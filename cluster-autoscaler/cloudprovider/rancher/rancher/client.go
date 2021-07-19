/*
Copyright 2021 The Kubernetes Authors.

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

package rancher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Cluster represents a Rancher cluster.
type Cluster struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NodePool represents a Rancher nodePool.
type NodePool struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ClusterID         string `json:"clusterId"`
	ControlPlane      bool   `json:"controlPlane"`
	Etcd              bool   `json:"etcd"`
	Worker            bool   `json:"worker"`
	Quantity          int    `json:"quantity"`
	State             string `json:"state"`
	DrainBeforeDelete bool   `json:"drainBeforeDelete"`
}

type nodePoolResponse struct {
	Data []NodePool `json:"data"`
}

// Node represents a Rancher node.
type Node struct {
	ID         string `json:"id"`
	Name       string `json:"nodeName"`
	ProviderID string `json:"providerId"`
	ClusterID  string `json:"clusterId"`
	NodePoolID string `json:"nodePoolId"`
	State      string `json:"state"`
}

type nodeResponse struct {
	Data []Node `json:"data"`
}

// Client is an HTTP client for RancherAPI.
type Client struct {
	url   string
	token string
	cli   *http.Client
}

// New returns a new client for rancher cli.
func New(url, token string) *Client {
	c := http.DefaultClient
	if os.Getenv("AUTOSCALER_HTTP_DEBUG") != "" {
		c.Transport = loggerTransport
	}

	return &Client{
		url:   url,
		token: token,
		cli:   c,
	}
}

// ClusterByID returns the cluster by id
func (c *Client) ClusterByID(id string) (*Cluster, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/clusters/%s", c.url, id), nil, nil)
	if err != nil {
		return nil, err
	}

	var np Cluster
	if err := json.Unmarshal(resp, &np); err != nil {
		return nil, err
	}

	return &np, err
}

// ResizeNodePool resizes the selected nodePool
func (c *Client) ResizeNodePool(id string, size int) (*NodePool, error) {
	url := fmt.Sprintf("%s/nodePools/%s", c.url, id)
	reqBody, err := json.Marshal(map[string]interface{}{
		"quantity": size,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(http.MethodPut, url, bytes.NewBuffer(reqBody), nil)
	if err != nil {
		return nil, err
	}

	var np NodePool
	if err := json.Unmarshal(resp, &np); err != nil {
		return nil, err
	}

	return &np, nil
}

// NodePoolsByCluster returns all node pools by cluster.
func (c *Client) NodePoolsByCluster(clusterID string) ([]NodePool, error) {
	url := fmt.Sprintf("%s/clusters/%s/nodepools", c.url, clusterID)
	resp, err := c.doRequest(http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, err
	}

	var body nodePoolResponse
	if err := json.Unmarshal(resp, &body); err != nil {
		return nil, err
	}

	if len(body.Data) == 0 {
		return nil, fmt.Errorf("clusterID: %q does not have nodePools", clusterID)
	}

	return body.Data, nil
}

// ScaleDownNode deletes the specific node and  scale down the node pool.
func (c *Client) ScaleDownNode(nodeID string) error {
	url := fmt.Sprintf("%s/nodes/%s?action=scaledown", c.url, nodeID)
	_, err := c.doRequest(http.MethodPost, url, nil, nil)
	return err
}

// NodePoolByID returns a node pool by id.
func (c *Client) NodePoolByID(id string) (*NodePool, error) {
	url := fmt.Sprintf("%s/nodePools/%s", c.url, id)
	resp, err := c.doRequest(http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, err
	}

	var np NodePool
	if err := json.Unmarshal(resp, &np); err != nil {
		return nil, err
	}

	return &np, nil
}

// NodesByNodePool returns all the nodes in a node pool.
func (c *Client) NodesByNodePool(nodePoolID string) ([]Node, error) {
	return c.nodesByFilters(map[string]string{"nodePoolId": nodePoolID})
}

// NodeByProviderID returns a node by providerID
func (c *Client) NodeByProviderID(providerID string) (*Node, error) {
	nodes, err := c.nodesByFilters(map[string]string{"providerId": providerID})
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("nodeID: %q does not exist", providerID)
	}

	return &nodes[0], nil
}

// NodeByNameAndCluster returns the node that match name and cluster
func (c *Client) NodeByNameAndCluster(name, cluster string) (*Node, error) {
	nodes, err := c.nodesByFilters(map[string]string{"name": name, "clusterId": cluster})
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("node: %q for cluster: %q does not exist", name, cluster)
	}

	return &nodes[0], nil
}

func (c *Client) nodesByFilters(filters map[string]string) ([]Node, error) {
	url := fmt.Sprintf("%s/nodes", c.url)
	resp, err := c.doRequest(http.MethodGet, url, nil, filters)
	if err != nil {
		return nil, err
	}

	var body nodeResponse
	if err := json.Unmarshal(resp, &body); err != nil {
		return nil, err
	}

	return body.Data, nil
}

func (c *Client) doRequest(verb, url string, body io.Reader, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest(verb, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("bearer %s", c.token))
	req.Header.Add("Content-Type", "application/json")
	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var b bytes.Buffer
	if _, err := io.Copy(&b, resp.Body); err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("rancher: authentication failed with credentials provided")
	case http.StatusForbidden:
		return nil, fmt.Errorf("rancher: %s is forbidden", url)
	case http.StatusNotFound:
		return nil, fmt.Errorf("rancher: %s resource not found", url)
	default:
		return nil, fmt.Errorf("rancher: %s invalid status code %d", url, resp.StatusCode)
	}

	return b.Bytes(), nil
}
