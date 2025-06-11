/*
Copyright 2019 The Kubernetes Authors.

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

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	ec2types "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestGetStaticEC2InstanceTypes(t *testing.T) {
	result, _ := GetStaticEC2InstanceTypes()
	assert.True(t, len(result) != 0)
}

func TestInstanceTypeTransform(t *testing.T) {
	rawInstanceType := ec2types.InstanceTypeInfo{
		InstanceType: ec2types.InstanceType("c4.xlarge"),
		ProcessorInfo: &ec2types.ProcessorInfo{
			SupportedArchitectures: []ec2types.ArchitectureType{ec2types.ArchitectureTypeX8664},
		},
		VCpuInfo: &ec2types.VCpuInfo{
			DefaultVCpus: aws.Int32(4),
		},
		MemoryInfo: &ec2types.MemoryInfo{
			SizeInMiB: aws.Int64(7680),
		},
	}

	instanceType := transformInstanceType(&rawInstanceType)

	assert.Equal(t, "c4.xlarge", instanceType.InstanceType)
	assert.Equal(t, int64(4), instanceType.VCPU)
	assert.Equal(t, int64(7680), instanceType.MemoryMb)
	assert.Equal(t, int64(0), instanceType.GPU)
	assert.Equal(t, "amd64", instanceType.Architecture)
}

func TestInterpretEc2SupportedArchitecure(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "arm64",
			expect: "arm64",
		},
		{
			input:  "i386",
			expect: "amd64",
		},
		{
			input:  "x86_64",
			expect: "amd64",
		},
		{
			input:  "x86_64_mac",
			expect: "amd64",
		},
		{
			input:  "anything default",
			expect: "amd64",
		},
	}

	for _, test := range tests {
		got := interpretEc2SupportedArchitecure(test.input)
		assert.Equal(t, test.expect, got)
	}
}

func TestGetGpuCount(t *testing.T) {
	gpuDeviceInfos := []ec2types.GpuDeviceInfo{
		{Count: aws.Int32(8)},
		{Count: aws.Int32(4)},
		{Count: aws.Int32(0)},
	}

	gpuInfo := ec2types.GpuInfo{Gpus: gpuDeviceInfos}

	assert.Equal(t, int32(12), getGpuCount(&gpuInfo))
}

func TestGetCurrentAwsRegionWithRegionEnv(t *testing.T) {
	region := "us-west-2"
	t.Setenv("AWS_REGION", region)

	result, err := GetCurrentAwsRegion()
	assert.Nil(t, err)
	assert.Equal(t, region, result)
}
