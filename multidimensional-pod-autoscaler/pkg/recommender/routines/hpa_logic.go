package routines

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	mpa_api "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/model"
	klog "k8s.io/klog/v2"
)

// NormalizationArg is used to pass all needed information between functions as one structure
type NormalizationArg struct {
	Key               model.MpaID
	ScaleUpBehavior   *autoscalingv2.HPAScalingRules
	ScaleDownBehavior *autoscalingv2.HPAScalingRules
	MinReplicas       int32
	MaxReplicas       int32
	CurrentReplicas   int32
	DesiredReplicas   int32
}

func (r *recommender) ReconcileHorizontalAutoscaling(ctx context.Context, mpaShared *mpa_types.MultidimPodAutoscaler, key model.MpaID) error {
	// make a copy so that we never mutate the shared informer cache (conversion can mutate the object)
	mpa := mpaShared.DeepCopy()
	mpaStatusOriginal := mpa.Status.DeepCopy()

	reference := fmt.Sprintf("%s/%s/%s", mpa.Spec.ScaleTargetRef.Kind, mpa.Namespace, mpa.Spec.ScaleTargetRef.Name)

	targetGV, err := schema.ParseGroupVersion(mpa.Spec.ScaleTargetRef.APIVersion)
	if err != nil {
		klog.Errorf("%s: FailedGetScale - error: %v", v1.EventTypeWarning, err.Error())
		r.eventRecorder.Event(mpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
		setCondition(mpa, mpa_types.AbleToScale, v1.ConditionFalse, "FailedGetScale", "the MPA recommender was unable to get the target's current scale: %v", err)
		if err := r.updateStatusIfNeeded(ctx, mpaStatusOriginal, mpa); err != nil {
			klog.Errorf("Error updating MPA status: %v", err.Error())
			utilruntime.HandleError(err)
		}
		return fmt.Errorf("invalid API version in scale target reference: %v", err)
	}

	targetGK := schema.GroupKind{
		Group: targetGV.Group,
		Kind:  mpa.Spec.ScaleTargetRef.Kind,
	}

	mappings, err := r.controllerFetcher.GetRESTMappings(targetGK)
	if err != nil {
		klog.Errorf("%s: FailedGetScale - error: %v", v1.EventTypeWarning, err.Error())
		r.eventRecorder.Event(mpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
		setCondition(mpa, mpa_types.AbleToScale, v1.ConditionFalse, "FailedGetScale", "the MPA recommender was unable to get the target's current scale: %v", err)
		if err := r.updateStatusIfNeeded(ctx, mpaStatusOriginal, mpa); err != nil {
			klog.Errorf("Error updating MPA status: %v", err.Error())
			utilruntime.HandleError(err)
		}
		return fmt.Errorf("unable to determine resource for scale target reference: %v", err)
	}

	scale, targetGR, err := r.scaleForResourceMappings(ctx, mpa.Namespace, mpa.Spec.ScaleTargetRef.Name, mappings)
	if err != nil {
		klog.Errorf("%s: FailedGetScale - error: %v", v1.EventTypeWarning, err.Error())
		r.eventRecorder.Event(mpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
		setCondition(mpa, mpa_types.AbleToScale, v1.ConditionFalse, "FailedGetScale", "the MPA recommender was unable to get the target's current scale: %v", err)
		if err := r.updateStatusIfNeeded(ctx, mpaStatusOriginal, mpa); err != nil {
			klog.Errorf("Error updating MPA status: %v", err.Error())
			utilruntime.HandleError(err)
		}
		return fmt.Errorf("failed to query scale subresource for %s: %v", reference, err)
	}
	setCondition(mpa, mpa_types.AbleToScale, v1.ConditionTrue, "SucceededGetScale", "the MPA recommender was able to get the target's current scale")
	klog.V(4).Infof("MPA recommender was able to get the target's current scale = %d for targetGR %v", scale.Spec.Replicas, targetGR)
	currentReplicas := scale.Spec.Replicas
	r.recordInitialRecommendation(currentReplicas, key)

	var (
		metricStatuses        []autoscalingv2.MetricStatus
		metricDesiredReplicas int32
		metricName            string
	)

	desiredReplicas := int32(0)
	rescaleReason := ""

	var minReplicas int32

	if mpa.Spec.Constraints.MinReplicas != nil {
		minReplicas = *mpa.Spec.Constraints.MinReplicas
	} else {
		// Default value is 1.
		minReplicas = 1
	}

	rescale := true

	if scale.Spec.Replicas == 0 && minReplicas != 0 {
		// Autoscaling is disabled for this resource
		desiredReplicas = 0
		rescale = false
		setCondition(mpa, mpa_types.ScalingActive, v1.ConditionFalse, "ScalingDisabled", "scaling is disabled since the replica count of the target is zero")
		klog.V(4).Infof("Scaling is disabled since the replica count of the target is zero.")
	} else if currentReplicas > *mpa.Spec.Constraints.MaxReplicas {
		rescaleReason = "Current number of replicas above Spec.Constraints.MaxReplicas"
		desiredReplicas = *mpa.Spec.Constraints.MaxReplicas
		klog.V(4).Infof("Current number of replicas above Spec.Constraints.MaxReplicas.")
	} else if currentReplicas < minReplicas {
		rescaleReason = "Current number of replicas below Spec.Constraints.MinReplicas"
		desiredReplicas = minReplicas
		klog.V(4).Infof("Current number of replicas below Spec.Constraints.MinReplicas.")
	} else {
		var metricTimestamp time.Time
		metricDesiredReplicas, metricName, metricStatuses, metricTimestamp, err = r.computeReplicasForMetrics(ctx, mpa, scale, mpa.Spec.Metrics)
		if err != nil {
			r.setCurrentReplicasInStatus(mpa, currentReplicas)
			if err := r.updateStatusIfNeeded(ctx, mpaStatusOriginal, mpa); err != nil {
				klog.Errorf("Error updating MPA status: %v", err.Error())
				utilruntime.HandleError(err)
			}
			r.eventRecorder.Event(mpa, v1.EventTypeWarning, "FailedComputeMetricsReplicas", err.Error())
			klog.Errorf("%s: FailedComputeMetricsReplicas - error: %v", v1.EventTypeWarning, err.Error())
			return fmt.Errorf("failed to compute desired number of replicas based on listed metrics for %s: %v", reference, err)
		}

		klog.V(4).Infof("metricDesiredReplicas = %d desired replicas (based on %s from %s) for %s", metricDesiredReplicas, metricName, metricTimestamp, reference)

		rescaleMetric := ""
		if metricDesiredReplicas > desiredReplicas {
			desiredReplicas = metricDesiredReplicas
			rescaleMetric = metricName
		}
		if desiredReplicas > currentReplicas {
			rescaleReason = fmt.Sprintf("%s above target", rescaleMetric)
		}
		if desiredReplicas < currentReplicas {
			rescaleReason = "All metrics below target"
		}
		if mpa.Spec.Constraints.Behavior == nil {
			desiredReplicas = r.normalizeDesiredReplicas(mpa, key, currentReplicas, desiredReplicas, minReplicas)
		} else {
			desiredReplicas = r.normalizeDesiredReplicasWithBehaviors(mpa, key, currentReplicas, desiredReplicas, minReplicas)
		}
		rescale = desiredReplicas != currentReplicas
	}

	if rescale {
		// scale.Spec.Replicas = desiredReplicas
		// klog.V(4).Infof("Updating the number of replicas to %d for MPA %v", desiredReplicas, key)
		// _, err = r.controllerFetcher.Scales(mpa.Namespace).Update(ctx, targetGR, scale, metav1.UpdateOptions{})
		r.setCurrentReplicasInStatus(mpa, currentReplicas)
		if err := r.updateStatusIfNeeded(ctx, mpaStatusOriginal, mpa); err != nil {
			klog.Errorf("Error updating MPA status: %v", err.Error())
			utilruntime.HandleError(err)
		}
		// if err != nil {
		// 	r.eventRecorder.Eventf(mpa, v1.EventTypeWarning, "FailedRescale", "New size: %d; reason: %s; error: %v", desiredReplicas, rescaleReason, err.Error())
		// 	klog.Errorf("%s: FailedRescale - New size: %d; reason: %s; error: %v", v1.EventTypeWarning, desiredReplicas, rescaleReason, err.Error())
		// 	setCondition(mpa, mpa_types.AbleToScale, v1.ConditionFalse, "FailedUpdateScale", "the MPA controller was unable to update the target scale: %v", err)
		// 	r.setCurrentReplicasInStatus(mpa, currentReplicas)
		// 	if err := r.updateStatusIfNeeded(ctx, mpaStatusOriginal, mpa); err != nil {
		// 		klog.Errorf("Error updating MPA status: %v", err.Error())
		// 		utilruntime.HandleError(err)
		// 	}
		// 	return fmt.Errorf("failed to rescale %s: %v", reference, err)
		// }
		setCondition(mpa, mpa_types.AbleToScale, v1.ConditionTrue, "SucceededRescale", "the MPA controller was able to update the target scale to %d", desiredReplicas)
		r.eventRecorder.Eventf(mpa, v1.EventTypeNormal, "SuccessfulRescale", "New size: %d; reason: %s", desiredReplicas, rescaleReason)
		// klog.V(4).Infof("%s: Successfully rescaled the number of replicas to %d for MPA %v", v1.EventTypeNormal, desiredReplicas, key)
		r.storeScaleEvent(mpa.Spec.Constraints.Behavior, key, currentReplicas, desiredReplicas)
		// klog.Infof("Successful rescaled of %s, old size: %d, new size: %d, reason: %s", mpa.Name, currentReplicas, desiredReplicas, rescaleReason)
	} else {
		klog.V(4).Infof("decided not to scale %s to %v (reason: %s) (the last scale time was %s)", reference, desiredReplicas, rescaleReason, mpa.Status.LastScaleTime)
		desiredReplicas = currentReplicas
	}

	r.setStatus(mpa, currentReplicas, desiredReplicas, metricStatuses, rescale)
	return r.updateStatusIfNeeded(ctx, mpaStatusOriginal, mpa)
}

// setCondition sets the specific condition type on the given MPA to the specified value with the
// given reason and message. The message and args are treated like a format string. The condition 
// will be added if it is not present.
func setCondition(mpa *mpa_types.MultidimPodAutoscaler, conditionType mpa_types.MultidimPodAutoscalerConditionType, status v1.ConditionStatus, reason, message string, args ...interface{}) {
	mpa.Status.Conditions = setConditionInList(mpa.Status.Conditions, conditionType, status, reason, message, args...)
}

// setConditionInList sets the specific condition type on the given MPA to the specified value with 
// the given reason and message. The message and args are treated like a format string. The
// condition will be added if it is not present.  The new list will be returned.
func setConditionInList(inputList []mpa_types.MultidimPodAutoscalerCondition, conditionType mpa_types.MultidimPodAutoscalerConditionType, status v1.ConditionStatus, reason, message string, args ...interface{}) []mpa_types.MultidimPodAutoscalerCondition {
	resList := inputList
	var existingCond *mpa_types.MultidimPodAutoscalerCondition
	for i, condition := range resList {
		if condition.Type == conditionType {
			// can't take a pointer to an iteration variable
			existingCond = &resList[i]
			break
		}
	}

	if existingCond == nil {
		resList = append(resList, mpa_types.MultidimPodAutoscalerCondition{
			Type: conditionType,
		})
		existingCond = &resList[len(resList)-1]
	}

	if existingCond.Status != status {
		existingCond.LastTransitionTime = metav1.Now()
	}

	existingCond.Status = status
	existingCond.Reason = reason
	existingCond.Message = fmt.Sprintf(message, args...)

	return resList
}

// updateStatusIfNeeded calls updateStatus only if the status of the new MPA is not the same as the
// old status
func (r *recommender) updateStatusIfNeeded(ctx context.Context, oldStatus *mpa_types.MultidimPodAutoscalerStatus, newMPA *mpa_types.MultidimPodAutoscaler) error {
	// skip a write if we wouldn't need to update
	if apiequality.Semantic.DeepEqual(oldStatus, &newMPA.Status) {
		return nil
	}

	patches := []patchRecord{{
		Op:    "add",
		Path:  "/status",
		Value: newMPA.Status,
	}}
	klog.V(4).Infof("Updating MPA status with the desired number of replicas = %d", newMPA.Status.DesiredReplicas)

	return patchMpa(r.mpaClient.MultidimPodAutoscalers(newMPA.Namespace), newMPA.Name, patches)
}

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func patchMpa(mpaClient mpa_api.MultidimPodAutoscalerInterface, mpaName string, patches []patchRecord) (err error) {
	bytes, err := json.Marshal(patches)
	if err != nil {
		klog.Errorf("Cannot marshal MPA status patches %+v. Reason: %+v", patches, err)
		return
	}

	updatedMPA, err := mpaClient.Patch(context.TODO(), mpaName, types.JSONPatchType, bytes, metav1.PatchOptions{})

	klog.V(4).Infof("MPA %s status updated (desiredReplicas = %d)", updatedMPA.Name, updatedMPA.Status.DesiredReplicas)

	return err
}

// updateStatus actually does the update request for the status of the given MPA
func (r *recommender) updateStatus(ctx context.Context, mpa *mpa_types.MultidimPodAutoscaler) error {
	_, err := r.mpaClient.MultidimPodAutoscalers(mpa.Namespace).UpdateStatus(ctx, mpa, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("%s: FailedUpdateStatus - error updating status for MPA %s (namespace %s): %v", v1.EventTypeWarning, mpa.Name, mpa.Namespace, err.Error())
		r.eventRecorder.Event(mpa, v1.EventTypeWarning, "FailedUpdateStatus", err.Error())
		return fmt.Errorf("failed to update status for %s: %v", mpa.Name, err)
	}
	klog.V(2).Infof("Successfully updated status (HPA-related) for %s", mpa.Name)
	return nil
}

// scaleForResourceMappings attempts to fetch the scale for the resource with the given name and
// namespace, trying each RESTMapping in turn until a working one is found. If none work, the first
// error is returned. It returns both the scale, as well as the group-resource from the working
// mapping.
func (r *recommender) scaleForResourceMappings(ctx context.Context, namespace, name string, mappings []*apimeta.RESTMapping) (*autoscalingv1.Scale, schema.GroupResource, error) {
	var firstErr error
	for i, mapping := range mappings {
		targetGR := mapping.Resource.GroupResource()
		scale, err := r.controllerFetcher.Scales(namespace).Get(ctx, targetGR, name, metav1.GetOptions{})
		if err == nil {
			return scale, targetGR, nil
		}

		// if this is the first error, remember it,
		// then go on and try other mappings until we find a good one
		if i == 0 {
			firstErr = err
		}
	}

	// make sure we handle an empty set of mappings
	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, schema.GroupResource{}, firstErr
}

func (r *recommender) recordInitialRecommendation(currentReplicas int32, key model.MpaID) {
	r.recommendationsLock.Lock()
	defer r.recommendationsLock.Unlock()
	if r.recommendations[key] == nil {
		r.recommendations[key] = []timestampedRecommendation{{currentReplicas, time.Now()}}
	}
}

// computeReplicasForMetrics computes the desired number of replicas for the metric specifications
// listed in the MPA, returning the maximum of the computed replica counts, a description of the
// associated metric, and the statuses of all metrics computed.
func (r *recommender) computeReplicasForMetrics(ctx context.Context, mpa *mpa_types.MultidimPodAutoscaler, scale *autoscalingv1.Scale, metricSpecs []autoscalingv2.MetricSpec) (replicas int32, metric string, statuses []autoscalingv2.MetricStatus, timestamp time.Time, err error) {
	if scale.Status.Selector == "" {
		errMsg := "selector is required"
		r.eventRecorder.Event(mpa, v1.EventTypeWarning, "SelectorRequired", errMsg)
		klog.Errorf("%s: SelectorRequired", v1.EventTypeWarning)
		setCondition(mpa, mpa_types.ScalingActive, v1.ConditionFalse, "InvalidSelector", "the MPA target's scale is missing a selector")
		return 0, "", nil, time.Time{}, fmt.Errorf(errMsg)
	}

	selector, err := labels.Parse(scale.Status.Selector)
	if err != nil {
		errMsg := fmt.Sprintf("couldn't convert selector into a corresponding internal selector object: %v", err)
		r.eventRecorder.Event(mpa, v1.EventTypeWarning, "InvalidSelector", errMsg)
		klog.Errorf("%s: InvalidSelector - error: %v", v1.EventTypeWarning, errMsg)
		setCondition(mpa, mpa_types.ScalingActive, v1.ConditionFalse, "InvalidSelector", errMsg)
		return 0, "", nil, time.Time{}, fmt.Errorf(errMsg)
	}
	klog.V(4).Infof("Label Selector parsed as %v", selector)

	specReplicas := scale.Spec.Replicas
	statusReplicas := scale.Status.Replicas
	statuses = make([]autoscalingv2.MetricStatus, len(metricSpecs))

	invalidMetricsCount := 0
	var invalidMetricError error
	var invalidMetricCondition mpa_types.MultidimPodAutoscalerCondition

	for i, metricSpec := range metricSpecs {
		replicaCountProposal, metricNameProposal, timestampProposal, condition, err := r.computeReplicasForMetric(ctx, mpa, metricSpec, specReplicas, statusReplicas, selector, &statuses[i])

		if err != nil {
			if invalidMetricsCount <= 0 {
				invalidMetricCondition = condition
				invalidMetricError = err
			}
			invalidMetricsCount++
		}
		if err == nil && (replicas == 0 || replicaCountProposal > replicas) {
			timestamp = timestampProposal
			replicas = replicaCountProposal
			metric = metricNameProposal
		}
	}

	// If all metrics are invalid or some are invalid and we would scale down,
	// return an error and set the condition of the MPA based on the first invalid metric.
	// Otherwise set the condition as scaling active as we're going to scale
	if invalidMetricsCount >= len(metricSpecs) || (invalidMetricsCount > 0 && replicas < specReplicas) {
		setCondition(mpa, mpa_types.AbleToScale, invalidMetricCondition.Status, invalidMetricCondition.Reason, invalidMetricCondition.Message)
		return 0, "", statuses, time.Time{}, fmt.Errorf("invalid metrics (%v invalid out of %v), first error is: %v", invalidMetricsCount, len(metricSpecs), invalidMetricError)
	}
	setCondition(mpa, mpa_types.ScalingActive, v1.ConditionTrue, "ValidMetricFound", "the MPA was able to successfully calculate a replica count from %s", metric)
	return replicas, metric, statuses, timestamp, nil
}

// Computes the desired number of replicas for a specific MPA and metric specification,
// returning the metric status and a proposed condition to be set on the MPA object.
func (r *recommender) computeReplicasForMetric(ctx context.Context, mpa *mpa_types.MultidimPodAutoscaler, spec autoscalingv2.MetricSpec, specReplicas, statusReplicas int32, selector labels.Selector, status *autoscalingv2.MetricStatus) (replicaCountProposal int32, metricNameProposal string, timestampProposal time.Time, condition mpa_types.MultidimPodAutoscalerCondition, err error) {
	switch spec.Type {
	case autoscalingv2.ObjectMetricSourceType:
		klog.V(4).Infof("Pulling metrics from the source of type ObjectMetricSourceType")
		metricSelector, err := metav1.LabelSelectorAsSelector(spec.Object.Metric.Selector)
		if err != nil {
			condition := r.getUnableComputeReplicaCountCondition(mpa, "FailedGetObjectMetric", err)
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get object metric value: %v", err)
		}
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err = r.computeStatusForObjectMetric(specReplicas, statusReplicas, spec, mpa, selector, status, metricSelector)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get object metric value: %v", err)
		}
	case autoscalingv2.PodsMetricSourceType:
		klog.V(4).Infof("Pulling metrics from the source of type PodMetricSourceType")
		metricSelector, err := metav1.LabelSelectorAsSelector(spec.Pods.Metric.Selector)
		if err != nil {
			condition := r.getUnableComputeReplicaCountCondition(mpa, "FailedGetPodsMetric", err)
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get pods metric value: %v", err)
		}
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err = r.computeStatusForPodsMetric(specReplicas, spec, mpa, selector, status, metricSelector)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get pods metric value: %v", err)
		}
	case autoscalingv2.ResourceMetricSourceType:
		klog.V(4).Infof("Pulling metrics from the source of type ResourceMetricSourceType")
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err = r.computeStatusForResourceMetric(ctx, specReplicas, spec, mpa, selector, status)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get %s resource metric value: %v", spec.Resource.Name, err)
		}
	case autoscalingv2.ContainerResourceMetricSourceType:
		klog.V(4).Infof("Pulling metrics from the source of type ContainerResourceMetricSourceType")
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err = r.computeStatusForContainerResourceMetric(ctx, specReplicas, spec, mpa, selector, status)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get %s container metric value: %v", spec.ContainerResource.Container, err)
		}
	case autoscalingv2.ExternalMetricSourceType:
		klog.V(4).Infof("Pulling metrics from the source of type ExternalMetricSourceType")
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err = r.computeStatusForExternalMetric(specReplicas, statusReplicas, spec, mpa, selector, status)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get %s external metric value: %v", spec.External.Metric.Name, err)
		}
	default:
		klog.Warningf("Unknown metric source type!")
		errMsg := fmt.Sprintf("unknown metric source type %q", string(spec.Type))
		err = fmt.Errorf(errMsg)
		condition := r.getUnableComputeReplicaCountCondition(mpa, "InvalidMetricSourceType", err)
		return 0, "", time.Time{}, condition, err
	}
	return replicaCountProposal, metricNameProposal, timestampProposal, mpa_types.MultidimPodAutoscalerCondition{}, nil
}

