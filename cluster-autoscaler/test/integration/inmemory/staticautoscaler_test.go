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

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	fcp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	tu "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const (
	unneededTime = time.Minute
)

func TestStaticAutoscalerRunOnce(t *testing.T) {
	cfg := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(unneededTime),
			integration.WithMaxScaleDownParallelism(10),
		)

	options := cfg.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		n1 := tu.BuildTestNode("n1", 1000, 1000, tu.IsReady(true))
		n3 := tu.BuildTestNode("n3", 1000, 1000, tu.IsReady(true))
		n4 := tu.BuildTestNode("n4", 1000, 1000, tu.IsReady(true))

		fakes.CloudProvider.AddNodeGroup("ng1",
			fcp.WithMinMax(1, 10),
			fcp.WithTargetSize(1),
			fcp.WithNode(n1))

		p1 := tu.BuildScheduledTestPod("p1", 600, 100, n1.Name)
		p2 := tu.BuildTestPod("p2", 600, 100, tu.MarkUnschedulable())

		fakes.K8s.AddPods(p1, p2)

		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)

		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1)

		fakes.K8s.DeletePod(p2.Namespace, p2.Name)

		// Detection and deletion steps.
		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)
		synctestutils.MustRunOnceAfter(t, autoscaler, 2*unneededTime)

		tg1, _ = fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 1, tg1)

		fakes.CloudProvider.AddNodeGroup("ng2", fcp.WithMinMax(0, 10), fcp.WithTargetSize(1), fcp.WithNode(n3))
		fakes.K8s.DeleteNode(n3.Name)

		// Detection and deletion steps.
		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)
		synctestutils.MustRunOnceAfter(t, autoscaler, 2*unneededTime)

		tg2, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()
		assert.Equal(t, 0, tg2)

		// Verify scale up to node group min size when cluster is empty.
		fakes.K8s.DeletePod(p1.Namespace, p1.Name)

		fakes.CloudProvider.AddNodeGroup("ng3", fcp.WithMinMax(3, 10), fcp.WithTargetSize(1), fcp.WithNode(n4))
		fakes.K8s.AddNode(n4)

		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)

		tg3, _ := fakes.CloudProvider.GetNodeGroup("ng3").TargetSize()
		assert.Equal(t, 3, tg3)
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

		n := tu.BuildTestNode("ng-node-0", 1000, 1000, tu.IsReady(true))
		fakes.CloudProvider.AddNodeGroup("ng", fcp.WithNode(n))
		fakes.K8s.AddPod(tu.BuildTestPod("pod", 600, 100, tu.MarkUnschedulable()))

		fakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 1)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)
		size, _ := fakes.CloudProvider.GetNodeGroup("ng").TargetSize()
		assert.Equal(t, 1, size, "Should not scale up when max cores limit is reached")

		fakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 2)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)
		newSize, _ := fakes.CloudProvider.GetNodeGroup("ng").TargetSize()
		assert.Equal(t, 2, newSize, "Should scale up after resource limit is increased")
	})
}

// Setup common infra for scale down delay tests
func setupScaleDownDelayTest(t *testing.T) (*integration.TestConfig, *integration.TestInfrastructure, context.Context, context.CancelFunc, *integration.FakeSet) {
	config := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(time.Minute),
			integration.WithScaleDownDelayAfterAdd(5*time.Minute),
			integration.WithScaleDownDelayAfterDelete(5*time.Minute),
			integration.WithScaleDownDelayAfterFailure(5*time.Minute),
			integration.WithMaxScaleDownParallelism(10),
			integration.WithMaxDrainParallelism(1),
		)

	infra := integration.SetupInfrastructure(t)
	ctx, cancel := context.WithCancel(context.Background())
	err := infra.StartAndSyncInformers(ctx)
	assert.NoError(t, err)

	return config, infra, ctx, cancel, infra.Fakes
}

