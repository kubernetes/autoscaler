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
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
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
