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
