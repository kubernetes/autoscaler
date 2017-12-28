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

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	azStorage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// AgentPool implements NodeGroup interface for agent pools deployed by acs-engine.
type AgentPool struct {
	AzureRef
	*AzureManager

	minSize int
	maxSize int

	template   map[string]interface{}
	parameters map[string]interface{}

	mutex        sync.Mutex
	targetSize   int
	provisioning bool
}

// VirtualMachineID contains VMID and ID of a virtual machine.
type VirtualMachineID struct {
	ID   string
	VMID string
}

// NewAgentPool creates a new AgentPool.
func NewAgentPool(name string, minSize, maxSize int, az *AzureManager) (*AgentPool, error) {
	as := &AgentPool{
		AzureRef: AzureRef{
			Name: name,
		},
		minSize:      minSize,
		maxSize:      maxSize,
		targetSize:   -1,
		AzureManager: az,
	}

	if err := as.initialize(); err != nil {
		return nil, err
	}

	return as, nil
}

func (as *AgentPool) initialize() error {
	deploy, err := as.deploymentsClient.Get(as.config.ResourceGroup, as.config.Deployment)
	if err != nil {
		glog.Errorf("deploymentsClient.Get(%s, %s) failed: %v", as.config.ResourceGroup, as.config.Deployment, err)
		return err
	}

	template, err := as.deploymentsClient.ExportTemplate(as.config.ResourceGroup, as.config.Deployment)
	if err != nil {
		glog.Errorf("deploymentsClient.ExportTemplate(%s, %s) failed: %v", as.config.ResourceGroup, as.config.Deployment, err)
		return err
	}

	as.parameters = *deploy.Properties.Parameters
	as.preprocessParameters()

	as.template = *template.Template
	return normalizeForK8sVMASScalingUp(as.template)
}

func (as *AgentPool) preprocessParameters() {
	// Delete type key from parameters.
	for k := range as.parameters {
		if v, ok := as.parameters[k].(map[string]interface{}); ok {
			delete(v, "type")
		}
	}

	// fulfill secure parameters.
	as.parameters["apiServerPrivateKey"] = map[string]string{"value": as.config.APIServerPrivateKey}
	as.parameters["caPrivateKey"] = map[string]string{"value": as.config.CAPrivateKey}
	as.parameters["clientPrivateKey"] = map[string]string{"value": as.config.ClientPrivateKey}
	as.parameters["kubeConfigPrivateKey"] = map[string]string{"value": as.config.KubeConfigPrivateKey}
	as.parameters["servicePrincipalClientId"] = map[string]string{"value": as.config.AADClientID}
	as.parameters["servicePrincipalClientSecret"] = map[string]string{"value": as.config.AADClientSecret}
	if as.config.WindowsAdminPassword != "" {
		as.parameters["windowsAdminPassword"] = map[string]string{"value": as.config.WindowsAdminPassword}
	}
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
func (as *AgentPool) Create() error {
	return cloudprovider.ErrAlreadyExist
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

// GetVMIndexes gets indexes of all virtual machines belongting to the agent pool.
func (as *AgentPool) GetVMIndexes() ([]int, map[int]VirtualMachineID, error) {
	instances, err := as.GetVirtualMachines()
	if err != nil {
		return nil, nil, err
	}

	indexes := make([]int, 0)
	indexToVM := make(map[int]VirtualMachineID)
	for _, instance := range instances {
		index, err := GetVMNameIndex(instance.StorageProfile.OsDisk.OsType, *instance.Name)
		if err != nil {
			return nil, nil, err
		}

		indexes = append(indexes, index)
		indexToVM[index] = VirtualMachineID{
			ID:   "azure://" + strings.ToLower(*instance.ID),
			VMID: "azure://" + strings.ToLower(*instance.VMID),
		}
	}

	sortedIndexes := sort.IntSlice(indexes)
	sortedIndexes.Sort()
	return sortedIndexes, indexToVM, nil
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (as *AgentPool) TargetSize() (int, error) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	if as.targetSize != -1 {
		return as.targetSize, nil
	}

	indexes, _, err := as.GetVMIndexes()
	if err != nil {
		return 0, err
	}

	as.targetSize = len(indexes)
	return as.targetSize, nil
}

// IncreaseSize increases agent pool size
func (as *AgentPool) IncreaseSize(delta int) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	indexes, _, err := as.GetVMIndexes()
	if err != nil {
		return err
	}

	curSize := len(indexes)
	if curSize+delta > as.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", curSize+delta, as.MaxSize())
	}

	highestUsedIndex := indexes[len(indexes)-1]
	countForTemplate := curSize + delta
	as.targetSize = countForTemplate
	if highestUsedIndex != 0 {
		countForTemplate += highestUsedIndex + 1 - curSize
	}
	as.parameters[as.Name+"Count"] = map[string]int{"value": countForTemplate}
	as.parameters[as.Name+"Offset"] = map[string]int{"value": highestUsedIndex + 1}

	cancel := make(chan struct{})
	newDeploymentName := fmt.Sprintf("cluster-autoscaler-%d", rand.New(rand.NewSource(time.Now().UnixNano())).Int31())
	newDeployment := resources.Deployment{
		Properties: &resources.DeploymentProperties{
			Template:   &as.template,
			Parameters: &as.parameters,
			Mode:       resources.Incremental,
		},
	}
	_, errChan := as.deploymentsClient.CreateOrUpdate(as.config.ResourceGroup, newDeploymentName, newDeployment, cancel)
	glog.V(3).Infof("Waiting for deploymentsClient.CreateOrUpdate(%s, %s, %s)", as.config.ResourceGroup, newDeploymentName, newDeployment)
	return <-errChan
}

