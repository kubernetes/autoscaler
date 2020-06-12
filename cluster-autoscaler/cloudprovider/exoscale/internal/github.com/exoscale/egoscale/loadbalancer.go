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

package egoscale

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	apiv2 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/api/v2"
	v2 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale/internal/v2"
)

// NetworkLoadBalancerServerStatus represents a Network Load Balancer service target server status.
type NetworkLoadBalancerServerStatus struct {
	InstanceIP net.IP
	Status     string
}

func nlbServerStatusFromAPI(st *v2.LoadBalancerServerStatus) *NetworkLoadBalancerServerStatus {
	return &NetworkLoadBalancerServerStatus{
		InstanceIP: net.ParseIP(optionalString(st.PublicIp)),
		Status:     optionalString(st.Status),
	}
}

// NetworkLoadBalancerServiceHealthcheck represents a Network Load Balancer service healthcheck.
type NetworkLoadBalancerServiceHealthcheck struct {
	Mode     string
	Port     uint16
	Interval time.Duration
	Timeout  time.Duration
	Retries  int64
	URI      string
	TLSSNI   string
}

// NetworkLoadBalancerService represents a Network Load Balancer service.
type NetworkLoadBalancerService struct {
	ID                string
	Name              string
	Description       string
	InstancePoolID    string
	Protocol          string
	Port              uint16
	TargetPort        uint16
	Strategy          string
	Healthcheck       NetworkLoadBalancerServiceHealthcheck
	State             string
	HealthcheckStatus []*NetworkLoadBalancerServerStatus
}

func nlbServiceFromAPI(svc *v2.LoadBalancerService) *NetworkLoadBalancerService {
	return &NetworkLoadBalancerService{
		ID:             optionalString(svc.Id),
		Name:           optionalString(svc.Name),
		Description:    optionalString(svc.Description),
		InstancePoolID: optionalString(svc.InstancePool.Id),
		Protocol:       optionalString(svc.Protocol),
		Port:           uint16(optionalInt64(svc.Port)),
		TargetPort:     uint16(optionalInt64(svc.TargetPort)),
		Strategy:       optionalString(svc.Strategy),
		Healthcheck: NetworkLoadBalancerServiceHealthcheck{
			Mode:     optionalString(svc.Healthcheck.Mode),
			Port:     uint16(optionalInt64(svc.Healthcheck.Port)),
			Interval: time.Duration(optionalInt64(svc.Healthcheck.Interval)) * time.Second,
			Timeout:  time.Duration(optionalInt64(svc.Healthcheck.Timeout)) * time.Second,
			Retries:  optionalInt64(svc.Healthcheck.Retries),
			URI:      optionalString(svc.Healthcheck.Uri),
			TLSSNI:   optionalString(svc.Healthcheck.TlsSni),
		},
		HealthcheckStatus: func() []*NetworkLoadBalancerServerStatus {
			statuses := make([]*NetworkLoadBalancerServerStatus, 0)

			if svc.HealthcheckStatus != nil {
				for _, st := range *svc.HealthcheckStatus {
					st := st
					statuses = append(statuses, nlbServerStatusFromAPI(&st))
				}
			}

			return statuses
		}(),
		State: optionalString(svc.State),
	}
}

// NetworkLoadBalancer represents a Network Load Balancer instance.
type NetworkLoadBalancer struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	IPAddress   net.IP
	Services    []*NetworkLoadBalancerService
	State       string

	c    *Client
	zone string
}

func nlbFromAPI(nlb *v2.LoadBalancer) *NetworkLoadBalancer {
	return &NetworkLoadBalancer{
		ID:          optionalString(nlb.Id),
		Name:        optionalString(nlb.Name),
		Description: optionalString(nlb.Description),
		CreatedAt:   *nlb.CreatedAt,
		IPAddress:   net.ParseIP(optionalString(nlb.Ip)),
		State:       optionalString(nlb.State),
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
	}
}

