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

package vmasclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"

	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/armclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

var _ Interface = &Client{}

const vmasResourceType = "Microsoft.Compute/availabilitySets"

// Client implements VMAS client Interface.
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

// New creates a new VMAS client with ratelimiting.
func New(config *azclients.ClientConfig) *Client {
	baseURI := config.ResourceManagerEndpoint
	authorizer := config.Authorizer
	apiVersion := APIVersion
	if strings.EqualFold(config.CloudName, AzureStackCloudName) && !config.DisableAzureStackCloud {
		apiVersion = AzureStackCloudAPIVersion
	}
	armClient := armclient.New(authorizer, *config, baseURI, apiVersion)
	rateLimiterReader, rateLimiterWriter := azclients.NewRateLimiter(config.RateLimitConfig)

	if azclients.RateLimitEnabled(config.RateLimitConfig) {
		klog.V(2).Infof("Azure AvailabilitySetsClient (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure AvailabilitySetsClient  (write ops) using rate limit config: QPS=%g, bucket=%d",
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

// Get gets a AvailabilitySet.
func (c *Client) Get(ctx context.Context, resourceGroupName string, vmasName string) (compute.AvailabilitySet, *retry.Error) {
	mc := metrics.NewMetricContext("vmas", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return compute.AvailabilitySet{}, retry.GetRateLimitError(false, "VMASGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VMASGet", "client throttled", c.RetryAfterReader)
		return compute.AvailabilitySet{}, rerr
	}

	result, rerr := c.getVMAS(ctx, resourceGroupName, vmasName)
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

// getVMAS gets a AvailabilitySet.
func (c *Client) getVMAS(ctx context.Context, resourceGroupName string, vmasName string) (compute.AvailabilitySet, *retry.Error) {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		vmasResourceType,
		vmasName,
	)
	result := compute.AvailabilitySet{}

	response, rerr := c.armClient.GetResource(ctx, resourceID)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmas.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmas.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}

// List gets a list of AvailabilitySets in the resource group.
func (c *Client) List(ctx context.Context, resourceGroupName string) ([]compute.AvailabilitySet, *retry.Error) {
	mc := metrics.NewMetricContext("vmas", "list", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return nil, retry.GetRateLimitError(false, "VMASList")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VMASList", "client throttled", c.RetryAfterReader)
		return nil, rerr
	}

	result, rerr := c.listVMAS(ctx, resourceGroupName)
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

// listVMAS gets a list of AvailabilitySets in the resource group.
func (c *Client) listVMAS(ctx context.Context, resourceGroupName string) ([]compute.AvailabilitySet, *retry.Error) {
	resourceID := armclient.GetResourceListID(c.subscriptionID, resourceGroupName, vmasResourceType)
	result := make([]compute.AvailabilitySet, 0)
	page := &AvailabilitySetListResultPage{}
	page.fn = c.listNextResults

	resp, rerr := c.armClient.GetResource(ctx, resourceID)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmas.list.request", resourceID, rerr.Error())
		return result, rerr
	}

	var err error
	page.vmaslr, err = c.listResponder(resp)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmas.list.respond", resourceID, err)
		return result, retry.GetError(resp, err)
	}

	for {
		result = append(result, page.Values()...)

		// Abort the loop when there's no nextLink in the response.
		if pointer.StringDeref(page.Response().NextLink, "") == "" {
			break
		}

		if err = page.NextWithContext(ctx); err != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmas.list.next", resourceID, err)
			return result, retry.GetError(page.Response().Response.Response, err)
		}
	}

	return result, nil
}

func (c *Client) listResponder(resp *http.Response) (result compute.AvailabilitySetListResult, err error) {
	err = autorest.Respond(
		resp,
		autorest.ByIgnoring(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return
}

// availabilitySetListResultPreparer prepares a request to retrieve the next set of results.
// It returns nil if no more results exist.
func (c *Client) availabilitySetListResultPreparer(ctx context.Context, vmaslr compute.AvailabilitySetListResult) (*http.Request, error) {
	if vmaslr.NextLink == nil || len(pointer.StringDeref(vmaslr.NextLink, "")) < 1 {
		return nil, nil
	}

	decorators := []autorest.PrepareDecorator{
		autorest.WithBaseURL(pointer.StringDeref(vmaslr.NextLink, "")),
	}
	return c.armClient.PrepareGetRequest(ctx, decorators...)
}

// listNextResults retrieves the next set of results, if any.
func (c *Client) listNextResults(ctx context.Context, lastResults compute.AvailabilitySetListResult) (result compute.AvailabilitySetListResult, err error) {
	req, err := c.availabilitySetListResultPreparer(ctx, lastResults)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "vmasclient", "listNextResults", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}

	resp, rerr := c.armClient.Send(ctx, req)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(rerr.Error(), "vmasclient", "listNextResults", resp, "Failure sending next results request")
	}

	result, err = c.listResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "vmasclient", "listNextResults", resp, "Failure responding to next results request")
	}

	return
}

// AvailabilitySetListResultPage  contains a page of AvailabilitySet values.
type AvailabilitySetListResultPage struct {
	fn     func(context.Context, compute.AvailabilitySetListResult) (compute.AvailabilitySetListResult, error)
	vmaslr compute.AvailabilitySetListResult
}

// NextWithContext advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
func (page *AvailabilitySetListResultPage) NextWithContext(ctx context.Context) (err error) {
	next, err := page.fn(ctx, page.vmaslr)
	if err != nil {
		return err
	}
	page.vmaslr = next
	return nil
}

// Next advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
// Deprecated: Use NextWithContext() instead.
func (page *AvailabilitySetListResultPage) Next() error {
	return page.NextWithContext(context.Background())
}

// NotDone returns true if the page enumeration should be started or is not yet complete.
func (page AvailabilitySetListResultPage) NotDone() bool {
	return !page.vmaslr.IsEmpty()
}

// Response returns the raw server response from the last page request.
func (page AvailabilitySetListResultPage) Response() compute.AvailabilitySetListResult {
	return page.vmaslr
}

// Values returns the slice of values for the current page or nil if there are no values.
func (page AvailabilitySetListResultPage) Values() []compute.AvailabilitySet {
	if page.vmaslr.IsEmpty() {
		return nil
	}
	return *page.vmaslr.Value
}
