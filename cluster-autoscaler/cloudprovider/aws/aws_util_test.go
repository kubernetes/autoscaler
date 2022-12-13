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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/ec2"
)

func TestGetStaticEC2InstanceTypes(t *testing.T) {
	result, _ := GetStaticEC2InstanceTypes()
	assert.True(t, len(result) != 0)
}

func TestInstanceTypeTransform(t *testing.T) {
	rawInstanceType := ec2.InstanceTypeInfo{
		InstanceType: aws.String("c4.xlarge"),
		ProcessorInfo: &ec2.ProcessorInfo{
			SupportedArchitectures: []*string{aws.String("x86_64")},
		},
		VCpuInfo: &ec2.VCpuInfo{
			DefaultVCpus: aws.Int64(4),
		},
		MemoryInfo: &ec2.MemoryInfo{
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
	gpuDeviceInfos := []*ec2.GpuDeviceInfo{
		{Count: aws.Int64(8)},
		{Count: aws.Int64(4)},
		{Count: aws.Int64(0)},
	}

	gpuInfo := ec2.GpuInfo{Gpus: gpuDeviceInfos}

	assert.Equal(t, int64(12), getGpuCount(&gpuInfo))
}

func TestGetCurrentAwsRegion(t *testing.T) {
	region := "us-west-2"
	if oldRegion, found := os.LookupEnv("AWS_REGION"); found {
		os.Unsetenv("AWS_REGION")
		defer os.Setenv("AWS_REGION", oldRegion)
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("{\"region\" : \"" + region + "\"}"))
	}))
	// Close the server when test finishes
	defer server.Close()

	ec2MetaDataServiceUrl = server.URL
	result, err := GetCurrentAwsRegion()

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, region, result)
}

func TestGetCurrentAwsRegionWithRegionEnv(t *testing.T) {
	region := "us-west-2"
	t.Setenv("AWS_REGION", region)

	result, err := GetCurrentAwsRegion()
	assert.Nil(t, err)
	assert.Equal(t, region, result)
}
