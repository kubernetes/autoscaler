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

package aws

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildAsg(t *testing.T) {
	asgCache := &asgCache{}

	asg, err := asgCache.buildAsgFromSpec("1:5:test-asg")
	assert.NoError(t, err)
	assert.Equal(t, asg.minSize, 1)
	assert.Equal(t, asg.maxSize, 5)
	assert.Equal(t, asg.Name, "test-asg")

	_, err = asgCache.buildAsgFromSpec("a")
	assert.Error(t, err)
	_, err = asgCache.buildAsgFromSpec("a:b:c")
	assert.Error(t, err)
	_, err = asgCache.buildAsgFromSpec("1:")
	assert.Error(t, err)
	_, err = asgCache.buildAsgFromSpec("1:2:")
	assert.Error(t, err)
}

func validateAsg(t *testing.T, asg *asg, name string, minSize int, maxSize int) {
	assert.Equal(t, name, asg.Name)
	assert.Equal(t, minSize, asg.minSize)
	assert.Equal(t, maxSize, asg.maxSize)
}

func TestCreatePlaceholders(t *testing.T) {
	registeredAsgName := aws.String("test-asg")
	registeredAsgRef := AwsRef{Name: *registeredAsgName}

	cases := []struct {
		name                string
		desiredCapacity     *int32
		activities          []autoscalingtypes.Activity
		groupLastUpdateTime time.Time
		describeErr         error
		asgToCheck          *string
	}{
		{
			name:            "add placeholders successful",
			desiredCapacity: aws.Int32(10),
		},
		{
			name:            "no placeholders needed",
			desiredCapacity: aws.Int32(0),
		},
		{
			name:            "DescribeScalingActivities failed",
			desiredCapacity: aws.Int32(1),
			describeErr:     errors.New("timeout"),
		},
		{
			name:            "early abort if AWS scaling up fails",
			desiredCapacity: aws.Int32(1),
			activities: []autoscalingtypes.Activity{
				{
					StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed,
					StartTime:  aws.Time(time.Unix(10, 0)),
				},
			},
			groupLastUpdateTime: time.Unix(9, 0),
		},
		{
			name:            "AWS scaling failed event before CA scale_up",
			desiredCapacity: aws.Int32(1),
			activities: []autoscalingtypes.Activity{
				{
					StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed,
					StartTime:  aws.Time(time.Unix(9, 0)),
				},
			},
			groupLastUpdateTime: time.Unix(10, 0),
		},
		{
			name:            "asg not registered",
			desiredCapacity: aws.Int32(10),
			activities: []autoscalingtypes.Activity{
				{
					StatusCode: autoscalingtypes.ScalingActivityStatusCodeFailed,
					StartTime:  aws.Time(time.Unix(10, 0)),
				},
			},
			groupLastUpdateTime: time.Unix(9, 0),
			asgToCheck:          aws.String("unregisteredAsgName"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shouldCallDescribeScalingActivities := true
			if *tc.desiredCapacity == int32(0) {
				shouldCallDescribeScalingActivities = false
			}

			asgName := registeredAsgName
			if tc.asgToCheck != nil {
				asgName = tc.asgToCheck
			}

			a := &autoScalingMock{}
			if shouldCallDescribeScalingActivities {
				a.On("DescribeScalingActivities",
					mock.Anything,
					&autoscaling.DescribeScalingActivitiesInput{
						AutoScalingGroupName: asgName,
					},
				).Return(
					&autoscaling.DescribeScalingActivitiesOutput{Activities: tc.activities},
					tc.describeErr,
				).Once()
			}

			asgCache := &asgCache{
				awsService: &awsWrapper{
					autoScalingI: a,
					ec2I:         nil,
				},
				registeredAsgs: map[AwsRef]*asg{
					registeredAsgRef: {
						AwsRef:         registeredAsgRef,
						lastUpdateTime: tc.groupLastUpdateTime,
					},
				},
			}

			groups := []autoscalingtypes.AutoScalingGroup{
				{
					AutoScalingGroupName: asgName,
					AvailabilityZones:    []string{"westeros-1a"},
					DesiredCapacity:      tc.desiredCapacity,
					Instances:            []autoscalingtypes.Instance{},
				},
			}
			groups = asgCache.createPlaceholdersForDesiredNonStartedInstances(groups)
			assert.Equal(t, int32(len(groups[0].Instances)), *tc.desiredCapacity)
			if tc.activities != nil && tc.activities[0].StatusCode == autoscalingtypes.ScalingActivityStatusCodeFailed && tc.activities[0].StartTime.After(tc.groupLastUpdateTime) && asgName == registeredAsgName {
				assert.Equal(t, *groups[0].Instances[0].HealthStatus, placeholderUnfulfillableStatus)
			} else if len(groups[0].Instances) > 0 {
				assert.Equal(t, *groups[0].Instances[0].HealthStatus, "")
			}
			a.AssertExpectations(t)
		})
	}
}

func TestIsTerminatingLifecycleState(t *testing.T) {
	testCases := []struct {
		lifecycle autoscalingtypes.LifecycleState
		expected  bool
	}{
		{autoscalingtypes.LifecycleStateTerminating, true},
		{autoscalingtypes.LifecycleStateTerminatingWait, true},
		{autoscalingtypes.LifecycleStateTerminatingProceed, true},
		{autoscalingtypes.LifecycleStateTerminatingRetained, true},
		{autoscalingtypes.LifecycleStateTerminated, true},
		{autoscalingtypes.LifecycleStatePending, false},
		{autoscalingtypes.LifecycleStateInService, false},
		{autoscalingtypes.LifecycleStateStandby, false},
		{autoscalingtypes.LifecycleStateDetaching, false},
		{"", false},
	}
	for _, tc := range testCases {
		t.Run(string(tc.lifecycle), func(t *testing.T) {
			assert.Equal(t, tc.expected, isTerminatingLifecycleState(tc.lifecycle))
		})
	}
}

