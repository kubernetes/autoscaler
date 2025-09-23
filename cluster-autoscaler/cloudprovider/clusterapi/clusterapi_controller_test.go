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

package clusterapi

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	fakediscovery "k8s.io/client-go/discovery/fake"
	clientgotesting "k8s.io/client-go/testing"
)

func TestControllerFindMachine(t *testing.T) {
	type testCase struct {
		description    string
		name           string
		namespace      string
		useAnnotation  bool
		lookupSucceeds bool
	}

	var testCases = []testCase{{
		description:    "lookup fails",
		lookupSucceeds: false,
		useAnnotation:  false,
		name:           "machine-does-not-exist",
		namespace:      "namespace-does-not-exist",
	}, {
		description:    "lookup fails in valid namespace",
		lookupSucceeds: false,
		useAnnotation:  false,
		name:           "machine-does-not-exist-in-existing-namespace",
	}, {
		description:    "lookup succeeds",
		lookupSucceeds: true,
		useAnnotation:  false,
	}, {
		description:    "lookup succeeds with annotation",
		lookupSucceeds: true,
		useAnnotation:  true,
	}}

	test := func(t *testing.T, tc testCase, testConfig *TestConfig) {
		controller := NewTestMachineController(t)
		defer controller.Stop()
		controller.AddTestConfigs(testConfig)

		machine, err := controller.findMachine(path.Join(tc.namespace, tc.name))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tc.lookupSucceeds && machine == nil {
			t.Error("expected success, findMachine failed")
		}

		if tc.lookupSucceeds && machine != nil {
			if machine.GetName() != tc.name {
				t.Errorf("expected %q, got %q", tc.name, machine.GetName())
			}
			if machine.GetNamespace() != tc.namespace {
				t.Errorf("expected %q, got %q", tc.namespace, machine.GetNamespace())
			}
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			testConfig := NewTestConfigBuilder().
				ForMachineSet().
				WithNodeCount(1).
				WithAnnotations(map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				}).
				Build()
			if tc.name == "" {
				tc.name = testConfig.machines[0].GetName()
			}
			if tc.namespace == "" {
				tc.namespace = testConfig.machines[0].GetNamespace()
			}
			if tc.useAnnotation {
				for i := range testConfig.machines {
					n := testConfig.nodes[i]
					annotations := n.GetAnnotations()
					_, ok := annotations[machineAnnotationKey]
					if !ok {
						t.Fatal("node did not contain machineAnnotationKey")
					}
					delete(annotations, machineAnnotationKey)
					n.SetAnnotations(annotations)
				}
			}
			test(t, tc, testConfig)
		})
	}
}

func TestControllerFindMachineOwner(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	// Test #1: Lookup succeeds
	testResult1, err := controller.findMachineOwner(testConfig.machines[0].DeepCopy())
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}
	if testResult1 == nil {
		t.Fatal("expected non-nil result")
	}
	if testConfig.spec.machineSetName != testResult1.GetName() {
		t.Errorf("expected %q, got %q", testConfig.spec.machineSetName, testResult1.GetName())
	}

	// Test #2: Lookup fails as the machine ownerref Name != machineset Name
	testMachine2 := testConfig.machines[0].DeepCopy()
	ownerRefs := testMachine2.GetOwnerReferences()
	ownerRefs[0].Name = "does-not-match-machineset"

	testMachine2.SetOwnerReferences(ownerRefs)

	testResult2, err := controller.findMachineOwner(testMachine2)
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}
	if testResult2 != nil {
		t.Fatal("expected nil result")
	}

	// Test #3: Delete the MachineSet and lookup should fail
	if err := controller.DeleteResource(controller.machineSetInformer, controller.machineSetResource, testConfig.machineSet); err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}
	testResult3, err := controller.findMachineOwner(testConfig.machines[0].DeepCopy())
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}
	if testResult3 != nil {
		t.Fatal("expected lookup to fail")
	}
}

