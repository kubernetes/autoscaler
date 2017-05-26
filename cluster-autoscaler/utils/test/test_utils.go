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
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
)

// BuildTestPod creates a pod with specified resources.
func BuildTestPod(name string, cpu int64, mem int64) *apiv1.Pod {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{},
					},
				},
			},
		},
	}

	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if mem >= 0 {
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}

	return pod
}

// BuildTestNode creates a node with specified capacity.
func BuildTestNode(name string, millicpu int64, mem int64) *apiv1.Node {
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     name,
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", name),
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourcePods: *resource.NewQuantity(100, resource.DecimalSI),
			},
		},
	}

	if millicpu >= 0 {
		node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewMilliQuantity(millicpu, resource.DecimalSI)
	}
	if mem >= 0 {
		node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(mem, resource.DecimalSI)
	}

	node.Status.Allocatable = node.Status.Capacity

	return node
}

// SetNodeReadyState sets node ready state.
func SetNodeReadyState(node *apiv1.Node, ready bool, lastTransition time.Time) {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == apiv1.NodeReady {
			node.Status.Conditions[i].LastTransitionTime = metav1.Time{Time: lastTransition}
			if ready {
				node.Status.Conditions[i].Status = apiv1.ConditionTrue
			} else {
				node.Status.Conditions[i].Status = apiv1.ConditionFalse
			}
			return
		}
	}
	condition := apiv1.NodeCondition{
		Type:               apiv1.NodeReady,
		Status:             apiv1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: lastTransition},
	}
	if ready {
		condition.Status = apiv1.ConditionTrue
	} else {
		condition.Status = apiv1.ConditionFalse
	}
	node.Status.Conditions = append(node.Status.Conditions, condition)
}

// RefJSON builds string reference to
func RefJSON(o runtime.Object) string {
	ref, err := apiv1.GetReference(api.Scheme, o)
	if err != nil {
		panic(err)
	}

	codec := testapi.Default.Codec()
	json := runtime.EncodeOrDie(codec, &apiv1.SerializedReference{Reference: *ref})
	return string(json)
}

// GetReplicaSetAnnotation returns a map containing annotation simulating pod being created by ReplicaSet
func GetReplicaSetAnnotation() map[string]string {
	return map[string]string{
		"kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\"}}",
	}
}
