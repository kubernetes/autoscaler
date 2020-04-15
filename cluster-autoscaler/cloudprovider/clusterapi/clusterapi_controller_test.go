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
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	fakekube "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
	"k8s.io/utils/pointer"
)

type testControllerShutdownFunc func()

type testConfig struct {
	spec              *testSpec
	machineDeployment *MachineDeployment
	machineSet        *MachineSet
	machines          []*Machine
	nodes             []*corev1.Node
}

type testSpec struct {
	annotations             map[string]string
	machineDeploymentName   string
	machineSetName          string
	namespace               string
	nodeCount               int
	rootIsMachineDeployment bool
}

const customCAPIGroup = "custom.x-k8s.io"

func mustCreateTestController(t *testing.T, testConfigs ...*testConfig) (*machineController, testControllerShutdownFunc) {
	t.Helper()

	nodeObjects := make([]runtime.Object, 0)
	machineObjects := make([]runtime.Object, 0)

	for _, config := range testConfigs {
		for i := range config.nodes {
			nodeObjects = append(nodeObjects, config.nodes[i])
		}

		for i := range config.machines {
			machineObjects = append(machineObjects, newUnstructuredFromMachine(config.machines[i]))
		}

		machineObjects = append(machineObjects, newUnstructuredFromMachineSet(config.machineSet))
		if config.machineDeployment != nil {
			machineObjects = append(machineObjects, newUnstructuredFromMachineDeployment(config.machineDeployment))
		}
	}

	kubeclientSet := fakekube.NewSimpleClientset(nodeObjects...)
	dynamicClientset := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme(), machineObjects...)
	discoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &clientgotesting.Fake{
			Resources: []*v1.APIResourceList{
				{
					GroupVersion: fmt.Sprintf("%s/v1beta1", customCAPIGroup),
					APIResources: []v1.APIResource{
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
				{
					GroupVersion: fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
					APIResources: []v1.APIResource{
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
	controller, err := newMachineController(dynamicClientset, kubeclientSet, discoveryClient)
	if err != nil {
		t.Fatal("failed to create test controller")
	}

	stopCh := make(chan struct{})
	if err := controller.run(stopCh); err != nil {
		t.Fatalf("failed to run controller: %v", err)
	}

	return controller, func() {
		close(stopCh)
	}
}

func createMachineSetTestConfig(namespace string, nodeCount int, annotations map[string]string) *testConfig {
	return createTestConfigs(createTestSpecs(namespace, 1, nodeCount, false, annotations)...)[0]
}

func createMachineSetTestConfigs(namespace string, configCount, nodeCount int, annotations map[string]string) []*testConfig {
	return createTestConfigs(createTestSpecs(namespace, configCount, nodeCount, false, annotations)...)
}

func createMachineDeploymentTestConfig(namespace string, nodeCount int, annotations map[string]string) *testConfig {
	return createTestConfigs(createTestSpecs(namespace, 1, nodeCount, true, annotations)...)[0]
}

func createMachineDeploymentTestConfigs(namespace string, configCount, nodeCount int, annotations map[string]string) []*testConfig {
	return createTestConfigs(createTestSpecs(namespace, configCount, nodeCount, true, annotations)...)
}

func createTestSpecs(namespace string, scalableResourceCount, nodeCount int, isMachineDeployment bool, annotations map[string]string) []testSpec {
	var specs []testSpec

	for i := 0; i < scalableResourceCount; i++ {
		specs = append(specs, testSpec{
			annotations:             annotations,
			machineDeploymentName:   fmt.Sprintf("machinedeployment-%d", i),
			machineSetName:          fmt.Sprintf("machineset-%d", i),
			namespace:               strings.ToLower(namespace),
			nodeCount:               nodeCount,
			rootIsMachineDeployment: isMachineDeployment,
		})
	}

	return specs
}

func createTestConfigs(specs ...testSpec) []*testConfig {
	var result []*testConfig

	for i, spec := range specs {
		config := &testConfig{
			spec:     &specs[i],
			nodes:    make([]*corev1.Node, spec.nodeCount),
			machines: make([]*Machine, spec.nodeCount),
		}

		config.machineSet = &MachineSet{
			TypeMeta: v1.TypeMeta{
				APIVersion: fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
				Kind:       "MachineSet",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      spec.machineSetName,
				Namespace: spec.namespace,
				UID:       types.UID(spec.machineSetName),
			},
		}

		if !spec.rootIsMachineDeployment {
			config.machineSet.ObjectMeta.Annotations = spec.annotations
			config.machineSet.Spec.Replicas = int32ptr(int32(spec.nodeCount))
		} else {
			config.machineDeployment = &MachineDeployment{
				TypeMeta: v1.TypeMeta{
					APIVersion: fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
					Kind:       "MachineDeployment",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:        spec.machineDeploymentName,
					Namespace:   spec.namespace,
					UID:         types.UID(spec.machineDeploymentName),
					Annotations: spec.annotations,
				},
				Spec: MachineDeploymentSpec{
					Replicas: int32ptr(int32(spec.nodeCount)),
				},
			}

			config.machineSet.OwnerReferences = make([]v1.OwnerReference, 1)
			config.machineSet.OwnerReferences[0] = v1.OwnerReference{
				Name: config.machineDeployment.Name,
				Kind: config.machineDeployment.Kind,
				UID:  config.machineDeployment.UID,
			}
		}

		machineOwner := v1.OwnerReference{
			Name: config.machineSet.Name,
			Kind: config.machineSet.Kind,
			UID:  config.machineSet.UID,
		}

		for j := 0; j < spec.nodeCount; j++ {
			config.nodes[j], config.machines[j] = makeLinkedNodeAndMachine(j, spec.namespace, machineOwner)
		}

		result = append(result, config)
	}

	return result
}

// makeLinkedNodeAndMachine creates a node and machine. The machine
// has its NodeRef set to the new node and the new machine's owner
// reference is set to owner.
func makeLinkedNodeAndMachine(i int, namespace string, owner v1.OwnerReference) (*corev1.Node, *Machine) {
	node := &corev1.Node{
		TypeMeta: v1.TypeMeta{
			Kind: "Node",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-node-%d", namespace, owner.Name, i),
			Annotations: map[string]string{
				machineAnnotationKey: fmt.Sprintf("%s/%s-%s-machine-%d", namespace, namespace, owner.Name, i),
			},
		},
		Spec: corev1.NodeSpec{
			ProviderID: fmt.Sprintf("test:////%s-%s-nodeid-%d", namespace, owner.Name, i),
		},
	}

	machine := &Machine{
		TypeMeta: v1.TypeMeta{
			APIVersion: fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
			Kind:       "Machine",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-machine-%d", namespace, owner.Name, i),
			Namespace: namespace,
			OwnerReferences: []v1.OwnerReference{{
				Name: owner.Name,
				Kind: owner.Kind,
				UID:  owner.UID,
			}},
		},
		Spec: MachineSpec{
			ProviderID: pointer.StringPtr(fmt.Sprintf("test:////%s-%s-nodeid-%d", namespace, owner.Name, i)),
		},
		Status: MachineStatus{
			NodeRef: &corev1.ObjectReference{
				Kind: node.Kind,
				Name: node.Name,
			},
		},
	}

	return node, machine
}

func int32ptr(v int32) *int32 {
	return &v
}

func addTestConfigs(t *testing.T, controller *machineController, testConfigs ...*testConfig) error {
	t.Helper()

	for _, config := range testConfigs {
		if config.machineDeployment != nil {

			if err := controller.machineDeploymentInformer.Informer().GetStore().Add(newUnstructuredFromMachineDeployment(config.machineDeployment)); err != nil {
				return err
			}
		}
		if err := controller.machineSetInformer.Informer().GetStore().Add(newUnstructuredFromMachineSet(config.machineSet)); err != nil {
			return err
		}
		for i := range config.machines {
			if err := controller.machineInformer.Informer().GetStore().Add(newUnstructuredFromMachine(config.machines[i])); err != nil {
				return err
			}
		}
		for i := range config.nodes {
			if err := controller.nodeInformer.GetStore().Add(config.nodes[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteTestConfigs(t *testing.T, controller *machineController, testConfigs ...*testConfig) error {
	t.Helper()

	for _, config := range testConfigs {
		for i := range config.nodes {
			if err := controller.nodeInformer.GetStore().Delete(config.nodes[i]); err != nil {
				return err
			}
		}
		for i := range config.machines {
			if err := controller.machineInformer.Informer().GetStore().Delete(config.machines[i]); err != nil {
				return err
			}
		}
		if err := controller.machineSetInformer.Informer().GetStore().Delete(config.machineSet); err != nil {
			return err
		}
		if config.machineDeployment != nil {
			if err := controller.machineDeploymentInformer.Informer().GetStore().Delete(config.machineDeployment); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestControllerFindMachineByID(t *testing.T) {
	type testCase struct {
		description    string
		name           string
		namespace      string
		lookupSucceeds bool
	}

	var testCases = []testCase{{
		description:    "lookup fails",
		lookupSucceeds: false,
		name:           "machine-does-not-exist",
		namespace:      "namespace-does-not-exist",
	}, {
		description:    "lookup fails in valid namespace",
		lookupSucceeds: false,
		name:           "machine-does-not-exist-in-existing-namespace",
	}, {
		description:    "lookup succeeds",
		lookupSucceeds: true,
	}}

	test := func(t *testing.T, tc testCase, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		machine, err := controller.findMachine(path.Join(tc.namespace, tc.name))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tc.lookupSucceeds && machine == nil {
			t.Error("expected success, findMachine failed")
		}

		if tc.lookupSucceeds && machine != nil {
			if machine.Name != tc.name {
				t.Errorf("expected %q, got %q", tc.name, machine.Name)
			}
			if machine.Namespace != tc.namespace {
				t.Errorf("expected %q, got %q", tc.namespace, machine.Namespace)
			}
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			testConfig := createMachineSetTestConfig(testNamespace, 1, map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			})
			if tc.name == "" {
				tc.name = testConfig.machines[0].Name
			}
			if tc.namespace == "" {
				tc.namespace = testConfig.machines[0].Namespace
			}
			test(t, tc, testConfig)
		})
	}
}

func TestControllerFindMachineOwner(t *testing.T) {
	testConfig := createMachineSetTestConfig(testNamespace, 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Test #1: Lookup succeeds
	testResult1, err := controller.findMachineOwner(testConfig.machines[0].DeepCopy())
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}
	if testResult1 == nil {
		t.Fatal("expected non-nil result")
	}
	if testConfig.spec.machineSetName != testResult1.Name {
		t.Errorf("expected %q, got %q", testConfig.spec.machineSetName, testResult1.Name)
	}

	// Test #2: Lookup fails as the machine UUID != machineset UUID
	testMachine2 := testConfig.machines[0].DeepCopy()
	testMachine2.OwnerReferences[0].UID = "does-not-match-machineset"
	testResult2, err := controller.findMachineOwner(testMachine2)
	if err != nil {
		t.Fatalf("unexpected error, got %v", err)
	}
	if testResult2 != nil {
		t.Fatal("expected nil result")
	}

	// Test #3: Delete the MachineSet and lookup should fail
	if err := controller.machineSetInformer.Informer().GetStore().Delete(testResult1); err != nil {
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
	testConfig := createMachineSetTestConfig(testNamespace, 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

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
	fakeProviderID := fmt.Sprintf("%s$s/%s", testConfig.machines[0].Namespace, testConfig.machines[0].Name)
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
	machine.Spec.ProviderID = pointer.StringPtr("does-not-match")
	if err := controller.machineInformer.Informer().GetStore().Update(machine); err != nil {
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
	testConfig := createMachineSetTestConfig(testNamespace, 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

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

func TestControllerMachinesInMachineSet(t *testing.T) {
	testConfig1 := createMachineSetTestConfig("testConfig1", 5, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig1)
	defer stop()

	// Construct a second set of objects and add the machines,
	// nodes and the additional machineset to the existing set of
	// test objects in the controller. This gives us two
	// machinesets, each with their own machines and linked nodes.
	testConfig2 := createMachineSetTestConfig("testConfig2", 5, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	if err := addTestConfigs(t, controller, testConfig2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	machinesInTestObjs1, err := controller.listMachines(testConfig1.spec.namespace, labels.Everything())
	if err != nil {
		t.Fatalf("error listing machines: %v", err)
	}

	machinesInTestObjs2, err := controller.listMachines(testConfig2.spec.namespace, labels.Everything())
	if err != nil {
		t.Fatalf("error listing machines: %v", err)
	}

	actual := len(machinesInTestObjs1) + len(machinesInTestObjs2)
	expected := len(testConfig1.machines) + len(testConfig2.machines)
	if actual != expected {
		t.Fatalf("expected %d machines, got %d", expected, actual)
	}

	// Sort results as order is not guaranteed.
	sort.Slice(machinesInTestObjs1, func(i, j int) bool {
		return machinesInTestObjs1[i].Name < machinesInTestObjs1[j].Name
	})
	sort.Slice(machinesInTestObjs2, func(i, j int) bool {
		return machinesInTestObjs2[i].Name < machinesInTestObjs2[j].Name
	})

	for i, m := range machinesInTestObjs1 {
		if m.Name != testConfig1.machines[i].Name {
			t.Errorf("expected %q, got %q", testConfig1.machines[i].Name, m.Name)
		}
		if m.Namespace != testConfig1.machines[i].Namespace {
			t.Errorf("expected %q, got %q", testConfig1.machines[i].Namespace, m.Namespace)
		}
	}

	for i, m := range machinesInTestObjs2 {
		if m.Name != testConfig2.machines[i].Name {
			t.Errorf("expected %q, got %q", testConfig2.machines[i].Name, m.Name)
		}
		if m.Namespace != testConfig2.machines[i].Namespace {
			t.Errorf("expected %q, got %q", testConfig2.machines[i].Namespace, m.Namespace)
		}
	}

	// Finally everything in the respective objects should be equal.
	if !reflect.DeepEqual(testConfig1.machines, machinesInTestObjs1) {
		t.Fatalf("expected %+v, got %+v", testConfig1.machines, machinesInTestObjs1)
	}
	if !reflect.DeepEqual(testConfig2.machines, machinesInTestObjs2) {
		t.Fatalf("expected %+v, got %+v", testConfig2.machines, machinesInTestObjs2)
	}
}

func TestControllerLookupNodeGroupForNonExistentNode(t *testing.T) {
	testConfig := createMachineSetTestConfig(testNamespace, 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

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

func TestControllerNodeGroupForNodeWithMissingMachineOwner(t *testing.T) {
	test := func(t *testing.T, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		machine := testConfig.machines[0].DeepCopy()
		machine.OwnerReferences = []v1.OwnerReference{}
		if err := controller.machineInformer.Informer().GetStore().Update(newUnstructuredFromMachine(machine)); err != nil {
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
		testConfig := createMachineSetTestConfig(testNamespace, 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})
		test(t, testConfig)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := createMachineDeploymentTestConfig(testNamespace, 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})
		test(t, testConfig)
	})
}

func TestControllerNodeGroupForNodeWithPositiveScalingBounds(t *testing.T) {
	test := func(t *testing.T, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		ng, err := controller.nodeGroupForNode(testConfig.nodes[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// We don't scale from 0 so nodes must belong to a
		// nodegroup that has a scale size of at least 1.
		if ng != nil {
			t.Fatalf("unexpected nodegroup: %v", ng)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		testConfig := createMachineSetTestConfig(testNamespace, 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "1",
		})
		test(t, testConfig)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := createMachineDeploymentTestConfig(testNamespace, 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "1",
		})
		test(t, testConfig)
	})
}

func TestControllerNodeGroups(t *testing.T) {
	assertNodegroupLen := func(t *testing.T, controller *machineController, expected int) {
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

	controller, stop := mustCreateTestController(t)
	defer stop()

	// Test #1: zero nodegroups
	assertNodegroupLen(t, controller, 0)

	// Test #2: add 5 machineset-based nodegroups
	machineSetConfigs := createMachineSetTestConfigs("MachineSet", 5, 1, annotations)
	if err := addTestConfigs(t, controller, machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 5)

	// Test #2: add 2 machinedeployment-based nodegroups
	machineDeploymentConfigs := createMachineDeploymentTestConfigs("MachineDeployment", 2, 1, annotations)
	if err := addTestConfigs(t, controller, machineDeploymentConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 7)

	// Test #3: delete 5 machineset-backed objects
	if err := deleteTestConfigs(t, controller, machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 2)

	// Test #4: delete 2 machinedeployment-backed objects
	if err := deleteTestConfigs(t, controller, machineDeploymentConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 0)

	annotations = map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "1",
	}

	// Test #5: machineset with no scaling bounds results in no nodegroups
	machineSetConfigs = createMachineSetTestConfigs("MachineSet", 5, 1, annotations)
	if err := addTestConfigs(t, controller, machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 0)

	// Test #6: machinedeployment with no scaling bounds results in no nodegroups
	machineDeploymentConfigs = createMachineDeploymentTestConfigs("MachineDeployment", 2, 1, annotations)
	if err := addTestConfigs(t, controller, machineDeploymentConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 0)

	annotations = map[string]string{
		nodeGroupMinSizeAnnotationKey: "-1",
		nodeGroupMaxSizeAnnotationKey: "1",
	}

	// Test #7: machineset with bad scaling bounds results in an error and no nodegroups
	machineSetConfigs = createMachineSetTestConfigs("MachineSet", 5, 1, annotations)
	if err := addTestConfigs(t, controller, machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := controller.nodeGroups(); err == nil {
		t.Fatalf("expected an error")
	}

	// Test #8: machinedeployment with bad scaling bounds results in an error and no nodegroups
	machineDeploymentConfigs = createMachineDeploymentTestConfigs("MachineDeployment", 2, 1, annotations)
	if err := addTestConfigs(t, controller, machineDeploymentConfigs...); err != nil {
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

	test := func(t *testing.T, tc testCase, testConfigs []*testConfig) {
		controller, stop := mustCreateTestController(t, testConfigs...)
		defer stop()

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
			test(t, tc, createMachineSetTestConfigs(testNamespace, tc.nodeGroups, tc.nodesPerGroup, annotations))
		}
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		for _, tc := range testCases {
			test(t, tc, createMachineDeploymentTestConfigs(testNamespace, tc.nodeGroups, tc.nodesPerGroup, annotations))
		}
	})
}

func TestControllerFindMachineFromNodeAnnotation(t *testing.T) {
	testConfig := createMachineSetTestConfig(testNamespace, 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Remove all the provider ID values on all the machines. We
	// want to force findMachineByProviderID() to fallback to
	// searching using the annotation on the node object.
	for _, machine := range testConfig.machines {
		machine.Spec.ProviderID = nil
		if err := controller.machineInformer.Informer().GetStore().Update(newUnstructuredFromMachine(machine)); err != nil {
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
	testConfig := createMachineSetTestConfig(testNamespace, 3, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Remove all linkage between node and machine.
	for _, machine := range testConfig.machines {
		machine.Spec.ProviderID = nil
		if err := controller.machineInformer.Informer().GetStore().Update(newUnstructuredFromMachine(machine)); err != nil {
			t.Fatalf("unexpected error updating machine, got %v", err)
		}
	}
	for _, machine := range testConfig.machines {
		machine.Status.NodeRef = nil
		if err := controller.machineInformer.Informer().GetStore().Update(newUnstructuredFromMachine(machine)); err != nil {
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

	// We removed all linkage - so we should get 0 nodes back.
	if len(nodeNames) != 0 {
		t.Fatalf("expected len=0, got len=%v", len(nodeNames))
	}
}

func TestControllerMachineSetNodeNamesUsingProviderID(t *testing.T) {
	testConfig := createMachineSetTestConfig(testNamespace, 3, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Remove Status.NodeRef.Name on all the machines. We want to
	// force machineSetNodeNames() to only consider the provider
	// ID for lookups.
	for _, machine := range testConfig.machines {
		machine.Status.NodeRef = nil
		if err := controller.machineInformer.Informer().GetStore().Update(newUnstructuredFromMachine(machine)); err != nil {
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
	testConfig := createMachineSetTestConfig(testNamespace, 3, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Remove all the provider ID values on all the machines. We
	// want to force machineSetNodeNames() to fallback to
	// searching using Status.NodeRef.Name.
	for _, machine := range testConfig.machines {
		machine.Spec.ProviderID = nil
		if err := controller.machineInformer.Informer().GetStore().Update(newUnstructuredFromMachine(machine)); err != nil {
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
	expected := "mygroup"
	if err := os.Setenv(CAPIGroupEnvVar, expected); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	observed := getCAPIGroup()
	if observed != expected {
		t.Fatalf("Wrong Version Group detected, expected %q, got %q", expected, observed)
	}

	expected = defaultCAPIGroup
	if err := os.Setenv(CAPIGroupEnvVar, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	observed = getCAPIGroup()
	if observed != expected {
		t.Fatalf("Wrong Version Group detected, expected %q, got %q", expected, observed)
	}
}

func TestControllerGetAPIVersionGroupWithMachineDeployments(t *testing.T) {
	testConfig := createMachineDeploymentTestConfig(testNamespace, 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "1",
	})
	if err := os.Setenv(CAPIGroupEnvVar, customCAPIGroup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testConfig.machineDeployment.TypeMeta.APIVersion = fmt.Sprintf("%s/v1beta1", customCAPIGroup)
	testConfig.machineSet.TypeMeta.APIVersion = fmt.Sprintf("%s/v1beta1", customCAPIGroup)
	for _, machine := range testConfig.machines {
		machine.TypeMeta.APIVersion = fmt.Sprintf("%s/v1beta1", customCAPIGroup)
	}
	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	machineDeployments, err := controller.listMachineDeployments(testNamespace, labels.Everything())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l := len(machineDeployments); l != 1 {
		t.Fatalf("Incorrect number of MachineDeployments, expected 1, got %d", l)
	}

	machineSets, err := controller.listMachineSets(testNamespace, labels.Everything())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l := len(machineSets); l != 1 {
		t.Fatalf("Incorrect number of MachineSets, expected 1, got %d", l)
	}

	machines, err := controller.listMachines(testNamespace, labels.Everything())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l := len(machines); l != 1 {
		t.Fatalf("Incorrect number of Machines, expected 1, got %d", l)
	}

	if err := os.Unsetenv(CAPIGroupEnvVar); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAPIGroupPreferredVersion(t *testing.T) {
	testCases := []struct {
		description      string
		APIGroup         string
		preferredVersion string
		error            bool
	}{
		{
			description:      "find version for default API group",
			APIGroup:         defaultCAPIGroup,
			preferredVersion: "v1alpha3",
			error:            false,
		},
		{
			description:      "find version for another API group",
			APIGroup:         customCAPIGroup,
			preferredVersion: "v1beta1",
			error:            false,
		},
		{
			description:      "API group does not exist",
			APIGroup:         "does.not.exist",
			preferredVersion: "",
			error:            true,
		},
	}

	discoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &clientgotesting.Fake{
			Resources: []*v1.APIResourceList{
				{
					GroupVersion: fmt.Sprintf("%s/v1beta1", customCAPIGroup),
				},
				{
					GroupVersion: fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			version, err := getAPIGroupPreferredVersion(discoveryClient, tc.APIGroup)
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
			APIGroup:     fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
			expected:     true,
			error:        false,
		},
		{
			description:  "false when it does not find resource",
			resourceName: "resourceDoesNotExist",
			APIGroup:     fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
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
			Resources: []*v1.APIResourceList{
				{
					GroupVersion: fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
					APIResources: []v1.APIResource{
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := machineKeyFromFailedProviderID(tc.providerID); got != tc.expected {
				t.Errorf("test case: %s, expected: %q, got: %q", tc.name, tc.expected, got)
			}
		})
	}
}
