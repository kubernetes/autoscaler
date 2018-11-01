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
	"k8s.io/api/core/v1"
	apimachv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterclientset "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	"strconv"
)

const (
	// MinSizeAnnotation sets a MachineDeployment's minimum size during autoscaling
	MinSizeAnnotation = "cluster-autoscaler/min-size"
	// MaxSizeAnnotation sets a MachineDeployment's maximum size during autoscaling
	MaxSizeAnnotation = "cluster-autoscaler/max-size"
)

// MachineDeploymentAttrs holds parsed-out attributes of a MD
type MachineDeploymentAttrs struct {
	minSize, maxSize int
}

// GetMachineDeploymentAttrs extracts MachineDeploymentAttrs from a given MachineDeployment
func GetMachineDeploymentAttrs(md *v1alpha1.MachineDeployment) *MachineDeploymentAttrs {
	attrs := &MachineDeploymentAttrs{}

	var err error

	if val, ok := md.Annotations[MinSizeAnnotation]; ok {
		attrs.minSize, err = strconv.Atoi(val)
		if err != nil {
			klog.Errorf("In %s: Invalid min-size: %v (%s)", md.Name, val, err)
			return nil
		}
	} else {
		return nil
	}

	if val, ok := md.Annotations[MaxSizeAnnotation]; ok {
		attrs.maxSize, err = strconv.Atoi(val)
		if err != nil {
			klog.Errorf("In %s: Invalid max-size: %v (%s)", md.Name, val, err)
			return nil
		}
	} else {
		return nil
	}

	return attrs
}

// MachineManager interface
type MachineManager interface {
	AllDeployments() []*v1alpha1.MachineDeployment
	DeploymentForNode(node *v1.Node) *v1alpha1.MachineDeployment
	NodesForDeployment(md *v1alpha1.MachineDeployment) []*v1.Node
	Refresh() error
	SetDeploymentSize(md *v1alpha1.MachineDeployment, size int) error
}

// ClusterapiMachineManager is a facade and cache for accessing the cluster's nodes, machines, and MachineDeployments
type ClusterapiMachineManager struct {
	k8sClient        *kubernetes.Clientset
	clusterApiClient clusterclientset.Interface

	// cache data structures.
	// each api object (Node, Machine, MachineDeployment etc.) is stored as a unique
	// pointer shared across all data structures.

	allDeploymentsByUid map[types.UID]*v1alpha1.MachineDeployment

	deploymentByMachineUid  map[types.UID]*v1alpha1.MachineDeployment
	nodeByMachineUid        map[types.UID]*v1.Node
	machinesByDeploymentUid map[types.UID][]*v1alpha1.Machine

	deploymentByNodeUid  map[types.UID]*v1alpha1.MachineDeployment
	machineByNodeUid     map[types.UID]*v1alpha1.Machine
	nodesByDeploymentUid map[types.UID][]*v1.Node
}

// NewMachineManager creates a new empty ClusterapiMachineManager. Call Refresh() to initialize it
func NewMachineManager(kubeConfig *rest.Config) (*ClusterapiMachineManager, error) {
	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	clusterApiClient, err := clusterclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	mm := &ClusterapiMachineManager{
		k8sClient:        k8sClient,
		clusterApiClient: clusterApiClient,
	}

	return mm, nil
}

// AllDeployments returns all MachineDeployments of the cluster
func (mm *ClusterapiMachineManager) AllDeployments() []*v1alpha1.MachineDeployment {
	result := make([]*v1alpha1.MachineDeployment, 0)
	for _, md := range mm.allDeploymentsByUid {
		result = append(result, md)
	}
	return result
}

// DeploymentForNode returns the MachineDeployment that created a specific node
func (mm *ClusterapiMachineManager) DeploymentForNode(node *v1.Node) *v1alpha1.MachineDeployment {
	return mm.deploymentByNodeUid[node.UID]
}

// NodesForDeployment returns all nodes that were created by a specific MachineDeployment
func (mm *ClusterapiMachineManager) NodesForDeployment(md *v1alpha1.MachineDeployment) []*v1.Node {
	return mm.nodesByDeploymentUid[md.UID]
}

