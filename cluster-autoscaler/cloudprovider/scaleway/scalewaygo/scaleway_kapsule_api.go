/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaygo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/version"
	"k8s.io/klog/v2"
)

const (
	defaultApiURL      string = "https://api.scaleway.com"
	defaultHTTPTimeout        = 30
	pageSizeListPools  int    = 100
	pageSizeListNodes  int    = 100
)

var (
	// ErrMissingClusterID is returned when no cluster id has been found
	// either in env variables or in config file
	ErrMissingClusterID = errors.New("cluster ID is not provided")
	// ErrMissingSecretKey is returned  when no secret key has been found
	// either in env variables or in config file
	ErrMissingSecretKey = errors.New("scaleway secret key is not provided")
	// ErrMissingRegion is returned when no region has been found
	// either in env variables or in config file
	ErrMissingRegion = errors.New("region is not provided")

	// ErrClientSide indicates an error on user side
	ErrClientSide = errors.New("400 error type")
	// ErrServerSide indicates an error on server side
	ErrServerSide = errors.New("500 error type")
	// ErrOther indicates a generic HTTP error
	ErrOther = errors.New("generic error type")
)

// Client is used to talk to Scaleway Kapsule API
type Client interface {
	ListPools(ctx context.Context, clusterID string) ([]PoolWithGenericNodeSpecs, error)
	UpdatePool(ctx context.Context, poolID string, size int) (Pool, error)
	ListNodes(ctx context.Context, cluserID string) ([]Node, error)
	DeleteNode(ctx context.Context, nodeID string) (Node, error)
}

