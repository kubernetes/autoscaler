/*
Copyright 2018 The Kubernetes Authors.

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

package hcloud

import "time"

// Ptr returns a pointer to p.
func Ptr[T any](p T) *T {
	return &p
}

// String returns a pointer to the passed string s.
//
// Deprecated: Use [Ptr] instead.
func String(s string) *string { return Ptr(s) }

// Int returns a pointer to the passed integer i.
//
// Deprecated: Use [Ptr] instead.
func Int(i int) *int { return Ptr(i) }

// Bool returns a pointer to the passed bool b.
//
// Deprecated: Use [Ptr] instead.
func Bool(b bool) *bool { return Ptr(b) }

// Duration returns a pointer to the passed time.Duration d.
//
// Deprecated: Use [Ptr] instead.
func Duration(d time.Duration) *time.Duration { return Ptr(d) }
