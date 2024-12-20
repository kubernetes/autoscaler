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
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	provreq_pods "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	"k8s.io/klog/v2"
)

const (
	defaultReservationTime    = 10 * time.Minute
	defaultExpirationTime     = 7 * 24 * time.Hour // 7 days
	defaultTerminalProvReqTTL = 7 * 24 * time.Hour // 7 days
	// defaultMaxUpdated is a limit for ProvisioningRequest to update conditions in one ClusterAutoscaler loop.
	defaultMaxUpdated = 20
)

type injector interface {
	TrySchedulePods(clusterSnapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, isNodeAcceptable func(*framework.NodeInfo) bool, breakOnFailure bool) ([]scheduling.Status, int, error)
}

type provReqProcessor struct {
	now        func() time.Time
	maxUpdated int
	client     *provreqclient.ProvisioningRequestClient
	injector   injector
}

// NewProvReqProcessor return ProvisioningRequestProcessor.
func NewProvReqProcessor(client *provreqclient.ProvisioningRequestClient) *provReqProcessor {
	return &provReqProcessor{now: time.Now, maxUpdated: defaultMaxUpdated, client: client, injector: scheduling.NewHintingSimulator()}
}

// Refresh implements loop.Observer interface and will be run at the start
// of every iteration of the main loop. It tries to fetch current
// ProvisioningRequests and processes up to p.maxUpdated of them.
func (p *provReqProcessor) Refresh() {
	provReqs, err := p.client.ProvisioningRequests()
	if err != nil {
		klog.Errorf("Failed to get ProvisioningRequests list, err: %v", err)
		return
	}
	p.refresh(provReqs)
}

// refresh iterates over ProvisioningRequests and apply:
// -BookingExpired condition for Provisioned ProvisioningRequest if capacity reservation time is expired.
// -Failed condition for ProvisioningRequest that were not provisioned during defaultExpirationTime.
// TODO(yaroslava): fetch reservation and expiration time from ProvisioningRequest
func (p *provReqProcessor) refresh(provReqs []*provreqwrapper.ProvisioningRequest) {
	expiredProvReq := []*provreqwrapper.ProvisioningRequest{}
	failedProvReq := []*provreqwrapper.ProvisioningRequest{}
	for _, provReq := range provReqs {
		if len(expiredProvReq) >= p.maxUpdated {
			break
		}
		if ok, found := provisioningrequest.SupportedProvisioningClasses[provReq.Spec.ProvisioningClassName]; !ok || !found {
			continue
		}
		conditions := provReq.Status.Conditions
		if apimeta.IsStatusConditionTrue(conditions, v1.BookingExpired) || apimeta.IsStatusConditionTrue(conditions, v1.Failed) {
			continue
		}
		provisioned := apimeta.FindStatusCondition(conditions, v1.Provisioned)
		if provisioned != nil && provisioned.Status == metav1.ConditionTrue {
			if provisioned.LastTransitionTime.Add(defaultReservationTime).Before(p.now()) {
				expiredProvReq = append(expiredProvReq, provReq)
			}
		} else if len(failedProvReq) < p.maxUpdated-len(expiredProvReq) {
			created := provReq.CreationTimestamp
			if created.Add(defaultExpirationTime).Before(p.now()) {
				failedProvReq = append(failedProvReq, provReq)
			}
		}
	}
	for _, provReq := range expiredProvReq {
		conditions.AddOrUpdateCondition(provReq, v1.BookingExpired, metav1.ConditionTrue, conditions.CapacityReservationTimeExpiredReason, conditions.CapacityReservationTimeExpiredMsg, metav1.NewTime(p.now()))
		_, updErr := p.client.UpdateProvisioningRequest(provReq.ProvisioningRequest)
		if updErr != nil {
			klog.Errorf("failed to add BookingExpired condition to ProvReq %s/%s, err: %v", provReq.Namespace, provReq.Name, updErr)
			continue
		}
	}
	for _, provReq := range failedProvReq {
		conditions.AddOrUpdateCondition(provReq, v1.Failed, metav1.ConditionTrue, conditions.ExpiredReason, conditions.ExpiredMsg, metav1.NewTime(p.now()))
		_, updErr := p.client.UpdateProvisioningRequest(provReq.ProvisioningRequest)
		if updErr != nil {
			klog.Errorf("failed to add Failed condition to ProvReq %s/%s, err: %v", provReq.Namespace, provReq.Name, updErr)
			continue
		}
	}
	p.DeleteOldProvReqs(provReqs)
}

