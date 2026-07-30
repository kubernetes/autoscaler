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

package inmemory

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const (
	plannerTestUnneededTime = 1 * time.Minute
	plannerStepDuration     = 10 * time.Second
)

// TestPlanner_AtomicAndNonAtomicScaleDown verifies that the scaledown planner
// correctly evaluates and processes both atomic and non-atomic node groups.
func TestPlanner_AtomicAndNonAtomicScaleDown(t *testing.T) {
	testConfig := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(plannerTestUnneededTime),
		)

	options := testConfig.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		// Create non-atomic node group with 2 nodes.
		naTemplate := test.BuildTestNode("node-na", 1000, 1000, test.IsReady(true))
		fakes.CloudProvider.AddNodeGroup("ng-nonatomic", fakecloudprovider.WithNodes(naTemplate, 2))

		// Create atomic node group with ZeroOrMaxNodeScaling option and 2 nodes.
		aTemplate := test.BuildTestNode("node-a", 1000, 1000, test.IsReady(true))
		atomicOpts := &config.NodeGroupAutoscalingOptions{
			ZeroOrMaxNodeScaling: true,
		}
		fakes.CloudProvider.AddNodeGroup("ng-atomic",
			fakecloudprovider.WithNodes(aTemplate, 2),
			fakecloudprovider.WithNGOptions(atomicOpts),
		)

		// Initial size check
		sizeNonAtomic, _ := fakes.CloudProvider.GetNodeGroup("ng-nonatomic").TargetSize()
		assert.Equal(t, 2, sizeNonAtomic)
		sizeAtomic, _ := fakes.CloudProvider.GetNodeGroup("ng-atomic").TargetSize()
		assert.Equal(t, 2, sizeAtomic)

		// Run CA loop once to mark empty nodes as unneeded.
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerStepDuration)

		// Advance time past unneededTime to allow scale down actuation.
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerTestUnneededTime+time.Second)

		// Run loop again to complete removal of all unneeded nodes.
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerStepDuration)

		// Verify both non-atomic and atomic empty nodes are scaled down.
		finalSizeNonAtomic, _ := fakes.CloudProvider.GetNodeGroup("ng-nonatomic").TargetSize()
		assert.Equal(t, 0, finalSizeNonAtomic, "Non-atomic empty nodes should be scaled down")

		finalSizeAtomic, _ := fakes.CloudProvider.GetNodeGroup("ng-atomic").TargetSize()
		assert.Equal(t, 0, finalSizeAtomic, "Atomic empty nodes should be scaled down")
	})
}

// TestPlanner_AtomicNodesProcessedAfterNonAtomicLimit verifies that when non-atomic simulation
// hits the unneededNodesLimit, non-atomic loop breaks, but atomic nodes are still evaluated.
func TestPlanner_AtomicNodesProcessedAfterNonAtomicLimit(t *testing.T) {
	testConfig := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(plannerTestUnneededTime),
			func(o *config.AutoscalingOptions) {
				// Set MaxScaleDownParallelism to 1 so unneededNodesLimit() is capped at 2.
				o.MaxScaleDownParallelism = 1
			},
		)

	options := testConfig.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		// Create 4 non-atomic nodes (exceeding limit of 2) in ng-nonatomic.
		naTemplate := test.BuildTestNode("node-na", 1000, 1000, test.IsReady(true))
		fakes.CloudProvider.AddNodeGroup("ng-nonatomic", fakecloudprovider.WithNodes(naTemplate, 4))

		// Create atomic node group with ZeroOrMaxNodeScaling option and 2 nodes.
		aTemplate := test.BuildTestNode("node-a", 1000, 1000, test.IsReady(true))
		atomicOpts := &config.NodeGroupAutoscalingOptions{
			ZeroOrMaxNodeScaling: true,
		}
		fakes.CloudProvider.AddNodeGroup("ng-atomic",
			fakecloudprovider.WithNodes(aTemplate, 2),
			fakecloudprovider.WithNGOptions(atomicOpts),
		)

		// Initial verification of target sizes
		sizeNA, _ := fakes.CloudProvider.GetNodeGroup("ng-nonatomic").TargetSize()
		assert.Equal(t, 4, sizeNA)
		sizeA, _ := fakes.CloudProvider.GetNodeGroup("ng-atomic").TargetSize()
		assert.Equal(t, 2, sizeA)

		// Run loop to mark unneeded
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerStepDuration)

		// Advance past unneeded time and run scale down loop
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerTestUnneededTime+time.Second)

		// Run subsequent iterations to allow scale down actuation of all evaluated nodes.
		for i := 0; i < 5; i++ {
			synctestutils.MustRunOnceAfter(t, autoscaler, plannerStepDuration)
		}

		// Check that atomic nodes were evaluated and scaled down.
		finalSizeAtomic, _ := fakes.CloudProvider.GetNodeGroup("ng-atomic").TargetSize()
		assert.Equal(t, 0, finalSizeAtomic, "Atomic nodes should be evaluated and scaled down even when non-atomic limit is reached")
	})
}

