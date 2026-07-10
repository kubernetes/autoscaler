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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
)

const (
	maxLUN                 = 64 // max number of LUNs per VM
	errStatusCode400       = "statuscode=400"
	errInvalidParameter    = `code="invalidparameter"`
	errTargetInstanceIDs   = `target="instanceids"`
	sourceSnapshot         = "snapshot"
	sourceVolume           = "volume"
	attachDiskMapKeySuffix = "attachdiskmap"
	detachDiskMapKeySuffix = "detachdiskmap"

	// WriteAcceleratorEnabled support for Azure Write Accelerator on Azure Disks
	// https://docs.microsoft.com/azure/virtual-machines/windows/how-to-enable-write-accelerator
	WriteAcceleratorEnabled = "writeacceleratorenabled"
)

// ExtendedLocation contains additional info about the location of resources.
type ExtendedLocation struct {
	// Name - The name of the extended location.
	Name string `json:"name,omitempty"`
	// Type - The type of the extended location.
	Type string `json:"type,omitempty"`
}

func FilterNonExistingDisks(ctx context.Context, clientFactory azclient.ClientFactory, unfilteredDisks []*armcompute.DataDisk) []*armcompute.DataDisk {
	logger := log.FromContextOrBackground(ctx).WithName("FilterNonExistingDisks")
	filteredDisks := []*armcompute.DataDisk{}
	for _, disk := range unfilteredDisks {
		filter := false
		if disk.ManagedDisk != nil && disk.ManagedDisk.ID != nil {
			diSKURI := *disk.ManagedDisk.ID
			exist, err := checkDiskExists(ctx, clientFactory, diSKURI)
			if err != nil {
				logger.Error(err, "checkDiskExists failed", "diskURI", diSKURI)
			} else {
				// only filter disk when checkDiskExists returns <false, nil>
				filter = !exist
				if filter {
					logger.Error(nil, "disk does not exist, removed from data disk list", "diskURI", diSKURI)
				}
			}
		}

		if !filter {
			filteredDisks = append(filteredDisks, disk)
		}
	}
	return filteredDisks
}

func checkDiskExists(ctx context.Context, clientFactory azclient.ClientFactory, diSKURI string) (bool, error) {
	diskName := path.Base(diSKURI)
	resourceGroup, subsID, err := getInfoFromDiSKURI(diSKURI)
	if err != nil {
		return false, err
	}
	diskClient, err := clientFactory.GetDiskClientForSub(subsID)
	if err != nil {
		return false, err
	}

	_, err = diskClient.Get(ctx, resourceGroup, diskName)
	if err != nil {
		rerr := &azcore.ResponseError{}
		if errors.As(err, &rerr) {
			if rerr.StatusCode == http.StatusNotFound {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}

// get resource group name, subs id from a managed disk URI, e.g. return {group-name}, {sub-id} according to
// /subscriptions/{sub-id}/resourcegroups/{group-name}/providers/microsoft.compute/disks/{disk-id}
// according to https://docs.microsoft.com/en-us/rest/api/compute/disks/get
func getInfoFromDiSKURI(diSKURI string) (string, string, error) {
	fields := strings.Split(diSKURI, "/")
	if len(fields) != 9 || strings.ToLower(fields[3]) != "resourcegroups" {
		return "", "", fmt.Errorf("invalid disk URI: %s", diSKURI)
	}
	return fields[4], fields[2], nil
}
