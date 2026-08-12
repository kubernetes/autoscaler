/*
Copyright 2024 The Kubernetes Authors.

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

package provreq

import (
	"context"
	"time"

	apiv1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"k8s.io/utils/lru"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/provisioningrequest"
	provreqconditions "sigs.k8s.io/cluster-autoscaler/pkg/provisioningrequest/conditions"
	provreqpods "sigs.k8s.io/cluster-autoscaler/pkg/provisioningrequest/pods"
	"sigs.k8s.io/cluster-autoscaler/pkg/provisioningrequest/provreqclient"
	"sigs.k8s.io/cluster-autoscaler/pkg/provisioningrequest/provreqwrapper"
)

// ProvisioningRequestPodsInjector creates in-memory pods from ProvisioningRequest and inject them to unscheduled pods list.
type ProvisioningRequestPodsInjector struct {
	initialRetryTime                   time.Duration
	maxBackoffTime                     time.Duration
	backoffDuration                    *lru.Cache
	clock                              clock.PassiveClock
	client                             *provreqclient.ProvisioningRequestClient
	lastProvisioningRequestProcessTime time.Time
	checkCapacityBatchProcessing       bool
	checkCapacityProcessorInstance     string
}

// IsAvailableForProvisioning checks if the provisioning request is the correct state for processing and provisioning has not been attempted recently.
func (p *ProvisioningRequestPodsInjector) IsAvailableForProvisioning(pr *provreqwrapper.ProvisioningRequest) bool {
	conditions := pr.Status.Conditions
	if apimeta.IsStatusConditionTrue(conditions, v1.Failed) || apimeta.IsStatusConditionTrue(conditions, v1.Provisioned) {
		p.backoffDuration.Remove(key(pr))
		return false
	}
	provisioned := apimeta.FindStatusCondition(conditions, v1.Provisioned)
	if provisioned != nil {
		if provisioned.Status != metav1.ConditionFalse {
			return false
		}
		val, found := p.backoffDuration.Get(key(pr))
		retryTime, ok := val.(time.Duration)
		if !found || !ok {
			retryTime = p.initialRetryTime
		}
		if provisioned.LastTransitionTime.Add(retryTime).Before(p.clock.Now()) {
			p.backoffDuration.Remove(key(pr))
			p.backoffDuration.Add(key(pr), min(2*retryTime, p.maxBackoffTime))
			return true
		}
		return false
	}
	return true
}

// MarkAsAccepted marks the ProvisioningRequest as accepted.
func (p *ProvisioningRequestPodsInjector) MarkAsAccepted(ctx context.Context, pr *provreqwrapper.ProvisioningRequest) error {
	provreqconditions.AddOrUpdateCondition(ctx, pr, v1.Accepted, metav1.ConditionTrue, provreqconditions.AcceptedReason, provreqconditions.AcceptedMsg, metav1.NewTime(p.clock.Now()))
	if _, err := p.client.UpdateProvisioningRequest(ctx, pr.ProvisioningRequest); err != nil {
		klog.Errorf("failed add Accepted condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, err)
		return err
	}
	p.UpdateLastProcessTime()
	return nil
}

// MarkAsFailed marks the ProvisioningRequest as failed.
func (p *ProvisioningRequestPodsInjector) MarkAsFailed(ctx context.Context, pr *provreqwrapper.ProvisioningRequest, reason string, message string) {
	provreqconditions.AddOrUpdateCondition(ctx, pr, v1.Failed, metav1.ConditionTrue, reason, message, metav1.NewTime(p.clock.Now()))
	if _, err := p.client.UpdateProvisioningRequest(ctx, pr.ProvisioningRequest); err != nil {
		klog.Errorf("failed add Failed condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, err)
	}
	p.UpdateLastProcessTime()
}

func (p *ProvisioningRequestPodsInjector) isSupportedClass(ctx context.Context, pr *provreqwrapper.ProvisioningRequest) bool {
	return provisioningrequest.SupportedProvisioningClass(ctx, pr.ProvisioningRequest, p.checkCapacityProcessorInstance)
}

func (p *ProvisioningRequestPodsInjector) isSupportedCheckCapacityClass(ctx context.Context, pr *provreqwrapper.ProvisioningRequest) bool {
	return provisioningrequest.SupportedCheckCapacityClass(ctx, pr.ProvisioningRequest, p.checkCapacityProcessorInstance)
}

func (p *ProvisioningRequestPodsInjector) shouldMarkAsAccepted(ctx context.Context, pr *provreqwrapper.ProvisioningRequest) bool {
	// Don't mark as accepted the check capacity ProvReq when batch processing is enabled.
	// It will be marked later, in parallel, during processing the requests.
	return !p.checkCapacityBatchProcessing || !p.isSupportedCheckCapacityClass(ctx, pr)
}

// GetPodsFromNextRequest picks one ProvisioningRequest meeting the condition passed using isSupportedClass function, marks it as accepted and returns pods from it.
func (p *ProvisioningRequestPodsInjector) GetPodsFromNextRequest(ctx context.Context) ([]*apiv1.Pod, error) {
	provReqs, err := p.client.ProvisioningRequests(ctx)
	if err != nil {
		return nil, err
	}
	for _, pr := range provReqs {
		if !p.isSupportedClass(ctx, pr) {
			continue
		}

		// Inject pods if ProvReq wasn't scaled up before or it has Provisioned == False condition more than defaultRetryTime
		if !p.IsAvailableForProvisioning(pr) {
			continue
		}

		podsFromProvReq, err := provreqpods.PodsForProvisioningRequest(pr)
		if err != nil {
			klog.Errorf("Failed to get pods for ProvisioningRequest %v", pr.Name)
			p.MarkAsFailed(ctx, pr, provreqconditions.FailedToCreatePodsReason, err.Error())
			continue
		}
		if p.shouldMarkAsAccepted(ctx, pr) {
			if err := p.MarkAsAccepted(ctx, pr); err != nil {
				continue
			}
			return podsFromProvReq, nil
		}
		p.UpdateLastProcessTime()
		return podsFromProvReq, nil
	}
	return nil, nil
}

// ProvisioningRequestWithPods contains a ProvisioningRequest Wrapper
// and its associated pods.
type ProvisioningRequestWithPods struct {
	PrWrapper *provreqwrapper.ProvisioningRequest
	Pods      []*apiv1.Pod
}

// GetCheckCapacityBatch returns up to the requested number of ProvisioningRequestWithPods.
// We do not mark the PRs as accepted here.
// If we fail to get the pods for a PR, we mark the PR as failed and issue an update.
func (p *ProvisioningRequestPodsInjector) GetCheckCapacityBatch(ctx context.Context, maxPrs int) ([]ProvisioningRequestWithPods, error) {
	provReqs, err := p.client.ProvisioningRequests(ctx)
	if err != nil {
		return nil, err
	}
	prsWithPods := make([]ProvisioningRequestWithPods, 0, min(maxPrs, len(provReqs)))
	for _, pr := range provReqs {
		if len(prsWithPods) >= maxPrs {
			break
		}
		if !p.isSupportedCheckCapacityClass(ctx, pr) {
			continue
		}
		if !p.IsAvailableForProvisioning(pr) {
			continue
		}

		pods, err := provreqpods.PodsForProvisioningRequest(pr)
		if err != nil {
			klog.Errorf("Failed to get pods for ProvisioningRequest %v", pr.Name)
			p.MarkAsFailed(ctx, pr, provreqconditions.FailedToCreatePodsReason, err.Error())
			continue
		}
		prsWithPods = append(prsWithPods, ProvisioningRequestWithPods{pr, pods})
	}
	return prsWithPods, nil
}

// Process pick one ProvisioningRequest, update Accepted condition and inject pods to unscheduled pods list.
func (p *ProvisioningRequestPodsInjector) Process(
	ctx context.Context,
	_ *ca_context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod,
) ([]*apiv1.Pod, error) {
	podsFromProvReq, err := p.GetPodsFromNextRequest(ctx)
	if err != nil {
		return unschedulablePods, err
	}

	return append(unschedulablePods, podsFromProvReq...), nil
}

// CleanUp cleans up the processor's internal structures.
func (p *ProvisioningRequestPodsInjector) CleanUp() {}

// NewProvisioningRequestPodsInjector creates a ProvisioningRequest filter processor.
func NewProvisioningRequestPodsInjector(client *provreqclient.ProvisioningRequestClient, initialBackoffTime, maxBackoffTime time.Duration, maxCacheSize int, checkCapacityBatchProcessing bool, checkCapacityProcessorInstance string) *ProvisioningRequestPodsInjector {
	return &ProvisioningRequestPodsInjector{
		initialRetryTime:                   initialBackoffTime,
		maxBackoffTime:                     maxBackoffTime,
		backoffDuration:                    lru.New(maxCacheSize),
		client:                             client,
		clock:                              clock.RealClock{},
		lastProvisioningRequestProcessTime: time.Now(),
		checkCapacityBatchProcessing:       checkCapacityBatchProcessing,
		checkCapacityProcessorInstance:     checkCapacityProcessorInstance,
	}
}

func key(pr *provreqwrapper.ProvisioningRequest) string {
	return string(pr.UID)
}

// LastProvisioningRequestProcessTime returns the time when the last provisioning request was processed.
func (p *ProvisioningRequestPodsInjector) LastProvisioningRequestProcessTime() time.Time {
	return p.lastProvisioningRequestProcessTime
}

// UpdateLastProcessTime updates the time we last processed a ProvisioningRequest
// to now. This time is used to skip waiting between loops if a request
// was processed in the last loop.
func (p *ProvisioningRequestPodsInjector) UpdateLastProcessTime() {
	p.lastProvisioningRequestProcessTime = p.clock.Now()
}
