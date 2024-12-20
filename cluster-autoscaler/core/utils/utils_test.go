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

package utils

import (
	"testing"
	"time"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetNodeResource(t *testing.T) {
	node := BuildTestNode("n1", 1000, 2*MiB)

	cores := getNodeResource(node, apiv1.ResourceCPU)
	assert.Equal(t, int64(1), cores)

	memory := getNodeResource(node, apiv1.ResourceMemory)
	assert.Equal(t, int64(2*MiB), memory)

	unknownResourceValue := getNodeResource(node, "unknown resource")
	assert.Equal(t, int64(0), unknownResourceValue)

	// if we have no resources in capacity we expect getNodeResource to return 0
	nodeWithMissingCapacity := BuildTestNode("n1", 1000, 2*MiB)
	nodeWithMissingCapacity.Status.Capacity = apiv1.ResourceList{}

	cores = getNodeResource(nodeWithMissingCapacity, apiv1.ResourceCPU)
	assert.Equal(t, int64(0), cores)

	memory = getNodeResource(nodeWithMissingCapacity, apiv1.ResourceMemory)
	assert.Equal(t, int64(0), memory)

	// if we have negative values in resources we expect getNodeResource to return 0
	nodeWithNegativeCapacity := BuildTestNode("n1", -1000, -2*MiB)
	nodeWithNegativeCapacity.Status.Capacity = apiv1.ResourceList{}

	cores = getNodeResource(nodeWithNegativeCapacity, apiv1.ResourceCPU)
	assert.Equal(t, int64(0), cores)

	memory = getNodeResource(nodeWithNegativeCapacity, apiv1.ResourceMemory)
	assert.Equal(t, int64(0), memory)

}

func TestGetNodeCoresAndMemory(t *testing.T) {
	node := BuildTestNode("n1", 2000, 2048*MiB)

	cores, memory := GetNodeCoresAndMemory(node)
	assert.Equal(t, int64(2), cores)
	assert.Equal(t, int64(2048*MiB), memory)

	// if we have no cpu/memory defined in capacity we expect getNodeCoresAndMemory to return 0s
	nodeWithMissingCapacity := BuildTestNode("n1", 1000, 2*MiB)
	nodeWithMissingCapacity.Status.Capacity = apiv1.ResourceList{}

	cores, memory = GetNodeCoresAndMemory(nodeWithMissingCapacity)
	assert.Equal(t, int64(0), cores)
	assert.Equal(t, int64(0), memory)
}

func TestGetOldestPod(t *testing.T) {
	p1 := BuildTestPod("p1", 500, 1000)
	p1.CreationTimestamp = metav1.NewTime(time.Now().Add(-1 * time.Minute))
	p2 := BuildTestPod("p2", 500, 1000)
	p2.CreationTimestamp = metav1.NewTime(time.Now().Add(+1 * time.Minute))
	p3 := BuildTestPod("p3", 500, 1000)
	p3.CreationTimestamp = metav1.NewTime(time.Now())

	assert.Equal(t, p1.CreationTimestamp.Time, GetOldestCreateTime([]*apiv1.Pod{p1, p2, p3}))
	assert.Equal(t, p1.CreationTimestamp.Time, GetOldestCreateTime([]*apiv1.Pod{p3, p2, p1}))
}
