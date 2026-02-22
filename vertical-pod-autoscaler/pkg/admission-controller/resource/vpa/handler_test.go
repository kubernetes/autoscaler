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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/version"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
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
	validPodScalingMode := vpa_types.PodScalingModeAuto
	scalingModeOff := vpa_types.ContainerScalingModeOff
	controlledValuesRequestsAndLimits := vpa_types.ContainerControlledValuesRequestsAndLimits
	inPlaceOrRecreateUpdateMode := vpa_types.UpdateModeInPlaceOrRecreate
	containerScalingModeRecsOnlyMode := vpa_types.ContainerScalingModeRecsOnly
	tests := []struct {
		name                                 string
		vpa                                  vpa_types.VerticalPodAutoscaler
		isCreate                             bool
		expectError                          error
		inPlaceOrRecreateFeatureGateDisabled bool
		PerVPAConfigDisabled                 bool
		ContainerScalingModeRecsOnlyDisabled bool
		emulatedVersion                      string
	}{
		{
			name: "empty update",
			vpa:  vpa_types.VerticalPodAutoscaler{},
		},
		{
			name:        "empty create",
			vpa:         vpa_types.VerticalPodAutoscaler{},
			isCreate:    true,
			expectError: errors.New("targetRef is required. If you're using v1beta1 version of the API, please migrate to v1"),
		},
		{
			name: "no update mode",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{},
				},
			},
			expectError: errors.New("updateMode is required if UpdatePolicy is used"),
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
			expectError: errors.New("unexpected UpdateMode value bad"),
		},
		{
			name: "creating VPA with InPlaceOrRecreate update mode not allowed by disabled feature gate",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &inPlaceOrRecreateUpdateMode,
					},
				},
			},
			isCreate:                             true,
			inPlaceOrRecreateFeatureGateDisabled: true,
			expectError:                          fmt.Errorf("in order to use UpdateMode %s, you must enable feature gate %s in the admission-controller args", vpa_types.UpdateModeInPlaceOrRecreate, features.InPlaceOrRecreate),
		},
		{
			name: "updating VPA with InPlaceOrRecreate update mode allowed by disabled feature gate",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &inPlaceOrRecreateUpdateMode,
					},
				},
			},
			isCreate:                             false,
			inPlaceOrRecreateFeatureGateDisabled: true,
			expectError:                          nil,
		},
		{
			name: "InPlaceOrRecreate update mode enabled by feature gate",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &inPlaceOrRecreateUpdateMode,
					},
				},
			},
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
			expectError: errors.New("minReplicas has to be positive, got 0"),
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
			expectError: errors.New("containerPolicies.ContainerName is required"),
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
			expectError: errors.New("unexpected Mode value bad"),
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
			expectError: errors.New("the current version of VPA object shouldn't specify more than one recommenders"),
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
			expectError: errors.New("max resource for cpu is lower than min"),
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
			expectError: fmt.Errorf("minAllowed: CPU [%v] must be a whole number of milli CPUs", badCPUResource.String()),
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
			expectError: fmt.Errorf("minAllowed: memory [%v] must be a whole number of bytes", resource.MustParse("100m")),
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
			expectError: fmt.Errorf("maxAllowed: CPU [%s] must be a whole number of milli CPUs", badCPUResource.String()),
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
			expectError: fmt.Errorf("maxAllowed: memory [%v] must be a whole number of bytes", resource.MustParse("500m")),
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
			expectError: errors.New("controlledValues shouldn't be specified if container scaling mode is off"),
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
		{
			name: "per-vpa config active and used",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
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
								OOMBumpUpRatio: resource.NewQuantity(2, resource.DecimalSI),
							},
						},
					},
				},
			},
			PerVPAConfigDisabled: false,
		},
		{
			name: "per-vpa config disabled and used",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
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
								OOMMinBumpUp: resource.NewQuantity(2, resource.DecimalSI),
							},
						},
					},
				},
			},
			PerVPAConfigDisabled: true,
			expectError:          errors.New("OOMBumpUpRatio and OOMMinBumpUp are not supported when feature flag PerVPAConfig is disabled"),
		},
		{
			name: "support for pod level stanzas disabled and used",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "c1",
								Mode:          &containerScalingModeRecsOnlyMode,
							},
						},
					},
				},
			},
			ContainerScalingModeRecsOnlyDisabled: true,
			expectError:                          errors.New("in order to use RecommendationOnly containerPolicies mode, you must enable feature gate PodLevelResourcesSupportForVPA in the admission-controller args"),
		},
		{
			name: "support for pod level active and used",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						PodPolicies: &vpa_types.PodResourcePolicies{
							Mode:             &validPodScalingMode,
							ControlledValues: &controlledValuesRequestsAndLimits,
						},
					},
				},
			},
			ContainerScalingModeRecsOnlyDisabled: false,
			emulatedVersion:                      "1.6",
		},
		{
			name: "support for pod level active and used should fail",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						PodPolicies: &vpa_types.PodResourcePolicies{
							Mode: &validPodScalingMode,
						},
					},
				},
			},
			ContainerScalingModeRecsOnlyDisabled: true,
			emulatedVersion:                      "1.6",
			expectError:                          errors.New("in order to use podPolicies stanza, you must enable feature gate PodLevelResourcesSupportForVPA in the admission-controller args"),
		},
		{
			name: "support for pod level and per-vpa config active and used",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						PodPolicies: &vpa_types.PodResourcePolicies{
							Mode:           &validPodScalingMode,
							OOMBumpUpRatio: resource.NewQuantity(2, resource.DecimalSI),
						},
					},
				},
			},
			PerVPAConfigDisabled:                 false,
			ContainerScalingModeRecsOnlyDisabled: false,
			emulatedVersion:                      "1.6",
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			if tc.inPlaceOrRecreateFeatureGateDisabled {
				featuregatetesting.SetFeatureGateEmulationVersionDuringTest(t, features.MutableFeatureGate, version.MustParse("1.5"))
				featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, !tc.inPlaceOrRecreateFeatureGateDisabled)
			}
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.PerVPAConfig, !tc.PerVPAConfigDisabled)

			if tc.emulatedVersion != "" && tc.emulatedVersion == "1.6" {
				featuregatetesting.SetFeatureGateEmulationVersionDuringTest(t, features.MutableFeatureGate, version.MustParse(tc.emulatedVersion))
				featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.PodLevelResourcesSupportForVPA, !tc.ContainerScalingModeRecsOnlyDisabled)
			}

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
