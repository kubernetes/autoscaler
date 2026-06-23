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

package inmemory

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/fake"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
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
			sdUnneededTime := 1 * time.Minute // arbitrary
			stepDuration := 10 * time.Second  // arbitrary
			stepsBetweenScaleDownAndGarbageCollection := 3
			stepsBetweenScaleUpAndNodeRegistration := 3
			nodeGarbageCollectionDelay := time.Duration(stepsBetweenScaleDownAndGarbageCollection) * stepDuration
			nodeRegistrationDelay := time.Duration(stepsBetweenScaleUpAndNodeRegistration) * stepDuration
			// maxNodeProvisionTime is intentionally higher than nodeRegistrationDelay so that CSR never hits a timeout when waiting for the upcoming Nodes.
			maxNodeProvisionTime := nodeRegistrationDelay + stepDuration

			options := integration.NewTestConfig().WithOverrides(func(opts *config.AutoscalingOptions) {
				opts.NodeGroupDefaults.ScaleDownUnneededTime = sdUnneededTime
				opts.NodeGroupDefaults.MaxNodeProvisionTime = maxNodeProvisionTime
			}).ResolveOptions()

			infra := integration.SetupInfrastructure(t)
			// Configure the behavior of the fake CloudProvider methods according to the test case.
			infra.Fakes.CloudProvider.ConfigureNodeGroupForNodeBehavior(tc.nodeGroupForNodeWorksForDeletedNodes)
			infra.Fakes.CloudProvider.ConfigureHasInstanceBehavior(tc.hasInstanceNotImplemented)

			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer synctestutils.TearDown(cancel)

				autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
				if err != nil {
					t.Fatalf("DefaultAutoscalingBuilder() unexpected error: %v", err)
				}

				// ======== STEP 0: Prepare the initial state of the test NodeGroup ========
				initialNodeCount := 10
				ngName := "ng"
				templateNode := testutils.BuildTestNode("ng-template", 1000, 2*units.GiB, testutils.IsReady(true))
				// The test NodeGroup is configured with artificial delays simulating real-world conditions that have impact on ClusterStateRegistry logic:
				// - Node objects are only deleted from the K8s API fake after nodeGarbageCollectionDelay elapses since the DeleteNodes() call. This is to simulate Node objects
				//   hanging around for some time after the VM deletion call, and being processed by subsequent CA loops.
				// - Node objects are only added to the K8s API fake after nodeRegistrationDelay elapses since the IncreaseSize() call. This is to simulate a delay between VM creation
				//   call and the VM registering as Node in K8s. During this delay, CA needs to model upcoming Nodes correctly in order not to scale-up again unnecessarily.
				fakeNodeGroup := infra.Fakes.CloudProvider.AddNodeGroup(ngName,
					fakecloudprovider.WithNodes(templateNode, initialNodeCount),
					fakecloudprovider.WithNodeGarbageCollectionDelay(nodeGarbageCollectionDelay),
					fakecloudprovider.WithNodeRegistrationDelay(nodeRegistrationDelay),
				)
				// Create scheduled Pods to block a portion of the initial Nodes from being scaled down. Each Pod requires a dedicated Node, and blocks CA from scaling it down.
				fullNodesCount := 6
				for i := range fullNodesCount {
					podName := fmt.Sprintf("scheduled-pods-%d", i)
					nodeName := fmt.Sprintf("%s-node-%d", ngName, i) // fakecloudprovider.WithNodes() names Nodes in this format.
					pod := testutils.BuildTestPod(podName, 800, 1*units.GiB, testutils.WithNodeName(nodeName))
					infra.Fakes.K8s.AddPod(pod)
				}

				// ======== STEP 1: Force CA to scale down some empty Nodes ========
				// The remaining Nodes from the NodeGroup that don't have a Pod scheduled are empty. Run CA loop once to start tracking them as unneeded.
				if err := synctestutils.RunOnceAfter(t, autoscaler, 0); err != nil {
					t.Fatalf("RunOnce() unexpected error: %v", err)
				}
				// Run CA loop again after the unneeded threshold has passed, CA should start scaling down the empty Nodes.
				if err := synctestutils.RunOnceAfter(t, autoscaler, sdUnneededTime+time.Microsecond); err != nil {
					t.Fatalf("RunOnce() unexpected error: %v", err)
				}
				// RunOnceAfter() above only returns after all goroutines in the bubble are blocked, so we're guaranteed that DeleteNodes() already decreased the
				// target size of the NodeGroup. We're also guaranteed that CA loops will still process the deleted Nodes for the next nodeGarbageCollectionDelay.
				assertNodeGroupSize(t, fakeNodeGroup, infra.Fakes.K8s, fullNodesCount, initialNodeCount)

				// ======== STEP 2: Force CA to scale up some Nodes in the same NodeGroup ========
				// Create some pending Pods so that CA needs to scale the same NodeGroup up still immediately after a previous scale-down.
				pendingPodsCount := 2
				for i := range pendingPodsCount {
					podName := fmt.Sprintf("pending-pods-%d", i)
					pod := testutils.BuildTestPod(podName, 800, 1*units.GiB, testutils.MarkUnschedulable())
					infra.Fakes.K8s.AddPod(pod)
				}
				// Run CA loop with the new pending Pods, CA should scale up a dedicated Node for each pending Pod.
				if err := synctestutils.RunOnceAfter(t, autoscaler, stepDuration); err != nil {
					t.Fatalf("RunOnce() unexpected error: %v", err)
				}
				// RunOnceAfter() above only returns after all goroutines in the bubble are blocked, so we're guaranteed that IncreaseSize() already increased the
				// target size of the NodeGroup back to the initial count. We're also guaranteed that CA loops will not see the new Nodes for the next nodeGarbageCollectionDelay,
				// so upcoming Nodes will have to be modeled.
				assertNodeGroupSize(t, fakeNodeGroup, infra.Fakes.K8s, fullNodesCount+pendingPodsCount, initialNodeCount)

				// ======== STEP 3: Assert that CA doesn't scale up again for the same Pods ========
				// Run CA loop after a short delay. The deleted Nodes should still be present, and the new Nodes still shouldn't.
				if err := synctestutils.RunOnceAfter(t, autoscaler, stepDuration); err != nil {
					t.Fatalf("RunOnce() unexpected error: %v", err)
				}
				// CA shouldn't scale up, because the pending Pods should be packed on upcoming Nodes from the previous scale-up. The sizes should be identical to the previous step.
				assertNodeGroupSize(t, fakeNodeGroup, infra.Fakes.K8s, fullNodesCount+pendingPodsCount, initialNodeCount)

				// ======== STEP 4: Assert that deleted Nodes disappear from the API at the expected time ========
				// Run CA loop after nodeGarbageCollectionDelay (3 times stepDuration) elapses since the scale-down.
				if err := synctestutils.RunOnceAfter(t, autoscaler, stepDuration); err != nil {
					t.Fatalf("RunOnce() unexpected error: %v", err)
				}
				// The Nodes should be gone from the API, so we should only see the original Nodes that had Pods scheduled. No changes to the targetSize are expected.
				assertNodeGroupSize(t, fakeNodeGroup, infra.Fakes.K8s, fullNodesCount+pendingPodsCount, fullNodesCount)

				// ======== STEP 5: Assert that scaled-up Nodes appear in the API at the expected time ========
				// Run CA loop after nodeRegistrationDelay elapses since the scale-up.
				if err := synctestutils.RunOnceAfter(t, autoscaler, stepDuration+time.Microsecond); err != nil {
					t.Fatalf("RunOnce() unexpected error: %v", err)
				}
				// The new Nodes should finally be visible in the API. No changes to the targetSize are expected.
				assertNodeGroupSize(t, fakeNodeGroup, infra.Fakes.K8s, fullNodesCount+pendingPodsCount, fullNodesCount+pendingPodsCount)

				// ======== STEP 6: Assert that the final state is stable  ========
				// Run CA loop after a long delay to verify that the final state is stable.
				if err := synctestutils.RunOnceAfter(t, autoscaler, 100*stepDuration); err != nil {
					t.Fatalf("RunOnce() unexpected error: %v", err)
				}
				// Sizes should be identical to the previous step.
				assertNodeGroupSize(t, fakeNodeGroup, infra.Fakes.K8s, fullNodesCount+pendingPodsCount, fullNodesCount+pendingPodsCount)
			})
		})
	}
}

func assertNodeGroupSize(t *testing.T, ng *fakecloudprovider.NodeGroup, k8s *fakek8s.Kubernetes, wantTargetSize, wantNodeCount int) {
	t.Helper()
	if gotSize, err := ng.TargetSize(); err != nil || wantTargetSize != gotSize {
		t.Fatalf("fakeNodeGroup.TargetSize(): want <%d, <nil>>, got <%d, %v>", wantTargetSize, gotSize, err)
	}
	if gotNodeCount := len(k8s.Nodes().Items); wantNodeCount != gotNodeCount {
		t.Fatalf("Fakes.K8s.Nodes(): want len=%d, got len=%d", wantNodeCount, gotNodeCount)
	}
}
