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
	"regexp"
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
	klog "k8s.io/klog/v2"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"k8s.io/legacy-cloud-providers/azure/retry"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

var (
	defaultVmssSizeRefreshPeriod = 15 * time.Second
	vmssInstancesRefreshPeriod   = 5 * time.Minute
	vmssContextTimeout           = 3 * time.Minute
	vmssSizeMutex                sync.Mutex
)

var scaleSetStatusCache struct {
	lastRefresh time.Time
	mutex       sync.Mutex
	scaleSets   map[string]compute.VirtualMachineScaleSet
}

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

	sizeMutex         sync.Mutex
	curSize           int64
	lastSizeRefresh   time.Time
	sizeRefreshPeriod time.Duration

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

	if az.config.VmssCacheTTL != 0 {
		scaleSet.sizeRefreshPeriod = time.Duration(az.config.VmssCacheTTL) * time.Second
	} else {
		scaleSet.sizeRefreshPeriod = defaultVmssSizeRefreshPeriod
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

func (scaleSet *ScaleSet) getVMSSInfo() (compute.VirtualMachineScaleSet, *retry.Error) {
	scaleSetStatusCache.mutex.Lock()
	defer scaleSetStatusCache.mutex.Unlock()

	if scaleSetStatusCache.lastRefresh.Add(scaleSet.sizeRefreshPeriod).After(time.Now()) {
		if status, exists := scaleSetStatusCache.scaleSets[scaleSet.Name]; exists {
			return status, nil
		}
	}

	var allVMSS []compute.VirtualMachineScaleSet
	var rerr *retry.Error

	allVMSS, rerr = scaleSet.getAllVMSSInfo()
	if rerr != nil {
		return compute.VirtualMachineScaleSet{}, rerr
	}

	var newStatus = make(map[string]compute.VirtualMachineScaleSet)
	for _, vmss := range allVMSS {
		newStatus[*vmss.Name] = vmss
	}

	scaleSetStatusCache.lastRefresh = time.Now()
	scaleSetStatusCache.scaleSets = newStatus

	if _, exists := scaleSetStatusCache.scaleSets[scaleSet.Name]; !exists {
		return compute.VirtualMachineScaleSet{}, &retry.Error{RawError: fmt.Errorf("could not find vmss: %s", scaleSet.Name)}
	}

	return scaleSetStatusCache.scaleSets[scaleSet.Name], nil
}

func (scaleSet *ScaleSet) getAllVMSSInfo() ([]compute.VirtualMachineScaleSet, *retry.Error) {
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	resourceGroup := scaleSet.manager.config.ResourceGroup
	setInfo, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.List(ctx, resourceGroup)
	if rerr != nil {
		return []compute.VirtualMachineScaleSet{}, rerr
	}

	return setInfo, nil
}

func (scaleSet *ScaleSet) getCurSize() (int64, error) {
	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	if scaleSet.lastSizeRefresh.Add(scaleSet.sizeRefreshPeriod).After(time.Now()) {
		return scaleSet.curSize, nil
	}

	klog.V(5).Infof("Get scale set size for %q", scaleSet.Name)
	set, rerr := scaleSet.getVMSSInfo()
	if rerr != nil {
		if isAzureRequestsThrottled(rerr) {
			// Log a warning and update the size refresh time so that it would retry after next sizeRefreshPeriod.
			klog.Warningf("getVMSSInfo() is throttled with message %v, would return the cached vmss size", rerr)
			scaleSet.lastSizeRefresh = time.Now()
			return scaleSet.curSize, nil
		}
		return -1, rerr.Error()
	}

	vmssSizeMutex.Lock()
	curSize := *set.Sku.Capacity
	vmssSizeMutex.Unlock()

	klog.V(5).Infof("Getting scale set (%q) capacity: %d\n", scaleSet.Name, curSize)

	if scaleSet.curSize != curSize {
		// Invalidate the instance cache if the capacity has changed.
		scaleSet.invalidateInstanceCache()
	}

	scaleSet.curSize = curSize
	scaleSet.lastSizeRefresh = time.Now()
	return scaleSet.curSize, nil
}

// GetScaleSetSize gets Scale Set size.
func (scaleSet *ScaleSet) GetScaleSetSize() (int64, error) {
	return scaleSet.getCurSize()
}

func (scaleSet *ScaleSet) waitForDeleteInstances(future *azure.Future, requiredIds *compute.VirtualMachineScaleSetVMInstanceRequiredIDs) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForAsyncOperationResult - DeleteInstances(%v)", requiredIds.InstanceIds)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForAsyncOperationResult(ctx, future)
	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)
	if isSuccess {
		klog.V(3).Infof("virtualMachineScaleSetsClient.WaitForAsyncOperationResult - DeleteInstances(%v) success", requiredIds.InstanceIds)
		return
	}
	klog.Errorf("virtualMachineScaleSetsClient.WaitForAsyncOperationResult - DeleteInstances for instances %v failed with error: %v", requiredIds.InstanceIds, err)
}

