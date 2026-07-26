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

package scope

import (
	"fmt"
	"hash/fnv"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
)

const (
	daemonSetKind = "DaemonSet"
	labelPrefix   = "__vpa_scope_"
	// AbsentLabelValue is used for scoped DaemonSet grouping when a node doesn't have scope label key.
	AbsentLabelValue = "__absent__"
	// EmptyLabelValue is the persisted sentinel for a scope value equal to the
	// empty string. It lets a scoped checkpoint for an empty scope value be
	// distinguished from a non-scoped checkpoint (which stores an empty scope value).
	EmptyLabelValue = "__empty__"
)

// IsScopedDaemonSet returns true when DaemonSetScope is enabled, target is DaemonSet, and scopeKey is set.
func IsScopedDaemonSet(kind, scopeKey string) bool {
	return features.Enabled(features.DaemonSetScope) && kind == daemonSetKind && scopeKey != ""
}

// AggregationLabelKey returns a deterministic synthetic pod label key for scopeKey.
func AggregationLabelKey(scopeKey string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(scopeKey))
	return fmt.Sprintf("%s%x", labelPrefix, h.Sum32())
}

// EncodeCheckpointScopeValue maps a runtime scope value to the value persisted in
// a checkpoint. The empty scope value is stored as EmptyLabelValue so that it can
// be told apart from a non-scoped checkpoint.
func EncodeCheckpointScopeValue(scopeValue string) string {
	if scopeValue == "" {
		return EmptyLabelValue
	}
	return scopeValue
}

// DecodeCheckpointScopeValue reverses EncodeCheckpointScopeValue when loading a
// checkpoint back into the model.
func DecodeCheckpointScopeValue(persisted string) string {
	if persisted == EmptyLabelValue {
		return ""
	}
	return persisted
}

// CheckpointName returns a stable, RFC 1123 compliant checkpoint object name for
// the given VPA, container and (already persisted) scope value. Non-scoped
// checkpoints (empty persistedScopeValue) keep the historical "<vpa>-<container>"
// name; scoped checkpoints append a short hash of the scope value so arbitrary
// label values cannot produce an invalid or overly long object name.
func CheckpointName(vpaName, container, persistedScopeValue string) string {
	if persistedScopeValue == "" {
		return fmt.Sprintf("%s-%s", vpaName, container)
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(persistedScopeValue))
	return fmt.Sprintf("%s-%s-%x", vpaName, container, h.Sum32())
}
