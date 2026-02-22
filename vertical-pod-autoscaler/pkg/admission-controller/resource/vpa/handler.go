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
	"context"
	"encoding/json"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
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
func (*resourceHandler) AdmissionResource() admission.AdmissionResource {
	return admission.Vpa
}

// GroupResource returns Group and Resource type this handler accepts.
func (*resourceHandler) GroupResource() metav1.GroupResource {
	return metav1.GroupResource{Group: "autoscaling.k8s.io", Resource: "verticalpodautoscalers"}
}

// DisallowIncorrectObjects decides whether incorrect objects (eg. unparsable, not passing validations) should be disallowed by Admission Server.
func (*resourceHandler) DisallowIncorrectObjects() bool {
	return true
}

// GetPatches builds patches for VPA in given admission request.
func (h *resourceHandler) GetPatches(_ context.Context, ar *admissionv1.AdmissionRequest) ([]resource.PatchRecord, field.ErrorList) {
	raw, isCreate := ar.Object.Raw, ar.Operation == admissionv1.Create
	vpa, err := parseVPA(raw)
	if err != nil {
		return nil, field.ErrorList{field.InternalError(field.NewPath("."), err)}
	}

	oldVPA := &vpa_types.VerticalPodAutoscaler{}

	if ar.Operation == admissionv1.Update {
		oldRaw := ar.OldObject.Raw
		oldVPA, err = parseVPA(oldRaw)
		if err != nil {
			return nil, field.ErrorList{field.InternalError(field.NewPath("."), err)}
		}
	}

	opts := getValidationOptionsForVPA(oldVPA)

	vpa, err = h.preProcessor.Process(vpa, isCreate)
	if err != nil {
		return nil, field.ErrorList{field.InternalError(field.NewPath("."), err)}
	}

	if allErrs := validateVPA(vpa, opts); len(allErrs) > 0 {
		return nil, allErrs
	}

	klog.V(4).InfoS("Processing vpa", "vpa", vpa)
	patches := []resource.PatchRecord{}
	if vpa.Spec.UpdatePolicy == nil {
		// Sets the default updatePolicy.
		// Changed from UpdateModeAuto to UpdateModeRecreate as part of Auto mode deprecation
		defaultUpdateMode := vpa_types.UpdateModeRecreate
		patches = append(patches, resource.PatchRecord{
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