func (r *recommender) getUnableComputeReplicaCountCondition(mpa runtime.Object, reason string, err error) (condition mpa_types.MultidimPodAutoscalerCondition) {
	r.eventRecorder.Event(mpa, v1.EventTypeWarning, reason, err.Error())
	klog.Errorf("%s: %s - error: %v", v1.EventTypeWarning, reason, err.Error())
	return mpa_types.MultidimPodAutoscalerCondition{
		Type:    mpa_types.ScalingActive,
		Status:  v1.ConditionFalse,
		Reason:  reason,
		Message: fmt.Sprintf("the MPA was unable to compute the replica count: %v", err),
	}
}

// computeStatusForObjectMetric computes the desired number of replicas for the specified metric of type ObjectMetricSourceType.
func (r *recommender) computeStatusForObjectMetric(specReplicas, statusReplicas int32, metricSpec autoscalingv2.MetricSpec, mpa *mpa_types.MultidimPodAutoscaler, selector labels.Selector, status *autoscalingv2.MetricStatus, metricSelector labels.Selector) (replicas int32, timestamp time.Time, metricName string, condition mpa_types.MultidimPodAutoscalerCondition, err error) {
	if metricSpec.Object.Target.Type == autoscalingv2.ValueMetricType {
		replicaCountProposal, usageProposal, timestampProposal, err := r.replicaCalc.GetObjectMetricReplicas(specReplicas, metricSpec.Object.Target.Value.MilliValue(), metricSpec.Object.Metric.Name, mpa.Namespace, &metricSpec.Object.DescribedObject, selector, metricSelector)
		if err != nil {
			condition := r.getUnableComputeReplicaCountCondition(mpa, "FailedGetObjectMetric", err)
			return 0, timestampProposal, "", condition, err
		}
		*status = autoscalingv2.MetricStatus{
			Type: autoscalingv2.ObjectMetricSourceType,
			Object: &autoscalingv2.ObjectMetricStatus{
				DescribedObject: metricSpec.Object.DescribedObject,
				Metric: autoscalingv2.MetricIdentifier{
					Name:     metricSpec.Object.Metric.Name,
					Selector: metricSpec.Object.Metric.Selector,
				},
				Current: autoscalingv2.MetricValueStatus{
					Value: resource.NewMilliQuantity(usageProposal, resource.DecimalSI),
				},
			},
		}
		return replicaCountProposal, timestampProposal, fmt.Sprintf("%s metric %s", metricSpec.Object.DescribedObject.Kind, metricSpec.Object.Metric.Name), mpa_types.MultidimPodAutoscalerCondition{}, nil
	} else if metricSpec.Object.Target.Type == autoscalingv2.AverageValueMetricType {
		replicaCountProposal, usageProposal, timestampProposal, err := r.replicaCalc.GetObjectPerPodMetricReplicas(statusReplicas, metricSpec.Object.Target.AverageValue.MilliValue(), metricSpec.Object.Metric.Name, mpa.Namespace, &metricSpec.Object.DescribedObject, metricSelector)
		if err != nil {
			condition := r.getUnableComputeReplicaCountCondition(mpa, "FailedGetObjectMetric", err)
			return 0, time.Time{}, "", condition, fmt.Errorf("failed to get %s object metric: %v", metricSpec.Object.Metric.Name, err)
		}
		*status = autoscalingv2.MetricStatus{
			Type: autoscalingv2.ObjectMetricSourceType,
			Object: &autoscalingv2.ObjectMetricStatus{
				Metric: autoscalingv2.MetricIdentifier{
					Name:     metricSpec.Object.Metric.Name,
					Selector: metricSpec.Object.Metric.Selector,
				},
				Current: autoscalingv2.MetricValueStatus{
					AverageValue: resource.NewMilliQuantity(usageProposal, resource.DecimalSI),
				},
			},
		}
		return replicaCountProposal, timestampProposal, fmt.Sprintf("external metric %s(%+v)", metricSpec.Object.Metric.Name, metricSpec.Object.Metric.Selector), mpa_types.MultidimPodAutoscalerCondition{}, nil
	}
	errMsg := "invalid object metric source: neither a value target nor an average value target was set"
	err = fmt.Errorf(errMsg)
	condition = r.getUnableComputeReplicaCountCondition(mpa, "FailedGetObjectMetric", err)
	return 0, time.Time{}, "", condition, err
}

