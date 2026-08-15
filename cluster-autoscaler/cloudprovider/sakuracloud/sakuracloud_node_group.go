/*
Copyright The Kubernetes Authors.

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

package sakuracloud

import (
	"fmt"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
)

// sakuracloudNodeGroup implements cloudprovider.NodeGroup for a set of SAKURA
// cloud servers tagged with ca-group-<id>. SAKURA cloud has no instance-group
// primitive, so the autoscaler itself creates and deletes servers.
type sakuracloudNodeGroup struct {
	id      string
	manager *sakuracloudManager
	config  *nodeGroupConfig

	mu           sync.Mutex
	targetSize   int
	provisioning int
}

// MaxSize returns maximum size of the node group.
func (n *sakuracloudNodeGroup) MaxSize() int {
	return n.config.MaxSize
}

// MinSize returns minimum size of the node group.
func (n *sakuracloudNodeGroup) MinSize() int {
	return n.config.MinSize
}

// GetOptions returns nil, meaning the global autoscaling options are used.
func (n *sakuracloudNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// TargetSize returns the current target size of the node group.
func (n *sakuracloudNodeGroup) TargetSize() (int, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.targetSize, nil
}

// IncreaseSize increases the node group size by provisioning delta servers.
// Provisioning runs asynchronously: disk copy from an archive takes minutes,
// which is far longer than a CA loop iteration.
func (n *sakuracloudNodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}
	n.mu.Lock()
	if n.targetSize+delta > n.MaxSize() {
		n.mu.Unlock()
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d", n.targetSize, n.targetSize+delta, n.MaxSize())
	}
	n.targetSize += delta
	n.provisioning += delta
	n.mu.Unlock()

	for i := 0; i < delta; i++ {
		go func() {
			err := n.manager.createServer(n.id, n.config)
			n.mu.Lock()
			n.provisioning--
			if err != nil {
				n.targetSize--
			}
			n.mu.Unlock()
			if err != nil {
				klog.Errorf("sakuracloud: failed to provision server for node group %s: %v", n.id, err)
			}
		}()
	}
	return nil
}

// AtomicIncreaseSize is not implemented.
func (n *sakuracloudNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes the given nodes (and their servers) from the group.
func (n *sakuracloudNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	for _, node := range nodes {
		serverName, err := serverNameFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return fmt.Errorf("cannot delete node %s: %w", node.Name, err)
		}
		server := n.manager.serverByName(serverName)
		if server == nil {
			klog.V(2).Infof("sakuracloud: server %s already gone, skipping", serverName)
			continue
		}
		if g, ok := server.groupName(); !ok || g != n.id {
			return fmt.Errorf("server %s does not belong to node group %s", serverName, n.id)
		}
		if err := n.manager.deleteServer(server); err != nil {
			return err
		}
		n.mu.Lock()
		n.targetSize--
		n.mu.Unlock()
	}
	if err := n.manager.refreshServers(); err != nil {
		klog.Errorf("sakuracloud: failed to refresh servers after delete: %v", err)
	}
	return nil
}

// ForceDeleteNodes deletes nodes without checking group constraints.
func (n *sakuracloudNodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize lowers the target without deleting existing servers.
// Used to shed target that never materialized (failed provisioning).
func (n *sakuracloudNodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	live := len(n.manager.serversInGroup(n.id))
	if n.targetSize+delta < live {
		return fmt.Errorf("attempt to delete existing nodes: target %d delta %d live %d", n.targetSize, delta, live)
	}
	n.targetSize += delta
	return nil
}

// Id returns the node group identifier.
func (n *sakuracloudNodeGroup) Id() string {
	return n.id
}

// Debug returns a debug string for the node group.
func (n *sakuracloudNodeGroup) Debug() string {
	return fmt.Sprintf("cluster ID: %s (min:%d max:%d target:%d)", n.id, n.MinSize(), n.MaxSize(), n.targetSize)
}

// Nodes returns instances that belong to this node group.
func (n *sakuracloudNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	servers := n.manager.serversInGroup(n.id)
	instances := make([]cloudprovider.Instance, 0, len(servers))
	for _, s := range servers {
		state := cloudprovider.InstanceCreating
		if s.status() == "up" {
			state = cloudprovider.InstanceRunning
		}
		instances = append(instances, cloudprovider.Instance{
			Id:     providerIDForServer(n.manager.zone, s.Name),
			Status: &cloudprovider.InstanceStatus{State: state},
		})
	}
	return instances, nil
}

// TemplateNodeInfo returns a node template used by scale-from-zero
// simulations, advertising the labels and taints from the group config.
func (n *sakuracloudNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	nodeName := fmt.Sprintf("%s-template", n.id)
	node := apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Labels: map[string]string{
				apiv1.LabelHostname:   nodeName,
				apiv1.LabelOSStable:   "linux",
				apiv1.LabelArchStable: "amd64",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewQuantity(int64(n.config.Core), resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(int64(n.config.MemoryGB)*1024*1024*1024, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
			},
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	for k, v := range n.config.Labels {
		node.Labels[k] = v
	}
	node.Spec.Taints = append(node.Spec.Taints, n.config.Taints...)

	return framework.NewNodeInfo(&node, nil, framework.NewPodInfo(cloudprovider.BuildKubeProxy(n.id), nil)), nil
}

// Exist returns true: all node groups come from static configuration.
func (n *sakuracloudNodeGroup) Exist() bool {
	return true
}

// Create is not supported; node groups are statically configured.
func (n *sakuracloudNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete is not supported; node groups are statically configured.
func (n *sakuracloudNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned always returns false.
func (n *sakuracloudNodeGroup) Autoprovisioned() bool {
	return false
}

// syncTargetSize reconciles the in-memory target with reality on Refresh:
// target must never drop below the number of live servers, and with nothing
// provisioning the live count is the truth.
func (n *sakuracloudNodeGroup) syncTargetSize() {
	n.mu.Lock()
	defer n.mu.Unlock()
	live := len(n.manager.serversInGroup(n.id))
	if n.provisioning == 0 && n.targetSize != live {
		klog.V(4).Infof("sakuracloud: node group %s target %d -> %d (live servers)", n.id, n.targetSize, live)
		n.targetSize = live
	} else if n.targetSize < live {
		n.targetSize = live
	}
}
