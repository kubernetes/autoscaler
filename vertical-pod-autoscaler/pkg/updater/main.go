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
	"fmt"
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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/recommendation"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	updater_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/config"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/inplace"
	updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/updater"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/server"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	defaultResyncPeriod          time.Duration = 10 * time.Minute
	scaleCacheEntryLifetime      time.Duration = time.Hour
	scaleCacheEntryFreshnessTime time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor  float64       = 1.
)

var config *updater_config.UpdaterConfig

func main() {
	config = updater_config.InitUpdaterFlags()
	klog.InitFlags(nil)
	common.InitLoggingFlags()

	leaderElection := defaultLeaderElectionConfiguration()
	componentbaseoptions.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	features.MutableFeatureGate.AddFlag(pflag.CommandLine)

	kube_flag.InitFlags()
	klog.V(1).InfoS("Vertical Pod Autoscaler Updater", "version", common.VerticalPodAutoscalerVersion())

	healthCheck := metrics.NewHealthCheck(config.UpdaterInterval * 5)
	server.Initialize(&config.CommonFlags.EnableProfiling, healthCheck, &config.Address)

	metrics_updater.Register()

	if !leaderElection.LeaderElect {
		run(healthCheck, config.CommonFlags)
	} else {
		id, err := os.Hostname()
		if err != nil {
			klog.ErrorS(err, "Unable to get hostname")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		id = id + "_" + string(uuid.NewUUID())

		kubeConfig := common.CreateKubeConfigOrDie(config.CommonFlags.KubeConfig, float32(config.CommonFlags.KubeApiQps), int(config.CommonFlags.KubeApiBurst))
		kubeClient := kube_client.NewForConfigOrDie(kubeConfig)

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
			klog.ErrorS(err, "Unable to create leader election lock")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
			Lock:            lock,
			LeaseDuration:   leaderElection.LeaseDuration.Duration,
			RenewDeadline:   leaderElection.RenewDeadline.Duration,
			RetryPeriod:     leaderElection.RetryPeriod.Duration,
			ReleaseOnCancel: true,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ context.Context) {
					run(healthCheck, config.CommonFlags)
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
	stopCh := make(chan struct{})
	defer close(stopCh)

	kubeConfig := common.CreateKubeConfigOrDie(commonFlag.KubeConfig, float32(commonFlag.KubeApiQps), int(commonFlag.KubeApiBurst))
	kubeClient := kube_client.NewForConfigOrDie(kubeConfig)
	vpaClient := vpa_clientset.NewForConfigOrDie(kubeConfig)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(commonFlag.VpaObjectNamespace))
	targetSelectorFetcher := target.NewVpaTargetSelectorFetcher(kubeConfig, kubeClient, factory)
	controllerFetcher := controllerfetcher.NewControllerFetcher(kubeConfig, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)

	var limitRangeCalculator limitrange.LimitRangeCalculator
	limitRangeCalculator, err := limitrange.NewLimitsRangeCalculator(factory)
	if err != nil {
		klog.ErrorS(err, "Failed to create limitRangeCalculator, falling back to not checking limits")
		limitRangeCalculator = limitrange.NewNoopLimitsCalculator()
	}

	factory.Start(stopCh)
	informerMap := factory.WaitForCacheSync(stopCh)
	for kind, synced := range informerMap {
		if !synced {
			klog.ErrorS(nil, fmt.Sprintf("Could not sync cache for the %s informer", kind.String()))
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}

	admissionControllerStatusNamespace := status.AdmissionControllerStatusNamespace
	if config.Namespace != "" {
		admissionControllerStatusNamespace = config.Namespace
	}

	ignoredNamespaces := strings.Split(commonFlag.IgnoredVpaObjectNamespaces, ",")

	recommendationProvider := recommendation.NewProvider(limitRangeCalculator, vpa_api_util.NewCappingRecommendationProcessor(limitRangeCalculator))

	calculators := []patch.Calculator{inplace.NewResourceInPlaceUpdatesCalculator(recommendationProvider), inplace.NewInPlaceUpdatedCalculator()}

	// TODO: use SharedInformerFactory in updater
	updater, err := updater.NewUpdater(
		kubeClient,
		vpaClient,
		config.MinReplicas,
		config.EvictionRateLimit,
		config.EvictionRateBurst,
		config.EvictionToleranceFraction,
		config.UseAdmissionControllerStatus,
		config.InPlaceSkipDisruptionBudget,
		admissionControllerStatusNamespace,
		vpa_api_util.NewCappingRecommendationProcessor(limitRangeCalculator),
		priority.NewScalingDirectionPodEvictionAdmission(),
		targetSelectorFetcher,
		controllerFetcher,
		priority.NewProcessor(),
		commonFlag.VpaObjectNamespace,
		ignoredNamespaces,
		calculators,
	)
	if err != nil {
		klog.ErrorS(err, "Failed to create updater")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	// Start updating health check endpoint.
	healthCheck.StartMonitoring()

	ticker := time.Tick(config.UpdaterInterval)
	for range ticker {
		ctx, cancel := context.WithTimeout(context.Background(), config.UpdaterInterval)
		updater.RunOnce(ctx)
		healthCheck.UpdateLastActivity()
		cancel()
	}
}
