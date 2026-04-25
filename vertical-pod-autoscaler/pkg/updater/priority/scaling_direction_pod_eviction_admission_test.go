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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpaautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
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
		assert.Len(t, sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements, 0)
	})

	t.Run("it should store EvictionRequirements from VPA in a map per Pod", func(t *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.LoopInit(nil, vpaToPodMap)
		assert.Len(t, sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements, 2)
		assert.Equal(t, sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements, expectedEvictionRequirements)
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
		assert.Len(t, sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements, 2)
		assert.Equal(t, sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements, newExpectedEvictionRequirements)
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
		assert.Len(t, sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements, 1)
		assert.Equal(t, sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements, newExpectedEvictionRequirements)
	})
}

func TestAdmitForSingleContainer(t *testing.T) {
	containerName := "test-container"
	pod := test.Pod().WithName("test-pod").
		AddContainer(
			test.Container().
				WithName(containerName).
				WithCPURequest(resource.MustParse("500m")).
				WithMemRequest(resource.MustParse("10Gi")).Get()).
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
		// Recommendation is higher than ContainerStatus requests, but lower than
		// PodSpec requests. This should be admitted (eviction requirement is
		// TargetHigherThanRequests), given that ContainerStatus has priority over
		// PodSpec.
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("400m", "4Gi").Get()

		assert.Equal(t, true, sdpea.Admit(podWithContainerStatus, recommendation))
	})

	t.Run("it should admit a Pod for eviction if no config is given", func(t *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{pod: {}}
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("500m", "10Gi").Get()

		assert.Equal(t, false, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should not admit a Pod for eviction because the CPU recommendation violates the EvictionRequirements, even though the memory recommendation fulfills them.", func(t *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{Resources: []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
		recommendation := test.Recommendation().WithContainer(containerName).WithTarget("500m", "11Gi").Get()

		assert.Equal(t, false, sdpea.Admit(pod, recommendation))
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
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
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().WithContainer(container2Name).WithTarget("300m", "10Gi").GetContainerResources(),
			},
		}

		assert.Equal(tt, false, sdpea.Admit(pod, recommendation))
	})
}

