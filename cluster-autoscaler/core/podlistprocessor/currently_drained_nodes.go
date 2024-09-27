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

package podlistprocessor

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/klog/v2"
)

type currentlyDrainedNodesPodListProcessor struct {
}

// NewCurrentlyDrainedNodesPodListProcessor returns a new processor adding pods
// from currently drained nodes to the unschedulable pods.
func NewCurrentlyDrainedNodesPodListProcessor() *currentlyDrainedNodesPodListProcessor {
	return &currentlyDrainedNodesPodListProcessor{}
}

// Process adds recreatable pods from currently drained nodes
func (p *currentlyDrainedNodesPodListProcessor) Process(context *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	recreatablePods := pod_util.FilterRecreatablePods(currentlyDrainedPods(context))
	return append(unschedulablePods, pod_util.ClearPodNodeNames(recreatablePods)...), nil
}

func (p *currentlyDrainedNodesPodListProcessor) CleanUp() {
}

func currentlyDrainedPods(context *context.AutoscalingContext) []*apiv1.Pod {
	var pods []*apiv1.Pod
	_, nodeNames := context.ScaleDownActuator.CheckStatus().DeletionsInProgress()
	for _, nodeName := range nodeNames {
		nodeInfo, err := context.ClusterSnapshot.GetNodeInfo(nodeName)
		if err != nil {
			klog.Warningf("Couldn't get node %v info, assuming the node got deleted already: %v", nodeName, err)
			continue
		}
		for _, podInfo := range nodeInfo.Pods {
			// Filter out pods that has deletion timestamp set
			if podInfo.Pod.DeletionTimestamp != nil {
				klog.Infof("Pod %v has deletion timestamp set, skipping injection to unschedulable pods list", podInfo.Pod.Name)
				continue
			}
			pods = append(pods, podInfo.Pod)
		}
	}
	return pods
}
