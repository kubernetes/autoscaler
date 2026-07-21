/*
Copyright The Kubernetes Authors.

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

package efficiency

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

// cloudProviderMockOptions overrides constraints enforced by the cloud provider infrastructure.
type cloudProviderMockOptions struct {
	NodeToNodeGroup   map[string]string
	NodeGroupMinSizes map[string]int
}

// nodeGroup represents a named node group with Nodes and Pods.
type nodeGroup struct {
	Name    string
	MinSize int
	MaxSize int
	Nodes   []*apiv1.Node
	Pods    []*apiv1.Pod

	// Overrides for optional nodeGroup to customize how autoscaling of a given nodeGroup works.
	NodeGroupOptions *config.NodeGroupAutoscalingOptions
}

type initialClusterState struct {
	Name                 string
	NodeGroups           []*nodeGroup
	CloudProviderOptions *cloudProviderMockOptions
}

// clusterStateFactory generates the initial cluster topology before a scaledown loop begins.
// Compared to static tables, factory state generation allows for dynamic initialization of more complex initial cluster states.
type clusterStateFactory func() *initialClusterState

func differentNodeSizesMemIrrelevant() clusterStateFactory {
	return func() *initialClusterState {
		return &initialClusterState{
			Name: "different node sizes, memory irrelevant",
			NodeGroups: []*nodeGroup{
				{
					Name:    "ng-small",
					MinSize: 0,
					MaxSize: 10,
					Nodes: []*apiv1.Node{
						testutils.BuildTestNode("small-node-1", 1000, 1000, testutils.IsReady(true), testutils.WithNodeLabels(map[string]string{apiv1.LabelInstanceTypeStable: "s-class"})),
						testutils.BuildTestNode("small-node-2", 1000, 1000, testutils.IsReady(true), testutils.WithNodeLabels(map[string]string{apiv1.LabelInstanceTypeStable: "s-class"})),
					},
					Pods: []*apiv1.Pod{
						testutils.SetRSPodSpec(testutils.BuildScheduledTestPod("pod-s1", 1000, 100, "small-node-1"), "rs"),
						testutils.SetRSPodSpec(testutils.BuildScheduledTestPod("pod-s2", 1000, 100, "small-node-2"), "rs"),
					},
				},
				{
					Name:    "ng-moderate",
					MinSize: 0,
					MaxSize: 10,
					Nodes: []*apiv1.Node{
						testutils.BuildTestNode("mod-node-1", 4000, 1000, testutils.IsReady(true), testutils.WithNodeLabels(map[string]string{apiv1.LabelInstanceTypeStable: "m-class"})),
						testutils.BuildTestNode("mod-node-2", 4000, 1000, testutils.IsReady(true), testutils.WithNodeLabels(map[string]string{apiv1.LabelInstanceTypeStable: "m-class"})),
						testutils.BuildTestNode("mod-node-3", 4000, 1000, testutils.IsReady(true), testutils.WithNodeLabels(map[string]string{apiv1.LabelInstanceTypeStable: "m-class"})),
					},
					Pods: []*apiv1.Pod{
						testutils.SetRSPodSpec(testutils.BuildScheduledTestPod("pod-m1", 1000, 100, "mod-node-1"), "rs"),
						testutils.SetRSPodSpec(testutils.BuildScheduledTestPod("pod-m2", 1000, 100, "mod-node-2"), "rs"),
						testutils.SetRSPodSpec(testutils.BuildScheduledTestPod("pod-m3", 1000, 100, "mod-node-3"), "rs"),
					},
				},
				{
					Name:    "ng-big",
					MinSize: 0,
					MaxSize: 10,
					Nodes: []*apiv1.Node{
						testutils.BuildTestNode("big-node-1", 9000, 1000, testutils.IsReady(true), testutils.WithNodeLabels(map[string]string{apiv1.LabelInstanceTypeStable: "xl-class"})),
					},
					Pods: []*apiv1.Pod{
						testutils.SetRSPodSpec(testutils.BuildScheduledTestPod("pod-b1", 1000, 100, "big-node-1"), "rs"),
					},
				},
			},
		}
	}
}
