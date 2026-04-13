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
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/processors/provreq"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
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

	// PartialCapacityCheckKey is a key for ProvReq's Parameters that enables
	// per-PodSet capacity evaluation. Supported values are "bookPartial" and "checkOnly".
	// By default this is not set, and checkCapacity evaluates all pods atomically.
	PartialCapacityCheckKey = "partialCapacityCheck"
	// When partialCapacityCheck Parameter is set to PartialCapacityCheckBookPartial,
	// the ProvisioningRequest condition is set to Provisioned=true if capacity is found for
	// some of the ProvReq PodSets.
	PartialCapacityCheckBookPartial = "bookPartial"
	// When partialCapacityCheck Parameter is set to PartialCapacityCheckCheckOnly,
	// the ProvisioningRequest condition is set to Provisioned=false even if capacity is found for
	// some of the ProvReq PodSets. If partial capacity is found, the condition message and ProvReq Details
	// will reflect the capacity state.
	PartialCapacityCheckCheckOnly = "checkOnly"
)

// Regex to match pod names created by PodsForProvisioningRequest.
var podSetIndexPattern = regexp.MustCompile(`-(\d+)-(\d+)$`)

type checkCapacityProvClass struct {
	autoscalingCtx                               *ca_context.AutoscalingContext
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
	autoscalingCtx *ca_context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	estimatorBuilder estimator.EstimatorBuilder,
	taintConfig taints.TaintConfig,
	schedulingSimulator *scheduling.HintingSimulator,
	quotasTrackerFactory *resourcequotas.TrackerFactory,
) {
	o.autoscalingCtx = autoscalingCtx
	o.schedulingSimulator = schedulingSimulator
	if autoscalingCtx.CheckCapacityBatchProcessing {
		o.checkCapacityProvisioningRequestBatchTimebox = autoscalingCtx.CheckCapacityProvisioningRequestBatchTimebox
		o.checkCapacityProvisioningRequestMaxBatchSize = autoscalingCtx.CheckCapacityProvisioningRequestMaxBatchSize
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

	o.autoscalingCtx.ClusterSnapshot.Fork()
	defer o.autoscalingCtx.ClusterSnapshot.Revert()

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
		prs = provreqclient.FilterOutProvisioningClass(prs, v1.ProvisioningClassCheckCapacity, o.autoscalingCtx.CheckCapacityProcessorInstance)
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
	o.autoscalingCtx.ClusterSnapshot.Fork()

	partialCapacityMode, partialCapacityCheckEnabled := getPartialCapacityCheckMode(provReq)

	if partialCapacityCheckEnabled {
		// Schedule per podset using nested forks. For each podset, if all its pods fit, commit the
		// inner fork into the outer fork so subsequent podsets see the consumed capacity. At the end,
		// commit or revert the outer fork depending on the result and mode.
		podsByPodSet := groupPodsByPodSet(unschedulablePods, provReq)
		schedulablePodSets := make([]string, 0)

		// PodSets are evaluated in spec order. Earlier PodSets that fit consume
		// capacity within the simulation, which may prevent later PodSets from
		// scheduling. In checkOnly mode the capacity is reverted after evaluation.
		for i, podSetSpec := range provReq.Spec.PodSets {
			pods := podsByPodSet[i]
			if len(pods) == 0 {
				klog.Warningf("No pods matched podset %d (%s) for ProvReq %s, treating as not schedulable", i, podSetSpec.PodTemplateRef.Name, provReq.Name)
				continue
			}
			o.autoscalingCtx.ClusterSnapshot.Fork()
			scheduled, _, err := o.schedulingSimulator.TrySchedulePods(o.autoscalingCtx.ClusterSnapshot, pods, true, clustersnapshot.SchedulingOptions{})
			if err != nil {
				o.autoscalingCtx.ClusterSnapshot.Revert()
				o.autoscalingCtx.ClusterSnapshot.Revert()
				return err
			}
			if len(scheduled) == len(pods) {
				schedulablePodSets = append(schedulablePodSets, podSetSpec.PodTemplateRef.Name)
				if commitErr := o.autoscalingCtx.ClusterSnapshot.Commit(); commitErr != nil {
					o.autoscalingCtx.ClusterSnapshot.Revert()
					return commitErr
				}
			} else {
				o.autoscalingCtx.ClusterSnapshot.Revert()
			}
		}

		schedulablePodSetsJSON, jsonErr := json.Marshal(schedulablePodSets)
		if jsonErr != nil {
			klog.Errorf("failed to marshal schedulablePodSets for ProvReq %s: %v", provReq.Name, jsonErr)
		} else {
			provReq.SetProvisioningClassDetail(conditions.SchedulablePodSetsDetailKey, v1.Detail(schedulablePodSetsJSON))
		}

		// Case 1: All podsets fit.
		if len(schedulablePodSets) == len(provReq.Spec.PodSets) {
			combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpSuccessful})
			conditions.AddOrUpdateCondition(provReq, v1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsFoundReason, conditions.CapacityIsFoundMsg, metav1.Now())
			if commitErr := o.autoscalingCtx.ClusterSnapshot.Commit(); commitErr != nil {
				o.autoscalingCtx.ClusterSnapshot.Revert()
				return commitErr
			}
			return nil
		}

		// Case 2: Some podsets fit.
		if len(schedulablePodSets) > 0 {
			msg := fmt.Sprintf("%s. Schedulable podsets: %s", conditions.PartialCapacityIsFoundMsg, strings.Join(schedulablePodSets, ","))
			handlePartialCapacityStatusUpdate(provReq, combinedStatus, partialCapacityMode, msg)
			if partialCapacityMode == PartialCapacityCheckBookPartial {
				if commitErr := o.autoscalingCtx.ClusterSnapshot.Commit(); commitErr != nil {
					o.autoscalingCtx.ClusterSnapshot.Revert()
					return commitErr
				}
			} else {
				o.autoscalingCtx.ClusterSnapshot.Revert()
			}
			return nil
		}

		// Case 3: No podsets fit.
		o.autoscalingCtx.ClusterSnapshot.Revert()
		provReq.DeleteProvisioningClassDetail(conditions.SchedulablePodSetsDetailKey)
		combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpNoOptionsAvailable})
		setCapacityNotFoundCondition(provReq)
		return nil
	}

	// Non-partial path: schedule all pods at once and break on first failure.
	scheduled, _, err := o.schedulingSimulator.TrySchedulePods(o.autoscalingCtx.ClusterSnapshot, unschedulablePods, true, clustersnapshot.SchedulingOptions{})
	if err == nil && len(scheduled) == len(unschedulablePods) {
		// Case 1: All capacity fits.
		allPodSetNames := make([]string, 0, len(provReq.Spec.PodSets))
		for _, ps := range provReq.Spec.PodSets {
			allPodSetNames = append(allPodSetNames, ps.PodTemplateRef.Name)
		}
		schedulablePodSetsJSON, jsonErr := json.Marshal(allPodSetNames)
		if jsonErr != nil {
			klog.Errorf("failed to marshal schedulablePodSets for ProvReq %s: %v", provReq.Name, jsonErr)
		} else {
			provReq.SetProvisioningClassDetail(conditions.SchedulablePodSetsDetailKey, v1.Detail(schedulablePodSetsJSON))
		}
		combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpSuccessful})
		conditions.AddOrUpdateCondition(provReq, v1.Provisioned, metav1.ConditionTrue, conditions.CapacityIsFoundReason, conditions.CapacityIsFoundMsg, metav1.Now())
		if commitError := o.autoscalingCtx.ClusterSnapshot.Commit(); commitError != nil {
			o.autoscalingCtx.ClusterSnapshot.Revert()
			return commitError
		}
		return nil
	}

	// Case 3: Capacity doesn't fit.
	o.autoscalingCtx.ClusterSnapshot.Revert()
	// Clear the detail from any previous iteration.
	provReq.DeleteProvisioningClassDetail(conditions.SchedulablePodSetsDetailKey)
	combinedStatus.Add(&status.ScaleUpStatus{Result: status.ScaleUpNoOptionsAvailable})
	setCapacityNotFoundCondition(provReq)
	return err
}