// CleanUp cleans up internal state
func (p *provReqProcessor) CleanUp() {}

// Process implements PodListProcessor.Process() and inject fake pods to the cluster snapshoot for Provisioned ProvReqs in order to
// reserve capacity from ScaleDown.
func (p *provReqProcessor) Process(context *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	err := p.bookCapacity(context)
	if err != nil {
		klog.Warningf("Failed to book capacity for ProvisioningRequests: %s", err)
	}
	return unschedulablePods, nil
}

// bookCapacity schedule fake pods for ProvisioningRequest that should have reserved capacity
// in the cluster.
func (p *provReqProcessor) bookCapacity(ctx *context.AutoscalingContext) error {
	provReqs, err := p.client.ProvisioningRequests()
	if err != nil {
		return fmt.Errorf("couldn't fetch ProvisioningRequests in the cluster: %v", err)
	}
	podsToCreate := []*apiv1.Pod{}
	for _, provReq := range provReqs {
		if !conditions.ShouldCapacityBeBooked(provReq) {
			continue
		}
		pods, err := provreq_pods.PodsForProvisioningRequest(provReq)
		if err != nil {
			// ClusterAutoscaler was able to create pods before, so we shouldn't have error here.
			// If there is an error, mark PR as invalid, because we won't be able to book capacity
			// for it anyway.
			conditions.AddOrUpdateCondition(provReq, v1.Failed, metav1.ConditionTrue, conditions.FailedToBookCapacityReason, fmt.Sprintf("Couldn't create pods, err: %v", err), metav1.Now())
			if _, err := p.client.UpdateProvisioningRequest(provReq.ProvisioningRequest); err != nil {
				klog.Errorf("failed to add Accepted condition to ProvReq %s/%s, err: %v", provReq.Namespace, provReq.Name, err)
			}
			continue
		}
		podsToCreate = append(podsToCreate, pods...)
	}
	if len(podsToCreate) == 0 {
		return nil
	}
	// Scheduling the pods to reserve capacity for provisioning request.
	if _, _, err = p.injector.TrySchedulePods(ctx.ClusterSnapshot, podsToCreate, scheduling.ScheduleAnywhere, false); err != nil {
		return err
	}
	return nil
}

// DeleteOldProvReqs delete ProvReq that have terminal state (Provisioned/Failed == True) more than a week.
func (p *provReqProcessor) DeleteOldProvReqs(provReqs []*provreqwrapper.ProvisioningRequest) {
	provReqQuota := klogx.NewLoggingQuota(30)
	for _, provReq := range provReqs {
		conditions := provReq.Status.Conditions
		provisioned := apimeta.FindStatusCondition(conditions, v1.Provisioned)
		failed := apimeta.FindStatusCondition(conditions, v1.Failed)
		if provisioned != nil && provisioned.LastTransitionTime.Add(defaultTerminalProvReqTTL).Before(p.now()) ||
			failed != nil && failed.LastTransitionTime.Add(defaultTerminalProvReqTTL).Before(p.now()) {
			klogx.V(4).UpTo(provReqQuota).Infof("Delete old ProvisioningRequest %s/%s", provReq.Namespace, provReq.Name)
			err := p.client.DeleteProvisioningRequest(provReq.ProvisioningRequest)
			if err != nil {
				klog.Warningf("Couldn't delete old %s/%s Provisioning Request, err: %v", provReq.Namespace, provReq.Name, err)
			}
		}
	}
}
