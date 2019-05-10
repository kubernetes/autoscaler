/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestUpdateResourceLimits(t *testing.T) {
	type testCase struct {
		pod                         *apiv1.Pod
		containerResources          []ContainerResources
		limitRanges                 []runtime.Object
		requestsExceedsRatioCPU     bool
		requestsExceedsRatioMemory  bool
		limitsRespectingRatioCPU    resource.Quantity
		limitsRespectingRatioMemory resource.Quantity
	}
	containerName := "container1"
	vpaName := "vpa1"
	labels := map[string]string{"app": "testingApp"}

	minRatio := test.Resources("5", "5")

	limitranges := []runtime.Object{
		&apiv1.LimitRange{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "limitRange-with-default-and-ratio",
			},
			Spec: apiv1.LimitRangeSpec{
				Limits: []apiv1.LimitRangeItem{
					{
						Type:    apiv1.LimitTypeContainer,
						Default: test.Resources("2000m", "2Gi"),
					},
					{
						Type:                 apiv1.LimitTypePod,
						MaxLimitRequestRatio: test.Resources("10", "10"),
					},
				},
			},
		},
		&apiv1.LimitRange{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "limitRange-with-only-ratio",
			},
			Spec: apiv1.LimitRangeSpec{
				Limits: []apiv1.LimitRangeItem{
					{
						Type:                 apiv1.LimitTypePod,
						MaxLimitRequestRatio: minRatio,
					},
				},
			},
		},
	}

	uninitialized := test.Pod().WithName("test_uninitialized").AddContainer(test.BuildTestContainer(containerName, "", "")).Get()
	uninitialized.ObjectMeta.Labels = labels

	initialized := test.Pod().WithName("test_initialized").AddContainer(test.BuildTestContainer(containerName, "1", "100Mi")).Get()
	initialized.ObjectMeta.Labels = labels

	withLimits := test.Pod().WithName("test_initialized").AddContainer(test.BuildTestContainer(containerName, "1", "100Mi")).Get()
	withLimits.ObjectMeta.Labels = labels
	withLimits.Spec.Containers[0].Resources.Limits = test.Resources("1500m", "800Mi")

	withHugeMemLimits := test.Pod().WithName("test_initialized").AddContainer(test.BuildTestContainer(containerName, "1", "10Gi")).Get()
	withHugeMemLimits.ObjectMeta.Labels = labels
	withHugeMemLimits.Spec.Containers[0].Resources.Limits = test.Resources("1500m", "80Gi")

	vpaBuilder := test.VerticalPodAutoscaler().
		WithName(vpaName).
		WithContainer(containerName).
		WithTarget("20m", "200Mi")
	vpa := vpaBuilder.Get()

	vpaWithHighMemory := vpaBuilder.WithTarget("2", "3Gi").Get()

	// short circuit recommendation provider
	vpaContainersResources := []ContainerResources{{
		Requests: vpa.Status.Recommendation.ContainerRecommendations[0].Target,
	}}
	vpaHighMemContainersResources := []ContainerResources{{
		Requests: vpaWithHighMemory.Status.Recommendation.ContainerRecommendations[0].Target,
	}}

	expectedMemory := func(crs []ContainerResources, ratio apiv1.ResourceList) resource.Quantity {
		return *resource.NewQuantity(
			int64(float64(
				crs[0].Requests.Memory().Value())*float64(ratio.Memory().Value())),
			crs[0].Requests.Memory().Format)
	}
	expectedCPU := func(crs []ContainerResources, ratio apiv1.ResourceList) resource.Quantity {
		return *resource.NewMilliQuantity(
			int64(float64(
				crs[0].Requests.Cpu().MilliValue())*float64(ratio.Cpu().Value())),
			crs[0].Requests.Cpu().Format)
	}

	testCases := []testCase{{
		pod:                         uninitialized,
		containerResources:          vpaContainersResources,
		limitRanges:                 limitranges,
		requestsExceedsRatioCPU:     true,
		requestsExceedsRatioMemory:  true,
		limitsRespectingRatioCPU:    expectedCPU(vpaContainersResources, minRatio),
		limitsRespectingRatioMemory: expectedMemory(vpaContainersResources, minRatio),
	}, {
		pod:                         initialized,
		containerResources:          vpaContainersResources,
		limitRanges:                 limitranges,
		requestsExceedsRatioCPU:     true,
		requestsExceedsRatioMemory:  true,
		limitsRespectingRatioCPU:    expectedCPU(vpaContainersResources, minRatio),
		limitsRespectingRatioMemory: expectedMemory(vpaContainersResources, minRatio),
	}, {
		pod:                         withLimits,
		containerResources:          vpaContainersResources,
		limitRanges:                 limitranges,
		requestsExceedsRatioCPU:     true,
		requestsExceedsRatioMemory:  false,
		limitsRespectingRatioCPU:    expectedCPU(vpaContainersResources, minRatio),
		limitsRespectingRatioMemory: resource.Quantity{},
	}, {
		pod:                         withHugeMemLimits,
		containerResources:          vpaHighMemContainersResources,
		limitRanges:                 limitranges,
		requestsExceedsRatioCPU:     false,
		requestsExceedsRatioMemory:  true,
		limitsRespectingRatioCPU:    resource.Quantity{},
		limitsRespectingRatioMemory: expectedMemory(vpaHighMemContainersResources, minRatio),
	}}

	// if admission controller is not allowed to adjust limits
	// the limits checher have to return always:
	// - no needed limits
	// - RequestsExceedsRatio always return false
	t.Run("test case for neverNeedsLimitsChecker", func(t *testing.T) {
		nlc := NewLimitsChecker(nil)
		hints := nlc.NeedsLimits(uninitialized, vpaContainersResources)
		hintsPtr, _ := hints.(*LimitRangeHints)
		if hintsPtr != nil {
			t.Errorf("%v NeedsLimits didn't not return nil: %v", nlc, hints)
		}
		if !hints.IsNil() {
			t.Errorf("%v NeedsLimits returned a LimitsHints not nil: %v", nlc, hints)
		}
		if hints.RequestsExceedsRatio(0, apiv1.ResourceMemory) != false {
			t.Errorf("%v RequestsExceedsRatio didn't not return false", hints)
		}
		hinted := hints.HintedLimit(0, apiv1.ResourceMemory)
		if !(&hinted).IsZero() {
			t.Errorf("%v RequestsExceedsRatio didn't not return zero quantity", hints)
		}
	})

	t.Run("test case for no Limit Range", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		factory := informers.NewSharedInformerFactory(cs, 0)
		lc := NewLimitsChecker(factory)
		hints := lc.NeedsLimits(uninitialized, vpaContainersResources)
		hintsPtr, _ := hints.(*LimitRangeHints)
		if hintsPtr != nil {
			t.Errorf("%v NeedsLimits didn't not return nil: %v", lc, hints)
		}
		if !hints.IsNil() {
			t.Errorf("%v NeedsLimits returned a LimitsHints not nil: %v", lc, hints)
		}
		if hints.RequestsExceedsRatio(0, apiv1.ResourceMemory) != false {
			t.Errorf("%v RequestsExceedsRatio didn't not return false", hints)
		}
		hinted := hints.HintedLimit(0, apiv1.ResourceMemory)
		if !(&hinted).IsZero() {
			t.Errorf("%v RequestsExceedsRatio didn't not return zero quantity", hints)
		}
	})

	for i, tc := range testCases {

		t.Run(fmt.Sprintf("test case number: %d", i), func(t *testing.T) {
			cs := fake.NewSimpleClientset(tc.limitRanges...)
			factory := informers.NewSharedInformerFactory(cs, 0)
			lc := NewLimitsChecker(factory)
			resources := tc.containerResources

			hints := lc.NeedsLimits(tc.pod, resources)
			assert.NotNil(t, hints, fmt.Sprintf("hints is: %+v", hints))

			if tc.requestsExceedsRatioCPU {
				assert.True(t, hints.RequestsExceedsRatio(0, apiv1.ResourceCPU))
			} else {
				assert.False(t, hints.RequestsExceedsRatio(0, apiv1.ResourceCPU))
			}

			if tc.requestsExceedsRatioMemory {
				assert.True(t, hints.RequestsExceedsRatio(0, apiv1.ResourceMemory))
			} else {
				assert.False(t, hints.RequestsExceedsRatio(0, apiv1.ResourceMemory))
			}

			hintedCPULimits := hints.HintedLimit(0, apiv1.ResourceCPU)
			hintedMemoryLimits := hints.HintedLimit(0, apiv1.ResourceMemory)
			assert.EqualValues(t, tc.limitsRespectingRatioCPU.Value(), hintedCPULimits.Value(), fmt.Sprintf("cpu limits doesn't match: %v != %v\n", tc.limitsRespectingRatioCPU.Value(), hintedCPULimits.Value()))
			assert.EqualValues(t, tc.limitsRespectingRatioMemory.Value(), hintedMemoryLimits.Value(), fmt.Sprintf("memory limits doesn't match: %v != %v\n", tc.limitsRespectingRatioMemory.Value(), hintedMemoryLimits.Value()))
		})

	}
}
