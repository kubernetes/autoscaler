// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.cloud.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
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

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DrgAttachmentTypeDrgRouteDistributionMatchCriteria) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum(string(m.AttachmentType)); !ok && m.AttachmentType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for AttachmentType: %s. Supported values are: %s.", m.AttachmentType, strings.Join(GetDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
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

var mappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum = map[string]DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum{
	"VCN":                       DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeVcn,
	"VIRTUAL_CIRCUIT":           DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeVirtualCircuit,
	"REMOTE_PEERING_CONNECTION": DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeRemotePeeringConnection,
	"IPSEC_TUNNEL":              DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeIpsecTunnel,
}

var mappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumLowerCase = map[string]DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum{
	"vcn":                       DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeVcn,
	"virtual_circuit":           DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeVirtualCircuit,
	"remote_peering_connection": DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeRemotePeeringConnection,
	"ipsec_tunnel":              DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeIpsecTunnel,
}

// GetDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumValues Enumerates the set of values for DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum
func GetDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumValues() []DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum {
	values := make([]DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum, 0)
	for _, v := range mappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumStringValues Enumerates the set of values in String for DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum
func GetDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumStringValues() []string {
	return []string{
		"VCN",
		"VIRTUAL_CIRCUIT",
		"REMOTE_PEERING_CONNECTION",
		"IPSEC_TUNNEL",
	}
}

// GetMappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum(val string) (DrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnum, bool) {
	enum, ok := mappingDrgAttachmentTypeDrgRouteDistributionMatchCriteriaAttachmentTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
