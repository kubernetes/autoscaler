/*
Copyright 2017 The Kubernetes Authors.

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

package errors

import (
	"fmt"
	"sort"
	"strings"
)

// AutoscalerErrorType describes a high-level category of a given error
type AutoscalerErrorType string

// AutoscalerError contains information about Autoscaler errors
type AutoscalerError interface {
	// Error implements golang error interface
	Error() string

	// Type returns the type of AutoscalerError
	Type() AutoscalerErrorType

	// AddPrefix adds a prefix to error message.
	// Returns the error it's called for convenient inline use.
	// Example:
	// if err := DoSomething(myObject); err != nil {
	//	return err.AddPrefix("can't do something with %v: ", myObject)
	// }
	AddPrefix(msg string, args ...interface{}) AutoscalerError
}

type autoscalerErrorImpl struct {
	errorType  AutoscalerErrorType
	wrappedErr error
	msg        string
}

const (
	// CloudProviderError is an error related to underlying infrastructure
	CloudProviderError AutoscalerErrorType = "cloudProviderError"
	// ApiCallError is an error related to communication with k8s API server
	ApiCallError AutoscalerErrorType = "apiCallError"
	// InternalError is an error inside Cluster Autoscaler
	InternalError AutoscalerErrorType = "internalError"
	// TransientError is an error that causes us to skip a single loop, but
	// does not require any additional action.
	TransientError AutoscalerErrorType = "transientError"
	// ConfigurationError is an error related to bad configuration provided
	// by a user.
	ConfigurationError AutoscalerErrorType = "configurationError"
	// NodeGroupDoesNotExistError signifies that a NodeGroup
	// does not exist.
	NodeGroupDoesNotExistError AutoscalerErrorType = "nodeGroupDoesNotExistError"
	// UnexpectedScaleDownStateError means Cluster Autoscaler thinks ongoing
	// scale down is already removing too much and so further node removals
	// shouldn't be attempted.
	UnexpectedScaleDownStateError AutoscalerErrorType = "unexpectedScaleDownStateError"
)

// NewAutoscalerError returns new autoscaler error with a message constructed from string
func NewAutoscalerError(errorType AutoscalerErrorType, msg string) AutoscalerError {
	return autoscalerErrorImpl{
		errorType: errorType,
		msg:       msg,
	}
}

// NewAutoscalerErrorf returns new autoscaler error with a message constructed from format string
func NewAutoscalerErrorf(errorType AutoscalerErrorType, msg string, args ...interface{}) AutoscalerError {
	return autoscalerErrorImpl{
		errorType: errorType,
		msg:       fmt.Sprintf(msg, args...),
	}
}

// ToAutoscalerError wraps an error to AutoscalerError with given type,
// unless it already is an AutoscalerError (in which case it's not modified).
// errors.Is() works correctly on the wrapped error.
func ToAutoscalerError(defaultType AutoscalerErrorType, err error) AutoscalerError {
	if e, ok := err.(AutoscalerError); ok {
		return e
	}
	if err == nil {
		return nil
	}
	return autoscalerErrorImpl{
		errorType:  defaultType,
		wrappedErr: err,
	}
}

// Error implements golang error interface
func (e autoscalerErrorImpl) Error() string {
	msg := e.msg
	if e.wrappedErr != nil {
		msg = msg + e.wrappedErr.Error()
	}
	return msg
}

// Unwrap returns the error wrapped via ToAutoscalerError or AddPrefix, so that errors.Is() works
// correctly.
func (e autoscalerErrorImpl) Unwrap() error {
	return e.wrappedErr
}

// Type returns the type of AutoscalerError
func (e autoscalerErrorImpl) Type() AutoscalerErrorType {
	return e.errorType
}

// AddPrefix returns an error wrapping this one, with the added prefix. errors.Is() works
// correctly on the wrapped error.
// Example usage:
// if err := DoSomething(myObject); err != nil {
//
//	return err.AddPrefix("can't do something with %v: ", myObject)
//
// }
func (e autoscalerErrorImpl) AddPrefix(msg string, args ...interface{}) AutoscalerError {
	return autoscalerErrorImpl{errorType: e.errorType, wrappedErr: e, msg: fmt.Sprintf(msg, args...)}
}

// Combine returns combined error to report from multiple errors.
func Combine(errs []AutoscalerError) AutoscalerError {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	uniqueMessages := make(map[string]bool)
	uniqueTypes := make(map[AutoscalerErrorType]bool)
	for _, err := range errs {
		uniqueTypes[err.Type()] = true
		uniqueMessages[err.Error()] = true
	}
	if len(uniqueTypes) == 1 && len(uniqueMessages) == 1 {
		return errs[0]
	}
	// sort to stabilize the results and easier log aggregation
	sort.Slice(errs, func(i, j int) bool {
		errA := errs[i]
		errB := errs[j]
		if errA.Type() == errB.Type() {
			return errs[i].Error() < errs[j].Error()
		}
		return errA.Type() < errB.Type()
	})
	firstErr := errs[0]
	printErrorTypes := len(uniqueTypes) > 1
	message := formatMessageFromErrors(errs, printErrorTypes)
	return NewAutoscalerError(firstErr.Type(), message)
}

func formatMessageFromErrors(errs []AutoscalerError, printErrorTypes bool) string {
	firstErr := errs[0]
	var builder strings.Builder
	builder.WriteString(firstErr.Error())
	builder.WriteString(" ...and other errors: [")
	formattedErrs := map[AutoscalerError]bool{
		firstErr: true,
	}
	for _, err := range errs {
		if _, has := formattedErrs[err]; has {
			continue
		}
		formattedErrs[err] = true
		var message string
		if printErrorTypes {
			message = fmt.Sprintf("[%s] %s", err.Type(), err.Error())
		} else {
			message = err.Error()
		}
		if len(formattedErrs) > 2 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%q", message))
	}
	builder.WriteString("]")
	return builder.String()
}
