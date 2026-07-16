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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup/equivalence"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/random"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestKarpenterSimulatorCorrectness(t *testing.T) {
	// Setup pods
	pod1 := BuildTestPod("p1", 100, 100)
	pod2 := BuildTestPod("p2", 200, 200)
	unschedulablePods := []*apiv1.Pod{pod1, pod2}
	podEquivalenceGroups := equivalence.BuildPodGroups(unschedulablePods)

	// Setup node groups
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	nodeGroups := provider.NodeGroups()

	// Setup node infos
	n1Template := BuildTestNode("ng1-template", 1000, 1000)
	n1Template.Labels["topology.kubernetes.io/zone"] = "us-east-1a"
	n1Template.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"
	n2Template := BuildTestNode("ng2-template", 1000, 1000)
	n2Template.Labels["topology.kubernetes.io/zone"] = "us-east-1b"
	n2Template.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"
	nodeInfos := map[string]*framework.NodeInfo{
		"ng1": framework.NewTestNodeInfo(n1Template),
		"ng2": framework.NewTestNodeInfo(n2Template),
	}

	// Setup context
	opts := config.AutoscalingOptions{
		PredicateParallelism: 1,
	}
	ctx := &ca_context.AutoscalingContext{
		AutoscalingOptions: opts,
		ClusterSnapshot:    testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:      provider,
		ExpanderStrategy:   random.NewStrategy(),
	}

	// Setup simulator
	orchestrator := &ScaleUpOrchestrator{
		autoscalingCtx: ctx,
		processors:     processors.DefaultProcessors(opts),
	}
	defaultSim := NewDefaultSimulator(orchestrator)
	karpenterSim := NewKarpenterSimulator(defaultSim, &DefaultKarpenterConverter{}, orchestrator.processors.NodeGroupSetProcessor)

	// Setup tracker
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Run simulation
	options, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.Len(t, options, 1, "Should have exactly one composite option")

	opt := options[0]
	assert.IsType(t, &CompositeNodeGroup{}, opt.NodeGroup)
	composite := opt.NodeGroup.(*CompositeNodeGroup)
	assert.NotEmpty(t, composite.Plan)

	assert.Equal(t, 2, len(opt.Pods), "All pods should be scheduled")
	assert.True(t, opt.NodeCount > 0)
	assert.NotEmpty(t, schedulable, "Schedulable pods should be recorded")
}

type chooseSpecificStrategy struct {
	targetGroup string
}

func (s *chooseSpecificStrategy) BestOption(options []expander.Option, nodeInfo map[string]*framework.NodeInfo) *expander.Option {
	for _, opt := range options {
		if opt.NodeGroup.Id() == s.targetGroup {
			return &opt
		}
	}
	if len(options) > 0 {
		return &options[0]
	}
	return nil
}

