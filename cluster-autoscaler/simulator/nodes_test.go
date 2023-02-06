/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/controller/daemon"
)

func TestBuildNodeInfoForNode(t *testing.T) {
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
			node: test.BuildTestNode("n", 1000, 10),
		},
		{
			name: "node with non-DS/mirror pods",
			node: test.BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				test.BuildScheduledTestPod("p1", 100, 1, "n"),
				test.BuildScheduledTestPod("p2", 100, 1, "n"),
			},
		},
		{
			name: "node with a mirror pod",
			node: test.BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				test.SetMirrorPodSpec(test.BuildScheduledTestPod("p1", 100, 1, "n")),
			},
			wantPods: []*apiv1.Pod{
				test.SetMirrorPodSpec(test.BuildScheduledTestPod("p1", 100, 1, "n")),
			},
		},
		{
			name: "node with a deleted mirror pod",
			node: test.BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				test.SetMirrorPodSpec(test.BuildScheduledTestPod("p1", 100, 1, "n")),
				setDeletionTimestamp(test.SetMirrorPodSpec(test.BuildScheduledTestPod("p2", 100, 1, "n"))),
			},
			wantPods: []*apiv1.Pod{
				test.SetMirrorPodSpec(test.BuildScheduledTestPod("p1", 100, 1, "n")),
			},
		},
		{
			name: "node with DS pods [forceDS=false, no daemon sets]",
			node: test.BuildTestNode("n", 1000, 10),
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
			node: test.BuildTestNode("n", 1000, 10),
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
			node: test.BuildTestNode("n", 1000, 10),
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
			node: test.BuildTestNode("n", 1000, 10),
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
			node: test.BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				test.BuildScheduledTestPod("p1", 100, 1, "n"),
				test.BuildScheduledTestPod("p2", 100, 1, "n"),
				test.SetMirrorPodSpec(test.BuildScheduledTestPod("p3", 100, 1, "n")),
				setDeletionTimestamp(test.SetMirrorPodSpec(test.BuildScheduledTestPod("p4", 100, 1, "n"))),
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: []*appsv1.DaemonSet{ds1, ds2, ds3},
			wantPods: []*apiv1.Pod{
				test.SetMirrorPodSpec(test.BuildScheduledTestPod("p3", 100, 1, "n")),
				buildDSPod(ds1, "n"),
			},
		},
		{
			name: "everything together [forceDS=true]",
			node: test.BuildTestNode("n", 1000, 10),
			pods: []*apiv1.Pod{
				test.BuildScheduledTestPod("p1", 100, 1, "n"),
				test.BuildScheduledTestPod("p2", 100, 1, "n"),
				test.SetMirrorPodSpec(test.BuildScheduledTestPod("p3", 100, 1, "n")),
				setDeletionTimestamp(test.SetMirrorPodSpec(test.BuildScheduledTestPod("p4", 100, 1, "n"))),
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: []*appsv1.DaemonSet{ds1, ds2, ds3},
			forceDS:    true,
			wantPods: []*apiv1.Pod{
				test.SetMirrorPodSpec(test.BuildScheduledTestPod("p3", 100, 1, "n")),
				buildDSPod(ds1, "n"),
				buildDSPod(ds2, "n"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nodeInfo, err := BuildNodeInfoForNode(tc.node, tc.pods, tc.daemonSets, tc.forceDS)

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, nodeInfo.Node(), tc.node)

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

func cleanPodMetadata(pod *apiv1.Pod) *apiv1.Pod {
	pod.Name = strings.Split(pod.Name, "-")[0]
	pod.OwnerReferences = nil
	return pod
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
