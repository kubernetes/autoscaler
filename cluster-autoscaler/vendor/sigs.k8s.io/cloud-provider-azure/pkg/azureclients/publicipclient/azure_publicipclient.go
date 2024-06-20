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

package publicipclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
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

const publicIPResourceType = "Microsoft.Network/publicIPAddresses"

// Client implements PublicIPAddress client Interface.
type Client struct {
	armClient              armclient.Interface
	subscriptionID         string
	cloudName              string
	disableAzureStackCloud bool

	// Rate limiting configures.
	rateLimiterReader flowcontrol.RateLimiter
	rateLimiterWriter flowcontrol.RateLimiter

	// ARM throttling configures.
	RetryAfterReader time.Time
	RetryAfterWriter time.Time

	computeAPIVersion string
}

// New creates a new PublicIPAddress client with ratelimiting.
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
		klog.V(2).Infof("Azure PublicIPAddressesClient (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure PublicIPAddressesClient (write ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPSWrite,
			config.RateLimitConfig.CloudProviderRateLimitBucketWrite)
	}

	computeAPIVersion := ComputeAPIVersion
	if strings.EqualFold(config.CloudName, AzureStackCloudName) && !config.DisableAzureStackCloud {
		computeAPIVersion = AzureStackComputeAPIVersion
	}

	client := &Client{
		armClient:              armClient,
		rateLimiterReader:      rateLimiterReader,
		rateLimiterWriter:      rateLimiterWriter,
		subscriptionID:         config.SubscriptionID,
		cloudName:              config.CloudName,
		disableAzureStackCloud: config.DisableAzureStackCloud,
		computeAPIVersion:      computeAPIVersion,
	}

	return client
}

// Get gets a PublicIPAddress.
func (c *Client) Get(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (network.PublicIPAddress, *retry.Error) {
	mc := metrics.NewMetricContext("public_ip_addresses", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return network.PublicIPAddress{}, retry.GetRateLimitError(false, "PublicIPGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PublicIPGet", "client throttled", c.RetryAfterReader)
		return network.PublicIPAddress{}, rerr
	}

	result, rerr := c.getPublicIPAddress(ctx, resourceGroupName, publicIPAddressName, expand)
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

// getPublicIPAddress gets a PublicIPAddress.
func (c *Client) getPublicIPAddress(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (network.PublicIPAddress, *retry.Error) {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		publicIPResourceType,
		publicIPAddressName,
	)
	result := network.PublicIPAddress{}

	response, rerr := c.armClient.GetResourceWithExpandQuery(ctx, resourceID, expand)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}

// GetVirtualMachineScaleSetPublicIPAddress gets a PublicIPAddress for VMSS VM.
func (c *Client) GetVirtualMachineScaleSetPublicIPAddress(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string, IPConfigurationName string, publicIPAddressName string, expand string) (network.PublicIPAddress, *retry.Error) {
	mc := metrics.NewMetricContext("vmss_public_ip_addresses", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return network.PublicIPAddress{}, retry.GetRateLimitError(false, "VMSSPublicIPGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VMSSPublicIPGet", "client throttled", c.RetryAfterReader)
		return network.PublicIPAddress{}, rerr
	}

	result, rerr := c.getVMSSPublicIPAddress(ctx, resourceGroupName, virtualMachineScaleSetName, virtualmachineIndex, networkInterfaceName, IPConfigurationName, publicIPAddressName, expand)
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

// getVMSSPublicIPAddress gets a PublicIPAddress for VMSS VM.
func (c *Client) getVMSSPublicIPAddress(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string, IPConfigurationName string, publicIPAddressName string, expand string) (network.PublicIPAddress, *retry.Error) {
	resourceID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachineScaleSets/%s/virtualMachines/%s/networkInterfaces/%s/ipconfigurations/%s/publicipaddresses/%s",
		autorest.Encode("path", c.subscriptionID),
		autorest.Encode("path", resourceGroupName),
		autorest.Encode("path", virtualMachineScaleSetName),
		autorest.Encode("path", virtualmachineIndex),
		autorest.Encode("path", networkInterfaceName),
		autorest.Encode("path", IPConfigurationName),
		autorest.Encode("path", publicIPAddressName),
	)

	result := network.PublicIPAddress{}
	response, rerr := c.armClient.GetResourceWithExpandAPIVersionQuery(ctx, resourceID, expand, c.computeAPIVersion)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmsspublicip.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmsspublicip.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}

// List gets a list of PublicIPAddress in the resource group.
func (c *Client) List(ctx context.Context, resourceGroupName string) ([]network.PublicIPAddress, *retry.Error) {
	mc := metrics.NewMetricContext("public_ip_addresses", "list", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return nil, retry.GetRateLimitError(false, "PublicIPList")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PublicIPList", "client throttled", c.RetryAfterReader)
		return nil, rerr
	}

	result, rerr := c.listPublicIPAddress(ctx, resourceGroupName)
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

// listPublicIPAddress gets a list of PublicIPAddress in the resource group.
func (c *Client) listPublicIPAddress(ctx context.Context, resourceGroupName string) ([]network.PublicIPAddress, *retry.Error) {
	resourceID := armclient.GetResourceListID(c.subscriptionID, resourceGroupName, publicIPResourceType)
	result := make([]network.PublicIPAddress, 0)
	page := &PublicIPAddressListResultPage{}
	page.fn = c.listNextResults

	resp, rerr := c.armClient.GetResource(ctx, resourceID)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.list.request", resourceID, rerr.Error())
		return result, rerr
	}

	var err error
	page.pialr, err = c.listResponder(resp)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.list.respond", resourceID, err)
		return result, retry.GetError(resp, err)
	}

	for {
		result = append(result, page.Values()...)

		// Abort the loop when there's no nextLink in the response.
		if pointer.StringDeref(page.Response().NextLink, "") == "" {
			break
		}

		if err = page.NextWithContext(ctx); err != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.list.next", resourceID, err)
			return result, retry.GetError(page.Response().Response.Response, err)
		}
	}

	return result, nil
}

