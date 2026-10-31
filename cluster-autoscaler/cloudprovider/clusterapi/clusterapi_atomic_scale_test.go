/*
Copyright The Kubernetes Authors.

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
	"math"
	"strings"
	"testing"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakescale "k8s.io/client-go/scale/fake"
	clientgotesting "k8s.io/client-go/testing"
)

func TestNodeGroupAtomicIncreaseSize(t *testing.T) {
	annotations := map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	}
	testCases := []struct {
		name  string
		build func() *testConfigBuilder
	}{
		{name: "MachineSet", build: func() *testConfigBuilder { return NewTestConfigBuilder().ForMachineSet() }},
		{name: "MachineDeployment", build: func() *testConfigBuilder { return NewTestConfigBuilder().ForMachineDeployment() }},
		{name: "MachinePool", build: func() *testConfigBuilder { return NewTestConfigBuilder().ForMachinePool() }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			testConfig := tc.build().WithNodeCount(3).WithAnnotations(annotations).Build()
			controller := NewTestMachineController(t)
			defer controller.Stop()
			if err := controller.AddTestConfigs(testConfig); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			nodegroups, err := controller.nodeGroups()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			ng := nodegroups[0].(*nodegroup)

			// When
			err = ng.AtomicIncreaseSize(context.Background(), 2)

			// Then
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gvr, err := ng.scalableResource.GroupVersionResource()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			scale, err := controller.managementScaleClient.Scales(ng.scalableResource.Namespace()).
				Get(context.Background(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if scale.Spec.Replicas != 5 {
				t.Errorf("expected 5 replicas, got %d", scale.Spec.Replicas)
			}
			targetSize, err := ng.TargetSize(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if targetSize != 5 {
				t.Errorf("expected cached target size 5, got %d", targetSize)
			}
		})
	}
}

func TestNodeGroupAtomicIncreaseSizeErrors(t *testing.T) {
	testCases := []struct {
		name     string
		delta    int
		errorMsg string
	}{
		{name: "non-positive delta", delta: 0, errorMsg: "size increase must be positive"},
		{name: "above maximum size", delta: 8, errorMsg: "size increase too large - desired:11 max:10"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			testConfig := NewTestConfigBuilder().ForMachineDeployment().WithNodeCount(3).WithAnnotations(map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			}).Build()
			controller := NewTestMachineController(t)
			defer controller.Stop()
			if err := controller.AddTestConfigs(testConfig); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			nodegroups, err := controller.nodeGroups()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// When
			err = nodegroups[0].AtomicIncreaseSize(context.Background(), tc.delta)

			// Then
			if err == nil || !strings.Contains(err.Error(), tc.errorMsg) {
				t.Fatalf("expected error containing %q, got %v", tc.errorMsg, err)
			}
		})
	}
}

func TestNodeGroupAtomicIncreaseSizeRejectsReplicaOverflow(t *testing.T) {
	// Given
	testConfig := NewTestConfigBuilder().ForMachineDeployment().WithNodeCount(1).WithAnnotations(map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: fmt.Sprintf("%d", math.MaxInt32+1),
	}).Build()
	controller := NewTestMachineController(t)
	defer controller.Stop()
	if err := controller.AddTestConfigs(testConfig); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodegroups, err := controller.nodeGroups()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When
	err = nodegroups[0].AtomicIncreaseSize(context.Background(), math.MaxInt32)

	// Then
	if err == nil || !strings.Contains(err.Error(), "size increase too large - desired:2147483648 max:2147483647") {
		t.Fatalf("expected replica overflow error, got %v", err)
	}
}

func TestNodeGroupAtomicIncreaseSizeAllowsMaxInt32Replicas(t *testing.T) {
	// Given
	testConfig := NewTestConfigBuilder().ForMachineDeployment().WithNodeCount(1).WithAnnotations(map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: fmt.Sprintf("%d", math.MaxInt32),
	}).Build()
	controller := NewTestMachineController(t)
	defer controller.Stop()
	if err := controller.AddTestConfigs(testConfig); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodegroups, err := controller.nodeGroups()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ng := nodegroups[0].(*nodegroup)

	// When
	err = ng.AtomicIncreaseSize(context.Background(), math.MaxInt32-1)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	targetSize, err := ng.TargetSize(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetSize != math.MaxInt32 {
		t.Errorf("expected target size %d, got %d", math.MaxInt32, targetSize)
	}
}

func TestNodeGroupAtomicIncreaseSizePreservesConcurrentUpdate(t *testing.T) {
	// Given
	testConfig := NewTestConfigBuilder().ForMachineDeployment().WithNodeCount(3).WithAnnotations(map[string]string{
		nodeGroupMinSizeAnnotationKey: "1",
		nodeGroupMaxSizeAnnotationKey: "10",
	}).Build()
	testConfig.machineDeployment.SetResourceVersion("1")
	controller := NewTestMachineController(t)
	defer controller.Stop()
	if err := controller.AddTestConfigs(testConfig); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodegroups, err := controller.nodeGroups()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ng := nodegroups[0].(*nodegroup)
	scaleClient := controller.managementScaleClient.(*fakescale.FakeScaleClient)
	scaleClient.PrependReactor("update", "*", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		updateAction := action.(clientgotesting.UpdateAction)
		scale := updateAction.GetObject().(*autoscalingv1.Scale)
		if scale.ResourceVersion != "1" {
			t.Errorf("expected resourceVersion 1, got %q", scale.ResourceVersion)
		}
		gvr, err := ng.scalableResource.GroupVersionResource()
		if err != nil {
			return true, nil, err
		}
		resource, err := controller.dynamicClientset.Resource(gvr).Namespace(ng.scalableResource.Namespace()).Get(context.Background(), ng.scalableResource.Name(), metav1.GetOptions{})
		if err != nil {
			return true, nil, err
		}
		if err := unstructured.SetNestedField(resource.Object, int64(4), "spec", "replicas"); err != nil {
			return true, nil, err
		}
		resource.SetResourceVersion("2")
		if _, err := controller.dynamicClientset.Resource(gvr).Namespace(resource.GetNamespace()).Update(context.Background(), resource, metav1.UpdateOptions{}); err != nil {
			return true, nil, err
		}
		return true, nil, apierrors.NewConflict(schema.GroupResource{Group: defaultCAPIGroup, Resource: resourceNameMachineDeployment}, scale.Name, fmt.Errorf("concurrent scale update"))
	})

	// When
	err = ng.AtomicIncreaseSize(context.Background(), 2)

	// Then
	if !apierrors.IsConflict(err) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	gvr, err := ng.scalableResource.GroupVersionResource()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	scale, err := controller.managementScaleClient.Scales(ng.scalableResource.Namespace()).
		Get(context.Background(), gvr.GroupResource(), ng.scalableResource.Name(), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scale.Spec.Replicas != 4 {
		t.Errorf("expected concurrent update to preserve 4 replicas, got %d", scale.Spec.Replicas)
	}
	targetSize, err := ng.TargetSize(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targetSize != 3 {
		t.Errorf("expected cached target size to remain 3, got %d", targetSize)
	}
}
