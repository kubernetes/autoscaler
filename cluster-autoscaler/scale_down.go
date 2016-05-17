/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package main

import (
	"fmt"
	"time"

	"k8s.io/contrib/cluster-autoscaler/config"
	"k8s.io/contrib/cluster-autoscaler/simulator"
	"k8s.io/contrib/cluster-autoscaler/utils/gce"
	kube_api "k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// ScaleDownResult represents the state of scale down.
type ScaleDownResult int

const (
	// ScaleDownError - scale down finished with error.
	ScaleDownError ScaleDownResult = iota
	// ScaleDownNoUnderutilized - no underutilized nodedes and no errors.
	ScaleDownNoUnderutilized ScaleDownResult = iota
	// ScaleDownNoNodeDeleted - underutilized nodes present but not available for deletion.
	ScaleDownNoNodeDeleted ScaleDownResult = iota
	// ScaleDownNodeDeleted - a node was deleted.
	ScaleDownNodeDeleted ScaleDownResult = iota
)

// CalculateUnderutilizedNodes calculates which nodes are underutilized.
func CalculateUnderutilizedNodes(nodes []*kube_api.Node,
	underutilizedNodes map[string]time.Time,
	utilizationThreshold float64,
	pods []*kube_api.Pod) map[string]time.Time {

	currentlyUnderutilizedNodes := make(map[string]struct{})
	nodeNameToNodeInfo := schedulercache.CreateNodeNameToInfoMap(pods)

	for _, node := range nodes {
		nodeInfo, found := nodeNameToNodeInfo[node.Name]
		if !found {
			glog.Errorf("Node info for %s not found", node.Name)
			continue
		}
		reservation, err := simulator.CalculateReservation(node, nodeInfo)

		if err != nil {
			glog.Warningf("Failed to calculate reservation for %s: %v", node.Name, err)
		}
		glog.V(4).Infof("Node %s - reservation %f", node.Name, reservation)

		if reservation >= utilizationThreshold {
			glog.V(4).Infof("Node %s is not suitable for removal - reservation to big (%f)", node.Name, reservation)
			continue
		}
		currentlyUnderutilizedNodes[node.Name] = struct{}{}
	}

	now := time.Now()
	result := make(map[string]time.Time)
	for name := range currentlyUnderutilizedNodes {
		if val, found := underutilizedNodes[name]; !found {
			result[name] = now
		} else {
			result[name] = val
		}
	}
	return result
}

// ScaleDown tries to scale down the cluster. It returns ScaleDownResult indicating if any node was
// removed and error if such occured.
func ScaleDown(
	nodes []*kube_api.Node,
	underutilizedNodes map[string]time.Time,
	underutilizationTime time.Duration,
	pods []*kube_api.Pod,
	gceManager *gce.GceManager,
	client *kube_client.Client) (ScaleDownResult, error) {

	now := time.Now()
	candidates := make([]*kube_api.Node, 0)
	for _, node := range nodes {
		if val, found := underutilizedNodes[node.Name]; found {
			if val.Add(underutilizationTime).Before(now) {
				candidates = append(candidates, node)
			}
		}
	}
	if len(candidates) == 0 {
		glog.Infof("No candidates for scale down")
		return ScaleDownNoUnderutilized, nil
	}

	nodeToRemove, err := simulator.FindNodeToRemove(candidates, nodes, pods, client)
	if err != nil {
		return ScaleDownError, fmt.Errorf("Find node to remove failed: %v", err)
	}
	if nodeToRemove == nil {
		glog.V(1).Infof("No node to remove")
		return ScaleDownNoNodeDeleted, nil
	}
	glog.Infof("Removing %s", nodeToRemove.Name)

	instanceConfig, err := config.InstanceConfigFromProviderId(nodeToRemove.Spec.ProviderID)
	if err != nil {
		return ScaleDownError, fmt.Errorf("Failed to get instance config for %s: %v", nodeToRemove.Name, err)
	}

	err = gceManager.DeleteInstances([]*config.InstanceConfig{instanceConfig})
	if err != nil {
		return ScaleDownError, fmt.Errorf("Failed to delete %v: %v", instanceConfig, err)
	}

	return ScaleDownNodeDeleted, nil
}
