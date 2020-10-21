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

// ISO defines the schema of an ISO image.
type ISO struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Deprecated  time.Time `json:"deprecated"`
}

// ISOGetResponse defines the schema of the response when retrieving a single ISO.
type ISOGetResponse struct {
	ISO ISO `json:"iso"`
}

// ISOListResponse defines the schema of the response when listing ISOs.
type ISOListResponse struct {
	ISOs []ISO `json:"isos"`
}
