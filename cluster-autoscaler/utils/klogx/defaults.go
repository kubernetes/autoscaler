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

import klog "k8s.io/klog/v2"

const (
	// MaxPodsLogged is the maximum number of pods for which we will
	// log detailed information every loop at verbosity < 5.
	MaxPodsLogged = 20
	// MaxPodsLoggedV5 is the maximum number of pods for which we will
	// log detailed information every loop at verbosity >= 5.
	MaxPodsLoggedV5 = 1000
)

// PodsLoggingQuota returns a new quota with default limit for pods at current verbosity.
func PodsLoggingQuota() *quota {
	if klog.V(5).Enabled() {
		return NewLoggingQuota(MaxPodsLoggedV5)
	}
	return NewLoggingQuota(MaxPodsLogged)
}
