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
	"fmt"
	"os"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	kubeinformers "k8s.io/client-go/informers"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
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
	machineDeploymentKind         = "MachineDeployment"
	machineSetKind                = "MachineSet"
	machineKind                   = "Machine"
	autoDiscovererTypeClusterAPI  = "clusterapi"
	autoDiscovererClusterNameKey  = "clusterName"
	autoDiscovererNamespaceKey    = "namespace"
)

// machineController watches for Nodes, Machines, MachineSets and
// MachineDeployments as they are added, updated and deleted on the
// cluster. Additionally, it adds indices to the node informers to
// satisfy lookup by node.Spec.ProviderID.
type machineController struct {
	workloadInformerFactory     kubeinformers.SharedInformerFactory
	managementInformerFactory   dynamicinformer.DynamicSharedInformerFactory
	machineDeploymentInformer   informers.GenericInformer
	machineInformer             informers.GenericInformer
	machineSetInformer          informers.GenericInformer
	nodeInformer                cache.SharedIndexInformer
	managementClient            dynamic.Interface
	managementScaleClient       scale.ScalesGetter
	machineSetResource          schema.GroupVersionResource
	machineResource             schema.GroupVersionResource
	machineDeploymentResource   schema.GroupVersionResource
	machineDeploymentsAvailable bool
	accessLock                  sync.Mutex
	autoDiscoverySpecs          []*clusterAPIAutoDiscoveryConfig
}

func indexMachineByProviderID(obj interface{}) ([]string, error) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, nil
	}

	providerID, found, err := unstructured.NestedString(u.UnstructuredContent(), "spec", "providerID")
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

func (c *machineController) findMachine(id string) (*unstructured.Unstructured, error) {
	return c.findResourceByKey(c.machineInformer.Informer().GetStore(), id)
}

func (c *machineController) findMachineSet(id string) (*unstructured.Unstructured, error) {
	return c.findResourceByKey(c.machineSetInformer.Informer().GetStore(), id)
}

func (c *machineController) findMachineDeployment(id string) (*unstructured.Unstructured, error) {
	return c.findResourceByKey(c.machineDeploymentInformer.Informer().GetStore(), id)
}

func (c *machineController) findResourceByKey(store cache.Store, key string) (*unstructured.Unstructured, error) {
	item, exists, err := store.GetByKey(key)
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

	// Verify the resource is allowed by the autodiscovery configuration
	if !c.allowedByAutoDiscoverySpecs(u) {
		return nil, nil
	}

	return u.DeepCopy(), nil
}

// findMachineOwner returns the machine set owner for machine, or nil
// if there is no owner. A DeepCopy() of the object is returned on
// success.
func (c *machineController) findMachineOwner(machine *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	machineOwnerRef := machineOwnerRef(machine)
	if machineOwnerRef == nil {
		return nil, nil
	}

	return c.findMachineSet(fmt.Sprintf("%s/%s", machine.GetNamespace(), machineOwnerRef.Name))
}

// findMachineSetOwner returns the owner for the machineSet, or nil
// if there is no owner. A DeepCopy() of the object is returned on
// success.
func (c *machineController) findMachineSetOwner(machineSet *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	machineSetOwnerRef := machineSetOwnerRef(machineSet)
	if machineSetOwnerRef == nil {
		return nil, nil
	}

	return c.findMachineDeployment(fmt.Sprintf("%s/%s", machineSet.GetNamespace(), machineSetOwnerRef.Name))
}

// run starts shared informers and waits for the informer cache to
// synchronize.
func (c *machineController) run(stopCh <-chan struct{}) error {
	c.workloadInformerFactory.Start(stopCh)
	c.managementInformerFactory.Start(stopCh)

	syncFuncs := []cache.InformerSynced{
		c.nodeInformer.HasSynced,
		c.machineInformer.Informer().HasSynced,
		c.machineSetInformer.Informer().HasSynced,
	}
	if c.machineDeploymentsAvailable {
		syncFuncs = append(syncFuncs, c.machineDeploymentInformer.Informer().HasSynced)
	}

	klog.V(4).Infof("waiting for caches to sync")
	if !cache.WaitForCacheSync(stopCh, syncFuncs...) {
		return fmt.Errorf("syncing caches failed")
	}

	return nil
}

