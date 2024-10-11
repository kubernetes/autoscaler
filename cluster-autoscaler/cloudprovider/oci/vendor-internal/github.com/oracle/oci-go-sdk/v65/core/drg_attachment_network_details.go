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

// DrgAttachmentNetworkDetails The representation of DrgAttachmentNetworkDetails
type DrgAttachmentNetworkDetails interface {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the network attached to the DRG.
	GetId() *string
}

type drgattachmentnetworkdetails struct {
	JsonData []byte
	Id       *string `mandatory:"false" json:"id"`
	Type     string  `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *drgattachmentnetworkdetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerdrgattachmentnetworkdetails drgattachmentnetworkdetails
	s := struct {
		Model Unmarshalerdrgattachmentnetworkdetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Id = s.Model.Id
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *drgattachmentnetworkdetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "VCN":
		mm := VcnDrgAttachmentNetworkDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "LOOPBACK":
		mm := LoopBackDrgAttachmentNetworkDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "IPSEC_TUNNEL":
		mm := IpsecTunnelDrgAttachmentNetworkDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "VIRTUAL_CIRCUIT":
		mm := VirtualCircuitDrgAttachmentNetworkDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "REMOTE_PEERING_CONNECTION":
		mm := RemotePeeringConnectionDrgAttachmentNetworkDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for DrgAttachmentNetworkDetails: %s.", m.Type)
		return *m, nil
	}
}

// GetId returns Id
func (m drgattachmentnetworkdetails) GetId() *string {
	return m.Id
}

func (m drgattachmentnetworkdetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m drgattachmentnetworkdetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DrgAttachmentNetworkDetailsTypeEnum Enum with underlying type: string
type DrgAttachmentNetworkDetailsTypeEnum string

// Set of constants representing the allowable values for DrgAttachmentNetworkDetailsTypeEnum
const (
	DrgAttachmentNetworkDetailsTypeVcn                     DrgAttachmentNetworkDetailsTypeEnum = "VCN"
	DrgAttachmentNetworkDetailsTypeIpsecTunnel             DrgAttachmentNetworkDetailsTypeEnum = "IPSEC_TUNNEL"
	DrgAttachmentNetworkDetailsTypeVirtualCircuit          DrgAttachmentNetworkDetailsTypeEnum = "VIRTUAL_CIRCUIT"
	DrgAttachmentNetworkDetailsTypeRemotePeeringConnection DrgAttachmentNetworkDetailsTypeEnum = "REMOTE_PEERING_CONNECTION"
)

var mappingDrgAttachmentNetworkDetailsTypeEnum = map[string]DrgAttachmentNetworkDetailsTypeEnum{
	"VCN":                       DrgAttachmentNetworkDetailsTypeVcn,
	"IPSEC_TUNNEL":              DrgAttachmentNetworkDetailsTypeIpsecTunnel,
	"VIRTUAL_CIRCUIT":           DrgAttachmentNetworkDetailsTypeVirtualCircuit,
	"REMOTE_PEERING_CONNECTION": DrgAttachmentNetworkDetailsTypeRemotePeeringConnection,
}

var mappingDrgAttachmentNetworkDetailsTypeEnumLowerCase = map[string]DrgAttachmentNetworkDetailsTypeEnum{
	"vcn":                       DrgAttachmentNetworkDetailsTypeVcn,
	"ipsec_tunnel":              DrgAttachmentNetworkDetailsTypeIpsecTunnel,
	"virtual_circuit":           DrgAttachmentNetworkDetailsTypeVirtualCircuit,
	"remote_peering_connection": DrgAttachmentNetworkDetailsTypeRemotePeeringConnection,
}

// GetDrgAttachmentNetworkDetailsTypeEnumValues Enumerates the set of values for DrgAttachmentNetworkDetailsTypeEnum
func GetDrgAttachmentNetworkDetailsTypeEnumValues() []DrgAttachmentNetworkDetailsTypeEnum {
	values := make([]DrgAttachmentNetworkDetailsTypeEnum, 0)
	for _, v := range mappingDrgAttachmentNetworkDetailsTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgAttachmentNetworkDetailsTypeEnumStringValues Enumerates the set of values in String for DrgAttachmentNetworkDetailsTypeEnum
func GetDrgAttachmentNetworkDetailsTypeEnumStringValues() []string {
	return []string{
		"VCN",
		"IPSEC_TUNNEL",
		"VIRTUAL_CIRCUIT",
		"REMOTE_PEERING_CONNECTION",
	}
}

// GetMappingDrgAttachmentNetworkDetailsTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgAttachmentNetworkDetailsTypeEnum(val string) (DrgAttachmentNetworkDetailsTypeEnum, bool) {
	enum, ok := mappingDrgAttachmentNetworkDetailsTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
