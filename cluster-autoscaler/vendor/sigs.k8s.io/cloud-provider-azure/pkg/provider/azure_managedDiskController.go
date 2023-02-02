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
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kwait "k8s.io/apimachinery/pkg/util/wait"
	cloudvolume "k8s.io/cloud-provider/volume"
	volumehelpers "k8s.io/cloud-provider/volume/helpers"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// ManagedDiskController : managed disk controller struct
type ManagedDiskController struct {
	common *controllerCommon
}

// ManagedDiskOptions specifies the options of managed disks.
type ManagedDiskOptions struct {
	// The SKU of storage account.
	StorageAccountType compute.DiskStorageAccountTypes
	// The name of the disk.
	DiskName string
	// The name of PVC.
	PVCName string
	// The name of resource group.
	ResourceGroup string
	// The AvailabilityZone to create the disk.
	AvailabilityZone string
	// The tags of the disk.
	Tags map[string]string
	// IOPS Caps for UltraSSD disk
	DiskIOPSReadWrite string
	// Throughput Cap (MBps) for UltraSSD disk
	DiskMBpsReadWrite string
	// if SourceResourceID is not empty, then it's a disk copy operation(for snapshot)
	SourceResourceID string
	// The type of source
	SourceType string
	// ResourceId of the disk encryption set to use for enabling encryption at rest.
	DiskEncryptionSetID string
	// DiskEncryption type, available values: EncryptionAtRestWithCustomerKey, EncryptionAtRestWithPlatformAndCustomerKeys
	DiskEncryptionType string
	// The size in GB.
	SizeGB int
	// The maximum number of VMs that can attach to the disk at the same time. Value greater than one indicates a disk that can be mounted on multiple VMs at the same time.
	MaxShares int32
	// Logical sector size in bytes for Ultra disks
	LogicalSectorSize int32
	// SkipGetDiskOperation indicates whether skip GetDisk operation(mainly due to throttling)
	SkipGetDiskOperation bool
	// NetworkAccessPolicy - Possible values include: 'AllowAll', 'AllowPrivate', 'DenyAll'
	NetworkAccessPolicy compute.NetworkAccessPolicy
	// DiskAccessID - ARM id of the DiskAccess resource for using private endpoints on disks.
	DiskAccessID *string
	// BurstingEnabled - Set to true to enable bursting beyond the provisioned performance target of the disk.
	BurstingEnabled *bool
	// SubscriptionID - specify a different SubscriptionID
	SubscriptionID string
	// Location - specify a different location
	Location string
}

