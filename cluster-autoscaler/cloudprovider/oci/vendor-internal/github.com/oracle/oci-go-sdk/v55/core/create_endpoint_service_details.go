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

// CreateEndpointServiceDetails Details for creating an endpoint service.
type CreateEndpointServiceDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to contain the
	// endpoint service.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// List of service IP addresses (in the service VCN) that handle requests to the endpoint service.
	ServiceIps []EndpointServiceIpDetails `mandatory:"true" json:"serviceIps"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the service VCN that the endpoint
	// service belongs to.
	VcnId *string `mandatory:"false" json:"vcnId"`

	// A description of the endpoint service. For Oracle services that use the "trusted" mode of the
	// private endpoint service, customers never see this description. Avoid entering confidential information.
	Description *string `mandatory:"false" json:"description"`

	// A friendly name for the endpoint service. Must be unique within the VCN. Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Some services want to restrict access to the resources represented by an endpoint service so
	// that only a single private endpoint in the customer VCN has access.
	// For example, the endpoint service might represent a particular service resource (such as a
	// particular database). The service might want to allow access to that particular resource
	// from only a single private endpoint.
	// Defaults to `false`.
	// **Note:** If the value is `true`, the `endpointFqdn` attribute cannot have a value. For more
	// information, see the discussion of DNS and FQDNs in
	// PrivateEndpoint.
	// Example: `true`
	AreMultiplePrivateEndpointsPerVcnAllowed *bool `mandatory:"false" json:"areMultiplePrivateEndpointsPerVcnAllowed"`

	// Reserved for future use.
	IsVcnMetadataEnabled *bool `mandatory:"false" json:"isVcnMetadataEnabled"`

	// The ports on the endpoint service IPs that are open for private endpoint traffic for this
	// endpoint service. If you provide no ports, all open ports on the service IPs are accessible.
	Ports []int `mandatory:"false" json:"ports"`

	// The default three-label FQDN to use for all private endpoints associated with this endpoint
	// service. For important information about how this attribute is used, see the discussion of DNS
	// and FQDNs in PrivateEndpoint.
	// If `areMultiplePrivateEndpointsPerVcnAllowed` is `true`, `endpointFqdn` cannot have a value.
	// Example: `xyz.oraclecloud.com`
	EndpointFqdn *string `mandatory:"false" json:"endpointFqdn"`

	// The Endpoint Service belong to a substrate or not
	IsSubstrate *bool `mandatory:"false" json:"isSubstrate"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m CreateEndpointServiceDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateEndpointServiceDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
