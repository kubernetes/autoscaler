package datacrunchclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// UploadStartupScript uploads a startup script to DataCrunch and returns its script ID.
func (c *Client) UploadStartupScript(name string, script string) (string, error) {
	type addScriptRequest struct {
		Name   string `json:"name"`
		Script string `json:"script"`
	}
	req := addScriptRequest{
		Name:   name,
		Script: script,
	}
	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	endpoint := c.baseURL + "/scripts"
	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.doRequest(httpReq)
	if err != nil {
		return "", err
	}

	scriptIDBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(scriptIDBytes), nil
}

// DeleteStartupScript deletes a startup script by its ID.
func (c *Client) DeleteStartupScript(scriptID string) error {
	type deleteScriptsRequest struct {
		Scripts []string `json:"scripts"`
	}
	body, err := json.Marshal(deleteScriptsRequest{Scripts: []string{scriptID}})
	if err != nil {
		return err
	}
	endpoint := c.baseURL + "/scripts"
	req, err := http.NewRequest("DELETE", endpoint, bytes.NewReader(body))
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

// ListStartupScripts returns all startup scripts for the project.
func (c *Client) ListStartupScripts() ([]StartupScript, error) {
	endpoint := c.baseURL + "/scripts"
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
	var scripts []StartupScript
	if err := json.NewDecoder(resp.Body).Decode(&scripts); err != nil {
		return nil, err
	}
	return scripts, nil
}