func TestAdmitForPodLevelResources(t *testing.T) {
	container1Name := "test-container-1"
	container2Name := "test-container-2"
	pod := test.Pod().
		WithName("test-pod").
		AddContainer(
			test.Container().
				WithName(container1Name).
				Get()).
		AddContainer(
			test.Container().
				WithName(container2Name).
				Get()).
		WithCPURequest(resource.MustParse("200m")).
		WithMemRequest(resource.MustParse("200Mi")).
		Get()

	podWithoutPodLevelResources := test.Pod().
		WithName("test-pod").
		AddContainer(
			test.Container().
				WithName(container1Name).
				Get()).
		AddContainer(
			test.Container().
				WithName(container2Name).
				Get()).
		Get()

	t.Run("it should admit pod for eviction if config is provided through a nil map", func(tt *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = nil
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: nil,
			PodRecommendations: test.Recommendation().
				WithPodLevelTarget("155m", "155Mi").
				GetPodLevelRecommendations(),
		}
		assert.Equal(tt, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit pod for eviction if config is provided through a map with nil value", func(tt *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{pod: {}}
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: nil,
			PodRecommendations: test.Recommendation().
				WithPodLevelTarget("155m", "155Mi").
				GetPodLevelRecommendations(),
		}
		assert.Equal(tt, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit pod for eviction if pod level resource does not exist but pod level recommendation does", func(tt *testing.T) {
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = nil
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: nil,
			PodRecommendations: test.Recommendation().
				WithPodLevelTarget("200m", "200Mi").
				GetPodLevelRecommendations(),
		}
		assert.Equal(tt, true, sdpea.Admit(podWithoutPodLevelResources, recommendation))
	})

	t.Run("it should admit pod for eviction if pod level recommendation fulfills EvictionRequirements through PodState requests", func(tt *testing.T) {
		podWithPodStatus := podWithoutPodLevelResources.DeepCopy()
		podWithPodStatus.Status = test.PodStatus().
			WithCPURequest(resource.MustParse("199m")).
			WithMemRequest(resource.MustParse("199Mi")).Get()

		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			podWithPodStatus: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: nil,
			PodRecommendations: test.Recommendation().
				WithPodLevelTarget("200m", "200Mi").
				GetPodLevelRecommendations(),
		}
		assert.Equal(tt, true, sdpea.Admit(podWithPodStatus, recommendation))
	})

	// TODO: Comment out this part once https://github.com/kubernetes/kubernetes/issues/137628 is fixed and backported to Kubernetes versions.
	// t.Run("it should not admit pod for eviction if pod level cpu recommendation violates EvictionRequirements through PodState requests", func(tt *testing.T) {
	// 	podWithPodStatus := podWithoutPodLevelResources.DeepCopy()
	// 	podWithPodStatus.Status = test.PodStatus().
	// 		WithCPURequest(resource.MustParse("9999m")).
	// 		WithMemRequest(resource.MustParse("199Mi")).Get()

	// 	evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
	// 		podWithPodStatus: {
	// 			{
	// 				Resources:         []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
	// 				ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
	// 			},
	// 		},
	// 	}
	// 	sdpea := NewScalingDirectionPodEvictionAdmission()
	// 	sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
	// 	recommendation := &vpaautoscalingv1.RecommendedPodResources{
	// 		ContainerRecommendations: nil,
	// 		PodRecommendations: test.Recommendation().
	// 			WithPodLevelTarget("200m", "200Mi").
	// 			GetPodLevelRecommendations(),
	// 	}
	// 	assert.Equal(tt, false, sdpea.Admit(podWithPodStatus, recommendation))
	// })

	// TODO: Comment out this part once https://github.com/kubernetes/kubernetes/issues/137628 is fixed and backported to Kubernetes versions.
	// t.Run("it should not admit a pod for eviction if pod level recommendation violates EvictionRequirements through PodState requests", func(tt *testing.T) {
	// 	podWithPodStatus := podWithoutPodLevelResources.DeepCopy()
	// 	podWithPodStatus.Status = test.PodStatus().
	// 		WithCPURequest(resource.MustParse("9999m")).
	// 		WithMemRequest(resource.MustParse("9999Mi")).Get()

	// 	evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
	// 		podWithPodStatus: {
	// 			{
	// 				Resources:         []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
	// 				ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
	// 			},
	// 		},
	// 	}
	// 	sdpea := NewScalingDirectionPodEvictionAdmission()
	// 	sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
	// 	recommendation := &vpaautoscalingv1.RecommendedPodResources{
	// 		ContainerRecommendations: nil,
	// 		PodRecommendations: test.Recommendation().
	// 			WithPodLevelTarget("200m", "200Mi").
	// 			GetPodLevelRecommendations(),
	// 	}
	// 	assert.Equal(tt, false, sdpea.Admit(podWithPodStatus, recommendation))
	// })

	t.Run("it should admit a Pod for eviction if pod level resource fulfills EvictionRequirements through PodSpec", func(tt *testing.T) {
		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			pod: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: nil,
			PodRecommendations: test.Recommendation().
				WithPodLevelTarget("201m", "201Mi").
				GetPodLevelRecommendations(),
		}
		assert.Equal(tt, true, sdpea.Admit(pod, recommendation))
	})

	t.Run("it should admit a Pod for eviction if pod level resource fulfills EvictionRequirements through PodSpec and container does not", func(tt *testing.T) {
		podWithExtraContainer := pod.DeepCopy()
		podWithExtraContainer.Spec.Containers = append(podWithExtraContainer.Spec.Containers,
			test.Container().WithName("test-container-3").
				WithCPURequest(resource.MustParse("20m")).
				WithMemRequest(resource.MustParse("20Mi")).Get(),
		)

		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			podWithExtraContainer: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().
					WithContainer("test-container-3").
					WithTarget("1m", "1Mi").
					GetContainerResources(),
			},
			PodRecommendations: test.Recommendation().
				WithPodLevelTarget("201m", "201Mi").
				GetPodLevelRecommendations(),
		}
		assert.Equal(tt, true, sdpea.Admit(podWithExtraContainer, recommendation))
	})

	t.Run("it should admit a Pod for eviction if pod level resource does not fulfill EvictionRequirements through PodSpec and container does", func(tt *testing.T) {
		podWithExtraContainer := pod.DeepCopy()
		podWithExtraContainer.Spec.Containers = append(podWithExtraContainer.Spec.Containers,
			test.Container().WithName("test-container-3").
				WithCPURequest(resource.MustParse("20m")).
				WithMemRequest(resource.MustParse("20Mi")).Get(),
		)

		evictionRequirements := map[*corev1.Pod][]*vpaautoscalingv1.EvictionRequirement{
			podWithExtraContainer: {
				{
					Resources:         []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory},
					ChangeRequirement: vpaautoscalingv1.TargetHigherThanRequests,
				},
			},
		}
		sdpea := NewScalingDirectionPodEvictionAdmission()
		sdpea.(*scalingDirectionPodEvictionAdmission).EvictionRequirements = evictionRequirements
		recommendation := &vpaautoscalingv1.RecommendedPodResources{
			ContainerRecommendations: []vpaautoscalingv1.RecommendedContainerResources{
				test.Recommendation().
					WithContainer("test-container-3").
					WithTarget("21m", "21Mi").
					GetContainerResources(),
			},
			PodRecommendations: test.Recommendation().
				WithPodLevelTarget("1m", "1Mi").
				GetPodLevelRecommendations(),
		}
		assert.Equal(tt, true, sdpea.Admit(podWithExtraContainer, recommendation))
	})
}
