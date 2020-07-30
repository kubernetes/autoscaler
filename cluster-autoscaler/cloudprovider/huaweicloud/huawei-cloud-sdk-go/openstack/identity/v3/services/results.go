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

package services

import (
	"encoding/json"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/internal"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/pagination"
)

type serviceResult struct {
	huaweicloudsdk.Result
}

// Extract interprets a GetResult, CreateResult or UpdateResult as a concrete
// Service. An error is returned if the original call or the extraction failed.
func (r serviceResult) Extract() (*Service, error) {
	var s struct {
		Service *Service `json:"service"`
	}
	err := r.ExtractInto(&s)
	return s.Service, err
}

// CreateResult is the response from a Create request. Call its Extract method
// to interpret it as a Service.
type CreateResult struct {
	serviceResult
}

// GetResult is the response from a Get request. Call its Extract method
// to interpret it as a Service.
type GetResult struct {
	serviceResult
}

// UpdateResult is the response from an Update request. Call its Extract method
// to interpret it as a Service.
type UpdateResult struct {
	serviceResult
}

// DeleteResult is the response from a Delete request. Call its ExtractErr
// method to interpret it as a Service.
type DeleteResult struct {
	huaweicloudsdk.ErrResult
}

// Service represents an OpenStack Service.
type Service struct {
	// ID is the unique ID of the service.
	ID string `json:"id"`

	// Type is the type of the service.
	Name string `json:"name"`

	// Type is the type of the service.
	Type string `json:"type"`

	// Enabled is whether or not the service is enabled.
	Enabled bool `json:"enabled"`

	// Links contains referencing links to the service.
	Links map[string]interface{} `json:"links"`

	// Extra is a collection of miscellaneous key/values.
	Extra map[string]interface{} `json:"-"`
}

// UnmarshalJSON ...
func (r *Service) UnmarshalJSON(b []byte) error {
	type tmp Service
	var s struct {
		tmp
		Extra map[string]interface{} `json:"extra"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Service(s.tmp)

	// Collect other fields and bundle them into Extra
	// but only if a field titled "extra" wasn't sent.
	if s.Extra != nil {
		r.Extra = s.Extra
	} else {
		var result interface{}
		err := json.Unmarshal(b, &result)
		if err != nil {
			return err
		}
		if resultMap, ok := result.(map[string]interface{}); ok {
			r.Extra = internal.RemainingKeys(Service{}, resultMap)
		}
	}

	return err
}

// ServicePage is a single page of Service results.
type ServicePage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if the ServicePage contains no results.
func (p ServicePage) IsEmpty() (bool, error) {
	services, err := ExtractServices(p)
	return len(services) == 0, err
}

// NextPageURL extracts the "next" link from the links section of the result.
func (p ServicePage) NextPageURL() (string, error) {
	var s struct {
		Links struct {
			Next     string `json:"next"`
			Previous string `json:"previous"`
		} `json:"links"`
	}
	err := p.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return s.Links.Next, err
}

// ExtractServices extracts a slice of Services from a Collection acquired
// from List.
func ExtractServices(r pagination.Page) ([]Service, error) {
	var s struct {
		Services []Service `json:"services"`
	}
	err := (r.(ServicePage)).ExtractInto(&s)
	return s.Services, err
}