func TestStaticAutoscalerScaleDownDelayAfterAdd(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		config, infra, ctx, cancel, fakes := setupScaleDownDelayTest(t)
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(config.ResolveOptions(), infra).Build(ctx)
		assert.NoError(t, err)

		n1 := tu.BuildTestNode("n1", 1000, 1000, tu.IsReady(true))
		n1.Labels["ng"] = "ng1"
		fakes.CloudProvider.AddNodeGroup("ng1", fcp.WithMinMax(0, 5), fcp.WithTargetSize(1), fcp.WithNode(n1))

		n2 := tu.BuildTestNode("n2", 1000, 1000, tu.IsReady(true))
		n2.Labels["ng"] = "ng2"
		fakes.CloudProvider.AddNodeGroup("ng2", fcp.WithMinMax(0, 5), fcp.WithTargetSize(1), fcp.WithNode(n2))

		p1 := tu.BuildScheduledTestPod("p1-n1", 600, 100, "n1")
		p2 := tu.BuildScheduledTestPod("p2-n2", 600, 100, "n2")
		p3 := tu.BuildTestPod("p-ng1", 600, 100, tu.MarkUnschedulable())
		p3.Spec.NodeSelector = map[string]string{"ng": "ng1"}

		fakes.K8s.AddNodes(n1, n2).AddPods(p1, p2, p3)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1)

		fakes.K8s.DeletePod(p3.Namespace, p3.Name)
		fakes.K8s.DeletePod(p1.Namespace, p1.Name)
		fakes.K8s.DeletePod(p2.Namespace, p2.Name)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
		synctestutils.MustRunOnceAfter(t, autoscaler, 2*time.Minute)

		tg1After, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		tg2After, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()

		assert.Equal(t, 2, tg1After, "ng1 scales up, but is delayed from downscaling because of AfterAdd constraint")
		assert.Equal(t, 0, tg2After, "ng2 scales down immediately")
	})
}

func TestStaticAutoscalerScaleDownDelayAfterDelete(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		config, infra, ctx, cancel, fakes := setupScaleDownDelayTest(t)
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(config.ResolveOptions(), infra).Build(ctx)
		assert.NoError(t, err)

		n1 := tu.BuildTestNode("n1", 1000, 1000, tu.IsReady(true))
		n1.Labels["ng"] = "ng1"
		fakes.CloudProvider.AddNodeGroup("ng1", fcp.WithMinMax(0, 5), fcp.WithTargetSize(1), fcp.WithNode(n1))

		n2 := tu.BuildTestNode("n2", 1000, 1000, tu.IsReady(true))
		n2.Labels["ng"] = "ng2"
		fakes.CloudProvider.AddNodeGroup("ng2", fcp.WithMinMax(0, 5), fcp.WithTargetSize(1), fcp.WithNode(n2))

		p1 := tu.BuildScheduledTestPod("p1-n1", 600, 100, "n1")
		p2 := tu.BuildScheduledTestPod("p2-n2", 600, 100, "n2")
		p3 := tu.BuildTestPod("p-ng2", 600, 100, tu.MarkUnschedulable())
		p3.Spec.NodeSelector = map[string]string{"ng": "ng2"}

		fakes.K8s.AddNodes(n1, n2).AddPods(p1, p2, p3)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

		tg2, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()
		assert.Equal(t, 2, tg2)

		fakes.K8s.DeletePod(p3.Namespace, p3.Name)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
		for i := 0; i < 7; i++ {
			synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)
		}

		tg2, _ = fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()
		assert.Equal(t, 1, tg2)

		fakes.K8s.DeletePod(p1.Namespace, p1.Name)
		fakes.K8s.DeletePod(p2.Namespace, p2.Name)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
		synctestutils.MustRunOnceAfter(t, autoscaler, 2*time.Minute)

		tg1After, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		tg2After, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()

		assert.Equal(t, 0, tg1After, "ng1 scales down cleanly")
		assert.Equal(t, 1, tg2After, "ng2 scaled down already, now it's delayed")
	})
}

