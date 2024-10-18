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

// CredentialRotationStatus Information regarding cluster's credential rotation.
type CredentialRotationStatus struct {

	// Credential rotation status of a kubernetes cluster
	// IN_PROGRESS: Issuing new credentials to kubernetes cluster control plane and worker nodes or retiring old credentials from kubernetes cluster control plane and worker nodes.
	// WAITING: Waiting for customer to invoke the complete rotation action or the automcatic complete rotation action.
	// COMPLETED: New credentials are functional on kuberentes cluster.
	Status CredentialRotationStatusStatusEnum `mandatory:"true" json:"status"`

	// Details of a kuberenetes cluster credential rotation status:
	// ISSUING_NEW_CREDENTIALS: Credential rotation is in progress. Starting to issue new credentials to kubernetes cluster control plane and worker nodes.
	// NEW_CREDENTIALS_ISSUED: New credentials are added. At this stage cluster has both old and new credentials and is awaiting old credentials retirement.
	// RETIRING_OLD_CREDENTIALS: Retirement of old credentials is in progress. Starting to remove old credentials from kubernetes cluster control plane and worker nodes.
	// COMPLETED: Credential rotation is complete. Old credentials are retired.
	StatusDetails CredentialRotationStatusStatusDetailsEnum `mandatory:"true" json:"statusDetails"`

	// The time by which retirement of old credentials should start.
	TimeAutoCompletionScheduled *common.SDKTime `mandatory:"false" json:"timeAutoCompletionScheduled"`
}

func (m CredentialRotationStatus) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m CredentialRotationStatus) ValidateEnumValue() (bool, error) {
	errMessage := []string{}
	if _, ok := GetMappingCredentialRotationStatusStatusEnum(string(m.Status)); !ok && m.Status != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for Status: %s. Supported values are: %s.", m.Status, strings.Join(GetCredentialRotationStatusStatusEnumStringValues(), ",")))
	}
	if _, ok := GetMappingCredentialRotationStatusStatusDetailsEnum(string(m.StatusDetails)); !ok && m.StatusDetails != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for StatusDetails: %s. Supported values are: %s.", m.StatusDetails, strings.Join(GetCredentialRotationStatusStatusDetailsEnumStringValues(), ",")))
	}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// CredentialRotationStatusStatusEnum Enum with underlying type: string
type CredentialRotationStatusStatusEnum string

// Set of constants representing the allowable values for CredentialRotationStatusStatusEnum
const (
	CredentialRotationStatusStatusInProgress CredentialRotationStatusStatusEnum = "IN_PROGRESS"
	CredentialRotationStatusStatusWaiting    CredentialRotationStatusStatusEnum = "WAITING"
	CredentialRotationStatusStatusCompleted  CredentialRotationStatusStatusEnum = "COMPLETED"
)

var mappingCredentialRotationStatusStatusEnum = map[string]CredentialRotationStatusStatusEnum{
	"IN_PROGRESS": CredentialRotationStatusStatusInProgress,
	"WAITING":     CredentialRotationStatusStatusWaiting,
	"COMPLETED":   CredentialRotationStatusStatusCompleted,
}

var mappingCredentialRotationStatusStatusEnumLowerCase = map[string]CredentialRotationStatusStatusEnum{
	"in_progress": CredentialRotationStatusStatusInProgress,
	"waiting":     CredentialRotationStatusStatusWaiting,
	"completed":   CredentialRotationStatusStatusCompleted,
}

// GetCredentialRotationStatusStatusEnumValues Enumerates the set of values for CredentialRotationStatusStatusEnum
func GetCredentialRotationStatusStatusEnumValues() []CredentialRotationStatusStatusEnum {
	values := make([]CredentialRotationStatusStatusEnum, 0)
	for _, v := range mappingCredentialRotationStatusStatusEnum {
		values = append(values, v)
	}
	return values
}

