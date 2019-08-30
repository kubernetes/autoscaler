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

// AWSInstanceType describes the instance type (m4.2xlarge, ...)
type AWSInstanceType string

// AWSPriceDimension describes the
type AWSPriceDimension string

// AWSRegion describes the Region of the resources
type AWSRegion string

// AWSAvailabilityZone describes the region's Availability Zone
type AWSAvailabilityZone string

// AWSSpotRequestID describes the ID of the spot request
type AWSSpotRequestID string

// AWSSpotRequestState describes the state of the spot request
type AWSSpotRequestState string

// AWSSpotRequestStatus describes the status of the spot request
type AWSSpotRequestStatus string

// AWSIamInstanceProfile describes the IamInstanceProfile
type AWSIamInstanceProfile string

// AWSLaunchConfigurationName describes the Launch Configuration name
type AWSLaunchConfigurationName string

// AWSAsgName describes the AGS name
type AWSAsgName string

const (
	// AWSSpotRequestStateOpen indicates the spot request has not been yet assessed
	AWSSpotRequestStateOpen = AWSSpotRequestState("open")
	// AWSSpotRequestStateFailed indicates the spot request could not be fulfilled
	AWSSpotRequestStateFailed = AWSSpotRequestState("failed")
	// AWSSpotRequestStatusNotAvailable there is not enough capacity available for the instances that you requested
	AWSSpotRequestStatusNotAvailable = AWSSpotRequestStatus("capacity-not-available")
	// AWSSpotRequestStatusOversubscribed there is not enough capacity available for the instances that you requested
	AWSSpotRequestStatusOversubscribed = AWSSpotRequestStatus("capacity-oversubscribed")
	// AWSSpotRequestStatusPriceToLow the request can't be fulfilled yet because your maximum price is below the Spot price.
	// In this case, no instance is launched and your request remains open
	AWSSpotRequestStatusPriceToLow = AWSSpotRequestStatus("price-too-low")
	// AWSSpotRequestStatusNotFulfillable the Spot request can't be fulfilled because one or more constraints are not valid
	// (for example, the Availability Zone does not exist). The status message indicates which constraint is not valid
	AWSSpotRequestStatusNotFulfillable = AWSSpotRequestStatus("constraint-not-fulfillable")
)
