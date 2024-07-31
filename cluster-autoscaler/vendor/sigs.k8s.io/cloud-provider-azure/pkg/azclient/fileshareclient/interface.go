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
package fileshareclient

import (
	"context"

	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=get,resource=Account,subResource=FileShare,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage,packageAlias=armstorage,clientName=FileSharesClient,expand=false,crossSubFactory=true
type Interface interface {
	utils.SubResourceGetFunc[armstorage.FileShare]
	Create(ctx context.Context, resourceGroupName string, resourceName string, parentResourceName string, resource armstorage.FileShare) (*armstorage.FileShare, error)
	Update(ctx context.Context, resourceGroupName string, resourceName string, parentResourceName string, resource armstorage.FileShare) (*armstorage.FileShare, error)
	Delete(ctx context.Context, resourceGroupName string, parentResourceName string, resourceName string) error
}
