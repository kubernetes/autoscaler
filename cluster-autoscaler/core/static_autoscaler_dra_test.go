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

package core

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	resourceapi "k8s.io/api/resource/v1alpha3"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	scaledownstatus "k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	featuretesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/kubernetes/pkg/features"
	schedconfig "k8s.io/kubernetes/pkg/scheduler/apis/config"
	schedconfiglatest "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"
	draplugin "k8s.io/kubernetes/pkg/scheduler/framework/plugins/dynamicresources"
)

const (
	exampleDriver = "dra.example.com"

	gpuDevice    = "gpuDevice"
	gpuAttribute = "gpuType"
	gpuTypeA     = "gpuA"
	gpuTypeB     = "gpuB"

	nicDevice    = "nicDevice"
	nicAttribute = "nicType"
	nicTypeA     = "nicA"

	globalDevice       = "globalDev"
	globalDevAttribute = "globalDevType"
	globalDevTypeA     = "globalDevA"
)

var (
	defaultDeviceClass = &resourceapi.DeviceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-class",
			Namespace: "default",
		},
		Spec: resourceapi.DeviceClassSpec{
			Selectors:     nil,
			Config:        nil,
			SuitableNodes: nil,
		},
	}
)

type fakeResourceClaimLister struct {
	claims []*resourceapi.ResourceClaim
}

func (l *fakeResourceClaimLister) List() ([]*resourceapi.ResourceClaim, error) {
	if l == nil {
		return nil, nil
	}
	return l.claims, nil
}

type fakeResourceSliceLister struct {
	slices []*resourceapi.ResourceSlice
}

func (l *fakeResourceSliceLister) List() ([]*resourceapi.ResourceSlice, error) {
	if l == nil {
		return nil, nil
	}
	return l.slices, nil
}

type fakeDeviceClassLister struct {
	devices []*resourceapi.DeviceClass
}

func (l *fakeDeviceClassLister) List() ([]*resourceapi.DeviceClass, error) {
	if l == nil {
		return nil, nil
	}
	return l.devices, nil
}

type fakeScaleUpStatusProcessor struct {
	lastStatus *status.ScaleUpStatus
}

func (f *fakeScaleUpStatusProcessor) Process(context *context.AutoscalingContext, status *status.ScaleUpStatus) {
	f.lastStatus = status
}

func (f *fakeScaleUpStatusProcessor) CleanUp() {
}

type fakeScaleDownStatusProcessor struct {
	lastStatus *scaledownstatus.ScaleDownStatus
}

func (f *fakeScaleDownStatusProcessor) Process(context *context.AutoscalingContext, status *scaledownstatus.ScaleDownStatus) {
	f.lastStatus = status
}

func (f *fakeScaleDownStatusProcessor) CleanUp() {
}

type testDeviceRequest struct {
	name      string
	selectors []string
	count     int64
	all       bool
}

type testDevice struct {
	name       string
	attributes map[string]string
	capacity   map[string]string
}

type testAllocation struct {
	request string
	driver  string
	pool    string
	device  string
}

type sliceNodeAvailability struct {
	node  string
	nodes []string
	all   bool
}

type testPod struct {
	pod    *apiv1.Pod
	claims []*resourceapi.ResourceClaim
}

type testNodeGroupDef struct {
	name               string
	cpu, mem           int64
	slicesTemplateFunc func(nodeName string) []*resourceapi.ResourceSlice
}

type noScaleUpDef struct {
	podName       string
	podNamePrefix string
	podCount      int
}

type noScaleDownDef struct {
	nodeName       string
	nodeNamePrefix string
	nodeCount      int
	reason         simulator.UnremovableReason
}

