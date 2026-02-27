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
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	fcp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	tu "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const (
	unneededTime = time.Minute
)

// TestStaticAutoscalerRunOnce is a comprehensive integration test that verifies the core
// lifecycle events of the Autoscaler's RunOnce loop in sequence:
// 1. Scale-up triggered by unschedulable pods.
// 2. Scale-down of empty/under-utilized nodes after pods are deleted.
// 3. Identification and removal of unregistered nodes (nodes in cloud provider but missing in K8s).
// 4. Scaling up a node group to satisfy its minimum size constraint.
func TestStaticAutoscalerRunOnce(t *testing.T) {
	cfg := integration.NewTestConfig().
		WithOverrides(
			integration.WithScaleDownUnneededTime(unneededTime),
			integration.WithMaxScaleDownParallelism(10),
			integration.WithMaxNodeGroupBinpackingDuration(time.Second),
		)

	options := cfg.ResolveOptions()

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer synctestutils.TearDown(cancel)

		infra := integration.SetupInfrastructure(t)
		fakes := infra.Fakes

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		n1 := tu.BuildTestNode("n1", 1000, 1000, tu.IsReady(true))
		n2 := tu.BuildTestNode("n2", 1000, 1000, tu.IsReady(true))
		n3 := tu.BuildTestNode("n3", 1000, 1000, tu.IsReady(true))

		p1 := tu.BuildScheduledTestPod("p1", 600, 100, n1.Name)
		p2 := tu.BuildTestPod("p2", 600, 100, tu.MarkUnschedulable())
		fakes.K8s.AddPods(p1, p2)

		fakes.CloudProvider.AddNodeGroup("ng1",
			fcp.WithNGSize(1, 10),
			fcp.WithTargetSize(1),
			fcp.WithNode(n1))

		// Scale up.
		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)

		tg1, _ := fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1)

		// Scale down.
		fakes.K8s.DeletePod(p2.Namespace, p2.Name)

		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)
		synctestutils.MustRunOnceAfter(t, autoscaler, 2*unneededTime)

		tg1, _ = fakes.CloudProvider.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 1, tg1)

		// Mark and remove unregistered nodes.
		fakes.CloudProvider.AddNodeGroup("ng2", fcp.WithNGSize(0, 10), fcp.WithTargetSize(1), fcp.WithNode(n2))
		fakes.K8s.DeleteNode(n2.Name)

		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)
		synctestutils.MustRunOnceAfter(t, autoscaler, 2*unneededTime)

		tg2, _ := fakes.CloudProvider.GetNodeGroup("ng2").TargetSize()
		assert.Equal(t, 0, tg2)

		// Scale up to node group min size.
		fakes.K8s.DeletePod(p1.Namespace, p1.Name)

		fakes.CloudProvider.AddNodeGroup("ng3", fcp.WithNGSize(3, 10), fcp.WithTargetSize(1), fcp.WithNode(n3))
		fakes.K8s.AddNode(n3)

		synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime)

		tg3, _ := fakes.CloudProvider.GetNodeGroup("ng3").TargetSize()
		assert.Equal(t, 3, tg3) // 2 new nodes are supposed to be scaled up.
	})
}

// TestScaleUpResourceLimits verifies that the Autoscaler respects global resource limits (e.g., CPU cores)
// during scale-up. It confirms that scale-up is blocked when limits are reached, and resumes when the limits are increased.
func TestScaleUpResourceLimits(t *testing.T) {
	config := integration.NewTestConfig()

	options := config.ResolveOptions()

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer synctestutils.TearDown(cancel)

		infra := integration.SetupInfrastructure(t)
		fakes := infra.Fakes

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		n := tu.BuildTestNode("ng-node-0", 1000, 1000, tu.IsReady(true))
		fakes.CloudProvider.AddNodeGroup("ng", fcp.WithNode(n))
		// Add a scheduled pod to consume resources on the existing node so that the unschedulable pod doesn't fit.
		fakes.K8s.AddPod(tu.BuildScheduledTestPod("consuming-pod", 600, 100, n.Name))
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
