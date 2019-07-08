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
	"net/http"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	cloudvolume "k8s.io/cloud-provider/volume"
	"k8s.io/klog"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-10-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

var (
	vmssSizeRefreshPeriod      = 15 * time.Second
	vmssInstancesRefreshPeriod = 5 * time.Minute
)

func init() {
	// In go-autorest SDK https://github.com/Azure/go-autorest/blob/master/autorest/sender.go#L242,
	// if ARM returns http.StatusTooManyRequests, the sender doesn't increase the retry attempt count,
	// hence the Azure clients will keep retrying forever until it get a status code other than 429.
	// So we explicitly removes http.StatusTooManyRequests from autorest.StatusCodesForRetry.
	// Refer https://github.com/Azure/go-autorest/issues/398.
	statusCodesForRetry := make([]int, 0)
	for _, code := range autorest.StatusCodesForRetry {
		if code != http.StatusTooManyRequests {
			statusCodesForRetry = append(statusCodesForRetry, code)
		}
	}
	autorest.StatusCodesForRetry = statusCodesForRetry
}

// ScaleSet implements NodeGroup interface.
type ScaleSet struct {
	azureRef
	manager *AzureManager

	minSize int
	maxSize int

	sizeMutex       sync.Mutex
	curSize         int64
	lastSizeRefresh time.Time

	instanceMutex       sync.Mutex
	instanceCache       []cloudprovider.Instance
	lastInstanceRefresh time.Time
}

// NewScaleSet creates a new NewScaleSet.
func NewScaleSet(spec *dynamic.NodeGroupSpec, az *AzureManager) (*ScaleSet, error) {
	scaleSet := &ScaleSet{
		azureRef: azureRef{
			Name: spec.Name,
		},
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
		manager: az,
		curSize: -1,
	}

	return scaleSet, nil
}

// MinSize returns minimum size of the node group.
func (scaleSet *ScaleSet) MinSize() int {
	return scaleSet.minSize
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (scaleSet *ScaleSet) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (scaleSet *ScaleSet) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (scaleSet *ScaleSet) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (scaleSet *ScaleSet) Autoprovisioned() bool {
	return false
}

// MaxSize returns maximum size of the node group.
func (scaleSet *ScaleSet) MaxSize() int {
	return scaleSet.maxSize
}

func (scaleSet *ScaleSet) getVMSSInfo() (compute.VirtualMachineScaleSet, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	resourceGroup := scaleSet.manager.config.ResourceGroup
	setInfo, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.Get(ctx, resourceGroup, scaleSet.Name)
	if err != nil {
		return compute.VirtualMachineScaleSet{}, err
	}

	return setInfo, nil
}

func (scaleSet *ScaleSet) getCurSize() (int64, error) {
	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	if scaleSet.lastSizeRefresh.Add(vmssSizeRefreshPeriod).After(time.Now()) {
		return scaleSet.curSize, nil
	}

	klog.V(5).Infof("Get scale set size for %q", scaleSet.Name)
	set, err := scaleSet.getVMSSInfo()
	if err != nil {
		if isAzureRequestsThrottled(err) {
			// Log a warning and update the size refresh time so that it would retry after next vmssSizeRefreshPeriod.
			klog.Warningf("getVMSSInfo() is throttled with message %v, would return the cached vmss size", err)
			scaleSet.lastSizeRefresh = time.Now()
			return scaleSet.curSize, nil
		}
		return -1, err
	}
	klog.V(5).Infof("Getting scale set (%q) capacity: %d\n", scaleSet.Name, *set.Sku.Capacity)

	if scaleSet.curSize != *set.Sku.Capacity {
		// Invalidate the instance cache if the capacity has changed.
		scaleSet.invalidateInstanceCache()
	}

	scaleSet.curSize = *set.Sku.Capacity
	scaleSet.lastSizeRefresh = time.Now()
	return scaleSet.curSize, nil
}

// GetScaleSetSize gets Scale Set size.
func (scaleSet *ScaleSet) GetScaleSetSize() (int64, error) {
	return scaleSet.getCurSize()
}

// updateVMSSCapacity invokes virtualMachineScaleSetsClient to update the capacity for VMSS.
func (scaleSet *ScaleSet) updateVMSSCapacity(size int64) {
	var op compute.VirtualMachineScaleSet
	var resp *http.Response
	var isSuccess bool
	var err error

	defer func() {
		if err != nil {
			klog.Errorf("Failed to update the capacity for vmss %s with error %v, invalidate the cache so as to get the real size from API", scaleSet.Name, err)
			// Invalidate the VMSS size cache in order to fetch the size from the API.
			scaleSet.sizeMutex.Lock()
			defer scaleSet.sizeMutex.Unlock()
			scaleSet.lastSizeRefresh = time.Now().Add(-1 * vmssSizeRefreshPeriod)
		}
	}()

	resourceGroup := scaleSet.manager.config.ResourceGroup
	op, err = scaleSet.getVMSSInfo()
	if err != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, err)
		return
	}

	op.Sku.Capacity = &size
	op.Identity = nil
	op.VirtualMachineScaleSetProperties.ProvisioningState = nil
	ctx, cancel := getContextWithCancel()
	defer cancel()
	klog.V(3).Infof("Waiting for virtualMachineScaleSetsClient.CreateOrUpdate(%s)", scaleSet.Name)
	resp, err = scaleSet.manager.azClient.virtualMachineScaleSetsClient.CreateOrUpdate(ctx, resourceGroup, scaleSet.Name, op)
	isSuccess, err = isSuccessHTTPResponse(resp, err)
	if isSuccess {
		klog.V(3).Infof("virtualMachineScaleSetsClient.CreateOrUpdate(%s) success", scaleSet.Name)
		scaleSet.invalidateInstanceCache()
		return
	}

	klog.Errorf("virtualMachineScaleSetsClient.CreateOrUpdate for scale set %q failed: %v", scaleSet.Name, err)
	return
}

