/*
Copyright 2026 The Kubernetes Authors.

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

package servicegatewayclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armnetwork "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/metrics"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

const UpdateTagsOperationName = "ServiceGatewaysClient.UpdateTags"

// UpdateTags updates the tags of a ServiceGateway.
func (client *Client) UpdateTags(ctx context.Context, resourceGroupName string, serviceGatewayName string, parameters armnetwork.TagsObject) (result *armnetwork.ServiceGateway, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "ServiceGateway", "update_tags")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, UpdateTagsOperationName, client.tracer, nil)
	defer endSpan(err)
	resp, err := client.ServiceGatewaysClient.UpdateTags(ctx, resourceGroupName, serviceGatewayName, parameters, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceGateway, nil
}

const GetAddressLocationsOperationName = "ServiceGatewaysClient.GetAddressLocations"

// GetAddressLocations gets the address locations of a ServiceGateway.
func (client *Client) GetAddressLocations(ctx context.Context, resourceGroupName string, serviceGatewayName string) (result []*armnetwork.ServiceGatewayAddressLocationResponse, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "ServiceGateway", "get_address_locations")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, GetAddressLocationsOperationName, client.tracer, nil)
	defer endSpan(err)
	pager := client.NewGetAddressLocationsPager(resourceGroupName, serviceGatewayName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, nextResult.Value...)
	}
	return result, nil
}

const GetServicesOperationName = "ServiceGatewaysClient.GetServices"

// GetServices gets the services of a ServiceGateway.
func (client *Client) GetServices(ctx context.Context, resourceGroupName string, serviceGatewayName string) (result []*armnetwork.ServiceGatewayService, err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "ServiceGateway", "get_services")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, GetServicesOperationName, client.tracer, nil)
	defer endSpan(err)
	pager := client.NewGetServicesPager(resourceGroupName, serviceGatewayName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, nextResult.Value...)
	}
	return result, nil
}

const UpdateAddressLocationsOperationName = "ServiceGatewaysClient.UpdateAddressLocations"

// UpdateAddressLocations updates the address locations of a ServiceGateway.
func (client *Client) UpdateAddressLocations(ctx context.Context, resourceGroupName string, serviceGatewayName string, parameters armnetwork.ServiceGatewayUpdateAddressLocationsRequest) (err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "ServiceGateway", "update_address_locations")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, UpdateAddressLocationsOperationName, client.tracer, nil)
	defer endSpan(err)
	_, err = utils.NewPollerWrapper(client.BeginUpdateAddressLocations(ctx, resourceGroupName, serviceGatewayName, parameters, nil)).WaitforPollerResp(ctx)
	return err
}

const UpdateServicesOperationName = "ServiceGatewaysClient.UpdateServices"

// UpdateServices updates the services of a ServiceGateway.
func (client *Client) UpdateServices(ctx context.Context, resourceGroupName string, serviceGatewayName string, parameters armnetwork.ServiceGatewayUpdateServicesRequest) (err error) {
	metricsCtx := metrics.BeginARMRequest(client.subscriptionID, resourceGroupName, "ServiceGateway", "update_services")
	defer func() { metricsCtx.Observe(ctx, err) }()
	ctx, endSpan := runtime.StartSpan(ctx, UpdateServicesOperationName, client.tracer, nil)
	defer endSpan(err)
	_, err = utils.NewPollerWrapper(client.BeginUpdateServices(ctx, resourceGroupName, serviceGatewayName, parameters, nil)).WaitforPollerResp(ctx)
	return err
}
