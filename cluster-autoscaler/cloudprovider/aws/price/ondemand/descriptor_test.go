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
	"fmt"

	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
)

func TestDescriptor_Price(t *testing.T) {
	d := NewDescriptor(
		newFakeInstanceInfoDescriber(
			buildInfo("m4.xlarge", "us-east-1", 4, 0, 16*1024, 0.111),
			buildInfo("m4.2xlarge", "us-east-1", 8, 0, 32*1024, 0.222),
		),
	)

	type testCase struct {
		instanceType      string
		availabilityZones []string
		expectError       bool
		expectPrice       float64
	}

	testCases := []testCase{
		{ // common case
			"m4.xlarge", []string{"us-east-1a", "us-east-1b"},
			false, 0.111,
		},
		{ // common case
			"m4.2xlarge", []string{"us-east-1a"},
			false, 0.222,
		},
		{ // error case: no availability zone
			"m4.xlarge", []string{},
			true, 0.0,
		},
		{ // error case: unknown instance type
			"m4.4xlarge", []string{"us-east-1a", "us-east-1b"},
			true, 0.0,
		},
	}

	for n, tc := range testCases {
		price, err := d.Price(tc.instanceType, tc.availabilityZones...)
		if tc.expectError {
			assert.Error(t, err, fmt.Sprintf("case %d", n))
		} else {
			assert.NoError(t, err, fmt.Sprintf("case %d", n))
		}
		assert.Equal(t, tc.expectPrice, price, fmt.Sprintf("case %d", n))
	}
}

type instanceInZone struct {
	instanceType string
	region       string
}

type instanceInfoBundle struct {
	instanceInZone instanceInZone
	info           *api.InstanceInfo
}

func buildInfo(instanceType, region string, cpu, gpu, mem int64, onDemandPrice float64) instanceInfoBundle {
	return instanceInfoBundle{
		instanceInZone{instanceType, region},
		&api.InstanceInfo{
			InstanceType:  instanceType,
			OnDemandPrice: onDemandPrice,
			VCPU:          cpu,
			GPU:           gpu,
			MemoryMb:      mem,
		},
	}
}

func newFakeInstanceInfoDescriber(iBundles ...instanceInfoBundle) *fakeInstanceInfoDescriber {
	c := make(map[instanceInZone]*api.InstanceInfo)

	for _, b := range iBundles {
		c[b.instanceInZone] = b.info
	}

	return &fakeInstanceInfoDescriber{
		c: c,
	}
}

type fakeInstanceInfoDescriber struct {
	c map[instanceInZone]*api.InstanceInfo
}

func (i *fakeInstanceInfoDescriber) DescribeInstanceInfo(instanceType string, region string) (*api.InstanceInfo, error) {
	iiz := instanceInZone{instanceType, region}
	if info, found := i.c[iiz]; found {
		return info, nil
	}

	return nil, fmt.Errorf("instance info not available for instance type %s in region %s", instanceType, region)
}
