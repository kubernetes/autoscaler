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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_fake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
)

// With no existing conditions the merge patch has no JSON Patch "test" op, so it
// must be guarded by metadata.resourceVersion instead.
func TestReconcileTargetConflictsMergePatchCarriesResourceVersion(t *testing.T) {
	older := conflictVpa{name: "vpa-1", age: 2 * time.Hour}
	newer := conflictVpa{name: "vpa-2", age: time.Hour}
	client := vpa_fake.NewSimpleClientset(older.build(), newer.build())

	listedNewer := newer.build()
	listedNewer.ResourceVersion = "42"
	listed := []*vpa_types.VerticalPodAutoscaler{older.build(), listedNewer}

	var patchBody []byte
	client.PrependReactor("patch", "verticalpodautoscalers", func(a k8stesting.Action) (bool, runtime.Object, error) {
		if pa, ok := a.(k8stesting.PatchAction); ok && pa.GetName() == "vpa-2" {
			patchBody = pa.GetPatch()
		}
		return false, nil, nil
	})

	u := &updater{vpaClient: client, eventRecorder: record.NewFakeRecorder(20)}
	u.reconcileTargetConflicts(listed)

	require.NotEmpty(t, patchBody, "expected a status patch for vpa-2")
	var body map[string]any
	require.NoError(t, json.Unmarshal(patchBody, &body))
	meta, ok := body["metadata"].(map[string]any)
	require.True(t, ok, "merge patch must carry metadata.resourceVersion as a concurrency guard")
	assert.Equal(t, "42", meta["resourceVersion"])
}

// A 409 from the resourceVersion guard must be retried on fresh data, even if
// the conditions themselves did not change.
func TestReconcileTargetConflictsRetriesOnResourceVersionConflict(t *testing.T) {
	older := conflictVpa{name: "vpa-1", age: 2 * time.Hour}
	newer := conflictVpa{name: "vpa-2", age: time.Hour}
	client := vpa_fake.NewSimpleClientset(older.build(), newer.build())

	listedNewer := newer.build()
	listedNewer.ResourceVersion = "42"
	listed := []*vpa_types.VerticalPodAutoscaler{older.build(), listedNewer}

	calls := 0
	client.PrependReactor("patch", "verticalpodautoscalers", func(k8stesting.Action) (bool, runtime.Object, error) {
		calls++
		if calls == 1 {
			return true, nil, apierrors.NewConflict(vpaGVR.GroupResource(), "vpa-2", fmt.Errorf("stale resourceVersion"))
		}
		return false, nil, nil
	})

	u := &updater{vpaClient: client, eventRecorder: record.NewFakeRecorder(20)}
	u.reconcileTargetConflicts(listed)

	assert.Equal(t, 2, statusPatchCount(client), "first patch conflicts, the retry succeeds")
	tc := findCondition(getVpa(t, client, "default", "vpa-2"), vpa_types.TargetConflict)
	require.NotNil(t, tc)
	assert.Equal(t, corev1.ConditionTrue, tc.Status)
}
