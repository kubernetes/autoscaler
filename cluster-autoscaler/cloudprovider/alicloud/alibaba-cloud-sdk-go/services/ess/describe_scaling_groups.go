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
// api document: https://help.aliyun.com/api/ess/describescalinggroups.html
func (client *Client) DescribeScalingGroups(request *DescribeScalingGroupsRequest) (response *DescribeScalingGroupsResponse, err error) {
	response = CreateDescribeScalingGroupsResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeScalingGroupsWithChan invokes the ess.DescribeScalingGroups API asynchronously
// api document: https://help.aliyun.com/api/ess/describescalinggroups.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
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
// api document: https://help.aliyun.com/api/ess/describescalinggroups.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
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
	ScalingGroupId10     string           `position:"Query" name:"ScalingGroupId.10"`
	ScalingGroupId12     string           `position:"Query" name:"ScalingGroupId.12"`
	ScalingGroupId13     string           `position:"Query" name:"ScalingGroupId.13"`
	ScalingGroupId14     string           `position:"Query" name:"ScalingGroupId.14"`
	ScalingGroupId15     string           `position:"Query" name:"ScalingGroupId.15"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	PageNumber           requests.Integer `position:"Query" name:"PageNumber"`
	PageSize             requests.Integer `position:"Query" name:"PageSize"`
	ScalingGroupName20   string           `position:"Query" name:"ScalingGroupName.20"`
	ScalingGroupName19   string           `position:"Query" name:"ScalingGroupName.19"`
	ScalingGroupId20     string           `position:"Query" name:"ScalingGroupId.20"`
	ScalingGroupName18   string           `position:"Query" name:"ScalingGroupName.18"`
	ScalingGroupName17   string           `position:"Query" name:"ScalingGroupName.17"`
	ScalingGroupName16   string           `position:"Query" name:"ScalingGroupName.16"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	ScalingGroupName     string           `position:"Query" name:"ScalingGroupName"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	ScalingGroupName1    string           `position:"Query" name:"ScalingGroupName.1"`
	ScalingGroupName2    string           `position:"Query" name:"ScalingGroupName.2"`
	ScalingGroupId2      string           `position:"Query" name:"ScalingGroupId.2"`
	ScalingGroupId1      string           `position:"Query" name:"ScalingGroupId.1"`
	ScalingGroupId6      string           `position:"Query" name:"ScalingGroupId.6"`
	ScalingGroupId16     string           `position:"Query" name:"ScalingGroupId.16"`
	ScalingGroupName7    string           `position:"Query" name:"ScalingGroupName.7"`
	ScalingGroupName11   string           `position:"Query" name:"ScalingGroupName.11"`
	ScalingGroupId5      string           `position:"Query" name:"ScalingGroupId.5"`
	ScalingGroupId17     string           `position:"Query" name:"ScalingGroupId.17"`
	ScalingGroupName8    string           `position:"Query" name:"ScalingGroupName.8"`
	ScalingGroupName10   string           `position:"Query" name:"ScalingGroupName.10"`
	ScalingGroupId4      string           `position:"Query" name:"ScalingGroupId.4"`
	ScalingGroupId18     string           `position:"Query" name:"ScalingGroupId.18"`
	ScalingGroupName9    string           `position:"Query" name:"ScalingGroupName.9"`
	ScalingGroupId3      string           `position:"Query" name:"ScalingGroupId.3"`
	ScalingGroupId19     string           `position:"Query" name:"ScalingGroupId.19"`
	ScalingGroupName3    string           `position:"Query" name:"ScalingGroupName.3"`
	ScalingGroupName15   string           `position:"Query" name:"ScalingGroupName.15"`
	ScalingGroupId9      string           `position:"Query" name:"ScalingGroupId.9"`
	ScalingGroupName4    string           `position:"Query" name:"ScalingGroupName.4"`
	ScalingGroupName14   string           `position:"Query" name:"ScalingGroupName.14"`
	ScalingGroupId8      string           `position:"Query" name:"ScalingGroupId.8"`
	ScalingGroupName5    string           `position:"Query" name:"ScalingGroupName.5"`
	ScalingGroupName13   string           `position:"Query" name:"ScalingGroupName.13"`
	ScalingGroupId7      string           `position:"Query" name:"ScalingGroupId.7"`
	ScalingGroupName6    string           `position:"Query" name:"ScalingGroupName.6"`
	ScalingGroupName12   string           `position:"Query" name:"ScalingGroupName.12"`
}

// DescribeScalingGroupsResponse is the response struct for api DescribeScalingGroups
type DescribeScalingGroupsResponse struct {
	*responses.BaseResponse
	TotalCount    int           `json:"TotalCount" xml:"TotalCount"`
	PageNumber    int           `json:"PageNumber" xml:"PageNumber"`
	PageSize      int           `json:"PageSize" xml:"PageSize"`
	RequestId     string        `json:"RequestId" xml:"RequestId"`
	ScalingGroups ScalingGroups `json:"ScalingGroups" xml:"ScalingGroups"`
}

// CreateDescribeScalingGroupsRequest creates a request to invoke DescribeScalingGroups API
func CreateDescribeScalingGroupsRequest() (request *DescribeScalingGroupsRequest) {
	request = &DescribeScalingGroupsRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ess", "2014-08-28", "DescribeScalingGroups", "ess", "openAPI")
	return
}

// CreateDescribeScalingGroupsResponse creates a response to parse from DescribeScalingGroups response
func CreateDescribeScalingGroupsResponse() (response *DescribeScalingGroupsResponse) {
	response = &DescribeScalingGroupsResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

// ScalingGroups is a nested struct in ess response
type ScalingGroups struct {
	ScalingGroup []ScalingGroup `json:"ScalingGroup" xml:"ScalingGroup"`
}

// ScalingGroup is a nested struct in ess response
type ScalingGroup struct {
	DefaultCooldown              int             `json:"DefaultCooldown" xml:"DefaultCooldown"`
	MaxSize                      int             `json:"MaxSize" xml:"MaxSize"`
	PendingWaitCapacity          int             `json:"PendingWaitCapacity" xml:"PendingWaitCapacity"`
	RemovingWaitCapacity         int             `json:"RemovingWaitCapacity" xml:"RemovingWaitCapacity"`
	PendingCapacity              int             `json:"PendingCapacity" xml:"PendingCapacity"`
	RemovingCapacity             int             `json:"RemovingCapacity" xml:"RemovingCapacity"`
	ScalingGroupName             string          `json:"ScalingGroupName" xml:"ScalingGroupName"`
	ActiveCapacity               int             `json:"ActiveCapacity" xml:"ActiveCapacity"`
	StandbyCapacity              int             `json:"StandbyCapacity" xml:"StandbyCapacity"`
	ProtectedCapacity            int             `json:"ProtectedCapacity" xml:"ProtectedCapacity"`
	ActiveScalingConfigurationId string          `json:"ActiveScalingConfigurationId" xml:"ActiveScalingConfigurationId"`
	LaunchTemplateId             string          `json:"LaunchTemplateId" xml:"LaunchTemplateId"`
	LaunchTemplateVersion        string          `json:"LaunchTemplateVersion" xml:"LaunchTemplateVersion"`
	ScalingGroupId               string          `json:"ScalingGroupId" xml:"ScalingGroupId"`
	RegionId                     string          `json:"RegionId" xml:"RegionId"`
	TotalCapacity                int             `json:"TotalCapacity" xml:"TotalCapacity"`
	MinSize                      int             `json:"MinSize" xml:"MinSize"`
	LifecycleState               string          `json:"LifecycleState" xml:"LifecycleState"`
	CreationTime                 string          `json:"CreationTime" xml:"CreationTime"`
	ModificationTime             string          `json:"ModificationTime" xml:"ModificationTime"`
	VpcId                        string          `json:"VpcId" xml:"VpcId"`
	VSwitchId                    string          `json:"VSwitchId" xml:"VSwitchId"`
	MultiAZPolicy                string          `json:"MultiAZPolicy" xml:"MultiAZPolicy"`
	HealthCheckType              string          `json:"HealthCheckType" xml:"HealthCheckType"`
	VSwitchIds                   VSwitchIds      `json:"VSwitchIds" xml:"VSwitchIds"`
	RemovalPolicies              RemovalPolicies `json:"RemovalPolicies" xml:"RemovalPolicies"`
	DBInstanceIds                DBInstanceIds   `json:"DBInstanceIds" xml:"DBInstanceIds"`
	LoadBalancerIds              LoadBalancerIds `json:"LoadBalancerIds" xml:"LoadBalancerIds"`
}

// VSwitchIds is a nested struct in ess response
type VSwitchIds struct {
	VSwitchId []string `json:"VSwitchId" xml:"VSwitchId"`
}

// RemovalPolicies is a nested struct in ess response
type RemovalPolicies struct {
	RemovalPolicy []string `json:"RemovalPolicy" xml:"RemovalPolicy"`
}

// DBInstanceIds is a nested struct in ess response
type DBInstanceIds struct {
	DBInstanceId []string `json:"DBInstanceId" xml:"DBInstanceId"`
}

// LoadBalancerIds is a nested struct in ess response
type LoadBalancerIds struct {
	LoadBalancerId []string `json:"LoadBalancerId" xml:"LoadBalancerId"`
}
