/*
Copyright 2016 The Kubernetes Authors.

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

// ErrorCode is the API error code type.
type PredicateErrorType int

// API error codes.
const (
	// Factory errors
	NotSchedulablePredicateError PredicateErrorType = iota
	InternalPredicateError
	GenericError
)

// PredicateError represents an error during predicate checking.
type PredicateError interface {
	ErrorType() PredicateErrorType
	PredicateName() string
	Message() string
	VerboseMessage() string
	Reasons() []string
}

// PredicateError passes error information regarding predicate checking
type simplePredicateError struct {
	errorType     PredicateErrorType
	predicateName string
	errorMessage  string
	reasons       []string
	// debugInfo contains additional info that predicate doesn't include,
	// but may be useful for debugging (e.g. taints on node blocking scale-up)
	debugInfo func() string
}

// PredicateName return name of predicate which failed.
func (pe *simplePredicateError) ErrorType() PredicateErrorType {
	return pe.errorType
}

// PredicateName return name of predicate which failed.
func (pe *simplePredicateError) PredicateName() string {
	return pe.predicateName
}

// Message returns error message.
func (pe *simplePredicateError) Message() string {
	if pe.errorMessage == "" {
		return "unknown error"
	} else {
		return pe.errorMessage
	}
}

// VerboseMessage generates verbose error message. Building verbose message may be expensive so number of calls should be
// limited.
func (pe *simplePredicateError) VerboseMessage() string {
	return fmt.Sprintf(
		"%s; predicateName=%s; reasons: %s; debugInfo=%s",
		pe.Message(),
		pe.predicateName,
		strings.Join(pe.reasons, ", "),
		pe.debugInfo())
}

// Reasons returns failure reasons from failed predicate as a slice of strings.
func (pe *simplePredicateError) Reasons() []string {
	return pe.reasons
}

// NewPredicateError creates a new predicate error from error and reasons.
func NewPredicateError(
	errorType PredicateErrorType,
	predicateName string,
	errorMessage string,
	reasons []string,
	debugInfo func() string,
) PredicateError {
	return &simplePredicateError{
		errorType:     errorType,
		predicateName: predicateName,
		errorMessage:  errorMessage,
		reasons:       reasons,
		debugInfo:     debugInfo,
	}
}

// GenericPredicateError return a generic instance of PredicateError to be used in context where predicate name is
// unknown (hack - to be removed)
func GenericPredicateError() PredicateError {
	return &simplePredicateError{
		errorType:    GenericError,
		errorMessage: "generic error",
	}
}

func emptyString() string {
	return ""
}
