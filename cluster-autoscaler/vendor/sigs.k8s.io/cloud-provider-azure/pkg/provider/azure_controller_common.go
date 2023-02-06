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

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"

	"k8s.io/apimachinery/pkg/types"
	kwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/flowcontrol"
	cloudprovider "k8s.io/cloud-provider"
	volerr "k8s.io/cloud-provider/volume/errors"
	"k8s.io/klog/v2"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

const (
	// Disk Caching is not supported for disks 4 TiB and larger
	// https://docs.microsoft.com/en-us/azure/virtual-machines/premium-storage-performance#disk-caching
	diskCachingLimit = 4096 // GiB

	maxLUN                 = 64 // max number of LUNs per VM
	errStatusCode400       = "statuscode=400"
	errInvalidParameter    = `code="invalidparameter"`
	errTargetInstanceIds   = `target="instanceids"`
	sourceSnapshot         = "snapshot"
	sourceVolume           = "volume"
	attachDiskMapKeySuffix = "attachdiskmap"
	detachDiskMapKeySuffix = "detachdiskmap"

	// WriteAcceleratorEnabled support for Azure Write Accelerator on Azure Disks
	// https://docs.microsoft.com/azure/virtual-machines/windows/how-to-enable-write-accelerator
	WriteAcceleratorEnabled = "writeacceleratorenabled"

	// see https://docs.microsoft.com/en-us/rest/api/compute/disks/createorupdate#create-a-managed-disk-by-copying-a-snapshot.
	diskSnapshotPath = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/snapshots/%s"

	// see https://docs.microsoft.com/en-us/rest/api/compute/disks/createorupdate#create-a-managed-disk-from-an-existing-managed-disk-in-the-same-or-different-subscription.
	managedDiskPath = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/disks/%s"
)

var defaultBackOff = kwait.Backoff{
	Steps:    20,
	Duration: 2 * time.Second,
	Factor:   1.5,
	Jitter:   0.0,
}

var (
	managedDiskPathRE  = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(?:.*)/providers/Microsoft.Compute/disks/(.+)`)
	diskSnapshotPathRE = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(?:.*)/providers/Microsoft.Compute/snapshots/(.+)`)
)

type controllerCommon struct {
	subscriptionID        string
	location              string
	extendedLocation      *ExtendedLocation
	storageEndpointSuffix string
	resourceGroup         string
	diskStateMap          sync.Map // <diskURI, attaching/detaching state>
	lockMap               *lockMap
	cloud                 *Cloud
	// disk queue that is waiting for attach or detach on specific node
	// <nodeName, map<diskURI, *AttachDiskOptions/DetachDiskOptions>>
	attachDiskMap sync.Map
	detachDiskMap sync.Map
	// attach/detach disk rate limiter
	diskOpRateLimiter flowcontrol.RateLimiter
}

// AttachDiskOptions attach disk options
type AttachDiskOptions struct {
	cachingMode             compute.CachingTypes
	diskName                string
	diskEncryptionSetID     string
	writeAcceleratorEnabled bool
	lun                     int32
}

// ExtendedLocation contains additional info about the location of resources.
type ExtendedLocation struct {
	// Name - The name of the extended location.
	Name string `json:"name,omitempty"`
	// Type - The type of the extended location.
	Type string `json:"type,omitempty"`
}

// getNodeVMSet gets the VMSet interface based on config.VMType and the real virtual machine type.
func (c *controllerCommon) getNodeVMSet(nodeName types.NodeName, crt azcache.AzureCacheReadType) (VMSet, error) {
	// 1. vmType is standard or vmssflex, return cloud.VMSet directly.
	// 1.1 all the nodes in the cluster are avset nodes.
	// 1.2 all the nodes in the cluster are vmssflex nodes.
	if c.cloud.VMType == consts.VMTypeStandard || c.cloud.VMType == consts.VMTypeVmssFlex {
		return c.cloud.VMSet, nil
	}

	// 2. vmType is Virtual Machine Scale Set (vmss), convert vmSet to ScaleSet.
	// 2.1 all the nodes in the cluster are vmss uniform nodes.
	// 2.2 mix node: the nodes in the cluster can be any of avset nodes, vmss uniform nodes and vmssflex nodes.
	ss, ok := c.cloud.VMSet.(*ScaleSet)
	if !ok {
		return nil, fmt.Errorf("error of converting vmSet (%q) to ScaleSet with vmType %q", c.cloud.VMSet, c.cloud.VMType)
	}

	vmManagementType, err := ss.getVMManagementTypeByNodeName(mapNodeNameToVMName(nodeName), crt)
	if err != nil {
		return nil, fmt.Errorf("getNodeVMSet: failed to check the node %s management type: %w", mapNodeNameToVMName(nodeName), err)
	}
	// 3. If the node is managed by availability set, then return ss.availabilitySet.
	if vmManagementType == ManagedByAvSet {
		// vm is managed by availability set.
		return ss.availabilitySet, nil
	}
	if vmManagementType == ManagedByVmssFlex {
		// 4. If the node is managed by vmss flex, then return ss.flexScaleSet.
		// vm is managed by vmss flex.
		return ss.flexScaleSet, nil
	}

	// 5. Node is managed by vmss
	return ss, nil

}

