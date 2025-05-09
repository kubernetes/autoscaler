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

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// PatchResourceTarget is the type of resource that can be patched.
type PatchResourceTarget string

const (
	// Pod refers to the pod resource itself.
	Pod PatchResourceTarget = "Pod"
	// Resize refers to the resize subresource of the pod.
	Resize PatchResourceTarget = "Resize"

	// Future subresources can be added here.
	//  e.g. Status PatchResourceTarget = "Status"
)

// Calculator is capable of calculating required patches for pod.
type Calculator interface {
	CalculatePatches(pod *core.Pod, vpa *vpa_types.VerticalPodAutoscaler) ([]resource.PatchRecord, error)
	// PatchResourceTarget returns the resource this calculator should calculate patches for.
	PatchResourceTarget() PatchResourceTarget
}
