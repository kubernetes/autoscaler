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
	"context"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
	testprovider "sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider/test"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	cacontext "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/bench/scaledown/efficiency/metrics"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/budgets"
	"sigs.k8s.io/cluster-autoscaler/pkg/core/scaledown/planner"
	. "sigs.k8s.io/cluster-autoscaler/pkg/core/test"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors"
	processorstest "sigs.k8s.io/cluster-autoscaler/pkg/processors/test"
	"sigs.k8s.io/cluster-autoscaler/pkg/resourcequotas"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/clustersnapshot"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/options"
	kubeutil "sigs.k8s.io/cluster-autoscaler/pkg/utils/kubernetes"
)

func init() {
	klog.InitFlags(nil)
}

// scaleDownScenario packs initialClusterState, metricsTracker, and optional config.AutoscalingOptions into one struct for individual scaledown run execution.
type scaleDownScenario struct {
	Name                 string
	InitialClusterState  *initialClusterState
	MetricsTracker       *metrics.Tracker
	AutoscalingOptsSetup func(opts *config.AutoscalingOptions)
}

type scaleDownDependencies struct {
	AutoscalingCtx *cacontext.AutoscalingContext
	Processors     *processors.AutoscalingProcessors
	Planner        *planner.Planner
	Actuator       *fakeActuator
	MetricsTracker *metrics.Tracker
}

func defaultAutoscalingOptions() config.AutoscalingOptions {
	return config.AutoscalingOptions{
		ScaleDownEnabled:           true,
		MaxScaleDownParallelism:    1,
		MaxDrainParallelism:        1,
		ScaleDownSimulationTimeout: 10 * time.Second,
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 1.0,
			ScaleDownUnneededTime:         0,
		},
	}
}

func buildScaleDownDependencies(b *testing.B, cs *initialClusterState, autoscalingOpts config.AutoscalingOptions) scaleDownDependencies {
	b.Helper()
	var allNodes []*apiv1.Node
	var allPods []*apiv1.Pod

	provider := testprovider.NewTestCloudProviderBuilder().Build()

	for _, ng := range cs.NodeGroups {
		allNodes = append(allNodes, ng.Nodes...)
		allPods = append(allPods, ng.Pods...)
		provider.AddNodeGroupWithCustomOptions(ng.Name, ng.MinSize, ng.MaxSize, len(ng.Nodes), ng.NodeGroupOptions)
		for _, node := range ng.Nodes {
			provider.AddNode(ng.Name, node)
		}
	}

	procs, templateRegistry := processorstest.NewTestProcessors(autoscalingOpts)
	registry := kubeutil.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	autoscalingCtx, err := NewScaleTestAutoscalingContext(autoscalingOpts, &fake.Clientset{}, registry, provider, nil, nil, templateRegistry)
	if err != nil {
		b.Fatalf("failed to instantiate AutoscalingContext: %v", err)
	}
	clustersnapshot.InitializeClusterSnapshotOrDie(b, autoscalingCtx.ClusterSnapshot, allNodes, allPods)

	a := NewFakeActuator(autoscalingCtx, &fakeActuationStatus{})
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{CustomResourcesProcessor: procs.CustomResourcesProcessor, QuotaProvider: resourcequotas.NewCloudMinProvider(provider)})
	p := planner.New(&autoscalingCtx, procs, options.NodeDeleteOptions{}, nil, factory)

	return scaleDownDependencies{
		AutoscalingCtx: &autoscalingCtx,
		Processors:     procs,
		Planner:        p,
		Actuator:       a,
	}
}

