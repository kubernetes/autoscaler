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
	"errors"
	"fmt"
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/ec2metadata"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/session"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/service/ec2"
)

var (
	ec2MetaDataServiceUrl = "http://169.254.169.254"
)

// GenerateEC2InstanceTypes returns a map of ec2 resources
func GenerateEC2InstanceTypes(sess *session.Session) (map[string]*InstanceType, error) {
	ec2Client := ec2.New(sess)
	input := ec2.DescribeInstanceTypesInput{}
	instanceTypes := make(map[string]*InstanceType)

	if err := ec2Client.DescribeInstanceTypesPages(&input, func(page *ec2.DescribeInstanceTypesOutput, isLastPage bool) bool {
		for _, rawInstanceType := range page.InstanceTypes {
			instanceTypes[*rawInstanceType.InstanceType] = transformInstanceType(rawInstanceType)
		}
		return !isLastPage
	}); err != nil {
		return nil, err
	}

	if len(instanceTypes) == 0 {
		return nil, errors.New("unable to load EC2 Instance Type list")
	}

	return instanceTypes, nil
}

// GetStaticEC2InstanceTypes return pregenerated ec2 instance type list
func GetStaticEC2InstanceTypes() (map[string]*InstanceType, string) {
	return InstanceTypes, StaticListLastUpdateTime
}

func transformInstanceType(rawInstanceType *ec2.InstanceTypeInfo) *InstanceType {
	instanceType := &InstanceType{
		InstanceType: *rawInstanceType.InstanceType,
	}
	if rawInstanceType.MemoryInfo != nil && rawInstanceType.MemoryInfo.SizeInMiB != nil {
		instanceType.MemoryMb = *rawInstanceType.MemoryInfo.SizeInMiB
	}
	if rawInstanceType.VCpuInfo != nil && rawInstanceType.VCpuInfo.DefaultVCpus != nil {
		instanceType.VCPU = *rawInstanceType.VCpuInfo.DefaultVCpus
	}
	if rawInstanceType.GpuInfo != nil && len(rawInstanceType.GpuInfo.Gpus) > 0 {
		instanceType.GPU = getGpuCount(rawInstanceType.GpuInfo)
	}
	if rawInstanceType.ProcessorInfo != nil && len(rawInstanceType.ProcessorInfo.SupportedArchitectures) > 0 {
		instanceType.Architecture = interpretEc2SupportedArchitecure(*rawInstanceType.ProcessorInfo.SupportedArchitectures[0])
	}
	return instanceType
}

func getGpuCount(gpuInfo *ec2.GpuInfo) int64 {
	var gpuCountSum int64
	for _, gpu := range gpuInfo.Gpus {
		if gpu.Count != nil {
			gpuCountSum += *gpu.Count
		}
	}
	return gpuCountSum
}

func interpretEc2SupportedArchitecure(archName string) string {
	switch archName {
	case "arm64":
		return "arm64"
	case "i386":
		return "amd64"
	case "x86_64":
		return "amd64"
	case "x86_64_mac":
		return "amd64"
	default:
		return "amd64"
	}
}

// GetCurrentAwsRegion return region of current cluster without building awsManager
func GetCurrentAwsRegion() (string, error) {
	region, present := os.LookupEnv("AWS_REGION")

	if !present {
		c := aws.NewConfig().
			WithEndpoint(ec2MetaDataServiceUrl)
		sess, err := session.NewSession()
		if err != nil {
			return "", fmt.Errorf("failed to create session")
		}
		return ec2metadata.New(sess, c).Region()
	}

	return region, nil
}
