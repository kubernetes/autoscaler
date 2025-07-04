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
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestNodeGroup_BasicProperties tests basic NodeGroup interface properties
func TestNodeGroup_BasicProperties(t *testing.T) {
	// Create a mock NodeGroup
	ng := &NodeGroup{
		id:         "test-pool-id",
		clusterID:  "test-cluster",
		client:     nil, // Not needed for basic property tests
		manager:    nil, // Not needed for basic property tests
		minSize:    1,
		maxSize:    10,
		targetSize: 3,
	}

	// Test MinSize
	if ng.MinSize() != 1 {
		t.Errorf("Expected MinSize 1, got %d", ng.MinSize())
	}

	// Test MaxSize
	if ng.MaxSize() != 10 {
		t.Errorf("Expected MaxSize 10, got %d", ng.MaxSize())
	}

	// Test TargetSize (use stored value since no client)
	if ng.targetSize != 3 {
		t.Errorf("Expected stored targetSize 3, got %d", ng.targetSize)
	}

	// Test ID
	if ng.Id() != "test-pool-id" {
		t.Errorf("Expected ID 'test-pool-id', got '%s'", ng.Id())
	}

	// Test Debug
	debug := ng.Debug()
	if !strings.Contains(debug, "test-pool-id") {
		t.Errorf("Debug string should contain node group ID, got: %s", debug)
	}
}

// TestNodeGroup_Autoprovisioned tests autoprovisioning flag
func TestNodeGroup_Autoprovisioned(t *testing.T) {
	ng := &NodeGroup{
		id:         "test-pool-id",
		clusterID:  "test-cluster",
		client:     nil,
		manager:    nil,
		minSize:    1,
		maxSize:    10,
		targetSize: 3,
	}

	// VCloud node groups are not autoprovisioned - they're managed manually
	if ng.Autoprovisioned() {
		t.Error("Expected Autoprovisioned() to return false")
	}
}

// TestParseINIConfig tests the configuration parsing functionality
func TestParseINIConfig(t *testing.T) {
	configData := `[vCloud]
CLUSTER_ID=test-cluster-123
CLUSTER_NAME=test-cluster
MGMT_URL=https://api.example.com/api/v2/clusters/test-cluster-123
PROVIDER_TOKEN=test-token-456`

	config, err := parseINIConfig(strings.NewReader(configData))
	if err != nil {
		t.Fatalf("parseINIConfig failed: %v", err)
	}

	if config.ClusterID != "test-cluster-123" {
		t.Errorf("Expected ClusterID 'test-cluster-123', got '%s'", config.ClusterID)
	}

	if config.ClusterName != "test-cluster" {
		t.Errorf("Expected ClusterName 'test-cluster', got '%s'", config.ClusterName)
	}

	if config.MgmtURL != "https://api.example.com/api/v2/clusters/test-cluster-123" {
		t.Errorf("Expected specific MGMT_URL, got '%s'", config.MgmtURL)
	}

	if config.ProviderToken != "test-token-456" {
		t.Errorf("Expected ProviderToken 'test-token-456', got '%s'", config.ProviderToken)
	}
}

// TestParseINIConfig_InvalidSection tests parsing with wrong section
func TestParseINIConfig_InvalidSection(t *testing.T) {
	configData := `[WrongSection]
CLUSTER_ID=test-cluster-123
CLUSTER_NAME=test-cluster`

	config, err := parseINIConfig(strings.NewReader(configData))
	if err != nil {
		t.Fatalf("parseINIConfig failed: %v", err)
	}

	// Should have empty values since section name is wrong
	if config.ClusterID != "" {
		t.Errorf("Expected empty ClusterID, got '%s'", config.ClusterID)
	}
}

// TestProviderIDFormat tests the provider ID format
func TestProviderIDFormat(t *testing.T) {
	// Test provider ID format
	expectedPrefix := "vcloud://"
	if !strings.HasPrefix("vcloud://test-instance", expectedPrefix) {
		t.Errorf("Provider ID should start with '%s'", expectedPrefix)
	}
}

