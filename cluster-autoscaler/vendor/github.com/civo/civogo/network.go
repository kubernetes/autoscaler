package civogo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Network represents a private network for instances to connect to
type Network struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Default bool   `json:"default,omitempty"`
	CIDR    string `json:"cidr,omitempty"`
	Label   string `json:"label,omitempty"`
	Status  string `json:"status,omitempty"`
}

type networkConfig struct {
	Label  string `json:"label"`
	Region string `json:"region"`
}

// NetworkResult represents the result from a network create/update call
type NetworkResult struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Result string `json:"result"`
}

// GetDefaultNetwork finds the default private network for an account
func (c *Client) GetDefaultNetwork() (*Network, error) {
	resp, err := c.SendGetRequest("/v2/networks")
	if err != nil {
		return nil, decodeError(err)
	}

	networks := make([]Network, 0)
	json.NewDecoder(bytes.NewReader(resp)).Decode(&networks)
	for _, network := range networks {
		if network.Default {
			return &network, nil
		}
	}

	return nil, errors.New("no default network found")
}

// NewNetwork creates a new private network
func (c *Client) NewNetwork(label string) (*NetworkResult, error) {
	nc := networkConfig{Label: label, Region: c.Region}
	body, err := c.SendPostRequest("/v2/networks", nc)
	if err != nil {
		return nil, decodeError(err)
	}

	var result = &NetworkResult{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// ListNetworks list all private networks
func (c *Client) ListNetworks() ([]Network, error) {
	resp, err := c.SendGetRequest("/v2/networks")
	if err != nil {
		return nil, decodeError(err)
	}

	networks := make([]Network, 0)
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&networks); err != nil {
		return nil, err
	}

	return networks, nil
}

// FindNetwork finds a network by either part of the ID or part of the name
func (c *Client) FindNetwork(search string) (*Network, error) {
	networks, err := c.ListNetworks()
	if err != nil {
		return nil, decodeError(err)
	}

	exactMatch := false
	partialMatchesCount := 0
	result := Network{}

	for _, value := range networks {
		if value.Name == search || value.ID == search || value.Label == search {
			exactMatch = true
			result = value
		} else if strings.Contains(value.Name, search) || strings.Contains(value.ID, search) || strings.Contains(value.Label, search) {
			if !exactMatch {
				result = value
				partialMatchesCount++
			}
		}
	}

	if exactMatch || partialMatchesCount == 1 {
		return &result, nil
	} else if partialMatchesCount > 1 {
		err := fmt.Errorf("unable to find %s because there were multiple matches", search)
		return nil, MultipleMatchesError.wrap(err)
	} else {
		err := fmt.Errorf("unable to find %s, zero matches", search)
		return nil, ZeroMatchesError.wrap(err)
	}
}

// RenameNetwork renames an existing private network
func (c *Client) RenameNetwork(label, id string) (*NetworkResult, error) {
	nc := networkConfig{Label: label, Region: c.Region}
	body, err := c.SendPutRequest("/v2/networks/"+id, nc)
	if err != nil {
		return nil, decodeError(err)
	}

	var result = &NetworkResult{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteNetwork deletes a private network
func (c *Client) DeleteNetwork(id string) (*SimpleResponse, error) {
	resp, err := c.SendDeleteRequest(fmt.Sprintf("/v2/networks/%s", id))
	if err != nil {
		return nil, decodeError(err)
	}

	return c.DecodeSimpleResponse(resp)
}
