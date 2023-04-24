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

package alicloud

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRRSACloudConfigEssClientCreation(t *testing.T) {
	t.Setenv(oidcProviderARN, "acs:ram::12345:oidc-provider/ack-rrsa-cb123")
	t.Setenv(oidcTokenFilePath, "/var/run/secrets/tokens/oidc-token")
	t.Setenv(roleARN, "acs:ram::12345:role/autoscaler-role")
	t.Setenv(roleSessionName, "session")
	t.Setenv(regionId, "cn-hangzhou")

	cfg := &cloudConfig{}
	assert.True(t, cfg.isValid())
	assert.True(t, cfg.RRSAEnabled)

	client, err := getEssClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
