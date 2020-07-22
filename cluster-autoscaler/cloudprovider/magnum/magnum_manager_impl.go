/*
Copyright 2020 The Kubernetes Authors.

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

package magnum

import (
	"fmt"
	"sort"
	"strings"

	"github.com/satori/go.uuid"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/clusters"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/orchestration/v1/stackresources"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/orchestration/v1/stacks"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	klog "k8s.io/klog/v2"
)

type nodeGroupStacks struct {
	stackID   string
	stackName string

	kubeMinionsStackID   string
	kubeMinionsStackName string
}

// magnumManagerImpl implements the magnumManager interface.
type magnumManagerImpl struct {
	clusterClient *gophercloud.ServiceClient
	heatClient    *gophercloud.ServiceClient

	clusterName string

	stackInfo                  map[string]nodeGroupStacks
	providerIDToNodeGroupCache map[string]string
}

// createMagnumManagerImpl creates an instance of magnumManagerImpl.
func createMagnumManagerImpl(clusterClient, heatClient *gophercloud.ServiceClient, opts config.AutoscalingOptions) (*magnumManagerImpl, error) {
	manager := magnumManagerImpl{
		clusterClient: clusterClient,
		heatClient:    heatClient,
		clusterName:   opts.ClusterName,
		stackInfo:     make(map[string]nodeGroupStacks),

		providerIDToNodeGroupCache: make(map[string]string),
	}

	return &manager, nil
}

func uniqueName(ng *nodegroups.NodeGroup) string {
	name := ng.Name
	id := ng.UUID

	idSegment := strings.Split(id, "-")[0]
	uniqueName := fmt.Sprintf("%s-%s", name, idSegment)

	return uniqueName
}

func (mgr *magnumManagerImpl) uniqueNameAndIDForNodeGroup(nodegroup string) (string, string, error) {
	ng, err := nodegroups.Get(mgr.clusterClient, mgr.clusterName, nodegroup).Extract()
	if err != nil {
		return "", "", fmt.Errorf("could not get node group: %v", err)
	}

	uniqueName := uniqueName(ng)

	return uniqueName, ng.UUID, nil
}

// fetchNodeGroupStackIDs fetches and caches the IDs and names of the
// nodegroup Heat stack and the related kube_minions stack.
//
// Calling this a second time for the same node group will return the
// cached result.
func (mgr *magnumManagerImpl) fetchNodeGroupStackIDs(nodegroup string) (nodeGroupStacks, error) {
	if stacks, ok := mgr.stackInfo[nodegroup]; ok {
		// Stack info has already been fetched
		return stacks, nil
	}

	// Get the node group.
	ng, err := nodegroups.Get(mgr.clusterClient, mgr.clusterName, nodegroup).Extract()
	if err != nil {
		return nodeGroupStacks{}, fmt.Errorf("could not get node group: %v", err)
	}
	stackID := ng.StackID

	// Get the Heat stack for the node group.
	stack, err := stacks.Find(mgr.heatClient, stackID).Extract()
	if err != nil {
		return nodeGroupStacks{}, fmt.Errorf("could not find node group stack: %v", err)
	}

	stackName := stack.Name

	// Get the kube_minions stack resource for the node group stack.
	minionsResource, err := stackresources.Get(mgr.heatClient, stackName, stackID, "kube_minions").Extract()
	if err != nil {
		return nodeGroupStacks{}, fmt.Errorf("could not get node group stack kube_minions resource: %v", err)
	}

	minionsStackID := minionsResource.PhysicalID

	// Get the stack for that resource.
	minionsStack, err := stacks.Find(mgr.heatClient, minionsStackID).Extract()
	if err != nil {
		return nodeGroupStacks{}, fmt.Errorf("could not find node group kube_minions stack: %v", err)
	}

	minionsStackName := minionsStack.Name

	mgr.stackInfo[nodegroup] = nodeGroupStacks{
		stackID:              stackID,
		stackName:            stackName,
		kubeMinionsStackID:   minionsStackID,
		kubeMinionsStackName: minionsStackName,
	}

	return mgr.stackInfo[nodegroup], nil
}

// autoDiscoverNodeGroups lists all node groups that belong to this cluster
// and finds the ones which are valid for autoscaling and that match the
// auto discovery configuration.
func (mgr *magnumManagerImpl) autoDiscoverNodeGroups(cfgs []magnumAutoDiscoveryConfig) ([]*nodegroups.NodeGroup, error) {
	ngs := []*nodegroups.NodeGroup{}

	pages, err := nodegroups.List(mgr.clusterClient, mgr.clusterName, nodegroups.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("could not fetch node group pages: %v", err)
	}
	groups, err := nodegroups.ExtractNodeGroups(pages)
	if err != nil {
		return nil, fmt.Errorf("could not extract node groups: %v", err)
	}

	for _, group := range groups {
		if group.Role == "master" {
			// Don't yet support autoscaling for master node groups.
			continue
		}

		// Listing node groups does not return the min/max node count,
		// have to use a Get for those properties.
		detail, err := nodegroups.Get(mgr.clusterClient, mgr.clusterName, group.UUID).Extract()
		if err != nil {
			return nil, fmt.Errorf("could not get detail for node group %s: %v", group.Name, err)
		}

		// Max node count must be set to be eligible for autoscaling.
		if detail.MaxNodeCount == nil {
			klog.V(4).Infof("Node group %s does not have max node count set", detail.Name)
			continue
		}

		// The group must match at least one auto discovery config.
		var matchesAny bool
		for _, cfg := range cfgs {
			for _, role := range cfg.Roles {
				if detail.Role == role {
					matchesAny = true
				}
			}
		}
		if !matchesAny {
			klog.V(2).Infof("Node group %s has max node count set but does not match any auto discovery configs", detail.Name)
			continue
		}

		ngs = append(ngs, detail)
	}

	return ngs, nil
}

// nodeGroupSize gets the current node count of the given node group.
func (mgr *magnumManagerImpl) nodeGroupSize(nodegroup string) (int, error) {
	ng, err := nodegroups.Get(mgr.clusterClient, mgr.clusterName, nodegroup).Extract()
	if err != nil {
		return 0, fmt.Errorf("could not get node group: %v", err)
	}
	return ng.NodeCount, nil
}

// updateNodeCount performs a cluster resize targeting the given node group.
func (mgr *magnumManagerImpl) updateNodeCount(nodegroup string, nodes int) error {
	resizeOpts := clusters.ResizeOpts{
		NodeCount: &nodes,
		NodeGroup: nodegroup,
	}

	resizeResult := clusters.Resize(mgr.clusterClient, mgr.clusterName, resizeOpts)
	_, err := resizeResult.Extract()
	if err != nil {
		return fmt.Errorf("could not resize cluster: %v", err)
	}
	return nil
}

// getNodes returns Instances with ProviderIDs and running states
// of all nodes that exist in OpenStack for a node group.
func (mgr *magnumManagerImpl) getNodes(nodegroup string) ([]cloudprovider.Instance, error) {
	var nodes []cloudprovider.Instance

	stackInfo, err := mgr.fetchNodeGroupStackIDs(nodegroup)
	if err != nil {
		return nil, fmt.Errorf("could not fetch stack IDs for node group %s: %v", nodegroup, err)
	}

	minionResourcesPages, err := stackresources.List(mgr.heatClient, stackInfo.kubeMinionsStackName, stackInfo.kubeMinionsStackID, nil).AllPages()
	if err != nil {
		return nil, fmt.Errorf("could not list minion resources: %v", err)
	}

	minionResources, err := stackresources.ExtractResources(minionResourcesPages)
	if err != nil {
		return nil, fmt.Errorf("could not extract minion resources: %v", err)
	}

	stack, err := stacks.Get(mgr.heatClient, stackInfo.kubeMinionsStackName, stackInfo.kubeMinionsStackID).Extract()
	if err != nil {
		return nil, fmt.Errorf("could not get kube_minions stack from heat: %v", err)
	}

	// mapping from minion index to server ID e.g
	// "0": "4c30961a-6e2f-42be-be01-5270e1546a89"
	//
	// The value in refs_map goes through several stages:
	// 1. The initial value is the node index (same as the key).
	// 2. It then changes to be "kube-minion".
	// 3. When a server has been created it changes to the server ID.
	refsMap := make(map[string]string)
	for _, output := range stack.Outputs {
		if output["output_key"] == "refs_map" {
			refsMapOutput := output["output_value"].(map[string]interface{})
			for index, ID := range refsMapOutput {
				refsMap[index] = ID.(string)
			}
		}
	}

	for _, minion := range minionResources {
		// Prepare fake provider ID in the format "fake:///nodegroup/index" in case the minion does not yet have a server ID in refs_map.
		// This fake provider ID is necessary to have in case a server can not be created (e.g quota exceeded).
		// The minion.Name is its index e.g "2".
		fakeName := fmt.Sprintf("fake:///%s/%s", nodegroup, minion.Name)
		instance := cloudprovider.Instance{Id: fakeName, Status: &cloudprovider.InstanceStatus{}}

		switch minion.Status {
		case "DELETE_COMPLETE":
			// Don't return this instance
			continue
		case "DELETE_IN_PROGRESS":
			serverID, found := refsMap[minion.Name]
			if !found || serverID == "kube-minion" {
				// If a server ID can't be found for this minion, assume it is already deleted.
				klog.V(4).Infof("Minion %q is DELETE_IN_PROGRESS but has no refs_map entry", minion.Name)
				continue
			}
			instance.Id = fmt.Sprintf("openstack:///%s", serverID)
			instance.Status.State = cloudprovider.InstanceDeleting
		case "INIT_COMPLETE", "CREATE_IN_PROGRESS":
			instance.Status.State = cloudprovider.InstanceCreating
		case "UPDATE_IN_PROGRESS":
			// UPDATE_IN_PROGRESS can either be a creating node that was moved to updating by a separate
			// stack update before going to a complete status, or an old node that has already completed and is
			// only temporarily in an updating status.
			// We need to differentiate between these two states for the instance status.

			// If the minion is not yet in the refs_map it must still be creating.
			serverID, found := refsMap[minion.Name]
			if !found || serverID == "kube-minion" {
				instance.Status.State = cloudprovider.InstanceCreating
				klog.V(4).Infof("Minion %q is UPDATE_IN_PROGRESS but has no refs_map entry", minion.Name)
				break
			}

			instance.Id = fmt.Sprintf("openstack:///%s", serverID)

			// Otherwise, have to check the stack resources for this minion, as they do not change even when the stack is updated.
			// There are several resources but the two important ones are kube-minion (provisioning the server)
			// and node_config_deployment (the heat-container-agent running on the node).
			// If these two are CREATE_COMPLETE then this node must be a running node, not one being created.

			minionStackID := minion.PhysicalID

			// Only the stack ID is known, not the stack name, so this operation has to be a Find.
			minionResources, err := stackresources.Find(mgr.heatClient, minionStackID).Extract()
			if err != nil {
				return nil, fmt.Errorf("could not get stack resources for minion %q", minion.Name)
			}

			// The Find returns a list of all resources for this node, we have to loop through and find
			// the right statuses.
			var minionServerStatus string
			var minionNodeDeploymentStatus string

			for _, resource := range minionResources {
				switch resource.Name {
				case "kube-minion":
					minionServerStatus = resource.Status
				case "node_config_deployment":
					minionNodeDeploymentStatus = resource.Status
				}
			}

			if minionServerStatus == "CREATE_COMPLETE" && minionNodeDeploymentStatus == "CREATE_COMPLETE" {
				// The minion is one that is already running.
				klog.V(4).Infof("Minion %q in UPDATE_IN_PROGRESS is an already running node", minion.Name)
				instance.Status.State = cloudprovider.InstanceRunning
			} else {
				// The minion is one that is still being created.
				klog.V(4).Infof("Minion %q in UPDATE_IN_PROGRESS is a new node", minion.Name)
				instance.Status.State = cloudprovider.InstanceCreating
			}
		case "CREATE_FAILED", "UPDATE_FAILED":
			instance.Status.State = cloudprovider.InstanceCreating

			errorClass := cloudprovider.OtherErrorClass

			// Check if the error message is for exceeding the project quota.
			if strings.Contains(strings.ToLower(minion.StatusReason), "quota") {
				errorClass = cloudprovider.OutOfResourcesErrorClass
			}

			instance.Status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
				ErrorClass:   errorClass,
				ErrorMessage: minion.StatusReason,
			}

			klog.V(3).Infof("Instance %s failed with reason: %s", minion.Name, minion.StatusReason)
		case "CREATE_COMPLETE", "UPDATE_COMPLETE":
			if serverID, found := refsMap[minion.Name]; found && serverID != "kube-minion" {
				instance.Id = fmt.Sprintf("openstack:///%s", serverID)
			}
			instance.Status.State = cloudprovider.InstanceRunning
		default:
			// If the minion is in an unknown state.
			klog.V(3).Infof("Ignoring minion %s in state %s", minion.Name, minion.Status)
			continue
		}

		nodes = append(nodes, instance)
	}

	return nodes, nil
}

// deleteNodes deletes nodes by resizing the cluster to a smaller size
// and specifying which nodes should be removed.
//
// The nodes are referenced by server ID for nodes which have them,
// or by the minion index for nodes which are creating or in an error state.
func (mgr *magnumManagerImpl) deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error {
	var nodesToRemove []string
	for _, nodeRef := range nodes {
		if nodeRef.IsFake {
			_, index, err := parseFakeProviderID(nodeRef.Name)
			if err != nil {
				return fmt.Errorf("error handling fake node: %v", err)
			}
			nodesToRemove = append(nodesToRemove, index)
			continue
		}
		klog.V(2).Infof("manager deleting node: %s", nodeRef.Name)
		nodesToRemove = append(nodesToRemove, nodeRef.SystemUUID)
	}

	resizeOpts := clusters.ResizeOpts{
		NodeCount:     &updatedNodeCount,
		NodesToRemove: nodesToRemove,
		NodeGroup:     nodegroup,
	}

	klog.V(2).Infof("resizeOpts: node_count=%d, remove=%v", *resizeOpts.NodeCount, resizeOpts.NodesToRemove)

	resizeResult := clusters.Resize(mgr.clusterClient, mgr.clusterName, resizeOpts)
	_, err := resizeResult.Extract()
	if err != nil {
		return fmt.Errorf("could not resize cluster: %v", err)
	}

	return nil
}

// nodeGroupForNode returns the UUID of the node group that the given node is a member of.
func (mgr *magnumManagerImpl) nodeGroupForNode(node *apiv1.Node) (string, error) {
	if groupUUID, ok := mgr.providerIDToNodeGroupCache[node.Spec.ProviderID]; ok {
		klog.V(5).Infof("nodeGroupForNode: already cached %s in node group %s", node.Spec.ProviderID, groupUUID)
		return groupUUID, nil
	}

	// There are two possibilities for "fake" nodes:
	// * The node is in creation and has a provider ID like fake:///
	// * The node has just been created but is not yet registered in kubernetes,
	//   in which case it is "fake" but has an openstack:/// provider ID as it
	//   does exist in OpenStack.
	// Only the first case needs to be handled specially.
	if isFakeNode(node) && strings.HasPrefix(node.Spec.ProviderID, "fake:///") {
		groupUUID, _, err := parseFakeProviderID(node.Spec.ProviderID)
		if err != nil {
			return "", err
		}
		klog.V(5).Infof("nodeGroupForNode: parsed fake node, %s in node group %s", node.Spec.ProviderID, groupUUID)
		return groupUUID, nil
	}

	// Otherwise, have to loop through all node groups and check the stack of each one.
	pages, err := nodegroups.List(mgr.clusterClient, mgr.clusterName, nodegroups.ListOpts{}).AllPages()
	if err != nil {
		return "", fmt.Errorf("could not fetch node group pages: %v", err)
	}
	allNodeGroups, err := nodegroups.ExtractNodeGroups(pages)
	if err != nil {
		return "", fmt.Errorf("could not extract node groups: %v", err)
	}

	allNodeGroupUUIDs := []string{}
	nodeGroupSizes := make(map[string]int)
	for _, group := range allNodeGroups {
		if group.Role == "master" {
			continue
		}
		allNodeGroupUUIDs = append(allNodeGroupUUIDs, group.UUID)
		nodeGroupSizes[group.UUID] = group.NodeCount
	}

	// Sort the node groups IDs to put the groups with the highest number of nodes first.
	// This should maximise the chance of finding the node early.
	sort.Slice(allNodeGroupUUIDs, func(i, j int) bool {
		return nodeGroupSizes[allNodeGroupUUIDs[i]] > nodeGroupSizes[allNodeGroupUUIDs[j]]
	})

	for _, ngUUID := range allNodeGroupUUIDs {
		klog.V(5).Infof("Checking node group %s, size %d", ngUUID, nodeGroupSizes[ngUUID])

		stackInfo, err := mgr.fetchNodeGroupStackIDs(ngUUID)
		if err != nil {
			return "", fmt.Errorf("could not fetch stack IDs for node group %s: %v", ngUUID, err)
		}

		minionsStack, err := stacks.Get(mgr.heatClient, stackInfo.kubeMinionsStackName, stackInfo.kubeMinionsStackID).Extract()
		if err != nil {
			return "", fmt.Errorf("could not get minions stack: %v", err)
		}

		for _, output := range minionsStack.Outputs {
			if output["output_key"] == "refs_map" {
				refsMapOutput, ok := output["output_value"].(map[string]interface{})
				if !ok {
					// Output value was nil, possibly because the node group is being deleted.
					return "", fmt.Errorf("could not check the minions stack refs_map")
				}

				// Temporarily hold the result and loop over the entire refs_map
				// instead of returning instantly, to cache as much as possible.
				var found bool
				for _, ID := range refsMapOutput {
					ID := ID.(string)

					// Check that the ID is a proper server UUID.
					// (It could be an index or "kube-minion" in the refs_map instead).
					_, err := uuid.FromString(ID)
					if err != nil {
						continue
					}

					// Convert to a providerID and cache it as belonging to this node group.
					providerID := fmt.Sprintf("openstack:///%s", ID)
					mgr.providerIDToNodeGroupCache[providerID] = ngUUID

					// If the node matches this provider ID, then remember that this is
					// the correct node group but keep looping through the refs_map
					// to cache other providerIDs.
					if providerID == node.Spec.ProviderID {
						found = true
					}
				}
				if found {
					klog.V(5).Infof("nodeGroupForNode: found node %s in node group %s", node.Spec.ProviderID, ngUUID)
					return ngUUID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find node group for node %s", node.Spec.ProviderID)
}
