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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type AutoScalingMock struct {
	mock.Mock
}

func (a *AutoScalingMock) DescribeAutoScalingGroupsPages(i *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error {
	args := a.Called(i, fn)
	return args.Error(0)
}

func (a *AutoScalingMock) DescribeLaunchConfigurations(i *autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	args := a.Called(i)
	return args.Get(0).(*autoscaling.DescribeLaunchConfigurationsOutput), nil
}

func (a *AutoScalingMock) DescribeTagsPages(i *autoscaling.DescribeTagsInput, fn func(*autoscaling.DescribeTagsOutput, bool) bool) error {
	args := a.Called(i, fn)
	return args.Error(0)
}

func (a *AutoScalingMock) SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	args := a.Called(input)
	return args.Get(0).(*autoscaling.SetDesiredCapacityOutput), nil
}

func (a *AutoScalingMock) TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	args := a.Called(input)
	return args.Get(0).(*autoscaling.TerminateInstanceInAutoScalingGroupOutput), nil
}

type EC2Mock struct {
	mock.Mock
}

func (e *EC2Mock) DescribeLaunchTemplateVersions(i *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	args := e.Called(i)
	return args.Get(0).(*ec2.DescribeLaunchTemplateVersionsOutput), nil
}

var testService = autoScalingWrapper{&AutoScalingMock{}, newLaunchConfigurationInstanceTypeCache()}

var testAwsManager = &AwsManager{
	asgCache: &asgCache{
		registeredAsgs: make([]*asg, 0),
		asgToInstances: make(map[AwsRef][]AwsInstanceRef),
		instanceToAsg:  make(map[AwsInstanceRef]*asg),
		interrupt:      make(chan struct{}),
		service:        testService,
	},
	autoScalingService: testService,
}

func newTestAwsManagerWithService(service autoScaling, autoDiscoverySpecs []asgAutoDiscoveryConfig) *AwsManager {
	wrapper := autoScalingWrapper{service, newLaunchConfigurationInstanceTypeCache()}
	return &AwsManager{
		autoScalingService: wrapper,
		asgCache: &asgCache{
			registeredAsgs:        make([]*asg, 0),
			asgToInstances:        make(map[AwsRef][]AwsInstanceRef),
			instanceToAsg:         make(map[AwsInstanceRef]*asg),
			explicitlyConfigured:  make(map[AwsRef]bool),
			interrupt:             make(chan struct{}),
			asgAutoDiscoverySpecs: autoDiscoverySpecs,
			service:               wrapper,
		},
	}
}

func newTestAwsManagerWithAsgs(t *testing.T, service autoScaling, specs []string) *AwsManager {
	m := newTestAwsManagerWithService(service, nil)
	m.asgCache.parseExplicitAsgs(specs)
	return m
}

func newTestAwsManagerWithAutoAsgs(t *testing.T, service autoScaling, specs []string, autoDiscoverySpecs []asgAutoDiscoveryConfig) *AwsManager {
	m := newTestAwsManagerWithService(service, autoDiscoverySpecs)
	m.asgCache.parseExplicitAsgs(specs)
	return m
}

func testNamedDescribeAutoScalingGroupsOutput(groupName string, desiredCap int64, instanceIds ...string) *autoscaling.DescribeAutoScalingGroupsOutput {
	instances := []*autoscaling.Instance{}
	for _, id := range instanceIds {
		instances = append(instances, &autoscaling.Instance{
			InstanceId:       aws.String(id),
			AvailabilityZone: aws.String("us-east-1a"),
		})
	}
	return &autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []*autoscaling.Group{
			{
				AutoScalingGroupName: aws.String(groupName),
				DesiredCapacity:      aws.Int64(desiredCap),
				MinSize:              aws.Int64(1),
				MaxSize:              aws.Int64(5),
				Instances:            instances,
				AvailabilityZones:    aws.StringSlice([]string{"us-east-1a"}),
			},
		},
	}
}

func testProvider(t *testing.T, m *AwsManager) *awsCloudProvider {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	provider, err := BuildAwsCloudProvider(m, resourceLimiter)
	assert.NoError(t, err)
	return provider.(*awsCloudProvider)
}

func TestBuildAwsCloudProvider(t *testing.T) {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	_, err := BuildAwsCloudProvider(testAwsManager, resourceLimiter)
	assert.NoError(t, err)
}

func TestName(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	assert.Equal(t, provider.Name(), cloudprovider.AwsProviderName)
}

