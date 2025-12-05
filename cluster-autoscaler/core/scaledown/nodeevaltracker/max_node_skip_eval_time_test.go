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

package nodeevaltracker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMaxNodeSkipEvalTime(t *testing.T) {
	type testCase struct {
		name                       string
		unprocessedNodes           [][]string
		wantMaxSkipEvalTimeSeconds []int
	}
	start := time.Now()
	testCases := []testCase{
		{
			name:                       "Only one node is skipped in one iteration",
			unprocessedNodes:           [][]string{{}, {"n1"}, {}, {}},
			wantMaxSkipEvalTimeSeconds: []int{0, 1, 0, 0},
		},
		{
			name:                       "No nodes are skipped in the first iteration",
			unprocessedNodes:           [][]string{{}, {"n1", "n2"}, {"n2", "n3"}, {}},
			wantMaxSkipEvalTimeSeconds: []int{0, 1, 2, 0},
		},
		{
			name:                       "Some nodes are skipped in the first iteration",
			unprocessedNodes:           [][]string{{"n1", "n2"}, {"n1", "n2"}, {"n2", "n3"}, {}},
			wantMaxSkipEvalTimeSeconds: []int{1, 2, 3, 0},
		},
		{
			name:                       "Overlapping node sets are skipped in different iteration",
			unprocessedNodes:           [][]string{{}, {"n1", "n2"}, {"n1"}, {"n2"}, {}},
			wantMaxSkipEvalTimeSeconds: []int{0, 1, 2, 1, 0},
		},
		{
			name:                       "Disjoint node sets are skipped in each iteration",
			unprocessedNodes:           [][]string{{"n1"}, {"n2"}, {"n3"}, {"n4"}, {}},
			wantMaxSkipEvalTimeSeconds: []int{1, 1, 1, 1, 0},
		},
		{
			name:                       "None of the nodes are skipped in each iteration",
			unprocessedNodes:           [][]string{{}, {}, {}},
			wantMaxSkipEvalTimeSeconds: []int{0, 0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			timestamp := start
			maxNodeSkipEvalTime := NewMaxNodeSkipEvalTime(start)
			for i := 0; i < len(tc.unprocessedNodes); i++ {
				timestamp = timestamp.Add(1 * time.Second)
				assert.Equal(t, time.Duration(tc.wantMaxSkipEvalTimeSeconds[i])*time.Second, maxNodeSkipEvalTime.Update(tc.unprocessedNodes[i], timestamp))
				assert.Equal(t, len(tc.unprocessedNodes[i]), len(maxNodeSkipEvalTime.nodeNamesWithTimeStamps))
			}
		})
	}
}
