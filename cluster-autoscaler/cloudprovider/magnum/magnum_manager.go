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
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const (
	defaultManager = "heat"
)

// magnumManager is an interface for the basic interactions with the cluster.
type magnumManager interface {
	nodeGroupSize(nodegroup string) (int, error)
	updateNodeCount(nodegroup string, nodes int) error
	getNodes(nodegroup string) ([]string, error)
	deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error
	getClusterStatus() (string, error)
	canUpdate() (bool, string, error)
	templateNodeInfo(nodegroup string) (*schedulerframework.NodeInfo, error)
}

// createMagnumManager creates the desired implementation of magnumManager.
// Currently reads the environment variable MAGNUM_MANAGER to find which to create,
// and falls back to a default if the variable is not found.
func createMagnumManager(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (magnumManager, error) {
	// For now get manager from env var, can consider adding flag later
	manager, ok := os.LookupEnv("MAGNUM_MANAGER")
	if !ok {
		manager = defaultManager
	}

	switch manager {
	case "heat":
		return createMagnumManagerHeat(configReader, discoverOpts, opts)
	}

	return nil, fmt.Errorf("magnum manager does not exist: %s", manager)
}
