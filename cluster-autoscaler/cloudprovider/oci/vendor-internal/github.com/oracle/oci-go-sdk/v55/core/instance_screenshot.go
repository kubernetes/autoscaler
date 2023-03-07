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
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"strings"
)

// InstanceScreenshot An instance's screenshot. It is a screen capture of the current state of the instance.
// Generally it will show the login screen of the machine.
type InstanceScreenshot struct {

	// The OCID of the screenshot object.
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the instance this screenshot was fetched from.
	InstanceId *string `mandatory:"true" json:"instanceId"`

	// The current state of the screenshot capture.
	LifecycleState InstanceScreenshotLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the screenshot was captured, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`
}

func (m InstanceScreenshot) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m InstanceScreenshot) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := mappingInstanceScreenshotLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetInstanceScreenshotLifecycleStateEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// InstanceScreenshotLifecycleStateEnum Enum with underlying type: string
type InstanceScreenshotLifecycleStateEnum string

// Set of constants representing the allowable values for InstanceScreenshotLifecycleStateEnum
const (
	InstanceScreenshotLifecycleStateCreating InstanceScreenshotLifecycleStateEnum = "CREATING"
	InstanceScreenshotLifecycleStateActive   InstanceScreenshotLifecycleStateEnum = "ACTIVE"
	InstanceScreenshotLifecycleStateFailed   InstanceScreenshotLifecycleStateEnum = "FAILED"
	InstanceScreenshotLifecycleStateDeleted  InstanceScreenshotLifecycleStateEnum = "DELETED"
	InstanceScreenshotLifecycleStateDeleting InstanceScreenshotLifecycleStateEnum = "DELETING"
)

var mappingInstanceScreenshotLifecycleStateEnum = map[string]InstanceScreenshotLifecycleStateEnum{
	"CREATING": InstanceScreenshotLifecycleStateCreating,
	"ACTIVE":   InstanceScreenshotLifecycleStateActive,
	"FAILED":   InstanceScreenshotLifecycleStateFailed,
	"DELETED":  InstanceScreenshotLifecycleStateDeleted,
	"DELETING": InstanceScreenshotLifecycleStateDeleting,
}

// GetInstanceScreenshotLifecycleStateEnumValues Enumerates the set of values for InstanceScreenshotLifecycleStateEnum
func GetInstanceScreenshotLifecycleStateEnumValues() []InstanceScreenshotLifecycleStateEnum {
	values := make([]InstanceScreenshotLifecycleStateEnum, 0)
	for _, v := range mappingInstanceScreenshotLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetInstanceScreenshotLifecycleStateEnumStringValues Enumerates the set of values in String for InstanceScreenshotLifecycleStateEnum
func GetInstanceScreenshotLifecycleStateEnumStringValues() []string {
	return []string{
		"CREATING",
		"ACTIVE",
		"FAILED",
		"DELETED",
		"DELETING",
	}
}
