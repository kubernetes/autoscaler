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
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
)

type autoScalingMock struct {
	mock.Mock
}

var _ autoScalingI = &autoScalingMock{}

func (a *autoScalingMock) DescribeAutoScalingGroups(ctx context.Context, i *autoscaling.DescribeAutoScalingGroupsInput, opts ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	args := a.Called(ctx, i)
	return args.Get(0).(*autoscaling.DescribeAutoScalingGroupsOutput), nil
}

func (a *autoScalingMock) DescribeLaunchConfigurations(ctx context.Context, i *autoscaling.DescribeLaunchConfigurationsInput, opts ...func(options *autoscaling.Options)) (*autoscaling.DescribeLaunchConfigurationsOutput, error) {
	args := a.Called(ctx, i)
	return args.Get(0).(*autoscaling.DescribeLaunchConfigurationsOutput), nil
}

func (a *autoScalingMock) DescribeScalingActivities(ctx context.Context, i *autoscaling.DescribeScalingActivitiesInput, opts ...func(options *autoscaling.Options)) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	args := a.Called(ctx, i)
	return args.Get(0).(*autoscaling.DescribeScalingActivitiesOutput), args.Error(1)
}

func (a *autoScalingMock) SetDesiredCapacity(ctx context.Context, input *autoscaling.SetDesiredCapacityInput, opts ...func(options *autoscaling.Options)) (*autoscaling.SetDesiredCapacityOutput, error) {
	args := a.Called(ctx, input)
	return args.Get(0).(*autoscaling.SetDesiredCapacityOutput), nil
}

func (a *autoScalingMock) TerminateInstanceInAutoScalingGroup(ctx context.Context, input *autoscaling.TerminateInstanceInAutoScalingGroupInput, opts ...func(options *autoscaling.Options)) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	args := a.Called(ctx, input)
	return args.Get(0).(*autoscaling.TerminateInstanceInAutoScalingGroupOutput), nil
}

type ec2Mock struct {
	mock.Mock
}

var _ ec2I = &ec2Mock{}

