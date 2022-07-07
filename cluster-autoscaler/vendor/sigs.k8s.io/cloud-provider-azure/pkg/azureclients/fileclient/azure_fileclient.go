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

package fileclient

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-02-01/storage"

	"k8s.io/klog/v2"

	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// Client implements the azure file client interface
type Client struct {
	fileSharesClient   storage.FileSharesClient
	fileServicesClient storage.FileServicesClient

	subscriptionID string
}

// ShareOptions contains the fields which are used to create file share.
type ShareOptions struct {
	Name       string
	Protocol   storage.EnabledProtocols
	RequestGiB int
	// supported values: ""(by default), "TransactionOptimized", "Cool", "Hot", "Premium"
	AccessTier string
	// supported values: ""(by default), "AllSquash", "NoRootSquash", "RootSquash"
	RootSquash string
	// Metadata - A name-value pair to associate with the share as metadata.
	Metadata map[string]*string
}

// New creates a azure file client
func New(config *azclients.ClientConfig) *Client {
	fileSharesClient := storage.NewFileSharesClientWithBaseURI(config.ResourceManagerEndpoint, config.SubscriptionID)
	fileSharesClient.Authorizer = config.Authorizer

	fileServicesClient := storage.NewFileServicesClientWithBaseURI(config.ResourceManagerEndpoint, config.SubscriptionID)
	fileServicesClient.Authorizer = config.Authorizer
	return &Client{
		fileSharesClient:   fileSharesClient,
		fileServicesClient: fileServicesClient,
		subscriptionID:     config.SubscriptionID,
	}
}

// CreateFileShare creates a file share
func (c *Client) CreateFileShare(resourceGroupName, accountName string, shareOptions *ShareOptions) error {
	mc := metrics.NewMetricContext("file_shares", "create", resourceGroupName, c.subscriptionID, "")

	if shareOptions == nil {
		return fmt.Errorf("share options is nil")
	}
	quota := int32(shareOptions.RequestGiB)
	fileShareProperties := &storage.FileShareProperties{
		ShareQuota: &quota,
	}
	if shareOptions.Protocol == storage.EnabledProtocolsNFS {
		fileShareProperties.EnabledProtocols = shareOptions.Protocol
	}
	if shareOptions.AccessTier != "" {
		fileShareProperties.AccessTier = storage.ShareAccessTier(shareOptions.AccessTier)
	}
	if shareOptions.RootSquash != "" {
		fileShareProperties.RootSquash = storage.RootSquashType(shareOptions.RootSquash)
	}
	if shareOptions.Metadata != nil {
		fileShareProperties.Metadata = shareOptions.Metadata
	}
	fileShare := storage.FileShare{
		Name:                &shareOptions.Name,
		FileShareProperties: fileShareProperties,
	}
	_, err := c.fileSharesClient.Create(context.Background(), resourceGroupName, accountName, shareOptions.Name, fileShare, "")
	var rerr *retry.Error
	if err != nil {
		rerr = &retry.Error{
			RawError: err,
		}
	}
	mc.Observe(rerr)

	return err
}

// DeleteFileShare deletes a file share
func (c *Client) DeleteFileShare(resourceGroupName, accountName, name string) error {
	mc := metrics.NewMetricContext("file_shares", "delete", resourceGroupName, c.subscriptionID, "")

	_, err := c.fileSharesClient.Delete(context.Background(), resourceGroupName, accountName, name, "")
	var rerr *retry.Error
	if err != nil {
		rerr = &retry.Error{
			RawError: err,
		}
	}
	mc.Observe(rerr)

	return err
}

// ResizeFileShare resizes a file share
func (c *Client) ResizeFileShare(resourceGroupName, accountName, name string, sizeGiB int) error {
	mc := metrics.NewMetricContext("file_shares", "resize", resourceGroupName, c.subscriptionID, "")
	var rerr *retry.Error

	quota := int32(sizeGiB)

	share, err := c.fileSharesClient.Get(context.Background(), resourceGroupName, accountName, name, storage.GetShareExpandStats, "")
	if err != nil {
		rerr = &retry.Error{
			RawError: err,
		}
		mc.Observe(rerr)
		return fmt.Errorf("failed to get file share (%s): %w", name, err)
	}
	if *share.FileShareProperties.ShareQuota >= quota {
		klog.Warningf("file share size(%dGi) is already greater or equal than requested size(%dGi), accountName: %s, shareName: %s",
			share.FileShareProperties.ShareQuota, sizeGiB, accountName, name)
		return nil
	}

	share.FileShareProperties.ShareQuota = &quota
	_, err = c.fileSharesClient.Update(context.Background(), resourceGroupName, accountName, name, share)
	if err != nil {
		rerr = &retry.Error{
			RawError: err,
		}
		mc.Observe(rerr)
		return fmt.Errorf("failed to update quota on file share(%s), err: %w", name, err)
	}

	mc.Observe(rerr)
	klog.V(4).Infof("resize file share completed, resourceGroupName(%s), accountName: %s, shareName: %s, sizeGiB: %d", resourceGroupName, accountName, name, sizeGiB)

	return nil
}

// GetFileShare gets a file share
func (c *Client) GetFileShare(resourceGroupName, accountName, name string) (storage.FileShare, error) {
	mc := metrics.NewMetricContext("file_shares", "get", resourceGroupName, c.subscriptionID, "")

	result, err := c.fileSharesClient.Get(context.Background(), resourceGroupName, accountName, name, storage.GetShareExpandStats, "")
	var rerr *retry.Error
	if err != nil {
		rerr = &retry.Error{
			RawError: err,
		}
	}
	mc.Observe(rerr)

	return result, err
}

// GetServiceProperties get service properties
func (c *Client) GetServiceProperties(resourceGroupName, accountName string) (storage.FileServiceProperties, error) {
	return c.fileServicesClient.GetServiceProperties(context.Background(), resourceGroupName, accountName)
}

// SetServiceProperties set service properties
func (c *Client) SetServiceProperties(resourceGroupName, accountName string, parameters storage.FileServiceProperties) (storage.FileServiceProperties, error) {
	return c.fileServicesClient.SetServiceProperties(context.Background(), resourceGroupName, accountName, parameters)
}
