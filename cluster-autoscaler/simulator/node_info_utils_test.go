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
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/controller/daemon"
)

var (
	ds1 = &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds1",
			Namespace: "ds1-namespace",
			UID:       types.UID("ds1"),
		},
	}
	ds2 = &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds2",
			Namespace: "ds2-namespace",
			UID:       types.UID("ds2"),
		},
	}
	ds3 = &appsv1.DaemonSet{
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
	testDaemonSets = []*appsv1.DaemonSet{ds1, ds2, ds3}
)

func TestSanitizedTemplateNodeInfoFromNodeGroup(t *testing.T) {
	exampleNode := BuildTestNode("n", 1000, 10)
	exampleNode.Spec.Taints = []apiv1.Taint{
		{Key: taints.ToBeDeletedTaint, Value: "2312532423", Effect: apiv1.TaintEffectNoSchedule},
	}

	for _, tc := range []struct {
		testName  string
		nodeGroup *fakeNodeGroup

		wantPods    []*apiv1.Pod
		wantCpError bool
	}{
		{
			testName:    "node group error results in an error",
			nodeGroup:   &fakeNodeGroup{templateNodeInfoErr: fmt.Errorf("test error")},
			wantCpError: true,
		},
		{
			testName: "simple template with no pods",
			nodeGroup: &fakeNodeGroup{
				templateNodeInfoResult: framework.NewNodeInfo(exampleNode, nil),
			},
			wantPods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				buildDSPod(ds2, "n"),
			},
		},
		{
			testName: "template with all kinds of pods",
			nodeGroup: &fakeNodeGroup{
				templateNodeInfoResult: framework.NewNodeInfo(exampleNode, nil,
					&framework.PodInfo{Pod: BuildScheduledTestPod("p1", 100, 1, "n")},
					&framework.PodInfo{Pod: BuildScheduledTestPod("p2", 100, 1, "n")},
					&framework.PodInfo{Pod: SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n"))},
					&framework.PodInfo{Pod: setDeletionTimestamp(SetMirrorPodSpec(BuildScheduledTestPod("p4", 100, 1, "n")))},
					&framework.PodInfo{Pod: buildDSPod(ds1, "n")},
					&framework.PodInfo{Pod: setDeletionTimestamp(buildDSPod(ds2, "n"))},
				),
			},
			wantPods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				buildDSPod(ds1, "n"),
				buildDSPod(ds2, "n"),
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			templateNodeInfo, err := SanitizedTemplateNodeInfoFromNodeGroup(tc.nodeGroup, testDaemonSets, taints.TaintConfig{})
			if tc.wantCpError {
				if err == nil || err.Type() != errors.CloudProviderError {
					t.Fatalf("TemplateNodeInfoFromNodeGroupTemplate(): want CloudProviderError, but got: %v (%T)", err, err)
				} else {
					return
				}
			}
			if err != nil {
				t.Fatalf("TemplateNodeInfoFromNodeGroupTemplate(): expected no error, but got %v", err)
			}

			// Verify that the taints are correctly sanitized.
			// Verify that the NodeInfo is sanitized using the node group id as base.
			// Pass empty string as nameSuffix so that it's auto-determined from the sanitized templateNodeInfo, because
			// TemplateNodeInfoFromNodeGroupTemplate randomizes the suffix.
			// Pass non-empty expectedPods to verify that the set of pods is changed as expected (e.g. DS pods added, non-DS/deleted pods removed).
			if err := verifyNodeInfoSanitization(tc.nodeGroup.templateNodeInfoResult, templateNodeInfo, tc.wantPods, "template-node-for-"+tc.nodeGroup.id, "", nil); err != nil {
				t.Fatalf("TemplateNodeInfoFromExampleNodeInfo(): NodeInfo wasn't properly sanitized: %v", err)
			}
		})
	}
}

