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

package vmssvmclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azclients "sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/armclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

var _ Interface = &Client{}

const (
	vmssResourceType = "Microsoft.Compute/virtualMachineScaleSets"
	vmResourceType   = "virtualMachines"
)

// Client implements VMSS client Interface.
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

// New creates a new vmssVM client with ratelimiting.
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
		klog.V(2).Infof("Azure vmssVM client (read ops) using rate limit config: QPS=%g, bucket=%d",
			config.RateLimitConfig.CloudProviderRateLimitQPS,
			config.RateLimitConfig.CloudProviderRateLimitBucket)
		klog.V(2).Infof("Azure vmssVM client (write ops) using rate limit config: QPS=%g, bucket=%d",
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

// Get gets a VirtualMachineScaleSetVM.
func (c *Client) Get(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, expand compute.InstanceViewTypes) (compute.VirtualMachineScaleSetVM, *retry.Error) {
	mc := metrics.NewMetricContext("vmssvm", "get", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return compute.VirtualMachineScaleSetVM{}, retry.GetRateLimitError(false, "VMSSVMGet")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VMSSVMGet", "client throttled", c.RetryAfterReader)
		return compute.VirtualMachineScaleSetVM{}, rerr
	}

	result, rerr := c.getVMSSVM(ctx, resourceGroupName, VMScaleSetName, instanceID, expand)
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

// getVMSSVM gets a VirtualMachineScaleSetVM.
func (c *Client) getVMSSVM(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, expand compute.InstanceViewTypes) (compute.VirtualMachineScaleSetVM, *retry.Error) {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		vmssResourceType,
		VMScaleSetName,
		vmResourceType,
		instanceID,
	)
	result := compute.VirtualMachineScaleSetVM{}

	response, rerr := c.armClient.GetResourceWithExpandQuery(ctx, resourceID, string(expand))
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.get.request", resourceID, rerr.Error())
		return result, rerr
	}

	err := autorest.Respond(
		response,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.get.respond", resourceID, err)
		return result, retry.GetError(response, err)
	}

	result.Response = autorest.Response{Response: response}
	return result, nil
}

// List gets a list of VirtualMachineScaleSetVMs in the virtualMachineScaleSet.
func (c *Client) List(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, expand string) ([]compute.VirtualMachineScaleSetVM, *retry.Error) {
	mc := metrics.NewMetricContext("vmssvm", "list", resourceGroupName, c.subscriptionID, "")

	// Report errors if the client is rate limited.
	if !c.rateLimiterReader.TryAccept() {
		mc.RateLimitedCount()
		return nil, retry.GetRateLimitError(false, "VMSSVMList")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterReader.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VMSSVMList", "client throttled", c.RetryAfterReader)
		return nil, rerr
	}

	result, rerr := c.listVMSSVM(ctx, resourceGroupName, virtualMachineScaleSetName, expand)
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

// listVMSSVM gets a list of VirtualMachineScaleSetVMs in the virtualMachineScaleSet.
func (c *Client) listVMSSVM(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, expand string) ([]compute.VirtualMachineScaleSetVM, *retry.Error) {
	resourceID := armclient.GetChildResourcesListID(
		c.subscriptionID,
		resourceGroupName,
		vmssResourceType,
		virtualMachineScaleSetName,
		vmResourceType,
	)

	result := make([]compute.VirtualMachineScaleSetVM, 0)
	page := &VirtualMachineScaleSetVMListResultPage{}
	page.fn = c.listNextResults

	resp, rerr := c.armClient.GetResourceWithExpandQuery(ctx, resourceID, expand)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.list.request", resourceID, rerr.Error())
		return result, rerr
	}

	var err error
	page.vmssvlr, err = c.listResponder(resp)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.list.respond", resourceID, err)
		return result, retry.GetError(resp, err)
	}

	for {
		result = append(result, page.Values()...)

		// Abort the loop when there's no nextLink in the response.
		if pointer.StringDeref(page.Response().NextLink, "") == "" {
			break
		}

		if err = page.NextWithContext(ctx); err != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.list.next", resourceID, err)
			return result, retry.GetError(page.Response().Response.Response, err)
		}
	}

	return result, nil
}

