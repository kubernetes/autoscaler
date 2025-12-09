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

package verda

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	t.Run("error with message", func(t *testing.T) {
		err := &APIError{
			StatusCode: 404,
			Message:    "Resource not found",
		}

		expected := "API error 404: Resource not found"
		if err.Error() != expected {
			t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("error with details", func(t *testing.T) {
		err := &APIError{
			StatusCode: 400,
			Message:    "Bad request",
			Details:    "Invalid parameter",
		}

		expected := "API error 400: Bad request (Invalid parameter)"
		if err.Error() != expected {
			t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
		}
	})
}

func TestValidationError_Error(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		err := &ValidationError{
			Field:   "email",
			Message: "Invalid email format",
		}

		expected := "validation error for field 'email': Invalid email format"
		if err.Error() != expected {
			t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
		}
	})
}
