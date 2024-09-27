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

package podsharding

import (
	"k8s.io/api/core/v1"
	pr_pods "k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
)

// NewOssPodSharder returns a PodSharder that shards pods based on the podShardingLabels specified.
func NewOssPodSharder(podShardingLabels map[string]string, provisioningRequestsEnabled bool) PodSharder {
	computeFunctions := []FeatureShardComputeFunction{
		{
			"label",
			computeLabelNameShard(podShardingLabels),
		},
	}

	if provisioningRequestsEnabled {
		computeFunctions = append(computeFunctions, FeatureShardComputeFunction{
			"provisioning",
			provisioningRequestShard,
		})
	}

	return NewCompositePodSharder(computeFunctions)
}

// computeLabelNameShard returns a function with nodeGroupDescriptor labels updated to contain the podShardingLabels if they are present in the pod.
func computeLabelNameShard(podShardingLabels map[string]string) func(*v1.Pod, *NodeGroupDescriptor) {
	return func(pod *v1.Pod, nodeGroupDescriptor *NodeGroupDescriptor) {
		if len(podShardingLabels) == 0 {
			return
		}

		for labelKey, labelValue := range podShardingLabels {
			if podLabelValue, ok := pod.Spec.NodeSelector[labelKey]; ok && podLabelValue == labelValue {
				nodeGroupDescriptor.Labels[labelKey] = labelValue
			}
		}
	}
}

func provisioningRequestShard(pod *v1.Pod, nodeGroupDescriptor *NodeGroupDescriptor) {
	provClass, found := pr_pods.ProvisioningClassName(pod)
	if !found {
		return
	}
	nodeGroupDescriptor.ProvisioningClassName = provClass
}
