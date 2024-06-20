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

package virtualnetworklinksclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"
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
	privateDNSZoneResourceType     = "Microsoft.Network/privateDnsZones"
	virtualNetworkLinkResourceType = "virtualNetworkLinks"
)

// Client implements virtualnetworklinksclient Interface.
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

// New creates a new virtualnetworklinks client.
func New(config *azclients.ClientConfig) *Client {
	apiVersion := APIVersion
	if strings.EqualFold(config.CloudName, AzureStackCloudName) && !config.DisableAzureStackCloud {
		klog.Warningf("Azure Stack is not supported for Virtual Network Link API")
	}
	armClient := armclient.New(config.Authorizer, *config, config.ResourceManagerEndpoint, apiVersion)

	rateLimiterReader, rateLimiterWriter := azclients.NewRateLimiter(config.RateLimitConfig)
	if azclients.RateLimitEnabled(config.RateLimitConfig) {
		klog.V(2).Infof("Azure VirtualNetworkLinksClient (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure VirtualNetworkLinksClient (write ops) using rate limit config: QPS=%g, bucket=%d",
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

// CreateOrUpdate creates or updates a virtual network link
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, privateZoneName string, virtualNetworkLinkName string, parameters privatedns.VirtualNetworkLink, etag string, waitForCompletion bool) *retry.Error {
	mc := metrics.NewMetricContext("virtual_network_links", "create_or_update", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "VirtualNetworkLinkCreateOrUpdate")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VirtualNetworkLinkCreateOrUpdate", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.createOrUpdateVirtualNetworkLink(ctx, resourceGroupName, privateZoneName, virtualNetworkLinkName, parameters, etag, waitForCompletion)
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

// createOrUpdateVirtualNetworkLink creates or updates a virtual network link.
func (c *Client) createOrUpdateVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string, parameters privatedns.VirtualNetworkLink, etag string, waitForCompletion bool) *retry.Error {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		privateDNSZoneResourceType,
		privateZoneName,
		virtualNetworkLinkResourceType,
		virtualNetworkLinkName,
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
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "virtualnetworklink.put.request", resourceID, rerr.Error())
		return rerr
	}

	if !waitForCompletion {
		return nil
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		_, rerr = c.createOrUpdateResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "virtualnetworklink.put.request", resourceID, rerr.Error())
			return rerr
		}
	}
	return nil
}

func (c *Client) createOrUpdateResponder(resp *http.Response) (*privatedns.VirtualNetworkLink, *retry.Error) {
	result := &privatedns.VirtualNetworkLink{}
	err := autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return result, retry.GetError(resp, err)
}

// Get gets a virtual network link
func (c *Client) Get(ctx context.Context, resourceGroupName string, privateZoneName string, virtualNetworkLinkName string) (result privatedns.VirtualNetworkLink, err *retry.Error) {
	mc := metrics.NewMetricContext("virtual_network_links", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return privatedns.VirtualNetworkLink{}, retry.GetRateLimitError(false, "VirtualNetworkLinkGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VirtualNetworkLinkGet", "client throttled", c.RetryAfterReader)
		return privatedns.VirtualNetworkLink{}, rerr
	}
	result, rerr := c.getVirtualNetworkLink(ctx, resourceGroupName, privateZoneName, virtualNetworkLinkName)

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

// getVirtualNetworkLink gets a virtual network link.
func (c *Client) getVirtualNetworkLink(ctx context.Context, resourceGroupName string, privateZoneName string, virtualNetworkLinkName string) (privatedns.VirtualNetworkLink, *retry.Error) {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		privateDNSZoneResourceType,
		privateZoneName,
		virtualNetworkLinkResourceType,
		virtualNetworkLinkName,
	)
	result := privatedns.VirtualNetworkLink{}
	response, rerr := c.armClient.GetResourceWithExpandQuery(ctx, resourceID, "")
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "virtualnetworklink.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "virtualnetworklink.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}
