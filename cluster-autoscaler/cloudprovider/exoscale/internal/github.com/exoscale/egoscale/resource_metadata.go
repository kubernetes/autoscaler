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

package egoscale

import "fmt"

// ResourceDetail represents extra details
type ResourceDetail ResourceTag

// ListRequest builds the ListResourceDetails request
func (detail ResourceDetail) ListRequest() (ListCommand, error) {
	if detail.ResourceType == "" {
		return nil, fmt.Errorf("the resourcetype parameter is required")
	}

	req := &ListResourceDetails{
		ResourceType: detail.ResourceType,
		ResourceID:   detail.ResourceID,
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListResourceDetails

// ListResourceDetails lists the resource tag(s) (but different from listTags...)
type ListResourceDetails struct {
	ResourceType string `json:"resourcetype" doc:"list by resource type"`
	ForDisplay   bool   `json:"fordisplay,omitempty" doc:"if set to true, only details marked with display=true, are returned. False by default"`
	Key          string `json:"key,omitempty" doc:"list by key"`
	Keyword      string `json:"keyword,omitempty" doc:"List by keyword"`
	Page         int    `json:"page,omitempty"`
	PageSize     int    `json:"pagesize,omitempty"`
	ResourceID   *UUID  `json:"resourceid,omitempty" doc:"list by resource id"`
	Value        string `json:"value,omitempty" doc:"list by key, value. Needs to be passed only along with key"`
	_            bool   `name:"listResourceDetails" description:"List resource detail(s)"`
}

// ListResourceDetailsResponse represents a list of resource details
type ListResourceDetailsResponse struct {
	Count          int           `json:"count"`
	ResourceDetail []ResourceTag `json:"resourcedetail"`
}
