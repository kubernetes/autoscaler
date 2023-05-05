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

package orchestrator

import (
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	"github.com/stretchr/testify/assert"
)

func TestCombinedConcurrentScaleUpErrors(t *testing.T) {
	cloudProviderErr := errors.NewAutoscalerError(errors.CloudProviderError, "provider error")
	internalErr := errors.NewAutoscalerError(errors.InternalError, "internal error")
	testCases := []struct {
		desc        string
		errors      []errors.AutoscalerError
		expectedErr errors.AutoscalerError
	}{
		{
			desc:        "no errors",
			errors:      []errors.AutoscalerError{},
			expectedErr: nil,
		},
		{
			desc:        "single error",
			errors:      []errors.AutoscalerError{internalErr},
			expectedErr: internalErr,
		},
		{
			desc: "two duplicated errors",
			errors: []errors.AutoscalerError{
				internalErr,
				internalErr,
			},
			expectedErr: internalErr,
		},
		{
			desc: "two different errors",
			errors: []errors.AutoscalerError{
				cloudProviderErr,
				internalErr,
			},
			expectedErr: errors.NewAutoscalerError(
				errors.CloudProviderError,
				"provider error ...and other concurrent errors: [\"[internalError] internal error\"]",
			),
		},
		{
			desc: "two different errors - reverse alphabetical order",
			errors: []errors.AutoscalerError{
				internalErr,
				cloudProviderErr,
			},
			expectedErr: errors.NewAutoscalerError(
				errors.CloudProviderError,
				"provider error ...and other concurrent errors: [\"[internalError] internal error\"]",
			),
		},
		{
			desc: "errors with the same type and different messages",
			errors: []errors.AutoscalerError{
				errors.NewAutoscalerError(errors.InternalError, "A"),
				errors.NewAutoscalerError(errors.InternalError, "B"),
				errors.NewAutoscalerError(errors.InternalError, "C"),
			},
			expectedErr: errors.NewAutoscalerError(
				errors.InternalError,
				"A ...and other concurrent errors: [\"B\", \"C\"]"),
		},
		{
			desc: "errors with the same type and some duplicated messages",
			errors: []errors.AutoscalerError{
				errors.NewAutoscalerError(errors.InternalError, "A"),
				errors.NewAutoscalerError(errors.InternalError, "B"),
				errors.NewAutoscalerError(errors.InternalError, "A"),
			},
			expectedErr: errors.NewAutoscalerError(
				errors.InternalError,
				"A ...and other concurrent errors: [\"B\"]"),
		},
		{
			desc: "some duplicated errors",
			errors: []errors.AutoscalerError{
				errors.NewAutoscalerError(errors.CloudProviderError, "A"),
				errors.NewAutoscalerError(errors.CloudProviderError, "A"),
				errors.NewAutoscalerError(errors.CloudProviderError, "B"),
				errors.NewAutoscalerError(errors.InternalError, "A"),
			},
			expectedErr: errors.NewAutoscalerError(
				errors.CloudProviderError,
				"A ...and other concurrent errors: [\"[cloudProviderError] B\", \"[internalError] A\"]"),
		},
		{
			desc: "different errors with quotes in messages",
			errors: []errors.AutoscalerError{
				errors.NewAutoscalerError(errors.InternalError, "\"first\""),
				errors.NewAutoscalerError(errors.InternalError, "\"second\""),
			},
			expectedErr: errors.NewAutoscalerError(
				errors.InternalError,
				"\"first\" ...and other concurrent errors: [\"\\\"second\\\"\"]"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			combinedErr := combineConcurrentScaleUpErrors(testCase.errors)
			assert.Equal(t, testCase.expectedErr, combinedErr)
		})
	}
}
