/*
Copyright 2022 The Kubernetes Authors.

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

package govultr

// ListOptions are the available fields that can be used with pagination
type ListOptions struct {
	PerPage int    `url:"per_page,omitempty"`
	Cursor  string `url:"cursor,omitempty"`
}

// Meta represents the available pagination information
type Meta struct {
	Total int `json:"total"`
	Links *Links
}

// Links represent the next/previous cursor in your pagination calls
type Links struct {
	Next string `json:"next"`
	Prev string `json:"prev"`
}
