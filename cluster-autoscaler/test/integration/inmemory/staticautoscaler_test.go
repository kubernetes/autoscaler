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
