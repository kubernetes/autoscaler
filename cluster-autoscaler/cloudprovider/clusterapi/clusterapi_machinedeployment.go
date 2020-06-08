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

type machineDeploymentScalableResource struct {
	controller        *machineController
	machineDeployment *MachineDeployment
	maxSize           int
	minSize           int
}

var _ scalableResource = (*machineDeploymentScalableResource)(nil)

func (r machineDeploymentScalableResource) ID() string {
	return path.Join(r.Namespace(), r.Name())
}

func (r machineDeploymentScalableResource) MaxSize() int {
	return r.maxSize
}

func (r machineDeploymentScalableResource) MinSize() int {
	return r.minSize
}

func (r machineDeploymentScalableResource) Name() string {
	return r.machineDeployment.Name
}

func (r machineDeploymentScalableResource) Namespace() string {
	return r.machineDeployment.Namespace
}

func (r machineDeploymentScalableResource) Nodes() ([]string, error) {
	var result []string

	if err := r.controller.filterAllMachineSets(func(machineSet *MachineSet) error {
		if machineSetIsOwnedByMachineDeployment(machineSet, r.machineDeployment) {
			providerIDs, err := r.controller.machineSetProviderIDs(machineSet)
			if err != nil {
				return err
			}
			result = append(result, providerIDs...)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (r machineDeploymentScalableResource) Replicas() (int32, error) {
	freshMachineDeployment, err := r.controller.getMachineDeployment(r.machineDeployment.Namespace, r.machineDeployment.Name, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	if freshMachineDeployment == nil {
		return 0, fmt.Errorf("unknown machineDeployment %s", r.machineDeployment.Name)
	}

	if freshMachineDeployment.Spec.Replicas == nil {
		klog.Warningf("MachineDeployment %q has nil spec.replicas. This is unsupported behaviour. Falling back to status.replicas.", r.machineDeployment.Name)
	}
	// If no value for replicas on the MachineSet spec, fallback to the status
	// TODO: Remove this fallback once defaulting is implemented for MachineSet Replicas
	return pointer.Int32PtrDerefOr(freshMachineDeployment.Spec.Replicas, freshMachineDeployment.Status.Replicas), nil
}

func (r machineDeploymentScalableResource) SetSize(nreplicas int32) error {
	u, err := r.controller.dynamicclient.Resource(*r.controller.machineDeploymentResource).Namespace(r.machineDeployment.Namespace).Get(context.TODO(), r.machineDeployment.Name, metav1.GetOptions{})

	if err != nil {
		return err
	}

	if u == nil {
		return fmt.Errorf("unknown machineDeployment %s", r.machineDeployment.Name)
	}

	u = u.DeepCopy()
	if err := unstructured.SetNestedField(u.Object, int64(nreplicas), "spec", "replicas"); err != nil {
		return fmt.Errorf("failed to set replica value: %v", err)
	}

	_, updateErr := r.controller.dynamicclient.Resource(*r.controller.machineDeploymentResource).Namespace(u.GetNamespace()).Update(context.TODO(), u, metav1.UpdateOptions{})
	return updateErr
}

func (r machineDeploymentScalableResource) UnmarkMachineForDeletion(machine *Machine) error {
	return unmarkMachineForDeletion(r.controller, machine)
}

func (r machineDeploymentScalableResource) MarkMachineForDeletion(machine *Machine) error {
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

func newMachineDeploymentScalableResource(controller *machineController, machineDeployment *MachineDeployment) (*machineDeploymentScalableResource, error) {
	minSize, maxSize, err := parseScalingBounds(machineDeployment.Annotations)
	if err != nil {
		return nil, fmt.Errorf("error validating min/max annotations: %v", err)
	}

	return &machineDeploymentScalableResource{
		controller:        controller,
		machineDeployment: machineDeployment,
		maxSize:           maxSize,
		minSize:           minSize,
	}, nil
}
