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

package clusterapi

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math/rand"
	"reflect"
	"testing"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/informers"
	fakekube "k8s.io/client-go/kubernetes/fake"
	fakescale "k8s.io/client-go/scale/fake"
	clientgotesting "k8s.io/client-go/testing"
	klog "k8s.io/klog/v2"
)

type scalableTestType string

const (
	machineSetType        scalableTestType = "MachineSet"
	machineDeploymentType scalableTestType = "MachineDeployment"
)

type testConfigBuilder struct {
	scalableType scalableTestType
	namespace    string
	clusterName  string
	namePrefix   string
	nodeCount    int
	annotations  map[string]string
	capacity     map[string]string
}

// NewTestConfigBuilder returns a builder for dynamically constructing mock ClusterAPI resources for testing.
func NewTestConfigBuilder() *testConfigBuilder {
	return &testConfigBuilder{
		namespace:   RandomString(6),
		clusterName: RandomString(6),
		namePrefix:  RandomString(6),
		nodeCount:   0,
		annotations: nil,
		capacity:    nil,
	}
}

func (b *testConfigBuilder) Build() *TestConfig {
	if len(b.scalableType) == 0 {
		panic("scalable API type must be specified")
	}
	isMachineDeployment := b.scalableType == machineDeploymentType
	configCount := 1

	return createTestConfigs(
		createTestSpecs(
			b.namespace,
			b.clusterName,
			b.namePrefix,
			configCount,
			b.nodeCount,
			isMachineDeployment,
			b.annotations,
			b.capacity,
		)[0],
	)[0]
}

func (b *testConfigBuilder) BuildMultiple(configCount int) []*TestConfig {
	if len(b.scalableType) == 0 {
		panic("scalable API type must be specified")
	}
	isMachineDeployment := b.scalableType == machineDeploymentType

	return createTestConfigs(
		createTestSpecs(
			b.namespace,
			b.clusterName,
			b.namePrefix,
			configCount,
			b.nodeCount,
			isMachineDeployment,
			b.annotations,
			b.capacity,
		)...,
	)
}

func (b *testConfigBuilder) ForMachineSet() *testConfigBuilder {
	b.scalableType = machineSetType
	return b
}

func (b *testConfigBuilder) ForMachineDeployment() *testConfigBuilder {
	b.scalableType = machineDeploymentType
	return b
}

func (b *testConfigBuilder) WithNamespace(n string) *testConfigBuilder {
	b.namespace = n
	return b
}

func (b *testConfigBuilder) WithClusterName(n string) *testConfigBuilder {
	b.clusterName = n
	return b
}

func (b *testConfigBuilder) WithNamePrefix(n string) *testConfigBuilder {
	b.namePrefix = n
	return b
}

func (b *testConfigBuilder) WithNodeCount(c int) *testConfigBuilder {
	b.nodeCount = c
	return b
}

func (b *testConfigBuilder) WithAnnotations(a map[string]string) *testConfigBuilder {
	if a == nil {
		// explicitly setting annotations to nil
		b.annotations = nil
	} else {
		if b.annotations == nil {
			b.annotations = map[string]string{}
		}
		maps.Insert(b.annotations, maps.All(a))
	}
	return b
}

func (b *testConfigBuilder) WithCapacity(c map[string]string) *testConfigBuilder {
	if c == nil {
		// explicitly setting capacity to nil
		b.capacity = nil
	} else {
		if b.capacity == nil {
			b.capacity = map[string]string{}
		}
		maps.Insert(b.capacity, maps.All(c))
	}
	return b
}

// TestConfig contains clusterspecific information about a single test configuration.
type TestConfig struct {
	spec              *TestSpec
	clusterName       string
	namespace         string
	machineDeployment *unstructured.Unstructured
	machineSet        *unstructured.Unstructured
	machineTemplate   *unstructured.Unstructured
	machinePool       *unstructured.Unstructured
	machines          []*unstructured.Unstructured
	nodes             []*corev1.Node
}