func TestControllerFindMachineByProviderID(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	// Remove all the "machine" annotation values on all the
	// nodes. We want to force findMachineByProviderID() to only
	// be successful by searching on provider ID.
	for _, node := range testConfig.nodes {
		delete(node.Annotations, machineAnnotationKey)
		if err := controller.nodeInformer.GetStore().Update(node); err != nil {
			t.Fatalf("unexpected error updating node, got %v", err)
		}
	}

	// Test #1: Verify underlying machine provider ID matches
	machine, err := controller.findMachineByProviderID(normalizedProviderString(testConfig.nodes[0].Spec.ProviderID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if machine == nil {
		t.Fatal("expected to find machine")
	}

	if !reflect.DeepEqual(machine, testConfig.machines[0]) {
		t.Fatalf("expected machines to be equal - expected %+v, got %+v", testConfig.machines[0], machine)
	}

	// Test #2: Verify machine returned by fake provider ID is correct machine
	fakeProviderID := fmt.Sprintf("%s$s/%s", testConfig.machines[0].GetNamespace(), testConfig.machines[0].GetName())
	machine, err = controller.findMachineByProviderID(normalizedProviderID(fakeProviderID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if machine != nil {
		t.Fatal("expected find to fail")
	}

	// Test #3: Verify machine is not found if it has a
	// non-existent or different provider ID.
	machine = testConfig.machines[0].DeepCopy()
	if err := unstructured.SetNestedField(machine.Object, "does-not-match", "spec", "providerID"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := controller.UpdateResource(controller.machineInformer, controller.machineResource, machine); err != nil {
		t.Fatalf("unexpected error updating machine, got %v", err)
	}

	machine, err = controller.findMachineByProviderID(normalizedProviderString(testConfig.nodes[0].Spec.ProviderID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if machine != nil {
		t.Fatal("expected find to fail")
	}
}

func TestControllerFindNodeByNodeName(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	// Test #1: Verify known node can be found
	node, err := controller.findNodeByNodeName(testConfig.nodes[0].Name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected lookup to be successful")
	}

	// Test #2: Verify non-existent node cannot be found
	node, err = controller.findNodeByNodeName(testConfig.nodes[0].Name + "non-existent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node != nil {
		t.Fatal("expected lookup to fail")
	}
}

func TestControllerListMachinesForScalableResource(t *testing.T) {
	test := func(t *testing.T, testConfig1 *TestConfig, testConfig2 *TestConfig) {
		controller := NewTestMachineController(t)
		defer controller.Stop()
		controller.AddTestConfigs(testConfig1)

		if err := controller.AddTestConfigs(testConfig2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scalableResource1 := testConfig1.machineSet
		if testConfig1.machineDeployment != nil {
			scalableResource1 = testConfig1.machineDeployment
		}

		scalableResource2 := testConfig2.machineSet
		if testConfig2.machineDeployment != nil {
			scalableResource2 = testConfig2.machineDeployment
		}

		machinesInScalableResource1, err := controller.listMachinesForScalableResource(scalableResource1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		machinesInScalableResource2, err := controller.listMachinesForScalableResource(scalableResource2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		actual := len(machinesInScalableResource1) + len(machinesInScalableResource2)
		expected := len(testConfig1.machines) + len(testConfig2.machines)
		if actual != expected {
			t.Fatalf("expected %d machines, got %d", expected, actual)
		}

		// Sort results as order is not guaranteed.
		sort.Slice(machinesInScalableResource1, func(i, j int) bool {
			return machinesInScalableResource1[i].GetName() < machinesInScalableResource1[j].GetName()
		})
		sort.Slice(machinesInScalableResource2, func(i, j int) bool {
			return machinesInScalableResource2[i].GetName() < machinesInScalableResource2[j].GetName()
		})

		for i, m := range machinesInScalableResource1 {
			if m.GetName() != testConfig1.machines[i].GetName() {
				t.Errorf("expected %q, got %q", testConfig1.machines[i].GetName(), m.GetName())
			}
			if m.GetNamespace() != testConfig1.machines[i].GetNamespace() {
				t.Errorf("expected %q, got %q", testConfig1.machines[i].GetNamespace(), m.GetNamespace())
			}
		}

		for i, m := range machinesInScalableResource2 {
			if m.GetName() != testConfig2.machines[i].GetName() {
				t.Errorf("expected %q, got %q", testConfig2.machines[i].GetName(), m.GetName())
			}
			if m.GetNamespace() != testConfig2.machines[i].GetNamespace() {
				t.Errorf("expected %q, got %q", testConfig2.machines[i].GetNamespace(), m.GetNamespace())
			}
		}

		// Finally everything in the respective objects should be equal.
		if !reflect.DeepEqual(testConfig1.machines, machinesInScalableResource1) {
			t.Fatalf("expected %+v, got %+v", testConfig1.machines, machinesInScalableResource2)
		}
		if !reflect.DeepEqual(testConfig2.machines, machinesInScalableResource2) {
			t.Fatalf("expected %+v, got %+v", testConfig2.machines, machinesInScalableResource2)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		namespace := RandomString(6)
		clusterName := RandomString(6)
		testConfig1 := NewTestConfigBuilder().
			ForMachineSet().
			WithNamespace(namespace).
			WithClusterName(clusterName).
			WithNodeCount(5).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).
			Build()

		// Construct a second set of objects and add the machines,
		// nodes and the additional machineset to the existing set of
		// test objects in the controller. This gives us two
		// machinesets, each with their own machines and linked nodes.
		testConfig2 := NewTestConfigBuilder().
			ForMachineSet().
			WithNamespace(namespace).
			WithClusterName(clusterName).
			WithNodeCount(5).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).
			Build()

		test(t, testConfig1, testConfig2)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		namespace := RandomString(6)
		clusterName := RandomString(6)
		testConfig1 := NewTestConfigBuilder().
			ForMachineDeployment().
			WithNamespace(namespace).
			WithClusterName(clusterName).
			WithNodeCount(5).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).
			Build()

		// Construct a second set of objects and add the machines,
		// nodes, machineset, and the additional machineset to the existing set of
		// test objects in the controller. This gives us two
		// machinedeployments, each with their own machineSet, machines and linked nodes.
		testConfig2 := NewTestConfigBuilder().
			ForMachineDeployment().
			WithNamespace(namespace).
			WithClusterName(clusterName).
			WithNodeCount(5).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).
			Build()

		test(t, testConfig1, testConfig2)
	})
}

func TestControllerLookupNodeGroupForNonExistentNode(t *testing.T) {
	test := func(t *testing.T, testConfig *TestConfig) {
		controller := NewTestMachineController(t)
		defer controller.Stop()
		controller.AddTestConfigs(testConfig)

		node := testConfig.nodes[0].DeepCopy()
		node.Spec.ProviderID = "does-not-exist"

		ng, err := controller.nodeGroupForNode(node)

		// Looking up a node that doesn't exist doesn't generate an
		// error. But, equally, the ng should actually be nil.
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ng != nil {
			t.Fatalf("unexpected nodegroup: %v", ng)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		testConfig := NewTestConfigBuilder().
			ForMachineSet().
			WithNodeCount(1).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).
			Build()
		test(t, testConfig)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := NewTestConfigBuilder().
			ForMachineDeployment().
			WithNodeCount(1).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).
			Build()
		test(t, testConfig)
	})
}

func TestControllerNodeGroupForNodeWithMissingMachineOwner(t *testing.T) {
	test := func(t *testing.T, testConfig *TestConfig) {
		controller := NewTestMachineController(t)
		defer controller.Stop()
		controller.AddTestConfigs(testConfig)

		machine := testConfig.machines[0].DeepCopy()
		machine.SetOwnerReferences([]metav1.OwnerReference{})

		if err := controller.UpdateResource(controller.machineInformer, controller.machineResource, machine); err != nil {
			t.Fatalf("unexpected error updating machine, got %v", err)
		}

		ng, err := controller.nodeGroupForNode(testConfig.nodes[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ng != nil {
			t.Fatalf("unexpected nodegroup: %v", ng)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		testConfig := NewTestConfigBuilder().
			ForMachineSet().
			WithNodeCount(1).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).
			Build()
		test(t, testConfig)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := NewTestConfigBuilder().
			ForMachineDeployment().
			WithNodeCount(1).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).
			Build()
		test(t, testConfig)
	})
}

func TestControllerNodeGroupForNodeWithMissingSetMachineOwner(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	machineSet := testConfig.machineSet.DeepCopy()
	machineSet.SetOwnerReferences([]metav1.OwnerReference{})

	if err := controller.UpdateResource(controller.machineSetInformer, controller.machineSetResource, machineSet); err != nil {
		t.Fatalf("unexpected error updating machine, got %v", err)
	}

	ng, err := controller.nodeGroupForNode(testConfig.nodes[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ng != nil {
		t.Fatalf("unexpected nodegroup: %v", ng)
	}
}

func TestControllerNodeGroupForNodeWithPositiveScalingBounds(t *testing.T) {
	test := func(t *testing.T, testConfig *TestConfig) {
		controller := NewTestMachineController(t)
		defer controller.Stop()
		controller.AddTestConfigs(testConfig)

		ng, err := controller.nodeGroupForNode(testConfig.nodes[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// We allow scaling if minSize=maxSize
		if ng == nil {
			t.Fatalf("unexpected nodegroup: %v", ng)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		testConfig := NewTestConfigBuilder().
			ForMachineSet().
			WithNodeCount(1).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "1",
			}).
			Build()
		test(t, testConfig)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := NewTestConfigBuilder().
			ForMachineDeployment().
			WithNodeCount(1).
			WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "1",
			}).
			Build()
		test(t, testConfig)
	})
}

func TestControllerNodeGroups(t *testing.T) {
	assertNodegroupLen := func(t *testing.T, controller *testMachineController, expected int) {
		t.Helper()
		nodegroups, err := controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := len(nodegroups); got != expected {
			t.Fatalf("expected %d, got %d", expected, got)
		}
	}

	annotations := map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "2",
	}

	controller := NewTestMachineController(t)
	defer controller.Stop()

	namespace := RandomString(6)
	clusterName := RandomString(6)

	// Test #1: zero nodegroups
	assertNodegroupLen(t, controller, 0)

	// Test #2: add 5 machineset-based nodegroups
	machineSetConfigs := NewTestConfigBuilder().
		ForMachineSet().
		WithNamespace(namespace).
		WithClusterName(clusterName).
		WithNodeCount(1).
		WithAnnotations(annotations).
		BuildMultiple(5)
	if err := controller.AddTestConfigs(machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 5)

	// Test #2: add 2 machinedeployment-based nodegroups
	machineDeploymentConfigs := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNamespace(namespace).
		WithClusterName(clusterName).
		WithNodeCount(1).
		WithAnnotations(annotations).
		BuildMultiple(2)
	if err := controller.AddTestConfigs(machineDeploymentConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 7)

	// Test #3: delete 5 machineset-backed objects
	if err := controller.DeleteTestConfigs(machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 2)

	// Test #4: delete 2 machinedeployment-backed objects
	if err := controller.DeleteTestConfigs(machineDeploymentConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 0)

	annotations = map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "1",
	}

	// Test #5: 5 machineset with minSize=maxSize results in a five nodegroup
	machineSetConfigs = NewTestConfigBuilder().
		ForMachineSet().
		WithNamespace(namespace).
		WithClusterName(clusterName).
		WithNodeCount(1).
		WithAnnotations(annotations).
		BuildMultiple(5)
	if err := controller.AddTestConfigs(machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 5)

	// Test #6: add 2 machinedeployment with minSize=maxSize
	machineDeploymentConfigs = NewTestConfigBuilder().
		ForMachineDeployment().
		WithNamespace(namespace).
		WithClusterName(clusterName).
		WithNodeCount(1).
		WithAnnotations(annotations).
		BuildMultiple(2)
	if err := controller.AddTestConfigs(machineDeploymentConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 7)

	annotations = map[string]string{
		nodeGroupMinSizeAnnotationKey: "-1",
		nodeGroupMaxSizeAnnotationKey: "1",
	}

	// Test #7: machineset with bad scaling bounds results in an error and no nodegroups
	machineSetConfigs = NewTestConfigBuilder().
		ForMachineSet().
		WithNamespace(namespace).
		WithClusterName(clusterName).
		WithNodeCount(5).
		WithAnnotations(annotations).
		BuildMultiple(1)
	if err := controller.AddTestConfigs(machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := controller.nodeGroups(); err == nil {
		t.Fatalf("expected an error")
	}

	// Test #8: machinedeployment with bad scaling bounds results in an error and no nodegroups
	machineDeploymentConfigs = NewTestConfigBuilder().
		ForMachineDeployment().
		WithNamespace(namespace).
		WithClusterName(clusterName).
		WithNodeCount(2).
		WithAnnotations(annotations).
		BuildMultiple(1)
	if err := controller.AddTestConfigs(machineDeploymentConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := controller.nodeGroups(); err == nil {
		t.Fatalf("expected an error")
	}
}

func TestControllerNodeGroupsNodeCount(t *testing.T) {
	type testCase struct {
		nodeGroups            int
		nodesPerGroup         int
		expectedNodeGroups    int
		expectedNodesPerGroup int
	}

	var testCases = []testCase{{
		nodeGroups:            0,
		nodesPerGroup:         0,
		expectedNodeGroups:    0,
		expectedNodesPerGroup: 0,
	}, {
		nodeGroups:            1,
		nodesPerGroup:         0,
		expectedNodeGroups:    0,
		expectedNodesPerGroup: 0,
	}, {
		nodeGroups:            2,
		nodesPerGroup:         10,
		expectedNodeGroups:    2,
		expectedNodesPerGroup: 10,
	}}

	test := func(t *testing.T, tc testCase, testConfigs []*TestConfig) {
		controller := NewTestMachineController(t)
		defer controller.Stop()
		controller.AddTestConfigs(testConfigs...)

		nodegroups, err := controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := len(nodegroups); got != tc.expectedNodeGroups {
			t.Fatalf("expected %d, got %d", tc.expectedNodeGroups, got)
		}

		for i := range nodegroups {
			nodes, err := nodegroups[i].Nodes()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := len(nodes); got != tc.expectedNodesPerGroup {
				t.Fatalf("expected %d, got %d", tc.expectedNodesPerGroup, got)
			}
		}
	}

	annotations := map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	}

	t.Run("MachineSet", func(t *testing.T) {
		for _, tc := range testCases {
			testConfigs := NewTestConfigBuilder().
				ForMachineSet().
				WithNodeCount(tc.nodesPerGroup).
				WithAnnotations(annotations).
				BuildMultiple(tc.nodeGroups)
			test(t, tc, testConfigs)
		}
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		for _, tc := range testCases {
			testConfigs := NewTestConfigBuilder().
				ForMachineDeployment().
				WithNodeCount(tc.nodesPerGroup).
				WithAnnotations(annotations).
				BuildMultiple(tc.nodeGroups)
			test(t, tc, testConfigs)
		}
	})
}

func TestControllerFindMachineFromNodeAnnotation(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	// Remove all the provider ID values on all the machines. We
	// want to force findMachineByProviderID() to fallback to
	// searching using the annotation on the node object.
	for _, machine := range testConfig.machines {
		unstructured.RemoveNestedField(machine.Object, "spec", "providerID")

		if err := controller.UpdateResource(controller.machineInformer, controller.machineResource, machine); err != nil {
			t.Fatalf("unexpected error updating machine, got %v", err)
		}
	}

	// Test #1: Verify machine can be found from node annotation
	machine, err := controller.findMachineByProviderID(normalizedProviderString(testConfig.nodes[0].Spec.ProviderID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if machine == nil {
		t.Fatal("expected to find machine")
	}
	if !reflect.DeepEqual(machine, testConfig.machines[0]) {
		t.Fatalf("expected machines to be equal - expected %+v, got %+v", testConfig.machines[0], machine)
	}

	// Test #2: Verify machine is not found if it has no
	// corresponding machine annotation.
	node := testConfig.nodes[0].DeepCopy()
	delete(node.Annotations, machineAnnotationKey)
	if err := controller.nodeInformer.GetStore().Update(node); err != nil {
		t.Fatalf("unexpected error updating node, got %v", err)
	}
	machine, err = controller.findMachineByProviderID(normalizedProviderString(testConfig.nodes[0].Spec.ProviderID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if machine != nil {
		t.Fatal("expected find to fail")
	}
}

func TestControllerMachineSetNodeNamesWithoutLinkage(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(3).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	// Remove all linkage between node and machine.
	for i := range testConfig.machines {
		machine := testConfig.machines[i].DeepCopy()

		unstructured.RemoveNestedField(machine.Object, "spec", "providerID")
		unstructured.RemoveNestedField(machine.Object, "status", "nodeRef")

		if err := controller.UpdateResource(controller.machineInformer, controller.machineResource, machine); err != nil {
			t.Fatalf("unexpected error updating machine, got %v", err)
		}
	}

	nodegroups, err := controller.nodeGroups()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l := len(nodegroups); l != 1 {
		t.Fatalf("expected 1 nodegroup, got %d", l)
	}

	ng := nodegroups[0]
	nodeNames, err := ng.Nodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nodeCount := 0
	for _, node := range nodeNames {
		if !isPendingMachineProviderID(normalizedProviderString(node.Id)) {
			nodeCount++
		}
	}

	// We removed all linkage - so we should get 0 nodes back.
	if nodeCount != 0 {
		t.Fatalf("expected len=0, got len=%v", nodeCount)
	}
}

func TestControllerMachineSetNodeNamesUsingProviderID(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(3).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	nodegroups, err := controller.nodeGroups()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l := len(nodegroups); l != 1 {
		t.Fatalf("expected 1 nodegroup, got %d", l)
	}

	ng := nodegroups[0]
	nodeNames, err := ng.Nodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodeNames) != len(testConfig.nodes) {
		t.Fatalf("expected len=%v, got len=%v", len(testConfig.nodes), len(nodeNames))
	}

	sort.Slice(nodeNames, func(i, j int) bool {
		return nodeNames[i].Id < nodeNames[j].Id
	})

	for i := range testConfig.nodes {
		if nodeNames[i].Id != testConfig.nodes[i].Spec.ProviderID {
			t.Fatalf("expected %q, got %q", testConfig.nodes[i].Spec.ProviderID, nodeNames[i].Id)
		}
	}
}

func TestControllerMachineSetNodeNamesUsingStatusNodeRefName(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(3).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	// Remove all the provider ID values on all the machines. We
	// want to force machineSetNodeNames() to fallback to
	// searching using Status.NodeRef.Name.
	for i := range testConfig.machines {
		machine := testConfig.machines[i].DeepCopy()

		unstructured.RemoveNestedField(machine.Object, "spec", "providerID")

		if err := controller.UpdateResource(controller.machineInformer, controller.machineResource, machine); err != nil {
			t.Fatalf("unexpected error updating machine, got %v", err)
		}
	}

	nodegroups, err := controller.nodeGroups()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l := len(nodegroups); l != 1 {
		t.Fatalf("expected 1 nodegroup, got %d", l)
	}

	nodeNames, err := nodegroups[0].Nodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodeNames) != len(testConfig.nodes) {
		t.Fatalf("expected len=%v, got len=%v", len(testConfig.nodes), len(nodeNames))
	}

	sort.Slice(nodeNames, func(i, j int) bool {
		return nodeNames[i].Id < nodeNames[j].Id
	})

	for i := range testConfig.nodes {
		if nodeNames[i].Id != testConfig.nodes[i].Spec.ProviderID {
			t.Fatalf("expected %q, got %q", testConfig.nodes[i].Spec.ProviderID, nodeNames[i].Id)
		}
	}
}

func TestControllerGetAPIVersionGroup(t *testing.T) {
	t.Run("CAPI_GROUP is set", func(t *testing.T) {
		expected := "mygroup"
		t.Setenv(CAPIGroupEnvVar, expected)
		observed := getCAPIGroup()
		if observed != expected {
			t.Fatalf("Wrong Version Group detected, expected %q, got %q", expected, observed)
		}
	})

	t.Run("CAPI_GROUP is empty", func(t *testing.T) {
		expected := defaultCAPIGroup
		t.Setenv(CAPIGroupEnvVar, "")
		observed := getCAPIGroup()
		if observed != expected {
			t.Fatalf("Wrong Version Group detected, expected %q, got %q", expected, observed)
		}
	})
}

func TestControllerGetAPIVersion(t *testing.T) {
	t.Run("CAPI_VERSION is not set", func(t *testing.T) {
		expected := ""
		observed := getCAPIVersion()
		if observed != expected {
			t.Fatalf("Wrong API Version detected, expected %q, got %q", expected, observed)
		}
	})

	t.Run("CAPI_VERSION is set", func(t *testing.T) {
		expected := "v1beta1"
		t.Setenv(CAPIVersionEnvVar, expected)
		observed := getCAPIVersion()
		if observed != expected {
			t.Fatalf("Wrong API Version detected, expected %q, got %q", expected, observed)
		}
	})
}

func TestControllerGetAPIVersionGroupWithMachineDeployments(t *testing.T) {
	testConfig := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "1",
		}).
		Build()
	t.Setenv(CAPIGroupEnvVar, customCAPIGroup)

	testConfig.machineDeployment.SetAPIVersion(fmt.Sprintf("%s/v1beta1", customCAPIGroup))
	testConfig.machineSet.SetAPIVersion(fmt.Sprintf("%s/v1beta1", customCAPIGroup))

	for _, machine := range testConfig.machines {
		machine.SetAPIVersion(fmt.Sprintf("%s/v1beta1", customCAPIGroup))
	}

	controller := NewTestMachineController(t)
	defer controller.Stop()
	controller.AddTestConfigs(testConfig)

	machineDeployments, err := controller.managementClient.Resource(controller.machineDeploymentResource).Namespace(testConfig.spec.namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l := len(machineDeployments.Items); l != 1 {
		t.Fatalf("Incorrect number of MachineDeployments, expected 1, got %d", l)
	}

	machineSets, err := controller.managementClient.Resource(controller.machineSetResource).Namespace(testConfig.spec.namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l := len(machineSets.Items); l != 1 {
		t.Fatalf("Incorrect number of MachineDeployments, expected 1, got %d", l)
	}

	machines, err := controller.managementClient.Resource(controller.machineResource).Namespace(testConfig.spec.namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if l := len(machines.Items); l != 1 {
		t.Fatalf("Incorrect number of Machines, expected 1, got %d", l)
	}
}

func TestGetAPIGroupPreferredVersion(t *testing.T) {
	customVersion := "v1"
	testCases := []struct {
		description      string
		APIGroup         string
		preferredVersion string
		envVar           string
		error            bool
	}{
		{
			description:      "find version for default API group",
			APIGroup:         defaultCAPIGroup,
			preferredVersion: "v1beta2",
			envVar:           "",
			error:            false,
		},
		{
			description:      "find version for another API group",
			APIGroup:         customCAPIGroup,
			preferredVersion: "v1beta1",
			envVar:           "",
			error:            false,
		},
		{
			description:      "find version for another API group while overriding version with env var",
			APIGroup:         customCAPIGroup,
			preferredVersion: customVersion,
			envVar:           customVersion,
			error:            false,
		},
		{
			description:      "API group does not exist",
			APIGroup:         "does.not.exist",
			preferredVersion: "",
			envVar:           "",
			error:            true,
		},
	}

	discoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &clientgotesting.Fake{
			Resources: []*metav1.APIResourceList{
				{
					GroupVersion: fmt.Sprintf("%s/v1beta1", customCAPIGroup),
				},
				{
					GroupVersion: fmt.Sprintf("%s/v1beta2", defaultCAPIGroup),
				},
				{
					GroupVersion: fmt.Sprintf("%s/%s", customCAPIGroup, customVersion),
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			t.Setenv(CAPIVersionEnvVar, tc.envVar)
			version, err := getCAPIGroupPreferredVersion(discoveryClient, tc.APIGroup)
			if (err != nil) != tc.error {
				t.Errorf("expected to have error: %t. Had an error: %t", tc.error, err != nil)
			}
			if version != tc.preferredVersion {
				t.Errorf("expected %v, got: %v", tc.preferredVersion, version)
			}
		})
	}
}

func TestGroupVersionHasResource(t *testing.T) {
	testCases := []struct {
		description  string
		APIGroup     string
		resourceName string
		expected     bool
		error        bool
	}{
		{
			description:  "true when it finds resource",
			resourceName: resourceNameMachineDeployment,
			APIGroup:     fmt.Sprintf("%s/v1beta2", defaultCAPIGroup),
			expected:     true,
			error:        false,
		},
		{
			description:  "false when it does not find resource",
			resourceName: "resourceDoesNotExist",
			APIGroup:     fmt.Sprintf("%s/v1beta2", defaultCAPIGroup),
			expected:     false,
			error:        false,
		},
		{
			description:  "error when invalid groupVersion",
			resourceName: resourceNameMachineDeployment,
			APIGroup:     "APIGroupDoesNotExist",
			expected:     false,
			error:        true,
		},
	}

	discoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &clientgotesting.Fake{
			Resources: []*metav1.APIResourceList{
				{
					GroupVersion: fmt.Sprintf("%s/v1beta2", defaultCAPIGroup),
					APIResources: []metav1.APIResource{
						{
							Name: resourceNameMachineDeployment,
						},
						{
							Name: resourceNameMachineSet,
						},
						{
							Name: resourceNameMachine,
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got, err := groupVersionHasResource(discoveryClient, tc.APIGroup, tc.resourceName)
			if (err != nil) != tc.error {
				t.Errorf("expected to have error: %t. Had an error: %t", tc.error, err != nil)
			}
			if got != tc.expected {
				t.Errorf("expected %v, got: %v", tc.expected, got)
			}
		})
	}
}

func TestIsFailedMachineProviderID(t *testing.T) {
	testCases := []struct {
		name       string
		providerID normalizedProviderID
		expected   bool
	}{
		{
			name:       "with the failed machine prefix",
			providerID: normalizedProviderID(fmt.Sprintf("%sfoo", failedMachinePrefix)),
			expected:   true,
		},
		{
			name:       "without the failed machine prefix",
			providerID: normalizedProviderID("foo"),
			expected:   false,
		},
		{
			name:       "with provider ID created by createFailedMachineNormalizedProviderID should return true",
			providerID: normalizedProviderID(createFailedMachineNormalizedProviderID("cluster-api", "id-0001")),
			expected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isFailedMachineProviderID(tc.providerID); got != tc.expected {
				t.Errorf("test case: %s, expected: %v, got: %v", tc.name, tc.expected, got)
			}
		})
	}
}

func TestMachineKeyFromFailedProviderID(t *testing.T) {
	testCases := []struct {
		name       string
		providerID normalizedProviderID
		expected   string
	}{
		{
			name:       "with a valid failed machine prefix",
			providerID: normalizedProviderID(fmt.Sprintf("%stest-namespace_foo", failedMachinePrefix)),
			expected:   "test-namespace/foo",
		},
		{
			name:       "with a machine with an underscore in the name",
			providerID: normalizedProviderID(fmt.Sprintf("%stest-namespace_foo_bar", failedMachinePrefix)),
			expected:   "test-namespace/foo_bar",
		},
		{
			name:       "with a provider ID created by createFailedMachineNormalizedProviderID",
			providerID: normalizedProviderID(createFailedMachineNormalizedProviderID("cluster-api", "id-0001")),
			expected:   "cluster-api/id-0001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := machineKeyFromFailedProviderID(tc.providerID); got != tc.expected {
				t.Errorf("test case: %s, expected: %q, got: %q", tc.name, tc.expected, got)
			}
		})
	}
}

func Test_machineController_allowedByAutoDiscoverySpecs(t *testing.T) {
	for _, tc := range []struct {
		name               string
		testSpec           TestSpec
		autoDiscoverySpecs []*clusterAPIAutoDiscoveryConfig
		additionalLabels   map[string]string
		shouldMatch        bool
	}{{
		name:     "autodiscovery specs includes permissive spec that should match any MachineSet",
		testSpec: createTestSpec(RandomString(6), RandomString(6), RandomString(6), 1, false, nil, nil),
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{labelSelector: labels.NewSelector()},
			{clusterName: "foo", namespace: "bar", labelSelector: labels.Nothing()},
		},
		shouldMatch: true,
	}, {
		name:     "autodiscovery specs includes permissive spec that should match any MachineDeployment",
		testSpec: createTestSpec(RandomString(6), RandomString(6), RandomString(6), 1, true, nil, nil),
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{labelSelector: labels.NewSelector()},
			{clusterName: "foo", namespace: "bar", labelSelector: labels.Nothing()},
		},
		shouldMatch: true,
	}, {
		name:             "autodiscovery specs includes a restrictive spec that should match specific MachineSet",
		testSpec:         createTestSpec("default", "foo", RandomString(6), 1, false, nil, nil),
		additionalLabels: map[string]string{"color": "green"},
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{clusterName: "foo", namespace: "default", labelSelector: labels.SelectorFromSet(labels.Set{"color": "green"})},
			{clusterName: "wombat", namespace: "bar", labelSelector: labels.Nothing()},
		},
		shouldMatch: true,
	}, {
		name:             "autodiscovery specs includes a restrictive spec that should match specific MachineDeployment",
		testSpec:         createTestSpec("default", "foo", RandomString(6), 1, true, nil, nil),
		additionalLabels: map[string]string{"color": "green"},
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{clusterName: "foo", namespace: "default", labelSelector: labels.SelectorFromSet(labels.Set{"color": "green"})},
			{clusterName: "wombat", namespace: "bar", labelSelector: labels.Nothing()},
		},
		shouldMatch: true,
	}, {
		name:             "autodiscovery specs does not include any specs that should match specific MachineSet",
		testSpec:         createTestSpec("default", "foo", RandomString(6), 1, false, nil, nil),
		additionalLabels: map[string]string{"color": "green"},
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{clusterName: "test", namespace: "default", labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue"})},
			{clusterName: "wombat", namespace: "bar", labelSelector: labels.Nothing()},
		},
		shouldMatch: false,
	}, {
		name:             "autodiscovery specs does not include any specs that should match specific MachineDeployment",
		testSpec:         createTestSpec("default", "foo", RandomString(6), 1, true, nil, nil),
		additionalLabels: map[string]string{"color": "green"},
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{clusterName: "test", namespace: "default", labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue"})},
			{clusterName: "wombat", namespace: "bar", labelSelector: labels.Nothing()},
		},
		shouldMatch: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			testConfigs := createTestConfigs(tc.testSpec)
			resource := testConfigs[0].machineSet
			if tc.testSpec.rootIsMachineDeployment {
				resource = testConfigs[0].machineDeployment
			}
			if tc.additionalLabels != nil {
				resource.SetLabels(labels.Merge(resource.GetLabels(), tc.additionalLabels))
			}
			c := &machineController{
				autoDiscoverySpecs: tc.autoDiscoverySpecs,
			}

			got := c.allowedByAutoDiscoverySpecs(resource)
			if got != tc.shouldMatch {
				t.Errorf("allowedByAutoDiscoverySpecs got = %v, want %v", got, tc.shouldMatch)
			}
		})
	}
}

