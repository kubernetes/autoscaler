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

package loadbalancerclient

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
	"k8s.io/utils/pointer"

	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/armclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

var _ Interface = &Client{}

const lbResourceType = "Microsoft.Network/loadBalancers"

// Client implements LoadBalancer client Interface.
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

type backendPoolsToBeMigrated struct {
	BackendPoolNames []string `json:"pools"`
}

// New creates a new LoadBalancer client with ratelimiting.
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
		klog.V(2).Infof("Azure LoadBalancersClient (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure LoadBalancersClient (write ops) using rate limit config: QPS=%g, bucket=%d",
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

// Get gets a LoadBalancer.
func (c *Client) Get(ctx context.Context, resourceGroupName string, loadBalancerName string, expand string) (network.LoadBalancer, *retry.Error) {
	mc := metrics.NewMetricContext("load_balancers", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return network.LoadBalancer{}, retry.GetRateLimitError(false, "LBGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("LBGet", "client throttled", c.RetryAfterReader)
		return network.LoadBalancer{}, rerr
	}

	result, rerr := c.getLB(ctx, resourceGroupName, loadBalancerName, expand)
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

// getLB gets a LoadBalancer.
func (c *Client) getLB(ctx context.Context, resourceGroupName string, loadBalancerName string, expand string) (network.LoadBalancer, *retry.Error) {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		lbResourceType,
		loadBalancerName,
	)
	result := network.LoadBalancer{}

	response, rerr := c.armClient.GetResourceWithExpandQuery(ctx, resourceID, expand)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}

// List gets a list of LoadBalancer in the resource group.
func (c *Client) List(ctx context.Context, resourceGroupName string) ([]network.LoadBalancer, *retry.Error) {
	mc := metrics.NewMetricContext("load_balancers", "list", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return nil, retry.GetRateLimitError(false, "LBList")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("LBList", "client throttled", c.RetryAfterReader)
		return nil, rerr
	}

	result, rerr := c.listLB(ctx, resourceGroupName)
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

// listLB gets a list of LoadBalancers in the resource group.
func (c *Client) listLB(ctx context.Context, resourceGroupName string) ([]network.LoadBalancer, *retry.Error) {
	resourceID := armclient.GetResourceListID(c.subscriptionID, resourceGroupName, "Microsoft.Network/loadBalancers")
	result := make([]network.LoadBalancer, 0)
	page := &LoadBalancerListResultPage{}
	page.fn = c.listNextResults

	resp, rerr := c.armClient.GetResource(ctx, resourceID)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.list.request", resourceID, rerr.Error())
		return result, rerr
	}

	var err error
	page.lblr, err = c.listResponder(resp)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.list.respond", resourceID, err)
		return result, retry.GetError(resp, err)
	}

	for {
		result = append(result, page.Values()...)

		// Abort the loop when there's no nextLink in the response.
		if pointer.StringDeref(page.Response().NextLink, "") == "" {
			break
		}

		if err = page.NextWithContext(ctx); err != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.list.next", resourceID, err)
			return result, retry.GetError(page.Response().Response.Response, err)
		}
	}

	return result, nil
}

// CreateOrUpdate creates or updates a LoadBalancer.
func (c *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, loadBalancerName string, parameters network.LoadBalancer, etag string) *retry.Error {
	mc := metrics.NewMetricContext("load_balancers", "create_or_update", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "LBCreateOrUpdate")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("LBCreateOrUpdate", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.createOrUpdateLB(ctx, resourceGroupName, loadBalancerName, parameters, etag)
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

// createOrUpdateLB creates or updates a LoadBalancer.
func (c *Client) createOrUpdateLB(ctx context.Context, resourceGroupName string, loadBalancerName string, parameters network.LoadBalancer, etag string) *retry.Error {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		lbResourceType,
		loadBalancerName,
	)
	decorators := []autorest.PrepareDecorator{}
	if etag != "" {
		decorators = append(decorators, autorest.WithHeader("If-Match", autorest.String(etag)))
	}

	response, rerr := c.armClient.PutResource(ctx, resourceID, parameters, decorators...)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.put.request", resourceID, rerr.Error())
		return rerr
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		_, rerr = c.createOrUpdateResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.put.respond", resourceID, rerr.Error())
			return rerr
		}
	}

	return nil
}

