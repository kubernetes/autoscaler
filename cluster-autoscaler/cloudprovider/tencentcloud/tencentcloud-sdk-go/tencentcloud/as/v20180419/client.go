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

package v20180419

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = "2018-04-19"

type Client struct {
	common.Client
}

// Deprecated
func NewClientWithSecretId(secretId, secretKey, region string) (client *Client, err error) {
	cpf := profile.NewClientProfile()
	client = &Client{}
	client.Init(region).WithSecretId(secretId, secretKey).WithProfile(cpf)
	return
}

func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func NewAttachInstancesRequest() (request *AttachInstancesRequest) {
	request = &AttachInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "AttachInstances")

	return
}

func NewAttachInstancesResponse() (response *AttachInstancesResponse) {
	response = &AttachInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AttachInstances
// 本接口（AttachInstances）用于将 CVM 实例添加到伸缩组。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPABOVEMAXSIZE = "ResourceInsufficient.AutoScalingGroupAboveMaxSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEABOVEMAXSIZE = "ResourceInsufficient.InServiceInstanceAboveMaxSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTFOUND = "ResourceNotFound.InstancesNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_CVMVPCINCONSISTENT = "ResourceUnavailable.CvmVpcInconsistent"
//	RESOURCEUNAVAILABLE_INSTANCESALREADYINAUTOSCALINGGROUP = "ResourceUnavailable.InstancesAlreadyInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) AttachInstances(request *AttachInstancesRequest) (response *AttachInstancesResponse, err error) {
	if request == nil {
		request = NewAttachInstancesRequest()
	}

	response = NewAttachInstancesResponse()
	err = c.Send(request, response)
	return
}

// AttachInstances
// 本接口（AttachInstances）用于将 CVM 实例添加到伸缩组。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPABOVEMAXSIZE = "ResourceInsufficient.AutoScalingGroupAboveMaxSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEABOVEMAXSIZE = "ResourceInsufficient.InServiceInstanceAboveMaxSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTFOUND = "ResourceNotFound.InstancesNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_CVMVPCINCONSISTENT = "ResourceUnavailable.CvmVpcInconsistent"
//	RESOURCEUNAVAILABLE_INSTANCESALREADYINAUTOSCALINGGROUP = "ResourceUnavailable.InstancesAlreadyInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) AttachInstancesWithContext(ctx context.Context, request *AttachInstancesRequest) (response *AttachInstancesResponse, err error) {
	if request == nil {
		request = NewAttachInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewAttachInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewAttachLoadBalancersRequest() (request *AttachLoadBalancersRequest) {
	request = &AttachLoadBalancersRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "AttachLoadBalancers")

	return
}

func NewAttachLoadBalancersResponse() (response *AttachLoadBalancersResponse) {
	response = &AttachLoadBalancersResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AttachLoadBalancers
// 此接口（AttachLoadBalancers）用于将负载均衡器添加到伸缩组。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	INVALIDPARAMETERVALUE_TARGETPORTDUPLICATED = "InvalidParameterValue.TargetPortDuplicated"
//	LIMITEXCEEDED_AFTERATTACHLBLIMITEXCEEDED = "LimitExceeded.AfterAttachLbLimitExceeded"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBPROJECTINCONSISTENT = "ResourceUnavailable.LbProjectInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) AttachLoadBalancers(request *AttachLoadBalancersRequest) (response *AttachLoadBalancersResponse, err error) {
	if request == nil {
		request = NewAttachLoadBalancersRequest()
	}

	response = NewAttachLoadBalancersResponse()
	err = c.Send(request, response)
	return
}

// AttachLoadBalancers
// 此接口（AttachLoadBalancers）用于将负载均衡器添加到伸缩组。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	INVALIDPARAMETERVALUE_TARGETPORTDUPLICATED = "InvalidParameterValue.TargetPortDuplicated"
//	LIMITEXCEEDED_AFTERATTACHLBLIMITEXCEEDED = "LimitExceeded.AfterAttachLbLimitExceeded"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBPROJECTINCONSISTENT = "ResourceUnavailable.LbProjectInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) AttachLoadBalancersWithContext(ctx context.Context, request *AttachLoadBalancersRequest) (response *AttachLoadBalancersResponse, err error) {
	if request == nil {
		request = NewAttachLoadBalancersRequest()
	}
	request.SetContext(ctx)

	response = NewAttachLoadBalancersResponse()
	err = c.Send(request, response)
	return
}

func NewClearLaunchConfigurationAttributesRequest() (request *ClearLaunchConfigurationAttributesRequest) {
	request = &ClearLaunchConfigurationAttributesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ClearLaunchConfigurationAttributes")

	return
}

func NewClearLaunchConfigurationAttributesResponse() (response *ClearLaunchConfigurationAttributesResponse) {
	response = &ClearLaunchConfigurationAttributesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ClearLaunchConfigurationAttributes
// 本接口（ClearLaunchConfigurationAttributes）用于将启动配置内的特定属性完全清空。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
func (c *Client) ClearLaunchConfigurationAttributes(request *ClearLaunchConfigurationAttributesRequest) (response *ClearLaunchConfigurationAttributesResponse, err error) {
	if request == nil {
		request = NewClearLaunchConfigurationAttributesRequest()
	}

	response = NewClearLaunchConfigurationAttributesResponse()
	err = c.Send(request, response)
	return
}

// ClearLaunchConfigurationAttributes
// 本接口（ClearLaunchConfigurationAttributes）用于将启动配置内的特定属性完全清空。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
func (c *Client) ClearLaunchConfigurationAttributesWithContext(ctx context.Context, request *ClearLaunchConfigurationAttributesRequest) (response *ClearLaunchConfigurationAttributesResponse, err error) {
	if request == nil {
		request = NewClearLaunchConfigurationAttributesRequest()
	}
	request.SetContext(ctx)

	response = NewClearLaunchConfigurationAttributesResponse()
	err = c.Send(request, response)
	return
}

func NewCompleteLifecycleActionRequest() (request *CompleteLifecycleActionRequest) {
	request = &CompleteLifecycleActionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "CompleteLifecycleAction")

	return
}

func NewCompleteLifecycleActionResponse() (response *CompleteLifecycleActionResponse) {
	response = &CompleteLifecycleActionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CompleteLifecycleAction
// 本接口（CompleteLifecycleAction）用于完成生命周期动作。
//
// * 用户通过调用本接口，指定一个具体的生命周期挂钩的结果（“CONITNUE”或者“ABANDON”）。如果一直不调用本接口，则生命周期挂钩会在超时后按照“DefaultResult”进行处理。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"
//	RESOURCENOTFOUND_LIFECYCLEHOOKINSTANCENOTFOUND = "ResourceNotFound.LifecycleHookInstanceNotFound"
//	RESOURCENOTFOUND_LIFECYCLEHOOKNOTFOUND = "ResourceNotFound.LifecycleHookNotFound"
//	RESOURCEUNAVAILABLE_LIFECYCLEACTIONRESULTHASSET = "ResourceUnavailable.LifecycleActionResultHasSet"
func (c *Client) CompleteLifecycleAction(request *CompleteLifecycleActionRequest) (response *CompleteLifecycleActionResponse, err error) {
	if request == nil {
		request = NewCompleteLifecycleActionRequest()
	}

	response = NewCompleteLifecycleActionResponse()
	err = c.Send(request, response)
	return
}

// CompleteLifecycleAction
// 本接口（CompleteLifecycleAction）用于完成生命周期动作。
//
// * 用户通过调用本接口，指定一个具体的生命周期挂钩的结果（“CONITNUE”或者“ABANDON”）。如果一直不调用本接口，则生命周期挂钩会在超时后按照“DefaultResult”进行处理。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"
//	RESOURCENOTFOUND_LIFECYCLEHOOKINSTANCENOTFOUND = "ResourceNotFound.LifecycleHookInstanceNotFound"
//	RESOURCENOTFOUND_LIFECYCLEHOOKNOTFOUND = "ResourceNotFound.LifecycleHookNotFound"
//	RESOURCEUNAVAILABLE_LIFECYCLEACTIONRESULTHASSET = "ResourceUnavailable.LifecycleActionResultHasSet"
func (c *Client) CompleteLifecycleActionWithContext(ctx context.Context, request *CompleteLifecycleActionRequest) (response *CompleteLifecycleActionResponse, err error) {
	if request == nil {
		request = NewCompleteLifecycleActionRequest()
	}
	request.SetContext(ctx)

	response = NewCompleteLifecycleActionResponse()
	err = c.Send(request, response)
	return
}

func NewCreateAutoScalingGroupRequest() (request *CreateAutoScalingGroupRequest) {
	request = &CreateAutoScalingGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "CreateAutoScalingGroup")

	return
}

func NewCreateAutoScalingGroupResponse() (response *CreateAutoScalingGroupResponse) {
	response = &CreateAutoScalingGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateAutoScalingGroup
// 本接口（CreateAutoScalingGroup）用于创建伸缩组
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLLBERROR = "InternalError.CallLbError"
//	INTERNALERROR_CALLTAGERROR = "InternalError.CallTagError"
//	INTERNALERROR_CALLTVPCERROR = "InternalError.CallTvpcError"
//	INTERNALERROR_CALLVPCERROR = "InternalError.CallVpcError"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETERVALUE_BASECAPACITYTOOLARGE = "InvalidParameterValue.BaseCapacityTooLarge"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"
//	INVALIDPARAMETERVALUE_DUPLICATEDFORWARDLB = "InvalidParameterValue.DuplicatedForwardLb"
//	INVALIDPARAMETERVALUE_DUPLICATEDSUBNET = "InvalidParameterValue.DuplicatedSubnet"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_GROUPNAMEDUPLICATED = "InvalidParameterValue.GroupNameDuplicated"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_INVALIDSUBNETID = "InvalidParameterValue.InvalidSubnetId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNOTFOUND = "InvalidParameterValue.LaunchConfigurationNotFound"
//	INVALIDPARAMETERVALUE_LBPROJECTINCONSISTENT = "InvalidParameterValue.LbProjectInconsistent"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ONLYVPC = "InvalidParameterValue.OnlyVpc"
//	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_SUBNETIDS = "InvalidParameterValue.SubnetIds"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_ZONEMISMATCHREGION = "InvalidParameterValue.ZoneMismatchRegion"
//	LIMITEXCEEDED = "LimitExceeded"
//	LIMITEXCEEDED_AUTOSCALINGGROUPLIMITEXCEEDED = "LimitExceeded.AutoScalingGroupLimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_LAUNCHCONFIGURATIONSTATUSABNORMAL = "ResourceUnavailable.LaunchConfigurationStatusAbnormal"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_PROJECTINCONSISTENT = "ResourceUnavailable.ProjectInconsistent"
//	RESOURCEUNAVAILABLE_ZONEUNAVAILABLE = "ResourceUnavailable.ZoneUnavailable"
func (c *Client) CreateAutoScalingGroup(request *CreateAutoScalingGroupRequest) (response *CreateAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewCreateAutoScalingGroupRequest()
	}

	response = NewCreateAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

// CreateAutoScalingGroup
// 本接口（CreateAutoScalingGroup）用于创建伸缩组
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLLBERROR = "InternalError.CallLbError"
//	INTERNALERROR_CALLTAGERROR = "InternalError.CallTagError"
//	INTERNALERROR_CALLTVPCERROR = "InternalError.CallTvpcError"
//	INTERNALERROR_CALLVPCERROR = "InternalError.CallVpcError"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETERVALUE_BASECAPACITYTOOLARGE = "InvalidParameterValue.BaseCapacityTooLarge"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"
//	INVALIDPARAMETERVALUE_DUPLICATEDFORWARDLB = "InvalidParameterValue.DuplicatedForwardLb"
//	INVALIDPARAMETERVALUE_DUPLICATEDSUBNET = "InvalidParameterValue.DuplicatedSubnet"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_GROUPNAMEDUPLICATED = "InvalidParameterValue.GroupNameDuplicated"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_INVALIDSUBNETID = "InvalidParameterValue.InvalidSubnetId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNOTFOUND = "InvalidParameterValue.LaunchConfigurationNotFound"
//	INVALIDPARAMETERVALUE_LBPROJECTINCONSISTENT = "InvalidParameterValue.LbProjectInconsistent"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ONLYVPC = "InvalidParameterValue.OnlyVpc"
//	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_SUBNETIDS = "InvalidParameterValue.SubnetIds"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_ZONEMISMATCHREGION = "InvalidParameterValue.ZoneMismatchRegion"
//	LIMITEXCEEDED = "LimitExceeded"
//	LIMITEXCEEDED_AUTOSCALINGGROUPLIMITEXCEEDED = "LimitExceeded.AutoScalingGroupLimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_LAUNCHCONFIGURATIONSTATUSABNORMAL = "ResourceUnavailable.LaunchConfigurationStatusAbnormal"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_PROJECTINCONSISTENT = "ResourceUnavailable.ProjectInconsistent"
//	RESOURCEUNAVAILABLE_ZONEUNAVAILABLE = "ResourceUnavailable.ZoneUnavailable"
func (c *Client) CreateAutoScalingGroupWithContext(ctx context.Context, request *CreateAutoScalingGroupRequest) (response *CreateAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewCreateAutoScalingGroupRequest()
	}
	request.SetContext(ctx)

	response = NewCreateAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

func NewCreateAutoScalingGroupFromInstanceRequest() (request *CreateAutoScalingGroupFromInstanceRequest) {
	request = &CreateAutoScalingGroupFromInstanceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "CreateAutoScalingGroupFromInstance")

	return
}

