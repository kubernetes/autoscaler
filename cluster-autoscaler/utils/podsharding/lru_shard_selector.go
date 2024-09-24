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