// Update updates a VirtualMachineScaleSetVM.
func (c *Client) Update(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters compute.VirtualMachineScaleSetVM, source string) (*compute.VirtualMachineScaleSetVM, *retry.Error) {
	mc := metrics.NewMetricContext("vmssvm", "update", resourceGroupName, c.subscriptionID, source)

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return nil, retry.GetRateLimitError(true, "VMSSVMUpdate")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VMSSVMUpdate", "client throttled", c.RetryAfterWriter)
		return nil, rerr
	}

	result, rerr := c.updateVMSSVM(ctx, resourceGroupName, VMScaleSetName, instanceID, parameters)
	mc.Observe(rerr)
	if rerr != nil {
		if rerr.IsThrottled() {
			// Update RetryAfterReader so that no more requests would be sent until RetryAfter expires.
			c.RetryAfterWriter = rerr.RetryAfter
		}
	}

	return result, rerr
}

// UpdateAsync updates a VirtualMachineScaleSetVM asynchronously
func (c *Client) UpdateAsync(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters compute.VirtualMachineScaleSetVM, source string) (*azure.Future, *retry.Error) {
	mc := metrics.NewMetricContext("vmssvm", "updateasync", resourceGroupName, c.subscriptionID, source)

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return nil, retry.GetRateLimitError(true, "VMSSVMUpdateAsync")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VMSSVMUpdateAsync", "client throttled", c.RetryAfterWriter)
		return nil, rerr
	}

	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		vmssResourceType,
		VMScaleSetName,
		vmResourceType,
		instanceID,
	)

	future, rerr := c.armClient.PutResourceAsync(ctx, resourceID, parameters)
	mc.Observe(rerr)
	if rerr != nil {
		if rerr.IsThrottled() {
			// Update RetryAfterReader so that no more requests would be sent until RetryAfter expires.
			c.RetryAfterWriter = rerr.RetryAfter
		}

		return nil, rerr
	}

	return future, nil
}

// WaitForUpdateResult waits for the response of the update request
func (c *Client) WaitForUpdateResult(ctx context.Context, future *azure.Future, resourceGroupName, source string) (*compute.VirtualMachineScaleSetVM, *retry.Error) {
	mc := metrics.NewMetricContext("vmss", "wait_for_update_result", resourceGroupName, c.subscriptionID, source)
	response, err := c.armClient.WaitForAsyncOperationResult(ctx, future, "VMSSWaitForUpdateResult")
	mc.Observe(retry.NewErrorOrNil(false, err))
	defer c.armClient.CloseResponse(ctx, response)

	if err != nil {
		if response != nil {
			klog.V(5).Infof("Received error in WaitForAsyncOperationResult: '%s', response code %d", err.Error(), response.StatusCode)
		} else {
			klog.V(5).Infof("Received error in WaitForAsyncOperationResult: '%s', no response", err.Error())
		}
		return nil, retry.GetError(response, err)
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		result, rerr := c.updateResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in WaitForAsyncOperationResult updateResponder: '%s'", rerr.Error())
		}

		return result, rerr
	}

	result := &compute.VirtualMachineScaleSetVM{}
	result.Response = autorest.Response{Response: response}
	return result, nil
}

// updateVMSSVM updates a VirtualMachineScaleSetVM.
func (c *Client) updateVMSSVM(ctx context.Context, resourceGroupName string, VMScaleSetName string, instanceID string, parameters compute.VirtualMachineScaleSetVM) (*compute.VirtualMachineScaleSetVM, *retry.Error) {
	resourceID := armclient.GetChildResourceID(
		c.subscriptionID,
		resourceGroupName,
		vmssResourceType,
		VMScaleSetName,
		vmResourceType,
		instanceID,
	)

	response, rerr := c.armClient.PutResource(ctx, resourceID, parameters)
	defer c.armClient.CloseResponse(ctx, response)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.put.request", resourceID, rerr.Error())
		return nil, rerr
	}

	if response != nil && response.StatusCode != http.StatusNoContent {
		result, rerr := c.updateResponder(response)
		if rerr != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.put.respond", resourceID, rerr.Error())
		}
		return result, rerr
	}

	result := &compute.VirtualMachineScaleSetVM{}
	result.Response = autorest.Response{Response: response}
	return result, nil
}

func (c *Client) updateResponder(resp *http.Response) (*compute.VirtualMachineScaleSetVM, *retry.Error) {
	result := &compute.VirtualMachineScaleSetVM{}
	err := autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return result, retry.GetError(resp, err)
}

