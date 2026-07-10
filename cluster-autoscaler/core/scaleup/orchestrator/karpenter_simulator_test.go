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

package orchestrator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/equivalence"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"

	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/annotations"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"

	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

func createKarpenterTestPod(name string, cpu, mem int64) *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{{
				Name: "c",
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpu, resource.DecimalSI),
						apiv1.ResourceMemory: *resource.NewQuantity(mem*1024*1024, resource.BinarySI),
					},
				},
			}},
		},
	}
}

func TestKarpenterSimulatorBasic(t *testing.T) {
	ctx := &ca_context.AutoscalingContext{
		ClusterSnapshot: testsnapshot.NewTestSnapshotOrDie(t),
	}

	// Setup cloud provider with node groups
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 1, 10, 1)

	ctx.CloudProvider = provider

	nodeGroups := provider.NodeGroups()

	// Build template NodeInfos
	node1 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-ng1-template",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	nodeInfos := map[string]*framework.NodeInfo{
		"ng1": framework.NewNodeInfo(node1, nil),
	}

	// Setup converter
	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	processor := nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})

	karpenterSim := NewKarpenterSimulator(nil, converter, processor, true)

	// Build unschedulable pods
	pod1 := createKarpenterTestPod("p1", 500, 1000)
	pod2 := createKarpenterTestPod("p2", 500, 1000)
	unschedulablePods := []*apiv1.Pod{pod1, pod2}

	podEquivalenceGroups := []*equivalence.PodGroup{
		{
			Pods: unschedulablePods,
		},
	}

	// Setup tracker
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Run simulation
	decisions, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.Len(t, decisions, 1, "Should have exactly one decision group")

	options := decisions[0]
	assert.NotEmpty(t, options)
	opt := options[0]
	assert.Equal(t, "ng1", opt.NodeGroup.Id())
	assert.Equal(t, 2, len(opt.Pods), "All pods should be scheduled")
	assert.True(t, opt.NodeCount > 0)
	assert.NotEmpty(t, schedulable, "Schedulable pods should be recorded")
}

