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

import (
	tcerr "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tchttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

func NewCommonClient(cred CredentialIface, region string, clientProfile *profile.ClientProfile) (c *Client) {
	return new(Client).Init(region).WithCredential(cred).WithProfile(clientProfile)
}

// SendOctetStream Invoke API with application/octet-stream content-type.
//
// Note:
//  1. only specific API can be invoked in such manner.
//  2. only TC3-HMAC-SHA256 signature method can be specified.
//  3. only POST request method can be specified
//  4. the request Must be a CommonRequest and called SetOctetStreamParameters
func (c *Client) SendOctetStream(request tchttp.Request, response tchttp.Response) (err error) {
	if c.profile.SignMethod != "TC3-HMAC-SHA256" {
		return tcerr.NewTencentCloudSDKError("ClientError", "Invalid signature method.", "")
	}
	if c.profile.HttpProfile.ReqMethod != "POST" {
		return tcerr.NewTencentCloudSDKError("ClientError", "Invalid request method.", "")
	}
	//cr, ok := request.(*tchttp.CommonRequest)
	//if !ok {
	//	return tcerr.NewTencentCloudSDKError("ClientError", "Invalid request, must be *CommonRequest!", "")
	//}
	//if !cr.IsOctetStream() {
	//	return tcerr.NewTencentCloudSDKError("ClientError", "Invalid request, does not meet the conditions for sending OctetStream", "")
	//}
	return c.Send(request, response)
}