func TestSanitizedTemplateNodeInfoFromNodeInfo(t *testing.T) {
	exampleNode := BuildTestNode("n", 1000, 10)
	exampleNode.Spec.Taints = []apiv1.Taint{
		{Key: taints.ToBeDeletedTaint, Value: "2312532423", Effect: apiv1.TaintEffectNoSchedule},
	}

	testCases := []struct {
		name       string
		pods       []*apiv1.Pod
		daemonSets []*appsv1.DaemonSet
		forceDS    bool

		wantPods  []*apiv1.Pod
		wantError bool
	}{
		{
			name: "node without any pods",
		},
		{
			name: "node with non-DS/mirror pods",
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 100, 1, "n"),
			},
			wantPods: []*apiv1.Pod{},
		},
		{
			name: "node with a mirror pod",
			pods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p1", 100, 1, "n")),
			},
			wantPods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p1", 100, 1, "n")),
			},
		},
		{
			name: "node with a deleted mirror pod",
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
			pods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: testDaemonSets,
			wantPods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
			},
		},
		{
			name: "node with a DS pod [forceDS=true, no daemon sets]",
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
			pods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: testDaemonSets,
			forceDS:    true,
			wantPods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				buildDSPod(ds2, "n"),
			},
		},
		{
			name: "everything together [forceDS=false]",
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 100, 1, "n"),
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				setDeletionTimestamp(SetMirrorPodSpec(BuildScheduledTestPod("p4", 100, 1, "n"))),
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: testDaemonSets,
			wantPods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				buildDSPod(ds1, "n"),
			},
		},
		{
			name: "everything together [forceDS=true]",
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p1", 100, 1, "n"),
				BuildScheduledTestPod("p2", 100, 1, "n"),
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				setDeletionTimestamp(SetMirrorPodSpec(BuildScheduledTestPod("p4", 100, 1, "n"))),
				buildDSPod(ds1, "n"),
				setDeletionTimestamp(buildDSPod(ds2, "n")),
			},
			daemonSets: testDaemonSets,
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
			nodeGroupId := "nodeGroupId"
			exampleNodeInfo := framework.NewNodeInfo(exampleNode, nil)
			for _, pod := range tc.pods {
				exampleNodeInfo.AddPod(&framework.PodInfo{Pod: pod})
			}

			templateNodeInfo, err := SanitizedTemplateNodeInfoFromNodeInfo(exampleNodeInfo, nodeGroupId, tc.daemonSets, tc.forceDS, taints.TaintConfig{})
			if tc.wantError {
				if err == nil {
					t.Fatal("TemplateNodeInfoFromExampleNodeInfo(): want error, but got nil")
				} else {
					return
				}
			}
			if err != nil {
				t.Fatalf("TemplateNodeInfoFromExampleNodeInfo(): expected no error, but got %v", err)
			}

			// Verify that the taints are correctly sanitized.
			// Verify that the NodeInfo is sanitized using the node group id as base.
			// Pass empty string as nameSuffix so that it's auto-determined from the sanitized templateNodeInfo, because
			// TemplateNodeInfoFromExampleNodeInfo randomizes the suffix.
			// Pass non-empty expectedPods to verify that the set of pods is changed as expected (e.g. DS pods added, non-DS/deleted pods removed).
			if err := verifyNodeInfoSanitization(exampleNodeInfo, templateNodeInfo, tc.wantPods, "template-node-for-"+nodeGroupId, "", nil); err != nil {
				t.Fatalf("TemplateNodeInfoFromExampleNodeInfo(): NodeInfo wasn't properly sanitized: %v", err)
			}
		})
	}
}

func TestNodeInfoSanitizedDeepCopy(t *testing.T) {
	nodeName := "template-node"
	templateNode := BuildTestNode(nodeName, 1000, 1000)
	templateNode.Spec.Taints = []apiv1.Taint{
		{Key: "startup-taint", Value: "true", Effect: apiv1.TaintEffectNoSchedule},
		{Key: taints.ToBeDeletedTaint, Value: "2312532423", Effect: apiv1.TaintEffectNoSchedule},
		{Key: "a", Value: "b", Effect: apiv1.TaintEffectNoSchedule},
	}
	pods := []*framework.PodInfo{
		{Pod: BuildTestPod("p1", 80, 0, WithNodeName(nodeName))},
		{Pod: BuildTestPod("p2", 80, 0, WithNodeName(nodeName))},
	}
	templateNodeInfo := framework.NewNodeInfo(templateNode, nil, pods...)

	suffix := "abc"
	freshNodeInfo := NodeInfoSanitizedDeepCopy(templateNodeInfo, suffix)
	// Verify that the taints are not sanitized (they should be sanitized in the template already).
	// Verify that the NodeInfo is sanitized using the template Node name as base.
	initialTaints := templateNodeInfo.Node().Spec.Taints
	if err := verifyNodeInfoSanitization(templateNodeInfo, freshNodeInfo, nil, templateNodeInfo.Node().Name, suffix, initialTaints); err != nil {
		t.Fatalf("FreshNodeInfoFromTemplateNodeInfo(): NodeInfo wasn't properly sanitized: %v", err)
	}
}

