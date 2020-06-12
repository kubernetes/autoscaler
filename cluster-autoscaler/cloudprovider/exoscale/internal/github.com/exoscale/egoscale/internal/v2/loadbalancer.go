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

package v2

import (
	"encoding/json"
	"time"
)

// UnmarshalJSON unmarshals a LoadBalancer structure into a temporary structure whose "CreatedAt" field of type
// string to be able to parse the original timestamp (ISO 8601) into a time.Time object, since json.Unmarshal()
// only supports RFC 3339 format.
func (lb *LoadBalancer) UnmarshalJSON(data []byte) error {
	var raw = struct {
		CreatedAt   string                 `json:"created-at,omitempty"`
		Description *string                `json:"description,omitempty"`
		Id          *string                `json:"id,omitempty"` // nolint:golint
		Ip          *string                `json:"ip,omitempty"` // nolint:golint
		Name        *string                `json:"name,omitempty"`
		Services    *[]LoadBalancerService `json:"services,omitempty"`
		State       *string                `json:"state,omitempty"`
	}{}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	createdAt, err := time.Parse(iso8601Format, raw.CreatedAt)
	if err != nil {
		return err
	}

	lb.CreatedAt = &createdAt
	lb.Description = raw.Description
	lb.Id = raw.Id
	lb.Ip = raw.Ip
	lb.Name = raw.Name
	lb.Services = raw.Services
	lb.State = raw.State

	return nil
}

// MarshalJSON returns the JSON encoding of a LoadBalancer structure after having formatted the CreatedAt field
// in the original timestamp (ISO 8601), since time.MarshalJSON() only supports RFC 3339 format.
func (lb *LoadBalancer) MarshalJSON() ([]byte, error) {
	var raw = struct {
		CreatedAt   string                 `json:"created-at,omitempty"`
		Description *string                `json:"description,omitempty"`
		Id          *string                `json:"id,omitempty"` // nolint:golint
		Ip          *string                `json:"ip,omitempty"` // nolint:golint
		Name        *string                `json:"name,omitempty"`
		Services    *[]LoadBalancerService `json:"services,omitempty"`
		State       *string                `json:"state,omitempty"`
	}{}

	raw.CreatedAt = lb.CreatedAt.Format(iso8601Format)
	raw.Description = lb.Description
	raw.Id = lb.Id
	raw.Ip = lb.Ip
	raw.Name = lb.Name
	raw.Services = lb.Services
	raw.State = lb.State

	return json.Marshal(raw)
}
