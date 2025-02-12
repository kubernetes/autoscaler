/*
Copyright 2024 The Kubernetes Authors.

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

package clustersnapshot

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
)

// SchedulingErrorType represents different possible schedulingError types.
type SchedulingErrorType int

const (
	// SchedulingInternalError denotes internal unexpected error while trying to schedule a pod
	SchedulingInternalError SchedulingErrorType = iota
	// FailingPredicateError means that a pod couldn't be scheduled on a particular node because of a failing scheduler predicate
	FailingPredicateError
	// NoNodesPassingPredicatesFoundError means that a pod couldn't be scheduled on any Node because of failing scheduler predicates
	NoNodesPassingPredicatesFoundError
)

// SchedulingError represents an error encountered while trying to schedule a Pod inside ClusterSnapshot.
// An interface is exported instead of the concrete type to avoid the dreaded https://go.dev/doc/faq#nil_error.
type SchedulingError interface {
	error

	// Type can be used to distinguish between different SchedulingError types.
	Type() SchedulingErrorType
	// Reasons provides a list of human-readable reasons explaining the error.
	Reasons() []string

	// FailingPredicateName returns the name of the predicate that failed. Only applicable to the FailingPredicateError type.
	FailingPredicateName() string
	// FailingPredicateReasons returns a list of human-readable reasons explaining why the predicate failed. Only applicable to the FailingPredicateError type.
	FailingPredicateReasons() []string
}

type schedulingError struct {
	errorType SchedulingErrorType
	pod       *apiv1.Pod

	// Only applicable to SchedulingInternalError:
	internalErrorMsg string

	// Only applicable to FailingPredicateError:
	failingPredicateName             string
	failingPredicateReasons          []string
	failingPredicateUnexpectedErrMsg string
	// debugInfo contains additional info that predicate doesn't include,
	// but may be useful for debugging (e.g. taints on node blocking scale-up)
	failingPredicateDebugInfo string
}

// Type returns if error was internal of names predicate failure.
func (se *schedulingError) Type() SchedulingErrorType {
	return se.errorType
}

// Is returns whether this error is the same as another error by comparing the types.
func (se *schedulingError) Is(otherErr error) bool {
	otherSchedErr, ok := otherErr.(SchedulingError)
	if !ok {
		return false
	}
	return se.Type() == otherSchedErr.Type()
}

// Error satisfies the builtin error interface.
func (se *schedulingError) Error() string {
	msg := ""

	switch se.errorType {
	case SchedulingInternalError:
		msg = fmt.Sprintf("unexpected error: %s", se.internalErrorMsg)
	case FailingPredicateError:
		details := []string{
			fmt.Sprintf("predicateReasons=[%s]", strings.Join(se.FailingPredicateReasons(), ", ")),
		}
		if se.failingPredicateDebugInfo != "" {
			details = append(details, fmt.Sprintf("debugInfo=%s", se.failingPredicateDebugInfo))
		}
		if se.failingPredicateUnexpectedErrMsg != "" {
			details = append(details, fmt.Sprintf("unexpectedError=%s", se.failingPredicateUnexpectedErrMsg))
		}
		msg = fmt.Sprintf("predicate %q didn't pass (%s)", se.FailingPredicateName(), strings.Join(details, "; "))
	case NoNodesPassingPredicatesFoundError:
		msg = fmt.Sprintf("couldn't find a matching Node with passing predicates")
	default:
		msg = fmt.Sprintf("SchedulingErrorType type %q unknown - this shouldn't happen", se.errorType)
	}

	return fmt.Sprintf("can't schedule pod %s/%s: %s", se.pod.Namespace, se.pod.Name, msg)
}

// Reasons returns a list of human-readable reasons for the error.
func (se *schedulingError) Reasons() []string {
	switch se.errorType {
	case FailingPredicateError:
		return se.FailingPredicateReasons()
	default:
		return []string{se.Error()}
	}
}

// FailingPredicateName returns the name of the predicate which failed.
func (se *schedulingError) FailingPredicateName() string {
	return se.failingPredicateName
}

// FailingPredicateReasons returns the failure reasons from the failed predicate as a slice of strings.
func (se *schedulingError) FailingPredicateReasons() []string {
	return se.failingPredicateReasons
}

// NewSchedulingInternalError creates a new schedulingError with SchedulingInternalError type.
func NewSchedulingInternalError(pod *apiv1.Pod, errMsg string) SchedulingError {
	return &schedulingError{
		errorType:        SchedulingInternalError,
		pod:              pod,
		internalErrorMsg: errMsg,
	}
}

// NewFailingPredicateError creates a new schedulingError with FailingPredicateError type.
func NewFailingPredicateError(pod *apiv1.Pod, predicateName string, predicateReasons []string, unexpectedErrMsg string, debugInfo string) SchedulingError {
	return &schedulingError{
		errorType:                        FailingPredicateError,
		pod:                              pod,
		failingPredicateName:             predicateName,
		failingPredicateReasons:          predicateReasons,
		failingPredicateUnexpectedErrMsg: unexpectedErrMsg,
		failingPredicateDebugInfo:        debugInfo,
	}
}

// NewNoNodesPassingPredicatesFoundError creates a new schedulingError with NoNodesPassingPredicatesFoundError type.
func NewNoNodesPassingPredicatesFoundError(pod *apiv1.Pod) SchedulingError {
	return &schedulingError{
		errorType: NoNodesPassingPredicatesFoundError,
		pod:       pod,
	}
}
