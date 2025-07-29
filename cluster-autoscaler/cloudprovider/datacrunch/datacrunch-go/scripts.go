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
	"io"
	"net/http"
)

// StartupScript represents a startup script in DataCrunch.
type StartupScript struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Script string `json:"script"`
}

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
