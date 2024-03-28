/*
Copyright 2017 The Kubernetes Authors.

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

package gpu

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/util"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/stretchr/testify/assert"
)

const (
	GPULabel = "TestGPULabel/accelerator"
)

func TestGetGpuInfoForMetrics(t *testing.T) {
	type testCase struct {
		desc         string
		config       *cloudprovider.GpuConfig
		gpus         map[string]struct{}
		node         *apiv1.Node
		templateNode *apiv1.Node
		wantResource string
		wantLabel    string
	}
	var tests = []testCase{
		{
			desc:         "no gpu config",
			node:         newNode(nil),
			templateNode: newNode(nil),
		},
		{
			desc: "gpu config but no gpu used",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "gpu-type-value",
				ResourceName: util.ResourceNvidiaGPU,
			},
			gpus: map[string]struct{}{
				"gpu-type-value": {},
			},
			node:         newNode(nil),
			templateNode: newNode(nil),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "unexpected-label",
		},
		{
			desc: "gpu type and no resources",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "gpu-type-value",
				ResourceName: util.ResourceNvidiaGPU,
			},
			gpus: map[string]struct{}{
				"gpu-type-value": {},
			},
			node:         newNode(nil),
			templateNode: newNode(nil),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "unexpected-label",
		},
		{
			desc: "unspecified gpu",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "",
				ResourceName: util.ResourceNvidiaGPU,
			},
			gpus: map[string]struct{}{
				"gpu-type-value": {},
			},
			node:         newNode(resource.NewQuantity(10, resource.DecimalSI)),
			templateNode: newNode(resource.NewQuantity(10, resource.DecimalSI)),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "generic",
		},
		{
			desc: "invalid gpu type",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "no-such-gpu",
				ResourceName: util.ResourceNvidiaGPU,
			},
			gpus: map[string]struct{}{
				"gpu-type-value": {},
			},
			node:         newNode(resource.NewQuantity(10, resource.DecimalSI)),
			templateNode: newNode(resource.NewQuantity(10, resource.DecimalSI)),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "not-listed",
		},
		{
			desc: "happy scenario",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "gpu-type-value",
				ResourceName: util.ResourceNvidiaGPU,
			},
			gpus: map[string]struct{}{
				"gpu-type-value": {},
			},
			node:         newNode(resource.NewQuantity(10, resource.DecimalSI)),
			templateNode: newNode(resource.NewQuantity(10, resource.DecimalSI)),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "gpu-type-value",
		},
		{
			desc: "missing gpu",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "gpu-type-value",
				ResourceName: util.ResourceNvidiaGPU,
			},
			gpus: map[string]struct{}{
				"gpu-type-value": {},
			},
			node:         newNode(nil),
			templateNode: newNode(resource.NewQuantity(10, resource.DecimalSI)),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "missing-gpu",
		},
		{
			desc: "no capacity",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "gpu-type-value",
				ResourceName: util.ResourceNvidiaGPU,
			},
			gpus: map[string]struct{}{
				"gpu-type-value": {},
			},
			node:         newNode(nil),
			templateNode: newNode(nil),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "unexpected-label",
		},
		{
			desc: "no capacity and no node group",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "gpu-type-value",
				ResourceName: util.ResourceNvidiaGPU,
			},
			gpus: map[string]struct{}{
				"gpu-type-value": {},
			},
			node:         newNode(nil),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "unexpected-label",
		},
		{
			desc: "no gpu labels but allocatable defined",
			config: &cloudprovider.GpuConfig{
				Label:        "label/gpu-type",
				Type:         "gpu-type-value",
				ResourceName: util.ResourceNvidiaGPU,
			},
			node:         newNode(resource.NewQuantity(10, resource.DecimalSI)),
			wantResource: util.ResourceNvidiaGPU,
			wantLabel:    "not-listed",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			var nodeGroup cloudprovider.NodeGroup
			if tc.templateNode != nil {
				var nodeInfo = schedulerframework.NewNodeInfo()
				nodeInfo.SetNode(tc.templateNode)

				provider := testprovider.NewTestAutoprovisioningCloudProvider(
					func(_ string, _ int) error { return nil },
					func(_ string, _ string) error { return nil },
					func(_ string) error { return nil },
					func(_ string) error { return nil },
					[]string{},
					map[string]*schedulerframework.NodeInfo{
						"my-node-group": nodeInfo,
					})

				provider.AddNodeGroup("my-node-group", 0, 100, 1)
				nodeGroup = provider.GetNodeGroup("my-node-group")
			}

			gotResource, gotLabel := GetGpuInfoForMetrics(tc.config, tc.gpus, tc.node, nodeGroup)

			assert.Equal(t, tc.wantResource, gotResource)
			assert.Equal(t, tc.wantLabel, gotLabel)
		})
	}
}

func newNode(gpuCapacity *resource.Quantity) *apiv1.Node {
	status := apiv1.NodeStatus{
		Capacity:    apiv1.ResourceList{},
		Allocatable: apiv1.ResourceList{},
	}
	if gpuCapacity != nil {
		status.Capacity[util.ResourceNvidiaGPU] = *gpuCapacity
		status.Allocatable[util.ResourceNvidiaGPU] = *gpuCapacity
	}
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "my-node",
			Labels: map[string]string{},
		},
		Status: status,
	}
}
