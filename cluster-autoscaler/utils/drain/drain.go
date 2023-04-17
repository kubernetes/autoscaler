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
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
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

// GetPodsForDeletionOnNodeDrain returns pods that should be deleted on node drain as well as some extra information
// about possibly problematic pods (unreplicated and DaemonSets).
func GetPodsForDeletionOnNodeDrain(
	podList []*apiv1.Pod,
	pdbs []*policyv1.PodDisruptionBudget,
	skipNodesWithSystemPods bool,
	skipNodesWithLocalStorage bool,
	skipNodesWithCustomControllerPods bool,
	listers kube_util.ListerRegistry,
	minReplica int32,
	currentTime time.Time) (pods []*apiv1.Pod, daemonSetPods []*apiv1.Pod, blockingPod *BlockingPod, err error) {

	pods = []*apiv1.Pod{}
	daemonSetPods = []*apiv1.Pod{}
	// filter kube-system PDBs to avoid doing it for every kube-system pod
	kubeSystemPDBs := make([]*policyv1.PodDisruptionBudget, 0)
	for _, pdb := range pdbs {
		if pdb.Namespace == "kube-system" {
			kubeSystemPDBs = append(kubeSystemPDBs, pdb)
		}
	}

	for _, pod := range podList {
		if pod_util.IsMirrorPod(pod) {
			continue
		}

		// Possibly skip a pod under deletion but only if it was being deleted for long enough
		// to avoid a situation when we delete the empty node immediately after the pod was marked for
		// deletion without respecting any graceful termination.
		if IsPodLongTerminating(pod, currentTime) {
			// pod is being deleted for long enough - no need to care about it.
			continue
		}

		isDaemonSetPod := false
		replicated := false
		safeToEvict := hasSafeToEvictAnnotation(pod)
		terminal := isPodTerminal(pod)

		if skipNodesWithCustomControllerPods {
			// TODO(vadasambar): remove this when we get rid of skipNodesWithCustomControllerPods
			replicated, isDaemonSetPod, blockingPod, err = legacyCheckForReplicatedPods(listers, pod, minReplica)
			if err != nil {
				return []*apiv1.Pod{}, []*apiv1.Pod{}, blockingPod, err
			}
		} else {
			replicated = ControllerRef(pod) != nil
			isDaemonSetPod = pod_util.IsDaemonSetPod(pod)
		}

		if isDaemonSetPod {
			daemonSetPods = append(daemonSetPods, pod)
			continue
		}

		if !safeToEvict && !terminal {
			if !replicated {
				return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: NotReplicated}, fmt.Errorf("%s/%s is not replicated", pod.Namespace, pod.Name)
			}
			if pod.Namespace == "kube-system" && skipNodesWithSystemPods {
				hasPDB, err := checkKubeSystemPDBs(pod, kubeSystemPDBs)
				if err != nil {
					return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: UnexpectedError}, fmt.Errorf("error matching pods to pdbs: %v", err)
				}
				if !hasPDB {
					return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: UnmovableKubeSystemPod}, fmt.Errorf("non-daemonset, non-mirrored, non-pdb-assigned kube-system pod present: %s", pod.Name)
				}
			}
			if HasBlockingLocalStorage(pod) && skipNodesWithLocalStorage {
				return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: LocalStorageRequested}, fmt.Errorf("pod with local storage present: %s", pod.Name)
			}
			if hasNotSafeToEvictAnnotation(pod) {
				return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: NotSafeToEvictAnnotation}, fmt.Errorf("pod annotated as not safe to evict present: %s", pod.Name)
			}
		}
		pods = append(pods, pod)
	}
	return pods, daemonSetPods, nil, nil
}

