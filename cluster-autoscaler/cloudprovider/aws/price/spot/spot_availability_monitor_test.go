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

package spot

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/api"
)

func TestAsgStatusCache_get(t *testing.T) {
	cases := []struct {
		name     string
		cache    map[api.AWSAsgName]*asgSpotStatus
		searched api.AWSAsgName
		expected *asgSpotStatus
	}{
		{
			name:     "getting unknown ASG name: return nil",
			cache:    map[api.AWSAsgName]*asgSpotStatus{},
			searched: "foo",
			expected: nil,
		},
		{
			name: "getting unknown ASG name: return nil",
			cache: map[api.AWSAsgName]*asgSpotStatus{
				"foo": {},
			},
			searched: "foo",
			expected: &asgSpotStatus{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := asgStatusCache{
				asgNames: make([]api.AWSAsgName, 0),
				cache:    c.cache,
				mux:      sync.RWMutex{},
			}

			asgStatus := cache.get(c.searched)

			assert.Equal(t, c.expected, asgStatus, c.name, "unexpected status returned")
		})
	}
}

func TestAsgStatusCache_exists(t *testing.T) {
	cases := []struct {
		name     string
		cache    map[api.AWSAsgName]*asgSpotStatus
		searched api.AWSAsgName
		expected bool
	}{
		{
			name:     "getting unknown ASG name: return nil",
			cache:    map[api.AWSAsgName]*asgSpotStatus{},
			searched: "foo",
			expected: false,
		},
		{
			name: "getting unknown ASG name: return nil",
			cache: map[api.AWSAsgName]*asgSpotStatus{
				"foo": {},
			},
			searched: "foo",
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := asgStatusCache{
				asgNames: make([]api.AWSAsgName, 0),
				cache:    c.cache,
				mux:      sync.RWMutex{},
			}

			asgStatus := cache.exists(c.searched)

			assert.Equal(t, c.expected, asgStatus, c.name, "unexpected status returned")
		})
	}
}

func TestAsgStatusCache_add(t *testing.T) {
	cases := []struct {
		name     string
		cache    asgStatusCache
		asgName  api.AWSAsgName
		status   *asgSpotStatus
		expected asgStatusCache
	}{
		{
			name: "adding new status: status added",
			cache: asgStatusCache{
				asgNames: []api.AWSAsgName{},
				cache:    map[api.AWSAsgName]*asgSpotStatus{},
				mux:      sync.RWMutex{},
			},
			asgName: "foo",
			status:  &asgSpotStatus{},
			expected: asgStatusCache{
				asgNames: []api.AWSAsgName{"foo"},
				cache: map[api.AWSAsgName]*asgSpotStatus{
					"foo": {},
				},
				mux: sync.RWMutex{},
			},
		},
		{
			name: "added existing ASG status: no change",
			cache: asgStatusCache{
				asgNames: []api.AWSAsgName{"foo"},
				cache: map[api.AWSAsgName]*asgSpotStatus{
					"foo": {},
				},
				mux: sync.RWMutex{},
			},
			asgName: "foo",
			status: &asgSpotStatus{
				Available: true,
			},
			expected: asgStatusCache{
				asgNames: []api.AWSAsgName{"foo"},
				cache: map[api.AWSAsgName]*asgSpotStatus{
					"foo": {},
				},
				mux: sync.RWMutex{},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := c.cache

			cache.add(c.asgName, c.status)

			assert.Equal(t, c.expected, cache, c.name, "unexpected status returned")
		})
	}
}

