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

package inplace

import (
	corev1 "k8s.io/api/core/v1"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

type managedLabel struct{}

// CalculatePatches returns patches that add VpaManagedLabel to the pod.
// The admission controller only labels pods on creation, so pods resized in-place need to be labeled here.
func (*managedLabel) CalculatePatches(pod *corev1.Pod, _ *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	patches := []resource_admission.PatchRecord{}
	if pod.Labels == nil {
		patches = append(patches, patch.GetAddEmptyLabelsPatch())
	}
	return append(patches, patch.GetAddLabelPatch(patch.VpaManagedLabel, "true")), nil
}

func (*managedLabel) PatchResourceTarget() patch.PatchResourceTarget {
	return patch.Pod
}

// NewManagedLabelCalculator returns a calculator that adds VpaManagedLabel to in-place updated pods.
func NewManagedLabelCalculator() patch.Calculator {
	return &managedLabel{}
}
