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
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_fake "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/fake"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

var conflictTestBase = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

func deploymentRef(name string) *autoscalingv1.CrossVersionObjectReference {
	return &autoscalingv1.CrossVersionObjectReference{
		Kind: "Deployment",
		Name: name,
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

// conflictVpa describes a VPA used in the tests. All VPAs target Deployment "app".
type conflictVpa struct {
	name      string
	namespace string               // defaults to "default"
	mode      vpa_types.UpdateMode // defaults to Recreate
	age       time.Duration        // larger age means created earlier
	policies  []vpa_types.ContainerResourcePolicy
	// noTargetRef drops the targetRef.
	noTargetRef bool
	// startupBoost sets a startupBoost config.
	startupBoost bool
	// priorMessage, if set, gives the VPA an existing TargetConflict=True
	// condition with this message.
	priorMessage string
}

func (c conflictVpa) ns() string {
	if c.namespace == "" {
		return "default"
	}
	return c.namespace
}

func (c conflictVpa) build() *vpa_types.VerticalPodAutoscaler {
	mode := c.mode
	if mode == "" {
		mode = vpa_types.UpdateModeRecreate
	}
	b := test.VerticalPodAutoscaler().WithName(c.name).WithNamespace(c.ns()).WithContainer("main").
		WithUpdateMode(mode).WithTargetRef(deploymentRef("app"))
	if c.priorMessage != "" {
		b = b.AppendCondition(vpa_types.TargetConflict, corev1.ConditionTrue, targetConflictReason,
			c.priorMessage, time.Now().Add(-time.Hour))
	}
	v := b.Get()
	v.CreationTimestamp = metav1.NewTime(conflictTestBase.Add(-c.age))
	if c.noTargetRef {
		v.Spec.TargetRef = nil
	}
	if c.startupBoost {
		v.Spec.StartupBoost = &vpa_types.StartupBoost{}
	}
	if c.policies != nil {
		v.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{ContainerPolicies: c.policies}
	}
	return v
}

// runReconcile creates the VPAs in a fake client and runs reconcileTargetConflicts
// on them, optionally passing them in reverse order.
func runReconcile(vpas []conflictVpa, reverse bool, ignoredNamespaces ...string) (*vpa_fake.Clientset, *record.FakeRecorder) {
	built := make([]*vpa_types.VerticalPodAutoscaler, 0, len(vpas))
	objs := make([]runtime.Object, 0, len(vpas))
	for _, c := range vpas {
		v := c.build()
		built = append(built, v)
		objs = append(objs, v)
	}
	if reverse {
		for i, j := 0, len(built)-1; i < j; i, j = i+1, j-1 {
			built[i], built[j] = built[j], built[i]
		}
	}
	client := vpa_fake.NewSimpleClientset(objs...)
	recorder := record.NewFakeRecorder(20)
	u := &updater{vpaClient: client, eventRecorder: recorder, ignoredNamespaces: ignoredNamespaces}
	u.reconcileTargetConflicts(built)
	return client, recorder
}

func TestReconcileTargetConflicts(t *testing.T) {
	cpu, mem := corev1.ResourceCPU, corev1.ResourceMemory
	cpuOnly := []vpa_types.ContainerResourcePolicy{{ContainerName: "*", ControlledResources: &[]corev1.ResourceName{cpu}}}
	memOnly := []vpa_types.ContainerResourcePolicy{{ContainerName: "*", ControlledResources: &[]corev1.ResourceName{mem}}}

	tests := []struct {
		name              string
		vpas              []conflictVpa
		ignoredNamespaces []string
		// wantConflict maps a VPA name to the name of the VPA that controls
		// the target instead of it. VPAs not listed must not be in conflict:
		// the condition is absent, or False if they had one before.
		wantConflict map[string]string
	}{
		{
			name:         "older VPA controls, newer one conflicts",
			vpas:         []conflictVpa{{name: "vpa-1", age: 2 * time.Hour}, {name: "vpa-2", age: time.Hour}},
			wantConflict: map[string]string{"vpa-2": "vpa-1"},
		},
		{
			name:         "creation time wins over name",
			vpas:         []conflictVpa{{name: "vpa-a", age: time.Hour}, {name: "vpa-b", age: 2 * time.Hour}},
			wantConflict: map[string]string{"vpa-a": "vpa-b"},
		},
		{
			name:         "same creation time, lower name controls",
			vpas:         []conflictVpa{{name: "vpa-1"}, {name: "vpa-2"}},
			wantConflict: map[string]string{"vpa-2": "vpa-1"},
		},
		{
			name: "three VPAs, only the oldest controls",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 3 * time.Hour},
				{name: "vpa-2", age: 2 * time.Hour},
				{name: "vpa-3", age: time.Hour},
			},
			wantConflict: map[string]string{"vpa-2": "vpa-1", "vpa-3": "vpa-1"},
		},
		{
			name: "Initial mode counts as active",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour, mode: vpa_types.UpdateModeInitial},
				{name: "vpa-2", age: time.Hour, mode: vpa_types.UpdateModeInitial},
			},
			wantConflict: map[string]string{"vpa-2": "vpa-1"},
		},
		{
			name: "Off VPA is ignored",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour, mode: vpa_types.UpdateModeOff},
				{name: "vpa-2", age: time.Hour},
			},
		},
		{
			name: "Off VPA with startupBoost is active",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour, mode: vpa_types.UpdateModeOff, startupBoost: true},
				{name: "vpa-2", age: time.Hour},
			},
			wantConflict: map[string]string{"vpa-2": "vpa-1"},
		},
		{
			name: "single VPA",
			vpas: []conflictVpa{{name: "vpa-1"}},
		},
		{
			name: "same target name in different namespaces",
			vpas: []conflictVpa{
				{name: "vpa-1", namespace: "ns-a", age: 2 * time.Hour},
				{name: "vpa-2", namespace: "ns-b", age: time.Hour},
			},
		},
		{
			name: "ignored namespace is skipped",
			vpas: []conflictVpa{
				{name: "vpa-1", namespace: "ignored-ns", age: 2 * time.Hour},
				{name: "vpa-2", namespace: "ignored-ns", age: time.Hour},
			},
			ignoredNamespaces: []string{"ignored-ns"},
		},
		{
			// Stronger() picks one VPA per pod regardless of the resources
			// each VPA controls, so disjoint resources still conflict.
			name: "different controlled resources still conflict",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour, policies: cpuOnly},
				{name: "vpa-2", age: time.Hour, policies: memOnly},
			},
			wantConflict: map[string]string{"vpa-2": "vpa-1"},
		},
		{
			name: "condition cleared when the other VPA is gone",
			vpas: []conflictVpa{{name: "vpa-1", priorMessage: "old"}},
		},
		{
			name: "controlling VPA's stale condition is cleared",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour, priorMessage: "old"},
				{name: "vpa-2", age: time.Hour},
			},
			wantConflict: map[string]string{"vpa-2": "vpa-1"},
		},
		{
			name: "condition cleared when a VPA turns Off or loses its targetRef",
			vpas: []conflictVpa{
				{name: "vpa-1", mode: vpa_types.UpdateModeOff, priorMessage: "old"},
				{name: "vpa-2", noTargetRef: true, priorMessage: "old"},
			},
		},
	}

	for _, tc := range tests {
		for _, reverse := range []bool{false, true} {
			name := tc.name
			if reverse {
				name += " (reversed input)"
			}
			t.Run(name, func(t *testing.T) {
				client, _ := runReconcile(tc.vpas, reverse, tc.ignoredNamespaces...)
				for _, c := range tc.vpas {
					got := getVpa(t, client, c.ns(), c.name)
					cond := findCondition(got, vpa_types.TargetConflict)
					if controlling, ok := tc.wantConflict[c.name]; ok {
						require.NotNil(t, cond, "expected TargetConflict condition on %s", c.name)
						assert.Equal(t, corev1.ConditionTrue, cond.Status)
						assert.Equal(t, targetConflictReason, cond.Reason)
						assert.Contains(t, cond.Message, controlling)
						assert.Equal(t, got.Generation, cond.ObservedGeneration)
						continue
					}
					if c.priorMessage != "" {
						require.NotNil(t, cond, "stale condition on %s should be updated, not left dangling", c.name)
						assert.Equal(t, corev1.ConditionFalse, cond.Status)
						assert.Equal(t, noTargetConflictReason, cond.Reason)
						assert.True(t, cond.LastTransitionTime.After(time.Now().Add(-10*time.Minute)),
							"LastTransitionTime should be bumped when the status changes")
						continue
					}
					assert.Nil(t, cond, "%s should not have a TargetConflict condition", c.name)
				}
			})
		}
	}
}

