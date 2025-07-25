package datacrunchclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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
