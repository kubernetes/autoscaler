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

import "time"

// InstanceInfoDescriber is an interface to describe instance information
type InstanceInfoDescriber interface {
	// DescribeInstanceInfo returns the corresponding aws instance info by given instance type and region.
	DescribeInstanceInfo(instanceType string, region string) (*InstanceInfo, error)
}

// SpotPriceHistoryDescriber is an interface to describe spot price history information
type SpotPriceHistoryDescriber interface {
	// DescribeSpotPriceHistory returns the spot price history for given instance type
	DescribeSpotPriceHistory(instanceType string, availabilityZone string, startTime time.Time) (*SpotPriceHistory, error)
}

// LaunchConfigurationDescriber is an interface to describe aws ec2 launch configurations
type LaunchConfigurationDescriber interface {
	// DescribeLaunchConfiguration returns the corresponding launch configuration by the given launch configuration name.
	DescribeLaunchConfiguration(launchConfigurationName string) (*EC2LaunchConfiguration, error)
}

// AutoscalingGroupDescriber is an interface to describe aws ec2 autoscaling groups
type AutoscalingGroupDescriber interface {
	// DescribeAutoscalingGroup returns the corresponding EC2AutoscalingGroup by the given autoscaling group name
	DescribeAutoscalingGroup(autoscalingGroupName string) (*EC2AutoscalingGroup, error)
}
