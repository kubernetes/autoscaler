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

package bench

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

var (
	nodesCount        = flag.Int("nodes-count", 500, "Number of nodes in the cluster")
	nodesPerNodeGroup = flag.Int("nodes-per-node-group", 50, "Number of nodes per node group")
	podsPerNode       = flag.Int("pods-per-node", 10, "Number of pods per node")
	surgeCount        = flag.Int("surge-count", 100, "Number of surge pods to schedule")
)

func BenchmarkRunOnceAffinitySurge_Hostname_Default(b *testing.B) {
	s := scenario{
		setup:  setupAffinitySurge(*nodesCount, *podsPerNode, *surgeCount, *nodesPerNodeGroup, apiv1.LabelHostname),
		verify: verifyTotalTargetSizeEquals(*nodesCount + *surgeCount),
		config: func(opts *config.AutoscalingOptions) {
			configureAffinitySurge(opts)
			opts.KarpenterSimulatorEnabled = false
		},
	}
	s.run(b)
}

func BenchmarkRunOnceAffinitySurge_Hostname_Karpenter(b *testing.B) {
	s := scenario{
		setup:  setupAffinitySurge(*nodesCount, *podsPerNode, *surgeCount, *nodesPerNodeGroup, apiv1.LabelHostname),
		verify: verifyTotalTargetSizeEquals(*nodesCount + *surgeCount),
		config: func(opts *config.AutoscalingOptions) {
			configureAffinitySurge(opts)
			opts.KarpenterSimulatorEnabled = true
		},
	}
	s.run(b)
}

// In the Zonal setup, each pod has an anti-affinity constraint against its
// application partner. Across the entire set of surge pods, at least 1/2
// of them are guaranteed to fit into at least one of the zones, even in
// the worst case where all surge pods come in anti-affinity pairs.
// By using the 'most-pods' expander, we guarantee that the chosen zone
// will fit at least this amount.
func BenchmarkRunOnceAffinitySurge_Zonal_Default(b *testing.B) {
	s := scenario{
		setup:  setupAffinitySurge(*nodesCount, *podsPerNode, *surgeCount, *nodesPerNodeGroup, apiv1.LabelTopologyZone),
		verify: verifyTotalTargetSizeAtLeast(*nodesCount + (*surgeCount / 2)),
		config: func(opts *config.AutoscalingOptions) {
			configureAffinitySurge(opts)
			opts.KarpenterSimulatorEnabled = false
		},
	}
	s.run(b)
}

// In Karpenter simulation, the generated NodeClaims are initially zone-agnostic.
// If both pods of an anti-affinity pair are in the surge set, scheduling the first
// pod results in an in-flight zone-agnostic NodeClaim. The second pod cannot be
// scheduled because the zone of the first pod is undecided. In the worst-case
// scenario, all surge pods are paired with each other, blocking exactly half of them.
func BenchmarkRunOnceAffinitySurge_Zonal_Karpenter(b *testing.B) {
	s := scenario{
		setup:  setupAffinitySurge(*nodesCount, *podsPerNode, *surgeCount, *nodesPerNodeGroup, apiv1.LabelTopologyZone),
		verify: verifyTotalTargetSizeAtLeast(*nodesCount + (*surgeCount / 2)),
		config: func(opts *config.AutoscalingOptions) {
			configureAffinitySurge(opts)
			opts.KarpenterSimulatorEnabled = true
		},
	}
	s.run(b)
}

func configureAffinitySurge(opts *config.AutoscalingOptions) {
	opts.MaxNodesPerScaleUp = 1000000
	opts.ScaleUpFromZero = true
	opts.MaxCoresTotal = 1000000000
	opts.MaxMemoryTotal = 1000000000
	opts.MaxNodesTotal = 1000000
	opts.PredicateParallelism = 16
	opts.ExpanderNames = expander.MostPodsExpanderName
	opts.BalanceSimilarNodeGroups = true
	// Use high binpacking limits to avoid benchmarks being cut by internal timeouts.
	opts.MaxBinpackingTime = 10 * time.Minute
	opts.MaxNodeGroupBinpackingDuration = 10 * time.Minute
}

