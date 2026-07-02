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
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/fake"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

// RunTestClusterStateRegistryScaleUpWithDeletedNodes runs an integration test validating that ClusterStateRegistry correctly handles scaling up while deleted Nodes
// from a previous scale-down are still hanging around in the K8s API. The provided setupFactory is used to create the StaticAutoscaler, the K8s fake, and the
// test NodeGroup used by the test logic. This setup allows running the test against any CloudProvider implementation that works with the integration testing framework.
func RunTestClusterStateRegistryScaleUpWithDeletedNodes(t *testing.T, setupFactory TestClusterStateRegistryScaleUpWithDeletedNodesSetupFactory) {
	// This is a regression test for a complex race-condition bug in ClusterStateRegistry.
	// Full details about the bug can be found in https://github.com/kubernetes/autoscaler/issues/9813

	sdUnneededTime := 1 * time.Minute // arbitrary
	stepDuration := 10 * time.Second  // arbitrary
	stepsBetweenScaleDownAndGarbageCollection := 3
	stepsBetweenScaleUpAndNodeRegistration := 3
	nodeGarbageCollectionDelay := time.Duration(stepsBetweenScaleDownAndGarbageCollection) * stepDuration
	nodeRegistrationDelay := time.Duration(stepsBetweenScaleUpAndNodeRegistration) * stepDuration
	// maxNodeProvisionTime is intentionally higher than nodeRegistrationDelay so that CSR never hits a timeout when waiting for the upcoming Nodes.
	maxNodeProvisionTime := nodeRegistrationDelay + stepDuration

	initialNodeCount := 10
	fullNodesCount := 6

	// These arguments are passed to the provided setupFactory, to be used when setting up StaticAutoscaler and the CloudProvider implementation.
	setupFactoryArgs := TestClusterStateRegistryScaleUpWithDeletedNodesSetupArgs{
		// Configure the AutoscalingOptions so that scale-down of empty Nodes actually happens, and set some delays to particular values
		// that are important for the test logic below.
		OptsOverride: func(opts *config.AutoscalingOptions) {
			opts.ScaleDownEnabled = true
			opts.ScaleDownDelayAfterAdd = 0
			opts.ScaleDownSimulationTimeout = 24 * time.Hour
			opts.MaxScaleDownParallelism = 10
			opts.NodeGroupDefaults.ScaleDownUtilizationThreshold = 0.5
			opts.NodeGroupDefaults.ScaleDownUnneededTime = sdUnneededTime
			opts.NodeGroupDefaults.MaxNodeProvisionTime = maxNodeProvisionTime
		},
		// Configure Node K8s object delays, to be simulated by the API fake used by the CloudProvider implementation.
		NodeGarbageCollectionDelay: nodeGarbageCollectionDelay,
		NodeRegistrationDelay:      nodeRegistrationDelay,
		// Configure how many initial Nodes the factory should create in the returned NodeGroup.
		NodeCount: initialNodeCount,
	}

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		autoscaler, nodeGroup, k8s := setupFactory(t, ctx, setupFactoryArgs)
		// Progress the time by nodeRegistrationDelay to make sure the initial Nodes created by the setup factory appear in the K8s fake.
		time.Sleep(nodeRegistrationDelay)
		// Validate that the initial state of the NodeGroup is as expected.
		assertNodeGroupSize(t, nodeGroup, k8s, initialNodeCount, initialNodeCount)
		// Find out how much CPU and memory is available on the Nodes created by the setup factory. The assertion above ensures that the Node
		// list is not empty.
		nodeCpu, nodeMemory := getNodeCpuAndMemory(&k8s.Nodes().Items[0])
		fullCpuRequest := int64(float64(nodeCpu) * 0.8)
		fullMemRequest := int64(float64(nodeMemory) * 0.8)

		// Create scheduled Pods to block a portion of the initial Nodes from being scaled down. Each Pod requires a dedicated Node, and blocks CA from scaling it down.
		for i := range fullNodesCount {
			node := k8s.Nodes().Items[i]
			pod := testutils.BuildTestPod(fmt.Sprintf("scheduled-pods-%d", i), fullCpuRequest, fullMemRequest, testutils.WithNodeName(node.Name))
			k8s.AddPod(pod)
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
		assertNodeGroupSize(t, nodeGroup, k8s, fullNodesCount, initialNodeCount)

		// ======== STEP 2: Force CA to scale up some Nodes in the same NodeGroup ========
		// Create some pending Pods so that CA needs to scale the same NodeGroup up still immediately after a previous scale-down.
		pendingPodsCount := 2
		for i := range pendingPodsCount {
			podName := fmt.Sprintf("pending-pods-%d", i)
			pod := testutils.BuildTestPod(podName, fullCpuRequest, fullMemRequest, testutils.MarkUnschedulable())
			k8s.AddPod(pod)
		}
		// Run CA loop with the new pending Pods, CA should scale up a dedicated Node for each pending Pod.
		if err := synctestutils.RunOnceAfter(t, autoscaler, stepDuration); err != nil {
			t.Fatalf("RunOnce() unexpected error: %v", err)
		}
		// RunOnceAfter() above only returns after all goroutines in the bubble are blocked, so we're guaranteed that IncreaseSize() already increased the
		// target size of the NodeGroup back to the initial count. We're also guaranteed that CA loops will not see the new Nodes for the next nodeGarbageCollectionDelay,
		// so upcoming Nodes will have to be modeled.
		assertNodeGroupSize(t, nodeGroup, k8s, fullNodesCount+pendingPodsCount, initialNodeCount)

		// ======== STEP 3: Assert that CA doesn't scale up again for the same Pods ========
		// Run CA loop after a short delay. The deleted Nodes should still be present, and the new Nodes still shouldn't.
		if err := synctestutils.RunOnceAfter(t, autoscaler, stepDuration); err != nil {
			t.Fatalf("RunOnce() unexpected error: %v", err)
		}
		// CA shouldn't scale up, because the pending Pods should be packed on upcoming Nodes from the previous scale-up. The sizes should be identical to the previous step.
		assertNodeGroupSize(t, nodeGroup, k8s, fullNodesCount+pendingPodsCount, initialNodeCount)

		// ======== STEP 4: Assert that deleted Nodes disappear from the API at the expected time ========
		// Run CA loop after nodeGarbageCollectionDelay (3 times stepDuration) elapses since the scale-down.
		if err := synctestutils.RunOnceAfter(t, autoscaler, stepDuration); err != nil {
			t.Fatalf("RunOnce() unexpected error: %v", err)
		}
		// The Nodes should be gone from the API, so we should only see the original Nodes that had Pods scheduled. No changes to the targetSize are expected.
		assertNodeGroupSize(t, nodeGroup, k8s, fullNodesCount+pendingPodsCount, fullNodesCount)

		// ======== STEP 5: Assert that scaled-up Nodes appear in the API at the expected time ========
		// Run CA loop after nodeRegistrationDelay elapses since the scale-up.
		if err := synctestutils.RunOnceAfter(t, autoscaler, stepDuration+time.Microsecond); err != nil {
			t.Fatalf("RunOnce() unexpected error: %v", err)
		}
		// The new Nodes should finally be visible in the API. No changes to the targetSize are expected.
		assertNodeGroupSize(t, nodeGroup, k8s, fullNodesCount+pendingPodsCount, fullNodesCount+pendingPodsCount)

		// ======== STEP 6: Assert that the final state is stable  ========
		// Run CA loop after a long delay to verify that the final state is stable.
		if err := synctestutils.RunOnceAfter(t, autoscaler, 100*stepDuration); err != nil {
			t.Fatalf("RunOnce() unexpected error: %v", err)
		}
		// Sizes should be identical to the previous step.
		assertNodeGroupSize(t, nodeGroup, k8s, fullNodesCount+pendingPodsCount, fullNodesCount+pendingPodsCount)
	})
}

