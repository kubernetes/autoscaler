/*
Copyright 2023 The Kubernetes Authors.

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

package upcloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
)

func TestUpCloudNodeGroup_Id(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{clusterID: uuid.New(), name: "test"}
	require.Equal(t, fmt.Sprintf("%s/%s", g.clusterID.String(), g.name), g.Id())
}

func TestUpCloudNodeGroup_MinSize(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{minSize: 1}
	require.Equal(t, 1, g.MinSize(t.Context()))
}

func TestUpCloudNodeGroup_MaxSize(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{maxSize: 1}
	require.Equal(t, 1, g.MaxSize(t.Context()))
}

func TestUpCloudNodeGroup_TargetSize(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{size: 1}
	size, err := g.TargetSize(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, size)
}

func TestUpCloudNodeGroup_IncreaseSize(t *testing.T) {
	t.Parallel()
	clusterID := uuid.New()
	svc := newMockService(clusterID)
	g := &upCloudNodeGroup{size: 1, maxSize: 20, name: "group1", svc: svc, clusterID: clusterID}
	require.NoError(t, g.IncreaseSize(t.Context(), 1))
	size, _ := g.TargetSize(t.Context())
	require.Equal(t, 2, size)
}

func TestUpCloudNodeGroup_DecreaseTargetSize(t *testing.T) {
	t.Parallel()

	clusterID := uuid.New()
	svc := newMockService(clusterID)
	g := &upCloudNodeGroup{size: 3, maxSize: 20, name: "group2", svc: svc, clusterID: clusterID}
	require.NoError(t, g.DecreaseTargetSize(t.Context(), -1))
	size, _ := g.TargetSize(t.Context())
	require.Equal(t, 2, size)
}

func TestUpCloudNodeGroup_DeleteNodes(t *testing.T) {
	t.Parallel()

	clusterID := uuid.New()
	svc := newMockService(clusterID)
	kng := svc.Clusters[clusterID.String()].NodeGroups[0]
	g := &upCloudNodeGroup{size: kng.Count, maxSize: 20, name: kng.Name, svc: svc, clusterID: clusterID}
	size, _ := g.TargetSize(t.Context())
	require.Equal(t, kng.Count, size)
	require.NoError(t, g.DeleteNodes(t.Context(), []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: "group1-node-1"}},
	}))
	size, _ = g.TargetSize(t.Context())
	require.Equal(t, kng.Count-1, size)
}

func TestUpCloudNodeGroup_ForceDeleteNodes(t *testing.T) {
	t.Parallel()

	clusterID := uuid.New()
	svc := newMockService(clusterID)
	kng := svc.Clusters[clusterID.String()].NodeGroups[0]
	g := &upCloudNodeGroup{size: kng.Count, maxSize: 20, name: kng.Name, svc: svc, clusterID: clusterID}
	size, _ := g.TargetSize(t.Context())
	require.Equal(t, kng.Count, size)
	require.NoError(t, g.ForceDeleteNodes(t.Context(), []*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: "group1-node-1"}},
	}))
	size, _ = g.TargetSize(t.Context())
	require.Equal(t, kng.Count-1, size)
}

func TestUpCloudNodeGroup_Nodes(t *testing.T) {
	t.Parallel()

	wantNodes := []cloudprovider.Instance{{
		Id: "test",
	}}
	g := &upCloudNodeGroup{nodes: wantNodes}
	gotNodes, err := g.Nodes(t.Context())
	require.NoError(t, err)
	require.Equal(t, wantNodes, gotNodes)
}

func TestUpCloudNodeGroup_Autoprovisioned(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{}
	require.False(t, g.Autoprovisioned(t.Context()))
}

func TestUpCloudNodeGroup_Create(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{}
	_, err := g.Create(t.Context())
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}

func TestUpCloudNodeGroup_Delete(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{}
	err := g.Delete(t.Context())
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}

func TestUpCloudNodeGroup_GetOptions(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{}
	_, err := g.GetOptions(t.Context(), config.NodeGroupAutoscalingOptions{})
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}

func TestUpCloudNodeGroup_Debug(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{name: "test"}
	require.NotEmpty(t, g.Debug(t.Context()))
}

func TestUpCloudNodeGroup_Exist(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{name: "test"}
	require.True(t, g.Exist(t.Context()))
}

func TestUpCloudNodeGroup_TemplateNodeInfo(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{}
	_, err := g.TemplateNodeInfo(t.Context())
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}

func TestUpCloudNodeGroup_AtomicIncreaseSize(t *testing.T) {
	t.Parallel()

	g := &upCloudNodeGroup{}
	require.ErrorIs(t, g.AtomicIncreaseSize(t.Context(), 1), cloudprovider.ErrNotImplemented)
}
