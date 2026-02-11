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

package config

import (
	"flag"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// RecommenderConfig holds all configuration for the recommender component
type RecommenderConfig struct {
	// Common flags
	CommonFlags *common.CommonFlags

	// Recommender-specific flags
	RecommenderName         string
	MetricsFetcherInterval  time.Duration
	CheckpointsGCInterval   time.Duration
	CheckpointsWriteTimeout time.Duration
	Address                 string
	Storage                 string
	MemorySaver             bool
	UpdateWorkerCount       int
	MinCheckpointsPerRun    int

	// Recommendation configuration
	SafetyMarginFraction       float64
	PodMinCPUMillicores        float64
	PodMinMemoryMb             float64
	TargetCPUPercentile        float64
	LowerBoundCPUPercentile    float64
	UpperBoundCPUPercentile    float64
	ConfidenceIntervalCPU      time.Duration
	TargetMemoryPercentile     float64
	LowerBoundMemoryPercentile float64
	UpperBoundMemoryPercentile float64
	ConfidenceIntervalMemory   time.Duration
	HumanizeMemory             bool
	RoundCPUMillicores         int
	RoundMemoryBytes           int

	// Prometheus history provider configuration
	PrometheusAddress         string
	PrometheusInsecure        bool
	PrometheusJobName         string
	HistoryLength             string
	HistoryResolution         string
	QueryTimeout              string
	PodLabelPrefix            string
	PodLabelsMetricName       string
	PodNamespaceLabel         string
	PodNameLabel              string
	CtrNamespaceLabel         string
	CtrPodNameLabel           string
	CtrNameLabel              string
	Username                  string
	Password                  string
	PrometheusBearerToken     string
	PrometheusBearerTokenFile string

	// External metrics provider configuration
	UseExternalMetrics   bool
	ExternalCpuMetric    string
	ExternalMemoryMetric string

	// Aggregation configuration
	MemoryAggregationInterval      time.Duration
	MemoryAggregationIntervalCount int64
	MemoryHistogramDecayHalfLife   time.Duration
	CpuHistogramDecayHalfLife      time.Duration
	OOMBumpUpRatio                 float64
	OOMMinBumpUp                   float64

	// Post processors configuration
	PostProcessorCPUasInteger bool
	MaxAllowedCPU             resource.QuantityValue
	MaxAllowedMemory          resource.QuantityValue
}

// DefaultRecommenderConfig returns a RecommenderConfig with default values
func DefaultRecommenderConfig() *RecommenderConfig {
	return &RecommenderConfig{
		CommonFlags: common.DefaultCommonConfig(),

		// Recommender-specific flags
		RecommenderName:         input.DefaultRecommenderName,
		MetricsFetcherInterval:  1 * time.Minute,
		CheckpointsGCInterval:   10 * time.Minute,
		CheckpointsWriteTimeout: time.Minute,
		Address:                 ":8942",
		Storage:                 "",
		MemorySaver:             false,
		UpdateWorkerCount:       10,
		MinCheckpointsPerRun:    10,

		// Recommendation configuration
		SafetyMarginFraction:       0.15,
		PodMinCPUMillicores:        25,
		PodMinMemoryMb:             250,
		TargetCPUPercentile:        0.9,
		LowerBoundCPUPercentile:    0.5,
		UpperBoundCPUPercentile:    0.95,
		ConfidenceIntervalCPU:      24 * time.Hour,
		TargetMemoryPercentile:     0.9,
		LowerBoundMemoryPercentile: 0.5,
		UpperBoundMemoryPercentile: 0.95,
		ConfidenceIntervalMemory:   24 * time.Hour,
		HumanizeMemory:             false,
		RoundCPUMillicores:         1,
		RoundMemoryBytes:           1,

		// Prometheus history provider flags
		PrometheusAddress:         "http://prometheus.monitoring.svc",
		PrometheusInsecure:        false,
		PrometheusJobName:         "kubernetes-cadvisor",
		HistoryLength:             "8d",
		HistoryResolution:         "1h",
		QueryTimeout:              "5m",
		PodLabelPrefix:            "pod_label_",
		PodLabelsMetricName:       "up{job=\"kubernetes-pods\"}",
		PodNamespaceLabel:         "kubernetes_namespace",
		PodNameLabel:              "kubernetes_pod_name",
		CtrNamespaceLabel:         "namespace",
		CtrPodNameLabel:           "pod_name",
		CtrNameLabel:              "name",
		Username:                  "",
		Password:                  "",
		PrometheusBearerToken:     "",
		PrometheusBearerTokenFile: "",

		// External metrics provider flags
		UseExternalMetrics:   false,
		ExternalCpuMetric:    "",
		ExternalMemoryMetric: "",

		// Aggregation configuration flags
		MemoryAggregationInterval:      model.DefaultMemoryAggregationInterval,
		MemoryAggregationIntervalCount: model.DefaultMemoryAggregationIntervalCount,
		MemoryHistogramDecayHalfLife:   model.DefaultMemoryHistogramDecayHalfLife,
		CpuHistogramDecayHalfLife:      model.DefaultCPUHistogramDecayHalfLife,
		OOMBumpUpRatio:                 model.DefaultOOMBumpUpRatio,
		OOMMinBumpUp:                   model.DefaultOOMMinBumpUp,

		// Post processors flags
		PostProcessorCPUasInteger: false,
		MaxAllowedCPU:             resource.QuantityValue{},
		MaxAllowedMemory:          resource.QuantityValue{},
	}
}

// InitRecommenderFlags initializes flags for the recommender component
func InitRecommenderFlags() *RecommenderConfig {
	config := DefaultRecommenderConfig()
	config.CommonFlags = common.InitCommonFlags()

	flag.StringVar(&config.RecommenderName, "recommender-name", config.RecommenderName, "Set the recommender name. Recommender will generate recommendations for VPAs that configure the same recommender name. If the recommender name is left as default it will also generate recommendations that don't explicitly specify recommender. You shouldn't run two recommenders with the same name in a cluster.")
	flag.DurationVar(&config.MetricsFetcherInterval, "recommender-interval", config.MetricsFetcherInterval, `How often metrics should be fetched`)
	flag.DurationVar(&config.CheckpointsGCInterval, "checkpoints-gc-interval", config.CheckpointsGCInterval, `How often orphaned checkpoints should be garbage collected`)
	flag.DurationVar(&config.CheckpointsWriteTimeout, "checkpoints-timeout", config.CheckpointsWriteTimeout, `Timeout for writing checkpoints since the start of the recommender's main loop`)
	flag.StringVar(&config.Address, "address", config.Address, "The address to expose Prometheus metrics.")
	flag.StringVar(&config.Storage, "storage", config.Storage, `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
	flag.BoolVar(&config.MemorySaver, "memory-saver", config.MemorySaver, `If true, only track pods which have an associated VPA`)
	flag.IntVar(&config.UpdateWorkerCount, "update-worker-count", config.UpdateWorkerCount, "Number of concurrent workers to update VPA recommendations and checkpoints. When increasing this setting, make sure the client-side rate limits ('kube-api-qps' and 'kube-api-burst') are either increased or turned off as well. Determines the minimum number of VPA checkpoints written per recommender loop.")
	// MinCheckpointsPerRun is deprecated but kept for warning/compatibility.
	flag.IntVar(&config.MinCheckpointsPerRun, "min-checkpoints", config.MinCheckpointsPerRun, "Minimum number of checkpoints to write per recommender's main loop. WARNING: this flag is deprecated and doesn't have any effect. It will be removed in a future release. Refer to update-worker-count to influence the minimum number of checkpoints written per loop.")

	// Recommendation configuration flags
	flag.Float64Var(&config.SafetyMarginFraction, "recommendation-margin-fraction", config.SafetyMarginFraction, `Fraction of usage added as the safety margin to the recommended request`)
	flag.Float64Var(&config.PodMinCPUMillicores, "pod-recommendation-min-cpu-millicores", config.PodMinCPUMillicores, `Minimum CPU recommendation for a pod`)
	flag.Float64Var(&config.PodMinMemoryMb, "pod-recommendation-min-memory-mb", config.PodMinMemoryMb, `Minimum memory recommendation for a pod`)
	flag.Float64Var(&config.TargetCPUPercentile, "target-cpu-percentile", config.TargetCPUPercentile, "CPU usage percentile that will be used as a base for CPU target recommendation. Doesn't affect CPU lower bound, CPU upper bound nor memory recommendations.")
	flag.Float64Var(&config.LowerBoundCPUPercentile, "recommendation-lower-bound-cpu-percentile", config.LowerBoundCPUPercentile, `CPU usage percentile that will be used for the lower bound on CPU recommendation.`)
	flag.Float64Var(&config.UpperBoundCPUPercentile, "recommendation-upper-bound-cpu-percentile", config.UpperBoundCPUPercentile, `CPU usage percentile that will be used for the upper bound on CPU recommendation.`)
	flag.DurationVar(&config.ConfidenceIntervalCPU, "confidence-interval-cpu", config.ConfidenceIntervalCPU, "The time interval used for computing the confidence multiplier for the CPU lower and upper bound. Default: 24h")
	flag.Float64Var(&config.TargetMemoryPercentile, "target-memory-percentile", config.TargetMemoryPercentile, "Memory usage percentile that will be used as a base for memory target recommendation. Doesn't affect memory lower bound nor memory upper bound.")
	flag.Float64Var(&config.LowerBoundMemoryPercentile, "recommendation-lower-bound-memory-percentile", config.LowerBoundMemoryPercentile, `Memory usage percentile that will be used for the lower bound on memory recommendation.`)
	flag.Float64Var(&config.UpperBoundMemoryPercentile, "recommendation-upper-bound-memory-percentile", config.UpperBoundMemoryPercentile, `Memory usage percentile that will be used for the upper bound on memory recommendation.`)
	flag.DurationVar(&config.ConfidenceIntervalMemory, "confidence-interval-memory", config.ConfidenceIntervalMemory, "The time interval used for computing the confidence multiplier for the memory lower and upper bound. Default: 24h")
	flag.BoolVar(&config.HumanizeMemory, "humanize-memory", config.HumanizeMemory, "DEPRECATED: Convert memory values in recommendations to the highest appropriate SI unit with up to 2 decimal places for better readability. This flag is deprecated and will be removed in a future version. Use --round-memory-bytes instead.")
	flag.IntVar(&config.RoundCPUMillicores, "round-cpu-millicores", config.RoundCPUMillicores, `CPU recommendation rounding factor in millicores. The CPU value will always be rounded up to the nearest multiple of this factor.`)
	flag.IntVar(&config.RoundMemoryBytes, "round-memory-bytes", config.RoundMemoryBytes, `Memory recommendation rounding factor in bytes. The Memory value will always be rounded up to the nearest multiple of this factor.`)

	// Prometheus history provider flags
	flag.StringVar(&config.PrometheusAddress, "prometheus-address", config.PrometheusAddress, `Where to reach for Prometheus metrics`)
	flag.BoolVar(&config.PrometheusInsecure, "prometheus-insecure", config.PrometheusInsecure, `Skip tls verify if https is used in the prometheus-address`)
	flag.StringVar(&config.PrometheusJobName, "prometheus-cadvisor-job-name", config.PrometheusJobName, `Name of the prometheus job name which scrapes the cAdvisor metrics`)
	flag.StringVar(&config.HistoryLength, "history-length", config.HistoryLength, `How much time back prometheus have to be queried to get historical metrics`)
	flag.StringVar(&config.HistoryResolution, "history-resolution", config.HistoryResolution, `Resolution at which Prometheus is queried for historical metrics`)
	flag.StringVar(&config.QueryTimeout, "prometheus-query-timeout", config.QueryTimeout, `How long to wait before killing long queries`)
	flag.StringVar(&config.PodLabelPrefix, "pod-label-prefix", config.PodLabelPrefix, `Which prefix to look for pod labels in metrics`)
	flag.StringVar(&config.PodLabelsMetricName, "metric-for-pod-labels", config.PodLabelsMetricName, `Which metric to look for pod labels in metrics`)
	flag.StringVar(&config.PodNamespaceLabel, "pod-namespace-label", config.PodNamespaceLabel, `Label name to look for pod namespaces`)
	flag.StringVar(&config.PodNameLabel, "pod-name-label", config.PodNameLabel, `Label name to look for pod names`)
	flag.StringVar(&config.CtrNamespaceLabel, "container-namespace-label", config.CtrNamespaceLabel, `Label name to look for container namespaces`)
	flag.StringVar(&config.CtrPodNameLabel, "container-pod-name-label", config.CtrPodNameLabel, `Label name to look for container pod names`)
	flag.StringVar(&config.CtrNameLabel, "container-name-label", config.CtrNameLabel, `Label name to look for container names`)
	flag.StringVar(&config.Username, "username", config.Username, "The username used in the prometheus server basic auth. Can also be set via the PROMETHEUS_USERNAME environment variable")
	flag.StringVar(&config.Password, "password", config.Password, "The password used in the prometheus server basic auth. Can also be set via the PROMETHEUS_PASSWORD environment variable")
	flag.StringVar(&config.PrometheusBearerToken, "prometheus-bearer-token", config.PrometheusBearerToken, "The bearer token used in the Prometheus server bearer token auth")
	flag.StringVar(&config.PrometheusBearerTokenFile, "prometheus-bearer-token-file", config.PrometheusBearerTokenFile, "Path to the bearer token file used for authentication by the Prometheus server")

	// External metrics provider flags
	flag.BoolVar(&config.UseExternalMetrics, "use-external-metrics", config.UseExternalMetrics, "ALPHA.  Use an external metrics provider instead of metrics_server.")
	flag.StringVar(&config.ExternalCpuMetric, "external-metrics-cpu-metric", config.ExternalCpuMetric, "ALPHA.  Metric to use with external metrics provider for CPU usage.")
	flag.StringVar(&config.ExternalMemoryMetric, "external-metrics-memory-metric", config.ExternalMemoryMetric, "ALPHA.  Metric to use with external metrics provider for memory usage.")

	// Aggregation configuration flags
	flag.DurationVar(&config.MemoryAggregationInterval, "memory-aggregation-interval", config.MemoryAggregationInterval, `The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval)`)
	flag.Int64Var(&config.MemoryAggregationIntervalCount, "memory-aggregation-interval-count", config.MemoryAggregationIntervalCount, `The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count.`)
	flag.DurationVar(&config.MemoryHistogramDecayHalfLife, "memory-histogram-decay-half-life", config.MemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	flag.DurationVar(&config.CpuHistogramDecayHalfLife, "cpu-histogram-decay-half-life", config.CpuHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
	flag.Float64Var(&config.OOMBumpUpRatio, "oom-bump-up-ratio", config.OOMBumpUpRatio, `Default memory bump up ratio when OOM occurs. This value applies to all VPAs unless overridden in the VPA spec. Default is 1.2.`)
	flag.Float64Var(&config.OOMMinBumpUp, "oom-min-bump-up-bytes", config.OOMMinBumpUp, `Default minimal increase of memory (in bytes) when OOM occurs. This value applies to all VPAs unless overridden in the VPA spec. Default is 100 * 1024 * 1024 (100Mi).`)

	// Post processors flags
	// CPU as integer to benefit for CPU management Static Policy ( https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#static-policy )
	flag.BoolVar(&config.PostProcessorCPUasInteger, "cpu-integer-post-processor-enabled", config.PostProcessorCPUasInteger, "Enable the cpu-integer recommendation post processor. The post processor will round up CPU recommendations to a whole CPU for pods which were opted in by setting an appropriate label on VPA object (experimental)")
	flag.Var(&config.MaxAllowedCPU, "container-recommendation-max-allowed-cpu", "Maximum amount of CPU that will be recommended for a container. VerticalPodAutoscaler-level maximum allowed takes precedence over the global maximum allowed.")
	flag.Var(&config.MaxAllowedMemory, "container-recommendation-max-allowed-memory", "Maximum amount of memory that will be recommended for a container. VerticalPodAutoscaler-level maximum allowed takes precedence over the global maximum allowed.")

	return config
}

// ValidateRecommenderConfig performs validation of the recommender flags
func ValidateRecommenderConfig(config *RecommenderConfig) {
	common.ValidateCommonConfig(config.CommonFlags)

	if config.MinCheckpointsPerRun != 10 { // Default value is 10
		klog.InfoS("DEPRECATION WARNING: The 'min-checkpoints' flag is deprecated and has no effect. It will be removed in a future release.")
	}

	if config.PrometheusBearerToken != "" && config.PrometheusBearerTokenFile != "" && config.Username != "" {
		klog.ErrorS(nil, "--bearer-token, --bearer-token-file and --username are mutually exclusive and can't be set together.")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if config.PrometheusBearerTokenFile != "" {
		fileContent, err := os.ReadFile(config.PrometheusBearerTokenFile)
		if err != nil {
			klog.ErrorS(err, "Unable to read bearer token file", "filename", config.PrometheusBearerTokenFile)
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		config.PrometheusBearerTokenFile = strings.TrimSpace(string(fileContent))
	}
}
