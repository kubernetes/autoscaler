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

package civocloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// KubernetesClusterPoolUpdateConfig is used to create a new cluster pool
type KubernetesClusterPoolUpdateConfig struct {
	ID     string `json:"id,omitempty"`
	Count  int    `json:"count,omitempty"`
	Size   string `json:"size,omitempty"`
	Region string `json:"region,omitempty"`
}

// ListKubernetesClusterPools returns all the pools for a kubernetes cluster
func (c *Client) ListKubernetesClusterPools(cid string) ([]KubernetesPool, error) {
	resp, err := c.SendGetRequest(fmt.Sprintf("/v2/kubernetes/clusters/%s/pools", cid))
	if err != nil {
		return nil, decodeError(err)
	}

	pools := make([]KubernetesPool, 0)
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&pools); err != nil {
		return nil, decodeError(err)
	}

	return pools, nil
}

// GetKubernetesClusterPool returns a pool for a kubernetes cluster
func (c *Client) GetKubernetesClusterPool(cid, pid string) (*KubernetesPool, error) {
	resp, err := c.SendGetRequest(fmt.Sprintf("/v2/kubernetes/clusters/%s/pools/%s", cid, pid))
	if err != nil {
		return nil, decodeError(err)
	}

	pool := &KubernetesPool{}
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&pool); err != nil {
		return nil, decodeError(err)
	}

	return pool, nil
}

// FindKubernetesClusterPool finds a pool by either part of the ID
func (c *Client) FindKubernetesClusterPool(cid, search string) (*KubernetesPool, error) {
	pools, err := c.ListKubernetesClusterPools(cid)
	if err != nil {
		return nil, decodeError(err)
	}

	exactMatch := false
	partialMatchesCount := 0
	result := KubernetesPool{}

	for _, value := range pools {
		if value.ID == search {
			exactMatch = true
			result = value
		} else if strings.Contains(value.ID, search) {
			if !exactMatch {
				result = value
				partialMatchesCount++
			}
		}
	}

	if exactMatch || partialMatchesCount == 1 {
		return &result, nil
	} else if partialMatchesCount > 1 {
		err := fmt.Errorf("unable to find %s because there were multiple matches", search)
		return nil, MultipleMatchesError.wrap(err)
	} else {
		err := fmt.Errorf("unable to find %s, zero matches", search)
		return nil, ZeroMatchesError.wrap(err)
	}
}

// DeleteKubernetesClusterPoolInstance deletes a instance from pool
func (c *Client) DeleteKubernetesClusterPoolInstance(cid, pid, id string) (*SimpleResponse, error) {
	resp, err := c.SendDeleteRequest(fmt.Sprintf("/v2/kubernetes/clusters/%s/pools/%s/instances/%s", cid, pid, id))
	if err != nil {
		return nil, decodeError(err)
	}

	return c.DecodeSimpleResponse(resp)
}

// UpdateKubernetesClusterPool updates a pool for a kubernetes cluster
func (c *Client) UpdateKubernetesClusterPool(cid, pid string, config *KubernetesClusterPoolUpdateConfig) (*KubernetesPool, error) {
	resp, err := c.SendPutRequest(fmt.Sprintf("/v2/kubernetes/clusters/%s/pools/%s", cid, pid), config)
	if err != nil {
		return nil, decodeError(err)
	}

	pool := &KubernetesPool{}
	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&pool); err != nil {
		return nil, decodeError(err)
	}

	return pool, nil
}