func Test_machineController_listScalableResources(t *testing.T) {
	uniqueMDConfig := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNodeCount(1).
		Build()

	mdTestConfigs := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNodeCount(1).
		BuildMultiple(5)
	mdTestConfigs = append(mdTestConfigs, uniqueMDConfig)

	allMachineDeployments := make([]*unstructured.Unstructured, 0, len(mdTestConfigs))
	for i := range mdTestConfigs {
		allMachineDeployments = append(allMachineDeployments, mdTestConfigs[i].machineDeployment)
	}

	uniqueMSConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		Build()

	msTestConfigs := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		BuildMultiple(5)
	msTestConfigs = append(msTestConfigs, uniqueMSConfig)

	allMachineSets := make([]*unstructured.Unstructured, 0, len(msTestConfigs))
	for i := range msTestConfigs {
		allMachineSets = append(allMachineSets, msTestConfigs[i].machineSet)
	}

	allTestConfigs := append(mdTestConfigs, msTestConfigs...)
	allScalableResources := append(allMachineDeployments, allMachineSets...)

	for _, tc := range []struct {
		name               string
		autoDiscoverySpecs []*clusterAPIAutoDiscoveryConfig
		want               []*unstructured.Unstructured
		wantErr            bool
	}{{
		name:               "undefined autodiscovery results in returning all scalable resources",
		autoDiscoverySpecs: nil,
		want:               allScalableResources,
		wantErr:            false,
	}, {
		name: "autodiscovery configuration to match against unique MachineSet only returns that MachineSet",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
		},
		want:    []*unstructured.Unstructured{uniqueMSConfig.machineSet},
		wantErr: false,
	}, {
		name: "autodiscovery configuration to match against unique MachineDeployment only returns that MachineDeployment",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
		},
		want:    []*unstructured.Unstructured{uniqueMDConfig.machineDeployment},
		wantErr: false,
	}, {
		name: "autodiscovery configuration to match against both unique MachineDeployment and unique MachineSet only returns those scalable resources",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
		},
		want:    []*unstructured.Unstructured{uniqueMDConfig.machineDeployment, uniqueMSConfig.machineSet},
		wantErr: false,
	}, {
		name: "autodiscovery configuration to match against both unique MachineDeployment, unique MachineSet, and a permissive config returns all scalable resources",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
			{labelSelector: labels.NewSelector()},
		},
		want:    allScalableResources,
		wantErr: false,
	}, {
		name: "autodiscovery configuration to match against both unique MachineDeployment, unique MachineSet, and a restrictive returns unique scalable resources",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
			{namespace: RandomString(6), clusterName: RandomString(6), labelSelector: labels.Nothing()},
		},
		want:    []*unstructured.Unstructured{uniqueMDConfig.machineDeployment, uniqueMSConfig.machineSet},
		wantErr: false,
	}, {
		name: "autodiscovery configuration to match against a restrictive config returns no scalable resources",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: RandomString(6), clusterName: RandomString(6), labelSelector: labels.Nothing()},
		},
		want:    nil,
		wantErr: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewTestMachineController(t)
			defer c.Stop()
			c.AddTestConfigs(allTestConfigs...)
			c.autoDiscoverySpecs = tc.autoDiscoverySpecs

			got, err := c.listScalableResources()
			if (err != nil) != tc.wantErr {
				t.Errorf("listScalableRsources() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if len(got) != len(tc.want) {
				t.Errorf("listScalableRsources() expected length of got to be = %v, got %v", len(tc.want), len(got))
			}

			// Sort results as order is not guaranteed.
			sort.Slice(got, func(i, j int) bool {
				return got[i].GetName() < got[j].GetName()
			})
			sort.Slice(tc.want, func(i, j int) bool {
				return tc.want[i].GetName() < tc.want[j].GetName()
			})

			if err == nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("listScalableRsources() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_machineController_nodeGroupForNode(t *testing.T) {
	uniqueMDConfig := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	mdTestConfigs := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		BuildMultiple(5)
	mdTestConfigs = append(mdTestConfigs, uniqueMDConfig)

	uniqueMSConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	msTestConfigs := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		BuildMultiple(5)
	msTestConfigs = append(msTestConfigs, uniqueMSConfig)

	allTestConfigs := append(mdTestConfigs, msTestConfigs...)

	for _, tc := range []struct {
		name               string
		autoDiscoverySpecs []*clusterAPIAutoDiscoveryConfig
		node               *corev1.Node
		scalableResource   *unstructured.Unstructured
		wantErr            bool
	}{{
		name:               "undefined autodiscovery results in returning MachineSet resource for given node",
		autoDiscoverySpecs: nil,
		node:               msTestConfigs[0].nodes[0],
		scalableResource:   msTestConfigs[0].machineSet,
		wantErr:            false,
	}, {
		name:               "undefined autodiscovery results in returning MachineDeployment resource for given node",
		autoDiscoverySpecs: nil,
		node:               mdTestConfigs[0].nodes[0],
		scalableResource:   mdTestConfigs[0].machineDeployment,
		wantErr:            false,
	}, {
		name: "autodiscovery configuration to match against a restrictive config does not return a nodegroup",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: RandomString(6), clusterName: RandomString(6), labelSelector: labels.Nothing()},
		},
		node:             msTestConfigs[0].nodes[0],
		scalableResource: nil,
		wantErr:          false,
	}, {
		name: "autodiscovery configuration to match against unique MachineSet returns nodegroup for that MachineSet",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
		},
		node:             uniqueMSConfig.nodes[0],
		scalableResource: uniqueMSConfig.machineSet,
		wantErr:          false,
	}, {
		name: "autodiscovery configuration to match against unique MachineDeployment returns nodegroup for that MachineDeployment",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
		},
		node:             uniqueMDConfig.nodes[0],
		scalableResource: uniqueMDConfig.machineDeployment,
		wantErr:          false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewTestMachineController(t)
			defer c.Stop()
			c.AddTestConfigs(allTestConfigs...)
			c.autoDiscoverySpecs = tc.autoDiscoverySpecs

			got, err := c.nodeGroupForNode(tc.node)
			if (err != nil) != tc.wantErr {
				t.Errorf("nodeGroupForNode() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if got == nil && tc.scalableResource != nil {
				t.Error("expected a node group to be returned, got nil")
				return
			}

			if tc.scalableResource == nil && got != nil {
				t.Errorf("expected nil node group, got: %v", got)
				return
			}

			if tc.scalableResource != nil && !reflect.DeepEqual(got.scalableResource.unstructured, tc.scalableResource) {
				t.Errorf("nodeGroupForNode() got = %v, want node group for scalable resource %v", got, tc.scalableResource)
			}
		})
	}
}

