/*
Copyright 2020 The Kubernetes Authors.

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

package patch

import (
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
)

// GetAddEmptyAnnotationsPatch returns a patch initializing empty annotations.
func GetAddEmptyAnnotationsPatch() resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  "/metadata/annotations",
		Value: map[string]string{},
	}
}

// GetAddAnnotationPatch returns a patch for an annotation.
func GetAddAnnotationPatch(annotationName, annotationValue string) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/metadata/annotations/%s", annotationName),
		Value: annotationValue,
	}
}

// GetAddResourceRequirementValuePatch returns a patch record to add resource requirements to a container.
func GetAddResourceRequirementValuePatch(i int, kind string, resource core.ResourceName, quantity resource.Quantity) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources/%s/%s", i, kind, resource),
		Value: quantity.String()}
}

// GetPatchInitializingEmptyResources returns a patch record to initialize an empty resources object for a container.
func GetPatchInitializingEmptyResources(i int) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources", i),
		Value: core.ResourceRequirements{},
	}
}

// GetPatchInitializingEmptyResourcesSubfield returns a patch record to initialize an empty subfield
// (e.g., "requests" or "limits") within a container's resources object.
func GetPatchInitializingEmptyResourcesSubfield(i int, kind string) resource_admission.PatchRecord {
	return resource_admission.PatchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources/%s", i, kind),
		Value: core.ResourceList{},
	}
}
