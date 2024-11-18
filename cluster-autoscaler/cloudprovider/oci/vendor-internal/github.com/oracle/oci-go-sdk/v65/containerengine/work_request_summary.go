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

// WorkRequestSummary The properties that define a work request summary.
type WorkRequestSummary struct {

	// The OCID of the work request.
	Id *string `mandatory:"false" json:"id"`

	// The type of work the work request is doing.
	OperationType WorkRequestOperationTypeEnum `mandatory:"false" json:"operationType,omitempty"`

	// The current status of the work request.
	Status WorkRequestStatusEnum `mandatory:"false" json:"status,omitempty"`

	// The OCID of the compartment in which the work request exists.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The resources this work request affects.
	Resources []WorkRequestResource `mandatory:"false" json:"resources"`

	// The time the work request was accepted.
	TimeAccepted *common.SDKTime `mandatory:"false" json:"timeAccepted"`

	// The time the work request was started.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// The time the work request was finished.
	TimeFinished *common.SDKTime `mandatory:"false" json:"timeFinished"`
}

func (m WorkRequestSummary) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m WorkRequestSummary) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := GetMappingWorkRequestOperationTypeEnum(string(m.OperationType)); !ok && m.OperationType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for OperationType: %s. Supported values are: %s.", m.OperationType, strings.Join(GetWorkRequestOperationTypeEnumStringValues(), ",")))
	}
	if _, ok := GetMappingWorkRequestStatusEnum(string(m.Status)); !ok && m.Status != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Status: %s. Supported values are: %s.", m.Status, strings.Join(GetWorkRequestStatusEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// WorkRequestSummaryOperationTypeEnum is an alias to type: WorkRequestOperationTypeEnum
// Consider using WorkRequestOperationTypeEnum instead
// Deprecated
type WorkRequestSummaryOperationTypeEnum = WorkRequestOperationTypeEnum

// Set of constants representing the allowable values for WorkRequestOperationTypeEnum
// Deprecated
const (
	WorkRequestSummaryOperationTypeClusterCreate         WorkRequestOperationTypeEnum = "CLUSTER_CREATE"
	WorkRequestSummaryOperationTypeClusterUpdate         WorkRequestOperationTypeEnum = "CLUSTER_UPDATE"
	WorkRequestSummaryOperationTypeClusterDelete         WorkRequestOperationTypeEnum = "CLUSTER_DELETE"
	WorkRequestSummaryOperationTypeCreateNamespace       WorkRequestOperationTypeEnum = "CREATE_NAMESPACE"
	WorkRequestSummaryOperationTypeNodepoolCreate        WorkRequestOperationTypeEnum = "NODEPOOL_CREATE"
	WorkRequestSummaryOperationTypeNodepoolUpdate        WorkRequestOperationTypeEnum = "NODEPOOL_UPDATE"
	WorkRequestSummaryOperationTypeNodepoolDelete        WorkRequestOperationTypeEnum = "NODEPOOL_DELETE"
	WorkRequestSummaryOperationTypeNodepoolReconcile     WorkRequestOperationTypeEnum = "NODEPOOL_RECONCILE"
	WorkRequestSummaryOperationTypeNodepoolCycling       WorkRequestOperationTypeEnum = "NODEPOOL_CYCLING"
	WorkRequestSummaryOperationTypeWorkrequestCancel     WorkRequestOperationTypeEnum = "WORKREQUEST_CANCEL"
	WorkRequestSummaryOperationTypeVirtualnodepoolCreate WorkRequestOperationTypeEnum = "VIRTUALNODEPOOL_CREATE"
	WorkRequestSummaryOperationTypeVirtualnodepoolUpdate WorkRequestOperationTypeEnum = "VIRTUALNODEPOOL_UPDATE"
	WorkRequestSummaryOperationTypeVirtualnodepoolDelete WorkRequestOperationTypeEnum = "VIRTUALNODEPOOL_DELETE"
	WorkRequestSummaryOperationTypeVirtualnodeDelete     WorkRequestOperationTypeEnum = "VIRTUALNODE_DELETE"
	WorkRequestSummaryOperationTypeEnableAddon           WorkRequestOperationTypeEnum = "ENABLE_ADDON"
	WorkRequestSummaryOperationTypeUpdateAddon           WorkRequestOperationTypeEnum = "UPDATE_ADDON"
	WorkRequestSummaryOperationTypeDisableAddon          WorkRequestOperationTypeEnum = "DISABLE_ADDON"
	WorkRequestSummaryOperationTypeReconcileAddon        WorkRequestOperationTypeEnum = "RECONCILE_ADDON"
)

// GetWorkRequestSummaryOperationTypeEnumValues Enumerates the set of values for WorkRequestOperationTypeEnum
// Consider using GetWorkRequestOperationTypeEnumValue
// Deprecated
var GetWorkRequestSummaryOperationTypeEnumValues = GetWorkRequestOperationTypeEnumValues

// WorkRequestSummaryStatusEnum is an alias to type: WorkRequestStatusEnum
// Consider using WorkRequestStatusEnum instead
// Deprecated
type WorkRequestSummaryStatusEnum = WorkRequestStatusEnum

// Set of constants representing the allowable values for WorkRequestStatusEnum
// Deprecated
const (
	WorkRequestSummaryStatusAccepted   WorkRequestStatusEnum = "ACCEPTED"
	WorkRequestSummaryStatusInProgress WorkRequestStatusEnum = "IN_PROGRESS"
	WorkRequestSummaryStatusFailed     WorkRequestStatusEnum = "FAILED"
	WorkRequestSummaryStatusSucceeded  WorkRequestStatusEnum = "SUCCEEDED"
	WorkRequestSummaryStatusCanceling  WorkRequestStatusEnum = "CANCELING"
	WorkRequestSummaryStatusCanceled   WorkRequestStatusEnum = "CANCELED"
)

// GetWorkRequestSummaryStatusEnumValues Enumerates the set of values for WorkRequestStatusEnum
// Consider using GetWorkRequestStatusEnumValue
// Deprecated
var GetWorkRequestSummaryStatusEnumValues = GetWorkRequestStatusEnumValues