func Test_machineController_nodeGroups(t *testing.T) {
	uniqueMDConfig := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	mdTestConfigs := NewTestConfigBuilder().
		ForMachineDeployment().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		BuildMultiple(5)
	mdTestConfigs = append(mdTestConfigs, uniqueMDConfig)

	allMachineDeployments := make([]*unstructured.Unstructured, 0, len(mdTestConfigs))
	for i := range mdTestConfigs {
		allMachineDeployments = append(allMachineDeployments, mdTestConfigs[i].machineDeployment)
	}

	uniqueMSConfig := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		Build()

	msTestConfigs := NewTestConfigBuilder().
		ForMachineSet().
		WithNodeCount(1).
		WithAnnotations(map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}).
		BuildMultiple(5)
	msTestConfigs = append(msTestConfigs, uniqueMSConfig)

	allMachineSets := make([]*unstructured.Unstructured, 0, len(msTestConfigs))
	for i := range msTestConfigs {
		allMachineSets = append(allMachineSets, msTestConfigs[i].machineSet)
	}

	allTestConfigs := append(mdTestConfigs, msTestConfigs...)
	allScalableResources := append(allMachineDeployments, allMachineSets...)

	for _, tc := range []struct {
		name                      string
		autoDiscoverySpecs        []*clusterAPIAutoDiscoveryConfig
		expectedScalableResources []*unstructured.Unstructured
		wantErr                   bool
	}{{
		name:                      "undefined autodiscovery results in returning nodegroups for all scalable resources",
		autoDiscoverySpecs:        nil,
		expectedScalableResources: allScalableResources,
		wantErr:                   false,
	}, {
		name: "autodiscovery configuration to match against unique MachineSet only returns nodegroup for that MachineSet",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
		},
		expectedScalableResources: []*unstructured.Unstructured{uniqueMSConfig.machineSet},
		wantErr:                   false,
	}, {
		name: "autodiscovery configuration to match against unique MachineDeployment only returns nodegroup for that MachineDeployment",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
		},
		expectedScalableResources: []*unstructured.Unstructured{uniqueMDConfig.machineDeployment},
		wantErr:                   false,
	}, {
		name: "autodiscovery configuration to match against both unique MachineDeployment and unique MachineSet only returns nodegroups for those scalable resources",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
		},
		expectedScalableResources: []*unstructured.Unstructured{uniqueMDConfig.machineDeployment, uniqueMSConfig.machineSet},
		wantErr:                   false,
	}, {
		name: "autodiscovery configuration to match against both unique MachineDeployment, unique MachineSet, and a permissive config returns nodegroups for all scalable resources",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
			{labelSelector: labels.NewSelector()},
		},
		expectedScalableResources: allScalableResources,
		wantErr:                   false,
	}, {
		name: "autodiscovery configuration to match against both unique MachineDeployment, unique MachineSet, and a restrictive returns nodegroups for unique scalable resources",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: uniqueMDConfig.namespace, clusterName: uniqueMDConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMDConfig.machineDeployment.GetLabels())},
			{namespace: uniqueMSConfig.namespace, clusterName: uniqueMSConfig.clusterName, labelSelector: labels.SelectorFromSet(uniqueMSConfig.machineSet.GetLabels())},
			{namespace: RandomString(6), clusterName: RandomString(6), labelSelector: labels.Nothing()},
		},
		expectedScalableResources: []*unstructured.Unstructured{uniqueMDConfig.machineDeployment, uniqueMSConfig.machineSet},
		wantErr:                   false,
	}, {
		name: "autodiscovery configuration to match against a restrictive config returns no nodegroups",
		autoDiscoverySpecs: []*clusterAPIAutoDiscoveryConfig{
			{namespace: RandomString(6), clusterName: RandomString(6), labelSelector: labels.Nothing()},
		},
		expectedScalableResources: nil,
		wantErr:                   false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewTestMachineController(t)
			defer c.Stop()
			c.AddTestConfigs(allTestConfigs...)
			c.autoDiscoverySpecs = tc.autoDiscoverySpecs

			got, err := c.nodeGroups()
			if (err != nil) != tc.wantErr {
				t.Errorf("nodeGroups() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if len(got) != len(tc.expectedScalableResources) {
				t.Errorf("nodeGroups() expected length of got to be = %v, got %v", len(tc.expectedScalableResources), len(got))
			}

			// Sort results as order is not guaranteed.
			sort.Slice(got, func(i, j int) bool {
				return got[i].(*nodegroup).scalableResource.Name() < got[j].(*nodegroup).scalableResource.Name()
			})
			sort.Slice(tc.expectedScalableResources, func(i, j int) bool {
				return tc.expectedScalableResources[i].GetName() < tc.expectedScalableResources[j].GetName()
			})

			if err == nil {
				for i := range got {
					if !reflect.DeepEqual(got[i].(*nodegroup).scalableResource.unstructured, tc.expectedScalableResources[i]) {
						t.Errorf("nodeGroups() got = %v, expected to consist of nodegroups for scalable resources: %v", got, tc.expectedScalableResources)
					}
				}
			}
		})
	}
}

