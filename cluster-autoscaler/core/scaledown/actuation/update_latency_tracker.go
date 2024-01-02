/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"
)

const sleepDurationWhenPolling = 50 * time.Millisecond
const waitForTaintingTimeoutDuration = 30 * time.Second

type nodeTaintStartTime struct {
	nodeName  string
	startTime time.Time
}

// UpdateLatencyTracker can be used to calculate round-trip time between CA and api-server
// when adding ToBeDeletedTaint to nodes
type UpdateLatencyTracker struct {
	startTimestamp     map[string]time.Time
	finishTimestamp    map[string]time.Time
	remainingNodeCount int
	nodeLister         kubernetes.NodeLister
	// Sends node tainting start timestamps to the tracker
	StartTimeChan            chan nodeTaintStartTime
	sleepDurationWhenPolling time.Duration
	// Passing a bool will wait for all the started nodes to get tainted and calculate
	// latency based on latencies observed. (If all the nodes did not get tained within
	// waitForTaintingTimeoutDuration after passing a bool, latency calculation will be
	// aborted and the ResultChan will be closed without returning a value) Closing the
	// AwaitOrStopChan without passing any bool will abort the latency calculation.
	AwaitOrStopChan chan bool
	// Communicate back the measured latency
	ResultChan chan time.Duration
	// now is used only to make the testing easier
	now func() time.Time
}

// NewUpdateLatencyTracker returns a new NewUpdateLatencyTracker object
func NewUpdateLatencyTracker(nodeLister kubernetes.NodeLister) *UpdateLatencyTracker {
	return &UpdateLatencyTracker{
		startTimestamp:           map[string]time.Time{},
		finishTimestamp:          map[string]time.Time{},
		remainingNodeCount:       0,
		nodeLister:               nodeLister,
		StartTimeChan:            make(chan nodeTaintStartTime),
		sleepDurationWhenPolling: sleepDurationWhenPolling,
		AwaitOrStopChan:          make(chan bool),
		ResultChan:               make(chan time.Duration),
		now:                      time.Now,
	}
}

// Start starts listening for node tainting start timestamps and update the timestamps that
// the taint appears for the first time for a particular node. Listen AwaitOrStopChan for stop/await signals
func (u *UpdateLatencyTracker) Start() {
	for {
		select {
		case _, ok := <-u.AwaitOrStopChan:
			if ok {
				u.await()
			}
			return
		case ntst := <-u.StartTimeChan:
			u.startTimestamp[ntst.nodeName] = ntst.startTime
			u.remainingNodeCount += 1
			continue
		default:
		}
		u.updateFinishTime()
		time.Sleep(u.sleepDurationWhenPolling)
	}
}

func (u *UpdateLatencyTracker) updateFinishTime() {
	for nodeName := range u.startTimestamp {
		if _, ok := u.finishTimestamp[nodeName]; ok {
			continue
		}
		node, err := u.nodeLister.Get(nodeName)
		if err != nil {
			klog.Errorf("Error getting node: %v", err)
			continue
		}
		if taints.HasToBeDeletedTaint(node) {
			u.finishTimestamp[node.Name] = u.now()
			u.remainingNodeCount -= 1
		}
	}
}

func (u *UpdateLatencyTracker) calculateLatency() time.Duration {
	var maxLatency time.Duration = 0
	for node, startTime := range u.startTimestamp {
		endTime, _ := u.finishTimestamp[node]
		currentLatency := endTime.Sub(startTime)
		if currentLatency > maxLatency {
			maxLatency = currentLatency
		}
	}
	return maxLatency
}

func (u *UpdateLatencyTracker) await() {
	waitingForTaintingStartTime := time.Now()
	for {
		switch {
		case u.remainingNodeCount == 0:
			latency := u.calculateLatency()
			u.ResultChan <- latency
			return
		case time.Now().After(waitingForTaintingStartTime.Add(waitForTaintingTimeoutDuration)):
			klog.Errorf("Timeout before tainting all nodes, latency measurement will be stale")
			close(u.ResultChan)
			return
		default:
			time.Sleep(u.sleepDurationWhenPolling)
			u.updateFinishTime()
		}
	}
}

// NewUpdateLatencyTrackerForTesting returns a UpdateLatencyTracker object with
// reduced sleepDurationWhenPolling and mock clock for testing
func NewUpdateLatencyTrackerForTesting(nodeLister kubernetes.NodeLister, now func() time.Time) *UpdateLatencyTracker {
	updateLatencyTracker := NewUpdateLatencyTracker(nodeLister)
	updateLatencyTracker.now = now
	updateLatencyTracker.sleepDurationWhenPolling = time.Millisecond
	return updateLatencyTracker
}
