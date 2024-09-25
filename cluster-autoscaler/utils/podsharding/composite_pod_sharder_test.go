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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
)

type shardFeatureParts map[string]string

type podFeatureParts []struct {
	uid        apitypes.UID
	nodeLables shardFeatureParts
}
type expectedShards map[apitypes.UID][]ShardSignature // uid -> list of shards

func TestCompositePodSharderBasic(t *testing.T) {
	tests := []struct {
		name           string
		podsFeatures   podFeatureParts
		expectedShards expectedShards
	}{
		{
			name:           "no pods",
			podsFeatures:   nil,
			expectedShards: nil,
		},
		{
			name: "only default shard",
			podsFeatures: podFeatureParts{
				{"p1", shardFeatureParts{}},
				{"p2", shardFeatureParts{}},
				{"p3", shardFeatureParts{}},
			},
			expectedShards: expectedShards{
				"p1": []ShardSignature{"_"},
				"p2": []ShardSignature{"_"},
				"p3": []ShardSignature{"_"},
			},
		},
		{
			name: "only non default shard",
			podsFeatures: podFeatureParts{
				{"p1", shardFeatureParts{"f1": "v1", "f2": "v2"}},
				{"p2", shardFeatureParts{"f1": "v1", "f2": "v2"}},
				{"p3", shardFeatureParts{"f1": "v1", "f2": "v2"}},
			},
			expectedShards: expectedShards{
				"p1": []ShardSignature{"Labels(f1=v1,f2=v2)"},
				"p2": []ShardSignature{"Labels(f1=v1,f2=v2)"},
				"p3": []ShardSignature{"Labels(f1=v1,f2=v2)"},
			},
		},
		{
			name: "mixed default and non default",
			podsFeatures: podFeatureParts{
				{"p1", shardFeatureParts{}},
				{"p2", shardFeatureParts{"f1": "v1", "f2": "v2"}},
				{"p3", shardFeatureParts{"f1": "v1", "f2": "v2"}},
			},
			expectedShards: expectedShards{
				"p1": []ShardSignature{"_"},
				"p2": []ShardSignature{"Labels(f1=v1,f2=v2)"},
				"p3": []ShardSignature{"Labels(f1=v1,f2=v2)"},
			},
		},
		{
			name: "different values for one feature",
			podsFeatures: podFeatureParts{
				{"p1", shardFeatureParts{"f1": "v1", "f2": "v2"}},
				{"p2", shardFeatureParts{"f1": "v1", "f2": "v2"}},
				{"p3", shardFeatureParts{"f1": "v1", "f2": "v3"}},
				{"p4", shardFeatureParts{"f1": "v1"}},
			},
			expectedShards: expectedShards{
				"p1": []ShardSignature{"Labels(f1=v1,f2=v2)"},
				"p2": []ShardSignature{"Labels(f1=v1,f2=v2)"},
				"p3": []ShardSignature{"Labels(f1=v1,f2=v3)"},
				"p4": []ShardSignature{"Labels(f1=v1)"},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var computeFunctions []FeatureShardComputeFunction
			for _, feature := range getAllFeaturesSorted(tc.podsFeatures) {
				featureCopy := feature
				computeFunctions = append(computeFunctions,
					FeatureShardComputeFunction{
						Feature: featureCopy,
						Function: func(pod *v1.Pod, nodeGroupDescriptor *NodeGroupDescriptor) {
							for _, podFeatures := range tc.podsFeatures {
								if podFeatures.uid == pod.UID {
									for k, v := range podFeatures.nodeLables {
										nodeGroupDescriptor.Labels[k] = v
									}
								}
							}
						},
					})
			}

			podSharder := NewCompositePodSharder(computeFunctions)

			var pods []*v1.Pod
			for _, podFeaturesEntry := range tc.podsFeatures {
				pods = append(pods, podForUID(podFeaturesEntry.uid))
			}
			podShards := podSharder.ComputePodShards(pods)

			allShardSignatures := allExpectedShardSignatures(tc.expectedShards)
			resultShardSignatures := getShardSignatures(podShards)

			assert.ElementsMatch(t, allShardSignatures, resultShardSignatures, "List of shards %v does not match expected %v one", resultShardSignatures, allShardSignatures)
			for _, podShard := range podShards {
				expectedUidsForShard := expectedUidsForShard(tc.expectedShards, podShard.Signature())
				assert.ElementsMatch(t, expectedUidsForShard, podShard.PodUidsSlice(), "List of pods does not match expected one for shard %s", podShard.Signature())
			}
		})
	}
}

