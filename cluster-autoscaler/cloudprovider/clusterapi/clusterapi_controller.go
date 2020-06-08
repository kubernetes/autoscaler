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
	"os"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	kubeinformers "k8s.io/client-go/informers"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

const (
	machineProviderIDIndex = "machineProviderIDIndex"
	nodeProviderIDIndex    = "nodeProviderIDIndex"
	defaultCAPIGroup       = "cluster.x-k8s.io"
	// CAPIGroupEnvVar contains the environment variable name which allows overriding defaultCAPIGroup.
	CAPIGroupEnvVar               = "CAPI_GROUP"
	resourceNameMachine           = "machines"
	resourceNameMachineSet        = "machinesets"
	resourceNameMachineDeployment = "machinedeployments"
	failedMachinePrefix           = "failed-machine-"
)

// machineController watches for Nodes, Machines, MachineSets and
// MachineDeployments as they are added, updated and deleted on the
// cluster. Additionally, it adds indices to the node informers to
// satisfy lookup by node.Spec.ProviderID.
type machineController struct {
	kubeInformerFactory       kubeinformers.SharedInformerFactory
	machineInformerFactory    dynamicinformer.DynamicSharedInformerFactory
	machineDeploymentInformer informers.GenericInformer
	machineInformer           informers.GenericInformer
	machineSetInformer        informers.GenericInformer
	nodeInformer              cache.SharedIndexInformer
	dynamicclient             dynamic.Interface
	machineSetResource        *schema.GroupVersionResource
	machineResource           *schema.GroupVersionResource
	machineDeploymentResource *schema.GroupVersionResource
	accessLock                sync.Mutex
}

type machineSetFilterFunc func(machineSet *MachineSet) error

func indexMachineByProviderID(obj interface{}) ([]string, error) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, nil
	}

	providerID, found, err := unstructured.NestedString(u.Object, "spec", "providerID")
	if err != nil || !found {
		return nil, nil
	}
	if providerID == "" {
		return nil, nil
	}

	return []string{string(normalizedProviderString(providerID))}, nil
}

func indexNodeByProviderID(obj interface{}) ([]string, error) {
	if node, ok := obj.(*corev1.Node); ok {
		if node.Spec.ProviderID != "" {
			return []string{string(normalizedProviderString(node.Spec.ProviderID))}, nil
		}
		return []string{}, nil
	}
	return []string{}, nil
}

func (c *machineController) findMachine(id string) (*Machine, error) {
	item, exists, err := c.machineInformer.Informer().GetStore().GetByKey(id)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	u, ok := item.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("internal error; unexpected type: %T", item)
	}

	machine := newMachineFromUnstructured(u.DeepCopy())
	if machine == nil {
		return nil, nil
	}

	return machine, nil
}

func (c *machineController) findMachineDeployment(id string) (*MachineDeployment, error) {
	item, exists, err := c.machineDeploymentInformer.Informer().GetStore().GetByKey(id)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	u, ok := item.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("internal error; unexpected type: %T", item)
	}

	machineDeployment := newMachineDeploymentFromUnstructured(u.DeepCopy())
	if machineDeployment == nil {
		return nil, nil
	}

	return machineDeployment, nil
}

// findMachineOwner returns the machine set owner for machine, or nil
// if there is no owner. A DeepCopy() of the object is returned on
// success.
func (c *machineController) findMachineOwner(machine *Machine) (*MachineSet, error) {
	machineOwnerRef := machineOwnerRef(machine)
	if machineOwnerRef == nil {
		return nil, nil
	}

	store := c.machineSetInformer.Informer().GetStore()
	item, exists, err := store.GetByKey(fmt.Sprintf("%s/%s", machine.Namespace, machineOwnerRef.Name))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	u, ok := item.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("internal error; unexpected type: %T", item)
	}

	u = u.DeepCopy()
	machineSet := newMachineSetFromUnstructured(u)
	if machineSet == nil {
		return nil, nil
	}

	if !machineIsOwnedByMachineSet(machine, machineSet) {
		return nil, nil
	}

	return machineSet, nil
}

