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
	resourceapi "k8s.io/api/resource/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/controller/daemon"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	ndf "k8s.io/component-helpers/nodedeclaredfeatures"
	"k8s.io/dynamic-resource-allocation/resourceclaim"
	"k8s.io/kubernetes/pkg/features"
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
	ds4 = &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds4",
			Namespace: "ds4-namespace",
			UID:       types.UID("ds4"),
		},
		Spec: appsv1.DaemonSetSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					PriorityClassName: labels.SystemNodeCriticalLabel,
				},
			},
		},
	}
	testDaemonSets = []*appsv1.DaemonSet{ds1, ds2, ds3, ds4}
)

func TestSanitizedTemplateNodeInfoFromNodeGroup(t *testing.T) {
	exampleNode := BuildTestNode("n", 1000, 10)
	exampleNode.Spec.Taints = []apiv1.Taint{
		{Key: taints.ToBeDeletedTaint, Value: "2312532423", Effect: apiv1.TaintEffectNoSchedule},
	}
	exampleNode.Labels = map[string]string{
		"custom":                      "label",
		apiv1.LabelInstanceTypeStable: "some-instance",
		apiv1.LabelTopologyRegion:     "some-region",
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
				buildDSPod(ds4, "n"),
			},
		},
		{
			testName: "template with all kinds of pods",
			nodeGroup: &fakeNodeGroup{
				templateNodeInfoResult: framework.NewNodeInfo(exampleNode, nil,
					framework.NewPodInfo(BuildScheduledTestPod("p1", 100, 1, "n"), nil),
					framework.NewPodInfo(BuildScheduledTestPod("p2", 100, 1, "n"), nil),
					framework.NewPodInfo(SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")), nil),
					framework.NewPodInfo(setDeletionTimestamp(SetMirrorPodSpec(BuildScheduledTestPod("p4", 100, 1, "n"))), nil),
					framework.NewPodInfo(buildDSPod(ds1, "n"), nil),
					framework.NewPodInfo(setDeletionTimestamp(buildDSPod(ds2, "n")), nil),
				),
			},
			wantPods: []*apiv1.Pod{
				SetMirrorPodSpec(BuildScheduledTestPod("p3", 100, 1, "n")),
				buildDSPod(ds1, "n"),
				buildDSPod(ds2, "n"),
				buildDSPod(ds4, "n"),
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
			if err := verifyNodeInfoSanitization(tc.nodeGroup.templateNodeInfoResult, templateNodeInfo, tc.wantPods, "template-node-for-"+tc.nodeGroup.id, "", true, nil, false, false /*wantsCSINode*/); err != nil {
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
	exampleNode.Labels = map[string]string{
		"custom":                      "label",
		apiv1.LabelInstanceTypeStable: "some-instance",
		apiv1.LabelTopologyRegion:     "some-region",
	}
	exampleNode.Status.DeclaredFeatures = []string{"test-feature=true"}

	testCases := []struct {
		name       string
		pods       []*apiv1.Pod
		daemonSets []*appsv1.DaemonSet
		forceDS    bool
		csiNode    *storagev1.CSINode

		wantPods    []*apiv1.Pod
		wantCSINode bool
		wantError   bool
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
				buildDSPod(ds4, "n"),
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
				buildDSPod(ds4, "n"),
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
				buildDSPod(ds4, "n"),
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
				buildDSPod(ds4, "n"),
			},
		},
		{
			name: "node with CSINode",
			csiNode: &storagev1.CSINode{
				ObjectMeta: metav1.ObjectMeta{
					Name: "n",
					UID:  types.UID("original-csi-node-uid"),
				},
				Spec: storagev1.CSINodeSpec{
					Drivers: []storagev1.CSINodeDriver{
						{
							Name:         "test-driver",
							NodeID:       "test-node-id",
							TopologyKeys: []string{"topology-key"},
						},
					},
				},
			},
			wantCSINode: true,
		},
		{
			name: "node with CSINode and pods",
			pods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
			},
			daemonSets: testDaemonSets,
			csiNode: &storagev1.CSINode{
				ObjectMeta: metav1.ObjectMeta{
					Name: "n",
					UID:  types.UID("original-csi-node-uid"),
				},
				Spec: storagev1.CSINodeSpec{
					Drivers: []storagev1.CSINodeDriver{
						{
							Name:         "test-driver",
							NodeID:       "test-node-id",
							TopologyKeys: []string{"topology-key"},
						},
					},
				},
			},
			wantPods: []*apiv1.Pod{
				buildDSPod(ds1, "n"),
				buildDSPod(ds4, "n"),
			},
			wantCSINode: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nodeGroupId := "nodeGroupId"
			exampleNodeInfo := framework.NewNodeInfo(exampleNode, nil)
			for _, pod := range tc.pods {
				exampleNodeInfo.AddPod(framework.NewPodInfo(pod, nil))
			}
			if tc.csiNode != nil {
				exampleNodeInfo.SetCSINode(tc.csiNode)
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
			if err := verifyNodeInfoSanitization(exampleNodeInfo, templateNodeInfo, tc.wantPods, "template-node-for-"+nodeGroupId, "", false, nil, false, tc.wantCSINode); err != nil {
				t.Fatalf("TemplateNodeInfoFromExampleNodeInfo(): NodeInfo wasn't properly sanitized: %v", err)
			}
		})
	}
}