func TestNodeGroups(t *testing.T) {
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, testService, []string{"1:5:test-asg"}))

	nodeGroups := provider.NodeGroups()
	assert.Equal(t, len(nodeGroups), 1)
	assert.Equal(t, nodeGroups[0].Id(), "test-asg")
	assert.Equal(t, nodeGroups[0].MinSize(), 1)
	assert.Equal(t, nodeGroups[0].MaxSize(), 5)
}

func TestAutoDiscoveredNodeGroups(t *testing.T) {
	service := &AutoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAutoAsgs(t, service, []string{}, []asgAutoDiscoveryConfig{
		{
			Tags: map[string]string{"test": ""},
		},
	}))

	service.On("DescribeTagsPages",
		&autoscaling.DescribeTagsInput{
			Filters: []*autoscaling.Filter{
				{Name: aws.String("key"), Values: aws.StringSlice([]string{"test"})},
			},
			MaxRecords: aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeTagsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeTagsOutput, bool) bool)
		fn(&autoscaling.DescribeTagsOutput{
			Tags: []*autoscaling.TagDescription{
				{ResourceId: aws.String("auto-asg")},
			}}, false)
	}).Return(nil).Once()

	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"auto-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("auto-asg", 1, "test-instance-id"), false)
	}).Return(nil)

	provider.Refresh()

	nodeGroups := provider.NodeGroups()
	assert.Equal(t, len(nodeGroups), 1)
	assert.Equal(t, nodeGroups[0].Id(), "auto-asg")
	assert.Equal(t, nodeGroups[0].MinSize(), 1)
	assert.Equal(t, nodeGroups[0].MaxSize(), 5)
}

func TestNodeGroupForNode(t *testing.T) {
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	service := &AutoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, service, []string{"1:5:test-asg"}))

	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("test-asg", 1, "test-instance-id"), false)
	}).Return(nil)

	provider.Refresh()

	group, err := provider.NodeGroupForNode(node)

	assert.NoError(t, err)
	assert.Equal(t, group.Id(), "test-asg")
	assert.Equal(t, group.MinSize(), 1)
	assert.Equal(t, group.MaxSize(), 5)

	nodes, err := group.Nodes()

	assert.NoError(t, err)

	assert.Equal(t, []cloudprovider.Instance{{Id: "aws:///us-east-1a/test-instance-id"}}, nodes)
	service.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

	// test node in cluster that is not in a group managed by cluster autoscaler
	nodeNotInGroup := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id-not-in-group",
		},
	}

	group, err = provider.NodeGroupForNode(nodeNotInGroup)

	assert.NoError(t, err)
	assert.Nil(t, group)
	service.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)
}

func TestNodeGroupForNodeWithNoProviderId(t *testing.T) {
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "",
		},
	}
	service := &AutoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, service, []string{"1:5:test-asg"}))
	group, err := provider.NodeGroupForNode(node)

	assert.NoError(t, err)
	assert.Equal(t, group, nil)
}

func TestAwsRefFromProviderId(t *testing.T) {
	tests := []struct {
		provID string
		expErr bool
		expRef *AwsInstanceRef
	}{
		{
			provID: "aws123",
			expErr: true,
		},
		{
			provID: "aws://test-az/test-instance-id",
			expErr: true,
		},
		{

			provID: "aws:///us-east-1a/i-260942b3",
			expErr: false,
			expRef: &AwsInstanceRef{
				Name:       "i-260942b3",
				ProviderID: "aws:///us-east-1a/i-260942b3",
			},
		},
		{
			provID: "aws:///us-east-1a/i-placeholder-some.arbitrary.cluster.local",
			expErr: false,
			expRef: &AwsInstanceRef{
				Name:       "i-placeholder-some.arbitrary.cluster.local",
				ProviderID: "aws:///us-east-1a/i-placeholder-some.arbitrary.cluster.local",
			},
		},
		{
			// ref: https://github.com/kubernetes/autoscaler/issues/2285
			provID: "aws:///eu-central-1c/i-placeholder-K3-EKS-spotr5xlasgsubnet02af43b02922e710f-10QH9H0C8PG7O-14",
			expErr: false,
			expRef: &AwsInstanceRef{
				Name:       "i-placeholder-K3-EKS-spotr5xlasgsubnet02af43b02922e710f-10QH9H0C8PG7O-14",
				ProviderID: "aws:///eu-central-1c/i-placeholder-K3-EKS-spotr5xlasgsubnet02af43b02922e710f-10QH9H0C8PG7O-14",
			},
		},
	}

	for _, test := range tests {
		got, err := AwsRefFromProviderId(test.provID)
		if test.expErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, got, test.expRef)
		}
	}
}