func TestAsgStatusCache_Update(t *testing.T) {
	cases := []struct {
		name     string
		cache    asgStatusCache
		asgName  api.AWSAsgName
		status   bool
		expected []api.AWSAsgName
		changed  bool
	}{
		{
			name: "updating non existing ASG status: no change",
			cache: asgStatusCache{
				asgNames: []api.AWSAsgName{"foo"},
				cache: map[api.AWSAsgName]*asgSpotStatus{
					"foo": {},
				},
				mux: sync.RWMutex{},
			},
			asgName:  "bar",
			status:   true,
			expected: []api.AWSAsgName{"foo"},
			changed:  false,
		},
		{
			name: "updating existing ASG status: change stored",
			cache: asgStatusCache{
				asgNames: []api.AWSAsgName{"foo"},
				cache: map[api.AWSAsgName]*asgSpotStatus{
					"foo": {
						statusChangeTime: fluxCompensator(time.Minute),
					},
				},
				mux: sync.RWMutex{},
			},
			asgName:  "foo",
			status:   true,
			expected: []api.AWSAsgName{"foo"},
			changed:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := c.cache
			oldChangeTime := time.Now()

			if c.changed {
				oldChangeTime = cache.get(c.asgName).statusChangeTime

			}

			cache.update(c.asgName, c.status)

			entry := cache.get(c.asgName)

			if c.changed {
				assert.Equal(t, c.expected, cache.asgNames,
					c.name, "unexpected entry in ASG name list")
				assert.Equal(t, c.status, entry.Available,
					c.name, "status should have been updated")

				assert.Equal(t, true, entry.statusChangeTime.After(oldChangeTime),
					c.name, "status change time should have been updated")
			} else {
				assert.Nil(t, entry,
					c.name, "unexpected entry in ASG name list")
			}
		})
	}
}

func TestAsgStatusCache_asgNameList(t *testing.T) {
	cases := []struct {
		name     string
		cache    []*asgSpotStatus
		expected []api.AWSAsgName
	}{
		{
			name:     "no ASG status cached: returns empty list",
			cache:    []*asgSpotStatus{},
			expected: []api.AWSAsgName{},
		},
		{
			name: "updating existing ASG status: change stored",
			cache: []*asgSpotStatus{
				{
					AsgName:          "foo",
					statusChangeTime: fluxCompensator(time.Minute * 10),
				},
				{
					AsgName:          "bar",
					statusChangeTime: fluxCompensator(time.Minute * 5),
				},
				{
					AsgName:          "foobar",
					statusChangeTime: fluxCompensator(time.Minute),
				},
			},
			expected: []api.AWSAsgName{"foo", "bar", "foobar"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := asgStatusCache{
				asgNames: []api.AWSAsgName{},
				cache:    map[api.AWSAsgName]*asgSpotStatus{},
				mux:      sync.RWMutex{},
			}

			for _, entry := range c.cache {
				cache.add(entry.AsgName, entry)
			}

			assert.Equal(t, c.expected, cache.asgNameList(),
				c.name, "unexpected entry in ASG name list")
		})
	}
}

func TestSpotRequestCache_refresh(t *testing.T) {
	cacheTime := fluxCompensator(time.Hour)

	cases := []struct {
		name     string
		cache    []*api.SpotRequest
		requests []*api.SpotRequest
		expected []*api.SpotRequest
		changed  bool
	}{
		{
			name:     "empty request list provided: no change to cache state",
			cache:    []*api.SpotRequest{},
			requests: []*api.SpotRequest{},
			expected: []*api.SpotRequest{},
			changed:  false,
		},
		{
			name:  "filled request list provided to empty cache: requests have been stored",
			cache: []*api.SpotRequest{},
			requests: []*api.SpotRequest{
				{
					ID: "12",
				},
				{
					ID: "15",
				},
			},
			expected: []*api.SpotRequest{
				{
					ID: "12",
				},
				{
					ID: "15",
				},
			},
			changed: true,
		},
		{
			name: "filled request list provided to filled cache: cached requests have been replaced",
			cache: []*api.SpotRequest{
				{
					ID: "1",
				},
			},
			requests: []*api.SpotRequest{
				{
					ID: "12",
				},
				{
					ID: "15",
				},
			},
			expected: []*api.SpotRequest{
				{
					ID: "12",
				},
				{
					ID: "15",
				},
			},
			changed: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := &spotRequestCache{
				createTime: cacheTime,
				cache:      c.cache,
				mux:        sync.RWMutex{},
			}

			oldEntryCount := len(cache.cache)

			cache.refresh(c.requests)

			assert.Equal(t, c.expected, cache.cache,
				c.name, "cache has not the wanted state")

			if c.changed {
				assert.NotEqual(t, oldEntryCount, len(cache.cache),
					c.name, "cache entry count should have changed")
				assert.Equal(t, true, cache.createTime.After(cacheTime),
					c.name, "cache change time should have been updated")
			} else {
				assert.Equal(t, oldEntryCount, len(cache.cache),
					c.name, "cache entry count should not have changed")
				assert.Equal(t, cacheTime, cache.createTime,
					c.name, "cache change time should not have been updated")
			}
		})
	}
}

