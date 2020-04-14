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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"
)

func newMachineDeploymentFromUnstructured(u *unstructured.Unstructured) *MachineDeployment {
	machineDeployment := MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       u.GetKind(),
			APIVersion: u.GetAPIVersion(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              u.GetName(),
			Namespace:         u.GetNamespace(),
			UID:               u.GetUID(),
			Labels:            u.GetLabels(),
			Annotations:       u.GetAnnotations(),
			OwnerReferences:   u.GetOwnerReferences(),
			DeletionTimestamp: u.GetDeletionTimestamp(),
		},
		Spec:   MachineDeploymentSpec{},
		Status: MachineDeploymentStatus{},
	}

	replicas, found, err := unstructured.NestedInt64(u.Object, "spec", "replicas")
	if err == nil && found {
		machineDeployment.Spec.Replicas = pointer.Int32Ptr(int32(replicas))
	}

	return &machineDeployment
}

func newMachineSetFromUnstructured(u *unstructured.Unstructured) *MachineSet {
	machineSet := MachineSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       u.GetKind(),
			APIVersion: u.GetAPIVersion(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              u.GetName(),
			Namespace:         u.GetNamespace(),
			UID:               u.GetUID(),
			Labels:            u.GetLabels(),
			Annotations:       u.GetAnnotations(),
			OwnerReferences:   u.GetOwnerReferences(),
			DeletionTimestamp: u.GetDeletionTimestamp(),
		},
		Spec:   MachineSetSpec{},
		Status: MachineSetStatus{},
	}

	replicas, found, err := unstructured.NestedInt64(u.Object, "spec", "replicas")
	if err == nil && found {
		machineSet.Spec.Replicas = pointer.Int32Ptr(int32(replicas))
	}

	return &machineSet
}

func newMachineFromUnstructured(u *unstructured.Unstructured) *Machine {
	machine := Machine{
		TypeMeta: metav1.TypeMeta{
			Kind:       u.GetKind(),
			APIVersion: u.GetAPIVersion(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              u.GetName(),
			Namespace:         u.GetNamespace(),
			UID:               u.GetUID(),
			Labels:            u.GetLabels(),
			Annotations:       u.GetAnnotations(),
			OwnerReferences:   u.GetOwnerReferences(),
			ClusterName:       u.GetClusterName(),
			DeletionTimestamp: u.GetDeletionTimestamp(),
		},
		Spec:   MachineSpec{},
		Status: MachineStatus{},
	}

	if providerID, _, _ := unstructured.NestedString(u.Object, "spec", "providerID"); providerID != "" {
		machine.Spec.ProviderID = pointer.StringPtr(providerID)
	}

	nodeRef := corev1.ObjectReference{}

	if nodeRefKind, _, _ := unstructured.NestedString(u.Object, "status", "nodeRef", "kind"); nodeRefKind != "" {
		nodeRef.Kind = nodeRefKind
	}

	if nodeRefName, _, _ := unstructured.NestedString(u.Object, "status", "nodeRef", "name"); nodeRefName != "" {
		nodeRef.Name = nodeRefName
	}

	if nodeRef.Name != "" || nodeRef.Kind != "" {
		machine.Status.NodeRef = &nodeRef
	}

	if failureMessage, _, _ := unstructured.NestedString(u.Object, "status", "failureMessage"); failureMessage != "" {
		machine.Status.FailureMessage = pointer.StringPtr(failureMessage)
	}

	return &machine
}

func newUnstructuredFromMachineSet(m *MachineSet) *unstructured.Unstructured {
	u := unstructured.Unstructured{}

	u.SetAPIVersion(m.APIVersion)
	u.SetAnnotations(m.Annotations)
	u.SetKind(m.Kind)
	u.SetLabels(m.Labels)
	u.SetName(m.Name)
	u.SetNamespace(m.Namespace)
	u.SetOwnerReferences(m.OwnerReferences)
	u.SetUID(m.UID)
	u.SetDeletionTimestamp(m.DeletionTimestamp)

	if m.Spec.Replicas != nil {
		unstructured.SetNestedField(u.Object, int64(*m.Spec.Replicas), "spec", "replicas")
	}

	return &u
}

func newUnstructuredFromMachineDeployment(m *MachineDeployment) *unstructured.Unstructured {
	u := unstructured.Unstructured{}

	u.SetAPIVersion(m.APIVersion)
	u.SetAnnotations(m.Annotations)
	u.SetKind(m.Kind)
	u.SetLabels(m.Labels)
	u.SetName(m.Name)
	u.SetNamespace(m.Namespace)
	u.SetOwnerReferences(m.OwnerReferences)
	u.SetUID(m.UID)
	u.SetDeletionTimestamp(m.DeletionTimestamp)

	if m.Spec.Replicas != nil {
		unstructured.SetNestedField(u.Object, int64(*m.Spec.Replicas), "spec", "replicas")
	}

	return &u
}

func newUnstructuredFromMachine(m *Machine) *unstructured.Unstructured {
	u := unstructured.Unstructured{}

	u.SetAPIVersion(m.APIVersion)
	u.SetAnnotations(m.Annotations)
	u.SetKind(m.Kind)
	u.SetLabels(m.Labels)
	u.SetName(m.Name)
	u.SetNamespace(m.Namespace)
	u.SetOwnerReferences(m.OwnerReferences)
	u.SetUID(m.UID)
	u.SetDeletionTimestamp(m.DeletionTimestamp)

	if m.Spec.ProviderID != nil && *m.Spec.ProviderID != "" {
		unstructured.SetNestedField(u.Object, *m.Spec.ProviderID, "spec", "providerID")
	}

	if m.Status.NodeRef != nil {
		if m.Status.NodeRef.Kind != "" {
			unstructured.SetNestedField(u.Object, m.Status.NodeRef.Kind, "status", "nodeRef", "kind")
		}
		if m.Status.NodeRef.Name != "" {
			unstructured.SetNestedField(u.Object, m.Status.NodeRef.Name, "status", "nodeRef", "name")
		}
	}

	if m.Status.FailureMessage != nil {
		unstructured.SetNestedField(u.Object, *m.Status.FailureMessage, "status", "failureMessage")
	}

	return &u
}
