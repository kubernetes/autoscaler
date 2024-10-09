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

package diskclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armcompute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

const PatchOperationName = "DisksClient.Patch"

func (client *Client) Patch(ctx context.Context, resourceGroupName string, resourceName string, parameters armcompute.DiskUpdate) (result *armcompute.Disk, err error) {
	ctx = utils.ContextWithClientName(ctx, "DisksClient")
	ctx = utils.ContextWithRequestMethod(ctx, "Patch")
	ctx = utils.ContextWithResourceGroupName(ctx, resourceGroupName)
	ctx = utils.ContextWithSubscriptionID(ctx, client.subscriptionID)
	ctx, endSpan := runtime.StartSpan(ctx, CreateOrUpdateOperationName, client.tracer, nil)
	defer endSpan(err)
	resp, err := utils.NewPollerWrapper(client.DisksClient.BeginUpdate(ctx, resourceGroupName, resourceName, parameters, nil)).WaitforPollerResp(ctx)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return &resp.Disk, nil
	}
	return nil, nil
}