func (e *ec2Mock) DescribeImages(ctx context.Context, i *ec2.DescribeImagesInput, opts ...func(options *ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	args := e.Called(ctx, i)
	return args.Get(0).(*ec2.DescribeImagesOutput), nil
}

func (e *ec2Mock) DescribeLaunchTemplateVersions(ctx context.Context, i *ec2.DescribeLaunchTemplateVersionsInput, opts ...func(options *ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	args := e.Called(ctx, i)
	return args.Get(0).(*ec2.DescribeLaunchTemplateVersionsOutput), nil
}

func (e *ec2Mock) GetInstanceTypesFromInstanceRequirements(ctx context.Context, i *ec2.GetInstanceTypesFromInstanceRequirementsInput, opts ...func(options *ec2.Options)) (*ec2.GetInstanceTypesFromInstanceRequirementsOutput, error) {
	args := e.Called(ctx, i)
	return args.Get(0).(*ec2.GetInstanceTypesFromInstanceRequirementsOutput), nil
}

type eksMock struct {
	mock.Mock
}

var _ eksI = &eksMock{}

func (k *eksMock) DescribeNodegroup(ctx context.Context, i *eks.DescribeNodegroupInput, opts ...func(options *eks.Options)) (*eks.DescribeNodegroupOutput, error) {
	args := k.Called(ctx, i)

	if args.Get(0) == nil && args.Get(1) == nil {
		return nil, nil
	} else if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	} else if args.Get(1) == nil {
		return args.Get(0).(*eks.DescribeNodegroupOutput), nil
	} else {
		return args.Get(0).(*eks.DescribeNodegroupOutput), args.Get(1).(error)
	}
}

var testAwsService = awsWrapper{&autoScalingMock{}, &ec2Mock{}, &eksMock{}}

func TestGetManagedNodegroup(t *testing.T) {
	k := &eksMock{}
	awsWrapper := &awsWrapper{
		autoScalingI: nil,
		ec2I:         nil,
		eksI:         k,
	}

	labelKey1 := "labelKey 1"
	labelKey2 := "labelKey 2"
	labelValue1 := "testValue 1"
	labelValue2 := "testValue 2"
	nodegroupName := "testNodegroup"
	clusterName := "testCluster"

	taintEffect1 := ekstypes.TaintEffectNoSchedule
	taintEffectTranslated1 := apiv1.TaintEffectNoSchedule
	taintKey1 := "key 1"
	taintValue1 := "value 1"
	taint1 := ekstypes.Taint{
		Effect: taintEffect1,
		Key:    &taintKey1,
		Value:  &taintValue1,
	}

	taintEffect2 := ekstypes.TaintEffectNoExecute
	taintEffectTranslated2 := apiv1.TaintEffectNoExecute
	taintKey2 := "key 2"
	taintValue2 := "value 2"
	taint2 := ekstypes.Taint{
		Effect: taintEffect2,
		Key:    &taintKey2,
		Value:  &taintValue2,
	}

	amiType := "testAmiType"
	diskSize := int32(100)
	capacityType := "testCapacityType"
	k8sVersion := "1.19"

	tagKey1 := "tag 1"
	tagValue1 := "value 1"
	tagKey2 := "tag 2"
	tagValue2 := "value 2"

	// Create test nodegroup
	testNodegroup := ekstypes.Nodegroup{
		AmiType:       ekstypes.AMITypes(amiType),
		ClusterName:   &clusterName,
		DiskSize:      &diskSize,
		Labels:        map[string]string{labelKey1: labelValue1, labelKey2: labelValue2},
		NodegroupName: &nodegroupName,
		CapacityType:  ekstypes.CapacityTypes(capacityType),
		Version:       &k8sVersion,
		Taints:        []ekstypes.Taint{taint1, taint2},
		Tags:          map[string]string{tagKey1: tagValue1, tagKey2: tagValue2},
	}

	k.On("DescribeNodegroup",
		mock.Anything,
		&eks.DescribeNodegroupInput{
			ClusterName:   &clusterName,
			NodegroupName: &nodegroupName,
		},
	).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	taintList, labelMap, tagMap, err := awsWrapper.getManagedNodegroupInfo(nodegroupName, clusterName)
	assert.Nil(t, err)
	assert.Equal(t, len(taintList), 2)
	assert.Equal(t, taintList[0].Effect, taintEffectTranslated1)
	assert.Equal(t, taintList[0].Key, taintKey1)
	assert.Equal(t, taintList[0].Value, taintValue1)
	assert.Equal(t, taintList[1].Effect, taintEffectTranslated2)
	assert.Equal(t, taintList[1].Key, taintKey2)
	assert.Equal(t, taintList[1].Value, taintValue2)
	assert.Equal(t, len(labelMap), 7)
	assert.Equal(t, labelMap[labelKey1], labelValue1)
	assert.Equal(t, labelMap[labelKey2], labelValue2)
	assert.Equal(t, labelMap["diskSize"], strconv.FormatInt(int64(diskSize), 10))
	assert.Equal(t, labelMap["amiType"], amiType)
	assert.Equal(t, labelMap["eks.amazonaws.com/capacityType"], capacityType)
	assert.Equal(t, labelMap["k8sVersion"], k8sVersion)
	assert.Equal(t, labelMap["eks.amazonaws.com/nodegroup"], nodegroupName)
	assert.Equal(t, len(tagMap), 2)
	assert.Equal(t, tagMap[tagKey1], tagValue1)
	assert.Equal(t, tagMap[tagKey2], tagValue2)
}

func TestGetManagedNodegroupWithNilValues(t *testing.T) {
	k := &eksMock{}
	awsWrapper := &awsWrapper{
		autoScalingI: nil,
		ec2I:         nil,
		eksI:         k,
	}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"

	amiType := "testAmiType"
	capacityType := "testCapacityType"
	k8sVersion := "1.19"

	// Create test nodegroup
	testNodegroup := ekstypes.Nodegroup{
		AmiType:       ekstypes.AMITypes(amiType),
		ClusterName:   &clusterName,
		DiskSize:      nil,
		Labels:        nil,
		NodegroupName: &nodegroupName,
		CapacityType:  ekstypes.CapacityTypes(capacityType),
		Version:       &k8sVersion,
		Taints:        nil,
		Tags:          nil,
	}

	k.On("DescribeNodegroup",
		mock.Anything,
		&eks.DescribeNodegroupInput{
			ClusterName:   &clusterName,
			NodegroupName: &nodegroupName,
		},
	).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	taintList, labelMap, tagMap, err := awsWrapper.getManagedNodegroupInfo(nodegroupName, clusterName)
	assert.Nil(t, err)
	assert.Equal(t, len(taintList), 0)
	assert.Equal(t, len(labelMap), 4)
	assert.Equal(t, labelMap["amiType"], amiType)
	assert.Equal(t, labelMap["eks.amazonaws.com/capacityType"], capacityType)
	assert.Equal(t, labelMap["k8sVersion"], k8sVersion)
	assert.Equal(t, labelMap["eks.amazonaws.com/nodegroup"], nodegroupName)
	assert.Equal(t, len(tagMap), 0)
}

func TestGetManagedNodegroupWithEmptyValues(t *testing.T) {
	k := &eksMock{}
	awsWrapper := &awsWrapper{
		autoScalingI: nil,
		ec2I:         nil,
		eksI:         k,
	}

	nodegroupName := "testNodegroup"
	clusterName := "testCluster"

	amiType := "testAmiType"
	capacityType := "testCapacityType"
	k8sVersion := "1.19"

	// Create test nodegroup
	testNodegroup := ekstypes.Nodegroup{
		AmiType:       ekstypes.AMITypes(amiType),
		ClusterName:   &clusterName,
		DiskSize:      nil,
		Labels:        make(map[string]string),
		NodegroupName: &nodegroupName,
		CapacityType:  ekstypes.CapacityTypes(capacityType),
		Version:       &k8sVersion,
		Taints:        make([]ekstypes.Taint, 0),
		Tags:          make(map[string]string),
	}

	k.On("DescribeNodegroup",
		mock.Anything,
		&eks.DescribeNodegroupInput{
			ClusterName:   &clusterName,
			NodegroupName: &nodegroupName,
		},
	).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	taintList, labelMap, tagMap, err := awsWrapper.getManagedNodegroupInfo(nodegroupName, clusterName)
	assert.Nil(t, err)
	assert.Equal(t, len(taintList), 0)
	assert.Equal(t, len(labelMap), 4)
	assert.Equal(t, labelMap["amiType"], amiType)
	assert.Equal(t, labelMap["eks.amazonaws.com/capacityType"], capacityType)
	assert.Equal(t, labelMap["k8sVersion"], k8sVersion)
	assert.Equal(t, labelMap["eks.amazonaws.com/nodegroup"], nodegroupName)
	assert.Equal(t, len(tagMap), 0)
}

func TestMoreThen100Groups(t *testing.T) {
	a := &autoScalingMock{}
	awsWrapper := &awsWrapper{
		autoScalingI: a,
		ec2I:         nil,
		eksI:         nil,
	}

	// Generate 101 ASG names
	names := make([]string, 101)
	for i := range names {
		names[i] = fmt.Sprintf("asg-%d", i)
	}

	// First batch, first 100 elements
	a.On("DescribeAutoScalingGroups",
		mock.Anything,
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: names[:100],
			MaxRecords:            aws.Int32(maxRecordsReturnedByAPI),
		},
	).Return(testNamedDescribeAutoScalingGroupsOutput("asg-1", 1, "test-instance-id"), nil)

	// Second batch, element 101
	a.On("DescribeAutoScalingGroups",
		mock.Anything,
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []string{"asg-100"},
			MaxRecords:            aws.Int32(maxRecordsReturnedByAPI),
		},
	).Return(testNamedDescribeAutoScalingGroupsOutput("asg-2", 1, "test-instance-id"), nil)

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
	a.On("DescribeLaunchConfigurations",
		mock.Anything,
		&autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []string{ltName},
			MaxRecords:               aws.Int32(50),
		},
	).Return(
		&autoscaling.DescribeLaunchConfigurationsOutput{
			LaunchConfigurations: []autoscalingtypes.LaunchConfiguration{
				{
					LaunchConfigurationName: aws.String(ltName),
					InstanceType:            aws.String(instanceType),
				},
			},
		},
		nil,
	)
	e.On("DescribeLaunchTemplateVersions",
		mock.Anything,
		&ec2.DescribeLaunchTemplateVersionsInput{
			LaunchTemplateName: aws.String(ltName),
			Versions:           []string{ltVersion},
		},
	).Return(
		&ec2.DescribeLaunchTemplateVersionsOutput{
			LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
				{
					LaunchTemplateData: &ec2types.ResponseLaunchTemplateData{
						InstanceType: ec2types.InstanceType(instanceType),
					},
				},
			},
		},
		nil,
	)

	t.Setenv("AWS_REGION", "fanghorn")

	awsWrapper := &awsWrapper{
		autoScalingI: a,
		ec2I:         e,
		eksI:         nil,
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

func TestGetInstanceTypesFromInstanceRequirementsOverrides(t *testing.T) {
	mixedInstancesPolicy := &mixedInstancesPolicy{
		launchTemplate: &launchTemplate{
			name:    "launchTemplateName",
			version: "1",
		},
		instanceRequirementsOverrides: &autoscalingtypes.InstanceRequirements{
			VCpuCount: &autoscalingtypes.VCpuCountRequest{
				Min: aws.Int32(4),
				Max: aws.Int32(8),
			},
			MemoryMiB: &autoscalingtypes.MemoryMiBRequest{
				Min: aws.Int32(4),
				Max: aws.Int32(8),
			},
			AcceleratorTypes:         []autoscalingtypes.AcceleratorType{autoscalingtypes.AcceleratorTypeGpu},
			AcceleratorManufacturers: []autoscalingtypes.AcceleratorManufacturer{autoscalingtypes.AcceleratorManufacturerNvidia},
			AcceleratorCount: &autoscalingtypes.AcceleratorCountRequest{
				Min: aws.Int32(4),
				Max: aws.Int32(8),
			},
		},
	}

	e := &ec2Mock{}
	awsWrapper := &awsWrapper{
		autoScalingI: nil,
		ec2I:         e,
		eksI:         nil,
	}

	e.On("DescribeLaunchTemplateVersions",
		mock.Anything,
		&ec2.DescribeLaunchTemplateVersionsInput{
			LaunchTemplateName: aws.String("launchTemplateName"),
			Versions:           []string{"1"},
		},
	).Return(
		&ec2.DescribeLaunchTemplateVersionsOutput{
			LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
				{
					LaunchTemplateData: &ec2types.ResponseLaunchTemplateData{
						ImageId: aws.String("123"),
					},
				},
			},
		},
		nil,
	)

	e.On("DescribeImages",
		mock.Anything,
		&ec2.DescribeImagesInput{
			ImageIds: []string{"123"},
		},
	).Return(
		&ec2.DescribeImagesOutput{
			Images: []ec2types.Image{
				{
					Architecture:       ec2types.ArchitectureValuesX8664,
					VirtualizationType: "xen",
				},
			},
		},
		nil,
	)

	requirements, err := awsWrapper.getRequirementsRequestFromAutoscaling(mixedInstancesPolicy.instanceRequirementsOverrides)
	assert.NoError(t, err)
	e.On("GetInstanceTypesFromInstanceRequirements",
		mock.Anything,
		&ec2.GetInstanceTypesFromInstanceRequirementsInput{
			ArchitectureTypes:    []ec2types.ArchitectureType{ec2types.ArchitectureTypeX8664},
			InstanceRequirements: requirements,
			VirtualizationTypes:  []ec2types.VirtualizationType{"xen"},
		},
	).Return(
		&ec2.GetInstanceTypesFromInstanceRequirementsOutput{
			InstanceTypes: []ec2types.InstanceTypeInfoFromInstanceRequirements{
				{
					InstanceType: aws.String("g4dn.xlarge"),
				},
			},
		},
		nil,
	)

	result, err := awsWrapper.getInstanceTypeFromRequirementsOverrides(mixedInstancesPolicy)
	assert.NoError(t, err)
	assert.Equal(t, "g4dn.xlarge", result)
}

