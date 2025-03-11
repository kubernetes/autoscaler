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

package orchestrator

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type asyncNodeGroupInitializer struct {
	nodeGroup              cloudprovider.NodeGroup
	nodeInfo               *framework.NodeInfo
	scaleUpExecutor        *scaleUpExecutor
	taintConfig            taints.TaintConfig
	daemonSets             []*appsv1.DaemonSet
	scaleUpStatusProcessor status.ScaleUpStatusProcessor
	context                *context.AutoscalingContext
	atomicScaleUp          bool
}

func newAsyncNodeGroupInitializer(
	nodeGroup cloudprovider.NodeGroup,
	nodeInfo *framework.NodeInfo,
	scaleUpExecutor *scaleUpExecutor,
	taintConfig taints.TaintConfig,
	daemonSets []*appsv1.DaemonSet,
	scaleUpStatusProcessor status.ScaleUpStatusProcessor,
	context *context.AutoscalingContext,
	atomicScaleUp bool,
) *asyncNodeGroupInitializer {
	return &asyncNodeGroupInitializer{
		nodeGroup,
		nodeInfo,
		scaleUpExecutor,
		taintConfig,
		daemonSets,
		scaleUpStatusProcessor,
		context,
		atomicScaleUp,
	}
}

func (s *asyncNodeGroupInitializer) InitializeNodeGroup(result nodegroups.AsyncNodeGroupCreationResult) {
	if result.Error != nil {
		klog.Errorf("Async node group creation failed. Async scale-up is cancelled. %v", result.Error)
		scaleUpStatus, _ := status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.ToAutoscalerError(errors.InternalError, result.Error))
		s.scaleUpStatusProcessor.Process(s.context, scaleUpStatus)
		return
	}
	mainCreatedNodeGroup := result.CreationResult.MainCreatedNodeGroup
	// If possible replace candidate node-info with node info based on crated node group. The latter
	// one should be more in line with nodes which will be created by node group.
	nodeInfo, aErr := utils.GetNodeInfoFromTemplate(mainCreatedNodeGroup, s.daemonSets, s.taintConfig)
	if aErr != nil {
		klog.Warningf("Cannot build node info for newly created main node group %s. Using fallback. Error: %v", mainCreatedNodeGroup.Id(), aErr)
		nodeInfo = s.nodeInfo
	}

	nodeInfos := make(map[string]*framework.NodeInfo)
	var scaleUpInfos []nodegroupset.ScaleUpInfo
	for _, nodeGroup := range result.CreationResult.AllCreatedNodeGroups() {
		if targetSize := result.TargetSizes[nodeGroup.Id()]; targetSize > 0 {
			nodeInfos[nodeGroup.Id()] = nodeInfo
			scaleUpInfo := nodegroupset.ScaleUpInfo{
				Group:       nodeGroup,
				CurrentSize: 0,
				NewSize:     targetSize,
				MaxSize:     nodeGroup.MaxSize(),
			}
			scaleUpInfos = append(scaleUpInfos, scaleUpInfo)
		}
	}
	klog.Infof("Starting initial scale-up for async created node groups. Scale ups: %v", scaleUpInfos)
	err, failedNodeGroups := s.scaleUpExecutor.ExecuteScaleUps(scaleUpInfos, nodeInfos, time.Now(), s.atomicScaleUp)
	if err != nil {
		var failedNodeGroupIds []string
		for _, failedNodeGroup := range failedNodeGroups {
			failedNodeGroupIds = append(failedNodeGroupIds, failedNodeGroup.Id())
		}
		klog.Errorf("Async scale-up for asynchronously created node group failed: %v (node groups: %v)", err, failedNodeGroupIds)
		return
	}
	klog.Infof("Initial scale-up succeeded. Scale ups: %v", scaleUpInfos)
}

func nodeGroupIds(nodeGroups []cloudprovider.NodeGroup) []string {
	var result []string
	for _, ng := range nodeGroups {
		result = append(result, ng.Id())
	}
	return result
}