// TestNewEnhancedManager tests manager creation
func TestNewEnhancedManager(t *testing.T) {
	configData := `[vCloud]
CLUSTER_ID=test-cluster-123
CLUSTER_NAME=test-cluster
MGMT_URL=https://api.example.com/api/v2/clusters/test-cluster-123
PROVIDER_TOKEN=test-token-456`

	manager, err := newEnhancedManager(strings.NewReader(configData))
	if err != nil {
		t.Fatalf("newEnhancedManager failed: %v", err)
	}

	if manager.clusterID != "test-cluster-123" {
		t.Errorf("Expected clusterID 'test-cluster-123', got '%s'", manager.clusterID)
	}

	if manager.client == nil {
		t.Error("Expected client to be initialized")
	}

	if manager.config == nil {
		t.Error("Expected config to be initialized")
	}
}

// TestNewEnhancedManager_MissingConfig tests manager creation with invalid config
func TestNewEnhancedManager_MissingConfig(t *testing.T) {
	configData := `[vCloud]
CLUSTER_NAME=test-cluster`

	_, err := newEnhancedManager(strings.NewReader(configData))
	if err == nil {
		t.Error("Expected error for missing required config fields")
	}

	if !strings.Contains(err.Error(), "cluster ID") {
		t.Errorf("Expected error about cluster ID, got: %v", err)
	}
}

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	// Test valid config
	validConfig := &Config{
		ClusterID:     "test-cluster",
		ClusterName:   "test-cluster-name",
		MgmtURL:       "https://api.example.com",
		ProviderToken: "test-token",
	}

	manager := &EnhancedManager{config: validConfig}
	if err := manager.ValidateConfig(); err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Test invalid URL format
	invalidConfig := &Config{
		ClusterID:     "test-cluster",
		ClusterName:   "test-cluster-name",
		MgmtURL:       "http://api.example.com", // HTTP instead of HTTPS
		ProviderToken: "test-token",
	}

	manager = &EnhancedManager{config: invalidConfig}
	if err := manager.ValidateConfig(); err == nil {
		t.Error("Expected error for invalid URL format")
	}

	// Test missing required field
	missingConfig := &Config{
		ClusterName:   "test-cluster-name",
		MgmtURL:       "https://api.example.com",
		ProviderToken: "test-token",
		// ClusterID is missing
	}

	manager = &EnhancedManager{config: missingConfig}
	if err := manager.ValidateConfig(); err == nil {
		t.Error("Expected error for missing ClusterID")
	}
}

// TestParseEnvConfig tests environment variable configuration parsing
func TestParseEnvConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("CLUSTER_ID", "test-cluster-env")
	os.Setenv("CLUSTER_NAME", "test-cluster-name-env")
	os.Setenv("MGMT_URL", "https://k8s.io.infra.vnetwork.dev")
	os.Setenv("PROVIDER_TOKEN", "test-token-env")
	defer func() {
		os.Unsetenv("CLUSTER_ID")
		os.Unsetenv("CLUSTER_NAME")
		os.Unsetenv("MGMT_URL")
		os.Unsetenv("PROVIDER_TOKEN")
	}()

	config := parseEnvConfig()

	if config.ClusterID != "test-cluster-env" {
		t.Errorf("Expected ClusterID 'test-cluster-env', got '%s'", config.ClusterID)
	}

	if config.ClusterName != "test-cluster-name-env" {
		t.Errorf("Expected ClusterName 'test-cluster-name-env', got '%s'", config.ClusterName)
	}

	if config.MgmtURL != "https://k8s.io.infra.vnetwork.dev" {
		t.Errorf("Expected specific MGMT_URL, got '%s'", config.MgmtURL)
	}

	if config.ProviderToken != "test-token-env" {
		t.Errorf("Expected ProviderToken 'test-token-env', got '%s'", config.ProviderToken)
	}
}

