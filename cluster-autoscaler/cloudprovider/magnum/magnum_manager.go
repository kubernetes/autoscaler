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
	"fmt"
	"io"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/compute/v2/flavors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

const (
	// Magnum microversion that must be requested to use the node groups API.
	microversionNodeGroups = "1.9"
	// Magnum microversion that must be requested to support scaling node groups to 0 nodes.
	microversionScaleToZero = "1.10"
	// Magnum interprets "latest" to mean the highest available microversion.
	microversionLatest = "latest"
)

// magnumManager is an interface for the basic interactions with the cluster.
type magnumManager interface {
	nodeGroupSize(nodegroup string) (int, error)
	updateNodeCount(nodegroup string, nodes int) error
	getNodes(nodegroup string) ([]cloudprovider.Instance, error)
	deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error
	autoDiscoverNodeGroups(cfgs []magnumAutoDiscoveryConfig) ([]*nodegroups.NodeGroup, error)
	fetchNodeGroupStackIDs(nodegroup string) (nodeGroupStacks, error)
	uniqueNameAndIDForNodeGroup(nodegroup string) (string, string, error)
	nodeGroupForNode(node *apiv1.Node) (string, error)
	getFlavorById(flavor string) (*flavors.Flavor, error)
}

// createMagnumManager creates the necessary OpenStack clients and returns
// an instance of magnumManagerImpl.
func createMagnumManager(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (magnumManager, error) {
	cfg, err := readConfig(configReader)
	if err != nil {
		return nil, err
	}

	provider, err := createProviderClient(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("could not create provider client: %v", err)
	}

	clusterClient, err := createClusterClient(cfg, provider, opts)
	if err != nil {
		return nil, err
	}

	clusterClient.Microversion = microversionLatest

	// This replaces the cluster name with a UUID if the name was given in the parameters.
	err = checkClusterUUID(provider, clusterClient, opts)
	if err != nil {
		return nil, fmt.Errorf("could not check cluster UUID: %v", err)
	}

	heatClient, err := createHeatClient(cfg, provider, opts)
	if err != nil {
		return nil, fmt.Errorf("could not create heat client: %v", err)
	}

	novaClient, err := createNovaClient(cfg, provider, opts)
	if err != nil {
		return nil, fmt.Errorf("could not create nova client: %v", err)
	}

	return createMagnumManagerImpl(clusterClient, heatClient, novaClient, opts)
}
