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
	core "k8s.io/api/core/v1"

	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

type unboostAnnotationPatchCalculator struct{}

// NewUnboostAnnotationCalculator returns a calculator for the unboost annotation patch.
func NewUnboostAnnotationCalculator() patch.Calculator {
	return &unboostAnnotationPatchCalculator{}
}

// PatchResourceTarget returns the Pod resource to apply calculator patches.
func (*unboostAnnotationPatchCalculator) PatchResourceTarget() patch.PatchResourceTarget {
	return patch.Pod
}

// CalculatePatches calculates the patch to remove the startup CPU boost annotation if the pod is ready to be unboosted.
func (c *unboostAnnotationPatchCalculator) CalculatePatches(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	if vpa_api_util.IsPodReadyAndStartupBoostDurationPassed(pod, vpa) {
		return []resource_admission.PatchRecord{
			patch.GetRemoveAnnotationPatch(annotations.StartupCPUBoostAnnotation),
		}, nil
	}
	return []resource_admission.PatchRecord{}, nil
}