// run starts shared informers and waits for the informer cache to
// synchronize.
func (c *machineController) run(stopCh <-chan struct{}) error {
	c.kubeInformerFactory.Start(stopCh)
	c.machineInformerFactory.Start(stopCh)

	syncFuncs := []cache.InformerSynced{
		c.nodeInformer.HasSynced,
		c.machineInformer.Informer().HasSynced,
		c.machineSetInformer.Informer().HasSynced,
	}
	if c.machineDeploymentResource != nil {
		syncFuncs = append(syncFuncs, c.machineDeploymentInformer.Informer().HasSynced)
	}

	klog.V(4).Infof("waiting for caches to sync")
	if !cache.WaitForCacheSync(stopCh, syncFuncs...) {
		return fmt.Errorf("syncing caches failed")
	}

	return nil
}

// findMachineByProviderID finds machine matching providerID. A
// DeepCopy() of the object is returned on success.
func (c *machineController) findMachineByProviderID(providerID normalizedProviderID) (*Machine, error) {
	objs, err := c.machineInformer.Informer().GetIndexer().ByIndex(machineProviderIDIndex, string(providerID))
	if err != nil {
		return nil, err
	}

	switch n := len(objs); {
	case n > 1:
		return nil, fmt.Errorf("internal error; expected len==1, got %v", n)
	case n == 1:
		u, ok := objs[0].(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("internal error; unexpected type %T", objs[0])
		}
		machine := newMachineFromUnstructured(u.DeepCopy())
		if machine != nil {
			return machine, nil
		}
	}

	if isFailedMachineProviderID(providerID) {
		machine, err := c.findMachine(machineKeyFromFailedProviderID(providerID))
		if err != nil {
			return nil, err
		}
		if machine != nil {
			return machine.DeepCopy(), nil
		}
	}

	// If the machine object has no providerID--maybe actuator
	// does not set this value (e.g., OpenStack)--then first
	// lookup the node using ProviderID. If that is successful
	// then the machine can be found using the annotation (should
	// it exist).
	node, err := c.findNodeByProviderID(providerID)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}
	return c.findMachine(node.Annotations[machineAnnotationKey])
}

func isFailedMachineProviderID(providerID normalizedProviderID) bool {
	return strings.HasPrefix(string(providerID), failedMachinePrefix)
}

func machineKeyFromFailedProviderID(providerID normalizedProviderID) string {
	namespaceName := strings.TrimPrefix(string(providerID), failedMachinePrefix)
	return strings.Replace(namespaceName, "_", "/", 1)
}