// GetCredentialRotationStatusStatusEnumStringValues Enumerates the set of values in String for CredentialRotationStatusStatusEnum
func GetCredentialRotationStatusStatusEnumStringValues() []string {
	return []string{
		"IN_PROGRESS",
		"WAITING",
		"COMPLETED",
	}
}

// GetMappingCredentialRotationStatusStatusEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCredentialRotationStatusStatusEnum(val string) (CredentialRotationStatusStatusEnum, bool) {
	enum, ok := mappingCredentialRotationStatusStatusEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}

// CredentialRotationStatusStatusDetailsEnum Enum with underlying type: string
type CredentialRotationStatusStatusDetailsEnum string

// Set of constants representing the allowable values for CredentialRotationStatusStatusDetailsEnum
const (
	CredentialRotationStatusStatusDetailsIssuingNewCredentials  CredentialRotationStatusStatusDetailsEnum = "ISSUING_NEW_CREDENTIALS"
	CredentialRotationStatusStatusDetailsNewCredentialsIssued   CredentialRotationStatusStatusDetailsEnum = "NEW_CREDENTIALS_ISSUED"
	CredentialRotationStatusStatusDetailsRetiringOldCredentials CredentialRotationStatusStatusDetailsEnum = "RETIRING_OLD_CREDENTIALS"
	CredentialRotationStatusStatusDetailsCompleted              CredentialRotationStatusStatusDetailsEnum = "COMPLETED"
)

var mappingCredentialRotationStatusStatusDetailsEnum = map[string]CredentialRotationStatusStatusDetailsEnum{
	"ISSUING_NEW_CREDENTIALS":  CredentialRotationStatusStatusDetailsIssuingNewCredentials,
	"NEW_CREDENTIALS_ISSUED":   CredentialRotationStatusStatusDetailsNewCredentialsIssued,
	"RETIRING_OLD_CREDENTIALS": CredentialRotationStatusStatusDetailsRetiringOldCredentials,
	"COMPLETED":                CredentialRotationStatusStatusDetailsCompleted,
}

var mappingCredentialRotationStatusStatusDetailsEnumLowerCase = map[string]CredentialRotationStatusStatusDetailsEnum{
	"issuing_new_credentials":  CredentialRotationStatusStatusDetailsIssuingNewCredentials,
	"new_credentials_issued":   CredentialRotationStatusStatusDetailsNewCredentialsIssued,
	"retiring_old_credentials": CredentialRotationStatusStatusDetailsRetiringOldCredentials,
	"completed":                CredentialRotationStatusStatusDetailsCompleted,
}

// GetCredentialRotationStatusStatusDetailsEnumValues Enumerates the set of values for CredentialRotationStatusStatusDetailsEnum
func GetCredentialRotationStatusStatusDetailsEnumValues() []CredentialRotationStatusStatusDetailsEnum {
	values := make([]CredentialRotationStatusStatusDetailsEnum, 0)
	for _, v := range mappingCredentialRotationStatusStatusDetailsEnum {
		values = append(values, v)
	}
	return values
}

// GetCredentialRotationStatusStatusDetailsEnumStringValues Enumerates the set of values in String for CredentialRotationStatusStatusDetailsEnum
func GetCredentialRotationStatusStatusDetailsEnumStringValues() []string {
	return []string{
		"ISSUING_NEW_CREDENTIALS",
		"NEW_CREDENTIALS_ISSUED",
		"RETIRING_OLD_CREDENTIALS",
		"COMPLETED",
	}
}

// GetMappingCredentialRotationStatusStatusDetailsEnum performs case Insensitive comparison on enum value and return the desired enum
func GetMappingCredentialRotationStatusStatusDetailsEnum(val string) (CredentialRotationStatusStatusDetailsEnum, bool) {
	enum, ok := mappingCredentialRotationStatusStatusDetailsEnumLowerCase[strings.ToLower(val)]
	return enum, ok
}
