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

package glogx

import (
	"k8s.io/klog"
)

type quota struct {
	limit int
	left  int
}

// NewLoggingQuota returns a quota object with limit & left set to the passed value.
func NewLoggingQuota(n int) *quota {
	return &quota{n, n}
}

// Left returns how much quota was left. If it was exceeded, the value will be negative.
func (q *quota) Left() int {
	return q.left
}

// Reset resets left quota to initial limit.
func (q *quota) Reset() {
	q.left = q.limit
}

// UpTo decreases quota for logging and reports whether there was any left.
// The returned value is a boolean of type glogx.Verbose.
func UpTo(quota *quota) klog.Verbose {
	quota.left--
	return quota.left >= 0
}

// Over reports whether quota for logging was exceeded.
// The returned value is a boolean of type glogx.Verbose.
func Over(quota *quota) klog.Verbose {
	return quota.left < 0
}

// V calls V from glog and wraps the result into glogx.Verbose.
func V(n klog.Level) Verbose {
	return Verbose(klog.V(n))
}

// Verbose is a wrapper for klog.Verbose that implements UpTo and Over.
type Verbose klog.Verbose

// UpTo calls UpTo from this package if called on true object.
// The returned value is a boolean of type klog.Verbose.
func (v Verbose) UpTo(quota *quota) klog.Verbose {
	if v {
		return UpTo(quota)
	}
	return klog.Verbose(false)
}

// Over calls Over from this package if called on true object.
// The returned value is a boolean of type klog.Verbose.
func (v Verbose) Over(quota *quota) klog.Verbose {
	if v {
		return Over(quota)
	}
	return klog.Verbose(false)
}
