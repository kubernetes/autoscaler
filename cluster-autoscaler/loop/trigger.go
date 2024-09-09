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

package loop

import (
	"context"
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	podv1 "k8s.io/kubernetes/pkg/api/v1/pod"
)

const maxPodChangeAge = 10 * time.Second

var (
	podsResource             = "pods"
	unschedulablePodSelector = fields.ParseSelectorOrDie("spec.nodeName==" + "" + ",status.phase!=" +
		string(apiv1.PodSucceeded) + ",status.phase!=" + string(apiv1.PodFailed))
)

// scalingTimesGetter exposes recent autoscaler activity
type scalingTimesGetter interface {
	LastScaleUpTime() time.Time
	LastScaleDownDeleteTime() time.Time
}

// provisioningRequestProcessingTimesGetter exposes recent provisioning request processing activity regardless of wether the
// ProvisioningRequest was marked as accepted or failed. This is because a ProvisioningRequest being processed indicates that
// there are other ProvisioningRequests that require processing regardless of the outcome of the current one. Thus, the next iteration
// should be started immediately.
type provisioningRequestProcessingTimesGetter interface {
	LastProvisioningRequestProcessTime() time.Time
}

// LoopTrigger object implements criteria used to start new autoscaling iteration
type LoopTrigger struct {
	initialized                          bool
	podObserver                          *UnschedulablePodObserver
	scanInterval                         time.Duration
	scalingTimesGetter                   scalingTimesGetter
	provisioningRequestProcessTimeGetter provisioningRequestProcessingTimesGetter
}

// NewLoopTrigger creates a LoopTrigger object
func NewLoopTrigger(scalingTimesGetter scalingTimesGetter, provisioningRequestProcessTimeGetter provisioningRequestProcessingTimesGetter, scanInterval time.Duration) *LoopTrigger {
	return &LoopTrigger{
		podObserver:                          nil,
		scanInterval:                         scanInterval,
		scalingTimesGetter:                   scalingTimesGetter,
		provisioningRequestProcessTimeGetter: provisioningRequestProcessTimeGetter,
	}
}

// Initialize initializes the LoopTrigger object by providing a pointer to the UnschedulablePodObserver
func (t *LoopTrigger) Initialize(podObserver *UnschedulablePodObserver) {
	t.podObserver = podObserver
	t.initialized = true
}

// Wait waits for the next autoscaling iteration
func (t *LoopTrigger) Wait(lastRun time.Time) errors.AutoscalerError {
	sleepStart := time.Now()
	defer metrics.UpdateDurationFromStart(metrics.LoopWait, sleepStart)

	// To improve scale-up throughput, Cluster Autoscaler starts new iteration
	// immediately if the previous one was productive.
	if !t.initialized {
		return errors.ToAutoscalerError(errors.InternalError, fmt.Errorf("LoopTrigger not initialized"))
	}

	if !t.scalingTimesGetter.LastScaleUpTime().Before(lastRun) ||
		!t.scalingTimesGetter.LastScaleDownDeleteTime().Before(lastRun) {
		t.triggerNextIteration("Autoscaler loop triggered immediately after a productive iteration")
		return nil
	}

	if t.provisioningRequestWasProcessed(lastRun) {
		t.triggerNextIteration("Autoscaler loop triggered immediately after a provisioning request was processed")
		return nil
	}

	// Unschedulable pod triggers autoscaling immediately.
	select {
	case <-time.After(t.scanInterval):
		klog.Infof("Autoscaler loop triggered by a %v timer", t.scanInterval)
	case <-t.podObserver.unschedulablePodChan:
		klog.Info("Autoscaler loop triggered by unschedulable pod appearing")
	}
	return nil
}

// UnschedulablePodObserver triggers a new loop if there are new unschedulable pods
type UnschedulablePodObserver struct {
	unschedulablePodChan <-chan any
}

// StartPodObserver creates an informer and starts a goroutine watching for newly added
// or updated pods. Each time a new unschedulable pod appears or a change causes a pod to become
// unschedulable, a message is sent to the UnschedulablePodObserver's channel.
func StartPodObserver(ctx context.Context, kubeClient kube_client.Interface) *UnschedulablePodObserver {
	podChan := make(chan any, 1)
	listWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), podsResource, apiv1.NamespaceAll, unschedulablePodSelector)
	informer := cache.NewSharedInformer(listWatch, &apiv1.Pod{}, time.Hour)
	addEventHandlerFunc := func(obj any) {
		if isRecentUnschedulablePod(obj) {
			klog.V(5).Infof(" filterPodChanUntilClose emits signal")
			select {
			case podChan <- struct{}{}:
			default:
			}
		}
	}
	updateEventHandlerFunc := func(old any, newOjb any) { addEventHandlerFunc(newOjb) }
	_, _ = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addEventHandlerFunc,
		UpdateFunc: updateEventHandlerFunc,
	})
	go informer.Run(ctx.Done())
	return &UnschedulablePodObserver{
		unschedulablePodChan: podChan,
	}
}

// triggerNextIteration logs a message if the next iteration was not triggered by unschedulable pods appearing, else it logs a message that the next iteration was triggered by unschedulable pods appearing
func (t *LoopTrigger) triggerNextIteration(message string) {
	select {
	case <-t.podObserver.unschedulablePodChan:
		klog.Info("Autoscaler loop triggered by unschedulable pod appearing")
	default:
		klog.Infof(message)
	}
}

func (t *LoopTrigger) provisioningRequestWasProcessed(lastRun time.Time) bool {
	return t.provisioningRequestProcessTimeGetter != nil && !t.provisioningRequestProcessTimeGetter.LastProvisioningRequestProcessTime().Before(lastRun)
}

// isRecentUnschedulablePod checks if the object is an unschedulable pod observed recently.
func isRecentUnschedulablePod(obj any) bool {
	pod, ok := obj.(*apiv1.Pod)
	if !ok {
		return false
	}
	if pod.Status.Phase == apiv1.PodSucceeded || pod.Status.Phase == apiv1.PodFailed {
		return false
	}
	if pod.Spec.NodeName != "" {
		return false
	}
	_, scheduledCondition := podv1.GetPodCondition(&pod.Status, apiv1.PodScheduled)
	if scheduledCondition == nil {
		return false
	}
	if scheduledCondition.Status != apiv1.ConditionFalse || scheduledCondition.Reason != "Unschedulable" {
		return false
	}
	if scheduledCondition.LastTransitionTime.Time.Add(maxPodChangeAge).Before(time.Now()) {
		return false
	}
	return true
}