func TestSpotRequestCache_findRequests(t *testing.T) {
	cacheTime := fluxCompensator(time.Hour)

	cases := []struct {
		name               string
		cache              []*api.SpotRequest
		availabilityZone   api.AWSAvailabilityZone
		iamInstanceProfile api.AWSIamInstanceProfile
		instanceType       api.AWSInstanceType
		expected           []*api.SpotRequest
	}{
		{
			name:     "search in empty cache: empty request list returned",
			cache:    []*api.SpotRequest{},
			expected: []*api.SpotRequest{},
		},
		{
			name: "no matching entry in cache: empty request list returned",
			cache: []*api.SpotRequest{
				{
					ID:               "5",
					AvailabilityZone: "eu-west-1c",
					InstanceProfile:  "123",
					InstanceType:     "c5.4xlarge",
				},
				{
					ID:               "34",
					AvailabilityZone: "eu-west-1c",
					InstanceProfile:  "222",
					InstanceType:     "c5.4xlarge",
				},
			},
			availabilityZone:   "eu-west-1c",
			iamInstanceProfile: "222",
			instanceType:       "m4.4xlarge",
			expected:           []*api.SpotRequest{},
		},
		{
			name: "matching entries in cache: matching request list returned",
			cache: []*api.SpotRequest{
				{
					ID:               "5",
					AvailabilityZone: "eu-west-1c",
					InstanceProfile:  "222",
					InstanceType:     "m4.4xlarge",
				},
				{
					ID:               "12",
					AvailabilityZone: "eu-west-1c",
					InstanceProfile:  "222",
					InstanceType:     "c5.4xlarge",
				},
				{
					ID:               "45",
					AvailabilityZone: "eu-west-1a",
					InstanceProfile:  "222",
					InstanceType:     "m4.4xlarge",
				},
				{
					ID:               "67",
					AvailabilityZone: "eu-west-1c",
					InstanceProfile:  "222",
					InstanceType:     "m4.4xlarge",
				},
			},
			availabilityZone:   "eu-west-1c",
			iamInstanceProfile: "222",
			instanceType:       "m4.4xlarge",
			expected: []*api.SpotRequest{
				{
					ID:               "5",
					AvailabilityZone: "eu-west-1c",
					InstanceProfile:  "222",
					InstanceType:     "m4.4xlarge",
				},
				{
					ID:               "67",
					AvailabilityZone: "eu-west-1c",
					InstanceProfile:  "222",
					InstanceType:     "m4.4xlarge",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := &spotRequestCache{
				createTime: cacheTime,
				cache:      c.cache,
				mux:        sync.RWMutex{},
			}

			requests := cache.findRequests(c.availabilityZone, c.iamInstanceProfile, c.instanceType)

			assert.Equal(t, c.expected, requests,
				c.name, "unexpected request list returned")
		})
	}
}

func TestSpotAvailabilityMonitor_requestsAllValid(t *testing.T) {
	cases := []struct {
		name        string
		awsRequests []*api.SpotRequest
		expected    bool
	}{
		{
			name: fmt.Sprintf("invalid state %s: returns false", api.AWSSpotRequestStateFailed),
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            api.AWSSpotRequestStateFailed,
					Status:           "something",
				},
			},
			expected: false,
		},
		{
			name: fmt.Sprintf("invalid status %s: returns false", api.AWSSpotRequestStatusNotAvailable),
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           api.AWSSpotRequestStatusNotAvailable,
				},
			},
			expected: false,
		},
		{
			name: fmt.Sprintf("invalid status %s: returns false", api.AWSSpotRequestStatusNotFulfillable),
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           api.AWSSpotRequestStatusNotFulfillable,
				},
			},
			expected: false,
		},
		{
			name: fmt.Sprintf("invalid status %s: returns false", api.AWSSpotRequestStatusOversubscribed),
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           api.AWSSpotRequestStatusOversubscribed,
				},
			},
			expected: false,
		},
		{
			name: fmt.Sprintf("invalid status %s: returns false", api.AWSSpotRequestStatusPriceToLow),
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           api.AWSSpotRequestStatusPriceToLow,
				},
			},
			expected: false,
		},
		{
			name: "invalid status in request list: returns false",
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "failed",
					Status:           "bad-parameters",
				},
				{
					ID:               "45",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
			expected: false,
		},
		{
			name: "only valid status in request list: returns true",
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "fulfilled",
					Status:           "active",
				},
				{
					ID:               "45",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
			expected: true,
		},
		{
			name:        "empty request list provided: return true",
			awsRequests: []*api.SpotRequest{},
			expected:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := newAwsEC2SpotRequestManagerMock([]*ec2.SpotInstanceRequest{})

			monitor := NewSpotAvailabilityMonitor(mock, time.Minute, time.Hour)

			status := monitor.requestsAllValid(c.awsRequests)

			assert.Equal(t, c.expected, status, c.name, "returned status as not the expected value")
		})
	}
}