// computeStatusForPodsMetric computes the desired number of replicas for the specified metric of type PodsMetricSourceType.
func (r *recommender) computeStatusForPodsMetric(currentReplicas int32, metricSpec autoscalingv2.MetricSpec, mpa *mpa_types.MultidimPodAutoscaler, selector labels.Selector, status *autoscalingv2.MetricStatus, metricSelector labels.Selector) (replicaCountProposal int32, timestampProposal time.Time, metricNameProposal string, condition mpa_types.MultidimPodAutoscalerCondition, err error) {
	replicaCountProposal, usageProposal, timestampProposal, err := r.replicaCalc.GetMetricReplicas(currentReplicas, metricSpec.Pods.Target.AverageValue.MilliValue(), metricSpec.Pods.Metric.Name, mpa.Namespace, selector, metricSelector)
	if err != nil {
		condition = r.getUnableComputeReplicaCountCondition(mpa, "FailedGetPodsMetric", err)
		return 0, timestampProposal, "", condition, err
	}
	*status = autoscalingv2.MetricStatus{
		Type: autoscalingv2.PodsMetricSourceType,
		Pods: &autoscalingv2.PodsMetricStatus{
			Metric: autoscalingv2.MetricIdentifier{
				Name:     metricSpec.Pods.Metric.Name,
				Selector: metricSpec.Pods.Metric.Selector,
			},
			Current: autoscalingv2.MetricValueStatus{
				AverageValue: resource.NewMilliQuantity(usageProposal, resource.DecimalSI),
			},
		},
	}

	return replicaCountProposal, timestampProposal, fmt.Sprintf("pods metric %s", metricSpec.Pods.Metric.Name), mpa_types.MultidimPodAutoscalerCondition{}, nil
}

