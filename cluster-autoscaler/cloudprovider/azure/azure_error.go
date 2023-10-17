/*
Copyright 2023 The Kubernetes Authors.

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

package azure

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// Unknown is for errors that have nil RawError body
const Unknown CloudProviderErrorReason = "Unknown"

// Errors on the sync path
const (
	// QuotaExceeded falls under OperationNotAllowed error code but we make it more specific here
	QuotaExceeded CloudProviderErrorReason = "QuotaExceeded"
	// OperationNotAllowed is an umbrella for a lot of errors returned by Azure
	OperationNotAllowed string = "OperationNotAllowed"
)

// AutoscalerErrorType describes a high-level category of a given error
type AutoscalerErrorType string

// AutoscalerErrorReason is a more detailed reason for the failed operation
type AutoscalerErrorReason string

// CloudProviderErrorReason providers more details on errors of type CloudProviderError
type CloudProviderErrorReason AutoscalerErrorReason

// AutoscalerError contains information about Autoscaler errors
type AutoscalerError interface {
	// Error implements golang error interface
	Error() string

	// Type returns the type of AutoscalerError
	Type() AutoscalerErrorType

	// Reason returns the reason of the AutoscalerError
	Reason() AutoscalerErrorReason

	// AddPrefix adds a prefix to error message.
	// Returns the error it's called for convenient inline use.
	// Example:
	// if err := DoSomething(myObject); err != nil {
	//	return err.AddPrefix("can't do something with %v: ", myObject)
	// }
	AddPrefix(msg string, args ...interface{}) AutoscalerError
}

type autoscalerErrorImpl struct {
	errorType   AutoscalerErrorType
	errorReason AutoscalerErrorReason
	msg         string
}

const (
	// CloudProviderError is an error related to underlying infrastructure
	CloudProviderError AutoscalerErrorType = "CloudProviderError"
	// ApiCallError is an error related to communication with k8s API server
	ApiCallError AutoscalerErrorType = "ApiCallError"
	// Timeout is an error related to nodes not joining the cluster in maxNodeProvisionTime
	Timeout AutoscalerErrorType = "Timeout"
	// InternalError is an error inside Cluster Autoscaler
	InternalError AutoscalerErrorType = "InternalError"
	// TransientError is an error that causes us to skip a single loop, but
	// does not require any additional action.
	TransientError AutoscalerErrorType = "TransientError"
	// ConfigurationError is an error related to bad configuration provided
	// by a user.
	ConfigurationError AutoscalerErrorType = "ConfigurationError"
	// NodeGroupDoesNotExistError signifies that a NodeGroup
	// does not exist.
	NodeGroupDoesNotExistError AutoscalerErrorType = "nodeGroupDoesNotExistError"
)

const (
	// NodeRegistration signifies an error with node registering
	NodeRegistration AutoscalerErrorReason = "NodeRegistration"
)

// NewAutoscalerError returns new autoscaler error with a message constructed from format string
func NewAutoscalerError(errorType AutoscalerErrorType, msg string, args ...interface{}) AutoscalerError {
	return autoscalerErrorImpl{
		errorType: errorType,
		msg:       fmt.Sprintf(msg, args...),
	}
}

// NewAutoscalerErrorWithReason returns new autoscaler error with a reason and a message constructed from format string
func NewAutoscalerErrorWithReason(errorType AutoscalerErrorType, reason AutoscalerErrorReason, msg string, args ...interface{}) AutoscalerError {
	return autoscalerErrorImpl{
		errorType:   errorType,
		errorReason: reason,
		msg:         fmt.Sprintf(msg, args...),
	}
}

// NewAutoscalerCloudProviderError returns new autoscaler error with a cloudprovider error type and a message constructed from format string
func NewAutoscalerCloudProviderError(errorReason CloudProviderErrorReason, msg string, args ...interface{}) AutoscalerError {
	return autoscalerErrorImpl{
		errorType:   CloudProviderError,
		errorReason: AutoscalerErrorReason(errorReason),
		msg:         fmt.Sprintf(msg, args...),
	}
}

// ToAutoscalerError converts an error to AutoscalerError with given type,
// unless it already is an AutoscalerError (in which case it's not modified).
func ToAutoscalerError(defaultType AutoscalerErrorType, err error) AutoscalerError {
	if err == nil {
		return nil
	}
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

func (e autoscalerErrorImpl) Reason() AutoscalerErrorReason {
	return e.errorReason
}

// AddPrefix adds a prefix to error message.
// Returns the error it's called for convenient inline use.
// Example:
// if err := DoSomething(myObject); err != nil {
//
//	return err.AddPrefix("can't do something with %v: ", myObject)
//
// }
func (e autoscalerErrorImpl) AddPrefix(msg string, args ...interface{}) AutoscalerError {
	e.msg = fmt.Sprintf(msg, args...) + e.msg
	return e
}

// ServiceRawError wraps the RawError returned by the k8s/cloudprovider
// Azure clients. The error body  should satisfy the autorest.ServiceError type
type ServiceRawError struct {
	ServiceError *azure.ServiceError `json:"error,omitempty"`
}

func azureToAutoscalerError(rerr *retry.Error) AutoscalerError {
	if rerr == nil {
		return nil
	}
	if rerr.RawError == nil {
		return NewAutoscalerCloudProviderError(Unknown, rerr.Error().Error())
	}

	re := ServiceRawError{}
	err := json.Unmarshal([]byte(rerr.RawError.Error()), &re)
	if err != nil {
		return NewAutoscalerCloudProviderError(Unknown, rerr.Error().Error())
	}
	se := re.ServiceError
	if se == nil {
		return NewAutoscalerCloudProviderError(Unknown, rerr.Error().Error())
	}
	var errCode CloudProviderErrorReason
	if se.Code == "" {
		errCode = Unknown
	} else if se.Code == OperationNotAllowed {
		errCode = getOperationNotAllowedReason(se)
	} else {
		errCode = CloudProviderErrorReason(se.Code)
	}
	return NewAutoscalerCloudProviderError(errCode, se.Message)
}

// getOperationNotAllowedReason renames the error code for quotas to a more human-readable error
func getOperationNotAllowedReason(se *azure.ServiceError) CloudProviderErrorReason {
	if strings.Contains(se.Message, "Quota increase") {
		return QuotaExceeded
	}
	return CloudProviderErrorReason(OperationNotAllowed)
}
