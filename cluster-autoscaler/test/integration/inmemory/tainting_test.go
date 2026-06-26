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
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

// TestTaintingConcurrency verifies that the ScaleDown process correctly taints and deletes nodes
// across different concurrency levels.
//
// Concurrency levels tested:
// - 5 (Default): Ensures standard behavior remains stable.
// - 10 (Medium): Validates throughput improvements without overwhelming the API server.
// - 100 (High): Tests the system's resilience under high parallel pressure, confirming that
//   the workqueue and context management handle extreme concurrency gracefully.
func TestTaintingConcurrency(t *testing.T) {
	for _, concurrency := range []int{5, 10, 100} {
		t.Run(fmt.Sprintf("concurrency-%d", concurrency), func(t *testing.T) {
			stepDuration := 10 * time.Second
			config := integration.NewTestConfig().
				WithOverrides(
					integration.WithCloudProviderName("gce"),
					integration.WithScaleDownUnneededTime(unneededTime),
					integration.WithMaxConcurrentNodesTainting(concurrency),
				)

			options := config.ResolveOptions()
			infra := integration.SetupInfrastructure(t)
			fakes := infra.Fakes

			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer synctestutils.TearDown(cancel)

				autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
				assert.NoError(t, err)

				// Create 10 empty nodes that should be deleted.
				templateNode := test.BuildTestNode("template", 1000, 1000, test.IsReady(true))
				fakes.CloudProvider.AddNodeGroup("ng1", fakecloudprovider.WithNGSize(0, 20), fakecloudprovider.WithNodes(templateNode, 10))

				// Run CA loop to mark nodes as unneeded.
				synctestutils.MustRunOnceAfter(t, autoscaler, stepDuration)
				
				// Run another CA loop after unneededTime, nodes should be deleted.
				synctestutils.MustRunOnceAfter(t, autoscaler, unneededTime+time.Nanosecond)

				// Check if nodes are gone.
				assert.Eventually(t, func() bool {
					return len(fakes.K8s.Nodes().Items) == 0
				}, 5*time.Second, 100*time.Millisecond, "All nodes should be deleted")
			})
		})
	}
}
