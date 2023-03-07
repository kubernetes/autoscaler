// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// InternalDrgAttachment A link between a DRG and VCN. For more information, see
// Overview of the Networking Service (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm).
type InternalDrgAttachment struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment containing the DRG attachment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the DRG.
	DrgId *string `mandatory:"true" json:"drgId"`

	// The DRG attachment's Oracle ID (OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The DRG attachment's current state.
	LifecycleState InternalDrgAttachmentLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VCN.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// NextHop target's MPLS label.
	MplsLabel *string `mandatory:"true" json:"mplsLabel"`

	// The string in the form ASN:mplsLabel.
	RouteTarget *string `mandatory:"true" json:"routeTarget"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the route table the DRG attachment is using.
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// The date and time the DRG attachment was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m InternalDrgAttachment) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InternalDrgAttachment) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInternalDrgAttachmentLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInternalDrgAttachmentLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InternalDrgAttachmentLifecycleStateEnum Enum with underlying type: string
type InternalDrgAttachmentLifecycleStateEnum string

// Set of constants representing the allowable values for InternalDrgAttachmentLifecycleStateEnum
const (
	InternalDrgAttachmentLifecycleStateAttaching InternalDrgAttachmentLifecycleStateEnum = "ATTACHING"
	InternalDrgAttachmentLifecycleStateAttached  InternalDrgAttachmentLifecycleStateEnum = "ATTACHED"
	InternalDrgAttachmentLifecycleStateDetaching InternalDrgAttachmentLifecycleStateEnum = "DETACHING"
	InternalDrgAttachmentLifecycleStateDetached  InternalDrgAttachmentLifecycleStateEnum = "DETACHED"
)

var mappingInternalDrgAttachmentLifecycleStateEnum = map[string]InternalDrgAttachmentLifecycleStateEnum{
	"ATTACHING": InternalDrgAttachmentLifecycleStateAttaching,
	"ATTACHED":  InternalDrgAttachmentLifecycleStateAttached,
	"DETACHING": InternalDrgAttachmentLifecycleStateDetaching,
	"DETACHED":  InternalDrgAttachmentLifecycleStateDetached,
}

// GetInternalDrgAttachmentLifecycleStateEnumValues Enumerates the set of values for InternalDrgAttachmentLifecycleStateEnum
func GetInternalDrgAttachmentLifecycleStateEnumValues() []InternalDrgAttachmentLifecycleStateEnum {
	values := make([]InternalDrgAttachmentLifecycleStateEnum, 0)
	for _, v := range mappingInternalDrgAttachmentLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInternalDrgAttachmentLifecycleStateEnumStringValues Enumerates the set of values in String for InternalDrgAttachmentLifecycleStateEnum
func GetInternalDrgAttachmentLifecycleStateEnumStringValues() []string {
	return []string{
		"ATTACHING",
		"ATTACHED",
		"DETACHING",
		"DETACHED",
	}
}
