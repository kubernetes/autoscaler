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
	"io/ioutil"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"net/http"

	"strings"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	metrics_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/klog"
)

// AdmissionServer is an admission webhook server that modifies pod resources request based on VPA recommendation
type AdmissionServer struct {
	recommendationProvider RecommendationProvider
	podPreProcessor        PodPreProcessor
	vpaPreProcessor        VpaPreProcessor
	limitsChecker          limitrange.LimitRangeCalculator
}

// NewAdmissionServer constructs new AdmissionServer
func NewAdmissionServer(recommendationProvider RecommendationProvider, podPreProcessor PodPreProcessor, vpaPreProcessor VpaPreProcessor, limitsChecker limitrange.LimitRangeCalculator) *AdmissionServer {
	return &AdmissionServer{recommendationProvider, podPreProcessor, vpaPreProcessor, limitsChecker}
}

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func (s *AdmissionServer) getPatchesForPodResourceRequest(raw []byte, namespace string) ([]patchRecord, error) {
	pod := v1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		return nil, err
	}
	if len(pod.Name) == 0 {
		pod.Name = pod.GenerateName + "%"
		pod.Namespace = namespace
	}
	klog.V(4).Infof("Admitting pod %v", pod.ObjectMeta)
	containersResources, annotationsPerContainer, vpaName, err := s.recommendationProvider.GetContainersResourcesForPod(&pod)
	if err != nil {
		return nil, err
	}
	pod, err = s.podPreProcessor.Process(pod)
	if err != nil {
		return nil, err
	}
	if annotationsPerContainer == nil {
		annotationsPerContainer = vpa_api_util.ContainerToAnnotationsMap{}
	}

	patches := []patchRecord{}
	updatesAnnotation := []string{}
	for i, containerResources := range containersResources {
		newPatches, newUpdatesAnnotation := s.getContainerPatch(pod, i, annotationsPerContainer, containerResources)
		patches = append(patches, newPatches...)
		updatesAnnotation = append(updatesAnnotation, newUpdatesAnnotation)
	}
	if len(updatesAnnotation) > 0 {
		vpaAnnotationValue := fmt.Sprintf("Pod resources updated by %s: %s", vpaName, strings.Join(updatesAnnotation, "; "))
		if pod.Annotations == nil {
			patches = append(patches, patchRecord{
				Op:    "add",
				Path:  "/metadata/annotations",
				Value: map[string]string{"vpaUpdates": vpaAnnotationValue}})
		} else {
			patches = append(patches, patchRecord{
				Op:    "add",
				Path:  "/metadata/annotations/vpaUpdates",
				Value: vpaAnnotationValue})
		}
	}
	return patches, nil
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

