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

package checkcapacity

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/provreq"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"

	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
)

const (
	// NoRetryParameterKey is a a key for ProvReq's Parameters that describes
	// if ProvisioningRequest should be retried in case CA cannot provision it.
	// Supported values are "true" and "false" - by default ProvisioningRequests are always retried.
	// Currently supported only for checkcapacity class.
	NoRetryParameterKey = "noRetry"
)

type checkCapacityProvClass struct {
	context                                      *context.AutoscalingContext
	client                                       *provreqclient.ProvisioningRequestClient
	schedulingSimulator                          *scheduling.HintingSimulator
	checkCapacityProvisioningRequestMaxBatchSize int
	checkCapacityProvisioningRequestBatchTimebox time.Duration
	provreqInjector                              *provreq.ProvisioningRequestPodsInjector
}

// New create check-capacity scale-up mode.
func New(
	client *provreqclient.ProvisioningRequestClient,
	provreqInjector *provreq.ProvisioningRequestPodsInjector,
) *checkCapacityProvClass {
	return &checkCapacityProvClass{client: client, provreqInjector: provreqInjector}
}

func (o *checkCapacityProvClass) Initialize(
	autoscalingContext *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	estimatorBuilder estimator.EstimatorBuilder,
	taintConfig taints.TaintConfig,
	schedulingSimulator *scheduling.HintingSimulator,
) {
	o.context = autoscalingContext
	o.schedulingSimulator = schedulingSimulator
	if autoscalingContext.CheckCapacityBatchProcessing {
		o.checkCapacityProvisioningRequestBatchTimebox = autoscalingContext.CheckCapacityProvisioningRequestBatchTimebox
		o.checkCapacityProvisioningRequestMaxBatchSize = autoscalingContext.CheckCapacityProvisioningRequestMaxBatchSize
	} else {
		o.checkCapacityProvisioningRequestMaxBatchSize = 1
	}
}

// Provision return if there is capacity in the cluster for pods from ProvisioningRequest.
func (o *checkCapacityProvClass) Provision(
	unschedulablePods []*apiv1.Pod,
	nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet,
	nodeInfos map[string]*framework.NodeInfo,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	combinedStatus := NewCombinedStatusSet()
	startTime := time.Now()

	o.context.ClusterSnapshot.Fork()
	defer o.context.ClusterSnapshot.Revert()

	// Gather ProvisioningRequests.
	prs, err := o.getProvisioningRequestsAndPods(unschedulablePods)
	if err != nil {
		return status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerErrorf(errors.InternalError, "Error fetching provisioning requests and associated pods: %s", err.Error()))
	} else if len(prs) == 0 {
		return &status.ScaleUpStatus{Result: status.ScaleUpNotTried}, nil
	}

	if o.provreqInjector != nil {
		// for more frequent iterations.
		// See https://github.com/kubernetes/autoscaler/pull/7271
		o.provreqInjector.UpdateLastProcessTime()
	}

	// Add accepted condition to ProvisioningRequests.
	for _, pr := range prs {
		conditions.AddOrUpdateCondition(pr.PrWrapper, v1.Accepted, metav1.ConditionTrue, conditions.AcceptedReason, conditions.AcceptedMsg, metav1.Now())
	}

	// Check Capacity. Add Provisioned or Failed conditions.
	processedPrs := o.checkCapacityBatch(prs, &combinedStatus, startTime)

	// Use client to update ProvisioningRequests conditions.
	updateRequests(o.client, processedPrs, &combinedStatus)

	return combinedStatus.Export()
}

func (o *checkCapacityProvClass) getProvisioningRequestsAndPods(unschedulablePods []*apiv1.Pod) ([]provreq.ProvisioningRequestWithPods, error) {
	if !o.isBatchEnabled() {
		klog.Info("Processing single provisioning request (non-batch)")
		prs := provreqclient.ProvisioningRequestsForPods(o.client, unschedulablePods)
		prs = provreqclient.FilterOutProvisioningClass(prs, v1.ProvisioningClassCheckCapacity)
		if len(prs) == 0 {
			return nil, nil
		}
		return []provreq.ProvisioningRequestWithPods{{PrWrapper: prs[0], Pods: unschedulablePods}}, nil
	}

	batch, err := o.provreqInjector.GetCheckCapacityBatch(o.checkCapacityProvisioningRequestMaxBatchSize)
	if err != nil {
		return nil, err
	}
	klog.Infof("Processing provisioning requests as batch of size %d", len(batch))
	return batch, nil
}

