/*
Copyright 2020 The Kubernetes Authors.

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

package ionoscloud

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/wait"
	ionos "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ionoscloud/ionos-cloud-sdk-go"
	"k8s.io/klog/v2"
)

const (
	// K8sNodeStateReady indicates that the Kubernetes Node was provisioned and joined the cluster.
	K8sNodeStateReady = "READY"
	// K8sNodeStateProvisioning indicates that the instance backing the Kubernetes Node is being provisioned.
	K8sNodeStateProvisioning = "PROVISIONING"
	// K8sNodeStateProvisioned indicates that the instance backing the Kubernetes Node is ready
	K8sNodeStateProvisioned = "PROVISIONED"
	// K8sNodeStateTerminating indicates that the Kubernetes Node is being deleted.
	K8sNodeStateTerminating = "TERMINATING"
	// K8sNodeStateRebuilding indicates that the Kubernetes Node is being rebuilt.
	K8sNodeStateRebuilding = "REBUILDING"
)

// APIClient provides a subset of API calls necessary to perform autoscaling.
type APIClient interface {
	K8sNodepoolsFindById(ctx context.Context, k8sClusterId string, nodepoolId string) ionos.ApiK8sNodepoolsFindByIdRequest
	K8sNodepoolsFindByIdExecute(r ionos.ApiK8sNodepoolsFindByIdRequest) (ionos.KubernetesNodePool, *ionos.APIResponse, error)
	K8sNodepoolsNodesGet(ctx context.Context, k8sClusterId string, nodepoolId string) ionos.ApiK8sNodepoolsNodesGetRequest
	K8sNodepoolsNodesGetExecute(r ionos.ApiK8sNodepoolsNodesGetRequest) (ionos.KubernetesNodes, *ionos.APIResponse, error)
	K8sNodepoolsNodesDelete(ctx context.Context, k8sClusterId string, nodepoolId string, nodeId string) ionos.ApiK8sNodepoolsNodesDeleteRequest
	K8sNodepoolsNodesDeleteExecute(r ionos.ApiK8sNodepoolsNodesDeleteRequest) (*ionos.APIResponse, error)
	K8sNodepoolsPut(ctx context.Context, k8sClusterId string, nodepoolId string) ionos.ApiK8sNodepoolsPutRequest
	K8sNodepoolsPutExecute(r ionos.ApiK8sNodepoolsPutRequest) (ionos.KubernetesNodePool, *ionos.APIResponse, error)
}

// NewAPIClient creates a new IonosCloud API client.
func NewAPIClient(cfg *Config, userAgent string) APIClient {
	config := ionos.NewConfiguration("", "", cfg.Token, cfg.Endpoint)
	if userAgent != "" {
		config.UserAgent = userAgent
	}
	if cfg.Insecure {
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint: gosec
			},
		}
	}
	for key, value := range cfg.AdditionalHeaders {
		config.AddDefaultHeader(key, value)
	}

	setLogLevel(config)
	config.Logger = klogAdapter{}
	// Depth > 0 is only important for listing resources. All other autoscaling related requests don't need it
	config.SetDepth(0)
	client := ionos.NewAPIClient(config)
	return client.KubernetesApi
}

func setLogLevel(config *ionos.Configuration) {
	switch {
	case klog.V(7).Enabled():
		config.LogLevel = ionos.Trace
	case klog.V(5).Enabled():
		config.LogLevel = ionos.Debug
	}
}

type klogAdapter struct{}

func (klogAdapter) Printf(format string, args ...interface{}) {
	klog.InfofDepth(1, "IONOSLOG "+format, args...)
}

// AutoscalingClient is a client abstraction used for autoscaling.
type AutoscalingClient struct {
	client    APIClient
	cfg       *Config
	userAgent string
}

// NewAutoscalingClient contructs a new autoscaling client.
func NewAutoscalingClient(config *Config, userAgent string) *AutoscalingClient {
	c := &AutoscalingClient{cfg: config, userAgent: userAgent}
	if config.Token != "" {
		c.client = NewAPIClient(config, userAgent)
	}
	return c
}

func (c *AutoscalingClient) getClient() (APIClient, error) {
	if c.client != nil {
		return c.client, nil
	}

	files, err := filepath.Glob(filepath.Join(c.cfg.TokensPath, "[a-zA-Z0-9]*"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("missing cloud config")
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		return nil, err
	}
	cloudConfig := struct {
		Tokens []string `json:"tokens"`
	}{}
	if err := json.Unmarshal(data, &cloudConfig); err != nil {
		return nil, err
	}
	if len(cloudConfig.Tokens) == 0 {
		return nil, fmt.Errorf("missing tokens for cloud config %s", filepath.Base(files[0]))
	}
	cfg := *c.cfg
	cfg.Token = cloudConfig.Tokens[0]
	return NewAPIClient(&cfg, c.userAgent), nil
}

// GetNodePool gets a node pool.
func (c *AutoscalingClient) GetNodePool(id string) (*ionos.KubernetesNodePool, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}
	req := client.K8sNodepoolsFindById(context.Background(), c.cfg.ClusterID, id)
	nodepool, resp, err := client.K8sNodepoolsFindByIdExecute(req)
	registerRequest("GetNodePool", resp, err)
	if err != nil {
		return nil, err
	}
	return &nodepool, nil
}

func resizeRequestBody(targetSize int32) ionos.KubernetesNodePoolForPut {
	return ionos.KubernetesNodePoolForPut{
		Properties: &ionos.KubernetesNodePoolPropertiesForPut{
			NodeCount: &targetSize,
		},
	}
}

// ResizeNodePool sets the target size of a node pool and starts the resize process.
// The node pool target size cannot be changed until this operation finishes.
func (c *AutoscalingClient) ResizeNodePool(id string, targetSize int32) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}
	req := client.K8sNodepoolsPut(context.Background(), c.cfg.ClusterID, id)
	req = req.KubernetesNodePool(resizeRequestBody(targetSize))
	_, resp, err := client.K8sNodepoolsPutExecute(req)
	registerRequest("ResizeNodePool", resp, err)
	return err
}

// WaitForNodePoolResize polls the node pool until it is in state ACTIVE.
func (c *AutoscalingClient) WaitForNodePoolResize(id string, size int) error {
	klog.V(1).Infof("Waiting for node pool %s to reach target size %d", id, size)
	return wait.PollUntilContextTimeout(context.Background(), c.cfg.PollInterval, c.cfg.PollTimeout, true,
		func(context.Context) (bool, error) {
			nodePool, err := c.GetNodePool(id)
			if err != nil {
				return false, fmt.Errorf("failed to fetch node pool %s: %w", id, err)
			}
			state := *nodePool.Metadata.State
			klog.V(5).Infof("Polled node pool %s: state=%s", id, state)
			return state == ionos.Active, nil
		})
}

// ListNodes lists nodes.
func (c *AutoscalingClient) ListNodes(nodePoolID string) ([]ionos.KubernetesNode, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}
	req := client.K8sNodepoolsNodesGet(context.Background(), c.cfg.ClusterID, nodePoolID)
	req = req.Depth(1)
	nodes, resp, err := client.K8sNodepoolsNodesGetExecute(req)
	registerRequest("ListNodes", resp, err)
	if err != nil {
		return nil, err
	}
	return *nodes.Items, nil
}

// DeleteNode starts node deletion.
func (c *AutoscalingClient) DeleteNode(nodePoolID, nodeID string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}
	req := client.K8sNodepoolsNodesDelete(context.Background(), c.cfg.ClusterID, nodePoolID, nodeID)
	resp, err := client.K8sNodepoolsNodesDeleteExecute(req)
	registerRequest("DeleteNode", resp, err)
	return err
}
