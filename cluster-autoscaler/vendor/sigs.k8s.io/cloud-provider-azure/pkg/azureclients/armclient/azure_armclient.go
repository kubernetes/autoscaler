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

package armclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/tracing"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
	"sigs.k8s.io/cloud-provider-azure/pkg/version"
)

// there is one sender per TLS renegotiation type, i.e. count of tls.RenegotiationSupport enums

type defaultSender struct {
	sender autorest.Sender
	init   *sync.Once
}

// each type of sender will be created on demand in sender()
var defaultSenders defaultSender

func init() {
	defaultSenders.init = &sync.Once{}
}

var _ Interface = &Client{}

// Client implements ARM client Interface.
type Client struct {
	client           autorest.Client
	baseURI          string
	apiVersion       string
	regionalEndpoint string
}

func sender() autorest.Sender {
	// note that we can't init defaultSenders in init() since it will
	// execute before calling code has had a chance to enable tracing
	defaultSenders.init.Do(func() {
		// copied from http.DefaultTransport with a TLS minimum version.
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second, // the same as default transport
				KeepAlive: 30 * time.Second, // the same as default transport
			}).DialContext,
			ForceAttemptHTTP2:     false,            // respect custom dialer (default is true)
			MaxIdleConns:          100,              // Zero means no limit, the same as default transport
			MaxIdleConnsPerHost:   100,              // Default is 2, ref:https://cs.opensource.google/go/go/+/go1.18.4:src/net/http/transport.go;l=58
			IdleConnTimeout:       90 * time.Second, // the same as default transport
			TLSHandshakeTimeout:   10 * time.Second, // the same as default transport
			ExpectContinueTimeout: 1 * time.Second,  // the same as default transport
			TLSClientConfig: &tls.Config{
				MinVersion:    tls.VersionTLS12,     //force to use TLS 1.2
				Renegotiation: tls.RenegotiateNever, // the same as default transport https://pkg.go.dev/crypto/tls#RenegotiationSupport
			},
		}
		var roundTripper http.RoundTripper = transport
		if tracing.IsEnabled() {
			roundTripper = tracing.NewTransport(transport)
		}
		j, _ := cookiejar.New(nil)
		defaultSenders.sender = &http.Client{Jar: j, Transport: roundTripper}

		// In go-autorest SDK https://github.com/Azure/go-autorest/blob/master/autorest/sender.go#L258-L287,
		// if ARM returns http.StatusTooManyRequests, the sender doesn't increase the retry attempt count,
		// hence the Azure clients will keep retrying forever until it get a status code other than 429.
		// So we explicitly removes http.StatusTooManyRequests from autorest.StatusCodesForRetry.
		// Refer https://github.com/Azure/go-autorest/issues/398.
		// TODO(feiskyer): Use autorest.SendDecorator to customize the retry policy when new Azure SDK is available.
		statusCodesForRetry := make([]int, 0)
		for _, code := range autorest.StatusCodesForRetry {
			if code != http.StatusTooManyRequests {
				statusCodesForRetry = append(statusCodesForRetry, code)
			}
		}
		autorest.StatusCodesForRetry = statusCodesForRetry
	})
	return defaultSenders.sender
}

// New creates a ARM client
func New(authorizer autorest.Authorizer, clientConfig azureclients.ClientConfig, baseURI, apiVersion string, sendDecoraters ...autorest.SendDecorator) *Client {
	restClient := autorest.NewClientWithUserAgent(clientConfig.UserAgent)
	restClient.Authorizer = authorizer
	restClient.Sender = sender()

	if clientConfig.UserAgent == "" {
		restClient.UserAgent = GetUserAgent(restClient)
	}

	if clientConfig.RestClientConfig.PollingDelay == nil {
		restClient.PollingDelay = 5 * time.Second
	} else {
		restClient.PollingDelay = *clientConfig.RestClientConfig.PollingDelay
	}

	if clientConfig.RestClientConfig.RetryAttempts == nil {
		restClient.RetryAttempts = 3
	} else {
		restClient.RetryAttempts = *clientConfig.RestClientConfig.RetryAttempts
	}

	if clientConfig.RestClientConfig.RetryDuration == nil {
		restClient.RetryDuration = 1 * time.Second
	} else {
		restClient.RetryDuration = *clientConfig.RestClientConfig.RetryDuration
	}

	backoff := clientConfig.Backoff
	if backoff == nil {
		backoff = &retry.Backoff{}
	}
	if backoff.Steps == 0 {
		// 1 steps means no retry.
		backoff.Steps = 1
	}

	url, _ := url.Parse(baseURI)

	client := &Client{
		client:           restClient,
		baseURI:          baseURI,
		apiVersion:       apiVersion,
		regionalEndpoint: fmt.Sprintf("%s.%s", clientConfig.Location, url.Host),
	}
	client.client.Sender = autorest.DecorateSender(client.client,
		autorest.DoCloseIfError(),
		retry.DoExponentialBackoffRetry(backoff),
		DoDumpRequest(10),
	)

	client.client.Sender = autorest.DecorateSender(client.client.Sender, sendDecoraters...)

	return client
}