func TestKarpenterSimulatorLabelRequirements(t *testing.T) {
	testCases := []struct {
		name              string
		nodeGroups        map[string]map[string]string // ngName -> labels
		podNodeSelector   map[string]string
		expectedNodeGroup string // Empty if should fail
		expander          expander.Strategy
	}{
		{
			name: "pod requesting custom label present on one node group",
			nodeGroups: map[string]map[string]string{
				"ng-custom": {"my-label": "foo", "topology.kubernetes.io/zone": "zone-a"},
				"ng-other":  {"topology.kubernetes.io/zone": "zone-b"},
			},
			podNodeSelector:   map[string]string{"my-label": "foo"},
			expectedNodeGroup: "ng-custom",
		},
		{
			name: "pod requesting custom label NOT present on any node group",
			nodeGroups: map[string]map[string]string{
				"ng1": {"my-label": "foo"},
			},
			podNodeSelector:   map[string]string{"my-label": "bar"},
			expectedNodeGroup: "",
		},
		{
			name: "pod requesting well-known label (zone) present on one node group",
			nodeGroups: map[string]map[string]string{
				"ng-zone-a": {"topology.kubernetes.io/zone": "zone-a"},
				"ng-zone-b": {"topology.kubernetes.io/zone": "zone-b"},
			},
			podNodeSelector:   map[string]string{"topology.kubernetes.io/zone": "zone-a"},
			expectedNodeGroup: "ng-zone-a",
		},
		{
			name: "pod requesting restricted label (hostname) - should NOT match (new nodes don't have hostname yet)",
			nodeGroups: map[string]map[string]string{
				"ng1": {"kubernetes.io/hostname": "host-1"},
				"ng2": {"kubernetes.io/hostname": "host-2"},
			},
			podNodeSelector:   map[string]string{"kubernetes.io/hostname": "host-1"},
			expectedNodeGroup: "",
		},
		{
			name: "pod requesting label in restricted domain (karpenter.sh/)",
			nodeGroups: map[string]map[string]string{
				"ng-restricted": {"karpenter.sh/custom": "value", "topology.kubernetes.io/zone": "zone-a"},
				"ng-other":      {"topology.kubernetes.io/zone": "zone-b"},
			},
			podNodeSelector:   map[string]string{"karpenter.sh/custom": "value"},
			expectedNodeGroup: "ng-restricted",
		},
		{
			name: "pod requesting multiple labels",
			nodeGroups: map[string]map[string]string{
				"ng1": {"label1": "v1", "label2": "v2"},
				"ng2": {"label1": "v1", "label2": "wrong"},
			},
			podNodeSelector:   map[string]string{"label1": "v1", "label2": "v2"},
			expectedNodeGroup: "ng1",
		},
		{
			name: "pod not requesting custom label - should allow scaling up group lacking it",
			nodeGroups: map[string]map[string]string{
				"ng-custom":   {"my-label": "foo", "topology.kubernetes.io/zone": "zone-a", "node.kubernetes.io/instance-type": "e2-standard-2"},
				"ng-standard": {"topology.kubernetes.io/zone": "zone-a", "node.kubernetes.io/instance-type": "e2-standard-2"},
			},
			podNodeSelector:   map[string]string{},
			expectedNodeGroup: "ng-standard",
			expander:          &chooseSpecificStrategy{targetGroup: "ng-standard"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup node groups
			provider := testprovider.NewTestCloudProviderBuilder().Build()
			nodeInfos := make(map[string]*framework.NodeInfo)
			var ngs []cloudprovider.NodeGroup

			for name, labels := range tc.nodeGroups {
				provider.AddNodeGroup(name, 1, 10, 1)
				ng := provider.GetNodeGroup(name)
				ngs = append(ngs, ng)

				node := BuildTestNode(name+"-template", 1000, 1000)
				for k, v := range labels {
					node.Labels[k] = v
				}
				nodeInfos[name] = framework.NewTestNodeInfo(node)
			}

			// Setup pod
			pod := BuildTestPod("p1", 100, 100, WithLabels(tc.podNodeSelector))
			pod.Spec.NodeSelector = tc.podNodeSelector
			unschedulablePods := []*apiv1.Pod{pod}
			podEquivalenceGroups := equivalence.BuildPodGroups(unschedulablePods)

			// Setup context & simulator
			opts := config.AutoscalingOptions{PredicateParallelism: 1}
			strategy := tc.expander
			if strategy == nil {
				strategy = random.NewStrategy()
			}
			ctx := &ca_context.AutoscalingContext{
				AutoscalingOptions: opts,
				ClusterSnapshot:    testsnapshot.NewTestSnapshotOrDie(t),
				CloudProvider:      provider,
				ExpanderStrategy:   strategy,
			}
			orchestrator := &ScaleUpOrchestrator{
				autoscalingCtx: ctx,
				processors:     processors.DefaultProcessors(opts),
			}
			karpenterSim := NewKarpenterSimulator(NewDefaultSimulator(orchestrator), &DefaultKarpenterConverter{}, orchestrator.processors.NodeGroupSetProcessor)

			// Setup tracker
			factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
				CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
				QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
			})
			tracker, _ := factory.NewQuotasTracker(ctx, nil)

			// Run simulation
			options, _, _, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, ngs, nodeInfos, tracker, time.Now(), false)

			// Verify
			assert.NoError(t, err)
			if tc.expectedNodeGroup == "" {
				assert.Empty(t, options)
			} else {
				assert.Len(t, options, 1)
				opt := options[0]
				assert.IsType(t, &CompositeNodeGroup{}, opt.NodeGroup)
				composite := opt.NodeGroup.(*CompositeNodeGroup)
				found := false
				for _, sui := range composite.Plan {
					if sui.Group.Id() == tc.expectedNodeGroup {
						found = true
						assert.True(t, sui.NewSize > sui.CurrentSize)
					}
				}
				assert.True(t, found, "Expected NodeGroup %s not found in plan", tc.expectedNodeGroup)
			}
		})
	}
}

