// cloudprovider/datacrunch/datacrunch-go/sshkeys.go

package datacrunchclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SSHKey represents a DataCrunch SSH key.
type SSHKey struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// ListSSHKeys returns all SSH keys for the project.
func (c *Client) ListSSHKeys() ([]SSHKey, error) {
	endpoint := c.baseURL + "/sshkeys"
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
	var keys []SSHKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}
	return keys, nil
}

// AddSSHKey adds a new SSH key and returns its ID.
func (c *Client) AddSSHKey(name, key string) (string, error) {
	endpoint := c.baseURL + "/sshkeys"
	body := map[string]string{
		"name": name,
		"key":  key,
	}
	b, _ := json.Marshal(body)
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
	if resp.StatusCode != 201 {
		return "", c.parseAPIError(resp.Body)
	}
	var id string
	if err := json.NewDecoder(resp.Body).Decode(&id); err != nil {
		return "", err
	}
	return id, nil
}

// DeleteSSHKeys deletes multiple SSH keys by their IDs.
func (c *Client) DeleteSSHKeys(ids []string) error {
	endpoint := c.baseURL + "/sshkeys"
	body := map[string][]string{
		"keys": ids,
	}
	b, _ := json.Marshal(body)
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
	if resp.StatusCode != 200 {
		return c.parseAPIError(resp.Body)
	}
	return nil
}

// GetSSHKey returns a single SSH key by ID.
func (c *Client) GetSSHKey(id string) (*SSHKey, error) {
	endpoint := fmt.Sprintf("%s/sshkeys/%s", c.baseURL, id)
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
	var key SSHKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, err
	}
	return &key, nil
}

// DeleteSSHKey deletes a single SSH key by ID.
func (c *Client) DeleteSSHKey(id string) error {
	endpoint := fmt.Sprintf("%s/sshkeys/%s", c.baseURL, id)
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return c.parseAPIError(resp.Body)
	}
	return nil
}
