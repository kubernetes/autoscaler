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

// VnicWorker Details of a vnicWorker.
type VnicWorker struct {

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the `vnicWorker`.
	Id *string `mandatory:"false" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment that contains the VNIC worker.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The `vnicWorker`'s current state.
	LifecycleState VnicWorkerLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of associated service VNIC.
	ServiceVnicId *string `mandatory:"false" json:"serviceVnicId"`

	// Details of vnicWorker IPs config.
	WorkerIpsConfig []VnicWorkerIpConfig `mandatory:"false" json:"workerIpsConfig"`

	// The MAC address of the vnicWorker.
	WorkerMacAddress *string `mandatory:"false" json:"workerMacAddress"`

	// The instance where vnicWorker is attached.
	WorkerInstanceId *string `mandatory:"false" json:"workerInstanceId"`

	// Which physical network interface card (NIC) the VNIC worker uses.
	// Certain bare metal instance shapes have two active physical NICs (0 and 1). If
	// you add a VNIC worker to one of these instances, you can specify which NIC
	// the VNIC worker will use. Note that it is required for NIC to have at least a single
	// VNIC attached before attaching a VNIC worker.
	WorkerNicIndex *int `mandatory:"false" json:"workerNicIndex"`

	// The VLAN tag assigned to `vnicWorker`.
	WorkerVlanTag *int `mandatory:"false" json:"workerVlanTag"`

	WorkerPrimaryIpConfig *VnicWorkerIpConfig `mandatory:"false" json:"workerPrimaryIpConfig"`

	// Specifies whether the `vnicWorker` had been enabled for forwarding traffic.
	IsEnabled *bool `mandatory:"false" json:"isEnabled"`

	// Specifies whether the `vnicWorker` is in draining mode.
	IsDraining *bool `mandatory:"false" json:"isDraining"`
}

func (m VnicWorker) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m VnicWorker) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if _, ok := mappingVnicWorkerLifecycleStateEnum[string(m.LifecycleState)]; !ok && m.LifecycleState != "" {
		errMessage = append(errMessage, fmt.Sprintf("unsupported enum value for LifecycleState: %s. Supported values are: %s.", m.LifecycleState, strings.Join(GetVnicWorkerLifecycleStateEnumStringValues(), ",")))
	}
	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}

// VnicWorkerLifecycleStateEnum Enum with underlying type: string
type VnicWorkerLifecycleStateEnum string

// Set of constants representing the allowable values for VnicWorkerLifecycleStateEnum
const (
	VnicWorkerLifecycleStateProvisioning VnicWorkerLifecycleStateEnum = "PROVISIONING"
	VnicWorkerLifecycleStateAvailable    VnicWorkerLifecycleStateEnum = "AVAILABLE"
	VnicWorkerLifecycleStateTerminating  VnicWorkerLifecycleStateEnum = "TERMINATING"
	VnicWorkerLifecycleStateTerminated   VnicWorkerLifecycleStateEnum = "TERMINATED"
)

var mappingVnicWorkerLifecycleStateEnum = map[string]VnicWorkerLifecycleStateEnum{
	"PROVISIONING": VnicWorkerLifecycleStateProvisioning,
	"AVAILABLE":    VnicWorkerLifecycleStateAvailable,
	"TERMINATING":  VnicWorkerLifecycleStateTerminating,
	"TERMINATED":   VnicWorkerLifecycleStateTerminated,
}

// GetVnicWorkerLifecycleStateEnumValues Enumerates the set of values for VnicWorkerLifecycleStateEnum
func GetVnicWorkerLifecycleStateEnumValues() []VnicWorkerLifecycleStateEnum {
	values := make([]VnicWorkerLifecycleStateEnum, 0)
	for _, v := range mappingVnicWorkerLifecycleStateEnum {
		values = append(values, v)
	}
	return values
}

// GetVnicWorkerLifecycleStateEnumStringValues Enumerates the set of values in String for VnicWorkerLifecycleStateEnum
func GetVnicWorkerLifecycleStateEnumStringValues() []string {
	return []string{
		"PROVISIONING",
		"AVAILABLE",
		"TERMINATING",
		"TERMINATED",
	}
}
