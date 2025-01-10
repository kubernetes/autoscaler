// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// NodePoolPodNetworkOptionDetails The CNI type and relevant network details for the pods of a given node pool
type NodePoolPodNetworkOptionDetails interface {
}

type nodepoolpodnetworkoptiondetails struct {
	JsonData []byte
	CniType  string `json:"cniType"`
}

// UnmarshalJSON unmarshals json
func (m *nodepoolpodnetworkoptiondetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalernodepoolpodnetworkoptiondetails nodepoolpodnetworkoptiondetails
	s := struct {
		Model Unmarshalernodepoolpodnetworkoptiondetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.CniType = s.Model.CniType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *nodepoolpodnetworkoptiondetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.CniType {
	case "OCI_VCN_IP_NATIVE":
		mm := OciVcnIpNativeNodePoolPodNetworkOptionDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "FLANNEL_OVERLAY":
		mm := FlannelOverlayNodePoolPodNetworkOptionDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for NodePoolPodNetworkOptionDetails: %s.", m.CniType)
		return *m, nil
	}
}

func (m nodepoolpodnetworkoptiondetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m nodepoolpodnetworkoptiondetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// NodePoolPodNetworkOptionDetailsCniTypeEnum Enum with underlying type: string
type NodePoolPodNetworkOptionDetailsCniTypeEnum string

// Set of constants representing the allowable values for NodePoolPodNetworkOptionDetailsCniTypeEnum
const (
	NodePoolPodNetworkOptionDetailsCniTypeOciVcnIpNative NodePoolPodNetworkOptionDetailsCniTypeEnum = "OCI_VCN_IP_NATIVE"
	NodePoolPodNetworkOptionDetailsCniTypeFlannelOverlay NodePoolPodNetworkOptionDetailsCniTypeEnum = "FLANNEL_OVERLAY"
)

var mappingNodePoolPodNetworkOptionDetailsCniTypeEnum = map[string]NodePoolPodNetworkOptionDetailsCniTypeEnum{
	"OCI_VCN_IP_NATIVE": NodePoolPodNetworkOptionDetailsCniTypeOciVcnIpNative,
	"FLANNEL_OVERLAY":   NodePoolPodNetworkOptionDetailsCniTypeFlannelOverlay,
}

var mappingNodePoolPodNetworkOptionDetailsCniTypeEnumLowerCase = map[string]NodePoolPodNetworkOptionDetailsCniTypeEnum{
	"oci_vcn_ip_native": NodePoolPodNetworkOptionDetailsCniTypeOciVcnIpNative,
	"flannel_overlay":   NodePoolPodNetworkOptionDetailsCniTypeFlannelOverlay,
}

// GetNodePoolPodNetworkOptionDetailsCniTypeEnumValues Enumerates the set of values for NodePoolPodNetworkOptionDetailsCniTypeEnum
func GetNodePoolPodNetworkOptionDetailsCniTypeEnumValues() []NodePoolPodNetworkOptionDetailsCniTypeEnum {
	values := make([]NodePoolPodNetworkOptionDetailsCniTypeEnum, 0)
	for _, v := range mappingNodePoolPodNetworkOptionDetailsCniTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetNodePoolPodNetworkOptionDetailsCniTypeEnumStringValues Enumerates the set of values in String for NodePoolPodNetworkOptionDetailsCniTypeEnum
func GetNodePoolPodNetworkOptionDetailsCniTypeEnumStringValues() []string {
	return []string{
		"OCI_VCN_IP_NATIVE",
		"FLANNEL_OVERLAY",
	}
}

// GetMappingNodePoolPodNetworkOptionDetailsCniTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingNodePoolPodNetworkOptionDetailsCniTypeEnum(val string) (NodePoolPodNetworkOptionDetailsCniTypeEnum, bool) {
	enum, ok := mappingNodePoolPodNetworkOptionDetailsCniTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