func TestSanitizedNodeInfo(t *testing.T) {
	nodeName := "template-node"
	templateNode := BuildTestNode(nodeName, 1000, 1000)
	templateNode.Spec.Taints = []apiv1.Taint{
		{Key: "startup-taint", Value: "true", Effect: apiv1.TaintEffectNoSchedule},
		{Key: taints.ToBeDeletedTaint, Value: "2312532423", Effect: apiv1.TaintEffectNoSchedule},
		{Key: "a", Value: "b", Effect: apiv1.TaintEffectNoSchedule},
	}
	templateNode.Labels = map[string]string{
		"custom":                      "label",
		apiv1.LabelInstanceTypeStable: "some-instance",
		apiv1.LabelTopologyRegion:     "some-region",
	}

	pods := []*framework.PodInfo{
		framework.NewPodInfo(BuildTestPod("p1", 80, 0, WithNodeName(nodeName)), nil),
		framework.NewPodInfo(BuildTestPod("p2", 80, 0, WithNodeName(nodeName)), nil),
	}
	templateNodeInfo := framework.NewNodeInfo(templateNode, nil, pods...)

	suffix := "abc"
	freshNodeInfo, err := SanitizedNodeInfo(templateNodeInfo, suffix)
	if err != nil {
		t.Fatalf("FreshNodeInfoFromTemplateNodeInfo(): want nil error, got %v", err)
	}
	// Verify that the taints are not sanitized (they should be sanitized in the template already).
	// Verify that the NodeInfo is sanitized using the template Node name as base.
	initialTaints := templateNodeInfo.Node().Spec.Taints
	if err := verifyNodeInfoSanitization(templateNodeInfo, freshNodeInfo, nil, templateNodeInfo.Node().Name, suffix, false, initialTaints, false, false /*wantCSINode*/); err != nil {
		t.Fatalf("FreshNodeInfoFromTemplateNodeInfo(): NodeInfo wasn't properly sanitized: %v", err)
	}
}

