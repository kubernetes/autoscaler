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

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"

	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/fileclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// CreateFileShare creates a file share, using a matching storage account type, account kind, etc.
// storage account will be created if specified account is not found
func (az *Cloud) CreateFileShare(ctx context.Context, accountOptions *AccountOptions, shareOptions *fileclient.ShareOptions) (string, string, error) {
	if accountOptions == nil {
		return "", "", fmt.Errorf("account options is nil")
	}
	if shareOptions == nil {
		return "", "", fmt.Errorf("share options is nil")
	}
	if accountOptions.ResourceGroup == "" {
		accountOptions.ResourceGroup = az.ResourceGroup
	}
	if accountOptions.SubscriptionID == "" {
		accountOptions.SubscriptionID = az.SubscriptionID
	}

	accountOptions.EnableHTTPSTrafficOnly = true
	if shareOptions.Protocol == storage.EnabledProtocolsNFS {
		accountOptions.EnableHTTPSTrafficOnly = false
	}

	accountName, accountKey, err := az.EnsureStorageAccount(ctx, accountOptions, consts.FileShareAccountNamePrefix)
	if err != nil {
		return "", "", fmt.Errorf("could not get storage key for storage account %s: %w", accountOptions.Name, err)
	}

	if err := az.createFileShare(ctx, accountOptions.SubscriptionID, accountOptions.ResourceGroup, accountName, shareOptions); err != nil {
		return "", "", fmt.Errorf("failed to create share %s in account %s: %w", shareOptions.Name, accountName, err)
	}
	klog.V(4).Infof("created share %s in account %s", shareOptions.Name, accountOptions.Name)
	return accountName, accountKey, nil
}

// DeleteFileShare deletes a file share using storage account name and key
func (az *Cloud) DeleteFileShare(ctx context.Context, subsID, resourceGroup, accountName, shareName string) error {
	if err := az.deleteFileShare(ctx, subsID, resourceGroup, accountName, shareName); err != nil {
		return err
	}
	klog.V(4).Infof("share %s deleted", shareName)
	return nil
}

// ResizeFileShare resizes a file share
func (az *Cloud) ResizeFileShare(ctx context.Context, subsID, resourceGroup, accountName, name string, sizeGiB int) error {
	return az.resizeFileShare(ctx, subsID, resourceGroup, accountName, name, sizeGiB)
}

// GetFileShare gets a file share
func (az *Cloud) GetFileShare(ctx context.Context, subsID, resourceGroupName, accountName, name string) (storage.FileShare, error) {
	return az.getFileShare(ctx, subsID, resourceGroupName, accountName, name)
}
