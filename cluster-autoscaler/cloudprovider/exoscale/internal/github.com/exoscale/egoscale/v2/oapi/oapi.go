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

// Exoscale API OpenAPI specs, as well as helpers and transition types exposed
// in the public-facing package.
package oapi

import "context"

//go:generate oapi-codegen -generate types,client -package oapi -o oapi.gen.go source.json

type oapiClient interface {
	ClientWithResponsesInterface

	GetOperationWithResponse(context.Context, string, ...RequestEditorFn) (*GetOperationResponse, error)
}

// OptionalString returns the dereferenced string value of v if not nil, otherwise an empty string.
func OptionalString(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

// OptionalInt64 returns the dereferenced int64 value of v if not nil, otherwise 0.
func OptionalInt64(v *int64) int64 {
	if v != nil {
		return *v
	}

	return 0
}

// NilableString returns the input string pointer v if the dereferenced string is non-empty, otherwise nil.
// This helper is intended for use with OAPI types containing nilable string properties.
func NilableString(v *string) *string {
	if v != nil && *v == "" {
		return nil
	}

	return v
}
