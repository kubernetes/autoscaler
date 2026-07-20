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

type resourcePair struct {
	Allocatable float64
	Requested   float64
}

type nodeResources struct {
	CPU resourcePair
	Mem resourcePair
}

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

// resourceUtilizationMetric computes resource utilization ratio for cpu and memory.
type resourceUtilizationMetric struct {
	metricReportingConfig
	nodeValues          map[string]any
	initialSet          bool
	initialCpuRatio     float64
	initialMemRatio     float64
	totalCPU            float64
	totalCPUAllocatable float64
	totalMem            float64
	totalMemAllocatable float64
	cpuRatio            float64
	memRatio            float64
	driver              string
}

func NewResourceUtilizationMetric(opts ...metricOption) *resourceUtilizationMetric {
	m := &resourceUtilizationMetric{
		nodeValues: make(map[string]any),
	}
	applyMetricOptions(&m.metricReportingConfig, opts...)
	return m
}

func (m *resourceUtilizationMetric) GetNodeValues() map[string]any {
	return m.nodeValues
}

func (m *resourceUtilizationMetric) Name() string {
	return "resource_utilization"
}

func calculateAllocatableRequested(currentTime time.Time, nodeInfo *framework.NodeInfo, nodeName string) (float64, float64, float64, float64, error) {
	cpuUtil, err := utilization.CalculateUtilizationOfResource(nodeInfo, apiv1.ResourceCPU, true, true, currentTime)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error calculating %s utilization for node %s, err: %v", apiv1.ResourceCPU, nodeName, err)
	}
	memUtil, err := utilization.CalculateUtilizationOfResource(nodeInfo, apiv1.ResourceMemory, true, true, currentTime)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("error calculating %s utilization for node %s, err: %v", apiv1.ResourceMemory, nodeName, err)
	}

	nodeAllocatableCPU, foundCPU := nodeInfo.Node().Status.Allocatable[apiv1.ResourceCPU]
	if !foundCPU {
		return 0, 0, 0, 0, fmt.Errorf("failed to get %v from %s", apiv1.ResourceCPU, nodeName)
	}
	nodeAllocatableMem, foundMem := nodeInfo.Node().Status.Allocatable[apiv1.ResourceMemory]
	if !foundMem {
		return 0, 0, 0, 0, fmt.Errorf("failed to get %v from %s", apiv1.ResourceMemory, nodeName)
	}
	return float64(nodeAllocatableCPU.MilliValue()), float64(nodeAllocatableMem.MilliValue()), cpuUtil, memUtil, nil
}

// ComputeNodeLevel TODO: no GPU yet or other factors influencing node efficiency
func (m *resourceUtilizationMetric) ComputeNodeLevel(nodeInfos []*framework.NodeInfo) error {
	currentTime := time.Now()
	m.nodeValues = make(map[string]any)

	if len(nodeInfos) == 0 {
		return nil
	}

	for _, nodeInfo := range nodeInfos {
		if nodeInfo == nil || nodeInfo.Node() == nil {
			continue
		}
		nodeName := nodeInfo.Node().Name
		cAlloc, mAlloc, cUtil, mUtil, err := calculateAllocatableRequested(currentTime, nodeInfo, nodeName)
		if err != nil {
			return err
		}

		m.nodeValues[nodeName] = nodeResources{
			CPU: resourcePair{Allocatable: cAlloc, Requested: cUtil * cAlloc},
			Mem: resourcePair{Allocatable: mAlloc, Requested: mUtil * mAlloc},
		}
	}
	return nil
}

func (m *resourceUtilizationMetric) ComputeClusterLevel() error {
	m.totalCPU, m.totalCPUAllocatable, m.totalMem, m.totalMemAllocatable = 0, 0, 0, 0

	for _, val := range m.nodeValues {
		resources, ok := val.(nodeResources)
		if !ok {
			return fmt.Errorf("unexpected type in m.nodeValues")
		}
		m.totalCPU += resources.CPU.Requested
		m.totalCPUAllocatable += resources.CPU.Allocatable
		m.totalMem += resources.Mem.Requested
		m.totalMemAllocatable += resources.Mem.Allocatable
	}

	m.memRatio, m.cpuRatio = 0, 0
	if m.totalMemAllocatable > 0 {
		m.memRatio = m.totalMem / m.totalMemAllocatable
	}
	if m.totalCPUAllocatable > 0 {
		m.cpuRatio = m.totalCPU / m.totalCPUAllocatable
	}
	m.driver = "MEM"
	if m.cpuRatio > m.memRatio {
		m.driver = "CPU"
	}
	if !m.initialSet {
		m.initialCpuRatio = m.cpuRatio
		m.initialMemRatio = m.memRatio
		m.initialSet = true
	}
	return nil
}

