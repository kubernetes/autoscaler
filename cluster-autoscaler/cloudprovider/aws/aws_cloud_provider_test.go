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
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"testing"
)

var testAwsManager = &AwsManager{
	asgCache: &asgCache{
		registeredAsgs: make(map[AwsRef]*asg, 0),
		asgToInstances: make(map[AwsRef][]AwsInstanceRef),
		instanceToAsg:  make(map[AwsInstanceRef]*asg),
		interrupt:      make(chan struct{}),
		awsService:     &testAwsService,
	},
	awsService: testAwsService,
}

func newTestAwsManagerWithMockServices(mockAutoScaling autoScalingI, mockEC2 ec2I, mockEKS eksI, autoDiscoverySpecs []asgAutoDiscoveryConfig, instanceStatus map[AwsInstanceRef]*string) *AwsManager {
	awsService := awsWrapper{mockAutoScaling, mockEC2, mockEKS}
	mgr := &AwsManager{
		awsService: awsService,
		asgCache: &asgCache{
			registeredAsgs:        make(map[AwsRef]*asg, 0),
			asgToInstances:        make(map[AwsRef][]AwsInstanceRef),
			instanceToAsg:         make(map[AwsInstanceRef]*asg),
			asgInstanceTypeCache:  newAsgInstanceTypeCache(&awsService),
			explicitlyConfigured:  make(map[AwsRef]bool),
			interrupt:             make(chan struct{}),
			asgAutoDiscoverySpecs: autoDiscoverySpecs,
			awsService:            &awsService,
			autoscalingOptions:    make(map[AwsRef]map[string]string),
		},
	}

	if instanceStatus != nil {
		mgr.asgCache.instanceStatus = instanceStatus
	}
	return mgr
}

func newTestAwsManagerWithAsgs(t *testing.T, mockAutoScaling autoScalingI, mockEC2 ec2I, specs []string) *AwsManager {
	m := newTestAwsManagerWithMockServices(mockAutoScaling, mockEC2, nil, nil, nil)
	m.asgCache.parseExplicitAsgs(specs)
	return m
}

func newTestAwsManagerWithAutoAsgs(t *testing.T, mockAutoScaling autoScalingI, mockEC2 ec2I, specs []string, autoDiscoverySpecs []asgAutoDiscoveryConfig) *AwsManager {
	m := newTestAwsManagerWithMockServices(mockAutoScaling, mockEC2, nil, autoDiscoverySpecs, nil)
	m.asgCache.parseExplicitAsgs(specs)
	return m
}

func testNamedDescribeAutoScalingGroupsOutput(groupName string, desiredCap int64, instanceIds ...string) *autoscaling.DescribeAutoScalingGroupsOutput {
	instances := []*autoscaling.Instance{}
	for _, id := range instanceIds {
		instances = append(instances, &autoscaling.Instance{
			InstanceId:       aws.String(id),
			AvailabilityZone: aws.String("us-east-1a"),
			LifecycleState:   aws.String(autoscaling.LifecycleStateInService),
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

func testSetASGInstanceLifecycle(asg *autoscaling.DescribeAutoScalingGroupsOutput, lifecycleState string) *autoscaling.DescribeAutoScalingGroupsOutput {
	for _, asg := range asg.AutoScalingGroups {
		for _, instance := range asg.Instances {
			instance.LifecycleState = aws.String(lifecycleState)
		}
	}
	return asg
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

func TestInstanceTypeFallback(t *testing.T) {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	do := cloudprovider.NodeGroupDiscoveryOptions{}
	opts := config.AutoscalingOptions{}

	t.Setenv("AWS_REGION", "non-existent-region")

	// This test ensures that no klog.Fatalf calls occur when constructing the AWS cloud provider.  Specifically it is
	// intended to ensure that instance type fallback works correctly in the event of an error enumerating instance
	// types.
	_ = BuildAWS(opts, do, resourceLimiter)
}

func TestName(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	assert.Equal(t, provider.Name(), cloudprovider.AwsProviderName)
}

func TestNodeGroups(t *testing.T) {
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, testAwsService, nil, []string{"1:5:test-asg"}))

	nodeGroups := provider.NodeGroups()
	assert.Equal(t, len(nodeGroups), 1)
	assert.Equal(t, nodeGroups[0].Id(), "test-asg")
	assert.Equal(t, nodeGroups[0].MinSize(), 1)
	assert.Equal(t, nodeGroups[0].MaxSize(), 5)
}

func TestAutoDiscoveredNodeGroups(t *testing.T) {
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAutoAsgs(t, a, nil, []string{}, []asgAutoDiscoveryConfig{
		{
			Tags: map[string]string{"test": ""},
		},
	}))

	a.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			Filters: []*autoscaling.Filter{
				{Name: aws.String("tag-key"), Values: aws.StringSlice([]string{"test"})},
			},
			MaxRecords: aws.Int64(maxRecordsReturnedByAPI),
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
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))

	a.On("DescribeAutoScalingGroupsPages",
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
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

	// test node in cluster that is not in a group managed by cluster autoscaler
	nodeNotInGroup := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id-not-in-group",
		},
	}

	group, err = provider.NodeGroupForNode(nodeNotInGroup)

	assert.NoError(t, err)
	assert.Nil(t, group)
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)
}

