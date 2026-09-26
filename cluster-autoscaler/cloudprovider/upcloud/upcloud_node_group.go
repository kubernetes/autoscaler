/*
Copyright 2023 The Kubernetes Authors.

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

package upcloud

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/request"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
)

// upCloudNodeGroup implements cloudprovide.NodeGroup interfaces
type upCloudNodeGroup struct {
	clusterID uuid.UUID
	name      string
	size      int
	minSize   int
	maxSize   int

	nodes []cloudprovider.Instance
	svc   upCloudService

	mu sync.Mutex
}

// Id returns an unique identifier of the node group.
func (u *upCloudNodeGroup) Id() string { //nolint: stylecheck
	id := fmt.Sprintf("%s/%s", u.clusterID.String(), u.name)
	// set log level higher because this get called a lot
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Id called", id)
	return id
}

// MinSize returns minimum size of the node group.
func (u *upCloudNodeGroup) MinSize(ctx context.Context) int {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.MinSize called", u.Id())
	return u.minSize
}

// MaxSize returns maximum size of the node group.
func (u *upCloudNodeGroup) MaxSize(ctx context.Context) int {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.MaxSize called", u.Id())
	return u.maxSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (u *upCloudNodeGroup) TargetSize(ctx context.Context) (int, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.TargetSize called (%d)", u.Id(), u.size)
	return u.size, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (u *upCloudNodeGroup) IncreaseSize(ctx context.Context, delta int) error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.IncreaseSize(%d) called", u.Id(), delta)
	if delta <= 0 {
		return fmt.Errorf("failed to increase node group size, delta=%d", delta)
	}
	size := u.size + delta
	if size > u.MaxSize(ctx) {
		return fmt.Errorf("failed to increase node group size, current=%d want=%d max=%d", u.size, size, u.MaxSize(ctx))
	}
	return u.scaleNodeGroup(ctx, size)
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (u *upCloudNodeGroup) DecreaseTargetSize(ctx context.Context, delta int) error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.DecreaseTargetSize(%d) called", u.Id(), delta)
	if delta >= 0 {
		return fmt.Errorf("failed to increase node group size, delta=%d", delta)
	}
	size := u.size + delta
	if size < u.MinSize(ctx) {
		return fmt.Errorf("failed to decrease node group size, current=%d want=%d min=%d", u.size, size, u.MinSize(ctx))
	}
	return u.scaleNodeGroup(ctx, size)
}

func (u *upCloudNodeGroup) scaleNodeGroup(ctx context.Context, size int) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	svcCtx, cancel := context.WithTimeout(ctx, timeoutModifyNodeGroup)
	defer cancel()
	klog.V(logInfo).Infof("scaling node group %s from %d to %d", u.Id(), u.size, size)
	_, err := u.svc.ModifyKubernetesNodeGroup(svcCtx, &request.ModifyKubernetesNodeGroupRequest{
		ClusterUUID: u.clusterID.String(),
		Name:        u.name,
		NodeGroup: request.ModifyKubernetesNodeGroup{
			Count: size,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to scale node group %s, %w", u.name, err)
	}
	nodeGroup, err := u.waitNodeGroupState(ctx, upcloud.KubernetesNodeGroupStateRunning, timeoutWaitNodeGroupState)
	if err != nil {
		return err
	}
	u.size = nodeGroup.Count
	return nil
}

func (u *upCloudNodeGroup) waitNodeGroupState(ctx context.Context, state upcloud.KubernetesNodeGroupState, timeout time.Duration) (*upcloud.KubernetesNodeGroupDetails, error) {
	deadline := time.Now().Add(timeout)
	i := 1
	klog.V(logInfo).Infof("waiting node group %s state %s", u.Id(), state)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting node group %s state %s cancelled; %w", u.Id(), state, ctx.Err())
		case <-ticker.C:
			svcCtx, cancel := context.WithTimeout(ctx, timeoutGetRequest)
			g, err := u.svc.GetKubernetesNodeGroup(svcCtx, &request.GetKubernetesNodeGroupRequest{
				ClusterUUID: u.clusterID.String(),
				Name:        u.name,
			})
			cancel()
			if err != nil {
				return g, fmt.Errorf("failed to fetch node group %s, %w", u.Id(), err)
			}
			if g.State == state {
				return g, nil
			}
			klog.V(logInfo).Infof("waiting(%d) node group %s state %s (%s)", i, u.Id(), state, g.State)
		}
		i++
	}
	return nil, fmt.Errorf("node group %s state check (%d) timed out", u.Id(), i)
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (u *upCloudNodeGroup) DeleteNodes(ctx context.Context, nodes []*apiv1.Node) error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.DeleteNodes called", u.Id())
	u.mu.Lock()
	defer u.mu.Unlock()

	for i := range nodes {
		if err := u.deleteNode(ctx, nodes[i].GetName()); err != nil {
			return err
		}
		nodeGroup, err := u.waitNodeGroupState(ctx, upcloud.KubernetesNodeGroupStateRunning, timeoutWaitNodeGroupState)
		if err != nil {
			return err
		}
		u.size = nodeGroup.Count
	}
	return nil
}

// ForceDeleteNodes deletes nodes from this node group without additional
// provider-side constraints. UpCloud exposes the same deletion flow here as the
// regular delete path.
func (u *upCloudNodeGroup) ForceDeleteNodes(ctx context.Context, nodes []*apiv1.Node) error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.ForceDeleteNodes called", u.Id())
	return u.DeleteNodes(ctx, nodes)
}

func (u *upCloudNodeGroup) deleteNode(ctx context.Context, nodeName string) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutDeleteNode)
	defer cancel()
	klog.V(logInfo).Infof("deleting UpCloud %s/node %s", u.Id(), nodeName)
	return u.svc.DeleteKubernetesNodeGroupNode(ctx, &request.DeleteKubernetesNodeGroupNodeRequest{
		ClusterUUID: u.clusterID.String(),
		Name:        u.name,
		NodeName:    nodeName,
	})
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (u *upCloudNodeGroup) Nodes(ctx context.Context) ([]cloudprovider.Instance, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Nodes called", u.Id())
	return u.nodes, nil
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (u *upCloudNodeGroup) Autoprovisioned(ctx context.Context) bool {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Autoprovisioned called", u.Id())
	return false
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (u *upCloudNodeGroup) Create(ctx context.Context) (cloudprovider.NodeGroup, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Create called", u.Id())
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (u *upCloudNodeGroup) Delete(ctx context.Context) error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Delete called", u.Id())
	return cloudprovider.ErrNotImplemented
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional.
func (u *upCloudNodeGroup) GetOptions(ctx context.Context, _ config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.GetOptions called", u.Id())
	return nil, cloudprovider.ErrNotImplemented
}

// Debug returns a string containing all information regarding this node group.
func (u *upCloudNodeGroup) Debug(ctx context.Context) string {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Debug called", u.Id())
	return fmt.Sprintf("Node group ID: %s (min:%d max:%d)", u.Id(), u.MinSize(ctx), u.MaxSize(ctx))
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (u *upCloudNodeGroup) Exist(ctx context.Context) bool {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Exist called", u.Id())
	return u.name != ""
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (u *upCloudNodeGroup) TemplateNodeInfo(ctx context.Context) (*framework.NodeInfo, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.TemplateNodeInfo called", u.Id())
	return nil, cloudprovider.ErrNotImplemented
}

// AtomicIncreaseSize tries to increase the size of the node group atomically.
//   - If the method returns nil, it guarantees that delta instances will be added to the node group
//     within its MaxNodeProvisionTime. The function should wait until node group size is updated.
//     The cloud provider is responsible for tracking and ensuring successful scale up asynchronously.
//   - If the method returns an error, it guarantees that no new instances will be added to the node group
//     as a result of this call. The cloud provider is responsible for ensuring that before returning from the method.
//
// Implementation is optional. If implemented, CA will take advantage of the method while scaling up
// GenericScaleUp ProvisioningClass, guaranteeing that all instances required for such a ProvisioningRequest
// are provisioned atomically.
func (u *upCloudNodeGroup) AtomicIncreaseSize(ctx context.Context, _ int) error {
	return cloudprovider.ErrNotImplemented
}