func NewCreateAutoScalingGroupFromInstanceResponse() (response *CreateAutoScalingGroupFromInstanceResponse) {
	response = &CreateAutoScalingGroupFromInstanceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateAutoScalingGroupFromInstance
// 本接口（CreateAutoScalingGroupFromInstance）用于根据实例创建启动配置及伸缩组。
//
// 说明：根据按包年包月计费的实例所创建的伸缩组，其扩容的实例为按量计费实例。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	CALLCVMERROR = "CallCvmError"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLVPCERROR = "InternalError.CallVpcError"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"
//	INVALIDPARAMETERVALUE_DUPLICATEDSUBNET = "InvalidParameterValue.DuplicatedSubnet"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED_AUTOSCALINGGROUPLIMITEXCEEDED = "LimitExceeded.AutoScalingGroupLimitExceeded"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_INSTANCESNOTFOUND = "ResourceNotFound.InstancesNotFound"
//	RESOURCEUNAVAILABLE_LAUNCHCONFIGURATIONSTATUSABNORMAL = "ResourceUnavailable.LaunchConfigurationStatusAbnormal"
//	RESOURCEUNAVAILABLE_PROJECTINCONSISTENT = "ResourceUnavailable.ProjectInconsistent"
//	RESOURCEUNAVAILABLE_STOPPEDINSTANCENOTALLOWATTACH = "ResourceUnavailable.StoppedInstanceNotAllowAttach"
func (c *Client) CreateAutoScalingGroupFromInstance(request *CreateAutoScalingGroupFromInstanceRequest) (response *CreateAutoScalingGroupFromInstanceResponse, err error) {
	if request == nil {
		request = NewCreateAutoScalingGroupFromInstanceRequest()
	}

	response = NewCreateAutoScalingGroupFromInstanceResponse()
	err = c.Send(request, response)
	return
}

// CreateAutoScalingGroupFromInstance
// 本接口（CreateAutoScalingGroupFromInstance）用于根据实例创建启动配置及伸缩组。
//
// 说明：根据按包年包月计费的实例所创建的伸缩组，其扩容的实例为按量计费实例。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	CALLCVMERROR = "CallCvmError"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLVPCERROR = "InternalError.CallVpcError"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"
//	INVALIDPARAMETERVALUE_DUPLICATEDSUBNET = "InvalidParameterValue.DuplicatedSubnet"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED_AUTOSCALINGGROUPLIMITEXCEEDED = "LimitExceeded.AutoScalingGroupLimitExceeded"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_INSTANCESNOTFOUND = "ResourceNotFound.InstancesNotFound"
//	RESOURCEUNAVAILABLE_LAUNCHCONFIGURATIONSTATUSABNORMAL = "ResourceUnavailable.LaunchConfigurationStatusAbnormal"
//	RESOURCEUNAVAILABLE_PROJECTINCONSISTENT = "ResourceUnavailable.ProjectInconsistent"
//	RESOURCEUNAVAILABLE_STOPPEDINSTANCENOTALLOWATTACH = "ResourceUnavailable.StoppedInstanceNotAllowAttach"
func (c *Client) CreateAutoScalingGroupFromInstanceWithContext(ctx context.Context, request *CreateAutoScalingGroupFromInstanceRequest) (response *CreateAutoScalingGroupFromInstanceResponse, err error) {
	if request == nil {
		request = NewCreateAutoScalingGroupFromInstanceRequest()
	}
	request.SetContext(ctx)

	response = NewCreateAutoScalingGroupFromInstanceResponse()
	err = c.Send(request, response)
	return
}

func NewCreateLaunchConfigurationRequest() (request *CreateLaunchConfigurationRequest) {
	request = &CreateLaunchConfigurationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "CreateLaunchConfiguration")

	return
}

func NewCreateLaunchConfigurationResponse() (response *CreateLaunchConfigurationResponse) {
	response = &CreateLaunchConfigurationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateLaunchConfiguration
// 本接口（CreateLaunchConfiguration）用于创建新的启动配置。
//
// * 启动配置，可以通过 `ModifyLaunchConfigurationAttributes` 修改少量字段。如需使用新的启动配置，建议重新创建启动配置。
//
// * 每个项目最多只能创建20个启动配置，详见[使用限制](https://cloud.tencent.com/document/product/377/3120)。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_HOSTNAMEUNAVAILABLE = "InvalidParameter.HostNameUnavailable"
//	INVALIDPARAMETER_INVALIDCOMBINATION = "InvalidParameter.InvalidCombination"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETER_PARAMETERMUSTBEDELETED = "InvalidParameter.ParameterMustBeDeleted"
//	INVALIDPARAMETERVALUE_CVMCONFIGURATIONERROR = "InvalidParameterValue.CvmConfigurationError"
//	INVALIDPARAMETERVALUE_HOSTNAMEILLEGAL = "InvalidParameterValue.HostNameIllegal"
//	INVALIDPARAMETERVALUE_INSTANCENAMEILLEGAL = "InvalidParameterValue.InstanceNameIllegal"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTED = "InvalidParameterValue.InstanceTypeNotSupported"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCETYPE = "InvalidParameterValue.InvalidInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATION = "InvalidParameterValue.InvalidLaunchConfiguration"
//	INVALIDPARAMETERVALUE_INVALIDSNAPSHOTID = "InvalidParameterValue.InvalidSnapshotId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_NOTSTRINGTYPEFLOAT = "InvalidParameterValue.NotStringTypeFloat"
//	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_TOOSHORT = "InvalidParameterValue.TooShort"
//	INVALIDPARAMETERVALUE_USERDATAFORMATERROR = "InvalidParameterValue.UserDataFormatError"
//	INVALIDPARAMETERVALUE_USERDATASIZEEXCEEDED = "InvalidParameterValue.UserDataSizeExceeded"
//	INVALIDPERMISSION = "InvalidPermission"
//	LIMITEXCEEDED_LAUNCHCONFIGURATIONQUOTANOTENOUGH = "LimitExceeded.LaunchConfigurationQuotaNotEnough"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_INSTANCEMARKETOPTIONS = "MissingParameter.InstanceMarketOptions"
func (c *Client) CreateLaunchConfiguration(request *CreateLaunchConfigurationRequest) (response *CreateLaunchConfigurationResponse, err error) {
	if request == nil {
		request = NewCreateLaunchConfigurationRequest()
	}

	response = NewCreateLaunchConfigurationResponse()
	err = c.Send(request, response)
	return
}

// CreateLaunchConfiguration
// 本接口（CreateLaunchConfiguration）用于创建新的启动配置。
//
// * 启动配置，可以通过 `ModifyLaunchConfigurationAttributes` 修改少量字段。如需使用新的启动配置，建议重新创建启动配置。
//
// * 每个项目最多只能创建20个启动配置，详见[使用限制](https://cloud.tencent.com/document/product/377/3120)。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_HOSTNAMEUNAVAILABLE = "InvalidParameter.HostNameUnavailable"
//	INVALIDPARAMETER_INVALIDCOMBINATION = "InvalidParameter.InvalidCombination"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETER_PARAMETERMUSTBEDELETED = "InvalidParameter.ParameterMustBeDeleted"
//	INVALIDPARAMETERVALUE_CVMCONFIGURATIONERROR = "InvalidParameterValue.CvmConfigurationError"
//	INVALIDPARAMETERVALUE_HOSTNAMEILLEGAL = "InvalidParameterValue.HostNameIllegal"
//	INVALIDPARAMETERVALUE_INSTANCENAMEILLEGAL = "InvalidParameterValue.InstanceNameIllegal"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTED = "InvalidParameterValue.InstanceTypeNotSupported"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCETYPE = "InvalidParameterValue.InvalidInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATION = "InvalidParameterValue.InvalidLaunchConfiguration"
//	INVALIDPARAMETERVALUE_INVALIDSNAPSHOTID = "InvalidParameterValue.InvalidSnapshotId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_NOTSTRINGTYPEFLOAT = "InvalidParameterValue.NotStringTypeFloat"
//	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_TOOSHORT = "InvalidParameterValue.TooShort"
//	INVALIDPARAMETERVALUE_USERDATAFORMATERROR = "InvalidParameterValue.UserDataFormatError"
//	INVALIDPARAMETERVALUE_USERDATASIZEEXCEEDED = "InvalidParameterValue.UserDataSizeExceeded"
//	INVALIDPERMISSION = "InvalidPermission"
//	LIMITEXCEEDED_LAUNCHCONFIGURATIONQUOTANOTENOUGH = "LimitExceeded.LaunchConfigurationQuotaNotEnough"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_INSTANCEMARKETOPTIONS = "MissingParameter.InstanceMarketOptions"
func (c *Client) CreateLaunchConfigurationWithContext(ctx context.Context, request *CreateLaunchConfigurationRequest) (response *CreateLaunchConfigurationResponse, err error) {
	if request == nil {
		request = NewCreateLaunchConfigurationRequest()
	}
	request.SetContext(ctx)

	response = NewCreateLaunchConfigurationResponse()
	err = c.Send(request, response)
	return
}

func NewCreateLifecycleHookRequest() (request *CreateLifecycleHookRequest) {
	request = &CreateLifecycleHookRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "CreateLifecycleHook")

	return
}

func NewCreateLifecycleHookResponse() (response *CreateLifecycleHookResponse) {
	response = &CreateLifecycleHookResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateLifecycleHook
// 本接口（CreateLifecycleHook）用于创建生命周期挂钩。
//
// * 您可以为生命周期挂钩配置消息通知，弹性伸缩会通知您的CMQ消息队列，通知内容形如：
//
// “`
//
// {
//
//	"Service": "Tencent Cloud Auto Scaling",
//
//	"Time": "2019-03-14T10:15:11Z",
//
//	"AppId": "1251783334",
//
//	"ActivityId": "asa-fznnvrja",
//
//	"AutoScalingGroupId": "asg-rrrrtttt",
//
//	"LifecycleHookId": "ash-xxxxyyyy",
//
//	"LifecycleHookName": "my-hook",
//
//	"LifecycleActionToken": "3080e1c9-0efe-4dd7-ad3b-90cd6618298f",
//
//	"InstanceId": "ins-aaaabbbb",
//
//	"LifecycleTransition": "INSTANCE_LAUNCHING",
//
//	"NotificationMetadata": ""
//
// }
//
// “`
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLEEERROR = "InternalError.CalleeError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_LIFECYCLEHOOKNAMEDUPLICATED = "InvalidParameterValue.LifecycleHookNameDuplicated"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_CMQQUEUENOTFOUND = "ResourceNotFound.CmqQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQQUEUENOTFOUND = "ResourceNotFound.TDMQCMQQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQTOPICNOTFOUND = "ResourceNotFound.TDMQCMQTopicNotFound"
//	RESOURCEUNAVAILABLE_CMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.CmqTopicHasNoSubscriber"
//	RESOURCEUNAVAILABLE_TDMQCMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.TDMQCMQTopicHasNoSubscriber"
func (c *Client) CreateLifecycleHook(request *CreateLifecycleHookRequest) (response *CreateLifecycleHookResponse, err error) {
	if request == nil {
		request = NewCreateLifecycleHookRequest()
	}

	response = NewCreateLifecycleHookResponse()
	err = c.Send(request, response)
	return
}

// CreateLifecycleHook
// 本接口（CreateLifecycleHook）用于创建生命周期挂钩。
//
// * 您可以为生命周期挂钩配置消息通知，弹性伸缩会通知您的CMQ消息队列，通知内容形如：
//
// “`
//
// {
//
//	"Service": "Tencent Cloud Auto Scaling",
//
//	"Time": "2019-03-14T10:15:11Z",
//
//	"AppId": "1251783334",
//
//	"ActivityId": "asa-fznnvrja",
//
//	"AutoScalingGroupId": "asg-rrrrtttt",
//
//	"LifecycleHookId": "ash-xxxxyyyy",
//
//	"LifecycleHookName": "my-hook",
//
//	"LifecycleActionToken": "3080e1c9-0efe-4dd7-ad3b-90cd6618298f",
//
//	"InstanceId": "ins-aaaabbbb",
//
//	"LifecycleTransition": "INSTANCE_LAUNCHING",
//
//	"NotificationMetadata": ""
//
// }
//
// “`
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLEEERROR = "InternalError.CalleeError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_LIFECYCLEHOOKNAMEDUPLICATED = "InvalidParameterValue.LifecycleHookNameDuplicated"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_CMQQUEUENOTFOUND = "ResourceNotFound.CmqQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQQUEUENOTFOUND = "ResourceNotFound.TDMQCMQQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQTOPICNOTFOUND = "ResourceNotFound.TDMQCMQTopicNotFound"
//	RESOURCEUNAVAILABLE_CMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.CmqTopicHasNoSubscriber"
//	RESOURCEUNAVAILABLE_TDMQCMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.TDMQCMQTopicHasNoSubscriber"
func (c *Client) CreateLifecycleHookWithContext(ctx context.Context, request *CreateLifecycleHookRequest) (response *CreateLifecycleHookResponse, err error) {
	if request == nil {
		request = NewCreateLifecycleHookRequest()
	}
	request.SetContext(ctx)

	response = NewCreateLifecycleHookResponse()
	err = c.Send(request, response)
	return
}

func NewCreateNotificationConfigurationRequest() (request *CreateNotificationConfigurationRequest) {
	request = &CreateNotificationConfigurationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "CreateNotificationConfiguration")

	return
}

