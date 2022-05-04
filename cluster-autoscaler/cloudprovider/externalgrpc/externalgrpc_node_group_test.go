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
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

func TestCloudProvider_Nodes(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()

	pbInstances := []*protos.Instance{
		{Id: "1", Status: &protos.InstanceStatus{
			InstanceState: protos.InstanceStatus_unspecified,
			ErrorInfo:     &protos.InstanceErrorInfo{},
		},
		},
		{Id: "2", Status: &protos.InstanceStatus{
			InstanceState: protos.InstanceStatus_instanceRunning,
			ErrorInfo:     &protos.InstanceErrorInfo{},
		},
		},
		{Id: "3", Status: &protos.InstanceStatus{
			InstanceState: protos.InstanceStatus_instanceRunning,
			ErrorInfo: &protos.InstanceErrorInfo{
				ErrorCode:          "error1",
				ErrorMessage:       "mock error",
				InstanceErrorClass: 1,
			},
		},
		},
	}

	// test correct call
	m.On(
		"NodeGroupNodes", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupNodesRequest) bool {
			return req.Id == "nodeGroup1"
		}),
	).Return(
		&protos.NodeGroupNodesResponse{
			Instances: pbInstances,
		}, nil,
	).Once()

	ng1 := NodeGroup{
		id:     "nodeGroup1",
		client: client,
	}

	instances, err := ng1.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(instances))
	for _, i := range instances {
		if i.Id == "1" {
			assert.Nil(t, i.Status)
		} else if i.Id == "2" {
			assert.Equal(t, cloudprovider.InstanceRunning, i.Status.State)
			assert.Nil(t, i.Status.ErrorInfo)
		} else if i.Id == "3" {
			assert.Equal(t, cloudprovider.InstanceRunning, i.Status.State)
			assert.Equal(t, cloudprovider.OutOfResourcesErrorClass, i.Status.ErrorInfo.ErrorClass)
			assert.Equal(t, "error1", i.Status.ErrorInfo.ErrorCode)
			assert.Equal(t, "mock error", i.Status.ErrorInfo.ErrorMessage)
		} else {
			assert.Fail(t, "instance not recognized")
		}
	}

	// test grpc error
	m.On(
		"NodeGroupNodes", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupNodesRequest) bool {
			return req.Id == "nodeGroup2"
		}),
	).Return(
		&protos.NodeGroupNodesResponse{
			Instances: pbInstances,
		},
		fmt.Errorf("mock error"),
	).Once()

	ng2 := NodeGroup{
		id:     "nodeGroup2",
		client: client,
	}

	_, err = ng2.Nodes()
	assert.Error(t, err)

}

func TestCloudProvider_TemplateNodeInfo(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()

	// test correct call
	apiv1Node1 := &apiv1.Node{}
	apiv1Node1.Name = "node1"

	apiv1Node2 := &apiv1.Node{}
	apiv1Node2.Name = "node2"

	m.On(
		"NodeGroupTemplateNodeInfo", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupTemplateNodeInfoRequest) bool {
			return req.Id == "nodeGroup1"
		}),
	).Return(
		&protos.NodeGroupTemplateNodeInfoResponse{
			NodeInfo: apiv1Node1,
		}, nil,
	).Once()

	m.On(
		"NodeGroupTemplateNodeInfo", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupTemplateNodeInfoRequest) bool {
			return req.Id == "nodeGroup2"
		}),
	).Return(
		&protos.NodeGroupTemplateNodeInfoResponse{
			NodeInfo: apiv1Node2,
		}, nil,
	).Once()

	ng1 := NodeGroup{
		id:     "nodeGroup1",
		client: client,
	}

	ng2 := NodeGroup{
		id:     "nodeGroup2",
		client: client,
	}

	nodeInfo1, err := ng1.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, apiv1Node1.Name, nodeInfo1.Node().Name)

	nodeInfo2, err := ng2.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, apiv1Node2.Name, nodeInfo2.Node().Name)

	// test cached answer
	nodeInfo1, err = ng1.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, apiv1Node1.Name, nodeInfo1.Node().Name)
	m.AssertNumberOfCalls(t, "NodeGroupTemplateNodeInfo", 2)

	// test nil TemplateNodeInfo
	m.On(
		"NodeGroupTemplateNodeInfo", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupTemplateNodeInfoRequest) bool {
			return req.Id == "nodeGroup3"
		}),
	).Return(
		&protos.NodeGroupTemplateNodeInfoResponse{
			NodeInfo: nil,
		}, nil,
	).Once()

	ng3 := NodeGroup{
		id:     "nodeGroup3",
		client: client,
	}

	nodeInfo3, err := ng3.TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Nil(t, nodeInfo3)

	// test grpc error
	m.On(
		"NodeGroupTemplateNodeInfo", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupTemplateNodeInfoRequest) bool {
			return req.Id == "nodeGroup4"
		}),
	).Return(
		&protos.NodeGroupTemplateNodeInfoResponse{
			NodeInfo: nil,
		},
		fmt.Errorf("mock error"),
	).Once()

	ng4 := NodeGroup{
		id:     "nodeGroup4",
		client: client,
	}

	_, err = ng4.TemplateNodeInfo()
	assert.Error(t, err)

}