func TestKarpenterSimulatorSmallestInstanceTypeResolution(t *testing.T) {
	// Setup pods
	pod1 := BuildTestPod("p1", 100, 100)
	unschedulablePods := []*apiv1.Pod{pod1}
	podEquivalenceGroups := equivalence.BuildPodGroups(unschedulablePods)

	// Setup node groups of different sizes
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-standard-2", 1, 10, 1) // e2-standard-2
	provider.AddNodeGroup("ng-standard-16", 1, 10, 1) // e2-standard-16
	nodeGroups := provider.NodeGroups()

	// Setup templates where ng-standard-2 has e2-standard-2 and ng-standard-16 has e2-standard-16
	nStandard2Template := BuildTestNode("ng-standard-2-template", 2000, 8000) // e2-standard-2 (~2 vCPU, 8 GiB)
	nStandard2Template.Labels["topology.kubernetes.io/zone"] = "us-east-1a"
	nStandard2Template.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"

	nStandard16Template := BuildTestNode("ng-standard-16-template", 16000, 64000) // e2-standard-16 (~16 vCPU, 64 GiB)
	nStandard16Template.Labels["topology.kubernetes.io/zone"] = "us-east-1a"
	nStandard16Template.Labels["node.kubernetes.io/instance-type"] = "e2-standard-16"

	nodeInfos := map[string]*framework.NodeInfo{
		"ng-standard-2":   framework.NewTestNodeInfo(nStandard2Template),
		"ng-standard-16": framework.NewTestNodeInfo(nStandard16Template),
	}

	// Setup context
	opts := config.AutoscalingOptions{
		PredicateParallelism: 1,
	}
	ctx := &ca_context.AutoscalingContext{
		AutoscalingOptions: opts,
		ClusterSnapshot:    testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:      provider,
		ExpanderStrategy:   random.NewStrategy(),
	}

	// Setup simulator
	orchestrator := &ScaleUpOrchestrator{
		autoscalingCtx: ctx,
		processors:     processors.DefaultProcessors(opts),
	}
	defaultSim := NewDefaultSimulator(orchestrator)
	karpenterSim := NewKarpenterSimulator(defaultSim, &DefaultKarpenterConverter{}, orchestrator.processors.NodeGroupSetProcessor)

	// Setup tracker
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Run simulation
	options, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.Len(t, options, 1, "Should have exactly one composite option")

	opt := options[0]
	assert.IsType(t, &CompositeNodeGroup{}, opt.NodeGroup)
	composite := opt.NodeGroup.(*CompositeNodeGroup)

	// We expect the solver to select the cheaper/smaller e2-standard-2 (ng-standard-2) physical node group rather than e2-standard-16.
	foundStandard2 := false
	foundStandard16 := false
	for _, sui := range composite.Plan {
		t.Logf("PLAN ITEM: Group=%s, Delta=%d", sui.Group.Id(), sui.NewSize-sui.CurrentSize)
		if sui.Group.Id() == "ng-standard-2" {
			foundStandard2 = true
			assert.True(t, sui.NewSize > sui.CurrentSize, "Should scale up ng-standard-2")
		}
		if sui.Group.Id() == "ng-standard-16" {
			foundStandard16 = true
		}
	}

	assert.True(t, foundStandard2, "Should have scaled up ng-standard-2")
	assert.False(t, foundStandard16, "Should NOT have scaled up ng-standard-16 (cost-optimization violation)")
	assert.NotEmpty(t, schedulable)
}

