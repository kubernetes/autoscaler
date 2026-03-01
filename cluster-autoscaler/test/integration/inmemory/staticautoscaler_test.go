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

package inmemory

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	tutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const (
	unneededTime = 1 * time.Minute
)

func TestStaticAutoscalerRunOnce(t *testing.T) {
	cfg := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(time.Minute),
		)

	options := cfg.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		err := infra.StartAndSyncInformers(ctx)
		assert.NoError(t, err)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		// Set up nodes
		n1 := tutils.BuildTestNode("n1", 1000, 1000, tutils.IsReady(true))
		n2 := tutils.BuildTestNode("n2", 1000, 1000, tutils.IsReady(true))
		n3 := tutils.BuildTestNode("n3", 1000, 1000, tutils.IsReady(true))
		n4 := tutils.BuildTestNode("n4", 1000, 1000, tutils.IsReady(true))

		fakes.CloudProvider.AddNodeGroup("ng1",
			fakecloudprovider.WithMinMax(1, 10),
			fakecloudprovider.WithTargetSize(1),
			fakecloudprovider.WithNode(n1))

		// Set up pods
		p1 := tutils.BuildScheduledTestPod("p1", 600, 100, n1.Name)
		p2 := tutils.BuildTestPod("p2", 600, 100, tutils.MarkUnschedulable())

		fakes.K8s.AddPod(p1)
		fakes.K8s.AddPod(p2)

		// Simulate scale up for the unschedulable pod (p2)
		synctestutils.MustRunOnceAfter(t, autoscaler, time.Hour)

		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1) // Scales up for p2

		// Mark unneeded nodes (Add n2 to ng1)
		fakes.K8s.AddNode(n2)
		fakes.CloudProvider.AddNodeGroup("ng1", fakecloudprovider.WithNode(n2), fakecloudprovider.WithTargetSize(2)) // update tg

		synctestutils.MustRunOnceAfter(t, autoscaler, 2*time.Hour)

		// Scale down
		fakes.K8s.DeletePod(p2.Namespace, p2.Name)
		synctestutils.MustRunOnceAfter(t, autoscaler, 3*time.Hour)

		// Mark unregistered nodes
		fakes.CloudProvider.AddNodeGroup("ng2", fakecloudprovider.WithMinMax(0, 10), fakecloudprovider.WithTargetSize(0), fakecloudprovider.WithNode(n3))
		fakes.K8s.AddNode(n3)
		synctestutils.MustRunOnceAfter(t, autoscaler, 4*time.Hour)

		// Remove unregistered nodes
		synctestutils.MustRunOnceAfter(t, autoscaler, 5*time.Hour)

		// Verify scale up to node group min size when cluster is empty
		fakes.K8s.DeletePod(p1.Namespace, p1.Name)

		fakes.CloudProvider.AddNodeGroup("ng3", fakecloudprovider.WithMinMax(3, 10), fakecloudprovider.WithTargetSize(1), fakecloudprovider.WithNode(n4))
		fakes.K8s.AddNode(n4)
		synctestutils.MustRunOnceAfter(t, autoscaler, 5*time.Hour)

		tg3, _ := fakes.CloudProvider.GetNodeGroup("ng3").TargetSize()
		assert.Equal(t, 3, tg3) // Scales up to min size 3
	})
}

func TestScaleUp_ResourceLimits(t *testing.T) {
	config := integration.NewTestConfig()

	options := config.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		n := tutils.BuildTestNode("ng-node-0", 1000, 1000, tutils.IsReady(true))
		fakes.CloudProvider.AddNodeGroup("ng", fakecloudprovider.WithNode(n))
		fakes.K8s.AddPod(tutils.BuildTestPod("pod", 600, 100, tutils.MarkUnschedulable()))

		// Scale-up should be blocked.
		fakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 1)

		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)
		size, _ := fakes.CloudProvider.GetNodeGroup("ng").TargetSize()
		assert.Equal(t, 1, size, "Should not scale up when max cores limit is reached")

		// Scale-up should succeed.
		fakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 2)

		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)
		newSize, _ := fakes.CloudProvider.GetNodeGroup("ng").TargetSize()
		assert.Equal(t, 2, newSize, "Should scale up after resource limit is increased")
	})
}