func createTestConfigs(specs ...TestSpec) []*TestConfig {
	result := make([]*TestConfig, 0, len(specs))

	for i, spec := range specs {
		config := &TestConfig{
			spec:        &specs[i],
			namespace:   spec.namespace,
			clusterName: spec.clusterName,
			nodes:       make([]*corev1.Node, spec.nodeCount),
			machines:    make([]*unstructured.Unstructured, spec.nodeCount),
		}

		machineSetLabels := map[string]string{
			"clusterName":    spec.clusterName,
			"machineSetName": spec.machineSetName,
		}

		config.machineSet = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       machineSetKind,
				"apiVersion": "cluster.x-k8s.io/v1beta2",
				"metadata": map[string]interface{}{
					"name":      spec.machineSetName,
					"namespace": spec.namespace,
					"uid":       spec.machineSetName,
				},
				"spec": map[string]interface{}{
					"clusterName": spec.clusterName,
					"replicas":    int64(spec.nodeCount),
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"infrastructureRef": map[string]interface{}{
								"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta1",
								"kind":       machineTemplateKind,
								"name":       "TestMachineTemplate",
							},
						},
					},
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
				"clusterName":           spec.clusterName,
				"machineDeploymentName": spec.machineDeploymentName,
			}

			config.machineDeployment = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       machineDeploymentKind,
					"apiVersion": "cluster.x-k8s.io/v1beta2",
					"metadata": map[string]interface{}{
						"name":      spec.machineDeploymentName,
						"namespace": spec.namespace,
						"uid":       spec.machineDeploymentName,
					},
					"spec": map[string]interface{}{
						"clusterName": spec.clusterName,
						"replicas":    int64(spec.nodeCount),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"infrastructureRef": map[string]interface{}{
									"apiGroup": "infrastructure.cluster.x-k8s.io",
									"kind":     machineTemplateKind,
									"name":     "TestMachineTemplate",
								},
							},
						},
					},
					"status": map[string]interface{}{},
				},
			}
			config.machineDeployment.SetAnnotations(spec.annotations)
			config.machineDeployment.SetLabels(machineDeploymentLabels)
			if err := unstructured.SetNestedStringMap(config.machineDeployment.Object, machineDeploymentLabels, "spec", "selector", "matchLabels"); err != nil {
				panic(err)
			}

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
		if err := unstructured.SetNestedStringMap(config.machineSet.Object, machineSetLabels, "spec", "selector", "matchLabels"); err != nil {
			panic(err)
		}

		machineOwner := metav1.OwnerReference{
			Name: config.machineSet.GetName(),
			Kind: config.machineSet.GetKind(),
			UID:  config.machineSet.GetUID(),
		}

		if spec.capacity != nil {
			klog.V(4).Infof("adding capacity to machine template")
			config.machineTemplate = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta1",
					"kind":       machineTemplateKind,
					"metadata": map[string]interface{}{
						"name":      "TestMachineTemplate",
						"namespace": spec.namespace,
						"uid":       "TestMachineTemplate",
					},
				},
			}
			if err := unstructured.SetNestedStringMap(config.machineTemplate.Object, spec.capacity, "status", "capacity"); err != nil {
				panic(err)
			}
		} else {
			klog.V(4).Infof("not adding capacity")
		}

		for j := 0; j < spec.nodeCount; j++ {
			config.nodes[j], config.machines[j] = makeLinkedNodeAndMachine(j, spec.namespace, spec.clusterName, machineOwner, machineSetLabels)
		}

		result = append(result, config)
	}

	return result
}

// TestSpec contains the clusterapi specific information for a single test specification.
type TestSpec struct {
	annotations             map[string]string
	capacity                map[string]string
	machineDeploymentName   string
	machineSetName          string
	machinePoolName         string
	clusterName             string
	namespace               string
	nodeCount               int
	rootIsMachineDeployment bool
}

