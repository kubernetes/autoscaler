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

package estimator

import (
	"sort"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// podScoreInfo contains Pod and score that corresponds to how important it is to handle the pod first.
type podScoreInfo struct {
	score float64
	pod   *apiv1.Pod
}

// DecreasingPodOrderer is the default implementation of the EstimationPodOrderer
// It implements sorting pods by pod score in decreasing order
type DecreasingPodOrderer struct {
}

// NewDecreasingPodOrderer returns the object of DecreasingPodOrderer
func NewDecreasingPodOrderer() *DecreasingPodOrderer {
	return &DecreasingPodOrderer{}
}

// Order is the processing func that sorts the pods based on the size of the pod
func (d *DecreasingPodOrderer) Order(pods []*apiv1.Pod, nodeTemplate *framework.NodeInfo, _ cloudprovider.NodeGroup) []*apiv1.Pod {
	podInfos := make([]*podScoreInfo, 0, len(pods))
	for _, pod := range pods {
		podInfos = append(podInfos, d.calculatePodScore(pod, nodeTemplate))
	}
	sort.Slice(podInfos, func(i, j int) bool { return podInfos[i].score > podInfos[j].score })
	podList := make([]*apiv1.Pod, 0, len(pods))
	for _, podInfo := range podInfos {
		podList = append(podList, podInfo.pod)
	}

	return podList
}

// calculatePodScore score for  pod and returns podScoreInfo structure.
// Score is defined as cpu_sum/node_capacity + mem_sum/node_capacity.
// Pods that have bigger requirements should be processed first, thus have higher scores.
func (d *DecreasingPodOrderer) calculatePodScore(pod *apiv1.Pod, nodeTemplate *framework.NodeInfo) *podScoreInfo {

	cpuSum := resource.Quantity{}
	memorySum := resource.Quantity{}

	for _, container := range pod.Spec.Containers {
		if request, ok := container.Resources.Requests[apiv1.ResourceCPU]; ok {
			cpuSum.Add(request)
		}
		if request, ok := container.Resources.Requests[apiv1.ResourceMemory]; ok {
			memorySum.Add(request)
		}
	}
	score := float64(0)
	if cpuAllocatable, ok := nodeTemplate.Node().Status.Allocatable[apiv1.ResourceCPU]; ok && cpuAllocatable.MilliValue() > 0 {
		score += float64(cpuSum.MilliValue()) / float64(cpuAllocatable.MilliValue())
	}
	if memAllocatable, ok := nodeTemplate.Node().Status.Allocatable[apiv1.ResourceMemory]; ok && memAllocatable.Value() > 0 {
		score += float64(memorySum.Value()) / float64(memAllocatable.Value())
	}

	return &podScoreInfo{
		score: score,
		pod:   pod,
	}
}