func TestStaticAutoscalerRunOnceWithScaleDownDelayPerNG(t *testing.T) {
	testCases := []struct {
		name                string
		beforeTest          func(t *testing.T, fakes *integration.FakeSet, autoscaler core.Autoscaler)
		expectedTargetSize1 int
		expectedTargetSize2 int
	}{
		{
			name: "ng1 scaled up recently - both ng1 and ng2 have under-utilized nodes",
			beforeTest: func(t *testing.T, fakes *integration.FakeSet, autoscaler core.Autoscaler) {
				// We want ng1 to scale up *now*, so it gets a 5m scale-down delay.

				// 1. Protect n1 and n2 from being unneeded while we scale up ng1
				p1 := tutils.BuildScheduledTestPod("p1-n1", 600, 100, "n1")
				p2 := tutils.BuildScheduledTestPod("p2-n2", 600, 100, "n2")
				fakes.K8s.AddPod(p1)
				fakes.K8s.AddPod(p2)

				// 2. Trigger scale up on ng1
				p := tutils.BuildTestPod("p-ng1", 600, 100, tutils.MarkUnschedulable())
				p.Spec.NodeSelector = map[string]string{"ng": "ng1"}
				fakes.K8s.AddPod(p)

				// Run CA. This should trigger a scale up for ng1 to accommodate pod.
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

				// Validate it actually scaled up to ensure the event was recorded
				tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
				assert.Equal(t, 2, tg1, "Setup failed: ng1 did not scale up")

				// The scale up was registered.
				// 3. Remove all pods. Now n1 and n2 become unneeded at THIS point in time.
				// Their scale down timer starts now (1m).
				fakes.K8s.DeletePod(p.Namespace, p.Name)
				fakes.K8s.DeletePod(p1.Namespace, p1.Name)
				fakes.K8s.DeletePod(p2.Namespace, p2.Name)

				// CA runs, notices n1 and n2 are unneeded.
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
			},
			expectedTargetSize1: 2, // ng1 was 2, delayed, so it doesn't scale down
			expectedTargetSize2: 0, // ng2 scales down
		},
		{
			name: "ng2 scaled down recently - both ng1 and ng2 have under-utilized nodes",
			beforeTest: func(t *testing.T, fakes *integration.FakeSet, autoscaler core.Autoscaler) {
				// 1. Add a temporary node to ng2 to scale it down *now*.
				n3 := tutils.BuildTestNode("n3", 1000, 1000, tutils.IsReady(true))
				n3.Labels["ng"] = "ng2"
				fakes.CloudProvider.AddNodeGroup("ng2", fakecloudprovider.WithNode(n3), fakecloudprovider.WithTargetSize(2))
				fakes.K8s.AddNode(n3)

				// CA loop to discover n3
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

				// 2. Fast forward 1m to satisfy unneeded time, triggering scale down of n3 (and potentially n1, n2 if they were unneeded, but only 1 scale down per NG is allowed usually, or they compete).
				// We want ONLY n3 to scale down first to register the 5m delay on ng2.

				// Wait for unneeded time (1m) to pass.
				// Actually, since n1, n2, n3 are all empty, CA might pick any.
				// To force n3, we can temporarily put pods on n1 and n2.
				p1 := tutils.BuildScheduledTestPod("p1-n1", 600, 100, "n1")
				p2 := tutils.BuildScheduledTestPod("p2-n2", 600, 100, "n2")
				fakes.K8s.AddPod(p1)
				fakes.K8s.AddPod(p2)

				synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute*2)

				// At this point n3 should be scaled down. ng2 now has a "ScaleDownDelayAfterDelete" (5m).

				// Remove pods on n1 and n2 so they become unneeded *now*
				fakes.K8s.DeletePod(p1.Namespace, p1.Name)
				fakes.K8s.DeletePod(p2.Namespace, p2.Name)
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
			},
			expectedTargetSize1: 0, // ng1 scales down
			expectedTargetSize2: 1, // ng2 delayed
		},
		{
			name: "ng1 had scale-down failure - both ng1 and ng2 have under-utilized nodes",
			beforeTest: func(t *testing.T, fakes *integration.FakeSet, autoscaler core.Autoscaler) {
				// 1. Add nExtra to ng1
				nExtra := tutils.BuildTestNode("nExtra", 1000, 1000, tutils.IsReady(true))
				nExtra.Labels["ng"] = "ng1"
				fakes.CloudProvider.AddNodeGroup("ng1", fakecloudprovider.WithNode(nExtra), fakecloudprovider.WithTargetSize(2))
				fakes.K8s.AddNode(nExtra)

				// CA loop to discover nExtra
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

				// Protect n1, n2
				p1 := tutils.BuildScheduledTestPod("p1-n1", 600, 100, "n1")
				p2 := tutils.BuildScheduledTestPod("p2-n2", 600, 100, "n2")
				fakes.K8s.AddPod(p1)
				fakes.K8s.AddPod(p2)

				// Inject delete error
				ng1 := fakes.CloudProvider.GetNodeGroup("ng1").(*fakecloudprovider.NodeGroup)
				originalDelete := ng1.DeleteNodesOverride
				ng1.DeleteNodesOverride = func(nodes []*apiv1.Node) error {
					return fmt.Errorf("simulated scale down failure")
				}

				// Wait 2m for nExtra to become unneeded and attempt scale down (which fails)
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute*2)

				// Restore to allow future scale-downs
				ng1.DeleteNodesOverride = originalDelete
				// Manually clean up nExtra so it doesn't interfere
				fakes.K8s.DeleteNode("nExtra")
				fakes.CloudProvider.AddNodeGroup("ng1", fakecloudprovider.WithTargetSize(1))
				_ = ng1.DeleteNodes([]*apiv1.Node{nExtra})
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

				// Remove pods on n1 and n2 so they become unneeded *now*
				fakes.K8s.DeletePod(p1.Namespace, p1.Name)
				fakes.K8s.DeletePod(p2.Namespace, p2.Name)
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
			},
			expectedTargetSize1: 1, // ng1 delayed
			expectedTargetSize2: 0, // ng2 scales down
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				config := integration.NewTestConfig().
					WithOverrides(
						integration.WithScaleDownUnneededTime(time.Minute),
						integration.WithScaleDownDelayAfterAdd(5*time.Minute),
						integration.WithScaleDownDelayAfterDelete(5*time.Minute),
						integration.WithScaleDownDelayAfterFailure(5*time.Minute),
						integration.WithMaxScaleDownParallelism(10),
						integration.WithMaxDrainParallelism(1),
					)

				options := config.ResolveOptions()
				infra := integration.SetupInfrastructure(t)
				fakes := infra.Fakes

				ctx, cancel := context.WithCancel(context.Background())
				defer synctestutils.TearDown(cancel)

				err := infra.StartAndSyncInformers(ctx)
				assert.NoError(t, err)

				n1 := tutils.BuildTestNode("n1", 1000, 1000, tutils.IsReady(true))
				n1.Labels["ng"] = "ng1"
				fakes.CloudProvider.AddNodeGroup("ng1", fakecloudprovider.WithMinMax(0, 5), fakecloudprovider.WithTargetSize(1), fakecloudprovider.WithNode(n1))
				fakes.K8s.AddNode(n1)

				n2 := tutils.BuildTestNode("n2", 1000, 1000, tutils.IsReady(true))
				n2.Labels["ng"] = "ng2"
				fakes.CloudProvider.AddNodeGroup("ng2", fakecloudprovider.WithMinMax(0, 5), fakecloudprovider.WithTargetSize(1), fakecloudprovider.WithNode(n2))
				fakes.K8s.AddNode(n2)

				autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
				assert.NoError(t, err)

				tc.beforeTest(t, fakes, autoscaler)

				// After beforeTest, n1 and n2 are unneeded (either they were already empty, or we just removed pods from them).
				// We wait 2 minutes (unneeded time is 1m) to trigger scale down.
				// However, the Delay should be 5m.
				// So the protected NG should NOT scale down, but the unprotected one SHOULD.
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute*2)

				// Check target sizes
				tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
				tg2, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()

				assert.Equal(t, tc.expectedTargetSize1, tg1, "target size for ng1")
				assert.Equal(t, tc.expectedTargetSize2, tg2, "target size for ng2")
			})
		})
	}
}