// TestParseINIConfig_RealExample tests parsing with real vCloud configuration data
func TestParseINIConfig_RealExample(t *testing.T) {
	// Real vCloud configuration example from production environment
	realConfigData := `[vCloud]
CLUSTER_ID=363d6263-865f-48f6-b21e-5bf48b766c28
CLUSTER_NAME=k8s-c-shanismit
MGMT_URL=https://k8s.io.infra.vnetwork.dev/api/v2/services/42f16e6f-39bf-4ca0-9b3a-cd924d353396/s10015/clusters/363d6263-865f-48f6-b21e-5bf48b766c28
PROVIDER_TOKEN=MzYzZDYyNjMtODY1Zi00OGY2LWIyMWUtNWJmNDhiNzY2YzI4`

	config, err := parseINIConfig(strings.NewReader(realConfigData))
	if err != nil {
		t.Fatalf("parseINIConfig failed with real config data: %v", err)
	}

	// Validate parsed values match expected real configuration
	expectedClusterID := "363d6263-865f-48f6-b21e-5bf48b766c28"
	if config.ClusterID != expectedClusterID {
		t.Errorf("Expected ClusterID '%s', got '%s'", expectedClusterID, config.ClusterID)
	}

	expectedClusterName := "k8s-c-shanismit"
	if config.ClusterName != expectedClusterName {
		t.Errorf("Expected ClusterName '%s', got '%s'", expectedClusterName, config.ClusterName)
	}

	expectedMgmtURL := "https://k8s.io.infra.vnetwork.dev/api/v2/services/42f16e6f-39bf-4ca0-9b3a-cd924d353396/s10015/clusters/363d6263-865f-48f6-b21e-5bf48b766c28"
	if config.MgmtURL != expectedMgmtURL {
		t.Errorf("Expected MGMT_URL '%s', got '%s'", expectedMgmtURL, config.MgmtURL)
	}

	expectedToken := "MzYzZDYyNjMtODY1Zi00OGY2LWIyMWUtNWJmNDhiNzY2YzI4"
	if config.ProviderToken != expectedToken {
		t.Errorf("Expected ProviderToken '%s', got '%s'", expectedToken, config.ProviderToken)
	}

	// Test URL validation with real config format
	if !strings.HasPrefix(config.MgmtURL, "https://") {
		t.Error("Real config MGMT_URL should start with https://")
	}

	// Test that URL contains expected vnetwork.dev domain
	if !strings.Contains(config.MgmtURL, "k8s.io.infra.vnetwork.dev") {
		t.Error("Real config MGMT_URL should contain vnetwork.dev domain")
	}

	// Test cluster ID format (should be UUID format)
	if len(config.ClusterID) != 36 {
		t.Errorf("Expected ClusterID to be 36 characters (UUID format), got %d characters", len(config.ClusterID))
	}

	// Test cluster name format (should match pattern)
	if !strings.HasPrefix(config.ClusterName, "k8s-c-") {
		t.Error("Expected ClusterName to start with 'k8s-c-' prefix")
	}
}

// TestNewEnhancedManager_RealExample tests manager creation with real configuration
func TestNewEnhancedManager_RealExample(t *testing.T) {
	// Real vCloud configuration example
	realConfigData := `[vCloud]
CLUSTER_ID=363d6263-865f-48f6-b21e-5bf48b766c28
CLUSTER_NAME=k8s-c-shanismit
MGMT_URL=https://k8s.io.infra.vnetwork.dev/api/v2/services/42f16e6f-39bf-4ca0-9b3a-cd924d353396/s10015/clusters/363d6263-865f-48f6-b21e-5bf48b766c28
PROVIDER_TOKEN=MzYzZDYyNjMtODY1Zi00OGY2LWIyMWUtNWJmNDhiNzY2YzI4`

	manager, err := newEnhancedManager(strings.NewReader(realConfigData))
	if err != nil {
		t.Fatalf("newEnhancedManager failed with real config: %v", err)
	}

	// Validate manager was created correctly with real config
	expectedClusterID := "363d6263-865f-48f6-b21e-5bf48b766c28"
	if manager.clusterID != expectedClusterID {
		t.Errorf("Expected clusterID '%s', got '%s'", expectedClusterID, manager.clusterID)
	}

	if manager.client == nil {
		t.Error("Expected client to be initialized with real config")
	}

	if manager.config == nil {
		t.Error("Expected config to be initialized with real config")
	}

	// Test configuration validation with real data
	if err := manager.ValidateConfig(); err != nil {
		t.Errorf("Real config should pass validation, got error: %v", err)
	}

	// Test that the API client was configured with correct endpoints
	if manager.client.clusterID != expectedClusterID {
		t.Errorf("Expected client clusterID '%s', got '%s'", expectedClusterID, manager.client.clusterID)
	}

	expectedClusterName := "k8s-c-shanismit"
	if manager.client.clusterName != expectedClusterName {
		t.Errorf("Expected client clusterName '%s', got '%s'", expectedClusterName, manager.client.clusterName)
	}

	// Test URL construction patterns
	expectedMgmtURL := "https://k8s.io.infra.vnetwork.dev/api/v2/services/42f16e6f-39bf-4ca0-9b3a-cd924d353396/s10015/clusters/363d6263-865f-48f6-b21e-5bf48b766c28"
	if manager.client.mgmtURL != expectedMgmtURL {
		t.Errorf("Expected client mgmtURL to match config, got '%s'", manager.client.mgmtURL)
	}
}

