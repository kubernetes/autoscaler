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

package clusterapi

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineDeploymentSpec is the internal autoscaler Schema for MachineDeploymentSpec
type MachineDeploymentSpec struct {
	// Number of desired machines. Defaults to 1.
	// This is a pointer to distinguish between explicit zero and not specified.
	Replicas *int32 `json:"replicas,omitempty"`

	// Label selector for machines. Existing MachineSets whose machines are
	// selected by this will be the ones affected by this deployment.
	// It must match the machine template's labels.
	Selector metav1.LabelSelector `json:"selector"`

	// Template describes the machines that will be created.
	Template MachineTemplateSpec `json:"template"`
}

// MachineDeploymentStatus is the internal autoscaler Schema for MachineDeploymentStatus
type MachineDeploymentStatus struct {
	// Number of desired machines. Defaults to 1.
	// This is a pointer to distinguish between explicit zero and not specified.
	Replicas int32 `json:"replicas,omitempty"`
}

// MachineDeployment is the internal autoscaler Schema for MachineDeployment
type MachineDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineDeploymentSpec   `json:"spec,omitempty"`
	Status MachineDeploymentStatus `json:"status,omitempty"`
}

// MachineDeploymentList is the internal autoscaler Schema for MachineDeploymentList
type MachineDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineDeployment `json:"items"`
}
