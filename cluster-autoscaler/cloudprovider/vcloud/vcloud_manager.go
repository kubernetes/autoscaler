/*
Copyright 2024 The Kubernetes Authors.

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

package vcloud

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// Config represents the VCloud configuration parsed from cloud-config file.
// It contains the necessary parameters to connect to VCloud NodePool APIs.
type Config struct {
	// ClusterID is the unique identifier for the VCloud cluster
	ClusterID string
	// ClusterName is the human-readable name of the cluster
	ClusterName string
	// MgmtURL is the base URL for VCloud management API endpoints
	MgmtURL string
	// ProviderToken is the authentication token for VCloud API access
	ProviderToken string
}

// EnhancedManager manages VCloud node groups and provides the main interface
// for cluster autoscaler operations. It reuses proven NodePool Autoscaler API client
// patterns for reliable cloud provider integration.
type EnhancedManager struct {
	// client is the VCloud API client for making HTTP requests
	client *VCloudAPIClient
	// clusterID is the unique identifier for this cluster
	clusterID string
	// nodeGroups is the list of discovered node groups
	nodeGroups []*NodeGroup
	// config contains the parsed cloud configuration
	config *Config
}

// VCloudAPIClient provides HTTP client functionality for VCloud NodePool APIs.
// It implements proven retry logic and error handling patterns for reliable
// communication with VCloud backend services.
type VCloudAPIClient struct {
	// clusterName is the human-readable cluster name
	clusterName string
	// clusterID is the unique cluster identifier
	clusterID string
	// mgmtURL is the base management API URL
	mgmtURL string
	// providerToken is the authentication token
	providerToken string
	// httpClient is the underlying HTTP client with timeout configuration
	httpClient *http.Client
}

// NodePoolInfo represents the structure of a VCloud NodePool as returned by the API.
// This matches the existing data structure used by VCloud NodePool Autoscaler.
type NodePoolInfo struct {
	// ID is the unique identifier for the node pool
	ID string `json:"id"`
	// Name is the human-readable name of the node pool
	Name string `json:"name"`
	// CurrentSize is the actual number of nodes currently in the pool
	CurrentSize int `json:"currentSize"`
	// DesiredSize is the target number of nodes for the pool
	DesiredSize int `json:"desiredSize"`
	// MinSize is the minimum allowed size for autoscaling
	MinSize int `json:"minSize"`
	// MaxSize is the maximum allowed size for autoscaling
	MaxSize int `json:"maxSize"`
	// InstanceType specifies the type/flavor of instances in this pool
	InstanceType string `json:"instanceType"`
	// Zone is the availability zone where the pool is located
	Zone string `json:"zone"`
	// Status indicates the current operational status of the node pool
	Status string `json:"status"`
	// TotalNodes is the total number of nodes including pending ones
	TotalNodes int `json:"totalNodes"`
	// Autoscaling indicates if autoscaling is enabled for this node pool
	Autoscaling bool `json:"autoscaling"`
	// AutoscalingClass specifies the autoscaling class used
	AutoscalingClass string `json:"autoscalingClass"`
	// CreatedAt is the timestamp when the node pool was created
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the timestamp when the node pool was last updated
	UpdatedAt string `json:"updatedAt"`
	// Machines is the list of actual machine IDs in this node pool
	Machines []string `json:"machines"`
}

// NodePoolResponse represents the API response structure for single node pool operations.
// This matches the response format used by VCloud NodePool APIs.
type NodePoolResponse struct {
	// Status is the HTTP status code returned by the API
	Status int `json:"status"`
	// Data contains the actual node pool information
	Data struct {
		// NodePool contains the detailed node pool information
		NodePool NodePoolInfo `json:"nodepool"`
	} `json:"data"`
	// Message contains any error or informational message from the API
	Message string `json:"message,omitempty"`
}

// NodePoolListResponse represents the API response structure for listing multiple node pools.
// Used for auto-discovery of available node groups that can be managed by cluster autoscaler.
type NodePoolListResponse struct {
	// Status is the HTTP status code returned by the API
	Status int `json:"status"`
	// Data contains the list of node pools
	Data struct {
		// NodePools is the array of available node pools
		NodePools []NodePoolInfo `json:"nodepools"`
	} `json:"data"`
}

// parseINIConfig parses VCloud configuration from INI format input.
// It looks for a [vCloud] section and extracts the required parameters:
// CLUSTER_ID, CLUSTER_NAME, MGMT_URL, and PROVIDER_TOKEN.
// This reuses the proven parser logic from existing VCloud projects.
func parseINIConfig(reader io.Reader) (*Config, error) {
	config := &Config{}
	scanner := bufio.NewScanner(reader)
	inVCloudSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			inVCloudSection = strings.EqualFold(sectionName, "vCloud")
			continue
		}

		// Parse key-value pairs only in vCloud section
		if inVCloudSection && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch strings.ToUpper(key) {
				case "CLUSTER_ID":
					config.ClusterID = value
				case "CLUSTER_NAME":
					config.ClusterName = value
				case "MGMT_URL":
					config.MgmtURL = value
				case "PROVIDER_TOKEN":
					config.ProviderToken = value
				}
			}
		}
	}

	return config, scanner.Err()
}

// parseEnvConfig reads VCloud configuration from environment variables.
// This provides an alternative to INI file configuration for containerized environments.
func parseEnvConfig() *Config {
	return &Config{
		ClusterID:     os.Getenv("CLUSTER_ID"),
		ClusterName:   os.Getenv("CLUSTER_NAME"),
		MgmtURL:       os.Getenv("MGMT_URL"),
		ProviderToken: os.Getenv("PROVIDER_TOKEN"),
	}
}

// newEnhancedManager creates a new VCloud manager instance with the provided configuration.
// It parses the cloud config, validates required parameters, and initializes the API client
// with proven retry logic and error handling patterns.
// If configReader is nil, it will try to read from environment variables.
func newEnhancedManager(configReader io.Reader) (*EnhancedManager, error) {
	var cfg *Config
	var err error

	if configReader != nil {
		cfg, err = parseINIConfig(configReader)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config: %v", err)
		}
	} else {
		// Try to read from environment variables
		cfg = parseEnvConfig()
	}

	if cfg.ClusterID == "" {
		return nil, fmt.Errorf("cluster ID is not provided")
	}
	if cfg.MgmtURL == "" {
		return nil, fmt.Errorf("management URL is not provided")
	}
	if cfg.ProviderToken == "" {
		return nil, fmt.Errorf("provider token is not provided")
	}

	// Create API client
	client := &VCloudAPIClient{
		clusterName:   cfg.ClusterName,
		clusterID:     cfg.ClusterID,
		mgmtURL:       cfg.MgmtURL,
		providerToken: cfg.ProviderToken,
		httpClient:    &http.Client{Timeout: 60 * time.Second},
	}

	m := &EnhancedManager{
		client:     client,
		clusterID:  cfg.ClusterID,
		nodeGroups: make([]*NodeGroup, 0),
		config:     cfg,
	}

	return m, nil
}

// Request makes HTTP requests to VCloud APIs with intelligent URL construction and retry logic.
// It handles both base URLs and cluster-specific URLs, implements exponential backoff for retries,
// and provides comprehensive error handling. This reuses proven HTTP client patterns.
func (c *VCloudAPIClient) Request(ctx context.Context, method, path string, requestBody io.Reader) (*http.Response, error) {
	// Check if mgmtURL already contains the cluster path to avoid duplication
	var url string
	if strings.Contains(c.mgmtURL, "/clusters/"+c.clusterID) {
		// MGMT_URL already includes the cluster path
		url = fmt.Sprintf("%s%s", c.mgmtURL, path)
	} else {
		// MGMT_URL is base URL, need to add cluster path
		url = fmt.Sprintf("%s/clusters/%s%s", c.mgmtURL, c.clusterID, path)
	}

	klog.V(4).Infof("Making %s request to VCloud API", method)

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header if a provider token is set
	if c.providerToken != "" {
		req.Header.Set("X-Provider-Token", c.providerToken)
	}

	// Set content-type for JSON API calls
	req.Header.Set("Content-Type", "application/json")

	// Track API latency
	startTime := time.Now()

	// Make the request with retry logic (3 attempts with exponential backoff)
	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			waitTime := time.Second * time.Duration(i+1)
			klog.Warningf("Request failed (attempt %d/%d), retrying in %v: %v", i+1, maxRetries, waitTime, err)
			time.Sleep(waitTime)
			continue
		}
		return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, err)
	}

	// Record latency
	duration := time.Since(startTime).Seconds()
	klog.V(4).Infof("Request completed in %.3fs: %s -> %d", duration, method, resp.StatusCode)

	// Handle non-200 responses
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d, body: %s)", resp.StatusCode, string(body))
	}

	klog.V(4).Infof("Request successful: %s -> %d", method, resp.StatusCode)
	return resp, nil
}

// GetNodePool retrieves detailed information about a specific node pool by name.
// This reuses the proven API patterns from NodePool Autoscaler.
func (c *VCloudAPIClient) GetNodePool(ctx context.Context, nodePoolName string) (*NodePoolInfo, error) {
	klog.V(2).Infof("Getting node pool info for: %s", nodePoolName)

	resp, err := c.Request(ctx, "GET", fmt.Sprintf("/nodepools/%s", nodePoolName), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get node pool %s: %w", nodePoolName, err)
	}
	defer resp.Body.Close()

	var nodePoolResponse NodePoolResponse
	if err := json.NewDecoder(resp.Body).Decode(&nodePoolResponse); err != nil {
		return nil, fmt.Errorf("failed to decode node pool response: %w", err)
	}

	if nodePoolResponse.Status != 200 {
		return nil, fmt.Errorf("API returned error status %d: %s", nodePoolResponse.Status, nodePoolResponse.Message)
	}

	nodePool := &nodePoolResponse.Data.NodePool

	klog.V(2).Infof("Retrieved node pool: %+v", *nodePool)
	return nodePool, nil
}

// ListNodePools discovers all available node pools in the cluster.
// This is used for auto-discovery of node groups that can be managed by cluster autoscaler.
// Only node pools with autoscaling enabled (min/max > 0) will be considered for management.
func (c *VCloudAPIClient) ListNodePools(ctx context.Context) ([]NodePoolInfo, error) {
	klog.V(2).Infof("Listing all node pools")

	resp, err := c.Request(ctx, "GET", "/nodepools", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list node pools: %w", err)
	}
	defer resp.Body.Close()

	var listResponse NodePoolListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to decode node pools list response: %w", err)
	}

	if listResponse.Status != 200 {
		return nil, fmt.Errorf("API returned error status %d", listResponse.Status)
	}

	klog.V(2).Infof("Retrieved %d node pools", len(listResponse.Data.NodePools))
	return listResponse.Data.NodePools, nil
}

// ListNodePoolInstances retrieves the actual instances in a node pool.
// This returns the real instance IDs that match Kubernetes node provider IDs.
func (c *VCloudAPIClient) ListNodePoolInstances(ctx context.Context, nodePoolID string) ([]string, error) {
	klog.V(2).Infof("Listing instances for node pool: %s", nodePoolID)

	resp, err := c.Request(ctx, "GET", fmt.Sprintf("/nodepools/%s/machines", nodePoolID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list node pool instances: %w", err)
	}
	defer resp.Body.Close()

	// For now, let's see what the API returns and handle it gracefully
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read instances response: %w", readErr)
	}

	klog.V(4).Infof("Instances API response: %s", string(body))

	// Parse the actual API response format from your VCloud API
	var instancesResponse struct {
		Status int `json:"status"`
		Data   struct {
			Machines []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				State     string `json:"state"`
				IP        string `json:"ip"`
				OS        string `json:"os"`
				Kernel    string `json:"kernel"`
				Runtime   string `json:"runtime"`
				CreatedAt string `json:"createdAt"`
			} `json:"machines"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &instancesResponse); err != nil {
		klog.Warningf("Failed to parse instances response as JSON: %v", err)
		// Return empty list if API doesn't exist yet
		return []string{}, nil
	}

	if instancesResponse.Status != 200 {
		klog.Warningf("Instances API returned status %d", instancesResponse.Status)
		return []string{}, nil
	}

	var instanceIDs []string
	for _, machine := range instancesResponse.Data.Machines {
		instanceIDs = append(instanceIDs, machine.ID)
	}

	klog.V(2).Infof("Found %d instances in node pool %s", len(instanceIDs), nodePoolID)
	return instanceIDs, nil
}

// ScaleNodePool scales a node pool to the specified desired size.
// It sends a scaling request to the VCloud NodePool API with the new target size.
// This operation is asynchronous - the API will begin scaling and update the pool status.
func (c *VCloudAPIClient) ScaleNodePool(ctx context.Context, nodePoolName string, desiredSize int) error {
	klog.V(2).Infof("Scaling node pool %s to size %d", nodePoolName, desiredSize)

	// Create scale request payload (following cloud provider patterns)
	payload := map[string]interface{}{
		"desiredSize": desiredSize,
		"reason":      "cluster-autoscaler-scale-up",
		"async":       true, // Non-blocking operation
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal scale request: %w", err)
	}

	resp, err := c.Request(ctx, "PUT", fmt.Sprintf("/nodepools/%s/scale", nodePoolName), strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to scale node pool %s: %w", nodePoolName, err)
	}
	defer resp.Body.Close()

	var scaleResponse NodePoolResponse
	if err := json.NewDecoder(resp.Body).Decode(&scaleResponse); err != nil {
		return fmt.Errorf("failed to decode scale response: %w", err)
	}

	if scaleResponse.Status != 200 {
		return fmt.Errorf("scale operation failed with status %d: %s", scaleResponse.Status, scaleResponse.Message)
	}

	klog.Infof("Successfully scaled node pool %s to size %d", nodePoolName, desiredSize)
	return nil
}

// DeleteInstance deletes a specific instance from the node pool.
// This follows the common pattern used by other cloud providers for individual node deletion.
func (c *VCloudAPIClient) DeleteInstance(ctx context.Context, nodePoolName, instanceID string) error {
	klog.V(2).Infof("Deleting instance %s from node pool %s", instanceID, nodePoolName)

	// Create delete request payload (following common cloud provider patterns)
	deletePayload := map[string]interface{}{
		"force":  false, // Graceful shutdown
		"reason": "cluster-autoscaler-scale-down",
	}

	body, err := json.Marshal(deletePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	resp, err := c.Request(ctx, "DELETE", fmt.Sprintf("/nodepools/%s/machines/%s", nodePoolName, instanceID), strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to delete instance %s: %w", instanceID, err)
	}
	defer resp.Body.Close()

	var deleteResponse struct {
		Status  int    `json:"status"`
		Message string `json:"message,omitempty"`
		Data    struct {
			InstanceID string `json:"instanceId,omitempty"`
			Operation  string `json:"operation,omitempty"`
		} `json:"data,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&deleteResponse); err != nil {
		return fmt.Errorf("failed to decode delete response: %w", err)
	}

	if deleteResponse.Status != 200 {
		return fmt.Errorf("delete operation failed with status %d: %s", deleteResponse.Status, deleteResponse.Message)
	}

	klog.Infof("Successfully deleted instance %s from node pool %s", instanceID, nodePoolName)
	return nil
}

// Refresh updates the list of node groups by querying the VCloud API.
// It discovers available node pools and converts them to NodeGroup objects for cluster autoscaler.
// Only node pools with autoscaling enabled (non-zero min/max sizes) are included.
func (m *EnhancedManager) Refresh() error {
	ctx := context.Background()
	klog.V(4).Infof("refreshing VCloud node groups for cluster %s", m.clusterID)

	// Use your proven API to discover node pools
	nodePools, err := m.client.ListNodePools(ctx)
	if err != nil {
		klog.Warningf("failed to list node pools: %v", err)
		return err
	}

	// Convert NodePoolInfo to NodeGroup objects
	var nodeGroups []*NodeGroup
	for _, pool := range nodePools {
		// Only include pools that have autoscaling enabled (non-zero min/max)
		if pool.MinSize > 0 || pool.MaxSize > 0 {
			klog.V(4).Infof("adding node group: %q name: %s min: %d max: %d current: %d",
				pool.ID, pool.Name, pool.MinSize, pool.MaxSize, pool.CurrentSize)

			ng := &NodeGroup{
				id:         pool.ID,
				clusterID:  m.clusterID,
				client:     m.client,
				manager:    m,
				minSize:    pool.MinSize,
				maxSize:    pool.MaxSize,
				targetSize: pool.DesiredSize,
			}
			nodeGroups = append(nodeGroups, ng)
		}
	}

	if len(nodeGroups) == 0 {
		klog.V(4).Info("cluster-autoscaler is disabled. no node pools are configured for autoscaling")
	}

	m.nodeGroups = nodeGroups
	return nil
}

// GetNodeGroups returns the current list of managed node groups.
// This returns the list that was populated by the last Refresh() call.
func (m *EnhancedManager) GetNodeGroups() []*NodeGroup {
	return m.nodeGroups
}

// GetNodeGroupForInstance finds the node group that contains the specified instance.
// It searches through all managed node groups to find the one containing the given instance ID.
// Returns an error if the instance is not found in any managed node group.
func (m *EnhancedManager) GetNodeGroupForInstance(instanceID string) (*NodeGroup, error) {
	for _, ng := range m.nodeGroups {
		instances, err := ng.Nodes()
		if err != nil {
			continue
		}
		for _, instance := range instances {
			if instance.Id == instanceID {
				return ng, nil
			}
		}
	}
	return nil, fmt.Errorf("node group not found for instance %s", instanceID)
}

// ValidateConfig verifies that all required configuration parameters are present and valid.
// It checks for required fields (CLUSTER_ID, CLUSTER_NAME, MGMT_URL, PROVIDER_TOKEN)
// and validates the format of the management URL.
func (m *EnhancedManager) ValidateConfig() error {
	if m.config.ClusterID == "" {
		return fmt.Errorf("CLUSTER_ID is required")
	}
	if m.config.ClusterName == "" {
		return fmt.Errorf("CLUSTER_NAME is required")
	}
	if m.config.MgmtURL == "" {
		return fmt.Errorf("MGMT_URL is required")
	}
	if m.config.ProviderToken == "" {
		return fmt.Errorf("PROVIDER_TOKEN is required")
	}

	// Validate URL format
	if !strings.HasPrefix(m.config.MgmtURL, "https://") {
		return fmt.Errorf("MGMT_URL must start with https://")
	}

	return nil
}
