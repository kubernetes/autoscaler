/*
Copyright 2020 The Kubernetes Authors.

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

package fileclient

import (
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-02-01/storage"
)

// Interface is the client interface for creating file shares, interface for test injection.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	CreateFileShare(resourceGroupName, accountName string, shareOptions *ShareOptions) error
	DeleteFileShare(resourceGroupName, accountName, name string) error
	ResizeFileShare(resourceGroupName, accountName, name string, sizeGiB int) error
	GetFileShare(resourceGroupName, accountName, name string) (storage.FileShare, error)
	GetServiceProperties(resourceGroupName, accountName string) (storage.FileServiceProperties, error)
	SetServiceProperties(resourceGroupName, accountName string, parameters storage.FileServiceProperties) (storage.FileServiceProperties, error)
}
