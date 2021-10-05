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

package egoscale

import "errors"

// ErrNotFound represents an error indicating a non-existent resource.
var ErrNotFound = errors.New("resource not found")

// ErrTooManyFound represents an error indicating multiple results found for a single resource.
var ErrTooManyFound = errors.New("multiple resources found")
