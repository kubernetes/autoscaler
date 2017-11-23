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

package price

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/price/ondemand"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/price/spot"
)

// ShapeDescriptor describes an interface to appraise an instance price of any shape
type ShapeDescriptor interface {
	// Price calls, depending whether the asg has a spot price or not, the spot or the on-demand price descriptor
	Price(asgName string) (price float64, err error)
}

type shapeDescriptor struct {
	autoscaling         api.AutoscalingGroupDescriber
	launchConfiguration api.LaunchConfigurationDescriber
	spot                spot.Descriptor
	onDemand            ondemand.Descriptor
}

// NewDescriptor is the constructor of a shapeDescriptor
func NewDescriptor(s *session.Session) *shapeDescriptor {
	as := autoscaling.New(s)
	return &shapeDescriptor{
		autoscaling:         api.NewEC2AutoscalingService(as),
		launchConfiguration: api.NewEC2LaunchConfigurationService(as),
		spot:                spot.NewDescriptor(api.NewEC2SpotPriceService(ec2.New(s))),
		onDemand:            ondemand.NewDescriptor(api.NewEC2InstanceInfoService(http.DefaultClient)),
	}
}

// Price calls, depending whether the asg has a spot price or not, the spot or the on-demand price descriptor
func (d *shapeDescriptor) Price(asgName string) (price float64, err error) {
	asg, err := d.autoscaling.DescribeAutoscalingGroup(asgName)
	if err != nil {
		return 0, err
	}

	lc, err := d.launchConfiguration.DescribeLaunchConfiguration(asg.LaunchConfigurationName)
	if err != nil {
		return 0, err
	}

	if lc.HasSpotMarkedBid {
		return d.spot.Price(lc.InstanceType, lc.SpotPrice, asg.AvailabilityZones...)
	}

	return d.onDemand.Price(lc.InstanceType, asg.AvailabilityZones...)
}
