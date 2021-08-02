/*
Copyright 2021 Oracle and/or its affiliates.

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

package oci

import (
	"context"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/core"
	"k8s.io/klog/v2"
	"strings"
	"sync"
	"time"
)

// ComputeMgmtClient wraps core.ComputeManagementClient exposing the functions we actually require.
type ComputeMgmtClient interface {
	GetInstancePool(context.Context, core.GetInstancePoolRequest) (core.GetInstancePoolResponse, error)
	UpdateInstancePool(context.Context, core.UpdateInstancePoolRequest) (core.UpdateInstancePoolResponse, error)
	GetInstancePoolInstance(context.Context, core.GetInstancePoolInstanceRequest) (core.GetInstancePoolInstanceResponse, error)
	ListInstancePoolInstances(context.Context, core.ListInstancePoolInstancesRequest) (core.ListInstancePoolInstancesResponse, error)
	DetachInstancePoolInstance(context.Context, core.DetachInstancePoolInstanceRequest) (core.DetachInstancePoolInstanceResponse, error)
}

// ComputeClient wraps core.ComputeClient exposing the functions we actually require.
type ComputeClient interface {
	ListVnicAttachments(ctx context.Context, request core.ListVnicAttachmentsRequest) (core.ListVnicAttachmentsResponse, error)
}

// VirtualNetworkClient wraps core.VirtualNetworkClient exposing the functions we actually require.
type VirtualNetworkClient interface {
	GetVnic(context.Context, core.GetVnicRequest) (core.GetVnicResponse, error)
}

type instancePoolCache struct {
	mu                   sync.Mutex
	poolCache            map[string]*core.InstancePool
	instanceSummaryCache map[string]*[]core.InstanceSummary
	targetSize           map[string]int
	unownedInstances     map[OciRef]bool

	computeManagementClient ComputeMgmtClient
	computeClient           ComputeClient
	virtualNetworkClient    VirtualNetworkClient
}

func newInstancePoolCache(computeManagementClient ComputeMgmtClient, computeClient ComputeClient, virtualNetworkClient VirtualNetworkClient) *instancePoolCache {
	return &instancePoolCache{
		poolCache:               map[string]*core.InstancePool{},
		instanceSummaryCache:    map[string]*[]core.InstanceSummary{},
		targetSize:              map[string]int{},
		unownedInstances:        map[OciRef]bool{},
		computeManagementClient: computeManagementClient,
		computeClient:           computeClient,
		virtualNetworkClient:    virtualNetworkClient,
	}
}

func (c *instancePoolCache) InstancePools() map[string]*core.InstancePool {
	result := map[string]*core.InstancePool{}
	for k, v := range c.poolCache {
		result[k] = v
	}
	return result
}

func (c *instancePoolCache) rebuild(staticInstancePools map[string]*InstancePoolNodeGroup, cfg CloudConfig) error {
	// Since we only support static instance-pools we don't need to worry about pruning.

	for id := range staticInstancePools {
		resp, err := c.computeManagementClient.GetInstancePool(context.Background(), core.GetInstancePoolRequest{
			InstancePoolId: common.String(id),
		})
		if err != nil {
			klog.Errorf("get instance pool %s failed: %v", id, err)
			return err
		}
		klog.V(5).Infof("GetInstancePool() response %v", resp.InstancePool)

		c.setInstancePool(&resp.InstancePool)

		// OCI instance-pools do not contain individual instance objects so they must be fetched separately.
		listInstancesResponse, err := c.computeManagementClient.ListInstancePoolInstances(context.Background(), core.ListInstancePoolInstancesRequest{
			InstancePoolId: common.String(id),
			CompartmentId:  common.String(cfg.Global.CompartmentID),
		})
		if err != nil {
			return err
		}
		klog.V(5).Infof("ListInstancePoolInstances() response %v", listInstancesResponse.Items)
		c.setInstanceSummaries(*resp.InstancePool.Id, &listInstancesResponse.Items)
	}
	// Reset unowned instances cache.
	c.unownedInstances = make(map[OciRef]bool)

	return nil
}

// removeInstance tries to remove the instance from the specified instance pool. If the instance isn't in the array,
// then it won't do anything removeInstance returns true if it actually removed the instance and reduced the size of
// the instance pool.
func (c *instancePoolCache) removeInstance(instancePool InstancePoolNodeGroup, instanceID string) bool {

	c.mu.Lock()
	defer c.mu.Unlock()

	if instanceID == "" {
		klog.Warning("instanceID is not set - skipping removal.")
		return false
	}

	// This instance pool must be in state RUNNING in order to detach a particular instance.
	err := c.waitForInstancePoolState(context.Background(), instancePool.Id(), core.InstancePoolLifecycleStateRunning)
	if err != nil {
		return false
	}

	_, err = c.computeManagementClient.DetachInstancePoolInstance(context.Background(), core.DetachInstancePoolInstanceRequest{
		InstancePoolId: common.String(instancePool.Id()),
		DetachInstancePoolInstanceDetails: core.DetachInstancePoolInstanceDetails{
			InstanceId:      common.String(instanceID),
			IsDecrementSize: common.Bool(true),
			IsAutoTerminate: common.Bool(true),
		},
	})

	if err == nil {
		// Decrease pool size in cache since IsDecrementSize was true
		c.targetSize[instancePool.Id()] -= 1
		return true
	}

	return false
}

// findInstanceByDetails attempts to find the given instance by details by searching
// through the configured instance-pools (ListInstancePoolInstances) for a match.
func (c *instancePoolCache) findInstanceByDetails(ociInstance OciRef) (*OciRef, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Minimum amount of information we need to make a positive match
	if ociInstance.InstanceID == "" && ociInstance.PrivateIPAddress == "" && ociInstance.PublicIPAddress == "" {
		return nil, errors.New("instance id or an IP address is required to resolve details")
	}

	if c.unownedInstances[ociInstance] {
		// We already know this instance is not part of a configured pool. Return early and avoid additional API calls.
		klog.V(4).Infof("Node " + ociInstance.Name + " is known to not be a member of any of the specified instance pool(s)")
		return nil, errInstanceInstancePoolNotFound
	}

	// Look for the instance in each of the specified pool(s)
	for _, nextInstancePool := range c.poolCache {
		// Skip searching instance pool if it's instance count is 0.
		if *nextInstancePool.Size == 0 {
			klog.V(4).Infof("skipping over instance pool %s since it is empty", *nextInstancePool.Id)
			continue
		}
		// List instances in the next pool
		listInstancePoolInstances, err := c.computeManagementClient.ListInstancePoolInstances(context.Background(), core.ListInstancePoolInstancesRequest{
			CompartmentId:  common.String(ociInstance.CompartmentID),
			InstancePoolId: nextInstancePool.Id,
		})
		if err != nil {
			return nil, err
		}

		for _, poolMember := range listInstancePoolInstances.Items {
			// Skip comparing this instance if it is not in the Running state
			if strings.ToLower(*poolMember.State) != strings.ToLower(string(core.InstanceLifecycleStateRunning)) {
				klog.V(4).Infof("skipping over instance %s: since it is not in the running state: %s", *poolMember.Id, *poolMember.State)
				continue
			}
			listVnicAttachments, err := c.computeClient.ListVnicAttachments(context.Background(), core.ListVnicAttachmentsRequest{
				CompartmentId: common.String(*poolMember.CompartmentId),
				InstanceId:    poolMember.Id,
			})
			if err != nil {
				klog.Errorf("list vNIC attachments for %s failed: %v", *poolMember.Id, err)
				return nil, err
			}
			klog.V(5).Infof("ListVnicAttachments() response for %s: %v", *poolMember.Id, listVnicAttachments.Items)
			for _, vnicAttachment := range listVnicAttachments.Items {
				// Skip this attachment if the vNIC is not live
				if core.VnicAttachmentLifecycleStateAttached != vnicAttachment.LifecycleState {
					klog.V(4).Infof("skipping vNIC on instance %s: since it is not active", *poolMember.Id)
					continue
				}
				getVnicResp, err := c.virtualNetworkClient.GetVnic(context.Background(), core.GetVnicRequest{
					VnicId: vnicAttachment.VnicId,
				})
				if err != nil {
					klog.Errorf("get vNIC for %s failed: %v", *poolMember.Id, err)
					return nil, err
				}
				klog.V(5).Infof("GetVnic() response for vNIC %s: %v", *vnicAttachment.Id, getVnicResp.Vnic)
				// Preferably we match by instanceID, but we can match by private or public IP
				if *poolMember.Id == ociInstance.InstanceID ||
					(getVnicResp.Vnic.PrivateIp != nil && *getVnicResp.Vnic.PrivateIp == ociInstance.PrivateIPAddress) ||
					(getVnicResp.Vnic.PublicIp != nil && *getVnicResp.Vnic.PublicIp == ociInstance.PublicIPAddress) {
					klog.V(4).Info(*poolMember.DisplayName, " is a member of "+*nextInstancePool.Id)
					// Return a complete instance details.
					if ociInstance.Name == "" {
						ociInstance.Name = *poolMember.DisplayName
					}
					ociInstance.InstanceID = *poolMember.Id
					ociInstance.PoolID = *nextInstancePool.Id
					ociInstance.CompartmentID = *poolMember.CompartmentId
					ociInstance.AvailabilityDomain = strings.Split(*poolMember.AvailabilityDomain, ":")[1]
					ociInstance.Shape = *poolMember.Shape
					ociInstance.PrivateIPAddress = *getVnicResp.Vnic.PrivateIp
					// Public IP is optional
					if getVnicResp.Vnic.PublicIp != nil {
						ociInstance.PublicIPAddress = *getVnicResp.Vnic.PublicIp
					}
					return &ociInstance, nil
				}
			}
		}
	}

	c.unownedInstances[ociInstance] = true
	klog.V(4).Infof(ociInstance.Name + " is not a member of any of the specified instance pool(s)")
	return nil, errInstanceInstancePoolNotFound
}

func (c *instancePoolCache) getInstancePool(id string) (*core.InstancePool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.getInstancePoolWithoutLock(id)
}

func (c *instancePoolCache) getInstancePoolWithoutLock(id string) (*core.InstancePool, error) {
	instancePool := c.poolCache[id]
	if instancePool == nil {
		return nil, errors.New("instance pool was not found in the cache")
	}

	return instancePool, nil
}

func (c *instancePoolCache) setInstancePool(np *core.InstancePool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.poolCache[*np.Id] = np
	c.targetSize[*np.Id] = *np.Size
}

func (c *instancePoolCache) getInstanceSummaries(poolID string) (*[]core.InstanceSummary, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.getInstanceSummariesWithoutLock(poolID)
}

func (c *instancePoolCache) getInstanceSummariesWithoutLock(poolID string) (*[]core.InstanceSummary, error) {
	instanceSummaries := c.instanceSummaryCache[poolID]
	if instanceSummaries == nil {
		return nil, errors.New("instance summaries for instance pool id " + poolID + " were not found in cache")
	}

	return instanceSummaries, nil
}

func (c *instancePoolCache) setInstanceSummaries(instancePoolID string, is *[]core.InstanceSummary) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.instanceSummaryCache[instancePoolID] = is
}

func (c *instancePoolCache) setSize(instancePoolID string, size int) error {

	if instancePoolID == "" {
		return errors.New("instance-pool is required")
	}

	getInstancePoolResp, err := c.computeManagementClient.GetInstancePool(context.Background(), core.GetInstancePoolRequest{
		InstancePoolId: common.String(instancePoolID),
	})
	if err != nil {
		return err
	}

	updateDetails := core.UpdateInstancePoolDetails{
		Size:                    common.Int(size),
		InstanceConfigurationId: getInstancePoolResp.InstanceConfigurationId,
	}

	_, err = c.computeManagementClient.UpdateInstancePool(context.Background(), core.UpdateInstancePoolRequest{
		InstancePoolId:            common.String(instancePoolID),
		UpdateInstancePoolDetails: updateDetails,
	})
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.targetSize[instancePoolID] = size

	return c.waitForInstancePoolState(context.Background(), instancePoolID, core.InstancePoolLifecycleStateRunning)
}

func (c *instancePoolCache) waitForInstancePoolState(ctx context.Context, instancePoolID string, desiredState core.InstancePoolLifecycleStateEnum) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	err := wait.PollImmediateUntil(
		// TODO we need a better implementation of this function
		internalPollInterval,
		func() (bool, error) {
			getInstancePoolResp, err := c.computeManagementClient.GetInstancePool(context.Background(), core.GetInstancePoolRequest{
				InstancePoolId: common.String(instancePoolID),
			})
			if err != nil {
				klog.Errorf("getInstancePool failed. Retrying: %+v", err)
				return false, err
			} else if getInstancePoolResp.LifecycleState != desiredState {
				klog.V(4).Infof("waiting for instance-pool %s to enter state: %s (current state: %s)", instancePoolID,
					desiredState, getInstancePoolResp.LifecycleState)
				return false, nil
			}
			klog.V(3).Infof("instance pool %s is in desired state: %s", instancePoolID, desiredState)

			return true, nil
		}, timeoutCtx.Done())
	if err != nil {
		// may be wait.ErrWaitTimeout
		return err
	}
	return nil
}

func (c *instancePoolCache) getSize(id string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	size, ok := c.targetSize[id]
	if !ok {
		return -1, errors.New("target size not found")
	}

	return size, nil
}
