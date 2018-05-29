/*
Copyright 2018 The Kubernetes Authors.

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

package nodegroups

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"
	"k8s.io/kubernetes/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// AutoprovisioningNodeGroupListProcessor adds autoprovisioning candidates to consider in scale-up.
type AutoprovisioningNodeGroupListProcessor struct {
}

// NewAutoprovisioningNodeGroupListProcessor creates an instance of NodeGroupListProcessor.
func NewAutoprovisioningNodeGroupListProcessor() NodeGroupListProcessor {
	return &AutoprovisioningNodeGroupListProcessor{}
}

// Process processes lists of unschedulable and sheduled pods before scaling of the cluster.
func (p *AutoprovisioningNodeGroupListProcessor) Process(context *context.AutoscalingContext, nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*schedulercache.NodeInfo,
	unschedulablePods []*apiv1.Pod) ([]cloudprovider.NodeGroup, map[string]*schedulercache.NodeInfo, error) {

	if !context.AutoscalingOptions.NodeAutoprovisioningEnabled {
		return nodeGroups, nodeInfos, nil
	}

	autoprovisionedNodeGroupCount := 0
	for _, group := range nodeGroups {
		if group.Autoprovisioned() {
			autoprovisionedNodeGroupCount++
		}
	}
	if autoprovisionedNodeGroupCount >= context.MaxAutoprovisionedNodeGroupCount {
		glog.V(4).Infof("Max autoprovisioned node group count reached")
		return nodeGroups, nodeInfos, nil
	}

	newGroupsCount := 0

	newNodeGroups := addAllMachineTypesForConfig(context, map[string]string{}, map[string]resource.Quantity{},
		nodeInfos, unschedulablePods)
	newGroupsCount += len(newNodeGroups)
	nodeGroups = append(nodeGroups, newNodeGroups...)

	gpuRequests := gpu.GetGpuRequests(unschedulablePods)
	for _, gpuRequestInfo := range gpuRequests {
		glog.V(4).Info("Adding node groups using GPU to NAP simulations")
		extraResources := map[string]resource.Quantity{
			gpu.ResourceNvidiaGPU: gpuRequestInfo.MaxRequest,
		}
		newNodeGroups := addAllMachineTypesForConfig(context, gpuRequestInfo.SystemLabels, extraResources,
			nodeInfos, gpuRequestInfo.Pods)
		newGroupsCount += len(newNodeGroups)
		nodeGroups = append(nodeGroups, newNodeGroups...)
	}
	glog.V(4).Infof("Considering %v potential node groups in NAP simulations", newGroupsCount)

	return nodeGroups, nodeInfos, nil
}

func addAllMachineTypesForConfig(context *context.AutoscalingContext, systemLabels map[string]string, extraResources map[string]resource.Quantity,
	nodeInfos map[string]*schedulercache.NodeInfo, unschedulablePods []*apiv1.Pod) []cloudprovider.NodeGroup {

	nodeGroups := make([]cloudprovider.NodeGroup, 0)
	machines, err := context.CloudProvider.GetAvailableMachineTypes()
	if err != nil {
		glog.Warningf("Failed to get machine types: %v", err)
		return nodeGroups
	}

	bestLabels := labels.BestLabelSet(unschedulablePods)
	taints := make([]apiv1.Taint, 0)
	for _, machineType := range machines {
		nodeGroup, err := context.CloudProvider.NewNodeGroup(machineType, bestLabels, systemLabels, taints, extraResources)
		if err != nil {
			// We don't check if a given node group setup is allowed.
			// It's fine if it isn't, just don't consider it an option.
			if err != cloudprovider.ErrIllegalConfiguration {
				glog.Warningf("Unable to build temporary node group for %s: %v", machineType, err)
			}
			continue
		}
		nodeInfo, err := nodeGroup.TemplateNodeInfo()
		if err != nil {
			glog.Warningf("Unable to build template for node group for %s: %v", nodeGroup.Id(), err)
			continue
		}
		nodeInfos[nodeGroup.Id()] = nodeInfo
		nodeGroups = append(nodeGroups, nodeGroup)
	}
	return nodeGroups
}
