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
	"k8s.io/client-go/informers"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	scheduler_config_latest "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"
)

// testFailer is an abstraction that covers both *testing.T and *testing.B.
type testFailer interface {
	Fatalf(format string, args ...any)
}

// NewTestNodeInfo returns a new NodeInfo without any DRA information - only to be used in test code.
// Production code should always take DRA objects into account.
func NewTestNodeInfo(node *apiv1.Node, pods ...*apiv1.Pod) *NodeInfo {
	nodeInfo := NewNodeInfo(node, nil)
	for _, pod := range pods {
		nodeInfo.AddPod(&PodInfo{Pod: pod, NeededResourceClaims: nil})
	}
	return nodeInfo
}

// NewTestFrameworkHandle creates a Handle that can be used in tests.
func NewTestFrameworkHandle() (*Handle, error) {
	defaultConfig, err := scheduler_config_latest.Default()
	if err != nil {
		return nil, err
	}
	fwHandle, err := NewHandle(informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(), 0), defaultConfig, true)
	if err != nil {
		return nil, err
	}
	return fwHandle, nil
}

// NewTestFrameworkHandleOrDie creates a Handle that can be used in tests.
func NewTestFrameworkHandleOrDie(t testFailer) *Handle {
	handle, err := NewTestFrameworkHandle()
	if err != nil {
		t.Fatalf("TestFrameworkHandleOrDie: couldn't create test framework handle: %v", err)
	}
	return handle
}
