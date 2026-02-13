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

package verda

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// StartupScriptService handles startup script API operations.
type StartupScriptService struct {
	client *Client
}

// CreateStartupScriptRequest represents a request to create a startup script.
type CreateStartupScriptRequest struct {
	Name   string `json:"name"`
	Script string `json:"script"`
}

// DeleteMultipleStartupScriptsRequest represents a request to delete multiple startup scripts.
type DeleteMultipleStartupScriptsRequest struct {
	Scripts []string `json:"scripts"`
}

func (s *StartupScriptService) GetAllStartupScripts(ctx context.Context) ([]StartupScript, error) {
	scripts, _, err := getRequest[[]StartupScript](ctx, s.client, "/scripts")
	if err != nil {
		return nil, err
	}
	return scripts, nil
}

func (s *StartupScriptService) GetStartupScriptByID(ctx context.Context, scriptID string) (*StartupScript, error) {
	path := fmt.Sprintf("/scripts/%s", scriptID)

	// API returns array even for single script lookup
	scripts, _, err := getRequest[[]StartupScript](ctx, s.client, path)
	if err != nil {
		return nil, err
	}

	if len(scripts) == 0 {
		return nil, fmt.Errorf("script not found: %s", scriptID)
	}

	return &scripts[0], nil
}

// AddStartupScript creates a script and refetches it since the API returns only the ID as plain text
func (s *StartupScriptService) AddStartupScript(ctx context.Context, req *CreateStartupScriptRequest) (*StartupScript, error) {
	return s.createWithPlainTextResponse(ctx, req)
}

// createWithPlainTextResponse handles API's inconsistent response format (sometimes JSON, sometimes plain text ID)
func (s *StartupScriptService) createWithPlainTextResponse(ctx context.Context, req *CreateStartupScriptRequest) (*StartupScript, error) {
	resp, err := s.client.makeRequest(ctx, http.MethodPost, "/scripts", req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    string(body),
			}
		}
		apiError.StatusCode = resp.StatusCode
		return nil, &apiError
	}

	// Try JSON first, fall back to plain text ID
	var script StartupScript
	if err := json.Unmarshal(body, &script); err == nil {
		return &script, nil
	}

	scriptID := strings.TrimSpace(string(body))
	scriptID = strings.Trim(scriptID, "\"")

	return s.GetStartupScriptByID(ctx, scriptID)
}

func (s *StartupScriptService) DeleteStartupScript(ctx context.Context, scriptID string) error {
	path := fmt.Sprintf("/scripts/%s", scriptID)
	_, err := deleteRequestNoResult(ctx, s.client, path)
	return err
}

func (s *StartupScriptService) DeleteMultipleStartupScripts(ctx context.Context, scriptIDs []string) error {
	req := &DeleteMultipleStartupScriptsRequest{
		Scripts: scriptIDs,
	}
	_, err := deleteRequestWithBody(ctx, s.client, "/scripts", req)
	return err
}