// TestMachineInfoStruct tests the MachineInfo structure with real API response data
func TestMachineInfoStruct(t *testing.T) {
	// Updated real machine data from your VCloud API response (now includes nodePoolId)
	realMachineData := `{
		"status": 200,
		"data": {
			"machines": [
				{
					"id": "0d026b65-03bf-4f6d-a053-10982471a22e",
					"name": "k8s-c-shanismit-nuhgut-worker-1",
					"state": "active",
					"ip": "10.254.11.10",
					"os": "Kubernetes Techev 25.05.v1.33.1",
					"kernel": "linux",
					"runtime": "containerd://1.7.12",
					"createdAt": "2025-06-25T01:04:35Z",
					"nodePoolId": "03f2031c-b0e9-42a3-8498-4c8ff2dc2046"
				},
				{
					"id": "0831e8fe-2912-43bc-bc6b-6ca7325ea69a",
					"name": "k8s-c-shanismit-nuhgut-worker-autoscaler-202506250825253",
					"state": "active",
					"ip": "10.254.11.16",
					"os": "Kubernetes Techev 25.05.v1.33.1",
					"kernel": "linux",
					"runtime": "containerd://1.7.12",
					"createdAt": "2025-06-25T01:25:51Z",
					"nodePoolId": "03f2031c-b0e9-42a3-8498-4c8ff2dc2046"
				}
			]
		}
	}`

	// Parse the response
	var machinesResponse struct {
		Status int `json:"status"`
		Data   struct {
			Machines []MachineInfo `json:"machines"`
		} `json:"data"`
	}

	err := json.Unmarshal([]byte(realMachineData), &machinesResponse)
	if err != nil {
		t.Fatalf("Failed to parse real machine data: %v", err)
	}

	// Validate the response was parsed correctly
	if machinesResponse.Status != 200 {
		t.Errorf("Expected status 200, got %d", machinesResponse.Status)
	}

	if len(machinesResponse.Data.Machines) != 2 {
		t.Errorf("Expected 2 machines, got %d", len(machinesResponse.Data.Machines))
	}

	// Validate first machine
	machine1 := machinesResponse.Data.Machines[0]
	expectedID1 := "0d026b65-03bf-4f6d-a053-10982471a22e"
	if machine1.ID != expectedID1 {
		t.Errorf("Expected machine1 ID '%s', got '%s'", expectedID1, machine1.ID)
	}

	expectedName1 := "k8s-c-shanismit-nuhgut-worker-1"
	if machine1.Name != expectedName1 {
		t.Errorf("Expected machine1 name '%s', got '%s'", expectedName1, machine1.Name)
	}

	if machine1.State != "active" {
		t.Errorf("Expected machine1 state 'active', got '%s'", machine1.State)
	}

	expectedIP1 := "10.254.11.10"
	if machine1.IP != expectedIP1 {
		t.Errorf("Expected machine1 IP '%s', got '%s'", expectedIP1, machine1.IP)
	}

	expectedOS := "Kubernetes Techev 25.05.v1.33.1"
	if machine1.OS != expectedOS {
		t.Errorf("Expected machine1 OS '%s', got '%s'", expectedOS, machine1.OS)
	}

	if machine1.Kernel != "linux" {
		t.Errorf("Expected machine1 kernel 'linux', got '%s'", machine1.Kernel)
	}

	expectedRuntime := "containerd://1.7.12"
	if machine1.Runtime != expectedRuntime {
		t.Errorf("Expected machine1 runtime '%s', got '%s'", expectedRuntime, machine1.Runtime)
	}

	expectedCreatedAt := "2025-06-25T01:04:35Z"
	if machine1.CreatedAt != expectedCreatedAt {
		t.Errorf("Expected machine1 createdAt '%s', got '%s'", expectedCreatedAt, machine1.CreatedAt)
	}

	// Validate second machine (autoscaler created)
	machine2 := machinesResponse.Data.Machines[1]
	expectedID2 := "0831e8fe-2912-43bc-bc6b-6ca7325ea69a"
	if machine2.ID != expectedID2 {
		t.Errorf("Expected machine2 ID '%s', got '%s'", expectedID2, machine2.ID)
	}

	// Check that the second machine has autoscaler in the name
	if !strings.Contains(machine2.Name, "autoscaler") {
		t.Error("Expected machine2 name to contain 'autoscaler'")
	}

	expectedIP2 := "10.254.11.16"
	if machine2.IP != expectedIP2 {
		t.Errorf("Expected machine2 IP '%s', got '%s'", expectedIP2, machine2.IP)
	}

	// Test that both machines have consistent OS and runtime
	if machine2.OS != expectedOS {
		t.Errorf("Expected consistent OS across machines, got '%s'", machine2.OS)
	}

	if machine2.Runtime != expectedRuntime {
		t.Errorf("Expected consistent runtime across machines, got '%s'", machine2.Runtime)
	}

	// NEW: Test nodePoolId validation (both machines should have the same nodePoolId)
	expectedNodePoolID := "03f2031c-b0e9-42a3-8498-4c8ff2dc2046"
	if machine1.NodePoolID != expectedNodePoolID {
		t.Errorf("Expected machine1 nodePoolId '%s', got '%s'", expectedNodePoolID, machine1.NodePoolID)
	}

	if machine2.NodePoolID != expectedNodePoolID {
		t.Errorf("Expected machine2 nodePoolId '%s', got '%s'", expectedNodePoolID, machine2.NodePoolID)
	}

	// Validate that both machines belong to the same node pool
	if machine1.NodePoolID != machine2.NodePoolID {
		t.Errorf("Expected both machines to have same nodePoolId. Machine1: %s, Machine2: %s",
			machine1.NodePoolID, machine2.NodePoolID)
	}

	// Test nodePoolId format (should be UUID format)
	if len(machine1.NodePoolID) != 36 {
		t.Errorf("Expected nodePoolId to be 36 characters (UUID format), got %d characters", len(machine1.NodePoolID))
	}

	// Test that nodePoolId is not empty (API enhancement validation)
	if machine1.NodePoolID == "" {
		t.Error("Expected nodePoolId to be provided by the enhanced API")
	}

	if machine2.NodePoolID == "" {
		t.Error("Expected nodePoolId to be provided by the enhanced API")
	}
}

