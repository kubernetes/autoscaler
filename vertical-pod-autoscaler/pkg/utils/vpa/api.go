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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
)

// VpaWithSelector is a pair of VPA and its selector.
type VpaWithSelector struct {
	Vpa      *vpa_types.VerticalPodAutoscaler
	Selector labels.Selector
}

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func patchVpaStatus(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpaName string, patches []patchRecord) (result *vpa_types.VerticalPodAutoscaler, err error) {
	bytes, err := json.Marshal(patches)
	if err != nil {
		klog.Errorf("Cannot marshal VPA status patches %+v. Reason: %+v", patches, err)
		return
	}

	return vpaClient.Patch(context.TODO(), vpaName, types.JSONPatchType, bytes, meta.PatchOptions{}, "status")
}

// UpdateVpaStatusIfNeeded updates the status field of the VPA API object.
func UpdateVpaStatusIfNeeded(vpaClient vpa_api.VerticalPodAutoscalerInterface, vpaName string, newStatus,
	oldStatus *vpa_types.VerticalPodAutoscalerStatus) (result *vpa_types.VerticalPodAutoscaler, err error) {
	patches := []patchRecord{{
		Op:    "add",
		Path:  "/status",
		Value: *newStatus,
	}}

	if !apiequality.Semantic.DeepEqual(*oldStatus, *newStatus) {
		return patchVpaStatus(vpaClient, vpaName, patches)
	}
	return nil, nil
}

// NewVpasLister returns VerticalPodAutoscalerLister configured to fetch all VPA objects from namespace,
// set namespace to k8sapiv1.NamespaceAll to select all namespaces.
// The method blocks until vpaLister is initially populated.
func NewVpasLister(vpaClient *vpa_clientset.Clientset, stopChannel <-chan struct{}, namespace string) vpa_lister.VerticalPodAutoscalerLister {
	vpaListWatch := cache.NewListWatchFromClient(vpaClient.AutoscalingV1().RESTClient(), "verticalpodautoscalers", namespace, fields.Everything())
	indexer, controller := cache.NewIndexerInformer(vpaListWatch,
		&vpa_types.VerticalPodAutoscaler{},
		1*time.Hour,
		&cache.ResourceEventHandlerFuncs{},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	vpaLister := vpa_lister.NewVerticalPodAutoscalerLister(indexer)
	go controller.Run(stopChannel)
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		klog.Fatalf("Failed to sync VPA cache during initialization")
	} else {
		klog.Info("Initial VPA synced successfully")
	}
	return vpaLister
}

// NewVpaCheckpointLister returns VerticalPodAutoscalerCheckpointLister configured to fetch all VPACheckpoint objects from namespace,
// set namespace to k8sapiv1.NamespaceAll to select all namespaces.
// The method blocks until vpaCheckpointLister is initially populated.
func NewVpaCheckpointLister(vpaClient *vpa_clientset.Clientset, stopChannel <-chan struct{}, namespace string) vpa_lister.VerticalPodAutoscalerCheckpointLister {
	vpaListWatch := cache.NewListWatchFromClient(vpaClient.AutoscalingV1().RESTClient(), "verticalpodautoscalercheckpoints", namespace, fields.Everything())
	indexer, controller := cache.NewIndexerInformer(vpaListWatch,
		&vpa_types.VerticalPodAutoscalerCheckpoint{},
		1*time.Hour,
		&cache.ResourceEventHandlerFuncs{},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	vpaCheckpointLister := vpa_lister.NewVerticalPodAutoscalerCheckpointLister(indexer)
	go controller.Run(stopChannel)
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		klog.Fatalf("Failed to sync VPA checkpoint cache during initialization")
	} else {
		klog.Info("Initial VPA checkpoint synced successfully")
	}
	return vpaCheckpointLister
}