const podNameFormatLen = 3

// groupPodsByPodSet buckets pods by their podset index, extracted from the pod name.
// Pod names are created in the format {GenerateName}{i}-{j}, where i is the podset index
// and j is the pod index within the podset. Pods not belonging to provReq are filtered out.
func groupPodsByPodSet(pods []*apiv1.Pod, provReq *provreqwrapper.ProvisioningRequest) map[int][]*apiv1.Pod {
	groups := make(map[int][]*apiv1.Pod)
	for _, pod := range pods {
		if pod.Annotations[v1.ProvisioningRequestPodAnnotationKey] != provReq.Name {
			continue
		}
		matches := podSetIndexPattern.FindStringSubmatch(pod.Name)
		if len(matches) == podNameFormatLen {
			idx, _ := strconv.Atoi(matches[1]) // safe: regex guarantees digits
			groups[idx] = append(groups[idx], pod)
		}
	}
	return groups
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

// setCapacityNotFoundCondition sets the appropriate condition on the ProvReq when no capacity is found,
// respecting the noRetry parameter.
func setCapacityNotFoundCondition(provReq *provreqwrapper.ProvisioningRequest) {
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
}

// handlePartialCapacityStatusUpdate handles the combineStatusSet update and ProvReq Provisioned status update for the partial capacity case
func handlePartialCapacityStatusUpdate(provReq *provreqwrapper.ProvisioningRequest, combinedStatus *combinedStatusSet, provReqMode, conditionMsg string) {
	provisionedStatus := metav1.ConditionFalse
	scaleupStatus := status.ScaleUpSuccessful
	if provReqMode == PartialCapacityCheckBookPartial {
		provisionedStatus = metav1.ConditionTrue
	}
	combinedStatus.Add(&status.ScaleUpStatus{Result: scaleupStatus})
	conditions.AddOrUpdateCondition(provReq, v1.Provisioned, provisionedStatus, conditions.PartialCapacityIsFoundReason, conditionMsg, metav1.Now())
}

func getPartialCapacityCheckMode(provReq *provreqwrapper.ProvisioningRequest) (string, bool) {
	val, ok := provReq.Spec.Parameters[PartialCapacityCheckKey]
	if !ok {
		return "", false
	}

	switch val {
	case PartialCapacityCheckBookPartial:
		return PartialCapacityCheckBookPartial, true
	case PartialCapacityCheckCheckOnly:
		return PartialCapacityCheckCheckOnly, true
	default:
		klog.Errorf("Ignoring Parameter %v with invalid value: %v in ProvisioningRequest: %v. Supported values are: %v, %v",
			PartialCapacityCheckKey, val, provReq.Name,
			PartialCapacityCheckBookPartial, PartialCapacityCheckCheckOnly,
		)
		return "", false
	}
}
