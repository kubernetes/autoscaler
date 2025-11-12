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

	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
)

func TestMaxResourceLimitReached(t *testing.T) {
	tests := []struct {
		name        string
		quotaID     string
		resources   []string
		wantReasons []string
	}{
		{
			name:        "simple test",
			quotaID:     "test",
			resources:   []string{"gpu"},
			wantReasons: []string{`exceeded quota: "test", resources: gpu`},
		},
		{
			name:        "multiple resources",
			quotaID:     "test",
			resources:   []string{"gpu1", "gpu3", "tpu", "ram"},
			wantReasons: []string{`exceeded quota: "test", resources: gpu1, gpu3, tpu, ram`},
		},
		{
			name:        "no resources",
			quotaID:     "test",
			resources:   []string{},
			wantReasons: []string{`exceeded quota: "test", resources: `},
		},
		{
			name:        "different quota ID",
			quotaID:     "project-quota",
			resources:   []string{"cpu"},
			wantReasons: []string{`exceeded quota: "project-quota", resources: cpu`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exceededQuotas := []resourcequotas.ExceededQuota{{ID: tt.quotaID, ExceededResources: tt.resources}}
			if got := NewMaxResourceLimitReached(exceededQuotas); !reflect.DeepEqual(got.Reasons(), tt.wantReasons) {
				t.Errorf("MaxResourceLimitReached(quotaID=%v, resources=%v) = %v, want %v", tt.quotaID, tt.resources, got.Reasons(), tt.wantReasons)
			}
		})
	}
}
