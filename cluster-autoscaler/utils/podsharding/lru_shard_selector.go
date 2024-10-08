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

// LruPodShardSelector is a PodShardSelector which picks shards in least-recently picked order
type LruPodShardSelector struct {
	lruMap     map[ShardSignature]uint64
	useCounter uint64
}

// NewLruPodShardSelector creates PodShardSelector which picks shards in least-recently picked order
func NewLruPodShardSelector() PodShardSelector {
	return &LruPodShardSelector{
		lruMap: make(map[ShardSignature]uint64),
	}
}

// SelectPodShard selects shard with least recently used shard signature
func (selector *LruPodShardSelector) SelectPodShard(shards []*PodShard) *PodShard {
	if len(shards) == 0 {
		return nil
	}

	var selectedShard *PodShard

	// fill in entries for unknown shards; those are added to the end of the queue
	for _, shard := range shards {
		signature := shard.Signature()
		if _, found := selector.lruMap[signature]; !found {
			selector.useCounter++
			selector.lruMap[signature] = selector.useCounter
		}
	}

	// select oldest entry
	selector.useCounter++
	selectedShardTime := selector.useCounter
	for _, shard := range shards {
		shardTime := selector.lruMap[shard.Signature()]
		if shardTime < selectedShardTime {
			selectedShard = shard
			selectedShardTime = shardTime
		}
	}
	selector.lruMap[selectedShard.Signature()] = selector.useCounter
	return selectedShard
}
