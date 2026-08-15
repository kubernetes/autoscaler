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

// DeployTarget represents a Deploy Target.
type DeployTarget struct {
	Description *string
	ID          *string
	Name        *string
	Type        *string
	Zone        *string
}

func deployTargetFromAPI(d *oapi.DeployTarget, zone string) *DeployTarget {
	return &DeployTarget{
		Description: d.Description,
		ID:          d.Id,
		Name:        d.Name,
		Type:        (*string)(d.Type),
		Zone:        &zone,
	}
}

// FindDeployTarget attempts to find a Deploy Target by name or ID.
func (c *Client) FindDeployTarget(ctx context.Context, zone, x string) (*DeployTarget, error) {
	res, err := c.ListDeployTargets(ctx, zone)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		if *r.ID == x || *r.Name == x {
			return c.GetDeployTarget(ctx, zone, *r.ID)
		}
	}

	return nil, apiv2.ErrNotFound
}

// GetDeployTarget returns the Deploy Target corresponding to the specified ID.
func (c *Client) GetDeployTarget(ctx context.Context, zone, id string) (*DeployTarget, error) {
	resp, err := c.GetDeployTargetWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return deployTargetFromAPI(resp.JSON200, zone), nil
}

// ListDeployTargets returns the list of existing Deploy Targets.
func (c *Client) ListDeployTargets(ctx context.Context, zone string) ([]*DeployTarget, error) {
	list := make([]*DeployTarget, 0)

	resp, err := c.ListDeployTargetsWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.DeployTargets != nil {
		for i := range *resp.JSON200.DeployTargets {
			list = append(list, deployTargetFromAPI(&(*resp.JSON200.DeployTargets)[i], zone))
		}
	}

	return list, nil
}
