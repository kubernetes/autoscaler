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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"
	componentbaseoptions "k8s.io/component-base/config/options"
	"k8s.io/klog/v2"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/checkpoint"
	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	input_metrics "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/metrics"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/routines"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_quality "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/quality"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	metrics_resources "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/resources"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/server"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	// aggregateContainerStateGCInterval defines how often expired AggregateContainerStates are garbage collected.
	aggregateContainerStateGCInterval               = 1 * time.Hour
	scaleCacheEntryLifetime           time.Duration = time.Hour
	scaleCacheEntryFreshnessTime      time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor       float64       = 1.
	scaleCacheLoopPeriod                            = 7 * time.Second
	defaultResyncPeriod               time.Duration = 10 * time.Minute
)

var config *recommender_config.RecommenderConfig

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
		run(ctx, healthCheck, config.CommonFlags)
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
					run(ctx, healthCheck, config.CommonFlags)
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

func run(ctx context.Context, healthCheck *metrics.HealthCheck, commonFlag *common.CommonFlags) {
	// Create a stop channel that will be used to signal shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)
	kubeConfig := common.CreateKubeConfigOrDie(commonFlag.KubeConfig, float32(commonFlag.KubeApiQps), int(commonFlag.KubeApiBurst))
	kubeClient := kube_client.NewForConfigOrDie(kubeConfig)
	clusterState := model.NewClusterState(aggregateContainerStateGCInterval)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(commonFlag.VpaObjectNamespace))
	controllerFetcher := controllerfetcher.NewControllerFetcher(kubeConfig, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	podLister, oomObserver := input.NewPodListerAndOOMObserver(ctx, kubeClient, commonFlag.VpaObjectNamespace, stopCh)

	factory.Start(stopCh)
	informerMap := factory.WaitForCacheSync(stopCh)
	for kind, synced := range informerMap {
		if !synced {
			klog.ErrorS(nil, fmt.Sprintf("Could not sync cache for the %s informer", kind.String()))
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}

	model.InitializeAggregationsConfig(model.NewAggregationsConfig(config.MemoryAggregationInterval, config.MemoryAggregationIntervalCount, config.MemoryHistogramDecayHalfLife, config.CpuHistogramDecayHalfLife, config.OOMBumpUpRatio, config.OOMMinBumpUp))

	useCheckpoints := config.Storage != "prometheus"

	var postProcessors []routines.RecommendationPostProcessor
	if config.PostProcessorCPUasInteger {
		postProcessors = append(postProcessors, &routines.IntegerCPUPostProcessor{})
	}

	globalMaxAllowed := initGlobalMaxAllowed()
	// CappingPostProcessor, should always come in the last position for post-processing
	postProcessors = append(postProcessors, routines.NewCappingRecommendationProcessor(globalMaxAllowed))
	var source input_metrics.PodMetricsLister
	if config.UseExternalMetrics {
		resourceMetrics := map[corev1.ResourceName]string{}
		if config.ExternalCpuMetric != "" {
			resourceMetrics[corev1.ResourceCPU] = config.ExternalCpuMetric
		}
		if config.ExternalMemoryMetric != "" {
			resourceMetrics[corev1.ResourceMemory] = config.ExternalMemoryMetric
		}
		externalClientOptions := &input_metrics.ExternalClientOptions{ResourceMetrics: resourceMetrics, ContainerNameLabel: config.CtrNameLabel}
		klog.V(1).InfoS("Using External Metrics", "options", externalClientOptions)
		source = input_metrics.NewExternalClient(kubeConfig, clusterState, *externalClientOptions)
	} else {
		klog.V(1).InfoS("Using Metrics Server")
		source = input_metrics.NewPodMetricsesSource(resourceclient.NewForConfigOrDie(kubeConfig))
	}

	ignoredNamespaces := strings.Split(commonFlag.IgnoredVpaObjectNamespaces, ",")

	clusterStateFeeder := input.ClusterStateFeederFactory{
		PodLister:           podLister,
		OOMObserver:         oomObserver,
		KubeClient:          kubeClient,
		MetricsClient:       input_metrics.NewMetricsClient(source, commonFlag.VpaObjectNamespace, "default-metrics-client"),
		VpaCheckpointClient: vpa_clientset.NewForConfigOrDie(kubeConfig).AutoscalingV1(),
		VpaLister:           vpa_api_util.NewVpasLister(vpa_clientset.NewForConfigOrDie(kubeConfig), make(chan struct{}), commonFlag.VpaObjectNamespace),
		VpaCheckpointLister: vpa_api_util.NewVpaCheckpointLister(vpa_clientset.NewForConfigOrDie(kubeConfig), make(chan struct{}), commonFlag.VpaObjectNamespace),
		ClusterState:        clusterState,
		SelectorFetcher:     target.NewVpaTargetSelectorFetcher(kubeConfig, kubeClient, factory),
		MemorySaveMode:      config.MemorySaver,
		ControllerFetcher:   controllerFetcher,
		RecommenderName:     config.RecommenderName,
		IgnoredNamespaces:   ignoredNamespaces,
		VpaObjectNamespace:  commonFlag.VpaObjectNamespace,
	}.Make()
	controllerFetcher.Start(ctx, scaleCacheLoopPeriod)

	recommender := routines.RecommenderFactory{
		ClusterState:       clusterState,
		ClusterStateFeeder: clusterStateFeeder,
		ControllerFetcher:  controllerFetcher,
		CheckpointWriter:   checkpoint.NewCheckpointWriter(clusterState, vpa_clientset.NewForConfigOrDie(kubeConfig).AutoscalingV1()),
		VpaClient:          vpa_clientset.NewForConfigOrDie(kubeConfig).AutoscalingV1(),
		PodResourceRecommender: logic.CreatePodResourceRecommender(logic.RecommendationConfig{
			SafetyMarginFraction:       config.SafetyMarginFraction,
			PodMinCPUMillicores:        config.PodMinCPUMillicores,
			PodMinMemoryMb:             config.PodMinMemoryMb,
			TargetCPUPercentile:        config.TargetCPUPercentile,
			LowerBoundCPUPercentile:    config.LowerBoundCPUPercentile,
			UpperBoundCPUPercentile:    config.UpperBoundCPUPercentile,
			ConfidenceIntervalCPU:      config.ConfidenceIntervalCPU,
			TargetMemoryPercentile:     config.TargetMemoryPercentile,
			LowerBoundMemoryPercentile: config.LowerBoundMemoryPercentile,
			UpperBoundMemoryPercentile: config.UpperBoundMemoryPercentile,
			ConfidenceIntervalMemory:   config.ConfidenceIntervalMemory,
		}),
		RecommendationFormat: logic.RecommendationFormat{
			HumanizeMemory:     config.HumanizeMemory,
			RoundCPUMillicores: config.RoundCPUMillicores,
			RoundMemoryBytes:   config.RoundMemoryBytes,
		},
		RecommendationPostProcessors: postProcessors,
		CheckpointsGCInterval:        config.CheckpointsGCInterval,
		CheckpointsWriteTimeout:      config.CheckpointsWriteTimeout,
		UseCheckpoints:               useCheckpoints,
		UpdateWorkerCount:            config.UpdateWorkerCount,
	}.Make()

	promQueryTimeout, err := time.ParseDuration(config.QueryTimeout)
	if err != nil {
		klog.ErrorS(err, "Could not parse --prometheus-query-timeout as a time.Duration")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if useCheckpoints {
		recommender.GetClusterStateFeeder().InitFromCheckpoints(ctx)
	} else {
		config := history.PrometheusHistoryProviderConfig{
			Address:                config.PrometheusAddress,
			Insecure:               config.PrometheusInsecure,
			QueryTimeout:           promQueryTimeout,
			HistoryLength:          config.HistoryLength,
			HistoryResolution:      config.HistoryResolution,
			PodLabelPrefix:         config.PodLabelPrefix,
			PodLabelsMetricName:    config.PodLabelsMetricName,
			PodNamespaceLabel:      config.PodNamespaceLabel,
			PodNameLabel:           config.PodNameLabel,
			CtrNamespaceLabel:      config.CtrNamespaceLabel,
			CtrPodNameLabel:        config.CtrPodNameLabel,
			CtrNameLabel:           config.CtrNameLabel,
			CadvisorMetricsJobName: config.PrometheusJobName,
			Namespace:              commonFlag.VpaObjectNamespace,
			Authentication: history.PrometheusCredentials{
				BearerToken: config.PrometheusBearerToken,
				Username:    config.Username,
				Password:    config.Password,
			},
		}
		provider, err := history.NewPrometheusHistoryProvider(config)
		if err != nil {
			klog.ErrorS(err, "Could not initialize history provider")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		recommender.GetClusterStateFeeder().InitFromHistoryProvider(provider)
	}

	// Start updating health check endpoint.
	healthCheck.StartMonitoring()

	ticker := time.Tick(config.MetricsFetcherInterval)
	for range ticker {
		recommender.RunOnce()
		healthCheck.UpdateLastActivity()
	}
}

func initGlobalMaxAllowed() corev1.ResourceList {
	result := make(corev1.ResourceList)
	if !config.MaxAllowedCPU.IsZero() {
		result[corev1.ResourceCPU] = config.MaxAllowedCPU.Quantity
	}
	if !config.MaxAllowedMemory.IsZero() {
		result[corev1.ResourceMemory] = config.MaxAllowedMemory.Quantity
	}

	return result
}
