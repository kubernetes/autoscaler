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

package mpa

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	apires "k8s.io/apimachinery/pkg/api/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	"k8s.io/klog/v2"
)

var (
	possibleUpdateModes = map[vpa_types.UpdateMode]interface{}{
		vpa_types.UpdateModeOff:      struct{}{},
		vpa_types.UpdateModeInitial:  struct{}{},
		vpa_types.UpdateModeRecreate: struct{}{},
		vpa_types.UpdateModeAuto:     struct{}{},
	}

	possibleScalingModes = map[vpa_types.ContainerScalingMode]interface{}{
		vpa_types.ContainerScalingModeAuto: struct{}{},
		vpa_types.ContainerScalingModeOff:  struct{}{},
	}
)

// resourceHandler builds patches for VPAs.
type resourceHandler struct {
	preProcessor PreProcessor
}

// NewResourceHandler creates new instance of resourceHandler.
func NewResourceHandler(preProcessor PreProcessor) resource.Handler {
	return &resourceHandler{preProcessor: preProcessor}
}

// AdmissionResource returns resource type this handler accepts.
func (h *resourceHandler) AdmissionResource() admission.AdmissionResource {
	return admission.Vpa
}

// GroupResource returns Group and Resource type this handler accepts.
func (h *resourceHandler) GroupResource() metav1.GroupResource {
	return metav1.GroupResource{Group: "autoscaling.k8s.io", Resource: "verticalpodautoscalers"}
}

// DisallowIncorrectObjects decides whether incorrect objects (eg. unparsable, not passing validations) should be disallowed by Admission Server.
func (h *resourceHandler) DisallowIncorrectObjects() bool {
	return true
}

// GetPatches builds patches for VPA in given admission request.
func (h *resourceHandler) GetPatches(ar *v1.AdmissionRequest) ([]resource.PatchRecord, error) {
	raw, isCreate := ar.Object.Raw, ar.Operation == v1.Create
	mpa, err := parseMPA(raw)
	if err != nil {
		return nil, err
	}

	mpa, err = h.preProcessor.Process(mpa, isCreate)
	if err != nil {
		return nil, err
	}

	err = ValidateMPA(mpa, isCreate)
	if err != nil {
		return nil, err
	}

	klog.V(4).Infof("Processing MPA: %v", mpa)
	patches := []resource.PatchRecord{}
	if mpa.Spec.UpdatePolicy == nil {
		// Sets the default updatePolicy.
		defaultUpdateMode := vpa_types.UpdateModeAuto
		patches = append(patches, resource.PatchRecord{
			Op:    "add",
			Path:  "/spec/updatePolicy",
			Value: vpa_types.PodUpdatePolicy{UpdateMode: &defaultUpdateMode}})
	}
	return patches, nil
}

func parseMPA(raw []byte) (*mpa_types.MultidimPodAutoscaler, error) {
	mpa := mpa_types.MultidimPodAutoscaler{}
	if err := json.Unmarshal(raw, &mpa); err != nil {
		return nil, err
	}
	return &mpa, nil
}

// ValidateMPA checks the correctness of MPA Spec and returns an error if there is a problem.
func ValidateMPA(mpa *mpa_types.MultidimPodAutoscaler, isCreate bool) error {
	if mpa.Spec.UpdatePolicy != nil {
		mode := mpa.Spec.UpdatePolicy.UpdateMode
		if mode == nil {
			return fmt.Errorf("UpdateMode is required if UpdatePolicy is used")
		}
		if _, found := possibleUpdateModes[*mode]; !found {
			return fmt.Errorf("unexpected UpdateMode value %s", *mode)
		}

		if minReplicas := mpa.Spec.Constraints.MinReplicas; minReplicas != nil && *minReplicas <= 0 {
			return fmt.Errorf("MinReplicas has to be positive, got %v", *minReplicas)
		}
	}

	if mpa.Spec.ResourcePolicy != nil {
		for _, policy := range mpa.Spec.ResourcePolicy.ContainerPolicies {
			if policy.ContainerName == "" {
				return fmt.Errorf("ContainerPolicies.ContainerName is required")
			}
			mode := policy.Mode
			if mode != nil {
				if _, found := possibleScalingModes[*mode]; !found {
					return fmt.Errorf("unexpected Mode value %s", *mode)
				}
			}
			for resource, min := range policy.MinAllowed {
				if err := validateResourceResolution(resource, min); err != nil {
					return fmt.Errorf("MinAllowed: %v", err)
				}
				max, found := policy.MaxAllowed[resource]
				if found && max.Cmp(min) < 0 {
					return fmt.Errorf("max resource for %v is lower than min", resource)
				}
			}

			for resource, max := range policy.MaxAllowed {
				if err := validateResourceResolution(resource, max); err != nil {
					return fmt.Errorf("MaxAllowed: %v", err)
				}
			}
			ControlledValues := policy.ControlledValues
			if mode != nil && ControlledValues != nil {
				if *mode == vpa_types.ContainerScalingModeOff && *ControlledValues == vpa_types.ContainerControlledValuesRequestsAndLimits {
					return fmt.Errorf("ControlledValues shouldn't be specified if container scaling mode is off.")
				}
			}
		}
	}

	if isCreate && mpa.Spec.ScaleTargetRef == nil {
		return fmt.Errorf("ScaleTargetRef is required.")
	}

	if len(mpa.Spec.Recommenders) > 1 {
		return fmt.Errorf("The current version of MPA object shouldn't specify more than one recommenders.")
	}

	return nil
}

func validateResourceResolution(name corev1.ResourceName, val apires.Quantity) error {
	switch name {
	case corev1.ResourceCPU:
		return validateCPUResolution(val)
	case corev1.ResourceMemory:
		return validateMemoryResolution(val)
	}
	return nil
}

func validateCPUResolution(val apires.Quantity) error {
	if _, precissionPreserved := val.AsScale(apires.Milli); !precissionPreserved {
		return fmt.Errorf("CPU [%s] must be a whole number of milli CPUs", val.String())
	}
	return nil
}

func validateMemoryResolution(val apires.Quantity) error {
	if _, precissionPreserved := val.AsScale(0); !precissionPreserved {
		return fmt.Errorf("Memory [%v] must be a whole number of bytes", val)
	}
	return nil
}
