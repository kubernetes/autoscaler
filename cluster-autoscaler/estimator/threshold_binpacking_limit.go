package estimator

type thresholdBinpackingLimit struct {
	maxNodes int
}

func (l *thresholdBinpackingLimit) GetLimit() int {
	return l.maxNodes
}

func NewThresholdBinpackingLimit(maxNodes int) BinpackingLimit {
	return &thresholdBinpackingLimit{
		maxNodes: maxNodes,
	}
}
