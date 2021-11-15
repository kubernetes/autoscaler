/*
Copyright 2021 The Kubernetes Authors.

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

package nodegroupset

import (
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestIsClusterAPINodeInfoSimilar(t *testing.T) {
	comparator := CreateClusterAPINodeInfoComparator([]string{})
	node1 := BuildTestNode("node1", 1000, 2000)
	node2 := BuildTestNode("node2", 1000, 2000)

	for _, tc := range []struct {
		description    string
		label          string
		value1         string
		value2         string
		removeOneLabel bool
	}{
		{
			description:    "topology.ebs.csi.aws.com/zone empty value",
			label:          "topology.ebs.csi.aws.com/zone",
			value1:         "",
			value2:         "",
			removeOneLabel: false,
		},
		{
			description:    "topology.ebs.csi.aws.com/zone different values",
			label:          "topology.ebs.csi.aws.com/zone",
			value1:         "foo",
			value2:         "bar",
			removeOneLabel: false,
		},
		{
			description:    "topology.ebs.csi.aws.com/zone one node labeled",
			label:          "topology.ebs.csi.aws.com/zone",
			value1:         "foo",
			value2:         "bar",
			removeOneLabel: true,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			node1.ObjectMeta.Labels[tc.label] = tc.value1
			node2.ObjectMeta.Labels[tc.label] = tc.value2
			if tc.removeOneLabel {
				delete(node2.ObjectMeta.Labels, tc.label)
			}
			checkNodesSimilar(t, node1, node2, comparator, true)
		})
	}
}

func TestFindSimilarNodeGroupsClusterAPIBasic(t *testing.T) {
	context := &context.AutoscalingContext{}
	ni1, ni2, ni3 := buildBasicNodeGroups(context)
	processor := &BalancingNodeGroupSetProcessor{Comparator: CreateClusterAPINodeInfoComparator([]string{})}
	basicSimilarNodeGroupsTest(t, context, processor, ni1, ni2, ni3)
}
