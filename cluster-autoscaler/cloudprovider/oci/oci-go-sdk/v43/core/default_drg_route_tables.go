// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// DefaultDrgRouteTables The default DRG route table for this DRG. Each network type
// has a default DRG route table.
// You can update a network type to use a different DRG route table, but
// each network type must have a default DRG route table. You cannot delete
// a default DRG route table.
type DefaultDrgRouteTables struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the default DRG route table to be assigned to DRG attachments
	// of type VCN on creation.
	Vcn *string `mandatory:"false" json:"vcn"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the default DRG route table assigned to DRG attachments
	// of type IPSEC_TUNNEL on creation.
	IpsecTunnel *string `mandatory:"false" json:"ipsecTunnel"`

	// The OCID of the default DRG route table to be assigned to DRG attachments
	// of type VIRTUAL_CIRCUIT on creation.
	VirtualCircuit *string `mandatory:"false" json:"virtualCircuit"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the default DRG route table to be assigned to DRG attachments
	// of type REMOTE_PEERING_CONNECTION on creation.
	RemotePeeringConnection *string `mandatory:"false" json:"remotePeeringConnection"`
}

func (m DefaultDrgRouteTables) String() string {
	return common.PointerString(m)
}
