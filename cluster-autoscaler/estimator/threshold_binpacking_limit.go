package estimator

type thresholdBinpackingLimit struct {
	maxNodes int
}

func (l *thresholdBinpackingLimit) GetLimit() int {
	return l.maxNodes
}

// NewThresholdBinpackingLimit returns a BinpackingLimit that caps maximum node
// count by the given static value
func NewThresholdBinpackingLimit(maxNodes int) BinpackingLimit {
	return &thresholdBinpackingLimit{
		maxNodes: maxNodes,
	}
}
