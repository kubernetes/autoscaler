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
package accountclient

import (
	"context"

	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=listbyrg,resource=Account,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage,packageAlias=armstorage,clientName=AccountsClient,expand=true,crossSubFactory=true
type Interface interface {
	utils.ListFunc[armstorage.Account]
	Create(ctx context.Context, resourceGroupName string, accountName string, resource *armstorage.AccountCreateParameters) (*armstorage.Account, error)
	GetProperties(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientGetPropertiesOptions) (*armstorage.Account, error)
	Delete(ctx context.Context, resourceGroupName string, accountName string) error
	ListKeys(ctx context.Context, resourceGroupName string, accountName string) ([]*armstorage.AccountKey, error)
}
