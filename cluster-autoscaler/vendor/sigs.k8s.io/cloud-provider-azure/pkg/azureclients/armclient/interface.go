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
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// PutResourcesResponse defines the response for PutResources.
type PutResourcesResponse struct {
	Response *http.Response
	Error    *retry.Error
}

// Interface is the client interface for ARM.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	// Send sends a http request to ARM service with possible retry to regional ARM endpoint.
	Send(ctx context.Context, request *http.Request, decorators ...autorest.SendDecorator) (*http.Response, *retry.Error)

	// PreparePutRequest prepares put request
	PreparePutRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error)

	// PreparePostRequest prepares post request
	PreparePostRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error)

	// PrepareGetRequest prepares get request
	PrepareGetRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error)

	// PrepareDeleteRequest preparse delete request
	PrepareDeleteRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error)

	// PrepareHeadRequest prepares head request
	PrepareHeadRequest(ctx context.Context, decorators ...autorest.PrepareDecorator) (*http.Request, error)

	// WaitForAsyncOperationCompletion waits for an operation completion
	WaitForAsyncOperationCompletion(ctx context.Context, future *azure.Future, asyncOperationName string) error

	// WaitForAsyncOperationResult waits for an operation result.
	WaitForAsyncOperationResult(ctx context.Context, future *azure.Future, asyncOperationName string) (*http.Response, error)

	// SendAsync send a request and return a future object representing the async result as well as the origin http response
	SendAsync(ctx context.Context, request *http.Request) (*azure.Future, *http.Response, *retry.Error)

	// PutResource puts a resource by resource ID
	PutResource(ctx context.Context, resourceID string, parameters interface{}, decorators ...autorest.PrepareDecorator) (*http.Response, *retry.Error)

	// PutResourceAsync puts a resource by resource ID in async mode
	PutResourceAsync(ctx context.Context, resourceID string, parameters interface{}, decorators ...autorest.PrepareDecorator) (*azure.Future, *retry.Error)

	// PutResourcesInBatches is similar with PutResources, but it sends sync request concurrently in batches.
	PutResourcesInBatches(ctx context.Context, resources map[string]interface{}, batchSize int) map[string]*PutResourcesResponse

	// PatchResource patches a resource by resource ID
	PatchResource(ctx context.Context, resourceID string, parameters interface{}, decorators ...autorest.PrepareDecorator) (*http.Response, *retry.Error)

	// PatchResourceAsync patches a resource by resource ID asynchronously
	PatchResourceAsync(ctx context.Context, resourceID string, parameters interface{}, decorators ...autorest.PrepareDecorator) (*azure.Future, *retry.Error)

	// HeadResource heads a resource by resource ID
	HeadResource(ctx context.Context, resourceID string) (*http.Response, *retry.Error)

	// GetResourceWithExpandQuery get a resource by resource ID with expand
	GetResourceWithExpandQuery(ctx context.Context, resourceID, expand string) (*http.Response, *retry.Error)

	// GetResourceWithExpandAPIVersionQuery get a resource by resource ID with expand and API version.
	GetResourceWithExpandAPIVersionQuery(ctx context.Context, resourceID, expand, apiVersion string) (*http.Response, *retry.Error)

	// GetResourceWithQueries get a resource by resource ID with queries.
	GetResourceWithQueries(ctx context.Context, resourceID string, queries map[string]interface{}) (*http.Response, *retry.Error)

	// GetResource get a resource with decorators by resource ID
	GetResource(ctx context.Context, resourceID string, decorators ...autorest.PrepareDecorator) (*http.Response, *retry.Error)

	// PostResource posts a resource by resource ID
	PostResource(ctx context.Context, resourceID, action string, parameters interface{}, queryParameters map[string]interface{}) (*http.Response, *retry.Error)

	// DeleteResource deletes a resource by resource ID
	DeleteResource(ctx context.Context, resourceID string, decorators ...autorest.PrepareDecorator) *retry.Error

	// DeleteResourceAsync delete a resource by resource ID and returns a future representing the async result
	DeleteResourceAsync(ctx context.Context, resourceID string, decorators ...autorest.PrepareDecorator) (*azure.Future, *retry.Error)

	// CloseResponse closes a response
	CloseResponse(ctx context.Context, response *http.Response)
}
