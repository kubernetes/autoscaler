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

type awsEC2LaunchConfigurationService interface {
	DescribeLaunchConfigurations(input *autoscaling.DescribeLaunchConfigurationsInput) (*autoscaling.DescribeLaunchConfigurationsOutput, error)
}

// EC2LaunchConfiguration holds AWS EC2 Launch Configuration information
type EC2LaunchConfiguration struct {
	// HasSpotMarkedBid is true if the launch configuration uses the spot marked
	HasSpotMarkedBid bool
	// SpotPrice is the bid price on the spot marked the autoscaling group uses
	SpotPrice float64
	// Name of the launch configuration
	Name string
	// InstanceType of the underlying instance described in the launch configuration
	InstanceType string
}

// NewEC2LaunchConfigurationService is the constructor of launchConfigurationService which is a wrapper for
// the AWS EC2 LaunchConfiguration API
func NewEC2LaunchConfigurationService(awsEC2Service awsEC2LaunchConfigurationService) *launchConfigurationService {
	return &launchConfigurationService{service: awsEC2Service}
}

type launchConfigurationService struct {
	service awsEC2LaunchConfigurationService
}

// DescribeLaunchConfiguration returns the corresponding launch configuration by the given launch configuration name.
func (lcs *launchConfigurationService) DescribeLaunchConfiguration(launchConfigurationName string) (*EC2LaunchConfiguration, error) {
	req := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{&launchConfigurationName},
	}

	for {
		res, err := lcs.service.DescribeLaunchConfigurations(req)
		if err != nil {
			return nil, err
		}

		for _, lc := range res.LaunchConfigurations {
			if *lc.LaunchConfigurationName == launchConfigurationName {
				p, err := stringRefToFloat64(lc.SpotPrice)
				if err != nil {
					return nil, fmt.Errorf("failed to parse price: %v", err)
				}
				return &EC2LaunchConfiguration{
					HasSpotMarkedBid: lc.SpotPrice != nil,
					SpotPrice:        p,
					Name:             *lc.LaunchConfigurationName,
					InstanceType:     *lc.InstanceType,
				}, nil
			}
		}

		req.NextToken = res.NextToken
		if req.NextToken == nil {
			break
		}
	}

	return nil, fmt.Errorf("launch configuration named %s not found", launchConfigurationName)
}
