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

package metrics

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/utilization"
)

// resource struct holds two values:
// 1) node allocatable value of given resource (whole number)
// 2) node utilization of a given resource (float in range <0,1>)
type resource struct {
	allocatable float64
	utilization float64
}

func resourceAllocatableRequested(resourceName apiv1.ResourceName, nodeInfo *framework.NodeInfo, currentTime time.Time) (resource, error) {
	var r resource
	util, err := utilization.CalculateUtilizationOfResource(nodeInfo, resourceName, true, true, currentTime)
	if err != nil {
		return r, err
	}
	r.utilization = util
	allocatable, found := nodeInfo.Node().Status.Allocatable[resourceName]
	if !found {
		return r, fmt.Errorf("failed to get allocatable %v from %s", resourceName, nodeInfo.Node().Name)
	}
	if resourceName == apiv1.ResourceCPU {
		r.allocatable = float64(allocatable.MilliValue())
	} else {
		r.allocatable = float64(allocatable.Value())
	}
	return r, nil
}

func (m *resourceUtilizationMetric) computeResourceUtilization(clusterState ClusterState) (float64, error) {
	var totalRequested, totalAllocatable, ratio float64
	for _, nodeInfo := range clusterState.nodeInfos {
		if nodeInfo == nil || nodeInfo.Node() == nil {
			continue
		}
		res, err := resourceAllocatableRequested(m.resourceName, nodeInfo, clusterState.currentTime)
		if err != nil {
			return 0, err
		}
		nodeRequested := res.utilization * res.allocatable
		totalRequested += nodeRequested
		totalAllocatable += res.allocatable
	}

	if totalAllocatable > 0 {
		ratio = totalRequested / totalAllocatable
	}

	klog.V(5).Infof("Metric: %s, Allocated: %.2f, Allocatable: %.2f, Ratio: %.2f",
		m.Name(), totalRequested, totalAllocatable, ratio)
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

func (m *resourceUtilizationMetric) Compute(clusterState ClusterState) (float64, error) {
	ratio, err := m.computeResourceUtilization(clusterState)
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
