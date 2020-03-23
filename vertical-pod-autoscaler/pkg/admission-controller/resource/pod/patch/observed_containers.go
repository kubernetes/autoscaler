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
	core "k8s.io/api/core/v1"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
)

type observedContainers struct{}

func (*observedContainers) CalculatePatches(pod *core.Pod, _ *vpa_types.VerticalPodAutoscaler) ([]resource_admission.PatchRecord, error) {
	vpaObservedContainersValue := annotations.GetVpaObservedContainersValue(pod)
	return []resource_admission.PatchRecord{GetAddAnnotationPatch(annotations.VpaObservedContainersLabel, vpaObservedContainersValue)}, nil
}

// NewObservedContainersCalculator returns calculator for
// observed containers patches.
func NewObservedContainersCalculator() Calculator {
	return &observedContainers{}
}
