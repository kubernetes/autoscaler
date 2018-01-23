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
)

type admissionServer struct {
	recommendationProvider logic.RecommendationProvider
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
	if ar.Request.Resource != podResource {
		glog.Errorf("expect resource to be %s", podResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	pod := v1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		glog.Error(err)
		return nil
	}
	if len(pod.Name) == 0 {
		pod.Name = pod.GenerateName + "%"
	}
	requests, err := s.recommendationProvider.GetRequestForPod(&pod)
	if err != nil {
		glog.Error(err)
		return nil
	}
	response := v1beta1.AdmissionResponse{}
	response.Allowed = true
	patches := []map[string]string{}
	for i, resources := range requests {
		for resource, request := range resources {
			patches = append(patches, map[string]string{
				"op":    "add",
				"path":  fmt.Sprintf("/spec/containers/%d/resources/requests/%s", i, resource),
				"value": request.String()})
		}
	}
	if len(patches) > 0 {
		patch, err := json.Marshal(patches)
		if err != nil {
			glog.Errorf("Cannot marshal the patch %v: %v", patches, err)
			return nil
		}
		patchType := v1beta1.PatchTypeJSONPatch
		response.PatchType = &patchType
		response.Patch = patch
		glog.V(4).Infof("Setting resource for pod %s: %v", pod.ObjectMeta.Name, patches)
	}
	glog.V(4).Infof("Replying with %v", response)
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
