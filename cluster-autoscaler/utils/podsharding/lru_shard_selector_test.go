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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLruShardSelector(t *testing.T) {
	selector := NewLruPodShardSelector()
	podShards := generatePodShards("a", "b", "c")

	var selectedShards1 []*PodShard
	for i := 0; i < len(podShards); i++ {
		selectedShards1 = append(selectedShards1, selector.SelectPodShard(podShards))
	}
	assert.ElementsMatch(t, signatures(podShards...), signatures(selectedShards1...), "not all shards selected")

	// another round of selection should be the same (including order)
	var selectedShards2 []*PodShard
	for i := 0; i < len(podShards); i++ {
		selectedShards2 = append(selectedShards2, selector.SelectPodShard(podShards))
	}
	assert.EqualValues(t, signatures(selectedShards1...), signatures(selectedShards2...), "different order of shard selection in second round")

	// we add new shard. It should be picked last.
	podShards = generatePodShards("a", "b", "c", "d")
	var selectedShards3 []*PodShard
	for i := 0; i < len(podShards)-1; i++ {
		selectedShards3 = append(selectedShards3, selector.SelectPodShard(podShards))
	}
	assert.EqualValues(t, signatures(selectedShards1...), signatures(selectedShards3...), "different order of shard selection for first 3 queries on third round")
	assert.Equal(t, ShardSignature("Labels(key=d)"), selector.SelectPodShard(podShards).Signature(), "newly added shard was not picked last")
}

func generatePodShards(lables ...string) []*PodShard {
	var result []*PodShard
	for _, label := range lables {
		result = append(result, &PodShard{
			NodeGroupDescriptor: NodeGroupDescriptor{
				Labels: map[string]string{
					"key": label,
				},
			},
		})
	}
	return result
}

func signatures(podShards ...*PodShard) []ShardSignature {
	var result []ShardSignature
	for _, podShard := range podShards {
		result = append(result, podShard.Signature())
	}
	return result
}
