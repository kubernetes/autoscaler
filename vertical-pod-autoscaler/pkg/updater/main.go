/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	kube_flag "k8s.io/component-base/cli/flag"
	componentbaseconfig "k8s.io/component-base/config"
	componentbaseoptions "k8s.io/component-base/config/options"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/updater"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/server"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

var (
	updaterInterval = flag.Duration("updater-interval", 1*time.Minute,
		`How often updater should run`)

	minReplicas = flag.Int("min-replicas", 2,
		`Minimum number of replicas to perform update`)

	evictionToleranceFraction = flag.Float64("eviction-tolerance", 0.5,
		`Fraction of replica count that can be evicted for update, if more than one pod can be evicted.`)

	evictionRateLimit = flag.Float64("eviction-rate-limit", -1,
		`Number of pods that can be evicted per seconds. A rate limit set to 0 or -1 will disable
		the rate limiter.`)

	evictionRateBurst = flag.Int("eviction-rate-burst", 1, `Burst of pods that can be evicted.`)

	address = flag.String("address", ":8943", "The address to expose Prometheus metrics.")

	useAdmissionControllerStatus = flag.Bool("use-admission-controller-status", true,
		"If true, updater will only evict pods when admission controller status is valid.")

	namespace = os.Getenv("NAMESPACE")
)

const (
	defaultResyncPeriod          time.Duration = 10 * time.Minute
	scaleCacheEntryLifetime      time.Duration = time.Hour
	scaleCacheEntryFreshnessTime time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor  float64       = 1.
)

func main() {
	commonFlags := common.InitCommonFlags()
	klog.InitFlags(nil)
	common.InitLoggingFlags()

	leaderElection := defaultLeaderElectionConfiguration()
	componentbaseoptions.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	kube_flag.InitFlags()
	klog.V(1).InfoS("Vertical Pod Autoscaler Updater", "version", common.VerticalPodAutoscalerVersion)

	if len(commonFlags.VpaObjectNamespace) > 0 && len(commonFlags.IgnoredVpaObjectNamespaces) > 0 {
		klog.Fatalf("--vpa-object-namespace and --ignored-vpa-object-namespaces are mutually exclusive and can't be set together.")
	}

	healthCheck := metrics.NewHealthCheck(*updaterInterval * 5)
	server.Initialize(&commonFlags.EnableProfiling, healthCheck, address)

	metrics_updater.Register()

	if !leaderElection.LeaderElect {
		run(healthCheck, commonFlags)
	} else {
		id, err := os.Hostname()
		if err != nil {
			klog.Fatalf("Unable to get hostname: %v", err)
		}
		id = id + "_" + string(uuid.NewUUID())

		config := common.CreateKubeConfigOrDie(commonFlags.KubeConfig, float32(commonFlags.KubeApiQps), int(commonFlags.KubeApiBurst))
		kubeClient := kube_client.NewForConfigOrDie(config)

		lock, err := resourcelock.New(
			leaderElection.ResourceLock,
			leaderElection.ResourceNamespace,
			leaderElection.ResourceName,
			kubeClient.CoreV1(),
			kubeClient.CoordinationV1(),
			resourcelock.ResourceLockConfig{
				Identity: id,
			},
		)
		if err != nil {
			klog.Fatalf("Unable to create leader election lock: %v", err)
		}

		leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
			Lock:            lock,
			LeaseDuration:   leaderElection.LeaseDuration.Duration,
			RenewDeadline:   leaderElection.RenewDeadline.Duration,
			RetryPeriod:     leaderElection.RetryPeriod.Duration,
			ReleaseOnCancel: true,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ context.Context) {
					run(healthCheck, commonFlags)
				},
				OnStoppedLeading: func() {
					klog.Fatal("lost master")
				},
			},
		})
	}
}

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

func defaultLeaderElectionConfiguration() componentbaseconfig.LeaderElectionConfiguration {
	return componentbaseconfig.LeaderElectionConfiguration{
		LeaderElect:       false,
		LeaseDuration:     metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline:     metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:       metav1.Duration{Duration: defaultRetryPeriod},
		ResourceLock:      resourcelock.LeasesResourceLock,
		ResourceName:      "vpa-updater",
		ResourceNamespace: metav1.NamespaceSystem,
	}
}

func run(healthCheck *metrics.HealthCheck, commonFlag *common.CommonFlags) {
	config := common.CreateKubeConfigOrDie(commonFlag.KubeConfig, float32(commonFlag.KubeApiQps), int(commonFlag.KubeApiBurst))
	kubeClient := kube_client.NewForConfigOrDie(config)
	vpaClient := vpa_clientset.NewForConfigOrDie(config)
	factory := informers.NewSharedInformerFactory(kubeClient, defaultResyncPeriod)
	targetSelectorFetcher := target.NewVpaTargetSelectorFetcher(config, kubeClient, factory)
	controllerFetcher := controllerfetcher.NewControllerFetcher(config, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	var limitRangeCalculator limitrange.LimitRangeCalculator
	limitRangeCalculator, err := limitrange.NewLimitsRangeCalculator(factory)
	if err != nil {
		klog.ErrorS(err, "Failed to create limitRangeCalculator, falling back to not checking limits")
		limitRangeCalculator = limitrange.NewNoopLimitsCalculator()
	}
	admissionControllerStatusNamespace := status.AdmissionControllerStatusNamespace
	if namespace != "" {
		admissionControllerStatusNamespace = namespace
	}

	ignoredNamespaces := strings.Split(commonFlag.IgnoredVpaObjectNamespaces, ",")

	// TODO: use SharedInformerFactory in updater
	updater, err := updater.NewUpdater(
		kubeClient,
		vpaClient,
		*minReplicas,
		*evictionRateLimit,
		*evictionRateBurst,
		*evictionToleranceFraction,
		*useAdmissionControllerStatus,
		admissionControllerStatusNamespace,
		vpa_api_util.NewCappingRecommendationProcessor(limitRangeCalculator),
		priority.NewScalingDirectionPodEvictionAdmission(),
		targetSelectorFetcher,
		controllerFetcher,
		priority.NewProcessor(),
		commonFlag.VpaObjectNamespace,
		ignoredNamespaces,
	)
	if err != nil {
		klog.Fatalf("Failed to create updater: %v", err)
	}

	// Start updating health check endpoint.
	healthCheck.StartMonitoring()

	ticker := time.Tick(*updaterInterval)
	for range ticker {
		ctx, cancel := context.WithTimeout(context.Background(), *updaterInterval)
		updater.RunOnce(ctx)
		healthCheck.UpdateLastActivity()
		cancel()
	}
}
