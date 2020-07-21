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

package endpoints

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/pagination"
)

type commonResult struct {
	huaweicloudsdk.Result
}

// Extract interprets a GetResult, CreateResult or UpdateResult as a concrete
// Endpoint. An error is returned if the original call or the extraction failed.
func (r commonResult) Extract() (*Endpoint, error) {
	var s struct {
		Endpoint *Endpoint `json:"endpoint"`
	}
	err := r.ExtractInto(&s)
	return s.Endpoint, err
}

// CreateResult is the response from a Create operation. Call its Extract
// method to interpret it as an Endpoint.
type CreateResult struct {
	commonResult
}

// UpdateResult is the response from an Update operation. Call its Extract
// method to interpret it as an Endpoint.
type UpdateResult struct {
	commonResult
}

// DeleteResult is the response from a Delete operation. Call its ExtractErr
// method to determine if the call succeeded or failed.
type DeleteResult struct {
	huaweicloudsdk.ErrResult
}

// Endpoint describes the entry point for another service's API.
type Endpoint struct {
	// ID is the unique ID of the endpoint.
	ID string `json:"id"`

	// Availability is the interface type of the Endpoint (admin, internal,
	// or public), referenced by the gophercloud.Availability type.
	Availability huaweicloudsdk.Availability `json:"interface"`

	// Name is the name of the Endpoint.
	Name string `json:"name"`

	// Region is the region the Endpoint is located in.
	Region string `json:"region"`

	// ServiceID is the ID of the service the Endpoint refers to.
	ServiceID string `json:"service_id"`

	// URL is the url of the Endpoint.
	URL string `json:"url"`
}

// EndpointPage is a single page of Endpoint results.
type EndpointPage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if no Endpoints were returned.
func (r EndpointPage) IsEmpty() (bool, error) {
	es, err := ExtractEndpoints(r)
	return len(es) == 0, err
}

// ExtractEndpoints extracts an Endpoint slice from a Page.
func ExtractEndpoints(r pagination.Page) ([]Endpoint, error) {
	var s struct {
		Endpoints []Endpoint `json:"endpoints"`
	}
	err := (r.(EndpointPage)).ExtractInto(&s)
	return s.Endpoints, err
}