// Crucial to pass -benchtime=1x to let the Loop know to run scaledown strategy exactly once.
func (scenario scaleDownScenario) run(b *testing.B) {
	opts := defaultAutoscalingOptions()
	if scenario.AutoscalingOptsSetup != nil {
		scenario.AutoscalingOptsSetup(&opts)
	}
	mt := scenario.MetricsTracker
	sdDeps := buildScaleDownDependencies(b, scenario.InitialClusterState, opts)

	nodeInfos, err := sdDeps.AutoscalingCtx.ClusterSnapshot.ListNodeInfos()
	if err != nil {
		b.Fatalf("failed to list node infos from snapshot: %s", err.Error())
	}

	mt.ComputeMetrics(metrics.NewClusterState(nil, nodeInfos, time.Now()), b)

	var scaleDownCandidates []*apiv1.Node
	var podDestinations []*apiv1.Node

	for scaleDownLoop := 0; ; scaleDownLoop++ {
		infos, err := sdDeps.AutoscalingCtx.ClusterSnapshot.ListNodeInfos()
		if err != nil {
			b.Fatalf("error retrieving nodes from cluster snapshot, err %s", err.Error())
		}

		var nodes []*apiv1.Node
		for _, nodeInfo := range infos {
			nodes = append(nodes, nodeInfo.Node())
		}

		if sdDeps.Processors == nil || sdDeps.Processors.ScaleDownNodeProcessor == nil {
			scaleDownCandidates = nodes
			podDestinations = nodes
		} else {
			scaleDownCandidates, err = sdDeps.Processors.ScaleDownNodeProcessor.GetScaleDownCandidates(context.TODO(), sdDeps.AutoscalingCtx, nodes)
			if err != nil {
				b.Fatalf("error getting scale-down candidates: %s", err.Error())
			}
			podDestinations, err = sdDeps.Processors.ScaleDownNodeProcessor.GetPodDestinationCandidates(sdDeps.AutoscalingCtx, nodes)
			if err != nil {
				b.Fatalf("error getting pod destination candidates: %s", err.Error())
			}
		}

		currentTime := time.Now()
		err = sdDeps.Planner.UpdateClusterState(context.TODO(), podDestinations, scaleDownCandidates, sdDeps.Actuator.actuationStatus, currentTime)
		if err != nil {
			b.Fatalf("error updating cluster state: %s", err.Error())
		}

		empty, needDrain := sdDeps.Planner.NodesToDelete(context.TODO(), currentTime)

		budgetProcessor := budgets.NewScaleDownBudgetProcessor(sdDeps.AutoscalingCtx)
		emptyGrouped, drainGrouped := budgetProcessor.CropNodes(context.TODO(), sdDeps.Actuator.actuationStatus, empty, needDrain)

		empty = []*apiv1.Node{}
		for _, group := range emptyGrouped {
			empty = append(empty, group.Nodes...)
		}

		needDrain = []*apiv1.Node{}
		for _, group := range drainGrouped {
			needDrain = append(needDrain, group.Nodes...)
		}

		// No more nodes to delete, final state reached.
		if len(empty) == 0 && len(needDrain) == 0 {
			b.ReportMetric(float64(scaleDownLoop), "loops")
			break
		}
		_, scaledDownNodes, typedErr := sdDeps.Actuator.startDeletion(empty, needDrain)
		if typedErr != nil {
			b.Fatalf("error processing nodes in fake actuator, err %s", typedErr.Error())
		}
		ni, err := sdDeps.AutoscalingCtx.ClusterSnapshot.ListNodeInfos()
		if err != nil {
			b.Fatalf("error listing node infos: %s", err.Error())
		}

		mt.ComputeMetrics(metrics.NewClusterState(scaledDownNodes, ni, currentTime), b)
	}
	mt.ReportToBenchmark(b)
}

func BenchmarkScaledownEfficiency(b *testing.B) {
	s := scaleDownScenario{
		Name:                "three ng, basic setup",
		InitialClusterState: differentNodeSizesMemIrrelevant(),
		MetricsTracker: metrics.NewTracker(
			metrics.NewResourceUtilizationMetric(apiv1.ResourceCPU),
			metrics.NewResourceUtilizationMetric(apiv1.ResourceMemory),
		),
		AutoscalingOptsSetup: func(opts *config.AutoscalingOptions) {
			opts.NodeGroupDefaults.ScaleDownUtilizationThreshold = 0.75
		},
	}
	s.run(b)
}
