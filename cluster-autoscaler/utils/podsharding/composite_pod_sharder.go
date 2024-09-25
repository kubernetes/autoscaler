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
	apitypes "k8s.io/apimachinery/pkg/types"
)

// FeatureShardComputeFunction function computes part of shard to which pod belongs, looking at specific feature.
// If function does not detect that pod should be separated from others based on handled feature it should return empty string
type FeatureShardComputeFunction struct {
	// Feature is feature name
	Feature string

	// Function is used to compute shard value for feature.
	// Result of computation is mutation done to passed NodeGroupDescriptor.
	// After mutations for all features are applied for given pod we end up with final NodeGroupDescriptor which represents
	// a shard pod belongs to.
	// NodeGroupDescriptor for selected shard is also later used for building NodeInfo to be used for extending set of pods in
	// shard using predicate checking.
	Function func(*v1.Pod, *NodeGroupDescriptor)
}

// CompositePodSharder is an implementation of PodSharder based on list of features with corresponding
// functions to compute shard value for given feature.
type CompositePodSharder struct {
	featureShardComputeFunctions []FeatureShardComputeFunction
}

// NewCompositePodSharder creates new instance of CompositePodSharder
func NewCompositePodSharder(computeFunctions []FeatureShardComputeFunction) *CompositePodSharder {
	return &CompositePodSharder{
		featureShardComputeFunctions: append([]FeatureShardComputeFunction{}, computeFunctions...),
	}
}

// ComputePodShards computes sharding for given set of pods
func (sharder *CompositePodSharder) ComputePodShards(pods []*v1.Pod) []*PodShard {
	podShards := make(map[ShardSignature]*PodShard)

	for _, pod := range pods {
		nodeGroupDescriptor := EmptyNodeGroupDescriptor()
		for _, computeFunction := range sharder.featureShardComputeFunctions {
			computeFunction.Function(pod, &nodeGroupDescriptor)
		}

		shardSignature := nodeGroupDescriptor.signature()
		if _, found := podShards[shardSignature]; !found {
			podShards[shardSignature] = &PodShard{
				PodUids:             make(map[apitypes.UID]bool),
				NodeGroupDescriptor: nodeGroupDescriptor,
			}
		}
		podShards[shardSignature].PodUids[pod.UID] = true
	}

	var result []*PodShard
	for _, podShard := range podShards {
		result = append(result, podShard)
	}
	return result
}
