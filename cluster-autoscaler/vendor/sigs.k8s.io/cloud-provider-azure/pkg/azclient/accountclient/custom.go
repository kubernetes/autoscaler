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

package accountclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

const CreateOperationName = "AccountsClient.Create"
const UpdateOperationName = "AccountsClient.Update"
const GetPropertiesOperationName = "AccountsClient.GetProperties"
const DeleteOperationName = "AccountsClient.Delete"
const ListKeysOperationName = "AccountsClient.ListKeys"

func (client *Client) Create(ctx context.Context, resourceGroupName string, resourceName string, resource *armstorage.AccountCreateParameters) (result *armstorage.Account, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "Account", "create")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, CreateOperationName, client.tracer, nil)
	defer endSpan(err)

	if resource == nil {
		resource = &armstorage.AccountCreateParameters{}
	}
	resp, err := utils.NewPollerWrapper(client.AccountsClient.BeginCreate(ctx, resourceGroupName, resourceName, *resource, nil)).WaitforPollerResp(ctx)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return &resp.Account, nil
	}
	return nil, nil
}

func (client *Client) Update(ctx context.Context, resourceGroupName string, resourceName string, parameters *armstorage.AccountUpdateParameters) (result *armstorage.Account, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "Account", "update")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, UpdateOperationName, client.tracer, nil)
	defer endSpan(err)

	if parameters == nil {
		parameters = &armstorage.AccountUpdateParameters{}
	}
	resp, err := client.AccountsClient.Update(ctx, resourceGroupName, resourceName, *parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Account, nil
}

func (client *Client) GetProperties(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientGetPropertiesOptions) (result *armstorage.Account, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "Account", "getProperties")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, GetPropertiesOperationName, client.tracer, nil)
	defer endSpan(err)

	resp, err := client.AccountsClient.GetProperties(ctx, resourceGroupName, accountName, options)
	if err != nil {
		return nil, err
	}
	//handle statuscode
	return &resp.Account, nil
}

// Delete deletes a Interface by name.
func (client *Client) Delete(ctx context.Context, resourceGroupName string, resourceName string) (err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "Account", "delete")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, DeleteOperationName, client.tracer, nil)
	defer endSpan(err)

	_, err = client.AccountsClient.Delete(ctx, resourceGroupName, resourceName, nil)
	return err
}

func (client *Client) ListKeys(ctx context.Context, resourceGroupName string, accountName string) (result []*armstorage.AccountKey, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "Account", "listKeys")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, ListKeysOperationName, client.tracer, nil)
	defer endSpan(err)

	resp, err := client.AccountsClient.ListKeys(ctx, resourceGroupName, accountName, nil)
	if err != nil {
		return nil, err
	}
	return resp.Keys, nil
}