func TestCreateSanitizedNodeInfo(t *testing.T) {
	oldNodeName := "old-node"
	basicNode := BuildTestNode(oldNodeName, 1000, 1000)

	labelsNode := basicNode.DeepCopy()
	labelsNode.Labels = map[string]string{
		apiv1.LabelHostname:           oldNodeName,
		"a":                           "b",
		"x":                           "y",
		apiv1.LabelInstanceTypeStable: "some-instance",
		apiv1.LabelTopologyRegion:     "some-region",
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

	nodeWithDeclaredFeatures := basicNode.DeepCopy()
	nodeWithDeclaredFeatures.Status.DeclaredFeatures = []string{"FeatureA,FeatureB"}

	resourceSlices := []*resourceapi.ResourceSlice{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "slice1", UID: "slice1Uid"},
			Spec: resourceapi.ResourceSliceSpec{
				NodeName: &oldNodeName,
				Pool: resourceapi.ResourcePool{
					Name:               "pool1",
					ResourceSliceCount: 1,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "slice2", UID: "slice2Uid"},
			Spec: resourceapi.ResourceSliceSpec{
				NodeName: &oldNodeName,
				Pool: resourceapi.ResourcePool{
					Name:               "pool2",
					ResourceSliceCount: 1,
				},
			},
		},
	}

	pod1 := BuildTestPod("pod1", 80, 0, WithNodeName(oldNodeName))
	pod2 := BuildTestPod("pod2", 80, 0, WithNodeName(oldNodeName))

	pod1WithClaims := BuildTestPod("pod1", 80, 0, WithNodeName(oldNodeName),
		WithResourceClaim("claim1", "pod1Claim1", "pod1ClaimTemplate"),
		WithResourceClaim("claim2", "pod1Claim2", "pod1ClaimTemplate"),
		WithResourceClaim("claim3", "sharedClaim1", "sharedClaimTemplate"),
		WithResourceClaim("claim4", "sharedClaim2", "sharedClaimTemplate"),
	)
	pod2WithClaims := BuildTestPod("pod2", 80, 0, WithNodeName(oldNodeName),
		WithResourceClaim("claim1", "pod2Claim1", "pod2ClaimTemplate"),
		WithResourceClaim("claim2", "pod2Claim2", "pod2ClaimTemplate"),
		WithResourceClaim("claim3", "sharedClaim1", "sharedClaimTemplate"),
		WithResourceClaim("claim4", "sharedClaim2", "sharedClaimTemplate"),
	)
	nodeAllocation := &resourceapi.AllocationResult{
		NodeSelector: &apiv1.NodeSelector{NodeSelectorTerms: []apiv1.NodeSelectorTerm{{
			MatchFields: []apiv1.NodeSelectorRequirement{
				{Key: "metadata.name", Operator: apiv1.NodeSelectorOpIn, Values: []string{oldNodeName}},
			}},
		}},
	}
	pod1Claim1 := &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "pod1claim1", UID: "pod1claim1Uid", Namespace: "default"}}
	pod1Claim2 := &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "pod1claim2", UID: "pod1claim2Uid", Namespace: "default"}}
	pod2Claim1 := &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "pod2claim1", UID: "pod2claim1Uid", Namespace: "default"}}
	pod2Claim2 := &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "pod2claim2", UID: "pod2claim2Uid", Namespace: "default"}}
	sharedClaim1 := &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "sharedClaim1", UID: "sharedClaim1Uid", Namespace: "default"}}
	sharedClaim2 := &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "sharedClaim2", UID: "sharedClaim2Uid", Namespace: "default"}}
	pod1ResourceClaims := []*resourceapi.ResourceClaim{
		drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(drautils.TestClaimWithPodOwnership(pod1WithClaims, pod1Claim1), nodeAllocation), pod1WithClaims),
		drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(drautils.TestClaimWithPodOwnership(pod1WithClaims, pod1Claim2), nodeAllocation), pod1WithClaims),
		drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim1, nil), pod1WithClaims, pod2WithClaims),
		drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim2, nil), pod1WithClaims, pod2WithClaims),
	}
	pod2ResourceClaims := []*resourceapi.ResourceClaim{
		drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(drautils.TestClaimWithPodOwnership(pod2WithClaims, pod2Claim1), nodeAllocation), pod2WithClaims),
		drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(drautils.TestClaimWithPodOwnership(pod2WithClaims, pod2Claim2), nodeAllocation), pod2WithClaims),
		drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim1, nil), pod1WithClaims, pod2WithClaims),
		drautils.TestClaimWithPodReservations(drautils.TestClaimWithAllocation(sharedClaim2, nil), pod1WithClaims, pod2WithClaims),
	}

	tests := []struct {
		testName string

		nodeInfo                    *framework.NodeInfo
		taintConfig                 *taints.TaintConfig
		nodeDeclaredFeaturesEnabled bool
		wantTaints                  []apiv1.Taint
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
			testName: "sanitize node with ResourceSlices",
			nodeInfo: framework.NewNodeInfo(basicNode, resourceSlices),
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
			nodeInfo: framework.NewNodeInfo(basicNode, nil, framework.NewPodInfo(pod1, nil), framework.NewPodInfo(pod2, nil)),
		},
		{
			testName: "sanitize pods with ResourceClaims",
			nodeInfo: framework.NewNodeInfo(basicNode, nil, framework.NewPodInfo(pod1WithClaims, pod1ResourceClaims), framework.NewPodInfo(pod2WithClaims, pod2ResourceClaims)),
		},
		{
			testName:    "sanitize everything",
			nodeInfo:    framework.NewNodeInfo(taintsLabelsNode, resourceSlices, framework.NewPodInfo(pod1WithClaims, pod1ResourceClaims), framework.NewPodInfo(pod2WithClaims, pod2ResourceClaims)),
			taintConfig: &taintConfig,
			wantTaints:  []apiv1.Taint{{Key: "a", Value: "b", Effect: apiv1.TaintEffectNoSchedule}},
		},
		{
			testName:                    "sanitize node with NodeDeclaredFeatures enabled",
			nodeInfo:                    framework.NewNodeInfo(nodeWithDeclaredFeatures, resourceSlices, framework.NewPodInfo(pod1WithClaims, pod1ResourceClaims), framework.NewPodInfo(pod2WithClaims, pod2ResourceClaims)),
			nodeDeclaredFeaturesEnabled: true,
		},
		{
			testName:                    "sanitize node with NodeDeclaredFeatures disabled",
			nodeInfo:                    framework.NewNodeInfo(nodeWithDeclaredFeatures, resourceSlices, framework.NewPodInfo(pod1WithClaims, pod1ResourceClaims), framework.NewPodInfo(pod2WithClaims, pod2ResourceClaims)),
			nodeDeclaredFeaturesEnabled: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			utilfeature.DefaultMutableFeatureGate.Set(fmt.Sprintf("%s=%v", features.NodeDeclaredFeatures, tc.nodeDeclaredFeaturesEnabled))
			newNameBase := "node"
			suffix := "abc"
			nodeInfo, err := createSanitizedNodeInfo(tc.nodeInfo, newNameBase, suffix, tc.taintConfig)
			if err != nil {
				t.Fatalf("sanitizeNodeInfo(): want nil error, got %v", err)
			}
			if err := verifyNodeInfoSanitization(tc.nodeInfo, nodeInfo, nil, newNameBase, suffix, false, tc.wantTaints, tc.nodeDeclaredFeaturesEnabled, false /*wantCSINode*/); err != nil {
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
func verifyNodeInfoSanitization(initialNodeInfo, sanitizedNodeInfo *framework.NodeInfo, expectedPods []*apiv1.Pod, nameBase, nameSuffix string, wantDeprecatedLabels bool, wantTaints []apiv1.Taint, nodeDeclaredFeaturesEnabled bool, wantsCSINode bool) error {
	if nameSuffix == "" {
		// Determine the suffix from the provided sanitized NodeInfo - it should be the last part of a dash-separated name.
		nameParts := strings.Split(sanitizedNodeInfo.Node().Name, "-")
		if len(nameParts) < 2 {
			return fmt.Errorf("sanitized NodeInfo name unexpected: want format <prefix>-<suffix>, got %q", sanitizedNodeInfo.Node().Name)
		}
		nameSuffix = nameParts[len(nameParts)-1]
	}
	// extract CSINode before initialNodeInfo gets overwritten below
	initialCSINode := initialNodeInfo.CSINode
	sanitizedCSINode := sanitizedNodeInfo.CSINode

	if expectedPods != nil {
		// If the sanitization is expected to change the set of pods, hack the initial NodeInfo to have the expected pods.
		// Then we can just compare things pod-by-pod as if the set didn't change.
		initialNodeInfo = framework.NewNodeInfo(initialNodeInfo.Node(), nil)
		for _, pod := range expectedPods {
			initialNodeInfo.AddPod(framework.NewPodInfo(pod, nil))
		}
	}

	// Verification below assumes the same set of pods between initialNodeInfo and sanitizedNodeInfo.
	wantNodeName := fmt.Sprintf("%s-%s", nameBase, nameSuffix)
	if err := verifySanitizedNode(initialNodeInfo.Node(), sanitizedNodeInfo.Node(), wantNodeName, wantDeprecatedLabels, wantTaints); err != nil {
		return err
	}
	if err := verifySanitizedNodeResourceSlices(initialNodeInfo.LocalResourceSlices, sanitizedNodeInfo.LocalResourceSlices, nameSuffix); err != nil {
		return err
	}
	if err := verifySanitizedPods(initialNodeInfo.Pods(), sanitizedNodeInfo.Pods(), wantNodeName, nameSuffix); err != nil {
		return err
	}

	gotDeclaredFeatures := sanitizedNodeInfo.GetNodeDeclaredFeatures()
	// Verify DeclaredFeatures on the NodeInfo struct
	if nodeDeclaredFeaturesEnabled {
		wantDeclaredFeatures := ndf.NewFeatureSet(initialNodeInfo.Node().Status.DeclaredFeatures...)
		if diff := cmp.Diff(wantDeclaredFeatures, gotDeclaredFeatures); diff != "" {
			return fmt.Errorf("sanitized NodeInfo.DeclaredFeatures unexpected, diff (-want +got): %s", diff)
		}
	} else if gotDeclaredFeatures.Len() != 0 {
		return fmt.Errorf("sanitized NodeInfo.DeclaredFeatures unexpected: got %v, want empty when feature gate is disabled", sanitizedNodeInfo.GetNodeDeclaredFeatures())
	}

	if wantsCSINode {
		if err := verifySanitizedCSINode(initialCSINode, sanitizedCSINode, sanitizedNodeInfo.Node()); err != nil {
			return err
		}
	} else {
		if sanitizedNodeInfo.CSINode != nil {
			return fmt.Errorf("unexpected CSINode %v in sanitized NodeInfo", sanitizedNodeInfo.CSINode)
		}
	}

	return nil
}

func verifySanitizedNode(initialNode, sanitizedNode *apiv1.Node, wantNodeName string, wantDeprecatedLabels bool, wantTaints []apiv1.Taint) error {
	if gotName := sanitizedNode.Name; gotName != wantNodeName {
		return fmt.Errorf("want sanitized Node name %q, got %q", wantNodeName, gotName)
	}
	if gotUid, oldUid := sanitizedNode.UID, initialNode.UID; gotUid == "" || gotUid == oldUid {
		return fmt.Errorf("sanitized Node UID wasn't randomized - got %q, old UID was %q", gotUid, oldUid)
	}

	wantLabels := make(map[string]string)
	for k, v := range initialNode.Labels {
		wantLabels[k] = v
	}
	wantLabels[apiv1.LabelHostname] = wantNodeName
	if wantDeprecatedLabels {
		labels.UpdateDeprecatedLabels(wantLabels)
	}
	if diff := cmp.Diff(wantLabels, sanitizedNode.Labels); diff != "" {
		return fmt.Errorf("sanitized Node labels unexpected, diff (-want +got): %s", diff)
	}

	if diff := cmp.Diff(wantTaints, sanitizedNode.Spec.Taints); diff != "" {
		return fmt.Errorf("sanitized Node taints unexpected, diff (-want +got): %s", diff)
	}

	if diff := cmp.Diff(initialNode, sanitizedNode,
		cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "Labels", "UID"),
		cmpopts.IgnoreFields(apiv1.NodeSpec{}, "Taints"),
	); diff != "" {
		return fmt.Errorf("sanitized Node unexpected diff (-want +got): %s", diff)
	}

	return nil
}

func verifySanitizedPods(initialPods, sanitizedPods []*framework.PodInfo, wantNodeName, nameSuffix string) error {
	if len(initialPods) != len(sanitizedPods) {
		return fmt.Errorf("want %d pods in sanitized NodeInfo, got %d", len(initialPods), len(sanitizedPods))
	}

	for i, sanitizedPod := range sanitizedPods {
		initialPod := initialPods[i]

		if sanitizedPod.Name == initialPod.Name || !strings.HasSuffix(sanitizedPod.Name, nameSuffix) {
			return fmt.Errorf("sanitized Pod name unexpected: want (different than %q, ending in %q), got %q", initialPod.Name, nameSuffix, sanitizedPod.Name)
		}
		if gotUid, oldUid := sanitizedPod.UID, initialPod.UID; gotUid == "" || gotUid == oldUid {
			return fmt.Errorf("sanitized Pod UID wasn't randomized - got %q, old UID was %q", gotUid, oldUid)
		}

		if gotNodeName := sanitizedPod.Spec.NodeName; gotNodeName != wantNodeName {
			return fmt.Errorf("want sanitized Pod.Spec.NodeName %q, got %q", wantNodeName, gotNodeName)
		}

		if err := verifySanitizedPodResourceClaimStatuses(initialPod.Status.ResourceClaimStatuses, sanitizedPod.Status.ResourceClaimStatuses, nameSuffix); err != nil {
			return fmt.Errorf("verifying Pod.Status.ResourceClaimStatuses in sanitized NodeInfo failed for pod %s: %v", sanitizedPod.Name, err)
		}

		if diff := cmp.Diff(initialPod.Pod, sanitizedPod.Pod,
			cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "UID"),
			cmpopts.IgnoreFields(apiv1.PodSpec{}, "NodeName"),
			cmpopts.IgnoreFields(apiv1.PodStatus{}, "ResourceClaimStatuses"),
		); diff != "" {
			return fmt.Errorf("sanitized Pod unexpected diff (-want +got): %s", diff)
		}

		if err := verifySanitizedPodResourceClaims(initialPod, sanitizedPod, nameSuffix); err != nil {
			return err
		}
	}

	return nil
}

