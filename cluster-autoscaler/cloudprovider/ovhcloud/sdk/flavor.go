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

import (
	"context"
	"fmt"
)

// Flavor defines instances types available on OVHcloud
type Flavor struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	State    string `json:"state"`
}

// ListFlavors allows to display flavors available for nodes templates
func (c *Client) ListFlavors(ctx context.Context, projectID string, clusterID string) ([]Flavor, error) {
	flavors := make([]Flavor, 0)

	return flavors, c.CallAPIWithContext(
		ctx,
		"GET",
		fmt.Sprintf("/cloud/project/%s/kube/%s/flavors", projectID, clusterID),
		nil,
		&flavors,
		nil,
		true,
	)
}
