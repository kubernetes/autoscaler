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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

var (
	recommenderName        = flag.String("recommender-name", input.DefaultRecommenderName, "Set the recommender name. Recommender will generate recommendations for VPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster.")
	metricsFetcherInterval = flag.Duration("recommender-interval", 1*time.Minute, `How often metrics should be fetched`)
	checkpointsGCInterval  = flag.Duration("checkpoints-gc-interval", 10*time.Minute, `How often orphaned checkpoints should be garbage collected`)
	address                = flag.String("address", ":8942", "The address to expose Prometheus metrics.")
	storage                = flag.String("storage", "", `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
	memorySaver            = flag.Bool("memory-saver", false, `If true, only track pods which have an associated VPA`)
	updateWorkerCount      = flag.Int("update-worker-count", 10, "Number of concurrent workers to update VPA recommendations and checkpoints. When increasing this setting, make sure the client-side rate limits ('kube-api-qps' and 'kube-api-burst') are either increased or turned off as well. Determines the minimum number of VPA checkpoints written per recommender loop.")
)

// Prometheus history provider flags
var (
	prometheusAddress         = flag.String("prometheus-address", "http://prometheus.monitoring.svc", `Where to reach for Prometheus metrics`)
	prometheusInsecure        = flag.Bool("prometheus-insecure", false, `Skip tls verify if https is used in the prometheus-address`)
	prometheusJobName         = flag.String("prometheus-cadvisor-job-name", "kubernetes-cadvisor", `Name of the prometheus job name which scrapes the cAdvisor metrics`)
	historyLength             = flag.String("history-length", "8d", `How much time back prometheus have to be queried to get historical metrics`)
	historyResolution         = flag.String("history-resolution", "1h", `Resolution at which Prometheus is queried for historical metrics`)
	queryTimeout              = flag.String("prometheus-query-timeout", "5m", `How long to wait before killing long queries`)
	podLabelPrefix            = flag.String("pod-label-prefix", "pod_label_", `Which prefix to look for pod labels in metrics`)
	podLabelsMetricName       = flag.String("metric-for-pod-labels", "up{job=\"kubernetes-pods\"}", `Which metric to look for pod labels in metrics`)
	podNamespaceLabel         = flag.String("pod-namespace-label", "kubernetes_namespace", `Label name to look for pod namespaces`)
	podNameLabel              = flag.String("pod-name-label", "kubernetes_pod_name", `Label name to look for pod names`)
	ctrNamespaceLabel         = flag.String("container-namespace-label", "namespace", `Label name to look for container namespaces`)
	ctrPodNameLabel           = flag.String("container-pod-name-label", "pod_name", `Label name to look for container pod names`)
	ctrNameLabel              = flag.String("container-name-label", "name", `Label name to look for container names`)
	username                  = flag.String("username", "", "The username used in the prometheus server basic auth. Can also be set via the PROMETHEUS_USERNAME environment variable")
	password                  = flag.String("password", "", "The password used in the prometheus server basic auth. Can also be set via the PROMETHEUS_PASSWORD environment variable")
	prometheusBearerToken     = flag.String("prometheus-bearer-token", "", "The bearer token used in the Prometheus server bearer token auth")
	prometheusBearerTokenFile = flag.String("prometheus-bearer-token-file", "", "Path to the bearer token file used for authentication by the Prometheus server")
)

// External metrics provider flags
var (
	useExternalMetrics   = flag.Bool("use-external-metrics", false, "ALPHA.  Use an external metrics provider instead of metrics_server.")
	externalCpuMetric    = flag.String("external-metrics-cpu-metric", "", "ALPHA.  Metric to use with external metrics provider for CPU usage.")
	externalMemoryMetric = flag.String("external-metrics-memory-metric", "", "ALPHA.  Metric to use with external metrics provider for memory usage.")
)

// Aggregation configuration flags
var (
	memoryAggregationInterval      = flag.Duration("memory-aggregation-interval", model.DefaultMemoryAggregationInterval, `The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval)`)
	memoryAggregationIntervalCount = flag.Int64("memory-aggregation-interval-count", model.DefaultMemoryAggregationIntervalCount, `The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count.`)
	memoryHistogramDecayHalfLife   = flag.Duration("memory-histogram-decay-half-life", model.DefaultMemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	cpuHistogramDecayHalfLife      = flag.Duration("cpu-histogram-decay-half-life", model.DefaultCPUHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
	oomBumpUpRatio                 = flag.Float64("oom-bump-up-ratio", model.DefaultOOMBumpUpRatio, `The memory bump up ratio when OOM occurred, default is 1.2.`)
	oomMinBumpUp                   = flag.Float64("oom-min-bump-up-bytes", model.DefaultOOMMinBumpUp, `The minimal increase of memory when OOM occurred in bytes, default is 100 * 1024 * 1024`)
)

// Post processors flags
var (
	// CPU as integer to benefit for CPU management Static Policy ( https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#static-policy )
	postProcessorCPUasInteger = flag.Bool("cpu-integer-post-processor-enabled", false, "Enable the cpu-integer recommendation post processor. The post processor will round up CPU recommendations to a whole CPU for pods which were opted in by setting an appropriate label on VPA object (experimental)")
	maxAllowedCPU             = resource.QuantityValue{}
	maxAllowedMemory          = resource.QuantityValue{}
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

func init() {
	flag.Var(&maxAllowedCPU, "container-recommendation-max-allowed-cpu", "Maximum amount of CPU that will be recommended for a container. VerticalPodAutoscaler-level maximum allowed takes precedence over the global maximum allowed.")
	flag.Var(&maxAllowedMemory, "container-recommendation-max-allowed-memory", "Maximum amount of memory that will be recommended for a container. VerticalPodAutoscaler-level maximum allowed takes precedence over the global maximum allowed.")
}

func main() {
	commonFlags := common.InitCommonFlags()
	klog.InitFlags(nil)
	common.InitLoggingFlags()

	leaderElection := defaultLeaderElectionConfiguration()
	componentbaseoptions.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	features.MutableFeatureGate.AddFlag(pflag.CommandLine)

	kube_flag.InitFlags()
	klog.V(1).InfoS("Vertical Pod Autoscaler Recommender", "version", common.VerticalPodAutoscalerVersion(), "recommenderName", *recommenderName)

	if len(commonFlags.VpaObjectNamespace) > 0 && len(commonFlags.IgnoredVpaObjectNamespaces) > 0 {
		klog.ErrorS(nil, "--vpa-object-namespace and --ignored-vpa-object-namespaces are mutually exclusive and can't be set together.")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if *routines.MinCheckpointsPerRun != 10 { // Default value is 10
		klog.InfoS("DEPRECATION WARNING: The 'min-checkpoints' flag is deprecated and has no effect. It will be removed in a future release.")
	}

	if *prometheusBearerToken != "" && *prometheusBearerTokenFile != "" && *username != "" {
		klog.ErrorS(nil, "--bearer-token, --bearer-token-file and --username are mutually exclusive and can't be set together.")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if *prometheusBearerTokenFile != "" {
		fileContent, err := os.ReadFile(*prometheusBearerTokenFile)
		if err != nil {
			klog.ErrorS(err, "Unable to read bearer token file", "filename", *prometheusBearerTokenFile)
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		*prometheusBearerToken = strings.TrimSpace(string(fileContent))
	}

	ctx := context.Background()

	healthCheck := metrics.NewHealthCheck(*metricsFetcherInterval * 5)
	metrics_recommender.Register()
	metrics_quality.Register()
	metrics_resources.Register()
	server.Initialize(&commonFlags.EnableProfiling, healthCheck, address)

	if !leaderElection.LeaderElect {
		run(ctx, healthCheck, commonFlags)
	} else {
		id, err := os.Hostname()
		if err != nil {
			klog.ErrorS(err, "Unable to get hostname")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
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
					run(ctx, healthCheck, commonFlags)
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
	config := common.CreateKubeConfigOrDie(commonFlag.KubeConfig, float32(commonFlag.KubeApiQps), int(commonFlag.KubeApiBurst))
	kubeClient := kube_client.NewForConfigOrDie(config)
	clusterState := model.NewClusterState(aggregateContainerStateGCInterval)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(commonFlag.VpaObjectNamespace))
	controllerFetcher := controllerfetcher.NewControllerFetcher(config, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	podLister, oomObserver := input.NewPodListerAndOOMObserver(ctx, kubeClient, commonFlag.VpaObjectNamespace, stopCh)

	factory.Start(stopCh)
	informerMap := factory.WaitForCacheSync(stopCh)
	for kind, synced := range informerMap {
		if !synced {
			klog.ErrorS(nil, fmt.Sprintf("Could not sync cache for the %s informer", kind.String()))
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}

	model.InitializeAggregationsConfig(model.NewAggregationsConfig(*memoryAggregationInterval, *memoryAggregationIntervalCount, *memoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife, *oomBumpUpRatio, *oomMinBumpUp))

	useCheckpoints := *storage != "prometheus"

	var postProcessors []routines.RecommendationPostProcessor
	if *postProcessorCPUasInteger {
		postProcessors = append(postProcessors, &routines.IntegerCPUPostProcessor{})
	}

	globalMaxAllowed := initGlobalMaxAllowed()
	// CappingPostProcessor, should always come in the last position for post-processing
	postProcessors = append(postProcessors, routines.NewCappingRecommendationProcessor(globalMaxAllowed))
	var source input_metrics.PodMetricsLister
	if *useExternalMetrics {
		resourceMetrics := map[apiv1.ResourceName]string{}
		if externalCpuMetric != nil && *externalCpuMetric != "" {
			resourceMetrics[apiv1.ResourceCPU] = *externalCpuMetric
		}
		if externalMemoryMetric != nil && *externalMemoryMetric != "" {
			resourceMetrics[apiv1.ResourceMemory] = *externalMemoryMetric
		}
		externalClientOptions := &input_metrics.ExternalClientOptions{ResourceMetrics: resourceMetrics, ContainerNameLabel: *ctrNameLabel}
		klog.V(1).InfoS("Using External Metrics", "options", externalClientOptions)
		source = input_metrics.NewExternalClient(config, clusterState, *externalClientOptions)
	} else {
		klog.V(1).InfoS("Using Metrics Server")
		source = input_metrics.NewPodMetricsesSource(resourceclient.NewForConfigOrDie(config))
	}

	ignoredNamespaces := strings.Split(commonFlag.IgnoredVpaObjectNamespaces, ",")

	clusterStateFeeder := input.ClusterStateFeederFactory{
		PodLister:           podLister,
		OOMObserver:         oomObserver,
		KubeClient:          kubeClient,
		MetricsClient:       input_metrics.NewMetricsClient(source, commonFlag.VpaObjectNamespace, "default-metrics-client"),
		VpaCheckpointClient: vpa_clientset.NewForConfigOrDie(config).AutoscalingV1(),
		VpaLister:           vpa_api_util.NewVpasLister(vpa_clientset.NewForConfigOrDie(config), make(chan struct{}), commonFlag.VpaObjectNamespace),
		VpaCheckpointLister: vpa_api_util.NewVpaCheckpointLister(vpa_clientset.NewForConfigOrDie(config), make(chan struct{}), commonFlag.VpaObjectNamespace),
		ClusterState:        clusterState,
		SelectorFetcher:     target.NewVpaTargetSelectorFetcher(config, kubeClient, factory),
		MemorySaveMode:      *memorySaver,
		ControllerFetcher:   controllerFetcher,
		RecommenderName:     *recommenderName,
		IgnoredNamespaces:   ignoredNamespaces,
		VpaObjectNamespace:  commonFlag.VpaObjectNamespace,
	}.Make()
	controllerFetcher.Start(ctx, scaleCacheLoopPeriod)

	recommender := routines.RecommenderFactory{
		ClusterState:                 clusterState,
		ClusterStateFeeder:           clusterStateFeeder,
		ControllerFetcher:            controllerFetcher,
		CheckpointWriter:             checkpoint.NewCheckpointWriter(clusterState, vpa_clientset.NewForConfigOrDie(config).AutoscalingV1()),
		VpaClient:                    vpa_clientset.NewForConfigOrDie(config).AutoscalingV1(),
		PodResourceRecommender:       logic.CreatePodResourceRecommender(),
		RecommendationPostProcessors: postProcessors,
		CheckpointsGCInterval:        *checkpointsGCInterval,
		UseCheckpoints:               useCheckpoints,
		UpdateWorkerCount:            *updateWorkerCount,
	}.Make()

	promQueryTimeout, err := time.ParseDuration(*queryTimeout)
	if err != nil {
		klog.ErrorS(err, "Could not parse --prometheus-query-timeout as a time.Duration")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if useCheckpoints {
		recommender.GetClusterStateFeeder().InitFromCheckpoints(ctx)
	} else {
		config := history.PrometheusHistoryProviderConfig{
			Address:                *prometheusAddress,
			Insecure:               *prometheusInsecure,
			QueryTimeout:           promQueryTimeout,
			HistoryLength:          *historyLength,
			HistoryResolution:      *historyResolution,
			PodLabelPrefix:         *podLabelPrefix,
			PodLabelsMetricName:    *podLabelsMetricName,
			PodNamespaceLabel:      *podNamespaceLabel,
			PodNameLabel:           *podNameLabel,
			CtrNamespaceLabel:      *ctrNamespaceLabel,
			CtrPodNameLabel:        *ctrPodNameLabel,
			CtrNameLabel:           *ctrNameLabel,
			CadvisorMetricsJobName: *prometheusJobName,
			Namespace:              commonFlag.VpaObjectNamespace,
			Authentication: history.PrometheusCredentials{
				BearerToken: *prometheusBearerToken,
				Username:    *username,
				Password:    *password,
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

	ticker := time.Tick(*metricsFetcherInterval)
	for range ticker {
		recommender.RunOnce()
		healthCheck.UpdateLastActivity()
	}
}

func initGlobalMaxAllowed() apiv1.ResourceList {
	result := make(apiv1.ResourceList)
	if !maxAllowedCPU.IsZero() {
		result[apiv1.ResourceCPU] = maxAllowedCPU.Quantity
	}
	if !maxAllowedMemory.IsZero() {
		result[apiv1.ResourceMemory] = maxAllowedMemory.Quantity
	}

	return result
}