// GetUserAgent gets the autorest client with a user agent that
// includes "kubernetes" and the full kubernetes git version string
// example:
// Azure-SDK-for-Go/7.0.1 arm-network/2016-09-01; kubernetes-cloudprovider/v1.17.0;
func GetUserAgent(client autorest.Client) string {
	k8sVersion := version.Get().GitVersion
	return fmt.Sprintf("%s; kubernetes-cloudprovider/%s", client.UserAgent, k8sVersion)
}

// NormalizeAzureRegion returns a normalized Azure region with white spaces removed and converted to lower case
func NormalizeAzureRegion(name string) string {
	region := ""
	for _, runeValue := range name {
		if !unicode.IsSpace(runeValue) {
			region += string(runeValue)
		}
	}
	return strings.ToLower(region)
}

// Send sends a http request to ARM service with possible retry to regional ARM endpoint.
func (c *Client) Send(_ context.Context, request *http.Request, decorators ...autorest.SendDecorator) (*http.Response, *retry.Error) {
	response, err := autorest.SendWithSender(
		c.client,
		request,
		decorators...,
	)

	if response == nil && err == nil {
		return response, retry.NewError(false, fmt.Errorf("Empty response and no HTTP code"))
	}

	return response, retry.GetError(response, err)
}

// PreparePutRequest prepares put request
func (c *Client) PreparePutRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error) {
	decorators = append(
		[]autorest.PrepareDecorator{
			autorest.AsContentType("application/json; charset=utf-8"),
			autorest.AsPut(),
			autorest.WithBaseURL(c.baseURI)},
		decorators...)
	return c.prepareRequest(ctx, decorators...)
}

// PreparePatchRequest prepares patch request
func (c *Client) PreparePatchRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error) {
	decorators = append(
		[]autorest.PrepareDecorator{
			autorest.AsContentType("application/json; charset=utf-8"),
			autorest.AsPatch(),
			autorest.WithBaseURL(c.baseURI)},
		decorators...)
	return c.prepareRequest(ctx, decorators...)
}

// PreparePostRequest prepares post request
func (c *Client) PreparePostRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error) {
	decorators = append(
		[]autorest.PrepareDecorator{
			autorest.AsContentType("application/json; charset=utf-8"),
			autorest.AsPost(),
			autorest.WithBaseURL(c.baseURI)},
		decorators...)
	return c.prepareRequest(ctx, decorators...)
}

// PrepareGetRequest prepares get request
func (c *Client) PrepareGetRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error) {
	decorators = append(
		[]autorest.PrepareDecorator{
			autorest.AsGet(),
			autorest.WithBaseURL(c.baseURI)},
		decorators...)
	return c.prepareRequest(ctx, decorators...)
}

// PrepareDeleteRequest preparse delete request
func (c *Client) PrepareDeleteRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error) {
	decorators = append(
		[]autorest.PrepareDecorator{
			autorest.AsDelete(),
			autorest.WithBaseURL(c.baseURI)},
		decorators...)
	return c.prepareRequest(ctx, decorators...)
}

// PrepareHeadRequest prepares head request
func (c *Client) PrepareHeadRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error) {
	decorators = append(
		[]autorest.PrepareDecorator{
			autorest.AsHead(),
			autorest.WithBaseURL(c.baseURI)},
		decorators...)
	return c.prepareRequest(ctx, decorators...)
}