func (r *recommender) computeStatusForResourceMetricGeneric(ctx context.Context, currentReplicas int32, target autoscalingv2.MetricTarget, resourceName v1.ResourceName, namespace string, container string, selector labels.Selector) (replicaCountProposal int32, metricStatus *autoscalingv2.MetricValueStatus, timestampProposal time.Time, metricNameProposal string, condition mpa_types.MultidimPodAutoscalerCondition, err error) {
	if target.AverageValue != nil {
		var rawProposal int64
		replicaCountProposal, rawProposal, timestampProposal, err := r.replicaCalc.GetRawResourceReplicas(ctx, currentReplicas, target.AverageValue.MilliValue(), resourceName, namespace, selector, container)
		if err != nil {
			return 0, nil, time.Time{}, "", condition, fmt.Errorf("failed to get %s usage: %v", resourceName, err)
		}
		metricNameProposal = fmt.Sprintf("%s resource", resourceName.String())
		status := autoscalingv2.MetricValueStatus{
			AverageValue: resource.NewMilliQuantity(rawProposal, resource.DecimalSI),
		}
		return replicaCountProposal, &status, timestampProposal, metricNameProposal, mpa_types.MultidimPodAutoscalerCondition{}, nil
	}

	if target.AverageUtilization == nil {
		errMsg := "invalid resource metric source: neither an average utilization target nor an average value (usage) target was set"
		return 0, nil, time.Time{}, "", condition, fmt.Errorf(errMsg)
	}

	targetUtilization := *target.AverageUtilization
	replicaCountProposal, percentageProposal, rawProposal, timestampProposal, err := r.replicaCalc.GetResourceReplicas(ctx, currentReplicas, targetUtilization, resourceName, namespace, selector, container)
	if err != nil {
		return 0, nil, time.Time{}, "", condition, fmt.Errorf("failed to get %s utilization: %v", resourceName, err)
	}

	metricNameProposal = fmt.Sprintf("%s resource utilization (percentage of request)", resourceName)
	status := autoscalingv2.MetricValueStatus{
		AverageUtilization: &percentageProposal,
		AverageValue:       resource.NewMilliQuantity(rawProposal, resource.DecimalSI),
	}
	klog.V(4).Infof("Current average utilization = %d average value = %v", percentageProposal, status.AverageValue)
	return replicaCountProposal, &status, timestampProposal, metricNameProposal, mpa_types.MultidimPodAutoscalerCondition{}, nil
}