func (o *checkCapacityProvClass) isBatchEnabled() bool {
	return o.provreqInjector != nil && o.checkCapacityProvisioningRequestMaxBatchSize > 1
}

func (o *checkCapacityProvClass) checkCapacityBatch(reqs []provreq.ProvisioningRequestWithPods, combinedStatus *combinedStatusSet, startTime time.Time) []*provreqwrapper.ProvisioningRequest {
	updates := make([]*provreqwrapper.ProvisioningRequest, 0, len(reqs))
	for _, req := range reqs {
		if err := o.checkCapacity(req.Pods, req.PrWrapper, combinedStatus); err != nil {
			klog.Errorf("error checking capacity %v", err)
			continue
		}

		updates = append(updates, req.PrWrapper)

		// timebox checkCapacity when batch processing.
		if o.isBatchEnabled() && time.Since(startTime) > o.checkCapacityProvisioningRequestBatchTimebox {
			klog.Infof("Batch timebox exceeded, processed %d check capacity provisioning requests this iteration", len(updates))
			break
		}
	}
	return updates
}

// checkCapacity checks if there is capacity, updates combinedStatus and Conditions. If capacity is found, it commits to the clusterSnapshot.
func (o *checkCapacityProvClass) checkCapacity(unschedulablePods []*apiv1.Pod, provReq *provreqwrapper.ProvisioningRequest, combinedStatus *combinedStatusSet) error {
	o.context.ClusterSnapshot.Fork()

	// Case 1: Capacity fits.
	scheduled, _, err := o.schedulingSimulator.TrySchedulePods(o.context.ClusterSnapshot, unschedulablePods, scheduling.ScheduleAnywhere, true)
	if err == nil && len(scheduled) == len(unschedulablePods) {
		commitError := o.context.ClusterSnapshot.Commit()
		if commitError != nil {
			o.context.ClusterSnapshot.Revert()
			return commitError
		}
		combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpSuccessful})
		conditions.AddOrUpdateCondition(provReq, v1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsFoundReason, conditions.CapacityIsFoundMsg, metav1.Now())
		return nil
	}
	// Case 2: Capacity doesn't fit.
	o.context.ClusterSnapshot.Revert()
	combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpNoOptionsAvailable})
	if noRetry, ok := provReq.Spec.Parameters[NoRetryParameterKey]; ok && noRetry == "true" {
		// Failed=true condition triggers retry in Kueue. Otherwise ProvisioningRequest with Provisioned=Failed
		// condition block capacity in Kueue even if it's in the middle of backoff waiting time.
		conditions.AddOrUpdateCondition(provReq, v1.Failed, metav1.ConditionTrue, conditions.CapacityIsNotFoundReason, "CA could not find requested capacity", metav1.Now())
	} else {
		if noRetry, ok := provReq.Spec.Parameters[NoRetryParameterKey]; ok && noRetry != "false" {
			klog.Errorf("Ignoring Parameter %v with invalid value: %v in ProvisioningRequest: %v. Supported values are: \"true\", \"false\"", NoRetryParameterKey, noRetry, provReq.Name)
		}
		conditions.AddOrUpdateCondition(provReq, v1.Provisioned, metav1.ConditionFalse, conditions.CapacityIsNotFoundReason, "Capacity is not found, CA will try to find it later.", metav1.Now())
	}
	return err
}