func verifySanitizedNodeResourceSlices(initialSlices, sanitizedSlices []*resourceapi.ResourceSlice, nameSuffix string) error {
	if len(initialSlices) != len(sanitizedSlices) {
		return fmt.Errorf("want %d LocalResourceSlices in sanitized NodeInfo, got %d", len(initialSlices), len(sanitizedSlices))
	}

	for i, newSlice := range sanitizedSlices {
		oldSlice := initialSlices[i]

		if newSlice.Name == oldSlice.Name || !strings.HasSuffix(newSlice.Name, nameSuffix) {
			return fmt.Errorf("sanitized ResourceSlice name unexpected: want (different than %q, ending in %q), got %q", oldSlice.Name, nameSuffix, newSlice.Name)
		}
		if gotUid, oldUid := newSlice.UID, oldSlice.UID; gotUid == "" || gotUid == oldUid {
			return fmt.Errorf("sanitized ResourceSlice UID wasn't randomized - got %q, old UID was %q", gotUid, oldUid)
		}

		// Don't verify ResourceSlice sanitization in detail, there are separate unit tests for that. Just assert that the Spec changed to confirm that it was sanitized.
		if cmp.Equal(oldSlice.Spec, newSlice.Spec) {
			return fmt.Errorf("sanitized ResourceSlice Spec is identical to original Spec: %v", newSlice.Spec)
		}
	}

	return nil
}