// AddService adds a service to the Network Load Balancer instance.
func (nlb *NetworkLoadBalancer) AddService(ctx context.Context,
	svc *NetworkLoadBalancerService) (*NetworkLoadBalancerService, error) {
	var (
		port                = int64(svc.Port)
		targetPort          = int64(svc.TargetPort)
		healthcheckPort     = int64(svc.Healthcheck.Port)
		healthcheckInterval = int64(svc.Healthcheck.Interval.Seconds())
		healthcheckTimeout  = int64(svc.Healthcheck.Timeout.Seconds())
	)

	// The API doesn't return the NLB service created directly, so in order to return a
	// *NetworkLoadBalancerService corresponding to the new service we have to manually
	// compare the list of services on the NLB instance before and after the service
	// creation, and identify the service that wasn't there before.
	// Note: in case of multiple services creation in parallel this technique is subject
	// to race condition as we could return an unrelated service. To prevent this, we
	// also compare the name of the new service to the name specified in the svc
	// parameter.
	services := make(map[string]struct{})
	for _, svc := range nlb.Services {
		services[svc.ID] = struct{}{}
	}

	resp, err := nlb.c.v2.AddServiceToLoadBalancerWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		v2.AddServiceToLoadBalancerJSONRequestBody{
			Name:        &svc.Name,
			Description: &svc.Description,
			Healthcheck: &v2.Healthcheck{
				Mode:     &svc.Healthcheck.Mode,
				Port:     &healthcheckPort,
				Interval: &healthcheckInterval,
				Timeout:  &healthcheckTimeout,
				Retries:  &svc.Healthcheck.Retries,
				Uri: func() *string {
					if strings.HasPrefix(svc.Healthcheck.Mode, "http") {
						return &svc.Healthcheck.URI
					}
					return nil
				}(),
				TlsSni: func() *string {
					if svc.Healthcheck.Mode == "https" {
						return &svc.Healthcheck.TLSSNI
					}
					return nil
				}(),
			},
			InstancePool: &v2.Resource{Id: &svc.InstancePoolID},
			Port:         &port,
			TargetPort:   &targetPort,
			Protocol:     &svc.Protocol,
			Strategy:     &svc.Strategy,
		})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	res, err := v2.NewPoller().
		WithTimeout(nlb.c.Timeout).
		Poll(ctx, nlb.c.v2.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	nlbUpdated, err := nlb.c.GetNetworkLoadBalancer(ctx, nlb.zone, *res.(*v2.Reference).Id)
	if err != nil {
		return nil, err
	}

	// Look for an unknown service: if we find one we hope it's the one we've just created.
	for _, s := range nlbUpdated.Services {
		if _, ok := services[svc.ID]; !ok && s.Name == svc.Name {
			return s, nil
		}
	}

	return nil, errors.New("unable to identify the service created")
}

