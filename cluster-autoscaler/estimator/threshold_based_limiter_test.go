/*
Copyright 2016 The Kubernetes Authors.

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

package estimator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type limiterOperation func(*testing.T, EstimationLimiter)

func expectDeny(t *testing.T, l EstimationLimiter) {
	assert.Equal(t, false, l.PermissionToAddNode())
}

func expectAllow(t *testing.T, l EstimationLimiter) {
	assert.Equal(t, true, l.PermissionToAddNode())
}

func resetLimiter(t *testing.T, l EstimationLimiter) {
	l.EndEstimation()
	l.StartEstimation([]*apiv1.Pod{}, nil, nil)
}

type dynamicThreshold struct {
	nodeLimit int
}

func (d *dynamicThreshold) GetDurationLimit() time.Duration {
	return 0
}

func (d *dynamicThreshold) GetNodeLimit(cloudprovider.NodeGroup, *EstimationContext) int {
	d.nodeLimit += 1
	return d.nodeLimit
}

func TestThresholdBasedLimiter(t *testing.T) {
	testCases := []struct {
		name            string
		maxNodes        int
		maxDuration     time.Duration
		startDelta      time.Duration
		operations      []limiterOperation
		expectNodeCount int
		thresholds      []Threshold
	}{
		{
			name:     "no limiting happens",
			maxNodes: 20,
			operations: []limiterOperation{
				expectAllow,
				expectAllow,
				expectAllow,
			},
			expectNodeCount: 3,
		},
		{
			name:       "time based trigger fires",
			startDelta: -10 * time.Second,
			operations: []limiterOperation{
				expectDeny,
				expectDeny,
			},
			expectNodeCount: 0,
			thresholds:      []Threshold{NewStaticThreshold(20, 5*time.Second)},
		},
		{
			name: "sequence of additions works until the threshold is hit",
			operations: []limiterOperation{
				expectAllow,
				expectAllow,
				expectAllow,
				expectDeny,
			},
			expectNodeCount: 3,
			thresholds:      []Threshold{NewStaticThreshold(3, 0)},
		},
		{
			name: "node counter is reset",
			operations: []limiterOperation{
				expectAllow,
				expectAllow,
				expectDeny,
				resetLimiter,
				expectAllow,
			},
			expectNodeCount: 1,
			thresholds:      []Threshold{NewStaticThreshold(2, 0)},
		},
		{
			name:       "timer is reset",
			startDelta: -10 * time.Second,
			operations: []limiterOperation{
				expectDeny,
				resetLimiter,
				expectAllow,
				expectAllow,
			},
			expectNodeCount: 2,
			thresholds:      []Threshold{NewStaticThreshold(20, 5*time.Second)},
		},
		{
			name:     "handles dynamic limits",
			maxNodes: 0,
			operations: []limiterOperation{
				expectAllow,
				expectAllow,
				expectDeny,
				resetLimiter,
				expectAllow,
				expectAllow,
				expectAllow,
				expectDeny,
			},
			expectNodeCount: 3,
			thresholds:      []Threshold{&dynamicThreshold{1}},
		},
		{
			name: "duration limit is set to runtime limit",
			operations: []limiterOperation{
				expectDeny,
				expectDeny,
				resetLimiter, // resets startDelta
				expectAllow,
				expectAllow,
			},
			expectNodeCount: 2,
			startDelta:      -120 * time.Second,
			thresholds:      []Threshold{NewStaticThreshold(2, 60*time.Second)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewThresholdBasedEstimationLimiter(tc.thresholds).(*thresholdBasedEstimationLimiter)
			limiter.StartEstimation([]*apiv1.Pod{}, nil, nil)

			if tc.startDelta != time.Duration(0) {
				limiter.start = limiter.start.Add(tc.startDelta)
			}

			for _, op := range tc.operations {
				op(t, limiter)
			}
			assert.Equal(t, tc.expectNodeCount, limiter.nodes)
			limiter.EndEstimation()
		})
	}
}
