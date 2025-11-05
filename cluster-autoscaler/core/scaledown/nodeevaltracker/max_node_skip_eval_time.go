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

package nodeevaltracker

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/metrics"
)

// MaxNodeSkipEvalTime is a time tracker for the biggest evaluation time of a node during ScaleDown
type MaxNodeSkipEvalTime struct {
	// lastEvalTime is the time of previous currentlyUnneededNodeNames parsing
	lastEvalTime time.Time
	// nodeNamesWithTimeStamps is maps of nodeNames with their time of last successful evaluation
	nodeNamesWithTimeStamps map[string]time.Time
}

// NewMaxNodeSkipEvalTime returns LongestNodeScaleDownEvalTime with lastEvalTime set to currentTime
func NewMaxNodeSkipEvalTime(currentTime time.Time) *MaxNodeSkipEvalTime {
	return &MaxNodeSkipEvalTime{lastEvalTime: currentTime}
}

// Retrieves the time of the last evaluation of a node.
func (l *MaxNodeSkipEvalTime) get(nodeName string) time.Time {
	if _, ok := l.nodeNamesWithTimeStamps[nodeName]; ok {
		return l.nodeNamesWithTimeStamps[nodeName]
	}
	return l.lastEvalTime
}

// getMin() returns the minimum time in nodeNamesWithTimeStamps or time of last evaluation
func (l *MaxNodeSkipEvalTime) getMin() time.Time {
	minimumTime := l.lastEvalTime
	for _, val := range l.nodeNamesWithTimeStamps {
		if minimumTime.After(val) {
			minimumTime = val
		}
	}
	return minimumTime
}

// Update returns the longest evaluation time for the nodes in nodeNamesWithTimeStamps
// and changes nodeNamesWithTimeStamps for nodeNames.
func (l *MaxNodeSkipEvalTime) Update(nodeNames []string, currentTime time.Time) time.Duration {
	newNodes := make(map[string]time.Time)
	for _, nodeName := range nodeNames {
		newNodes[nodeName] = l.get(nodeName)
	}
	l.nodeNamesWithTimeStamps = newNodes
	l.lastEvalTime = currentTime
	minimumTime := l.getMin()
	longestDuration := currentTime.Sub(minimumTime)
	metrics.ObserveMaxNodeSkipEvalDurationSeconds(longestDuration)
	return longestDuration
}
