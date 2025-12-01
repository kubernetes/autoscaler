/*
Copyright 2025 The Kubernetes Authors.

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

package binpacking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testingclock "k8s.io/utils/clock/testing"
)

func TestTimeLimiter(t *testing.T) {
	testCases := []struct {
		name                  string
		maxBinpackingDuration time.Duration
		timeAdvanced          time.Duration
		expectStop            bool
	}{
		{
			name:                  "Time limit not exceeded",
			maxBinpackingDuration: 10 * time.Second,
			timeAdvanced:          5 * time.Second,
			expectStop:            false,
		},
		{
			name:                  "Time limit exceeded",
			maxBinpackingDuration: 10 * time.Second,
			timeAdvanced:          15 * time.Second,
			expectStop:            true,
		},
		{
			name:                  "Time limit met exactly",
			maxBinpackingDuration: 10 * time.Second,
			timeAdvanced:          10 * time.Second,
			expectStop:            false,
		},
		{
			name:                  "Zero duration",
			maxBinpackingDuration: 0,
			timeAdvanced:          1 * time.Millisecond,
			expectStop:            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClock := testingclock.NewFakeClock(time.Now())
			limiter := newTimeLimiterWithClock(tc.maxBinpackingDuration, fakeClock.Now)

			limiter.InitBinpacking(nil, nil)
			fakeClock.Step(tc.timeAdvanced)

			assert.Equal(t, tc.expectStop, limiter.StopBinpacking(nil, nil))
		})
	}
}
