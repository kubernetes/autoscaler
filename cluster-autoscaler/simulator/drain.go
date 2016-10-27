/*
Copyright 2015 The Kubernetes Authors.

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
	"k8s.io/contrib/cluster-autoscaler/utils/drain"

	"k8s.io/kubernetes/pkg/api"
	unversionedclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// FastGetPodsToMove returns a list of pods that should be moved elsewhere if the node
// is drained. Raises error if there is an unreplicated pod and force option was not specified.
// Based on kubectl drain code. It makes an assumption that RC, DS, Jobs and RS were deleted
// along with their pods (no abandoned pods with dangling created-by annotation). Usefull for fast
// checks. Doesn't check i
func FastGetPodsToMove(nodeInfo *schedulercache.NodeInfo, skipNodesWithSystemPods bool, skipNodesWithLocalStorage bool) ([]*api.Pod, error) {
	return drain.GetPodsForDeletionOnNodeDrain(
		nodeInfo.Pods(),
		api.Codecs.UniversalDecoder(),
		skipNodesWithSystemPods,
		skipNodesWithLocalStorage,
		false,
		nil,
		0)
}

// DetailedGetPodsForMove returns a list of pods that should be moved elsewhere if the node
// is drained. Raises error if there is an unreplicated pod and force option was not specified.
// Based on kubectl drain code. It checks whether RC, DS, Jobs and RS that created these pods
// still exist.
func DetailedGetPodsForMove(nodeInfo *schedulercache.NodeInfo, skipNodesWithSystemPods bool,
	skipNodesWithLocalStorage bool, client *unversionedclient.Client, minReplicaCount int32) ([]*api.Pod, error) {
	return drain.GetPodsForDeletionOnNodeDrain(
		nodeInfo.Pods(),
		api.Codecs.UniversalDecoder(),
		skipNodesWithSystemPods,
		skipNodesWithLocalStorage,
		true,
		client,
		minReplicaCount)
}
