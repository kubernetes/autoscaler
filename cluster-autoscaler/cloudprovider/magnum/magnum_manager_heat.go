/*
Copyright 2019 The Kubernetes Authors.

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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"gopkg.in/gcfg.v1"
	netutil "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/clusters"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/orchestration/v1/stackresources"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/orchestration/v1/stacks"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	certutil "k8s.io/client-go/util/cert"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const (
	stackStatusUpdateInProgress = "UPDATE_IN_PROGRESS"
	stackStatusUpdateComplete   = "UPDATE_COMPLETE"
)

// statusesPreventingUpdate is a set of statuses that would prevent
// the cluster from successfully scaling.
//
// TODO: If it becomes possible to update even in UPDATE_FAILED state then it can be removed here
// https://storyboard.openstack.org/#!/story/2005056
var statusesPreventingUpdate = sets.NewString(
	clusterStatusUpdateInProgress,
	clusterStatusUpdateFailed,
)

// magnumManagerHeat implements the magnumManager interface.
//
// Most interactions with the cluster are done directly with magnum,
// but scaling down requires an intermediate step using heat to
// delete the specific nodes that the autoscaler has picked for removal.
type magnumManagerHeat struct {
	clusterClient *gophercloud.ServiceClient
	heatClient    *gophercloud.ServiceClient
	clusterName   string

	stackName string
	stackID   string

	kubeMinionsStackName string
	kubeMinionsStackID   string

	waitTimeStep time.Duration
}

// createMagnumManagerHeat sets up cluster and stack clients and returns
// an magnumManagerHeat.
func createMagnumManagerHeat(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (*magnumManagerHeat, error) {
	var cfg Config
	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			klog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	if opts.ClusterName == "" {
		klog.Fatalf("The cluster-name parameter must be set")
	}

	authOpts := toAuthOptsExt(cfg)

	provider, err := openstack.NewClient(cfg.Global.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate client: %v", err)
	}

	if cfg.Global.CAFile != "" {
		roots, err := certutil.NewPool(cfg.Global.CAFile)
		if err != nil {
			return nil, err
		}
		config := &tls.Config{}
		config.RootCAs = roots
		provider.HTTPClient.Transport = netutil.SetOldTransportDefaults(&http.Transport{TLSClientConfig: config})

	}

	userAgent := gophercloud.UserAgent{}
	userAgent.Prepend(fmt.Sprintf("cluster-autoscaler/%s", version.ClusterAutoscalerVersion))
	userAgent.Prepend(fmt.Sprintf("cluster/%s", opts.ClusterName))
	provider.UserAgent = userAgent

	klog.V(5).Infof("Using user-agent %s", userAgent.Join())

	err = openstack.AuthenticateV3(provider, authOpts, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("could not authenticate: %v", err)
	}

	clusterClient, err := openstack.NewContainerInfraV1(provider, gophercloud.EndpointOpts{Type: "container-infra", Name: "magnum", Region: cfg.Global.Region})
	if err != nil {
		return nil, fmt.Errorf("could not create container-infra client: %v", err)
	}

	heatClient, err := openstack.NewOrchestrationV1(provider, gophercloud.EndpointOpts{Type: "orchestration", Name: "heat", Region: cfg.Global.Region})
	if err != nil {
		return nil, fmt.Errorf("could not create orchestration client: %v", err)
	}

	manager := magnumManagerHeat{
		clusterClient: clusterClient,
		clusterName:   opts.ClusterName,
		heatClient:    heatClient,
		waitTimeStep:  waitForStatusTimeStep,
	}

	// Check that the cluster exists, and get the ID of its heat stack
	cluster, err := clusters.Get(manager.clusterClient, manager.clusterName).Extract()
	if err != nil {
		return nil, fmt.Errorf("unable to access cluster (%s): %v", manager.clusterName, err)
	}
	manager.stackID = cluster.StackID

	// Prefer to use the cluster UUID if the cluster name was given in the parameters
	if cluster.UUID != opts.ClusterName {
		klog.V(0).Infof("Using cluster UUID %s instead of name %s", cluster.UUID, opts.ClusterName)
		manager.clusterName = cluster.UUID

		userAgent := gophercloud.UserAgent{}
		userAgent.Prepend(fmt.Sprintf("cluster-autoscaler/%s", version.ClusterAutoscalerVersion))
		userAgent.Prepend(fmt.Sprintf("cluster/%s", cluster.UUID))
		provider.UserAgent = userAgent

		klog.V(5).Infof("Using updated user-agent %s", userAgent.Join())
	}

	// Need both the stack name and ID to use in GET requests for the stack, so get name and store that on the manager
	manager.stackName, err = manager.getStackName(manager.stackID)
	if err != nil {
		return nil, fmt.Errorf("could not store stack name on manager: %v", err)
	}

	// Need to be able top access the nested stack which has a mapping between minion indices and IP/ID of the node.
	// The cluster stack has an ID for this nested stack but we need the name as well.
	manager.kubeMinionsStackName, manager.kubeMinionsStackID, err = manager.getKubeMinionsStack(manager.stackName, manager.stackID)
	if err != nil {
		return nil, fmt.Errorf("could not store kube minions stack name/ID on manager: %v", err)
	}

	return &manager, nil
}

// nodeGroupSize gets the current cluster size as reported by magnum.
// The nodegroup argument is ignored as this implementation of magnumManager
// assumes that only a single node group exists.
func (mgr *magnumManagerHeat) nodeGroupSize(nodegroup string) (int, error) {
	cluster, err := clusters.Get(mgr.clusterClient, mgr.clusterName).Extract()
	if err != nil {
		return 0, fmt.Errorf("could not get cluster: %v", err)
	}
	return cluster.NodeCount, nil
}

// updateNodeCount replaces the cluster node_count in magnum.
func (mgr *magnumManagerHeat) updateNodeCount(nodegroup string, nodes int) error {
	updateOpts := []clusters.UpdateOptsBuilder{
		UpdateOptsInt{Op: clusters.ReplaceOp, Path: "/node_count", Value: nodes},
	}
	_, err := clusters.Update(mgr.clusterClient, mgr.clusterName, updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("could not update cluster: %v", err)
	}
	return nil
}

// getNodes should return ProviderIDs for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
// This can not be done with heat currently but a change has been merged upstream
// that will allow this.
func (mgr *magnumManagerHeat) getNodes(nodegroup string) ([]string, error) {
	// TODO: get node ProviderIDs by getting nova instance IDs from heat
	// Waiting for https://review.openstack.org/#/c/639053/ to be able to get
	// nova instance IDs from the kube_minions stack resource.
	// This works fine being empty for now anyway.
	return []string{}, nil
}

// deleteNodes deletes nodes by passing a comma separated list of names or IPs
// of minions to remove to heat, and simultaneously sets the new number of minions on the stack.
// The magnum node_count is then set to the new value (does not cause any more nodes to be removed).
//
// TODO: The two step process is required until https://storyboard.openstack.org/#!/story/2005052
// is complete, which will allow resizing with specific nodes to be deleted as a single Magnum operation.
func (mgr *magnumManagerHeat) deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error {
	stackIndices, err := mgr.findStackIndices(nodes)
	if err != nil {
		return fmt.Errorf("could not find stack indices for nodes to be deleted: %v", err)
	}
	minionsToRemove := strings.Join(stackIndices, ",")

	updateOpts := stacks.UpdateOpts{
		Parameters: map[string]interface{}{
			"minions_to_remove": minionsToRemove,
			"number_of_minions": updatedNodeCount,
		},
	}

	updateResult := stacks.UpdatePatch(mgr.heatClient, mgr.stackName, mgr.stackID, updateOpts)
	err = updateResult.ExtractErr()
	if err != nil {
		return fmt.Errorf("stack patch failed: %v", err)
	}

	// Wait for the stack to do its thing before updating the cluster node_count
	err = mgr.waitForStackStatus(stackStatusUpdateInProgress, waitForUpdateStatusTimeout)
	if err != nil {
		return fmt.Errorf("error waiting for stack %s status: %v", stackStatusUpdateInProgress, err)
	}
	err = mgr.waitForStackStatus(stackStatusUpdateComplete, waitForCompleteStatusTimout)
	if err != nil {
		return fmt.Errorf("error waiting for stack %s status: %v", stackStatusUpdateComplete, err)
	}

	err = mgr.updateNodeCount(nodegroup, updatedNodeCount)
	if err != nil {
		return fmt.Errorf("could not set new cluster size: %v", err)
	}
	return nil
}

// getClusterStatus returns the current status of the magnum cluster.
func (mgr *magnumManagerHeat) getClusterStatus() (string, error) {
	cluster, err := clusters.Get(mgr.clusterClient, mgr.clusterName).Extract()
	if err != nil {
		return "", fmt.Errorf("could not get cluster: %v", err)
	}
	return cluster.Status, nil
}

// canUpdate checks if the cluster status is present in a set of statuses that
// prevent the cluster from being updated.
// Returns if updating is possible and the status for convenience.
func (mgr *magnumManagerHeat) canUpdate() (bool, string, error) {
	clusterStatus, err := mgr.getClusterStatus()
	if err != nil {
		return false, "", fmt.Errorf("could not get cluster status: %v", err)
	}
	return !statusesPreventingUpdate.Has(clusterStatus), clusterStatus, nil
}

// getStackStatus returns the current status of the heat stack used by the magnum cluster.
func (mgr *magnumManagerHeat) getStackStatus() (string, error) {
	stack, err := stacks.Get(mgr.heatClient, mgr.stackName, mgr.stackID).Extract()
	if err != nil {
		return "", fmt.Errorf("could not get stack from heat: %v", err)
	}
	return stack.Status, nil
}

// templateNodeInfo returns a NodeInfo with a node template based on the VM flavor
// that is used to created minions in a given node group.
func (mgr *magnumManagerHeat) templateNodeInfo(nodegroup string) (*schedulerframework.NodeInfo, error) {
	// TODO: create a node template by getting the minion flavor from the heat stack.
	return nil, cloudprovider.ErrNotImplemented
}

// waitForStackStatus checks periodically to see if the heat stack has entered a given status.
// Returns when the status is observed or the timeout is reached.
func (mgr *magnumManagerHeat) waitForStackStatus(status string, timeout time.Duration) error {
	klog.V(2).Infof("Waiting for stack %s status", status)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(mgr.waitTimeStep) {
		currentStatus, err := mgr.getStackStatus()
		if err != nil {
			return fmt.Errorf("error waiting for stack status: %v", err)
		}
		if currentStatus == status {
			klog.V(0).Infof("Waited for stack %s status", status)
			return nil
		}
	}
	return fmt.Errorf("timeout (%v) waiting for stack status %s", timeout, status)
}

// getStackName finds the name of a stack matching a given ID.
func (mgr *magnumManagerHeat) getStackName(stackID string) (string, error) {
	stack, err := stacks.Find(mgr.heatClient, stackID).Extract()
	if err != nil {
		return "", fmt.Errorf("could not find stack with ID %s: %v", mgr.stackID, err)
	}
	klog.V(0).Infof("For stack ID %s, stack name is %s", mgr.stackID, stack.Name)
	return stack.Name, nil
}

// getKubeMinionsStack finds the nested kube_minions stack belonging to the main cluster stack,
// and returns its name and ID.
func (mgr *magnumManagerHeat) getKubeMinionsStack(stackName, stackID string) (name string, ID string, err error) {
	minionsResource, err := stackresources.Get(mgr.heatClient, stackName, stackID, "kube_minions").Extract()
	if err != nil {
		return "", "", fmt.Errorf("could not get kube_minions stack resource: %v", err)
	}

	stack, err := stacks.Find(mgr.heatClient, minionsResource.PhysicalID).Extract()
	if err != nil {
		return "", "", fmt.Errorf("could not find stack matching resource ID in heat: %v", err)
	}

	klog.V(0).Infof("Found nested kube_minions stack: name %s, ID %s", stack.Name, minionsResource.PhysicalID)

	return stack.Name, minionsResource.PhysicalID, nil
}

// findStackIndices finds the stack indices of a set of nodes.
//
// The heat stack stores a mapping between minion indices and their stack IDs.
// The stack IDs are equal to a property of the minions, this could be IP or could be machine ID.
// i.e {'0': '188.185.64.117'}
// or  {'0': 'f12070fef4144bef82812aff177b83c1'}
// Deleting minions in heat can be done with either the index of the minion or the ID associated with it,
// but since the ID could be one of several things it is useful to be able to resolve back to indices.
func (mgr *magnumManagerHeat) findStackIndices(nodeRefs []NodeRef) ([]string, error) {
	stack, err := stacks.Get(mgr.heatClient, mgr.kubeMinionsStackName, mgr.kubeMinionsStackID).Extract()
	if err != nil {
		return nil, fmt.Errorf("could not get kube_minions nested stack from heat: %v", err)
	}

	var IDToIndex = make(map[string]string)
	for _, output := range stack.Outputs {
		if output["output_key"] == "refs_map" {
			minionsMapInterface := output["output_value"].(map[string]interface{})
			for index, ID := range minionsMapInterface {
				IDToIndex[ID.(string)] = index
			}
		}
	}

	var indices []string

	notFound := 0
	for _, ref := range nodeRefs {
		if index, found := stackIndexFromID(IDToIndex, ref); found {
			klog.V(0).Infof("Resolved node %s to stack index %s", ref.Name, index)
			indices = append(indices, index)
		} else {
			klog.V(0).Infof("Could not resolve node %+v to a stack index", ref)
			notFound += 1
		}
	}

	if notFound > 0 {
		return nil, fmt.Errorf("%d nodes could not be resolved to stack indices", notFound)
	}

	return indices, nil
}

// stackIndexFromID finds the index of a given node from the heat kube_minions output map,
// which is provided by findStackIndices (inverted).
// The boolean return value specifies if the index was found or not.
func stackIndexFromID(IDToIndex map[string]string, nodeRef NodeRef) (string, bool) {
	// Kubernetes stores machine UUID without dashes, openstack expects with dashes.
	// Parsing the MachineID and getting the string output gives the correct format.
	// If the MachineID does not parse (maybe it is empty) then it will not be checked, it will not cause an error.
	id, err := uuid.FromString(nodeRef.MachineID)
	if err == nil {
		machineID := id.String()
		if index, found := IDToIndex[machineID]; found {
			return index, found
		}
	}

	for _, IP := range nodeRef.IPs {
		if index, found := IDToIndex[IP]; found {
			return index, found
		}
	}

	return "", false
}

// UpdateOptsInt has a value of type int rather than string.
//
// A Magnum API running with python2 accepts a string for
// replacing node_count, but with python3 only an integer type
// is accepted.
//
// TODO: Can remove when this is provided by gophercloud
// https://github.com/gophercloud/gophercloud/issues/1458
type UpdateOptsInt struct {
	Op    clusters.UpdateOp `json:"op" required:"true"`
	Path  string            `json:"path" required:"true"`
	Value int               `json:"value,omitempty"`
}

// ToClustersUpdateMap assembles a request body based on the contents of
// UpdateOpts.
func (opts UpdateOptsInt) ToClustersUpdateMap() (map[string]interface{}, error) {
	return gophercloud.BuildRequestBody(opts, "")
}
