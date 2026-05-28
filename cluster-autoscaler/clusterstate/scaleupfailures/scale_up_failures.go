/*
Copyright The Kubernetes Authors.

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
)

// Record contains information about a failure of a scale-up.
type Record struct {
	ErrorInfo cloudprovider.InstanceErrorInfo
	Delta     int
	Time      time.Time
}

// Registry contains information about scale-up failures.
type Registry struct {
	mu       sync.Mutex
	failures map[string][]Record
}

// NewRegistry returns a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		failures: make(map[string][]Record),
	}
}

// RegisterScaleUp records when the last scale up happened for a nodegroup.
func (s *Registry) RegisterScaleUp(_ cloudprovider.NodeGroup,
	_ int, _ time.Time) {
}

// RegisterScaleDown records when the last scale down happened for a nodegroup.
func (s *Registry) RegisterScaleDown(_ cloudprovider.NodeGroup,
	_ string, _ time.Time, _ time.Time) {
}

// RegisterFailedScaleUp records when the last scale up failed for a nodegroup.
func (s *Registry) RegisterFailedScaleUp(nodeGroup cloudprovider.NodeGroup, delta int,
	errorInfo cloudprovider.InstanceErrorInfo, currentTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures[nodeGroup.Id()] = append(s.failures[nodeGroup.Id()], Record{ErrorInfo: errorInfo, Delta: delta, Time: currentTime})
}

// RegisterFailedScaleDown records failed scale-down for a nodegroup.
func (s *Registry) RegisterFailedScaleDown(_ cloudprovider.NodeGroup,
	_ string, _ time.Time) {
}

// Refresh clears the scale-up failures.
func (s *Registry) Refresh() {
	s.clear()
}

// Clear clears the scale-up failures.
func (s *Registry) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures = make(map[string][]Record)
}

// Get returns the scale-up failures.
func (s *Registry) Get() map[string][]Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string][]Record)
	for nodeGroupId, failures := range s.failures {
		failuresCopy := make([]Record, len(failures))
		copy(failuresCopy, failures)
		result[nodeGroupId] = failuresCopy
	}
	return result
}
