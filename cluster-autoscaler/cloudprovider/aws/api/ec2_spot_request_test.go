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
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestSpotRequestManager_listDelta(t *testing.T) {
	cases := []struct {
		name          string
		requests      []*ec2.SpotInstanceRequest
		expected      []*SpotRequest
		lastCheckTime time.Time
		expectedError string
		error         string
	}{
		{
			name:          "error fetching list: handle error",
			requests:      []*ec2.SpotInstanceRequest{},
			expected:      []*SpotRequest{},
			expectedError: "could not retrieve AWS Spot Request list: AWS died",
			error:         "AWS died",
		},
		{
			name:     "no request exists: returns empty list",
			requests: []*ec2.SpotInstanceRequest{},
			expected: []*SpotRequest{},
		},
		{
			name:          "no request is young enough: returns empty list",
			lastCheckTime: time.Now(),
			requests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("1", "open",
					"capacity-not-available", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Hour)),
			},
			expected: []*SpotRequest{},
		},
		{
			name:          "no request has the requested state: returns empty list",
			lastCheckTime: time.Now(),
			requests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("1", "fulfilled",
					"capacity-not-available", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Hour)),
			},
			expected: []*SpotRequest{},
		},
		{
			name:          "multiple requests found: returns filtered list",
			lastCheckTime: fluxCompensator(time.Minute * 10),
			requests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1c", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			expected: []*SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "failed",
					Status:           "bad-parameters",
				},
				{
					ID:               "15",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := newAwsEC2SpotRequestManagerMock(c.requests)
			mock.setError(c.error)

			service := NewEC2SpotRequestManager(mock)
			list, err := service.listDelta(c.lastCheckTime)

			if len(c.error) > 0 {
				assert.Nil(t, list, c.name, "request list should be nil")
				assert.NotNil(t, err, c.name, "awaits an error")

				if err != nil {
					assert.Equal(t, c.expectedError, err.Error(), c.name, "unexpected error")
				}
			} else {
				assert.Nil(t, err, c.name, "no error should have append")
				assert.NotNil(t, list, c.name, "awaits a list")

				if list != nil {
					assert.Equal(t, c.expected, list, c.name, "return list is not valid")
				}
			}
		})
	}
}

func TestSpotRequestManager_List(t *testing.T) {
	cases := []struct {
		name          string
		requests      []*ec2.SpotInstanceRequest
		expected      []*SpotRequest
		lastCheckTime time.Time
		expectedError string
		error         string
	}{
		{
			name:          "returns all found spot requests",
			lastCheckTime: fluxCompensator(time.Minute * 10),
			requests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1c", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			expected: []*SpotRequest{
				{
					ID:               "12",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "fulfilled",
					Status:           "active",
				},
				{
					ID:               "13",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1c",
					State:            "open",
					Status:           "active",
				},
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "failed",
					Status:           "bad-parameters",
				},
				{
					ID:               "15",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := newAwsEC2SpotRequestManagerMock(c.requests)
			mock.setError(c.error)

			service := NewEC2SpotRequestManager(mock)
			service.lastFetchTime = c.lastCheckTime
			list, err := service.List()

			if len(c.error) > 0 {
				assert.Nil(t, list, c.name, "request list should be nil")
				assert.NotNil(t, err, c.name, "awaits an error")

				if err != nil {
					assert.Equal(t, c.expectedError, err.Error(), c.name, "unexpected error")
				}
			} else {
				assert.Nil(t, err, c.name, "no error should have append")
				assert.NotNil(t, list, c.name, "awaits a list")

				if list != nil {
					assert.Equal(t, c.expected, list, c.name, "return list is not valid")
				}
			}
		})
	}
}

func TestSpotRequestManager_ListDelta(t *testing.T) {
	cases := []struct {
		name          string
		requests      []*ec2.SpotInstanceRequest
		expected      []*SpotRequest
		lastCheckTime time.Time
		expectedError string
		error         string
	}{
		{
			name:          "error fetching list: last fetch time should not be updated",
			requests:      []*ec2.SpotInstanceRequest{},
			expected:      []*SpotRequest{},
			expectedError: "could not retrieve AWS Spot Request list: AWS died",
			error:         "AWS died",
		},
		{
			name:          "returns a delta of all new/updated spot requests since the last fetch date",
			lastCheckTime: fluxCompensator(time.Minute * 10),
			requests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1c", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			expected: []*SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "failed",
					Status:           "bad-parameters",
				},
				{
					ID:               "15",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := newAwsEC2SpotRequestManagerMock(c.requests)
			mock.setError(c.error)

			service := NewEC2SpotRequestManager(mock)
			service.lastFetchTime = c.lastCheckTime
			list, err := service.ListDelta()

			if len(c.error) > 0 {
				assert.Nil(t, list, c.name, "request list should be nil")
				assert.NotNil(t, err, c.name, "awaits an error")

				assert.Equal(t, c.lastCheckTime, service.lastFetchTime, "on error lastFetchTime should not be updated")

				if err != nil {
					assert.Equal(t, c.expectedError, err.Error(), c.name, "unexpected error")
				}
			} else {
				assert.Nil(t, err, c.name, "no error should have append")
				assert.NotNil(t, list, c.name, "awaits a list")

				assert.Equal(t, true, service.lastFetchTime.After(c.lastCheckTime), "on successful fetch lastFetchTime should be updated")

				if list != nil {
					assert.Equal(t, c.expected, list, c.name, "return list is not valid")
				}
			}
		})
	}
}

