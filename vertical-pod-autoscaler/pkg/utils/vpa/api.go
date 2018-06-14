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
	"fmt"
	"strings"
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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
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

// UpdateVpaStatus updates the status field of the VPA API object.
// It prevents race conditions by verifying that the lastUpdateTime of the
// API object and its model representation are equal.
func UpdateVpaStatus(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpa *model.Vpa) (result *vpa_types.VerticalPodAutoscaler, err error) {
	status := vpa_types.VerticalPodAutoscalerStatus{
		Conditions: vpa.Conditions.AsList(),
	}
	if vpa.Recommendation != nil {
		status.Recommendation = vpa.Recommendation
	}
	patches := []patchRecord{{
		Op:    "add",
		Path:  "/status",
		Value: status,
	}}
	return patchVpa(vpaClient, (*vpa).ID.VpaName, patches)
}

// NewAllVpasLister returns VerticalPodAutoscalerLister configured to fetch all VPA objects.
// The method blocks until vpaLister is initially populated.
func NewAllVpasLister(vpaClient *vpa_clientset.Clientset, stopChannel <-chan struct{}) vpa_lister.VerticalPodAutoscalerLister {
	vpaListWatch := cache.NewListWatchFromClient(vpaClient.PocV1alpha1().RESTClient(), "verticalpodautoscalers", apiv1.NamespaceAll, fields.Everything())
	indexer, controller := cache.NewIndexerInformer(vpaListWatch,
		&vpa_types.VerticalPodAutoscaler{},
		1*time.Hour,
		&cache.ResourceEventHandlerFuncs{},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	vpaLister := vpa_lister.NewVerticalPodAutoscalerLister(indexer)
	go controller.Run(stopChannel)
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		glog.Fatalf("Failed to sync VPA cache during initialization")
	} else {
		glog.Info("Initial VPA synced successfully")
	}
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

// GetUpdateMode returns the updatePolicy.updateMode for a given VPA.
// If the mode is not specified it returns the default (UpdateModeAuto).
func GetUpdateMode(vpa *vpa_types.VerticalPodAutoscaler) vpa_types.UpdateMode {
	if vpa.Spec.UpdatePolicy == nil || vpa.Spec.UpdatePolicy.UpdateMode == nil || *vpa.Spec.UpdatePolicy.UpdateMode == "" {
		return vpa_types.UpdateModeAuto
	}
	return *vpa.Spec.UpdatePolicy.UpdateMode
}

// GetContainerResourcePolicy returns the ContainerResourcePolicy for a given policy
// and container name. It returns nil if there is no policy specified for the container.
func GetContainerResourcePolicy(containerName string, policy *vpa_types.PodResourcePolicy) *vpa_types.ContainerResourcePolicy {
	var defaultPolicy *vpa_types.ContainerResourcePolicy
	if policy != nil {
		for i, containerPolicy := range policy.ContainerPolicies {
			if containerPolicy.ContainerName == containerName {
				return &policy.ContainerPolicies[i]
			}
			if containerPolicy.ContainerName == vpa_types.DefaultContainerResourcePolicy {
				defaultPolicy = &policy.ContainerPolicies[i]
			}
		}
	}
	return defaultPolicy
}

// CreateOrUpdateVpaCheckpoint updates the status field of the VPA Checkpoint API object.
// If object doesn't exits it is created.
func CreateOrUpdateVpaCheckpoint(vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointInterface,
	vpaCheckpoint *vpa_types.VerticalPodAutoscalerCheckpoint) error {
	patches := make([]patchRecord, 0)
	patches = append(patches, patchRecord{
		Op:    "replace",
		Path:  "/status",
		Value: vpaCheckpoint.Status,
	})
	bytes, err := json.Marshal(patches)
	if err != nil {
		return fmt.Errorf("Cannot marshal VPA checkpoint status patches %+v. Reason: %+v", patches, err)
	}
	_, err = vpaCheckpointClient.Patch(vpaCheckpoint.ObjectMeta.Name, types.JSONPatchType, bytes)
	if err != nil && strings.Contains(err.Error(), fmt.Sprintf("\"%s\" not found", vpaCheckpoint.ObjectMeta.Name)) {
		_, err = vpaCheckpointClient.Create(vpaCheckpoint)
	}
	if err != nil {
		return fmt.Errorf("Cannot save checkpotint for vpa %v container %v. Reason: %+v", vpaCheckpoint.ObjectMeta.Name, vpaCheckpoint.Spec.ContainerName, err)
	}
	return nil
}
