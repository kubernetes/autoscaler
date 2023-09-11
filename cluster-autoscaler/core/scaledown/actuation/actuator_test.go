/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/budgets"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/utilization"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

type startDeletionTestCase struct {
	emptyNodes            []*budgets.NodeGroupView
	drainNodes            []*budgets.NodeGroupView
	pods                  map[string][]*apiv1.Pod
	failedPodDrain        map[string]bool
	failedNodeDeletion    map[string]bool
	failedNodeTaint       map[string]bool
	wantStatus            *status.ScaleDownStatus
	wantErr               error
	wantDeletedPods       []string
	wantDeletedNodes      []string
	wantTaintUpdates      map[string][][]apiv1.Taint
	wantNodeDeleteResults map[string]status.NodeDeleteResult
}

func getStartDeletionTestCases(testNg *testprovider.TestNodeGroup, ignoreDaemonSetsUtilization bool, suffix string) map[string]startDeletionTestCase {
	toBeDeletedTaint := apiv1.Taint{Key: taints.ToBeDeletedTaint, Effect: apiv1.TaintEffectNoSchedule}

	dsUtilInfo := generateUtilInfo(2./8., 2./8.)

	if ignoreDaemonSetsUtilization {
		dsUtilInfo = generateUtilInfo(0./8., 0./8.)
	}

	atomic2 := sizedNodeGroup("atomic-2", 2, true)
	atomic4 := sizedNodeGroup("atomic-4", 4, true)
	// We need separate groups since previous test cases *change state of groups*.
	// TODO(aleksandra-malinowska): refactor this test to isolate test cases.
	atomic2pods := sizedNodeGroup("atomic-2-pods", 2, true)
	atomic4taints := sizedNodeGroup("atomic-4-taints", 4, true)
	atomic6 := sizedNodeGroup("atomic-6", 6, true)
	atomic2mixed := sizedNodeGroup("atomic-2-mixed", 2, true)
	atomic2drain := sizedNodeGroup("atomic-2-drain", 2, true)

	testCases := map[string]startDeletionTestCase{
		"nothing to delete": {
			emptyNodes: nil,
			drainNodes: nil,
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNoNodeDeleted,
			},
		},
		"empty node deletion": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 2),
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"empty atomic node deletion": {
			emptyNodes: generateNodeGroupViewList(atomic2, 0, 2),
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("atomic-2-node-0"),
						NodeGroup:   atomic2,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("atomic-2-node-1"),
						NodeGroup:   atomic2,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
				},
			},
			wantDeletedNodes: []string{"atomic-2-node-0", "atomic-2-node-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"atomic-2-node-0": {
					{toBeDeletedTaint},
				},
				"atomic-2-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"atomic-2-node-0": {ResultType: status.NodeDeleteOk},
				"atomic-2-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"deletion with drain": {
			drainNodes: generateNodeGroupViewList(testNg, 0, 2),
			pods: map[string][]*apiv1.Pod{
				"test-node-0": removablePods(2, "test-node-0"),
				"test-node-1": removablePods(2, "test-node-1"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "test-node-0"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "test-node-1"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-1"},
			wantDeletedPods:  []string{"test-node-0-pod-0", "test-node-0-pod-1", "test-node-1-pod-0", "test-node-1-pod-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"empty and drain deletion work correctly together": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 2),
			drainNodes: generateNodeGroupViewList(testNg, 2, 4),
			pods: map[string][]*apiv1.Pod{
				"test-node-2": removablePods(2, "test-node-2"),
				"test-node-3": removablePods(2, "test-node-3"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-2"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "test-node-2"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("test-node-3"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "test-node-3"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-1", "test-node-2", "test-node-3"},
			wantDeletedPods:  []string{"test-node-2-pod-0", "test-node-2-pod-1", "test-node-3-pod-0", "test-node-3-pod-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
				"test-node-2": {
					{toBeDeletedTaint},
				},
				"test-node-3": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteOk},
				"test-node-2": {ResultType: status.NodeDeleteOk},
				"test-node-3": {ResultType: status.NodeDeleteOk},
			},
		},
		"two atomic groups can be scaled down together": {
			emptyNodes: generateNodeGroupViewList(atomic2mixed, 1, 2),
			drainNodes: append(generateNodeGroupViewList(atomic2mixed, 0, 1),
				generateNodeGroupViewList(atomic2drain, 0, 2)...),
			pods: map[string][]*apiv1.Pod{
				"atomic-2-mixed-node-0": removablePods(2, "atomic-2-mixed-node-0"),
				"atomic-2-drain-node-0": removablePods(1, "atomic-2-drain-node-0"),
				"atomic-2-drain-node-1": removablePods(2, "atomic-2-drain-node-1"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("atomic-2-mixed-node-1"),
						NodeGroup:   atomic2mixed,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("atomic-2-mixed-node-0"),
						NodeGroup:   atomic2mixed,
						EvictedPods: removablePods(2, "atomic-2-mixed-node-0"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("atomic-2-drain-node-0"),
						NodeGroup:   atomic2drain,
						EvictedPods: removablePods(1, "atomic-2-drain-node-0"),
						UtilInfo:    generateUtilInfo(1./8., 1./8.),
					},
					{
						Node:        generateNode("atomic-2-drain-node-1"),
						NodeGroup:   atomic2drain,
						EvictedPods: removablePods(2, "atomic-2-drain-node-1"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{
				"atomic-2-mixed-node-0",
				"atomic-2-mixed-node-1",
				"atomic-2-drain-node-0",
				"atomic-2-drain-node-1",
			},
			wantDeletedPods: []string{"atomic-2-mixed-node-0-pod-0", "atomic-2-mixed-node-0-pod-1", "atomic-2-drain-node-0-pod-0", "atomic-2-drain-node-1-pod-0", "atomic-2-drain-node-1-pod-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"atomic-2-mixed-node-0": {
					{toBeDeletedTaint},
				},
				"atomic-2-mixed-node-1": {
					{toBeDeletedTaint},
				},
				"atomic-2-drain-node-0": {
					{toBeDeletedTaint},
				},
				"atomic-2-drain-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"atomic-2-mixed-node-0": {ResultType: status.NodeDeleteOk},
				"atomic-2-mixed-node-1": {ResultType: status.NodeDeleteOk},
				"atomic-2-drain-node-0": {ResultType: status.NodeDeleteOk},
				"atomic-2-drain-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"atomic empty and drain deletion work correctly together": {
			emptyNodes: generateNodeGroupViewList(atomic4, 0, 2),
			drainNodes: generateNodeGroupViewList(atomic4, 2, 4),
			pods: map[string][]*apiv1.Pod{
				"atomic-4-node-2": removablePods(2, "atomic-4-node-2"),
				"atomic-4-node-3": removablePods(2, "atomic-4-node-3"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("atomic-4-node-0"),
						NodeGroup:   atomic4,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("atomic-4-node-1"),
						NodeGroup:   atomic4,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("atomic-4-node-2"),
						NodeGroup:   atomic4,
						EvictedPods: removablePods(2, "atomic-4-node-2"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("atomic-4-node-3"),
						NodeGroup:   atomic4,
						EvictedPods: removablePods(2, "atomic-4-node-3"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"atomic-4-node-0", "atomic-4-node-1", "atomic-4-node-2", "atomic-4-node-3"},
			wantDeletedPods:  []string{"atomic-4-node-2-pod-0", "atomic-4-node-2-pod-1", "atomic-4-node-3-pod-0", "atomic-4-node-3-pod-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"atomic-4-node-0": {
					{toBeDeletedTaint},
				},
				"atomic-4-node-1": {
					{toBeDeletedTaint},
				},
				"atomic-4-node-2": {
					{toBeDeletedTaint},
				},
				"atomic-4-node-3": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"atomic-4-node-0": {ResultType: status.NodeDeleteOk},
				"atomic-4-node-1": {ResultType: status.NodeDeleteOk},
				"atomic-4-node-2": {ResultType: status.NodeDeleteOk},
				"atomic-4-node-3": {ResultType: status.NodeDeleteOk},
			},
		},
		"failure to taint empty node stops deletion and cleans already applied taints": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 4),
			drainNodes: generateNodeGroupViewList(testNg, 4, 5),
			pods: map[string][]*apiv1.Pod{
				"test-node-4": removablePods(2, "test-node-4"),
			},
			failedNodeTaint: map[string]bool{"test-node-2": true},
			wantStatus: &status.ScaleDownStatus{
				Result:          status.ScaleDownError,
				ScaledDownNodes: nil,
			},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
					{},
				},
				"test-node-1": {
					{toBeDeletedTaint},
					{},
				},
			},
			wantErr: cmpopts.AnyError,
		},
		"failure to taint empty atomic node stops deletion and cleans already applied taints": {
			emptyNodes: generateNodeGroupViewList(atomic4taints, 0, 4),
			drainNodes: generateNodeGroupViewList(testNg, 4, 5),
			pods: map[string][]*apiv1.Pod{
				"test-node-4": removablePods(2, "test-node-4"),
			},
			failedNodeTaint: map[string]bool{"atomic-4-taints-node-2": true},
			wantStatus: &status.ScaleDownStatus{
				Result:          status.ScaleDownError,
				ScaledDownNodes: nil,
			},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"atomic-4-taints-node-0": {
					{toBeDeletedTaint},
					{},
				},
				"atomic-4-taints-node-1": {
					{toBeDeletedTaint},
					{},
				},
			},
			wantErr: cmpopts.AnyError,
		},
		"failure to taint drain node stops further deletion and cleans already applied taints": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 2),
			drainNodes: generateNodeGroupViewList(testNg, 2, 6),
			pods: map[string][]*apiv1.Pod{
				"test-node-2": removablePods(2, "test-node-2"),
				"test-node-3": removablePods(2, "test-node-3"),
				"test-node-4": removablePods(2, "test-node-4"),
				"test-node-5": removablePods(2, "test-node-5"),
			},
			failedNodeTaint: map[string]bool{"test-node-2": true},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownError,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteOk},
			},
			wantErr: cmpopts.AnyError,
		},
		"failure to taint drain atomic node stops further deletion and cleans already applied taints": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 2),
			drainNodes: generateNodeGroupViewList(atomic6, 0, 6),
			pods: map[string][]*apiv1.Pod{
				"atomic-6-node-0": removablePods(2, "atomic-6-node-0"),
				"atomic-6-node-1": removablePods(2, "atomic-6-node-1"),
				"atomic-6-node-2": removablePods(2, "atomic-6-node-2"),
				"atomic-6-node-3": removablePods(2, "atomic-6-node-3"),
				"atomic-6-node-4": removablePods(2, "atomic-6-node-4"),
				"atomic-6-node-5": removablePods(2, "atomic-6-node-5"),
			},
			failedNodeTaint: map[string]bool{"atomic-6-node-2": true},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownError,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteOk},
			},
			wantErr: cmpopts.AnyError,
		},
		"nodes that failed drain are correctly reported in results": {
			drainNodes: generateNodeGroupViewList(testNg, 0, 4),
			pods: map[string][]*apiv1.Pod{
				"test-node-0": removablePods(3, "test-node-0"),
				"test-node-1": removablePods(3, "test-node-1"),
				"test-node-2": removablePods(3, "test-node-2"),
				"test-node-3": removablePods(3, "test-node-3"),
			},
			failedPodDrain: map[string]bool{
				"test-node-0-pod-0": true,
				"test-node-0-pod-1": true,
				"test-node-2-pod-1": true,
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(3, "test-node-0"),
						UtilInfo:    generateUtilInfo(3./8., 3./8.),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(3, "test-node-1"),
						UtilInfo:    generateUtilInfo(3./8., 3./8.),
					},
					{
						Node:        generateNode("test-node-2"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(3, "test-node-2"),
						UtilInfo:    generateUtilInfo(3./8., 3./8.),
					},
					{
						Node:        generateNode("test-node-3"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(3, "test-node-3"),
						UtilInfo:    generateUtilInfo(3./8., 3./8.),
					},
				},
			},
			wantDeletedNodes: []string{"test-node-1", "test-node-3"},
			wantDeletedPods: []string{
				"test-node-0-pod-2",
				"test-node-1-pod-0", "test-node-1-pod-1", "test-node-1-pod-2",
				"test-node-2-pod-0", "test-node-2-pod-2",
				"test-node-3-pod-0", "test-node-3-pod-1", "test-node-3-pod-2",
			},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
					{},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
				"test-node-2": {
					{toBeDeletedTaint},
					{},
				},
				"test-node-3": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {
					ResultType: status.NodeDeleteErrorFailedToEvictPods,
					Err:        cmpopts.AnyError,
					PodEvictionResults: map[string]status.PodEvictionResult{
						"test-node-0-pod-0": {Pod: removablePod("test-node-0-pod-0", "test-node-0"), Err: cmpopts.AnyError, TimedOut: true},
						"test-node-0-pod-1": {Pod: removablePod("test-node-0-pod-1", "test-node-0"), Err: cmpopts.AnyError, TimedOut: true},
						"test-node-0-pod-2": {Pod: removablePod("test-node-0-pod-2", "test-node-0")},
					},
				},
				"test-node-1": {ResultType: status.NodeDeleteOk},
				"test-node-2": {
					ResultType: status.NodeDeleteErrorFailedToEvictPods,
					Err:        cmpopts.AnyError,
					PodEvictionResults: map[string]status.PodEvictionResult{
						"test-node-2-pod-0": {Pod: removablePod("test-node-2-pod-0", "test-node-2")},
						"test-node-2-pod-1": {Pod: removablePod("test-node-2-pod-1", "test-node-2"), Err: cmpopts.AnyError, TimedOut: true},
						"test-node-2-pod-2": {Pod: removablePod("test-node-2-pod-2", "test-node-2")},
					},
				},
				"test-node-3": {ResultType: status.NodeDeleteOk},
			},
		},
		"nodes that failed deletion are correctly reported in results": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 2),
			drainNodes: generateNodeGroupViewList(testNg, 2, 4),
			pods: map[string][]*apiv1.Pod{
				"test-node-2": removablePods(2, "test-node-2"),
				"test-node-3": removablePods(2, "test-node-3"),
			},
			failedNodeDeletion: map[string]bool{
				"test-node-1": true,
				"test-node-3": true,
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-2"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "test-node-2"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("test-node-3"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "test-node-3"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-2"},
			wantDeletedPods: []string{
				"test-node-2-pod-0", "test-node-2-pod-1",
				"test-node-3-pod-0", "test-node-3-pod-1",
			},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
					{},
				},
				"test-node-2": {
					{toBeDeletedTaint},
				},
				"test-node-3": {
					{toBeDeletedTaint},
					{},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteErrorFailedToDelete, Err: cmpopts.AnyError},
				"test-node-2": {ResultType: status.NodeDeleteOk},
				"test-node-3": {ResultType: status.NodeDeleteErrorFailedToDelete, Err: cmpopts.AnyError},
			},
		},
		"DS pods are evicted from empty nodes, but don't block deletion on error": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 2),
			pods: map[string][]*apiv1.Pod{
				"test-node-0": generateDsPods(2, "test-node-0"),
				"test-node-1": generateDsPods(2, "test-node-1"),
			},
			failedPodDrain: map[string]bool{"test-node-1-ds-pod-0": true},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    dsUtilInfo,
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    dsUtilInfo,
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-1"},
			wantDeletedPods:  []string{"test-node-0-ds-pod-0", "test-node-0-ds-pod-1", "test-node-1-ds-pod-1"},
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"DS pods and deletion with drain": {
			drainNodes: generateNodeGroupViewList(testNg, 0, 2),
			pods: map[string][]*apiv1.Pod{
				"test-node-0": generateDsPods(2, "test-node-0"),
				"test-node-1": generateDsPods(2, "test-node-1"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:      generateNode("test-node-0"),
						NodeGroup: testNg,
						// this is nil because DaemonSetEvictionForOccupiedNodes is
						// not enabled for drained nodes in this test suite
						EvictedPods: nil,
						UtilInfo:    dsUtilInfo,
					},
					{
						Node:      generateNode("test-node-1"),
						NodeGroup: testNg,
						// this is nil because DaemonSetEvictionForOccupiedNodes is
						// not enabled for drained nodes in this test suite
						EvictedPods: nil,
						UtilInfo:    dsUtilInfo,
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-1"},
			// same as evicted pods
			wantDeletedPods: nil,
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteOk},
			},
		},
		"DS pods and empty and drain deletion work correctly together": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 2),
			drainNodes: generateNodeGroupViewList(testNg, 2, 4),
			pods: map[string][]*apiv1.Pod{
				"test-node-2": removablePods(2, "test-node-2"),
				"test-node-3": generateDsPods(2, "test-node-3"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-2"),
						NodeGroup:   testNg,
						EvictedPods: removablePods(2, "test-node-2"),
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:      generateNode("test-node-3"),
						NodeGroup: testNg,
						// this is nil because DaemonSetEvictionForOccupiedNodes is
						// not enabled for drained nodes in this test suite
						EvictedPods: nil,
						UtilInfo:    dsUtilInfo,
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0", "test-node-1", "test-node-2", "test-node-3"},
			// same as evicted pods
			wantDeletedPods: nil,
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
				},
				"test-node-2": {
					{toBeDeletedTaint},
				},
				"test-node-3": {
					{toBeDeletedTaint},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteOk},
				"test-node-1": {ResultType: status.NodeDeleteOk},
				"test-node-2": {ResultType: status.NodeDeleteOk},
				"test-node-3": {ResultType: status.NodeDeleteOk},
			},
		},
		"nodes with pods are not deleted if the node is passed as empty": {
			emptyNodes: generateNodeGroupViewList(testNg, 0, 2),
			pods: map[string][]*apiv1.Pod{
				"test-node-0": removablePods(2, "test-node-0"),
				"test-node-1": removablePods(2, "test-node-1"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: nil,
			wantDeletedPods:  nil,
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
					{},
				},
				"test-node-1": {
					{toBeDeletedTaint},
					{},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0": {ResultType: status.NodeDeleteErrorInternal, Err: cmpopts.AnyError},
				"test-node-1": {ResultType: status.NodeDeleteErrorInternal, Err: cmpopts.AnyError},
			},
		},
		"atomic nodes with pods are not deleted if the node is passed as empty": {
			emptyNodes: append(
				generateNodeGroupViewList(testNg, 0, 2),
				generateNodeGroupViewList(atomic2pods, 0, 2)...,
			),
			pods: map[string][]*apiv1.Pod{
				"test-node-1":          removablePods(2, "test-node-1"),
				"atomic-2-pods-node-1": removablePods(2, "atomic-2-pods-node-1"),
			},
			wantStatus: &status.ScaleDownStatus{
				Result: status.ScaleDownNodeDeleteStarted,
				ScaledDownNodes: []*status.ScaleDownNode{
					{
						Node:        generateNode("test-node-0"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("test-node-1"),
						NodeGroup:   testNg,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
					{
						Node:        generateNode("atomic-2-pods-node-0"),
						NodeGroup:   atomic2pods,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(0, 0),
					},
					{
						Node:        generateNode("atomic-2-pods-node-1"),
						NodeGroup:   atomic2pods,
						EvictedPods: nil,
						UtilInfo:    generateUtilInfo(2./8., 2./8.),
					},
				},
			},
			wantDeletedNodes: []string{"test-node-0"},
			wantDeletedPods:  nil,
			wantTaintUpdates: map[string][][]apiv1.Taint{
				"test-node-0": {
					{toBeDeletedTaint},
				},
				"test-node-1": {
					{toBeDeletedTaint},
					{},
				},
				"atomic-2-pods-node-0": {
					{toBeDeletedTaint},
					{},
				},
				"atomic-2-pods-node-1": {
					{toBeDeletedTaint},
					{},
				},
			},
			wantNodeDeleteResults: map[string]status.NodeDeleteResult{
				"test-node-0":          {ResultType: status.NodeDeleteOk},
				"test-node-1":          {ResultType: status.NodeDeleteErrorInternal, Err: cmpopts.AnyError},
				"atomic-2-pods-node-0": {ResultType: status.NodeDeleteErrorFailedToDelete, Err: cmpopts.AnyError},
				"atomic-2-pods-node-1": {ResultType: status.NodeDeleteErrorInternal, Err: cmpopts.AnyError},
			},
		},
	}

	testCasesWithNGNames := map[string]startDeletionTestCase{}
	for k, v := range testCases {
		testCasesWithNGNames[k+" "+suffix] = v
	}

	return testCasesWithNGNames
}

func TestStartDeletion(t *testing.T) {
	testNg1 := testprovider.NewTestNodeGroup("test", 100, 0, 3, true, false, "n1-standard-2", nil, nil)
	opts1 := &config.NodeGroupAutoscalingOptions{
		IgnoreDaemonSetsUtilization: false,
	}
	testNg1.SetOptions(opts1)
	testNg2 := testprovider.NewTestNodeGroup("test", 100, 0, 3, true, false, "n1-standard-2", nil, nil)
	opts2 := &config.NodeGroupAutoscalingOptions{
		IgnoreDaemonSetsUtilization: true,
	}
	testNg2.SetOptions(opts2)

	testSets := []map[string]startDeletionTestCase{
		// IgnoreDaemonSetsUtilization is false
		getStartDeletionTestCases(testNg1, opts1.IgnoreDaemonSetsUtilization, "testNg1"),
		// IgnoreDaemonSetsUtilization is true
		getStartDeletionTestCases(testNg2, opts2.IgnoreDaemonSetsUtilization, "testNg2"),
	}

	for _, testSet := range testSets {
		for tn, tc := range testSet {
			t.Run(tn, func(t *testing.T) {
				// This is needed because the tested code starts goroutines that can technically live longer than the execution
				// of a single test case, and the goroutines eventually access tc in fakeClient hooks below.
				tc := tc
				// Insert all nodes into a map to support live node updates and GETs.
				allEmptyNodes, allDrainNodes := []*apiv1.Node{}, []*apiv1.Node{}
				nodesByName := make(map[string]*apiv1.Node)
				nodesLock := sync.Mutex{}
				for _, bucket := range tc.emptyNodes {
					allEmptyNodes = append(allEmptyNodes, bucket.Nodes...)
					for _, node := range allEmptyNodes {
						nodesByName[node.Name] = node
					}
				}
				for _, bucket := range tc.drainNodes {
					allDrainNodes = append(allDrainNodes, bucket.Nodes...)
					for _, node := range bucket.Nodes {
						nodesByName[node.Name] = node
					}
				}

				// Set up a fake k8s client to hook and verify certain actions.
				fakeClient := &fake.Clientset{}
				type nodeTaints struct {
					nodeName string
					taints   []apiv1.Taint
				}
				taintUpdates := make(chan nodeTaints, 10)
				deletedNodes := make(chan string, 10)
				deletedPods := make(chan string, 10)

				ds := generateDaemonSet()

				// We're faking the whole k8s client, and some of the code needs to get live nodes and pods, so GET on nodes and pods has to be set up.
				fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
					nodesLock.Lock()
					defer nodesLock.Unlock()
					getAction := action.(core.GetAction)
					node, found := nodesByName[getAction.GetName()]
					if !found {
						return true, nil, fmt.Errorf("node %q not found", getAction.GetName())
					}
					return true, node, nil
				})
				fakeClient.Fake.AddReactor("get", "pods",
					func(action core.Action) (bool, runtime.Object, error) {
						return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
					})
				// Hook node update to gather all taint updates, and to fail the update for certain nodes to simulate errors.
				fakeClient.Fake.AddReactor("update", "nodes",
					func(action core.Action) (bool, runtime.Object, error) {
						nodesLock.Lock()
						defer nodesLock.Unlock()
						update := action.(core.UpdateAction)
						obj := update.GetObject().(*apiv1.Node)
						if tc.failedNodeTaint[obj.Name] {
							return true, nil, fmt.Errorf("SIMULATED ERROR: won't taint")
						}
						nt := nodeTaints{
							nodeName: obj.Name,
						}
						for _, taint := range obj.Spec.Taints {
							nt.taints = append(nt.taints, taint)
						}
						taintUpdates <- nt
						nodesByName[obj.Name] = obj.DeepCopy()
						return true, obj, nil
					})
				// Hook eviction creation to gather which pods were evicted, and to fail the eviction for certain pods to simulate errors.
				fakeClient.Fake.AddReactor("create", "pods",
					func(action core.Action) (bool, runtime.Object, error) {
						createAction := action.(core.CreateAction)
						if createAction == nil {
							return false, nil, nil
						}
						eviction := createAction.GetObject().(*policyv1beta1.Eviction)
						if eviction == nil {
							return false, nil, nil
						}
						if tc.failedPodDrain[eviction.Name] {
							return true, nil, fmt.Errorf("SIMULATED ERROR: won't evict")
						}
						deletedPods <- eviction.Name
						return true, nil, nil
					})

				// Hook node deletion at the level of cloud provider, to gather which nodes were deleted, and to fail the deletion for
				// certain nodes to simulate errors.
				provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
					if tc.failedNodeDeletion[node] {
						return fmt.Errorf("SIMULATED ERROR: won't remove node")
					}
					deletedNodes <- node
					return nil
				})
				for _, bucket := range tc.emptyNodes {
					bucket.Group.(*testprovider.TestNodeGroup).SetCloudProvider(provider)
					provider.InsertNodeGroup(bucket.Group)
					for _, node := range bucket.Nodes {
						provider.AddNode(bucket.Group.Id(), node)
					}
				}
				for _, bucket := range tc.drainNodes {
					bucket.Group.(*testprovider.TestNodeGroup).SetCloudProvider(provider)
					provider.InsertNodeGroup(bucket.Group)
					for _, node := range bucket.Nodes {
						provider.AddNode(bucket.Group.Id(), node)
					}
				}

				// Set up other needed structures and options.
				opts := config.AutoscalingOptions{
					MaxScaleDownParallelism:        10,
					MaxDrainParallelism:            5,
					MaxPodEvictionTime:             0,
					DaemonSetEvictionForEmptyNodes: true,
				}

				allPods := []*apiv1.Pod{}

				for _, pods := range tc.pods {
					allPods = append(allPods, pods...)
				}

				podLister := kube_util.NewTestPodLister(allPods)
				pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})
				dsLister, err := kube_util.NewTestDaemonSetLister([]*appsv1.DaemonSet{ds})
				if err != nil {
					t.Fatalf("Couldn't create daemonset lister")
				}

				registry := kube_util.NewListerRegistry(nil, nil, podLister, pdbLister, dsLister, nil, nil, nil, nil)
				ctx, err := NewScaleTestAutoscalingContext(opts, fakeClient, registry, provider, nil, nil)
				if err != nil {
					t.Fatalf("Couldn't set up autoscaling context: %v", err)
				}
				csr := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, ctx.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}))
				for _, bucket := range tc.emptyNodes {
					for _, node := range bucket.Nodes {
						err := ctx.ClusterSnapshot.AddNodeWithPods(node, tc.pods[node.Name])
						if err != nil {
							t.Fatalf("Couldn't add node %q to snapshot: %v", node.Name, err)
						}
					}
				}
				for _, bucket := range tc.drainNodes {
					for _, node := range bucket.Nodes {
						pods, found := tc.pods[node.Name]
						if !found {
							t.Fatalf("Drain node %q doesn't have pods defined in the test case.", node.Name)
						}
						err := ctx.ClusterSnapshot.AddNodeWithPods(node, pods)
						if err != nil {
							t.Fatalf("Couldn't add node %q to snapshot: %v", node.Name, err)
						}
					}
				}

				// Create Actuator, run StartDeletion, and verify the error.
				ndt := deletiontracker.NewNodeDeletionTracker(0)
				ndb := NewNodeDeletionBatcher(&ctx, csr, ndt, 0*time.Second)
				evictor := Evictor{EvictionRetryTime: 0, DsEvictionRetryTime: 0, DsEvictionEmptyNodeTimeout: 0, PodEvictionHeadroom: DefaultPodEvictionHeadroom}
				actuator := Actuator{
					ctx: &ctx, clusterState: csr, nodeDeletionTracker: ndt,
					nodeDeletionScheduler: NewGroupDeletionScheduler(&ctx, ndt, ndb, evictor),
					budgetProcessor:       budgets.NewScaleDownBudgetProcessor(&ctx),
					configGetter:          nodegroupconfig.NewDefaultNodeGroupConfigProcessor(ctx.NodeGroupDefaults),
				}
				gotStatus, gotErr := actuator.StartDeletion(allEmptyNodes, allDrainNodes)
				if diff := cmp.Diff(tc.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
					t.Errorf("StartDeletion error diff (-want +got):\n%s", diff)
				}

				// Verify ScaleDownStatus looks as expected.
				ignoreSdNodeOrder := cmpopts.SortSlices(func(a, b *status.ScaleDownNode) bool { return a.Node.Name < b.Node.Name })
				ignoreTimestamps := cmpopts.IgnoreFields(status.ScaleDownStatus{}, "NodeDeleteResultsAsOf")
				cmpNg := cmp.Comparer(func(a, b *testprovider.TestNodeGroup) bool { return a.Id() == b.Id() })
				statusCmpOpts := cmp.Options{ignoreSdNodeOrder, ignoreTimestamps, cmpNg, cmpopts.EquateEmpty()}
				if diff := cmp.Diff(tc.wantStatus, gotStatus, statusCmpOpts); diff != "" {
					t.Errorf("StartDeletion status diff (-want +got):\n%s", diff)
				}

				// Verify that all expected nodes were deleted using the cloud provider hook.
				var gotDeletedNodes []string
			nodesLoop:
				for i := 0; i < len(tc.wantDeletedNodes); i++ {
					select {
					case deletedNode := <-deletedNodes:
						gotDeletedNodes = append(gotDeletedNodes, deletedNode)
					case <-time.After(3 * time.Second):
						t.Errorf("Timeout while waiting for deleted nodes.")
						break nodesLoop
					}
				}
				ignoreStrOrder := cmpopts.SortSlices(func(a, b string) bool { return a < b })
				if diff := cmp.Diff(tc.wantDeletedNodes, gotDeletedNodes, ignoreStrOrder); diff != "" {
					t.Errorf("deletedNodes diff (-want +got):\n%s", diff)
				}

				// Verify that all expected pods were deleted using the fake k8s client hook.
				var gotDeletedPods []string
			podsLoop:
				for i := 0; i < len(tc.wantDeletedPods); i++ {
					select {
					case deletedPod := <-deletedPods:
						gotDeletedPods = append(gotDeletedPods, deletedPod)
					case <-time.After(3 * time.Second):
						t.Errorf("Timeout while waiting for deleted pods.")
						break podsLoop
					}
				}
				if diff := cmp.Diff(tc.wantDeletedPods, gotDeletedPods, ignoreStrOrder); diff != "" {
					t.Errorf("deletedPods diff (-want +got):\n%s", diff)
				}

				// Verify that all expected taint updates happened using the fake k8s client hook.
				allUpdatesCount := 0
				for _, updates := range tc.wantTaintUpdates {
					allUpdatesCount += len(updates)
				}
				gotTaintUpdates := make(map[string][][]apiv1.Taint)
			taintsLoop:
				for i := 0; i < allUpdatesCount; i++ {
					select {
					case taintUpdate := <-taintUpdates:
						gotTaintUpdates[taintUpdate.nodeName] = append(gotTaintUpdates[taintUpdate.nodeName], taintUpdate.taints)
					case <-time.After(3 * time.Second):
						t.Errorf("Timeout while waiting for taint updates.")
						break taintsLoop
					}
				}
				ignoreTaintValue := cmpopts.IgnoreFields(apiv1.Taint{}, "Value")
				if diff := cmp.Diff(tc.wantTaintUpdates, gotTaintUpdates, ignoreTaintValue, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("taintUpdates diff (-want +got):\n%s", diff)
				}

				// Wait for all expected deletions to be reported in NodeDeletionTracker. Reporting happens shortly after the deletion
				// in cloud provider we sync to above and so this will usually not wait at all. However, it can still happen
				// that there is a delay between cloud provider deletion and reporting, in which case the results are not there yet
				// and we need to wait for them before asserting.
				err = waitForDeletionResultsCount(actuator.nodeDeletionTracker, len(tc.wantNodeDeleteResults), 3*time.Second, 200*time.Millisecond)
				if err != nil {
					t.Errorf("Timeout while waiting for node deletion results")
				}

				// Run StartDeletion again to gather node deletion results for deletions started in the previous call, and verify
				// that they look as expected.
				gotNextStatus, gotNextErr := actuator.StartDeletion(nil, nil)
				if gotNextErr != nil {
					t.Errorf("StartDeletion unexpected error: %v", gotNextErr)
				}
				if diff := cmp.Diff(tc.wantNodeDeleteResults, gotNextStatus.NodeDeleteResults, cmpopts.EquateEmpty(), cmpopts.EquateErrors()); diff != "" {
					t.Errorf("NodeDeleteResults diff (-want +got):\n%s", diff)
				}
			})
		}
	}
}

