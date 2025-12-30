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
	apiv1 "k8s.io/api/core/v1"
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
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/app"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/checkpoint"
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

var config *app.RecommenderConfig

func main() {
	config = app.DefaultRecommenderConfig()
	config.CommonFlags = common.InitCommonFlags()

	fs := pflag.CommandLine
	fs.StringVar(&config.RecommenderName, "recommender-name", config.RecommenderName, "Set the recommender name. Recommender will generate recommendations for VPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster.")
	fs.DurationVar(&config.MetricsFetcherInterval, "recommender-interval", config.MetricsFetcherInterval, `How often metrics should be fetched`)
	fs.DurationVar(&config.CheckpointsGCInterval, "checkpoints-gc-interval", config.CheckpointsGCInterval, `How often orphaned checkpoints should be garbage collected`)
	fs.StringVar(&config.Address, "address", ":8942", "The address to expose Prometheus metrics.")
	fs.StringVar(&config.Storage, "storage", config.Storage, `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
	fs.BoolVar(&config.MemorySaver, "memory-saver", false, `If true, only track pods which have an associated VPA`)
	fs.IntVar(&config.UpdateWorkerCount, "update-worker-count", 10, "Number of concurrent workers to update VPA recommendations and checkpoints. When increasing this setting, make sure the client-side rate limits ('kube-api-qps' and 'kube-api-burst') are either increased or turned off as well. Determines the minimum number of VPA checkpoints written per recommender loop.")

	// Prometheus history provider flags
	fs.StringVar(&config.PrometheusAddress, "prometheus-address", config.PrometheusAddress, `Where to reach for Prometheus metrics`)
	fs.BoolVar(&config.PrometheusInsecure, "prometheus-insecure", config.PrometheusInsecure, `Skip tls verify if https is used in the prometheus-address`)
	fs.StringVar(&config.PrometheusJobName, "prometheus-cadvisor-job-name", config.PrometheusJobName, `Name of the prometheus job name which scrapes the cAdvisor metrics`)
	fs.StringVar(&config.HistoryLength, "history-length", config.HistoryLength, `How much time back prometheus have to be queried to get historical metrics`)
	fs.StringVar(&config.HistoryResolution, "history-resolution", config.HistoryResolution, `Resolution at which Prometheus is queried for historical metrics`)
	fs.StringVar(&config.QueryTimeout, "prometheus-query-timeout", config.QueryTimeout, `How long to wait before killing long queries`)
	fs.StringVar(&config.PodLabelPrefix, "pod-label-prefix", config.PodLabelPrefix, `Which prefix to look for pod labels in metrics`)
	fs.StringVar(&config.PodLabelsMetricName, "metric-for-pod-labels", config.PodLabelsMetricName, `Which metric to look for pod labels in metrics`)
	fs.StringVar(&config.PodNamespaceLabel, "pod-namespace-label", config.PodNamespaceLabel, `Label name to look for pod namespaces`)
	fs.StringVar(&config.PodNameLabel, "pod-name-label", config.PodNameLabel, `Label name to look for pod names`)
	fs.StringVar(&config.CtrNamespaceLabel, "container-namespace-label", config.CtrNamespaceLabel, `Label name to look for container namespaces`)
	fs.StringVar(&config.CtrPodNameLabel, "container-pod-name-label", config.CtrPodNameLabel, `Label name to look for container pod names`)
	fs.StringVar(&config.CtrNameLabel, "container-name-label", config.CtrNameLabel, `Label name to look for container names`)
	fs.StringVar(&config.Username, "username", config.Username, "The username used in the prometheus server basic auth. Can also be set via the PROMETHEUS_USERNAME environment variable")
	fs.StringVar(&config.Password, "password", config.Password, "The password used in the prometheus server basic auth. Can also be set via the PROMETHEUS_PASSWORD environment variable")
	fs.StringVar(&config.PrometheusBearerToken, "prometheus-bearer-token", config.PrometheusBearerToken, "The bearer token used in the Prometheus server bearer token auth")
	fs.StringVar(&config.PrometheusBearerTokenFile, "prometheus-bearer-token-file", config.PrometheusBearerTokenFile, "Path to the bearer token file used for authentication by the Prometheus server")

	// External metrics provider flags
	fs.BoolVar(&config.UseExternalMetrics, "use-external-metrics", config.UseExternalMetrics, "ALPHA.  Use an external metrics provider instead of metrics_server.")
	fs.StringVar(&config.ExternalCpuMetric, "external-metrics-cpu-metric", config.ExternalCpuMetric, "ALPHA.  Metric to use with external metrics provider for CPU usage.")
	fs.StringVar(&config.ExternalMemoryMetric, "external-metrics-memory-metric", config.ExternalMemoryMetric, "ALPHA.  Metric to use with external metrics provider for memory usage.")

	// Aggregation configuration flags
	fs.DurationVar(&config.MemoryAggregationInterval, "memory-aggregation-interval", config.MemoryAggregationInterval, `The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval)`)
	fs.Int64Var(&config.MemoryAggregationIntervalCount, "memory-aggregation-interval-count", config.MemoryAggregationIntervalCount, `The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count.`)
	fs.DurationVar(&config.MemoryHistogramDecayHalfLife, "memory-histogram-decay-half-life", config.MemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	fs.DurationVar(&config.CpuHistogramDecayHalfLife, "cpu-histogram-decay-half-life", config.CpuHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
	fs.Float64Var(&config.OOMBumpUpRatio, "oom-bump-up-ratio", config.OOMBumpUpRatio, `Default memory bump up ratio when OOM occurs. This value applies to all VPAs unless overridden in the VPA spec. Default is 1.2.`)
	fs.Float64Var(&config.OOMMinBumpUp, "oom-min-bump-up-bytes", config.OOMMinBumpUp, `Default minimal increase of memory (in bytes) when OOM occurs. This value applies to all VPAs unless overridden in the VPA spec. Default is 100 * 1024 * 1024 (100Mi).`)

	// Post processors flags
	// CPU as integer to benefit for CPU management Static Policy ( https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#static-policy )
	fs.BoolVar(&config.PostProcessorCPUasInteger, "cpu-integer-post-processor-enabled", config.PostProcessorCPUasInteger, "Enable the cpu-integer recommendation post processor. The post processor will round up CPU recommendations to a whole CPU for pods which were opted in by setting an appropriate label on VPA object (experimental)")
	fs.Var(&config.MaxAllowedCPU, "container-recommendation-max-allowed-cpu", "Maximum amount of CPU that will be recommended for a container. VerticalPodAutoscaler-level maximum allowed takes precedence over the global maximum allowed.")
	fs.Var(&config.MaxAllowedMemory, "container-recommendation-max-allowed-memory", "Maximum amount of memory that will be recommended for a container. VerticalPodAutoscaler-level maximum allowed takes precedence over the global maximum allowed.")

	klog.InitFlags(nil)
	common.InitLoggingFlags()

	leaderElection := defaultLeaderElectionConfiguration()
	componentbaseoptions.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	features.MutableFeatureGate.AddFlag(pflag.CommandLine)

	kube_flag.InitFlags()
	klog.V(1).InfoS("Vertical Pod Autoscaler Recommender", "version", common.VerticalPodAutoscalerVersion(), "recommenderName", config.RecommenderName)

	common.ValidateCommonConfig(config.CommonFlags)
	app.ValidateRecommenderConfig(config)

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
		resourceMetrics := map[apiv1.ResourceName]string{}
		if config.ExternalCpuMetric != "" {
			resourceMetrics[apiv1.ResourceCPU] = config.ExternalCpuMetric
		}
		if config.ExternalMemoryMetric != "" {
			resourceMetrics[apiv1.ResourceMemory] = config.ExternalMemoryMetric
		}
		externalClientOptions := &input_metrics.ExternalClientOptions{ResourceMetrics: resourceMetrics, ContainerNameLabel: *&config.CtrNameLabel}
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
		ClusterState:                 clusterState,
		ClusterStateFeeder:           clusterStateFeeder,
		ControllerFetcher:            controllerFetcher,
		CheckpointWriter:             checkpoint.NewCheckpointWriter(clusterState, vpa_clientset.NewForConfigOrDie(kubeConfig).AutoscalingV1()),
		VpaClient:                    vpa_clientset.NewForConfigOrDie(kubeConfig).AutoscalingV1(),
		PodResourceRecommender:       logic.CreatePodResourceRecommender(),
		RecommendationPostProcessors: postProcessors,
		CheckpointsGCInterval:        config.CheckpointsGCInterval,
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

func initGlobalMaxAllowed() apiv1.ResourceList {
	result := make(apiv1.ResourceList)
	if !config.MaxAllowedCPU.IsZero() {
		result[apiv1.ResourceCPU] = config.MaxAllowedCPU.Quantity
	}
	if !config.MaxAllowedMemory.IsZero() {
		result[apiv1.ResourceMemory] = config.MaxAllowedMemory.Quantity
	}

	return result
}
