/*
Copyright 2024 The Kubernetes Authors.

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

package framework

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	schedulerinterface "k8s.io/kube-scheduler/framework"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestNodeInfo(t *testing.T) {
	nodeName := "test-node"
	node := test.BuildTestNode(nodeName, 1000, 1024)
	pods := []*apiv1.Pod{
		// Use pods requesting host-ports to make sure that NodeInfo fields other than node and Pods also
		// get set correctly (in this case - the UsedPorts field).
		test.BuildTestPod("hostport-pod-0", 100, 16, test.WithHostPort(1337)),
		test.BuildTestPod("hostport-pod-1", 100, 16, test.WithHostPort(1338)),
		test.BuildTestPod("hostport-pod-2", 100, 16, test.WithHostPort(1339)),
		test.BuildTestPod("regular-pod-0", 100, 16),
		test.BuildTestPod("regular-pod-1", 100, 16),
		test.BuildTestPod("regular-pod-2", 100, 16),
	}
	extraPod := test.BuildTestPod("extra-pod", 1, 1)
	schedulerNodeInfo := newSchedNodeInfo(node, pods)
	slices := []*resourceapi.ResourceSlice{
		{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-node-slice-0",
			},
			Spec: resourceapi.ResourceSliceSpec{
				NodeName: &nodeName,
				Driver:   "test.driver.com",
				Pool:     resourceapi.ResourcePool{Name: nodeName, Generation: 13, ResourceSliceCount: 2},
				Devices:  []resourceapi.Device{{Name: "device-0"}, {Name: "device-1"}},
			}},
		{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-node-slice-1",
			},
			Spec: resourceapi.ResourceSliceSpec{
				NodeName: &nodeName,
				Driver:   "test.driver.com",
				Pool:     resourceapi.ResourcePool{Name: nodeName, Generation: 13, ResourceSliceCount: 2},
				Devices:  []resourceapi.Device{{Name: "device-2"}, {Name: "device-3"}},
			},
		},
	}

	for _, tc := range []struct {
		testName                string
		modFn                   func(info schedulerinterface.NodeInfo) *NodeInfo
		wantSchedNodeInfo       schedulerinterface.NodeInfo
		wantLocalResourceSlices []*resourceapi.ResourceSlice
		wantPods                []*PodInfo
	}{
		{
			testName: "wrapping via NewNodeInfo",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				return NewNodeInfo(info.Node(), nil, testPodInfos(pods, false)...)
			},
			wantSchedNodeInfo: schedulerNodeInfo,
			wantPods:          testPodInfos(pods, false),
		},
		{
			testName: "wrapping via NewNodeInfo with DRA objects",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				return NewNodeInfo(info.Node(), slices, testPodInfos(pods, true)...)
			},
			wantSchedNodeInfo:       schedulerNodeInfo,
			wantLocalResourceSlices: slices,
			wantPods:                testPodInfos(pods, true),
		},
		{
			testName: "wrapping via NewTestNodeInfo",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				var pods []*apiv1.Pod
				for _, pod := range info.GetPods() {
					pods = append(pods, pod.GetPod())
				}
				return NewTestNodeInfo(info.Node(), pods...)
			},
			wantSchedNodeInfo: schedulerNodeInfo,
			wantPods:          testPodInfos(pods, false),
		},
		{
			testName: "wrapping via WrapSchedulerNodeInfo",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				return WrapSchedulerNodeInfo(info, nil, nil)
			},
			wantSchedNodeInfo: schedulerNodeInfo,
			wantPods:          testPodInfos(pods, false),
		},
		{
			testName: "wrapping via WrapSchedulerNodeInfo with DRA objects",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				podInfos := testPodInfos(pods, true)
				extraInfos := make(map[types.UID]PodExtraInfo)
				for _, podInfo := range podInfos {
					extraInfos[podInfo.Pod.UID] = podInfo.PodExtraInfo
				}
				return WrapSchedulerNodeInfo(schedulerNodeInfo, slices, extraInfos)
			},
			wantSchedNodeInfo:       schedulerNodeInfo,
			wantLocalResourceSlices: slices,
			wantPods:                testPodInfos(pods, true),
		},
		{
			testName: "wrapping via SetNode+AddPod",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				result := NewNodeInfo(nil, nil)
				result.SetNode(info.Node())
				for _, pod := range info.GetPods() {
					result.AddPod(&PodInfo{Pod: pod.GetPod()})
				}
				return result
			},
			wantSchedNodeInfo: schedulerNodeInfo,
			wantPods:          testPodInfos(pods, false),
		},
		{
			testName: "wrapping via SetNode+AddPod with DRA objects",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				result := NewNodeInfo(nil, nil)
				result.LocalResourceSlices = slices
				result.SetNode(info.Node())
				for _, podInfo := range testPodInfos(pods, true) {
					result.AddPod(podInfo)
				}
				return result
			},
			wantSchedNodeInfo:       schedulerNodeInfo,
			wantLocalResourceSlices: slices,
			wantPods:                testPodInfos(pods, true),
		},
		{
			testName: "removing pods",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				result := NewNodeInfo(info.Node(), slices, testPodInfos(pods, true)...)
				for _, pod := range []*apiv1.Pod{pods[0], pods[2], pods[4]} {
					if err := result.RemovePod(pod); err != nil {
						t.Errorf("RemovePod unexpected error: %v", err)
					}
				}
				return result
			},
			wantSchedNodeInfo:       newSchedNodeInfo(node, []*apiv1.Pod{pods[1], pods[3], pods[5]}),
			wantLocalResourceSlices: slices,
			wantPods:                testPodInfos([]*apiv1.Pod{pods[1], pods[3], pods[5]}, true),
		},
		{
			testName: "wrapping via WrapSchedulerNodeInfo and adding more pods",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				result := WrapSchedulerNodeInfo(info, nil, nil)
				result.AddPod(testPodInfos([]*apiv1.Pod{extraPod}, false)[0])
				return result
			},
			wantSchedNodeInfo: newSchedNodeInfo(node, append(pods, extraPod)),
			wantPods:          testPodInfos(append(pods, extraPod), false),
		},
		{
			testName: "wrapping via WrapSchedulerNodeInfo and adding more pods using DRA",
			modFn: func(info schedulerinterface.NodeInfo) *NodeInfo {
				result := WrapSchedulerNodeInfo(info, nil, nil)
				result.AddPod(testPodInfos([]*apiv1.Pod{extraPod}, true)[0])
				return result
			},
			wantSchedNodeInfo: newSchedNodeInfo(node, append(pods, extraPod)),
			wantPods:          append(testPodInfos(pods, false), testPodInfos([]*apiv1.Pod{extraPod}, true)...),
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			wrappedNodeInfo := tc.modFn(schedulerNodeInfo.Snapshot())
			// Assert that the scheduler NodeInfo object is as expected.
			nodeInfoCmpOpts := []cmp.Option{
				// The Node is the only unexported field in this type, and we want to compare it.
				cmp.AllowUnexported(schedulerimpl.NodeInfo{}),
				// Generation is expected to be different.
				cmpopts.IgnoreFields(schedulerimpl.NodeInfo{}, "Generation"),
				// The pod order changes in a particular way whenever schedulerimpl.RemovePod() is called. Instead of
				// relying on that schedulerimpl implementation detail in assertions, just ignore the order.
				cmpopts.SortSlices(func(p1, p2 schedulerinterface.PodInfo) bool {
					return p1.GetPod().Name < p2.GetPod().Name
				}),
				cmpopts.IgnoreUnexported(schedulerimpl.PodInfo{}),
			}
			if diff := cmp.Diff(tc.wantSchedNodeInfo, wrappedNodeInfo.ToScheduler(), nodeInfoCmpOpts...); diff != "" {
				t.Errorf("ToScheduler() output differs from expected, diff (-want +got): %s", diff)
			}

			// Assert that the Node() method matches the scheduler object.
			if diff := cmp.Diff(tc.wantSchedNodeInfo.Node(), wrappedNodeInfo.Node()); diff != "" {
				t.Errorf("Node() output differs from expected, diff  (-want +got): %s", diff)
			}

			// Assert that LocalResourceSlices are as expected.
			if diff := cmp.Diff(tc.wantLocalResourceSlices, wrappedNodeInfo.LocalResourceSlices); diff != "" {
				t.Errorf("LocalResourceSlices differ from expected, diff  (-want +got): %s", diff)
			}

			// Assert that the pods list in the wrapper is as expected.
			// The pod order changes in a particular way whenever schedulerimpl.RemovePod() is called. Instead of
			// relying on that schedulerimpl implementation detail in assertions, just ignore the order.
			podsInfosIgnoreOrderOpt := cmpopts.SortSlices(func(p1, p2 *PodInfo) bool {
				return p1.Name < p2.Name
			})
			if diff := cmp.Diff(tc.wantPods, wrappedNodeInfo.Pods(), podsInfosIgnoreOrderOpt); diff != "" {
				t.Errorf("Pods() output differs from expected, diff (-want +got): %s", diff)
			}

			// Assert that the extra info map only contains information about pods in the list. This verifies that
			// the map is properly cleaned up during RemovePod.
			possiblePodUids := make(map[types.UID]bool)
			for _, pod := range tc.wantPods {
				possiblePodUids[pod.UID] = true
			}
			for podUid := range wrappedNodeInfo.podsExtraInfo {
				if !possiblePodUids[podUid] {
					t.Errorf("podsExtraInfo contains entry for unexpected UID %q", podUid)
				}
			}
		})
	}
}

func TestDeepCopyNodeInfo(t *testing.T) {
	nodeName := "node"
	node := test.BuildTestNode(nodeName, 1000, 1000)
	pods := []*PodInfo{
		{Pod: test.BuildTestPod("p1", 80, 0, test.WithNodeName(node.Name))},
		{
			Pod: test.BuildTestPod("p2", 80, 0, test.WithNodeName(node.Name)),
			PodExtraInfo: PodExtraInfo{
				NeededResourceClaims: []*resourceapi.ResourceClaim{
					{ObjectMeta: v1.ObjectMeta{Name: "claim1"}, Spec: resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "req1"}}}}},
					{ObjectMeta: v1.ObjectMeta{Name: "claim2"}, Spec: resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "req2"}}}}},
				},
			},
		},
	}
	slices := []*resourceapi.ResourceSlice{
		{ObjectMeta: v1.ObjectMeta{Name: "slice1"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &nodeName}},
		{ObjectMeta: v1.ObjectMeta{Name: "slice2"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &nodeName}},
	}

	for _, tc := range []struct {
		testName string
		nodeInfo *NodeInfo
	}{
		{
			testName: "empty NodeInfo",
			nodeInfo: NewNodeInfo(nil, nil),
		},
		{
			testName: "NodeInfo with only Node set",
			nodeInfo: NewNodeInfo(node, nil),
		},
		{
			testName: "NodeInfo with only Pods set",
			nodeInfo: NewNodeInfo(nil, nil, pods...),
		},
		{
			testName: "NodeInfo with both Node and Pods set",
			nodeInfo: NewNodeInfo(node, nil, pods...),
		},
		{
			testName: "NodeInfo with Node, ResourceSlices, and Pods set",
			nodeInfo: NewNodeInfo(node, slices, pods...),
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			// Verify that the contents are identical after copying.
			nodeInfoCopy := tc.nodeInfo.DeepCopy()
			if diff := cmp.Diff(tc.nodeInfo, nodeInfoCopy,
				cmp.AllowUnexported(schedulerimpl.NodeInfo{}, NodeInfo{}),
				// We don't care about this field staying the same, and it differs because it's a global counter bumped
				// on every AddPod.
				cmpopts.IgnoreFields(schedulerimpl.NodeInfo{}, "Generation"),
				cmpopts.IgnoreUnexported(schedulerimpl.PodInfo{}),
			); diff != "" {
				t.Errorf("nodeInfo differs after DeepCopyNodeInfo, diff (-want +got): %s", diff)
			}

			// Verify that the object addresses changed in the copy.
			if tc.nodeInfo == nodeInfoCopy {
				t.Error("nodeInfo address identical after DeepCopyNodeInfo")
			}
			if tc.nodeInfo.ToScheduler() == nodeInfoCopy.ToScheduler() {
				t.Error("schedulerimpl.NodeInfo address identical after DeepCopyNodeInfo")
			}
			for i := range len(tc.nodeInfo.LocalResourceSlices) {
				if tc.nodeInfo.LocalResourceSlices[i] == nodeInfoCopy.LocalResourceSlices[i] {
					t.Errorf("%d-th LocalResourceSlice address identical after DeepCopyNodeInfo", i)
				}
			}
			for podIndex := range len(tc.nodeInfo.Pods()) {
				oldPodInfo := tc.nodeInfo.Pods()[podIndex]
				newPodInfo := nodeInfoCopy.Pods()[podIndex]
				if oldPodInfo == newPodInfo {
					t.Errorf("%d-th PodInfo address identical after DeepCopyNodeInfo", podIndex)
				}
				if oldPodInfo.Pod == newPodInfo.Pod {
					t.Errorf("%d-th PodInfo.Pod address identical after DeepCopyNodeInfo", podIndex)
				}
				for claimIndex := range len(newPodInfo.NeededResourceClaims) {
					if oldPodInfo.NeededResourceClaims[podIndex] == newPodInfo.NeededResourceClaims[podIndex] {
						t.Errorf("%d-th PodInfo - %d-th NeededResourceClaim address identical after DeepCopyNodeInfo", podIndex, claimIndex)
					}
				}
			}
		})
	}
}

func TestNodeInfoResourceClaims(t *testing.T) {
	node := test.BuildTestNode("node", 1000, 1000)
	pods := []*apiv1.Pod{
		test.BuildTestPod("pod-0", 100, 16),
		test.BuildTestPod("pod-1", 100, 16),
		test.BuildTestPod("pod-2", 100, 16),
	}

	for _, tc := range []struct {
		testName   string
		nodeInfo   *NodeInfo
		wantClaims []*resourceapi.ResourceClaim
	}{
		{
			testName:   "no pods",
			nodeInfo:   NewNodeInfo(node, nil),
			wantClaims: nil,
		},
		{
			testName:   "pods but no claims",
			nodeInfo:   NewNodeInfo(node, nil, testPodInfos(pods, false)...),
			wantClaims: nil,
		},
		{
			testName: "pods with claims, shared claims are not repeated",
			nodeInfo: NewNodeInfo(node, nil, testPodInfos(pods, true)...),
			wantClaims: []*resourceapi.ResourceClaim{
				testClaim("pod-0-claim-0"),
				testClaim("pod-0-claim-1"),
				testClaim("pod-0-claim-2"),
				testClaim("pod-1-claim-0"),
				testClaim("pod-1-claim-1"),
				testClaim("pod-1-claim-2"),
				testClaim("pod-2-claim-0"),
				testClaim("pod-2-claim-1"),
				testClaim("pod-2-claim-2"),
				testClaim("shared-claim-0"),
				testClaim("shared-claim-1"),
				testClaim("shared-claim-2"),
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			claims := tc.nodeInfo.ResourceClaims()
			ignoreClaimOrder := cmpopts.SortSlices(func(c1, c2 *resourceapi.ResourceClaim) bool { return c1.Name < c2.Name })
			if diff := cmp.Diff(tc.wantClaims, claims, ignoreClaimOrder, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("NodeInfo.ResourceClaims(): unexpected output (-want +got): %s", diff)
			}
		})
	}
}

func testPodInfos(pods []*apiv1.Pod, addClaims bool) []*PodInfo {
	var result []*PodInfo
	for _, pod := range pods {
		podInfo := &PodInfo{Pod: pod}
		if addClaims {
			for i := range 3 {
				podInfo.NeededResourceClaims = append(podInfo.NeededResourceClaims, testClaim(fmt.Sprintf("%s-claim-%d", pod.Name, i)))
				podInfo.NeededResourceClaims = append(podInfo.NeededResourceClaims, testClaim(fmt.Sprintf("shared-claim-%d", i)))
			}
		}
		result = append(result, podInfo)
	}
	return result
}

func testClaim(claimName string) *resourceapi.ResourceClaim {
	return &resourceapi.ResourceClaim{
		ObjectMeta: v1.ObjectMeta{Name: claimName, UID: types.UID(claimName)},
		Spec: resourceapi.ResourceClaimSpec{
			Devices: resourceapi.DeviceClaim{
				Requests: []resourceapi.DeviceRequest{
					{Name: "request-0"},
					{Name: "request-1"},
				},
			},
		},
	}
}

func newSchedNodeInfo(node *apiv1.Node, pods []*apiv1.Pod) *schedulerimpl.NodeInfo {
	result := schedulerimpl.NewNodeInfo(pods...)
	result.SetNode(node)
	return result
}
