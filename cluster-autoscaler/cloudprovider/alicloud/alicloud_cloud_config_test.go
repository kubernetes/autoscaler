package alicloud

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccessKeyCloudConfigIsValid(t *testing.T) {
	t.Setenv(accessKeyId, "id")
	t.Setenv(accessKeySecret, "secret")
	t.Setenv(regionId, "cn-hangzhou")

	cfg := &cloudConfig{}
	assert.True(t, cfg.isValid())
	assert.False(t, cfg.RRSAEnabled)
}

func TestRRSACloudConfigIsValid(t *testing.T) {
	t.Setenv(oidcProviderARN, "acs:ram::12345:oidc-provider/ack-rrsa-cb123")
	t.Setenv(oidcTokenFilePath, "/var/run/secrets/tokens/oidc-token")
	t.Setenv(roleARN, "acs:ram::12345:role/autoscaler-role")
	t.Setenv(roleSessionName, "session")
	t.Setenv(regionId, "cn-hangzhou")

	cfg := &cloudConfig{}
	assert.True(t, cfg.isValid())
	assert.True(t, cfg.RRSAEnabled)
}
