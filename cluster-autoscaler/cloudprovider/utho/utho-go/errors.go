/*
Copyright 2025 The Kubernetes Authors.

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

package utho

import (
	"fmt"
	"net/http"
)

// ErrorResponse represents an error response from the Utho API.
type ErrorResponse struct {
	Response *http.Response
	Errors   []Error `json:"errors"`
}

// Error represents a single error returned by the Utho API.
type Error struct {
	Message     string      `json:"message"`
	LongMessage string      `json:"long_message"`
	Code        string      `json:"code"`
	Meta        interface{} `json:"meta,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %+v",
		e.Response.Request.Method, e.Response.Request.URL,
		e.Response.StatusCode, e.Errors)
}
