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
	"path"
	"sort"
	"strings"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/utils/pointer"
)

const (
	testNamespace = "test-namespace"
)

func TestNodeGroupNewNodeGroupConstructor(t *testing.T) {
	type testCase struct {
		description string
		annotations map[string]string
		errors      bool
		replicas    int32
		minSize     int
		maxSize     int
		nodeCount   int
	}

	var testCases = []testCase{{
		description: "errors because minSize is invalid",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "-1",
			nodeGroupMaxSizeAnnotationKey: "0",
		},
		errors: true,
	}, {
		description: "errors because maxSize is invalid",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "0",
			nodeGroupMaxSizeAnnotationKey: "-1",
		},
		errors: true,
	}, {
		description: "errors because minSize > maxSize",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "0",
		},
		errors: true,
	}, {
		description: "errors because maxSize < minSize",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "0",
		},
		errors: true,
	}, {
		description: "no error: min=0, max=0",
		minSize:     0,
		maxSize:     0,
		replicas:    0,
		errors:      false,
	}, {
		description: "no error: min=0, max=1",
		annotations: map[string]string{
			nodeGroupMaxSizeAnnotationKey: "1",
		},
		minSize:  0,
		maxSize:  1,
		replicas: 0,
		errors:   false,
	}, {
		description: "no error: min=1, max=10, replicas=5",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		},
		minSize:   1,
		maxSize:   10,
		replicas:  5,
		nodeCount: 5,
		errors:    false,
	}}

	newNodeGroup := func(t *testing.T, controller *machineController, testConfig *testConfig) (*nodegroup, error) {
		if testConfig.machineDeployment != nil {
			return newNodegroupFromMachineDeployment(controller, testConfig.machineDeployment)
		}
		return newNodegroupFromMachineSet(controller, testConfig.machineSet)
	}

	test := func(t *testing.T, tc testCase, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t)
		defer stop()

		ng, err := newNodeGroup(t, controller, testConfig)
		if tc.errors && err == nil {
			t.Fatal("expected an error")
		}

		if !tc.errors && ng == nil {
			t.Fatalf("test case logic error: %v", err)
		}

		if tc.errors {
			// if the test case is expected to error then
			// don't assert the remainder
			return
		}

		if ng == nil {
			t.Fatal("expected nodegroup to be non-nil")
		}

		var expectedName string

		switch v := (ng.scalableResource).(type) {
		case *machineSetScalableResource:
			expectedName = testConfig.spec.machineSetName
		case *machineDeploymentScalableResource:
			expectedName = testConfig.spec.machineDeploymentName
		default:
			t.Fatalf("unexpected type: %T", v)
		}

		expectedID := path.Join(testConfig.spec.namespace, expectedName)
		expectedDebug := fmt.Sprintf(debugFormat, expectedID, tc.minSize, tc.maxSize, tc.replicas)

		if ng.Name() != expectedName {
			t.Errorf("expected %q, got %q", expectedName, ng.Name())
		}

		if ng.Namespace() != testConfig.spec.namespace {
			t.Errorf("expected %q, got %q", testConfig.spec.namespace, ng.Namespace())
		}

		if ng.MinSize() != tc.minSize {
			t.Errorf("expected %v, got %v", tc.minSize, ng.MinSize())
		}

		if ng.MaxSize() != tc.maxSize {
			t.Errorf("expected %v, got %v", tc.maxSize, ng.MaxSize())
		}

		if ng.Id() != expectedID {
			t.Errorf("expected %q, got %q", expectedID, ng.Id())
		}

		if ng.Debug() != expectedDebug {
			t.Errorf("expected %q, got %q", expectedDebug, ng.Debug())
		}

		if _, err := ng.TemplateNodeInfo(); err != cloudprovider.ErrNotImplemented {
			t.Error("expected error")
		}

		if exists := ng.Exist(); !exists {
			t.Errorf("expected %t, got %t", true, exists)
		}

		if _, err := ng.Create(); err != cloudprovider.ErrAlreadyExist {
			t.Error("expected error")
		}

		if err := ng.Delete(); err != cloudprovider.ErrNotImplemented {
			t.Error("expected error")
		}

		if result := ng.Autoprovisioned(); result {
			t.Errorf("expected %t, got %t", false, result)
		}

		// We test ng.Nodes() in TestControllerNodeGroupsNodeCount
	}

	t.Run("MachineSet", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				test(t, tc, createMachineSetTestConfig(testNamespace, tc.nodeCount, tc.annotations))
			})
		}
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				test(t, tc, createMachineDeploymentTestConfig(testNamespace, tc.nodeCount, tc.annotations))
			})
		}
	})
}

