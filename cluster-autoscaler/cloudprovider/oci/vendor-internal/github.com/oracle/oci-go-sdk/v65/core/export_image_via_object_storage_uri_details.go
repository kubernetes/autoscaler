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

// ExportImageViaObjectStorageUriDetails The representation of ExportImageViaObjectStorageUriDetails
type ExportImageViaObjectStorageUriDetails struct {

	// The Object Storage URL to export the image to. See Object
	// Storage URLs (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/imageimportexport.htm#URLs)
	// and Using Pre-Authenticated Requests (https://docs.cloud.oracle.com/iaas/Content/Object/Tasks/usingpreauthenticatedrequests.htm)
	// for constructing URLs for image import/export.
	DestinationUri *string `mandatory:"true" json:"destinationUri"`

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
func (m ExportImageViaObjectStorageUriDetails) GetExportFormat() ExportImageDetailsExportFormatEnum {
	return m.ExportFormat
}

func (m ExportImageViaObjectStorageUriDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ExportImageViaObjectStorageUriDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingExportImageDetailsExportFormatEnum(string(m.ExportFormat)); !ok && m.ExportFormat != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ExportFormat: %s. Supported values are: %s.", m.ExportFormat, strings.Join(GetExportImageDetailsExportFormatEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m ExportImageViaObjectStorageUriDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeExportImageViaObjectStorageUriDetails ExportImageViaObjectStorageUriDetails
	s := struct {
		DiscriminatorParam string `json:"destinationType"`
		MarshalTypeExportImageViaObjectStorageUriDetails
	}{
		"objectStorageUri",
		(MarshalTypeExportImageViaObjectStorageUriDetails)(m),
	}

	return json.Marshal(&s)
}