func TestNodeGroupForNodeWithNoProviderId(t *testing.T) {
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "",
		},
	}
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))
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
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	a.On("DescribeAutoScalingGroupsPages",
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

	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)
}

func TestIncreaseSize(t *testing.T) {
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	a.On("SetDesiredCapacity", &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asgs[0].Id()),
		DesiredCapacity:      aws.Int64(3),
		HonorCooldown:        aws.Bool(false),
	}).Return(&autoscaling.SetDesiredCapacityOutput{})

	a.On("DescribeAutoScalingGroupsPages",
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
	a.AssertNumberOfCalls(t, "SetDesiredCapacity", 1)
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

	newSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, newSize)
}

func TestBelongs(t *testing.T) {
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	a.On("DescribeAutoScalingGroupsPages",
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
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

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
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)
}

func TestDeleteNodes(t *testing.T) {
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	a.On("TerminateInstanceInAutoScalingGroup", &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String("test-instance-id"),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}).Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{
		Activity: &autoscaling.Activity{Description: aws.String("Deleted instance")},
	})

	// Look up the current number of instances...
	var expectedInstancesCount int64 = 2
	a.On("DescribeAutoScalingGroupsPages",
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
	a.AssertNumberOfCalls(t, "TerminateInstanceInAutoScalingGroup", 1)
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

	newSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, newSize)
}

func TestDeleteNodesTerminatingInstances(t *testing.T) {
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	a.On("TerminateInstanceInAutoScalingGroup", &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String("test-instance-id"),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}).Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{
		Activity: &autoscaling.Activity{Description: aws.String("Deleted instance")},
	})

	// Look up the current number of instances...
	var expectedInstancesCount int64 = 2
	a.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testSetASGInstanceLifecycle(testNamedDescribeAutoScalingGroupsOutput("test-asg", expectedInstancesCount, "test-instance-id", "second-test-instance-id"), autoscaling.LifecycleStateTerminatingWait), false)
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
	a.AssertNumberOfCalls(t, "TerminateInstanceInAutoScalingGroup", 0) // instances which are terminating don't need to be terminated again
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

	newSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 2, newSize)
}

func TestDeleteNodesTerminatedInstances(t *testing.T) {
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	a.On("TerminateInstanceInAutoScalingGroup", &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String("test-instance-id"),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}).Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{
		Activity: &autoscaling.Activity{Description: aws.String("Deleted instance")},
	})

	expectedInstancesCount := 2
	a.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testSetASGInstanceLifecycle(testNamedDescribeAutoScalingGroupsOutput("test-asg", int64(expectedInstancesCount), "test-instance-id", "second-test-instance-id"), autoscaling.LifecycleStateTerminated), false)
	}).Return(nil)

	// load ASG state into cache
	provider.Refresh()

	initialSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, expectedInstancesCount, initialSize)

	// try deleting a node, but all of them are already in a
	// Terminated state, so we should see no calls to Terminate.
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	err = asgs[0].DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, err)
	// we expect no calls to TerminateInstanceInAutoScalingGroup,
	// because the Node we tried to Delete was already terminating.
	a.AssertNumberOfCalls(t, "TerminateInstanceInAutoScalingGroup", 0)
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 1)

	newSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	// we expect TargetSize to stay the same, even though there are
	// two instances in Terminated state - TargetSize was already
	// adjusted for them in a previous loop.
	assert.Equal(t, initialSize, newSize)
}

