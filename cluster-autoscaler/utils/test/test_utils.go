/*
Copyright 2016 The Kubernetes Authors.

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

package test

import (
	"k8s.io/kubernetes/pkg/api/resource"
	kube_api "k8s.io/kubernetes/pkg/api/v1"
)

// BuildTestPod creates a pod with specified resources.
func BuildTestPod(name string, cpu int64, mem int64) *kube_api.Pod {
	pod := &kube_api.Pod{
		ObjectMeta: kube_api.ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
		Spec: kube_api.PodSpec{
			Containers: []kube_api.Container{
				{
					Resources: kube_api.ResourceRequirements{
						Requests: kube_api.ResourceList{},
					},
				},
			},
		},
	}

	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[kube_api.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if mem >= 0 {
		pod.Spec.Containers[0].Resources.Requests[kube_api.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}

	return pod
}

// BuildTestNode creates a node with specified capacity.
func BuildTestNode(name string, cpu int64, mem int64) *kube_api.Node {
	node := &kube_api.Node{
		ObjectMeta: kube_api.ObjectMeta{
			Name: name,
		},
		Status: kube_api.NodeStatus{
			Capacity: kube_api.ResourceList{
				kube_api.ResourcePods: *resource.NewQuantity(100, resource.DecimalSI),
			},
		},
	}

	if cpu >= 0 {
		node.Status.Capacity[kube_api.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if mem >= 0 {
		node.Status.Capacity[kube_api.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}

	node.Status.Allocatable = node.Status.Capacity

	return node
}
