/*
Copyright 2020 The Kubernetes Authors.

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

package rancher

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	provisioningv1 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/provisioning.cattle.io/v1"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/utils/pointer"
)

const (
	testCluster   = "test-cluster"
	testNamespace = "default"
	nodeGroupDev  = "dev"
	nodeGroupProd = "ng-long-prod"
)

func TestNodeGroupNodes(t *testing.T) {
	tests := []struct {
		name                string
		nodeGroup           nodeGroup
		expectedNodes       int
		expectedErrContains string
		machines            func() []runtime.Object
	}{
		{
			name:          "normal",
			nodeGroup:     nodeGroup{name: nodeGroupDev},
			expectedNodes: 2,
			machines: func() []runtime.Object {
				return []runtime.Object{
					newMachine(nodeGroupDev, 0),
					newMachine(nodeGroupDev, 1),
				}
			},
		},
		{
			name:          "mixed machines",
			nodeGroup:     nodeGroup{name: nodeGroupDev},
			expectedNodes: 3,
			machines: func() []runtime.Object {
				return []runtime.Object{
					newMachine(nodeGroupDev, 0),
					newMachine(nodeGroupDev, 1),
					newMachine(nodeGroupDev, 2),
					newMachine(nodeGroupProd, 0),
					newMachine(nodeGroupProd, 1),
				}
			},
		},
		{
			name:          "no matching machines",
			nodeGroup:     nodeGroup{name: nodeGroupDev},
			expectedNodes: 0,
			machines: func() []runtime.Object {
				return []runtime.Object{
					newMachine(nodeGroupProd, 0),
					newMachine(nodeGroupProd, 1),
				}
			},
		},
		{
			name:                "machine without provider id",
			nodeGroup:           nodeGroup{name: nodeGroupDev},
			expectedNodes:       0,
			expectedErrContains: "could not find providerID in machine",
			machines: func() []runtime.Object {
				machine := newMachine(nodeGroupDev, 0)
				_ = unstructured.SetNestedMap(machine.Object, map[string]interface{}{}, "spec")
				return []runtime.Object{machine}
			},
		},
		{
			name:          "machine without provider id during provisioning",
			nodeGroup:     nodeGroup{name: nodeGroupDev},
			expectedNodes: 1,
			machines: func() []runtime.Object {
				machineProvisioning := newMachine(nodeGroupDev, 0)
				_ = unstructured.SetNestedMap(machineProvisioning.Object, map[string]interface{}{}, "spec")
				_ = unstructured.SetNestedField(machineProvisioning.Object, machinePhaseProvisioning, "status", "phase")
				return []runtime.Object{
					machineProvisioning,
					newMachine(nodeGroupDev, 1),
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := setup(tc.machines())
			if err != nil {
				t.Fatal(err)
			}

			tc.nodeGroup.provider = provider

			nodes, err := tc.nodeGroup.Nodes()
			if err != nil {
				if tc.expectedErrContains == "" || !strings.Contains(err.Error(), tc.expectedErrContains) {
					t.Fatalf("expected err to contain %q, got %q", tc.expectedErrContains, err)
				}
			}
			if len(nodes) != tc.expectedNodes {
				t.Fatalf("expected %v nodes, got %v", tc.expectedNodes, len(nodes))
			}
		})
	}
}

func TestNodeGroupDeleteNodes(t *testing.T) {
	tests := []struct {
		name                string
		nodeGroup           nodeGroup
		expectedTargetSize  int
		expectedErrContains string
		machines            []runtime.Object
		toDelete            []*corev1.Node
	}{
		{
			name: "delete node",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 1,
				minSize:  0,
				maxSize:  2,
			},
			expectedTargetSize: 0,
			machines:           []runtime.Object{newMachine(nodeGroupDev, 0)},
			toDelete: []*corev1.Node{
				newNode(nodeName(nodeGroupDev, 0)),
			},
		},
		{
			name: "delete multiple nodes",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 3,
				minSize:  0,
				maxSize:  3,
			},
			expectedTargetSize: 1,
			machines:           []runtime.Object{newMachine(nodeGroupDev, 0), newMachine(nodeGroupDev, 1), newMachine(nodeGroupDev, 2)},
			toDelete: []*corev1.Node{
				newNode(nodeName(nodeGroupDev, 0)),
				newNode(nodeName(nodeGroupDev, 2)),
			},
		},
		{
			name: "delete unknown node",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 1,
				minSize:  0,
				maxSize:  2,
			},
			expectedTargetSize:  1,
			expectedErrContains: fmt.Sprintf("node with providerID rke2://%s not found in node group %s", nodeName(nodeGroupDev, 42), nodeGroupDev),
			machines:            []runtime.Object{newMachine(nodeGroupDev, 0)},
			toDelete: []*corev1.Node{
				newNode(nodeName(nodeGroupDev, 42)),
			},
		},
		{
			name: "delete more nodes than min size",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 2,
				minSize:  1,
				maxSize:  2,
			},
			expectedTargetSize:  2,
			expectedErrContains: "node group size would be below minimum size - desired: 0, min: 1",
			machines:            []runtime.Object{newMachine(nodeGroupDev, 0), newMachine(nodeGroupDev, 1)},
			toDelete: []*corev1.Node{
				{ObjectMeta: v1.ObjectMeta{Name: nodeName(nodeGroupDev, 0)}},
				{ObjectMeta: v1.ObjectMeta{Name: nodeName(nodeGroupDev, 1)}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := setup(tc.machines)
			if err != nil {
				t.Fatal(err)
			}

			tc.nodeGroup.provider = provider

			if err := provider.Refresh(); err != nil {
				t.Fatal(err)
			}

			// store delta before deleting nodes
			delta := tc.nodeGroup.replicas - tc.expectedTargetSize

			if err := tc.nodeGroup.DeleteNodes(tc.toDelete); err != nil {
				if tc.expectedErrContains == "" || !strings.Contains(err.Error(), tc.expectedErrContains) {
					t.Fatalf("expected err to contain %q, got %q", tc.expectedErrContains, err)
				}
			}

			targetSize, err := tc.nodeGroup.TargetSize()
			if err != nil {
				t.Fatal(err)
			}

			if tc.expectedTargetSize != targetSize {
				t.Fatalf("expected target size %v, got %v", tc.expectedTargetSize, targetSize)
			}

			// ensure we get fresh machines
			tc.nodeGroup.machines = nil
			machines, err := tc.nodeGroup.listMachines()
			if err != nil {
				t.Fatal(err)
			}

			annotationCount := 0
			for _, machine := range machines {
				if _, ok := machine.GetAnnotations()[machineDeleteAnnotationKey]; ok {
					annotationCount++
				}
			}
			if annotationCount != delta {
				t.Fatalf("expected %v machines to have the deleted annotation, got %v", delta, annotationCount)
			}
		})
	}
}

func TestIncreaseTargetSize(t *testing.T) {
	tests := []struct {
		name                string
		delta               int
		nodeGroup           nodeGroup
		expectedErrContains string
	}{
		{
			name: "valid increase",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 0,
				minSize:  0,
				maxSize:  2,
			},
			delta: 2,
		},
		{
			name: "too large",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 1,
				minSize:  0,
				maxSize:  2,
			},
			delta:               2,
			expectedErrContains: "size increase too large, desired: 3 max: 2",
		},
		{
			name: "negative",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 2,
				minSize:  0,
				maxSize:  2,
			},
			delta:               -2,
			expectedErrContains: "size increase must be positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := setup(nil)
			if err != nil {
				t.Fatal(err)
			}

			tc.nodeGroup.provider = provider
			if err := tc.nodeGroup.IncreaseSize(tc.delta); err != nil {
				if tc.expectedErrContains == "" || !strings.Contains(err.Error(), tc.expectedErrContains) {
					t.Fatalf("expected err to contain %q, got %q", tc.expectedErrContains, err)
				}
			}
		})
	}
}

func TestDecreaseTargetSize(t *testing.T) {
	tests := []struct {
		name                string
		delta               int
		nodeGroup           nodeGroup
		expectedErrContains string
	}{
		{
			name: "valid decrease",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 2,
				minSize:  0,
				maxSize:  2,
			},
			delta: -2,
		},
		{
			name: "too large",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 1,
				minSize:  0,
				maxSize:  2,
			},
			delta:               -2,
			expectedErrContains: "attempt to delete existing nodes targetSize: 1 delta: -2 existingNodes: 0",
		},
		{
			name: "positive",
			nodeGroup: nodeGroup{
				name:     nodeGroupDev,
				replicas: 2,
				minSize:  0,
				maxSize:  2,
			},
			delta:               2,
			expectedErrContains: "size decrease must be negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := setup(nil)
			if err != nil {
				t.Fatal(err)
			}

			tc.nodeGroup.provider = provider
			if err := tc.nodeGroup.DecreaseTargetSize(tc.delta); err != nil {
				if tc.expectedErrContains == "" || !strings.Contains(err.Error(), tc.expectedErrContains) {
					t.Fatalf("expected err to contain %q, got %q", tc.expectedErrContains, err)
				}
			}
		})
	}
}

func TestTemplateNodeInfo(t *testing.T) {
	provider, err := setup(nil)
	if err != nil {
		t.Fatal(err)
	}

	ng := nodeGroup{
		name:     nodeGroupDev,
		replicas: 2,
		minSize:  0,
		maxSize:  2,
		provider: provider,
		resources: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("2"),
			corev1.ResourceMemory:           resource.MustParse("2Gi"),
			corev1.ResourceEphemeralStorage: resource.MustParse("100Gi"),
		},
	}

	nodeInfo, err := ng.TemplateNodeInfo()
	if err != nil {
		t.Fatal(err)
	}

	if nodeInfo.GetAllocatable().GetMilliCPU() != ng.resources.Cpu().MilliValue() {
		t.Fatalf("expected nodeInfo to have %v MilliCPU, got %v",
			ng.resources.Cpu().MilliValue(), nodeInfo.GetAllocatable().GetMilliCPU())
	}

	if nodeInfo.GetAllocatable().GetMemory() != ng.resources.Memory().Value() {
		t.Fatalf("expected nodeInfo to have %v Memory, got %v",
			ng.resources.Memory().Value(), nodeInfo.GetAllocatable().GetMemory())
	}

	if nodeInfo.GetAllocatable().GetEphemeralStorage() != ng.resources.StorageEphemeral().Value() {
		t.Fatalf("expected nodeInfo to have %v ephemeral storage, got %v",
			ng.resources.StorageEphemeral().Value(), nodeInfo.GetAllocatable().GetEphemeralStorage())
	}
}

func TestNewNodeGroupFromMachinePool(t *testing.T) {
	provider, err := setup(nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                string
		machinePool         provisioningv1.RKEMachinePool
		expectedErrContains string
		expectedResources   corev1.ResourceList
	}{
		{
			name: "valid",
			machinePool: provisioningv1.RKEMachinePool{
				Name:     nodeGroupDev,
				Quantity: pointer.Int32(1),
				MachineDeploymentAnnotations: map[string]string{
					minSizeAnnotation:                  "0",
					maxSizeAnnotation:                  "3",
					resourceCPUAnnotation:              "2",
					resourceMemoryAnnotation:           "4Gi",
					resourceEphemeralStorageAnnotation: "50Gi",
				},
			},
			expectedResources: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("2"),
				corev1.ResourceMemory:           resource.MustParse("4Gi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("50Gi"),
			},
		},
		{
			name:                "missing size annotations",
			expectedErrContains: "missing min size annotation",
			machinePool: provisioningv1.RKEMachinePool{
				Name:     nodeGroupDev,
				Quantity: pointer.Int32(1),
				MachineDeploymentAnnotations: map[string]string{
					resourceCPUAnnotation:              "2",
					resourceMemoryAnnotation:           "4Gi",
					resourceEphemeralStorageAnnotation: "50Gi",
				},
			},
		},
		{
			name: "missing resource annotations",
			machinePool: provisioningv1.RKEMachinePool{
				Name:     nodeGroupDev,
				Quantity: pointer.Int32(1),
				MachineDeploymentAnnotations: map[string]string{
					minSizeAnnotation: "0",
					maxSizeAnnotation: "3",
				},
			},
			expectedResources: corev1.ResourceList{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ng, err := newNodeGroupFromMachinePool(provider, tc.machinePool)
			if err != nil {
				if tc.expectedErrContains == "" || !strings.Contains(err.Error(), tc.expectedErrContains) {
					t.Fatalf("expected err to contain %q, got %q", tc.expectedErrContains, err)
				}
				return
			}

			if ng.replicas != int(*tc.machinePool.Quantity) {
				t.Fatalf("expected nodegroup replicas %v, got %v", ng.replicas, tc.machinePool.Quantity)
			}

			if !reflect.DeepEqual(tc.expectedResources, ng.resources) {
				t.Fatalf("expected resources %v do not match node group resources %v", tc.expectedResources, ng.resources)
			}
		})
	}
}

func setup(machines []runtime.Object) (*RancherCloudProvider, error) {
	config := &cloudConfig{
		ClusterName:       testCluster,
		ClusterNamespace:  testNamespace,
		ClusterAPIVersion: "v1alpha4",
	}

	machinePools := []provisioningv1.RKEMachinePool{
		{
			Name:     nodeGroupDev,
			Quantity: pointer.Int32(1),
			MachineDeploymentAnnotations: map[string]string{
				minSizeAnnotation:                  "0",
				maxSizeAnnotation:                  "3",
				resourceCPUAnnotation:              "2",
				resourceMemoryAnnotation:           "4Gi",
				resourceEphemeralStorageAnnotation: "50Gi",
			},
		},
		{
			Name:     nodeGroupProd,
			Quantity: pointer.Int32(3),
			MachineDeploymentAnnotations: map[string]string{
				minSizeAnnotation:                  "0",
				maxSizeAnnotation:                  "3",
				resourceCPUAnnotation:              "2",
				resourceMemoryAnnotation:           "4Gi",
				resourceEphemeralStorageAnnotation: "50Gi",
			},
		},
	}

	pools, err := machinePoolsToUnstructured(machinePools)
	if err != nil {
		return nil, err
	}

	return &RancherCloudProvider{
		resourceLimiter: &cloudprovider.ResourceLimiter{},
		client: fakedynamic.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				machineGVR(config.ClusterAPIVersion): "kindList",
			},
			append(machines, newCluster(testCluster, testNamespace, pools))...,
		),
		config: config,
	}, nil
}

func newMachine(nodeGroupName string, num int) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Machine",
			"apiVersion": "cluster.x-k8s.io/v1alpha4",
			"metadata": map[string]interface{}{
				"name":      nodeName(nodeGroupName, num),
				"namespace": testNamespace,
				"labels": map[string]interface{}{
					machineDeploymentNameLabelKey:  fmt.Sprintf("%s-%s", testCluster, nodeGroupName),
					rancherMachinePoolNameLabelKey: nodeGroupName,
				},
			},
			"spec": map[string]interface{}{
				"clusterName": testCluster,
				"providerID":  testProviderID + nodeName(nodeGroupName, num),
			},
			"status": map[string]interface{}{
				"phase": "Running",
			},
		},
	}
}

func nodeName(nodeGroupName string, num int) string {
	return fmt.Sprintf("%s-%s-123456-%v", testCluster, nodeGroupName, num)
}

func newNode(name string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: v1.ObjectMeta{Name: name},
		Spec: corev1.NodeSpec{
			ProviderID: testProviderID + name,
		},
	}
}
