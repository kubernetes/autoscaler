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

// CreateScanProxyDetails Details for adding a RAC's scan listener information.
type CreateScanProxyDetails struct {

	// Type indicating whether Scan listener is specified by its FQDN or list of IPs
	ScanListenerType ScanProxyScanListenerTypeEnum `mandatory:"true" json:"scanListenerType"`

	// The FQDN/IPs and port information of customer's Real Application Cluster (RAC)'s SCAN
	// listeners.
	ScanListenerInfo []ScanListenerInfo `mandatory:"true" json:"scanListenerInfo"`

	// The protocol to be used for communication between client, scanProxy and RAC's scan
	// listeners
	Protocol ScanProxyProtocolEnum `mandatory:"false" json:"protocol,omitempty"`

	ScanListenerWallet *WalletInfo `mandatory:"false" json:"scanListenerWallet"`
}

func (m CreateScanProxyDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CreateScanProxyDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingScanProxyScanListenerTypeEnum[string(m.ScanListenerType)]; !ok && m.ScanListenerType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ScanListenerType: %s. Supported values are: %s.", m.ScanListenerType, strings.Join(GetScanProxyScanListenerTypeEnumStringValues(), ",")))
	}

	if _, ok := mappingScanProxyProtocolEnum[string(m.Protocol)]; !ok && m.Protocol != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Protocol: %s. Supported values are: %s.", m.Protocol, strings.Join(GetScanProxyProtocolEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