// SetScaleSetSize sets ScaleSet size.
func (scaleSet *ScaleSet) SetScaleSetSize(size int64) {
	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	// Proactively set the VMSS size so autoscaler makes better decisions.
	scaleSet.curSize = size
	scaleSet.lastSizeRefresh = time.Now()
	go scaleSet.updateVMSSCapacity(size)
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (scaleSet *ScaleSet) TargetSize() (int, error) {
	size, err := scaleSet.GetScaleSetSize()
	return int(size), err
}

// IncreaseSize increases Scale Set size
func (scaleSet *ScaleSet) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	if int(size)+delta > scaleSet.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, scaleSet.MaxSize())
	}

	scaleSet.SetScaleSetSize(size + int64(delta))
	return nil
}

// GetScaleSetVms returns list of nodes for the given scale set.
// Note that the list results is not used directly because their resource ID format
// is not consistent with Get results.
func (scaleSet *ScaleSet) GetScaleSetVms() ([]string, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	resourceGroup := scaleSet.manager.config.ResourceGroup
	vmList, err := scaleSet.manager.azClient.virtualMachineScaleSetVMsClient.List(ctx, resourceGroup, scaleSet.Name, "", "", "")
	if err != nil {
		klog.Errorf("VirtualMachineScaleSetVMsClient.List failed for %s: %v", scaleSet.Name, err)
		return nil, err
	}

	allVMs := make([]string, 0)
	for _, vm := range vmList {
		// The resource ID is empty string, which indicates the instance may be in deleting state.
		if len(*vm.ID) == 0 {
			continue
		}

		resourceID, err := convertResourceGroupNameToLower(*vm.ID)
		if err != nil {
			// This shouldn't happen. Log a waring message for tracking.
			klog.Warningf("GetScaleSetVms.convertResourceGroupNameToLower failed with error: %v", err)
			continue
		}

		allVMs = append(allVMs, resourceID)
	}

	return allVMs, nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (scaleSet *ScaleSet) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}

	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	nodes, err := scaleSet.Nodes()
	if err != nil {
		return err
	}

	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}

	scaleSet.SetScaleSetSize(size + int64(delta))
	return nil
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (scaleSet *ScaleSet) Belongs(node *apiv1.Node) (bool, error) {
	klog.V(6).Infof("Check if node belongs to this scale set: scaleset:%v, node:%v\n", scaleSet, node)

	ref := &azureRef{
		Name: node.Spec.ProviderID,
	}

	targetAsg, err := scaleSet.manager.GetAsgForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known scale set", node.Name)
	}
	if !strings.EqualFold(targetAsg.Id(), scaleSet.Id()) {
		return false, nil
	}
	return true, nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (scaleSet *ScaleSet) DeleteInstances(instances []*azureRef) error {
	if len(instances) == 0 {
		return nil
	}

	klog.V(3).Infof("Deleting vmss instances %+v", instances)

	commonAsg, err := scaleSet.manager.GetAsgForInstance(instances[0])
	if err != nil {
		return err
	}

	instanceIDs := []string{}
	for _, instance := range instances {
		asg, err := scaleSet.manager.GetAsgForInstance(instance)
		if err != nil {
			return err
		}

		if !strings.EqualFold(asg.Id(), commonAsg.Id()) {
			return fmt.Errorf("cannot delete instance (%s) which don't belong to the same Scale Set (%q)", instance.Name, commonAsg)
		}

		instanceID, err := getLastSegment(instance.Name)
		if err != nil {
			klog.Errorf("getLastSegment failed with error: %v", err)
			return err
		}

		instanceIDs = append(instanceIDs, instanceID)
	}

	requiredIds := &compute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIds: &instanceIDs,
	}
	ctx, cancel := getContextWithCancel()
	defer cancel()
	resourceGroup := scaleSet.manager.config.ResourceGroup
	_, err = scaleSet.manager.azClient.virtualMachineScaleSetsClient.DeleteInstances(ctx, resourceGroup, commonAsg.Id(), *requiredIds)
	return err
}

