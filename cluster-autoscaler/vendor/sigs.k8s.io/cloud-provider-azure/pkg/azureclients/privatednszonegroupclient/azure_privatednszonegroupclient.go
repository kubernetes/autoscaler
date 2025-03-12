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

package privatednszonegroupclient

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
	peResourceType                  = "Microsoft.Network/privateEndpoints"
	privateDNSZoneGroupResourceType = "privateDnsZoneGroups"
)

// Client implements privatednszonegroupclient client Interface.
type Client struct {
	armClient      armclient.Interface
	subscriptionID string
	cloudName      string

	// Rate limiting configures.
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter

	// ARM throttling configures.
	RetryAfterReader time.Time
	RetryAfterWriter time.Time
}

// New creates a new private dns zone group client with ratelimiting.
func New(config *azclients.ClientConfig) *Client {
	baseURI := config.ResourceManagerEndpoint
	authorizer := config.Authorizer
	apiVersion := APIVersion
	if strings.EqualFold(config.CloudName, AzureStackCloudName) && !config.DisableAzureStackCloud {
		klog.Warningf("Azure Stack is not supported for Private DNS Zone Group API")
	}
	armClient := armclient.New(authorizer, *config, baseURI, apiVersion)
	rateLimiterReader, rateLimiterWriter := azclients.NewRateLimiter(config.RateLimitConfig)

	if azclients.RateLimitEnabled(config.RateLimitConfig) {
		klog.V(2).Infof("Azure PrivateDNSZoneGroupClient (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure PrivateDNSZoneGroupClient (write ops) using rate limit config: QPS=%g, bucket=%d",
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

// CreateOrUpdate creates or updates a private DNS zone group.
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName, privateEndpointName, privateDNSZoneGroupName string, parameters network.PrivateDNSZoneGroup, etag string, waitForCompletion bool) *retry.Error {
	mc := metrics.NewMetricContext("private_dns_zone_group", "create_or_update", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "PrivateDNSZoneGroupCreateOrUpdate")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PrivateDNSZoneGroupCreateOrUpdate", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.createOrUpdatePrivateDNSZoneGroup(ctx, resourceGroupName, privateEndpointName, privateDNSZoneGroupName, parameters, etag, waitForCompletion)
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

// createOrUpdatePrivateDNSZoneGroup creates or updates a private DNS zone group.
func (c *Client) createOrUpdatePrivateDNSZoneGroup(ctx context.Context, resourceGroupName, privateEndpointName, privateDNSZoneGroupName string, parameters network.PrivateDNSZoneGroup, etag string, waitForCompletion bool) *retry.Error {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		peResourceType,
		privateEndpointName,
		privateDNSZoneGroupResourceType,
		privateDNSZoneGroupName)
	decorators := []autorest.PrepareDecorator{}
	if etag != "" {
		decorators = append(decorators, autorest.WithHeader("If-Match", autorest.String(etag)))
	}

	var response *http.Response
	var rerr *retry.Error
	if waitForCompletion {
		response, rerr = c.armClient.PutResource(ctx, resourceID, parameters, decorators...)
		defer c.armClient.CloseResponse(ctx, response)

	} else {
		_, rerr = c.armClient.PutResourceAsync(ctx, resourceID, parameters, decorators...)
	}
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privatednszonegroup.put.request", resourceID, rerr.Error())
		return rerr
	}
	if !waitForCompletion {
		return nil
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		_, rerr = c.createOrUpdateResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privatednszonegroup.put.respond", resourceID, rerr.Error())
			return rerr
		}
	}

	return nil
}

func (c *Client) createOrUpdateResponder(resp *http.Response) (*network.PrivateDNSZoneGroup, *retry.Error) {
	result := &network.PrivateDNSZoneGroup{}
	err := autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return result, retry.GetError(resp, err)
}

// Get gets a private dns zone group.
func (c *Client) Get(ctx context.Context, resourceGroupName, privateEndpointName, privateDNSZoneGroupName string) (network.PrivateDNSZoneGroup, *retry.Error) {
	mc := metrics.NewMetricContext("private_dns_zone_group", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return network.PrivateDNSZoneGroup{}, retry.GetRateLimitError(false, "PrivateDNSZoneGroupGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PrivateDNSZoneGroupGet", "client throttled", c.RetryAfterReader)
		return network.PrivateDNSZoneGroup{}, rerr
	}

	result, rerr := c.getPrivateDNSZoneGroup(ctx, resourceGroupName, privateEndpointName, privateDNSZoneGroupName)
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

// getPrivateDNSZoneGroup gets a private DNS zone group.
func (c *Client) getPrivateDNSZoneGroup(ctx context.Context, resourceGroupName, privateEndpointName, privateDNSZoneGroupName string) (network.PrivateDNSZoneGroup, *retry.Error) {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		peResourceType,
		privateEndpointName,
		privateDNSZoneGroupResourceType,
		privateDNSZoneGroupName,
	)
	result := network.PrivateDNSZoneGroup{}

	response, rerr := c.armClient.GetResource(ctx, resourceID)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privatednszonegroup.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privatednszonegroup.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}
