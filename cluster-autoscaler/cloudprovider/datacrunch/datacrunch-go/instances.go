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
	"io"
	"net/http"
	"net/url"
)

// APIError represents an error response from the DataCrunch API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Instance represents a DataCrunch instance.
type Instance struct {
	ID            string   `json:"id"`
	IP            string   `json:"ip"`
	Status        string   `json:"status"`
	CreatedAt     string   `json:"created_at"`
	CPU           CPU      `json:"cpu"`
	GPU           GPU      `json:"gpu"`
	GPUMemory     GPUMem   `json:"gpu_memory"`
	Memory        Memory   `json:"memory"`
	Storage       Storage  `json:"storage"`
	Hostname      string   `json:"hostname"`
	Description   string   `json:"description"`
	Location      string   `json:"location"`
	PricePerHour  float64  `json:"price_per_hour"`
	IsSpot        bool     `json:"is_spot"`
	InstanceType  string   `json:"instance_type"`
	Image         string   `json:"image"`
	OSName        string   `json:"os_name"`
	StartupScript string   `json:"startup_script_id"`
	SSHKeyIDs     []string `json:"ssh_key_ids"`
	OSVolumeID    string   `json:"os_volume_id"`
	JupyterToken  string   `json:"jupyter_token"`
	Contract      string   `json:"contract"`
	Pricing       string   `json:"pricing"`
}

// CPU represents a CPU in DataCrunch.
type CPU struct {
	Description   string `json:"description"`
	NumberOfCores int    `json:"number_of_cores"`
}

// GPU represents a GPU in DataCrunch.
type GPU struct {
	Description  string `json:"description"`
	NumberOfGPUs int    `json:"number_of_gpus"`
}

// GPUMem represents a GPU memory in DataCrunch.
type GPUMem struct {
	Description     string `json:"description"`
	SizeInGigabytes int    `json:"size_in_gigabytes"`
}

// Memory represents a memory in DataCrunch.
type Memory struct {
	Description     string `json:"description"`
	SizeInGigabytes int    `json:"size_in_gigabytes"`
}

// Storage represents a storage in DataCrunch.
type Storage struct {
	Description string `json:"description"`
}

// InstanceList is a list of instances.
type InstanceList []Instance

// DeployInstanceRequest is the request body for creating a new instance.
type DeployInstanceRequest struct {
	InstanceType    string         `json:"instance_type"`
	Image           string         `json:"image"`
	SSHKeyIDs       []string       `json:"ssh_key_ids,omitempty"`
	StartupScriptID string         `json:"startup_script_id,omitempty"`
	Hostname        string         `json:"hostname"`
	Description     string         `json:"description"`
	LocationCode    string         `json:"location_code,omitempty"`
	OSVolume        *OSVolume      `json:"os_volume,omitempty"`
	IsSpot          bool           `json:"is_spot"`
	Coupon          string         `json:"coupon,omitempty"`
	Volumes         []DeployVolume `json:"volumes,omitempty"`
	ExistingVolumes []string       `json:"existing_volumes,omitempty"`
	Contract        string         `json:"contract,omitempty"`
	Pricing         string         `json:"pricing,omitempty"`
}

// OSVolume represents an OS volume in DataCrunch when creating a new instance.
type OSVolume struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// DeployVolume represents a deploy volume in DataCrunch when creating a new instance.
type DeployVolume struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	Type string `json:"type"`
}

// DeployInstanceResponse is the response for creating a new instance (instance ID).
type DeployInstanceResponse string

// InstanceActionRequest is the request body for performing an action on an instance.
type InstanceActionRequest struct {
	Action    string   `json:"action"`
	ID        string   `json:"id"`
	VolumeIDs []string `json:"volume_ids,omitempty"`
}

// InstanceType represents an instance type.
type InstanceType struct {
	BestFor             []string `json:"best_for"`
	CPU                 CPU      `json:"cpu"`
	DeployWarning       string   `json:"deploy_warning"`
	Description         string   `json:"description"`
	GPU                 GPU      `json:"gpu"`
	GPUMemory           GPUMem   `json:"gpu_memory"`
	Memory              Memory   `json:"memory"`
	Model               string   `json:"model"`
	ID                  string   `json:"id"`
	InstanceType        string   `json:"instance_type"`
	Name                string   `json:"name"`
	P2P                 string   `json:"p2p"`
	PricePerHour        string   `json:"price_per_hour"`
	SpotPrice           string   `json:"spot_price"`
	DynamicPrice        string   `json:"dynamic_price"`
	MaxDynamicPrice     string   `json:"max_dynamic_price"`
	ServerlessPrice     string   `json:"serverless_price"`
	ServerlessSpotPrice string   `json:"serverless_spot_price"`
	Storage             Storage  `json:"storage"`
	Currency            string   `json:"currency"`
	Manufacturer        string   `json:"manufacturer"`
	DisplayName         string   `json:"display_name"`
}