// WaitForAsyncOperationCompletion waits for an operation completion
func (c *Client) WaitForAsyncOperationCompletion(ctx context.Context, future *azure.Future, asyncOperationName string) error {
	err := future.WaitForCompletionRef(ctx, c.client)
	if err != nil {
		klog.V(5).Infof("Received error in WaitForCompletionRef: '%v'", err)
		return err
	}

	var done bool
	done, err = future.DoneWithContext(ctx, c.client)
	if err != nil {
		klog.V(5).Infof("Received error in DoneWithContext: '%v'", err)
		return autorest.NewErrorWithError(err, asyncOperationName, "Result", future.Response(), "Polling failure")
	}
	if !done {
		return azure.NewAsyncOpIncompleteError(asyncOperationName)
	}

	return nil
}

// WaitForAsyncOperationResult waits for an operation result.
func (c *Client) WaitForAsyncOperationResult(ctx context.Context, future *azure.Future, _ string) (*http.Response, error) {
	if err := future.WaitForCompletionRef(ctx, c.client); err != nil {
		klog.V(5).Infof("Received error in WaitForAsyncOperationCompletion: '%v'", err)
		return nil, err
	}
	return future.GetResult(c.client)
}

// SendAsync send a request and return a future object representing the async result as well as the origin http response
func (c *Client) SendAsync(ctx context.Context, request *http.Request) (*azure.Future, *http.Response, *retry.Error) {
	asyncResponse, rerr := c.Send(ctx, request)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "sendAsync.send", html.EscapeString(request.URL.String()), rerr.Error())
		return nil, nil, rerr
	}

	future, err := azure.NewFutureFromResponse(asyncResponse)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "sendAsync.respond", html.EscapeString(request.URL.String()), err)
		return nil, asyncResponse, retry.GetError(asyncResponse, err)
	}

	return &future, asyncResponse, nil
}

// GetResourceWithExpandQuery get a resource by resource ID with expand
func (c *Client) GetResourceWithExpandQuery(ctx context.Context, resourceID, expand string) (*http.Response, *retry.Error) {
	var decorators []autorest.PrepareDecorator
	if expand != "" {
		queryParameters := map[string]interface{}{
			"$expand": autorest.Encode("query", expand),
		}
		decorators = append(decorators, autorest.WithQueryParameters(queryParameters))
	}
	return c.GetResource(ctx, resourceID, decorators...)
}

// GetResourceWithExpandAPIVersionQuery get a resource by resource ID with expand and API version.
func (c *Client) GetResourceWithExpandAPIVersionQuery(ctx context.Context, resourceID, expand, apiVersion string) (*http.Response, *retry.Error) {
	decorators := []autorest.PrepareDecorator{
		withAPIVersion(apiVersion),
	}
	if expand != "" {
		decorators = append(decorators, autorest.WithQueryParameters(map[string]interface{}{
			"$expand": autorest.Encode("query", expand),
		}))
	}

	return c.GetResource(ctx, resourceID, decorators...)
}

// GetResourceWithQueries get a resource by resource ID with queries.
func (c *Client) GetResourceWithQueries(ctx context.Context, resourceID string, queries map[string]interface{}) (*http.Response, *retry.Error) {

	queryParameters := make(map[string]interface{})
	for queryKey, queryValue := range queries {
		queryParameters[queryKey] = autorest.Encode("query", queryValue)
	}

	decorators := []autorest.PrepareDecorator{
		autorest.WithQueryParameters(queryParameters),
	}

	return c.GetResource(ctx, resourceID, decorators...)
}

// GetResourceWithDecorators get a resource with decorators by resource ID
func (c *Client) GetResource(ctx context.Context, resourceID string, decorators ...autorest.PrepareDecorator) (*http.Response, *retry.Error) {
	getDecorators := append([]autorest.PrepareDecorator{
		autorest.WithPathParameters("{resourceID}", map[string]interface{}{"resourceID": resourceID}),
	}, decorators...)
	request, err := c.PrepareGetRequest(ctx, getDecorators...)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "get.prepare", resourceID, err)
		return nil, retry.NewError(false, err)
	}

	return c.Send(ctx, request, DoHackRegionalRetryForGET(c))
}