func TestGetInstanceTypesFromInstanceRequirementsInLaunchTemplate(t *testing.T) {
	launchTemplate := &launchTemplate{
		name:    "launchTemplateName",
		version: "1",
	}

	e := &ec2Mock{}
	awsWrapper := &awsWrapper{
		autoScalingI: nil,
		ec2I:         e,
		eksI:         nil,
	}

	instanceRequirements := &ec2types.InstanceRequirements{
		VCpuCount: &ec2types.VCpuCountRange{
			Min: aws.Int32(4),
			Max: aws.Int32(8),
		},
		MemoryMiB: &ec2types.MemoryMiB{
			Min: aws.Int32(4),
			Max: aws.Int32(8),
		},
		AcceleratorTypes:         []ec2types.AcceleratorType{ec2types.AcceleratorTypeGpu},
		AcceleratorManufacturers: []ec2types.AcceleratorManufacturer{ec2types.AcceleratorManufacturerNvidia},
		AcceleratorCount: &ec2types.AcceleratorCount{
			Min: aws.Int32(4),
			Max: aws.Int32(8),
		},
	}

	e.On("DescribeLaunchTemplateVersions",
		mock.Anything,
		&ec2.DescribeLaunchTemplateVersionsInput{
			LaunchTemplateName: aws.String("launchTemplateName"),
			Versions:           []string{"1"},
		},
	).Return(
		&ec2.DescribeLaunchTemplateVersionsOutput{
			LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
				{
					LaunchTemplateData: &ec2types.ResponseLaunchTemplateData{
						ImageId:              aws.String("123"),
						InstanceRequirements: instanceRequirements,
					},
				},
			},
		},
		nil,
	)

	e.On("DescribeImages",
		mock.Anything,
		&ec2.DescribeImagesInput{
			ImageIds: []string{"123"},
		},
	).Return(
		&ec2.DescribeImagesOutput{
			Images: []ec2types.Image{
				{
					Architecture:       ec2types.ArchitectureValuesX8664,
					VirtualizationType: "xen",
				},
			},
		},
		nil,
	)

	requirements, err := awsWrapper.getRequirementsRequestFromEC2(instanceRequirements)
	assert.NoError(t, err)
	e.On("GetInstanceTypesFromInstanceRequirements",
		mock.Anything,
		&ec2.GetInstanceTypesFromInstanceRequirementsInput{
			ArchitectureTypes:    []ec2types.ArchitectureType{ec2types.ArchitectureTypeX8664},
			InstanceRequirements: requirements,
			VirtualizationTypes:  []ec2types.VirtualizationType{"xen"},
		},
	).Return(
		&ec2.GetInstanceTypesFromInstanceRequirementsOutput{
			InstanceTypes: []ec2types.InstanceTypeInfoFromInstanceRequirements{
				{
					InstanceType: aws.String("g4dn.xlarge"),
				},
			},
		},
		nil,
	)

	result, err := awsWrapper.getInstanceTypeByLaunchTemplate(launchTemplate)
	assert.NoError(t, err)
	assert.Equal(t, "g4dn.xlarge", result)
}

