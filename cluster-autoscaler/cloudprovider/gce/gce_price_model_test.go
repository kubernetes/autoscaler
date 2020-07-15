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

package gce

import (
	"math"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	"github.com/stretchr/testify/assert"
)

func testNode(t *testing.T, nodeName string, instanceType string, millicpu int64, mem int64, gpuType string, gpuCount int64, isPreemptible bool) *apiv1.Node {
	node := BuildTestNode(nodeName, millicpu, mem)
	labels, err := BuildGenericLabels(GceRef{
		Name:    "kubernetes-minion-group",
		Project: "mwielgus-proj",
		Zone:    "us-central1-b"},
		instanceType,
		nodeName,
		OperatingSystemLinux)
	assert.NoError(t, err)
	if isPreemptible {
		labels[preemptibleLabel] = "true"
	}
	if gpuCount > 0 {
		node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(gpuCount, resource.DecimalSI)
		node.Status.Allocatable[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(gpuCount, resource.DecimalSI)
		if gpuType != "" {
			labels[GPULabel] = gpuType
		}
	}
	node.Labels = labels
	return node
}

// this test is meant to cover all the branches in pricing logic, not all possible types of instances
func TestGetNodePrice(t *testing.T) {
	// tests assert that price(cheaperNode) < priceComparisonCoefficient * price(expensiveNode)
	cases := map[string]struct {
		cheaperNode                *apiv1.Node
		expensiveNode              *apiv1.Node
		priceComparisonCoefficient float64
	}{
		// instance types
		"e2 is cheaper than n1": {
			cheaperNode:                testNode(t, "e2", "e2-standard-8", 8000, 32*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "n1", "n1-standard-8", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 1,
		},
		"custom nodes are more expensive than n1": {
			cheaperNode:                testNode(t, "n1", "n1-standard-8", 8000, 30*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "custom", "custom-8", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 1,
		},
		"custom nodes are not extremely expensive": {
			cheaperNode:                testNode(t, "custom", "custom-8", 8000, 30*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "n1", "n1-standard-8", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 1.2,
		},
		"custom node price scales linearly": {
			cheaperNode:                testNode(t, "small_custom", "custom-1", 1000, 3.75*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "large_custom", "custom-8", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 1.0 / 7.9,
		},
		"custom node price scales linearly 2": {
			cheaperNode:                testNode(t, "large_custom", "custom-8", 8000, 30*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "small_custom", "custom-1", 1000, 3.75*units.GiB, "", 0, false),
			priceComparisonCoefficient: 8.1,
		},
		// GPUs
		"accelerators are expensive": {
			cheaperNode: testNode(t, "no_accelerators", "n1-standard-8", 8000, 30*units.GiB, "", 0, false),
			// #NotFunny
			expensiveNode:              testNode(t, "large hadron collider", "n1-standard-8", 8000, 30*units.GiB, "nvidia-tesla-v100", 1, false),
			priceComparisonCoefficient: 0.5,
		},
		"GPUs of unknown type are still expensive": {
			cheaperNode:                testNode(t, "no_accelerators", "n1-standard-8", 8000, 30*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "cyclotron", "n1-standard-8", 8000, 30*units.GiB, "", 1, false),
			priceComparisonCoefficient: 0.5,
		},
		"different GPUs have different prices": {
			cheaperNode:                testNode(t, "cheap gpu", "n1-standard-8", 8000, 30*units.GiB, "nvidia-tesla-t4", 1, false),
			expensiveNode:              testNode(t, "large hadron collider", "n1-standard-8", 8000, 30*units.GiB, "nvidia-tesla-v100", 1, false),
			priceComparisonCoefficient: 0.5,
		},
		"more GPUs is more expensive": {
			cheaperNode:                testNode(t, "1 gpu", "n1-standard-8", 8000, 30*units.GiB, "nvidia-tesla-v100", 1, false),
			expensiveNode:              testNode(t, "2 gpus", "n1-standard-8", 8000, 30*units.GiB, "nvidia-tesla-v100", 2, false),
			priceComparisonCoefficient: 0.7,
		},
		// Preemptibles
		"preemtpibles are cheap": {
			cheaperNode:                testNode(t, "preempted_i_can_be", "n1-standard-8", 8000, 30*units.GiB, "", 0, true),
			expensiveNode:              testNode(t, "ondemand", "n1-standard-8", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 0.25,
		},
		"custom preemptibles are also cheap": {
			cheaperNode:                testNode(t, "preempted_i_can_be", "custom-8", 8000, 30*units.GiB, "", 0, true),
			expensiveNode:              testNode(t, "ondemand", "custom-8", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 0.25,
		},
		"preemtpibles GPUs are (relatively) cheap": {
			cheaperNode:                testNode(t, "preempted_i_can_be", "n1-standard-8", 8000, 30*units.GiB, "nvidia-tesla-v100", 2, true),
			expensiveNode:              testNode(t, "ondemand", "n1-standard-8", 8000, 30*units.GiB, "nvidia-tesla-v100", 2, false),
			priceComparisonCoefficient: 0.5,
		},
		// Unknown instances
		"unknown cost is similar to its node family": {
			cheaperNode:                testNode(t, "unknown", "n1-unknown", 8000, 30*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "known", "n1-standard-8", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 1.001,
		},
		"unknown cost is similar to its node family 2": {
			cheaperNode:                testNode(t, "unknown", "n1-standard-8", 8000, 30*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "known", "n1-unknown", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 1.001,
		},
		// Custom instances
		"big custom from cheap family is cheaper than small custom from expensive family": {
			cheaperNode:                testNode(t, "unknown", "e2-custom", 9000, 32*units.GiB, "", 0, false),
			expensiveNode:              testNode(t, "known", "n1-custom", 8000, 30*units.GiB, "", 0, false),
			priceComparisonCoefficient: 1.001,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			model := &GcePriceModel{}
			now := time.Now()

			price1, err := model.NodePrice(tc.cheaperNode, now, now.Add(time.Hour))
			assert.NoError(t, err)
			price2, err := model.NodePrice(tc.expensiveNode, now, now.Add(time.Hour))
			assert.NoError(t, err)
			if price1 >= tc.priceComparisonCoefficient*price2 {
				t.Errorf("Failed price comparison, price1=%v price2=%v price2*coefficient=%v", price1, price2, price2*tc.priceComparisonCoefficient)
			}
		})
	}
}

func TestGetPodPrice(t *testing.T) {
	pod1 := BuildTestPod("a1", 100, 500*units.MiB)
	pod2 := BuildTestPod("a2", 2*100, 2*500*units.MiB)

	model := &GcePriceModel{}
	now := time.Now()

	price1, err := model.PodPrice(pod1, now, now.Add(time.Hour))
	assert.NoError(t, err)
	price2, err := model.PodPrice(pod2, now, now.Add(time.Hour))
	assert.NoError(t, err)
	// 2 times bigger pod should cost twice as much.
	assert.True(t, math.Abs(price1*2-price2) < 0.001)
}
