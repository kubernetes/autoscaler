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

package resource

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodeinfosprovider"
	processorstest "k8s.io/autoscaler/cluster-autoscaler/processors/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	utils_test "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

type nodeGroupConfig struct {
	Name string
	Min  int
	Max  int
	Size int
	CPU  int64
	Mem  int64
}

type deltaForNodeTestCase struct {
	nodeGroupConfig nodeGroupConfig
	expectedOutput  Delta
}

func TestDeltaForNode(t *testing.T) {
	testCases := []deltaForNodeTestCase{
		{
			nodeGroupConfig: nodeGroupConfig{Name: "ng1", Min: 3, Max: 10, Size: 5, CPU: 8, Mem: 16},
			expectedOutput:  Delta{"cpu": 8, "memory": 16},
		},
		{
			nodeGroupConfig: nodeGroupConfig{Name: "ng2", Min: 1, Max: 20, Size: 9, CPU: 4, Mem: 32},
			expectedOutput:  Delta{"cpu": 4, "memory": 32},
		},
	}

	for _, testCase := range testCases {
		cp := testprovider.NewTestCloudProviderBuilder().Build()
		autoscalingCtx := newAutoscalingContext(t, cp)
		processors := processorstest.NewTestProcessors(&autoscalingCtx)

		ng := testCase.nodeGroupConfig
		group, nodes := newNodeGroup(t, cp, ng.Name, ng.Min, ng.Max, ng.Size, ng.CPU, ng.Mem)
		err := autoscalingCtx.ClusterSnapshot.SetClusterState(nodes, nil, nil)
		assert.NoError(t, err)
		nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&autoscalingCtx, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, time.Now())

		rm := NewManager(processors.CustomResourcesProcessor)
		delta, err := rm.DeltaForNode(&autoscalingCtx, nodeInfos[ng.Name], group)
		assert.NoError(t, err)
		assert.Equal(t, testCase.expectedOutput, delta)
	}
}

type resourceLeftTestCase struct {
	nodeGroupConfig nodeGroupConfig
	clusterCPULimit int64
	clusterMemLimit int64
	expectedOutput  Limits
}

func TestResourcesLeft(t *testing.T) {
	testCases := []resourceLeftTestCase{
		{
			// cpu left: 1000 - 8 * 5 = 960; memory left: 1000 - 16 * 5 = 920
			nodeGroupConfig: nodeGroupConfig{Name: "ng1", Min: 3, Max: 10, Size: 5, CPU: 8, Mem: 16},
			clusterCPULimit: 1000,
			clusterMemLimit: 1000,
			expectedOutput:  Limits{"cpu": 960, "memory": 920},
		},
		{
			// cpu left: 1000 - 4 * 100 = 600; memory left: 1000 - 8 * 100 = 200
			nodeGroupConfig: nodeGroupConfig{Name: "ng2", Min: 3, Max: 100, Size: 100, CPU: 4, Mem: 8},
			clusterCPULimit: 1000,
			clusterMemLimit: 1000,
			expectedOutput:  Limits{"cpu": 600, "memory": 200},
		},
	}

	for _, testCase := range testCases {
		cp := newCloudProvider(t, 1000, 1000)
		autoscalingCtx := newAutoscalingContext(t, cp)
		processors := processorstest.NewTestProcessors(&autoscalingCtx)

		ng := testCase.nodeGroupConfig
		_, nodes := newNodeGroup(t, cp, ng.Name, ng.Min, ng.Max, ng.Size, ng.CPU, ng.Mem)
		err := autoscalingCtx.ClusterSnapshot.SetClusterState(nodes, nil, nil)
		assert.NoError(t, err)
		nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&autoscalingCtx, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, time.Now())

		rm := NewManager(processors.CustomResourcesProcessor)
		left, err := rm.ResourcesLeft(&autoscalingCtx, nodeInfos, nodes)
		assert.NoError(t, err)
		assert.Equal(t, testCase.expectedOutput, left)
	}
}

type applyLimitsTestCase struct {
	nodeGroupConfig nodeGroupConfig
	resourcesLeft   Limits
	newNodeCount    int
	expectedOutput  int
}

