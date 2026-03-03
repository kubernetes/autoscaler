/*
Copyright 2025 The Kubernetes Authors.

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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
)

type inPlaceUpdate struct{}

// CalculatePatches returns a patch that adds a "vpaInPlaceUpdated" annotation
// to the pod, marking it as having been requested to be updated in-place by VPA.
func (*inPlaceUpdate) CalculatePatches(pod *corev1.Pod, _ *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	vpaInPlaceUpdatedValue := annotations.GetVpaInPlaceUpdatedValue()
	return []resource_admission.PatchRecord{patch.GetAddAnnotationPatch(annotations.VpaInPlaceUpdatedLabel, vpaInPlaceUpdatedValue)}, nil
}

func (*inPlaceUpdate) PatchResourceTarget() patch.PatchResourceTarget {
	return patch.Pod
}

// NewInPlaceUpdatedCalculator returns calculator for
// observed containers patches.
func NewInPlaceUpdatedCalculator() patch.Calculator {
	return &inPlaceUpdate{}
}
