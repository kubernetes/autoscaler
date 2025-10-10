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
	podinjectionbackoff "k8s.io/autoscaler/cluster-autoscaler/processors/podinjection/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/fake"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
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
		return unschedulablePods, fmt.Errorf("failed to list nodeInfos from cluster snapshot: %v", err)
	}
	scheduledPods := podsFromNodeInfos(nodeInfos)

	allPods, err := ctx.AllPodLister().List()
	if err != nil {
		return unschedulablePods, fmt.Errorf("failed to list all pods from all pod lister: %v", err)
	}
	schedulingGatedPods := kube_util.SchedulingGatedPods(allPods)

	groupedPods := groupPods(append(append(scheduledPods, unschedulablePods...), schedulingGatedPods...), controllers)

	var podsToInject []*apiv1.Pod

	for _, groupedPod := range groupedPods {
		var fakePodCount = groupedPod.fakePodCount()
		fakePods := makeFakePods(groupedPod.ownerUid, groupedPod.sample, fakePodCount)
		podsToInject = append(podsToInject, fakePods...)
	}

	unschedulablePodsAfterProcessing := append(unschedulablePods, podsToInject...)

	return unschedulablePodsAfterProcessing, nil
}

// CleanUp is called at CA termination
func (p *PodInjectionPodListProcessor) CleanUp() {
}

// makeFakePods creates podCount number of copies of the sample pod
// makeFakePods also adds annotation to the pod to be marked as "fake"
func makeFakePods(ownerUid types.UID, samplePod *apiv1.Pod, podCount int) []*apiv1.Pod {
	var fakePods []*apiv1.Pod
	for i := 1; i <= podCount; i++ {
		newPod := fake.WithFakePodAnnotation(samplePod.DeepCopy())
		newPod.Name = fmt.Sprintf("%s-copy-%d", samplePod.Name, i)
		newPod.UID = types.UID(fmt.Sprintf("%s-%d", string(ownerUid), i))
		newPod.Spec.NodeName = ""
		fakePods = append(fakePods, newPod)
	}
	return fakePods
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
		for _, podInfo := range nodeInfo.Pods() {
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
