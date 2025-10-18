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

package proxmox

import (
	"strings"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func TestProxmoxCloudProvider_Name(t *testing.T) {
	provider := &proxmoxCloudProvider{}
	expected := cloudprovider.ProxmoxProviderName
	if provider.Name() != expected {
		t.Errorf("Expected provider name %s, got %s", expected, provider.Name())
	}
}

func TestBuildProxmox(t *testing.T) {
	config := `{
		"api_endpoint": "https://proxmox.example.com:8006/api2/json",
		"username": "test@pam",
		"password": "testpass",
		"node_groups": [
			{
				"name": "test-pool",
				"min_size": 1,
				"max_size": 5,
				"proxmox_node": "pve1",
				"template_id": 9000,
				"vmid_start": 200,
				"vmid_end": 299,
				"vm_config": {
					"cores": 2,
					"memory": 2048,
					"storage": "local-lvm",
					"network": "vmbr0"
				}
			}
		]
	}`

	// Create a temporary file with the config
	configFile := strings.NewReader(config)

	do := cloudprovider.NodeGroupDiscoveryOptions{}
	rl := &cloudprovider.ResourceLimiter{}

	manager, err := newManager(configFile, do)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	provider := newProxmoxCloudProvider(manager, rl)
	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	if provider.Name() != cloudprovider.ProxmoxProviderName {
		t.Errorf("Expected provider name %s, got %s", cloudprovider.ProxmoxProviderName, provider.Name())
	}

	// Test that we have node groups
	nodeGroups := provider.NodeGroups()
	if len(nodeGroups) != 1 {
		t.Errorf("Expected 1 node group, got %d", len(nodeGroups))
	}

	if len(nodeGroups) > 0 {
		ng := nodeGroups[0]
		if ng.Id() != "test-pool" {
			t.Errorf("Expected node group ID 'test-pool', got %s", ng.Id())
		}
		if ng.MinSize() != 1 {
			t.Errorf("Expected min size 1, got %d", ng.MinSize())
		}
		if ng.MaxSize() != 5 {
			t.Errorf("Expected max size 5, got %d", ng.MaxSize())
		}
	}
}

func TestToProviderID(t *testing.T) {
	nodeID := "123"
	expected := "proxmox://123"
	result := toProviderID(nodeID)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestToNodeID(t *testing.T) {
	providerID := "proxmox://123"
	expected := "proxmox://123"
	result := toNodeID(providerID)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
