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
	"testing"

	"github.com/stretchr/testify/assert"
	featuregatetesting "k8s.io/component-base/featuregate/testing"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
)

func TestIsScopedDaemonSetRequiresFeatureGate(t *testing.T) {
	assert.False(t, IsScopedDaemonSet("DaemonSet", "node.kubernetes.io/instance-type"))

	featuregatetesting.SetFeatureGateDuringTest(t, features.MutableFeatureGate, features.DaemonSetScope, true)
	assert.True(t, IsScopedDaemonSet("DaemonSet", "node.kubernetes.io/instance-type"))
	assert.False(t, IsScopedDaemonSet("Deployment", "node.kubernetes.io/instance-type"))
	assert.False(t, IsScopedDaemonSet("DaemonSet", ""))
}

func TestEncodeDecodeCheckpointScopeValue(t *testing.T) {
	// The empty scope value must round-trip through a non-empty sentinel so it is
	// distinguishable from a non-scoped checkpoint (which stores an empty value).
	assert.Equal(t, EmptyLabelValue, EncodeCheckpointScopeValue(""))
	assert.Equal(t, "", DecodeCheckpointScopeValue(EmptyLabelValue))

	for _, scopeValue := range []string{"worker", "master", AbsentLabelValue} {
		encoded := EncodeCheckpointScopeValue(scopeValue)
		assert.NotEmpty(t, encoded)
		assert.Equal(t, scopeValue, DecodeCheckpointScopeValue(encoded))
	}
}

func TestCheckpointName(t *testing.T) {
	// A non-scoped checkpoint keeps the historical name format.
	assert.Equal(t, "vpa-1-agent", CheckpointName("vpa-1", "agent", ""))

	// Scoped checkpoints append a stable hash of the persisted scope value.
	worker := CheckpointName("vpa-1", "agent", "worker")
	master := CheckpointName("vpa-1", "agent", "master")
	assert.NotEqual(t, worker, master)
	assert.NotEqual(t, "vpa-1-agent", worker)
	// The name is deterministic.
	assert.Equal(t, worker, CheckpointName("vpa-1", "agent", "worker"))
}
