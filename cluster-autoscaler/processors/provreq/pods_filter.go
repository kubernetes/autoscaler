/*
Copyright 2023 The Kubernetes Authors.

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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	provreqpods "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
)

const maxProvReqEvent = 50

// EventManager is an interface for handling events for provisioning request.
type EventManager interface {
	LogIgnoredInScaleUpEvent(context *context.AutoscalingContext, now time.Time, pod *apiv1.Pod, prName string)
	Reset()
}

type defaultEventManager struct {
	loggedEvents int
	limit        int
}

// NewDefautlEventManager return basic event manager.
func NewDefautlEventManager() *defaultEventManager {
	return &defaultEventManager{limit: maxProvReqEvent}
}

// LogIgnoredInScaleUpEvent adds event about ignored scale up for unscheduled pod, that consumes Provisioning Request.
func (e *defaultEventManager) LogIgnoredInScaleUpEvent(context *context.AutoscalingContext, now time.Time, pod *apiv1.Pod, prName string) {
	message := fmt.Sprintf("Unschedulable pod didn't trigger scale-up, because it's consuming ProvisioningRequest %s/%s", pod.Namespace, prName)
	if e.loggedEvents < e.limit {
		context.Recorder.Event(pod, apiv1.EventTypeNormal, "", message)
		e.loggedEvents++
	}
}

// Reset resets event manager internal structure. It will be called once before handling all pods.
func (e *defaultEventManager) Reset() {
	e.loggedEvents = 0
}

// ProvisioningRequestPodsFilter filter out pods that consumes Provisioning Request
type ProvisioningRequestPodsFilter struct {
	eventManager EventManager
}

// Process filters out all pods that are consuming a Provisioning Request from unschedulable pods list.
func (p *ProvisioningRequestPodsFilter) Process(
	context *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod,
) ([]*apiv1.Pod, error) {
	now := time.Now()
	p.eventManager.Reset()
	loggingQuota := klogx.PodsLoggingQuota()
	result := make([]*apiv1.Pod, 0, len(unschedulablePods))
	for _, pod := range unschedulablePods {
		prName, found := provisioningRequestName(pod)
		if !found {
			result = append(result, pod)
			continue
		}
		klogx.V(1).UpTo(loggingQuota).Infof("Ignoring unschedulable pod %s/%s as it consumes ProvisioningRequest: %s/%s", pod.Namespace, pod.Name, pod.Namespace, prName)
		p.eventManager.LogIgnoredInScaleUpEvent(context, now, pod, prName)
	}
	klogx.V(1).Over(loggingQuota).Infof("There are also %v other pods which were ignored", -loggingQuota.Left())
	return result, nil
}

// CleanUp cleans up the processor's internal structures.
func (p *ProvisioningRequestPodsFilter) CleanUp() {}

// NewProvisioningRequestPodsFilter creates a ProvisioningRequest filter processor.
func NewProvisioningRequestPodsFilter(e EventManager) pods.PodListProcessor {
	return &ProvisioningRequestPodsFilter{e}
}

func provisioningRequestName(pod *corev1.Pod) (string, bool) {
	if pod == nil || pod.Annotations == nil {
		return "", false
	}
	provReqName, found := pod.Annotations[v1.ProvisioningRequestPodAnnotationKey]
	if !found {
		provReqName, found = pod.Annotations[provreqpods.DeprecatedProvisioningRequestPodAnnotationKey]
	}
	return provReqName, found
}
