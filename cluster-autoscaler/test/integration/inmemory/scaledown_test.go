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
	corev1 "k8s.io/api/core/v1"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	synctestutils "k8s.io/autoscaler/cluster-autoscaler/test/integration/synctest"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	k8s_testing "k8s.io/client-go/testing"
)

func TestScaleDown_PartialFailure(t *testing.T) {
	testCases := []struct {
		name                         string
		partialTaintActuationEnabled bool
		// If PartialTaintActuationEnabled is true, the non-atomic nodegroup will partially delete.
		// If PartialTaintActuationEnabled is false, the legacy code synchronous logic is used, which aborts the entire targeted batch on any failure.
		expectedTargetSizeNonAtomic int
		expectedTargetSizeAtomic    int
	}{
		{
			name:                         "PartialTaintActuation enabled",
			partialTaintActuationEnabled: true,
			expectedTargetSizeNonAtomic:  1, // 1 node deleted, 1 node failed
			expectedTargetSizeAtomic:     2, // atomic group completely aborted
		},
		{
			name:                         "PartialTaintActuation disabled",
			partialTaintActuationEnabled: false,
			expectedTargetSizeNonAtomic:  2,
			expectedTargetSizeAtomic:     2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configBuilder := integration.NewTestConfig().
				WithOverrides(
					integration.WithScaleDownUnneededTime(time.Minute),
					func(o *config.AutoscalingOptions) {
						o.PartialTaintActuationEnabled = tc.partialTaintActuationEnabled
					},
				)

			options := configBuilder.ResolveOptions()
			infra := integration.SetupInfrastructure(t)
			fakes := infra.Fakes

			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer synctestutils.TearDown(cancel)

				autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
				assert.NoError(t, err)

				// Create atomic node group
				atomicOpts := &config.NodeGroupAutoscalingOptions{ZeroOrMaxNodeScaling: true}
				nAtomicTemplate := test.BuildTestNode("ng-atomic-node-template", 1000, 1000, test.IsReady(true))
				fakes.CloudProvider.AddNodeGroup("ng-atomic",
					fakecloudprovider.WithNodes(nAtomicTemplate, 2),
					fakecloudprovider.WithNodeGroupOptions(atomicOpts),
				)
				// Create non-atomic node group
				nNonAtomicTemplate := test.BuildTestNode("ng-nonatomic-node-template", 1000, 1000, test.IsReady(true))
				fakes.CloudProvider.AddNodeGroup("ng-nonatomic",
					fakecloudprovider.WithNodes(nNonAtomicTemplate, 2),
				)

				// Create happy-path node group with no errors
				nHappyPathTemplate := test.BuildTestNode("ng-happy-path-node-template", 1000, 1000, test.IsReady(true))
				fakes.CloudProvider.AddNodeGroup("ng-happy-path",
					fakecloudprovider.WithNodes(nHappyPathTemplate, 2),
				)

				nodeToFailAtomic := "ng-atomic-node-0"
				nodeToFailNonAtomic := "ng-nonatomic-node-0"

				// Simulate API failure when applying ToBeDeletedTaint to specific nodes
				fakes.KubeClient.Fake.PrependReactor("update", "nodes", func(action k8s_testing.Action) (handled bool, ret apimachineryruntime.Object, err error) {
					updateAction, ok := action.(k8s_testing.UpdateAction)
					if !ok {
						return false, nil, nil
					}
					node, ok := updateAction.GetObject().(*corev1.Node)
					if !ok {
						return false, nil, nil
					}

					if node.Name == nodeToFailAtomic || node.Name == nodeToFailNonAtomic {
						if taints.HasToBeDeletedTaint(node) {
							return true, nil, fmt.Errorf("simulated network error updating taint on node %s", node.Name)
						}
					}
					return false, nil, nil
				})

				// 1st loop: Marks non-atomic nodes as unneeded. Atomic node groups bypass unneeded time and attempt scale-down immediately, but fail due to taint errors.
				synctestutils.RunOnceAfter(t, autoscaler, 10*time.Second)

				// 2nd loop: timer exceeds unneeded time, deletion triggered
				synctestutils.RunOnceAfter(t, autoscaler, time.Minute+time.Second)

				// Verify expected target sizes after partial deletion
				atomicGroup := fakes.CloudProvider.GetNodeGroup("ng-atomic")
				assert.NotNil(t, atomicGroup, "expected ng-atomic group to exist")
				atomicSize, err := atomicGroup.TargetSize()
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTargetSizeAtomic, atomicSize, "atomic target size mismatch")

				nonAtomicGroup := fakes.CloudProvider.GetNodeGroup("ng-nonatomic")
				assert.NotNil(t, nonAtomicGroup, "expected ng-nonatomic group to exist")
				nonAtomicSize, err := nonAtomicGroup.TargetSize()
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTargetSizeNonAtomic, nonAtomicSize, fmt.Sprintf("nonatomic target size mismatch, PartialTaintActuation enabled: %v", tc.partialTaintActuationEnabled))

				happyPathGroup := fakes.CloudProvider.GetNodeGroup("ng-happy-path")
				assert.NotNil(t, happyPathGroup, "expected ng-happy-path group to exist")
				happyPathSize, err := happyPathGroup.TargetSize()
				assert.NoError(t, err)

				if tc.partialTaintActuationEnabled {
					assert.Equal(t, 0, happyPathSize, "expected happy path nodegroup to be fully scaled down")
				} else {
					assert.Equal(t, 2, happyPathSize, "expected happy path nodegroup to remain untouched in legacy fallback")
				}

				// Validate end state: the exact correct nodes are deleted, and all remaining nodes have no taints
				var remainingNodeNames []string
				for _, n := range fakes.K8s.Nodes().Items {
					remainingNodeNames = append(remainingNodeNames, n.Name)
					assert.False(t, taints.HasToBeDeletedTaint(&n), "node %s should have clean taints", n.Name)
				}

				if tc.partialTaintActuationEnabled {
					assert.ElementsMatch(t, []string{"ng-atomic-node-0", "ng-atomic-node-1", "ng-nonatomic-node-0"}, remainingNodeNames, "expected successfully tainted non-atomic sibling to be deleted while failed nodes remain")
				} else {
					assert.ElementsMatch(t, []string{"ng-atomic-node-0", "ng-atomic-node-1", "ng-nonatomic-node-0", "ng-nonatomic-node-1", "ng-happy-path-node-0", "ng-happy-path-node-1"}, remainingNodeNames, "with PartialTaintActuation disabled, a taint failure aborts the entire scale-down batch and rolls back all nodes to remain untouched")
				}
			})
		})
	}
}

