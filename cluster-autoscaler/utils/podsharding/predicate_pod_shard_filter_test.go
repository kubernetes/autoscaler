/*
Copyright 2023 The Kubernetes Authors.

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

package podsharding

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestPredicatePodShardFilterFilterPods(t *testing.T) {
	tests := []struct {
		name             string
		selectedPodShard *PodShard
		allPodShards     []*PodShard
		pods             []*apiv1.Pod
		want             PodFilteringResult
		wantErr          bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machineTypes := []string{}
			machineTempaltes := make(map[string]*framework.NodeInfo, len(machineTypes))
			for _, machineType := range machineTypes {
				frameworkNode := framework.NewNodeInfo()
				frameworkNode.SetNode(test.BuildTestNode(machineType+"-node", 1, 1))
				machineTempaltes[machineType] = frameworkNode
			}
			predicateChecker, err := predicatechecker.NewTestPredicateChecker()
			assert.NoError(t, err)

			ctx := &context.AutoscalingContext{
				ClusterSnapshot:  clustersnapshot.NewBasicClusterSnapshot(),
				PredicateChecker: predicateChecker,
			}
			p := NewPredicatePodShardFilter()
			got, err := p.FilterPods(ctx, tt.selectedPodShard, tt.allPodShards, tt.pods)
			if (err != nil) != tt.wantErr {
				t.Errorf("PredicatePodShardFilter.FilterPods() error = %v\nwantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PredicatePodShardFilter.FilterPods() = %v\nwant %v", got, tt.want)
			}
		})
	}
}

func createTestPodShard(provReqClass string, labels []string, podUIDs ...string) *PodShard {
	podUIDsMap := make(map[types.UID]bool, len(podUIDs))
	for _, podUID := range podUIDs {
		podUIDsMap[types.UID(podUID)] = true
	}
	labelsMap := make(map[string]string, len(labels))
	for _, label := range labels {
		labelsMap[label] = "true"
	}

	return &PodShard{
		PodUids: podUIDsMap,
		NodeGroupDescriptor: NodeGroupDescriptor{
			Labels:                labelsMap,
			ProvisioningClassName: provReqClass,
		},
	}
}

func createTestPod(uid string) *apiv1.Pod {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Namespace:         "default",
			Name:              uid,
			CreationTimestamp: metav1.NewTime(time.Date(2024, 8, 9, 12, 13, 14, 0, time.UTC)),
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{},
					},
				},
			},
		},
	}
	return pod
}
