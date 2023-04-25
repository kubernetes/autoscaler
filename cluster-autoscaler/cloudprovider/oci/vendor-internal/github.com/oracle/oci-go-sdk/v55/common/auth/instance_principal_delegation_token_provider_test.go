// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
)

func TestInstancePrincipalDelegationTokenConfigurationProvider_ErrorInput(t *testing.T) {
	delegationToken := ""
	region := common.StringToRegion("us-ashburn-1")
	configurationProvider, err := InstancePrincipalDelegationTokenConfigurationProvider(&delegationToken)
	assert.Nil(t, configurationProvider)
	assert.NotNil(t, err)

	configurationProvider, err = InstancePrincipalDelegationTokenConfigurationProvider(nil)
	assert.Nil(t, configurationProvider)
	assert.NotNil(t, err)

	configurationProviderForRegion, err := InstancePrincipalDelegationTokenConfigurationProviderForRegion(&delegationToken, region)
	assert.Nil(t, configurationProviderForRegion)
	assert.NotNil(t, err)

	configurationProviderForRegionNotFound, err := InstancePrincipalDelegationTokenConfigurationProviderForRegion(&delegationToken, "")
	assert.Nil(t, configurationProviderForRegionNotFound)
	assert.NotNil(t, err)
}
