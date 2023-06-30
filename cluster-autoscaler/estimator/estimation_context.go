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
type EstimationContext struct {
	similarNodeGroups   []cloudprovider.NodeGroup
	currentNodeCount    int
	clusterMaxNodeLimit int
}

// NewEstimationContext creates new estimation context with static properties
func NewEstimationContext(clusterMaxNodeLimit int) *EstimationContext {
	return &EstimationContext{clusterMaxNodeLimit: clusterMaxNodeLimit}
}

// NewEstimationContextUpdate creates a patch for estimation context with runtime properties.
// This patch is used to update existing context.
func NewEstimationContextUpdate(similarNodeGroups []cloudprovider.NodeGroup, currentNodeCount int) *EstimationContext {
	return &EstimationContext{
		similarNodeGroups: similarNodeGroups,
		currentNodeCount:  currentNodeCount,
	}
}

// GetSimilarNodeGroups returns array of similar node groups
func (c *EstimationContext) GetSimilarNodeGroups() []cloudprovider.NodeGroup {
	return c.similarNodeGroups
}

// GetClusterMaxNodeLimit returns maximum node number allowed for the cluster
func (c *EstimationContext) GetClusterMaxNodeLimit() int {
	return c.clusterMaxNodeLimit
}

// GetCurrentNodeCount returns current number of nodes in the cluster
func (c *EstimationContext) GetCurrentNodeCount() int {
	return c.currentNodeCount
}

func updateContext(baseContext *EstimationContext, otherContext *EstimationContext) *EstimationContext {
	if baseContext == nil {
		return otherContext
	} else if otherContext == nil {
		return baseContext
	}
	baseContext.similarNodeGroups = otherContext.similarNodeGroups
	baseContext.currentNodeCount = otherContext.currentNodeCount
	return baseContext
}
