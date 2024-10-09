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

package rancher

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	provisioningv1 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/rancher/provisioning.cattle.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

// nodeGroup implements nodeGroup for rancher machine pools.
type nodeGroup struct {
	provider  *RancherCloudProvider
	name      string
	labels    map[string]string
	taints    []corev1.Taint
	minSize   int
	maxSize   int
	resources corev1.ResourceList
	replicas  int
	machines  []unstructured.Unstructured
}

type node struct {
	instance cloudprovider.Instance
	machine  unstructured.Unstructured
}

var (
	// errMissingMinSizeAnnotation is the error returned when a machine pool does
	// not have the min size annotations attached.
	errMissingMinSizeAnnotation = errors.New("missing min size annotation")

	// errMissingMaxSizeAnnotation is the error returned when a machine pool does
	// not have the max size annotations attached.
	errMissingMaxSizeAnnotation = errors.New("missing max size annotation")

	// errMissingResourceAnnotation is the error returned when a machine pool does
	// not have all the resource annotations attached.
	errMissingResourceAnnotation = errors.New("missing resource annotation")
)

const podCapacity = 110

// Id returns node group id/name.
func (ng *nodeGroup) Id() string {
	return ng.name
}

// MinSize returns minimum size of the node group.
func (ng *nodeGroup) MinSize() int {
	return ng.minSize
}

// MaxSize returns maximum size of the node group.
func (ng *nodeGroup) MaxSize() int {
	return ng.maxSize
}

// Debug returns a debug string for the node group.
func (ng *nodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", ng.Id(), ng.MinSize(), ng.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (ng *nodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes, err := ng.nodes()
	if err != nil {
		return nil, err
	}

	instances := make([]cloudprovider.Instance, 0, len(nodes))
	for _, node := range nodes {
		instances = append(instances, node.instance)
	}

	return instances, nil
}

// DeleteNodes deletes the specified nodes from the node group.
func (ng *nodeGroup) DeleteNodes(toDelete []*corev1.Node) error {
	if ng.replicas-len(toDelete) < ng.MinSize() {
		return fmt.Errorf("node group size would be below minimum size - desired: %d, min: %d",
			ng.replicas-len(toDelete), ng.MinSize())
	}

	for _, del := range toDelete {
		node, err := ng.findNodeByProviderID(del.Spec.ProviderID)
		if err != nil {
			return err
		}

		klog.V(4).Infof("marking machine for deletion: %v", node.instance.Id)

		if err := node.markMachineForDeletion(ng); err != nil {
			return fmt.Errorf("unable to mark machine %s for deletion: %w", del.Name, err)
		}

		if err := ng.setSize(ng.replicas - 1); err != nil {
			// rollback deletion mark
			_ = node.unmarkMachineForDeletion(ng)
			return fmt.Errorf("unable to set node group size: %w", err)
		}
	}

	return nil
}

func (ng *nodeGroup) findNodeByProviderID(providerID string) (*node, error) {
	nodes, err := ng.nodes()
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if node.instance.Id == providerID {
			return &node, nil
		}
	}

	return nil, fmt.Errorf("node with providerID %s not found in node group %s", providerID, ng.name)
}

// IncreaseSize increases NodeGroup size.
func (ng *nodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	newSize := ng.replicas + delta
	if newSize > ng.MaxSize() {
		return fmt.Errorf("size increase too large, desired: %d max: %d", newSize, ng.MaxSize())
	}

	return ng.setSize(newSize)
}

// AtomicIncreaseSize is not implemented.
func (ng *nodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (ng *nodeGroup) TargetSize() (int, error) {
	return ng.replicas, nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (ng *nodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}

	nodes, err := ng.Nodes()
	if err != nil {
		return fmt.Errorf("failed to get node group nodes: %w", err)
	}

	if ng.replicas+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize: %d delta: %d existingNodes: %d",
			ng.replicas, delta, len(nodes))
	}

	return ng.setSize(ng.replicas + delta)
}

// TemplateNodeInfo returns a node template for this node group.
func (ng *nodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s-%d", ng.provider.config.ClusterName, ng.Id(), rand.Int63()),
			Labels: ng.labels,
		},
		Spec: corev1.NodeSpec{
			Taints: ng.taints,
		},
		Status: corev1.NodeStatus{
			Capacity:   ng.resources,
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}

	node.Status.Capacity[corev1.ResourcePods] = *resource.NewQuantity(podCapacity, resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	// Setup node info template
	nodeInfo := framework.NewNodeInfo(node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(ng.Id())})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side.