func (c *machineController) findScalableResourceByProviderID(providerID normalizedProviderID) (*unstructured.Unstructured, error) {
	machine, err := c.findMachineByProviderID(providerID)
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

	if c.machineDeploymentsAvailable {
		machineDeployment, err := c.findMachineSetOwner(machineSet)
		if err != nil {
			return nil, err
		}

		// If a matching machineDeployment was found return it
		if machineDeployment != nil {
			return machineDeployment, nil
		}
	}

	return machineSet, nil
}

// findMachineByProviderID finds machine matching providerID. A
// DeepCopy() of the object is returned on success.
func (c *machineController) findMachineByProviderID(providerID normalizedProviderID) (*unstructured.Unstructured, error) {
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
		return u.DeepCopy(), nil
	}

	if isFailedMachineProviderID(providerID) {
		return c.findMachine(machineKeyFromFailedProviderID(providerID))
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

	machineID, ok := node.Annotations[machineAnnotationKey]
	if !ok {
		machineID = node.Annotations[deprecatedMachineAnnotationKey]
	}
	return c.findMachine(machineID)
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
	managementClient dynamic.Interface,
	workloadClient kubeclient.Interface,
	managementDiscoveryClient discovery.DiscoveryInterface,
	managementScaleClient scale.ScalesGetter,
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions,
) (*machineController, error) {
	workloadInformerFactory := kubeinformers.NewSharedInformerFactory(workloadClient, 0)
	managementInformerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(managementClient, 0, metav1.NamespaceAll, nil)

	autoDiscoverySpecs, err := parseAutoDiscovery(discoveryOpts.NodeGroupAutoDiscoverySpecs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auto discovery configuration: %v", err)
	}

	CAPIGroup := getCAPIGroup()
	CAPIVersion, err := getAPIGroupPreferredVersion(managementDiscoveryClient, CAPIGroup)
	if err != nil {
		return nil, fmt.Errorf("could not find preferred version for CAPI group %q: %v", CAPIGroup, err)
	}
	klog.Infof("Using version %q for API group %q", CAPIVersion, CAPIGroup)

	var gvrMachineDeployment schema.GroupVersionResource
	var machineDeploymentInformer informers.GenericInformer

	machineDeploymentAvailable, err := groupVersionHasResource(managementDiscoveryClient,
		fmt.Sprintf("%s/%s", CAPIGroup, CAPIVersion), resourceNameMachineDeployment)
	if err != nil {
		return nil, fmt.Errorf("failed to validate if resource %q is available for group %q: %v",
			resourceNameMachineDeployment, fmt.Sprintf("%s/%s", CAPIGroup, CAPIVersion), err)
	}

	if machineDeploymentAvailable {
		gvrMachineDeployment = schema.GroupVersionResource{
			Group:    CAPIGroup,
			Version:  CAPIVersion,
			Resource: resourceNameMachineDeployment,
		}
		machineDeploymentInformer = managementInformerFactory.ForResource(gvrMachineDeployment)
		machineDeploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})
	}

	gvrMachineSet := schema.GroupVersionResource{
		Group:    CAPIGroup,
		Version:  CAPIVersion,
		Resource: resourceNameMachineSet,
	}
	machineSetInformer := managementInformerFactory.ForResource(gvrMachineSet)
	machineSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	gvrMachine := schema.GroupVersionResource{
		Group:    CAPIGroup,
		Version:  CAPIVersion,
		Resource: resourceNameMachine,
	}
	machineInformer := managementInformerFactory.ForResource(gvrMachine)
	machineInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	nodeInformer := workloadInformerFactory.Core().V1().Nodes().Informer()
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
		autoDiscoverySpecs:          autoDiscoverySpecs,
		workloadInformerFactory:     workloadInformerFactory,
		managementInformerFactory:   managementInformerFactory,
		machineDeploymentInformer:   machineDeploymentInformer,
		machineInformer:             machineInformer,
		machineSetInformer:          machineSetInformer,
		nodeInformer:                nodeInformer,
		managementClient:            managementClient,
		managementScaleClient:       managementScaleClient,
		machineSetResource:          gvrMachineSet,
		machineResource:             gvrMachine,
		machineDeploymentResource:   gvrMachineDeployment,
		machineDeploymentsAvailable: machineDeploymentAvailable,
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

