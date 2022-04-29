package civogo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ApplicationSize represents an available size for applications to launch
type ApplicationSize struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	NiceName          string `json:"nice_name,omitempty"`
	CPUCores          int    `json:"cpu_cores,omitempty"`
	RAMMegabytes      int    `json:"ram_mb,omitempty"`
	DiskGigabytes     int    `json:"disk_gb,omitempty"`
	TransferTerabytes int    `json:"transfer_tb,omitempty"`
	Description       string `json:"description,omitempty"`
	Selectable        bool   `json:"selectable,omitempty"`
}

// ListApplicationSizes returns all availble sizes of applications
func (c *Client) ListApplicationSizes() ([]ApplicationSize, error) {
	resp, err := c.SendGetRequest("/v2/applications/sizes")
	if err != nil {
		return nil, decodeError(err)
	}

	sizes := make([]ApplicationSize, 0)
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&sizes); err != nil {
		return nil, err
	}

	return sizes, nil
}

// FindApplicationSizes finds a application size name by either part of the ID or part of the name
func (c *Client) FindApplicationSizes(search string) (*ApplicationSize, error) {
	appSize, err := c.ListApplicationSizes()
	if err != nil {
		return nil, decodeError(err)
	}

	exactMatch := false
	partialMatchesCount := 0
	result := ApplicationSize{}

	for _, value := range appSize {
		if value.Name == search || value.ID == search {
			exactMatch = true
			result = value
		} else if strings.Contains(value.Name, search) || strings.Contains(value.ID, search) {
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