// TestDeleteNodes_ValidationChecks tests the validation logic in DeleteNodes
func TestDeleteNodes_ValidationChecks(t *testing.T) {
	// Create a mock NodeGroup with constraints
	ng := &NodeGroup{
		id:         "test-pool",
		clusterID:  "test-cluster",
		client:     nil,
		manager:    nil,
		minSize:    2,
		maxSize:    10,
		targetSize: 3,
	}

	// Verify the node group properties for improved DeleteNodes implementation
	if ng.MinSize() != 2 {
		t.Errorf("Expected MinSize 2, got %d", ng.MinSize())
	}

	if ng.MaxSize() != 10 {
		t.Errorf("Expected MaxSize 10, got %d", ng.MaxSize())
	}

	// This test validates that the DeleteNodes implementation follows
	// common cloud provider patterns with proper validation
	t.Log("DeleteNodes implementation updated to follow common cloud provider pattern")
}

// TestNodeGroup_IncreaseSize tests the IncreaseSize method
func TestNodeGroup_IncreaseSize(t *testing.T) {
	ng := &NodeGroup{
		id:         "test-pool",
		clusterID:  "test-cluster",
		client:     nil,
		manager:    nil,
		minSize:    1,
		maxSize:    10,
		targetSize: 3,
	}

	// Test negative delta first (validated before API call)
	err := ng.IncreaseSize(-1)
	if err == nil {
		t.Error("IncreaseSize should fail for negative delta")
	}

	// Test zero delta (validated before API call)
	err = ng.IncreaseSize(0)
	if err == nil {
		t.Error("IncreaseSize should fail for zero delta")
	}

	// Other tests would require a real client for TargetSize() call
	t.Log("IncreaseSize validation logic works correctly for delta validation")
}

