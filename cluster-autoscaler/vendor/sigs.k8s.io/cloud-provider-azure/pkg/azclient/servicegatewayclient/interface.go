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

// +azure:enableclientgen:=true
package servicegatewayclient

import (
	"context"

	armnetwork "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=get;createorupdate;delete;list,resource=ServiceGateway,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9,packageAlias=armnetwork,clientName=ServiceGatewaysClient,expand=false,rateLimitKey=serviceGatewayRateLimit
type Interface interface {
	utils.GetFunc[armnetwork.ServiceGateway]
	utils.CreateOrUpdateFunc[armnetwork.ServiceGateway]
	utils.DeleteFunc[armnetwork.ServiceGateway]
	utils.ListFunc[armnetwork.ServiceGateway]
	UpdateTags(ctx context.Context, resourceGroupName string, serviceGatewayName string, parameters armnetwork.TagsObject) (*armnetwork.ServiceGateway, error)
	GetAddressLocations(ctx context.Context, resourceGroupName string, serviceGatewayName string) ([]*armnetwork.ServiceGatewayAddressLocationResponse, error)
	GetServices(ctx context.Context, resourceGroupName string, serviceGatewayName string) ([]*armnetwork.ServiceGatewayService, error)
	UpdateAddressLocations(ctx context.Context, resourceGroupName string, serviceGatewayName string, parameters armnetwork.ServiceGatewayUpdateAddressLocationsRequest) error
	UpdateServices(ctx context.Context, resourceGroupName string, serviceGatewayName string, parameters armnetwork.ServiceGatewayUpdateServicesRequest) error
}