// updateVMSSCapacity invokes virtualMachineScaleSetsClient to update the capacity for VMSS.
func (scaleSet *ScaleSet) updateVMSSCapacity(future *azure.Future) {
	var err error

	defer func() {
		if err != nil {
			klog.Errorf("Failed to update the capacity for vmss %s with error %v, invalidate the cache so as to get the real size from API", scaleSet.Name, err)
			// Invalidate the VMSS size cache in order to fetch the size from the API.
			scaleSet.invalidateStatusCacheWithLock()
		}
	}()

	ctx, cancel := getContextWithCancel()
	defer cancel()

	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.WaitForAsyncOperationResult - updateVMSSCapacity(%s)", scaleSet.Name)
	httpResponse, err := scaleSet.manager.azClient.virtualMachineScaleSetsClient.WaitForAsyncOperationResult(ctx, future)

	isSuccess, err := isSuccessHTTPResponse(httpResponse, err)
	if isSuccess {
		klog.V(3).Infof("virtualMachineScaleSetsClient.WaitForAsyncOperationResult - updateVMSSCapacity(%s) success", scaleSet.Name)
		scaleSet.invalidateInstanceCache()
		return
	}

	klog.Errorf("virtualMachineScaleSetsClient.WaitForAsyncOperationResult - updateVMSSCapacity for scale set %q failed: %v", scaleSet.Name, err)
}

// SetScaleSetSize sets ScaleSet size.
func (scaleSet *ScaleSet) SetScaleSetSize(size int64) error {
	scaleSet.sizeMutex.Lock()
	defer scaleSet.sizeMutex.Unlock()

	vmssInfo, rerr := scaleSet.getVMSSInfo()
	if rerr != nil {
		klog.Errorf("Failed to get information for VMSS (%q): %v", scaleSet.Name, rerr)
		return rerr.Error()
	}

	// Update the new capacity to cache.
	vmssSizeMutex.Lock()
	vmssInfo.Sku.Capacity = &size
	vmssSizeMutex.Unlock()

	// Compose a new VMSS for updating.
	op := compute.VirtualMachineScaleSet{
		Name:     vmssInfo.Name,
		Sku:      vmssInfo.Sku,
		Location: vmssInfo.Location,
	}
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()
	klog.V(3).Infof("Waiting for virtualMachineScaleSetsClient.CreateOrUpdateAsync(%s)", scaleSet.Name)
	future, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.CreateOrUpdateAsync(ctx, scaleSet.manager.config.ResourceGroup, scaleSet.Name, op)
	if rerr != nil {
		klog.Errorf("virtualMachineScaleSetsClient.CreateOrUpdate for scale set %q failed: %v", scaleSet.Name, rerr)
		return rerr.Error()
	}

	// Proactively set the VMSS size so autoscaler makes better decisions.
	scaleSet.curSize = size
	scaleSet.lastSizeRefresh = time.Now()

	go scaleSet.updateVMSSCapacity(future)

	return nil
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

	if size == -1 {
		return fmt.Errorf("the scale set %s is under initialization, skipping IncreaseSize", scaleSet.Name)
	}

	if int(size)+delta > scaleSet.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, scaleSet.MaxSize())
	}

	return scaleSet.SetScaleSetSize(size + int64(delta))
}

