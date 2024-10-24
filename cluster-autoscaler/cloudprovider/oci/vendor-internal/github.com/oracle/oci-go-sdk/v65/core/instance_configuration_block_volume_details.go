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

// InstanceConfigurationBlockVolumeDetails Create new block volumes or attach to an existing volume. Specify either createDetails or volumeId.
type InstanceConfigurationBlockVolumeDetails struct {
	AttachDetails InstanceConfigurationAttachVolumeDetails `mandatory:"false" json:"attachDetails"`

	CreateDetails *InstanceConfigurationCreateVolumeDetails `mandatory:"false" json:"createDetails"`

	// The OCID of the volume.
	VolumeId *string `mandatory:"false" json:"volumeId"`
}

func (m InstanceConfigurationBlockVolumeDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceConfigurationBlockVolumeDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// UnmarshalJSON unmarshals from json
func (m *InstanceConfigurationBlockVolumeDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		AttachDetails instanceconfigurationattachvolumedetails  `json:"attachDetails"`
		CreateDetails *InstanceConfigurationCreateVolumeDetails `json:"createDetails"`
		VolumeId      *string                                   `json:"volumeId"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	nn, e = model.AttachDetails.UnmarshalPolymorphicJSON(model.AttachDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.AttachDetails = nn.(InstanceConfigurationAttachVolumeDetails)
	} else {
		m.AttachDetails = nil
	}

	m.CreateDetails = model.CreateDetails

	m.VolumeId = model.VolumeId

	return
}
