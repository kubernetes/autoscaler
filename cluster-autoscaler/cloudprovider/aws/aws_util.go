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
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var (
	ec2MetaDataServiceUrl = "http://169.254.169.254"
)

// GenerateEC2InstanceTypes returns a map of ec2 resources
func GenerateEC2InstanceTypes(cfg aws.Config) (map[string]*InstanceType, error) {
	ctx := context.Background()
	ec2Client := ec2.NewFromConfig(cfg)
	input := ec2.DescribeInstanceTypesInput{}
	instanceTypes := make(map[string]*InstanceType)

	// Use pagination
	var nextToken *string
	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}
		page, err := ec2Client.DescribeInstanceTypes(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, rawInstanceType := range page.InstanceTypes {
			instanceTypes[string(rawInstanceType.InstanceType)] = transformInstanceType(&rawInstanceType)
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
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

func transformInstanceType(rawInstanceType *ec2types.InstanceTypeInfo) *InstanceType {
	instanceType := &InstanceType{
		InstanceType: string(rawInstanceType.InstanceType),
	}
	if rawInstanceType.MemoryInfo != nil && rawInstanceType.MemoryInfo.SizeInMiB != nil {
		instanceType.MemoryMb = *rawInstanceType.MemoryInfo.SizeInMiB
	}
	if rawInstanceType.VCpuInfo != nil && rawInstanceType.VCpuInfo.DefaultVCpus != nil {
		instanceType.VCPU = int64(*rawInstanceType.VCpuInfo.DefaultVCpus)
	}
	if rawInstanceType.GpuInfo != nil && len(rawInstanceType.GpuInfo.Gpus) > 0 {
		instanceType.GPU = getGpuCount(rawInstanceType.GpuInfo)
	}
	if rawInstanceType.ProcessorInfo != nil && len(rawInstanceType.ProcessorInfo.SupportedArchitectures) > 0 {
		instanceType.Architecture = interpretEc2SupportedArchitecure(string(rawInstanceType.ProcessorInfo.SupportedArchitectures[0]))
	}
	return instanceType
}

func getGpuCount(gpuInfo *ec2types.GpuInfo) int64 {
	var gpuCountSum int64
	for _, gpu := range gpuInfo.Gpus {
		if gpu.Count != nil {
			gpuCountSum += int64(*gpu.Count)
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
		ctx := context.Background()
		cfg, err := config.LoadDefaultConfig(ctx, config.WithEC2IMDSEndpoint(ec2MetaDataServiceUrl))
		if err != nil {
			return "", fmt.Errorf("failed to load config: %w", err)
		}
		client := imds.NewFromConfig(cfg)
		result, err := client.GetRegion(ctx, &imds.GetRegionInput{})
		if err != nil {
			return "", fmt.Errorf("failed to get region from metadata: %w", err)
		}
		return result.Region, nil
	}

	return region, nil
}