func createTestSpecs(namespace, clusterName, namePrefix string, scalableResourceCount, nodeCount int, isMachineDeployment bool, annotations map[string]string, capacity map[string]string) []TestSpec {
	var specs []TestSpec

	for i := 0; i < scalableResourceCount; i++ {
		specs = append(specs, createTestSpec(namespace, clusterName, fmt.Sprintf("%s-%d", namePrefix, i), nodeCount, isMachineDeployment, annotations, capacity))
	}

	return specs
}

func createTestSpec(namespace, clusterName, name string, nodeCount int, isMachineDeployment bool, annotations map[string]string, capacity map[string]string) TestSpec {
	return TestSpec{
		annotations:             annotations,
		capacity:                capacity,
		machineDeploymentName:   name,
		machineSetName:          name,
		clusterName:             clusterName,
		namespace:               namespace,
		nodeCount:               nodeCount,
		rootIsMachineDeployment: isMachineDeployment,
	}
}

// Part of RandomString utility function
const charSet = "0123456789abcdefghijklmnopqrstuvwxyz"

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomString returns a random alphanumeric string.
func RandomString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = charSet[rnd.Intn(len(charSet))]
	}
	return string(result)
}

// makeLinkedNodeAndMachine creates a node and machine. The machine
// has its NodeRef set to the new node and the new machine's owner
// reference is set to owner.
func makeLinkedNodeAndMachine(i int, namespace, clusterName string, owner metav1.OwnerReference, machineLabels map[string]string) (*corev1.Node, *unstructured.Unstructured) {
	node := &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind: "Node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-node-%d", namespace, owner.Name, i),
			Annotations: map[string]string{
				clusterNameAnnotationKey:      clusterName,
				clusterNamespaceAnnotationKey: namespace,
				machineAnnotationKey:          fmt.Sprintf("%s-%s-machine-%d", namespace, owner.Name, i),
			},
		},
		Spec: corev1.NodeSpec{
			ProviderID: fmt.Sprintf("test:////%s-%s-nodeid-%d", namespace, owner.Name, i),
		},
	}

	machine := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       machineKind,
			"apiVersion": "cluster.x-k8s.io/v1beta2",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-%s-machine-%d", namespace, owner.Name, i),
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"clusterName": clusterName,
				"providerID":  fmt.Sprintf("test:////%s-%s-nodeid-%d", namespace, owner.Name, i),
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

type testControllerShutdownFunc func()

const customCAPIGroup = "custom.x-k8s.io"
const fifteenSecondDuration = time.Second * 15

type testMachineController struct {
	*machineController

	stopCh           chan struct{}
	testingTb        testing.TB
	dynamicClientset *fakedynamic.FakeDynamicClient
}