func TestTargetSize(t *testing.T) {
	service := &AutoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, service, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("test-asg", 2, "test-instance-id", "second-test-instance-id"), false)
	}).Return(nil)

	provider.Refresh()

	targetSize, err := asgs[0].TargetSize()
	assert.Equal(t, targetSize, 2)
	assert.NoError(t, err)

	service.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)
}

func TestIncreaseSize(t *testing.T) {
	service := &AutoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, service, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	service.On("SetDesiredCapacity", &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asgs[0].Id()),
		DesiredCapacity:      aws.Int64(3),
		HonorCooldown:        aws.Bool(false),
	}).Return(&autoscaling.SetDesiredCapacityOutput{})

	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("test-asg", 2, "test-instance-id", "second-test-instance-id"), false)
	}).Return(nil)

	provider.Refresh()

	initialSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, initialSize)

	err = asgs[0].IncreaseSize(1)
	assert.NoError(t, err)
	service.AssertNumberOfCalls(t, "SetDesiredCapacity", 1)
	service.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

	newSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, newSize)
}

func TestBelongs(t *testing.T) {
	service := &AutoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, service, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{asgs[0].Id()}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("test-asg", 1, "test-instance-id"), false)
	}).Return(nil)

	provider.Refresh()

	invalidNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/invalid-instance-id",
		},
	}
	_, err := asgs[0].(*AwsNodeGroup).Belongs(invalidNode)
	assert.Error(t, err)
	service.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

	validNode := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	belongs, err := asgs[0].(*AwsNodeGroup).Belongs(validNode)
	assert.Equal(t, belongs, true)
	assert.NoError(t, err)
	// As "test-instance-id" is already known to be managed by test-asg since
	// the first `Belongs` call, no additional DescribAutoScalingGroupsPages
	// call is made.
	service.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)
}

func TestDeleteNodes(t *testing.T) {
	service := &AutoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, service, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	service.On("TerminateInstanceInAutoScalingGroup", &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String("test-instance-id"),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}).Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{
		Activity: &autoscaling.Activity{Description: aws.String("Deleted instance")},
	})

	// Look up the current number of instances...
	var expectedInstancesCount int64 = 2
	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("test-asg", expectedInstancesCount, "test-instance-id", "second-test-instance-id"), false)
		// we expect the instance count to be 1 after the call to DeleteNodes
		expectedInstancesCount = 1
	}).Return(nil)

	provider.Refresh()

	initialSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, initialSize)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	err = asgs[0].DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
	service.AssertNumberOfCalls(t, "TerminateInstanceInAutoScalingGroup", 1)
	service.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 2)

	newSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, newSize)
}

func TestDeleteNodesWithPlaceholder(t *testing.T) {
	service := &AutoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, service, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	service.On("SetDesiredCapacity", &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asgs[0].Id()),
		DesiredCapacity:      aws.Int64(1),
		HonorCooldown:        aws.Bool(false),
	}).Return(&autoscaling.SetDesiredCapacityOutput{})

	// Look up the current number of instances...
	var expectedInstancesCount int64 = 2
	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("test-asg", expectedInstancesCount, "test-instance-id"), false)
		// we expect the instance count to be 1 after the call to DeleteNodes
		expectedInstancesCount = 1
	}).Return(nil)

	provider.Refresh()

	initialSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, initialSize)

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/i-placeholder-test-asg-1",
		},
	}
	err = asgs[0].DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
	service.AssertNumberOfCalls(t, "SetDesiredCapacity", 1)
	service.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 2)

	newSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, newSize)
}

func TestDeleteNodesAfterMultipleRefreshes(t *testing.T) {
	service := &AutoScalingMock{}
	manager := newTestAwsManagerWithAsgs(t, service, []string{"1:5:test-asg"})
	provider := testProvider(t, manager)
	asgs := provider.NodeGroups()

	service.On("TerminateInstanceInAutoScalingGroup", &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String("test-instance-id"),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}).Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{
		Activity: &autoscaling.Activity{Description: aws.String("Deleted instance")},
	})

	// Look up the current number of instances...
	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("test-asg", 2, "test-instance-id", "second-test-instance-id"), false)
	}).Return(nil)

	provider.Refresh()
	// Call the manager directly as otherwise the call would be a noop as its within less then 60s
	manager.forceRefresh()

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	err := asgs[0].DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
}

func TestGetResourceLimiter(t *testing.T) {
	service := &AutoScalingMock{}
	m := newTestAwsManagerWithService(service, nil)

	provider := testProvider(t, m)
	_, err := provider.GetResourceLimiter()
	assert.NoError(t, err)
}

func TestCleanup(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	err := provider.Cleanup()
	assert.NoError(t, err)
}