// CreateManagedDisk: create managed disk
func (c *ManagedDiskController) CreateManagedDisk(ctx context.Context, options *ManagedDiskOptions) (string, error) {
	var err error
	klog.V(4).Infof("azureDisk - creating new managed Name:%s StorageAccountType:%s Size:%v", options.DiskName, options.StorageAccountType, options.SizeGB)

	var createZones []string
	if len(options.AvailabilityZone) > 0 {
		requestedZone := c.common.cloud.GetZoneID(options.AvailabilityZone)
		if requestedZone != "" {
			createZones = append(createZones, requestedZone)
		}
	}

	// insert original tags to newTags
	newTags := make(map[string]*string)
	azureDDTag := "kubernetes-azure-dd"
	newTags[consts.CreatedByTag] = &azureDDTag
	if options.Tags != nil {
		for k, v := range options.Tags {
			// Azure won't allow / (forward slash) in tags
			newKey := strings.Replace(k, "/", "-", -1)
			newValue := strings.Replace(v, "/", "-", -1)
			newTags[newKey] = &newValue
		}
	}

	diskSizeGB := int32(options.SizeGB)
	diskSku := options.StorageAccountType

	rg := c.common.resourceGroup
	if options.ResourceGroup != "" {
		rg = options.ResourceGroup
	}
	if options.SubscriptionID != "" && !strings.EqualFold(options.SubscriptionID, c.common.subscriptionID) && options.ResourceGroup == "" {
		return "", fmt.Errorf("resourceGroup must be specified when subscriptionID(%s) is not empty", options.SubscriptionID)
	}
	subsID := c.common.subscriptionID
	if options.SubscriptionID != "" {
		subsID = options.SubscriptionID
	}

	creationData, err := getValidCreationData(subsID, rg, options.SourceResourceID, options.SourceType)
	if err != nil {
		return "", err
	}
	diskProperties := compute.DiskProperties{
		DiskSizeGB:      &diskSizeGB,
		CreationData:    &creationData,
		BurstingEnabled: options.BurstingEnabled,
	}

	if options.NetworkAccessPolicy != "" {
		diskProperties.NetworkAccessPolicy = options.NetworkAccessPolicy
		if options.NetworkAccessPolicy == compute.AllowPrivate {
			if options.DiskAccessID == nil {
				return "", fmt.Errorf("DiskAccessID should not be empty when NetworkAccessPolicy is AllowPrivate")
			}
			diskProperties.DiskAccessID = options.DiskAccessID
		} else {
			if options.DiskAccessID != nil {
				return "", fmt.Errorf("DiskAccessID(%s) must be empty when NetworkAccessPolicy(%s) is not AllowPrivate", *options.DiskAccessID, options.NetworkAccessPolicy)
			}
		}
	}

	if diskSku == compute.UltraSSDLRS || diskSku == consts.PremiumV2LRS {
		if options.DiskIOPSReadWrite == "" {
			if diskSku == compute.UltraSSDLRS {
				diskIOPSReadWrite := int64(consts.DefaultDiskIOPSReadWrite)
				diskProperties.DiskIOPSReadWrite = to.Int64Ptr(diskIOPSReadWrite)
			}
		} else {
			v, err := strconv.Atoi(options.DiskIOPSReadWrite)
			if err != nil {
				return "", fmt.Errorf("AzureDisk - failed to parse DiskIOPSReadWrite: %w", err)
			}
			diskIOPSReadWrite := int64(v)
			diskProperties.DiskIOPSReadWrite = to.Int64Ptr(diskIOPSReadWrite)
		}

		if options.DiskMBpsReadWrite == "" {
			if diskSku == compute.UltraSSDLRS {
				diskMBpsReadWrite := int64(consts.DefaultDiskMBpsReadWrite)
				diskProperties.DiskMBpsReadWrite = to.Int64Ptr(diskMBpsReadWrite)
			}
		} else {
			v, err := strconv.Atoi(options.DiskMBpsReadWrite)
			if err != nil {
				return "", fmt.Errorf("AzureDisk - failed to parse DiskMBpsReadWrite: %w", err)
			}
			diskMBpsReadWrite := int64(v)
			diskProperties.DiskMBpsReadWrite = to.Int64Ptr(diskMBpsReadWrite)
		}

		if options.LogicalSectorSize != 0 {
			klog.V(2).Infof("AzureDisk - requested LogicalSectorSize: %v", options.LogicalSectorSize)
			diskProperties.CreationData.LogicalSectorSize = to.Int32Ptr(options.LogicalSectorSize)
		}
	} else {
		if options.DiskIOPSReadWrite != "" {
			return "", fmt.Errorf("AzureDisk - DiskIOPSReadWrite parameter is only applicable in UltraSSD_LRS disk type")
		}
		if options.DiskMBpsReadWrite != "" {
			return "", fmt.Errorf("AzureDisk - DiskMBpsReadWrite parameter is only applicable in UltraSSD_LRS disk type")
		}
		if options.LogicalSectorSize != 0 {
			return "", fmt.Errorf("AzureDisk - LogicalSectorSize parameter is only applicable in UltraSSD_LRS disk type")
		}
	}

	if options.DiskEncryptionSetID != "" {
		if strings.Index(strings.ToLower(options.DiskEncryptionSetID), "/subscriptions/") != 0 {
			return "", fmt.Errorf("AzureDisk - format of DiskEncryptionSetID(%s) is incorrect, correct format: %s", options.DiskEncryptionSetID, consts.DiskEncryptionSetIDFormat)
		}
		encryptionType := compute.EncryptionTypeEncryptionAtRestWithCustomerKey
		if options.DiskEncryptionType != "" {
			encryptionType = compute.EncryptionType(options.DiskEncryptionType)
			klog.V(4).Infof("azureDisk - DiskEncryptionType: %s, DiskEncryptionSetID: %s", options.DiskEncryptionType, options.DiskEncryptionSetID)
		}
		diskProperties.Encryption = &compute.Encryption{
			DiskEncryptionSetID: &options.DiskEncryptionSetID,
			Type:                encryptionType,
		}
	} else {
		if options.DiskEncryptionType != "" {
			return "", fmt.Errorf("AzureDisk - DiskEncryptionType(%s) should be empty when DiskEncryptionSetID is not set", options.DiskEncryptionType)
		}
	}

	if options.MaxShares > 1 {
		diskProperties.MaxShares = &options.MaxShares
	}

	location := c.common.location
	if options.Location != "" {
		location = options.Location
	}
	model := compute.Disk{
		Location: &location,
		Tags:     newTags,
		Sku: &compute.DiskSku{
			Name: diskSku,
		},
		DiskProperties: &diskProperties,
	}

	if el := c.common.extendedLocation; el != nil {
		model.ExtendedLocation = &compute.ExtendedLocation{
			Name: to.StringPtr(el.Name),
			Type: compute.ExtendedLocationTypes(el.Type),
		}
	}

	if len(createZones) > 0 {
		model.Zones = &createZones
	}

	if rerr := c.common.cloud.DisksClient.CreateOrUpdate(ctx, subsID, rg, options.DiskName, model); rerr != nil {
		return "", rerr.Error()
	}

	diskID := fmt.Sprintf(managedDiskPath, subsID, rg, options.DiskName)

	if options.SkipGetDiskOperation {
		klog.Warningf("azureDisk - GetDisk(%s, StorageAccountType:%s) is throttled, unable to confirm provisioningState in poll process", options.DiskName, options.StorageAccountType)
	} else {
		err = kwait.ExponentialBackoff(defaultBackOff, func() (bool, error) {
			provisionState, id, err := c.GetDisk(ctx, subsID, rg, options.DiskName)
			if err == nil {
				if id != "" {
					diskID = id
				}
			} else {
				// We are waiting for provisioningState==Succeeded
				// We don't want to hand-off managed disks to k8s while they are
				//still being provisioned, this is to avoid some race conditions
				return false, err
			}
			if strings.ToLower(provisionState) == "succeeded" {
				return true, nil
			}
			return false, nil
		})

		if err != nil {
			klog.Warningf("azureDisk - created new MD Name:%s StorageAccountType:%s Size:%v but was unable to confirm provisioningState in poll process", options.DiskName, options.StorageAccountType, options.SizeGB)
		}
	}

	klog.V(2).Infof("azureDisk - created new MD Name:%s StorageAccountType:%s Size:%v", options.DiskName, options.StorageAccountType, options.SizeGB)
	return diskID, nil
}