func (s *AdmissionServer) getContainerPatch(pod v1.Pod, i int, annotationsPerContainer vpa_api_util.ContainerToAnnotationsMap, containerResources vpa_api_util.ContainerResources) ([]patchRecord, string) {
	var patches []patchRecord
	// Add empty resources object if missing
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

func parseVPA(raw []byte) (*vpa_types.VerticalPodAutoscaler, error) {
	vpa := vpa_types.VerticalPodAutoscaler{}
	if err := json.Unmarshal(raw, &vpa); err != nil {
		return nil, err
	}
	return &vpa, nil
}

var (
	possibleUpdateModes = map[vpa_types.UpdateMode]interface{}{
		vpa_types.UpdateModeOff:      struct{}{},
		vpa_types.UpdateModeInitial:  struct{}{},
		vpa_types.UpdateModeRecreate: struct{}{},
		vpa_types.UpdateModeAuto:     struct{}{},
	}

	possibleScalingModes = map[vpa_types.ContainerScalingMode]interface{}{
		vpa_types.ContainerScalingModeAuto: struct{}{},
		vpa_types.ContainerScalingModeOff:  struct{}{},
	}
)

func validateVPA(vpa *vpa_types.VerticalPodAutoscaler, isCreate bool) error {
	if vpa.Spec.UpdatePolicy != nil {
		mode := vpa.Spec.UpdatePolicy.UpdateMode
		if mode == nil {
			return fmt.Errorf("UpdateMode is required if UpdatePolicy is used")
		}
		if _, found := possibleUpdateModes[*mode]; !found {
			return fmt.Errorf("unexpected UpdateMode value %s", *mode)
		}
	}

	if vpa.Spec.ResourcePolicy != nil {
		for _, policy := range vpa.Spec.ResourcePolicy.ContainerPolicies {
			if policy.ContainerName == "" {
				return fmt.Errorf("ContainerPolicies.ContainerName is required")
			}
			mode := policy.Mode
			if mode != nil {
				if _, found := possibleScalingModes[*mode]; !found {
					return fmt.Errorf("unexpected Mode value %s", *mode)
				}
			}
			for resource, min := range policy.MinAllowed {
				max, found := policy.MaxAllowed[resource]
				if found && max.Cmp(min) < 0 {
					return fmt.Errorf("max resource for %v is lower than min", resource)
				}
			}
		}
	}

	if isCreate && vpa.Spec.TargetRef == nil {
		return fmt.Errorf("TargetRef is required. If you're using v1beta1 version of the API, please migrate to v1.")
	}

	return nil
}

func (s *AdmissionServer) getPatchesForVPADefaults(raw []byte, isCreate bool) ([]patchRecord, error) {
	vpa, err := parseVPA(raw)
	if err != nil {
		return nil, err
	}

	vpa, err = s.vpaPreProcessor.Process(vpa, isCreate)
	if err != nil {
		return nil, err
	}

	err = validateVPA(vpa, isCreate)
	if err != nil {
		return nil, err
	}

	klog.V(4).Infof("Processing vpa: %v", vpa)
	patches := []patchRecord{}
	if vpa.Spec.UpdatePolicy == nil {
		// Sets the default updatePolicy.
		defaultUpdateMode := vpa_types.UpdateModeAuto
		patches = append(patches, patchRecord{
			Op:    "add",
			Path:  "/spec/updatePolicy",
			Value: vpa_types.PodUpdatePolicy{UpdateMode: &defaultUpdateMode}})
	}
	return patches, nil
}

func (s *AdmissionServer) admit(data []byte) (*v1beta1.AdmissionResponse, metrics_admission.AdmissionStatus, metrics_admission.AdmissionResource) {
	// we don't block the admission by default, even on unparsable JSON
	response := v1beta1.AdmissionResponse{}
	response.Allowed = true

	ar := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(data, &ar); err != nil {
		klog.Error(err)
		return &response, metrics_admission.Error, metrics_admission.Unknown
	}
	// The externalAdmissionHookConfiguration registered via selfRegistration
	// asks the kube-apiserver only to send admission requests regarding pods & VPA objects.
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	vpaResourceV1beta2 := metav1.GroupVersionResource{Group: "autoscaling.k8s.io", Version: "v1beta2", Resource: "verticalpodautoscalers"}
	vpaResourceV1 := metav1.GroupVersionResource{Group: "autoscaling.k8s.io", Version: "v1", Resource: "verticalpodautoscalers"}

	var patches []patchRecord
	var err error
	resource := metrics_admission.Unknown

	switch ar.Request.Resource {
	case podResource:
		patches, err = s.getPatchesForPodResourceRequest(ar.Request.Object.Raw, ar.Request.Namespace)
		resource = metrics_admission.Pod
	case vpaResourceV1, vpaResourceV1beta2:
		patches, err = s.getPatchesForVPADefaults(ar.Request.Object.Raw, ar.Request.Operation == v1beta1.Create)
		resource = metrics_admission.Vpa
		// we don't let in problematic VPA objects - late validation
		if err != nil {
			status := metav1.Status{}
			status.Status = "Failure"
			status.Message = err.Error()
			response.Result = &status
			response.Allowed = false
		}
	default:
		patches, err = nil, fmt.Errorf("expected the resource to be %v, %v or %v", podResource, vpaResourceV1beta2, vpaResourceV1)
	}

	if err != nil {
		klog.Error(err)
		return &response, metrics_admission.Error, resource
	}

	if len(patches) > 0 {
		patch, err := json.Marshal(patches)
		if err != nil {
			klog.Errorf("Cannot marshal the patch %v: %v", patches, err)
			return &response, metrics_admission.Error, resource
		}
		patchType := v1beta1.PatchTypeJSONPatch
		response.PatchType = &patchType
		response.Patch = patch
		klog.V(4).Infof("Sending patches: %v", patches)
	}

	var status metrics_admission.AdmissionStatus
	if len(patches) > 0 {
		status = metrics_admission.Applied
	} else {
		status = metrics_admission.Skipped
	}
	if resource == metrics_admission.Pod {
		metrics_admission.OnAdmittedPod(status == metrics_admission.Applied)
	}

	return &response, status, resource
}

// Serve is a handler function of AdmissionServer
func (s *AdmissionServer) Serve(w http.ResponseWriter, r *http.Request) {
	timer := metrics_admission.NewAdmissionLatency()

	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		timer.Observe(metrics_admission.Error, metrics_admission.Unknown)
		return
	}

	reviewResponse, status, resource := s.admit(body)
	ar := v1beta1.AdmissionReview{
		Response: reviewResponse,
	}

	resp, err := json.Marshal(ar)
	if err != nil {
		klog.Error(err)
		timer.Observe(metrics_admission.Error, resource)
		return
	}

	if _, err := w.Write(resp); err != nil {
		klog.Error(err)
		timer.Observe(metrics_admission.Error, resource)
		return
	}

	timer.Observe(status, resource)
}
