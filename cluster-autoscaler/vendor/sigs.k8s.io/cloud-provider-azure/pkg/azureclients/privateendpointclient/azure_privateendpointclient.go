/*
Copyright 2021 The Kubernetes Authors.

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

// Package privateendpointclient implements the client for Private Endpoint.
package privateendpointclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
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

const (
	PEResourceType = "Microsoft.Network/privateEndpoints"
)

// Client implements privateendpointclient Interface.
type Client struct {
	armClient      armclient.Interface
	cloudName      string
	subscriptionID string

	// Rate limiting configures.
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter

	// ARM throttling configures.
	RetryAfterReader time.Time
	RetryAfterWriter time.Time
}

// New creates a new private endpoint client.
func New(config *azclients.ClientConfig) *Client {
	apiVersion := APIVersion
	if strings.EqualFold(config.CloudName, AzureStackCloudName) && !config.DisableAzureStackCloud {
		// To check whether Azure Stack Cloud supports a resource and the supported API version, refer to:
		// https://docs.microsoft.com/en-us/azure-stack/user/azure-stack-profiles-azure-resource-manager-versions?view=azs-2108
		klog.Warningf("Azure Stack is not supported for Private Endpoint API")
	}
	armClient := armclient.New(config.Authorizer, *config, config.ResourceManagerEndpoint, apiVersion)

	rateLimiterReader, rateLimiterWriter := azclients.NewRateLimiter(config.RateLimitConfig)
	if azclients.RateLimitEnabled(config.RateLimitConfig) {
		klog.V(2).Infof("Azure PrivateEndpointsClient (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure PrivateEndpointsClient (write ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPSWrite,
			config.RateLimitConfig.CloudProviderRateLimitBucketWrite)
	}

	client := &Client{
		armClient:         armClient,
		rateLimiterReader: rateLimiterReader,
		rateLimiterWriter: rateLimiterWriter,
		subscriptionID:    config.SubscriptionID,
		cloudName:         config.CloudName,
	}
	return client
}

// CreateOrUpdate creates or updates a private endpoint.
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, endpointName string, privateEndpoint network.PrivateEndpoint, etag string, waitForCompletion bool) *retry.Error {
	mc := metrics.NewMetricContext("private_endpoints", "create_or_update", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "PrivateEndpointCreateOrUpdate")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PrivateEndpointCreateOrUpdate", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.createOrUpdatePE(ctx, resourceGroupName, endpointName, privateEndpoint, etag, waitForCompletion)
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

func (c *Client) createOrUpdatePE(ctx context.Context, resourceGroupName string, endpointName string, parameters network.PrivateEndpoint, etag string, waitForCompletion bool) *retry.Error {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		PEResourceType,
		endpointName,
	)
	decorators := []autorest.PrepareDecorator{}
	if etag != "" {
		decorators = append(decorators, autorest.WithHeader("If-Match", autorest.String(etag)))
	}

	var response *http.Response
	var rerr *retry.Error
	if waitForCompletion {
		response, rerr = c.armClient.PutResource(ctx, resourceID, parameters, decorators...)
	} else {
		_, rerr = c.armClient.PutResourceAsync(ctx, resourceID, parameters, decorators...)
	}
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privateendpoint.put.request", resourceID, rerr.Error())
		return rerr
	}

	if !waitForCompletion {
		return nil
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		_, rerr = c.createOrUpdateResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privateendpoint.put.request", resourceID, rerr.Error())
			return rerr
		}
	}
	return nil
}

func (c *Client) createOrUpdateResponder(resp *http.Response) (*network.PrivateEndpoint, *retry.Error) {
	result := &network.PrivateEndpoint{}
	err := autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return result, retry.GetError(resp, err)
}

// Get gets the private endpoint
func (c *Client) Get(ctx context.Context, resourceGroupName string, privateEndpointName string, expand string) (network.PrivateEndpoint, *retry.Error) {
	mc := metrics.NewMetricContext("private_endpoints", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return network.PrivateEndpoint{}, retry.GetRateLimitError(false, "PrivateEndpointGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PrivateEndpointGet", "client throttled", c.RetryAfterReader)
		return network.PrivateEndpoint{}, rerr
	}
	result, rerr := c.getPE(ctx, resourceGroupName, privateEndpointName, expand)

	mc.Observe(rerr)
	if rerr != nil {
		if rerr.IsThrottled() {
			// Update RetryAfterReader so that no more requests would be sent until RetryAfter expires.
			c.RetryAfterReader = rerr.RetryAfter
		}
		return result, rerr
	}
	return result, nil
}

// getPE gets a private endpoint.
func (c *Client) getPE(ctx context.Context, resourceGroupName string, privateEndpointName string, expand string) (network.PrivateEndpoint, *retry.Error) {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		PEResourceType,
		privateEndpointName,
	)
	result := network.PrivateEndpoint{}
	response, rerr := c.armClient.GetResourceWithExpandQuery(ctx, resourceID, expand)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privateendpoint.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privateendpoint.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}
