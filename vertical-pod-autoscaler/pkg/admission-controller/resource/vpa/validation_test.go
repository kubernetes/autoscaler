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
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/version"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/utils/ptr"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
)

const (
	cpu    = corev1.ResourceCPU
	memory = corev1.ResourceMemory
)

func TestAllowCPUBoost(t *testing.T) {
	tests := []struct {
		name        string
		oldObj      *vpa_types.VerticalPodAutoscaler
		featureFlag bool
		expected    bool
	}{
		{
			name:        "feature enabled returns true",
			oldObj:      nil,
			featureFlag: true,
			expected:    true,
		},
		{
			name:        "feature disabled and oldObj nil returns false",
			oldObj:      nil,
			featureFlag: false,
			expected:    false,
		},
		{
			name: "feature disabled but top-level StartupBoost.CPU set returns true",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Factor: ptr.To(int32(2)),
						},
					},
				},
			},
			featureFlag: false,
			expected:    true,
		},
		{
			name: "feature disabled but container StartupBoost set returns true",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "app",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{
										Factor: ptr.To(int32(2)),
									},
								},
							},
						},
					},
				},
			},
			featureFlag: false,
			expected:    true,
		},
		{
			name: "feature disabled and no startup boost configured returns false",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "app",
							},
						},
					},
				},
			},
			featureFlag: false,
			expected:    false,
		},
		{
			name: "feature disabled with empty container policies returns false",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{},
					},
				},
			},
			featureFlag: false,
			expected:    false,
		},
		{
			name: "feature disabled with ResourcePolicy nil returns false",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					StartupBoost: &vpa_types.StartupBoost{},
				},
			},
			featureFlag: false,
			expected:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.CPUStartupBoost, tc.featureFlag)
			result := allowCPUBoost(tc.oldObj)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAllowPerVPAConfig(t *testing.T) {
	tests := []struct {
		name        string
		oldObj      *vpa_types.VerticalPodAutoscaler
		featureFlag bool
		expected    bool
	}{
		{
			name:        "feature enabled returns true",
			oldObj:      nil,
			featureFlag: true,
			expected:    true,
		},
		{
			name:        "feature disabled and oldObj nil returns false",
			oldObj:      nil,
			featureFlag: false,
			expected:    false,
		},
		{
			name: "feature disabled but EvictAfterOOMSeconds set returns true",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						EvictAfterOOMSeconds: ptr.To(int32(600)),
					},
				},
			},
			featureFlag: false,
			expected:    true,
		},
		{
			name: "feature disabled but OOMBumpUpRatio set returns true",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:  "app",
								OOMBumpUpRatio: resource.NewQuantity(2, resource.DecimalSI),
							},
						},
					},
				},
			},
			featureFlag: false,
			expected:    true,
		},
		{
			name: "feature disabled but OOMMinBumpUp set returns true",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "app",
								OOMMinBumpUp:  resource.NewQuantity(100, resource.BinarySI),
							},
						},
					},
				},
			},
			featureFlag: false,
			expected:    true,
		},
		{
			name: "feature disabled and no per-vpa config returns false",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "app",
							},
						},
					},
				},
			},
			featureFlag: false,
			expected:    false,
		},
		{
			name: "feature disabled with empty container policies returns false",
			oldObj: &vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{},
					},
				},
			},
			featureFlag: false,
			expected:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.PerVPAConfig, tc.featureFlag)
			result := allowPerVPAConfig(tc.oldObj)
			assert.Equal(t, tc.expected, result)
		})
	}
}

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
	inPlaceUpdateMode := vpa_types.UpdateModeInPlace
	badCPUBoostFactor := int32(0)
	validCPUBoostFactor := int32(2)
	badCPUBoostQuantity := resource.MustParse("187500u")
	validCPUBoostQuantity := resource.MustParse("100m")
	badCPUBoostType := vpa_types.StartupBoostType("bad")
	validCPUBoostTypeFactor := vpa_types.FactorStartupBoostType
	validCPUBoostTypeQuantity := vpa_types.QuantityStartupBoostType

	tests := []struct {
		name        string
		vpa         vpa_types.VerticalPodAutoscaler
		expectError error
		opts        VPAValidationOptions
	}{
		{
			name: "empty update",
			vpa:  vpa_types.VerticalPodAutoscaler{},
		},
		{
			name:        "empty create",
			vpa:         vpa_types.VerticalPodAutoscaler{},
			opts:        VPAValidationOptions{IsVPACreate: true},
			expectError: errors.New("spec.targetRef: Required value: If you're using v1beta1 version of the API, please migrate to v1"),
		},
		{
			name: "no update mode",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{},
				},
			},
			expectError: errors.New("spec.updatePolicy.updateMode: Required value: updateMode is required if UpdatePolicy is used"),
		},
		{
			name: "bad update mode",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &badUpdateMode,
					},
				},
			},
			expectError: errors.New("spec.updatePolicy.updateMode: Unsupported value: \"bad\": supported values: \"InPlace\", \"InPlaceOrRecreate\", \"Initial\", \"Off\", \"Recreate\""),
		},
		{
			name: "InPlaceOrRecreate update mode set",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &inPlaceOrRecreateUpdateMode,
					},
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
				},
			},
		},
		{
			name: "zero minReplicas",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						MinReplicas: &badMinReplicas,
						UpdateMode:  &validUpdateMode,
					},
				},
			},
			expectError: errors.New("spec.updatePolicy.minReplicas: Invalid value: 0: minReplicas has to be positive"),
		},
		{
			name: "no policy name",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{{}},
					},
				},
			},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].containerName: Required value"),
		},
		{
			name: "invalid scaling mode",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
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
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].mode: Unsupported value: \"bad\": supported values: \"Auto\", \"Off\""),
		},
		{
			name: "more than one recommender",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
						{Name: "test1"},
						{Name: "test2"},
					},
				},
			},
			expectError: errors.New("spec.recommenders: Too many: 2: must have at most 1 item"),
		},
		{
			name: "bad limits",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
							},
						},
					},
				},
			},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].maxAllowed[cpu]: Invalid value: \"10\": max resource for cpu is lower than min of \"100\""),
		},
		{
			name: "bad minAllowed cpu value",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("187500u"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("275m"),
								},
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("spec.resourcePolicy.containerPolicies[0].minAllowed[cpu]: Invalid value: \"%v\": must be a whole number of milli CPUs", badCPUResource.String()),
		},
		{
			name: "bad minAllowed memory value",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed: corev1.ResourceList{
									cpu:    resource.MustParse("1m"),
									memory: resource.MustParse("100m"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu:    resource.MustParse("275m"),
									memory: resource.MustParse("500M"),
								},
							},
						},
					},
				},
			},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].minAllowed[memory]: Invalid value: \"100m\": must be a whole number of bytes"),
		},
		{
			name: "bad maxAllowed cpu value",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed:    corev1.ResourceList{},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("187500u"),
								},
							},
						},
					},
				},
			},
			expectError: fmt.Errorf("spec.resourcePolicy.containerPolicies[0].maxAllowed[cpu]: Invalid value: \"%s\": must be a whole number of milli CPUs", badCPUResource.String()),
		},
		{
			name: "bad maxAllowed memory value",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("1m")},
								MaxAllowed: corev1.ResourceList{
									cpu:    resource.MustParse("275m"),
									memory: resource.MustParse("500m"),
								},
							},
						},
					},
				},
			},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].maxAllowed[memory]: Invalid value: \"500m\": must be a whole number of bytes"),
		},
		{
			name: "scaling off with controlled values requests and limits",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
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
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].controlledValues: Forbidden: controlledValues shouldn't be specified if container scaling mode is off"),
		},
		{
			name: "all valid",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: corev1.ResourceList{
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
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Factor: &validCPUBoostFactor,
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: false},
			expectError: fmt.Errorf("spec.startupBoost: Forbidden: in order to use startupBoost, you must enable feature gate %s in the admission-controller args", features.CPUStartupBoost),
		},
		{
			name: "container startupBoost with feature gate disabled",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
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
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: false},
			expectError: fmt.Errorf("spec.resourcePolicy.containerPolicies[0].startupBoost: Forbidden: in order to use startupBoost, you must enable feature gate %s in the admission-controller args", features.CPUStartupBoost),
		},
		{
			name: "top-level startupBoost with bad factor",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Type:   vpa_types.FactorStartupBoostType,
							Factor: &badCPUBoostFactor,
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
			expectError: errors.New("spec.startupBoost.cpu.factor: Invalid value: 0: must be >= 1 for type Factor"),
		},
		{
			name: "container startupBoost with bad factor",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
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
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].startupBoost.cpu.factor: Invalid value: 0: must be >= 1 for type Factor"),
		},
		{
			name: "top-level startupBoost with bad quantity",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
			expectError: fmt.Errorf("spec.startupBoost.cpu.quantity: Invalid value: \"%v\": must be a whole number of milli CPUs", &badCPUBoostQuantity),
		},
		{
			name: "container startupBoost with bad quantity",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
			expectError: fmt.Errorf("spec.resourcePolicy.containerPolicies[0].startupBoost.cpu.quantity: Invalid value: \"%v\": must be a whole number of milli CPUs", &badCPUBoostQuantity),
		},
		{
			name: "top-level startupBoost with bad type",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Type: badCPUBoostType,
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
			expectError: fmt.Errorf("spec.startupBoost.cpu.type: Unsupported value: \"%v\": supported values: \"%s\", \"%s\"", badCPUBoostType, vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType),
		},
		{
			name: "container startupBoost with bad type",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},

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
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
			expectError: fmt.Errorf("spec.resourcePolicy.containerPolicies[0].startupBoost.cpu.type: Unsupported value: \"%v\": supported values: \"%s\", \"%s\"", badCPUBoostType, vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType),
		},
		{
			name: "top-level startupBoost with empty type",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
			expectError: fmt.Errorf("spec.startupBoost.cpu.type: Required value: must be either %s or %s", vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType),
		},
		{
			name: "container startupBoost with empty type",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								StartupBoost: &vpa_types.StartupBoost{
									CPU: &vpa_types.GenericStartupBoost{},
								},
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
			expectError: fmt.Errorf("spec.resourcePolicy.containerPolicies[0].startupBoost.cpu.type: Required value: must be either %s or %s", vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType),
		},
		{
			name: "top-level startupBoost with valid factor",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			opts: VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
		},
		{
			name: "container startupBoost with valid factor",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			opts: VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
		},
		{
			name: "top-level startupBoost with valid quantity",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			opts: VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
		},
		{
			name: "container startupBoost with valid quantity",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
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
			opts: VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
		},
		{
			name: "top-level and container startupBoost",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{
							Type:   validCPUBoostTypeFactor,
							Factor: &validCPUBoostFactor,
						},
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
			opts: VPAValidationOptions{IsVPACreate: true, AllowCPUStartupBoost: true},
		},
		{
			name: "per-vpa config active and used",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								OOMBumpUpRatio: resource.NewQuantity(2, resource.DecimalSI),
							},
						},
					},
				},
			},
			opts: VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
		},
		{
			name: "per-vpa config active and used evictOOMThreshold",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode:           &validUpdateMode,
						EvictAfterOOMSeconds: ptr.To(int32(600)),
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								OOMBumpUpRatio: resource.NewQuantity(2, resource.DecimalSI),
							},
						},
					},
				},
			},
			opts: VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
		},
		{
			name: "Invalid oomBumpUpRatio (negative value)",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:  "*",
								Mode:           &validScalingMode,
								OOMBumpUpRatio: ptr.To(resource.MustParse("-1")),
								OOMMinBumpUp:   ptr.To(resource.MustParse("104857600")),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].oomBumpUpRatio: Invalid value: -1: must be greater than or equal to 1.0"),
		},
		{
			name: "Invalid oomBumpUpRatio (less than 1)",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:  "*",
								Mode:           &validScalingMode,
								OOMBumpUpRatio: ptr.To(resource.MustParse("0.5")),
								OOMMinBumpUp:   ptr.To(resource.MustParse("104857600")),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].oomBumpUpRatio: Invalid value: 0.5: must be greater than or equal to 1.0"),
		},
		{
			name: "Invalid oomMinBumpUp (negative value)",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:  "*",
								Mode:           &validScalingMode,
								OOMBumpUpRatio: ptr.To(resource.MustParse("2")),
								OOMMinBumpUp:   ptr.To(resource.MustParse("-1")),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].oomMinBumpUp: Invalid value: -1: must be greater than or equal to 0"),
		},
		{
			name: "per-vpa config disabled and used",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								OOMMinBumpUp: resource.NewQuantity(2, resource.DecimalSI),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: false},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].oomMinBumpUp: Forbidden: not supported when feature flag PerVPAConfig is disabled"),
		},
		{
			name: "per-vpa config active with MemoryAggregationIntervalSeconds",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								MemoryAggregationIntervalSeconds: ptr.To(int32(3600)),
							},
						},
					},
				},
			},
			opts: VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
		},
		{
			name: "per-vpa config disabled with MemoryAggregationIntervalSeconds",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								MemoryAggregationIntervalSeconds: ptr.To(int32(3600)),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: false},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].memoryAggregationIntervalSeconds: Forbidden: not supported when feature flag PerVPAConfig is disabled"),
		},
		{
			name: "per-vpa config active with MemoryAggregationIntervalCount",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								MemoryAggregationIntervalCount: ptr.To(int64(10)),
							},
						},
					},
				},
			},
			opts: VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
		},
		{
			name: "per-vpa config disabled with MemoryAggregationIntervalCount",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "loot box",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu: resource.MustParse("10"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu: resource.MustParse("100"),
								},
								MemoryAggregationIntervalCount: ptr.To(int64(10)),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: false},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].memoryAggregationIntervalCount: Forbidden: not supported when feature flag PerVPAConfig is disabled"),
		},
		{
			name: "invalid MemoryAggregationIntervalSeconds zero",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:                    "loot box",
								Mode:                             &validScalingMode,
								MemoryAggregationIntervalSeconds: ptr.To(int32(0)),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].memoryAggregationIntervalSeconds: Invalid value: 0: must be greater than or equal to 1"),
		},
		{
			name: "invalid MemoryAggregationIntervalSeconds negative",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:                    "loot box",
								Mode:                             &validScalingMode,
								MemoryAggregationIntervalSeconds: ptr.To(int32(-5)),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].memoryAggregationIntervalSeconds: Invalid value: -5: must be greater than or equal to 1"),
		},
		{
			name: "invalid MemoryAggregationIntervalCount zero",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:                  "loot box",
								Mode:                           &validScalingMode,
								MemoryAggregationIntervalCount: ptr.To(int64(0)),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].memoryAggregationIntervalCount: Invalid value: 0: must be greater than or equal to 1"),
		},
		{
			name: "invalid MemoryAggregationIntervalCount negative",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &validUpdateMode,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName:                  "loot box",
								Mode:                           &validScalingMode,
								MemoryAggregationIntervalCount: ptr.To(int64(-3)),
							},
						},
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowPerVPAConfig: true},
			expectError: errors.New("spec.resourcePolicy.containerPolicies[0].memoryAggregationIntervalCount: Invalid value: -3: must be greater than or equal to 1"),
		},
		{
			name: "creating VPA with InPlace update mode not allowed by disabled feature gate",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode: &inPlaceUpdateMode,
					},
				},
			},
			opts:        VPAValidationOptions{IsVPACreate: true, AllowInPlace: false},
			expectError: fmt.Errorf("spec.updatePolicy.updateMode: Forbidden: in order to use UpdateMode %s, you must enable feature gate %s in the admission-controller args", vpa_types.UpdateModeInPlace, features.InPlace),
		},
		{
			name: "InPlace update mode with minReplicas",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode:  &inPlaceUpdateMode,
						MinReplicas: &validMinReplicas,
					},
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
				},
			},
			opts: VPAValidationOptions{IsVPACreate: true, AllowInPlace: true},
		},
		{
			name: "InPlace update mode with complete resource policy",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					TargetRef: &autoscalingv1.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "my-app",
					},
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode:  &inPlaceUpdateMode,
						MinReplicas: &validMinReplicas,
					},
					ResourcePolicy: &vpa_types.PodResourcePolicy{
						ContainerPolicies: []vpa_types.ContainerResourcePolicy{
							{
								ContainerName: "test-container",
								Mode:          &validScalingMode,
								MinAllowed: corev1.ResourceList{
									cpu:    resource.MustParse("10m"),
									memory: resource.MustParse("100Mi"),
								},
								MaxAllowed: corev1.ResourceList{
									cpu:    resource.MustParse("1000m"),
									memory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
			opts: VPAValidationOptions{IsVPACreate: true, AllowInPlace: true},
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			if !tc.opts.AllowCPUStartupBoost {
				featuregatetesting.SetFeatureGateEmulationVersionDuringTest(t, features.MutableFeatureGate, version.MustParse("1.5"))
			} else {
				featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.CPUStartupBoost, tc.opts.AllowCPUStartupBoost)
			}
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.PerVPAConfig, tc.opts.AllowPerVPAConfig)
			errs := validateVPA(&tc.vpa, tc.opts)
			if tc.expectError == nil {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
				assert.Len(t, errs, 1)
				assert.Equal(t, tc.expectError.Error(), errs[0].Error())
			}
		})
	}
}