func TestDeleteNodesWithPlaceholder(t *testing.T) {
	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"}))
	asgs := provider.NodeGroups()

	a.On("SetDesiredCapacity", &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asgs[0].Id()),
		DesiredCapacity:      aws.Int64(1),
		HonorCooldown:        aws.Bool(false),
	}).Return(&autoscaling.SetDesiredCapacityOutput{})

	// Look up the current number of instances...
	var expectedInstancesCount int64 = 2
	a.On("DescribeAutoScalingGroupsPages",
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

	a.On("DescribeScalingActivities",
		&autoscaling.DescribeScalingActivitiesInput{
			AutoScalingGroupName: aws.String("test-asg"),
		},
	).Return(&autoscaling.DescribeScalingActivitiesOutput{}, nil)

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
	a.AssertNumberOfCalls(t, "SetDesiredCapacity", 1)
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 2)

	newSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, newSize)
}

func TestDeleteNodesAfterMultipleRefreshes(t *testing.T) {
	a := &autoScalingMock{}
	manager := newTestAwsManagerWithAsgs(t, a, nil, []string{"1:5:test-asg"})
	provider := testProvider(t, manager)
	asgs := provider.NodeGroups()

	a.On("TerminateInstanceInAutoScalingGroup", &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String("test-instance-id"),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}).Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{
		Activity: &autoscaling.Activity{Description: aws.String("Deleted instance")},
	})

	// Look up the current number of instances...
	a.On("DescribeAutoScalingGroupsPages",
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
	mockAutoScaling := &autoScalingMock{}
	mockEC2 := &ec2Mock{}
	m := newTestAwsManagerWithMockServices(mockAutoScaling, mockEC2, nil, nil, nil)

	provider := testProvider(t, m)
	_, err := provider.GetResourceLimiter()
	assert.NoError(t, err)
}

func TestCleanup(t *testing.T) {
	provider := testProvider(t, testAwsManager)
	err := provider.Cleanup()
	assert.NoError(t, err)
}

func TestHasInstance(t *testing.T) {
	nodeStatus := "Healthy"
	mgr := &AwsManager{
		asgCache: &asgCache{
			registeredAsgs: make(map[AwsRef]*asg, 0),
			asgToInstances: make(map[AwsRef][]AwsInstanceRef),
			instanceToAsg:  make(map[AwsInstanceRef]*asg),
			interrupt:      make(chan struct{}),
			awsService:     &testAwsService,
			instanceStatus: map[AwsInstanceRef]*string{
				{
					ProviderID: "aws:///us-east-1a/test-instance-id",
					Name:       "test-instance-id",
				}: &nodeStatus,
			},
		},
		awsService: testAwsService,
	}
	provider := testProvider(t, mgr)

	// Case 1: correct node - present in AWS
	node1 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-1",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	present, err := provider.HasInstance(node1)
	assert.NoError(t, err)
	assert.True(t, present)

	// Case 2: incorrect node - fargate is unsupported
	node2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fargate-1",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id",
		},
	}
	present, err = provider.HasInstance(node2)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
	assert.True(t, present)

	// Case 3: correct node - not present in AWS
	node3 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-2",
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id-2",
		},
	}
	present, err = provider.HasInstance(node3)
	assert.ErrorContains(t, err, nodeNotPresentErr)
	assert.False(t, present)

	// Case 4: correct node - not autoscaled -> not present in AWS -> no warning
	node4 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-2",
			Annotations: map[string]string{
				"k8s.io/cluster-autoscaler-enabled": "false",
			},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "aws:///us-east-1a/test-instance-id-2",
		},
	}
	present, err = provider.HasInstance(node4)
	assert.NoError(t, err)
	assert.False(t, present)
}

