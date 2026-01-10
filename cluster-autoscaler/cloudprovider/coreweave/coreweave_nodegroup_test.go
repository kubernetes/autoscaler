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

package coreweave

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func makeTestNodeGroup(name string, uid string, min, max, target int64, options ...NodeGroupOption) *CoreWeaveNodeGroup {
	// Create a dynamic client with a fake nodepool object
	// This is a minimal setup to test the CoreWeaveNodeGroup functionality
	dynamicClientset := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: coreWeaveGroup, Version: coreWeaveVersion, Resource: coreWeaveResource}: "kindList",
		},
	)
	obj := map[string]any{
		"apiVersion": coreWeaveGroup + "/" + coreWeaveVersion,
		"kind":       coreWeaveResource,
		"metadata": map[string]any{
			"name":      name,
			"namespace": "default",
			"uid":       uid,
		},
		"spec": map[string]any{
			"minNodes":    min,
			"maxNodes":    max,
			"targetNodes": target,
		},
	}

	for _, o := range options {
		o(obj)
	}

	u := &unstructured.Unstructured{Object: obj}
	_, _ = dynamicClientset.Resource(CoreWeaveNodeGroupResource).Namespace("default").Create(context.TODO(), u.DeepCopy(), metav1.CreateOptions{})
	// fake a Kubernetes client to simulate node operations
	fakeClient := k8sfake.NewSimpleClientset(
		&apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node1",
				Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
			},
		},
		&apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node2",
				Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
			},
		},
	)
	client := fakeClient
	// Create the CoreWeaveNodePool instance
	np, _ := NewCoreWeaveNodePool(u, dynamicClientset, client)
	return NewCoreWeaveNodeGroup(np)
}

// NodeGroupOption is used for makeTestNodeGroup options
type NodeGroupOption func(obj map[string]any)

func withInstanceType(instanceType string) NodeGroupOption {
	return func(obj map[string]any) {
		obj["spec"].(map[string]any)["instanceType"] = instanceType
	}
}

func withNodeLabels(nodeLabels map[string]string) NodeGroupOption {
	return func(obj map[string]any) {
		labelsInterface := make(map[string]any, len(nodeLabels))
		for k, v := range nodeLabels {
			labelsInterface[k] = v
		}

		obj["spec"].(map[string]any)["nodeLabels"] = labelsInterface
	}
}

func withNodeTaints(nodeTaints []apiv1.Taint) NodeGroupOption {
	return func(obj map[string]any) {
		taints := make([]any, len(nodeTaints))
		for i, t := range nodeTaints {
			tUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&t)
			if err != nil {
				panic(err)
			}
			taints[i] = tUnstructured
		}
		obj["spec"].(map[string]any)["nodeTaints"] = taints
	}
}

func TestId(t *testing.T) {
	ng := makeTestNodeGroup("ng-1", "uid-1", 1, 5, 3)
	if ng.Id() != "uid-1" {
		t.Errorf("expected id 'uid-1', got %s", ng.Id())
	}
}

func TestMinMaxTargetSize(t *testing.T) {
	ng := makeTestNodeGroup("ng-1", "uid-1", 2, 10, 5)
	if ng.MinSize() != 2 {
		t.Errorf("expected min size 2, got %d", ng.MinSize())
	}
	if ng.MaxSize() != 10 {
		t.Errorf("expected max size 10, got %d", ng.MaxSize())
	}
	size, err := ng.TargetSize()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if size != 5 {
		t.Errorf("expected target size 5, got %d", size)
	}
}

func TestIncreaseSize(t *testing.T) {
	ng := makeTestNodeGroup("ng-1", "uid-1", 1, 5, 3)
	err := ng.IncreaseSize(2)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented or nil, got %v", err)
	}
}

func TestDeleteNodes(t *testing.T) {
	initialTargetSize := int64(3)

	testCases := map[string]struct {
		nodesToDelete      []*apiv1.Node
		expectedTargetSize int
		expectedError      error
	}{
		"reduce-target-size-by-one-node": {
			nodesToDelete: []*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node1",
						Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
					},
				},
			},
			expectedTargetSize: 2,
		},
		"reduce-target-size-by-three-node": {
			nodesToDelete: []*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node1",
						Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node2",
						Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node3",
						Labels: map[string]string{coreWeaveNodePoolUID: "uid-1"},
					},
				},
			},
			expectedTargetSize: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ng := makeTestNodeGroup("ng-1", "uid-1", 0, 5, initialTargetSize)

			err := ng.DeleteNodes(tc.nodesToDelete)
			if tc.expectedError != nil {
				require.Equal(t, tc.expectedError, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, ng.nodepool.GetTargetSize(), tc.expectedTargetSize)
		})
	}
}

