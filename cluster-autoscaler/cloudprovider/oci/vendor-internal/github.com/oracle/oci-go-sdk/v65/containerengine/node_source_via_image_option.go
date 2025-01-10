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

// NodeSourceViaImageOption An image can be specified as the source of nodes when launching a node pool using the `nodeSourceDetails` object.
type NodeSourceViaImageOption struct {

	// The user-friendly name of the entity corresponding to the OCID.
	SourceName *string `mandatory:"false" json:"sourceName"`

	// The OCID of the image.
	ImageId *string `mandatory:"false" json:"imageId"`
}

// GetSourceName returns SourceName
func (m NodeSourceViaImageOption) GetSourceName() *string {
	return m.SourceName
}

func (m NodeSourceViaImageOption) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m NodeSourceViaImageOption) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m NodeSourceViaImageOption) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeNodeSourceViaImageOption NodeSourceViaImageOption
	s := struct {
		DiscriminatorParam string `json:"sourceType"`
		MarshalTypeNodeSourceViaImageOption
	}{
		"IMAGE",
		(MarshalTypeNodeSourceViaImageOption)(m),
	}

	return json.Marshal(&s)
}