// AttachDisk attaches a disk to vm
// parameter async indicates whether allow multiple batch disk attach on one node in parallel
// return (lun, error)
func (c *controllerCommon) AttachDisk(ctx context.Context, async bool, diskName, diskURI string, nodeName types.NodeName,
	cachingMode compute.CachingTypes, disk *compute.Disk) (int32, error) {
	diskEncryptionSetID := ""
	writeAcceleratorEnabled := false

	// there is possibility that disk is nil when GetDisk is throttled
	// don't check disk state when GetDisk is throttled
	if disk != nil {
		if disk.ManagedBy != nil && (disk.MaxShares == nil || *disk.MaxShares <= 1) {
			vmset, err := c.getNodeVMSet(nodeName, azcache.CacheReadTypeUnsafe)
			if err != nil {
				return -1, err
			}
			attachedNode, err := vmset.GetNodeNameByProviderID(*disk.ManagedBy)
			if err != nil {
				return -1, err
			}
			if strings.EqualFold(string(nodeName), string(attachedNode)) {
				klog.Warningf("volume %s is actually attached to current node %s, invalidate vm cache and return error", diskURI, nodeName)
				// update VM(invalidate vm cache)
				if errUpdate := c.UpdateVM(ctx, nodeName); errUpdate != nil {
					return -1, errUpdate
				}
				lun, _, err := c.GetDiskLun(diskName, diskURI, nodeName)
				return lun, err
			}

			attachErr := fmt.Sprintf(
				"disk(%s) already attached to node(%s), could not be attached to node(%s)",
				diskURI, *disk.ManagedBy, nodeName)
			klog.V(2).Infof("found dangling volume %s attached to node %s, could not be attached to node(%s)", diskURI, attachedNode, nodeName)
			return -1, volerr.NewDanglingError(attachErr, attachedNode, "")
		}

		if disk.DiskProperties != nil {
			if disk.DiskProperties.DiskSizeGB != nil && *disk.DiskProperties.DiskSizeGB >= diskCachingLimit && cachingMode != compute.CachingTypesNone {
				// Disk Caching is not supported for disks 4 TiB and larger
				// https://docs.microsoft.com/en-us/azure/virtual-machines/premium-storage-performance#disk-caching
				cachingMode = compute.CachingTypesNone
				klog.Warningf("size of disk(%s) is %dGB which is bigger than limit(%dGB), set cacheMode as None",
					diskURI, *disk.DiskProperties.DiskSizeGB, diskCachingLimit)
			}

			if disk.DiskProperties.Encryption != nil &&
				disk.DiskProperties.Encryption.DiskEncryptionSetID != nil {
				diskEncryptionSetID = *disk.DiskProperties.Encryption.DiskEncryptionSetID
			}

			if disk.DiskProperties.DiskState != compute.Unattached && (disk.MaxShares == nil || *disk.MaxShares <= 1) {
				return -1, fmt.Errorf("state of disk(%s) is %s, not in expected %s state", diskURI, disk.DiskProperties.DiskState, compute.Unattached)
			}
		}

		if v, ok := disk.Tags[WriteAcceleratorEnabled]; ok {
			if v != nil && strings.EqualFold(*v, "true") {
				writeAcceleratorEnabled = true
			}
		}
	}

	options := AttachDiskOptions{
		lun:                     -1,
		diskName:                diskName,
		cachingMode:             cachingMode,
		diskEncryptionSetID:     diskEncryptionSetID,
		writeAcceleratorEnabled: writeAcceleratorEnabled,
	}
	node := strings.ToLower(string(nodeName))
	diskuri := strings.ToLower(diskURI)
	if err := c.insertAttachDiskRequest(diskuri, node, &options); err != nil {
		return -1, err
	}

	c.lockMap.LockEntry(node)
	unlock := false
	defer func() {
		if !unlock {
			c.lockMap.UnlockEntry(node)
		}
	}()

	diskMap, err := c.cleanAttachDiskRequests(node)
	if err != nil {
		return -1, err
	}

	lun, err := c.SetDiskLun(nodeName, diskuri, diskMap)
	if err != nil {
		return -1, err
	}

	klog.V(2).Infof("Trying to attach volume %s lun %d to node %s, diskMap: %s", diskURI, lun, nodeName, diskMap)
	if len(diskMap) == 0 {
		return lun, nil
	}

	vmset, err := c.getNodeVMSet(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return -1, err
	}
	c.diskStateMap.Store(disk, "attaching")
	defer c.diskStateMap.Delete(disk)
	future, err := vmset.AttachDisk(ctx, nodeName, diskMap)
	if err != nil {
		return -1, err
	}

	if async && c.diskOpRateLimiter.TryAccept() {
		// unlock and wait for attach disk complete
		unlock = true
		c.lockMap.UnlockEntry(node)
	} else {
		if async {
			klog.Warningf("azureDisk - switch to batch operation due to rate limited, QPS: %f", c.diskOpRateLimiter.QPS())
		}
	}
	return lun, vmset.WaitForUpdateResult(ctx, future, nodeName, "attach_disk")
}

