/*
Copyright 2022 The Kubernetes Authors.

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

package blobclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"

	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/armclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

var _ Interface = &Client{}

// Client implements the blobclient interface
type Client struct {
	blobServicesClient storage.BlobServicesClient
	armClient          armclient.Interface
	subscriptionID     string
	cloudName          string
	baseURI            string
	authorizer         autorest.Authorizer

	// Rate limiting configures.
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter

	// ARM throttling configures.
	RetryAfterReader time.Time
	RetryAfterWriter time.Time

	// now allows for injecting fake or real now time into code
	now func() time.Time
}

// New creates a blobContainersClient
func New(config *azclients.ClientConfig) *Client {
	baseURI := config.ResourceManagerEndpoint
	authorizer := config.Authorizer
	apiVersion := APIVersion

	blobServicesClient := storage.NewBlobServicesClientWithBaseURI(baseURI, config.SubscriptionID)
	blobServicesClient.Authorizer = authorizer

	if strings.EqualFold(config.CloudName, AzureStackCloudName) && !config.DisableAzureStackCloud {
		apiVersion = AzureStackCloudAPIVersion
	}

	klog.V(2).Infof("Azure BlobClient using API version: %s", apiVersion)
	armClient := armclient.New(authorizer, *config, baseURI, apiVersion)
	rateLimiterReader, rateLimiterWriter := azclients.NewRateLimiter(config.RateLimitConfig)

	if azclients.RateLimitEnabled(config.RateLimitConfig) {
		klog.V(2).Infof("Azure BlobClient (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure BlobClient (write ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPSWrite,
			config.RateLimitConfig.CloudProviderRateLimitBucketWrite)
	}

	client := &Client{
		blobServicesClient: blobServicesClient,
		armClient:          armClient,
		rateLimiterReader:  rateLimiterReader,
		rateLimiterWriter:  rateLimiterWriter,
		subscriptionID:     config.SubscriptionID,
		cloudName:          config.CloudName,
		now:                time.Now,
		baseURI:            baseURI,
		authorizer:         authorizer,
	}

	return client
}

// CreateContainer creates a blob container
func (c *Client) CreateContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string, parameters storage.BlobContainer) *retry.Error {
	if subsID == "" {
		subsID = c.subscriptionID
	}

	mc := metrics.NewMetricContext("blob_container", "create", resourceGroupName, subsID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "CreateBlobContainer")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(c.now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("CreateBlobContainer", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.createContainer(ctx, subsID, resourceGroupName, accountName, containerName, parameters)
	mc.Observe(rerr)
	if rerr != nil {
		if rerr.IsThrottled() {
			// Update RetryAfterReader so that no more requests would be sent until RetryAfter expires.
			c.RetryAfterWriter = rerr.RetryAfter
		}

		return rerr
	}

	return nil
}

func (c *Client) createContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string, parameters storage.BlobContainer) *retry.Error {
	// resourceID format: "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/blobServices/default/containers/{containerName}"
	resourceID := armclient.GetChildResourceID(
		subsID,
		resourceGroupName,
		"Microsoft.Storage/storageAccounts",
		accountName,
		"blobServices/default/containers",
		containerName,
	)

	response, rerr := c.armClient.PutResource(ctx, resourceID, parameters)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "blob_container.put.request", resourceID, rerr.Error())
		return rerr
	}

	container := storage.BlobContainer{}
	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&container))
	container.Response = autorest.Response{Response: response}

	return retry.GetError(response, err)
}

// DeleteContainer deletes a blob container
func (c *Client) DeleteContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string) *retry.Error {
	if subsID == "" {
		subsID = c.subscriptionID
	}

	mc := metrics.NewMetricContext("blob_container", "delete", resourceGroupName, subsID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "BlobContainerDelete")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(c.now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("BlobContainerDelete", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.deleteContainer(ctx, subsID, resourceGroupName, accountName, containerName)
	mc.Observe(rerr)
	if rerr != nil {
		if rerr.IsThrottled() {
			// Update RetryAfterReader so that no more requests would be sent until RetryAfter expires.
			c.RetryAfterWriter = rerr.RetryAfter
		}

		return rerr
	}

	return nil
}

func (c *Client) deleteContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string) *retry.Error {
	// resourceID format: "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/blobServices/default/containers/{containerName}"
	resourceID := armclient.GetChildResourceID(
		subsID,
		resourceGroupName,
		"Microsoft.Storage/storageAccounts",
		accountName,
		"blobServices/default/containers",
		containerName,
	)

	return c.armClient.DeleteResource(ctx, resourceID)
}

// GetContainer gets a blob container
func (c *Client) GetContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string) (storage.BlobContainer, *retry.Error) {
	if subsID == "" {
		subsID = c.subscriptionID
	}

	mc := metrics.NewMetricContext("blob_container", "get", resourceGroupName, subsID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return storage.BlobContainer{}, retry.GetRateLimitError(false, "GetBlobContainer")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(c.now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("GetBlobContainer", "client throttled", c.RetryAfterReader)
		return storage.BlobContainer{}, rerr
	}

	container, rerr := c.getContainer(ctx, subsID, resourceGroupName, accountName, containerName)
	mc.Observe(rerr)
	if rerr != nil {
		if rerr.IsThrottled() {
			// Update RetryAfterReader so that no more requests would be sent until RetryAfter expires.
			c.RetryAfterReader = rerr.RetryAfter
		}

		return container, rerr
	}

	return container, nil
}

func (c *Client) getContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string) (storage.BlobContainer, *retry.Error) {
	// resourceID format: "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Storage/storageAccounts/{accountName}/blobServices/default/containers/{containerName}"
	resourceID := armclient.GetChildResourceID(
		subsID,
		resourceGroupName,
		"Microsoft.Storage/storageAccounts",
		accountName,
		"blobServices/default/containers",
		containerName,
	)

	response, rerr := c.armClient.GetResource(ctx, resourceID)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "blob_container.get.request", resourceID, rerr.Error())
		return storage.BlobContainer{}, rerr
	}

	container := storage.BlobContainer{}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&container))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "blob_container.get.request", resourceID, err)
		return container, retry.GetError(response, err)
	}

	container.Response = autorest.Response{Response: response}
	return container, nil
}

func (c *Client) GetServiceProperties(ctx context.Context, subsID, resourceGroupName, accountName string) (storage.BlobServiceProperties, error) {
	blobServicesClient := c.blobServicesClient
	if subsID != c.subscriptionID {
		blobServicesClient = storage.NewBlobServicesClientWithBaseURI(c.baseURI, c.subscriptionID)
		blobServicesClient.Authorizer = c.authorizer
	}
	return blobServicesClient.GetServiceProperties(ctx, resourceGroupName, accountName)
}

func (c *Client) SetServiceProperties(ctx context.Context, subsID, resourceGroupName, accountName string, parameters storage.BlobServiceProperties) (storage.BlobServiceProperties, error) {
	blobServicesClient := c.blobServicesClient
	if subsID != c.subscriptionID {
		blobServicesClient = storage.NewBlobServicesClientWithBaseURI(c.baseURI, c.subscriptionID)
		blobServicesClient.Authorizer = c.authorizer
	}
	return blobServicesClient.SetServiceProperties(ctx, resourceGroupName, accountName, parameters)
}
