package testing

import (
	"fmt"
	"net/http"
	"strings"
)

// HasHeader compares the header values and identifies if the actual header
// set includes all values specified in the expect set. Returns an error if not.
func HasHeader(expect, actual http.Header) error {
	var errs errors
	for key, es := range expect {
		as := actual.Values(key)
		if len(as) == 0 {
			errs = append(errs, fmt.Errorf("expect %v header in %v",
				key, actual))
			continue
		}

		// Join the list of header values together for consistent
		// comparison between common separated sets, and individual header
		// key/value pairs repeated.
		e := strings.Join(es, ", ")
		a := strings.Join(as, ", ")
		if e != a {
			errs = append(errs, fmt.Errorf("expect %v=%v to match %v",
				key, e, a))
			continue
		}
	}

	return errs
}

// AssertHasHeader compares the header values and identifies if the actual
// header set includes all values specified in the expect set. Emits a testing
// error, and returns false if the headers are not equal.
func AssertHasHeader(t T, expect, actual http.Header) bool {
	t.Helper()

	if err := HasHeader(expect, actual); err != nil {
		for _, e := range err.(errors) {
			t.Error(e)
		}
		return false
	}
	return true
}

// HasHeaderKeys validates that header set contains all keys expected. Returns
// an error if a header key is not in the header set.
func HasHeaderKeys(keys []string, actual http.Header) error {
	var errs errors
	for _, key := range keys {
		if vs := actual.Values(key); len(vs) == 0 {
			errs = append(errs, fmt.Errorf("expect %v key in %v", key, actual))
			continue
		}
	}
	return errs
}

// AssertHasHeaderKeys validates that header set contains all keys expected.
// Emits a testing error, and returns false if the headers are not equal.
func AssertHasHeaderKeys(t T, keys []string, actual http.Header) bool {
	t.Helper()

	if err := HasHeaderKeys(keys, actual); err != nil {
		for _, e := range err.(errors) {
			t.Error(e)
		}
		return false
	}
	return true
}

// NotHaveHeaderKeys validates that header set does not contains any of the
// keys. Returns an error if a header key is found in the header set.
func NotHaveHeaderKeys(keys []string, actual http.Header) error {
	var errs errors
	for _, k := range keys {
		if vs := actual.Values(k); len(vs) != 0 {
			errs = append(errs, fmt.Errorf("expect %v key not in %v", k, actual))
			continue
		}
	}
	return errs
}

// AssertNotHaveHeaderKeys validates that header set does not contains any of
// the keys. Emits a testing error, and returns false if the header contains
// the any of the keys equal.
func AssertNotHaveHeaderKeys(t T, keys []string, actual http.Header) bool {
	t.Helper()

	if err := NotHaveHeaderKeys(keys, actual); err != nil {
		for _, e := range err.(errors) {
			t.Error(e)
		}
		return false
	}
	return true
}

// QueryItem provides an escaped key and value struct for values from a raw
// query string.
type QueryItem struct {
	Key   string
	Value string
}

// ParseRawQuery returns a slice of QueryItems extracted from the raw query
// string. The parsed QueryItems preserve escaping of key and values.
//
// All + escape characters are replaced with %20 for consistent escaping
// pattern.
func ParseRawQuery(rawQuery string) (items []QueryItem) {
	for _, item := range strings.Split(rawQuery, `&`) {
		parts := strings.SplitN(item, `=`, 2)
		var value string
		if len(parts) > 1 {
			value = parts[1]
		}
		items = append(items, QueryItem{
			// Go Query encoder escapes space as `+` whereas smithy protocol
			// tests expect `%20`.
			Key:   strings.ReplaceAll(parts[0], `+`, `%20`),
			Value: strings.ReplaceAll(value, `+`, `%20`),
		})
	}
	return items
}

// HasQuery validates that the expected set of query items are present in
// the actual set. Returns an error if any of the expected set are not found in
// the actual.
func HasQuery(expect, actual []QueryItem) error {
	var errs errors
	for _, item := range expect {
		var found bool
		for _, v := range actual {
			if item.Key == v.Key && item.Value == v.Value {
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, fmt.Errorf("expect %v query item in %v", item, actual))
		}
	}
	return errs
}

// AssertHasQuery validates that the expected set of query items are
// present in the actual set. Emits a testing error, and returns false if any
// of the expected set are not found in the actual.
func AssertHasQuery(t T, expect, actual []QueryItem) bool {
	t.Helper()

	if err := HasQuery(expect, actual); err != nil {
		for _, e := range err.(errors) {
			t.Error(e.Error())
		}
		return false
	}
	return true
}

// HasQueryKeys validates that the actual set of query items contains the keys
// provided. Returns an error if any key is not found.
func HasQueryKeys(keys []string, actual []QueryItem) error {
	var errs errors
	for _, key := range keys {
		var found bool
		for _, v := range actual {
			if key == v.Key {
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, fmt.Errorf("expect %v query key in %v", key, actual))
		}
	}
	return errs
}

// AssertHasQueryKeys validates that the actual set of query items contains the
// keys provided. Emits a testing error if any key is not found.
func AssertHasQueryKeys(t T, keys []string, actual []QueryItem) bool {
	t.Helper()

	if err := HasQueryKeys(keys, actual); err != nil {
		for _, e := range err.(errors) {
			t.Error(e)
		}
		return false
	}
	return true
}

// NotHaveQueryKeys validates that the actual set of query items does not
// contain the keys provided. Returns an error if any key is found.
func NotHaveQueryKeys(keys []string, actual []QueryItem) error {
	var errs errors
	for _, key := range keys {
		for _, v := range actual {
			if key == v.Key {
				errs = append(errs, fmt.Errorf("expect %v query key not in %v",
					key, actual))
				continue
			}
		}
	}
	return errs
}

// AssertNotHaveQueryKeys validates that the actual set of query items does not
// contains the keys provided. Emits a testing error if any key is found.
func AssertNotHaveQueryKeys(t T, keys []string, actual []QueryItem) bool {
	t.Helper()

	if err := NotHaveQueryKeys(keys, actual); err != nil {
		for _, e := range err.(errors) {
			t.Error(e)
		}
		return false
	}
	return true
}
