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
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

func TestUpdateVpaIfNeeded(t *testing.T) {
	updatedVpa := test.VerticalPodAutoscaler().WithName("vpa").WithNamespace("test").WithContainer(containerName).
		AppendCondition(vpa_types.RecommendationProvided, corev1.ConditionTrue, "reason", "msg", anytime).Get()
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
				AppendCondition(vpa_types.RecommendationProvided, corev1.ConditionTrue, "reason", "msg", anytime).Get(),
			expectedUpdate: false,
		}, {
			caseName:   "Updates on recommendation change.",
			updatedVpa: updatedVpa,
			observedVpa: observedVpaBuilder.WithTarget("10", "200").
				AppendCondition(vpa_types.RecommendationProvided, corev1.ConditionTrue, "reason", "msg", anytime).Get(),
			expectedUpdate: true,
		}, {
			caseName:   "Updates on condition change.",
			updatedVpa: updatedVpa,
			observedVpa: observedVpaBuilder.WithTarget("5", "200").
				AppendCondition(vpa_types.RecommendationProvided, corev1.ConditionFalse, "reason", "msg", anytime).Get(),
			expectedUpdate: true,
		}, {
			caseName:   "Updates on condition added.",
			updatedVpa: updatedVpa,
			observedVpa: observedVpaBuilder.WithTarget("5", "200").
				AppendCondition(vpa_types.RecommendationProvided, corev1.ConditionTrue, "reason", "msg", anytime).
				AppendCondition(vpa_types.LowConfidence, corev1.ConditionTrue, "reason", "msg", anytime).Get(),
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
		pod             *corev1.Pod
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
	pod.OwnerReferences = []metav1.OwnerReference{
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
	vpaA.Spec.TargetRef = &autoscalingv1.CrossVersionObjectReference{
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
		MinAllowed: corev1.ResourceList{
			corev1.ResourceCPU: *resource.NewScaledQuantity(10, 1),
		},
	}
	containerPolicy2 := vpa_types.ContainerResourcePolicy{
		ContainerName: "container2",
		MaxAllowed: corev1.ResourceList{
			corev1.ResourceMemory: *resource.NewScaledQuantity(100, 1),
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
		MinAllowed: corev1.ResourceList{
			corev1.ResourceCPU: *resource.NewScaledQuantity(20, 1),
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
		pod         *corev1.Pod
		ctrlFetcher controllerfetcher.ControllerFetcher
		expected    *controllerfetcher.ControllerKeyWithAPIVersion
	}{
		{
			name: "should return nil for Pod without ownerReferences",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: nil,
				},
			},
			ctrlFetcher: &controllerfetcher.NilControllerFetcher{},
			expected:    nil,
		},
		{
			name: "should return nil for Pod with ownerReference with controller=nil",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
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
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
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
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "bar",
					OwnerReferences: []metav1.OwnerReference{
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
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "bar",
					OwnerReferences: []metav1.OwnerReference{
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
	now := metav1.Now()
	past := metav1.Time{Time: now.Add(-2 * time.Minute)}
	duration60 := int32(60)
	duration180 := int32(180)
	duration300 := int32(300)
	testCases := []struct {
		name     string
		pod      *corev1.Pod
		vpa      *vpa_types.VerticalPodAutoscaler
		expected bool
	}{
		{
			name:     "No StartupBoost config",
			pod:      &corev1.Pod{},
			vpa:      &vpa_types.VerticalPodAutoscaler{},
			expected: false,
		},
		{
			name:     "No duration in StartupBoost, no annotation",
			pod:      &corev1.Pod{},
			vpa:      test.VerticalPodAutoscaler().WithContainer(containerName).WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, 0).Get(),
			expected: false,
		},
		{
			name: "No duration in StartupBoost, with annotation and pod ready",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			vpa:      test.VerticalPodAutoscaler().WithContainer(containerName).WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, 0).Get(),
			expected: true,
		},
		{
			name: "Pod not ready",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			vpa:      test.VerticalPodAutoscaler().WithContainer(containerName).WithCPUStartupBoost(vpa_types.FactorStartupBoostType, nil, nil, 60).Get(),
			expected: false,
		},
		{
			name: "Duration passed",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:               corev1.PodReady,
							Status:             corev1.ConditionTrue,
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
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:               corev1.PodReady,
							Status:             corev1.ConditionTrue,
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
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1"},
						{Name: "c2"},
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:               corev1.PodReady,
							Status:             corev1.ConditionTrue,
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
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1"},
						{Name: "c2"},
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:               corev1.PodReady,
							Status:             corev1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: now.Add(-4 * time.Minute)},
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
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1"},
						{Name: "c2"},
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:               corev1.PodReady,
							Status:             corev1.ConditionTrue,
							LastTransitionTime: metav1.Time{Time: now.Add(-4 * time.Minute)},
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
		pod      *corev1.Pod
		expected bool
	}{
		{
			name:     "No annotations",
			pod:      &corev1.Pod{},
			expected: false,
		},
		{
			name: "Annotation present",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"startup-cpu-boost": "",
					},
				},
			},
			expected: true,
		},
		{
			name: "Annotation not present",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
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

func TestDetermineManagedContainers(t *testing.T) {
	podScalingModeAuto := vpa_types.PodScalingModeAuto
	containerScalingModeAuto := vpa_types.ContainerScalingModeAuto
	containerScalingModeRecsOnly := vpa_types.ContainerScalingModeRecsOnly
	status := vpa_types.VerticalPodAutoscalerStatus{
		Recommendation: &vpa_types.RecommendedPodResources{
			ContainerRecommendations: []vpa_types.RecommendedContainerResources{
				test.Recommendation().WithContainer("c1").GetContainerResources(),
				test.Recommendation().WithContainer("c2").GetContainerResources(),
				test.Recommendation().WithContainer("c3").GetContainerResources(),
			},
		},
	}
	tests := []struct {
		name               string
		vpa                *vpa_types.VerticalPodAutoscaler
		expectedContainers []string
	}{
		{
			name: "only pod level scaling is set", // i.e. by default all containers mode is auto
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						PodPolicies: &vpa_types.PodResourcePolicies{
							Mode: &podScalingModeAuto,
						},
					},
				},
				Status: status,
			},
			expectedContainers: []string{"c1", "c2", "c3"},
		},
		{
			name: "pod level scaling is set and one container is in auto mode",
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "c1",
								Mode:          &containerScalingModeAuto,
							},
							{
								ContainerName: "*",
								Mode:          &containerScalingModeRecsOnly,
							},
						},
						PodPolicies: &vpa_types.PodResourcePolicies{
							Mode: &podScalingModeAuto,
						},
					},
				},
				Status: status,
			},
			expectedContainers: []string{"c1"},
		},
		{
			name: "pod level scaling is set and all containers are in recommendation only mode",
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "*",
								Mode:          &containerScalingModeRecsOnly,
							},
						},
						PodPolicies: &vpa_types.PodResourcePolicies{
							Mode: &podScalingModeAuto,
						},
					},
				},
				Status: status,
			},
			expectedContainers: []string{},
		},
		{
			name: "pod level scaling is set and all containers are in auto mode",
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "*",
								Mode:          &containerScalingModeAuto,
							},
						},
						PodPolicies: &vpa_types.PodResourcePolicies{
							Mode: &podScalingModeAuto,
						},
					},
				},
				Status: status,
			},
			expectedContainers: []string{"c1", "c2", "c3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := determineManagedContainers(tt.vpa)
			assert.Equal(t, tt.expectedContainers, actual, "doesn't match")
		})
	}
}

