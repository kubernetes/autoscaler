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
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	scalingModeOff := vpa_types.ContainerScalingModeOff
	controlledValuesRequestsAndLimits := vpa_types.ContainerControlledValuesRequestsAndLimits
	inPlaceOrRecreateUpdateMode := vpa_types.UpdateModeInPlaceOrRecreate
	badCPUBoostFactor := int32(0)
	validCPUBoostFactor := int32(2)
	badCPUBoostQuantity := resource.MustParse("187500u")
	validCPUBoostQuantity := resource.MustParse("100m")
	badCPUBoostType := vpa_types.StartupBoostType("bad")
	validCPUBoostTypeFactor := vpa_types.FactorStartupBoostType
	validCPUBoostTypeQuantity := vpa_types.QuantityStartupBoostType

	tests := []struct {
		name                                 string
		vpa                                  vpa_types.VerticalPodAutoscaler
		isCreate                             bool
		expectError                          error
		inPlaceOrRecreateFeatureGateDisabled bool
		PerVPAConfigDisabled                 bool
		cpuStartupBoostFeatureGateDisabled   bool
	}{
		{
			name: "empty update",
			vpa:  vpa_types.VerticalPodAutoscaler{},
		},
		{
			name:        "empty create",
			vpa:         vpa_types.VerticalPodAutoscaler{},
			isCreate:    true,
			expectError: fmt.Errorf("targetRef is required. If you're using v1beta1 version of the API, please migrate to v1"),
		},
		{
			name: "no update mode",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{},
				},
			},
			expectError: fmt.Errorf("updateMode is required if UpdatePolicy is used"),
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
			expectError: fmt.Errorf("minReplicas has to be positive, got 0"),
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
			expectError: fmt.Errorf("containerPolicies.ContainerName is required"),
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
			expectError: fmt.Errorf("the current version of VPA object shouldn't specify more than one recommenders"),
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
			expectError: fmt.Errorf("controlledValues shouldn't be specified if container scaling mode is off"),
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
			name: "top-level startupBoost with feature gate disabled",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Factor: &validCPUBoostFactor,
						},
					},
				},
			},
			isCreate:                           true,
			cpuStartupBoostFeatureGateDisabled: true,
			expectError:                        fmt.Errorf("invalid startupBoost: in order to use startupBoost, you must enable feature gate %s in the admission-controller args", features.CPUStartupBoost),
		},
		{
			name: "container startupBoost with feature gate disabled",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										Factor: &validCPUBoostFactor,
									},
								},
							},
						},
					},
				},
			},
			isCreate:                           true,
			cpuStartupBoostFeatureGateDisabled: true,
			expectError:                        fmt.Errorf("invalid startupBoost in container loot box: in order to use startupBoost, you must enable feature gate %s in the admission-controller args", features.CPUStartupBoost),
		},
		{
			name: "top-level startupBoost with bad factor",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Type:   vpa_types.FactorStartupBoostType,
							Factor: &badCPUBoostFactor,
						},
					},
				},
			},
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost: invalid StartupBoost.CPU.Factor: must be >= 1 for Type Factor"),
		},
		{
			name: "container startupBoost with bad factor",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										Type:   vpa_types.FactorStartupBoostType,
										Factor: &badCPUBoostFactor,
									},
								},
							},
						},
					},
				},
			},
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost in container loot box: invalid StartupBoost.CPU.Factor: must be >= 1 for Type Factor"),
		},
		{
			name: "top-level startupBoost with bad quantity",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Type:     validCPUBoostTypeQuantity,
							Quantity: &badCPUBoostQuantity,
						},
					},
				},
			},
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost: invalid StartupBoost.CPU.Quantity: CPU [%v] must be a whole number of milli CPUs", &badCPUBoostQuantity),
		},
		{
			name: "container startupBoost with bad quantity",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										Type:     validCPUBoostTypeQuantity,
										Quantity: &badCPUBoostQuantity,
									},
								},
							},
						},
					},
				},
			},
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost in container loot box: invalid StartupBoost.CPU.Quantity: CPU [%v] must be a whole number of milli CPUs", &badCPUBoostQuantity),
		},
		{
			name: "top-level startupBoost with bad type",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Type: badCPUBoostType,
						},
					},
				},
			},
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost: unexpected StartupBoost.CPU.Type value bad"),
		},
		{
			name: "container startupBoost with bad type",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										Type: badCPUBoostType,
									},
								},
							},
						},
					},
				},
			},
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost in container loot box: unexpected StartupBoost.CPU.Type value bad"),
		},
		{
			name: "top-level startupBoost with valid factor",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Type:   validCPUBoostTypeFactor,
							Factor: &validCPUBoostFactor,
						},
					},
				},
			},
			isCreate: true,
		},
		{
			name: "container startupBoost with valid factor",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										Type:   validCPUBoostTypeFactor,
										Factor: &validCPUBoostFactor,
									},
								},
							},
						},
					},
				},
			},
			isCreate: true,
		},
		{
			name: "top-level startupBoost with valid quantity",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Type:     validCPUBoostTypeQuantity,
							Quantity: &validCPUBoostQuantity,
						},
					},
				},
			},
			isCreate: true,
		},
		{
			name: "container startupBoost with valid quantity",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscaling.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										Type:     validCPUBoostTypeQuantity,
										Quantity: &validCPUBoostQuantity,
									},
								},
							},
						},
					},
				},
			},
			isCreate: true,
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
			expectError:          fmt.Errorf("OOMBumpUpRatio and OOMMinBumpUp are not supported when feature flag PerVPAConfig is disabled"),
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.PerVPAConfig, !tc.PerVPAConfigDisabled)
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, !tc.inPlaceOrRecreateFeatureGateDisabled)
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.CPUStartupBoost, !tc.cpuStartupBoostFeatureGateDisabled)
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