func TestStaticAutoscalerDynamicResources(t *testing.T) {
	//// Uncomment to get logs from the DRA plugin.
	//var fs flag.FlagSet
	//klog.InitFlags(&fs)
	//assert.NoError(t, fs.Set("v", "5"))

	featuretesting.SetFeatureGateDuringTest(t, feature.DefaultFeatureGate, features.DynamicResourceAllocation, true)
	schedConfig, err := schedconfiglatest.Default()
	assert.NoError(t, err)
	schedConfig.Profiles[0].Plugins.PreFilter.Enabled = append(schedConfig.Profiles[0].Plugins.PreFilter.Enabled, schedconfig.Plugin{Name: draplugin.Name})
	schedConfig.Profiles[0].Plugins.Filter.Enabled = append(schedConfig.Profiles[0].Plugins.PreFilter.Enabled, schedconfig.Plugin{Name: draplugin.Name})
	schedConfig.Profiles[0].Plugins.Reserve.Enabled = append(schedConfig.Profiles[0].Plugins.PreFilter.Enabled, schedconfig.Plugin{Name: draplugin.Name})

	now := time.Now()

	node1GpuA1slice := &testNodeGroupDef{name: "node1GpuA1slice", cpu: 1000, mem: 1000, slicesTemplateFunc: nodeTemplateResourceSlices(exampleDriver, 1, 0, []testDevice{{name: gpuDevice + "-0", attributes: map[string]string{gpuAttribute: gpuTypeA}}})}
	node1GpuB1slice := &testNodeGroupDef{name: "node1GpuB1slice", cpu: 1000, mem: 1000, slicesTemplateFunc: nodeTemplateResourceSlices(exampleDriver, 1, 0, []testDevice{{name: gpuDevice + "-0", attributes: map[string]string{gpuAttribute: gpuTypeB}}})}
	node3GpuA1slice := &testNodeGroupDef{name: "node3GpuA1slice", cpu: 1000, mem: 1000, slicesTemplateFunc: nodeTemplateResourceSlices(exampleDriver, 1, 0, testDevices(gpuDevice, 3, map[string]string{gpuAttribute: gpuTypeA}, nil))}
	node3GpuA3slice := &testNodeGroupDef{name: "node3GpuA3slice", cpu: 1000, mem: 1000, slicesTemplateFunc: nodeTemplateResourceSlices(exampleDriver, 3, 0, testDevices(gpuDevice, 3, map[string]string{gpuAttribute: gpuTypeA}, nil))}
	node1Nic1slice := &testNodeGroupDef{name: "node1Nic1slice", cpu: 1000, mem: 1000, slicesTemplateFunc: nodeTemplateResourceSlices(exampleDriver, 1, 0, []testDevice{{name: nicDevice + "-0", attributes: map[string]string{nicAttribute: nicTypeA}}})}
	node1Gpu1Nic1slice := &testNodeGroupDef{name: "node1Gpu1Nic1slice", cpu: 1000, mem: 1000, slicesTemplateFunc: nodeTemplateResourceSlices(exampleDriver, 1, 0, []testDevice{
		{name: gpuDevice + "-0", attributes: map[string]string{gpuAttribute: gpuTypeA}},
		{name: nicDevice + "-0", attributes: map[string]string{nicAttribute: nicTypeA}},
	})}

	baseBigPod := BuildTestPod("", 600, 100)
	baseSmallPod := BuildTestPod("", 100, 100)

	req1GpuA := testDeviceRequest{name: "req1GpuA", count: 1, selectors: singleAttrSelector(exampleDriver, gpuAttribute, gpuTypeA)}
	req2GpuA := testDeviceRequest{name: "req2GpuA", count: 2, selectors: singleAttrSelector(exampleDriver, gpuAttribute, gpuTypeA)}
	req1GpuB := testDeviceRequest{name: "req1GpuB", count: 1, selectors: singleAttrSelector(exampleDriver, gpuAttribute, gpuTypeB)}
	req1Nic := testDeviceRequest{name: "req1Nic", count: 1, selectors: singleAttrSelector(exampleDriver, nicAttribute, nicTypeA)}
	req1Global := testDeviceRequest{name: "req1Global", count: 1, selectors: singleAttrSelector(exampleDriver, globalDevAttribute, globalDevTypeA)}

	sharedGpuBClaim := testResourceClaim("sharedGpuBClaim", nil, "", []testDeviceRequest{req1GpuB}, nil, nil)
	sharedAllocatedGlobalClaim := testResourceClaim("sharedGlobalClaim", nil, "", []testDeviceRequest{req1Global}, []testAllocation{{request: req1Global.name, driver: exampleDriver, pool: "global-pool", device: globalDevice + "-0"}}, nil)

	testCases := map[string]struct {
		nodeGroups           map[*testNodeGroupDef]int
		templatePods         map[string][]testPod
		pods                 []testPod
		extraResourceSlices  []*resourceapi.ResourceSlice
		extraResourceClaims  []*resourceapi.ResourceClaim
		expectedScaleUps     map[string]int
		expectedScaleDowns   map[string][]string
		expectedNoScaleUps   []noScaleUpDef
		expectedNoScaleDowns []noScaleDownDef
	}{
		"scale-up: one pod per node, one device per node": {
			// 1xGPU nodes, pods requesting 1xGPU: 1 scheduled, 3 unschedulable -> 3 nodes needed
			nodeGroups: map[*testNodeGroupDef]int{node1GpuA1slice: 1},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 3, []testDeviceRequest{req1GpuA}),
				scheduledPod(baseSmallPod, "scheduled-0", node1GpuA1slice.name+"-0", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-0"}}),
			),
			expectedScaleUps: map[string]int{node1GpuA1slice.name: 3},
		},
		"scale-up: multiple pods per node, pods requesting one device": {
			// 3xGPU nodes, pods requesting 1xGPU: 2 scheduled, 10 unschedulable -> 3 nodes needed
			nodeGroups: map[*testNodeGroupDef]int{node3GpuA1slice: 1},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 10, []testDeviceRequest{req1GpuA}),
				scheduledPod(baseSmallPod, "scheduled-0", node3GpuA1slice.name+"-0", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-0"}}),
				scheduledPod(baseSmallPod, "scheduled-1", node3GpuA1slice.name+"-0", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-1"}}),
			),
			expectedScaleUps: map[string]int{node3GpuA1slice.name: 3},
		},
		"scale-up: multiple pods per node, pods requesting multiple identical devices": {
			// 3xGPU nodes, 1 scheduled pod requesting 1xGPU, 4 unschedulable pods requesting 2xGPU -> 3 nodes needed"
			nodeGroups: map[*testNodeGroupDef]int{node3GpuA1slice: 1},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 4, []testDeviceRequest{req2GpuA}),
				scheduledPod(baseSmallPod, "scheduled-0", node3GpuA1slice.name+"-0", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-0"}}),
			),
			expectedScaleUps: map[string]int{node3GpuA1slice.name: 3},
		},
		"scale-up: multiple devices in one slice work correctly": {
			// 3xGPU nodes, 1 scheduled pod requesting 2xGPU, 5 unschedulable pods requesting 1xGPU -> 2 nodes needed"
			nodeGroups: map[*testNodeGroupDef]int{node3GpuA1slice: 1},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 5, []testDeviceRequest{req1GpuA}),
				scheduledPod(baseSmallPod, "scheduled-0", node3GpuA1slice.name+"-0", map[*testDeviceRequest][]string{&req2GpuA: {gpuDevice + "-0", gpuDevice + "-1"}}),
			),
			expectedScaleUps: map[string]int{node3GpuA1slice.name: 2},
		},
		"scale-up: multiple devices in multiple slices work correctly": {
			// 3xGPU nodes, 1 scheduled pod requesting 2xGPU, 5 unschedulable pods requesting 1xGPU -> 2 nodes needed"
			nodeGroups: map[*testNodeGroupDef]int{node3GpuA3slice: 1},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 5, []testDeviceRequest{req1GpuA}),
				scheduledPod(baseSmallPod, "scheduled-0", node3GpuA3slice.name+"-0", map[*testDeviceRequest][]string{&req2GpuA: {gpuDevice + "-0", gpuDevice + "-1"}}),
			),
			expectedScaleUps: map[string]int{node3GpuA3slice.name: 2},
		},
		"scale-up: one pod per node, pods requesting multiple different devices": {
			nodeGroups: map[*testNodeGroupDef]int{node1Gpu1Nic1slice: 1},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 3, []testDeviceRequest{req1GpuA, req1Nic}),
				scheduledPod(baseSmallPod, "scheduled-0", node1Gpu1Nic1slice.name+"-0", map[*testDeviceRequest][]string{&req1Nic: {nicDevice + "-0"}, &req1GpuA: {gpuDevice + "-0"}}),
			),
			expectedScaleUps: map[string]int{node1Gpu1Nic1slice.name: 3},
		},
		"scale-up: scale from 0 nodes in a node group": {
			nodeGroups: map[*testNodeGroupDef]int{node1Gpu1Nic1slice: 0},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 3, []testDeviceRequest{req1GpuA, req1Nic}),
			),
			expectedScaleUps: map[string]int{node1Gpu1Nic1slice.name: 3},
		},
		"scale-up: scale from 0 nodes in a node group, with pods on the template nodes consuming DRA resources": {
			nodeGroups: map[*testNodeGroupDef]int{node3GpuA1slice: 0},
			templatePods: map[string][]testPod{
				node3GpuA1slice.name: {
					scheduledPod(baseSmallPod, "template-0", node3GpuA1slice.name+"-template", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-0"}}),
					scheduledPod(baseSmallPod, "template-1", node3GpuA1slice.name+"-template", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-1"}}),
				},
			},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 3, []testDeviceRequest{req1GpuA}),
			),
			expectedScaleUps: map[string]int{node3GpuA1slice.name: 3},
		},
		"scale-up: scale from 0 nodes in a node group, with pods on the template nodes consuming DRA resources, including shared claims": {
			nodeGroups:          map[*testNodeGroupDef]int{node3GpuA1slice: 0},
			extraResourceClaims: []*resourceapi.ResourceClaim{sharedAllocatedGlobalClaim},
			templatePods: map[string][]testPod{
				node3GpuA1slice.name: {
					scheduledPod(baseSmallPod, "template-0", node3GpuA1slice.name+"-template", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-0"}}, sharedAllocatedGlobalClaim),
					scheduledPod(baseSmallPod, "template-1", node3GpuA1slice.name+"-template", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-1"}}, sharedAllocatedGlobalClaim),
				},
			},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 3, []testDeviceRequest{req1GpuA}, sharedAllocatedGlobalClaim),
			),
			expectedScaleUps: map[string]int{node3GpuA1slice.name: 3},
		},
		"no scale-up: pods requesting multiple different devices, but they're on different nodes": {
			nodeGroups: map[*testNodeGroupDef]int{node1GpuA1slice: 1, node1Nic1slice: 1},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 3, []testDeviceRequest{req1GpuA, req1Nic}),
			),
		},
		"scale-up: pods requesting a shared, unallocated claim": {
			extraResourceClaims: []*resourceapi.ResourceClaim{sharedGpuBClaim},
			nodeGroups:          map[*testNodeGroupDef]int{node1GpuB1slice: 1},
			pods: append(
				unscheduledPods(baseSmallPod, "unschedulable", 13, nil, sharedGpuBClaim),
				scheduledPod(baseSmallPod, "scheduled-0", node1GpuB1slice.name+"-0", map[*testDeviceRequest][]string{&req1GpuB: {gpuDevice + "-0"}}),
			),
			// All pods request a shared claim to a node-local resource - only 1 node can work.
			expectedScaleUps: map[string]int{node1GpuB1slice.name: 1},
			// The claim is bound to a node, and the node only fits 10 pods because of CPU. The 3 extra pods shouldn't trigger a scale-up.
			expectedNoScaleUps: []noScaleUpDef{{podNamePrefix: "unschedulable", podCount: 3}},
		},
		"scale-down: empty single-device nodes": {
			nodeGroups: map[*testNodeGroupDef]int{node1GpuA1slice: 3},
			pods: []testPod{
				scheduledPod(baseBigPod, "scheduled-0", node1GpuA1slice.name+"-1", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-0"}}),
			},
			expectedScaleDowns:   map[string][]string{node1GpuA1slice.name: {node1GpuA1slice.name + "-0", node1GpuA1slice.name + "-2"}},
			expectedNoScaleDowns: []noScaleDownDef{{nodeName: node1GpuA1slice.name + "-1", reason: simulator.NotUnderutilized}},
		},
		"scale-down: single-device nodes with drain": {
			nodeGroups: map[*testNodeGroupDef]int{node3GpuA1slice: 3},
			pods: []testPod{
				scheduledPod(baseSmallPod, "scheduled-0", node3GpuA1slice.name+"-0", map[*testDeviceRequest][]string{&req2GpuA: {gpuDevice + "-0", gpuDevice + "-1"}}),
				scheduledPod(baseSmallPod, "scheduled-1", node3GpuA1slice.name+"-1", map[*testDeviceRequest][]string{&req2GpuA: {gpuDevice + "-0", gpuDevice + "-1"}}),
				scheduledPod(baseSmallPod, "scheduled-2", node3GpuA1slice.name+"-2", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-0"}}),
				scheduledPod(baseSmallPod, "scheduled-3", node3GpuA1slice.name+"-2", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-1"}}),
			},
			expectedScaleDowns: map[string][]string{node3GpuA1slice.name: {node3GpuA1slice.name + "-2"}},
			expectedNoScaleDowns: []noScaleDownDef{
				{nodeName: node1GpuA1slice.name + "-0", reason: simulator.NoPlaceToMovePods},
				{nodeName: node1GpuA1slice.name + "-1", reason: simulator.NoPlaceToMovePods},
			},
		},
		"no scale-down: no place to reschedule": {
			nodeGroups: map[*testNodeGroupDef]int{node3GpuA1slice: 3},
			pods: []testPod{
				scheduledPod(baseBigPod, "scheduled-0", node3GpuA1slice.name+"-0", map[*testDeviceRequest][]string{&req2GpuA: {gpuDevice + "-0", gpuDevice + "-1"}}),
				scheduledPod(baseBigPod, "scheduled-1", node3GpuA1slice.name+"-1", map[*testDeviceRequest][]string{&req2GpuA: {gpuDevice + "-0", gpuDevice + "-1"}}),
				scheduledPod(baseBigPod, "scheduled-2", node3GpuA1slice.name+"-1", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-2"}}),
				scheduledPod(baseSmallPod, "scheduled-3", node3GpuA1slice.name+"-2", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-0"}}),
				scheduledPod(baseSmallPod, "scheduled-4", node3GpuA1slice.name+"-2", map[*testDeviceRequest][]string{&req1GpuA: {gpuDevice + "-1"}}),
			},
			expectedNoScaleDowns: []noScaleDownDef{
				{nodeName: node3GpuA1slice.name + "-0", reason: simulator.NoPlaceToMovePods},
				{nodeName: node3GpuA1slice.name + "-1", reason: simulator.NotUnderutilized},
				{nodeName: node3GpuA1slice.name + "-2", reason: simulator.NoPlaceToMovePods},
			},
		},
		// TODO(DRA): Write.
		//"expendable pods ResourceClaims are removed from the snapshot": {
		//
		//},
		//"pods with nominatedNodeName using ResourceClaims are correctly scheduled": {
		//
		//},
		// TODO(DRA): Write.
		//"duplicating unschedulabel pods using ResourceClaims works correctly": {
		//
		//},
	}

	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			var nodeGroups []*nodeGroup
			var allNodes []*apiv1.Node
			allResourceSlices := tc.extraResourceSlices
			for nodeGroupDef, count := range tc.nodeGroups {
				var nodes []*apiv1.Node
				for i := range count {
					node := BuildTestNode(fmt.Sprintf("%s-%d", nodeGroupDef.name, i), nodeGroupDef.cpu, nodeGroupDef.mem)
					SetNodeReadyState(node, true, now)
					nodes = append(nodes, node)
					if nodeGroupDef.slicesTemplateFunc != nil {
						slicesForNode := nodeGroupDef.slicesTemplateFunc(node.Name)
						allResourceSlices = append(allResourceSlices, slicesForNode...)
					}
				}

				templateNode := BuildTestNode(fmt.Sprintf("%s-template", nodeGroupDef.name), nodeGroupDef.cpu, nodeGroupDef.mem)
				templateSlices := nodeGroupDef.slicesTemplateFunc(templateNode.Name)
				templateNodeInfo := framework.NewNodeInfo(templateNode, templateSlices)
				for _, pod := range tc.templatePods[nodeGroupDef.name] {
					WithDSController()(pod.pod)
					templateNodeInfo.AddPod(&framework.PodInfo{Pod: pod.pod, NeededResourceClaims: pod.claims})
				}

				nodeGroups = append(nodeGroups, &nodeGroup{
					name:     nodeGroupDef.name,
					template: templateNodeInfo,
					min:      0,
					max:      10,
					nodes:    nodes,
				})
				allNodes = append(allNodes, nodes...)
			}

			var allPods []*apiv1.Pod
			allResourceClaims := tc.extraResourceClaims
			for _, pod := range tc.pods {
				allPods = append(allPods, pod.pod)
				allResourceClaims = append(allResourceClaims, pod.claims...)
			}

			allExpectedScaleDowns := 0

			mocks := newCommonMocks()
			mocks.readyNodeLister.SetNodes(allNodes)
			mocks.allNodeLister.SetNodes(allNodes)
			mocks.daemonSetLister.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil)
			mocks.podDisruptionBudgetLister.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil)
			mocks.allPodLister.On("List").Return(allPods, nil)
			for nodeGroup, delta := range tc.expectedScaleUps {
				mocks.onScaleUp.On("ScaleUp", nodeGroup, delta).Return(nil).Once()
			}
			for nodeGroup, nodes := range tc.expectedScaleDowns {
				for _, node := range nodes {
					mocks.onScaleDown.On("ScaleDown", nodeGroup, node).Return(nil).Once()
					allExpectedScaleDowns++
				}
			}
			mocks.resourceClaimLister = &fakeResourceClaimLister{claims: allResourceClaims}
			mocks.resourceSliceLister = &fakeResourceSliceLister{slices: allResourceSlices}
			mocks.deviceClassLister = &fakeDeviceClassLister{devices: []*resourceapi.DeviceClass{defaultDeviceClass}}

			setupConfig := &autoscalerSetupConfig{
				autoscalingOptions: config.AutoscalingOptions{
					NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
						ScaleDownUnneededTime:         time.Minute,
						ScaleDownUnreadyTime:          time.Minute,
						ScaleDownUtilizationThreshold: 0.7,
						MaxNodeProvisionTime:          time.Hour,
					},
					EstimatorName:                  estimator.BinpackingEstimatorName,
					MaxBinpackingTime:              1 * time.Hour,
					MaxNodeGroupBinpackingDuration: 1 * time.Hour,
					ScaleDownSimulationTimeout:     1 * time.Hour,
					OkTotalUnreadyCount:            9999999,
					MaxTotalUnreadyPercentage:      1.0,
					ScaleDownEnabled:               true,
					MaxScaleDownParallelism:        10,
					MaxDrainParallelism:            10,
					NodeDeletionBatcherInterval:    0 * time.Second,
					NodeDeleteDelayAfterTaint:      1 * time.Millisecond,
					MaxNodesTotal:                  1000,
					MaxCoresTotal:                  1000,
					MaxMemoryTotal:                 100000000,
					SchedulerConfig:                schedConfig,
					ScaleUpFromZero:                true,
					EnableDynamicResources:         true,
				},
				nodeGroups:             nodeGroups,
				nodeStateUpdateTime:    now,
				mocks:                  mocks,
				optionsBlockDefaulting: true,
				nodesDeleted:           make(chan bool, allExpectedScaleDowns),
			}

			autoscaler, err := setupAutoscaler(setupConfig)
			assert.NoError(t, err)

			scaleUpProcessor := &fakeScaleUpStatusProcessor{}
			scaleDownProcessor := &fakeScaleDownStatusProcessor{}
			autoscaler.processors.ScaleUpStatusProcessor = scaleUpProcessor
			autoscaler.processors.ScaleDownStatusProcessor = scaleDownProcessor

			if len(tc.expectedScaleDowns) > 0 {
				err = autoscaler.RunOnce(now)
				assert.NoError(t, err)
			}

			err = autoscaler.RunOnce(now.Add(2 * time.Minute))
			assert.NoError(t, err)

			for range allExpectedScaleDowns {
				select {
				case <-setupConfig.nodesDeleted:
					return
				case <-time.After(20 * time.Second):
					t.Fatalf("Node deletes not finished")
				}
			}

			if len(tc.expectedNoScaleUps) > 0 || len(tc.expectedNoScaleDowns) > 0 {
				err = autoscaler.RunOnce(now.Add(4 * time.Minute))
				assert.NoError(t, err)

				for _, noScaleUp := range tc.expectedNoScaleUps {
					assertNoScaleUpReported(t, scaleUpProcessor.lastStatus, noScaleUp)
				}
				for _, noScaleDown := range tc.expectedNoScaleDowns {
					assertNoScaleDownReported(t, scaleDownProcessor.lastStatus, noScaleDown)
				}
			}

			mock.AssertExpectationsForObjects(t, setupConfig.mocks.allPodLister,
				setupConfig.mocks.podDisruptionBudgetLister, setupConfig.mocks.daemonSetLister, setupConfig.mocks.onScaleUp, setupConfig.mocks.onScaleDown)
		})
	}
}

