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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	ionos "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ionoscloud/ionos-cloud-sdk-go"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

const (
	// K8sStateActive indicates that the cluster/nodepool resource is active.
	K8sStateActive = "ACTIVE"
	// K8sStateUpdating indicates that the cluster/nodepool resource is being updated.
	K8sStateUpdating = "UPDATING"
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
func NewAPIClient(token, endpoint, userAgent string, insecure bool) APIClient {
	config := ionos.NewConfiguration("", "", token, endpoint)
	if userAgent != "" {
		config.UserAgent = userAgent
	}
	if insecure {
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint: gosec
			},
		}
	}
	config.Debug = klog.V(6).Enabled()
	// Depth > 0 is only important for listing resources. All other autoscaling related requests don't need it
	config.SetDepth(0)
	client := ionos.NewAPIClient(config)
	return client.KubernetesApi
}

// AutoscalingClient is a client abstraction used for autoscaling.
type AutoscalingClient struct {
	clientProvider
	clusterId    string
	pollTimeout  time.Duration
	pollInterval time.Duration
}

// NewAutoscalingClient contructs a new autoscaling client.
func NewAutoscalingClient(config *Config, userAgent string) (*AutoscalingClient, error) {
	c := &AutoscalingClient{
		clientProvider: newClientProvider(config, userAgent),
		clusterId:      config.ClusterId,
		pollTimeout:    config.PollTimeout,
		pollInterval:   config.PollInterval,
	}
	return c, nil
}

func newClientProvider(config *Config, userAgent string) clientProvider {
	if config.Token != "" {
		return defaultClientProvider{
			token:     config.Token,
			userAgent: userAgent,
		}
	}
	return customClientProvider{
		cloudConfigDir: config.TokensPath,
		endpoint:       config.Endpoint,
		userAgent:      userAgent,
		insecure:       config.Insecure,
	}
}

// clientProvider initializes an authenticated Ionos Cloud API client using pre-configured values
type clientProvider interface {
	GetClient() (APIClient, error)
}

type defaultClientProvider struct {
	token     string
	userAgent string
}

func (p defaultClientProvider) GetClient() (APIClient, error) {
	return NewAPIClient(p.token, "", p.userAgent, false), nil
}

type customClientProvider struct {
	cloudConfigDir string
	endpoint       string
	userAgent      string
	insecure       bool
}

func (p customClientProvider) GetClient() (APIClient, error) {
	files, err := filepath.Glob(filepath.Join(p.cloudConfigDir, "[a-zA-Z0-9]*"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("missing cloud config")
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
	return NewAPIClient(cloudConfig.Tokens[0], p.endpoint, p.userAgent, p.insecure), nil
}

// GetNodePool gets a node pool.
func (c *AutoscalingClient) GetNodePool(id string) (*ionos.KubernetesNodePool, error) {
	client, err := c.GetClient()
	if err != nil {
		return nil, err
	}
	req := client.K8sNodepoolsFindById(context.Background(), c.clusterId, id)
	nodepool, _, err := client.K8sNodepoolsFindByIdExecute(req)
	if err != nil {
		return nil, err
	}
	return &nodepool, nil
}

func resizeRequestBody(targetSize int) ionos.KubernetesNodePoolForPut {
	return ionos.KubernetesNodePoolForPut{
		Properties: &ionos.KubernetesNodePoolPropertiesForPut{
			NodeCount: pointer.Int32Ptr(int32(targetSize)),
		},
	}
}

// ResizeNodePool sets the target size of a node pool and starts the resize process.
// The node pool target size cannot be changed until this operation finishes.
func (c *AutoscalingClient) ResizeNodePool(id string, targetSize int) error {
	client, err := c.GetClient()
	if err != nil {
		return err
	}
	req := client.K8sNodepoolsPut(context.Background(), c.clusterId, id)
	req = req.KubernetesNodePool(resizeRequestBody(targetSize))
	_, _, err = client.K8sNodepoolsPutExecute(req)
	return err
}

// WaitForNodePoolResize polls the node pool until it is in state ACTIVE.
func (c *AutoscalingClient) WaitForNodePoolResize(id string, size int) error {
	klog.V(1).Infof("Waiting for node pool %s to reach target size %d", id, size)
	return wait.PollImmediate(c.pollInterval, c.pollTimeout, func() (bool, error) {
		nodePool, err := c.GetNodePool(id)
		if err != nil {
			return false, fmt.Errorf("failed to fetch node pool %s: %w", id, err)
		}
		state := *nodePool.Metadata.State
		klog.V(5).Infof("Polled node pool %s: state=%s", id, state)
		return state == K8sStateActive, nil
	})
}

// ListNodes lists nodes.
func (c *AutoscalingClient) ListNodes(id string) ([]ionos.KubernetesNode, error) {
	client, err := c.GetClient()
	if err != nil {
		return nil, err
	}
	req := client.K8sNodepoolsNodesGet(context.Background(), c.clusterId, id)
	req = req.Depth(1)
	nodes, _, err := client.K8sNodepoolsNodesGetExecute(req)
	if err != nil {
		return nil, err
	}
	return *nodes.Items, nil
}

// DeleteNode starts node deletion.
func (c *AutoscalingClient) DeleteNode(id, nodeId string) error {
	client, err := c.GetClient()
	if err != nil {
		return err
	}
	req := client.K8sNodepoolsNodesDelete(context.Background(), c.clusterId, id, nodeId)
	_, err = client.K8sNodepoolsNodesDeleteExecute(req)
	return err
}
