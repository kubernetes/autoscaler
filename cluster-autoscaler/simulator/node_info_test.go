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

package simulator

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/controller/daemon"
)

func TestSanitizePod(t *testing.T) {
	pod := BuildTestPod("p1", 80, 0)
	pod.Spec.NodeName = "n1"

	node := BuildTestNode("node", 1000, 1000)

	resNode := sanitizeNode(node, "test-group", nil)
	res := sanitizePod(pod, resNode.Name, "abc")
	assert.Equal(t, res.Spec.NodeName, resNode.Name)
	assert.Equal(t, res.Name, "p1-abc")
}

func TestSanitizeLabels(t *testing.T) {
	oldNode := BuildTestNode("ng1-1", 1000, 1000)
	oldNode.Labels = map[string]string{
		apiv1.LabelHostname: "abc",
		"x":                 "y",
	}
	node := sanitizeNode(oldNode, "bzium", nil)
	assert.NotEqual(t, node.Labels[apiv1.LabelHostname], "abc", nil)
	assert.Equal(t, node.Labels["x"], "y")
	assert.NotEqual(t, node.Name, oldNode.Name)
	assert.Equal(t, node.Labels[apiv1.LabelHostname], node.Name)
}

func TestTemplateNodeInfoFromExampleNodeInfo(t *testing.T) {
	ds1 := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds1",
			Namespace: "ds1-namespace",
			UID:       types.UID("ds1"),
		},
	}

	ds2 := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds2",
			Namespace: "ds2-namespace",
			UID:       types.UID("ds2"),
		},
	}

	ds3 := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds3",
			Namespace: "ds3-namespace",
			UID:       types.UID("ds3"),
		},
		Spec: appsv1.DaemonSetSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					NodeSelector: map[string]string{"key": "value"},
				},
			},
		},
	}

	testCases := []struct {
		name       string
		node       *apiv1.Node
		pods       []*apiv1.Pod
		daemonSets []*appsv1.DaemonSet
		forceDS    bool

		wantPods  []*apiv1.Pod
		wantError bool
	}{
		{
			name: "node without any pods",
			node: BuildTestNode("n", 1000, 10),
		},
		{
			name: "node with non-DS/mirror pods",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 100, 1, "n"),
			},
		},
		{
			name: "node with a mirror pod",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p1", 100, 1, "n")),
			},
			wantPods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p1", 100, 1, "n")),
			},
		},
		{
			name: "node with a deleted mirror pod",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p1", 100, 1, "n")),
				setDeletionTimestamp(SetMirrorPodSpec(BuildScheduledTestPod("p2", 100, 1, "n"))),
			},
			wantPods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p1", 100, 1, "n")),
			},
		},
		{
			name: "node with DS pods [forceDS=false, no daemon sets]",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			wantPods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
			},
		},
		{
			name: "node with DS pods [forceDS=false, some daemon sets]",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: []*appsv1.DaemonSet{ds1, ds2, ds3},
			wantPods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
			},
		},
		{
			name: "node with a DS pod [forceDS=true, no daemon sets]",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			wantPods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
			},
			forceDS: true,
		},
		{
			name: "node with a DS pod [forceDS=true, some daemon sets]",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: []*appsv1.DaemonSet{ds1, ds2, ds3},
			forceDS:    true,
			wantPods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				buildDSPod(ds2, "n"),
			},
		},
		{
			name: "everything together [forceDS=false]",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 100, 1, "n"),
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				setDeletionTimestamp(SetMirrorPodSpec(BuildScheduledTestPod("p4", 100, 1, "n"))),
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: []*appsv1.DaemonSet{ds1, ds2, ds3},
			wantPods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				buildDSPod(ds1, "n"),
			},
		},
		{
			name: "everything together [forceDS=true]",
			node: BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 100, 1, "n"),
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				setDeletionTimestamp(SetMirrorPodSpec(BuildScheduledTestPod("p4", 100, 1, "n"))),
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: []*appsv1.DaemonSet{ds1, ds2, ds3},
			forceDS:    true,
			wantPods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				buildDSPod(ds1, "n"),
				buildDSPod(ds2, "n"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exampleNodeInfo := framework.NewNodeInfo(tc.node, nil)
			for _, pod := range tc.pods {
				exampleNodeInfo.AddPod(&framework.PodInfo{Pod: pod})
			}
			nodeInfo, err := TemplateNodeInfoFromExampleNodeInfo(exampleNodeInfo, "nodeGroupId", tc.daemonSets, tc.forceDS, taints.TaintConfig{})

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, cleanNodeMetadata(nodeInfo.Node()), cleanNodeMetadata(tc.node))

				// clean pod metadata for comparison purposes
				var wantPods, pods []*apiv1.Pod
				for _, pod := range tc.wantPods {
					wantPods = append(wantPods, cleanPodMetadata(pod))
				}
				for _, podInfo := range nodeInfo.Pods {
					pods = append(pods, cleanPodMetadata(podInfo.Pod))
				}
				assert.ElementsMatch(t, tc.wantPods, pods)
			}
		})
	}
}

func TestTemplateNodeInfoFromNodeGroupTemplate(t *testing.T) {
	// TODO(DRA): Write.
}

func TestFreshNodeInfoFromTemplateNodeInfo(t *testing.T) {
	// TODO(DRA): Write.
}

func TestDeepCopyNodeInfo(t *testing.T) {
	// TODO(DRA): Write.
}

func cleanPodMetadata(pod *apiv1.Pod) *apiv1.Pod {
	pod.Name = strings.Split(pod.Name, "-")[0]
	pod.UID = ""
	pod.OwnerReferences = nil
	pod.Spec.NodeName = ""
	return pod
}

func cleanNodeMetadata(node *apiv1.Node) *apiv1.Node {
	node.UID = ""
	node.Name = ""
	return node
}

func buildDSPod(ds *appsv1.DaemonSet, nodeName string) *apiv1.Pod {
	pod := daemon.NewPod(ds, nodeName)
	pod.Name = ds.Name
	ptrVal := true
	pod.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{Kind: "DaemonSet", UID: ds.UID, Controller: &ptrVal},
	}
	return pod
}

func setDeletionTimestamp(pod *apiv1.Pod) *apiv1.Pod {
	now := metav1.NewTime(time.Now())
	pod.DeletionTimestamp = &now
	return pod
}