func TestApplyLimits(t *testing.T) {
	testCases := []applyLimitsTestCase{
		{
			nodeGroupConfig: nodeGroupConfig{Name: "ng1", Min: 3, Max: 10, Size: 5, CPU: 8, Mem: 16},
			resourcesLeft:   Limits{"cpu": 80, "memory": 160},
			newNodeCount:    10,
			expectedOutput:  10,
		},
		{
			nodeGroupConfig: nodeGroupConfig{Name: "ng2", Min: 3, Max: 10, Size: 5, CPU: 8, Mem: 16},
			resourcesLeft:   Limits{"cpu": 80, "memory": 100},
			newNodeCount:    10,
			expectedOutput:  6, // limited by memory: 100 / 16 = 6
		},
		{
			nodeGroupConfig: nodeGroupConfig{Name: "ng3", Min: 3, Max: 10, Size: 5, CPU: 8, Mem: 16},
			resourcesLeft:   Limits{"cpu": 39, "memory": 160},
			newNodeCount:    10,
			expectedOutput:  4, // limited by CPU: 39 / 8 = 4
		},
		{
			nodeGroupConfig: nodeGroupConfig{Name: "ng4", Min: 3, Max: 10, Size: 5, CPU: 8, Mem: 16},
			resourcesLeft:   Limits{"cpu": 40, "memory": 80},
			newNodeCount:    10,
			expectedOutput:  5, // limited by CPU and memory
		},
	}

	for _, testCase := range testCases {
		cp := testprovider.NewTestCloudProviderBuilder().Build()
		autoscalingCtx := newAutoscalingContext(t, cp)
		processors := processorstest.NewTestProcessors(&autoscalingCtx)

		ng := testCase.nodeGroupConfig
		group, nodes := newNodeGroup(t, cp, ng.Name, ng.Min, ng.Max, ng.Size, ng.CPU, ng.Mem)
		err := autoscalingCtx.ClusterSnapshot.SetClusterState(nodes, nil, nil)
		assert.NoError(t, err)
		nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&autoscalingCtx, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, time.Now())

		rm := NewManager(processors.CustomResourcesProcessor)
		newCount, err := rm.ApplyLimits(&autoscalingCtx, testCase.newNodeCount, testCase.resourcesLeft, nodeInfos[testCase.nodeGroupConfig.Name], group)
		assert.NoError(t, err)
		assert.Equal(t, testCase.expectedOutput, newCount)
	}
}

type checkDeltaWithinLimitsTestCase struct {
	resourcesLeft  Limits
	resourcesDelta Delta
	expectedOutput LimitsCheckResult
}

func TestCheckDeltaWithinLimits(t *testing.T) {
	testCases := []checkDeltaWithinLimitsTestCase{
		{
			resourcesLeft:  Limits{"cpu": 10, "memory": 20},
			resourcesDelta: Delta{"cpu": 8, "memory": 16},
			expectedOutput: LimitsCheckResult{Exceeded: false, ExceededResources: []string{}},
		},
		{
			resourcesLeft:  Limits{"cpu": 10, "memory": 20},
			resourcesDelta: Delta{"cpu": 12, "memory": 16},
			expectedOutput: LimitsCheckResult{Exceeded: true, ExceededResources: []string{"cpu"}},
		},
		{
			resourcesLeft:  Limits{"cpu": 10, "memory": 20},
			resourcesDelta: Delta{"cpu": 8, "memory": 32},
			expectedOutput: LimitsCheckResult{Exceeded: true, ExceededResources: []string{"memory"}},
		},
		{
			resourcesLeft:  Limits{"cpu": 10, "memory": 20},
			resourcesDelta: Delta{"cpu": 16, "memory": 96},
			expectedOutput: LimitsCheckResult{Exceeded: true, ExceededResources: []string{"cpu", "memory"}},
		},
	}

	for _, testCase := range testCases {
		result := CheckDeltaWithinLimits(testCase.resourcesLeft, testCase.resourcesDelta)
		assert.Equal(t, testCase.expectedOutput, result)
	}
}

