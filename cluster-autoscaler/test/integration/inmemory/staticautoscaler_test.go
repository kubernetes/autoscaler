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
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const (
	unneededTime = 1 * time.Minute
)

func TestStaticAutoscaler_FullLifecycle(t *testing.T) {
	testConfig := integration.NewConfig().
		WithOverrides(
			integration.WithCloudProviderName("gce"),
			integration.WithScaleDownUnneededTime(unneededTime),
		)

	integration.RunTest(t, testConfig, func(ctx *integration.TestContext) {
		ctx.BuildAutoscaler()
		fakeK8s := ctx.Fakes.K8s
		fakeCloud := ctx.Fakes.Cloud

		n1 := test.BuildTestNode("ng1-node-0", 1000, 1000, test.IsReady(true))
		fakeCloud.AddNodeGroup("ng1", fakecloudprovider.WithNode(n1))

		fakeK8s.AddPod(test.BuildScheduledTestPod("p1", 600, 100, n1.Name))
		p2 := test.BuildTestPod("p2", 600, 100, test.MarkUnschedulable())
		fakeK8s.AddPod(p2)

		ctx.MustRunOnceAfter(1 * time.Minute)

		tg1, _ := fakeCloud.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1)

		assert.Equal(t, 2, len(fakeK8s.Nodes().Items))

		fakeK8s.DeletePod(p2.Namespace, p2.Name)

		// Detection and deletion steps.
		ctx.MustRunOnceAfter(unneededTime)
		ctx.MustRunOnceAfter(unneededTime)

		finalSize, _ := fakeCloud.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 1, finalSize)
	})
}

func TestScaleUp_ResourceLimits(t *testing.T) {
	testConfig := integration.NewConfig()

	integration.RunTest(t, testConfig, func(ctx *integration.TestContext) {
		ctx.BuildAutoscaler()
		fakeK8s := ctx.Fakes.K8s
		fakeCloud := ctx.Fakes.Cloud

		n1 := test.BuildTestNode("ng1-node-0", 1000, 1000, test.IsReady(true))
		fakeCloud.AddNodeGroup("ng1", fakecloudprovider.WithNode(n1))

		p1 := test.BuildTestPod("p1", 600, 100, test.MarkUnschedulable())
		fakeK8s.AddPod(p1)

		// Scale-up should be blocked.
		fakeCloud.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 1)

		ctx.MustRunOnceAfter(1 * time.Minute)
		size, _ := fakeCloud.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 1, size, "Should not scale up when max cores limit is reached")

		// Scale-up should succeed.
		fakeCloud.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 2)

		ctx.MustRunOnceAfter(1 * time.Minute)
		newSize, _ := fakeCloud.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, newSize, "Should scale up after resource limit is increased")
	})
}
