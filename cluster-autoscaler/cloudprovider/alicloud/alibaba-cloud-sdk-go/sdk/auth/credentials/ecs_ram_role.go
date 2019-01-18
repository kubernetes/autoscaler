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

// StsRoleNameOnEcsCredential is deprecated: Use EcsRamRoleCredential in this package instead.
type StsRoleNameOnEcsCredential struct {
	RoleName string
}

// NewStsRoleNameOnEcsCredential is deprecated: Use NewEcsRamRoleCredential in this package instead.
func NewStsRoleNameOnEcsCredential(roleName string) *StsRoleNameOnEcsCredential {
	return &StsRoleNameOnEcsCredential{
		RoleName: roleName,
	}
}

// ToEcsRamRoleCredential is deprecated
func (oldCred *StsRoleNameOnEcsCredential) ToEcsRamRoleCredential() *EcsRamRoleCredential {
	return &EcsRamRoleCredential{
		RoleName: oldCred.RoleName,
	}
}

// EcsRamRoleCredential is kind of credential on ECS
type EcsRamRoleCredential struct {
	RoleName string
}

// NewEcsRamRoleCredential returns EcsRamRoleCredential
func NewEcsRamRoleCredential(roleName string) *EcsRamRoleCredential {
	return &EcsRamRoleCredential{
		RoleName: roleName,
	}
}
