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

package resource

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
)

func TestCalculateCoresAndMemoryTotal(t *testing.T) {
	nodeConfigs := []NodeConfig{
		{"n1", 2000, 7500 * utils.MiB, 0, true, "ng1"},
		{"n2", 2000, 7500 * utils.MiB, 0, true, "ng1"},
		{"n3", 2000, 7500 * utils.MiB, 0, true, "ng1"},
		{"n4", 12000, 8000 * utils.MiB, 0, true, "ng1"},
		{"n5", 16000, 7500 * utils.MiB, 0, true, "ng1"},
		{"n6", 8000, 6000 * utils.MiB, 0, true, "ng1"},
		{"n7", 6000, 16000 * utils.MiB, 0, true, "ng1"},
	}
	nodes := make([]*apiv1.Node, len(nodeConfigs))
	for i, n := range nodeConfigs {
		node := BuildTestNode(n.Name, n.Cpu, n.Memory)
		SetNodeReadyState(node, n.Ready, time.Now())
		nodes[i] = node
	}

	nodes[6].Spec.Taints = []apiv1.Taint{
		{
			Key:    taints.ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}

	options := config.AutoscalingOptions{
		MaxCloudProviderNodeDeletionTime: 5 * time.Minute,
	}
	context, err := NewScaleTestAutoscalingContext(options, nil, nil, nil, nil, nil)
	assert.NoError(t, err)
	coresTotal, memoryTotal := coresMemoryTotal(&context, nodes, time.Now())

	assert.Equal(t, int64(42), coresTotal)
	assert.Equal(t, int64(44000*utils.MiB), memoryTotal)
}

func TestCheckDeltaWithinLimits(t *testing.T) {
	type testcase struct {
		limits            Limits
		delta             Delta
		exceededResources []string
	}
	tests := []testcase{
		{
			limits:            Limits{"a": 10},
			delta:             Delta{"a": 10},
			exceededResources: []string{},
		},
		{
			limits:            Limits{"a": 10},
			delta:             Delta{"a": 11},
			exceededResources: []string{"a"},
		},
		{
			limits:            Limits{"a": 10},
			delta:             Delta{"b": 10},
			exceededResources: []string{},
		},
		{
			limits:            Limits{"a": limitUnknown},
			delta:             Delta{"a": 0},
			exceededResources: []string{},
		},
		{
			limits:            Limits{"a": limitUnknown},
			delta:             Delta{"a": 1},
			exceededResources: []string{"a"},
		},
		{
			limits:            Limits{"a": 10, "b": 20, "c": 30},
			delta:             Delta{"a": 11, "b": 20, "c": 31},
			exceededResources: []string{"a", "c"},
		},
	}

	for _, test := range tests {
		checkResult := test.limits.CheckDeltaWithinLimits(test.delta)
		assert.Equal(t, LimitsCheckResult{test.exceededResources}, checkResult)
	}
}
