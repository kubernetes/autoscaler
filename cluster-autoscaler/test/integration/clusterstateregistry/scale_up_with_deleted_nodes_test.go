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

package clusterstateregistry

import (
	"testing"
)

// TestClusterStateRegistryScaleUpWithDeletedNodes is an integration test validating that ClusterStateRegistry correctly handles scaling up while deleted Nodes
// from a previous scale-down are still hanging around in the K8s API.
func TestClusterStateRegistryScaleUpWithDeletedNodes(t *testing.T) {
	// This is a regression test for a complex race-condition bug in ClusterStateRegistry.
	// Full details about the bug can be found in https://github.com/kubernetes/autoscaler/issues/9813
	for _, tc := range []struct {
		testName string
		// nodeGroupForNodeWorksForDeletedNodes is used to test different supported behaviors of the CloudProvider.NodeGroupForNode() method.
		nodeGroupForNodeWorksForDeletedNodes bool
		// hasInstanceNotImplemented is used to test different supported behaviors of the CloudProvider.HasInstance() method.
		// Note that the behavior of HasInstance() returning still returning true for Nodes deleted via NodeGroup.DeleteNodes() is not supported
		// and this test would not pass with it.
		hasInstanceNotImplemented bool // The field is negated to match the parameter to ConfigureHasInstanceBehavior().
	}{
		{
			testName:                             "NodeGroupForNode_works_for_deleted_HasInstance_implemented",
			nodeGroupForNodeWorksForDeletedNodes: true,
			hasInstanceNotImplemented:            false,
		},
		{
			testName:                             "NodeGroupForNode_works_for_deleted_HasInstance_not_implemented",
			nodeGroupForNodeWorksForDeletedNodes: true,
			hasInstanceNotImplemented:            true,
		},
		{
			testName:                             "NodeGroupForNode_returns_nil_for_deleted_HasInstance_implemented",
			nodeGroupForNodeWorksForDeletedNodes: false,
			hasInstanceNotImplemented:            false,
		},
		{
			testName:                             "NodeGroupForNode_returns_nil_for_deleted_HasInstance_not_implemented",
			nodeGroupForNodeWorksForDeletedNodes: false,
			hasInstanceNotImplemented:            true,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			RunTestClusterStateRegistryScaleUpWithDeletedNodes(t, tc.nodeGroupForNodeWorksForDeletedNodes, tc.hasInstanceNotImplemented)
		})
	}
}
