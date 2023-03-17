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
)

// SkippedReasons contains information why given node group was skipped.
type SkippedReasons struct {
	messages []string
}

// Reasons returns a slice of reasons why the node group was not considered for scale up.
func (sr *SkippedReasons) Reasons() []string {
	return sr.messages
}

var (
	// BackoffReason node group is in backoff.
	BackoffReason = &SkippedReasons{[]string{"in backoff after failed scale-up"}}
	// MaxLimitReachedReason node group reached max size limit.
	MaxLimitReachedReason = &SkippedReasons{[]string{"max node group size reached"}}
	// NotReadyReason node group is not ready.
	NotReadyReason = &SkippedReasons{[]string{"not ready for scale-up"}}
)

// MaxResourceLimitReached returns a reason describing which cluster wide resource limits were reached.
func MaxResourceLimitReached(resources []string) *SkippedReasons {
	return &SkippedReasons{[]string{fmt.Sprintf("max cluster %s limit reached", strings.Join(resources, ", "))}}
}