// PodMatchesVPA returns true iff the vpaWithSelector matches the Pod.
func PodMatchesVPA(pod *core.Pod, vpaWithSelector *VpaWithSelector) bool {
	return PodLabelsMatchVPA(pod.Namespace, labels.Set(pod.GetLabels()), vpaWithSelector.Vpa.Namespace, vpaWithSelector.Selector)
}

// PodLabelsMatchVPA returns true iff the vpaWithSelector matches the pod labels.
func PodLabelsMatchVPA(podNamespace string, labels labels.Set, vpaNamespace string, vpaSelector labels.Selector) bool {
	if podNamespace != vpaNamespace {
		return false
	}
	return vpaSelector.Matches(labels)
}

// stronger returns true iff a is before b in the order to control a Pod (that matches both VPAs).
func stronger(a, b *vpa_types.VerticalPodAutoscaler) bool {
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
func GetControllingVPAForPod(ctx context.Context, pod *core.Pod, vpas []*VpaWithSelector, ctrlFetcher controllerfetcher.ControllerFetcher) *VpaWithSelector {

	var ownerRefrence *meta.OwnerReference
	for i := range pod.OwnerReferences {
		r := pod.OwnerReferences[i]
		if r.Controller != nil && *r.Controller {
			ownerRefrence = &r
		}
	}
	if ownerRefrence == nil {
		// If the pod has no ownerReference, it cannot be under a VPA.
		return nil
	}
	k := &controllerfetcher.ControllerKeyWithAPIVersion{
		ControllerKey: controllerfetcher.ControllerKey{
			Namespace: pod.Namespace,
			Kind:      ownerRefrence.Kind,
			Name:      ownerRefrence.Name,
		},
		ApiVersion: ownerRefrence.APIVersion,
	}
	parentController, err := ctrlFetcher.FindTopMostWellKnownOrScalable(ctx, k)
	if err != nil {
		klog.Errorf("fail to get pod controller: pod=%s err=%s", klog.KObj(pod), err.Error())
		return nil
	}
	if parentController == nil {
		return nil
	}

	var controlling *VpaWithSelector
	var controllingVpa *vpa_types.VerticalPodAutoscaler
	// Choose the strongest VPA from the ones that match this Pod.
	for _, vpaWithSelector := range vpas {
		if vpaWithSelector.Vpa.Spec.TargetRef == nil {
			continue
		}
		if vpaWithSelector.Vpa.Spec.TargetRef.Kind != parentController.Kind ||
			vpaWithSelector.Vpa.Namespace != parentController.Namespace ||
			vpaWithSelector.Vpa.Spec.TargetRef.Name != parentController.Name {
			continue // This pod is not associated to the right controller
		}
		if PodMatchesVPA(pod, vpaWithSelector) && stronger(vpaWithSelector.Vpa, controllingVpa) {
			controlling = vpaWithSelector
			controllingVpa = controlling.Vpa
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

// GetContainerControlledValues returns controlled resource values
func GetContainerControlledValues(name string, vpaResourcePolicy *vpa_types.PodResourcePolicy) vpa_types.ContainerControlledValues {
	containerPolicy := GetContainerResourcePolicy(name, vpaResourcePolicy)
	if containerPolicy == nil || containerPolicy.ControlledValues == nil {
		return vpa_types.ContainerControlledValuesRequestsAndLimits
	}
	return *containerPolicy.ControlledValues
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
	_, err = vpaCheckpointClient.Patch(context.TODO(), vpaCheckpoint.ObjectMeta.Name, types.JSONPatchType, bytes, meta.PatchOptions{})
	if err != nil && strings.Contains(err.Error(), fmt.Sprintf("\"%s\" not found", vpaCheckpoint.ObjectMeta.Name)) {
		_, err = vpaCheckpointClient.Create(context.TODO(), vpaCheckpoint, meta.CreateOptions{})
	}
	if err != nil {
		return fmt.Errorf("Cannot save checkpoint for vpa %s/%s container %s. Reason: %+v", vpaCheckpoint.Namespace, vpaCheckpoint.Name, vpaCheckpoint.Spec.ContainerName, err)
	}
	return nil
}