// updateRequests calls the client to update ProvisioningRequests, in parallel.
func updateRequests(client *provreqclient.ProvisioningRequestClient, prWrappers []*provreqwrapper.ProvisioningRequest, combinedStatus *combinedStatusSet) {
	wg := sync.WaitGroup{}
	wg.Add(len(prWrappers))
	lock := sync.Mutex{}
	for _, wrapper := range prWrappers {
		go func() {
			provReq := wrapper.ProvisioningRequest
			_, updErr := client.UpdateProvisioningRequest(provReq)
			if updErr != nil {
				err := fmt.Errorf("failed to update ProvReq %s/%s, err: %v", provReq.Namespace, provReq.Name, updErr)
				scaleUpStatus, _ := status.UpdateScaleUpError(&status.ScaleUpStatus{}, errors.NewAutoscalerErrorf(errors.InternalError, "error during ScaleUp: %s", err.Error()))
				lock.Lock()
				combinedStatus.Add(scaleUpStatus)
				lock.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// combinedStatusSet is a helper struct to combine multiple ScaleUpStatuses into one. It keeps track of the best result and all errors that occurred during the ScaleUp process.
type combinedStatusSet struct {
	Result        status.ScaleUpResult
	ScaleupErrors map[*errors.AutoscalerError]bool
}

// Add adds a ScaleUpStatus to the combinedStatusSet.
func (c *combinedStatusSet) Add(newStatus *status.ScaleUpStatus) {
	// This represents the priority of the ScaleUpResult. The final result is the one with the highest priority in the set.
	resultPriority := map[status.ScaleUpResult]int{
		status.ScaleUpNotTried:           0,
		status.ScaleUpNoOptionsAvailable: 1,
		status.ScaleUpError:              2,
		status.ScaleUpSuccessful:         3,
	}

	// If even one scaleUpSuccessful is present, the final result is ScaleUpSuccessful.
	// If no ScaleUpSuccessful is present, and even one ScaleUpError is present, the final result is ScaleUpError.
	// If no ScaleUpSuccessful or ScaleUpError is present, and even one ScaleUpNoOptionsAvailable is present, the final result is ScaleUpNoOptionsAvailable.
	// If no ScaleUpSuccessful, ScaleUpError or ScaleUpNoOptionsAvailable is present, the final result is ScaleUpNotTried.
	if resultPriority[c.Result] < resultPriority[newStatus.Result] {
		c.Result = newStatus.Result
	}
	if newStatus.ScaleUpError != nil {
		if _, found := c.ScaleupErrors[newStatus.ScaleUpError]; !found {
			c.ScaleupErrors[newStatus.ScaleUpError] = true
		}
	}
}

// formatMessageFromBatchErrors formats a message from a list of errors.
func (c *combinedStatusSet) formatMessageFromBatchErrors(errs []errors.AutoscalerError) string {
	firstErr := errs[0]
	var builder strings.Builder
	builder.WriteString(firstErr.Error())
	builder.WriteString(" ...and other concurrent errors: [")
	formattedErrs := map[errors.AutoscalerError]bool{
		firstErr: true,
	}
	for _, err := range errs {
		if _, has := formattedErrs[err]; has {
			continue
		}
		formattedErrs[err] = true
		message := err.Error()
		if len(formattedErrs) > 2 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%q", message))
	}
	builder.WriteString("]")
	return builder.String()
}

// combineBatchScaleUpErrors combines multiple errors into one. If there is only one error, it returns that error. If there are multiple errors, it combines them into one error with a message that contains all the errors.
func (c *combinedStatusSet) combineBatchScaleUpErrors() *errors.AutoscalerError {
	if len(c.ScaleupErrors) == 0 {
		return nil
	}
	if len(c.ScaleupErrors) == 1 {
		for err := range c.ScaleupErrors {
			return err
		}
	}
	uniqueMessages := make(map[string]bool)
	for err := range c.ScaleupErrors {
		uniqueMessages[(*err).Error()] = true
	}
	if len(uniqueMessages) == 1 {
		for err := range c.ScaleupErrors {
			return err
		}
	}
	// sort to stabilize the results and easier log aggregation
	errs := make([]errors.AutoscalerError, 0, len(c.ScaleupErrors))
	for err := range c.ScaleupErrors {
		errs = append(errs, *err)
	}
	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Error() < errs[j].Error()
	})
	message := c.formatMessageFromBatchErrors(errs)
	combinedErr := errors.NewAutoscalerError(errors.InternalError, message)
	return &combinedErr
}

// Export converts the combinedStatusSet into a ScaleUpStatus.
func (c *combinedStatusSet) Export() (*status.ScaleUpStatus, errors.AutoscalerError) {
	result := &status.ScaleUpStatus{Result: c.Result}
	if len(c.ScaleupErrors) > 0 {
		result.ScaleUpError = c.combineBatchScaleUpErrors()
	}

	var resErr errors.AutoscalerError

	if result.Result == status.ScaleUpError {
		resErr = *result.ScaleUpError
	}

	return result, resErr
}

// NewCombinedStatusSet creates a new combinedStatusSet.
func NewCombinedStatusSet() combinedStatusSet {
	return combinedStatusSet{
		Result:        status.ScaleUpNotTried,
		ScaleupErrors: make(map[*errors.AutoscalerError]bool),
	}
}
