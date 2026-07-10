/*
Copyright 2024 The Kubernetes Authors.

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

package privatelinkserviceclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

const DeletePEConnectionOperationName = "PrivateLinkServicesClient.DeletePrivateEndpointConnection"

func (client *Client) DeletePrivateEndpointConnection(ctx context.Context, resourceGroupName string, serviceName string, peConnectionName string) (err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "PrivateLinkService", "deletePrivateEndpointConnection")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, DeletePEConnectionOperationName, client.tracer, nil)
	defer endSpan(err)

	_, err = utils.NewPollerWrapper(
		client.BeginDeletePrivateEndpointConnection(ctx, resourceGroupName, serviceName, peConnectionName, nil),
	).WaitforPollerResp(ctx)

	return err
}