func TestKarpenterSimulatorMultiZoneReliabilityBalancing(t *testing.T) {
	// Setup pods - we want to trigger a scale-up of 30 nodes.
	var unschedulablePods []*apiv1.Pod
	for i := 0; i < 30; i++ {
		unschedulablePods = append(unschedulablePods, BuildTestPod(fmt.Sprintf("p%d", i), 500, 500))
	}
	podEquivalenceGroups := equivalence.BuildPodGroups(unschedulablePods)

	// Setup three node groups representing Zone-A, Zone-B, Zone-C
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-zone-a", 1, 100, 1)
	provider.AddNodeGroup("ng-zone-b", 1, 100, 1)
	provider.AddNodeGroup("ng-zone-c", 1, 100, 1)
	nodeGroups := provider.NodeGroups()

	// Setup node templates with identical taints (none), identical resources, but different zones.
	nATemplate := BuildTestNode("ng-zone-a-template", 1000, 1000)
	nATemplate.Labels["topology.kubernetes.io/zone"] = "zone-a"
	nATemplate.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"

	nBTemplate := BuildTestNode("ng-zone-b-template", 1000, 1000)
	nBTemplate.Labels["topology.kubernetes.io/zone"] = "zone-b"
	nBTemplate.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"

	nCTemplate := BuildTestNode("ng-zone-c-template", 1000, 1000)
	nCTemplate.Labels["topology.kubernetes.io/zone"] = "zone-c"
	nCTemplate.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"

	nodeInfos := map[string]*framework.NodeInfo{
		"ng-zone-a": framework.NewTestNodeInfo(nATemplate),
		"ng-zone-b": framework.NewTestNodeInfo(nBTemplate),
		"ng-zone-c": framework.NewTestNodeInfo(nCTemplate),
	}

	// Setup context
	opts := config.AutoscalingOptions{
		PredicateParallelism: 1,
	}
	ctx := &ca_context.AutoscalingContext{
		AutoscalingOptions: opts,
		ClusterSnapshot:    testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:      provider,
		ExpanderStrategy:   random.NewStrategy(),
	}

	// Setup simulator
	orchestrator := &ScaleUpOrchestrator{
		autoscalingCtx: ctx,
		processors:     processors.DefaultProcessors(opts),
	}
	defaultSim := NewDefaultSimulator(orchestrator)
	karpenterSim := NewKarpenterSimulator(defaultSim, &DefaultKarpenterConverter{}, orchestrator.processors.NodeGroupSetProcessor)

	// Setup tracker
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Run simulation
	options, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.Len(t, options, 1, "Should have exactly one composite option representing scaled-up capacity")

	opt := options[0]
	assert.IsType(t, &CompositeNodeGroup{}, opt.NodeGroup)
	composite := opt.NodeGroup.(*CompositeNodeGroup)

	// Verify that the final plan contains balanced allocation (e.g., 5 nodes for Zone-A, 5 for Zone-B, 5 for Zone-C)
	assert.Len(t, composite.Plan, 3, "Plan should include all 3 node groups for balancing")

	totalNodes := 0
	for _, sui := range composite.Plan {
		delta := sui.NewSize - sui.CurrentSize
		totalNodes += delta
		assert.Equal(t, 5, delta, "Each zone should receive exactly 5 nodes to preserve zonal reliability")
	}
	assert.Equal(t, 15, totalNodes, "Should have provisioned exactly 15 nodes in total")
	assert.NotEmpty(t, schedulable)
}

