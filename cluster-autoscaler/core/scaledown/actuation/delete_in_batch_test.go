/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	clusterstate_utils "k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	kube_record "k8s.io/client-go/tools/record"
)

func TestAddNodeToBucket(t *testing.T) {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	ctx, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{}, nil, nil, provider, nil, nil)
	if err != nil {
		t.Fatalf("Couldn't set up autoscaling context: %v", err)
	}
	nodeGroup1 := "ng-1"
	nodeGroup2 := "ng-2"
	nodes1 := generateNodes(5, "ng-1")
	nodes2 := generateNodes(5, "ng-2")
	provider.AddNodeGroup(nodeGroup1, 1, 10, 5)
	provider.AddNodeGroup(nodeGroup2, 1, 10, 5)
	for _, node := range nodes1 {
		provider.AddNode(nodeGroup1, node)
	}
	for _, node := range nodes2 {
		provider.AddNode(nodeGroup2, node)
	}
	testcases := []struct {
		name        string
		nodes       []*apiv1.Node
		wantBatches int
		drained     bool
	}{
		{
			name:        "Add 1 node",
			nodes:       []*apiv1.Node{nodes1[0]},
			wantBatches: 1,
		},
		{
			name:        "Add nodes that belong to one nodeGroup",
			nodes:       nodes1,
			wantBatches: 1,
		},
		{
			name:        "Add 3 nodes that belong to 2 nodeGroups",
			nodes:       []*apiv1.Node{nodes1[0], nodes2[0], nodes2[1]},
			wantBatches: 2,
		},
		{
			name:        "Add 3 nodes that belong to 2 nodeGroups, all nodes are drained",
			nodes:       []*apiv1.Node{nodes1[0], nodes2[0], nodes2[1]},
			wantBatches: 2,
			drained:     true,
		},
	}
	for _, test := range testcases {
		d := NodeDeletionBatcher{
			ctx:                   &ctx,
			clusterState:          nil,
			nodeDeletionTracker:   nil,
			deletionsPerNodeGroup: make(map[string][]*apiv1.Node),
			drainedNodeDeletions:  make(map[string]bool),
		}
		batchCount := 0
		for _, node := range test.nodes {
			_, first, err := d.addNodeToBucket(node, test.drained)
			if err != nil {
				t.Errorf("addNodeToBucket return error %q when addidng node %v", err, node)
			}
			if first {
				batchCount += 1
			}
		}
		if batchCount != test.wantBatches {
			t.Errorf("Want %d batches, got %d batches", test.wantBatches, batchCount)
		}

	}
}

func TestRemove(t *testing.T) {
	testCases := []struct {
		name           string
		err            bool
		numNodes       int
		failedDeletion int
		addNgToBucket  bool
	}{
		{
			name:          "Remove NodeGroup that is not present in bucket",
			err:           true,
			addNgToBucket: false,
		},
		{
			name:           "Regular successful remove",
			err:            false,
			numNodes:       5,
			failedDeletion: 0,
			addNgToBucket:  true,
		},
		{
			name:           "Unsuccessful remove",
			numNodes:       5,
			failedDeletion: 1,
			addNgToBucket:  true,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test := test
			fakeClient := &fake.Clientset{}
			fakeLogRecorder, _ := clusterstate_utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")

			failedNodeDeletion := make(map[string]bool)
			deletedNodes := make(chan string, 10)
			notDeletedNodes := make(chan string, 10)
			// Hook node deletion at the level of cloud provider, to gather which nodes were deleted, and to fail the deletion for
			// certain nodes to simulate errors.
			provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
				if failedNodeDeletion[node] {
					notDeletedNodes <- node
					return fmt.Errorf("SIMULATED ERROR: won't remove node")
				}
				deletedNodes <- node
				return nil
			})

			fakeClient.Fake.AddReactor("update", "nodes",
				func(action core.Action) (bool, runtime.Object, error) {
					update := action.(core.UpdateAction)
					obj := update.GetObject().(*apiv1.Node)
					return true, obj, nil
				})

			ctx, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{}, fakeClient, nil, provider, nil, nil)
			clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, fakeLogRecorder, NewBackoff())
			if err != nil {
				t.Fatalf("Couldn't set up autoscaling context: %v", err)
			}

			ng := "ng"
			provider.AddNodeGroup(ng, 1, 10, test.numNodes)

			d := NodeDeletionBatcher{
				ctx:                   &ctx,
				clusterState:          clusterStateRegistry,
				nodeDeletionTracker:   deletiontracker.NewNodeDeletionTracker(1 * time.Minute),
				deletionsPerNodeGroup: make(map[string][]*apiv1.Node),
				drainedNodeDeletions:  make(map[string]bool),
			}
			nodes := generateNodes(test.numNodes, ng)
			failedDeletion := test.failedDeletion
			for _, node := range nodes {
				if failedDeletion > 0 {
					failedNodeDeletion[node.Name] = true
					failedDeletion -= 1
				}
				provider.AddNode(ng, node)
			}
			if test.addNgToBucket {
				for _, node := range nodes {
					node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
						Key:    taints.ToBeDeletedTaint,
						Effect: apiv1.TaintEffectNoSchedule,
					})
					_, _, err := d.addNodeToBucket(node, true)
					if err != nil {
						t.Errorf("addNodeToBucket return error %q when addidng node %v", err, node)
					}
				}
			}

			err = d.remove(ng)
			if test.err {
				if err == nil {
					t.Errorf("remove() should return error, but return nil")
				}
				return
			}
			if err != nil {
				t.Errorf("remove() return error, but shouldn't")
			}
			if test.failedDeletion == 0 {
				gotDeletedNodes := []string{}
				for i := 0; i < test.numNodes; i++ {
					select {
					case deletedNode := <-deletedNodes:
						gotDeletedNodes = append(gotDeletedNodes, deletedNode)
					case <-time.After(4 * time.Second):
						t.Errorf("Timeout while waiting for deleted nodes.")
					}
				}
			} else {
				select {
				case <-notDeletedNodes:
				case <-time.After(4 * time.Second):
					t.Errorf("Timeout while waiting for deleted nodes.")
				}
			}
			if len(d.deletionsPerNodeGroup) > 0 {
				t.Errorf("Number of bathces hasn't reach 0 after remove(), got: %v", len(d.deletionsPerNodeGroup))
			}
			if len(d.drainedNodeDeletions) > 0 {
				t.Errorf(" Drained node map is not empty, got: %v", len(d.drainedNodeDeletions))
			}
		})
	}
}