func (c *controllerCommon) insertAttachDiskRequest(diskURI, nodeName string, options *AttachDiskOptions) error {
	var diskMap map[string]*AttachDiskOptions
	attachDiskMapKey := nodeName + attachDiskMapKeySuffix
	c.lockMap.LockEntry(attachDiskMapKey)
	defer c.lockMap.UnlockEntry(attachDiskMapKey)
	v, ok := c.attachDiskMap.Load(nodeName)
	if ok {
		if diskMap, ok = v.(map[string]*AttachDiskOptions); !ok {
			return fmt.Errorf("convert attachDiskMap failure on node(%s)", nodeName)
		}
	} else {
		diskMap = make(map[string]*AttachDiskOptions)
		c.attachDiskMap.Store(nodeName, diskMap)
	}
	// insert attach disk request to queue
	_, ok = diskMap[diskURI]
	if ok {
		klog.V(2).Infof("azureDisk - duplicated attach disk(%s) request on node(%s)", diskURI, nodeName)
	} else {
		diskMap[diskURI] = options
	}
	return nil
}

// clean up attach disk requests
// return original attach disk requests
func (c *controllerCommon) cleanAttachDiskRequests(nodeName string) (map[string]*AttachDiskOptions, error) {
	var diskMap map[string]*AttachDiskOptions

	attachDiskMapKey := nodeName + attachDiskMapKeySuffix
	c.lockMap.LockEntry(attachDiskMapKey)
	defer c.lockMap.UnlockEntry(attachDiskMapKey)
	v, ok := c.attachDiskMap.Load(nodeName)
	if !ok {
		return diskMap, nil
	}
	if diskMap, ok = v.(map[string]*AttachDiskOptions); !ok {
		return diskMap, fmt.Errorf("convert attachDiskMap failure on node(%s)", nodeName)
	}
	c.attachDiskMap.Store(nodeName, make(map[string]*AttachDiskOptions))
	return diskMap, nil
}

