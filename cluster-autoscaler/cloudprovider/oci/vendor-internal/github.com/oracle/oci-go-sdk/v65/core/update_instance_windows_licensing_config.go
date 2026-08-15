// Copyright (c) 2016, 2018, 2025, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// UpdateInstanceWindowsLicensingConfig The default windows licensing config.
type UpdateInstanceWindowsLicensingConfig struct {

	// License Type for the OS license.
	// * `OCI_PROVIDED` - OCI provided license (e.g. metered $/OCPU-hour).
	// * `BRING_YOUR_OWN_LICENSE` - Bring your own license.
	LicenseType UpdateInstanceLicensingConfigLicenseTypeEnum `mandatory:"false" json:"licenseType,omitempty"`
}

// GetLicenseType returns LicenseType
func (m UpdateInstanceWindowsLicensingConfig) GetLicenseType() UpdateInstanceLicensingConfigLicenseTypeEnum {
	return m.LicenseType
}

func (m UpdateInstanceWindowsLicensingConfig) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m UpdateInstanceWindowsLicensingConfig) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingUpdateInstanceLicensingConfigLicenseTypeEnum(string(m.LicenseType)); !ok && m.LicenseType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LicenseType: %s. Supported values are: %s.", m.LicenseType, strings.Join(GetUpdateInstanceLicensingConfigLicenseTypeEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m UpdateInstanceWindowsLicensingConfig) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeUpdateInstanceWindowsLicensingConfig UpdateInstanceWindowsLicensingConfig
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeUpdateInstanceWindowsLicensingConfig
	}{
		"WINDOWS",
		(MarshalTypeUpdateInstanceWindowsLicensingConfig)(m),
	}

	return json.Marshal(&s)
}
