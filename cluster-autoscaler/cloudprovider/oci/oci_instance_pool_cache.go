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
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/core"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/workrequests"
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

// WorkRequestClient wraps workrequests.WorkRequestClient exposing the functions we actually require.
type WorkRequestClient interface {
	GetWorkRequest(context.Context, workrequests.GetWorkRequestRequest) (workrequests.GetWorkRequestResponse, error)
	ListWorkRequests(context.Context, workrequests.ListWorkRequestsRequest) (workrequests.ListWorkRequestsResponse, error)
	ListWorkRequestErrors(context.Context, workrequests.ListWorkRequestErrorsRequest) (workrequests.ListWorkRequestErrorsResponse, error)
}

type instancePoolCache struct {
	mu                   sync.Mutex
	poolCache            map[string]*core.InstancePool
	instanceSummaryCache map[string]*[]core.InstanceSummary
	unownedInstances     map[OciRef]bool

	computeManagementClient ComputeMgmtClient
	computeClient           ComputeClient
	virtualNetworkClient    VirtualNetworkClient
	workRequestsClient      WorkRequestClient
}

func newInstancePoolCache(computeManagementClient ComputeMgmtClient, computeClient ComputeClient, virtualNetworkClient VirtualNetworkClient, workRequestsClient WorkRequestClient) *instancePoolCache {
	return &instancePoolCache{
		poolCache:               map[string]*core.InstancePool{},
		instanceSummaryCache:    map[string]*[]core.InstanceSummary{},
		unownedInstances:        map[OciRef]bool{},
		computeManagementClient: computeManagementClient,
		computeClient:           computeClient,
		virtualNetworkClient:    virtualNetworkClient,
		workRequestsClient:      workRequestsClient,
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
		getInstancePoolResp, err := c.computeManagementClient.GetInstancePool(context.Background(), core.GetInstancePoolRequest{
			InstancePoolId: common.String(id),
		})
		if err != nil {
			klog.Errorf("get instance pool %s failed: %v", id, err)
			return err
		}
		klog.V(6).Infof("GetInstancePool() response %v", getInstancePoolResp.InstancePool)

		c.setInstancePool(&getInstancePoolResp.InstancePool)

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
		c.setInstanceSummaries(id, &instanceSummaries)
		// Compare instance pool's size with the latest number of InstanceSummaries. If found, look for unrecoverable
		// errors such as quota or capacity issues in scaling pool.
		if len(*c.instanceSummaryCache[id]) < *c.poolCache[id].Size {
			klog.V(4).Infof("Instance pool %s has only %d instances created while requested count is %d. ",
				*getInstancePoolResp.InstancePool.DisplayName, len(*c.instanceSummaryCache[id]), *c.poolCache[id].Size)

			if getInstancePoolResp.LifecycleState != core.InstancePoolLifecycleStateRunning {
				lastWorkRequest, err := c.lastStartedWorkRequest(*getInstancePoolResp.CompartmentId, id)

				// The last started work request may be many minutes old depending on sync interval
				// and exponential backoff time of OCI retried OCI operations.
				if err == nil && *lastWorkRequest.OperationType == ociInstancePoolLaunchOp &&
					lastWorkRequest.Status == workrequests.WorkRequestSummaryStatusFailed {
					unrecoverableErrorMsg := c.firstUnrecoverableErrorForWorkRequest(*lastWorkRequest.Id)
					if unrecoverableErrorMsg != "" {
						klog.V(4).Infof("Creating placeholder instances for %s.", *getInstancePoolResp.InstancePool.DisplayName)
						for i := len(*c.instanceSummaryCache[id]); i < *c.poolCache[id].Size; i++ {
							c.addUnfulfilledInstanceToCache(id, fmt.Sprintf("%s%s-%d", instanceIDUnfulfilled,
								*getInstancePoolResp.InstancePool.Id, i), *getInstancePoolResp.InstancePool.CompartmentId,
								fmt.Sprintf("%s-%d", *getInstancePoolResp.InstancePool.DisplayName, i))
						}
					}
				}
			}
		}
	}

	// Reset unowned instances cache.
	c.unownedInstances = make(map[OciRef]bool)

	return nil
}

