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

package snapshot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/utils/ptr"
)

var (
	node1Name = "node1"
	node2Name = "node2"
	node3Name = "node3"

	driverName1 = "ebs.csi.aws.com"
	driverName2 = "pd.csi.google.com"

	node1CSINode = &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: node1Name,
			UID:  types.UID(node1Name + "-uid"),
		},
		Spec: storagev1.CSINodeSpec{
			Drivers: []storagev1.CSINodeDriver{
				{
					Name:   driverName1,
					NodeID: node1Name,
					Allocatable: &storagev1.VolumeNodeResources{
						Count: ptr.To[int32](10),
					},
				},
			},
		},
	}

	node2CSINode = &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: node2Name,
			UID:  types.UID(node2Name + "-uid"),
		},
		Spec: storagev1.CSINodeSpec{
			Drivers: []storagev1.CSINodeDriver{
				{
					Name:   driverName1,
					NodeID: node2Name,
					Allocatable: &storagev1.VolumeNodeResources{
						Count: ptr.To[int32](20),
					},
				},
				{
					Name:   driverName2,
					NodeID: node2Name,
					Allocatable: &storagev1.VolumeNodeResources{
						Count: ptr.To[int32](5),
					},
				},
			},
		},
	}

	node3CSINode = &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: node3Name,
			UID:  types.UID(node3Name + "-uid"),
		},
		Spec: storagev1.CSINodeSpec{
			Drivers: []storagev1.CSINodeDriver{
				{
					Name:   driverName2,
					NodeID: node3Name,
					Allocatable: &storagev1.VolumeNodeResources{
						Count: ptr.To[int32](15),
					},
				},
			},
		},
	}

	node1 = test.BuildTestNode(node1Name, 1000, 1000)
	node2 = test.BuildTestNode(node2Name, 2000, 2000)
	node3 = test.BuildTestNode(node3Name, 1500, 1500)
)

func TestSnapshotCSINodes(t *testing.T) {
	for _, tc := range []struct {
		testName string

		csiNodes map[string]*storagev1.CSINode

		csiNodesModFun        func(snapshot *Snapshot) error
		wantCSINodesModFunErr error

		nodeName          string
		wantCSINode       *storagev1.CSINode
		wantCSINodeFound  bool
		wantCSINodeGetErr error

		wantAllCSINodes []*storagev1.CSINode
	}{
		{
			testName: "Get(): non-existent CSI node returns error",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
			},
			nodeName:          node2Name,
			wantCSINodeGetErr: cmpopts.AnyError,
		},
		{
			testName: "Get(): existing CSI node is correctly returned",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
				node2Name: node2CSINode.DeepCopy(),
			},
			nodeName:         node2Name,
			wantCSINode:      node2CSINode,
			wantCSINodeFound: true,
		},
		{
			testName: "AddCSINode(): adding duplicate CSI node is an error",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
			},
			csiNodesModFun: func(snapshot *Snapshot) error {
				return snapshot.AddCSINode(node1CSINode.DeepCopy())
			},
			wantCSINodesModFunErr: cmpopts.AnyError,
			wantAllCSINodes:       []*storagev1.CSINode{node1CSINode}, // unchanged on error
		},
		{
			testName: "AddCSINode(): new CSI node is correctly added",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
			},
			csiNodesModFun: func(snapshot *Snapshot) error {
				return snapshot.AddCSINode(node2CSINode.DeepCopy())
			},
			wantAllCSINodes: []*storagev1.CSINode{node1CSINode, node2CSINode},
		},
		{
			testName: "AddCSINodes(): adding multiple CSI nodes works correctly",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
			},
			csiNodesModFun: func(snapshot *Snapshot) error {
				return snapshot.AddCSINodes([]*storagev1.CSINode{node2CSINode.DeepCopy(), node3CSINode.DeepCopy()})
			},
			wantAllCSINodes: []*storagev1.CSINode{node1CSINode, node2CSINode, node3CSINode},
		},
		{
			testName: "AddCSINodes(): adding duplicate in batch is an error",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
			},
			csiNodesModFun: func(snapshot *Snapshot) error {
				return snapshot.AddCSINodes([]*storagev1.CSINode{node1CSINode.DeepCopy(), node2CSINode.DeepCopy()})
			},
			wantCSINodesModFunErr: cmpopts.AnyError,
			wantAllCSINodes:       []*storagev1.CSINode{node1CSINode}, // unchanged on error
		},
		{
			testName: "RemoveCSINode(): removing non-existent CSI node is a no-op",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
			},
			csiNodesModFun: func(snapshot *Snapshot) error {
				snapshot.RemoveCSINode(node2Name)
				return nil
			},
			wantAllCSINodes: []*storagev1.CSINode{node1CSINode},
		},
		{
			testName: "RemoveCSINode(): removing existing CSI node works correctly",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
				node2Name: node2CSINode.DeepCopy(),
			},
			csiNodesModFun: func(snapshot *Snapshot) error {
				snapshot.RemoveCSINode(node1Name)
				return nil
			},
			wantAllCSINodes: []*storagev1.CSINode{node2CSINode},
		},
		{
			testName: "CSINodes().List(): all CSI nodes are returned",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
				node2Name: node2CSINode.DeepCopy(),
				node3Name: node3CSINode.DeepCopy(),
			},
			wantAllCSINodes: []*storagev1.CSINode{node1CSINode, node2CSINode, node3CSINode},
		},
		{
			testName:        "NewEmptySnapshot(): creates empty snapshot",
			wantAllCSINodes: []*storagev1.CSINode{},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			var snapshot *Snapshot
			if tc.csiNodes != nil {
				snapshot = NewSnapshot(tc.csiNodes)
			} else {
				snapshot = NewEmptySnapshot()
			}

			if tc.csiNodesModFun != nil {
				err := tc.csiNodesModFun(snapshot)
				if diff := cmp.Diff(tc.wantCSINodesModFunErr, err, cmpopts.EquateErrors()); diff != "" {
					t.Fatalf("Snapshot modification: unexpected error (-want +got): %s", diff)
				}
			}

			if tc.nodeName != "" {
				csiNode, err := snapshot.Get(tc.nodeName)
				if diff := cmp.Diff(tc.wantCSINodeGetErr, err, cmpopts.EquateErrors()); diff != "" {
					t.Fatalf("Snapshot.Get(): unexpected error (-want +got): %s", diff)
				}
				if tc.wantCSINodeFound {
					if diff := cmp.Diff(tc.wantCSINode, csiNode, cmpopts.EquateEmpty()); diff != "" {
						t.Errorf("Snapshot.Get(): unexpected output (-want +got): %s", diff)
					}
				}
			}

			if tc.wantAllCSINodes != nil {
				allCSINodes, err := snapshot.CSINodes().List()
				if err != nil {
					t.Fatalf("CSINodes().List(): unexpected error: %v", err)
				}
				if diff := cmp.Diff(tc.wantAllCSINodes, allCSINodes, cmpopts.EquateEmpty(), test.IgnoreObjectOrder[*storagev1.CSINode]()); diff != "" {
					t.Errorf("Snapshot: unexpected CSINode state (-want +got): %s", diff)
				}
			}
		})
	}
}

