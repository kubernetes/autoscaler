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
package blobcontainerclient

import (
	"context"

	armstorage "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// +azure:client:verbs=get,resource=Account,subResource=BlobContainer,packageName=github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage,packageAlias=armstorage,clientName=BlobContainersClient,expand=false,crossSubFactory=true
type Interface interface {
	utils.SubResourceGetFunc[armstorage.BlobContainer]
	CreateContainer(ctx context.Context, resourceGroupName, accountName, containerName string, parameters armstorage.BlobContainer) (*armstorage.BlobContainer, error)
	DeleteContainer(ctx context.Context, resourceGroupName, accountName, containerName string) error
	List(ctx context.Context, resourceGroupName string, parentResourceName string) (result []*armstorage.ListContainerItem, rerr error)
}