func TestSpotRequestManager_CancelRequests(t *testing.T) {
	cases := []struct {
		name          string
		awsRequests   []*ec2.SpotInstanceRequest
		requests      []*SpotRequest
		expectedError string
		error         string
	}{
		{
			name:          "error cancelling requests: handle error",
			awsRequests:   []*ec2.SpotInstanceRequest{},
			requests:      []*SpotRequest{},
			expectedError: "could not cancel spot requests: AWS died",
			error:         "AWS died",
		},
		{
			// used CLI to test API: aws ec2 --region=eu-west-1 cancel-spot-instance-requests --spot-instance-request-ids ""
			name:          "empty request list provided: return error",
			awsRequests:   []*ec2.SpotInstanceRequest{},
			requests:      []*SpotRequest{},
			expectedError: "could not cancel spot requests: the request must contain the parameter SpotInstanceRequestId",
			error:         "the request must contain the parameter SpotInstanceRequestId",
		},
		{
			// used CLI to test API: aws ec2 --region=eu-west-1 cancel-spot-instance-requests --spot-instance-request-ids "sir-n29rnope"
			name:          "unknown spot request id: return error",
			awsRequests:   []*ec2.SpotInstanceRequest{},
			requests:      []*SpotRequest{},
			expectedError: "could not cancel spot requests: the spot instance request ID 'sir-n29rnope' does not exist",
			error:         "the spot instance request ID 'sir-n29rnope' does not exist",
		},
		{
			// used CLI to test API: aws ec2 --region=eu-west-1 cancel-spot-instance-requests --spot-instance-request-ids "sir-n29rnope" "<some-known-id>"
			name: "mixed known and unknown request ids: returns error",
			awsRequests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			requests: []*SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "failed",
					Status:           "bad-parameters",
				},
				{
					ID:               "unknown",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
			expectedError: "could not cancel spot requests: the spot instance request ID 'sir-n29rnope' does not exist",
			error:         "the spot instance request ID 'sir-n29rnope' does not exist",
		},
		{
			// used CLI to test API: aws ec2 --region=eu-west-1 cancel-spot-instance-requests --spot-instance-request-ids "<some-known-id>"
			name: "only known spot request ids provided: no error",
			awsRequests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			requests: []*SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "failed",
					Status:           "bad-parameters",
				},
				{
					ID:               "15",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := newAwsEC2SpotRequestManagerMock(c.awsRequests)
			mock.setError(c.error)

			service := NewEC2SpotRequestManager(mock)
			err := service.CancelRequests(c.requests)

			if len(c.error) > 0 {
				assert.NotNil(t, err, c.name, "awaits an error")

				if err != nil {
					assert.Equal(t, c.expectedError, err.Error(), c.name, "unexpected error")
				}
			} else {
				assert.Nil(t, err, c.name, "no error should have append")
			}
		})
	}
}

var _ AwsEC2SpotRequestManager = &awsEC2SpotRequestManagerMock{}

func newAwsEC2SpotRequestManagerMock(requests []*ec2.SpotInstanceRequest) *awsEC2SpotRequestManagerMock {
	return &awsEC2SpotRequestManagerMock{requests, ""}
}

type awsEC2SpotRequestManagerMock struct {
	requests []*ec2.SpotInstanceRequest
	error    string
}

func (m *awsEC2SpotRequestManagerMock) setError(errorMessage string) {
	m.error = errorMessage
}

func (m *awsEC2SpotRequestManagerMock) CancelSpotInstanceRequests(input *ec2.CancelSpotInstanceRequestsInput) (*ec2.CancelSpotInstanceRequestsOutput, error) {
	if len(m.error) > 0 {
		return nil, errors.New(m.error)
	}

	canceledIds := make([]*ec2.CancelledSpotInstanceRequest, len(m.requests))

idloop:
	for _, id := range input.SpotInstanceRequestIds {
		for _, request := range m.requests {
			if aws.StringValue(id) == aws.StringValue(request.SpotInstanceRequestId) {
				canceledIds = append(canceledIds, &ec2.CancelledSpotInstanceRequest{
					SpotInstanceRequestId: request.SpotInstanceRequestId,
					State:                 request.State,
				})
				request.State = aws.String("cancelled")
				continue idloop
			}
		}

		return nil, fmt.Errorf("the spot instance request ID '%s' does not exist", aws.StringValue(id))
	}

	return &ec2.CancelSpotInstanceRequestsOutput{CancelledSpotInstanceRequests: canceledIds}, nil
}

func (m *awsEC2SpotRequestManagerMock) DescribeSpotInstanceRequests(input *ec2.DescribeSpotInstanceRequestsInput) (*ec2.DescribeSpotInstanceRequestsOutput, error) {
	if len(m.error) > 0 {
		return nil, errors.New(m.error)
	}

	return &ec2.DescribeSpotInstanceRequestsOutput{SpotInstanceRequests: m.requests}, nil
}

func newSpotInstanceRequestInstance(id, state, status, iamInstanceProfile, instanceType, availabilityZone string, created *time.Time) *ec2.SpotInstanceRequest {
	if created == nil {
		created = aws.Time(time.Now())
	}

	return &ec2.SpotInstanceRequest{
		SpotInstanceRequestId:    aws.String(id),
		LaunchedAvailabilityZone: aws.String(availabilityZone),
		LaunchSpecification: &ec2.LaunchSpecification{
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
				Name: aws.String(iamInstanceProfile),
			},
			InstanceType: aws.String(instanceType),
		},
		State: aws.String(state),
		Status: &ec2.SpotInstanceStatus{
			Code:       aws.String(status),
			UpdateTime: created,
		},
		CreateTime: created,
	}
}

func fluxCompensatorAWS(travelRange time.Duration) *time.Time {
	past := fluxCompensator(travelRange)
	return &past
}

func fluxCompensator(travelRange time.Duration) time.Time {
	return time.Now().Add(-1 * travelRange)
}
