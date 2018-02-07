/*
Copyright 2017 The Kubernetes Authors.

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

package api

import (
	"encoding/json"
	"time"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/poc.autoscaling.k8s.io/v1alpha1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	"k8s.io/client-go/tools/cache"
)

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func patchVpa(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpaName string, patches []patchRecord) (result *vpa_types.VerticalPodAutoscaler, err error) {
	bytes, err := json.Marshal(patches)
	if err != nil {
		glog.Errorf("Cannot marshal VPA status patches %+v. Reason: %+v", patches, err)
		return
	}

	return vpaClient.Patch(vpaName, types.JSONPatchType, bytes)
}

// UpdateVpa updates VPA object status.
func UpdateVpaStatus(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpa *model.Vpa) (result *vpa_types.VerticalPodAutoscaler, err error) {
	status := vpa_types.VerticalPodAutoscalerStatus{
		LastUpdateTime:	metav1.Now(),
		Conditions:	vpa.Conditions.AsList(),
	}
	if vpa.Recommendation != nil {
		status.Recommendation = *vpa.Recommendation
	}

	patches := make([]patchRecord, 0)
	if !vpa.LastUpdateTime.IsZero() {
		// Verify that Status was not updated in the meantime.
		patches = append(patches, patchRecord{
			Op:    "test",
			Path:  "/status/lastUpdateTime",
			Value: vpa.LastUpdateTime,
		})
	}
	patches = append(patches, patchRecord{
		Op:    "add",
		Path:  "/status",
		Value: status,
	})

	
	return patchVpa(vpaClient, (*vpa).ID.VpaName, patches)
}

// NewAllVpasLister returns VerticalPodAutoscalerLister configured to fetch all VPA objects.
func NewAllVpasLister(vpaClient *vpa_clientset.Clientset, stopChannel <-chan struct{}) vpa_lister.VerticalPodAutoscalerLister {
	vpaListWatch := cache.NewListWatchFromClient(vpaClient.PocV1alpha1().RESTClient(), "verticalpodautoscalers", apiv1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	vpaLister := vpa_lister.NewVerticalPodAutoscalerLister(store)
	vpaReflector := cache.NewReflector(vpaListWatch, &vpa_types.VerticalPodAutoscaler{}, store, time.Hour)
	go vpaReflector.Run(stopChannel)
	return vpaLister
}

// PodMatchesVPA returns true iff the VPA's selector matches the Pod and they are in the same namespace.
func PodMatchesVPA(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) bool {
	if pod.Namespace != vpa.Namespace {
		return false
	}
	selector, err := metav1.LabelSelectorAsSelector(vpa.Spec.Selector)
	if err != nil {
		glog.Errorf("error processing VPA object: failed to create pod selector: %v", err)
		return false
	}
	return selector.Matches(labels.Set(pod.GetLabels()))
}

// stronger returns true iff a is before b in the order to control a Pod (that matches both VPAs).
func stronger(a, b *vpa_types.VerticalPodAutoscaler) bool {
	// Assume a is not nil and each valid object is before nil object.
	if b == nil {
		return true
	}
	// Compare creation timestamps of the VPA objects. This is the clue of the stronger logic.
	var aTime, bTime metav1.Time
	aTime = a.GetCreationTimestamp()
	bTime = b.GetCreationTimestamp()
	if !aTime.Equal(&bTime) {
		return aTime.Before(&bTime)
	}
	// If the timestamps are the same (unlikely, but possible e.g. in test environments): compare by name to have a complete deterministic order.
	return a.GetName() < b.GetName()
}

// GetControllingVPAForPod chooses the earliest created VPA from the input list that matches the given Pod.
func GetControllingVPAForPod(pod *apiv1.Pod, vpas []*vpa_types.VerticalPodAutoscaler) *vpa_types.VerticalPodAutoscaler {
	var controlling *vpa_types.VerticalPodAutoscaler
	// Choose the strongest VPA from the ones that match this Pod.
	for _, vpa := range vpas {
		if PodMatchesVPA(pod, vpa) && stronger(vpa, controlling) {
			controlling = vpa
		}
	}
	return controlling
}
