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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
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
	l.StartEstimation([]PodEquivalenceGroup{}, nil, nil)
}

type dynamicThreshold struct {
	nodeLimit int
}

func (d *dynamicThreshold) DurationLimit(cloudprovider.NodeGroup, EstimationContext) time.Duration {
	return 0
}

func (d *dynamicThreshold) NodeLimit(cloudprovider.NodeGroup, EstimationContext) int {
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
			name: "binpacking is stopped if at least one threshold has negative max nodes limit",
			operations: []limiterOperation{
				expectDeny,
			},
			expectNodeCount: 0,
			thresholds: []Threshold{
				NewStaticThreshold(-1, 0),
				NewStaticThreshold(10, 0),
			},
		},
		{
			name: "binpacking is stopped if at least one threshold has negative max duration limit",
			operations: []limiterOperation{
				expectDeny,
			},
			expectNodeCount: 0,
			thresholds: []Threshold{
				NewStaticThreshold(100, -1),
				NewStaticThreshold(10, 60*time.Minute),
			},
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
			limiter.StartEstimation([]PodEquivalenceGroup{}, nil, nil)

			if tc.startDelta != time.Duration(0) {
				limiter.start = limiter.start.Add(tc.startDelta)
			}

			for _, op := range tc.operations {
				op(t, limiter)
			}
			assert.Equalf(t, tc.expectNodeCount, limiter.nodes, "Number of allowed nodes does not match expectation")
			limiter.EndEstimation()
		})
	}
}

func TestMinLimit(t *testing.T) {
	type testCase[V interface{ int | time.Duration }] struct {
		name        string
		baseLimit   V
		targetLimit V
		want        V
	}
	tests := []testCase[int]{
		{name: "At least one negative", baseLimit: -10, targetLimit: 10, want: -1},
		{name: "Negative and not set", baseLimit: -10, targetLimit: 0, want: -1},
		{name: "Both negative", baseLimit: -10, targetLimit: -10, want: -1},
		{name: "Both not set", baseLimit: 0, targetLimit: 0, want: 0},
		{name: "At least one not set", baseLimit: 0, targetLimit: 10, want: 10},
		{name: "Both set", baseLimit: 5, targetLimit: 10, want: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getMinLimit(tt.baseLimit, tt.targetLimit), "getMinLimit(%v, %v)", tt.baseLimit, tt.targetLimit)
		})
	}
	testsTime := []testCase[time.Duration]{
		{name: "At least one negative duration", baseLimit: time.Now().Sub(time.Now().Add(5 * time.Minute)), targetLimit: time.Duration(10), want: -1},
		{name: "Negative and not set durations", baseLimit: time.Now().Sub(time.Now().Add(5 * time.Minute)), targetLimit: time.Duration(0), want: -1},
		{name: "Both negative durations", baseLimit: time.Now().Sub(time.Now().Add(5 * time.Minute)), targetLimit: time.Duration(-10), want: -1},
		{name: "Both not set durations", baseLimit: time.Duration(0), targetLimit: time.Duration(0)},
		{name: "At least one not set duration", baseLimit: time.Duration(0), targetLimit: time.Duration(10), want: 10},
		{name: "Both set durations", baseLimit: time.Duration(5), targetLimit: time.Duration(10), want: 5},
	}
	for _, tt := range testsTime {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getMinLimit(tt.baseLimit, tt.targetLimit), "getMinLimit(%v, %v)", tt.baseLimit, tt.targetLimit)
		})
	}
}
