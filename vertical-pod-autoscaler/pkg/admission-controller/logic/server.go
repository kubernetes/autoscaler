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
	"net/http"

	"strings"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	metrics_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// AdmissionServer is an admission webhook server that modifies pod resources request based on VPA recommendation
type AdmissionServer struct {
	recommendationProvider RecommendationProvider
	podPreProcessor        PodPreProcessor
}

// NewAdmissionServer constructs new AdmissionServer
func NewAdmissionServer(recommendationProvider RecommendationProvider, podPreProcessor PodPreProcessor) *AdmissionServer {
	return &AdmissionServer{recommendationProvider, podPreProcessor}
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
	glog.V(4).Infof("Admitting pod %v", pod.ObjectMeta)
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

		// Add resources empty object if missing
		if pod.Spec.Containers[i].Resources.Limits == nil &&
			pod.Spec.Containers[i].Resources.Requests == nil {
			patches = append(patches, patchRecord{
				Op:    "add",
				Path:  fmt.Sprintf("/spec/containers/%d/resources", i),
				Value: v1.ResourceRequirements{},
			})
		}

		// Add request empty map if missing
		if pod.Spec.Containers[i].Resources.Requests == nil {
			patches = append(patches, patchRecord{
				Op:    "add",
				Path:  fmt.Sprintf("/spec/containers/%d/resources/requests", i),
				Value: v1.ResourceList{}})
		}

		annotations, found := annotationsPerContainer[pod.Spec.Containers[i].Name]
		if !found {
			annotations = make([]string, 0)
		}
		for resource, request := range containerResources.Requests {
			// Set request
			patches = append(patches, patchRecord{
				Op:    "add",
				Path:  fmt.Sprintf("/spec/containers/%d/resources/requests/%s", i, resource),
				Value: request.String()})
			annotations = append(annotations, fmt.Sprintf("%s request", resource))
		}

		updatesAnnotation = append(updatesAnnotation, fmt.Sprintf("container %d: ", i)+strings.Join(annotations, ", "))
	}
	if len(updatesAnnotation) > 0 {
		vpaAnnotationValue := fmt.Sprintf("Pod resources updated by %s: ", vpaName) + strings.Join(updatesAnnotation, "; ")
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

func getPatchesForVPADefaults(raw []byte) ([]patchRecord, error) {
	vpa := vpa_types.VerticalPodAutoscaler{}
	if err := json.Unmarshal(raw, &vpa); err != nil {
		return nil, err
	}
	glog.V(4).Infof("Processing vpa: %v", vpa)
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
	ar := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(data, &ar); err != nil {
		glog.Error(err)
		return nil, metrics_admission.Error, metrics_admission.Unknown
	}
	// The externalAdmissionHookConfiguration registered via selfRegistration
	// asks the kube-apiserver only to send admission requests regarding pods & VPA objects.
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	vpaResource := metav1.GroupVersionResource{Group: "poc.autoscaling.k8s.io", Version: "v1alpha1", Resource: "verticalpodautoscalers"}
	var patches []patchRecord
	var err error
	resource := metrics_admission.Unknown

	switch ar.Request.Resource {
	case podResource:
		patches, err = s.getPatchesForPodResourceRequest(ar.Request.Object.Raw, ar.Request.Namespace)
		resource = metrics_admission.Pod
	case vpaResource:
		patches, err = getPatchesForVPADefaults(ar.Request.Object.Raw)
		resource = metrics_admission.Vpa
	default:
		patches, err = nil, fmt.Errorf("expected the resource to be %v or %v", podResource, vpaResource)
	}

	if err != nil {
		glog.Error(err)
		return nil, metrics_admission.Error, resource
	}
	response := v1beta1.AdmissionResponse{}
	response.Allowed = true
	if len(patches) > 0 {
		patch, err := json.Marshal(patches)
		if err != nil {
			glog.Errorf("Cannot marshal the patch %v: %v", patches, err)
			return nil, metrics_admission.Error, resource
		}
		patchType := v1beta1.PatchTypeJSONPatch
		response.PatchType = &patchType
		response.Patch = patch
		glog.V(4).Infof("Sending patches: %v", patches)
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
		glog.Errorf("contentType=%s, expect application/json", contentType)
		timer.Observe(metrics_admission.Error, metrics_admission.Unknown)
		return
	}

	reviewResponse, status, resource := s.admit(body)
	ar := v1beta1.AdmissionReview{
		Response: reviewResponse,
	}

	resp, err := json.Marshal(ar)
	if err != nil {
		glog.Error(err)
		timer.Observe(metrics_admission.Error, resource)
		return
	}

	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
		timer.Observe(metrics_admission.Error, resource)
		return
	}

	timer.Observe(status, resource)
}