func TestStaticAutoscalerScaleDownDelayAfterFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		config, infra, ctx, cancel, fakes := setupScaleDownDelayTest(t)
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(config.ResolveOptions(), infra).Build(ctx)
		assert.NoError(t, err)

		n1 := tu.BuildTestNode("n1", 1000, 1000, tu.IsReady(true))
		n1.Labels["ng"] = "ng1"
		fakes.CloudProvider.AddNodeGroup("ng1", fcp.WithMinMax(0, 5), fcp.WithTargetSize(1), fcp.WithNode(n1))

		n2 := tu.BuildTestNode("n2", 1000, 1000, tu.IsReady(true))
		n2.Labels["ng"] = "ng2"
		fakes.CloudProvider.AddNodeGroup("ng2", fcp.WithMinMax(0, 5), fcp.WithTargetSize(1), fcp.WithNode(n2))

		p1 := tu.BuildScheduledTestPod("p1-n1", 600, 100, "n1")
		p2 := tu.BuildScheduledTestPod("p2-n2", 600, 100, "n2")
		p3 := tu.BuildTestPod("p-ng1", 600, 100, tu.MarkUnschedulable())
		p3.Spec.NodeSelector = map[string]string{"ng": "ng1"}

		fakes.K8s.AddNodes(n1, n2).AddPods(p1, p2, p3)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1)

		fakes.K8s.DeletePod(p3.Namespace, p3.Name)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

		ng1 := fakes.CloudProvider.GetNodeGroup("ng1").(*fcp.NodeGroup)
		originalDelete := ng1.DeleteNodesOverride
		ng1.DeleteNodesOverride = func(nodes []*apiv1.Node) error {
			return fmt.Errorf("simulated scale down failure")
		}

		for i := 0; i < 7; i++ {
			synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)
		}

		tg1, _ = fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1)

		ng1.DeleteNodesOverride = originalDelete

		fakes.K8s.DeletePod(p1.Namespace, p1.Name)
		fakes.K8s.DeletePod(p2.Namespace, p2.Name)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
		synctestutils.MustRunOnceAfter(t, autoscaler, 2*time.Minute)

		tg1After, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		tg2After, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()

		assert.Equal(t, 2, tg1After, "ng1 should not scale down due to simulated test failure earlier")
		assert.Equal(t, 0, tg2After, "ng2 can freely scale down")
	})
}

// Setup common config for long unregistered node tests
func setUpLongUnregisteredNodeTest(t *testing.T, forceDelete bool) (*integration.FakeSet, core.Autoscaler, context.CancelFunc) {
	cfg := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(time.Minute),
		)
	options := cfg.ResolveOptions()
	options.ForceDeleteLongUnregisteredNodes = forceDelete

	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	ctx, cancel := context.WithCancel(context.Background())

	autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
	assert.NoError(t, err)

	return fakes, autoscaler, cancel
}

func TestStaticAutoscalerRunOnceKeepALongUnregisteredNode(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		fakes, autoscaler, cancel := setUpLongUnregisteredNodeTest(t, false)
		defer synctestutils.TearDown(cancel)

		n1 := tu.BuildTestNode("n1", 1000, 1000, tu.IsReady(true))
		n2 := tu.BuildTestNode("n2", 1000, 1000, tu.IsReady(true))

		p1 := tu.BuildScheduledTestPod("p1", 600, 100, "n1")
		p2 := tu.BuildTestPod("p2", 600, 100, tu.MarkUnschedulable())

		fakes.CloudProvider.AddNodeGroup("ng1", fcp.WithMinMax(2, 10), fcp.WithTargetSize(2), fcp.WithNode(n1))

		brokenNode := tu.BuildTestNode("broken", 1000, 1000)
		fakes.CloudProvider.AddNode("ng1", brokenNode)

		fakes.K8s.AddNodes(n1).AddPods(p1)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)
		synctestutils.MustRunOnceAfter(t, autoscaler, 15*time.Minute)

		fakes.K8s.AddPod(p2)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)
		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 3, tg1)

		fakes.K8s.AddNode(n2)
		fakes.CloudProvider.AddNode("ng1", n2)

		fakes.K8s.DeletePod(p1.Namespace, p1.Name)
		fakes.K8s.DeletePod(p2.Namespace, p2.Name)

		synctestutils.MustRunOnceAfter(t, autoscaler, 2*time.Minute)
		// check if broken node is removed.
		ng, err := fakes.CloudProvider.NodeGroupForNode(brokenNode)
		assert.NoError(t, err)
		assert.Nil(t, ng)
	})
}

