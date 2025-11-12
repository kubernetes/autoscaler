/*
Copyright 2023 The Kubernetes Authors.

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

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
)

// SkippedReasons contains information why given node group was skipped.
type SkippedReasons struct {
	messages []string
}

// NewSkippedReasons creates new SkippedReason object.
func NewSkippedReasons(m string) *SkippedReasons {
	return &SkippedReasons{[]string{m}}
}

// Reasons returns a slice of reasons why the node group was not considered for scale up.
func (sr *SkippedReasons) Reasons() []string {
	return sr.messages
}

var (
	// BackoffReason node group is in backoff.
	BackoffReason = NewSkippedReasons("in backoff after failed scale-up")
	// MaxLimitReachedReason node group reached max size limit.
	MaxLimitReachedReason = NewSkippedReasons("max node group size reached")
	// NotReadyReason node group is not ready.
	NotReadyReason = NewSkippedReasons("not ready for scale-up")
)

// MaxResourceLimitReached contains information why given node group was skipped.
type MaxResourceLimitReached struct {
	messages  []string
	resources []string
}

// Reasons returns a slice of reasons why the node group was not considered for scale up.
func (sr *MaxResourceLimitReached) Reasons() []string {
	return sr.messages
}

// Resources returns a slice of resources which were missing in the node group.
func (sr *MaxResourceLimitReached) Resources() []string {
	return sr.resources
}

// NewMaxResourceLimitReached returns a reason describing which resource limits were reached.
func NewMaxResourceLimitReached(exceededQuotas []resourcequotas.ExceededQuota) *MaxResourceLimitReached {
	var messages []string
	resources := make(sets.Set[string])
	for _, quota := range exceededQuotas {
		msg := fmt.Sprintf("exceeded quota: %q, resources: %s", quota.ID, strings.Join(quota.ExceededResources, ", "))
		messages = append(messages, msg)
		resources.Insert(quota.ExceededResources...)
	}
	return &MaxResourceLimitReached{
		messages:  messages,
		resources: resources.UnsortedList(),
	}
}
