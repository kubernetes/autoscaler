// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"encoding/json"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// DrgAttachmentTypeDrgRouteDistributionMatchCriteria The attachment type from which the DRG will import routes. Routes will be imported from
// all attachments of this type.
type DrgAttachmentTypeDrgRouteDistributionMatchCriteria struct {

	// The type of the network resource to be included in this match. A match for a network type implies that all
	// DRG attachments of that type insert routes into the table.
	AttachmentType DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum `mandatory:"true" json:"attachmentType"`
}

func (m DrgAttachmentTypeDrgRouteDistributionMatchCriteria) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m DrgAttachmentTypeDrgRouteDistributionMatchCriteria) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeDrgAttachmentTypeDrgRouteDistributionMatchCriteria DrgAttachmentTypeDrgRouteDistributionMatchCriteria
	s := struct {
		DiscriminatorParam string `json:"matchType"`
		MarshalTypeDrgAttachmentTypeDrgRouteDistributionMatchCriteria
	}{
		"DRG_ATTACHMENT_TYPE",
		(MarshalTypeDrgAttachmentTypeDrgRouteDistributionMatchCriteria)(m),
	}

	return json.Marshal(&s)
}

// DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum Enum with underlying type: string
type DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum string

// Set of constants representing the allowable values for DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum
const (
	DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeVcn                     DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum = "VCN"
	DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeVirtualCircuit          DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum = "VIRTUAL_CIRCUIT"
	DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeRemotePeeringConnection DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum = "REMOTE_PEERING_CONNECTION"
	DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeIpsecTunnel             DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum = "IPSEC_TUNNEL"
)

var mappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentType = map[string]DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum{
	"VCN":                       DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeVcn,
	"VIRTUAL_CIRCUIT":           DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeVirtualCircuit,
	"REMOTE_PEERING_CONNECTION": DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeRemotePeeringConnection,
	"IPSEC_TUNNEL":              DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeIpsecTunnel,
}

// GetDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumValues Enumerates the set of values for DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum
func GetDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumValues() []DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum {
	values := make([]DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum, 0)
	for _, v := range mappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentType {
		values = append(values, v)
	}
	return values
}
