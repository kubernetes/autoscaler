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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSetSize(t *testing.T) {
	initialReplicas := int32(1)
	updatedReplicas := int32(5)

	testConfig := createMachineSetTestConfig(testNamespace, int(initialReplicas), nil)
	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	sr, err := newMachineSetScalableResource(controller, testConfig.machineSet)
	if err != nil {
		t.Fatal(err)
	}

	err = sr.SetSize(updatedReplicas)
	if err != nil {
		t.Fatal(err)
	}

	// fetch machineSet
	u, err := sr.controller.dynamicclient.Resource(*sr.controller.machineSetResource).Namespace(testConfig.machineSet.Namespace).
		Get(context.TODO(), testConfig.machineSet.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	replicas, found, err := unstructured.NestedInt64(u.Object, "spec", "replicas")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("spec.replicas not found")
	}

	got := int32(replicas)
	if got != updatedReplicas {
		t.Errorf("expected %v, got: %v", updatedReplicas, got)
	}
}

func TestReplicas(t *testing.T) {
	initialReplicas := int32(1)
	updatedReplicas := int32(5)

	testConfig := createMachineSetTestConfig(testNamespace, int(initialReplicas), nil)
	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	sr, err := newMachineSetScalableResource(controller, testConfig.machineSet)
	if err != nil {
		t.Fatal(err)
	}

	i, err := sr.Replicas()
	if err != nil {
		t.Fatal(err)
	}

	if i != initialReplicas {
		t.Errorf("expected %v, got: %v", initialReplicas, i)
	}

	// fetch and update machineSet
	u, err := sr.controller.dynamicclient.Resource(*sr.controller.machineSetResource).Namespace(testConfig.machineSet.Namespace).
		Get(context.TODO(), testConfig.machineSet.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if err := unstructured.SetNestedField(u.Object, int64(updatedReplicas), "spec", "replicas"); err != nil {
		t.Fatal(err)
	}

	_, err = sr.controller.dynamicclient.Resource(*sr.controller.machineSetResource).Namespace(u.GetNamespace()).
		Update(context.TODO(), u, metav1.UpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	i, err = sr.Replicas()
	if err != nil {
		t.Fatal(err)
	}

	if i != updatedReplicas {
		t.Errorf("expected %v, got: %v", updatedReplicas, i)
	}
}

func TestSetSizeAndReplicas(t *testing.T) {
	initialReplicas := int32(1)
	updatedReplicas := int32(5)

	testConfig := createMachineSetTestConfig(testNamespace, int(initialReplicas), nil)
	controller, stop := mustCreateTestController(t, testConfig)
	defer stop()

	sr, err := newMachineSetScalableResource(controller, testConfig.machineSet)
	if err != nil {
		t.Fatal(err)
	}

	i, err := sr.Replicas()
	if err != nil {
		t.Fatal(err)
	}

	if i != initialReplicas {
		t.Errorf("expected %v, got: %v", initialReplicas, i)
	}

	err = sr.SetSize(updatedReplicas)
	if err != nil {
		t.Fatal(err)
	}

	i, err = sr.Replicas()
	if err != nil {
		t.Fatal(err)
	}

	if i != updatedReplicas {
		t.Errorf("expected %v, got: %v", updatedReplicas, i)
	}
}
