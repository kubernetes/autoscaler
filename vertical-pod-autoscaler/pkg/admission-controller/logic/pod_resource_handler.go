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

package logic

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/klog"
)

const (
	vpaAnnotationLabel = "vpaUpdates"
)

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

// podResourceHandler builds patches for Pods.
type podResourceHandler struct {
	podPreProcessor        PodPreProcessor
	recommendationProvider RecommendationProvider
	vpaMatcher             VpaMatcher
}

// newPodResourceHandler creates new instance of podResourceHandler.
func newPodResourceHandler(podPreProcessor PodPreProcessor, recommendationProvider RecommendationProvider, vpaMatcher VpaMatcher) ResourceHandler {
	return &podResourceHandler{
		podPreProcessor:        podPreProcessor,
		recommendationProvider: recommendationProvider,
		vpaMatcher:             vpaMatcher,
	}
}

// AdmissionResource returns resource type this handler accepts.
func (h *podResourceHandler) AdmissionResource() admission.AdmissionResource {
	return admission.Pod
}

// GroupResource returns Group and Resource type this handler accepts.
func (h *podResourceHandler) GroupResource() metav1.GroupResource {
	return metav1.GroupResource{Group: "", Resource: "pods"}
}

// DisallowIncorrectObjects decides whether incorrect objects (eg. unparsable, not passing validations) should be disallowed by Admission Server.
func (h *podResourceHandler) DisallowIncorrectObjects() bool {
	// Incorrect Pods are validated by API Server.
	return false
}

// GetPatches builds patches for Pod in given admission request.
func (h *podResourceHandler) GetPatches(ar *v1beta1.AdmissionRequest) ([]patchRecord, error) {
	if ar.Resource.Version != "v1" {
		return nil, fmt.Errorf("only v1 Pods are supported")
	}
	raw, namespace := ar.Object.Raw, ar.Namespace
	pod := v1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		return nil, err
	}
	if len(pod.Name) == 0 {
		pod.Name = pod.GenerateName + "%"
		pod.Namespace = namespace
	}
	klog.V(4).Infof("Admitting pod %v", pod.ObjectMeta)
	controllingVpa := h.vpaMatcher.GetMatchingVPA(&pod)
	if controllingVpa == nil {
		klog.V(4).Infof("No matching VPA found for pod %s/%s", pod.Namespace, pod.Name)
		return []patchRecord{}, nil
	}
	containersResources, annotationsPerContainer, err := h.recommendationProvider.GetContainersResourcesForPod(&pod, controllingVpa)
	if err != nil {
		return nil, err
	}
	pod, err = h.podPreProcessor.Process(pod)
	if err != nil {
		return nil, err
	}
	if annotationsPerContainer == nil {
		annotationsPerContainer = vpa_api_util.ContainerToAnnotationsMap{}
	}

	patches := []patchRecord{}
	updatesAnnotation := []string{}
	for i, containerResources := range containersResources {
		newPatches, newUpdatesAnnotation := getContainerPatch(pod, i, annotationsPerContainer, containerResources)
		patches = append(patches, newPatches...)
		updatesAnnotation = append(updatesAnnotation, newUpdatesAnnotation)
	}

	if pod.Annotations == nil {
		patches = append(patches, getAddEmptyAnnotationsPatch())
	}
	if len(updatesAnnotation) > 0 {
		vpaAnnotationValue := fmt.Sprintf("Pod resources updated by %s: %s", controllingVpa.Name, strings.Join(updatesAnnotation, "; "))
		patches = append(patches, getAddAnnotationPatch(vpaAnnotationLabel, vpaAnnotationValue))
	}
	vpaObservedContainersValue := annotations.GetVpaObservedContainersValue(&pod)
	patches = append(patches, getAddAnnotationPatch(annotations.VpaObservedContainersLabel, vpaObservedContainersValue))

	return patches, nil
}

func getContainerPatch(pod v1.Pod, i int, annotationsPerContainer vpa_api_util.ContainerToAnnotationsMap, containerResources vpa_api_util.ContainerResources) ([]patchRecord, string) {
	var patches []patchRecord
	// Add empty resources object if missing.
	if pod.Spec.Containers[i].Resources.Limits == nil &&
		pod.Spec.Containers[i].Resources.Requests == nil {
		patches = append(patches, getPatchInitializingEmptyResources(i))
	}

	annotations, found := annotationsPerContainer[pod.Spec.Containers[i].Name]
	if !found {
		annotations = make([]string, 0)
	}

	patches, annotations = appendPatchesAndAnnotations(patches, annotations, pod.Spec.Containers[i].Resources.Requests, i, containerResources.Requests, "requests", "request")
	patches, annotations = appendPatchesAndAnnotations(patches, annotations, pod.Spec.Containers[i].Resources.Limits, i, containerResources.Limits, "limits", "limit")

	updatesAnnotation := fmt.Sprintf("container %d: ", i) + strings.Join(annotations, ", ")
	return patches, updatesAnnotation
}

func getAddEmptyAnnotationsPatch() patchRecord {
	return patchRecord{
		Op:    "add",
		Path:  "/metadata/annotations",
		Value: map[string]string{},
	}
}

func getAddAnnotationPatch(annotationName, annotationValue string) patchRecord {
	return patchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/metadata/annotations/%s", annotationName),
		Value: annotationValue,
	}
}

func getPatchInitializingEmptyResources(i int) patchRecord {
	return patchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources", i),
		Value: v1.ResourceRequirements{},
	}
}

func getPatchInitializingEmptyResourcesSubfield(i int, kind string) patchRecord {
	return patchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources/%s", i, kind),
		Value: v1.ResourceList{},
	}
}

func getAddResourceRequirementValuePatch(i int, kind string, resource v1.ResourceName, quantity resource.Quantity) patchRecord {
	return patchRecord{
		Op:    "add",
		Path:  fmt.Sprintf("/spec/containers/%d/resources/%s/%s", i, kind, resource),
		Value: quantity.String()}
}

func appendPatchesAndAnnotations(patches []patchRecord, annotations []string, current v1.ResourceList, containerIndex int, resources v1.ResourceList, fieldName, resourceName string) ([]patchRecord, []string) {
	// Add empty object if it's missing and we're about to fill it.
	if current == nil && len(resources) > 0 {
		patches = append(patches, getPatchInitializingEmptyResourcesSubfield(containerIndex, fieldName))
	}
	for resource, request := range resources {
		patches = append(patches, getAddResourceRequirementValuePatch(containerIndex, fieldName, resource, request))
		annotations = append(annotations, fmt.Sprintf("%s %s", resource, resourceName))
	}
	return patches, annotations
}
