// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Work Requests API
//
// Many of the API operations that you use to create and configure Compute resources do not take effect
// immediately. In these cases, the operation spawns an asynchronous workflow to fulfill the request.
// Work requests provide visibility into the status of these in-progress, long-running workflows.
// For more information about work requests and the operations that spawn work requests, see
// Viewing the State of a Compute Work Request (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/viewingworkrequestcompute.htm).
//

package workrequests

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// WorkRequest An asynchronous work request.
type WorkRequest struct {

	// The asynchronous operation tracked by this work request.
	OperationType *string `mandatory:"true" json:"operationType"`

	// The status of the work request.
	Status WorkRequestStatusEnum `mandatory:"true" json:"status"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the work request.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment
	// that contains the work request.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The resources that are affected by this work request.
	Resources []WorkRequestResource `mandatory:"true" json:"resources"`

	// The percentage complete of the operation tracked by this work request.
	PercentComplete *float32 `mandatory:"true" json:"percentComplete"`

	// The date and time the work request was created, in the format defined by RFC3339.
	TimeAccepted *common.SDKTime `mandatory:"true" json:"timeAccepted"`

	// The date and time the work request transitioned from `ACCEPTED` to `IN_PROGRESS`,
	// in the format defined by RFC3339.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// The date and time the work request reached a terminal state, either `FAILED` or `SUCCEEDED`.
	// Format is defined by RFC3339.
	TimeFinished *common.SDKTime `mandatory:"false" json:"timeFinished"`
}

func (m WorkRequest) String() string {
	return common.PointerString(m)
}

// WorkRequestStatusEnum Enum with underlying type: string
type WorkRequestStatusEnum string

// Set of constants representing the allowable values for WorkRequestStatusEnum
const (
	WorkRequestStatusAccepted   WorkRequestStatusEnum = "ACCEPTED"
	WorkRequestStatusInProgress WorkRequestStatusEnum = "IN_PROGRESS"
	WorkRequestStatusFailed     WorkRequestStatusEnum = "FAILED"
	WorkRequestStatusSucceeded  WorkRequestStatusEnum = "SUCCEEDED"
	WorkRequestStatusCanceling  WorkRequestStatusEnum = "CANCELING"
	WorkRequestStatusCanceled   WorkRequestStatusEnum = "CANCELED"
)

var mappingWorkRequestStatus = map[string]WorkRequestStatusEnum{
	"ACCEPTED":    WorkRequestStatusAccepted,
	"IN_PROGRESS": WorkRequestStatusInProgress,
	"FAILED":      WorkRequestStatusFailed,
	"SUCCEEDED":   WorkRequestStatusSucceeded,
	"CANCELING":   WorkRequestStatusCanceling,
	"CANCELED":    WorkRequestStatusCanceled,
}

// GetWorkRequestStatusEnumValues Enumerates the set of values for WorkRequestStatusEnum
func GetWorkRequestStatusEnumValues() []WorkRequestStatusEnum {
	values := make([]WorkRequestStatusEnum, 0)
	for _, v := range mappingWorkRequestStatus {
		values = append(values, v)
	}
	return values
}
