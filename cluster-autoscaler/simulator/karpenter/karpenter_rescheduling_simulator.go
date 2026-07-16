/*
Copyright 2026 The Kubernetes Authors.

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

package karpenter

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/utils/clock"
	karpscheduling "sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling"
	karpstate "sigs.k8s.io/karpenter/pkg/controllers/state"
)

// KarpenterReschedulingSimulator implements PodSchedulingSimulator using Karpenter's solver.
type KarpenterReschedulingSimulator struct{}

// NewKarpenterReschedulingSimulator returns a new KarpenterReschedulingSimulator.
func NewKarpenterReschedulingSimulator() *KarpenterReschedulingSimulator {
	return &KarpenterReschedulingSimulator{}
}

// TrySchedulePods attempts to schedule provided pods on any acceptable nodes using the Karpenter solver.
func (s *KarpenterReschedulingSimulator) TrySchedulePods(snapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, breakOnFailure bool, opts clustersnapshot.SchedulingOptions) ([]clustersnapshot.Status, int, error) {
	ctx := context.Background()
	clk := clock.RealClock{}

	// 1. Get NodeInfos from snapshot
	nodeInfos, err := snapshot.ListNodeInfos()
	if err != nil {
		return nil, 0, err
	}
	
	// 2. Hydrate the cluster state with pruning
	relevantPods, relevantNodes, err := HydrateClusterState(ctx, snapshot, nodeInfos, pods)
	if err != nil {
		return nil, 0, err
	}

	// 3. Initialize optimized DirectClient and state.Cluster with relevant pods only
	directClient := NewDirectClient(snapshot, append(relevantPods, pods...), nil, nil)
	cluster := karpstate.NewCluster(clk, directClient, nil)
	for _, node := range relevantNodes {
		if err := cluster.UpdateNode(ctx, node); err != nil {
			return nil, 0, err
		}
	}
	for _, pod := range relevantPods {
		if err := cluster.UpdatePod(ctx, pod); err != nil {
			return nil, 0, err
		}
	}

	// Collect state nodes
	var stateNodes []*karpstate.StateNode
	for n := range cluster.Nodes() {
		stateNodes = append(stateNodes, n)
	}

	// Filter state nodes based on IsNodeAcceptable predicate
	var filteredStateNodes []*karpstate.StateNode
	for _, sn := range stateNodes {
		ni, err := snapshot.GetNodeInfo(sn.Name())
		if err == nil && (opts.IsNodeAcceptable == nil || opts.IsNodeAcceptable(ni)) {
			filteredStateNodes = append(filteredStateNodes, sn)
		}
	}

	// Initialize Topology
	topology, err := karpscheduling.NewTopology(ctx, directClient, cluster, stateNodes, nil, nil, pods)
	if err != nil {
		return nil, 0, err
	}

	// Initialize NewScheduler
	scheduler := karpscheduling.NewScheduler(
		ctx,
		directClient,
		nil,
		cluster,
		filteredStateNodes,
		topology,
		nil,
		nil,
		&NoopRecorder{},
		clk,
		nil,
		nil, // dynamicresources.Allocator
	)

	results, err := scheduler.Solve(ctx, pods)
	if err != nil {
		return nil, 0, err
	}

	var statuses []clustersnapshot.Status
	for _, existingNode := range results.ExistingNodes {
		for _, pod := range existingNode.Pods {
			statuses = append(statuses, clustersnapshot.Status{
				Pod:      pod,
				NodeName: existingNode.Name(),
			})
		}
	}

	// Calculate overflowing controller count
	overflowingControllers := make(map[types.UID]bool)
	for p := range results.PodErrors {
		if owner := metav1.GetControllerOf(p); owner != nil {
			overflowingControllers[owner.UID] = true
		} else {
			overflowingControllers[p.UID] = true
		}
	}

	return statuses, len(overflowingControllers), nil
}

// DropOldHints is a no-op as Karpenter solver does not use hints.
func (s *KarpenterReschedulingSimulator) DropOldHints() {}