// computeStatusForResourceMetric computes the desired number of replicas for the specified metric of type ResourceMetricSourceType.
func (r *recommender) computeStatusForResourceMetric(ctx context.Context, currentReplicas int32, metricSpec autoscalingv2.MetricSpec, mpa *mpa_types.MultidimPodAutoscaler, selector labels.Selector, status *autoscalingv2.MetricStatus) (replicaCountProposal int32, timestampProposal time.Time, metricNameProposal string, condition mpa_types.MultidimPodAutoscalerCondition, err error) {
	replicaCountProposal, metricValueStatus, timestampProposal, metricNameProposal, condition, err := r.computeStatusForResourceMetricGeneric(ctx, currentReplicas, metricSpec.Resource.Target, metricSpec.Resource.Name, mpa.Namespace, "", selector)
	if err != nil {
		condition = r.getUnableComputeReplicaCountCondition(mpa, "FailedGetResourceMetric", err)
		return replicaCountProposal, timestampProposal, metricNameProposal, condition, err
	}
	*status = autoscalingv2.MetricStatus{
		Type: autoscalingv2.ResourceMetricSourceType,
		Resource: &autoscalingv2.ResourceMetricStatus{
			Name:    metricSpec.Resource.Name,
			Current: *metricValueStatus,
		},
	}
	return replicaCountProposal, timestampProposal, metricNameProposal, condition, nil
}

// computeStatusForContainerResourceMetric computes the desired number of replicas for the specified metric of type ResourceMetricSourceType.
func (r *recommender) computeStatusForContainerResourceMetric(ctx context.Context, currentReplicas int32, metricSpec autoscalingv2.MetricSpec, mpa *mpa_types.MultidimPodAutoscaler, selector labels.Selector, status *autoscalingv2.MetricStatus) (replicaCountProposal int32, timestampProposal time.Time, metricNameProposal string, condition mpa_types.MultidimPodAutoscalerCondition, err error) {
	replicaCountProposal, metricValueStatus, timestampProposal, metricNameProposal, condition, err := r.computeStatusForResourceMetricGeneric(ctx, currentReplicas, metricSpec.ContainerResource.Target, metricSpec.ContainerResource.Name, mpa.Namespace, metricSpec.ContainerResource.Container, selector)
	if err != nil {
		condition = r.getUnableComputeReplicaCountCondition(mpa, "FailedGetContainerResourceMetric", err)
		return replicaCountProposal, timestampProposal, metricNameProposal, condition, err
	}
	*status = autoscalingv2.MetricStatus{
		Type: autoscalingv2.ContainerResourceMetricSourceType,
		ContainerResource: &autoscalingv2.ContainerResourceMetricStatus{
			Name:      metricSpec.ContainerResource.Name,
			Container: metricSpec.ContainerResource.Container,
			Current:   *metricValueStatus,
		},
	}
	return replicaCountProposal, timestampProposal, metricNameProposal, condition, nil
}

// computeStatusForExternalMetric computes the desired number of replicas for the specified metric of type ExternalMetricSourceType.
func (r *recommender) computeStatusForExternalMetric(specReplicas, statusReplicas int32, metricSpec autoscalingv2.MetricSpec, mpa *mpa_types.MultidimPodAutoscaler, selector labels.Selector, status *autoscalingv2.MetricStatus) (replicaCountProposal int32, timestampProposal time.Time, metricNameProposal string, condition mpa_types.MultidimPodAutoscalerCondition, err error) {
	if metricSpec.External.Target.AverageValue != nil {
		replicaCountProposal, usageProposal, timestampProposal, err := r.replicaCalc.GetExternalPerPodMetricReplicas(statusReplicas, metricSpec.External.Target.AverageValue.MilliValue(), metricSpec.External.Metric.Name, mpa.Namespace, metricSpec.External.Metric.Selector)
		if err != nil {
			condition = r.getUnableComputeReplicaCountCondition(mpa, "FailedGetExternalMetric", err)
			return 0, time.Time{}, "", condition, fmt.Errorf("failed to get %s external metric: %v", metricSpec.External.Metric.Name, err)
		}
		*status = autoscalingv2.MetricStatus{
			Type: autoscalingv2.ExternalMetricSourceType,
			External: &autoscalingv2.ExternalMetricStatus{
				Metric: autoscalingv2.MetricIdentifier{
					Name:     metricSpec.External.Metric.Name,
					Selector: metricSpec.External.Metric.Selector,
				},
				Current: autoscalingv2.MetricValueStatus{
					AverageValue: resource.NewMilliQuantity(usageProposal, resource.DecimalSI),
				},
			},
		}
		return replicaCountProposal, timestampProposal, fmt.Sprintf("external metric %s(%+v)", metricSpec.External.Metric.Name, metricSpec.External.Metric.Selector), mpa_types.MultidimPodAutoscalerCondition{}, nil
	}
	if metricSpec.External.Target.Value != nil {
		replicaCountProposal, usageProposal, timestampProposal, err := r.replicaCalc.GetExternalMetricReplicas(specReplicas, metricSpec.External.Target.Value.MilliValue(), metricSpec.External.Metric.Name, mpa.Namespace, metricSpec.External.Metric.Selector, selector)
		if err != nil {
			condition = r.getUnableComputeReplicaCountCondition(mpa, "FailedGetExternalMetric", err)
			return 0, time.Time{}, "", condition, fmt.Errorf("failed to get external metric %s: %v", metricSpec.External.Metric.Name, err)
		}
		*status = autoscalingv2.MetricStatus{
			Type: autoscalingv2.ExternalMetricSourceType,
			External: &autoscalingv2.ExternalMetricStatus{
				Metric: autoscalingv2.MetricIdentifier{
					Name:     metricSpec.External.Metric.Name,
					Selector: metricSpec.External.Metric.Selector,
				},
				Current: autoscalingv2.MetricValueStatus{
					Value: resource.NewMilliQuantity(usageProposal, resource.DecimalSI),
				},
			},
		}
		return replicaCountProposal, timestampProposal, fmt.Sprintf("external metric %s(%+v)", metricSpec.External.Metric.Name, metricSpec.External.Metric.Selector), mpa_types.MultidimPodAutoscalerCondition{}, nil
	}
	errMsg := "invalid external metric source: neither a value target nor an average value target was set"
	err = fmt.Errorf(errMsg)
	condition = r.getUnableComputeReplicaCountCondition(mpa, "FailedGetExternalMetric", err)
	return 0, time.Time{}, "", condition, fmt.Errorf(errMsg)
}

