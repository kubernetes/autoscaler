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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/server"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

var (
	recommenderName        = flag.String("recommender-name", input.DefaultRecommenderName, "Set the recommender name. Recommender will generate recommendations for VPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster.")
	metricsFetcherInterval = flag.Duration("recommender-interval", 1*time.Minute, `How often metrics should be fetched`)
	checkpointsGCInterval  = flag.Duration("checkpoints-gc-interval", 10*time.Minute, `How often orphaned checkpoints should be garbage collected`)
	prometheusAddress      = flag.String("prometheus-address", "", `Where to reach for Prometheus metrics`)
	prometheusJobName      = flag.String("prometheus-cadvisor-job-name", "kubernetes-cadvisor", `Name of the prometheus job name which scrapes the cAdvisor metrics`)
	address                = flag.String("address", ":8942", "The address to expose Prometheus metrics.")
	kubeconfig             = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	kubeApiQps             = flag.Float64("kube-api-qps", 5.0, `QPS limit when making requests to Kubernetes apiserver`)
	kubeApiBurst           = flag.Float64("kube-api-burst", 10.0, `QPS burst limit when making requests to Kubernetes apiserver`)
	enableProfiling        = flag.Bool("profiling", false, "Is debug/pprof endpoint enabled")

	storage = flag.String("storage", "", `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
	// prometheus history provider configs
	historyLength              = flag.String("history-length", "8d", `How much time back prometheus have to be queried to get historical metrics`)
	historyResolution          = flag.String("history-resolution", "1h", `Resolution at which Prometheus is queried for historical metrics`)
	queryTimeout               = flag.String("prometheus-query-timeout", "5m", `How long to wait before killing long queries`)
	podLabelPrefix             = flag.String("pod-label-prefix", "pod_label_", `Which prefix to look for pod labels in metrics`)
	podLabelsMetricName        = flag.String("metric-for-pod-labels", "up{job=\"kubernetes-pods\"}", `Which metric to look for pod labels in metrics`)
	podNamespaceLabel          = flag.String("pod-namespace-label", "kubernetes_namespace", `Label name to look for pod namespaces`)
	podNameLabel               = flag.String("pod-name-label", "kubernetes_pod_name", `Label name to look for pod names`)
	ctrNamespaceLabel          = flag.String("container-namespace-label", "namespace", `Label name to look for container namespaces`)
	ctrPodNameLabel            = flag.String("container-pod-name-label", "pod_name", `Label name to look for container pod names`)
	ctrNameLabel               = flag.String("container-name-label", "name", `Label name to look for container names`)
	vpaObjectNamespace         = flag.String("vpa-object-namespace", apiv1.NamespaceAll, "Namespace to search for VPA objects and pod stats. Empty means all namespaces will be used. Must not be used if ignored-vpa-object-namespaces is set.")
	ignoredVpaObjectNamespaces = flag.String("ignored-vpa-object-namespaces", "", "Comma separated list of namespaces to ignore. Must not be used if vpa-object-namespace is used.")
	username                   = flag.String("username", "", "The username used in the prometheus server basic auth")
	password                   = flag.String("password", "", "The password used in the prometheus server basic auth")
	memorySaver                = flag.Bool("memory-saver", false, `If true, only track pods which have an associated VPA`)
	// external metrics provider config
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

func main() {
	klog.InitFlags(nil)

	leaderElection := defaultLeaderElectionConfiguration()
	componentbaseoptions.BindLeaderElectionFlags(&leaderElection, pflag.CommandLine)

	kube_flag.InitFlags()
	klog.V(1).Infof("Vertical Pod Autoscaler %s Recommender: %v", common.VerticalPodAutoscalerVersion, *recommenderName)

	if len(*vpaObjectNamespace) > 0 && len(*ignoredVpaObjectNamespaces) > 0 {
		klog.Fatalf("--vpa-object-namespace and --ignored-vpa-object-namespaces are mutually exclusive and can't be set together.")
	}

	healthCheck := metrics.NewHealthCheck(*metricsFetcherInterval * 5)
	metrics_recommender.Register()
	metrics_quality.Register()
	server.Initialize(enableProfiling, healthCheck, address)

	if !leaderElection.LeaderElect {
		run(healthCheck)
	} else {
		id, err := os.Hostname()
		if err != nil {
			klog.Fatalf("Unable to get hostname: %v", err)
		}
		id = id + "_" + string(uuid.NewUUID())

		config := common.CreateKubeConfigOrDie(*kubeconfig, float32(*kubeApiQps), int(*kubeApiBurst))
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
					run(healthCheck)
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
		ResourceName:      "vpa-recommender",
		ResourceNamespace: metav1.NamespaceSystem,
	}
}

func run(healthCheck *metrics.HealthCheck) {
	config := common.CreateKubeConfigOrDie(*kubeconfig, float32(*kubeApiQps), int(*kubeApiBurst))
	kubeClient := kube_client.NewForConfigOrDie(config)
	clusterState := model.NewClusterState(aggregateContainerStateGCInterval)
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(*vpaObjectNamespace))
	controllerFetcher := controllerfetcher.NewControllerFetcher(config, kubeClient, factory, scaleCacheEntryFreshnessTime, scaleCacheEntryLifetime, scaleCacheEntryJitterFactor)
	podLister, oomObserver := input.NewPodListerAndOOMObserver(kubeClient, *vpaObjectNamespace)

	model.InitializeAggregationsConfig(model.NewAggregationsConfig(*memoryAggregationInterval, *memoryAggregationIntervalCount, *memoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife, *oomBumpUpRatio, *oomMinBumpUp))

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
		klog.V(1).Infof("Using External Metrics: %+v", externalClientOptions)
		source = input_metrics.NewExternalClient(config, clusterState, *externalClientOptions)
	} else {
		klog.V(1).Infof("Using Metrics Server.")
		source = input_metrics.NewPodMetricsesSource(resourceclient.NewForConfigOrDie(config))
	}

	ignoredNamespaces := strings.Split(*ignoredVpaObjectNamespaces, ",")

	clusterStateFeeder := input.ClusterStateFeederFactory{
		PodLister:           podLister,
		OOMObserver:         oomObserver,
		KubeClient:          kubeClient,
		MetricsClient:       input_metrics.NewMetricsClient(source, *vpaObjectNamespace, "default-metrics-client"),
		VpaCheckpointClient: vpa_clientset.NewForConfigOrDie(config).AutoscalingV1(),
		VpaCheckpointLister: vpa_api_util.NewVpaCheckpointLister(vpa_clientset.NewForConfigOrDie(config), make(chan struct{}), *vpaObjectNamespace),
		VpaLister:           vpa_api_util.NewVpasLister(vpa_clientset.NewForConfigOrDie(config), make(chan struct{}), *vpaObjectNamespace),
		ClusterState:        clusterState,
		SelectorFetcher:     target.NewVpaTargetSelectorFetcher(config, kubeClient, factory),
		MemorySaveMode:      *memorySaver,
		ControllerFetcher:   controllerFetcher,
		RecommenderName:     *recommenderName,
		IgnoredNamespaces:   ignoredNamespaces,
	}.Make()
	controllerFetcher.Start(context.Background(), scaleCacheLoopPeriod)

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
	}.Make()

	promQueryTimeout, err := time.ParseDuration(*queryTimeout)
	if err != nil {
		klog.Fatalf("Could not parse --prometheus-query-timeout as a time.Duration: %v", err)
	}

	if useCheckpoints {
		recommender.GetClusterStateFeeder().InitFromCheckpoints()
	} else {
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
			Namespace:              *vpaObjectNamespace,
			PrometheusBasicAuthTransport: history.PrometheusBasicAuthTransport{
				Username: *username,
				Password: *password,
			},
		}
		provider, err := history.NewPrometheusHistoryProvider(config)
		if err != nil {
			klog.Fatalf("Could not initialize history provider: %v", err)
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
