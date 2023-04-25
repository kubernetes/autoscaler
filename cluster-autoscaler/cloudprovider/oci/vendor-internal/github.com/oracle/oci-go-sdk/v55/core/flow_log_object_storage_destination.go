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
	"encoding/json"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// FlowLogObjectStorageDestination Information to identify the Object Storage bucket where the flow logs will be stored.
type FlowLogObjectStorageDestination struct {

	// The Object Storage bucket name.
	BucketName *string `mandatory:"false" json:"bucketName"`

	// The Object Storage namespace.
	NamespaceName *string `mandatory:"false" json:"namespaceName"`
}

func (m FlowLogObjectStorageDestination) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m FlowLogObjectStorageDestination) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// MarshalJSON marshals to json representation
func (m FlowLogObjectStorageDestination) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeFlowLogObjectStorageDestination FlowLogObjectStorageDestination
	s := struct {
		DiscriminatorParam string `json:"destinationType"`
		MarshalTypeFlowLogObjectStorageDestination
	}{
		"OBJECT_STORAGE",
		(MarshalTypeFlowLogObjectStorageDestination)(m),
	}

	return json.Marshal(&s)
}
