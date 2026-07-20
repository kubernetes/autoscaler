/*
Copyright 2026 The Kubernetes Authors.

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
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/planner"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	processorstest "k8s.io/autoscaler/cluster-autoscaler/processors/test"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	kubeutil "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// metricsTrackerFactory generates a new instance of a metricsTracker for every iteration of a benchmark scenario.
// It allows individual test scenarios to configure and inject only efficiency metrics they need to evaluate.
type metricsTrackerFactory func() *metricsTracker

// scaleDownScenario packs InitialClusterState, ScaleDownStrategy, Metrics, and optional AutoscalingOpts into one struct for individual scenario execution.
type scaleDownScenario struct {
	Name                string
	InitialClusterState clusterStateFactory
	Metrics             metricsTrackerFactory
	ScaleDownStrategy   scaleDownStrategyFactory
	AutoscalingOpts     func(opts *config.AutoscalingOptions)
}

func defaultAutoscalingOptions() config.AutoscalingOptions {
	return config.AutoscalingOptions{
		ScaleDownEnabled:           true,
		MaxScaleDownParallelism:    1,
		ScaleDownSimulationTimeout: 10 * time.Second,
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 1.0,
			ScaleDownUnneededTime:         0,
		},
	}
}

func buildScaleDownDependencies(b *testing.B, cs *initialClusterState, autoscalingOpts config.AutoscalingOptions) scaleDownDependencies {
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

	fakeActuator := NewFakeActuator(autoscalingCtx, &fakeActuationStatus{}, b)
	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{CustomResourcesProcessor: procs.CustomResourcesProcessor, QuotaProvider: resourcequotas.NewCloudMinProvider(provider)})
	p := planner.New(&autoscalingCtx, procs, options.NodeDeleteOptions{}, nil, factory)

	return scaleDownDependencies{
		AutoscalingCtx: &autoscalingCtx,
		Processors:     procs,
		Planner:        p,
		Actuator:       fakeActuator,
	}
}

// runIteration executes a benchmarking iteration using newer b.Loop (https://go.dev/blog/testing-b-loop).
// Crucial to pass -benchtime=1x to let the Loop know to run scaledown strategy exactly once.
func (scenario scaleDownScenario) runIteration(b *testing.B) {
	cs := scenario.InitialClusterState()

	opts := defaultAutoscalingOptions()
	if scenario.AutoscalingOpts != nil {
		scenario.AutoscalingOpts(&opts)
	}
	mt := scenario.Metrics()
	deps := buildScaleDownDependencies(b, cs, opts)

	mt.Compute(deps.AutoscalingCtx.ClusterSnapshot, nil, b)
	mt.Report("Initial Benchmark State", b)

	if b.N != 1 {
		b.Fatalf("This benchmark must be executed exactly once per run, use '-benchtime=1x' flag.")
	}

	if scenario.ScaleDownStrategy != nil {
		for b.Loop() {
			scenario.ScaleDownStrategy(b, mt, deps)
		}
	}

	mt.ReportBenchmarks(b)
}

func BenchmarkRunOnceScaleDown(b *testing.B) {
	s := scaleDownScenario{
		Name:                "three ng, basic setup",
		InitialClusterState: differentNodeSizesMemIrrelevant(),
		ScaleDownStrategy:   vanillaScaleDownStrategy,
		Metrics: func() *metricsTracker {
			return NewMetricsTracker(
				NewNodeCountMetric(withNodeReporting(false)),
				NewResourceUtilizationMetric(),
				NewResourceFragmentationMetric(),
				NewEvictionCountMetric(),
				NewEvictionCostMetric(withClusterReporting(false)),
				NewNodeCostMetric(),
			)
		},
		AutoscalingOpts: func(opts *config.AutoscalingOptions) {
			opts.NodeGroupDefaults.ScaleDownUtilizationThreshold = 0.75
		},
	}
	s.runIteration(b)
}
