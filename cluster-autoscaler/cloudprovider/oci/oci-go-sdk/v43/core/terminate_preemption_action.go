// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"encoding/json"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/oci-go-sdk/v43/common"
)

// TerminatePreemptionAction Terminates the preemptible instance when it is interrupted for eviction.
type TerminatePreemptionAction struct {

	// Whether to preserve the boot volume that was used to launch the preemptible instance when the instance is terminated. Defaults to false if not specified.
	PreserveBootVolume *bool `mandatory:"false" json:"preserveBootVolume"`
}

func (m TerminatePreemptionAction) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m TerminatePreemptionAction) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeTerminatePreemptionAction TerminatePreemptionAction
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeTerminatePreemptionAction
	}{
		"TERMINATE",
		(MarshalTypeTerminatePreemptionAction)(m),
	}

	return json.Marshal(&s)
}