func Test_isDeletingMachineProviderID(t *testing.T) {
	type testCase struct {
		description    string
		providerID     string
		expectedReturn bool
	}

	testCases := []testCase{
		{
			description:    "proper provider ID without deletion prefix should return false",
			providerID:     "fake-provider://a.provider.id-0001",
			expectedReturn: false,
		},
		{
			description:    "provider ID with deletion prefix should return true",
			providerID:     fmt.Sprintf("%s%s_%s", deletingMachinePrefix, "cluster-api", "id-0001"),
			expectedReturn: true,
		},
		{
			description:    "empty provider ID should return false",
			providerID:     "",
			expectedReturn: false,
		},
		{
			description:    "provider ID created with createDeletingMachineNormalizedProviderID returns true",
			providerID:     createDeletingMachineNormalizedProviderID("cluster-api", "id-0001"),
			expectedReturn: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			observed := isDeletingMachineProviderID(normalizedProviderID(tc.providerID))
			if observed != tc.expectedReturn {
				t.Fatalf("unexpected return for provider ID %q, expected %t, observed %t", tc.providerID, tc.expectedReturn, observed)
			}
		})
	}

}

// TestNodeHasValidProviderID tests all permutations of provider IDs
// to determine whether the providerID is the standard cloud provider ID
// or has been modified by CAS CAPI provider
func TestNodeHasValidProviderID(t *testing.T) {
	type testCase struct {
		description    string
		providerID     normalizedProviderID
		expectedReturn bool
	}

	testCases := []testCase{
		{
			description:    "real looking provider ID should return true",
			providerID:     normalizedProviderID("fake-provider://a.provider.id-0001"),
			expectedReturn: true,
		},
		{
			description:    "provider ID created with createDeletingMachineNormalizedProviderID should return false",
			providerID:     normalizedProviderID(createDeletingMachineNormalizedProviderID("cluster-api", "id-0001")),
			expectedReturn: false,
		},
		{
			description:    "provider ID created with createPendingDeletionMachineNormalizedProviderID should return false",
			providerID:     normalizedProviderID(createPendingMachineProviderID("cluster-api", "id-0001")),
			expectedReturn: false,
		},
		{
			description:    "provider ID created with createFailedMachineNormalizedProviderID should return false",
			providerID:     normalizedProviderID(createFailedMachineNormalizedProviderID("cluster-api", "id-0001")),
			expectedReturn: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			observed := isProviderIDNormalized(tc.providerID)
			if observed != tc.expectedReturn {
				t.Fatalf("unexpected return for provider ID %q, expected %t, observed %t", tc.providerID, tc.expectedReturn, observed)
			}
		})
	}
}

