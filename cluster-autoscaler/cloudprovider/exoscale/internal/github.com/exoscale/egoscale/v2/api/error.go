/*
Copyright 2021 The Kubernetes Authors.

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

package api

import "errors"

var (
	// ErrNotFound represents an error indicating a non-existent resource.
	ErrNotFound = errors.New("resource not found")

	// ErrTooManyFound represents an error indicating multiple results found for a single resource.
	ErrTooManyFound = errors.New("multiple resources found")

	// ErrInvalidRequest represents an error indicating that the caller's request is invalid.
	ErrInvalidRequest = errors.New("invalid request")

	// ErrAPIError represents an error indicating an API-side issue.
	ErrAPIError = errors.New("API error")
)
