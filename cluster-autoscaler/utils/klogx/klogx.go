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

package klogx

import (
	klog "k8s.io/klog/v2"
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

// V calls V from glog and wraps the result into glogx.Verbose.
func V(n klog.Level) Verbose {
	return Verbose{
		enabled: true,
		v:       klog.V(n)}
}

// Verbose is a wrapper for klog.Verbose that implements UpTo and Over.
// It provides a subset of methods exposed by klog.Verbose.
type Verbose struct {
	enabled bool
	v       klog.Verbose
}

func (v Verbose) enable(b bool) Verbose {
	return Verbose{
		enabled: b,
		v:       v.v}
}

// UpTo calls UpTo from this package if called on true object.
// The returned value is of type Verbose.
func (v Verbose) UpTo(quota *quota) Verbose {
	if v.v.Enabled() {
		quota.left--
		return v.enable(quota.left >= 0)
	}
	return v.enable(false)
}

// Over calls Over from this package if called on true object.
// The returned value is of type Verbose.
func (v Verbose) Over(quota *quota) Verbose {
	if v.v.Enabled() {
		return v.enable(quota.left < 0)
	}
	return v.enable(false)
}

// Infof is a wrapper for klog.Infof that logs if the quota
// allows for it.
func (v Verbose) Infof(format string, args ...interface{}) {
	if v.enabled {
		v.v.Infof(format, args...)
	}
}

// Info is a wrapper for klog.Info that logs if the quota
// allows for it.
func (v Verbose) Info(args ...interface{}) {
	if v.enabled {
		v.v.Info(args...)
	}
}