// DeleteManagedDisk : delete managed disk
func (c *ManagedDiskController) DeleteManagedDisk(ctx context.Context, diskURI string) error {
	resourceGroup, subsID, err := getInfoFromDiskURI(diskURI)
	if err != nil {
		return err
	}

	if _, ok := c.common.diskStateMap.Load(strings.ToLower(diskURI)); ok {
		return fmt.Errorf("failed to delete disk(%s) since it's in attaching or detaching state", diskURI)
	}

	diskName := path.Base(diskURI)
	disk, rerr := c.common.cloud.DisksClient.Get(ctx, subsID, resourceGroup, diskName)
	if rerr != nil {
		if rerr.HTTPStatusCode == http.StatusNotFound {
			klog.V(2).Infof("azureDisk - disk(%s) is already deleted", diskURI)
			return nil
		}
		// ignore GetDisk throttling
		if !rerr.IsThrottled() && !strings.Contains(rerr.RawError.Error(), consts.RateLimited) {
			return rerr.Error()
		}
	}
	if disk.ManagedBy != nil {
		return fmt.Errorf("disk(%s) already attached to node(%s), could not be deleted", diskURI, *disk.ManagedBy)
	}

	if rerr := c.common.cloud.DisksClient.Delete(ctx, subsID, resourceGroup, diskName); rerr != nil {
		return rerr.Error()
	}
	// We don't need poll here, k8s will immediately stop referencing the disk
	// the disk will be eventually deleted - cleanly - by ARM

	klog.V(2).Infof("azureDisk - deleted a managed disk: %s", diskURI)

	return nil
}

// GetDisk return: disk provisionState, diskID, error
func (c *ManagedDiskController) GetDisk(ctx context.Context, subsID, resourceGroup, diskName string) (string, string, error) {
	result, rerr := c.common.cloud.DisksClient.Get(ctx, subsID, resourceGroup, diskName)
	if rerr != nil {
		return "", "", rerr.Error()
	}

	if result.DiskProperties != nil && (*result.DiskProperties).ProvisioningState != nil {
		return *(*result.DiskProperties).ProvisioningState, *result.ID, nil
	}
	return "", "", nil
}

