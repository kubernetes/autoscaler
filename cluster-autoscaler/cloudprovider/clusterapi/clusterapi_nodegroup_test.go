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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	gpuapis "k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/tools/cache"
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
		expectNil   bool
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
		expectNil:   true,
	}, {
		description: "no error: min=0, max=1",
		annotations: map[string]string{
			nodeGroupMaxSizeAnnotationKey: "1",
		},
		minSize:   0,
		maxSize:   1,
		replicas:  0,
		errors:    false,
		expectNil: true,
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
		expectNil: true,
	}, {
		description: "no error and expect notNil: min=max=2",
		annotations: map[string]string{
			nodeGroupMinSizeAnnotationKey: "2",
			nodeGroupMaxSizeAnnotationKey: "2",
		},
		nodeCount: 1,
		minSize:   2,
		maxSize:   2,
		replicas:  1,
		errors:    false,
		expectNil: false,
	}}

	newNodeGroup := func(controller *machineController, testConfig *testConfig) (*nodegroup, error) {
		if testConfig.machineDeployment != nil {
			return newNodeGroupFromScalableResource(controller, testConfig.machineDeployment)
		}
		return newNodeGroupFromScalableResource(controller, testConfig.machineSet)
	}

	test := func(t *testing.T, tc testCase, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		ng, err := newNodeGroup(controller, testConfig)
		if tc.errors && err == nil {
			t.Fatal("expected an error")
		}

		if tc.errors {
			// if the test case is expected to error then
			// don't assert the remainder
			return
		}

		if tc.expectNil && ng == nil {
			// if the test case is expected to return nil then
			// don't assert the remainder
			return
		}

		if ng == nil {
			t.Fatal("expected nodegroup to be non-nil")
		}

		var expectedName, expectedKind string

		if testConfig.machineDeployment != nil {
			expectedKind = machineDeploymentKind
			expectedName = testConfig.spec.machineDeploymentName
		} else {
			expectedKind = machineSetKind
			expectedName = testConfig.spec.machineSetName
		}

		expectedID := path.Join(expectedKind, testConfig.spec.namespace, expectedName)
		expectedDebug := fmt.Sprintf(debugFormat, expectedID, tc.minSize, tc.maxSize, tc.replicas)

		if ng.scalableResource.Name() != expectedName {
			t.Errorf("expected %q, got %q", expectedName, ng.scalableResource.Name())
		}

		if ng.scalableResource.Namespace() != testConfig.spec.namespace {
			t.Errorf("expected %q, got %q", testConfig.spec.namespace, ng.scalableResource.Namespace())
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
				test(t, tc, createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), tc.nodeCount, tc.annotations, nil))
			})
		}
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				test(t, tc, createMachineDeploymentTestConfig(RandomString(6), RandomString(6), RandomString(6), tc.nodeCount, tc.annotations, nil))
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

		ng := nodegroups[0].(*nodegroup)
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

		gvr, err := ng.scalableResource.GroupVersionResource()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scalableResource, err := ng.machineController.managementScaleClient.Scales(testConfig.spec.namespace).
			Get(context.TODO(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if scalableResource.Spec.Replicas != tc.initial {
			t.Errorf("expected %v, got %v", tc.initial, scalableResource.Spec.Replicas)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				annotations := map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				}
				test(t, &tc, createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
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
				test(t, &tc, createMachineDeploymentTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
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

		ng := nodegroups[0].(*nodegroup)
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

		gvr, err := ng.scalableResource.GroupVersionResource()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scalableResource, err := ng.machineController.managementScaleClient.Scales(ng.scalableResource.Namespace()).
			Get(context.TODO(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if scalableResource.Spec.Replicas != tc.expected {
			t.Errorf("expected %v, got %v", tc.expected, scalableResource.Spec.Replicas)
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
		test(t, &tc, createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		tc := testCase{
			description: "increase by 1",
			initial:     3,
			expected:    4,
			delta:       1,
		}
		test(t, &tc, createMachineDeploymentTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
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

		ng := nodegroups[0].(*nodegroup)

		gvr, err := ng.scalableResource.GroupVersionResource()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// DecreaseTargetSize should only decrease the size when the current target size of the nodeGroup
		// is bigger than the number existing instances for that group. We force such a scenario with targetSizeIncrement.
		scalableResource, err := controller.managementScaleClient.Scales(testConfig.spec.namespace).
			Get(context.TODO(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ch := make(chan error)
		checkDone := func(obj interface{}) (bool, error) {
			u, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return false, nil
			}
			if u.GetResourceVersion() != scalableResource.GetResourceVersion() {
				return false, nil
			}
			ng, err := newNodeGroupFromScalableResource(controller, u)
			if err != nil {
				return true, fmt.Errorf("unexpected error: %v", err)
			}
			if ng == nil {
				return false, nil
			}
			currReplicas, err := ng.TargetSize()
			if err != nil {
				return true, fmt.Errorf("unexpected error: %v", err)
			}

			if currReplicas != int(tc.initial)+int(tc.targetSizeIncrement) {
				return true, fmt.Errorf("expected %v, got %v", tc.initial+tc.targetSizeIncrement, currReplicas)
			}

			if err := ng.DecreaseTargetSize(tc.delta); (err != nil) != tc.expectedError {
				return true, fmt.Errorf("expected error: %v, got: %v", tc.expectedError, err)
			}

			scalableResource, err := controller.managementScaleClient.Scales(testConfig.spec.namespace).
				Get(context.TODO(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
			if err != nil {
				return true, fmt.Errorf("unexpected error: %v", err)
			}

			if scalableResource.Spec.Replicas != tc.expected {
				return true, fmt.Errorf("expected %v, got %v", tc.expected, scalableResource.Spec.Replicas)
			}
			return true, nil
		}
		handler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				match, err := checkDone(obj)
				if match {
					ch <- err
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				match, err := checkDone(newObj)
				if match {
					ch <- err
				}
			},
		}
		if _, err := controller.machineSetInformer.Informer().AddEventHandler(handler); err != nil {
			t.Fatalf("unexpected error adding event handler for machineSetInformer: %v", err)
		}
		if _, err := controller.machineDeploymentInformer.Informer().AddEventHandler(handler); err != nil {
			t.Fatalf("unexpected error adding event handler for machineDeploymentInformer: %v", err)
		}

		scalableResource.Spec.Replicas += tc.targetSizeIncrement

		_, err = ng.machineController.managementScaleClient.Scales(ng.scalableResource.Namespace()).
			Update(context.TODO(), gvr.GroupResource(), scalableResource, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lastErr := fmt.Errorf("no updates received yet")
		for lastErr != nil {
			select {
			case err = <-ch:
				lastErr = err
			case <-time.After(1 * time.Second):
				t.Fatal(fmt.Errorf("timeout while waiting for update. Last error was: %v", lastErr))
			}
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
		test(t, &tc, createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
	})

	t.Run("MachineSet", func(t *testing.T) {
		tc := testCase{
			description:         "A node group with target size 4 but only 3 existing instances should decrease by 1",
			initial:             3,
			targetSizeIncrement: 1,
			expected:            3,
			delta:               -1,
		}
		test(t, &tc, createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
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
		test(t, &tc, createMachineDeploymentTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
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

		ng := nodegroups[0].(*nodegroup)
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

		gvr, err := ng.scalableResource.GroupVersionResource()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scalableResource, err := ng.machineController.managementScaleClient.Scales(testConfig.spec.namespace).
			Get(context.TODO(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if scalableResource.Spec.Replicas != tc.initial {
			t.Errorf("expected %v, got %v", tc.initial, scalableResource.Spec.Replicas)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				annotations := map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				}
				test(t, &tc, createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
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
				test(t, &tc, createMachineDeploymentTestConfig(RandomString(6), RandomString(6), RandomString(6), int(tc.initial), annotations, nil))
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

		ng := nodegroups[0].(*nodegroup)
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
			machine, err := controller.managementClient.Resource(controller.machineResource).
				Namespace(testConfig.spec.namespace).
				Get(context.TODO(), testConfig.machines[i].GetName(), metav1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, found := machine.GetAnnotations()[machineDeleteAnnotationKey]; !found {
				t.Errorf("expected annotation %q on machine %s", machineDeleteAnnotationKey, machine.GetName())
			}
		}

		gvr, err := ng.scalableResource.GroupVersionResource()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scalableResource, err := ng.machineController.managementScaleClient.Scales(testConfig.spec.namespace).
			Get(context.TODO(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if scalableResource.Spec.Replicas != 5 {
			t.Errorf("expected 5, got %v", scalableResource.Spec.Replicas)
		}
	}

	// Note: 10 is an upper bound for the number of nodes/replicas
	// Going beyond 10 will break the sorting that happens in the
	// test() function because sort.Strings() will not do natural
	// sorting and the expected semantics in test() will fail.

	t.Run("MachineSet", func(t *testing.T) {
		test(
			t,
			createMachineSetTestConfig(
				RandomString(6),
				RandomString(6),
				RandomString(6),
				10,
				map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				},
				nil,
			),
		)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		test(
			t,
			createMachineDeploymentTestConfig(
				RandomString(6),
				RandomString(6),
				RandomString(6),
				10,
				map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				},
				nil,
			),
		)
	})
}

func TestNodeGroupMachineSetDeleteNodesWithMismatchedNodes(t *testing.T) {
	test := func(t *testing.T, expected int, testConfigs []*testConfig) {
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

		expectedErrSubstring := "doesn't belong to node group"

		if !strings.Contains(err0.Error(), expectedErrSubstring) {
			t.Errorf("expected error: %q to contain: %q", err0.Error(), expectedErrSubstring)
		}

		// Deleting nodes that are not in ng1 should fail.
		err1 := ng1.DeleteNodes(testConfig0.nodes)
		if err1 == nil {
			t.Error("expected an error")
		}

		if !strings.Contains(err1.Error(), expectedErrSubstring) {
			t.Errorf("expected error: %q to contain: %q", err0.Error(), expectedErrSubstring)
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
		namespace := RandomString(6)
		clusterName := RandomString(6)
		testConfig0 := createMachineSetTestConfigs(namespace, clusterName, RandomString(6), 1, 2, annotations, nil)
		testConfig1 := createMachineSetTestConfigs(namespace, clusterName, RandomString(6), 1, 2, annotations, nil)
		test(t, 2, append(testConfig0, testConfig1...))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		namespace := RandomString(6)
		clusterName := RandomString(6)
		testConfig0 := createMachineDeploymentTestConfigs(namespace, clusterName, RandomString(6), 1, 2, annotations, nil)
		testConfig1 := createMachineDeploymentTestConfigs(namespace, clusterName, RandomString(6), 1, 2, annotations, nil)
		test(t, 2, append(testConfig0, testConfig1...))
	})
}

func TestNodeGroupDeleteNodesTwice(t *testing.T) {
	addDeletionTimestampToMachine := func(controller *machineController, node *corev1.Node) error {
		m, err := controller.findMachineByProviderID(normalizedProviderString(node.Spec.ProviderID))
		if err != nil {
			return err
		}

		// Simulate delete that would have happened if the
		// Machine API controllers were running Don't actually
		// delete since the fake client does not support
		// finalizers.
		now := metav1.Now()

		m.SetDeletionTimestamp(&now)

		if _, err := controller.managementClient.Resource(controller.machineResource).
			Namespace(m.GetNamespace()).Update(context.TODO(), m, metav1.UpdateOptions{}); err != nil {
			return err
		}

		return nil
	}

	// This is the size we expect the NodeGroup to be after we have called DeleteNodes.
	// We need at least 8 nodes for this test to be valid.
	expectedSize := 7

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

		ng := nodegroups[0].(*nodegroup)
		nodeNames, err := ng.Nodes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that the test case is valid before executing DeleteNodes
		// 1. We must have at least 1 more node than the expected size otherwise DeleteNodes is a no-op
		// 2. MinSize must be less than the expected size otherwise a second call to DeleteNodes may
		//    not make the nodegroup size less than the expected size.
		if len(nodeNames) <= expectedSize {
			t.Fatalf("expected more nodes than the expected size: %d <= %d", len(nodeNames), expectedSize)
		}
		if ng.MinSize() >= expectedSize {
			t.Fatalf("expected min size to be less than expected size: %d >= %d", ng.MinSize(), expectedSize)
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

		// These are the nodes which are over the final expectedSize
		nodesToBeDeleted := testConfig.nodes[expectedSize:]

		// Assert that we have no DeletionTimestamp
		for i := expectedSize; i < len(testConfig.machines); i++ {
			if !testConfig.machines[i].GetDeletionTimestamp().IsZero() {
				t.Fatalf("unexpected DeletionTimestamp")
			}
		}

		// Delete all nodes over the expectedSize
		if err := ng.DeleteNodes(nodesToBeDeleted); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, node := range nodesToBeDeleted {
			if err := addDeletionTimestampToMachine(controller, node); err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
		}

		// Wait for the machineset to have been updated
		if err := wait.PollImmediate(100*time.Millisecond, 5*time.Second, func() (bool, error) {
			nodegroups, err = controller.nodeGroups()
			if err != nil {
				return false, err
			}
			targetSize, err := nodegroups[0].TargetSize()
			if err != nil {
				return false, err
			}
			return targetSize == expectedSize, nil
		}); err != nil {
			t.Fatalf("unexpected error waiting for nodegroup to be expected size: %v", err)
		}

		nodegroups, err = controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ng = nodegroups[0].(*nodegroup)

		// Check the nodegroup is at the expected size
		actualSize, err := ng.TargetSize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if actualSize != expectedSize {
			t.Fatalf("expected %d nodes, got %d", expectedSize, actualSize)
		}

		// Check that the machines deleted in the last run have DeletionTimestamp's
		// when fetched from the API
		for _, node := range nodesToBeDeleted {
			// Ensure the update has propogated
			if err := wait.PollImmediate(100*time.Millisecond, 5*time.Minute, func() (bool, error) {
				m, err := controller.findMachineByProviderID(normalizedProviderString(node.Spec.ProviderID))
				if err != nil {
					return false, err
				}
				return !m.GetDeletionTimestamp().IsZero(), nil
			}); err != nil {
				t.Fatalf("unexpected error waiting for machine to have deletion timestamp: %v", err)
			}
		}

		// Attempt to delete the nodes again which verifies
		// that nodegroup.DeleteNodes() skips over nodes that
		// have a non-nil DeletionTimestamp value.
		if err := ng.DeleteNodes(nodesToBeDeleted); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		gvr, err := ng.scalableResource.GroupVersionResource()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scalableResource, err := ng.machineController.managementScaleClient.Scales(testConfig.spec.namespace).
			Get(context.TODO(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})

		if scalableResource.Spec.Replicas != int32(expectedSize) {
			t.Errorf("expected %v, got %v", expectedSize, scalableResource.Spec.Replicas)
		}
	}

	// Note: 10 is an upper bound for the number of nodes/replicas
	// Going beyond 10 will break the sorting that happens in the
	// test() function because sort.Strings() will not do natural
	// sorting and the expected semantics in test() will fail.

	t.Run("MachineSet", func(t *testing.T) {
		test(
			t,
			createMachineSetTestConfig(
				RandomString(6),
				RandomString(6),
				RandomString(6),
				10,
				map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				},
				nil,
			),
		)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		test(
			t,
			createMachineDeploymentTestConfig(
				RandomString(6),
				RandomString(6),
				RandomString(6),
				10,
				map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				},
				nil,
			),
		)
	})
}

func TestNodeGroupDeleteNodesSequential(t *testing.T) {
	// This is the size we expect the NodeGroup to be after we have called DeleteNodes.
	// We need at least 8 nodes for this test to be valid.
	expectedSize := 7

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

		ng := nodegroups[0].(*nodegroup)
		nodeNames, err := ng.Nodes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that the test case is valid before executing DeleteNodes
		// 1. We must have at least 1 more node than the expected size otherwise DeleteNodes is a no-op
		// 2. MinSize must be less than the expected size otherwise a second call to DeleteNodes may
		//    not make the nodegroup size less than the expected size.
		if len(nodeNames) <= expectedSize {
			t.Fatalf("expected more nodes than the expected size: %d <= %d", len(nodeNames), expectedSize)
		}
		if ng.MinSize() >= expectedSize {
			t.Fatalf("expected min size to be less than expected size: %d >= %d", ng.MinSize(), expectedSize)
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

		// These are the nodes which are over the final expectedSize
		nodesToBeDeleted := testConfig.nodes[expectedSize:]

		// Assert that we have no DeletionTimestamp
		for i := expectedSize; i < len(testConfig.machines); i++ {
			if !testConfig.machines[i].GetDeletionTimestamp().IsZero() {
				t.Fatalf("unexpected DeletionTimestamp")
			}
		}

		// When the core autoscaler scales down nodes, it fetches the node group for each node in separate
		// go routines and then scales them down individually, we use a lock to ensure the scale down is
		// performed sequentially but this means that the cached scalable resource may not be up to date.
		// We need to replicate this out of date nature by constructing the node groups then deleting.
		nodeToNodeGroup := make(map[*corev1.Node]*nodegroup)

		for _, node := range nodesToBeDeleted {
			nodeGroup, err := controller.nodeGroupForNode(node)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			nodeToNodeGroup[node] = nodeGroup
		}

		for node, nodeGroup := range nodeToNodeGroup {
			if err := nodeGroup.DeleteNodes([]*corev1.Node{node}); err != nil {
				t.Fatalf("unexpected error deleting node: %v", err)
			}
		}

		nodegroups, err = controller.nodeGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ng = nodegroups[0].(*nodegroup)

		// Check the nodegroup is at the expected size
		actualSize, err := ng.scalableResource.Replicas()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if actualSize != expectedSize {
			t.Fatalf("expected %d nodes, got %d", expectedSize, actualSize)
		}

		gvr, err := ng.scalableResource.GroupVersionResource()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scalableResource, err := ng.machineController.managementScaleClient.Scales(testConfig.spec.namespace).
			Get(context.TODO(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if scalableResource.Spec.Replicas != int32(expectedSize) {
			t.Errorf("expected %v, got %v", expectedSize, scalableResource.Spec.Replicas)
		}
	}

	// Note: 10 is an upper bound for the number of nodes/replicas
	// Going beyond 10 will break the sorting that happens in the
	// test() function because sort.Strings() will not do natural
	// sorting and the expected semantics in test() will fail.

	t.Run("MachineSet", func(t *testing.T) {
		test(
			t,
			createMachineSetTestConfig(
				RandomString(6),
				RandomString(6),
				RandomString(6),
				10,
				map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				},
				nil,
			),
		)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		test(
			t,
			createMachineDeploymentTestConfig(
				RandomString(6),
				RandomString(6),
				RandomString(6),
				10,
				map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				},
				nil,
			),
		)
	})
}

func TestNodeGroupWithFailedMachine(t *testing.T) {
	test := func(t *testing.T, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		// Simulate a failed machine
		machine := testConfig.machines[3].DeepCopy()

		unstructured.RemoveNestedField(machine.Object, "spec", "providerID")
		if err := unstructured.SetNestedField(machine.Object, "FailureMessage", "status", "failureMessage"); err != nil {
			t.Fatalf("unexpected error setting nested field: %v", err)
		}

		if err := updateResource(controller.managementClient, controller.machineInformer, controller.machineResource, machine); err != nil {
			t.Fatalf("unexpected error updating machine, got %v", err)
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

		sort.SliceStable(nodeNames, func(i, j int) bool {
			return nodeNames[i].Id < nodeNames[j].Id
		})

		// The failed machine key is sorted to the first index
		failedMachineID := fmt.Sprintf("%s%s_%s", failedMachinePrefix, machine.GetNamespace(), machine.GetName())
		if nodeNames[0].Id != failedMachineID {
			t.Fatalf("expected %q, got %q", failedMachineID, nodeNames[0].Id)
		}

		for i := 1; i < len(nodeNames); i++ {
			// Fix the indexing due the failed machine being removed from the list
			var nodeIndex int
			if i < 4 {
				// for nodes 0, 1, 2
				nodeIndex = i - 1
			} else {
				// for nodes 4 onwards
				nodeIndex = i
			}

			if nodeNames[i].Id != testConfig.nodes[nodeIndex].Spec.ProviderID {
				t.Fatalf("expected %q, got %q", testConfig.nodes[nodeIndex].Spec.ProviderID, nodeNames[i].Id)
			}
		}
	}

	// Note: 10 is an upper bound for the number of nodes/replicas
	// Going beyond 10 will break the sorting that happens in the
	// test() function because sort.Strings() will not do natural
	// sorting and the expected semantics in test() will fail.

	t.Run("MachineSet", func(t *testing.T) {
		test(
			t,
			createMachineSetTestConfig(
				RandomString(6),
				RandomString(6),
				RandomString(6),
				10,
				map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				},
				nil,
			),
		)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		test(
			t,
			createMachineDeploymentTestConfig(
				RandomString(6),
				RandomString(6),
				RandomString(6),
				10,
				map[string]string{
					nodeGroupMinSizeAnnotationKey: "1",
					nodeGroupMaxSizeAnnotationKey: "10",
				},
				nil,
			),
		)
	})
}

func TestNodeGroupTemplateNodeInfo(t *testing.T) {
	enableScaleAnnotations := map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	}

	type testCaseConfig struct {
		nodeLabels         map[string]string
		includeNodes       bool
		expectedErr        error
		expectedCapacity   map[corev1.ResourceName]int64
		expectedNodeLabels map[string]string
	}

	testCases := []struct {
		name                 string
		nodeGroupAnnotations map[string]string
		config               testCaseConfig
	}{
		{
			name: "When the NodeGroup cannot scale from zero",
			config: testCaseConfig{
				expectedErr: cloudprovider.ErrNotImplemented,
			},
		},
		{
			name: "When the NodeGroup can scale from zero",
			nodeGroupAnnotations: map[string]string{
				memoryKey:   "2048Mi",
				cpuKey:      "2",
				gpuTypeKey:  gpuapis.ResourceNvidiaGPU,
				gpuCountKey: "1",
			},
			config: testCaseConfig{
				expectedErr: nil,
				nodeLabels: map[string]string{
					"kubernetes.io/os":   "linux",
					"kubernetes.io/arch": "amd64",
				},
				expectedCapacity: map[corev1.ResourceName]int64{
					corev1.ResourceCPU:        2,
					corev1.ResourceMemory:     2048 * 1024 * 1024,
					corev1.ResourcePods:       110,
					gpuapis.ResourceNvidiaGPU: 1,
				},
				expectedNodeLabels: map[string]string{
					"kubernetes.io/os":       "linux",
					"kubernetes.io/arch":     "amd64",
					"kubernetes.io/hostname": "random value",
				},
			},
		},
		{
			name: "When the NodeGroup can scale from zero, the label capacity annotations merge with the pre-built node labels and take precedence if the same key is defined in both",
			nodeGroupAnnotations: map[string]string{
				memoryKey:   "2048Mi",
				cpuKey:      "2",
				gpuTypeKey:  gpuapis.ResourceNvidiaGPU,
				gpuCountKey: "1",
				labelsKey:   "kubernetes.io/arch=arm64,my-custom-label=custom-value",
			},
			config: testCaseConfig{
				expectedErr: nil,
				expectedCapacity: map[corev1.ResourceName]int64{
					corev1.ResourceCPU:        2,
					corev1.ResourceMemory:     2048 * 1024 * 1024,
					corev1.ResourcePods:       110,
					gpuapis.ResourceNvidiaGPU: 1,
				},
				expectedNodeLabels: map[string]string{
					"kubernetes.io/os":       "linux",
					"kubernetes.io/arch":     "arm64",
					"kubernetes.io/hostname": "random value",
					"my-custom-label":        "custom-value",
				},
			},
		},
		{
			name: "When the NodeGroup can scale from zero and the Node still exists, it includes the known node labels",
			nodeGroupAnnotations: map[string]string{
				memoryKey: "2048Mi",
				cpuKey:    "2",
			},
			config: testCaseConfig{
				includeNodes: true,
				expectedErr:  nil,
				nodeLabels: map[string]string{
					"kubernetes.io/os":                 "windows",
					"kubernetes.io/arch":               "arm64",
					"node.kubernetes.io/instance-type": "instance1",
				},
				expectedCapacity: map[corev1.ResourceName]int64{
					corev1.ResourceCPU:    2,
					corev1.ResourceMemory: 2048 * 1024 * 1024,
					corev1.ResourcePods:   110,
				},
				expectedNodeLabels: map[string]string{
					"kubernetes.io/hostname":           "random value",
					"kubernetes.io/os":                 "windows",
					"kubernetes.io/arch":               "arm64",
					"node.kubernetes.io/instance-type": "instance1",
				},
			},
		},
	}

	test := func(t *testing.T, testConfig *testConfig, config testCaseConfig) {
		if config.includeNodes {
			for i := range testConfig.nodes {
				testConfig.nodes[i].SetLabels(config.nodeLabels)
			}
		} else {
			testConfig.nodes = []*corev1.Node{}
		}

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
		nodeInfo, err := ng.TemplateNodeInfo()
		if config.expectedErr != nil {
			if err != config.expectedErr {
				t.Fatalf("expected error: %v, but got: %v", config.expectedErr, err)
			}
			return
		}

		nodeAllocatable := nodeInfo.Node().Status.Allocatable
		nodeCapacity := nodeInfo.Node().Status.Capacity
		for resource, expectedCapacity := range config.expectedCapacity {
			if gotAllocatable, ok := nodeAllocatable[resource]; !ok {
				t.Errorf("Expected allocatable to have resource %q, resource not found", resource)
			} else if gotAllocatable.Value() != expectedCapacity {
				t.Errorf("Expected allocatable %q: %+v, Got: %+v", resource, expectedCapacity, gotAllocatable.Value())
			}

			if gotCapactiy, ok := nodeCapacity[resource]; !ok {
				t.Errorf("Expected capacity to have resource %q, resource not found", resource)
			} else if gotCapactiy.Value() != expectedCapacity {
				t.Errorf("Expected capacity %q: %+v, Got: %+v", resource, expectedCapacity, gotCapactiy.Value())
			}
		}

		if len(nodeInfo.Node().GetLabels()) != len(config.expectedNodeLabels) {
			t.Errorf("Expected node labels to have len: %d, but got: %d, labels are: %v", len(config.expectedNodeLabels), len(nodeInfo.Node().GetLabels()), nodeInfo.Node().GetLabels())
		}
		for key, value := range nodeInfo.Node().GetLabels() {
			// Exclude the hostname label as it is randomized
			if key != corev1.LabelHostname {
				if expected, ok := config.expectedNodeLabels[key]; ok {
					if value != expected {
						t.Errorf("Expected node label %q: %q, Got: %q", key, config.expectedNodeLabels[key], value)
					}
				} else {
					t.Errorf("Expected node label %q to exist in node", key)
				}
			}
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("MachineSet", func(t *testing.T) {
				test(
					t,
					createMachineSetTestConfig(
						testNamespace,
						RandomString(6),
						RandomString(6),
						10,
						cloudprovider.JoinStringMaps(enableScaleAnnotations, tc.nodeGroupAnnotations),
						nil,
					),
					tc.config,
				)
			})

			t.Run("MachineDeployment", func(t *testing.T) {
				test(
					t,
					createMachineDeploymentTestConfig(
						testNamespace,
						RandomString(6),
						RandomString(6),
						10,
						cloudprovider.JoinStringMaps(enableScaleAnnotations, tc.nodeGroupAnnotations),
						nil,
					),
					tc.config,
				)
			})
		})
	}

}

func TestNodeGroupGetOptions(t *testing.T) {
	enableScaleAnnotations := map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	}

	defaultOptions := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.1,
		ScaleDownGpuUtilizationThreshold: 0.2,
		ScaleDownUnneededTime:            time.Second,
		ScaleDownUnreadyTime:             time.Minute,
		MaxNodeProvisionTime:             15 * time.Minute,
	}

	cases := []struct {
		desc     string
		opts     map[string]string
		expected *config.NodeGroupAutoscalingOptions
	}{
		{
			desc:     "return provided defaults on empty metadata",
			opts:     map[string]string{},
			expected: &defaultOptions,
		},
		{
			desc: "return specified options",
			opts: map[string]string{
				config.DefaultScaleDownGpuUtilizationThresholdKey: "0.6",
				config.DefaultScaleDownUtilizationThresholdKey:    "0.7",
				config.DefaultScaleDownUnneededTimeKey:            "1h",
				config.DefaultScaleDownUnreadyTimeKey:             "30m",
				config.DefaultMaxNodeProvisionTimeKey:             "60m",
			},
			expected: &config.NodeGroupAutoscalingOptions{
				ScaleDownGpuUtilizationThreshold: 0.6,
				ScaleDownUtilizationThreshold:    0.7,
				ScaleDownUnneededTime:            time.Hour,
				ScaleDownUnreadyTime:             30 * time.Minute,
				MaxNodeProvisionTime:             60 * time.Minute,
			},
		},
		{
			desc: "complete partial options specs with defaults",
			opts: map[string]string{
				config.DefaultScaleDownGpuUtilizationThresholdKey: "0.1",
				config.DefaultScaleDownUnneededTimeKey:            "1m",
			},
			expected: &config.NodeGroupAutoscalingOptions{
				ScaleDownGpuUtilizationThreshold: 0.1,
				ScaleDownUtilizationThreshold:    defaultOptions.ScaleDownUtilizationThreshold,
				ScaleDownUnneededTime:            time.Minute,
				ScaleDownUnreadyTime:             defaultOptions.ScaleDownUnreadyTime,
				MaxNodeProvisionTime:             15 * time.Minute,
			},
		},
		{
			desc: "keep defaults on unparsable options values",
			opts: map[string]string{
				config.DefaultScaleDownGpuUtilizationThresholdKey: "foo",
				config.DefaultScaleDownUnneededTimeKey:            "bar",
			},
			expected: &defaultOptions,
		},
	}

	test := func(t *testing.T, testConfig *testConfig, expectedOptions *config.NodeGroupAutoscalingOptions) {
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
		opts, err := ng.GetOptions(defaultOptions)
		assert.NoError(t, err)
		assert.Equal(t, expectedOptions, opts)
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			annotations := map[string]string{}
			for k, v := range c.opts {
				annotations[nodeGroupAutoscalingOptionsKeyPrefix+k] = v
			}

			t.Run("MachineSet", func(t *testing.T) {
				test(
					t,
					createMachineSetTestConfig(
						testNamespace,
						RandomString(6),
						RandomString(6),
						10,
						cloudprovider.JoinStringMaps(enableScaleAnnotations, annotations),
						nil,
					),
					c.expected,
				)
			})

			t.Run("MachineDeployment", func(t *testing.T) {
				test(
					t,
					createMachineDeploymentTestConfig(
						testNamespace,
						RandomString(6),
						RandomString(6),
						10,
						cloudprovider.JoinStringMaps(enableScaleAnnotations, annotations),
						nil,
					),
					c.expected,
				)
			})
		})
	}
}
