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

// DrgRouteDistribution A route distribution establishes how routes get imported into DRG route tables and exported through the DRG attachments.
// A route distribution is a list of statements. Each statement consists of a set of matches, all of which must be `True` in order for
// the statement's action to take place. Each statement determines which routes are propagated.
// You can assign a route distribution as a route table's import distribution. The statements in an import
// route distribution specify how how incoming route advertisements through a referenced attachment or all attachments of a certain type are inserted into the route table.
// You can assign a route distribution as a DRG attachment's export distribution unless the
// attachment has the type VCN. Exporting routes through a VCN attachment is unsupported. Export
// route distribution statements specify how routes in a DRG attachment's assigned table are
// advertised out through the attachment. When a DRG is created, a route distribution is created
// with a single ACCEPT statement with match criteria MATCH_ALL. By default, all DRG attachments
// (except for those of type VCN), are assigned this distribution.
//
// The two auto-generated DRG route tables (one as the default for VCN attachments, and the other for all other types of attachments)
// are each assigned an auto generated import route distribution. The default VCN table's import distribution has a single statement with match criteria MATCH_ALL to import routes from
// each DRG attachment type. The other table's import distribution has a statement to import routes from attachments with the VCN type.
// The route distribution is always in the same compartment as the DRG.
type DrgRouteDistribution struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG that contains this route distribution.
	DrgId *string `mandatory:"true" json:"drgId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the route distribution.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The route distribution's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The route distribution's current state.
	LifecycleState DrgRouteDistributionLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the route distribution was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Whether this distribution defines how routes get imported into route tables or exported through DRG attachments.
	DistributionType DrgRouteDistributionDistributionTypeEnum `mandatory:"true" json:"distributionType"`

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
}

func (m DrgRouteDistribution) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DrgRouteDistribution) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDrgRouteDistributionLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetDrgRouteDistributionLifecycleStateEnumStringValues(), ",")))
	}
	if _, ok := GetMappingDrgRouteDistributionDistributionTypeEnum(string(m.DistributionType)); !ok && m.DistributionType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for DistributionType: %s. Supported values are: %s.", m.DistributionType, strings.Join(GetDrgRouteDistributionDistributionTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DrgRouteDistributionLifecycleStateEnum Enum with underlying type: string
type DrgRouteDistributionLifecycleStateEnum string

// Set of constants representing the allowable values for DrgRouteDistributionLifecycleStateEnum
const (
	DrgRouteDistributionLifecycleStateProvisioning DrgRouteDistributionLifecycleStateEnum = "PROVISIONING"
	DrgRouteDistributionLifecycleStateAvailable    DrgRouteDistributionLifecycleStateEnum = "AVAILABLE"
	DrgRouteDistributionLifecycleStateTerminating  DrgRouteDistributionLifecycleStateEnum = "TERMINATING"
	DrgRouteDistributionLifecycleStateTerminated   DrgRouteDistributionLifecycleStateEnum = "TERMINATED"
)

var mappingDrgRouteDistributionLifecycleStateEnum = map[string]DrgRouteDistributionLifecycleStateEnum{
	"PROVISIONING": DrgRouteDistributionLifecycleStateProvisioning,
	"AVAILABLE":    DrgRouteDistributionLifecycleStateAvailable,
	"TERMINATING":  DrgRouteDistributionLifecycleStateTerminating,
	"TERMINATED":   DrgRouteDistributionLifecycleStateTerminated,
}

var mappingDrgRouteDistributionLifecycleStateEnumLowerCase = map[string]DrgRouteDistributionLifecycleStateEnum{
	"provisioning": DrgRouteDistributionLifecycleStateProvisioning,
	"available":    DrgRouteDistributionLifecycleStateAvailable,
	"terminating":  DrgRouteDistributionLifecycleStateTerminating,
	"terminated":   DrgRouteDistributionLifecycleStateTerminated,
}

// GetDrgRouteDistributionLifecycleStateEnumValues Enumerates the set of values for DrgRouteDistributionLifecycleStateEnum
func GetDrgRouteDistributionLifecycleStateEnumValues() []DrgRouteDistributionLifecycleStateEnum {
	values := make([]DrgRouteDistributionLifecycleStateEnum, 0)
	for _, v := range mappingDrgRouteDistributionLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgRouteDistributionLifecycleStateEnumStringValues Enumerates the set of values in String for DrgRouteDistributionLifecycleStateEnum
func GetDrgRouteDistributionLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}

// GetMappingDrgRouteDistributionLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgRouteDistributionLifecycleStateEnum(val string) (DrgRouteDistributionLifecycleStateEnum, bool) {
	enum, ok := mappingDrgRouteDistributionLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// DrgRouteDistributionDistributionTypeEnum Enum with underlying type: string
type DrgRouteDistributionDistributionTypeEnum string

// Set of constants representing the allowable values for DrgRouteDistributionDistributionTypeEnum
const (
	DrgRouteDistributionDistributionTypeImport DrgRouteDistributionDistributionTypeEnum = "IMPORT"
	DrgRouteDistributionDistributionTypeExport DrgRouteDistributionDistributionTypeEnum = "EXPORT"
)

var mappingDrgRouteDistributionDistributionTypeEnum = map[string]DrgRouteDistributionDistributionTypeEnum{
	"IMPORT": DrgRouteDistributionDistributionTypeImport,
	"EXPORT": DrgRouteDistributionDistributionTypeExport,
}

var mappingDrgRouteDistributionDistributionTypeEnumLowerCase = map[string]DrgRouteDistributionDistributionTypeEnum{
	"import": DrgRouteDistributionDistributionTypeImport,
	"export": DrgRouteDistributionDistributionTypeExport,
}

// GetDrgRouteDistributionDistributionTypeEnumValues Enumerates the set of values for DrgRouteDistributionDistributionTypeEnum
func GetDrgRouteDistributionDistributionTypeEnumValues() []DrgRouteDistributionDistributionTypeEnum {
	values := make([]DrgRouteDistributionDistributionTypeEnum, 0)
	for _, v := range mappingDrgRouteDistributionDistributionTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgRouteDistributionDistributionTypeEnumStringValues Enumerates the set of values in String for DrgRouteDistributionDistributionTypeEnum
func GetDrgRouteDistributionDistributionTypeEnumStringValues() []string {
	return []string{
		"IMPORT",
		"EXPORT",
	}
}

// GetMappingDrgRouteDistributionDistributionTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgRouteDistributionDistributionTypeEnum(val string) (DrgRouteDistributionDistributionTypeEnum, bool) {
	enum, ok := mappingDrgRouteDistributionDistributionTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
