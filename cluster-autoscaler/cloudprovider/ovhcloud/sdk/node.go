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

package sdk

import "time"

// Node defines the instance deployed on OVHcloud
type Node struct {
	ID         string `json:"id"`
	InstanceID string `json:"instanceId"`
	NodePoolID string `json:"nodePoolId"`
	ProjectID  string `json:"projectId"`

	Name     string `json:"name"`
	Flavor   string `json:"flavor"`
	Version  string `json:"version"`
	UpToDate bool   `json:"isUpToDate"`
	Status   string `json:"status"`

	IP        *string `json:"ip,omitempty"`
	PrivateIP *string `json:"privateIp,omitempty"`

	CreatedAt  time.Time `json:"createdAt"`
	DeployedAt time.Time `json:"deployedAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