func assertNoScaleUpReported(t *testing.T, status *status.ScaleUpStatus, wantNoScaleUp noScaleUpDef) {
	matchingPrefix := 0
	for _, noScaleUpPod := range status.PodsRemainUnschedulable {
		if wantNoScaleUp.podName != "" && wantNoScaleUp.podName == noScaleUpPod.Pod.Name {
			return
		}
		if wantNoScaleUp.podNamePrefix != "" && strings.HasPrefix(noScaleUpPod.Pod.Name, wantNoScaleUp.podNamePrefix) {
			matchingPrefix++
		}
	}
	assert.Equal(t, wantNoScaleUp.podCount, matchingPrefix)
}

func assertNoScaleDownReported(t *testing.T, status *scaledownstatus.ScaleDownStatus, wantNoScaleDown noScaleDownDef) {
	matchingPrefix := 0
	for _, unremovableNode := range status.UnremovableNodes {
		if wantNoScaleDown.nodeName != "" && wantNoScaleDown.nodeName == unremovableNode.Node.Name {
			assert.Equal(t, wantNoScaleDown.reason, unremovableNode.Reason)
			return
		}
		if wantNoScaleDown.nodeNamePrefix != "" && strings.HasPrefix(unremovableNode.Node.Name, wantNoScaleDown.nodeNamePrefix) {
			assert.Equal(t, wantNoScaleDown.reason, unremovableNode.Reason)
			matchingPrefix++
		}
	}
	assert.Equal(t, wantNoScaleDown.nodeCount, matchingPrefix)
}

