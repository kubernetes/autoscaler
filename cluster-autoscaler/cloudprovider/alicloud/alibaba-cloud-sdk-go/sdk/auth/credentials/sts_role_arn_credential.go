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

package credentials

// StsRoleArnCredential is deprecated: Use RamRoleArnCredential in this package instead.
type StsRoleArnCredential struct {
	AccessKeyId           string
	AccessKeySecret       string
	RoleArn               string
	RoleSessionName       string
	RoleSessionExpiration int
}

// RamRoleArnCredential is going to replace StsRoleArnCredential
type RamRoleArnCredential struct {
	AccessKeyId           string
	AccessKeySecret       string
	RoleArn               string
	RoleSessionName       string
	RoleSessionExpiration int
}

// NewStsRoleArnCredential is deprecated: Use RamRoleArnCredential in this package instead.
func NewStsRoleArnCredential(accessKeyId, accessKeySecret, roleArn, roleSessionName string, roleSessionExpiration int) *StsRoleArnCredential {
	return &StsRoleArnCredential{
		AccessKeyId:           accessKeyId,
		AccessKeySecret:       accessKeySecret,
		RoleArn:               roleArn,
		RoleSessionName:       roleSessionName,
		RoleSessionExpiration: roleSessionExpiration,
	}
}

// ToRamRoleArnCredential returns RamRoleArnCredential
func (oldCred *StsRoleArnCredential) ToRamRoleArnCredential() *RamRoleArnCredential {
	return &RamRoleArnCredential{
		AccessKeyId:           oldCred.AccessKeyId,
		AccessKeySecret:       oldCred.AccessKeySecret,
		RoleArn:               oldCred.RoleArn,
		RoleSessionName:       oldCred.RoleSessionName,
		RoleSessionExpiration: oldCred.RoleSessionExpiration,
	}
}

// NewRamRoleArnCredential returns RamRoleArnCredential
func NewRamRoleArnCredential(accessKeyId, accessKeySecret, roleArn, roleSessionName string, roleSessionExpiration int) *RamRoleArnCredential {
	return &RamRoleArnCredential{
		AccessKeyId:           accessKeyId,
		AccessKeySecret:       accessKeySecret,
		RoleArn:               roleArn,
		RoleSessionName:       roleSessionName,
		RoleSessionExpiration: roleSessionExpiration,
	}
}