// UpdateService updates the specified Network Load Balancer service.
func (nlb *NetworkLoadBalancer) UpdateService(ctx context.Context, svc *NetworkLoadBalancerService) error {
	var (
		port                = int64(svc.Port)
		targetPort          = int64(svc.TargetPort)
		healthcheckPort     = int64(svc.Healthcheck.Port)
		healthcheckInterval = int64(svc.Healthcheck.Interval.Seconds())
		healthcheckTimeout  = int64(svc.Healthcheck.Timeout.Seconds())
	)

	resp, err := nlb.c.v2.UpdateLoadBalancerServiceWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		svc.ID,
		v2.UpdateLoadBalancerServiceJSONRequestBody{
			Name:        &svc.Name,
			Description: &svc.Description,
			Port:        &port,
			TargetPort:  &targetPort,
			Protocol:    &svc.Protocol,
			Strategy:    &svc.Strategy,
			Healthcheck: &v2.Healthcheck{
				Mode:     &svc.Healthcheck.Mode,
				Port:     &healthcheckPort,
				Interval: &healthcheckInterval,
				Timeout:  &healthcheckTimeout,
				Retries:  &svc.Healthcheck.Retries,
				Uri: func() *string {
					if strings.HasPrefix(svc.Healthcheck.Mode, "http") {
						return &svc.Healthcheck.URI
					}
					return nil
				}(),
				TlsSni: func() *string {
					if svc.Healthcheck.Mode == "https" {
						return &svc.Healthcheck.TLSSNI
					}
					return nil
				}(),
			},
		})
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	_, err = v2.NewPoller().
		WithTimeout(nlb.c.Timeout).
		Poll(ctx, nlb.c.v2.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteService deletes the specified service from the Network Load Balancer instance.
func (nlb *NetworkLoadBalancer) DeleteService(ctx context.Context, svc *NetworkLoadBalancerService) error {
	resp, err := nlb.c.v2.DeleteLoadBalancerServiceWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		svc.ID,
	)
	if err != nil {
		return err
	}

	_, err = v2.NewPoller().
		WithTimeout(nlb.c.Timeout).
		Poll(ctx, nlb.c.v2.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateNetworkLoadBalancer creates a Network Load Balancer instance in the specified zone.
func (c *Client) CreateNetworkLoadBalancer(ctx context.Context, zone string,
	nlb *NetworkLoadBalancer) (*NetworkLoadBalancer, error) {
	resp, err := c.v2.CreateLoadBalancerWithResponse(
		apiv2.WithZone(ctx, zone),
		v2.CreateLoadBalancerJSONRequestBody{
			Name:        &nlb.Name,
			Description: &nlb.Description,
		})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	res, err := v2.NewPoller().
		WithTimeout(c.Timeout).
		Poll(ctx, c.v2.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetNetworkLoadBalancer(ctx, zone, *res.(*v2.Reference).Id)
}

// ListNetworkLoadBalancers returns the list of existing Network Load Balancers in the
// specified zone.
func (c *Client) ListNetworkLoadBalancers(ctx context.Context, zone string) ([]*NetworkLoadBalancer, error) {
	var list = make([]*NetworkLoadBalancer, 0)

	resp, err := c.v2.ListLoadBalancersWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	if resp.JSON200.LoadBalancers != nil {
		for i := range *resp.JSON200.LoadBalancers {
			nlb := nlbFromAPI(&(*resp.JSON200.LoadBalancers)[i])
			nlb.c = c
			nlb.zone = zone

			list = append(list, nlb)
		}
	}

	return list, nil
}

// GetNetworkLoadBalancer returns the Network Load Balancer instance corresponding to the
// specified ID in the specified zone.
func (c *Client) GetNetworkLoadBalancer(ctx context.Context, zone, id string) (*NetworkLoadBalancer, error) {
	resp, err := c.v2.GetLoadBalancerWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return nil, ErrNotFound

		default:
			return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
		}
	}

	nlb := nlbFromAPI(resp.JSON200)
	nlb.c = c
	nlb.zone = zone

	return nlb, nil
}

// UpdateNetworkLoadBalancer updates the specified Network Load Balancer instance in the specified zone.
func (c *Client) UpdateNetworkLoadBalancer(ctx context.Context, zone string,
	nlb *NetworkLoadBalancer) (*NetworkLoadBalancer, error) {
	resp, err := c.v2.UpdateLoadBalancerWithResponse(
		apiv2.WithZone(ctx, zone),
		nlb.ID,
		v2.UpdateLoadBalancerJSONRequestBody{
			Name:        &nlb.Name,
			Description: &nlb.Description,
		})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return nil, ErrNotFound

		default:
			return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
		}
	}

	res, err := v2.NewPoller().
		WithTimeout(c.Timeout).
		Poll(ctx, c.v2.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetNetworkLoadBalancer(ctx, zone, *res.(*v2.Reference).Id)
}

// DeleteNetworkLoadBalancer deletes the specified Network Load Balancer instance in the specified zone.
func (c *Client) DeleteNetworkLoadBalancer(ctx context.Context, zone, id string) error {
	resp, err := c.v2.DeleteLoadBalancerWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return ErrNotFound

		default:
			return fmt.Errorf("unexpected response from API: %s", resp.Status())
		}
	}

	_, err = v2.NewPoller().
		WithTimeout(c.Timeout).
		Poll(ctx, c.v2.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
