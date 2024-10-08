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

package ess

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/requests"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/responses"
)

// DescribeScalingGroups invokes the ess.DescribeScalingGroups API synchronously
func (client *Client) DescribeScalingGroups(request *DescribeScalingGroupsRequest) (response *DescribeScalingGroupsResponse, err error) {
	response = CreateDescribeScalingGroupsResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeScalingGroupsWithChan invokes the ess.DescribeScalingGroups API asynchronously
func (client *Client) DescribeScalingGroupsWithChan(request *DescribeScalingGroupsRequest) (<-chan *DescribeScalingGroupsResponse, <-chan error) {
	responseChan := make(chan *DescribeScalingGroupsResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeScalingGroups(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DescribeScalingGroupsWithCallback invokes the ess.DescribeScalingGroups API asynchronously
func (client *Client) DescribeScalingGroupsWithCallback(request *DescribeScalingGroupsRequest, callback func(response *DescribeScalingGroupsResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeScalingGroupsResponse
		var err error
		defer close(result)
		response, err = client.DescribeScalingGroups(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DescribeScalingGroupsRequest is the request struct for api DescribeScalingGroups
type DescribeScalingGroupsRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ResourceGroupId      string           `position:"Query" name:"ResourceGroupId"`
	GroupType            string           `position:"Query" name:"GroupType"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	ScalingGroupId       *[]string        `position:"Query" name:"ScalingGroupId"  type:"Repeated"`
	PageNumber           requests.Integer `position:"Query" name:"PageNumber"`
	PageSize             requests.Integer `position:"Query" name:"PageSize"`
	ScalingGroupName20   string           `position:"Query" name:"ScalingGroupName.20"`
	ScalingGroupName19   string           `position:"Query" name:"ScalingGroupName.19"`
	ScalingGroupName18   string           `position:"Query" name:"ScalingGroupName.18"`
	ScalingGroupName17   string           `position:"Query" name:"ScalingGroupName.17"`
	ScalingGroupName16   string           `position:"Query" name:"ScalingGroupName.16"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	ScalingGroupName     string           `position:"Query" name:"ScalingGroupName"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	ScalingGroupName1    string           `position:"Query" name:"ScalingGroupName.1"`
	ScalingGroupName2    string           `position:"Query" name:"ScalingGroupName.2"`
	ScalingGroupName7    string           `position:"Query" name:"ScalingGroupName.7"`
	ScalingGroupName11   string           `position:"Query" name:"ScalingGroupName.11"`
	ScalingGroupName8    string           `position:"Query" name:"ScalingGroupName.8"`
	ScalingGroupName10   string           `position:"Query" name:"ScalingGroupName.10"`
	ScalingGroupName9    string           `position:"Query" name:"ScalingGroupName.9"`
	ScalingGroupName3    string           `position:"Query" name:"ScalingGroupName.3"`
	ScalingGroupName15   string           `position:"Query" name:"ScalingGroupName.15"`
	ScalingGroupName4    string           `position:"Query" name:"ScalingGroupName.4"`
	ScalingGroupName14   string           `position:"Query" name:"ScalingGroupName.14"`
	ScalingGroupName5    string           `position:"Query" name:"ScalingGroupName.5"`
	ScalingGroupName13   string           `position:"Query" name:"ScalingGroupName.13"`
	ScalingGroupName6    string           `position:"Query" name:"ScalingGroupName.6"`
	ScalingGroupName12   string           `position:"Query" name:"ScalingGroupName.12"`
}

// DescribeScalingGroupsResponse is the response struct for api DescribeScalingGroups
type DescribeScalingGroupsResponse struct {
	*responses.BaseResponse
	RequestId     string        `json:"RequestId" xml:"RequestId"`
	PageNumber    int           `json:"PageNumber" xml:"PageNumber"`
	PageSize      int           `json:"PageSize" xml:"PageSize"`
	TotalCount    int           `json:"TotalCount" xml:"TotalCount"`
	ScalingGroups ScalingGroups `json:"ScalingGroups" xml:"ScalingGroups"`
}

// CreateDescribeScalingGroupsRequest creates a request to invoke DescribeScalingGroups API
func CreateDescribeScalingGroupsRequest() (request *DescribeScalingGroupsRequest) {
	request = &DescribeScalingGroupsRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ess", "2014-08-28", "DescribeScalingGroups", "ess", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDescribeScalingGroupsResponse creates a response to parse from DescribeScalingGroups response
func CreateDescribeScalingGroupsResponse() (response *DescribeScalingGroupsResponse) {
	response = &DescribeScalingGroupsResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
