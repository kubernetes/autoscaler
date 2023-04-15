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
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	provisioningv1 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/provisioning.cattle.io/v1"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/utils/pointer"
)

const testProviderID = "rke2://"

func TestNodeGroups(t *testing.T) {
	tests := []struct {
		name                string
		machinePools        []provisioningv1.RKEMachinePool
		expectedGroups      int
		expectedErrContains string
		expectedResources   corev1.ResourceList
		clusterNameOverride string
	}{
		{
			name: "normal",
			machinePools: []provisioningv1.RKEMachinePool{
				{
					Name:     nodeGroupDev,
					Quantity: pointer.Int32(1),
					MachineDeploymentAnnotations: map[string]string{
						minSizeAnnotation: "0",
						maxSizeAnnotation: "3",
					},
				},
			},
			expectedGroups: 1,
		},
		{
			name: "without size annotations",
			machinePools: []provisioningv1.RKEMachinePool{
				{
					Name:     nodeGroupDev,
					Quantity: pointer.Int32(1),
				},
			},
			expectedGroups: 0,
		},
		{
			name:                "missing quantity",
			expectedGroups:      0,
			expectedErrContains: "machine pool quantity is not set",
		},
		{
			name:                "missing cluster",
			expectedGroups:      0,
			expectedErrContains: "clusters.provisioning.cattle.io \"some-other-cluster\" not found",
			clusterNameOverride: "some-other-cluster",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pools, err := machinePoolsToUnstructured(tc.machinePools)
			if err != nil {
				t.Fatal(err)
			}

			config := &cloudConfig{
				ClusterName:      "test-cluster",
				ClusterNamespace: "default",
			}

			cluster := newCluster(config.ClusterName, config.ClusterNamespace, pools)

			if tc.clusterNameOverride != "" {
				config.ClusterName = tc.clusterNameOverride
			}

			provider := RancherCloudProvider{
				resourceLimiter: &cloudprovider.ResourceLimiter{},
				client: fakedynamic.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						clusterGVR(): "kindList",
					},
					cluster,
				),
				config: config,
			}

			if err := provider.Refresh(); err != nil {
				if tc.expectedErrContains == "" || !strings.Contains(err.Error(), tc.expectedErrContains) {
					t.Fatalf("expected err to contain %q, got %q", tc.expectedErrContains, err)
				}
			}

			if len(provider.NodeGroups()) != tc.expectedGroups {
				t.Fatalf("expected %q groups, got %q", tc.expectedGroups, len(provider.NodeGroups()))
			}
		})
	}
}

func TestNodeGroupForNode(t *testing.T) {
	provider, err := setup([]runtime.Object{newMachine(nodeGroupDev, 0), newMachine(nodeGroupProd, 0)})
	if err != nil {
		t.Fatal(err)
	}

	if err := provider.Refresh(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                 string
		node                 *corev1.Node
		nodeGroupId          string
		expectedNodeGroupNil bool
	}{
		{
			name: "match dev",
			node: &corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Name:        nodeName(nodeGroupDev, 0),
					Annotations: map[string]string{machineNodeAnnotationKey: nodeName(nodeGroupDev, 0)},
				},
				Spec: corev1.NodeSpec{
					ProviderID: testProviderID + nodeName(nodeGroupDev, 0),
				},
			},
			nodeGroupId: nodeGroupDev,
		},
		{
			name: "match prod",
			node: &corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Name:        nodeName(nodeGroupProd, 0),
					Annotations: map[string]string{machineNodeAnnotationKey: nodeName(nodeGroupProd, 0)},
				},
				Spec: corev1.NodeSpec{
					ProviderID: testProviderID + nodeName(nodeGroupProd, 0),
				},
			},
			nodeGroupId: nodeGroupProd,
		},
		{
			name: "not rke2 node",
			node: &corev1.Node{
				ObjectMeta: v1.ObjectMeta{Name: nodeName(nodeGroupDev, 0)},
				Spec: corev1.NodeSpec{
					ProviderID: "whatever://" + nodeName(nodeGroupDev, 0),
				},
			},
			expectedNodeGroupNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ng, err := provider.NodeGroupForNode(tc.node)
			if err != nil {
				t.Fatal(err)
			}

			if !tc.expectedNodeGroupNil {
				if ng == nil {
					t.Fatalf("expected node group from node %s", tc.node.Name)
				}

				if tc.nodeGroupId != ng.Id() {
					t.Fatalf("expected node group id %s, got %s", tc.nodeGroupId, ng.Id())
				}
			} else {
				if ng != nil {
					t.Fatalf("expected node group to be nil, got %v", ng)
				}
			}
		})
	}

}

func newCluster(name, namespace string, machinePools interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Cluster",
			"apiVersion": rancherProvisioningGroup + "/" + rancherProvisioningVersion,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"rkeConfig": map[string]interface{}{
					"machinePools": machinePools,
				},
			},
			"status": map[string]interface{}{},
		},
	}
}
