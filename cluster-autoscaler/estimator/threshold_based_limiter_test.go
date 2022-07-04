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

func resetLimiter(t *testing.T, l EstimationLimiter) {
	l.EndEstimation()
	l.StartEstimation([]*apiv1.Pod{}, nil)
}

func TestThresholdBasedLimiter(t *testing.T) {
	testCases := []struct {
		name            string
		maxNodes        int
		maxDuration     time.Duration
		startDelta      time.Duration
		operations      []limiterOperation
		expectNodeCount int
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := &thresholdBasedEstimationLimiter{
				maxNodes:    tc.maxNodes,
				maxDuration: tc.maxDuration,
			}
			limiter.StartEstimation([]*apiv1.Pod{}, nil)

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