func TestVeryInvalidateVPA(t *testing.T) {
	vpa := vpa_types.VerticalPodAutoscaler{
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{
					{
						ContainerName: "loot box",
						StartupBoost: &vpa_types.StartupBoost{
							CPU: &vpa_types.GenericStartupBoost{},
						},
						ControlledResources: &[]corev1.ResourceName{"cpu", "memory", "broken"},
					},
				},
			},
			StartupBoost: &vpa_types.StartupBoost{},
			UpdatePolicy: &vpa_types.PodUpdatePolicy{
				MinReplicas:          ptr.To(int32(-1)),
				UpdateMode:           ptr.To(vpa_types.UpdateMode("bad")),
				EvictAfterOOMSeconds: ptr.To(int32(-1)),
			},
			Recommenders: []*vpa_types.VerticalPodAutoscalerRecommenderSelector{
				{Name: "test1"},
				{Name: "test2"},
			},
		},
	}
	expectFieldErrors := []string{
		"spec.targetRef",
		"spec.updatePolicy.updateMode",
		"spec.updatePolicy.minReplicas",
		"spec.updatePolicy.evictAfterOOMSeconds",
		"spec.recommenders",
		"spec.startupBoost",
		"spec.resourcePolicy.containerPolicies[0].startupBoost",
		"spec.resourcePolicy.containerPolicies[0].memoryAggregationIntervalCount",
		"spec.resourcePolicy.containerPolicies[0].memoryAggregationIntervalSeconds",
	}

	errs := validateVPA(&vpa, VPAValidationOptions{IsVPACreate: true})

	assert.Len(t, errs, 7)
	for _, err := range errs {
		assert.Contains(t, expectFieldErrors, err.Field)
	}
}
