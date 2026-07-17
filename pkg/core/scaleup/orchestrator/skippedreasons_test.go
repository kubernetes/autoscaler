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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/autoscaler/cluster-autoscaler/resourcequotas"
)

func TestMaxResourceLimitReached(t *testing.T) {
	tests := []struct {
		name           string
		exceededQuotas []resourcequotas.ExceededQuota
		wantReasons    []string
		wantResources  []string
	}{
		{
			name: "simple-case",
			exceededQuotas: []resourcequotas.ExceededQuota{
				{ID: "test", ExceededResources: []string{"gpu"}},
			},
			wantReasons:   []string{`exceeded quota: "test", resources: gpu`},
			wantResources: []string{"gpu"},
		},
		{
			name: "multiple-resources",
			exceededQuotas: []resourcequotas.ExceededQuota{
				{ID: "test", ExceededResources: []string{"gpu1", "gpu3", "tpu", "ram"}},
			},
			wantReasons:   []string{`exceeded quota: "test", resources: gpu1, gpu3, tpu, ram`},
			wantResources: []string{"gpu1", "gpu3", "tpu", "ram"},
		},
		{
			name: "multiple-exceeded-quotas",
			exceededQuotas: []resourcequotas.ExceededQuota{
				{ID: "cluster-quota", ExceededResources: []string{"gpu", "cpu"}},
				{ID: "other-quota", ExceededResources: []string{"cpu", "nodes"}},
			},
			wantReasons:   []string{`exceeded quota: "cluster-quota", resources: gpu, cpu`, `exceeded quota: "other-quota", resources: cpu, nodes`},
			wantResources: []string{"gpu", "cpu", "nodes"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMaxResourceLimitReached(tt.exceededQuotas)
			if diff := cmp.Diff(tt.wantReasons, got.Reasons()); diff != "" {
				t.Errorf("Resources() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantResources, got.Resources(), cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Resources() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
