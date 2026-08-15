/*
Copyright 2018 The Kubernetes Authors.

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

package alicloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
)

func TestBuildGenericLabels(t *testing.T) {
	template := &sgTemplate{
		InstanceType: &instanceType{
			instanceTypeID: "gn5-4c-8g",
			vcpu:           4,
			memoryInBytes:  8 * 1024 * 1024 * 1024,
			gpu:            1,
		},
		Region: "cn-hangzhou",
		Zone:   "cn-hangzhou-a",
	}
	nodeName := "virtual-node"
	labels := buildGenericLabels(template, nodeName)
	assert.Equal(t, labels[apiv1.LabelInstanceTypeStable], template.InstanceType.instanceTypeID)
}

func TestExtractLabelsFromAsg(t *testing.T) {
	template := &sgTemplate{
		InstanceType: &instanceType{
			instanceTypeID: "gn5-4c-8g",
			vcpu:           4,
			memoryInBytes:  8 * 1024 * 1024 * 1024,
			gpu:            1,
		},
		Region: "cn-hangzhou",
		Zone:   "cn-hangzhou-a",
		Tags: map[string]string{
			"workload_type": "cpu",
		},
	}
	labels := template.Tags
	assert.Equal(t, labels["workload_type"], "cpu")
}
