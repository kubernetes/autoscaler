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
		name                                 string
		vpa                                  vpa_types.VerticalPodAutoscaler
		isCreate                             bool
		expectError                          error
		inPlaceOrRecreateFeatureGateDisabled bool
		PerVPAConfigDisabled                 bool
		inPlaceFeatureGateDisabled           bool
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
								MinAllowed:    corev1.ResourceList{},
								MaxAllowed: corev1.ResourceList{
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
			expectError: errors.New("invalid startupBoost: invalid startupBoost.cpu.factor: must be >= 1 for type Factor"),
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
			expectError: errors.New("invalid startupBoost in container loot box: invalid startupBoost.cpu.factor: must be >= 1 for type Factor"),
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
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost: invalid startupBoost.cpu.quantity: CPU [%v] must be a whole number of milli CPUs", &badCPUBoostQuantity),
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
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost in container loot box: invalid startupBoost.cpu.quantity: CPU [%v] must be a whole number of milli CPUs", &badCPUBoostQuantity),
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
			expectError: fmt.Errorf("invalid startupBoost: startupBoost.cpu.type field is required and must be either %s or %s, got %v", vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType, badCPUBoostType),
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
			expectError: fmt.Errorf("invalid startupBoost in container loot box: startupBoost.cpu.type field is required and must be either %s or %s, got %v", vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType, badCPUBoostType),
		},
		{
			name: "top-level startupBoost with empty type",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					StartupBoost: &vpa_types.StartupBoost{
						CPU: &vpa_types.GenericStartupBoost{},
					},
				},
			},
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost: startupBoost.cpu.type field is required and must be either %s or %s", vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType),
		},
		{
			name: "container startupBoost with empty type",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
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
			isCreate:    true,
			expectError: fmt.Errorf("invalid startupBoost in container loot box: startupBoost.cpu.type field is required and must be either %s or %s", vpa_types.FactorStartupBoostType, vpa_types.QuantityStartupBoostType),
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
			isCreate: true,
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
			isCreate: true,
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
			isCreate: true,
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
			isCreate: true,
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
			PerVPAConfigDisabled: false,
		},
		{
			name: "per-vpa config active and used evictOOMThreshold",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
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
			PerVPAConfigDisabled: true,
			expectError:          errors.New("OOMBumpUpRatio and OOMMinBumpUp are not supported when feature flag PerVPAConfig is disabled"),
		},
		{
			name: "InPlace update mode with minReplicas",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
					UpdatePolicy: &vpa_types.PodUpdatePolicy{
						UpdateMode:  &inPlaceUpdateMode,
						MinReplicas: &validMinReplicas,
					},
				},
			},
			inPlaceFeatureGateDisabled: false,
			expectError:                nil,
		},
		{
			name: "InPlace update mode with complete resource policy",
			vpa: vpa_types.VerticalPodAutoscaler{
				Spec: vpa_types.VerticalPodAutoscalerSpec{
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
			inPlaceFeatureGateDisabled: false,
			expectError:                nil,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			if tc.inPlaceOrRecreateFeatureGateDisabled {
				featuregatetesting.SetFeatureGateEmulationVersionDuringTest(t, features.MutableFeatureGate, version.MustParse("1.5"))
				featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlaceOrRecreate, !tc.inPlaceOrRecreateFeatureGateDisabled)
			} else {
				featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.CPUStartupBoost, !tc.cpuStartupBoostFeatureGateDisabled)
			}
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.PerVPAConfig, !tc.PerVPAConfigDisabled)
			featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.InPlace, !tc.inPlaceFeatureGateDisabled)
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
