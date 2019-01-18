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

package signers

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/auth/credentials"
)

// StsTokenSigner is kind of Signer
type StsTokenSigner struct {
	credential *credentials.StsTokenCredential
}

// NewStsTokenSigner returns StsTokenSigner
func NewStsTokenSigner(credential *credentials.StsTokenCredential) (*StsTokenSigner, error) {
	return &StsTokenSigner{
		credential: credential,
	}, nil
}

// GetName returns "HMAC-SHA1"
func (*StsTokenSigner) GetName() string {
	return "HMAC-SHA1"
}

// GetType returns ""
func (*StsTokenSigner) GetType() string {
	return ""
}

// GetVersion returns ""
func (*StsTokenSigner) GetVersion() string {
	return "1.0"
}

// GetAccessKeyId returns accessKeyId
func (signer *StsTokenSigner) GetAccessKeyId() (accessKeyId string, err error) {
	return signer.credential.AccessKeyId, nil
}

// GetExtraParam returns params
func (signer *StsTokenSigner) GetExtraParam() map[string]string {
	return map[string]string{"SecurityToken": signer.credential.AccessKeyStsToken}
}

// Sign creates signer
func (signer *StsTokenSigner) Sign(stringToSign, secretSuffix string) string {
	secret := signer.credential.AccessKeySecret + secretSuffix
	return ShaHmac1(stringToSign, secret)
}

// Shutdown doesn't implement
func (signer *StsTokenSigner) Shutdown() {}
