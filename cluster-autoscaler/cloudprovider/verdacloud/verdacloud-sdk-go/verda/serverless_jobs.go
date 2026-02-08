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
	"fmt"
	"net/url"
)

// ServerlessJobsService handles serverless job API operations.
type ServerlessJobsService struct {
	client *Client
}

// GetJobDeployments retrieves all job deployments.
func (s *ServerlessJobsService) GetJobDeployments(ctx context.Context) ([]JobDeploymentShortInfo, error) {
	jobs, _, err := getRequest[[]JobDeploymentShortInfo](ctx, s.client, "/job-deployments")
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (s *ServerlessJobsService) CreateJobDeployment(ctx context.Context, req *CreateJobDeploymentRequest) (*JobDeployment, error) {
	// Validate required fields for create
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Compute == nil {
		return nil, fmt.Errorf("compute is required")
	}
	if req.Scaling == nil {
		return nil, fmt.Errorf("scaling is required")
	}
	if len(req.Containers) == 0 {
		return nil, fmt.Errorf("at least one container is required")
	}
	for i, c := range req.Containers {
		if c.Image == "" {
			return nil, fmt.Errorf("containers[%d].image is required", i)
		}
	}

	job, _, err := postRequest[JobDeployment](ctx, s.client, "/job-deployments", req)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *ServerlessJobsService) GetJobDeploymentByName(ctx context.Context, jobName string) (*JobDeployment, error) {
	path := fmt.Sprintf("/job-deployments/%s", jobName)
	job, _, err := getRequest[JobDeployment](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *ServerlessJobsService) UpdateJobDeployment(ctx context.Context, jobName string, req *UpdateJobDeploymentRequest) (*JobDeployment, error) {
	path := fmt.Sprintf("/job-deployments/%s", jobName)
	job, _, err := patchRequest[JobDeployment](ctx, s.client, path, req)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// DeleteJobDeployment removes a job with optional timeout in milliseconds
func (s *ServerlessJobsService) DeleteJobDeployment(ctx context.Context, jobName string, timeoutMs int) error {
	path := fmt.Sprintf("/job-deployments/%s", jobName)

	if timeoutMs > 0 {
		params := url.Values{}
		params.Set("timeout", fmt.Sprintf("%d", timeoutMs))
		path += "?" + params.Encode()
	}

	_, err := deleteRequestNoResult(ctx, s.client, path)
	return err
}

func (s *ServerlessJobsService) GetJobDeploymentScaling(ctx context.Context, jobName string) (*ContainerScalingOptions, error) {
	if jobName == "" {
		return nil, fmt.Errorf("jobName is required")
	}
	path := fmt.Sprintf("/job-deployments/%s/scaling", jobName)
	scaling, _, err := getRequest[ContainerScalingOptions](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &scaling, nil
}

func (s *ServerlessJobsService) UpdateJobDeploymentScaling(ctx context.Context, jobName string, req *UpdateScalingOptionsRequest) (*ContainerScalingOptions, error) {
	if jobName == "" {
		return nil, fmt.Errorf("jobName is required")
	}
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	path := fmt.Sprintf("/job-deployments/%s/scaling", jobName)
	scaling, _, err := patchRequest[ContainerScalingOptions](ctx, s.client, path, req)
	if err != nil {
		return nil, err
	}
	return &scaling, nil
}

func (s *ServerlessJobsService) PurgeJobDeploymentQueue(ctx context.Context, jobName string) error {
	path := fmt.Sprintf("/job-deployments/%s/purge-queue", jobName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, nil)
	return err
}

func (s *ServerlessJobsService) PauseJobDeployment(ctx context.Context, jobName string) error {
	path := fmt.Sprintf("/job-deployments/%s/pause", jobName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, nil)
	return err
}

func (s *ServerlessJobsService) ResumeJobDeployment(ctx context.Context, jobName string) error {
	path := fmt.Sprintf("/job-deployments/%s/resume", jobName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, nil)
	return err
}

func (s *ServerlessJobsService) GetJobDeploymentStatus(ctx context.Context, jobName string) (*JobDeploymentStatus, error) {
	path := fmt.Sprintf("/job-deployments/%s/status", jobName)
	status, _, err := getRequest[JobDeploymentStatus](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *ServerlessJobsService) GetJobEnvironmentVariables(ctx context.Context, jobName string) ([]ContainerEnvVar, error) {
	if jobName == "" {
		return nil, fmt.Errorf("jobName is required")
	}
	path := fmt.Sprintf("/job-deployments/%s/environment-variables", jobName)
	envVars, _, err := getRequest[[]ContainerEnvVar](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return envVars, nil
}

func (s *ServerlessJobsService) AddJobEnvironmentVariables(ctx context.Context, jobName string, req *EnvironmentVariablesRequest) error {
	if jobName == "" {
		return fmt.Errorf("jobName is required")
	}
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.ContainerName == "" {
		return fmt.Errorf("container_name is required")
	}
	if len(req.Env) == 0 {
		return fmt.Errorf("env array cannot be empty")
	}
	path := fmt.Sprintf("/job-deployments/%s/environment-variables", jobName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, req)
	return err
}

func (s *ServerlessJobsService) UpdateJobEnvironmentVariables(ctx context.Context, jobName string, req *EnvironmentVariablesRequest) error {
	if jobName == "" {
		return fmt.Errorf("jobName is required")
	}
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.ContainerName == "" {
		return fmt.Errorf("container_name is required")
	}
	if len(req.Env) == 0 {
		return fmt.Errorf("env array cannot be empty")
	}
	path := fmt.Sprintf("/job-deployments/%s/environment-variables", jobName)
	_, _, err := patchRequest[interface{}](ctx, s.client, path, req)
	return err
}

func (s *ServerlessJobsService) DeleteJobEnvironmentVariables(ctx context.Context, jobName string, req *DeleteEnvironmentVariablesRequest) error {
	if jobName == "" {
		return fmt.Errorf("jobName is required")
	}
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.ContainerName == "" {
		return fmt.Errorf("container_name is required")
	}
	if len(req.Env) == 0 {
		return fmt.Errorf("env array cannot be empty")
	}
	path := fmt.Sprintf("/job-deployments/%s/environment-variables", jobName)
	_, err := deleteRequestWithBody(ctx, s.client, path, req)
	return err
}
