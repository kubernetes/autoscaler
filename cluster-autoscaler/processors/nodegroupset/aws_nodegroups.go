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
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// this is a label used by eksctl to identify instances.
	AwsIgnoredLabelEksctlInstanceId = "alpha.eksctl.io/instance-id"

	// this is a label used by eksctl to identify "node group" names.
	AwsIgnoredLabelEksctlNodegroupName = "alpha.eksctl.io/nodegroup-name"

	// this is a label used by eks to identify "node group".
	AwsIgnoredLabelEksNodegroup = "eks.amazonaws.com/nodegroup"

	// this is a label used by the AWS CNI for custom networking.
	AwsIgnoredLabelK8sEniconfig = "k8s.amazonaws.com/eniConfig"

	// this is a label used by the AWS for spot.
	AwsIgnoredLabelLifecycle = "lifecycle"

	// this is a label used by the AWS EBS CSI driver as a target for Persistent Volume Node Affinity
	AwsIgnoredLabelEbsCsiZone = "topology.ebs.csi.aws.com/zone"
)

// CreateAwsNodeInfoComparator returns a comparator that checks if two nodes should be considered
// part of the same NodeGroupSet. This is true if they match usual conditions checked by IsCloudProviderNodeInfoSimilar,
// even if they have different AWS-specific labels.
func CreateAwsNodeInfoComparator(extraIgnoredLabels []string) NodeInfoComparator {
	awsIgnoredLabels := map[string]bool{
		AwsIgnoredLabelEksctlInstanceId:    true,
		AwsIgnoredLabelEksctlNodegroupName: true,
		AwsIgnoredLabelEksNodegroup:        true,
		AwsIgnoredLabelK8sEniconfig:        true,
		AwsIgnoredLabelLifecycle:           true,
		AwsIgnoredLabelEbsCsiZone:          true,
		GceIgnoredLabelGkeZone:             true,
	}

	for k, v := range BasicIgnoredLabels {
		awsIgnoredLabels[k] = v
	}

	for _, k := range extraIgnoredLabels {
		awsIgnoredLabels[k] = true
	}

	return func(n1, n2 *schedulerframework.NodeInfo) bool {
		return IsCloudProviderNodeInfoSimilar(n1, n2, awsIgnoredLabels)
	}
}