func TestReconcileTargetConflictsEvents(t *testing.T) {
	warning := "Warning " + targetConflictReason
	normal := "Normal " + noTargetConflictReason

	tests := []struct {
		name       string
		vpas       []conflictVpa
		wantEvents []string
	}{
		{
			name:       "new conflict emits one warning",
			vpas:       []conflictVpa{{name: "vpa-1", age: 2 * time.Hour}, {name: "vpa-2", age: time.Hour}},
			wantEvents: []string{warning},
		},
		{
			name: "persisting conflict emits nothing",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour},
				{name: "vpa-2", age: time.Hour, priorMessage: `Conflict: VPA "vpa-1" controls it`},
			},
		},
		{
			name: "changed controlling VPA updates the message but emits nothing",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 2 * time.Hour},
				{name: "vpa-2", age: time.Hour, priorMessage: `Conflict: VPA "vpa-9" controls it`},
			},
		},
		{
			name: "third VPA joining emits one warning, for the new one",
			vpas: []conflictVpa{
				{name: "vpa-1", age: 3 * time.Hour},
				{name: "vpa-2", age: 2 * time.Hour, priorMessage: `Conflict: VPA "vpa-1" controls it`},
				{name: "vpa-3", age: time.Hour},
			},
			wantEvents: []string{warning},
		},
		{
			name:       "resolved conflict emits a normal event",
			vpas:       []conflictVpa{{name: "vpa-1", priorMessage: "old"}},
			wantEvents: []string{normal},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, recorder := runReconcile(tc.vpas, false)
			var got []string
			for len(recorder.Events) > 0 {
				f := strings.Fields(<-recorder.Events)
				got = append(got, f[0]+" "+f[1])
			}
			assert.ElementsMatch(t, tc.wantEvents, got)
		})
	}
}
