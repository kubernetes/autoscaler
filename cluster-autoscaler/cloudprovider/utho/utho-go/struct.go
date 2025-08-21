/*
Copyright 2025 The Kubernetes Authors.

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

package utho

// BasicResponse represents a basic API response with status and message.
type BasicResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// CreateResponse represents the response returned after creating a resource.
type CreateResponse struct {
	ID        string `json:"id"`
	AppStatus string `json:"app_status,omitempty"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// CreateBasicResponse represents a basic response for create operations.
type CreateBasicResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// UpdateResponse represents the response returned after updating a resource.
type UpdateResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DeleteResponse represents the response returned after deleting a resource.
type DeleteResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// Dclocation represents a datacenter location.
type Dclocation struct {
	Location string `json:"location"`
	Country  string `json:"country"`
	Dc       string `json:"dc"`
	Dccc     string `json:"dccc"`
}
