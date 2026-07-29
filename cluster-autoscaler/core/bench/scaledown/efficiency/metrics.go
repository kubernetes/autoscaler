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

package efficiency

import (
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	"k8s.io/klog/v2"
)

type metricsDependencies struct {
	scaledDownNodes []*status.ScaleDownNode
	nodeInfos       []*framework.NodeInfo
	currentTime     time.Time
}

// scaleDownMetric is the unified interface all metrics implement.
type scaleDownMetric interface {
	// Name returns name of the scaleDownMetric.
	Name() string
	// Compute calculates single cluster-level metric value using metricsDependencies.
	Compute(dependencies metricsDependencies) (float64, error)
	// Summarize processes metric recorded timeline and derives final benchmarking values returned in a map.
	Summarize(metricValues []float64) map[string]float64
}

// metricsTracker orchestrates the computation, storing, and reporting of multiple scaleDownMetric.
type metricsTracker struct {
	metrics         []scaleDownMetric
	metricsTimeline map[string][]float64
}

// NewMetricsTracker creates a new instance of metricsTracker with given list of scaleDownMetrics.
func NewMetricsTracker(metrics ...scaleDownMetric) *metricsTracker {
	return &metricsTracker{
		metrics:         metrics,
		metricsTimeline: make(map[string][]float64),
	}
}

// ReportToBenchmark registers final benchmark values to benchmark runner across a set of configured scaleDownMetrics.
func (mt *metricsTracker) ReportToBenchmark(b *testing.B) {
	for _, m := range mt.metrics {
		metricValues := mt.metricsTimeline[m.Name()]
		benchVals := m.Summarize(metricValues)
		for name, val := range benchVals {
			b.ReportMetric(val, name)
		}
	}
}

// ComputeMetrics computes metric value across a set of configured scaleDownMetrics.
func (mt *metricsTracker) ComputeMetrics(dependencies metricsDependencies, t testing.TB) {
	for _, m := range mt.metrics {
		val, err := m.Compute(dependencies)
		if err != nil {
			t.Errorf("failed to compute metric %s, err %v", m.Name(), err)
			continue
		}
		mt.metricsTimeline[m.Name()] = append(mt.metricsTimeline[m.Name()], val)
	}
}

type resource struct {
	allocatable float64
	requested   float64
}

func resourceAllocatableRequested(resourceName apiv1.ResourceName, nodeInfo *framework.NodeInfo, nodeName string, currentTime time.Time) (resource, error) {
	var r resource
	util, err := utilization.CalculateUtilizationOfResource(nodeInfo, resourceName, true, true, currentTime)
	if err != nil {
		return r, err
	}
	r.requested = util
	allocatable, found := nodeInfo.Node().Status.Allocatable[resourceName]
	if !found {
		return r, fmt.Errorf("failed to get allocatable %v from %s", resourceName, nodeName)
	}
	if resourceName == apiv1.ResourceCPU {
		r.allocatable = float64(allocatable.MilliValue())
	} else {
		r.allocatable = float64(allocatable.Value())
	}
	return r, nil
}

func computeResourceUtilization(metricName string, resourceName apiv1.ResourceName, dependencies metricsDependencies) (float64, error) {
	var totalRequested, totalAllocatable, ratio float64
	for _, nodeInfo := range dependencies.nodeInfos {
		if nodeInfo == nil || nodeInfo.Node() == nil {
			continue
		}
		nodeName := nodeInfo.Node().Name
		res, err := resourceAllocatableRequested(resourceName, nodeInfo, nodeName, dependencies.currentTime)
		if err != nil {
			klog.V(5).ErrorS(err, fmt.Sprintf("failed to calculate node resource %s", resourceName.String()))
			return 0, err
		}
		nodeRequested := res.requested * res.allocatable
		totalRequested += nodeRequested
		totalAllocatable += res.allocatable
	}

	if totalAllocatable > 0 {
		ratio = totalRequested / totalAllocatable
	}

	klog.V(5).Infof("Metric: %s, Allocated: %.2f, Allocatable: %.2f, Ratio: %.2f",
		metricName, totalRequested, totalAllocatable, ratio)
	return ratio, nil
}

// resourceUtilizationMetric computes utilization ratio for given resource of generic ResourceName.
type resourceUtilizationMetric struct {
	resourceName apiv1.ResourceName
}

// NewResourceUtilizationMetric creates a new instance of resourceUtilizationMetric.
func NewResourceUtilizationMetric(rn apiv1.ResourceName) *resourceUtilizationMetric {
	return &resourceUtilizationMetric{resourceName: rn}
}

func (m *resourceUtilizationMetric) Name() string {
	return fmt.Sprintf("%s_utilization", m.resourceName.String())
}

func (m *resourceUtilizationMetric) Compute(dependencies metricsDependencies) (float64, error) {
	ratio, err := computeResourceUtilization(m.Name(), m.resourceName, dependencies)
	if err != nil {
		return 0, err
	}
	return ratio, nil
}

func (m *resourceUtilizationMetric) Summarize(metricValues []float64) map[string]float64 {
	first := metricValues[0]
	last := metricValues[len(metricValues)-1]
	diff := last - first
	r := map[string]float64{
		fmt.Sprintf("%s_%%_init", m.Name()):     first * 100,
		fmt.Sprintf("%s_%%_final", m.Name()):    last * 100,
		fmt.Sprintf("%s_%%_improved", m.Name()): diff * 100,
	}
	return r
}
