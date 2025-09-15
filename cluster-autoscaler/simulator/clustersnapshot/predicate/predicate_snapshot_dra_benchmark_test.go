/*
Copyright 2025 The Kubernetes Authors.

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

package predicate

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/util/feature"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	featuretesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/kubernetes/pkg/features"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func createTestResourceSlice(nodeName string, devicesPerSlice int, slicesPerNode int, driver string) *resourceapi.ResourceSlice {
	sliceId := uuid.New().String()
	name := fmt.Sprintf("rs-%s", sliceId)
	uid := types.UID(fmt.Sprintf("rs-%s-uid", sliceId))
	devices := make([]resourceapi.Device, devicesPerSlice)
	for deviceIndex := 0; deviceIndex < devicesPerSlice; deviceIndex++ {
		deviceName := fmt.Sprintf("rs-dev-%s-%d", sliceId, deviceIndex)
		devices[deviceIndex] = resourceapi.Device{Name: deviceName}
	}

	return &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: uid},
		Spec: resourceapi.ResourceSliceSpec{
			NodeName: &nodeName,
			Driver:   driver,
			Pool: resourceapi.ResourcePool{
				Name:               nodeName,
				ResourceSliceCount: int64(slicesPerNode),
			},
			Devices: devices,
		},
	}
}

func createTestResourceClaim(requestsPerClaim int, devicesPerRequest int, driver string, deviceClass string) *resourceapi.ResourceClaim {
	claimId := uuid.New().String()
	name := fmt.Sprintf("claim-%s", claimId)
	uid := types.UID(fmt.Sprintf("claim-%s-uid", claimId))
	expression := fmt.Sprintf(`device.driver == "%s"`, driver)

	requests := make([]resourceapi.DeviceRequest, requestsPerClaim)
	for requestIndex := 0; requestIndex < requestsPerClaim; requestIndex++ {
		requests[requestIndex] = resourceapi.DeviceRequest{
			Name: fmt.Sprintf("deviceRequest-%d", requestIndex),
			Exactly: &resourceapi.ExactDeviceRequest{
				DeviceClassName: deviceClass,
				Selectors:       []resourceapi.DeviceSelector{{CEL: &resourceapi.CELDeviceSelector{Expression: expression}}},
				AllocationMode:  resourceapi.DeviceAllocationModeExactCount,
				Count:           int64(devicesPerRequest),
			},
		}
	}

	return &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name, UID: uid, Namespace: "default"},
		Spec: resourceapi.ResourceClaimSpec{
			Devices: resourceapi.DeviceClaim{Requests: requests},
		},
	}
}

// allocateResourceSlicesForClaim attempts to allocate devices from the provided ResourceSlices
// to satisfy the requests in the given ResourceClaim. It iterates through the claim's device
// requests and, for each request, tries to find enough available devices in the provided slices.
//
// The function returns a new ResourceClaim object with the allocation result (if successful)
// and a boolean indicating whether all requests in the claim were satisfied.
//
// If not all requests can be satisfied with the given slices, the returned ResourceClaim will
// have a partial or empty allocation, and the boolean will be false.
// The original ResourceClaim object is not modified.
func allocateResourceSlicesForClaim(claim *resourceapi.ResourceClaim, nodeName string, slices ...*resourceapi.ResourceSlice) (*resourceapi.ResourceClaim, bool) {
	allocatedDevices := make([]resourceapi.DeviceRequestAllocationResult, 0, len(claim.Spec.Devices.Requests))
	sliceIndex, deviceIndex := 0, 0
	requestSatisfied := true

allocationLoop:
	for _, request := range claim.Spec.Devices.Requests {
		for devicesRequired := request.Exactly.Count; devicesRequired > 0; devicesRequired-- {
			// Skipping resource slices until we find one with at least a single device available
			for sliceIndex < len(slices) && deviceIndex >= len(slices[sliceIndex].Spec.Devices) {
				sliceIndex++
				deviceIndex = 0
			}

			// In case in the previous look we weren't able to find a resource slice containing
			// at least a single device and there's still device request pending from resource
			// claim - terminate allocation loop and indicate the request wasn't fully satisfied
			if sliceIndex >= len(slices) {
				requestSatisfied = false
				break allocationLoop
			}

			slice := slices[sliceIndex]
			device := slice.Spec.Devices[deviceIndex]
			deviceAllocation := resourceapi.DeviceRequestAllocationResult{
				Request: request.Name,
				Driver:  slice.Spec.Driver,
				Pool:    slice.Spec.Pool.Name,
				Device:  device.Name,
			}

			allocatedDevices = append(allocatedDevices, deviceAllocation)
			deviceIndex++
		}
	}

	allocation := &resourceapi.AllocationResult{
		NodeSelector: selectorForNode(nodeName),
		Devices:      resourceapi.DeviceAllocationResult{Results: allocatedDevices},
	}

	return drautils.TestClaimWithAllocation(claim, allocation), requestSatisfied
}

func selectorForNode(node string) *apiv1.NodeSelector {
	return &apiv1.NodeSelector{
		NodeSelectorTerms: []apiv1.NodeSelectorTerm{
			{
				MatchFields: []apiv1.NodeSelectorRequirement{
					{
						Key:      "metadata.name",
						Operator: apiv1.NodeSelectorOpIn,
						Values:   []string{node},
					},
				},
			},
		},
	}
}

// BenchmarkScheduleRevert measures the performance of scheduling pods which interact with Dynamic Resources Allocation
// API onto nodes within a cluster snapshot, followed by snapshot manipulation operations (fork, commit, revert).
//
// The benchmark iterates tests for various configurations, varying:
// - The number of nodes in the initial snapshot.
// - The number of pods being scheduled, categorized by whether they use shared or pod-owned ResourceClaims.
// - The number of snapshot operations (Fork, Commit, Revert) performed before/after scheduling.
//
// For each configuration and snapshot type, the benchmark performs the following steps:
//  1. Initializes a cluster snapshot with a predefined set of nodes, ResourceSlices, DeviceClasses, and pre-allocated ResourceClaims (both shared and potentially pod-owned).
//  2. Iterates through a subset of the nodes based on the configuration.
//  3. For each node:
//     a. Performs the configured number of snapshot Forks.
//     b. Adds the node's NodeInfo (including its ResourceSlices) to the snapshot.
//     c. Schedules a configured number of pods that reference a shared ResourceClaim onto the node.
//     d. Schedules a configured number of pods that reference their own pre-allocated pod-owned ResourceClaims onto the node.
//     e. Performs the configured number of snapshot Commits.
//     f. Performs the configured number of snapshot Reverts.
//
// This benchmark helps evaluate the efficiency of:
// - Scheduling pods with different types of DRA claims.
// - Adding nodes with DRA resources to the snapshot.
// - The overhead of snapshot Fork, Commit, and Revert operations, especially in scenarios involving DRA objects.
func BenchmarkScheduleRevert(b *testing.B) {
	featuretesting.SetFeatureGateDuringTest(b, feature.DefaultFeatureGate, features.DynamicResourceAllocation, true)

	const maxNodesCount = 100
	const devicesPerSlice = 100
	const maxPodsCount = 100
	const deviceClassName = "defaultClass"
	const driverName = "driver.foo.com"

	configurations := map[string]struct {
		nodesCount int

		sharedClaimPods int
		ownedClaimPods  int
		forks           int
		commits         int
		reverts         int
	}{
		// SHARED CLAIMS
		"100x32/SharedClaims/ForkRevert":           {sharedClaimPods: 32, nodesCount: 100, forks: 1, reverts: 1},
		"100x32/SharedClaims/ForkCommit":           {sharedClaimPods: 32, nodesCount: 100, forks: 1, commits: 1},
		"100x32/SharedClaims/ForkForkCommitRevert": {sharedClaimPods: 32, nodesCount: 100, forks: 2, reverts: 1, commits: 1},
		"100x32/SharedClaims/Fork":                 {sharedClaimPods: 32, nodesCount: 100, forks: 1},
		"100x32/SharedClaims/Fork5Revert5":         {sharedClaimPods: 32, nodesCount: 100, forks: 5, reverts: 5},
		"100x1/SharedClaims/ForkRevert":            {sharedClaimPods: 1, nodesCount: 100, forks: 1, reverts: 1},
		"100x1/SharedClaims/ForkCommit":            {sharedClaimPods: 1, nodesCount: 100, forks: 1, commits: 1},
		"100x1/SharedClaims/ForkForkCommitRevert":  {sharedClaimPods: 1, nodesCount: 100, forks: 2, reverts: 1, commits: 1},
		"100x1/SharedClaims/Fork":                  {sharedClaimPods: 1, nodesCount: 100, forks: 1},
		"100x1/SharedClaims/Fork5Revert5":          {sharedClaimPods: 1, nodesCount: 100, forks: 5, reverts: 5},
		"10x32/SharedClaims/ForkRevert":            {sharedClaimPods: 32, nodesCount: 10, forks: 1, reverts: 1},
		"10x32/SharedClaims/ForkCommit":            {sharedClaimPods: 32, nodesCount: 10, forks: 1, commits: 1},
		"10x32/SharedClaims/ForkForkCommitRevert":  {sharedClaimPods: 32, nodesCount: 10, forks: 2, reverts: 1, commits: 1},
		"10x32/SharedClaims/Fork":                  {sharedClaimPods: 32, nodesCount: 10, forks: 1},
		"10x32/SharedClaims/Fork5Revert5":          {sharedClaimPods: 32, nodesCount: 10, forks: 5, reverts: 5},
		"10x1/SharedClaims/ForkRevert":             {sharedClaimPods: 1, nodesCount: 10, forks: 1, reverts: 1},
		"10x1/SharedClaims/ForkCommit":             {sharedClaimPods: 1, nodesCount: 10, forks: 1, commits: 1},
		"10x1/SharedClaims/ForkForkCommitRevert":   {sharedClaimPods: 1, nodesCount: 10, forks: 2, reverts: 1, commits: 1},
		"10x1/SharedClaims/Fork":                   {sharedClaimPods: 1, nodesCount: 10, forks: 1},
		"10x1/SharedClaims/Fork5Revert5":           {sharedClaimPods: 1, nodesCount: 10, forks: 5, reverts: 5},
		// POD OWNED CLAIMS
		"100x100/OwnedClaims/ForkRevert":           {ownedClaimPods: 100, nodesCount: 100, forks: 1, reverts: 1},
		"100x100/OwnedClaims/ForkCommit":           {ownedClaimPods: 100, nodesCount: 100, forks: 1, commits: 1},
		"100x100/OwnedClaims/ForkForkCommitRevert": {ownedClaimPods: 100, nodesCount: 100, forks: 2, reverts: 1, commits: 1},
		"100x100/OwnedClaims/Fork":                 {ownedClaimPods: 100, nodesCount: 100, forks: 1},
		"100x100/OwnedClaims/Fork5Revert5":         {ownedClaimPods: 100, nodesCount: 100, forks: 5, reverts: 5},
		"100х1/OwnedClaims/ForkRevert":             {ownedClaimPods: 1, nodesCount: 100, forks: 1, reverts: 1},
		"100x1/OwnedClaims/ForkCommit":             {ownedClaimPods: 1, nodesCount: 100, forks: 1, commits: 1},
		"100x1/OwnedClaims/ForkForkCommitRevert":   {ownedClaimPods: 1, nodesCount: 100, forks: 2, reverts: 1, commits: 1},
		"100x1/OwnedClaims/Fork":                   {ownedClaimPods: 1, nodesCount: 100, forks: 1},
		"100x1/OwnedClaims/Fork5Revert5":           {ownedClaimPods: 1, nodesCount: 100, forks: 5, reverts: 5},
		"10x100/OwnedClaims/ForkRevert":            {ownedClaimPods: 100, nodesCount: 10, forks: 1, reverts: 1},
		"10x100/OwnedClaims/ForkCommit":            {ownedClaimPods: 100, nodesCount: 10, forks: 1, commits: 1},
		"10x100/OwnedClaims/ForkForkCommitRevert":  {ownedClaimPods: 100, nodesCount: 10, forks: 2, reverts: 1, commits: 1},
		"10x100/OwnedClaims/Fork":                  {ownedClaimPods: 100, nodesCount: 10, forks: 1},
		"10x100/OwnedClaims/Fork5Revert5":          {ownedClaimPods: 100, nodesCount: 10, forks: 5, reverts: 5},
		"10х1/OwnedClaims/ForkRevert":              {ownedClaimPods: 1, nodesCount: 10, forks: 1, reverts: 1},
		"10x1/OwnedClaims/ForkCommit":              {ownedClaimPods: 1, nodesCount: 10, forks: 1, commits: 1},
		"10x1/OwnedClaims/ForkForkCommitRevert":    {ownedClaimPods: 1, nodesCount: 10, forks: 2, reverts: 1, commits: 1},
		"10x1/OwnedClaims/Fork":                    {ownedClaimPods: 1, nodesCount: 10, forks: 1},
		"10x1/OwnedClaims/Fork5Revert5":            {ownedClaimPods: 1, nodesCount: 10, forks: 5, reverts: 5},
		// MIXED CLAIMS
		"100x32x50/MixedClaims/ForkRevert":           {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 100, forks: 1, reverts: 1},
		"100x32x50/MixedClaims/ForkCommit":           {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 100, forks: 1, commits: 1},
		"100x32x50/MixedClaims/ForkForkCommitRevert": {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 100, forks: 2, reverts: 1, commits: 1},
		"100x32x50/MixedClaims/Fork":                 {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 100, forks: 1},
		"100x32x50/MixedClaims/Fork5Revert5":         {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 100, forks: 5, reverts: 5},
		"100x1x1/MixedClaims/ForkRevert":             {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 100, forks: 1, reverts: 1},
		"100x1x1/MixedClaims/ForkCommit":             {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 100, forks: 1, commits: 1},
		"100x1x1/MixedClaims/ForkForkCommitRevert":   {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 100, forks: 2, reverts: 1, commits: 1},
		"100x1x1/MixedClaims/Fork":                   {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 100, forks: 1},
		"100x1x1/MixedClaims/Fork5Revert5":           {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 100, forks: 5, reverts: 5},
		"10x32x50/MixedClaims/ForkRevert":            {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 10, forks: 1, reverts: 1},
		"10x32x50/MixedClaims/ForkCommit":            {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 10, forks: 1, commits: 1},
		"10x32x50/MixedClaims/ForkForkCommitRevert":  {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 10, forks: 2, reverts: 1, commits: 1},
		"10x32x50/MixedClaims/Fork":                  {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 10, forks: 1},
		"10x32x50/MixedClaims/Fork5Revert5":          {sharedClaimPods: 32, ownedClaimPods: 50, nodesCount: 10, forks: 5, reverts: 5},
		"10x1x1/MixedClaims/ForkRevert":              {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 10, forks: 1, reverts: 1},
		"10x1x1/MixedClaims/ForkCommit":              {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 10, forks: 1, commits: 1},
		"10x1x1/MixedClaims/ForkForkCommitRevert":    {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 10, forks: 2, reverts: 1, commits: 1},
		"10x1x1/MixedClaims/Fork":                    {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 10, forks: 1},
		"10x1x1/MixedClaims/Fork5Revert5":            {sharedClaimPods: 1, ownedClaimPods: 1, nodesCount: 10, forks: 5, reverts: 5},
	}

	devicesClasses := map[string]*resourceapi.DeviceClass{
		deviceClassName: {ObjectMeta: metav1.ObjectMeta{Name: deviceClassName, UID: "defaultClassUid"}},
	}

	nodeInfos := make([]*framework.NodeInfo, maxNodesCount)
	sharedClaims := make([]*resourceapi.ResourceClaim, maxNodesCount)
	ownedClaims := make([][]*resourceapi.ResourceClaim, maxNodesCount)
	owningPods := make([][]*apiv1.Pod, maxNodesCount)
	for nodeIndex := 0; nodeIndex < maxNodesCount; nodeIndex++ {
		nodeName := fmt.Sprintf("node-%d", nodeIndex)
		node := BuildTestNode(nodeName, 10000, 10000)
		nodeSlice := createTestResourceSlice(node.Name, devicesPerSlice, 1, driverName)
		nodeInfo := framework.NewNodeInfo(node, []*resourceapi.ResourceSlice{nodeSlice})

		sharedClaim := createTestResourceClaim(devicesPerSlice, 1, driverName, deviceClassName)
		sharedClaim, satisfied := allocateResourceSlicesForClaim(sharedClaim, nodeName, nodeSlice)
		if !satisfied {
			b.Errorf("Error during setup, claim allocation cannot be satistied")
		}

		claimsOnNode := make([]*resourceapi.ResourceClaim, maxPodsCount)
		podsOnNode := make([]*apiv1.Pod, maxPodsCount)
		for podIndex := 0; podIndex < maxPodsCount; podIndex++ {
			podName := fmt.Sprintf("pod-%d-%d", nodeIndex, podIndex)
			ownedClaim := createTestResourceClaim(1, 1, driverName, deviceClassName)
			pod := BuildTestPod(
				podName,
				1,
				1,
				WithResourceClaim(ownedClaim.Name, ownedClaim.Name, ""),
			)

			ownedClaim = drautils.TestClaimWithPodOwnership(pod, ownedClaim)
			ownedClaim, satisfied := allocateResourceSlicesForClaim(ownedClaim, nodeName, nodeSlice)
			if !satisfied {
				b.Errorf("Error during setup, claim allocation cannot be satistied")
			}

			podsOnNode[podIndex] = pod
			claimsOnNode[podIndex] = ownedClaim
		}

		nodeInfos[nodeIndex] = nodeInfo
		sharedClaims[nodeIndex] = sharedClaim
		ownedClaims[nodeIndex] = claimsOnNode
		owningPods[nodeIndex] = podsOnNode
	}

	b.ResetTimer()
	for snapshotName, snapshotFactory := range snapshots {
		b.Run(snapshotName, func(b *testing.B) {
			for cfgName, cfg := range configurations {
				b.Run(cfgName, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						snapshot, err := snapshotFactory()
						if err != nil {
							b.Errorf("Failed to create a snapshot: %v", err)
						}

						draSnapshot := drasnapshot.NewSnapshot(
							nil,
							nil,
							nil,
							devicesClasses,
						)

						draSnapshot.AddClaims(sharedClaims)
						for nodeIndex := 0; nodeIndex < cfg.nodesCount; nodeIndex++ {
							draSnapshot.AddClaims(ownedClaims[nodeIndex])
						}

						err = snapshot.SetClusterState(nil, nil, draSnapshot)
						if err != nil {
							b.Errorf("Failed to set cluster state: %v", err)
						}

						for nodeIndex := 0; nodeIndex < cfg.nodesCount; nodeIndex++ {
							nodeInfo := nodeInfos[nodeIndex]
							for i := 0; i < cfg.forks; i++ {
								snapshot.Fork()
							}

							err := snapshot.AddNodeInfo(nodeInfo)
							if err != nil {
								b.Errorf("Failed to add node info to snapshot: %v", err)
							}

							sharedClaim := sharedClaims[nodeIndex]
							for podIndex := 0; podIndex < cfg.sharedClaimPods; podIndex++ {
								pod := BuildTestPod(
									fmt.Sprintf("pod-%d", podIndex),
									1,
									1,
									WithResourceClaim(sharedClaim.Name, sharedClaim.Name, ""),
								)

								err := snapshot.SchedulePod(pod, nodeInfo.Node().Name)
								if err != nil {
									b.Errorf(
										"Failed to schedule a pod %s to node %s: %v",
										pod.Name,
										nodeInfo.Node().Name,
										err,
									)
								}
							}

							for podIndex := 0; podIndex < cfg.ownedClaimPods; podIndex++ {
								owningPod := owningPods[nodeIndex][podIndex]
								err := snapshot.SchedulePod(owningPod, nodeInfo.Node().Name)
								if err != nil {
									b.Errorf(
										"Failed to schedule a pod %s to node %s: %v",
										owningPod.Name,
										nodeInfo.Node().Name,
										err,
									)
								}
							}

							for i := 0; i < cfg.commits; i++ {
								snapshot.Commit()
							}

							for i := 0; i < cfg.reverts; i++ {
								snapshot.Revert()
							}
						}
					}
				})
			}
		})
	}
}