// Config is used to deserialize config file passed with flag `cloud-config`
type Config struct {
	ClusterID string `json:"cluster_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	ApiUrl    string `json:"api_url"`
	UserAgent string
}

// NewClient returns a new Client able to talk to Scaleway API
func NewClient(cfg Config) (*client, error) {
	if cfg.ClusterID == "" {
		return nil, ErrMissingClusterID
	}
	if cfg.SecretKey == "" {
		return nil, ErrMissingSecretKey
	}
	if cfg.Region == "" {
		return nil, ErrMissingRegion
	}
	if cfg.ApiUrl == "" {
		cfg.ApiUrl = defaultApiURL
	}

	hc := &http.Client{
		Timeout: defaultHTTPTimeout * time.Second,
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}

	return &client{
		httpClient: hc,
		apiURL:     cfg.ApiUrl,
		token:      cfg.SecretKey,
		userAgent:  fmt.Sprintf("%s/%s cluster-id/%s", cfg.UserAgent, version.ClusterAutoscalerVersion, cfg.ClusterID),
		region:     cfg.Region,
	}, nil
}

// scalewayRequest contains all the contents related to performing a request on the Scaleway API.
type scalewayRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   io.Reader
}

// client contains necessary information to perform API calls
type client struct {
	httpClient *http.Client
	apiURL     string
	token      string
	userAgent  string
	region     string
}

func (req *scalewayRequest) getURL(apiURL string) (*url.URL, error) {
	completeURL, err := url.Parse(apiURL + req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid url %s: %s", apiURL+req.Path, err)
	}
	completeURL.RawQuery = req.Query.Encode()

	return completeURL, nil
}

func (c *client) ApiURL() string {
	return c.apiURL
}

func (c *client) Token() string {
	return c.token
}

func (c *client) Region() string {
	return c.region
}

// do perform a single HTTP request based on the generic Request object.
func (c *client) do(ctx context.Context, req *scalewayRequest, res interface{}) error {
	if req == nil {
		return errors.New("request must be non-nil")
	}

	// build URL
	completeURL, err := req.getURL(c.apiURL)
	if err != nil {
		return err
	}

	// build request
	httpRequest, err := http.NewRequest(req.Method, completeURL.String(), req.Body)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	httpRequest.Header.Set("User-Agent", c.userAgent)
	httpRequest.Header.Set("X-Auth-Token", c.token)
	if req.Body != nil {
		httpRequest.Header.Set("Content-Type", "application/json")
	}

	httpRequest = httpRequest.WithContext(ctx)

	// execute request
	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}

	defer func() {
		if err := httpResponse.Body.Close(); err != nil {
			klog.Errorf("failed to close response body: %v", err)
		}
	}()

	ct := httpResponse.Header.Get("Content-Type")
	if ct != "application/json" {
		return fmt.Errorf("unexpected content-type: %s with status: %s", ct, httpResponse.Status)
	}

	err = json.NewDecoder(httpResponse.Body).Decode(&res)
	if err != nil {
		return fmt.Errorf("could not parse %s response body: %w", ct, err)
	}

	switch {
	case httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300:
		return nil
	case httpResponse.StatusCode >= 400 && httpResponse.StatusCode < 500:
		err = ErrClientSide
	case httpResponse.StatusCode >= 500 && httpResponse.StatusCode < 600:
		err = ErrServerSide
	default:
		err = ErrOther

	}

	return fmt.Errorf("%d %v %v: %w", httpResponse.StatusCode, httpRequest.Method, httpRequest.URL, err)
}

// NodeStatus is the state in which a node might be
type NodeStatus string

const (
	// NodeStatusCreating indicates that node is provisioning the underlying instance/BM
	NodeStatusCreating = NodeStatus("creating")
	// NodeStatusStarting indicates that node is being configured and/or booting
	NodeStatusStarting = NodeStatus("starting")
	// NodeStatusRegistering indicates that underlying node has booted and k8s services are starting
	NodeStatusRegistering = NodeStatus("registering")
	// NodeStatusNotReady indicates that k8s has marked this node as `NotReady`
	NodeStatusNotReady = NodeStatus("not_ready")
	// NodeStatusReady indicates that node is ready for use
	NodeStatusReady = NodeStatus("ready")
	// NodeStatusDeleting indicates that node is being deleted
	NodeStatusDeleting = NodeStatus("deleting")
	// NodeStatusDeleted indicates that node is deleted
	NodeStatusDeleted = NodeStatus("deleted")
	// NodeStatusLocked indicates that node has been locked for legal reasons
	NodeStatusLocked = NodeStatus("locked")
	// NodeStatusRebooting indicates that node is rebooting
	NodeStatusRebooting = NodeStatus("rebooting")
	// NodeStatusCreationError indicates that node failed to create
	NodeStatusCreationError = NodeStatus("creation_error")
	// NodeStatusUpgrading indicates that this node CP is currently upgrading k8s version
	NodeStatusUpgrading = NodeStatus("upgrading")
)

// Node represents an instance running in a scaleway pool
type Node struct {
	// ID: the ID of the node
	ID string `json:"id"`
	// PoolID: the pool ID of the node
	PoolID string `json:"pool_id"`
	// ClusterID: the cluster ID of the node
	ClusterID string `json:"cluster_id"`
	// ProviderID: the underlying instance ID
	ProviderID string `json:"provider_id"`
	// Name: the name of the node
	Name string `json:"name"`
	// Status: the status of the node
	Status NodeStatus `json:"status"`
	// CreatedAt: the date at which the node was created
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: the date at which the node was last updated
	UpdatedAt *time.Time `json:"updated_at"`
}

// PoolStatus is the state in which a pool might be (unused)
type PoolStatus string

// These are possible statuses for a scaleway pool
const (
	PoolStatusReady     = PoolStatus("ready")
	PoolStatusDeleting  = PoolStatus("deleting")
	PoolStatusDeleted   = PoolStatus("deleted")
	PoolStatusScaling   = PoolStatus("scaling")
	PoolStatusWarning   = PoolStatus("warning")
	PoolStatusLocked    = PoolStatus("locked")
	PoolStatusUpgrading = PoolStatus("upgrading")
)

// Pool is the abstraction used to gather nodes with the same specs
type Pool struct {
	// ID: the ID of the pool
	ID string `json:"id"`
	// ClusterID: the cluster ID of the pool
	ClusterID string `json:"cluster_id"`
	// CreatedAt: the date at which the pool was created
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: the date at which the pool was last updated
	UpdatedAt *time.Time `json:"updated_at"`
	// Name: the name of the pool
	Name string `json:"name"`
	// Status: the status of the pool
	Status PoolStatus `json:"status"`
	// Version: the version of the pool
	Version string `json:"version"`
	// NodeType: the node type is the type of Scaleway Instance wanted for the pool
	NodeType string `json:"node_type"`
	// Autoscaling: the enablement of the autoscaling feature for the pool
	Autoscaling bool `json:"autoscaling"`
	// Size: the size (number of nodes) of the pool
	Size int `json:"size"`
	// MinSize: the minimum size of the pool
	MinSize int `json:"min_size"`
	// MaxSize: the maximum size of the pool
	MaxSize int `json:"max_size"`
	// Zone: the zone where the nodes will be spawn in
	Zone string `json:"zone"`
}

// GenericNodeSpecs represents NodeType specs used for scale-up simulations.
// it is used to select the appropriate pool to scale-up.
type GenericNodeSpecs struct {
	NodePricePerHour float32           `json:"node_price_per_hour"`
	Capacity         map[string]int64  `json:"capacity"`
	Allocatable      map[string]int64  `json:"allocatable"`
	Labels           map[string]string `json:"labels"`
	Taints           map[string]string `json:"taints"`
}

// PoolWithGenericNodeSpecs contains the requested `Pool` with additional `Specs` information
type PoolWithGenericNodeSpecs struct {
	Pool  Pool             `json:"pool"`
	Specs GenericNodeSpecs `json:"specs"`
}

// ListPoolsResponse is returned from `ListPools` method
type ListPoolsResponse struct {
	// Pools: the paginated returned pools
	Pools []PoolWithGenericNodeSpecs `json:"pools"`
}

// ListPools returns pools associated to a cluster id
func (c *client) ListPools(ctx context.Context, clusterID string) ([]PoolWithGenericNodeSpecs, error) {
	klog.V(4).Info("ListPools,ClusterID=", clusterID)

	var currentPage = 1
	var pools []PoolWithGenericNodeSpecs
	for {
		paginatedPools, err := c.listPoolsPaginated(ctx, clusterID, currentPage, pageSizeListPools)
		if err != nil {
			return []PoolWithGenericNodeSpecs{}, err
		}
		pools = append(pools, paginatedPools...)

		if len(paginatedPools) < pageSizeListPools || len(paginatedPools) == 0 {
			break
		}

		currentPage++
	}

	return pools, nil
}

func (c *client) listPoolsPaginated(ctx context.Context, clusterID string, page, pageSize int) ([]PoolWithGenericNodeSpecs, error) {
	if len(clusterID) == 0 {
		return nil, errors.New("clusterID cannot be empty in request")
	}

	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("page_size", strconv.Itoa(pageSize))

	scwReq := &scalewayRequest{
		Method: "GET",
		Path:   fmt.Sprintf("/k8s/v1/regions/%s/clusters/%s/pools-autoscaler", c.region, clusterID),
		Query:  query,
	}

	var resp ListPoolsResponse
	err := c.do(ctx, scwReq, &resp)

	return resp.Pools, err
}

// UpdatePoolRequest is passed to `UpdatePool` method
type UpdatePoolRequest struct {
	// PoolID: the ID of the pool to update
	PoolID string `json:"-"`
	// Size: the new size for the pool
	Size int `json:"size"`
}

// UpdatePool is used to resize a pool, to decrease pool size `DeleteNode` should be used instead
func (c *client) UpdatePool(ctx context.Context, poolID string, size int) (Pool, error) {
	klog.V(4).Info("UpdatePool,PoolID=", poolID)

	if len(poolID) == 0 {
		return Pool{}, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scalewayRequest{
		Method: "PATCH",
		Path:   fmt.Sprintf("/k8s/v1/regions/%s/pools/%s", c.region, poolID),
	}

	buf, err := json.Marshal(UpdatePoolRequest{PoolID: poolID, Size: size})
	if err != nil {
		return Pool{}, err
	}
	scwReq.Body = bytes.NewReader(buf)

	var resp Pool
	err = c.do(ctx, scwReq, &resp)

	return resp, err
}

// ListNodesResponse is returned from `ListNodes` method
type ListNodesResponse struct {
	// Nodes: the paginated returned nodes
	Nodes []Node `json:"nodes"`
}

// ListNodes returns the Nodes associated to a Cluster
func (c *client) ListNodes(ctx context.Context, clusterID string) ([]Node, error) {
	klog.V(4).Info("ListNodes,ClusterID=", clusterID)

	var currentPage = 1
	var nodes []Node
	for {
		paginatedNodes, err := c.listNodesPaginated(ctx, clusterID, currentPage, pageSizeListPools)
		if err != nil {
			return []Node{}, err
		}
		nodes = append(nodes, paginatedNodes...)

		if len(paginatedNodes) < int(pageSizeListNodes) || len(paginatedNodes) == 0 {
			break
		}

		currentPage++
	}

	return nodes, nil
}

func (c *client) listNodesPaginated(ctx context.Context, clusterID string, page, pageSize int) ([]Node, error) {
	if len(clusterID) == 0 {
		return nil, errors.New("clusterID cannot be empty in request")
	}

	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("page_size", strconv.Itoa(pageSize))

	scwReq := &scalewayRequest{
		Method: "GET",
		Path:   fmt.Sprintf("/k8s/v1/regions/%s/clusters/%s/nodes", c.region, clusterID),
		Query:  query,
	}

	var resp ListNodesResponse
	err := c.do(ctx, scwReq, &resp)

	return resp.Nodes, err
}

// DeleteNode asynchronously deletes a Node by its id
func (c *client) DeleteNode(ctx context.Context, nodeID string) (Node, error) {
	klog.V(4).Info("DeleteNode,NodeID=", nodeID)

	if len(nodeID) == 0 {
		return Node{}, errors.New("field NodeID cannot be empty in request")
	}

	scwReq := &scalewayRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/k8s/v1/regions/%s/nodes/%s", c.region, nodeID),
	}

	var resp Node
	err := c.do(ctx, scwReq, &resp)

	return resp, err
}
