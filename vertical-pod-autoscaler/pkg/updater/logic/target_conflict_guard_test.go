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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_fake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
)

var vpaGVR = schema.GroupVersionResource{Group: "autoscaling.k8s.io", Version: "v1", Resource: "verticalpodautoscalers"}

// appendStoredCondition simulates another writer (e.g. the recommender) adding
// a condition to the VPA stored in the API server, behind the updater's back.
func appendStoredCondition(t *testing.T, client *vpa_fake.Clientset, name, reason string) {
	t.Helper()
	obj, err := client.Tracker().Get(vpaGVR, "default", name)
	require.NoError(t, err)
	stored := obj.(*vpa_types.VerticalPodAutoscaler).DeepCopy()
	stored.Status.Conditions = append(stored.Status.Conditions, vpa_types.VerticalPodAutoscalerCondition{
		Type:   vpa_types.RecommendationProvided,
		Status: corev1.ConditionTrue,
		Reason: reason,
	})
	require.NoError(t, client.Tracker().Update(vpaGVR, stored, "default"))
}

func statusPatchCount(client *vpa_fake.Clientset) int {
	n := 0
	for _, a := range client.Actions() {
		if a.GetVerb() == "patch" && a.GetSubresource() == "status" {
			n++
		}
	}
	return n
}

// vpa-1 controls the target; vpa-2 already carries an outdated TargetConflict
// condition (so its patch is guarded) and is listed before the concurrent write.
func concurrentWriteSetup() (older, newer conflictVpa) {
	older = conflictVpa{name: "vpa-1", age: 2 * time.Hour}
	newer = conflictVpa{name: "vpa-2", age: time.Hour, priorMessage: `Conflict: VPA "vpa-9" controls it`}
	return older, newer
}

func TestReconcileTargetConflictsKeepsConcurrentConditionWrite(t *testing.T) {
	older, newer := concurrentWriteSetup()
	client := vpa_fake.NewSimpleClientset(older.build(), newer.build())
	// What the updater listed, before the other writer got in.
	listed := []*vpa_types.VerticalPodAutoscaler{older.build(), newer.build()}
	appendStoredCondition(t, client, "vpa-2", "concurrent")
	client.ClearActions()

	u := &updater{vpaClient: client, eventRecorder: record.NewFakeRecorder(20)}
	u.reconcileTargetConflicts(listed)

	got := getVpa(t, client, "default", "vpa-2")
	tc := findCondition(got, vpa_types.TargetConflict)
	require.NotNil(t, tc)
	assert.Equal(t, corev1.ConditionTrue, tc.Status)
	assert.Contains(t, tc.Message, "vpa-1", "message should be recomputed and name the controlling VPA")
	rp := findCondition(got, vpa_types.RecommendationProvided)
	require.NotNil(t, rp, "the concurrently written condition must not be overwritten")
	assert.Equal(t, "concurrent", rp.Reason)
	assert.Equal(t, 2, statusPatchCount(client), "first patch is rejected by the guard, the retry succeeds")
}

func TestReconcileTargetConflictsGivesUpWhenConditionsKeepChanging(t *testing.T) {
	older, newer := concurrentWriteSetup()
	client := vpa_fake.NewSimpleClientset(older.build(), newer.build())
	listed := []*vpa_types.VerticalPodAutoscaler{older.build(), newer.build()}
	recorder := record.NewFakeRecorder(20)

	// Every patch attempt is preceded by yet another concurrent write.
	writes := 0
	client.PrependReactor("patch", "verticalpodautoscalers", func(k8stesting.Action) (bool, runtime.Object, error) {
		writes++
		appendStoredCondition(t, client, "vpa-2", fmt.Sprintf("concurrent-%d", writes))
		return false, nil, nil
	})
	client.ClearActions()

	u := &updater{vpaClient: client, eventRecorder: recorder}
	u.reconcileTargetConflicts(listed)

	assert.Equal(t, targetConflictPatchAttempts, statusPatchCount(client), "retries must be bounded")
	tc := findCondition(getVpa(t, client, "default", "vpa-2"), vpa_types.TargetConflict)
	require.NotNil(t, tc)
	assert.Contains(t, tc.Message, "vpa-9", "nothing was written, the old condition is left as is")
	assert.Empty(t, recorder.Events, "no event for a persisting conflict")
}

func TestReconcileTargetConflictsStopsRetryWhenGenerationChanged(t *testing.T) {
	older, newer := concurrentWriteSetup()
	client := vpa_fake.NewSimpleClientset(older.build(), newer.build())
	listed := []*vpa_types.VerticalPodAutoscaler{older.build(), newer.build()}
	recorder := record.NewFakeRecorder(20)

	// Another writer adds a condition and the VPA spec changes (generation bump).
	appendStoredCondition(t, client, "vpa-2", "concurrent")
	obj, err := client.Tracker().Get(vpaGVR, "default", "vpa-2")
	require.NoError(t, err)
	stored := obj.(*vpa_types.VerticalPodAutoscaler).DeepCopy()
	stored.Generation = 2
	require.NoError(t, client.Tracker().Update(vpaGVR, stored, "default"))
	client.ClearActions()

	u := &updater{vpaClient: client, eventRecorder: recorder}
	u.reconcileTargetConflicts(listed)

	assert.Equal(t, 1, statusPatchCount(client), "no retry once the generation changed")
	tc := findCondition(getVpa(t, client, "default", "vpa-2"), vpa_types.TargetConflict)
	require.NotNil(t, tc)
	assert.Contains(t, tc.Message, "vpa-9", "nothing was written for a stale decision")
	assert.Empty(t, recorder.Events)
}