// setCurrentReplicasInStatus sets the current replica count in the status of the MPA.
func (r *recommender) setCurrentReplicasInStatus(mpa *mpa_types.MultidimPodAutoscaler, currentReplicas int32) {
	r.setStatus(mpa, currentReplicas, mpa.Status.DesiredReplicas, mpa.Status.CurrentMetrics, false)
}

// setStatus recreates the status of the given MPA, updating the current and desired replicas, as
// well as the metric statuses
func (r *recommender) setStatus(mpa *mpa_types.MultidimPodAutoscaler, currentReplicas, desiredReplicas int32, metricStatuses []autoscalingv2.MetricStatus, rescale bool) {
	mpa.Status = mpa_types.MultidimPodAutoscalerStatus{
		CurrentReplicas: currentReplicas,
		DesiredReplicas: desiredReplicas,
		LastScaleTime:   mpa.Status.LastScaleTime,
		CurrentMetrics:  metricStatuses,
		Conditions:      mpa.Status.Conditions,
		// Keep VPA-related untouched.
		Recommendation:  mpa.Status.Recommendation,
	}

	if rescale {
		now := metav1.NewTime(time.Now())
		mpa.Status.LastScaleTime = &now
	}
}

// normalizeDesiredReplicas takes the metrics desired replicas value and normalizes it based on the appropriate conditions (i.e., < maxReplicas, > minReplicas, etc...)
func (r *recommender) normalizeDesiredReplicas(mpa *mpa_types.MultidimPodAutoscaler, key model.MpaID, currentReplicas int32, prenormalizedDesiredReplicas int32, minReplicas int32) int32 {
	stabilizedRecommendation := r.stabilizeRecommendation(key, prenormalizedDesiredReplicas)
	if stabilizedRecommendation != prenormalizedDesiredReplicas {
		setCondition(mpa, mpa_types.AbleToScale, v1.ConditionTrue, "ScaleDownStabilized", "recent recommendations were higher than current one, applying the highest recent recommendation")
	} else {
		setCondition(mpa, mpa_types.AbleToScale, v1.ConditionTrue, "ReadyForNewScale", "recommended size matches current size")
	}

	desiredReplicas, condition, reason := convertDesiredReplicasWithRules(currentReplicas, stabilizedRecommendation, minReplicas, *mpa.Spec.Constraints.MaxReplicas)

	if desiredReplicas == stabilizedRecommendation {
		setCondition(mpa, mpa_types.ScalingLimited, v1.ConditionFalse, condition, reason)
	} else {
		setCondition(mpa, mpa_types.ScalingLimited, v1.ConditionTrue, condition, reason)
	}

	return desiredReplicas
}

// convertDesiredReplicas performs the actual normalization, without depending on `HorizontalController` or `HorizontalPodAutoscaler`
func convertDesiredReplicasWithRules(currentReplicas, desiredReplicas, hpaMinReplicas, hpaMaxReplicas int32) (int32, string, string) {
	var minimumAllowedReplicas int32
	var maximumAllowedReplicas int32

	var possibleLimitingCondition string
	var possibleLimitingReason string

	minimumAllowedReplicas = hpaMinReplicas

	// Do not scaleup too much to prevent incorrect rapid increase of the number of master replicas
	// caused by bogus CPU usage report from heapster/kubelet (like in issue #32304).
	scaleUpLimit := calculateScaleUpLimit(currentReplicas)

	if hpaMaxReplicas > scaleUpLimit {
		maximumAllowedReplicas = scaleUpLimit
		possibleLimitingCondition = "ScaleUpLimit"
		possibleLimitingReason = "the desired replica count is increasing faster than the maximum scale rate"
	} else {
		maximumAllowedReplicas = hpaMaxReplicas
		possibleLimitingCondition = "TooManyReplicas"
		possibleLimitingReason = "the desired replica count is more than the maximum replica count"
	}

	if desiredReplicas < minimumAllowedReplicas {
		possibleLimitingCondition = "TooFewReplicas"
		possibleLimitingReason = "the desired replica count is less than the minimum replica count"
		return minimumAllowedReplicas, possibleLimitingCondition, possibleLimitingReason
	} else if desiredReplicas > maximumAllowedReplicas {
		return maximumAllowedReplicas, possibleLimitingCondition, possibleLimitingReason
	}

	return desiredReplicas, "DesiredWithinRange", "the desired count is within the acceptable range"
}

func calculateScaleUpLimit(currentReplicas int32) int32 {
	return int32(math.Max(scaleUpLimitFactor*float64(currentReplicas), scaleUpLimitMinimum))
}

// convertDesiredReplicasWithBehaviorRate performs the actual normalization given the constraint
// rate. It doesn't consider the stabilizationWindow, it is done separately.
func (r *recommender) convertDesiredReplicasWithBehaviorRate(args NormalizationArg) (int32, string, string) {
	var possibleLimitingReason, possibleLimitingMessage string

	if args.DesiredReplicas > args.CurrentReplicas {
		r.scaleUpEventsLock.RLock()
		defer r.scaleUpEventsLock.RUnlock()
		r.scaleDownEventsLock.RLock()
		defer r.scaleDownEventsLock.RUnlock()
		scaleUpLimit := calculateScaleUpLimitWithScalingRules(args.CurrentReplicas, r.scaleUpEvents[args.Key], r.scaleDownEvents[args.Key], args.ScaleUpBehavior)

		if scaleUpLimit < args.CurrentReplicas {
			// We shouldn't scale up further until the scaleUpEvents will be cleaned up
			scaleUpLimit = args.CurrentReplicas
		}
		maximumAllowedReplicas := args.MaxReplicas
		if maximumAllowedReplicas > scaleUpLimit {
			maximumAllowedReplicas = scaleUpLimit
			possibleLimitingReason = "ScaleUpLimit"
			possibleLimitingMessage = "the desired replica count is increasing faster than the maximum scale rate"
		} else {
			possibleLimitingReason = "TooManyReplicas"
			possibleLimitingMessage = "the desired replica count is more than the maximum replica count"
		}
		if args.DesiredReplicas > maximumAllowedReplicas {
			return maximumAllowedReplicas, possibleLimitingReason, possibleLimitingMessage
		}
	} else if args.DesiredReplicas < args.CurrentReplicas {
		r.scaleUpEventsLock.RLock()
		defer r.scaleUpEventsLock.RUnlock()
		r.scaleDownEventsLock.RLock()
		defer r.scaleDownEventsLock.RUnlock()
		scaleDownLimit := calculateScaleDownLimitWithBehaviors(args.CurrentReplicas, r.scaleUpEvents[args.Key], r.scaleDownEvents[args.Key], args.ScaleDownBehavior)

		if scaleDownLimit > args.CurrentReplicas {
			// We shouldn't scale down further until the scaleDownEvents will be cleaned up
			scaleDownLimit = args.CurrentReplicas
		}
		minimumAllowedReplicas := args.MinReplicas
		if minimumAllowedReplicas < scaleDownLimit {
			minimumAllowedReplicas = scaleDownLimit
			possibleLimitingReason = "ScaleDownLimit"
			possibleLimitingMessage = "the desired replica count is decreasing faster than the maximum scale rate"
		} else {
			possibleLimitingMessage = "the desired replica count is less than the minimum replica count"
			possibleLimitingReason = "TooFewReplicas"
		}
		if args.DesiredReplicas < minimumAllowedReplicas {
			return minimumAllowedReplicas, possibleLimitingReason, possibleLimitingMessage
		}
	}
	return args.DesiredReplicas, "DesiredWithinRange", "the desired count is within the acceptable range"
}

