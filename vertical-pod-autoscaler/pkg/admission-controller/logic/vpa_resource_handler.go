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
	"encoding/json"
	"fmt"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	"k8s.io/klog"
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

// vpaResourceHandler builds patches for VPAs.
type vpaResourceHandler struct {
	vpaPreProcessor VpaPreProcessor
}

// newVpaResourceHandler creates new instance of vpaResourceHandler.
func newVpaResourceHandler(vpaPreProcessor VpaPreProcessor) ResourceHandler {
	return &vpaResourceHandler{vpaPreProcessor: vpaPreProcessor}
}

// AdmissionResource returns resource type this handler accepts.
func (h *vpaResourceHandler) AdmissionResource() admission.AdmissionResource {
	return admission.Vpa
}

// GroupResource returns Group and Resource type this handler accepts.
func (h *vpaResourceHandler) GroupResource() metav1.GroupResource {
	return metav1.GroupResource{Group: "autoscaling.k8s.io", Resource: "verticalpodautoscalers"}
}

// DisallowIncorrectObjects decides whether incorrect objects (eg. unparsable, not passing validations) should be disallowed by Admission Server.
func (h *vpaResourceHandler) DisallowIncorrectObjects() bool {
	return true
}

// GetPatches builds patches for VPA in given admission request.
func (h *vpaResourceHandler) GetPatches(ar *v1beta1.AdmissionRequest) ([]patchRecord, error) {
	raw, isCreate := ar.Object.Raw, ar.Operation == v1beta1.Create
	vpa, err := parseVPA(raw)
	if err != nil {
		return nil, err
	}

	vpa, err = h.vpaPreProcessor.Process(vpa, isCreate)
	if err != nil {
		return nil, err
	}

	err = validateVPA(vpa, isCreate)
	if err != nil {
		return nil, err
	}

	klog.V(4).Infof("Processing vpa: %v", vpa)
	patches := []patchRecord{}
	if vpa.Spec.UpdatePolicy == nil {
		// Sets the default updatePolicy.
		defaultUpdateMode := vpa_types.UpdateModeAuto
		patches = append(patches, patchRecord{
			Op:    "add",
			Path:  "/spec/updatePolicy",
			Value: vpa_types.PodUpdatePolicy{UpdateMode: &defaultUpdateMode}})
	}
	return patches, nil
}

func parseVPA(raw []byte) (*vpa_types.VerticalPodAutoscaler, error) {
	vpa := vpa_types.VerticalPodAutoscaler{}
	if err := json.Unmarshal(raw, &vpa); err != nil {
		return nil, err
	}
	return &vpa, nil
}

func validateVPA(vpa *vpa_types.VerticalPodAutoscaler, isCreate bool) error {
	if vpa.Spec.UpdatePolicy != nil {
		mode := vpa.Spec.UpdatePolicy.UpdateMode
		if mode == nil {
			return fmt.Errorf("UpdateMode is required if UpdatePolicy is used")
		}
		if _, found := possibleUpdateModes[*mode]; !found {
			return fmt.Errorf("unexpected UpdateMode value %s", *mode)
		}
	}

	if vpa.Spec.ResourcePolicy != nil {
		for _, policy := range vpa.Spec.ResourcePolicy.ContainerPolicies {
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
				max, found := policy.MaxAllowed[resource]
				if found && max.Cmp(min) < 0 {
					return fmt.Errorf("max resource for %v is lower than min", resource)
				}
			}
		}
	}

	if isCreate && vpa.Spec.TargetRef == nil {
		return fmt.Errorf("TargetRef is required. If you're using v1beta1 version of the API, please migrate to v1")
	}

	return nil
}
