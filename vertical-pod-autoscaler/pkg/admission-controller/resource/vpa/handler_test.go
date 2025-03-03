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

package vpa

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

const (
	cpu    = apiv1.ResourceCPU
	memory = apiv1.ResourceMemory
)

func TestValidateVPA(t *testing.T) {
	badUpdateMode := vpa_types.UpdateMode("bad")
	validUpdateMode := vpa_types.UpdateModeOff
	badMinReplicas := int32(0)
	validMinReplicas := int32(1)
	badScalingMode := vpa_types.ContainerScalingMode("bad")
	badCPUResource := resource.MustParse("187500u")
	validScalingMode := vpa_types.ContainerScalingModeAuto
	scalingModeOff := vpa_types.ContainerScalingModeOff
	controlledValuesRequestsAndLimits := vpa_types.ContainerControlledValuesRequestsAndLimits
	tests := []struct {
		name        string
		vpa         vpa_types.VerticalPodAutoscaler
		isCreate    bool
		expectError error
	}{
		{
			name: "empty update",
			vpa:  vpa_types.VerticalPodAutoscaler{},
		},
		{
			name:        "empty create",
			vpa:         vpa_types.VerticalPodAutoscaler{},
			isCreate:    true,
			expectError: fmt.Errorf("TargetRef is required. If you're using v1beta1 version of the API, please migrate to v1"),
		},
		{
			name: "no update mode",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{},
				},
			},
			expectError: fmt.Errorf("UpdateMode is required if UpdatePolicy is used"),
		},
		{
			name: "bad update mode",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &badUpdateMode,
					},
				},
			},
			expectError: fmt.Errorf("unexpected UpdateMode value bad"),
		},
		{
			name: "zero minReplicas",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						MinReplicas: &badMinReplicas,
						UpdateMode:  &validUpdateMode,
					},
				},
			},
			expectError: fmt.Errorf("MinReplicas has to be positive, got 0"),
		},
		{
			name: "no policy name",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{{}},
					},
				},
			},
			expectError: fmt.Errorf("ContainerPolicies.ContainerName is required"),
		},
		{
			name: "invalid scaling mode",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &badScalingMode,
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("unexpected Mode value bad"),
		},
		{
			name: "more than one recommender",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
						{Name: "test1"},
						{Name: "test2"},
					},
				},
			},
			expectError: fmt.Errorf("The current version of VPA object shouldn't specify more than one recommenders."),
		},
		{
			name: "bad limits",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed: apiv1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								MaxAllowed: apiv1.ResourceList{
									cpu: resource.MustParse("10"),
								},
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("max resource for cpu is lower than min"),
		},
		{
			name: "bad minAllowed cpu value",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed: apiv1.ResourceList{
									cpu: resource.MustParse("187500u"),
								},
								MaxAllowed: apiv1.ResourceList{
									cpu: resource.MustParse("275m"),
								},
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("MinAllowed: CPU [%v] must be a whole number of milli CPUs", badCPUResource.String()),
		},
		{
			name: "bad minAllowed memory value",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed: apiv1.ResourceList{
									cpu:    resource.MustParse("1m"),
									memory: resource.MustParse("100m"),
								},
								MaxAllowed: apiv1.ResourceList{
									cpu:    resource.MustParse("275m"),
									memory: resource.MustParse("500M"),
								},
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("MinAllowed: Memory [%v] must be a whole number of bytes", resource.MustParse("100m")),
		},
		{
			name: "bad maxAllowed cpu value",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed:    apiv1.ResourceList{},
								MaxAllowed: apiv1.ResourceList{
									cpu: resource.MustParse("187500u"),
								},
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("MaxAllowed: CPU [%s] must be a whole number of milli CPUs", badCPUResource.String()),
		},
		{
			name: "bad maxAllowed memory value",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed: apiv1.ResourceList{
									cpu: resource.MustParse("1m")},
								MaxAllowed: apiv1.ResourceList{
									cpu:    resource.MustParse("275m"),
									memory: resource.MustParse("500m"),
								},
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("MaxAllowed: Memory [%v] must be a whole number of bytes", resource.MustParse("500m")),
		},
		{
			name: "scaling off with controlled values requests and limits",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:    "loot box",
								Mode:             &scalingModeOff,
								ControlledValues: &controlledValuesRequestsAndLimits,
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("ControlledValues shouldn't be specified if container scaling mode is off."),
		},
		{
			name: "all valid",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: apiv1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: apiv1.ResourceList{
									cpu: resource.MustParse("100"),
								},
							},
						},
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode:  &validUpdateMode,
						MinReplicas: &validMinReplicas,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			err := ValidateVPA(&tc.vpa, tc.isCreate)
			if tc.expectError == nil {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					assert.Equal(t, tc.expectError.Error(), err.Error())
				}
			}
		})
	}
}