func legacyCheckForReplicatedPods(listers kube_util.ListerRegistry, pod *apiv1.Pod, minReplica int32) (replicated bool, isDaemonSetPod bool, blockingPod *BlockingPod, err error) {
	replicated = false
	refKind := ""
	checkReferences := listers != nil
	isDaemonSetPod = false

	controllerRef := ControllerRef(pod)
	if controllerRef != nil {
		refKind = controllerRef.Kind
	}

	// For now, owner controller must be in the same namespace as the pod
	// so OwnerReference doesn't have its own Namespace field
	controllerNamespace := pod.Namespace
	if refKind == "ReplicationController" {
		if checkReferences {
			rc, err := listers.ReplicationControllerLister().ReplicationControllers(controllerNamespace).Get(controllerRef.Name)
			// Assume a reason for an error is because the RC is either
			// gone/missing or that the rc has too few replicas configured.
			// TODO: replace the minReplica check with pod disruption budget.
			if err == nil && rc != nil {
				if rc.Spec.Replicas != nil && *rc.Spec.Replicas < minReplica {
					return replicated, isDaemonSetPod, &BlockingPod{Pod: pod, Reason: MinReplicasReached}, fmt.Errorf("replication controller for %s/%s has too few replicas spec: %d min: %d",
						pod.Namespace, pod.Name, rc.Spec.Replicas, minReplica)
				}
				replicated = true
			} else {
				return replicated, isDaemonSetPod, &BlockingPod{Pod: pod, Reason: ControllerNotFound}, fmt.Errorf("replication controller for %s/%s is not available, err: %v", pod.Namespace, pod.Name, err)
			}
		} else {
			replicated = true
		}
	} else if pod_util.IsDaemonSetPod(pod) {
		isDaemonSetPod = true
		// don't have listener for other DaemonSet kind
		// TODO: we should use a generic client for checking the reference.
		if checkReferences && refKind == "DaemonSet" {
			_, err := listers.DaemonSetLister().DaemonSets(controllerNamespace).Get(controllerRef.Name)
			if apierrors.IsNotFound(err) {
				return replicated, isDaemonSetPod, &BlockingPod{Pod: pod, Reason: ControllerNotFound}, fmt.Errorf("daemonset for %s/%s is not present, err: %v", pod.Namespace, pod.Name, err)
			} else if err != nil {
				return replicated, isDaemonSetPod, &BlockingPod{Pod: pod, Reason: UnexpectedError}, fmt.Errorf("error when trying to get daemonset for %s/%s , err: %v", pod.Namespace, pod.Name, err)
			}
		}
	} else if refKind == "Job" {
		if checkReferences {
			job, err := listers.JobLister().Jobs(controllerNamespace).Get(controllerRef.Name)

			// Assume the only reason for an error is because the Job is
			// gone/missing, not for any other cause.  TODO(mml): something more
			// sophisticated than this
			if err == nil && job != nil {
				replicated = true
			} else {
				return replicated, isDaemonSetPod, &BlockingPod{Pod: pod, Reason: ControllerNotFound}, fmt.Errorf("job for %s/%s is not available: err: %v", pod.Namespace, pod.Name, err)
			}
		} else {
			replicated = true
		}
	} else if refKind == "ReplicaSet" {
		if checkReferences {
			rs, err := listers.ReplicaSetLister().ReplicaSets(controllerNamespace).Get(controllerRef.Name)

			// Assume the only reason for an error is because the RS is
			// gone/missing, not for any other cause.  TODO(mml): something more
			// sophisticated than this
			if err == nil && rs != nil {
				if rs.Spec.Replicas != nil && *rs.Spec.Replicas < minReplica {
					return replicated, isDaemonSetPod, &BlockingPod{Pod: pod, Reason: MinReplicasReached}, fmt.Errorf("replication controller for %s/%s has too few replicas spec: %d min: %d",
						pod.Namespace, pod.Name, rs.Spec.Replicas, minReplica)
				}
				replicated = true
			} else {
				return replicated, isDaemonSetPod, &BlockingPod{Pod: pod, Reason: ControllerNotFound}, fmt.Errorf("replication controller for %s/%s is not available, err: %v", pod.Namespace, pod.Name, err)
			}
		} else {
			replicated = true
		}
	} else if refKind == "StatefulSet" {
		if checkReferences {
			ss, err := listers.StatefulSetLister().StatefulSets(controllerNamespace).Get(controllerRef.Name)

			// Assume the only reason for an error is because the StatefulSet is
			// gone/missing, not for any other cause.  TODO(mml): something more
			// sophisticated than this
			if err == nil && ss != nil {
				replicated = true
			} else {
				return replicated, isDaemonSetPod, &BlockingPod{Pod: pod, Reason: ControllerNotFound}, fmt.Errorf("statefulset for %s/%s is not available: err: %v", pod.Namespace, pod.Name, err)
			}
		} else {
			replicated = true
		}
	}

	return replicated, isDaemonSetPod, &BlockingPod{}, nil
}

// ControllerRef returns the OwnerReference to pod's controller.
func ControllerRef(pod *apiv1.Pod) *metav1.OwnerReference {
	return metav1.GetControllerOf(pod)
}

// isPodTerminal checks whether the pod is in a terminal state.
func isPodTerminal(pod *apiv1.Pod) bool {
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

// This only checks if a matching PDB exist and therefore if it makes sense to attempt drain simulation,
// as we check for allowed-disruptions later anyway (for all pods with PDB, not just in kube-system)
func checkKubeSystemPDBs(pod *apiv1.Pod, pdbs []*policyv1.PodDisruptionBudget) (bool, error) {
	for _, pdb := range pdbs {
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			return false, err
		}
		if selector.Matches(labels.Set(pod.Labels)) {
			return true, nil
		}
	}

	return false, nil
}

// This checks if pod has PodSafeToEvictKey annotation
func hasSafeToEvictAnnotation(pod *apiv1.Pod) bool {
	return pod.GetAnnotations()[PodSafeToEvictKey] == "true"
}

// This checks if pod has PodSafeToEvictKey annotation set to false
func hasNotSafeToEvictAnnotation(pod *apiv1.Pod) bool {
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
