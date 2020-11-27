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
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	uuid "github.com/satori/go.uuid"
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
	K8sNodepoolsNodesDeleteExecute(r ionos.ApiK8sNodepoolsNodesDeleteRequest) (map[string]interface{}, *ionos.APIResponse, error)
	K8sNodepoolsPut(ctx context.Context, k8sClusterId string, nodepoolId string) ionos.ApiK8sNodepoolsPutRequest
	K8sNodepoolsPutExecute(r ionos.ApiK8sNodepoolsPutRequest) (ionos.KubernetesNodePoolForPut, *ionos.APIResponse, error)
}

var apiClientFactory = NewAPIClient

// NewAPIClient creates a new IonosCloud API client.
func NewAPIClient(token, endpoint string, insecure bool) APIClient {
	config := ionos.NewConfiguration("", "", token)
	if endpoint != "" {
		config.Servers[0].URL = endpoint
	}
	if insecure {
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint: gosec
			},
		}
	}
	config.Debug = klog.V(6).Enabled()
	client := ionos.NewAPIClient(config)
	return client.KubernetesApi
}

// AutoscalingClient is a client abstraction used for autoscaling.
type AutoscalingClient struct {
	client       APIClient
	clusterId    string
	endpoint     string
	insecure     bool
	pollTimeout  time.Duration
	pollInterval time.Duration
	tokens       map[string]string
}

// NewAutoscalingClient contructs a new autoscaling client.
func NewAutoscalingClient(config *Config) (*AutoscalingClient, error) {
	c := &AutoscalingClient{
		clusterId:    config.ClusterId,
		endpoint:     config.Endpoint,
		insecure:     config.Insecure,
		pollTimeout:  config.PollTimeout,
		pollInterval: config.PollInterval,
	}
	if config.Token != "" {
		c.client = apiClientFactory(config.Token, config.Endpoint, config.Insecure)
	} else if config.TokensPath != "" {
		tokens, err := loadTokensFromFilesystem(config.TokensPath)
		if err != nil {
			return nil, err
		}
		c.tokens = tokens
	}
	return c, nil
}

// loadTokensFromFilesystem loads a mapping of node pool UUIDs to JWT tokens from the given path.
func loadTokensFromFilesystem(path string) (map[string]string, error) {
	klog.V(3).Infof("Loading tokens from: %s", path)
	filenames, _ := filepath.Glob(filepath.Join(path, "*"))
	tokens := make(map[string]string)
	for _, filename := range filenames {
		name := filepath.Base(filename)
		if _, err := uuid.FromString(name); err != nil {
			continue
		}
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		tokens[name] = string(data)
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%s did not contain any valid entries", path)
	}
	return tokens, nil
}

// apiClientFor returns an IonosCloud API client for the given node pool.
// If the cache was not initialized with a client a new one will be created using a cached token.
func (c *AutoscalingClient) apiClientFor(id string) (APIClient, error) {
	if c.client != nil {
		return c.client, nil
	}
	token, exists := c.tokens[id]
	if !exists {
		return nil, fmt.Errorf("missing token for node pool %s", id)
	}
	return apiClientFactory(token, c.endpoint, c.insecure), nil
}

// GetNodePool gets a node pool.
func (c *AutoscalingClient) GetNodePool(id string) (*ionos.KubernetesNodePool, error) {
	client, err := c.apiClientFor(id)
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

func resizeRequestBody(targetSize int) ionos.KubernetesNodePoolPropertiesForPut {
	return ionos.KubernetesNodePoolPropertiesForPut{
		NodeCount: pointer.Int32Ptr(int32(targetSize)),
	}
}

// ResizeNodePool sets the target size of a node pool and starts the resize process.
// The node pool target size cannot be changed until this operation finishes.
func (c *AutoscalingClient) ResizeNodePool(id string, targetSize int) error {
	client, err := c.apiClientFor(id)
	if err != nil {
		return err
	}
	req := client.K8sNodepoolsPut(context.Background(), c.clusterId, id)
	req = req.KubernetesNodePoolProperties(resizeRequestBody(targetSize))
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
	client, err := c.apiClientFor(id)
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
	client, err := c.apiClientFor(id)
	if err != nil {
		return err
	}
	req := client.K8sNodepoolsNodesDelete(context.Background(), c.clusterId, id, nodeId)
	_, _, err = client.K8sNodepoolsNodesDeleteExecute(req)
	return err
}
