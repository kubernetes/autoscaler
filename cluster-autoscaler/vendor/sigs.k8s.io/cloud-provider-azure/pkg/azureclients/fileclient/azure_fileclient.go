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

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"
	"github.com/Azure/go-autorest/autorest"

	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// Client implements the azure file client interface
type Client struct {
	fileSharesClient   storage.FileSharesClient
	fileServicesClient storage.FileServicesClient

	subscriptionID string
	baseURI        string
	authorizer     autorest.Authorizer
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
func New(config *azclients.ClientConfig) Interface {
	baseURI := config.ResourceManagerEndpoint
	authorizer := config.Authorizer
	fileSharesClient := storage.NewFileSharesClientWithBaseURI(baseURI, config.SubscriptionID)
	fileSharesClient.Authorizer = authorizer

	fileServicesClient := storage.NewFileServicesClientWithBaseURI(baseURI, config.SubscriptionID)
	fileServicesClient.Authorizer = authorizer
	return &Client{
		fileSharesClient:   fileSharesClient,
		fileServicesClient: fileServicesClient,
		subscriptionID:     config.SubscriptionID,
		baseURI:            baseURI,
		authorizer:         authorizer,
	}
}

func (c *Client) WithSubscriptionID(subscriptionID string) Interface {
	if subscriptionID == "" || subscriptionID == c.subscriptionID {
		return c
	}

	return New(&azclients.ClientConfig{
		SubscriptionID:          subscriptionID,
		ResourceManagerEndpoint: c.baseURI,
		Authorizer:              c.authorizer,
	})
}

// CreateFileShare creates a file share
// expand - optional, used to expand the properties within share's properties. Valid values are: snapshots.
// Should be passed as a string with delimiter ','
func (c *Client) CreateFileShare(ctx context.Context, resourceGroupName, accountName string, shareOptions *ShareOptions, expand string) (storage.FileShare, error) {
	mc := metrics.NewMetricContext("file_shares", "create", resourceGroupName, c.subscriptionID, "")

	if shareOptions == nil {
		return storage.FileShare{}, fmt.Errorf("share options is nil")
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
	FileShare, err := c.fileSharesClient.Create(ctx, resourceGroupName, accountName, shareOptions.Name, fileShare, expand)
	var rerr *retry.Error
	if err != nil {
		rerr = &retry.Error{
			RawError: err,
		}
	}
	mc.Observe(rerr)

	return FileShare, err
}

// DeleteFileShare deletes a file share
// xMsSnapshot - optional, used to delete a snapshot.
// It is a DateTime value that uniquely identifies the share snapshot. e.g. "2017-05-10T17:52:33.9551861Z"
func (c *Client) DeleteFileShare(ctx context.Context, resourceGroupName, accountName, name, xMsSnapshot string) error {
	mc := metrics.NewMetricContext("file_shares", "delete", resourceGroupName, c.subscriptionID, "")

	_, err := c.fileSharesClient.Delete(ctx, resourceGroupName, accountName, name, xMsSnapshot, "")
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
func (c *Client) ResizeFileShare(ctx context.Context, resourceGroupName, accountName, name string, sizeGiB int) error {
	mc := metrics.NewMetricContext("file_shares", "resize", resourceGroupName, c.subscriptionID, "")
	var rerr *retry.Error

	quota := int32(sizeGiB)

	share, err := c.fileSharesClient.Get(ctx, resourceGroupName, accountName, name, "stats", "")
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
	_, err = c.fileSharesClient.Update(ctx, resourceGroupName, accountName, name, share)
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
// xMsSnapshot - optional, used to retrieve properties of a snapshot.
// It is a DateTime value that uniquely identifies the share snapshot. e.g. "2017-05-10T17:52:33.9551861Z"
func (c *Client) GetFileShare(ctx context.Context, resourceGroupName, accountName, name, xMsSnapshot string) (storage.FileShare, error) {
	mc := metrics.NewMetricContext("file_shares", "get", resourceGroupName, c.subscriptionID, "")

	result, err := c.fileSharesClient.Get(ctx, resourceGroupName, accountName, name, "stats", xMsSnapshot)
	var rerr *retry.Error
	if err != nil {
		rerr = &retry.Error{
			RawError: err,
		}
	}
	mc.Observe(rerr)

	return result, err
}

// ListFileShare gets a file share list
// expand - optional, used to expand the properties within share's properties. Valid values are: deleted,
// snapshots. Should be passed as a string with delimiter ','
func (c *Client) ListFileShare(ctx context.Context, resourceGroupName, accountName, filter, expand string) ([]storage.FileShareItem, error) {
	mc := metrics.NewMetricContext("file_shares", "list", resourceGroupName, c.subscriptionID, "")

	page, err := c.fileSharesClient.List(ctx, resourceGroupName, accountName, "", filter, expand)
	var rerr *retry.Error
	if err != nil {
		rerr = &retry.Error{
			RawError: err,
		}
	}
	mc.Observe(rerr)

	result := make([]storage.FileShareItem, 0)

	for {
		result = append(result, page.Values()...)

		// Abort the loop when there's no nextLink in the response.
		if pointer.StringDeref(page.Response().NextLink, "") == "" {
			break
		}

		if err = page.NextWithContext(ctx); err != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "snapshot.list.next", resourceGroupName, err)
			mc.Observe(retry.GetError(page.Response().Response.Response, err))
		}
	}

	return result, err
}

// GetServiceProperties get service properties
func (c *Client) GetServiceProperties(ctx context.Context, resourceGroupName, accountName string) (storage.FileServiceProperties, error) {
	return c.fileServicesClient.GetServiceProperties(ctx, resourceGroupName, accountName)
}

// SetServiceProperties set service properties
func (c *Client) SetServiceProperties(ctx context.Context, resourceGroupName, accountName string, parameters storage.FileServiceProperties) (storage.FileServiceProperties, error) {
	return c.fileServicesClient.SetServiceProperties(ctx, resourceGroupName, accountName, parameters)
}
