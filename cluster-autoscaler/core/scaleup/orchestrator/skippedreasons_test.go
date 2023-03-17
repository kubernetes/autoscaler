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

package orchestrator

import (
	"reflect"
	"testing"
)

func TestMaxResourceLimitReached(t *testing.T) {
	tests := []struct {
		name        string
		resources   []string
		wantReasons []string
	}{
		{
			name:        "simple test",
			resources:   []string{"gpu"},
			wantReasons: []string{"max cluster gpu limit reached"},
		},
		{
			name:        "multiple resources",
			resources:   []string{"gpu1", "gpu3", "tpu", "ram"},
			wantReasons: []string{"max cluster gpu1, gpu3, tpu, ram limit reached"},
		},
		{
			name:        "no resources",
			wantReasons: []string{"max cluster  limit reached"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxResourceLimitReached(tt.resources); !reflect.DeepEqual(got.Reasons(), tt.wantReasons) {
				t.Errorf("MaxResourceLimitReached(%v) = %v, want %v", tt.resources, got.Reasons(), tt.wantReasons)
			}
		})
	}
}
