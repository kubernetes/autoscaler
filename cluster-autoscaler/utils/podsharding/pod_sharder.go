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
	"fmt"
	"sort"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

// ShardSignature represents PodShard signature
type ShardSignature string

// NodeGroupDescriptor encapsulates limitations for node groups usable for pod shard.
type NodeGroupDescriptor struct {
	Labels                map[string]string
	SystemLabels          map[string]string
	Taints                []apiv1.Taint
	ExtraResources        map[string]resource.Quantity
	ProvisioningClassName string
}

func (descriptor *NodeGroupDescriptor) signature() ShardSignature {
	builder := strings.Builder{}

	writeMapSorted := func(header string, m map[string]string) {
		if len(m) == 0 {
			return
		}
		builder.WriteString(header)
		builder.WriteString("(")
		var keys []string
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		first := true
		for _, key := range keys {
			if !first {
				builder.WriteString(",")
			}
			first = false
			builder.WriteString(key)
			builder.WriteString("=")
			builder.WriteString(m[key])
		}
		builder.WriteString(")")
	}

	writeMapSorted("Labels", descriptor.Labels)
	writeMapSorted("SystemLabels", descriptor.SystemLabels)

	extraResourcesStringified := make(map[string]string)
	for k, v := range descriptor.ExtraResources {
		extraResourcesStringified[k] = v.String()
	}
	writeMapSorted("ExtraResources", extraResourcesStringified)

	var taintsStringified []string
	for _, taint := range descriptor.Taints {
		taintsStringified = append(taintsStringified, fmt.Sprintf("%v/%v/%v", taint.Key, taint.Value, taint.Effect))
	}
	sort.Strings(taintsStringified)
	if len(taintsStringified) > 0 {
		builder.WriteString("Taints(")
		builder.WriteString(strings.Join(taintsStringified, ","))
		builder.WriteString(")")
	}
	if descriptor.ProvisioningClassName != "" {
		builder.WriteString(fmt.Sprintf("ProvisioningClassName(%s)", descriptor.ProvisioningClassName))
	}
	if builder.Len() == 0 {
		return "_"
	}
	return ShardSignature(builder.String())
}

// PodShard represents a subset of unschedulable pods which can be processed independently from others in scale-up algorithm.
type PodShard struct {
	PodUids             map[types.UID]bool
	NodeGroupDescriptor NodeGroupDescriptor
}

// Signature computes string signature for pod shard
func (shard *PodShard) Signature() ShardSignature {
	return shard.NodeGroupDescriptor.signature()
}

// PodUidsSlice list all pods with entries in in PodShard.PodUids map (with value equal to True) as a slice.
func (shard *PodShard) PodUidsSlice() []types.UID {
	var result []types.UID
	for uid := range shard.PodUids {
		result = append(result, uid)
	}
	return result
}

// PodShardSelector is responsible for selecting shard to be used in given loop iteration
type PodShardSelector interface {
	// SelectPodShard is responsible for selecting shard to be used in given loop iteration
	SelectPodShard(podShards []*PodShard) *PodShard
}

// PodSharder is used to compute sharding for given set of pods
type PodSharder interface {
	// ComputePodShards computes sharding for given set of pods
	ComputePodShards(pods []*apiv1.Pod) []*PodShard
}

// PodFilteringResult represents return value of PodShardFilter.FilterPods.
type PodFilteringResult struct {
	// Pods is the result of filtering
	Pods []*apiv1.Pod
}

// PodShardFilter filters pod list against PodShard
type PodShardFilter interface {
	// FilterPods filters pod list against selected PodShard
	FilterPods(context *context.AutoscalingContext, selectedPodShard *PodShard, allPodShards []*PodShard, pods []*apiv1.Pod) (PodFilteringResult, error)
}

// EmptyNodeGroupDescriptor creates an empty NodeGroupDescriptor
func EmptyNodeGroupDescriptor() NodeGroupDescriptor {
	return NodeGroupDescriptor{
		Labels:         make(map[string]string),
		SystemLabels:   make(map[string]string),
		ExtraResources: make(map[string]resource.Quantity),
	}
}

// DeepCopy crates a deep copy (with copied maps) of NodeGroupDescriptor
func (descriptor NodeGroupDescriptor) DeepCopy() NodeGroupDescriptor {
	copy := EmptyNodeGroupDescriptor()
	for k, v := range descriptor.Labels {
		copy.Labels[k] = v
	}
	for k, v := range descriptor.SystemLabels {
		copy.SystemLabels[k] = v
	}
	copy.Taints = append(copy.Taints, descriptor.Taints...)
	for k, v := range descriptor.ExtraResources {
		copy.ExtraResources[k] = v
	}
	copy.ProvisioningClassName = descriptor.ProvisioningClassName
	return copy
}