func TestNodeGroupIncreaseSizeErrors(t *testing.T) {
	type testCase struct {
		description string
		delta       int
		initial     int32
		errorMsg    string
	}

	testCases := []testCase{{
		description: "errors because delta is negative",
		delta:       -1,
		initial:     3,
		errorMsg:    "size increase must be positive",
	}, {
		description: "errors because initial+delta > maxSize",
		delta:       8,
		initial:     3,
		errorMsg:    "size increase too large - desired:11 max:10",
	}}

	test := func(t *testing.T, tc *testCase, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		nodegroups, err := controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if l := len(nodegroups); l != 1 {
			t.Fatalf("expected 1 nodegroup, got %d", l)
		}

		ng := nodegroups[0]
		currReplicas, err := ng.TargetSize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if currReplicas != int(tc.initial) {
			t.Errorf("expected %v, got %v", tc.initial, currReplicas)
		}

		errors := len(tc.errorMsg) > 0

		err = ng.IncreaseSize(tc.delta)
		if errors && err == nil {
			t.Fatal("expected an error")
		}

		if !errors && err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(err.Error(), tc.errorMsg) {
			t.Errorf("expected error message to contain %q, got %q", tc.errorMsg, err.Error())
		}

		switch v := (ng.scalableResource).(type) {
		case *machineSetScalableResource:
			// A nodegroup is immutable; get a fresh copy.
			ms, err := ng.machineController.getMachineSet(ng.Namespace(), ng.Name(), v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(ms.Spec.Replicas, 0); actual != tc.initial {
				t.Errorf("expected %v, got %v", tc.initial, actual)
			}
		case *machineDeploymentScalableResource:
			// A nodegroup is immutable; get a fresh copy.
			md, err := ng.machineController.getMachineDeployment(ng.Namespace(), ng.Name(), v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(md.Spec.Replicas, 0); actual != tc.initial {
				t.Errorf("expected %v, got %v", tc.initial, actual)
			}
		default:
			t.Errorf("unexpected type: %T", v)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				annotations := map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				}
				test(t, &tc, createMachineSetTestConfig(testNamespace, int(tc.initial), annotations))
			})
		}
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				annotations := map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				}
				test(t, &tc, createMachineDeploymentTestConfig(testNamespace, int(tc.initial), annotations))
			})
		}
	})
}