func TestKarpenterSimulatorDaemonSets(t *testing.T) {
	// Setup 2 node groups with same physical IT (e2-standard-2) but different DS compatibility.
	// ng-ds has a DS requiring 200 CPU.
	// ng-nodspod has no DS.
	// Pod requires 900 CPU. It can only fit on ng-nodspod.
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-ds", 1, 10, 1)
	provider.AddNodeGroup("ng-nods", 1, 10, 1)
	nodeGroups := provider.NodeGroups()

	nodeInfos := make(map[string]*framework.NodeInfo)

	// ng-ds template node
	nodeDS := BuildTestNode("ng-ds-template", 1000, 1000)
	nodeDS.Labels[apiv1.LabelInstanceTypeStable] = "e2-standard-2"
	nodeDS.Labels["ng"] = "ng-ds"
	niDS := framework.NewNodeInfo(nodeDS, nil)
	// Add DS pod to ng-ds
	dsPod := BuildTestPod("ds-pod", 200, 0)
	dsPod.OwnerReferences = []metav1.OwnerReference{
		{
			Kind:       "DaemonSet",
			Name:       "ds1",
			UID:        "ds1-uid",
			APIVersion: "apps/v1",
		},
	}
	dsPod.Spec.NodeSelector = map[string]string{"ng": "ng-ds"}
	niDS.AddPod(framework.NewPodInfo(dsPod, nil))
	nodeInfos["ng-ds"] = niDS

	// ng-nods template node
	nodeNoDS := BuildTestNode("ng-nods-template", 1000, 1000)
	nodeNoDS.Labels[apiv1.LabelInstanceTypeStable] = "e2-standard-2"
	nodeNoDS.Labels["ng"] = "ng-nods"
	niNoDS := framework.NewNodeInfo(nodeNoDS, nil)
	nodeInfos["ng-nods"] = niNoDS

	// Pod requiring 900 CPU
	pod := BuildTestPod("pod1", 900, 0)
	unschedulablePods := []*apiv1.Pod{pod}

	// Setup context
	opts := config.AutoscalingOptions{
		KarpenterSimulatorEnabled: true,
	}
	ctx := &ca_context.AutoscalingContext{
		AutoscalingOptions: opts,
		ClusterSnapshot:    testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:      provider,
		ExpanderStrategy:   random.NewStrategy(),
	}

	// Setup simulator
	converter := &DefaultKarpenterConverter{}
	processorsInstance := processors.DefaultProcessors(opts)
	karpenterSim := NewKarpenterSimulator(nil, converter, processorsInstance.NodeGroupSetProcessor)

	// Setup tracker
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Pod equivalence groups
	podEquivalenceGroups := []*equivalence.PodGroup{
		{
			Pods: []*apiv1.Pod{pod},
		},
	}

	// Run simulation
	options, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.Len(t, options, 1)

	opt := options[0]
	assert.IsType(t, &CompositeNodeGroup{}, opt.NodeGroup)
	composite := opt.NodeGroup.(*CompositeNodeGroup)

	// Should have chosen ng-nods because ng-ds has DS overhead and pod won't fit
	assert.Len(t, composite.Plan, 1)
	assert.Equal(t, "ng-nods", composite.Plan[0].Group.Id())
	assert.Equal(t, 1, composite.Plan[0].NewSize-composite.Plan[0].CurrentSize)
	assert.NotEmpty(t, schedulable)
}

