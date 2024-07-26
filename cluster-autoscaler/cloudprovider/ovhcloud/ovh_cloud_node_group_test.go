/*
Copyright 2020 The Kubernetes Authors.

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

package ovhcloud

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud/sdk"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

func newTestNodeGroup(t *testing.T, flavor string) *NodeGroup {
	cfg := `{
		"project_id": "projectID",
		"cluster_id": "clusterID",
		"authentication_type": "consumer",
		"application_endpoint": "ovh-eu",
		"application_key": "key",
		"application_secret": "secret",
		"application_consumer_key": "consumer_key"
	}`

	manager, err := NewManager(bytes.NewBufferString(cfg))
	if err != nil {
		assert.FailNow(t, "failed to create manager", err)
	}

	client := &sdk.ClientMock{}
	ctx := context.Background()

	client.On("ListClusterFlavors", ctx, "projectID", "clusterID").Return(
		[]sdk.Flavor{
			{
				Name:     "b2-7",
				Category: "b",
				State:    "available",
				VCPUs:    2,
				GPUs:     0,
				RAM:      7,
			},
			{
				Name:     "t1-45",
				Category: "t",
				State:    "available",
				VCPUs:    8,
				GPUs:     1,
				RAM:      45,
			},
			{
				Name:     "unknown",
				Category: "",
				State:    "unavailable",
				VCPUs:    2,
				GPUs:     0,
				RAM:      7,
			},
		}, nil,
	)
	manager.Client = client

	ng := &NodeGroup{
		Manager: manager,
		NodePool: &sdk.NodePool{
			ID:           "id",
			Name:         fmt.Sprintf("pool-%s", flavor),
			Flavor:       flavor,
			Autoscale:    true,
			DesiredNodes: 3,
			MinNodes:     1,
			MaxNodes:     5,
			Autoscaling: &sdk.NodePoolAutoscaling{
				ScaleDownUtilizationThreshold: 3.2,
				ScaleDownUnneededTimeSeconds:  10,
				ScaleDownUnreadyTimeSeconds:   20,
			},
		},

		CurrentSize: 3,
	}

	return ng
}

func (ng *NodeGroup) mockCallUpdateNodePool(newDesiredNodes uint32, nodesToRemove []string) {
	ng.Manager.Client.(*sdk.ClientMock).On(
		"UpdateNodePool",
		context.Background(),
		ng.Manager.ProjectID,
		ng.Manager.ClusterID,
		ng.ID,
		&sdk.UpdateNodePoolOpts{
			DesiredNodes:  &newDesiredNodes,
			NodesToRemove: nodesToRemove,
		},
	).Return(
		&sdk.NodePool{
			ID:           ng.ID,
			Name:         ng.Name,
			Flavor:       ng.Flavor,
			Autoscale:    ng.Autoscale,
			DesiredNodes: newDesiredNodes,
			MinNodes:     ng.MinNodes,
			MaxNodes:     ng.MaxNodes,
		},
		nil,
	)
}

func (ng *NodeGroup) mockCallListNodePoolNodes() {
	ng.Manager.Client.(*sdk.ClientMock).On(
		"ListNodePoolNodes",
		context.Background(),
		ng.Manager.ProjectID,
		ng.Manager.ClusterID,
		ng.ID,
	).Return(
		[]sdk.Node{
			{
				ID:         "id-1",
				Name:       "node-1",
				Status:     "READY",
				InstanceID: "instance-1",
			},
			{
				ID:         "id-2",
				Name:       "node-2",
				Status:     "INSTALLING",
				InstanceID: "",
			},
			{
				ID:         "id-3",
				Name:       "node-3",
				Status:     "ERROR",
				InstanceID: "",
			},
		}, nil,
	)
}

func (ng *NodeGroup) mockCallCreateNodePool() {
	ng.Manager.Client.(*sdk.ClientMock).On(
		"CreateNodePool",
		context.Background(),
		ng.Manager.ProjectID,
		ng.Manager.ClusterID,
		&sdk.CreateNodePoolOpts{
			FlavorName:   "b2-7",
			Name:         &ng.Name,
			DesiredNodes: &ng.DesiredNodes,
			MinNodes:     &ng.MinNodes,
			MaxNodes:     &ng.MaxNodes,
			Autoscale:    ng.Autoscale,
		},
	).Return(
		&sdk.NodePool{
			ID:           ng.ID,
			Name:         ng.Name,
			Flavor:       ng.Flavor,
			Autoscale:    ng.Autoscale,
			DesiredNodes: ng.DesiredNodes,
			MinNodes:     ng.MinNodes,
			MaxNodes:     ng.MaxNodes,
		}, nil,
	)
}

func (ng *NodeGroup) mockCallDeleteNodePool() {
	ng.Manager.Client.(*sdk.ClientMock).On("DeleteNodePool", context.Background(), "projectID", "clusterID", "id").Return(&sdk.NodePool{}, nil)
}

func TestOVHCloudNodeGroup_MaxSize(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check default node group max size", func(t *testing.T) {
		max := ng.MaxSize()

		assert.Equal(t, 5, max)
	})
}

func TestOVHCloudNodeGroup_MinSize(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check default node group min size", func(t *testing.T) {
		min := ng.MinSize()

		assert.Equal(t, 1, min)
	})
}

func TestOVHCloudNodeGroup_TargetSize(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check default node group target size", func(t *testing.T) {
		targetSize, err := ng.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 3, targetSize)
	})
}

func TestOVHCloudNodeGroup_IncreaseSize(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check increase size below max size", func(t *testing.T) {
		ng.mockCallUpdateNodePool(4, nil)

		err := ng.IncreaseSize(1)
		assert.NoError(t, err)

		targetSize, err := ng.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 4, targetSize)
	})

	t.Run("check increase size above max size", func(t *testing.T) {
		err := ng.IncreaseSize(5)
		assert.Error(t, err)
	})

	t.Run("check increase size with negative delta", func(t *testing.T) {
		err := ng.IncreaseSize(-1)
		assert.Error(t, err)
	})
}

func TestOVHCloudNodeGroup_DeleteNodes(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check delete nodes above min size", func(t *testing.T) {
		ng.mockCallUpdateNodePool(2, []string{"openstack:///instance-1"})

		err := ng.DeleteNodes([]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"nodepool": ng.Id(),
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "openstack:///instance-1",
				},
			},
		})
		assert.NoError(t, err)

		targetSize, err := ng.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 2, targetSize)
	})

	t.Run("check delete nodes empty nodes array", func(t *testing.T) {
		ng.mockCallUpdateNodePool(2, []string{})

		err := ng.DeleteNodes(nil)
		assert.NoError(t, err)

		targetSize, err := ng.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, 2, targetSize)
	})

	t.Run("check delete nodes below min size", func(t *testing.T) {
		err := ng.DeleteNodes([]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"nodepool": ng.Id(),
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
					Labels: map[string]string{
						"nodepool": ng.Id(),
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-3",
					Labels: map[string]string{
						"nodepool": ng.Id(),
					},
				},
			},
		})
		assert.Error(t, err)
	})
}

func TestOVHCloudNodeGroup_DecreaseTargetSize(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")
	err := ng.DecreaseTargetSize(-1)
	assert.Error(t, err)
}

func TestOVHCloudNodeGroup_Id(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check node group id fetch", func(t *testing.T) {
		name := ng.Id()

		assert.Equal(t, "pool-b2-7", name)
	})
}

func TestOVHCloudNodeGroup_Debug(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check node group debug print", func(t *testing.T) {
		debug := ng.Debug()

		assert.Equal(t, "pool-b2-7 (3:1:5)", debug)
	})
}

func TestOVHCloudNodeGroup_Nodes(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check nodes list in node group", func(t *testing.T) {
		ng.mockCallListNodePoolNodes()

		nodes, err := ng.Nodes()
		assert.NoError(t, err)

		assert.Equal(t, 3, len(nodes))

		assert.Equal(t, "openstack:///instance-1", nodes[0].Id)
		assert.Equal(t, cloudprovider.InstanceRunning, nodes[0].Status.State)

		assert.Equal(t, "openstack:///", nodes[1].Id)
		assert.Equal(t, cloudprovider.InstanceCreating, nodes[1].Status.State)

		assert.Equal(t, "openstack:///", nodes[2].Id)
		assert.Equal(t, cloudprovider.OtherErrorClass, nodes[2].Status.ErrorInfo.ErrorClass)
		assert.Equal(t, "ERROR", nodes[2].Status.ErrorInfo.ErrorCode)
	})
}

func TestOVHCloudNodeGroup_TemplateNodeInfo(t *testing.T) {
	t.Run("template for b2-7 flavor", func(t *testing.T) {
		ng := newTestNodeGroup(t, "b2-7")
		template, err := ng.TemplateNodeInfo()
		assert.NoError(t, err)

		node := template.Node()
		assert.NotNil(t, node)

		assert.Contains(t, node.ObjectMeta.Name, fmt.Sprintf("%s-node-", ng.Id()))
		assert.Equal(t, map[string]string{"nodepool": ng.Id()}, node.Labels)
		assert.Equal(t, map[string]string(nil), node.Annotations)
		assert.Equal(t, []string(nil), node.Finalizers)
		assert.Equal(t, []v1.Taint(nil), node.Spec.Taints)

		assert.Equal(t, *resource.NewQuantity(110, resource.DecimalSI), node.Status.Capacity[apiv1.ResourcePods])
		assert.Equal(t, *resource.NewQuantity(2, resource.DecimalSI), node.Status.Capacity[apiv1.ResourceCPU])
		assert.Equal(t, *resource.NewQuantity(0, resource.DecimalSI), node.Status.Capacity[gpu.ResourceNvidiaGPU])
		assert.Equal(t, *resource.NewQuantity(7516192768, resource.DecimalSI), node.Status.Capacity[apiv1.ResourceMemory])
	})

	t.Run("template for t1-45 flavor", func(t *testing.T) {
		ng := newTestNodeGroup(t, "t1-45")
		template, err := ng.TemplateNodeInfo()
		assert.NoError(t, err)

		node := template.Node()
		assert.NotNil(t, node)

		assert.Contains(t, node.ObjectMeta.Name, fmt.Sprintf("%s-node-", ng.Id()))
		assert.Equal(t, map[string]string{"nodepool": ng.Id()}, node.Labels)
		assert.Equal(t, map[string]string(nil), node.Annotations)
		assert.Equal(t, []string(nil), node.Finalizers)
		assert.Equal(t, []v1.Taint(nil), node.Spec.Taints)

		assert.Equal(t, *resource.NewQuantity(110, resource.DecimalSI), node.Status.Capacity[apiv1.ResourcePods])
		assert.Equal(t, *resource.NewQuantity(8, resource.DecimalSI), node.Status.Capacity[apiv1.ResourceCPU])
		assert.Equal(t, *resource.NewQuantity(1, resource.DecimalSI), node.Status.Capacity[gpu.ResourceNvidiaGPU])
		assert.Equal(t, *resource.NewQuantity(48318382080, resource.DecimalSI), node.Status.Capacity[apiv1.ResourceMemory])
	})

	t.Run("template for b2-7 flavor with node pool templates", func(t *testing.T) {
		ng := newTestNodeGroup(t, "t1-45")

		// Setup node pool templates
		ng.Template.Metadata.Labels = map[string]string{
			"label1": "labelValue1",
		}
		ng.Template.Metadata.Annotations = map[string]string{
			"annotation1": "annotationValue1",
		}
		ng.Template.Metadata.Finalizers = []string{
			"finalizer1",
		}
		ng.Template.Spec.Taints = []v1.Taint{
			{
				Key:    "taintKey1",
				Value:  "taintValue1",
				Effect: "taintEffect1",
			},
		}

		template, err := ng.TemplateNodeInfo()
		assert.NoError(t, err)

		node := template.Node()
		assert.NotNil(t, node)

		assert.Contains(t, node.ObjectMeta.Name, fmt.Sprintf("%s-node-", ng.Id()))
		assert.Equal(t, map[string]string{"nodepool": ng.Id(), "label1": "labelValue1"}, node.Labels)
		assert.Equal(t, ng.Template.Metadata.Annotations, node.Annotations)
		assert.Equal(t, ng.Template.Metadata.Finalizers, node.Finalizers)
		assert.Equal(t, ng.Template.Spec.Taints, node.Spec.Taints)

		assert.Equal(t, *resource.NewQuantity(110, resource.DecimalSI), node.Status.Capacity[apiv1.ResourcePods])
		assert.Equal(t, *resource.NewQuantity(8, resource.DecimalSI), node.Status.Capacity[apiv1.ResourceCPU])
		assert.Equal(t, *resource.NewQuantity(1, resource.DecimalSI), node.Status.Capacity[gpu.ResourceNvidiaGPU])
		assert.Equal(t, *resource.NewQuantity(48318382080, resource.DecimalSI), node.Status.Capacity[apiv1.ResourceMemory])
	})
}

func TestOVHCloudNodeGroup_Exist(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check exist is true", func(t *testing.T) {
		exist := ng.Exist()

		assert.True(t, exist)
	})
}

func TestOVHCloudNodeGroup_Create(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check create node group", func(t *testing.T) {
		ng.mockCallCreateNodePool()

		newGroup, err := ng.Create()
		assert.NoError(t, err)

		targetSize, err := newGroup.TargetSize()
		assert.NoError(t, err)

		assert.Equal(t, "pool-b2-7", newGroup.Id())
		assert.Equal(t, 3, targetSize)
		assert.Equal(t, 1, newGroup.MinSize())
		assert.Equal(t, 5, newGroup.MaxSize())
	})
}

func TestOVHCloudNodeGroup_Delete(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check delete node group", func(t *testing.T) {
		ng.mockCallDeleteNodePool()

		err := ng.Delete()
		assert.NoError(t, err)
	})
}

func TestOVHCloudNodeGroup_Autoprovisioned(t *testing.T) {
	ng := newTestNodeGroup(t, "b2-7")

	t.Run("check auto-provisioned is false", func(t *testing.T) {
		provisioned := ng.Autoprovisioned()
		assert.False(t, provisioned)
	})
}

func TestOVHCloudNodeGroup_IsGpu(t *testing.T) {
	t.Run("has GPU cores", func(t *testing.T) {
		ng := newTestNodeGroup(t, "custom-t1-45")
		ng.Manager.FlavorsCache = map[string]sdk.Flavor{
			"custom-t1-45": {
				GPUs: 1,
			},
		}
		ng.Manager.FlavorsCacheExpirationTime = time.Now().Add(time.Minute)
		assert.True(t, ng.isGpu())
	})

	t.Run("doesn't have GPU cores", func(t *testing.T) {
		ng := newTestNodeGroup(t, "custom-b2-7")
		ng.Manager.FlavorsCache = map[string]sdk.Flavor{
			"custom-b2-7": {
				GPUs: 0,
			},
		}
		assert.False(t, ng.isGpu())
	})

	t.Run("not found but belong to GPU category", func(t *testing.T) {
		ng := newTestNodeGroup(t, GPUMachineCategory+"1-123")
		assert.True(t, ng.isGpu())
	})
}

func TestOVHCloudNodeGroup_GetOptions(t *testing.T) {
	t.Run("check get autoscaling options", func(t *testing.T) {
		ng := newTestNodeGroup(t, "b2-7")
		opts, err := ng.GetOptions(config.NodeGroupAutoscalingOptions{})

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%.1f", 3.2), fmt.Sprintf("%.1f", opts.ScaleDownUtilizationThreshold))
		assert.Equal(t, float64(0), opts.ScaleDownGpuUtilizationThreshold)
		assert.Equal(t, 10*time.Second, opts.ScaleDownUnneededTime)
		assert.Equal(t, 20*time.Second, opts.ScaleDownUnreadyTime)
	})

	t.Run("check get autoscaling options on gpu machine", func(t *testing.T) {
		ng := newTestNodeGroup(t, "t1-45")
		opts, err := ng.GetOptions(config.NodeGroupAutoscalingOptions{})

		assert.NoError(t, err)
		assert.Equal(t, float64(0), opts.ScaleDownUtilizationThreshold)
		assert.Equal(t, fmt.Sprintf("%.1f", 3.2), fmt.Sprintf("%.1f", opts.ScaleDownGpuUtilizationThreshold))
		assert.Equal(t, 10*time.Second, opts.ScaleDownUnneededTime)
		assert.Equal(t, 20*time.Second, opts.ScaleDownUnreadyTime)
	})
}
