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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// CpeDeviceShapeDetail The detailed information about a particular CPE device type. Compare with
// CpeDeviceShapeSummary.
type CpeDeviceShapeDetail struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the CPE device shape.
	// This value uniquely identifies the type of CPE device.
	CpeDeviceShapeId *string `mandatory:"false" json:"cpeDeviceShapeId"`

	CpeDeviceInfo *CpeDeviceInfo `mandatory:"false" json:"cpeDeviceInfo"`

	// For certain CPE devices types, the customer can provide answers to
	// questions that are specific to the device type. This attribute contains
	// a list of those questions. The Networking service merges the answers with
	// other information and renders a set of CPE configuration content. To
	// provide the answers, use
	// UpdateTunnelCpeDeviceConfig.
	Parameters []CpeDeviceConfigQuestion `mandatory:"false" json:"parameters"`

	// A template of CPE device configuration information that will be merged with the customer's
	// answers to the questions to render the final CPE device configuration content. Also see:
	//   * GetCpeDeviceConfigContent
	//   * GetIpsecCpeDeviceConfigContent
	//   * GetTunnelCpeDeviceConfigContent
	Template *string `mandatory:"false" json:"template"`
}

func (m CpeDeviceShapeDetail) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CpeDeviceShapeDetail) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
