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

	apiv2 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2/api"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2/oapi"
)

// Quota represents an Exoscale organization quota.
type Quota struct {
	Resource *string
	Usage    *int64
	Limit    *int64
}

func quotaFromAPI(q *oapi.Quota) *Quota {
	return &Quota{
		Resource: q.Resource,
		Usage:    q.Usage,
		Limit:    q.Limit,
	}
}

// ListQuotas returns the list of Exoscale organization quotas.
func (c *Client) ListQuotas(ctx context.Context, zone string) ([]*Quota, error) {
	list := make([]*Quota, 0)

	resp, err := c.ListQuotasWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.Quotas != nil {
		for i := range *resp.JSON200.Quotas {
			list = append(list, quotaFromAPI(&(*resp.JSON200.Quotas)[i]))
		}
	}

	return list, nil
}

// GetQuota returns the current Exoscale organization quota for the specified resource.
func (c *Client) GetQuota(ctx context.Context, zone, resource string) (*Quota, error) {
	resp, err := c.GetQuotaWithResponse(apiv2.WithZone(ctx, zone), resource)
	if err != nil {
		return nil, err
	}

	return quotaFromAPI(resp.JSON200), nil
}
