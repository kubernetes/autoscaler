/*
Copyright 2025 The Kubernetes Authors.

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

package core

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	sdplanner "k8s.io/autoscaler/cluster-autoscaler/core/scaledown/planner"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/predicate"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	csinodeprovider "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/provider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	informers "k8s.io/client-go/informers"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	v1storagelister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/utils/ptr"
)

type testNodeGroupCSI struct {
	name                string
	cpu, mem            int64
	csiNodeTemplateFunc func(nodeName string) *storagev1.CSINode
}

type fakeCSINodeLister struct {
	nodes map[string]*storagev1.CSINode
}

var _ v1storagelister.CSINodeLister = (*fakeCSINodeLister)(nil)

func (l *fakeCSINodeLister) List(_ labels.Selector) ([]*storagev1.CSINode, error) {
	out := make([]*storagev1.CSINode, 0, len(l.nodes))
	for _, n := range l.nodes {
		out = append(out, n)
	}
	return out, nil
}

func (l *fakeCSINodeLister) Get(name string) (*storagev1.CSINode, error) {
	n, ok := l.nodes[name]
	if !ok {
		return nil, fmt.Errorf("csinode %q not found", name)
	}
	return n, nil
}

func TestStaticAutoscalerCSI(t *testing.T) {
	now := time.Now()

	const (
		driverName     = "ebs.csi.aws.com"
		storageClass   = "sc-ebs"
		volumeLimit    = int32(10)
		templateVolUse = 2
		podVolUse      = 6
	)

	node1GroupCSI := &testNodeGroupCSI{
		name:                "node1CSINode",
		cpu:                 1000,
		mem:                 1000,
		csiNodeTemplateFunc: csiNodeTemplate(driverName, ptr.To(volumeLimit)),
	}

	testCases := map[string]struct {
		nodeGroups         map[*testNodeGroupCSI]int
		templatePods       map[string][]*apiv1.Pod
		pods               []*apiv1.Pod
		expectedScaleUps   map[string]int
		expectedScaleDowns map[string][]string
	}{
		"scale-up: CSI volume limit blocks packing, requires 2 extra nodes": {
			nodeGroups: map[*testNodeGroupCSI]int{node1GroupCSI: 1},
			templatePods: map[string][]*apiv1.Pod{
				// A DaemonSet-like pod that consumes some CSI volumes on every node.
				node1GroupCSI.name: {buildCSITestPod("template-ds", templateVolUse, node1GroupCSI.name+"-0")},
			},
			// Two pending pods, each using podVolUse volumes.
			pods: []*apiv1.Pod{
				buildCSITestPod("pending-0", podVolUse, ""),
				buildCSITestPod("pending-1", podVolUse, ""),
				buildCSITestPod("pending-2", podVolUse, ""),
			},
			// One existing node has the CSI pod already; only one pending pod can fit because (2 + 6 + 6) > 10.
			// CA should add one node for the second pending pod.
			expectedScaleUps: map[string]int{node1GroupCSI.name: 2},
		},
		"scale-from-zero: CSI volume limit blocks packing, requires >1 new nodes": {
			// No nodes in the node group initially.
			nodeGroups: map[*testNodeGroupCSI]int{node1GroupCSI: 0},
			templatePods: map[string][]*apiv1.Pod{
				// DaemonSet-like pod is part of the template (will run on every new node), so it should be
				// accounted for in binpacking even when scaling from zero.
				node1GroupCSI.name: {buildCSITestPod("template-ds", templateVolUse, node1GroupCSI.name+
					"-0")},
			},
			pods: []*apiv1.Pod{
				buildCSITestPod("pending-0", podVolUse, ""),
				buildCSITestPod("pending-1", podVolUse, ""),
				buildCSITestPod("pending-2", podVolUse, ""),
			},
			// With volumeLimit=10, templateVolUse=2 and podVolUse=6:
			// - 2+6 fits, but 2+6+6 doesn't, so each new node can host at most one pending pod.
			// Scaling from 0 with 3 pending pods should require 3 new nodes (>1).
			expectedScaleUps: map[string]int{node1GroupCSI.name: 3},
		},
		"scale-from-1: existing node can support limits, no scaling needed": {
			nodeGroups: map[*testNodeGroupCSI]int{node1GroupCSI: 1},
			templatePods: map[string][]*apiv1.Pod{
				// A DaemonSet-like pod that consumes some CSI volumes on every node.
				node1GroupCSI.name: {buildCSITestPod("template-ds", templateVolUse, node1GroupCSI.name+"-0")},
			},
			// Two pending pods, each using podVolUse volumes.
			pods: []*apiv1.Pod{
				buildCSITestPod("pending-0", podVolUse, ""),
			},
		},
		"scale-down: single pod left with 6 volumes": {
			nodeGroups: map[*testNodeGroupCSI]int{node1GroupCSI: 3},
			pods: []*apiv1.Pod{
				// Keep one node non-empty; the other two should be removable.
				buildCSITestPod("scheduled-0", podVolUse, node1GroupCSI.name+"-0"),
			},
			expectedScaleDowns: map[string][]string{node1GroupCSI.name: {node1GroupCSI.name + "-1", node1GroupCSI.name + "-2"}},
		},
	}

	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			var nodeGroups []*nodeGroup
			var allNodes []*apiv1.Node
			csiNodes := map[string]*storagev1.CSINode{}
			var allPods []*apiv1.Pod

			// Collect k8s objects required by the kube-scheduler NodeVolumeLimits plugin:
			// StorageClass, CSIDriver and PVCs referenced by pods.
			var k8sObjects []runtime.Object

			for nodeGroupDef, count := range tc.nodeGroups {
				var nodes []*apiv1.Node
				for i := 0; i < count; i++ {
					node := BuildTestNode(fmt.Sprintf("%s-%d", nodeGroupDef.name, i), nodeGroupDef.cpu, nodeGroupDef.mem)
					// Make nodes old enough for scale-down consideration
					nodeCreationTime := now.Add(-10 * time.Minute)
					node.CreationTimestamp = metav1.NewTime(nodeCreationTime)
					SetNodeReadyState(node, true, nodeCreationTime)
					nodes = append(nodes, node)
					allNodes = append(allNodes, node)

					// CSINode for the real node in the cluster (used by StaticAutoscaler snapshotting and CSI-aware scheduling).
					csiNode := nodeGroupDef.csiNodeTemplateFunc(node.Name)
					csiNode.UID = types.UID(fmt.Sprintf("%s-uid", node.Name))
					csiNodes[node.Name] = csiNode
				}

				// Template nodeInfo for simulating new nodes in this node group.
				templateNode := BuildTestNode(fmt.Sprintf("%s-template", nodeGroupDef.name), nodeGroupDef.cpu, nodeGroupDef.mem)
				templateNodeInfo := framework.NewNodeInfo(templateNode, nil)
				if nodeGroupDef.csiNodeTemplateFunc != nil {
					templateNodeInfo.SetCSINode(nodeGroupDef.csiNodeTemplateFunc(templateNode.Name))
				}
				for _, p := range tc.templatePods[nodeGroupDef.name] {
					templateNodeInfo.AddPod(framework.NewPodInfo(p, nil))
				}

				nodeGroups = append(nodeGroups, &nodeGroup{
					name:     nodeGroupDef.name,
					template: templateNodeInfo,
					min:      0,
					max:      10,
					nodes:    nodes,
				})
			}

			allPods = append(allPods, tc.pods...)

			// Create StorageClass + CSIDriver used by PVCs.
			sc := &storagev1.StorageClass{
				ObjectMeta:        metav1.ObjectMeta{Name: storageClass},
				Provisioner:       driverName,
				VolumeBindingMode: ptr.To(storagev1.VolumeBindingWaitForFirstConsumer),
			}
			csiDriver := &storagev1.CSIDriver{
				ObjectMeta: metav1.ObjectMeta{Name: driverName},
			}
			k8sObjects = append(k8sObjects, sc, csiDriver)

			k8sObjects = append(k8sObjects, pvcsForPods(allPods, storageClass)...)

			// Create a framework handle with informer-backed listers for StorageClass/PVC/CSIDriver.
			client := clientsetfake.NewSimpleClientset(k8sObjects...)
			informerFactory := informers.NewSharedInformerFactory(client, 0)
			fwHandle, err := framework.NewHandle(context.Background(), informerFactory, nil, false, true)
			require.NoError(t, err)
			stopCh := make(chan struct{})
			t.Cleanup(func() { close(stopCh) })
			informerFactory.Start(stopCh)
			informerFactory.WaitForCacheSync(stopCh)

			mocks := newCommonMocks()
			mocks.readyNodeLister.SetNodes(allNodes)
			mocks.allNodeLister.SetNodes(allNodes)
			mocks.daemonSetLister.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil)
			mocks.podDisruptionBudgetLister.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil)
			mocks.allPodLister.On("List").Return(allPods, nil)
			for nodeGroup, delta := range tc.expectedScaleUps {
				mocks.onScaleUp.On("ScaleUp", nodeGroup, delta).Return(nil).Once()
			}

			allExpectedScaleDowns := 0

			for nodeGroup, nodes := range tc.expectedScaleDowns {
				for _, node := range nodes {
					mocks.onScaleDown.On("ScaleDown", nodeGroup, node).Return(nil).Once()
					allExpectedScaleDowns++
				}
			}

			setupConfig := &autoscalerSetupConfig{
				autoscalingOptions: config.AutoscalingOptions{
					NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
						ScaleDownUnneededTime:         time.Minute,
						ScaleDownUnreadyTime:          time.Minute,
						ScaleDownUtilizationThreshold: 0.7,
						MaxNodeProvisionTime:          time.Hour,
					},
					ScaleDownEnabled:               true,
					MaxNodesTotal:                  1000,
					MaxCoresTotal:                  1000,
					MaxMemoryTotal:                 100000000,
					ScaleUpFromZero:                true,
					CSINodeAwareSchedulingEnabled:  true,
					PredicateParallelism:           1,
					MaxNodeGroupBinpackingDuration: 1 * time.Hour,
				},
				nodeGroups:             nodeGroups,
				nodeStateUpdateTime:    now,
				mocks:                  mocks,
				optionsBlockDefaulting: false,
				// setupCloudProvider sends on this channel from the ScaleDown callback; it must be able to
				// accommodate multiple deletions without blocking the test.
				nodesDeleted: make(chan bool, allExpectedScaleDowns+1),
			}

			autoscaler, err := setupAutoscaler(setupConfig)
			require.NoError(t, err)

			// Replace framework handle + snapshot with CSI-aware snapshot using the handle that has PVC/SC/CSIDriver informers.
			autoscaler.AutoscalingContext.FrameworkHandle = fwHandle
			autoscaler.AutoscalingContext.ClusterSnapshot = predicate.NewPredicateSnapshot(store.NewBasicSnapshotStore(), fwHandle, true, 1, true)

			// Provide CSI nodes snapshotting for the real nodes.
			autoscaler.AutoscalingContext.CsiProvider = csinodeprovider.NewCSINodeProvider(&fakeCSINodeLister{nodes: csiNodes})

			// IMPORTANT: setupAutoscaler() builds the scale-down planner with whatever ClusterSnapshot existed at that time.
			// We replaced ClusterSnapshot above (to use a CSI-aware PredicateSnapshot), so we must rebuild the planner so its
			// internal removal simulator uses the same snapshot instance that RunOnce() initializes.
			deleteOptions := options.NewNodeDeleteOptions(autoscaler.AutoscalingOptions)
			drainabilityRules := rules.Default(deleteOptions)
			newSDPlanner := sdplanner.New(autoscaler.AutoscalingContext, autoscaler.processors, deleteOptions, drainabilityRules)
			autoscaler.scaleDownPlanner = newSDPlanner
			autoscaler.processorCallbacks.scaleDownPlanner = newSDPlanner

			scaleUpProcessor := &fakeScaleUpStatusProcessor{}
			scaleDownProcessor := &fakeScaleDownStatusProcessor{}
			autoscaler.processors.ScaleUpStatusProcessor = scaleUpProcessor
			autoscaler.processors.ScaleDownStatusProcessor = scaleDownProcessor

			if len(tc.expectedScaleDowns) > 0 {
				err = autoscaler.RunOnce(now)
				assert.NoError(t, err)
			}

			// Run one autoscaler loop.
			require.NoError(t, autoscaler.RunOnce(now.Add(2*time.Minute)))

			// Scale-down is a multi-iteration process (mark unneeded -> taint -> delete after delay).
			// Run one more loop to allow the actuator to call the cloud provider ScaleDown callbacks.
			if len(tc.expectedScaleDowns) > 0 {
				require.NoError(t, autoscaler.RunOnce(now.Add(4*time.Minute)))
				for range allExpectedScaleDowns {
					select {
					case <-setupConfig.nodesDeleted:
						// ok
					case <-time.After(20 * time.Second):
						t.Fatalf("Scale-down deletes not finished")
					}
				}
			}

			// If we expected a scale-up but didn't call ScaleUp(), fail with a helpful snapshot of the scale-up status.
			if len(tc.expectedScaleUps) > 0 && len(mocks.onScaleUp.Calls) == 0 {
				if scaleUpProcessor.lastStatus == nil {
					t.Fatalf("expected ScaleUp to be called, but it wasn't (scaleUpStatus was nil)")
				}
				var reasonDump string
				if len(scaleUpProcessor.lastStatus.PodsRemainUnschedulable) > 0 {
					noScale := scaleUpProcessor.lastStatus.PodsRemainUnschedulable[0]
					reasonDump = fmt.Sprintf(" firstPod=%s/%s rejected=%v skipped=%v",
						noScale.Pod.Namespace, noScale.Pod.Name, noScale.RejectedNodeGroups, noScale.SkippedNodeGroups)
				}
				t.Fatalf("expected ScaleUp to be called, but it wasn't (triggered=%d remainUnschedulable=%d awaitEval=%d result=%v)%s",
					len(scaleUpProcessor.lastStatus.PodsTriggeredScaleUp),
					len(scaleUpProcessor.lastStatus.PodsRemainUnschedulable),
					len(scaleUpProcessor.lastStatus.PodsAwaitEvaluation),
					scaleUpProcessor.lastStatus.Result,
					reasonDump,
				)
			}

			if len(tc.expectedScaleUps) == 0 && scaleUpProcessor.lastStatus != nil && len(scaleUpProcessor.lastStatus.PodsTriggeredScaleUp) > 0 {
				t.Fatalf("unexpected scale-up triggered: %v", scaleUpProcessor.lastStatus.PodsTriggeredScaleUp)
			}

			if len(tc.expectedScaleDowns) > 0 && len(mocks.onScaleDown.Calls) == 0 {
				if scaleDownProcessor.lastStatus == nil {
					t.Fatalf("expected ScaleDown to be called, but it wasn't (scaleDownStatus was nil)")
				}
				t.Fatalf("expected ScaleDown to be called, but it wasn't (triggered=%d result=%v)",
					len(scaleDownProcessor.lastStatus.UnremovableNodes),
					scaleDownProcessor.lastStatus.Result,
				)
			}

			mock.AssertExpectationsForObjects(t,
				mocks.allPodLister,
				mocks.podDisruptionBudgetLister,
				mocks.daemonSetLister,
				mocks.onScaleUp,
				mocks.onScaleDown,
			)
		})
	}
}

func csiNodeTemplate(driver string, attachCount *int32) func(nodeName string) *storagev1.CSINode {
	return func(nodeName string) *storagev1.CSINode {
		return &storagev1.CSINode{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
			},
			Spec: storagev1.CSINodeSpec{
				Drivers: []storagev1.CSINodeDriver{
					{
						Name:   driver,
						NodeID: nodeName,
						Allocatable: &storagev1.VolumeNodeResources{
							Count: attachCount,
						},
					},
				},
			},
		}
	}
}

func buildCSITestPod(name string, volumeCount int, nodeName string) *apiv1.Pod {
	pod1 := BuildTestPod(name, 6, 100)
	pod1.Namespace = "default"
	pod1.UID = types.UID(name)
	for i := 0; i < volumeCount; i++ {
		claimName := fmt.Sprintf("%s-pvc-%d", name, i)
		pod1.Spec.Volumes = append(pod1.Spec.Volumes, apiv1.Volume{
			Name: fmt.Sprintf("test-volume-%d", i),
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
				},
			},
		})
	}

	if nodeName != "" {
		pod1.Spec.NodeName = nodeName
	} else {
		MarkUnschedulable()(pod1)
	}

	return pod1
}

func pvcsForPods(pods []*apiv1.Pod, storageClass string) []runtime.Object {
	var k8sObjects []runtime.Object
	for _, p := range pods {
		for _, vol := range p.Spec.Volumes {
			if vol.PersistentVolumeClaim == nil {
				continue
			}
			pvcName := vol.PersistentVolumeClaim.ClaimName
			pvc := &apiv1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcName,
					Namespace: p.Namespace,
					UID:       types.UID(p.Namespace + "/" + pvcName),
				},
				Spec: apiv1.PersistentVolumeClaimSpec{
					StorageClassName: ptr.To(storageClass),
				},
			}
			k8sObjects = append(k8sObjects, pvc)
		}
	}
	return k8sObjects
}
