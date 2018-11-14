/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/google/uuid"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"strconv"
)

func buildTestMachineDeployment(name string, replicas, minSize, maxSize int) *v1alpha1.MachineDeployment {
	md := &v1alpha1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
			UID:       types.UID(uuid.New().String()),
			SelfLink:  fmt.Sprintf("/apis/cluster.k8s.io/v1alpha1/namespaces/kube-system/machinedeployments/%s", name),
			Labels:    map[string]string{},
		},
		Spec: v1alpha1.MachineDeploymentSpec{
			Replicas: int32Ptr(int32(replicas)),
		},
	}

	if minSize < maxSize {
		md.ObjectMeta.Annotations = map[string]string{
			MinSizeAnnotation: strconv.Itoa(minSize),
			MaxSizeAnnotation: strconv.Itoa(maxSize),
		}
	}

	return md
}

func buildTestMachineSet(owner *v1alpha1.MachineDeployment, name string, replicas int) *v1alpha1.MachineSet {
	ms := &v1alpha1.MachineSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
			UID:       types.UID(uuid.New().String()),
			SelfLink:  fmt.Sprintf("/apis/cluster.k8s.io/v1alpha1/namespaces/kube-system/machinesets/%s", name),
			Labels:    map[string]string{},
		},
		Spec: v1alpha1.MachineSetSpec{
			Replicas: int32Ptr(int32(replicas)),
		},
	}

	if nil != owner {
		ms.ObjectMeta.OwnerReferences = test.GenerateOwnerReferences(owner.Name, "MachineDeployment", "cluster.k8s.io/v1alpha1", owner.UID)
	}

	return ms
}

func buildTestMachine(owner *v1alpha1.MachineSet, name string, node *apiv1.Node) *v1alpha1.Machine {
	m := &v1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
			UID:       types.UID(uuid.New().String()),
			SelfLink:  fmt.Sprintf("/apis/cluster.k8s.io/v1alpha1/namespaces/kube-system/machines/%s", name),
			Labels:    map[string]string{},
		},
		Spec: v1alpha1.MachineSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
		Status: v1alpha1.MachineStatus{
			ProviderStatus: nil,
		},
	}

	if nil != owner {
		m.ObjectMeta.OwnerReferences = test.GenerateOwnerReferences(owner.Name, "MachineSet", "cluster.k8s.io/v1alpha1", owner.UID)
	}

	if nil != node {
		m.Status.NodeRef = &apiv1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node.Name,
			UID:        node.UID,
		}
	}

	return m
}

func buildTestNode(name string) *apiv1.Node {
	n := test.BuildTestNode(name, 100, 100)
	n.UID = types.UID(uuid.New().String())

	return n
}
