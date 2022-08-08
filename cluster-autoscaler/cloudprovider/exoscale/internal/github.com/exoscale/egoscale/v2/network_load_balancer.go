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
	"net"
	"time"

	apiv2 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2/api"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/v2/oapi"
)

// NetworkLoadBalancer represents a Network Load Balancer.
type NetworkLoadBalancer struct {
	CreatedAt   *time.Time
	Description *string
	ID          *string `req-for:"update,delete"`
	IPAddress   *net.IP
	Labels      *map[string]string
	Name        *string `req-for:"create"`
	Services    []*NetworkLoadBalancerService
	State       *string
	Zone        *string
}

func nlbFromAPI(nlb *oapi.LoadBalancer, zone string) *NetworkLoadBalancer {
	return &NetworkLoadBalancer{
		CreatedAt:   nlb.CreatedAt,
		Description: nlb.Description,
		ID:          nlb.Id,
		IPAddress: func() (v *net.IP) {
			if nlb.Ip != nil {
				ip := net.ParseIP(*nlb.Ip)
				v = &ip
			}
			return
		}(),
		Labels: func() (v *map[string]string) {
			if nlb.Labels != nil && len(nlb.Labels.AdditionalProperties) > 0 {
				v = &nlb.Labels.AdditionalProperties
			}
			return
		}(),
		Name: nlb.Name,
		Services: func() []*NetworkLoadBalancerService {
			services := make([]*NetworkLoadBalancerService, 0)
			if nlb.Services != nil {
				for _, svc := range *nlb.Services {
					svc := svc
					services = append(services, nlbServiceFromAPI(&svc))
				}
			}
			return services
		}(),
		State: (*string)(nlb.State),
		Zone:  &zone,
	}
}

// CreateNetworkLoadBalancer creates a Network Load Balancer.
func (c *Client) CreateNetworkLoadBalancer(
	ctx context.Context,
	zone string,
	nlb *NetworkLoadBalancer,
) (*NetworkLoadBalancer, error) {
	if err := validateOperationParams(nlb, "create"); err != nil {
		return nil, err
	}

	resp, err := c.CreateLoadBalancerWithResponse(
		apiv2.WithZone(ctx, zone),
		oapi.CreateLoadBalancerJSONRequestBody{
			Description: nlb.Description,
			Labels: func() (v *oapi.Labels) {
				if nlb.Labels != nil {
					v = &oapi.Labels{AdditionalProperties: *nlb.Labels}
				}
				return
			}(),
			Name: *nlb.Name,
		})
	if err != nil {
		return nil, err
	}

	res, err := oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, oapi.OperationPoller(c, zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetNetworkLoadBalancer(ctx, zone, *res.(*oapi.Reference).Id)
}

// DeleteNetworkLoadBalancer deletes a Network Load Balancer.
func (c *Client) DeleteNetworkLoadBalancer(ctx context.Context, zone string, nlb *NetworkLoadBalancer) error {
	if err := validateOperationParams(nlb, "delete"); err != nil {
		return err
	}

	resp, err := c.DeleteLoadBalancerWithResponse(apiv2.WithZone(ctx, zone), *nlb.ID)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, oapi.OperationPoller(c, zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// FindNetworkLoadBalancer attempts to find a Network Load Balancer by name or ID.
func (c *Client) FindNetworkLoadBalancer(ctx context.Context, zone, x string) (*NetworkLoadBalancer, error) {
	res, err := c.ListNetworkLoadBalancers(ctx, zone)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		if *r.ID == x || *r.Name == x {
			return c.GetNetworkLoadBalancer(ctx, zone, *r.ID)
		}
	}

	return nil, apiv2.ErrNotFound
}

// GetNetworkLoadBalancer returns the Network Load Balancer corresponding to the specified ID.
func (c *Client) GetNetworkLoadBalancer(ctx context.Context, zone, id string) (*NetworkLoadBalancer, error) {
	resp, err := c.GetLoadBalancerWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return nlbFromAPI(resp.JSON200, zone), nil
}

// ListNetworkLoadBalancers returns the list of existing Network Load Balancers in the specified zone.
func (c *Client) ListNetworkLoadBalancers(ctx context.Context, zone string) ([]*NetworkLoadBalancer, error) {
	list := make([]*NetworkLoadBalancer, 0)

	resp, err := c.ListLoadBalancersWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.LoadBalancers != nil {
		for i := range *resp.JSON200.LoadBalancers {
			list = append(list, nlbFromAPI(&(*resp.JSON200.LoadBalancers)[i], zone))
		}
	}

	return list, nil
}

// UpdateNetworkLoadBalancer updates a Network Load Balancer.
func (c *Client) UpdateNetworkLoadBalancer(ctx context.Context, zone string, nlb *NetworkLoadBalancer) error {
	if err := validateOperationParams(nlb, "update"); err != nil {
		return err
	}

	resp, err := c.UpdateLoadBalancerWithResponse(
		apiv2.WithZone(ctx, zone),
		*nlb.ID,
		oapi.UpdateLoadBalancerJSONRequestBody{
			Description: oapi.NilableString(nlb.Description),
			Labels: func() (v *oapi.Labels) {
				if nlb.Labels != nil {
					v = &oapi.Labels{AdditionalProperties: *nlb.Labels}
				}
				return
			}(),
			Name: nlb.Name,
		})
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, oapi.OperationPoller(c, zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
