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

	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

type limiterOperation func(*testing.T, EstimationLimiter)

func expectDeny(t *testing.T, l EstimationLimiter) {
	assert.Equal(t, false, l.PermissionToAddNode())
}

func expectAllow(t *testing.T, l EstimationLimiter) {
	assert.Equal(t, true, l.PermissionToAddNode())
}

func resetLimiter(_ *testing.T, l EstimationLimiter) {
	l.EndEstimation()
	l.StartEstimation([]*apiv1.Pod{}, nil, nil)
}

type dynamicLimit struct {
	nodeLimit int
}

func (d *dynamicLimit) GetDurationLimit() time.Duration {
	return 0
}

func (d *dynamicLimit) GetNodeLimit() int {
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
		runtimeLimits   []EstimationLimit
		staticLimits    []EstimationLimit
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
			name:        "time based trigger fires",
			maxNodes:    20,
			maxDuration: 5 * time.Second,
			startDelta:  -10 * time.Second,
			operations: []limiterOperation{
				expectDeny,
				expectDeny,
			},
			expectNodeCount: 0,
		},
		{
			name:     "sequence of additions works until the threshold is hit",
			maxNodes: 3,
			operations: []limiterOperation{
				expectAllow,
				expectAllow,
				expectAllow,
				expectDeny,
			},
			expectNodeCount: 3,
		},
		{
			name:     "node counter is reset",
			maxNodes: 2,
			operations: []limiterOperation{
				expectAllow,
				expectAllow,
				expectDeny,
				resetLimiter,
				expectAllow,
			},
			expectNodeCount: 1,
		},
		{
			name:        "timer is reset",
			maxNodes:    20,
			maxDuration: 5 * time.Second,
			startDelta:  -10 * time.Second,
			operations: []limiterOperation{
				expectDeny,
				resetLimiter,
				expectAllow,
				expectAllow,
			},
			expectNodeCount: 2,
		},
		{
			name:     "node counter is reset to default value having runtime limit",
			maxNodes: 3,
			operations: []limiterOperation{
				expectAllow,
				expectAllow,
				expectDeny,
				resetLimiter, // Drops runtime limit, falls back to limit from maxNodes
				expectAllow,
				expectAllow,
				expectAllow,
				expectDeny,
			},
			expectNodeCount: 3,
			runtimeLimits:   []EstimationLimit{NewThresholdEstimationLimit(2, 0)},
		},
		{
			name:     "node cap is set to runtime limit",
			maxNodes: 0,
			operations: []limiterOperation{
				expectAllow,
				expectAllow,
				expectDeny,
				resetLimiter, // Drops runtime limits
				expectAllow,
				expectAllow,
				expectAllow,
				expectAllow,
				expectAllow,
			},
			expectNodeCount: 5,
			runtimeLimits:   []EstimationLimit{NewThresholdEstimationLimit(2, 0)},
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
			staticLimits:    []EstimationLimit{&dynamicLimit{nodeLimit: 1}},
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
			runtimeLimits:   []EstimationLimit{NewThresholdEstimationLimit(2, 60*time.Second)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var limits []EstimationLimit
			if tc.staticLimits != nil {
				limits = tc.staticLimits
			} else {
				limits = []EstimationLimit{NewThresholdEstimationLimit(tc.maxNodes, tc.maxDuration)}
			}
			limiter := NewThresholdBasedEstimationLimiter(limits).(*thresholdBasedEstimationLimiter)
			limiter.StartEstimation([]*apiv1.Pod{}, nil, tc.runtimeLimits)

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
