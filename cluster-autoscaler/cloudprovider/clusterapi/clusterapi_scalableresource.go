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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// scalableResource is a resource that can be scaled up and down by
// adjusting its replica count field.
type scalableResource interface {
	// Id returns an unique identifier of the resource
	ID() string

	// MaxSize returns maximum size of the resource
	MaxSize() int

	// MinSize returns minimum size of the resource
	MinSize() int

	// Name returns the name of the resource
	Name() string

	// Namespace returns the namespace the resource is in
	Namespace() string

	// Nodes returns a list of all machines that already have or should become nodes that belong to this
	// resource
	Nodes() ([]string, error)

	// SetSize() sets the replica count of the resource
	SetSize(nreplicas int32) error

	// Replicas returns the current replica count of the resource
	Replicas() (int32, error)

	// MarkMachineForDeletion marks machine for deletion
	MarkMachineForDeletion(machine *Machine) error

	// UnmarkMachineForDeletion unmarks machine for deletion
	UnmarkMachineForDeletion(machine *Machine) error
}

func unmarkMachineForDeletion(controller *machineController, machine *Machine) error {
	u, err := controller.dynamicclient.Resource(*controller.machineResource).Namespace(machine.Namespace).Get(context.TODO(), machine.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	annotations := u.GetAnnotations()
	if _, ok := annotations[machineDeleteAnnotationKey]; ok {
		delete(annotations, machineDeleteAnnotationKey)
		u.SetAnnotations(annotations)
		_, updateErr := controller.dynamicclient.Resource(*controller.machineResource).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})
		return updateErr
	}
	return nil
}