func TestStaticAutoscalerRunOncePodsWithPriorities(t *testing.T) {
	cfg := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(time.Minute),
			integration.WithMaxScaleDownParallelism(10),
			integration.WithMaxDrainParallelism(1),
			// Pods with priority < 10 are considered "expendable" and won't block scale down.
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

		n1 := tu.BuildTestNode("n1", 100, 1000, tu.IsReady(true)) // Target for scale down test
		n2 := tu.BuildTestNode("n2", 1000, 1000, tu.IsReady(true), tu.WithNodeLabel("kubernetes.io/hostname", "n2"))
		n3 := tu.BuildTestNode("n3", 1000, 1000, tu.IsReady(true))

		p1 := tu.BuildTestPod("p1", 40, 0, tu.WithNodeName("n1"), tu.WithPriority(1), tu.WithControllerOwnerRef("rs", "ReplicaSet", "uid"))
		p2 := tu.BuildTestPod("p2", 400, 0, tu.WithNodeName("n2"), tu.WithPriority(1), tu.WithControllerOwnerRef("rs", "ReplicaSet", "uid"))
		p3 := tu.BuildTestPod("p3", 400, 0, tu.WithNodeName("n2"), tu.WithPriority(100), tu.WithControllerOwnerRef("rs", "ReplicaSet", "uid"))
		// p4, p5, p6 are all Pending and with High Priority.
		p4 := tu.BuildTestPod("p4", 500, 0, tu.MarkUnschedulable(), tu.WithPriority(100), tu.WithControllerOwnerRef("rs", "ReplicaSet", "uid"), tu.WithNodeSelector(map[string]string{"kubernetes.io/hostname": "n2"}))
		p5 := tu.BuildTestPod("p5", 800, 0, tu.MarkUnschedulable(), tu.WithPriority(100), tu.WithControllerOwnerRef("rs", "ReplicaSet", "uid"), tu.WithNominatedNodeName("n3"))
		p6 := tu.BuildTestPod("p6", 1000, 0, tu.MarkUnschedulable(), tu.WithPriority(100), tu.WithControllerOwnerRef("rs", "ReplicaSet", "uid"))

		fakes.CloudProvider.AddNodeGroup("ng1", fcp.WithMinMax(0, 10), fcp.WithTargetSize(1), fcp.WithNode(n1))
		fakes.CloudProvider.AddNodeGroup("ng2", fcp.WithMinMax(0, 10), fcp.WithTargetSize(2), fcp.WithNode(n2), fcp.WithNode(n3))

		fakes.K8s.AddNodes(n1, n2, n3).AddPods(p1, p2, p3, p4, p5, p6)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
		tg2, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()
		assert.Equal(t, 3, tg2, "Autoscaler should scale up ng2 to accomodate p6.")

		fakes.K8s.DeletePod(p6.Namespace, p6.Name)

		// Detection and deletion of unneeded node.
		synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)
		synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute)

		fakes.K8s.DeletePod(p4.Namespace, p4.Name)
		// p4 is now scheduled on n2.
		p4 = tu.BuildTestPod("p4", 500, 0, tu.WithNodeName("n2"), tu.WithPriority(100), tu.WithControllerOwnerRef("rs", "ReplicaSet", "uid"))
		fakes.K8s.AddPod(p4)

		synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)
		synctestutils.MustRunOnceAfter(t, autoscaler, 2*time.Minute)

		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 0, tg1) // n1 is scaled down because p1 is expendable
	})
}

