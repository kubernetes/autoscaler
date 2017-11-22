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
	"testing"

	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/stretchr/testify/assert"
)

func TestLaunchConfigurationService_DescribeLaunchConfiguration(t *testing.T) {
	type testCase struct {
		lcName      string
		service     awsEC2LaunchConfigurationService
		expectError bool
	}
	type cases []testCase

	var (
		lcName1 = "k8s-LaunchConfigurationWorker-TTTTTTTTTTTTT"
		lcName2 = "k8s-LaunchConfigurationWorker-YYYYYYYYYYYYY"
		lcName3 = "k8s-LaunchConfigurationWorker-XXXXXXXXXXXXX"
	)

	//NewEC2LaunchConfigurationService()

	tcs := cases{
		{ // good case: common case
			lcName1,
			newLCFakeService(lcName1, "m3.xlarge", stringSlice("token-a"), nil),
			false,
		},
		{ // good case: common case
			lcName2,
			newLCFakeService(lcName2, "m3.xlarge", stringSlice("token-a"), nil),
			false,
		},
		{ // good case: common case
			lcName3,
			newLCFakeService(lcName2, "m3.xlarge", stringSlice("token-a"), nil),
			true,
		},
	}

	for id, tc := range tcs {
		service := NewEC2LaunchConfigurationService(tc.service)
		out, err := service.DescribeLaunchConfiguration(tc.lcName)
		if tc.expectError {
			assert.Error(t, err, fmt.Sprintf("case %d", id))
			assert.Nil(t, out, fmt.Sprintf("case %d", id))
		} else {
			assert.NoError(t, err, fmt.Sprintf("case %d", id))
			assert.NotNil(t, out, fmt.Sprintf("case %d", id))
		}

	}
}

func newLCFakeService(
	name string,
	instanceType string,
	tokens []string,
	err error,
) *fakeLCService {
	return &fakeLCService{
		mocks: map[string]*autoscaling.LaunchConfiguration{
			name: {
				LaunchConfigurationName: aws.String(name),
				LaunchConfigurationARN:  aws.String(fmt.Sprintf("arn:aws:ec2:launchconfiguration:123456789:%s", name)),
				InstanceType:            aws.String(instanceType),
			},
		},
		err:    err,
		tokens: tokens,
	}
}

type fakeLCService struct {
	mocks  map[string]*autoscaling.LaunchConfiguration
	err    error
	tokens []string
}

func (lcs *fakeLCService) DescribeLaunchConfigurations(input *autoscaling.DescribeLaunchConfigurationsInput) (output *autoscaling.DescribeLaunchConfigurationsOutput, err error) {
	if lcs.err != nil {
		return nil, err
	}

	output = new(autoscaling.DescribeLaunchConfigurationsOutput)

	if len(lcs.tokens) != 0 {
		if input.NextToken == nil {
			output.NextToken = &lcs.tokens[0]
			return
		}

		for i, token := range lcs.tokens {
			if *input.NextToken == token {
				next := i + 1
				if next < len(lcs.tokens) {
					nextToken := lcs.tokens[next]
					output.NextToken = &nextToken
				} else {
					goto respond
				}
				return
			}
		}

		return nil, errors.New("invalid token")
	}

respond:
	output.LaunchConfigurations = make([]*autoscaling.LaunchConfiguration, 0)
	for _, name := range input.LaunchConfigurationNames {
		if item, found := lcs.mocks[*name]; found {
			output.LaunchConfigurations = append(output.LaunchConfigurations, item)
		}
	}

	return
}

func stringSlice(s ...string) []string {
	return s
}