func singleAttrSelector(driver, attribute, value string) []string {
	return []string{fmt.Sprintf("device.attributes[%q].%s == %q", driver, attribute, value)}
}

func unscheduledPods(basePod *apiv1.Pod, podBaseName string, podCount int, requests []testDeviceRequest, sharedClaims ...*resourceapi.ResourceClaim) []testPod {
	var result []testPod
	for i := range podCount {
		pod := unscheduledPod(basePod, fmt.Sprintf("%s-%d", podBaseName, i), podBaseName, requests, sharedClaims...)
		result = append(result, pod)
	}
	return result
}

func scheduledPod(basePod *apiv1.Pod, podName, nodeName string, requests map[*testDeviceRequest][]string, sharedClaims ...*resourceapi.ResourceClaim) testPod {
	allocations := map[string][]testAllocation{}
	var reqs []testDeviceRequest
	for request, devices := range requests {
		reqs = append(reqs, *request)
		for _, device := range devices {
			allocations[request.name] = append(allocations[request.name], testAllocation{
				request: request.name,
				driver:  exampleDriver,
				pool:    nodeName,
				device:  device,
			})
		}
	}
	return createTestPod(basePod, podName, podName+"-controller", nodeName, len(requests), reqs, allocations, sharedClaims)
}

func unscheduledPod(basePod *apiv1.Pod, podName, controllerName string, requests []testDeviceRequest, extraClaims ...*resourceapi.ResourceClaim) testPod {
	pod := createTestPod(basePod, podName, controllerName, "", len(requests), requests, nil, extraClaims)
	MarkUnschedulable()(pod.pod)
	return pod
}