// InstanceTypeList is a list of instance types.
type InstanceTypeList []InstanceType

// PriceHistoryEntry represents a price history entry for an instance type.
type PriceHistoryEntry struct {
	Date                string  `json:"date"`
	FixedPricePerHour   float64 `json:"fixed_price_per_hour"`
	DynamicPricePerHour float64 `json:"dynamic_price_per_hour"`
	Currency            string  `json:"currency"`
}

// PriceHistory is a map of instance types to price history entries.
type PriceHistory map[string][]PriceHistoryEntry

// InstanceAvailability represents instance availability for a location.
type InstanceAvailability struct {
	LocationCode   string   `json:"location_code"`
	Availabilities []string `json:"availabilities"`
}

// InstanceAvailabilityList is a list of instance availabilities.
type InstanceAvailabilityList []InstanceAvailability

// ListInstances returns all instances, optionally filtered by status.
func (c *Client) ListInstances(status string) (InstanceList, error) {
	endpoint := c.baseURL + "/instances"
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
	var list InstanceList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	return list, nil
}

// GetInstance returns a single instance by ID.
func (c *Client) GetInstance(id string) (*Instance, error) {
	endpoint := fmt.Sprintf("%s/instances/%s", c.baseURL, id)
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
	var inst Instance
	if err := json.NewDecoder(resp.Body).Decode(&inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

// DeployInstance deploys a new instance and returns its ID.
func (c *Client) DeployInstance(reqBody DeployInstanceRequest) (string, error) {
	endpoint := c.baseURL + "/instances"
	b, err := json.Marshal(reqBody)
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
	id, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(id), nil
}

// PerformInstanceAction performs an action (boot, start, shutdown, delete, etc.) on an instance.
func (c *Client) PerformInstanceAction(reqBody InstanceActionRequest) error {
	endpoint := c.baseURL + "/instances"
	b, err := json.Marshal(reqBody)
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

// ListInstanceTypes returns all available instance types.
func (c *Client) ListInstanceTypes() (InstanceTypeList, error) {
	endpoint := c.baseURL + "/instance-types"
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
	var list InstanceTypeList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	return list, nil
}

// GetInstanceTypePriceHistory returns price history for all instance types.
func (c *Client) GetInstanceTypePriceHistory(currency string, numOfMonths int) (PriceHistory, error) {
	endpoint := c.baseURL + "/instance-types/price-history"
	params := url.Values{}
	if currency != "" {
		params.Set("currency", currency)
	}
	if numOfMonths > 0 {
		params.Set("num_of_months", fmt.Sprintf("%d", numOfMonths))
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
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
	var ph PriceHistory
	if err := json.NewDecoder(resp.Body).Decode(&ph); err != nil {
		return nil, err
	}
	return ph, nil
}

// ListInstanceAvailability returns all instance type availabilities for all locations.
func (c *Client) ListInstanceAvailability(isSpot bool, locationCode string) (InstanceAvailabilityList, error) {
	endpoint := c.baseURL + "/instance-availability"
	params := url.Values{}
	if isSpot {
		params.Set("is_spot", "true")
	}
	if locationCode != "" {
		params.Set("location_code", locationCode)
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
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
	var list InstanceAvailabilityList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	return list, nil
}

// GetInstanceTypeAvailability returns availability for a specific instance type.
func (c *Client) GetInstanceTypeAvailability(instanceType string, isSpot bool, locationCode string) (bool, error) {

	endpoint := fmt.Sprintf("%s/instance-availability/%s", c.baseURL, url.PathEscape(instanceType))
	params := url.Values{}
	if isSpot {
		params.Set("is_spot", "true")
	}
	if locationCode != "" {
		params.Set("location_code", locationCode)
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, c.parseAPIError(resp.Body)
	}
	var available bool

	if err := json.NewDecoder(resp.Body).Decode(&available); err != nil {
		return false, err
	}

	return available, nil
}

// parseAPIError parses an API error response.
func (c *Client) parseAPIError(r io.Reader) error {
	var apiErr APIError
	if err := json.NewDecoder(r).Decode(&apiErr); err != nil {
		return fmt.Errorf("API error: failed to decode error response: %w", err)
	}
	return fmt.Errorf("API error: %s - %s", apiErr.Code, apiErr.Message)
}
