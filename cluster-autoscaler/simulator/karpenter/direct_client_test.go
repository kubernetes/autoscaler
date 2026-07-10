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
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
)

func TestDirectClient_Get_CSINode(t *testing.T) {
	// Setup snapshot
	snapshot := testsnapshot.NewTestSnapshotOrDie(t)

	// Create a StorageClass
	sc1 := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sc-1",
		},
		Provisioner: "kubernetes.io/gce-pd", // In-tree provisioner that translates to pd.csi.storage.gke.io
	}
	sc2 := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sc-2",
		},
		Provisioner: "my-custom-driver",
	}

	// Create a real CSINode for node1
	csiNode1 := &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Spec: storagev1.CSINodeSpec{
			Drivers: []storagev1.CSINodeDriver{
				{
					Name:         "pd.csi.storage.gke.io",
					NodeID:       "node1-gke-id",
					TopologyKeys: []string{"topology.gke.io/zone"},
				},
			},
		},
	}

	// Populate snapshot
	err := snapshot.SetClusterState(
		[]*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node2"}}, // No CSINode for node2 in snapshot
		},
		nil,
		nil,
		csisnapshot.NewSnapshot(map[string]*storagev1.CSINode{"node1": csiNode1}),
		nil,
		nil,
		[]*storagev1.StorageClass{sc1, sc2},
	)
	assert.NoError(t, err)

	// Initialize DirectClient
	directClient := NewDirectClient(snapshot, nil, nil, nil, nil)

	// Test case 1: Get existing CSINode (node1)
	ctx := context.Background()
	var gotCSINode1 storagev1.CSINode
	err = directClient.Get(ctx, types.NamespacedName{Name: "node1"}, &gotCSINode1)
	assert.NoError(t, err)
	assert.Equal(t, "node1", gotCSINode1.Name)
	assert.Len(t, gotCSINode1.Spec.Drivers, 1)
	assert.Equal(t, "pd.csi.storage.gke.io", gotCSINode1.Spec.Drivers[0].Name)
	assert.Equal(t, "node1-gke-id", gotCSINode1.Spec.Drivers[0].NodeID)
	assert.Equal(t, []string{"topology.gke.io/zone"}, gotCSINode1.Spec.Drivers[0].TopologyKeys)

	// Test case 2: Get non-existing CSINode (node2) -> Should be mocked
	var gotCSINode2 storagev1.CSINode
	err = directClient.Get(ctx, types.NamespacedName{Name: "node2"}, &gotCSINode2)
	assert.NoError(t, err)
	assert.Equal(t, "node2", gotCSINode2.Name)
	assert.Len(t, gotCSINode2.Spec.Drivers, 2)

	// Verify drivers in mock
	drivers := map[string]storagev1.CSINodeDriver{}
	for _, d := range gotCSINode2.Spec.Drivers {
		drivers[d.Name] = d
	}

	// kubernetes.io/gce-pd translates to pd.csi.storage.gke.io
	d1, ok := drivers["pd.csi.storage.gke.io"]
	assert.True(t, ok)
	assert.Equal(t, "node2", d1.NodeID)
	assert.Contains(t, d1.TopologyKeys, apiv1.LabelZoneFailureDomainStable)
	assert.Contains(t, d1.TopologyKeys, apiv1.LabelHostname)

	// my-custom-driver remains my-custom-driver
	d2, ok := drivers["my-custom-driver"]
	assert.True(t, ok)
	assert.Equal(t, "node2", d2.NodeID)
	assert.Contains(t, d2.TopologyKeys, apiv1.LabelZoneFailureDomainStable)
	assert.Contains(t, d2.TopologyKeys, apiv1.LabelHostname)
}
