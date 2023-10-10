/*
Copyright 2021 The Kubernetes Authors.

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

package tencentcloud

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common/profile"

// Tencent Cloud Service Http Endpoint
const (
	ASHttpEndpoint  = "as.tencentcloudapi.com"
	VPCHttpEndpoint = "vpc.tencentcloudapi.com"
	CVMHttpEndpoint = "cvm.tencentcloudapi.com"
)

func newASClientProfile() *profile.ClientProfile {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = ASHttpEndpoint
	return cpf
}

func newVPCClientProfile() *profile.ClientProfile {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = VPCHttpEndpoint
	return cpf
}

func newCVMClientProfile() *profile.ClientProfile {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = CVMHttpEndpoint
	return cpf
}