// DetachDisk detaches a disk from VM
func (c *controllerCommon) DetachDisk(ctx context.Context, diskName, diskURI string, nodeName types.NodeName) error {
	if _, err := c.cloud.InstanceID(ctx, nodeName); err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			// if host doesn't exist, no need to detach
			klog.Warningf("azureDisk - failed to get azure instance id(%s), DetachDisk(%s) will assume disk is already detached",
				nodeName, diskURI)
			return nil
		}
		klog.Warningf("failed to get azure instance id (%v)", err)
		return fmt.Errorf("failed to get azure instance id for node %q: %w", nodeName, err)
	}

	vmset, err := c.getNodeVMSet(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return err
	}

	node := strings.ToLower(string(nodeName))
	disk := strings.ToLower(diskURI)
	if err := c.insertDetachDiskRequest(diskName, disk, node); err != nil {
		return err
	}

	c.lockMap.LockEntry(node)
	defer c.lockMap.UnlockEntry(node)
	diskMap, err := c.cleanDetachDiskRequests(node)
	if err != nil {
		return err
	}

	klog.V(2).Infof("Trying to detach volume %s from node %s, diskMap: %s", diskURI, nodeName, diskMap)
	if len(diskMap) > 0 {
		c.diskStateMap.Store(disk, "detaching")
		defer c.diskStateMap.Delete(disk)
		if err = vmset.DetachDisk(ctx, nodeName, diskMap); err != nil {
			if isInstanceNotFoundError(err) {
				// if host doesn't exist, no need to detach
				klog.Warningf("azureDisk - got InstanceNotFoundError(%v), DetachDisk(%s) will assume disk is already detached",
					err, diskURI)
				return nil
			}
		}
	} else {
		lun, _, errGetLun := c.GetDiskLun(diskName, diskURI, nodeName)
		if errGetLun == nil || !strings.Contains(errGetLun.Error(), consts.CannotFindDiskLUN) {
			return fmt.Errorf("disk(%s) is still attached to node(%s) on lun(%d), error: %w", diskURI, nodeName, lun, errGetLun)
		}
	}

	if err != nil {
		klog.Errorf("azureDisk - detach disk(%s, %s) failed, err: %v", diskName, diskURI, err)
		return err
	}

	klog.V(2).Infof("azureDisk - detach disk(%s, %s) succeeded", diskName, diskURI)
	return nil
}

// UpdateVM updates a vm
func (c *controllerCommon) UpdateVM(ctx context.Context, nodeName types.NodeName) error {
	vmset, err := c.getNodeVMSet(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		return err
	}
	node := strings.ToLower(string(nodeName))
	c.lockMap.LockEntry(node)
	defer c.lockMap.UnlockEntry(node)
	return vmset.UpdateVM(ctx, nodeName)
}

func (c *controllerCommon) insertDetachDiskRequest(diskName, diskURI, nodeName string) error {
	var diskMap map[string]string
	detachDiskMapKey := nodeName + detachDiskMapKeySuffix
	c.lockMap.LockEntry(detachDiskMapKey)
	defer c.lockMap.UnlockEntry(detachDiskMapKey)
	v, ok := c.detachDiskMap.Load(nodeName)
	if ok {
		if diskMap, ok = v.(map[string]string); !ok {
			return fmt.Errorf("convert detachDiskMap failure on node(%s)", nodeName)
		}
	} else {
		diskMap = make(map[string]string)
		c.detachDiskMap.Store(nodeName, diskMap)
	}
	// insert detach disk request to queue
	_, ok = diskMap[diskURI]
	if ok {
		klog.V(2).Infof("azureDisk - duplicated detach disk(%s) request on node(%s)", diskURI, nodeName)
	} else {
		diskMap[diskURI] = diskName
	}
	return nil
}

// clean up detach disk requests
// return original detach disk requests
func (c *controllerCommon) cleanDetachDiskRequests(nodeName string) (map[string]string, error) {
	var diskMap map[string]string

	detachDiskMapKey := nodeName + detachDiskMapKeySuffix
	c.lockMap.LockEntry(detachDiskMapKey)
	defer c.lockMap.UnlockEntry(detachDiskMapKey)
	v, ok := c.detachDiskMap.Load(nodeName)
	if !ok {
		return diskMap, nil
	}
	if diskMap, ok = v.(map[string]string); !ok {
		return diskMap, fmt.Errorf("convert detachDiskMap failure on node(%s)", nodeName)
	}
	// clean up original requests in disk map
	c.detachDiskMap.Store(nodeName, make(map[string]string))
	return diskMap, nil
}

// getNodeDataDisks invokes vmSet interfaces to get data disks for the node.
func (c *controllerCommon) getNodeDataDisks(nodeName types.NodeName, crt azcache.AzureCacheReadType) ([]compute.DataDisk, *string, error) {
	vmset, err := c.getNodeVMSet(nodeName, crt)
	if err != nil {
		return nil, nil, err
	}

	return vmset.GetDataDisks(nodeName, crt)
}