func (c *machineController) scalableResourceProviderIDs(scalableResource *unstructured.Unstructured) ([]string, error) {
	machines, err := c.listMachinesForScalableResource(scalableResource)
	if err != nil {
		return nil, fmt.Errorf("error listing machines: %v", err)
	}

	var providerIDs []string
	for _, machine := range machines {
		providerID, found, err := unstructured.NestedString(machine.UnstructuredContent(), "spec", "providerID")
		if err != nil {
			return nil, err
		}

		if found {
			if providerID != "" {
				providerIDs = append(providerIDs, providerID)
				continue
			}
		}

		klog.Warningf("Machine %q has no providerID", machine.GetName())

		failureMessage, found, err := unstructured.NestedString(machine.UnstructuredContent(), "status", "failureMessage")
		if err != nil {
			return nil, err
		}

		if found {
			klog.V(4).Infof("Status.FailureMessage of machine %q is %q", machine.GetName(), failureMessage)
			// Provide a fake ID to allow the autoscaler to track machines that will never
			// become nodes and mark the nodegroup unhealthy after maxNodeProvisionTime.
			// Fake ID needs to be recognised later and converted into a machine key.
			// Use an underscore as a separator between namespace and name as it is not a
			// valid character within a namespace name.
			providerIDs = append(providerIDs, fmt.Sprintf("%s%s_%s", failedMachinePrefix, machine.GetNamespace(), machine.GetName()))
			continue
		}

		_, found, err = unstructured.NestedFieldCopy(machine.UnstructuredContent(), "status", "nodeRef")
		if err != nil {
			return nil, err
		}

		if !found {
			klog.V(4).Infof("Status.NodeRef of machine %q is currently nil", machine.GetName())
			continue
		}

		nodeRefKind, found, err := unstructured.NestedString(machine.UnstructuredContent(), "status", "nodeRef", "kind")
		if err != nil {
			return nil, err
		}

		if found && nodeRefKind != "Node" {
			klog.Errorf("Status.NodeRef of machine %q does not reference a node (rather %q)", machine.GetName(), nodeRefKind)
			continue
		}

		nodeRefName, found, err := unstructured.NestedString(machine.UnstructuredContent(), "status", "nodeRef", "name")
		if err != nil {
			return nil, err
		}

		if found {
			node, err := c.findNodeByNodeName(nodeRefName)
			if err != nil {
				return nil, fmt.Errorf("unknown node %q", nodeRefName)
			}

			if node != nil {
				providerIDs = append(providerIDs, node.Spec.ProviderID)
			}
		}
	}

	klog.V(4).Infof("nodegroup %s has nodes %v", scalableResource.GetName(), providerIDs)

	return providerIDs, nil
}

func (c *machineController) nodeGroups() ([]*nodegroup, error) {
	scalableResources, err := c.listScalableResources()
	if err != nil {
		return nil, err
	}

	nodegroups := make([]*nodegroup, 0, len(scalableResources))

	for _, r := range scalableResources {
		ng, err := newNodeGroupFromScalableResource(c, r)
		if err != nil {
			return nil, err
		}

		if ng != nil {
			nodegroups = append(nodegroups, ng)
		}
	}
	return nodegroups, nil
}

