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
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// IsAwsNodeInfoSimilar adds AWS specific node labels to the list of ignored labels.
func IsAwsNodeInfoSimilar(n1, n2 *schedulernodeinfo.NodeInfo) bool {
	awsIgnoredLabels := map[string]bool{
		"alpha.eksctl.io/instance-id":    true, // this is a label used by eksctl to identify instances.
		"alpha.eksctl.io/nodegroup-name": true, // this is a label used by eksctl to identify "node group" names.
		"eks.amazonaws.com/nodegroup":    true, // this is a label used by eks to identify "node group".
		"k8s.amazonaws.com/eniConfig":    true, // this is a label used by the AWS CNI for custom networking.
		"lifecycle":                      true, // this is a label used by the AWS for spot.
	}

	for k, v := range BasicIgnoredLabels {
		awsIgnoredLabels[k] = v
	}
	return IsCloudProviderNodeInfoSimilar(n1, n2, awsIgnoredLabels)
}
