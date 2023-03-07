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

// EndpointService Required for Oracle services that offer customers private endpoints for private access to the
// service.
// An endpoint service is an object that resides in the Oracle service's VCN and represents the IP addresses
// for accessing the service. An endpoint service can be associated with one or more private
// endpoints, which reside in customer VCNs (see PrivateEndpoint).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm).
type EndpointService struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the endpoint service.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment that contains the
	// endpoint service.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// List of service IP addresses (in the service VCN) that handle requests to the endpoint service.
	ServiceIps []EndpointServiceIpDetails `mandatory:"true" json:"serviceIps"`

	// The ports on the endpoint service IPs that are open for private endpoint traffic for this
	// endpoint service. If you provide no ports, all open ports on the service IPs are accessible.
	Ports []int `mandatory:"true" json:"ports"`

	// The three-label FQDN to use for all private endpoints associated with this endpoint
	// service. This attribute's value cannot be changed.
	// For important information about how this attribute is used, see the discussion
	// of DNS and FQDNs in PrivateEndpoint.
	// Example: `xyz.oraclecloud.com`
	EndpointFqdn *string `mandatory:"true" json:"endpointFqdn"`

	// The date and time the endpoint service was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The endpoint service's current lifecycle state.
	LifecycleState EndpointServiceLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the service VCN that the endpoint
	// service belongs to.
	VcnId *string `mandatory:"false" json:"vcnId"`

	// A description of the endpoint service. For Oracle services that use the "trusted" mode of the
	// private endpoint service, customers never see this description.
	Description *string `mandatory:"false" json:"description"`

	// A friendly name for the endpoint service. Must be unique within the VCN. Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Some Oracle services want to restrict access to the resources represented by an endpoint service so
	// that only a single private endpoint in the customer VCN has access.
	// For example, the endpoint service might represent a particular service resource (such as a
	// particular database). The service might want to allow access to that particular resource
	// from only a single private endpoint.
	// Defaults to `false`.
	// Example: `true`
	AreMultiplePrivateEndpointsPerVcnAllowed *bool `mandatory:"false" json:"areMultiplePrivateEndpointsPerVcnAllowed"`

	// Reserved for future use.
	IsVcnMetadataEnabled *bool `mandatory:"false" json:"isVcnMetadataEnabled"`

	// ES from substrate or not
	IsSubstrate *bool `mandatory:"false" json:"isSubstrate"`

	// RCE substrate anycast IP
	ReverseConnectionAnycastIp *string `mandatory:"false" json:"reverseConnectionAnycastIp"`

	// MPLS label that identifies the substrate endpoint service
	ReverseConnectionMplsLabel *int `mandatory:"false" json:"reverseConnectionMplsLabel"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m EndpointService) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m EndpointService) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingEndpointServiceLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetEndpointServiceLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// EndpointServiceLifecycleStateEnum Enum with underlying type: string
type EndpointServiceLifecycleStateEnum string

// Set of constants representing the allowable values for EndpointServiceLifecycleStateEnum
const (
	EndpointServiceLifecycleStateProvisioning EndpointServiceLifecycleStateEnum = "PROVISIONING"
	EndpointServiceLifecycleStateAvailable    EndpointServiceLifecycleStateEnum = "AVAILABLE"
	EndpointServiceLifecycleStateTerminating  EndpointServiceLifecycleStateEnum = "TERMINATING"
	EndpointServiceLifecycleStateTerminated   EndpointServiceLifecycleStateEnum = "TERMINATED"
	EndpointServiceLifecycleStateUpdating     EndpointServiceLifecycleStateEnum = "UPDATING"
	EndpointServiceLifecycleStateFailed       EndpointServiceLifecycleStateEnum = "FAILED"
)

var mappingEndpointServiceLifecycleStateEnum = map[string]EndpointServiceLifecycleStateEnum{
	"PROVISIONING": EndpointServiceLifecycleStateProvisioning,
	"AVAILABLE":    EndpointServiceLifecycleStateAvailable,
	"TERMINATING":  EndpointServiceLifecycleStateTerminating,
	"TERMINATED":   EndpointServiceLifecycleStateTerminated,
	"UPDATING":     EndpointServiceLifecycleStateUpdating,
	"FAILED":       EndpointServiceLifecycleStateFailed,
}

// GetEndpointServiceLifecycleStateEnumValues Enumerates the set of values for EndpointServiceLifecycleStateEnum
func GetEndpointServiceLifecycleStateEnumValues() []EndpointServiceLifecycleStateEnum {
	values := make([]EndpointServiceLifecycleStateEnum, 0)
	for _, v := range mappingEndpointServiceLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetEndpointServiceLifecycleStateEnumStringValues Enumerates the set of values in String for EndpointServiceLifecycleStateEnum
func GetEndpointServiceLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
		"UPDATING",
		"FAILED",
	}
}
