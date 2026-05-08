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

package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/annotations"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const customNodeGroupLabel = "custom-nodegroup"

func withNodeSelector(selector map[string]string) func(*apiv1.Pod) {
	return func(pod *apiv1.Pod) {
		pod.Spec.NodeSelector = selector
	}
}

func TestStaticAutoscalerSalvoScaleUp(t *testing.T) {
	setupSalvoTest := func(t *testing.T, salvoEnabled bool, salvoBudget time.Duration) (
		*StaticAutoscaler,
		*onScaleUpMock,
	) {
		n1 := BuildTestNode("n1", 1000, 1000)
		n1.Labels = map[string]string{customNodeGroupLabel: "ng1"}
		SetNodeReadyState(n1, true, time.Now())

		n2 := BuildTestNode("n2", 1000, 1000)
		n2.Labels = map[string]string{customNodeGroupLabel: "ng2"}
		SetNodeReadyState(n2, true, time.Now())

		tn1 := BuildTestNode("tn1", 1000, 1000)
		tn1.Labels = map[string]string{customNodeGroupLabel: "ng1"}
		tni1 := framework.NewTestNodeInfo(tn1)

		tn2 := BuildTestNode("tn2", 1000, 1000)
		tn2.Labels = map[string]string{customNodeGroupLabel: "ng2"}
		tni2 := framework.NewTestNodeInfo(tn2)

		mocks := newCommonMocks()
		mocks.readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
		mocks.allNodeLister.SetNodes([]*apiv1.Node{n1, n2})

		busy1 := BuildTestPod("busy1", 1000, 1000, WithNodeName("n1"))
		busy2 := BuildTestPod("busy2", 1000, 1000, WithNodeName("n2"))
		p1 := BuildTestPod("p1", 1000, 1000, MarkUnschedulable(), withNodeSelector(map[string]string{customNodeGroupLabel: "ng1"}))
		p2 := BuildTestPod("p2", 1000, 1000, MarkUnschedulable(), withNodeSelector(map[string]string{customNodeGroupLabel: "ng2"}))

		mocks.allPodLister.On("List").Return([]*apiv1.Pod{busy1, busy2, p1, p2}, nil).Once()
		mocks.daemonSetLister.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Once()
		mocks.podDisruptionBudgetLister.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

		autoscaler, err := setupAutoscaler(&autoscalerSetupConfig{
			nodeGroups: []*nodeGroup{
				{name: "ng1", nodes: []*apiv1.Node{n1}, template: tni1, min: 1, max: 10},
				{name: "ng2", nodes: []*apiv1.Node{n2}, template: tni2, min: 1, max: 10},
			},
			autoscalingOptions: config.AutoscalingOptions{
				NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
					ScaleDownUnneededTime:         time.Minute,
					ScaleDownUnreadyTime:          time.Minute,
					ScaleDownUtilizationThreshold: 0.5,
					MaxNodeProvisionTime:          10 * time.Second,
				},
				EstimatorName:           estimator.BinpackingEstimatorName,
				EnforceNodeGroupMinSize: true,

				MaxNodesTotal:                  10,
				MaxCoresTotal:                  100,
				MaxMemoryTotal:                 100000,
				MaxNodeGroupBinpackingDuration: 1 * time.Second,
				SalvoScaleUp:                   salvoEnabled,
				SalvoScaleUpBudget:             salvoBudget,
			},
			clusterStateConfig: clusterstate.ClusterStateRegistryConfig{
				OkTotalUnreadyCount: 1,
			},
			mocks:        mocks,
			nodesDeleted: make(chan bool, 1),
		})
		assert.NoError(t, err)
		autoscaler.initialized = true

		return autoscaler, mocks.onScaleUp
	}

	t.Run("Salvo disabled - only single scale up is triggered", func(t *testing.T) {
		autoscaler, onScaleUpMock := setupSalvoTest(t, false, 1*time.Minute)

		// Legacy behavior expects exactly one scale-up matching ng1 or ng2
		onScaleUpMock.On("ScaleUp", mock.MatchedBy(func(id string) bool {
			return id == "ng1" || id == "ng2"
		}), 1).Return(nil).Once()

		err := autoscaler.RunOnce(time.Now())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, onScaleUpMock)
	})

	t.Run("Salvo enabled - multiple sequential scale ups are triggered in a single loop", func(t *testing.T) {
		autoscaler, onScaleUpMock := setupSalvoTest(t, true, 1*time.Minute)

		// Both scale ups should be triggered sequentially in a single RunOnce loop
		onScaleUpMock.On("ScaleUp", "ng1", 1).Return(nil).Once()
		onScaleUpMock.On("ScaleUp", "ng2", 1).Return(nil).Once()

		err := autoscaler.RunOnce(time.Now())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, onScaleUpMock)

		// Verify that virtual upcoming nodes were correctly injected for both node groups
		nodeInfos, snapErr := autoscaler.AutoscalingContext.ClusterSnapshot.ListNodeInfos()
		assert.NoError(t, snapErr)

		var node1Info, node2Info *framework.NodeInfo
		for _, ni := range nodeInfos {
			if ni.Node().Annotations[annotations.NodeUpcomingAnnotation] == "true" {
				if ni.Node().Labels[customNodeGroupLabel] == "ng1" {
					node1Info = ni
				} else if ni.Node().Labels[customNodeGroupLabel] == "ng2" {
					node2Info = ni
				}
			}
		}

		// Because map iteration order is not deterministic, either ng1 or ng2 will be processed first.
		// The final iteration intentionally skips snapshot injection, so exactly one of the two upcoming nodes will be present.
		node1Present := node1Info != nil
		node2Present := node2Info != nil
		assert.True(t, node1Present != node2Present, "Expected exactly one upcoming node to be present in snapshot")

		var presentNodeInfo *framework.NodeInfo
		var expectedPodName string
		if node1Present {
			presentNodeInfo = node1Info
			expectedPodName = "p1"
		} else {
			presentNodeInfo = node2Info
			expectedPodName = "p2"
		}

		pods := presentNodeInfo.Pods()
		assert.Len(t, pods, 1)
		assert.Equal(t, expectedPodName, pods[0].Name)
	})

	t.Run("Salvo enabled - zero budget timeout halts the loop", func(t *testing.T) {
		autoscaler, onScaleUpMock := setupSalvoTest(t, true, 0*time.Nanosecond)

		// Only the first scale up should be triggered before the budget is exhausted
		onScaleUpMock.On("ScaleUp", mock.Anything, 1).Return(nil).Once()

		err := autoscaler.RunOnce(time.Now())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, onScaleUpMock)
	})
}