func TestSpotAvailabilityMonitor_AsgAvailability(t *testing.T) {
	statusCreateTime := fluxCompensator(time.Hour)
	cases := []struct {
		name               string
		asgName            string
		availabilityZone   string
		iamInstanceProfile string
		instanceType       string
		awsRequests        []*api.SpotRequest
		awsStatusCache     map[api.AWSAsgName]*asgSpotStatus
		expectedCache      map[api.AWSAsgName]*asgSpotStatus
		expectedStatus     bool
	}{
		{
			name:               "uncached ASG status check with invalid spot request: add ASG to cache and negative status",
			asgName:            "myasg",
			availabilityZone:   "eu-west-1b",
			iamInstanceProfile: "123",
			instanceType:       "m4.4xlarge",
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.4xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1b",
					State:            "failed",
					Status:           "bad-parameters",
				},
				{
					ID:               "45",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
			awsStatusCache: map[api.AWSAsgName]*asgSpotStatus{},
			expectedCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1b",
					IamInstanceProfile: "123",
					InstanceType:       "m4.4xlarge",
					Available:          false,
				},
			},
			expectedStatus: false,
		},
		{
			name:               "uncached ASG status check with no known spot request: add ASG to cache and positive status",
			asgName:            "myasg",
			availabilityZone:   "eu-west-1b",
			iamInstanceProfile: "123",
			instanceType:       "m4.4xlarge",
			awsRequests: []*api.SpotRequest{
				{
					ID:               "14",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "failed",
					Status:           "bad-parameters",
				},
				{
					ID:               "45",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "123",
					AvailabilityZone: "eu-west-1a",
					State:            "open",
					Status:           "active",
				},
			},
			awsStatusCache: map[api.AWSAsgName]*asgSpotStatus{},
			expectedCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1b",
					IamInstanceProfile: "123",
					InstanceType:       "m4.4xlarge",
					Available:          true,
				},
			},
			expectedStatus: true,
		},
		{
			name:               "cached ASG status check: returns status",
			asgName:            "myasg",
			availabilityZone:   "eu-west-1b",
			iamInstanceProfile: "123",
			instanceType:       "m4.4xlarge",
			awsRequests:        []*api.SpotRequest{},
			awsStatusCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1b",
					IamInstanceProfile: "123",
					InstanceType:       "m4.4xlarge",
					Available:          false,
					statusChangeTime:   statusCreateTime,
				},
			},
			expectedCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1b",
					IamInstanceProfile: "123",
					InstanceType:       "m4.4xlarge",
					Available:          false,
					statusChangeTime:   statusCreateTime,
				},
			},
			expectedStatus: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := newAwsEC2SpotRequestManagerMock([]*ec2.SpotInstanceRequest{})

			monitor := NewSpotAvailabilityMonitor(mock, time.Minute, time.Hour)
			monitor.requestCache.cache = c.awsRequests
			monitor.statusCache.cache = c.awsStatusCache

			status := monitor.AsgAvailability(c.asgName, c.iamInstanceProfile, c.availabilityZone, c.instanceType)

			assert.Equal(t, c.expectedStatus, status, c.name, "returned status as not the expected value")
			assert.Equal(t, len(c.expectedCache), len(monitor.statusCache.cache), c.name, "status cache has not the expected entry length")

			for _, expectedEntry := range c.expectedCache {
				entry := monitor.statusCache.get(expectedEntry.AsgName)

				assert.Equal(t, expectedEntry.AsgName, entry.AsgName, c.name, "unexpected ASG name")
				assert.Equal(t, expectedEntry.AvailabilityZone, entry.AvailabilityZone, c.name, "unexpected Availability Zone")
				assert.Equal(t, expectedEntry.IamInstanceProfile, entry.IamInstanceProfile, c.name, "unexpected Iam Instance Profile")
				assert.Equal(t, expectedEntry.InstanceType, entry.InstanceType, c.name, "unexpected Instance Type")
				assert.Equal(t, expectedEntry.Available, entry.Available, c.name, "unexpected status")
			}
		})
	}
}

