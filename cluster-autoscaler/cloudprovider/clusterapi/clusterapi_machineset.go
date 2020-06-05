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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	klog "k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

type machineSetScalableResource struct {
	controller *machineController
	machineSet *MachineSet
	maxSize    int
	minSize    int
}

var _ scalableResource = (*machineSetScalableResource)(nil)

func (r machineSetScalableResource) ID() string {
	return path.Join(r.Namespace(), r.Name())
}

func (r machineSetScalableResource) MaxSize() int {
	return r.maxSize
}

func (r machineSetScalableResource) MinSize() int {
	return r.minSize
}

func (r machineSetScalableResource) Name() string {
	return r.machineSet.Name
}

func (r machineSetScalableResource) Namespace() string {
	return r.machineSet.Namespace
}

func (r machineSetScalableResource) Nodes() ([]string, error) {
	return r.controller.machineSetProviderIDs(r.machineSet)
}

func (r machineSetScalableResource) Replicas() (int32, error) {
	freshMachineSet, err := r.controller.getMachineSet(r.machineSet.Namespace, r.machineSet.Name, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	if freshMachineSet == nil {
		return 0, fmt.Errorf("unknown machineSet %s", r.machineSet.Name)
	}

	if freshMachineSet.Spec.Replicas == nil {
		klog.Warningf("MachineSet %q has nil spec.replicas. This is unsupported behaviour. Falling back to status.replicas.", r.machineSet.Name)
	}

	// If no value for replicas on the MachineSet spec, fallback to the status
	// TODO: Remove this fallback once defaulting is implemented for MachineSet Replicas
	return pointer.Int32PtrDerefOr(freshMachineSet.Spec.Replicas, freshMachineSet.Status.Replicas), nil
}

func (r machineSetScalableResource) SetSize(nreplicas int32) error {
	u, err := r.controller.dynamicclient.Resource(*r.controller.machineSetResource).Namespace(r.machineSet.Namespace).Get(context.TODO(), r.machineSet.Name, metav1.GetOptions{})

	if err != nil {
		return err
	}

	if u == nil {
		return fmt.Errorf("unknown machineSet %s", r.machineSet.Name)
	}

	u = u.DeepCopy()
	if err := unstructured.SetNestedField(u.Object, int64(nreplicas), "spec", "replicas"); err != nil {
		return fmt.Errorf("failed to set replica value: %v", err)
	}

	_, updateErr := r.controller.dynamicclient.Resource(*r.controller.machineSetResource).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})
	return updateErr
}

func (r machineSetScalableResource) MarkMachineForDeletion(machine *Machine) error {
	u, err := r.controller.dynamicclient.Resource(*r.controller.machineResource).Namespace(machine.Namespace).Get(context.TODO(), machine.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	u = u.DeepCopy()

	annotations := u.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[machineDeleteAnnotationKey] = time.Now().String()
	u.SetAnnotations(annotations)

	_, updateErr := r.controller.dynamicclient.Resource(*r.controller.machineResource).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})
	return updateErr
}

func (r machineSetScalableResource) UnmarkMachineForDeletion(machine *Machine) error {
	return unmarkMachineForDeletion(r.controller, machine)
}

func newMachineSetScalableResource(controller *machineController, machineSet *MachineSet) (*machineSetScalableResource, error) {
	minSize, maxSize, err := parseScalingBounds(machineSet.Annotations)
	if err != nil {
		return nil, fmt.Errorf("error validating min/max annotations: %v", err)
	}

	return &machineSetScalableResource{
		controller: controller,
		machineSet: machineSet,
		maxSize:    maxSize,
		minSize:    minSize,
	}, nil
}
