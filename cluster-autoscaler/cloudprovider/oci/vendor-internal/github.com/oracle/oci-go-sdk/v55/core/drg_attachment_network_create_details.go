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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// DrgAttachmentNetworkCreateDetails The representation of DrgAttachmentNetworkCreateDetails
type DrgAttachmentNetworkCreateDetails interface {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the network attached to the DRG.
	GetId() *string
}

type drgattachmentnetworkcreatedetails struct {
	JsonData []byte
	Id       *string `mandatory:"true" json:"id"`
	Type     string  `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *drgattachmentnetworkcreatedetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerdrgattachmentnetworkcreatedetails drgattachmentnetworkcreatedetails
	s := struct {
		Model Unmarshalerdrgattachmentnetworkcreatedetails
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
func (m *drgattachmentnetworkcreatedetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "IPSEC_TUNNEL":
		mm := IpsecTunnelDrgAttachmentNetworkCreateDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "VIRTUAL_CIRCUIT":
		mm := VirtualCircuitDrgAttachmentNetworkCreateDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "REMOTE_PEERING_CONNECTION":
		mm := RemotePeeringConnectionDrgAttachmentNetworkCreateDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "INTERNET":
		mm := InternetDrgAttachmentNetworkCreateDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "VCN":
		mm := VcnDrgAttachmentNetworkCreateDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

// GetId returns Id
func (m drgattachmentnetworkcreatedetails) GetId() *string {
	return m.Id
}

func (m drgattachmentnetworkcreatedetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m drgattachmentnetworkcreatedetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// DrgAttachmentNetworkCreateDetailsTypeEnum Enum with underlying type: string
type DrgAttachmentNetworkCreateDetailsTypeEnum string

// Set of constants representing the allowable values for DrgAttachmentNetworkCreateDetailsTypeEnum
const (
	DrgAttachmentNetworkCreateDetailsTypeVcn                     DrgAttachmentNetworkCreateDetailsTypeEnum = "VCN"
	DrgAttachmentNetworkCreateDetailsTypeVirtualCircuit          DrgAttachmentNetworkCreateDetailsTypeEnum = "VIRTUAL_CIRCUIT"
	DrgAttachmentNetworkCreateDetailsTypeRemotePeeringConnection DrgAttachmentNetworkCreateDetailsTypeEnum = "REMOTE_PEERING_CONNECTION"
	DrgAttachmentNetworkCreateDetailsTypeIpsecTunnel             DrgAttachmentNetworkCreateDetailsTypeEnum = "IPSEC_TUNNEL"
	DrgAttachmentNetworkCreateDetailsTypeInternet                DrgAttachmentNetworkCreateDetailsTypeEnum = "INTERNET"
)

var mappingDrgAttachmentNetworkCreateDetailsTypeEnum = map[string]DrgAttachmentNetworkCreateDetailsTypeEnum{
	"VCN":                       DrgAttachmentNetworkCreateDetailsTypeVcn,
	"VIRTUAL_CIRCUIT":           DrgAttachmentNetworkCreateDetailsTypeVirtualCircuit,
	"REMOTE_PEERING_CONNECTION": DrgAttachmentNetworkCreateDetailsTypeRemotePeeringConnection,
	"IPSEC_TUNNEL":              DrgAttachmentNetworkCreateDetailsTypeIpsecTunnel,
	"INTERNET":                  DrgAttachmentNetworkCreateDetailsTypeInternet,
}

// GetDrgAttachmentNetworkCreateDetailsTypeEnumValues Enumerates the set of values for DrgAttachmentNetworkCreateDetailsTypeEnum
func GetDrgAttachmentNetworkCreateDetailsTypeEnumValues() []DrgAttachmentNetworkCreateDetailsTypeEnum {
	values := make([]DrgAttachmentNetworkCreateDetailsTypeEnum, 0)
	for _, v := range mappingDrgAttachmentNetworkCreateDetailsTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetDrgAttachmentNetworkCreateDetailsTypeEnumStringValues Enumerates the set of values in String for DrgAttachmentNetworkCreateDetailsTypeEnum
func GetDrgAttachmentNetworkCreateDetailsTypeEnumStringValues() []string {
	return []string{
		"VCN",
		"VIRTUAL_CIRCUIT",
		"REMOTE_PEERING_CONNECTION",
		"IPSEC_TUNNEL",
		"INTERNET",
	}
}
