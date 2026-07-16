/*
Copyright 2023 The Kubernetes Authors.

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

package blobcontainerclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/metrics"
)

const ListOperationName = "BlobContainersClient.List"
const CreateOperationName = "BlobContainersClient.Create"
const DeleteOperationName = "BlobContainersClient.Delete"

// List gets a list of BlobContainer in the resource group.
func (client *Client) List(ctx context.Context, resourceGroupName string, parentResourceName string) (result []*armstorage.ListContainerItem, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "BlobContainer", "list")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, ListOperationName, client.tracer, nil)
	defer endSpan(err)

	pager := client.NewListPager(resourceGroupName, parentResourceName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, nextResult.Value...)
	}
	return result, nil
}

func (client *Client) CreateContainer(ctx context.Context, resourceGroupName, accountName, containerName string, parameters armstorage.BlobContainer) (result *armstorage.BlobContainer, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "BlobContainer", "create")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, CreateOperationName, client.tracer, nil)
	defer endSpan(err)

	resp, err := client.Create(ctx, resourceGroupName, accountName, containerName, parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.BlobContainer, nil
}

func (client *Client) DeleteContainer(ctx context.Context, resourceGroupName, accountName, containerName string) (err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "BlobContainer", "delete")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, DeleteOperationName, client.tracer, nil)
	defer endSpan(err)

	_, err = client.Delete(ctx, resourceGroupName, accountName, containerName, nil)
	return err
}