func createTestPod(basePod *apiv1.Pod, podName, controllerName, nodeName string, claimCount int, requests []testDeviceRequest, allocations map[string][]testAllocation, sharedClaims []*resourceapi.ResourceClaim) testPod {
	pod := basePod.DeepCopy()
	pod.Name = podName
	pod.UID = types.UID(podName)
	pod.Spec.NodeName = nodeName
	pod.OwnerReferences = GenerateOwnerReferences(controllerName, "ReplicaSet", "apps/v1", types.UID(controllerName))
	claims := resourceClaimsForPod(pod, nodeName, claimCount, requests, allocations)
	for i, claim := range claims {
		claimRef := fmt.Sprintf("claim-ref-%d", i)
		claimTemplateName := fmt.Sprintf("%s-%s-template", pod.Name, claimRef) // For completeness only.
		WithResourceClaim(claimRef, claim.Name, claimTemplateName)(pod)
	}
	for i, extraClaim := range sharedClaims {
		WithResourceClaim(fmt.Sprintf("claim-ref-extra-%d", i), extraClaim.Name, "")(pod)
	}
	return testPod{pod: pod, claims: claims}
}

func resourceClaimsForPod(pod *apiv1.Pod, nodeName string, claimCount int, requests []testDeviceRequest, allocations map[string][]testAllocation) []*resourceapi.ResourceClaim {
	if claimCount == 0 || len(requests) == 0 {
		return nil
	}

	slices.SortFunc(requests, func(a, b testDeviceRequest) int {
		if a.name < b.name {
			return -1
		}
		if a.name == b.name {
			return 0
		}
		return 1
	})

	requestsPerClaim := int64(math.Ceil(float64(len(requests)) / float64(claimCount)))

	requestIndex := 0
	var claims []*resourceapi.ResourceClaim
	for claimIndex := range claimCount {
		name := fmt.Sprintf("%s-claim-%d", pod.Name, claimIndex)

		var claimRequests []testDeviceRequest
		var claimAllocations []testAllocation
		for range requestsPerClaim {
			request := requests[requestIndex]
			claimRequests = append(claimRequests, request)
			if allocs, found := allocations[request.name]; found {
				claimAllocations = append(claimAllocations, allocs...)
			}

			requestIndex++
			if requestIndex >= len(requests) {
				break
			}
		}

		claims = append(claims, testResourceClaim(name, pod, nodeName, claimRequests, claimAllocations, nil))
	}

	return claims
}