// GetVirtualMachines returns list of nodes for the given agent pool.
func (as *AgentPool) GetVirtualMachines() (instances []compute.VirtualMachine, err error) {
	result, err := as.virtualMachinesClient.List(as.config.ResourceGroup)
	if err != nil {
		return nil, err
	}

	moreResult := (result.Value != nil && len(*result.Value) > 0)
	for moreResult {
		for _, instance := range *result.Value {
			if instance.Tags == nil {
				continue
			}

			tags := *instance.Tags
			vmPoolName := tags["poolName"]
			if *vmPoolName != as.Id() {
				continue
			}

			instances = append(instances, instance)
		}

		moreResult = false
		if result.NextLink != nil {
			result, err = as.virtualMachinesClient.ListNextResults(result)
			if err != nil {
				glog.Errorf("virtualMachinesClient.ListNextResults failed: %v", err)
				return nil, err
			}

			moreResult = (result.Value != nil && len(*result.Value) > 0)
		}
	}

	return instances, nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (as *AgentPool) DecreaseTargetSize(delta int) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	nodes, err := as.GetVirtualMachines()
	if err != nil {
		return err
	}

	curTargetSize := as.targetSize
	if curTargetSize+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			curTargetSize, delta, len(nodes))
	}

	as.targetSize = curTargetSize + delta
	return nil
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (as *AgentPool) Belongs(node *apiv1.Node) (bool, error) {
	glog.V(6).Infof("Check if node belongs to this agent pool: AgentPool:%v, node:%v\n", as, node)

	ref := &AzureRef{
		Name: strings.ToLower(node.Spec.ProviderID),
	}

	targetAsg, err := as.GetNodeGroupForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known agent pool", node.Name)
	}
	if targetAsg.Id() != as.Id() {
		return false, nil
	}
	return true, nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (as *AgentPool) DeleteInstances(instances []*AzureRef) error {
	if len(instances) == 0 {
		return nil
	}

	commonAsg, err := as.GetNodeGroupForInstance(instances[0])
	if err != nil {
		return err
	}

	for _, instance := range instances {
		asg, err := as.GetNodeGroupForInstance(instance)
		if err != nil {
			return err
		}

		if asg != commonAsg {
			return fmt.Errorf("cannot delete instance (%s) which don't belong to the same node pool (%q)", instance.GetKey(), commonAsg)
		}
	}

	for _, instance := range instances {
		name, err := resourceName((*instance).Name)
		if err != nil {
			glog.Errorf("Get name for instance %q failed: %v", *instance, err)
			return err
		}

		err = as.deleteVirtualMachine(name)
		if err != nil {
			glog.Errorf("Delete virtual machine %q failed: %v", name, err)
			return err
		}
	}

	return nil
}

// DeleteNodes deletes the nodes from the group.
func (as *AgentPool) DeleteNodes(nodes []*apiv1.Node) error {
	glog.V(8).Infof("Delete nodes requested: %v\n", nodes)
	indexes, _, err := as.GetVMIndexes()
	if err != nil {
		return err
	}

	if len(indexes) <= as.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]*AzureRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := as.Belongs(node)
		if err != nil {
			return err
		}

		if belongs != true {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, as.Id())
		}

		azureRef := &AzureRef{
			Name: strings.ToLower(node.Spec.ProviderID),
		}
		refs = append(refs, azureRef)
	}

	return as.DeleteInstances(refs)
}

