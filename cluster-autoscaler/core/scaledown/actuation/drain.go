/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	acontext "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// DefaultEvictionRetryTime is the time after CA retries failed pod eviction.
	DefaultEvictionRetryTime = 10 * time.Second
	// DefaultPodEvictionHeadroom is the extra time we wait to catch situations when the pod is ignoring SIGTERM and
	// is killed with SIGKILL after MaxGracefulTerminationTime
	DefaultPodEvictionHeadroom = 30 * time.Second
	// DefaultDsEvictionEmptyNodeTimeout is the time to evict all DaemonSet pods on empty node
	DefaultDsEvictionEmptyNodeTimeout = 10 * time.Second
	// DefaultDsEvictionRetryTime is a time between retries to create eviction that uses for DaemonSet eviction for empty nodes
	DefaultDsEvictionRetryTime = 3 * time.Second
)

type evictionRegister interface {
	RegisterEviction(*apiv1.Pod)
}

// Evictor can be used to evict pods from nodes.
type Evictor struct {
	EvictionRetryTime          time.Duration
	DsEvictionRetryTime        time.Duration
	DsEvictionEmptyNodeTimeout time.Duration
	PodEvictionHeadroom        time.Duration
	evictionRegister           evictionRegister
	deleteOptions              simulator.NodeDeleteOptions
}

// NewDefaultEvictor returns an instance of Evictor using the default parameters.
func NewDefaultEvictor(deleteOptions simulator.NodeDeleteOptions, evictionRegister evictionRegister) Evictor {
	return Evictor{
		EvictionRetryTime:          DefaultEvictionRetryTime,
		DsEvictionRetryTime:        DefaultDsEvictionRetryTime,
		DsEvictionEmptyNodeTimeout: DefaultDsEvictionEmptyNodeTimeout,
		PodEvictionHeadroom:        DefaultPodEvictionHeadroom,
		evictionRegister:           evictionRegister,
		deleteOptions:              deleteOptions,
	}
}

// DrainNode works like DrainNodeWithPods, but lists of pods to evict don't have to be provided. All non-mirror, non-DS pods on the
// node are evicted. Mirror pods are not evicted. DaemonSet pods are evicted if DaemonSetEvictionForOccupiedNodes is enabled, or
// if they have the EnableDsEvictionKey annotation.
func (e Evictor) DrainNode(ctx *acontext.AutoscalingContext, nodeInfo *framework.NodeInfo) (map[string]status.PodEvictionResult, error) {
	dsPodsToEvict, nonDsPodsToEvict := podsToEvict(ctx, nodeInfo)
	return e.DrainNodeWithPods(ctx, nodeInfo.Node(), nonDsPodsToEvict, dsPodsToEvict)
}

