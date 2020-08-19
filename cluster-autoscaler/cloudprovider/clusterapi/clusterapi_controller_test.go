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
	"math/rand"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/informers"
	fakekube "k8s.io/client-go/kubernetes/fake"
	fakescale "k8s.io/client-go/scale/fake"
	clientgotesting "k8s.io/client-go/testing"
)

type testControllerShutdownFunc func()

type testConfig struct {
	spec              *testSpec
	machineDeployment *unstructured.Unstructured
	machineSet        *unstructured.Unstructured
	machines          []*unstructured.Unstructured
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
			machineObjects = append(machineObjects, config.machines[i])
		}

		machineObjects = append(machineObjects, config.machineSet)
		if config.machineDeployment != nil {
			machineObjects = append(machineObjects, config.machineDeployment)
		}
	}

	kubeclientSet := fakekube.NewSimpleClientset(nodeObjects...)
	dynamicClientset := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme(), machineObjects...)
	discoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &clientgotesting.Fake{
			Resources: []*metav1.APIResourceList{
				{
					GroupVersion: fmt.Sprintf("%s/v1beta1", customCAPIGroup),
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
				{
					GroupVersion: fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
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

	scaleClient := &fakescale.FakeScaleClient{Fake: clientgotesting.Fake{}}
	scaleReactor := func(action clientgotesting.Action) (bool, runtime.Object, error) {
		resource := action.GetResource().Resource
		if resource != resourceNameMachineSet && resource != resourceNameMachineDeployment {
			// Do not attempt to react to resources that are not MachineSet or MachineDeployment
			return false, nil, nil
		}

		subresource := action.GetSubresource()
		if subresource != "scale" {
			// Handle a bug in the client-go fakeNamespaceScaleClient, where the action namespace and subresource are
			// switched for update actions
			if action.GetVerb() == "update" && action.GetNamespace() == "scale" {
				subresource = "scale"
			} else {
				// Do not attempt to respond to anything but scale subresource requests
				return false, nil, nil
			}
		}

		gvr := schema.GroupVersionResource{
			Group:    action.GetResource().Group,
			Version:  "v1alpha3",
			Resource: resource,
		}

		switch action.GetVerb() {
		case "get":
			action, ok := action.(clientgotesting.GetAction)
			if !ok {
				return true, nil, fmt.Errorf("failed to convert Action to GetAction: %T", action)
			}

			u, err := dynamicClientset.Resource(gvr).Namespace(action.GetNamespace()).Get(context.TODO(), action.GetName(), metav1.GetOptions{})
			if err != nil {
				return true, nil, err
			}

			replicas, found, err := unstructured.NestedInt64(u.Object, "spec", "replicas")
			if err != nil {
				return true, nil, err
			}

			if !found {
				replicas = 0
			}

			result := &autoscalingv1.Scale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      u.GetName(),
					Namespace: u.GetNamespace(),
				},
				Spec: autoscalingv1.ScaleSpec{
					Replicas: int32(replicas),
				},
			}

			return true, result, nil
		case "update":
			action, ok := action.(clientgotesting.UpdateAction)
			if !ok {
				return true, nil, fmt.Errorf("failed to convert Action to UpdateAction: %T", action)
			}

			s, ok := action.GetObject().(*autoscalingv1.Scale)
			if !ok {
				return true, nil, fmt.Errorf("failed to convert Resource to Scale: %T", s)
			}

			u, err := dynamicClientset.Resource(gvr).Namespace(s.Namespace).Get(context.TODO(), s.Name, metav1.GetOptions{})
			if err != nil {
				return true, nil, fmt.Errorf("failed to fetch underlying %s resource: %s/%s", resource, s.Namespace, s.Name)
			}

			if err := unstructured.SetNestedField(u.Object, int64(s.Spec.Replicas), "spec", "replicas"); err != nil {
				return true, nil, err
			}

			_, err = dynamicClientset.Resource(gvr).Namespace(s.Namespace).Update(context.TODO(), u, metav1.UpdateOptions{})
			if err != nil {
				return true, nil, err
			}

			return true, s, nil
		default:
			return true, nil, fmt.Errorf("unknown verb: %v", action.GetVerb())
		}
	}
	scaleClient.AddReactor("*", "*", scaleReactor)

	controller, err := newMachineController(dynamicClientset, kubeclientSet, discoveryClient, scaleClient)
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

func createMachineSetTestConfig(namespace, namePrefix string, nodeCount int, annotations map[string]string) *testConfig {
	return createTestConfigs(createTestSpecs(namespace, namePrefix, 1, nodeCount, false, annotations)...)[0]
}

func createMachineSetTestConfigs(namespace, namePrefix string, configCount, nodeCount int, annotations map[string]string) []*testConfig {
	return createTestConfigs(createTestSpecs(namespace, namePrefix, configCount, nodeCount, false, annotations)...)
}

func createMachineDeploymentTestConfig(namespace, namePrefix string, nodeCount int, annotations map[string]string) *testConfig {
	return createTestConfigs(createTestSpecs(namespace, namePrefix, 1, nodeCount, true, annotations)...)[0]
}

func createMachineDeploymentTestConfigs(namespace, namePrefix string, configCount, nodeCount int, annotations map[string]string) []*testConfig {
	return createTestConfigs(createTestSpecs(namespace, namePrefix, configCount, nodeCount, true, annotations)...)
}

func createTestSpecs(namespace, namePrefix string, scalableResourceCount, nodeCount int, isMachineDeployment bool, annotations map[string]string) []testSpec {
	var specs []testSpec

	for i := 0; i < scalableResourceCount; i++ {
		specs = append(specs, testSpec{
			annotations:             annotations,
			machineDeploymentName:   fmt.Sprintf("%s-%d", namePrefix, i),
			machineSetName:          fmt.Sprintf("%s-%d", namePrefix, i),
			namespace:               namespace,
			nodeCount:               nodeCount,
			rootIsMachineDeployment: isMachineDeployment,
		})
	}

	return specs
}

func createTestConfigs(specs ...testSpec) []*testConfig {
	result := make([]*testConfig, 0, len(specs))

	for i, spec := range specs {
		config := &testConfig{
			spec:     &specs[i],
			nodes:    make([]*corev1.Node, spec.nodeCount),
			machines: make([]*unstructured.Unstructured, spec.nodeCount),
		}

		machineSetLabels := map[string]string{
			"machineSetName": spec.machineSetName,
		}

		config.machineSet = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1alpha3",
				"metadata": map[string]interface{}{
					"name":      spec.machineSetName,
					"namespace": spec.namespace,
					"uid":       spec.machineSetName,
				},
				"spec": map[string]interface{}{
					"replicas": int64(spec.nodeCount),
				},
				"status": map[string]interface{}{},
			},
		}

		config.machineSet.SetAnnotations(make(map[string]string))

		if !spec.rootIsMachineDeployment {
			config.machineSet.SetAnnotations(spec.annotations)
		} else {
			machineSetLabels["machineDeploymentName"] = spec.machineDeploymentName

			machineDeploymentLabels := map[string]string{
				"machineDeploymentName": spec.machineDeploymentName,
			}

			config.machineDeployment = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       machineDeploymentKind,
					"apiVersion": "cluster.x-k8s.io/v1alpha3",
					"metadata": map[string]interface{}{
						"name":      spec.machineDeploymentName,
						"namespace": spec.namespace,
						"uid":       spec.machineDeploymentName,
					},
					"spec": map[string]interface{}{
						"replicas": int64(spec.nodeCount),
					},
					"status": map[string]interface{}{},
				},
			}
			config.machineDeployment.SetAnnotations(spec.annotations)
			config.machineDeployment.SetLabels(machineDeploymentLabels)
			unstructured.SetNestedStringMap(config.machineDeployment.Object, machineDeploymentLabels, "spec", "selector", "matchLabels")

			ownerRefs := []metav1.OwnerReference{
				{
					Name: config.machineDeployment.GetName(),
					Kind: config.machineDeployment.GetKind(),
					UID:  config.machineDeployment.GetUID(),
				},
			}
			config.machineSet.SetOwnerReferences(ownerRefs)
		}
		config.machineSet.SetLabels(machineSetLabels)
		unstructured.SetNestedStringMap(config.machineSet.Object, machineSetLabels, "spec", "selector", "matchLabels")

		machineOwner := metav1.OwnerReference{
			Name: config.machineSet.GetName(),
			Kind: config.machineSet.GetKind(),
			UID:  config.machineSet.GetUID(),
		}

		for j := 0; j < spec.nodeCount; j++ {
			config.nodes[j], config.machines[j] = makeLinkedNodeAndMachine(j, spec.namespace, machineOwner, machineSetLabels)
		}

		result = append(result, config)
	}

	return result
}

