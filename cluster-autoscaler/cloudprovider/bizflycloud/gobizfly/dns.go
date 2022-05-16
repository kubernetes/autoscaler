// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

const (
	zonePath   = "/zones"
	recordPath = "/record"
)

type dnsService struct {
	client *Client
}

var _ DNSService = (*dnsService)(nil)

type DNSService interface {
	ListZones(ctx context.Context, opts *ListOptions) (*ListZoneResp, error)
	CreateZone(ctx context.Context, czpl *createZonePayload) (*ExtendedZone, error)
	GetZone(ctx context.Context, zoneID string) (*ExtendedZone, error)
	DeleteZone(ctx context.Context, zoneID string) error
	CreateRecord(ctx context.Context, zoneID string, crpl *CreateRecordPayload) (*Record, error)
	GetRecord(ctx context.Context, recordID string) (*Record, error)
	UpdateRecord(ctx context.Context, recordID string, urpl *UpdateRecordPayload) (*ExtendedRecord, error)
	DeleteRecord(ctx context.Context, recordID string) error
}

type Zone struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Deleted    int      `json:"deleted"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	TenantId   string   `json:"tenant_id"`
	NameServer []string `json:"nameserver"`
	TTL        int      `json:"ttl"`
	Active     bool     `json:"active"`
	RecordsSet []string `json:"records_set"`
}

type ExtendedZone struct {
	Zone
	RecordsSet []RecordSet `json:"records_set"`
}

type RecordSet struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Type              string            `json:"type"`
	TTL               int               `json:"ttl"`
	Data              []string          `json:"data"`
	RoutingPolicyData RoutingPolicyData `json:"routing_policy_data"`
}
type Meta struct {
	MaxResults int `json:"max_results"`
	Total      int `json:"total"`
	Page       int `json:"page"`
}

type ListZoneResp struct {
	Zones []Zone `json:"zones"`
	Meta  Meta   `json:"_meta"`
}

type createZonePayload struct {
	Name        string `json:"name"`
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
}

type Addrs struct {
	HN  []string `json:"HN"`
	HCM []string `json:"HCM"`
	SG  []string `json:"SG"`
	USA []string `json:"USA"`
}
type RoutingData struct {
	AddrsV4 Addrs `json:"addrs_v4"`
	AddrsV6 Addrs `json:"addrs_v6"`
}

type RoutingPolicyData struct {
	RoutingData RoutingData `json:"routing_data,omitempty"`
	HealthCheck struct {
		TCPConnect struct {
			TCPPort int `json:"tcp_port"`
		} `json:"tcp_connect,omitempty"`
		HTTPStatus struct {
			HTTPPort int    `json:"http_port"`
			URLPath  string `json:"url_path"`
			VHost    string `json:"vhost"`
			OkCodes  []int  `json:"ok_codes"`
			Interval int    `json:"internal"`
		} `json:"http_status,omitempty"`
	} `json:"healthcheck,omitempty"`
}

type CreateRecordPayload struct {
	Name              string            `json:"name"`
	Type              string            `json:"type"`
	TTL               int               `json:"ttl"`
	Data              []string          `json:"data"`
	RoutingPolicyData RoutingPolicyData `json:"routing_policy_data"`
}

type RecordData struct {
	Value    string `json:"value"`
	Priority int    `json:"priority"`
}
type UpdateRecordPayload struct {
	Name              string            `json:"name,omitempty"`
	Type              string            `json:"type,omitempty"`
	TTL               int               `json:"ttl,omitempty"`
	Data              []RecordData      `json:"data,omitempty"`
	RoutingPolicyData RoutingPolicyData `json:"routing_policy_data,omitempty"`
}

type Record struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Delete            int               `json:"deleted"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
	TenantID          string            `json:"tenant_id"`
	ZoneID            string            `json:"zone_id"`
	Type              string            `json:"type"`
	TTL               int               `json:"ttl"`
	Data              []string          `json:"data"`
	RoutingPolicyData RoutingPolicyData `json:"routing_policy_data"`
}

type ExtendedRecord struct {
	Record
	Data []RecordData `json:"data"`
}

type Records struct {
	Records []Record `json:"records"`
}

func (d dnsService) resourcePath() string {
	return zonePath
}

func (d dnsService) zoneItemPath(id string) string {
	return strings.Join([]string{zonePath, id}, "/")
}

func (d dnsService) recordItemPath(id string) string {
	return strings.Join([]string{recordPath, id}, "/")
}

func (d *dnsService) ListZones(ctx context.Context, opts *ListOptions) (*ListZoneResp, error) {
	req, err := d.client.NewRequest(ctx, http.MethodGet, dnsName, d.resourcePath(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data *ListZoneResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *dnsService) CreateZone(ctx context.Context, czpl *createZonePayload) (*ExtendedZone, error) {
	req, err := d.client.NewRequest(ctx, http.MethodPost, dnsName, d.resourcePath(), czpl)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data *ExtendedZone
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *dnsService) GetZone(ctx context.Context, zoneID string) (*ExtendedZone, error) {
	req, err := d.client.NewRequest(ctx, http.MethodGet, dnsName, d.zoneItemPath(zoneID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(ctx, req)
	var data *ExtendedZone
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *dnsService) DeleteZone(ctx context.Context, zoneID string) error {
	req, err := d.client.NewRequest(ctx, http.MethodDelete, dnsName, d.zoneItemPath(zoneID), nil)
	if err != nil {
		return err
	}
	resp, err := d.client.Do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (d *dnsService) CreateRecord(ctx context.Context, zoneID string, crpl *CreateRecordPayload) (*Record, error) {
	req, err := d.client.NewRequest(ctx, http.MethodPost, dnsName, strings.Join([]string{d.zoneItemPath(zoneID), "record"}, "/"), crpl)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data *Record
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *dnsService) GetRecord(ctx context.Context, recordID string) (*Record, error) {
	req, err := d.client.NewRequest(ctx, http.MethodGet, dnsName, d.recordItemPath(recordID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data *Record
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *dnsService) UpdateRecord(ctx context.Context, recordID string, urpl *UpdateRecordPayload) (*ExtendedRecord, error) {
	req, err := d.client.NewRequest(ctx, http.MethodPut, dnsName, d.recordItemPath(recordID), urpl)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data *ExtendedRecord
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *dnsService) DeleteRecord(ctx context.Context, recordID string) error {
	req, err := d.client.NewRequest(ctx, http.MethodDelete, dnsName, d.recordItemPath(recordID), nil)
	if err != nil {
		return err
	}
	resp, err := d.client.Do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
