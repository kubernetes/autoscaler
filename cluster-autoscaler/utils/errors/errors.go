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
	errorType AutoscalerErrorType
	msg       string
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
)

// NewAutoscalerError returns new autoscaler error with a message constructed from format string
func NewAutoscalerError(errorType AutoscalerErrorType, msg string, args ...interface{}) AutoscalerError {
	return autoscalerErrorImpl{
		errorType: errorType,
		msg:       fmt.Sprintf(msg, args...),
	}
}

// ToAutoscalerError converts an error to AutoscalerError with given type,
// unless it already is an AutoscalerError (in which case it's not modified).
func ToAutoscalerError(defaultType AutoscalerErrorType, err error) AutoscalerError {
	if e, ok := err.(AutoscalerError); ok {
		return e
	}
	return NewAutoscalerError(defaultType, "%v", err)
}

// Error implements golang error interface
func (e autoscalerErrorImpl) Error() string {
	return e.msg
}

// Type returns the type of AutoscalerError
func (e autoscalerErrorImpl) Type() AutoscalerErrorType {
	return e.errorType
}

// AddPrefix adds a prefix to error message.
// Returns the error it's called for convenient inline use.
// Example:
// if err := DoSomething(myObject); err != nil {
//	return err.AddPrefix("can't do something with %v: ", myObject)
// }
func (e autoscalerErrorImpl) AddPrefix(msg string, args ...interface{}) AutoscalerError {
	e.msg = fmt.Sprintf(msg, args...) + e.msg
	return e
}