func TestNodeGroupIncreaseSize(t *testing.T) {
	type testCase struct {
		description string
		delta       int
		initial     int32
		expected    int32
	}

	test := func(t *testing.T, tc *testCase, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		nodegroups, err := controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if l := len(nodegroups); l != 1 {
			t.Fatalf("expected 1 nodegroup, got %d", l)
		}

		ng := nodegroups[0]
		currReplicas, err := ng.TargetSize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if currReplicas != int(tc.initial) {
			t.Errorf("initially expected %v, got %v", tc.initial, currReplicas)
		}

		if err := ng.IncreaseSize(tc.delta); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		switch v := (ng.scalableResource).(type) {
		case *machineSetScalableResource:
			// A nodegroup is immutable; get a fresh copy.
			ms, err := ng.machineController.getMachineSet(ng.Namespace(), ng.Name(), v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(ms.Spec.Replicas, 0); actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		case *machineDeploymentScalableResource:
			// A nodegroup is immutable; get a fresh copy.
			md, err := ng.machineController.getMachineDeployment(ng.Namespace(), ng.Name(), v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(md.Spec.Replicas, 0); actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		default:
			t.Errorf("unexpected type: %T", v)
		}
	}

	annotations := map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	}

	t.Run("MachineSet", func(t *testing.T) {
		tc := testCase{
			description: "increase by 1",
			initial:     3,
			expected:    4,
			delta:       1,
		}
		test(t, &tc, createMachineSetTestConfig(testNamespace, int(tc.initial), annotations))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		tc := testCase{
			description: "increase by 1",
			initial:     3,
			expected:    4,
			delta:       1,
		}
		test(t, &tc, createMachineDeploymentTestConfig(testNamespace, int(tc.initial), annotations))
	})
}

func TestNodeGroupDecreaseTargetSize(t *testing.T) {
	type testCase struct {
		description         string
		delta               int
		initial             int32
		targetSizeIncrement int32
		expected            int32
		expectedError       bool
	}

	test := func(t *testing.T, tc *testCase, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		nodegroups, err := controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if l := len(nodegroups); l != 1 {
			t.Fatalf("expected 1 nodegroup, got %d", l)
		}

		ng := nodegroups[0]
		// DecreaseTargetSize should only decrease the size when the current target size of the nodeGroup
		// is bigger than the number existing instances for that group. We force such a scenario with targetSizeIncrement.
		switch v := (ng.scalableResource).(type) {
		case *machineSetScalableResource:
			testConfig.machineSet.Spec.Replicas = int32ptr(*testConfig.machineSet.Spec.Replicas + tc.targetSizeIncrement)
			ms, err := ng.machineapiClient.MachineSets(ng.Namespace()).Update(testConfig.machineSet)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err := controller.machineSetInformer.Informer().GetStore().Add(ms); err != nil {
				t.Fatalf("failed to add new machine: %v", err)
			}
		case *machineDeploymentScalableResource:
			testConfig.machineDeployment.Spec.Replicas = int32ptr(*testConfig.machineDeployment.Spec.Replicas + tc.targetSizeIncrement)
			md, err := ng.machineapiClient.MachineDeployments(ng.Namespace()).Update(testConfig.machineDeployment)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err := controller.machineDeploymentInformer.Informer().GetStore().Add(md); err != nil {
				t.Fatalf("failed to add new machine: %v", err)
			}
		default:
			t.Errorf("unexpected type: %T", v)
		}
		// A nodegroup is immutable; get a fresh copy after adding targetSizeIncrement.
		nodegroups, err = controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ng = nodegroups[0]

		currReplicas, err := ng.TargetSize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if currReplicas != int(tc.initial)+int(tc.targetSizeIncrement) {
			t.Errorf("initially expected %v, got %v", tc.initial, currReplicas)
		}

		if err := ng.DecreaseTargetSize(tc.delta); (err != nil) != tc.expectedError {
			t.Fatalf("expected error: %v, got: %v", tc.expectedError, err)
		}

		switch v := (ng.scalableResource).(type) {
		case *machineSetScalableResource:
			// A nodegroup is immutable; get a fresh copy.
			ms, err := ng.machineController.getMachineSet(ng.Namespace(), ng.Name(), v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(ms.Spec.Replicas, 0); actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		case *machineDeploymentScalableResource:
			// A nodegroup is immutable; get a fresh copy.
			md, err := ng.machineController.getMachineDeployment(ng.Namespace(), ng.Name(), v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(md.Spec.Replicas, 0); actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		default:
			t.Errorf("unexpected type: %T", v)
		}
	}

	annotations := map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	}

	t.Run("MachineSet", func(t *testing.T) {
		tc := testCase{
			description:         "Same number of existing instances and node group target size should error",
			initial:             3,
			targetSizeIncrement: 0,
			expected:            3,
			delta:               -1,
			expectedError:       true,
		}
		test(t, &tc, createMachineSetTestConfig(testNamespace, int(tc.initial), annotations))
	})

	t.Run("MachineSet", func(t *testing.T) {
		tc := testCase{
			description:         "A node group with targe size 4 but only 3 existing instances should decrease by 1",
			initial:             3,
			targetSizeIncrement: 1,
			expected:            3,
			delta:               -1,
		}
		test(t, &tc, createMachineSetTestConfig(testNamespace, int(tc.initial), annotations))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		tc := testCase{
			description:         "Same number of existing instances and node group target size should error",
			initial:             3,
			targetSizeIncrement: 0,
			expected:            3,
			delta:               -1,
			expectedError:       true,
		}
		test(t, &tc, createMachineDeploymentTestConfig(testNamespace, int(tc.initial), annotations))
	})
}

func TestNodeGroupDecreaseSizeErrors(t *testing.T) {
	type testCase struct {
		description string
		delta       int
		initial     int32
		errorMsg    string
	}

	testCases := []testCase{{
		description: "errors because delta is positive",
		delta:       1,
		initial:     3,
		errorMsg:    "size decrease must be negative",
	}, {
		description: "errors because initial+delta < len(nodes)",
		delta:       -1,
		initial:     3,
		errorMsg:    "attempt to delete existing nodes targetSize:3 delta:-1 existingNodes: 3",
	}}

	test := func(t *testing.T, tc *testCase, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		nodegroups, err := controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if l := len(nodegroups); l != 1 {
			t.Fatalf("expected 1 nodegroup, got %d", l)
		}

		ng := nodegroups[0]
		currReplicas, err := ng.TargetSize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if currReplicas != int(tc.initial) {
			t.Errorf("expected %v, got %v", tc.initial, currReplicas)
		}

		errors := len(tc.errorMsg) > 0

		err = ng.DecreaseTargetSize(tc.delta)
		if errors && err == nil {
			t.Fatal("expected an error")
		}

		if !errors && err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(err.Error(), tc.errorMsg) {
			t.Errorf("expected error message to contain %q, got %q", tc.errorMsg, err.Error())
		}

		switch v := (ng.scalableResource).(type) {
		case *machineSetScalableResource:
			// A nodegroup is immutable; get a fresh copy.
			ms, err := ng.machineController.getMachineSet(ng.Namespace(), ng.Name(), v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(ms.Spec.Replicas, 0); actual != tc.initial {
				t.Errorf("expected %v, got %v", tc.initial, actual)
			}
		case *machineDeploymentScalableResource:
			// A nodegroup is immutable; get a fresh copy.
			md, err := ng.machineController.getMachineDeployment(ng.Namespace(), ng.Name(), v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(md.Spec.Replicas, 0); actual != tc.initial {
				t.Errorf("expected %v, got %v", tc.initial, actual)
			}
		default:
			t.Errorf("unexpected type: %T", v)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				annotations := map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				}
				test(t, &tc, createMachineSetTestConfig(testNamespace, int(tc.initial), annotations))
			})
		}
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				annotations := map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				}
				test(t, &tc, createMachineDeploymentTestConfig(testNamespace, int(tc.initial), annotations))
			})
		}
	})
}

