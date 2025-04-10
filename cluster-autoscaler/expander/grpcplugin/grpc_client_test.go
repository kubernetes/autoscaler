/*
Copyright 2021 The Kubernetes Authors.

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

package grpcplugin

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/expander/grpcplugin/protos"
	"k8s.io/autoscaler/cluster-autoscaler/expander/mocks"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/expander"

	_ "github.com/golang/mock/mockgen/model"
)

var (
	nodes = []*v1.Node{
		BuildTestNode("n1", 1000, 1000),
		BuildTestNode("n2", 1000, 1000),
		BuildTestNode("n3", 1000, 1000),
		BuildTestNode("n4", 1000, 1000),
	}

	eoT2Micro = expander.Option{
		Debug:     "t2.micro",
		NodeGroup: test.NewTestNodeGroup("my-asg.t2.micro", 10, 1, 1, true, false, "t2.micro", nil, nil),
	}
	eoT2Large = expander.Option{
		Debug:     "t2.large",
		NodeGroup: test.NewTestNodeGroup("my-asg.t2.large", 10, 1, 1, true, false, "t2.large", nil, nil),
	}
	eoT3Large = expander.Option{
		Debug:     "t3.large",
		NodeGroup: test.NewTestNodeGroup("my-asg.t3.large", 10, 1, 1, true, false, "t3.large", nil, nil),
	}
	eoM44XLarge = expander.Option{
		Debug:     "m4.4xlarge",
		NodeGroup: test.NewTestNodeGroup("my-asg.m4.4xlarge", 10, 1, 1, true, false, "m4.4xlarge", nil, nil),
	}
	eoT2MicroWithSimilar = expander.Option{
		Debug:             "t2.micro",
		NodeGroup:         test.NewTestNodeGroup("my-asg.t2.micro", 10, 1, 1, true, false, "t2.micro", nil, nil),
		SimilarNodeGroups: []cloudprovider.NodeGroup{test.NewTestNodeGroup("my-similar-asg.t2.micro", 10, 1, 1, true, false, "t2.micro", nil, nil)},
	}
	options = []expander.Option{eoT2Micro, eoT2Large, eoT3Large, eoM44XLarge}

	grpcEoT2Micro = protos.Option{
		NodeGroupId: eoT2Micro.NodeGroup.Id(),
		NodeCount:   int32(eoT2Micro.NodeCount),
		Debug:       eoT2Micro.Debug,
		Pod:         eoT2Micro.Pods,
	}
	grpcEoT2Large = protos.Option{
		NodeGroupId: eoT2Large.NodeGroup.Id(),
		NodeCount:   int32(eoT2Large.NodeCount),
		Debug:       eoT2Large.Debug,
		Pod:         eoT2Large.Pods,
	}
	grpcEoT3Large = protos.Option{
		NodeGroupId: eoT3Large.NodeGroup.Id(),
		NodeCount:   int32(eoT3Large.NodeCount),
		Debug:       eoT3Large.Debug,
		Pod:         eoT3Large.Pods,
	}
	grpcEoM44XLarge = protos.Option{
		NodeGroupId: eoM44XLarge.NodeGroup.Id(),
		NodeCount:   int32(eoM44XLarge.NodeCount),
		Debug:       eoM44XLarge.Debug,
		Pod:         eoM44XLarge.Pods,
	}
	grpcEoT2MicroWithSimilar = protos.Option{
		NodeGroupId:         eoT2Micro.NodeGroup.Id(),
		NodeCount:           int32(eoT2Micro.NodeCount),
		Debug:               eoT2Micro.Debug,
		Pod:                 eoT2Micro.Pods,
		SimilarNodeGroupIds: []string{eoT2MicroWithSimilar.SimilarNodeGroups[0].Id()},
	}
	grpcEoT2MicroWithSimilarWithExtraOptions = protos.Option{
		NodeGroupId:         eoT2Micro.NodeGroup.Id(),
		NodeCount:           int32(eoT2Micro.NodeCount),
		Debug:               eoT2Micro.Debug,
		Pod:                 eoT2Micro.Pods,
		SimilarNodeGroupIds: []string{eoT2MicroWithSimilar.SimilarNodeGroups[0].Id(), "extra-ng-id"},
	}
)

func TestPopulateOptionsForGrpc(t *testing.T) {
	testCases := []struct {
		desc         string
		opts         []expander.Option
		expectedOpts []*protos.Option
		expectedMap  map[string]expander.Option
	}{
		{
			desc:         "empty options",
			opts:         []expander.Option{},
			expectedOpts: []*protos.Option{},
			expectedMap:  map[string]expander.Option{},
		},
		{
			desc:         "one option",
			opts:         []expander.Option{eoT2Micro},
			expectedOpts: []*protos.Option{&grpcEoT2Micro},
			expectedMap:  map[string]expander.Option{eoT2Micro.NodeGroup.Id(): eoT2Micro},
		},
		{
			desc:         "many options",
			opts:         options,
			expectedOpts: []*protos.Option{&grpcEoT2Micro, &grpcEoT2Large, &grpcEoT3Large, &grpcEoM44XLarge},
			expectedMap: map[string]expander.Option{
				eoT2Micro.NodeGroup.Id():   eoT2Micro,
				eoT2Large.NodeGroup.Id():   eoT2Large,
				eoT3Large.NodeGroup.Id():   eoT3Large,
				eoM44XLarge.NodeGroup.Id(): eoM44XLarge,
			},
		},
		{
			desc:         "similar nodegroups are included",
			opts:         []expander.Option{eoT2MicroWithSimilar},
			expectedOpts: []*protos.Option{&grpcEoT2MicroWithSimilar},
			expectedMap:  map[string]expander.Option{eoT2MicroWithSimilar.NodeGroup.Id(): eoT2MicroWithSimilar},
		},
	}
	for _, tc := range testCases {
		grpcOptionsSlice, nodeGroupIDOptionMap := populateOptionsForGRPC(tc.opts)
		assert.Equal(t, tc.expectedOpts, grpcOptionsSlice)
		assert.Equal(t, tc.expectedMap, nodeGroupIDOptionMap)
	}
}

func makeFakeNodeInfos() map[string]*framework.NodeInfo {
	nodeInfos := make(map[string]*framework.NodeInfo)
	for i, opt := range options {
		nodeInfo := framework.NewTestNodeInfo(nodes[i])
		nodeInfos[opt.NodeGroup.Id()] = nodeInfo
	}
	return nodeInfos
}

func TestPopulateNodeInfoForGRPC(t *testing.T) {
	nodeInfos := makeFakeNodeInfos()
	grpcNodeInfoMap := populateNodeInfoForGRPC(nodeInfos)

	expectedGrpcNodeInfoMap := make(map[string]*v1.Node)
	for i, opt := range options {
		expectedGrpcNodeInfoMap[opt.NodeGroup.Id()] = nodes[i]
	}
	assert.Equal(t, expectedGrpcNodeInfoMap, grpcNodeInfoMap)
}

func TestValidTransformAndSanitizeOptionsFromGRPC(t *testing.T) {
	testCases := []struct {
		desc                  string
		responseOptions       []*protos.Option
		expectedOptions       []expander.Option
		nodegroupIDOptionaMap map[string]expander.Option
	}{
		{
			desc:            "valid transform and sanitize options",
			responseOptions: []*protos.Option{&grpcEoT2Micro, &grpcEoT3Large, &grpcEoM44XLarge},
			nodegroupIDOptionaMap: map[string]expander.Option{
				eoT2Micro.NodeGroup.Id():   eoT2Micro,
				eoT2Large.NodeGroup.Id():   eoT2Large,
				eoT3Large.NodeGroup.Id():   eoT3Large,
				eoM44XLarge.NodeGroup.Id(): eoM44XLarge,
			},
			expectedOptions: []expander.Option{eoT2Micro, eoT3Large, eoM44XLarge},
		},
		{
			desc:            "similar ngs are retained in proto options are retained",
			responseOptions: []*protos.Option{&grpcEoT2MicroWithSimilar},
			nodegroupIDOptionaMap: map[string]expander.Option{
				eoT2MicroWithSimilar.NodeGroup.Id(): eoT2MicroWithSimilar,
			},
			expectedOptions: []expander.Option{eoT2MicroWithSimilar},
		},
		{
			desc:            "extra similar ngs added to expander response are ignored",
			responseOptions: []*protos.Option{&grpcEoT2MicroWithSimilarWithExtraOptions},
			nodegroupIDOptionaMap: map[string]expander.Option{
				eoT2MicroWithSimilar.NodeGroup.Id(): eoT2MicroWithSimilar,
			},
			expectedOptions: []expander.Option{eoT2MicroWithSimilar},
		},
	}
	for _, tc := range testCases {
		ret := transformAndSanitizeOptionsFromGRPC(tc.responseOptions, tc.nodegroupIDOptionaMap)
		assert.Equal(t, tc.expectedOptions, ret)
	}
}

func TestAnInvalidTransformAndSanitizeOptionsFromGRPC(t *testing.T) {
	responseOptionsSlice := []*protos.Option{&grpcEoT2Micro, &grpcEoT3Large, &grpcEoM44XLarge}
	nodeGroupIDOptionMap := map[string]expander.Option{
		eoT2Micro.NodeGroup.Id(): eoT2Micro,
		eoT2Large.NodeGroup.Id(): eoT2Large,
		eoT3Large.NodeGroup.Id(): eoT3Large,
	}

	ret := transformAndSanitizeOptionsFromGRPC(responseOptionsSlice, nodeGroupIDOptionMap)
	assert.Equal(t, []expander.Option{eoT2Micro, eoT3Large}, ret)
}

func TestBestOptionsValid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockExpanderClient(ctrl)
	g := &grpcclientstrategy{mockClient}

	nodeInfos := makeFakeNodeInfos()
	grpcNodeInfoMap := make(map[string]*v1.Node)
	for i, opt := range options {
		grpcNodeInfoMap[opt.NodeGroup.Id()] = nodes[i]
	}
	expectedBestOptionsReq := &protos.BestOptionsRequest{
		Options: []*protos.Option{&grpcEoT2Micro, &grpcEoT2Large, &grpcEoT3Large, &grpcEoM44XLarge},
		NodeMap: grpcNodeInfoMap,
	}

	mockClient.EXPECT().BestOptions(
		gomock.Any(), gomock.Eq(expectedBestOptionsReq),
	).Return(&protos.BestOptionsResponse{Options: []*protos.Option{&grpcEoT3Large}}, nil)

	resp := g.BestOptions(options, nodeInfos)

	assert.Equal(t, resp, []expander.Option{eoT3Large})
}

func TestBestOptionsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockExpanderClient(ctrl)
	g := grpcclientstrategy{mockClient}

	testCases := []struct {
		desc         string
		mockResponse protos.BestOptionsResponse
	}{
		{
			desc:         "empty bestOptions response",
			mockResponse: protos.BestOptionsResponse{},
		},
		{
			desc:         "empty bestOptions response, options nil",
			mockResponse: protos.BestOptionsResponse{Options: nil},
		},
		{
			desc:         "empty bestOptions response, empty options slice",
			mockResponse: protos.BestOptionsResponse{Options: []*protos.Option{}},
		},
	}
	for _, tc := range testCases {
		grpcNodeInfoMap := populateNodeInfoForGRPC(makeFakeNodeInfos())
		mockClient.EXPECT().BestOptions(
			gomock.Any(), gomock.Eq(
				&protos.BestOptionsRequest{
					Options: []*protos.Option{&grpcEoT2Micro, &grpcEoT2Large, &grpcEoT3Large, &grpcEoM44XLarge},
					NodeMap: grpcNodeInfoMap,
				})).Return(&tc.mockResponse, nil)
		resp := g.BestOptions(options, makeFakeNodeInfos())

		assert.Nil(t, resp)
	}
}

// All test cases should error, and no options should be filtered
func TestBestOptionsErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockExpanderClient(ctrl)
	g := grpcclientstrategy{mockClient}

	badProtosOption := protos.Option{
		NodeGroupId: "badID",
		NodeCount:   int32(eoM44XLarge.NodeCount),
		Debug:       eoM44XLarge.Debug,
		Pod:         eoM44XLarge.Pods,
	}

	testCases := []struct {
		desc         string
		client       grpcclientstrategy
		nodeInfo     map[string]*framework.NodeInfo
		mockResponse protos.BestOptionsResponse
		errResponse  error
	}{
		{
			desc:         "Bad gRPC client config",
			client:       grpcclientstrategy{nil},
			nodeInfo:     makeFakeNodeInfos(),
			mockResponse: protos.BestOptionsResponse{},
			errResponse:  nil,
		},
		{
			desc:         "gRPC error response",
			client:       g,
			nodeInfo:     makeFakeNodeInfos(),
			mockResponse: protos.BestOptionsResponse{},
			errResponse:  errors.New("timeout error"),
		},
		{
			desc:         "bad bestOptions response, options invalid - nil",
			client:       g,
			nodeInfo:     makeFakeNodeInfos(),
			mockResponse: protos.BestOptionsResponse{Options: []*protos.Option{&grpcEoT2Micro, nil, &grpcEoT2Large, &grpcEoT3Large, &grpcEoM44XLarge}},
			errResponse:  nil,
		},
		{
			desc:         "bad bestOptions response, options invalid - nonExistent nodeID",
			client:       g,
			nodeInfo:     makeFakeNodeInfos(),
			mockResponse: protos.BestOptionsResponse{Options: []*protos.Option{&grpcEoT2Micro, &badProtosOption, &grpcEoT2Large, &grpcEoT3Large, &grpcEoM44XLarge}},
			errResponse:  nil,
		},
	}
	for _, tc := range testCases {
		grpcNodeInfoMap := populateNodeInfoForGRPC(tc.nodeInfo)
		if tc.client.grpcClient != nil {
			mockClient.EXPECT().BestOptions(
				gomock.Any(), gomock.Eq(
					&protos.BestOptionsRequest{
						Options: []*protos.Option{&grpcEoT2Micro, &grpcEoT2Large, &grpcEoT3Large, &grpcEoM44XLarge},
						NodeMap: grpcNodeInfoMap,
					})).Return(&tc.mockResponse, tc.errResponse)
		}
		resp := tc.client.BestOptions(options, tc.nodeInfo)

		assert.Equal(t, resp, options)
	}
}