func verifySanitizedPodResourceClaims(initialPod, sanitizedPod *framework.PodInfo, nameSuffix string) error {
	initialClaims := initialPod.NeededResourceClaims
	sanitizedClaims := sanitizedPod.NeededResourceClaims
	owningPod := initialPod.Pod

	if len(initialClaims) != len(sanitizedClaims) {
		return fmt.Errorf("want %d NeededResourceClaims in sanitized NodeInfo, got %d", len(initialClaims), len(sanitizedClaims))
	}

	for i, sanitizedClaim := range sanitizedClaims {
		initialClaim := initialClaims[i]

		// Pod-owned claims should be sanitized, other claims shouldn't.
		err := resourceclaim.IsForPod(owningPod, initialClaim)
		isPodOwned := err == nil
		if isPodOwned {
			// Pod-owned claim, verify that it was sanitized.
			if sanitizedClaim.Name == initialClaim.Name || !strings.HasSuffix(sanitizedClaim.Name, nameSuffix) {
				return fmt.Errorf("sanitized ResourceClaim name unexpected: want (different than %q, ending in %q), got %q", initialClaim.Name, nameSuffix, sanitizedClaim.Name)
			}
			if gotUid, oldUid := sanitizedClaim.UID, initialClaim.UID; gotUid == "" || gotUid == oldUid {
				return fmt.Errorf("sanitized ResourceClaim UID wasn't randomized - got %q, old UID was %q", gotUid, oldUid)
			}

			// Don't verify ResourceClaim sanitization in detail, there are separate unit tests for that. Just assert that the Status changed to confirm that it was sanitized.
			if cmp.Equal(initialClaim.Status, sanitizedClaim.Status) {
				return fmt.Errorf("sanitized ResourceClaim Status is identical to original Status: %v", sanitizedClaim.Status)
			}
		} else {
			// Shared claim, verify that it wasn't sanitized.
			if diff := cmp.Diff(initialClaim, sanitizedClaim); diff != "" {
				return fmt.Errorf("shared ResourceClaim unexpectedly sanitized: diff from original (-want +got): %s", diff)
			}
		}
	}

	return nil
}