// PutResource puts a resource by resource ID
func (c *Client) PutResource(ctx context.Context, resourceID string, parameters interface{}, decorators ...autorest.PrepareDecorator) (*http.Response, *retry.Error) {
	future, rerr := c.PutResourceAsync(ctx, resourceID, parameters, decorators...)
	if rerr != nil {
		return nil, rerr
	}

	response, err := c.WaitForAsyncOperationResult(ctx, future, "armclient.PutResource")
	if err != nil {
		if response != nil {
			klog.V(5).Infof("Received error in WaitForAsyncOperationResult: '%s', response code %d", err.Error(), response.StatusCode)
		} else {
			klog.V(5).Infof("Received error in WaitForAsyncOperationResult: '%s', no response", err.Error())
		}

		retriableErr := retry.GetError(response, err)
		if !retriableErr.Retriable &&
			strings.Contains(strings.ToUpper(err.Error()), strings.ToUpper("InternalServerError")) {
			klog.V(5).Infof("Received InternalServerError in WaitForAsyncOperationResult: '%s', setting error retriable", err.Error())
			retriableErr.Retriable = true
		}
		return nil, retriableErr
	}

	return response, nil
}

func (c *Client) waitAsync(ctx context.Context, futures map[string]*azure.Future, previousResponses map[string]*PutResourcesResponse) {
	wg := sync.WaitGroup{}
	var responseLock sync.Mutex
	for resourceID, future := range futures {
		wg.Add(1)
		go func(resourceID string, future *azure.Future) {
			defer wg.Done()
			response, err := c.WaitForAsyncOperationResult(ctx, future, "armclient.PutResource")
			if err != nil {
				if response != nil {
					klog.V(5).Infof("Received error in WaitForAsyncOperationResult: '%s', response code %d", err.Error(), response.StatusCode)
				} else {
					klog.V(5).Infof("Received error in WaitForAsyncOperationResult: '%s', no response", err.Error())
				}

				retriableErr := retry.GetError(response, err)
				if !retriableErr.Retriable &&
					strings.Contains(strings.ToUpper(err.Error()), strings.ToUpper("InternalServerError")) {
					klog.V(5).Infof("Received InternalServerError in WaitForAsyncOperationResult: '%s', setting error retriable", err.Error())
					retriableErr.Retriable = true
				}

				responseLock.Lock()
				previousResponses[resourceID] = &PutResourcesResponse{
					Error: retriableErr,
				}
				responseLock.Unlock()
				return
			}
		}(resourceID, future)
	}
	wg.Wait()
}

// PutResourcesInBatches is similar with PutResources, but it sends sync request concurrently in batches.
func (c *Client) PutResourcesInBatches(ctx context.Context, resources map[string]interface{}, batchSize int) map[string]*PutResourcesResponse {
	if len(resources) == 0 {
		return nil
	}

	if batchSize <= 0 {
		klog.V(4).Infof("PutResourcesInBatches: batch size %d, put resources in sequence", batchSize)
		batchSize = 1
	}

	if batchSize > len(resources) {
		klog.V(4).Infof("PutResourcesInBatches: batch size %d, but the number of the resources is %d", batchSize, len(resources))
		batchSize = len(resources)
	}
	klog.V(4).Infof("PutResourcesInBatches: send sync requests in parallel with the batch size %d", batchSize)

	rateLimiter := make(chan struct{}, batchSize)

	// Concurrent sync requests in batches.
	futures := make(map[string]*azure.Future)
	responses := make(map[string]*PutResourcesResponse)
	wg := sync.WaitGroup{}
	var responseLock, futuresLock sync.Mutex
	for resourceID, parameters := range resources {
		rateLimiter <- struct{}{}
		wg.Add(1)
		go func(resourceID string, parameters interface{}) {
			defer wg.Done()
			defer func() { <-rateLimiter }()
			future, rerr := c.PutResourceAsync(ctx, resourceID, parameters)
			if rerr != nil {
				responseLock.Lock()
				responses[resourceID] = &PutResourcesResponse{
					Error: rerr,
				}
				responseLock.Unlock()
				return
			}

			futuresLock.Lock()
			futures[resourceID] = future
			futuresLock.Unlock()
		}(resourceID, parameters)
	}
	wg.Wait()
	close(rateLimiter)

	// Concurrent async requests.
	c.waitAsync(ctx, futures, responses)

	return responses
}

