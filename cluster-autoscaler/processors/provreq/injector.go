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
	"time"

	apiv1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest"
	provreqconditions "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	provreqpods "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"k8s.io/utils/lru"
)

// ProvisioningRequestPodsInjector creates in-memory pods from ProvisioningRequest and inject them to unscheduled pods list.
type ProvisioningRequestPodsInjector struct {
	initialRetryTime                   time.Duration
	maxBackoffTime                     time.Duration
	backoffDuration                    *lru.Cache
	clock                              clock.PassiveClock
	client                             *provreqclient.ProvisioningRequestClient
	lastProvisioningRequestProcessTime time.Time
}

// IsAvailableForProvisioning checks if the provisioning request is the correct state for processing and provisioning has not been attempted recently.
func (p *ProvisioningRequestPodsInjector) IsAvailableForProvisioning(pr *provreqwrapper.ProvisioningRequest) bool {
	conditions := pr.Status.Conditions
	if apimeta.IsStatusConditionTrue(conditions, v1.Failed) || apimeta.IsStatusConditionTrue(conditions, v1.Provisioned) {
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

// MarkAsFailed marks the ProvisioningRequest as failed.
func (p *ProvisioningRequestPodsInjector) MarkAsFailed(pr *provreqwrapper.ProvisioningRequest, reason string, message string) {
	provreqconditions.AddOrUpdateCondition(pr, v1.Failed, metav1.ConditionTrue, reason, message, metav1.NewTime(p.clock.Now()))
	if _, err := p.client.UpdateProvisioningRequest(pr.ProvisioningRequest); err != nil {
		klog.Errorf("failed add Failed condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, err)
	}
	p.lastProvisioningRequestProcessTime = p.clock.Now()
}

// GetPodsFromNextRequest picks one ProvisioningRequest meeting the condition passed using isSupportedClass function, marks it as accepted and returns pods from it.
func (p *ProvisioningRequestPodsInjector) GetPodsFromNextRequest(
	isSupportedClass func(*provreqwrapper.ProvisioningRequest) bool,
) ([]*apiv1.Pod, error) {
	provReqs, err := p.client.ProvisioningRequests()
	if err != nil {
		return nil, err
	}
	for _, pr := range provReqs {
		if !isSupportedClass(pr) {
			continue
		}
		conditions := pr.Status.Conditions
		if apimeta.IsStatusConditionTrue(conditions, v1.Failed) || apimeta.IsStatusConditionTrue(conditions, v1.Provisioned) {
			p.backoffDuration.Remove(key(pr))
			continue
		}

		// Inject pods if ProvReq wasn't scaled up before or it has Provisioned == False condition more than defaultRetryTime
		if !p.IsAvailableForProvisioning(pr) {
			continue
		}

		podsFromProvReq, err := provreqpods.PodsForProvisioningRequest(pr)
		if err != nil {
			klog.Errorf("Failed to get pods for ProvisioningRequest %v", pr.Name)
			p.MarkAsFailed(pr, provreqconditions.FailedToCreatePodsReason, err.Error())
			continue
		}

		return podsFromProvReq, nil
	}
	return nil, nil
}

// Process pick one ProvisioningRequest, update Accepted condition and inject pods to unscheduled pods list.
func (p *ProvisioningRequestPodsInjector) Process(
	_ *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod,
) ([]*apiv1.Pod, error) {
	podsFromProvReq, err := p.GetPodsFromNextRequest(
		func(pr *provreqwrapper.ProvisioningRequest) bool {
			_, found := provisioningrequest.SupportedProvisioningClasses[pr.Spec.ProvisioningClassName]
			if !found {
				klog.Warningf("Provisioning Class %s is not supported for ProvReq %s/%s", pr.Spec.ProvisioningClassName, pr.Namespace, pr.Name)
			}
			return found
		})

	if err != nil {
		return unschedulablePods, err
	}

	return append(unschedulablePods, podsFromProvReq...), nil
}

// CleanUp cleans up the processor's internal structures.
func (p *ProvisioningRequestPodsInjector) CleanUp() {}

// NewProvisioningRequestPodsInjector creates a ProvisioningRequest filter processor.
func NewProvisioningRequestPodsInjector(kubeConfig *rest.Config, initialBackoffTime, maxBackoffTime time.Duration, maxCacheSize int) (*ProvisioningRequestPodsInjector, error) {
	client, err := provreqclient.NewProvisioningRequestClient(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &ProvisioningRequestPodsInjector{initialRetryTime: initialBackoffTime, maxBackoffTime: maxBackoffTime, backoffDuration: lru.New(maxCacheSize), client: client, clock: clock.RealClock{}, lastProvisioningRequestProcessTime: time.Now()}, nil
}

func key(pr *provreqwrapper.ProvisioningRequest) string {
	return string(pr.UID)
}

// LastProvisioningRequestProcessTime returns the time when the last provisioning request was processed.
func (p *ProvisioningRequestPodsInjector) LastProvisioningRequestProcessTime() time.Time {
	return p.lastProvisioningRequestProcessTime
}