// findNodeByNodeName finds the Node object keyed by name.. Returns
// nil if it cannot be found. A DeepCopy() of the object is returned
// on success.
func (c *machineController) findNodeByNodeName(name string) (*corev1.Node, error) {
	item, exists, err := c.nodeInformer.GetIndexer().GetByKey(name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	node, ok := item.(*corev1.Node)
	if !ok {
		return nil, fmt.Errorf("internal error; unexpected type %T", item)
	}

	return node.DeepCopy(), nil
}

// machinesInMachineSet returns all the machines that belong to
// machineSet. For each machine in the set a DeepCopy() of the object
// is returned.
func (c *machineController) machinesInMachineSet(machineSet *MachineSet) ([]*Machine, error) {
	machines, err := c.listMachines(machineSet.Namespace, labels.SelectorFromSet(machineSet.Labels))
	if err != nil {
		return nil, err
	}
	if machines == nil {
		return nil, nil
	}

	var result []*Machine

	for _, machine := range machines {
		if machineIsOwnedByMachineSet(machine, machineSet) {
			result = append(result, machine)
		}
	}

	return result, nil
}

// getCAPIGroup returns a string that specifies the group for the API.
// It will return either the value from the
// CAPI_GROUP environment variable, or the default value i.e cluster.x-k8s.io.
func getCAPIGroup() string {
	g := os.Getenv(CAPIGroupEnvVar)
	if g == "" {
		g = defaultCAPIGroup
	}
	klog.V(4).Infof("Using API Group %q", g)
	return g
}

// newMachineController constructs a controller that watches Nodes,
// Machines and MachineSet as they are added, updated and deleted on
// the cluster.
func newMachineController(
	dynamicclient dynamic.Interface,
	kubeclient kubeclient.Interface,
	discoveryclient discovery.DiscoveryInterface,
) (*machineController, error) {
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeclient, 0)
	informerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicclient, 0, metav1.NamespaceAll, nil)

	CAPIGroup := getCAPIGroup()
	CAPIVersion, err := getAPIGroupPreferredVersion(discoveryclient, CAPIGroup)
	if err != nil {
		return nil, fmt.Errorf("could not find preferred version for CAPI group %q: %v", CAPIGroup, err)
	}
	klog.Infof("Using version %q for API group %q", CAPIVersion, CAPIGroup)

	var gvrMachineDeployment *schema.GroupVersionResource
	var machineDeploymentInformer informers.GenericInformer

	machineDeployment, err := groupVersionHasResource(discoveryclient,
		fmt.Sprintf("%s/%s", CAPIGroup, CAPIVersion), resourceNameMachineDeployment)
	if err != nil {
		return nil, fmt.Errorf("failed to validate if resource %q is available for group %q: %v",
			resourceNameMachineDeployment, fmt.Sprintf("%s/%s", CAPIGroup, CAPIVersion), err)
	}

	if machineDeployment {
		gvrMachineDeployment = &schema.GroupVersionResource{
			Group:    CAPIGroup,
			Version:  CAPIVersion,
			Resource: resourceNameMachineDeployment,
		}
		machineDeploymentInformer = informerFactory.ForResource(*gvrMachineDeployment)
		machineDeploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})
	}

	gvrMachineSet := &schema.GroupVersionResource{
		Group:    CAPIGroup,
		Version:  CAPIVersion,
		Resource: resourceNameMachineSet,
	}
	machineSetInformer := informerFactory.ForResource(*gvrMachineSet)
	machineSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	gvrMachine := &schema.GroupVersionResource{
		Group:    CAPIGroup,
		Version:  CAPIVersion,
		Resource: resourceNameMachine,
	}
	machineInformer := informerFactory.ForResource(*gvrMachine)
	machineInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	nodeInformer := kubeInformerFactory.Core().V1().Nodes().Informer()
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	if err := machineInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
		machineProviderIDIndex: indexMachineByProviderID,
	}); err != nil {
		return nil, fmt.Errorf("cannot add machine indexer: %v", err)
	}

	if err := nodeInformer.GetIndexer().AddIndexers(cache.Indexers{
		nodeProviderIDIndex: indexNodeByProviderID,
	}); err != nil {
		return nil, fmt.Errorf("cannot add node indexer: %v", err)
	}

	return &machineController{
		kubeInformerFactory:       kubeInformerFactory,
		machineInformerFactory:    informerFactory,
		machineDeploymentInformer: machineDeploymentInformer,
		machineInformer:           machineInformer,
		machineSetInformer:        machineSetInformer,
		nodeInformer:              nodeInformer,
		dynamicclient:             dynamicclient,
		machineSetResource:        gvrMachineSet,
		machineResource:           gvrMachine,
		machineDeploymentResource: gvrMachineDeployment,
	}, nil
}

func groupVersionHasResource(client discovery.DiscoveryInterface, groupVersion, resourceName string) (bool, error) {
	resourceList, err := client.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return false, fmt.Errorf("failed to get ServerGroups: %v", err)
	}

	for _, r := range resourceList.APIResources {
		klog.Infof("Resource %q available", r.Name)
		if r.Name == resourceName {
			return true, nil
		}
	}
	return false, nil
}

func getAPIGroupPreferredVersion(client discovery.DiscoveryInterface, APIGroup string) (string, error) {
	groupList, err := client.ServerGroups()
	if err != nil {
		return "", fmt.Errorf("failed to get ServerGroups: %v", err)
	}

	for _, group := range groupList.Groups {
		if group.Name == APIGroup {
			return group.PreferredVersion.Version, nil
		}
	}

	return "", fmt.Errorf("failed to find API group %q", APIGroup)
}

