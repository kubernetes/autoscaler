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
	"time"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"
	componentbaseoptions "k8s.io/component-base/config/options"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/routines"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_quality "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/quality"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	metrics_resources "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/resources"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/server"
)

var config *recommender_config.RecommenderConfig

const (
	// defaultResyncPeriod is the default resync period for shared informer factories.
	defaultResyncPeriod time.Duration = 10 * time.Minute
)

func main() {
	// Leader election needs to be initialized before any other flag, because it may be used in other flag's validation.
	leaderElection := defaultLeaderElectionConfiguration()
	componentbaseoptions.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	config = recommender_config.InitRecommenderFlags()

	klog.V(1).InfoS("Vertical Pod Autoscaler Recommender", "version", common.VerticalPodAutoscalerVersion(), "recommenderName", config.RecommenderName)

	ctx := context.Background()

	healthCheck := metrics.NewHealthCheck(config.MetricsFetcherInterval * 5)
	metrics_recommender.Register()
	metrics_quality.Register()
	metrics_resources.Register()
	server.Initialize(&config.CommonFlags.EnableProfiling, healthCheck, &config.Address)

	if !leaderElection.LeaderElect {
		run(ctx, healthCheck, config)
	} else {
		id, err := os.Hostname()
		if err != nil {
			klog.ErrorS(err, "Unable to get hostname")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		id = id + "_" + string(uuid.NewUUID())

		kubeconfig := common.CreateKubeConfigOrDie(config.CommonFlags.KubeConfig, float32(config.CommonFlags.KubeApiQps), int(config.CommonFlags.KubeApiBurst))
		kubeClient := kube_client.NewForConfigOrDie(kubeconfig)

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

		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock:            lock,
			LeaseDuration:   leaderElection.LeaseDuration.Duration,
			RenewDeadline:   leaderElection.RenewDeadline.Duration,
			RetryPeriod:     leaderElection.RetryPeriod.Duration,
			ReleaseOnCancel: true,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(_ context.Context) {
					run(ctx, healthCheck, config)
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
		LeaderElect:   false,
		LeaseDuration: metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline: metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:   metav1.Duration{Duration: defaultRetryPeriod},
		ResourceLock:  resourcelock.LeasesResourceLock,
		// This was changed from "vpa-recommender" to avoid conflicts with managed VPA deployments.
		ResourceName:      "vpa-recommender-lease",
		ResourceNamespace: metav1.NamespaceSystem,
	}
}

func run(ctx context.Context, healthCheck *metrics.HealthCheck, config *recommender_config.RecommenderConfig) {
	// Create a stop channel that will be used to signal shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)
	kubeConfig := common.CreateKubeConfigOrDie(config.CommonFlags.KubeConfig, float32(config.CommonFlags.KubeApiQps), int(config.CommonFlags.KubeApiBurst))
	kubeClient := kube_client.NewForConfigOrDie(kubeConfig)
	vpaClient := vpa_clientset.NewForConfigOrDie(kubeConfig)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(config.CommonFlags.VpaObjectNamespace))

	controller, err := routines.NewRecommenderController(
		ctx,
		kubeConfig,
		kubeClient,
		vpaClient,
		factory,
		config,
		healthCheck,
		stopCh,
	)

	if err != nil {
		klog.ErrorS(err, "Failed to create recommender controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	factory.Start(stopCh)
	informerMap := factory.WaitForCacheSync(stopCh)
	for kind, synced := range informerMap {
		if !synced {
			klog.ErrorS(nil, fmt.Sprintf("Could not sync cache for the %s informer", kind.String()))
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}

	if err := controller.Run(ctx); err != nil {
		klog.ErrorS(err, "Recommender controller exited with error")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
