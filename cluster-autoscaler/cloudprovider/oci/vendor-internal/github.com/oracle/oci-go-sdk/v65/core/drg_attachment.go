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

// DrgAttachment A DRG attachment serves as a link between a DRG and a network resource. A DRG can be attached to a VCN,
// IPSec tunnel, remote peering connection, or virtual circuit.
// For more information, see Overview of the Networking Service (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm).
type DrgAttachment struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the DRG attachment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG.
	DrgId *string `mandatory:"true" json:"drgId"`

	// The DRG attachment's Oracle ID (OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The DRG attachment's current state.
	LifecycleState DrgAttachmentLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The date and time the DRG attachment was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the DRG route table that is assigned to this attachment.
	// The DRG route table manages traffic inside the DRG.
	DrgRouteTableId *string `mandatory:"false" json:"drgRouteTableId"`

	NetworkDetails DrgAttachmentNetworkDetails `mandatory:"false" json:"networkDetails"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the route table the DRG attachment is using.
	// For information about why you would associate a route table with a DRG attachment, see:
	//   * Transit Routing: Access to Multiple VCNs in Same Region (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/transitrouting.htm)
	//   * Transit Routing: Private Access to Oracle Services (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/transitroutingoracleservices.htm)
	// This field is deprecated. Instead, use the `networkDetails` field to view the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the attached resource.
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN.
	// This field is deprecated. Instead, use the `networkDetails` field to view the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the attached resource.
	VcnId *string `mandatory:"false" json:"vcnId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the export route distribution used to specify how routes in the assigned DRG route table
	// are advertised to the attachment.
	// If this value is null, no routes are advertised through this attachment.
	ExportDrgRouteDistributionId *string `mandatory:"false" json:"exportDrgRouteDistributionId"`

	// Indicates whether the DRG attachment and attached network live in a different tenancy than the DRG.
	// Example: `false`
	IsCrossTenancy *bool `mandatory:"false" json:"isCrossTenancy"`
}

func (m DrgAttachment) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m DrgAttachment) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingDrgAttachmentLifecycleStateEnum(string(m.LifecycleState)); !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetDrgAttachmentLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *DrgAttachment) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DisplayName                  *string                           `json:"displayName"`
		TimeCreated                  *common.SDKTime                   `json:"timeCreated"`
		DrgRouteTableId              *string                           `json:"drgRouteTableId"`
		NetworkDetails               drgattachmentnetworkdetails       `json:"networkDetails"`
		DefinedTags                  map[string]map[string]interface{} `json:"definedTags"`
		FreeformTags                 map[string]string                 `json:"freeformTags"`
		RouteTableId                 *string                           `json:"routeTableId"`
		VcnId                        *string                           `json:"vcnId"`
		ExportDrgRouteDistributionId *string                           `json:"exportDrgRouteDistributionId"`
		IsCrossTenancy               *bool                             `json:"isCrossTenancy"`
		CompartmentId                *string                           `json:"compartmentId"`
		DrgId                        *string                           `json:"drgId"`
		Id                           *string                           `json:"id"`
		LifecycleState               DrgAttachmentLifecycleStateEnum   `json:"lifecycleState"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DisplayName = model.DisplayName

	m.TimeCreated = model.TimeCreated

	m.DrgRouteTableId = model.DrgRouteTableId

	nn, e = model.NetworkDetails.UnmarshalPolymorphicJSON(model.NetworkDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.NetworkDetails = nn.(DrgAttachmentNetworkDetails)
	} else {
		m.NetworkDetails = nil
	}

	m.DefinedTags = model.DefinedTags

	m.FreeformTags = model.FreeformTags

	m.RouteTableId = model.RouteTableId

	m.VcnId = model.VcnId

	m.ExportDrgRouteDistributionId = model.ExportDrgRouteDistributionId

	m.IsCrossTenancy = model.IsCrossTenancy

	m.CompartmentId = model.CompartmentId

	m.DrgId = model.DrgId

	m.Id = model.Id

	m.LifecycleState = model.LifecycleState

	return
}

// DrgAttachmentLifecycleStateEnum Enum with underlying type: string
type DrgAttachmentLifecycleStateEnum string

// Set of constants representing the allowable values for DrgAttachmentLifecycleStateEnum
const (
	DrgAttachmentLifecycleStateAttaching DrgAttachmentLifecycleStateEnum = "ATTACHING"
	DrgAttachmentLifecycleStateAttached  DrgAttachmentLifecycleStateEnum = "ATTACHED"
	DrgAttachmentLifecycleStateDetaching DrgAttachmentLifecycleStateEnum = "DETACHING"
	DrgAttachmentLifecycleStateDetached  DrgAttachmentLifecycleStateEnum = "DETACHED"
)

var mappingDrgAttachmentLifecycleStateEnum = map[string]DrgAttachmentLifecycleStateEnum{
	"ATTACHING": DrgAttachmentLifecycleStateAttaching,
	"ATTACHED":  DrgAttachmentLifecycleStateAttached,
	"DETACHING": DrgAttachmentLifecycleStateDetaching,
	"DETACHED":  DrgAttachmentLifecycleStateDetached,
}

var mappingDrgAttachmentLifecycleStateEnumLowerCase = map[string]DrgAttachmentLifecycleStateEnum{
	"attaching": DrgAttachmentLifecycleStateAttaching,
	"attached":  DrgAttachmentLifecycleStateAttached,
	"detaching": DrgAttachmentLifecycleStateDetaching,
	"detached":  DrgAttachmentLifecycleStateDetached,
}

// GetDrgAttachmentLifecycleStateEnumValues Enumerates the set of values for DrgAttachmentLifecycleStateEnum
func GetDrgAttachmentLifecycleStateEnumValues() []DrgAttachmentLifecycleStateEnum {
	values := make([]DrgAttachmentLifecycleStateEnum, 0)
	for _, v := range mappingDrgAttachmentLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgAttachmentLifecycleStateEnumStringValues Enumerates the set of values in String for DrgAttachmentLifecycleStateEnum
func GetDrgAttachmentLifecycleStateEnumStringValues() []string {
	return []string{
		"ATTACHING",
		"ATTACHED",
		"DETACHING",
		"DETACHED",
	}
}

// GetMappingDrgAttachmentLifecycleStateEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgAttachmentLifecycleStateEnum(val string) (DrgAttachmentLifecycleStateEnum, bool) {
	enum, ok := mappingDrgAttachmentLifecycleStateEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