func TestGetPodControlledValues(t *testing.T) {
	requestsAndLimits := vpa_types.ContainerControlledValuesRequestsAndLimits
	requestsOnly := vpa_types.ContainerControlledValuesRequestsOnly
	for _, tc := range []struct {
		name     string
		policy   *vpa_types.PodResourcePolicy
		expected vpa_types.ContainerControlledValues
	}{
		{
			name:     "default pod level policy is RequestAndLimits",
			policy:   nil,
			expected: vpa_types.ContainerControlledValuesRequestsAndLimits,
		},
		{
			name: "pod level policy should be RequestsAndLimits",
			policy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					ControlledValues: &requestsAndLimits,
				},
			},
			expected: vpa_types.ContainerControlledValuesRequestsAndLimits,
		},
		{
			name: "pod level policy should be RequestsOnly",
			policy: &vpa_types.PodResourcePolicy{
				PodPolicies: &vpa_types.PodResourcePolicies{
					ControlledValues: &requestsOnly,
				},
			},
			expected: vpa_types.ContainerControlledValuesRequestsOnly,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := GetPodControlledValues(tc.policy)
			assert.Equal(t, got, tc.expected)
		})
	}
}

func TestFilterContainerRecommendations(t *testing.T) {
	tests := []struct {
		name                  string
		vpa                   *vpa_types.VerticalPodAutoscaler
		expectedContainerRecs []vpa_types.RecommendedContainerResources
	}{
		{
			name:                  "nil VPA object",
			vpa:                   nil,
			expectedContainerRecs: nil,
		},
		{
			name: "Empty VPA status stanza",
			vpa: test.VerticalPodAutoscaler().
				Get(),
			expectedContainerRecs: nil,
		},
		{
			name: "Container level policies are set and empty VPA status stanza",
			vpa: test.VerticalPodAutoscaler().
				WithContainer("container1").
				WithScalingMode("container1", vpa_types.ContainerScalingModeRecsOnly).
				WithContainer("container2").
				WithScalingMode("container2", vpa_types.ContainerScalingModeRecsOnly).
				Get(),
			expectedContainerRecs: nil,
		},
		{
			name: "All containers use Auto mode and container level recommendations are present",
			vpa: test.VerticalPodAutoscaler().
				WithContainer("container1").
				AppendRecommendation(
					test.Recommendation().
						WithContainer("container1").
						WithTarget("1", "1").
						GetContainerResources()).
				WithContainer("container2").
				AppendRecommendation(
					test.Recommendation().
						WithContainer("container2").
						WithTarget("2", "2").
						GetContainerResources()).
				Get(),
			expectedContainerRecs: []vpa_types.RecommendedContainerResources{
				test.Recommendation().WithContainer("container1").WithTarget("1", "1").GetContainerResources(),
				test.Recommendation().WithContainer("container2").WithTarget("2", "2").GetContainerResources(),
			},
		},
		{
			name: "One container uses RecommendationOnly mode the other uses Auto mode and container level recommendations are present",
			vpa: test.VerticalPodAutoscaler().
				WithContainer("container1").
				WithScalingMode("container1", vpa_types.ContainerScalingModeRecsOnly).
				AppendRecommendation(
					test.Recommendation().
						WithContainer("container1").
						WithTarget("1", "1").
						GetContainerResources()).
				WithContainer("container2").
				AppendRecommendation(
					test.Recommendation().
						WithContainer("container2").
						WithTarget("2", "2").
						GetContainerResources()).
				Get(),
			expectedContainerRecs: []vpa_types.RecommendedContainerResources{
				test.Recommendation().WithContainer("container2").WithTarget("2", "2").GetContainerResources(),
			},
		},
		{
			name: "All containers use RecommendationOnly mode and container level recommendations are present",
			vpa: test.VerticalPodAutoscaler().
				WithContainer("container1").
				WithScalingMode("container1", vpa_types.ContainerScalingModeRecsOnly).
				AppendRecommendation(
					test.Recommendation().
						WithContainer("container1").
						WithTarget("1", "1").
						GetContainerResources()).
				WithContainer("container2").
				WithScalingMode("container2", vpa_types.ContainerScalingModeRecsOnly).
				AppendRecommendation(
					test.Recommendation().
						WithContainer("container2").
						WithTarget("2", "2").
						GetContainerResources()).
				WithPodLevelTarget("10", "10").
				Get(),
			expectedContainerRecs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FilterContainerRecommendations(tt.vpa)
			assert.Equal(t, tt.expectedContainerRecs, actual, "doesn't match")
		})
	}
}

func TestIsPodLevelScalingModeEnabled(t *testing.T) {
	tests := []struct {
		name     string
		vpa      *vpa_types.VerticalPodAutoscaler
		expected bool
	}{
		{
			name:     "nil VPA object",
			vpa:      nil,
			expected: false,
		},
		{
			name: "VPA object with nil PodPolicies",
			vpa: test.VerticalPodAutoscaler().
				Get(),
			expected: false,
		},
		{
			name: "Pod level scaling mode is set to Off",
			vpa: test.VerticalPodAutoscaler().
				WithPodLevelScalingMode(vpa_types.PodScalingModeOff).
				Get(),
			expected: false,
		},
		{
			name: "Pod level scaling mode is set to Auto",
			vpa: test.VerticalPodAutoscaler().
				WithPodLevelScalingMode(vpa_types.PodScalingModeAuto).
				Get(),
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsPodLevelScalingModeEnabled(tt.vpa)
			assert.Equal(t, tt.expected, actual, "doesn't match")
		})
	}
}
