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

package logic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	k8stesting "k8s.io/client-go/testing"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// A VPA that has no status/conditions yet (the updater ran before the
// recommender populated it) must still get the condition, via a merge patch
// rather than a JSON Patch "add" on a missing parent path.
func TestReconcileTargetConflictsVPAWithoutStatus(t *testing.T) {
	client, _ := runReconcile([]conflictVpa{
		{name: "vpa-1", age: 2 * time.Hour},
		{name: "vpa-2", age: time.Hour},
	}, false)

	got := getVpa(t, client, "default", "vpa-2")
	require.Empty(t, got.Status.Recommendation, "test setup: VPA should have no recommendation")
	cond := findCondition(got, vpa_types.TargetConflict)
	require.NotNil(t, cond)
	assert.Equal(t, corev1.ConditionTrue, cond.Status)

	patches := 0
	for _, a := range client.Actions() {
		if a.GetVerb() != "patch" || a.GetSubresource() != "status" {
			continue
		}
		patches++
		pa, ok := a.(k8stesting.PatchAction)
		require.True(t, ok)
		assert.Equal(t, types.MergePatchType, pa.GetPatchType(),
			"no existing conditions: must not use a JSON Patch add on a possibly missing /status")
	}
	assert.Equal(t, 1, patches)
}