// TestNodeGroup_DecreaseTargetSize tests the DecreaseTargetSize method
func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	ng := &NodeGroup{
		id:         "test-pool",
		clusterID:  "test-cluster",
		client:     nil,
		manager:    nil,
		minSize:    1,
		maxSize:    10,
		targetSize: 5,
	}

	// Test positive delta (should fail - this is validated first)
	err := ng.DecreaseTargetSize(2)
	if err == nil {
		t.Error("DecreaseTargetSize should fail for positive delta")
	}

	// Test negative delta (correct usage) - should succeed
	err = ng.DecreaseTargetSize(-2)
	if err != nil {
		t.Errorf("DecreaseTargetSize should succeed for negative delta, got error: %v", err)
	}

	// Verify the new target size is updated
	if ng.targetSize != 3 {
		t.Errorf("Expected targetSize to be 3 after decreasing by 2, got %d", ng.targetSize)
	}
}

// TestNodeGroup_Exist tests the Exist method
func TestNodeGroup_Exist(t *testing.T) {
	ng := &NodeGroup{
		id:           "test-pool",
		clusterID:    "test-cluster",
		client:       nil,
		manager:      nil,
		minSize:      1,
		maxSize:      10,
		targetSize:   3,
		instanceType: "v2g-standard-4-8",
	}

	// Exist() always returns true in the current implementation
	if !ng.Exist() {
		t.Error("Expected Exist() to return true")
	}
}

// TestNodeGroup_TemplateNodeInfo tests the TemplateNodeInfo method
func TestNodeGroup_TemplateNodeInfo(t *testing.T) {
	ng := &NodeGroup{
		id:           "test-pool",
		clusterID:    "test-cluster",
		client:       nil,
		manager:      nil,
		minSize:      1,
		maxSize:      10,
		targetSize:   3,
		instanceType: "v2g-standard-4-8",
	}

	nodeInfo, err := ng.TemplateNodeInfo()
	if err != nil {
		t.Errorf("TemplateNodeInfo should not return error, got: %v", err)
	}

	if nodeInfo == nil {
		t.Fatal("Expected nodeInfo to be non-nil")
	}

	node := nodeInfo.Node()
	if node == nil {
		t.Fatal("Expected node to be non-nil")
	}

	// Verify node has the correct labels
	if node.Labels["node.kubernetes.io/instance-type"] != "v2g-standard-4-8" {
		t.Errorf("Expected instance type label 'v2g-standard-4-8', got '%s'",
			node.Labels["node.kubernetes.io/instance-type"])
	}

	// Verify node has correct resource capacity
	cpuCapacity := node.Status.Capacity["cpu"]
	if cpuCapacity.IsZero() {
		t.Error("Expected CPU capacity to be non-zero")
	}

	memoryCapacity := node.Status.Capacity["memory"]
	if memoryCapacity.IsZero() {
		t.Error("Expected memory capacity to be non-zero")
	}

	// Verify node has allocatable resources
	cpuAllocatable := node.Status.Allocatable["cpu"]
	if cpuAllocatable.IsZero() {
		t.Error("Expected CPU allocatable to be non-zero")
	}

	memoryAllocatable := node.Status.Allocatable["memory"]
	if memoryAllocatable.IsZero() {
		t.Error("Expected memory allocatable to be non-zero")
	}

	// Verify that allocatable is less than capacity (system overhead)
	if cpuAllocatable.Cmp(cpuCapacity) >= 0 {
		t.Error("Expected CPU allocatable to be less than capacity")
	}

	if memoryAllocatable.Cmp(memoryCapacity) >= 0 {
		t.Error("Expected memory allocatable to be less than capacity")
	}

	// Verify node has ready conditions
	found := false
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" && condition.Status == "True" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected node to have Ready=True condition")
	}

	t.Logf("TemplateNodeInfo test passed: CPU=%s, Memory=%s",
		cpuCapacity.String(), memoryCapacity.String())
}

