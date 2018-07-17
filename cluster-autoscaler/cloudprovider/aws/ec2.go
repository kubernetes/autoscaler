/*
Copyright 2018 The Kubernetes Authors.

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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type ec2I interface {
	DescribeLaunchTemplateVersions(input *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error)
}

type ec2Wrapper struct {
	ec2I
}

func (m ec2Wrapper) getInstanceTypeByLT(name string, version string) (string, error) {
	params := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: aws.String(name),
		Versions:           []*string{aws.String(version)},
	}

	describeData, err := m.DescribeLaunchTemplateVersions(params)
	if err != nil {
		return "", err
	}

	if len(describeData.LaunchTemplateVersions) == 0 {
		return "", fmt.Errorf("Unable to find template versions")
	}

	lt := describeData.LaunchTemplateVersions[0]
	instanceType := lt.LaunchTemplateData.InstanceType

	if instanceType == nil {
		return "", fmt.Errorf("Unable to find instance type within launch template")
	}

	return aws.StringValue(instanceType), nil
}
