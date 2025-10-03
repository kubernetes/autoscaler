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

package eligibility

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
)

type testCase struct {
	desc                        string
	nodes                       []*apiv1.Node
	pods                        []*apiv1.Pod
	draSnapshot                 *drasnapshot.Snapshot
	draEnabled                  bool
	wantUnneeded                []string
	wantUnremovable             []*simulator.UnremovableNode
	scaleDownUnready            bool
	ignoreDaemonSetsUtilization bool
}

func getTestCases(ignoreDaemonSetsUtilization bool, suffix string, now time.Time) []testCase {
	regularNode := BuildTestNode("regular", 1000, 10)
	SetNodeReadyState(regularNode, true, time.Time{})

	justDeletedNode := BuildTestNode("justDeleted", 1000, 10)
	justDeletedNode.Spec.Taints = []apiv1.Taint{{Key: taints.ToBeDeletedTaint, Value: strconv.FormatInt(now.Unix()-30, 10)}}
	SetNodeReadyState(justDeletedNode, true, time.Time{})

	noScaleDownNode := BuildTestNode("noScaleDown", 1000, 10)
	noScaleDownNode.Annotations = map[string]string{ScaleDownDisabledKey: "true"}
	SetNodeReadyState(noScaleDownNode, true, time.Time{})

	unreadyNode := BuildTestNode("unready", 1000, 10)
	SetNodeReadyState(unreadyNode, false, time.Time{})

	bigPod := BuildTestPod("bigPod", 600, 0)
	bigPod.Spec.NodeName = "regular"

	smallPod := BuildTestPod("smallPod", 100, 0)
	smallPod.Spec.NodeName = "regular"

	dsPod := BuildTestPod("dsPod", 500, 0, WithDSController())
	dsPod.Spec.NodeName = "regular"

	brokenUtilNode := BuildTestNode("regular", 0, 0)
	resourceSliceNodeName := "regular"
	regularNodeIncompleteResourceSlice := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "regularNodeIncompleteResourceSlice", UID: "regularNodeIncompleteResourceSlice"},
		Spec: resourceapi.ResourceSliceSpec{
			Driver:   "driver.foo.com",
			NodeName: &resourceSliceNodeName,
			Pool: resourceapi.ResourcePool{
				Name:               "regular-pool",
				ResourceSliceCount: 999,
			},
			Devices: []resourceapi.Device{{Name: "dev1"}},
		},
	}
	testCases := []testCase{
		{
			desc:             "regular node stays",
			nodes:            []*apiv1.Node{regularNode},
			wantUnneeded:     []string{"regular"},
			wantUnremovable:  []*simulator.UnremovableNode{},
			scaleDownUnready: true,
		},
		{
			desc:             "recently deleted node is filtered out",
			nodes:            []*apiv1.Node{regularNode, justDeletedNode},
			wantUnneeded:     []string{"regular"},
			wantUnremovable:  []*simulator.UnremovableNode{{Node: justDeletedNode, Reason: simulator.CurrentlyBeingDeleted}},
			scaleDownUnready: true,
		},
		{
			desc:             "marked no scale down is filtered out",
			nodes:            []*apiv1.Node{noScaleDownNode, regularNode},
			wantUnneeded:     []string{"regular"},
			wantUnremovable:  []*simulator.UnremovableNode{{Node: noScaleDownNode, Reason: simulator.ScaleDownDisabledAnnotation}},
			scaleDownUnready: true,
		},
		{
			desc:             "highly utilized node is filtered out",
			nodes:            []*apiv1.Node{regularNode},
			pods:             []*apiv1.Pod{bigPod},
			wantUnneeded:     []string{},
			wantUnremovable:  []*simulator.UnremovableNode{{Node: regularNode, Reason: simulator.NotUnderutilized}},
			scaleDownUnready: true,
		},
		{
			desc:             "underutilized node stays",
			nodes:            []*apiv1.Node{regularNode},
			pods:             []*apiv1.Pod{smallPod},
			wantUnneeded:     []string{"regular"},
			wantUnremovable:  []*simulator.UnremovableNode{},
			scaleDownUnready: true,
		},
		{
			desc:             "node is filtered out if utilization can't be calculated",
			nodes:            []*apiv1.Node{brokenUtilNode},
			pods:             []*apiv1.Pod{smallPod},
			wantUnneeded:     []string{},
			wantUnremovable:  []*simulator.UnremovableNode{{Node: brokenUtilNode, Reason: simulator.UnexpectedError}},
			scaleDownUnready: true,
		},
		{
			desc:             "unready node stays",
			nodes:            []*apiv1.Node{unreadyNode},
			wantUnneeded:     []string{"unready"},
			wantUnremovable:  []*simulator.UnremovableNode{},
			scaleDownUnready: true,
		},
		{
			desc:             "unready node is filtered oud when scale-down of unready is disabled",
			nodes:            []*apiv1.Node{unreadyNode},
			wantUnneeded:     []string{},
			wantUnremovable:  []*simulator.UnremovableNode{{Node: unreadyNode, Reason: simulator.ScaleDownUnreadyDisabled}},
			scaleDownUnready: false,
		},
		{
			desc:             "Node is not filtered out because of DRA issues if DRA is disabled",
			nodes:            []*apiv1.Node{regularNode},
			pods:             []*apiv1.Pod{smallPod},
			draSnapshot:      drasnapshot.NewSnapshot(nil, map[string][]*resourceapi.ResourceSlice{"regular": {regularNodeIncompleteResourceSlice}}, nil, nil),
			draEnabled:       false,
			wantUnneeded:     []string{"regular"},
			wantUnremovable:  []*simulator.UnremovableNode{},
			scaleDownUnready: true,
		},
		{
			desc:             "Node is filtered out because of DRA issues if DRA is enabled",
			nodes:            []*apiv1.Node{regularNode},
			pods:             []*apiv1.Pod{smallPod},
			draSnapshot:      drasnapshot.NewSnapshot(nil, map[string][]*resourceapi.ResourceSlice{"regular": {regularNodeIncompleteResourceSlice}}, nil, nil),
			draEnabled:       true,
			wantUnneeded:     []string{},
			wantUnremovable:  []*simulator.UnremovableNode{{Node: regularNode, Reason: simulator.UnexpectedError}},
			scaleDownUnready: true,
		},
	}

	finalTestCases := []testCase{}
	for _, tc := range testCases {
		tc.desc = tc.desc + " " + suffix
		if ignoreDaemonSetsUtilization {
			tc.ignoreDaemonSetsUtilization = true
		}
		finalTestCases = append(finalTestCases, tc)
	}

	if ignoreDaemonSetsUtilization {
		finalTestCases = append(testCases, testCase{
			desc:                        "high utilization daemonsets node is filtered out",
			nodes:                       []*apiv1.Node{regularNode},
			pods:                        []*apiv1.Pod{smallPod, dsPod},
			wantUnneeded:                []string{},
			wantUnremovable:             []*simulator.UnremovableNode{{Node: regularNode, Reason: simulator.NotUnderutilized}},
			scaleDownUnready:            true,
			ignoreDaemonSetsUtilization: false,
		},
			testCase{
				desc:                        "high utilization daemonsets node stays",
				nodes:                       []*apiv1.Node{regularNode},
				pods:                        []*apiv1.Pod{smallPod, dsPod},
				wantUnneeded:                []string{"regular"},
				wantUnremovable:             []*simulator.UnremovableNode{},
				scaleDownUnready:            true,
				ignoreDaemonSetsUtilization: true,
			})
	}

	return finalTestCases
}

