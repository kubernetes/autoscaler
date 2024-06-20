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

package privatednsclient

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
	privateDNSZoneResourceType = "Microsoft.Network/privateDnsZones"
)

// Client implements privatednsclient Interface.
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

// New creates a new private dns client with ratelimiting.
func New(config *azclients.ClientConfig) *Client {
	baseURI := config.ResourceManagerEndpoint
	authorizer := config.Authorizer
	apiVersion := APIVersion
	if strings.EqualFold(config.CloudName, AzureStackCloudName) && !config.DisableAzureStackCloud {
		klog.Warningf("Azure Stack is not supported for Private DNS Zone API")
	}
	armClient := armclient.New(authorizer, *config, baseURI, apiVersion)
	rateLimiterReader, rateLimiterWriter := azclients.NewRateLimiter(config.RateLimitConfig)

	if azclients.RateLimitEnabled(config.RateLimitConfig) {
		klog.V(2).Infof("Azure PrivateDNSZoneClient (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure PrivateDNSZoneClient (write ops) using rate limit config: QPS=%g, bucket=%d",
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

// CreateOrUpdate creates or updates a private DNS zone.
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName, privateZoneName string, parameters privatedns.PrivateZone, etag string, waitForCompletion bool) *retry.Error {
	mc := metrics.NewMetricContext("private_dns_zone", "create_or_update", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "PrivateDNSZoneCreateOrUpdate")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PrivateDNSZoneCreateOrUpdate", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.createOrUpdatePrivateDNSZone(ctx, resourceGroupName, privateZoneName, parameters, etag, waitForCompletion)
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

// createOrUpdatePrivateDNSZone creates or updates a private DNS zone.
func (c *Client) createOrUpdatePrivateDNSZone(ctx context.Context, resourceGroupName, privateDNSZoneName string, parameters privatedns.PrivateZone, etag string, waitForCompletion bool) *retry.Error {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		privateDNSZoneResourceType,
		privateDNSZoneName,
	)
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
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privatednszone.put.request", resourceID, rerr.Error())
		return rerr
	}
	if !waitForCompletion {
		return nil
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		_, rerr = c.createOrUpdateResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privatednszone.put.respond", resourceID, rerr.Error())
			return rerr
		}
	}

	return nil
}

func (c *Client) createOrUpdateResponder(resp *http.Response) (*privatedns.PrivateZone, *retry.Error) {
	result := &privatedns.PrivateZone{}
	err := autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return result, retry.GetError(resp, err)
}

// Get gets a private dns zone.
func (c *Client) Get(ctx context.Context, resourceGroupName, privateZoneName string) (privatedns.PrivateZone, *retry.Error) {
	mc := metrics.NewMetricContext("private_dns_zones", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return privatedns.PrivateZone{}, retry.GetRateLimitError(false, "PrivateDNSZoneGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PrivateDNSZoneGet", "client throttled", c.RetryAfterReader)
		return privatedns.PrivateZone{}, rerr
	}

	result, rerr := c.getPrivateDNSZone(ctx, resourceGroupName, privateZoneName)
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

// getPrivateDNSZone gets a private DNS zone.
func (c *Client) getPrivateDNSZone(ctx context.Context, resourceGroupName, privateDNSZoneName string) (privatedns.PrivateZone, *retry.Error) {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		privateDNSZoneResourceType,
		privateDNSZoneName,
	)
	result := privatedns.PrivateZone{}

	response, rerr := c.armClient.GetResource(ctx, resourceID)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privatednszone.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "privatednszone.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}
