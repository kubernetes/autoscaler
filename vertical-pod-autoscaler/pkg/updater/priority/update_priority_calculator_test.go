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

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

const (
	containerName = "container1"
)

func TestSortPriority(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})

	pod1 := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "2", "")).Get()
	pod2 := test.Pod().WithName("POD2").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get()
	pod3 := test.Pod().WithName("POD3").AddContainer(test.BuildTestContainer(containerName, "1", "")).Get()
	pod4 := test.Pod().WithName("POD4").AddContainer(test.BuildTestContainer(containerName, "3", "")).Get()

	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get()

	timestampNow := pod1.Status.StartTime.Time.Add(time.Hour * 24)
	calculator.AddPod(pod1, recommendation, timestampNow)
	calculator.AddPod(pod2, recommendation, timestampNow)
	calculator.AddPod(pod3, recommendation, timestampNow)
	calculator.AddPod(pod4, recommendation, timestampNow)

	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{pod3, pod1, pod4, pod2}, result, "Wrong priority order")
}

func TestSortPriorityMultiResource(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})

	pod1 := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "60M")).Get()
	pod2 := test.Pod().WithName("POD2").AddContainer(test.BuildTestContainer(containerName, "3", "90M")).Get()

	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("6", "100M").Get()

	timestampNow := pod1.Status.StartTime.Time.Add(time.Hour * 24)
	calculator.AddPod(pod1, recommendation, timestampNow)
	calculator.AddPod(pod2, recommendation, timestampNow)

	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
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

	pod1 := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "3", "10M")).Get()

	pod2 := test.Pod().WithName("POD2").AddContainer(test.BuildTestContainer(containerName, "4", "10M")).Get()
	container2 := test.BuildTestContainer(containerName2, "2", "20M")
	pod2.Spec.Containers = append(pod2.Spec.Containers, container2)

	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("6", "20M").Get()
	cpuRec, _ := resource.ParseQuantity("4")
	memRec, _ := resource.ParseQuantity("20M")
	container2rec := vpa_types.RecommendedContainerResources{
		ContainerName: containerName2,
		Target:        map[apiv1.ResourceName]resource.Quantity{apiv1.ResourceCPU: cpuRec, apiv1.ResourceMemory: memRec}}
	recommendation.ContainerRecommendations = append(recommendation.ContainerRecommendations, container2rec)

	timestampNow := pod1.Status.StartTime.Time.Add(time.Hour * 24)
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})
	calculator.AddPod(pod1, recommendation, timestampNow)
	calculator.AddPod(pod2, recommendation, timestampNow)

	// Expect pod1 to have resourceDiff=2.0 (100% change to CPU, 100% change to memory).
	podPriority1 := calculator.getUpdatePriority(pod1, recommendation)
	assert.Equal(t, 2.0, podPriority1.resourceDiff)
	// Expect pod2 to have resourceDiff=1.0 (66% change to CPU, 33% change to memory).
	podPriority2 := calculator.getUpdatePriority(pod2, recommendation)
	assert.Equal(t, 1.0, podPriority2.resourceDiff)

	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{pod1, pod2}, result, "Wrong priority order")
}

func TestSortPriorityResourcesDecrease(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})

	pod1 := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get()
	pod2 := test.Pod().WithName("POD2").AddContainer(test.BuildTestContainer(containerName, "7", "")).Get()
	pod3 := test.Pod().WithName("POD3").AddContainer(test.BuildTestContainer(containerName, "10", "")).Get()

	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("5", "").Get()

	timestampNow := pod1.Status.StartTime.Time.Add(time.Hour * 24)
	calculator.AddPod(pod1, recommendation, timestampNow)
	calculator.AddPod(pod2, recommendation, timestampNow)
	calculator.AddPod(pod3, recommendation, timestampNow)

	// Expect the following order:
	// 1. pod1 - wants to grow by 1 unit.
	// 2. pod3 - can reclaim 5 units.
	// 3. pod2 - can reclaim 2 units.
	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{pod1, pod3, pod2}, result, "Wrong priority order")
}

func TestUpdateNotRequired(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})

	pod1 := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get()

	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("4", "").Get()

	timestampNow := pod1.Status.StartTime.Time.Add(time.Hour * 24)
	calculator.AddPod(pod1, recommendation, timestampNow)

	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod should not be updated")
}

