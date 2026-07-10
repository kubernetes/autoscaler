/*
Copyright The Kubernetes Authors.

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

// Upstream source: k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1
// These types are copied from the upstream CapacityBuffer API to avoid pulling in the entire
// autoscaler module and its transitive dependencies. Keep in sync with upstream as needed.

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:defaulter-gen=TypeMeta
// +groupName=autoscaling.x-k8s.io
package v1alpha1 // doc.go is discovered by codegen

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	// Group is the API group for CapacityBuffer resources.
	Group = "autoscaling.x-k8s.io"
)

func init() {
	gv := schema.GroupVersion{Group: Group, Version: "v1alpha1"}
	v1.AddToGroupVersion(scheme.Scheme, gv)
	scheme.Scheme.AddKnownTypes(gv,
		&CapacityBuffer{},
		&CapacityBufferList{},
	)
}
