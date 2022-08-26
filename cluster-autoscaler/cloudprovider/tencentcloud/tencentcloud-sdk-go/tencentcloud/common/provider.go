/*
Copyright 2016 The Kubernetes Authors.

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

package common

import tcerr "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"

var (
	envNotSet        = tcerr.NewTencentCloudSDKError(creErr, "could not find environmental variable", "")
	fileDoseNotExist = tcerr.NewTencentCloudSDKError(creErr, "could not find config file", "")
	noCvmRole        = tcerr.NewTencentCloudSDKError(creErr, "get cvm role name failed, Please confirm whether the role is bound", "")
)

// Provider provide credential to build client.
//
// Now there are four kinds provider:
//
//	 EnvProvider : get credential from your Variable environment
//	 ProfileProvider : get credential from your profile
//		CvmRoleProvider : get credential from your cvm role
//	 RoleArnProvider : get credential from your role arn
type Provider interface {
	// GetCredential get the credential interface
	GetCredential() (CredentialIface, error)
}