func TestUpdateRequiredOnMilliQuantities(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})

	pod1 := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "10m", "")).Get()

	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("900m", "").Get()

	timestampNow := pod1.Status.StartTime.Time.Add(time.Hour * 24)
	calculator.AddPod(pod1, recommendation, timestampNow)

	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{pod1}, result, "Pod should be updated")
}

func TestUseProcessor(t *testing.T) {

	processedRecommendation := test.Recommendation().WithContainer(containerName).WithTarget("4", "10M").Get()
	recommendationProcessor := &test.RecommendationProcessorMock{}
	recommendationProcessor.On("Apply").Return(processedRecommendation, nil)

	calculator := NewUpdatePriorityCalculator(
		nil, nil, nil, recommendationProcessor)

	pod1 := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "10M")).Get()

	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("5", "5M").Get()
	timestampNow := pod1.Status.StartTime.Time.Add(time.Hour * 24)
	calculator.AddPod(pod1, recommendation, timestampNow)

	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod should not be updated")
}

// Verify that a pod that lives for more than podLifetimeUpdateThreshold is
// updated if it has at least one container with the request:
// 1. outside the [MinRecommended...MaxRecommended] range or
// 2. diverging from the target by more than MinChangePriority.
func TestUpdateLonglivedPods(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		nil, nil, &UpdateConfig{MinChangePriority: 0.5}, &test.FakeRecommendationProcessor{})

	pods := []*apiv1.Pod{
		test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get(),
		test.Pod().WithName("POD2").AddContainer(test.BuildTestContainer(containerName, "1", "")).Get(),
		test.Pod().WithName("POD3").AddContainer(test.BuildTestContainer(containerName, "7", "")).Get(),
	}

	// Both pods are within the recommended range.
	recommendation := test.Recommendation().WithContainer(containerName).
		WithTarget("5", "").
		WithLowerBound("1", "").
		WithUpperBound("6", "").Get()

	// Pretend that the test pods started 13 hours ago.
	timestampNow := pods[0].Status.StartTime.Time.Add(time.Hour * 13)
	for i := 0; i < 3; i++ {
		calculator.AddPod(pods[i], recommendation, timestampNow)
	}
	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{pods[1], pods[2]}, result, "Exactly POD2 and POD3 should be updated")
}

// Verify that a pod that lives for less than podLifetimeUpdateThreshold is
// updated only if the request is outside the [MinRecommended...MaxRecommended]
// range for at least one container.
func TestUpdateShortlivedPods(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		nil, nil, &UpdateConfig{MinChangePriority: 0.5}, &test.FakeRecommendationProcessor{})

	pods := []*apiv1.Pod{
		test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get(),
		test.Pod().WithName("POD2").AddContainer(test.BuildTestContainer(containerName, "1", "")).Get(),
		test.Pod().WithName("POD3").AddContainer(test.BuildTestContainer(containerName, "7", "")).Get(),
	}

	// Both pods are within the recommended range.
	recommendation := test.Recommendation().WithContainer(containerName).
		WithTarget("5", "").
		WithLowerBound("1", "").
		WithUpperBound("6", "").Get()

	// Pretend that the test pods started 11 hours ago.
	timestampNow := pods[0].Status.StartTime.Time.Add(time.Hour * 11)
	for i := 0; i < 3; i++ {
		calculator.AddPod(pods[i], recommendation, timestampNow)
	}
	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{pods[2]}, result, "Only POD3 should be updated")
}

func TestUpdatePodWithQuickOOM(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		nil, nil, &UpdateConfig{MinChangePriority: 0.5}, &test.FakeRecommendationProcessor{})

	pod := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get()

	// Pretend that the test pod started 11 hours ago.
	timestampNow := pod.Status.StartTime.Time.Add(time.Hour * 11)

	pod.Status.ContainerStatuses = []apiv1.ContainerStatus{
		{
			LastTerminationState: apiv1.ContainerState{
				Terminated: &apiv1.ContainerStateTerminated{
					Reason:     "OOMKilled",
					FinishedAt: metav1.NewTime(timestampNow.Add(-1 * 3 * time.Minute)),
					StartedAt:  metav1.NewTime(timestampNow.Add(-1 * 5 * time.Minute)),
				},
			},
		},
	}

	// Pod is within the recommended range.
	recommendation := test.Recommendation().WithContainer(containerName).
		WithTarget("5", "").
		WithLowerBound("1", "").
		WithUpperBound("6", "").Get()

	calculator.AddPod(pod, recommendation, timestampNow)
	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{pod}, result, "Pod should be updated")
}

