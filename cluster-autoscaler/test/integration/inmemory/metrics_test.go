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

	"github.com/stretchr/testify/assert"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/component-base/metrics/testutil"
)

func TestMetrics_ScaledUpNodes(t *testing.T) {
	// Initialize fakes and configuration options.
	// This happens outside the synctest bubble to keep the setup clean.
	config := integration.NewTestConfig().WithOverrides() // override CA options if needed

	options := config.ResolveOptions()
	infra := integration.SetupInfrastructure(t)
	fakes := infra.Fakes

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		metrics.RegisterAll(true)
		// This ensures all background goroutines wake up and exit when the test finishes.
		defer synctestutils.TearDown(cancel)

		autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
		assert.NoError(t, err)

		n := test.BuildTestNode("node", 1000, 1000, test.IsReady(true))
		fakes.CloudProvider.AddNodeGroup("ng", fakecloudprovider.WithNode(n))

		fakes.K8s.AddPod(test.BuildTestPod("p1", 600, 100, test.MarkUnschedulable()))
		fakes.K8s.AddPod(test.BuildTestPod("p2", 600, 100, test.MarkUnschedulable()))

		err = synctestutils.RunOnceAfter(t, autoscaler, unneededTime)
		assert.NoError(t, err)

		testutil.AssertVectorCount(t, "cluster_autoscaler_scaled_up_nodes_total", map[string]string{}, 2)
	})
}
