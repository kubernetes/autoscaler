/*
Copyright 2021 The Kubernetes Authors.

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

package nodegroupset

import (
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestNodeLabelComparison(t *testing.T) {
	labels := []string{"node.kubernetes.io/instance-type", "kubernetes.io/arch"}
	comparator := CreateLabelNodeInfoComparator(labels)
	node1 := BuildTestNode("node1", 1000, 2000)
	node2 := BuildTestNode("node2", 1000, 2000)

	for _, tc := range []struct {
		description string
		labels1     map[string]string
		labels2     map[string]string
		isSimilar   bool
	}{
		{
			description: "both labels match",
			labels1:     map[string]string{"node.kubernetes.io/instance-type": "m5.4xlarge", "kubernetes.io/arch": "amd64"},
			labels2:     map[string]string{"node.kubernetes.io/instance-type": "m5.4xlarge", "kubernetes.io/arch": "amd64"},
			isSimilar:   true,
		},
		{
			description: "one label doesn't match",
			labels1:     map[string]string{"node.kubernetes.io/instance-type": "m5.4xlarge", "kubernetes.io/arch": "amd64"},
			labels2:     map[string]string{"node.kubernetes.io/instance-type": "m5.4xlarge", "kubernetes.io/arch": "i386"},
			isSimilar:   false,
		},
		{
			description: "unspecified labels are not considered",
			labels1:     map[string]string{"node.kubernetes.io/instance-type": "m5.4xlarge", "kubernetes.io/arch": "amd64", "unspecified-label": "eu-west1"},
			labels2:     map[string]string{"node.kubernetes.io/instance-type": "m5.4xlarge", "kubernetes.io/arch": "amd64", "unspecified-label": "eu-west2"},
			isSimilar:   true,
		},
		{
			description: "no labels are set",
			labels1:     map[string]string{},
			labels2:     map[string]string{},
			isSimilar:   false,
		},
		{
			description: "single label matches, label is unset on second group",
			labels1:     map[string]string{"node.kubernetes.io/instance-type": "m5.4xlarge", "kubernetes.io/arch": "amd64"},
			labels2:     map[string]string{"kubernetes.io/arch": "amd64"},
			isSimilar:   false,
		},
		{
			description: "single label matches, label is unset on first group",
			labels1:     map[string]string{"kubernetes.io/arch": "amd64"},
			labels2:     map[string]string{"node.kubernetes.io/instance-type": "m5.4xlarge", "kubernetes.io/arch": "amd64"},
			isSimilar:   false,
		},
		{
			description: "labels are explicitly set to be empty",
			labels1:     map[string]string{"node.kubernetes.io/instance-type": "", "kubernetes.io/arch": ""},
			labels2:     map[string]string{"node.kubernetes.io/instance-type": "", "kubernetes.io/arch": ""},
			isSimilar:   true,
		},
		{
			description: "one labels is explicitly set to be empty",
			labels1:     map[string]string{"node.kubernetes.io/instance-type": "", "kubernetes.io/arch": "amd64"},
			labels2:     map[string]string{"node.kubernetes.io/instance-type": "", "kubernetes.io/arch": "amd64"},
			isSimilar:   true,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			node1.ObjectMeta.Labels = tc.labels1
			node2.ObjectMeta.Labels = tc.labels2
			checkNodesSimilar(t, node1, node2, comparator, tc.isSimilar)
		})
	}
}