func TestCloudProvider_GetOptions(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()

	// test correct call
	m.On(
		"NodeGroupGetOptions", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupAutoscalingOptionsRequest) bool {
			return req.Id == "nodeGroup1"
		}),
	).Return(
		&protos.NodeGroupAutoscalingOptionsResponse{
			NodeGroupAutoscalingOptions: &protos.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold:    0.6,
				ScaleDownGpuUtilizationThreshold: 0.7,
				ScaleDownUnneededTime:            &v1.Duration{Duration: time.Minute},
				ScaleDownUnreadyTime:             &v1.Duration{Duration: time.Hour},
			},
		},
		nil,
	)

	ng1 := NodeGroup{
		id:     "nodeGroup1",
		client: client,
	}
	defaultsOpts := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.6,
		ScaleDownGpuUtilizationThreshold: 0.7,
		ScaleDownUnneededTime:            time.Minute,
		ScaleDownUnreadyTime:             time.Hour,
	}

	opts, err := ng1.GetOptions(defaultsOpts)
	assert.NoError(t, err)
	assert.Equal(t, 0.6, opts.ScaleDownUtilizationThreshold)
	assert.Equal(t, 0.7, opts.ScaleDownGpuUtilizationThreshold)
	assert.Equal(t, time.Minute, opts.ScaleDownUnneededTime)
	assert.Equal(t, time.Hour, opts.ScaleDownUnreadyTime)

	// test grpc error
	m.On(
		"NodeGroupGetOptions", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupAutoscalingOptionsRequest) bool {
			return req.Id == "nodeGroup2"
		}),
	).Return(
		&protos.NodeGroupAutoscalingOptionsResponse{},
		fmt.Errorf("mock error"),
	)

	ng2 := NodeGroup{
		id:     "nodeGroup2",
		client: client,
	}

	_, err = ng2.GetOptions(defaultsOpts)
	assert.Error(t, err)

	// test no opts
	m.On(
		"NodeGroupGetOptions", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupAutoscalingOptionsRequest) bool {
			return req.Id == "nodeGroup3"
		}),
	).Return(
		&protos.NodeGroupAutoscalingOptionsResponse{},
		nil,
	)

	ng3 := NodeGroup{
		id:     "nodeGroup3",
		client: client,
	}

	opts, err = ng3.GetOptions(defaultsOpts)
	assert.NoError(t, err)
	assert.Nil(t, opts)
}

