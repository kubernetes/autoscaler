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

package podlistprocessor

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	klog "k8s.io/klog/v2"
)

type filterOutDaemonSetPodListProcessor struct {
}

// NewFilterOutDaemonSetPodListProcessor creates a PodListProcessor filtering out daemon set pods
func NewFilterOutDaemonSetPodListProcessor() *filterOutDaemonSetPodListProcessor {
	return &filterOutDaemonSetPodListProcessor{}
}

// Process filters out pods which are daemon set pods.
func (p *filterOutDaemonSetPodListProcessor) Process(context *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	// Scale-up cannot help unschedulable Daemon Set pods, as those require a specific node
	// for scheduling. To improve that we are filtering them here, as the CA won't be
	// able to help them so there is no point to in passing them to scale-up logic.

	klog.V(4).Infof("Filtering out daemon set pods")

	var nonDaemonSetPods []*apiv1.Pod
	for _, pod := range unschedulablePods {
		if !podutils.IsDaemonSetPod(pod) {
			nonDaemonSetPods = append(nonDaemonSetPods, pod)
		}
	}

	klog.V(4).Infof("Filtered out %v daemon set pods, %v unschedulable pods left", len(unschedulablePods)-len(nonDaemonSetPods), len(nonDaemonSetPods))
	return nonDaemonSetPods, nil
}

func (p *filterOutDaemonSetPodListProcessor) CleanUp() {
}
