/*
Copyright 2025 The Kubernetes Authors.

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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_fake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func deploymentRef(name string) *autoscalingv1.CrossVersionObjectReference {
	return &autoscalingv1.CrossVersionObjectReference{
		Kind: "Deployment",
		Name: name,
	}
}

func newConflictTestUpdater(vpaClient *vpa_fake.Clientset, ignoredNamespaces ...string) *updater {
	return &updater{
		vpaClient:         vpaClient,
		eventRecorder:     record.NewFakeRecorder(10),
		ignoredNamespaces: ignoredNamespaces,
	}
}

func getVpa(t *testing.T, client *vpa_fake.Clientset, namespace, name string) *vpa_types.VerticalPodAutoscaler {
	t.Helper()
	vpa, err := client.AutoscalingV1().VerticalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err)
	return vpa
}

func findCondition(vpa *vpa_types.VerticalPodAutoscaler, condType vpa_types.VerticalPodAutoscalerConditionType) *vpa_types.VerticalPodAutoscalerCondition {
	for i := range vpa.Status.Conditions {
		if vpa.Status.Conditions[i].Type == condType {
			return &vpa.Status.Conditions[i]
		}
	}
	return nil
}

func TestReconcileTargetConflicts_TwoActiveVPAsConflict(t *testing.T) {
	vpa1 := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).WithTargetRef(deploymentRef("app")).Get()
	vpa2 := test.VerticalPodAutoscaler().WithName("vpa-2").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeInPlaceOrRecreate).WithTargetRef(deploymentRef("app")).Get()

	client := vpa_fake.NewSimpleClientset(vpa1, vpa2)
	u := newConflictTestUpdater(client)

	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpa1, vpa2})

	for _, name := range []string{"vpa-1", "vpa-2"} {
		got := getVpa(t, client, "default", name)
		cond := findCondition(got, vpa_types.TargetConflict)
		require.NotNil(t, cond, "expected TargetConflict condition on %s", name)
		assert.Equal(t, corev1.ConditionTrue, cond.Status)
		assert.Equal(t, targetConflictReason, cond.Reason)
		assert.Contains(t, cond.Message, "vpa-1")
		assert.Contains(t, cond.Message, "vpa-2")
	}
}

func TestReconcileTargetConflicts_InitialModeIncluded(t *testing.T) {
	vpa1 := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeInitial).WithTargetRef(deploymentRef("app")).Get()
	vpa2 := test.VerticalPodAutoscaler().WithName("vpa-2").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeInitial).WithTargetRef(deploymentRef("app")).Get()

	client := vpa_fake.NewSimpleClientset(vpa1, vpa2)
	u := newConflictTestUpdater(client)

	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpa1, vpa2})

	for _, name := range []string{"vpa-1", "vpa-2"} {
		got := getVpa(t, client, "default", name)
		cond := findCondition(got, vpa_types.TargetConflict)
		require.NotNil(t, cond, "Initial-mode VPAs should still be flagged per maintainer decision")
		assert.Equal(t, corev1.ConditionTrue, cond.Status)
	}
}

func TestReconcileTargetConflicts_OffModeExcluded(t *testing.T) {
	vpa1 := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeOff).WithTargetRef(deploymentRef("app")).Get()
	vpa2 := test.VerticalPodAutoscaler().WithName("vpa-2").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).WithTargetRef(deploymentRef("app")).Get()

	client := vpa_fake.NewSimpleClientset(vpa1, vpa2)
	u := newConflictTestUpdater(client)

	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpa1, vpa2})

	got1 := getVpa(t, client, "default", "vpa-1")
	assert.Nil(t, findCondition(got1, vpa_types.TargetConflict), "Off-mode VPA should not be flagged")

	got2 := getVpa(t, client, "default", "vpa-2")
	assert.Nil(t, findCondition(got2, vpa_types.TargetConflict), "single active VPA on a target should not be flagged")
}

func TestReconcileTargetConflicts_NoConflictSingleVPA(t *testing.T) {
	vpa1 := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).WithTargetRef(deploymentRef("app")).Get()

	client := vpa_fake.NewSimpleClientset(vpa1)
	u := newConflictTestUpdater(client)

	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpa1})

	got := getVpa(t, client, "default", "vpa-1")
	assert.Nil(t, findCondition(got, vpa_types.TargetConflict), "single VPA should not get a TargetConflict condition")
}

func TestReconcileTargetConflicts_NamespaceIsolation(t *testing.T) {
	vpa1 := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("ns-a").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).WithTargetRef(deploymentRef("app")).Get()
	vpa2 := test.VerticalPodAutoscaler().WithName("vpa-2").WithNamespace("ns-b").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).WithTargetRef(deploymentRef("app")).Get()

	client := vpa_fake.NewSimpleClientset(vpa1, vpa2)
	u := newConflictTestUpdater(client)

	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpa1, vpa2})

	got1 := getVpa(t, client, "ns-a", "vpa-1")
	assert.Nil(t, findCondition(got1, vpa_types.TargetConflict), "same name/kind in different namespaces should not conflict")

	got2 := getVpa(t, client, "ns-b", "vpa-2")
	assert.Nil(t, findCondition(got2, vpa_types.TargetConflict))
}

func TestReconcileTargetConflicts_IgnoredNamespaceRespected(t *testing.T) {
	vpa1 := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("ignored-ns").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).WithTargetRef(deploymentRef("app")).Get()
	vpa2 := test.VerticalPodAutoscaler().WithName("vpa-2").WithNamespace("ignored-ns").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeInPlaceOrRecreate).WithTargetRef(deploymentRef("app")).Get()

	client := vpa_fake.NewSimpleClientset(vpa1, vpa2)
	u := newConflictTestUpdater(client, "ignored-ns")

	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpa1, vpa2})

	got1 := getVpa(t, client, "ignored-ns", "vpa-1")
	assert.Nil(t, findCondition(got1, vpa_types.TargetConflict), "VPAs in ignored namespaces should be skipped")
}

func TestReconcileTargetConflicts_ClearsOnResolution(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	// vpa1 previously had a recorded conflict; vpa2 (the other party) is gone now.
	vpa1 := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).WithTargetRef(deploymentRef("app")).
		AppendCondition(vpa_types.TargetConflict, corev1.ConditionTrue, targetConflictReason,
			"Conflict: multiple active VPAs target the same object: vpa-1, vpa-2", past).Get()

	client := vpa_fake.NewSimpleClientset(vpa1)
	u := newConflictTestUpdater(client)

	// Only vpa1 remains active on this target now.
	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpa1})

	got := getVpa(t, client, "default", "vpa-1")
	cond := findCondition(got, vpa_types.TargetConflict)
	require.NotNil(t, cond)
	assert.Equal(t, corev1.ConditionFalse, cond.Status)
	assert.Equal(t, noTargetConflictReason, cond.Reason)
	assert.True(t, cond.LastTransitionTime.After(past), "LastTransitionTime should be bumped on status change")
}

func TestReconcileTargetConflicts_NoEventOnRepeatedConflict(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	vpa1 := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).WithTargetRef(deploymentRef("app")).
		AppendCondition(vpa_types.TargetConflict, corev1.ConditionTrue, targetConflictReason,
			"Conflict: multiple active VPAs target the same object: vpa-1, vpa-2", past).Get()
	vpa2 := test.VerticalPodAutoscaler().WithName("vpa-2").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeInPlaceOrRecreate).WithTargetRef(deploymentRef("app")).
		AppendCondition(vpa_types.TargetConflict, corev1.ConditionTrue, targetConflictReason,
			"Conflict: multiple active VPAs target the same object: vpa-1, vpa-2", past).Get()

	client := vpa_fake.NewSimpleClientset(vpa1, vpa2)
	recorder := record.NewFakeRecorder(10)
	u := &updater{vpaClient: client, eventRecorder: recorder}

	// Conflict still active on this run -- should not re-fire an event.
	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpa1, vpa2})

	select {
	case ev := <-recorder.Events:
		t.Fatalf("expected no event on unchanged conflict state, got: %s", ev)
	default:
		// expected: no event
	}
}

func TestReconcileTargetConflicts_ClearsOnIneligibleTransition(t *testing.T) {
	past := time.Now().Add(-time.Hour)

	// vpa1 flips from Recreate to Off. It was previously flagged as
	// conflicting and should have that condition cleared even though it no
	// longer participates in grouping.
	vpaTurnedOff := test.VerticalPodAutoscaler().WithName("vpa-1").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeOff).WithTargetRef(deploymentRef("app")).
		AppendCondition(vpa_types.TargetConflict, corev1.ConditionTrue, targetConflictReason,
			"Conflict: multiple active VPAs target the same object: vpa-1, vpa-2", past).Get()

	// vpa2 drops its targetRef entirely. Same expectation: stale condition
	// should be cleared.
	vpaNoTargetRef := test.VerticalPodAutoscaler().WithName("vpa-2").WithNamespace("default").WithContainer("main").
		WithUpdateMode(vpa_types.UpdateModeRecreate).
		AppendCondition(vpa_types.TargetConflict, corev1.ConditionTrue, targetConflictReason,
			"Conflict: multiple active VPAs target the same object: vpa-1, vpa-2", past).Get()
	vpaNoTargetRef.Spec.TargetRef = nil

	client := vpa_fake.NewSimpleClientset(vpaTurnedOff, vpaNoTargetRef)
	u := newConflictTestUpdater(client)

	u.reconcileTargetConflicts([]*vpa_types.VerticalPodAutoscaler{vpaTurnedOff, vpaNoTargetRef})

	got1 := getVpa(t, client, "default", "vpa-1")
	cond1 := findCondition(got1, vpa_types.TargetConflict)
	require.NotNil(t, cond1, "Off-mode VPA should still have its stale condition updated, not left dangling")
	assert.Equal(t, corev1.ConditionFalse, cond1.Status, "condition should be cleared once VPA turns Off")
	assert.Equal(t, noTargetConflictReason, cond1.Reason)

	got2 := getVpa(t, client, "default", "vpa-2")
	cond2 := findCondition(got2, vpa_types.TargetConflict)
	require.NotNil(t, cond2, "VPA with no targetRef should still have its stale condition updated, not left dangling")
	assert.Equal(t, corev1.ConditionFalse, cond2.Status, "condition should be cleared once targetRef is removed")
}
