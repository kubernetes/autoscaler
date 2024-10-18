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

// FlannelOverlayClusterPodNetworkOptionDetails Network options specific to using the flannel (FLANNEL_OVERLAY) CNI
type FlannelOverlayClusterPodNetworkOptionDetails struct {
}

func (m FlannelOverlayClusterPodNetworkOptionDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m FlannelOverlayClusterPodNetworkOptionDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m FlannelOverlayClusterPodNetworkOptionDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeFlannelOverlayClusterPodNetworkOptionDetails FlannelOverlayClusterPodNetworkOptionDetails
	s := struct {
		DiscriminatorParam string `json:"cniType"`
		MarshalTypeFlannelOverlayClusterPodNetworkOptionDetails
	}{
		"FLANNEL_OVERLAY",
		(MarshalTypeFlannelOverlayClusterPodNetworkOptionDetails)(m),
	}

	return json.Marshal(&s)
}
