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
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/ec2"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/eks"
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

func (a *autoScalingMock) DescribeScalingActivities(i *autoscaling.DescribeScalingActivitiesInput) (*autoscaling.DescribeScalingActivitiesOutput, error) {
	args := a.Called(i)
	return args.Get(0).(*autoscaling.DescribeScalingActivitiesOutput), args.Error(1)
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

func (e *ec2Mock) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	args := e.Called(input)
	return args.Get(0).(*ec2.DescribeImagesOutput), nil
}

func (e *ec2Mock) DescribeLaunchTemplateVersions(i *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	args := e.Called(i)
	return args.Get(0).(*ec2.DescribeLaunchTemplateVersionsOutput), nil
}

func (e *ec2Mock) GetInstanceTypesFromInstanceRequirementsPages(input *ec2.GetInstanceTypesFromInstanceRequirementsInput, fn func(*ec2.GetInstanceTypesFromInstanceRequirementsOutput, bool) bool) error {
	args := e.Called(input, fn)
	return args.Error(0)
}

type eksMock struct {
	mock.Mock
}

func (k *eksMock) DescribeNodegroup(i *eks.DescribeNodegroupInput) (*eks.DescribeNodegroupOutput, error) {
	args := k.Called(i)

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

	taintEffect1 := "effect 1"
	taintKey1 := "key 1"
	taintValue1 := "value 1"
	taint1 := eks.Taint{
		Effect: &taintEffect1,
		Key:    &taintKey1,
		Value:  &taintValue1,
	}

	taintEffect2 := "effect 2"
	taintKey2 := "key 2"
	taintValue2 := "value 2"
	taint2 := eks.Taint{
		Effect: &taintEffect2,
		Key:    &taintKey2,
		Value:  &taintValue2,
	}

	amiType := "testAmiType"
	diskSize := int64(100)
	capacityType := "testCapacityType"
	k8sVersion := "1.19"

	// Create test nodegroup
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      &diskSize,
		Labels:        map[string]*string{labelKey1: &labelValue1, labelKey2: &labelValue2},
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        []*eks.Taint{&taint1, &taint2},
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	taintList, labelMap, err := awsWrapper.getManagedNodegroupInfo(nodegroupName, clusterName)
	assert.Nil(t, err)
	assert.Equal(t, len(taintList), 2)
	assert.Equal(t, taintList[0].Effect, apiv1.TaintEffect(taintEffect1))
	assert.Equal(t, taintList[0].Key, taintKey1)
	assert.Equal(t, taintList[0].Value, taintValue1)
	assert.Equal(t, taintList[1].Effect, apiv1.TaintEffect(taintEffect2))
	assert.Equal(t, taintList[1].Key, taintKey2)
	assert.Equal(t, taintList[1].Value, taintValue2)
	assert.Equal(t, len(labelMap), 7)
	assert.Equal(t, labelMap[labelKey1], labelValue1)
	assert.Equal(t, labelMap[labelKey2], labelValue2)
	assert.Equal(t, labelMap["diskSize"], strconv.FormatInt(diskSize, 10))
	assert.Equal(t, labelMap["amiType"], amiType)
	assert.Equal(t, labelMap["capacityType"], capacityType)
	assert.Equal(t, labelMap["k8sVersion"], k8sVersion)
	assert.Equal(t, labelMap["eks.amazonaws.com/nodegroup"], nodegroupName)
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
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      nil,
		Labels:        nil,
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        nil,
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	taintList, labelMap, err := awsWrapper.getManagedNodegroupInfo(nodegroupName, clusterName)
	assert.Nil(t, err)
	assert.Equal(t, len(taintList), 0)
	assert.Equal(t, len(labelMap), 4)
	assert.Equal(t, labelMap["amiType"], amiType)
	assert.Equal(t, labelMap["capacityType"], capacityType)
	assert.Equal(t, labelMap["k8sVersion"], k8sVersion)
	assert.Equal(t, labelMap["eks.amazonaws.com/nodegroup"], nodegroupName)
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
	testNodegroup := eks.Nodegroup{
		AmiType:       &amiType,
		ClusterName:   &clusterName,
		DiskSize:      nil,
		Labels:        make(map[string]*string),
		NodegroupName: &nodegroupName,
		CapacityType:  &capacityType,
		Version:       &k8sVersion,
		Taints:        make([]*eks.Taint, 0),
	}

	k.On("DescribeNodegroup", &eks.DescribeNodegroupInput{
		ClusterName:   &clusterName,
		NodegroupName: &nodegroupName,
	}).Return(&eks.DescribeNodegroupOutput{Nodegroup: &testNodegroup}, nil)

	taintList, labelMap, err := awsWrapper.getManagedNodegroupInfo(nodegroupName, clusterName)
	assert.Nil(t, err)
	assert.Equal(t, len(taintList), 0)
	assert.Equal(t, len(labelMap), 4)
	assert.Equal(t, labelMap["amiType"], amiType)
	assert.Equal(t, labelMap["capacityType"], capacityType)
	assert.Equal(t, labelMap["k8sVersion"], k8sVersion)
	assert.Equal(t, labelMap["eks.amazonaws.com/nodegroup"], nodegroupName)
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
		instanceRequirementsOverrides: &autoscaling.InstanceRequirements{
			VCpuCount: &autoscaling.VCpuCountRequest{
				Min: aws.Int64(4),
				Max: aws.Int64(8),
			},
			MemoryMiB: &autoscaling.MemoryMiBRequest{
				Min: aws.Int64(4),
				Max: aws.Int64(8),
			},
			AcceleratorTypes:         []*string{aws.String(autoscaling.AcceleratorTypeGpu)},
			AcceleratorManufacturers: []*string{aws.String(autoscaling.AcceleratorManufacturerNvidia)},
			AcceleratorCount: &autoscaling.AcceleratorCountRequest{
				Min: aws.Int64(4),
				Max: aws.Int64(8),
			},
		},
	}

	e := &ec2Mock{}
	awsWrapper := &awsWrapper{
		autoScalingI: nil,
		ec2I:         e,
		eksI:         nil,
	}

	e.On("DescribeLaunchTemplateVersions", &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String("launchTemplateName"),
		Versions:           []*string{aws.String("1")},
	}).Return(&ec2.DescribeLaunchTemplateVersionsOutput{
		LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
			{
				LaunchTemplateData: &ec2.ResponseLaunchTemplateData{
					ImageId: aws.String("123"),
				},
			},
		},
	})

	e.On("DescribeImages", &ec2.DescribeImagesInput{
		ImageIds: []*string{aws.String("123")},
	}).Return(&ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			{
				Architecture:       aws.String("x86_64"),
				VirtualizationType: aws.String("xen"),
			},
		},
	})

	requirements, err := awsWrapper.getRequirementsRequestFromAutoscaling(mixedInstancesPolicy.instanceRequirementsOverrides)
	assert.NoError(t, err)
	e.On("GetInstanceTypesFromInstanceRequirementsPages",
		&ec2.GetInstanceTypesFromInstanceRequirementsInput{
			ArchitectureTypes:    []*string{aws.String("x86_64")},
			InstanceRequirements: requirements,
			VirtualizationTypes:  []*string{aws.String("xen")},
		},
		mock.AnythingOfType("func(*ec2.GetInstanceTypesFromInstanceRequirementsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*ec2.GetInstanceTypesFromInstanceRequirementsOutput, bool) bool)
		fn(&ec2.GetInstanceTypesFromInstanceRequirementsOutput{
			InstanceTypes: []*ec2.InstanceTypeInfoFromInstanceRequirements{
				{
					InstanceType: aws.String("g4dn.xlarge"),
				},
			},
		}, false)
	}).Return(nil)

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

	instanceRequirements := &ec2.InstanceRequirements{
		VCpuCount: &ec2.VCpuCountRange{
			Min: aws.Int64(4),
			Max: aws.Int64(8),
		},
		MemoryMiB: &ec2.MemoryMiB{
			Min: aws.Int64(4),
			Max: aws.Int64(8),
		},
		AcceleratorTypes:         []*string{aws.String(autoscaling.AcceleratorTypeGpu)},
		AcceleratorManufacturers: []*string{aws.String(autoscaling.AcceleratorManufacturerNvidia)},
		AcceleratorCount: &ec2.AcceleratorCount{
			Min: aws.Int64(4),
			Max: aws.Int64(8),
		},
	}

	e.On("DescribeLaunchTemplateVersions", &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String("launchTemplateName"),
		Versions:           []*string{aws.String("1")},
	}).Return(&ec2.DescribeLaunchTemplateVersionsOutput{
		LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
			{
				LaunchTemplateData: &ec2.ResponseLaunchTemplateData{
					ImageId:              aws.String("123"),
					InstanceRequirements: instanceRequirements,
				},
			},
		},
	})

	e.On("DescribeImages", &ec2.DescribeImagesInput{
		ImageIds: []*string{aws.String("123")},
	}).Return(&ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			{
				Architecture:       aws.String("x86_64"),
				VirtualizationType: aws.String("xen"),
			},
		},
	})

	requirements, err := awsWrapper.getRequirementsRequestFromEC2(instanceRequirements)
	assert.NoError(t, err)
	e.On("GetInstanceTypesFromInstanceRequirementsPages",
		&ec2.GetInstanceTypesFromInstanceRequirementsInput{
			ArchitectureTypes:    []*string{aws.String("x86_64")},
			InstanceRequirements: requirements,
			VirtualizationTypes:  []*string{aws.String("xen")},
		},
		mock.AnythingOfType("func(*ec2.GetInstanceTypesFromInstanceRequirementsOutput, bool) bool"),
	).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(*ec2.GetInstanceTypesFromInstanceRequirementsOutput, bool) bool)
		fn(&ec2.GetInstanceTypesFromInstanceRequirementsOutput{
			InstanceTypes: []*ec2.InstanceTypeInfoFromInstanceRequirements{
				{
					InstanceType: aws.String("g4dn.xlarge"),
				},
			},
		}, false)
	}).Return(nil)

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
		expectedData         *ec2.ResponseLaunchTemplateData
		expectedErr          error
	}{
		{
			"no launch template version found",
			&ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{},
			},
			nil,
			errors.New("unable to find template versions for launch template launchTemplateName"),
		},
		{
			"no data found for launch template",
			&ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
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
				LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
					{
						LaunchTemplateName: aws.String("launchTemplateName"),
						LaunchTemplateData: &ec2.ResponseLaunchTemplateData{
							ImageId: aws.String("123"),
						},
					},
				},
			},
			&ec2.ResponseLaunchTemplateData{
				ImageId: aws.String("123"),
			},
			nil,
		},
	}

	describeTemplateInput := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String("launchTemplateName"),
		Versions:           []*string{aws.String("1")},
	}

	for _, testCase := range testCases {
		e.On("DescribeLaunchTemplateVersions", describeTemplateInput).Return(testCase.describeTemplateData).Once()

		describeData, err := awsWrapper.getLaunchTemplateData("launchTemplateName", "1")
		assert.Equal(t, testCase.expectedData, describeData)
		assert.Equal(t, testCase.expectedErr, err)
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