// TestClusterStateRegistryScaleUpWithDeletedNodesSetupArgs groups arguments to TestClusterStateRegistryScaleUpWithDeletedNodesSetupFactory.
type TestClusterStateRegistryScaleUpWithDeletedNodesSetupArgs struct {
	// OptsOverride should be applied to the AutoscalingOptions used by the created Autoscaler.
	OptsOverride integration.AutoscalingOptionOverride
	// NodeRegistrationDelay is the delay that Nodes should appear after in the K8s fake, since the NodeGroup.IncreaseSize() call. This behavior should likely be
	// configured in the fake implementation of the API used to create new VMs in NodeGroup.IncreaseSize().
	NodeRegistrationDelay time.Duration
	// NodeRegistrationDelay is the delay that Nodes should disappear after from the K8s fake, since the NodeGroup.DeleteNodes() call. This behavior should likely be
	// configured in the fake implementation of the API used to delete VMs in NodeGroup.DeleteNodes().
	NodeGarbageCollectionDelay time.Duration

	// NodeCount specifies how many initial Nodes there should be in the created NodeGroup. The Nodes in the NodeGroup should have a reasonable CPU and memory allocatable values,
	// and should otherwise allow Pods scheduling on them (should be Ready, shouldn't have any taints etc.).
	NodeCount int
}

// TestClusterStateRegistryScaleUpWithDeletedNodesSetupFactory creates the Autoscaler, the K8s fake, and the test NodeGroup to be used in the test scenario, based on the provided arguments.
// See the field comments in TestClusterStateRegistryScaleUpWithDeletedNodesSetupArgs to see how the arguments should be used during the creation.
// The provided context object should be propagated to any goroutines etc. that would need to be cleaned up when the test finishes and the context is cancelled. In particular, it should be
// propagated to the StaticAutoscaler Builder.
type TestClusterStateRegistryScaleUpWithDeletedNodesSetupFactory func(*testing.T, context.Context, TestClusterStateRegistryScaleUpWithDeletedNodesSetupArgs) (core.Autoscaler, cloudprovider.NodeGroup, *fakek8s.Kubernetes)

func assertNodeGroupSize(t *testing.T, ng cloudprovider.NodeGroup, k8s *fakek8s.Kubernetes, wantTargetSize, wantNodeCount int) {
	t.Helper()
	if gotSize, err := ng.TargetSize(); err != nil || wantTargetSize != gotSize {
		t.Fatalf("fakeNodeGroup.TargetSize(): want <%d, <nil>>, got <%d, %v>", wantTargetSize, gotSize, err)
	}
	if gotNodeCount := len(k8s.Nodes().Items); wantNodeCount != gotNodeCount {
		t.Fatalf("Fakes.K8s.Nodes(): want len=%d, got len=%d", wantNodeCount, gotNodeCount)
	}
}

func getNodeCpuAndMemory(node *apiv1.Node) (milliCpu int64, bytesMemory int64) {
	cpuQuantity := node.Status.Allocatable[apiv1.ResourceCPU]
	memQuantity := node.Status.Allocatable[apiv1.ResourceMemory]
	return cpuQuantity.MilliValue(), memQuantity.Value()
}