func (m *resourceUtilizationMetric) ReportNodeLevel(t testing.TB) {
	for nodeName, val := range m.nodeValues {
		resources, ok := val.(nodeResources)
		if !ok {
			t.Errorf("unexpected type in m.nodeValues, expected NodeResources")
			continue
		}
		cpuRatio := resources.CPU.Requested / resources.CPU.Allocatable
		memRatio := resources.Mem.Requested / resources.Mem.Allocatable

		driver := "Mem"
		if cpuRatio > memRatio {
			driver = "CPU"
		}
		t.Logf("Node: %s, CPU ratio: %.2f, Mem ratio: %.2f, Efficiency: %.2f (%s driven)", nodeName, cpuRatio, memRatio, math.Max(memRatio, cpuRatio), driver)
	}
}

func (m *resourceUtilizationMetric) ReportClusterLevel(t testing.TB) {
	t.Logf("Allocated CPU: %.2f, Allocatable CPU: %.2f, CPU ratio: %.2f",
		m.totalCPU, m.totalCPUAllocatable, m.cpuRatio)
	t.Logf("Allocated Mem: %.2f, Allocatable Mem: %.2f, Mem ratio: %.2f",
		m.totalMem, m.totalMemAllocatable, m.memRatio)

	driver := "Mem"
	if m.cpuRatio > m.memRatio {
		driver = "CPU"
	}
	t.Logf("Overall efficiency: %.2f (%s driven)", math.Max(m.cpuRatio, m.memRatio), driver)
}

func (m *resourceUtilizationMetric) ReportBenchmark(b *testing.B) {
	cpuImprovement := m.cpuRatio - m.initialCpuRatio
	memImprovement := m.memRatio - m.initialMemRatio
	b.ReportMetric(m.initialCpuRatio*100, "util_cpu_%_init")
	b.ReportMetric(m.initialMemRatio*100, "util_mem_%_init")
	b.ReportMetric(m.cpuRatio*100, "util_cpu_%_final")
	b.ReportMetric(m.memRatio*100, "util_mem_%_final")
	b.ReportMetric(cpuImprovement*100, "util_cpu_%_improved")
	b.ReportMetric(memImprovement*100, "util_mem_%_improved")
}

type nodeFreeResources struct {
	FreeCPU float64
	FreeMem float64
}

// resourceFragmentationMetric computes value of resource fragmentation and Herfindahl-Hirschman Index.
// HHI originally measures market concentration and competitiveness, here repurposed to measure consolidation of free resources in the cluster.
// If HHI is lower, it indicates more competition, it higher it indicates consolidation.
// In this context, higher values mean free resources are consolidated to fewer nodes, meaning lower fragmentation, therefore computed as 1-HHI.
type resourceFragmentationMetric struct {
	metricReportingConfig
	nodeValues     map[string]any
	initialSet     bool
	initialFragCPU float64
	initialFragMem float64
	totalFreeCPU   float64
	totalFreeMem   float64
	hhiCPU         float64
	hhiMem         float64
	fragCPU        float64
	fragMem        float64
}

func NewResourceFragmentationMetric(opts ...metricOption) *resourceFragmentationMetric {
	m := &resourceFragmentationMetric{
		nodeValues: make(map[string]any),
	}
	applyMetricOptions(&m.metricReportingConfig, opts...)
	return m
}

func (m *resourceFragmentationMetric) GetNodeValues() map[string]any {
	return m.nodeValues
}

func (m *resourceFragmentationMetric) Name() string {
	return "resource_fragmentation"
}

// ComputeNodeLevel could reuse resourceUtilizationMetric which already computes needed values to avoid code duplication but that introduces unnecessary state holding.
func (m *resourceFragmentationMetric) ComputeNodeLevel(nodeInfos []*framework.NodeInfo) error {
	currentTime := time.Now()
	m.nodeValues = make(map[string]any)
	if len(nodeInfos) == 0 {
		return nil
	}

	for _, nodeInfo := range nodeInfos {
		if nodeInfo == nil || nodeInfo.Node() == nil {
			continue
		}
		nodeName := nodeInfo.Node().Name
		cAlloc, mAlloc, cUtil, mUtil, err := calculateAllocatableRequested(currentTime, nodeInfo, nodeName)
		if err != nil {
			return err
		}

		freeCPU := cAlloc - cUtil*cAlloc
		freeMem := mAlloc - mUtil*mAlloc

		// may signal overcommit, where packing may be dangerous and add to the node throttling
		if freeCPU < 0 {
			freeCPU = 0
		}
		if freeMem < 0 {
			freeMem = 0
		}
		m.nodeValues[nodeName] = nodeFreeResources{
			FreeCPU: freeCPU,
			FreeMem: freeMem,
		}
	}
	return nil
}