func TestDecreaseTargetSize(t *testing.T) {
	testCases := map[string]struct {
		delta              int
		expectedTargetSize int
		expectedError      error
	}{
		"positive-delta": {
			delta:              2,
			expectedTargetSize: 1,
		},
		"negative-delta": {
			delta:              -2,
			expectedTargetSize: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ng := makeTestNodeGroup("ng-1", "uid-1", 1, 5, 3)

			err := ng.DecreaseTargetSize(tc.delta)
			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedTargetSize, ng.nodepool.GetTargetSize())
		})
	}
}

func TestBuildResourceList(t *testing.T) {
	cpuInstance := &InstanceType{
		VCPU:               192,
		MemoryKi:           1583282428,
		GPU:                0,
		EphemeralStorageKi: 29299982,
		Architecture:       "amd64",
		MaxPods:            110,
	}

	gpuInstance := &InstanceType{
		VCPU:               128,
		MemoryKi:           1055337468,
		GPU:                8,
		EphemeralStorageKi: 7499362648,
		Architecture:       "amd64",
		MaxPods:            110,
	}

	highPodNoStorageInstance := &InstanceType{
		VCPU:               64,
		MemoryKi:           1024,
		GPU:                0,
		EphemeralStorageKi: 0,
		Architecture:       "amd64",
		MaxPods:            300,
	}

	testCases := map[string]struct {
		nodePool          *CoreWeaveNodePool
		instanceType      *InstanceType
		expectedResources apiv1.ResourceList
	}{
		"standard cpu node": {
			nodePool:     makeTestNodePool("uid-1", "ng-1", 1, 5, 3),
			instanceType: cpuInstance,
			expectedResources: apiv1.ResourceList{
				apiv1.ResourceCPU:              *resource.NewQuantity(192, resource.DecimalSI),
				apiv1.ResourceMemory:           *resource.NewQuantity(1583282428*1024, resource.BinarySI),
				apiv1.ResourceEphemeralStorage: *resource.NewQuantity(29299982*1024, resource.BinarySI),
				apiv1.ResourcePods:             *resource.NewQuantity(110, resource.DecimalSI),
			},
		},
		"gpu node": {
			nodePool:     makeTestNodePool("uid-2", "ng-2", 1, 5, 3),
			instanceType: gpuInstance,
			expectedResources: apiv1.ResourceList{
				apiv1.ResourceCPU:              *resource.NewQuantity(128, resource.DecimalSI),
				apiv1.ResourceMemory:           *resource.NewQuantity(1055337468*1024, resource.BinarySI),
				apiv1.ResourceEphemeralStorage: *resource.NewQuantity(7499362648*1024, resource.BinarySI),
				apiv1.ResourcePods:             *resource.NewQuantity(110, resource.DecimalSI),
				gpu.ResourceNvidiaGPU:          *resource.NewQuantity(8, resource.DecimalSI),
			},
		},
		"custom max pods and zero storage": {
			nodePool:     makeTestNodePool("uid-3", "ng-3", 1, 5, 3),
			instanceType: highPodNoStorageInstance,
			expectedResources: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewQuantity(64, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(1024*1024, resource.BinarySI),
				apiv1.ResourcePods:   *resource.NewQuantity(300, resource.DecimalSI),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ng := NewCoreWeaveNodeGroup(tc.nodePool)
			resources := ng.buildResourceList(tc.instanceType)

			require.Equal(t, len(tc.expectedResources), len(resources), "resource list length mismatch")

			for resName, expectedQty := range tc.expectedResources {
				actualQty, exists := resources[resName]
				require.True(t, exists, "expected resource %s missing", resName)

				require.True(t, expectedQty.Equal(actualQty),
					"resource %s mismatch: expected %v, got %v", resName, expectedQty, actualQty)
			}
		})
	}
}