// NewTestMachineController returns a new machineController wrapped by a test harness and associated with a testing interface.
func NewTestMachineController(t testing.TB) *testMachineController {
	t.Helper()

	kubeclientSet := fakekube.NewSimpleClientset()
	dynamicClientset := fakedynamic.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: "cluster.x-k8s.io", Version: "v1beta1", Resource: "machinedeployments"}:              "kindList",
			{Group: "cluster.x-k8s.io", Version: "v1beta1", Resource: "machines"}:                        "kindList",
			{Group: "cluster.x-k8s.io", Version: "v1beta1", Resource: "machinesets"}:                     "kindList",
			{Group: "cluster.x-k8s.io", Version: "v1beta2", Resource: "machinedeployments"}:              "kindList",
			{Group: "cluster.x-k8s.io", Version: "v1beta2", Resource: "machines"}:                        "kindList",
			{Group: "cluster.x-k8s.io", Version: "v1beta2", Resource: "machinesets"}:                     "kindList",
			{Group: "cluster.x-k8s.io", Version: "v1beta2", Resource: "machinepools"}:                    "kindList",
			{Group: "custom.x-k8s.io", Version: "v1beta1", Resource: "machinepools"}:                     "kindList",
			{Group: "custom.x-k8s.io", Version: "v1beta1", Resource: "machinedeployments"}:               "kindList",
			{Group: "custom.x-k8s.io", Version: "v1beta1", Resource: "machines"}:                         "kindList",
			{Group: "custom.x-k8s.io", Version: "v1beta1", Resource: "machinesets"}:                      "kindList",
			{Group: "infrastructure.cluster.x-k8s.io", Version: "v1beta1", Resource: "machinetemplates"}: "kindList",
		},
	)
	discoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &clientgotesting.Fake{
			Resources: []*metav1.APIResourceList{
				{
					GroupVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
					APIResources: []metav1.APIResource{
						{
							Name: "machinetemplates",
						},
					},
				},
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
						{
							Name: resourceNameMachinePool,
						},
					},
				},
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
						{
							Name: resourceNameMachinePool,
						},
					},
				},
			},
		},
	}

	scaleClient := &fakescale.FakeScaleClient{Fake: clientgotesting.Fake{}}
	scaleReactor := func(action clientgotesting.Action) (bool, runtime.Object, error) {
		resource := action.GetResource().Resource
		if resource != resourceNameMachineSet && resource != resourceNameMachineDeployment && resource != resourceNameMachinePool {
			// Do not attempt to react to resources that are not MachineSet, MachineDeployment, or MachinePool
			return false, nil, nil
		}

		subresource := action.GetSubresource()
		if subresource != "scale" {
			// Handle a bug in the client-go fakeNamespaceScaleClient, where the action namespace and subresource are
			// switched for update actions
			if action.GetVerb() != "update" || action.GetNamespace() != "scale" {
				// Do not attempt to respond to anything but scale subresource requests
				return false, nil, nil
			}
		}

		gvr := schema.GroupVersionResource{
			Group:    action.GetResource().Group,
			Version:  "v1beta2",
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

			replicas, found, err := unstructured.NestedInt64(u.UnstructuredContent(), "spec", "replicas")
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
		case "patch":
			action, ok := action.(clientgotesting.PatchAction)
			if !ok {
				return true, nil, fmt.Errorf("failed to convert Action to PatchAction: %T", action)
			}

			pt := action.GetPatchType()
			if pt != types.MergePatchType {
				return true, nil, fmt.Errorf("unexpected patch type: expected = %s, got = %s", types.MergePatchType, pt)
			}

			var scale autoscalingv1.Scale
			err := json.Unmarshal(action.GetPatch(), &scale)
			if err != nil {
				return true, nil, fmt.Errorf("couldn't unmarshal patch: %w", err)
			}

			_, err = dynamicClientset.Resource(gvr).Namespace(action.GetNamespace()).Patch(context.TODO(), action.GetName(), pt, action.GetPatch(), metav1.PatchOptions{})
			if err != nil {
				return true, nil, err
			}

			newReplicas := scale.Spec.Replicas

			return true, &autoscalingv1.Scale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      action.GetName(),
					Namespace: action.GetNamespace(),
				},
				Spec: autoscalingv1.ScaleSpec{
					Replicas: newReplicas,
				},
			}, nil
		default:
			return true, nil, fmt.Errorf("unknown verb: %v", action.GetVerb())
		}
	}
	scaleClient.AddReactor("*", "*", scaleReactor)

	stopCh := make(chan struct{})
	controller, err := newMachineController(dynamicClientset, kubeclientSet, discoveryClient, scaleClient, cloudprovider.NodeGroupDiscoveryOptions{}, stopCh)
	if err != nil {
		t.Fatal("failed to create test controller")
	}

	if err := controller.run(); err != nil {
		t.Fatalf("failed to run controller: %v", err)
	}

	return &testMachineController{
		machineController: controller,
		stopCh:            stopCh,
		testingTb:         t,
		dynamicClientset:  dynamicClientset,
	}
}

