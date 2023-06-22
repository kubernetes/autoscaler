package estimator

// BinpackingLimit returns nodes count limit, that is intended to be used as a cap for binpacking.
// Return value of 0 means that no limit is set.
type BinpackingLimit interface {
	GetLimit() int
}
