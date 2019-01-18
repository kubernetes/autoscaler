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

// BaseCredential is deprecated: Use AccessKeyCredential in this package instead.
type BaseCredential struct {
	AccessKeyId     string
	AccessKeySecret string
}

// AccessKeyCredential is kind of credential
type AccessKeyCredential struct {
	AccessKeyId     string
	AccessKeySecret string
}

// NewBaseCredential is deprecated: Use NewAccessKeyCredential in this package instead.
func NewBaseCredential(accessKeyId, accessKeySecret string) *BaseCredential {
	return &BaseCredential{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
}

// ToAccessKeyCredential returns AccessKeyCredential
func (baseCred *BaseCredential) ToAccessKeyCredential() *AccessKeyCredential {
	return &AccessKeyCredential{
		AccessKeyId:     baseCred.AccessKeyId,
		AccessKeySecret: baseCred.AccessKeySecret,
	}
}

// NewAccessKeyCredential returns AccessKeyCredential
func NewAccessKeyCredential(accessKeyId, accessKeySecret string) *AccessKeyCredential {
	return &AccessKeyCredential{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
}