func Test_machineKeyFromDeletingMachineProviderID(t *testing.T) {
	type testCase struct {
		description    string
		providerID     string
		expectedReturn string
	}

	testCases := []testCase{
		{
			description:    "real looking provider ID with no deletion prefix returns provider id",
			providerID:     "fake-provider://a.provider.id-0001",
			expectedReturn: "fake-provider://a.provider.id-0001",
		},
		{
			description:    "namespace_name provider ID with no deletion prefix returns proper provider ID",
			providerID:     "cluster-api_id-0001",
			expectedReturn: "cluster-api/id-0001",
		},
		{
			description:    "namespace_name provider ID with deletion prefix returns proper provider ID",
			providerID:     fmt.Sprintf("%s%s_%s", deletingMachinePrefix, "cluster-api", "id-0001"),
			expectedReturn: "cluster-api/id-0001",
		},
		{
			description:    "namespace_name provider ID with deletion prefix and two underscores returns proper provider ID",
			providerID:     fmt.Sprintf("%s%s_%s", deletingMachinePrefix, "cluster-api", "id_0001"),
			expectedReturn: "cluster-api/id_0001",
		},
		{
			description:    "provider ID created with createDeletingMachineNormalizedProviderID returns proper provider ID",
			providerID:     createDeletingMachineNormalizedProviderID("cluster-api", "id-0001"),
			expectedReturn: "cluster-api/id-0001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			observed := machineKeyFromDeletingMachineProviderID(normalizedProviderID(tc.providerID))
			if observed != tc.expectedReturn {
				t.Fatalf("unexpected return for provider ID %q, expected %q, observed %q", tc.providerID, tc.expectedReturn, observed)
			}
		})
	}
}