func TestDeleteNodesWithPlaceholderAndStaleCache(t *testing.T) {
	// This test validates the scenario where ASG cache is not in sync with Autoscaling configuration.
	// we are taking an example where ASG size is 10, cache as 3 instances "i-0000", "i-0001" and "i-0002
	// But ASG has 6 instances i-0000 to i-10005. When DeleteInstances is called with 2 instances ("i-0000", "i-0001" )
	// and placeholders, CAS will terminate only these 2 instances after reducing ASG size by the count of placeholders

	a := &autoScalingMock{}
	provider := testProvider(t, newTestAwsManagerWithAsgs(t, a, nil, []string{"1:10:test-asg"}))
	asgs := provider.NodeGroups()
	commonAsg := &asg{
		AwsRef:  AwsRef{Name: asgs[0].Id()},
		minSize: asgs[0].MinSize(),
		maxSize: asgs[0].MaxSize(),
	}

	// desired capacity will be set as 6 as ASG has 4 placeholders
	a.On("SetDesiredCapacity", &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asgs[0].Id()),
		DesiredCapacity:      aws.Int64(6),
		HonorCooldown:        aws.Bool(false),
	}).Return(&autoscaling.SetDesiredCapacityOutput{})

	// Look up the current number of instances...
	var expectedInstancesCount int64 = 10
	a.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"test-asg"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("test-asg", expectedInstancesCount, "i-0000", "i-0001", "i-0002", "i-0003", "i-0004", "i-0005"), false)

		expectedInstancesCount = 4
	}).Return(nil)

	a.On("DescribeScalingActivities",
		&autoscaling.DescribeScalingActivitiesInput{
			AutoScalingGroupName: aws.String("test-asg"),
		},
	).Return(&autoscaling.DescribeScalingActivitiesOutput{}, nil)

	provider.Refresh()

	initialSize, err := asgs[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 10, initialSize)

	var awsInstanceRefs []AwsInstanceRef
	instanceToAsg := make(map[AwsInstanceRef]*asg)

	var nodes []*apiv1.Node
	for i := 3; i <= 9; i++ {
		providerId := fmt.Sprintf("aws:///us-east-1a/i-placeholder-test-asg-%d", i)
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: providerId,
			},
		}
		nodes = append(nodes, node)
		awsInstanceRef := AwsInstanceRef{
			ProviderID: providerId,
			Name:       fmt.Sprintf("i-placeholder-test-asg-%d", i),
		}
		awsInstanceRefs = append(awsInstanceRefs, awsInstanceRef)
		instanceToAsg[awsInstanceRef] = commonAsg
	}

	for i := 0; i <= 2; i++ {
		providerId := fmt.Sprintf("aws:///us-east-1a/i-000%d", i)
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: providerId,
			},
		}
		// only setting 2 instances to be terminated out of 3 active instances
		if i < 2 {
			nodes = append(nodes, node)
			a.On("TerminateInstanceInAutoScalingGroup", &autoscaling.TerminateInstanceInAutoScalingGroupInput{
				InstanceId:                     aws.String(fmt.Sprintf("i-000%d", i)),
				ShouldDecrementDesiredCapacity: aws.Bool(true),
			}).Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{
				Activity: &autoscaling.Activity{Description: aws.String("Deleted instance")},
			})
		}
		awsInstanceRef := AwsInstanceRef{
			ProviderID: providerId,
			Name:       fmt.Sprintf("i-000%d", i),
		}
		awsInstanceRefs = append(awsInstanceRefs, awsInstanceRef)
		instanceToAsg[awsInstanceRef] = commonAsg
	}

	// modifying provider to bring disparity between ASG and cache
	provider.awsManager.asgCache.asgToInstances[AwsRef{Name: "test-asg"}] = awsInstanceRefs
	provider.awsManager.asgCache.instanceToAsg = instanceToAsg

	// calling delete nodes 2 nodes and remaining placeholders
	err = asgs[0].DeleteNodes(nodes)
	assert.NoError(t, err)
	a.AssertNumberOfCalls(t, "SetDesiredCapacity", 1)
	a.AssertNumberOfCalls(t, "DescribeAutoScalingGroupsPages", 2)

	// This ensures only 2 instances are terminated which are mocked in this unit test
	a.AssertNumberOfCalls(t, "TerminateInstanceInAutoScalingGroup", 2)

}