func (c *Client) listResponder(resp *http.Response) (result compute.VirtualMachineScaleSetVMListResult, err error) {
	err = autorest.Respond(
		resp,
		autorest.ByIgnoring(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result))
	result.Response = autorest.Response{Response: resp}
	return
}

// virtualMachineScaleSetListResultPreparer prepares a request to retrieve the next set of results.
// It returns nil if no more results exist.
func (c *Client) virtualMachineScaleSetVMListResultPreparer(ctx context.Context, vmssvmlr compute.VirtualMachineScaleSetVMListResult) (*http.Request, error) {
	if vmssvmlr.NextLink == nil || len(pointer.StringDeref(vmssvmlr.NextLink, "")) < 1 {
		return nil, nil
	}

	decorators := []autorest.PrepareDecorator{
		autorest.WithBaseURL(pointer.StringDeref(vmssvmlr.NextLink, "")),
	}
	return c.armClient.PrepareGetRequest(ctx, decorators...)
}

// listNextResults retrieves the next set of results, if any.
func (c *Client) listNextResults(ctx context.Context, lastResults compute.VirtualMachineScaleSetVMListResult) (result compute.VirtualMachineScaleSetVMListResult, err error) {
	req, err := c.virtualMachineScaleSetVMListResultPreparer(ctx, lastResults)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "vmssvmclient", "listNextResults", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}

	resp, rerr := c.armClient.Send(ctx, req)
	defer c.armClient.CloseResponse(ctx, resp)
	if rerr != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(rerr.Error(), "vmssvmclient", "listNextResults", resp, "Failure sending next results request")
	}

	result, err = c.listResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "vmssvmclient", "listNextResults", resp, "Failure responding to next results request")
	}

	return
}

// VirtualMachineScaleSetVMListResultPage contains a page of VirtualMachineScaleSetVM values.
type VirtualMachineScaleSetVMListResultPage struct {
	fn      func(context.Context, compute.VirtualMachineScaleSetVMListResult) (compute.VirtualMachineScaleSetVMListResult, error)
	vmssvlr compute.VirtualMachineScaleSetVMListResult
}

// NextWithContext advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
func (page *VirtualMachineScaleSetVMListResultPage) NextWithContext(ctx context.Context) (err error) {
	next, err := page.fn(ctx, page.vmssvlr)
	if err != nil {
		return err
	}
	page.vmssvlr = next
	return nil
}

// Next advances to the next page of values.  If there was an error making
// the request the page does not advance and the error is returned.
// Deprecated: Use NextWithContext() instead.
func (page *VirtualMachineScaleSetVMListResultPage) Next() error {
	return page.NextWithContext(context.Background())
}

// NotDone returns true if the page enumeration should be started or is not yet complete.
func (page VirtualMachineScaleSetVMListResultPage) NotDone() bool {
	return !page.vmssvlr.IsEmpty()
}

// Response returns the raw server response from the last page request.
func (page VirtualMachineScaleSetVMListResultPage) Response() compute.VirtualMachineScaleSetVMListResult {
	return page.vmssvlr
}

// Values returns the slice of values for the current page or nil if there are no values.
func (page VirtualMachineScaleSetVMListResultPage) Values() []compute.VirtualMachineScaleSetVM {
	if page.vmssvlr.IsEmpty() {
		return nil
	}
	return *page.vmssvlr.Value
}