// GetScaleSetVms returns list of nodes for the given scale set.
func (scaleSet *ScaleSet) GetScaleSetVms() ([]compute.VirtualMachineScaleSetVM, *retry.Error) {
	klog.V(4).Infof("GetScaleSetVms: starts")
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	resourceGroup := scaleSet.manager.config.ResourceGroup
	vmList, rerr := scaleSet.manager.azClient.virtualMachineScaleSetVMsClient.List(ctx, resourceGroup, scaleSet.Name, "")
	klog.V(4).Infof("GetScaleSetVms: scaleSet.Name: %s, vmList: %v", scaleSet.Name, vmList)
	if rerr != nil {
		klog.Errorf("VirtualMachineScaleSetVMsClient.List failed for %s: %v", scaleSet.Name, rerr)
		return nil, rerr
	}

	return vmList, nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (scaleSet *ScaleSet) DecreaseTargetSize(delta int) error {
	// VMSS size would be changed automatically after the Node deletion, hence this operation is not required.
	// To prevent some unreproducible bugs, an extra refresh of cache is needed.
	scaleSet.invalidateInstanceCache()
	_, err := scaleSet.GetScaleSetSize()
	if err != nil {
		klog.Warningf("DecreaseTargetSize: failed with error: %v", err)
	}
	return err
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

	klog.V(3).Infof("Deleting vmss instances %v", instances)

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

		if cpi, found := scaleSet.getInstanceByProviderID(instance.Name); found && cpi.Status != nil && cpi.Status.State == cloudprovider.InstanceDeleting {
			klog.V(3).Infof("Skipping deleting instance %s as its current state is deleting", instance.Name)
			continue
		}

		instanceID, err := getLastSegment(instance.Name)
		if err != nil {
			klog.Errorf("getLastSegment failed with error: %v", err)
			return err
		}

		instanceIDs = append(instanceIDs, instanceID)
	}

	// nothing to delete
	if len(instanceIDs) == 0 {
		klog.V(3).Infof("No new instances eligible for deletion, skipping")
		return nil
	}

	requiredIds := &compute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIds: &instanceIDs,
	}

	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()
	resourceGroup := scaleSet.manager.config.ResourceGroup

	scaleSet.instanceMutex.Lock()
	klog.V(3).Infof("Calling virtualMachineScaleSetsClient.DeleteInstancesAsync(%v)", requiredIds.InstanceIds)
	future, rerr := scaleSet.manager.azClient.virtualMachineScaleSetsClient.DeleteInstancesAsync(ctx, resourceGroup, commonAsg.Id(), *requiredIds)
	scaleSet.instanceMutex.Unlock()
	if rerr != nil {
		klog.Errorf("virtualMachineScaleSetsClient.DeleteInstancesAsync for instances %v failed: %v", requiredIds.InstanceIds, rerr)
		return rerr.Error()
	}

	// Proactively decrement scale set size so that we don't
	// go below minimum node count if cache data is stale
	scaleSet.sizeMutex.Lock()
	scaleSet.curSize -= int64(len(instanceIDs))
	scaleSet.sizeMutex.Unlock()

	go scaleSet.waitForDeleteInstances(future, requiredIds)
	return nil
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

	var vmssType *InstanceType
	for k := range InstanceTypes {
		if strings.EqualFold(k, *template.Sku.Name) {
			vmssType = InstanceTypes[k]
			break
		}
	}

	promoRe := regexp.MustCompile(`(?i)_promo`)
	if promoRe.MatchString(*template.Sku.Name) {
		if vmssType == nil {
			// We didn't find an exact match but this is a promo type, check for matching standard
			klog.V(1).Infof("No exact match found for %s, checking standard types", *template.Sku.Name)
			skuName := promoRe.ReplaceAllString(*template.Sku.Name, "")
			for k := range InstanceTypes {
				if strings.EqualFold(k, skuName) {
					vmssType = InstanceTypes[k]
					break
				}
			}
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
	// Labels from the Scale Set's Tags
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromScaleSet(template.Tags))

	// Taints from the Scale Set's Tags
	node.Spec.Taints = extractTaintsFromScaleSet(template.Tags)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func extractLabelsFromScaleSet(tags map[string]*string) map[string]string {
	result := make(map[string]string)

	for tagName, tagValue := range tags {
		splits := strings.Split(tagName, nodeLabelTagName)
		if len(splits) > 1 {
			label := strings.Replace(splits[1], "_", "/", -1)
			if label != "" {
				result[label] = *tagValue
			}
		}
	}

	return result
}

func extractTaintsFromScaleSet(tags map[string]*string) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for tagName, tagValue := range tags {
		// The tag value must be in the format <tag>:NoSchedule
		r, _ := regexp.Compile("(.*):(?:NoSchedule|NoExecute|PreferNoSchedule)")

		if r.MatchString(*tagValue) {
			splits := strings.Split(tagName, nodeTaintTagName)
			if len(splits) > 1 {
				values := strings.SplitN(*tagValue, ":", 2)
				if len(values) > 1 {
					taintKey := strings.Replace(splits[1], "_", "/", -1)
					taints = append(taints, apiv1.Taint{
						Key:    taintKey,
						Value:  values[0],
						Effect: apiv1.TaintEffect(values[1]),
					})
				}
			}
		}
	}

	return taints
}