// makeLinkedNodeAndMachine creates a node and machine. The machine
// has its NodeRef set to the new node and the new machine's owner
// reference is set to owner.
func makeLinkedNodeAndMachine(i int, namespace string, owner metav1.OwnerReference, machineLabels map[string]string) (*corev1.Node, *unstructured.Unstructured) {
	node := &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind: "Node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-node-%d", namespace, owner.Name, i),
			Annotations: map[string]string{
				machineAnnotationKey: fmt.Sprintf("%s/%s-%s-machine-%d", namespace, namespace, owner.Name, i),
			},
		},
		Spec: corev1.NodeSpec{
			ProviderID: fmt.Sprintf("test:////%s-%s-nodeid-%d", namespace, owner.Name, i),
		},
	}

	machine := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       machineKind,
			"apiVersion": "cluster.x-k8s.io/v1alpha3",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-%s-machine-%d", namespace, owner.Name, i),
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"providerID": fmt.Sprintf("test:////%s-%s-nodeid-%d", namespace, owner.Name, i),
			},
			"status": map[string]interface{}{
				"nodeRef": map[string]interface{}{
					"kind": node.Kind,
					"name": node.Name,
				},
			},
		},
	}
	machine.SetOwnerReferences([]metav1.OwnerReference{owner})
	machine.SetLabels(machineLabels)

	return node, machine
}