// TestPlanner_SimulationTimeoutSkipsAtomicAndNonAtomic verifies that a simulation timeout
// during non-atomic loop properly skips remaining non-atomic nodes and all atomic nodes.
func TestPlanner_SimulationTimeoutSkipsAtomicAndNonAtomic(t *testing.T) {
	testConfig := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(plannerTestUnneededTime),
			func(o *config.AutoscalingOptions) {
				// Set simulation timeout to negative duration to force immediate simulation timeout.
				o.ScaleDownSimulationTimeout = -1 * time.Second
			},
		)

	options := testConfig.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		// Add destination helper node where pods can schedule if simulation runs.
		destNode := test.BuildTestNode("dest-node", 10000, 10000, test.IsReady(true))
		fakes.K8s.AddNode(destNode)

		naTemplate := test.BuildTestNode("node-na", 1000, 1000, test.IsReady(true))
		fakes.CloudProvider.AddNodeGroup("ng-nonatomic", fakecloudprovider.WithNodes(naTemplate, 2))

		aTemplate := test.BuildTestNode("node-a", 1000, 1000, test.IsReady(true))
		atomicOpts := &config.NodeGroupAutoscalingOptions{
			ZeroOrMaxNodeScaling: true,
		}
		fakes.CloudProvider.AddNodeGroup("ng-atomic",
			fakecloudprovider.WithNodes(aTemplate, 2),
			fakecloudprovider.WithNGOptions(atomicOpts),
		)

		// Add scheduled pods to nodes so scale down simulation is required.
		fakes.K8s.AddPod(test.BuildScheduledTestPod("pod-na-0", 100, 100, "ng-nonatomic-node-0"))
		fakes.K8s.AddPod(test.BuildScheduledTestPod("pod-na-1", 100, 100, "ng-nonatomic-node-1"))
		fakes.K8s.AddPod(test.BuildScheduledTestPod("pod-a-0", 100, 100, "ng-atomic-node-0"))
		fakes.K8s.AddPod(test.BuildScheduledTestPod("pod-a-1", 100, 100, "ng-atomic-node-1"))

		// Run CA loop. With simulation timeout <= 0, nodes should be skipped.
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerStepDuration)
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerTestUnneededTime+time.Second)

		// Node groups should remain at original size because scale down simulation was skipped due to timeout.
		sizeNA, _ := fakes.CloudProvider.GetNodeGroup("ng-nonatomic").TargetSize()
		assert.Equal(t, 2, sizeNA)
		sizeA, _ := fakes.CloudProvider.GetNodeGroup("ng-atomic").TargetSize()
		assert.Equal(t, 2, sizeA)
	})
}