func TestDontUpdatePodWithOOMAfterLongRun(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		nil, nil, &UpdateConfig{MinChangePriority: 0.5}, &test.FakeRecommendationProcessor{})

	pod := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get()

	// Pretend that the test pod started 11 hours ago.
	timestampNow := pod.Status.StartTime.Time.Add(time.Hour * 11)

	pod.Status.ContainerStatuses = []apiv1.ContainerStatus{
		{
			LastTerminationState: apiv1.ContainerState{
				Terminated: &apiv1.ContainerStateTerminated{
					Reason:     "OOMKilled",
					FinishedAt: metav1.NewTime(timestampNow.Add(-1 * 3 * time.Minute)),
					StartedAt:  metav1.NewTime(timestampNow.Add(-1 * 60 * time.Minute)),
				},
			},
		},
	}

	// Pod is within the recommended range.
	recommendation := test.Recommendation().WithContainer(containerName).
		WithTarget("5", "").
		WithLowerBound("1", "").
		WithUpperBound("6", "").Get()

	calculator.AddPod(pod, recommendation, timestampNow)
	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod shouldn't be updated")
}

func TestDontUpdatePodWithOOMOnlyOnOneContainer(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		nil, nil, &UpdateConfig{MinChangePriority: 0.5}, &test.FakeRecommendationProcessor{})

	pod := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get()

	// Pretend that the test pod started 11 hours ago.
	timestampNow := pod.Status.StartTime.Time.Add(time.Hour * 11)

	pod.Status.ContainerStatuses = []apiv1.ContainerStatus{
		{
			LastTerminationState: apiv1.ContainerState{
				Terminated: &apiv1.ContainerStateTerminated{
					Reason:     "OOMKilled",
					FinishedAt: metav1.NewTime(timestampNow.Add(-1 * 3 * time.Minute)),
					StartedAt:  metav1.NewTime(timestampNow.Add(-1 * 5 * time.Minute)),
				},
			},
		},
		{},
	}

	// Pod is within the recommended range.
	recommendation := test.Recommendation().WithContainer(containerName).
		WithTarget("5", "").
		WithLowerBound("1", "").
		WithUpperBound("6", "").Get()

	calculator.AddPod(pod, recommendation, timestampNow)
	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod shouldn't be updated")
}

func TestNoPods(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})
	result := calculator.GetSortedPods(NewDefaultPodEvictionAdmission())
	assert.Exactly(t, []*apiv1.Pod{}, result)
}

type pod1Admission struct{}

func (p *pod1Admission) LoopInit([]*apiv1.Pod, map[*vpa_types.VerticalPodAutoscaler][]*apiv1.Pod) {}
func (p *pod1Admission) Admit(pod *apiv1.Pod, recommendation *vpa_types.RecommendedPodResources) bool {
	return pod.Name == "POD1"
}
func (p *pod1Admission) CleanUp() {}

func TestAdmission(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})

	pod1 := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "2", "")).Get()
	pod2 := test.Pod().WithName("POD2").AddContainer(test.BuildTestContainer(containerName, "4", "")).Get()
	pod3 := test.Pod().WithName("POD3").AddContainer(test.BuildTestContainer(containerName, "1", "")).Get()
	pod4 := test.Pod().WithName("POD4").AddContainer(test.BuildTestContainer(containerName, "3", "")).Get()

	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("10", "").Get()

	timestampNow := pod1.Status.StartTime.Time.Add(time.Hour * 24)
	calculator.AddPod(pod1, recommendation, timestampNow)
	calculator.AddPod(pod2, recommendation, timestampNow)
	calculator.AddPod(pod3, recommendation, timestampNow)
	calculator.AddPod(pod4, recommendation, timestampNow)

	result := calculator.GetSortedPods(&pod1Admission{})
	assert.Exactly(t, []*apiv1.Pod{pod1}, result, "Wrong priority order")
}

// Verify getUpdatePriorty does not encounter a NPE when there is no
// recommendation for a container.
func TestNoRecommendationForContainer(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil, nil, &test.FakeRecommendationProcessor{})
	pod := test.Pod().WithName("POD1").AddContainer(test.BuildTestContainer(containerName, "5", "10")).Get()

	result := calculator.getUpdatePriority(pod, nil)
	assert.NotNil(t, result)
}
