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

// DrgAttachmentNetworkCreateDetails The representation of DrgAttachmentNetworkCreateDetails
type DrgAttachmentNetworkCreateDetails interface {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the network attached to the DRG.
	GetId() *string
}

type drgattachmentnetworkcreatedetails struct {
	JsonData []byte
	Id       *string `mandatory:"false" json:"id"`
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
	case "VCN":
		mm := VcnDrgAttachmentNetworkCreateDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for DrgAttachmentNetworkCreateDetails: %s.", m.Type)
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
	DrgAttachmentNetworkCreateDetailsTypeVcn DrgAttachmentNetworkCreateDetailsTypeEnum = "VCN"
)

var mappingDrgAttachmentNetworkCreateDetailsTypeEnum = map[string]DrgAttachmentNetworkCreateDetailsTypeEnum{
	"VCN": DrgAttachmentNetworkCreateDetailsTypeVcn,
}

var mappingDrgAttachmentNetworkCreateDetailsTypeEnumLowerCase = map[string]DrgAttachmentNetworkCreateDetailsTypeEnum{
	"vcn": DrgAttachmentNetworkCreateDetailsTypeVcn,
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
	}
}

// GetMappingDrgAttachmentNetworkCreateDetailsTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingDrgAttachmentNetworkCreateDetailsTypeEnum(val string) (DrgAttachmentNetworkCreateDetailsTypeEnum, bool) {
	enum, ok := mappingDrgAttachmentNetworkCreateDetailsTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