// GetDiskLun finds the lun on the host that the vhd is attached to, given a vhd's diskName and diskURI.
func (c *controllerCommon) GetDiskLun(diskName, diskURI string, nodeName types.NodeName) (int32, *string, error) {
	// getNodeDataDisks need to fetch the cached data/fresh data if cache expired here
	// to ensure we get LUN based on latest entry.
	disks, provisioningState, err := c.getNodeDataDisks(nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		klog.Errorf("error of getting data disks for node %s: %v", nodeName, err)
		return -1, provisioningState, err
	}

	for _, disk := range disks {
		if disk.Lun != nil && (disk.Name != nil && diskName != "" && strings.EqualFold(*disk.Name, diskName)) ||
			(disk.Vhd != nil && disk.Vhd.URI != nil && diskURI != "" && strings.EqualFold(*disk.Vhd.URI, diskURI)) ||
			(disk.ManagedDisk != nil && strings.EqualFold(*disk.ManagedDisk.ID, diskURI)) {
			if disk.ToBeDetached != nil && *disk.ToBeDetached {
				klog.Warningf("azureDisk - find disk(ToBeDetached): lun %d name %s uri %s", *disk.Lun, diskName, diskURI)
			} else {
				// found the disk
				klog.V(2).Infof("azureDisk - find disk: lun %d name %s uri %s", *disk.Lun, diskName, diskURI)
				return *disk.Lun, provisioningState, nil
			}
		}
	}
	return -1, provisioningState, fmt.Errorf("%s for disk %s", consts.CannotFindDiskLUN, diskName)
}

// SetDiskLun find unused luns and allocate lun for every disk in diskMap.
// Return lun of diskURI, -1 if all luns are used.
func (c *controllerCommon) SetDiskLun(nodeName types.NodeName, diskURI string, diskMap map[string]*AttachDiskOptions) (int32, error) {
	disks, _, err := c.getNodeDataDisks(nodeName, azcache.CacheReadTypeDefault)
	if err != nil {
		klog.Errorf("error of getting data disks for node %s: %v", nodeName, err)
		return -1, err
	}

	lun := int32(-1)
	_, isDiskInMap := diskMap[diskURI]
	used := make([]bool, maxLUN)
	for _, disk := range disks {
		if disk.Lun != nil {
			used[*disk.Lun] = true
			if !isDiskInMap {
				// find lun of diskURI since diskURI is not in diskMap
				if disk.ManagedDisk != nil && strings.EqualFold(*disk.ManagedDisk.ID, diskURI) {
					lun = *disk.Lun
				}
			}
		}
	}
	if !isDiskInMap && lun < 0 {
		return -1, fmt.Errorf("could not find disk(%s) in current disk list(len: %d) nor in diskMap(%v)", diskURI, len(disks), diskMap)
	}
	if len(diskMap) == 0 {
		// attach disk request is empty, return directly
		return lun, nil
	}

	// allocate lun for every disk in diskMap
	var diskLuns []int32
	count := 0
	for k, v := range used {
		if !v {
			diskLuns = append(diskLuns, int32(k))
			count++
			if count >= len(diskMap) {
				break
			}
		}
	}

	if len(diskLuns) != len(diskMap) {
		return -1, fmt.Errorf("could not find enough disk luns(current: %d) for diskMap(%v, len=%d), diskURI(%s)",
			len(diskLuns), diskMap, len(diskMap), diskURI)
	}

	count = 0
	for uri, opt := range diskMap {
		if opt == nil {
			return -1, fmt.Errorf("unexpected nil pointer in diskMap(%v), diskURI(%s)", diskMap, diskURI)
		}
		if strings.EqualFold(uri, diskURI) {
			lun = diskLuns[count]
		}
		opt.lun = diskLuns[count]
		count++
	}
	if lun < 0 {
		return lun, fmt.Errorf("could not find lun of diskURI(%s), diskMap(%v)", diskURI, diskMap)
	}
	return lun, nil
}

