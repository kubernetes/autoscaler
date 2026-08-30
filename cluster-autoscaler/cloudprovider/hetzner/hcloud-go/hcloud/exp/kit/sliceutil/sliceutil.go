package sliceutil

// Transform each element of the given slice and returns a new slice with the result.
//
// Experimental: `exp` package is experimental, breaking changes may occur within minor releases.
func Transform[Slice ~[]E, E any, R any](s Slice, fn func(e E) R) []R {
	result := make([]R, 0, len(s))
	for _, e := range s {
		result = append(result, fn(e))
	}
	return result
}
