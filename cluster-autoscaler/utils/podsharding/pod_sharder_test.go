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

package podsharding

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNodeGroupDescriptorSignatureAndDeepCopy(t *testing.T) {
	type fields struct {
		Labels                map[string]string
		SystemLabels          map[string]string
		Taints                []apiv1.Taint
		ExtraResources        map[string]resource.Quantity
		ProvisioningClassName string
	}
	tests := []struct {
		name   string
		fields fields
		want   ShardSignature
	}{
		{
			name: "simple pod",
			fields: fields{
				Labels: map[string]string{
					"key": "value",
				},
				SystemLabels:   map[string]string{},
				ExtraResources: map[string]resource.Quantity{},
			},
			want: ShardSignature("Labels(key=value)"),
		},
		{
			name: "two labels",
			fields: fields{
				Labels: map[string]string{
					"other-key": "another-value",
					"key":       "value",
				},
				SystemLabels:   map[string]string{},
				ExtraResources: map[string]resource.Quantity{},
			},
			want: ShardSignature("Labels(key=value,other-key=another-value)"),
		},
		{
			name: "system labels",
			fields: fields{
				Labels: map[string]string{
					"key": "value",
				},
				SystemLabels: map[string]string{
					"system-key": "system-value",
					"key":        "value",
				},
				ExtraResources: map[string]resource.Quantity{},
			},
			want: ShardSignature("Labels(key=value)SystemLabels(key=value,system-key=system-value)"),
		},
		{
			name: "extra resources",
			fields: fields{
				Labels: map[string]string{
					"key": "value",
				},
				SystemLabels: map[string]string{
					"key": "value",
				},
				ExtraResources: map[string]resource.Quantity{
					"resource":         *resource.NewMilliQuantity(1500, resource.DecimalSI),
					"another-resource": *resource.NewMilliQuantity(500, resource.DecimalSI),
				},
			},
			want: ShardSignature("Labels(key=value)SystemLabels(key=value)ExtraResources(another-resource=500m,resource=1500m)"),
		},
		{
			name: "taints",
			fields: fields{
				Labels: map[string]string{
					"key": "value",
				},
				SystemLabels: map[string]string{
					"key": "value",
				},
				ExtraResources: map[string]resource.Quantity{
					"resource": *resource.NewMilliQuantity(1500, resource.DecimalSI),
				},
				Taints: []apiv1.Taint{
					{
						Key:    "key",
						Value:  "value",
						Effect: apiv1.TaintEffectPreferNoSchedule,
					},
					{
						Key:    "another-key",
						Value:  "other-value",
						Effect: apiv1.TaintEffectNoSchedule,
					},
				},
			},
			want: ShardSignature("Labels(key=value)SystemLabels(key=value)ExtraResources(resource=1500m)Taints(another-key/other-value/NoSchedule,key/value/PreferNoSchedule)"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descriptor := &NodeGroupDescriptor{
				Labels:                tt.fields.Labels,
				SystemLabels:          tt.fields.SystemLabels,
				Taints:                tt.fields.Taints,
				ExtraResources:        tt.fields.ExtraResources,
				ProvisioningClassName: tt.fields.ProvisioningClassName,
			}
			if got := descriptor.signature(); got != tt.want {
				t.Errorf("NodeGroupDescriptor.signature() = %v, want %v", got, tt.want)
			}

			original := NodeGroupDescriptor{
				Labels:                tt.fields.Labels,
				SystemLabels:          tt.fields.SystemLabels,
				Taints:                tt.fields.Taints,
				ExtraResources:        tt.fields.ExtraResources,
				ProvisioningClassName: tt.fields.ProvisioningClassName,
			}
			copy := original.DeepCopy()
			if diff := cmp.Diff(copy, original); diff != "" {
				t.Errorf("NodeGroupDescriptor.DeepCopy() diff (-copy +original): %v", diff)
			}
		})
	}
}