// UpdateVMs updates a list of VirtualMachineScaleSetVM from map[instanceID]compute.VirtualMachineScaleSetVM.
// If the batch size > 0, it will send sync requests concurrently in batches, or it will send sync requests in sequence.
// No matter what the batch size is, it will process the async requests concurrently in one single batch.
func (c *Client) UpdateVMs(ctx context.Context, resourceGroupName string, VMScaleSetName string, instances map[string]compute.VirtualMachineScaleSetVM, source string, batchSize int) *retry.Error {
	mc := metrics.NewMetricContext("vmssvm", "update_vms", resourceGroupName, c.subscriptionID, source)

	// Report errors if the client is rate limited.
	if !c.rateLimiterWriter.TryAccept() {
		mc.RateLimitedCount()
		return retry.GetRateLimitError(true, "VMSSVMUpdateVMs")
	}

	// Report errors if the client is throttled.
	if c.RetryAfterWriter.After(time.Now()) {
		mc.ThrottledCount()
		rerr := retry.GetThrottlingError("VMSSVMUpdateVMs", "client throttled", c.RetryAfterWriter)
		return rerr
	}

	rerr := c.updateVMSSVMs(ctx, resourceGroupName, VMScaleSetName, instances, batchSize)
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

// updateVMSSVMs updates a list of VirtualMachineScaleSetVM from map[instanceID]compute.VirtualMachineScaleSetVM.
func (c *Client) updateVMSSVMs(ctx context.Context, resourceGroupName string, VMScaleSetName string, instances map[string]compute.VirtualMachineScaleSetVM, batchSize int) *retry.Error {
	resources := make(map[string]interface{})
	for instanceID, parameter := range instances {
		resourceID := armclient.GetChildResourceID(
			c.subscriptionID,
			resourceGroupName,
			vmssResourceType,
			VMScaleSetName,
			vmResourceType,
			instanceID,
		)
		resources[resourceID] = parameter
	}

	responses := c.armClient.PutResourcesInBatches(ctx, resources, batchSize)
	errors, retryIDs := c.parseResp(ctx, responses, true)
	if len(retryIDs) > 0 {
		retryResources := make(map[string]interface{})
		for _, id := range retryIDs {
			retryResources[id] = resources[id]
		}
		resps := c.armClient.PutResourcesInBatches(ctx, retryResources, batchSize)
		errs, _ := c.parseResp(ctx, resps, false)
		errors = append(errors, errs...)
	}

	// Aggregate errors.
	if len(errors) > 0 {
		rerr := &retry.Error{}
		errs := make([]error, 0)
		for _, err := range errors {
			if !err.Retriable && strings.Contains(err.Error().Error(), consts.ConcurrentRequestConflictMessage) {
				err.Retriable = true
				err.RetryAfter = time.Now().Add(5 * time.Second)
			}

			if err.IsThrottled() && err.RetryAfter.After(rerr.RetryAfter) {
				rerr.RetryAfter = err.RetryAfter
			}
			errs = append(errs, err.Error())
		}
		rerr.RawError = utilerrors.Flatten(utilerrors.NewAggregate(errs))
		return rerr
	}

	return nil
}

func (c *Client) parseResp(
	ctx context.Context,
	responses map[string]*armclient.PutResourcesResponse,
	shouldRetry bool,
) ([]*retry.Error, []string) {
	var (
		errors   []*retry.Error
		retryIDs []string
	)
	for resourceID, resp := range responses {
		if resp == nil {
			continue
		}

		defer c.armClient.CloseResponse(ctx, resp.Response)
		if resp.Error != nil {
			klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.put.request", resourceID, resp.Error.Error())

			errMsg := resp.Error.Error().Error()
			if strings.Contains(errMsg, consts.VmssVMNotActiveErrorMessage) {
				klog.V(2).Infof("VMSS VM %s is not active, skip updating it.", resourceID)
				continue
			}
			if strings.Contains(errMsg, consts.ParentResourceNotFoundMessageCode) {
				klog.V(2).Infof("The parent resource of VMSS VM %s is not found, skip updating it.", resourceID)
				continue
			}
			if strings.Contains(errMsg, consts.CannotUpdateVMBeingDeletedMessagePrefix) &&
				strings.Contains(errMsg, consts.CannotUpdateVMBeingDeletedMessageSuffix) {
				klog.V(2).Infof("The VM %s is being deleted, skip updating it.", resourceID)
				continue
			}

			if retry.IsSuccessHTTPResponse(resp.Response) &&
				strings.Contains(
					strings.ToLower(errMsg),
					strings.ToLower(consts.OperationPreemptedErrorMessage),
				) {
				if shouldRetry {
					klog.V(2).Infof("The operation on VM %s is preempted, will retry.", resourceID)
					retryIDs = append(retryIDs, resourceID)
					continue
				}
				klog.V(2).Infof("The operation on VM %s is preempted, will not retry.", resourceID)
			}

			errors = append(errors, resp.Error)
			continue
		}

		if resp.Response != nil && resp.Response.StatusCode != http.StatusNoContent {
			_, rerr := c.updateResponder(resp.Response)
			if rerr != nil {
				klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "vmssvm.put.respond", resourceID, rerr.Error())
				errors = append(errors, rerr)
			}
		}
	}
	return errors, retryIDs
}