// PatchResource patches a resource by resource ID
func (c *Client) PatchResource(ctx context.Context, resourceID string, parameters interface{}, decorators ...autorest.PrepareDecorator) (*http.Response, *retry.Error) {
	future, rerr := c.PatchResourceAsync(ctx, resourceID, parameters, decorators...)
	if rerr != nil {
		return nil, rerr
	}
	response, err := c.WaitForAsyncOperationResult(ctx, future, "armclient.PatchResource")
	if err != nil {
		if response != nil {
			klog.V(5).Infof("Received error in WaitForAsyncOperationResult: '%s', response code %d", err.Error(), response.StatusCode)
		} else {
			klog.V(5).Infof("Received error in WaitForAsyncOperationResult: '%s', no response", err.Error())
		}

		retriableErr := retry.GetError(response, err)
		if !retriableErr.Retriable &&
			strings.Contains(strings.ToUpper(err.Error()), strings.ToUpper("InternalServerError")) {
			klog.V(5).Infof("Received InternalServerError in WaitForAsyncOperationResult: '%s', setting error retriable", err.Error())
			retriableErr.Retriable = true
		}
		return nil, retriableErr
	}

	return response, nil
}

// PatchResourceAsync patches a resource by resource ID asynchronously
func (c *Client) PatchResourceAsync(ctx context.Context, resourceID string, parameters interface{}, decorators ...autorest.PrepareDecorator) (*azure.Future, *retry.Error) {
	decorators = append(decorators,
		autorest.WithPathParameters("{resourceID}", map[string]interface{}{"resourceID": resourceID}),
		autorest.WithJSON(parameters),
	)

	request, err := c.PreparePatchRequest(ctx, decorators...)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "patch.prepare", resourceID, err)
		return nil, retry.NewError(false, err)
	}

	future, resp, clientErr := c.SendAsync(ctx, request)
	defer c.CloseResponse(ctx, resp)
	if clientErr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "patch.send", resourceID, clientErr.Error())
		return nil, clientErr
	}
	return future, clientErr
}

// PutResourceAsync puts a resource by resource ID in async mode
func (c *Client) PutResourceAsync(ctx context.Context, resourceID string, parameters interface{}, decorators ...autorest.PrepareDecorator) (*azure.Future, *retry.Error) {
	decorators = append(decorators,
		autorest.WithPathParameters("{resourceID}", map[string]interface{}{"resourceID": resourceID}),
		autorest.WithJSON(parameters),
	)

	request, err := c.PreparePutRequest(ctx, decorators...)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "put.prepare", resourceID, err)
		return nil, retry.NewError(false, err)
	}

	future, resp, rErr := c.SendAsync(ctx, request)
	defer c.CloseResponse(ctx, resp)
	if rErr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "put.send", resourceID, rErr.Error())
		return nil, rErr
	}

	return future, nil
}

// PostResource posts a resource by resource ID
func (c *Client) PostResource(ctx context.Context, resourceID, action string, parameters interface{}, queryParameters map[string]interface{}) (*http.Response, *retry.Error) {
	pathParameters := map[string]interface{}{
		"resourceID": resourceID,
		"action":     action,
	}

	decorators := []autorest.PrepareDecorator{
		autorest.WithPathParameters("{resourceID}/{action}", pathParameters),
		autorest.WithJSON(parameters),
	}
	if len(queryParameters) > 0 {
		decorators = append(decorators, autorest.WithQueryParameters(queryParameters))
	}

	request, err := c.PreparePostRequest(ctx, decorators...)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "post.prepare", resourceID, err)
		return nil, retry.NewError(false, err)
	}

	return c.Send(ctx, request)
}

// DeleteResource deletes a resource by resource ID
func (c *Client) DeleteResource(ctx context.Context, resourceID string, _ ...autorest.PrepareDecorator) *retry.Error {
	future, clientErr := c.DeleteResourceAsync(ctx, resourceID)
	if clientErr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "delete.request", resourceID, clientErr.Error())
		return clientErr
	}

	if future == nil {
		return nil
	}
	if err := future.WaitForCompletionRef(ctx, c.client); err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "delete.wait", resourceID, err)
		return retry.NewError(true, err)
	}

	return nil
}

// HeadResource heads a resource by resource ID
func (c *Client) HeadResource(ctx context.Context, resourceID string) (*http.Response, *retry.Error) {
	decorators := []autorest.PrepareDecorator{
		autorest.WithPathParameters("{resourceID}", map[string]interface{}{"resourceID": resourceID}),
	}
	request, err := c.PrepareHeadRequest(ctx, decorators...)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "head.prepare", resourceID, err)
		return nil, retry.NewError(false, err)
	}

	return c.Send(ctx, request)
}