func TestKarpenterSimulatorClusteringWithDifferentLabels(t *testing.T) {
	// Setup 2 node groups with same physical IT (e2-standard-2) and same DS (none),
	// but different custom labels.
	// ng1 has label=val1.
	// ng2 has label=val2.
	// Pod requires label=val2.
	// They should be clustered into the SAME InstanceType (because same IT and same DS).
	// But they should be routed correctly based on offering requirements.
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	nodeGroups := provider.NodeGroups()

	nodeInfos := make(map[string]*framework.NodeInfo)

	// ng1 template node
	node1 := BuildTestNode("ng1-template", 1000, 1000)
	node1.Labels[apiv1.LabelInstanceTypeStable] = "e2-standard-2"
	node1.Labels["custom-label"] = "val1"
	nodeInfos["ng1"] = framework.NewNodeInfo(node1, nil)

	// ng2 template node
	node2 := BuildTestNode("ng2-template", 1000, 1000)
	node2.Labels[apiv1.LabelInstanceTypeStable] = "e2-standard-2"
	node2.Labels["custom-label"] = "val2"
	nodeInfos["ng2"] = framework.NewNodeInfo(node2, nil)

	// Pod requiring custom-label=val2
	pod := BuildTestPod("pod1", 100, 0)
	pod.Spec.NodeSelector = map[string]string{"custom-label": "val2"}
	unschedulablePods := []*apiv1.Pod{pod}

	// Setup context
	opts := config.AutoscalingOptions{
		KarpenterSimulatorEnabled: true,
	}
	ctx := &ca_context.AutoscalingContext{
		AutoscalingOptions: opts,
		ClusterSnapshot:    testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:      provider,
		ExpanderStrategy:   random.NewStrategy(),
	}

	// Setup simulator
	converter := &DefaultKarpenterConverter{}
	processorsInstance := processors.DefaultProcessors(opts)
	karpenterSim := NewKarpenterSimulator(nil, converter, processorsInstance.NodeGroupSetProcessor)

	// Setup tracker
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	// Pod equivalence groups
	podEquivalenceGroups := []*equivalence.PodGroup{
		{
			Pods: []*apiv1.Pod{pod},
		},
	}

	// Run simulation
	options, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.Len(t, options, 1)

	opt := options[0]
	assert.IsType(t, &CompositeNodeGroup{}, opt.NodeGroup)
	composite := opt.NodeGroup.(*CompositeNodeGroup)

	// Should have chosen ng2
	assert.Len(t, composite.Plan, 1)
	assert.Equal(t, "ng2", composite.Plan[0].Group.Id())
	assert.Equal(t, 1, composite.Plan[0].NewSize-composite.Plan[0].CurrentSize)
	assert.NotEmpty(t, schedulable)
}

