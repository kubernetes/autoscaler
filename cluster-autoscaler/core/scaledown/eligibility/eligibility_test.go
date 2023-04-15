/*
Copyright 2022 The Kubernetes Authors.

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

package eligibility

import (
	"strconv"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFilterOutUnremovable(t *testing.T) {
	now := time.Now()

	regularNode := BuildTestNode("regular", 1000, 10)
	SetNodeReadyState(regularNode, true, time.Time{})

	justDeletedNode := BuildTestNode("justDeleted", 1000, 10)
	justDeletedNode.Spec.Taints = []apiv1.Taint{{Key: taints.ToBeDeletedTaint, Value: strconv.FormatInt(now.Unix()-30, 10)}}
	SetNodeReadyState(justDeletedNode, true, time.Time{})

	noScaleDownNode := BuildTestNode("noScaleDown", 1000, 10)
	noScaleDownNode.Annotations = map[string]string{ScaleDownDisabledKey: "true"}
	SetNodeReadyState(noScaleDownNode, true, time.Time{})

	unreadyNode := BuildTestNode("unready", 1000, 10)
	SetNodeReadyState(unreadyNode, false, time.Time{})

	bigPod := BuildTestPod("bigPod", 600, 0)
	bigPod.Spec.NodeName = "regular"

	smallPod := BuildTestPod("smallPod", 100, 0)
	smallPod.Spec.NodeName = "regular"

	testCases := []struct {
		desc             string
		nodes            []*apiv1.Node
		pods             []*apiv1.Pod
		want             []string
		scaleDownUnready bool
	}{
		{
			desc:             "regular node stays",
			nodes:            []*apiv1.Node{regularNode},
			want:             []string{"regular"},
			scaleDownUnready: true,
		},
		{
			desc:             "recently deleted node is filtered out",
			nodes:            []*apiv1.Node{regularNode, justDeletedNode},
			want:             []string{"regular"},
			scaleDownUnready: true,
		},
		{
			desc:             "marked no scale down is filtered out",
			nodes:            []*apiv1.Node{noScaleDownNode, regularNode},
			want:             []string{"regular"},
			scaleDownUnready: true,
		},
		{
			desc:             "highly utilized node is filtered out",
			nodes:            []*apiv1.Node{regularNode},
			pods:             []*apiv1.Pod{bigPod},
			want:             []string{},
			scaleDownUnready: true,
		},
		{
			desc:             "underutilized node stays",
			nodes:            []*apiv1.Node{regularNode},
			pods:             []*apiv1.Pod{smallPod},
			want:             []string{"regular"},
			scaleDownUnready: true,
		},
		{
			desc:             "unready node stays",
			nodes:            []*apiv1.Node{unreadyNode},
			want:             []string{"unready"},
			scaleDownUnready: true,
		},
		{
			desc:             "unready node is filtered oud when scale-down of unready is disabled",
			nodes:            []*apiv1.Node{unreadyNode},
			want:             []string{},
			scaleDownUnready: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			c := NewChecker(&staticThresholdGetter{0.5})
			options := config.AutoscalingOptions{
				UnremovableNodeRecheckTimeout: 5 * time.Minute,
				ScaleDownUnreadyEnabled:       tc.scaleDownUnready,
			}
			provider := testprovider.NewTestCloudProvider(nil, nil)
			provider.AddNodeGroup("ng1", 1, 10, 2)
			for _, n := range tc.nodes {
				provider.AddNode("ng1", n)
			}
			context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil, nil)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, tc.nodes, tc.pods)
			if err != nil {
				t.Fatalf("Could not create autoscaling context: %v", err)
			}
			unremovableNodes := unremovable.NewNodes()
			got, _, _ := c.FilterOutUnremovable(&context, tc.nodes, now, unremovableNodes)
			assert.Equal(t, tc.want, got)
		})
	}
}

type staticThresholdGetter struct {
	threshold float64
}

func (s *staticThresholdGetter) GetScaleDownUtilizationThreshold(_ *context.AutoscalingContext, _ cloudprovider.NodeGroup) (float64, error) {
	return s.threshold, nil
}

func (s *staticThresholdGetter) GetScaleDownGpuUtilizationThreshold(_ *context.AutoscalingContext, _ cloudprovider.NodeGroup) (float64, error) {
	return s.threshold, nil
}
