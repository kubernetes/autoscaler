// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

const (
	firewallBasePath = "/firewalls"
)

var _ FirewallService = (*firewall)(nil)

type firewall struct {
	client *Client
}

type BaseFirewall struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Tags           []string       `json:"tags"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
	RevisionNumber int            `json:"revision_number"`
	ProjectID      string         `json:"project_id"`
	ServersCount   int            `json:"servers_count"`
	RulesCount     int            `json:"rules_count"`
	InBound        []FirewallRule `json:"inbound"`
	OutBound       []FirewallRule `json:"outbound"`
}

type Firewall struct {
	BaseFirewall
	Servers []string `json:"servers"`
}

type FirewallDetail struct {
	BaseFirewall
	Servers []*Server `json:"servers"`
}

type FirewallRule struct {
	ID             string   `json:"id"`
	FirewallID     string   `json:"security_group_id"`
	EtherType      string   `json:"ethertype"`
	Direction      string   `json:"direction"`
	Protocol       string   `json:"protocol"`
	PortRangeMin   int      `json:"port_range_min"`
	PortRangeMax   int      `json:"port_range_max"`
	RemoteIPPrefix string   `json:"remote_ip_prefix"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	RevisionNumber int      `json:"revision_number"`
	ProjectID      string   `json:"project_id"`
	Type           string   `json:"type"`
	CIDR           string   `json:"cidr"`
	PortRange      string   `json:"port_range"`
}

type FirewallRuleCreateRequest struct {
	Type      string `json:"type"`
	Protocol  string `json:"protocol"`
	PortRange string `json:"port_range"`
	CIDR      string `json:"cidr"`
}

type FirewallSingleRuleCreateRequest struct {
	FirewallRuleCreateRequest
	Direction string `json:"direction"`
}

type FirewallRequestPayload struct {
	Name     string                      `json:"name"`
	InBound  []FirewallRuleCreateRequest `json:"inbound,omitempty"`
	OutBound []FirewallRuleCreateRequest `json:"outbound,omitempty"`
	Targets  []string                    `json:"targets,omitempty"`
}

type FirewallDeleteResponse struct {
	Message string `json:"message"`
}

type FirewallRemoveServerRequest struct {
	Servers []string `json:"servers"`
}

type FirewallRuleCreateResponse struct {
	Rule FirewallRule `json:"security_group_rule"`
}

type FirewallService interface {
	List(ctx context.Context, opts *ListOptions) ([]*Firewall, error)
	Create(ctx context.Context, fcr *FirewallRequestPayload) (*FirewallDetail, error)
	Get(ctx context.Context, id string) (*FirewallDetail, error)
	Delete(ctx context.Context, id string) (*FirewallDeleteResponse, error)
	RemoveServer(ctx context.Context, id string, rsfr *FirewallRemoveServerRequest) (*Firewall, error)
	Update(ctx context.Context, id string, ufr *FirewallRequestPayload) (*FirewallDetail, error)
	DeleteRule(ctx context.Context, id string) (*FirewallDeleteResponse, error)
	CreateRule(ctx context.Context, fwID string, fsrcr *FirewallSingleRuleCreateRequest) (*FirewallRule, error)
}

// List lists all firewall.
func (f *firewall) List(ctx context.Context, opts *ListOptions) ([]*Firewall, error) {

	req, err := f.client.NewRequest(ctx, http.MethodGet, serverServiceName, firewallBasePath, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var firewalls []*Firewall

	if err := json.NewDecoder(resp.Body).Decode(&firewalls); err != nil {
		return nil, err
	}

	return firewalls, nil
}

// Create a firewall.
func (f *firewall) Create(ctx context.Context, fcr *FirewallRequestPayload) (*FirewallDetail, error) {

	req, err := f.client.NewRequest(ctx, http.MethodPost, serverServiceName, firewallBasePath, fcr)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var firewall *FirewallDetail

	if err := json.NewDecoder(resp.Body).Decode(&firewall); err != nil {
		return nil, err
	}

	return firewall, nil
}

// Get detail a firewall.
func (f *firewall) Get(ctx context.Context, id string) (*FirewallDetail, error) {

	req, err := f.client.NewRequest(ctx, http.MethodGet, serverServiceName, firewallBasePath+"/"+id, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var firewall *FirewallDetail

	if err := json.NewDecoder(resp.Body).Decode(&firewall); err != nil {
		return nil, err
	}

	return firewall, nil
}

// Remove servers from a firewall.
func (f *firewall) RemoveServer(ctx context.Context, id string, rsfr *FirewallRemoveServerRequest) (*Firewall, error) {

	req, err := f.client.NewRequest(ctx, http.MethodDelete, serverServiceName, strings.Join([]string{firewallBasePath, id, "servers"}, "/"), rsfr)

	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var firewall *Firewall

	if err := json.NewDecoder(resp.Body).Decode(&firewall); err != nil {
		return nil, err
	}

	return firewall, nil
}

// Update Firewall
func (f *firewall) Update(ctx context.Context, id string, ufr *FirewallRequestPayload) (*FirewallDetail, error) {

	req, err := f.client.NewRequest(ctx, http.MethodPatch, serverServiceName, firewallBasePath+"/"+id, ufr)

	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var firewall *FirewallDetail

	if err := json.NewDecoder(resp.Body).Decode(&firewall); err != nil {
		return nil, err
	}

	return firewall, nil
}

// Delete a Firewall
func (f *firewall) Delete(ctx context.Context, id string) (*FirewallDeleteResponse, error) {

	req, err := f.client.NewRequest(ctx, http.MethodDelete, serverServiceName, firewallBasePath+"/"+id, nil)

	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var dwr *FirewallDeleteResponse

	if err := json.NewDecoder(resp.Body).Decode(&dwr); err != nil {
		return nil, err
	}

	return dwr, nil
}

// Delete a rule in a firewall
func (f *firewall) DeleteRule(ctx context.Context, id string) (*FirewallDeleteResponse, error) {
	req, err := f.client.NewRequest(ctx, http.MethodDelete, serverServiceName, firewallBasePath+"/"+id, nil)

	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var dwr *FirewallDeleteResponse

	if err := json.NewDecoder(resp.Body).Decode(&dwr); err != nil {
		return nil, err
	}

	return dwr, nil
}

// Create a new rule in a firewall
func (f *firewall) CreateRule(ctx context.Context, fwID string, fsrcr *FirewallSingleRuleCreateRequest) (*FirewallRule, error) {
	req, err := f.client.NewRequest(ctx, http.MethodPost, serverServiceName, strings.Join([]string{firewallBasePath, fwID, "rules"}, "/"), fsrcr)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var firewallRuleCreateResponse struct {
		Rule *FirewallRule `json:"security_group_rule"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&firewallRuleCreateResponse); err != nil {
		return nil, err
	}
	return firewallRuleCreateResponse.Rule, nil
}
