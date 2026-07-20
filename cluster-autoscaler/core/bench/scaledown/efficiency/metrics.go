package efficiency

import (
	"fmt"
	"math"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
)

// metricReportingConfig holds level of reporting for scaleDownEfficiencyMetric.
type metricReportingConfig struct {
	ReportNodeLevel    bool
	ReportClusterLevel bool
}

// GetReportingConfig returns metricReportingConfig for scaleDownEfficiencyMetric.
func (c metricReportingConfig) GetReportingConfig() metricReportingConfig {
	return c
}

type metricOption func(*metricReportingConfig)

func withNodeReporting(enabled bool) metricOption {
	return func(c *metricReportingConfig) {
		c.ReportNodeLevel = enabled
	}
}

func withClusterReporting(enabled bool) metricOption {
	return func(c *metricReportingConfig) {
		c.ReportClusterLevel = enabled
	}
}

// defaultMetricReportingConfig sets reporting to false because it is verbose.
// Better to explicitly turn on if desired due to verbose output.
func defaultMetricReportingConfig() metricReportingConfig {
	return metricReportingConfig{
		ReportNodeLevel:    false,
		ReportClusterLevel: false,
	}
}

func applyMetricOptions(config *metricReportingConfig, opts ...metricOption) {
	*config = defaultMetricReportingConfig()
	for _, opt := range opts {
		opt(config)
	}
}

type scaleDownNodesReceiver interface {
	SetScaleDownNodes([]*status.ScaleDownNode)
}

// scaleDownEfficiencyMetric is the unified interface all metrics implement.
type scaleDownEfficiencyMetric interface {
	// Name returns name of the scaleDownEfficiencyMetric.
	Name() string
	// GetReportingConfig returns reporting config set for the instance of scaleDownEfficiencyMetric.
	GetReportingConfig() metricReportingConfig
	// GetNodeValues returns value of the scaleDownEfficiencyMetric for each node in the current snapshot.
	GetNodeValues() map[string]any
	// ComputeNodeLevel calculates the metric for each node.
	ComputeNodeLevel(nodeInfos []*framework.NodeInfo) error
	// ComputeClusterLevel aggregates the node-level data.
	ComputeClusterLevel() error
	// ReportNodeLevel reports per-node metrics summary.
	ReportNodeLevel(t testing.TB)
	// ReportClusterLevel reports summarized metric per whole cluster.
	ReportClusterLevel(t testing.TB)
	// ReportBenchmark calls b.ReportMetric for benchmarking output.
	ReportBenchmark(b *testing.B)
}

// metricsTracker orchestrates the computation, reporting, and aggregation of multiple scaleDownEfficiencyMetric.
type metricsTracker struct {
	metrics []scaleDownEfficiencyMetric
}

// NewMetricsTracker creates a new instance of metricsTracker with given list of scaleDownEfficiencyMetric.
func NewMetricsTracker(metrics ...scaleDownEfficiencyMetric) *metricsTracker {
	return &metricsTracker{
		metrics: metrics,
	}
}

// Compute computes both node-level and cluster-level metrics across a set of metrics.
func (mt *metricsTracker) Compute(snapshot clustersnapshot.ClusterSnapshot, scaledDownNodes []*status.ScaleDownNode, t testing.TB) {
	var err error
	nodeInfos, err := snapshot.ListNodeInfos()
	if err != nil {
		t.Errorf("failed to list node infos from snapshot: %v", err)
		return
	}

	for _, m := range mt.metrics {
		if receiverMetric, ok := m.(scaleDownNodesReceiver); ok {
			receiverMetric.SetScaleDownNodes(scaledDownNodes)
		}
		err := m.ComputeNodeLevel(nodeInfos)
		if err != nil {
			t.Errorf("failed to compute node metric %s: %v", m.Name(), err)
			continue
		}

		if m.GetNodeValues() == nil {
			t.Errorf("cannot compute cluster level for %s without prior node level computation", m.Name())
			continue
		}
		err = m.ComputeClusterLevel()
		if err != nil {
			t.Errorf("failed to compute cluster metric %s: %v", m.Name(), err)
			continue
		}
	}
}

// Report reports node-level and cluster-level metrics based on configuration of reporting level across a set of metrics.
// Each scaleDownEfficiencyMetric defines desired levels for itself in metricReportingConfig.
func (mt *metricsTracker) Report(stage string, t testing.TB) {
	reporting := false
	for i, m := range mt.metrics {
		nodeVals := m.GetNodeValues()
		config := m.GetReportingConfig()

		if !reporting && (config.ReportNodeLevel || config.ReportClusterLevel) {
			t.Logf("=== Metrics Report: %s ===", stage)
			reporting = true
		}

		if config.ReportNodeLevel && len(nodeVals) > 0 {
			t.Logf("--- %s (Node Level) ---", m.Name())
			m.ReportNodeLevel(t)
		}
		if config.ReportClusterLevel && len(nodeVals) > 0 {
			t.Logf("--- %s (Cluster Level) ---", m.Name())
			m.ReportClusterLevel(t)
		}
		if i == len(mt.metrics)-1 && reporting {
			t.Logf("======================================")
		}
	}

}

// ReportBenchmarks registers the raw metric with the benchmark runner across a set of configured metrics.
func (mt *metricsTracker) ReportBenchmarks(b *testing.B) {
	for _, m := range mt.metrics {
		m.ReportBenchmark(b)
	}
}

// nodeCountMetric counts number of nodes in the cluster snapshot.
type nodeCountMetric struct {
	metricReportingConfig
	nodeValues       map[string]any
	initialSet       bool
	initialNodeCount int
	nodeCount        int
}

func NewNodeCountMetric(opts ...metricOption) *nodeCountMetric {
	m := &nodeCountMetric{
		nodeValues: make(map[string]any),
	}
	applyMetricOptions(&m.metricReportingConfig, opts...)
	return m
}

func (m *nodeCountMetric) GetNodeValues() map[string]any {
	return m.nodeValues
}

func (m *nodeCountMetric) Name() string {
	return "node_count"
}

func (m *nodeCountMetric) ComputeNodeLevel(nodeInfos []*framework.NodeInfo) error {
	m.nodeValues = make(map[string]any)
	if len(nodeInfos) == 0 {
		return nil
	}
	for _, nodeInfo := range nodeInfos {
		if nodeInfo != nil && nodeInfo.Node() != nil {
			m.nodeValues[nodeInfo.Node().Name] = 1
		}
	}
	return nil
}

func (m *nodeCountMetric) ComputeClusterLevel() error {
	m.nodeCount = len(m.nodeValues)
	if !m.initialSet {
		m.initialNodeCount = m.nodeCount
		m.initialSet = true
	}
	return nil
}

func (m *nodeCountMetric) ReportNodeLevel(t testing.TB) {
}

func (m *nodeCountMetric) ReportClusterLevel(t testing.TB) {
	t.Logf("Total nodes: %d", m.nodeCount)
}

func (m *nodeCountMetric) ReportBenchmark(b *testing.B) {
	nodesRemoved := m.initialNodeCount - m.nodeCount
	b.ReportMetric(float64(m.initialNodeCount), "node_count_init")
	b.ReportMetric(float64(m.nodeCount), "node_count_final")
	b.ReportMetric(float64(nodesRemoved), "node_count_removed")
}
