// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// Use the Core Services API to manage resources such as virtual cloud networks (VCNs),
// compute instances, and block storage volumes. For more information, see the console
// documentation for the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services.
// The required permissions are documented in the
// Details for the Core Services (https://docs.cloud.oracle.com/iaas/Content/Identity/Reference/corepolicyreference.htm) article.
//

package core

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"strings"
)

// TunnelPhaseOneDetails IPSec tunnel details specific to ISAKMP phase one.
type TunnelPhaseOneDetails struct {

	// Indicates whether custom phase one configuration is enabled.
	// If this option is not enabled, default settings are proposed.
	IsCustomPhaseOneConfig *bool `mandatory:"false" json:"isCustomPhaseOneConfig"`

	// The total configured lifetime of the IKE security association.
	Lifetime *int64 `mandatory:"false" json:"lifetime"`

	// The remaining lifetime before the key is refreshed.
	RemainingLifetime *int64 `mandatory:"false" json:"remainingLifetime"`

	// The proposed custom authentication algorithm.
	CustomAuthenticationAlgorithm *string `mandatory:"false" json:"customAuthenticationAlgorithm"`

	// The negotiated authentication algorithm.
	NegotiatedAuthenticationAlgorithm *string `mandatory:"false" json:"negotiatedAuthenticationAlgorithm"`

	// The proposed custom encryption algorithm.
	CustomEncryptionAlgorithm *string `mandatory:"false" json:"customEncryptionAlgorithm"`

	// The negotiated encryption algorithm.
	NegotiatedEncryptionAlgorithm *string `mandatory:"false" json:"negotiatedEncryptionAlgorithm"`

	// The proposed custom Diffie-Hellman group.
	CustomDhGroup *string `mandatory:"false" json:"customDhGroup"`

	// The negotiated Diffie-Hellman group.
	NegotiatedDhGroup *string `mandatory:"false" json:"negotiatedDhGroup"`

	// Indicates whether IKE phase one is established.
	IsIkeEstablished *bool `mandatory:"false" json:"isIkeEstablished"`

	// The date and time we retrieved the remaining lifetime, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	RemainingLifetimeLastRetrieved *common.SDKTime `mandatory:"false" json:"remainingLifetimeLastRetrieved"`
}

func (m TunnelPhaseOneDetails) String() string {
	return common.PointerString(m)
}

// ValidateEnumValue returns an error when providing an unsupported enum value
// This function is being called during constructing API request process
// Not recommended for calling this function directly
func (m TunnelPhaseOneDetails) ValidateEnumValue() (bool, error) {
	errMessage := []string{}

	if len(errMessage) > 0 {
		return true, fmt.Errorf(strings.Join(errMessage, "\n"))
	}
	return false, nil
}
