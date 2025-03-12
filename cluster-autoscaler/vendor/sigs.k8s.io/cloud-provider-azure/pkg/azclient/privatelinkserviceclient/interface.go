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

// +azure:enableclientgen:=true
package privatelinkserviceclient

import (
	armnetwork "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=get;createorupdate;delete;list,resource=PrivateLinkService,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4,packageAlias=armnetwork,clientName=PrivateLinkServicesClient,expand=true,rateLimitKey=privateLinkServiceRateLimit
type Interface interface {
	utils.GetWithExpandFunc[armnetwork.PrivateLinkService]
	utils.CreateOrUpdateFunc[armnetwork.PrivateLinkService]
	utils.DeleteFunc[armnetwork.PrivateLinkService]
	utils.ListFunc[armnetwork.PrivateLinkService]
}
