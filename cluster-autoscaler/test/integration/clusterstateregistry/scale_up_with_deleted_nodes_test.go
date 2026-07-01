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
	"context"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/fake"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

// TestClusterStateRegistryScaleUpWithDeletedNodes is an integration test validating that ClusterStateRegistry correctly handles scaling up while deleted Nodes
// from a previous scale-down are still hanging around in the K8s API. The test uses the fake CloudProvider with different semantics of certain methods being
// tested.
func TestClusterStateRegistryScaleUpWithDeletedNodes(t *testing.T) {
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
			autoscalerFactory := func(t *testing.T, ctx context.Context, args TestClusterStateRegistryScaleUpWithDeletedNodesSetupArgs) (core.Autoscaler, cloudprovider.NodeGroup, *fakek8s.Kubernetes) {
				options := integration.NewTestConfig().WithOverrides(args.OptsOverride).ResolveOptions()
				infra := integration.SetupInfrastructure(t)
				// Configure the behavior of the fake CloudProvider methods according to the test case.
				infra.Fakes.CloudProvider.ConfigureNodeGroupForNodeBehavior(tc.nodeGroupForNodeWorksForDeletedNodes)
				infra.Fakes.CloudProvider.ConfigureHasInstanceBehavior(tc.hasInstanceNotImplemented)

				autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
				if err != nil {
					t.Fatalf("DefaultAutoscalingBuilder() unexpected error: %v", err)
				}

				templateNode := testutils.BuildTestNode("ng-template", 1000, 2*units.GiB, testutils.IsReady(true))
				nodeGroup := infra.Fakes.CloudProvider.AddNodeGroup("ng",
					fakecloudprovider.WithNodes(templateNode, args.NodeCount),
					fakecloudprovider.WithNodeGarbageCollectionDelay(args.NodeGarbageCollectionDelay),
					fakecloudprovider.WithNodeRegistrationDelay(args.NodeRegistrationDelay),
				)

				return autoscaler, nodeGroup, infra.Fakes.K8s
			}

			RunTestClusterStateRegistryScaleUpWithDeletedNodes(t, autoscalerFactory)
		})
	}
}
