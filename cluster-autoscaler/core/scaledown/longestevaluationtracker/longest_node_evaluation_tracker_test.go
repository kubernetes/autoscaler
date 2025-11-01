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

package longestevaluationtracker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLongestUnprocessedNodeScaleDownTime(t *testing.T) {
	type testCase struct {
		name                         string
		unprocessedNodes             [][]string
		wantLongestScaleDownEvalTime []time.Duration
	}
	start := time.Now()
	testCases := []testCase{
		{
			name:             "All nodes processed in the first iteration",
			unprocessedNodes: [][]string{nil, {"n1", "n2"}, {"n2", "n3"}, {}, {}},
			wantLongestScaleDownEvalTime: []time.Duration{time.Duration(0), time.Duration(1 * time.Second), time.Duration(2 * time.Second),
				time.Duration(3 * time.Second), time.Duration(0)},
		},
		{
			name:             "Not all nodes processed in the first iteration",
			unprocessedNodes: [][]string{{"n1", "n2"}, {"n1", "n2"}, {"n2", "n3"}, {}, {}},
			wantLongestScaleDownEvalTime: []time.Duration{time.Duration(1 * time.Second), time.Duration(2 * time.Second), time.Duration(3 * time.Second),
				time.Duration(4 * time.Second), time.Duration(0)},
		},
		{
			name:             "Different nodes processed in each iteration",
			unprocessedNodes: [][]string{{"n1"}, {"n2"}, {"n3"}, {"n4"}, {}, {}},
			wantLongestScaleDownEvalTime: []time.Duration{time.Duration(1 * time.Second), time.Duration(2 * time.Second), time.Duration(2 * time.Second),
				time.Duration(2 * time.Second), time.Duration(2 * time.Second), time.Duration(0)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			timestamp := start
			longestScaleDownEvalT := NewLongestNodeScaleDownEvalTime(start)
			for i := 0; i < len(tc.unprocessedNodes); i++ {
				timestamp = timestamp.Add(1 * time.Second)
				assert.Equal(t, tc.wantLongestScaleDownEvalTime[i], longestScaleDownEvalT.Update(tc.unprocessedNodes[i], timestamp))
				assert.Equal(t, len(tc.unprocessedNodes[i]), len(longestScaleDownEvalT.NodeNamesWithTimeStamps))
			}
		})
	}
}