func addTestConfigs(t *testing.T, controller *machineController, testConfigs ...*testConfig) error {
	t.Helper()

	for _, config := range testConfigs {
		if config.machineDeployment != nil {
			if err := createResource(controller.managementClient, controller.machineDeploymentInformer, controller.machineDeploymentResource, config.machineDeployment); err != nil {
				return err
			}
		}
		if err := createResource(controller.managementClient, controller.machineSetInformer, controller.machineSetResource, config.machineSet); err != nil {
			return err
		}

		for i := range config.machines {
			if err := createResource(controller.managementClient, controller.machineInformer, controller.machineResource, config.machines[i]); err != nil {
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

func selectorFromScalableResource(u *unstructured.Unstructured) (labels.Selector, error) {
	unstructuredSelector, found, err := unstructured.NestedMap(u.Object, "spec", "selector")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("expected field spec.selector on scalable resource type")
	}

	labelSelector := &metav1.LabelSelector{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredSelector, labelSelector); err != nil {
		return nil, err
	}

	return metav1.LabelSelectorAsSelector(labelSelector)
}

func createResource(client dynamic.Interface, informer informers.GenericInformer, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error {
	if _, err := client.Resource(gvr).Namespace(resource.GetNamespace()).Create(context.TODO(), resource, metav1.CreateOptions{}); err != nil {
		return err
	}

	return wait.PollImmediateInfinite(time.Microsecond, func() (bool, error) {
		_, err := informer.Lister().ByNamespace(resource.GetNamespace()).Get(resource.GetName())
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		return true, nil
	})
}

func updateResource(client dynamic.Interface, informer informers.GenericInformer, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error {
	updateResult, err := client.Resource(gvr).Namespace(resource.GetNamespace()).Update(context.TODO(), resource, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return wait.PollImmediateInfinite(time.Microsecond, func() (bool, error) {
		result, err := informer.Lister().ByNamespace(resource.GetNamespace()).Get(resource.GetName())
		if err != nil {
			return false, err
		}
		return reflect.DeepEqual(updateResult, result), nil
	})
}

func deleteResource(client dynamic.Interface, informer informers.GenericInformer, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error {
	if err := client.Resource(gvr).Namespace(resource.GetNamespace()).Delete(context.TODO(), resource.GetName(), metav1.DeleteOptions{}); err != nil {
		return err
	}

	return wait.PollImmediateInfinite(time.Microsecond, func() (bool, error) {
		_, err := informer.Lister().ByNamespace(resource.GetNamespace()).Get(resource.GetName())
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
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
			if err := deleteResource(controller.managementClient, controller.machineInformer, controller.machineResource, config.machines[i]); err != nil {
				return err
			}
		}
		if err := deleteResource(controller.managementClient, controller.machineSetInformer, controller.machineSetResource, config.machineSet); err != nil {
			return err
		}
		if config.machineDeployment != nil {
			if err := deleteResource(controller.managementClient, controller.machineDeploymentInformer, controller.machineDeploymentResource, config.machineDeployment); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestControllerFindMachine(t *testing.T) {
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
			testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			})
			if tc.name == "" {
				tc.name = testConfig.machines[0].GetName()
			}
			if tc.namespace == "" {
				tc.namespace = testConfig.machines[0].GetNamespace()
			}
			test(t, tc, testConfig)
		})
	}
}

func TestControllerFindMachineOwner(t *testing.T) {
	testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
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
	if err := deleteResource(controller.managementClient, controller.machineSetInformer, controller.machineSetResource, testConfig.machineSet); err != nil {
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
	testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
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

	if err := updateResource(controller.managementClient, controller.machineInformer, controller.machineResource, machine); err != nil {
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
	testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
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

func TestControllerListMachinesForScalableResource(t *testing.T) {
	test := func(t *testing.T, testConfig1 *testConfig, testConfig2 *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig1)
		defer stop()

		if err := addTestConfigs(t, controller, testConfig2); err != nil {
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
		testConfig1 := createMachineSetTestConfig(namespace, RandomString(6), 5, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})

		// Construct a second set of objects and add the machines,
		// nodes and the additional machineset to the existing set of
		// test objects in the controller. This gives us two
		// machinesets, each with their own machines and linked nodes.
		testConfig2 := createMachineSetTestConfig(namespace, RandomString(6), 5, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})

		test(t, testConfig1, testConfig2)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		namespace := RandomString(6)
		testConfig1 := createMachineDeploymentTestConfig(namespace, RandomString(6), 5, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})

		// Construct a second set of objects and add the machines,
		// nodes, machineset, and the additional machineset to the existing set of
		// test objects in the controller. This gives us two
		// machinedeployments, each with their own machineSet, machines and linked nodes.
		testConfig2 := createMachineDeploymentTestConfig(namespace, RandomString(6), 5, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})

		test(t, testConfig1, testConfig2)
	})
}

func TestControllerLookupNodeGroupForNonExistentNode(t *testing.T) {
	test := func(t *testing.T, testConfig *testConfig) {
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

	t.Run("MachineSet", func(t *testing.T) {
		testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})
		test(t, testConfig)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := createMachineDeploymentTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})
		test(t, testConfig)
	})
}