func testResourceClaim(claimName string, owningPod *apiv1.Pod, nodeName string, requests []testDeviceRequest, allocations []testAllocation, reservedFor []*apiv1.Pod) *resourceapi.ResourceClaim {
	var deviceRequests []resourceapi.DeviceRequest
	for _, request := range requests {
		var selectors []resourceapi.DeviceSelector
		for _, selector := range request.selectors {
			selectors = append(selectors, resourceapi.DeviceSelector{CEL: &resourceapi.CELDeviceSelector{Expression: selector}})
		}
		deviceRequest := resourceapi.DeviceRequest{
			Name:            request.name,
			DeviceClassName: "default-class",
			AdminAccess:     false,
			Selectors:       selectors,
		}
		if request.all {
			deviceRequest.AllocationMode = resourceapi.DeviceAllocationModeAll
		} else {
			deviceRequest.AllocationMode = resourceapi.DeviceAllocationModeExactCount
			deviceRequest.Count = request.count
		}
		deviceRequests = append(deviceRequests, deviceRequest)
	}

	claim := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimName,
			Namespace: "default",
			UID:       types.UID(claimName),
		},
		Spec: resourceapi.ResourceClaimSpec{
			Devices: resourceapi.DeviceClaim{
				Requests: deviceRequests,
			},
		},
	}
	if owningPod != nil {
		claim.OwnerReferences = GenerateOwnerReferences(owningPod.Name, "Pod", "v1", owningPod.UID)
	}
	if len(allocations) > 0 {
		var deviceAllocations []resourceapi.DeviceRequestAllocationResult
		for _, allocation := range allocations {
			deviceAllocations = append(deviceAllocations, resourceapi.DeviceRequestAllocationResult{
				Driver:  allocation.driver,
				Request: allocation.request,
				Pool:    allocation.pool,
				Device:  allocation.device,
			})
		}
		var nodeSelector *apiv1.NodeSelector
		if nodeName != "" {
			nodeSelector = &apiv1.NodeSelector{
				NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
					MatchFields: []apiv1.NodeSelectorRequirement{
						{
							Key:      "metadata.name",
							Operator: apiv1.NodeSelectorOpIn,
							Values:   []string{nodeName},
						},
					},
				}},
			}
		}
		var podReservations []resourceapi.ResourceClaimConsumerReference
		if owningPod != nil {
			podReservations = []resourceapi.ResourceClaimConsumerReference{
				{
					APIGroup: "",
					Resource: "pods",
					Name:     owningPod.Name,
					UID:      owningPod.UID,
				},
			}
		} else {
			for _, pod := range podReservations {
				podReservations = append(podReservations, resourceapi.ResourceClaimConsumerReference{
					APIGroup: "",
					Resource: "pods",
					Name:     pod.Name,
					UID:      pod.UID,
				})
			}
		}
		claim.Status = resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{
					Results: deviceAllocations,
				},
				NodeSelector: nodeSelector,
			},
			ReservedFor: podReservations,
		}
	}
	return claim
}

