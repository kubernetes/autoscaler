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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// ClusterMetadata The properties that define meta data for a cluster.
type ClusterMetadata struct {

	// The time the cluster was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The user who created the cluster.
	CreatedByUserId *string `mandatory:"false" json:"createdByUserId"`

	// The OCID of the work request which created the cluster.
	CreatedByWorkRequestId *string `mandatory:"false" json:"createdByWorkRequestId"`

	// The time the cluster was deleted.
	TimeDeleted *common.SDKTime `mandatory:"false" json:"timeDeleted"`

	// The user who deleted the cluster.
	DeletedByUserId *string `mandatory:"false" json:"deletedByUserId"`

	// The OCID of the work request which deleted the cluster.
	DeletedByWorkRequestId *string `mandatory:"false" json:"deletedByWorkRequestId"`

	// The time the cluster was updated.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// The user who updated the cluster.
	UpdatedByUserId *string `mandatory:"false" json:"updatedByUserId"`

	// The OCID of the work request which updated the cluster.
	UpdatedByWorkRequestId *string `mandatory:"false" json:"updatedByWorkRequestId"`

	// The time until which the cluster credential is valid.
	TimeCredentialExpiration *common.SDKTime `mandatory:"false" json:"timeCredentialExpiration"`
}

func (m ClusterMetadata) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m ClusterMetadata) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