// DrainNodeWithPods performs drain logic on the node. Marks the node as unschedulable and later removes all pods, giving
// them up to MaxGracefulTerminationTime to finish. The list of pods to evict has to be provided.
func (e Evictor) DrainNodeWithPods(ctx *acontext.AutoscalingContext, node *apiv1.Node, pods []*apiv1.Pod, daemonSetPods []*apiv1.Pod) (map[string]status.PodEvictionResult, error) {
	evictionResults := make(map[string]status.PodEvictionResult)
	retryUntil := time.Now().Add(ctx.MaxPodEvictionTime)
	confirmations := make(chan status.PodEvictionResult, len(pods))
	daemonSetConfirmations := make(chan status.PodEvictionResult, len(daemonSetPods))
	for _, pod := range pods {
		evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: nil}
		go func(podToEvict *apiv1.Pod) {
			confirmations <- evictPod(ctx, podToEvict, false, retryUntil, e.EvictionRetryTime, e.evictionRegister)
		}(pod)
	}

	// Perform eviction of daemonset. We don't want to raise an error if daemonsetPod wasn't evict properly
	for _, daemonSetPod := range daemonSetPods {
		go func(podToEvict *apiv1.Pod) {
			daemonSetConfirmations <- evictPod(ctx, podToEvict, true, retryUntil, e.EvictionRetryTime, e.evictionRegister)
		}(daemonSetPod)

	}

	podsEvictionCounter := 0
	for i := 0; i < len(pods)+len(daemonSetPods); i++ {
		select {
		case evictionResult := <-confirmations:
			podsEvictionCounter++
			evictionResults[evictionResult.Pod.Name] = evictionResult
			if evictionResult.WasEvictionSuccessful() {
				metrics.RegisterEvictions(1)
			}
		case <-daemonSetConfirmations:
		case <-time.After(retryUntil.Sub(time.Now()) + 5*time.Second):
			if podsEvictionCounter < len(pods) {
				// All pods initially had results with TimedOut set to true, so the ones that didn't receive an actual result are correctly marked as timed out.
				return evictionResults, errors.NewAutoscalerError(errors.ApiCallError, "Failed to drain node %s/%s: timeout when waiting for creating evictions", node.Namespace, node.Name)
			}
			klog.Infof("Timeout when waiting for creating daemonSetPods eviction")
		}
	}

	evictionErrs := make([]error, 0)
	for _, result := range evictionResults {
		if !result.WasEvictionSuccessful() {
			evictionErrs = append(evictionErrs, result.Err)
		}
	}
	if len(evictionErrs) != 0 {
		return evictionResults, errors.NewAutoscalerError(errors.ApiCallError, "Failed to drain node %s/%s, due to following errors: %v", node.Namespace, node.Name, evictionErrs)
	}

	// Evictions created successfully, wait maxGracefulTerminationSec + podEvictionHeadroom to see if pods really disappeared.
	var allGone bool
	for start := time.Now(); time.Now().Sub(start) < time.Duration(ctx.MaxGracefulTerminationSec)*time.Second+e.PodEvictionHeadroom; time.Sleep(5 * time.Second) {
		allGone = true
		for _, pod := range pods {
			podreturned, err := ctx.ClientSet.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if err == nil && (podreturned == nil || podreturned.Spec.NodeName == node.Name) {
				klog.V(1).Infof("Not deleted yet %s/%s", pod.Namespace, pod.Name)
				allGone = false
				break
			}
			if err != nil && !kube_errors.IsNotFound(err) {
				klog.Errorf("Failed to check pod %s/%s: %v", pod.Namespace, pod.Name, err)
				allGone = false
				break
			}
		}
		if allGone {
			klog.V(1).Infof("All pods removed from %s", node.Name)
			// Let the deferred function know there is no need for cleanup
			return evictionResults, nil
		}
	}

	for _, pod := range pods {
		podReturned, err := ctx.ClientSet.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
		if err == nil && (podReturned == nil || podReturned.Spec.NodeName == node.Name) {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: nil}
		} else if err != nil && !kube_errors.IsNotFound(err) {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: err}
		} else {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: false, Err: nil}
		}
	}

	return evictionResults, errors.NewAutoscalerError(errors.TransientError, "Failed to drain node %s/%s: pods remaining after timeout", node.Namespace, node.Name)
}