func (c *machineController) machineSetProviderIDs(machineSet *MachineSet) ([]string, error) {
	machines, err := c.machinesInMachineSet(machineSet)
	if err != nil {
		return nil, fmt.Errorf("error listing machines: %v", err)
	}

	var providerIDs []string
	for _, machine := range machines {
		if machine.Spec.ProviderID == nil || *machine.Spec.ProviderID == "" {
			klog.Warningf("Machine %q has no providerID", machine.Name)
		}

		if machine.Spec.ProviderID != nil && *machine.Spec.ProviderID != "" {
			providerIDs = append(providerIDs, *machine.Spec.ProviderID)
			continue
		}

		if machine.Status.FailureMessage != nil {
			klog.V(4).Infof("Status.FailureMessage of machine %q is %q", machine.Name, *machine.Status.FailureMessage)
			// Provide a fake ID to allow the autoscaler to track machines that will never
			// become nodes and mark the nodegroup unhealthy after maxNodeProvisionTime.
			// Fake ID needs to be recognised later and converted into a machine key.
			// Use an underscore as a separator between namespace and name as it is not a
			// valid character within a namespace name.
			providerIDs = append(providerIDs, fmt.Sprintf("%s%s_%s", failedMachinePrefix, machine.Namespace, machine.Name))
			continue
		}

		if machine.Status.NodeRef == nil {
			klog.V(4).Infof("Status.NodeRef of machine %q is currently nil", machine.Name)
			continue
		}

		if machine.Status.NodeRef.Kind != "Node" {
			klog.Errorf("Status.NodeRef of machine %q does not reference a node (rather %q)", machine.Name, machine.Status.NodeRef.Kind)
			continue
		}

		node, err := c.findNodeByNodeName(machine.Status.NodeRef.Name)
		if err != nil {
			return nil, fmt.Errorf("unknown node %q", machine.Status.NodeRef.Name)
		}

		if node != nil {
			providerIDs = append(providerIDs, node.Spec.ProviderID)
		}
	}

	klog.V(4).Infof("nodegroup %s has nodes %v", machineSet.Name, providerIDs)
	return providerIDs, nil
}

func (c *machineController) filterAllMachineSets(f machineSetFilterFunc) error {
	return c.filterMachineSets(metav1.NamespaceAll, f)
}

func (c *machineController) filterMachineSets(namespace string, f machineSetFilterFunc) error {
	machineSets, err := c.listMachineSets(namespace, labels.Everything())
	if err != nil {
		return nil
	}
	for _, machineSet := range machineSets {
		if err := f(machineSet); err != nil {
			return err
		}
	}
	return nil
}