func TestNodeGroupDeleteNodes(t *testing.T) {
	test := func(t *testing.T, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

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

		sort.SliceStable(nodeNames, func(i, j int) bool {
			return nodeNames[i].Id < nodeNames[j].Id
		})

		for i := 0; i < len(nodeNames); i++ {
			if nodeNames[i].Id != testConfig.nodes[i].Spec.ProviderID {
				t.Fatalf("expected %q, got %q", testConfig.nodes[i].Spec.ProviderID, nodeNames[i].Id)
			}
		}

		if err := ng.DeleteNodes(testConfig.nodes[5:]); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		for i := 5; i < len(testConfig.machines); i++ {
			machine, err := controller.getMachine(testConfig.machines[i].Namespace, testConfig.machines[i].Name, v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, found := machine.Annotations[machineDeleteAnnotationKey]; !found {
				t.Errorf("expected annotation %q on machine %s", machineDeleteAnnotationKey, machine.Name)
			}
		}

		switch v := (ng.scalableResource).(type) {
		case *machineSetScalableResource:
			updatedMachineSet, err := controller.getMachineSet(testConfig.machineSet.Namespace, testConfig.machineSet.Name, v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(updatedMachineSet.Spec.Replicas, 0); actual != 5 {
				t.Fatalf("expected 5 nodes, got %v", actual)
			}
		case *machineDeploymentScalableResource:
			updatedMachineDeployment, err := controller.getMachineDeployment(testConfig.machineDeployment.Namespace, testConfig.machineDeployment.Name, v1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual := pointer.Int32PtrDerefOr(updatedMachineDeployment.Spec.Replicas, 0); actual != 5 {
				t.Fatalf("expected 5 nodes, got %v", actual)
			}
		default:
			t.Errorf("unexpected type: %T", v)
		}
	}

	// Note: 10 is an upper bound for the number of nodes/replicas
	// Going beyond 10 will break the sorting that happens in the
	// test() function because sort.Strings() will not do natural
	// sorting and the expected semantics in test() will fail.

	t.Run("MachineSet", func(t *testing.T) {
		test(t, createMachineSetTestConfig(testNamespace, 10, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		test(t, createMachineDeploymentTestConfig(testNamespace, 10, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		}))
	})
}

func TestNodeGroupMachineSetDeleteNodesWithMismatchedNodes(t *testing.T) {
	test := func(t *testing.T, expected int, testConfigs []*testConfig) {
		t.Helper()
		testConfig0, testConfig1 := testConfigs[0], testConfigs[1]
		controller, stop := mustCreateTestController(t, testConfigs...)
		defer stop()

		nodegroups, err := controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if l := len(nodegroups); l != expected {
			t.Fatalf("expected %d, got %d", expected, l)
		}

		ng0, err := controller.nodeGroupForNode(testConfig0.nodes[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ng1, err := controller.nodeGroupForNode(testConfig1.nodes[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Deleting nodes that are not in ng0 should fail.
		err0 := ng0.DeleteNodes(testConfig1.nodes)
		if err0 == nil {
			t.Error("expected an error")
		}

		expectedErr0 := `node "test-namespace1-machineset-0-nodeid-0" doesn't belong to node group "test-namespace0/machineset-0"`
		if testConfig0.machineDeployment != nil {
			expectedErr0 = `node "test-namespace1-machineset-0-nodeid-0" doesn't belong to node group "test-namespace0/machinedeployment-0"`
		}

		if !strings.Contains(err0.Error(), expectedErr0) {
			t.Errorf("expected: %q, got: %q", expectedErr0, err0.Error())
		}

		// Deleting nodes that are not in ng1 should fail.
		err1 := ng1.DeleteNodes(testConfig0.nodes)
		if err1 == nil {
			t.Error("expected an error")
		}

		expectedErr1 := `node "test-namespace0-machineset-0-nodeid-0" doesn't belong to node group "test-namespace1/machineset-0"`
		if testConfig1.machineDeployment != nil {
			expectedErr1 = `node "test-namespace0-machineset-0-nodeid-0" doesn't belong to node group "test-namespace1/machinedeployment-0"`
		}

		if !strings.Contains(err1.Error(), expectedErr1) {
			t.Errorf("expected: %q, got: %q", expectedErr1, err1.Error())
		}

		// Deleting from correct node group should fail because
		// replicas would become <= 0.
		if err := ng0.DeleteNodes(testConfig0.nodes); err == nil {
			t.Error("expected error")
		}

		// Deleting from correct node group should fail because
		// replicas would become <= 0.
		if err := ng1.DeleteNodes(testConfig1.nodes); err == nil {
			t.Error("expected error")
		}
	}

	annotations := map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "3",
	}

	t.Run("MachineSet", func(t *testing.T) {
		testConfig0 := createMachineSetTestConfigs(testNamespace+"0", 1, 2, annotations)
		testConfig1 := createMachineSetTestConfigs(testNamespace+"1", 1, 2, annotations)
		test(t, 2, append(testConfig0, testConfig1...))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig0 := createMachineDeploymentTestConfigs(testNamespace+"0", 1, 2, annotations)
		testConfig1 := createMachineDeploymentTestConfigs(testNamespace+"1", 1, 2, annotations)
		test(t, 2, append(testConfig0, testConfig1...))
	})
}

	addDeletionTimestamp := func(t *testing.T, controller *machineController, machine *Machine) error {
		// Simulate delete that would have happened if the
		// Machine API controllers were running Don't actually
		// delete since the fake client does not support
		// finalizers.
		now := v1.Now()
		machine.DeletionTimestamp = &now
		return controller.machineInformer.Informer().GetStore().Update(newUnstructuredFromMachine(machine))
	}

		// Assert that we have no DeletionTimestamp
		for i := 7; i < len(testConfig.machines); i++ {
			if !testConfig.machines[i].ObjectMeta.DeletionTimestamp.IsZero() {
				t.Fatalf("unexpected DeletionTimestamp")
			}
		}
		if err := ng.DeleteNodes(testConfig.nodes[7:]); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for i := 7; i < len(testConfig.machines); i++ {
			if err := addDeletionTimestamp(t, controller, testConfig.machines[i]); err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if testConfig.machines[i].ObjectMeta.DeletionTimestamp.IsZero() {
				t.Fatalf("expected a DeletionTimestamp")
			}
		}
		// TODO(frobware) We have a flaky test here because we
		// just called Delete and Update and the next call to
		// controller.nodeGroups() will sometimes get stale
		// objects from the (fakeclient) store. To fix this we
		// should update the test machinery so that individual
		// tests can have callbacks on Add/Update/Delete on
		// each of the respective informers. We should then
		// override those callbacks here in this test to add
		// rendezvous points so that we wait until all objects
		// have been updated before we go and get them again.
		//
		// Running this test with a 500ms duration I see:
		//
		// $ ./stress ./clusterapi.test -test.run TestNodeGroupDeleteNodesTwice -test.count 5 | ts | ts -i
		// 00:00:05 Feb 27 14:29:36 0 runs so far, 0 failures
		// 00:00:05 Feb 27 14:29:41 8 runs so far, 0 failures
		// 00:00:05 Feb 27 14:29:46 16 runs so far, 0 failures
		// 00:00:05 Feb 27 14:29:51 24 runs so far, 0 failures
		// 00:00:05 Feb 27 14:29:56 32 runs so far, 0 failures
		// ...
		// 00:00:05 Feb 27 14:31:01 112 runs so far, 0 failures
		// 00:00:05 Feb 27 14:31:06 120 runs so far, 0 failures
		// 00:00:05 Feb 27 14:31:11 128 runs so far, 0 failures
		// 00:00:05 Feb 27 14:31:16 136 runs so far, 0 failures
		// 00:00:05 Feb 27 14:31:21 144 runs so far, 0 failures
		//
		// To make sure we don't run into any flakes in CI
		// I've chosen to make this sleep duration 3s.
		time.Sleep(3 * time.Second)

		nodegroups, err = controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ng = nodegroups[0]

		// Attempt to delete the nodes again which verifies
		// that nodegroup.DeleteNodes() skips over nodes that
		// have a non-nil DeletionTimestamp value.
		if err := ng.DeleteNodes(testConfig.nodes[7:]); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		actualSize, err := ng.TargetSize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedSize := len(testConfig.machines) - len(testConfig.machines[7:])
		if actualSize != expectedSize {
			t.Fatalf("expected %d nodes, got %d", expectedSize, actualSize)
		}
