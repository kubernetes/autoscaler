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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

const (
	// InputStateFilter request state filter key for spot request listing
	InputStateFilter = "state"
)

// AwsEC2SpotRequestManager wraps the necessary AWS API methods
type AwsEC2SpotRequestManager interface {
	CancelSpotInstanceRequests(input *ec2.CancelSpotInstanceRequestsInput) (*ec2.CancelSpotInstanceRequestsOutput, error)
	DescribeSpotInstanceRequests(input *ec2.DescribeSpotInstanceRequestsInput) (*ec2.DescribeSpotInstanceRequestsOutput, error)
}

// SpotRequest provides all information necessary to assess spot ASG availability
type SpotRequest struct {
	ID               AWSSpotRequestID
	InstanceProfile  AWSIamInstanceProfile
	AvailabilityZone AWSAvailabilityZone
	InstanceType     AWSInstanceType
	State            AWSSpotRequestState
	Status           AWSSpotRequestStatus
}

// SpotRequestManager defines the interface to interact with spot requests
type SpotRequestManager interface {
	List() ([]*SpotRequest, error)
	ListDelta() ([]*SpotRequest, error)
	CancelRequests([]*SpotRequest) error
}

var _ SpotRequestManager = &spotRequestService{}

// NewEC2SpotRequestManager is the constructor of spotRequestService which is a wrapper for the AWS EC2 API
func NewEC2SpotRequestManager(awsEC2Service AwsEC2SpotRequestManager) *spotRequestService {
	return &spotRequestService{
		service:       awsEC2Service,
		lastFetchTime: time.Time{},
	}
}

type spotRequestService struct {
	service       AwsEC2SpotRequestManager
	lastFetchTime time.Time
}

// CancelRequests cancels all open spot requests from the provided list
func (srs *spotRequestService) CancelRequests(requests []*SpotRequest) error {
	ids := make([]*string, 0)

	for _, request := range requests {
		if request.State == AWSSpotRequestStateOpen {
			ids = append(ids, aws.String(string(request.ID)))
		}
	}

	options := &ec2.CancelSpotInstanceRequestsInput{
		SpotInstanceRequestIds: ids,
	}

	_, err := srs.service.CancelSpotInstanceRequests(options)
	if err != nil {
		return errors.Wrap(err, "could not cancel spot requests")
	}
	klog.V(3).Infof("canceled %d spot requests from AWS: %v", len(ids), ids)

	return nil
}

// Lists returns all aws spot requests
func (srs *spotRequestService) List() ([]*SpotRequest, error) {
	requests, err := srs.listDelta(time.Time{})

	if err == nil {
		klog.V(3).Infof("retrieved %d spot requests from AWS", len(requests))
	}

	return requests, err
}

// ListDelta returns all new or updated aws spot requests since the last listing
func (srs *spotRequestService) ListDelta() ([]*SpotRequest, error) {
	requests, err := srs.listDelta(srs.lastFetchTime)

	if err == nil {
		klog.V(3).Infof("retrieved %d new or updated spot requests from AWS", len(requests))
		srs.lastFetchTime = time.Now()
	}

	return requests, err
}

// listDelta returns spot requests newer as a provided point in time
func (srs *spotRequestService) listDelta(filterTime time.Time) ([]*SpotRequest, error) {
	list := make([]*SpotRequest, 0)

	arguments := srs.listArguments()

	awsSpotRequests, err := srs.service.DescribeSpotInstanceRequests(arguments)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve AWS Spot Request list")
	}

	klog.V(2).Infof("filter %d spot requests using provided time: %v",
		len(awsSpotRequests.SpotInstanceRequests), filterTime)

	for _, request := range awsSpotRequests.SpotInstanceRequests {
		if aws.TimeValue(request.Status.UpdateTime).After(filterTime) {
			converted := srs.convertAwsSpotRequest(request)
			list = append(list, converted)
		}
	}

	return list, nil
}

func (srs *spotRequestService) convertAwsSpotRequest(request *ec2.SpotInstanceRequest) *SpotRequest {
	converted := new(SpotRequest)

	converted.ID = AWSSpotRequestID(aws.StringValue(request.SpotInstanceRequestId))
	converted.AvailabilityZone = AWSAvailabilityZone(aws.StringValue(request.LaunchedAvailabilityZone))
	converted.InstanceProfile = AWSIamInstanceProfile(aws.StringValue(request.LaunchSpecification.IamInstanceProfile.Name))
	converted.InstanceType = AWSInstanceType(aws.StringValue(request.LaunchSpecification.InstanceType))
	converted.State = AWSSpotRequestState(aws.StringValue(request.State))
	converted.Status = AWSSpotRequestStatus(aws.StringValue(request.Status.Code))

	return converted
}

func (srs *spotRequestService) listArguments() *ec2.DescribeSpotInstanceRequestsInput {
	return &ec2.DescribeSpotInstanceRequestsInput{}
}
