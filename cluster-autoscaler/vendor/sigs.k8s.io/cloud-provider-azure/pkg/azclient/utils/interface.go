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

package utils

import "context"

// Get gets the service resource
type GetFunc[Type interface{}] interface {
	Get(ctx context.Context, resourceGroupName string, resourceName string) (result *Type, rerr error)
}

// Get gets the service resource
type SubResourceGetFunc[Type interface{}] interface {
	Get(ctx context.Context, resourceGroupName string, parentResourceName string, resourceName string) (result *Type, rerr error)
}

// Get gets the service resource
type GetWithExpandFunc[Type interface{}] interface {
	Get(ctx context.Context, resourceGroupName string, resourceName string, expand *string) (result *Type, rerr error)
}

// Get gets the service resource
type SubResourceGetWithExpandFunc[Type interface{}] interface {
	Get(ctx context.Context, resourceGroupName string, parentResourceName string, resourceName string, expand *string) (result *Type, rerr error)
}

// List gets a list of service resource in the resource group.
type ListFunc[Type interface{}] interface {
	List(ctx context.Context, resourceGroupName string) (result []*Type, rerr error)
}

// List gets a list of service resource in the resource group.
type SubResourceListFunc[Type interface{}] interface {
	List(ctx context.Context, resourceGroupName string, parentResourceName string) (result []*Type, rerr error)
}

// CreateOrUpdate creates or updates a service resource.
type CreateOrUpdateFunc[Type interface{}] interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, resourceParam Type) (*Type, error)
}

// CreateOrUpdate creates or updates a service resource.
type SubResourceCreateOrUpdateFunc[Type interface{}] interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, parentResourceName string, resourceName string, resourceParam Type) (*Type, error)
}

// Delete deletes a service resource by name.
type DeleteFunc[Type interface{}] interface {
	Delete(ctx context.Context, resourceGroupName string, resourceName string) error
}

// Delete deletes a service resource by name.
type SubResourceDeleteFunc[Type interface{}] interface {
	Delete(ctx context.Context, resourceGroupName string, parentResourceName string, resourceName string) error
}
