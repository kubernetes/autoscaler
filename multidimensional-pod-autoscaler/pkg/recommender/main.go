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
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/input"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/common"
	mpa_clientset "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/routines"
	metrics_recommender "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/metrics/recommender"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/history"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
	metrics_quality "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/quality"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	kube_flag "k8s.io/component-base/cli/flag"
	klog "k8s.io/klog/v2"

	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	hpa_metrics "k8s.io/kubernetes/pkg/controller/podautoscaler/metrics"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/custom_metrics"
	"k8s.io/metrics/pkg/client/external_metrics"
)

var (
	recommenderName        = flag.String("recommender-name", input.DefaultRecommenderName, "Set the recommender name. Recommender will generate recommendations for MPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster.")
	metricsFetcherInterval = flag.Duration("recommender-interval", 1*time.Minute, `How often metrics should be fetched`)
	checkpointsGCInterval  = flag.Duration("checkpoints-gc-interval", 10*time.Minute, `How often orphaned checkpoints should be garbage collected`)
	prometheusAddress      = flag.String("prometheus-address", "", `Where to reach for Prometheus metrics`)
	prometheusJobName      = flag.String("prometheus-cadvisor-job-name", "kubernetes-cadvisor", `Name of the prometheus job name which scrapes the cAdvisor metrics`)
	address                = flag.String("address", ":8942", "The address to expose Prometheus metrics.")
	kubeconfig             = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	kubeApiQps             = flag.Float64("kube-api-qps", 5.0, `QPS limit when making requests to Kubernetes apiserver`)
	kubeApiBurst           = flag.Float64("kube-api-burst", 10.0, `QPS burst limit when making requests to Kubernetes apiserver`)

	storage = flag.String("storage", "", `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
	// prometheus history provider configs
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
	mpaObjectNamespace  = flag.String("mpa-object-namespace", apiv1.NamespaceAll, "Namespace to search for MPA objects and pod stats. Empty means all namespaces will be used.")
)

// Aggregation configuration flags
var (
	memoryAggregationInterval      = flag.Duration("memory-aggregation-interval", vpa_model.DefaultMemoryAggregationInterval, `The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval)`)
	memoryAggregationIntervalCount = flag.Int64("memory-aggregation-interval-count", vpa_model.DefaultMemoryAggregationIntervalCount, `The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count.`)
	memoryHistogramDecayHalfLife   = flag.Duration("memory-histogram-decay-half-life", vpa_model.DefaultMemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	cpuHistogramDecayHalfLife      = flag.Duration("cpu-histogram-decay-half-life", vpa_model.DefaultCPUHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
)

// HPA-related flags
var (
	// horizontalPodAutoscalerSyncPeriod is the period for syncing the number of pods in MPA.
	hpaSyncPeriod                   = flag.Duration("hpa-sync-period", 15 * time.Second, `The period for syncing the number of pods in horizontal pod autoscaler.`)
	// horizontalPodAutoscalerUpscaleForbiddenWindow is a period after which next upscale allowed.
	hpaUpscaleForbiddenWindow       = flag.Duration("hpa-upscale-forbidden-window", 3 * time.Minute, `The period after which next upscale allowed.`)
	// horizontalPodAutoscalerDownscaleForbiddenWindow is a period after which next downscale allowed.
	hpaDownscaleForbiddenWindow     = flag.Duration("hpa-downscale-forbidden-window", 5 * time.Minute, `The period after which next downscale allowed.`)
	// HorizontalPodAutoscalerDowncaleStabilizationWindow is a period for which autoscaler will look
	// backwards and not scale down below any recommendation it made during that period.
	hpaDownscaleStabilizationWindow = flag.Duration("hpa-downscale-stabilization-window", 5 * time.Minute, `The period for which autoscaler will look backwards and not scale down below any recommendation it made during that period.`)
	// horizontalPodAutoscalerTolerance is the tolerance for when resource usage suggests upscaling/downscaling
	hpaTolerance                    = flag.Float64("hpa-tolerance", 0.1, `The tolerance for when resource usage suggests horizontally upscaling/downscaling.`)
	// HorizontalPodAutoscalerCPUInitializationPeriod is the period after pod start when CPU samples
	// might be skipped.
	hpaCPUInitializationPeriod      = flag.Duration("hpa-cpu-initialization-period", 5 * time.Minute, `The period after pod start when CPU samples might be skipped.`)
	// HorizontalPodAutoscalerInitialReadinessDelay is period after pod start during which readiness
	// changes are treated as readiness being set for the first time. The only effect of this is
	// that HPA will disregard CPU samples from unready pods that had last readiness change during
	// that period.
	hpaInitialReadinessDelay        = flag.Duration("hpa-initial-readiness-delay", 30 * time.Second, `The period after pod start during which readiness changes are treated as readiness being set for the first time.`)
	concurrentHPASyncs = flag.Int64("concurrent-hpa-syncs", 5, `The number of horizontal pod autoscaler objects that are allowed to sync concurrently. Larger number = more responsive MPA objects processing, but more CPU (and network) load.`)
)

const (
	discoveryResetPeriod time.Duration = 5 * time.Minute
)

func main() {
	klog.InitFlags(nil)
	kube_flag.InitFlags()
	klog.V(1).Infof("Multi-dimensional Pod Autoscaler %s Recommender: %v", common.MultidimPodAutoscalerVersion, recommenderName)

	config := common.CreateKubeConfigOrDie(*kubeconfig, float32(*kubeApiQps), int(*kubeApiBurst))

	vpa_model.InitializeAggregationsConfig(vpa_model.NewAggregationsConfig(*memoryAggregationInterval, *memoryAggregationIntervalCount, *memoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife))

	healthCheck := metrics.NewHealthCheck(*metricsFetcherInterval*5, true)
	metrics.Initialize(*address, healthCheck)
	metrics_recommender.Register()
	metrics_quality.Register()

	useCheckpoints := *storage != "prometheus"

	// For HPA.
	kubeClient := kube_client.NewForConfigOrDie(config)
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
	ctx := context.Background()  // TODO: Add a deadline to this ctx?
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
	
	recommender := routines.NewRecommender(config, *checkpointsGCInterval, useCheckpoints, *mpaObjectNamespace, *recommenderName, kubeClient.CoreV1(), metricsClient, *hpaSyncPeriod, *hpaDownscaleStabilizationWindow, *hpaTolerance, *hpaCPUInitializationPeriod, *hpaInitialReadinessDelay)
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
		if (vpaOrHpa == "vpa") {
			vpaOrHpa = "hpa"
		} else if (vpaOrHpa == "hpa") {
			vpaOrHpa = "vpa"
		}
	}
}