func TestKarpenterSimulatorInstanceTypeSelection(t *testing.T) {
	testCases := []struct {
		name              string
		podCPU            int64
		podMem            int64
		expectedNodeGroup string
	}{
		{
			name:              "small pod fits e2-standard-2",
			podCPU:            1000,
			podMem:            2000,
			expectedNodeGroup: "ng-standard-2",
		},
		{
			name:              "large pod requires e2-standard-16",
			podCPU:            5000,
			podMem:            16000,
			expectedNodeGroup: "ng-standard-16",
		},
		{
			name:              "huge pod fits nothing",
			podCPU:            32000,
			podMem:            128000,
			expectedNodeGroup: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &ca_context.AutoscalingContext{
				ClusterSnapshot: testsnapshot.NewTestSnapshotOrDie(t),
			}

			provider := testprovider.NewTestCloudProviderBuilder().Build()
			provider.AddNodeGroup("ng-standard-2", 1, 10, 1)
			provider.AddNodeGroup("ng-standard-16", 1, 10, 1)
			ctx.CloudProvider = provider

			ngs := provider.NodeGroups()

			nodeStandard2 := &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-std-2",
					Labels: map[string]string{
						apiv1.LabelInstanceTypeStable: "e2-standard-2",
						apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
						apiv1.LabelOSStable:           "linux",
					},
				},
				Status: apiv1.NodeStatus{
					Capacity: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("2"),
						apiv1.ResourceMemory: resource.MustParse("8Gi"),
						apiv1.ResourcePods:   resource.MustParse("110"),
					},
					Allocatable: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("2"),
						apiv1.ResourceMemory: resource.MustParse("8Gi"),
						apiv1.ResourcePods:   resource.MustParse("110"),
					},
				},
			}

			nodeStandard16 := &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-std-16",
					Labels: map[string]string{
						apiv1.LabelInstanceTypeStable: "e2-standard-16",
						apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
						apiv1.LabelOSStable:           "linux",
					},
				},
				Status: apiv1.NodeStatus{
					Capacity: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("16"),
						apiv1.ResourceMemory: resource.MustParse("64Gi"),
						apiv1.ResourcePods:   resource.MustParse("110"),
					},
					Allocatable: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("16"),
						apiv1.ResourceMemory: resource.MustParse("64Gi"),
						apiv1.ResourcePods:   resource.MustParse("110"),
					},
				},
			}

			nodeInfos := map[string]*framework.NodeInfo{
				"ng-standard-2":  framework.NewNodeInfo(nodeStandard2, nil),
				"ng-standard-16": framework.NewNodeInfo(nodeStandard16, nil),
			}

			converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
			processor := nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})

			karpenterSim := NewKarpenterSimulator(nil, converter, processor, true)

			pod := createKarpenterTestPod("pod1", tc.podCPU, tc.podMem)
			unschedulablePods := []*apiv1.Pod{pod}
			podEquivalenceGroups := []*equivalence.PodGroup{{Pods: unschedulablePods}}

			factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
				CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
				QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
			})
			tracker, _ := factory.NewQuotasTracker(ctx, nil)

			// Run simulation
			decisions, _, _, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, ngs, nodeInfos, tracker, time.Now(), false)

			// Verify
			assert.NoError(t, err)
			if tc.expectedNodeGroup == "" {
				assert.Empty(t, decisions)
			} else {
				assert.NotEmpty(t, decisions)
				found := false
				for _, opts := range decisions {
					for _, opt := range opts {
						if opt.NodeGroup.Id() == tc.expectedNodeGroup {
							found = true
							assert.True(t, opt.NodeCount > 0)
						}
					}
				}
				assert.True(t, found, "Expected NodeGroup %s not found in decisions", tc.expectedNodeGroup)
			}
		})
	}
}

func TestKarpenterSimulatorDaemonSets(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-ds", 1, 10, 1)
	provider.AddNodeGroup("ng-nods", 1, 10, 1)

	ctx := &ca_context.AutoscalingContext{
		ClusterSnapshot: testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:   provider,
	}

	nodeGroups := provider.NodeGroups()

	// Node for ng-ds with a DS pod taking 1.5 CPU
	nodeDS := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-ds",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
				"ds-target":                   "ng-ds",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}
	dsPod := createKarpenterTestPod("ds-pod", 1500, 1000)
	dsPod.Spec.NodeSelector = map[string]string{"ds-target": "ng-ds"}
	dsPod.OwnerReferences = []metav1.OwnerReference{{Kind: "DaemonSet", Name: "ds"}}
	dsPodInfo := framework.NewPodInfo(dsPod, nil)
	nodeInfoDS := framework.NewNodeInfo(nodeDS, nil, dsPodInfo)

	// Node for ng-nods with no DS
	nodeNoDS := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-nods",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
				"ds-target":                   "ng-nods",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}
	nodeInfoNoDS := framework.NewNodeInfo(nodeNoDS, nil)

	nodeInfos := map[string]*framework.NodeInfo{
		"ng-ds":   nodeInfoDS,
		"ng-nods": nodeInfoNoDS,
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	processor := nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	karpenterSim := NewKarpenterSimulator(nil, converter, processor, true)

	// Pod requiring 1 CPU (1000m). On ng-ds, remaining is 2000m - 1500m = 500m (won't fit). On ng-nods, remaining is 2000m (fits).
	pod := createKarpenterTestPod("pod1", 1000, 1000)
	unschedulablePods := []*apiv1.Pod{pod}
	podEquivalenceGroups := []*equivalence.PodGroup{{Pods: []*apiv1.Pod{pod}}}

	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Run simulation
	decisions, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.NotEmpty(t, decisions)

	foundNoDS := false
	for _, opts := range decisions {
		for _, opt := range opts {
			if opt.NodeGroup.Id() == "ng-nods" {
				foundNoDS = true
				assert.True(t, opt.NodeCount > 0)
			}
		}
	}
	assert.True(t, foundNoDS, "Should have chosen ng-nods because ng-ds has DS overhead and pod won't fit")
	assert.NotEmpty(t, schedulable)
}