func TestScaleDown_HappyPath(t *testing.T) {
	atomicNodeGroupConfig := fakecloudprovider.WithNodeGroupOptions(&config.NodeGroupAutoscalingOptions{ZeroOrMaxNodeScaling: true})
	nonAtomicNodeGroupConfig := fakecloudprovider.WithNodeGroupOptions(&config.NodeGroupAutoscalingOptions{ZeroOrMaxNodeScaling: false})

	testCases := []struct {
		name            string
		nodeGroupConfig fakecloudprovider.NodeGroupOption
	}{
		{
			name:            "Atomic nodegroup",
			nodeGroupConfig: atomicNodeGroupConfig,
		},
		{
			name:            "Non-atomic nodegroup",
			nodeGroupConfig: nonAtomicNodeGroupConfig,
		},
	}

	for _, tc := range testCases {
		for _, partialTaintActuationEnabled := range []bool{true, false} {
			testName := fmt.Sprintf("%s/PartialTaintActuation_enabled_%v", tc.name, partialTaintActuationEnabled)
			t.Run(testName, func(t *testing.T) {
				configBuilder := integration.NewTestConfig().
					WithOverrides(
						integration.WithScaleDownUnneededTime(time.Minute),
						func(o *config.AutoscalingOptions) {
							o.PartialTaintActuationEnabled = partialTaintActuationEnabled
						},
					)

				options := configBuilder.ResolveOptions()
				infra := integration.SetupInfrastructure(t)
				fakes := infra.Fakes

				synctest.Test(t, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer synctestutils.TearDown(cancel)

					autoscaler, _, err := integration.DefaultAutoscalingBuilder(options, infra).Build(ctx)
					assert.NoError(t, err)

					nTemplate := test.BuildTestNode("ng-node-template", 1000, 1000, test.IsReady(true))
					fakes.CloudProvider.AddNodeGroup("ng",
						fakecloudprovider.WithNodes(nTemplate, 2),
						tc.nodeGroupConfig,
					)

					// 1st loop: Marks non-atomic nodes as unneeded.
					synctestutils.MustRunOnceAfter(t, autoscaler, 10*time.Second)

					// 2nd loop: timer exceeds unneeded time, deletion triggered for non-atomic nodes
					synctestutils.MustRunOnceAfter(t, autoscaler, time.Minute+time.Second)

					// Verify expected target sizes after deletion
					group := fakes.CloudProvider.GetNodeGroup("ng")
					assert.NotNil(t, group, "expected ng group to exist")
					size, err := group.TargetSize()
					assert.NoError(t, err)
					assert.Equal(t, 0, size, "expected target size to be 0 for happy path scaledown")

					// Validate end state: all nodes are deleted
					var remainingNodeNames []string
					for _, n := range fakes.K8s.Nodes().Items {
						remainingNodeNames = append(remainingNodeNames, n.Name)
						assert.False(t, taints.HasToBeDeletedTaint(&n), "node %s should have clean taints", n.Name)
					}
					assert.Empty(t, remainingNodeNames, "all nodes should be successfully deleted")
				})
			})
		}
	}
}
