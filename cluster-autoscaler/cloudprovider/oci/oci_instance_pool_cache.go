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
	"math"
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
	unownedInstances     map[OciRef]bool

	computeManagementClient ComputeMgmtClient
	computeClient           ComputeClient
	virtualNetworkClient    VirtualNetworkClient
}

func newInstancePoolCache(computeManagementClient ComputeMgmtClient, computeClient ComputeClient, virtualNetworkClient VirtualNetworkClient) *instancePoolCache {
	return &instancePoolCache{
		poolCache:               map[string]*core.InstancePool{},
		instanceSummaryCache:    map[string]*[]core.InstanceSummary{},
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
		klog.V(6).Infof("GetInstancePool() response %v", resp.InstancePool)

		c.setInstancePool(&resp.InstancePool)

		var instanceSummaries []core.InstanceSummary
		var page *string
		for {
			// OCI instance-pools do not contain individual instance objects so they must be fetched separately.
			listInstancePoolInstances, err := c.computeManagementClient.ListInstancePoolInstances(context.Background(), core.ListInstancePoolInstancesRequest{
				InstancePoolId: common.String(id),
				CompartmentId:  common.String(cfg.Global.CompartmentID),
				Page:           page,
			})
			if err != nil {
				return err
			}

			instanceSummaries = append(instanceSummaries, listInstancePoolInstances.Items...)

			if page = listInstancePoolInstances.OpcNextPage; listInstancePoolInstances.OpcNextPage == nil {
				break
			}
		}
		c.setInstanceSummaries(*resp.InstancePool.Id, &instanceSummaries)
	}

	// Reset unowned instances cache.
	c.unownedInstances = make(map[OciRef]bool)

	return nil
}

// removeInstance tries to remove the instance from the specified instance pool. If the instance isn't in the array,
// then it won't do anything removeInstance returns true if it actually removed the instance and reduced the size of
// the instance pool.
func (c *instancePoolCache) removeInstance(instancePool InstancePoolNodeGroup, instanceID string) bool {

	if instanceID == "" {
		klog.Warning("instanceID is not set - skipping removal.")
		return false
	}

	_, err := c.computeManagementClient.DetachInstancePoolInstance(context.Background(), core.DetachInstancePoolInstanceRequest{
		InstancePoolId: common.String(instancePool.Id()),
		DetachInstancePoolInstanceDetails: core.DetachInstancePoolInstanceDetails{
			InstanceId:      common.String(instanceID),
			IsDecrementSize: common.Bool(true),
			IsAutoTerminate: common.Bool(true),
		},
	})

	if err == nil {
		c.mu.Lock()
		// Decrease pool size in cache since IsDecrementSize was true
		c.poolCache[instancePool.Id()].Size = common.Int(*c.poolCache[instancePool.Id()].Size - 1)
		c.mu.Unlock()
		return true
	}

	klog.Errorf("error detaching instance %s from pool: %v", instanceID, err)
	return false
}

// findInstanceByDetails attempts to find the given instance by details by searching
// through the configured instance-pools (ListInstancePoolInstances) for a match.
func (c *instancePoolCache) findInstanceByDetails(ociInstance OciRef) (*OciRef, error) {

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
		// Skip searching instance pool if we happen tp know (prior labels) the pool ID and this is not it
		if (ociInstance.PoolID != "") && (ociInstance.PoolID != *nextInstancePool.Id) {
			klog.V(5).Infof("skipping over instance pool %s since it is not the one we are looking for", *nextInstancePool.Id)
			continue
		}

		var page *string
		var instanceSummaries []core.InstanceSummary
		for {
			// List instances in the next pool
			listInstancePoolInstancesReq := core.ListInstancePoolInstancesRequest{}
			listInstancePoolInstancesReq.CompartmentId = common.String(ociInstance.CompartmentID)
			listInstancePoolInstancesReq.InstancePoolId = nextInstancePool.Id
			listInstancePoolInstancesReq.Page = page

			listInstancePoolInstances, err := c.computeManagementClient.ListInstancePoolInstances(context.Background(), listInstancePoolInstancesReq)
			if err != nil {
				return nil, err
			}

			instanceSummaries = append(instanceSummaries, listInstancePoolInstances.Items...)

			if page = listInstancePoolInstances.OpcNextPage; listInstancePoolInstances.OpcNextPage == nil {
				break
			}
		}

		for _, poolMember := range instanceSummaries {
			// Skip comparing this instance if it is not in the Running state
			if strings.ToLower(*poolMember.State) != strings.ToLower(string(core.InstanceLifecycleStateRunning)) {
				klog.V(4).Infof("skipping over instance %s: since it is not in the running state: %s", *poolMember.Id, *poolMember.State)
				continue
			}
			// Skip this instance if we happen to know (prior labels) the instance ID and this is not it
			if (ociInstance.InstanceID != "") && (ociInstance.InstanceID != *poolMember.Id) {
				klog.V(5).Infof("skipping over instance %s since it is not the one we are looking for", *poolMember.Id)
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
			klog.V(6).Infof("ListVnicAttachments() response for %s: %v", *poolMember.Id, listVnicAttachments.Items)
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
				klog.V(6).Infof("GetVnic() response for vNIC %s: %v", *vnicAttachment.Id, getVnicResp.Vnic)
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
	c.poolCache[*np.Id].Size = np.Size
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

	scaleDelta := int(math.Abs(float64(*getInstancePoolResp.Size - size)))

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

	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, maxScalingWaitTime(scaleDelta, 20, 10*time.Minute))
	// Ensure this context is always canceled so channels, go routines, etc. always complete.
	defer cancelFunc()

	// Wait for the number of Running instances in this pool to reach size
	err = c.waitForRunningInstanceCount(ctx, size, instancePoolID, *getInstancePoolResp.CompartmentId)
	if err != nil {
		return err
	}

	// Allow an additional time for the pool State to reach Running
	ctx, _ = context.WithTimeout(ctx, 10*time.Minute)
	err = c.waitForState(ctx, instancePoolID, core.InstancePoolLifecycleStateRunning)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.poolCache[instancePoolID].Size = common.Int(size)
	c.mu.Unlock()

	return nil
}

func (c *instancePoolCache) waitForState(ctx context.Context, instancePoolID string, desiredState core.InstancePoolLifecycleStateEnum) error {
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
				deadline, _ := ctx.Deadline()
				klog.V(4).Infof("waiting for instance-pool %s to enter state: %s (current state: %s) (remaining time %v)",
					instancePoolID, desiredState, getInstancePoolResp.LifecycleState, deadline.Sub(time.Now()).Round(time.Second))
				return false, nil
			}
			klog.V(3).Infof("instance pool %s is in desired state: %s", instancePoolID, desiredState)

			return true, nil
		}, ctx.Done()) // context timeout
	if err != nil {
		// may be wait.ErrWaitTimeout
		return err
	}
	return nil
}

