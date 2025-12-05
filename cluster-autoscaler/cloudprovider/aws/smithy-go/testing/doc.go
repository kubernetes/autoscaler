// Package testing provides utilities for testing smith clients and protocols.
package testing

// T provides the testing interface for capturing failures with testing assert
// utilities.
type T interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Helper()
}