func (ng *nodeGroup) Exist() bool {
	return ng.Id() != ""
}

// Create creates the node group on the cloud provider side.
func (ng *nodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
func (ng *nodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (ng *nodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (ng *nodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (ng *nodeGroup) setSize(size int) error {
	machinePools, err := ng.provider.getMachinePools()
	if err != nil {
		return err
	}

	found := false
	for i := range machinePools {
		if machinePools[i].Name == ng.name {
			machinePools[i].Quantity = pointer.Int32Ptr(int32(size))
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unable to set size of group %s of cluster %s: group not found",
			ng.name, ng.provider.config.ClusterName)
	}

	if err := ng.provider.updateMachinePools(machinePools); err != nil {
		return err
	}

	ng.replicas = size
	return nil
}

// nodes returns all nodes of this node group that have a provider ID set by
// getting the underlying machines and extracting the providerID, which
// corresponds to the name of the k8s node object.
func (ng *nodeGroup) nodes() ([]node, error) {
	machines, err := ng.listMachines()
	if err != nil {
		return nil, err
	}

	nodes := make([]node, 0, len(machines))
	for _, machine := range machines {
		phase, found, err := unstructured.NestedString(machine.UnstructuredContent(), "status", "phase")
		if err != nil {
			return nil, err
		}

		if !found {
			return nil, fmt.Errorf("machine %s/%s does not have status.phase field", machine.GetName(), machine.GetNamespace())
		}

		providerID, found, err := unstructured.NestedString(machine.UnstructuredContent(), "spec", "providerID")
		if err != nil {
			return nil, err
		}

		if !found {
			if phase == machinePhaseProvisioning {
				// if the provider ID is missing during provisioning, we
				// ignore this node to avoid errors in the autoscaler.
				continue
			}

			return nil, fmt.Errorf("could not find providerID in machine: %s/%s", machine.GetName(), machine.GetNamespace())
		}

		state := cloudprovider.InstanceRunning

		switch phase {
		case machinePhasePending, machinePhaseProvisioning:
			state = cloudprovider.InstanceCreating
		case machinePhaseDeleting:
			state = cloudprovider.InstanceDeleting
		}

		nodes = append(nodes, node{
			machine: machine,
			instance: cloudprovider.Instance{
				Id: providerID,
				Status: &cloudprovider.InstanceStatus{
					State: state,
				},
			},
		})
	}

	return nodes, nil
}

// listMachines returns the unstructured objects of all cluster-api machines
// in a node group. The machines are found using the deployment name label.
func (ng *nodeGroup) listMachines() ([]unstructured.Unstructured, error) {
	if ng.machines != nil {
		return ng.machines, nil
	}

	machinesList, err := ng.provider.client.Resource(machineGVR(ng.provider.config.ClusterAPIVersion)).
		Namespace(ng.provider.config.ClusterNamespace).List(
		context.TODO(), metav1.ListOptions{
			// we find all machines belonging to an rke2 machinePool by the
			// deployment name, since it is just <cluster name>-<machinePool name>
			LabelSelector: fmt.Sprintf("%s=%s-%s", machineDeploymentNameLabelKey, ng.provider.config.ClusterName, ng.name),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not list machines: %w", err)
	}

	ng.machines = machinesList.Items
	return machinesList.Items, nil
}

func (ng *nodeGroup) machineByName(name string) (*unstructured.Unstructured, error) {
	machines, err := ng.listMachines()
	if err != nil {
		return nil, fmt.Errorf("error listing machines in node group %s: %w", ng.name, err)
	}

	for _, machine := range machines {
		if machine.GetName() == name {
			return &machine, nil
		}
	}

	return nil, fmt.Errorf("machine %s not found in list", name)
}

// markMachineForDeletion sets an annotation on the cluster-api machine
// object, inidicating that this node is a candidate to be removed on scale
// down of the controlling resource (machineSet/machineDeployment).
func (n *node) markMachineForDeletion(ng *nodeGroup) error {
	u, err := ng.provider.client.Resource(machineGVR(ng.provider.config.ClusterAPIVersion)).Namespace(n.machine.GetNamespace()).
		Get(context.TODO(), n.machine.GetName(), metav1.GetOptions{})
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

	_, err = ng.provider.client.Resource(machineGVR(ng.provider.config.ClusterAPIVersion)).Namespace(u.GetNamespace()).
		Update(context.TODO(), u, metav1.UpdateOptions{})

	return err
}

// unmarkMachineForDeletion removes the machine delete annotation.
func (n *node) unmarkMachineForDeletion(ng *nodeGroup) error {
	u, err := ng.provider.client.Resource(machineGVR(ng.provider.config.ClusterAPIVersion)).Namespace(n.machine.GetNamespace()).
		Get(context.TODO(), n.machine.GetName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	u = u.DeepCopy()

	annotations := u.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	delete(annotations, machineDeleteAnnotationKey)
	u.SetAnnotations(annotations)

	_, err = ng.provider.client.Resource(machineGVR(ng.provider.config.ClusterAPIVersion)).Namespace(u.GetNamespace()).
		Update(context.TODO(), u, metav1.UpdateOptions{})

	return err
}

func newNodeGroupFromMachinePool(provider *RancherCloudProvider, machinePool provisioningv1.RKEMachinePool) (*nodeGroup, error) {
	if machinePool.Quantity == nil {
		return nil, errors.New("machine pool quantity is not set")
	}

	minSize, maxSize, err := parseScalingAnnotations(machinePool.MachineDeploymentAnnotations)
	if err != nil {
		return nil, fmt.Errorf("error parsing scaling annotations: %w", err)
	}

	resources, err := parseResourceAnnotations(machinePool.MachineDeploymentAnnotations)
	if err != nil {
		if !errors.Is(err, errMissingResourceAnnotation) {
			return nil, fmt.Errorf("error parsing resource annotations: %w", err)
		}
		// if the resource labels are missing, we simply initialize an empty
		// list. The autoscaler can still work but won't scale up from 0 if a
		// pod requests any resources.
		resources = corev1.ResourceList{}
	}

	return &nodeGroup{
		provider:  provider,
		name:      machinePool.Name,
		labels:    machinePool.Labels,
		taints:    machinePool.Taints,
		minSize:   minSize,
		maxSize:   maxSize,
		replicas:  int(*machinePool.Quantity),
		resources: resources,
	}, nil
}

func parseResourceAnnotations(annotations map[string]string) (corev1.ResourceList, error) {
	cpu, ok := annotations[resourceCPUAnnotation]
	if !ok {
		return nil, errMissingResourceAnnotation
	}

	cpuResources, err := resource.ParseQuantity(cpu)
	if err != nil {
		return nil, fmt.Errorf("unable to parse cpu resources: %q: %w", cpu, err)
	}
	memory, ok := annotations[resourceMemoryAnnotation]
	if !ok {
		return nil, errMissingResourceAnnotation
	}

	memoryResources, err := resource.ParseQuantity(memory)
	if err != nil {
		return nil, fmt.Errorf("unable to parse memory resources: %q: %w", memory, err)
	}
	ephemeralStorage, ok := annotations[resourceEphemeralStorageAnnotation]
	if !ok {
		return nil, errMissingResourceAnnotation
	}

	ephemeralStorageResources, err := resource.ParseQuantity(ephemeralStorage)
	if err != nil {
		return nil, fmt.Errorf("unable to parse ephemeral storage resources: %q: %w", ephemeralStorage, err)
	}

	return corev1.ResourceList{
		corev1.ResourceCPU:              cpuResources,
		corev1.ResourceMemory:           memoryResources,
		corev1.ResourceEphemeralStorage: ephemeralStorageResources,
	}, nil
}

func parseScalingAnnotations(annotations map[string]string) (int, int, error) {
	min, ok := annotations[minSizeAnnotation]
	if !ok {
		return 0, 0, errMissingMinSizeAnnotation
	}

	minSize, err := strconv.Atoi(min)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to parse min size: %s", min)
	}

	max, ok := annotations[maxSizeAnnotation]
	if !ok {
		return 0, 0, errMissingMaxSizeAnnotation
	}

	maxSize, err := strconv.Atoi(max)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to parse min size: %s", min)
	}

	if minSize < 0 || maxSize < 0 {
		return 0, 0, fmt.Errorf("invalid min or max size supplied: %v/%v", minSize, maxSize)
	}

	return minSize, maxSize, nil
}
