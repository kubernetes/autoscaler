/*
Copyright 2022 The Kubernetes Authors.

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

package externalgrpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/anypb"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
)

func TestCloudProvider_NodeGroups(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()
	c := newExternalGrpcCloudProvider(client, nil)

	m.On("Refresh", mock.Anything, mock.Anything).Return(&protos.RefreshResponse{}, nil)

	// test answer with multiple node groups
	m.On(
		"NodeGroups", mock.Anything, mock.Anything,
	).Return(
		&protos.NodeGroupsResponse{
			NodeGroups: []*protos.NodeGroup{
				{Id: "1", MinSize: 10, MaxSize: 20, Debug: "test1"},
				{Id: "2", MinSize: 30, MaxSize: 40, Debug: "test2"},
			},
		}, nil,
	).Times(2)

	ngs := c.NodeGroups()
	assert.Equal(t, 2, len(ngs))
	for _, ng := range ngs {
		if ng.Id() == "1" {
			assert.Equal(t, 10, ng.MinSize())
			assert.Equal(t, 20, ng.MaxSize())
			assert.Equal(t, "test1", ng.Debug())
		} else if ng.Id() == "2" {
			assert.Equal(t, 30, ng.MinSize())
			assert.Equal(t, 40, ng.MaxSize())
			assert.Equal(t, "test2", ng.Debug())
		} else {
			assert.Fail(t, "node group id not recognized")
		}
	}

	// test cached answer
	m.AssertNumberOfCalls(t, "NodeGroups", 1)
	ngs = c.NodeGroups()
	assert.Equal(t, 2, len(ngs))
	m.AssertNumberOfCalls(t, "NodeGroups", 1)

	// test answer after refresh to clear cached answer
	err := c.Refresh()
	assert.NoError(t, err)
	ngs = c.NodeGroups()
	assert.Equal(t, 2, len(ngs))
	m.AssertNumberOfCalls(t, "NodeGroups", 2)

	// test empty answer
	err = c.Refresh()
	assert.NoError(t, err)

	m.On(
		"NodeGroups", mock.Anything, mock.Anything,
	).Return(
		&protos.NodeGroupsResponse{
			NodeGroups: make([]*protos.NodeGroup, 0),
		}, nil,
	).Once()

	ngs = c.NodeGroups()
	assert.NotNil(t, ngs)
	assert.Equal(t, 0, len(ngs))

	// test grpc error
	err = c.Refresh()
	assert.NoError(t, err)

	m.On(
		"NodeGroups", mock.Anything, mock.Anything,
	).Return(
		&protos.NodeGroupsResponse{
			NodeGroups: make([]*protos.NodeGroup, 0),
		},
		fmt.Errorf("mock error"),
	).Once()

	ngs = c.NodeGroups()
	assert.NotNil(t, ngs)
	assert.Equal(t, 0, len(ngs))

}

func TestCloudProvider_NodeGroupForNode(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()
	c := newExternalGrpcCloudProvider(client, nil)

	m.On("Refresh", mock.Anything, mock.Anything).Return(&protos.RefreshResponse{}, nil)

	// test correct answer
	m.On(
		"NodeGroupForNode", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupForNodeRequest) bool {
			return req.Node.Name == "node1"
		}),
	).Return(
		&protos.NodeGroupForNodeResponse{
			NodeGroup: &protos.NodeGroup{Id: "1", MinSize: 10, MaxSize: 20, Debug: "test1"},
		}, nil,
	)
	m.On(
		"NodeGroupForNode", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupForNodeRequest) bool {
			return req.Node.Name == "node2"
		}),
	).Return(
		&protos.NodeGroupForNodeResponse{
			NodeGroup: &protos.NodeGroup{Id: "2", MinSize: 30, MaxSize: 40, Debug: "test2"},
		}, nil,
	)

	apiv1Node1 := &apiv1.Node{}
	apiv1Node1.Name = "node1"
	apiv1Node1.Spec.ProviderID = "providerId://node1"

	ng1, err := c.NodeGroupForNode(apiv1Node1)
	assert.NoError(t, err)
	assert.Equal(t, "1", ng1.Id())
	assert.Equal(t, 10, ng1.MinSize())
	assert.Equal(t, 20, ng1.MaxSize())
	assert.Equal(t, "test1", ng1.Debug())

	apiv1Node2 := &apiv1.Node{}
	apiv1Node2.Name = "node2"
	apiv1Node2.Spec.ProviderID = "providerId://node2"

	ng2, err := c.NodeGroupForNode(apiv1Node2)
	assert.NoError(t, err)
	assert.Equal(t, "2", ng2.Id())
	assert.Equal(t, 30, ng2.MinSize())
	assert.Equal(t, 40, ng2.MaxSize())
	assert.Equal(t, "test2", ng2.Debug())

	// test cached answer
	ng1, err = c.NodeGroupForNode(apiv1Node1)
	assert.NoError(t, err)
	assert.Equal(t, "1", ng1.Id())
	m.AssertNumberOfCalls(t, "NodeGroupForNode", 2)

	// test clear cache
	err = c.Refresh()
	assert.NoError(t, err)

	ng1, err = c.NodeGroupForNode(apiv1Node1)
	assert.NoError(t, err)
	assert.Equal(t, "1", ng1.Id())
	m.AssertNumberOfCalls(t, "NodeGroupForNode", 3)

	//test no node group for node
	err = c.Refresh()
	assert.NoError(t, err)

	m.On(
		"NodeGroupForNode", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupForNodeRequest) bool {
			return req.Node.Name == "node3"
		}),
	).Return(
		&protos.NodeGroupForNodeResponse{
			NodeGroup: &protos.NodeGroup{Id: ""},
		}, nil,
	)

	apiv1Node3 := &apiv1.Node{}
	apiv1Node3.Name = "node3"
	apiv1Node3.Spec.ProviderID = "providerId://node3"

	ng3, err := c.NodeGroupForNode(apiv1Node3)
	assert.NoError(t, err)
	assert.Equal(t, nil, ng3)

	//test grpc error
	err = c.Refresh()
	assert.NoError(t, err)

	m.On(
		"NodeGroupForNode", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupForNodeRequest) bool {
			return req.Node.Name == "node4"
		}),
	).Return(
		&protos.NodeGroupForNodeResponse{
			NodeGroup: &protos.NodeGroup{Id: ""},
		},
		fmt.Errorf("mock error"),
	)

	apiv1Node4 := &apiv1.Node{}
	apiv1Node4.Name = "node4"
	apiv1Node4.Spec.ProviderID = "providerId://node4"

	_, err = c.NodeGroupForNode(apiv1Node4)
	assert.Error(t, err)

	//test error is not cached
	_, err = c.NodeGroupForNode(apiv1Node4)
	assert.Error(t, err)
	m.AssertNumberOfCalls(t, "NodeGroupForNode", 6)

	//test nil node param
	_, err = c.NodeGroupForNode(nil)
	assert.Error(t, err)
}

func TestCloudProvider_Pricing(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()
	c := newExternalGrpcCloudProvider(client, nil)

	model, errPricing := c.Pricing()
	assert.NoError(t, errPricing)
	assert.NotNil(t, model)

	// test correct NodePrice call
	m.On(
		"PricingNodePrice", mock.Anything, mock.MatchedBy(func(req *protos.PricingNodePriceRequest) bool {
			return req.Node.Name == "node1"
		}),
	).Return(
		&protos.PricingNodePriceResponse{Price: 100},
		nil,
	)
	m.On(
		"PricingNodePrice", mock.Anything, mock.MatchedBy(func(req *protos.PricingNodePriceRequest) bool {
			return req.Node.Name == "node2"
		}),
	).Return(
		&protos.PricingNodePriceResponse{Price: 200},
		nil,
	)

	apiv1Node1 := &apiv1.Node{}
	apiv1Node1.Name = "node1"

	price, err := model.NodePrice(apiv1Node1, time.Time{}, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, float64(100), price)

	apiv1Node2 := &apiv1.Node{}
	apiv1Node2.Name = "node2"

	price, err = model.NodePrice(apiv1Node2, time.Time{}, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, float64(200), price)

	// test grpc error for NodePrice
	m.On(
		"PricingNodePrice", mock.Anything, mock.MatchedBy(func(req *protos.PricingNodePriceRequest) bool {
			return req.Node.Name == "node3"
		}),
	).Return(
		&protos.PricingNodePriceResponse{},
		fmt.Errorf("mock error"),
	)

	apiv1Node3 := &apiv1.Node{}
	apiv1Node3.Name = "node3"

	_, err = model.NodePrice(apiv1Node3, time.Time{}, time.Time{})
	assert.Error(t, err)

	// test correct PodPrice call
	m.On(
		"PricingPodPrice", mock.Anything, mock.MatchedBy(func(req *protos.PricingPodPriceRequest) bool {
			return req.Pod.Name == "pod1"
		}),
	).Return(
		&protos.PricingPodPriceResponse{Price: 100},
		nil,
	)
	m.On(
		"PricingPodPrice", mock.Anything, mock.MatchedBy(func(req *protos.PricingPodPriceRequest) bool {
			return req.Pod.Name == "pod2"
		}),
	).Return(
		&protos.PricingPodPriceResponse{Price: 200},
		nil,
	)

	apiv1Pod1 := &apiv1.Pod{}
	apiv1Pod1.Name = "pod1"

	price, err = model.PodPrice(apiv1Pod1, time.Time{}, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, float64(100), price)

	apiv1Pod2 := &apiv1.Pod{}
	apiv1Pod2.Name = "pod2"

	price, err = model.PodPrice(apiv1Pod2, time.Time{}, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, float64(200), price)

	// test grpc error for PodPrice
	m.On(
		"PricingPodPrice", mock.Anything, mock.MatchedBy(func(req *protos.PricingPodPriceRequest) bool {
			return req.Pod.Name == "pod3"
		}),
	).Return(
		&protos.PricingPodPriceResponse{},
		fmt.Errorf("mock error"),
	)

	apiv1Pod3 := &apiv1.Pod{}
	apiv1Pod3.Name = "pod3"

	_, err = model.PodPrice(apiv1Pod3, time.Time{}, time.Time{})
	assert.Error(t, err)
}

func TestCloudProvider_GPULabel(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()
	c := newExternalGrpcCloudProvider(client, nil)

	m.On("Refresh", mock.Anything, mock.Anything).Return(&protos.RefreshResponse{}, nil)

	// test correct call
	m.On(
		"GPULabel", mock.Anything, mock.Anything,
	).Return(
		&protos.GPULabelResponse{Label: "gpu_label"},
		nil,
	)

	label := c.GPULabel()
	assert.Equal(t, "gpu_label", label)

	// test cache
	label = c.GPULabel()
	assert.Equal(t, "gpu_label", label)
	m.AssertNumberOfCalls(t, "GPULabel", 1)

	// test grpc error
	client2, m2, teardown2 := setupTest(t)
	defer teardown2()
	c2 := newExternalGrpcCloudProvider(client2, nil)

	m2.On("Refresh", mock.Anything, mock.Anything).Return(&protos.RefreshResponse{}, nil)

	m2.On(
		"GPULabel", mock.Anything, mock.Anything,
	).Return(
		&protos.GPULabelResponse{Label: "gpu_label"},
		fmt.Errorf("mock error"),
	)
	label = c2.GPULabel()
	assert.Equal(t, "", label)

	//test error is not cached
	label = c2.GPULabel()
	assert.Equal(t, "", label)
	m2.AssertNumberOfCalls(t, "GPULabel", 2)

}

func TestCloudProvider_GetAvailableGPUTypes(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()
	c := newExternalGrpcCloudProvider(client, nil)

	m.On("Refresh", mock.Anything, mock.Anything).Return(&protos.RefreshResponse{}, nil)

	// test correct call
	pbGpuTypes := make(map[string]*anypb.Any)
	pbGpuTypes["type1"] = &anypb.Any{}
	pbGpuTypes["type2"] = &anypb.Any{}

	m.On(
		"GetAvailableGPUTypes", mock.Anything, mock.Anything,
	).Return(
		&protos.GetAvailableGPUTypesResponse{GpuTypes: pbGpuTypes},
		nil,
	)

	gpuTypes := c.GetAvailableGPUTypes()
	assert.NotNil(t, gpuTypes)
	assert.NotNil(t, gpuTypes["type1"])
	assert.NotNil(t, gpuTypes["type2"])

	// test cache
	gpuTypes = c.GetAvailableGPUTypes()
	assert.NotNil(t, gpuTypes)
	assert.NotNil(t, gpuTypes["type1"])
	assert.NotNil(t, gpuTypes["type2"])
	m.AssertNumberOfCalls(t, "GetAvailableGPUTypes", 1)

	// test no gpu types
	client2, m2, teardown2 := setupTest(t)
	defer teardown2()
	c2 := newExternalGrpcCloudProvider(client2, nil)

	m2.On(
		"GetAvailableGPUTypes", mock.Anything, mock.Anything,
	).Return(
		&protos.GetAvailableGPUTypesResponse{GpuTypes: nil},
		nil,
	)

	gpuTypes = c2.GetAvailableGPUTypes()
	assert.NotNil(t, gpuTypes)
	assert.Equal(t, 0, len(gpuTypes))

	// test grpc error
	client3, m3, teardown3 := setupTest(t)
	defer teardown3()
	c3 := newExternalGrpcCloudProvider(client3, nil)

	m3.On(
		"GetAvailableGPUTypes", mock.Anything, mock.Anything,
	).Return(
		&protos.GetAvailableGPUTypesResponse{GpuTypes: nil},
		fmt.Errorf("mock error"),
	)

	gpuTypes = c3.GetAvailableGPUTypes()
	assert.Nil(t, gpuTypes)

	// test error is not cahced
	gpuTypes = c3.GetAvailableGPUTypes()
	assert.Nil(t, gpuTypes)
	m3.AssertNumberOfCalls(t, "GetAvailableGPUTypes", 2)

}

func TestCloudProvider_Cleanup(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()
	c := newExternalGrpcCloudProvider(client, nil)

	// test correct call
	m.On(
		"Cleanup", mock.Anything, mock.Anything,
	).Return(
		&protos.CleanupResponse{},
		nil,
	).Once()

	err := c.Cleanup()
	assert.NoError(t, err)

	// test grpc error
	m.On(
		"Cleanup", mock.Anything, mock.Anything,
	).Return(
		&protos.CleanupResponse{},
		fmt.Errorf("mock error"),
	).Once()

	err = c.Cleanup()
	assert.Error(t, err)
}

func TestCloudProvider_Refresh(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()
	c := newExternalGrpcCloudProvider(client, nil)

	// test correct call
	m.On(
		"Refresh", mock.Anything, mock.Anything,
	).Return(
		&protos.RefreshResponse{},
		nil,
	).Once()

	err := c.Refresh()
	assert.NoError(t, err)

	// test grpc error
	m.On(
		"Refresh", mock.Anything, mock.Anything,
	).Return(
		&protos.RefreshResponse{},
		fmt.Errorf("mock error"),
	).Once()

	err = c.Refresh()
	assert.Error(t, err)
}
