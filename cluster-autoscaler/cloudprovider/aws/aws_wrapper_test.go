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
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type autoScalingMock struct {
	mock.Mock
}

func (a *autoScalingMock) DescribeAutoScalingGroupsPages(i *autoscaling.DescribeAutoScalingGroupsInput, fn func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool) error {
	args := a.Called(i, fn)
	return args.Error(0)
}

func (a *autoScalingMock) DescribeLaunchConfigurations(i *autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	args := a.Called(i)
	return args.Get(0).(*autoscaling.DescribeLaunchConfigurationsOutput), nil
}

func (a *autoScalingMock) DescribeTagsPages(i *autoscaling.DescribeTagsInput, fn func(*autoscaling.DescribeTagsOutput, bool) bool) error {
	args := a.Called(i, fn)
	return args.Error(0)
}

func (a *autoScalingMock) SetDesiredCapacity(input *autoscaling.SetDesiredCapacityInput) (*autoscaling.SetDesiredCapacityOutput, error) {
	args := a.Called(input)
	return args.Get(0).(*autoscaling.SetDesiredCapacityOutput), nil
}

func (a *autoScalingMock) TerminateInstanceInAutoScalingGroup(input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	args := a.Called(input)
	return args.Get(0).(*autoscaling.TerminateInstanceInAutoScalingGroupOutput), nil
}

type ec2Mock struct {
	mock.Mock
}

func (e *ec2Mock) DescribeLaunchTemplateVersions(i *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	args := e.Called(i)
	return args.Get(0).(*ec2.DescribeLaunchTemplateVersionsOutput), nil
}

var testAwsService = awsWrapper{&autoScalingMock{}, &ec2Mock{}}

func TestMoreThen100Groups(t *testing.T) {
	a := &autoScalingMock{}
	awsWrapper := &awsWrapper{
		autoScalingI: a,
		ec2I:         nil,
	}

	// Generate 101 ASG names
	names := make([]string, 101)
	for i := 0; i < len(names); i++ {
		names[i] = fmt.Sprintf("asg-%d", i)
	}

	// First batch, first 100 elements
	a.On("DescribeAutoScalingGroupsPages",
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
	a.On("DescribeAutoScalingGroupsPages",
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice([]string{"asg-100"}),
			MaxRecords:            aws.Int64(maxRecordsReturnedByAPI),
		},
		mock.AnythingOfType("func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*autoscaling.DescribeAutoScalingGroupsOutput, bool) bool)
		fn(testNamedDescribeAutoScalingGroupsOutput("asg-2", 1, "test-instance-id"), false)
	}).Return(nil)

	asgs, err := awsWrapper.getAutoscalingGroupsByNames(names)
	assert.Nil(t, err)
	assert.Equal(t, len(asgs), 2)
	assert.Equal(t, *asgs[0].AutoScalingGroupName, "asg-1")
	assert.Equal(t, *asgs[1].AutoScalingGroupName, "asg-2")
}

func TestGetInstanceTypesForAsgs(t *testing.T) {
	asgName, ltName, ltVersion, instanceType := "testasg", "launcher", "1", "t2.large"
	ltSpec := launchTemplate{
		name:    ltName,
		version: ltVersion,
	}

	a := &autoScalingMock{}
	e := &ec2Mock{}
	a.On("DescribeLaunchConfigurations", &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{aws.String(ltName)},
		MaxRecords:               aws.Int64(50),
	}).Return(&autoscaling.DescribeLaunchConfigurationsOutput{
		LaunchConfigurations: []*autoscaling.LaunchConfiguration{
			{
				LaunchConfigurationName: aws.String(ltName),
				InstanceType:            aws.String(instanceType),
			},
		},
	})
	e.On("DescribeLaunchTemplateVersions", &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String(ltName),
		Versions:           []*string{aws.String(ltVersion)},
	}).Return(&ec2.DescribeLaunchTemplateVersionsOutput{
		LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
			{
				LaunchTemplateData: &ec2.ResponseLaunchTemplateData{
					InstanceType: aws.String(instanceType),
				},
			},
		},
	})

	// #1449 Without AWS_REGION getRegion() lookup runs till timeout during tests.
	defer resetAWSRegion(os.LookupEnv("AWS_REGION"))
	os.Setenv("AWS_REGION", "fanghorn")

	awsWrapper := &awsWrapper{
		autoScalingI: a,
		ec2I:         e,
	}

	cases := []struct {
		name                    string
		launchConfigurationName string
		launchTemplate          *launchTemplate
		mixedInstancesPolicy    *mixedInstancesPolicy
	}{
		{
			"AsgWithLaunchConfiguration",
			ltName,
			nil,
			nil,
		},
		{
			"AsgWithLaunchTemplate",
			"",
			&ltSpec,
			nil,
		},
		{
			"AsgWithLaunchTemplateMixedInstancePolicyOverride",
			"",
			nil,
			&mixedInstancesPolicy{
				instanceTypesOverrides: []string{instanceType},
			},
		},
		{
			"AsgWithLaunchTemplateMixedInstancePolicyNoOverride",
			"",
			nil,
			&mixedInstancesPolicy{
				launchTemplate: &ltSpec,
			},
		},
	}

	for _, tc := range cases {
		results, err := awsWrapper.getInstanceTypesForAsgs([]*asg{
			{
				AwsRef:                  AwsRef{Name: asgName},
				LaunchConfigurationName: tc.launchConfigurationName,
				LaunchTemplate:          tc.launchTemplate,
				MixedInstancesPolicy:    tc.mixedInstancesPolicy,
			},
		})
		assert.NoError(t, err)

		foundInstanceType, exists := results[asgName]
		assert.NoErrorf(t, err, "%s had error %v", tc.name, err)
		assert.Truef(t, exists, "%s did not find asg", tc.name)
		assert.Equalf(t, foundInstanceType, instanceType, "%s had %s, expected %s", tc.name, foundInstanceType, instanceType)
	}
}

func TestBuildLaunchTemplateFromSpec(t *testing.T) {
	assert := assert.New(t)

	units := []struct {
		name string
		in   *autoscaling.LaunchTemplateSpecification
		exp  *launchTemplate
	}{
		{
			name: "non-default, specified version",
			in: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: aws.String("foo"),
				Version:            aws.String("1"),
			},
			exp: &launchTemplate{
				name:    "foo",
				version: "1",
			},
		},
		{
			name: "non-default, specified $Latest",
			in: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: aws.String("foo"),
				Version:            aws.String("$Latest"),
			},
			exp: &launchTemplate{
				name:    "foo",
				version: "$Latest",
			},
		},
		{
			name: "specified $Default",
			in: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: aws.String("foo"),
				Version:            aws.String("$Default"),
			},
			exp: &launchTemplate{
				name:    "foo",
				version: "$Default",
			},
		},
		{
			name: "no version specified",
			in: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateName: aws.String("foo"),
				Version:            nil,
			},
			exp: &launchTemplate{
				name:    "foo",
				version: "$Default",
			},
		},
	}

	for _, unit := range units {
		got := buildLaunchTemplateFromSpec(unit.in)
		assert.Equal(unit.exp, got)
	}
}