// DeleteResourceAsync delete a resource by resource ID and returns a future representing the async result
func (c *Client) DeleteResourceAsync(ctx context.Context, resourceID string, decorators ...autorest.PrepareDecorator) (*azure.Future, *retry.Error) {
	decorators = append(decorators,
		autorest.WithPathParameters("{resourceID}", map[string]interface{}{"resourceID": resourceID}),
	)

	deleteRequest, err := c.PrepareDeleteRequest(ctx, decorators...)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "deleteAsync.prepare", resourceID, err)
		return nil, retry.NewError(false, err)
	}

	resp, rerr := c.Send(ctx, deleteRequest)
	defer c.CloseResponse(ctx, resp)
	if rerr != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "deleteAsync.send", resourceID, rerr.Error())
		return nil, rerr
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusAccepted, http.StatusNoContent, http.StatusNotFound))
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "deleteAsync.respond", resourceID, err)
		return nil, retry.GetError(resp, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	future, err := azure.NewFutureFromResponse(resp)
	if err != nil {
		klog.V(5).Infof("Received error in %s: resourceID: %s, error: %s", "deleteAsync.future", resourceID, err)
		return nil, retry.GetError(resp, err)
	}

	return &future, nil
}

// CloseResponse closes a response
func (c *Client) CloseResponse(_ context.Context, response *http.Response) {
	if response != nil && response.Body != nil {
		if err := response.Body.Close(); err != nil {
			klog.Errorf("Error closing the response body: %v", err)
		}
	}
}

func (c *Client) prepareRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error) {
	decorators = append(
		decorators,
		withAPIVersion(c.apiVersion))
	preparer := autorest.CreatePreparer(decorators...)
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

func withAPIVersion(apiVersion string) autorest.PrepareDecorator {
	const apiVersionKey = "api-version"
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r, err := p.Prepare(r)
			if err == nil {
				if r.URL == nil {
					return r, fmt.Errorf("Error in withAPIVersion: Invoked with a nil URL")
				}

				v := r.URL.Query()
				if len(v.Get(apiVersionKey)) > 0 {
					return r, nil
				}

				v.Add(apiVersionKey, apiVersion)
				r.URL.RawQuery = v.Encode()
			}
			return r, err
		})
	}
}

// GetResourceID gets Azure resource ID
func GetResourceID(subscriptionID, resourceGroupName, resourceType, resourceName string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s",
		autorest.Encode("path", subscriptionID),
		autorest.Encode("path", resourceGroupName),
		resourceType,
		autorest.Encode("path", resourceName))
}

// GetResourceListID gets Azure resource list ID
func GetResourceListID(subscriptionID, resourceGroupName, resourceType string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s",
		autorest.Encode("path", subscriptionID),
		autorest.Encode("path", resourceGroupName),
		resourceType)
}

// GetChildResourceID gets Azure child resource ID
func GetChildResourceID(subscriptionID, resourceGroupName, resourceType, resourceName, childResourceType, childResourceName string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s/%s/%s",
		autorest.Encode("path", subscriptionID),
		autorest.Encode("path", resourceGroupName),
		resourceType,
		autorest.Encode("path", resourceName),
		childResourceType,
		autorest.Encode("path", childResourceName))
}

// GetChildResourcesListID gets Azure child resources list ID
func GetChildResourcesListID(subscriptionID, resourceGroupName, resourceType, resourceName, childResourceType string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s/%s",
		autorest.Encode("path", subscriptionID),
		autorest.Encode("path", resourceGroupName),
		resourceType,
		autorest.Encode("path", resourceName),
		childResourceType)
}

// GetProviderResourceID gets Azure RP resource ID
func GetProviderResourceID(subscriptionID, providerNamespace string) string {
	return fmt.Sprintf("/subscriptions/%s/providers/%s",
		autorest.Encode("path", subscriptionID),
		providerNamespace)
}

// GetProviderResourcesListID gets Azure RP resources list ID
func GetProviderResourcesListID(subscriptionID string) string {
	return fmt.Sprintf("/subscriptions/%s/providers", autorest.Encode("path", subscriptionID))
}
