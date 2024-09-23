/*
Copyright 2024 The Kubernetes Authors.

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

package aws

import (
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// CreateNodeInfoComparator returns a comparator that checks if two nodes should be considered
// part of the same NodeGroupSet. This is true if they match usual conditions checked by IsCloudProviderNodeInfoSimilar,
// even if they have different AWS-specific labels.
func CreateNodeInfoComparator(extraIgnoredLabels []string, ratioOpts config.NodeGroupDifferenceRatios) nodegroupset.NodeInfoComparator {
	awsIgnoredLabels := map[string]bool{
		"alpha.eksctl.io/instance-id":    true, // this is a label used by eksctl to identify instances.
		"alpha.eksctl.io/nodegroup-name": true, // this is a label used by eksctl to identify "node group" names.
		"eks.amazonaws.com/nodegroup":    true, // this is a label used by eks to identify "node group".
		"k8s.amazonaws.com/eniConfig":    true, // this is a label used by the AWS CNI for custom networking.
		"lifecycle":                      true, // this is a label used by the AWS for spot.
		"topology.ebs.csi.aws.com/zone":  true, // this is a label used by the AWS EBS CSI driver as a target for Persistent Volume Node Affinity
	}

	for k, v := range nodegroupset.BasicIgnoredLabels {
		awsIgnoredLabels[k] = v
	}

	for _, k := range extraIgnoredLabels {
		awsIgnoredLabels[k] = true
	}

	return func(n1, n2 *schedulerframework.NodeInfo) bool {
		return nodegroupset.IsCloudProviderNodeInfoSimilar(n1, n2, awsIgnoredLabels, ratioOpts)
	}
}
