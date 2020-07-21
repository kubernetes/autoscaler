/*
Copyright 2020 The Kubernetes Authors.

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

package tokens

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"

// CreateURL generates the URL used to create new Tokens.
func CreateURL(client *huaweicloudsdk.ServiceClient) string {
	return client.ServiceURL("tokens")
}

// GetURL generates the URL used to Validate Tokens.
func GetURL(client *huaweicloudsdk.ServiceClient, token string) string {
	return client.ServiceURL("tokens", token)
}
