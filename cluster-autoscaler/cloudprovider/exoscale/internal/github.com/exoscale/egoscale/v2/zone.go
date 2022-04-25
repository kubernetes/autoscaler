/*
Copyright 2021 The Kubernetes Authors.

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
	"context"
)

// ListZones returns the list of Exoscale zones.
func (c *Client) ListZones(ctx context.Context) ([]string, error) {
	list := make([]string, 0)

	resp, err := c.ListZonesWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if resp.JSON200.Zones != nil {
		for i := range *resp.JSON200.Zones {
			zone := &(*resp.JSON200.Zones)[i]
			list = append(list, string(*zone.Name))
		}
	}

	return list, nil
}