// DisksAreAttached checks if a list of volumes are attached to the node with the specified NodeName.
func (c *controllerCommon) DisksAreAttached(diskNames []string, nodeName types.NodeName) (map[string]bool, error) {
	attached := make(map[string]bool)
	for _, diskName := range diskNames {
		attached[diskName] = false
	}

	// doing stalled read for getNodeDataDisks to ensure we don't call ARM
	// for every reconcile call. The cache is invalidated after Attach/Detach
	// disk. So the new entry will be fetched and cached the first time reconcile
	// loop runs after the Attach/Disk OP which will reflect the latest model.
	disks, _, err := c.getNodeDataDisks(nodeName, azcache.CacheReadTypeUnsafe)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			// if host doesn't exist, no need to detach
			klog.Warningf("azureDisk - Cannot find node %s, DisksAreAttached will assume disks %v are not attached to it.",
				nodeName, diskNames)
			return attached, nil
		}

		return attached, err
	}

	for _, disk := range disks {
		for _, diskName := range diskNames {
			if disk.Name != nil && diskName != "" && strings.EqualFold(*disk.Name, diskName) {
				attached[diskName] = true
			}
		}
	}

	return attached, nil
}

func filterDetachingDisks(unfilteredDisks []compute.DataDisk) []compute.DataDisk {
	filteredDisks := []compute.DataDisk{}
	for _, disk := range unfilteredDisks {
		if disk.ToBeDetached != nil && *disk.ToBeDetached {
			if disk.Name != nil {
				klog.V(2).Infof("Filtering disk: %s with ToBeDetached flag set.", *disk.Name)
			}
		} else {
			filteredDisks = append(filteredDisks, disk)
		}
	}
	return filteredDisks
}

func (c *controllerCommon) filterNonExistingDisks(ctx context.Context, unfilteredDisks []compute.DataDisk) []compute.DataDisk {
	filteredDisks := []compute.DataDisk{}
	for _, disk := range unfilteredDisks {
		filter := false
		if disk.ManagedDisk != nil && disk.ManagedDisk.ID != nil {
			diskURI := *disk.ManagedDisk.ID
			exist, err := c.cloud.checkDiskExists(ctx, diskURI)
			if err != nil {
				klog.Errorf("checkDiskExists(%s) failed with error: %v", diskURI, err)
			} else {
				// only filter disk when checkDiskExists returns <false, nil>
				filter = !exist
				if filter {
					klog.Errorf("disk(%s) does not exist, removed from data disk list", diskURI)
				}
			}
		}

		if !filter {
			filteredDisks = append(filteredDisks, disk)
		}
	}
	return filteredDisks
}

func (c *controllerCommon) checkDiskExists(ctx context.Context, diskURI string) (bool, error) {
	diskName := path.Base(diskURI)
	resourceGroup, subsID, err := getInfoFromDiskURI(diskURI)
	if err != nil {
		return false, err
	}

	if _, rerr := c.cloud.DisksClient.Get(ctx, subsID, resourceGroup, diskName); rerr != nil {
		if rerr.HTTPStatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, rerr.Error()
	}

	return true, nil
}

func getValidCreationData(subscriptionID, resourceGroup, sourceResourceID, sourceType string) (compute.CreationData, error) {
	if sourceResourceID == "" {
		return compute.CreationData{
			CreateOption: compute.Empty,
		}, nil
	}

	switch sourceType {
	case sourceSnapshot:
		if match := diskSnapshotPathRE.FindString(sourceResourceID); match == "" {
			sourceResourceID = fmt.Sprintf(diskSnapshotPath, subscriptionID, resourceGroup, sourceResourceID)
		}

	case sourceVolume:
		if match := managedDiskPathRE.FindString(sourceResourceID); match == "" {
			sourceResourceID = fmt.Sprintf(managedDiskPath, subscriptionID, resourceGroup, sourceResourceID)
		}
	default:
		return compute.CreationData{
			CreateOption: compute.Empty,
		}, nil
	}

	splits := strings.Split(sourceResourceID, "/")
	if len(splits) > 9 {
		if sourceType == sourceSnapshot {
			return compute.CreationData{}, fmt.Errorf("sourceResourceID(%s) is invalid, correct format: %s", sourceResourceID, diskSnapshotPathRE)
		}
		return compute.CreationData{}, fmt.Errorf("sourceResourceID(%s) is invalid, correct format: %s", sourceResourceID, managedDiskPathRE)
	}
	return compute.CreationData{
		CreateOption:     compute.Copy,
		SourceResourceID: &sourceResourceID,
	}, nil
}

func isInstanceNotFoundError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, strings.ToLower(consts.VmssVMNotActiveErrorMessage)) {
		return true
	}
	return strings.Contains(errMsg, errStatusCode400) && strings.Contains(errMsg, errInvalidParameter) && strings.Contains(errMsg, errTargetInstanceIds)
}
