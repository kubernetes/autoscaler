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

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	"github.com/stretchr/testify/assert"
)

func TestGetNodePrice(t *testing.T) {
	labels1, _ := BuildGenericLabels(GceRef{
		Name:    "kubernetes-minion-group",
		Project: "mwielgus-proj",
		Zone:    "us-central1-b"},
		"n1-standard-8",
		"sillyname",
		OperatingSystemLinux)

	labels2, _ := BuildGenericLabels(GceRef{
		Name:    "kubernetes-minion-group",
		Project: "mwielgus-proj",
		Zone:    "us-central1-b"},
		"n1-standard-8",
		"sillyname",
		OperatingSystemLinux)
	labels2[preemptibleLabel] = "true"

	labels3, _ := BuildGenericLabels(GceRef{
		Name:    "kubernetes-minion-group",
		Project: "mwielgus-proj",
		Zone:    "us-central1-b"},
		"n1-custom",
		"sillyname",
		OperatingSystemLinux)

	labels4, _ := BuildGenericLabels(GceRef{
		Name:    "kubernetes-minion-group",
		Project: "mwielgus-proj",
		Zone:    "us-central1-b"},
		"n1-unknown",
		"sillyname",
		OperatingSystemLinux)

	labels5, _ := BuildGenericLabels(GceRef{
		Name:    "kubernetes-minion-group",
		Project: "mwielgus-proj",
		Zone:    "us-central1-b"},
		"e2-custom",
		"sillyname",
		OperatingSystemLinux)

	model := &GcePriceModel{}
	now := time.Now()

	// regular
	node1 := BuildTestNode("sillyname1", 8000, 30*units.GiB)
	node1.Labels = labels1
	price1, err := model.NodePrice(node1, now, now.Add(time.Hour))
	assert.NoError(t, err)

	// preemptible
	node2 := BuildTestNode("sillyname2", 8000, 30*units.GiB)
	node2.Labels = labels2
	price2, err := model.NodePrice(node2, now, now.Add(time.Hour))
	assert.NoError(t, err)
	// preemptible nodes should be way cheaper than regular.
	assert.True(t, price1 > 3*price2)

	// custom node
	node3 := BuildTestNode("sillyname3", 8000, 30*units.GiB)
	node3.Labels = labels3
	price3, err := model.NodePrice(node3, now, now.Add(time.Hour))
	assert.NoError(t, err)
	// custom nodes should be slightly more expensive than regular.
	assert.True(t, price1 < price3)
	assert.True(t, price1*1.2 > price3)

	// regular with gpu
	node4 := BuildTestNode("sillyname4", 8000, 30*units.GiB)
	node4.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	node4.Labels = labels1
	price4, _ := model.NodePrice(node4, now, now.Add(time.Hour))

	// preemptible with gpu
	node5 := BuildTestNode("sillyname5", 8000, 30*units.GiB)
	node5.Labels = labels2
	node5.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	price5, _ := model.NodePrice(node5, now, now.Add(time.Hour))

	// Nodes with GPU are way more expensive than regular.
	// Being preemptible doesn't bring much of a discount (less than 50%).
	assert.True(t, price4 > price5)
	assert.True(t, price4 < 1.5*price5)
	assert.True(t, price4 > 2*price1)

	// small custom node
	node6 := BuildTestNode("sillyname6", 1000, 3750*units.MiB)
	node6.Labels = labels3
	price6, err := model.NodePrice(node6, now, now.Add(time.Hour))
	assert.NoError(t, err)
	// 8 times smaller node should be 8 times less expensive.
	assert.True(t, math.Abs(price3-8*price6) < 0.1)

	// unknown instance type
	node7 := BuildTestNode("sillyname7", 8000, 30*units.GiB)
	node7.Labels = labels4
	price7, err := model.NodePrice(node7, now, now.Add(time.Hour))
	assert.NoError(t, err)
	// Unknown instance type should have similar pricing to its node family
	assert.True(t, math.Abs(price1-price7) < 0.1)

	// custom node from cheaper family
	node8 := BuildTestNode("sillyname8", 9000, 32*units.GiB)
	node8.Labels = labels5
	price8, err := model.NodePrice(node8, now, now.Add(time.Hour))
	assert.NoError(t, err)
	// Bigger custom e2 node should be cheaper than smaller custom n1 node
	assert.True(t, price8 < price3)
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
