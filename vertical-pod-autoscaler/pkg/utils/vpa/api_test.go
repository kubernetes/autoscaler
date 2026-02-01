/*
Copyright 2018 The Kubernetes Authors.

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

package api

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/autoscaling/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_fake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

const (
	containerName = "container1"
)

var (
	anytime = time.Unix(0, 0)
)

func init() {
	flag.Set("alsologtostderr", "true") //nolint:errcheck
	flag.Set("v", "5")                  //nolint:errcheck
}

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := meta.ParseToLabelSelector(selector)
	parsedSelector, _ := meta.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

func TestUpdateVpaIfNeeded(t *testing.T) {
	updatedVpa := test.VerticalPodAutoscaler().WithName("vpa").WithNamespace("test").WithContainer(containerName).
		AppendCondition(vpa_types.RecommendationProvided, core.ConditionTrue, "reason", "msg", anytime).Get()
	recommendation := test.Recommendation().WithContainer(containerName).WithTarget("5", "200").Get()
	updatedVpa.Status.Recommendation = recommendation
	observedVpaBuilder := test.VerticalPodAutoscaler().WithName("vpa").WithNamespace("test").WithContainer(containerName)

	testCases := []struct {
		caseName       string
		updatedVpa     *vpa_types.VerticalPodAutoscaler
		observedVpa    *vpa_types.VerticalPodAutoscaler
		expectedUpdate bool
	}{
		{
			caseName:   "Doesn't update if no changes.",
			updatedVpa: updatedVpa,
			observedVpa: observedVpaBuilder.WithTarget("5", "200").
				AppendCondition(vpa_types.RecommendationProvided, core.ConditionTrue, "reason", "msg", anytime).Get(),
			expectedUpdate: false,
		}, {
			caseName:   "Updates on recommendation change.",
			updatedVpa: updatedVpa,
			observedVpa: observedVpaBuilder.WithTarget("10", "200").
				AppendCondition(vpa_types.RecommendationProvided, core.ConditionTrue, "reason", "msg", anytime).Get(),
			expectedUpdate: true,
		}, {
			caseName:   "Updates on condition change.",
			updatedVpa: updatedVpa,
			observedVpa: observedVpaBuilder.WithTarget("5", "200").
				AppendCondition(vpa_types.RecommendationProvided, core.ConditionFalse, "reason", "msg", anytime).Get(),
			expectedUpdate: true,
		}, {
			caseName:   "Updates on condition added.",
			updatedVpa: updatedVpa,
			observedVpa: observedVpaBuilder.WithTarget("5", "200").
				AppendCondition(vpa_types.RecommendationProvided, core.ConditionTrue, "reason", "msg", anytime).
				AppendCondition(vpa_types.LowConfidence, core.ConditionTrue, "reason", "msg", anytime).Get(),
			expectedUpdate: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			fakeClient := vpa_fake.NewSimpleClientset(&vpa_types.VerticalPodAutoscalerList{Items: []vpa_types.VerticalPodAutoscaler{*tc.observedVpa}}) //nolint:staticcheck // https://github.com/kubernetes/autoscaler/issues/8954
			_, err := UpdateVpaStatusIfNeeded(fakeClient.AutoscalingV1().VerticalPodAutoscalers(tc.updatedVpa.Namespace),
				tc.updatedVpa.Name, &tc.updatedVpa.Status, &tc.observedVpa.Status)
			assert.NoError(t, err, "Unexpected error occurred.")
			actions := fakeClient.Actions()
			if tc.expectedUpdate {
				assert.Equal(t, 1, len(actions), "Unexpected number of actions")
			} else {
				assert.Equal(t, 0, len(actions), "Unexpected number of actions")
			}
		})
	}
}

func TestPodMatchesVPA(t *testing.T) {
	type testCase struct {
		pod             *core.Pod
		vpaWithSelector VpaWithSelector
		result          bool
	}

	pod := test.Pod().WithName("test-pod").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).Get()
	pod.Labels = map[string]string{"app": "testingApp"}

	vpaBuilder := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithTarget("2", "200M").
		WithMinAllowed(containerName, "1", "100M").
		WithMaxAllowed(containerName, "3", "1G")

	vpa := vpaBuilder.Get()
	otherNamespaceVPA := vpaBuilder.WithNamespace("other").Get()

	testCases := []testCase{
		{pod, VpaWithSelector{vpa, parseLabelSelector("app = testingApp")}, true},
		{pod, VpaWithSelector{otherNamespaceVPA, parseLabelSelector("app = testingApp")}, false},
		{pod, VpaWithSelector{vpa, parseLabelSelector("app = other")}, false}}

	for _, tc := range testCases {
		actual := PodMatchesVPA(tc.pod, &tc.vpaWithSelector)
		assert.Equal(t, tc.result, actual)
	}
}

func TestGetControllingVPAForPod(t *testing.T) {
	ctx := context.Background()

	pod := test.Pod().WithName("test-pod").AddContainer(test.Container().WithName(containerName).WithCPURequest(resource.MustParse("1")).WithMemRequest(resource.MustParse("100M")).Get()).Get()
	pod.Labels = map[string]string{"app": "testingApp"}
	pod.OwnerReferences = []meta.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
			Name:       "test-sts",
			Controller: ptr.To(true),
		},
	}

	vpaBuilder := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithTarget("2", "200M").
		WithMinAllowed(containerName, "1", "100M").
		WithMaxAllowed(containerName, "3", "1G")
	vpaA := vpaBuilder.WithCreationTimestamp(time.Unix(5, 0)).Get()
	vpaB := vpaBuilder.WithCreationTimestamp(time.Unix(10, 0)).Get()
	nonMatchingVPA := vpaBuilder.WithCreationTimestamp(time.Unix(2, 0)).Get()
	vpaA.Spec.TargetRef = &apiv1.CrossVersionObjectReference{
		Kind:       "StatefulSet",
		Name:       "test-sts",
		APIVersion: "apps/v1",
	}
	chosen := GetControllingVPAForPod(ctx, pod, []*VpaWithSelector{
		{vpaB, parseLabelSelector("app = testingApp")},
		{vpaA, parseLabelSelector("app = testingApp")},
		{nonMatchingVPA, parseLabelSelector("app = other")},
	}, &controllerfetcher.FakeControllerFetcher{})
	assert.Equal(t, vpaA, chosen.Vpa)

	// For some Pods (which are *not* under VPA), controllerFetcher.FindTopMostWellKnownOrScalable will return `nil`, e.g. when the Pod owner is a custom resource, which doesn't implement the /scale subresource
	// See pkg/target/controller_fetcher/controller_fetcher_test.go:393 for testing this behavior
	// This test case makes sure that GetControllingVPAForPod will just return `nil` in that case as well
	chosen = GetControllingVPAForPod(ctx, pod, []*VpaWithSelector{{vpaA, parseLabelSelector("app = testingApp")}}, &controllerfetcher.NilControllerFetcher{})
	assert.Nil(t, chosen)
}

func TestGetContainerResourcePolicy(t *testing.T) {
	containerPolicy1 := vpa_types.ContainerResourcePolicy{
		ContainerName: "container1",
		MinAllowed: core.ResourceList{
			core.ResourceCPU: *resource.NewScaledQuantity(10, 1),
		},
	}
	containerPolicy2 := vpa_types.ContainerResourcePolicy{
		ContainerName: "container2",
		MaxAllowed: core.ResourceList{
			core.ResourceMemory: *resource.NewScaledQuantity(100, 1),
		},
	}
	policy := vpa_types.PodResourcePolicy{
		ContainerPolicies: []vpa_types.ContainerResourcePolicy{
			containerPolicy1, containerPolicy2,
		},
	}
	assert.Equal(t, &containerPolicy1, GetContainerResourcePolicy("container1", &policy))
	assert.Equal(t, &containerPolicy2, GetContainerResourcePolicy("container2", &policy))
	assert.Nil(t, GetContainerResourcePolicy("container3", &policy))

	// Add the wildcard ("*") policy.
	defaultPolicy := vpa_types.ContainerResourcePolicy{
		ContainerName: "*",
		MinAllowed: core.ResourceList{
			core.ResourceCPU: *resource.NewScaledQuantity(20, 1),
		},
	}
	policy = vpa_types.PodResourcePolicy{
		ContainerPolicies: []vpa_types.ContainerResourcePolicy{
			containerPolicy1, defaultPolicy, containerPolicy2,
		},
	}
	assert.Equal(t, &containerPolicy1, GetContainerResourcePolicy("container1", &policy))
	assert.Equal(t, &containerPolicy2, GetContainerResourcePolicy("container2", &policy))
	assert.Equal(t, &defaultPolicy, GetContainerResourcePolicy("container3", &policy))
}

func TestGetContainerControlledResources(t *testing.T) {
	requestsAndLimits := vpa_types.ContainerControlledValuesRequestsAndLimits
	requestsOnly := vpa_types.ContainerControlledValuesRequestsOnly
	for _, tc := range []struct {
		name          string
		containerName string
		policy        *vpa_types.PodResourcePolicy
		expected      vpa_types.ContainerControlledValues
	}{
		{
			name:          "default policy is RequestAndLimits",
			containerName: "any",
			policy:        nil,
			expected:      vpa_types.ContainerControlledValuesRequestsAndLimits,
		}, {
			name:          "container default policy is RequestsAndLimits",
			containerName: "any",
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsAndLimits,
				}},
			},
			expected: vpa_types.ContainerControlledValuesRequestsAndLimits,
		}, {
			name:          "container default policy is RequestsOnly",
			containerName: "any",
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsOnly,
				}},
			},
			expected: vpa_types.ContainerControlledValuesRequestsOnly,
		}, {
			name:          "RequestAndLimits is used when no policy for given container specified",
			containerName: "other",
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    "some",
					ControlledValues: &requestsOnly,
				}},
			},
			expected: vpa_types.ContainerControlledValuesRequestsAndLimits,
		}, {
			name:          "RequestsOnly specified explicitly",
			containerName: "some",
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    "some",
					ControlledValues: &requestsOnly,
				}},
			},
			expected: vpa_types.ContainerControlledValuesRequestsOnly,
		}, {
			name:          "RequestsAndLimits specified explicitly overrides default",
			containerName: "some",
			policy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
					ContainerName:    vpa_types.DefaultContainerResourcePolicy,
					ControlledValues: &requestsOnly,
				}, {
					ContainerName:    "some",
					ControlledValues: &requestsAndLimits,
				}},
			},
			expected: vpa_types.ContainerControlledValuesRequestsAndLimits,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := GetContainerControlledValues(tc.containerName, tc.policy)
			assert.Equal(t, got, tc.expected)
		})
	}
}

func TestFindParentControllerForPod(t *testing.T) {
	for _, tc := range []struct {
		name        string
		pod         *core.Pod
		ctrlFetcher controllerfetcher.ControllerFetcher
		expected    *controllerfetcher.ControllerKeyWithAPIVersion
	}{
		{
			name: "should return nil for Pod without ownerReferences",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					OwnerReferences: nil,
				},
			},
			ctrlFetcher: &controllerfetcher.NilControllerFetcher{},
			expected:    nil,
		},
		{
			name: "should return nil for Pod with ownerReference with controller=nil",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					OwnerReferences: []meta.OwnerReference{
						{
							APIVersion: "apps/v1",
							Controller: nil,
							Kind:       "ReplicaSet",
							Name:       "foo",
						},
					},
				},
			},
			ctrlFetcher: &controllerfetcher.FakeControllerFetcher{},
			expected:    nil,
		},
		{
			name: "should return nil for Pod with ownerReference with controller=false",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					OwnerReferences: []meta.OwnerReference{
						{
							APIVersion: "apps/v1",
							Controller: ptr.To(false),
							Kind:       "ReplicaSet",
							Name:       "foo",
						},
					},
				},
			},
			ctrlFetcher: &controllerfetcher.FakeControllerFetcher{},
			expected:    nil,
		},
		{
			name: "should pass the Pod ownerReference to the fake ControllerFetcher",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Namespace: "bar",
					OwnerReferences: []meta.OwnerReference{
						{
							APIVersion: "apps/v1",
							Controller: ptr.To(true),
							Kind:       "ReplicaSet",
							Name:       "foo",
						},
					},
				},
			},
			ctrlFetcher: &controllerfetcher.FakeControllerFetcher{},
			expected: &controllerfetcher.ControllerKeyWithAPIVersion{
				ControllerKey: controllerfetcher.ControllerKey{
					Namespace: "bar",
					Kind:      "ReplicaSet",
					Name:      "foo",
				},
				ApiVersion: "apps/v1",
			},
		},
		{
			name: "should not return an error for Node owner reference",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Namespace: "bar",
					OwnerReferences: []meta.OwnerReference{
						{
							APIVersion: "v1",
							Controller: ptr.To(true),
							Kind:       "Node",
							Name:       "foo",
						},
					},
				},
			},
			ctrlFetcher: &controllerfetcher.FakeControllerFetcher{},
			expected:    nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FindParentControllerForPod(context.Background(), tc.pod, tc.ctrlFetcher)
			assert.NoError(t, err, "Unexpected error occurred.")
			assert.Equal(t, got, tc.expected)
		})
	}
}

func TestIsPodReadyAndStartupBoostDurationPassed(t *testing.T) {
	now := meta.Now()
	past := meta.Time{Time: now.Add(-2 * time.Minute)}
	duration60 := int32(60)
	duration180 := int32(180)
	duration300 := int32(300)
	testCases := []struct {
		name     string
		pod      *core.Pod
		vpa      *vpa_types.VerticalPodAutoscaler
		expected bool
	}{
		{
			name:     "No StartupBoost config",
			pod:      &core.Pod{},
			vpa:      &vpa_types.VerticalPodAutoscaler{},
			expected: false,
		},
		{
			name:     "No duration in StartupBoost, no annotation",
			pod:      &core.Pod{},
			vpa:      test.VerticalPodAutoscaler().WithContainer(containerName).WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, 0).Get(),
			expected: false,
		},
		{
			name: "No duration in StartupBoost, with annotation",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
			},
			vpa:      test.VerticalPodAutoscaler().WithContainer(containerName).WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, 0).Get(),
			expected: true,
		},
		{
			name: "Pod not ready",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Status: core.PodStatus{
					Conditions: []core.PodCondition{
						{
							Type:   core.PodReady,
							Status: core.ConditionFalse,
						},
					},
				},
			},
			vpa:      test.VerticalPodAutoscaler().WithContainer(containerName).WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, 60).Get(),
			expected: false,
		},
		{
			name: "Duration passed",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Status: core.PodStatus{
					Conditions: []core.PodCondition{
						{
							Type:               core.PodReady,
							Status:             core.ConditionTrue,
							LastTransitionTime: past,
						},
					},
				},
			},
			vpa:      test.VerticalPodAutoscaler().WithContainer(containerName).WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, 60).Get(),
			expected: true,
		},
		{
			name: "Duration not passed",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Status: core.PodStatus{
					Conditions: []core.PodCondition{
						{
							Type:               core.PodReady,
							Status:             core.ConditionTrue,
							LastTransitionTime: now,
						},
					},
				},
			},
			vpa:      test.VerticalPodAutoscaler().WithContainer(containerName).WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, 60).Get(),
			expected: false,
		},
		{
			name: "Container-level boost duration",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{Name: "c1"},
						{Name: "c2"},
					},
				},
				Status: core.PodStatus{
					Conditions: []core.PodCondition{
						{
							Type:               core.PodReady,
							Status:             core.ConditionTrue,
							LastTransitionTime: past,
						},
					},
				},
			},
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "c1",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										DurationSeconds: &duration180,
									},
								},
							},
							{
								ContainerName: "c2",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										DurationSeconds: &duration60,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "Container-level boost duration passed",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{Name: "c1"},
						{Name: "c2"},
					},
				},
				Status: core.PodStatus{
					Conditions: []core.PodCondition{
						{
							Type:               core.PodReady,
							Status:             core.ConditionTrue,
							LastTransitionTime: meta.Time{Time: now.Add(-4 * time.Minute)},
						},
					},
				},
			},
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "c1",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										DurationSeconds: &duration180,
									},
								},
							},
							{
								ContainerName: "c2",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										DurationSeconds: &duration60,
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "Pod-level boost duration is higher",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{Name: "c1"},
						{Name: "c2"},
					},
				},
				Status: core.PodStatus{
					Conditions: []core.PodCondition{
						{
							Type:               core.PodReady,
							Status:             core.ConditionTrue,
							LastTransitionTime: meta.Time{Time: now.Add(-4 * time.Minute)},
						},
					},
				},
			},
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							DurationSeconds: &duration300,
						},
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "c1",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										DurationSeconds: &duration180,
									},
								},
							},
							{
								ContainerName: "c2",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										DurationSeconds: &duration60,
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsPodReadyAndStartupBoostDurationPassed(tc.pod, tc.vpa))
		})
	}
}

func TestPodHasCPUBoostInProgressAnnotation(t *testing.T) {
	testCases := []struct {
		name     string
		pod      *core.Pod
		expected bool
	}{
		{
			name:     "No annotations",
			pod:      &core.Pod{},
			expected: false,
		},
		{
			name: "Annotation present",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
			},
			expected: true,
		},
		{
			name: "Annotation not present",
			pod: &core.Pod{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{
						"another-annotation": "true",
					},
				},
			},
			expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, PodHasCPUBoostInProgressAnnotation(tc.pod))
		})
	}
}
