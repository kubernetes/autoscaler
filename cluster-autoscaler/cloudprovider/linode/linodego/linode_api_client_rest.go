/*
Copyright 2016 The Kubernetes Authors.

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

package linodego

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	klog "k8s.io/klog/v2"
)

// LKELinodeStatus constants reflect the current status of an node in a LKE Pool
const (
	LKELinodeReady    LKELinodeStatus = "ready"
	LKELinodeNotReady LKELinodeStatus = "not_ready"
)

// LKELinodeStatus represents the status of a node in a LKE Pool
type LKELinodeStatus string

// PageOptions are the pagination parameters for List endpoints
type PageOptions struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	Results int `json:"results"`
}

// ListOptions are the pagination and filtering parameters for endpoints
type ListOptions struct {
	*PageOptions
	PageSize int
	Filter   string
}

// LKEClusterPoolResponse is the struct for unmarshaling response from the list of LKE pools API call
type LKEClusterPoolResponse struct {
	Pools []LKEClusterPool `json:"data"`
	PageOptions
}

// LKEClusterPool represents a LKE Pool
type LKEClusterPool struct {
	ID      int                    `json:"id"`
	Count   int                    `json:"count"`
	Type    string                 `json:"type"`
	Disks   []LKEClusterPoolDisk   `json:"disks"`
	Linodes []LKEClusterPoolLinode `json:"nodes"`
}

// LKEClusterPoolDisk represents a node disk in an LKEClusterPool object
type LKEClusterPoolDisk struct {
	Size int    `json:"size"`
	Type string `json:"type"`
}

// LKEClusterPoolLinode represents a node in a LKE Pool
type LKEClusterPoolLinode struct {
	ID         string          `json:"id"`
	InstanceID int             `json:"instance_id"`
	Status     LKELinodeStatus `json:"status"`
}

// LKEClusterPoolCreateOptions fields are those accepted by CreateLKEClusterPool
type LKEClusterPoolCreateOptions struct {
	Count int                  `json:"count"`
	Type  string               `json:"type"`
	Disks []LKEClusterPoolDisk `json:"disks"`
}

// SetUserAgent sets a custom user-agent for HTTP requests
func (c *Client) SetUserAgent(ua string) *Client {
	c.userAgent = ua
	return c
}

// SetBaseURL sets the base URL of the Linode v4 API (https://api.linode.com/v4)
func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

// NewClient factory to create new Client struct
func NewClient(hc *http.Client) (client Client) {
	return Client{
		hc:        hc,
		baseURL:   "https://api.linode.com/v4",
		userAgent: "kubernetes/cluster-autoscaler",
	}
}

// Client is the struct to perform API calls
type Client struct {
	hc        *http.Client
	baseURL   string
	userAgent string
}

// request performs is used to perform a generic call to the Linode API and returns
// the body of the response, or error if such occurs or the response is not valid
func (c *Client) request(ctx context.Context, method, url string, jsonData []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			klog.Errorf("failed to close response body: %v", err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		return nil, fmt.Errorf("Unexpected Content-Type: %s with status: %s", ct, resp.Status)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}
	return nil, fmt.Errorf("%v %v: %d %v", req.Method, req.URL, resp.StatusCode, string(body))

}

// CreateLKEClusterPool creates a LKE Pool for for a LKE Cluster
func (c *Client) CreateLKEClusterPool(ctx context.Context, clusterID int, createOpts LKEClusterPoolCreateOptions) (*LKEClusterPool, error) {
	bodyReq, err := json.Marshal(createOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	url := fmt.Sprintf("%s/lke/clusters/%d/pools", c.baseURL, clusterID)
	bodyResp, err := c.request(ctx, "POST", url, bodyReq)
	if err != nil {
		return nil, err
	}
	newPool := &LKEClusterPool{}
	err = json.Unmarshal(bodyResp, newPool)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return newPool, nil
}

// DeleteLKEClusterPool deletes the LKE Pool with the specified id
func (c *Client) DeleteLKEClusterPool(ctx context.Context, clusterID, id int) error {
	url := fmt.Sprintf("%s/lke/clusters/%d/pools/%d", c.baseURL, clusterID, id)
	_, err := c.request(ctx, "DELETE", url, []byte{})
	return err
}

// listLKEClusterPoolsPaginated lists LKE Pools in a paginated request
// and the total number of pages the complete response is composed of
func (c *Client) listLKEClusterPoolsPaginated(ctx context.Context, clusterID int, opts *ListOptions, page int) ([]LKEClusterPool, int, error) {
	url := fmt.Sprintf("%s/lke/clusters/%d/pools?pages=%d", c.baseURL, clusterID, page)
	body, err := c.request(ctx, "GET", url, []byte{})
	if err != nil {
		return nil, 0, err
	}
	poolResp := &LKEClusterPoolResponse{}
	err = json.Unmarshal(body, poolResp)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return poolResp.Pools, poolResp.PageOptions.Pages, nil
}

// ListLKEClusterPools lists LKE Pools
func (c *Client) ListLKEClusterPools(ctx context.Context, clusterID int, opts *ListOptions) ([]LKEClusterPool, error) {
	// get the first response with the number of pages in it
	pools, pages, err := c.listLKEClusterPoolsPaginated(ctx, clusterID, opts, 1)
	if err != nil {
		return nil, err
	}
	// call again the API to get the results in the other pages
	for p := 2; p <= pages; p++ {
		poolsForPage, _, err := c.listLKEClusterPoolsPaginated(ctx, clusterID, opts, p)
		if err != nil {
			return nil, err
		}
		pools = append(pools, poolsForPage...)
	}
	return pools, nil
}