func TestGetLaunchTemplateData(t *testing.T) {
	e := &ec2Mock{}
	awsWrapper := &awsWrapper{
		ec2I: e,
	}

	testCases := []struct {
		testName             string
		describeTemplateData *ec2.DescribeLaunchTemplateVersionsOutput
		expectedData         *ec2types.ResponseLaunchTemplateData
		expectedErr          error
	}{
		{
			"no launch template version found",
			&ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{},
			},
			nil,
			errors.New("unable to find template versions for launch template launchTemplateName"),
		},
		{
			"no data found for launch template",
			&ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
					{
						LaunchTemplateName: aws.String("launchTemplateName"),
						LaunchTemplateData: nil,
					},
				},
			},
			nil,
			errors.New("no data found for launch template launchTemplateName, version 1"),
		},
		{
			"launch template data found successfully",
			&ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
					{
						LaunchTemplateName: aws.String("launchTemplateName"),
						LaunchTemplateData: &ec2types.ResponseLaunchTemplateData{
							ImageId: aws.String("123"),
						},
					},
				},
			},
			&ec2types.ResponseLaunchTemplateData{
				ImageId: aws.String("123"),
			},
			nil,
		},
	}

	describeTemplateInput := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String("launchTemplateName"),
		Versions:           []string{"1"},
	}

	for _, testCase := range testCases {
		e.On("DescribeLaunchTemplateVersions",
			mock.Anything,
			describeTemplateInput,
		).Return(
			testCase.describeTemplateData,
			nil,
		).Once()

		describeData, err := awsWrapper.getLaunchTemplateData("launchTemplateName", "1")
		assert.Equal(t, testCase.expectedData, describeData)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestBuildLaunchTemplateFromSpec(t *testing.T) {
	assert := assert.New(t)

	units := []struct {
		name string
		in   *autoscalingtypes.LaunchTemplateSpecification
		exp  *launchTemplate
	}{
		{
			name: "non-default, specified version",
			in: &autoscalingtypes.LaunchTemplateSpecification{
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
			in: &autoscalingtypes.LaunchTemplateSpecification{
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
			in: &autoscalingtypes.LaunchTemplateSpecification{
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
			in: &autoscalingtypes.LaunchTemplateSpecification{
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

func TestGetInstanceTypesFromInstanceRequirementsWithEmptyList(t *testing.T) {
	e := &ec2Mock{}
	awsWrapper := &awsWrapper{
		autoScalingI: nil,
		ec2I:         e,
		eksI:         nil,
	}
	requirements := &ec2types.InstanceRequirementsRequest{}

	e.On("DescribeImages",
		mock.Anything,
		&ec2.DescribeImagesInput{
			ImageIds: []string{"123"},
		},
	).Return(
		&ec2.DescribeImagesOutput{
			Images: []ec2types.Image{
				{
					Architecture:       ec2types.ArchitectureValuesX8664,
					VirtualizationType: "xen",
				},
			},
		},
		nil,
	)
	e.On("GetInstanceTypesFromInstanceRequirements",
		mock.Anything,
		&ec2.GetInstanceTypesFromInstanceRequirementsInput{
			ArchitectureTypes:    []ec2types.ArchitectureType{ec2types.ArchitectureTypeX8664},
			InstanceRequirements: requirements,
			VirtualizationTypes:  []ec2types.VirtualizationType{"xen"},
		},
	).Return(
		&ec2.GetInstanceTypesFromInstanceRequirementsOutput{
			InstanceTypes: []ec2types.InstanceTypeInfoFromInstanceRequirements{},
		},
		nil,
	)

	result, err := awsWrapper.getInstanceTypeFromInstanceRequirements("123", requirements)
	assert.Error(t, err)
	exp := fmt.Errorf("no instance types found for requirements")
	assert.EqualError(t, err, exp.Error())
	assert.Equal(t, "", result)
}

func TestTaintEksTranslator(t *testing.T) {
	key := "key"
	value := "value"

	taintEffect1 := ekstypes.TaintEffectNoSchedule
	taintEffectTranslated1 := apiv1.TaintEffectNoSchedule
	taint1 := ekstypes.Taint{
		Effect: taintEffect1,
		Key:    &key,
		Value:  &value,
	}

	t1, err := taintEksTranslator(taint1)
	assert.Nil(t, err)
	assert.Equal(t, t1, taintEffectTranslated1)

	taintEffect2 := ekstypes.TaintEffectNoSchedule
	taintEffectTranslated2 := apiv1.TaintEffectNoSchedule
	taint2 := ekstypes.Taint{
		Effect: taintEffect2,
		Key:    &key,
		Value:  &value,
	}
	t2, err := taintEksTranslator(taint2)
	assert.Nil(t, err)
	assert.Equal(t, t2, taintEffectTranslated2)

	taintEffect3 := ekstypes.TaintEffectNoExecute
	taintEffectTranslated3 := apiv1.TaintEffectNoExecute
	taint3 := ekstypes.Taint{
		Effect: taintEffect3,
		Key:    &key,
		Value:  &value,
	}
	t3, err := taintEksTranslator(taint3)
	assert.Nil(t, err)
	assert.Equal(t, t3, taintEffectTranslated3)

	taintEffect4 := "TAINT_NO_EXISTS"
	taint4 := ekstypes.Taint{
		Effect: ekstypes.TaintEffect(taintEffect4),
		Key:    &key,
		Value:  &value,
	}
	_, err = taintEksTranslator(taint4)
	assert.Error(t, err)
}
