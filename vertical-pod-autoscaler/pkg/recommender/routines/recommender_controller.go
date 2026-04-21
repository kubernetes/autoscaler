/*
Copyright The Kubernetes Authors.

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

package routines

import (
	"context"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/checkpoint"
	recommender_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/config"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	input_metrics "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/metrics"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	// aggregateContainerStateGCInterval defines how often expired AggregateContainerStates are garbage collected.
	aggregateContainerStateGCInterval               = 1 * time.Hour
	scaleCacheEntryLifetime           time.Duration = time.Hour
	scaleCacheEntryFreshnessTime      time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor       float64       = 1.
	scaleCacheLoopPeriod                            = 7 * time.Second
)

// RecommenderController wraps the Recommender with the full lifecycle
// (informer setup, ticker loop, health checks) needed to run as a standalone controller.
// It implements controllercontext.Controller.
type RecommenderController struct {
	recommender Recommender
	interval    time.Duration
	healthCheck *metrics.HealthCheck
}

// NewRecommenderController creates a RecommenderController
func NewRecommenderController(
	ctx context.Context,
	kubeConfig *rest.Config,
	kubeClient kube_client.Interface,
	vpaClient *vpa_clientset.Clientset,
	factory informers.SharedInformerFactory,
	config *recommender_config.RecommenderConfig,
	healthCheck *metrics.HealthCheck,
	stopCh <-chan struct{},
) (*RecommenderController, error) {
	commonFlags := config.CommonFlags

	clusterState := model.NewClusterState(aggregateContainerStateGCInterval)
	controllerFetcher := controllerfetcher.NewControllerFetcher(kubeConfig, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor, stopCh)
	podLister, oomObserver := input.NewPodListerAndOOMObserver(ctx, kubeClient, commonFlags.VpaObjectNamespace, stopCh)

	model.InitializeAggregationsConfig(model.NewAggregationsConfig(
		config.MemoryAggregationInterval,
		config.MemoryAggregationIntervalCount,
		config.MemoryHistogramDecayHalfLife,
		config.CpuHistogramDecayHalfLife,
		config.OOMBumpUpRatio,
		config.OOMMinBumpUp,
	))

	useCheckpoints := config.Storage != "prometheus"

	var postProcessors []RecommendationPostProcessor
	if config.PostProcessorCPUasInteger {
		postProcessors = append(postProcessors, &IntegerCPUPostProcessor{})
	}

	globalMaxAllowed := initGlobalMaxAllowed(config)
	postProcessors = append(postProcessors, NewCappingRecommendationProcessor(globalMaxAllowed))

	var source input_metrics.PodMetricsLister
	if config.UseExternalMetrics {
		resourceMetrics := map[corev1.ResourceName]string{}
		if config.ExternalCpuMetric != "" {
			resourceMetrics[corev1.ResourceCPU] = config.ExternalCpuMetric
		}
		if config.ExternalMemoryMetric != "" {
			resourceMetrics[corev1.ResourceMemory] = config.ExternalMemoryMetric
		}
		externalClientOptions := &input_metrics.ExternalClientOptions{
			ResourceMetrics:    resourceMetrics,
			ContainerNameLabel: config.CtrNameLabel,
		}
		klog.V(1).InfoS("Using External Metrics", "options", externalClientOptions)
		source = input_metrics.NewExternalClient(kubeConfig, clusterState, *externalClientOptions)
	} else {
		klog.V(1).InfoS("Using Metrics Server")
		source = input_metrics.NewPodMetricsesSource(resourceclient.NewForConfigOrDie(kubeConfig))
	}

	ignoredNamespaces := strings.Split(commonFlags.IgnoredVpaObjectNamespaces, ",")

	clusterStateFeeder := input.ClusterStateFeederFactory{
		PodLister:           podLister,
		OOMObserver:         oomObserver,
		KubeClient:          kubeClient,
		MetricsClient:       input_metrics.NewMetricsClient(source, commonFlags.VpaObjectNamespace, "default-metrics-client"),
		VpaCheckpointClient: vpaClient.AutoscalingV1(),
		VpaLister:           vpa_api_util.NewVpasLister(vpaClient, stopCh, commonFlags.VpaObjectNamespace),
		VpaCheckpointLister: vpa_api_util.NewVpaCheckpointLister(vpaClient, stopCh, commonFlags.VpaObjectNamespace),
		ClusterState:        clusterState,
		SelectorFetcher:     target.NewVpaTargetSelectorFetcher(kubeConfig, kubeClient, factory, stopCh),
		MemorySaveMode:      config.MemorySaver,
		ControllerFetcher:   controllerFetcher,
		RecommenderName:     config.RecommenderName,
		IgnoredNamespaces:   ignoredNamespaces,
		VpaObjectNamespace:  commonFlags.VpaObjectNamespace,
	}.Make()
	controllerFetcher.Start(ctx, scaleCacheLoopPeriod)

	recommender := RecommenderFactory{
		ClusterState:       clusterState,
		ClusterStateFeeder: clusterStateFeeder,
		ControllerFetcher:  controllerFetcher,
		CheckpointWriter:   checkpoint.NewCheckpointWriter(clusterState, vpaClient.AutoscalingV1()),
		VpaClient:          vpaClient.AutoscalingV1(),
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

	if err := initHistoryProvider(ctx, recommender, config); err != nil {
		return nil, err
	}

	return &RecommenderController{
		recommender: recommender,
		interval:    config.MetricsFetcherInterval,
		healthCheck: healthCheck,
	}, nil
}

// Run starts the recommender loop and blocks until ctx is cancelled.
// It implements controllercontext.Controller.
func (c *RecommenderController) Run(ctx context.Context) error {
	c.healthCheck.StartMonitoring()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			c.recommender.RunOnce()
			c.healthCheck.UpdateLastActivity()
		}
	}
}

func initGlobalMaxAllowed(config *recommender_config.RecommenderConfig) corev1.ResourceList {
	result := make(corev1.ResourceList)
	if !config.MaxAllowedCPU.IsZero() {
		result[corev1.ResourceCPU] = config.MaxAllowedCPU.Quantity
	}
	if !config.MaxAllowedMemory.IsZero() {
		result[corev1.ResourceMemory] = config.MaxAllowedMemory.Quantity
	}
	return result
}

func initHistoryProvider(ctx context.Context, rec Recommender, config *recommender_config.RecommenderConfig) error {
	useCheckpoints := config.Storage != "prometheus"
	if useCheckpoints {
		rec.GetClusterStateFeeder().InitFromCheckpoints(ctx)
	} else {
		promQueryTimeout, err := time.ParseDuration(config.QueryTimeout)
		if err != nil {
			return err
		}
		histConfig := history.PrometheusHistoryProviderConfig{
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
			Namespace:              config.CommonFlags.VpaObjectNamespace,
			Authentication: history.PrometheusCredentials{
				BearerToken: config.PrometheusBearerToken,
				Username:    config.Username,
				Password:    config.Password,
			},
		}
		provider, err := history.NewPrometheusHistoryProvider(histConfig)
		if err != nil {
			return err
		}
		rec.GetClusterStateFeeder().InitFromHistoryProvider(provider)
	}
	return nil
}
