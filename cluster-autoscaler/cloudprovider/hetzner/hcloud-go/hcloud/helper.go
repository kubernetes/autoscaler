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

// String returns a pointer to the passed string s.
func String(s string) *string { return &s }

// Int returns a pointer to the passed integer i.
func Int(i int) *int { return &i }

// Bool returns a pointer to the passed bool b.
func Bool(b bool) *bool { return &b }

// Duration returns a pointer to the passed time.Duration d.
func Duration(d time.Duration) *time.Duration { return &d }

func intSlice(is []int) *[]int          { return &is }
func stringSlice(ss []string) *[]string { return &ss }
