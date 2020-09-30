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

package azure

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"testing"
)

func TestExtractLabelsFromScaleSet(t *testing.T) {
	expectedNodeLabelKey := "zip"
	expectedNodeLabelValue := "zap"
	extraNodeLabelValue := "buzz"
	blankString := ""

	tags := map[string]*string{
		fmt.Sprintf("%s%s", nodeLabelTagName, expectedNodeLabelKey): &expectedNodeLabelValue,
		"fizz": &extraNodeLabelValue,
		"bip":  &blankString,
	}

	labels := extractLabelsFromScaleSet(tags)
	assert.Len(t, labels, 1)
	assert.Equal(t, expectedNodeLabelValue, labels[expectedNodeLabelKey])
}

func TestExtractTaintsFromScaleSet(t *testing.T) {
	noScheduleTaintValue := "foo:NoSchedule"
	noExecuteTaintValue := "bar:NoExecute"
	preferNoScheduleTaintValue := "fizz:PreferNoSchedule"
	noSplitTaintValue := "some_value"
	blankTaintValue := ""
	regularTagValue := "baz"

	tags := map[string]*string{
		fmt.Sprintf("%s%s", nodeTaintTagName, "dedicated"):                          &noScheduleTaintValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "group"):                              &noExecuteTaintValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "app"):                                &preferNoScheduleTaintValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "k8s.io_testing_underscore_to_slash"): &preferNoScheduleTaintValue,
		"bar": &regularTagValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "blank"):   &blankTaintValue,
		fmt.Sprintf("%s%s", nodeTaintTagName, "nosplit"): &noSplitTaintValue,
	}

	expectedTaints := []apiv1.Taint{
		{
			Key:    "dedicated",
			Value:  "foo",
			Effect: apiv1.TaintEffectNoSchedule,
		},
		{
			Key:    "group",
			Value:  "bar",
			Effect: apiv1.TaintEffectNoExecute,
		},
		{
			Key:    "app",
			Value:  "fizz",
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
		{
			Key:    "k8s.io/testing/underscore/to/slash",
			Value:  "fizz",
			Effect: apiv1.TaintEffectPreferNoSchedule,
		},
	}

	taints := extractTaintsFromScaleSet(tags)
	assert.Len(t, taints, 4)
	assert.Equal(t, makeTaintSet(expectedTaints), makeTaintSet(taints))
}

func makeTaintSet(taints []apiv1.Taint) map[apiv1.Taint]bool {
	set := make(map[apiv1.Taint]bool)
	for _, taint := range taints {
		set[taint] = true
	}
	return set
}
