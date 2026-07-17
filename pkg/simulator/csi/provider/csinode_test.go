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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewCSINodeProviderFromInformers(t *testing.T) {
	tests := []struct {
		name     string
		csiNodes []*storagev1.CSINode
	}{
		{
			name:     "empty",
			csiNodes: []*storagev1.CSINode{},
		},
		{
			name: "non-empty",
			csiNodes: []*storagev1.CSINode{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
					Spec: storagev1.CSINodeSpec{
						Drivers: []storagev1.CSINodeDriver{
							{
								Name: "driver1",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
					Spec: storagev1.CSINodeSpec{
						Drivers: []storagev1.CSINodeDriver{
							{
								Name: "driver2",
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			csiNodes := make([]runtime.Object, len(test.csiNodes))
			for i, csiNode := range test.csiNodes {
				csiNodes[i] = csiNode
			}
			clientset := fake.NewSimpleClientset(csiNodes...)
			informerFactory := informers.NewSharedInformerFactory(clientset, 0)

			// Create the provider first to ensure the informer is created
			provider := NewCSINodeProviderFromInformers(informerFactory)
			require.NotNil(t, provider)

			informerFactory.Start(t.Context().Done())
			informerFactory.WaitForCacheSync(t.Context().Done())

			snapshot, err := provider.Snapshot()
			require.NoError(t, err)
			require.NotNil(t, snapshot)

			snapshotCSINodes := snapshot.CSINodes()
			require.NotNil(t, snapshotCSINodes)

			_, err = snapshotCSINodes.Get("non-existent")
			assert.Error(t, err)

			for _, csiNode := range test.csiNodes {
				retrievedCSINode, err := snapshotCSINodes.Get(csiNode.Name)
				require.NoError(t, err)
				assert.Equal(t, csiNode.Name, retrievedCSINode.Name)
				assert.Equal(t, len(csiNode.Spec.Drivers), len(retrievedCSINode.Spec.Drivers))
			}
		})
	}
}