func TestCloudProvider_TargetSize(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()

	// test correct call
	m.On(
		"NodeGroupTargetSize", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupTargetSizeRequest) bool {
			return req.Id == "nodeGroup1"
		}),
	).Return(
		&protos.NodeGroupTargetSizeResponse{
			TargetSize: 1,
		}, nil,
	).Once()

	ng1 := NodeGroup{
		id:     "nodeGroup1",
		client: client,
	}

	size, err := ng1.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, size)

	// test grpc error
	m.On(
		"NodeGroupTargetSize", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupTargetSizeRequest) bool {
			return req.Id == "nodeGroup2"
		}),
	).Return(
		&protos.NodeGroupTargetSizeResponse{},
		fmt.Errorf("mock error"),
	).Once()

	ng2 := NodeGroup{
		id:     "nodeGroup2",
		client: client,
	}

	_, err = ng2.TargetSize()
	assert.Error(t, err)

}

func TestCloudProvider_IncreaseSize(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()

	// test correct call
	m.On(
		"NodeGroupIncreaseSize", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupIncreaseSizeRequest) bool {
			return req.Id == "nodeGroup1"
		}),
	).Return(
		&protos.NodeGroupIncreaseSizeResponse{}, nil,
	).Once()

	ng1 := NodeGroup{
		id:     "nodeGroup1",
		client: client,
	}

	err := ng1.IncreaseSize(1)
	assert.NoError(t, err)

	// test grpc error
	m.On(
		"NodeGroupIncreaseSize", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupIncreaseSizeRequest) bool {
			return req.Id == "nodeGroup2"
		}),
	).Return(
		&protos.NodeGroupIncreaseSizeResponse{},
		fmt.Errorf("mock error"),
	).Once()

	ng2 := NodeGroup{
		id:     "nodeGroup2",
		client: client,
	}

	err = ng2.IncreaseSize(1)
	assert.Error(t, err)

}

func TestCloudProvider_DecreaseSize(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()

	// test correct call
	m.On(
		"NodeGroupDecreaseTargetSize", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupDecreaseTargetSizeRequest) bool {
			return req.Id == "nodeGroup1"
		}),
	).Return(
		&protos.NodeGroupDecreaseTargetSizeResponse{}, nil,
	).Once()

	ng1 := NodeGroup{
		id:     "nodeGroup1",
		client: client,
	}

	err := ng1.DecreaseTargetSize(1)
	assert.NoError(t, err)

	// test grpc error
	m.On(
		"NodeGroupDecreaseTargetSize", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupDecreaseTargetSizeRequest) bool {
			return req.Id == "nodeGroup2"
		}),
	).Return(
		&protos.NodeGroupDecreaseTargetSizeResponse{},
		fmt.Errorf("mock error"),
	).Once()

	ng2 := NodeGroup{
		id:     "nodeGroup2",
		client: client,
	}

	err = ng2.DecreaseTargetSize(1)
	assert.Error(t, err)

}

func TestCloudProvider_DeleteNodes(t *testing.T) {
	client, m, teardown := setupTest(t)
	defer teardown()

	apiv1Node1 := &apiv1.Node{}
	apiv1Node1.Name = "node1"

	apiv1Node2 := &apiv1.Node{}
	apiv1Node2.Name = "node2"

	nodes := []*apiv1.Node{apiv1Node1, apiv1Node2}

	// test correct call
	m.On(
		"NodeGroupDeleteNodes", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupDeleteNodesRequest) bool {
			return req.Id == "nodeGroup1"
		}),
	).Return(
		&protos.NodeGroupDeleteNodesResponse{}, nil,
	).Once()

	ng1 := NodeGroup{
		id:     "nodeGroup1",
		client: client,
	}

	err := ng1.DeleteNodes(nodes)
	assert.NoError(t, err)

	// test grpc error
	m.On(
		"NodeGroupDeleteNodes", mock.Anything, mock.MatchedBy(func(req *protos.NodeGroupDeleteNodesRequest) bool {
			return req.Id == "nodeGroup2"
		}),
	).Return(
		&protos.NodeGroupDeleteNodesResponse{},
		fmt.Errorf("mock error"),
	).Once()

	ng2 := NodeGroup{
		id:     "nodeGroup2",
		client: client,
	}

	err = ng2.DeleteNodes(nodes)
	assert.Error(t, err)

}
