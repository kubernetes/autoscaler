/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package estimator

import (
	"testing"

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"

	"github.com/stretchr/testify/assert"
)

func TestEstimate(t *testing.T) {
	cpuPerPod := int64(500)
	memoryPerPod := int64(1000 * 1024 * 1024)

	pod := &kube_api.Pod{
		Spec: kube_api.PodSpec{
			Containers: []kube_api.Container{
				{
					Resources: kube_api.ResourceRequirements{
						Requests: kube_api.ResourceList{
							kube_api.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod, resource.DecimalSI),
							kube_api.ResourceMemory: *resource.NewQuantity(memoryPerPod, resource.DecimalSI),
						},
					},
				},
			},
		},
	}

	estimator := NewBasicNodeEstimator()

	for i := 0; i < 5; i++ {
		estimator.Add(pod)
	}

	assert.Equal(t, int64(500*5), estimator.cpuSum.MilliValue())
	assert.Equal(t, int64(5*memoryPerPod), estimator.memorySum.Value())
	assert.Equal(t, 5, estimator.count)

	node := &kube_api.Node{
		Status: kube_api.NodeStatus{
			Capacity: kube_api.ResourceList{
				kube_api.ResourceCPU:    *resource.NewMilliQuantity(3*cpuPerPod, resource.DecimalSI),
				kube_api.ResourceMemory: *resource.NewQuantity(2*memoryPerPod, resource.DecimalSI),
				kube_api.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}
	estimate, report := estimator.Estimate(node)
	assert.Contains(t, estimator.GetDebug(), "CPU")
	assert.Contains(t, report, "CPU")
	assert.Equal(t, 3, estimate)
}

func TestEstimateWithPorts(t *testing.T) {
	cpuPerPod := int64(500)
	memoryPerPod := int64(1000 * 1024 * 1024)

	pod := &kube_api.Pod{
		Spec: kube_api.PodSpec{
			Containers: []kube_api.Container{
				{
					Resources: kube_api.ResourceRequirements{
						Requests: kube_api.ResourceList{
							kube_api.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod, resource.DecimalSI),
							kube_api.ResourceMemory: *resource.NewQuantity(memoryPerPod, resource.DecimalSI),
						},
					},
					Ports: []kube_api.ContainerPort{
						{
							HostPort: 5555,
						},
					},
				},
			},
		},
	}

	estimator := NewBasicNodeEstimator()
	for i := 0; i < 5; i++ {
		estimator.Add(pod)
	}
	node := &kube_api.Node{
		Status: kube_api.NodeStatus{
			Capacity: kube_api.ResourceList{
				kube_api.ResourceCPU:    *resource.NewMilliQuantity(3*cpuPerPod, resource.DecimalSI),
				kube_api.ResourceMemory: *resource.NewQuantity(2*memoryPerPod, resource.DecimalSI),
				kube_api.ResourcePods:   *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}

	estimate, report := estimator.Estimate(node)
	assert.Contains(t, estimator.GetDebug(), "CPU")
	assert.Contains(t, report, "CPU")
	assert.Equal(t, 5, estimate)
}
