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

// Meta defines the schema of meta information which may be included
// in responses.
type Meta struct {
	Pagination *MetaPagination `json:"pagination"`
}

// MetaPagination defines the schema of pagination information.
type MetaPagination struct {
	Page         int `json:"page"`
	PerPage      int `json:"per_page"`
	PreviousPage int `json:"previous_page"`
	NextPage     int `json:"next_page"`
	LastPage     int `json:"last_page"`
	TotalEntries int `json:"total_entries"`
}

// MetaResponse defines the schema of a response containing
// meta information.
type MetaResponse struct {
	Meta Meta `json:"meta"`
}
