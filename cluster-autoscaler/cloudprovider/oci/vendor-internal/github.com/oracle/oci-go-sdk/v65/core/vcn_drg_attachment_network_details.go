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

// VcnDrgAttachmentNetworkDetails Specifies details within the VCN.
type VcnDrgAttachmentNetworkDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the network attached to the DRG.
	Id *string `mandatory:"false" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table the DRG attachment is using.
	// For information about why you would associate a route table with a DRG attachment, see:
	//   * Transit Routing: Access to Multiple VCNs in Same Region (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/transitrouting.htm)
	//   * Transit Routing: Private Access to Oracle Services (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/transitroutingoracleservices.htm)
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// Indicates whether the VCN CIDRs or the individual subnet CIDRs are imported from the attachment.
	// Routes from the VCN ingress route table are always imported.
	VcnRouteType VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum `mandatory:"false" json:"vcnRouteType,omitempty"`
}

// GetId returns Id
func (m VcnDrgAttachmentNetworkDetails) GetId() *string {
	return m.Id
}

func (m VcnDrgAttachmentNetworkDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VcnDrgAttachmentNetworkDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum(string(m.VcnRouteType)); !ok && m.VcnRouteType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for VcnRouteType: %s. Supported values are: %s.", m.VcnRouteType, strings.Join(GetVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m VcnDrgAttachmentNetworkDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeVcnDrgAttachmentNetworkDetails VcnDrgAttachmentNetworkDetails
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeVcnDrgAttachmentNetworkDetails
	}{
		"VCN",
		(MarshalTypeVcnDrgAttachmentNetworkDetails)(m),
	}

	return json.Marshal(&s)
}

// VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum Enum with underlying type: string
type VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum string

// Set of constants representing the allowable values for VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum
const (
	VcnDrgAttachmentNetworkDetailsVcnRouteTypeVcnCidrs    VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum = "VCN_CIDRS"
	VcnDrgAttachmentNetworkDetailsVcnRouteTypeSubnetCidrs VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum = "SUBNET_CIDRS"
)

var mappingVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum = map[string]VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum{
	"VCN_CIDRS":    VcnDrgAttachmentNetworkDetailsVcnRouteTypeVcnCidrs,
	"SUBNET_CIDRS": VcnDrgAttachmentNetworkDetailsVcnRouteTypeSubnetCidrs,
}

var mappingVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnumLowerCase = map[string]VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum{
	"vcn_cidrs":    VcnDrgAttachmentNetworkDetailsVcnRouteTypeVcnCidrs,
	"subnet_cidrs": VcnDrgAttachmentNetworkDetailsVcnRouteTypeSubnetCidrs,
}

// GetVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnumValues Enumerates the set of values for VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum
func GetVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnumValues() []VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum {
	values := make([]VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum, 0)
	for _, v := range mappingVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnumStringValues Enumerates the set of values in String for VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum
func GetVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnumStringValues() []string {
	return []string{
		"VCN_CIDRS",
		"SUBNET_CIDRS",
	}
}

// GetMappingVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum(val string) (VcnDrgAttachmentNetworkDetailsVcnRouteTypeEnum, bool) {
	enum, ok := mappingVcnDrgAttachmentNetworkDetailsVcnRouteTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
