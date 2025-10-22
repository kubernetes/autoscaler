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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog/v2"
)

// ProxmoxClient interface for Proxmox VE API operations
type ProxmoxClient interface {
	// CreateVM creates a new VM from a template
	CreateVM(ctx context.Context, nodeID string, templateID int, vmID int, config VMConfig, nodeGroup string) error
	CreatePool(ctx context.Context, name string) error
	// DeleteVM deletes a VM
	DeleteVM(ctx context.Context, nodeID string, vmID int) error
	GetVM(ctx context.Context, nodeName string, vmID int) (*VM, error)
	// GetVMs returns list of VMs in a node
	GetVMs(ctx context.Context, nodeID string) ([]VM, error)
	// GetNodes returns list of Proxmox nodes
	GetNodes(ctx context.Context) ([]ProxmoxNode, error)
	// StartVM starts a VM
	StartVM(ctx context.Context, nodeID string, vmID int) error
	// StopVM stops a VM
	StopVM(ctx context.Context, nodeID string, vmID int) error
	GetNextFreeVMID(ctx context.Context) (int, error)
}

// Manager handles Proxmox communication and data caching of node groups
type Manager struct {
	client     ProxmoxClient
	nodeGroups []*NodeGroup
	config     *Config
}

// Config is the configuration of the Proxmox cloud provider
type Config struct {
	// APIEndpoint is the Proxmox VE API endpoint (e.g., https://proxmox.example.com:8006/api2/json)
	APIEndpoint string `json:"api_endpoint"`

	// Username for Proxmox VE authentication
	Username string `json:"username"`

	// Password for Proxmox VE authentication
	Password string `json:"password"`

	// TokenID for Proxmox VE API token authentication (alternative to username/password)
	TokenID string `json:"token_id"`

	// TokenSecret for Proxmox VE API token authentication
	TokenSecret string `json:"token_secret"`

	// InsecureSkipTLSVerify skips TLS certificate verification
	InsecureSkipTLSVerify bool `json:"insecure_skip_tls_verify"`

	// NodeGroups configuration
	NodeGroups []NodeGroupConfig `json:"node_groups"`
}

// NodeGroupConfig represents configuration for a node group
type NodeGroupConfig struct {
	// Name is the unique identifier for this node group
	Name string `json:"name"`

	// MinSize is the minimum number of nodes in this group
	MinSize int `json:"min_size"`

	// MaxSize is the maximum number of nodes in this group
	MaxSize int `json:"max_size"`

	// ProxmoxNode is the Proxmox node where VMs should be created
	ProxmoxNode string `json:"proxmox_node"`

	// TemplateID is the VM template ID to clone from
	TemplateID int `json:"template_id"`

	// VMConfig contains VM creation parameters
	VMConfig VMConfig `json:"vm_config"`
}

// VMConfig represents VM configuration parameters
type VMConfig struct {
	Cores   int      `json:"cores"`
	Memory  int      `json:"memory"`
	Storage string   `json:"storage"`
	Network string   `json:"network"`
	Tags    []string `json:"tags"`
}

// VM represents a Proxmox VM
type VM struct {
	ID     int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Node   string `json:"node"`
	Tags   string `json:"tags"`
	UUID   string `json:"uuid"`
}

// ProxmoxNode represents a Proxmox cluster node
type ProxmoxNode struct {
	ID     string `json:"node"`
	Status string `json:"status"`
	Online bool   `json:"online"`
}

func newManager(configReader io.Reader, do cloudprovider.NodeGroupDiscoveryOptions) (*Manager, error) {
	cfg := &Config{}
	if configReader != nil {
		body, err := io.ReadAll(configReader)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.APIEndpoint == "" {
		return nil, errors.New("Proxmox API endpoint is not provided")
	}

	// Validate authentication
	if (cfg.Username == "" || cfg.Password == "") && (cfg.TokenID == "" || cfg.TokenSecret == "") {
		return nil, errors.New("Proxmox authentication credentials are not provided")
	}

	// Create Proxmox client using the go-proxmox library
	client, err := NewClient(
		cfg.APIEndpoint,
		cfg.Username,
		cfg.Password,
		cfg.TokenID,
		cfg.TokenSecret,
		cfg.InsecureSkipTLSVerify,
		"not-available",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Proxmox client: %v", err)
	}

	manager := &Manager{
		client: client,
		config: cfg,
	}

	// Initialize node groups
	if err := manager.initializeNodeGroups(); err != nil {
		return nil, fmt.Errorf("failed to initialize node groups: %v", err)
	}

	return manager, nil
}

func (m *Manager) initializeNodeGroups() error {
	for _, ngConfig := range m.config.NodeGroups {
		ng := &NodeGroup{
			id:            ngConfig.Name,
			manager:       m,
			minSize:       ngConfig.MinSize,
			maxSize:       ngConfig.MaxSize,
			proxmoxNode:   ngConfig.ProxmoxNode,
			templateID:    ngConfig.TemplateID,
			vmConfig:      ngConfig.VMConfig,
			createdVMs:    make(map[int]string),
			instanceCache: NewInstanceCache(),
		}
		m.nodeGroups = append(m.nodeGroups, ng)
		klog.V(4).Infof("Initialized Proxmox node group: %s", ngConfig.Name)
	}

	return nil
}

// Refresh refreshes the cache of node groups and their instances
func (m *Manager) Refresh() error {
	klog.V(4).Info("Refreshing Proxmox node groups")
	for _, ng := range m.nodeGroups {
		if err := ng.refresh(); err != nil {
			klog.Errorf("Failed to refresh node group %s: %v", ng.id, err)
			return err
		}
	}
	return nil
}

// GetNodeGroup returns a node group by ID
func (m *Manager) GetNodeGroup(id string) *NodeGroup {
	for _, ng := range m.nodeGroups {
		if ng.id == id {
			return ng
		}
	}
	return nil
}
