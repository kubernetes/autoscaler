/*
Copyright 2019 The Kubernetes Authors.

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

package utils

import (
	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	klog "k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// GetNodeGroupSizeMap return a map of node group id and its target size
func GetNodeGroupSizeMap(cloudProvider cloudprovider.CloudProvider) map[string]int {
	nodeGroupSize := make(map[string]int)
	for _, nodeGroup := range cloudProvider.NodeGroups() {
		size, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Errorf("Error while checking node group size %s: %v", nodeGroup.Id(), err)
			continue
		}
		nodeGroupSize[nodeGroup.Id()] = size
	}
	return nodeGroupSize
}

// FilterOutNodes filters out nodesToFilterOut from nodes
func FilterOutNodes(nodes []*apiv1.Node, nodesToFilterOut []*apiv1.Node) []*apiv1.Node {
	var filtered []*apiv1.Node
	for _, node := range nodes {
		found := false
		for _, nodeToFilter := range nodesToFilterOut {
			if nodeToFilter.Name == node.Name {
				found = true
			}
		}
		if !found {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

// PodSpecSemanticallyEqual returns true if two pod specs are similar after dropping
// the fields we don't care about
// Due to the generated suffixes, a strict DeepEquals check will fail and generate
// an equivalence group per pod which is undesirable.
// Projected volumes do not impact scheduling so we should ignore them
func PodSpecSemanticallyEqual(p1 apiv1.PodSpec, p2 apiv1.PodSpec) bool {
	p1Spec := sanitizePodSpec(p1)
	p2Spec := sanitizePodSpec(p2)
	return apiequality.Semantic.DeepEqual(p1Spec, p2Spec)
}

func sanitizePodSpec(podSpec apiv1.PodSpec) apiv1.PodSpec {
	dropProjectedVolumesAndMounts(&podSpec)
	dropHostname(&podSpec)
	dropEnv(&podSpec)
	return podSpec
}

func dropEnv(podSpec *apiv1.PodSpec) {
	for i := range podSpec.Containers {
		podSpec.Containers[i].Env = nil
	}
	for i := range podSpec.InitContainers {
		podSpec.InitContainers[i].Env = nil
	}
}

func dropProjectedVolumesAndMounts(podSpec *apiv1.PodSpec) {
	projectedVolumeNames := map[string]bool{}
	var volumes []apiv1.Volume
	for _, v := range podSpec.Volumes {
		if v.Projected == nil {
			volumes = append(volumes, v)
		} else {
			projectedVolumeNames[v.Name] = true
		}
	}
	podSpec.Volumes = volumes

	for i := range podSpec.Containers {
		var volumeMounts []apiv1.VolumeMount
		for _, mount := range podSpec.Containers[i].VolumeMounts {
			if ok := projectedVolumeNames[mount.Name]; !ok {
				volumeMounts = append(volumeMounts, mount)
			}
		}
		podSpec.Containers[i].VolumeMounts = volumeMounts
	}

	for i := range podSpec.InitContainers {
		var volumeMounts []apiv1.VolumeMount
		for _, mount := range podSpec.InitContainers[i].VolumeMounts {
			if ok := projectedVolumeNames[mount.Name]; !ok {
				volumeMounts = append(volumeMounts, mount)
			}
		}
		podSpec.InitContainers[i].VolumeMounts = volumeMounts
	}
}

func dropHostname(podSpec *apiv1.PodSpec) {
	podSpec.Hostname = ""
}