func TestBuildNodeLabels(t *testing.T) {
	nodeName := "test-node-0"
	instanceTypeName := "turin-gp-l"

	instanceType := &InstanceType{
		Architecture: "amd64",
	}

	instanceTypeArm := &InstanceType{
		Architecture: "arm64",
	}

	instanceTypeEmptyArch := &InstanceType{
		Architecture: "",
	}

	testCases := map[string]struct {
		nodePool       *CoreWeaveNodePool
		instanceType   *InstanceType
		expectedLabels map[string]string
	}{
		"standard labels": {
			nodePool:     makeTestNodePool("uid-1", "ng-1", 1, 5, 3),
			instanceType: instanceType,
			expectedLabels: map[string]string{
				apiv1.LabelInstanceTypeStable: instanceTypeName,
				apiv1.LabelArchStable:         "amd64",
				apiv1.LabelOSStable:           cloudprovider.DefaultOS,
				coreWeaveNodePoolName:         "ng-1",
				coreWeaveNodePoolUID:          "uid-1",
			},
		},
		"with custom node labels": {
			nodePool: makeTestNodePool("uid-1", "ng-1", 1, 5, 3, withNodeLabels(map[string]string{
				"custom-label": "custom-value",
				"environment":  "prod",
			})),
			instanceType: instanceType,
			expectedLabels: map[string]string{
				apiv1.LabelInstanceTypeStable: instanceTypeName,
				apiv1.LabelArchStable:         "amd64",
				apiv1.LabelOSStable:           cloudprovider.DefaultOS,
				coreWeaveNodePoolName:         "ng-1",
				coreWeaveNodePoolUID:          "uid-1",
				"custom-label":                "custom-value",
				"environment":                 "prod",
			},
		},
		"architecture override arm64 arch": {
			nodePool:     makeTestNodePool("uid-arm", "ng-arm", 1, 1, 1),
			instanceType: instanceTypeArm,
			expectedLabels: map[string]string{
				apiv1.LabelInstanceTypeStable: instanceTypeName,
				apiv1.LabelArchStable:         "arm64",
				apiv1.LabelOSStable:           cloudprovider.DefaultOS,
				coreWeaveNodePoolName:         "ng-arm",
				coreWeaveNodePoolUID:          "uid-arm",
			},
		},
		"architecture override empty arch": {
			nodePool:     makeTestNodePool("uid-arm", "ng-arm", 1, 1, 1),
			instanceType: instanceTypeEmptyArch,
			expectedLabels: map[string]string{
				apiv1.LabelInstanceTypeStable: instanceTypeName,
				apiv1.LabelArchStable:         cloudprovider.DefaultArch,
				apiv1.LabelOSStable:           cloudprovider.DefaultOS,
				coreWeaveNodePoolName:         "ng-arm",
				coreWeaveNodePoolUID:          "uid-arm",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ng := NewCoreWeaveNodeGroup(tc.nodePool)
			labels := ng.buildNodeLabels(nodeName, instanceTypeName, tc.instanceType)

			for k, v := range tc.expectedLabels {
				require.Equal(t, v, labels[k], "label %s mismatch", k)
			}
			require.Equal(t, len(tc.expectedLabels), len(labels), "unexpected number of labels")
		})
	}
}

func TestBuildNodeFromInstanceType(t *testing.T) {
	instanceTypeName := "turin-gp-l"
	instanceType := &InstanceType{
		VCPU:               192,
		MemoryKi:           1583282428,
		GPU:                0,
		EphemeralStorageKi: 29299982,
		Architecture:       "amd64",
		MaxPods:            110,
	}

	customTaints := []apiv1.Taint{
		{Key: "dedicated", Value: "gpu", Effect: apiv1.TaintEffectNoSchedule},
	}
	customLabels := map[string]string{
		"custom-key": "custom-value",
	}

	testCases := map[string]struct {
		nodePool     *CoreWeaveNodePool
		expectTaints bool
		expectLabels bool
	}{
		"basic node": {
			nodePool:     makeTestNodePool("uid-1", "ng-basic", 1, 5, 3),
			expectTaints: false,
			expectLabels: false,
		},
		"node with taints and labels": {
			nodePool: makeTestNodePool("uid-2", "ng-complex", 1, 5, 3,
				withNodeTaints(customTaints),
				withNodeLabels(customLabels),
			),
			expectTaints: true,
			expectLabels: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ng := NewCoreWeaveNodeGroup(tc.nodePool)
			node, err := ng.buildNodeFromInstanceType(instanceTypeName, instanceType)

			require.NoError(t, err)
			require.NotNil(t, node)

			require.NotEmpty(t, node.Status.Capacity)
			require.Equal(t, node.Status.Capacity, node.Status.Allocatable, "capacity should equal allocatable for template")

			cpuQty := node.Status.Capacity[apiv1.ResourceCPU]
			require.Equal(t, int64(192), cpuQty.Value())

			require.Equal(t, instanceTypeName, node.Labels[apiv1.LabelInstanceTypeStable])
			require.Equal(t, tc.nodePool.GetUID(), node.Labels[coreWeaveNodePoolUID])

			if tc.expectLabels {
				require.Equal(t, "custom-value", node.Labels["custom-key"], "custom label missing")
			}

			if tc.expectTaints {
				require.Len(t, node.Spec.Taints, 1)
				require.Equal(t, "dedicated", node.Spec.Taints[0].Key)
			} else {
				require.Empty(t, node.Spec.Taints)
			}

			require.NotEmpty(t, node.Status.Conditions)
			isReady := false
			for _, cond := range node.Status.Conditions {
				if cond.Type == apiv1.NodeReady && cond.Status == apiv1.ConditionTrue {
					isReady = true
					break
				}
			}
			require.True(t, isReady, "node must have Ready condition set to True")
		})
	}
}

