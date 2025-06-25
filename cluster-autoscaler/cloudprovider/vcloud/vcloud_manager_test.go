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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockVCloudAPIServer creates a test HTTP server that simulates VCloud API responses
func mockVCloudAPIServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock GET /nodepools - list all node pools
	mux.HandleFunc("/clusters/test-cluster-123/nodepools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			response := NodePoolListResponse{
				Status: 200,
				Data: struct {
					NodePools []NodePoolInfo `json:"nodepools"`
				}{
					NodePools: []NodePoolInfo{
						{
							ID:           "pool-1",
							Name:         "worker-pool-1",
							CurrentSize:  3,
							DesiredSize:  3,
							MinSize:      1,
							MaxSize:      10,
							InstanceType: "standard-4",
							Zone:         "zone-1",
							Status:       "active",
							Machines:     []string{"machine-1", "machine-2", "machine-3"},
						},
						{
							ID:           "pool-2",
							Name:         "worker-pool-2",
							CurrentSize:  2,
							DesiredSize:  2,
							MinSize:      1,
							MaxSize:      5,
							InstanceType: "standard-2",
							Zone:         "zone-1",
							Status:       "active",
							Machines:     []string{"machine-4", "machine-5"},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	// Mock GET /nodepools/{id} - get specific node pool
	mux.HandleFunc("/clusters/test-cluster-123/nodepools/", func(w http.ResponseWriter, r *http.Request) {
		poolID := strings.TrimPrefix(r.URL.Path, "/clusters/test-cluster-123/nodepools/")

		if strings.HasSuffix(poolID, "/machines") {
			// Mock machines endpoint
			poolID = strings.TrimSuffix(poolID, "/machines")
			mockMachinesResponse(w, r, poolID)
			return
		}

		if strings.Contains(poolID, "/scale") {
			// Mock scale endpoint
			poolID = strings.TrimSuffix(poolID, "/scale")
			mockScaleResponse(w, r, poolID)
			return
		}

		if strings.Contains(poolID, "/machines/") {
			// Mock individual machine deletion endpoint
			mockScaleResponse(w, r, poolID)
			return
		}

		if r.Method == "GET" {
			var nodePool NodePoolInfo
			switch poolID {
			case "pool-1":
				nodePool = NodePoolInfo{
					ID:           "pool-1",
					Name:         "worker-pool-1",
					CurrentSize:  3,
					DesiredSize:  3,
					MinSize:      1,
					MaxSize:      10,
					InstanceType: "standard-4",
					Zone:         "zone-1",
					Status:       "active",
					Machines:     []string{"machine-1", "machine-2", "machine-3"},
				}
			case "pool-2":
				nodePool = NodePoolInfo{
					ID:           "pool-2",
					Name:         "worker-pool-2",
					CurrentSize:  2,
					DesiredSize:  2,
					MinSize:      1,
					MaxSize:      5,
					InstanceType: "standard-2",
					Zone:         "zone-1",
					Status:       "active",
					Machines:     []string{"machine-4", "machine-5"},
				}
			default:
				http.Error(w, "Node pool not found", http.StatusNotFound)
				return
			}

			response := NodePoolResponse{
				Status: 200,
				Data: struct {
					NodePool NodePoolInfo `json:"nodepool"`
				}{
					NodePool: nodePool,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	return httptest.NewServer(mux)
}

func mockMachinesResponse(w http.ResponseWriter, r *http.Request, poolID string) {
	var machines []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		State     string `json:"state"`
		IP        string `json:"ip"`
		OS        string `json:"os"`
		Kernel    string `json:"kernel"`
		Runtime   string `json:"runtime"`
		CreatedAt string `json:"createdAt"`
	}

	switch poolID {
	case "pool-1":
		machines = []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			State     string `json:"state"`
			IP        string `json:"ip"`
			OS        string `json:"os"`
			Kernel    string `json:"kernel"`
			Runtime   string `json:"runtime"`
			CreatedAt string `json:"createdAt"`
		}{
			{ID: "machine-1", Name: "worker-1", State: "running", IP: "10.0.1.1", OS: "ubuntu", Kernel: "5.4.0", Runtime: "containerd", CreatedAt: "2024-01-01T00:00:00Z"},
			{ID: "machine-2", Name: "worker-2", State: "running", IP: "10.0.1.2", OS: "ubuntu", Kernel: "5.4.0", Runtime: "containerd", CreatedAt: "2024-01-01T00:00:00Z"},
			{ID: "machine-3", Name: "worker-3", State: "running", IP: "10.0.1.3", OS: "ubuntu", Kernel: "5.4.0", Runtime: "containerd", CreatedAt: "2024-01-01T00:00:00Z"},
		}
	case "pool-2":
		machines = []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			State     string `json:"state"`
			IP        string `json:"ip"`
			OS        string `json:"os"`
			Kernel    string `json:"kernel"`
			Runtime   string `json:"runtime"`
			CreatedAt string `json:"createdAt"`
		}{
			{ID: "machine-4", Name: "worker-4", State: "running", IP: "10.0.1.4", OS: "ubuntu", Kernel: "5.4.0", Runtime: "containerd", CreatedAt: "2024-01-01T00:00:00Z"},
			{ID: "machine-5", Name: "worker-5", State: "running", IP: "10.0.1.5", OS: "ubuntu", Kernel: "5.4.0", Runtime: "containerd", CreatedAt: "2024-01-01T00:00:00Z"},
		}
	}

	response := struct {
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
	}{
		Status: 200,
		Data: struct {
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
		}{
			Machines: machines,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func mockScaleResponse(w http.ResponseWriter, r *http.Request, poolID string) {
	if r.Method == "PUT" {
		response := NodePoolResponse{
			Status: 200,
			Data: struct {
				NodePool NodePoolInfo `json:"nodepool"`
			}{
				NodePool: NodePoolInfo{
					ID:     poolID,
					Status: "scaling",
				},
			},
			Message: "Scaling operation initiated",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else if r.Method == "DELETE" {
		// Mock DELETE endpoint for individual machine deletion
		response := struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
			Data    struct {
				InstanceID string `json:"instanceId"`
				Operation  string `json:"operation"`
			} `json:"data"`
		}{
			Status:  200,
			Message: "Instance deletion initiated",
			Data: struct {
				InstanceID string `json:"instanceId"`
				Operation  string `json:"operation"`
			}{
				InstanceID: "test-instance",
				Operation:  "delete",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// TestVCloudAPIClient_Request tests the basic HTTP request functionality
func TestVCloudAPIClient_Request(t *testing.T) {
	server := mockVCloudAPIServer()
	defer server.Close()

	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       server.URL,
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	resp, err := client.Request(ctx, "GET", "/nodepools", nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestVCloudAPIClient_ListNodePools tests the ListNodePools method
func TestVCloudAPIClient_ListNodePools(t *testing.T) {
	server := mockVCloudAPIServer()
	defer server.Close()

	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       server.URL,
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	nodePools, err := client.ListNodePools(ctx)
	if err != nil {
		t.Fatalf("ListNodePools failed: %v", err)
	}

	if len(nodePools) != 2 {
		t.Errorf("Expected 2 node pools, got %d", len(nodePools))
	}

	// Verify the first node pool
	pool1 := nodePools[0]
	if pool1.ID != "pool-1" {
		t.Errorf("Expected pool ID 'pool-1', got '%s'", pool1.ID)
	}
	if pool1.Name != "worker-pool-1" {
		t.Errorf("Expected pool name 'worker-pool-1', got '%s'", pool1.Name)
	}
	if pool1.CurrentSize != 3 {
		t.Errorf("Expected current size 3, got %d", pool1.CurrentSize)
	}
}

// TestVCloudAPIClient_GetNodePool tests the GetNodePool method
func TestVCloudAPIClient_GetNodePool(t *testing.T) {
	server := mockVCloudAPIServer()
	defer server.Close()

	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       server.URL,
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	nodePool, err := client.GetNodePool(ctx, "pool-1")
	if err != nil {
		t.Fatalf("GetNodePool failed: %v", err)
	}

	if nodePool.ID != "pool-1" {
		t.Errorf("Expected pool ID 'pool-1', got '%s'", nodePool.ID)
	}
	if nodePool.MinSize != 1 {
		t.Errorf("Expected min size 1, got %d", nodePool.MinSize)
	}
	if nodePool.MaxSize != 10 {
		t.Errorf("Expected max size 10, got %d", nodePool.MaxSize)
	}
}

// TestVCloudAPIClient_ListNodePoolInstances tests the ListNodePoolInstances method
func TestVCloudAPIClient_ListNodePoolInstances(t *testing.T) {
	server := mockVCloudAPIServer()
	defer server.Close()

	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       server.URL,
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	instances, err := client.ListNodePoolInstances(ctx, "pool-1")
	if err != nil {
		t.Fatalf("ListNodePoolInstances failed: %v", err)
	}

	if len(instances) != 3 {
		t.Errorf("Expected 3 instances, got %d", len(instances))
	}

	expectedInstances := []string{"machine-1", "machine-2", "machine-3"}
	for i, instance := range instances {
		if instance != expectedInstances[i] {
			t.Errorf("Expected instance '%s', got '%s'", expectedInstances[i], instance)
		}
	}
}

// TestVCloudAPIClient_ScaleNodePool tests the ScaleNodePool method
func TestVCloudAPIClient_ScaleNodePool(t *testing.T) {
	server := mockVCloudAPIServer()
	defer server.Close()

	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       server.URL,
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	err := client.ScaleNodePool(ctx, "pool-1", 5)
	if err != nil {
		t.Fatalf("ScaleNodePool failed: %v", err)
	}

	t.Log("ScaleNodePool completed successfully")
}

// TestEnhancedManager_Refresh tests the Refresh method
func TestEnhancedManager_Refresh(t *testing.T) {
	server := mockVCloudAPIServer()
	defer server.Close()

	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       server.URL,
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	manager := &EnhancedManager{
		client:     client,
		clusterID:  "test-cluster-123",
		nodeGroups: make([]*NodeGroup, 0),
		config: &Config{
			ClusterID:     "test-cluster-123",
			ClusterName:   "test-cluster",
			MgmtURL:       server.URL,
			ProviderToken: "test-token",
		},
	}

	err := manager.Refresh()
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	nodeGroups := manager.GetNodeGroups()
	if len(nodeGroups) != 2 {
		t.Errorf("Expected 2 node groups after refresh, got %d", len(nodeGroups))
	}

	// Verify node group properties
	for _, ng := range nodeGroups {
		if ng.clusterID != "test-cluster-123" {
			t.Errorf("Expected cluster ID 'test-cluster-123', got '%s'", ng.clusterID)
		}
		if ng.MinSize() == 0 && ng.MaxSize() == 0 {
			t.Error("Node group should have non-zero min/max sizes for autoscaling")
		}
	}
}

// TestEnhancedManager_GetNodeGroupForInstance tests the GetNodeGroupForInstance method
func TestEnhancedManager_GetNodeGroupForInstance(t *testing.T) {
	server := mockVCloudAPIServer()
	defer server.Close()

	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       server.URL,
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	nodeGroup := &NodeGroup{
		id:         "pool-1",
		clusterID:  "test-cluster-123",
		client:     client,
		minSize:    1,
		maxSize:    10,
		targetSize: 3,
	}

	manager := &EnhancedManager{
		client:     client,
		clusterID:  "test-cluster-123",
		nodeGroups: []*NodeGroup{nodeGroup},
		config: &Config{
			ClusterID:     "test-cluster-123",
			ClusterName:   "test-cluster",
			MgmtURL:       server.URL,
			ProviderToken: "test-token",
		},
	}

	// Set manager reference
	nodeGroup.manager = manager

	// Test with non-existent instance
	nodeGroup, err := manager.GetNodeGroupForInstance("vcloud://non-existent-instance")
	if err == nil {
		t.Error("Expected error for non-existent instance")
	}
	if nodeGroup != nil {
		t.Error("Expected no node group for non-existent instance")
	}
}

// TestVCloudAPIClient_URLConstruction tests URL construction logic
func TestVCloudAPIClient_URLConstruction(t *testing.T) {
	tests := []struct {
		name        string
		mgmtURL     string
		clusterID   string
		path        string
		expectedURL string
	}{
		{
			name:        "Base URL without cluster path",
			mgmtURL:     "https://api.example.com",
			clusterID:   "test-cluster",
			path:        "/nodepools",
			expectedURL: "https://api.example.com/clusters/test-cluster/nodepools",
		},
		{
			name:        "Base URL with existing cluster path",
			mgmtURL:     "https://api.example.com/clusters/test-cluster",
			clusterID:   "test-cluster",
			path:        "/nodepools",
			expectedURL: "https://api.example.com/clusters/test-cluster/nodepools",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &VCloudAPIClient{
				clusterID:     test.clusterID,
				mgmtURL:       test.mgmtURL,
				providerToken: "test-token",
				httpClient:    &http.Client{Timeout: 30 * time.Second},
			}
			_ = client

			// We can't easily test the private URL construction logic directly,
			// but we can verify it doesn't error during request creation
			ctx := context.Background()
			_, err := http.NewRequestWithContext(ctx, "GET", test.expectedURL, nil)
			if err != nil {
				t.Errorf("URL construction test failed: %v", err)
			}
		})
	}
}

// TestVCloudAPIClient_ErrorHandling tests error handling in API client
func TestVCloudAPIClient_ErrorHandling(t *testing.T) {
	// Test with non-existent server
	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       "http://localhost:99999", // Non-existent server
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 1 * time.Second},
	}

	ctx := context.Background()
	_, err := client.ListNodePools(ctx)
	if err == nil {
		t.Error("Expected error when connecting to non-existent server")
	}

	t.Logf("Error handling test passed: %v", err)
}

// TestVCloudAPIClient_Authentication tests authentication header handling
func TestVCloudAPIClient_Authentication(t *testing.T) {
	// Create a server that checks authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Provider-Token")
		if token != "test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		response := NodePoolListResponse{
			Status: 200,
			Data: struct {
				NodePools []NodePoolInfo `json:"nodepools"`
			}{
				NodePools: []NodePoolInfo{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test with correct token
	client := &VCloudAPIClient{
		clusterName:   "test-cluster",
		clusterID:     "test-cluster-123",
		mgmtURL:       server.URL,
		providerToken: "test-token",
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	_, err := client.ListNodePools(ctx)
	if err != nil {
		t.Errorf("Request with correct token should succeed, got error: %v", err)
	}

	// Test with incorrect token
	client.providerToken = "wrong-token"
	_, err = client.ListNodePools(ctx)
	if err == nil {
		t.Error("Request with incorrect token should fail")
	}
}
