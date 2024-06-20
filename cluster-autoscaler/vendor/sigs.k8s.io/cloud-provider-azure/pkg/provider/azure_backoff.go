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

package provider

import (
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	pipErrorMessageRE = regexp.MustCompile(`(?:.*)/subscriptions/(?:.*)/resourceGroups/(.*)/providers/Microsoft.Network/publicIPAddresses/([^\s]+)(?:.*)`)
)

// RequestBackoff if backoff is disabled in cloud provider it
// returns a new Backoff object steps = 1
// This is to make sure that the requested command executes
// at least once
func (az *Cloud) RequestBackoff() (resourceRequestBackoff wait.Backoff) {
	if az.CloudProviderBackoff {
		return az.ResourceRequestBackoff
	}
	resourceRequestBackoff = wait.Backoff{
		Steps: 1,
	}
	return resourceRequestBackoff
}

// Event creates a event for the specified object.
func (az *Cloud) Event(obj runtime.Object, eventType, reason, message string) {
	if obj != nil && reason != "" {
		az.eventRecorder.Event(obj, eventType, reason, message)
	}
}