func TestKarpenterSimulatorQuotaCappingAndBalancing(t *testing.T) {
	// Setup pods - we want to trigger a scale-up of 30 nodes (each fits 2 pods, so 15 nodes needed).
	var unschedulablePods []*apiv1.Pod
	for i := 0; i < 30; i++ {
		unschedulablePods = append(unschedulablePods, BuildTestPod(fmt.Sprintf("p%d", i), 500, 500))
	}
	podEquivalenceGroups := equivalence.BuildPodGroups(unschedulablePods)

	// Setup three node groups representing Zone-A, Zone-B, Zone-C
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng-zone-a", 1, 100, 1)
	provider.AddNodeGroup("ng-zone-b", 1, 100, 1)
	provider.AddNodeGroup("ng-zone-c", 1, 100, 1)
	nodeGroups := provider.NodeGroups()

	// Setup node templates and existing nodes
	nATemplate := BuildTestNode("ng-zone-a-template", 1000, 1000)
	nATemplate.Labels["topology.kubernetes.io/zone"] = "zone-a"
	nATemplate.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"
	nATemplate.Labels[nodeGroupLabel] = "ng-zone-a"
	nodeA := BuildTestNode("ng-zone-a-node-1", 1000, 1000)
	nodeA.Labels["topology.kubernetes.io/zone"] = "zone-a"
	nodeA.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"
	nodeA.Labels[nodeGroupLabel] = "ng-zone-a"
	provider.AddNode("ng-zone-a", nodeA)

	nBTemplate := BuildTestNode("ng-zone-b-template", 1000, 1000)
	nBTemplate.Labels["topology.kubernetes.io/zone"] = "zone-b"
	nBTemplate.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"
	nBTemplate.Labels[nodeGroupLabel] = "ng-zone-b"
	nodeB := BuildTestNode("ng-zone-b-node-1", 1000, 1000)
	nodeB.Labels["topology.kubernetes.io/zone"] = "zone-b"
	nodeB.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"
	nodeB.Labels[nodeGroupLabel] = "ng-zone-b"
	provider.AddNode("ng-zone-b", nodeB)

	nCTemplate := BuildTestNode("ng-zone-c-template", 1000, 1000)
	nCTemplate.Labels["topology.kubernetes.io/zone"] = "zone-c"
	nCTemplate.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"
	nCTemplate.Labels[nodeGroupLabel] = "ng-zone-c"
	nodeC := BuildTestNode("ng-zone-c-node-1", 1000, 1000)
	nodeC.Labels["topology.kubernetes.io/zone"] = "zone-c"
	nodeC.Labels["node.kubernetes.io/instance-type"] = "e2-standard-2"
	nodeC.Labels[nodeGroupLabel] = "ng-zone-c"
	provider.AddNode("ng-zone-c", nodeC)

	existingNodes := []*apiv1.Node{nodeA, nodeB, nodeC}

	nodeInfos := map[string]*framework.NodeInfo{
		"ng-zone-a": framework.NewTestNodeInfo(nATemplate),
		"ng-zone-b": framework.NewTestNodeInfo(nBTemplate),
		"ng-zone-c": framework.NewTestNodeInfo(nCTemplate),
	}

	// Setup context
	opts := config.AutoscalingOptions{
		PredicateParallelism: 1,
	}
	ctx := &ca_context.AutoscalingContext{
		AutoscalingOptions: opts,
		ClusterSnapshot:    testsnapshot.NewTestSnapshotOrDie(t),
		CloudProvider:      provider,
		ExpanderStrategy:   random.NewStrategy(),
	}
	err := ctx.ClusterSnapshot.SetClusterState(existingNodes, nil, nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	// Setup simulator
	orchestrator := &ScaleUpOrchestrator{
		autoscalingCtx: ctx,
		processors:     processors.DefaultProcessors(opts),
	}
	// Override NodeGroupSetProcessor to allow balancing between these three groups
	orchestrator.processors.NodeGroupSetProcessor = nodegroupset.NewDefaultNodeGroupSetProcessor([]string{"topology.kubernetes.io/zone", nodeGroupLabel}, config.NodeGroupDifferenceRatios{})

	defaultSim := NewDefaultSimulator(orchestrator)
	karpenterSim := NewKarpenterSimulator(defaultSim, &DefaultKarpenterConverter{}, orchestrator.processors.NodeGroupSetProcessor)

	// Setup tracker with a shared quota of 10 nodes total across all three groups
	quotas := []resourcequotas.Quota{
		&resourcequotas.FakeQuota{
			Name:        "shared-node-limit",
			AppliesToFn: matchNodeGroups([]string{"ng-zone-a", "ng-zone-b", "ng-zone-c"}),
			LimitsVal:   map[string]int64{"nodes": 10},
		},
	}

	cloudQuotasProvider := resourcequotas.NewCloudQuotasProvider(provider)
	fakeQuotasProvider := resourcequotas.NewFakeProvider(quotas)
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		QuotaProvider:            resourcequotas.NewCombinedQuotasProvider([]resourcequotas.Provider{cloudQuotasProvider, fakeQuotasProvider}),
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, existingNodes)

	// Run simulation
	options, skipped, schedulable, err := karpenterSim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, existingNodes, nodeGroups, nodeInfos, tracker, time.Now(), false)

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, skipped)
	assert.Len(t, options, 1, "Should have exactly one composite option representing scaled-up capacity")

	opt := options[0]
	assert.IsType(t, &CompositeNodeGroup{}, opt.NodeGroup)
	composite := opt.NodeGroup.(*CompositeNodeGroup)

	// Verify that the final plan contains balanced allocation capped by the quota.
	assert.Len(t, composite.Plan, 3, "Plan should include all 3 node groups for balancing")

	totalNodes := 0
	for _, sui := range composite.Plan {
		delta := sui.NewSize - sui.CurrentSize
		totalNodes += delta
		assert.True(t, delta == 2 || delta == 3, fmt.Sprintf("Group %s delta %d should be either 2 or 3", sui.Group.Id(), delta))
	}
	assert.Equal(t, 7, totalNodes, "Should have provisioned exactly 7 nodes (capped by quota)")

	assert.Len(t, opt.Pods, 14, "Should have scheduled exactly 14 pods in the allowed 7 nodes")
	assert.NotEmpty(t, schedulable)
}