func TestControllerNodeGroupForNodeWithMissingMachineOwner(t *testing.T) {
	test := func(t *testing.T, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		machine := testConfig.machines[0].DeepCopy()
		machine.SetOwnerReferences([]metav1.OwnerReference{})

		if err := updateResource(controller.managementClient, controller.machineInformer, controller.machineResource, machine); err != nil {
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
		testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})
		test(t, testConfig)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := createMachineDeploymentTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "10",
		})
		test(t, testConfig)
	})
}

func TestControllerNodeGroupForNodeWithMissingSetMachineOwner(t *testing.T) {
	testConfig := createMachineDeploymentTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	machineSet := testConfig.machineSet.DeepCopy()
	machineSet.SetOwnerReferences([]metav1.OwnerReference{})

	if err := updateResource(controller.managementClient, controller.machineSetInformer, controller.machineSetResource, machineSet); err != nil {
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
		testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
			nodeGroupMinSizeAnnotationKey: "1",
			nodeGroupMaxSizeAnnotationKey: "1",
		})
		test(t, testConfig)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := createMachineDeploymentTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
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

	namespace := RandomString(6)

	// Test #1: zero nodegroups
	assertNodegroupLen(t, controller, 0)

	// Test #2: add 5 machineset-based nodegroups
	machineSetConfigs := createMachineSetTestConfigs(namespace, RandomString(6), 5, 1, annotations)
	if err := addTestConfigs(t, controller, machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 5)

	// Test #2: add 2 machinedeployment-based nodegroups
	machineDeploymentConfigs := createMachineDeploymentTestConfigs(namespace, RandomString(6), 2, 1, annotations)
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
	machineSetConfigs = createMachineSetTestConfigs(namespace, RandomString(6), 5, 1, annotations)
	if err := addTestConfigs(t, controller, machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 0)

	// Test #6: machinedeployment with no scaling bounds results in no nodegroups
	machineDeploymentConfigs = createMachineDeploymentTestConfigs(namespace, RandomString(6), 2, 1, annotations)
	if err := addTestConfigs(t, controller, machineDeploymentConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNodegroupLen(t, controller, 0)

	annotations = map[string]string{
		nodeGroupMinSizeAnnotationKey: "-1",
		nodeGroupMaxSizeAnnotationKey: "1",
	}

	// Test #7: machineset with bad scaling bounds results in an error and no nodegroups
	machineSetConfigs = createMachineSetTestConfigs(namespace, RandomString(6), 5, 1, annotations)
	if err := addTestConfigs(t, controller, machineSetConfigs...); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := controller.nodeGroups(); err == nil {
		t.Fatalf("expected an error")
	}

	// Test #8: machinedeployment with bad scaling bounds results in an error and no nodegroups
	machineDeploymentConfigs = createMachineDeploymentTestConfigs(namespace, RandomString(6), 2, 1, annotations)
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
			test(t, tc, createMachineSetTestConfigs(RandomString(6), RandomString(6), tc.nodeGroups, tc.nodesPerGroup, annotations))
		}
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		for _, tc := range testCases {
			test(t, tc, createMachineDeploymentTestConfigs(RandomString(6), RandomString(6), tc.nodeGroups, tc.nodesPerGroup, annotations))
		}
	})
}