// Id returns AgentPool id.
func (as *AgentPool) Id() string {
	return as.Name
}

// Debug returns a debug string for the agent pool.
func (as *AgentPool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", as.Id(), as.MinSize(), as.MaxSize())
}

// TemplateNodeInfo returns a node template for this agent pool.
func (as *AgentPool) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Nodes returns a list of all nodes that belong to this node group.
func (as *AgentPool) Nodes() ([]string, error) {
	instances, err := as.GetVirtualMachines()
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0, len(instances))
	for _, instance := range instances {
		// Convert to lower because instance.ID is in different in different API calls (e.g. GET and LIST).
		name := "azure://" + strings.ToLower(*instance.ID)
		nodes = append(nodes, name)
	}

	return nodes, nil
}

func (as *AgentPool) deleteBlob(accountName, vhdContainer, vhdBlob string) error {
	storageKeysResult, err := as.storageAccountsClient.ListKeys(as.config.ResourceGroup, accountName)
	if err != nil {
		return err
	}

	keys := *storageKeysResult.Keys
	client, err := azStorage.NewBasicClientOnSovereignCloud(accountName, to.String(keys[0].Value), as.env)
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
	vm, err := as.virtualMachinesClient.Get(as.config.ResourceGroup, name, "")
	if err != nil {
		glog.Errorf("failed to get VM: %s/%s: %s", as.config.ResourceGroup, name, err.Error())
		return err
	}

	vhd := vm.VirtualMachineProperties.StorageProfile.OsDisk.Vhd
	managedDisk := vm.VirtualMachineProperties.StorageProfile.OsDisk.ManagedDisk
	if vhd == nil && managedDisk == nil {
		glog.Errorf("failed to get a valid os disk URI for VM: %s/%s", as.config.ResourceGroup, name)
		return fmt.Errorf("os disk does not have a VHD URI")
	}

	osDiskName := vm.VirtualMachineProperties.StorageProfile.OsDisk.Name
	var nicName string
	nicID := (*vm.VirtualMachineProperties.NetworkProfile.NetworkInterfaces)[0].ID
	if nicID == nil {
		glog.Warningf("NIC ID is not set for VM (%s/%s)", as.config.ResourceGroup, name)
	} else {
		nicName, err = resourceName(*nicID)
		if err != nil {
			return err
		}
		glog.Infof("found nic name for VM (%s/%s): %s", as.config.ResourceGroup, name, nicName)
	}
	glog.Infof("deleting VM: %s/%s", as.config.ResourceGroup, name)
	_, deleteErrChan := as.virtualMachinesClient.Delete(as.config.ResourceGroup, name, nil)
	glog.Infof("waiting for vm deletion: %s/%s", as.config.ResourceGroup, name)
	if err := <-deleteErrChan; err != nil {
		return err
	}

	if len(nicName) > 0 {
		glog.Infof("deleting nic: %s/%s", as.config.ResourceGroup, nicName)
		_, nicErrChan := as.interfacesClient.Delete(as.config.ResourceGroup, nicName, nil)
		glog.Infof("waiting for nic deletion: %s/%s", as.config.ResourceGroup, nicName)
		if nicErr := <-nicErrChan; nicErr != nil {
			return nicErr
		}
	}

	if vhd != nil {
		accountName, vhdContainer, vhdBlob, err := splitBlobURI(*vhd.URI)
		if err != nil {
			return err
		}

		glog.Infof("found os disk storage reference: %s %s %s", accountName, vhdContainer, vhdBlob)

		glog.Infof("deleting blob: %s/%s", vhdContainer, vhdBlob)
		if err = as.deleteBlob(accountName, vhdContainer, vhdBlob); err != nil {
			return err
		}
	} else if managedDisk != nil {
		if osDiskName == nil {
			glog.Warningf("osDisk is not set for VM %s/%s", as.config.ResourceGroup, name)
		} else {
			glog.Infof("deleting managed disk: %s/%s", as.config.ResourceGroup, *osDiskName)
			_, diskErrChan := as.disksClient.Delete(as.config.ResourceGroup, *osDiskName, nil)

			if err := <-diskErrChan; err != nil {
				return err
			}
		}
	}

	return nil
}