func (c *machineController) nodeGroupForNode(node *corev1.Node) (*nodegroup, error) {
	scalableResource, err := c.findScalableResourceByProviderID(normalizedProviderString(node.Spec.ProviderID))
	if err != nil {
		return nil, err
	}
	if scalableResource == nil {
		return nil, nil
	}

	nodegroup, err := newNodeGroupFromScalableResource(c, scalableResource)
	if err != nil {
		return nil, fmt.Errorf("failed to build nodegroup for node %q: %v", node.Name, err)
	}

	// the nodegroup will be nil if it doesn't match the autodiscovery configuration
	// or if it doesn't meet the scaling requirements
	if nodegroup == nil {
		return nil, nil
	}

	klog.V(4).Infof("node %q is in nodegroup %q", node.Name, nodegroup.Id())
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

func (c *machineController) listMachinesForScalableResource(r *unstructured.Unstructured) ([]*unstructured.Unstructured, error) {
	switch r.GetKind() {
	case machineSetKind, machineDeploymentKind:
		unstructuredSelector, found, err := unstructured.NestedMap(r.UnstructuredContent(), "spec", "selector")
		if err != nil {
			return nil, err
		}

		if !found {
			return nil, fmt.Errorf("expected field spec.selector on scalable resource type")
		}

		labelSelector := &metav1.LabelSelector{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredSelector, labelSelector); err != nil {
			return nil, err
		}

		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, err
		}

		return listResources(c.machineInformer.Lister().ByNamespace(r.GetNamespace()), clusterNameFromResource(r), selector)
	default:
		return nil, fmt.Errorf("unknown scalable resource kind %s", r.GetKind())
	}
}

func (c *machineController) listScalableResources() ([]*unstructured.Unstructured, error) {
	scalableResources, err := c.listResources(c.machineSetInformer.Lister())
	if err != nil {
		return nil, err
	}

	if c.machineDeploymentsAvailable {
		machineDeployments, err := c.listResources(c.machineDeploymentInformer.Lister())
		if err != nil {
			return nil, err
		}

		scalableResources = append(scalableResources, machineDeployments...)
	}
	return scalableResources, nil
}

func (c *machineController) listResources(lister cache.GenericLister) ([]*unstructured.Unstructured, error) {
	if len(c.autoDiscoverySpecs) == 0 {
		return listResources(lister.ByNamespace(metav1.NamespaceAll), "", labels.Everything())
	}

	var results []*unstructured.Unstructured
	tracker := map[string]bool{}
	for _, spec := range c.autoDiscoverySpecs {
		resources, err := listResources(lister.ByNamespace(spec.namespace), spec.clusterName, spec.labelSelector)
		if err != nil {
			return nil, err
		}
		for i := range resources {
			r := resources[i]
			key := fmt.Sprintf("%s-%s-%s", r.GetKind(), r.GetNamespace(), r.GetName())
			if _, ok := tracker[key]; !ok {
				results = append(results, r)
				tracker[key] = true
			}
		}
	}

	return results, nil
}

func listResources(lister cache.GenericNamespaceLister, clusterName string, selector labels.Selector) ([]*unstructured.Unstructured, error) {
	objs, err := lister.List(selector)
	if err != nil {
		return nil, err
	}

	results := make([]*unstructured.Unstructured, 0, len(objs))
	for _, x := range objs {
		u, ok := x.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("expected unstructured resource from lister, not %T", x)
		}

		// if clusterName is not empty and the clusterName does not match the resource, do not return it as part of the results
		if clusterName != "" && clusterNameFromResource(u) != clusterName {
			continue
		}

		// if we are listing MachineSets, do not return MachineSets that are owned by a MachineDeployment
		if u.GetKind() == machineSetKind && machineSetHasMachineDeploymentOwnerRef(u) {
			continue
		}

		results = append(results, u.DeepCopy())
	}

	return results, nil
}

func (c *machineController) allowedByAutoDiscoverySpecs(r *unstructured.Unstructured) bool {
	// If no autodiscovery configuration fall back to previous behavior of allowing all
	if len(c.autoDiscoverySpecs) == 0 {
		return true
	}

	for _, spec := range c.autoDiscoverySpecs {
		if allowedByAutoDiscoverySpec(spec, r) {
			return true
		}
	}

	return false
}
