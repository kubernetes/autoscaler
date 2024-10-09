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

package podinjection

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/dynamicresources"
	podinjectionbackoff "k8s.io/autoscaler/cluster-autoscaler/processors/podinjection/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

const (
	// FakePodAnnotationKey the key for pod type
	FakePodAnnotationKey = "podtype"
	// FakePodAnnotationValue the value for a fake pod
	FakePodAnnotationValue = "fakepod"
)

// PodInjectionPodListProcessor is a PodListProcessor used to inject fake pods to consider replica count in the respective controllers for the scale-up.
// For each controller, #fake pods injected = #replicas specified the controller - #scheduled pods - #finished pods - #unschedulable pods
type PodInjectionPodListProcessor struct {
	fakePodControllerBackoffRegistry *podinjectionbackoff.ControllerRegistry
}

// controller is a struct that can be used to abstract different pod controllers
type controller struct {
	uid             types.UID
	desiredReplicas int
}

// NewPodInjectionPodListProcessor return an instance of PodInjectionPodListProcessor
func NewPodInjectionPodListProcessor(fakePodRegistry *podinjectionbackoff.ControllerRegistry) *PodInjectionPodListProcessor {
	return &PodInjectionPodListProcessor{
		fakePodControllerBackoffRegistry: fakePodRegistry,
	}
}

// Process updates unschedulablePods by injecting fake pods to match target replica count
func (p *PodInjectionPodListProcessor) Process(ctx *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {

	controllers := listControllers(ctx)
	controllers = p.skipBackedoffControllers(controllers)

	nodeInfos, err := ctx.ClusterSnapshot.ListNodeInfos()
	if err != nil {
		klog.Errorf("Failed to list nodeInfos from cluster snapshot: %v", err)
		return unschedulablePods, fmt.Errorf("failed to list nodeInfos from cluster snapshot: %v", err)
	}
	scheduledPods := podsFromNodeInfos(nodeInfos)

	groupedPods := groupPods(append(scheduledPods, unschedulablePods...), controllers, ctx.ClusterSnapshot, ctx.EnableDynamicResources)
	var podsToInject []*apiv1.Pod

	for _, groupedPod := range groupedPods {
		var fakePodCount = groupedPod.fakePodCount()
		fakePods, err := makeFakePods(groupedPod.ownerUid, groupedPod.sample, fakePodCount)
		if err != nil {
			return nil, err
		}
		for _, podInfo := range fakePods {
			podsToInject = append(podsToInject, podInfo.Pod)

			if ctx.EnableDynamicResources && dynamicresources.PodNeedsResourceClaims(podInfo.Pod) {
				// If the pod we're duplicating owns any ResourceClaims, we need to duplicate and inject them into the cluster snapshot in order
				// for the pods to be able to schedule.
				err := ctx.ClusterSnapshot.AddResourceClaims(podInfo.NeededResourceClaims)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	unschedulablePodsAfterProcessing := append(unschedulablePods, podsToInject...)

	return unschedulablePodsAfterProcessing, nil
}

// CleanUp is called at CA termination
func (p *PodInjectionPodListProcessor) CleanUp() {
}

// makeFakePods creates podCount number of copies of the sample pod
// makeFakePods also adds annotation to the pod to be marked as "fake"
func makeFakePods(ownerUid types.UID, samplePod *framework.PodInfo, podCount int) ([]*framework.PodInfo, error) {
	var fakePods []*framework.PodInfo
	for i := 1; i <= podCount; i++ {
		newPod := withFakePodAnnotation(samplePod.DeepCopy())
		nameSuffix := fmt.Sprintf("copy-%d", i)
		newPod.Name = fmt.Sprintf("%s-%s", samplePod.Name, nameSuffix)
		newPod.UID = types.UID(fmt.Sprintf("%s-%d", string(ownerUid), i))

		// Filter to only the claims owned by the duplicated pod, clear allocations to be sure, sanitize. These will need to be added to the cluster snapshot
		// in order for the fake pod to be able to schedule.
		newResourceClaims, err := dynamicresources.SanitizePodResourceClaims(newPod, samplePod.Pod, samplePod.NeededResourceClaims, true, nameSuffix, "", "", nil)
		if err != nil {
			return nil, err
		}

		fakePods = append(fakePods, &framework.PodInfo{Pod: newPod, NeededResourceClaims: newResourceClaims})
	}
	return fakePods, nil
}

// withFakePodAnnotation adds annotation of key `FakePodAnnotationKey` with value `FakePodAnnotationValue` to passed pod.
// withFakePodAnnotation also creates a new annotations map if original pod.Annotations is nil
func withFakePodAnnotation(pod *apiv1.Pod) *apiv1.Pod {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string, 1)
	}
	pod.Annotations[FakePodAnnotationKey] = FakePodAnnotationValue
	return pod
}

// fakePodCount calculate the fake pod count that should be injected from this podGroup
func (p *podGroup) fakePodCount() int {
	// Controllers with no unschedulable pods are ignored
	if p.podCount == 0 || p.sample == nil {
		return 0
	}
	fakePodCount := p.desiredReplicas - p.podCount
	if fakePodCount <= 0 {
		return 0
	}
	return fakePodCount
}

// podsFromNodeInfos return all the pods in the nodeInfos
func podsFromNodeInfos(nodeInfos []*framework.NodeInfo) []*apiv1.Pod {
	var pods []*apiv1.Pod
	for _, nodeInfo := range nodeInfos {
		for _, podInfo := range nodeInfo.Pods {
			pods = append(pods, podInfo.Pod)
		}
	}
	return pods
}

// listControllers returns the list of controllers that can be used to inject fake pods
func listControllers(ctx *context.AutoscalingContext) []controller {
	var controllers []controller
	controllers = append(controllers, createReplicaSetControllers(ctx)...)
	controllers = append(controllers, createJobControllers(ctx)...)
	controllers = append(controllers, createStatefulSetControllers(ctx)...)
	return controllers
}

// IsFake returns true if the a pod is marked as fake and false otherwise
func IsFake(pod *apiv1.Pod) bool {
	if pod.Annotations == nil {
		return false
	}
	return pod.Annotations[FakePodAnnotationKey] == FakePodAnnotationValue
}

func (p *PodInjectionPodListProcessor) skipBackedoffControllers(controllers []controller) []controller {
	var filteredControllers []controller
	backoffRegistry := p.fakePodControllerBackoffRegistry
	now := time.Now()
	for _, controller := range controllers {
		if backoffUntil := backoffRegistry.BackOffUntil(controller.uid, now); backoffUntil.After(now) {
			klog.Warningf("Skipping generating fake pods for controller in backoff until (%s): %v", backoffUntil.Format(time.TimeOnly), controller.uid)
			continue
		}
		filteredControllers = append(filteredControllers, controller)
	}
	return filteredControllers
}
