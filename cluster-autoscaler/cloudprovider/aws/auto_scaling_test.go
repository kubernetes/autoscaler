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
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMoreThen100Groups(t *testing.T) {
	service := &AutoScalingMock{}
	autoScalingWrapper := &autoScalingWrapper{
		autoScaling: service,
	}

	// Generate 101 ASG names
	names := make([]string, 101)
	for i := 0; i < len(names); i++ {
		names[i] = fmt.Sprintf("asg-%d", i)
	}

	// First batch, first 100 elements
	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice(names[:100]),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("asg-1", 1, "test-instance-id"), false)
	}).Return(nil)

	// Second batch, element 101
	service.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"asg-100"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("asg-2", 1, "test-instance-id"), false)
	}).Return(nil)

	asgs, err := autoScalingWrapper.getAutoscalingGroupsByNames(names)
	assert.Nil(t, err)
	assert.Equal(t, len(asgs), 2)
	assert.Equal(t, *asgs[0].AutoScalingGroupName, "asg-1")
	assert.Equal(t, *asgs[1].AutoScalingGroupName, "asg-2")
}

func TestLaunchConfigurationCache(t *testing.T) {
	c := newLaunchConfigurationInstanceTypeCache()
	err := c.Add(instanceTypeCachedObject{
		name:         "123",
		instanceType: "t2.medium",
	})
	require.NoError(t, err)
	obj, ok, err := c.GetByKey("123")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "t2.medium", obj.(instanceTypeCachedObject).instanceType)
}
