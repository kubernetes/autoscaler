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

package estimator

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// EstimationContext stores static and runtime state of autoscaling, used by Estimator
type EstimationContext interface {
	SimilarNodeGroups() []cloudprovider.NodeGroup
	ClusterMaxNodeLimit() int
	CurrentNodeCount() int
}

type estimationContext struct {
	similarNodeGroups   []cloudprovider.NodeGroup
	currentNodeCount    int
	clusterMaxNodeLimit int
}

// NewEstimationContext creates a patch for estimation context with runtime properties.
// This patch is used to update existing context.
func NewEstimationContext(clusterMaxNodeLimit int, similarNodeGroups []cloudprovider.NodeGroup, currentNodeCount int) EstimationContext {
	return &estimationContext{
		similarNodeGroups:   similarNodeGroups,
		currentNodeCount:    currentNodeCount,
		clusterMaxNodeLimit: clusterMaxNodeLimit,
	}
}

// SimilarNodeGroups returns array of similar node groups
func (c *estimationContext) SimilarNodeGroups() []cloudprovider.NodeGroup {
	return c.similarNodeGroups
}

// ClusterMaxNodeLimit returns maximum node number allowed for the cluster
func (c *estimationContext) ClusterMaxNodeLimit() int {
	return c.clusterMaxNodeLimit
}

// CurrentNodeCount returns current number of nodes in the cluster
func (c *estimationContext) CurrentNodeCount() int {
	return c.currentNodeCount
}