func TestStartDeletionInBatchBasic(t *testing.T) {
	deleteInterval := 1 * time.Second

	for _, test := range []struct {
		name                   string
		deleteCalls            int
		numNodesToDelete       map[string][]int //per node group and per call
		failedRequests         map[string]bool  //per node group
		wantSuccessfulDeletion map[string]int   //per node group
	}{
		{
			name:        "Succesfull deletion for all node group",
			deleteCalls: 1,
			numNodesToDelete: map[string][]int{
				"test-ng-1": {4},
				"test-ng-2": {5},
				"test-ng-3": {1},
			},
			wantSuccessfulDeletion: map[string]int{
				"test-ng-1": 4,
				"test-ng-2": 5,
				"test-ng-3": 1,
			},
		},
		{
			name:        "Node deletion failed for one group",
			deleteCalls: 1,
			numNodesToDelete: map[string][]int{
				"test-ng-1": {4},
				"test-ng-2": {5},
				"test-ng-3": {1},
			},
			failedRequests: map[string]bool{
				"test-ng-1": true,
			},
			wantSuccessfulDeletion: map[string]int{
				"test-ng-1": 0,
				"test-ng-2": 5,
				"test-ng-3": 1,
			},
		},
		{
			name:        "Node deletion failed for one group two times",
			deleteCalls: 2,
			numNodesToDelete: map[string][]int{
				"test-ng-1": {4, 3},
				"test-ng-2": {5},
				"test-ng-3": {1},
			},
			failedRequests: map[string]bool{
				"test-ng-1": true,
			},
			wantSuccessfulDeletion: map[string]int{
				"test-ng-1": 0,
				"test-ng-2": 5,
				"test-ng-3": 1,
			},
		},
		{
			name:        "Node deletion failed for all groups",
			deleteCalls: 2,
			numNodesToDelete: map[string][]int{
				"test-ng-1": {4, 3},
				"test-ng-2": {5},
				"test-ng-3": {1},
			},
			failedRequests: map[string]bool{
				"test-ng-1": true,
				"test-ng-2": true,
				"test-ng-3": true,
			},
			wantSuccessfulDeletion: map[string]int{
				"test-ng-1": 0,
				"test-ng-2": 0,
				"test-ng-3": 0,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			test := test
			gotFailedRequest := func(nodeGroupId string) bool {
				val, _ := test.failedRequests[nodeGroupId]
				return val
			}
			deletedResult := make(chan string)
			fakeClient := &fake.Clientset{}
			provider := testprovider.NewTestCloudProvider(nil, func(nodeGroupId string, node string) error {
				if gotFailedRequest(nodeGroupId) {
					return fmt.Errorf("SIMULATED ERROR: won't remove node")
				}
				deletedResult <- nodeGroupId
				return nil
			})
			// 2d array represent the waves of pushing nodes to delete.
			deleteNodes := [][]*apiv1.Node{}

			for i := 0; i < test.deleteCalls; i++ {
				deleteNodes = append(deleteNodes, []*apiv1.Node{})
			}
			testNg1 := testprovider.NewTestNodeGroup("test-ng-1", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
			testNg2 := testprovider.NewTestNodeGroup("test-ng-2", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
			testNg3 := testprovider.NewTestNodeGroup("test-ng-3", 0, 100, 3, true, false, "n1-standard-2", nil, nil)
			testNg := map[string]*testprovider.TestNodeGroup{
				"test-ng-1": testNg1,
				"test-ng-2": testNg2,
				"test-ng-3": testNg3,
			}

			for ngName, numNodes := range test.numNodesToDelete {
				ng := testNg[ngName]
				provider.InsertNodeGroup(ng)
				ng.SetCloudProvider(provider)
				for i, num := range numNodes {
					singleBucketList := generateNodeGroupViewList(ng, 0, num)
					bucket := singleBucketList[0]
					deleteNodes[i] = append(deleteNodes[i], bucket.Nodes...)
					for _, node := range bucket.Nodes {
						provider.AddNode(bucket.Group.Id(), node)
					}
				}
			}
			opts := config.AutoscalingOptions{
				MaxScaleDownParallelism:        10,
				MaxDrainParallelism:            5,
				MaxPodEvictionTime:             0,
				DaemonSetEvictionForEmptyNodes: true,
			}

			podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
			pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})
			registry := kube_util.NewListerRegistry(nil, nil, podLister, pdbLister, nil, nil, nil, nil, nil)
			ctx, err := NewScaleTestAutoscalingContext(opts, fakeClient, registry, provider, nil, nil)
			if err != nil {
				t.Fatalf("Couldn't set up autoscaling context: %v", err)
			}
			csr := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, ctx.LogRecorder, NewBackoff(), nodegroupconfig.NewDefaultNodeGroupConfigProcessor(config.NodeGroupAutoscalingOptions{MaxNodeProvisionTime: 15 * time.Minute}))
			ndt := deletiontracker.NewNodeDeletionTracker(0)
			ndb := NewNodeDeletionBatcher(&ctx, csr, ndt, deleteInterval)
			evictor := Evictor{EvictionRetryTime: 0, DsEvictionRetryTime: 0, DsEvictionEmptyNodeTimeout: 0, PodEvictionHeadroom: DefaultPodEvictionHeadroom}
			actuator := Actuator{
				ctx: &ctx, clusterState: csr, nodeDeletionTracker: ndt,
				nodeDeletionScheduler: NewGroupDeletionScheduler(&ctx, ndt, ndb, evictor),
				budgetProcessor:       budgets.NewScaleDownBudgetProcessor(&ctx),
			}

			for _, nodes := range deleteNodes {
				actuator.StartDeletion(nodes, []*apiv1.Node{})
				time.Sleep(deleteInterval)
			}
			wantDeletedNodes := 0
			for _, num := range test.wantSuccessfulDeletion {
				wantDeletedNodes += num
			}
			gotDeletedNodes := map[string]int{
				"test-ng-1": 0,
				"test-ng-2": 0,
				"test-ng-3": 0,
			}
			for i := 0; i < wantDeletedNodes; i++ {
				select {
				case ngId := <-deletedResult:
					gotDeletedNodes[ngId]++
				case <-time.After(1 * time.Second):
					t.Errorf("Timeout while waiting for deleted nodes.")
					break
				}
			}
			if diff := cmp.Diff(test.wantSuccessfulDeletion, gotDeletedNodes); diff != "" {
				t.Errorf("Successful deleteions per node group diff (-want +got):\n%s", diff)
			}
		})
	}
}

