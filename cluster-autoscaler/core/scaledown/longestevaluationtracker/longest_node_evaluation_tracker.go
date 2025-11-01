/*
Copyright 2025 The Kubernetes Authors.

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

package longestevaluationtracker

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/metrics"
)

// LongestNodeScaleDownEvalTime is a time tracker for the longest evaluation time of a node during ScaleDown
type LongestNodeScaleDownEvalTime struct {
	// lastEvalTime is the time of previous currentlyUnneededNodeNames parsing
	lastEvalTime time.Time
	// NodeNamesWithTimeStamps is maps of nodeNames with their time of last successful evaluation
	NodeNamesWithTimeStamps map[string]time.Time
}

// NewLongestNodeScaleDownEvalTime returns LongestNodeScaleDownEvalTime with lastEvalTime set to currentTime
func NewLongestNodeScaleDownEvalTime(currentTime time.Time) *LongestNodeScaleDownEvalTime {
	return &LongestNodeScaleDownEvalTime{lastEvalTime: currentTime}
}

// Retrieves the time of the last evaluation of a node.
func (l *LongestNodeScaleDownEvalTime) get(nodeName string) time.Time {
	if _, ok := l.NodeNamesWithTimeStamps[nodeName]; ok {
		return l.NodeNamesWithTimeStamps[nodeName]
	}
	return l.lastEvalTime
}

// getMin() returns the minimum time in NodeNamesWithTimeStamps or time of last evaluation
func (l *LongestNodeScaleDownEvalTime) getMin() time.Time {
	minimumTime := l.lastEvalTime
	for _, val := range l.NodeNamesWithTimeStamps {
		if minimumTime.After(val) {
			minimumTime = val
		}
	}
	return minimumTime
}

// Update returns the longest evaluation time for the nodes in NodeNamesWithTimeStamps
// and changes NodeNamesWithTimeStamps for nodeNames.
func (l *LongestNodeScaleDownEvalTime) Update(nodeNames []string, currentTime time.Time) time.Duration {
	var longestTime time.Duration
	minimumTime := l.getMin()
	if len(nodeNames) == 0 {
		// if l.minimumTime is lastEvalTime, then in previous iteration we also processed all the nodes, so the longest time is 0
		// otherwise -> report the longest time from previous iteration and reset the minimumTime
		if minimumTime.Equal(l.lastEvalTime) {
			longestTime = 0
		} else {
			longestTime = currentTime.Sub(minimumTime)
		}
		l.NodeNamesWithTimeStamps = make(map[string]time.Time)
	} else {
		newNodes := make(map[string]time.Time, len(nodeNames))
		for _, nodeName := range nodeNames {
			newNodes[nodeName] = l.get(nodeName)
		}
		l.NodeNamesWithTimeStamps = newNodes
		longestTime = currentTime.Sub(minimumTime)
	}
	l.lastEvalTime = currentTime
	metrics.ObserveLongestUnneededNodeScaleDownEvalDurationSeconds(longestTime)
	return longestTime
}