// Refresh reloads the ClusterapiMachineManager's cached representation of the cluster state
func (mm *ClusterapiMachineManager) Refresh() error {
	newAllDeploymentsByUid := make(map[types.UID]*v1alpha1.MachineDeployment)

	newDeploymentByMachineUid := make(map[types.UID]*v1alpha1.MachineDeployment)
	newNodeByMachineUid := make(map[types.UID]*v1.Node)
	newMachinesByDeploymentUid := make(map[types.UID][]*v1alpha1.Machine)

	newDeploymentByNodeUid := make(map[types.UID]*v1alpha1.MachineDeployment)
	newMachineByNodeUid := make(map[types.UID]*v1alpha1.Machine)
	newNodesByDeploymentUid := make(map[types.UID][]*v1.Node)

	machines, err := mm.clusterApiClient.ClusterV1alpha1().Machines("kube-system").List(apimachv1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range machines.Items {
		machine := &machines.Items[i]

		// TODO consider fetching all machines, nodes, ms's and mds in one API call each and cross-refing them in memory,
		// so that for n machines we have just 4 API calls rather than 3n+1

		var node *v1.Node

		if nodeRef := machine.Status.NodeRef; nodeRef != nil {
			node, err = mm.k8sClient.CoreV1().Nodes().Get(nodeRef.Name, apimachv1.GetOptions{})
			if err != nil {
				return err
			}
			newNodeByMachineUid[machine.UID] = node
			newMachineByNodeUid[node.UID] = machine
		}

		if msRef, ok := findRefByKind(machine.OwnerReferences, "MachineSet"); ok {
			ms, err := mm.clusterApiClient.ClusterV1alpha1().MachineSets("kube-system").Get(msRef.Name, apimachv1.GetOptions{})
			if err != nil {
				return err
			}

			if mdRef, ok := findRefByKind(ms.OwnerReferences, "MachineDeployment"); ok {
				md, err := mm.clusterApiClient.ClusterV1alpha1().MachineDeployments("kube-system").Get(mdRef.Name, apimachv1.GetOptions{})
				if err != nil {
					return err
				}

				if nil == GetMachineDeploymentAttrs(md) {
					klog.Infof("MachineDeployment %s has no valid autoscaler annotations; ignoring.", md.Name)
					continue
				}

				newAllDeploymentsByUid[md.UID] = md
				newDeploymentByMachineUid[machine.UID] = md
				newMachinesByDeploymentUid[md.UID] = append(newMachinesByDeploymentUid[md.UID], machine)

				if node != nil {
					newDeploymentByNodeUid[node.UID] = md
					newNodesByDeploymentUid[md.UID] = append(newNodesByDeploymentUid[md.UID], node)
				}
			}
		}
	}

	// So far we've only found MachineDeployments containing at least one machine.
	// Iterate over all MachineDeployments directly to also find ones with no machines.
	mds, err := mm.clusterApiClient.ClusterV1alpha1().MachineDeployments("kube-system").List(apimachv1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range mds.Items {
		md := &mds.Items[i]
		if nil == GetMachineDeploymentAttrs(md) {
			klog.Infof("MachineDeployment %s has no valid autoscaler annotations; ignoring.", md.Name)
			continue
		}
		if _, ok := newAllDeploymentsByUid[md.UID]; !ok {
			newAllDeploymentsByUid[md.UID] = md
		}
	}

	mm.allDeploymentsByUid = newAllDeploymentsByUid

	mm.deploymentByMachineUid = newDeploymentByMachineUid
	mm.nodeByMachineUid = newNodeByMachineUid
	mm.machinesByDeploymentUid = newMachinesByDeploymentUid

	mm.deploymentByNodeUid = newDeploymentByNodeUid
	mm.machineByNodeUid = newMachineByNodeUid
	mm.nodesByDeploymentUid = newNodesByDeploymentUid

	return nil
}

// SetDeploymentSize sets a MachineDeployment's replica count
func (mm *ClusterapiMachineManager) SetDeploymentSize(md *v1alpha1.MachineDeployment, size int) error {
	// check that we know the md
	internalMd := mm.allDeploymentsByUid[md.UID]
	if internalMd == nil {
		// shouldn't happen as autoscaler should ony pass us mds that we handed out previously
		return fmt.Errorf("STRANGE: MachineDeployment not cached: %v", md.Name)
	}

	internalMd.Spec.Replicas = int32Ptr(int32(size))
	md.Spec.Replicas = int32Ptr(int32(size))

	_, err := mm.clusterApiClient.ClusterV1alpha1().MachineDeployments("kube-system").Update(md)
	return err
}

func findRefByKind(orefs []apimachv1.OwnerReference, kind string) (apimachv1.OwnerReference, bool) {
	for _, ownerRef := range orefs {
		if *ownerRef.Controller && ownerRef.Kind == kind {
			return ownerRef, true
		}
	}
	return apimachv1.OwnerReference{}, false
}

func int32Ptr(i int32) *int32 {
	return &i
}