func NewCreateNotificationConfigurationResponse() (response *CreateNotificationConfigurationResponse) {
	response = &CreateNotificationConfigurationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateNotificationConfiguration
// 本接口（CreateNotificationConfiguration）用于创建通知。
//
// 通知到 CMQ 主题或队列时，消息内容如下：
//
// “`
//
// {
//
//	"Service": "Tencent Cloud Auto Scaling",
//
//	"CreatedTime": "2021-10-11T10:15:11Z", // 活动创建时间
//
//	"AppId": "100000000",
//
//	"ActivityId": "asa-fznnvrja", // 伸缩活动ID
//
//	"AutoScalingGroupId": "asg-pc2oqu2z", // 伸缩组ID
//
//	"ActivityType": "SCALE_OUT",  // 伸缩活动类型
//
//	"StatusCode": "SUCCESSFUL",   // 伸缩活动结果
//
//	"Description": "Activity was launched in response to a difference between desired capacity and actual capacity,
//
//	scale out 1 instance(s).", // 伸缩活动描述
//
//	"StartTime": "2021-10-11T10:15:11Z",  // 活动开始时间
//
//	"EndTime": "2021-10-11T10:15:32Z",    // 活动结束时间
//
//	"DetailedStatusMessageSet": [ // 活动内部错误集合（非空不代表活动失败）
//
//	    {
//
//	        "Code": "InvalidInstanceType",
//
//	        "Zone": "ap-guangzhou-2",
//
//	        "InstanceId": "",
//
//	        "InstanceChargeType": "POSTPAID_BY_HOUR",
//
//	        "SubnetId": "subnet-4t5mgeuu",
//
//	        "Message": "The specified instance type `S5.LARGE8` is invalid in `subnet-4t5mgeuu`, `ap-guangzhou-2`.",
//
//	        "InstanceType": "S5.LARGE8"
//
//	    }
//
//	]
//
// }
//
// “`
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_CONFLICTNOTIFICATIONTARGET = "InvalidParameterValue.ConflictNotificationTarget"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"
//	INVALIDPARAMETERVALUE_USERGROUPIDNOTFOUND = "InvalidParameterValue.UserGroupIdNotFound"
//	LIMITEXCEEDED = "LimitExceeded"
//	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_CMQQUEUENOTFOUND = "ResourceNotFound.CmqQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQQUEUENOTFOUND = "ResourceNotFound.TDMQCMQQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQTOPICNOTFOUND = "ResourceNotFound.TDMQCMQTopicNotFound"
//	RESOURCEUNAVAILABLE_CMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.CmqTopicHasNoSubscriber"
//	RESOURCEUNAVAILABLE_TDMQCMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.TDMQCMQTopicHasNoSubscriber"
func (c *Client) CreateNotificationConfiguration(request *CreateNotificationConfigurationRequest) (response *CreateNotificationConfigurationResponse, err error) {
	if request == nil {
		request = NewCreateNotificationConfigurationRequest()
	}

	response = NewCreateNotificationConfigurationResponse()
	err = c.Send(request, response)
	return
}

// CreateNotificationConfiguration
// 本接口（CreateNotificationConfiguration）用于创建通知。
//
// 通知到 CMQ 主题或队列时，消息内容如下：
//
// “`
//
// {
//
//	"Service": "Tencent Cloud Auto Scaling",
//
//	"CreatedTime": "2021-10-11T10:15:11Z", // 活动创建时间
//
//	"AppId": "100000000",
//
//	"ActivityId": "asa-fznnvrja", // 伸缩活动ID
//
//	"AutoScalingGroupId": "asg-pc2oqu2z", // 伸缩组ID
//
//	"ActivityType": "SCALE_OUT",  // 伸缩活动类型
//
//	"StatusCode": "SUCCESSFUL",   // 伸缩活动结果
//
//	"Description": "Activity was launched in response to a difference between desired capacity and actual capacity,
//
//	scale out 1 instance(s).", // 伸缩活动描述
//
//	"StartTime": "2021-10-11T10:15:11Z",  // 活动开始时间
//
//	"EndTime": "2021-10-11T10:15:32Z",    // 活动结束时间
//
//	"DetailedStatusMessageSet": [ // 活动内部错误集合（非空不代表活动失败）
//
//	    {
//
//	        "Code": "InvalidInstanceType",
//
//	        "Zone": "ap-guangzhou-2",
//
//	        "InstanceId": "",
//
//	        "InstanceChargeType": "POSTPAID_BY_HOUR",
//
//	        "SubnetId": "subnet-4t5mgeuu",
//
//	        "Message": "The specified instance type `S5.LARGE8` is invalid in `subnet-4t5mgeuu`, `ap-guangzhou-2`.",
//
//	        "InstanceType": "S5.LARGE8"
//
//	    }
//
//	]
//
// }
//
// “`
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_CONFLICTNOTIFICATIONTARGET = "InvalidParameterValue.ConflictNotificationTarget"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"
//	INVALIDPARAMETERVALUE_USERGROUPIDNOTFOUND = "InvalidParameterValue.UserGroupIdNotFound"
//	LIMITEXCEEDED = "LimitExceeded"
//	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_CMQQUEUENOTFOUND = "ResourceNotFound.CmqQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQQUEUENOTFOUND = "ResourceNotFound.TDMQCMQQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQTOPICNOTFOUND = "ResourceNotFound.TDMQCMQTopicNotFound"
//	RESOURCEUNAVAILABLE_CMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.CmqTopicHasNoSubscriber"
//	RESOURCEUNAVAILABLE_TDMQCMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.TDMQCMQTopicHasNoSubscriber"
func (c *Client) CreateNotificationConfigurationWithContext(ctx context.Context, request *CreateNotificationConfigurationRequest) (response *CreateNotificationConfigurationResponse, err error) {
	if request == nil {
		request = NewCreateNotificationConfigurationRequest()
	}
	request.SetContext(ctx)

	response = NewCreateNotificationConfigurationResponse()
	err = c.Send(request, response)
	return
}

func NewCreateScalingPolicyRequest() (request *CreateScalingPolicyRequest) {
	request = &CreateScalingPolicyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "CreateScalingPolicy")

	return
}

func NewCreateScalingPolicyResponse() (response *CreateScalingPolicyResponse) {
	response = &CreateScalingPolicyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateScalingPolicy
// 本接口（CreateScalingPolicy）用于创建告警触发策略。
//
// 可能返回的错误码:
//
//	INTERNALERROR_CALLNOTIFICATIONERROR = "InternalError.CallNotificationError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SCALINGPOLICYNAMEDUPLICATE = "InvalidParameterValue.ScalingPolicyNameDuplicate"
//	INVALIDPARAMETERVALUE_THRESHOLDOUTOFRANGE = "InvalidParameterValue.ThresholdOutOfRange"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_USERGROUPIDNOTFOUND = "InvalidParameterValue.UserGroupIdNotFound"
//	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) CreateScalingPolicy(request *CreateScalingPolicyRequest) (response *CreateScalingPolicyResponse, err error) {
	if request == nil {
		request = NewCreateScalingPolicyRequest()
	}

	response = NewCreateScalingPolicyResponse()
	err = c.Send(request, response)
	return
}

// CreateScalingPolicy
// 本接口（CreateScalingPolicy）用于创建告警触发策略。
//
// 可能返回的错误码:
//
//	INTERNALERROR_CALLNOTIFICATIONERROR = "InternalError.CallNotificationError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SCALINGPOLICYNAMEDUPLICATE = "InvalidParameterValue.ScalingPolicyNameDuplicate"
//	INVALIDPARAMETERVALUE_THRESHOLDOUTOFRANGE = "InvalidParameterValue.ThresholdOutOfRange"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_USERGROUPIDNOTFOUND = "InvalidParameterValue.UserGroupIdNotFound"
//	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) CreateScalingPolicyWithContext(ctx context.Context, request *CreateScalingPolicyRequest) (response *CreateScalingPolicyResponse, err error) {
	if request == nil {
		request = NewCreateScalingPolicyRequest()
	}
	request.SetContext(ctx)

	response = NewCreateScalingPolicyResponse()
	err = c.Send(request, response)
	return
}

func NewCreateScheduledActionRequest() (request *CreateScheduledActionRequest) {
	request = &CreateScheduledActionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "CreateScheduledAction")

	return
}

func NewCreateScheduledActionResponse() (response *CreateScheduledActionResponse) {
	response = &CreateScheduledActionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateScheduledAction
// 本接口（CreateScheduledAction）用于创建定时任务。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_CRONEXPRESSIONILLEGAL = "InvalidParameterValue.CronExpressionIllegal"
//	INVALIDPARAMETERVALUE_ENDTIMEBEFORESTARTTIME = "InvalidParameterValue.EndTimeBeforeStartTime"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONNAMEINCLUDEILLEGALCHAR = "InvalidParameterValue.InvalidScheduledActionNameIncludeIllegalChar"
//	INVALIDPARAMETERVALUE_SCHEDULEDACTIONNAMEDUPLICATE = "InvalidParameterValue.ScheduledActionNameDuplicate"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_STARTTIMEBEFORECURRENTTIME = "InvalidParameterValue.StartTimeBeforeCurrentTime"
//	INVALIDPARAMETERVALUE_TIMEFORMAT = "InvalidParameterValue.TimeFormat"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"
//	LIMITEXCEEDED_SCHEDULEDACTIONLIMITEXCEEDED = "LimitExceeded.ScheduledActionLimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) CreateScheduledAction(request *CreateScheduledActionRequest) (response *CreateScheduledActionResponse, err error) {
	if request == nil {
		request = NewCreateScheduledActionRequest()
	}

	response = NewCreateScheduledActionResponse()
	err = c.Send(request, response)
	return
}

// CreateScheduledAction
// 本接口（CreateScheduledAction）用于创建定时任务。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_CRONEXPRESSIONILLEGAL = "InvalidParameterValue.CronExpressionIllegal"
//	INVALIDPARAMETERVALUE_ENDTIMEBEFORESTARTTIME = "InvalidParameterValue.EndTimeBeforeStartTime"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONNAMEINCLUDEILLEGALCHAR = "InvalidParameterValue.InvalidScheduledActionNameIncludeIllegalChar"
//	INVALIDPARAMETERVALUE_SCHEDULEDACTIONNAMEDUPLICATE = "InvalidParameterValue.ScheduledActionNameDuplicate"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_STARTTIMEBEFORECURRENTTIME = "InvalidParameterValue.StartTimeBeforeCurrentTime"
//	INVALIDPARAMETERVALUE_TIMEFORMAT = "InvalidParameterValue.TimeFormat"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"
//	LIMITEXCEEDED_SCHEDULEDACTIONLIMITEXCEEDED = "LimitExceeded.ScheduledActionLimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) CreateScheduledActionWithContext(ctx context.Context, request *CreateScheduledActionRequest) (response *CreateScheduledActionResponse, err error) {
	if request == nil {
		request = NewCreateScheduledActionRequest()
	}
	request.SetContext(ctx)

	response = NewCreateScheduledActionResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteAutoScalingGroupRequest() (request *DeleteAutoScalingGroupRequest) {
	request = &DeleteAutoScalingGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DeleteAutoScalingGroup")

	return
}

func NewDeleteAutoScalingGroupResponse() (response *DeleteAutoScalingGroupResponse) {
	response = &DeleteAutoScalingGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteAutoScalingGroup
// 本接口（DeleteAutoScalingGroup）用于删除指定伸缩组，删除前提是伸缩组内无实例且当前未在执行伸缩活动。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLERROR = "InternalError.CallError"
//	INTERNALERROR_CALLMONITORERROR = "InternalError.CallMonitorError"
//	INTERNALERROR_CALLTAGERROR = "InternalError.CallTagError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	RESOURCEINUSE_ACTIVITYINPROGRESS = "ResourceInUse.ActivityInProgress"
//	RESOURCEINUSE_INSTANCEINGROUP = "ResourceInUse.InstanceInGroup"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) DeleteAutoScalingGroup(request *DeleteAutoScalingGroupRequest) (response *DeleteAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewDeleteAutoScalingGroupRequest()
	}

	response = NewDeleteAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

// DeleteAutoScalingGroup
// 本接口（DeleteAutoScalingGroup）用于删除指定伸缩组，删除前提是伸缩组内无实例且当前未在执行伸缩活动。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLERROR = "InternalError.CallError"
//	INTERNALERROR_CALLMONITORERROR = "InternalError.CallMonitorError"
//	INTERNALERROR_CALLTAGERROR = "InternalError.CallTagError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	RESOURCEINUSE_ACTIVITYINPROGRESS = "ResourceInUse.ActivityInProgress"
//	RESOURCEINUSE_INSTANCEINGROUP = "ResourceInUse.InstanceInGroup"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) DeleteAutoScalingGroupWithContext(ctx context.Context, request *DeleteAutoScalingGroupRequest) (response *DeleteAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewDeleteAutoScalingGroupRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteLaunchConfigurationRequest() (request *DeleteLaunchConfigurationRequest) {
	request = &DeleteLaunchConfigurationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DeleteLaunchConfiguration")

	return
}