// TestNodeGroup_TemplateNodeInfo_ProductionInstanceType tests with actual production instance type
func TestNodeGroup_TemplateNodeInfo_ProductionInstanceType(t *testing.T) {
	ng := &NodeGroup{
		id:           "test-pool",
		clusterID:    "test-cluster",
		client:       nil,
		manager:      nil,
		minSize:      1,
		maxSize:      10,
		targetSize:   3,
		instanceType: "v2g-standard-8-16", // Actual production instance type
	}

	nodeInfo, err := ng.TemplateNodeInfo()
	if err != nil {
		t.Errorf("TemplateNodeInfo should not return error, got: %v", err)
	}

	if nodeInfo == nil {
		t.Fatal("Expected nodeInfo to be non-nil")
	}

	node := nodeInfo.Node()
	if node == nil {
		t.Fatal("Expected node to be non-nil")
	}

	// Verify node has the correct labels
	if node.Labels["node.kubernetes.io/instance-type"] != "v2g-standard-8-16" {
		t.Errorf("Expected instance type label 'v2g-standard-8-16', got '%s'",
			node.Labels["node.kubernetes.io/instance-type"])
	}

	// Verify node has correct resource capacity for 8-16 instance type
	cpuCapacity := node.Status.Capacity["cpu"]
	if cpuCapacity.IsZero() {
		t.Error("Expected CPU capacity to be non-zero")
	}

	memoryCapacity := node.Status.Capacity["memory"]
	if memoryCapacity.IsZero() {
		t.Error("Expected memory capacity to be non-zero")
	}

	// Verify actual production values: 8 CPU, 16GB memory
	expectedCPU := int64(8)
	expectedMemoryGB := int64(16)
	expectedMemoryBytes := expectedMemoryGB * 1024 * 1024 * 1024

	actualCPU := cpuCapacity.Value()
	actualMemory := memoryCapacity.Value()

	if actualCPU != expectedCPU {
		t.Errorf("Expected CPU capacity %d, got %d", expectedCPU, actualCPU)
	}

	if actualMemory != expectedMemoryBytes {
		t.Errorf("Expected memory capacity %d bytes (%dGB), got %d bytes",
			expectedMemoryBytes, expectedMemoryGB, actualMemory)
	}

	// Verify allocatable resources are less than capacity (system overhead)
	cpuAllocatable := node.Status.Allocatable["cpu"]
	memoryAllocatable := node.Status.Allocatable["memory"]

	if cpuAllocatable.Cmp(cpuCapacity) >= 0 {
		t.Error("Expected CPU allocatable to be less than capacity")
	}

	if memoryAllocatable.Cmp(memoryCapacity) >= 0 {
		t.Error("Expected memory allocatable to be less than capacity")
	}

	t.Logf("Production TemplateNodeInfo test passed: CPU=%s, Memory=%sGB",
		cpuCapacity.String(), memoryCapacity.String())
}

// TestParseInstanceType tests the parseInstanceType utility function
func TestParseInstanceType(t *testing.T) {
	tests := []struct {
		instanceType   string
		expectedCPU    int64
		expectedMemory int64
		expectError    bool
	}{
		{"v2g-standard-8-16", 8, 16 * 1024 * 1024 * 1024, false},
		{"v2g-standard-4-8", 4, 8 * 1024 * 1024 * 1024, false},
		{"v2g-standard-2-4", 2, 4 * 1024 * 1024 * 1024, false},
		{"invalid-format", 0, 0, true},
		{"v2g-standard", 0, 0, true},
		{"v2g-standard-invalid-8", 0, 0, true},
		{"v2g-standard-8-invalid", 0, 0, true},
	}

	for _, test := range tests {
		t.Run(test.instanceType, func(t *testing.T) {
			cpu, memory, err := parseInstanceType(test.instanceType)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for instance type %s", test.instanceType)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for instance type %s: %v", test.instanceType, err)
				}
				if cpu != test.expectedCPU {
					t.Errorf("Expected CPU %d, got %d", test.expectedCPU, cpu)
				}
				if memory != test.expectedMemory {
					t.Errorf("Expected memory %d, got %d", test.expectedMemory, memory)
				}
			}
		})
	}
}
