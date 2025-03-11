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
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest"
	provreqconditions "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	provreqpods "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	defaultRetryTime = 10 * time.Minute
)

// ProvisioningRequestPodsInjector creates in-memory pods from ProvisioningRequest and inject them to unscheduled pods list.
type ProvisioningRequestPodsInjector struct {
	client *provreqclient.ProvisioningRequestClient
	clock  clock.PassiveClock
}

// Process pick one ProvisioningRequest, update Accepted condition and inject pods to unscheduled pods list.
func (p *ProvisioningRequestPodsInjector) Process(
	_ *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod,
) ([]*apiv1.Pod, error) {
	provReqs, err := p.client.ProvisioningRequests()
	if err != nil {
		return nil, err
	}
	for _, pr := range provReqs {
		if ok, found := provisioningrequest.SupportedProvisioningClasses[pr.Spec.ProvisioningClassName]; !ok || !found {
			klog.Warningf("Provisioning Class %s is not supported", pr.Spec.ProvisioningClassName)
			continue
		}
		conditions := pr.Status.Conditions
		if apimeta.IsStatusConditionTrue(conditions, v1.Failed) || apimeta.IsStatusConditionTrue(conditions, v1.Provisioned) {
			continue
		}

		provisioned := apimeta.FindStatusCondition(conditions, v1.Provisioned)

		//TODO(yaroslava): support exponential backoff
		// Inject pods if ProvReq wasn't scaled up before or it has Provisioned == False condition more than defaultRetryTime
		inject := true
		if provisioned != nil {
			if provisioned.Status == metav1.ConditionFalse && provisioned.LastTransitionTime.Add(defaultRetryTime).Before(p.clock.Now()) {
				inject = true
			} else {
				inject = false
			}
		}
		if inject {
			provreqpods, err := provreqpods.PodsForProvisioningRequest(pr)
			if err != nil {
				klog.Errorf("Failed to get pods for ProvisioningRequest %v", pr.Name)
				provreqconditions.AddOrUpdateCondition(pr, v1.Failed, metav1.ConditionTrue, provreqconditions.FailedToCreatePodsReason, err.Error(), metav1.NewTime(p.clock.Now()))
				if _, err := p.client.UpdateProvisioningRequest(pr.ProvisioningRequest); err != nil {
					klog.Errorf("failed add Failed condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, err)
				}
				continue
			}
			provreqconditions.AddOrUpdateCondition(pr, v1.Accepted, metav1.ConditionTrue, provreqconditions.AcceptedReason, provreqconditions.AcceptedMsg, metav1.NewTime(p.clock.Now()))
			if _, err := p.client.UpdateProvisioningRequest(pr.ProvisioningRequest); err != nil {
				klog.Errorf("failed add Accepted condition to ProvReq %s/%s, err: %v", pr.Namespace, pr.Name, err)
				continue
			}
			unschedulablePods := append(unschedulablePods, provreqpods...)
			return unschedulablePods, nil
		}
	}
	return unschedulablePods, nil
}

// CleanUp cleans up the processor's internal structures.
func (p *ProvisioningRequestPodsInjector) CleanUp() {}

// NewProvisioningRequestPodsInjector creates a ProvisioningRequest filter processor.
func NewProvisioningRequestPodsInjector(kubeConfig *rest.Config) (pods.PodListProcessor, error) {
	client, err := provreqclient.NewProvisioningRequestClient(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &ProvisioningRequestPodsInjector{client: client, clock: clock.RealClock{}}, nil
}
