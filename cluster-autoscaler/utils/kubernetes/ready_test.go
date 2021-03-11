/*
Copyright 2021 The Kubernetes Authors.

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

package kubernetes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestGetReadiness(t *testing.T) {
	testCases := []struct {
		desc           string
		condition      apiv1.NodeConditionType
		status         apiv1.ConditionStatus
		taintKey       string
		expectedResult bool
	}{
		{"ready", apiv1.NodeReady, apiv1.ConditionTrue, "", true},
		{"unready and unready taint", apiv1.NodeReady, apiv1.ConditionFalse, apiv1.TaintNodeNotReady, false},
		{"readiness unknown and unready taint", apiv1.NodeReady, apiv1.ConditionUnknown, apiv1.TaintNodeNotReady, false},
		{"disk pressure and disk pressure taint", apiv1.NodeDiskPressure, apiv1.ConditionTrue, apiv1.TaintNodeDiskPressure, false},
		{"network unavailable and network unavailable taint", apiv1.NodeNetworkUnavailable, apiv1.ConditionTrue, apiv1.TaintNodeNetworkUnavailable, false},
		{"ready but unready taint", apiv1.NodeReady, apiv1.ConditionTrue, apiv1.TaintNodeNotReady, false},
		{"no disk pressure but disk pressure taint", apiv1.NodeDiskPressure, apiv1.ConditionFalse, apiv1.TaintNodeDiskPressure, false},
		{"network available but network unavailable taint", apiv1.NodeNetworkUnavailable, apiv1.ConditionFalse, apiv1.TaintNodeNetworkUnavailable, false},
	}
	for _, tc := range testCases {
		createTestNode := func(timeSinceCreation time.Duration) *apiv1.Node {
			node := BuildTestNode("n1", 1000, 1000)
			node.CreationTimestamp.Time = time.Time{}
			testedTime := node.CreationTimestamp.Time.Add(timeSinceCreation)

			SetNodeCondition(node, tc.condition, tc.status, testedTime)
			if tc.condition != apiv1.NodeReady {
				SetNodeCondition(node, apiv1.NodeReady, apiv1.ConditionTrue, testedTime)
			}

			if tc.taintKey != "" {
				node.Spec.Taints = []apiv1.Taint{{
					Key:       tc.taintKey,
					Effect:    apiv1.TaintEffectNoSchedule,
					TimeAdded: &metav1.Time{Time: testedTime},
				}}
			}

			return node
		}
		t.Run(tc.desc, func(t *testing.T) {
			node := createTestNode(1 * time.Minute)
			isReady, _, err := GetReadinessState(node)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, isReady)
		})
	}
}
