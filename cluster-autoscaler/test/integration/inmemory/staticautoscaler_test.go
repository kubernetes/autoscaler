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
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
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
