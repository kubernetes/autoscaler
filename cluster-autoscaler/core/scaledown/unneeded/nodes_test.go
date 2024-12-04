/*
Copyright 2022 The Kubernetes Authors.

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

package unneeded

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/resource"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodes"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdate(t *testing.T) {
	initialTimestamp := time.Now()
	finalTimestamp := initialTimestamp.Add(1 * time.Minute)
	testCases := []struct {
		desc           string
		initialNodes   []simulator.NodeToBeRemoved
		finalNodes     []simulator.NodeToBeRemoved
		wantTimestamps map[string]time.Time
		wantVersions   map[string]string
	}{
		{
			desc: "added then deleted",
			initialNodes: []simulator.NodeToBeRemoved{
				makeNode("n1", "v1"),
				makeNode("n2", "v1"),
				makeNode("n3", "v1"),
			},
			finalNodes: []simulator.NodeToBeRemoved{},
		},
		{
			desc:         "added in last call",
			initialNodes: []simulator.NodeToBeRemoved{},
			finalNodes: []simulator.NodeToBeRemoved{
				makeNode("n1", "v1"),
				makeNode("n2", "v1"),
				makeNode("n3", "v1"),
			},
			wantTimestamps: map[string]time.Time{"n1": finalTimestamp, "n2": finalTimestamp, "n3": finalTimestamp},
			wantVersions:   map[string]string{"n1": "v1", "n2": "v1", "n3": "v1"},
		},
		{
			desc: "single one remaining",
			initialNodes: []simulator.NodeToBeRemoved{
				makeNode("n1", "v1"),
				makeNode("n2", "v1"),
				makeNode("n3", "v1"),
			},
			finalNodes: []simulator.NodeToBeRemoved{
				makeNode("n2", "v2"),
			},
			wantTimestamps: map[string]time.Time{"n2": initialTimestamp},
			wantVersions:   map[string]string{"n2": "v2"},
		},
		{
			desc: "single one older",
			initialNodes: []simulator.NodeToBeRemoved{
				makeNode("n2", "v1"),
			},
			finalNodes: []simulator.NodeToBeRemoved{
				makeNode("n1", "v2"),
				makeNode("n2", "v2"),
				makeNode("n3", "v2"),
			},
			wantTimestamps: map[string]time.Time{"n1": finalTimestamp, "n2": initialTimestamp, "n3": finalTimestamp},
			wantVersions:   map[string]string{"n1": "v2", "n2": "v2", "n3": "v2"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			nodes := NewNodes(nil, nil)
			nodes.Update(tc.initialNodes, initialTimestamp)
			nodes.Update(tc.finalNodes, finalTimestamp)
			wantNodes := len(tc.wantTimestamps)
			assert.Equal(t, wantNodes, len(nodes.AsList()))
			assert.Equal(t, wantNodes, len(nodes.byName))
			for _, n := range nodes.AsList() {
				nn, found := nodes.byName[n.Name]
				assert.True(t, found)
				assert.Equal(t, tc.wantTimestamps[n.Name], nn.since)
				assert.Equal(t, tc.wantVersions[n.Name], version(nn.ntbr))
			}
		})
	}
}

const testVersion = "testVersion"

func makeNode(name, version string) simulator.NodeToBeRemoved {
	n := BuildTestNode(name, 1000, 10)
	n.Annotations = map[string]string{testVersion: version}
	return simulator.NodeToBeRemoved{Node: n}
}

func version(n simulator.NodeToBeRemoved) string {
	return n.Node.Annotations[testVersion]
}

func TestRemovableAt(t *testing.T) {
	testCases := []struct {
		name                string
		numEmpty            int
		numDrain            int
		minSize             int
		targetSize          int
		numOngoingDeletions int
		numEmptyToRemove    int
		numDrainToRemove    int
	}{
		{
			name:                "Node group min size is not reached",
			numEmpty:            3,
			numDrain:            2,
			minSize:             1,
			targetSize:          10,
			numOngoingDeletions: 2,
			numEmptyToRemove:    3,
			numDrainToRemove:    2,
		},
		{
			name:                "Node group min size is reached for drain nodes",
			numEmpty:            3,
			numDrain:            5,
			minSize:             1,
			targetSize:          10,
			numOngoingDeletions: 2,
			numEmptyToRemove:    3,
			numDrainToRemove:    4,
		},
		{
			name:                "Node group min size is reached for empty and drain nodes",
			numEmpty:            3,
			numDrain:            5,
			minSize:             1,
			targetSize:          5,
			numOngoingDeletions: 2,
			numEmptyToRemove:    2,
			numDrainToRemove:    0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ng := testprovider.NewTestNodeGroup("ng", 100, tc.minSize, tc.targetSize, true, false, "", nil, nil)
			empty := []simulator.NodeToBeRemoved{}
			for i := 0; i < tc.numEmpty; i++ {
				empty = append(empty, simulator.NodeToBeRemoved{
					Node: BuildTestNode(fmt.Sprintf("empty-%d", i), 10, 100),
				})
			}
			drain := []simulator.NodeToBeRemoved{}
			for i := 0; i < tc.numDrain; i++ {
				drain = append(drain, simulator.NodeToBeRemoved{
					Node:             BuildTestNode(fmt.Sprintf("drain-%d", i), 10, 100),
					PodsToReschedule: []*apiv1.Pod{BuildTestPod(fmt.Sprintf("pod-%d", i), 1, 1)},
				})
			}

			removableNodes := append(empty, drain...)
			provider := testprovider.NewTestCloudProvider(nil, nil)
			provider.InsertNodeGroup(ng)
			for _, node := range removableNodes {
				provider.AddNode("ng", node.Node)
			}

			as := &fakeActuationStatus{deletionCount: map[string]int{"ng": tc.numOngoingDeletions}}

			rsLister, err := kube_util.NewTestReplicaSetLister(nil)
			assert.NoError(t, err)
			registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, rsLister, nil)
			ctx, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{ScaleDownSimulationTimeout: 5 * time.Minute}, &fake.Clientset{}, registry, provider, nil, nil)
			assert.NoError(t, err)

			n := NewNodes(&fakeScaleDownTimeGetter{}, &resource.LimitsFinder{})
			n.Update(removableNodes, time.Now())
			gotEmptyToRemove, gotDrainToRemove, _ := n.RemovableAt(&ctx, nodes.ScaleDownContext{
				ActuationStatus:     as,
				ResourcesLeft:       resource.Limits{},
				ResourcesWithLimits: []string{},
			}, time.Now())
			if len(gotDrainToRemove) != tc.numDrainToRemove || len(gotEmptyToRemove) != tc.numEmptyToRemove {
				t.Errorf("%s: getNodesToRemove() return %d, %d, want %d, %d", tc.name, len(gotEmptyToRemove), len(gotDrainToRemove), tc.numEmptyToRemove, tc.numDrainToRemove)
			}
		})
	}
}

type fakeActuationStatus struct {
	recentEvictions []*apiv1.Pod
	deletionCount   map[string]int
}

func (f *fakeActuationStatus) RecentEvictions() []*apiv1.Pod {
	return f.recentEvictions
}

func (f *fakeActuationStatus) DeletionsInProgress() ([]string, []string) {
	return nil, nil
}

func (f *fakeActuationStatus) DeletionResults() (map[string]status.NodeDeleteResult, time.Time) {
	return nil, time.Time{}
}

func (f *fakeActuationStatus) DeletionsCount(nodeGroup string) int {
	return f.deletionCount[nodeGroup]
}

type fakeScaleDownTimeGetter struct{}

func (f *fakeScaleDownTimeGetter) GetScaleDownUnneededTime(cloudprovider.NodeGroup) (time.Duration, error) {
	return 0 * time.Second, nil
}

func (f *fakeScaleDownTimeGetter) GetScaleDownUnreadyTime(cloudprovider.NodeGroup) (time.Duration, error) {
	return 0 * time.Second, nil
}
