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
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/mocks"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	"sigs.k8s.io/cluster-autoscaler/pkg/config/dynamic"
)

func TestClusterMaxNodes(t *testing.T) {
	t.Parallel()

	clusterID := uuid.New()
	mock := newMockService(clusterID)
	want := 10
	got, err := clusterMaxNodes(context.TODO(), mock, clusterID, 10)
	require.NoError(t, err)
	require.Equal(t, want, got)

	got, err = clusterMaxNodes(context.TODO(), mock, clusterID, 0)
	require.NoError(t, err)
	require.Equal(t, mock.Plans[0].MaxNodes, got)

	_, err = clusterMaxNodes(context.TODO(), mock, clusterID, 100)
	require.Error(t, err)
}

func TestClusterPlanByName(t *testing.T) {
	t.Parallel()

	want := upcloud.KubernetesPlan{
		Name: "match",
	}
	mock := mocks.UpCloudService{
		Plans: []upcloud.KubernetesPlan{want},
	}
	got, err := clusterPlanByName(context.TODO(), &mock, want.Name)
	require.NoError(t, err)
	require.Equal(t, want.Name, got.Name)
}

func TestManager(t *testing.T) {
	t.Parallel()

	clusterID := uuid.New()
	upCfg := upCloudConfig{ClusterID: clusterID.String()}
	svc := newMockService(clusterID)

	m, err := newManager(
		context.Background(),
		svc,
		upCfg,
		config.AutoscalingOptions{},
		cloudprovider.NodeGroupDiscoveryOptions{
			NodeGroupSpecs: []string{"1:2:one", "11:20:two"},
		},
	)
	require.NoError(t, err)
	require.Equal(t, upCfg.ClusterID, m.clusterID.String())
	require.Equal(t, dynamic.NodeGroupSpec{Name: "one", MinSize: 1, MaxSize: 2}, m.nodeGroupSpecs["one"])
	require.Equal(t, dynamic.NodeGroupSpec{Name: "two", MinSize: 11, MaxSize: 20}, m.nodeGroupSpecs["two"])
	require.NoError(t, m.refresh())
	require.Positive(t, len(m.nodeGroups))
	require.Equal(t, len(svc.Clusters[clusterID.String()].NodeGroups), len(m.nodeGroups))
}

func newMockService(clusterID uuid.UUID) *mocks.UpCloudService {
	return &mocks.UpCloudService{
		Clusters: map[string]upcloud.KubernetesCluster{
			clusterID.String(): {
				UUID: clusterID.String(),
				Plan: "dev",
				NodeGroups: []upcloud.KubernetesNodeGroup{
					{
						Count: 2,
						Name:  "group1",
						State: upcloud.KubernetesNodeGroupStateRunning,
					},
					{
						Count: 3,
						Name:  "group2",
						State: upcloud.KubernetesNodeGroupStateRunning,
					},
				},
			},
		},
		Plans: []upcloud.KubernetesPlan{{
			Name:     "dev",
			MaxNodes: 20,
		}},
	}
}