// CreateOrUpdate creates or updates a PublicIPAddress.
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, publicIPAddressName string, parameters network.PublicIPAddress) *retry.Error {
	mc := metrics.NewMetricContext("public_ip_addresses", "create_or_update", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "PublicIPCreateOrUpdate")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PublicIPCreateOrUpdate", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.createOrUpdatePublicIP(ctx, resourceGroupName, publicIPAddressName, parameters)
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

// createOrUpdatePublicIP creates or updates a PublicIPAddress.
func (c *Client) createOrUpdatePublicIP(ctx context.Context, resourceGroupName string, publicIPAddressName string, parameters network.PublicIPAddress) *retry.Error {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		publicIPResourceType,
		publicIPAddressName,
	)

	response, rerr := c.armClient.PutResource(ctx, resourceID, parameters)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.put.request", resourceID, rerr.Error())
		return rerr
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		_, rerr = c.createOrUpdateResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.put.respond", resourceID, rerr.Error())
			return rerr
		}
	}

	return nil
}

func (c *Client) createOrUpdateResponder(resp *http.Response) (*network.PublicIPAddress, *retry.Error) {
	result := &network.PublicIPAddress{}
	err := autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return result, retry.GetError(resp, err)
}

// Delete deletes a PublicIPAddress by name.
func (c *Client) Delete(ctx context.Context, resourceGroupName string, publicIPAddressName string) *retry.Error {
	mc := metrics.NewMetricContext("public_ip_addresses", "delete", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "PublicIPDelete")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("PublicIPDelete", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.deletePublicIP(ctx, resourceGroupName, publicIPAddressName)
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

// deletePublicIP deletes a PublicIPAddress by name.
func (c *Client) deletePublicIP(ctx context.Context, resourceGroupName string, publicIPAddressName string) *retry.Error {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		publicIPResourceType,
		publicIPAddressName,
	)

	return c.armClient.DeleteResource(ctx, resourceID)
}

func (c *Client) listResponder(resp *http.Response) (result network.PublicIPAddressListResult, err error) {
	err = autorest.Respond(
		resp,
		autorest.ByIgnoring(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return
}

// publicIPAddressListResultPreparer prepares a request to retrieve the next set of results.
// It returns nil if no more results exist.
func (c *Client) publicIPAddressListResultPreparer(ctx context.Context, lr network.PublicIPAddressListResult) (*http.Request, error) {
	if lr.NextLink == nil || len(pointer.StringDeref(lr.NextLink, "")) < 1 {
		return nil, nil
	}

	decorators := []autorest.PrepareDecorator{
		autorest.WithBaseURL(pointer.StringDeref(lr.NextLink, "")),
	}
	return c.armClient.PrepareGetRequest(ctx, decorators...)
}

// listNextResults retrieves the next set of results, if any.
func (c *Client) listNextResults(ctx context.Context, lastResults network.PublicIPAddressListResult) (result network.PublicIPAddressListResult, err error) {
	req, err := c.publicIPAddressListResultPreparer(ctx, lastResults)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "publicipclient", "listNextResults", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}

	resp, rerr := c.armClient.Send(ctx, req)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(rerr.Error(), "publicipclient", "listNextResults", resp, "Failure sending next results request")
	}

	result, err = c.listResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "publicipclient", "listNextResults", resp, "Failure responding to next results request")
	}

	return
}