func TestResourceManagerWithGpuResource(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 0, cloudprovider.ResourceNameMemory: 0, "gpu": 0},
		map[string]int64{cloudprovider.ResourceNameCores: 320, cloudprovider.ResourceNameMemory: 640, "gpu": 16},
	)
	provider.SetResourceLimiter(resourceLimiter)

	autoscalingCtx := newAutoscalingContext(t, provider)
	processors := processorstest.NewTestProcessors(&autoscalingCtx)

	n1 := newNode(t, "n1", 8, 16)
	utils_test.AddGpusToNode(n1, 4)
	n1.Labels[provider.GPULabel()] = "gpu"
	provider.AddNodeGroup("ng1", 3, 10, 1)
	provider.AddNode("ng1", n1)
	ng1, err := provider.NodeGroupForNode(n1)
	assert.NoError(t, err)

	nodes := []*corev1.Node{n1}
	err = autoscalingCtx.ClusterSnapshot.SetClusterState(nodes, nil, nil)
	assert.NoError(t, err)
	nodeInfos, _ := nodeinfosprovider.NewDefaultTemplateNodeInfoProvider(nil, false).Process(&autoscalingCtx, nodes, []*appsv1.DaemonSet{}, taints.TaintConfig{}, time.Now())

	rm := NewManager(processors.CustomResourcesProcessor)

	delta, err := rm.DeltaForNode(&autoscalingCtx, nodeInfos["ng1"], ng1)
	assert.Equal(t, int64(8), delta[cloudprovider.ResourceNameCores])
	assert.Equal(t, int64(16), delta[cloudprovider.ResourceNameMemory])
	assert.Equal(t, int64(4), delta["gpu"])

	left, err := rm.ResourcesLeft(&autoscalingCtx, nodeInfos, nodes)
	assert.NoError(t, err)
	assert.Equal(t, Limits{"cpu": 312, "memory": 624, "gpu": 12}, left) // cpu: 320-8*1=312; memory: 640-16*1=624; gpu: 16-4*1=12

	result := CheckDeltaWithinLimits(left, delta)
	assert.False(t, result.Exceeded)
	assert.Zero(t, len(result.ExceededResources))

	newNodeCount, err := rm.ApplyLimits(&autoscalingCtx, 10, left, nodeInfos["ng1"], ng1)
	assert.Equal(t, 3, newNodeCount) // gpu left / grpu per node: 12 / 4 = 3
}

