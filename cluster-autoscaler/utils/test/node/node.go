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

package node

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

// BuildTestNode creates a node with specified capacity.
func BuildTestNode(name string, millicpu int64, mem int64) *apiv1.Node {
	node := baseNode(name)

	if millicpu >= 0 {
		node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewMilliQuantity(millicpu, resource.DecimalSI)
	}
	if mem >= 0 {
		node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}

	node.Status.Allocatable = apiv1.ResourceList{}
	for k, v := range node.Status.Capacity {
		node.Status.Allocatable[k] = v
	}

	return node
}

func baseNode(name string) *apiv1.Node {
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     name,
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", name),
			Labels:   map[string]string{},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: name,
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourcePods: *resource.NewQuantity(100, resource.DecimalSI),
			},
		},
	}
}

// AddEphemeralStorageToNode adds ephemeral storage capacity to a given node.
func AddEphemeralStorageToNode(node *apiv1.Node, eph int64) *apiv1.Node {
	if eph >= 0 {
		node.Status.Capacity[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(eph, resource.DecimalSI)
		node.Status.Allocatable[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(eph, resource.DecimalSI)
	}
	return node
}

// AddGpusToNode adds GPU capacity to given node. Default accelerator type is used.
func AddGpusToNode(node *apiv1.Node, gpusCount int64) {
	node.Spec.Taints = append(
		node.Spec.Taints,
		apiv1.Taint{
			Key:    test.ResourceNvidiaGPU,
			Value:  "present",
			Effect: "NoSchedule",
		})
	node.Status.Capacity[test.ResourceNvidiaGPU] = *resource.NewQuantity(gpusCount, resource.DecimalSI)
	node.Status.Allocatable[test.ResourceNvidiaGPU] = *resource.NewQuantity(gpusCount, resource.DecimalSI)
	AddGpuLabelToNode(node)
}

// AddGpuLabelToNode adds GPULabel to give node. This is used to mock intermediate result that GPU on node is not ready
func AddGpuLabelToNode(node *apiv1.Node) {
	node.Labels[test.GPULabel] = test.DefaultGPUType
}

// GetGPULabel return GPULabel on the node. This is only used in unit tests.
func GetGPULabel() string {
	return test.GPULabel
}

// GetGpuConfigFromNode returns the GPU of the node if it has one. This is only used in unit tests.
func GetGpuConfigFromNode(node *apiv1.Node) *cloudprovider.GpuConfig {
	gpuType, hasGpuLabel := node.Labels[test.GPULabel]
	gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[test.ResourceNvidiaGPU]
	if hasGpuLabel || (hasGpuAllocatable && !gpuAllocatable.IsZero()) {
		return &cloudprovider.GpuConfig{
			Label:        test.GPULabel,
			Type:         gpuType,
			ResourceName: test.ResourceNvidiaGPU,
		}
	}
	return nil
}

// SetNodeReadyState sets node ready state to either ConditionTrue or ConditionFalse.
func SetNodeReadyState(node *apiv1.Node, ready bool, lastTransition time.Time) {
	if ready {
		SetNodeCondition(node, apiv1.NodeReady, apiv1.ConditionTrue, lastTransition)
	} else {
		SetNodeCondition(node, apiv1.NodeReady, apiv1.ConditionFalse, lastTransition)
		node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
			Key:    "node.kubernetes.io/not-ready",
			Value:  "true",
			Effect: apiv1.TaintEffectNoSchedule,
		})
	}
}

// SetNodeNotReadyTaint sets the not ready taint on node.
func SetNodeNotReadyTaint(node *apiv1.Node) {
	node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{Key: apiv1.TaintNodeNotReady, Effect: apiv1.TaintEffectNoSchedule})
}

// RemoveNodeNotReadyTaint removes the not ready taint.
func RemoveNodeNotReadyTaint(node *apiv1.Node) {
	var final []apiv1.Taint
	for i := range node.Spec.Taints {
		if node.Spec.Taints[i].Key == apiv1.TaintNodeNotReady {
			continue
		}
		final = append(final, node.Spec.Taints[i])
	}
	node.Spec.Taints = final
}

// SetNodeCondition sets node condition.
func SetNodeCondition(node *apiv1.Node, conditionType apiv1.NodeConditionType, status apiv1.ConditionStatus, lastTransition time.Time) {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == conditionType {
			node.Status.Conditions[i].LastTransitionTime = metav1.Time{Time: lastTransition}
			node.Status.Conditions[i].Status = status
			return
		}
	}
	// Condition doesn't exist yet.
	condition := apiv1.NodeCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Time{Time: lastTransition},
	}
	node.Status.Conditions = append(node.Status.Conditions, condition)
}