func (c *testMachineController) Stop() {
	close(c.stopCh)
}

func (c *testMachineController) AddTestConfigs(testConfigs ...*TestConfig) error {
	c.testingTb.Helper()

	for _, config := range testConfigs {
		if config.machineDeployment != nil {
			if err := c.CreateResource(c.machineDeploymentInformer, c.machineDeploymentResource, config.machineDeployment); err != nil {
				return err
			}
		}
		if err := c.CreateResource(c.machineSetInformer, c.machineSetResource, config.machineSet); err != nil {
			return err
		}

		if config.machinePool != nil {
			if err := c.CreateResource(c.machinePoolInformer, c.machinePoolResource, config.machinePool); err != nil {
				return err
			}
		}

		if config.machineTemplate != nil {
			c.dynamicClientset.Tracker().Add(config.machineTemplate)
		}

		for i := range config.machines {
			if err := c.CreateResource(c.machineInformer, c.machineResource, config.machines[i]); err != nil {
				return err
			}
		}

		for i := range config.nodes {
			if err := c.nodeInformer.GetStore().Add(config.nodes[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *testMachineController) DeleteTestConfigs(testConfigs ...*TestConfig) error {
	c.testingTb.Helper()

	for _, config := range testConfigs {
		for i := range config.nodes {
			if err := c.nodeInformer.GetStore().Delete(config.nodes[i]); err != nil {
				return err
			}
		}
		for i := range config.machines {
			if err := c.DeleteResource(c.machineInformer, c.machineResource, config.machines[i]); err != nil {
				return err
			}
		}
		if err := c.DeleteResource(c.machineSetInformer, c.machineSetResource, config.machineSet); err != nil {
			return err
		}
		if config.machineDeployment != nil {
			if err := c.DeleteResource(c.machineDeploymentInformer, c.machineDeploymentResource, config.machineDeployment); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *testMachineController) CreateResource(informer informers.GenericInformer, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error {
	if _, err := c.managementClient.Resource(gvr).Namespace(resource.GetNamespace()).Create(context.TODO(), resource, metav1.CreateOptions{}); err != nil {
		return err
	}

	deadlineCtx, deadlineFn := context.WithTimeout(context.Background(), fifteenSecondDuration)
	defer deadlineFn()
	return wait.PollUntilContextTimeout(deadlineCtx, time.Microsecond, fifteenSecondDuration, true, func(_ context.Context) (bool, error) {
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

func (c *testMachineController) UpdateResource(informer informers.GenericInformer, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error {
	updateResult, err := c.managementClient.Resource(gvr).Namespace(resource.GetNamespace()).Update(context.TODO(), resource, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	deadlineCtx, deadlineFn := context.WithTimeout(context.Background(), fifteenSecondDuration)
	defer deadlineFn()
	return wait.PollUntilContextTimeout(deadlineCtx, time.Microsecond, fifteenSecondDuration, true, func(_ context.Context) (bool, error) {
		result, err := informer.Lister().ByNamespace(resource.GetNamespace()).Get(resource.GetName())
		if err != nil {
			return false, err
		}
		return reflect.DeepEqual(updateResult, result), nil
	})
}

func (c *testMachineController) DeleteResource(informer informers.GenericInformer, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error {
	if err := c.managementClient.Resource(gvr).Namespace(resource.GetNamespace()).Delete(context.TODO(), resource.GetName(), metav1.DeleteOptions{}); err != nil {
		return err
	}

	deadlineCtx, deadlineFn := context.WithTimeout(context.Background(), fifteenSecondDuration)
	defer deadlineFn()
	return wait.PollUntilContextTimeout(deadlineCtx, time.Microsecond, fifteenSecondDuration, true, func(_ context.Context) (bool, error) {
		_, err := informer.Lister().ByNamespace(resource.GetNamespace()).Get(resource.GetName())
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
}