func testDevices(namePrefix string, count int, attributes map[string]string, capacity map[string]string) []testDevice {
	var result []testDevice
	for i := range count {
		result = append(result, testDevice{name: fmt.Sprintf("%s-%d", namePrefix, i), attributes: attributes, capacity: capacity})
	}
	return result
}

func nodeTemplateResourceSlices(driver string, poolSliceCount, poolGen int64, deviceDefs []testDevice) func(nodeName string) []*resourceapi.ResourceSlice {
	return func(nodeName string) []*resourceapi.ResourceSlice {
		return testResourceSlices(driver, nodeName, poolSliceCount, poolGen, sliceNodeAvailability{node: nodeName}, deviceDefs)
	}
}

func testResourceSlices(driver, poolName string, poolSliceCount, poolGen int64, avail sliceNodeAvailability, deviceDefs []testDevice) []*resourceapi.ResourceSlice {
	var slices []*resourceapi.ResourceSlice
	for sliceIndex := range poolSliceCount {
		slice := &resourceapi.ResourceSlice{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s-gen%d-slice%d", driver, poolName, poolGen, sliceIndex),
			},
			Spec: resourceapi.ResourceSliceSpec{
				Driver: driver,
				Pool: resourceapi.ResourcePool{
					Name:               poolName,
					Generation:         poolGen,
					ResourceSliceCount: poolSliceCount,
				},
			},
		}

		if avail.node != "" {
			slice.Spec.NodeName = avail.node
		} else if avail.all {
			slice.Spec.AllNodes = true
		} else if len(avail.nodes) > 0 {
			slice.Spec.NodeSelector = &apiv1.NodeSelector{
				NodeSelectorTerms: []apiv1.NodeSelectorTerm{
					{MatchFields: []apiv1.NodeSelectorRequirement{{Key: "metadata.name", Operator: apiv1.NodeSelectorOpIn, Values: avail.nodes}}},
				},
			}
		}

		slices = append(slices, slice)
	}

	var devices []resourceapi.Device
	for _, deviceDef := range deviceDefs {
		device := resourceapi.Device{
			Name: deviceDef.name,
			Basic: &resourceapi.BasicDevice{
				Attributes: map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{},
				Capacity:   map[resourceapi.QualifiedName]resource.Quantity{},
			},
		}
		for name, val := range deviceDef.attributes {
			val := val
			device.Basic.Attributes[resourceapi.QualifiedName(driver+"/"+name)] = resourceapi.DeviceAttribute{StringValue: &val}
		}
		for name, quantity := range deviceDef.capacity {
			device.Basic.Capacity[resourceapi.QualifiedName(name)] = resource.MustParse(quantity)
		}
		devices = append(devices, device)
	}

	devPerSlice := int64(math.Ceil(float64(len(devices)) / float64(poolSliceCount)))
	addedToSlice := int64(0)
	sliceIndex := 0
	for _, device := range devices {
		if addedToSlice >= devPerSlice {
			sliceIndex += 1
			addedToSlice = 0
		}
		slice := slices[sliceIndex]
		slice.Spec.Devices = append(slice.Spec.Devices, device)
		addedToSlice += 1
	}
	return slices
}
