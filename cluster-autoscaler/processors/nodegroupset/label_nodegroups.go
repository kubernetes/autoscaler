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
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
)

// CreateLabelNodeInfoComparator returns a comparator that checks for node group similarity using the given labels.
func CreateLabelNodeInfoComparator(labels []string) NodeInfoComparator {
	return func(n1, n2 *framework.NodeInfo) bool {
		return areLabelsSame(n1, n2, labels)
	}
}

func areLabelsSame(n1, n2 *framework.NodeInfo, labels []string) bool {
	for _, label := range labels {
		val1, exists := n1.Node().ObjectMeta.Labels[label]
		if !exists {
			klog.V(8).Infof("%s label not present on %s", label, n1.Node().Name)
			return false
		}
		val2, exists := n2.Node().ObjectMeta.Labels[label]
		if !exists {
			klog.V(8).Infof("%s label not present on %s", label, n1.Node().Name)
			return false
		}
		if val1 != val2 {
			klog.V(8).Infof("%s label did not match. %s: %s, %s: %s", label, n1.Node().Name, val1, n2.Node().Name, val2)
			return false
		}
	}
	return true
}