// TemplateNodeInfo returns a node template for this scale set.
func (scaleSet *ScaleSet) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	template, rerr := scaleSet.getVMSSInfo()
	if rerr != nil {
		return nil, rerr.Error()
	}

	node, err := scaleSet.buildNodeFromTemplate(template)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(scaleSet.Name))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// Nodes returns a list of all nodes that belong to this node group.
func (scaleSet *ScaleSet) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(4).Infof("Nodes: starts, scaleSet.Name: %s", scaleSet.Name)
	curSize, err := scaleSet.getCurSize()
	if err != nil {
		klog.Errorf("Failed to get current size for vmss %q: %v", scaleSet.Name, err)
		return nil, err
	}

	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	if int64(len(scaleSet.instanceCache)) == curSize &&
		scaleSet.lastInstanceRefresh.Add(vmssInstancesRefreshPeriod).After(time.Now()) {
		klog.V(4).Infof("Nodes: returns with curSize %d", curSize)
		return scaleSet.instanceCache, nil
	}

	klog.V(4).Infof("Nodes: starts to get VMSS VMs")
	vms, rerr := scaleSet.GetScaleSetVms()
	if rerr != nil {
		if isAzureRequestsThrottled(rerr) {
			// Log a warning and update the instance refresh time so that it would retry after next vmssInstancesRefreshPeriod.
			klog.Warningf("GetScaleSetVms() is throttled with message %v, would return the cached instances", rerr)
			scaleSet.lastInstanceRefresh = time.Now()
			return scaleSet.instanceCache, nil
		}
		return nil, rerr.Error()
	}

	scaleSet.instanceCache = buildInstanceCache(vms)
	scaleSet.lastInstanceRefresh = time.Now()
	klog.V(4).Infof("Nodes: returns")
	return scaleSet.instanceCache, nil
}

// Note that the GetScaleSetVms() results is not used directly because for the List endpoint,
// their resource ID format is not consistent with Get endpoint
func buildInstanceCache(vms []compute.VirtualMachineScaleSetVM) []cloudprovider.Instance {
	instances := []cloudprovider.Instance{}

	for _, vm := range vms {
		// The resource ID is empty string, which indicates the instance may be in deleting state.
		if len(*vm.ID) == 0 {
			continue
		}

		resourceID, err := convertResourceGroupNameToLower(*vm.ID)
		if err != nil {
			// This shouldn't happen. Log a waring message for tracking.
			klog.Warningf("buildInstanceCache.convertResourceGroupNameToLower failed with error: %v", err)
			continue
		}

		instances = append(instances, cloudprovider.Instance{
			Id:     "azure://" + resourceID,
			Status: instanceStatusFromVM(vm),
		})
	}

	return instances
}

func (scaleSet *ScaleSet) getInstanceByProviderID(providerID string) (cloudprovider.Instance, bool) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()
	for _, instance := range scaleSet.instanceCache {
		if instance.Id == providerID {
			return instance, true
		}
	}
	return cloudprovider.Instance{}, false
}

// instanceStatusFromVM converts the VM provisioning state to cloudprovider.InstanceStatus
func instanceStatusFromVM(vm compute.VirtualMachineScaleSetVM) *cloudprovider.InstanceStatus {
	if vm.ProvisioningState == nil {
		return nil
	}

	status := &cloudprovider.InstanceStatus{}
	switch *vm.ProvisioningState {
	case string(compute.ProvisioningStateDeleting):
		status.State = cloudprovider.InstanceDeleting
	case string(compute.ProvisioningStateCreating):
		status.State = cloudprovider.InstanceCreating
	default:
		status.State = cloudprovider.InstanceRunning
	}

	return status
}

func (scaleSet *ScaleSet) invalidateInstanceCache() {
	scaleSet.instanceMutex.Lock()
	// Set the instanceCache as outdated.
	scaleSet.lastInstanceRefresh = time.Now().Add(-1 * vmssInstancesRefreshPeriod)
	scaleSet.instanceMutex.Unlock()
}

func (scaleSet *ScaleSet) invalidateStatusCacheWithLock() {
	scaleSet.sizeMutex.Lock()
	scaleSet.lastSizeRefresh = time.Now().Add(-1 * scaleSet.sizeRefreshPeriod)
	scaleSet.sizeMutex.Unlock()

	scaleSetStatusCache.mutex.Lock()
	scaleSetStatusCache.lastRefresh = time.Now().Add(-1 * scaleSet.sizeRefreshPeriod)
	scaleSetStatusCache.mutex.Unlock()
}
