/*
Copyright 2020 The Kubernetes Authors.

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

package sdk

import (
	"fmt"
)

// APIError represents an error that can occurred while calling the API.
type APIError struct {
	// Error message.
	Message string
	// HTTP code.
	Code int
	// ID of the request
	QueryID string
}

func (err *APIError) Error() string {
	return fmt.Sprintf("Error %d: %q", err.Code, err.Message)
}

type (
	// Error struct
	Error struct {
		StatusCode int
		Method     string
		Path       string
		Type       string
		Message    string
	}
)

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}