// DeleteNodes deletes the nodes from the group.
func (scaleSet *ScaleSet) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(8).Infof("Delete nodes requested: %q\n", nodes)
	size, err := scaleSet.GetScaleSetSize()
	if err != nil {
		return err
	}

	if int(size) <= scaleSet.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	refs := make([]*azureRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := scaleSet.Belongs(node)
		if err != nil {
			return err
		}

		if belongs != true {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, scaleSet.Id())
		}

		ref := &azureRef{
			Name: node.Spec.ProviderID,
		}
		refs = append(refs, ref)
	}

	return scaleSet.DeleteInstances(refs)
}

// Id returns ScaleSet id.
func (scaleSet *ScaleSet) Id() string {
	return scaleSet.Name
}

// Debug returns a debug string for the Scale Set.
func (scaleSet *ScaleSet) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", scaleSet.Id(), scaleSet.MinSize(), scaleSet.MaxSize())
}

func buildInstanceOS(template compute.VirtualMachineScaleSet) string {
	instanceOS := cloudprovider.DefaultOS
	if template.VirtualMachineProfile != nil && template.VirtualMachineProfile.OsProfile != nil && template.VirtualMachineProfile.OsProfile.WindowsConfiguration != nil {
		instanceOS = "windows"
	}

	return instanceOS
}

func buildGenericLabels(template compute.VirtualMachineScaleSet, nodeName string) map[string]string {
	result := make(map[string]string)

	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = buildInstanceOS(template)
	result[apiv1.LabelInstanceType] = *template.Sku.Name
	result[apiv1.LabelZoneRegion] = strings.ToLower(*template.Location)

	if template.Zones != nil && len(*template.Zones) > 0 {
		failureDomains := make([]string, len(*template.Zones))
		for k, v := range *template.Zones {
			failureDomains[k] = strings.ToLower(*template.Location) + "-" + v
		}

		result[apiv1.LabelZoneFailureDomain] = strings.Join(failureDomains[:], cloudvolume.LabelMultiZoneDelimiter)
	} else {
		result[apiv1.LabelZoneFailureDomain] = "0"
	}

	result[apiv1.LabelHostname] = nodeName
	return result
}

func (scaleSet *ScaleSet) buildNodeFromTemplate(template compute.VirtualMachineScaleSet) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", scaleSet.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	var vmssType *instanceType
	for k := range InstanceTypes {
		if strings.EqualFold(k, *template.Sku.Name) {
			vmssType = InstanceTypes[k]
			break
		}
	}
	if vmssType == nil {
		return nil, fmt.Errorf("instance type %q not supported", *template.Sku.Name)
	}
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(vmssType.VCPU, resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(vmssType.GPU, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(vmssType.MemoryMb*1024*1024, resource.DecimalSI)

	// TODO: set real allocatable.
	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	if template.Tags != nil {
		for k, v := range template.Tags {
			if v != nil {
				node.Labels[k] = *v
			} else {
				node.Labels[k] = ""
			}

		}
	}

	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))
	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

// TemplateNodeInfo returns a node template for this scale set.
func (scaleSet *ScaleSet) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	template, err := scaleSet.getVMSSInfo()
	if err != nil {
		return nil, err
	}

	node, err := scaleSet.buildNodeFromTemplate(template)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulernodeinfo.NewNodeInfo(cloudprovider.BuildKubeProxy(scaleSet.Name))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// Nodes returns a list of all nodes that belong to this node group.
func (scaleSet *ScaleSet) Nodes() ([]cloudprovider.Instance, error) {
	curSize, err := scaleSet.getCurSize()
	if err != nil {
		klog.Errorf("Failed to get current size for vmss %q: %v", scaleSet.Name, err)
		return nil, err
	}

	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	if int64(len(scaleSet.instanceCache)) == curSize &&
		scaleSet.lastInstanceRefresh.Add(vmssInstancesRefreshPeriod).After(time.Now()) {
		return scaleSet.instanceCache, nil
	}

	vms, err := scaleSet.GetScaleSetVms()
	if err != nil {
		if isAzureRequestsThrottled(err) {
			// Log a warning and update the instance refresh time so that it would retry after next vmssInstancesRefreshPeriod.
			klog.Warningf("GetScaleSetVms() is throttled with message %v, would return the cached instances", err)
			scaleSet.lastInstanceRefresh = time.Now()
			return scaleSet.instanceCache, nil
		}
		return nil, err
	}

	instances := make([]cloudprovider.Instance, len(vms))
	for i := range vms {
		name := "azure://" + vms[i]
		instances[i] = cloudprovider.Instance{Id: name}
	}

	scaleSet.instanceCache = instances
	scaleSet.lastInstanceRefresh = time.Now()
	return instances, nil
}

func (scaleSet *ScaleSet) invalidateInstanceCache() {
	scaleSet.instanceMutex.Lock()
	// Set the instanceCache as outdated.
	scaleSet.lastInstanceRefresh = time.Now().Add(-1 * vmssInstancesRefreshPeriod)
	scaleSet.instanceMutex.Unlock()
}
