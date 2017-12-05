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
	"errors"
	"testing"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/stretchr/testify/assert"
)

func TestAutoscalingService_DescribeAutoscalingGroup(t *testing.T) {
	type testCase struct {
		asName       string
		expectError  bool
		expectResult bool
	}
	type cases []testCase

	asName1 := "k8s-AutoscalingGroupWorker-TTTTTTTTTTTTT"
	asName2 := "k8s-AutoscalingGroupWorker-YYYYYYYYYYYYY"
	asName3 := "k8s-AutoscalingGroupWorker-XXXXXXXXXXXXX"
	lcName1 := "k8s-LaunchConfigurationWorker-TTTTTTTTTTTTT"
	lcName2 := "k8s-LaunchConfigurationWorker-YYYYYYYYYYYYY"
	azName1 := "us-east-1a"
	azName2 := "us-east-1b"

	service := NewEC2AutoscalingService(newFakeAutoscalingService(
		newAutoscalingMock(asName1, lcName1, azName1, azName2),
		newAutoscalingMock(asName2, lcName2, azName1, azName2),
	))

	tcs := cases{
		{ // good case: common case
			asName1,
			false,
			true,
		},
		{ // good case: common case
			asName2,
			false,
			true,
		},
		{ // error case: unknown autoscaling group
			asName3,
			true,
			false,
		},
	}

	for id, tc := range tcs {
		out, err := service.DescribeAutoscalingGroup(tc.asName)
		if tc.expectError {
			assert.Error(t, err, fmt.Sprintf("case %d", id))
			assert.Nil(t, out, fmt.Sprintf("case %d", id))
		} else {
			assert.NoError(t, err, fmt.Sprintf("case %d", id))
			assert.NotNil(t, out, fmt.Sprintf("case %d", id))
		}
		if tc.expectResult {
			assert.Equal(t, tc.asName, out.Name, fmt.Sprintf("case %d", id))
		}

	}
}

func newFakeAutoscalingService(ams ...autoscalingMock) *fakeAutoscalingService {
	m := make(map[string]*autoscaling.Group)

	for _, am := range ams {
		m[am.name] = am.asg
	}

	return &fakeAutoscalingService{m, []string{"token-a", "token-b"}}
}

func newAutoscalingMock(asName, lcName string, availabilityZones ...string) autoscalingMock {
	return autoscalingMock{
		asg: &autoscaling.Group{
			AvailabilityZones:       stringToStringSliceRef(availabilityZones...),
			LaunchConfigurationName: aws.String(lcName),
			AutoScalingGroupName:    aws.String(asName),
		},
		name: asName,
	}
}

type autoscalingMock struct {
	asg  *autoscaling.Group
	name string
}

type fakeAutoscalingService struct {
	mocks  map[string]*autoscaling.Group
	tokens []string
}

func (lcs *fakeAutoscalingService) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (output *autoscaling.DescribeAutoScalingGroupsOutput, err error) {
	output = new(autoscaling.DescribeAutoScalingGroupsOutput)

	output.NextToken, err = nextToken(lcs.tokens, input.NextToken)
	if err != nil {
		return
	}

	if output.NextToken != nil {
		return
	}

	output.AutoScalingGroups = make([]*autoscaling.Group, 0)
	for _, name := range input.AutoScalingGroupNames {
		if item, found := lcs.mocks[*name]; found {
			output.AutoScalingGroups = append(output.AutoScalingGroups, item)
		}
	}

	return
}

func nextToken(tokenChain []string, userToken *string) (*string, error) {
	tokenChainLen := len(tokenChain)
	if tokenChainLen == 0 {
		return nil, nil
	}

	if userToken == nil {
		return &tokenChain[0], nil
	}

	for i, token := range tokenChain {
		if *userToken == token {
			next := i + 1
			if next < tokenChainLen {
				return &tokenChain[next], nil
			}
			return nil, nil
		}
	}

	return nil, errors.New("invalid token")
}
