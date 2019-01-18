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

package auth

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/auth/signers"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/requests"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/responses"
	"reflect"
)

// Signer sign client token
type Signer interface {
	GetName() string
	GetType() string
	GetVersion() string
	GetAccessKeyId() (string, error)
	GetExtraParam() map[string]string
	Sign(stringToSign, secretSuffix string) string
	Shutdown()
}

// NewSignerWithCredential create signer with credential
func NewSignerWithCredential(credential Credential, commonApi func(request *requests.CommonRequest, signer interface{}) (response *responses.CommonResponse, err error)) (signer Signer, err error) {
	switch instance := credential.(type) {
	case *credentials.AccessKeyCredential:
		{
			signer, err = signers.NewAccessKeySigner(instance)
		}
	case *credentials.StsTokenCredential:
		{
			signer, err = signers.NewStsTokenSigner(instance)
		}

	case *credentials.RamRoleArnCredential:
		{
			signer, err = signers.NewRamRoleArnSigner(instance, commonApi)
		}
	case *credentials.RsaKeyPairCredential:
		{
			signer, err = signers.NewSignerKeyPair(instance, commonApi)
		}
	case *credentials.EcsRamRoleCredential:
		{
			signer, err = signers.NewEcsRamRoleSigner(instance, commonApi)
		}
	case *credentials.BaseCredential: // deprecated user interface
		{
			signer, err = signers.NewAccessKeySigner(instance.ToAccessKeyCredential())
		}
	case *credentials.StsRoleArnCredential: // deprecated user interface
		{
			signer, err = signers.NewRamRoleArnSigner(instance.ToRamRoleArnCredential(), commonApi)
		}
	case *credentials.StsRoleNameOnEcsCredential: // deprecated user interface
		{
			signer, err = signers.NewEcsRamRoleSigner(instance.ToEcsRamRoleCredential(), commonApi)
		}
	default:
		message := fmt.Sprintf(errors.UnsupportedCredentialErrorMessage, reflect.TypeOf(credential))
		err = errors.NewClientError(errors.UnsupportedCredentialErrorCode, message, nil)
	}
	return
}

// Sign will generate signer token
func Sign(request requests.AcsRequest, signer Signer, regionId string) (err error) {
	switch request.GetStyle() {
	case requests.ROA:
		{
			signRoaRequest(request, signer, regionId)
		}
	case requests.RPC:
		{
			err = signRpcRequest(request, signer, regionId)
		}
	default:
		message := fmt.Sprintf(errors.UnknownRequestTypeErrorMessage, reflect.TypeOf(request))
		err = errors.NewClientError(errors.UnknownRequestTypeErrorCode, message, nil)
	}

	return
}
