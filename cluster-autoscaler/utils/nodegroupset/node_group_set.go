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

package nodegroupset

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// FindSimilarNodeGroups returns a list of NodeGroups similar to the given one.
// Two groups are similar if the NodeInfos for them compare equal using IsNodeInfoSimilar.
func FindSimilarNodeGroups(nodeGroup cloudprovider.NodeGroup, cloudProvider cloudprovider.CloudProvider,
	nodeInfosForGroups map[string]*schedulercache.NodeInfo) ([]cloudprovider.NodeGroup, errors.AutoscalerError) {
	result := []cloudprovider.NodeGroup{}
	nodeGroupId := nodeGroup.Id()
	nodeInfo, found := nodeInfosForGroups[nodeGroupId]
	if !found {
		return []cloudprovider.NodeGroup{}, errors.NewAutoscalerError(
			errors.InternalError,
			"failed to find template node for node group %s",
			nodeGroupId)
	}
	for _, ng := range cloudProvider.NodeGroups() {
		ngId := ng.Id()
		if ngId == nodeGroupId {
			continue
		}
		ngNodeInfo, found := nodeInfosForGroups[ngId]
		if !found {
			glog.Warningf("Failed to find nodeInfo for group %v", ngId)
			continue
		}
		if IsNodeInfoSimilar(nodeInfo, ngNodeInfo) {
			result = append(result, ng)
		}
	}
	return result, nil
}
