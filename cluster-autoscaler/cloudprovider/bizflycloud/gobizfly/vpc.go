// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

const (
	vpcPath = "/vpc-networks"
)

var _ VPCService = (*vpcService)(nil)

type vpcService struct {
	client *Client
}

type VPCService interface {
	List(ctx context.Context) ([]*VPC, error)
	Get(ctx context.Context, vpcID string) (*VPC, error)
	Update(ctx context.Context, vpcID string, uvpl *UpdateVPCPayload) (*VPC, error)
	Create(ctx context.Context, cvpl *CreateVPCPayload) (*VPC, error)
	Delete(ctx context.Context, vpcID string) error
}

type VPC struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	TenantID              string   `json:"tenant_id"`
	AdminStateUp          bool     `json:"admin_state_up"`
	MTU                   int      `json:"mtu"`
	Status                string   `json:"status"`
	Subnets               []Subnet `json:"subnets"`
	Shared                bool     `json:"shared"`
	AvailabilityZoneHints []string `json:"availability_zone_hints"`
	AvailabilityZones     []string `json:"availability_zones"`
	IPv4AddressScope      string   `json:"ipv4_address_scope"`
	IPv6AddressScope      string   `json:"ipv6_address_scope"`
	RouterExternal        bool     `json:"router:external"`
	Description           string   `json:"description"`
	PortSecurityEnabled   bool     `json:"port_security_enabled"`
	QosPolicyID           string   `json:"qos_policy_id"`
	Tags                  []string `json:"tags"`
	CreatedAt             string   `json:"created_at"`
	UpdatedAt             string   `json:"updated_at"`
	RevisionNumber        int      `json:"revision_number"`
	IsDefault             bool     `json:"is_default"`
}

type Subnet struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	TenantID        string              `json:"tenant_id"`
	NetworkID       string              `json:"network_id"`
	IPVersion       int                 `json:"ip_version"`
	SubnetPoolID    string              `json:"subnet_pool_id"`
	EnableDHCP      bool                `json:"enable_dhcp"`
	IPv6RaMode      string              `json:"ipv6_ra_mode"`
	IPv6AddressMode string              `json:"ipv6_address_mode"`
	GatewayIP       string              `json:"gateway_ip"`
	CIDR            string              `json:"cidr"`
	AllocationPools []map[string]string `json:"allocation_pools"`
	HostRoutes      []string            `json:"host_routes"`
	DNSNameServers  []string            `json:"dns_nameservers"`
	Description     string              `json:"description"`
	ServiceTypes    []string            `json:"service_types"`
	Tags            []string            `json:"tags"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
	RevisionNumber  int                 `json:"revision_number"`
	ProjectID       string              `json:"project_id"`
}

type CreateVPCPayload struct {
	Name        string `json:"name"`
	CIDR        string `json:"cidr,omitempty"`
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

type UpdateVPCPayload struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	CIDR        string `json:"cidr,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

func (v vpcService) resourcePath() string {
	return vpcPath
}

func (v vpcService) itemPath(id string) string {
	return strings.Join([]string{vpcPath, id}, "/")
}

func (v vpcService) List(ctx context.Context) ([]*VPC, error) {
	req, err := v.client.NewRequest(ctx, http.MethodGet, serverServiceName, v.resourcePath(), nil)
	if err != nil {
		return nil, err
	}
	var data []*VPC
	resp, err := v.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (v vpcService) Get(ctx context.Context, vpcID string) (*VPC, error) {
	req, err := v.client.NewRequest(ctx, http.MethodGet, serverServiceName, v.itemPath(vpcID), nil)
	if err != nil {
		return nil, err
	}
	var data *struct {
		Network *VPC `json:"network"`
	}
	resp, err := v.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Network, nil
}

func (v vpcService) Update(ctx context.Context, vpcID string, uvpl *UpdateVPCPayload) (*VPC, error) {
	req, err := v.client.NewRequest(ctx, http.MethodPut, serverServiceName, v.itemPath(vpcID), uvpl)
	if err != nil {
		return nil, err
	}
	var data *struct {
		Network *VPC `json:"network"`
	}
	resp, err := v.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Network, nil
}

func (v vpcService) Create(ctx context.Context, cvpl *CreateVPCPayload) (*VPC, error) {
	req, err := v.client.NewRequest(ctx, http.MethodPost, serverServiceName, v.resourcePath(), cvpl)
	if err != nil {
		return nil, err
	}
	var data *VPC
	resp, err := v.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (v vpcService) Delete(ctx context.Context, vpcID string) error {
	req, err := v.client.NewRequest(ctx, http.MethodDelete, serverServiceName, v.itemPath(vpcID), nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
