/*
Copyright 2019 The Kubernetes Authors.

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

package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func TestNodeGroup_TargetSize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		numberOfNodes := 3

		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count: numberOfNodes,
		})

		size, err := ng.TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, numberOfNodes, size, "target size is not correct")
	})
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		numberOfNodes := 3
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count:    numberOfNodes,
			MinNodes: 1,
			MaxNodes: 10,
		})

		delta := 2

		newCount := numberOfNodes + delta
		client.On("UpdateNodePool",
			ctx,
			ng.clusterID,
			ng.id,
			&godo.KubernetesNodePoolUpdateRequest{
				Count: &newCount,
			},
		).Return(
			&godo.KubernetesNodePool{Count: newCount},
			&godo.Response{},
			nil,
		).Once()

		err := ng.IncreaseSize(delta)
		assert.NoError(t, err)
	})

	t.Run("successful increase to maximum", func(t *testing.T) {
		numberOfNodes := 2
		maxNodes := 3
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count:     numberOfNodes,
			AutoScale: true,
			MinNodes:  1,
			MaxNodes:  maxNodes,
		})

		delta := 1
		newCount := numberOfNodes + delta
		client.On("UpdateNodePool",
			ctx,
			ng.clusterID,
			ng.id,
			&godo.KubernetesNodePoolUpdateRequest{
				Count: &newCount,
			},
		).Return(
			&godo.KubernetesNodePool{Count: newCount},
			&godo.Response{},
			nil,
		).Once()

		err := ng.IncreaseSize(delta)
		assert.NoError(t, err)
	})

	t.Run("negative increase", func(t *testing.T) {
		numberOfNodes := 3
		delta := -1
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count: numberOfNodes,
		})
		err := ng.IncreaseSize(delta)

		exp := fmt.Errorf("delta must be positive, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("zero increase", func(t *testing.T) {
		numberOfNodes := 3
		delta := 0
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count: numberOfNodes,
		})

		exp := fmt.Errorf("delta must be positive, have: %d", delta)

		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("large increase above maximum", func(t *testing.T) {
		numberOfNodes := 195
		delta := 10

		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count: numberOfNodes,
		})

		exp := fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			numberOfNodes, numberOfNodes+delta, ng.MaxSize())

		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase is too large")
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		numberOfNodes := 5
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count: numberOfNodes,
		})

		delta := -2

		newCount := numberOfNodes + delta
		client.On("UpdateNodePool",
			ctx,
			ng.clusterID,
			ng.id,
			&godo.KubernetesNodePoolUpdateRequest{
				Count: &newCount,
			},
		).Return(
			&godo.KubernetesNodePool{Count: newCount},
			&godo.Response{},
			nil,
		).Once()

		err := ng.DecreaseTargetSize(delta)
		assert.NoError(t, err)
	})

	t.Run("successful decrease to minimum", func(t *testing.T) {
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count:     2,
			AutoScale: true,
			MinNodes:  1,
			MaxNodes:  5,
		})

		delta := -1
		newCount := ng.nodePool.Count + delta
		client.On("UpdateNodePool",
			ctx,
			ng.clusterID,
			ng.id,
			&godo.KubernetesNodePoolUpdateRequest{
				Count: &newCount,
			},
		).Return(
			&godo.KubernetesNodePool{Count: newCount},
			&godo.Response{},
			nil,
		).Once()

		err := ng.DecreaseTargetSize(delta)
		assert.NoError(t, err)
	})

	t.Run("positive decrease", func(t *testing.T) {
		numberOfNodes := 5
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count: numberOfNodes,
		})

		delta := 1
		err := ng.DecreaseTargetSize(delta)

		exp := fmt.Errorf("delta must be negative, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("zero decrease", func(t *testing.T) {
		numberOfNodes := 5
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count: numberOfNodes,
		})

		delta := 0
		exp := fmt.Errorf("delta must be negative, have: %d", delta)

		err := ng.DecreaseTargetSize(delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("small decrease below minimum", func(t *testing.T) {
		delta := -2
		numberOfNodes := 3
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count:    numberOfNodes,
			MinNodes: 2,
			MaxNodes: 5,
		})

		exp := fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			numberOfNodes, numberOfNodes+delta, ng.MinSize())
		err := ng.DecreaseTargetSize(delta)
		assert.EqualError(t, err, exp.Error(), "size decrease is too small")
	})
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count: 3,
		})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "1"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "2"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "3"}}},
		}

		// this should be called three times (the number of nodes)
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "1", nil).Return(&godo.Response{}, nil).Once()
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "2", nil).Return(&godo.Response{}, nil).Once()
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "3", nil).Return(&godo.Response{}, nil).Once()

		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
	})

	t.Run("client deleting node fails", func(t *testing.T) {
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{Count: 3})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "1"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "2"}}},
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "3"}}},
		}

		// client is called twice, first run is successfully but the second one
		// fails with a random error. In this case DeleteNodes() should return
		// immediately.
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "1", nil).
			Return(&godo.Response{}, nil).Once()
		client.On("DeleteNode", ctx, ng.clusterID, ng.id, "2", nil).
			Return(&godo.Response{}, errors.New("random error")).Once()

		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Nodes: []*godo.KubernetesNode{
				{
					ID: "1",
					Status: &godo.KubernetesNodeStatus{
						State: "provisioning",
					},
					DropletID: "droplet-1",
				},
				{
					ID: "2",
					Status: &godo.KubernetesNodeStatus{
						State: "running",
					},
					DropletID: "droplet-2",
				},
				{
					ID: "3",
					Status: &godo.KubernetesNodeStatus{
						State: "deleting",
					},
					DropletID: "droplet-3",
				},
				{
					ID: "4",
					Status: &godo.KubernetesNodeStatus{
						State:   "unknown",
						Message: "some-message",
					},
					DropletID: "droplet-4",
				},
				{
					// no status
					ID:        "5",
					DropletID: "droplet-5",
				},
			},
			Count: 5,
		})

		exp := []cloudprovider.Instance{
			{
				Id: "digitalocean://droplet-1",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			},
			{
				Id: "digitalocean://droplet-2",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceRunning,
				},
			},
			{
				Id: "digitalocean://droplet-3",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceDeleting,
				},
			},
			{
				Id: "digitalocean://droplet-4",
				Status: &cloudprovider.InstanceStatus{
					ErrorInfo: &cloudprovider.InstanceErrorInfo{
						ErrorClass:   cloudprovider.OtherErrorClass,
						ErrorCode:    "no-code-digitalocean",
						ErrorMessage: "some-message",
					},
				},
			},
			{
				Id: "digitalocean://droplet-5",
			},
		}

		nodes, err := ng.Nodes()
		assert.NoError(t, err)
		assert.Equal(t, exp, nodes, "nodes do not match")
	})

	t.Run("failure (nil node pool)", func(t *testing.T) {
		client := &doClientMock{}
		ng := testNodeGroup(client, nil)

		_, err := ng.Nodes()
		assert.Error(t, err, "Nodes() should return an error")
	})
}

func TestNodeGroup_Debug(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{
			Count:    3,
			MinNodes: 1,
			MaxNodes: 200,
		})

		d := ng.Debug()
		exp := "cluster ID: 1 (min:1 max:200)"
		assert.Equal(t, exp, d, "debug string do not match")
	})
}

func TestNodeGroup_Exist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &doClientMock{}
		ng := testNodeGroup(client, &godo.KubernetesNodePool{Count: 3})

		exist := ng.Exist()
		assert.Equal(t, true, exist, "node pool should exist")
	})

	t.Run("failure", func(t *testing.T) {
		client := &doClientMock{}
		ng := testNodeGroup(client, nil)

		exist := ng.Exist()
		assert.Equal(t, false, exist, "node pool should not exist")
	})
}

func testNodeGroup(client nodeGroupClient, np *godo.KubernetesNodePool) *NodeGroup {
	var minNodes, maxNodes int
	if np != nil {
		minNodes = np.MinNodes
		maxNodes = np.MaxNodes
	}
	return &NodeGroup{
		id:        "1",
		clusterID: "1",
		client:    client,
		nodePool:  np,
		minSize:   minNodes,
		maxSize:   maxNodes,
	}
}

func TestGenerateWorkerName(t *testing.T) {
	t.Run("generate worker node name", func(t *testing.T) {
		prefix := "testpool"
		expectedLength := generatedWorkerNameSuffixLength
		g := generateWorkerName(prefix)
		parts := strings.Split(g, "-")
		assert.Equal(t, 2, len(parts), "incorrect number of components for generated worker name")
		assert.Equal(t, prefix, parts[0], "unexpected prefix in generated worker name")
		assert.Equal(t, expectedLength, len(parts[1]), "incorrect suffix length for generated worker name")
	})
}

func TestParseToQuantity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rl, err := parseToQuantity(3800, 110, "8Gi")
		require.NoError(t, err)

		expCPU := resource.NewMilliQuantity(3800, resource.DecimalSI)
		expPods := resource.NewQuantity(110, resource.DecimalSI)
		expMem := resource.MustParse("8Gi")

		assert.True(t, expCPU.Equal(rl[apiv1.ResourceCPU]))
		assert.True(t, expPods.Equal(rl[apiv1.ResourcePods]))
		assert.True(t, expMem.Equal(rl[apiv1.ResourceMemory]))

		// compare with regular CPU quantity
		rl, err = parseToQuantity(4000, 110, "8Gi")
		require.NoError(t, err)
		fullCPU := resource.NewQuantity(4, resource.DecimalSI)
		assert.True(t, fullCPU.Equal(rl[apiv1.ResourceCPU]))
	})

	t.Run("invalid memory string", func(t *testing.T) {
		_, err := parseToQuantity(100, 10, "not-a-valid-quantity")
		assert.Error(t, err)
	})
}

func TestToNodeInfoTemplate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		resp := &godo.KubernetesNodePoolTemplate{
			Template: &godo.KubernetesNodeTemplate{
				Name: "my-pool",
				Labels: map[string]string{
					"pool": "primary",
				},
				Taints: []string{"dedicated=test:NoSchedule"},
				Capacity: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 8000,
					Pods:          110,
					Memory:        "32Gi",
				},
				Allocatable: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 7800,
					Pods:          110,
					Memory:        "31Gi",
				},
			},
		}

		ni, err := toNodeInfoTemplate(resp)
		require.NoError(t, err)
		require.NotNil(t, ni)

		node := ni.Node()
		require.NotNil(t, node)

		assert.True(t, strings.HasPrefix(node.Name, "my-pool-"))
		assert.Len(t, strings.TrimPrefix(node.Name, "my-pool-"), generatedWorkerNameSuffixLength)

		assert.Equal(t, cloudprovider.DefaultOS, node.Labels[apiv1.LabelOSStable])
		assert.Equal(t, cloudprovider.DefaultArch, node.Labels[apiv1.LabelArchStable])
		assert.Equal(t, "primary", node.Labels["pool"])

		require.Len(t, node.Spec.Taints, 1)
		assert.Equal(t, "dedicated", node.Spec.Taints[0].Key)
		assert.Equal(t, "test", node.Spec.Taints[0].Value)
		assert.Equal(t, apiv1.TaintEffectNoSchedule, node.Spec.Taints[0].Effect)

		assert.Equal(t, apiv1.NodeRunning, node.Status.Phase)

		capCPU := resource.NewMilliQuantity(8000, resource.DecimalSI)
		allCPU := resource.NewMilliQuantity(7800, resource.DecimalSI)
		assert.True(t, capCPU.Equal(node.Status.Capacity[apiv1.ResourceCPU]))
		assert.True(t, allCPU.Equal(node.Status.Allocatable[apiv1.ResourceCPU]))

		memCap := resource.MustParse("32Gi")
		memAll := resource.MustParse("31Gi")
		assert.True(t, memCap.Equal(node.Status.Capacity[apiv1.ResourceMemory]))
		assert.True(t, memAll.Equal(node.Status.Allocatable[apiv1.ResourceMemory]))

		podsCap := resource.NewQuantity(110, resource.DecimalSI)
		assert.True(t, podsCap.Equal(node.Status.Capacity[apiv1.ResourcePods]))
		assert.True(t, podsCap.Equal(node.Status.Allocatable[apiv1.ResourcePods]))

		assert.NotEmpty(t, node.Status.Conditions)
	})

	t.Run("template labels override defaults", func(t *testing.T) {
		resp := &godo.KubernetesNodePoolTemplate{
			Template: &godo.KubernetesNodeTemplate{
				Name: "pool",
				Labels: map[string]string{
					apiv1.LabelOSStable: "windows",
				},
				Capacity: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 2000,
					Pods:          30,
					Memory:        "4Gi",
				},
				Allocatable: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 1900,
					Pods:          30,
					Memory:        "4Gi",
				},
			},
		}

		ni, err := toNodeInfoTemplate(resp)
		require.NoError(t, err)
		assert.Equal(t, "windows", ni.Node().Labels[apiv1.LabelOSStable])
		assert.Equal(t, cloudprovider.DefaultArch, ni.Node().Labels[apiv1.LabelArchStable])
	})

	t.Run("success with NVIDIA GPU", func(t *testing.T) {
		resp := &godo.KubernetesNodePoolTemplate{
			Template: &godo.KubernetesNodeTemplate{
				Name: "gpu-pool",
				Capacity: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 4000,
					Pods:          110,
					Memory:        "16Gi",
				},
				Allocatable: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 3800,
					Pods:          110,
					Memory:        "15Gi",
				},
				Gpu: &godo.KubernetesNodePoolGPUResources{
					Vendor: "nvidia",
					Model:  "a100",
					Count:  4,
				},
			},
		}

		ni, err := toNodeInfoTemplate(resp)
		require.NoError(t, err)

		node := ni.Node()
		gpuName := apiv1.ResourceName("nvidia.com/gpu")
		expGPU := resource.MustParse("4")
		assert.True(t, expGPU.Equal(node.Status.Capacity[gpuName]))
		assert.True(t, expGPU.Equal(node.Status.Allocatable[gpuName]))
	})

	t.Run("success with AMD GPU", func(t *testing.T) {
		resp := &godo.KubernetesNodePoolTemplate{
			Template: &godo.KubernetesNodeTemplate{
				Name: "amd-gpu-pool",
				Capacity: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 8000,
					Pods:          110,
					Memory:        "32Gi",
				},
				Allocatable: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 7900,
					Pods:          110,
					Memory:        "31Gi",
				},
				Gpu: &godo.KubernetesNodePoolGPUResources{
					Vendor: "amd",
					Model:  "mi300x",
					Count:  8,
				},
			},
		}

		ni, err := toNodeInfoTemplate(resp)
		require.NoError(t, err)

		node := ni.Node()
		gpuName := apiv1.ResourceName("amd.com/gpu")
		expGPU := resource.MustParse("8")
		assert.True(t, expGPU.Equal(node.Status.Capacity[gpuName]))
		assert.True(t, expGPU.Equal(node.Status.Allocatable[gpuName]))
	})

	t.Run("allocatable parse error", func(t *testing.T) {
		resp := &godo.KubernetesNodePoolTemplate{
			Template: &godo.KubernetesNodeTemplate{
				Name: "pool",
				Capacity: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 1000,
					Pods:          10,
					Memory:        "1Gi",
				},
				Allocatable: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 900,
					Pods:          10,
					Memory:        "NOT_VALID",
				},
			},
		}

		_, err := toNodeInfoTemplate(resp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create allocatable resources")
	})

	t.Run("capacity parse error", func(t *testing.T) {
		resp := &godo.KubernetesNodePoolTemplate{
			Template: &godo.KubernetesNodeTemplate{
				Name: "pool",
				Capacity: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 1000,
					Pods:          10,
					Memory:        "NOT_VALID",
				},
				Allocatable: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 900,
					Pods:          10,
					Memory:        "1Gi",
				},
			},
		}

		_, err := toNodeInfoTemplate(resp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create capacity resources")
	})

	t.Run("taints parse error", func(t *testing.T) {
		resp := &godo.KubernetesNodePoolTemplate{
			Template: &godo.KubernetesNodeTemplate{
				Name: "pool",
				Capacity: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 1000,
					Pods:          10,
					Memory:        "1Gi",
				},
				Allocatable: &godo.KubernetesNodePoolResources{
					CpuMilliCores: 900,
					Pods:          10,
					Memory:        "1Gi",
				},
				Taints: []string{"this-is-not-a-valid-taint-string-format"},
			},
		}

		_, err := toNodeInfoTemplate(resp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse taints")
	})
}

type doClientMock struct {
	mock.Mock
}

func (m *doClientMock) ListNodePools(ctx context.Context, clusterID string, opts *godo.ListOptions) ([]*godo.KubernetesNodePool, *godo.Response, error) {
	args := m.Called(ctx, clusterID, nil)
	return args.Get(0).([]*godo.KubernetesNodePool), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *doClientMock) UpdateNodePool(ctx context.Context, clusterID, poolID string, req *godo.KubernetesNodePoolUpdateRequest) (*godo.KubernetesNodePool, *godo.Response, error) {
	args := m.Called(ctx, clusterID, poolID, req)
	return args.Get(0).(*godo.KubernetesNodePool), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *doClientMock) DeleteNode(ctx context.Context, clusterID, poolID, nodeID string, req *godo.KubernetesNodeDeleteRequest) (*godo.Response, error) {
	args := m.Called(ctx, clusterID, poolID, nodeID, nil)
	return args.Get(0).(*godo.Response), args.Error(1)
}

func (m *doClientMock) GetNodePoolTemplate(ctx context.Context, clusterID string, nodePoolName string) (*godo.KubernetesNodePoolTemplate, *godo.Response, error) {
	args := m.Called(ctx, clusterID, nodePoolName)
	return args.Get(0).(*godo.KubernetesNodePoolTemplate), args.Get(1).(*godo.Response), args.Error(2)
}
