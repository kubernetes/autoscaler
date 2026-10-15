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

package patch

import (
	corev1 "k8s.io/api/core/v1"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// VpaManagedLabel marks pods managed by a VPA, so they can be found with a label selector.
const VpaManagedLabel = "vpa-autoscaler.k8s.io/managed"

type managedLabel struct{}

func (*managedLabel) CalculatePatches(_ *corev1.Pod, _ *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	return []resource_admission.PatchRecord{GetAddLabelPatch(VpaManagedLabel, "true")}, nil
}

func (*managedLabel) PatchResourceTarget() PatchResourceTarget {
	return Pod
}

// NewManagedLabelCalculator returns a calculator that adds VpaManagedLabel to pods on creation.
func NewManagedLabelCalculator() Calculator {
	return &managedLabel{}
}
