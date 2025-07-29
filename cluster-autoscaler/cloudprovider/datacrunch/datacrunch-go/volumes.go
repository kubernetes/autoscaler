/*
Copyright 2019 The Kubernetes Authors.

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

package datacrunchclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Volume represents a DataCrunch volume based on GetVolumePublicResponseDto
type Volume struct {
	ID                       string                  `json:"id"`
	InstanceID               string                  `json:"instance_id"`
	Instances                []VolumeInstanceInfo    `json:"instances"`
	Name                     string                  `json:"name"`
	CreatedAt                string                  `json:"created_at"`
	Status                   string                  `json:"status"`
	Size                     int                     `json:"size"`
	IsOSVolume               bool                    `json:"is_os_volume"`
	Target                   string                  `json:"target"`
	Type                     string                  `json:"type"`
	Location                 string                  `json:"location"`
	SSHKeyIDs                []string                `json:"ssh_key_ids"`
	PseudoPath               string                  `json:"pseudo_path"`
	CreateDirectoryCommand   string                  `json:"create_directory_command"`
	MountCommand             string                  `json:"mount_command"`
	FilesystemToFstabCommand string                  `json:"filesystem_to_fstab_command"`
	Contract                 string                  `json:"contract"`
	BaseHourlyCost           float64                 `json:"base_hourly_cost"`
	MonthlyPrice             float64                 `json:"monthly_price"`
	Currency                 string                  `json:"currency"`
	LongTerm                 *VolumeLongTermContract `json:"long_term,omitempty"`
}

// VolumeInstanceInfo represents instance information attached to a volume
type VolumeInstanceInfo struct {
	ID                  string `json:"id"`
	AutoRentalExtension *bool  `json:"auto_rental_extension"`
	IP                  string `json:"ip"`
	InstanceType        string `json:"instance_type"`
	Status              string `json:"status"`
	OSVolumeID          string `json:"os_volume_id"`
	Hostname            string `json:"hostname"`
}

// VolumeLongTermContract represents long term contract details for a volume
type VolumeLongTermContract struct {
	EndDate             string  `json:"end_date"`
	LongTermPeriod      string  `json:"long_term_period"`
	DiscountPercentage  float64 `json:"discount_percentage"`
	AutoRentalExtension bool    `json:"auto_rental_extension"`
	NextPeriodPrice     float64 `json:"next_period_price"`
	CurrentPeriodPrice  float64 `json:"current_period_price"`
}

// VolumeInTrash represents a volume in trash based on GetVolumeInTrashPublicResponseDto
type VolumeInTrash struct {
	ID             string               `json:"id"`
	InstanceID     string               `json:"instance_id"`
	Instances      []VolumeInstanceInfo `json:"instances"`
	Name           string               `json:"name"`
	CreatedAt      string               `json:"created_at"`
	Status         string               `json:"status"`
	Size           int                  `json:"size"`
	IsOSVolume     bool                 `json:"is_os_volume"`
	Target         string               `json:"target"`
	Type           string               `json:"type"`
	Location       string               `json:"location"`
	SSHKeyIDs      []string             `json:"ssh_key_ids"`
	Contract       string               `json:"contract"`
	BaseHourlyCost float64              `json:"base_hourly_cost"`
	MonthlyPrice   float64              `json:"monthly_price"`
	Currency       string               `json:"currency"`
	DeletedAt      string               `json:"deleted_at"`
}

// CreateVolumeRequest represents the request body for creating a volume based on CreateVolumePublicDto
type CreateVolumeRequest struct {
	Type         string   `json:"type"`
	LocationCode string   `json:"location_code,omitempty"`
	Size         int      `json:"size"`
	InstanceID   string   `json:"instance_id,omitempty"`
	InstanceIDs  []string `json:"instance_ids,omitempty"`
	Name         string   `json:"name"`
}

// VolumeActionRequest represents the request body for performing actions on volumes based on PerformVolumeActionPublicDto
type VolumeActionRequest struct {
	Action       string      `json:"action"`
	ID           interface{} `json:"id"` // Can be string or []string
	Size         *int        `json:"size,omitempty"`
	InstanceID   string      `json:"instance_id,omitempty"`
	InstanceIDs  []string    `json:"instance_ids,omitempty"`
	Name         string      `json:"name,omitempty"`
	Type         string      `json:"type,omitempty"`
	IsPermanent  *bool       `json:"is_permanent,omitempty"`
	LocationCode string      `json:"location_code,omitempty"`
}

// DeleteVolumeRequest represents the request body for deleting a volume based on DeleteVolumePublicDto
type DeleteVolumeRequest struct {
	IsPermanent bool `json:"is_permanent"`
}

// VolumeType represents volume type information based on VolumeType model
type VolumeType struct {
	Type                 string          `json:"type"`
	Price                VolumeTypePrice `json:"price"`
	IsSharedFS           bool            `json:"is_shared_fs"`
	BurstBandwidth       int             `json:"burst_bandwidth"`
	ContinuousBandwidth  int             `json:"continuous_bandwidth"`
	InternalNetworkSpeed int             `json:"internal_network_speed"`
	IOPS                 string          `json:"iops"`
}

// VolumeTypePrice represents pricing information for volume types
type VolumeTypePrice struct {
	PricePerMonthPerGB float64 `json:"price_per_month_per_gb"`
	CPSPerGB           float64 `json:"cps_per_gb"`
	Currency           string  `json:"currency"`
}

// ListVolumes returns all volumes, optionally filtered by status
func (c *Client) ListVolumes(status string) ([]Volume, error) {
	endpoint := c.baseURL + "/volumes"
	if status != "" {
		endpoint += "?status=" + url.QueryEscape(status)
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, c.parseAPIError(resp.Body)
	}
	var volumes []Volume
	if err := json.NewDecoder(resp.Body).Decode(&volumes); err != nil {
		return nil, err
	}
	return volumes, nil
}

// CreateVolume creates a new volume and returns the volume ID
func (c *Client) CreateVolume(request CreateVolumeRequest) (string, error) {
	endpoint := c.baseURL + "/volumes"
	b, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		return "", c.parseAPIError(resp.Body)
	}
	var volumeID string
	if err := json.NewDecoder(resp.Body).Decode(&volumeID); err != nil {
		return "", err
	}
	return volumeID, nil
}

// PerformVolumeAction performs an action on one or multiple volumes
func (c *Client) PerformVolumeAction(request VolumeActionRequest) error {
	endpoint := c.baseURL + "/volumes"
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		return c.parseAPIError(resp.Body)
	}
	return nil
}

// ListVolumesInTrash returns all volumes that are in trash
func (c *Client) ListVolumesInTrash() ([]VolumeInTrash, error) {
	endpoint := c.baseURL + "/volumes/trash"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, c.parseAPIError(resp.Body)
	}
	var volumes []VolumeInTrash
	if err := json.NewDecoder(resp.Body).Decode(&volumes); err != nil {
		return nil, err
	}
	return volumes, nil
}

// GetVolume returns a single volume by ID
func (c *Client) GetVolume(volumeID string) (*Volume, error) {
	endpoint := fmt.Sprintf("%s/volumes/%s", c.baseURL, volumeID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, c.parseAPIError(resp.Body)
	}
	var volume Volume
	if err := json.NewDecoder(resp.Body).Decode(&volume); err != nil {
		return nil, err
	}
	return &volume, nil
}

// DeleteVolume deletes a volume by ID
func (c *Client) DeleteVolume(volumeID string, isPermanent bool) error {
	endpoint := fmt.Sprintf("%s/volumes/%s", c.baseURL, volumeID)
	request := DeleteVolumeRequest{IsPermanent: isPermanent}
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		return c.parseAPIError(resp.Body)
	}
	return nil
}

// ListVolumeTypes returns all available volume types
func (c *Client) ListVolumeTypes() ([]VolumeType, error) {
	endpoint := c.baseURL + "/volume-types"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, c.parseAPIError(resp.Body)
	}
	var volumeTypes []VolumeType
	if err := json.NewDecoder(resp.Body).Decode(&volumeTypes); err != nil {
		return nil, err
	}
	return volumeTypes, nil
}
