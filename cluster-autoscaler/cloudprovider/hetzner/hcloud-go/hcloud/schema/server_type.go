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

// ServerType defines the schema of a server type.
type ServerType struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Cores       int                      `json:"cores"`
	Memory      float32                  `json:"memory"`
	Disk        int                      `json:"disk"`
	StorageType string                   `json:"storage_type"`
	CPUType     string                   `json:"cpu_type"`
	Prices      []PricingServerTypePrice `json:"prices"`
}

// ServerTypeListResponse defines the schema of the response when
// listing server types.
type ServerTypeListResponse struct {
	ServerTypes []ServerType `json:"server_types"`
}

// ServerTypeGetResponse defines the schema of the response when
// retrieving a single server type.
type ServerTypeGetResponse struct {
	ServerType ServerType `json:"server_type"`
}