func verifySanitizedPodResourceClaimStatuses(initialStatuses, sanitizedStatuses []apiv1.PodResourceClaimStatus, nameSuffix string) error {
	if len(initialStatuses) != len(sanitizedStatuses) {
		return fmt.Errorf("want %d Pod.Status.ResourceClaimStatuses in sanitized NodeInfo, got %d", len(initialStatuses), len(sanitizedStatuses))
	}

	for i, sanitizedStatus := range sanitizedStatuses {
		initialStatus := initialStatuses[i]

		if initialStatus.Name != sanitizedStatus.Name {
			return fmt.Errorf("sanitized ResourceClaimStatus name unexpected: want %q, got %q", initialStatus.Name, sanitizedStatus.Name)
		}

		if initialStatus.ResourceClaimName != nil {
			if sanitizedStatus.ResourceClaimName == nil {
				return fmt.Errorf("sanitized ResourceClaimStatus %q: ResourceClaimName unexpectedly nil", initialStatus.Name)
			}
			initialClaimName := *initialStatus.ResourceClaimName
			sanitizedClaimName := *sanitizedStatus.ResourceClaimName

			if sanitizedClaimName == initialClaimName || !strings.HasSuffix(sanitizedClaimName, nameSuffix) {
				return fmt.Errorf("sanitized ResourceClaimStatus %q: ResourceClaimName unexpected: want (different than %q, ending in %q), got %q", initialStatus.Name, initialClaimName, nameSuffix, sanitizedClaimName)
			}
		}
	}
	return nil
}