func TestStaticAutoscalerInstanceCreationErrors(t *testing.T) {
	testCases := []struct {
		name                   string
		forceDeleteEnabled     bool
		forceDeleteImplemented bool
	}{
		{
			name:                   "forceDeleteEnabled=false,forceDeleteImplemented=false",
			forceDeleteEnabled:     false,
			forceDeleteImplemented: false,
		},
		{
			name:                   "forceDeleteEnabled=true,forceDeleteImplemented=false",
			forceDeleteEnabled:     true,
			forceDeleteImplemented: false,
		},
		{
			name:                   "forceDeleteEnabled=true,forceDeleteImplemented=true",
			forceDeleteEnabled:     true,
			forceDeleteImplemented: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := integration.NewTestConfig().
				WithOverrides(
					integration.WithScaleDownUnneededTime(time.Minute),
					integration.WithExpendablePodsPriorityCutoff(10),
					integration.WithForceDeleteFailedNodes(tc.forceDeleteEnabled),
					integration.WithMaxNodeGroupBinpackingDuration(1*time.Second),
					integration.WithMaxTotalUnreadyPercentage(100),
				)
			options := cfg.ResolveOptions()
			infra := integration.SetupInfrastructure(t)
			fakes := infra.Fakes

			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer synctestutils.TearDown(cancel)

				autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
				assert.NoError(t, err)

				templateNodeA := tu.BuildTestNode("A-template", 1000, 1000, tu.IsReady(true))
				fakes.CloudProvider.AddNodeGroup("A", fcp.WithMinMax(0, 6), fcp.WithTargetSize(1), fcp.WithNode(templateNodeA))
				fakes.K8s.AddNode(templateNodeA)

				ngA := fakes.CloudProvider.GetNodeGroup("A").(*fcp.NodeGroup)

				for i := 1; i <= 5; i++ {
					p := tu.BuildTestPod(fmt.Sprintf("p-%d", i), 800, 100, tu.MarkUnschedulable())
					fakes.K8s.AddPod(p)
				}

				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

				// Helper to simulate instances for NodeGroup A
				setupNodeGroupA := func() {
					ngA.SetInstanceState("A-template", cloudprovider.InstanceRunning)
					ngA.SetInstanceState("A-node-1", cloudprovider.InstanceCreating)

					errors := []struct {
						id       string
						errClass cloudprovider.InstanceErrorClass
						errCode  string
					}{
						{"A-node-2", cloudprovider.OutOfResourcesErrorClass, "RESOURCE_POOL_EXHAUSTED"},
						{"A-node-3", cloudprovider.OutOfResourcesErrorClass, "RESOURCE_POOL_EXHAUSTED"},
						{"A-node-4", cloudprovider.OutOfResourcesErrorClass, "QUOTA"},
						{"A-node-5", cloudprovider.OtherErrorClass, "OTHER"},
					}

					for _, e := range errors {
						ngA.SetInstanceState(e.id, cloudprovider.InstanceCreating)
						ngA.SetInstanceError(e.id, &cloudprovider.InstanceErrorInfo{
							ErrorClass: e.errClass,
							ErrorCode:  e.errCode,
						})
					}
				}
				setupNodeGroupA()

				for i := 1; i <= 5; i++ {
					fakes.K8s.DeletePod("default", fmt.Sprintf("p-%d", i))
				}

				// CA acts upon unready and errored nodes natively.
				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

				instancesA, _ := ngA.Nodes()
				assert.Len(t, instancesA, 2)
				instanceIds := make(map[string]bool)
				for _, instance := range instancesA {
					instanceIds[instance.Id] = true
				}
				assert.True(t, instanceIds["A-template"])
				assert.True(t, instanceIds["A-node-1"])

				setupNodeGroupA()

				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

				instancesA, _ = ngA.Nodes()
				assert.Len(t, instancesA, 2)

				for i := 2; i <= 5; i++ {
					id := fmt.Sprintf("A-node-%d", i)
					ngA.SetInstanceState(id, cloudprovider.InstanceRunning)
					ngA.SetInstanceError(id, nil)
				}

				synctestutils.MustRunOnceAfter(t, autoscaler, time.Second)

				instancesA, _ = ngA.Nodes()
				assert.Len(t, instancesA, 6)
			})
		})
	}
}
