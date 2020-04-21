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

package packet

import (
	"math"
	"testing"
	"time"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	"github.com/stretchr/testify/assert"
)

func TestGetNodePrice(t *testing.T) {
	labelsPool1 := BuildGenericLabels("pool1", "t1.small.x86")
	plan1 := InstanceTypes["t1.small.x86"]

	labelsPool2 := BuildGenericLabels("pool2", "c1.xlarge.x86")
	plan2 := InstanceTypes["c1.xlarge.x86"]

	model := &PacketPriceModel{}
	now := time.Now()

	node1 := BuildTestNode("node1", plan1.CPU*1000, plan1.MemoryMb*1024*1024)
	node1.Labels = labelsPool1
	price1, err := model.NodePrice(node1, now, now.Add(time.Hour))
	assert.NoError(t, err)

	node2 := BuildTestNode("node2", plan2.CPU*1000, plan2.MemoryMb*1024*1024)
	node2.Labels = labelsPool2
	price2, err := model.NodePrice(node2, now, now.Add(time.Hour))
	assert.NoError(t, err)

	assert.True(t, price1 == 0.07)
	assert.True(t, price2 == 1.75)
}

func TestGetPodPrice(t *testing.T) {
	pod1 := BuildTestPod("pod1", 100, 500*units.MiB)
	pod2 := BuildTestPod("pod2", 2*100, 2*500*units.MiB)

	model := &PacketPriceModel{}
	now := time.Now()

	price1, err := model.PodPrice(pod1, now, now.Add(time.Hour))
	assert.NoError(t, err)
	price2, err := model.PodPrice(pod2, now, now.Add(time.Hour))
	assert.NoError(t, err)
	// 2 times bigger pod should cost twice as much.
	assert.True(t, math.Abs(price1*2-price2) < 0.001)
}