func sizedNodeGroup(id string, size int, atomic bool) cloudprovider.NodeGroup {
	ng := testprovider.NewTestNodeGroup(id, 10000, 0, size, true, false, "n1-standard-2", nil, nil)
	ng.SetOptions(&config.NodeGroupAutoscalingOptions{
		ZeroOrMaxNodeScaling: atomic,
	})
	return ng
}

func generateNodes(from, to int, prefix string) []*apiv1.Node {
	var result []*apiv1.Node
	for i := from; i < to; i++ {
		name := fmt.Sprintf("node-%d", i)
		if prefix != "" {
			name = prefix + "-" + name
		}
		result = append(result, generateNode(name))
	}
	return result
}

func generateNodeGroupViewList(ng cloudprovider.NodeGroup, from, to int) []*budgets.NodeGroupView {
	return []*budgets.NodeGroupView{
		{
			Group: ng,
			Nodes: generateNodes(from, to, ng.Id()),
		},
	}
}

func generateNode(name string) *apiv1.Node {
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("8"),
				apiv1.ResourceMemory: resource.MustParse("8G"),
			},
		},
	}
}

func removablePods(count int, prefix string) []*apiv1.Pod {
	var result []*apiv1.Pod
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("pod-%d", i)
		if prefix != "" {
			name = prefix + "-" + name
		}
		result = append(result, removablePod(name, prefix))
	}
	return result
}

