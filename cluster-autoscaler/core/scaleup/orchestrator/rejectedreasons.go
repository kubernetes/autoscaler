/*
Copyright 2024 The Kubernetes Authors.

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

package orchestrator

// RejectedReasons contains information why given node group was rejected as a scale-up option.
type RejectedReasons struct {
	messages []string
}

// NewRejectedReasons creates new RejectedReason object.
func NewRejectedReasons(m string) *RejectedReasons {
	return &RejectedReasons{[]string{m}}
}

// Reasons returns a slice of reasons why the node group was not considered for scale up.
func (sr *RejectedReasons) Reasons() []string {
	return sr.messages
}

var (
	// AllOrNothingReason means the node group was rejected because not all pods would fit it when using all-or-nothing strategy.
	AllOrNothingReason = NewRejectedReasons("not all pods would fit and scale-up is using all-or-nothing strategy")
)
