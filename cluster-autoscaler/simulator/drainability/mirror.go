/*
Copyright 2023 The Kubernetes Authors.

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

package drainability

import (
	"k8s.io/autoscaler/cluster-autoscaler/utils/pod"

	apiv1 "k8s.io/api/core/v1"
)

// MirrorPodRule is a drainability rule on how to handle mirror pods.
type MirrorPodRule struct{}

// NewMirrorPodRule creates a new MirrorPodRule.
func NewMirrorPodRule() *MirrorPodRule {
	return &MirrorPodRule{}
}

// Drainable decides what to do with mirror pods on node drain.
func (m *MirrorPodRule) Drainable(p *apiv1.Pod) Status {
	if pod.IsMirrorPod(p) {
		return NewSkipStatus()
	}
	return NewUndefinedStatus()
}
