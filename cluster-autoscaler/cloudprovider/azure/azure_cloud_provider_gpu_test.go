/*
Copyright The Kubernetes Authors.

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

package azure

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	cloudprovidermocks "sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider/mocks"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/customresources"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodeinfosprovider"
	drasnapshot "sigs.k8s.io/cluster-autoscaler/pkg/simulator/dynamicresources/snapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/gpu"
)

func TestGetNodeGpuConfig(t *testing.T) {
	provider := &AzureCloudProvider{}
	devicePluginConfig := &cloudprovider.GpuConfig{
		Label: GPULabel, Type: "nvidia", ExtendedResourceName: gpu.ResourceNvidiaGPU,
	}
	draConfig := &cloudprovider.GpuConfig{
		Label: GPULabel, Type: "nvidia", DraDriverName: "gpu.nvidia.com",
	}
	tests := []struct {
		name        string
		labels      map[string]string
		allocatable apiv1.ResourceList
		want        *cloudprovider.GpuConfig
	}{
		{
			name: "CPU node",
		},
		{
			name:   "DRA marker alone does not invent a GPU",
			labels: map[string]string{DraGPULabel: "true"},
		},
		{
			name:   "GPU without device plugin is not inferred to use DRA",
			labels: map[string]string{GPULabel: "nvidia"},
			want:   devicePluginConfig,
		},
		{
			name:   "DRA GPU without extended resources",
			labels: map[string]string{GPULabel: "nvidia", "kubernetes.azure.com/gpu-dra-driver": "true"},
			want:   draConfig,
		},
		{
			name:        "explicit DRA mode overrides extended resource detection",
			labels:      map[string]string{GPULabel: "nvidia", DraGPULabel: "true"},
			allocatable: apiv1.ResourceList{gpu.ResourceNvidiaGPU: resource.MustParse("1")},
			want:        draConfig,
		},
		{
			name:   "DRA disabled",
			labels: map[string]string{GPULabel: "nvidia", DraGPULabel: "false"},
			want:   devicePluginConfig,
		},
		{
			name:   "empty DRA marker",
			labels: map[string]string{GPULabel: "nvidia", DraGPULabel: ""},
			want:   devicePluginConfig,
		},
		{
			name:   "DRA marker requires exact true value",
			labels: map[string]string{GPULabel: "nvidia", DraGPULabel: "TRUE"},
			want:   devicePluginConfig,
		},
		{
			name:        "device plugin GPU",
			labels:      map[string]string{GPULabel: "nvidia"},
			allocatable: apiv1.ResourceList{gpu.ResourceNvidiaGPU: resource.MustParse("1")},
			want:        devicePluginConfig,
		},
		{
			name:        "GPU allocatable without GPU label",
			allocatable: apiv1.ResourceList{gpu.ResourceNvidiaGPU: resource.MustParse("1")},
			want:        &cloudprovider.GpuConfig{Label: GPULabel, ExtendedResourceName: gpu.ResourceNvidiaGPU},
		},
		{
			name:        "DirectX device plugin",
			labels:      map[string]string{GPULabel: "nvidia"},
			allocatable: apiv1.ResourceList{gpu.ResourceDirectX: resource.MustParse("1")},
			want:        &cloudprovider.GpuConfig{Label: GPULabel, Type: "nvidia", ExtendedResourceName: gpu.ResourceDirectX},
		},
		{
			name:        "AMD device plugin",
			labels:      map[string]string{GPULabel: "amd"},
			allocatable: apiv1.ResourceList{gpu.ResourceAMDGPU: resource.MustParse("1")},
			want:        &cloudprovider.GpuConfig{Label: GPULabel, Type: "amd", ExtendedResourceName: gpu.ResourceAMDGPU},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node := &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: tc.labels},
				Status:     apiv1.NodeStatus{Allocatable: tc.allocatable},
			}
			original := node.DeepCopy()
			assert.Equal(t, tc.want, provider.GetNodeGpuConfig(context.Background(), node))
			assert.Equal(t, original, node)
		})
	}
}

func TestAzureGpuProcessor(t *testing.T) {
	ctx := context.Background()
	autoscalingCtx := &ca_context.AutoscalingContext{CloudProvider: &AzureCloudProvider{}}
	processor := &customresources.GpuCustomResourcesProcessor{}
	tests := []struct {
		name         string
		labels       map[string]string
		allocatable  apiv1.ResourceList
		templateGPUs int64
		wantReady    bool
		wantTarget   customresources.CustomResourceTarget
	}{
		{
			name:      "unmanaged DRA GPU",
			labels:    map[string]string{GPULabel: "nvidia", DraGPULabel: "true"},
			wantReady: true,
		},
		{
			name:         "managed GPU with missing device plugin",
			labels:       map[string]string{GPULabel: "nvidia"},
			templateGPUs: 2,
			wantTarget:   customresources.CustomResourceTarget{ResourceType: "nvidia", ResourceCount: 2},
		},
		{
			name:         "device plugin reporting zero GPUs",
			labels:       map[string]string{GPULabel: "nvidia", DraGPULabel: "false"},
			allocatable:  apiv1.ResourceList{gpu.ResourceNvidiaGPU: resource.MustParse("0")},
			templateGPUs: 2,
			wantTarget:   customresources.CustomResourceTarget{ResourceType: "nvidia", ResourceCount: 2},
		},
		{
			name:        "device plugin reporting GPUs",
			labels:      map[string]string{GPULabel: "nvidia"},
			allocatable: apiv1.ResourceList{gpu.ResourceNvidiaGPU: resource.MustParse("2")},
			wantReady:   true,
			wantTarget:  customresources.CustomResourceTarget{ResourceType: "nvidia", ResourceCount: 2},
		},
		{
			name:        "DirectX device plugin",
			labels:      map[string]string{GPULabel: "nvidia"},
			allocatable: apiv1.ResourceList{gpu.ResourceDirectX: resource.MustParse("1")},
			wantReady:   true,
			wantTarget:  customresources.CustomResourceTarget{ResourceType: "nvidia", ResourceCount: 1},
		},
		{
			name:      "CPU node",
			wantReady: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node := &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: tc.name, Labels: tc.labels},
				Status: apiv1.NodeStatus{
					Allocatable: tc.allocatable,
					Conditions:  []apiv1.NodeCondition{{Type: apiv1.NodeReady, Status: apiv1.ConditionTrue}},
				},
			}
			original := node.DeepCopy()
			nodes := []*apiv1.Node{node}
			allNodes, readyNodes := processor.FilterOutNodesWithUnreadyResources(ctx, autoscalingCtx, nodes, nodes, nil, nil)
			require.Len(t, allNodes, 1)
			assert.Equal(t, tc.wantReady, len(readyNodes) == 1)
			assert.Equal(t, tc.wantReady, allNodes[0].Status.Conditions[0].Status == apiv1.ConditionTrue)
			var nodeGroup cloudprovider.NodeGroup
			if tc.templateGPUs > 0 {
				templateNode := node.DeepCopy()
				templateNode.Status.Capacity = apiv1.ResourceList{
					gpu.ResourceNvidiaGPU: *resource.NewQuantity(tc.templateGPUs, resource.DecimalSI),
				}
				group := &cloudprovidermocks.NodeGroup{}
				group.On("TemplateNodeInfo", mock.Anything).Return(framework.NewTestNodeInfo(templateNode), nil).Once()
				t.Cleanup(func() { group.AssertExpectations(t) })
				nodeGroup = group
			}
			targets, err := processor.GetNodeResourceTargets(ctx, autoscalingCtx, node, nodeGroup)
			require.NoError(t, err)
			assert.Equal(t, []customresources.CustomResourceTarget{tc.wantTarget}, targets)
			assert.Equal(t, original, node)
		})
	}
}

func TestAzureDraReadiness(t *testing.T) {
	ctx := context.Background()
	templateSlice := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "template-slice"},
		Spec: resourceapi.ResourceSliceSpec{
			Driver:   NvidiaDraGPUDriver,
			NodeName: ptr.To("template"),
			Pool:     resourceapi.ResourcePool{Name: "template", Generation: 1, ResourceSliceCount: 1},
			Devices:  []resourceapi.Device{{Name: "gpu-0"}, {Name: "gpu-1"}},
		},
	}
	firstSlice := templateSlice.DeepCopy()
	firstSlice.Name = "node-slice-0"
	firstSlice.Spec.NodeName = ptr.To("gpu-node")
	firstSlice.Spec.Pool.Name = "gpu-node"
	firstSlice.Spec.Pool.ResourceSliceCount = 2
	firstSlice.Spec.Devices = firstSlice.Spec.Devices[:1]
	secondSlice := firstSlice.DeepCopy()
	secondSlice.Name = "node-slice-1"
	secondSlice.Spec.Devices[0].Name = "gpu-1"

	tests := []struct {
		name      string
		draLabel  string
		nodeReady bool
		slices    []*resourceapi.ResourceSlice
		wantReady bool
	}{
		{
			name: "complete DRA pool", draLabel: "true", nodeReady: true,
			slices: []*resourceapi.ResourceSlice{firstSlice, secondSlice}, wantReady: true,
		},
		{
			name: "incomplete DRA pool", draLabel: "true", nodeReady: true,
			slices: []*resourceapi.ResourceSlice{firstSlice},
		},
		{
			name: "DRA slices not yet published", draLabel: "true", nodeReady: true,
		},
		{
			name: "slices do not replace explicit DRA marker", nodeReady: true,
			slices: []*resourceapi.ResourceSlice{firstSlice, secondSlice},
		},
		{
			name: "DRA marker does not override Kubernetes readiness", draLabel: "true",
			slices: []*resourceapi.ResourceSlice{firstSlice, secondSlice},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node := &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "gpu-node",
					Labels: map[string]string{GPULabel: "nvidia", DraGPULabel: tc.draLabel},
				},
				Status: apiv1.NodeStatus{
					Conditions: []apiv1.NodeCondition{{Type: apiv1.NodeReady, Status: apiv1.ConditionFalse}},
				},
			}
			var readyNodes []*apiv1.Node
			if tc.nodeReady {
				node.Status.Conditions[0].Status = apiv1.ConditionTrue
				readyNodes = []*apiv1.Node{node}
			}
			original := node.DeepCopy()
			nodeGroup := &cloudprovidermocks.NodeGroup{}
			if tc.draLabel == "true" && tc.nodeReady {
				nodeGroup.On("Id").Return("dra-pool").Once()
				nodeGroup.On("TemplateNodeInfo", mock.Anything).
					Return(framework.NewNodeInfo(node, []*resourceapi.ResourceSlice{templateSlice}), nil).Once()
			}
			provider := &azureDraTestProvider{AzureCloudProvider: &AzureCloudProvider{}, nodeGroup: nodeGroup}
			autoscalingCtx := &ca_context.AutoscalingContext{
				CloudProvider:            provider,
				TemplateNodeInfoRegistry: nodeinfosprovider.NewTemplateNodeInfoRegistry(nil),
			}
			snapshot := drasnapshot.NewSnapshot(nil, map[string][]*resourceapi.ResourceSlice{node.Name: tc.slices}, nil, nil)
			processor := customresources.NewDefaultCustomResourcesProcessor(true, false)
			defer processor.CleanUp()
			allNodes, readyNodes := processor.FilterOutNodesWithUnreadyResources(ctx, autoscalingCtx, []*apiv1.Node{node}, readyNodes, snapshot, nil)
			require.Len(t, allNodes, 1)
			assert.Equal(t, tc.wantReady, len(readyNodes) == 1)
			assert.Equal(t, tc.wantReady, allNodes[0].Status.Conditions[0].Status == apiv1.ConditionTrue)
			assert.Equal(t, original, node)
			nodeGroup.AssertExpectations(t)
		})
	}
}

type azureDraTestProvider struct {
	*AzureCloudProvider
	nodeGroup cloudprovider.NodeGroup
}

func (p *azureDraTestProvider) NodeGroupForNode(context.Context, *apiv1.Node) (cloudprovider.NodeGroup, error) {
	return p.nodeGroup, nil
}