func TestKarpenterSimulatorCustomDaemonSetAnnotation(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-custom-ds", 1, 10, 1)
	provider.AddNodeGroup("ng-nods", 1, 10, 1)

	ctx := &ca_context.AutoscalingContext{
		ClusterSnapshot: testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:   provider,
	}

	nodeGroups := provider.NodeGroups()

	// Node for ng-custom-ds with a custom annotated DS pod taking 1.5 CPU
	nodeCustomDS := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-custom-ds",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
				"ds-target":                   "ng-custom-ds",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}
	customDSPod := createKarpenterTestPod("custom-ds-pod", 1500, 1000)
	customDSPod.Spec.NodeSelector = map[string]string{"ds-target": "ng-custom-ds"}
	customDSPod.Annotations = map[string]string{podutils.DaemonSetPodAnnotationKey: "true"}
	customDSPodInfo := framework.NewPodInfo(customDSPod, nil)
	nodeInfoCustomDS := framework.NewNodeInfo(nodeCustomDS, nil, customDSPodInfo)

	// Node for ng-nods with no DS
	nodeNoDS := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-nods",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
				"ds-target":                   "ng-nods",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}
	nodeInfoNoDS := framework.NewNodeInfo(nodeNoDS, nil)

	nodeInfos := map[string]*framework.NodeInfo{
		"ng-custom-ds": nodeInfoCustomDS,
		"ng-nods":      nodeInfoNoDS,
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	processor := nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	karpenterSim := NewKarpenterSimulator(nil, converter, processor, true)

	pod := createKarpenterTestPod("pod1", 1000, 1000)
	unschedulablePods := []*apiv1.Pod{pod}
	podEquivalenceGroups := []*equivalence.PodGroup{{Pods: []*apiv1.Pod{pod}}}

	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	decisions, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.NotEmpty(t, decisions)

	foundNoDS := false
	for _, opts := range decisions {
		for _, opt := range opts {
			if opt.NodeGroup.Id() == "ng-nods" {
				foundNoDS = true
				assert.True(t, opt.NodeCount > 0)
			}
		}
	}
	assert.True(t, foundNoDS, "Should have chosen ng-nods because ng-custom-ds has custom DS annotation overhead")
	assert.NotEmpty(t, schedulable)
}

func TestKarpenterSimulatorClusteringWithDifferentLabels(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)

	ctx := &ca_context.AutoscalingContext{
		ClusterSnapshot: testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:   provider,
	}

	nodeGroups := provider.NodeGroups()

	node1 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-ng1",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
				"custom-label":                "val1",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	node2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-ng2",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
				"custom-label":                "val2",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	nodeInfos := map[string]*framework.NodeInfo{
		"ng1": framework.NewNodeInfo(node1, nil),
		"ng2": framework.NewNodeInfo(node2, nil),
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	processor := nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	karpenterSim := NewKarpenterSimulator(nil, converter, processor, true)

	// Pod requires custom-label=val2
	pod := createKarpenterTestPod("pod1", 500, 1000)
	pod.Spec.NodeSelector = map[string]string{"custom-label": "val2"}
	unschedulablePods := []*apiv1.Pod{pod}
	podEquivalenceGroups := []*equivalence.PodGroup{{Pods: []*apiv1.Pod{pod}}}

	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Run simulation
	decisions, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.NotEmpty(t, decisions)

	foundNg2 := false
	for _, opts := range decisions {
		for _, opt := range opts {
			if opt.NodeGroup.Id() == "ng2" {
				foundNg2 = true
				assert.True(t, opt.NodeCount > 0)
			}
		}
	}
	assert.True(t, foundNg2, "Should have chosen ng2 matching pod nodeSelector")
	assert.NotEmpty(t, schedulable)
}

