/*
Copyright 2016 The Kubernetes Authors.

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

package drain

import (
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// PodLongTerminatingExtraThreshold - time after which a pod, that is terminating and that has run over its terminationGracePeriod, should be ignored and considered as deleted
	PodLongTerminatingExtraThreshold = 30 * time.Second
)

const (
	// PodSafeToEvictKey - annotation that ignores constraints to evict a pod like not being replicated, being on
	// kube-system namespace or having a local storage.
	PodSafeToEvictKey = "cluster-autoscaler.kubernetes.io/safe-to-evict"
	// SafeToEvictLocalVolumesKey - annotation that ignores (doesn't block on) a local storage volume during node scale down
	SafeToEvictLocalVolumesKey = "cluster-autoscaler.kubernetes.io/safe-to-evict-local-volumes"
)

// BlockingPod represents a pod which is blocking the scale down of a node.
type BlockingPod struct {
	Pod    *apiv1.Pod
	Reason BlockingPodReason
}

// BlockingPodReason represents a reason why a pod is blocking the scale down of a node.
type BlockingPodReason int

const (
	// NoReason - sanity check, this should never be set explicitly. If this is found in the wild, it means that it was
	// implicitly initialized and might indicate a bug.
	NoReason BlockingPodReason = iota
	// ControllerNotFound - pod is blocking scale down because its controller can't be found.
	ControllerNotFound
	// MinReplicasReached - pod is blocking scale down because its controller already has the minimum number of replicas.
	MinReplicasReached
	// NotReplicated - pod is blocking scale down because it's not replicated.
	NotReplicated
	// LocalStorageRequested - pod is blocking scale down because it requests local storage.
	LocalStorageRequested
	// NotSafeToEvictAnnotation - pod is blocking scale down because it has a "not safe to evict" annotation.
	NotSafeToEvictAnnotation
	// UnmovableKubeSystemPod - pod is blocking scale down because it's a non-daemonset, non-mirrored, non-pdb-assigned kube-system pod.
	UnmovableKubeSystemPod
	// NotEnoughPdb - pod is blocking scale down because it doesn't have enough PDB left.
	NotEnoughPdb
	// UnexpectedError - pod is blocking scale down because of an unexpected error.
	UnexpectedError
)

// ControllerRef returns the OwnerReference to pod's controller.
func ControllerRef(pod *apiv1.Pod) *metav1.OwnerReference {
	return metav1.GetControllerOf(pod)
}

// IsPodTerminal checks whether the pod is in a terminal state.
func IsPodTerminal(pod *apiv1.Pod) bool {
	// pod will never be restarted
	if pod.Spec.RestartPolicy == apiv1.RestartPolicyNever && (pod.Status.Phase == apiv1.PodSucceeded || pod.Status.Phase == apiv1.PodFailed) {
		return true
	}
	// pod has run to completion and succeeded
	if pod.Spec.RestartPolicy == apiv1.RestartPolicyOnFailure && pod.Status.Phase == apiv1.PodSucceeded {
		return true
	}
	// kubelet has rejected this pod, due to eviction or some other constraint
	return pod.Status.Phase == apiv1.PodFailed
}

// HasBlockingLocalStorage returns true if pod has any local storage
// without pod annotation `<SafeToEvictLocalVolumeKey>: <volume-name-1>,<volume-name-2>...`
func HasBlockingLocalStorage(pod *apiv1.Pod) bool {
	isNonBlocking := getNonBlockingVolumes(pod)
	for _, volume := range pod.Spec.Volumes {
		if isLocalVolume(&volume) && !isNonBlocking[volume.Name] {
			return true
		}
	}
	return false
}

func getNonBlockingVolumes(pod *apiv1.Pod) map[string]bool {
	isNonBlocking := map[string]bool{}
	annotationVal := pod.GetAnnotations()[SafeToEvictLocalVolumesKey]
	if annotationVal != "" {
		vols := strings.Split(annotationVal, ",")
		for _, v := range vols {
			isNonBlocking[v] = true
		}
	}
	return isNonBlocking
}

func isLocalVolume(volume *apiv1.Volume) bool {
	return volume.HostPath != nil || (volume.EmptyDir != nil && volume.EmptyDir.Medium != apiv1.StorageMediumMemory)
}

// HasSafeToEvictAnnotation checks if pod has PodSafeToEvictKey annotation.
func HasSafeToEvictAnnotation(pod *apiv1.Pod) bool {
	return pod.GetAnnotations()[PodSafeToEvictKey] == "true"
}

// HasNotSafeToEvictAnnotation checks if pod has PodSafeToEvictKey annotation
// set to false.
func HasNotSafeToEvictAnnotation(pod *apiv1.Pod) bool {
	return pod.GetAnnotations()[PodSafeToEvictKey] == "false"
}

// IsPodLongTerminating checks if a pod has been terminating for a long time (pod's terminationGracePeriod + an additional const buffer)
func IsPodLongTerminating(pod *apiv1.Pod, currentTime time.Time) bool {
	// pod has not even been deleted
	if pod.DeletionTimestamp == nil {
		return false
	}

	gracePeriod := pod.Spec.TerminationGracePeriodSeconds
	if gracePeriod == nil {
		defaultGracePeriod := int64(apiv1.DefaultTerminationGracePeriodSeconds)
		gracePeriod = &defaultGracePeriod
	}
	return pod.DeletionTimestamp.Time.Add(time.Duration(*gracePeriod) * time.Second).Add(PodLongTerminatingExtraThreshold).Before(currentTime)
}
