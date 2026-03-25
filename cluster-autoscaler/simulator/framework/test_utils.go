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
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/informers"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	schedulerinterface "k8s.io/kube-scheduler/framework"
	scheduler_config_latest "k8s.io/kubernetes/pkg/scheduler/apis/config/latest"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

// NodeInfoCmpOptions returns a set of cmp.Options suitable for comparing NodeInfo and PodInfo objects.
func NodeInfoCmpOptions() []cmp.Option {
	return []cmp.Option{
		// Allow unexported fields in PodInfo and NodeInfo to enable full comparison.
		cmp.AllowUnexported(schedulerimpl.NodeInfo{}, NodeInfo{}, schedulerimpl.PodInfo{}, PodInfo{}),
		// Generation is expected to be different.
		cmpopts.IgnoreFields(schedulerimpl.NodeInfo{}, "Generation"),
		// Ignore cachedResource as it is lazily initialized and may differ between fresh and processed PodInfo objects.
		cmpopts.IgnoreFields(schedulerimpl.PodInfo{}, "cachedResource"),
		// The pod order changes in a particular way whenever schedulerimpl.RemovePod() is called. Instead of
		// relying on that schedulerimpl implementation detail in assertions, just ignore the order.
		cmp.FilterValues(func(x, y []schedulerinterface.PodInfo) bool { return true },
			cmpopts.SortSlices(func(p1, p2 schedulerinterface.PodInfo) bool {
				return p1.GetPod().Name < p2.GetPod().Name
			})),
		// Transform PodInfo to its embedded schedulerimpl.PodInfo for comparison when needed.
		cmp.Transformer("unwrapPodInfo", func(pi schedulerinterface.PodInfo) *schedulerimpl.PodInfo {
			if pi == nil {
				return nil
			}
			if internal, ok := pi.(*PodInfo); ok {
				return internal.PodInfo
			}
			if impl, ok := pi.(*schedulerimpl.PodInfo); ok {
				return impl
			}
			return nil
		}),
		// Sort our own *PodInfo slices as well.
		cmp.FilterValues(func(x, y []*PodInfo) bool { return true },
			cmpopts.SortSlices(func(p1, p2 *PodInfo) bool { return p1.Name < p2.Name })),
	}
}

// testFailer is an abstraction that covers both *testing.T and *testing.B.
type testFailer interface {
	Fatalf(format string, args ...any)
}

// NewTestNodeInfo returns a new NodeInfo without any DRA information - only to be used in test code.
// Production code should always take DRA objects into account.
func NewTestNodeInfo(node *apiv1.Node, pods ...*apiv1.Pod) *NodeInfo {
	nodeInfo := NewNodeInfo(node, nil)
	for _, pod := range pods {
		nodeInfo.AddPod(NewPodInfo(pod, nil))
	}
	return nodeInfo
}

// NewTestNodeInfoWithCSI returns a new NodeInfo object with CSINode information, but no DRA related
// information. It is meant to be used only from tests.
func NewTestNodeInfoWithCSI(node *apiv1.Node, csiNode *storagev1.CSINode, pods ...*apiv1.Pod) *NodeInfo {
	nodeInfo := NewNodeInfo(node, nil)
	for _, pod := range pods {
		nodeInfo.AddPod(NewPodInfo(pod, nil))
	}
	nodeInfo.CSINode = csiNode
	return nodeInfo
}

// NewTestFrameworkHandle creates a Handle that can be used in tests.
func NewTestFrameworkHandle() (*Handle, error) {
	defaultConfig, err := scheduler_config_latest.Default()
	if err != nil {
		return nil, err
	}
	fwHandle, err := NewHandle(context.Background(), informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(), 0), defaultConfig, true, true)
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
