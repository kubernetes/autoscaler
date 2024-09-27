/*
Copyright 2024 The Kubernetes Authors.

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

package framework

import (
	apiv1 "k8s.io/api/core/v1"
)

// NewTestNodeInfo returns a new NodeInfo without any DRA information - only to be used in test code.
// Production code should always take DRA objects into account.
func NewTestNodeInfo(node *apiv1.Node, pods ...*apiv1.Pod) *NodeInfo {
	nodeInfo := NewNodeInfo(node, nil)
	for _, pod := range pods {
		nodeInfo.AddPod(&PodInfo{Pod: pod, NeededResourceClaims: nil})
	}
	return nodeInfo
}
