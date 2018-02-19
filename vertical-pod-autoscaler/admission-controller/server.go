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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/admission-controller/logic"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
)

type admissionServer struct {
	recommendationProvider logic.RecommendationProvider
}

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func (s *admissionServer) getPatchesForPodResourceRequest(raw []byte, namespace string) ([]patchRecord, error) {
	pod := v1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		return nil, err
	}
	if len(pod.Name) == 0 {
		pod.Name = pod.GenerateName + "%"
		pod.Namespace = namespace
	}
	glog.V(4).Infof("Admitting pod %v", pod.ObjectMeta)
	requests, err := s.recommendationProvider.GetRequestForPod(&pod)
	if err != nil {
		return nil, err
	}
	patches := []patchRecord{}
	for i, resources := range requests {
		for resource, request := range resources {
			patches = append(patches, patchRecord{
				Op:    "add",
				Path:  fmt.Sprintf("/spec/containers/%d/resources/requests/%s", i, resource),
				Value: request.String()})
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
	if vpa.Spec.UpdatePolicy.UpdateMode == "" {
		// Sets the whole updatePolicy. Note that if it contained mutiple fields
		// this needed to be changed to respect their original values.
		patches = append(patches, patchRecord{
			Op:    "add",
			Path:  "/spec/updatePolicy",
			Value: vpa_types.PodUpdatePolicy{UpdateMode: vpa_types.UpdateModeAuto}})
	}
	return patches, nil
}

// only allow pods to pull images from specific registry.
func (s *admissionServer) admit(data []byte) *v1beta1.AdmissionResponse {
	ar := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(data, &ar); err != nil {
		glog.Error(err)
		return nil
	}
	// The externalAdmissionHookConfiguration registered via selfRegistration
	// asks the kube-apiserver only sends admission request regarding pods.
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	vpaResource := metav1.GroupVersionResource{Group: "poc.autoscaling.k8s.io", Version: "v1alpha1", Resource: "verticalpodautoscalers"}
	var patches []patchRecord
	var err error

	switch ar.Request.Resource {
	case podResource:
		patches, err = s.getPatchesForPodResourceRequest(ar.Request.Object.Raw, ar.Request.Namespace)
	case vpaResource:
		patches, err = getPatchesForVPADefaults(ar.Request.Object.Raw)
	default:
		patches, err = nil, fmt.Errorf("expected the resource to be %v or %v", podResource, vpaResource)
	}

	if err != nil {
		glog.Error(err)
		return nil
	}
	response := v1beta1.AdmissionResponse{}
	response.Allowed = true
	if len(patches) > 0 {
		patch, err := json.Marshal(patches)
		if err != nil {
			glog.Errorf("Cannot marshal the patch %v: %v", patches, err)
			return nil
		}
		patchType := v1beta1.PatchTypeJSONPatch
		response.PatchType = &patchType
		response.Patch = patch
		glog.V(4).Infof("Sending patches: %v", patches)
	}
	return &response
}

func (s *admissionServer) serve(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	reviewResponse := s.admit(body)
	ar := v1beta1.AdmissionReview{
		Response: reviewResponse,
	}

	resp, err := json.Marshal(ar)
	if err != nil {
		glog.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
	}
}
