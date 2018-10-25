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
	core "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1beta1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1beta1"
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

// UpdateVpaStatusIfNeeded updates the status field of the VPA API object.
// It prevents race conditions by verifying that the lastUpdateTime of the
// API object and its model representation are equal.
func UpdateVpaStatusIfNeeded(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpa *model.Vpa,
	oldStatus *vpa_types.VerticalPodAutoscalerStatus) (result *vpa_types.VerticalPodAutoscaler, err error) {
	newStatus := &vpa_types.VerticalPodAutoscalerStatus{
		Conditions: vpa.Conditions.AsList(),
	}
	if vpa.Recommendation != nil {
		newStatus.Recommendation = vpa.Recommendation
	}
	patches := []patchRecord{{
		Op:    "add",
		Path:  "/status",
		Value: *newStatus,
	}}

	if !apiequality.Semantic.DeepEqual(*oldStatus, *newStatus) {
		return patchVpa(vpaClient, (*vpa).ID.VpaName, patches)
	}
	return nil, nil
}

// NewAllVpasLister returns VerticalPodAutoscalerLister configured to fetch all VPA objects.
// The method blocks until vpaLister is initially populated.
func NewAllVpasLister(vpaClient vpa_clientset.Interface, stopChannel <-chan struct{}) vpa_lister.VerticalPodAutoscalerLister {
	vpaListWatch := cache.NewListWatchFromClient(vpaClient.AutoscalingV1beta1().RESTClient(), "verticalpodautoscalers", core.NamespaceAll, fields.Everything())
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

// NewAllCPsLister returns ClusterProportionalScalerLister configured to fetch all CPS objects.
// The method blocks until the lister is initially populated.
func NewAllCPSLister(cpsClient vpa_clientset.Interface, stopChannel <-chan struct{}) vpa_lister.ClusterProportionalScalerLister {
	cpsListWatch := cache.NewListWatchFromClient(cpsClient.AutoscalingV1beta1().RESTClient(), "clusterproportionalscalers", core.NamespaceAll, fields.Everything())
	indexer, controller := cache.NewIndexerInformer(cpsListWatch,
		&vpa_types.ClusterProportionalScaler{},
		1*time.Hour,
		&cache.ResourceEventHandlerFuncs{},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	cpsLister := vpa_lister.NewClusterProportionalScalerLister(indexer)
	go controller.Run(stopChannel)
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		glog.Fatalf("Failed to sync VPA cache during initialization")
	} else {
		glog.Info("Initial VPA synced successfully")
	}
	return cpsLister
}

// PodMatchesVPA returns true iff the VPA's selector matches the Pod and they are in the same namespace.
func PodMatchesVPA(pod *core.Pod, vpa ScalerDuck) bool {
	if pod.Namespace != vpa.GetNamespace() {
		return false
	}
	selector, err := meta.LabelSelectorAsSelector(vpa.GetSelector())
	if err != nil {
		glog.Errorf("error processing VPA object: failed to create pod selector: %v", err)
		return false
	}

	return selector.Matches(labels.Set(pod.GetLabels()))
}

// stronger returns true iff a is before b in the order to control a Pod (that matches both VPAs).
func stronger(a, b ScalerDuck) bool {
	// Assume a is not nil and each valid object is before nil object.
	if b == nil {
		return true
	}
	// Compare creation timestamps of the VPA objects. This is the clue of the stronger logic.
	var aTime, bTime meta.Time
	aTime = a.GetCreationTimestamp()
	bTime = b.GetCreationTimestamp()
	if !aTime.Equal(&bTime) {
		return aTime.Before(&bTime)
	}
	// If the timestamps are the same (unlikely, but possible e.g. in test environments): compare by name to have a complete deterministic order.
	return a.GetName() < b.GetName()
}

// GetControllingVPAForPod chooses the earliest created VPA from the input list that matches the given Pod.
func GetControllingVPAForPod(pod *core.Pod, vpas []ScalerDuck) ScalerDuck {
	var controlling ScalerDuck
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

// GetCPSUpdateMode returns the updatePolicy.updateMode for a given CPS.
// If the mode is not specified it returns the default (UpdateModeAuto).
func GetCPSUpdateMode(cps *vpa_types.ClusterProportionalScaler) vpa_types.UpdateMode {
	if cps.Spec.UpdatePolicy == nil || cps.Spec.UpdatePolicy.UpdateMode == nil || *cps.Spec.UpdatePolicy.UpdateMode == "" {
		return vpa_types.UpdateModeAuto
	}
	return *cps.Spec.UpdatePolicy.UpdateMode
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
		return fmt.Errorf("Cannot save checkpoint for vpa %v container %v. Reason: %+v", vpaCheckpoint.ObjectMeta.Name, vpaCheckpoint.Spec.ContainerName, err)
	}
	return nil
}

// GetContainerResourcePolicy returns the ContainerResourcePolicy for a given policy
// and container name. It returns nil if there is no policy specified for the container.
func GetContainerResourcePolicy(containerName string, policy *vpa_types.PodResourcePolicy) *vpa_types.ContainerResourcePolicy {
	var defaultPolicy *vpa_types.ContainerResourcePolicy
	if policy != nil {
		containerPolicies := policy.ContainerPolicies
		for i := range containerPolicies {
			if containerPolicies[i].ContainerName == containerName {
				return &containerPolicies[i]
			}
			if containerPolicies[i].ContainerName == vpa_types.DefaultContainerResourcePolicy {
				defaultPolicy = &containerPolicies[i]
			}
		}
	}
	return defaultPolicy
}

// GetCPSContainerResourcePolicy returns the ContainerResourcePolicy for a given policy
// and container name. It returns nil if there is no policy specified for the container.
func GetCPSContainerResourcePolicy(containerName string, policy *vpa_types.CPPodResourcePolicy) *vpa_types.ContainerResourcePolicy {
	var defaultPolicy *vpa_types.ContainerResourcePolicy
	if policy != nil {
		containerPolicies := policy.ContainerPolicies
		for i := range containerPolicies {
			if containerPolicies[i].ContainerName == containerName {
				return &containerPolicies[i].ContainerResourcePolicy
			}
			if containerPolicies[i].ContainerName == vpa_types.DefaultContainerResourcePolicy {
				defaultPolicy = &containerPolicies[i].ContainerResourcePolicy
			}
		}
	}
	return defaultPolicy
}