// EvictDaemonSetPods creates eviction objects for all DaemonSet pods on the node.
func (e Evictor) EvictDaemonSetPods(ctx *acontext.AutoscalingContext, nodeInfo *framework.NodeInfo, timeNow time.Time) error {
	nodeToDelete := nodeInfo.Node()
	_, daemonSetPods, _, err := simulator.GetPodsToMove(nodeInfo, e.deleteOptions, nil, []*policyv1.PodDisruptionBudget{}, timeNow)
	if err != nil {
		return fmt.Errorf("failed to get DaemonSet pods for %s (error: %v)", nodeToDelete.Name, err)
	}

	daemonSetPods = daemonset.PodsToEvict(daemonSetPods, ctx.DaemonSetEvictionForEmptyNodes)

	dsEviction := make(chan status.PodEvictionResult, len(daemonSetPods))

	// Perform eviction of DaemonSet pods
	for _, daemonSetPod := range daemonSetPods {
		go func(podToEvict *apiv1.Pod) {
			dsEviction <- evictPod(ctx, podToEvict, true, timeNow.Add(e.DsEvictionEmptyNodeTimeout), e.DsEvictionRetryTime, e.evictionRegister)
		}(daemonSetPod)
	}
	// Wait for creating eviction of DaemonSet pods
	var failedPodErrors []string
	for range daemonSetPods {
		select {
		case status := <-dsEviction:
			if status.Err != nil {
				failedPodErrors = append(failedPodErrors, status.Err.Error())
			}
		// adding waitBetweenRetries in order to have a bigger time interval than evictPod()
		case <-time.After(e.DsEvictionEmptyNodeTimeout):
			return fmt.Errorf("failed to create DaemonSet eviction for %v seconds on the %s", e.DsEvictionEmptyNodeTimeout, nodeToDelete.Name)
		}
	}
	if len(failedPodErrors) > 0 {

		return fmt.Errorf("following DaemonSet pod failed to evict on the %s:\n%s", nodeToDelete.Name, fmt.Errorf(strings.Join(failedPodErrors, "\n")))
	}
	return nil
}

func evictPod(ctx *acontext.AutoscalingContext, podToEvict *apiv1.Pod, isDaemonSetPod bool, retryUntil time.Time, waitBetweenRetries time.Duration, evictionRegister evictionRegister) status.PodEvictionResult {
	ctx.Recorder.Eventf(podToEvict, apiv1.EventTypeNormal, "ScaleDown", "deleting pod for node scale down")

	maxTermination := int64(apiv1.DefaultTerminationGracePeriodSeconds)
	if podToEvict.Spec.TerminationGracePeriodSeconds != nil {
		if *podToEvict.Spec.TerminationGracePeriodSeconds < int64(ctx.MaxGracefulTerminationSec) {
			maxTermination = *podToEvict.Spec.TerminationGracePeriodSeconds
		} else {
			maxTermination = int64(ctx.MaxGracefulTerminationSec)
		}
	}

	var lastError error
	for first := true; first || time.Now().Before(retryUntil); time.Sleep(waitBetweenRetries) {
		first = false
		eviction := &policyv1beta1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: podToEvict.Namespace,
				Name:      podToEvict.Name,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: &maxTermination,
			},
		}
		lastError = ctx.ClientSet.CoreV1().Pods(podToEvict.Namespace).Evict(context.TODO(), eviction)
		if lastError == nil || kube_errors.IsNotFound(lastError) {
			if evictionRegister != nil {
				evictionRegister.RegisterEviction(podToEvict)
			}
			return status.PodEvictionResult{Pod: podToEvict, TimedOut: false, Err: nil}
		}
	}
	if !isDaemonSetPod {
		klog.Errorf("Failed to evict pod %s, error: %v", podToEvict.Name, lastError)
		ctx.Recorder.Eventf(podToEvict, apiv1.EventTypeWarning, "ScaleDownFailed", "failed to delete pod for ScaleDown")
	}
	return status.PodEvictionResult{Pod: podToEvict, TimedOut: true, Err: fmt.Errorf("failed to evict pod %s/%s within allowed timeout (last error: %v)", podToEvict.Namespace, podToEvict.Name, lastError)}
}

func podsToEvict(ctx *acontext.AutoscalingContext, nodeInfo *framework.NodeInfo) (dsPods, nonDsPods []*apiv1.Pod) {
	for _, podInfo := range nodeInfo.Pods {
		if pod_util.IsMirrorPod(podInfo.Pod) {
			continue
		} else if pod_util.IsDaemonSetPod(podInfo.Pod) {
			dsPods = append(dsPods, podInfo.Pod)
		} else {
			nonDsPods = append(nonDsPods, podInfo.Pod)
		}
	}
	dsPodsToEvict := daemonset.PodsToEvict(dsPods, ctx.DaemonSetEvictionForOccupiedNodes)
	return dsPodsToEvict, nonDsPods
}
