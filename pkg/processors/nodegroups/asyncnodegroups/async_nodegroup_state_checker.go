/*
Copyright 2024 The Kubernetes Authors.

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

package asyncnodegroups

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// AsyncNodeGroupStateChecker is responsible for checking the state of a node group
type AsyncNodeGroupStateChecker interface {
	// IsUpcoming checks if the node group is being asynchronously created, is scheduled to be
	// asynchronously created or is being initiated after asynchronous creation. Upcoming node groups
	// are reported as non-existing by the NodeGroup.Exist method, but are listed by the cloud provider.
	// Upcoming node group may be scaled up or down, if cloud provider supports in-memory accounting.
	// When cloud provider does not support asynchronous node group creation,
	// method always return false.
	IsUpcoming(nodeGroup cloudprovider.NodeGroup) bool

	CleanUp()
}

// NoOpAsyncNodeGroupStateChecker is a no-op implementation of AsyncNodeGroupStateChecker.
type NoOpAsyncNodeGroupStateChecker struct {
}

// IsUpcoming returns false by default
func (*NoOpAsyncNodeGroupStateChecker) IsUpcoming(nodeGroup cloudprovider.NodeGroup) bool {
	return false
}

// CleanUp cleans up internal structures.
func (*NoOpAsyncNodeGroupStateChecker) CleanUp() {
}

// MockAsyncNodeGroupStateChecker is a mock AsyncNodeGroupStateChecker to be used in tests
type MockAsyncNodeGroupStateChecker struct {
	IsUpcomingNodeGroup map[string]bool
}

// IsUpcoming simulates checking if node group is upcoming.
func (p *MockAsyncNodeGroupStateChecker) IsUpcoming(nodeGroup cloudprovider.NodeGroup) bool {
	return p.IsUpcomingNodeGroup[nodeGroup.Id()]
}

// CleanUp doesn't do anything; it's here to satisfy the interface
func (p *MockAsyncNodeGroupStateChecker) CleanUp() {
}

// NewDefaultAsyncNodeGroupStateChecker creates an instance of AsyncNodeGroupStateChecker.
func NewDefaultAsyncNodeGroupStateChecker() AsyncNodeGroupStateChecker {
	return &NoOpAsyncNodeGroupStateChecker{}
}