// TestPlanner_AtomicNodePdbHandling verifies scale down behavior when atomic nodes host pods subject to PDBs.
func TestPlanner_AtomicNodePdbHandling(t *testing.T) {
	testConfig := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(plannerTestUnneededTime),
		)

	options := testConfig.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		destNode := test.BuildTestNode("dest-node", 10000, 10000, test.IsReady(true))
		fakes.K8s.AddNode(destNode)

		aTemplate := test.BuildTestNode("node-a", 1000, 1000, test.IsReady(true))
		atomicOpts := &config.NodeGroupAutoscalingOptions{
			ZeroOrMaxNodeScaling: true,
		}
		fakes.CloudProvider.AddNodeGroup("ng-atomic",
			fakecloudprovider.WithNodes(aTemplate, 2),
			fakecloudprovider.WithNGOptions(atomicOpts),
		)

		// Create PDB allowing at most 1 pod disruption at a time
		maxUnavailable := intstr.FromInt32(1)
		pdb := &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app-pdb",
				Namespace: "default",
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				MaxUnavailable: &maxUnavailable,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
		}
		_, err = fakes.KubeClient.PolicyV1().PodDisruptionBudgets("default").Create(ctx, pdb, metav1.CreateOptions{})
		assert.NoError(t, err)

		pod0 := test.BuildScheduledTestPod("pod-a-0", 100, 100, "ng-atomic-node-0")
		pod0.Labels = map[string]string{"app": "test"}
		pod1 := test.BuildScheduledTestPod("pod-a-1", 100, 100, "ng-atomic-node-1")
		pod1.Labels = map[string]string{"app": "test"}
		fakes.K8s.AddPod(pod0)
		fakes.K8s.AddPod(pod1)

		// Run CA loop once to mark unneeded and run simulation.
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerStepDuration)
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerTestUnneededTime+time.Second)

		// Run loop to actuate.
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerStepDuration)

		sizeA, _ := fakes.CloudProvider.GetNodeGroup("ng-atomic").TargetSize()
		assert.LessOrEqual(t, sizeA, 2, "Atomic node group evaluation with PDB succeeded")
	})
}

// TestPlanner_IncompleteAtomicNodeGroupPrefiltered verifies that an atomic node group
// with fewer unneeded nodes than its target size is prefiltered out before simulation,
// and its target size remains unchanged.
func TestPlanner_IncompleteAtomicNodeGroupPrefiltered(t *testing.T) {
	testConfig := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(plannerTestUnneededTime),
		)

	options := testConfig.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		// Create atomic node group with ZeroOrMaxNodeScaling option and 2 nodes.
		aTemplate := test.BuildTestNode("node", 1000, 1000, test.IsReady(true))
		atomicOpts := &config.NodeGroupAutoscalingOptions{
			ZeroOrMaxNodeScaling: true,
		}
		fakes.CloudProvider.AddNodeGroup("ng-atomic",
			fakecloudprovider.WithNodes(aTemplate, 2),
			fakecloudprovider.WithNGOptions(atomicOpts),
		)

		// Add a non-removable pod to node 0 of ng-atomic so node 0 is NOT unneeded.
		blockingPod := test.BuildScheduledTestPod("blocking-pod", 100, 100, "ng-atomic-node-0")
		blockingPod.Annotations = map[string]string{
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
		}
		fakes.K8s.AddPod(blockingPod)

		// Run CA loop once to process candidates.
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerStepDuration)

		// Advance past unneeded time and run scale down loop.
		synctestutils.MustRunOnceAfter(t, autoscaler, plannerTestUnneededTime+time.Second)

		// The atomic node group should NOT scale down because only 1 of 2 nodes was unneeded.
		sizeA, _ := fakes.CloudProvider.GetNodeGroup("ng-atomic").TargetSize()
		assert.Equal(t, 2, sizeA, "Incomplete atomic node group should be prefiltered and remain at target size 2")
	})
}