func Test_createDeletingMachineNormalizedProviderID(t *testing.T) {
	type testCase struct {
		description    string
		namespace      string
		name           string
		expectedReturn string
	}

	testCases := []testCase{
		{
			description:    "namespace and name return proper normalized ID",
			namespace:      "cluster-api",
			name:           "id-0001",
			expectedReturn: fmt.Sprintf("%s%s_%s", deletingMachinePrefix, "cluster-api", "id-0001"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			observed := createDeletingMachineNormalizedProviderID(tc.namespace, tc.name)
			if observed != tc.expectedReturn {
				t.Fatalf("unexpected return for (namespace %q, name %q), expected %q, observed %q", tc.namespace, tc.name, tc.expectedReturn, observed)
			}
		})
	}
}

// Test_createPendingMachineProviderID tests the creation of a pending machine provider ID
func Test_createPendingMachineProviderID(t *testing.T) {
	type testCase struct {
		description    string
		namespace      string
		name           string
		expectedReturn string
	}

	testCases := []testCase{
		{
			description:    "namespace and name return proper normalized ID",
			namespace:      "cluster-api",
			name:           "id-0001",
			expectedReturn: fmt.Sprintf("%s%s_%s", pendingMachinePrefix, "cluster-api", "id-0001"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			observed := createPendingMachineProviderID(tc.namespace, tc.name)
			if observed != tc.expectedReturn {
				t.Fatalf("unexpected return for (namespace %q, name %q), expected %q, observed %q", tc.namespace, tc.name, tc.expectedReturn, observed)
			}
		})
	}
}
