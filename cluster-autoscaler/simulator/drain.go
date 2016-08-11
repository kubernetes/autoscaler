/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package simulator

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// FastGetPodsToMove returns a list of pods that should be moved elsewhere if the node
// is drained. Raises error if there is an unreplicated pod and force option was not specified.
// Based on kubectl drain code. It makes an assumption that RC, DS, Jobs and RS were deleted
// along with their pods (no abandoned pods with dangling created-by annotation). Usefull for fast
// checks.
func FastGetPodsToMove(nodeInfo *schedulercache.NodeInfo, force bool,
	skipNodesWithSystemPods bool, skipNodesWithLocalStorage bool) ([]*api.Pod, error) {
	pods := make([]*api.Pod, 0)
	unreplicatedPodNames := []string{}
	for _, pod := range nodeInfo.Pods() {
		if IsMirrorPod(pod) {
			continue
		}

		replicated := false
		daemonsetPod := false

		creatorKind, err := CreatorRefKind(pod)
		if err != nil {
			return []*api.Pod{}, err
		}
		if creatorKind == "ReplicationController" {
			replicated = true
		} else if creatorKind == "DaemonSet" {
			daemonsetPod = true
		} else if creatorKind == "Job" {
			replicated = true
		} else if creatorKind == "ReplicaSet" {
			replicated = true
		}

		if !daemonsetPod && pod.Namespace == "kube-system" && skipNodesWithSystemPods {
			return []*api.Pod{}, fmt.Errorf("non-deamons set, non-mirrored, kube-system pod present: %s", pod.Name)
		}

		if !daemonsetPod && hasLocalStorage(pod) && skipNodesWithLocalStorage {
			return []*api.Pod{}, fmt.Errorf("pod with local storage present: %s", pod.Name)
		}

		switch {
		case daemonsetPod:
			break
		case !replicated:
			unreplicatedPodNames = append(unreplicatedPodNames, pod.Name)
			if force {
				pods = append(pods, pod)
			}
		default:
			pods = append(pods, pod)
		}
	}
	if !force && len(unreplicatedPodNames) > 0 {
		return []*api.Pod{}, fmt.Errorf("unreplicated pods present")
	}
	return pods, nil
}

// CreatorRefKind returns the kind of the creator of the pod.
func CreatorRefKind(pod *api.Pod) (string, error) {
	creatorRef, found := pod.ObjectMeta.Annotations[controller.CreatedByAnnotation]
	if !found {
		return "", nil
	}
	var sr api.SerializedReference
	if err := runtime.DecodeInto(api.Codecs.UniversalDecoder(), []byte(creatorRef), &sr); err != nil {
		return "", err
	}
	return sr.Reference.Kind, nil
}

// IsMirrorPod checks whether the pod is a mirror pod.
func IsMirrorPod(pod *api.Pod) bool {
	_, found := pod.ObjectMeta.Annotations[types.ConfigMirrorAnnotationKey]
	return found
}

func hasLocalStorage(pod *api.Pod) bool {
	for _, volume := range pod.Spec.Volumes {
		if isLocalVolume(&volume) {
			return true
		}
	}
	return false
}

func isLocalVolume(volume *api.Volume) bool {
	return volume.HostPath != nil || volume.EmptyDir != nil
}
