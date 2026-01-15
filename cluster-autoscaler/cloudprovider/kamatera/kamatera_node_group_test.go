/*
Copyright 2016 The Kubernetes Authors.

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

package kamatera

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNodeGroup_IncreaseSize(t *testing.T) {
	tests := []struct {
		name             string
		poweronOnScaleUp bool
	}{
		{
			name:             "default",
			poweronOnScaleUp: false,
		},
		{
			name:             "poweron on scale up",
			poweronOnScaleUp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := kamateraClientMock{}
			ctx := context.Background()
			mgr := manager{
				client:    &client,
				instances: make(map[string]*Instance),
				config: &kamateraConfig{
					providerIDPrefix: "rke2://",
					PoweronOnScaleUp: tt.poweronOnScaleUp,
				},
			}
			serverName1 := mockKamateraServerName()
			serverProviderID1 := formatKamateraProviderID("rke2://", serverName1)
			serverName2 := mockKamateraServerName()
			serverProviderID2 := formatKamateraProviderID("rke2://", serverName2)
			serverName3 := mockKamateraServerName()
			serverProviderID3 := formatKamateraProviderID("rke2://", serverName3)
			serverConfig := mockServerConfig("test", []string{})
			ng := NodeGroup{
				id:      "ng1",
				manager: &mgr,
				minSize: 1,
				maxSize: 7,
				instances: map[string]*Instance{
					serverProviderID1: {
						Id:      serverProviderID1,
						Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
						PowerOn: true,
					},
					serverProviderID2: {
						Id:      serverProviderID2,
						Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
						PowerOn: true,
					},
					serverProviderID3: {
						Id:      serverProviderID3,
						Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
						PowerOn: true,
					},
				},
				serverConfig: serverConfig,
			}

			// test error on bad delta values
			err := ng.IncreaseSize(0)
			assert.Error(t, err)
			assert.Equal(t, "delta must be positive, have: 0", err.Error())

			err = ng.IncreaseSize(-1)
			assert.Error(t, err)
			assert.Equal(t, "delta must be positive, have: -1", err.Error())

			// test error on a too large increase of nodes
			err = ng.IncreaseSize(5)
			assert.Error(t, err)
			assert.Equal(t, "size increase is too large. current: 3 desired: 8 max: 7", err.Error())

			// test ok to add a node
			client.On(
				"CreateServers", ctx, 1, serverConfig,
			).Return(
				[]Server{{Name: mockKamateraServerName(), PowerOn: true}}, nil,
			).Once().On(
				"ListServers", ctx, mgr.instances, "", "rke2://",
			).Return(
				[]Server{}, nil,
			).Once()
			err = ng.IncreaseSize(1)
			assert.NoError(t, err)
			assert.Equal(t, 4, len(ng.instances))

			// test ok to add multiple nodes
			client.On(
				"CreateServers", ctx, 2, serverConfig,
			).Return(
				[]Server{
					{Name: mockKamateraServerName(), PowerOn: true},
					{Name: mockKamateraServerName(), PowerOn: true},
				}, nil,
			).Once().On(
				"ListServers", ctx, mgr.instances, "", "rke2://",
			).Return(
				[]Server{}, nil,
			).Once()
			err = ng.IncreaseSize(2)
			assert.NoError(t, err)
			assert.Equal(t, 6, len(ng.instances))

			// test error on API call error
			client.On(
				"CreateServers", ctx, 1, serverConfig,
			).Return(
				[]Server{}, fmt.Errorf("error on API call"),
			).Once().On(
				"ListServers", ctx, mgr.instances, "", "rke2://",
			).Return(
				[]Server{}, nil,
			).Once()
			err = ng.IncreaseSize(1)
			assert.Error(t, err, "no error on injected API call error")
			assert.Equal(t, "error on API call", err.Error())
		})
	}
}

func TestNodeGroup_IncreaseSize_withPoweredOffServers(t *testing.T) {
	tests := []struct {
		name             string
		poweronOnScaleUp bool
	}{
		{
			name:             "default",
			poweronOnScaleUp: false,
		},
		{
			name:             "poweron on scale up",
			poweronOnScaleUp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := kamateraClientMock{}
			ctx := context.Background()
			PoweredOffServerName1 := mockKamateraServerName()
			PoweredOffServerProviderID1 := formatKamateraProviderID("rke2://", PoweredOffServerName1)
			PoweredOffServerName2 := mockKamateraServerName()
			PoweredOffServerProviderID2 := formatKamateraProviderID("rke2://", PoweredOffServerName2)
			PoweredOffServerName3 := mockKamateraServerName()
			PoweredOffServerProviderID3 := formatKamateraProviderID("rke2://", PoweredOffServerName3)
			mgr := manager{
				client: &client,
				instances: map[string]*Instance{
					PoweredOffServerProviderID1: {
						Id:      PoweredOffServerProviderID1,
						Tags:    []string{"tag1", "tag2"},
						PowerOn: false,
					},
					PoweredOffServerProviderID2: {
						Id:      PoweredOffServerProviderID2,
						Tags:    []string{"tag1", "tag3"},
						PowerOn: false,
					},
					PoweredOffServerProviderID3: {
						Id:      PoweredOffServerProviderID3,
						Tags:    []string{"tag3", "tag2"},
						PowerOn: false,
					},
				},
				config: &kamateraConfig{
					providerIDPrefix: "rke2://",
					PoweronOnScaleUp: tt.poweronOnScaleUp,
				},
			}
			serverName1 := mockKamateraServerName()
			serverProviderID1 := formatKamateraProviderID("rke2://", serverName1)
			serverName2 := mockKamateraServerName()
			serverProviderID2 := formatKamateraProviderID("rke2://", serverName2)
			serverName3 := mockKamateraServerName()
			serverProviderID3 := formatKamateraProviderID("rke2://", serverName3)
			serverConfig := mockServerConfig("test", []string{"tag1", "tag2"})
			ng := NodeGroup{
				id:      "ng1",
				manager: &mgr,
				minSize: 1,
				maxSize: 7,
				instances: map[string]*Instance{
					serverProviderID1: {
						Id:      serverProviderID1,
						Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
						PowerOn: true,
					},
					serverProviderID2: {
						Id:      serverProviderID2,
						Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
						PowerOn: true,
					},
					serverProviderID3: {
						Id:      serverProviderID3,
						Status:  &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
						PowerOn: true,
					},
				},
				serverConfig: serverConfig,
			}

			// test ok to add a node
			client.On(
				"ListServers", ctx, mgr.instances, "", "rke2://",
			).Return(
				[]Server{}, nil,
			).Once()
			if tt.poweronOnScaleUp {
				client.On(
					"PoweronServer", ctx, PoweredOffServerName1,
				).Return(
					nil,
				).Once()
			} else {
				client.On(
					"CreateServers", ctx, 1, serverConfig,
				).Return(
					[]Server{{Name: mockKamateraServerName(), PowerOn: true}}, nil,
				).Once()
			}
			err := ng.IncreaseSize(1)
			assert.NoError(t, err)
			assert.Equal(t, 4, len(ng.instances))

			// test ok to add multiple nodes
			mgr.instances[PoweredOffServerProviderID3].Tags = []string{"tag1", "tag2"}
			if tt.poweronOnScaleUp {
				client.On(
					"CreateServers", ctx, 1, serverConfig,
				).Return(
					[]Server{
						{Name: mockKamateraServerName(), PowerOn: true},
					}, nil,
				).Once().On(
					"ListServers", ctx, mgr.instances, "", "rke2://",
				).Return(
					[]Server{}, nil,
				).Once().On(
					"PoweronServer", ctx, PoweredOffServerName3,
				).Return(
					nil,
				).Once()
			} else {
				client.On(
					"CreateServers", ctx, 2, serverConfig,
				).Return(
					[]Server{
						{Name: mockKamateraServerName(), PowerOn: true},
						{Name: mockKamateraServerName(), PowerOn: true},
					}, nil,
				).Once().On(
					"ListServers", ctx, mgr.instances, "", "rke2://",
				).Return(
					[]Server{}, nil,
				).Once()
			}
			err = ng.IncreaseSize(2)
			assert.NoError(t, err)
			assert.Equal(t, 6, len(ng.instances))

			// test error on API call error
			client.On(
				"CreateServers", ctx, 1, serverConfig,
			).Return(
				[]Server{}, fmt.Errorf("error on API call"),
			).Once().On(
				"ListServers", ctx, mgr.instances, "", "rke2://",
			).Return(
				[]Server{}, nil,
			).Once()
			err = ng.IncreaseSize(1)
			assert.Error(t, err, "no error on injected API call error")
			assert.Equal(t, "error on API call", err.Error())
		})
	}
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	ng := &NodeGroup{}
	err := ng.DecreaseTargetSize(-1)
	assert.Error(t, err)
	assert.Equal(t, "Not implemented", err.Error())
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	tests := []struct {
		name                string
		poweroffOnScaleDown bool
	}{
		{
			name:                "default",
			poweroffOnScaleDown: false,
		},
		{
			name:                "poweroff on scale down",
			poweroffOnScaleDown: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := kamateraClientMock{}
			ctx := context.Background()
			mgr := manager{
				client:    &client,
				instances: make(map[string]*Instance),
				config: &kamateraConfig{
					providerIDPrefix:    "rke2://",
					PoweroffOnScaleDown: tt.poweroffOnScaleDown,
				},
			}
			serverName1 := mockKamateraServerName()
			serverProviderID1 := formatKamateraProviderID("rke2://", serverName1)
			serverName2 := mockKamateraServerName()
			serverProviderID2 := formatKamateraProviderID("rke2://", serverName2)
			serverName3 := mockKamateraServerName()
			serverProviderID3 := formatKamateraProviderID("rke2://", serverName3)
			serverName4 := mockKamateraServerName()
			serverProviderID4 := formatKamateraProviderID("rke2://", serverName4)
			serverName5 := mockKamateraServerName()
			serverProviderID5 := formatKamateraProviderID("rke2://", serverName5)
			serverName6 := mockKamateraServerName()
			serverProviderID6 := formatKamateraProviderID("rke2://", serverName6)
			ng := NodeGroup{
				id:      "ng1",
				minSize: 1,
				maxSize: 6,
				instances: map[string]*Instance{
					serverProviderID1: {Id: serverProviderID1, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
					serverProviderID2: {Id: serverProviderID2, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
					serverProviderID3: {Id: serverProviderID3, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
					serverProviderID4: {Id: serverProviderID4, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
					serverProviderID5: {Id: serverProviderID5, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
					serverProviderID6: {Id: serverProviderID6, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
				},
				manager: &mgr,
			}

			var expectedMethodName string
			if tt.poweroffOnScaleDown {
				expectedMethodName = "PoweroffServer"
			} else {
				expectedMethodName = "DeleteServer"
			}

			// test of deleting nodes
			client.On(
				expectedMethodName, ctx, serverName1,
			).Return(nil).Once().On(
				expectedMethodName, ctx, serverName2,
			).Return(nil).Once().On(
				expectedMethodName, ctx, serverName6,
			).Return(nil).Once()

			err := ng.DeleteNodes([]*apiv1.Node{
				{Spec: apiv1.NodeSpec{ProviderID: serverProviderID1}},
				{Spec: apiv1.NodeSpec{ProviderID: serverProviderID2}},
				{Spec: apiv1.NodeSpec{ProviderID: serverProviderID6}},
			})

			assert.NoError(t, err)
			if tt.poweroffOnScaleDown {
				assert.Equal(t, 6, len(ng.instances))
			} else {
				assert.Equal(t, 3, len(ng.instances))
			}
			targetSize, err := ng.TargetSize()
			assert.Equal(t, 3, targetSize)
			assert.Equal(t, serverProviderID3, ng.instances[serverProviderID3].Id)
			assert.Equal(t, serverProviderID4, ng.instances[serverProviderID4].Id)
			assert.Equal(t, serverProviderID5, ng.instances[serverProviderID5].Id)

			// test error on deleting a node we are not managing
			err = ng.DeleteNodes([]*apiv1.Node{{Spec: apiv1.NodeSpec{ProviderID: mockKamateraServerName()}}})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot find this node in the node group")

			// test error on deleting a node when the API call fails
			client.On(
				expectedMethodName, ctx, serverName4,
			).Return(fmt.Errorf("error on API call")).Once()
			err = ng.DeleteNodes([]*apiv1.Node{
				{Spec: apiv1.NodeSpec{ProviderID: serverProviderID4}},
			})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error on API call")
		})
	}
}

func TestNodeGroup_Nodes(t *testing.T) {
	client := kamateraClientMock{}
	mgr := manager{
		client:    &client,
		instances: make(map[string]*Instance),
		config:    &kamateraConfig{providerIDPrefix: "rke2://"},
	}
	providerIDPrefix := "rke2://"
	mgr.config = &kamateraConfig{providerIDPrefix: providerIDPrefix}
	serverName1 := mockKamateraServerName()
	serverProviderID1 := formatKamateraProviderID(providerIDPrefix, serverName1)
	serverName2 := mockKamateraServerName()
	serverProviderID2 := formatKamateraProviderID(providerIDPrefix, serverName2)
	serverName3 := mockKamateraServerName()
	serverProviderID3 := formatKamateraProviderID(providerIDPrefix, serverName3)
	ng := NodeGroup{
		id:      "ng1",
		minSize: 1,
		maxSize: 6,
		instances: map[string]*Instance{
			serverProviderID1: {Id: serverProviderID1, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
			serverProviderID2: {Id: serverProviderID2, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
			serverProviderID3: {Id: serverProviderID3, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		},
		manager: &mgr,
	}

	// test nodes returned from Nodes() are only the ones we are expecting
	// Instance.Id should be prefixed with the configured provider ID prefix to match node.Spec.ProviderID
	instancesList, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(instancesList))
	var serverIds []string
	for _, instance := range instancesList {
		serverIds = append(serverIds, instance.Id)
	}
	assert.Equal(t, 3, len(serverIds))
	assert.Contains(t, serverIds, formatKamateraProviderID(providerIDPrefix, serverName1))
	assert.Contains(t, serverIds, formatKamateraProviderID(providerIDPrefix, serverName2))
	assert.Contains(t, serverIds, formatKamateraProviderID(providerIDPrefix, serverName3))
}

func TestNodeGroup_getResourceList(t *testing.T) {
	ng := &NodeGroup{}
	_, err := ng.getResourceList()
	assert.ErrorContains(t, err, "failed to parse server config ram")
	ng.serverConfig.Ram = "1024mb"
	_, err = ng.getResourceList()
	assert.ErrorContains(t, err, "failed to parse server config ram")
	ng.serverConfig.Ram = "1024"
	_, err = ng.getResourceList()
	assert.ErrorContains(t, err, "failed to parse server config cpu")
	ng.serverConfig.Cpu = "55PX"
	_, err = ng.getResourceList()
	assert.ErrorContains(t, err, "failed to parse server config cpu")
	ng.serverConfig.Cpu = "55P"
	rl, err := ng.getResourceList()
	assert.NoError(t, err)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
		apiv1.ResourceCPU:     *resource.NewQuantity(int64(55), resource.DecimalSI),
		apiv1.ResourceMemory:  *resource.NewQuantity(int64(1024*1024*1024), resource.DecimalSI),
		apiv1.ResourceStorage: *resource.NewQuantity(int64(0*1024*1024*1024), resource.DecimalSI),
	}, rl)
	ng.serverConfig.Disks = []string{"size=oops"}
	_, err = ng.getResourceList()
	assert.ErrorContains(t, err, "invalid syntax")
	ng.serverConfig.Disks = []string{"size=50"}
	rl, err = ng.getResourceList()
	assert.NoError(t, err)
	assert.Equal(t, apiv1.ResourceList{
		apiv1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
		apiv1.ResourceCPU:     *resource.NewQuantity(int64(55), resource.DecimalSI),
		apiv1.ResourceMemory:  *resource.NewQuantity(int64(1024*1024*1024), resource.DecimalSI),
		apiv1.ResourceStorage: *resource.NewQuantity(int64(50*1024*1024*1024), resource.DecimalSI),
	}, rl)
}

func TestNodeGroup_TemplateNodeInfo(t *testing.T) {
	ng := &NodeGroup{
		serverConfig: ServerConfig{
			Ram:   "1024",
			Cpu:   "5D",
			Disks: []string{"size=50"},
		},
	}
	nodeInfo, err := ng.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, nodeInfo.Node().Status.Capacity, apiv1.ResourceList{
		apiv1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
		apiv1.ResourceCPU:     *resource.NewQuantity(int64(5), resource.DecimalSI),
		apiv1.ResourceMemory:  *resource.NewQuantity(int64(1024*1024*1024), resource.DecimalSI),
		apiv1.ResourceStorage: *resource.NewQuantity(int64(50*1024*1024*1024), resource.DecimalSI),
	})
	assert.Equal(t, map[string]string{}, nodeInfo.Node().Labels)

	// test with template labels
	ng.templateLabels = []string{"disktype=ssd", "kubernetes.io/os=linux"}
	nodeInfo, err = ng.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"disktype":         "ssd",
		"kubernetes.io/os": "linux",
	}, nodeInfo.Node().Labels)

	// test with invalid label format (missing =)
	ng.templateLabels = []string{"invalidlabel", "valid=label"}
	nodeInfo, err = ng.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"valid": "label",
	}, nodeInfo.Node().Labels)

	// test with label containing = in value
	ng.templateLabels = []string{"key=value=with=equals"}
	nodeInfo, err = ng.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"key": "value=with=equals",
	}, nodeInfo.Node().Labels)
}

func TestNodeGroup_Others(t *testing.T) {
	client := kamateraClientMock{}
	mgr := manager{
		client:    &client,
		instances: make(map[string]*Instance),
	}
	serverName1 := mockKamateraServerName()
	serverName2 := mockKamateraServerName()
	serverName3 := mockKamateraServerName()
	ng := NodeGroup{
		id:      "ng1",
		minSize: 1,
		maxSize: 7,
		instances: map[string]*Instance{
			serverName1: {Id: serverName1, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
			serverName2: {Id: serverName2, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
			serverName3: {Id: serverName3, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		},
		manager: &mgr,
	}
	assert.Equal(t, 1, ng.MinSize())
	assert.Equal(t, 7, ng.MaxSize())
	ts, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, ts)
	assert.Equal(t, "ng1", ng.Id())
	assert.Equal(t, "node group ID: ng1 (min:1 max:7)", ng.Debug())
	extendedDebug := strings.Split(ng.extendedDebug(), "\n")
	assert.Equal(t, 4, len(extendedDebug))
	assert.Contains(t, extendedDebug, "node group ID: ng1 (min:1 max:7)")
	for _, serverName := range []string{serverName1, serverName2, serverName3} {
		assert.Contains(t, extendedDebug, fmt.Sprintf("instance ID: %s state: Running powerOn: false", serverName))
	}
	assert.Equal(t, true, ng.Exist())
	assert.Equal(t, false, ng.Autoprovisioned())
	_, err = ng.Create()
	assert.Error(t, err)
	assert.Equal(t, "Not implemented", err.Error())
	err = ng.Delete()
	assert.Error(t, err)
	assert.Equal(t, "Not implemented", err.Error())
}

func TestNodeGroup_AtomicIncreaseSize(t *testing.T) {
	ng := &NodeGroup{}
	err := ng.AtomicIncreaseSize(1)
	assert.Error(t, err)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestNodeGroup_ForceDeleteNodes(t *testing.T) {
	ng := &NodeGroup{}
	err := ng.ForceDeleteNodes([]*apiv1.Node{})
	assert.Error(t, err)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestNodeGroup_GetOptions(t *testing.T) {
	ng := &NodeGroup{}
	opts, err := ng.GetOptions(config.NodeGroupAutoscalingOptions{})
	assert.Nil(t, opts)
	assert.Error(t, err)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestNodeGroup_findInstanceForNode_EmptyProviderID(t *testing.T) {
	serverName1 := mockKamateraServerName()
	serverProviderID1 := formatKamateraProviderID("", serverName1)
	serverName2 := mockKamateraServerName()
	serverProviderID2 := formatKamateraProviderID("", serverName2)

	// Create a fake kubernetes client
	fakeClient := fake.NewSimpleClientset()

	ng := NodeGroup{
		id: "ng1",
		instances: map[string]*Instance{
			serverProviderID1: {Id: serverProviderID1, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
			serverProviderID2: {Id: serverProviderID2, Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		},
		manager: &manager{kubeClient: fakeClient, config: &kamateraConfig{providerIDPrefix: defaultKamateraProviderIDPrefix}},
	}

	// Test finding an instance when ProviderID is empty but node name matches instance ID
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "",
		},
	}
	node.Name = serverName1

	instance, err := ng.findInstanceForNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, serverProviderID1, instance.Id)
	// Verify that ProviderID was set on the node object with kamatera:// prefix
	// (even though the kubernetes update may fail)
	assert.Equal(t, formatKamateraProviderID("", serverName1), node.Spec.ProviderID)

	// Test not finding when neither ProviderID nor name matches
	node2 := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "",
		},
	}
	node2.Name = mockKamateraServerName()

	instance2, err := ng.findInstanceForNode(node2)
	assert.NoError(t, err)
	assert.Nil(t, instance2)
}
