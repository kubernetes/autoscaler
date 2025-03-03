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

package orchestrator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups/asyncnodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	processorstest "k8s.io/autoscaler/cluster-autoscaler/processors/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNodePoolAsyncInitialization(t *testing.T) {
	testCases := []struct {
		name             string
		failingScaleUps  map[string]bool
		expectedScaleUps map[string]int
	}{
		{
			name:             "scale up upcoming node group",
			expectedScaleUps: map[string]int{"async-ng": 3},
		},
		{
			name:            "failing initial scale up",
			failingScaleUps: map[string]bool{"async-ng": true},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scaledUpGroups := make(map[string]int)
			provider := testprovider.NewTestCloudProvider(
				func(nodeGroup string, increase int) error {
					if tc.failingScaleUps[nodeGroup] {
						return fmt.Errorf("Simulated error")
					}
					scaledUpGroups[nodeGroup] += increase
					return nil
				}, nil)
			options := config.AutoscalingOptions{
				NodeAutoprovisioningEnabled: true,
				AsyncNodeGroupsEnabled:      true,
			}
			listers := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil)
			context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
			assert.NoError(t, err)
			p1 := BuildTestPod("p1", 2, 1000)
			upcomingNodeGroup := provider.BuildNodeGroup("upcoming-ng", 0, 100, 0, false, true, "T1", nil)
			createdNodeGroup := provider.BuildNodeGroup("async-ng", 0, 100, 0, false, true, "T1", nil)
			option := expander.Option{
				NodeGroup: upcomingNodeGroup,
				Pods:      []*apiv1.Pod{p1},
			}
			processors := processorstest.NewTestProcessors(&context)
			processors.AsyncNodeGroupStateChecker = &asyncnodegroups.MockAsyncNodeGroupStateChecker{IsUpcomingNodeGroup: map[string]bool{upcomingNodeGroup.Id(): true}}
			nodeInfo := framework.NewTestNodeInfo(BuildTestNode("t1", 100, 0))
			executor := newScaleUpExecutor(&context, processors.ScaleStateNotifier, processors.AsyncNodeGroupStateChecker)
			scaleUpStatusProcessor := &fakeScaleUpStatusProcessor{}
			initializer := NewAsyncNodeGroupInitializer(&option, nodeInfo, executor, taints.TaintConfig{}, nil, scaleUpStatusProcessor, &context, false)
			initializer.SetTargetSize(upcomingNodeGroup.Id(), 3)
			asyncResult := nodegroups.AsyncNodeGroupCreationResult{
				CreationResult: nodegroups.CreateNodeGroupResult{MainCreatedNodeGroup: createdNodeGroup},
				CreatedToUpcomingMapping: map[string]string{
					createdNodeGroup.Id(): upcomingNodeGroup.Id(),
				},
			}
			initializer.InitializeNodeGroup(asyncResult)
			assert.Equal(t, len(scaledUpGroups), len(tc.expectedScaleUps))
			for groupName, increase := range tc.expectedScaleUps {
				assert.Equal(t, increase, scaledUpGroups[groupName])
			}
			if len(tc.failingScaleUps) > 0 {
				expectedErr := errors.ToAutoscalerError(errors.CloudProviderError, fmt.Errorf("Simulated error")).AddPrefix("failed to increase node group size: ")
				assert.Equal(t, scaleUpStatusProcessor.lastStatus, &status.ScaleUpStatus{
					Result:                 status.ScaleUpError,
					ScaleUpError:           &expectedErr,
					CreateNodeGroupResults: []nodegroups.CreateNodeGroupResult{asyncResult.CreationResult},
					FailedResizeNodeGroups: []cloudprovider.NodeGroup{createdNodeGroup},
					PodsTriggeredScaleUp:   option.Pods,
				})
			}
		})
	}
}

type fakeScaleUpStatusProcessor struct {
	lastStatus *status.ScaleUpStatus
}

func (f *fakeScaleUpStatusProcessor) Process(_ *context.AutoscalingContext, status *status.ScaleUpStatus) {
	f.lastStatus = status
}

func (f *fakeScaleUpStatusProcessor) CleanUp() {
}
