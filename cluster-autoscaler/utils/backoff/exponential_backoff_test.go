/*
Copyright 2018 The Kubernetes Authors.

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

package backoff

import (
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"

	"github.com/stretchr/testify/assert"
)

func nodeGroup(id string) cloudprovider.NodeGroup {
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup(id, 1, 10, 1)
	return provider.GetNodeGroup(id)
}

var nodeGroup1 = nodeGroup("id1")
var nodeGroup2 = nodeGroup("id2")

var quotaError = cloudprovider.InstanceErrorInfo{ErrorClass: cloudprovider.OutOfResourcesErrorClass, ErrorCode: "QUOTA_EXCEEDED", ErrorMessage: "Not enough CPU"}
var ipSpaceExhaustedError = cloudprovider.InstanceErrorInfo{ErrorClass: cloudprovider.OtherErrorClass, ErrorCode: "IP_SPACE_EXHAUSTED", ErrorMessage: "IP space has been exhausted"}

var noBackOff = Status{IsBackedOff: false}
var backoffWithQuotaError = Status{
	IsBackedOff: true,
	ErrorInfo:   quotaError,
}
var backoffWithIpSpaceExhaustedError = Status{
	IsBackedOff: true,
	ErrorInfo:   ipSpaceExhaustedError,
}

func TestBackoffTwoKeys(t *testing.T) {
	backoff := NewIdBasedExponentialBackoff(10*time.Minute, time.Hour, 3*time.Hour)
	startTime := time.Now()
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup1, nil, startTime))
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup2, nil, startTime))
	backoff.Backoff(nodeGroup1, nil, quotaError, startTime.Add(time.Minute))
	assert.Equal(t, backoffWithQuotaError, backoff.BackoffStatus(nodeGroup1, nil, startTime.Add(2*time.Minute)))
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup2, nil, startTime))
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup1, nil, startTime.Add(11*time.Minute+1*time.Millisecond)))
}

func TestMaxBackoff(t *testing.T) {
	backoff := NewIdBasedExponentialBackoff(1*time.Minute, 3*time.Minute, 3*time.Hour)
	startTime := time.Now()
	backoff.Backoff(nodeGroup1, nil, ipSpaceExhaustedError, startTime)
	assert.Equal(t, backoffWithIpSpaceExhaustedError, backoff.BackoffStatus(nodeGroup1, nil, startTime))
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup1, nil, startTime.Add(1*time.Minute+1*time.Millisecond)))
	backoff.Backoff(nodeGroup1, nil, ipSpaceExhaustedError, startTime.Add(1*time.Minute))
	assert.Equal(t, backoffWithIpSpaceExhaustedError, backoff.BackoffStatus(nodeGroup1, nil, startTime.Add(1*time.Minute)))
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup1, nil, startTime.Add(3*time.Minute)))
	backoff.Backoff(nodeGroup1, nil, ipSpaceExhaustedError, startTime.Add(3*time.Minute))
	assert.Equal(t, backoffWithIpSpaceExhaustedError, backoff.BackoffStatus(nodeGroup1, nil, startTime.Add(3*time.Minute)))
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup1, nil, startTime.Add(6*time.Minute)))
}

func TestRemoveBackoff(t *testing.T) {
	backoff := NewIdBasedExponentialBackoff(1*time.Minute, 3*time.Minute, 3*time.Hour)
	startTime := time.Now()
	backoff.Backoff(nodeGroup1, nil, quotaError, startTime)
	assert.Equal(t, backoffWithQuotaError, backoff.BackoffStatus(nodeGroup1, nil, startTime))
	backoff.RemoveBackoff(nodeGroup1, nil)
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup1, nil, startTime))
}

func TestResetStaleBackoffData(t *testing.T) {
	backoff := NewIdBasedExponentialBackoff(1*time.Minute, 3*time.Minute, 3*time.Hour)
	startTime := time.Now()
	backoff.Backoff(nodeGroup1, nil, quotaError, startTime)
	backoff.Backoff(nodeGroup2, nil, quotaError, startTime.Add(time.Hour))
	backoff.RemoveStaleBackoffData(startTime.Add(time.Hour))
	assert.Equal(t, 2, len(backoff.(*exponentialBackoff).backoffInfo))
	backoff.RemoveStaleBackoffData(startTime.Add(4 * time.Hour))
	assert.Equal(t, 1, len(backoff.(*exponentialBackoff).backoffInfo))
	backoff.RemoveStaleBackoffData(startTime.Add(5 * time.Hour))
	assert.Equal(t, 0, len(backoff.(*exponentialBackoff).backoffInfo))
}

func TestIncreaseExistingBackoff(t *testing.T) {
	backoff := NewIdBasedExponentialBackoff(1*time.Second, 10*time.Minute, 3*time.Hour)
	currentTime := time.Date(2023, 12, 12, 12, 0, 0, 0, time.UTC)
	backoff.Backoff(nodeGroup1, nil, quotaError, currentTime)
	// NG in backoff for one second here
	assert.Equal(t, backoffWithQuotaError, backoff.BackoffStatus(nodeGroup1, nil, currentTime))
	// Come out of backoff
	currentTime = currentTime.Add(1*time.Second + 1*time.Millisecond)
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup1, nil, currentTime))
	// Confirm existing backoff duration and error info have been increased by backing off again
	backoff.Backoff(nodeGroup1, nil, ipSpaceExhaustedError, currentTime)
	// Backoff should be for 2 seconds now
	assert.Equal(t, backoffWithIpSpaceExhaustedError, backoff.BackoffStatus(nodeGroup1, nil, currentTime))
	currentTime = currentTime.Add(1 * time.Second)
	// Doing backoff during existing backoff should change error info and backoff end period but doesn't change the duration.
	backoff.Backoff(nodeGroup1, nil, quotaError, currentTime)
	assert.Equal(t, backoffWithQuotaError, backoff.BackoffStatus(nodeGroup1, nil, currentTime))
	currentTime = currentTime.Add(2*time.Second + 1*time.Millisecond)
	assert.Equal(t, noBackOff, backoff.BackoffStatus(nodeGroup1, nil, currentTime))
	// Result: existing backoff duration was scaled up beyond initial duration
}
