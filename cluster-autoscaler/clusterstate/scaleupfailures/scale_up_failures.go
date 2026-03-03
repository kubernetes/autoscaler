/*
Copyright 2016 The Kubernetes Authors.

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

package scaleupfailures

import (
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
)

// ScaleUpFailure contains information about a failure of a scale-up.
type ScaleUpFailure struct {
	NodeGroup cloudprovider.NodeGroup
	Reason    metrics.FailedScaleUpReason
	Time      time.Time
}

// ScaleUpFailuresRegistry contains information about scale-up failures.
type ScaleUpFailuresRegistry struct {
	mu       sync.Mutex
	failures map[string][]ScaleUpFailure
}

// NewScaleUpFailuresRegistry returns a new ScaleUpFailuresRegistry.
func NewScaleUpFailuresRegistry() *ScaleUpFailuresRegistry {
	return &ScaleUpFailuresRegistry{
		failures: make(map[string][]ScaleUpFailure),
	}
}

// Register registers a scale-up failure.
func (s *ScaleUpFailuresRegistry) Register(nodeGroup cloudprovider.NodeGroup, reason metrics.FailedScaleUpReason, time time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures[nodeGroup.Id()] = append(s.failures[nodeGroup.Id()], ScaleUpFailure{NodeGroup: nodeGroup, Reason: reason, Time: time})
}

// Clear clears the scale-up failures.
func (s *ScaleUpFailuresRegistry) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures = make(map[string][]ScaleUpFailure)
}

// Get returns the scale-up failures.
func (s *ScaleUpFailuresRegistry) Get() map[string][]ScaleUpFailure {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string][]ScaleUpFailure)
	for nodeGroupId, failures := range s.failures {
		result[nodeGroupId] = failures
	}
	return result
}
