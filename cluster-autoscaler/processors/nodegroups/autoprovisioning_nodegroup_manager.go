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
	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// AutoprovisioningNodeGroupManager is responsible for creating/deleting autoprovisioned node groups.
type AutoprovisioningNodeGroupManager struct {
}

// NewAutoprovisioningNodeGroupManager creates an instance of NodeGroupManager.
func NewAutoprovisioningNodeGroupManager() NodeGroupManager {
	return &AutoprovisioningNodeGroupManager{}
}

// CreateNodeGroup creates autoprovisioned node group.
func (p *AutoprovisioningNodeGroupManager) CreateNodeGroup(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (cloudprovider.NodeGroup, errors.AutoscalerError) {
	if !context.AutoscalingOptions.NodeAutoprovisioningEnabled {
		return nil, errors.NewAutoscalerError(errors.InternalError, "tried to create a node group %s, but autoprovisioning is disabled", nodeGroup.Id())
	}

	oldId := nodeGroup.Id()
	err := nodeGroup.Create()
	if err != nil {
		context.LogRecorder.Eventf(apiv1.EventTypeWarning, "FailedToCreateNodeGroup",
			"NodeAutoprovisioning: attempt to create node group %v failed: %v", oldId, err)
		// TODO(maciekpytel): add some metric here after figuring out failure scenarios
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}
	newId := nodeGroup.Id()
	if newId != oldId {
		glog.V(2).Infof("Created node group %s based on template node group %s, will use new node group in scale-up", newId, oldId)
	}
	context.LogRecorder.Eventf(apiv1.EventTypeNormal, "CreatedNodeGroup",
		"NodeAutoprovisioning: created new node group %v", newId)
	metrics.RegisterNodeGroupCreation()
	return nodeGroup, nil
}

// RemoveUnneededNodeGroups removes node groups that are not needed anymore.
func (p *AutoprovisioningNodeGroupManager) RemoveUnneededNodeGroups(context *context.AutoscalingContext) error {
	if !context.AutoscalingOptions.NodeAutoprovisioningEnabled {
		return nil
	}
	nodeGroups := context.CloudProvider.NodeGroups()
	for _, nodeGroup := range nodeGroups {
		if !nodeGroup.Autoprovisioned() {
			continue
		}
		targetSize, err := nodeGroup.TargetSize()
		if err != nil {
			return err
		}
		if targetSize > 0 {
			continue
		}
		nodes, err := nodeGroup.Nodes()
		if err != nil {
			return err
		}
		if len(nodes) > 0 {
			continue
		}
		ngId := nodeGroup.Id()
		if err := nodeGroup.Delete(); err != nil {
			context.LogRecorder.Eventf(apiv1.EventTypeWarning, "FailedToDeleteNodeGroup",
				"NodeAutoprovisioning: attempt to delete node group %v failed: %v", ngId, err)
			// TODO(maciekpytel): add some metric here after figuring out failure scenarios
			return err
		}
		context.LogRecorder.Eventf(apiv1.EventTypeNormal, "DeletedNodeGroup",
			"NodeAutoprovisioning: removed node group %v", ngId)
		metrics.RegisterNodeGroupDeletion()
	}
	return nil
}

// CleanUp cleans up the processor's internal structures.
func (p *AutoprovisioningNodeGroupManager) CleanUp() {
}