// stabilizeRecommendation:
// - replaces old recommendation with the newest recommendation,
// - returns max of recommendations that are not older than downscaleStabilisationWindow.
func (r *recommender) stabilizeRecommendation(key model.MpaID, prenormalizedDesiredReplicas int32) int32 {
	maxRecommendation := prenormalizedDesiredReplicas
	foundOldSample := false
	oldSampleIndex := 0
	cutoff := time.Now().Add(-r.downscaleStabilisationWindow)

	r.recommendationsLock.Lock()
	defer r.recommendationsLock.Unlock()
	for i, rec := range r.recommendations[key] {
		if rec.timestamp.Before(cutoff) {
			foundOldSample = true
			oldSampleIndex = i
		} else if rec.recommendation > maxRecommendation {
			maxRecommendation = rec.recommendation
		}
	}
	if foundOldSample {
		r.recommendations[key][oldSampleIndex] = timestampedRecommendation{prenormalizedDesiredReplicas, time.Now()}
	} else {
		r.recommendations[key] = append(r.recommendations[key], timestampedRecommendation{prenormalizedDesiredReplicas, time.Now()})
	}
	return maxRecommendation
}

// normalizeDesiredReplicasWithBehaviors takes the metrics desired replicas value and normalizes it:
// 1. Apply the basic conditions (i.e. < maxReplicas, > minReplicas, etc...)
// 2. Apply the scale up/down limits from the mpaSpec.Constraints.Behaviors (i.e., add no more than 4 pods)
// 3. Apply the constraints period (i.e., add no more than 4 pods per minute)
// 4. Apply the stabilization (i.e., add no more than 4 pods per minute, and pick the smallest recommendation during last 5 minutes)
func (r *recommender) normalizeDesiredReplicasWithBehaviors(mpa *mpa_types.MultidimPodAutoscaler, key model.MpaID, currentReplicas, prenormalizedDesiredReplicas, minReplicas int32) int32 {
	r.maybeInitScaleDownStabilizationWindow(mpa)
	normalizationArg := NormalizationArg{
		Key:               key,
		ScaleUpBehavior:   mpa.Spec.Constraints.Behavior.ScaleUp,
		ScaleDownBehavior: mpa.Spec.Constraints.Behavior.ScaleDown,
		MinReplicas:       minReplicas,
		MaxReplicas:       *mpa.Spec.Constraints.MaxReplicas,
		CurrentReplicas:   currentReplicas,
		DesiredReplicas:   prenormalizedDesiredReplicas}
	stabilizedRecommendation, reason, message := r.stabilizeRecommendationWithBehaviors(normalizationArg)
	normalizationArg.DesiredReplicas = stabilizedRecommendation
	if stabilizedRecommendation != prenormalizedDesiredReplicas {
		// "ScaleUpStabilized" || "ScaleDownStabilized"
		setCondition(mpa, mpa_types.AbleToScale, v1.ConditionTrue, reason, message)
	} else {
		setCondition(mpa, mpa_types.AbleToScale, v1.ConditionTrue, "ReadyForNewScale", "recommended size matches current size")
	}
	desiredReplicas, reason, message := r.convertDesiredReplicasWithBehaviorRate(normalizationArg)
	if desiredReplicas == stabilizedRecommendation {
		setCondition(mpa, mpa_types.ScalingLimited, v1.ConditionFalse, reason, message)
	} else {
		setCondition(mpa, mpa_types.ScalingLimited, v1.ConditionTrue, reason, message)
	}

	return desiredReplicas
}

func (r *recommender) maybeInitScaleDownStabilizationWindow(mpa *mpa_types.MultidimPodAutoscaler) {
	behavior := mpa.Spec.Constraints.Behavior
	if behavior != nil && behavior.ScaleDown != nil && behavior.ScaleDown.StabilizationWindowSeconds == nil {
		stabilizationWindowSeconds := (int32)(r.downscaleStabilisationWindow.Seconds())
		mpa.Spec.Constraints.Behavior.ScaleDown.StabilizationWindowSeconds = &stabilizationWindowSeconds
	}
}

// stabilizeRecommendationWithBehaviors:
// - replaces old recommendation with the newest recommendation,
// - returns {max,min} of recommendations that are not older than constraints.Scale{Up,Down}.DelaySeconds
func (r *recommender) stabilizeRecommendationWithBehaviors(args NormalizationArg) (int32, string, string) {
	now := time.Now()

	foundOldSample := false
	oldSampleIndex := 0

	upRecommendation := args.DesiredReplicas
	upDelaySeconds := *args.ScaleUpBehavior.StabilizationWindowSeconds
	upCutoff := now.Add(-time.Second * time.Duration(upDelaySeconds))

	downRecommendation := args.DesiredReplicas
	downDelaySeconds := *args.ScaleDownBehavior.StabilizationWindowSeconds
	downCutoff := now.Add(-time.Second * time.Duration(downDelaySeconds))

	// Calculate the upper and lower stabilization limits.
	r.recommendationsLock.Lock()
	defer r.recommendationsLock.Unlock()
	for i, rec := range r.recommendations[args.Key] {
		if rec.timestamp.After(upCutoff) {
			upRecommendation = min(rec.recommendation, upRecommendation)
		}
		if rec.timestamp.After(downCutoff) {
			downRecommendation = max(rec.recommendation, downRecommendation)
		}
		if rec.timestamp.Before(upCutoff) && rec.timestamp.Before(downCutoff) {
			foundOldSample = true
			oldSampleIndex = i
		}
	}

	// Bring the recommendation to within the upper and lower limits (stabilize).
	recommendation := args.CurrentReplicas
	if recommendation < upRecommendation {
		recommendation = upRecommendation
	}
	if recommendation > downRecommendation {
		recommendation = downRecommendation
	}

	// Record the unstabilized recommendation.
	if foundOldSample {
		r.recommendations[args.Key][oldSampleIndex] = timestampedRecommendation{args.DesiredReplicas, time.Now()}
	} else {
		r.recommendations[args.Key] = append(r.recommendations[args.Key], timestampedRecommendation{args.DesiredReplicas, time.Now()})
	}

	// Determine a human-friendly message.
	var reason, message string
	if args.DesiredReplicas >= args.CurrentReplicas {
		reason = "ScaleUpStabilized"
		message = "recent recommendations were lower than current one, applying the lowest recent recommendation"
	} else {
		reason = "ScaleDownStabilized"
		message = "recent recommendations were higher than current one, applying the highest recent recommendation"
	}
	return recommendation, reason, message
}

