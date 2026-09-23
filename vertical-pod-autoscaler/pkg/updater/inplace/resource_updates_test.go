/*
Copyright The Kubernetes Authors.

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

package inplace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

type fakeRecommendationProvider struct {
	resources []vpa_api_util.ContainerResources
	err       error
}

func (frp *fakeRecommendationProvider) GetContainersResourcesForPod(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]vpa_api_util.ContainerResources, vpa_api_util.ContainerToAnnotationsMap, error) {
	return frp.resources, nil, frp.err
}

func TestCalculatePatches_MultiContainerResourceUpdates(t *testing.T) {
	now := metav1.Now()
	past := metav1.Time{Time: now.Add(-5 * time.Minute)}
	ten := int32(10)
	thousand := int32(1000)

	tests := []struct {
		name               string
		pod                *corev1.Pod
		vpa                *vpa_types.VerticalPodAutoscaler
		recommendResources []vpa_api_util.ContainerResources
		recommendError     error
		expectPatches      []resource_admission.PatchRecord
		expectError        error
	}{
		{
			name: "UpdateMode Off - Only unboost expired container",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotations.GetStartupCPUBoostAnnotationKey("c1"): "{\"requests\":{\"cpu\":\"100m\"}}",
						annotations.GetStartupCPUBoostAnnotationKey("c2"): "{\"requests\":{\"cpu\":\"100m\"}}",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
						{Name: "c2", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodReady, Status: corev1.ConditionTrue, LastTransitionTime: past},
					},
				},
			},
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{UpdateMode: &[]vpa_types.UpdateMode{vpa_types.UpdateModeOff}[0]},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{ContainerName: "c1", StartupBoost: &vpa_types.StartupBoost{CPU: &vpa_types.GenericStartupBoost{DurationSeconds: &ten}}},
							{ContainerName: "c2", StartupBoost: &vpa_types.StartupBoost{CPU: &vpa_types.GenericStartupBoost{DurationSeconds: &thousand}}},
						},
					},
				},
			},
			recommendResources: make([]vpa_api_util.ContainerResources, 2),
			expectPatches: []resource_admission.PatchRecord{
				{Op: "add", Path: "/spec/containers/0/resources/requests/cpu", Value: "100m"},
			},
		},
		{
			name: "UpdateMode Auto - ignores boosted unexpired containers naturally",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotations.GetStartupCPUBoostAnnotationKey("c1"): "{\"requests\":{\"cpu\":\"100m\"}}",
						annotations.GetStartupCPUBoostAnnotationKey("c2"): "{\"requests\":{\"cpu\":\"100m\"}}",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
						{Name: "c2", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodReady, Status: corev1.ConditionTrue, LastTransitionTime: past},
					},
				},
			},
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{UpdateMode: &[]vpa_types.UpdateMode{vpa_types.UpdateModeInPlaceOrRecreate}[0]},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{ContainerName: "c1", StartupBoost: &vpa_types.StartupBoost{CPU: &vpa_types.GenericStartupBoost{DurationSeconds: &ten}}},
							{ContainerName: "c2", StartupBoost: &vpa_types.StartupBoost{CPU: &vpa_types.GenericStartupBoost{DurationSeconds: &thousand}}},
						},
					},
				},
			},
			recommendResources: []vpa_api_util.ContainerResources{
				{Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")}},
				{Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")}},
			},
			expectPatches: []resource_admission.PatchRecord{
				{Op: "add", Path: "/spec/containers/0/resources/requests/cpu", Value: "200m"},
			},
		},
		{
			name: "No Unboost - both containers unexpired so they stay boosted safely",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotations.GetStartupCPUBoostAnnotationKey("c1"): "{\"requests\":{\"cpu\":\"100m\"}}",
						annotations.GetStartupCPUBoostAnnotationKey("c2"): "{\"requests\":{\"cpu\":\"100m\"}}",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
						{Name: "c2", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodReady, Status: corev1.ConditionTrue, LastTransitionTime: past},
					},
				},
			},
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{UpdateMode: &[]vpa_types.UpdateMode{vpa_types.UpdateModeInPlaceOrRecreate}[0]},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{ContainerName: "c1", StartupBoost: &vpa_types.StartupBoost{CPU: &vpa_types.GenericStartupBoost{DurationSeconds: &thousand}}},
							{ContainerName: "c2", StartupBoost: &vpa_types.StartupBoost{CPU: &vpa_types.GenericStartupBoost{DurationSeconds: &thousand}}},
						},
					},
				},
			},
			recommendResources: []vpa_api_util.ContainerResources{
				{Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")}},
				{Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")}},
			},
			expectPatches: []resource_admission.PatchRecord{},
		},
		{
			name: "Regular Recommendation - no startup boost annotations or config",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
						{Name: "c2", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
					},
				},
			},
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{UpdateMode: &[]vpa_types.UpdateMode{vpa_types.UpdateModeInPlaceOrRecreate}[0]},
				},
			},
			recommendResources: []vpa_api_util.ContainerResources{
				{Requests: corev1.ResourceList{"cpu": resource.MustParse("300m")}},
				{Requests: corev1.ResourceList{"cpu": resource.MustParse("300m")}},
			},
			expectPatches: []resource_admission.PatchRecord{
				{Op: "add", Path: "/spec/containers/0/resources/requests/cpu", Value: "300m"},
				{Op: "add", Path: "/spec/containers/1/resources/requests/cpu", Value: "300m"},
			},
		},
		// Pod with UpdateMode Off and no startup boost is not going to end up here due to check in updater.go
		// but this extra validation makes sure it yields no patches
		{
			name: "Regular Recommendation - UpdateMode Off - no startup boost annotations yields no patches",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "c1", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
						{Name: "c2", Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{"cpu": resource.MustParse("1m")}}},
					},
				},
			},
			vpa: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{UpdateMode: &[]vpa_types.UpdateMode{vpa_types.UpdateModeOff}[0]},
				},
			},
			recommendResources: []vpa_api_util.ContainerResources{
				{Requests: corev1.ResourceList{"cpu": resource.MustParse("300m")}},
				{Requests: corev1.ResourceList{"cpu": resource.MustParse("300m")}},
			},
			expectPatches: []resource_admission.PatchRecord{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			frp := fakeRecommendationProvider{resources: tc.recommendResources, err: tc.recommendError}
			calculator := resourcesInplaceUpdatesPatchCalculator{recommendationProvider: &frp}

			patches, err := calculator.CalculatePatches(tc.pod, tc.vpa)
			if tc.expectError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectPatches, patches)
		})
	}
}
