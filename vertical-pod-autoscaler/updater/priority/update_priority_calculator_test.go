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

package priority

import (
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/apimock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/test"
	"testing"
)

const (
	containerName = "container1"
)

func TestSortPriority(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", containerName, "2", "", nil)
	pod2 := test.BuildTestPod("POD2", containerName, "4", "", nil)
	pod3 := test.BuildTestPod("POD3", containerName, "1", "", nil)
	pod4 := test.BuildTestPod("POD4", containerName, "3", "", nil)

	recommendation := test.Recommendation(containerName, "10", "")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)
	calculator.AddPod(pod3, recommendation)
	calculator.AddPod(pod4, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod3, pod1, pod4, pod2}, result, "Wrong priority order")
}

func TestSortPriorityMultiResource(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", containerName, "4", "60M", nil)
	pod2 := test.BuildTestPod("POD2", containerName, "3", "90M", nil)

	recommendation := test.Recommendation(containerName, "6", "100M")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod1, pod2}, result, "Wrong priority order")
}

func TestSortPriorityMultiContainers(t *testing.T) {

	containerName2 := "container2"

	pod1 := test.BuildTestPod("POD1", containerName, "3", "10M", nil)

	pod2 := test.BuildTestPod("POD2", containerName, "4", "10M", nil)
	container2 := test.BuildTestContainer(containerName2, "3", "20M")
	pod2.Spec.Containers = append(pod1.Spec.Containers, container2)

	recommendation := test.Recommendation(containerName, "6", "20M")
	cpuRec, _ := resource.ParseQuantity("4")
	memRec, _ := resource.ParseQuantity("20M")
	container2rec := apimock.ContainerRecommendation{
		Name:      containerName2,
		Resources: map[apiv1.ResourceName]resource.Quantity{apiv1.ResourceCPU: cpuRec, apiv1.ResourceMemory: memRec}}
	recommendation.Containers = append(recommendation.Containers, container2rec)

	calculator := NewUpdatePriorityCalculator(nil, nil)
	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod2, pod1}, result, "Wrong priority order")
}

func TestSortPriorityResorucesDecrease(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", containerName, "4", "", nil)
	pod2 := test.BuildTestPod("POD2", containerName, "10", "", nil)

	recommendation := test.Recommendation(containerName, "5", "")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod2, pod1}, result, "Wrong priority order")
}

func TestUpdateNotRequired(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", containerName, "4", "", nil)

	recommendation := test.Recommendation(containerName, "4", "")

	calculator.AddPod(pod1, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod should not be updated")
}

func TestUsePolicy(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		test.BuildTestPolicy(containerName, "1", "4", "10M", "100M"), nil)

	pod1 := test.BuildTestPod("POD1", containerName, "4", "10M", nil)

	recommendation := test.Recommendation(containerName, "5", "5M")

	calculator.AddPod(pod1, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod should not be updated")
}

func TestChangeTooSmall(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, &UpdateConfig{0.5})

	pod1 := test.BuildTestPod("POD1", containerName, "4", "", nil)
	pod2 := test.BuildTestPod("POD2", containerName, "1", "", nil)

	recommendation := test.Recommendation(containerName, "5", "")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod2}, result, "Only POD2 should be updated")
}

func TestNoPods(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)
	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result)
}