func max(a, b int32) int32 {
	if a >= b {
		return a
	}
	return b
}

func min(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
}

// calculateScaleUpLimitWithScalingRules returns the maximum number of pods that could be added for the given HPAScalingRules
func calculateScaleUpLimitWithScalingRules(currentReplicas int32, scaleUpEvents, scaleDownEvents []timestampedScaleEvent, scalingRules *autoscalingv2.HPAScalingRules) int32 {
	var result int32
	var proposed int32
	var selectPolicyFn func(int32, int32) int32
	if *scalingRules.SelectPolicy == autoscalingv2.DisabledPolicySelect {
		return currentReplicas // Scaling is disabled
	} else if *scalingRules.SelectPolicy == autoscalingv2.MinChangePolicySelect {
		result = math.MaxInt32
		selectPolicyFn = min // For scaling up, the lowest change ('min' policy) produces a minimum value
	} else {
		result = math.MinInt32
		selectPolicyFn = max // Use the default policy otherwise to produce a highest possible change
	}
	for _, policy := range scalingRules.Policies {
		replicasAddedInCurrentPeriod := getReplicasChangePerPeriod(policy.PeriodSeconds, scaleUpEvents)
		replicasDeletedInCurrentPeriod := getReplicasChangePerPeriod(policy.PeriodSeconds, scaleDownEvents)
		periodStartReplicas := currentReplicas - replicasAddedInCurrentPeriod + replicasDeletedInCurrentPeriod
		if policy.Type == autoscalingv2.PodsScalingPolicy {
			proposed = periodStartReplicas + policy.Value
		} else if policy.Type == autoscalingv2.PercentScalingPolicy {
			// the proposal has to be rounded up because the proposed change might not increase the replica count causing the target to never scale up
			proposed = int32(math.Ceil(float64(periodStartReplicas) * (1 + float64(policy.Value)/100)))
		}
		result = selectPolicyFn(result, proposed)
	}
	return result
}

// calculateScaleDownLimitWithBehavior returns the maximum number of pods that could be deleted for the given HPAScalingRules
func calculateScaleDownLimitWithBehaviors(currentReplicas int32, scaleUpEvents, scaleDownEvents []timestampedScaleEvent, scalingRules *autoscalingv2.HPAScalingRules) int32 {
	var result int32
	var proposed int32
	var selectPolicyFn func(int32, int32) int32
	if *scalingRules.SelectPolicy == autoscalingv2.DisabledPolicySelect {
		return currentReplicas // Scaling is disabled
	} else if *scalingRules.SelectPolicy == autoscalingv2.MinChangePolicySelect {
		result = math.MinInt32
		selectPolicyFn = max // For scaling down, the lowest change ('min' policy) produces a maximum value
	} else {
		result = math.MaxInt32
		selectPolicyFn = min // Use the default policy otherwise to produce a highest possible change
	}
	for _, policy := range scalingRules.Policies {
		replicasAddedInCurrentPeriod := getReplicasChangePerPeriod(policy.PeriodSeconds, scaleUpEvents)
		replicasDeletedInCurrentPeriod := getReplicasChangePerPeriod(policy.PeriodSeconds, scaleDownEvents)
		periodStartReplicas := currentReplicas - replicasAddedInCurrentPeriod + replicasDeletedInCurrentPeriod
		if policy.Type == autoscalingv2.PodsScalingPolicy {
			proposed = periodStartReplicas - policy.Value
		} else if policy.Type == autoscalingv2.PercentScalingPolicy {
			proposed = int32(float64(periodStartReplicas) * (1 - float64(policy.Value)/100))
		}
		result = selectPolicyFn(result, proposed)
	}
	return result
}

// getReplicasChangePerPeriod function find all the replica changes per period
func getReplicasChangePerPeriod(periodSeconds int32, scaleEvents []timestampedScaleEvent) int32 {
	period := time.Second * time.Duration(periodSeconds)
	cutoff := time.Now().Add(-period)
	var replicas int32
	for _, rec := range scaleEvents {
		if rec.timestamp.After(cutoff) {
			replicas += rec.replicaChange
		}
	}
	return replicas

}

// storeScaleEvent stores (adds or replaces outdated) scale event.
// outdated events to be replaced were marked as outdated in the `markScaleEventsOutdated` function
func (r *recommender) storeScaleEvent(behavior *autoscalingv2.HorizontalPodAutoscalerBehavior, key model.MpaID, prevReplicas, newReplicas int32) {
	if behavior == nil {
		return  // we should not store any event as they will not be used
	}
	var oldSampleIndex int
	var longestPolicyPeriod int32
	foundOldSample := false
	if newReplicas > prevReplicas {
		longestPolicyPeriod = getLongestPolicyPeriod(behavior.ScaleUp)

		r.scaleUpEventsLock.Lock()
		defer r.scaleUpEventsLock.Unlock()
		markScaleEventsOutdated(r.scaleUpEvents[key], longestPolicyPeriod)
		replicaChange := newReplicas - prevReplicas
		for i, event := range r.scaleUpEvents[key] {
			if event.outdated {
				foundOldSample = true
				oldSampleIndex = i
			}
		}
		newEvent := timestampedScaleEvent{replicaChange, time.Now(), false}
		if foundOldSample {
			r.scaleUpEvents[key][oldSampleIndex] = newEvent
		} else {
			r.scaleUpEvents[key] = append(r.scaleUpEvents[key], newEvent)
		}
	} else {
		longestPolicyPeriod = getLongestPolicyPeriod(behavior.ScaleDown)

		r.scaleDownEventsLock.Lock()
		defer r.scaleDownEventsLock.Unlock()
		markScaleEventsOutdated(r.scaleDownEvents[key], longestPolicyPeriod)
		replicaChange := prevReplicas - newReplicas
		for i, event := range r.scaleDownEvents[key] {
			if event.outdated {
				foundOldSample = true
				oldSampleIndex = i
			}
		}
		newEvent := timestampedScaleEvent{replicaChange, time.Now(), false}
		if foundOldSample {
			r.scaleDownEvents[key][oldSampleIndex] = newEvent
		} else {
			r.scaleDownEvents[key] = append(r.scaleDownEvents[key], newEvent)
		}
	}
}

func getLongestPolicyPeriod(scalingRules *autoscalingv2.HPAScalingRules) int32 {
	var longestPolicyPeriod int32
	for _, policy := range scalingRules.Policies {
		if policy.PeriodSeconds > longestPolicyPeriod {
			longestPolicyPeriod = policy.PeriodSeconds
		}
	}
	return longestPolicyPeriod
}

// markScaleEventsOutdated set 'outdated=true' flag for all scale events that are not used by any MPA object
func markScaleEventsOutdated(scaleEvents []timestampedScaleEvent, longestPolicyPeriod int32) {
	period := time.Second * time.Duration(longestPolicyPeriod)
	cutoff := time.Now().Add(-period)
	for i, event := range scaleEvents {
		if event.timestamp.Before(cutoff) {
			// outdated scale event are marked for later reuse
			scaleEvents[i].outdated = true
		}
	}
}