func (m *resourceFragmentationMetric) ComputeClusterLevel() error {
	m.totalFreeCPU, m.totalFreeMem, m.hhiCPU, m.hhiMem = 0, 0, 0, 0
	m.fragCPU, m.fragMem = -1.0, -1.0

	for _, val := range m.nodeValues {
		res, ok := val.(nodeFreeResources)
		if !ok {
			return fmt.Errorf("unexpected type in m.nodeValues, expected NodeFreeResources")
		}
		m.totalFreeCPU += res.FreeCPU
		m.totalFreeMem += res.FreeMem
	}

	for _, val := range m.nodeValues {
		res, ok := val.(nodeFreeResources)
		if !ok {
			return fmt.Errorf("unexpected type in m.nodeValues, expected NodeFreeResources")
		}
		if m.totalFreeCPU > 0 {
			shareCPU := res.FreeCPU / m.totalFreeCPU
			m.hhiCPU += shareCPU * shareCPU

			if m.totalFreeMem > 0 {
				shareMem := res.FreeMem / m.totalFreeMem
				m.hhiMem += shareMem * shareMem
			}
		}
	}
	if m.totalFreeCPU > 0 {
		m.fragCPU = 1.0 - m.hhiCPU
	}
	if m.totalFreeMem > 0 {
		m.fragMem = 1.0 - m.hhiMem
	}
	if !m.initialSet {
		m.initialFragCPU = m.fragCPU
		m.initialFragMem = m.fragMem
		m.initialSet = true
	}
	return nil
}

func (m *resourceFragmentationMetric) ReportNodeLevel(t testing.TB) {
	for nodeName, val := range m.nodeValues {
		res, ok := val.(nodeFreeResources)
		if !ok {
			t.Errorf("unexpected type in m.nodeValues, expected NodeFreeResources")
			return
		}
		percCPU, percMem := 0.0, 0.0
		if m.totalFreeCPU > 0 {
			percCPU = (res.FreeCPU / m.totalFreeCPU) * 100
		}
		if m.totalFreeMem > 0 {
			percMem = (res.FreeMem / m.totalFreeMem) * 100
		}
		t.Logf("Node: %s, Free CPU: %.2f (%.2f%%), Free Mem: %.2f (%.2f%%)",
			nodeName, res.FreeCPU, percCPU, res.FreeMem, percMem)
	}
}

func (m *resourceFragmentationMetric) ReportClusterLevel(t testing.TB) {
	if m.fragCPU == -1.0 {
		t.Logf("Resource: CPU, Total free: 0.00, HHI: N/A, Fragmentation: Undefined (cluster fully allocated)")
	} else {
		t.Logf("Resource: CPU, Total free: %.2f, HHI: %.2f, Fragmentation: %.2f", m.totalFreeCPU, m.hhiCPU, m.fragCPU)
	}

	if m.fragMem == -1.0 {
		t.Logf("Resource: Mem, Total free: 0.00, HHI: N/A, Fragmentation: Undefined (cluster fully allocated)")
	} else {
		t.Logf("Resource: Mem, Total free: %.2f, HHI: %.2f, Fragmentation: %.2f", m.totalFreeMem, m.hhiMem, m.fragMem)
	}

}

// ReportBenchmark for resourceFragmentationMetric has to handle edge cases (fully allocated cluster).
// For initial and final states, -100% denotes fully allocated cluster, for delta the number is reset to 0.
// Delta: fragCpuDelta < 0 = fragmentation decreased, fragCpuDelta > 0 = fragmentation increased.
func (m *resourceFragmentationMetric) ReportBenchmark(b *testing.B) {
	initialCpuDelta := m.initialFragCPU
	finalCpuDelta := m.fragCPU
	if initialCpuDelta == -1.0 {
		initialCpuDelta = 0.0
	}
	if finalCpuDelta == -1.0 {
		finalCpuDelta = 0.0
	}
	fragCpuDelta := finalCpuDelta - initialCpuDelta
	b.ReportMetric(m.initialFragCPU*100, "frag_cpu_%_init")
	b.ReportMetric(m.fragCPU*100, "frag_cpu_%_final")
	b.ReportMetric(fragCpuDelta*100, "frag_cpu_%_delta")

	initialMemDelta := m.initialFragMem
	finalMemDelta := m.fragMem
	if initialMemDelta == -1.0 {
		initialMemDelta = 0.0
	}
	if finalMemDelta == -1.0 {
		finalMemDelta = 0.0
	}
	fragMemDelta := finalMemDelta - initialMemDelta
	b.ReportMetric(m.initialFragMem*100, "frag_mem_%_init")
	b.ReportMetric(m.fragMem*100, "frag_mem_%_final")
	b.ReportMetric(fragMemDelta*100, "frag_mem_%_delta")
}