func (c *Client) createOrUpdateResponder(resp *http.Response) (*network.LoadBalancer, *retry.Error) {
	result := &network.LoadBalancer{}
	err := autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return result, retry.GetError(resp, err)
}

// Delete deletes a LoadBalancer by name.
func (c *Client) Delete(ctx context.Context, resourceGroupName string, loadBalancerName string) *retry.Error {
	mc := metrics.NewMetricContext("load_balancers", "delete", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "LBDelete")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("LBDelete", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.deleteLB(ctx, resourceGroupName, loadBalancerName)
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

// deleteLB deletes a LoadBalancer by name.
func (c *Client) deleteLB(ctx context.Context, resourceGroupName string, loadBalancerName string) *retry.Error {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		lbResourceType,
		loadBalancerName,
	)

	return c.armClient.DeleteResource(ctx, resourceID)
}

func (c *Client) listResponder(resp *http.Response) (result network.LoadBalancerListResult, err error) {
	err = autorest.Respond(
		resp,
		autorest.ByIgnoring(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return
}

// loadBalancerListResultPreparer prepares a request to retrieve the next set of results.
// It returns nil if no more results exist.
func (c *Client) loadBalancerListResultPreparer(ctx context.Context, lblr network.LoadBalancerListResult) (*http.Request, error) {
	if lblr.NextLink == nil || len(pointer.StringDeref(lblr.NextLink, "")) < 1 {
		return nil, nil
	}

	decorators := []autorest.PrepareDecorator{
		autorest.WithBaseURL(pointer.StringDeref(lblr.NextLink, "")),
	}
	return c.armClient.PrepareGetRequest(ctx, decorators...)
}

// listNextResults retrieves the next set of results, if any.
func (c *Client) listNextResults(ctx context.Context, lastResults network.LoadBalancerListResult) (result network.LoadBalancerListResult, err error) {
	req, err := c.loadBalancerListResultPreparer(ctx, lastResults)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "loadbalancerclient", "listNextResults", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}

	resp, rerr := c.armClient.Send(ctx, req)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(rerr.Error(), "loadbalancerclient", "listNextResults", resp, "Failure sending next results request")
	}

	result, err = c.listResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "loadbalancerclient", "listNextResults", resp, "Failure responding to next results request")
	}

	return
}

// LoadBalancerListResultPage contains a page of LoadBalancer values.
type LoadBalancerListResultPage struct {
	fn   func(context.Context, network.LoadBalancerListResult) (network.LoadBalancerListResult, error)
	lblr network.LoadBalancerListResult
}

// NextWithContext advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
func (page *LoadBalancerListResultPage) NextWithContext(ctx context.Context) (err error) {
	next, err := page.fn(ctx, page.lblr)
	if err != nil {
		return err
	}
	page.lblr = next
	return nil
}

// Next advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
// Deprecated: Use NextWithContext() instead.
func (page *LoadBalancerListResultPage) Next() error {
	return page.NextWithContext(context.Background())
}

// NotDone returns true if the page enumeration should be started or is not yet complete.
func (page LoadBalancerListResultPage) NotDone() bool {
	return !page.lblr.IsEmpty()
}

// Response returns the raw server response from the last page request.
func (page LoadBalancerListResultPage) Response() network.LoadBalancerListResult {
	return page.lblr
}

// Values returns the slice of values for the current page or nil if there are no values.
func (page LoadBalancerListResultPage) Values() []network.LoadBalancer {
	if page.lblr.IsEmpty() {
		return nil
	}
	return *page.lblr.Value
}

