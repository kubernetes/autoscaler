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

package ondemand

import (
	"errors"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
)

// Descriptor describes the price interface
type Descriptor interface {
	// Price returns the current instance price per hour in USD.
	Price(instanceType string, availabilityZones ...string) (float64, error)
}

// NewDescriptor is the constructor of the constructor
func NewDescriptor(service api.InstanceInfoDescriber) *descriptor {
	return &descriptor{
		service: service,
	}
}

type descriptor struct {
	service api.InstanceInfoDescriber
}

// Price returns the current instance price per hour in USD.
func (d *descriptor) Price(instanceType string, availabilityZones ...string) (float64, error) {
	if len(availabilityZones) == 0 {
		return 0, errors.New("no availability zone given")
	}
	region := regionOfAvailabilityZone(availabilityZones[0])
	info, err := d.service.DescribeInstanceInfo(instanceType, region)
	if err != nil {
		return 0, fmt.Errorf("failed to obtain instance info for %s in zone %s: %v", instanceType, region, err)
	}

	return info.OnDemandPrice, nil
}

func regionOfAvailabilityZone(availabilityZone string) string {
	return availabilityZone[0 : len(availabilityZone)-1]
}