func TestTemplateNodeInfo(t *testing.T) {
	validInstanceType := "turin-gp-l"

	nodeLabels := map[string]string{
		"environment": "production",
		"team":        "ai-research",
	}

	nodeTaints := []apiv1.Taint{
		{
			Key:    "dedicated",
			Value:  "gpu",
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}

	testCases := map[string]struct {
		nodePool      *CoreWeaveNodePool
		expectedError string
		validateNode  func(t *testing.T, node *apiv1.Node)
	}{
		"success basic case": {
			nodePool: makeTestNodePool("uid-1", "ng-basic", 1, 5, 3,
				withInstanceType(validInstanceType),
			),
			validateNode: func(t *testing.T, node *apiv1.Node) {
				cpu := node.Status.Capacity[apiv1.ResourceCPU]
				require.Equal(t, int64(192), cpu.Value())

				require.Equal(t, "uid-1", node.Labels[coreWeaveNodePoolUID])
			},
		},
		"with node labels": {
			nodePool: makeTestNodePool("uid-2", "ng-labels", 1, 5, 3,
				withInstanceType(validInstanceType),
				withNodeLabels(nodeLabels),
			),
			validateNode: func(t *testing.T, node *apiv1.Node) {
				require.Equal(t, "production", node.Labels["environment"])
				require.Equal(t, "ai-research", node.Labels["team"])
				require.Equal(t, validInstanceType, node.Labels[apiv1.LabelInstanceTypeStable])
			},
		},
		"with node taints": {
			nodePool: makeTestNodePool("uid-3", "ng-taints", 1, 5, 3,
				withInstanceType(validInstanceType),
				withNodeTaints(nodeTaints),
			),
			validateNode: func(t *testing.T, node *apiv1.Node) {
				require.Len(t, node.Spec.Taints, 1)
				require.Equal(t, "dedicated", node.Spec.Taints[0].Key)
				require.Equal(t, "gpu", node.Spec.Taints[0].Value)
				require.Equal(t, apiv1.TaintEffectNoSchedule, node.Spec.Taints[0].Effect)
			},
		},
		"with both labels and taints": {
			nodePool: makeTestNodePool("uid-4", "ng-both", 1, 5, 3,
				withInstanceType(validInstanceType),
				withNodeLabels(nodeLabels),
				withNodeTaints(nodeTaints),
			),
			validateNode: func(t *testing.T, node *apiv1.Node) {
				require.Equal(t, "production", node.Labels["environment"])
				require.Len(t, node.Spec.Taints, 1)
				require.Equal(t, "dedicated", node.Spec.Taints[0].Key)
			},
		},
		"architecture override check": {
			nodePool: makeTestNodePool("uid-5", "ng-arm", 1, 5, 3,
				withInstanceType("gd-1xgh200"),
			),
			validateNode: func(t *testing.T, node *apiv1.Node) {
				require.Equal(t, "arm64", node.Labels[apiv1.LabelArchStable])

				cpu := node.Status.Capacity[apiv1.ResourceCPU]
				require.Equal(t, int64(72), cpu.Value())
			},
		},
		"missing instance type error": {
			nodePool: makeTestNodePool("uid-err-1", "ng-err-1", 1, 5, 3,
				withInstanceType(""),
			),
			expectedError: "node pool ng-err-1 has no instance type defined",
		},
		"unknown instance type error": {
			nodePool: makeTestNodePool("uid-err-2", "ng-err-2", 1, 5, 3,
				withInstanceType("non-existent-type"),
			),
			expectedError: "unknown instance type: non-existent-type",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ng := NewCoreWeaveNodeGroup(tc.nodePool)
			nodeInfo, err := ng.TemplateNodeInfo()

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
				require.Nil(t, nodeInfo)
			} else {
				require.NoError(t, err)
				require.NotNil(t, nodeInfo)

				node := nodeInfo.Node()
				require.NotNil(t, node)

				if tc.validateNode != nil {
					tc.validateNode(t, node)
				}
			}
		})
	}
}
