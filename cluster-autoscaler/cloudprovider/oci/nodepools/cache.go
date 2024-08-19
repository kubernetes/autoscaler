/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"context"
	"net/http"
	"sync"

	"k8s.io/klog/v2"

	"github.com/pkg/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
)

func newNodePoolCache(okeClient *oke.ContainerEngineClient) *nodePoolCache {
	return &nodePoolCache{
		cache:      map[string]*oke.NodePool{},
		targetSize: map[string]int{},
		okeClient:  okeClient,
	}
}

type nodePoolCache struct {
	mu         sync.Mutex
	cache      map[string]*oke.NodePool
	targetSize map[string]int

	okeClient okeClient
}

func (c *nodePoolCache) nodePools() map[string]*oke.NodePool {
	result := map[string]*oke.NodePool{}
	for k, v := range c.cache {
		result[k] = v
	}
	return result
}

func (c *nodePoolCache) rebuild(staticNodePools map[string]NodePool, maxGetNodepoolRetries int) (httpStatusCode int, err error) {
	klog.Infof("rebuilding cache")
	var statusCode int
	for id := range staticNodePools {
		var resp oke.GetNodePoolResponse
		for i := 1; i <= maxGetNodepoolRetries; i++ {
			// prevent us from getting a node pool at the same time that we're performing delete actions on the node pool.
			c.mu.Lock()
			resp, err = c.okeClient.GetNodePool(context.Background(), oke.GetNodePoolRequest{
				NodePoolId: common.String(id),
			})
			c.mu.Unlock()
			httpResp := resp.HTTPResponse()
			statusCode = httpResp.StatusCode
			if err != nil {
				klog.Errorf("Failed to fetch the nodepool : %v. Retries available : %v", id, maxGetNodepoolRetries-i)
			} else {
				break
			}
		}
		if err != nil {
			klog.Errorf("Failed to fetch the nodepool : %v", id)
			return statusCode, err
		}
		c.set(&resp.NodePool)
	}
	return statusCode, nil
}

// removeInstance tries to remove the instance from the node pool.
func (c *nodePoolCache) removeInstance(nodePoolID, instanceID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	klog.Infof("Deleting instance %q from node pool %q", instanceID, nodePoolID)
	// always try to remove the instance. This call is idempotent
	scaleDown := true
	resp, err := c.okeClient.DeleteNode(context.Background(), oke.DeleteNodeRequest{
		NodePoolId:      &nodePoolID,
		NodeId:          &instanceID,
		IsDecrementSize: &scaleDown,
	})

	klog.Infof("Delete Node API returned response: %v, err: %v", resp, err)
	httpResp := resp.HTTPResponse()
	var success bool
	if httpResp != nil {
		statusCode := httpResp.StatusCode
		// status returned should be a 202, but let's accept any 2XX codes anyway
		statusSuccess := statusCode >= 200 && statusCode < 300
		success = statusSuccess ||
			// 409 means the instance is already going to be processed for deletion
			statusCode == http.StatusConflict ||
			// 404 means it is probably already deleted and our cache may be stale
			statusCode == http.StatusNotFound
		if !success {
			status := httpResp.Status
			klog.Infof("Received error status %s while deleting node %q", status, instanceID)

			// statuses that we might expect but are still errors:
			// 400s (if cluster still uses TA or is v1 based)
			// 401 unauthorized
			// 412 etag mismatch
			// 429 too many requests
			// 500 internal server errors
			return errors.Errorf("received error status %s while deleting node %q", status, instanceID)
		} else if statusSuccess {
			// since delete node endpoint scales down by 1, we need to update the cache's target size by -1 too
			c.targetSize[nodePoolID]--
		}
	}

	if !success && err != nil {
		return err
	}

	nodePool := c.cache[nodePoolID]
	// theoretical max number of nodes inside a cluster is 1000
	// so at most we'll be copying 1000 nodes
	newNodeSlice := make([]oke.Node, 0, len(nodePool.Nodes))
	for _, node := range nodePool.Nodes {
		if *node.Id != instanceID {
			newNodeSlice = append(newNodeSlice, node)
		} else {
			klog.Infof("Deleting instance %q from cache", instanceID)
		}
	}
	nodePool.Nodes = newNodeSlice

	return nil
}

func (c *nodePoolCache) getByInstance(instanceID string) (*oke.NodePool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, nodePool := range c.cache {
		for _, node := range nodePool.Nodes {
			if *node.Id == instanceID {
				return nodePool, nil
			}
		}
	}

	return nil, errors.New("node pool not found for node in cache")
}

func (c *nodePoolCache) get(id string) (*oke.NodePool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.getWithoutLock(id)
}

func (c *nodePoolCache) getWithoutLock(id string) (*oke.NodePool, error) {
	nodePool := c.cache[id]
	if nodePool == nil {
		return nil, errors.New("node pool was not found in cache")
	}

	return nodePool, nil
}

func (c *nodePoolCache) set(np *oke.NodePool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[*np.Id] = np
	c.targetSize[*np.Id] = *np.NodeConfigDetails.Size
}

func (c *nodePoolCache) setSize(id string, size int) error {

	_, err := c.okeClient.UpdateNodePool(context.Background(), oke.UpdateNodePoolRequest{
		NodePoolId: common.String(id),
		UpdateNodePoolDetails: oke.UpdateNodePoolDetails{
			NodeConfigDetails: &oke.UpdateNodePoolNodeConfigDetails{
				Size: common.Int(size),
			},
		},
	})
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.targetSize[id] = size
	return nil
}

func (c *nodePoolCache) getSize(id string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	size, ok := c.targetSize[id]
	if !ok {
		return -1, errors.New("target size not found")
	}

	return size, nil
}