func setupAffinitySurge(nodesCount, podsPerNode, surgePodsCount, nodesPerGroup int, topologyKey string) func(*integration.FakeSet) error {
	return func(clusterFakes *integration.FakeSet) error {
		clusterFakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 1000000000)
		clusterFakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameMemory, 0, 1000000000)

		numGroups := nodesCount / nodesPerGroup
		totalPods := (nodesCount * podsPerNode) + surgePodsCount

		// Use fixed seed for reproducibility
		r := rand.New(rand.NewSource(42))

		// Partition indices into scheduled and surge
		indices := r.Perm(totalPods)
		isSurge := make(map[int]bool)
		for i := 0; i < surgePodsCount; i++ {
			isSurge[indices[i]] = true
		}

		scheduledPodIdx := 0
		surgePodIdx := 0

		for i := 0; i < numGroups; i++ {
			ngName := fmt.Sprintf("ng-affinity-%d", i)
			zone := fmt.Sprintf("zone-%d", i%3)
			instanceType := fmt.Sprintf("type-%d", i/3)

			nTemplate := BuildTestNode(fmt.Sprintf("%s-template", ngName), nodeCPU, nodeMem)
			if nTemplate.Labels == nil {
				nTemplate.Labels = make(map[string]string)
			}
			nTemplate.Labels["kubernetes.io/hostname"] = nTemplate.Name
			nTemplate.Labels["topology.kubernetes.io/zone"] = zone
			nTemplate.Labels["node.kubernetes.io/instance-type"] = instanceType
			SetNodeReadyState(nTemplate, true, time.Now())

			clusterFakes.CloudProvider.AddNodeGroup(ngName,
				testprovider.WithTemplate(framework.NewNodeInfo(nTemplate, nil)),
				testprovider.WithNGSize(0, 1000000),
			)

			ng := clusterFakes.CloudProvider.GetNodeGroup(ngName)
			if err := ng.(*testprovider.NodeGroup).DecreaseTargetSize(-nodesPerGroup); err != nil {
				return err
			}
			ngID := ng.Id()

			for j := 0; j < nodesPerGroup; j++ {
				nodeName := fmt.Sprintf("%s-node-%d", ngID, j)

				node := BuildTestNode(nodeName, nodeCPU, nodeMem)
				if node.Labels == nil {
					node.Labels = make(map[string]string)
				}
				node.Labels["kubernetes.io/hostname"] = nodeName
				node.Labels["topology.kubernetes.io/zone"] = zone
				node.Labels["node.kubernetes.io/instance-type"] = instanceType
				SetNodeReadyState(node, true, time.Now())
				clusterFakes.K8s.AddNode(node)
				clusterFakes.CloudProvider.AddNode(ngID, node)

				for k := 0; k < podsPerNode; k++ {
					// Find the next non-surge pod index to schedule on this node
					for isSurge[scheduledPodIdx+surgePodIdx] {
						surgePodIdx++
					}
					podIdx := scheduledPodIdx + surgePodIdx

					podName := fmt.Sprintf("scheduled-pod-%d", podIdx)
					appLabels := map[string]string{"app": fmt.Sprintf("app-%d", podIdx/2)}
					pod := BuildTestPod(podName, nodeCPU/int64(podsPerNode), nodeMem/int64(podsPerNode),
						WithNodeName(nodeName),
						WithLabels(appLabels),
						WithPodAntiAffinity(appLabels, topologyKey),
					)

					clusterFakes.K8s.AddPod(pod)
					scheduledPodIdx++
				}
			}
		}

		// Add remaining surge pods
		for podIdx := 0; podIdx < totalPods; podIdx++ {
			if isSurge[podIdx] {
				podName := fmt.Sprintf("surge-pod-%d", podIdx)
				appLabels := map[string]string{"app": fmt.Sprintf("app-%d", podIdx/2)}
				pod := BuildTestPod(podName, nodeCPU, nodeMem,
					MarkUnschedulable(),
					WithLabels(appLabels),
					WithPodAntiAffinity(appLabels, topologyKey),
				)
				clusterFakes.K8s.AddPod(pod)
			}
		}

		return nil
	}
}

func totalTargetSize(clusterFakes *integration.FakeSet) (int, error) {
	total := 0
	for _, ng := range clusterFakes.CloudProvider.NodeGroups() {
		targetSize, err := ng.TargetSize()
		if err != nil {
			return 0, err
		}
		total += targetSize
	}
	return total, nil
}

func verifyTotalTargetSizeEquals(expected int) func(*integration.FakeSet) error {
	return func(clusterFakes *integration.FakeSet) error {
		total, err := totalTargetSize(clusterFakes)
		if err != nil {
			return err
		}
		if total != expected {
			return fmt.Errorf("verify failed: expected total target size %d, got %d", expected, total)
		}
		return nil
	}
}

func verifyTotalTargetSizeAtLeast(expectedMin int) func(*integration.FakeSet) error {
	return func(clusterFakes *integration.FakeSet) error {
		total, err := totalTargetSize(clusterFakes)
		if err != nil {
			return err
		}
		if total < expectedMin {
			return fmt.Errorf("verify failed: expected total target size to be at least %d, got %d", expectedMin, total)
		}
		return nil
	}
}
