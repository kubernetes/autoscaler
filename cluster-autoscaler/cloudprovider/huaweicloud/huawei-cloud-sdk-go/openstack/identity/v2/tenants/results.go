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

package tenants

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/pagination"
)

// Tenant is a grouping of users in the identity service.
type Tenant struct {
	// ID is a unique identifier for this tenant.
	ID string `json:"id"`

	// Name is a friendlier user-facing name for this tenant.
	Name string `json:"name"`

	// Description is a human-readable explanation of this Tenant's purpose.
	Description string `json:"description"`

	// Enabled indicates whether or not a tenant is active.
	Enabled bool `json:"enabled"`
}

// TenantPage is a single page of Tenant results.
type TenantPage struct {
	pagination.LinkedPageBase
}

// IsEmpty determines whether or not a page of Tenants contains any results.
func (r TenantPage) IsEmpty() (bool, error) {
	tenants, err := ExtractTenants(r)
	return len(tenants) == 0, err
}

// NextPageURL extracts the "next" link from the tenants_links section of the result.
func (r TenantPage) NextPageURL() (string, error) {
	var s struct {
		Links []huaweicloudsdk.Link `json:"tenants_links"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return huaweicloudsdk.ExtractNextURL(s.Links)
}

// ExtractTenants returns a slice of Tenants contained in a single page of
// results.
func ExtractTenants(r pagination.Page) ([]Tenant, error) {
	var s struct {
		Tenants []Tenant `json:"tenants"`
	}
	err := (r.(TenantPage)).ExtractInto(&s)
	return s.Tenants, err
}

type tenantResult struct {
	huaweicloudsdk.Result
}

// Extract interprets any tenantResults as a Tenant.
func (r tenantResult) Extract() (*Tenant, error) {
	var s struct {
		Tenant *Tenant `json:"tenant"`
	}
	err := r.ExtractInto(&s)
	return s.Tenant, err
}

// GetResult is the response from a Get request. Call its Extract method to
// interpret it as a Tenant.
type GetResult struct {
	tenantResult
}

// CreateResult is the response from a Create request. Call its Extract method
// to interpret it as a Tenant.
type CreateResult struct {
	tenantResult
}

// DeleteResult is the response from a Get request. Call its ExtractErr method
// to determine if the call succeeded or failed.
type DeleteResult struct {
	huaweicloudsdk.ErrResult
}

// UpdateResult is the response from a Update request. Call its Extract method
// to interpret it as a Tenant.
type UpdateResult struct {
	tenantResult
}
