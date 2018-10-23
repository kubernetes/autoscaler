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

// AccessKeySigner is wrapper of AccessKeyCredential
type AccessKeySigner struct {
	credential *credentials.AccessKeyCredential
}

// GetExtraParam doesn't implement
func (signer *AccessKeySigner) GetExtraParam() map[string]string {
	return nil
}

// NewAccessKeySigner returns AccessKeySigner
func NewAccessKeySigner(credential *credentials.AccessKeyCredential) (*AccessKeySigner, error) {
	return &AccessKeySigner{
		credential: credential,
	}, nil
}

// GetName returns "HMAC-SHA1"
func (*AccessKeySigner) GetName() string {
	return "HMAC-SHA1"
}

// GetType returns ""
func (*AccessKeySigner) GetType() string {
	return ""
}

// GetVersion returns "1.0"
func (*AccessKeySigner) GetVersion() string {
	return "1.0"
}

// GetAccessKeyId returns accessKeyId
func (signer *AccessKeySigner) GetAccessKeyId() (accessKeyId string, err error) {
	return signer.credential.AccessKeyId, nil
}

// Sign returns a signer
func (signer *AccessKeySigner) Sign(stringToSign, secretSuffix string) string {
	secret := signer.credential.AccessKeySecret + secretSuffix
	return ShaHmac1(stringToSign, secret)
}

// Shutdown doesn't implement
func (signer *AccessKeySigner) Shutdown() {}