func TestSanitizeNodeInfo(t *testing.T) {
	oldNodeName := "old-node"
	basicNode := BuildTestNode(oldNodeName, 1000, 1000)

	labelsNode := basicNode.DeepCopy()
	labelsNode.Labels = map[string]string{
		apiv1.LabelHostname: oldNodeName,
		"a":                 "b",
		"x":                 "y",
	}

	taintsNode := basicNode.DeepCopy()
	taintsNode.Spec.Taints = []apiv1.Taint{
		{Key: "startup-taint", Value: "true", Effect: apiv1.TaintEffectNoSchedule},
		{Key: taints.ToBeDeletedTaint, Value: "2312532423", Effect: apiv1.TaintEffectNoSchedule},
		{Key: "a", Value: "b", Effect: apiv1.TaintEffectNoSchedule},
	}
	taintConfig := taints.NewTaintConfig(config.AutoscalingOptions{StartupTaints: []string{"startup-taint"}})

	taintsLabelsNode := labelsNode.DeepCopy()
	taintsLabelsNode.Spec.Taints = taintsNode.Spec.Taints

	pods := []*framework.PodInfo{
		{Pod: BuildTestPod("p1", 80, 0, WithNodeName(oldNodeName))},
		{Pod: BuildTestPod("p2", 80, 0, WithNodeName(oldNodeName))},
	}

	for _, tc := range []struct {
		testName string

		nodeInfo    *framework.NodeInfo
		taintConfig *taints.TaintConfig

		wantTaints []apiv1.Taint
	}{
		{
			testName: "sanitize node",
			nodeInfo: framework.NewTestNodeInfo(basicNode),
		},
		{
			testName: "sanitize node labels",
			nodeInfo: framework.NewTestNodeInfo(labelsNode),
		},
		{
			testName:    "sanitize node taints - disabled",
			nodeInfo:    framework.NewTestNodeInfo(taintsNode),
			taintConfig: nil,
			wantTaints:  taintsNode.Spec.Taints,
		},
		{
			testName:    "sanitize node taints - enabled",
			nodeInfo:    framework.NewTestNodeInfo(taintsNode),
			taintConfig: &taintConfig,
			wantTaints:  []apiv1.Taint{{Key: "a", Value: "b", Effect: apiv1.TaintEffectNoSchedule}},
		},
		{
			testName: "sanitize pods",
			nodeInfo: framework.NewNodeInfo(basicNode, nil, pods...),
		},
		{
			testName:    "sanitize everything",
			nodeInfo:    framework.NewNodeInfo(taintsLabelsNode, nil, pods...),
			taintConfig: &taintConfig,
			wantTaints:  []apiv1.Taint{{Key: "a", Value: "b", Effect: apiv1.TaintEffectNoSchedule}},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			newNameBase := "node"
			suffix := "abc"
			sanitizedNodeInfo := sanitizeNodeInfo(tc.nodeInfo, newNameBase, suffix, tc.taintConfig)
			if err := verifyNodeInfoSanitization(tc.nodeInfo, sanitizedNodeInfo, nil, newNameBase, suffix, tc.wantTaints); err != nil {
				t.Fatalf("sanitizeNodeInfo(): NodeInfo wasn't properly sanitized: %v", err)
			}
		})
	}
}