func TestSpotAvailabilityMonitor_updateRequestCache(t *testing.T) {
	cases := []struct {
		name          string
		awsRequests   []*ec2.SpotInstanceRequest
		requests      []*api.SpotRequest
		expectedError string
		error         string
	}{
		{
			name:          "error fetching AWS spot requests: returns an error",
			awsRequests:   []*ec2.SpotInstanceRequest{},
			requests:      []*api.SpotRequest{},
			expectedError: "could not retrieve AWS Spot Request list: AWS died",
			error:         "AWS died",
		},
		{
			name:        "no spot request found: empty cache",
			awsRequests: []*ec2.SpotInstanceRequest{},
			requests:    []*api.SpotRequest{},
		},
		{
			name: "only known spot request ids provided: no error",
			awsRequests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					"active", "222",
					"m4.2xlarge", "eu-west-1c", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			requests: []*api.SpotRequest{
				{
					ID:               "13",
					InstanceType:     "m4.2xlarge",
					InstanceProfile:  "222",
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
			mock := newAwsEC2SpotRequestManagerMock(c.awsRequests)
			mock.setError(c.error)

			var err error
			monitor := NewSpotAvailabilityMonitor(mock, time.Minute, time.Hour)

			err = monitor.updateRequestCache()

			if len(c.error) > 0 {
				assert.NotNil(t, err, c.name, "awaits an error")

				if err != nil {
					assert.Equal(t, c.expectedError, err.Error(), c.name, "unexpected error")
				}
			} else {
				assert.Nil(t, err, c.name, "no error should have append")
				if err == nil {
					assert.Equal(t, c.requests, monitor.requestCache.cache, c.name, "request state has not the awaited state")
				}
			}
		})
	}
}

func TestSpotAvailabilityMonitor_roundtrip(t *testing.T) {
	statusCreateTime := fluxCompensator(time.Hour)
	cases := []struct {
		name            string
		awsSpotRequests []*ec2.SpotInstanceRequest
		awsStatusCache  map[api.AWSAsgName]*asgSpotStatus
		expectedCache   map[api.AWSAsgName]*asgSpotStatus
		expectedError   string
		error           string
	}{
		{
			name:            "error raised while using AWS APIs: return the error",
			awsSpotRequests: []*ec2.SpotInstanceRequest{},
			awsStatusCache:  map[api.AWSAsgName]*asgSpotStatus{},
			expectedCache:   map[api.AWSAsgName]*asgSpotStatus{},
			expectedError:   "could not retrieve AWS Spot Request list: AWS Died",
			error:           "AWS Died",
		},
		{
			name: "no ASG status entries: no changes, only spot request cache update",
			awsSpotRequests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					"active", "222",
					"m4.2xlarge", "eu-west-1c", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			awsStatusCache: map[api.AWSAsgName]*asgSpotStatus{},
			expectedCache:  map[api.AWSAsgName]*asgSpotStatus{},
		},
		{
			name: "valid spot request for known available ASG found: no change",
			awsSpotRequests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					"active", "222",
					"m4.2xlarge", "eu-west-1c", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			awsStatusCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1c",
					IamInstanceProfile: "222",
					InstanceType:       "m4.2xlarge",
					Available:          true,
					statusChangeTime:   statusCreateTime,
				},
			},
			expectedCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1c",
					IamInstanceProfile: "222",
					InstanceType:       "m4.2xlarge",
					Available:          true,
					statusChangeTime:   statusCreateTime,
				},
			},
		},
		{
			name: "invalid spot request for known available ASG found: ASG status changes to unavailable",
			awsSpotRequests: []*ec2.SpotInstanceRequest{
				newSpotInstanceRequestInstance("12", "fulfilled",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*35)),
				newSpotInstanceRequestInstance("13", "open",
					string(api.AWSSpotRequestStatusNotAvailable), "222",
					"m4.2xlarge", "eu-west-1c", fluxCompensatorAWS(time.Minute*30)),
				newSpotInstanceRequestInstance("14", "failed",
					"bad-parameters", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute*5)),
				newSpotInstanceRequestInstance("15", "open",
					"active", "123",
					"m4.2xlarge", "eu-west-1a", fluxCompensatorAWS(time.Minute)),
			},
			awsStatusCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1c",
					IamInstanceProfile: "222",
					InstanceType:       "m4.2xlarge",
					Available:          true,
					statusChangeTime:   statusCreateTime,
				},
			},
			expectedCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1c",
					IamInstanceProfile: "222",
					InstanceType:       "m4.2xlarge",
					Available:          false,
					statusChangeTime:   statusCreateTime,
				},
			},
		},
		{
			name:            "unavailable spot ASG gets a 'true' status but is still in the exclusion period: ASG stays unavailable",
			awsSpotRequests: []*ec2.SpotInstanceRequest{},
			awsStatusCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1c",
					IamInstanceProfile: "222",
					InstanceType:       "m4.2xlarge",
					Available:          false,
					statusChangeTime:   fluxCompensator(time.Minute * 30),
				},
			},
			expectedCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1c",
					IamInstanceProfile: "222",
					InstanceType:       "m4.2xlarge",
					Available:          false,
				},
			},
		},
		{
			name:            "unavailable spot ASG gets a 'true' status and is out of the exclusion period: ASG becomes available",
			awsSpotRequests: []*ec2.SpotInstanceRequest{},
			awsStatusCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1c",
					IamInstanceProfile: "222",
					InstanceType:       "m4.2xlarge",
					Available:          false,
					statusChangeTime:   fluxCompensator(time.Minute * 90),
				},
			},
			expectedCache: map[api.AWSAsgName]*asgSpotStatus{
				"myasg": {
					AsgName:            "myasg",
					AvailabilityZone:   "eu-west-1c",
					IamInstanceProfile: "222",
					InstanceType:       "m4.2xlarge",
					Available:          true,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := newAwsEC2SpotRequestManagerMock(c.awsSpotRequests)
			mock.setError(c.error)

			asgNames := make([]api.AWSAsgName, 0)
			for _, status := range c.awsStatusCache {
				asgNames = append(asgNames, status.AsgName)
			}

			monitor := NewSpotAvailabilityMonitor(mock, time.Minute, time.Hour)
			monitor.statusCache.asgNames = asgNames
			monitor.statusCache.cache = c.awsStatusCache

			err := monitor.roundtrip()

			if len(c.error) > 0 {
				assert.NotNil(t, err, c.name, "awaits an error")

				if err != nil {
					assert.Equal(t, c.expectedError, err.Error(), c.name, "unexpected error")
				}
			} else {
				assert.Nil(t, err, c.name, "no error should have append")
				for _, expectedEntry := range c.expectedCache {
					entry := monitor.statusCache.get(expectedEntry.AsgName)
					assert.NotNil(t, entry, c.name, "ASG status should be known for", expectedEntry.AsgName)

					if entry != nil {
						assert.Equal(t, expectedEntry.AsgName, entry.AsgName, c.name, "unexpected ASG name")
						assert.Equal(t, expectedEntry.AvailabilityZone, entry.AvailabilityZone, c.name, "unexpected Availability Zone")
						assert.Equal(t, expectedEntry.IamInstanceProfile, entry.IamInstanceProfile, c.name, "unexpected Iam Instance Profile")
						assert.Equal(t, expectedEntry.InstanceType, entry.InstanceType, c.name, "unexpected Instance Type")
						assert.Equal(t, expectedEntry.Available, entry.Available, c.name, "unexpected status")
					}
				}
			}
		})
	}
}

var _ api.AwsEC2SpotRequestManager = &awsEC2SpotRequestManagerMock{}

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

	startTime := time.Time{}
	searchedStates := make([]*string, 0)

	for _, filter := range input.Filters {
		switch aws.StringValue(filter.Name) {
		case api.InputStateFilter:
			for _, state := range filter.Values {
				searchedStates = append(searchedStates, state)
			}
		}
	}

	requests := make([]*ec2.SpotInstanceRequest, 0)

	for _, request := range m.requests {
		if aws.TimeValue(request.CreateTime).After(startTime) {
			for _, state := range searchedStates {
				if aws.StringValue(request.State) == aws.StringValue(state) {
					requests = append(requests, request)
					break
				}
			}
		}
	}

	return &ec2.DescribeSpotInstanceRequestsOutput{SpotInstanceRequests: requests}, nil
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
