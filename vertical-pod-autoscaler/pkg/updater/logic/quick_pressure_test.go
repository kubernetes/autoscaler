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

package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/mock/gomock"
	"golang.org/x/time/rate"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	target_mock "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/mock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/restriction"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

func TestClampToCurrentRequests(t *testing.T) {
	vpa := test.VerticalPodAutoscaler().WithContainer("c").WithTarget("2", "700Mi").Get()
	pod := test.Pod().WithName("p").AddContainer(test.Container().WithName("c").
		WithCPURequest(resource.MustParse("4")).WithMemRequest(resource.MustParse("256Mi")).Get()).Get()
	c := clampToCurrentRequests(vpa, pod)
	tgt := c.Status.Recommendation.ContainerRecommendations[0].Target
	assert.Equal(t, "4", tgt.Cpu().String(), "CPU decrease clamped to the current request")
	assert.Equal(t, "700Mi", tgt.Memory().String(), "memory increase kept")
	assert.Equal(t, "2", vpa.Status.Recommendation.ContainerRecommendations[0].Target.Cpu().String(), "original VPA untouched")
}

func TestRunOnce_QuickPressure(t *testing.T) {
	testCases := []struct {
		name             string
		gateEnabled      bool
		youngPod         bool
		canInPlaceUpdate utils.InPlaceDecision
		inPlaceFails     bool
		// cappedCPU, when set, is the CPU target the recommendation processor returns after capping.
		cappedCPU         string
		expectedInPlace   int
		expectedEvictions int
	}{
		{
			name:              "pressure-only pod whose resize fails is not evicted",
			gateEnabled:       true,
			youngPod:          true,
			canInPlaceUpdate:  utils.InPlaceApproved,
			inPlaceFails:      true,
			expectedInPlace:   1,
			expectedEvictions: 0,
		},
		{
			name:              "pressure-only pod that cannot be resized in place is not evicted",
			gateEnabled:       true,
			youngPod:          true,
			canInPlaceUpdate:  utils.InPlaceEvict,
			expectedInPlace:   0,
			expectedEvictions: 0,
		},
		{
			name:              "pressure-only pod is resized in place",
			gateEnabled:       true,
			youngPod:          true,
			canInPlaceUpdate:  utils.InPlaceApproved,
			expectedInPlace:   1,
			expectedEvictions: 0,
		},
		{
			// maxAllowed below the current CPU request would lower CPU even after the clamp.
			name:              "pressure-only pod is skipped when capping would lower a resource",
			gateEnabled:       true,
			youngPod:          true,
			canInPlaceUpdate:  utils.InPlaceApproved,
			cappedCPU:         "500m",
			expectedInPlace:   0,
			expectedEvictions: 0,
		},
		{
			name:              "gate off: young pod is not admitted",
			gateEnabled:       false,
			youngPod:          true,
			canInPlaceUpdate:  utils.InPlaceApproved,
			expectedInPlace:   0,
			expectedEvictions: 0,
		},
		{
			name:              "pod admitted by the lifetime gate still falls back to eviction",
			gateEnabled:       true,
			youngPod:          false,
			canInPlaceUpdate:  utils.InPlaceApproved,
			inPlaceFails:      true,
			expectedInPlace:   1,
			expectedEvictions: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.ReactiveMemoryPressureDetection, tc.gateEnabled)
			ctrl := gomock.NewController(t)

			containerName := "container1"
			rc := corev1.ReplicationController{
				TypeMeta:   metav1.TypeMeta{Kind: "ReplicationController", APIVersion: "apps/v1"},
				ObjectMeta: metav1.ObjectMeta{Name: "rc", Namespace: "default"},
			}
			// The memory request is inside [lowerBound, upperBound] but below target, so only the lifetime
			// gate or quick pressure can admit the pod.
			vpaObj := test.VerticalPodAutoscaler().
				WithContainer(containerName).
				WithTarget("1", "200M").
				WithLowerBound("1", "50M").
				WithUpperBound("1", "300M").
				WithUpdateMode(vpa_types.UpdateModeInPlaceOrRecreate).
				WithTargetRef(&autoscalingv1.CrossVersionObjectReference{Kind: rc.Kind, Name: rc.Name, APIVersion: rc.APIVersion}).
				Get()
			enabled := vpa_types.PressureDetectionEnabled
			vpaObj.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{
				{ContainerName: containerName, PressureDetection: &enabled},
			}}
			stamp := metav1.NewTime(time.Now().Add(-time.Minute))
			vpaObj.Status.Recommendation.ContainerRecommendations[0].LastPressureTime = &stamp

			pod := test.Pod().WithName("test_0").
				AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).
				WithCreator(&rc.ObjectMeta, &rc.TypeMeta).
				Get()
			pod.Labels = map[string]string{"app": "testingApp"}
			if tc.youngPod {
				pod.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Minute)}
			}

			eviction := &test.PodsEvictionRestrictionMock{}
			eviction.On("CanEvict", pod).Return(true)
			eviction.On("Evict", pod, nil).Return(nil)
			inplace := &test.PodsInPlaceRestrictionMock{}
			inplace.On("CanInPlaceUpdate", pod).Return(tc.canInPlaceUpdate)
			if tc.inPlaceFails {
				inplace.On("InPlaceUpdate", pod, mock.Anything).Return(errors.New("in-place update failed"))
			} else {
				inplace.On("InPlaceUpdate", pod, mock.Anything).Return(nil)
			}
			vpaLister := &test.VerticalPodAutoscalerListerMock{}
			vpaLister.On("List").Return([]*vpa_types.VerticalPodAutoscaler{vpaObj}, nil).Once()
			podLister := &test.PodListerMock{}
			podLister.On("List").Return([]*corev1.Pod{pod}, nil)
			mockSelectorFetcher := target_mock.NewMockVpaTargetSelectorFetcher(ctrl)
			mockSelectorFetcher.EXPECT().Fetch(gomock.Eq(vpaObj)).Return(parseLabelSelector("app = testingApp"), nil)

			var processor vpa_api_util.RecommendationProcessor = &test.FakeRecommendationProcessor{}
			if tc.cappedCPU != "" {
				capped := &test.RecommendationProcessorMock{}
				capped.On("Apply").Return(test.Recommendation().WithContainer(containerName).
					WithTarget(tc.cappedCPU, "200M").WithLowerBound(tc.cappedCPU, "50M").WithUpperBound("2", "300M").Get(), nil)
				processor = capped
			}
			updater := &updater{
				vpaLister:                    vpaLister,
				podLister:                    podLister,
				restrictionFactory:           &restriction.FakePodsRestrictionFactory{Eviction: eviction, InPlace: inplace},
				evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
				inPlaceRateLimiter:           rate.NewLimiter(rate.Inf, 0),
				evictionAdmission:            priority.NewDefaultPodEvictionAdmission(),
				recommendationProcessor:      processor,
				selectorFetcher:              mockSelectorFetcher,
				controllerFetcher:            controllerfetcher.FakeControllerFetcher{},
				useAdmissionControllerStatus: true,
				statusValidator:              newFakeValidator(true),
				priorityProcessor:            priority.NewProcessor(),
				podLifetimeUpdateThreshold:   12 * time.Hour,
				pressureQuickUpdateWindow:    10 * time.Minute,
				pressureQuickUpdateFraction:  0.1,
			}
			updater.RunOnce(context.Background())
			inplace.AssertNumberOfCalls(t, "InPlaceUpdate", tc.expectedInPlace)
			eviction.AssertNumberOfCalls(t, "Evict", tc.expectedEvictions)
		})
	}
}