func TestStaticAutoscalerRunOnceWithALongUnregisteredNode(t *testing.T) {
	for _, forceDeleteLongUnregisteredNodes := range []bool{false, true} {
		t.Run(fmt.Sprintf("forceDeleteLongUnregisteredNodes=%v", forceDeleteLongUnregisteredNodes), func(t *testing.T) {
			cfg := integration.NewTestConfig().
				WithOverrides(
					integration.WithScaleDownUnneededTime(time.Minute),
				)
			options := cfg.ResolveOptions()
			options.ForceDeleteLongUnregisteredNodes = forceDeleteLongUnregisteredNodes

			infra := integration.SetupInfrastructure(t)
			fakes := infra.Fakes

			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer synctestutils.TearDown(cancel)

				autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
				assert.NoError(t, err)

				n1 := tutils.BuildTestNode("n1", 1000, 1000, tutils.IsReady(true))
				n2 := tutils.BuildTestNode("n2", 1000, 1000, tutils.IsReady(true))

				p1 := tutils.BuildScheduledTestPod("p1", 600, 100, "n1")
				p2 := tutils.BuildTestPod("p2", 600, 100, tutils.MarkUnschedulable())

				fakes.CloudProvider.AddNodeGroup("ng1", fakecloudprovider.WithMinMax(2, 10), fakecloudprovider.WithTargetSize(2), fakecloudprovider.WithNode(n1))

				// broken node, that will be just hanging out there during
				// the test (it can't be removed since that would validate group min size)
				brokenNode := tutils.BuildTestNode("broken", 1000, 1000)
				fakes.CloudProvider.AddNode("ng1", brokenNode)

				fakes.K8s.AddNode(n1)
				fakes.K8s.AddPod(p1)

				// Initial run to register n1 and discover brokenNode as unregistered
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)

				// broken node failed to register in time, run again
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)

				if !forceDeleteLongUnregisteredNodes {
					// Add unschedulable pod p2 to trigger scale up
					fakes.K8s.AddPod(p2)

					synctestutils.MustRunOnceAfter(t, autoscaler, time.Hour)
					tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
					assert.Equal(t, 3, tg1) // scale up to 3 for p2

					// Register the nodes that were scaled up so they don't timeout
					fakes.K8s.AddNode(n2)
					fakes.CloudProvider.AddNode("ng1", n2)
					fakes.CloudProvider.GetNodeGroup("ng1").(*fakecloudprovider.NodeGroup).SetTargetSize(3)
				}

				// Remove broken node
				fakes.K8s.DeletePod(p1.Namespace, p1.Name)
				fakes.K8s.DeletePod(p2.Namespace, p2.Name)
				synctestutils.MustRunOnceAfter(t, autoscaler, 2*time.Hour)

				// check if broken node is removed
				ng, err := fakes.CloudProvider.NodeGroupForNode(brokenNode)
				assert.NoError(t, err)
				assert.Nil(t, ng) // SHOULD return nil if node is deleted
			})
		})
	}
}