// mockPricingModel for tests
type mockPricingModel struct{}

func (m *mockPricingModel) NodePrice(node *apiv1.Node, start, end time.Time) (float64, error) {
	switch node.Labels[apiv1.LabelInstanceTypeStable] {
	case "e2-standard-2":
		return 0.067, nil
	case "e2-standard-16":
		return 0.536, nil
	default:
		return 1.0, nil
	}
}

func (m *mockPricingModel) PodPrice(pod *apiv1.Pod, start, end time.Time) (float64, error) {
	return 0.0, nil
}

func TestKarpenterSimulatorSalvoBatchingCoupling(t *testing.T) {
	ctx := &ca_context.AutoscalingContext{
		ClusterSnapshot: testsnapshot.NewTestSnapshotOrDie(t),
	}

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 1, 2000, 1)

	node1 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-ng1-template",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	nodeInfos := map[string]*framework.NodeInfo{
		"ng1": framework.NewNodeInfo(node1, nil),
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	processor := nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})

	// Build 1050 unschedulable pods
	var unschedulablePods []*apiv1.Pod
	var podEquivalenceGroups []*equivalence.PodGroup
	for i := 0; i < 1050; i++ {
		p := createKarpenterTestPod(fmt.Sprintf("p%d", i), 500, 1000)
		unschedulablePods = append(unschedulablePods, p)
		podEquivalenceGroups = append(podEquivalenceGroups, &equivalence.PodGroup{Pods: []*apiv1.Pod{p}})
	}

	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Case 1: Salvo enabled -> Batch capped at 1000 pods
	salvoSim := NewKarpenterSimulator(nil, converter, processor, true)
	decisionsSalvo, _, _, err := salvoSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, provider.NodeGroups(), nodeInfos, tracker, time.Now(), false)
	assert.NoError(t, err)
	assert.NotEmpty(t, decisionsSalvo)
	totalPodsSalvo := 0
	for _, opts := range decisionsSalvo {
		for _, opt := range opts {
			totalPodsSalvo += len(opt.Pods)
		}
	}
	assert.Equal(t, 1000, totalPodsSalvo, "When Salvo is enabled, batch size should be capped at 1000")

	// Case 2: Salvo disabled -> Batching disabled, evaluates all 1050 pods
	noSalvoSim := NewKarpenterSimulator(nil, converter, processor, false)
	decisionsNoSalvo, _, _, err := noSalvoSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, provider.NodeGroups(), nodeInfos, tracker, time.Now(), false)
	assert.NoError(t, err)
	assert.NotEmpty(t, decisionsNoSalvo)
	totalPodsNoSalvo := 0
	for _, opts := range decisionsNoSalvo {
		for _, opt := range opts {
			totalPodsNoSalvo += len(opt.Pods)
		}
	}
	assert.Equal(t, 1050, totalPodsNoSalvo, "When Salvo is disabled, all 1050 pods should be evaluated in single pass")
}

