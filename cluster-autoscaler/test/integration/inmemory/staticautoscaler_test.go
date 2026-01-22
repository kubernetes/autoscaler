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
	autoscalerbuilder "k8s.io/autoscaler/cluster-autoscaler/builder"
	"k8s.io/autoscaler/cluster-autoscaler/loop"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/fake"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	unneededTime = 1 * time.Minute
)

func TestStaticAutoscaler_FullLifecycle(t *testing.T) {
	config := integration.NewConfig().
		WithOverrides(
			integration.WithCloudProviderName("gce"),
			integration.WithScaleDownUnneededTime(unneededTime),
		)

	options := config.ResolveOptions()
	kubeClient := fake.NewClientset()
	fakeK8s := fakek8s.NewKubernetes(kubeClient, informers.NewSharedInformerFactory(kubeClient, 0))
	fakeCloud := fakecloudprovider.NewCloudProvider(fakeK8s)

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer integration.TearDown(cancel)

		autoscaler, _, err := autoscalerbuilder.New(options).
			WithKubeClient(kubeClient).
			WithInformerFactory(fakeK8s.InformerFactory).
			WithCloudProvider(fakeCloud).
			WithPodObserver(&loop.UnschedulablePodObserver{}).
			Build(ctx, integration.DebuggingSnapshotter(false))

		assert.NoError(t, err)

		n := test.BuildTestNode("ng1-node-0", 1000, 1000, test.IsReady(true))
		fakeCloud.AddNodeGroup("ng1", fakecloudprovider.WithNode(n))
		fakeK8s.AddPod(test.BuildScheduledTestPod("p1", 600, 100, n.Name))

		p := test.BuildTestPod("p2", 600, 100, test.MarkUnschedulable())
		fakeK8s.AddPod(p)

		integration.MustRunOnceAfter(t, autoscaler, unneededTime)

		tg1, _ := fakeCloud.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 2, tg1)

		assert.Equal(t, 2, len(fakeK8s.Nodes().Items))

		fakeK8s.DeletePod(p.Namespace, p.Name)

		// Detection and deletion steps.
		integration.MustRunOnceAfter(t, autoscaler, unneededTime)
		integration.MustRunOnceAfter(t, autoscaler, unneededTime)

		finalSize, _ := fakeCloud.GetNodeGroup("ng1").TargetSize()
		assert.Equal(t, 1, finalSize)
	})
}

func TestScaleUp_ResourceLimits(t *testing.T) {
	config := integration.NewConfig()

	options := config.ResolveOptions()
	kubeClient := fake.NewClientset()
	fakeK8s := fakek8s.NewKubernetes(kubeClient, informers.NewSharedInformerFactory(kubeClient, 0))
	fakeCloud := fakecloudprovider.NewCloudProvider(fakeK8s)

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer integration.TearDown(cancel)

		autoscaler, _, err := autoscalerbuilder.New(options).
			WithKubeClient(kubeClient).
			WithInformerFactory(fakeK8s.InformerFactory).
			WithCloudProvider(fakeCloud).
			WithPodObserver(&loop.UnschedulablePodObserver{}).
			Build(ctx, integration.DebuggingSnapshotter(false))

		assert.NoError(t, err)

		n := test.BuildTestNode("ng-node-0", 1000, 1000, test.IsReady(true))
		fakeCloud.AddNodeGroup("ng", fakecloudprovider.WithNode(n))
		fakeK8s.AddPod(test.BuildTestPod("pod", 600, 100, test.MarkUnschedulable()))

		// Scale-up should be blocked.
		fakeCloud.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 1)

		integration.MustRunOnceAfter(t, autoscaler, unneededTime)
		size, _ := fakeCloud.GetNodeGroup("ng").TargetSize()
		assert.Equal(t, 1, size, "Should not scale up when max cores limit is reached")

		// Scale-up should succeed.
		fakeCloud.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 2)

		integration.MustRunOnceAfter(t, autoscaler, unneededTime)
		newSize, _ := fakeCloud.GetNodeGroup("ng").TargetSize()
		assert.Equal(t, 2, newSize, "Should scale up after resource limit is increased")
	})
}