// GetLBBackendPool gets a LoadBalancer backend pool.
func (c *Client) GetLBBackendPool(ctx context.Context, resourceGroupName string, loadBalancerName string, backendPoolName string, expand string) (network.BackendAddressPool, *retry.Error) {
	mc := metrics.NewMetricContext("load_balancers", "get_backend_pool", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return network.BackendAddressPool{}, retry.GetRateLimitError(false, "LBGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("LBBackendPoolGet", "client throttled", c.RetryAfterReader)
		return network.BackendAddressPool{}, rerr
	}

	result, rerr := c.getLBBackendPool(ctx, resourceGroupName, loadBalancerName, backendPoolName, expand)
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

// getLBBackendPool gets a LoadBalancer backend pool.
func (c *Client) getLBBackendPool(ctx context.Context, resourceGroupName string, loadBalancerName string, backendPoolName string, expand string) (network.BackendAddressPool, *retry.Error) {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		lbResourceType,
		loadBalancerName,
		"backendAddressPools",
		backendPoolName,
	)
	result := network.BackendAddressPool{}

	response, rerr := c.armClient.GetResourceWithExpandQuery(ctx, resourceID, expand)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.backendpool.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.backendpool.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}

// CreateOrUpdateBackendPools creates or updates a LoadBalancer backend pool.
func (c *Client) CreateOrUpdateBackendPools(ctx context.Context, resourceGroupName string, loadBalancerName string, backendPoolName string, parameters network.BackendAddressPool, etag string) *retry.Error {
	mc := metrics.NewMetricContext("load_balancers", "create_or_update_backend_pools", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "LBCreateOrUpdateBackendPools")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("LBCreateOrUpdateBackendPools", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.createOrUpdateLBBackendPool(ctx, resourceGroupName, loadBalancerName, backendPoolName, parameters, etag)
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

// createOrUpdateLBBackendPool creates or updates a LoadBalancer.
func (c *Client) createOrUpdateLBBackendPool(ctx context.Context, resourceGroupName string, loadBalancerName string, backendPoolName string, parameters network.BackendAddressPool, etag string) *retry.Error {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		lbResourceType,
		loadBalancerName,
		"backendAddressPools",
		backendPoolName,
	)
	decorators := []autorest.PrepareDecorator{}
	if etag != "" {
		decorators = append(decorators, autorest.WithHeader("If-Match", autorest.String(etag)))
	}

	response, rerr := c.armClient.PutResource(ctx, resourceID, parameters, decorators...)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.backendpool.put.request", resourceID, rerr.Error())
		return rerr
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		_, rerr = c.createOrUpdateBackendPoolResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancer.backendpool.put.respond", resourceID, rerr.Error())
			return rerr
		}
	}

	return nil
}

// DeleteLBBackendPool deletes a LoadBalancer backend pool by name.
func (c *Client) DeleteLBBackendPool(ctx context.Context, resourceGroupName, loadBalancerName, backendPoolName string) *retry.Error {
	mc := metrics.NewMetricContext("load_balancers", "delete_backend_pool", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "LBDeleteBackendPool")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("LBDeleteBackendPool", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.deleteLBBackendPool(ctx, resourceGroupName, loadBalancerName, backendPoolName)
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

func (c *Client) deleteLBBackendPool(ctx context.Context, resourceGroupName, loadBalancerName, backendPoolName string) *retry.Error {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		lbResourceType,
		loadBalancerName,
		"backendAddressPools",
		backendPoolName,
	)
	return c.armClient.DeleteResource(ctx, resourceID)
}

func (c *Client) createOrUpdateBackendPoolResponder(resp *http.Response) (*network.BackendAddressPool, *retry.Error) {
	result := &network.BackendAddressPool{}
	err := autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return result, retry.GetError(resp, err)
}

// MigrateToIPBasedBackendPool migrates a NIC-based backend pool to IP-based.
func (c *Client) MigrateToIPBasedBackendPool(ctx context.Context, resourceGroupName string, loadBalancerName string, backendPoolNames []string) *retry.Error {
	mc := metrics.NewMetricContext("load_balancers", "migrate_to_ip_based_backend_pool", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "LBMigrateToIPBasedBackendPool")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("LBMigrateToIPBasedBackendPool", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	parameters := backendPoolsToBeMigrated{
		BackendPoolNames: backendPoolNames,
	}
	rerr := c.migrateToIPBasedBackendPool(ctx, resourceGroupName, loadBalancerName, parameters)
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

func (c *Client) migrateToIPBasedBackendPool(ctx context.Context, resourceGroupName string, loadBalancerName string, parameters backendPoolsToBeMigrated) *retry.Error {
	resourceID := armclient.GetResourceID(
		c.subscriptionID,
		resourceGroupName,
		lbResourceType,
		loadBalancerName,
	)

	response, rerr := c.armClient.PostResource(ctx, resourceID, "migrateToIpBased", parameters, map[string]interface{}{})
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "loadbalancerbackendpool.migrate.request", resourceID, rerr.Error())
		return rerr
	}

	klog.Infof("Response: %v", response)
	return nil
}
