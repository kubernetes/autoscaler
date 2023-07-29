/*
Copyright 2022 Haoran Qiu.

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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	mpa_clientset "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned"
	mpa_api "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1alpha1"
	mpa_lister "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1alpha1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// MpaWithSelector is a pair of MPA and its selector.
type MpaWithSelector struct {
	Mpa      *mpa_types.MultidimPodAutoscaler
	Selector labels.Selector
}

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func patchMpa(mpaClient mpa_api.MultidimPodAutoscalerInterface, mpaName string, patches []patchRecord) (result *mpa_types.MultidimPodAutoscaler, err error) {
	bytes, err := json.Marshal(patches)
	if err != nil {
		klog.Errorf("Cannot marshal MPA status patches %+v. Reason: %+v", patches, err)
		return
	}

	return mpaClient.Patch(context.TODO(), mpaName, types.JSONPatchType, bytes, meta.PatchOptions{})
}

// NewMpasLister returns MultidimPodAutoscalerLister configured to fetch all MPA objects from
// namespace, set namespace to k8sapiv1.NamespaceAll to select all namespaces.
// The method blocks until mpaLister is initially populated.
func NewMpasLister(mpaClient *mpa_clientset.Clientset, stopChannel <-chan struct{}, namespace string) mpa_lister.MultidimPodAutoscalerLister {
	mpaListWatch := cache.NewListWatchFromClient(mpaClient.AutoscalingV1alpha1().RESTClient(), "multidimpodautoscalers", namespace, fields.Everything())
	indexer, controller := cache.NewIndexerInformer(mpaListWatch,
		&mpa_types.MultidimPodAutoscaler{},
		1*time.Hour,
		&cache.ResourceEventHandlerFuncs{},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	mpaLister := mpa_lister.NewMultidimPodAutoscalerLister(indexer)
	go controller.Run(stopChannel)
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		klog.Fatalf("Failed to sync MPA cache during initialization")
	} else {
		klog.Info("Initial MPA synced successfully")
	}
	return mpaLister
}

// CreateOrUpdateMpaCheckpoint updates the status field of the MPA Checkpoint API object.
// If object doesn't exits it is created.
func CreateOrUpdateMpaCheckpoint(mpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointInterface,
	mpaCheckpoint *vpa_types.VerticalPodAutoscalerCheckpoint) error {
	patches := make([]patchRecord, 0)
	patches = append(patches, patchRecord{
		Op:    "replace",
		Path:  "/status",
		Value: mpaCheckpoint.Status,
	})
	bytes, err := json.Marshal(patches)
	if err != nil {
		return fmt.Errorf("Cannot marshal MPA checkpoint status patches %+v. Reason: %+v", patches, err)
	}
	_, err = mpaCheckpointClient.Patch(context.TODO(), mpaCheckpoint.ObjectMeta.Name, types.JSONPatchType, bytes, meta.PatchOptions{})
	if err != nil && strings.Contains(err.Error(), fmt.Sprintf("\"%s\" not found", mpaCheckpoint.ObjectMeta.Name)) {
		_, err = mpaCheckpointClient.Create(context.TODO(), mpaCheckpoint, meta.CreateOptions{})
	}
	if err != nil {
		return fmt.Errorf("Cannot save checkpoint for mpa %v container %v. Reason: %+v", mpaCheckpoint.ObjectMeta.Name, mpaCheckpoint.Spec.ContainerName, err)
	}
	return nil
}

// UpdateMpaStatusIfNeeded updates the status field of the MPA API object.
func UpdateMpaStatusIfNeeded(mpaClient mpa_api.MultidimPodAutoscalerInterface, mpaName string, newStatus,
	oldStatus *mpa_types.MultidimPodAutoscalerStatus) (result *mpa_types.MultidimPodAutoscaler, err error) {
	patches := []patchRecord{{
		Op:    "add",
		Path:  "/status",
		Value: *newStatus,
	}}

	if !apiequality.Semantic.DeepEqual(*oldStatus, *newStatus) {
		return patchMpa(mpaClient, mpaName, patches)
	}
	return nil, nil
}

// GetUpdateMode returns the updatePolicy.updateMode for a given MPA.
// If the mode is not specified it returns the default (UpdateModeAuto).
func GetUpdateMode(mpa *mpa_types.MultidimPodAutoscaler) vpa_types.UpdateMode {
	if mpa.Spec.UpdatePolicy == nil || mpa.Spec.UpdatePolicy.UpdateMode == nil || *mpa.Spec.UpdatePolicy.UpdateMode == "" {
		return vpa_types.UpdateModeAuto
	}
	return *mpa.Spec.UpdatePolicy.UpdateMode
}

// PodMatchesMPA returns true iff the mpaWithSelector matches the Pod.
func PodMatchesMPA(pod *core.Pod, mpaWithSelector *MpaWithSelector) bool {
	return PodLabelsMatchMPA(pod.Namespace, labels.Set(pod.GetLabels()), mpaWithSelector.Mpa.Namespace, mpaWithSelector.Selector)
}

// PodLabelsMatchMPA returns true iff the mpaWithSelector matches the pod labels.
func PodLabelsMatchMPA(podNamespace string, labels labels.Set, mpaNamespace string, mpaSelector labels.Selector) bool {
	if podNamespace != mpaNamespace {
		return false
	}
	return mpaSelector.Matches(labels)
}

// stronger returns true iff a is before b in the order to control a Pod (that matches both MPAs).
func stronger(a, b *mpa_types.MultidimPodAutoscaler) bool {
	// Assume a is not nil and each valid object is before nil object.
	if b == nil {
		return true
	}
	// Compare creation timestamps of the MPA objects. This is the clue of the stronger logic.
	var aTime, bTime meta.Time
	aTime = a.GetCreationTimestamp()
	bTime = b.GetCreationTimestamp()
	if !aTime.Equal(&bTime) {
		return aTime.Before(&bTime)
	}
	// If the timestamps are the same (unlikely, but possible e.g. in test environments): compare by name to have a complete deterministic order.
	return a.GetName() < b.GetName()
}

// GetControllingMPAForPod chooses the earliest created MPA from the input list that matches the given Pod.
func GetControllingMPAForPod(pod *core.Pod, mpas []*MpaWithSelector) *MpaWithSelector {
	var controlling *MpaWithSelector
	var controllingMpa *mpa_types.MultidimPodAutoscaler
	// Choose the strongest MPA from the ones that match this Pod.
	for _, mpaWithSelector := range mpas {
		if PodMatchesMPA(pod, mpaWithSelector) && stronger(mpaWithSelector.Mpa, controllingMpa) {
			controlling = mpaWithSelector
			controllingMpa = controlling.Mpa
		}
	}
	return controlling
}