func TestStaticAutoscalerRunOncePodsWithPriorities(t *testing.T) {
	cfg := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(time.Minute),
			integration.WithMaxScaleDownParallelism(10),
			integration.WithMaxDrainParallelism(1),
			integration.WithExpendablePodsPriorityCutoff(10),
		)
	options := cfg.ResolveOptions()

	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		n1 := tutils.BuildTestNode("n1", 100, 1000, tutils.IsReady(true))
		n2 := tutils.BuildTestNode("n2", 1000, 1000, tutils.IsReady(true))
		n2.Labels["kubernetes.io/hostname"] = "n2"
		n3 := tutils.BuildTestNode("n3", 1000, 1000, tutils.IsReady(true))

		var priority100 int32 = 100
		var priority1 int32 = 1

		ownerRef := metav1.OwnerReference{
			APIVersion: "apps/v1",
			Kind:       "ReplicaSet",
			Name:       "rs",
			UID:        "uid",
		}

		p1 := tutils.BuildTestPod("p1", 40, 0)
		p1.Spec.NodeName = "n1"
		p1.Spec.Priority = &priority1
		p1.OwnerReferences = []metav1.OwnerReference{ownerRef}

		p2 := tutils.BuildTestPod("p2", 400, 0)
		p2.Spec.NodeName = "n2"
		p2.Spec.Priority = &priority1
		p2.OwnerReferences = []metav1.OwnerReference{ownerRef}

		p3 := tutils.BuildTestPod("p3", 400, 0)
		p3.Spec.NodeName = "n2"
		p3.Spec.Priority = &priority100
		p3.OwnerReferences = []metav1.OwnerReference{ownerRef}

		p4 := tutils.BuildTestPod("p4", 500, 0, tutils.MarkUnschedulable())
		p4.Spec.Priority = &priority100
		p4.OwnerReferences = []metav1.OwnerReference{ownerRef}
		// Pin p4 to n2 so that in subsequent rounds the scheduler doesn't allocate p4 onto the upcoming node (n4).
		// If p4 takes space on the upcoming node, p6 will be unable to schedule, triggering a secondary scale-up.
		p4.Spec.NodeSelector = map[string]string{"kubernetes.io/hostname": "n2"}

		p5 := tutils.BuildTestPod("p5", 800, 0, tutils.MarkUnschedulable())
		p5.Spec.Priority = &priority100
		p5.Status.NominatedNodeName = "n3"
		p5.OwnerReferences = []metav1.OwnerReference{ownerRef}

		p6 := tutils.BuildTestPod("p6", 1000, 0, tutils.MarkUnschedulable())
		p6.Spec.Priority = &priority100
		p6.OwnerReferences = []metav1.OwnerReference{ownerRef}

		fakes.CloudProvider.AddNodeGroup("ng1", fakecloudprovider.WithMinMax(0, 10), fakecloudprovider.WithTargetSize(1), fakecloudprovider.WithNode(n1))
		fakes.CloudProvider.AddNodeGroup("ng2", fakecloudprovider.WithMinMax(0, 10), fakecloudprovider.WithTargetSize(2), fakecloudprovider.WithNode(n2), fakecloudprovider.WithNode(n3))

		fakes.K8s.AddNode(n1)
		fakes.K8s.AddNode(n2)
		fakes.K8s.AddNode(n3)

		fakes.K8s.AddPod(p1)
		fakes.K8s.AddPod(p2)
		fakes.K8s.AddPod(p3)
		fakes.K8s.AddPod(p4)
		fakes.K8s.AddPod(p5)
		fakes.K8s.AddPod(p6)

		// Scale up (trigger scale up for p4/p5/p6)
		synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)
		tg2, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()
		assert.Equal(t, 3, tg2)

		// Mark unneeded nodes (p6 goes away)
		fakes.K8s.DeletePod(p6.Namespace, p6.Name)
		fakes.CloudProvider.GetNodeGroup("ng2").(*fakecloudprovider.NodeGroup).SetTargetSize(2)
		synctestutils.MustRunOnceAfter(t, autoscaler, time.Hour*2)

		// Scale down
		fakes.K8s.DeletePod(p4.Namespace, p4.Name)
		p4 = tutils.BuildScheduledTestPod("p4", 500, 0, "n2")
		p4.Spec.Priority = &priority100
		p4.OwnerReferences = []metav1.OwnerReference{ownerRef}
		fakes.K8s.AddPod(p4)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Hour*3)

		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 0, tg1) // n1 is scaled down because p1 is expendable
	})
}
