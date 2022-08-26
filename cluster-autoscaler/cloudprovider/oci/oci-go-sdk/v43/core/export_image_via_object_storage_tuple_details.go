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
	"encoding/json"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// ExportImageViaObjectStorageTupleDetails The representation of ExportImageViaObjectStorageTupleDetails
type ExportImageViaObjectStorageTupleDetails struct {

	// The Object Storage bucket to export the image to.
	BucketName *string `mandatory:"true" json:"bucketName"`

	// The Object Storage namespace to export the image to.
	NamespaceName *string `mandatory:"true" json:"namespaceName"`

	// The Object Storage object name for the exported image.
	ObjectName *string `mandatory:"true" json:"objectName"`

	// The format to export the image to. The default value is `OCI`.
	// The following image formats are available:
	// - `OCI` - Oracle Cloud Infrastructure file with a QCOW2 image and Oracle Cloud Infrastructure metadata (.oci).
	// Use this format to export a custom image that you want to import into other tenancies or regions.
	// - `QCOW2` - QEMU Copy On Write (.qcow2)
	// - `VDI` - Virtual Disk Image (.vdi) for Oracle VM VirtualBox
	// - `VHD` - Virtual Hard Disk (.vhd) for Hyper-V
	// - `VMDK` - Virtual Machine Disk (.vmdk)
	ExportFormat ExportImageDetailsExportFormatEnum `mandatory:"false" json:"exportFormat,omitempty"`
}

// GetExportFormat returns ExportFormat
func (m ExportImageViaObjectStorageTupleDetails) GetExportFormat() ExportImageDetailsExportFormatEnum {
	return m.ExportFormat
}

func (m ExportImageViaObjectStorageTupleDetails) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m ExportImageViaObjectStorageTupleDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeExportImageViaObjectStorageTupleDetails ExportImageViaObjectStorageTupleDetails
	s := struct {
		DiscriminatorParam string `json:"destinationType"`
		MarshalTypeExportImageViaObjectStorageTupleDetails
	}{
		"objectStorageTuple",
		(MarshalTypeExportImageViaObjectStorageTupleDetails)(m),
	}

	return json.Marshal(&s)
}
