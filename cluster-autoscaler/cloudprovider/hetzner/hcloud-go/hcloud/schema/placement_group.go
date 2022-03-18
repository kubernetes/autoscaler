/*
Copyright 2018 The Kubernetes Authors.

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

package schema

import "time"

type PlacementGroup struct {
	ID      int               `json:"id"`
	Name    string            `json:"name"`
	Labels  map[string]string `json:"labels"`
	Created time.Time         `json:"created"`
	Servers []int             `json:"servers"`
	Type    string            `json:"type"`
}

type PlacementGroupListResponse struct {
	PlacementGroups []PlacementGroup `json:"placement_groups"`
}

type PlacementGroupGetResponse struct {
	PlacementGroup PlacementGroup `json:"placement_group"`
}

type PlacementGroupCreateRequest struct {
	Name   string             `json:"name"`
	Labels *map[string]string `json:"labels,omitempty"`
	Type   string             `json:"type"`
}

type PlacementGroupCreateResponse struct {
	PlacementGroup PlacementGroup `json:"placement_group"`
	Action         *Action        `json:"action"`
}

type PlacementGroupUpdateRequest struct {
	Name   *string            `json:"name,omitempty"`
	Labels *map[string]string `json:"labels,omitempty"`
}

type PlacementGroupUpdateResponse struct {
	PlacementGroup PlacementGroup `json:"placement_group"`
}
