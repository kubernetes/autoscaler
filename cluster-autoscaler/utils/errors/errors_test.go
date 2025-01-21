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

package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombine(t *testing.T) {
	cloudProviderErr := NewAutoscalerError(CloudProviderError, "provider error")
	internalErr := NewAutoscalerError(InternalError, "internal error")
	testCases := []struct {
		desc        string
		errors      []AutoscalerError
		expectedErr AutoscalerError
	}{
		{
			desc:        "no errors",
			errors:      []AutoscalerError{},
			expectedErr: nil,
		},
		{
			desc:        "single error",
			errors:      []AutoscalerError{internalErr},
			expectedErr: internalErr,
		},
		{
			desc: "two duplicated errors",
			errors: []AutoscalerError{
				internalErr,
				internalErr,
			},
			expectedErr: internalErr,
		},
		{
			desc: "two different errors",
			errors: []AutoscalerError{
				cloudProviderErr,
				internalErr,
			},
			expectedErr: NewAutoscalerError(
				CloudProviderError,
				"provider error ...and other errors: [\"[internalError] internal error\"]",
			),
		},
		{
			desc: "two different errors - reverse alphabetical order",
			errors: []AutoscalerError{
				internalErr,
				cloudProviderErr,
			},
			expectedErr: NewAutoscalerError(
				CloudProviderError,
				"provider error ...and other errors: [\"[internalError] internal error\"]",
			),
		},
		{
			desc: "errors with the same type and different messages",
			errors: []AutoscalerError{
				NewAutoscalerError(InternalError, "A"),
				NewAutoscalerError(InternalError, "B"),
				NewAutoscalerError(InternalError, "C"),
			},
			expectedErr: NewAutoscalerError(
				InternalError,
				"A ...and other errors: [\"B\", \"C\"]"),
		},
		{
			desc: "errors with the same type and some duplicated messages",
			errors: []AutoscalerError{
				NewAutoscalerError(InternalError, "A"),
				NewAutoscalerError(InternalError, "B"),
				NewAutoscalerError(InternalError, "A"),
			},
			expectedErr: NewAutoscalerError(
				InternalError,
				"A ...and other errors: [\"B\"]"),
		},
		{
			desc: "some duplicated errors",
			errors: []AutoscalerError{
				NewAutoscalerError(CloudProviderError, "A"),
				NewAutoscalerError(CloudProviderError, "A"),
				NewAutoscalerError(CloudProviderError, "B"),
				NewAutoscalerError(InternalError, "A"),
			},
			expectedErr: NewAutoscalerError(
				CloudProviderError,
				"A ...and other errors: [\"[cloudProviderError] B\", \"[internalError] A\"]"),
		},
		{
			desc: "different errors with quotes in messages",
			errors: []AutoscalerError{
				NewAutoscalerError(InternalError, "\"first\""),
				NewAutoscalerError(InternalError, "\"second\""),
			},
			expectedErr: NewAutoscalerError(
				InternalError,
				"\"first\" ...and other errors: [\"\\\"second\\\"\"]"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			combinedErr := Combine(testCase.errors)
			assert.Equal(t, testCase.expectedErr, combinedErr)
		})
	}
}
