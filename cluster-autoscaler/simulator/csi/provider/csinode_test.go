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

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	v1storagelister "k8s.io/client-go/listers/storage/v1"
)

// testCSINodeLister is a fake CSINode lister for testing.
type testCSINodeLister struct {
	csiNodes []*storagev1.CSINode
}

func (l *testCSINodeLister) List(selector labels.Selector) ([]*storagev1.CSINode, error) {
	return l.csiNodes, nil
}

func (l *testCSINodeLister) Get(name string) (*storagev1.CSINode, error) {
	for _, csiNode := range l.csiNodes {
		if csiNode.Name == name {
			return csiNode, nil
		}
	}
	return nil, apierrors.NewNotFound(storagev1.Resource("csinode"), name)
}

func newTestCSINodeLister(csiNodes []*storagev1.CSINode) v1storagelister.CSINodeLister {
	return &testCSINodeLister{csiNodes: csiNodes}
}

func TestNewCSINodeProvider(t *testing.T) {
	csiNode1 := &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-1",
		},
		Spec: storagev1.CSINodeSpec{
			Drivers: []storagev1.CSINodeDriver{
				{
					Name:         "driver1",
					NodeID:       "node-id-1",
					TopologyKeys: []string{"topology-key-1"},
				},
			},
		},
	}

	csiNode2 := &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-2",
		},
		Spec: storagev1.CSINodeSpec{
			Drivers: []storagev1.CSINodeDriver{
				{
					Name:         "driver2",
					NodeID:       "node-id-2",
					TopologyKeys: []string{"topology-key-2"},
				},
			},
		},
	}

	csiNodes := []*storagev1.CSINode{csiNode1, csiNode2}
	lister := newTestCSINodeLister(csiNodes)

	provider := NewCSINodeProvider(lister)
	require.NotNil(t, provider)

	snapshot, err := provider.Snapshot()
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	// Verify that snapshot contains all CSINodes
	snapshotCSINodes := snapshot.CSINodes()
	require.NotNil(t, snapshotCSINodes)

	// Get CSINode by name from snapshot
	retrievedCSINode1, err := snapshotCSINodes.Get("node-1")
	require.NoError(t, err)
	assert.Equal(t, csiNode1.Name, retrievedCSINode1.Name)
	assert.Equal(t, len(csiNode1.Spec.Drivers), len(retrievedCSINode1.Spec.Drivers))
	assert.Equal(t, csiNode1.Spec.Drivers[0].Name, retrievedCSINode1.Spec.Drivers[0].Name)

	retrievedCSINode2, err := snapshotCSINodes.Get("node-2")
	require.NoError(t, err)
	assert.Equal(t, csiNode2.Name, retrievedCSINode2.Name)
	assert.Equal(t, len(csiNode2.Spec.Drivers), len(retrievedCSINode2.Spec.Drivers))
	assert.Equal(t, csiNode2.Spec.Drivers[0].Name, retrievedCSINode2.Spec.Drivers[0].Name)

	// Verify non-existent CSINode returns error
	_, err = snapshotCSINodes.Get("non-existent-node")
	assert.Error(t, err)
}

func TestNewCSINodeProviderFromInformers(t *testing.T) {
	// Create fake clientset
	clientset := fake.NewSimpleClientset()

	// Create informer factory
	informerFactory := informers.NewSharedInformerFactory(clientset, 0)

	// Create provider from informers - this verifies the constructor works
	provider := NewCSINodeProviderFromInformers(informerFactory)
	require.NotNil(t, provider)

	// Verify provider can create snapshot (even if empty)
	snapshot, err := provider.Snapshot()
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	// Verify snapshot is valid
	snapshotCSINodes := snapshot.CSINodes()
	require.NotNil(t, snapshotCSINodes)
}

func TestSnapshot_EmptyLister(t *testing.T) {
	lister := newTestCSINodeLister([]*storagev1.CSINode{})
	provider := NewCSINodeProvider(lister)

	snapshot, err := provider.Snapshot()
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	// Verify snapshot is empty but valid
	snapshotCSINodes := snapshot.CSINodes()
	require.NotNil(t, snapshotCSINodes)

	_, err = snapshotCSINodes.Get("non-existent")
	assert.Error(t, err)
}
