// This file is part of gobizfly

package gobizfly

// SliceContains - Check data in slice
func SliceContains(slice interface{}, val interface{}) (int, bool) {
	switch v := slice.(type) {
	case string:
		if slice == val {
			return 1, true
		}
		return -1, false
	default:
		for i, item := range v.([]string) {
			if item == val {
				return i, true
			}
		}
		return -1, false
	}
}
