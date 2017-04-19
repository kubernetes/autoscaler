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

package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
)

type AutoScalingMock struct {
	mock.Mock
}

func (a *AutoScalingMock) DescribeAutoScalingGroups(i *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return &autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []*autoscaling.Group{
			{
				DesiredCapacity: aws.Int64(2),
				Instances: []*autoscaling.Instance{
					{
						InstanceId: aws.String("test-instance-id"),
					},
					{
						InstanceId: aws.String("second-test-instance-id"),
					},
				},
			},
		},
	}, nil
}

func (a *AutoScalingMock) DescribeTags(i *autoscaling.DescribeTagsInput) (*autoscaling.DescribeTagsOutput, error) {
	return &autoscaling.DescribeTagsOutput{
		Tags: []*autoscaling.TagDescription{
			{
				Key:               aws.String("foo"),
				Value:             aws.String("bar"),
				ResourceId:        aws.String("asg-123456"),
				ResourceType:      aws.String("auto-scaling-group"),
				PropagateAtLaunch: aws.Bool(false),
			},
		},
	}, nil
}

func (a *AutoScalingMock) SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	args := a.Called(input)
	return args.Get(0).(*autoscaling.SetDesiredCapacityOutput), nil
}

func (a *AutoScalingMock) TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	args := a.Called(input)
	return args.Get(0).(*autoscaling.TerminateInstanceInAutoScalingGroupOutput), nil
}

var testAwsManager = &AwsManager{
	asgs:     make([]*asgInformation, 0),
	service:  &AutoScalingMock{},
	asgCache: make(map[AwsRef]*Asg),
}

func testProvider(t *testing.T, m *AwsManager) *awsCloudProvider {
	provider, err := buildStaticallyDiscoveringProvider(m, nil)
	assert.NoError(t, err)
	return provider
}

func TestBuildAwsCloudProvider(t *testing.T) {
	m := testAwsManager
	_, err := buildStaticallyDiscoveringProvider(m, []string{"bad spec"})
	assert.Error(t, err)

	_, err = buildStaticallyDiscoveringProvider(m, nil)
	assert.NoError(t, err)
}

func TestAddNodeGroup(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	err := provider.addNodeGroup("bad spec")
	assert.Error(t, err)
	assert.Equal(t, len(provider.asgs), 0)

	err = provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.asgs), 1)
}

func TestName(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	assert.Equal(t, provider.Name(), "aws")
}

func TestNodeGroups(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	assert.Equal(t, len(provider.NodeGroups()), 0)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.NodeGroups()), 1)
}

func TestNodeGroupForNode(t *testing.T) {
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	provider := testProvider(t, testAwsManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	group, err := provider.NodeGroupForNode(node)

	assert.NoError(t, err)
	assert.Equal(t, group.Id(), "test-asg")
	assert.Equal(t, group.MinSize(), 1)
	assert.Equal(t, group.MaxSize(), 5)

	// test node in cluster that is not in a group managed by cluster autoscaler
	nodeNotInGroup := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id-not-in-group",
		},
	}

	group, err = provider.NodeGroupForNode(nodeNotInGroup)
	assert.NoError(t, err)
	assert.Nil(t, group)
}

func TestAwsRefFromProviderId(t *testing.T) {
	_, err := AwsRefFromProviderId("aws123")
	assert.Error(t, err)
	_, err = AwsRefFromProviderId("aws://test-az/test-instance-id")
	assert.Error(t, err)

	awsRef, err := AwsRefFromProviderId("aws:///us-east-1a/i-260942b3")
	assert.NoError(t, err)
	assert.Equal(t, awsRef, &AwsRef{Name: "i-260942b3"})
}

func TestMaxSize(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.asgs), 1)
	assert.Equal(t, provider.asgs[0].MaxSize(), 5)
}

func TestMinSize(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.asgs), 1)
	assert.Equal(t, provider.asgs[0].MinSize(), 1)
}

func TestTargetSize(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	targetSize, err := provider.asgs[0].TargetSize()
	assert.Equal(t, targetSize, 2)
	assert.NoError(t, err)
}

func TestIncreaseSize(t *testing.T) {
	service := &AutoScalingMock{}
	m := &AwsManager{
		asgs:     make([]*asgInformation, 0),
		service:  service,
		asgCache: make(map[AwsRef]*Asg),
	}
	provider := testProvider(t, m)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.asgs), 1)

	service.On("SetDesiredCapacity", &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(provider.asgs[0].Name),
		DesiredCapacity:      aws.Int64(3),
		HonorCooldown:        aws.Bool(false),
	}).Return(&autoscaling.SetDesiredCapacityOutput{})

	err = provider.asgs[0].IncreaseSize(1)
	assert.NoError(t, err)
	service.AssertNumberOfCalls(t, "SetDesiredCapacity", 1)
}

func TestBelongs(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)

	invalidNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/invalid-instance-id",
		},
	}
	_, err = provider.asgs[0].Belongs(invalidNode)
	assert.Error(t, err)

	validNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	belongs, err := provider.asgs[0].Belongs(validNode)
	assert.Equal(t, belongs, true)
	assert.NoError(t, err)
}

func TestDeleteNodes(t *testing.T) {
	service := &AutoScalingMock{}
	m := &AwsManager{
		asgs:     make([]*asgInformation, 0),
		service:  service,
		asgCache: make(map[AwsRef]*Asg),
	}

	service.On("TerminateInstanceInAutoScalingGroup", &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String("test-instance-id"),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}).Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{
		Activity: &autoscaling.Activity{Description: aws.String("Deleted instance")},
	})

	provider := testProvider(t, m)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	err = provider.asgs[0].DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
	service.AssertNumberOfCalls(t, "TerminateInstanceInAutoScalingGroup", 1)
}

func TestId(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	err := provider.addNodeGroup("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, len(provider.asgs), 1)
	assert.Equal(t, provider.asgs[0].Id(), "test-asg")
}

func TestDebug(t *testing.T) {
	asg := Asg{
		awsManager: testAwsManager,
		minSize:    5,
		maxSize:    55,
	}
	asg.Name = "test-asg"
	assert.Equal(t, asg.Debug(), "test-asg (5:55)")
}

func TestBuildAsg(t *testing.T) {
	_, err := buildAsgFromSpec("a", nil)
	assert.Error(t, err)
	_, err = buildAsgFromSpec("a:b:c", nil)
	assert.Error(t, err)
	_, err = buildAsgFromSpec("1:", nil)
	assert.Error(t, err)
	_, err = buildAsgFromSpec("1:2:", nil)
	assert.Error(t, err)

	asg, err := buildAsgFromSpec("111:222:test-name", nil)
	assert.NoError(t, err)
	assert.Equal(t, 111, asg.MinSize())
	assert.Equal(t, 222, asg.MaxSize())
	assert.Equal(t, "test-name", asg.Name)
}
