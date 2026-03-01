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
	"net/url"
	"strings"
)

// InstanceService handles instance-related API operations.
type InstanceService struct {
	client *Client
}

// Get retrieves all instances, optionally filtered by status.
func (s *InstanceService) Get(ctx context.Context, status string) ([]Instance, error) {
	path := "/instances"
	if status != "" {
		params := url.Values{}
		params.Set("status", status)
		path += "?" + params.Encode()
	}

	instances, _, err := getRequest[[]Instance](ctx, s.client, path)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (s *InstanceService) GetByID(ctx context.Context, id string) (*Instance, error) {
	path := fmt.Sprintf("/instances/%s", id)

	instance, _, err := getRequest[Instance](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

func (s *InstanceService) Create(ctx context.Context, req CreateInstanceRequest) (*Instance, error) {
	if req.LocationCode == "" {
		req.LocationCode = LocationFIN01
	}

	if req.SSHKeyIDs == nil {
		req.SSHKeyIDs = []string{}
	}

	return s.createWithPlainTextResponse(ctx, req)
}

// createWithPlainTextResponse handles API's inconsistent response format (sometimes JSON, sometimes plain text ID)
func (s *InstanceService) createWithPlainTextResponse(ctx context.Context, req CreateInstanceRequest) (*Instance, error) {
	resp, err := s.client.makeRequest(ctx, http.MethodPost, "/instances", req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		s.client.Logger.Debug("Instance creation failed with status %d: %s", resp.StatusCode, string(body))

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var instance Instance
	if err := json.Unmarshal(body, &instance); err != nil {
		// Fall back to plain text ID
		instanceID := strings.TrimSpace(string(body))
		return s.GetByID(ctx, instanceID)
	}

	return &instance, nil
}

func (s *InstanceService) Action(ctx context.Context, ids []string, action string, volumeIDs []string) error {
	req := InstanceActionRequest{
		Action:    action,
		ID:        ids,
		VolumeIDs: volumeIDs,
	}

	_, _, err := putRequest[any](ctx, s.client, "/instances", req)
	return err
}

func (s *InstanceService) GetLocationAvailabilities(ctx context.Context) ([]LocationAvailability, error) {
	availabilities, _, err := getRequest[[]LocationAvailability](ctx, s.client, "/instance-availability")
	if err != nil {
		return nil, err
	}

	return availabilities, nil
}

func (s *InstanceService) CheckInstanceTypeAvailability(ctx context.Context, instanceType string) (bool, error) {
	path := fmt.Sprintf("/instance-availability/%s", instanceType)

	// API returns "true"/"false" as JSON string, not boolean
	resp, err := s.client.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return false, &APIError{
				StatusCode: resp.StatusCode,
				Message:    string(body),
			}
		}
		apiError.StatusCode = resp.StatusCode
		return false, &apiError
	}

	var boolResult bool
	if err := json.Unmarshal(body, &boolResult); err == nil {
		return boolResult, nil
	}

	var stringResult string
	if err := json.Unmarshal(body, &stringResult); err == nil {
		return stringResult == trueString, nil
	}

	return false, fmt.Errorf("unexpected response format: %s", string(body))
}

func (s *InstanceService) Boot(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionBoot, nil)
}

func (s *InstanceService) Start(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionStart, nil)
}

func (s *InstanceService) Shutdown(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionShutdown, nil)
}

func (s *InstanceService) Delete(ctx context.Context, volumeIDs []string, ids ...string) error {
	return s.Action(ctx, ids, ActionDelete, volumeIDs)
}

func (s *InstanceService) Discontinue(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionDiscontinue, nil)
}

// Hibernate shuts down and archives an instance - must be shut down first or API will error.
// Volumes are detached and the instance is deleted during hibernation.
func (s *InstanceService) Hibernate(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionHibernate, nil)
}

func (s *InstanceService) ConfigureSpot(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionConfigureSpot, nil)
}

func (s *InstanceService) ForceShutdown(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionForceShutdown, nil)
}

func (s *InstanceService) DeleteStuck(ctx context.Context, volumeIDs []string, ids ...string) error {
	return s.Action(ctx, ids, ActionDeleteStuck, volumeIDs)
}

func (s *InstanceService) Deploy(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionDeploy, nil)
}

func (s *InstanceService) Transfer(ctx context.Context, ids ...string) error {
	return s.Action(ctx, ids, ActionTransfer, nil)
}