func TestKarpenterSharedPhysicalInstanceTypeAcrossPools(t *testing.T) {
	ctx := &ca_context.AutoscalingContext{
		ClusterSnapshot: testsnapshot.NewTestSnapshotOrDie(t),
	}

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-untainted", 1, 10, 1)
	provider.AddNodeGroup("ng-tainted", 1, 10, 1)
	ctx.CloudProvider = provider

	nodeUntainted := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-untainted",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	nodeTainted := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-tainted",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-2",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
				"dedicated":                   "special",
			},
		},
		Spec: apiv1.NodeSpec{
			Taints: []apiv1.Taint{
				{Key: "dedicated", Value: "special", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	nodeInfos := map[string]*framework.NodeInfo{
		"ng-untainted": framework.NewNodeInfo(nodeUntainted, nil),
		"ng-tainted":   framework.NewNodeInfo(nodeTainted, nil),
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	processor := nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	karpenterSim := NewKarpenterSimulator(nil, converter, processor, true)

	res, err := converter.Convert(provider.NodeGroups(), nodeInfos)
	assert.NoError(t, err)
	assert.Len(t, res.NodePools, 2, "Should create 2 NodePools for untainted and tainted groups")

	for _, np := range res.NodePools {
		its := res.InstanceTypes[np.Name]
		assert.Len(t, its, 1)
		assert.Equal(t, "e2-standard-2", its[0].Name, "Both NodePools must retain genuine physical InstanceType name without synthetic suffixing")
	}

	// Verify 2D mappings
	for _, np := range res.NodePools {
		ngs := res.NodeGroupsFor(np.Name, "e2-standard-2")
		assert.Len(t, ngs, 1)
	}

	assert.Equal(t, res.PoolForNodeGroup("ng-untainted"), res.PoolForNodeGroup("ng-untainted"))
	assert.NotEqual(t, res.PoolForNodeGroup("ng-untainted"), res.PoolForNodeGroup("ng-tainted"))

	// Verify pod requesting physical instance type and matching taint schedules onto ng-tainted
	pod := createKarpenterTestPod("pod-tainted", 500, 1000)
	pod.Spec.NodeSelector = map[string]string{
		apiv1.LabelInstanceTypeStable: "e2-standard-2",
		"dedicated":                   "special",
	}
	pod.Spec.Tolerations = []apiv1.Toleration{{Key: "dedicated", Value: "special", Effect: apiv1.TaintEffectNoSchedule}}

	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	decisions, skipped, schedulable, err := karpenterSim.Simulate(ctx, []*equivalence.PodGroup{{Pods: []*apiv1.Pod{pod}}}, []*apiv1.Pod{pod}, nil, provider.NodeGroups(), nodeInfos, tracker, time.Now(), false)
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.NotEmpty(t, decisions)
	assert.NotEmpty(t, schedulable)

	foundTainted := false
	for _, opts := range decisions {
		for _, opt := range opts {
			if opt.NodeGroup.Id() == "ng-tainted" {
				foundTainted = true
				assert.Equal(t, 1, opt.NodeCount)
			}
		}
	}
	assert.True(t, foundTainted, "Pod with instance-type selector and toleration must schedule onto ng-tainted")
}

func TestKarpenterConverter_IgnoredLabelsAndFiltering(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-filtered", 1, 10, 1)

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-filtered",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable:                    "e2-standard-2",
				apiv1.LabelHostname:                              "node-filtered-host",
				apiv1.LabelTopologyZone:                          "us-central1-a",
				apiv1.LabelTopologyRegion:                        "us-central1",
				apiv1.LabelZoneFailureDomain:                     "us-central1-a",
				apiv1.LabelZoneRegion:                            "us-central1",
				"beta.kubernetes.io/instance-type":               "e2-standard-2",
				"beta.kubernetes.io/arch":                        "amd64",
				"beta.kubernetes.io/os":                          "linux",
				karpenterv1.CapacityTypeLabelKey:                 karpenterv1.CapacityTypeOnDemand,
				"cluster-autoscaler.kubernetes.io/template-node": "true",
				"cloud.google.com/gke-nodepool":                  "filtered-pool",
				"app":                                            "web",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
	}
	nodeInfos := map[string]*framework.NodeInfo{
		"ng-filtered": framework.NewNodeInfo(node, nil),
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, []string{"cloud.google.com/gke-nodepool"})
	res, err := converter.Convert(provider.NodeGroups(), nodeInfos)
	assert.NoError(t, err)
	assert.Len(t, res.NodePools, 1)

	poolName := res.NodePools[0].Name
	its := res.InstanceTypes[poolName]
	assert.Len(t, its, 1)
	it := its[0]

	// Verify ignored labels do not exist in static requirements
	assert.False(t, it.Requirements.Has(apiv1.LabelHostname))
	assert.False(t, it.Requirements.Has("cluster-autoscaler.kubernetes.io/template-node"))
	assert.False(t, it.Requirements.Has("cloud.google.com/gke-nodepool"))
	assert.False(t, it.Requirements.Has("beta.kubernetes.io/arch"))
	assert.False(t, it.Requirements.Has("beta.kubernetes.io/os"))

	// Verify custom label 'app' is present on offering requirements
	assert.Len(t, it.Offerings, 1)
	assert.True(t, it.Offerings[0].Requirements.Has("app"))
}

func TestKarpenterConverter_BetaArchAndOSNormalization(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-arm", 1, 10, 1)

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-arm",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "t2a-standard-1",
				"beta.kubernetes.io/arch":     "arm64",
				"beta.kubernetes.io/os":       "linux",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("1"),
				apiv1.ResourceMemory: resource.MustParse("4Gi"),
			},
		},
	}
	nodeInfos := map[string]*framework.NodeInfo{
		"ng-arm": framework.NewNodeInfo(node, nil),
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	res, err := converter.Convert(provider.NodeGroups(), nodeInfos)
	assert.NoError(t, err)

	poolName := res.NodePools[0].Name
	its := res.InstanceTypes[poolName]
	assert.Len(t, its, 1)
	it := its[0]

	reqArch := it.Requirements.Get(apiv1.LabelArchStable)
	assert.NotNil(t, reqArch)
	assert.Equal(t, []string{"arm64"}, reqArch.Values())

	reqOS := it.Requirements.Get(apiv1.LabelOSStable)
	assert.NotNil(t, reqOS)
	assert.Equal(t, []string{"linux"}, reqOS.Values())
}