func (c *instancePoolCache) addUnfulfilledInstanceToCache(instancePoolID, instanceID, compartmentID, name string) {
	*c.instanceSummaryCache[instancePoolID] = append(*c.instanceSummaryCache[instancePoolID], core.InstanceSummary{
		Id:            common.String(instanceID),
		CompartmentId: common.String(compartmentID),
		State:         common.String(instanceStateUnfulfilled),
		DisplayName:   common.String(name),
	})
}

// removeInstance tries to remove the instance from the specified instance pool. If the instance isn't in the array,
// then it won't do anything removeInstance returns true if it actually removed the instance and reduced the size of
// the instance pool.
func (c *instancePoolCache) removeInstance(instancePool InstancePoolNodeGroup, instanceID string) bool {

	if instanceID == "" {
		klog.Warning("instanceID is not set - skipping removal.")
		return false
	}

	var err error
	if strings.Contains(instanceID, instanceIDUnfulfilled) {
		// For an unfulfilled instance, reduce the target size of the instance pool and remove the placeholder instance from cache.
		err = c.setSize(instancePool.Id(), *c.poolCache[instancePool.Id()].Size-1)
	} else {
		_, err = c.computeManagementClient.DetachInstancePoolInstance(context.Background(), core.DetachInstancePoolInstanceRequest{
			InstancePoolId: common.String(instancePool.Id()),
			DetachInstancePoolInstanceDetails: core.DetachInstancePoolInstanceDetails{
				InstanceId:      common.String(instanceID),
				IsDecrementSize: common.Bool(true),
				IsAutoTerminate: common.Bool(true),
			},
		})
	}

	if err == nil {
		c.mu.Lock()
		// Decrease pool size in cache
		c.poolCache[instancePool.Id()].Size = common.Int(*c.poolCache[instancePool.Id()].Size - 1)
		// Since we're removing the instance from cache, we don't need to expire the pool cache
		c.removeInstanceSummaryFromCache(instancePool.Id(), instanceID)
		c.mu.Unlock()
		return true
	}

	klog.Errorf("error detaching instance %s from pool: %v", instanceID, err)
	return false
}

