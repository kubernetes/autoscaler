/*
Copyright 2016 The Kubernetes Authors.

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

package utilization

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/kubelet/types"

	"github.com/stretchr/testify/assert"
)

func TestCalculate(t *testing.T) {
	testTime := time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
	pod := BuildTestPod("p1", 100, 200000)
	pod2 := BuildTestPod("p2", -1, -1)

	podWithInitContainers := BuildTestPod("p-init", 100, 200000)
	restartAlways := apiv1.ContainerRestartPolicyAlways
	podWithInitContainers.Spec.InitContainers = []apiv1.Container{
		// restart always
		{
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewMilliQuantity(50, resource.DecimalSI),
					apiv1.ResourceMemory: *resource.NewQuantity(100000, resource.DecimalSI),
				},
			},
			RestartPolicy: &restartAlways,
		},
		// non-restartable, should be excluded from calculations
		{
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewMilliQuantity(5, resource.DecimalSI),
					apiv1.ResourceMemory: *resource.NewQuantity(100, resource.DecimalSI),
				},
			},
		},
	}

	podWithLargeNonRestartableInitContainers := BuildTestPod("p-large-init", 100, 200000)
	podWithLargeNonRestartableInitContainers.Spec.InitContainers = []apiv1.Container{
		// large non-restartable should "overwhelm" the pod utilization calc
		// see formula: https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/753-sidecar-containers#resources-calculation-for-scheduling-and-pod-admission
		{
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:    *resource.NewMilliQuantity(50000, resource.DecimalSI),
					apiv1.ResourceMemory: *resource.NewQuantity(100000000, resource.DecimalSI),
				},
			},
		},
	}
	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})
	nodeInfo := framework.NewTestNodeInfo(node, pod, pod, pod2)

	gpuConfig := getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err := Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)
	assert.Equal(t, 0.1, utilInfo.CpuUtil)

	node2 := BuildTestNode("node2", 2000, -1)
	nodeInfo = framework.NewTestNodeInfo(node2, pod, pod, pod2)

	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	_, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.Error(t, err)

	node3 := BuildTestNode("node3", 2000, 2000000)
	SetNodeReadyState(node3, true, time.Time{})
	nodeInfo = framework.NewTestNodeInfo(node3, pod, podWithInitContainers, podWithLargeNonRestartableInitContainers)

	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 50.25, utilInfo.Utilization, 0.01)
	assert.InEpsilon(t, 25.125, utilInfo.CpuUtil, 0.005)

	daemonSetPod3 := BuildTestPod("p3", 100, 200000)
	daemonSetPod3.OwnerReferences = GenerateOwnerReferences("ds", "DaemonSet", "apps/v1", "")

	daemonSetPod4 := BuildTestPod("p4", 100, 200000)
	daemonSetPod4.OwnerReferences = GenerateOwnerReferences("ds", "CustomDaemonSet", "crd/v1", "")
	daemonSetPod4.Annotations = map[string]string{"cluster-autoscaler.kubernetes.io/daemonset-pod": "true"}

	nodeInfo = framework.NewTestNodeInfo(node, pod, pod, pod2, daemonSetPod3, daemonSetPod4)
	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, true, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.5/10, utilInfo.Utilization, 0.01)

	nodeInfo = framework.NewTestNodeInfo(node, pod, pod2, daemonSetPod3)
	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	terminatedPod := BuildTestPod("podTerminated", 100, 200000)
	terminatedPod.DeletionTimestamp = &metav1.Time{Time: testTime.Add(-10 * time.Minute)}
	nodeInfo = framework.NewTestNodeInfo(node, pod, pod, pod2, terminatedPod)
	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	mirrorPod := BuildTestPod("p4", 100, 200000)
	mirrorPod.Annotations = map[string]string{
		types.ConfigMirrorAnnotationKey: "",
	}

	nodeInfo = framework.NewTestNodeInfo(node, pod, pod, pod2, mirrorPod)
	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, false, true, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/9.0, utilInfo.Utilization, 0.01)

	nodeInfo = framework.NewTestNodeInfo(node, pod, pod2, mirrorPod)
	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 2.0/10, utilInfo.Utilization, 0.01)

	nodeInfo = framework.NewTestNodeInfo(node, pod, mirrorPod, daemonSetPod3)
	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, true, true, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1.0/8.0, utilInfo.Utilization, 0.01)

	gpuNode := BuildTestNode("gpu_node", 2000, 2000000)
	AddGpusToNode(gpuNode, 1)
	gpuPod := BuildTestPod("gpu_pod", 100, 200000)
	RequestGpuForPod(gpuPod, 1)
	TolerateGpuForPod(gpuPod)
	nodeInfo = framework.NewTestNodeInfo(gpuNode, pod, pod, gpuPod)
	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1/1, utilInfo.Utilization, 0.01)

	// Node with Unready GPU
	gpuNode = BuildTestNode("gpu_node", 2000, 2000000)
	AddGpuLabelToNode(gpuNode)
	nodeInfo = framework.NewTestNodeInfo(gpuNode, pod, pod)
	gpuConfig = getGpuConfigFromNode(nodeInfo.Node(), false)
	utilInfo, err = Calculate(nodeInfo, false, false, false, gpuConfig, testTime)
	assert.NoError(t, err)
	assert.Zero(t, utilInfo.Utilization)
}

func TestCalculateWithDynamicResources(t *testing.T) {
	now := time.Date(2024, 12, 4, 0, 0, 0, 0, time.UTC)
	node := BuildTestNode("node", 1000, 1000)
	gpuNode := BuildTestNode("gpuNode", 1000, 1000)
	AddGpusToNode(gpuNode, 1)
	AddGpuLabelToNode(gpuNode)
	gpuConfig := getGpuConfigFromNode(gpuNode, false)
	gpuConfigDra := getGpuConfigFromNode(gpuNode, true)
	pod1 := BuildTestPod("pod1", 250, 0, WithNodeName("node"))
	pod2 := BuildTestPod("pod2", 250, 0, WithNodeName("node"))
	resourceSlice1 := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "node-slice1", UID: "node-slice1"},
		Spec: resourceapi.ResourceSliceSpec{
			Driver: "driver.foo.com",
			Pool: resourceapi.ResourcePool{
				Name:               "node-pool1",
				ResourceSliceCount: 1,
			},
			Devices: []resourceapi.Device{
				{Name: "dev1"},
				{Name: "dev2"},
				{Name: "dev3"},
				{Name: "dev4"},
				{Name: "dev5"},
			},
		},
	}
	resourceSlice2 := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "node-slice2", UID: "node-slice2"},
		Spec: resourceapi.ResourceSliceSpec{
			Driver: "driver.bar.com",
			Pool: resourceapi.ResourcePool{
				Name:               "node-pool2",
				ResourceSliceCount: 1,
			},
			Devices: []resourceapi.Device{
				{Name: "dev1"},
				{Name: "dev2"},
				{Name: "dev3"},
				{Name: "dev4"},
				{Name: "dev5"},
			},
		},
	}
	incompleteResourceSlice := &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{Name: "incompleteResourceSlice", UID: "incompleteResourceSlice"},
		Spec: resourceapi.ResourceSliceSpec{
			Driver: "driver.foo.com",
			Pool: resourceapi.ResourcePool{
				Name:               "node-pool3",
				ResourceSliceCount: 999,
			},
			Devices: []resourceapi.Device{
				{Name: "dev1"},
				{Name: "dev2"},
				{Name: "dev3"},
			},
		},
	}
	pod1Claim1 := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1-claim1", UID: "pod1-claim1"},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{
					Results: []resourceapi.DeviceRequestAllocationResult{
						{Request: "req1", Driver: "driver.foo.com", Pool: "node-pool1", Device: "dev1"},
						{Request: "req2", Driver: "driver.foo.com", Pool: "node-pool1", Device: "dev2"},
						{Request: "req3", Driver: "driver.foo.com", Pool: "node-pool1", Device: "dev3"},
					},
				},
			},
		},
	}
	pod1Claim2 := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1-claim2", UID: "pod1-claim2"},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{
					Results: []resourceapi.DeviceRequestAllocationResult{
						{Request: "req4", Driver: "driver.bar.com", Pool: "node-pool2", Device: "dev1"},
					},
				},
			},
		},
	}
	pod2Claim := &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "pod2-claim", UID: "pod2-claim"},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{
					Results: []resourceapi.DeviceRequestAllocationResult{
						{Request: "req1", Driver: "driver.foo.com", Pool: "node-pool1", Device: "dev4"},
						{Request: "req2", Driver: "driver.bar.com", Pool: "node-pool2", Device: "dev2"},
						{Request: "req3", Driver: "driver.bar.com", Pool: "node-pool2", Device: "dev3"},
					},
				},
			},
		},
	}
	nodeInfoNoDra := framework.NewTestNodeInfo(node, pod1, pod2)
	nodeInfoSlicesNoClaims := framework.NewNodeInfo(node, []*resourceapi.ResourceSlice{resourceSlice1}, framework.NewPodInfo(pod1, nil), framework.NewPodInfo(pod2, nil))
	nodeInfoSlicesAndClaimsPool1Higher := framework.NewNodeInfo(node, []*resourceapi.ResourceSlice{resourceSlice1, resourceSlice2},
		framework.NewPodInfo(pod1, []*resourceapi.ResourceClaim{pod1Claim1, pod1Claim2}),
		framework.NewPodInfo(pod2, []*resourceapi.ResourceClaim{pod2Claim}))
	nodeInfoSlicesAndClaimsPool2Higher := framework.NewNodeInfo(node, []*resourceapi.ResourceSlice{resourceSlice1, resourceSlice2},
		framework.NewPodInfo(pod2, []*resourceapi.ResourceClaim{pod2Claim}))
	nodeInfoIncompleteSlices := framework.NewNodeInfo(node, []*resourceapi.ResourceSlice{resourceSlice1, resourceSlice2, incompleteResourceSlice},
		framework.NewPodInfo(pod1, []*resourceapi.ResourceClaim{pod1Claim1, pod1Claim2}),
		framework.NewPodInfo(pod2, []*resourceapi.ResourceClaim{pod2Claim}))
	nodeInfoGpuAndDra := framework.NewNodeInfo(gpuNode, []*resourceapi.ResourceSlice{resourceSlice1, resourceSlice2},
		framework.NewPodInfo(pod1, []*resourceapi.ResourceClaim{pod1Claim1, pod1Claim2}),
		framework.NewPodInfo(pod2, []*resourceapi.ResourceClaim{pod2Claim}))

	for _, tc := range []struct {
		testName     string
		draEnabled   bool
		nodeInfo     *framework.NodeInfo
		gpuConfig    *cloudprovider.GpuConfig
		wantUtilInfo Info
		wantErr      error
	}{
		{
			testName:     "no DRA resources, DRA disabled -> normal resource util returned",
			nodeInfo:     nodeInfoNoDra,
			draEnabled:   false,
			wantUtilInfo: Info{CpuUtil: 0.5, Utilization: 0.5, ResourceName: apiv1.ResourceCPU},
		},
		{
			testName:     "DRA slices present, but no claims, DRA disabled -> normal resource util returned",
			nodeInfo:     nodeInfoSlicesNoClaims,
			draEnabled:   false,
			wantUtilInfo: Info{CpuUtil: 0.5, Utilization: 0.5, ResourceName: apiv1.ResourceCPU},
		},
		{
			testName:     "DRA slices and claims present, DRA disabled -> normal resource util returned",
			nodeInfo:     nodeInfoSlicesAndClaimsPool1Higher,
			draEnabled:   false,
			wantUtilInfo: Info{CpuUtil: 0.5, Utilization: 0.5, ResourceName: apiv1.ResourceCPU},
		},
		{
			testName:     "no DRA resources, DRA enabled -> normal resource util returned",
			nodeInfo:     nodeInfoNoDra,
			draEnabled:   true,
			wantUtilInfo: Info{CpuUtil: 0.5, Utilization: 0.5, ResourceName: apiv1.ResourceCPU},
		},
		{
			testName:     "DRA slices present, but no claims, DRA enabled -> DRA util returned despite being lower than CPU",
			nodeInfo:     nodeInfoSlicesNoClaims,
			draEnabled:   true,
			wantUtilInfo: Info{DynamicResourceUtil: 0, Utilization: 0, ResourceName: apiv1.ResourceName("driver.foo.com/node-pool1")},
		},
		{
			testName:     "DRA slices and claims present, DRA enabled -> DRA util returned despite being lower than CPU",
			nodeInfo:     nodeInfoSlicesAndClaimsPool2Higher,
			draEnabled:   true,
			wantUtilInfo: Info{DynamicResourceUtil: 0.4, Utilization: 0.4, ResourceName: apiv1.ResourceName("driver.bar.com/node-pool2")},
		},
		{
			testName:     "DRA slices and claims present, DRA enabled -> DRA util returned",
			nodeInfo:     nodeInfoSlicesAndClaimsPool1Higher,
			draEnabled:   true,
			wantUtilInfo: Info{DynamicResourceUtil: 0.8, Utilization: 0.8, ResourceName: apiv1.ResourceName("driver.foo.com/node-pool1")},
		},
		{
			testName:     "DRA slices and claims present, DRA enabled, GPU config passed -> GPU util returned",
			nodeInfo:     nodeInfoGpuAndDra,
			gpuConfig:    gpuConfig,
			draEnabled:   true,
			wantUtilInfo: Info{Utilization: 0, ResourceName: gpuConfig.ExtendedResourceName},
		},
		{
			testName:     "DRA slices and claims present, DRA enabled, DRA GPU config passed -> DRA util returned",
			nodeInfo:     nodeInfoGpuAndDra,
			gpuConfig:    gpuConfigDra,
			draEnabled:   true,
			wantUtilInfo: Info{DynamicResourceUtil: 0.8, Utilization: 0.8, ResourceName: apiv1.ResourceName("driver.foo.com/node-pool1")},
		},
		{
			testName:     "DRA slices and claims present, DRA enabled, error while calculating DRA util -> error returned",
			nodeInfo:     nodeInfoIncompleteSlices,
			draEnabled:   true,
			wantUtilInfo: Info{},
			wantErr:      cmpopts.AnyError,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			utilInfo, err := Calculate(tc.nodeInfo, false, false, tc.draEnabled, tc.gpuConfig, now)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Calculate(): unexpected error (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantUtilInfo, utilInfo); diff != "" {
				t.Errorf("Calculate(): unexpected output (-want +got): %s", diff)
			}
		})
	}
}

func getGpuConfigFromNode(node *apiv1.Node, dra bool) *cloudprovider.GpuConfig {
	gpuLabel := "cloud.google.com/gke-accelerator"
	gpuType, hasGpuLabel := node.Labels[gpuLabel]
	gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[gpu.ResourceNvidiaGPU]
	if hasGpuLabel || (hasGpuAllocatable && !gpuAllocatable.IsZero()) {
		if !dra {
			return &cloudprovider.GpuConfig{
				Label:                gpuLabel,
				Type:                 gpuType,
				ExtendedResourceName: gpu.ResourceNvidiaGPU,
			}
		}

		return &cloudprovider.GpuConfig{
			Label:         gpuLabel,
			Type:          gpuType,
			DraDriverName: "gpu.nvidia.com",
		}
	}
	return nil
}
