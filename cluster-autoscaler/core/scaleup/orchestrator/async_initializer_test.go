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
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
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
	scaleUpSize := 3
	failingNodeGroupName := "failing-ng"
	provider := testprovider.NewTestCloudProviderBuilder().WithOnScaleUp(func(nodeGroup string, increase int) error {
		if nodeGroup == failingNodeGroupName {
			return fmt.Errorf("Simulated error")
		}
		return nil
	}).Build()
	pod := BuildTestPod("p1", 2, 1000)
	failingNodeGroup := provider.BuildNodeGroup(failingNodeGroupName, 0, 100, 0, false, true, "T1", nil)
	successfulNodeGroup := provider.BuildNodeGroup("async-ng", 0, 100, 0, false, true, "T1", nil)
	failedScaleUpErr := errors.ToAutoscalerError(errors.CloudProviderError, fmt.Errorf("Simulated error")).AddPrefix("failed to increase node group size: ")
	testCases := []struct {
		name       string
		nodeGroup  *testprovider.TestNodeGroup
		wantStatus status.ScaleUpStatus
	}{
		{
			name:      "scale up upcoming node group",
			nodeGroup: successfulNodeGroup,
			wantStatus: status.ScaleUpStatus{
				Result: status.ScaleUpSuccessful,
				ScaleUpInfos: []nodegroupset.ScaleUpInfo{
					{
						Group:       successfulNodeGroup,
						CurrentSize: 0,
						NewSize:     scaleUpSize,
						MaxSize:     successfulNodeGroup.MaxSize(),
					},
				},
				CreateNodeGroupResults: []nodegroups.CreateNodeGroupResult{
					{MainCreatedNodeGroup: successfulNodeGroup},
				},
				PodsTriggeredScaleUp: []*apiv1.Pod{pod},
			},
		},
		{
			name:      "failing initial scale up",
			nodeGroup: failingNodeGroup,
			wantStatus: status.ScaleUpStatus{
				Result:       status.ScaleUpError,
				ScaleUpError: &failedScaleUpErr,
				CreateNodeGroupResults: []nodegroups.CreateNodeGroupResult{
					{MainCreatedNodeGroup: failingNodeGroup},
				},
				FailedResizeNodeGroups: []cloudprovider.NodeGroup{failingNodeGroup},
				PodsTriggeredScaleUp:   []*apiv1.Pod{pod},
			},
		},
	}
	listers := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	upcomingNodeGroup := provider.BuildNodeGroup("upcoming-ng", 0, 100, 0, false, true, "T1", nil)
	options := config.AutoscalingOptions{AsyncNodeGroupsEnabled: true}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, listers, provider, nil, nil)
	assert.NoError(t, err)
	option := expander.Option{NodeGroup: upcomingNodeGroup, Pods: []*apiv1.Pod{pod}}
	processors := processorstest.NewTestProcessors(&context)
	processors.AsyncNodeGroupStateChecker = &asyncnodegroups.MockAsyncNodeGroupStateChecker{IsUpcomingNodeGroup: map[string]bool{upcomingNodeGroup.Id(): true}}
	nodeInfo := framework.NewTestNodeInfo(BuildTestNode("t1", 100, 0))
	executor := newScaleUpExecutor(&context, processors.ScaleStateNotifier, processors.AsyncNodeGroupStateChecker)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scaleUpStatusProcessor := &fakeScaleUpStatusProcessor{}
			initializer := NewAsyncNodeGroupInitializer(&option, nodeInfo, executor, taints.TaintConfig{}, nil, scaleUpStatusProcessor, &context, false)
			initializer.SetTargetSize(upcomingNodeGroup.Id(), int64(scaleUpSize))
			asyncResult := nodegroups.AsyncNodeGroupCreationResult{
				CreationResult: nodegroups.CreateNodeGroupResult{MainCreatedNodeGroup: tc.nodeGroup},
				CreatedToUpcomingMapping: map[string]string{
					tc.nodeGroup.Id(): upcomingNodeGroup.Id(),
				},
			}
			initializer.InitializeNodeGroup(asyncResult)
			assert.Equal(t, *scaleUpStatusProcessor.lastStatus, tc.wantStatus)
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