// waitForRunningInstanceCount waits for the number of instances in the instance pool reaches target.
func (c *instancePoolCache) waitForRunningInstanceCount(ctx context.Context, size int, instancePoolID, compartmentID string) error {

	progressCan, errChan := c.monitorScalingProgress(ctx, size, instancePoolID, compartmentID)
	for {
		// Reset progress timeout channel each time (any) progress message is received
		select {
		case _ = <-progressCan:
			// received a message on progress channel
		case err := <-errChan:
			// received a message on error channel
			if err != nil {
				klog.V(4).Infof("received an error while waiting for scale to complete on %s: %v", instancePoolID, err)
				return err
			}
			return nil
		case <-time.After(10 * time.Minute):
			// timeout waiting for completion or update in count
			return errors.New("timeout waiting for instance-pool " + instancePoolID + " scaling operation to make progress")
		}
	}
}

// monitorScalingProgress monitors the progress of the scaling operation of instancePoolID and sends incremental changes
// to the number of running instances to the int channel and errors to the error channel.
func (c *instancePoolCache) monitorScalingProgress(ctx context.Context, target int, instancePoolID, compartmentID string) (<-chan int, <-chan error) {

	errCh := make(chan error, 1)
	progressCh := make(chan int)

	go func() {
		defer close(errCh)
		defer close(progressCh)

		previousNumInstances := 0
		ticker := time.NewTicker(internalPollInterval)
		sendProgress := func(p int) {
			select {
			case progressCh <- p:
				// Send progress
				break
			default:
				break
			}
		}

		for {
			select {
			case <-ticker.C:
				// ticker fired, recheck
				break
			case <-ctx.Done():
				// max/context timeout
				errCh <- ctx.Err()
				return
			}

			var page *string
			numRunningInstances := 0
			for {
				// List instances in the pool
				listInstancePoolInstances, err := c.computeManagementClient.ListInstancePoolInstances(context.Background(), core.ListInstancePoolInstancesRequest{
					InstancePoolId: common.String(instancePoolID),
					CompartmentId:  common.String(compartmentID),
					Page:           page,
				})
				if err != nil {
					klog.Errorf("list instance pool instances for pool %s failed: %v", instancePoolID, err)
					errCh <- err
					return
				}

				for _, poolMember := range listInstancePoolInstances.Items {
					if strings.ToLower(*poolMember.State) == strings.ToLower(string(core.InstanceLifecycleStateRunning)) {
						numRunningInstances++
					}
				}

				if page = listInstancePoolInstances.OpcNextPage; listInstancePoolInstances.OpcNextPage == nil {
					break
				}
			}

			if previousNumInstances == 0 {
				previousNumInstances = numRunningInstances
			}

			if numRunningInstances == target {
				klog.V(4).Infof("running instances in %s has reached the target of %d", instancePoolID, target)
				sendProgress(target)
				errCh <- nil
				// done
				return
			} else if previousNumInstances != numRunningInstances {
				deadline, _ := ctx.Deadline()
				klog.V(4).Infof("running instances in %s has not yet reached the target of %d (current count: %d) "+
					"(remaining time %v)", instancePoolID, target, numRunningInstances,
					deadline.Sub(time.Now()).Round(time.Second))
				sendProgress(numRunningInstances)
				previousNumInstances = numRunningInstances
				// continue
			}
		}
	}()

	return progressCh, errCh
}

func (c *instancePoolCache) getSize(id string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pool, ok := c.poolCache[id]
	if !ok {
		return -1, errors.New("target size not found")
	}

	return *pool.Size, nil
}

// maxScalingWaitTime estimates the maximum amount of time, as a duration, that to scale size instances.
// note, larger scale operations are broken up internally to smaller batches. This is an internal detail
// and can be overridden on a tenancy basis. 20 is a good default.
func maxScalingWaitTime(size, batchSize int, timePerBatch time.Duration) time.Duration {
	buffer := 60 * time.Second

	if size <= batchSize {
		return timePerBatch + buffer
	}

	maxScalingWaitTime := (timePerBatch) * time.Duration(size/batchSize)

	// add additional batch for any remainders
	if size%batchSize > 0 {
		maxScalingWaitTime = maxScalingWaitTime + timePerBatch
	}

	return maxScalingWaitTime + buffer
}
