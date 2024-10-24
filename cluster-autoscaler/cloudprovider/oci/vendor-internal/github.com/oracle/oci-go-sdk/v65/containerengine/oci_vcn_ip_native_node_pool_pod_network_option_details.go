// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// OciVcnIpNativeNodePoolPodNetworkOptionDetails Network options specific to using the OCI VCN Native CNI
type OciVcnIpNativeNodePoolPodNetworkOptionDetails struct {

	// The OCIDs of the subnets in which to place pods for this node pool. This can be one of the node pool subnet IDs
	PodSubnetIds []string `mandatory:"true" json:"podSubnetIds"`

	// The max number of pods per node in the node pool. This value will be limited by the number of VNICs attachable to the node pool shape
	MaxPodsPerNode *int `mandatory:"false" json:"maxPodsPerNode"`

	// The OCIDs of the Network Security Group(s) to associate pods for this node pool with. For more information about NSGs, see NetworkSecurityGroup.
	PodNsgIds []string `mandatory:"false" json:"podNsgIds"`
}

func (m OciVcnIpNativeNodePoolPodNetworkOptionDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m OciVcnIpNativeNodePoolPodNetworkOptionDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m OciVcnIpNativeNodePoolPodNetworkOptionDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeOciVcnIpNativeNodePoolPodNetworkOptionDetails OciVcnIpNativeNodePoolPodNetworkOptionDetails
	s := struct {
		DiscriminatorParam string `json:"cniType"`
		MarshalTypeOciVcnIpNativeNodePoolPodNetworkOptionDetails
	}{
		"OCI_VCN_IP_NATIVE",
		(MarshalTypeOciVcnIpNativeNodePoolPodNetworkOptionDetails)(m),
	}

	return json.Marshal(&s)
}