func TestKarpenterConverter_NamespaceScopedDaemonSetKeys(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-sys-ds", 1, 10, 1)
	provider.AddNodeGroup("ng-mon-ds", 1, 10, 1)

	nodeSys := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-sys",
			Labels: map[string]string{apiv1.LabelInstanceTypeStable: "e2-standard-2"},
		},
		Status: apiv1.NodeStatus{Capacity: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("2")}},
	}
	isController := true
	dsPodSys := createKarpenterTestPod("fluentd-sys", 500, 500)
	dsPodSys.Namespace = "kube-system"
	dsPodSys.OwnerReferences = []metav1.OwnerReference{{Kind: "DaemonSet", Name: "fluentd", Controller: &isController}}
	nodeInfoSys := framework.NewNodeInfo(nodeSys, nil, framework.NewPodInfo(dsPodSys, nil))

	nodeMon := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-mon",
			Labels: map[string]string{apiv1.LabelInstanceTypeStable: "e2-standard-2"},
		},
		Status: apiv1.NodeStatus{Capacity: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("2")}},
	}
	dsPodMon := createKarpenterTestPod("fluentd-mon", 500, 500)
	dsPodMon.Namespace = "monitoring"
	dsPodMon.OwnerReferences = []metav1.OwnerReference{{Kind: "DaemonSet", Name: "fluentd", Controller: &isController}}
	nodeInfoMon := framework.NewNodeInfo(nodeMon, nil, framework.NewPodInfo(dsPodMon, nil))

	nodeInfos := map[string]*framework.NodeInfo{
		"ng-sys-ds": nodeInfoSys,
		"ng-mon-ds": nodeInfoMon,
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	res, err := converter.Convert(provider.NodeGroups(), nodeInfos)
	assert.NoError(t, err)

	poolName := res.NodePools[0].Name
	its := res.InstanceTypes[poolName]
	assert.Len(t, its, 2, "Distinct namespace-scoped DS keys should separate into distinct instance types")
}

