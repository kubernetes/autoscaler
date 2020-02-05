/*
Copyright 2019 The Kubernetes Authors.

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

package simulator

import (
	"fmt"
	"strings"
)

// PredicateErrorType is type of predicate error
type PredicateErrorType int

const (
	// NotSchedulablePredicateError means that one of the filters retuned that pod does not fit a node
	NotSchedulablePredicateError PredicateErrorType = iota
	// InternalPredicateError denotes internal unexpected error while calling PredicateChecker
	InternalPredicateError
)

// PredicateError is a structure representing error returned from predicate checking simulation.
type PredicateError struct {
	errorType     PredicateErrorType
	predicateName string
	errorMessage  string
	reasons       []string
	// debugInfo contains additional info that predicate doesn't include,
	// but may be useful for debugging (e.g. taints on node blocking scale-up)
	debugInfo func() string
}

// ErrorType returns if error was internal of names predicate failure.
func (pe *PredicateError) ErrorType() PredicateErrorType {
	return pe.errorType
}

// PredicateName return name of predicate which failed.
func (pe *PredicateError) PredicateName() string {
	return pe.predicateName
}

// Message returns error message.
func (pe *PredicateError) Message() string {
	if pe.errorMessage == "" {
		return "unknown error"
	}
	return pe.errorMessage
}

// VerboseMessage generates verbose error message. Building verbose message may be expensive so number of calls should be
// limited.
func (pe *PredicateError) VerboseMessage() string {
	return fmt.Sprintf(
		"%s; predicateName=%s; reasons: %s; debugInfo=%s",
		pe.Message(),
		pe.predicateName,
		strings.Join(pe.reasons, ", "),
		pe.debugInfo())
}

// Reasons returns failure reasons from failed predicate as a slice of strings.
func (pe *PredicateError) Reasons() []string {
	return pe.reasons
}

// NewPredicateError creates a new predicate error from error and reasons.
func NewPredicateError(
	errorType PredicateErrorType,
	predicateName string,
	errorMessage string,
	reasons []string,
	debugInfo func() string,
) *PredicateError {
	return &PredicateError{
		errorType:     errorType,
		predicateName: predicateName,
		errorMessage:  errorMessage,
		reasons:       reasons,
		debugInfo:     debugInfo,
	}
}

// GenericPredicateError return a generic instance of PredicateError to be used in context where predicate name is not
// know.
func GenericPredicateError() *PredicateError {
	return &PredicateError{
		errorType:    NotSchedulablePredicateError,
		errorMessage: "generic predicate failure",
	}
}

func emptyString() string {
	return ""
}
