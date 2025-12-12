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

// ContainerDeploymentsService handles container deployment API operations.
type ContainerDeploymentsService struct {
	client *Client
}

// GetDeployments retrieves all container deployments
// projectID is optional - if empty, uses default project
func (s *ContainerDeploymentsService) GetDeployments(ctx context.Context) ([]ContainerDeployment, error) {
	path := "/container-deployments"

	// Note: projectId query parameter may be required by some API environments
	// The API typically uses the default project from authentication context
	// If you need explicit project support, use GetDeploymentsForProject

	deployments, _, err := getRequest[[]ContainerDeployment](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

// GetDeploymentsForProject retrieves container deployments for a specific project
func (s *ContainerDeploymentsService) GetDeploymentsForProject(ctx context.Context, projectID string) ([]ContainerDeployment, error) {
	path := "/container-deployments"

	if projectID != "" {
		params := url.Values{}
		params.Set("projectId", projectID)
		path += "?" + params.Encode()
	}

	deployments, _, err := getRequest[[]ContainerDeployment](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

func (s *ContainerDeploymentsService) CreateDeployment(ctx context.Context, req *CreateDeploymentRequest) (*ContainerDeployment, error) {
	// Validate required fields for create
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Compute.Name == "" {
		return nil, fmt.Errorf("compute is required")
	}
	if req.Scaling.MaxReplicaCount == 0 {
		return nil, fmt.Errorf("scaling.max_replica_count is required")
	}
	if len(req.Containers) == 0 {
		return nil, fmt.Errorf("at least one container is required")
	}
	for i, c := range req.Containers {
		if c.Image == "" {
			return nil, fmt.Errorf("containers[%d].image is required", i)
		}
		if c.ExposedPort == 0 {
			return nil, fmt.Errorf("containers[%d].exposed_port is required", i)
		}
	}

	deployment, _, err := postRequest[ContainerDeployment](ctx, s.client, "/container-deployments", req)
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (s *ContainerDeploymentsService) GetDeploymentByName(ctx context.Context, deploymentName string) (*ContainerDeployment, error) {
	path := fmt.Sprintf("/container-deployments/%s", deploymentName)
	deployment, _, err := getRequest[ContainerDeployment](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (s *ContainerDeploymentsService) UpdateDeployment(ctx context.Context, deploymentName string, req *UpdateDeploymentRequest) (*ContainerDeployment, error) {
	// Validate required fields for update
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if deploymentName == "" {
		return nil, fmt.Errorf("deploymentName is required")
	}
	// Note: UpdateDeployment is a PATCH operation, so partial updates are allowed.
	// Containers are optional - you can update just scaling, compute, or other fields.

	path := fmt.Sprintf("/container-deployments/%s", deploymentName)
	deployment, _, err := patchRequest[ContainerDeployment](ctx, s.client, path, req)
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// DeleteDeployment removes a deployment with timeout in milliseconds (0-300000ms)
// If timeoutMs <= 0, uses the API default of 60000ms (60 seconds)
func (s *ContainerDeploymentsService) DeleteDeployment(ctx context.Context, deploymentName string, timeoutMs int) error {
	if deploymentName == "" {
		return fmt.Errorf("deploymentName is required")
	}

	path := fmt.Sprintf("/container-deployments/%s", deploymentName)

	// Use default timeout of 60000ms if not specified
	// Valid range: 0-300000ms
	timeout := timeoutMs
	if timeout <= 0 {
		timeout = 60000 // default 60 seconds
	} else if timeout > 300000 {
		timeout = 300000 // max 300 seconds
	}

	params := url.Values{}
	params.Set("timeout", fmt.Sprintf("%d", timeout))
	path += "?" + params.Encode()

	_, err := deleteRequestNoResult(ctx, s.client, path)
	return err
}

func (s *ContainerDeploymentsService) GetDeploymentStatus(ctx context.Context, deploymentName string) (*DeploymentStatus, error) {
	path := fmt.Sprintf("/container-deployments/%s/status", deploymentName)
	status, _, err := getRequest[DeploymentStatus](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *ContainerDeploymentsService) RestartDeployment(ctx context.Context, deploymentName string) error {
	path := fmt.Sprintf("/container-deployments/%s/restart", deploymentName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, nil)
	return err
}

func (s *ContainerDeploymentsService) PauseDeployment(ctx context.Context, deploymentName string) error {
	path := fmt.Sprintf("/container-deployments/%s/pause", deploymentName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, nil)
	return err
}

func (s *ContainerDeploymentsService) ResumeDeployment(ctx context.Context, deploymentName string) error {
	path := fmt.Sprintf("/container-deployments/%s/resume", deploymentName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, nil)
	return err
}

func (s *ContainerDeploymentsService) PurgeDeploymentQueue(ctx context.Context, deploymentName string) error {
	path := fmt.Sprintf("/container-deployments/%s/purge-queue", deploymentName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, nil)
	return err
}

func (s *ContainerDeploymentsService) GetDeploymentScaling(ctx context.Context, deploymentName string) (*ContainerScalingOptions, error) {
	if deploymentName == "" {
		return nil, fmt.Errorf("deploymentName is required")
	}
	path := fmt.Sprintf("/container-deployments/%s/scaling", deploymentName)
	scaling, _, err := getRequest[ContainerScalingOptions](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &scaling, nil
}

func (s *ContainerDeploymentsService) UpdateDeploymentScaling(ctx context.Context, deploymentName string, req *UpdateScalingOptionsRequest) (*ContainerScalingOptions, error) {
	if deploymentName == "" {
		return nil, fmt.Errorf("deploymentName is required")
	}
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	path := fmt.Sprintf("/container-deployments/%s/scaling", deploymentName)
	scaling, _, err := patchRequest[ContainerScalingOptions](ctx, s.client, path, req)
	if err != nil {
		return nil, err
	}
	return &scaling, nil
}

func (s *ContainerDeploymentsService) GetDeploymentReplicas(ctx context.Context, deploymentName string) (*DeploymentReplicas, error) {
	path := fmt.Sprintf("/container-deployments/%s/replicas", deploymentName)
	replicas, _, err := getRequest[DeploymentReplicas](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return &replicas, nil
}

func (s *ContainerDeploymentsService) GetEnvironmentVariables(ctx context.Context, deploymentName string) ([]ContainerEnvVar, error) {
	if deploymentName == "" {
		return nil, fmt.Errorf("deploymentName is required")
	}
	path := fmt.Sprintf("/container-deployments/%s/environment-variables", deploymentName)
	envVars, _, err := getRequest[[]ContainerEnvVar](ctx, s.client, path)
	if err != nil {
		return nil, err
	}
	return envVars, nil
}

func (s *ContainerDeploymentsService) AddEnvironmentVariables(ctx context.Context, deploymentName string, req *EnvironmentVariablesRequest) error {
	if deploymentName == "" {
		return fmt.Errorf("deploymentName is required")
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
	path := fmt.Sprintf("/container-deployments/%s/environment-variables", deploymentName)
	_, _, err := postRequest[interface{}](ctx, s.client, path, req)
	return err
}

func (s *ContainerDeploymentsService) UpdateEnvironmentVariables(ctx context.Context, deploymentName string, req *EnvironmentVariablesRequest) error {
	if deploymentName == "" {
		return fmt.Errorf("deploymentName is required")
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
	path := fmt.Sprintf("/container-deployments/%s/environment-variables", deploymentName)
	_, _, err := patchRequest[interface{}](ctx, s.client, path, req)
	return err
}

func (s *ContainerDeploymentsService) DeleteEnvironmentVariables(ctx context.Context, deploymentName string, req *DeleteEnvironmentVariablesRequest) error {
	if deploymentName == "" {
		return fmt.Errorf("deploymentName is required")
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
	path := fmt.Sprintf("/container-deployments/%s/environment-variables", deploymentName)
	_, err := deleteRequestWithBody(ctx, s.client, path, req)
	return err
}

func (s *ContainerDeploymentsService) GetServerlessComputeResources(ctx context.Context) ([]ComputeResource, error) {
	resources, _, err := getRequest[[]ComputeResource](ctx, s.client, "/serverless-compute-resources")
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (s *ContainerDeploymentsService) GetSecrets(ctx context.Context) ([]Secret, error) {
	secrets, _, err := getRequest[[]Secret](ctx, s.client, "/secrets")
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (s *ContainerDeploymentsService) CreateSecret(ctx context.Context, req *CreateSecretRequest) error {
	_, _, err := postRequest[interface{}](ctx, s.client, "/secrets", req)
	return err
}

// DeleteSecret removes a secret - force deletes even if in use (dangerous)
func (s *ContainerDeploymentsService) DeleteSecret(ctx context.Context, secretName string, force bool) error {
	path := fmt.Sprintf("/secrets/%s", secretName)

	if force {
		params := url.Values{}
		params.Set("force", "true")
		path += "?" + params.Encode()
	}

	_, err := deleteRequestNoResult(ctx, s.client, path)
	return err
}

func (s *ContainerDeploymentsService) GetFileSecrets(ctx context.Context) ([]FileSecret, error) {
	secrets, _, err := getRequest[[]FileSecret](ctx, s.client, "/file-secrets")
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (s *ContainerDeploymentsService) CreateFileSecret(ctx context.Context, req *CreateFileSecretRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Files) == 0 {
		return fmt.Errorf("files map cannot be empty")
	}
	_, _, err := postRequest[interface{}](ctx, s.client, "/file-secrets", req)
	return err
}

func (s *ContainerDeploymentsService) DeleteFileSecret(ctx context.Context, secretName string, force bool) error {
	path := fmt.Sprintf("/file-secrets/%s", secretName)

	if force {
		params := url.Values{}
		params.Set("force", "true")
		path += "?" + params.Encode()
	}

	_, err := deleteRequestNoResult(ctx, s.client, path)
	return err
}

func (s *ContainerDeploymentsService) GetRegistryCredentials(ctx context.Context) ([]RegistryCredentials, error) {
	credentials, _, err := getRequest[[]RegistryCredentials](ctx, s.client, "/container-registry-credentials")
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func (s *ContainerDeploymentsService) CreateRegistryCredentials(ctx context.Context, req *CreateRegistryCredentialsRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Type == "" {
		return fmt.Errorf("type is required")
	}
	_, _, err := postRequest[interface{}](ctx, s.client, "/container-registry-credentials", req)
	return err
}

func (s *ContainerDeploymentsService) DeleteRegistryCredentials(ctx context.Context, credentialsName string, force bool) error {
	path := fmt.Sprintf("/container-registry-credentials/%s", credentialsName)

	if force {
		params := url.Values{}
		params.Set("force", "true")
		path += "?" + params.Encode()
	}

	_, err := deleteRequestNoResult(ctx, s.client, path)
	return err
}