func (c *machineController) machineSetNodeGroups() ([]*nodegroup, error) {
	var nodegroups []*nodegroup

	if err := c.filterAllMachineSets(func(machineSet *MachineSet) error {
		if machineSetHasMachineDeploymentOwnerRef(machineSet) {
			return nil
		}
		ng, err := newNodegroupFromMachineSet(c, machineSet)
		if err != nil {
			return err
		}
		if ng.MaxSize()-ng.MinSize() > 0 && pointer.Int32PtrDerefOr(machineSet.Spec.Replicas, 0) > 0 {
			nodegroups = append(nodegroups, ng)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return nodegroups, nil
}

func (c *machineController) machineDeploymentNodeGroups() ([]*nodegroup, error) {
	machineDeployments, err := c.listMachineDeployments(metav1.NamespaceAll, labels.Everything())
	if err != nil {
		return nil, err
	}

	var nodegroups []*nodegroup

	for _, md := range machineDeployments {
		ng, err := newNodegroupFromMachineDeployment(c, md)
		if err != nil {
			return nil, err
		}
		// add nodegroup iff it has the capacity to scale
		if ng.MaxSize()-ng.MinSize() > 0 && pointer.Int32PtrDerefOr(md.Spec.Replicas, 0) > 0 {
			nodegroups = append(nodegroups, ng)
		}
	}

	return nodegroups, nil
}

func (c *machineController) nodeGroups() ([]*nodegroup, error) {
	machineSets, err := c.machineSetNodeGroups()
	if err != nil {
		return nil, err
	}

	if c.machineDeploymentResource != nil {
		machineDeployments, err := c.machineDeploymentNodeGroups()
		if err != nil {
			return nil, err
		}
		machineSets = append(machineSets, machineDeployments...)
	}

	return machineSets, nil
}

func (c *machineController) nodeGroupForNode(node *corev1.Node) (*nodegroup, error) {
	machine, err := c.findMachineByProviderID(normalizedProviderString(node.Spec.ProviderID))
	if err != nil {
		return nil, err
	}
	if machine == nil {
		return nil, nil
	}

	machineSet, err := c.findMachineOwner(machine)
	if err != nil {
		return nil, err
	}

	if machineSet == nil {
		return nil, nil
	}

	if c.machineDeploymentResource != nil {
		if ref := machineSetMachineDeploymentRef(machineSet); ref != nil {
			key := fmt.Sprintf("%s/%s", machineSet.Namespace, ref.Name)
			machineDeployment, err := c.findMachineDeployment(key)
			if err != nil {
				return nil, fmt.Errorf("unknown MachineDeployment %q: %v", key, err)
			}
			if machineDeployment == nil {
				return nil, fmt.Errorf("unknown MachineDeployment %q", key)
			}
			nodegroup, err := newNodegroupFromMachineDeployment(c, machineDeployment)
			if err != nil {
				return nil, fmt.Errorf("failed to build nodegroup for node %q: %v", node.Name, err)
			}
			// We don't scale from 0 so nodes must belong
			// to a nodegroup that has a scale size of at
			// least 1.
			if nodegroup.MaxSize()-nodegroup.MinSize() < 1 {
				return nil, nil
			}
			return nodegroup, nil
		}
	}

	nodegroup, err := newNodegroupFromMachineSet(c, machineSet)
	if err != nil {
		return nil, fmt.Errorf("failed to build nodegroup for node %q: %v", node.Name, err)
	}

	// We don't scale from 0 so nodes must belong to a nodegroup
	// that has a scale size of at least 1.
	if nodegroup.MaxSize()-nodegroup.MinSize() < 1 {
		return nil, nil
	}

	klog.V(4).Infof("node %q is in nodegroup %q", node.Name, machineSet.Name)
	return nodegroup, nil
}

// findNodeByProviderID find the Node object keyed by provideID.
// Returns nil if it cannot be found. A DeepCopy() of the object is
// returned on success.
func (c *machineController) findNodeByProviderID(providerID normalizedProviderID) (*corev1.Node, error) {
	objs, err := c.nodeInformer.GetIndexer().ByIndex(nodeProviderIDIndex, string(providerID))
	if err != nil {
		return nil, err
	}

	switch n := len(objs); {
	case n == 0:
		return nil, nil
	case n > 1:
		return nil, fmt.Errorf("internal error; expected len==1, got %v", n)
	}

	node, ok := objs[0].(*corev1.Node)
	if !ok {
		return nil, fmt.Errorf("internal error; unexpected type %T", objs[0])
	}

	return node.DeepCopy(), nil
}

func (c *machineController) getMachine(namespace, name string, options metav1.GetOptions) (*Machine, error) {
	u, err := c.dynamicclient.Resource(*c.machineResource).Namespace(namespace).Get(context.TODO(), name, options)
	if err != nil {
		return nil, err
	}
	return newMachineFromUnstructured(u.DeepCopy()), nil
}

func (c *machineController) getMachineSet(namespace, name string, options metav1.GetOptions) (*MachineSet, error) {
	u, err := c.dynamicclient.Resource(*c.machineSetResource).Namespace(namespace).Get(context.TODO(), name, options)
	if err != nil {
		return nil, err
	}
	return newMachineSetFromUnstructured(u.DeepCopy()), nil
}

func (c *machineController) getMachineDeployment(namespace, name string, options metav1.GetOptions) (*MachineDeployment, error) {
	u, err := c.dynamicclient.Resource(*c.machineDeploymentResource).Namespace(namespace).Get(context.TODO(), name, options)
	if err != nil {
		return nil, err
	}
	return newMachineDeploymentFromUnstructured(u.DeepCopy()), nil
}

func (c *machineController) listMachines(namespace string, selector labels.Selector) ([]*Machine, error) {
	objs, err := c.machineInformer.Lister().ByNamespace(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	var machines []*Machine

	for _, x := range objs {
		u := x.(*unstructured.Unstructured).DeepCopy()
		if machine := newMachineFromUnstructured(u); machine != nil {
			machines = append(machines, machine)
		}
	}

	return machines, nil
}

func (c *machineController) listMachineSets(namespace string, selector labels.Selector) ([]*MachineSet, error) {
	objs, err := c.machineSetInformer.Lister().ByNamespace(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	var machineSets []*MachineSet

	for _, x := range objs {
		u := x.(*unstructured.Unstructured).DeepCopy()
		if machineSet := newMachineSetFromUnstructured(u); machineSet != nil {
			machineSets = append(machineSets, machineSet)
		}
	}

	return machineSets, nil
}

func (c *machineController) listMachineDeployments(namespace string, selector labels.Selector) ([]*MachineDeployment, error) {
	objs, err := c.machineDeploymentInformer.Lister().ByNamespace(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	var machineDeployments []*MachineDeployment

	for _, x := range objs {
		u := x.(*unstructured.Unstructured).DeepCopy()
		if machineDeployment := newMachineDeploymentFromUnstructured(u); machineDeployment != nil {
			machineDeployments = append(machineDeployments, machineDeployment)
		}
	}

	return machineDeployments, nil
}