func TestSnapshotCSINodeLister(t *testing.T) {
	csiNodes := map[string]*storagev1.CSINode{
		node1Name: node1CSINode.DeepCopy(),
		node2Name: node2CSINode.DeepCopy(),
	}
	snapshot := NewSnapshot(csiNodes)
	lister := snapshot.CSINodes()

	t.Run("Get(): retrieves CSI node by name", func(t *testing.T) {
		csiNode, err := lister.Get(node1Name)
		if err != nil {
			t.Fatalf("Get(): unexpected error: %v", err)
		}
		if diff := cmp.Diff(node1CSINode, csiNode, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Get(): unexpected CSINode (-want +got): %s", diff)
		}
	})

	t.Run("Get(): non-existent node returns error", func(t *testing.T) {
		_, err := lister.Get(node3Name)
		if err == nil {
			t.Fatal("Get(): expected error for non-existent node, got nil")
		}
	})

	t.Run("List(): returns all CSI nodes", func(t *testing.T) {
		allCSINodes, err := lister.List()
		if err != nil {
			t.Fatalf("List(): unexpected error: %v", err)
		}
		wantCSINodes := []*storagev1.CSINode{node1CSINode, node2CSINode}
		if diff := cmp.Diff(wantCSINodes, allCSINodes, cmpopts.EquateEmpty(), test.IgnoreObjectOrder[*storagev1.CSINode]()); diff != "" {
			t.Errorf("List(): unexpected CSINodes (-want +got): %s", diff)
		}
	})
}

func TestSnapshotAddCSINodeInfoToNodeInfo(t *testing.T) {
	for _, tc := range []struct {
		testName string

		csiNodes map[string]*storagev1.CSINode

		nodeName    string
		wantCSINode *storagev1.CSINode
		wantErr     error
	}{
		{
			testName: "AddCSINodeInfoToNodeInfo(): missing CSI node is an error",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
			},
			nodeName: node2Name,
			wantErr:  cmpopts.AnyError,
		},
		{
			testName: "AddCSINodeInfoToNodeInfo(): CSI node info is correctly added to NodeInfo",
			csiNodes: map[string]*storagev1.CSINode{
				node1Name: node1CSINode.DeepCopy(),
				node2Name: node2CSINode.DeepCopy(),
			},
			nodeName:    node1Name,
			wantCSINode: node1CSINode,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			snapshot := NewSnapshot(tc.csiNodes)

			node := test.BuildTestNode(tc.nodeName, 1000, 1000)
			nodeInfo := framework.NewNodeInfo(node, nil)
			resultNodeInfo, err := snapshot.AddCSINodeInfoToNodeInfo(nodeInfo)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("AddCSINodeInfoToNodeInfo(): unexpected error (-want +got): %s", diff)
			}

			if tc.wantCSINode != nil {
				if resultNodeInfo == nil {
					t.Fatal("AddCSINodeInfoToNodeInfo(): returned nil NodeInfo")
				}
				gotCSINode := resultNodeInfo.CSINode
				if diff := cmp.Diff(tc.wantCSINode, gotCSINode, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("AddCSINodeInfoToNodeInfo(): unexpected CSINode (-want +got): %s", diff)
				}
			}
		})
	}
}