func TestKarpenterSimulator_SalvoCandidatePlacementAndTopologyPreservation(t *testing.T) {
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-salvo", 1, 10, 1)

	templateNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "template-node-for-ng-salvo",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-4",
				apiv1.LabelTopologyZone:       "us-central1-a",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("4"),
				apiv1.ResourceMemory: resource.MustParse("16Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("4"),
				apiv1.ResourceMemory: resource.MustParse("16Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}
	templateNodeInfos := map[string]*framework.NodeInfo{
		"ng-salvo": framework.NewNodeInfo(templateNode, nil),
	}

	// 1. Existing non-Salvo node with a running pod
	existingNonSalvoNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-node-real",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-4",
				apiv1.LabelTopologyZone:       "us-central1-a",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("4"),
				apiv1.ResourceMemory: resource.MustParse("16Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("4"),
				apiv1.ResourceMemory: resource.MustParse("16Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}
	existingPod := createKarpenterTestPod("existing-pod", 100, 100)
	existingPod.Spec.NodeName = "existing-node-real"
	existingPod.Labels = map[string]string{"app": "redis-leader"}

	// 2. Upcoming Salvo node injected from previous Salvo batch (with remaining capacity)
	salvoNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "template-node-for-ng-salvo-0-salvo-0",
			Labels: map[string]string{
				apiv1.LabelInstanceTypeStable: "e2-standard-4",
				apiv1.LabelTopologyZone:       "us-central1-a",
				apiv1.LabelArchStable:         karpenterv1.ArchitectureAmd64,
				apiv1.LabelOSStable:           "linux",
			},
			Annotations: map[string]string{
				annotations.NodeUpcomingAnnotation: "true",
				annotations.NodeSalvoAnnotation:    "true",
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("4"),
				apiv1.ResourceMemory: resource.MustParse("16Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("4"),
				apiv1.ResourceMemory: resource.MustParse("16Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	snapshot := testsnapshot.NewTestSnapshotOrDie(t)
	err := snapshot.AddNodeInfo(framework.NewNodeInfo(existingNonSalvoNode, nil, framework.NewPodInfo(existingPod, nil)))
	assert.NoError(t, err)
	err = snapshot.AddNodeInfo(framework.NewNodeInfo(salvoNode, nil))
	assert.NoError(t, err)

	ctx := &ca_context.AutoscalingContext{
		AutoscalingOptions: config.AutoscalingOptions{
			KarpenterSimulatorEnabled: true,
			SalvoScaleUp:              true,
		},
		CloudProvider:   provider,
		ClusterSnapshot: snapshot,
	}

	converter := NewDefaultKarpenterConverter(&mockPricingModel{}, nil)
	processor := nodegroupset.NewDefaultNodeGroupSetProcessor([]string{}, config.NodeGroupDifferenceRatios{})
	karpenterSim := NewKarpenterSimulator(nil, converter, processor, true)

	// Pending pod fits on existing salvoNode (500m CPU) and has podAffinity to redis-leader (running on non-Salvo node)
	pendingPod := createKarpenterTestPod("pending-pod", 500, 500)
	pendingPod.Spec.Affinity = &apiv1.Affinity{
		PodAffinity: &apiv1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "redis-leader"},
					},
					TopologyKey: apiv1.LabelTopologyZone,
				},
			},
		},
	}

	// scaleUpPod requires 3800m CPU, exceeding salvoNode remaining capacity after pendingPod (3500m), triggering scale up of ng-salvo
	scaleUpPod := createKarpenterTestPod("scaleup-pod", 3800, 1000)

	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	unschedulablePods := []*apiv1.Pod{pendingPod, scaleUpPod}
	podEquivalenceGroups := []*equivalence.PodGroup{{Pods: unschedulablePods}}

	decisions, _, schedulable, err := karpenterSim.Simulate(
		ctx,
		podEquivalenceGroups,
		unschedulablePods,
		nil,
		provider.NodeGroups(),
		templateNodeInfos,
		tracker,
		time.Now(),
		false,
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, decisions, "scaleUpPod should trigger scale-up decision for ng-salvo")
	assert.NotEmpty(t, schedulable)
}
