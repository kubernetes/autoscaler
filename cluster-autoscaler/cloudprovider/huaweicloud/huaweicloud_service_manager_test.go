/*
Copyright 2020 The Kubernetes Authors.

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

package huaweicloud

import (
	"reflect"
	"testing"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func Test_extractTaintsFromTags(t *testing.T) {
	tests := []struct {
		name string
		args map[string]string
		want []apiv1.Taint
	}{
		{
			name: "tag in right format",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:NoSchedule",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "bar", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "empty taint key should be ignored",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_": "bar:NoSchedule",
			},
			want: []apiv1.Taint{},
		},
		{
			name: "invalid tag key should be ignored",
			args: map[string]string{
				"invalidTagKey": "bar:NoSchedule",
			},
			want: []apiv1.Taint{},
		},
		{
			name: "invalid taint effect should be ignored",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:InvalidEffect",
			},
			want: []apiv1.Taint{},
		},
		{
			name: "empty taint value",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": ":NoSchedule",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "one tag with valid tag, one tag with invalid key, ignore the invalid one",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:NoSchedule",
				"invalidTagKey": ":NoSchedule",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "bar", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "one tag with valid key/value, one tag with invalid value, ignore the invalid one",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:NoSchedule",
				"k8s.io_cluster-autoscaler_node-template_taint_bar": "invalidTagValue",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "bar", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
		{
			name: "one tag with valid key/value, one tag with invalid value length, ignore the invalid one",
			args: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "bar:NoSchedule",
				"k8s.io_cluster-autoscaler_node-template_taint_bar": "foo:NoSchedule:more",
			},
			want: []apiv1.Taint{
				{Key: "foo", Value: "bar", Effect: apiv1.TaintEffectNoSchedule},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTaintsFromTags(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTaintsFromTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildGenericLabels(t *testing.T) {
	template := &asgTemplate{
		name:   "foo",
		region: "foo",
		zone:   "foo",
	}
	tests := []struct {
		name string
		tags map[string]string
		want map[string]string
	}{
		{
			name: "tags contain taints key, ignore it when extract labels",
			tags: map[string]string{
				"k8s.io_cluster-autoscaler_node-template_taint_foo": "true:PreferNoSchedule",
				"foo": "bar",
			},
			want: map[string]string{
				apiv1.LabelArchStable:         cloudprovider.DefaultArch,
				apiv1.LabelOSStable:           cloudprovider.DefaultOS,
				apiv1.LabelInstanceTypeStable: template.name,
				apiv1.LabelTopologyRegion:     template.region,
				apiv1.LabelTopologyZone:       template.zone,
				apiv1.LabelHostname:           "foo",
				"foo":                         "bar",
			},
		},
		{
			name: "tags don't contain taints key",
			tags: map[string]string{
				"foo": "bar",
			},
			want: map[string]string{
				apiv1.LabelArchStable:         cloudprovider.DefaultArch,
				apiv1.LabelOSStable:           cloudprovider.DefaultOS,
				apiv1.LabelInstanceTypeStable: template.name,
				apiv1.LabelTopologyRegion:     template.region,
				apiv1.LabelTopologyZone:       template.zone,
				apiv1.LabelHostname:           "foo",
				"foo":                         "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template.tags = tt.tags
			if got := buildGenericLabels(template, "foo"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildGenericLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
