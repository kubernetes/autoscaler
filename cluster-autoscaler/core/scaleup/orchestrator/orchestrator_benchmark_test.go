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
	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

func runKarpenterSimulatorBenchmark(b *testing.B, salvoEnabled bool) {
	ctx := &ca_context.AutoscalingContext{
		ClusterSnapshot: testsnapshot.NewTestSnapshotOrDie(b),
		AutoscalingOptions: config.AutoscalingOptions{
			KarpenterSimulatorEnabled: true,
			SalvoScaleUp:       salvoEnabled,
		},
	}

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.AddNodeGroup("ng1", 1, 5000, 1)

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

	var unschedulablePods []*apiv1.Pod
	var podEquivalenceGroups []*equivalence.PodGroup
	for i := 0; i < 1500; i++ {
		p := createKarpenterTestPod(fmt.Sprintf("p%d", i), 500, 1000)
		unschedulablePods = append(unschedulablePods, p)
		podEquivalenceGroups = append(podEquivalenceGroups, &equivalence.PodGroup{Pods: []*apiv1.Pod{p}})
	}

	factory := resourcequotas.NewTrackerFactory(resourcequotas.TrackerOptions{
		CustomResourcesProcessor: customresources.NewDefaultCustomResourcesProcessor(false, false),
		QuotaProvider:            resourcequotas.NewCloudQuotasProvider(provider),
	})
	tracker, _ := factory.NewQuotasTracker(ctx, nil)

	sim := NewKarpenterSimulator(nil, converter, processor, salvoEnabled)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = sim.Simulate(ctx, podEquivalenceGroups, unschedulablePods, nil, provider.NodeGroups(), nodeInfos, tracker, time.Now(), false)
	}
}

func BenchmarkKarpenterSimulator_SalvoEnabled(b *testing.B) {
	runKarpenterSimulatorBenchmark(b, true)
}

func BenchmarkKarpenterSimulator_SalvoDisabled(b *testing.B) {
	runKarpenterSimulatorBenchmark(b, false)
}
