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

package app

import (
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/common"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/routines"
)

// RecommenderConfig holds all configuration for the recommender component
type RecommenderConfig struct {
	// Common flags
	CommonFlags *common.CommonFlags

	// Recommender-specific flags
	RecommenderName        string
	MetricsFetcherInterval time.Duration
	CheckpointsGCInterval  time.Duration
	Address                string
	Storage                string
	MemorySaver            bool
	UpdateWorkerCount      int

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
		RecommenderName:        input.DefaultRecommenderName,
		MetricsFetcherInterval: 1 * time.Minute,
		CheckpointsGCInterval:  10 * time.Minute,
		Address:                ":8942",
		Storage:                "",
		MemorySaver:            false,
		UpdateWorkerCount:      10,

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

// ValidateRecommenderConfig performs validation of the recommender flags
func ValidateRecommenderConfig(config *RecommenderConfig) {
	if *routines.MinCheckpointsPerRun != 10 { // Default value is 10
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
