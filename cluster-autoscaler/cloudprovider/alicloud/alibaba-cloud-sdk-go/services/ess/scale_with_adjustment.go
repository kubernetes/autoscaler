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

// ScaleWithAdjustment invokes the ess.ScaleWithAdjustment API synchronously
func (client *Client) ScaleWithAdjustment(request *ScaleWithAdjustmentRequest) (response *ScaleWithAdjustmentResponse, err error) {
	response = CreateScaleWithAdjustmentResponse()
	err = client.DoAction(request, response)
	return
}

// ScaleWithAdjustmentWithChan invokes the ess.ScaleWithAdjustment API asynchronously
func (client *Client) ScaleWithAdjustmentWithChan(request *ScaleWithAdjustmentRequest) (<-chan *ScaleWithAdjustmentResponse, <-chan error) {
	responseChan := make(chan *ScaleWithAdjustmentResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ScaleWithAdjustment(request)
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

// ScaleWithAdjustmentWithCallback invokes the ess.ScaleWithAdjustment API asynchronously
func (client *Client) ScaleWithAdjustmentWithCallback(request *ScaleWithAdjustmentRequest, callback func(response *ScaleWithAdjustmentResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ScaleWithAdjustmentResponse
		var err error
		defer close(result)
		response, err = client.ScaleWithAdjustment(request)
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

// ScaleWithAdjustmentRequest is the request struct for api ScaleWithAdjustment
type ScaleWithAdjustmentRequest struct {
	*requests.RpcRequest
	ClientToken            string                                  `position:"Query" name:"ClientToken"`
	ScalingGroupId         string                                  `position:"Query" name:"ScalingGroupId"`
	LifecycleHookContext   ScaleWithAdjustmentLifecycleHookContext `position:"Query" name:"LifecycleHookContext"  type:"Struct"`
	InstanceType           *[]string                               `position:"Query" name:"InstanceType"  type:"Repeated"`
	SyncActivity           requests.Boolean                        `position:"Query" name:"SyncActivity"`
	Allocation             *[]ScaleWithAdjustmentAllocation        `position:"Query" name:"Allocation"  type:"Repeated"`
	AdjustmentValue        requests.Integer                        `position:"Query" name:"AdjustmentValue"`
	ResourceOwnerAccount   string                                  `position:"Query" name:"ResourceOwnerAccount"`
	ActivityMetadata       string                                  `position:"Query" name:"ActivityMetadata"`
	AdjustmentType         string                                  `position:"Query" name:"AdjustmentType"`
	ParallelTask           requests.Boolean                        `position:"Query" name:"ParallelTask"`
	Overrides              ScaleWithAdjustmentOverrides            `position:"Query" name:"Overrides"  type:"Struct"`
	OwnerId                requests.Integer                        `position:"Query" name:"OwnerId"`
	SpotStrategy           string                                  `position:"Query" name:"SpotStrategy"`
	VSwitchId              *[]string                               `position:"Query" name:"VSwitchId"  type:"Repeated"`
	MinAdjustmentMagnitude requests.Integer                        `position:"Query" name:"MinAdjustmentMagnitude"`
}

// ScaleWithAdjustmentLifecycleHookContext is a repeated param struct in ScaleWithAdjustmentRequest
type ScaleWithAdjustmentLifecycleHookContext struct {
	DisableLifecycleHook    string    `name:"DisableLifecycleHook"`
	IgnoredLifecycleHookIds *[]string `name:"IgnoredLifecycleHookIds" type:"Repeated"`
}

// ScaleWithAdjustmentAllocation is a repeated param struct in ScaleWithAdjustmentRequest
type ScaleWithAdjustmentAllocation struct {
	VSwitchId *[]string `name:"VSwitchId" type:"Repeated"`
	Count     string    `name:"Count"`
}

// ScaleWithAdjustmentOverrides is a repeated param struct in ScaleWithAdjustmentRequest
type ScaleWithAdjustmentOverrides struct {
	Memory            string                                               `name:"Memory"`
	ContainerOverride *[]ScaleWithAdjustmentOverridesContainerOverrideItem `name:"ContainerOverride" type:"Repeated"`
	Cpu               string                                               `name:"Cpu"`
}

// ScaleWithAdjustmentOverridesContainerOverrideItem is a repeated param struct in ScaleWithAdjustmentRequest
type ScaleWithAdjustmentOverridesContainerOverrideItem struct {
	Memory         string                                                                 `name:"Memory"`
	Arg            *[]string                                                              `name:"Arg" type:"Repeated"`
	EnvironmentVar *[]ScaleWithAdjustmentOverridesContainerOverrideItemEnvironmentVarItem `name:"EnvironmentVar" type:"Repeated"`
	Name           string                                                                 `name:"Name"`
	Cpu            string                                                                 `name:"Cpu"`
	Command        *[]string                                                              `name:"Command" type:"Repeated"`
}

// ScaleWithAdjustmentOverridesContainerOverrideItemEnvironmentVarItem is a repeated param struct in ScaleWithAdjustmentRequest
type ScaleWithAdjustmentOverridesContainerOverrideItemEnvironmentVarItem struct {
	Value string `name:"Value"`
	Key   string `name:"Key"`
}

// ScaleWithAdjustmentResponse is the response struct for api ScaleWithAdjustment
type ScaleWithAdjustmentResponse struct {
	*responses.BaseResponse
	ScalingActivityId string `json:"ScalingActivityId" xml:"ScalingActivityId"`
	RequestId         string `json:"RequestId" xml:"RequestId"`
	ActivityType      string `json:"ActivityType" xml:"ActivityType"`
}

// CreateScaleWithAdjustmentRequest creates a request to invoke ScaleWithAdjustment API
func CreateScaleWithAdjustmentRequest() (request *ScaleWithAdjustmentRequest) {
	request = &ScaleWithAdjustmentRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ess", "2014-08-28", "ScaleWithAdjustment", "ess", "openAPI")
	request.Method = requests.POST
	return
}

// CreateScaleWithAdjustmentResponse creates a response to parse from ScaleWithAdjustment response
func CreateScaleWithAdjustmentResponse() (response *ScaleWithAdjustmentResponse) {
	response = &ScaleWithAdjustmentResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
