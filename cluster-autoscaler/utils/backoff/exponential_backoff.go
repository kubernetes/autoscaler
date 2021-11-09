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
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// Backoff handles backing off executions.
type exponentialBackoff struct {
	maxBackoffDuration     time.Duration
	initialBackoffDuration time.Duration
	backoffResetTimeout    time.Duration
	backoffInfo            map[string]exponentialBackoffInfo
	nodeGroupKey           func(nodeGroup cloudprovider.NodeGroup) string
}

type exponentialBackoffInfo struct {
	duration            time.Duration
	backoffUntil        time.Time
	lastFailedExecution time.Time
}

// NewExponentialBackoff creates an instance of exponential backoff.
func NewExponentialBackoff(
	initialBackoffDuration time.Duration,
	maxBackoffDuration time.Duration,
	backoffResetTimeout time.Duration,
	nodeGroupKey func(nodeGroup cloudprovider.NodeGroup) string) Backoff {
	return &exponentialBackoff{
		maxBackoffDuration:     maxBackoffDuration,
		initialBackoffDuration: initialBackoffDuration,
		backoffResetTimeout:    backoffResetTimeout,
		backoffInfo:            make(map[string]exponentialBackoffInfo),
		nodeGroupKey:           nodeGroupKey,
	}
}

// NewIdBasedExponentialBackoff creates an instance of exponential backoff with node group Id used as a key.
func NewIdBasedExponentialBackoff(initialBackoffDuration time.Duration, maxBackoffDuration time.Duration, backoffResetTimeout time.Duration) Backoff {
	return NewExponentialBackoff(
		initialBackoffDuration,
		maxBackoffDuration,
		backoffResetTimeout,
		func(nodeGroup cloudprovider.NodeGroup) string {
			return nodeGroup.Id()
		})
}

// Backoff execution for the given node group. Returns time till execution is backed off.
func (b *exponentialBackoff) Backoff(nodeGroup cloudprovider.NodeGroup, nodeInfo *schedulerframework.NodeInfo, errorClass cloudprovider.InstanceErrorClass, errorCode string, currentTime time.Time) time.Time {
	duration := b.initialBackoffDuration
	key := b.nodeGroupKey(nodeGroup)
	if backoffInfo, found := b.backoffInfo[key]; found {
		// Multiple concurrent scale-ups failing shouldn't cause
		// backoff duration to increase exponentially
		duration = backoffInfo.duration
		if backoffInfo.backoffUntil.Before(currentTime) {
			// NodeGroup is not currently in backoff, but was recently
			// Increase backoff duration exponentially
			duration = 2 * backoffInfo.duration
			if duration > b.maxBackoffDuration {
				duration = b.maxBackoffDuration
			}
		}
	}
	backoffUntil := currentTime.Add(duration)
	b.backoffInfo[key] = exponentialBackoffInfo{
		duration:            duration,
		backoffUntil:        backoffUntil,
		lastFailedExecution: currentTime,
	}
	return backoffUntil
}

// IsBackedOff returns true if execution is backed off for the given node group.
func (b *exponentialBackoff) IsBackedOff(nodeGroup cloudprovider.NodeGroup, nodeInfo *schedulerframework.NodeInfo, currentTime time.Time) bool {
	backoffInfo, found := b.backoffInfo[b.nodeGroupKey(nodeGroup)]
	return found && backoffInfo.backoffUntil.After(currentTime)
}

// RemoveBackoff removes backoff data for the given node group.
func (b *exponentialBackoff) RemoveBackoff(nodeGroup cloudprovider.NodeGroup, nodeInfo *schedulerframework.NodeInfo) {
	delete(b.backoffInfo, b.nodeGroupKey(nodeGroup))
}

// RemoveStaleBackoffData removes stale backoff data.
func (b *exponentialBackoff) RemoveStaleBackoffData(currentTime time.Time) {
	for key, backoffInfo := range b.backoffInfo {
		if backoffInfo.lastFailedExecution.Add(b.backoffResetTimeout).Before(currentTime) {
			delete(b.backoffInfo, key)
		}
	}
}