// verifyNodeInfoSanitization verifies whether sanitizedNodeInfo was correctly sanitized starting from initialNodeInfo, with the provided
// nameBase and nameSuffix. The expected taints aren't auto-determined, so wantTaints should always be provided.
//
// If nameSuffix is an empty string, the suffix will be determined from sanitizedNodeInfo. This is useful if
// the test doesn't know/control the name suffix (e.g. because it's randomized by the tested function).
//
// If expectedPods is nil, the set of pods is expected not to change between initialNodeInfo and sanitizedNodeInfo. If the sanitization is
// expected to change the set of pods, the expected set should be passed to expectedPods.
func verifyNodeInfoSanitization(initialNodeInfo, sanitizedNodeInfo *framework.NodeInfo, expectedPods []*apiv1.Pod, nameBase, nameSuffix string, wantTaints []apiv1.Taint) error {
	if nameSuffix == "" {
		// Determine the suffix from the provided sanitized NodeInfo - it should be the last part of a dash-separated name.
		nameParts := strings.Split(sanitizedNodeInfo.Node().Name, "-")
		if len(nameParts) < 2 {
			return fmt.Errorf("sanitized NodeInfo name unexpected: want format <prefix>-<suffix>, got %q", sanitizedNodeInfo.Node().Name)
		}
		nameSuffix = nameParts[len(nameParts)-1]
	}
	if expectedPods != nil {
		// If the sanitization is expected to change the set of pods, hack the initial NodeInfo to have the expected pods.
		// Then we can just compare things pod-by-pod as if the set didn't change.
		initialNodeInfo = framework.NewNodeInfo(initialNodeInfo.Node(), nil)
		for _, pod := range expectedPods {
			initialNodeInfo.AddPod(&framework.PodInfo{Pod: pod})
		}
	}

	// Verification below assumes the same set of pods between initialNodeInfo and sanitizedNodeInfo.
	wantNodeName := fmt.Sprintf("%s-%s", nameBase, nameSuffix)
	if gotName := sanitizedNodeInfo.Node().Name; gotName != wantNodeName {
		return fmt.Errorf("want sanitized Node name %q, got %q", wantNodeName, gotName)
	}
	if gotUid, oldUid := sanitizedNodeInfo.Node().UID, initialNodeInfo.Node().UID; gotUid == "" || gotUid == oldUid {
		return fmt.Errorf("sanitized Node UID wasn't randomized - got %q, old UID was %q", gotUid, oldUid)
	}
	wantLabels := make(map[string]string)
	for k, v := range initialNodeInfo.Node().Labels {
		wantLabels[k] = v
	}
	wantLabels[apiv1.LabelHostname] = wantNodeName
	if diff := cmp.Diff(wantLabels, sanitizedNodeInfo.Node().Labels); diff != "" {
		return fmt.Errorf("sanitized Node labels unexpected, diff (-want +got): %s", diff)
	}
	if diff := cmp.Diff(wantTaints, sanitizedNodeInfo.Node().Spec.Taints); diff != "" {
		return fmt.Errorf("sanitized Node taints unexpected, diff (-want +got): %s", diff)
	}
	if diff := cmp.Diff(initialNodeInfo.Node(), sanitizedNodeInfo.Node(),
		cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "Labels", "UID"),
		cmpopts.IgnoreFields(apiv1.NodeSpec{}, "Taints"),
	); diff != "" {
		return fmt.Errorf("sanitized Node unexpected diff (-want +got): %s", diff)
	}

	oldPods := initialNodeInfo.Pods()
	newPods := sanitizedNodeInfo.Pods()
	if len(oldPods) != len(newPods) {
		return fmt.Errorf("want %d pods in sanitized NodeInfo, got %d", len(oldPods), len(newPods))
	}
	for i, newPod := range newPods {
		oldPod := oldPods[i]

		if newPod.Name == oldPod.Name || !strings.HasSuffix(newPod.Name, nameSuffix) {
			return fmt.Errorf("sanitized Pod name unexpected: want (different than %q, ending in %q), got %q", oldPod.Name, nameSuffix, newPod.Name)
		}
		if gotUid, oldUid := newPod.UID, oldPod.UID; gotUid == "" || gotUid == oldUid {
			return fmt.Errorf("sanitized Pod UID wasn't randomized - got %q, old UID was %q", gotUid, oldUid)
		}
		if gotNodeName := newPod.Spec.NodeName; gotNodeName != wantNodeName {
			return fmt.Errorf("want sanitized Pod.Spec.NodeName %q, got %q", wantNodeName, gotNodeName)
		}
		if diff := cmp.Diff(oldPod, newPod,
			cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "UID"),
			cmpopts.IgnoreFields(apiv1.PodSpec{}, "NodeName"),
		); diff != "" {
			return fmt.Errorf("sanitized Pod unexpected diff (-want +got): %s", diff)
		}
	}
	return nil
}

func buildDSPod(ds *appsv1.DaemonSet, nodeName string) *apiv1.Pod {
	pod := daemon.NewPod(ds, nodeName)
	pod.Name = fmt.Sprintf("%s-pod-%d", ds.Name, rand.Int63())
	ptrVal := true
	pod.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{Kind: "DaemonSet", UID: ds.UID, Name: ds.Name, Controller: &ptrVal},
	}
	return pod
}

func setDeletionTimestamp(pod *apiv1.Pod) *apiv1.Pod {
	now := metav1.NewTime(time.Now())
	pod.DeletionTimestamp = &now
	return pod
}

type fakeNodeGroup struct {
	id                     string
	templateNodeInfoResult *framework.NodeInfo
	templateNodeInfoErr    error
}

func (f *fakeNodeGroup) Id() string {
	return f.id
}

func (f *fakeNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return f.templateNodeInfoResult, f.templateNodeInfoErr
}