func NewDeleteLaunchConfigurationResponse() (response *DeleteLaunchConfigurationResponse) {
	response = &DeleteLaunchConfigurationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteLaunchConfiguration
// 本接口（DeleteLaunchConfiguration）用于删除启动配置。
//
// * 若启动配置在伸缩组中属于生效状态，则该启动配置不允许删除。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	RESOURCEINUSE_LAUNCHCONFIGURATIONIDINUSE = "ResourceInUse.LaunchConfigurationIdInUse"
//	RESOURCENOTFOUND_LAUNCHCONFIGURATIONIDNOTFOUND = "ResourceNotFound.LaunchConfigurationIdNotFound"
func (c *Client) DeleteLaunchConfiguration(request *DeleteLaunchConfigurationRequest) (response *DeleteLaunchConfigurationResponse, err error) {
	if request == nil {
		request = NewDeleteLaunchConfigurationRequest()
	}

	response = NewDeleteLaunchConfigurationResponse()
	err = c.Send(request, response)
	return
}

// DeleteLaunchConfiguration
// 本接口（DeleteLaunchConfiguration）用于删除启动配置。
//
// * 若启动配置在伸缩组中属于生效状态，则该启动配置不允许删除。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	RESOURCEINUSE_LAUNCHCONFIGURATIONIDINUSE = "ResourceInUse.LaunchConfigurationIdInUse"
//	RESOURCENOTFOUND_LAUNCHCONFIGURATIONIDNOTFOUND = "ResourceNotFound.LaunchConfigurationIdNotFound"
func (c *Client) DeleteLaunchConfigurationWithContext(ctx context.Context, request *DeleteLaunchConfigurationRequest) (response *DeleteLaunchConfigurationResponse, err error) {
	if request == nil {
		request = NewDeleteLaunchConfigurationRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteLaunchConfigurationResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteLifecycleHookRequest() (request *DeleteLifecycleHookRequest) {
	request = &DeleteLifecycleHookRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DeleteLifecycleHook")

	return
}

func NewDeleteLifecycleHookResponse() (response *DeleteLifecycleHookResponse) {
	response = &DeleteLifecycleHookResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteLifecycleHook
// 本接口（DeleteLifecycleHook）用于删除生命周期挂钩。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_LIFECYCLEHOOKNOTFOUND = "ResourceNotFound.LifecycleHookNotFound"
func (c *Client) DeleteLifecycleHook(request *DeleteLifecycleHookRequest) (response *DeleteLifecycleHookResponse, err error) {
	if request == nil {
		request = NewDeleteLifecycleHookRequest()
	}

	response = NewDeleteLifecycleHookResponse()
	err = c.Send(request, response)
	return
}

// DeleteLifecycleHook
// 本接口（DeleteLifecycleHook）用于删除生命周期挂钩。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_LIFECYCLEHOOKNOTFOUND = "ResourceNotFound.LifecycleHookNotFound"
func (c *Client) DeleteLifecycleHookWithContext(ctx context.Context, request *DeleteLifecycleHookRequest) (response *DeleteLifecycleHookResponse, err error) {
	if request == nil {
		request = NewDeleteLifecycleHookRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteLifecycleHookResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteNotificationConfigurationRequest() (request *DeleteNotificationConfigurationRequest) {
	request = &DeleteNotificationConfigurationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DeleteNotificationConfiguration")

	return
}

func NewDeleteNotificationConfigurationResponse() (response *DeleteNotificationConfigurationResponse) {
	response = &DeleteNotificationConfigurationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteNotificationConfiguration
// 本接口（DeleteNotificationConfiguration）用于删除特定的通知。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGNOTIFICATIONID = "InvalidParameterValue.InvalidAutoScalingNotificationId"
//	RESOURCENOTFOUND_AUTOSCALINGNOTIFICATIONNOTFOUND = "ResourceNotFound.AutoScalingNotificationNotFound"
func (c *Client) DeleteNotificationConfiguration(request *DeleteNotificationConfigurationRequest) (response *DeleteNotificationConfigurationResponse, err error) {
	if request == nil {
		request = NewDeleteNotificationConfigurationRequest()
	}

	response = NewDeleteNotificationConfigurationResponse()
	err = c.Send(request, response)
	return
}

// DeleteNotificationConfiguration
// 本接口（DeleteNotificationConfiguration）用于删除特定的通知。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGNOTIFICATIONID = "InvalidParameterValue.InvalidAutoScalingNotificationId"
//	RESOURCENOTFOUND_AUTOSCALINGNOTIFICATIONNOTFOUND = "ResourceNotFound.AutoScalingNotificationNotFound"
func (c *Client) DeleteNotificationConfigurationWithContext(ctx context.Context, request *DeleteNotificationConfigurationRequest) (response *DeleteNotificationConfigurationResponse, err error) {
	if request == nil {
		request = NewDeleteNotificationConfigurationRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteNotificationConfigurationResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteScalingPolicyRequest() (request *DeleteScalingPolicyRequest) {
	request = &DeleteScalingPolicyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DeleteScalingPolicy")

	return
}

func NewDeleteScalingPolicyResponse() (response *DeleteScalingPolicyResponse) {
	response = &DeleteScalingPolicyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteScalingPolicy
// 本接口（DeleteScalingPolicy）用于删除告警触发策略。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"
//	RESOURCENOTFOUND_SCALINGPOLICYNOTFOUND = "ResourceNotFound.ScalingPolicyNotFound"
func (c *Client) DeleteScalingPolicy(request *DeleteScalingPolicyRequest) (response *DeleteScalingPolicyResponse, err error) {
	if request == nil {
		request = NewDeleteScalingPolicyRequest()
	}

	response = NewDeleteScalingPolicyResponse()
	err = c.Send(request, response)
	return
}

// DeleteScalingPolicy
// 本接口（DeleteScalingPolicy）用于删除告警触发策略。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"
//	RESOURCENOTFOUND_SCALINGPOLICYNOTFOUND = "ResourceNotFound.ScalingPolicyNotFound"
func (c *Client) DeleteScalingPolicyWithContext(ctx context.Context, request *DeleteScalingPolicyRequest) (response *DeleteScalingPolicyResponse, err error) {
	if request == nil {
		request = NewDeleteScalingPolicyRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteScalingPolicyResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteScheduledActionRequest() (request *DeleteScheduledActionRequest) {
	request = &DeleteScheduledActionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DeleteScheduledAction")

	return
}

func NewDeleteScheduledActionResponse() (response *DeleteScheduledActionResponse) {
	response = &DeleteScheduledActionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteScheduledAction
// 本接口（DeleteScheduledAction）用于删除特定的定时任务。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONID = "InvalidParameterValue.InvalidScheduledActionId"
//	RESOURCENOTFOUND_SCHEDULEDACTIONNOTFOUND = "ResourceNotFound.ScheduledActionNotFound"
func (c *Client) DeleteScheduledAction(request *DeleteScheduledActionRequest) (response *DeleteScheduledActionResponse, err error) {
	if request == nil {
		request = NewDeleteScheduledActionRequest()
	}

	response = NewDeleteScheduledActionResponse()
	err = c.Send(request, response)
	return
}

// DeleteScheduledAction
// 本接口（DeleteScheduledAction）用于删除特定的定时任务。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONID = "InvalidParameterValue.InvalidScheduledActionId"
//	RESOURCENOTFOUND_SCHEDULEDACTIONNOTFOUND = "ResourceNotFound.ScheduledActionNotFound"
func (c *Client) DeleteScheduledActionWithContext(ctx context.Context, request *DeleteScheduledActionRequest) (response *DeleteScheduledActionResponse, err error) {
	if request == nil {
		request = NewDeleteScheduledActionRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteScheduledActionResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAccountLimitsRequest() (request *DescribeAccountLimitsRequest) {
	request = &DescribeAccountLimitsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeAccountLimits")

	return
}

func NewDescribeAccountLimitsResponse() (response *DescribeAccountLimitsResponse) {
	response = &DescribeAccountLimitsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAccountLimits
// 本接口（DescribeAccountLimits）用于查询用户账户在弹性伸缩中的资源限制。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
func (c *Client) DescribeAccountLimits(request *DescribeAccountLimitsRequest) (response *DescribeAccountLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeAccountLimitsRequest()
	}

	response = NewDescribeAccountLimitsResponse()
	err = c.Send(request, response)
	return
}

// DescribeAccountLimits
// 本接口（DescribeAccountLimits）用于查询用户账户在弹性伸缩中的资源限制。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
func (c *Client) DescribeAccountLimitsWithContext(ctx context.Context, request *DescribeAccountLimitsRequest) (response *DescribeAccountLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeAccountLimitsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAccountLimitsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAutoScalingActivitiesRequest() (request *DescribeAutoScalingActivitiesRequest) {
	request = &DescribeAutoScalingActivitiesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeAutoScalingActivities")

	return
}

func NewDescribeAutoScalingActivitiesResponse() (response *DescribeAutoScalingActivitiesResponse) {
	response = &DescribeAutoScalingActivitiesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAutoScalingActivities
// 本接口（DescribeAutoScalingActivities）用于查询伸缩组的伸缩活动记录。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDACTIVITYID = "InvalidParameterValue.InvalidActivityId"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	LIMITEXCEEDED_FILTERVALUESTOOLONG = "LimitExceeded.FilterValuesTooLong"
func (c *Client) DescribeAutoScalingActivities(request *DescribeAutoScalingActivitiesRequest) (response *DescribeAutoScalingActivitiesResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingActivitiesRequest()
	}

	response = NewDescribeAutoScalingActivitiesResponse()
	err = c.Send(request, response)
	return
}

// DescribeAutoScalingActivities
// 本接口（DescribeAutoScalingActivities）用于查询伸缩组的伸缩活动记录。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDACTIVITYID = "InvalidParameterValue.InvalidActivityId"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	LIMITEXCEEDED_FILTERVALUESTOOLONG = "LimitExceeded.FilterValuesTooLong"
func (c *Client) DescribeAutoScalingActivitiesWithContext(ctx context.Context, request *DescribeAutoScalingActivitiesRequest) (response *DescribeAutoScalingActivitiesResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingActivitiesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAutoScalingActivitiesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAutoScalingAdvicesRequest() (request *DescribeAutoScalingAdvicesRequest) {
	request = &DescribeAutoScalingAdvicesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeAutoScalingAdvices")

	return
}

func NewDescribeAutoScalingAdvicesResponse() (response *DescribeAutoScalingAdvicesResponse) {
	response = &DescribeAutoScalingAdvicesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAutoScalingAdvices
// 此接口用于查询伸缩组配置建议。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
func (c *Client) DescribeAutoScalingAdvices(request *DescribeAutoScalingAdvicesRequest) (response *DescribeAutoScalingAdvicesResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingAdvicesRequest()
	}

	response = NewDescribeAutoScalingAdvicesResponse()
	err = c.Send(request, response)
	return
}

// DescribeAutoScalingAdvices
// 此接口用于查询伸缩组配置建议。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
func (c *Client) DescribeAutoScalingAdvicesWithContext(ctx context.Context, request *DescribeAutoScalingAdvicesRequest) (response *DescribeAutoScalingAdvicesResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingAdvicesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAutoScalingAdvicesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAutoScalingGroupLastActivitiesRequest() (request *DescribeAutoScalingGroupLastActivitiesRequest) {
	request = &DescribeAutoScalingGroupLastActivitiesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeAutoScalingGroupLastActivities")

	return
}

func NewDescribeAutoScalingGroupLastActivitiesResponse() (response *DescribeAutoScalingGroupLastActivitiesResponse) {
	response = &DescribeAutoScalingGroupLastActivitiesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAutoScalingGroupLastActivities
// 本接口（DescribeAutoScalingGroupLastActivities）用于查询伸缩组的最新一次伸缩活动记录。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_NORESOURCEPERMISSION = "InvalidParameterValue.NoResourcePermission"
func (c *Client) DescribeAutoScalingGroupLastActivities(request *DescribeAutoScalingGroupLastActivitiesRequest) (response *DescribeAutoScalingGroupLastActivitiesResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingGroupLastActivitiesRequest()
	}

	response = NewDescribeAutoScalingGroupLastActivitiesResponse()
	err = c.Send(request, response)
	return
}

// DescribeAutoScalingGroupLastActivities
// 本接口（DescribeAutoScalingGroupLastActivities）用于查询伸缩组的最新一次伸缩活动记录。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_NORESOURCEPERMISSION = "InvalidParameterValue.NoResourcePermission"
func (c *Client) DescribeAutoScalingGroupLastActivitiesWithContext(ctx context.Context, request *DescribeAutoScalingGroupLastActivitiesRequest) (response *DescribeAutoScalingGroupLastActivitiesResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingGroupLastActivitiesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAutoScalingGroupLastActivitiesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAutoScalingGroupsRequest() (request *DescribeAutoScalingGroupsRequest) {
	request = &DescribeAutoScalingGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeAutoScalingGroups")

	return
}

func NewDescribeAutoScalingGroupsResponse() (response *DescribeAutoScalingGroupsResponse) {
	response = &DescribeAutoScalingGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAutoScalingGroups
// 本接口（DescribeAutoScalingGroups）用于查询伸缩组信息。
//
// * 可以根据伸缩组ID、伸缩组名称或者启动配置ID等信息来查询伸缩组的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的伸缩组。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPERMISSION = "InvalidPermission"
func (c *Client) DescribeAutoScalingGroups(request *DescribeAutoScalingGroupsRequest) (response *DescribeAutoScalingGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingGroupsRequest()
	}

	response = NewDescribeAutoScalingGroupsResponse()
	err = c.Send(request, response)
	return
}

// DescribeAutoScalingGroups
// 本接口（DescribeAutoScalingGroups）用于查询伸缩组信息。
//
// * 可以根据伸缩组ID、伸缩组名称或者启动配置ID等信息来查询伸缩组的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的伸缩组。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPERMISSION = "InvalidPermission"
func (c *Client) DescribeAutoScalingGroupsWithContext(ctx context.Context, request *DescribeAutoScalingGroupsRequest) (response *DescribeAutoScalingGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAutoScalingGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAutoScalingInstancesRequest() (request *DescribeAutoScalingInstancesRequest) {
	request = &DescribeAutoScalingInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeAutoScalingInstances")

	return
}

func NewDescribeAutoScalingInstancesResponse() (response *DescribeAutoScalingInstancesResponse) {
	response = &DescribeAutoScalingInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAutoScalingInstances
// 本接口（DescribeAutoScalingInstances）用于查询弹性伸缩关联实例的信息。
//
// * 可以根据实例ID、伸缩组ID等信息来查询实例的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的实例。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	LIMITEXCEEDED_FILTERVALUESTOOLONG = "LimitExceeded.FilterValuesTooLong"
func (c *Client) DescribeAutoScalingInstances(request *DescribeAutoScalingInstancesRequest) (response *DescribeAutoScalingInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingInstancesRequest()
	}

	response = NewDescribeAutoScalingInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeAutoScalingInstances
// 本接口（DescribeAutoScalingInstances）用于查询弹性伸缩关联实例的信息。
//
// * 可以根据实例ID、伸缩组ID等信息来查询实例的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的实例。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	LIMITEXCEEDED_FILTERVALUESTOOLONG = "LimitExceeded.FilterValuesTooLong"
func (c *Client) DescribeAutoScalingInstancesWithContext(ctx context.Context, request *DescribeAutoScalingInstancesRequest) (response *DescribeAutoScalingInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeAutoScalingInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAutoScalingInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeLaunchConfigurationsRequest() (request *DescribeLaunchConfigurationsRequest) {
	request = &DescribeLaunchConfigurationsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeLaunchConfigurations")

	return
}

func NewDescribeLaunchConfigurationsResponse() (response *DescribeLaunchConfigurationsResponse) {
	response = &DescribeLaunchConfigurationsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeLaunchConfigurations
// 本接口（DescribeLaunchConfigurations）用于查询启动配置的信息。
//
// * 可以根据启动配置ID、启动配置名称等信息来查询启动配置的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的启动配置。
//
// 可能返回的错误码:
//
//	INVALIDLAUNCHCONFIGURATION = "InvalidLaunchConfiguration"
//	INVALIDLAUNCHCONFIGURATIONID = "InvalidLaunchConfigurationId"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPERMISSION = "InvalidPermission"
func (c *Client) DescribeLaunchConfigurations(request *DescribeLaunchConfigurationsRequest) (response *DescribeLaunchConfigurationsResponse, err error) {
	if request == nil {
		request = NewDescribeLaunchConfigurationsRequest()
	}

	response = NewDescribeLaunchConfigurationsResponse()
	err = c.Send(request, response)
	return
}

// DescribeLaunchConfigurations
// 本接口（DescribeLaunchConfigurations）用于查询启动配置的信息。
//
// * 可以根据启动配置ID、启动配置名称等信息来查询启动配置的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的启动配置。
//
// 可能返回的错误码:
//
//	INVALIDLAUNCHCONFIGURATION = "InvalidLaunchConfiguration"
//	INVALIDLAUNCHCONFIGURATIONID = "InvalidLaunchConfigurationId"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPERMISSION = "InvalidPermission"
func (c *Client) DescribeLaunchConfigurationsWithContext(ctx context.Context, request *DescribeLaunchConfigurationsRequest) (response *DescribeLaunchConfigurationsResponse, err error) {
	if request == nil {
		request = NewDescribeLaunchConfigurationsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeLaunchConfigurationsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeLifecycleHooksRequest() (request *DescribeLifecycleHooksRequest) {
	request = &DescribeLifecycleHooksRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeLifecycleHooks")

	return
}

func NewDescribeLifecycleHooksResponse() (response *DescribeLifecycleHooksResponse) {
	response = &DescribeLifecycleHooksResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeLifecycleHooks
// 本接口（DescribeLifecycleHooks）用于查询生命周期挂钩信息。
//
// * 可以根据伸缩组ID、生命周期挂钩ID或者生命周期挂钩名称等信息来查询生命周期挂钩的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的生命周期挂钩。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeLifecycleHooks(request *DescribeLifecycleHooksRequest) (response *DescribeLifecycleHooksResponse, err error) {
	if request == nil {
		request = NewDescribeLifecycleHooksRequest()
	}

	response = NewDescribeLifecycleHooksResponse()
	err = c.Send(request, response)
	return
}

// DescribeLifecycleHooks
// 本接口（DescribeLifecycleHooks）用于查询生命周期挂钩信息。
//
// * 可以根据伸缩组ID、生命周期挂钩ID或者生命周期挂钩名称等信息来查询生命周期挂钩的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的生命周期挂钩。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeLifecycleHooksWithContext(ctx context.Context, request *DescribeLifecycleHooksRequest) (response *DescribeLifecycleHooksResponse, err error) {
	if request == nil {
		request = NewDescribeLifecycleHooksRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeLifecycleHooksResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNotificationConfigurationsRequest() (request *DescribeNotificationConfigurationsRequest) {
	request = &DescribeNotificationConfigurationsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeNotificationConfigurations")

	return
}

func NewDescribeNotificationConfigurationsResponse() (response *DescribeNotificationConfigurationsResponse) {
	response = &DescribeNotificationConfigurationsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNotificationConfigurations
// 本接口 (DescribeNotificationConfigurations) 用于查询一个或多个通知的详细信息。
//
// 可以根据通知ID、伸缩组ID等信息来查询通知的详细信息。过滤信息详细请见过滤器`Filter`。
//
// 如果参数为空，返回当前用户一定数量（Limit所指定的数量，默认为20）的通知。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGNOTIFICATIONID = "InvalidParameterValue.InvalidAutoScalingNotificationId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
func (c *Client) DescribeNotificationConfigurations(request *DescribeNotificationConfigurationsRequest) (response *DescribeNotificationConfigurationsResponse, err error) {
	if request == nil {
		request = NewDescribeNotificationConfigurationsRequest()
	}

	response = NewDescribeNotificationConfigurationsResponse()
	err = c.Send(request, response)
	return
}

// DescribeNotificationConfigurations
// 本接口 (DescribeNotificationConfigurations) 用于查询一个或多个通知的详细信息。
//
// 可以根据通知ID、伸缩组ID等信息来查询通知的详细信息。过滤信息详细请见过滤器`Filter`。
//
// 如果参数为空，返回当前用户一定数量（Limit所指定的数量，默认为20）的通知。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGNOTIFICATIONID = "InvalidParameterValue.InvalidAutoScalingNotificationId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
func (c *Client) DescribeNotificationConfigurationsWithContext(ctx context.Context, request *DescribeNotificationConfigurationsRequest) (response *DescribeNotificationConfigurationsResponse, err error) {
	if request == nil {
		request = NewDescribeNotificationConfigurationsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNotificationConfigurationsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePaiInstancesRequest() (request *DescribePaiInstancesRequest) {
	request = &DescribePaiInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribePaiInstances")

	return
}

func NewDescribePaiInstancesResponse() (response *DescribePaiInstancesResponse) {
	response = &DescribePaiInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePaiInstances
// 本接口（DescribePaiInstances）用于查询PAI实例信息。
//
// * 可以根据实例ID、实例域名等信息来查询PAI实例的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的PAI实例。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
func (c *Client) DescribePaiInstances(request *DescribePaiInstancesRequest) (response *DescribePaiInstancesResponse, err error) {
	if request == nil {
		request = NewDescribePaiInstancesRequest()
	}

	response = NewDescribePaiInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribePaiInstances
// 本接口（DescribePaiInstances）用于查询PAI实例信息。
//
// * 可以根据实例ID、实例域名等信息来查询PAI实例的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的PAI实例。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
func (c *Client) DescribePaiInstancesWithContext(ctx context.Context, request *DescribePaiInstancesRequest) (response *DescribePaiInstancesResponse, err error) {
	if request == nil {
		request = NewDescribePaiInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePaiInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeScalingPoliciesRequest() (request *DescribeScalingPoliciesRequest) {
	request = &DescribeScalingPoliciesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeScalingPolicies")

	return
}

func NewDescribeScalingPoliciesResponse() (response *DescribeScalingPoliciesResponse) {
	response = &DescribeScalingPoliciesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeScalingPolicies
// 本接口（DescribeScalingPolicies）用于查询告警触发策略。
//
// 可能返回的错误码:
//
//	INTERNALERROR_CALLMONITORERROR = "InternalError.CallMonitorError"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) DescribeScalingPolicies(request *DescribeScalingPoliciesRequest) (response *DescribeScalingPoliciesResponse, err error) {
	if request == nil {
		request = NewDescribeScalingPoliciesRequest()
	}

	response = NewDescribeScalingPoliciesResponse()
	err = c.Send(request, response)
	return
}

// DescribeScalingPolicies
// 本接口（DescribeScalingPolicies）用于查询告警触发策略。
//
// 可能返回的错误码:
//
//	INTERNALERROR_CALLMONITORERROR = "InternalError.CallMonitorError"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) DescribeScalingPoliciesWithContext(ctx context.Context, request *DescribeScalingPoliciesRequest) (response *DescribeScalingPoliciesResponse, err error) {
	if request == nil {
		request = NewDescribeScalingPoliciesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeScalingPoliciesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeScheduledActionsRequest() (request *DescribeScheduledActionsRequest) {
	request = &DescribeScheduledActionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DescribeScheduledActions")

	return
}

func NewDescribeScheduledActionsResponse() (response *DescribeScheduledActionsResponse) {
	response = &DescribeScheduledActionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeScheduledActions
// 本接口 (DescribeScheduledActions) 用于查询一个或多个定时任务的详细信息。
//
// * 可以根据定时任务ID、定时任务名称或者伸缩组ID等信息来查询定时任务的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（Limit所指定的数量，默认为20）的定时任务。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONID = "InvalidParameterValue.InvalidScheduledActionId"
//	RESOURCENOTFOUND_SCHEDULEDACTIONNOTFOUND = "ResourceNotFound.ScheduledActionNotFound"
func (c *Client) DescribeScheduledActions(request *DescribeScheduledActionsRequest) (response *DescribeScheduledActionsResponse, err error) {
	if request == nil {
		request = NewDescribeScheduledActionsRequest()
	}

	response = NewDescribeScheduledActionsResponse()
	err = c.Send(request, response)
	return
}

// DescribeScheduledActions
// 本接口 (DescribeScheduledActions) 用于查询一个或多个定时任务的详细信息。
//
// * 可以根据定时任务ID、定时任务名称或者伸缩组ID等信息来查询定时任务的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（Limit所指定的数量，默认为20）的定时任务。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONID = "InvalidParameterValue.InvalidScheduledActionId"
//	RESOURCENOTFOUND_SCHEDULEDACTIONNOTFOUND = "ResourceNotFound.ScheduledActionNotFound"
func (c *Client) DescribeScheduledActionsWithContext(ctx context.Context, request *DescribeScheduledActionsRequest) (response *DescribeScheduledActionsResponse, err error) {
	if request == nil {
		request = NewDescribeScheduledActionsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeScheduledActionsResponse()
	err = c.Send(request, response)
	return
}

func NewDetachInstancesRequest() (request *DetachInstancesRequest) {
	request = &DetachInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DetachInstances")

	return
}

func NewDetachInstancesResponse() (response *DetachInstancesResponse) {
	response = &DetachInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DetachInstances
// 本接口（DetachInstances）用于从伸缩组移出 CVM 实例，本接口不会销毁实例。
//
// * 如果移出指定实例后，伸缩组内处于`IN_SERVICE`状态的实例数量小于伸缩组最小值，接口将报错
//
// * 如果伸缩组处于`DISABLED`状态，移出操作不校验`IN_SERVICE`实例数量和最小值的关系
//
// * 对于伸缩组配置的 CLB，实例在离开伸缩组时，AS 会进行解挂载动作
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEBELOWMINSIZE = "ResourceInsufficient.InServiceInstanceBelowMinSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_INSTANCEINOPERATION = "ResourceUnavailable.InstanceInOperation"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) DetachInstances(request *DetachInstancesRequest) (response *DetachInstancesResponse, err error) {
	if request == nil {
		request = NewDetachInstancesRequest()
	}

	response = NewDetachInstancesResponse()
	err = c.Send(request, response)
	return
}

// DetachInstances
// 本接口（DetachInstances）用于从伸缩组移出 CVM 实例，本接口不会销毁实例。
//
// * 如果移出指定实例后，伸缩组内处于`IN_SERVICE`状态的实例数量小于伸缩组最小值，接口将报错
//
// * 如果伸缩组处于`DISABLED`状态，移出操作不校验`IN_SERVICE`实例数量和最小值的关系
//
// * 对于伸缩组配置的 CLB，实例在离开伸缩组时，AS 会进行解挂载动作
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEBELOWMINSIZE = "ResourceInsufficient.InServiceInstanceBelowMinSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_INSTANCEINOPERATION = "ResourceUnavailable.InstanceInOperation"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) DetachInstancesWithContext(ctx context.Context, request *DetachInstancesRequest) (response *DetachInstancesResponse, err error) {
	if request == nil {
		request = NewDetachInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDetachInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDetachLoadBalancersRequest() (request *DetachLoadBalancersRequest) {
	request = &DetachLoadBalancersRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DetachLoadBalancers")

	return
}

func NewDetachLoadBalancersResponse() (response *DetachLoadBalancersResponse) {
	response = &DetachLoadBalancersResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DetachLoadBalancers
// 本接口（DetachLoadBalancers）用于从伸缩组移出负载均衡器，本接口不会销毁负载均衡器。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETER_LOADBALANCERNOTINAUTOSCALINGGROUP = "InvalidParameter.LoadBalancerNotInAutoScalingGroup"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	LIMITEXCEEDED_AFTERATTACHLBLIMITEXCEEDED = "LimitExceeded.AfterAttachLbLimitExceeded"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBPROJECTINCONSISTENT = "ResourceUnavailable.LbProjectInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) DetachLoadBalancers(request *DetachLoadBalancersRequest) (response *DetachLoadBalancersResponse, err error) {
	if request == nil {
		request = NewDetachLoadBalancersRequest()
	}

	response = NewDetachLoadBalancersResponse()
	err = c.Send(request, response)
	return
}

// DetachLoadBalancers
// 本接口（DetachLoadBalancers）用于从伸缩组移出负载均衡器，本接口不会销毁负载均衡器。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETER_LOADBALANCERNOTINAUTOSCALINGGROUP = "InvalidParameter.LoadBalancerNotInAutoScalingGroup"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	LIMITEXCEEDED_AFTERATTACHLBLIMITEXCEEDED = "LimitExceeded.AfterAttachLbLimitExceeded"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBPROJECTINCONSISTENT = "ResourceUnavailable.LbProjectInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) DetachLoadBalancersWithContext(ctx context.Context, request *DetachLoadBalancersRequest) (response *DetachLoadBalancersResponse, err error) {
	if request == nil {
		request = NewDetachLoadBalancersRequest()
	}
	request.SetContext(ctx)

	response = NewDetachLoadBalancersResponse()
	err = c.Send(request, response)
	return
}

func NewDisableAutoScalingGroupRequest() (request *DisableAutoScalingGroupRequest) {
	request = &DisableAutoScalingGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "DisableAutoScalingGroup")

	return
}

func NewDisableAutoScalingGroupResponse() (response *DisableAutoScalingGroupResponse) {
	response = &DisableAutoScalingGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisableAutoScalingGroup
// 本接口（DisableAutoScalingGroup）用于停用指定伸缩组。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) DisableAutoScalingGroup(request *DisableAutoScalingGroupRequest) (response *DisableAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewDisableAutoScalingGroupRequest()
	}

	response = NewDisableAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

// DisableAutoScalingGroup
// 本接口（DisableAutoScalingGroup）用于停用指定伸缩组。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) DisableAutoScalingGroupWithContext(ctx context.Context, request *DisableAutoScalingGroupRequest) (response *DisableAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewDisableAutoScalingGroupRequest()
	}
	request.SetContext(ctx)

	response = NewDisableAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

func NewEnableAutoScalingGroupRequest() (request *EnableAutoScalingGroupRequest) {
	request = &EnableAutoScalingGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "EnableAutoScalingGroup")

	return
}

func NewEnableAutoScalingGroupResponse() (response *EnableAutoScalingGroupResponse) {
	response = &EnableAutoScalingGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// EnableAutoScalingGroup
// 本接口（EnableAutoScalingGroup）用于启用指定伸缩组。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) EnableAutoScalingGroup(request *EnableAutoScalingGroupRequest) (response *EnableAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewEnableAutoScalingGroupRequest()
	}

	response = NewEnableAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

// EnableAutoScalingGroup
// 本接口（EnableAutoScalingGroup）用于启用指定伸缩组。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
func (c *Client) EnableAutoScalingGroupWithContext(ctx context.Context, request *EnableAutoScalingGroupRequest) (response *EnableAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewEnableAutoScalingGroupRequest()
	}
	request.SetContext(ctx)

	response = NewEnableAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

func NewExecuteScalingPolicyRequest() (request *ExecuteScalingPolicyRequest) {
	request = &ExecuteScalingPolicyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ExecuteScalingPolicy")

	return
}

func NewExecuteScalingPolicyResponse() (response *ExecuteScalingPolicyResponse) {
	response = &ExecuteScalingPolicyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ExecuteScalingPolicy
// 本接口（ExecuteScalingPolicy）用于执行伸缩策略。
//
// * 可以根据伸缩策略ID执行伸缩策略。
//
// * 伸缩策略所属伸缩组处于伸缩活动时，会拒绝执行伸缩策略。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"
//	RESOURCEINUSE_AUTOSCALINGGROUPNOTACTIVE = "ResourceInUse.AutoScalingGroupNotActive"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_SCALINGPOLICYNOTFOUND = "ResourceNotFound.ScalingPolicyNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPABNORMALSTATUS = "ResourceUnavailable.AutoScalingGroupAbnormalStatus"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
func (c *Client) ExecuteScalingPolicy(request *ExecuteScalingPolicyRequest) (response *ExecuteScalingPolicyResponse, err error) {
	if request == nil {
		request = NewExecuteScalingPolicyRequest()
	}

	response = NewExecuteScalingPolicyResponse()
	err = c.Send(request, response)
	return
}

// ExecuteScalingPolicy
// 本接口（ExecuteScalingPolicy）用于执行伸缩策略。
//
// * 可以根据伸缩策略ID执行伸缩策略。
//
// * 伸缩策略所属伸缩组处于伸缩活动时，会拒绝执行伸缩策略。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"
//	RESOURCEINUSE_AUTOSCALINGGROUPNOTACTIVE = "ResourceInUse.AutoScalingGroupNotActive"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_SCALINGPOLICYNOTFOUND = "ResourceNotFound.ScalingPolicyNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPABNORMALSTATUS = "ResourceUnavailable.AutoScalingGroupAbnormalStatus"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
func (c *Client) ExecuteScalingPolicyWithContext(ctx context.Context, request *ExecuteScalingPolicyRequest) (response *ExecuteScalingPolicyResponse, err error) {
	if request == nil {
		request = NewExecuteScalingPolicyRequest()
	}
	request.SetContext(ctx)

	response = NewExecuteScalingPolicyResponse()
	err = c.Send(request, response)
	return
}

func NewModifyAutoScalingGroupRequest() (request *ModifyAutoScalingGroupRequest) {
	request = &ModifyAutoScalingGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ModifyAutoScalingGroup")

	return
}

func NewModifyAutoScalingGroupResponse() (response *ModifyAutoScalingGroupResponse) {
	response = &ModifyAutoScalingGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyAutoScalingGroup
// 本接口（ModifyAutoScalingGroup）用于修改伸缩组。
//
// 可能返回的错误码:
//
//	INTERNALERROR_CALLVPCERROR = "InternalError.CallVpcError"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETERVALUE_BASECAPACITYTOOLARGE = "InvalidParameterValue.BaseCapacityTooLarge"
//	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"
//	INVALIDPARAMETERVALUE_DUPLICATEDSUBNET = "InvalidParameterValue.DuplicatedSubnet"
//	INVALIDPARAMETERVALUE_GROUPNAMEDUPLICATED = "InvalidParameterValue.GroupNameDuplicated"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_INVALIDSUBNETID = "InvalidParameterValue.InvalidSubnetId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNOTFOUND = "InvalidParameterValue.LaunchConfigurationNotFound"
//	INVALIDPARAMETERVALUE_LBPROJECTINCONSISTENT = "InvalidParameterValue.LbProjectInconsistent"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ONLYVPC = "InvalidParameterValue.OnlyVpc"
//	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_SUBNETIDS = "InvalidParameterValue.SubnetIds"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCEUNAVAILABLE_LAUNCHCONFIGURATIONSTATUSABNORMAL = "ResourceUnavailable.LaunchConfigurationStatusAbnormal"
//	RESOURCEUNAVAILABLE_PROJECTINCONSISTENT = "ResourceUnavailable.ProjectInconsistent"
func (c *Client) ModifyAutoScalingGroup(request *ModifyAutoScalingGroupRequest) (response *ModifyAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewModifyAutoScalingGroupRequest()
	}

	response = NewModifyAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

// ModifyAutoScalingGroup
// 本接口（ModifyAutoScalingGroup）用于修改伸缩组。
//
// 可能返回的错误码:
//
//	INTERNALERROR_CALLVPCERROR = "InternalError.CallVpcError"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETERVALUE_BASECAPACITYTOOLARGE = "InvalidParameterValue.BaseCapacityTooLarge"
//	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"
//	INVALIDPARAMETERVALUE_DUPLICATEDSUBNET = "InvalidParameterValue.DuplicatedSubnet"
//	INVALIDPARAMETERVALUE_GROUPNAMEDUPLICATED = "InvalidParameterValue.GroupNameDuplicated"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_INVALIDSUBNETID = "InvalidParameterValue.InvalidSubnetId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNOTFOUND = "InvalidParameterValue.LaunchConfigurationNotFound"
//	INVALIDPARAMETERVALUE_LBPROJECTINCONSISTENT = "InvalidParameterValue.LbProjectInconsistent"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ONLYVPC = "InvalidParameterValue.OnlyVpc"
//	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_SUBNETIDS = "InvalidParameterValue.SubnetIds"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCEUNAVAILABLE_LAUNCHCONFIGURATIONSTATUSABNORMAL = "ResourceUnavailable.LaunchConfigurationStatusAbnormal"
//	RESOURCEUNAVAILABLE_PROJECTINCONSISTENT = "ResourceUnavailable.ProjectInconsistent"
func (c *Client) ModifyAutoScalingGroupWithContext(ctx context.Context, request *ModifyAutoScalingGroupRequest) (response *ModifyAutoScalingGroupResponse, err error) {
	if request == nil {
		request = NewModifyAutoScalingGroupRequest()
	}
	request.SetContext(ctx)

	response = NewModifyAutoScalingGroupResponse()
	err = c.Send(request, response)
	return
}

func NewModifyDesiredCapacityRequest() (request *ModifyDesiredCapacityRequest) {
	request = &ModifyDesiredCapacityRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ModifyDesiredCapacity")

	return
}

func NewModifyDesiredCapacityResponse() (response *ModifyDesiredCapacityResponse) {
	response = &ModifyDesiredCapacityResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyDesiredCapacity
// 本接口（ModifyDesiredCapacity）用于修改指定伸缩组的期望实例数
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPABNORMALSTATUS = "ResourceUnavailable.AutoScalingGroupAbnormalStatus"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPDISABLED = "ResourceUnavailable.AutoScalingGroupDisabled"
func (c *Client) ModifyDesiredCapacity(request *ModifyDesiredCapacityRequest) (response *ModifyDesiredCapacityResponse, err error) {
	if request == nil {
		request = NewModifyDesiredCapacityRequest()
	}

	response = NewModifyDesiredCapacityResponse()
	err = c.Send(request, response)
	return
}

// ModifyDesiredCapacity
// 本接口（ModifyDesiredCapacity）用于修改指定伸缩组的期望实例数
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPABNORMALSTATUS = "ResourceUnavailable.AutoScalingGroupAbnormalStatus"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPDISABLED = "ResourceUnavailable.AutoScalingGroupDisabled"
func (c *Client) ModifyDesiredCapacityWithContext(ctx context.Context, request *ModifyDesiredCapacityRequest) (response *ModifyDesiredCapacityResponse, err error) {
	if request == nil {
		request = NewModifyDesiredCapacityRequest()
	}
	request.SetContext(ctx)

	response = NewModifyDesiredCapacityResponse()
	err = c.Send(request, response)
	return
}

func NewModifyLaunchConfigurationAttributesRequest() (request *ModifyLaunchConfigurationAttributesRequest) {
	request = &ModifyLaunchConfigurationAttributesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ModifyLaunchConfigurationAttributes")

	return
}

func NewModifyLaunchConfigurationAttributesResponse() (response *ModifyLaunchConfigurationAttributesResponse) {
	response = &ModifyLaunchConfigurationAttributesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyLaunchConfigurationAttributes
// 本接口（ModifyLaunchConfigurationAttributes）用于修改启动配置部分属性。
//
// * 修改启动配置后，已经使用该启动配置扩容的存量实例不会发生变更，此后使用该启动配置的新增实例会按照新的配置进行扩容。
//
// * 本接口支持修改部分简单类型。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_HOSTNAMEUNAVAILABLE = "InvalidParameter.HostNameUnavailable"
//	INVALIDPARAMETER_INVALIDCOMBINATION = "InvalidParameter.InvalidCombination"
//	INVALIDPARAMETER_PARAMETERMUSTBEDELETED = "InvalidParameter.ParameterMustBeDeleted"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CVMCONFIGURATIONERROR = "InvalidParameterValue.CvmConfigurationError"
//	INVALIDPARAMETERVALUE_HOSTNAMEILLEGAL = "InvalidParameterValue.HostNameIllegal"
//	INVALIDPARAMETERVALUE_IMAGENOTFOUND = "InvalidParameterValue.ImageNotFound"
//	INVALIDPARAMETERVALUE_INSTANCENAMEILLEGAL = "InvalidParameterValue.InstanceNameIllegal"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTED = "InvalidParameterValue.InstanceTypeNotSupported"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCETYPE = "InvalidParameterValue.InvalidInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"
//	INVALIDPARAMETERVALUE_NOTSTRINGTYPEFLOAT = "InvalidParameterValue.NotStringTypeFloat"
//	INVALIDPARAMETERVALUE_USERDATAFORMATERROR = "InvalidParameterValue.UserDataFormatError"
//	INVALIDPARAMETERVALUE_USERDATASIZEEXCEEDED = "InvalidParameterValue.UserDataSizeExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_LAUNCHCONFIGURATIONIDNOTFOUND = "ResourceNotFound.LaunchConfigurationIdNotFound"
func (c *Client) ModifyLaunchConfigurationAttributes(request *ModifyLaunchConfigurationAttributesRequest) (response *ModifyLaunchConfigurationAttributesResponse, err error) {
	if request == nil {
		request = NewModifyLaunchConfigurationAttributesRequest()
	}

	response = NewModifyLaunchConfigurationAttributesResponse()
	err = c.Send(request, response)
	return
}

// ModifyLaunchConfigurationAttributes
// 本接口（ModifyLaunchConfigurationAttributes）用于修改启动配置部分属性。
//
// * 修改启动配置后，已经使用该启动配置扩容的存量实例不会发生变更，此后使用该启动配置的新增实例会按照新的配置进行扩容。
//
// * 本接口支持修改部分简单类型。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_HOSTNAMEUNAVAILABLE = "InvalidParameter.HostNameUnavailable"
//	INVALIDPARAMETER_INVALIDCOMBINATION = "InvalidParameter.InvalidCombination"
//	INVALIDPARAMETER_PARAMETERMUSTBEDELETED = "InvalidParameter.ParameterMustBeDeleted"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CVMCONFIGURATIONERROR = "InvalidParameterValue.CvmConfigurationError"
//	INVALIDPARAMETERVALUE_HOSTNAMEILLEGAL = "InvalidParameterValue.HostNameIllegal"
//	INVALIDPARAMETERVALUE_IMAGENOTFOUND = "InvalidParameterValue.ImageNotFound"
//	INVALIDPARAMETERVALUE_INSTANCENAMEILLEGAL = "InvalidParameterValue.InstanceNameIllegal"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTED = "InvalidParameterValue.InstanceTypeNotSupported"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCETYPE = "InvalidParameterValue.InvalidInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"
//	INVALIDPARAMETERVALUE_NOTSTRINGTYPEFLOAT = "InvalidParameterValue.NotStringTypeFloat"
//	INVALIDPARAMETERVALUE_USERDATAFORMATERROR = "InvalidParameterValue.UserDataFormatError"
//	INVALIDPARAMETERVALUE_USERDATASIZEEXCEEDED = "InvalidParameterValue.UserDataSizeExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_LAUNCHCONFIGURATIONIDNOTFOUND = "ResourceNotFound.LaunchConfigurationIdNotFound"
func (c *Client) ModifyLaunchConfigurationAttributesWithContext(ctx context.Context, request *ModifyLaunchConfigurationAttributesRequest) (response *ModifyLaunchConfigurationAttributesResponse, err error) {
	if request == nil {
		request = NewModifyLaunchConfigurationAttributesRequest()
	}
	request.SetContext(ctx)

	response = NewModifyLaunchConfigurationAttributesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyLoadBalancerTargetAttributesRequest() (request *ModifyLoadBalancerTargetAttributesRequest) {
	request = &ModifyLoadBalancerTargetAttributesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ModifyLoadBalancerTargetAttributes")

	return
}

func NewModifyLoadBalancerTargetAttributesResponse() (response *ModifyLoadBalancerTargetAttributesResponse) {
	response = &ModifyLoadBalancerTargetAttributesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyLoadBalancerTargetAttributes
// 本接口（ModifyLoadBalancerTargetAttributes）用于修改伸缩组内负载均衡器的目标规则属性。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETER_LOADBALANCERNOTINAUTOSCALINGGROUP = "InvalidParameter.LoadBalancerNotInAutoScalingGroup"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	INVALIDPARAMETERVALUE_TARGETPORTDUPLICATED = "InvalidParameterValue.TargetPortDuplicated"
//	LIMITEXCEEDED_AFTERATTACHLBLIMITEXCEEDED = "LimitExceeded.AfterAttachLbLimitExceeded"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBPROJECTINCONSISTENT = "ResourceUnavailable.LbProjectInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) ModifyLoadBalancerTargetAttributes(request *ModifyLoadBalancerTargetAttributesRequest) (response *ModifyLoadBalancerTargetAttributesResponse, err error) {
	if request == nil {
		request = NewModifyLoadBalancerTargetAttributesRequest()
	}

	response = NewModifyLoadBalancerTargetAttributesResponse()
	err = c.Send(request, response)
	return
}

// ModifyLoadBalancerTargetAttributes
// 本接口（ModifyLoadBalancerTargetAttributes）用于修改伸缩组内负载均衡器的目标规则属性。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETER_LOADBALANCERNOTINAUTOSCALINGGROUP = "InvalidParameter.LoadBalancerNotInAutoScalingGroup"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	INVALIDPARAMETERVALUE_TARGETPORTDUPLICATED = "InvalidParameterValue.TargetPortDuplicated"
//	LIMITEXCEEDED_AFTERATTACHLBLIMITEXCEEDED = "LimitExceeded.AfterAttachLbLimitExceeded"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBPROJECTINCONSISTENT = "ResourceUnavailable.LbProjectInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) ModifyLoadBalancerTargetAttributesWithContext(ctx context.Context, request *ModifyLoadBalancerTargetAttributesRequest) (response *ModifyLoadBalancerTargetAttributesResponse, err error) {
	if request == nil {
		request = NewModifyLoadBalancerTargetAttributesRequest()
	}
	request.SetContext(ctx)

	response = NewModifyLoadBalancerTargetAttributesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyLoadBalancersRequest() (request *ModifyLoadBalancersRequest) {
	request = &ModifyLoadBalancersRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ModifyLoadBalancers")

	return
}

func NewModifyLoadBalancersResponse() (response *ModifyLoadBalancersResponse) {
	response = &ModifyLoadBalancersResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyLoadBalancers
// 本接口（ModifyLoadBalancers）用于修改伸缩组的负载均衡器。
//
// * 本接口用于为伸缩组指定新的负载均衡器配置，采用`完全覆盖`风格，无论之前配置如何，`统一按照接口参数配置为新的负载均衡器`。
//
// * 如果要为伸缩组清空负载均衡器，则在调用本接口时仅指定伸缩组ID，不指定具体负载均衡器。
//
// * 本接口会立即修改伸缩组的负载均衡器，并生成一个伸缩活动，异步修改存量实例的负载均衡器。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	INVALIDPARAMETERVALUE_LBPROJECTINCONSISTENT = "InvalidParameterValue.LbProjectInconsistent"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) ModifyLoadBalancers(request *ModifyLoadBalancersRequest) (response *ModifyLoadBalancersResponse, err error) {
	if request == nil {
		request = NewModifyLoadBalancersRequest()
	}

	response = NewModifyLoadBalancersResponse()
	err = c.Send(request, response)
	return
}

// ModifyLoadBalancers
// 本接口（ModifyLoadBalancers）用于修改伸缩组的负载均衡器。
//
// * 本接口用于为伸缩组指定新的负载均衡器配置，采用`完全覆盖`风格，无论之前配置如何，`统一按照接口参数配置为新的负载均衡器`。
//
// * 如果要为伸缩组清空负载均衡器，则在调用本接口时仅指定伸缩组ID，不指定具体负载均衡器。
//
// * 本接口会立即修改伸缩组的负载均衡器，并生成一个伸缩活动，异步修改存量实例的负载均衡器。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"
//	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"
//	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"
//	INVALIDPARAMETERVALUE_LBPROJECTINCONSISTENT = "InvalidParameterValue.LbProjectInconsistent"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"
//	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"
//	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"
//	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) ModifyLoadBalancersWithContext(ctx context.Context, request *ModifyLoadBalancersRequest) (response *ModifyLoadBalancersResponse, err error) {
	if request == nil {
		request = NewModifyLoadBalancersRequest()
	}
	request.SetContext(ctx)

	response = NewModifyLoadBalancersResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNotificationConfigurationRequest() (request *ModifyNotificationConfigurationRequest) {
	request = &ModifyNotificationConfigurationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ModifyNotificationConfiguration")

	return
}

func NewModifyNotificationConfigurationResponse() (response *ModifyNotificationConfigurationResponse) {
	response = &ModifyNotificationConfigurationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNotificationConfiguration
// 本接口（ModifyNotificationConfiguration）用于修改通知。
//
// * 通知的接收端类型不支持修改。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_CONFLICTNOTIFICATIONTARGET = "InvalidParameterValue.ConflictNotificationTarget"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGNOTIFICATIONID = "InvalidParameterValue.InvalidAutoScalingNotificationId"
//	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"
//	INVALIDPARAMETERVALUE_USERGROUPIDNOTFOUND = "InvalidParameterValue.UserGroupIdNotFound"
//	RESOURCENOTFOUND_AUTOSCALINGNOTIFICATIONNOTFOUND = "ResourceNotFound.AutoScalingNotificationNotFound"
//	RESOURCENOTFOUND_CMQQUEUENOTFOUND = "ResourceNotFound.CmqQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQQUEUENOTFOUND = "ResourceNotFound.TDMQCMQQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQTOPICNOTFOUND = "ResourceNotFound.TDMQCMQTopicNotFound"
//	RESOURCEUNAVAILABLE_TDMQCMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.TDMQCMQTopicHasNoSubscriber"
func (c *Client) ModifyNotificationConfiguration(request *ModifyNotificationConfigurationRequest) (response *ModifyNotificationConfigurationResponse, err error) {
	if request == nil {
		request = NewModifyNotificationConfigurationRequest()
	}

	response = NewModifyNotificationConfigurationResponse()
	err = c.Send(request, response)
	return
}

// ModifyNotificationConfiguration
// 本接口（ModifyNotificationConfiguration）用于修改通知。
//
// * 通知的接收端类型不支持修改。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_CONFLICTNOTIFICATIONTARGET = "InvalidParameterValue.ConflictNotificationTarget"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGNOTIFICATIONID = "InvalidParameterValue.InvalidAutoScalingNotificationId"
//	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"
//	INVALIDPARAMETERVALUE_USERGROUPIDNOTFOUND = "InvalidParameterValue.UserGroupIdNotFound"
//	RESOURCENOTFOUND_AUTOSCALINGNOTIFICATIONNOTFOUND = "ResourceNotFound.AutoScalingNotificationNotFound"
//	RESOURCENOTFOUND_CMQQUEUENOTFOUND = "ResourceNotFound.CmqQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQQUEUENOTFOUND = "ResourceNotFound.TDMQCMQQueueNotFound"
//	RESOURCENOTFOUND_TDMQCMQTOPICNOTFOUND = "ResourceNotFound.TDMQCMQTopicNotFound"
//	RESOURCEUNAVAILABLE_TDMQCMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.TDMQCMQTopicHasNoSubscriber"
func (c *Client) ModifyNotificationConfigurationWithContext(ctx context.Context, request *ModifyNotificationConfigurationRequest) (response *ModifyNotificationConfigurationResponse, err error) {
	if request == nil {
		request = NewModifyNotificationConfigurationRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNotificationConfigurationResponse()
	err = c.Send(request, response)
	return
}

func NewModifyScalingPolicyRequest() (request *ModifyScalingPolicyRequest) {
	request = &ModifyScalingPolicyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ModifyScalingPolicy")

	return
}

func NewModifyScalingPolicyResponse() (response *ModifyScalingPolicyResponse) {
	response = &ModifyScalingPolicyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyScalingPolicy
// 本接口（ModifyScalingPolicy）用于修改告警触发策略。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"
//	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SCALINGPOLICYNAMEDUPLICATE = "InvalidParameterValue.ScalingPolicyNameDuplicate"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_SCALINGPOLICYNOTFOUND = "ResourceNotFound.ScalingPolicyNotFound"
func (c *Client) ModifyScalingPolicy(request *ModifyScalingPolicyRequest) (response *ModifyScalingPolicyResponse, err error) {
	if request == nil {
		request = NewModifyScalingPolicyRequest()
	}

	response = NewModifyScalingPolicyResponse()
	err = c.Send(request, response)
	return
}

// ModifyScalingPolicy
// 本接口（ModifyScalingPolicy）用于修改告警触发策略。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"
//	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SCALINGPOLICYNAMEDUPLICATE = "InvalidParameterValue.ScalingPolicyNameDuplicate"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_SCALINGPOLICYNOTFOUND = "ResourceNotFound.ScalingPolicyNotFound"
func (c *Client) ModifyScalingPolicyWithContext(ctx context.Context, request *ModifyScalingPolicyRequest) (response *ModifyScalingPolicyResponse, err error) {
	if request == nil {
		request = NewModifyScalingPolicyRequest()
	}
	request.SetContext(ctx)

	response = NewModifyScalingPolicyResponse()
	err = c.Send(request, response)
	return
}

func NewModifyScheduledActionRequest() (request *ModifyScheduledActionRequest) {
	request = &ModifyScheduledActionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ModifyScheduledAction")

	return
}

func NewModifyScheduledActionResponse() (response *ModifyScheduledActionResponse) {
	response = &ModifyScheduledActionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyScheduledAction
// 本接口（ModifyScheduledAction）用于修改定时任务。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_CRONEXPRESSIONILLEGAL = "InvalidParameterValue.CronExpressionIllegal"
//	INVALIDPARAMETERVALUE_ENDTIMEBEFORESTARTTIME = "InvalidParameterValue.EndTimeBeforeStartTime"
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONID = "InvalidParameterValue.InvalidScheduledActionId"
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONNAMEINCLUDEILLEGALCHAR = "InvalidParameterValue.InvalidScheduledActionNameIncludeIllegalChar"
//	INVALIDPARAMETERVALUE_SCHEDULEDACTIONNAMEDUPLICATE = "InvalidParameterValue.ScheduledActionNameDuplicate"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_STARTTIMEBEFORECURRENTTIME = "InvalidParameterValue.StartTimeBeforeCurrentTime"
//	INVALIDPARAMETERVALUE_TIMEFORMAT = "InvalidParameterValue.TimeFormat"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	LIMITEXCEEDED_SCHEDULEDACTIONLIMITEXCEEDED = "LimitExceeded.ScheduledActionLimitExceeded"
//	RESOURCENOTFOUND_SCHEDULEDACTIONNOTFOUND = "ResourceNotFound.ScheduledActionNotFound"
func (c *Client) ModifyScheduledAction(request *ModifyScheduledActionRequest) (response *ModifyScheduledActionResponse, err error) {
	if request == nil {
		request = NewModifyScheduledActionRequest()
	}

	response = NewModifyScheduledActionResponse()
	err = c.Send(request, response)
	return
}

// ModifyScheduledAction
// 本接口（ModifyScheduledAction）用于修改定时任务。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_CRONEXPRESSIONILLEGAL = "InvalidParameterValue.CronExpressionIllegal"
//	INVALIDPARAMETERVALUE_ENDTIMEBEFORESTARTTIME = "InvalidParameterValue.EndTimeBeforeStartTime"
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONID = "InvalidParameterValue.InvalidScheduledActionId"
//	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONNAMEINCLUDEILLEGALCHAR = "InvalidParameterValue.InvalidScheduledActionNameIncludeIllegalChar"
//	INVALIDPARAMETERVALUE_SCHEDULEDACTIONNAMEDUPLICATE = "InvalidParameterValue.ScheduledActionNameDuplicate"
//	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"
//	INVALIDPARAMETERVALUE_STARTTIMEBEFORECURRENTTIME = "InvalidParameterValue.StartTimeBeforeCurrentTime"
//	INVALIDPARAMETERVALUE_TIMEFORMAT = "InvalidParameterValue.TimeFormat"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"
//	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"
//	LIMITEXCEEDED_SCHEDULEDACTIONLIMITEXCEEDED = "LimitExceeded.ScheduledActionLimitExceeded"
//	RESOURCENOTFOUND_SCHEDULEDACTIONNOTFOUND = "ResourceNotFound.ScheduledActionNotFound"
func (c *Client) ModifyScheduledActionWithContext(ctx context.Context, request *ModifyScheduledActionRequest) (response *ModifyScheduledActionResponse, err error) {
	if request == nil {
		request = NewModifyScheduledActionRequest()
	}
	request.SetContext(ctx)

	response = NewModifyScheduledActionResponse()
	err = c.Send(request, response)
	return
}

func NewPreviewPaiDomainNameRequest() (request *PreviewPaiDomainNameRequest) {
	request = &PreviewPaiDomainNameRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "PreviewPaiDomainName")

	return
}

func NewPreviewPaiDomainNameResponse() (response *PreviewPaiDomainNameResponse) {
	response = &PreviewPaiDomainNameResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// PreviewPaiDomainName
// 本接口（PreviewPaiDomainName）用于预览PAI实例域名。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDPAIDOMAINNAMETYPE = "InvalidParameterValue.InvalidPaiDomainNameType"
func (c *Client) PreviewPaiDomainName(request *PreviewPaiDomainNameRequest) (response *PreviewPaiDomainNameResponse, err error) {
	if request == nil {
		request = NewPreviewPaiDomainNameRequest()
	}

	response = NewPreviewPaiDomainNameResponse()
	err = c.Send(request, response)
	return
}

// PreviewPaiDomainName
// 本接口（PreviewPaiDomainName）用于预览PAI实例域名。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDPAIDOMAINNAMETYPE = "InvalidParameterValue.InvalidPaiDomainNameType"
func (c *Client) PreviewPaiDomainNameWithContext(ctx context.Context, request *PreviewPaiDomainNameRequest) (response *PreviewPaiDomainNameResponse, err error) {
	if request == nil {
		request = NewPreviewPaiDomainNameRequest()
	}
	request.SetContext(ctx)

	response = NewPreviewPaiDomainNameResponse()
	err = c.Send(request, response)
	return
}

func NewRemoveInstancesRequest() (request *RemoveInstancesRequest) {
	request = &RemoveInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "RemoveInstances")

	return
}

func NewRemoveInstancesResponse() (response *RemoveInstancesResponse) {
	response = &RemoveInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RemoveInstances
// 本接口（RemoveInstances）用于从伸缩组删除 CVM 实例。根据当前的产品逻辑，如果实例由弹性伸缩自动创建，则实例会被销毁；如果实例系创建后加入伸缩组的，则会从伸缩组中移除，保留实例。
//
// * 如果删除指定实例后，伸缩组内处于`IN_SERVICE`状态的实例数量小于伸缩组最小值，接口将报错
//
// * 如果伸缩组处于`DISABLED`状态，删除操作不校验`IN_SERVICE`实例数量和最小值的关系
//
// * 对于伸缩组配置的 CLB，实例在离开伸缩组时，AS 会进行解挂载动作
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEBELOWMINSIZE = "ResourceInsufficient.InServiceInstanceBelowMinSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_INSTANCEINOPERATION = "ResourceUnavailable.InstanceInOperation"
func (c *Client) RemoveInstances(request *RemoveInstancesRequest) (response *RemoveInstancesResponse, err error) {
	if request == nil {
		request = NewRemoveInstancesRequest()
	}

	response = NewRemoveInstancesResponse()
	err = c.Send(request, response)
	return
}

// RemoveInstances
// 本接口（RemoveInstances）用于从伸缩组删除 CVM 实例。根据当前的产品逻辑，如果实例由弹性伸缩自动创建，则实例会被销毁；如果实例系创建后加入伸缩组的，则会从伸缩组中移除，保留实例。
//
// * 如果删除指定实例后，伸缩组内处于`IN_SERVICE`状态的实例数量小于伸缩组最小值，接口将报错
//
// * 如果伸缩组处于`DISABLED`状态，删除操作不校验`IN_SERVICE`实例数量和最小值的关系
//
// * 对于伸缩组配置的 CLB，实例在离开伸缩组时，AS 会进行解挂载动作
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEBELOWMINSIZE = "ResourceInsufficient.InServiceInstanceBelowMinSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_INSTANCEINOPERATION = "ResourceUnavailable.InstanceInOperation"
func (c *Client) RemoveInstancesWithContext(ctx context.Context, request *RemoveInstancesRequest) (response *RemoveInstancesResponse, err error) {
	if request == nil {
		request = NewRemoveInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewRemoveInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewScaleInInstancesRequest() (request *ScaleInInstancesRequest) {
	request = &ScaleInInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ScaleInInstances")

	return
}

func NewScaleInInstancesResponse() (response *ScaleInInstancesResponse) {
	response = &ScaleInInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ScaleInInstances
// 为伸缩组指定数量缩容实例，返回缩容活动的 ActivityId。
//
// * 伸缩组需要未处于活动中
//
// * 根据伸缩组的`TerminationPolicies`策略，选择被缩容的实例，可参考[缩容处理](https://cloud.tencent.com/document/product/377/8563)
//
// * 接口只会选择`IN_SERVICE`实例缩容，如果需要缩容其他状态实例，可以使用 [DetachInstances](https://cloud.tencent.com/document/api/377/20436) 或 [RemoveInstances](https://cloud.tencent.com/document/api/377/20431) 接口
//
// * 接口会减少期望实例数，新的期望实例数需要大于等于最小实例数
//
// * 缩容如果失败或者部分成功，最后期望实例数只会扣减实际缩容成功的实例数量
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
func (c *Client) ScaleInInstances(request *ScaleInInstancesRequest) (response *ScaleInInstancesResponse, err error) {
	if request == nil {
		request = NewScaleInInstancesRequest()
	}

	response = NewScaleInInstancesResponse()
	err = c.Send(request, response)
	return
}

// ScaleInInstances
// 为伸缩组指定数量缩容实例，返回缩容活动的 ActivityId。
//
// * 伸缩组需要未处于活动中
//
// * 根据伸缩组的`TerminationPolicies`策略，选择被缩容的实例，可参考[缩容处理](https://cloud.tencent.com/document/product/377/8563)
//
// * 接口只会选择`IN_SERVICE`实例缩容，如果需要缩容其他状态实例，可以使用 [DetachInstances](https://cloud.tencent.com/document/api/377/20436) 或 [RemoveInstances](https://cloud.tencent.com/document/api/377/20431) 接口
//
// * 接口会减少期望实例数，新的期望实例数需要大于等于最小实例数
//
// * 缩容如果失败或者部分成功，最后期望实例数只会扣减实际缩容成功的实例数量
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
func (c *Client) ScaleInInstancesWithContext(ctx context.Context, request *ScaleInInstancesRequest) (response *ScaleInInstancesResponse, err error) {
	if request == nil {
		request = NewScaleInInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewScaleInInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewScaleOutInstancesRequest() (request *ScaleOutInstancesRequest) {
	request = &ScaleOutInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "ScaleOutInstances")

	return
}

func NewScaleOutInstancesResponse() (response *ScaleOutInstancesResponse) {
	response = &ScaleOutInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ScaleOutInstances
// 为伸缩组指定数量扩容实例，返回扩容活动的 ActivityId。
//
// * 伸缩组需要未处于活动中
//
// * 接口会增加期望实例数，新的期望实例数需要小于等于最大实例数
//
// * 扩容如果失败或者部分成功，最后期望实例数只会增加实际成功的实例数量
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPABOVEMAXSIZE = "ResourceInsufficient.AutoScalingGroupAboveMaxSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
func (c *Client) ScaleOutInstances(request *ScaleOutInstancesRequest) (response *ScaleOutInstancesResponse, err error) {
	if request == nil {
		request = NewScaleOutInstancesRequest()
	}

	response = NewScaleOutInstancesResponse()
	err = c.Send(request, response)
	return
}

// ScaleOutInstances
// 为伸缩组指定数量扩容实例，返回扩容活动的 ActivityId。
//
// * 伸缩组需要未处于活动中
//
// * 接口会增加期望实例数，新的期望实例数需要小于等于最大实例数
//
// * 扩容如果失败或者部分成功，最后期望实例数只会增加实际成功的实例数量
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPABOVEMAXSIZE = "ResourceInsufficient.AutoScalingGroupAboveMaxSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
func (c *Client) ScaleOutInstancesWithContext(ctx context.Context, request *ScaleOutInstancesRequest) (response *ScaleOutInstancesResponse, err error) {
	if request == nil {
		request = NewScaleOutInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewScaleOutInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewSetInstancesProtectionRequest() (request *SetInstancesProtectionRequest) {
	request = &SetInstancesProtectionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "SetInstancesProtection")

	return
}

func NewSetInstancesProtectionResponse() (response *SetInstancesProtectionResponse) {
	response = &SetInstancesProtectionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// SetInstancesProtection
// 本接口（SetInstancesProtection）用于设置实例保护。
//
// 实例设置保护之后，当发生不健康替换、报警策略、期望值变更等触发缩容时，将不对此实例缩容操作。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
func (c *Client) SetInstancesProtection(request *SetInstancesProtectionRequest) (response *SetInstancesProtectionResponse, err error) {
	if request == nil {
		request = NewSetInstancesProtectionRequest()
	}

	response = NewSetInstancesProtectionResponse()
	err = c.Send(request, response)
	return
}

// SetInstancesProtection
// 本接口（SetInstancesProtection）用于设置实例保护。
//
// 实例设置保护之后，当发生不健康替换、报警策略、期望值变更等触发缩容时，将不对此实例缩容操作。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
func (c *Client) SetInstancesProtectionWithContext(ctx context.Context, request *SetInstancesProtectionRequest) (response *SetInstancesProtectionResponse, err error) {
	if request == nil {
		request = NewSetInstancesProtectionRequest()
	}
	request.SetContext(ctx)

	response = NewSetInstancesProtectionResponse()
	err = c.Send(request, response)
	return
}

func NewStartAutoScalingInstancesRequest() (request *StartAutoScalingInstancesRequest) {
	request = &StartAutoScalingInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "StartAutoScalingInstances")

	return
}

func NewStartAutoScalingInstancesResponse() (response *StartAutoScalingInstancesResponse) {
	response = &StartAutoScalingInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// StartAutoScalingInstances
// 本接口（StartAutoScalingInstances）用于开启伸缩组内 CVM 实例。
//
// * 开机成功，实例转为`IN_SERVICE`状态后，会增加期望实例数，期望实例数不可超过设置的最大值
//
// * 本接口支持批量操作，每次请求开机实例的上限为100
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPABOVEMAXSIZE = "ResourceInsufficient.AutoScalingGroupAboveMaxSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEABOVEMAXSIZE = "ResourceInsufficient.InServiceInstanceAboveMaxSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) StartAutoScalingInstances(request *StartAutoScalingInstancesRequest) (response *StartAutoScalingInstancesResponse, err error) {
	if request == nil {
		request = NewStartAutoScalingInstancesRequest()
	}

	response = NewStartAutoScalingInstancesResponse()
	err = c.Send(request, response)
	return
}

// StartAutoScalingInstances
// 本接口（StartAutoScalingInstances）用于开启伸缩组内 CVM 实例。
//
// * 开机成功，实例转为`IN_SERVICE`状态后，会增加期望实例数，期望实例数不可超过设置的最大值
//
// * 本接口支持批量操作，每次请求开机实例的上限为100
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPABOVEMAXSIZE = "ResourceInsufficient.AutoScalingGroupAboveMaxSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEABOVEMAXSIZE = "ResourceInsufficient.InServiceInstanceAboveMaxSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) StartAutoScalingInstancesWithContext(ctx context.Context, request *StartAutoScalingInstancesRequest) (response *StartAutoScalingInstancesResponse, err error) {
	if request == nil {
		request = NewStartAutoScalingInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewStartAutoScalingInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewStopAutoScalingInstancesRequest() (request *StopAutoScalingInstancesRequest) {
	request = &StopAutoScalingInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "StopAutoScalingInstances")

	return
}

func NewStopAutoScalingInstancesResponse() (response *StopAutoScalingInstancesResponse) {
	response = &StopAutoScalingInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// StopAutoScalingInstances
// 本接口（StopAutoScalingInstances）用于关闭伸缩组内 CVM 实例。
//
// * 关机方式采用`SOFT_FIRST`方式，表示在正常关闭失败后进行强制关闭
//
// * 关闭`IN_SERVICE`状态的实例，会减少期望实例数，期望实例数不可低于设置的最小值
//
// * 使用`STOP_CHARGING`选项关机，待关机的实例需要满足[关机不收费条件](https://cloud.tencent.com/document/product/213/19918)
//
// * 本接口支持批量操作，每次请求关机实例的上限为100
//
// 可能返回的错误码:
//
//	CALLCVMERROR = "CallCvmError"
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEBELOWMINSIZE = "ResourceInsufficient.InServiceInstanceBelowMinSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_INSTANCEINOPERATION = "ResourceUnavailable.InstanceInOperation"
//	RESOURCEUNAVAILABLE_INSTANCENOTSUPPORTSTOPCHARGING = "ResourceUnavailable.InstanceNotSupportStopCharging"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) StopAutoScalingInstances(request *StopAutoScalingInstancesRequest) (response *StopAutoScalingInstancesResponse, err error) {
	if request == nil {
		request = NewStopAutoScalingInstancesRequest()
	}

	response = NewStopAutoScalingInstancesResponse()
	err = c.Send(request, response)
	return
}

// StopAutoScalingInstances
// 本接口（StopAutoScalingInstances）用于关闭伸缩组内 CVM 实例。
//
// * 关机方式采用`SOFT_FIRST`方式，表示在正常关闭失败后进行强制关闭
//
// * 关闭`IN_SERVICE`状态的实例，会减少期望实例数，期望实例数不可低于设置的最小值
//
// * 使用`STOP_CHARGING`选项关机，待关机的实例需要满足[关机不收费条件](https://cloud.tencent.com/document/product/213/19918)
//
// * 本接口支持批量操作，每次请求关机实例的上限为100
//
// 可能返回的错误码:
//
//	CALLCVMERROR = "CallCvmError"
//	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"
//	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"
//	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"
//	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"
//	RESOURCEINSUFFICIENT_INSERVICEINSTANCEBELOWMINSIZE = "ResourceInsufficient.InServiceInstanceBelowMinSize"
//	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"
//	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"
//	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"
//	RESOURCEUNAVAILABLE_INSTANCEINOPERATION = "ResourceUnavailable.InstanceInOperation"
//	RESOURCEUNAVAILABLE_INSTANCENOTSUPPORTSTOPCHARGING = "ResourceUnavailable.InstanceNotSupportStopCharging"
//	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"
func (c *Client) StopAutoScalingInstancesWithContext(ctx context.Context, request *StopAutoScalingInstancesRequest) (response *StopAutoScalingInstancesResponse, err error) {
	if request == nil {
		request = NewStopAutoScalingInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewStopAutoScalingInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewUpgradeLaunchConfigurationRequest() (request *UpgradeLaunchConfigurationRequest) {
	request = &UpgradeLaunchConfigurationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "UpgradeLaunchConfiguration")

	return
}

func NewUpgradeLaunchConfigurationResponse() (response *UpgradeLaunchConfigurationResponse) {
	response = &UpgradeLaunchConfigurationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UpgradeLaunchConfiguration
// 本接口（UpgradeLaunchConfiguration）用于升级启动配置。
//
// * 本接口用于升级启动配置，采用“完全覆盖”风格，无论之前参数如何，统一按照接口参数设置为新的配置。对于非必填字段，不填写则按照默认值赋值。
//
// * 升级修改启动配置后，已经使用该启动配置扩容的存量实例不会发生变更，此后使用该启动配置的新增实例会按照新的配置进行扩容。
//
// 可能返回的错误码:
//
//	CALLCVMERROR = "CallCvmError"
//	INTERNALERROR = "InternalError"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INVALIDCOMBINATION = "InvalidParameter.InvalidCombination"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETER_PARAMETERMUSTBEDELETED = "InvalidParameter.ParameterMustBeDeleted"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CVMCONFIGURATIONERROR = "InvalidParameterValue.CvmConfigurationError"
//	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"
//	INVALIDPARAMETERVALUE_HOSTNAMEILLEGAL = "InvalidParameterValue.HostNameIllegal"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTED = "InvalidParameterValue.InstanceTypeNotSupported"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCETYPE = "InvalidParameterValue.InvalidInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"
//	INVALIDPARAMETERVALUE_NOTSTRINGTYPEFLOAT = "InvalidParameterValue.NotStringTypeFloat"
//	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_USERDATAFORMATERROR = "InvalidParameterValue.UserDataFormatError"
//	INVALIDPARAMETERVALUE_USERDATASIZEEXCEEDED = "InvalidParameterValue.UserDataSizeExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_LAUNCHCONFIGURATIONIDNOTFOUND = "ResourceNotFound.LaunchConfigurationIdNotFound"
func (c *Client) UpgradeLaunchConfiguration(request *UpgradeLaunchConfigurationRequest) (response *UpgradeLaunchConfigurationResponse, err error) {
	if request == nil {
		request = NewUpgradeLaunchConfigurationRequest()
	}

	response = NewUpgradeLaunchConfigurationResponse()
	err = c.Send(request, response)
	return
}

// UpgradeLaunchConfiguration
// 本接口（UpgradeLaunchConfiguration）用于升级启动配置。
//
// * 本接口用于升级启动配置，采用“完全覆盖”风格，无论之前参数如何，统一按照接口参数设置为新的配置。对于非必填字段，不填写则按照默认值赋值。
//
// * 升级修改启动配置后，已经使用该启动配置扩容的存量实例不会发生变更，此后使用该启动配置的新增实例会按照新的配置进行扩容。
//
// 可能返回的错误码:
//
//	CALLCVMERROR = "CallCvmError"
//	INTERNALERROR = "InternalError"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETER_INVALIDCOMBINATION = "InvalidParameter.InvalidCombination"
//	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"
//	INVALIDPARAMETER_PARAMETERMUSTBEDELETED = "InvalidParameter.ParameterMustBeDeleted"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CVMCONFIGURATIONERROR = "InvalidParameterValue.CvmConfigurationError"
//	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"
//	INVALIDPARAMETERVALUE_HOSTNAMEILLEGAL = "InvalidParameterValue.HostNameIllegal"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTED = "InvalidParameterValue.InstanceTypeNotSupported"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCETYPE = "InvalidParameterValue.InvalidInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"
//	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"
//	INVALIDPARAMETERVALUE_NOTSTRINGTYPEFLOAT = "InvalidParameterValue.NotStringTypeFloat"
//	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_USERDATAFORMATERROR = "InvalidParameterValue.UserDataFormatError"
//	INVALIDPARAMETERVALUE_USERDATASIZEEXCEEDED = "InvalidParameterValue.UserDataSizeExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_LAUNCHCONFIGURATIONIDNOTFOUND = "ResourceNotFound.LaunchConfigurationIdNotFound"
func (c *Client) UpgradeLaunchConfigurationWithContext(ctx context.Context, request *UpgradeLaunchConfigurationRequest) (response *UpgradeLaunchConfigurationResponse, err error) {
	if request == nil {
		request = NewUpgradeLaunchConfigurationRequest()
	}
	request.SetContext(ctx)

	response = NewUpgradeLaunchConfigurationResponse()
	err = c.Send(request, response)
	return
}

func NewUpgradeLifecycleHookRequest() (request *UpgradeLifecycleHookRequest) {
	request = &UpgradeLifecycleHookRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("as", APIVersion, "UpgradeLifecycleHook")

	return
}

func NewUpgradeLifecycleHookResponse() (response *UpgradeLifecycleHookResponse) {
	response = &UpgradeLifecycleHookResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UpgradeLifecycleHook
// 本接口（UpgradeLifecycleHook）用于升级生命周期挂钩。
//
// * 本接口用于升级生命周期挂钩，采用“完全覆盖”风格，无论之前参数如何，统一按照接口参数设置为新的配置。对于非必填字段，不填写则按照默认值赋值。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLCMQERROR = "InternalError.CallCmqError"
//	INTERNALERROR_CALLSTSERROR = "InternalError.CallStsError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"
//	INVALIDPARAMETERVALUE_LIFECYCLEHOOKNAMEDUPLICATED = "InvalidParameterValue.LifecycleHookNameDuplicated"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_LIFECYCLEHOOKNOTFOUND = "ResourceNotFound.LifecycleHookNotFound"
//	RESOURCEUNAVAILABLE_CMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.CmqTopicHasNoSubscriber"
func (c *Client) UpgradeLifecycleHook(request *UpgradeLifecycleHookRequest) (response *UpgradeLifecycleHookResponse, err error) {
	if request == nil {
		request = NewUpgradeLifecycleHookRequest()
	}

	response = NewUpgradeLifecycleHookResponse()
	err = c.Send(request, response)
	return
}

// UpgradeLifecycleHook
// 本接口（UpgradeLifecycleHook）用于升级生命周期挂钩。
//
// * 本接口用于升级生命周期挂钩，采用“完全覆盖”风格，无论之前参数如何，统一按照接口参数设置为新的配置。对于非必填字段，不填写则按照默认值赋值。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CALLCMQERROR = "InternalError.CallCmqError"
//	INTERNALERROR_CALLSTSERROR = "InternalError.CallStsError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"
//	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"
//	INVALIDPARAMETERVALUE_LIFECYCLEHOOKNAMEDUPLICATED = "InvalidParameterValue.LifecycleHookNameDuplicated"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND_LIFECYCLEHOOKNOTFOUND = "ResourceNotFound.LifecycleHookNotFound"
//	RESOURCEUNAVAILABLE_CMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.CmqTopicHasNoSubscriber"
func (c *Client) UpgradeLifecycleHookWithContext(ctx context.Context, request *UpgradeLifecycleHookRequest) (response *UpgradeLifecycleHookResponse, err error) {
	if request == nil {
		request = NewUpgradeLifecycleHookRequest()
	}
	request.SetContext(ctx)

	response = NewUpgradeLifecycleHookResponse()
	err = c.Send(request, response)
	return
}
