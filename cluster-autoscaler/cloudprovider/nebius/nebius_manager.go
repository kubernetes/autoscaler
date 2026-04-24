/*
Copyright The Kubernetes Authors.

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

package nebius

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/nebius/sdk"
	"k8s.io/klog/v2"
)

const (
	// defaultMinSize is the default minimum size for a node group when not specified.
	defaultMinSize = 0

	// defaultMaxSize is the default maximum size for a node group when not specified.
	defaultMaxSize = 100
)

// nebiusAPI abstracts Nebius SDK calls for testability.
type nebiusAPI interface {
	ListNodeGroups(ctx context.Context, req *sdk.ListNodeGroupsRequest) (*sdk.ListNodeGroupsResponse, error)
	ListInstances(ctx context.Context, req *sdk.ListInstancesRequest) (*sdk.ListInstancesResponse, error)
	GetNodeGroup(ctx context.Context, req *sdk.GetNodeGroupRequest) (*sdk.NodeGroup, error)
	UpdateNodeGroup(ctx context.Context, req *sdk.UpdateNodeGroupRequest) error
	DeleteInstance(ctx context.Context, req *sdk.DeleteInstanceRequest) error
}

// Config is the configuration for the Nebius cloud provider.
type Config struct {
	// ClusterID is the ID of the Nebius MK8S cluster.
	ClusterID string `json:"cluster_id"`

	// IAMToken is the Nebius IAM token used for authentication.
	// If not set, uses NEBIUS_IAM_TOKEN environment variable.
	IAMToken string `json:"iam_token"`

	// IAMTokenFile is a path to a file containing the IAM token.
	// Compatible with Kubernetes secrets mounted as files.
	// If both iam_token and iam_token_file are set, iam_token takes precedence.
	IAMTokenFile string `json:"iam_token_file,omitempty"`

	// ParentID is the parent folder/project ID where instances live.
	// If not set, uses NEBIUS_PARENT_ID environment variable.
	ParentID string `json:"parent_id"`

	// Domain is the Nebius API domain. Defaults to api.eu.nebius.com.
	Domain string `json:"domain,omitempty"`
}

// Manager handles Nebius communication and data caching of
// node groups (node groups in MK8S).
type Manager struct {
	mu         sync.Mutex
	client     nebiusAPI
	closer     io.Closer // SDK closer for Cleanup()
	clusterID  string
	parentID   string
	nodeGroups []*NodeGroup
}

func newManager(configReader io.Reader) (*Manager, error) {
	cfg := &Config{}
	if configReader != nil {
		body, err := io.ReadAll(configReader)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, cfg); err != nil {
			return nil, err
		}
	}

	// Read token from file if specified and no inline token is set.
	if cfg.IAMToken == "" && cfg.IAMTokenFile != "" {
		tokenBytes, err := os.ReadFile(cfg.IAMTokenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read IAM token from file %s: %w", cfg.IAMTokenFile, err)
		}
		cfg.IAMToken = strings.TrimSpace(string(tokenBytes))
	}

	// Fall back to environment variables.
	if cfg.IAMToken == "" {
		cfg.IAMToken = os.Getenv("NEBIUS_IAM_TOKEN")
	}
	if cfg.ClusterID == "" {
		cfg.ClusterID = os.Getenv("NEBIUS_CLUSTER_ID")
	}
	if cfg.ParentID == "" {
		cfg.ParentID = os.Getenv("NEBIUS_PARENT_ID")
	}

	if cfg.IAMToken == "" {
		return nil, errors.New("nebius IAM token is not provided (set iam_token, iam_token_file in config, or NEBIUS_IAM_TOKEN env var)")
	}
	if cfg.ClusterID == "" {
		return nil, errors.New("nebius cluster ID is not provided (set cluster_id in config or NEBIUS_CLUSTER_ID env var)")
	}
	if cfg.ParentID == "" {
		return nil, errors.New("nebius parent ID is not provided (set parent_id in config or NEBIUS_PARENT_ID env var)")
	}

	ctx := context.Background()

	client, err := sdk.NewClient(ctx, cfg.IAMToken, cfg.Domain)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize Nebius client: %w", err)
	}

	m := &Manager{
		client:     client,
		closer:     client,
		clusterID:  cfg.ClusterID,
		parentID:   cfg.ParentID,
		nodeGroups: make([]*NodeGroup, 0),
	}

	return m, nil
}

// Refresh refreshes the cache holding the nodegroups. This is called by the CA
// based on the `--scan-interval`. By default it's 10 seconds.
func (m *Manager) Refresh(ctx context.Context) error {
	// List all node groups for the cluster, handling pagination.
	var allNodeGroups []*sdk.NodeGroup
	pageToken := ""
	for {
		resp, err := m.client.ListNodeGroups(ctx, &sdk.ListNodeGroupsRequest{
			ParentID:  m.clusterID,
			PageToken: pageToken,
		})
		if err != nil {
			return fmt.Errorf("failed to list node groups for cluster %s: %w", m.clusterID, err)
		}
		allNodeGroups = append(allNodeGroups, resp.Items...)
		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	// List all instances in the parent folder to cache instance membership, handling pagination.
	// NOTE: The Nebius ListInstances API does not support filtering by label,
	// so we must list all instances and filter client-side. Instances not belonging
	// to any node group are discarded below.
	var allInstances []*sdk.Instance
	pageToken = ""
	for {
		resp, err := m.client.ListInstances(ctx, &sdk.ListInstancesRequest{
			ParentID:  m.parentID,
			PageToken: pageToken,
		})
		if err != nil {
			klog.Warningf("Failed to list instances for parent %s: %v. Node membership detection will be unavailable.", m.parentID, err)
			allInstances = nil
			break
		}
		allInstances = append(allInstances, resp.Items...)
		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	if len(allInstances) > 1000 {
		klog.Warningf("ListInstances returned %d instances for parent %s. "+
			"Consider using a dedicated parent folder to reduce API overhead.", len(allInstances), m.parentID)
	}

	// Build instance map by node group ID.
	instancesByNodeGroup := make(map[string]map[string]struct{})
	for _, instance := range allInstances {
		if instance.Metadata == nil {
			continue
		}
		if ngID, ok := instance.Metadata.Labels[nodeGroupIDLabel]; ok && ngID != "" {
			provID := toProviderID(instance.Metadata.ID)
			if instancesByNodeGroup[ngID] == nil {
				instancesByNodeGroup[ngID] = make(map[string]struct{})
			}
			instancesByNodeGroup[ngID][provID] = struct{}{}
		}
	}

	var groups []*NodeGroup
	for _, ng := range allNodeGroups {
		if ng.Metadata == nil {
			continue
		}

		minSize := defaultMinSize
		maxSize := defaultMaxSize

		// Read min/max from autoscaling spec.
		if ng.Spec != nil && ng.Spec.Autoscaling != nil {
			if ng.Spec.Autoscaling.MinNodeCount > 0 {
				minSize = int(ng.Spec.Autoscaling.MinNodeCount)
			}
			if ng.Spec.Autoscaling.MaxNodeCount > 0 {
				maxSize = int(ng.Spec.Autoscaling.MaxNodeCount)
			}
		}

		ngID := ng.Metadata.ID
		instances := instancesByNodeGroup[ngID]
		if instances == nil {
			instances = make(map[string]struct{})
		}

		klog.V(4).Infof("Adding node group: id=%q name=%q min=%d max=%d instances=%d",
			ngID, ng.Metadata.Name, minSize, maxSize, len(instances))

		var currentTargetSize int
		if ng.Status != nil {
			currentTargetSize = int(ng.Status.TargetNodeCount)
		}

		groups = append(groups, &NodeGroup{
			id:         ngID,
			manager:    m,
			nodeGroup:  ng,
			minSize:    minSize,
			maxSize:    maxSize,
			targetSize: currentTargetSize,
			instances:  instances,
		})
	}

	if len(groups) == 0 {
		klog.V(4).Info("No node groups found for cluster")
	}

	m.mu.Lock()
	m.nodeGroups = groups
	m.mu.Unlock()
	return nil
}

// getNodeGroups returns a snapshot of the current node groups.
func (m *Manager) getNodeGroups() []*NodeGroup {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nodeGroups
}

// setNodeGroupSize updates the node group's target size via the Nebius API.
//
// The Nebius MK8S API uses a oneOf for node group size: either Autoscaling{min, max}
// or FixedNodeCount. There is no "desired count" field within the autoscaling spec.
// To set a specific target size, we must switch to FixedNodeCount mode. On the next
// Refresh(), min/max will fall back to defaults until autoscaling is re-enabled
// externally.
func (m *Manager) setNodeGroupSize(ctx context.Context, nodeGroupID string, targetSize int) error {
	// Get current node group state.
	ng, err := m.client.GetNodeGroup(ctx, &sdk.GetNodeGroupRequest{
		ID: nodeGroupID,
	})
	if err != nil {
		return fmt.Errorf("failed to get node group %s: %w", nodeGroupID, err)
	}

	spec := ng.Spec
	if spec == nil {
		spec = &sdk.NodeGroupSpec{}
	}

	if spec.Autoscaling != nil {
		klog.Warningf("Node group %s is switching from autoscaling to fixed mode (target size %d). "+
			"The Nebius MK8S API does not support setting a desired count within autoscaling bounds.", nodeGroupID, targetSize)
	}

	fixedCount := int64(targetSize)
	updateReq := &sdk.UpdateNodeGroupRequest{
		Metadata: ng.Metadata,
		Spec: &sdk.NodeGroupSpec{
			Version:        spec.Version,
			FixedNodeCount: &fixedCount,
		},
	}

	if err := m.client.UpdateNodeGroup(ctx, updateReq); err != nil {
		return fmt.Errorf("failed to update node group %s to size %d: %w", nodeGroupID, targetSize, err)
	}

	klog.V(4).Infof("Node group %s update to size %d started", nodeGroupID, targetSize)
	return nil
}

// deleteInstances deletes specific compute instances by their provider IDs and
// then updates the node group target size to reflect the removal. Returns the
// number of instances successfully deleted and any error. If a deletion fails
// mid-way, the target size is still adjusted to account for instances that were
// successfully deleted.
func (m *Manager) deleteInstances(ctx context.Context, nodeGroupID string, providerIDs []string, currentSize int) (int, error) {
	deleted := 0
	for _, providerID := range providerIDs {
		instanceID := strings.TrimPrefix(providerID, nebiusProviderIDPrefix)
		if err := m.client.DeleteInstance(ctx, &sdk.DeleteInstanceRequest{
			ID: instanceID,
		}); err != nil {
			// Adjust target size for instances we did successfully delete.
			if deleted > 0 {
				adjustedSize := currentSize - deleted
				if sizeErr := m.setNodeGroupSize(ctx, nodeGroupID, adjustedSize); sizeErr != nil {
					klog.Errorf("Failed to adjust node group %s size after partial deletion: %v", nodeGroupID, sizeErr)
				}
			}
			return deleted, fmt.Errorf("failed to delete instance %s from node group %s: %w", instanceID, nodeGroupID, err)
		}
		deleted++
		klog.V(4).Infof("Deleted instance %s from node group %s", instanceID, nodeGroupID)
	}

	newTargetSize := currentSize - deleted
	return deleted, m.setNodeGroupSize(ctx, nodeGroupID, newTargetSize)
}

// Cleanup cleans up resources used by the manager.
func (m *Manager) Cleanup() error {
	if m.closer != nil {
		return m.closer.Close()
	}
	return nil
}