func TestCompositePodSharderNodeDescriptor(t *testing.T) {
	buildComputeFunction := func(labelKey string) FeatureShardComputeFunction {
		return FeatureShardComputeFunction{
			Feature: labelKey,
			Function: func(pod *v1.Pod, nodeGroupDescriptor *NodeGroupDescriptor) {
				for podLabelKey, podLabelValue := range pod.Labels {
					if podLabelKey == labelKey {
						nodeGroupDescriptor.Labels[podLabelKey] = podLabelValue

					}
				}
			},
		}
	}

	computeFunctions := []FeatureShardComputeFunction{
		buildComputeFunction("a"),
		buildComputeFunction("b"),
	}

	sharder := NewCompositePodSharder(computeFunctions)

	p1 := podForUID("p1")
	p1.Labels = make(map[string]string)
	p1.Labels["a"] = "a1"
	p1.Labels["b"] = "b1"

	p2 := podForUID("p2")
	p2.Labels = make(map[string]string)
	p2.Labels["a"] = "a1"
	p2.Labels["b"] = "b1"

	p3 := podForUID("p3")
	p3.Labels = make(map[string]string)
	p3.Labels["a"] = "a1"
	p3.Labels["b"] = "b2"

	// p1 and p2 are in single shard (a1/b1)
	// p3 is in separate shard (a1/b2)

	shards := sharder.ComputePodShards([]*v1.Pod{p1, p2, p3})
	assert.ElementsMatch(t, []ShardSignature{"Labels(a=a1,b=b1)", "Labels(a=a1,b=b2)"}, getShardSignatures(shards))

	shard1 := getShardForSignature("Labels(a=a1,b=b1)", shards)
	shard2 := getShardForSignature("Labels(a=a1,b=b2)", shards)

	assertNodeGroupDescriptorEqual(t, NodeGroupDescriptor{
		Labels: map[string]string{
			"a": "a1",
			"b": "b1",
		},
	}, shard1.NodeGroupDescriptor)

	assertNodeGroupDescriptorEqual(t, NodeGroupDescriptor{
		Labels: map[string]string{
			"a": "a1",
			"b": "b2",
		},
	}, shard2.NodeGroupDescriptor)
}

func getShardSignatures(podShards []*PodShard) []ShardSignature {
	var resultShardSignatures []ShardSignature
	for _, podShard := range podShards {
		resultShardSignatures = append(resultShardSignatures, podShard.Signature())
	}
	return resultShardSignatures
}

func getShardForSignature(signature ShardSignature, podShards []*PodShard) *PodShard {
	for _, podShard := range podShards {
		if podShard.Signature() == signature {
			return podShard
		}
	}
	return nil
}
func allExpectedShardSignatures(expectedShards expectedShards) []ShardSignature {
	allShardSignaturesSet := make(map[ShardSignature]bool)
	for _, shardSignatures := range expectedShards {
		for _, shardSignature := range shardSignatures {
			allShardSignaturesSet[shardSignature] = true
		}
	}
	var allShardSignatures []ShardSignature
	for shardSignature := range allShardSignaturesSet {
		allShardSignatures = append(allShardSignatures, shardSignature)
	}
	return allShardSignatures
}

func expectedUidsForShard(expectedShards expectedShards, shardSignature ShardSignature) []apitypes.UID {
	var uidsForShards []apitypes.UID
	for uid, shardSignatures := range expectedShards {
		if containsShard(shardSignatures, shardSignature) {
			uidsForShards = append(uidsForShards, uid)
		}
	}
	return uidsForShards
}

func containsShard(slice []ShardSignature, element ShardSignature) bool {
	for _, sliceElement := range slice {
		if element == sliceElement {
			return true
		}
	}
	return false
}

func getAllFeaturesSorted(features podFeatureParts) []string {
	allFeaturesSet := make(map[string]bool)
	for _, podFeaturesEntry := range features {
		for feature := range podFeaturesEntry.nodeLables {
			allFeaturesSet[feature] = true
		}
	}
	var allFeatures []string
	for feature := range allFeaturesSet {
		allFeatures = append(allFeatures, feature)
	}
	sort.Strings(allFeatures)
	return allFeatures
}
