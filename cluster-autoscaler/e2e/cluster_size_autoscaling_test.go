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

package e2e

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"
	"testing/synctest"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestWaitForStableReadySchedulableNodeCount(t *testing.T) {
	readyNode := func(name string) v1.Node {
		return v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{{Type: v1.NodeReady, Status: v1.ConditionTrue}},
			},
		}
	}
	baseline := []v1.Node{readyNode("node-1"), readyNode("node-2"), readyNode("node-3"), readyNode("control-plane")}
	baseline[0].Spec.Taints = []v1.Taint{{Key: "DeletionCandidateOfClusterAutoscaler", Effect: v1.TaintEffectPreferNoSchedule}}
	baseline[3].Spec.Taints = []v1.Taint{{Key: "node-role.kubernetes.io/control-plane", Effect: v1.TaintEffectNoSchedule}}

	deleting := readyNode("deleting-node")
	deleting.Spec.Unschedulable = true
	deleting.Spec.Taints = []v1.Taint{{Key: "ToBeDeletedByClusterAutoscaler", Effect: v1.TaintEffectNoSchedule}}
	otherDeleting := *deleting.DeepCopy()
	otherDeleting.Name = "other-deleting-node"
	uncordonedDeleting := *deleting.DeepCopy()
	uncordonedDeleting.Spec.Unschedulable = false
	terminating := readyNode("terminating-node")
	terminating.Spec.Unschedulable = true
	deletionTime := metav1.NewTime(time.Unix(1, 0))
	terminating.DeletionTimestamp = &deletionTime
	extra := readyNode("extra-node")
	disabled := readyNode("disabled-node")
	disabled.Spec.Taints = []v1.Taint{{Key: disabledTaint, Effect: v1.TaintEffectNoSchedule}}
	unready := readyNode("unready-node")
	unready.Status.Conditions[0].Status = v1.ConditionFalse
	networkUnavailable := readyNode("network-unavailable-node")
	networkUnavailable.Status.Conditions = append(networkUnavailable.Status.Conditions, v1.NodeCondition{
		Type: v1.NodeNetworkUnavailable, Status: v1.ConditionTrue,
	})
	cordoned := readyNode("cordoned-node")
	cordoned.Spec.Unschedulable = true
	noExecute := readyNode("no-execute-node")
	noExecute.Spec.Taints = []v1.Taint{{Key: "example.com/blocked", Effect: v1.TaintEffectNoExecute}}

	type nodeState struct {
		after      time.Duration
		extraNodes []v1.Node
		listErr    error
	}
	for _, tc := range []struct {
		name         string
		states       []nodeState
		cancelAfter  time.Duration
		wantDuration time.Duration
		wantErr      error
	}{
		{
			name:         "stable baseline ignores control-plane and soft deletion taints",
			wantDuration: nodeCountStableFor,
		},
		{
			name:         "unrelated blocking taints do not indicate deletion",
			states:       []nodeState{{extraNodes: []v1.Node{disabled}}},
			wantDuration: nodeCountStableFor,
		},
		{
			name:         "excludes unready and unschedulable nodes from the baseline",
			states:       []nodeState{{extraNodes: []v1.Node{unready, networkUnavailable, cordoned, noExecute}}},
			wantDuration: nodeCountStableFor,
		},
		{
			name: "waits for cordoned nodes to finish deletion",
			states: []nodeState{
				{extraNodes: []v1.Node{deleting}},
				{after: 40 * time.Second},
			},
			wantDuration: 40*time.Second + nodeCountStableFor,
		},
		{
			name: "waits for deletion taints without cordoning",
			states: []nodeState{
				{extraNodes: []v1.Node{uncordonedDeleting}},
				{after: 40 * time.Second},
			},
			wantDuration: 40*time.Second + nodeCountStableFor,
		},
		{
			name: "waits for terminating nodes",
			states: []nodeState{
				{extraNodes: []v1.Node{terminating}},
				{after: 40 * time.Second},
			},
			wantDuration: 40*time.Second + nodeCountStableFor,
		},
		{
			name: "resets stability when a node starts deleting",
			states: []nodeState{
				{after: 15 * time.Second, extraNodes: []v1.Node{deleting}},
				{after: 40 * time.Second},
			},
			wantDuration: 40*time.Second + nodeCountStableFor,
		},
		{
			name: "waits through autoscaler restart uncordoning leftover nodes",
			states: []nodeState{
				{extraNodes: []v1.Node{deleting, otherDeleting}},
				{after: 30 * time.Second, extraNodes: []v1.Node{readyNode(deleting.Name), readyNode(otherDeleting.Name)}},
				{after: 60 * time.Second},
			},
			wantDuration: 60*time.Second + nodeCountStableFor,
		},
		{
			name: "resets stability when the ready node count changes",
			states: []nodeState{
				{after: 15 * time.Second, extraNodes: []v1.Node{extra}},
				{after: 40 * time.Second},
			},
			wantDuration: 40*time.Second + nodeCountStableFor,
		},
		{
			name: "resets stability after a list error",
			states: []nodeState{
				{after: 15 * time.Second, listErr: errors.New("temporary node list failure")},
				{after: 20 * time.Second},
			},
			wantDuration: 20*time.Second + nodeCountStableFor,
		},
		{
			name:         "times out if deletion never finishes",
			states:       []nodeState{{extraNodes: []v1.Node{deleting}}},
			wantDuration: 2 * time.Minute,
			wantErr:      context.DeadlineExceeded,
		},
		{
			name:         "honors context cancellation",
			states:       []nodeState{{extraNodes: []v1.Node{deleting}}},
			cancelAfter:  10 * time.Second,
			wantDuration: 10 * time.Second,
			wantErr:      context.Canceled,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				if tc.cancelAfter > 0 {
					go func() {
						time.Sleep(tc.cancelAfter)
						cancel()
					}()
				}

				start := time.Now()
				client := fake.NewClientset()
				client.PrependReactor("list", "nodes", func(action clienttesting.Action) (bool, runtime.Object, error) {
					var state nodeState
					for _, candidate := range tc.states {
						if time.Since(start) >= candidate.after {
							state = candidate
						}
					}
					if state.listErr != nil {
						return true, nil, state.listErr
					}
					selector := action.(clienttesting.ListAction).GetListRestrictions().Fields
					nodes := &v1.NodeList{}
					for _, node := range append(slices.Clone(baseline), state.extraNodes...) {
						if selector.Matches(fields.Set{
							"metadata.name":      node.Name,
							"spec.unschedulable": strconv.FormatBool(node.Spec.Unschedulable),
						}) {
							nodes.Items = append(nodes.Items, *node.DeepCopy())
						}
					}
					return true, nodes, nil
				})

				nodes, err := waitForStableReadySchedulableNodeCount(ctx, client, 3, 2*time.Minute)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("waitForStableReadySchedulableNodeCount() error = %v, want %v", err, tc.wantErr)
				}
				if elapsed := time.Since(start); elapsed < tc.wantDuration || elapsed > tc.wantDuration+nodeCountPollInterval {
					t.Errorf("returned after %v, want between %v and %v", elapsed, tc.wantDuration, tc.wantDuration+nodeCountPollInterval)
				}
				if err != nil {
					return
				}
				var names []string
				for _, node := range nodes.Items {
					names = append(names, node.Name)
				}
				if want := []string{"node-1", "node-2", "node-3"}; !slices.Equal(names, want) {
					t.Errorf("ready schedulable nodes = %v, want %v", names, want)
				}
			})
		})
	}
}