// PublicIPAddressListResultPage contains a page of PublicIPAddress values.
type PublicIPAddressListResultPage struct {
	fn    func(context.Context, network.PublicIPAddressListResult) (network.PublicIPAddressListResult, error)
	pialr network.PublicIPAddressListResult
}

// NextWithContext advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
func (page *PublicIPAddressListResultPage) NextWithContext(ctx context.Context) (err error) {
	next, err := page.fn(ctx, page.pialr)
	if err != nil {
		return err
	}
	page.pialr = next
	return nil
}

// Next advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
// Deprecated: Use NextWithContext() instead.
func (page *PublicIPAddressListResultPage) Next() error {
	return page.NextWithContext(context.Background())
}

// NotDone returns true if the page enumeration should be started or is not yet complete.
func (page PublicIPAddressListResultPage) NotDone() bool {
	return !page.pialr.IsEmpty()
}

// Response returns the raw server response from the last page request.
func (page PublicIPAddressListResultPage) Response() network.PublicIPAddressListResult {
	return page.pialr
}

// Values returns the slice of values for the current page or nil if there are no values.
func (page PublicIPAddressListResultPage) Values() []network.PublicIPAddress {
	if page.pialr.IsEmpty() {
		return nil
	}
	return *page.pialr.Value
}

// ListAll gets all of PublicIPAddress in the subscription.
func (c *Client) ListAll(ctx context.Context) ([]network.PublicIPAddress, *retry.Error) {
	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		return nil, retry.GetRateLimitError(false, "PublicIPListAll")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		rerr := retry.GetThrottlingError("PublicIPListAll", "client throttled", c.RetryAfterReader)
		return nil, rerr
	}

	result, rerr := c.listAllPublicIPAddress(ctx)
	if rerr != nil {
		if rerr.IsThrottled() {
			// Update RetryAfterReader so that no more requests would be sent until RetryAfter expires.
			c.RetryAfterReader = rerr.RetryAfter
		}

		return result, rerr
	}

	return result, nil
}

// listAllPublicIPAddress gets all of PublicIPAddress in the subscription.
func (c *Client) listAllPublicIPAddress(ctx context.Context) ([]network.PublicIPAddress, *retry.Error) {
	resourceID := armclient.GetProviderResourceID(c.subscriptionID, publicIPResourceType)
	result := make([]network.PublicIPAddress, 0)
	page := &PublicIPAddressListResultPage{}
	page.fn = c.listNextResults

	resp, rerr := c.armClient.GetResource(ctx, resourceID)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.listall.request", resourceID, rerr.Error())
		return result, rerr
	}

	var err error
	page.pialr, err = c.listResponder(resp)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.listall.respond", resourceID, err)
		return result, retry.GetError(resp, err)
	}

	for {
		result = append(result, page.Values()...)

		// Abort the loop when there's no nextLink in the response.
		if pointer.StringDeref(page.Response().NextLink, "") == "" {
			break
		}

		if err = page.NextWithContext(ctx); err != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "publicip.listall.next", resourceID, err)
			return result, retry.GetError(page.Response().Response.Response, err)
		}
	}

	return result, nil
}
