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
	"testing"
	"time"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/assert"
)

const (
	containerName = "container1"
)

func TestSortPriority(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", containerName, "2", "", nil, nil)
	pod2 := test.BuildTestPod("POD2", containerName, "4", "", nil, nil)
	pod3 := test.BuildTestPod("POD3", containerName, "1", "", nil, nil)
	pod4 := test.BuildTestPod("POD4", containerName, "3", "", nil, nil)

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

	pod1 := test.BuildTestPod("POD1", containerName, "4", "60M", nil, nil)
	pod2 := test.BuildTestPod("POD2", containerName, "3", "90M", nil, nil)

	recommendation := test.Recommendation(containerName, "6", "100M")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod1, pod2}, result, "Wrong priority order")
}

// Creates 2 pods:
// POD1
//   container1: request={3 CPU, 10 MB}, recommended={6 CPU, 20 MB}
// POD2
//   container1: request={4 CPU, 10 MB}, recommended={6 CPU, 20 MB}
//   container2: request={2 CPU, 20 MB}, recommended={4 CPU, 20 MB}
//   total:      request={6 CPU, 30 MB}, recommneded={10 CPU, 40 MB}
//
// Verify that the total resource diff is calculated as expected and that the
// pods are ordered accordingly.
func TestSortPriorityMultiContainers(t *testing.T) {
	containerName2 := "container2"

	pod1 := test.BuildTestPod("POD1", containerName, "3", "10M", nil, nil)

	pod2 := test.BuildTestPod("POD2", containerName, "4", "10M", nil, nil)
	container2 := test.BuildTestContainer(containerName2, "2", "20M")
	pod2.Spec.Containers = append(pod2.Spec.Containers, container2)

	recommendation := test.Recommendation(containerName, "6", "20M")
	cpuRec, _ := resource.ParseQuantity("4")
	memRec, _ := resource.ParseQuantity("20M")
	container2rec := vpa_types.RecommendedContainerResources{
		Name:   containerName2,
		Target: map[apiv1.ResourceName]resource.Quantity{apiv1.ResourceCPU: cpuRec, apiv1.ResourceMemory: memRec}}
	recommendation.ContainerRecommendations = append(recommendation.ContainerRecommendations, container2rec)

	calculator := NewUpdatePriorityCalculator(nil, nil)
	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	// Expect pod1 to have resourceDiff=2.0 (100% change to CPU, 100% change to memory).
	podPriority1 := calculator.getUpdatePriority(pod1, recommendation)
	assert.Equal(t, 2.0, podPriority1.resourceDiff)
	// Expect pod2 to have resourceDiff=1.0 (66% change to CPU, 33% change to memory).
	podPriority2 := calculator.getUpdatePriority(pod2, recommendation)
	assert.Equal(t, 1.0, podPriority2.resourceDiff)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod1, pod2}, result, "Wrong priority order")
}

func TestSortPriorityResourcesDecrease(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", containerName, "4", "", nil, nil)
	pod2 := test.BuildTestPod("POD2", containerName, "7", "", nil, nil)
	pod3 := test.BuildTestPod("POD3", containerName, "10", "", nil, nil)

	recommendation := test.Recommendation(containerName, "5", "")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)
	calculator.AddPod(pod3, recommendation)

	// Expect the following order:
	// 1. pod1 - wants to grow by 1 unit.
	// 2. pod3 - can reclaim 5 units.
	// 3. pod2 - can reclaim 2 units.
	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod1, pod3, pod2}, result, "Wrong priority order")
}

func TestUpdateNotRequired(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", containerName, "4", "", nil, nil)

	recommendation := test.Recommendation(containerName, "4", "")

	calculator.AddPod(pod1, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod should not be updated")
}

func TestUpdateRequiredOnMilliQuantities(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", containerName, "10m", "", nil, nil)

	recommendation := test.Recommendation(containerName, "900m", "")

	calculator.AddPod(pod1, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod1}, result, "Pod should be updated")
}

func TestUsePolicy(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		test.BuildTestPolicy(containerName, "1", "4", "10M", "100M"), nil)

	pod1 := test.BuildTestPod("POD1", containerName, "4", "10M", nil, nil)

	recommendation := test.Recommendation(containerName, "5", "5M")

	calculator.AddPod(pod1, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod should not be updated")
}

// Verify that a pod that lives for more than podLifetimeUpdateThreshold is
// updated if it has at least one container with the request:
// 1. outside the [MinRecommended...MaxRecommended] range or
// 2. diverging from the target by more than MinChangePriority.
func TestUpdateLonglivedPods(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		nil, &UpdateConfig{MinChangePriority: 0.5})

	pods := []*apiv1.Pod{
		test.BuildTestPod("POD1", containerName, "4", "", nil, nil),
		test.BuildTestPod("POD2", containerName, "1", "", nil, nil),
		test.BuildTestPod("POD3", containerName, "7", "", nil, nil),
	}

	recommendation := test.Recommendation(containerName, "5", "")
	// Both pods are within the recommended range.
	test.AddMinRecommended(recommendation, "1", "")
	test.AddMaxRecommended(recommendation, "6", "")

	for i := 0; i < 3; i++ {
		pods[i].Status.StartTime.Time = time.Unix(0, 0) // The pods are long-lived.
		calculator.AddPod(pods[i], recommendation)
	}
	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pods[1], pods[2]}, result, "Exactly POD2 and POD3 should be updated")
}

// Verify that a pod that lives for less than podLifetimeUpdateThreshold is
// updated only if the request is outside the [MinRecommended...MaxRecommended]
// range for at least one container.
func TestUpdateShortlivedPods(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		nil, &UpdateConfig{MinChangePriority: 0.5})

	pods := []*apiv1.Pod{
		test.BuildTestPod("POD1", containerName, "4", "", nil, nil),
		test.BuildTestPod("POD2", containerName, "1", "", nil, nil),
		test.BuildTestPod("POD3", containerName, "7", "", nil, nil),
	}

	recommendation := test.Recommendation(containerName, "5", "")
	// Both pods are within the recommended range.
	test.AddMinRecommended(recommendation, "1", "")
	test.AddMaxRecommended(recommendation, "6", "")

	for i := 0; i < 3; i++ {
		calculator.AddPod(pods[i], recommendation)
	}
	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pods[2]}, result, "Only POD3 should be updated")
}

func TestNoPods(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)
	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result)
}
