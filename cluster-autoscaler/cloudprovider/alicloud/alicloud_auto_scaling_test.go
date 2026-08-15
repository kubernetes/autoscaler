/*
Copyright 2018 The Kubernetes Authors.

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

package alicloud

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/services/ess"
)

var instancesOfPageOne = []ess.ScalingInstance{
	{InstanceId: "instance-1"},
	{InstanceId: "instance-2"},
	{InstanceId: "instance-3"},
	{InstanceId: "instance-4"},
	{InstanceId: "instance-5"},
	{InstanceId: "instance-6"},
	{InstanceId: "instance-7"},
	{InstanceId: "instance-8"},
	{InstanceId: "instance-9"},
	{InstanceId: "instance-10"},
}

var instancesOfPageTwo = []ess.ScalingInstance{
	{InstanceId: "instance-11"},
	{InstanceId: "instance-12"},
	{InstanceId: "instance-13"},
	{InstanceId: "instance-14"},
	{InstanceId: "instance-15"},
}

type mockAutoScaling struct {
	mock.Mock
}

func (as *mockAutoScaling) ScaleWithAdjustment(req *ess.ScaleWithAdjustmentRequest) (*ess.ScaleWithAdjustmentResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) DescribeScalingGroups(req *ess.DescribeScalingGroupsRequest) (*ess.DescribeScalingGroupsResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) DescribeScalingConfigurations(req *ess.DescribeScalingConfigurationsRequest) (*ess.DescribeScalingConfigurationsResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) DescribeScalingRules(req *ess.DescribeScalingRulesRequest) (*ess.DescribeScalingRulesResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) CreateScalingRule(req *ess.CreateScalingRuleRequest) (*ess.CreateScalingRuleResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) ModifyScalingGroup(req *ess.ModifyScalingGroupRequest) (*ess.ModifyScalingGroupResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) RemoveInstances(req *ess.RemoveInstancesRequest) (*ess.RemoveInstancesResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) ExecuteScalingRule(req *ess.ExecuteScalingRuleRequest) (*ess.ExecuteScalingRuleResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) ModifyScalingRule(req *ess.ModifyScalingRuleRequest) (*ess.ModifyScalingRuleResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) DeleteScalingRule(req *ess.DeleteScalingRuleRequest) (*ess.DeleteScalingRuleResponse, error) {
	return nil, nil
}

func (as *mockAutoScaling) DescribeScalingInstances(req *ess.DescribeScalingInstancesRequest) (*ess.DescribeScalingInstancesResponse, error) {
	instances := make([]ess.ScalingInstance, 0)

	pageNumber, err := req.PageNumber.GetValue()
	if err != nil {
		return nil, fmt.Errorf("invalid page number")
	}
	if pageNumber == 1 {
		instances = instancesOfPageOne
	} else if pageNumber == 2 {
		instances = instancesOfPageTwo
	} else {
		return nil, fmt.Errorf("exceed total num")
	}

	return &ess.DescribeScalingInstancesResponse{
		ScalingInstances: ess.ScalingInstances{ScalingInstance: instances},
		TotalCount:       len(instancesOfPageOne) + len(instancesOfPageTwo),
	}, nil
}

func newMockAutoScalingWrapper() *autoScalingWrapper {
	return &autoScalingWrapper{
		autoScaling: &mockAutoScaling{},
		cfg:         &cloudConfig{},
	}
}

func TestRRSACloudConfigEssClientCreation(t *testing.T) {
	t.Setenv(oidcProviderARN, "acs:ram::12345:oidc-provider/ack-rrsa-cb123")
	t.Setenv(oidcTokenFilePath, "/var/run/secrets/tokens/oidc-token")
	t.Setenv(roleARN, "acs:ram::12345:role/autoscaler-role")
	t.Setenv(roleSessionName, "session")
	t.Setenv(regionId, "cn-hangzhou")

	cfg := &cloudConfig{}
	assert.True(t, cfg.isValid())
	assert.True(t, cfg.RRSAEnabled)

	client, err := getEssClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetScalingInstancesByGroup(t *testing.T) {
	wrapper := newMockAutoScalingWrapper()
	instances, err := wrapper.getScalingInstancesByGroup("asg-123")
	assert.NoError(t, err)
	assert.Equal(t, len(instancesOfPageOne)+len(instancesOfPageTwo), len(instances))
}