func verifySanitizedCSINode(initialCSINode, sanitizedCSINode *storagev1.CSINode, templateNode *apiv1.Node) error {
	if sanitizedCSINode == nil {
		return fmt.Errorf("sanitized CSINode is nil")
	}

	// Verify name matches template node name
	if sanitizedCSINode.Name != templateNode.Name {
		return fmt.Errorf("sanitized CSINode name unexpected: want %q, got %q", templateNode.Name, sanitizedCSINode.Name)
	}

	// Verify UID is different from original
	if sanitizedCSINode.UID == "" || sanitizedCSINode.UID == initialCSINode.UID {
		return fmt.Errorf("sanitized CSINode UID wasn't randomized - got %q, old UID was %q", sanitizedCSINode.UID, initialCSINode.UID)
	}

	// Verify owner references point to template node
	if len(sanitizedCSINode.OwnerReferences) != 1 {
		return fmt.Errorf("sanitized CSINode should have exactly one owner reference, got %d", len(sanitizedCSINode.OwnerReferences))
	}
	ownerRef := sanitizedCSINode.OwnerReferences[0]
	if ownerRef.Kind != "Node" {
		return fmt.Errorf("sanitized CSINode owner reference kind unexpected: want %q, got %q", "Node", ownerRef.Kind)
	}
	if ownerRef.Name != templateNode.Name {
		return fmt.Errorf("sanitized CSINode owner reference name unexpected: want %q, got %q", templateNode.Name, ownerRef.Name)
	}
	if ownerRef.UID != templateNode.UID {
		return fmt.Errorf("sanitized CSINode owner reference UID unexpected: want %q, got %q", templateNode.UID, ownerRef.UID)
	}

	// Verify spec is preserved (deep copied)
	if diff := cmp.Diff(initialCSINode.Spec, sanitizedCSINode.Spec); diff != "" {
		return fmt.Errorf("sanitized CSINode spec unexpected diff (-want +got): %s", diff)
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
