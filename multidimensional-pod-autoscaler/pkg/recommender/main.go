/*
Copyright 2024 The Kubernetes Authors.

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
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/checkpoint"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/routines"
	"k8s.io/metrics/pkg/client/custom_metrics"
	"k8s.io/metrics/pkg/client/external_metrics"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/common"

	mpa_clientset "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned"
	input_metrics "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/input/metrics"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/target"
	metrics_recommender "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/metrics/recommender"
	mpa_api_util "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/mpa"
	vpa_common "k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_quality "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/quality"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/server"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	kube_flag "k8s.io/component-base/cli/flag"
	componentbaseconfig "k8s.io/component-base/config"
	componentbaseoptions "k8s.io/component-base/config/options"
	klog "k8s.io/klog/v2"
	hpa_metrics "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

var (
	recommenderName            = flag.String("recommender-name", input.DefaultRecommenderName, "Set the recommender name. Recommender will generate recommendations for MPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster.")
	metricsFetcherInterval     = flag.Duration("recommender-interval", 1*time.Minute, `How often metrics should be fetched`)
	checkpointsGCInterval      = flag.Duration("checkpoints-gc-interval", 10*time.Minute, `How often orphaned checkpoints should be garbage collected`)
	address                    = flag.String("address", ":8942", "The address to expose Prometheus metrics.")
	storage                    = flag.String("storage", "", `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
	memorySaver                = flag.Bool("memory-saver", false, `If true, only track pods which have an associated MPA`)
	mpaObjectNamespace         = flag.String("mpa-object-namespace", apiv1.NamespaceAll, "Namespace to search for MPA objects and pod stats. Empty means all namespaces will be used.")
	ignoredMpaObjectNamespaces = flag.String("ignored-mpa-object-namespaces", "", "Comma separated list of namespaces to ignore when searching for MPA objects. Empty means no namespaces will be ignored.")
)

// prometheus history provider configs
var (
	prometheusAddress   = flag.String("prometheus-address", "", `Where to reach for Prometheus metrics`)
	prometheusJobName   = flag.String("prometheus-cadvisor-job-name", "kubernetes-cadvisor", `Name of the prometheus job name which scrapes the cAdvisor metrics`)
	historyLength       = flag.String("history-length", "8d", `How much time back prometheus have to be queried to get historical metrics`)
	historyResolution   = flag.String("history-resolution", "1h", `Resolution at which Prometheus is queried for historical metrics`)
	queryTimeout        = flag.String("prometheus-query-timeout", "5m", `How long to wait before killing long queries`)
	podLabelPrefix      = flag.String("pod-label-prefix", "pod_label_", `Which prefix to look for pod labels in metrics`)
	podLabelsMetricName = flag.String("metric-for-pod-labels", "up{job=\"kubernetes-pods\"}", `Which metric to look for pod labels in metrics`)
	podNamespaceLabel   = flag.String("pod-namespace-label", "kubernetes_namespace", `Label name to look for pod namespaces`)
	podNameLabel        = flag.String("pod-name-label", "kubernetes_pod_name", `Label name to look for pod names`)
	ctrNamespaceLabel   = flag.String("container-namespace-label", "namespace", `Label name to look for container namespaces`)
	ctrPodNameLabel     = flag.String("container-pod-name-label", "pod_name", `Label name to look for container pod names`)
	ctrNameLabel        = flag.String("container-name-label", "name", `Label name to look for container names`)
	username            = flag.String("username", "", "The username used in the prometheus server basic auth")
	password            = flag.String("password", "", "The password used in the prometheus server basic auth")
)

// External metrics provider flags
var (
	useExternalMetrics   = flag.Bool("use-external-metrics", false, "ALPHA.  Use an external metrics provider instead of metrics_server.")
	externalCpuMetric    = flag.String("external-metrics-cpu-metric", "", "ALPHA.  Metric to use with external metrics provider for CPU usage.")
	externalMemoryMetric = flag.String("external-metrics-memory-metric", "", "ALPHA.  Metric to use with external metrics provider for memory usage.")
)

// Aggregation configuration flags
var (
	memoryAggregationInterval      = flag.Duration("memory-aggregation-interval", vpa_model.DefaultMemoryAggregationInterval, `The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval)`)
	memoryAggregationIntervalCount = flag.Int64("memory-aggregation-interval-count", vpa_model.DefaultMemoryAggregationIntervalCount, `The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count.`)
	memoryHistogramDecayHalfLife   = flag.Duration("memory-histogram-decay-half-life", vpa_model.DefaultMemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	cpuHistogramDecayHalfLife      = flag.Duration("cpu-histogram-decay-half-life", vpa_model.DefaultCPUHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
	oomBumpUpRatio                 = flag.Float64("oom-bump-up-ratio", vpa_model.DefaultOOMBumpUpRatio, `Specifies the memory bump up ratio when OOM occurred.`)
	oomMinBumpUp                   = flag.Float64("oom-min-bump-up", vpa_model.DefaultOOMMinBumpUp, `Specifies the minimal increase of memory when OOM occurred in bytes..`)
)

// HPA-related flags
var (
	// horizontalPodAutoscalerSyncPeriod is the period for syncing the number of pods in MPA.
	hpaSyncPeriod = flag.Duration("hpa-sync-period", 15*time.Second, `The period for syncing the number of pods in horizontal pod autoscaler.`)
	// horizontalPodAutoscalerUpscaleForbiddenWindow is a period after which next upscale allowed.
	hpaUpscaleForbiddenWindow = flag.Duration("hpa-upscale-forbidden-window", 3*time.Minute, `The period after which next upscale allowed.`)
	// horizontalPodAutoscalerDownscaleForbiddenWindow is a period after which next downscale allowed.
	hpaDownscaleForbiddenWindow = flag.Duration("hpa-downscale-forbidden-window", 5*time.Minute, `The period after which next downscale allowed.`)
	// HorizontalPodAutoscalerDowncaleStabilizationWindow is a period for which autoscaler will look
	// backwards and not scale down below any recommendation it made during that period.
	hpaDownscaleStabilizationWindow = flag.Duration("hpa-downscale-stabilization-window", 5*time.Minute, `The period for which autoscaler will look backwards and not scale down below any recommendation it made during that period.`)
	// horizontalPodAutoscalerTolerance is the tolerance for when resource usage suggests upscaling/downscaling
	hpaTolerance = flag.Float64("hpa-tolerance", 0.1, `The tolerance for when resource usage suggests horizontally upscaling/downscaling.`)
	// HorizontalPodAutoscalerCPUInitializationPeriod is the period after pod start when CPU samples
	// might be skipped.
	hpaCPUInitializationPeriod = flag.Duration("hpa-cpu-initialization-period", 5*time.Minute, `The period after pod start when CPU samples might be skipped.`)
	// HorizontalPodAutoscalerInitialReadinessDelay is period after pod start during which readiness
	// changes are treated as readiness being set for the first time. The only effect of this is
	// that HPA will disregard CPU samples from unready pods that had last readiness change during
	// that period.
	hpaInitialReadinessDelay = flag.Duration("hpa-initial-readiness-delay", 30*time.Second, `The period after pod start during which readiness changes are treated as readiness being set for the first time.`)
	concurrentHPASyncs       = flag.Int64("concurrent-hpa-syncs", 5, `The number of horizontal pod autoscaler objects that are allowed to sync concurrently. Larger number = more responsive MPA objects processing, but more CPU (and network) load.`)
)

// Post processors flags
var (
	// CPU as integer to benefit for CPU management Static Policy ( https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#static-policy )
	postProcessorCPUasInteger = flag.Bool("cpu-integer-post-processor-enabled", false, "Enable the cpu-integer recommendation post processor. The post processor will round up CPU recommendations to a whole CPU for pods which were opted in by setting an appropriate label on VPA object (experimental)")
)

const (
	// aggregateContainerStateGCInterval defines how often expired AggregateContainerStates are garbage collected.
	aggregateContainerStateGCInterval               = 1 * time.Hour
	scaleCacheEntryLifetime           time.Duration = time.Hour
	scaleCacheEntryFreshnessTime      time.Duration = 10 * time.Minute
	scaleCacheEntryJitterFactor       float64       = 1.
	scaleCacheLoopPeriod                            = 7 * time.Second
	defaultResyncPeriod               time.Duration = 10 * time.Minute
	discoveryResetPeriod              time.Duration = 5 * time.Minute
)

func main() {
	commonFlags := vpa_common.InitCommonFlags()
	klog.InitFlags(nil)
	vpa_common.InitLoggingFlags()

	leaderElection := defaultLeaderElectionConfiguration()
	componentbaseoptions.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	kube_flag.InitFlags()
	klog.V(1).Infof("Multi-dimensional Pod Autoscaler %s Recommender: %v", common.MultidimPodAutoscalerVersion, recommenderName)

	if len(*mpaObjectNamespace) > 0 && len(*ignoredMpaObjectNamespaces) > 0 {
		klog.Fatalf("--mpa-object-namespace and --ignored-mpa-object-namespaces are mutually exclusive and can't be set together.")
	}

	healthCheck := metrics.NewHealthCheck(*metricsFetcherInterval * 5)
	metrics_recommender.Register()
	metrics_quality.Register()
	server.Initialize(&commonFlags.EnableProfiling, healthCheck, address)

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
		LeaderElect:   false,
		LeaseDuration: metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline: metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:   metav1.Duration{Duration: defaultRetryPeriod},
		ResourceLock:  resourcelock.LeasesResourceLock,
		// This was changed from "vpa-recommender" to avoid conflicts with managed VPA deployments.
		ResourceName:      "mpa-recommender-lease",
		ResourceNamespace: metav1.NamespaceSystem,
	}
}

func run(healthCheck *metrics.HealthCheck, commonFlag *vpa_common.CommonFlags) {
	config := common.CreateKubeConfigOrDie(commonFlag.KubeConfig, float32(commonFlag.KubeApiQps), int(commonFlag.KubeApiBurst))
	kubeClient := kube_client.NewForConfigOrDie(config)
	clusterState := model.NewClusterState(aggregateContainerStateGCInterval)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(*ignoredMpaObjectNamespaces))
	controllerFetcher := controllerfetcher.NewControllerFetcher(config, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	podLister, oomObserver := input.NewPodListerAndOOMObserver(kubeClient, commonFlag.IgnoredVpaObjectNamespaces)

	vpa_model.InitializeAggregationsConfig(vpa_model.NewAggregationsConfig(*memoryAggregationInterval, *memoryAggregationIntervalCount, *memoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife, *oomBumpUpRatio, *oomMinBumpUp))

	useCheckpoints := *storage != "prometheus"

	var postProcessors []routines.RecommendationPostProcessor
	if *postProcessorCPUasInteger {
		postProcessors = append(postProcessors, &routines.IntegerCPUPostProcessor{})
	}

	// CappingPostProcessor, should always come in the last position for post-processing
	postProcessors = append(postProcessors, &routines.CappingPostProcessor{})
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

	ignoredNamespaces := strings.Split(*ignoredMpaObjectNamespaces, ",")

	clusterStateFeeder := input.ClusterStateFeederFactory{
		PodLister:           podLister,
		OOMObserver:         oomObserver,
		KubeClient:          kubeClient,
		MetricsClient:       input_metrics.NewMetricsClient(source, *mpaObjectNamespace, "default-metrics-client"),
		MpaCheckpointClient: mpa_clientset.NewForConfigOrDie(config).AutoscalingV1alpha1(),
		MpaLister:           mpa_api_util.NewMpasLister(mpa_clientset.NewForConfigOrDie(config), make(chan struct{}), *mpaObjectNamespace),
		ClusterState:        clusterState,
		SelectorFetcher:     target.NewMpaTargetSelectorFetcher(config, kubeClient, factory),
		MemorySaveMode:      *memorySaver,
		ControllerFetcher:   controllerFetcher,
		RecommenderName:     *recommenderName,
		IgnoredNamespaces:   ignoredNamespaces,
	}.Make()
	controllerFetcher.Start(context.Background(), scaleCacheLoopPeriod)

	// For HPA.
	// Use a discovery client capable of being refreshed.
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		klog.Fatalf("Could not create discoveryClient: %v", err)
	}
	cachedDiscoveryClient := cacheddiscovery.NewMemCacheClient(discoveryClient)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscoveryClient)
	go wait.Until(func() {
		mapper.Reset()
	}, discoveryResetPeriod, make(chan struct{}))
	mpaClient := mpa_clientset.NewForConfigOrDie(config)
	apiVersionsGetter := custom_metrics.NewAvailableAPIsGetter(mpaClient.Discovery())
	ctx := context.Background() // TODO: Add a deadline to this ctx?
	// invalidate the discovery information roughly once per resync interval our API
	// information is *at most* two resync intervals old.
	go custom_metrics.PeriodicallyInvalidate(
		apiVersionsGetter,
		*hpaSyncPeriod,
		ctx.Done())

	metricsClient := hpa_metrics.NewRESTMetricsClient(
		resourceclient.NewForConfigOrDie(config),
		custom_metrics.NewForConfig(config, mapper, apiVersionsGetter),
		external_metrics.NewForConfigOrDie(config),
	)

	recommender := routines.RecommenderFactory{
		ClusterState:                 clusterState,
		ClusterStateFeeder:           clusterStateFeeder,
		ControllerFetcher:            controllerFetcher,
		CheckpointWriter:             checkpoint.NewCheckpointWriter(clusterState, mpa_clientset.NewForConfigOrDie(config).AutoscalingV1alpha1()),
		MpaClient:                    mpa_clientset.NewForConfigOrDie(config).AutoscalingV1alpha1(),
		SelectorFetcher:              target.NewMpaTargetSelectorFetcher(config, kubeClient, factory),
		PodResourceRecommender:       logic.CreatePodResourceRecommender(),
		RecommendationPostProcessors: postProcessors,
		CheckpointsGCInterval:        *checkpointsGCInterval,
		UseCheckpoints:               useCheckpoints,

		// HPA-related flags
		EvtNamespacer:                 kubeClient.CoreV1(),
		PodInformer:                   factory.Core().V1().Pods(),
		MetricsClient:                 metricsClient,
		ResyncPeriod:                  *hpaSyncPeriod,
		DownscaleStabilisationWindow:  *hpaDownscaleStabilizationWindow,
		Tolerance:                     *hpaTolerance,
		CpuInitializationPeriod:       *hpaCPUInitializationPeriod,
		DelayOfInitialReadinessStatus: *hpaInitialReadinessDelay,
	}.Make()

	klog.Infof("MPA Recommender created!")

	promQueryTimeout, err := time.ParseDuration(*queryTimeout)
	if err != nil {
		klog.Fatalf("Could not parse --prometheus-query-timeout as a time.Duration: %v", err)
	}

	if useCheckpoints {
		recommender.GetClusterStateFeeder().InitFromCheckpoints()
	} else {
		klog.Info("Creating Prometheus history provider...")
		config := history.PrometheusHistoryProviderConfig{
			Address:                *prometheusAddress,
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
			Namespace:              *mpaObjectNamespace,
			PrometheusBasicAuthTransport: history.PrometheusBasicAuthTransport{
				Username: *username,
				Password: *password,
			},
		}
		provider, err := history.NewPrometheusHistoryProvider(config)
		if err != nil {
			klog.Fatalf("Could not initialize history provider: %v", err)
		}
		klog.Info("History provider initialized!")
		recommender.GetClusterStateFeeder().InitFromHistoryProvider(provider)
		klog.Info("Recommender initialized!")
	}

	ticker := time.Tick(*metricsFetcherInterval)
	klog.Info("Start running MPA Recommender...")
	var vpaOrHpa = "vpa"
	for range ticker {
		recommender.RunOnce(int(*concurrentHPASyncs), vpaOrHpa)
		healthCheck.UpdateLastActivity()
		klog.Info("Health check completed.")
		if vpaOrHpa == "vpa" {
			vpaOrHpa = "hpa"
		} else if vpaOrHpa == "hpa" {
			vpaOrHpa = "vpa"
		}
	}
}
