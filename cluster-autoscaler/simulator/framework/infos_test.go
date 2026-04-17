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
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/klog/v2"
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
		modFn                   func() *NodeInfo
		wantSchedNodeInfo       *schedulerimpl.NodeInfo
		wantLocalResourceSlices []*resourceapi.ResourceSlice
		wantPods                []*PodInfo
	}{
		{
			testName: "wrapping via NewNodeInfo",
			modFn: func() *NodeInfo {
				return NewNodeInfo(node, nil, testPodInfos(pods, false)...)
			},
			wantSchedNodeInfo: schedulerNodeInfo,
			wantPods:          testPodInfos(pods, false),
		},
		{
			testName: "wrapping via NewNodeInfo with DRA objects",
			modFn: func() *NodeInfo {
				return NewNodeInfo(node, slices, testPodInfos(pods, true)...)
			},
			wantSchedNodeInfo:       schedulerNodeInfo,
			wantLocalResourceSlices: slices,
			wantPods:                testPodInfos(pods, true),
		},
		{
			testName: "wrapping via NewTestNodeInfo",
			modFn: func() *NodeInfo {
				return NewTestNodeInfo(node, pods...)
			},
			wantSchedNodeInfo: schedulerNodeInfo,
			wantPods:          testPodInfos(pods, false),
		},
		{
			testName: "wrapping via SetNode+AddPod",
			modFn: func() *NodeInfo {
				result := NewNodeInfo(nil, nil)
				result.SetNode(node)
				for _, pod := range pods {
					result.AddPod(NewPodInfo(pod, nil))
				}
				return result
			},
			wantSchedNodeInfo: schedulerNodeInfo,
			wantPods:          testPodInfos(pods, false),
		},
		{
			testName: "wrapping via SetNode+AddPod with DRA objects",
			modFn: func() *NodeInfo {
				result := NewNodeInfo(nil, nil)
				result.LocalResourceSlices = slices
				result.SetNode(node)
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
			modFn: func() *NodeInfo {
				result := NewNodeInfo(node, slices, testPodInfos(pods, true)...)
				for _, pod := range []*apiv1.Pod{pods[0], pods[2], pods[4]} {
					if err := result.RemovePod(klog.Background(), pod); err != nil {
						t.Errorf("RemovePod unexpected error: %v", err)
					}
				}
				return result
			},
			wantSchedNodeInfo:       newSchedNodeInfo(node, []*apiv1.Pod{pods[1], pods[3], pods[5]}),
			wantLocalResourceSlices: slices,
			wantPods:                testPodInfos([]*apiv1.Pod{pods[1], pods[3], pods[5]}, true),
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			wrappedNodeInfo := tc.modFn()
			// Assert that the scheduler NodeInfo object is as expected.
			nodeInfoCmpOpts := NodeInfoCmpOptions()
			if diff := cmp.Diff(tc.wantSchedNodeInfo, wrappedNodeInfo.NodeInfo, nodeInfoCmpOpts...); diff != "" {
				t.Errorf("Scheduler Node Info output differs from expected, diff (-want +got): %s", diff)
			}
			// Assert that the Node() method matches the scheduler object.
			if diff := cmp.Diff(tc.wantSchedNodeInfo.Node(), wrappedNodeInfo.Node()); diff != "" {
				t.Errorf("Node() output differs from expected, diff  (-want +got): %s", diff)
			}
			// Assert that LocalResourceSlices are as expected.
			if diff := cmp.Diff(tc.wantLocalResourceSlices, wrappedNodeInfo.LocalResourceSlices); diff != "" {
				t.Errorf("LocalResourceSlices differ from expected, diff  (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantPods, wrappedNodeInfo.Pods(), nodeInfoCmpOpts...); diff != "" {
				t.Errorf("Pods() output differs from expected, diff (-want +got): %s", diff)
			}
		})
	}
}

func TestDeepCopyNodeInfo(t *testing.T) {
	nodeName := "node"
	node := test.BuildTestNode(nodeName, 1000, 1000)
	pods := []*PodInfo{
		NewPodInfo(test.BuildTestPod("p1", 80, 0, test.WithNodeName(node.Name)), nil),
		NewPodInfo(test.BuildTestPod("p2", 80, 0, test.WithNodeName(node.Name)), []*resourceapi.ResourceClaim{
			{ObjectMeta: v1.ObjectMeta{Name: "claim1"}, Spec: resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "req1"}}}}},
			{ObjectMeta: v1.ObjectMeta{Name: "claim2"}, Spec: resourceapi.ResourceClaimSpec{Devices: resourceapi.DeviceClaim{Requests: []resourceapi.DeviceRequest{{Name: "req2"}}}}},
		}),
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
			if diff := cmp.Diff(tc.nodeInfo, nodeInfoCopy, NodeInfoCmpOptions()...); diff != "" {
				t.Errorf("nodeInfo differs after DeepCopyNodeInfo, diff (-want +got): %s", diff)
			}

			// Verify that the object addresses changed in the copy.
			if tc.nodeInfo == nodeInfoCopy {
				t.Error("nodeInfo address identical after DeepCopyNodeInfo")
			}
			if tc.nodeInfo.NodeInfo == nodeInfoCopy.NodeInfo {
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
					if oldPodInfo.NeededResourceClaims[claimIndex] == newPodInfo.NeededResourceClaims[claimIndex] {
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
		podInfo := NewPodInfo(pod, nil)
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