// findInstanceByDetails attempts to find the given instance by details by searching
// through the configured instance-pools (ListInstancePoolInstances) for a match.
func (c *instancePoolCache) findInstanceByDetails(ociInstance OciRef) (*OciRef, error) {

	// Unfilled instance placeholder
	if strings.Contains(ociInstance.Name, instanceIDUnfulfilled) {
		instIndex := strings.LastIndex(ociInstance.Name, "-")
		ociInstance.PoolID = strings.Replace(ociInstance.Name[:instIndex], instanceIDUnfulfilled, "", 1)
		return &ociInstance, nil
	}
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

	isScaleUp := size > *getInstancePoolResp.Size
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

	c.mu.Lock()
	c.poolCache[instancePoolID].Size = common.Int(size)
	c.mu.Unlock()

	// Just return Immediately if this was a scale down to be consistent with DetachInstancePoolInstance
	if !isScaleUp {
		return nil
	}

	// Only wait for scale up (not scale down)
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

			// Fail scale (up) operation fast by watching for unrecoverable errors such as quota or capacity issues
			lastWorkRequest, err := c.lastStartedWorkRequest(compartmentID, instancePoolID)
			if err == nil && *lastWorkRequest.OperationType == ociInstancePoolLaunchOp &&
				lastWorkRequest.Status == workrequests.WorkRequestSummaryStatusInProgress {
				unrecoverableErrorMsg := c.firstUnrecoverableErrorForWorkRequest(*lastWorkRequest.Id)
				if unrecoverableErrorMsg != "" {
					errCh <- errors.New(unrecoverableErrorMsg)
					return
				}
			}

			var page *string
			numRunningInstances := 0
			for {
				// Next, wait until the number of instances in the pool reaches the target
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

// removeInstanceSummaryFromCache removes looks through the pool cache for an InstanceSummary with the specified ID and
// removes it if found
func (c *instancePoolCache) removeInstanceSummaryFromCache(instancePoolID, instanceID string) {
	var instanceSummaries []core.InstanceSummary

	if instanceSummaryCache, found := c.instanceSummaryCache[instancePoolID]; found {
		for _, instanceSummary := range *instanceSummaryCache {
			if instanceSummary.Id != nil && *instanceSummary.Id != instanceID {
				instanceSummaries = append(instanceSummaries, instanceSummary)
			}
		}
		c.instanceSummaryCache[instancePoolID] = &instanceSummaries
	}
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

// lastStartedWorkRequest returns the *last started* work request for the specified resource or an error if none are found
func (c *instancePoolCache) lastStartedWorkRequest(compartmentID, resourceID string) (workrequests.WorkRequestSummary, error) {

	klog.V(6).Infof("Looking for the last started work request for resource %s.", resourceID)
	listWorkRequests, err := c.workRequestsClient.ListWorkRequests(context.Background(), workrequests.ListWorkRequestsRequest{
		CompartmentId: common.String(compartmentID),
		Limit:         common.Int(100),
		ResourceId:    common.String(resourceID),
	})
	if err != nil {
		klog.Errorf("list work requests for %s failed: %v", resourceID, err)
		return workrequests.WorkRequestSummary{}, err
	}

	var lastStartedWorkRequest = workrequests.WorkRequestSummary{}
	for i, nextWorkRequest := range listWorkRequests.Items {
		if i == 0 && nextWorkRequest.TimeStarted != nil {
			lastStartedWorkRequest = nextWorkRequest
		} else {
			if nextWorkRequest.TimeStarted != nil && nextWorkRequest.TimeStarted.After(lastStartedWorkRequest.TimeStarted.Time) {
				lastStartedWorkRequest = nextWorkRequest
			}
		}
	}

	if lastStartedWorkRequest.TimeStarted != nil {
		return lastStartedWorkRequest, nil
	}

	return workrequests.WorkRequestSummary{}, errors.New("no work requests found")
}

// firstUnrecoverableErrorForWorkRequest returns the first non-recoverable error message associated with the specified
// work-request ID, or the empty string if none are found.
func (c *instancePoolCache) firstUnrecoverableErrorForWorkRequest(workRequestID string) string {

	klog.V(6).Infof("Looking for non-recoverable errors for work request %s.", workRequestID)
	// Look through the error logs looking for known unrecoverable error messages(s)
	workRequestErrors, _ := c.workRequestsClient.ListWorkRequestErrors(context.Background(),
		workrequests.ListWorkRequestErrorsRequest{WorkRequestId: common.String(workRequestID),
			SortOrder: workrequests.ListWorkRequestErrorsSortOrderDesc})
	for _, nextErr := range workRequestErrors.Items {
		// Abort wait for certain unrecoverable errors such as capacity and quota issues
		if strings.Contains(strings.ToLower(*nextErr.Message), strings.ToLower("QuotaExceeded")) ||
			strings.Contains(strings.ToLower(*nextErr.Message), strings.ToLower("LimitExceeded")) ||
			strings.Contains(strings.ToLower(*nextErr.Message), strings.ToLower("OutOfCapacity")) {
			klog.V(4).Infof("Found unrecoverable error(s) in work request %s.", workRequestID)
			return *nextErr.Message
		}
	}
	klog.V(6).Infof("No non-recoverable errors for work request %s found.", workRequestID)
	return ""
}
