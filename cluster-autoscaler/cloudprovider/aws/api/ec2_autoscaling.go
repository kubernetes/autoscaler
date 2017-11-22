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

package api

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/autoscaling"
)

type awsEC2AutoscalingGroupService interface {
	DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

// EC2AutoscalingGroup holds AWS Autoscaling Group information
type EC2AutoscalingGroup struct {
	Name                    string
	LaunchConfigurationName string
	AvailabilityZones       []string
}

// NewEC2AutoscalingService is the constructor of autoscalingService which is a wrapper for the AWS EC2
// Autoscaling Group API
func NewEC2AutoscalingService(awsEC2Service awsEC2AutoscalingGroupService) *autoscalingService {
	return &autoscalingService{service: awsEC2Service}
}

type autoscalingService struct {
	service awsEC2AutoscalingGroupService
}

// DescribeAutoscalingGroup returns the corresponding EC2AutoscalingGroup by the given autoscaling group name
func (ass *autoscalingService) DescribeAutoscalingGroup(autoscalingGroupName string) (*EC2AutoscalingGroup, error) {
	req := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&autoscalingGroupName},
	}

	for {
		res, err := ass.service.DescribeAutoScalingGroups(req)
		if err != nil {
			return nil, err
		}

		for _, group := range res.AutoScalingGroups {
			if *group.AutoScalingGroupName == autoscalingGroupName {
				return &EC2AutoscalingGroup{
					Name: *group.AutoScalingGroupName,
					LaunchConfigurationName: *group.LaunchConfigurationName,
					AvailabilityZones:       stringRefToStringSlice(group.AvailabilityZones...),
				}, nil
			}
		}

		req.NextToken = res.NextToken
		if req.NextToken == nil {
			break
		}
	}

	return nil, fmt.Errorf("autoscaling group named %s not found", autoscalingGroupName)
}