func TestGetPodsUpdateOrderQuickPressureBudget(t *testing.T) {
	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.ReactiveMemoryPressureDetection, true)
	containerName := "container1"
	vpaObj := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithTarget("1", "200M").
		WithLowerBound("1", "50M").
		WithUpperBound("1", "300M").
		WithUpdateMode(vpa_types.UpdateModeInPlace).
		Get()
	enabled := vpa_types.PressureDetectionEnabled
	vpaObj.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{
		{ContainerName: containerName, PressureDetection: &enabled},
	}}
	stamp := metav1.NewTime(time.Now().Add(-time.Minute))
	vpaObj.Status.Recommendation.ContainerRecommendations[0].LastPressureTime = &stamp
	var candidates []*corev1.Pod
	for _, name := range []string{"a", "b"} {
		pod := test.Pod().WithName(name).
			AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).
			Get()
		pod.UID = types.UID(name)
		pod.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Minute)}
		candidates = append(candidates, pod)
	}
	u := &updater{
		recommendationProcessor:     &test.FakeRecommendationProcessor{},
		priorityProcessor:           priority.NewProcessor(),
		evictionAdmission:           priority.NewDefaultPodEvictionAdmission(),
		podLifetimeUpdateThreshold:  12 * time.Hour,
		pressureQuickUpdateWindow:   10 * time.Minute,
		pressureQuickUpdateFraction: 0.1,
	}
	// Two candidates out of 20 controlled Pods: 10% of 20 admits both, 10% of the 2 candidates would admit one.
	_, pressureOnly := u.getPodsUpdateOrder(candidates, vpaObj, 20)
	assert.Len(t, pressureOnly, 2)
}
