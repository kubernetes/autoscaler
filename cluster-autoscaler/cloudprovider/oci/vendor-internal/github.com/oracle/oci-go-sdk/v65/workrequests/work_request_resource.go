// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// WorkRequestResource A resource that is created or operated on by an asynchronous operation that is tracked by
// a work request.
type WorkRequestResource struct {

	// The way in which this resource was affected by the operation that spawned the work
	// request.
	ActionType WorkRequestResourceActionTypeEnum `mandatory:"true" json:"actionType"`

	// The resource type the work request affects.
	EntityType *string `mandatory:"true" json:"entityType"`

	// An OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) or other unique identifier for the
	// resource.
	Identifier *string `mandatory:"true" json:"identifier"`

	// The URI path that you can use for a GET request to access the resource metadata.
	EntityUri *string `mandatory:"false" json:"entityUri"`
}

func (m WorkRequestResource) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m WorkRequestResource) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingWorkRequestResourceActionTypeEnum(string(m.ActionType)); !ok && m.ActionType != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for ActionType: %s. Supported values are: %s.", m.ActionType, strings.Join(GetWorkRequestResourceActionTypeEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// WorkRequestResourceActionTypeEnum Enum with underlying type: string
type WorkRequestResourceActionTypeEnum string

// Set of constants representing the allowable values for WorkRequestResourceActionTypeEnum
const (
	WorkRequestResourceActionTypeCreated    WorkRequestResourceActionTypeEnum = "CREATED"
	WorkRequestResourceActionTypeUpdated    WorkRequestResourceActionTypeEnum = "UPDATED"
	WorkRequestResourceActionTypeDeleted    WorkRequestResourceActionTypeEnum = "DELETED"
	WorkRequestResourceActionTypeRelated    WorkRequestResourceActionTypeEnum = "RELATED"
	WorkRequestResourceActionTypeInProgress WorkRequestResourceActionTypeEnum = "IN_PROGRESS"
)

var mappingWorkRequestResourceActionTypeEnum = map[string]WorkRequestResourceActionTypeEnum{
	"CREATED":     WorkRequestResourceActionTypeCreated,
	"UPDATED":     WorkRequestResourceActionTypeUpdated,
	"DELETED":     WorkRequestResourceActionTypeDeleted,
	"RELATED":     WorkRequestResourceActionTypeRelated,
	"IN_PROGRESS": WorkRequestResourceActionTypeInProgress,
}

var mappingWorkRequestResourceActionTypeEnumLowerCase = map[string]WorkRequestResourceActionTypeEnum{
	"created":     WorkRequestResourceActionTypeCreated,
	"updated":     WorkRequestResourceActionTypeUpdated,
	"deleted":     WorkRequestResourceActionTypeDeleted,
	"related":     WorkRequestResourceActionTypeRelated,
	"in_progress": WorkRequestResourceActionTypeInProgress,
}

// GetWorkRequestResourceActionTypeEnumValues Enumerates the set of values for WorkRequestResourceActionTypeEnum
func GetWorkRequestResourceActionTypeEnumValues() []WorkRequestResourceActionTypeEnum {
	values := make([]WorkRequestResourceActionTypeEnum, 0)
	for _, v := range mappingWorkRequestResourceActionTypeEnum {
		values = append(values, v)
	}
	return values
}

// GetWorkRequestResourceActionTypeEnumStringValues Enumerates the set of values in String for WorkRequestResourceActionTypeEnum
func GetWorkRequestResourceActionTypeEnumStringValues() []string {
	return []string{
		"CREATED",
		"UPDATED",
		"DELETED",
		"RELATED",
		"IN_PROGRESS",
	}
}

// GetMappingWorkRequestResourceActionTypeEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingWorkRequestResourceActionTypeEnum(val string) (WorkRequestResourceActionTypeEnum, bool) {
	enum, ok := mappingWorkRequestResourceActionTypeEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
