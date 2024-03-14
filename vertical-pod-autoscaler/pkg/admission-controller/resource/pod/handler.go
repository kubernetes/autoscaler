/*
Copyright 2018 The Kubernetes Authors.

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

package pod

import (
	"context"
	"encoding/json"
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	resource_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/vpa"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	"k8s.io/klog/v2"
)

// resourceHandler builds patches for Pods.
type resourceHandler struct {
	preProcessor     PreProcessor
	vpaMatcher       vpa.Matcher
	patchCalculators []patch.Calculator
}

// NewResourceHandler creates new instance of resourceHandler.
func NewResourceHandler(preProcessor PreProcessor, vpaMatcher vpa.Matcher, patchCalculators []patch.Calculator) resource_admission.Handler {
	return &resourceHandler{
		preProcessor:     preProcessor,
		vpaMatcher:       vpaMatcher,
		patchCalculators: patchCalculators,
	}
}

// AdmissionResource returns resource type this handler accepts.
func (h *resourceHandler) AdmissionResource() admission.AdmissionResource {
	return admission.Pod
}

// GroupResource returns Group and Resource type this handler accepts.
func (h *resourceHandler) GroupResource() metav1.GroupResource {
	return metav1.GroupResource{Group: "", Resource: "pods"}
}

// DisallowIncorrectObjects decides whether incorrect objects (eg. unparsable, not passing validations) should be disallowed by Admission Server.
func (h *resourceHandler) DisallowIncorrectObjects() bool {
	// Incorrect Pods are validated by API Server.
	return false
}

// GetPatches builds patches for Pod in given admission request.
func (h *resourceHandler) GetPatches(ctx context.Context, ar *admissionv1.AdmissionRequest) ([]resource_admission.PatchRecord, error) {
	if ar.Resource.Version != "v1" {
		return nil, fmt.Errorf("only v1 Pods are supported")
	}
	raw, namespace := ar.Object.Raw, ar.Namespace
	pod := corev1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		return nil, err
	}
	if len(pod.Name) == 0 {
		pod.Name = pod.GenerateName + "%"
		pod.Namespace = namespace
	}

	// We're watching for updates now also because of in-place VPA, but we only want to act on the ones
	// that were triggered by the updater so we don't violate disruption quotas
	if ar.Operation == admissionv1.Update {
		// The patcher will remove this annotation, it's just the signal that the updater okayed this
		if _, ok := pod.Annotations["autoscaling.k8s.io/resize"]; !ok {
			return nil, nil
		}
	}

	klog.V(4).InfoS("Admitting pod", "pod", klog.KObj(&pod))
	controllingVpa := h.vpaMatcher.GetMatchingVPA(ctx, &pod)
	if controllingVpa == nil {
		klog.V(4).InfoS("No matching VPA found for pod", "pod", klog.KObj(&pod))
		return []resource_admission.PatchRecord{}, nil
	}
	pod, err := h.preProcessor.Process(pod)
	if err != nil {
		return nil, err
	}

	patches := []resource_admission.PatchRecord{}
	if pod.Annotations == nil {
		patches = append(patches, patch.GetAddEmptyAnnotationsPatch())
	}
	// TODO(jkyros): this is weird here, work it out with the above block somehow, but
	// we need to have this annotation patched out
	if ar.Operation == admissionv1.Update {
		// TODO(jkyros): the squiggle is escaping the / in the annotation
		patches = append(patches, patch.GetRemoveAnnotationPatch("autoscaling.k8s.io~1resize"))
	}

	for _, c := range h.patchCalculators {
		partialPatches, err := c.CalculatePatches(&pod, controllingVpa)
		if err != nil {
			return []resource_admission.PatchRecord{}, err
		}
		patches = append(patches, partialPatches...)
	}

	return patches, nil
}
