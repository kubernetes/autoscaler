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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// DrgRouteTable All routing inside the DRG is driven by the contents of DRG route tables.
// DRG route tables contain rules which route packets to a particular network destination,
// represented as a DRG attachment.
// The routing decision for a packet entering a DRG is determined by the rules in the DRG route table
// assigned to the attachment-of-entry.
// Each DRG attachment can inject routes in any DRG route table, provided there is a statement corresponding to the attachment in the route table's `importDrgRouteDistribution`.
// You can also insert static routes into the DRG route tables.
// The DRG route table is always in the same compartment as the DRG. There must always be a default
// DRG route table for each attachment type.
type DrgRouteTable struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the
	// DRG route table.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment the DRG is in. The DRG route table
	// is always in the same compartment as the DRG.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG the DRG that contains this route table.
	DrgId *string `mandatory:"true" json:"drgId"`

	// The date and time the DRG route table was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The DRG route table's current state.
	LifecycleState DrgRouteTableLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// If you want traffic to be routed using ECMP across your virtual circuits or IPSec tunnels to
	// your on-premises network, enable ECMP on the DRG route table to which these attachments
	// import routes.
	IsEcmpEnabled *bool `mandatory:"true" json:"isEcmpEnabled"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the import route distribution used to specify how incoming route advertisements from
	// referenced attachments are inserted into the DRG route table.
	ImportDrgRouteDistributionId *string `mandatory:"false" json:"importDrgRouteDistributionId"`
}

func (m DrgRouteTable) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DrgRouteTable) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDrgRouteTableLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetDrgRouteTableLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DrgRouteTableLifecycleStateEnum Enum with underlying type: string
type DrgRouteTableLifecycleStateEnum string

// Set of constants representing the allowable values for DrgRouteTableLifecycleStateEnum
const (
	DrgRouteTableLifecycleStateProvisioning DrgRouteTableLifecycleStateEnum = "PROVISIONING"
	DrgRouteTableLifecycleStateAvailable    DrgRouteTableLifecycleStateEnum = "AVAILABLE"
	DrgRouteTableLifecycleStateTerminating  DrgRouteTableLifecycleStateEnum = "TERMINATING"
	DrgRouteTableLifecycleStateTerminated   DrgRouteTableLifecycleStateEnum = "TERMINATED"
)

var mappingDrgRouteTableLifecycleStateEnum = map[string]DrgRouteTableLifecycleStateEnum{
	"PROVISIONING": DrgRouteTableLifecycleStateProvisioning,
	"AVAILABLE":    DrgRouteTableLifecycleStateAvailable,
	"TERMINATING":  DrgRouteTableLifecycleStateTerminating,
	"TERMINATED":   DrgRouteTableLifecycleStateTerminated,
}

var mappingDrgRouteTableLifecycleStateEnumLowerCase = map[string]DrgRouteTableLifecycleStateEnum{
	"provisioning": DrgRouteTableLifecycleStateProvisioning,
	"available":    DrgRouteTableLifecycleStateAvailable,
	"terminating":  DrgRouteTableLifecycleStateTerminating,
	"terminated":   DrgRouteTableLifecycleStateTerminated,
}

// GetDrgRouteTableLifecycleStateEnumValues Enumerates the set of values for DrgRouteTableLifecycleStateEnum
func GetDrgRouteTableLifecycleStateEnumValues() []DrgRouteTableLifecycleStateEnum {
	values := make([]DrgRouteTableLifecycleStateEnum, 0)
	for _, v := range mappingDrgRouteTableLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgRouteTableLifecycleStateEnumStringValues Enumerates the set of values in String for DrgRouteTableLifecycleStateEnum
func GetDrgRouteTableLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingDrgRouteTableLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgRouteTableLifecycleStateEnum(val string) (DrgRouteTableLifecycleStateEnum, bool) {
	enum, ok := mappingDrgRouteTableLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