func TestControllerFindMachineFromNodeAnnotation(t *testing.T) {
	testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Remove all the provider ID values on all the machines. We
	// want to force findMachineByProviderID() to fallback to
	// searching using the annotation on the node object.
	for _, machine := range testConfig.machines {
		unstructured.RemoveNestedField(machine.Object, "spec", "providerID")

		if err := updateResource(controller.managementClient, controller.machineInformer, controller.machineResource, machine); err != nil {
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
	testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 3, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Remove all linkage between node and machine.
	for i := range testConfig.machines {
		machine := testConfig.machines[i].DeepCopy()

		unstructured.RemoveNestedField(machine.Object, "spec", "providerID")
		unstructured.RemoveNestedField(machine.Object, "status", "nodeRef")

		if err := updateResource(controller.managementClient, controller.machineInformer, controller.machineResource, machine); err != nil {
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
	testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 3, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Remove Status.NodeRef.Name on all the machines. We want to
	// force machineSetNodeNames() to only consider the provider
	// ID for lookups.
	for i := range testConfig.machines {
		machine := testConfig.machines[i].DeepCopy()

		unstructured.RemoveNestedField(machine.Object, "status", "nodeRef")

		if err := updateResource(controller.managementClient, controller.machineInformer, controller.machineResource, machine); err != nil {
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
	testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), 3, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	})

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	// Remove all the provider ID values on all the machines. We
	// want to force machineSetNodeNames() to fallback to
	// searching using Status.NodeRef.Name.
	for i := range testConfig.machines {
		machine := testConfig.machines[i].DeepCopy()

		unstructured.RemoveNestedField(machine.Object, "spec", "providerID")

		if err := updateResource(controller.managementClient, controller.machineInformer, controller.machineResource, machine); err != nil {
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
	testConfig := createMachineDeploymentTestConfig(RandomString(6), RandomString(6), 1, map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "1",
	})
	if err := os.Setenv(CAPIGroupEnvVar, customCAPIGroup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testConfig.machineDeployment.SetAPIVersion(fmt.Sprintf("%s/v1beta1", customCAPIGroup))
	testConfig.machineSet.SetAPIVersion(fmt.Sprintf("%s/v1beta1", customCAPIGroup))

	for _, machine := range testConfig.machines {
		machine.SetAPIVersion(fmt.Sprintf("%s/v1beta1", customCAPIGroup))
	}

	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

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
			Resources: []*metav1.APIResourceList{
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
			Resources: []*metav1.APIResourceList{
				{
					GroupVersion: fmt.Sprintf("%s/v1alpha3", defaultCAPIGroup),
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

const CharSet = "0123456789abcdefghijklmnopqrstuvwxyz"

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomString returns a random alphanumeric string.
func RandomString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = CharSet[rnd.Intn(len(CharSet))]
	}
	return string(result)
}