func newTestAsgCacheForDelete(a *autoScalingMock, instanceCount int) (*asgCache, *asg, []*AwsInstanceRef) {
	awsService := awsWrapper{a, nil, nil}
	testAsg := &asg{AwsRef: AwsRef{Name: "test-asg"}, curSize: instanceCount}
	cache := &asgCache{
		instanceToAsg:     make(map[AwsInstanceRef]*asg),
		instanceLifecycle: make(map[AwsInstanceRef]autoscalingtypes.LifecycleState),
		awsService:        &awsService,
	}
	instances := make([]*AwsInstanceRef, instanceCount)
	for i := 0; i < instanceCount; i++ {
		ref := AwsInstanceRef{Name: fmt.Sprintf("i-%04d", i), ProviderID: fmt.Sprintf("aws:///us-east-1a/i-%04d", i)}
		cache.instanceToAsg[ref] = testAsg
		cache.instanceLifecycle[ref] = autoscalingtypes.LifecycleStateInService
		instances[i] = &ref
	}
	return cache, testAsg, instances
}

func instanceIdsRange(from, to int) []string {
	ids := make([]string, 0, to-from)
	for i := from; i < to; i++ {
		ids = append(ids, fmt.Sprintf("i-%04d", i))
	}
	return ids
}

func terminateInput(ids []string) *autoscaling.TerminateInstanceInAutoScalingGroupInput {
	return &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		AutoScalingGroupName:           aws.String("test-asg"),
		InstanceIds:                    ids,
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}
}

func TestDeleteInstancesBatches(t *testing.T) {
	a := &autoScalingMock{}
	cache, testAsg, instances := newTestAsgCacheForDelete(a, 250)
	out := &autoscaling.TerminateInstanceInAutoScalingGroupOutput{}

	a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(instanceIdsRange(0, 100))).Return(out, nil).Once()
	a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(instanceIdsRange(100, 200))).Return(out, nil).Once()
	a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(instanceIdsRange(200, 250))).Return(out, nil).Once()

	err := cache.DeleteInstances(instances)
	assert.NoError(t, err)
	a.AssertExpectations(t)
	assert.Equal(t, 0, testAsg.curSize)
	for _, instance := range instances {
		assert.Equal(t, autoscalingtypes.LifecycleStateTerminating, cache.instanceLifecycle[*instance])
	}
}

func TestDeleteInstancesBatchFailureFallsBackToSingleInstances(t *testing.T) {
	a := &autoScalingMock{}
	cache, testAsg, instances := newTestAsgCacheForDelete(a, 150)
	out := &autoscaling.TerminateInstanceInAutoScalingGroupOutput{}
	rejected := errors.New("terminate failed")

	a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(instanceIdsRange(0, 100))).Return(out, nil).Once()
	// the second batch is rejected as a whole because of i-0120, so each instance is retried on its own
	a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(instanceIdsRange(100, 150))).Return(nil, rejected).Once()
	for i := 100; i < 150; i++ {
		ids := instanceIdsRange(i, i+1)
		if i == 120 {
			a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(ids)).Return(nil, rejected).Once()
			continue
		}
		a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(ids)).Return(out, nil).Once()
	}

	err := cache.DeleteInstances(instances)
	assert.ErrorIs(t, err, rejected)
	assert.ErrorContains(t, err, "i-0120")
	a.AssertExpectations(t)
	// every instance except i-0120 is terminated
	assert.Equal(t, 1, testAsg.curSize)
	for i, instance := range instances {
		want := autoscalingtypes.LifecycleStateTerminating
		if i == 120 {
			want = autoscalingtypes.LifecycleStateInService
		}
		assert.Equal(t, want, cache.instanceLifecycle[*instance], instance.Name)
	}
}

func TestDeleteInstancesSingleInstanceFailure(t *testing.T) {
	a := &autoScalingMock{}
	cache, testAsg, instances := newTestAsgCacheForDelete(a, 1)
	rejected := errors.New("terminate failed")

	// a single-instance request is not retried
	a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(instanceIdsRange(0, 1))).Return(nil, rejected).Once()

	err := cache.DeleteInstances(instances)
	assert.ErrorIs(t, err, rejected)
	a.AssertExpectations(t)
	assert.Equal(t, 1, testAsg.curSize)
	assert.Equal(t, autoscalingtypes.LifecycleStateInService, cache.instanceLifecycle[*instances[0]])
}

func TestDeleteInstancesDeduplicatesInstances(t *testing.T) {
	a := &autoScalingMock{}
	cache, testAsg, instances := newTestAsgCacheForDelete(a, 3)
	// the same instance passed twice, once via the same pointer and once via an equal copy
	dup := *instances[1]
	input := append(instances, instances[0], &dup)

	a.On("TerminateInstanceInAutoScalingGroup", mock.Anything, terminateInput(instanceIdsRange(0, 3))).
		Return(&autoscaling.TerminateInstanceInAutoScalingGroupOutput{}, nil).Once()

	err := cache.DeleteInstances(input)
	assert.NoError(t, err)
	a.AssertExpectations(t)
	assert.Equal(t, 0, testAsg.curSize)
}
