/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sdk

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/auth/signers"
	"testing"
)

func TestRRSAClientInit(t *testing.T) {
	oidcProviderARN := "acs:ram::12345:oidc-provider/ack-rrsa-cb123"
	oidcTokenFilePath := "/var/run/secrets/tokens/oidc-token"
	roleARN := "acs:ram::12345:role/autoscaler-role"
	roleSessionName := "session"
	regionId := "cn-hangzhou"

	client, err := NewClientWithRRSA(regionId, roleARN, oidcProviderARN, oidcTokenFilePath, roleSessionName)
	assert.NoError(t, err)
	assert.IsType(t, &signers.OIDCSigner{}, client.signer)
}
