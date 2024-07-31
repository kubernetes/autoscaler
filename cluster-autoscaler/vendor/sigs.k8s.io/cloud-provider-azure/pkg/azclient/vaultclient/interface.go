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
package vaultclient

import (
	"context"

	armkeyvault "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=get;listbyrg,resource=Vault,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault,packageAlias=armkeyvault,clientName=VaultsClient,expand=false
type Interface interface {
	utils.GetFunc[armkeyvault.Vault]
	CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, resource armkeyvault.VaultCreateOrUpdateParameters) (*armkeyvault.Vault, error)
	utils.DeleteFunc[armkeyvault.Vault]
	PurgeDeleted(ctx context.Context, vaultName string, location string) error
	utils.ListFunc[armkeyvault.Vault]
}
