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

package actuation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	kubelet_config "k8s.io/kubernetes/pkg/kubelet/apis/config"
)

func TestPriorityEvictor(t *testing.T) {
	deletedPods := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	n1 := BuildTestNode("n1", 1000, 1000)
	p1 := BuildTestPod("p1", 100, 0, WithNodeName(n1.Name))
	p2 := BuildTestPod("p2", 300, 0, WithNodeName(n1.Name))
	p3 := BuildTestPod("p3", 150, 0, WithNodeName(n1.Name))

	priority100 := int32(100)
	priority2000 := int32(2000)
	priority2000000005 := int32(2000000005)
	p1.Spec.Priority = &priority2000000005
	p2.Spec.Priority = &priority2000
	p3.Spec.Priority = &priority100

	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1beta1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		deletedPods <- eviction.Name
		return true, nil, nil
	})

	options := config.AutoscalingOptions{
		MaxGracefulTerminationSec: 20,
		MaxPodEvictionTime:        5 * time.Second,
	}
	ctx, err := NewScaleTestAutoscalingContext(options, fakeClient, nil, nil, nil, nil)
	assert.NoError(t, err)

	evictor := Evictor{
		EvictionRetryTime:   0,
		PodEvictionHeadroom: DefaultPodEvictionHeadroom,
		shutdownGracePeriodByPodPriority: []kubelet_config.ShutdownGracePeriodByPodPriority{
			{
				Priority:                   0,
				ShutdownGracePeriodSeconds: 3,
			},
			{
				Priority:                   1000,
				ShutdownGracePeriodSeconds: 2,
			},
			{
				Priority:                   2000000000,
				ShutdownGracePeriodSeconds: 1,
			},
		},
		fullDsEviction: true,
	}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, ctx.ClusterSnapshot, []*apiv1.Node{n1}, []*apiv1.Pod{p1, p2, p3})
	nodeInfo, err := ctx.ClusterSnapshot.GetNodeInfo(n1.Name)
	assert.NoError(t, err)
	_, err = evictor.DrainNode(&ctx, nodeInfo)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))

	assert.Equal(t, p3.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
	assert.Equal(t, p1.Name, deleted[2])
}

func TestGroupByPriority(t *testing.T) {
	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	p3 := BuildTestPod("p3", 150, 0)
	p4 := BuildTestPod("p4", 100, 0)
	p5 := BuildTestPod("p5", 300, 0)

	p6 := BuildTestPod("p6", 100, 0)
	p7 := BuildTestPod("p7", 300, 0)
	p8 := BuildTestPod("p8", 150, 0)
	p9 := BuildTestPod("p9", 100, 0)
	p10 := BuildTestPod("p10", 300, 0)

	priority0 := int32(0)
	priority100 := int32(100)
	priority500 := int32(500)
	priority1000 := int32(1000)
	priority2000000005 := int32(2000000005)
	p1.Spec.Priority = &priority2000000005
	p2.Spec.Priority = &priority500
	p3.Spec.Priority = &priority100
	p4.Spec.Priority = &priority0
	p5.Spec.Priority = &priority1000

	p6.Spec.Priority = &priority2000000005
	p7.Spec.Priority = &priority500
	p8.Spec.Priority = &priority100
	p9.Spec.Priority = &priority0
	p10.Spec.Priority = &priority1000

	shutdownGracePeriodByPodPriority := []kubelet_config.ShutdownGracePeriodByPodPriority{
		{
			Priority:                   10,
			ShutdownGracePeriodSeconds: 4,
		},
		{
			Priority:                   1000,
			ShutdownGracePeriodSeconds: 3,
		},
		{
			Priority:                   2000,
			ShutdownGracePeriodSeconds: 2,
		},
		{
			Priority:                   2000000000,
			ShutdownGracePeriodSeconds: 1,
		},
	}

	wantGroups := []podEvictionGroup{
		{
			ShutdownGracePeriodByPodPriority: shutdownGracePeriodByPodPriority[0],
			FullEvictionPods:                 []*apiv1.Pod{p2, p3, p4},
			BestEffortEvictionPods:           []*apiv1.Pod{p7, p8, p9},
		},
		{
			ShutdownGracePeriodByPodPriority: shutdownGracePeriodByPodPriority[1],
			FullEvictionPods:                 []*apiv1.Pod{p5},
			BestEffortEvictionPods:           []*apiv1.Pod{p10},
		},
		{
			ShutdownGracePeriodByPodPriority: shutdownGracePeriodByPodPriority[2],
		},
		{
			ShutdownGracePeriodByPodPriority: shutdownGracePeriodByPodPriority[3],
			FullEvictionPods:                 []*apiv1.Pod{p1},
			BestEffortEvictionPods:           []*apiv1.Pod{p6},
		},
	}

	groups := groupByPriority(shutdownGracePeriodByPodPriority, []*apiv1.Pod{p1, p2, p3, p4, p5}, []*apiv1.Pod{p6, p7, p8, p9, p10})
	assert.Equal(t, wantGroups, groups)
}

func TestParseShutdownGracePeriodsAndPriorities(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []kubelet_config.ShutdownGracePeriodByPodPriority
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "Incorrect string - incorrect priority grace period pairs",
			input: "1:2,34",
			want:  nil,
		},
		{
			name:  "Incorrect string - trailing ,",
			input: "1:2, 3:4,",
			want:  nil,
		},
		{
			name:  "Incorrect string - trailing space",
			input: "1:2,3:4 ",
			want:  nil,
		},
		{
			name:  "Non integers - 1",
			input: "1:2,3:a",
			want:  nil,
		},
		{
			name:  "Non integers - 2",
			input: "1:2,3:23.2",
			want:  nil,
		},
		{
			name:  "parsable input",
			input: "1:2,3:4",
			want: []kubelet_config.ShutdownGracePeriodByPodPriority{
				{1, 2},
				{3, 4},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shutdownGracePeriodByPodPriority := ParseShutdownGracePeriodsAndPriorities(tc.input)
			assert.Equal(t, tc.want, shutdownGracePeriodByPodPriority)
		})
	}
}
