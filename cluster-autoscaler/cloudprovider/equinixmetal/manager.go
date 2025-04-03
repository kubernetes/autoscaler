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

package equinixmetal

import (
	"fmt"
	"io"
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

const (
	defaultManager = "rest"
)

// NodeRef stores the name, machineID and providerID of a node.
type NodeRef struct {
	Name       string
	MachineID  string
	ProviderID string
	IPs        []string
}

// equinixMetalManager is an interface for the basic interactions with the cluster.
type equinixMetalManager interface {
	nodeGroupSize(nodegroup string) (int, error)
	createNodes(nodegroup string, nodes int) error
	getNodes(nodegroup string) ([]string, error)
	getNodeNames(nodegroup string) ([]string, error)
	deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error
	templateNodeInfo(nodegroup string) (*framework.NodeInfo, error)
	NodeGroupForNode(labels map[string]string, nodeId string) (string, error)
}

// createEquinixMetalManager creates the desired implementation of equinixMetalManager.
// Currently reads the environment variable EQUINIX_METAL_MANAGER to find which to create,
// and falls back to a default if the variable is not found.
func createEquinixMetalManager(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (equinixMetalManager, error) {
	// For now get manager from env var, can consider adding flag later
	manager, ok := os.LookupEnv("EQUINIX_METAL_MANAGER")
	if !ok {
		if manager, ok = os.LookupEnv("PACKET_MANAGER"); !ok {
			manager = defaultManager
		}
	}

	switch manager {
	case "rest":
		return createEquinixMetalManagerRest(configReader, discoverOpts, opts)
	}

	return nil, fmt.Errorf("equinix metal manager does not exist: %s", manager)
}