func TestResourceManagerWithDraResource(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 0, cloudprovider.ResourceNameMemory: 0, "gpu.nvidia.com:nvidia l4": 0},
		map[string]int64{cloudprovider.ResourceNameCores: 320, cloudprovider.ResourceNameMemory: 640, "gpu.nvidia.com:nvidia l4": 16},
	)
	provider.SetResourceLimiter(resourceLimiter)

	autoscalingCtx := newAutoscalingContext(t, provider)
	processors := processorstest.NewTestProcessors(&autoscalingCtx)

	n1 := newNode(t, "n1", 8, 16)
	provider.AddNodeGroup("ng1", 3, 10, 1)
	provider.AddNode("ng1", n1)
	ngi := framework.NewTestNodeInfo(n1)
	ngi.LocalResourceSlices = []*resource.ResourceSlice{
		{
			Spec: resource.ResourceSliceSpec{
				Driver:   "gpu.nvidia.com",
				NodeName: ptr.To("n1"),
				Pool: resource.ResourcePool{
					Name:               "n1",
					Generation:         1,
					ResourceSliceCount: 1,
				},
				Devices: []resource.Device{
					{
						Name: "gpu-0",
						Attributes: map[resource.QualifiedName]resource.DeviceAttribute{
							"uuid": {
								StringValue: ptr.To("GPU-be40a35e-4b7b-16cf-09c4-2fd3e9166701"),
							},
							"productName": {
								StringValue: ptr.To("NVIDIA L4"),
							},
						},
					},
					{
						Name: "gpu-0-partition-0",
						Attributes: map[resource.QualifiedName]resource.DeviceAttribute{
							"uuid": {
								StringValue: ptr.To("GPU-be40a35e-4b7b-16cf-09c4-2fd3e9166701"),
							},
							"productName": {
								StringValue: ptr.To("NVIDIA L4"),
							},
						},
					},
					{
						Name: "gpu-0-partition-1",
						Attributes: map[resource.QualifiedName]resource.DeviceAttribute{
							"uuid": {
								StringValue: ptr.To("GPU-be40a35e-4b7b-16cf-09c4-2fd3e9166701"),
							},
							"productName": {
								StringValue: ptr.To("NVIDIA L4"),
							},
						},
					},
					{
						Name: "gpu-1",
						Attributes: map[resource.QualifiedName]resource.DeviceAttribute{
							"uuid": {
								StringValue: ptr.To("GPU-39807e58-5aca-9931-8819-3920ede08201"),
							},
							"productName": {
								StringValue: ptr.To("NVIDIA L4"),
							},
						},
					},
					{
						Name: "gpu-2",
						Attributes: map[resource.QualifiedName]resource.DeviceAttribute{
							"uuid": {
								StringValue: ptr.To("GPU-8ed40f31-c63d-c26b-46ea-5c024ec56802"),
							},
							"productName": {
								StringValue: ptr.To("NVIDIA L4"),
							},
						},
					},
					{
						Name: "gpu-3",
						Attributes: map[resource.QualifiedName]resource.DeviceAttribute{
							"uuid": {
								StringValue: ptr.To("GPU-f095b087-5812-4317-b9d1-193b8af1cd03"),
							},
							"productName": {
								StringValue: ptr.To("NVIDIA L4"),
							},
						},
					},
				},
			},
		},
	}
	provider.SetMachineTemplates(map[string]*framework.NodeInfo{
		"ng1": ngi,
	})

	ng1, err := provider.NodeGroupForNode(n1)
	assert.NoError(t, err)

	nodes := []*corev1.Node{n1}
	err = autoscalingCtx.ClusterSnapshot.SetClusterState(nodes, nil, nil)
	assert.NoError(t, err)

	rm := NewManager(processors.CustomResourcesProcessor)

	delta, err := rm.DeltaForNode(&autoscalingCtx, ngi, ng1)
	assert.Equal(t, int64(8), delta[cloudprovider.ResourceNameCores])
	assert.Equal(t, int64(16), delta[cloudprovider.ResourceNameMemory])
	assert.Equal(t, int64(4), delta["gpu.nvidia.com:nvidia l4"])

	left, err := rm.ResourcesLeft(&autoscalingCtx, map[string]*framework.NodeInfo{"ng1": ngi}, nodes)
	assert.NoError(t, err)
	assert.Equal(t, Limits{"cpu": 312, "memory": 624, "gpu.nvidia.com:nvidia l4": 12}, left) // cpu: 320-8*1=312; memory: 640-16*1=624; gpu: 16-4*1=12
	result := CheckDeltaWithinLimits(left, delta)
	assert.False(t, result.Exceeded)
	assert.Zero(t, len(result.ExceededResources))

	newNodeCount, err := rm.ApplyLimits(&autoscalingCtx, 10, left, ngi, ng1)
	assert.Equal(t, 3, newNodeCount) // gpu left / grpu per node: 12 / 4 = 3
}

func newCloudProvider(t *testing.T, cpu, mem int64) *testprovider.TestCloudProvider {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	assert.NotNil(t, provider)

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 0, cloudprovider.ResourceNameMemory: 0},
		map[string]int64{cloudprovider.ResourceNameCores: cpu, cloudprovider.ResourceNameMemory: mem},
	)
	provider.SetResourceLimiter(resourceLimiter)
	return provider
}

func newAutoscalingContext(t *testing.T, provider cloudprovider.CloudProvider) ca_context.AutoscalingContext {
	podLister := kube_util.NewTestPodLister([]*corev1.Pod{})
	listers := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)
	autoscalingCtx, err := test.NewScaleTestAutoscalingContext(config.AutoscalingOptions{}, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)
	return autoscalingCtx
}

func newNode(t *testing.T, name string, cpu, mem int64) *corev1.Node {
	return utils_test.BuildTestNode(name, cpu*1000, mem)
}

func newNodeGroup(t *testing.T, provider *testprovider.TestCloudProvider, name string, min, max, size int, cpu, mem int64) (cloudprovider.NodeGroup, []*corev1.Node) {
	provider.AddNodeGroup(name, min, max, size)
	nodes := make([]*corev1.Node, 0)
	for index := 0; index < size; index++ {
		node := newNode(t, fmt.Sprint(name, index), cpu, mem)
		provider.AddNode(name, node)
		nodes = append(nodes, node)
	}

	groups := provider.NodeGroups()
	for _, group := range groups {
		if group.Id() == name {
			return group, nodes
		}
	}
	assert.FailNowf(t, "node group %s not found", name)
	return nil, nil
}
