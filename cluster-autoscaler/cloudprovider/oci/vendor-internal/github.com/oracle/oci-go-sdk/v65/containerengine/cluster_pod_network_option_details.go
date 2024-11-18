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

// ClusterPodNetworkOptionDetails The CNI type and relevant network details potentially applicable to the node pools of the cluster
type ClusterPodNetworkOptionDetails interface {
}

type clusterpodnetworkoptiondetails struct {
	JsonData []byte
	CniType  string `json:"cniType"`
}

// UnmarshalJSON unmarshals json
func (m *clusterpodnetworkoptiondetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerclusterpodnetworkoptiondetails clusterpodnetworkoptiondetails
	s := struct {
		Model Unmarshalerclusterpodnetworkoptiondetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.CniType = s.Model.CniType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *clusterpodnetworkoptiondetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.CniType {
	case "FLANNEL_OVERLAY":
		mm := FlannelOverlayClusterPodNetworkOptionDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "OCI_VCN_IP_NATIVE":
		mm := OciVcnIpNativeClusterPodNetworkOptionDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		common.Logf("Recieved unsupported enum value for ClusterPodNetworkOptionDetails: %s.", m.CniType)
		return *m, nil
	}
}

func (m clusterpodnetworkoptiondetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m clusterpodnetworkoptiondetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// ClusterPodNetworkOptionDetailsCniTypeEnum Enum with underlying type: string
type ClusterPodNetworkOptionDetailsCniTypeEnum string

// Set of constants representing the allowable values for ClusterPodNetworkOptionDetailsCniTypeEnum
const (
	ClusterPodNetworkOptionDetailsCniTypeOciVcnIpNative ClusterPodNetworkOptionDetailsCniTypeEnum = "OCI_VCN_IP_NATIVE"
	ClusterPodNetworkOptionDetailsCniTypeFlannelOverlay ClusterPodNetworkOptionDetailsCniTypeEnum = "FLANNEL_OVERLAY"
)

var mappingClusterPodNetworkOptionDetailsCniTypeEnum = map[string]ClusterPodNetworkOptionDetailsCniTypeEnum{
	"OCI_VCN_IP_NATIVE": ClusterPodNetworkOptionDetailsCniTypeOciVcnIpNative,
	"FLANNEL_OVERLAY":   ClusterPodNetworkOptionDetailsCniTypeFlannelOverlay,
}

var mappingClusterPodNetworkOptionDetailsCniTypeEnumLowerCase = map[string]ClusterPodNetworkOptionDetailsCniTypeEnum{
	"oci_vcn_ip_native": ClusterPodNetworkOptionDetailsCniTypeOciVcnIpNative,
	"flannel_overlay":   ClusterPodNetworkOptionDetailsCniTypeFlannelOverlay,
}

// GetClusterPodNetworkOptionDetailsCniTypeEnumValues Enumerates the set of values for ClusterPodNetworkOptionDetailsCniTypeEnum
func GetClusterPodNetworkOptionDetailsCniTypeEnumValues() []ClusterPodNetworkOptionDetailsCniTypeEnum {
	values := make([]ClusterPodNetworkOptionDetailsCniTypeEnum, 0)
	for _, v := range mappingClusterPodNetworkOptionDetailsCniTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetClusterPodNetworkOptionDetailsCniTypeEnumStringValues Enumerates the set of values in String for ClusterPodNetworkOptionDetailsCniTypeEnum
func GetClusterPodNetworkOptionDetailsCniTypeEnumStringValues() []string {
	return []string{
		"OCI_VCN_IP_NATIVE",
		"FLANNEL_OVERLAY",
	}
}

// GetMappingClusterPodNetworkOptionDetailsCniTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingClusterPodNetworkOptionDetailsCniTypeEnum(val string) (ClusterPodNetworkOptionDetailsCniTypeEnum, bool) {
	enum, ok := mappingClusterPodNetworkOptionDetailsCniTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