// ResizeDisk Expand the disk to new size
func (c *ManagedDiskController) ResizeDisk(ctx context.Context, diskURI string, oldSize resource.Quantity, newSize resource.Quantity, supportOnlineResize bool) (resource.Quantity, error) {
	diskName := path.Base(diskURI)
	resourceGroup, subsID, err := getInfoFromDiskURI(diskURI)
	if err != nil {
		return oldSize, err
	}

	result, rerr := c.common.cloud.DisksClient.Get(ctx, subsID, resourceGroup, diskName)
	if rerr != nil {
		return oldSize, rerr.Error()
	}

	if result.DiskProperties == nil || result.DiskProperties.DiskSizeGB == nil {
		return oldSize, fmt.Errorf("DiskProperties of disk(%s) is nil", diskName)
	}

	// Azure resizes in chunks of GiB (not GB)
	requestGiB, err := volumehelpers.RoundUpToGiBInt32(newSize)
	if err != nil {
		return oldSize, err
	}

	newSizeQuant := resource.MustParse(fmt.Sprintf("%dGi", requestGiB))

	klog.V(2).Infof("azureDisk - begin to resize disk(%s) with new size(%d), old size(%v)", diskName, requestGiB, oldSize)
	// If disk already of greater or equal size than requested we return
	if *result.DiskProperties.DiskSizeGB >= requestGiB {
		return newSizeQuant, nil
	}

	if !supportOnlineResize && result.DiskProperties.DiskState != compute.Unattached {
		return oldSize, fmt.Errorf("azureDisk - disk resize is only supported on Unattached disk, current disk state: %s, already attached to %s", result.DiskProperties.DiskState, to.String(result.ManagedBy))
	}

	diskParameter := compute.DiskUpdate{
		DiskUpdateProperties: &compute.DiskUpdateProperties{
			DiskSizeGB: &requestGiB,
		},
	}

	if rerr := c.common.cloud.DisksClient.Update(ctx, subsID, resourceGroup, diskName, diskParameter); rerr != nil {
		return oldSize, rerr.Error()
	}

	klog.V(2).Infof("azureDisk - resize disk(%s) with new size(%d) completed", diskName, requestGiB)
	return newSizeQuant, nil
}

// get resource group name, subs id from a managed disk URI, e.g. return {group-name}, {sub-id} according to
// /subscriptions/{sub-id}/resourcegroups/{group-name}/providers/microsoft.compute/disks/{disk-id}
// according to https://docs.microsoft.com/en-us/rest/api/compute/disks/get
func getInfoFromDiskURI(diskURI string) (string, string, error) {
	fields := strings.Split(diskURI, "/")
	if len(fields) != 9 || strings.ToLower(fields[3]) != "resourcegroups" {
		return "", "", fmt.Errorf("invalid disk URI: %s", diskURI)
	}
	return fields[4], fields[2], nil
}

// GetLabelsForVolume implements PVLabeler.GetLabelsForVolume
func (c *Cloud) GetLabelsForVolume(ctx context.Context, pv *v1.PersistentVolume) (map[string]string, error) {
	// Ignore if not AzureDisk.
	if pv.Spec.AzureDisk == nil {
		return nil, nil
	}

	// Ignore any volumes that are being provisioned
	if pv.Spec.AzureDisk.DiskName == cloudvolume.ProvisionedVolumeName {
		return nil, nil
	}

	return c.GetAzureDiskLabels(ctx, pv.Spec.AzureDisk.DataDiskURI)
}

// GetAzureDiskLabels gets availability zone labels for Azuredisk.
func (c *Cloud) GetAzureDiskLabels(ctx context.Context, diskURI string) (map[string]string, error) {
	// Get disk's resource group.
	diskName := path.Base(diskURI)
	resourceGroup, subsID, err := getInfoFromDiskURI(diskURI)
	if err != nil {
		klog.Errorf("Failed to get resource group for AzureDisk %s: %v", diskName, err)
		return nil, err
	}

	labels := map[string]string{
		consts.LabelFailureDomainBetaRegion: c.Location,
	}
	// no azure credential is set, return nil
	if c.DisksClient == nil {
		return labels, nil
	}
	disk, rerr := c.DisksClient.Get(ctx, subsID, resourceGroup, diskName)
	if rerr != nil {
		klog.Errorf("Failed to get information for AzureDisk %s: %v", diskName, rerr)
		return nil, rerr.Error()
	}

	// Check whether availability zone is specified.
	if disk.Zones == nil || len(*disk.Zones) == 0 {
		klog.V(4).Infof("Azure disk %s is not zoned", diskName)
		return labels, nil
	}

	zones := *disk.Zones
	zoneID, err := strconv.Atoi(zones[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse zone %v for AzureDisk %v: %w", zones, diskName, err)
	}

	zone := c.makeZone(c.Location, zoneID)
	klog.V(4).Infof("Got zone %s for Azure disk %s", zone, diskName)
	labels[consts.LabelFailureDomainBetaZone] = zone
	return labels, nil
}
