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

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/vpa"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	metrics_admission "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/admission"
	"k8s.io/klog"
)

// AdmissionServer is an admission webhook server that modifies pod resources request based on VPA recommendation
type AdmissionServer struct {
	limitsChecker    limitrange.LimitRangeCalculator
	resourceHandlers map[metav1.GroupResource]resource.Handler
}

// NewAdmissionServer constructs new AdmissionServer
func NewAdmissionServer(podPreProcessor pod.PreProcessor,
	vpaPreProcessor vpa.PreProcessor,
	limitsChecker limitrange.LimitRangeCalculator,
	vpaMatcher vpa.Matcher,
	patchCalculators []patch.Calculator) *AdmissionServer {
	as := &AdmissionServer{limitsChecker, map[metav1.GroupResource]resource.Handler{}}
	as.RegisterResourceHandler(pod.NewResourceHandler(podPreProcessor, vpaMatcher, patchCalculators))
	as.RegisterResourceHandler(vpa.NewResourceHandler(vpaPreProcessor))
	return as
}

// RegisterResourceHandler allows to register a custom logic for handling given types of resources.
func (s *AdmissionServer) RegisterResourceHandler(resourceHandler resource.Handler) {
	s.resourceHandlers[resourceHandler.GroupResource()] = resourceHandler
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

	var patches []resource.PatchRecord
	var err error
	resource := metrics_admission.Unknown

	admittedGroupResource := metav1.GroupResource{
		Group:    ar.Request.Resource.Group,
		Resource: ar.Request.Resource.Resource,
	}

	handler, ok := s.resourceHandlers[admittedGroupResource]
	if ok {
		patches, err = handler.GetPatches(ar.Request)
		resource = handler.AdmissionResource()

		if handler.DisallowIncorrectObjects() && err != nil {
			// we don't let in problematic objects - late validation
			status := metav1.Status{}
			status.Status = "Failure"
			status.Message = err.Error()
			response.Result = &status
			response.Allowed = false
		}
	} else {
		patches, err = nil, fmt.Errorf("not supported resource type: %v", admittedGroupResource)
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
