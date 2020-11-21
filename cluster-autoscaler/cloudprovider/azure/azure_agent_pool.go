/*
Copyright 2017 The Kubernetes Authors.

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

package azure

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	azStorage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest/to"

	apiv1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	clusterAutoscalerDeploymentPrefix = `cluster-autoscaler-`
	defaultMaxDeploymentsCount        = 10
)

// AgentPool implements NodeGroup interface for agent pools deployed by aks-engine.
type AgentPool struct {
	azureRef
	manager *AzureManager

	minSize int
	maxSize int

	template   map[string]interface{}
	parameters map[string]interface{}

	mutex       sync.Mutex
	lastRefresh time.Time
	curSize     int64
}

// NewAgentPool creates a new AgentPool.
func NewAgentPool(spec *dynamic.NodeGroupSpec, az *AzureManager) (*AgentPool, error) {
	as := &AgentPool{
		azureRef: azureRef{
			Name: spec.Name,
		},
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
		manager: az,
		curSize: -1,
	}

	if err := as.initialize(); err != nil {
		return nil, err
	}

	return as, nil
}

func (as *AgentPool) initialize() error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	template, err := as.manager.azClient.deploymentsClient.ExportTemplate(ctx, as.manager.config.ResourceGroup, as.manager.config.Deployment)
	if err != nil {
		klog.Errorf("deploymentsClient.ExportTemplate(%s, %s) failed: %v", as.manager.config.ResourceGroup, as.manager.config.Deployment, err)
		return err
	}

	as.template = template.Template.(map[string]interface{})
	as.parameters = as.manager.config.DeploymentParameters
	return normalizeForK8sVMASScalingUp(as.template)
}

// MinSize returns minimum size of the node group.
func (as *AgentPool) MinSize() int {
	return as.minSize
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (as *AgentPool) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (as *AgentPool) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (as *AgentPool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (as *AgentPool) Autoprovisioned() bool {
	return false
}

// MaxSize returns maximum size of the node group.
func (as *AgentPool) MaxSize() int {
	return as.maxSize
}

// Id returns AgentPool id.
func (as *AgentPool) Id() string {
	return as.Name
}

func (as *AgentPool) getVMsFromCache() ([]compute.VirtualMachine, error) {
	allVMs := as.manager.azureCache.getVirtualMachines()
	if _, exists := allVMs[as.Name]; !exists {
		return []compute.VirtualMachine{}, fmt.Errorf("could not find VMs with poolName: %s", as.Name)
	}
	return allVMs[as.Name], nil
}

// GetVMIndexes gets indexes of all virtual machines belonging to the agent pool.
func (as *AgentPool) GetVMIndexes() ([]int, map[int]string, error) {
	klog.V(6).Infof("GetVMIndexes: starts for as %v", as)

	instances, err := as.getVMsFromCache()
	if err != nil {
		return nil, nil, err
	}
	klog.V(6).Infof("GetVMIndexes: got instances, length = %d", len(instances))

	indexes := make([]int, 0)
	indexToVM := make(map[int]string)
	for _, instance := range instances {
		index, err := GetVMNameIndex(instance.StorageProfile.OsDisk.OsType, *instance.Name)
		if err != nil {
			return nil, nil, err
		}

		indexes = append(indexes, index)
		resourceID, err := convertResourceGroupNameToLower("azure://" + *instance.ID)
		if err != nil {
			return nil, nil, err
		}
		indexToVM[index] = resourceID
	}

	sortedIndexes := sort.IntSlice(indexes)
	sortedIndexes.Sort()
	return sortedIndexes, indexToVM, nil
}

func (as *AgentPool) getCurSize() (int64, error) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	if as.lastRefresh.Add(15 * time.Second).After(time.Now()) {
		return as.curSize, nil
	}

	klog.V(5).Infof("Get agent pool size for %q", as.Name)
	indexes, _, err := as.GetVMIndexes()
	if err != nil {
		return 0, err
	}
	klog.V(5).Infof("Returning agent pool (%q) size: %d\n", as.Name, len(indexes))

	if as.curSize != int64(len(indexes)) {
		klog.V(6).Infof("getCurSize:as.curSize(%d) != real size (%d), invalidating cache", as.curSize, len(indexes))
		as.manager.invalidateCache()
	}

	as.curSize = int64(len(indexes))
	as.lastRefresh = time.Now()
	return as.curSize, nil
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (as *AgentPool) TargetSize() (int, error) {
	size, err := as.getCurSize()
	if err != nil {
		return -1, err
	}

	return int(size), nil
}

func (as *AgentPool) getAllSucceededAndFailedDeployments() (succeededAndFailedDeployments []resources.DeploymentExtended, err error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	deploymentsFilter := "provisioningState eq 'Succeeded' or provisioningState eq 'Failed'"
	succeededAndFailedDeployments, err = as.manager.azClient.deploymentsClient.List(ctx, as.manager.config.ResourceGroup, deploymentsFilter, nil)
	if err != nil {
		klog.Errorf("getAllSucceededAndFailedDeployments: failed to list succeeded or failed deployments with error: %v", err)
		return nil, err
	}

	return succeededAndFailedDeployments, err
}

// deleteOutdatedDeployments keeps the newest deployments in the resource group and delete others,
// since Azure resource group deployments have a hard cap of 800, outdated deployments must be deleted
// to prevent the `DeploymentQuotaExceeded` error. see: issue #2154.
func (as *AgentPool) deleteOutdatedDeployments() (err error) {
	deployments, err := as.getAllSucceededAndFailedDeployments()
	if err != nil {
		return err
	}

	for i := len(deployments) - 1; i >= 0; i-- {
		klog.V(4).Infof("deleteOutdatedDeployments: found deployments[i].Name: %s", *deployments[i].Name)
		if deployments[i].Name != nil && !strings.HasPrefix(*deployments[i].Name, clusterAutoscalerDeploymentPrefix) {
			deployments = append(deployments[:i], deployments[i+1:]...)
		}
	}

	if int64(len(deployments)) <= as.manager.config.MaxDeploymentsCount {
		klog.V(4).Infof("deleteOutdatedDeployments: the number of deployments (%d) is under threshold, skip deleting", len(deployments))
		return err
	}

	sort.Slice(deployments, func(i, j int) bool {
		return deployments[i].Properties.Timestamp.Time.After(deployments[j].Properties.Timestamp.Time)
	})

	toBeDeleted := deployments[as.manager.config.MaxDeploymentsCount:]

	ctx, cancel := getContextWithCancel()
	defer cancel()

	errList := make([]error, 0)
	for _, deployment := range toBeDeleted {
		klog.V(4).Infof("deleteOutdatedDeployments: starts deleting outdated deployment (%s)", *deployment.Name)
		_, err := as.manager.azClient.deploymentsClient.Delete(ctx, as.manager.config.ResourceGroup, *deployment.Name)
		if err != nil {
			errList = append(errList, err)
		}
	}

	return utilerrors.NewAggregate(errList)
}

// IncreaseSize increases agent pool size
func (as *AgentPool) IncreaseSize(delta int) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	if as.curSize == -1 {
		return fmt.Errorf("the availability set %s is under initialization, skipping IncreaseSize", as.Name)
	}

	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	err := as.deleteOutdatedDeployments()
	if err != nil {
		klog.Warningf("IncreaseSize: failed to cleanup outdated deployments with err: %v.", err)
	}

	klog.V(6).Infof("IncreaseSize: invalidating cache")
	as.manager.invalidateCache()

	indexes, _, err := as.GetVMIndexes()
	if err != nil {
		return err
	}

	curSize := len(indexes)
	if curSize+delta > as.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", curSize+delta, as.MaxSize())
	}

	highestUsedIndex := indexes[len(indexes)-1]
	expectedSize := curSize + delta
	countForTemplate := expectedSize
	if highestUsedIndex != 0 {
		countForTemplate += highestUsedIndex + 1 - curSize
	}
	as.parameters[as.Name+"Count"] = map[string]int{"value": countForTemplate}
	as.parameters[as.Name+"Offset"] = map[string]int{"value": highestUsedIndex + 1}

	newDeploymentName := fmt.Sprintf("cluster-autoscaler-%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int31())
	newDeployment := resources.Deployment{
		Properties: &resources.DeploymentProperties{
			Template:   &as.template,
			Parameters: &as.parameters,
			Mode:       resources.Incremental,
		},
	}
	ctx, cancel := getContextWithCancel()
	defer cancel()
	klog.V(3).Infof("Waiting for deploymentsClient.CreateOrUpdate(%s, %s, %v)", as.manager.config.ResourceGroup, newDeploymentName, newDeployment)
	resp, err := as.manager.azClient.deploymentsClient.CreateOrUpdate(ctx, as.manager.config.ResourceGroup, newDeploymentName, newDeployment)
	isSuccess, realError := isSuccessHTTPResponse(resp, err)
	if isSuccess {
		klog.V(3).Infof("deploymentsClient.CreateOrUpdate(%s, %s, %v) success", as.manager.config.ResourceGroup, newDeploymentName, newDeployment)

		// Update cache after scale success.
		as.curSize = int64(expectedSize)
		as.lastRefresh = time.Now()
		klog.V(6).Info("IncreaseSize: invalidating cache")
		as.manager.invalidateCache()
		return nil
	}

	klog.Errorf("deploymentsClient.CreateOrUpdate for deployment %q failed: %v", newDeploymentName, realError)
	return realError
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (as *AgentPool) DecreaseTargetSize(delta int) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	nodes, err := as.getVMsFromCache()
	if err != nil {
		return err
	}

	curTargetSize := int(as.curSize)
	if curTargetSize+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			curTargetSize, delta, len(nodes))
	}

	as.curSize = int64(curTargetSize + delta)
	as.lastRefresh = time.Now()
	return nil
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (as *AgentPool) Belongs(node *apiv1.Node) (bool, error) {
	klog.V(6).Infof("Check if node belongs to this agent pool: AgentPool:%v, node:%v\n", as, node)

	ref := &azureRef{
		Name: node.Spec.ProviderID,
	}

	targetAsg, err := as.manager.GetNodeGroupForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known agent pool", node.Name)
	}
	if !strings.EqualFold(targetAsg.Id(), as.Name) {
		return false, nil
	}
	return true, nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (as *AgentPool) DeleteInstances(instances []*azureRef) error {
	if len(instances) == 0 {
		return nil
	}

	commonAsg, err := as.manager.GetNodeGroupForInstance(instances[0])
	if err != nil {
		return err
	}

	for _, instance := range instances {
		asg, err := as.manager.GetNodeGroupForInstance(instance)
		if err != nil {
			return err
		}

		if !strings.EqualFold(asg.Id(), commonAsg.Id()) {
			return fmt.Errorf("cannot delete instance (%s) which don't belong to the same node pool (%q)", instance.GetKey(), commonAsg)
		}
	}

	for _, instance := range instances {
		name, err := resourceName((*instance).Name)
		if err != nil {
			klog.Errorf("Get name for instance %q failed: %v", *instance, err)
			return err
		}

		err = as.deleteVirtualMachine(name)
		if err != nil {
			klog.Errorf("Delete virtual machine %q failed: %v", name, err)
			return err
		}
	}

	klog.V(6).Infof("DeleteInstances: invalidating cache")
	as.manager.invalidateCache()
	return nil
}

// DeleteNodes deletes the nodes from the group.
func (as *AgentPool) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(6).Infof("Delete nodes requested: %v\n", nodes)
	indexes, _, err := as.GetVMIndexes()
	if err != nil {
		return err
	}

	if len(indexes) <= as.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]*azureRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := as.Belongs(node)
		if err != nil {
			return err
		}

		if belongs != true {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, as.Name)
		}

		ref := &azureRef{
			Name: node.Spec.ProviderID,
		}
		refs = append(refs, ref)
	}

	err = as.deleteOutdatedDeployments()
	if err != nil {
		klog.Warningf("DeleteNodes: failed to cleanup outdated deployments with err: %v.", err)
	}

	return as.DeleteInstances(refs)
}

// Debug returns a debug string for the agent pool.
func (as *AgentPool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", as.Name, as.MinSize(), as.MaxSize())
}

// TemplateNodeInfo returns a node template for this agent pool.
func (as *AgentPool) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Nodes returns a list of all nodes that belong to this node group.
func (as *AgentPool) Nodes() ([]cloudprovider.Instance, error) {
	instances, err := as.getVMsFromCache()
	if err != nil {
		return nil, err
	}

	nodes := make([]cloudprovider.Instance, 0, len(instances))
	for _, instance := range instances {
		if len(*instance.ID) == 0 {
			continue
		}

		// To keep consistent with providerID from kubernetes cloud provider, convert
		// resourceGroupName in the ID to lower case.
		resourceID, err := convertResourceGroupNameToLower("azure://" + *instance.ID)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, cloudprovider.Instance{Id: resourceID})
	}

	return nodes, nil
}

func (as *AgentPool) deleteBlob(accountName, vhdContainer, vhdBlob string) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	storageKeysResult, rerr := as.manager.azClient.storageAccountsClient.ListKeys(ctx, as.manager.config.ResourceGroup, accountName)
	if rerr != nil {
		return rerr.Error()
	}

	keys := *storageKeysResult.Keys
	client, err := azStorage.NewBasicClientOnSovereignCloud(accountName, to.String(keys[0].Value), as.manager.env)
	if err != nil {
		return err
	}

	bs := client.GetBlobService()
	containerRef := bs.GetContainerReference(vhdContainer)
	blobRef := containerRef.GetBlobReference(vhdBlob)

	return blobRef.Delete(&azStorage.DeleteBlobOptions{})
}

// deleteVirtualMachine deletes a VM and any associated OS disk
func (as *AgentPool) deleteVirtualMachine(name string) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	vm, rerr := as.manager.azClient.virtualMachinesClient.Get(ctx, as.manager.config.ResourceGroup, name, "")
	if rerr != nil {
		if exists, _ := checkResourceExistsFromRetryError(rerr); !exists {
			klog.V(2).Infof("VirtualMachine %s/%s has already been removed", as.manager.config.ResourceGroup, name)
			return nil
		}

		klog.Errorf("failed to get VM: %s/%s: %s", as.manager.config.ResourceGroup, name, rerr.Error())
		return rerr.Error()
	}

	vhd := vm.VirtualMachineProperties.StorageProfile.OsDisk.Vhd
	managedDisk := vm.VirtualMachineProperties.StorageProfile.OsDisk.ManagedDisk
	if vhd == nil && managedDisk == nil {
		klog.Errorf("failed to get a valid os disk URI for VM: %s/%s", as.manager.config.ResourceGroup, name)
		return fmt.Errorf("os disk does not have a VHD URI")
	}

	osDiskName := vm.VirtualMachineProperties.StorageProfile.OsDisk.Name
	var nicName string
	nicID := (*vm.VirtualMachineProperties.NetworkProfile.NetworkInterfaces)[0].ID
	if nicID == nil {
		klog.Warningf("NIC ID is not set for VM (%s/%s)", as.manager.config.ResourceGroup, name)
	} else {
		nicName, err := resourceName(*nicID)
		if err != nil {
			return err
		}
		klog.Infof("found nic name for VM (%s/%s): %s", as.manager.config.ResourceGroup, name, nicName)
	}

	klog.Infof("deleting VM: %s/%s", as.manager.config.ResourceGroup, name)
	deleteCtx, deleteCancel := getContextWithCancel()
	defer deleteCancel()

	klog.Infof("waiting for VirtualMachine deletion: %s/%s", as.manager.config.ResourceGroup, name)
	rerr = as.manager.azClient.virtualMachinesClient.Delete(deleteCtx, as.manager.config.ResourceGroup, name)
	_, realErr := checkResourceExistsFromRetryError(rerr)
	if realErr != nil {
		return realErr
	}
	klog.V(2).Infof("VirtualMachine %s/%s removed", as.manager.config.ResourceGroup, name)

	if len(nicName) > 0 {
		klog.Infof("deleting nic: %s/%s", as.manager.config.ResourceGroup, nicName)
		interfaceCtx, interfaceCancel := getContextWithCancel()
		defer interfaceCancel()
		rerr := as.manager.azClient.interfacesClient.Delete(interfaceCtx, as.manager.config.ResourceGroup, nicName)
		klog.Infof("waiting for nic deletion: %s/%s", as.manager.config.ResourceGroup, nicName)
		_, realErr := checkResourceExistsFromRetryError(rerr)
		if realErr != nil {
			return realErr
		}
		klog.V(2).Infof("interface %s/%s removed", as.manager.config.ResourceGroup, nicName)
	}

	if vhd != nil {
		accountName, vhdContainer, vhdBlob, err := splitBlobURI(*vhd.URI)
		if err != nil {
			return err
		}

		klog.Infof("found os disk storage reference: %s %s %s", accountName, vhdContainer, vhdBlob)

		klog.Infof("deleting blob: %s/%s", vhdContainer, vhdBlob)
		if err = as.deleteBlob(accountName, vhdContainer, vhdBlob); err != nil {
			_, realErr := checkResourceExistsFromError(err)
			if realErr != nil {
				return realErr
			}
			klog.V(2).Infof("Blob %s/%s removed", as.manager.config.ResourceGroup, vhdBlob)
		}
	} else if managedDisk != nil {
		if osDiskName == nil {
			klog.Warningf("osDisk is not set for VM %s/%s", as.manager.config.ResourceGroup, name)
		} else {
			klog.Infof("deleting managed disk: %s/%s", as.manager.config.ResourceGroup, *osDiskName)
			disksCtx, disksCancel := getContextWithCancel()
			defer disksCancel()
			rerr := as.manager.azClient.disksClient.Delete(disksCtx, as.manager.config.ResourceGroup, *osDiskName)
			_, realErr := checkResourceExistsFromRetryError(rerr)
			if realErr != nil {
				return realErr
			}
			klog.V(2).Infof("disk %s/%s removed", as.manager.config.ResourceGroup, *osDiskName)
		}
	}

	return nil
}

// getAzureRef gets AzureRef for the as.
func (as *AgentPool) getAzureRef() azureRef {
	return as.azureRef
}
