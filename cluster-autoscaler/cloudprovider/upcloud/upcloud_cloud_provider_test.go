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
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/mocks"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
)

func TestUpCloudCloudProvider_NodeGroups(t *testing.T) {
	t.Parallel()

	clusterID := uuid.New()
	svc := newMockService(clusterID)
	p := newUpCloudCloudProvider(clusterID, svc)
	require.NoError(t, p.Refresh(t.Context()))
	require.NoError(t, svc.AppendNodeGroup(context.TODO(), clusterID, upcloud.KubernetesNodeGroup{Count: 3, Name: "group3"}))
	// node group length should still be 2 as refresh is not yet called
	require.Len(t, p.NodeGroups(t.Context()), 2)
	require.NoError(t, p.Refresh(t.Context()))
	require.Len(t, p.NodeGroups(t.Context()), 3)
}

func TestUpCloudCloudProvider_Name(t *testing.T) {
	t.Parallel()

	p := upCloudCloudProvider{}
	require.Equal(t, ProviderName, p.Name())
}

func TestUpCloudCloudProvider_NodeGroupForNode(t *testing.T) {
	t.Parallel()

	clusterID := uuid.New()
	svc := newMockService(clusterID)
	p := newUpCloudCloudProvider(clusterID, svc)
	require.NoError(t, p.Refresh(t.Context()))

	group, err := p.NodeGroupForNode(t.Context(), &v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: fmt.Sprintf("upcloud:////%s", "group1-1"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, group)
	require.Equal(t, fmt.Sprintf("%s/group1", clusterID.String()), group.Id())

	// nodes of other providers should return `nil` as group and error
	group, err = p.NodeGroupForNode(t.Context(), &v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: fmt.Sprintf("fake:////%s", "group1-1"),
		},
	})
	require.NoError(t, err)
	require.Nil(t, group)
}

func TestUpCloudCloudProvider_GetResourceLimiter(t *testing.T) {
	t.Parallel()

	rl := cloudprovider.NewResourceLimiter(map[string]int64{"min": 1}, nil)
	p := upCloudCloudProvider{
		manager: &manager{
			clusterID: uuid.New(),
		},
		resourceLimiter: rl,
	}
	l, err := p.GetResourceLimiter(t.Context())
	require.NoError(t, err)
	require.Equal(t, rl, l)
}

func TestBuildCloudConfig(t *testing.T) {
	want := upCloudConfig{
		ClusterID: uuid.NewString(),
		Username:  "uks-username",
		Password:  "uks-passwd",
		UserAgent: "uks-agent",
	}
	_, err := buildCloudConfig(config.AutoscalingOptions{UserAgent: want.UserAgent})
	require.Error(t, err)

	t.Setenv(envUpCloudClusterID, want.ClusterID)
	_, err = buildCloudConfig(config.AutoscalingOptions{UserAgent: want.UserAgent})
	require.Error(t, err)

	t.Setenv(envUpCloudUsername, want.Username)
	_, err = buildCloudConfig(config.AutoscalingOptions{UserAgent: want.UserAgent})
	require.Error(t, err)

	t.Setenv(envUpCloudPassword, want.Password)
	got, err := buildCloudConfig(config.AutoscalingOptions{UserAgent: want.UserAgent})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestBuildCloudConfig_WithToken(t *testing.T) {
	t.Setenv(envUpCloudClusterID, uuid.NewString())
	t.Setenv(envUpCloudToken, "ucat_test_token")

	got, err := buildCloudConfig(config.AutoscalingOptions{UserAgent: "uks-agent"})
	require.NoError(t, err)
	require.Equal(t, upCloudConfig{
		ClusterID: got.ClusterID,
		Token:     "ucat_test_token",
		UserAgent: "uks-agent",
	}, got)
}

func TestBuildCloudConfig_PrefersTokenOverBasicAuth(t *testing.T) {
	t.Setenv(envUpCloudClusterID, uuid.NewString())
	t.Setenv(envUpCloudToken, "ucat_test_token")
	t.Setenv(envUpCloudUsername, "uks-username")
	t.Setenv(envUpCloudPassword, "uks-passwd")

	got, err := buildCloudConfig(config.AutoscalingOptions{UserAgent: "uks-agent"})
	require.NoError(t, err)
	require.Equal(t, "ucat_test_token", got.Token)
	require.Empty(t, got.Username)
	require.Empty(t, got.Password)
}

func TestNewUpCloudService_WithToken(t *testing.T) {
	_, err := newUpCloudService(upCloudConfig{Token: "ucat_test_token"})
	require.NoError(t, err)
}

func TestUpCloudCloudProvider_GPULabel(t *testing.T) {
	t.Parallel()

	p := upCloudCloudProvider{}
	require.Empty(t, p.GPULabel(t.Context()))
}

func TestUpCloudCloudProvider_GetAvailableGPUTypes(t *testing.T) {
	t.Parallel()

	p := upCloudCloudProvider{}
	require.Nil(t, p.GetAvailableGPUTypes(t.Context()))
}

func TestUpCloudCloudProvider_Cleanup(t *testing.T) {
	t.Parallel()

	p := upCloudCloudProvider{}
	require.Nil(t, p.Cleanup(t.Context()))
}

func TestUpCloudCloudProvider_GetNodeGpuConfig(t *testing.T) {
	t.Parallel()

	p := upCloudCloudProvider{}
	require.Nil(t, p.GetNodeGpuConfig(t.Context(), &v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: fmt.Sprintf("upcloud:////%s", uuid.NewString()),
		},
	}))
}

func TestUpCloudCloudProvider_ErrNotImplemented(t *testing.T) {
	t.Parallel()

	p := upCloudCloudProvider{}

	_, err := p.HasInstance(t.Context(), nil)
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)

	_, err = p.Pricing(t.Context())
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)

	_, err = p.GetAvailableMachineTypes(t.Context())
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)

	_, err = p.NewNodeGroup(t.Context(), "", nil, nil, nil, nil)
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}

func newUpCloudCloudProvider(clusterID uuid.UUID, svc *mocks.UpCloudService) upCloudCloudProvider {
	if svc == nil {
		svc = &mocks.UpCloudService{}
	}
	return upCloudCloudProvider{
		manager: &manager{
			clusterID: clusterID,
			svc:       svc,
		},
		resourceLimiter: cloudprovider.NewResourceLimiter(map[string]int64{"min": 1}, nil),
	}
}
