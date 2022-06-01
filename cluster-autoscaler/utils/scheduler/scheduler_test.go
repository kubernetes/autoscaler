/*
Copyright 2017 The Kubernetes Authors.

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

package scheduler

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestCreateNodeNameToInfoMap(t *testing.T) {
	p1 := BuildTestPod("p1", 1500, 200000)
	p1.Spec.NodeName = "node1"
	p2 := BuildTestPod("p2", 3000, 200000)
	p2.Spec.NodeName = "node2"
	p3 := BuildTestPod("p3", 3000, 200000)
	p3.Spec.NodeName = "node3"

	var priority int32 = 100
	podWaitingForPreemption := BuildTestPod("w1", 1500, 200000)
	podWaitingForPreemption.Spec.Priority = &priority
	podWaitingForPreemption.Status.NominatedNodeName = "node1"

	n1 := BuildTestNode("node1", 2000, 2000000)
	n2 := BuildTestNode("node2", 2000, 2000000)

	res := CreateNodeNameToInfoMap([]*apiv1.Pod{p1, p2, p3, podWaitingForPreemption}, []*apiv1.Node{n1, n2})
	assert.Equal(t, 2, len(res))
	assert.Equal(t, p1, res["node1"].Pods[0].Pod)
	assert.Equal(t, podWaitingForPreemption, res["node1"].Pods[1].Pod)
	assert.Equal(t, n1, res["node1"].Node())
	assert.Equal(t, p2, res["node2"].Pods[0].Pod)
	assert.Equal(t, n2, res["node2"].Node())
}

func TestResourceList(t *testing.T) {
	tests := []struct {
		resource *schedulerframework.Resource
		expected apiv1.ResourceList
	}{
		{
			resource: &schedulerframework.Resource{},
			expected: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceCPU:              *resource.NewScaledQuantity(0, -3),
				apiv1.ResourceMemory:           *resource.NewQuantity(0, resource.BinarySI),
				apiv1.ResourcePods:             *resource.NewQuantity(0, resource.BinarySI),
				apiv1.ResourceEphemeralStorage: *resource.NewQuantity(0, resource.BinarySI),
			},
		},
		{
			resource: &schedulerframework.Resource{
				MilliCPU:         4,
				Memory:           2000,
				EphemeralStorage: 5000,
				AllowedPodNumber: 80,
				ScalarResources: map[apiv1.ResourceName]int64{
					"scalar.test/scalar1":        1,
					"hugepages-test":             2,
					"attachable-volumes-aws-ebs": 39,
				},
			},
			expected: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceCPU:                      *resource.NewScaledQuantity(4, -3),
				apiv1.ResourceMemory:                   *resource.NewQuantity(2000, resource.BinarySI),
				apiv1.ResourcePods:                     *resource.NewQuantity(80, resource.BinarySI),
				apiv1.ResourceEphemeralStorage:         *resource.NewQuantity(5000, resource.BinarySI),
				"scalar.test/" + "scalar1":             *resource.NewQuantity(1, resource.DecimalSI),
				"attachable-volumes-aws-ebs":           *resource.NewQuantity(39, resource.DecimalSI),
				apiv1.ResourceHugePagesPrefix + "test": *resource.NewQuantity(2, resource.BinarySI),
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			rl := ResourceToResourceList(test.resource)
			if !reflect.DeepEqual(test.expected, rl) {
				t.Errorf("expected: %#v, got: %#v", test.expected, rl)
			}
		})
	}
}
