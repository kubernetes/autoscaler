/*
Copyright 2019 The Kubernetes Authors.

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
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

// CreateGceNodeInfoComparator returns a comparator that checks if two nodes should be considered
// part of the same NodeGroupSet. This is true if they match usual conditions checked by IsCloudProviderNodeInfoSimilar,
// even if they have different GCE-specific labels.
func CreateGceNodeInfoComparator(extraIgnoredLabels []string) NodeInfoComparator {
	gceIgnoredLabels := map[string]bool{
		"topology.gke.io/zone": true,
	}

	for k, v := range BasicIgnoredLabels {
		gceIgnoredLabels[k] = v
	}

	for _, k := range extraIgnoredLabels {
		gceIgnoredLabels[k] = true
	}

	return func(n1, n2 *schedulerframework.NodeInfo) bool {
		return IsCloudProviderNodeInfoSimilar(n1, n2, gceIgnoredLabels)
	}
}
