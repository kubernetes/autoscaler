/*
Copyright 2023 The Kubernetes Authors.

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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/utils/ptr"

	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestLoopInit(t *testing.T) {
	podEvictionRequirements := []*vpaautoscalingv1.EvictionRequirement{
		{
			Resources:         []corev1.ResourceName{corev1.ResourceCPU},
			ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
		},
	}
	container1Name := "test-container-1"
	container2Name := "test-container-2"
	pod := test.Pod().WithName("test-pod").
		AddContainer(test.Container().WithName(container1Name).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("10Gi")).Get()).
		AddContainer(test.Container().WithName(container2Name).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("10Gi")).Get()).
		Get()
	pod2 := test.Pod().WithName("test-pod-2").
		AddContainer(test.Container().WithName(container1Name).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("10Gi")).Get()).
		AddContainer(test.Container().WithName(container2Name).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("10Gi")).Get()).
		Get()
	expectedEvictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
		pod:  podEvictionRequirements,
		pod2: podEvictionRequirements,
	}
	testVPA := test.VerticalPodAutoscaler().
		WithName("test-vpa").
		WithContainer(container1Name).
		WithEvictionRequirements(podEvictionRequirements).
		Get()
	vpaToPodMap := map[*vpaautoscalingv1.VerticalPodAutoscaler][]*corev1.Pod{testVPA: {pod, pod2}}

	t.Run("it should not require UpdateMode and EvictionRequirements.", func(t *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.LoopInit(nil, vpaToPodMap)

		newTestVPA := test.VerticalPodAutoscaler().
			WithName("test-vpa").
			WithContainer(container1Name).
			Get()

		newVpaToPodMap := map[*vpaautoscalingv1.VerticalPodAutoscaler][]*corev1.Pod{newTestVPA: {pod, pod2}}

		sdpea.LoopInit(nil, newVpaToPodMap)
		assert.Len(t, storedEvictionRequirements(sdpea), 0)
	})

	t.Run("it should store EvictionRequirements from VPA in a map per Pod", func(t *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.LoopInit(nil, vpaToPodMap)
		assert.Len(t, storedEvictionRequirements(sdpea), 2)
		assert.Equal(t, storedEvictionRequirements(sdpea), expectedEvictionRequirements)
	})

	t.Run("it should change the stored EvictionRequirement if EvictionRequirement on the VPA changed", func(t *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.LoopInit(nil, vpaToPodMap)

		newPodEvictionRequirements := []*vpaautoscalingv1.EvictionRequirement{
			{
				Resources:         []corev1.ResourceName{corev1.ResourceMemory},
				ChangeRequirement: vpaautoscalingv1.TargetLowerThanRequests,
			},
		}
		newTestVPA := test.VerticalPodAutoscaler().
			WithName("test-vpa").
			WithContainer(container1Name).
			WithEvictionRequirements(newPodEvictionRequirements).
			Get()
		newVpaToPodMap := map[*vpaautoscalingv1.VerticalPodAutoscaler][]*corev1.Pod{newTestVPA: {pod, pod2}}
		newExpectedEvictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod:  newPodEvictionRequirements,
			pod2: newPodEvictionRequirements,
		}
		sdpea.LoopInit(nil, newVpaToPodMap)
		assert.Len(t, storedEvictionRequirements(sdpea), 2)
		assert.Equal(t, storedEvictionRequirements(sdpea), newExpectedEvictionRequirements)
	})

	t.Run("it should remove a Pod from the stored EvictionRequirements when it no longer exists", func(t *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.LoopInit(nil, vpaToPodMap)

		newPodEvictionRequirements := []*vpaautoscalingv1.EvictionRequirement{
			{
				Resources:         []corev1.ResourceName{corev1.ResourceMemory},
				ChangeRequirement: vpaautoscalingv1.TargetLowerThanRequests,
			},
		}
		newTestVPA := test.VerticalPodAutoscaler().
			WithName("test-vpa").
			WithContainer(container1Name).
			WithEvictionRequirements(newPodEvictionRequirements).
			Get()
		newVpaToPodMap := map[*vpaautoscalingv1.VerticalPodAutoscaler][]*corev1.Pod{newTestVPA: {pod2}}
		newExpectedEvictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod2: newPodEvictionRequirements,
		}
		sdpea.LoopInit(nil, newVpaToPodMap)
		assert.Len(t, storedEvictionRequirements(sdpea), 1)
		assert.Equal(t, storedEvictionRequirements(sdpea), newExpectedEvictionRequirements)
	})
}

func TestAdmitForSingleContainer(t *testing.T) {
	containerName := "test-container"
	pod := test.Pod().WithName("test-pod").
		AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("10Gi")).Get()).
		Get()
	podWithoutRequests := test.Pod().WithName("test-pod-without-requests").
		AddContainer(test.Container().WithName("test-container-without-requests").Get()).
		Get()

	t.Run("it should admit a Pod if no recommendation is present yet for a Pod", func(t *testing.T) {
		scalingDirectionPodEvictionAdmission := NewScalingDirectionPodEvictionAdmission()

		assert.Equal(t, true, scalingDirectionPodEvictionAdmission.Admit(pod, nil))
	})

	t.Run("it should admit a Pod for eviction if no resource request is present for a Pod", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{podWithoutRequests: {
			{
				Resources:         []corev1.ResourceName{corev1.ResourceCPU},
				ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
			},
		}}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer("test-container-without-requests").WithTarget("600m", "10Gi").Get()

		assert.Equal(t, true, sdpea.Admit(podWithoutRequests, recommendation))
	})

	t.Run("it should admit a Pod for eviction if recommendation is higher than ContainerStatus requests", func(t *testing.T) {
		podWithContainerStatus := pod.DeepCopy()
		podWithContainerStatus.Status.ContainerStatuses = []corev1.ContainerStatus{
			test.ContainerStatus().WithName(containerName).
				WithCPURequest(resource.MustParse("100m")).
				WithCPULimit(resource.MustParse("100m")).
				WithMemRequest(resource.MustParse("1Gi")).
				WithMemLimit(resource.MustParse("1Gi")).Get()}

		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			podWithContainerStatus: {
				{Resources: []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		// Recommendation is higher than ContainerStatus requests, but lower than
		// PodSpec requests. This should be admitted (eviction requirement is
		// TargetHigherThanRequests), given that ContainerStatus has priority over
		// PodSpec.
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("400m", "4Gi").Get()

		assert.Equal(t, true, sdpea.Admit(podWithContainerStatus, recommendation))
	})

	t.Run("it should admit a Pod for eviction if no config is given", func(t *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{pod: {}})
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("600m", "10Gi").Get()

		assert.Equal(t, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit a Pod for eviction if Container CPU is scaled up and config allows scaling up CPU", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("600m", "10Gi").Get()

		assert.Equal(t, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit a Pod for eviction if Container CPU is scaled down and config allows scaling down CPU", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetLowerThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("400m", "10Gi").Get()

		assert.Equal(t, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should not admit a Pod for eviction if Container CPU is scaled down and config allows only scaling up CPU", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{Resources: []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("500m", "10Gi").Get()

		assert.Equal(t, false, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit a Pod for eviction even if Container CPU is scaled down and config allows only scaling up CPU, because memory is scaled up", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{Resources: []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("500m", "11Gi").Get()

		assert.Equal(t, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit a Pod for eviction if Container memory is scaled up and config allows scaling up memory", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("500m", "11Gi").Get()

		assert.Equal(t, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit a Pod for eviction if Container memory is scaled down and config allows scaling down memory", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetLowerThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("500m", "9Gi").Get()

		assert.Equal(t, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should not admit a Pod for eviction if Container memory is scaled down and config allows only scaling up memory", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("500m", "9Gi").Get()

		assert.Equal(t, false, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit a Pod for eviction if Container CPU is scaled up, memory is scaled down and config allows scaling up CPU and scaling down memory", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
				{
					Resources:         []corev1.ResourceName{corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetLowerThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("600m", "9Gi").Get()

		assert.Equal(t, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should not admit a Pod for eviction if Container CPU is scaled up and config allows only scaling down CPU", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetLowerThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("1000m", "9Gi").Get()

		assert.Equal(t, false, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should not admit a Pod for eviction if Container CPU is scaled up, memory is scaled down and config allows scaling up CPU and scaling up memory", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
				{
					Resources:         []corev1.ResourceName{corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("600m", "9Gi").Get()

		assert.Equal(t, false, sdpea.Admit(pod, recommendation))
	})
}

func TestAdmitForMultipleContainer(t *testing.T) {
	container1Name := "test-container-1"
	container2Name := "test-container-2"
	pod := test.Pod().WithName("test-pod").
		AddContainer(test.Container().WithName(container1Name).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("10Gi")).Get()).
		AddContainer(test.Container().WithName(container2Name).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("10Gi")).Get()).
		Get()

	t.Run("it should admit the Pod if both containers fulfill the EvictionRequirements", func(tt *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().WithContainer(container1Name).WithTarget("600m", "10Gi").GetContainerResources(),
				test.Recommendation().WithContainer(container2Name).WithTarget("700m", "10Gi").GetContainerResources(),
			},
		}

		assert.Equal(tt, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit the Pod if only one container fulfills the EvictionRequirements", func(tt *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().WithContainer(container1Name).WithTarget("200m", "10Gi").GetContainerResources(),
				test.Recommendation().WithContainer(container2Name).WithTarget("700m", "10Gi").GetContainerResources(),
			},
		}

		assert.Equal(tt, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should not admit the Pod if no container fulfills the EvictionRequirements", func(tt *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().WithContainer(container1Name).WithTarget("200m", "10Gi").GetContainerResources(),
				test.Recommendation().WithContainer(container2Name).WithTarget("300m", "10Gi").GetContainerResources(),
			},
		}

		assert.Equal(tt, false, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should not admit the Pod even if there is a container that doesn't have a Recommendation and the other one doesn't fulfill the EvictionRequirements", func(tt *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		setEvictionRequirements(sdpea, evictionRequirements)
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().WithContainer(container2Name).WithTarget("300m", "10Gi").GetContainerResources(),
			},
		}

		assert.Equal(tt, false, sdpea.Admit(pod, recommendation))
	})
}

func TestAdmitTargetHigherThanRequestsAndOOM(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.EvictionRequirementOOMScaleUp, true)
	container1Name := "test-container-1"
	container2Name := "test-container-2"
	memoryOOMRequirement := &vpaautoscalingv1.EvictionRequirement{
		Resources:         []corev1.ResourceName{corev1.ResourceMemory},
		ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequestsAndOOM,
	}
	terminatedStatus := func(containerName, reason string, finishedAgo time.Duration) corev1.ContainerStatus {
		return corev1.ContainerStatus{
			Name: containerName,
			LastTerminationState: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{
					Reason:     reason,
					StartedAt:  metav1.NewTime(time.Now().Add(-finishedAgo - time.Minute)),
					FinishedAt: metav1.NewTime(time.Now().Add(-finishedAgo)),
				},
			},
		}
	}
	newPod := func(statuses ...corev1.ContainerStatus) *corev1.Pod {
		builder := test.Pod().WithName("test-pod").
			AddContainer(test.Container().WithName(container1Name).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("1Gi")).Get()).
			AddContainer(test.Container().WithName(container2Name).WithCPURequest(resource.MustParse("500m")).WithMemRequest(resource.MustParse("1Gi")).Get())
		for _, status := range statuses {
			builder = builder.AddContainerStatus(status)
		}
		return builder.Get()
	}
	newAdmission := func(pod *corev1.Pod, evictAfterOOMSeconds *int32, reqs []*vpaautoscalingv1.EvictionRequirement, opts ...ScalingDirectionPodEvictionAdmissionOption) PodEvictionAdmission {
		vpa := test.VerticalPodAutoscaler().
			WithName("test-vpa").
			WithContainer(container1Name).
			WithEvictionRequirements(reqs).
			WithEvictAfterOOMSeconds(evictAfterOOMSeconds).
			Get()
		sdpea := NewScalingDirectionPodEvictionAdmission(opts...)
		sdpea.LoopInit(nil, map[*vpaautoscalingv1.VerticalPodAutoscaler][]*corev1.Pod{vpa: {pod}})
		return sdpea
	}
	recommendation := func(container1Target, container2Target string) *vpaautoscalingv1.RecommendedPodResources {
		return &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().WithContainer(container1Name).WithTarget("500m", container1Target).GetContainerResources(),
				test.Recommendation().WithContainer(container2Name).WithTarget("500m", container2Target).GetContainerResources(),
			},
		}
	}

	t.Run("it should admit a Pod if memory is scaled up for a container that recently OOMed", func(t *testing.T) {
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", time.Minute))
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, true, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should admit a Pod if memory is scaled up for a container that is currently in terminated state due to OOM", func(t *testing.T) {
		status := terminatedStatus(container1Name, "OOMKilled", time.Minute)
		status.State.Terminated = status.LastTerminationState.Terminated
		status.LastTerminationState.Terminated = nil
		pod := newPod(status)
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, true, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should not admit a Pod if memory is scaled up for a container that has not OOMed", func(t *testing.T) {
		pod := newPod()
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, false, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should not admit a Pod if the container OOMed longer ago than the threshold", func(t *testing.T) {
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", 20*time.Minute))
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, false, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should not admit a Pod if the container was terminated for a reason other than OOM", func(t *testing.T) {
		pod := newPod(terminatedStatus(container1Name, "Error", time.Minute))
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, false, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should not admit a Pod if memory is scaled down for a container that recently OOMed", func(t *testing.T) {
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", time.Minute))
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, false, sdpea.Admit(pod, recommendation("512Mi", "1Gi")))
	})

	t.Run("it should not admit a Pod if the container that recently OOMed is not the one scaled up", func(t *testing.T) {
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", time.Minute))
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, false, sdpea.Admit(pod, recommendation("1Gi", "1.5Gi")))
	})

	t.Run("it should admit a Pod only if both EvictionRequirements are fulfilled", func(t *testing.T) {
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", time.Minute))
		reqs := []*vpaautoscalingv1.EvictionRequirement{
			memoryOOMRequirement,
			{
				Resources:         []corev1.ResourceName{corev1.ResourceCPU},
				ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
			},
		}
		sdpea := newAdmission(pod, nil, reqs)
		cpuAndMemoryScaledUp := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().WithContainer(container1Name).WithTarget("600m", "1.5Gi").GetContainerResources(),
			},
		}

		assert.Equal(t, true, sdpea.Admit(pod, cpuAndMemoryScaledUp))
		assert.Equal(t, false, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should use the threshold passed with WithEvictAfterOOMThreshold", func(t *testing.T) {
		thirtyMinutes := 30 * time.Minute
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", 20*time.Minute))
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement}, WithEvictAfterOOMThreshold(thirtyMinutes))

		assert.Equal(t, true, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should use the per-VPA evictAfterOOMSeconds when PerVPAConfig is enabled", func(t *testing.T) {
		thirtyMinutes := int32(30 * 60)
		featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.PerVPAConfig, true)
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", 20*time.Minute))
		sdpea := newAdmission(pod, ptr.To(thirtyMinutes), []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, true, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should ignore the per-VPA evictAfterOOMSeconds when PerVPAConfig is disabled", func(t *testing.T) {
		thirtyMinutes := int32(30 * 60)
		featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.PerVPAConfig, false)
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", 20*time.Minute))
		sdpea := newAdmission(pod, ptr.To(thirtyMinutes), []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, false, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})

	t.Run("it should not admit a Pod when EvictionRequirementOOMScaleUp is disabled", func(t *testing.T) {
		featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.EvictionRequirementOOMScaleUp, false)
		pod := newPod(terminatedStatus(container1Name, "OOMKilled", time.Minute))
		sdpea := newAdmission(pod, nil, []*vpaautoscalingv1.EvictionRequirement{memoryOOMRequirement})

		assert.Equal(t, false, sdpea.Admit(pod, recommendation("1.5Gi", "1Gi")))
	})
}

func setEvictionRequirements(admission PodEvictionAdmission, reqs map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement) {
	configs := make(map[*corev1.Pod]vpaEvictionConfig)
	for pod, podReqs := range reqs {
		configs[pod] = vpaEvictionConfig{requirements: podReqs}
	}
	admission.(*scalingDirectionPodEvictionAdmission).evictionConfigs = configs
}

func storedEvictionRequirements(admission PodEvictionAdmission) map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement {
	reqs := make(map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement)
	for pod, config := range admission.(*scalingDirectionPodEvictionAdmission).evictionConfigs {
		reqs[pod] = config.requirements
	}
	return reqs
}