func TestFilterOutUnremovable(t *testing.T) {
	now := time.Now()
	for _, tc := range append(getTestCases(false, "IgnoreDaemonSetUtilization=false", now),
		getTestCases(true, "IgnoreDaemonsetUtilization=true", now)...) {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			options := config.AutoscalingOptions{
				DynamicResourceAllocationEnabled: tc.draEnabled,
				UnremovableNodeRecheckTimeout:    5 * time.Minute,
				ScaleDownUnreadyEnabled:          tc.scaleDownUnready,
				NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
					ScaleDownUtilizationThreshold:    config.DefaultScaleDownUtilizationThreshold,
					ScaleDownGpuUtilizationThreshold: config.DefaultScaleDownGpuUtilizationThreshold,
					ScaleDownUnneededTime:            config.DefaultScaleDownUnneededTime,
					ScaleDownUnreadyTime:             config.DefaultScaleDownUnreadyTime,
					IgnoreDaemonSetsUtilization:      tc.ignoreDaemonSetsUtilization,
				},
			}
			s := nodegroupconfig.NewDefaultNodeGroupConfigProcessor(options.NodeGroupDefaults)
			c := NewChecker(s)
			provider := testprovider.NewTestCloudProviderBuilder().Build()
			provider.AddNodeGroup("ng1", 1, 10, 2)
			for _, n := range tc.nodes {
				provider.AddNode("ng1", n)
			}
			context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil, nil)
			if err != nil {
				t.Fatalf("Could not create autoscaling context: %v", err)
			}
			if err := context.ClusterSnapshot.SetClusterState(tc.nodes, tc.pods, tc.draSnapshot); err != nil {
				t.Fatalf("Could not SetClusterState: %v", err)
			}
			unremovableNodes := unremovable.NewNodes()
			gotUnneeded, _, gotUnremovable := c.FilterOutUnremovable(&context, tc.nodes, now, unremovableNodes)
			if diff := cmp.Diff(tc.wantUnneeded, gotUnneeded); diff != "" {
				t.Errorf("FilterOutUnremovable(): unexpected unneeded (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantUnremovable, gotUnremovable); diff != "" {
				t.Errorf("FilterOutUnremovable(): unexpected unremovable (-want +got): %s", diff)
			}
		})
	}
}
