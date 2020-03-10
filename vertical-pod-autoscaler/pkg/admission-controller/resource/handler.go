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

package resource

import (
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
)

// PatchRecord represents a single patch for modifying a resource.
type PatchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

// Handler represents a handler for a resource in Admission Server
type Handler interface {
	// GroupResource returns Group and Resource type this handler accepts.
	GroupResource() metav1.GroupResource
	// AdmissionResource returns resource type this handler accepts.
	AdmissionResource() admission.AdmissionResource
	// DisallowIncorrectObjects returns whether incorrect objects (eg. unparsable, not passing validations) should be disallowed by Admission Server.
	DisallowIncorrectObjects() bool
	// GetPatches returns patches for given AdmissionRequest
	GetPatches(*v1beta1.AdmissionRequest) ([]PatchRecord, error)
}