func removablePod(name string, node string) *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "true",
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: node,
			Containers: []apiv1.Container{
				{
					Name: "test-container",
					Resources: apiv1.ResourceRequirements{
						Requests: map[apiv1.ResourceName]resource.Quantity{
							apiv1.ResourceCPU:    resource.MustParse("1"),
							apiv1.ResourceMemory: resource.MustParse("1G"),
						},
					},
				},
			},
		},
	}
}

func generateDsPods(count int, node string) []*apiv1.Pod {

	var result []*apiv1.Pod
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("ds-pod-%d", i)
		result = append(result, generateDsPod(name, node))
	}
	return result
}

func generateDsPod(name string, node string) *apiv1.Pod {
	pod := removablePod(fmt.Sprintf("%s-%s", node, name), node)
	pod.OwnerReferences = GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", "some-uid")
	return pod
}

func generateDaemonSet() *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds",
			Namespace: "default",
			SelfLink:  "/apiv1s/apps/v1/namespaces/default/daemonsets/ds",
		},
	}
}

func generateUtilInfo(cpuUtil, memUtil float64) utilization.Info {
	var higherUtilName apiv1.ResourceName
	var higherUtilVal float64
	if cpuUtil > memUtil {
		higherUtilName = apiv1.ResourceCPU
		higherUtilVal = cpuUtil
	} else {
		higherUtilName = apiv1.ResourceMemory
		higherUtilVal = memUtil
	}
	return utilization.Info{
		CpuUtil:      cpuUtil,
		MemUtil:      memUtil,
		ResourceName: higherUtilName,
		Utilization:  higherUtilVal,
	}
}

func waitForDeletionResultsCount(ndt *deletiontracker.NodeDeletionTracker, resultsCount int, timeout, retryTime time.Duration) error {
	// This is quite ugly, but shouldn't matter much since in most cases there shouldn't be a need to wait at all, and
	// the function should return quickly after the first if check.
	// An alternative could be to turn NodeDeletionTracker into an interface, and use an implementation which allows
	// synchronizing calls to EndDeletion in the test code.
	for retryUntil := time.Now().Add(timeout); time.Now().Before(retryUntil); time.Sleep(retryTime) {
		if results, _ := ndt.DeletionResults(); len(results) == resultsCount {
			return nil
		}
	}
	return fmt.Errorf("timed out while waiting for node deletion results")
}