func TestSnapshotForkCommitRevert(t *testing.T) {
	initialCSINodes := map[string]*storagev1.CSINode{
		node1Name: node1CSINode.DeepCopy(),
	}
	initialState := NewSnapshot(initialCSINodes)

	addedCSINode := node2CSINode.DeepCopy()

	modifiedCSINodes := map[string]*storagev1.CSINode{
		node1Name: node1CSINode.DeepCopy(),
		node2Name: addedCSINode.DeepCopy(),
	}
	// Expected state after modifications are applied
	modifiedState := NewSnapshot(modifiedCSINodes)

	applyModifications := func(t *testing.T, s *Snapshot) {
		t.Helper()

		if err := s.AddCSINode(addedCSINode.DeepCopy()); err != nil {
			t.Fatalf("failed to add CSI node %s: %v", addedCSINode.Name, err)
		}
	}

	compareSnapshots := func(t *testing.T, want, got *Snapshot, msg string) {
		t.Helper()
		if diff := cmp.Diff(want, got, snapshotFlattenedComparer(), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("%s: Snapshot state mismatch (-want +got):\n%s", msg, diff)
		}
	}

	t.Run("Fork", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		compareSnapshots(t, modifiedState, snapshot, "After Fork and Modify")
	})

	t.Run("ForkRevert", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Revert()
		compareSnapshots(t, initialState, snapshot, "After Fork, Modify, Revert")
	})

	t.Run("ForkCommit", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Commit()
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Commit")
	})

	t.Run("ForkForkRevertRevert", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Fork()

		// Apply further modifications in second fork (add node3)
		furtherModifiedCSINode := node3CSINode.DeepCopy()
		if err := snapshot.AddCSINode(furtherModifiedCSINode); err != nil {
			t.Fatalf("AddCSINode failed in second fork: %v", err)
		}

		snapshot.Revert() // Revert second fork
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Fork, Modify, Revert")

		snapshot.Revert() // Revert first fork
		compareSnapshots(t, initialState, snapshot, "After Fork, Modify, Fork, Modify, Revert, Revert")
	})

	t.Run("ForkForkCommitRevert", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Commit() // Commit second fork into first fork
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Fork, Modify, Commit")

		snapshot.Revert() // Revert first fork (which now contains committed changes)
		compareSnapshots(t, initialState, snapshot, "After Fork, Fork, Modify, Commit, Revert")
	})

	t.Run("ForkForkRevertCommit", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Fork()
		// Apply further modifications in second fork (add node3)
		furtherModifiedCSINode := node3CSINode.DeepCopy()
		if err := snapshot.AddCSINode(furtherModifiedCSINode); err != nil {
			t.Fatalf("AddCSINode failed in second fork: %v", err)
		}

		snapshot.Revert() // Revert second fork
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Fork, Modify, Revert")

		snapshot.Commit() // Commit first fork (with original modifications)
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Fork, Modify, Revert, Commit")
	})

	t.Run("CommitNoFork", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Commit() // Should be a no-op
		compareSnapshots(t, initialState, snapshot, "After Commit with no Fork")
	})

	t.Run("RevertNoFork", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Revert() // Should be a no-op
		compareSnapshots(t, initialState, snapshot, "After Revert with no Fork")
	})

	t.Run("ForkCommitRevert", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Commit()
		// Now try to revert the committed changes (should be no-op as only base layer exists)
		snapshot.Revert()
		compareSnapshots(t, modifiedState, snapshot, "After Fork, Modify, Commit, Revert")
	})

	t.Run("ForkRevertFork", func(t *testing.T) {
		snapshot := CloneTestSnapshot(initialState)
		snapshot.Fork()
		applyModifications(t, snapshot)
		snapshot.Revert()
		compareSnapshots(t, initialState, snapshot, "After Fork, Modify, Revert")

		snapshot.Fork() // Fork again from the reverted (initial) state
		differentCSINode := node3CSINode.DeepCopy()
		if err := snapshot.AddCSINode(differentCSINode); err != nil {
			t.Fatalf("AddCSINode failed in second fork: %v", err)
		}

		expectedState := CloneTestSnapshot(initialState)
		expectedState.AddCSINode(differentCSINode.DeepCopy()) // Apply same change to expected state

		compareSnapshots(t, expectedState, snapshot, "After Fork, Modify, Revert, Fork, Modify")
	})
}
