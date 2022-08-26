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

package v20170312

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = "2017-03-12"

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

func NewAcceptAttachCcnInstancesRequest() (request *AcceptAttachCcnInstancesRequest) {
	request = &AcceptAttachCcnInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AcceptAttachCcnInstances")

	return
}

func NewAcceptAttachCcnInstancesResponse() (response *AcceptAttachCcnInstancesResponse) {
	response = &AcceptAttachCcnInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AcceptAttachCcnInstances
// 本接口（AcceptAttachCcnInstances）用于跨账号关联实例时，云联网所有者接受并同意关联操作。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CCNNOTATTACHED = "UnsupportedOperation.CcnNotAttached"
//	UNSUPPORTEDOPERATION_INVALIDINSTANCESTATE = "UnsupportedOperation.InvalidInstanceState"
//	UNSUPPORTEDOPERATION_ISNOTFINANCEACCOUNT = "UnsupportedOperation.IsNotFinanceAccount"
//	UNSUPPORTEDOPERATION_NOTPENDINGCCNINSTANCE = "UnsupportedOperation.NotPendingCcnInstance"
//	UNSUPPORTEDOPERATION_UNABLECROSSFINANCE = "UnsupportedOperation.UnableCrossFinance"
func (c *Client) AcceptAttachCcnInstances(request *AcceptAttachCcnInstancesRequest) (response *AcceptAttachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewAcceptAttachCcnInstancesRequest()
	}

	response = NewAcceptAttachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

// AcceptAttachCcnInstances
// 本接口（AcceptAttachCcnInstances）用于跨账号关联实例时，云联网所有者接受并同意关联操作。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CCNNOTATTACHED = "UnsupportedOperation.CcnNotAttached"
//	UNSUPPORTEDOPERATION_INVALIDINSTANCESTATE = "UnsupportedOperation.InvalidInstanceState"
//	UNSUPPORTEDOPERATION_ISNOTFINANCEACCOUNT = "UnsupportedOperation.IsNotFinanceAccount"
//	UNSUPPORTEDOPERATION_NOTPENDINGCCNINSTANCE = "UnsupportedOperation.NotPendingCcnInstance"
//	UNSUPPORTEDOPERATION_UNABLECROSSFINANCE = "UnsupportedOperation.UnableCrossFinance"
func (c *Client) AcceptAttachCcnInstancesWithContext(ctx context.Context, request *AcceptAttachCcnInstancesRequest) (response *AcceptAttachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewAcceptAttachCcnInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewAcceptAttachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewAddBandwidthPackageResourcesRequest() (request *AddBandwidthPackageResourcesRequest) {
	request = &AddBandwidthPackageResourcesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AddBandwidthPackageResources")

	return
}

func NewAddBandwidthPackageResourcesResponse() (response *AddBandwidthPackageResourcesResponse) {
	response = &AddBandwidthPackageResourcesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AddBandwidthPackageResources
// 接口用于添加带宽包资源，包括[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)和[负载均衡](https://cloud.tencent.com/document/product/214/517)等
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RESOURCEALREADYEXISTED = "InvalidParameterValue.ResourceAlreadyExisted"
//	INVALIDPARAMETERVALUE_RESOURCEIDMALFORMED = "InvalidParameterValue.ResourceIdMalformed"
//	INVALIDPARAMETERVALUE_RESOURCENOTFOUND = "InvalidParameterValue.ResourceNotFound"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDRESOURCEINTERNETCHARGETYPE = "UnsupportedOperation.InvalidResourceInternetChargeType"
//	UNSUPPORTEDOPERATION_INVALIDRESOURCEPROTOCOL = "UnsupportedOperation.InvalidResourceProtocol"
func (c *Client) AddBandwidthPackageResources(request *AddBandwidthPackageResourcesRequest) (response *AddBandwidthPackageResourcesResponse, err error) {
	if request == nil {
		request = NewAddBandwidthPackageResourcesRequest()
	}

	response = NewAddBandwidthPackageResourcesResponse()
	err = c.Send(request, response)
	return
}

// AddBandwidthPackageResources
// 接口用于添加带宽包资源，包括[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)和[负载均衡](https://cloud.tencent.com/document/product/214/517)等
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RESOURCEALREADYEXISTED = "InvalidParameterValue.ResourceAlreadyExisted"
//	INVALIDPARAMETERVALUE_RESOURCEIDMALFORMED = "InvalidParameterValue.ResourceIdMalformed"
//	INVALIDPARAMETERVALUE_RESOURCENOTFOUND = "InvalidParameterValue.ResourceNotFound"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDRESOURCEINTERNETCHARGETYPE = "UnsupportedOperation.InvalidResourceInternetChargeType"
//	UNSUPPORTEDOPERATION_INVALIDRESOURCEPROTOCOL = "UnsupportedOperation.InvalidResourceProtocol"
func (c *Client) AddBandwidthPackageResourcesWithContext(ctx context.Context, request *AddBandwidthPackageResourcesRequest) (response *AddBandwidthPackageResourcesResponse, err error) {
	if request == nil {
		request = NewAddBandwidthPackageResourcesRequest()
	}
	request.SetContext(ctx)

	response = NewAddBandwidthPackageResourcesResponse()
	err = c.Send(request, response)
	return
}

func NewAddIp6RulesRequest() (request *AddIp6RulesRequest) {
	request = &AddIp6RulesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AddIp6Rules")

	return
}

func NewAddIp6RulesResponse() (response *AddIp6RulesResponse) {
	response = &AddIp6RulesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AddIp6Rules
// 1. 该接口用于在转换实例下添加IPV6转换规则。
//
// 2. 支持在同一个转换实例下批量添加转换规则，一个账户在一个地域最多50个。
//
// 3. 一个完整的转换规则包括vip6:vport6:protocol:vip:vport，其中vip6:vport6:protocol必须是唯一。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) AddIp6Rules(request *AddIp6RulesRequest) (response *AddIp6RulesResponse, err error) {
	if request == nil {
		request = NewAddIp6RulesRequest()
	}

	response = NewAddIp6RulesResponse()
	err = c.Send(request, response)
	return
}

// AddIp6Rules
// 1. 该接口用于在转换实例下添加IPV6转换规则。
//
// 2. 支持在同一个转换实例下批量添加转换规则，一个账户在一个地域最多50个。
//
// 3. 一个完整的转换规则包括vip6:vport6:protocol:vip:vport，其中vip6:vport6:protocol必须是唯一。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) AddIp6RulesWithContext(ctx context.Context, request *AddIp6RulesRequest) (response *AddIp6RulesResponse, err error) {
	if request == nil {
		request = NewAddIp6RulesRequest()
	}
	request.SetContext(ctx)

	response = NewAddIp6RulesResponse()
	err = c.Send(request, response)
	return
}

func NewAddTemplateMemberRequest() (request *AddTemplateMemberRequest) {
	request = &AddTemplateMemberRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AddTemplateMember")

	return
}

func NewAddTemplateMemberResponse() (response *AddTemplateMemberResponse) {
	response = &AddTemplateMemberResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AddTemplateMember
// 增加模版对象中的IP地址、协议端口、IP地址组、协议端口组。当前仅支持北京、泰国、北美地域请求。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) AddTemplateMember(request *AddTemplateMemberRequest) (response *AddTemplateMemberResponse, err error) {
	if request == nil {
		request = NewAddTemplateMemberRequest()
	}

	response = NewAddTemplateMemberResponse()
	err = c.Send(request, response)
	return
}

// AddTemplateMember
// 增加模版对象中的IP地址、协议端口、IP地址组、协议端口组。当前仅支持北京、泰国、北美地域请求。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) AddTemplateMemberWithContext(ctx context.Context, request *AddTemplateMemberRequest) (response *AddTemplateMemberResponse, err error) {
	if request == nil {
		request = NewAddTemplateMemberRequest()
	}
	request.SetContext(ctx)

	response = NewAddTemplateMemberResponse()
	err = c.Send(request, response)
	return
}

func NewAllocateAddressesRequest() (request *AllocateAddressesRequest) {
	request = &AllocateAddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AllocateAddresses")

	return
}

func NewAllocateAddressesResponse() (response *AllocateAddressesResponse) {
	response = &AllocateAddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AllocateAddresses
// 本接口 (AllocateAddresses) 用于申请一个或多个[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）。
//
// * EIP 是专为动态云计算设计的静态 IP 地址。借助 EIP，您可以快速将 EIP 重新映射到您的另一个实例上，从而屏蔽实例故障。
//
// * 您的 EIP 与腾讯云账户相关联，而不是与某个实例相关联。在您选择显式释放该地址，或欠费超过24小时之前，它会一直与您的腾讯云账户保持关联。
//
// * 一个腾讯云账户在每个地域能申请的 EIP 最大配额有所限制，可参见 [EIP 产品简介](https://cloud.tencent.com/document/product/213/5733)，上述配额可通过 DescribeAddressQuota 接口获取。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED = "AddressQuotaLimitExceeded"
//	ADDRESSQUOTALIMITEXCEEDED_DAILYALLOCATE = "AddressQuotaLimitExceeded.DailyAllocate"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTAVAILABLE = "InvalidParameterValue.AddressIpNotAvailable"
//	INVALIDPARAMETERVALUE_BANDWIDTHOUTOFRANGE = "InvalidParameterValue.BandwidthOutOfRange"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_BANDWIDTHTOOSMALL = "InvalidParameterValue.BandwidthTooSmall"
//	INVALIDPARAMETERVALUE_COMBINATION = "InvalidParameterValue.Combination"
//	INVALIDPARAMETERVALUE_INVALIDDEDICATEDCLUSTERID = "InvalidParameterValue.InvalidDedicatedClusterId"
//	INVALIDPARAMETERVALUE_INVALIDTAG = "InvalidParameterValue.InvalidTag"
//	INVALIDPARAMETERVALUE_MIXEDADDRESSIPSETTYPE = "InvalidParameterValue.MixedAddressIpSetType"
//	INVALIDPARAMETERVALUE_RESOURCENOTSUPPORT = "InvalidParameterValue.ResourceNotSupport"
//	INVALIDPARAMETERVALUE_TAGNOTEXISTED = "InvalidParameterValue.TagNotExisted"
//	INVALIDPARAMETERVALUE_UNAVAILABLEZONE = "InvalidParameterValue.UnavailableZone"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
func (c *Client) AllocateAddresses(request *AllocateAddressesRequest) (response *AllocateAddressesResponse, err error) {
	if request == nil {
		request = NewAllocateAddressesRequest()
	}

	response = NewAllocateAddressesResponse()
	err = c.Send(request, response)
	return
}

// AllocateAddresses
// 本接口 (AllocateAddresses) 用于申请一个或多个[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）。
//
// * EIP 是专为动态云计算设计的静态 IP 地址。借助 EIP，您可以快速将 EIP 重新映射到您的另一个实例上，从而屏蔽实例故障。
//
// * 您的 EIP 与腾讯云账户相关联，而不是与某个实例相关联。在您选择显式释放该地址，或欠费超过24小时之前，它会一直与您的腾讯云账户保持关联。
//
// * 一个腾讯云账户在每个地域能申请的 EIP 最大配额有所限制，可参见 [EIP 产品简介](https://cloud.tencent.com/document/product/213/5733)，上述配额可通过 DescribeAddressQuota 接口获取。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED = "AddressQuotaLimitExceeded"
//	ADDRESSQUOTALIMITEXCEEDED_DAILYALLOCATE = "AddressQuotaLimitExceeded.DailyAllocate"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTAVAILABLE = "InvalidParameterValue.AddressIpNotAvailable"
//	INVALIDPARAMETERVALUE_BANDWIDTHOUTOFRANGE = "InvalidParameterValue.BandwidthOutOfRange"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_BANDWIDTHTOOSMALL = "InvalidParameterValue.BandwidthTooSmall"
//	INVALIDPARAMETERVALUE_COMBINATION = "InvalidParameterValue.Combination"
//	INVALIDPARAMETERVALUE_INVALIDDEDICATEDCLUSTERID = "InvalidParameterValue.InvalidDedicatedClusterId"
//	INVALIDPARAMETERVALUE_INVALIDTAG = "InvalidParameterValue.InvalidTag"
//	INVALIDPARAMETERVALUE_MIXEDADDRESSIPSETTYPE = "InvalidParameterValue.MixedAddressIpSetType"
//	INVALIDPARAMETERVALUE_RESOURCENOTSUPPORT = "InvalidParameterValue.ResourceNotSupport"
//	INVALIDPARAMETERVALUE_TAGNOTEXISTED = "InvalidParameterValue.TagNotExisted"
//	INVALIDPARAMETERVALUE_UNAVAILABLEZONE = "InvalidParameterValue.UnavailableZone"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
func (c *Client) AllocateAddressesWithContext(ctx context.Context, request *AllocateAddressesRequest) (response *AllocateAddressesResponse, err error) {
	if request == nil {
		request = NewAllocateAddressesRequest()
	}
	request.SetContext(ctx)

	response = NewAllocateAddressesResponse()
	err = c.Send(request, response)
	return
}

func NewAllocateIp6AddressesBandwidthRequest() (request *AllocateIp6AddressesBandwidthRequest) {
	request = &AllocateIp6AddressesBandwidthRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AllocateIp6AddressesBandwidth")

	return
}

func NewAllocateIp6AddressesBandwidthResponse() (response *AllocateIp6AddressesBandwidthResponse) {
	response = &AllocateIp6AddressesBandwidthResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AllocateIp6AddressesBandwidth
// 该接口用于给IPv6地址初次分配公网带宽
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTFOUND = "InvalidParameterValue.AddressIpNotFound"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTINVPC = "InvalidParameterValue.AddressIpNotInVpc"
//	INVALIDPARAMETERVALUE_ADDRESSPUBLISHED = "InvalidParameterValue.AddressPublished"
//	INVALIDPARAMETERVALUE_INVALIDIPV6 = "InvalidParameterValue.InvalidIpv6"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
func (c *Client) AllocateIp6AddressesBandwidth(request *AllocateIp6AddressesBandwidthRequest) (response *AllocateIp6AddressesBandwidthResponse, err error) {
	if request == nil {
		request = NewAllocateIp6AddressesBandwidthRequest()
	}

	response = NewAllocateIp6AddressesBandwidthResponse()
	err = c.Send(request, response)
	return
}

// AllocateIp6AddressesBandwidth
// 该接口用于给IPv6地址初次分配公网带宽
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTFOUND = "InvalidParameterValue.AddressIpNotFound"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTINVPC = "InvalidParameterValue.AddressIpNotInVpc"
//	INVALIDPARAMETERVALUE_ADDRESSPUBLISHED = "InvalidParameterValue.AddressPublished"
//	INVALIDPARAMETERVALUE_INVALIDIPV6 = "InvalidParameterValue.InvalidIpv6"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
func (c *Client) AllocateIp6AddressesBandwidthWithContext(ctx context.Context, request *AllocateIp6AddressesBandwidthRequest) (response *AllocateIp6AddressesBandwidthResponse, err error) {
	if request == nil {
		request = NewAllocateIp6AddressesBandwidthRequest()
	}
	request.SetContext(ctx)

	response = NewAllocateIp6AddressesBandwidthResponse()
	err = c.Send(request, response)
	return
}

func NewAssignIpv6AddressesRequest() (request *AssignIpv6AddressesRequest) {
	request = &AssignIpv6AddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssignIpv6Addresses")

	return
}

func NewAssignIpv6AddressesResponse() (response *AssignIpv6AddressesResponse) {
	response = &AssignIpv6AddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssignIpv6Addresses
// 本接口（AssignIpv6Addresses）用于弹性网卡申请`IPv6`地址。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// * 一个弹性网卡支持绑定的IP地址是有限制的，更多资源限制信息详见<a href="/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 可以指定`IPv6`地址申请，地址类型不能为主`IP`，`IPv6`地址暂时只支持作为辅助`IP`。
//
// * 地址必须要在弹性网卡所在子网内，而且不能被占用。
//
// * 在弹性网卡上申请一个到多个辅助`IPv6`地址，接口会在弹性网卡所在子网段内返回指定数量的辅助`IPv6`地址。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED_ADDRESS = "LimitExceeded.Address"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE_ADDRESS = "ResourceInUse.Address"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_UNASSIGNCIDRBLOCK = "UnsupportedOperation.UnassignCidrBlock"
func (c *Client) AssignIpv6Addresses(request *AssignIpv6AddressesRequest) (response *AssignIpv6AddressesResponse, err error) {
	if request == nil {
		request = NewAssignIpv6AddressesRequest()
	}

	response = NewAssignIpv6AddressesResponse()
	err = c.Send(request, response)
	return
}

// AssignIpv6Addresses
// 本接口（AssignIpv6Addresses）用于弹性网卡申请`IPv6`地址。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// * 一个弹性网卡支持绑定的IP地址是有限制的，更多资源限制信息详见<a href="/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 可以指定`IPv6`地址申请，地址类型不能为主`IP`，`IPv6`地址暂时只支持作为辅助`IP`。
//
// * 地址必须要在弹性网卡所在子网内，而且不能被占用。
//
// * 在弹性网卡上申请一个到多个辅助`IPv6`地址，接口会在弹性网卡所在子网段内返回指定数量的辅助`IPv6`地址。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED_ADDRESS = "LimitExceeded.Address"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE_ADDRESS = "ResourceInUse.Address"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_UNASSIGNCIDRBLOCK = "UnsupportedOperation.UnassignCidrBlock"
func (c *Client) AssignIpv6AddressesWithContext(ctx context.Context, request *AssignIpv6AddressesRequest) (response *AssignIpv6AddressesResponse, err error) {
	if request == nil {
		request = NewAssignIpv6AddressesRequest()
	}
	request.SetContext(ctx)

	response = NewAssignIpv6AddressesResponse()
	err = c.Send(request, response)
	return
}

func NewAssignIpv6CidrBlockRequest() (request *AssignIpv6CidrBlockRequest) {
	request = &AssignIpv6CidrBlockRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssignIpv6CidrBlock")

	return
}

func NewAssignIpv6CidrBlockResponse() (response *AssignIpv6CidrBlockResponse) {
	response = &AssignIpv6CidrBlockResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssignIpv6CidrBlock
// 本接口（AssignIpv6CidrBlock）用于分配IPv6网段。
//
// * 使用本接口前，您需要已有VPC实例，如果没有可通过接口<a href="https://cloud.tencent.com/document/api/215/15774" title="CreateVpc" target="_blank">CreateVpc</a>创建。
//
// * 每个VPC只能申请一个IPv6网段
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED_CIDRBLOCK = "LimitExceeded.CidrBlock"
//	RESOURCEINSUFFICIENT_CIDRBLOCK = "ResourceInsufficient.CidrBlock"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) AssignIpv6CidrBlock(request *AssignIpv6CidrBlockRequest) (response *AssignIpv6CidrBlockResponse, err error) {
	if request == nil {
		request = NewAssignIpv6CidrBlockRequest()
	}

	response = NewAssignIpv6CidrBlockResponse()
	err = c.Send(request, response)
	return
}

// AssignIpv6CidrBlock
// 本接口（AssignIpv6CidrBlock）用于分配IPv6网段。
//
// * 使用本接口前，您需要已有VPC实例，如果没有可通过接口<a href="https://cloud.tencent.com/document/api/215/15774" title="CreateVpc" target="_blank">CreateVpc</a>创建。
//
// * 每个VPC只能申请一个IPv6网段
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED_CIDRBLOCK = "LimitExceeded.CidrBlock"
//	RESOURCEINSUFFICIENT_CIDRBLOCK = "ResourceInsufficient.CidrBlock"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) AssignIpv6CidrBlockWithContext(ctx context.Context, request *AssignIpv6CidrBlockRequest) (response *AssignIpv6CidrBlockResponse, err error) {
	if request == nil {
		request = NewAssignIpv6CidrBlockRequest()
	}
	request.SetContext(ctx)

	response = NewAssignIpv6CidrBlockResponse()
	err = c.Send(request, response)
	return
}

func NewAssignIpv6SubnetCidrBlockRequest() (request *AssignIpv6SubnetCidrBlockRequest) {
	request = &AssignIpv6SubnetCidrBlockRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssignIpv6SubnetCidrBlock")

	return
}

func NewAssignIpv6SubnetCidrBlockResponse() (response *AssignIpv6SubnetCidrBlockResponse) {
	response = &AssignIpv6SubnetCidrBlockResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssignIpv6SubnetCidrBlock
// 本接口（AssignIpv6SubnetCidrBlock）用于分配IPv6子网段。
//
// * 给子网分配 `IPv6` 网段，要求子网所属 `VPC` 已获得 `IPv6` 网段。如果尚未分配，请先通过接口 `AssignIpv6CidrBlock` 给子网所属 `VPC` 分配一个 `IPv6` 网段。否则无法分配 `IPv6` 子网段。
//
// * 每个子网只能分配一个IPv6网段。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETRANGE = "InvalidParameterValue.SubnetRange"
//	LIMITEXCEEDED_SUBNETCIDRBLOCK = "LimitExceeded.SubnetCidrBlock"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) AssignIpv6SubnetCidrBlock(request *AssignIpv6SubnetCidrBlockRequest) (response *AssignIpv6SubnetCidrBlockResponse, err error) {
	if request == nil {
		request = NewAssignIpv6SubnetCidrBlockRequest()
	}

	response = NewAssignIpv6SubnetCidrBlockResponse()
	err = c.Send(request, response)
	return
}

// AssignIpv6SubnetCidrBlock
// 本接口（AssignIpv6SubnetCidrBlock）用于分配IPv6子网段。
//
// * 给子网分配 `IPv6` 网段，要求子网所属 `VPC` 已获得 `IPv6` 网段。如果尚未分配，请先通过接口 `AssignIpv6CidrBlock` 给子网所属 `VPC` 分配一个 `IPv6` 网段。否则无法分配 `IPv6` 子网段。
//
// * 每个子网只能分配一个IPv6网段。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETRANGE = "InvalidParameterValue.SubnetRange"
//	LIMITEXCEEDED_SUBNETCIDRBLOCK = "LimitExceeded.SubnetCidrBlock"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) AssignIpv6SubnetCidrBlockWithContext(ctx context.Context, request *AssignIpv6SubnetCidrBlockRequest) (response *AssignIpv6SubnetCidrBlockResponse, err error) {
	if request == nil {
		request = NewAssignIpv6SubnetCidrBlockRequest()
	}
	request.SetContext(ctx)

	response = NewAssignIpv6SubnetCidrBlockResponse()
	err = c.Send(request, response)
	return
}

func NewAssignPrivateIpAddressesRequest() (request *AssignPrivateIpAddressesRequest) {
	request = &AssignPrivateIpAddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssignPrivateIpAddresses")

	return
}

func NewAssignPrivateIpAddressesResponse() (response *AssignPrivateIpAddressesResponse) {
	response = &AssignPrivateIpAddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssignPrivateIpAddresses
// 本接口（AssignPrivateIpAddresses）用于弹性网卡申请内网 IP。
//
// * 一个弹性网卡支持绑定的IP地址是有限制的，更多资源限制信息详见<a href="/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 可以指定内网IP地址申请，内网IP地址类型不能为主IP，主IP已存在，不能修改，内网IP必须要弹性网卡所在子网内，而且不能被占用。
//
// * 在弹性网卡上申请一个到多个辅助内网IP，接口会在弹性网卡所在子网网段内返回指定数量的辅助内网IP。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_RESERVED = "InvalidParameterValue.Reserved"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_RESOURCEMISMATCH = "UnsupportedOperation.ResourceMismatch"
func (c *Client) AssignPrivateIpAddresses(request *AssignPrivateIpAddressesRequest) (response *AssignPrivateIpAddressesResponse, err error) {
	if request == nil {
		request = NewAssignPrivateIpAddressesRequest()
	}

	response = NewAssignPrivateIpAddressesResponse()
	err = c.Send(request, response)
	return
}

// AssignPrivateIpAddresses
// 本接口（AssignPrivateIpAddresses）用于弹性网卡申请内网 IP。
//
// * 一个弹性网卡支持绑定的IP地址是有限制的，更多资源限制信息详见<a href="/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 可以指定内网IP地址申请，内网IP地址类型不能为主IP，主IP已存在，不能修改，内网IP必须要弹性网卡所在子网内，而且不能被占用。
//
// * 在弹性网卡上申请一个到多个辅助内网IP，接口会在弹性网卡所在子网网段内返回指定数量的辅助内网IP。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_RESERVED = "InvalidParameterValue.Reserved"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_RESOURCEMISMATCH = "UnsupportedOperation.ResourceMismatch"
func (c *Client) AssignPrivateIpAddressesWithContext(ctx context.Context, request *AssignPrivateIpAddressesRequest) (response *AssignPrivateIpAddressesResponse, err error) {
	if request == nil {
		request = NewAssignPrivateIpAddressesRequest()
	}
	request.SetContext(ctx)

	response = NewAssignPrivateIpAddressesResponse()
	err = c.Send(request, response)
	return
}

func NewAssociateAddressRequest() (request *AssociateAddressRequest) {
	request = &AssociateAddressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssociateAddress")

	return
}

func NewAssociateAddressResponse() (response *AssociateAddressResponse) {
	response = &AssociateAddressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssociateAddress
// 本接口 (AssociateAddress) 用于将[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）绑定到实例或弹性网卡的指定内网 IP 上。
//
// * 将 EIP 绑定到实例（CVM）上，其本质是将 EIP 绑定到实例上主网卡的主内网 IP 上。
//
// * 将 EIP 绑定到主网卡的主内网IP上，绑定过程会把其上绑定的普通公网 IP 自动解绑并释放。
//
// * 将 EIP 绑定到指定网卡的内网 IP上（非主网卡的主内网IP），则必须先解绑该 EIP，才能再绑定新的。
//
// * 将 EIP 绑定到NAT网关，请使用接口[AssociateNatGatewayAddress](https://cloud.tencent.com/document/product/215/36722)
//
// * EIP 如果欠费或被封堵，则不能被绑定。
//
// * 只有状态为 UNBIND 的 EIP 才能够被绑定。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_ADDRESSENIINFONOTFOUND = "FailedOperation.AddressEniInfoNotFound"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDINSTANCEID_ALREADYBINDEIP = "InvalidInstanceId.AlreadyBindEip"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDNETWORKINTERFACEID_NOTFOUND = "InvalidNetworkInterfaceId.NotFound"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTAPPLICABLE = "InvalidParameterValue.AddressNotApplicable"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_COMBINATION = "InvalidParameterValue.Combination"
//	INVALIDPARAMETERVALUE_INSTANCEDOESNOTSUPPORTANYCAST = "InvalidParameterValue.InstanceDoesNotSupportAnycast"
//	INVALIDPARAMETERVALUE_INSTANCEHASWANIP = "InvalidParameterValue.InstanceHasWanIP"
//	INVALIDPARAMETERVALUE_INSTANCENOWANIP = "InvalidParameterValue.InstanceNoWanIP"
//	INVALIDPARAMETERVALUE_INSTANCENORMALPUBLICIPBLOCKED = "InvalidParameterValue.InstanceNormalPublicIpBlocked"
//	INVALIDPARAMETERVALUE_INSTANCENOTMATCHASSOCIATEENI = "InvalidParameterValue.InstanceNotMatchAssociateEni"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEINTERNETCHARGETYPE = "InvalidParameterValue.InvalidInstanceInternetChargeType"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCESTATE = "InvalidParameterValue.InvalidInstanceState"
//	INVALIDPARAMETERVALUE_LBALREADYBINDEIP = "InvalidParameterValue.LBAlreadyBindEip"
//	INVALIDPARAMETERVALUE_MISSINGASSOCIATEENTITY = "InvalidParameterValue.MissingAssociateEntity"
//	INVALIDPARAMETERVALUE_NETWORKINTERFACENOTFOUND = "InvalidParameterValue.NetworkInterfaceNotFound"
//	INVALIDPARAMETERVALUE_RESOURCENOTSUPPORT = "InvalidParameterValue.ResourceNotSupport"
//	INVALIDPRIVATEIPADDRESS_ALREADYBINDEIP = "InvalidPrivateIpAddress.AlreadyBindEip"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INCORRECTADDRESSRESOURCETYPE = "UnsupportedOperation.IncorrectAddressResourceType"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
//	UNSUPPORTEDOPERATION_INVALIDADDRESSINTERNETCHARGETYPE = "UnsupportedOperation.InvalidAddressInternetChargeType"
//	UNSUPPORTEDOPERATION_ISPNOTSUPPORTED = "UnsupportedOperation.IspNotSupported"
func (c *Client) AssociateAddress(request *AssociateAddressRequest) (response *AssociateAddressResponse, err error) {
	if request == nil {
		request = NewAssociateAddressRequest()
	}

	response = NewAssociateAddressResponse()
	err = c.Send(request, response)
	return
}

// AssociateAddress
// 本接口 (AssociateAddress) 用于将[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）绑定到实例或弹性网卡的指定内网 IP 上。
//
// * 将 EIP 绑定到实例（CVM）上，其本质是将 EIP 绑定到实例上主网卡的主内网 IP 上。
//
// * 将 EIP 绑定到主网卡的主内网IP上，绑定过程会把其上绑定的普通公网 IP 自动解绑并释放。
//
// * 将 EIP 绑定到指定网卡的内网 IP上（非主网卡的主内网IP），则必须先解绑该 EIP，才能再绑定新的。
//
// * 将 EIP 绑定到NAT网关，请使用接口[AssociateNatGatewayAddress](https://cloud.tencent.com/document/product/215/36722)
//
// * EIP 如果欠费或被封堵，则不能被绑定。
//
// * 只有状态为 UNBIND 的 EIP 才能够被绑定。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_ADDRESSENIINFONOTFOUND = "FailedOperation.AddressEniInfoNotFound"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDINSTANCEID_ALREADYBINDEIP = "InvalidInstanceId.AlreadyBindEip"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDNETWORKINTERFACEID_NOTFOUND = "InvalidNetworkInterfaceId.NotFound"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTAPPLICABLE = "InvalidParameterValue.AddressNotApplicable"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_COMBINATION = "InvalidParameterValue.Combination"
//	INVALIDPARAMETERVALUE_INSTANCEDOESNOTSUPPORTANYCAST = "InvalidParameterValue.InstanceDoesNotSupportAnycast"
//	INVALIDPARAMETERVALUE_INSTANCEHASWANIP = "InvalidParameterValue.InstanceHasWanIP"
//	INVALIDPARAMETERVALUE_INSTANCENOWANIP = "InvalidParameterValue.InstanceNoWanIP"
//	INVALIDPARAMETERVALUE_INSTANCENORMALPUBLICIPBLOCKED = "InvalidParameterValue.InstanceNormalPublicIpBlocked"
//	INVALIDPARAMETERVALUE_INSTANCENOTMATCHASSOCIATEENI = "InvalidParameterValue.InstanceNotMatchAssociateEni"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCEINTERNETCHARGETYPE = "InvalidParameterValue.InvalidInstanceInternetChargeType"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCESTATE = "InvalidParameterValue.InvalidInstanceState"
//	INVALIDPARAMETERVALUE_LBALREADYBINDEIP = "InvalidParameterValue.LBAlreadyBindEip"
//	INVALIDPARAMETERVALUE_MISSINGASSOCIATEENTITY = "InvalidParameterValue.MissingAssociateEntity"
//	INVALIDPARAMETERVALUE_NETWORKINTERFACENOTFOUND = "InvalidParameterValue.NetworkInterfaceNotFound"
//	INVALIDPARAMETERVALUE_RESOURCENOTSUPPORT = "InvalidParameterValue.ResourceNotSupport"
//	INVALIDPRIVATEIPADDRESS_ALREADYBINDEIP = "InvalidPrivateIpAddress.AlreadyBindEip"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INCORRECTADDRESSRESOURCETYPE = "UnsupportedOperation.IncorrectAddressResourceType"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
//	UNSUPPORTEDOPERATION_INVALIDADDRESSINTERNETCHARGETYPE = "UnsupportedOperation.InvalidAddressInternetChargeType"
//	UNSUPPORTEDOPERATION_ISPNOTSUPPORTED = "UnsupportedOperation.IspNotSupported"
func (c *Client) AssociateAddressWithContext(ctx context.Context, request *AssociateAddressRequest) (response *AssociateAddressResponse, err error) {
	if request == nil {
		request = NewAssociateAddressRequest()
	}
	request.SetContext(ctx)

	response = NewAssociateAddressResponse()
	err = c.Send(request, response)
	return
}

func NewAssociateDhcpIpWithAddressIpRequest() (request *AssociateDhcpIpWithAddressIpRequest) {
	request = &AssociateDhcpIpWithAddressIpRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssociateDhcpIpWithAddressIp")

	return
}

func NewAssociateDhcpIpWithAddressIpResponse() (response *AssociateDhcpIpWithAddressIpResponse) {
	response = &AssociateDhcpIpWithAddressIpResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssociateDhcpIpWithAddressIp
// 本接口（AssociateDhcpIpWithAddressIp）用于DhcpIp绑定弹性公网IP（EIP）。<br />
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_BINDEIP = "UnsupportedOperation.BindEIP"
func (c *Client) AssociateDhcpIpWithAddressIp(request *AssociateDhcpIpWithAddressIpRequest) (response *AssociateDhcpIpWithAddressIpResponse, err error) {
	if request == nil {
		request = NewAssociateDhcpIpWithAddressIpRequest()
	}

	response = NewAssociateDhcpIpWithAddressIpResponse()
	err = c.Send(request, response)
	return
}

// AssociateDhcpIpWithAddressIp
// 本接口（AssociateDhcpIpWithAddressIp）用于DhcpIp绑定弹性公网IP（EIP）。<br />
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_BINDEIP = "UnsupportedOperation.BindEIP"
func (c *Client) AssociateDhcpIpWithAddressIpWithContext(ctx context.Context, request *AssociateDhcpIpWithAddressIpRequest) (response *AssociateDhcpIpWithAddressIpResponse, err error) {
	if request == nil {
		request = NewAssociateDhcpIpWithAddressIpRequest()
	}
	request.SetContext(ctx)

	response = NewAssociateDhcpIpWithAddressIpResponse()
	err = c.Send(request, response)
	return
}

func NewAssociateDirectConnectGatewayNatGatewayRequest() (request *AssociateDirectConnectGatewayNatGatewayRequest) {
	request = &AssociateDirectConnectGatewayNatGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssociateDirectConnectGatewayNatGateway")

	return
}

func NewAssociateDirectConnectGatewayNatGatewayResponse() (response *AssociateDirectConnectGatewayNatGatewayResponse) {
	response = &AssociateDirectConnectGatewayNatGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssociateDirectConnectGatewayNatGateway
// 将专线网关与NAT网关绑定，专线网关默认路由指向NAT网关
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_VPGTYPENOTMATCH = "InvalidParameterValue.VpgTypeNotMatch"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) AssociateDirectConnectGatewayNatGateway(request *AssociateDirectConnectGatewayNatGatewayRequest) (response *AssociateDirectConnectGatewayNatGatewayResponse, err error) {
	if request == nil {
		request = NewAssociateDirectConnectGatewayNatGatewayRequest()
	}

	response = NewAssociateDirectConnectGatewayNatGatewayResponse()
	err = c.Send(request, response)
	return
}

// AssociateDirectConnectGatewayNatGateway
// 将专线网关与NAT网关绑定，专线网关默认路由指向NAT网关
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_VPGTYPENOTMATCH = "InvalidParameterValue.VpgTypeNotMatch"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) AssociateDirectConnectGatewayNatGatewayWithContext(ctx context.Context, request *AssociateDirectConnectGatewayNatGatewayRequest) (response *AssociateDirectConnectGatewayNatGatewayResponse, err error) {
	if request == nil {
		request = NewAssociateDirectConnectGatewayNatGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewAssociateDirectConnectGatewayNatGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewAssociateNatGatewayAddressRequest() (request *AssociateNatGatewayAddressRequest) {
	request = &AssociateNatGatewayAddressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssociateNatGatewayAddress")

	return
}

func NewAssociateNatGatewayAddressResponse() (response *AssociateNatGatewayAddressResponse) {
	response = &AssociateNatGatewayAddressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssociateNatGatewayAddress
// 本接口(AssociateNatGatewayAddress)用于NAT网关绑定弹性IP（EIP）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED_ADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.AddressQuotaLimitExceeded"
//	LIMITEXCEEDED_DAILYALLOCATEADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.DailyAllocateAddressQuotaLimitExceeded"
//	LIMITEXCEEDED_PUBLICIPADDRESSPERNATGATEWAYLIMITEXCEEDED = "LimitExceeded.PublicIpAddressPerNatGatewayLimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTBGPIP = "UnsupportedOperation.PublicIpAddressIsNotBGPIp"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSNOTBILLEDBYTRAFFIC = "UnsupportedOperation.PublicIpAddressNotBilledByTraffic"
func (c *Client) AssociateNatGatewayAddress(request *AssociateNatGatewayAddressRequest) (response *AssociateNatGatewayAddressResponse, err error) {
	if request == nil {
		request = NewAssociateNatGatewayAddressRequest()
	}

	response = NewAssociateNatGatewayAddressResponse()
	err = c.Send(request, response)
	return
}

// AssociateNatGatewayAddress
// 本接口(AssociateNatGatewayAddress)用于NAT网关绑定弹性IP（EIP）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED_ADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.AddressQuotaLimitExceeded"
//	LIMITEXCEEDED_DAILYALLOCATEADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.DailyAllocateAddressQuotaLimitExceeded"
//	LIMITEXCEEDED_PUBLICIPADDRESSPERNATGATEWAYLIMITEXCEEDED = "LimitExceeded.PublicIpAddressPerNatGatewayLimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTBGPIP = "UnsupportedOperation.PublicIpAddressIsNotBGPIp"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSNOTBILLEDBYTRAFFIC = "UnsupportedOperation.PublicIpAddressNotBilledByTraffic"
func (c *Client) AssociateNatGatewayAddressWithContext(ctx context.Context, request *AssociateNatGatewayAddressRequest) (response *AssociateNatGatewayAddressResponse, err error) {
	if request == nil {
		request = NewAssociateNatGatewayAddressRequest()
	}
	request.SetContext(ctx)

	response = NewAssociateNatGatewayAddressResponse()
	err = c.Send(request, response)
	return
}

func NewAssociateNetworkAclSubnetsRequest() (request *AssociateNetworkAclSubnetsRequest) {
	request = &AssociateNetworkAclSubnetsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssociateNetworkAclSubnets")

	return
}

func NewAssociateNetworkAclSubnetsResponse() (response *AssociateNetworkAclSubnetsResponse) {
	response = &AssociateNetworkAclSubnetsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssociateNetworkAclSubnets
// 本接口（AssociateNetworkAclSubnets）用于网络ACL关联vpc下的子网。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) AssociateNetworkAclSubnets(request *AssociateNetworkAclSubnetsRequest) (response *AssociateNetworkAclSubnetsResponse, err error) {
	if request == nil {
		request = NewAssociateNetworkAclSubnetsRequest()
	}

	response = NewAssociateNetworkAclSubnetsResponse()
	err = c.Send(request, response)
	return
}

// AssociateNetworkAclSubnets
// 本接口（AssociateNetworkAclSubnets）用于网络ACL关联vpc下的子网。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) AssociateNetworkAclSubnetsWithContext(ctx context.Context, request *AssociateNetworkAclSubnetsRequest) (response *AssociateNetworkAclSubnetsResponse, err error) {
	if request == nil {
		request = NewAssociateNetworkAclSubnetsRequest()
	}
	request.SetContext(ctx)

	response = NewAssociateNetworkAclSubnetsResponse()
	err = c.Send(request, response)
	return
}

func NewAssociateNetworkInterfaceSecurityGroupsRequest() (request *AssociateNetworkInterfaceSecurityGroupsRequest) {
	request = &AssociateNetworkInterfaceSecurityGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AssociateNetworkInterfaceSecurityGroups")

	return
}

func NewAssociateNetworkInterfaceSecurityGroupsResponse() (response *AssociateNetworkInterfaceSecurityGroupsResponse) {
	response = &AssociateNetworkInterfaceSecurityGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssociateNetworkInterfaceSecurityGroups
// 本接口（AssociateNetworkInterfaceSecurityGroups）用于弹性网卡绑定安全组（SecurityGroup）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) AssociateNetworkInterfaceSecurityGroups(request *AssociateNetworkInterfaceSecurityGroupsRequest) (response *AssociateNetworkInterfaceSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewAssociateNetworkInterfaceSecurityGroupsRequest()
	}

	response = NewAssociateNetworkInterfaceSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

// AssociateNetworkInterfaceSecurityGroups
// 本接口（AssociateNetworkInterfaceSecurityGroups）用于弹性网卡绑定安全组（SecurityGroup）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) AssociateNetworkInterfaceSecurityGroupsWithContext(ctx context.Context, request *AssociateNetworkInterfaceSecurityGroupsRequest) (response *AssociateNetworkInterfaceSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewAssociateNetworkInterfaceSecurityGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewAssociateNetworkInterfaceSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewAttachCcnInstancesRequest() (request *AttachCcnInstancesRequest) {
	request = &AttachCcnInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AttachCcnInstances")

	return
}

func NewAttachCcnInstancesResponse() (response *AttachCcnInstancesResponse) {
	response = &AttachCcnInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AttachCcnInstances
// 本接口（AttachCcnInstances）用于将网络实例加载到云联网实例中，网络实例包括VPC和专线网关。<br />
//
// 每个云联网能够关联的网络实例个数是有限的，详请参考产品文档。如果需要扩充请联系在线客服。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CCNATTACHBMVPCLIMITEXCEEDED = "InvalidParameterValue.CcnAttachBmvpcLimitExceeded"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_APPIDNOTFOUND = "UnsupportedOperation.AppIdNotFound"
//	UNSUPPORTEDOPERATION_CCNATTACHED = "UnsupportedOperation.CcnAttached"
//	UNSUPPORTEDOPERATION_CCNROUTETABLENOTEXIST = "UnsupportedOperation.CcnRouteTableNotExist"
//	UNSUPPORTEDOPERATION_INSTANCEANDRTBNOTMATCH = "UnsupportedOperation.InstanceAndRtbNotMatch"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_ISNOTFINANCEACCOUNT = "UnsupportedOperation.IsNotFinanceAccount"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
//	UNSUPPORTEDOPERATION_UNABLECROSSBORDER = "UnsupportedOperation.UnableCrossBorder"
//	UNSUPPORTEDOPERATION_UNABLECROSSFINANCE = "UnsupportedOperation.UnableCrossFinance"
func (c *Client) AttachCcnInstances(request *AttachCcnInstancesRequest) (response *AttachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewAttachCcnInstancesRequest()
	}

	response = NewAttachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

// AttachCcnInstances
// 本接口（AttachCcnInstances）用于将网络实例加载到云联网实例中，网络实例包括VPC和专线网关。<br />
//
// 每个云联网能够关联的网络实例个数是有限的，详请参考产品文档。如果需要扩充请联系在线客服。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CCNATTACHBMVPCLIMITEXCEEDED = "InvalidParameterValue.CcnAttachBmvpcLimitExceeded"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_APPIDNOTFOUND = "UnsupportedOperation.AppIdNotFound"
//	UNSUPPORTEDOPERATION_CCNATTACHED = "UnsupportedOperation.CcnAttached"
//	UNSUPPORTEDOPERATION_CCNROUTETABLENOTEXIST = "UnsupportedOperation.CcnRouteTableNotExist"
//	UNSUPPORTEDOPERATION_INSTANCEANDRTBNOTMATCH = "UnsupportedOperation.InstanceAndRtbNotMatch"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_ISNOTFINANCEACCOUNT = "UnsupportedOperation.IsNotFinanceAccount"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
//	UNSUPPORTEDOPERATION_UNABLECROSSBORDER = "UnsupportedOperation.UnableCrossBorder"
//	UNSUPPORTEDOPERATION_UNABLECROSSFINANCE = "UnsupportedOperation.UnableCrossFinance"
func (c *Client) AttachCcnInstancesWithContext(ctx context.Context, request *AttachCcnInstancesRequest) (response *AttachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewAttachCcnInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewAttachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewAttachClassicLinkVpcRequest() (request *AttachClassicLinkVpcRequest) {
	request = &AttachClassicLinkVpcRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AttachClassicLinkVpc")

	return
}

func NewAttachClassicLinkVpcResponse() (response *AttachClassicLinkVpcResponse) {
	response = &AttachClassicLinkVpcResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AttachClassicLinkVpc
// 本接口(AttachClassicLinkVpc)用于创建私有网络和基础网络设备互通。
//
// * 私有网络和基础网络设备必须在同一个地域。
//
// * 私有网络和基础网络的区别详见vpc产品文档-<a href="https://cloud.tencent.com/document/product/215/30720">私有网络与基础网络</a>。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CIDRUNSUPPORTEDCLASSICLINK = "UnsupportedOperation.CIDRUnSupportedClassicLink"
//	UNSUPPORTEDOPERATION_CLASSICINSTANCEIDALREADYEXISTS = "UnsupportedOperation.ClassicInstanceIdAlreadyExists"
func (c *Client) AttachClassicLinkVpc(request *AttachClassicLinkVpcRequest) (response *AttachClassicLinkVpcResponse, err error) {
	if request == nil {
		request = NewAttachClassicLinkVpcRequest()
	}

	response = NewAttachClassicLinkVpcResponse()
	err = c.Send(request, response)
	return
}

// AttachClassicLinkVpc
// 本接口(AttachClassicLinkVpc)用于创建私有网络和基础网络设备互通。
//
// * 私有网络和基础网络设备必须在同一个地域。
//
// * 私有网络和基础网络的区别详见vpc产品文档-<a href="https://cloud.tencent.com/document/product/215/30720">私有网络与基础网络</a>。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CIDRUNSUPPORTEDCLASSICLINK = "UnsupportedOperation.CIDRUnSupportedClassicLink"
//	UNSUPPORTEDOPERATION_CLASSICINSTANCEIDALREADYEXISTS = "UnsupportedOperation.ClassicInstanceIdAlreadyExists"
func (c *Client) AttachClassicLinkVpcWithContext(ctx context.Context, request *AttachClassicLinkVpcRequest) (response *AttachClassicLinkVpcResponse, err error) {
	if request == nil {
		request = NewAttachClassicLinkVpcRequest()
	}
	request.SetContext(ctx)

	response = NewAttachClassicLinkVpcResponse()
	err = c.Send(request, response)
	return
}

func NewAttachNetworkInterfaceRequest() (request *AttachNetworkInterfaceRequest) {
	request = &AttachNetworkInterfaceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AttachNetworkInterface")

	return
}

func NewAttachNetworkInterfaceResponse() (response *AttachNetworkInterfaceResponse) {
	response = &AttachNetworkInterfaceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AttachNetworkInterface
// 本接口（AttachNetworkInterface）用于弹性网卡绑定云服务器。
//
// * 一个云服务器可以绑定多个弹性网卡，但只能绑定一个主网卡。更多限制信息详见<a href="https://cloud.tencent.com/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 一个弹性网卡只能同时绑定一个云服务器。
//
// * 只有运行中或者已关机状态的云服务器才能绑定弹性网卡，查看云服务器状态详见<a href="https://cloud.tencent.com/document/api/213/9452#InstanceStatus">腾讯云服务器信息</a>。
//
// * 弹性网卡绑定的云服务器必须是私有网络的，而且云服务器所在可用区必须和弹性网卡子网的可用区相同。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ATTACHMENTALREADYEXISTS = "UnsupportedOperation.AttachmentAlreadyExists"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDINSTANCEFAMILY = "UnsupportedOperation.UnsupportedInstanceFamily"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
//	UNSUPPORTEDOPERATION_ZONEMISMATCH = "UnsupportedOperation.ZoneMismatch"
func (c *Client) AttachNetworkInterface(request *AttachNetworkInterfaceRequest) (response *AttachNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewAttachNetworkInterfaceRequest()
	}

	response = NewAttachNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

// AttachNetworkInterface
// 本接口（AttachNetworkInterface）用于弹性网卡绑定云服务器。
//
// * 一个云服务器可以绑定多个弹性网卡，但只能绑定一个主网卡。更多限制信息详见<a href="https://cloud.tencent.com/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 一个弹性网卡只能同时绑定一个云服务器。
//
// * 只有运行中或者已关机状态的云服务器才能绑定弹性网卡，查看云服务器状态详见<a href="https://cloud.tencent.com/document/api/213/9452#InstanceStatus">腾讯云服务器信息</a>。
//
// * 弹性网卡绑定的云服务器必须是私有网络的，而且云服务器所在可用区必须和弹性网卡子网的可用区相同。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ATTACHMENTALREADYEXISTS = "UnsupportedOperation.AttachmentAlreadyExists"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDINSTANCEFAMILY = "UnsupportedOperation.UnsupportedInstanceFamily"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
//	UNSUPPORTEDOPERATION_ZONEMISMATCH = "UnsupportedOperation.ZoneMismatch"
func (c *Client) AttachNetworkInterfaceWithContext(ctx context.Context, request *AttachNetworkInterfaceRequest) (response *AttachNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewAttachNetworkInterfaceRequest()
	}
	request.SetContext(ctx)

	response = NewAttachNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

func NewAuditCrossBorderComplianceRequest() (request *AuditCrossBorderComplianceRequest) {
	request = &AuditCrossBorderComplianceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "AuditCrossBorderCompliance")

	return
}

func NewAuditCrossBorderComplianceResponse() (response *AuditCrossBorderComplianceResponse) {
	response = &AuditCrossBorderComplianceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AuditCrossBorderCompliance
// 本接口（AuditCrossBorderCompliance）用于服务商操作合规化资质审批。
//
// * 服务商只能操作提交到本服务商的审批单，后台会校验身份。即只授权给服务商的`APPID` 调用本接口。
//
// * `APPROVED` 状态的审批单，可以再次操作为 `DENY`；`DENY` 状态的审批单，也可以再次操作为 `APPROVED`。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) AuditCrossBorderCompliance(request *AuditCrossBorderComplianceRequest) (response *AuditCrossBorderComplianceResponse, err error) {
	if request == nil {
		request = NewAuditCrossBorderComplianceRequest()
	}

	response = NewAuditCrossBorderComplianceResponse()
	err = c.Send(request, response)
	return
}

// AuditCrossBorderCompliance
// 本接口（AuditCrossBorderCompliance）用于服务商操作合规化资质审批。
//
// * 服务商只能操作提交到本服务商的审批单，后台会校验身份。即只授权给服务商的`APPID` 调用本接口。
//
// * `APPROVED` 状态的审批单，可以再次操作为 `DENY`；`DENY` 状态的审批单，也可以再次操作为 `APPROVED`。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) AuditCrossBorderComplianceWithContext(ctx context.Context, request *AuditCrossBorderComplianceRequest) (response *AuditCrossBorderComplianceResponse, err error) {
	if request == nil {
		request = NewAuditCrossBorderComplianceRequest()
	}
	request.SetContext(ctx)

	response = NewAuditCrossBorderComplianceResponse()
	err = c.Send(request, response)
	return
}

func NewCheckAssistantCidrRequest() (request *CheckAssistantCidrRequest) {
	request = &CheckAssistantCidrRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CheckAssistantCidr")

	return
}

func NewCheckAssistantCidrResponse() (response *CheckAssistantCidrResponse) {
	response = &CheckAssistantCidrResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CheckAssistantCidr
// 本接口(CheckAssistantCidr)用于检查辅助CIDR是否与存量路由、对等连接（对端VPC的CIDR）等资源存在冲突。如果存在重叠，则返回重叠的资源。（接口灰度中，如需使用请提工单。）
//
// * 检测辅助CIDR是否与当前VPC的主CIDR和辅助CIDR存在重叠。
//
// * 检测辅助CIDR是否与当前VPC的路由的目的端存在重叠。
//
// * 检测辅助CIDR是否与当前VPC的对等连接，对端VPC下的主CIDR或辅助CIDR存在重叠。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CheckAssistantCidr(request *CheckAssistantCidrRequest) (response *CheckAssistantCidrResponse, err error) {
	if request == nil {
		request = NewCheckAssistantCidrRequest()
	}

	response = NewCheckAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

// CheckAssistantCidr
// 本接口(CheckAssistantCidr)用于检查辅助CIDR是否与存量路由、对等连接（对端VPC的CIDR）等资源存在冲突。如果存在重叠，则返回重叠的资源。（接口灰度中，如需使用请提工单。）
//
// * 检测辅助CIDR是否与当前VPC的主CIDR和辅助CIDR存在重叠。
//
// * 检测辅助CIDR是否与当前VPC的路由的目的端存在重叠。
//
// * 检测辅助CIDR是否与当前VPC的对等连接，对端VPC下的主CIDR或辅助CIDR存在重叠。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CheckAssistantCidrWithContext(ctx context.Context, request *CheckAssistantCidrRequest) (response *CheckAssistantCidrResponse, err error) {
	if request == nil {
		request = NewCheckAssistantCidrRequest()
	}
	request.SetContext(ctx)

	response = NewCheckAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

func NewCheckDefaultSubnetRequest() (request *CheckDefaultSubnetRequest) {
	request = &CheckDefaultSubnetRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CheckDefaultSubnet")

	return
}

func NewCheckDefaultSubnetResponse() (response *CheckDefaultSubnetResponse) {
	response = &CheckDefaultSubnetResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CheckDefaultSubnet
// 本接口（CheckDefaultSubnet）用于预判是否可建默认子网。
//
// 可能返回的错误码:
//
//	RESOURCEINSUFFICIENT_CIDRBLOCK = "ResourceInsufficient.CidrBlock"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CheckDefaultSubnet(request *CheckDefaultSubnetRequest) (response *CheckDefaultSubnetResponse, err error) {
	if request == nil {
		request = NewCheckDefaultSubnetRequest()
	}

	response = NewCheckDefaultSubnetResponse()
	err = c.Send(request, response)
	return
}

// CheckDefaultSubnet
// 本接口（CheckDefaultSubnet）用于预判是否可建默认子网。
//
// 可能返回的错误码:
//
//	RESOURCEINSUFFICIENT_CIDRBLOCK = "ResourceInsufficient.CidrBlock"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CheckDefaultSubnetWithContext(ctx context.Context, request *CheckDefaultSubnetRequest) (response *CheckDefaultSubnetResponse, err error) {
	if request == nil {
		request = NewCheckDefaultSubnetRequest()
	}
	request.SetContext(ctx)

	response = NewCheckDefaultSubnetResponse()
	err = c.Send(request, response)
	return
}

func NewCheckNetDetectStateRequest() (request *CheckNetDetectStateRequest) {
	request = &CheckNetDetectStateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CheckNetDetectState")

	return
}

func NewCheckNetDetectStateResponse() (response *CheckNetDetectStateResponse) {
	response = &CheckNetDetectStateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CheckNetDetectState
// 本接口(CheckNetDetectState)用于验证网络探测。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_NEXTHOPMISMATCH = "InvalidParameter.NextHopMismatch"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_NETDETECTINVPC = "InvalidParameterValue.NetDetectInVpc"
//	INVALIDPARAMETERVALUE_NETDETECTNOTFOUNDIP = "InvalidParameterValue.NetDetectNotFoundIp"
//	INVALIDPARAMETERVALUE_NETDETECTSAMEIP = "InvalidParameterValue.NetDetectSameIp"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) CheckNetDetectState(request *CheckNetDetectStateRequest) (response *CheckNetDetectStateResponse, err error) {
	if request == nil {
		request = NewCheckNetDetectStateRequest()
	}

	response = NewCheckNetDetectStateResponse()
	err = c.Send(request, response)
	return
}

// CheckNetDetectState
// 本接口(CheckNetDetectState)用于验证网络探测。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_NEXTHOPMISMATCH = "InvalidParameter.NextHopMismatch"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_NETDETECTINVPC = "InvalidParameterValue.NetDetectInVpc"
//	INVALIDPARAMETERVALUE_NETDETECTNOTFOUNDIP = "InvalidParameterValue.NetDetectNotFoundIp"
//	INVALIDPARAMETERVALUE_NETDETECTSAMEIP = "InvalidParameterValue.NetDetectSameIp"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) CheckNetDetectStateWithContext(ctx context.Context, request *CheckNetDetectStateRequest) (response *CheckNetDetectStateResponse, err error) {
	if request == nil {
		request = NewCheckNetDetectStateRequest()
	}
	request.SetContext(ctx)

	response = NewCheckNetDetectStateResponse()
	err = c.Send(request, response)
	return
}

func NewCloneSecurityGroupRequest() (request *CloneSecurityGroupRequest) {
	request = &CloneSecurityGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CloneSecurityGroup")

	return
}

func NewCloneSecurityGroupResponse() (response *CloneSecurityGroupResponse) {
	response = &CloneSecurityGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CloneSecurityGroup
// 本接口（CloneSecurityGroup）用于根据存量的安全组，克隆创建出同样规则配置的安全组。仅克隆安全组及其规则信息，不会克隆安全组标签信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CloneSecurityGroup(request *CloneSecurityGroupRequest) (response *CloneSecurityGroupResponse, err error) {
	if request == nil {
		request = NewCloneSecurityGroupRequest()
	}

	response = NewCloneSecurityGroupResponse()
	err = c.Send(request, response)
	return
}

// CloneSecurityGroup
// 本接口（CloneSecurityGroup）用于根据存量的安全组，克隆创建出同样规则配置的安全组。仅克隆安全组及其规则信息，不会克隆安全组标签信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CloneSecurityGroupWithContext(ctx context.Context, request *CloneSecurityGroupRequest) (response *CloneSecurityGroupResponse, err error) {
	if request == nil {
		request = NewCloneSecurityGroupRequest()
	}
	request.SetContext(ctx)

	response = NewCloneSecurityGroupResponse()
	err = c.Send(request, response)
	return
}

func NewCreateAddressTemplateRequest() (request *CreateAddressTemplateRequest) {
	request = &CreateAddressTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateAddressTemplate")

	return
}

func NewCreateAddressTemplateResponse() (response *CreateAddressTemplateResponse) {
	response = &CreateAddressTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateAddressTemplate
// 本接口（CreateAddressTemplate）用于创建IP地址模版
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) CreateAddressTemplate(request *CreateAddressTemplateRequest) (response *CreateAddressTemplateResponse, err error) {
	if request == nil {
		request = NewCreateAddressTemplateRequest()
	}

	response = NewCreateAddressTemplateResponse()
	err = c.Send(request, response)
	return
}

// CreateAddressTemplate
// 本接口（CreateAddressTemplate）用于创建IP地址模版
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) CreateAddressTemplateWithContext(ctx context.Context, request *CreateAddressTemplateRequest) (response *CreateAddressTemplateResponse, err error) {
	if request == nil {
		request = NewCreateAddressTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewCreateAddressTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewCreateAddressTemplateGroupRequest() (request *CreateAddressTemplateGroupRequest) {
	request = &CreateAddressTemplateGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateAddressTemplateGroup")

	return
}

func NewCreateAddressTemplateGroupResponse() (response *CreateAddressTemplateGroupResponse) {
	response = &CreateAddressTemplateGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateAddressTemplateGroup
// 本接口（CreateAddressTemplateGroup）用于创建IP地址模版集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) CreateAddressTemplateGroup(request *CreateAddressTemplateGroupRequest) (response *CreateAddressTemplateGroupResponse, err error) {
	if request == nil {
		request = NewCreateAddressTemplateGroupRequest()
	}

	response = NewCreateAddressTemplateGroupResponse()
	err = c.Send(request, response)
	return
}

// CreateAddressTemplateGroup
// 本接口（CreateAddressTemplateGroup）用于创建IP地址模版集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) CreateAddressTemplateGroupWithContext(ctx context.Context, request *CreateAddressTemplateGroupRequest) (response *CreateAddressTemplateGroupResponse, err error) {
	if request == nil {
		request = NewCreateAddressTemplateGroupRequest()
	}
	request.SetContext(ctx)

	response = NewCreateAddressTemplateGroupResponse()
	err = c.Send(request, response)
	return
}

func NewCreateAndAttachNetworkInterfaceRequest() (request *CreateAndAttachNetworkInterfaceRequest) {
	request = &CreateAndAttachNetworkInterfaceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateAndAttachNetworkInterface")

	return
}

func NewCreateAndAttachNetworkInterfaceResponse() (response *CreateAndAttachNetworkInterfaceResponse) {
	response = &CreateAndAttachNetworkInterfaceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateAndAttachNetworkInterface
// 本接口（CreateAndAttachNetworkInterface）用于创建弹性网卡并绑定云服务器。
//
// * 创建弹性网卡时可以指定内网IP，并且可以指定一个主IP，指定的内网IP必须在弹性网卡所在子网内，而且不能被占用。
//
// * 创建弹性网卡时可以指定需要申请的内网IP数量，系统会随机生成内网IP地址。
//
// * 一个弹性网卡支持绑定的IP地址是有限制的，更多资源限制信息详见<a href="/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 创建弹性网卡同时可以绑定已有安全组。
//
// * 创建弹性网卡同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_RESERVED = "InvalidParameterValue.Reserved"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDINSTANCEFAMILY = "UnsupportedOperation.UnsupportedInstanceFamily"
func (c *Client) CreateAndAttachNetworkInterface(request *CreateAndAttachNetworkInterfaceRequest) (response *CreateAndAttachNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewCreateAndAttachNetworkInterfaceRequest()
	}

	response = NewCreateAndAttachNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

// CreateAndAttachNetworkInterface
// 本接口（CreateAndAttachNetworkInterface）用于创建弹性网卡并绑定云服务器。
//
// * 创建弹性网卡时可以指定内网IP，并且可以指定一个主IP，指定的内网IP必须在弹性网卡所在子网内，而且不能被占用。
//
// * 创建弹性网卡时可以指定需要申请的内网IP数量，系统会随机生成内网IP地址。
//
// * 一个弹性网卡支持绑定的IP地址是有限制的，更多资源限制信息详见<a href="/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 创建弹性网卡同时可以绑定已有安全组。
//
// * 创建弹性网卡同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_RESERVED = "InvalidParameterValue.Reserved"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDINSTANCEFAMILY = "UnsupportedOperation.UnsupportedInstanceFamily"
func (c *Client) CreateAndAttachNetworkInterfaceWithContext(ctx context.Context, request *CreateAndAttachNetworkInterfaceRequest) (response *CreateAndAttachNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewCreateAndAttachNetworkInterfaceRequest()
	}
	request.SetContext(ctx)

	response = NewCreateAndAttachNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

func NewCreateAssistantCidrRequest() (request *CreateAssistantCidrRequest) {
	request = &CreateAssistantCidrRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateAssistantCidr")

	return
}

func NewCreateAssistantCidrResponse() (response *CreateAssistantCidrResponse) {
	response = &CreateAssistantCidrResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateAssistantCidr
// 本接口(CreateAssistantCidr)用于批量创建辅助CIDR。（接口灰度中，如需使用请提工单。）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETOVERLAPASSISTCIDR = "InvalidParameterValue.SubnetOverlapAssistCidr"
//	INVALIDPARAMETERVALUE_SUBNETRANGE = "InvalidParameterValue.SubnetRange"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateAssistantCidr(request *CreateAssistantCidrRequest) (response *CreateAssistantCidrResponse, err error) {
	if request == nil {
		request = NewCreateAssistantCidrRequest()
	}

	response = NewCreateAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

// CreateAssistantCidr
// 本接口(CreateAssistantCidr)用于批量创建辅助CIDR。（接口灰度中，如需使用请提工单。）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETOVERLAPASSISTCIDR = "InvalidParameterValue.SubnetOverlapAssistCidr"
//	INVALIDPARAMETERVALUE_SUBNETRANGE = "InvalidParameterValue.SubnetRange"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateAssistantCidrWithContext(ctx context.Context, request *CreateAssistantCidrRequest) (response *CreateAssistantCidrResponse, err error) {
	if request == nil {
		request = NewCreateAssistantCidrRequest()
	}
	request.SetContext(ctx)

	response = NewCreateAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

func NewCreateBandwidthPackageRequest() (request *CreateBandwidthPackageRequest) {
	request = &CreateBandwidthPackageRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateBandwidthPackage")

	return
}

func NewCreateBandwidthPackageResponse() (response *CreateBandwidthPackageResponse) {
	response = &CreateBandwidthPackageResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateBandwidthPackage
// 本接口 (CreateBandwidthPackage) 支持创建[设备带宽包](https://cloud.tencent.com/document/product/684/15245#bwptype)和[IP带宽包](https://cloud.tencent.com/document/product/684/15245#bwptype)。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDPARAMETERVALUE_COMBINATION = "InvalidParameterValue.Combination"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TAGNOTEXISTED = "InvalidParameterValue.TagNotExisted"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDRESOURCEINTERNETCHARGETYPE = "UnsupportedOperation.InvalidResourceInternetChargeType"
func (c *Client) CreateBandwidthPackage(request *CreateBandwidthPackageRequest) (response *CreateBandwidthPackageResponse, err error) {
	if request == nil {
		request = NewCreateBandwidthPackageRequest()
	}

	response = NewCreateBandwidthPackageResponse()
	err = c.Send(request, response)
	return
}

// CreateBandwidthPackage
// 本接口 (CreateBandwidthPackage) 支持创建[设备带宽包](https://cloud.tencent.com/document/product/684/15245#bwptype)和[IP带宽包](https://cloud.tencent.com/document/product/684/15245#bwptype)。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDPARAMETERVALUE_COMBINATION = "InvalidParameterValue.Combination"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TAGNOTEXISTED = "InvalidParameterValue.TagNotExisted"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDRESOURCEINTERNETCHARGETYPE = "UnsupportedOperation.InvalidResourceInternetChargeType"
func (c *Client) CreateBandwidthPackageWithContext(ctx context.Context, request *CreateBandwidthPackageRequest) (response *CreateBandwidthPackageResponse, err error) {
	if request == nil {
		request = NewCreateBandwidthPackageRequest()
	}
	request.SetContext(ctx)

	response = NewCreateBandwidthPackageResponse()
	err = c.Send(request, response)
	return
}

func NewCreateCcnRequest() (request *CreateCcnRequest) {
	request = &CreateCcnRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateCcn")

	return
}

func NewCreateCcnResponse() (response *CreateCcnResponse) {
	response = &CreateCcnResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateCcn
// 本接口（CreateCcn）用于创建云联网（CCN）。<br />
//
// * 创建云联网同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 每个账号能创建的云联网实例个数是有限的，详请参考产品文档。如果需要扩充请联系在线客服。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	UNAUTHORIZEDOPERATION_NOREALNAMEAUTHENTICATION = "UnauthorizedOperation.NoRealNameAuthentication"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INSUFFICIENTFUNDS = "UnsupportedOperation.InsufficientFunds"
//	UNSUPPORTEDOPERATION_PREPAIDCCNONLYSUPPORTINTERREGIONLIMIT = "UnsupportedOperation.PrepaidCcnOnlySupportInterRegionLimit"
//	UNSUPPORTEDOPERATION_USERANDCCNCHARGETYPENOTMATCH = "UnsupportedOperation.UserAndCcnChargeTypeNotMatch"
func (c *Client) CreateCcn(request *CreateCcnRequest) (response *CreateCcnResponse, err error) {
	if request == nil {
		request = NewCreateCcnRequest()
	}

	response = NewCreateCcnResponse()
	err = c.Send(request, response)
	return
}

// CreateCcn
// 本接口（CreateCcn）用于创建云联网（CCN）。<br />
//
// * 创建云联网同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 每个账号能创建的云联网实例个数是有限的，详请参考产品文档。如果需要扩充请联系在线客服。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	UNAUTHORIZEDOPERATION_NOREALNAMEAUTHENTICATION = "UnauthorizedOperation.NoRealNameAuthentication"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INSUFFICIENTFUNDS = "UnsupportedOperation.InsufficientFunds"
//	UNSUPPORTEDOPERATION_PREPAIDCCNONLYSUPPORTINTERREGIONLIMIT = "UnsupportedOperation.PrepaidCcnOnlySupportInterRegionLimit"
//	UNSUPPORTEDOPERATION_USERANDCCNCHARGETYPENOTMATCH = "UnsupportedOperation.UserAndCcnChargeTypeNotMatch"
func (c *Client) CreateCcnWithContext(ctx context.Context, request *CreateCcnRequest) (response *CreateCcnResponse, err error) {
	if request == nil {
		request = NewCreateCcnRequest()
	}
	request.SetContext(ctx)

	response = NewCreateCcnResponse()
	err = c.Send(request, response)
	return
}

func NewCreateCustomerGatewayRequest() (request *CreateCustomerGatewayRequest) {
	request = &CreateCustomerGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateCustomerGateway")

	return
}

func NewCreateCustomerGatewayResponse() (response *CreateCustomerGatewayResponse) {
	response = &CreateCustomerGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateCustomerGateway
// 本接口（CreateCustomerGateway）用于创建对端网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	VPCLIMITEXCEEDED = "VpcLimitExceeded"
func (c *Client) CreateCustomerGateway(request *CreateCustomerGatewayRequest) (response *CreateCustomerGatewayResponse, err error) {
	if request == nil {
		request = NewCreateCustomerGatewayRequest()
	}

	response = NewCreateCustomerGatewayResponse()
	err = c.Send(request, response)
	return
}

// CreateCustomerGateway
// 本接口（CreateCustomerGateway）用于创建对端网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	VPCLIMITEXCEEDED = "VpcLimitExceeded"
func (c *Client) CreateCustomerGatewayWithContext(ctx context.Context, request *CreateCustomerGatewayRequest) (response *CreateCustomerGatewayResponse, err error) {
	if request == nil {
		request = NewCreateCustomerGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewCreateCustomerGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewCreateDefaultSecurityGroupRequest() (request *CreateDefaultSecurityGroupRequest) {
	request = &CreateDefaultSecurityGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateDefaultSecurityGroup")

	return
}

func NewCreateDefaultSecurityGroupResponse() (response *CreateDefaultSecurityGroupResponse) {
	response = &CreateDefaultSecurityGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateDefaultSecurityGroup
// 本接口（CreateDefaultSecurityGroup）用于创建（如果项目下未存在默认安全组，则创建；已存在则获取。）默认安全组（SecurityGroup）。
//
// * 每个账户下每个地域的每个项目的<a href="https://cloud.tencent.com/document/product/213/12453">安全组数量限制</a>。
//
// * 默认安全组会放通所有IPv4规则，在创建后通常您需要再调用CreateSecurityGroupPolicies将安全组的规则设置为需要的规则。
//
// * 创建安全组同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateDefaultSecurityGroup(request *CreateDefaultSecurityGroupRequest) (response *CreateDefaultSecurityGroupResponse, err error) {
	if request == nil {
		request = NewCreateDefaultSecurityGroupRequest()
	}

	response = NewCreateDefaultSecurityGroupResponse()
	err = c.Send(request, response)
	return
}

// CreateDefaultSecurityGroup
// 本接口（CreateDefaultSecurityGroup）用于创建（如果项目下未存在默认安全组，则创建；已存在则获取。）默认安全组（SecurityGroup）。
//
// * 每个账户下每个地域的每个项目的<a href="https://cloud.tencent.com/document/product/213/12453">安全组数量限制</a>。
//
// * 默认安全组会放通所有IPv4规则，在创建后通常您需要再调用CreateSecurityGroupPolicies将安全组的规则设置为需要的规则。
//
// * 创建安全组同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateDefaultSecurityGroupWithContext(ctx context.Context, request *CreateDefaultSecurityGroupRequest) (response *CreateDefaultSecurityGroupResponse, err error) {
	if request == nil {
		request = NewCreateDefaultSecurityGroupRequest()
	}
	request.SetContext(ctx)

	response = NewCreateDefaultSecurityGroupResponse()
	err = c.Send(request, response)
	return
}

func NewCreateDefaultVpcRequest() (request *CreateDefaultVpcRequest) {
	request = &CreateDefaultVpcRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateDefaultVpc")

	return
}

func NewCreateDefaultVpcResponse() (response *CreateDefaultVpcResponse) {
	response = &CreateDefaultVpcResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateDefaultVpc
// 本接口（CreateDefaultVpc）用于创建默认私有网络(VPC）。
//
// 默认VPC适用于快速入门和启动公共实例，您可以像使用任何其他VPC一样使用默认VPC。如果您想创建标准VPC，即指定VPC名称、VPC网段、子网网段、子网可用区，请使用常规创建VPC接口（CreateVpc）
//
// 正常情况，本接口并不一定生产默认VPC，而是根据用户账号的网络属性（DescribeAccountAttributes）来决定的
//
// * 支持基础网络、VPC，返回VpcId为0
//
// * 只支持VPC，返回默认VPC信息
//
// 您也可以通过 Force 参数，强制返回默认VPC
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINSUFFICIENT_CIDRBLOCK = "ResourceInsufficient.CidrBlock"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateDefaultVpc(request *CreateDefaultVpcRequest) (response *CreateDefaultVpcResponse, err error) {
	if request == nil {
		request = NewCreateDefaultVpcRequest()
	}

	response = NewCreateDefaultVpcResponse()
	err = c.Send(request, response)
	return
}

// CreateDefaultVpc
// 本接口（CreateDefaultVpc）用于创建默认私有网络(VPC）。
//
// 默认VPC适用于快速入门和启动公共实例，您可以像使用任何其他VPC一样使用默认VPC。如果您想创建标准VPC，即指定VPC名称、VPC网段、子网网段、子网可用区，请使用常规创建VPC接口（CreateVpc）
//
// 正常情况，本接口并不一定生产默认VPC，而是根据用户账号的网络属性（DescribeAccountAttributes）来决定的
//
// * 支持基础网络、VPC，返回VpcId为0
//
// * 只支持VPC，返回默认VPC信息
//
// 您也可以通过 Force 参数，强制返回默认VPC
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINSUFFICIENT_CIDRBLOCK = "ResourceInsufficient.CidrBlock"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateDefaultVpcWithContext(ctx context.Context, request *CreateDefaultVpcRequest) (response *CreateDefaultVpcResponse, err error) {
	if request == nil {
		request = NewCreateDefaultVpcRequest()
	}
	request.SetContext(ctx)

	response = NewCreateDefaultVpcResponse()
	err = c.Send(request, response)
	return
}

func NewCreateDhcpIpRequest() (request *CreateDhcpIpRequest) {
	request = &CreateDhcpIpRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateDhcpIp")

	return
}

func NewCreateDhcpIpResponse() (response *CreateDhcpIpResponse) {
	response = &CreateDhcpIpResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateDhcpIp
// 本接口（CreateDhcpIp）用于创建DhcpIp
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateDhcpIp(request *CreateDhcpIpRequest) (response *CreateDhcpIpResponse, err error) {
	if request == nil {
		request = NewCreateDhcpIpRequest()
	}

	response = NewCreateDhcpIpResponse()
	err = c.Send(request, response)
	return
}

// CreateDhcpIp
// 本接口（CreateDhcpIp）用于创建DhcpIp
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateDhcpIpWithContext(ctx context.Context, request *CreateDhcpIpRequest) (response *CreateDhcpIpResponse, err error) {
	if request == nil {
		request = NewCreateDhcpIpRequest()
	}
	request.SetContext(ctx)

	response = NewCreateDhcpIpResponse()
	err = c.Send(request, response)
	return
}

func NewCreateDirectConnectGatewayRequest() (request *CreateDirectConnectGatewayRequest) {
	request = &CreateDirectConnectGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateDirectConnectGateway")

	return
}

func NewCreateDirectConnectGatewayResponse() (response *CreateDirectConnectGatewayResponse) {
	response = &CreateDirectConnectGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateDirectConnectGateway
// 本接口（CreateDirectConnectGateway）用于创建专线网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_VPGHAGROUPNOTFOUND = "InvalidParameter.VpgHaGroupNotFound"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_UNABLECROSSBORDER = "UnsupportedOperation.UnableCrossBorder"
func (c *Client) CreateDirectConnectGateway(request *CreateDirectConnectGatewayRequest) (response *CreateDirectConnectGatewayResponse, err error) {
	if request == nil {
		request = NewCreateDirectConnectGatewayRequest()
	}

	response = NewCreateDirectConnectGatewayResponse()
	err = c.Send(request, response)
	return
}

// CreateDirectConnectGateway
// 本接口（CreateDirectConnectGateway）用于创建专线网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_VPGHAGROUPNOTFOUND = "InvalidParameter.VpgHaGroupNotFound"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_UNABLECROSSBORDER = "UnsupportedOperation.UnableCrossBorder"
func (c *Client) CreateDirectConnectGatewayWithContext(ctx context.Context, request *CreateDirectConnectGatewayRequest) (response *CreateDirectConnectGatewayResponse, err error) {
	if request == nil {
		request = NewCreateDirectConnectGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewCreateDirectConnectGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewCreateDirectConnectGatewayCcnRoutesRequest() (request *CreateDirectConnectGatewayCcnRoutesRequest) {
	request = &CreateDirectConnectGatewayCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateDirectConnectGatewayCcnRoutes")

	return
}

func NewCreateDirectConnectGatewayCcnRoutesResponse() (response *CreateDirectConnectGatewayCcnRoutesResponse) {
	response = &CreateDirectConnectGatewayCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateDirectConnectGatewayCcnRoutes
// 本接口（CreateDirectConnectGatewayCcnRoutes）用于创建专线网关的云联网路由（IDC网段）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateDirectConnectGatewayCcnRoutes(request *CreateDirectConnectGatewayCcnRoutesRequest) (response *CreateDirectConnectGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewCreateDirectConnectGatewayCcnRoutesRequest()
	}

	response = NewCreateDirectConnectGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// CreateDirectConnectGatewayCcnRoutes
// 本接口（CreateDirectConnectGatewayCcnRoutes）用于创建专线网关的云联网路由（IDC网段）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateDirectConnectGatewayCcnRoutesWithContext(ctx context.Context, request *CreateDirectConnectGatewayCcnRoutesRequest) (response *CreateDirectConnectGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewCreateDirectConnectGatewayCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewCreateDirectConnectGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewCreateFlowLogRequest() (request *CreateFlowLogRequest) {
	request = &CreateFlowLogRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateFlowLog")

	return
}

func NewCreateFlowLogResponse() (response *CreateFlowLogResponse) {
	response = &CreateFlowLogResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateFlowLog
// 本接口（CreateFlowLog）用于创建流日志
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) CreateFlowLog(request *CreateFlowLogRequest) (response *CreateFlowLogResponse, err error) {
	if request == nil {
		request = NewCreateFlowLogRequest()
	}

	response = NewCreateFlowLogResponse()
	err = c.Send(request, response)
	return
}

// CreateFlowLog
// 本接口（CreateFlowLog）用于创建流日志
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) CreateFlowLogWithContext(ctx context.Context, request *CreateFlowLogRequest) (response *CreateFlowLogResponse, err error) {
	if request == nil {
		request = NewCreateFlowLogRequest()
	}
	request.SetContext(ctx)

	response = NewCreateFlowLogResponse()
	err = c.Send(request, response)
	return
}

func NewCreateHaVipRequest() (request *CreateHaVipRequest) {
	request = &CreateHaVipRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateHaVip")

	return
}

func NewCreateHaVipResponse() (response *CreateHaVipResponse) {
	response = &CreateHaVipResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateHaVip
// 本接口（CreateHaVip）用于创建高可用虚拟IP（HAVIP）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_INVALIDBUSINESS = "InvalidParameterValue.InvalidBusiness"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateHaVip(request *CreateHaVipRequest) (response *CreateHaVipResponse, err error) {
	if request == nil {
		request = NewCreateHaVipRequest()
	}

	response = NewCreateHaVipResponse()
	err = c.Send(request, response)
	return
}

// CreateHaVip
// 本接口（CreateHaVip）用于创建高可用虚拟IP（HAVIP）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_INVALIDBUSINESS = "InvalidParameterValue.InvalidBusiness"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateHaVipWithContext(ctx context.Context, request *CreateHaVipRequest) (response *CreateHaVipResponse, err error) {
	if request == nil {
		request = NewCreateHaVipRequest()
	}
	request.SetContext(ctx)

	response = NewCreateHaVipResponse()
	err = c.Send(request, response)
	return
}

func NewCreateIp6TranslatorsRequest() (request *CreateIp6TranslatorsRequest) {
	request = &CreateIp6TranslatorsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateIp6Translators")

	return
}

func NewCreateIp6TranslatorsResponse() (response *CreateIp6TranslatorsResponse) {
	response = &CreateIp6TranslatorsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateIp6Translators
// 1. 该接口用于创建IPV6转换IPV4实例，支持批量
//
// 2. 同一个账户在一个地域最多允许创建10个转换实例
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
func (c *Client) CreateIp6Translators(request *CreateIp6TranslatorsRequest) (response *CreateIp6TranslatorsResponse, err error) {
	if request == nil {
		request = NewCreateIp6TranslatorsRequest()
	}

	response = NewCreateIp6TranslatorsResponse()
	err = c.Send(request, response)
	return
}

// CreateIp6Translators
// 1. 该接口用于创建IPV6转换IPV4实例，支持批量
//
// 2. 同一个账户在一个地域最多允许创建10个转换实例
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
func (c *Client) CreateIp6TranslatorsWithContext(ctx context.Context, request *CreateIp6TranslatorsRequest) (response *CreateIp6TranslatorsResponse, err error) {
	if request == nil {
		request = NewCreateIp6TranslatorsRequest()
	}
	request.SetContext(ctx)

	response = NewCreateIp6TranslatorsResponse()
	err = c.Send(request, response)
	return
}

func NewCreateLocalGatewayRequest() (request *CreateLocalGatewayRequest) {
	request = &CreateLocalGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateLocalGateway")

	return
}

func NewCreateLocalGatewayResponse() (response *CreateLocalGatewayResponse) {
	response = &CreateLocalGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateLocalGateway
// 该接口用于创建用于CDC的本地网关。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_LOCALGATEWAYALREADYEXISTS = "UnsupportedOperation.LocalGatewayAlreadyExists"
func (c *Client) CreateLocalGateway(request *CreateLocalGatewayRequest) (response *CreateLocalGatewayResponse, err error) {
	if request == nil {
		request = NewCreateLocalGatewayRequest()
	}

	response = NewCreateLocalGatewayResponse()
	err = c.Send(request, response)
	return
}

// CreateLocalGateway
// 该接口用于创建用于CDC的本地网关。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_LOCALGATEWAYALREADYEXISTS = "UnsupportedOperation.LocalGatewayAlreadyExists"
func (c *Client) CreateLocalGatewayWithContext(ctx context.Context, request *CreateLocalGatewayRequest) (response *CreateLocalGatewayResponse, err error) {
	if request == nil {
		request = NewCreateLocalGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewCreateLocalGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewCreateNatGatewayRequest() (request *CreateNatGatewayRequest) {
	request = &CreateNatGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateNatGateway")

	return
}

func NewCreateNatGatewayResponse() (response *CreateNatGatewayResponse) {
	response = &CreateNatGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateNatGateway
// 本接口(CreateNatGateway)用于创建NAT网关。
//
// 在对新建的NAT网关做其他操作前，需先确认此网关已被创建完成（DescribeNatGateway接口返回的实例State字段为AVAILABLE）。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED = "AddressQuotaLimitExceeded"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSSTATE = "InvalidAddressState"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDVPCID_MALFORMED = "InvalidVpcId.Malformed"
//	INVALIDVPCID_NOTFOUND = "InvalidVpcId.NotFound"
//	LIMITEXCEEDED_ADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.AddressQuotaLimitExceeded"
//	LIMITEXCEEDED_DAILYALLOCATEADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.DailyAllocateAddressQuotaLimitExceeded"
//	LIMITEXCEEDED_NATGATEWAYPERVPCLIMITEXCEEDED = "LimitExceeded.NatGatewayPerVpcLimitExceeded"
//	LIMITEXCEEDED_PUBLICIPADDRESSPERNATGATEWAYLIMITEXCEEDED = "LimitExceeded.PublicIpAddressPerNatGatewayLimitExceeded"
//	RESOURCEINUSE_ADDRESS = "ResourceInUse.Address"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_NOREALNAMEAUTHENTICATION = "UnauthorizedOperation.NoRealNameAuthentication"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INSUFFICIENTFUNDS = "UnsupportedOperation.InsufficientFunds"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTBGPIP = "UnsupportedOperation.PublicIpAddressIsNotBGPIp"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTEXISTED = "UnsupportedOperation.PublicIpAddressIsNotExisted"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSNOTBILLEDBYTRAFFIC = "UnsupportedOperation.PublicIpAddressNotBilledByTraffic"
func (c *Client) CreateNatGateway(request *CreateNatGatewayRequest) (response *CreateNatGatewayResponse, err error) {
	if request == nil {
		request = NewCreateNatGatewayRequest()
	}

	response = NewCreateNatGatewayResponse()
	err = c.Send(request, response)
	return
}

// CreateNatGateway
// 本接口(CreateNatGateway)用于创建NAT网关。
//
// 在对新建的NAT网关做其他操作前，需先确认此网关已被创建完成（DescribeNatGateway接口返回的实例State字段为AVAILABLE）。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED = "AddressQuotaLimitExceeded"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSSTATE = "InvalidAddressState"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDVPCID_MALFORMED = "InvalidVpcId.Malformed"
//	INVALIDVPCID_NOTFOUND = "InvalidVpcId.NotFound"
//	LIMITEXCEEDED_ADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.AddressQuotaLimitExceeded"
//	LIMITEXCEEDED_DAILYALLOCATEADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.DailyAllocateAddressQuotaLimitExceeded"
//	LIMITEXCEEDED_NATGATEWAYPERVPCLIMITEXCEEDED = "LimitExceeded.NatGatewayPerVpcLimitExceeded"
//	LIMITEXCEEDED_PUBLICIPADDRESSPERNATGATEWAYLIMITEXCEEDED = "LimitExceeded.PublicIpAddressPerNatGatewayLimitExceeded"
//	RESOURCEINUSE_ADDRESS = "ResourceInUse.Address"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_NOREALNAMEAUTHENTICATION = "UnauthorizedOperation.NoRealNameAuthentication"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INSUFFICIENTFUNDS = "UnsupportedOperation.InsufficientFunds"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTBGPIP = "UnsupportedOperation.PublicIpAddressIsNotBGPIp"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTEXISTED = "UnsupportedOperation.PublicIpAddressIsNotExisted"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSNOTBILLEDBYTRAFFIC = "UnsupportedOperation.PublicIpAddressNotBilledByTraffic"
func (c *Client) CreateNatGatewayWithContext(ctx context.Context, request *CreateNatGatewayRequest) (response *CreateNatGatewayResponse, err error) {
	if request == nil {
		request = NewCreateNatGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewCreateNatGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewCreateNatGatewayDestinationIpPortTranslationNatRuleRequest() (request *CreateNatGatewayDestinationIpPortTranslationNatRuleRequest) {
	request = &CreateNatGatewayDestinationIpPortTranslationNatRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateNatGatewayDestinationIpPortTranslationNatRule")

	return
}

func NewCreateNatGatewayDestinationIpPortTranslationNatRuleResponse() (response *CreateNatGatewayDestinationIpPortTranslationNatRuleResponse) {
	response = &CreateNatGatewayDestinationIpPortTranslationNatRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateNatGatewayDestinationIpPortTranslationNatRule
// 本接口(CreateNatGatewayDestinationIpPortTranslationNatRule)用于创建NAT网关端口转发规则。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED = "AddressQuotaLimitExceeded"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSSTATE = "InvalidAddressState"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDVPCID_MALFORMED = "InvalidVpcId.Malformed"
//	INVALIDVPCID_NOTFOUND = "InvalidVpcId.NotFound"
//	LIMITEXCEEDED_ADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.AddressQuotaLimitExceeded"
//	LIMITEXCEEDED_DAILYALLOCATEADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.DailyAllocateAddressQuotaLimitExceeded"
//	LIMITEXCEEDED_NATGATEWAYPERVPCLIMITEXCEEDED = "LimitExceeded.NatGatewayPerVpcLimitExceeded"
//	LIMITEXCEEDED_PUBLICIPADDRESSPERNATGATEWAYLIMITEXCEEDED = "LimitExceeded.PublicIpAddressPerNatGatewayLimitExceeded"
//	RESOURCEINUSE_ADDRESS = "ResourceInUse.Address"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_NOREALNAMEAUTHENTICATION = "UnauthorizedOperation.NoRealNameAuthentication"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INSUFFICIENTFUNDS = "UnsupportedOperation.InsufficientFunds"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTBGPIP = "UnsupportedOperation.PublicIpAddressIsNotBGPIp"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTEXISTED = "UnsupportedOperation.PublicIpAddressIsNotExisted"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSNOTBILLEDBYTRAFFIC = "UnsupportedOperation.PublicIpAddressNotBilledByTraffic"
func (c *Client) CreateNatGatewayDestinationIpPortTranslationNatRule(request *CreateNatGatewayDestinationIpPortTranslationNatRuleRequest) (response *CreateNatGatewayDestinationIpPortTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewCreateNatGatewayDestinationIpPortTranslationNatRuleRequest()
	}

	response = NewCreateNatGatewayDestinationIpPortTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

// CreateNatGatewayDestinationIpPortTranslationNatRule
// 本接口(CreateNatGatewayDestinationIpPortTranslationNatRule)用于创建NAT网关端口转发规则。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED = "AddressQuotaLimitExceeded"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSSTATE = "InvalidAddressState"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDVPCID_MALFORMED = "InvalidVpcId.Malformed"
//	INVALIDVPCID_NOTFOUND = "InvalidVpcId.NotFound"
//	LIMITEXCEEDED_ADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.AddressQuotaLimitExceeded"
//	LIMITEXCEEDED_DAILYALLOCATEADDRESSQUOTALIMITEXCEEDED = "LimitExceeded.DailyAllocateAddressQuotaLimitExceeded"
//	LIMITEXCEEDED_NATGATEWAYPERVPCLIMITEXCEEDED = "LimitExceeded.NatGatewayPerVpcLimitExceeded"
//	LIMITEXCEEDED_PUBLICIPADDRESSPERNATGATEWAYLIMITEXCEEDED = "LimitExceeded.PublicIpAddressPerNatGatewayLimitExceeded"
//	RESOURCEINUSE_ADDRESS = "ResourceInUse.Address"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_NOREALNAMEAUTHENTICATION = "UnauthorizedOperation.NoRealNameAuthentication"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INSUFFICIENTFUNDS = "UnsupportedOperation.InsufficientFunds"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTBGPIP = "UnsupportedOperation.PublicIpAddressIsNotBGPIp"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSISNOTEXISTED = "UnsupportedOperation.PublicIpAddressIsNotExisted"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSNOTBILLEDBYTRAFFIC = "UnsupportedOperation.PublicIpAddressNotBilledByTraffic"
func (c *Client) CreateNatGatewayDestinationIpPortTranslationNatRuleWithContext(ctx context.Context, request *CreateNatGatewayDestinationIpPortTranslationNatRuleRequest) (response *CreateNatGatewayDestinationIpPortTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewCreateNatGatewayDestinationIpPortTranslationNatRuleRequest()
	}
	request.SetContext(ctx)

	response = NewCreateNatGatewayDestinationIpPortTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

func NewCreateNatGatewaySourceIpTranslationNatRuleRequest() (request *CreateNatGatewaySourceIpTranslationNatRuleRequest) {
	request = &CreateNatGatewaySourceIpTranslationNatRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateNatGatewaySourceIpTranslationNatRule")

	return
}

func NewCreateNatGatewaySourceIpTranslationNatRuleResponse() (response *CreateNatGatewaySourceIpTranslationNatRuleResponse) {
	response = &CreateNatGatewaySourceIpTranslationNatRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateNatGatewaySourceIpTranslationNatRule
// 本接口(CreateNatGatewaySourceIpTranslationNatRule)用于创建NAT网关SNAT规则
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_NATSNATRULEEXISTS = "InvalidParameterValue.NatSnatRuleExists"
//	LIMITEXCEEDED = "LimitExceeded"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION_NATGATEWAYTYPENOTSUPPORTSNAT = "UnsupportedOperation.NatGatewayTypeNotSupportSNAT"
func (c *Client) CreateNatGatewaySourceIpTranslationNatRule(request *CreateNatGatewaySourceIpTranslationNatRuleRequest) (response *CreateNatGatewaySourceIpTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewCreateNatGatewaySourceIpTranslationNatRuleRequest()
	}

	response = NewCreateNatGatewaySourceIpTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

// CreateNatGatewaySourceIpTranslationNatRule
// 本接口(CreateNatGatewaySourceIpTranslationNatRule)用于创建NAT网关SNAT规则
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_NATSNATRULEEXISTS = "InvalidParameterValue.NatSnatRuleExists"
//	LIMITEXCEEDED = "LimitExceeded"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION_NATGATEWAYTYPENOTSUPPORTSNAT = "UnsupportedOperation.NatGatewayTypeNotSupportSNAT"
func (c *Client) CreateNatGatewaySourceIpTranslationNatRuleWithContext(ctx context.Context, request *CreateNatGatewaySourceIpTranslationNatRuleRequest) (response *CreateNatGatewaySourceIpTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewCreateNatGatewaySourceIpTranslationNatRuleRequest()
	}
	request.SetContext(ctx)

	response = NewCreateNatGatewaySourceIpTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

func NewCreateNetDetectRequest() (request *CreateNetDetectRequest) {
	request = &CreateNetDetectRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateNetDetect")

	return
}

func NewCreateNetDetectResponse() (response *CreateNetDetectResponse) {
	response = &CreateNetDetectResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateNetDetect
// 本接口(CreateNetDetect)用于创建网络探测。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_NEXTHOPMISMATCH = "InvalidParameter.NextHopMismatch"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_NETDETECTINVPC = "InvalidParameterValue.NetDetectInVpc"
//	INVALIDPARAMETERVALUE_NETDETECTNOTFOUNDIP = "InvalidParameterValue.NetDetectNotFoundIp"
//	INVALIDPARAMETERVALUE_NETDETECTSAMEIP = "InvalidParameterValue.NetDetectSameIp"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) CreateNetDetect(request *CreateNetDetectRequest) (response *CreateNetDetectResponse, err error) {
	if request == nil {
		request = NewCreateNetDetectRequest()
	}

	response = NewCreateNetDetectResponse()
	err = c.Send(request, response)
	return
}

// CreateNetDetect
// 本接口(CreateNetDetect)用于创建网络探测。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_NEXTHOPMISMATCH = "InvalidParameter.NextHopMismatch"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_NETDETECTINVPC = "InvalidParameterValue.NetDetectInVpc"
//	INVALIDPARAMETERVALUE_NETDETECTNOTFOUNDIP = "InvalidParameterValue.NetDetectNotFoundIp"
//	INVALIDPARAMETERVALUE_NETDETECTSAMEIP = "InvalidParameterValue.NetDetectSameIp"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) CreateNetDetectWithContext(ctx context.Context, request *CreateNetDetectRequest) (response *CreateNetDetectResponse, err error) {
	if request == nil {
		request = NewCreateNetDetectRequest()
	}
	request.SetContext(ctx)

	response = NewCreateNetDetectResponse()
	err = c.Send(request, response)
	return
}

func NewCreateNetworkAclRequest() (request *CreateNetworkAclRequest) {
	request = &CreateNetworkAclRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateNetworkAcl")

	return
}

func NewCreateNetworkAclResponse() (response *CreateNetworkAclResponse) {
	response = &CreateNetworkAclResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateNetworkAcl
// 本接口（CreateNetworkAcl）用于创建新的<a href="https://cloud.tencent.com/document/product/215/20088">网络ACL</a>。
//
// * 新建的网络ACL的入站和出站规则默认都是全部拒绝，在创建后通常您需要再调用ModifyNetworkAclEntries将网络ACL的规则设置为需要的规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateNetworkAcl(request *CreateNetworkAclRequest) (response *CreateNetworkAclResponse, err error) {
	if request == nil {
		request = NewCreateNetworkAclRequest()
	}

	response = NewCreateNetworkAclResponse()
	err = c.Send(request, response)
	return
}

// CreateNetworkAcl
// 本接口（CreateNetworkAcl）用于创建新的<a href="https://cloud.tencent.com/document/product/215/20088">网络ACL</a>。
//
// * 新建的网络ACL的入站和出站规则默认都是全部拒绝，在创建后通常您需要再调用ModifyNetworkAclEntries将网络ACL的规则设置为需要的规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateNetworkAclWithContext(ctx context.Context, request *CreateNetworkAclRequest) (response *CreateNetworkAclResponse, err error) {
	if request == nil {
		request = NewCreateNetworkAclRequest()
	}
	request.SetContext(ctx)

	response = NewCreateNetworkAclResponse()
	err = c.Send(request, response)
	return
}

func NewCreateNetworkInterfaceRequest() (request *CreateNetworkInterfaceRequest) {
	request = &CreateNetworkInterfaceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateNetworkInterface")

	return
}

func NewCreateNetworkInterfaceResponse() (response *CreateNetworkInterfaceResponse) {
	response = &CreateNetworkInterfaceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateNetworkInterface
// 本接口（CreateNetworkInterface）用于创建弹性网卡。
//
// * 创建弹性网卡时可以指定内网IP，并且可以指定一个主IP，指定的内网IP必须在弹性网卡所在子网内，而且不能被占用。
//
// * 创建弹性网卡时可以指定需要申请的内网IP数量，系统会随机生成内网IP地址。
//
// * 一个弹性网卡支持绑定的IP地址是有限制的，更多资源限制信息详见<a href="/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 创建弹性网卡同时可以绑定已有安全组。
//
// * 创建弹性网卡同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_RESERVED = "InvalidParameterValue.Reserved"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateNetworkInterface(request *CreateNetworkInterfaceRequest) (response *CreateNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewCreateNetworkInterfaceRequest()
	}

	response = NewCreateNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

// CreateNetworkInterface
// 本接口（CreateNetworkInterface）用于创建弹性网卡。
//
// * 创建弹性网卡时可以指定内网IP，并且可以指定一个主IP，指定的内网IP必须在弹性网卡所在子网内，而且不能被占用。
//
// * 创建弹性网卡时可以指定需要申请的内网IP数量，系统会随机生成内网IP地址。
//
// * 一个弹性网卡支持绑定的IP地址是有限制的，更多资源限制信息详见<a href="/document/product/576/18527">弹性网卡使用限制</a>。
//
// * 创建弹性网卡同时可以绑定已有安全组。
//
// * 创建弹性网卡同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_RESERVED = "InvalidParameterValue.Reserved"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateNetworkInterfaceWithContext(ctx context.Context, request *CreateNetworkInterfaceRequest) (response *CreateNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewCreateNetworkInterfaceRequest()
	}
	request.SetContext(ctx)

	response = NewCreateNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

func NewCreateRouteTableRequest() (request *CreateRouteTableRequest) {
	request = &CreateRouteTableRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateRouteTable")

	return
}

func NewCreateRouteTableResponse() (response *CreateRouteTableResponse) {
	response = &CreateRouteTableResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateRouteTable
// 本接口(CreateRouteTable)用于创建路由表。
//
// * 创建了VPC后，系统会创建一个默认路由表，所有新建的子网都会关联到默认路由表。默认情况下您可以直接使用默认路由表来管理您的路由策略。当您的路由策略较多时，您可以调用创建路由表接口创建更多路由表管理您的路由策略。
//
// * 创建路由表同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateRouteTable(request *CreateRouteTableRequest) (response *CreateRouteTableResponse, err error) {
	if request == nil {
		request = NewCreateRouteTableRequest()
	}

	response = NewCreateRouteTableResponse()
	err = c.Send(request, response)
	return
}

// CreateRouteTable
// 本接口(CreateRouteTable)用于创建路由表。
//
// * 创建了VPC后，系统会创建一个默认路由表，所有新建的子网都会关联到默认路由表。默认情况下您可以直接使用默认路由表来管理您的路由策略。当您的路由策略较多时，您可以调用创建路由表接口创建更多路由表管理您的路由策略。
//
// * 创建路由表同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateRouteTableWithContext(ctx context.Context, request *CreateRouteTableRequest) (response *CreateRouteTableResponse, err error) {
	if request == nil {
		request = NewCreateRouteTableRequest()
	}
	request.SetContext(ctx)

	response = NewCreateRouteTableResponse()
	err = c.Send(request, response)
	return
}

func NewCreateRoutesRequest() (request *CreateRoutesRequest) {
	request = &CreateRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateRoutes")

	return
}

func NewCreateRoutesResponse() (response *CreateRoutesResponse) {
	response = &CreateRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateRoutes
// 本接口(CreateRoutes)用于创建路由策略。
//
// * 向指定路由表批量新增路由策略。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CIDRNOTINPEERVPC = "InvalidParameterValue.CidrNotInPeerVpc"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER_WITHGUESS = "UnknownParameter.WithGuess"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CDCSUBNETNOTSUPPORTUNLOCALGATEWAY = "UnsupportedOperation.CdcSubnetNotSupportUnLocalGateway"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
//	UNSUPPORTEDOPERATION_ECMPWITHCCNROUTE = "UnsupportedOperation.EcmpWithCcnRoute"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
//	UNSUPPORTEDOPERATION_NORMALSUBNETNOTSUPPORTLOCALGATEWAY = "UnsupportedOperation.NormalSubnetNotSupportLocalGateway"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) CreateRoutes(request *CreateRoutesRequest) (response *CreateRoutesResponse, err error) {
	if request == nil {
		request = NewCreateRoutesRequest()
	}

	response = NewCreateRoutesResponse()
	err = c.Send(request, response)
	return
}

// CreateRoutes
// 本接口(CreateRoutes)用于创建路由策略。
//
// * 向指定路由表批量新增路由策略。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CIDRNOTINPEERVPC = "InvalidParameterValue.CidrNotInPeerVpc"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER_WITHGUESS = "UnknownParameter.WithGuess"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CDCSUBNETNOTSUPPORTUNLOCALGATEWAY = "UnsupportedOperation.CdcSubnetNotSupportUnLocalGateway"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
//	UNSUPPORTEDOPERATION_ECMPWITHCCNROUTE = "UnsupportedOperation.EcmpWithCcnRoute"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
//	UNSUPPORTEDOPERATION_NORMALSUBNETNOTSUPPORTLOCALGATEWAY = "UnsupportedOperation.NormalSubnetNotSupportLocalGateway"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) CreateRoutesWithContext(ctx context.Context, request *CreateRoutesRequest) (response *CreateRoutesResponse, err error) {
	if request == nil {
		request = NewCreateRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewCreateRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewCreateSecurityGroupRequest() (request *CreateSecurityGroupRequest) {
	request = &CreateSecurityGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateSecurityGroup")

	return
}

func NewCreateSecurityGroupResponse() (response *CreateSecurityGroupResponse) {
	response = &CreateSecurityGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateSecurityGroup
// 本接口（CreateSecurityGroup）用于创建新的安全组（SecurityGroup）。
//
// * 每个账户下每个地域的每个项目的<a href="https://cloud.tencent.com/document/product/213/12453">安全组数量限制</a>。
//
// * 新建的安全组的入站和出站规则默认都是全部拒绝，在创建后通常您需要再调用CreateSecurityGroupPolicies将安全组的规则设置为需要的规则。
//
// * 创建安全组同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateSecurityGroup(request *CreateSecurityGroupRequest) (response *CreateSecurityGroupResponse, err error) {
	if request == nil {
		request = NewCreateSecurityGroupRequest()
	}

	response = NewCreateSecurityGroupResponse()
	err = c.Send(request, response)
	return
}

// CreateSecurityGroup
// 本接口（CreateSecurityGroup）用于创建新的安全组（SecurityGroup）。
//
// * 每个账户下每个地域的每个项目的<a href="https://cloud.tencent.com/document/product/213/12453">安全组数量限制</a>。
//
// * 新建的安全组的入站和出站规则默认都是全部拒绝，在创建后通常您需要再调用CreateSecurityGroupPolicies将安全组的规则设置为需要的规则。
//
// * 创建安全组同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateSecurityGroupWithContext(ctx context.Context, request *CreateSecurityGroupRequest) (response *CreateSecurityGroupResponse, err error) {
	if request == nil {
		request = NewCreateSecurityGroupRequest()
	}
	request.SetContext(ctx)

	response = NewCreateSecurityGroupResponse()
	err = c.Send(request, response)
	return
}

func NewCreateSecurityGroupPoliciesRequest() (request *CreateSecurityGroupPoliciesRequest) {
	request = &CreateSecurityGroupPoliciesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateSecurityGroupPolicies")

	return
}

func NewCreateSecurityGroupPoliciesResponse() (response *CreateSecurityGroupPoliciesResponse) {
	response = &CreateSecurityGroupPoliciesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateSecurityGroupPolicies
// 本接口（CreateSecurityGroupPolicies）用于创建安全组规则（SecurityGroupPolicy）。
//
// 在 SecurityGroupPolicySet 参数中：
//
// <ul>
//
// <li>Version 安全组规则版本号，用户每次更新安全规则版本会自动加1，防止您更新的路由规则已过期，不填不考虑冲突。</li>
//
// <li>在创建出站和入站规则（Egress 和 Ingress）时：<ul>
//
// <li>Protocol 字段支持输入TCP, UDP, ICMP, ICMPV6, GRE, ALL。</li>
//
// <li>CidrBlock 字段允许输入符合cidr格式标准的任意字符串。在基础网络中，如果 CidrBlock 包含您的账户内的云服务器之外的设备在腾讯云的内网 IP，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。</li>
//
// <li>Ipv6CidrBlock 字段允许输入符合IPv6 cidr格式标准的任意字符串。在基础网络中，如果Ipv6CidrBlock 包含您的账户内的云服务器之外的设备在腾讯云的内网 IPv6，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。</li>
//
// <li>SecurityGroupId 字段允许输入与待修改的安全组位于相同项目中的安全组 ID，包括这个安全组 ID 本身，代表安全组下所有云服务器的内网 IP。使用这个字段时，这条规则用来匹配网络报文的过程中会随着被使用的这个 ID 所关联的云服务器变化而变化，不需要重新修改。</li>
//
// <li>Port 字段允许输入一个单独端口号，或者用减号分隔的两个端口号代表端口范围，例如80或8000-8010。只有当 Protocol 字段是 TCP 或 UDP 时，Port 字段才被接受，即 Protocol 字段不是 TCP 或 UDP 时，Protocol 和 Port 排他关系，不允许同时输入，否则会接口报错。</li>
//
// <li>Action 字段只允许输入 ACCEPT 或 DROP。</li>
//
// <li>CidrBlock, Ipv6CidrBlock, SecurityGroupId, AddressTemplate 四者是排他关系，不允许同时输入，Protocol + Port 和 ServiceTemplate 二者是排他关系，不允许同时输入。</li>
//
// <li>一次请求中只能创建单个方向的规则, 如果需要指定索引（PolicyIndex）参数, 多条规则的索引必须一致。</li>
//
// </ul></li></ul>
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	LIMITEXCEEDED_SECURITYGROUPPOLICYSET = "LimitExceeded.SecurityGroupPolicySet"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER_WITHGUESS = "UnknownParameter.WithGuess"
//	UNSUPPORTEDOPERATION_CLBPOLICYLIMIT = "UnsupportedOperation.ClbPolicyLimit"
//	UNSUPPORTEDOPERATION_DUPLICATEPOLICY = "UnsupportedOperation.DuplicatePolicy"
//	UNSUPPORTEDOPERATION_VERSIONMISMATCH = "UnsupportedOperation.VersionMismatch"
func (c *Client) CreateSecurityGroupPolicies(request *CreateSecurityGroupPoliciesRequest) (response *CreateSecurityGroupPoliciesResponse, err error) {
	if request == nil {
		request = NewCreateSecurityGroupPoliciesRequest()
	}

	response = NewCreateSecurityGroupPoliciesResponse()
	err = c.Send(request, response)
	return
}

// CreateSecurityGroupPolicies
// 本接口（CreateSecurityGroupPolicies）用于创建安全组规则（SecurityGroupPolicy）。
//
// 在 SecurityGroupPolicySet 参数中：
//
// <ul>
//
// <li>Version 安全组规则版本号，用户每次更新安全规则版本会自动加1，防止您更新的路由规则已过期，不填不考虑冲突。</li>
//
// <li>在创建出站和入站规则（Egress 和 Ingress）时：<ul>
//
// <li>Protocol 字段支持输入TCP, UDP, ICMP, ICMPV6, GRE, ALL。</li>
//
// <li>CidrBlock 字段允许输入符合cidr格式标准的任意字符串。在基础网络中，如果 CidrBlock 包含您的账户内的云服务器之外的设备在腾讯云的内网 IP，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。</li>
//
// <li>Ipv6CidrBlock 字段允许输入符合IPv6 cidr格式标准的任意字符串。在基础网络中，如果Ipv6CidrBlock 包含您的账户内的云服务器之外的设备在腾讯云的内网 IPv6，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。</li>
//
// <li>SecurityGroupId 字段允许输入与待修改的安全组位于相同项目中的安全组 ID，包括这个安全组 ID 本身，代表安全组下所有云服务器的内网 IP。使用这个字段时，这条规则用来匹配网络报文的过程中会随着被使用的这个 ID 所关联的云服务器变化而变化，不需要重新修改。</li>
//
// <li>Port 字段允许输入一个单独端口号，或者用减号分隔的两个端口号代表端口范围，例如80或8000-8010。只有当 Protocol 字段是 TCP 或 UDP 时，Port 字段才被接受，即 Protocol 字段不是 TCP 或 UDP 时，Protocol 和 Port 排他关系，不允许同时输入，否则会接口报错。</li>
//
// <li>Action 字段只允许输入 ACCEPT 或 DROP。</li>
//
// <li>CidrBlock, Ipv6CidrBlock, SecurityGroupId, AddressTemplate 四者是排他关系，不允许同时输入，Protocol + Port 和 ServiceTemplate 二者是排他关系，不允许同时输入。</li>
//
// <li>一次请求中只能创建单个方向的规则, 如果需要指定索引（PolicyIndex）参数, 多条规则的索引必须一致。</li>
//
// </ul></li></ul>
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	LIMITEXCEEDED_SECURITYGROUPPOLICYSET = "LimitExceeded.SecurityGroupPolicySet"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER_WITHGUESS = "UnknownParameter.WithGuess"
//	UNSUPPORTEDOPERATION_CLBPOLICYLIMIT = "UnsupportedOperation.ClbPolicyLimit"
//	UNSUPPORTEDOPERATION_DUPLICATEPOLICY = "UnsupportedOperation.DuplicatePolicy"
//	UNSUPPORTEDOPERATION_VERSIONMISMATCH = "UnsupportedOperation.VersionMismatch"
func (c *Client) CreateSecurityGroupPoliciesWithContext(ctx context.Context, request *CreateSecurityGroupPoliciesRequest) (response *CreateSecurityGroupPoliciesResponse, err error) {
	if request == nil {
		request = NewCreateSecurityGroupPoliciesRequest()
	}
	request.SetContext(ctx)

	response = NewCreateSecurityGroupPoliciesResponse()
	err = c.Send(request, response)
	return
}

func NewCreateSecurityGroupWithPoliciesRequest() (request *CreateSecurityGroupWithPoliciesRequest) {
	request = &CreateSecurityGroupWithPoliciesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateSecurityGroupWithPolicies")

	return
}

func NewCreateSecurityGroupWithPoliciesResponse() (response *CreateSecurityGroupWithPoliciesResponse) {
	response = &CreateSecurityGroupWithPoliciesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateSecurityGroupWithPolicies
// 本接口（CreateSecurityGroupWithPolicies）用于创建新的安全组（SecurityGroup），并且可以同时添加安全组规则（SecurityGroupPolicy）。
//
// * 每个账户下每个地域的每个项目的<a href="https://cloud.tencent.com/document/product/213/12453">安全组数量限制</a>。
//
// * 新建的安全组的入站和出站规则默认都是全部拒绝，在创建后通常您需要再调用CreateSecurityGroupPolicies将安全组的规则设置为需要的规则。
//
// 安全组规则说明：
//
// * Version安全组规则版本号，用户每次更新安全规则版本会自动加1，防止您更新的路由规则已过期，不填不考虑冲突。
//
// * Protocol字段支持输入TCP, UDP, ICMP, ICMPV6, GRE, ALL。
//
// * CidrBlock字段允许输入符合cidr格式标准的任意字符串。(展开)在基础网络中，如果CidrBlock包含您的账户内的云服务器之外的设备在腾讯云的内网IP，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。
//
// * Ipv6CidrBlock字段允许输入符合IPv6 cidr格式标准的任意字符串。(展开)在基础网络中，如果Ipv6CidrBlock包含您的账户内的云服务器之外的设备在腾讯云的内网IPv6，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。
//
// * SecurityGroupId字段允许输入与待修改的安全组位于相同项目中的安全组ID，包括这个安全组ID本身，代表安全组下所有云服务器的内网IP。使用这个字段时，这条规则用来匹配网络报文的过程中会随着被使用的这个ID所关联的云服务器变化而变化，不需要重新修改。
//
// * Port字段允许输入一个单独端口号，或者用减号分隔的两个端口号代表端口范围，例如80或8000-8010。只有当Protocol字段是TCP或UDP时，Port字段才被接受，即Protocol字段不是TCP或UDP时，Protocol和Port排他关系，不允许同时输入，否则会接口报错。
//
// * Action字段只允许输入ACCEPT或DROP。
//
// * CidrBlock, Ipv6CidrBlock, SecurityGroupId, AddressTemplate四者是排他关系，不允许同时输入，Protocol + Port和ServiceTemplate二者是排他关系，不允许同时输入。
//
// * 一次请求中只能创建单个方向的规则, 如果需要指定索引（PolicyIndex）参数, 多条规则的索引必须一致。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateSecurityGroupWithPolicies(request *CreateSecurityGroupWithPoliciesRequest) (response *CreateSecurityGroupWithPoliciesResponse, err error) {
	if request == nil {
		request = NewCreateSecurityGroupWithPoliciesRequest()
	}

	response = NewCreateSecurityGroupWithPoliciesResponse()
	err = c.Send(request, response)
	return
}

// CreateSecurityGroupWithPolicies
// 本接口（CreateSecurityGroupWithPolicies）用于创建新的安全组（SecurityGroup），并且可以同时添加安全组规则（SecurityGroupPolicy）。
//
// * 每个账户下每个地域的每个项目的<a href="https://cloud.tencent.com/document/product/213/12453">安全组数量限制</a>。
//
// * 新建的安全组的入站和出站规则默认都是全部拒绝，在创建后通常您需要再调用CreateSecurityGroupPolicies将安全组的规则设置为需要的规则。
//
// 安全组规则说明：
//
// * Version安全组规则版本号，用户每次更新安全规则版本会自动加1，防止您更新的路由规则已过期，不填不考虑冲突。
//
// * Protocol字段支持输入TCP, UDP, ICMP, ICMPV6, GRE, ALL。
//
// * CidrBlock字段允许输入符合cidr格式标准的任意字符串。(展开)在基础网络中，如果CidrBlock包含您的账户内的云服务器之外的设备在腾讯云的内网IP，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。
//
// * Ipv6CidrBlock字段允许输入符合IPv6 cidr格式标准的任意字符串。(展开)在基础网络中，如果Ipv6CidrBlock包含您的账户内的云服务器之外的设备在腾讯云的内网IPv6，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。
//
// * SecurityGroupId字段允许输入与待修改的安全组位于相同项目中的安全组ID，包括这个安全组ID本身，代表安全组下所有云服务器的内网IP。使用这个字段时，这条规则用来匹配网络报文的过程中会随着被使用的这个ID所关联的云服务器变化而变化，不需要重新修改。
//
// * Port字段允许输入一个单独端口号，或者用减号分隔的两个端口号代表端口范围，例如80或8000-8010。只有当Protocol字段是TCP或UDP时，Port字段才被接受，即Protocol字段不是TCP或UDP时，Protocol和Port排他关系，不允许同时输入，否则会接口报错。
//
// * Action字段只允许输入ACCEPT或DROP。
//
// * CidrBlock, Ipv6CidrBlock, SecurityGroupId, AddressTemplate四者是排他关系，不允许同时输入，Protocol + Port和ServiceTemplate二者是排他关系，不允许同时输入。
//
// * 一次请求中只能创建单个方向的规则, 如果需要指定索引（PolicyIndex）参数, 多条规则的索引必须一致。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateSecurityGroupWithPoliciesWithContext(ctx context.Context, request *CreateSecurityGroupWithPoliciesRequest) (response *CreateSecurityGroupWithPoliciesResponse, err error) {
	if request == nil {
		request = NewCreateSecurityGroupWithPoliciesRequest()
	}
	request.SetContext(ctx)

	response = NewCreateSecurityGroupWithPoliciesResponse()
	err = c.Send(request, response)
	return
}

func NewCreateServiceTemplateRequest() (request *CreateServiceTemplateRequest) {
	request = &CreateServiceTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateServiceTemplate")

	return
}

func NewCreateServiceTemplateResponse() (response *CreateServiceTemplateResponse) {
	response = &CreateServiceTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateServiceTemplate
// 本接口（CreateServiceTemplate）用于创建协议端口模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) CreateServiceTemplate(request *CreateServiceTemplateRequest) (response *CreateServiceTemplateResponse, err error) {
	if request == nil {
		request = NewCreateServiceTemplateRequest()
	}

	response = NewCreateServiceTemplateResponse()
	err = c.Send(request, response)
	return
}

// CreateServiceTemplate
// 本接口（CreateServiceTemplate）用于创建协议端口模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) CreateServiceTemplateWithContext(ctx context.Context, request *CreateServiceTemplateRequest) (response *CreateServiceTemplateResponse, err error) {
	if request == nil {
		request = NewCreateServiceTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewCreateServiceTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewCreateServiceTemplateGroupRequest() (request *CreateServiceTemplateGroupRequest) {
	request = &CreateServiceTemplateGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateServiceTemplateGroup")

	return
}

func NewCreateServiceTemplateGroupResponse() (response *CreateServiceTemplateGroupResponse) {
	response = &CreateServiceTemplateGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateServiceTemplateGroup
// 本接口（CreateServiceTemplateGroup）用于创建协议端口模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) CreateServiceTemplateGroup(request *CreateServiceTemplateGroupRequest) (response *CreateServiceTemplateGroupResponse, err error) {
	if request == nil {
		request = NewCreateServiceTemplateGroupRequest()
	}

	response = NewCreateServiceTemplateGroupResponse()
	err = c.Send(request, response)
	return
}

// CreateServiceTemplateGroup
// 本接口（CreateServiceTemplateGroup）用于创建协议端口模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) CreateServiceTemplateGroupWithContext(ctx context.Context, request *CreateServiceTemplateGroupRequest) (response *CreateServiceTemplateGroupResponse, err error) {
	if request == nil {
		request = NewCreateServiceTemplateGroupRequest()
	}
	request.SetContext(ctx)

	response = NewCreateServiceTemplateGroupResponse()
	err = c.Send(request, response)
	return
}

func NewCreateSubnetRequest() (request *CreateSubnetRequest) {
	request = &CreateSubnetRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateSubnet")

	return
}

func NewCreateSubnetResponse() (response *CreateSubnetResponse) {
	response = &CreateSubnetResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateSubnet
// 本接口(CreateSubnet)用于创建子网。
//
// * 创建子网前必须创建好 VPC。
//
// * 子网创建成功后，子网网段不能修改。子网网段必须在VPC网段内，可以和VPC网段相同（VPC有且只有一个子网时），建议子网网段在VPC网段内，预留网段给其他子网使用。
//
// * 您可以创建的最小网段子网掩码为28（有16个IP地址），最大网段子网掩码为16（65,536个IP地址）。
//
// * 同一个VPC内，多个子网的网段不能重叠。
//
// * 子网创建后会自动关联到默认路由表。
//
// * 创建子网同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETRANGE = "InvalidParameterValue.SubnetRange"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_ZONECONFLICT = "InvalidParameterValue.ZoneConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_APPIDMISMATCH = "UnsupportedOperation.AppIdMismatch"
//	UNSUPPORTEDOPERATION_DCGATEWAYSNOTFOUNDINVPC = "UnsupportedOperation.DcGatewaysNotFoundInVpc"
func (c *Client) CreateSubnet(request *CreateSubnetRequest) (response *CreateSubnetResponse, err error) {
	if request == nil {
		request = NewCreateSubnetRequest()
	}

	response = NewCreateSubnetResponse()
	err = c.Send(request, response)
	return
}

// CreateSubnet
// 本接口(CreateSubnet)用于创建子网。
//
// * 创建子网前必须创建好 VPC。
//
// * 子网创建成功后，子网网段不能修改。子网网段必须在VPC网段内，可以和VPC网段相同（VPC有且只有一个子网时），建议子网网段在VPC网段内，预留网段给其他子网使用。
//
// * 您可以创建的最小网段子网掩码为28（有16个IP地址），最大网段子网掩码为16（65,536个IP地址）。
//
// * 同一个VPC内，多个子网的网段不能重叠。
//
// * 子网创建后会自动关联到默认路由表。
//
// * 创建子网同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETRANGE = "InvalidParameterValue.SubnetRange"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_ZONECONFLICT = "InvalidParameterValue.ZoneConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_APPIDMISMATCH = "UnsupportedOperation.AppIdMismatch"
//	UNSUPPORTEDOPERATION_DCGATEWAYSNOTFOUNDINVPC = "UnsupportedOperation.DcGatewaysNotFoundInVpc"
func (c *Client) CreateSubnetWithContext(ctx context.Context, request *CreateSubnetRequest) (response *CreateSubnetResponse, err error) {
	if request == nil {
		request = NewCreateSubnetRequest()
	}
	request.SetContext(ctx)

	response = NewCreateSubnetResponse()
	err = c.Send(request, response)
	return
}

func NewCreateSubnetsRequest() (request *CreateSubnetsRequest) {
	request = &CreateSubnetsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateSubnets")

	return
}

func NewCreateSubnetsResponse() (response *CreateSubnetsResponse) {
	response = &CreateSubnetsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateSubnets
// 本接口(CreateSubnets)用于批量创建子网。
//
// * 创建子网前必须创建好 VPC。
//
// * 子网创建成功后，子网网段不能修改。子网网段必须在VPC网段内，可以和VPC网段相同（VPC有且只有一个子网时），建议子网网段在VPC网段内，预留网段给其他子网使用。
//
// * 您可以创建的最小网段子网掩码为28（有16个IP地址），最大网段子网掩码为16（65,536个IP地址）。
//
// * 同一个VPC内，多个子网的网段不能重叠。
//
// * 子网创建后会自动关联到默认路由表。
//
// * 创建子网同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETRANGE = "InvalidParameterValue.SubnetRange"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_ZONECONFLICT = "InvalidParameterValue.ZoneConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_APPIDMISMATCH = "UnsupportedOperation.AppIdMismatch"
//	UNSUPPORTEDOPERATION_DCGATEWAYSNOTFOUNDINVPC = "UnsupportedOperation.DcGatewaysNotFoundInVpc"
func (c *Client) CreateSubnets(request *CreateSubnetsRequest) (response *CreateSubnetsResponse, err error) {
	if request == nil {
		request = NewCreateSubnetsRequest()
	}

	response = NewCreateSubnetsResponse()
	err = c.Send(request, response)
	return
}

// CreateSubnets
// 本接口(CreateSubnets)用于批量创建子网。
//
// * 创建子网前必须创建好 VPC。
//
// * 子网创建成功后，子网网段不能修改。子网网段必须在VPC网段内，可以和VPC网段相同（VPC有且只有一个子网时），建议子网网段在VPC网段内，预留网段给其他子网使用。
//
// * 您可以创建的最小网段子网掩码为28（有16个IP地址），最大网段子网掩码为16（65,536个IP地址）。
//
// * 同一个VPC内，多个子网的网段不能重叠。
//
// * 子网创建后会自动关联到默认路由表。
//
// * 创建子网同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETRANGE = "InvalidParameterValue.SubnetRange"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_ZONECONFLICT = "InvalidParameterValue.ZoneConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_APPIDMISMATCH = "UnsupportedOperation.AppIdMismatch"
//	UNSUPPORTEDOPERATION_DCGATEWAYSNOTFOUNDINVPC = "UnsupportedOperation.DcGatewaysNotFoundInVpc"
func (c *Client) CreateSubnetsWithContext(ctx context.Context, request *CreateSubnetsRequest) (response *CreateSubnetsResponse, err error) {
	if request == nil {
		request = NewCreateSubnetsRequest()
	}
	request.SetContext(ctx)

	response = NewCreateSubnetsResponse()
	err = c.Send(request, response)
	return
}

func NewCreateVpcRequest() (request *CreateVpcRequest) {
	request = &CreateVpcRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateVpc")

	return
}

func NewCreateVpcResponse() (response *CreateVpcResponse) {
	response = &CreateVpcResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateVpc
// 本接口(CreateVpc)用于创建私有网络(VPC)。
//
// * 用户可以创建的最小网段子网掩码为28（有16个IP地址），最大网段子网掩码为16（65,536个IP地址），如果需要规划VPC网段请参见[网络规划](https://cloud.tencent.com/document/product/215/30313)。
//
// * 同一个地域能创建的VPC资源个数也是有限制的，详见 <a href="https://cloud.tencent.com/doc/product/215/537" title="VPC使用限制">VPC使用限制</a>，如果需要申请更多资源，请提交[工单申请](https://console.cloud.tencent.com/workorder/category)。
//
// * 创建VPC同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateVpc(request *CreateVpcRequest) (response *CreateVpcResponse, err error) {
	if request == nil {
		request = NewCreateVpcRequest()
	}

	response = NewCreateVpcResponse()
	err = c.Send(request, response)
	return
}

// CreateVpc
// 本接口(CreateVpc)用于创建私有网络(VPC)。
//
// * 用户可以创建的最小网段子网掩码为28（有16个IP地址），最大网段子网掩码为16（65,536个IP地址），如果需要规划VPC网段请参见[网络规划](https://cloud.tencent.com/document/product/215/30313)。
//
// * 同一个地域能创建的VPC资源个数也是有限制的，详见 <a href="https://cloud.tencent.com/doc/product/215/537" title="VPC使用限制">VPC使用限制</a>，如果需要申请更多资源，请提交[工单申请](https://console.cloud.tencent.com/workorder/category)。
//
// * 创建VPC同时可以绑定标签, 应答里的标签列表代表添加成功的标签。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreateVpcWithContext(ctx context.Context, request *CreateVpcRequest) (response *CreateVpcResponse, err error) {
	if request == nil {
		request = NewCreateVpcRequest()
	}
	request.SetContext(ctx)

	response = NewCreateVpcResponse()
	err = c.Send(request, response)
	return
}

func NewCreateVpcEndPointRequest() (request *CreateVpcEndPointRequest) {
	request = &CreateVpcEndPointRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateVpcEndPoint")

	return
}

func NewCreateVpcEndPointResponse() (response *CreateVpcEndPointResponse) {
	response = &CreateVpcEndPointResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateVpcEndPoint
// 创建终端节点。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_RESERVED = "InvalidParameterValue.Reserved"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	RESOURCEUNAVAILABLE_SERVICEWHITELISTNOTADDED = "ResourceUnavailable.ServiceWhiteListNotAdded"
//	UNSUPPORTEDOPERATION_ENDPOINTSERVICE = "UnsupportedOperation.EndPointService"
//	UNSUPPORTEDOPERATION_SPECIALENDPOINTSERVICE = "UnsupportedOperation.SpecialEndPointService"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) CreateVpcEndPoint(request *CreateVpcEndPointRequest) (response *CreateVpcEndPointResponse, err error) {
	if request == nil {
		request = NewCreateVpcEndPointRequest()
	}

	response = NewCreateVpcEndPointResponse()
	err = c.Send(request, response)
	return
}

// CreateVpcEndPoint
// 创建终端节点。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_RESERVED = "InvalidParameterValue.Reserved"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	RESOURCEUNAVAILABLE_SERVICEWHITELISTNOTADDED = "ResourceUnavailable.ServiceWhiteListNotAdded"
//	UNSUPPORTEDOPERATION_ENDPOINTSERVICE = "UnsupportedOperation.EndPointService"
//	UNSUPPORTEDOPERATION_SPECIALENDPOINTSERVICE = "UnsupportedOperation.SpecialEndPointService"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) CreateVpcEndPointWithContext(ctx context.Context, request *CreateVpcEndPointRequest) (response *CreateVpcEndPointResponse, err error) {
	if request == nil {
		request = NewCreateVpcEndPointRequest()
	}
	request.SetContext(ctx)

	response = NewCreateVpcEndPointResponse()
	err = c.Send(request, response)
	return
}

func NewCreateVpcEndPointServiceRequest() (request *CreateVpcEndPointServiceRequest) {
	request = &CreateVpcEndPointServiceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateVpcEndPointService")

	return
}

func NewCreateVpcEndPointServiceResponse() (response *CreateVpcEndPointServiceResponse) {
	response = &CreateVpcEndPointServiceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateVpcEndPointService
// 创建终端节点服务。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) CreateVpcEndPointService(request *CreateVpcEndPointServiceRequest) (response *CreateVpcEndPointServiceResponse, err error) {
	if request == nil {
		request = NewCreateVpcEndPointServiceRequest()
	}

	response = NewCreateVpcEndPointServiceResponse()
	err = c.Send(request, response)
	return
}

// CreateVpcEndPointService
// 创建终端节点服务。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) CreateVpcEndPointServiceWithContext(ctx context.Context, request *CreateVpcEndPointServiceRequest) (response *CreateVpcEndPointServiceResponse, err error) {
	if request == nil {
		request = NewCreateVpcEndPointServiceRequest()
	}
	request.SetContext(ctx)

	response = NewCreateVpcEndPointServiceResponse()
	err = c.Send(request, response)
	return
}

func NewCreateVpcEndPointServiceWhiteListRequest() (request *CreateVpcEndPointServiceWhiteListRequest) {
	request = &CreateVpcEndPointServiceWhiteListRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateVpcEndPointServiceWhiteList")

	return
}

func NewCreateVpcEndPointServiceWhiteListResponse() (response *CreateVpcEndPointServiceWhiteListResponse) {
	response = &CreateVpcEndPointServiceWhiteListResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateVpcEndPointServiceWhiteList
// 创建终端服务白名单。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
func (c *Client) CreateVpcEndPointServiceWhiteList(request *CreateVpcEndPointServiceWhiteListRequest) (response *CreateVpcEndPointServiceWhiteListResponse, err error) {
	if request == nil {
		request = NewCreateVpcEndPointServiceWhiteListRequest()
	}

	response = NewCreateVpcEndPointServiceWhiteListResponse()
	err = c.Send(request, response)
	return
}

// CreateVpcEndPointServiceWhiteList
// 创建终端服务白名单。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
func (c *Client) CreateVpcEndPointServiceWhiteListWithContext(ctx context.Context, request *CreateVpcEndPointServiceWhiteListRequest) (response *CreateVpcEndPointServiceWhiteListResponse, err error) {
	if request == nil {
		request = NewCreateVpcEndPointServiceWhiteListRequest()
	}
	request.SetContext(ctx)

	response = NewCreateVpcEndPointServiceWhiteListResponse()
	err = c.Send(request, response)
	return
}

func NewCreateVpnConnectionRequest() (request *CreateVpnConnectionRequest) {
	request = &CreateVpnConnectionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateVpnConnection")

	return
}

func NewCreateVpnConnectionResponse() (response *CreateVpnConnectionResponse) {
	response = &CreateVpnConnectionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateVpnConnection
// 本接口（CreateVpnConnection）用于创建VPN通道。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	INVALIDPARAMETERVALUE_VPNCONNCIDRCONFLICT = "InvalidParameterValue.VpnConnCidrConflict"
//	INVALIDPARAMETERVALUE_VPNCONNHEALTHCHECKIPCONFLICT = "InvalidParameterValue.VpnConnHealthCheckIpConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) CreateVpnConnection(request *CreateVpnConnectionRequest) (response *CreateVpnConnectionResponse, err error) {
	if request == nil {
		request = NewCreateVpnConnectionRequest()
	}

	response = NewCreateVpnConnectionResponse()
	err = c.Send(request, response)
	return
}

// CreateVpnConnection
// 本接口（CreateVpnConnection）用于创建VPN通道。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	INVALIDPARAMETERVALUE_VPNCONNCIDRCONFLICT = "InvalidParameterValue.VpnConnCidrConflict"
//	INVALIDPARAMETERVALUE_VPNCONNHEALTHCHECKIPCONFLICT = "InvalidParameterValue.VpnConnHealthCheckIpConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) CreateVpnConnectionWithContext(ctx context.Context, request *CreateVpnConnectionRequest) (response *CreateVpnConnectionResponse, err error) {
	if request == nil {
		request = NewCreateVpnConnectionRequest()
	}
	request.SetContext(ctx)

	response = NewCreateVpnConnectionResponse()
	err = c.Send(request, response)
	return
}

func NewCreateVpnGatewayRequest() (request *CreateVpnGatewayRequest) {
	request = &CreateVpnGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateVpnGateway")

	return
}

func NewCreateVpnGatewayResponse() (response *CreateVpnGatewayResponse) {
	response = &CreateVpnGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateVpnGateway
// 本接口（CreateVpnGateway）用于创建VPN网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_VPNCONNCIDRCONFLICT = "InvalidParameterValue.VpnConnCidrConflict"
//	INVALIDVPCID_MALFORMED = "InvalidVpcId.Malformed"
//	INVALIDVPCID_NOTFOUND = "InvalidVpcId.NotFound"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_NOREALNAMEAUTHENTICATION = "UnauthorizedOperation.NoRealNameAuthentication"
//	UNSUPPORTEDOPERATION_INSUFFICIENTFUNDS = "UnsupportedOperation.InsufficientFunds"
func (c *Client) CreateVpnGateway(request *CreateVpnGatewayRequest) (response *CreateVpnGatewayResponse, err error) {
	if request == nil {
		request = NewCreateVpnGatewayRequest()
	}

	response = NewCreateVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

// CreateVpnGateway
// 本接口（CreateVpnGateway）用于创建VPN网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPARAMETERVALUE_VPNCONNCIDRCONFLICT = "InvalidParameterValue.VpnConnCidrConflict"
//	INVALIDVPCID_MALFORMED = "InvalidVpcId.Malformed"
//	INVALIDVPCID_NOTFOUND = "InvalidVpcId.NotFound"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_NOREALNAMEAUTHENTICATION = "UnauthorizedOperation.NoRealNameAuthentication"
//	UNSUPPORTEDOPERATION_INSUFFICIENTFUNDS = "UnsupportedOperation.InsufficientFunds"
func (c *Client) CreateVpnGatewayWithContext(ctx context.Context, request *CreateVpnGatewayRequest) (response *CreateVpnGatewayResponse, err error) {
	if request == nil {
		request = NewCreateVpnGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewCreateVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewCreateVpnGatewayRoutesRequest() (request *CreateVpnGatewayRoutesRequest) {
	request = &CreateVpnGatewayRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "CreateVpnGatewayRoutes")

	return
}

func NewCreateVpnGatewayRoutesResponse() (response *CreateVpnGatewayRoutesResponse) {
	response = &CreateVpnGatewayRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateVpnGatewayRoutes
// 创建路由型VPN网关的目的路由
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateVpnGatewayRoutes(request *CreateVpnGatewayRoutesRequest) (response *CreateVpnGatewayRoutesResponse, err error) {
	if request == nil {
		request = NewCreateVpnGatewayRoutesRequest()
	}

	response = NewCreateVpnGatewayRoutesResponse()
	err = c.Send(request, response)
	return
}

// CreateVpnGatewayRoutes
// 创建路由型VPN网关的目的路由
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateVpnGatewayRoutesWithContext(ctx context.Context, request *CreateVpnGatewayRoutesRequest) (response *CreateVpnGatewayRoutesResponse, err error) {
	if request == nil {
		request = NewCreateVpnGatewayRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewCreateVpnGatewayRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteAddressTemplateRequest() (request *DeleteAddressTemplateRequest) {
	request = &DeleteAddressTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteAddressTemplate")

	return
}

func NewDeleteAddressTemplateResponse() (response *DeleteAddressTemplateResponse) {
	response = &DeleteAddressTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteAddressTemplate
// 本接口（DeleteAddressTemplate）用于删除IP地址模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteAddressTemplate(request *DeleteAddressTemplateRequest) (response *DeleteAddressTemplateResponse, err error) {
	if request == nil {
		request = NewDeleteAddressTemplateRequest()
	}

	response = NewDeleteAddressTemplateResponse()
	err = c.Send(request, response)
	return
}

// DeleteAddressTemplate
// 本接口（DeleteAddressTemplate）用于删除IP地址模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteAddressTemplateWithContext(ctx context.Context, request *DeleteAddressTemplateRequest) (response *DeleteAddressTemplateResponse, err error) {
	if request == nil {
		request = NewDeleteAddressTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteAddressTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteAddressTemplateGroupRequest() (request *DeleteAddressTemplateGroupRequest) {
	request = &DeleteAddressTemplateGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteAddressTemplateGroup")

	return
}

func NewDeleteAddressTemplateGroupResponse() (response *DeleteAddressTemplateGroupResponse) {
	response = &DeleteAddressTemplateGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteAddressTemplateGroup
// 本接口（DeleteAddressTemplateGroup）用于删除IP地址模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteAddressTemplateGroup(request *DeleteAddressTemplateGroupRequest) (response *DeleteAddressTemplateGroupResponse, err error) {
	if request == nil {
		request = NewDeleteAddressTemplateGroupRequest()
	}

	response = NewDeleteAddressTemplateGroupResponse()
	err = c.Send(request, response)
	return
}

// DeleteAddressTemplateGroup
// 本接口（DeleteAddressTemplateGroup）用于删除IP地址模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteAddressTemplateGroupWithContext(ctx context.Context, request *DeleteAddressTemplateGroupRequest) (response *DeleteAddressTemplateGroupResponse, err error) {
	if request == nil {
		request = NewDeleteAddressTemplateGroupRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteAddressTemplateGroupResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteAssistantCidrRequest() (request *DeleteAssistantCidrRequest) {
	request = &DeleteAssistantCidrRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteAssistantCidr")

	return
}

func NewDeleteAssistantCidrResponse() (response *DeleteAssistantCidrResponse) {
	response = &DeleteAssistantCidrResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteAssistantCidr
// 本接口(DeleteAssistantCidr)用于删除辅助CIDR。（接口灰度中，如需使用请提工单。）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteAssistantCidr(request *DeleteAssistantCidrRequest) (response *DeleteAssistantCidrResponse, err error) {
	if request == nil {
		request = NewDeleteAssistantCidrRequest()
	}

	response = NewDeleteAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

// DeleteAssistantCidr
// 本接口(DeleteAssistantCidr)用于删除辅助CIDR。（接口灰度中，如需使用请提工单。）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteAssistantCidrWithContext(ctx context.Context, request *DeleteAssistantCidrRequest) (response *DeleteAssistantCidrResponse, err error) {
	if request == nil {
		request = NewDeleteAssistantCidrRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteBandwidthPackageRequest() (request *DeleteBandwidthPackageRequest) {
	request = &DeleteBandwidthPackageRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteBandwidthPackage")

	return
}

func NewDeleteBandwidthPackageResponse() (response *DeleteBandwidthPackageResponse) {
	response = &DeleteBandwidthPackageResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteBandwidthPackage
// 接口支持删除共享带宽包，包括[设备带宽包](https://cloud.tencent.com/document/product/684/15246#.E8.AE.BE.E5.A4.87.E5.B8.A6.E5.AE.BD.E5.8C.85)和[IP带宽包](https://cloud.tencent.com/document/product/684/15246#ip-.E5.B8.A6.E5.AE.BD.E5.8C.85)
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEINUSE = "InvalidParameterValue.BandwidthPackageInUse"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDADDRESSSTATE = "UnsupportedOperation.InvalidAddressState"
func (c *Client) DeleteBandwidthPackage(request *DeleteBandwidthPackageRequest) (response *DeleteBandwidthPackageResponse, err error) {
	if request == nil {
		request = NewDeleteBandwidthPackageRequest()
	}

	response = NewDeleteBandwidthPackageResponse()
	err = c.Send(request, response)
	return
}

// DeleteBandwidthPackage
// 接口支持删除共享带宽包，包括[设备带宽包](https://cloud.tencent.com/document/product/684/15246#.E8.AE.BE.E5.A4.87.E5.B8.A6.E5.AE.BD.E5.8C.85)和[IP带宽包](https://cloud.tencent.com/document/product/684/15246#ip-.E5.B8.A6.E5.AE.BD.E5.8C.85)
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEINUSE = "InvalidParameterValue.BandwidthPackageInUse"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDADDRESSSTATE = "UnsupportedOperation.InvalidAddressState"
func (c *Client) DeleteBandwidthPackageWithContext(ctx context.Context, request *DeleteBandwidthPackageRequest) (response *DeleteBandwidthPackageResponse, err error) {
	if request == nil {
		request = NewDeleteBandwidthPackageRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteBandwidthPackageResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteCcnRequest() (request *DeleteCcnRequest) {
	request = &DeleteCcnRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteCcn")

	return
}

func NewDeleteCcnResponse() (response *DeleteCcnResponse) {
	response = &DeleteCcnResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteCcn
// 本接口（DeleteCcn）用于删除云联网。
//
// * 删除后，云联网关联的所有实例间路由将被删除，网络将会中断，请务必确认
//
// * 删除云联网是不可逆的操作，请谨慎处理。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_BANDWIDTHNOTEXPIRED = "UnsupportedOperation.BandwidthNotExpired"
//	UNSUPPORTEDOPERATION_CCNHASFLOWLOG = "UnsupportedOperation.CcnHasFlowLog"
func (c *Client) DeleteCcn(request *DeleteCcnRequest) (response *DeleteCcnResponse, err error) {
	if request == nil {
		request = NewDeleteCcnRequest()
	}

	response = NewDeleteCcnResponse()
	err = c.Send(request, response)
	return
}

// DeleteCcn
// 本接口（DeleteCcn）用于删除云联网。
//
// * 删除后，云联网关联的所有实例间路由将被删除，网络将会中断，请务必确认
//
// * 删除云联网是不可逆的操作，请谨慎处理。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_BANDWIDTHNOTEXPIRED = "UnsupportedOperation.BandwidthNotExpired"
//	UNSUPPORTEDOPERATION_CCNHASFLOWLOG = "UnsupportedOperation.CcnHasFlowLog"
func (c *Client) DeleteCcnWithContext(ctx context.Context, request *DeleteCcnRequest) (response *DeleteCcnResponse, err error) {
	if request == nil {
		request = NewDeleteCcnRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteCcnResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteCustomerGatewayRequest() (request *DeleteCustomerGatewayRequest) {
	request = &DeleteCustomerGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteCustomerGateway")

	return
}

func NewDeleteCustomerGatewayResponse() (response *DeleteCustomerGatewayResponse) {
	response = &DeleteCustomerGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteCustomerGateway
// 本接口（DeleteCustomerGateway）用于删除对端网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteCustomerGateway(request *DeleteCustomerGatewayRequest) (response *DeleteCustomerGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteCustomerGatewayRequest()
	}

	response = NewDeleteCustomerGatewayResponse()
	err = c.Send(request, response)
	return
}

// DeleteCustomerGateway
// 本接口（DeleteCustomerGateway）用于删除对端网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteCustomerGatewayWithContext(ctx context.Context, request *DeleteCustomerGatewayRequest) (response *DeleteCustomerGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteCustomerGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteCustomerGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteDhcpIpRequest() (request *DeleteDhcpIpRequest) {
	request = &DeleteDhcpIpRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteDhcpIp")

	return
}

func NewDeleteDhcpIpResponse() (response *DeleteDhcpIpResponse) {
	response = &DeleteDhcpIpResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteDhcpIp
// 本接口（DeleteDhcpIp）用于删除DhcpIp。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteDhcpIp(request *DeleteDhcpIpRequest) (response *DeleteDhcpIpResponse, err error) {
	if request == nil {
		request = NewDeleteDhcpIpRequest()
	}

	response = NewDeleteDhcpIpResponse()
	err = c.Send(request, response)
	return
}

// DeleteDhcpIp
// 本接口（DeleteDhcpIp）用于删除DhcpIp。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteDhcpIpWithContext(ctx context.Context, request *DeleteDhcpIpRequest) (response *DeleteDhcpIpResponse, err error) {
	if request == nil {
		request = NewDeleteDhcpIpRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteDhcpIpResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteDirectConnectGatewayRequest() (request *DeleteDirectConnectGatewayRequest) {
	request = &DeleteDirectConnectGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteDirectConnectGateway")

	return
}

func NewDeleteDirectConnectGatewayResponse() (response *DeleteDirectConnectGatewayResponse) {
	response = &DeleteDirectConnectGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteDirectConnectGateway
// 本接口（DeleteDirectConnectGateway）用于删除专线网关。
//
// <li>如果是 NAT 网关，删除专线网关后，NAT 规则以及 ACL 策略都被清理了。</li>
//
// <li>删除专线网关后，系统会删除路由表中跟该专线网关相关的路由策略。</li>
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`QueryTask`接口
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteDirectConnectGateway(request *DeleteDirectConnectGatewayRequest) (response *DeleteDirectConnectGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteDirectConnectGatewayRequest()
	}

	response = NewDeleteDirectConnectGatewayResponse()
	err = c.Send(request, response)
	return
}

// DeleteDirectConnectGateway
// 本接口（DeleteDirectConnectGateway）用于删除专线网关。
//
// <li>如果是 NAT 网关，删除专线网关后，NAT 规则以及 ACL 策略都被清理了。</li>
//
// <li>删除专线网关后，系统会删除路由表中跟该专线网关相关的路由策略。</li>
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`QueryTask`接口
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteDirectConnectGatewayWithContext(ctx context.Context, request *DeleteDirectConnectGatewayRequest) (response *DeleteDirectConnectGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteDirectConnectGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteDirectConnectGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteDirectConnectGatewayCcnRoutesRequest() (request *DeleteDirectConnectGatewayCcnRoutesRequest) {
	request = &DeleteDirectConnectGatewayCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteDirectConnectGatewayCcnRoutes")

	return
}

func NewDeleteDirectConnectGatewayCcnRoutesResponse() (response *DeleteDirectConnectGatewayCcnRoutesResponse) {
	response = &DeleteDirectConnectGatewayCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteDirectConnectGatewayCcnRoutes
// 本接口（DeleteDirectConnectGatewayCcnRoutes）用于删除专线网关的云联网路由（IDC网段）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteDirectConnectGatewayCcnRoutes(request *DeleteDirectConnectGatewayCcnRoutesRequest) (response *DeleteDirectConnectGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDeleteDirectConnectGatewayCcnRoutesRequest()
	}

	response = NewDeleteDirectConnectGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// DeleteDirectConnectGatewayCcnRoutes
// 本接口（DeleteDirectConnectGatewayCcnRoutes）用于删除专线网关的云联网路由（IDC网段）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteDirectConnectGatewayCcnRoutesWithContext(ctx context.Context, request *DeleteDirectConnectGatewayCcnRoutesRequest) (response *DeleteDirectConnectGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDeleteDirectConnectGatewayCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteDirectConnectGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteFlowLogRequest() (request *DeleteFlowLogRequest) {
	request = &DeleteFlowLogRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteFlowLog")

	return
}

func NewDeleteFlowLogResponse() (response *DeleteFlowLogResponse) {
	response = &DeleteFlowLogResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteFlowLog
// 本接口（DeleteFlowLog）用于删除流日志
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteFlowLog(request *DeleteFlowLogRequest) (response *DeleteFlowLogResponse, err error) {
	if request == nil {
		request = NewDeleteFlowLogRequest()
	}

	response = NewDeleteFlowLogResponse()
	err = c.Send(request, response)
	return
}

// DeleteFlowLog
// 本接口（DeleteFlowLog）用于删除流日志
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteFlowLogWithContext(ctx context.Context, request *DeleteFlowLogRequest) (response *DeleteFlowLogResponse, err error) {
	if request == nil {
		request = NewDeleteFlowLogRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteFlowLogResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteHaVipRequest() (request *DeleteHaVipRequest) {
	request = &DeleteHaVipRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteHaVip")

	return
}

func NewDeleteHaVipResponse() (response *DeleteHaVipResponse) {
	response = &DeleteHaVipResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteHaVip
// 本接口（DeleteHaVip）用于删除高可用虚拟IP（HAVIP）。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteHaVip(request *DeleteHaVipRequest) (response *DeleteHaVipResponse, err error) {
	if request == nil {
		request = NewDeleteHaVipRequest()
	}

	response = NewDeleteHaVipResponse()
	err = c.Send(request, response)
	return
}

// DeleteHaVip
// 本接口（DeleteHaVip）用于删除高可用虚拟IP（HAVIP）。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteHaVipWithContext(ctx context.Context, request *DeleteHaVipRequest) (response *DeleteHaVipResponse, err error) {
	if request == nil {
		request = NewDeleteHaVipRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteHaVipResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteIp6TranslatorsRequest() (request *DeleteIp6TranslatorsRequest) {
	request = &DeleteIp6TranslatorsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteIp6Translators")

	return
}

func NewDeleteIp6TranslatorsResponse() (response *DeleteIp6TranslatorsResponse) {
	response = &DeleteIp6TranslatorsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteIp6Translators
// 1. 该接口用于释放IPV6转换实例，支持批量。
//
// 2.  如果IPV6转换实例建立有转换规则，会一并删除。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DeleteIp6Translators(request *DeleteIp6TranslatorsRequest) (response *DeleteIp6TranslatorsResponse, err error) {
	if request == nil {
		request = NewDeleteIp6TranslatorsRequest()
	}

	response = NewDeleteIp6TranslatorsResponse()
	err = c.Send(request, response)
	return
}

// DeleteIp6Translators
// 1. 该接口用于释放IPV6转换实例，支持批量。
//
// 2.  如果IPV6转换实例建立有转换规则，会一并删除。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DeleteIp6TranslatorsWithContext(ctx context.Context, request *DeleteIp6TranslatorsRequest) (response *DeleteIp6TranslatorsResponse, err error) {
	if request == nil {
		request = NewDeleteIp6TranslatorsRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteIp6TranslatorsResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteLocalGatewayRequest() (request *DeleteLocalGatewayRequest) {
	request = &DeleteLocalGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteLocalGateway")

	return
}

func NewDeleteLocalGatewayResponse() (response *DeleteLocalGatewayResponse) {
	response = &DeleteLocalGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteLocalGateway
// 该接口用于删除CDC的本地网关。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteLocalGateway(request *DeleteLocalGatewayRequest) (response *DeleteLocalGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteLocalGatewayRequest()
	}

	response = NewDeleteLocalGatewayResponse()
	err = c.Send(request, response)
	return
}

// DeleteLocalGateway
// 该接口用于删除CDC的本地网关。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteLocalGatewayWithContext(ctx context.Context, request *DeleteLocalGatewayRequest) (response *DeleteLocalGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteLocalGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteLocalGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteNatGatewayRequest() (request *DeleteNatGatewayRequest) {
	request = &DeleteNatGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteNatGateway")

	return
}

func NewDeleteNatGatewayResponse() (response *DeleteNatGatewayResponse) {
	response = &DeleteNatGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteNatGateway
// 本接口（DeleteNatGateway）用于删除NAT网关。
//
// 删除 NAT 网关后，系统会自动删除路由表中包含此 NAT 网关的路由项，同时也会解绑弹性公网IP（EIP）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteNatGateway(request *DeleteNatGatewayRequest) (response *DeleteNatGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteNatGatewayRequest()
	}

	response = NewDeleteNatGatewayResponse()
	err = c.Send(request, response)
	return
}

// DeleteNatGateway
// 本接口（DeleteNatGateway）用于删除NAT网关。
//
// 删除 NAT 网关后，系统会自动删除路由表中包含此 NAT 网关的路由项，同时也会解绑弹性公网IP（EIP）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteNatGatewayWithContext(ctx context.Context, request *DeleteNatGatewayRequest) (response *DeleteNatGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteNatGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteNatGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteNatGatewayDestinationIpPortTranslationNatRuleRequest() (request *DeleteNatGatewayDestinationIpPortTranslationNatRuleRequest) {
	request = &DeleteNatGatewayDestinationIpPortTranslationNatRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteNatGatewayDestinationIpPortTranslationNatRule")

	return
}

func NewDeleteNatGatewayDestinationIpPortTranslationNatRuleResponse() (response *DeleteNatGatewayDestinationIpPortTranslationNatRuleResponse) {
	response = &DeleteNatGatewayDestinationIpPortTranslationNatRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteNatGatewayDestinationIpPortTranslationNatRule
// 本接口（DeleteNatGatewayDestinationIpPortTranslationNatRule）用于删除NAT网关端口转发规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteNatGatewayDestinationIpPortTranslationNatRule(request *DeleteNatGatewayDestinationIpPortTranslationNatRuleRequest) (response *DeleteNatGatewayDestinationIpPortTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewDeleteNatGatewayDestinationIpPortTranslationNatRuleRequest()
	}

	response = NewDeleteNatGatewayDestinationIpPortTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

// DeleteNatGatewayDestinationIpPortTranslationNatRule
// 本接口（DeleteNatGatewayDestinationIpPortTranslationNatRule）用于删除NAT网关端口转发规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteNatGatewayDestinationIpPortTranslationNatRuleWithContext(ctx context.Context, request *DeleteNatGatewayDestinationIpPortTranslationNatRuleRequest) (response *DeleteNatGatewayDestinationIpPortTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewDeleteNatGatewayDestinationIpPortTranslationNatRuleRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteNatGatewayDestinationIpPortTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteNatGatewaySourceIpTranslationNatRuleRequest() (request *DeleteNatGatewaySourceIpTranslationNatRuleRequest) {
	request = &DeleteNatGatewaySourceIpTranslationNatRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteNatGatewaySourceIpTranslationNatRule")

	return
}

func NewDeleteNatGatewaySourceIpTranslationNatRuleResponse() (response *DeleteNatGatewaySourceIpTranslationNatRuleResponse) {
	response = &DeleteNatGatewaySourceIpTranslationNatRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteNatGatewaySourceIpTranslationNatRule
// 本接口（DeleteNatGatewaySourceIpTranslationNatRule）用于删除NAT网关端口SNAT转发规则。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DeleteNatGatewaySourceIpTranslationNatRule(request *DeleteNatGatewaySourceIpTranslationNatRuleRequest) (response *DeleteNatGatewaySourceIpTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewDeleteNatGatewaySourceIpTranslationNatRuleRequest()
	}

	response = NewDeleteNatGatewaySourceIpTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

// DeleteNatGatewaySourceIpTranslationNatRule
// 本接口（DeleteNatGatewaySourceIpTranslationNatRule）用于删除NAT网关端口SNAT转发规则。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DeleteNatGatewaySourceIpTranslationNatRuleWithContext(ctx context.Context, request *DeleteNatGatewaySourceIpTranslationNatRuleRequest) (response *DeleteNatGatewaySourceIpTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewDeleteNatGatewaySourceIpTranslationNatRuleRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteNatGatewaySourceIpTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteNetDetectRequest() (request *DeleteNetDetectRequest) {
	request = &DeleteNetDetectRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteNetDetect")

	return
}

func NewDeleteNetDetectResponse() (response *DeleteNetDetectResponse) {
	response = &DeleteNetDetectResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteNetDetect
// 本接口(DeleteNetDetect)用于删除网络探测实例。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteNetDetect(request *DeleteNetDetectRequest) (response *DeleteNetDetectResponse, err error) {
	if request == nil {
		request = NewDeleteNetDetectRequest()
	}

	response = NewDeleteNetDetectResponse()
	err = c.Send(request, response)
	return
}

// DeleteNetDetect
// 本接口(DeleteNetDetect)用于删除网络探测实例。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteNetDetectWithContext(ctx context.Context, request *DeleteNetDetectRequest) (response *DeleteNetDetectResponse, err error) {
	if request == nil {
		request = NewDeleteNetDetectRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteNetDetectResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteNetworkAclRequest() (request *DeleteNetworkAclRequest) {
	request = &DeleteNetworkAclRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteNetworkAcl")

	return
}

func NewDeleteNetworkAclResponse() (response *DeleteNetworkAclResponse) {
	response = &DeleteNetworkAclResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteNetworkAcl
// 本接口（DeleteNetworkAcl）用于删除网络ACL。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteNetworkAcl(request *DeleteNetworkAclRequest) (response *DeleteNetworkAclResponse, err error) {
	if request == nil {
		request = NewDeleteNetworkAclRequest()
	}

	response = NewDeleteNetworkAclResponse()
	err = c.Send(request, response)
	return
}

// DeleteNetworkAcl
// 本接口（DeleteNetworkAcl）用于删除网络ACL。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteNetworkAclWithContext(ctx context.Context, request *DeleteNetworkAclRequest) (response *DeleteNetworkAclResponse, err error) {
	if request == nil {
		request = NewDeleteNetworkAclRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteNetworkAclResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteNetworkInterfaceRequest() (request *DeleteNetworkInterfaceRequest) {
	request = &DeleteNetworkInterfaceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteNetworkInterface")

	return
}

func NewDeleteNetworkInterfaceResponse() (response *DeleteNetworkInterfaceResponse) {
	response = &DeleteNetworkInterfaceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteNetworkInterface
// 本接口（DeleteNetworkInterface）用于删除弹性网卡。
//
// * 弹性网卡上绑定了云服务器时，不能被删除。
//
// * 删除指定弹性网卡，弹性网卡必须先和子机解绑才能删除。删除之后弹性网卡上所有内网IP都将被退还。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) DeleteNetworkInterface(request *DeleteNetworkInterfaceRequest) (response *DeleteNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewDeleteNetworkInterfaceRequest()
	}

	response = NewDeleteNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

// DeleteNetworkInterface
// 本接口（DeleteNetworkInterface）用于删除弹性网卡。
//
// * 弹性网卡上绑定了云服务器时，不能被删除。
//
// * 删除指定弹性网卡，弹性网卡必须先和子机解绑才能删除。删除之后弹性网卡上所有内网IP都将被退还。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) DeleteNetworkInterfaceWithContext(ctx context.Context, request *DeleteNetworkInterfaceRequest) (response *DeleteNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewDeleteNetworkInterfaceRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteRouteTableRequest() (request *DeleteRouteTableRequest) {
	request = &DeleteRouteTableRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteRouteTable")

	return
}

func NewDeleteRouteTableResponse() (response *DeleteRouteTableResponse) {
	response = &DeleteRouteTableResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteRouteTable
// 删除路由表
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteRouteTable(request *DeleteRouteTableRequest) (response *DeleteRouteTableResponse, err error) {
	if request == nil {
		request = NewDeleteRouteTableRequest()
	}

	response = NewDeleteRouteTableResponse()
	err = c.Send(request, response)
	return
}

// DeleteRouteTable
// 删除路由表
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteRouteTableWithContext(ctx context.Context, request *DeleteRouteTableRequest) (response *DeleteRouteTableResponse, err error) {
	if request == nil {
		request = NewDeleteRouteTableRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteRouteTableResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteRoutesRequest() (request *DeleteRoutesRequest) {
	request = &DeleteRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteRoutes")

	return
}

func NewDeleteRoutesResponse() (response *DeleteRoutesResponse) {
	response = &DeleteRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteRoutes
// 本接口(DeleteRoutes)用于对某个路由表批量删除路由策略（Route）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER_WITHGUESS = "UnknownParameter.WithGuess"
//	UNSUPPORTEDOPERATION_DISABLEDNOTIFYCCN = "UnsupportedOperation.DisabledNotifyCcn"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) DeleteRoutes(request *DeleteRoutesRequest) (response *DeleteRoutesResponse, err error) {
	if request == nil {
		request = NewDeleteRoutesRequest()
	}

	response = NewDeleteRoutesResponse()
	err = c.Send(request, response)
	return
}

// DeleteRoutes
// 本接口(DeleteRoutes)用于对某个路由表批量删除路由策略（Route）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER_WITHGUESS = "UnknownParameter.WithGuess"
//	UNSUPPORTEDOPERATION_DISABLEDNOTIFYCCN = "UnsupportedOperation.DisabledNotifyCcn"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) DeleteRoutesWithContext(ctx context.Context, request *DeleteRoutesRequest) (response *DeleteRoutesResponse, err error) {
	if request == nil {
		request = NewDeleteRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteSecurityGroupRequest() (request *DeleteSecurityGroupRequest) {
	request = &DeleteSecurityGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteSecurityGroup")

	return
}

func NewDeleteSecurityGroupResponse() (response *DeleteSecurityGroupResponse) {
	response = &DeleteSecurityGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteSecurityGroup
// 本接口（DeleteSecurityGroup）用于删除安全组（SecurityGroup）。
//
// * 只有当前账号下的安全组允许被删除。
//
// * 安全组实例ID如果在其他安全组的规则中被引用，则无法直接删除。这种情况下，需要先进行规则修改，再删除安全组。
//
// * 删除的安全组无法再找回，请谨慎调用。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDSECURITYGROUPID_MALFORMED = "InvalidSecurityGroupID.Malformed"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupID.NotFound"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteSecurityGroup(request *DeleteSecurityGroupRequest) (response *DeleteSecurityGroupResponse, err error) {
	if request == nil {
		request = NewDeleteSecurityGroupRequest()
	}

	response = NewDeleteSecurityGroupResponse()
	err = c.Send(request, response)
	return
}

// DeleteSecurityGroup
// 本接口（DeleteSecurityGroup）用于删除安全组（SecurityGroup）。
//
// * 只有当前账号下的安全组允许被删除。
//
// * 安全组实例ID如果在其他安全组的规则中被引用，则无法直接删除。这种情况下，需要先进行规则修改，再删除安全组。
//
// * 删除的安全组无法再找回，请谨慎调用。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDSECURITYGROUPID_MALFORMED = "InvalidSecurityGroupID.Malformed"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupID.NotFound"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteSecurityGroupWithContext(ctx context.Context, request *DeleteSecurityGroupRequest) (response *DeleteSecurityGroupResponse, err error) {
	if request == nil {
		request = NewDeleteSecurityGroupRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteSecurityGroupResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteSecurityGroupPoliciesRequest() (request *DeleteSecurityGroupPoliciesRequest) {
	request = &DeleteSecurityGroupPoliciesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteSecurityGroupPolicies")

	return
}

func NewDeleteSecurityGroupPoliciesResponse() (response *DeleteSecurityGroupPoliciesResponse) {
	response = &DeleteSecurityGroupPoliciesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteSecurityGroupPolicies
// 本接口（DeleteSecurityGroupPolicies）用于用于删除安全组规则（SecurityGroupPolicy）。
//
// * SecurityGroupPolicySet.Version 用于指定要操作的安全组的版本。传入 Version 版本号若不等于当前安全组的最新版本，将返回失败；若不传 Version 则直接删除指定PolicyIndex的规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VERSIONMISMATCH = "UnsupportedOperation.VersionMismatch"
func (c *Client) DeleteSecurityGroupPolicies(request *DeleteSecurityGroupPoliciesRequest) (response *DeleteSecurityGroupPoliciesResponse, err error) {
	if request == nil {
		request = NewDeleteSecurityGroupPoliciesRequest()
	}

	response = NewDeleteSecurityGroupPoliciesResponse()
	err = c.Send(request, response)
	return
}

// DeleteSecurityGroupPolicies
// 本接口（DeleteSecurityGroupPolicies）用于用于删除安全组规则（SecurityGroupPolicy）。
//
// * SecurityGroupPolicySet.Version 用于指定要操作的安全组的版本。传入 Version 版本号若不等于当前安全组的最新版本，将返回失败；若不传 Version 则直接删除指定PolicyIndex的规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VERSIONMISMATCH = "UnsupportedOperation.VersionMismatch"
func (c *Client) DeleteSecurityGroupPoliciesWithContext(ctx context.Context, request *DeleteSecurityGroupPoliciesRequest) (response *DeleteSecurityGroupPoliciesResponse, err error) {
	if request == nil {
		request = NewDeleteSecurityGroupPoliciesRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteSecurityGroupPoliciesResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteServiceTemplateRequest() (request *DeleteServiceTemplateRequest) {
	request = &DeleteServiceTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteServiceTemplate")

	return
}

func NewDeleteServiceTemplateResponse() (response *DeleteServiceTemplateResponse) {
	response = &DeleteServiceTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteServiceTemplate
// 本接口（DeleteServiceTemplate）用于删除协议端口模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteServiceTemplate(request *DeleteServiceTemplateRequest) (response *DeleteServiceTemplateResponse, err error) {
	if request == nil {
		request = NewDeleteServiceTemplateRequest()
	}

	response = NewDeleteServiceTemplateResponse()
	err = c.Send(request, response)
	return
}

// DeleteServiceTemplate
// 本接口（DeleteServiceTemplate）用于删除协议端口模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteServiceTemplateWithContext(ctx context.Context, request *DeleteServiceTemplateRequest) (response *DeleteServiceTemplateResponse, err error) {
	if request == nil {
		request = NewDeleteServiceTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteServiceTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteServiceTemplateGroupRequest() (request *DeleteServiceTemplateGroupRequest) {
	request = &DeleteServiceTemplateGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteServiceTemplateGroup")

	return
}

func NewDeleteServiceTemplateGroupResponse() (response *DeleteServiceTemplateGroupResponse) {
	response = &DeleteServiceTemplateGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteServiceTemplateGroup
// 本接口（DeleteServiceTemplateGroup）用于删除协议端口模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteServiceTemplateGroup(request *DeleteServiceTemplateGroupRequest) (response *DeleteServiceTemplateGroupResponse, err error) {
	if request == nil {
		request = NewDeleteServiceTemplateGroupRequest()
	}

	response = NewDeleteServiceTemplateGroupResponse()
	err = c.Send(request, response)
	return
}

// DeleteServiceTemplateGroup
// 本接口（DeleteServiceTemplateGroup）用于删除协议端口模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteServiceTemplateGroupWithContext(ctx context.Context, request *DeleteServiceTemplateGroupRequest) (response *DeleteServiceTemplateGroupResponse, err error) {
	if request == nil {
		request = NewDeleteServiceTemplateGroupRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteServiceTemplateGroupResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteSubnetRequest() (request *DeleteSubnetRequest) {
	request = &DeleteSubnetRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteSubnet")

	return
}

func NewDeleteSubnetResponse() (response *DeleteSubnetResponse) {
	response = &DeleteSubnetResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteSubnet
// 本接口（DeleteSubnet）用于用于删除子网(Subnet)。
//
// * 删除子网前，请清理该子网下所有资源，包括云服务器、负载均衡、云数据、noSql、弹性网卡等资源。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteSubnet(request *DeleteSubnetRequest) (response *DeleteSubnetResponse, err error) {
	if request == nil {
		request = NewDeleteSubnetRequest()
	}

	response = NewDeleteSubnetResponse()
	err = c.Send(request, response)
	return
}

// DeleteSubnet
// 本接口（DeleteSubnet）用于用于删除子网(Subnet)。
//
// * 删除子网前，请清理该子网下所有资源，包括云服务器、负载均衡、云数据、noSql、弹性网卡等资源。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteSubnetWithContext(ctx context.Context, request *DeleteSubnetRequest) (response *DeleteSubnetResponse, err error) {
	if request == nil {
		request = NewDeleteSubnetRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteSubnetResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteTemplateMemberRequest() (request *DeleteTemplateMemberRequest) {
	request = &DeleteTemplateMemberRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteTemplateMember")

	return
}

func NewDeleteTemplateMemberResponse() (response *DeleteTemplateMemberResponse) {
	response = &DeleteTemplateMemberResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteTemplateMember
// 删除模版对象中的IP地址、协议端口、IP地址组、协议端口组。当前仅支持北京、泰国、北美地域请求。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteTemplateMember(request *DeleteTemplateMemberRequest) (response *DeleteTemplateMemberResponse, err error) {
	if request == nil {
		request = NewDeleteTemplateMemberRequest()
	}

	response = NewDeleteTemplateMemberResponse()
	err = c.Send(request, response)
	return
}

// DeleteTemplateMember
// 删除模版对象中的IP地址、协议端口、IP地址组、协议端口组。当前仅支持北京、泰国、北美地域请求。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteTemplateMemberWithContext(ctx context.Context, request *DeleteTemplateMemberRequest) (response *DeleteTemplateMemberResponse, err error) {
	if request == nil {
		request = NewDeleteTemplateMemberRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteTemplateMemberResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteVpcRequest() (request *DeleteVpcRequest) {
	request = &DeleteVpcRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteVpc")

	return
}

func NewDeleteVpcResponse() (response *DeleteVpcResponse) {
	response = &DeleteVpcResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteVpc
// 本接口（DeleteVpc）用于删除私有网络。
//
// * 删除前请确保 VPC 内已经没有相关资源，例如云服务器、云数据库、NoSQL、VPN网关、专线网关、负载均衡、对等连接、与之互通的基础网络设备等。
//
// * 删除私有网络是不可逆的操作，请谨慎处理。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_APPIDMISMATCH = "UnsupportedOperation.AppIdMismatch"
func (c *Client) DeleteVpc(request *DeleteVpcRequest) (response *DeleteVpcResponse, err error) {
	if request == nil {
		request = NewDeleteVpcRequest()
	}

	response = NewDeleteVpcResponse()
	err = c.Send(request, response)
	return
}

// DeleteVpc
// 本接口（DeleteVpc）用于删除私有网络。
//
// * 删除前请确保 VPC 内已经没有相关资源，例如云服务器、云数据库、NoSQL、VPN网关、专线网关、负载均衡、对等连接、与之互通的基础网络设备等。
//
// * 删除私有网络是不可逆的操作，请谨慎处理。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_APPIDMISMATCH = "UnsupportedOperation.AppIdMismatch"
func (c *Client) DeleteVpcWithContext(ctx context.Context, request *DeleteVpcRequest) (response *DeleteVpcResponse, err error) {
	if request == nil {
		request = NewDeleteVpcRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteVpcResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteVpcEndPointRequest() (request *DeleteVpcEndPointRequest) {
	request = &DeleteVpcEndPointRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteVpcEndPoint")

	return
}

func NewDeleteVpcEndPointResponse() (response *DeleteVpcEndPointResponse) {
	response = &DeleteVpcEndPointResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteVpcEndPoint
// 删除终端节点。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteVpcEndPoint(request *DeleteVpcEndPointRequest) (response *DeleteVpcEndPointResponse, err error) {
	if request == nil {
		request = NewDeleteVpcEndPointRequest()
	}

	response = NewDeleteVpcEndPointResponse()
	err = c.Send(request, response)
	return
}

// DeleteVpcEndPoint
// 删除终端节点。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteVpcEndPointWithContext(ctx context.Context, request *DeleteVpcEndPointRequest) (response *DeleteVpcEndPointResponse, err error) {
	if request == nil {
		request = NewDeleteVpcEndPointRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteVpcEndPointResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteVpcEndPointServiceRequest() (request *DeleteVpcEndPointServiceRequest) {
	request = &DeleteVpcEndPointServiceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteVpcEndPointService")

	return
}

func NewDeleteVpcEndPointServiceResponse() (response *DeleteVpcEndPointServiceResponse) {
	response = &DeleteVpcEndPointServiceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteVpcEndPointService
// 删除终端节点服务。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteVpcEndPointService(request *DeleteVpcEndPointServiceRequest) (response *DeleteVpcEndPointServiceResponse, err error) {
	if request == nil {
		request = NewDeleteVpcEndPointServiceRequest()
	}

	response = NewDeleteVpcEndPointServiceResponse()
	err = c.Send(request, response)
	return
}

// DeleteVpcEndPointService
// 删除终端节点服务。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteVpcEndPointServiceWithContext(ctx context.Context, request *DeleteVpcEndPointServiceRequest) (response *DeleteVpcEndPointServiceResponse, err error) {
	if request == nil {
		request = NewDeleteVpcEndPointServiceRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteVpcEndPointServiceResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteVpcEndPointServiceWhiteListRequest() (request *DeleteVpcEndPointServiceWhiteListRequest) {
	request = &DeleteVpcEndPointServiceWhiteListRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteVpcEndPointServiceWhiteList")

	return
}

func NewDeleteVpcEndPointServiceWhiteListResponse() (response *DeleteVpcEndPointServiceWhiteListResponse) {
	response = &DeleteVpcEndPointServiceWhiteListResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteVpcEndPointServiceWhiteList
// 删除终端节点服务白名单。
//
// 可能返回的错误码:
//
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
func (c *Client) DeleteVpcEndPointServiceWhiteList(request *DeleteVpcEndPointServiceWhiteListRequest) (response *DeleteVpcEndPointServiceWhiteListResponse, err error) {
	if request == nil {
		request = NewDeleteVpcEndPointServiceWhiteListRequest()
	}

	response = NewDeleteVpcEndPointServiceWhiteListResponse()
	err = c.Send(request, response)
	return
}

// DeleteVpcEndPointServiceWhiteList
// 删除终端节点服务白名单。
//
// 可能返回的错误码:
//
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
func (c *Client) DeleteVpcEndPointServiceWhiteListWithContext(ctx context.Context, request *DeleteVpcEndPointServiceWhiteListRequest) (response *DeleteVpcEndPointServiceWhiteListResponse, err error) {
	if request == nil {
		request = NewDeleteVpcEndPointServiceWhiteListRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteVpcEndPointServiceWhiteListResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteVpnConnectionRequest() (request *DeleteVpnConnectionRequest) {
	request = &DeleteVpnConnectionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteVpnConnection")

	return
}

func NewDeleteVpnConnectionResponse() (response *DeleteVpnConnectionResponse) {
	response = &DeleteVpnConnectionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteVpnConnection
// 本接口(DeleteVpnConnection)用于删除VPN通道。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteVpnConnection(request *DeleteVpnConnectionRequest) (response *DeleteVpnConnectionResponse, err error) {
	if request == nil {
		request = NewDeleteVpnConnectionRequest()
	}

	response = NewDeleteVpnConnectionResponse()
	err = c.Send(request, response)
	return
}

// DeleteVpnConnection
// 本接口(DeleteVpnConnection)用于删除VPN通道。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DeleteVpnConnectionWithContext(ctx context.Context, request *DeleteVpnConnectionRequest) (response *DeleteVpnConnectionResponse, err error) {
	if request == nil {
		request = NewDeleteVpnConnectionRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteVpnConnectionResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteVpnGatewayRequest() (request *DeleteVpnGatewayRequest) {
	request = &DeleteVpnGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteVpnGateway")

	return
}

func NewDeleteVpnGatewayResponse() (response *DeleteVpnGatewayResponse) {
	response = &DeleteVpnGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteVpnGateway
// 本接口（DeleteVpnGateway）用于删除VPN网关。目前只支持删除运行中的按量计费的IPSEC网关实例。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDVPNGATEWAYID_MALFORMED = "InvalidVpnGatewayId.Malformed"
//	INVALIDVPNGATEWAYID_NOTFOUND = "InvalidVpnGatewayId.NotFound"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteVpnGateway(request *DeleteVpnGatewayRequest) (response *DeleteVpnGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteVpnGatewayRequest()
	}

	response = NewDeleteVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

// DeleteVpnGateway
// 本接口（DeleteVpnGateway）用于删除VPN网关。目前只支持删除运行中的按量计费的IPSEC网关实例。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDVPNGATEWAYID_MALFORMED = "InvalidVpnGatewayId.Malformed"
//	INVALIDVPNGATEWAYID_NOTFOUND = "InvalidVpnGatewayId.NotFound"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteVpnGatewayWithContext(ctx context.Context, request *DeleteVpnGatewayRequest) (response *DeleteVpnGatewayResponse, err error) {
	if request == nil {
		request = NewDeleteVpnGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteVpnGatewayRoutesRequest() (request *DeleteVpnGatewayRoutesRequest) {
	request = &DeleteVpnGatewayRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DeleteVpnGatewayRoutes")

	return
}

func NewDeleteVpnGatewayRoutesResponse() (response *DeleteVpnGatewayRoutesResponse) {
	response = &DeleteVpnGatewayRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteVpnGatewayRoutes
// 本接口（DeleteVpnGatewayCcnRoutes）用于删除VPN网关路由
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteVpnGatewayRoutes(request *DeleteVpnGatewayRoutesRequest) (response *DeleteVpnGatewayRoutesResponse, err error) {
	if request == nil {
		request = NewDeleteVpnGatewayRoutesRequest()
	}

	response = NewDeleteVpnGatewayRoutesResponse()
	err = c.Send(request, response)
	return
}

// DeleteVpnGatewayRoutes
// 本接口（DeleteVpnGatewayCcnRoutes）用于删除VPN网关路由
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DeleteVpnGatewayRoutesWithContext(ctx context.Context, request *DeleteVpnGatewayRoutesRequest) (response *DeleteVpnGatewayRoutesResponse, err error) {
	if request == nil {
		request = NewDeleteVpnGatewayRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteVpnGatewayRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAccountAttributesRequest() (request *DescribeAccountAttributesRequest) {
	request = &DescribeAccountAttributesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeAccountAttributes")

	return
}

func NewDescribeAccountAttributesResponse() (response *DescribeAccountAttributesResponse) {
	response = &DescribeAccountAttributesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAccountAttributes
// 本接口（DescribeAccountAttributes）用于查询用户账号私有属性。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeAccountAttributes(request *DescribeAccountAttributesRequest) (response *DescribeAccountAttributesResponse, err error) {
	if request == nil {
		request = NewDescribeAccountAttributesRequest()
	}

	response = NewDescribeAccountAttributesResponse()
	err = c.Send(request, response)
	return
}

// DescribeAccountAttributes
// 本接口（DescribeAccountAttributes）用于查询用户账号私有属性。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeAccountAttributesWithContext(ctx context.Context, request *DescribeAccountAttributesRequest) (response *DescribeAccountAttributesResponse, err error) {
	if request == nil {
		request = NewDescribeAccountAttributesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAccountAttributesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAddressQuotaRequest() (request *DescribeAddressQuotaRequest) {
	request = &DescribeAddressQuotaRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeAddressQuota")

	return
}

func NewDescribeAddressQuotaResponse() (response *DescribeAddressQuotaResponse) {
	response = &DescribeAddressQuotaResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAddressQuota
// 本接口 (DescribeAddressQuota) 用于查询您账户的[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）在当前地域的配额信息。配额详情可参见 [EIP 产品简介](https://cloud.tencent.com/document/product/213/5733)。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
func (c *Client) DescribeAddressQuota(request *DescribeAddressQuotaRequest) (response *DescribeAddressQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeAddressQuotaRequest()
	}

	response = NewDescribeAddressQuotaResponse()
	err = c.Send(request, response)
	return
}

// DescribeAddressQuota
// 本接口 (DescribeAddressQuota) 用于查询您账户的[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）在当前地域的配额信息。配额详情可参见 [EIP 产品简介](https://cloud.tencent.com/document/product/213/5733)。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
func (c *Client) DescribeAddressQuotaWithContext(ctx context.Context, request *DescribeAddressQuotaRequest) (response *DescribeAddressQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeAddressQuotaRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAddressQuotaResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAddressTemplateGroupsRequest() (request *DescribeAddressTemplateGroupsRequest) {
	request = &DescribeAddressTemplateGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeAddressTemplateGroups")

	return
}

func NewDescribeAddressTemplateGroupsResponse() (response *DescribeAddressTemplateGroupsResponse) {
	response = &DescribeAddressTemplateGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAddressTemplateGroups
// 本接口（DescribeAddressTemplateGroups）用于查询IP地址模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeAddressTemplateGroups(request *DescribeAddressTemplateGroupsRequest) (response *DescribeAddressTemplateGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeAddressTemplateGroupsRequest()
	}

	response = NewDescribeAddressTemplateGroupsResponse()
	err = c.Send(request, response)
	return
}

// DescribeAddressTemplateGroups
// 本接口（DescribeAddressTemplateGroups）用于查询IP地址模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeAddressTemplateGroupsWithContext(ctx context.Context, request *DescribeAddressTemplateGroupsRequest) (response *DescribeAddressTemplateGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeAddressTemplateGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAddressTemplateGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAddressTemplatesRequest() (request *DescribeAddressTemplatesRequest) {
	request = &DescribeAddressTemplatesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeAddressTemplates")

	return
}

func NewDescribeAddressTemplatesResponse() (response *DescribeAddressTemplatesResponse) {
	response = &DescribeAddressTemplatesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAddressTemplates
// 本接口（DescribeAddressTemplates）用于查询IP地址模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeAddressTemplates(request *DescribeAddressTemplatesRequest) (response *DescribeAddressTemplatesResponse, err error) {
	if request == nil {
		request = NewDescribeAddressTemplatesRequest()
	}

	response = NewDescribeAddressTemplatesResponse()
	err = c.Send(request, response)
	return
}

// DescribeAddressTemplates
// 本接口（DescribeAddressTemplates）用于查询IP地址模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeAddressTemplatesWithContext(ctx context.Context, request *DescribeAddressTemplatesRequest) (response *DescribeAddressTemplatesResponse, err error) {
	if request == nil {
		request = NewDescribeAddressTemplatesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAddressTemplatesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAddressesRequest() (request *DescribeAddressesRequest) {
	request = &DescribeAddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeAddresses")

	return
}

func NewDescribeAddressesResponse() (response *DescribeAddressesResponse) {
	response = &DescribeAddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAddresses
// 本接口 (DescribeAddresses) 用于查询一个或多个[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）的详细信息。
//
// * 如果参数为空，返回当前用户一定数量（Limit所指定的数量，默认为20）的 EIP。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_INVALIDFILTER = "InvalidParameter.InvalidFilter"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
func (c *Client) DescribeAddresses(request *DescribeAddressesRequest) (response *DescribeAddressesResponse, err error) {
	if request == nil {
		request = NewDescribeAddressesRequest()
	}

	response = NewDescribeAddressesResponse()
	err = c.Send(request, response)
	return
}

// DescribeAddresses
// 本接口 (DescribeAddresses) 用于查询一个或多个[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）的详细信息。
//
// * 如果参数为空，返回当前用户一定数量（Limit所指定的数量，默认为20）的 EIP。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_INVALIDFILTER = "InvalidParameter.InvalidFilter"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
func (c *Client) DescribeAddressesWithContext(ctx context.Context, request *DescribeAddressesRequest) (response *DescribeAddressesResponse, err error) {
	if request == nil {
		request = NewDescribeAddressesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAddressesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAssistantCidrRequest() (request *DescribeAssistantCidrRequest) {
	request = &DescribeAssistantCidrRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeAssistantCidr")

	return
}

func NewDescribeAssistantCidrResponse() (response *DescribeAssistantCidrResponse) {
	response = &DescribeAssistantCidrResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAssistantCidr
// 本接口（DescribeAssistantCidr）用于查询辅助CIDR列表。（接口灰度中，如需使用请提工单。）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeAssistantCidr(request *DescribeAssistantCidrRequest) (response *DescribeAssistantCidrResponse, err error) {
	if request == nil {
		request = NewDescribeAssistantCidrRequest()
	}

	response = NewDescribeAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

// DescribeAssistantCidr
// 本接口（DescribeAssistantCidr）用于查询辅助CIDR列表。（接口灰度中，如需使用请提工单。）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeAssistantCidrWithContext(ctx context.Context, request *DescribeAssistantCidrRequest) (response *DescribeAssistantCidrResponse, err error) {
	if request == nil {
		request = NewDescribeAssistantCidrRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeBandwidthPackageBillUsageRequest() (request *DescribeBandwidthPackageBillUsageRequest) {
	request = &DescribeBandwidthPackageBillUsageRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeBandwidthPackageBillUsage")

	return
}

func NewDescribeBandwidthPackageBillUsageResponse() (response *DescribeBandwidthPackageBillUsageResponse) {
	response = &DescribeBandwidthPackageBillUsageResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeBandwidthPackageBillUsage
// 本接口 (DescribeBandwidthPackageBillUsage) 用于查询后付费共享带宽包当前的计费用量.
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
func (c *Client) DescribeBandwidthPackageBillUsage(request *DescribeBandwidthPackageBillUsageRequest) (response *DescribeBandwidthPackageBillUsageResponse, err error) {
	if request == nil {
		request = NewDescribeBandwidthPackageBillUsageRequest()
	}

	response = NewDescribeBandwidthPackageBillUsageResponse()
	err = c.Send(request, response)
	return
}

// DescribeBandwidthPackageBillUsage
// 本接口 (DescribeBandwidthPackageBillUsage) 用于查询后付费共享带宽包当前的计费用量.
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
func (c *Client) DescribeBandwidthPackageBillUsageWithContext(ctx context.Context, request *DescribeBandwidthPackageBillUsageRequest) (response *DescribeBandwidthPackageBillUsageResponse, err error) {
	if request == nil {
		request = NewDescribeBandwidthPackageBillUsageRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeBandwidthPackageBillUsageResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeBandwidthPackageQuotaRequest() (request *DescribeBandwidthPackageQuotaRequest) {
	request = &DescribeBandwidthPackageQuotaRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeBandwidthPackageQuota")

	return
}

func NewDescribeBandwidthPackageQuotaResponse() (response *DescribeBandwidthPackageQuotaResponse) {
	response = &DescribeBandwidthPackageQuotaResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeBandwidthPackageQuota
// 接口用于查询账户在当前地域的带宽包上限数量以及使用数量
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
func (c *Client) DescribeBandwidthPackageQuota(request *DescribeBandwidthPackageQuotaRequest) (response *DescribeBandwidthPackageQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeBandwidthPackageQuotaRequest()
	}

	response = NewDescribeBandwidthPackageQuotaResponse()
	err = c.Send(request, response)
	return
}

// DescribeBandwidthPackageQuota
// 接口用于查询账户在当前地域的带宽包上限数量以及使用数量
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
func (c *Client) DescribeBandwidthPackageQuotaWithContext(ctx context.Context, request *DescribeBandwidthPackageQuotaRequest) (response *DescribeBandwidthPackageQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeBandwidthPackageQuotaRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeBandwidthPackageQuotaResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeBandwidthPackageResourcesRequest() (request *DescribeBandwidthPackageResourcesRequest) {
	request = &DescribeBandwidthPackageResourcesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeBandwidthPackageResources")

	return
}

func NewDescribeBandwidthPackageResourcesResponse() (response *DescribeBandwidthPackageResourcesResponse) {
	response = &DescribeBandwidthPackageResourcesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeBandwidthPackageResources
// 本接口 (DescribeBandwidthPackageResources) 用于根据共享带宽包唯一ID查询共享带宽包内的资源列表，支持按条件过滤查询结果和分页查询。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_RESOURCEIDMALFORMED = "InvalidParameterValue.ResourceIdMalformed"
func (c *Client) DescribeBandwidthPackageResources(request *DescribeBandwidthPackageResourcesRequest) (response *DescribeBandwidthPackageResourcesResponse, err error) {
	if request == nil {
		request = NewDescribeBandwidthPackageResourcesRequest()
	}

	response = NewDescribeBandwidthPackageResourcesResponse()
	err = c.Send(request, response)
	return
}

// DescribeBandwidthPackageResources
// 本接口 (DescribeBandwidthPackageResources) 用于根据共享带宽包唯一ID查询共享带宽包内的资源列表，支持按条件过滤查询结果和分页查询。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_RESOURCEIDMALFORMED = "InvalidParameterValue.ResourceIdMalformed"
func (c *Client) DescribeBandwidthPackageResourcesWithContext(ctx context.Context, request *DescribeBandwidthPackageResourcesRequest) (response *DescribeBandwidthPackageResourcesResponse, err error) {
	if request == nil {
		request = NewDescribeBandwidthPackageResourcesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeBandwidthPackageResourcesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeBandwidthPackagesRequest() (request *DescribeBandwidthPackagesRequest) {
	request = &DescribeBandwidthPackagesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeBandwidthPackages")

	return
}

func NewDescribeBandwidthPackagesResponse() (response *DescribeBandwidthPackagesResponse) {
	response = &DescribeBandwidthPackagesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeBandwidthPackages
// 接口用于查询带宽包详细信息，包括带宽包唯一标识ID，类型，计费模式，名称，资源信息等
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_INVALIDBANDWIDTHPACKAGECHARGETYPE = "InvalidParameterValue.InvalidBandwidthPackageChargeType"
func (c *Client) DescribeBandwidthPackages(request *DescribeBandwidthPackagesRequest) (response *DescribeBandwidthPackagesResponse, err error) {
	if request == nil {
		request = NewDescribeBandwidthPackagesRequest()
	}

	response = NewDescribeBandwidthPackagesResponse()
	err = c.Send(request, response)
	return
}

// DescribeBandwidthPackages
// 接口用于查询带宽包详细信息，包括带宽包唯一标识ID，类型，计费模式，名称，资源信息等
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_INVALIDBANDWIDTHPACKAGECHARGETYPE = "InvalidParameterValue.InvalidBandwidthPackageChargeType"
func (c *Client) DescribeBandwidthPackagesWithContext(ctx context.Context, request *DescribeBandwidthPackagesRequest) (response *DescribeBandwidthPackagesResponse, err error) {
	if request == nil {
		request = NewDescribeBandwidthPackagesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeBandwidthPackagesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCcnAttachedInstancesRequest() (request *DescribeCcnAttachedInstancesRequest) {
	request = &DescribeCcnAttachedInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeCcnAttachedInstances")

	return
}

func NewDescribeCcnAttachedInstancesResponse() (response *DescribeCcnAttachedInstancesResponse) {
	response = &DescribeCcnAttachedInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeCcnAttachedInstances
// 本接口（DescribeCcnAttachedInstances）用于查询云联网实例下已关联的网络实例。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	UNSUPPORTEDOPERATION_APPIDNOTFOUND = "UnsupportedOperation.AppIdNotFound"
func (c *Client) DescribeCcnAttachedInstances(request *DescribeCcnAttachedInstancesRequest) (response *DescribeCcnAttachedInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeCcnAttachedInstancesRequest()
	}

	response = NewDescribeCcnAttachedInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeCcnAttachedInstances
// 本接口（DescribeCcnAttachedInstances）用于查询云联网实例下已关联的网络实例。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	UNSUPPORTEDOPERATION_APPIDNOTFOUND = "UnsupportedOperation.AppIdNotFound"
func (c *Client) DescribeCcnAttachedInstancesWithContext(ctx context.Context, request *DescribeCcnAttachedInstancesRequest) (response *DescribeCcnAttachedInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeCcnAttachedInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeCcnAttachedInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCcnRegionBandwidthLimitsRequest() (request *DescribeCcnRegionBandwidthLimitsRequest) {
	request = &DescribeCcnRegionBandwidthLimitsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeCcnRegionBandwidthLimits")

	return
}

func NewDescribeCcnRegionBandwidthLimitsResponse() (response *DescribeCcnRegionBandwidthLimitsResponse) {
	response = &DescribeCcnRegionBandwidthLimitsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeCcnRegionBandwidthLimits
// 本接口（DescribeCcnRegionBandwidthLimits）用于查询云联网各地域出带宽上限，该接口只返回已关联网络实例包含的地域
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
func (c *Client) DescribeCcnRegionBandwidthLimits(request *DescribeCcnRegionBandwidthLimitsRequest) (response *DescribeCcnRegionBandwidthLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeCcnRegionBandwidthLimitsRequest()
	}

	response = NewDescribeCcnRegionBandwidthLimitsResponse()
	err = c.Send(request, response)
	return
}

// DescribeCcnRegionBandwidthLimits
// 本接口（DescribeCcnRegionBandwidthLimits）用于查询云联网各地域出带宽上限，该接口只返回已关联网络实例包含的地域
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
func (c *Client) DescribeCcnRegionBandwidthLimitsWithContext(ctx context.Context, request *DescribeCcnRegionBandwidthLimitsRequest) (response *DescribeCcnRegionBandwidthLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeCcnRegionBandwidthLimitsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeCcnRegionBandwidthLimitsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCcnRoutesRequest() (request *DescribeCcnRoutesRequest) {
	request = &DescribeCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeCcnRoutes")

	return
}

func NewDescribeCcnRoutesResponse() (response *DescribeCcnRoutesResponse) {
	response = &DescribeCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeCcnRoutes
// 本接口（DescribeCcnRoutes）用于查询已加入云联网（CCN）的路由
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeCcnRoutes(request *DescribeCcnRoutesRequest) (response *DescribeCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeCcnRoutesRequest()
	}

	response = NewDescribeCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// DescribeCcnRoutes
// 本接口（DescribeCcnRoutes）用于查询已加入云联网（CCN）的路由
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeCcnRoutesWithContext(ctx context.Context, request *DescribeCcnRoutesRequest) (response *DescribeCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCcnsRequest() (request *DescribeCcnsRequest) {
	request = &DescribeCcnsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeCcns")

	return
}

func NewDescribeCcnsResponse() (response *DescribeCcnsResponse) {
	response = &DescribeCcnsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeCcns
// 本接口（DescribeCcns）用于查询云联网（CCN）列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeCcns(request *DescribeCcnsRequest) (response *DescribeCcnsResponse, err error) {
	if request == nil {
		request = NewDescribeCcnsRequest()
	}

	response = NewDescribeCcnsResponse()
	err = c.Send(request, response)
	return
}

// DescribeCcns
// 本接口（DescribeCcns）用于查询云联网（CCN）列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeCcnsWithContext(ctx context.Context, request *DescribeCcnsRequest) (response *DescribeCcnsResponse, err error) {
	if request == nil {
		request = NewDescribeCcnsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeCcnsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClassicLinkInstancesRequest() (request *DescribeClassicLinkInstancesRequest) {
	request = &DescribeClassicLinkInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeClassicLinkInstances")

	return
}

func NewDescribeClassicLinkInstancesResponse() (response *DescribeClassicLinkInstancesResponse) {
	response = &DescribeClassicLinkInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClassicLinkInstances
// 本接口(DescribeClassicLinkInstances)用于查询私有网络和基础网络设备互通列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeClassicLinkInstances(request *DescribeClassicLinkInstancesRequest) (response *DescribeClassicLinkInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeClassicLinkInstancesRequest()
	}

	response = NewDescribeClassicLinkInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeClassicLinkInstances
// 本接口(DescribeClassicLinkInstances)用于查询私有网络和基础网络设备互通列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeClassicLinkInstancesWithContext(ctx context.Context, request *DescribeClassicLinkInstancesRequest) (response *DescribeClassicLinkInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeClassicLinkInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClassicLinkInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCrossBorderComplianceRequest() (request *DescribeCrossBorderComplianceRequest) {
	request = &DescribeCrossBorderComplianceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeCrossBorderCompliance")

	return
}

func NewDescribeCrossBorderComplianceResponse() (response *DescribeCrossBorderComplianceResponse) {
	response = &DescribeCrossBorderComplianceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeCrossBorderCompliance
// 本接口（DescribeCrossBorderCompliance）用于查询用户创建的合规化资质审批单。
//
// 服务商可以查询服务名下的任意 `APPID` 创建的审批单；非服务商，只能查询自己审批单。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeCrossBorderCompliance(request *DescribeCrossBorderComplianceRequest) (response *DescribeCrossBorderComplianceResponse, err error) {
	if request == nil {
		request = NewDescribeCrossBorderComplianceRequest()
	}

	response = NewDescribeCrossBorderComplianceResponse()
	err = c.Send(request, response)
	return
}

// DescribeCrossBorderCompliance
// 本接口（DescribeCrossBorderCompliance）用于查询用户创建的合规化资质审批单。
//
// 服务商可以查询服务名下的任意 `APPID` 创建的审批单；非服务商，只能查询自己审批单。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeCrossBorderComplianceWithContext(ctx context.Context, request *DescribeCrossBorderComplianceRequest) (response *DescribeCrossBorderComplianceResponse, err error) {
	if request == nil {
		request = NewDescribeCrossBorderComplianceRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeCrossBorderComplianceResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCustomerGatewayVendorsRequest() (request *DescribeCustomerGatewayVendorsRequest) {
	request = &DescribeCustomerGatewayVendorsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeCustomerGatewayVendors")

	return
}

func NewDescribeCustomerGatewayVendorsResponse() (response *DescribeCustomerGatewayVendorsResponse) {
	response = &DescribeCustomerGatewayVendorsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeCustomerGatewayVendors
// 本接口（DescribeCustomerGatewayVendors）用于查询可支持的对端网关厂商信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeCustomerGatewayVendors(request *DescribeCustomerGatewayVendorsRequest) (response *DescribeCustomerGatewayVendorsResponse, err error) {
	if request == nil {
		request = NewDescribeCustomerGatewayVendorsRequest()
	}

	response = NewDescribeCustomerGatewayVendorsResponse()
	err = c.Send(request, response)
	return
}

// DescribeCustomerGatewayVendors
// 本接口（DescribeCustomerGatewayVendors）用于查询可支持的对端网关厂商信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeCustomerGatewayVendorsWithContext(ctx context.Context, request *DescribeCustomerGatewayVendorsRequest) (response *DescribeCustomerGatewayVendorsResponse, err error) {
	if request == nil {
		request = NewDescribeCustomerGatewayVendorsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeCustomerGatewayVendorsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeCustomerGatewaysRequest() (request *DescribeCustomerGatewaysRequest) {
	request = &DescribeCustomerGatewaysRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeCustomerGateways")

	return
}

func NewDescribeCustomerGatewaysResponse() (response *DescribeCustomerGatewaysResponse) {
	response = &DescribeCustomerGatewaysResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeCustomerGateways
// 本接口（DescribeCustomerGateways）用于查询对端网关列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeCustomerGateways(request *DescribeCustomerGatewaysRequest) (response *DescribeCustomerGatewaysResponse, err error) {
	if request == nil {
		request = NewDescribeCustomerGatewaysRequest()
	}

	response = NewDescribeCustomerGatewaysResponse()
	err = c.Send(request, response)
	return
}

// DescribeCustomerGateways
// 本接口（DescribeCustomerGateways）用于查询对端网关列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeCustomerGatewaysWithContext(ctx context.Context, request *DescribeCustomerGatewaysRequest) (response *DescribeCustomerGatewaysResponse, err error) {
	if request == nil {
		request = NewDescribeCustomerGatewaysRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeCustomerGatewaysResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeDhcpIpsRequest() (request *DescribeDhcpIpsRequest) {
	request = &DescribeDhcpIpsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeDhcpIps")

	return
}

func NewDescribeDhcpIpsResponse() (response *DescribeDhcpIpsResponse) {
	response = &DescribeDhcpIpsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeDhcpIps
// 本接口（DescribeDhcpIps）用于查询DhcpIp列表
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeDhcpIps(request *DescribeDhcpIpsRequest) (response *DescribeDhcpIpsResponse, err error) {
	if request == nil {
		request = NewDescribeDhcpIpsRequest()
	}

	response = NewDescribeDhcpIpsResponse()
	err = c.Send(request, response)
	return
}

// DescribeDhcpIps
// 本接口（DescribeDhcpIps）用于查询DhcpIp列表
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeDhcpIpsWithContext(ctx context.Context, request *DescribeDhcpIpsRequest) (response *DescribeDhcpIpsResponse, err error) {
	if request == nil {
		request = NewDescribeDhcpIpsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeDhcpIpsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeDirectConnectGatewayCcnRoutesRequest() (request *DescribeDirectConnectGatewayCcnRoutesRequest) {
	request = &DescribeDirectConnectGatewayCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeDirectConnectGatewayCcnRoutes")

	return
}

func NewDescribeDirectConnectGatewayCcnRoutesResponse() (response *DescribeDirectConnectGatewayCcnRoutesResponse) {
	response = &DescribeDirectConnectGatewayCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeDirectConnectGatewayCcnRoutes
// 本接口（DescribeDirectConnectGatewayCcnRoutes）用于查询专线网关的云联网路由（IDC网段）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeDirectConnectGatewayCcnRoutes(request *DescribeDirectConnectGatewayCcnRoutesRequest) (response *DescribeDirectConnectGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeDirectConnectGatewayCcnRoutesRequest()
	}

	response = NewDescribeDirectConnectGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// DescribeDirectConnectGatewayCcnRoutes
// 本接口（DescribeDirectConnectGatewayCcnRoutes）用于查询专线网关的云联网路由（IDC网段）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeDirectConnectGatewayCcnRoutesWithContext(ctx context.Context, request *DescribeDirectConnectGatewayCcnRoutesRequest) (response *DescribeDirectConnectGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeDirectConnectGatewayCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeDirectConnectGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeDirectConnectGatewaysRequest() (request *DescribeDirectConnectGatewaysRequest) {
	request = &DescribeDirectConnectGatewaysRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeDirectConnectGateways")

	return
}

func NewDescribeDirectConnectGatewaysResponse() (response *DescribeDirectConnectGatewaysResponse) {
	response = &DescribeDirectConnectGatewaysResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeDirectConnectGateways
// 本接口（DescribeDirectConnectGateways）用于查询专线网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
func (c *Client) DescribeDirectConnectGateways(request *DescribeDirectConnectGatewaysRequest) (response *DescribeDirectConnectGatewaysResponse, err error) {
	if request == nil {
		request = NewDescribeDirectConnectGatewaysRequest()
	}

	response = NewDescribeDirectConnectGatewaysResponse()
	err = c.Send(request, response)
	return
}

// DescribeDirectConnectGateways
// 本接口（DescribeDirectConnectGateways）用于查询专线网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
func (c *Client) DescribeDirectConnectGatewaysWithContext(ctx context.Context, request *DescribeDirectConnectGatewaysRequest) (response *DescribeDirectConnectGatewaysResponse, err error) {
	if request == nil {
		request = NewDescribeDirectConnectGatewaysRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeDirectConnectGatewaysResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeFlowLogRequest() (request *DescribeFlowLogRequest) {
	request = &DescribeFlowLogRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeFlowLog")

	return
}

func NewDescribeFlowLogResponse() (response *DescribeFlowLogResponse) {
	response = &DescribeFlowLogResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeFlowLog
// 本接口（DescribeFlowLog）用于查询流日志实例信息
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeFlowLog(request *DescribeFlowLogRequest) (response *DescribeFlowLogResponse, err error) {
	if request == nil {
		request = NewDescribeFlowLogRequest()
	}

	response = NewDescribeFlowLogResponse()
	err = c.Send(request, response)
	return
}

// DescribeFlowLog
// 本接口（DescribeFlowLog）用于查询流日志实例信息
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeFlowLogWithContext(ctx context.Context, request *DescribeFlowLogRequest) (response *DescribeFlowLogResponse, err error) {
	if request == nil {
		request = NewDescribeFlowLogRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeFlowLogResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeFlowLogsRequest() (request *DescribeFlowLogsRequest) {
	request = &DescribeFlowLogsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeFlowLogs")

	return
}

func NewDescribeFlowLogsResponse() (response *DescribeFlowLogsResponse) {
	response = &DescribeFlowLogsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeFlowLogs
// 本接口（DescribeFlowLogs）用于查询获取流日志集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeFlowLogs(request *DescribeFlowLogsRequest) (response *DescribeFlowLogsResponse, err error) {
	if request == nil {
		request = NewDescribeFlowLogsRequest()
	}

	response = NewDescribeFlowLogsResponse()
	err = c.Send(request, response)
	return
}

// DescribeFlowLogs
// 本接口（DescribeFlowLogs）用于查询获取流日志集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeFlowLogsWithContext(ctx context.Context, request *DescribeFlowLogsRequest) (response *DescribeFlowLogsResponse, err error) {
	if request == nil {
		request = NewDescribeFlowLogsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeFlowLogsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeGatewayFlowMonitorDetailRequest() (request *DescribeGatewayFlowMonitorDetailRequest) {
	request = &DescribeGatewayFlowMonitorDetailRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeGatewayFlowMonitorDetail")

	return
}

func NewDescribeGatewayFlowMonitorDetailResponse() (response *DescribeGatewayFlowMonitorDetailResponse) {
	response = &DescribeGatewayFlowMonitorDetailResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeGatewayFlowMonitorDetail
// 本接口（DescribeGatewayFlowMonitorDetail）用于查询网关流量监控明细。
//
// * 只支持单个网关实例查询。即入参 `VpnId`、 `DirectConnectGatewayId`、 `PeeringConnectionId`、 `NatId` 最多只支持传一个，且必须传一个。
//
// * 如果网关有流量，但调用本接口没有返回数据，请在控制台对应网关详情页确认是否开启网关流量监控。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeGatewayFlowMonitorDetail(request *DescribeGatewayFlowMonitorDetailRequest) (response *DescribeGatewayFlowMonitorDetailResponse, err error) {
	if request == nil {
		request = NewDescribeGatewayFlowMonitorDetailRequest()
	}

	response = NewDescribeGatewayFlowMonitorDetailResponse()
	err = c.Send(request, response)
	return
}

// DescribeGatewayFlowMonitorDetail
// 本接口（DescribeGatewayFlowMonitorDetail）用于查询网关流量监控明细。
//
// * 只支持单个网关实例查询。即入参 `VpnId`、 `DirectConnectGatewayId`、 `PeeringConnectionId`、 `NatId` 最多只支持传一个，且必须传一个。
//
// * 如果网关有流量，但调用本接口没有返回数据，请在控制台对应网关详情页确认是否开启网关流量监控。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeGatewayFlowMonitorDetailWithContext(ctx context.Context, request *DescribeGatewayFlowMonitorDetailRequest) (response *DescribeGatewayFlowMonitorDetailResponse, err error) {
	if request == nil {
		request = NewDescribeGatewayFlowMonitorDetailRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeGatewayFlowMonitorDetailResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeGatewayFlowQosRequest() (request *DescribeGatewayFlowQosRequest) {
	request = &DescribeGatewayFlowQosRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeGatewayFlowQos")

	return
}

func NewDescribeGatewayFlowQosResponse() (response *DescribeGatewayFlowQosResponse) {
	response = &DescribeGatewayFlowQosResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeGatewayFlowQos
// 本接口（DescribeGatewayFlowQos）用于查询网关来访IP流控带宽。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) DescribeGatewayFlowQos(request *DescribeGatewayFlowQosRequest) (response *DescribeGatewayFlowQosResponse, err error) {
	if request == nil {
		request = NewDescribeGatewayFlowQosRequest()
	}

	response = NewDescribeGatewayFlowQosResponse()
	err = c.Send(request, response)
	return
}

// DescribeGatewayFlowQos
// 本接口（DescribeGatewayFlowQos）用于查询网关来访IP流控带宽。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) DescribeGatewayFlowQosWithContext(ctx context.Context, request *DescribeGatewayFlowQosRequest) (response *DescribeGatewayFlowQosResponse, err error) {
	if request == nil {
		request = NewDescribeGatewayFlowQosRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeGatewayFlowQosResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeHaVipsRequest() (request *DescribeHaVipsRequest) {
	request = &DescribeHaVipsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeHaVips")

	return
}

func NewDescribeHaVipsResponse() (response *DescribeHaVipsResponse) {
	response = &DescribeHaVipsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeHaVips
// 本接口（DescribeHaVips）用于查询高可用虚拟IP（HAVIP）列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeHaVips(request *DescribeHaVipsRequest) (response *DescribeHaVipsResponse, err error) {
	if request == nil {
		request = NewDescribeHaVipsRequest()
	}

	response = NewDescribeHaVipsResponse()
	err = c.Send(request, response)
	return
}

// DescribeHaVips
// 本接口（DescribeHaVips）用于查询高可用虚拟IP（HAVIP）列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeHaVipsWithContext(ctx context.Context, request *DescribeHaVipsRequest) (response *DescribeHaVipsResponse, err error) {
	if request == nil {
		request = NewDescribeHaVipsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeHaVipsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeIp6AddressesRequest() (request *DescribeIp6AddressesRequest) {
	request = &DescribeIp6AddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeIp6Addresses")

	return
}

func NewDescribeIp6AddressesResponse() (response *DescribeIp6AddressesResponse) {
	response = &DescribeIp6AddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeIp6Addresses
// 该接口用于查询IPV6地址信息
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_NETWORKINTERFACEIDMALFORMED = "InvalidParameterValue.NetworkInterfaceIdMalformed"
func (c *Client) DescribeIp6Addresses(request *DescribeIp6AddressesRequest) (response *DescribeIp6AddressesResponse, err error) {
	if request == nil {
		request = NewDescribeIp6AddressesRequest()
	}

	response = NewDescribeIp6AddressesResponse()
	err = c.Send(request, response)
	return
}

// DescribeIp6Addresses
// 该接口用于查询IPV6地址信息
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_NETWORKINTERFACEIDMALFORMED = "InvalidParameterValue.NetworkInterfaceIdMalformed"
func (c *Client) DescribeIp6AddressesWithContext(ctx context.Context, request *DescribeIp6AddressesRequest) (response *DescribeIp6AddressesResponse, err error) {
	if request == nil {
		request = NewDescribeIp6AddressesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeIp6AddressesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeIp6TranslatorQuotaRequest() (request *DescribeIp6TranslatorQuotaRequest) {
	request = &DescribeIp6TranslatorQuotaRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeIp6TranslatorQuota")

	return
}

func NewDescribeIp6TranslatorQuotaResponse() (response *DescribeIp6TranslatorQuotaResponse) {
	response = &DescribeIp6TranslatorQuotaResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeIp6TranslatorQuota
// 查询账户在指定地域IPV6转换实例和规则的配额
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
func (c *Client) DescribeIp6TranslatorQuota(request *DescribeIp6TranslatorQuotaRequest) (response *DescribeIp6TranslatorQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeIp6TranslatorQuotaRequest()
	}

	response = NewDescribeIp6TranslatorQuotaResponse()
	err = c.Send(request, response)
	return
}

// DescribeIp6TranslatorQuota
// 查询账户在指定地域IPV6转换实例和规则的配额
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
func (c *Client) DescribeIp6TranslatorQuotaWithContext(ctx context.Context, request *DescribeIp6TranslatorQuotaRequest) (response *DescribeIp6TranslatorQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeIp6TranslatorQuotaRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeIp6TranslatorQuotaResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeIp6TranslatorsRequest() (request *DescribeIp6TranslatorsRequest) {
	request = &DescribeIp6TranslatorsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeIp6Translators")

	return
}

func NewDescribeIp6TranslatorsResponse() (response *DescribeIp6TranslatorsResponse) {
	response = &DescribeIp6TranslatorsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeIp6Translators
// 1. 该接口用于查询账户下的IPV6转换实例及其绑定的转换规则信息
//
// 2. 支持过滤查询
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribeIp6Translators(request *DescribeIp6TranslatorsRequest) (response *DescribeIp6TranslatorsResponse, err error) {
	if request == nil {
		request = NewDescribeIp6TranslatorsRequest()
	}

	response = NewDescribeIp6TranslatorsResponse()
	err = c.Send(request, response)
	return
}

// DescribeIp6Translators
// 1. 该接口用于查询账户下的IPV6转换实例及其绑定的转换规则信息
//
// 2. 支持过滤查询
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribeIp6TranslatorsWithContext(ctx context.Context, request *DescribeIp6TranslatorsRequest) (response *DescribeIp6TranslatorsResponse, err error) {
	if request == nil {
		request = NewDescribeIp6TranslatorsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeIp6TranslatorsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeIpGeolocationDatabaseUrlRequest() (request *DescribeIpGeolocationDatabaseUrlRequest) {
	request = &DescribeIpGeolocationDatabaseUrlRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeIpGeolocationDatabaseUrl")

	return
}

func NewDescribeIpGeolocationDatabaseUrlResponse() (response *DescribeIpGeolocationDatabaseUrlResponse) {
	response = &DescribeIpGeolocationDatabaseUrlResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeIpGeolocationDatabaseUrl
// 本接口（DescribeIpGeolocationDatabaseUrl）用于获取IP地理位置库下载链接。
//
// 可能返回的错误码:
//
//	AUTHFAILURE = "AuthFailure"
//	INTERNALERROR = "InternalError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DescribeIpGeolocationDatabaseUrl(request *DescribeIpGeolocationDatabaseUrlRequest) (response *DescribeIpGeolocationDatabaseUrlResponse, err error) {
	if request == nil {
		request = NewDescribeIpGeolocationDatabaseUrlRequest()
	}

	response = NewDescribeIpGeolocationDatabaseUrlResponse()
	err = c.Send(request, response)
	return
}

// DescribeIpGeolocationDatabaseUrl
// 本接口（DescribeIpGeolocationDatabaseUrl）用于获取IP地理位置库下载链接。
//
// 可能返回的错误码:
//
//	AUTHFAILURE = "AuthFailure"
//	INTERNALERROR = "InternalError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DescribeIpGeolocationDatabaseUrlWithContext(ctx context.Context, request *DescribeIpGeolocationDatabaseUrlRequest) (response *DescribeIpGeolocationDatabaseUrlResponse, err error) {
	if request == nil {
		request = NewDescribeIpGeolocationDatabaseUrlRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeIpGeolocationDatabaseUrlResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeIpGeolocationInfosRequest() (request *DescribeIpGeolocationInfosRequest) {
	request = &DescribeIpGeolocationInfosRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeIpGeolocationInfos")

	return
}

func NewDescribeIpGeolocationInfosResponse() (response *DescribeIpGeolocationInfosResponse) {
	response = &DescribeIpGeolocationInfosResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeIpGeolocationInfos
// 本接口（DescribeIpGeolocationInfos）用于查询IP地址信息，包括地理位置信息和网络信息。
//
// 本接口目前处于内测中，如需使用，请提交 [工单申请](https://console.cloud.tencent.com/workorder/category?level1_id=6&level2_id=660&source=0&data_title=%E5%BC%B9%E6%80%A7%E5%85%AC%E7%BD%91%20EIP&level3_id=662&queue=96&scene_code=16400&step=2)。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_COMBINATION = "InvalidParameterValue.Combination"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeIpGeolocationInfos(request *DescribeIpGeolocationInfosRequest) (response *DescribeIpGeolocationInfosResponse, err error) {
	if request == nil {
		request = NewDescribeIpGeolocationInfosRequest()
	}

	response = NewDescribeIpGeolocationInfosResponse()
	err = c.Send(request, response)
	return
}

// DescribeIpGeolocationInfos
// 本接口（DescribeIpGeolocationInfos）用于查询IP地址信息，包括地理位置信息和网络信息。
//
// 本接口目前处于内测中，如需使用，请提交 [工单申请](https://console.cloud.tencent.com/workorder/category?level1_id=6&level2_id=660&source=0&data_title=%E5%BC%B9%E6%80%A7%E5%85%AC%E7%BD%91%20EIP&level3_id=662&queue=96&scene_code=16400&step=2)。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_COMBINATION = "InvalidParameterValue.Combination"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeIpGeolocationInfosWithContext(ctx context.Context, request *DescribeIpGeolocationInfosRequest) (response *DescribeIpGeolocationInfosResponse, err error) {
	if request == nil {
		request = NewDescribeIpGeolocationInfosRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeIpGeolocationInfosResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeLocalGatewayRequest() (request *DescribeLocalGatewayRequest) {
	request = &DescribeLocalGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeLocalGateway")

	return
}

func NewDescribeLocalGatewayResponse() (response *DescribeLocalGatewayResponse) {
	response = &DescribeLocalGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeLocalGateway
// 该接口用于查询CDC的本地网关。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeLocalGateway(request *DescribeLocalGatewayRequest) (response *DescribeLocalGatewayResponse, err error) {
	if request == nil {
		request = NewDescribeLocalGatewayRequest()
	}

	response = NewDescribeLocalGatewayResponse()
	err = c.Send(request, response)
	return
}

// DescribeLocalGateway
// 该接口用于查询CDC的本地网关。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeLocalGatewayWithContext(ctx context.Context, request *DescribeLocalGatewayRequest) (response *DescribeLocalGatewayResponse, err error) {
	if request == nil {
		request = NewDescribeLocalGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeLocalGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNatGatewayDestinationIpPortTranslationNatRulesRequest() (request *DescribeNatGatewayDestinationIpPortTranslationNatRulesRequest) {
	request = &DescribeNatGatewayDestinationIpPortTranslationNatRulesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNatGatewayDestinationIpPortTranslationNatRules")

	return
}

func NewDescribeNatGatewayDestinationIpPortTranslationNatRulesResponse() (response *DescribeNatGatewayDestinationIpPortTranslationNatRulesResponse) {
	response = &DescribeNatGatewayDestinationIpPortTranslationNatRulesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNatGatewayDestinationIpPortTranslationNatRules
// 本接口（DescribeNatGatewayDestinationIpPortTranslationNatRules）用于查询NAT网关端口转发规则对象数组。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeNatGatewayDestinationIpPortTranslationNatRules(request *DescribeNatGatewayDestinationIpPortTranslationNatRulesRequest) (response *DescribeNatGatewayDestinationIpPortTranslationNatRulesResponse, err error) {
	if request == nil {
		request = NewDescribeNatGatewayDestinationIpPortTranslationNatRulesRequest()
	}

	response = NewDescribeNatGatewayDestinationIpPortTranslationNatRulesResponse()
	err = c.Send(request, response)
	return
}

// DescribeNatGatewayDestinationIpPortTranslationNatRules
// 本接口（DescribeNatGatewayDestinationIpPortTranslationNatRules）用于查询NAT网关端口转发规则对象数组。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeNatGatewayDestinationIpPortTranslationNatRulesWithContext(ctx context.Context, request *DescribeNatGatewayDestinationIpPortTranslationNatRulesRequest) (response *DescribeNatGatewayDestinationIpPortTranslationNatRulesResponse, err error) {
	if request == nil {
		request = NewDescribeNatGatewayDestinationIpPortTranslationNatRulesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNatGatewayDestinationIpPortTranslationNatRulesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNatGatewayDirectConnectGatewayRouteRequest() (request *DescribeNatGatewayDirectConnectGatewayRouteRequest) {
	request = &DescribeNatGatewayDirectConnectGatewayRouteRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNatGatewayDirectConnectGatewayRoute")

	return
}

func NewDescribeNatGatewayDirectConnectGatewayRouteResponse() (response *DescribeNatGatewayDirectConnectGatewayRouteResponse) {
	response = &DescribeNatGatewayDirectConnectGatewayRouteResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNatGatewayDirectConnectGatewayRoute
// 查询专线绑定NAT的路由
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeNatGatewayDirectConnectGatewayRoute(request *DescribeNatGatewayDirectConnectGatewayRouteRequest) (response *DescribeNatGatewayDirectConnectGatewayRouteResponse, err error) {
	if request == nil {
		request = NewDescribeNatGatewayDirectConnectGatewayRouteRequest()
	}

	response = NewDescribeNatGatewayDirectConnectGatewayRouteResponse()
	err = c.Send(request, response)
	return
}

// DescribeNatGatewayDirectConnectGatewayRoute
// 查询专线绑定NAT的路由
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeNatGatewayDirectConnectGatewayRouteWithContext(ctx context.Context, request *DescribeNatGatewayDirectConnectGatewayRouteRequest) (response *DescribeNatGatewayDirectConnectGatewayRouteResponse, err error) {
	if request == nil {
		request = NewDescribeNatGatewayDirectConnectGatewayRouteRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNatGatewayDirectConnectGatewayRouteResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNatGatewaySourceIpTranslationNatRulesRequest() (request *DescribeNatGatewaySourceIpTranslationNatRulesRequest) {
	request = &DescribeNatGatewaySourceIpTranslationNatRulesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNatGatewaySourceIpTranslationNatRules")

	return
}

func NewDescribeNatGatewaySourceIpTranslationNatRulesResponse() (response *DescribeNatGatewaySourceIpTranslationNatRulesResponse) {
	response = &DescribeNatGatewaySourceIpTranslationNatRulesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNatGatewaySourceIpTranslationNatRules
// 本接口（DescribeNatGatewaySourceIpTranslationNatRules）用于查询NAT网关SNAT转发规则对象数组。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNatGatewaySourceIpTranslationNatRules(request *DescribeNatGatewaySourceIpTranslationNatRulesRequest) (response *DescribeNatGatewaySourceIpTranslationNatRulesResponse, err error) {
	if request == nil {
		request = NewDescribeNatGatewaySourceIpTranslationNatRulesRequest()
	}

	response = NewDescribeNatGatewaySourceIpTranslationNatRulesResponse()
	err = c.Send(request, response)
	return
}

// DescribeNatGatewaySourceIpTranslationNatRules
// 本接口（DescribeNatGatewaySourceIpTranslationNatRules）用于查询NAT网关SNAT转发规则对象数组。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNatGatewaySourceIpTranslationNatRulesWithContext(ctx context.Context, request *DescribeNatGatewaySourceIpTranslationNatRulesRequest) (response *DescribeNatGatewaySourceIpTranslationNatRulesResponse, err error) {
	if request == nil {
		request = NewDescribeNatGatewaySourceIpTranslationNatRulesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNatGatewaySourceIpTranslationNatRulesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNatGatewaysRequest() (request *DescribeNatGatewaysRequest) {
	request = &DescribeNatGatewaysRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNatGateways")

	return
}

func NewDescribeNatGatewaysResponse() (response *DescribeNatGatewaysResponse) {
	response = &DescribeNatGatewaysResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNatGateways
// 本接口（DescribeNatGateways）用于查询 NAT 网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
func (c *Client) DescribeNatGateways(request *DescribeNatGatewaysRequest) (response *DescribeNatGatewaysResponse, err error) {
	if request == nil {
		request = NewDescribeNatGatewaysRequest()
	}

	response = NewDescribeNatGatewaysResponse()
	err = c.Send(request, response)
	return
}

// DescribeNatGateways
// 本接口（DescribeNatGateways）用于查询 NAT 网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
func (c *Client) DescribeNatGatewaysWithContext(ctx context.Context, request *DescribeNatGatewaysRequest) (response *DescribeNatGatewaysResponse, err error) {
	if request == nil {
		request = NewDescribeNatGatewaysRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNatGatewaysResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNetDetectStatesRequest() (request *DescribeNetDetectStatesRequest) {
	request = &DescribeNetDetectStatesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNetDetectStates")

	return
}

func NewDescribeNetDetectStatesResponse() (response *DescribeNetDetectStatesResponse) {
	response = &DescribeNetDetectStatesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNetDetectStates
// 本接口(DescribeNetDetectStates)用于查询网络探测验证结果列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNetDetectStates(request *DescribeNetDetectStatesRequest) (response *DescribeNetDetectStatesResponse, err error) {
	if request == nil {
		request = NewDescribeNetDetectStatesRequest()
	}

	response = NewDescribeNetDetectStatesResponse()
	err = c.Send(request, response)
	return
}

// DescribeNetDetectStates
// 本接口(DescribeNetDetectStates)用于查询网络探测验证结果列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNetDetectStatesWithContext(ctx context.Context, request *DescribeNetDetectStatesRequest) (response *DescribeNetDetectStatesResponse, err error) {
	if request == nil {
		request = NewDescribeNetDetectStatesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNetDetectStatesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNetDetectsRequest() (request *DescribeNetDetectsRequest) {
	request = &DescribeNetDetectsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNetDetects")

	return
}

func NewDescribeNetDetectsResponse() (response *DescribeNetDetectsResponse) {
	response = &DescribeNetDetectsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNetDetects
// 本接口（DescribeNetDetects）用于查询网络探测列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNetDetects(request *DescribeNetDetectsRequest) (response *DescribeNetDetectsResponse, err error) {
	if request == nil {
		request = NewDescribeNetDetectsRequest()
	}

	response = NewDescribeNetDetectsResponse()
	err = c.Send(request, response)
	return
}

// DescribeNetDetects
// 本接口（DescribeNetDetects）用于查询网络探测列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNetDetectsWithContext(ctx context.Context, request *DescribeNetDetectsRequest) (response *DescribeNetDetectsResponse, err error) {
	if request == nil {
		request = NewDescribeNetDetectsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNetDetectsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNetworkAclsRequest() (request *DescribeNetworkAclsRequest) {
	request = &DescribeNetworkAclsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNetworkAcls")

	return
}

func NewDescribeNetworkAclsResponse() (response *DescribeNetworkAclsResponse) {
	response = &DescribeNetworkAclsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNetworkAcls
// 本接口（DescribeNetworkAcls）用于查询网络ACL列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
func (c *Client) DescribeNetworkAcls(request *DescribeNetworkAclsRequest) (response *DescribeNetworkAclsResponse, err error) {
	if request == nil {
		request = NewDescribeNetworkAclsRequest()
	}

	response = NewDescribeNetworkAclsResponse()
	err = c.Send(request, response)
	return
}

// DescribeNetworkAcls
// 本接口（DescribeNetworkAcls）用于查询网络ACL列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
func (c *Client) DescribeNetworkAclsWithContext(ctx context.Context, request *DescribeNetworkAclsRequest) (response *DescribeNetworkAclsResponse, err error) {
	if request == nil {
		request = NewDescribeNetworkAclsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNetworkAclsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNetworkInterfaceLimitRequest() (request *DescribeNetworkInterfaceLimitRequest) {
	request = &DescribeNetworkInterfaceLimitRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNetworkInterfaceLimit")

	return
}

func NewDescribeNetworkInterfaceLimitResponse() (response *DescribeNetworkInterfaceLimitResponse) {
	response = &DescribeNetworkInterfaceLimitResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNetworkInterfaceLimit
// 本接口（DescribeNetworkInterfaceLimit）根据CVM实例ID或弹性网卡ID查询弹性网卡配额，返回该CVM实例或弹性网卡能绑定的弹性网卡配额，以及弹性网卡可以分配的IP配额
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNetworkInterfaceLimit(request *DescribeNetworkInterfaceLimitRequest) (response *DescribeNetworkInterfaceLimitResponse, err error) {
	if request == nil {
		request = NewDescribeNetworkInterfaceLimitRequest()
	}

	response = NewDescribeNetworkInterfaceLimitResponse()
	err = c.Send(request, response)
	return
}

// DescribeNetworkInterfaceLimit
// 本接口（DescribeNetworkInterfaceLimit）根据CVM实例ID或弹性网卡ID查询弹性网卡配额，返回该CVM实例或弹性网卡能绑定的弹性网卡配额，以及弹性网卡可以分配的IP配额
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNetworkInterfaceLimitWithContext(ctx context.Context, request *DescribeNetworkInterfaceLimitRequest) (response *DescribeNetworkInterfaceLimitResponse, err error) {
	if request == nil {
		request = NewDescribeNetworkInterfaceLimitRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNetworkInterfaceLimitResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeNetworkInterfacesRequest() (request *DescribeNetworkInterfacesRequest) {
	request = &DescribeNetworkInterfacesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeNetworkInterfaces")

	return
}

func NewDescribeNetworkInterfacesResponse() (response *DescribeNetworkInterfacesResponse) {
	response = &DescribeNetworkInterfacesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeNetworkInterfaces
// 本接口（DescribeNetworkInterfaces）用于查询弹性网卡列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNetworkInterfaces(request *DescribeNetworkInterfacesRequest) (response *DescribeNetworkInterfacesResponse, err error) {
	if request == nil {
		request = NewDescribeNetworkInterfacesRequest()
	}

	response = NewDescribeNetworkInterfacesResponse()
	err = c.Send(request, response)
	return
}

// DescribeNetworkInterfaces
// 本接口（DescribeNetworkInterfaces）用于查询弹性网卡列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeNetworkInterfacesWithContext(ctx context.Context, request *DescribeNetworkInterfacesRequest) (response *DescribeNetworkInterfacesResponse, err error) {
	if request == nil {
		request = NewDescribeNetworkInterfacesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeNetworkInterfacesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeProductQuotaRequest() (request *DescribeProductQuotaRequest) {
	request = &DescribeProductQuotaRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeProductQuota")

	return
}

func NewDescribeProductQuotaResponse() (response *DescribeProductQuotaResponse) {
	response = &DescribeProductQuotaResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeProductQuota
// 本接口用于查询网络产品的配额信息
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DescribeProductQuota(request *DescribeProductQuotaRequest) (response *DescribeProductQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeProductQuotaRequest()
	}

	response = NewDescribeProductQuotaResponse()
	err = c.Send(request, response)
	return
}

// DescribeProductQuota
// 本接口用于查询网络产品的配额信息
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DescribeProductQuotaWithContext(ctx context.Context, request *DescribeProductQuotaRequest) (response *DescribeProductQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeProductQuotaRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeProductQuotaResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeRouteConflictsRequest() (request *DescribeRouteConflictsRequest) {
	request = &DescribeRouteConflictsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeRouteConflicts")

	return
}

func NewDescribeRouteConflictsResponse() (response *DescribeRouteConflictsResponse) {
	response = &DescribeRouteConflictsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeRouteConflicts
// 本接口（DescribeRouteConflicts）用于查询自定义路由策略与云联网路由策略冲突列表
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeRouteConflicts(request *DescribeRouteConflictsRequest) (response *DescribeRouteConflictsResponse, err error) {
	if request == nil {
		request = NewDescribeRouteConflictsRequest()
	}

	response = NewDescribeRouteConflictsResponse()
	err = c.Send(request, response)
	return
}

// DescribeRouteConflicts
// 本接口（DescribeRouteConflicts）用于查询自定义路由策略与云联网路由策略冲突列表
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeRouteConflictsWithContext(ctx context.Context, request *DescribeRouteConflictsRequest) (response *DescribeRouteConflictsResponse, err error) {
	if request == nil {
		request = NewDescribeRouteConflictsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeRouteConflictsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeRouteTablesRequest() (request *DescribeRouteTablesRequest) {
	request = &DescribeRouteTablesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeRouteTables")

	return
}

func NewDescribeRouteTablesResponse() (response *DescribeRouteTablesResponse) {
	response = &DescribeRouteTablesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeRouteTables
//
//	本接口（DescribeRouteTables）用于查询路由表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeRouteTables(request *DescribeRouteTablesRequest) (response *DescribeRouteTablesResponse, err error) {
	if request == nil {
		request = NewDescribeRouteTablesRequest()
	}

	response = NewDescribeRouteTablesResponse()
	err = c.Send(request, response)
	return
}

// DescribeRouteTables
//
//	本接口（DescribeRouteTables）用于查询路由表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeRouteTablesWithContext(ctx context.Context, request *DescribeRouteTablesRequest) (response *DescribeRouteTablesResponse, err error) {
	if request == nil {
		request = NewDescribeRouteTablesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeRouteTablesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeSecurityGroupAssociationStatisticsRequest() (request *DescribeSecurityGroupAssociationStatisticsRequest) {
	request = &DescribeSecurityGroupAssociationStatisticsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeSecurityGroupAssociationStatistics")

	return
}

func NewDescribeSecurityGroupAssociationStatisticsResponse() (response *DescribeSecurityGroupAssociationStatisticsResponse) {
	response = &DescribeSecurityGroupAssociationStatisticsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeSecurityGroupAssociationStatistics
// 本接口（DescribeSecurityGroupAssociationStatistics）用于查询安全组关联的实例统计。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupAssociationStatistics(request *DescribeSecurityGroupAssociationStatisticsRequest) (response *DescribeSecurityGroupAssociationStatisticsResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupAssociationStatisticsRequest()
	}

	response = NewDescribeSecurityGroupAssociationStatisticsResponse()
	err = c.Send(request, response)
	return
}

// DescribeSecurityGroupAssociationStatistics
// 本接口（DescribeSecurityGroupAssociationStatistics）用于查询安全组关联的实例统计。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupAssociationStatisticsWithContext(ctx context.Context, request *DescribeSecurityGroupAssociationStatisticsRequest) (response *DescribeSecurityGroupAssociationStatisticsResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupAssociationStatisticsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeSecurityGroupAssociationStatisticsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeSecurityGroupLimitsRequest() (request *DescribeSecurityGroupLimitsRequest) {
	request = &DescribeSecurityGroupLimitsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeSecurityGroupLimits")

	return
}

func NewDescribeSecurityGroupLimitsResponse() (response *DescribeSecurityGroupLimitsResponse) {
	response = &DescribeSecurityGroupLimitsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeSecurityGroupLimits
// 本接口(DescribeSecurityGroupLimits)用于查询用户安全组配额。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupLimits(request *DescribeSecurityGroupLimitsRequest) (response *DescribeSecurityGroupLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupLimitsRequest()
	}

	response = NewDescribeSecurityGroupLimitsResponse()
	err = c.Send(request, response)
	return
}

// DescribeSecurityGroupLimits
// 本接口(DescribeSecurityGroupLimits)用于查询用户安全组配额。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupLimitsWithContext(ctx context.Context, request *DescribeSecurityGroupLimitsRequest) (response *DescribeSecurityGroupLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupLimitsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeSecurityGroupLimitsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeSecurityGroupPoliciesRequest() (request *DescribeSecurityGroupPoliciesRequest) {
	request = &DescribeSecurityGroupPoliciesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeSecurityGroupPolicies")

	return
}

func NewDescribeSecurityGroupPoliciesResponse() (response *DescribeSecurityGroupPoliciesResponse) {
	response = &DescribeSecurityGroupPoliciesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeSecurityGroupPolicies
// 本接口（DescribeSecurityGroupPolicies）用于查询安全组规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupPolicies(request *DescribeSecurityGroupPoliciesRequest) (response *DescribeSecurityGroupPoliciesResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupPoliciesRequest()
	}

	response = NewDescribeSecurityGroupPoliciesResponse()
	err = c.Send(request, response)
	return
}

// DescribeSecurityGroupPolicies
// 本接口（DescribeSecurityGroupPolicies）用于查询安全组规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupPoliciesWithContext(ctx context.Context, request *DescribeSecurityGroupPoliciesRequest) (response *DescribeSecurityGroupPoliciesResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupPoliciesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeSecurityGroupPoliciesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeSecurityGroupReferencesRequest() (request *DescribeSecurityGroupReferencesRequest) {
	request = &DescribeSecurityGroupReferencesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeSecurityGroupReferences")

	return
}

func NewDescribeSecurityGroupReferencesResponse() (response *DescribeSecurityGroupReferencesResponse) {
	response = &DescribeSecurityGroupReferencesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeSecurityGroupReferences
// 本接口（DescribeSecurityGroupReferences）用于查询安全组被引用信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupReferences(request *DescribeSecurityGroupReferencesRequest) (response *DescribeSecurityGroupReferencesResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupReferencesRequest()
	}

	response = NewDescribeSecurityGroupReferencesResponse()
	err = c.Send(request, response)
	return
}

// DescribeSecurityGroupReferences
// 本接口（DescribeSecurityGroupReferences）用于查询安全组被引用信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupReferencesWithContext(ctx context.Context, request *DescribeSecurityGroupReferencesRequest) (response *DescribeSecurityGroupReferencesResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupReferencesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeSecurityGroupReferencesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeSecurityGroupsRequest() (request *DescribeSecurityGroupsRequest) {
	request = &DescribeSecurityGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeSecurityGroups")

	return
}

func NewDescribeSecurityGroupsResponse() (response *DescribeSecurityGroupsResponse) {
	response = &DescribeSecurityGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeSecurityGroups
// 本接口（DescribeSecurityGroups）用于查询安全组。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroups(request *DescribeSecurityGroupsRequest) (response *DescribeSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupsRequest()
	}

	response = NewDescribeSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

// DescribeSecurityGroups
// 本接口（DescribeSecurityGroups）用于查询安全组。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSecurityGroupsWithContext(ctx context.Context, request *DescribeSecurityGroupsRequest) (response *DescribeSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeSecurityGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeServiceTemplateGroupsRequest() (request *DescribeServiceTemplateGroupsRequest) {
	request = &DescribeServiceTemplateGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeServiceTemplateGroups")

	return
}

func NewDescribeServiceTemplateGroupsResponse() (response *DescribeServiceTemplateGroupsResponse) {
	response = &DescribeServiceTemplateGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeServiceTemplateGroups
// 本接口（DescribeServiceTemplateGroups）用于查询协议端口模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeServiceTemplateGroups(request *DescribeServiceTemplateGroupsRequest) (response *DescribeServiceTemplateGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeServiceTemplateGroupsRequest()
	}

	response = NewDescribeServiceTemplateGroupsResponse()
	err = c.Send(request, response)
	return
}

// DescribeServiceTemplateGroups
// 本接口（DescribeServiceTemplateGroups）用于查询协议端口模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeServiceTemplateGroupsWithContext(ctx context.Context, request *DescribeServiceTemplateGroupsRequest) (response *DescribeServiceTemplateGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeServiceTemplateGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeServiceTemplateGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeServiceTemplatesRequest() (request *DescribeServiceTemplatesRequest) {
	request = &DescribeServiceTemplatesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeServiceTemplates")

	return
}

func NewDescribeServiceTemplatesResponse() (response *DescribeServiceTemplatesResponse) {
	response = &DescribeServiceTemplatesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeServiceTemplates
// 本接口（DescribeServiceTemplates）用于查询协议端口模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeServiceTemplates(request *DescribeServiceTemplatesRequest) (response *DescribeServiceTemplatesResponse, err error) {
	if request == nil {
		request = NewDescribeServiceTemplatesRequest()
	}

	response = NewDescribeServiceTemplatesResponse()
	err = c.Send(request, response)
	return
}

// DescribeServiceTemplates
// 本接口（DescribeServiceTemplates）用于查询协议端口模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) DescribeServiceTemplatesWithContext(ctx context.Context, request *DescribeServiceTemplatesRequest) (response *DescribeServiceTemplatesResponse, err error) {
	if request == nil {
		request = NewDescribeServiceTemplatesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeServiceTemplatesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeSubnetsRequest() (request *DescribeSubnetsRequest) {
	request = &DescribeSubnetsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeSubnets")

	return
}

func NewDescribeSubnetsResponse() (response *DescribeSubnetsResponse) {
	response = &DescribeSubnetsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeSubnets
// 本接口（DescribeSubnets）用于查询子网列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSubnets(request *DescribeSubnetsRequest) (response *DescribeSubnetsResponse, err error) {
	if request == nil {
		request = NewDescribeSubnetsRequest()
	}

	response = NewDescribeSubnetsResponse()
	err = c.Send(request, response)
	return
}

// DescribeSubnets
// 本接口（DescribeSubnets）用于查询子网列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_EMPTY = "InvalidParameterValue.Empty"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeSubnetsWithContext(ctx context.Context, request *DescribeSubnetsRequest) (response *DescribeSubnetsResponse, err error) {
	if request == nil {
		request = NewDescribeSubnetsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeSubnetsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeTaskResultRequest() (request *DescribeTaskResultRequest) {
	request = &DescribeTaskResultRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeTaskResult")

	return
}

func NewDescribeTaskResultResponse() (response *DescribeTaskResultResponse) {
	response = &DescribeTaskResultResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeTaskResult
// 查询EIP异步任务执行结果
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeTaskResult(request *DescribeTaskResultRequest) (response *DescribeTaskResultResponse, err error) {
	if request == nil {
		request = NewDescribeTaskResultRequest()
	}

	response = NewDescribeTaskResultResponse()
	err = c.Send(request, response)
	return
}

// DescribeTaskResult
// 查询EIP异步任务执行结果
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeTaskResultWithContext(ctx context.Context, request *DescribeTaskResultRequest) (response *DescribeTaskResultResponse, err error) {
	if request == nil {
		request = NewDescribeTaskResultRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeTaskResultResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeTemplateLimitsRequest() (request *DescribeTemplateLimitsRequest) {
	request = &DescribeTemplateLimitsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeTemplateLimits")

	return
}

func NewDescribeTemplateLimitsResponse() (response *DescribeTemplateLimitsResponse) {
	response = &DescribeTemplateLimitsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeTemplateLimits
// 本接口（DescribeTemplateLimits）用于查询参数模板配额列表。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeTemplateLimits(request *DescribeTemplateLimitsRequest) (response *DescribeTemplateLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeTemplateLimitsRequest()
	}

	response = NewDescribeTemplateLimitsResponse()
	err = c.Send(request, response)
	return
}

// DescribeTemplateLimits
// 本接口（DescribeTemplateLimits）用于查询参数模板配额列表。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeTemplateLimitsWithContext(ctx context.Context, request *DescribeTemplateLimitsRequest) (response *DescribeTemplateLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeTemplateLimitsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeTemplateLimitsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcEndPointRequest() (request *DescribeVpcEndPointRequest) {
	request = &DescribeVpcEndPointRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcEndPoint")

	return
}

func NewDescribeVpcEndPointResponse() (response *DescribeVpcEndPointResponse) {
	response = &DescribeVpcEndPointResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcEndPoint
// 查询终端节点列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCENOTFOUND_SVCNOTEXIST = "ResourceNotFound.SvcNotExist"
func (c *Client) DescribeVpcEndPoint(request *DescribeVpcEndPointRequest) (response *DescribeVpcEndPointResponse, err error) {
	if request == nil {
		request = NewDescribeVpcEndPointRequest()
	}

	response = NewDescribeVpcEndPointResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcEndPoint
// 查询终端节点列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCENOTFOUND_SVCNOTEXIST = "ResourceNotFound.SvcNotExist"
func (c *Client) DescribeVpcEndPointWithContext(ctx context.Context, request *DescribeVpcEndPointRequest) (response *DescribeVpcEndPointResponse, err error) {
	if request == nil {
		request = NewDescribeVpcEndPointRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcEndPointResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcEndPointServiceRequest() (request *DescribeVpcEndPointServiceRequest) {
	request = &DescribeVpcEndPointServiceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcEndPointService")

	return
}

func NewDescribeVpcEndPointServiceResponse() (response *DescribeVpcEndPointServiceResponse) {
	response = &DescribeVpcEndPointServiceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcEndPointService
// 查询终端节点服务列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcEndPointService(request *DescribeVpcEndPointServiceRequest) (response *DescribeVpcEndPointServiceResponse, err error) {
	if request == nil {
		request = NewDescribeVpcEndPointServiceRequest()
	}

	response = NewDescribeVpcEndPointServiceResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcEndPointService
// 查询终端节点服务列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcEndPointServiceWithContext(ctx context.Context, request *DescribeVpcEndPointServiceRequest) (response *DescribeVpcEndPointServiceResponse, err error) {
	if request == nil {
		request = NewDescribeVpcEndPointServiceRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcEndPointServiceResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcEndPointServiceWhiteListRequest() (request *DescribeVpcEndPointServiceWhiteListRequest) {
	request = &DescribeVpcEndPointServiceWhiteListRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcEndPointServiceWhiteList")

	return
}

func NewDescribeVpcEndPointServiceWhiteListResponse() (response *DescribeVpcEndPointServiceWhiteListResponse) {
	response = &DescribeVpcEndPointServiceWhiteListResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcEndPointServiceWhiteList
// 查询终端节点服务的服务白名单列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
func (c *Client) DescribeVpcEndPointServiceWhiteList(request *DescribeVpcEndPointServiceWhiteListRequest) (response *DescribeVpcEndPointServiceWhiteListResponse, err error) {
	if request == nil {
		request = NewDescribeVpcEndPointServiceWhiteListRequest()
	}

	response = NewDescribeVpcEndPointServiceWhiteListResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcEndPointServiceWhiteList
// 查询终端节点服务的服务白名单列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
func (c *Client) DescribeVpcEndPointServiceWhiteListWithContext(ctx context.Context, request *DescribeVpcEndPointServiceWhiteListRequest) (response *DescribeVpcEndPointServiceWhiteListResponse, err error) {
	if request == nil {
		request = NewDescribeVpcEndPointServiceWhiteListRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcEndPointServiceWhiteListResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcInstancesRequest() (request *DescribeVpcInstancesRequest) {
	request = &DescribeVpcInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcInstances")

	return
}

func NewDescribeVpcInstancesResponse() (response *DescribeVpcInstancesResponse) {
	response = &DescribeVpcInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcInstances
//
//	本接口（DescribeVpcInstances）用于查询VPC下的云主机实例列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcInstances(request *DescribeVpcInstancesRequest) (response *DescribeVpcInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeVpcInstancesRequest()
	}

	response = NewDescribeVpcInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcInstances
//
//	本接口（DescribeVpcInstances）用于查询VPC下的云主机实例列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcInstancesWithContext(ctx context.Context, request *DescribeVpcInstancesRequest) (response *DescribeVpcInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeVpcInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcIpv6AddressesRequest() (request *DescribeVpcIpv6AddressesRequest) {
	request = &DescribeVpcIpv6AddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcIpv6Addresses")

	return
}

func NewDescribeVpcIpv6AddressesResponse() (response *DescribeVpcIpv6AddressesResponse) {
	response = &DescribeVpcIpv6AddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcIpv6Addresses
// 本接口（DescribeVpcIpv6Addresses）用于查询 `VPC` `IPv6` 信息。
//
// 只能查询已使用的`IPv6`信息，当查询未使用的IP时，本接口不会报错，但不会出现在返回结果里。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcIpv6Addresses(request *DescribeVpcIpv6AddressesRequest) (response *DescribeVpcIpv6AddressesResponse, err error) {
	if request == nil {
		request = NewDescribeVpcIpv6AddressesRequest()
	}

	response = NewDescribeVpcIpv6AddressesResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcIpv6Addresses
// 本接口（DescribeVpcIpv6Addresses）用于查询 `VPC` `IPv6` 信息。
//
// 只能查询已使用的`IPv6`信息，当查询未使用的IP时，本接口不会报错，但不会出现在返回结果里。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcIpv6AddressesWithContext(ctx context.Context, request *DescribeVpcIpv6AddressesRequest) (response *DescribeVpcIpv6AddressesResponse, err error) {
	if request == nil {
		request = NewDescribeVpcIpv6AddressesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcIpv6AddressesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcLimitsRequest() (request *DescribeVpcLimitsRequest) {
	request = &DescribeVpcLimitsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcLimits")

	return
}

func NewDescribeVpcLimitsResponse() (response *DescribeVpcLimitsResponse) {
	response = &DescribeVpcLimitsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcLimits
// 获取私有网络配额，部分私有网络的配额有地域属性。
//
// LimitTypes取值范围：
//
// * appid-max-vpcs （每个开发商每个地域可创建的VPC数）
//
// * vpc-max-subnets（每个VPC可创建的子网数）
//
// * vpc-max-route-tables（每个VPC可创建的路由表数）
//
// * route-table-max-policies（每个路由表可添加的策略数）
//
// * vpc-max-vpn-gateways（每个VPC可创建的VPN网关数）
//
// * appid-max-custom-gateways（每个开发商可创建的对端网关数）
//
// * appid-max-vpn-connections（每个开发商可创建的VPN通道数）
//
// * custom-gateway-max-vpn-connections（每个对端网关可创建的VPN通道数）
//
// * vpn-gateway-max-custom-gateways（每个VPNGW可以创建的通道数）
//
// * vpc-max-network-acls（每个VPC可创建的网络ACL数）
//
// * network-acl-max-inbound-policies（每个网络ACL可添加的入站规则数）
//
// * network-acl-max-outbound-policies（每个网络ACL可添加的出站规则数）
//
// * vpc-max-vpcpeers（每个VPC可创建的对等连接数）
//
// * vpc-max-available-vpcpeers（每个VPC可创建的有效对等连接数）
//
// * vpc-max-basic-network-interconnections（每个VPC可创建的基础网络云主机与VPC互通数）
//
// * direct-connection-max-snats（每个专线网关可创建的SNAT数）
//
// * direct-connection-max-dnats（每个专线网关可创建的DNAT数）
//
// * direct-connection-max-snapts（每个专线网关可创建的SNAPT数）
//
// * direct-connection-max-dnapts（每个专线网关可创建的DNAPT数）
//
// * vpc-max-nat-gateways（每个VPC可创建的NAT网关数）
//
// * nat-gateway-max-eips（每个NAT可以购买的外网IP数量）
//
// * vpc-max-enis（每个VPC可创建弹性网卡数）
//
// * vpc-max-havips（每个VPC可创建HAVIP数）
//
// * eni-max-private-ips（每个ENI可以绑定的内网IP数（ENI未绑定子机））
//
// * nat-gateway-max-dnapts（每个NAT网关可创建的DNAPT数）
//
// * vpc-max-ipv6s（每个VPC可分配的IPv6地址数）
//
// * eni-max-ipv6s（每个ENI可分配的IPv6地址数）
//
// * vpc-max-assistant_cidrs（每个VPC可分配的辅助CIDR数）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DescribeVpcLimits(request *DescribeVpcLimitsRequest) (response *DescribeVpcLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeVpcLimitsRequest()
	}

	response = NewDescribeVpcLimitsResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcLimits
// 获取私有网络配额，部分私有网络的配额有地域属性。
//
// LimitTypes取值范围：
//
// * appid-max-vpcs （每个开发商每个地域可创建的VPC数）
//
// * vpc-max-subnets（每个VPC可创建的子网数）
//
// * vpc-max-route-tables（每个VPC可创建的路由表数）
//
// * route-table-max-policies（每个路由表可添加的策略数）
//
// * vpc-max-vpn-gateways（每个VPC可创建的VPN网关数）
//
// * appid-max-custom-gateways（每个开发商可创建的对端网关数）
//
// * appid-max-vpn-connections（每个开发商可创建的VPN通道数）
//
// * custom-gateway-max-vpn-connections（每个对端网关可创建的VPN通道数）
//
// * vpn-gateway-max-custom-gateways（每个VPNGW可以创建的通道数）
//
// * vpc-max-network-acls（每个VPC可创建的网络ACL数）
//
// * network-acl-max-inbound-policies（每个网络ACL可添加的入站规则数）
//
// * network-acl-max-outbound-policies（每个网络ACL可添加的出站规则数）
//
// * vpc-max-vpcpeers（每个VPC可创建的对等连接数）
//
// * vpc-max-available-vpcpeers（每个VPC可创建的有效对等连接数）
//
// * vpc-max-basic-network-interconnections（每个VPC可创建的基础网络云主机与VPC互通数）
//
// * direct-connection-max-snats（每个专线网关可创建的SNAT数）
//
// * direct-connection-max-dnats（每个专线网关可创建的DNAT数）
//
// * direct-connection-max-snapts（每个专线网关可创建的SNAPT数）
//
// * direct-connection-max-dnapts（每个专线网关可创建的DNAPT数）
//
// * vpc-max-nat-gateways（每个VPC可创建的NAT网关数）
//
// * nat-gateway-max-eips（每个NAT可以购买的外网IP数量）
//
// * vpc-max-enis（每个VPC可创建弹性网卡数）
//
// * vpc-max-havips（每个VPC可创建HAVIP数）
//
// * eni-max-private-ips（每个ENI可以绑定的内网IP数（ENI未绑定子机））
//
// * nat-gateway-max-dnapts（每个NAT网关可创建的DNAPT数）
//
// * vpc-max-ipv6s（每个VPC可分配的IPv6地址数）
//
// * eni-max-ipv6s（每个ENI可分配的IPv6地址数）
//
// * vpc-max-assistant_cidrs（每个VPC可分配的辅助CIDR数）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DescribeVpcLimitsWithContext(ctx context.Context, request *DescribeVpcLimitsRequest) (response *DescribeVpcLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeVpcLimitsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcLimitsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcPrivateIpAddressesRequest() (request *DescribeVpcPrivateIpAddressesRequest) {
	request = &DescribeVpcPrivateIpAddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcPrivateIpAddresses")

	return
}

func NewDescribeVpcPrivateIpAddressesResponse() (response *DescribeVpcPrivateIpAddressesResponse) {
	response = &DescribeVpcPrivateIpAddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcPrivateIpAddresses
// 本接口（DescribeVpcPrivateIpAddresses）用于查询VPC内网IP信息。<br />
//
// 只能查询已使用的IP信息，当查询未使用的IP时，本接口不会报错，但不会出现在返回结果里。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcPrivateIpAddresses(request *DescribeVpcPrivateIpAddressesRequest) (response *DescribeVpcPrivateIpAddressesResponse, err error) {
	if request == nil {
		request = NewDescribeVpcPrivateIpAddressesRequest()
	}

	response = NewDescribeVpcPrivateIpAddressesResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcPrivateIpAddresses
// 本接口（DescribeVpcPrivateIpAddresses）用于查询VPC内网IP信息。<br />
//
// 只能查询已使用的IP信息，当查询未使用的IP时，本接口不会报错，但不会出现在返回结果里。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcPrivateIpAddressesWithContext(ctx context.Context, request *DescribeVpcPrivateIpAddressesRequest) (response *DescribeVpcPrivateIpAddressesResponse, err error) {
	if request == nil {
		request = NewDescribeVpcPrivateIpAddressesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcPrivateIpAddressesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcResourceDashboardRequest() (request *DescribeVpcResourceDashboardRequest) {
	request = &DescribeVpcResourceDashboardRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcResourceDashboard")

	return
}

func NewDescribeVpcResourceDashboardResponse() (response *DescribeVpcResourceDashboardResponse) {
	response = &DescribeVpcResourceDashboardResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcResourceDashboard
// 本接口(DescribeVpcResourceDashboard)用于查看VPC资源信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcResourceDashboard(request *DescribeVpcResourceDashboardRequest) (response *DescribeVpcResourceDashboardResponse, err error) {
	if request == nil {
		request = NewDescribeVpcResourceDashboardRequest()
	}

	response = NewDescribeVpcResourceDashboardResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcResourceDashboard
// 本接口(DescribeVpcResourceDashboard)用于查看VPC资源信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcResourceDashboardWithContext(ctx context.Context, request *DescribeVpcResourceDashboardRequest) (response *DescribeVpcResourceDashboardResponse, err error) {
	if request == nil {
		request = NewDescribeVpcResourceDashboardRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcResourceDashboardResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcTaskResultRequest() (request *DescribeVpcTaskResultRequest) {
	request = &DescribeVpcTaskResultRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcTaskResult")

	return
}

func NewDescribeVpcTaskResultResponse() (response *DescribeVpcTaskResultResponse) {
	response = &DescribeVpcTaskResultResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcTaskResult
// 本接口（DescribeVpcTaskResult）用于查询VPC任务执行结果。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeVpcTaskResult(request *DescribeVpcTaskResultRequest) (response *DescribeVpcTaskResultResponse, err error) {
	if request == nil {
		request = NewDescribeVpcTaskResultRequest()
	}

	response = NewDescribeVpcTaskResultResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcTaskResult
// 本接口（DescribeVpcTaskResult）用于查询VPC任务执行结果。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeVpcTaskResultWithContext(ctx context.Context, request *DescribeVpcTaskResultRequest) (response *DescribeVpcTaskResultResponse, err error) {
	if request == nil {
		request = NewDescribeVpcTaskResultRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcTaskResultResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcsRequest() (request *DescribeVpcsRequest) {
	request = &DescribeVpcsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpcs")

	return
}

func NewDescribeVpcsResponse() (response *DescribeVpcsResponse) {
	response = &DescribeVpcsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcs
// 本接口（DescribeVpcs）用于查询私有网络列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcs(request *DescribeVpcsRequest) (response *DescribeVpcsResponse, err error) {
	if request == nil {
		request = NewDescribeVpcsRequest()
	}

	response = NewDescribeVpcsResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcs
// 本接口（DescribeVpcs）用于查询私有网络列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpcsWithContext(ctx context.Context, request *DescribeVpcsRequest) (response *DescribeVpcsResponse, err error) {
	if request == nil {
		request = NewDescribeVpcsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpnConnectionsRequest() (request *DescribeVpnConnectionsRequest) {
	request = &DescribeVpnConnectionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpnConnections")

	return
}

func NewDescribeVpnConnectionsResponse() (response *DescribeVpnConnectionsResponse) {
	response = &DescribeVpnConnectionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpnConnections
//
//	本接口（DescribeVpnConnections）查询VPN通道列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpnConnections(request *DescribeVpnConnectionsRequest) (response *DescribeVpnConnectionsResponse, err error) {
	if request == nil {
		request = NewDescribeVpnConnectionsRequest()
	}

	response = NewDescribeVpnConnectionsResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpnConnections
//
//	本接口（DescribeVpnConnections）查询VPN通道列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpnConnectionsWithContext(ctx context.Context, request *DescribeVpnConnectionsRequest) (response *DescribeVpnConnectionsResponse, err error) {
	if request == nil {
		request = NewDescribeVpnConnectionsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpnConnectionsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpnGatewayCcnRoutesRequest() (request *DescribeVpnGatewayCcnRoutesRequest) {
	request = &DescribeVpnGatewayCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpnGatewayCcnRoutes")

	return
}

func NewDescribeVpnGatewayCcnRoutesResponse() (response *DescribeVpnGatewayCcnRoutesResponse) {
	response = &DescribeVpnGatewayCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpnGatewayCcnRoutes
// 本接口（DescribeVpnGatewayCcnRoutes）用于查询VPN网关云联网路由
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpnGatewayCcnRoutes(request *DescribeVpnGatewayCcnRoutesRequest) (response *DescribeVpnGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeVpnGatewayCcnRoutesRequest()
	}

	response = NewDescribeVpnGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpnGatewayCcnRoutes
// 本接口（DescribeVpnGatewayCcnRoutes）用于查询VPN网关云联网路由
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpnGatewayCcnRoutesWithContext(ctx context.Context, request *DescribeVpnGatewayCcnRoutesRequest) (response *DescribeVpnGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeVpnGatewayCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpnGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpnGatewayRoutesRequest() (request *DescribeVpnGatewayRoutesRequest) {
	request = &DescribeVpnGatewayRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpnGatewayRoutes")

	return
}

func NewDescribeVpnGatewayRoutesResponse() (response *DescribeVpnGatewayRoutesResponse) {
	response = &DescribeVpnGatewayRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpnGatewayRoutes
// 查询路由型VPN网关的目的路由
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeVpnGatewayRoutes(request *DescribeVpnGatewayRoutesRequest) (response *DescribeVpnGatewayRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeVpnGatewayRoutesRequest()
	}

	response = NewDescribeVpnGatewayRoutesResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpnGatewayRoutes
// 查询路由型VPN网关的目的路由
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeVpnGatewayRoutesWithContext(ctx context.Context, request *DescribeVpnGatewayRoutesRequest) (response *DescribeVpnGatewayRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeVpnGatewayRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpnGatewayRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpnGatewaysRequest() (request *DescribeVpnGatewaysRequest) {
	request = &DescribeVpnGatewaysRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DescribeVpnGateways")

	return
}

func NewDescribeVpnGatewaysResponse() (response *DescribeVpnGatewaysResponse) {
	response = &DescribeVpnGatewaysResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpnGateways
// 本接口（DescribeVpnGateways）用于查询VPN网关列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDVPNGATEWAYID_MALFORMED = "InvalidVpnGatewayId.Malformed"
//	INVALIDVPNGATEWAYID_NOTFOUND = "InvalidVpnGatewayId.NotFound"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpnGateways(request *DescribeVpnGatewaysRequest) (response *DescribeVpnGatewaysResponse, err error) {
	if request == nil {
		request = NewDescribeVpnGatewaysRequest()
	}

	response = NewDescribeVpnGatewaysResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpnGateways
// 本接口（DescribeVpnGateways）用于查询VPN网关列表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_FILTERINVALIDKEY = "InvalidParameter.FilterInvalidKey"
//	INVALIDPARAMETER_FILTERNOTDICT = "InvalidParameter.FilterNotDict"
//	INVALIDPARAMETER_FILTERVALUESNOTLIST = "InvalidParameter.FilterValuesNotList"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDVPNGATEWAYID_MALFORMED = "InvalidVpnGatewayId.Malformed"
//	INVALIDVPNGATEWAYID_NOTFOUND = "InvalidVpnGatewayId.NotFound"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeVpnGatewaysWithContext(ctx context.Context, request *DescribeVpnGatewaysRequest) (response *DescribeVpnGatewaysResponse, err error) {
	if request == nil {
		request = NewDescribeVpnGatewaysRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpnGatewaysResponse()
	err = c.Send(request, response)
	return
}

func NewDetachCcnInstancesRequest() (request *DetachCcnInstancesRequest) {
	request = &DetachCcnInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DetachCcnInstances")

	return
}

func NewDetachCcnInstancesResponse() (response *DetachCcnInstancesResponse) {
	response = &DetachCcnInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DetachCcnInstances
// 本接口（DetachCcnInstances）用于从云联网实例中解关联指定的网络实例。<br />
//
// 解关联网络实例后，相应的路由策略会一并删除。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_APPIDMISMATCH = "UnsupportedOperation.AppIdMismatch"
//	UNSUPPORTEDOPERATION_APPIDNOTFOUND = "UnsupportedOperation.AppIdNotFound"
//	UNSUPPORTEDOPERATION_CCNNOTATTACHED = "UnsupportedOperation.CcnNotAttached"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DetachCcnInstances(request *DetachCcnInstancesRequest) (response *DetachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewDetachCcnInstancesRequest()
	}

	response = NewDetachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

// DetachCcnInstances
// 本接口（DetachCcnInstances）用于从云联网实例中解关联指定的网络实例。<br />
//
// 解关联网络实例后，相应的路由策略会一并删除。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_APPIDMISMATCH = "UnsupportedOperation.AppIdMismatch"
//	UNSUPPORTEDOPERATION_APPIDNOTFOUND = "UnsupportedOperation.AppIdNotFound"
//	UNSUPPORTEDOPERATION_CCNNOTATTACHED = "UnsupportedOperation.CcnNotAttached"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) DetachCcnInstancesWithContext(ctx context.Context, request *DetachCcnInstancesRequest) (response *DetachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewDetachCcnInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDetachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDetachClassicLinkVpcRequest() (request *DetachClassicLinkVpcRequest) {
	request = &DetachClassicLinkVpcRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DetachClassicLinkVpc")

	return
}

func NewDetachClassicLinkVpcResponse() (response *DetachClassicLinkVpcResponse) {
	response = &DetachClassicLinkVpcResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DetachClassicLinkVpc
// 本接口(DetachClassicLinkVpc)用于删除私有网络和基础网络设备互通。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DetachClassicLinkVpc(request *DetachClassicLinkVpcRequest) (response *DetachClassicLinkVpcResponse, err error) {
	if request == nil {
		request = NewDetachClassicLinkVpcRequest()
	}

	response = NewDetachClassicLinkVpcResponse()
	err = c.Send(request, response)
	return
}

// DetachClassicLinkVpc
// 本接口(DetachClassicLinkVpc)用于删除私有网络和基础网络设备互通。
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DetachClassicLinkVpcWithContext(ctx context.Context, request *DetachClassicLinkVpcRequest) (response *DetachClassicLinkVpcResponse, err error) {
	if request == nil {
		request = NewDetachClassicLinkVpcRequest()
	}
	request.SetContext(ctx)

	response = NewDetachClassicLinkVpcResponse()
	err = c.Send(request, response)
	return
}

func NewDetachNetworkInterfaceRequest() (request *DetachNetworkInterfaceRequest) {
	request = &DetachNetworkInterfaceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DetachNetworkInterface")

	return
}

func NewDetachNetworkInterfaceResponse() (response *DetachNetworkInterfaceResponse) {
	response = &DetachNetworkInterfaceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DetachNetworkInterface
// 本接口（DetachNetworkInterface）用于弹性网卡解绑云服务器。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) DetachNetworkInterface(request *DetachNetworkInterfaceRequest) (response *DetachNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewDetachNetworkInterfaceRequest()
	}

	response = NewDetachNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

// DetachNetworkInterface
// 本接口（DetachNetworkInterface）用于弹性网卡解绑云服务器。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) DetachNetworkInterfaceWithContext(ctx context.Context, request *DetachNetworkInterfaceRequest) (response *DetachNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewDetachNetworkInterfaceRequest()
	}
	request.SetContext(ctx)

	response = NewDetachNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

func NewDisableCcnRoutesRequest() (request *DisableCcnRoutesRequest) {
	request = &DisableCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisableCcnRoutes")

	return
}

func NewDisableCcnRoutesResponse() (response *DisableCcnRoutesResponse) {
	response = &DisableCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisableCcnRoutes
// 本接口（DisableCcnRoutes）用于禁用已经启用的云联网（CCN）路由
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DisableCcnRoutes(request *DisableCcnRoutesRequest) (response *DisableCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDisableCcnRoutesRequest()
	}

	response = NewDisableCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// DisableCcnRoutes
// 本接口（DisableCcnRoutes）用于禁用已经启用的云联网（CCN）路由
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DisableCcnRoutesWithContext(ctx context.Context, request *DisableCcnRoutesRequest) (response *DisableCcnRoutesResponse, err error) {
	if request == nil {
		request = NewDisableCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDisableCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDisableGatewayFlowMonitorRequest() (request *DisableGatewayFlowMonitorRequest) {
	request = &DisableGatewayFlowMonitorRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisableGatewayFlowMonitor")

	return
}

func NewDisableGatewayFlowMonitorResponse() (response *DisableGatewayFlowMonitorResponse) {
	response = &DisableGatewayFlowMonitorResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisableGatewayFlowMonitor
// 本接口（DisableGatewayFlowMonitor）用于关闭网关流量监控。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) DisableGatewayFlowMonitor(request *DisableGatewayFlowMonitorRequest) (response *DisableGatewayFlowMonitorResponse, err error) {
	if request == nil {
		request = NewDisableGatewayFlowMonitorRequest()
	}

	response = NewDisableGatewayFlowMonitorResponse()
	err = c.Send(request, response)
	return
}

// DisableGatewayFlowMonitor
// 本接口（DisableGatewayFlowMonitor）用于关闭网关流量监控。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) DisableGatewayFlowMonitorWithContext(ctx context.Context, request *DisableGatewayFlowMonitorRequest) (response *DisableGatewayFlowMonitorResponse, err error) {
	if request == nil {
		request = NewDisableGatewayFlowMonitorRequest()
	}
	request.SetContext(ctx)

	response = NewDisableGatewayFlowMonitorResponse()
	err = c.Send(request, response)
	return
}

func NewDisableRoutesRequest() (request *DisableRoutesRequest) {
	request = &DisableRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisableRoutes")

	return
}

func NewDisableRoutesResponse() (response *DisableRoutesResponse) {
	response = &DisableRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisableRoutes
// 本接口（DisableRoutes）用于禁用已启用的子网路由
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_DISABLEDNOTIFYCCN = "UnsupportedOperation.DisabledNotifyCcn"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) DisableRoutes(request *DisableRoutesRequest) (response *DisableRoutesResponse, err error) {
	if request == nil {
		request = NewDisableRoutesRequest()
	}

	response = NewDisableRoutesResponse()
	err = c.Send(request, response)
	return
}

// DisableRoutes
// 本接口（DisableRoutes）用于禁用已启用的子网路由
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_DISABLEDNOTIFYCCN = "UnsupportedOperation.DisabledNotifyCcn"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) DisableRoutesWithContext(ctx context.Context, request *DisableRoutesRequest) (response *DisableRoutesResponse, err error) {
	if request == nil {
		request = NewDisableRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDisableRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateAddressRequest() (request *DisassociateAddressRequest) {
	request = &DisassociateAddressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisassociateAddress")

	return
}

func NewDisassociateAddressResponse() (response *DisassociateAddressResponse) {
	response = &DisassociateAddressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateAddress
// 本接口 (DisassociateAddress) 用于解绑[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）。
//
// * 支持CVM实例，弹性网卡上的EIP解绑
//
// * 不支持NAT上的EIP解绑。NAT上的EIP解绑请参考[DisassociateNatGatewayAddress](https://cloud.tencent.com/document/api/215/36716)
//
// * 只有状态为 BIND 和 BIND_ENI 的 EIP 才能进行解绑定操作。
//
// * EIP 如果被封堵，则不能进行解绑定操作。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED_DAILYALLOCATE = "AddressQuotaLimitExceeded.DailyAllocate"
//	FAILEDOPERATION_ADDRESSENIINFONOTFOUND = "FailedOperation.AddressEniInfoNotFound"
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDADDRESSIDSTATUS_NOTPERMIT = "InvalidAddressIdStatus.NotPermit"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_ONLYSUPPORTEDFORMASTERNETWORKCARD = "InvalidParameterValue.OnlySupportedForMasterNetworkCard"
//	INVALIDPARAMETERVALUE_RESOURCENOTSUPPORT = "InvalidParameterValue.ResourceNotSupport"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
func (c *Client) DisassociateAddress(request *DisassociateAddressRequest) (response *DisassociateAddressResponse, err error) {
	if request == nil {
		request = NewDisassociateAddressRequest()
	}

	response = NewDisassociateAddressResponse()
	err = c.Send(request, response)
	return
}

// DisassociateAddress
// 本接口 (DisassociateAddress) 用于解绑[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）。
//
// * 支持CVM实例，弹性网卡上的EIP解绑
//
// * 不支持NAT上的EIP解绑。NAT上的EIP解绑请参考[DisassociateNatGatewayAddress](https://cloud.tencent.com/document/api/215/36716)
//
// * 只有状态为 BIND 和 BIND_ENI 的 EIP 才能进行解绑定操作。
//
// * EIP 如果被封堵，则不能进行解绑定操作。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED_DAILYALLOCATE = "AddressQuotaLimitExceeded.DailyAllocate"
//	FAILEDOPERATION_ADDRESSENIINFONOTFOUND = "FailedOperation.AddressEniInfoNotFound"
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDADDRESSIDSTATUS_NOTPERMIT = "InvalidAddressIdStatus.NotPermit"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_ONLYSUPPORTEDFORMASTERNETWORKCARD = "InvalidParameterValue.OnlySupportedForMasterNetworkCard"
//	INVALIDPARAMETERVALUE_RESOURCENOTSUPPORT = "InvalidParameterValue.ResourceNotSupport"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
func (c *Client) DisassociateAddressWithContext(ctx context.Context, request *DisassociateAddressRequest) (response *DisassociateAddressResponse, err error) {
	if request == nil {
		request = NewDisassociateAddressRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateAddressResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateDhcpIpWithAddressIpRequest() (request *DisassociateDhcpIpWithAddressIpRequest) {
	request = &DisassociateDhcpIpWithAddressIpRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisassociateDhcpIpWithAddressIp")

	return
}

func NewDisassociateDhcpIpWithAddressIpResponse() (response *DisassociateDhcpIpWithAddressIpResponse) {
	response = &DisassociateDhcpIpWithAddressIpResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateDhcpIpWithAddressIp
// 本接口（DisassociateDhcpIpWithAddressIp）用于将DhcpIp已绑定的弹性公网IP（EIP）解除绑定。<br />
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_UNBINDEIP = "UnsupportedOperation.UnbindEIP"
func (c *Client) DisassociateDhcpIpWithAddressIp(request *DisassociateDhcpIpWithAddressIpRequest) (response *DisassociateDhcpIpWithAddressIpResponse, err error) {
	if request == nil {
		request = NewDisassociateDhcpIpWithAddressIpRequest()
	}

	response = NewDisassociateDhcpIpWithAddressIpResponse()
	err = c.Send(request, response)
	return
}

// DisassociateDhcpIpWithAddressIp
// 本接口（DisassociateDhcpIpWithAddressIp）用于将DhcpIp已绑定的弹性公网IP（EIP）解除绑定。<br />
//
// >?本接口为异步接口，可调用 [DescribeVpcTaskResult](https://cloud.tencent.com/document/api/215/59037) 接口查询任务执行结果，待任务执行成功后再进行其他操作。
//
// >
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_UNBINDEIP = "UnsupportedOperation.UnbindEIP"
func (c *Client) DisassociateDhcpIpWithAddressIpWithContext(ctx context.Context, request *DisassociateDhcpIpWithAddressIpRequest) (response *DisassociateDhcpIpWithAddressIpResponse, err error) {
	if request == nil {
		request = NewDisassociateDhcpIpWithAddressIpRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateDhcpIpWithAddressIpResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateDirectConnectGatewayNatGatewayRequest() (request *DisassociateDirectConnectGatewayNatGatewayRequest) {
	request = &DisassociateDirectConnectGatewayNatGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisassociateDirectConnectGatewayNatGateway")

	return
}

func NewDisassociateDirectConnectGatewayNatGatewayResponse() (response *DisassociateDirectConnectGatewayNatGatewayResponse) {
	response = &DisassociateDirectConnectGatewayNatGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateDirectConnectGatewayNatGateway
// 将专线网关与NAT网关解绑，解绑之后，专线网关将不能通过NAT网关访问公网
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DisassociateDirectConnectGatewayNatGateway(request *DisassociateDirectConnectGatewayNatGatewayRequest) (response *DisassociateDirectConnectGatewayNatGatewayResponse, err error) {
	if request == nil {
		request = NewDisassociateDirectConnectGatewayNatGatewayRequest()
	}

	response = NewDisassociateDirectConnectGatewayNatGatewayResponse()
	err = c.Send(request, response)
	return
}

// DisassociateDirectConnectGatewayNatGateway
// 将专线网关与NAT网关解绑，解绑之后，专线网关将不能通过NAT网关访问公网
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DisassociateDirectConnectGatewayNatGatewayWithContext(ctx context.Context, request *DisassociateDirectConnectGatewayNatGatewayRequest) (response *DisassociateDirectConnectGatewayNatGatewayResponse, err error) {
	if request == nil {
		request = NewDisassociateDirectConnectGatewayNatGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateDirectConnectGatewayNatGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateNatGatewayAddressRequest() (request *DisassociateNatGatewayAddressRequest) {
	request = &DisassociateNatGatewayAddressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisassociateNatGatewayAddress")

	return
}

func NewDisassociateNatGatewayAddressResponse() (response *DisassociateNatGatewayAddressResponse) {
	response = &DisassociateNatGatewayAddressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateNatGatewayAddress
// 本接口（DisassociateNatGatewayAddress）用于NAT网关解绑弹性IP。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSDISASSOCIATE = "UnsupportedOperation.PublicIpAddressDisassociate"
func (c *Client) DisassociateNatGatewayAddress(request *DisassociateNatGatewayAddressRequest) (response *DisassociateNatGatewayAddressResponse, err error) {
	if request == nil {
		request = NewDisassociateNatGatewayAddressRequest()
	}

	response = NewDisassociateNatGatewayAddressResponse()
	err = c.Send(request, response)
	return
}

// DisassociateNatGatewayAddress
// 本接口（DisassociateNatGatewayAddress）用于NAT网关解绑弹性IP。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_PUBLICIPADDRESSDISASSOCIATE = "UnsupportedOperation.PublicIpAddressDisassociate"
func (c *Client) DisassociateNatGatewayAddressWithContext(ctx context.Context, request *DisassociateNatGatewayAddressRequest) (response *DisassociateNatGatewayAddressResponse, err error) {
	if request == nil {
		request = NewDisassociateNatGatewayAddressRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateNatGatewayAddressResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateNetworkAclSubnetsRequest() (request *DisassociateNetworkAclSubnetsRequest) {
	request = &DisassociateNetworkAclSubnetsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisassociateNetworkAclSubnets")

	return
}

func NewDisassociateNetworkAclSubnetsResponse() (response *DisassociateNetworkAclSubnetsResponse) {
	response = &DisassociateNetworkAclSubnetsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateNetworkAclSubnets
// 本接口（DisassociateNetworkAclSubnets）用于网络ACL解关联vpc下的子网。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) DisassociateNetworkAclSubnets(request *DisassociateNetworkAclSubnetsRequest) (response *DisassociateNetworkAclSubnetsResponse, err error) {
	if request == nil {
		request = NewDisassociateNetworkAclSubnetsRequest()
	}

	response = NewDisassociateNetworkAclSubnetsResponse()
	err = c.Send(request, response)
	return
}

// DisassociateNetworkAclSubnets
// 本接口（DisassociateNetworkAclSubnets）用于网络ACL解关联vpc下的子网。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) DisassociateNetworkAclSubnetsWithContext(ctx context.Context, request *DisassociateNetworkAclSubnetsRequest) (response *DisassociateNetworkAclSubnetsResponse, err error) {
	if request == nil {
		request = NewDisassociateNetworkAclSubnetsRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateNetworkAclSubnetsResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateNetworkInterfaceSecurityGroupsRequest() (request *DisassociateNetworkInterfaceSecurityGroupsRequest) {
	request = &DisassociateNetworkInterfaceSecurityGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisassociateNetworkInterfaceSecurityGroups")

	return
}

func NewDisassociateNetworkInterfaceSecurityGroupsResponse() (response *DisassociateNetworkInterfaceSecurityGroupsResponse) {
	response = &DisassociateNetworkInterfaceSecurityGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateNetworkInterfaceSecurityGroups
// 本接口（DisassociateNetworkInterfaceSecurityGroups）用于弹性网卡解绑安全组。支持弹性网卡完全解绑安全组。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DisassociateNetworkInterfaceSecurityGroups(request *DisassociateNetworkInterfaceSecurityGroupsRequest) (response *DisassociateNetworkInterfaceSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewDisassociateNetworkInterfaceSecurityGroupsRequest()
	}

	response = NewDisassociateNetworkInterfaceSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

// DisassociateNetworkInterfaceSecurityGroups
// 本接口（DisassociateNetworkInterfaceSecurityGroups）用于弹性网卡解绑安全组。支持弹性网卡完全解绑安全组。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DisassociateNetworkInterfaceSecurityGroupsWithContext(ctx context.Context, request *DisassociateNetworkInterfaceSecurityGroupsRequest) (response *DisassociateNetworkInterfaceSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewDisassociateNetworkInterfaceSecurityGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateNetworkInterfaceSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateVpcEndPointSecurityGroupsRequest() (request *DisassociateVpcEndPointSecurityGroupsRequest) {
	request = &DisassociateVpcEndPointSecurityGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DisassociateVpcEndPointSecurityGroups")

	return
}

func NewDisassociateVpcEndPointSecurityGroupsResponse() (response *DisassociateVpcEndPointSecurityGroupsResponse) {
	response = &DisassociateVpcEndPointSecurityGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateVpcEndPointSecurityGroups
// 终端节点解绑安全组。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DisassociateVpcEndPointSecurityGroups(request *DisassociateVpcEndPointSecurityGroupsRequest) (response *DisassociateVpcEndPointSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewDisassociateVpcEndPointSecurityGroupsRequest()
	}

	response = NewDisassociateVpcEndPointSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

// DisassociateVpcEndPointSecurityGroups
// 终端节点解绑安全组。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DisassociateVpcEndPointSecurityGroupsWithContext(ctx context.Context, request *DisassociateVpcEndPointSecurityGroupsRequest) (response *DisassociateVpcEndPointSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewDisassociateVpcEndPointSecurityGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateVpcEndPointSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDownloadCustomerGatewayConfigurationRequest() (request *DownloadCustomerGatewayConfigurationRequest) {
	request = &DownloadCustomerGatewayConfigurationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "DownloadCustomerGatewayConfiguration")

	return
}

func NewDownloadCustomerGatewayConfigurationResponse() (response *DownloadCustomerGatewayConfigurationResponse) {
	response = &DownloadCustomerGatewayConfigurationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DownloadCustomerGatewayConfiguration
// 本接口(DownloadCustomerGatewayConfiguration)用于下载VPN通道配置。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DownloadCustomerGatewayConfiguration(request *DownloadCustomerGatewayConfigurationRequest) (response *DownloadCustomerGatewayConfigurationResponse, err error) {
	if request == nil {
		request = NewDownloadCustomerGatewayConfigurationRequest()
	}

	response = NewDownloadCustomerGatewayConfigurationResponse()
	err = c.Send(request, response)
	return
}

// DownloadCustomerGatewayConfiguration
// 本接口(DownloadCustomerGatewayConfiguration)用于下载VPN通道配置。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DownloadCustomerGatewayConfigurationWithContext(ctx context.Context, request *DownloadCustomerGatewayConfigurationRequest) (response *DownloadCustomerGatewayConfigurationResponse, err error) {
	if request == nil {
		request = NewDownloadCustomerGatewayConfigurationRequest()
	}
	request.SetContext(ctx)

	response = NewDownloadCustomerGatewayConfigurationResponse()
	err = c.Send(request, response)
	return
}

func NewEnableCcnRoutesRequest() (request *EnableCcnRoutesRequest) {
	request = &EnableCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "EnableCcnRoutes")

	return
}

func NewEnableCcnRoutesResponse() (response *EnableCcnRoutesResponse) {
	response = &EnableCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// EnableCcnRoutes
// 本接口（EnableCcnRoutes）用于启用已经加入云联网（CCN）的路由。<br />
//
// 本接口会校验启用后，是否与已有路由冲突，如果冲突，则无法启用，失败处理。路由冲突时，需要先禁用与之冲突的路由，才能启用该路由。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
func (c *Client) EnableCcnRoutes(request *EnableCcnRoutesRequest) (response *EnableCcnRoutesResponse, err error) {
	if request == nil {
		request = NewEnableCcnRoutesRequest()
	}

	response = NewEnableCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// EnableCcnRoutes
// 本接口（EnableCcnRoutes）用于启用已经加入云联网（CCN）的路由。<br />
//
// 本接口会校验启用后，是否与已有路由冲突，如果冲突，则无法启用，失败处理。路由冲突时，需要先禁用与之冲突的路由，才能启用该路由。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
func (c *Client) EnableCcnRoutesWithContext(ctx context.Context, request *EnableCcnRoutesRequest) (response *EnableCcnRoutesResponse, err error) {
	if request == nil {
		request = NewEnableCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewEnableCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewEnableGatewayFlowMonitorRequest() (request *EnableGatewayFlowMonitorRequest) {
	request = &EnableGatewayFlowMonitorRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "EnableGatewayFlowMonitor")

	return
}

func NewEnableGatewayFlowMonitorResponse() (response *EnableGatewayFlowMonitorResponse) {
	response = &EnableGatewayFlowMonitorResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// EnableGatewayFlowMonitor
// 本接口（EnableGatewayFlowMonitor）用于开启网关流量监控。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) EnableGatewayFlowMonitor(request *EnableGatewayFlowMonitorRequest) (response *EnableGatewayFlowMonitorResponse, err error) {
	if request == nil {
		request = NewEnableGatewayFlowMonitorRequest()
	}

	response = NewEnableGatewayFlowMonitorResponse()
	err = c.Send(request, response)
	return
}

// EnableGatewayFlowMonitor
// 本接口（EnableGatewayFlowMonitor）用于开启网关流量监控。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) EnableGatewayFlowMonitorWithContext(ctx context.Context, request *EnableGatewayFlowMonitorRequest) (response *EnableGatewayFlowMonitorResponse, err error) {
	if request == nil {
		request = NewEnableGatewayFlowMonitorRequest()
	}
	request.SetContext(ctx)

	response = NewEnableGatewayFlowMonitorResponse()
	err = c.Send(request, response)
	return
}

func NewEnableRoutesRequest() (request *EnableRoutesRequest) {
	request = &EnableRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "EnableRoutes")

	return
}

func NewEnableRoutesResponse() (response *EnableRoutesResponse) {
	response = &EnableRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// EnableRoutes
// 本接口（EnableRoutes）用于启用已禁用的子网路由。<br />
//
// 本接口会校验启用后，是否与已有路由冲突，如果冲突，则无法启用，失败处理。路由冲突时，需要先禁用与之冲突的路由，才能启用该路由。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
//	UNSUPPORTEDOPERATION_ECMPWITHCCNROUTE = "UnsupportedOperation.EcmpWithCcnRoute"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) EnableRoutes(request *EnableRoutesRequest) (response *EnableRoutesResponse, err error) {
	if request == nil {
		request = NewEnableRoutesRequest()
	}

	response = NewEnableRoutesResponse()
	err = c.Send(request, response)
	return
}

// EnableRoutes
// 本接口（EnableRoutes）用于启用已禁用的子网路由。<br />
//
// 本接口会校验启用后，是否与已有路由冲突，如果冲突，则无法启用，失败处理。路由冲突时，需要先禁用与之冲突的路由，才能启用该路由。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
//	UNSUPPORTEDOPERATION_ECMPWITHCCNROUTE = "UnsupportedOperation.EcmpWithCcnRoute"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) EnableRoutesWithContext(ctx context.Context, request *EnableRoutesRequest) (response *EnableRoutesResponse, err error) {
	if request == nil {
		request = NewEnableRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewEnableRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewEnableVpcEndPointConnectRequest() (request *EnableVpcEndPointConnectRequest) {
	request = &EnableVpcEndPointConnectRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "EnableVpcEndPointConnect")

	return
}

func NewEnableVpcEndPointConnectResponse() (response *EnableVpcEndPointConnectResponse) {
	response = &EnableVpcEndPointConnectResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// EnableVpcEndPointConnect
// 是否接受终端节点连接请求。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) EnableVpcEndPointConnect(request *EnableVpcEndPointConnectRequest) (response *EnableVpcEndPointConnectResponse, err error) {
	if request == nil {
		request = NewEnableVpcEndPointConnectRequest()
	}

	response = NewEnableVpcEndPointConnectResponse()
	err = c.Send(request, response)
	return
}

// EnableVpcEndPointConnect
// 是否接受终端节点连接请求。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) EnableVpcEndPointConnectWithContext(ctx context.Context, request *EnableVpcEndPointConnectRequest) (response *EnableVpcEndPointConnectResponse, err error) {
	if request == nil {
		request = NewEnableVpcEndPointConnectRequest()
	}
	request.SetContext(ctx)

	response = NewEnableVpcEndPointConnectResponse()
	err = c.Send(request, response)
	return
}

func NewGetCcnRegionBandwidthLimitsRequest() (request *GetCcnRegionBandwidthLimitsRequest) {
	request = &GetCcnRegionBandwidthLimitsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "GetCcnRegionBandwidthLimits")

	return
}

func NewGetCcnRegionBandwidthLimitsResponse() (response *GetCcnRegionBandwidthLimitsResponse) {
	response = &GetCcnRegionBandwidthLimitsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// GetCcnRegionBandwidthLimits
// 本接口（GetCcnRegionBandwidthLimits）用于查询云联网相关地域带宽信息，其中预付费模式的云联网仅支持地域间限速，后付费模式的云联网支持地域间限速和地域出口限速。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) GetCcnRegionBandwidthLimits(request *GetCcnRegionBandwidthLimitsRequest) (response *GetCcnRegionBandwidthLimitsResponse, err error) {
	if request == nil {
		request = NewGetCcnRegionBandwidthLimitsRequest()
	}

	response = NewGetCcnRegionBandwidthLimitsResponse()
	err = c.Send(request, response)
	return
}

// GetCcnRegionBandwidthLimits
// 本接口（GetCcnRegionBandwidthLimits）用于查询云联网相关地域带宽信息，其中预付费模式的云联网仅支持地域间限速，后付费模式的云联网支持地域间限速和地域出口限速。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) GetCcnRegionBandwidthLimitsWithContext(ctx context.Context, request *GetCcnRegionBandwidthLimitsRequest) (response *GetCcnRegionBandwidthLimitsResponse, err error) {
	if request == nil {
		request = NewGetCcnRegionBandwidthLimitsRequest()
	}
	request.SetContext(ctx)

	response = NewGetCcnRegionBandwidthLimitsResponse()
	err = c.Send(request, response)
	return
}

func NewHaVipAssociateAddressIpRequest() (request *HaVipAssociateAddressIpRequest) {
	request = &HaVipAssociateAddressIpRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "HaVipAssociateAddressIp")

	return
}

func NewHaVipAssociateAddressIpResponse() (response *HaVipAssociateAddressIpResponse) {
	response = &HaVipAssociateAddressIpResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// HaVipAssociateAddressIp
// 本接口（HaVipAssociateAddressIp）用于高可用虚拟IP（HAVIP）绑定弹性公网IP（EIP）。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_BINDEIP = "UnsupportedOperation.BindEIP"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) HaVipAssociateAddressIp(request *HaVipAssociateAddressIpRequest) (response *HaVipAssociateAddressIpResponse, err error) {
	if request == nil {
		request = NewHaVipAssociateAddressIpRequest()
	}

	response = NewHaVipAssociateAddressIpResponse()
	err = c.Send(request, response)
	return
}

// HaVipAssociateAddressIp
// 本接口（HaVipAssociateAddressIp）用于高可用虚拟IP（HAVIP）绑定弹性公网IP（EIP）。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_BINDEIP = "UnsupportedOperation.BindEIP"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) HaVipAssociateAddressIpWithContext(ctx context.Context, request *HaVipAssociateAddressIpRequest) (response *HaVipAssociateAddressIpResponse, err error) {
	if request == nil {
		request = NewHaVipAssociateAddressIpRequest()
	}
	request.SetContext(ctx)

	response = NewHaVipAssociateAddressIpResponse()
	err = c.Send(request, response)
	return
}

func NewHaVipDisassociateAddressIpRequest() (request *HaVipDisassociateAddressIpRequest) {
	request = &HaVipDisassociateAddressIpRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "HaVipDisassociateAddressIp")

	return
}

func NewHaVipDisassociateAddressIpResponse() (response *HaVipDisassociateAddressIpResponse) {
	response = &HaVipDisassociateAddressIpResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// HaVipDisassociateAddressIp
// 本接口（HaVipDisassociateAddressIp）用于将高可用虚拟IP（HAVIP）已绑定的弹性公网IP（EIP）解除绑定。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_UNBINDEIP = "UnsupportedOperation.UnbindEIP"
func (c *Client) HaVipDisassociateAddressIp(request *HaVipDisassociateAddressIpRequest) (response *HaVipDisassociateAddressIpResponse, err error) {
	if request == nil {
		request = NewHaVipDisassociateAddressIpRequest()
	}

	response = NewHaVipDisassociateAddressIpResponse()
	err = c.Send(request, response)
	return
}

// HaVipDisassociateAddressIp
// 本接口（HaVipDisassociateAddressIp）用于将高可用虚拟IP（HAVIP）已绑定的弹性公网IP（EIP）解除绑定。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_UNBINDEIP = "UnsupportedOperation.UnbindEIP"
func (c *Client) HaVipDisassociateAddressIpWithContext(ctx context.Context, request *HaVipDisassociateAddressIpRequest) (response *HaVipDisassociateAddressIpResponse, err error) {
	if request == nil {
		request = NewHaVipDisassociateAddressIpRequest()
	}
	request.SetContext(ctx)

	response = NewHaVipDisassociateAddressIpResponse()
	err = c.Send(request, response)
	return
}

func NewInquirePriceCreateDirectConnectGatewayRequest() (request *InquirePriceCreateDirectConnectGatewayRequest) {
	request = &InquirePriceCreateDirectConnectGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "InquirePriceCreateDirectConnectGateway")

	return
}

func NewInquirePriceCreateDirectConnectGatewayResponse() (response *InquirePriceCreateDirectConnectGatewayResponse) {
	response = &InquirePriceCreateDirectConnectGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquirePriceCreateDirectConnectGateway
// 本接口（DescribePriceCreateDirectConnectGateway）用于创建专线网关询价。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_UNBINDEIP = "UnsupportedOperation.UnbindEIP"
func (c *Client) InquirePriceCreateDirectConnectGateway(request *InquirePriceCreateDirectConnectGatewayRequest) (response *InquirePriceCreateDirectConnectGatewayResponse, err error) {
	if request == nil {
		request = NewInquirePriceCreateDirectConnectGatewayRequest()
	}

	response = NewInquirePriceCreateDirectConnectGatewayResponse()
	err = c.Send(request, response)
	return
}

// InquirePriceCreateDirectConnectGateway
// 本接口（DescribePriceCreateDirectConnectGateway）用于创建专线网关询价。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_UNBINDEIP = "UnsupportedOperation.UnbindEIP"
func (c *Client) InquirePriceCreateDirectConnectGatewayWithContext(ctx context.Context, request *InquirePriceCreateDirectConnectGatewayRequest) (response *InquirePriceCreateDirectConnectGatewayResponse, err error) {
	if request == nil {
		request = NewInquirePriceCreateDirectConnectGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewInquirePriceCreateDirectConnectGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceCreateVpnGatewayRequest() (request *InquiryPriceCreateVpnGatewayRequest) {
	request = &InquiryPriceCreateVpnGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "InquiryPriceCreateVpnGateway")

	return
}

func NewInquiryPriceCreateVpnGatewayResponse() (response *InquiryPriceCreateVpnGatewayResponse) {
	response = &InquiryPriceCreateVpnGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceCreateVpnGateway
// 本接口（InquiryPriceCreateVpnGateway）用于创建VPN网关询价。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) InquiryPriceCreateVpnGateway(request *InquiryPriceCreateVpnGatewayRequest) (response *InquiryPriceCreateVpnGatewayResponse, err error) {
	if request == nil {
		request = NewInquiryPriceCreateVpnGatewayRequest()
	}

	response = NewInquiryPriceCreateVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceCreateVpnGateway
// 本接口（InquiryPriceCreateVpnGateway）用于创建VPN网关询价。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) InquiryPriceCreateVpnGatewayWithContext(ctx context.Context, request *InquiryPriceCreateVpnGatewayRequest) (response *InquiryPriceCreateVpnGatewayResponse, err error) {
	if request == nil {
		request = NewInquiryPriceCreateVpnGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceCreateVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceRenewVpnGatewayRequest() (request *InquiryPriceRenewVpnGatewayRequest) {
	request = &InquiryPriceRenewVpnGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "InquiryPriceRenewVpnGateway")

	return
}

func NewInquiryPriceRenewVpnGatewayResponse() (response *InquiryPriceRenewVpnGatewayResponse) {
	response = &InquiryPriceRenewVpnGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceRenewVpnGateway
// 本接口（InquiryPriceRenewVpnGateway）用于续费VPN网关询价。目前仅支持IPSEC类型网关的询价。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) InquiryPriceRenewVpnGateway(request *InquiryPriceRenewVpnGatewayRequest) (response *InquiryPriceRenewVpnGatewayResponse, err error) {
	if request == nil {
		request = NewInquiryPriceRenewVpnGatewayRequest()
	}

	response = NewInquiryPriceRenewVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceRenewVpnGateway
// 本接口（InquiryPriceRenewVpnGateway）用于续费VPN网关询价。目前仅支持IPSEC类型网关的询价。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) InquiryPriceRenewVpnGatewayWithContext(ctx context.Context, request *InquiryPriceRenewVpnGatewayRequest) (response *InquiryPriceRenewVpnGatewayResponse, err error) {
	if request == nil {
		request = NewInquiryPriceRenewVpnGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceRenewVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceResetVpnGatewayInternetMaxBandwidthRequest() (request *InquiryPriceResetVpnGatewayInternetMaxBandwidthRequest) {
	request = &InquiryPriceResetVpnGatewayInternetMaxBandwidthRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "InquiryPriceResetVpnGatewayInternetMaxBandwidth")

	return
}

func NewInquiryPriceResetVpnGatewayInternetMaxBandwidthResponse() (response *InquiryPriceResetVpnGatewayInternetMaxBandwidthResponse) {
	response = &InquiryPriceResetVpnGatewayInternetMaxBandwidthResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceResetVpnGatewayInternetMaxBandwidth
// 本接口（InquiryPriceResetVpnGatewayInternetMaxBandwidth）调整VPN网关带宽上限询价。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) InquiryPriceResetVpnGatewayInternetMaxBandwidth(request *InquiryPriceResetVpnGatewayInternetMaxBandwidthRequest) (response *InquiryPriceResetVpnGatewayInternetMaxBandwidthResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResetVpnGatewayInternetMaxBandwidthRequest()
	}

	response = NewInquiryPriceResetVpnGatewayInternetMaxBandwidthResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceResetVpnGatewayInternetMaxBandwidth
// 本接口（InquiryPriceResetVpnGatewayInternetMaxBandwidth）调整VPN网关带宽上限询价。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) InquiryPriceResetVpnGatewayInternetMaxBandwidthWithContext(ctx context.Context, request *InquiryPriceResetVpnGatewayInternetMaxBandwidthRequest) (response *InquiryPriceResetVpnGatewayInternetMaxBandwidthResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResetVpnGatewayInternetMaxBandwidthRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceResetVpnGatewayInternetMaxBandwidthResponse()
	err = c.Send(request, response)
	return
}

func NewMigrateNetworkInterfaceRequest() (request *MigrateNetworkInterfaceRequest) {
	request = &MigrateNetworkInterfaceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "MigrateNetworkInterface")

	return
}

func NewMigrateNetworkInterfaceResponse() (response *MigrateNetworkInterfaceResponse) {
	response = &MigrateNetworkInterfaceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// MigrateNetworkInterface
// 本接口（MigrateNetworkInterface）用于弹性网卡迁移。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) MigrateNetworkInterface(request *MigrateNetworkInterfaceRequest) (response *MigrateNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewMigrateNetworkInterfaceRequest()
	}

	response = NewMigrateNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

// MigrateNetworkInterface
// 本接口（MigrateNetworkInterface）用于弹性网卡迁移。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) MigrateNetworkInterfaceWithContext(ctx context.Context, request *MigrateNetworkInterfaceRequest) (response *MigrateNetworkInterfaceResponse, err error) {
	if request == nil {
		request = NewMigrateNetworkInterfaceRequest()
	}
	request.SetContext(ctx)

	response = NewMigrateNetworkInterfaceResponse()
	err = c.Send(request, response)
	return
}

func NewMigratePrivateIpAddressRequest() (request *MigratePrivateIpAddressRequest) {
	request = &MigratePrivateIpAddressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "MigratePrivateIpAddress")

	return
}

func NewMigratePrivateIpAddressResponse() (response *MigratePrivateIpAddressResponse) {
	response = &MigratePrivateIpAddressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// MigratePrivateIpAddress
//
//	本接口（MigratePrivateIpAddress）用于弹性网卡内网IP迁移。
//
// * 该接口用于将一个内网IP从一个弹性网卡上迁移到另外一个弹性网卡，主IP地址不支持迁移。
//
// * 迁移前后的弹性网卡必须在同一个子网内。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_ATTACHMENTNOTFOUND = "UnauthorizedOperation.AttachmentNotFound"
//	UNAUTHORIZEDOPERATION_PRIMARYIP = "UnauthorizedOperation.PrimaryIp"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_PRIMARYIP = "UnsupportedOperation.PrimaryIp"
func (c *Client) MigratePrivateIpAddress(request *MigratePrivateIpAddressRequest) (response *MigratePrivateIpAddressResponse, err error) {
	if request == nil {
		request = NewMigratePrivateIpAddressRequest()
	}

	response = NewMigratePrivateIpAddressResponse()
	err = c.Send(request, response)
	return
}

// MigratePrivateIpAddress
//
//	本接口（MigratePrivateIpAddress）用于弹性网卡内网IP迁移。
//
// * 该接口用于将一个内网IP从一个弹性网卡上迁移到另外一个弹性网卡，主IP地址不支持迁移。
//
// * 迁移前后的弹性网卡必须在同一个子网内。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_ATTACHMENTNOTFOUND = "UnauthorizedOperation.AttachmentNotFound"
//	UNAUTHORIZEDOPERATION_PRIMARYIP = "UnauthorizedOperation.PrimaryIp"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
//	UNSUPPORTEDOPERATION_PRIMARYIP = "UnsupportedOperation.PrimaryIp"
func (c *Client) MigratePrivateIpAddressWithContext(ctx context.Context, request *MigratePrivateIpAddressRequest) (response *MigratePrivateIpAddressResponse, err error) {
	if request == nil {
		request = NewMigratePrivateIpAddressRequest()
	}
	request.SetContext(ctx)

	response = NewMigratePrivateIpAddressResponse()
	err = c.Send(request, response)
	return
}

func NewModifyAddressAttributeRequest() (request *ModifyAddressAttributeRequest) {
	request = &ModifyAddressAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyAddressAttribute")

	return
}

func NewModifyAddressAttributeResponse() (response *ModifyAddressAttributeResponse) {
	response = &ModifyAddressAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyAddressAttribute
// 本接口 (ModifyAddressAttribute) 用于修改[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）的名称。
//
// 可能返回的错误码:
//
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
func (c *Client) ModifyAddressAttribute(request *ModifyAddressAttributeRequest) (response *ModifyAddressAttributeResponse, err error) {
	if request == nil {
		request = NewModifyAddressAttributeRequest()
	}

	response = NewModifyAddressAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyAddressAttribute
// 本接口 (ModifyAddressAttribute) 用于修改[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）的名称。
//
// 可能返回的错误码:
//
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
func (c *Client) ModifyAddressAttributeWithContext(ctx context.Context, request *ModifyAddressAttributeRequest) (response *ModifyAddressAttributeResponse, err error) {
	if request == nil {
		request = NewModifyAddressAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyAddressAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyAddressInternetChargeTypeRequest() (request *ModifyAddressInternetChargeTypeRequest) {
	request = &ModifyAddressInternetChargeTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyAddressInternetChargeType")

	return
}

func NewModifyAddressInternetChargeTypeResponse() (response *ModifyAddressInternetChargeTypeResponse) {
	response = &ModifyAddressInternetChargeTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyAddressInternetChargeType
// 该接口用于调整具有带宽属性弹性公网IP的网络计费模式
//
// * 支持BANDWIDTH_PREPAID_BY_MONTH和TRAFFIC_POSTPAID_BY_HOUR两种网络计费模式之间的切换。
//
// * 每个弹性公网IP支持调整两次，次数超出则无法调整。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDADDRESSIDSTATE_INARREARS = "InvalidAddressIdState.InArrears"
//	INVALIDADDRESSSTATE = "InvalidAddressState"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTCALCIP = "InvalidParameterValue.AddressNotCalcIP"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_INTERNETCHARGETYPENOTCHANGED = "InvalidParameterValue.InternetChargeTypeNotChanged"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	LIMITEXCEEDED = "LimitExceeded"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
//	UNSUPPORTEDOPERATION_NATNOTSUPPORTED = "UnsupportedOperation.NatNotSupported"
func (c *Client) ModifyAddressInternetChargeType(request *ModifyAddressInternetChargeTypeRequest) (response *ModifyAddressInternetChargeTypeResponse, err error) {
	if request == nil {
		request = NewModifyAddressInternetChargeTypeRequest()
	}

	response = NewModifyAddressInternetChargeTypeResponse()
	err = c.Send(request, response)
	return
}

// ModifyAddressInternetChargeType
// 该接口用于调整具有带宽属性弹性公网IP的网络计费模式
//
// * 支持BANDWIDTH_PREPAID_BY_MONTH和TRAFFIC_POSTPAID_BY_HOUR两种网络计费模式之间的切换。
//
// * 每个弹性公网IP支持调整两次，次数超出则无法调整。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDADDRESSIDSTATE_INARREARS = "InvalidAddressIdState.InArrears"
//	INVALIDADDRESSSTATE = "InvalidAddressState"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTCALCIP = "InvalidParameterValue.AddressNotCalcIP"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_INTERNETCHARGETYPENOTCHANGED = "InvalidParameterValue.InternetChargeTypeNotChanged"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	LIMITEXCEEDED = "LimitExceeded"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INVALIDACTION = "UnsupportedOperation.InvalidAction"
//	UNSUPPORTEDOPERATION_NATNOTSUPPORTED = "UnsupportedOperation.NatNotSupported"
func (c *Client) ModifyAddressInternetChargeTypeWithContext(ctx context.Context, request *ModifyAddressInternetChargeTypeRequest) (response *ModifyAddressInternetChargeTypeResponse, err error) {
	if request == nil {
		request = NewModifyAddressInternetChargeTypeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyAddressInternetChargeTypeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyAddressTemplateAttributeRequest() (request *ModifyAddressTemplateAttributeRequest) {
	request = &ModifyAddressTemplateAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyAddressTemplateAttribute")

	return
}

func NewModifyAddressTemplateAttributeResponse() (response *ModifyAddressTemplateAttributeResponse) {
	response = &ModifyAddressTemplateAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyAddressTemplateAttribute
// 本接口（ModifyAddressTemplateAttribute）用于修改IP地址模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) ModifyAddressTemplateAttribute(request *ModifyAddressTemplateAttributeRequest) (response *ModifyAddressTemplateAttributeResponse, err error) {
	if request == nil {
		request = NewModifyAddressTemplateAttributeRequest()
	}

	response = NewModifyAddressTemplateAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyAddressTemplateAttribute
// 本接口（ModifyAddressTemplateAttribute）用于修改IP地址模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) ModifyAddressTemplateAttributeWithContext(ctx context.Context, request *ModifyAddressTemplateAttributeRequest) (response *ModifyAddressTemplateAttributeResponse, err error) {
	if request == nil {
		request = NewModifyAddressTemplateAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyAddressTemplateAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyAddressTemplateGroupAttributeRequest() (request *ModifyAddressTemplateGroupAttributeRequest) {
	request = &ModifyAddressTemplateGroupAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyAddressTemplateGroupAttribute")

	return
}

func NewModifyAddressTemplateGroupAttributeResponse() (response *ModifyAddressTemplateGroupAttributeResponse) {
	response = &ModifyAddressTemplateGroupAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyAddressTemplateGroupAttribute
// 本接口（ModifyAddressTemplateGroupAttribute）用于修改IP地址模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) ModifyAddressTemplateGroupAttribute(request *ModifyAddressTemplateGroupAttributeRequest) (response *ModifyAddressTemplateGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifyAddressTemplateGroupAttributeRequest()
	}

	response = NewModifyAddressTemplateGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyAddressTemplateGroupAttribute
// 本接口（ModifyAddressTemplateGroupAttribute）用于修改IP地址模板集合
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) ModifyAddressTemplateGroupAttributeWithContext(ctx context.Context, request *ModifyAddressTemplateGroupAttributeRequest) (response *ModifyAddressTemplateGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifyAddressTemplateGroupAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyAddressTemplateGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyAddressesBandwidthRequest() (request *ModifyAddressesBandwidthRequest) {
	request = &ModifyAddressesBandwidthRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyAddressesBandwidth")

	return
}

func NewModifyAddressesBandwidthResponse() (response *ModifyAddressesBandwidthResponse) {
	response = &ModifyAddressesBandwidthResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyAddressesBandwidth
// 本接口（ModifyAddressesBandwidth）用于调整[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)(简称EIP)带宽，支持后付费EIP, 预付费EIP和带宽包EIP
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_BANDWIDTHOUTOFRANGE = "InvalidParameterValue.BandwidthOutOfRange"
//	INVALIDPARAMETERVALUE_BANDWIDTHTOOSMALL = "InvalidParameterValue.BandwidthTooSmall"
//	INVALIDPARAMETERVALUE_INCONSISTENTINSTANCEINTERNETCHARGETYPE = "InvalidParameterValue.InconsistentInstanceInternetChargeType"
//	INVALIDPARAMETERVALUE_INSTANCENOCALCIP = "InvalidParameterValue.InstanceNoCalcIP"
//	INVALIDPARAMETERVALUE_INSTANCENOWANIP = "InvalidParameterValue.InstanceNoWanIP"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RESOURCEEXPIRED = "InvalidParameterValue.ResourceExpired"
//	INVALIDPARAMETERVALUE_RESOURCENOTEXISTED = "InvalidParameterValue.ResourceNotExisted"
//	INVALIDPARAMETERVALUE_RESOURCENOTSUPPORT = "InvalidParameterValue.ResourceNotSupport"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
func (c *Client) ModifyAddressesBandwidth(request *ModifyAddressesBandwidthRequest) (response *ModifyAddressesBandwidthResponse, err error) {
	if request == nil {
		request = NewModifyAddressesBandwidthRequest()
	}

	response = NewModifyAddressesBandwidthResponse()
	err = c.Send(request, response)
	return
}

// ModifyAddressesBandwidth
// 本接口（ModifyAddressesBandwidth）用于调整[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)(简称EIP)带宽，支持后付费EIP, 预付费EIP和带宽包EIP
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	INVALIDPARAMETERVALUE_BANDWIDTHOUTOFRANGE = "InvalidParameterValue.BandwidthOutOfRange"
//	INVALIDPARAMETERVALUE_BANDWIDTHTOOSMALL = "InvalidParameterValue.BandwidthTooSmall"
//	INVALIDPARAMETERVALUE_INCONSISTENTINSTANCEINTERNETCHARGETYPE = "InvalidParameterValue.InconsistentInstanceInternetChargeType"
//	INVALIDPARAMETERVALUE_INSTANCENOCALCIP = "InvalidParameterValue.InstanceNoCalcIP"
//	INVALIDPARAMETERVALUE_INSTANCENOWANIP = "InvalidParameterValue.InstanceNoWanIP"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RESOURCEEXPIRED = "InvalidParameterValue.ResourceExpired"
//	INVALIDPARAMETERVALUE_RESOURCENOTEXISTED = "InvalidParameterValue.ResourceNotExisted"
//	INVALIDPARAMETERVALUE_RESOURCENOTSUPPORT = "InvalidParameterValue.ResourceNotSupport"
//	UNSUPPORTEDOPERATION_ACTIONNOTFOUND = "UnsupportedOperation.ActionNotFound"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INSTANCESTATENOTSUPPORTED = "UnsupportedOperation.InstanceStateNotSupported"
func (c *Client) ModifyAddressesBandwidthWithContext(ctx context.Context, request *ModifyAddressesBandwidthRequest) (response *ModifyAddressesBandwidthResponse, err error) {
	if request == nil {
		request = NewModifyAddressesBandwidthRequest()
	}
	request.SetContext(ctx)

	response = NewModifyAddressesBandwidthResponse()
	err = c.Send(request, response)
	return
}

func NewModifyAssistantCidrRequest() (request *ModifyAssistantCidrRequest) {
	request = &ModifyAssistantCidrRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyAssistantCidr")

	return
}

func NewModifyAssistantCidrResponse() (response *ModifyAssistantCidrResponse) {
	response = &ModifyAssistantCidrResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyAssistantCidr
// 本接口(ModifyAssistantCidr)用于批量修改辅助CIDR，支持新增和删除。（接口灰度中，如需使用请提工单。）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETOVERLAPASSISTCIDR = "InvalidParameterValue.SubnetOverlapAssistCidr"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyAssistantCidr(request *ModifyAssistantCidrRequest) (response *ModifyAssistantCidrResponse, err error) {
	if request == nil {
		request = NewModifyAssistantCidrRequest()
	}

	response = NewModifyAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

// ModifyAssistantCidr
// 本接口(ModifyAssistantCidr)用于批量修改辅助CIDR，支持新增和删除。（接口灰度中，如需使用请提工单。）
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_SUBNETCONFLICT = "InvalidParameterValue.SubnetConflict"
//	INVALIDPARAMETERVALUE_SUBNETOVERLAPASSISTCIDR = "InvalidParameterValue.SubnetOverlapAssistCidr"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyAssistantCidrWithContext(ctx context.Context, request *ModifyAssistantCidrRequest) (response *ModifyAssistantCidrResponse, err error) {
	if request == nil {
		request = NewModifyAssistantCidrRequest()
	}
	request.SetContext(ctx)

	response = NewModifyAssistantCidrResponse()
	err = c.Send(request, response)
	return
}

func NewModifyBandwidthPackageAttributeRequest() (request *ModifyBandwidthPackageAttributeRequest) {
	request = &ModifyBandwidthPackageAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyBandwidthPackageAttribute")

	return
}

func NewModifyBandwidthPackageAttributeResponse() (response *ModifyBandwidthPackageAttributeResponse) {
	response = &ModifyBandwidthPackageAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyBandwidthPackageAttribute
// 接口用于修改带宽包属性，包括带宽包名字等
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_INVALIDBANDWIDTHPACKAGECHARGETYPE = "InvalidParameterValue.InvalidBandwidthPackageChargeType"
func (c *Client) ModifyBandwidthPackageAttribute(request *ModifyBandwidthPackageAttributeRequest) (response *ModifyBandwidthPackageAttributeResponse, err error) {
	if request == nil {
		request = NewModifyBandwidthPackageAttributeRequest()
	}

	response = NewModifyBandwidthPackageAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyBandwidthPackageAttribute
// 接口用于修改带宽包属性，包括带宽包名字等
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_INVALIDBANDWIDTHPACKAGECHARGETYPE = "InvalidParameterValue.InvalidBandwidthPackageChargeType"
func (c *Client) ModifyBandwidthPackageAttributeWithContext(ctx context.Context, request *ModifyBandwidthPackageAttributeRequest) (response *ModifyBandwidthPackageAttributeResponse, err error) {
	if request == nil {
		request = NewModifyBandwidthPackageAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyBandwidthPackageAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyCcnAttachedInstancesAttributeRequest() (request *ModifyCcnAttachedInstancesAttributeRequest) {
	request = &ModifyCcnAttachedInstancesAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyCcnAttachedInstancesAttribute")

	return
}

func NewModifyCcnAttachedInstancesAttributeResponse() (response *ModifyCcnAttachedInstancesAttributeResponse) {
	response = &ModifyCcnAttachedInstancesAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyCcnAttachedInstancesAttribute
// 修改CCN关联实例属性，目前仅修改备注description
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyCcnAttachedInstancesAttribute(request *ModifyCcnAttachedInstancesAttributeRequest) (response *ModifyCcnAttachedInstancesAttributeResponse, err error) {
	if request == nil {
		request = NewModifyCcnAttachedInstancesAttributeRequest()
	}

	response = NewModifyCcnAttachedInstancesAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyCcnAttachedInstancesAttribute
// 修改CCN关联实例属性，目前仅修改备注description
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyCcnAttachedInstancesAttributeWithContext(ctx context.Context, request *ModifyCcnAttachedInstancesAttributeRequest) (response *ModifyCcnAttachedInstancesAttributeResponse, err error) {
	if request == nil {
		request = NewModifyCcnAttachedInstancesAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyCcnAttachedInstancesAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyCcnAttributeRequest() (request *ModifyCcnAttributeRequest) {
	request = &ModifyCcnAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyCcnAttribute")

	return
}

func NewModifyCcnAttributeResponse() (response *ModifyCcnAttributeResponse) {
	response = &ModifyCcnAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyCcnAttribute
// 本接口（ModifyCcnAttribute）用于修改云联网（CCN）的相关属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyCcnAttribute(request *ModifyCcnAttributeRequest) (response *ModifyCcnAttributeResponse, err error) {
	if request == nil {
		request = NewModifyCcnAttributeRequest()
	}

	response = NewModifyCcnAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyCcnAttribute
// 本接口（ModifyCcnAttribute）用于修改云联网（CCN）的相关属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyCcnAttributeWithContext(ctx context.Context, request *ModifyCcnAttributeRequest) (response *ModifyCcnAttributeResponse, err error) {
	if request == nil {
		request = NewModifyCcnAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyCcnAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyCcnRegionBandwidthLimitsTypeRequest() (request *ModifyCcnRegionBandwidthLimitsTypeRequest) {
	request = &ModifyCcnRegionBandwidthLimitsTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyCcnRegionBandwidthLimitsType")

	return
}

func NewModifyCcnRegionBandwidthLimitsTypeResponse() (response *ModifyCcnRegionBandwidthLimitsTypeResponse) {
	response = &ModifyCcnRegionBandwidthLimitsTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyCcnRegionBandwidthLimitsType
// 本接口（ModifyCcnRegionBandwidthLimitsType）用于修改后付费云联网实例修改带宽限速策略。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_NOTPOSTPAIDCCNOPERATION = "UnsupportedOperation.NotPostpaidCcnOperation"
func (c *Client) ModifyCcnRegionBandwidthLimitsType(request *ModifyCcnRegionBandwidthLimitsTypeRequest) (response *ModifyCcnRegionBandwidthLimitsTypeResponse, err error) {
	if request == nil {
		request = NewModifyCcnRegionBandwidthLimitsTypeRequest()
	}

	response = NewModifyCcnRegionBandwidthLimitsTypeResponse()
	err = c.Send(request, response)
	return
}

// ModifyCcnRegionBandwidthLimitsType
// 本接口（ModifyCcnRegionBandwidthLimitsType）用于修改后付费云联网实例修改带宽限速策略。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_NOTPOSTPAIDCCNOPERATION = "UnsupportedOperation.NotPostpaidCcnOperation"
func (c *Client) ModifyCcnRegionBandwidthLimitsTypeWithContext(ctx context.Context, request *ModifyCcnRegionBandwidthLimitsTypeRequest) (response *ModifyCcnRegionBandwidthLimitsTypeResponse, err error) {
	if request == nil {
		request = NewModifyCcnRegionBandwidthLimitsTypeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyCcnRegionBandwidthLimitsTypeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyCustomerGatewayAttributeRequest() (request *ModifyCustomerGatewayAttributeRequest) {
	request = &ModifyCustomerGatewayAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyCustomerGatewayAttribute")

	return
}

func NewModifyCustomerGatewayAttributeResponse() (response *ModifyCustomerGatewayAttributeResponse) {
	response = &ModifyCustomerGatewayAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyCustomerGatewayAttribute
// 本接口（ModifyCustomerGatewayAttribute）用于修改对端网关信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyCustomerGatewayAttribute(request *ModifyCustomerGatewayAttributeRequest) (response *ModifyCustomerGatewayAttributeResponse, err error) {
	if request == nil {
		request = NewModifyCustomerGatewayAttributeRequest()
	}

	response = NewModifyCustomerGatewayAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyCustomerGatewayAttribute
// 本接口（ModifyCustomerGatewayAttribute）用于修改对端网关信息。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyCustomerGatewayAttributeWithContext(ctx context.Context, request *ModifyCustomerGatewayAttributeRequest) (response *ModifyCustomerGatewayAttributeResponse, err error) {
	if request == nil {
		request = NewModifyCustomerGatewayAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyCustomerGatewayAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyDhcpIpAttributeRequest() (request *ModifyDhcpIpAttributeRequest) {
	request = &ModifyDhcpIpAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyDhcpIpAttribute")

	return
}

func NewModifyDhcpIpAttributeResponse() (response *ModifyDhcpIpAttributeResponse) {
	response = &ModifyDhcpIpAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyDhcpIpAttribute
// 本接口（ModifyDhcpIpAttribute）用于修改DhcpIp属性
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyDhcpIpAttribute(request *ModifyDhcpIpAttributeRequest) (response *ModifyDhcpIpAttributeResponse, err error) {
	if request == nil {
		request = NewModifyDhcpIpAttributeRequest()
	}

	response = NewModifyDhcpIpAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyDhcpIpAttribute
// 本接口（ModifyDhcpIpAttribute）用于修改DhcpIp属性
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyDhcpIpAttributeWithContext(ctx context.Context, request *ModifyDhcpIpAttributeRequest) (response *ModifyDhcpIpAttributeResponse, err error) {
	if request == nil {
		request = NewModifyDhcpIpAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyDhcpIpAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyDirectConnectGatewayAttributeRequest() (request *ModifyDirectConnectGatewayAttributeRequest) {
	request = &ModifyDirectConnectGatewayAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyDirectConnectGatewayAttribute")

	return
}

func NewModifyDirectConnectGatewayAttributeResponse() (response *ModifyDirectConnectGatewayAttributeResponse) {
	response = &ModifyDirectConnectGatewayAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyDirectConnectGatewayAttribute
// 本接口（ModifyDirectConnectGatewayAttribute）用于修改专线网关属性
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_DIRECTCONNECTGATEWAYISUPDATINGCOMMUNITY = "UnsupportedOperation.DirectConnectGatewayIsUpdatingCommunity"
func (c *Client) ModifyDirectConnectGatewayAttribute(request *ModifyDirectConnectGatewayAttributeRequest) (response *ModifyDirectConnectGatewayAttributeResponse, err error) {
	if request == nil {
		request = NewModifyDirectConnectGatewayAttributeRequest()
	}

	response = NewModifyDirectConnectGatewayAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyDirectConnectGatewayAttribute
// 本接口（ModifyDirectConnectGatewayAttribute）用于修改专线网关属性
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_DIRECTCONNECTGATEWAYISUPDATINGCOMMUNITY = "UnsupportedOperation.DirectConnectGatewayIsUpdatingCommunity"
func (c *Client) ModifyDirectConnectGatewayAttributeWithContext(ctx context.Context, request *ModifyDirectConnectGatewayAttributeRequest) (response *ModifyDirectConnectGatewayAttributeResponse, err error) {
	if request == nil {
		request = NewModifyDirectConnectGatewayAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyDirectConnectGatewayAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyFlowLogAttributeRequest() (request *ModifyFlowLogAttributeRequest) {
	request = &ModifyFlowLogAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyFlowLogAttribute")

	return
}

func NewModifyFlowLogAttributeResponse() (response *ModifyFlowLogAttributeResponse) {
	response = &ModifyFlowLogAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyFlowLogAttribute
// 本接口（ModifyFlowLogAttribute）用于修改流日志属性
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) ModifyFlowLogAttribute(request *ModifyFlowLogAttributeRequest) (response *ModifyFlowLogAttributeResponse, err error) {
	if request == nil {
		request = NewModifyFlowLogAttributeRequest()
	}

	response = NewModifyFlowLogAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyFlowLogAttribute
// 本接口（ModifyFlowLogAttribute）用于修改流日志属性
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) ModifyFlowLogAttributeWithContext(ctx context.Context, request *ModifyFlowLogAttributeRequest) (response *ModifyFlowLogAttributeResponse, err error) {
	if request == nil {
		request = NewModifyFlowLogAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyFlowLogAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyGatewayFlowQosRequest() (request *ModifyGatewayFlowQosRequest) {
	request = &ModifyGatewayFlowQosRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyGatewayFlowQos")

	return
}

func NewModifyGatewayFlowQosResponse() (response *ModifyGatewayFlowQosResponse) {
	response = &ModifyGatewayFlowQosResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyGatewayFlowQos
// 本接口（ModifyGatewayFlowQos）用于调整网关流控带宽。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) ModifyGatewayFlowQos(request *ModifyGatewayFlowQosRequest) (response *ModifyGatewayFlowQosResponse, err error) {
	if request == nil {
		request = NewModifyGatewayFlowQosRequest()
	}

	response = NewModifyGatewayFlowQosResponse()
	err = c.Send(request, response)
	return
}

// ModifyGatewayFlowQos
// 本接口（ModifyGatewayFlowQos）用于调整网关流控带宽。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) ModifyGatewayFlowQosWithContext(ctx context.Context, request *ModifyGatewayFlowQosRequest) (response *ModifyGatewayFlowQosResponse, err error) {
	if request == nil {
		request = NewModifyGatewayFlowQosRequest()
	}
	request.SetContext(ctx)

	response = NewModifyGatewayFlowQosResponse()
	err = c.Send(request, response)
	return
}

func NewModifyHaVipAttributeRequest() (request *ModifyHaVipAttributeRequest) {
	request = &ModifyHaVipAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyHaVipAttribute")

	return
}

func NewModifyHaVipAttributeResponse() (response *ModifyHaVipAttributeResponse) {
	response = &ModifyHaVipAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyHaVipAttribute
// 本接口（ModifyHaVipAttribute）用于修改高可用虚拟IP（HAVIP）属性
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyHaVipAttribute(request *ModifyHaVipAttributeRequest) (response *ModifyHaVipAttributeResponse, err error) {
	if request == nil {
		request = NewModifyHaVipAttributeRequest()
	}

	response = NewModifyHaVipAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyHaVipAttribute
// 本接口（ModifyHaVipAttribute）用于修改高可用虚拟IP（HAVIP）属性
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyHaVipAttributeWithContext(ctx context.Context, request *ModifyHaVipAttributeRequest) (response *ModifyHaVipAttributeResponse, err error) {
	if request == nil {
		request = NewModifyHaVipAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyHaVipAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyIp6AddressesBandwidthRequest() (request *ModifyIp6AddressesBandwidthRequest) {
	request = &ModifyIp6AddressesBandwidthRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyIp6AddressesBandwidth")

	return
}

func NewModifyIp6AddressesBandwidthResponse() (response *ModifyIp6AddressesBandwidthResponse) {
	response = &ModifyIp6AddressesBandwidthResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyIp6AddressesBandwidth
// 该接口用于修改IPV6地址访问internet的带宽
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDADDRESSIDSTATE_INARREARS = "InvalidAddressIdState.InArrears"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTFOUND = "InvalidParameterValue.AddressIpNotFound"
//	INVALIDPARAMETERVALUE_INVALIDIPV6 = "InvalidParameterValue.InvalidIpv6"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	UNSUPPORTEDOPERATION_ADDRESSIPINARREAR = "UnsupportedOperation.AddressIpInArrear"
//	UNSUPPORTEDOPERATION_ADDRESSIPINTERNETCHARGETYPENOTPERMIT = "UnsupportedOperation.AddressIpInternetChargeTypeNotPermit"
//	UNSUPPORTEDOPERATION_ADDRESSIPNOTSUPPORTINSTANCE = "UnsupportedOperation.AddressIpNotSupportInstance"
//	UNSUPPORTEDOPERATION_ADDRESSIPSTATUSNOTPERMIT = "UnsupportedOperation.AddressIpStatusNotPermit"
func (c *Client) ModifyIp6AddressesBandwidth(request *ModifyIp6AddressesBandwidthRequest) (response *ModifyIp6AddressesBandwidthResponse, err error) {
	if request == nil {
		request = NewModifyIp6AddressesBandwidthRequest()
	}

	response = NewModifyIp6AddressesBandwidthResponse()
	err = c.Send(request, response)
	return
}

// ModifyIp6AddressesBandwidth
// 该接口用于修改IPV6地址访问internet的带宽
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_NOTSUPPORTED = "InvalidAccount.NotSupported"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDADDRESSIDSTATE_INARREARS = "InvalidAddressIdState.InArrears"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTFOUND = "InvalidParameterValue.AddressIpNotFound"
//	INVALIDPARAMETERVALUE_INVALIDIPV6 = "InvalidParameterValue.InvalidIpv6"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	UNSUPPORTEDOPERATION_ADDRESSIPINARREAR = "UnsupportedOperation.AddressIpInArrear"
//	UNSUPPORTEDOPERATION_ADDRESSIPINTERNETCHARGETYPENOTPERMIT = "UnsupportedOperation.AddressIpInternetChargeTypeNotPermit"
//	UNSUPPORTEDOPERATION_ADDRESSIPNOTSUPPORTINSTANCE = "UnsupportedOperation.AddressIpNotSupportInstance"
//	UNSUPPORTEDOPERATION_ADDRESSIPSTATUSNOTPERMIT = "UnsupportedOperation.AddressIpStatusNotPermit"
func (c *Client) ModifyIp6AddressesBandwidthWithContext(ctx context.Context, request *ModifyIp6AddressesBandwidthRequest) (response *ModifyIp6AddressesBandwidthResponse, err error) {
	if request == nil {
		request = NewModifyIp6AddressesBandwidthRequest()
	}
	request.SetContext(ctx)

	response = NewModifyIp6AddressesBandwidthResponse()
	err = c.Send(request, response)
	return
}

func NewModifyIp6RuleRequest() (request *ModifyIp6RuleRequest) {
	request = &ModifyIp6RuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyIp6Rule")

	return
}

func NewModifyIp6RuleResponse() (response *ModifyIp6RuleResponse) {
	response = &ModifyIp6RuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyIp6Rule
// 该接口用于修改IPV6转换规则，当前仅支持修改转换规则名称，IPV4地址和IPV4端口号
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_IPV6RULENOTCHANGE = "InvalidParameterValue.IPv6RuleNotChange"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) ModifyIp6Rule(request *ModifyIp6RuleRequest) (response *ModifyIp6RuleResponse, err error) {
	if request == nil {
		request = NewModifyIp6RuleRequest()
	}

	response = NewModifyIp6RuleResponse()
	err = c.Send(request, response)
	return
}

// ModifyIp6Rule
// 该接口用于修改IPV6转换规则，当前仅支持修改转换规则名称，IPV4地址和IPV4端口号
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_IPV6RULENOTCHANGE = "InvalidParameterValue.IPv6RuleNotChange"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
func (c *Client) ModifyIp6RuleWithContext(ctx context.Context, request *ModifyIp6RuleRequest) (response *ModifyIp6RuleResponse, err error) {
	if request == nil {
		request = NewModifyIp6RuleRequest()
	}
	request.SetContext(ctx)

	response = NewModifyIp6RuleResponse()
	err = c.Send(request, response)
	return
}

func NewModifyIp6TranslatorRequest() (request *ModifyIp6TranslatorRequest) {
	request = &ModifyIp6TranslatorRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyIp6Translator")

	return
}

func NewModifyIp6TranslatorResponse() (response *ModifyIp6TranslatorResponse) {
	response = &ModifyIp6TranslatorResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyIp6Translator
// 该接口用于修改IP6转换实例属性，当前仅支持修改实例名称。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) ModifyIp6Translator(request *ModifyIp6TranslatorRequest) (response *ModifyIp6TranslatorResponse, err error) {
	if request == nil {
		request = NewModifyIp6TranslatorRequest()
	}

	response = NewModifyIp6TranslatorResponse()
	err = c.Send(request, response)
	return
}

// ModifyIp6Translator
// 该接口用于修改IP6转换实例属性，当前仅支持修改实例名称。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) ModifyIp6TranslatorWithContext(ctx context.Context, request *ModifyIp6TranslatorRequest) (response *ModifyIp6TranslatorResponse, err error) {
	if request == nil {
		request = NewModifyIp6TranslatorRequest()
	}
	request.SetContext(ctx)

	response = NewModifyIp6TranslatorResponse()
	err = c.Send(request, response)
	return
}

func NewModifyIpv6AddressesAttributeRequest() (request *ModifyIpv6AddressesAttributeRequest) {
	request = &ModifyIpv6AddressesAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyIpv6AddressesAttribute")

	return
}

func NewModifyIpv6AddressesAttributeResponse() (response *ModifyIpv6AddressesAttributeResponse) {
	response = &ModifyIpv6AddressesAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyIpv6AddressesAttribute
// 本接口（ModifyIpv6AddressesAttribute）用于修改弹性网卡内网IPv6地址属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ATTACHMENTNOTFOUND = "UnsupportedOperation.AttachmentNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) ModifyIpv6AddressesAttribute(request *ModifyIpv6AddressesAttributeRequest) (response *ModifyIpv6AddressesAttributeResponse, err error) {
	if request == nil {
		request = NewModifyIpv6AddressesAttributeRequest()
	}

	response = NewModifyIpv6AddressesAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyIpv6AddressesAttribute
// 本接口（ModifyIpv6AddressesAttribute）用于修改弹性网卡内网IPv6地址属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ATTACHMENTNOTFOUND = "UnsupportedOperation.AttachmentNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) ModifyIpv6AddressesAttributeWithContext(ctx context.Context, request *ModifyIpv6AddressesAttributeRequest) (response *ModifyIpv6AddressesAttributeResponse, err error) {
	if request == nil {
		request = NewModifyIpv6AddressesAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyIpv6AddressesAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyLocalGatewayRequest() (request *ModifyLocalGatewayRequest) {
	request = &ModifyLocalGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyLocalGateway")

	return
}

func NewModifyLocalGatewayResponse() (response *ModifyLocalGatewayResponse) {
	response = &ModifyLocalGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyLocalGateway
// 该接口用于修改CDC的本地网关。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyLocalGateway(request *ModifyLocalGatewayRequest) (response *ModifyLocalGatewayResponse, err error) {
	if request == nil {
		request = NewModifyLocalGatewayRequest()
	}

	response = NewModifyLocalGatewayResponse()
	err = c.Send(request, response)
	return
}

// ModifyLocalGateway
// 该接口用于修改CDC的本地网关。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyLocalGatewayWithContext(ctx context.Context, request *ModifyLocalGatewayRequest) (response *ModifyLocalGatewayResponse, err error) {
	if request == nil {
		request = NewModifyLocalGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewModifyLocalGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNatGatewayAttributeRequest() (request *ModifyNatGatewayAttributeRequest) {
	request = &ModifyNatGatewayAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyNatGatewayAttribute")

	return
}

func NewModifyNatGatewayAttributeResponse() (response *ModifyNatGatewayAttributeResponse) {
	response = &ModifyNatGatewayAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNatGatewayAttribute
// 本接口（ModifyNatGatewayAttribute）用于修改NAT网关的属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyNatGatewayAttribute(request *ModifyNatGatewayAttributeRequest) (response *ModifyNatGatewayAttributeResponse, err error) {
	if request == nil {
		request = NewModifyNatGatewayAttributeRequest()
	}

	response = NewModifyNatGatewayAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyNatGatewayAttribute
// 本接口（ModifyNatGatewayAttribute）用于修改NAT网关的属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyNatGatewayAttributeWithContext(ctx context.Context, request *ModifyNatGatewayAttributeRequest) (response *ModifyNatGatewayAttributeResponse, err error) {
	if request == nil {
		request = NewModifyNatGatewayAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNatGatewayAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNatGatewayDestinationIpPortTranslationNatRuleRequest() (request *ModifyNatGatewayDestinationIpPortTranslationNatRuleRequest) {
	request = &ModifyNatGatewayDestinationIpPortTranslationNatRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyNatGatewayDestinationIpPortTranslationNatRule")

	return
}

func NewModifyNatGatewayDestinationIpPortTranslationNatRuleResponse() (response *ModifyNatGatewayDestinationIpPortTranslationNatRuleResponse) {
	response = &ModifyNatGatewayDestinationIpPortTranslationNatRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNatGatewayDestinationIpPortTranslationNatRule
// 本接口（ModifyNatGatewayDestinationIpPortTranslationNatRule）用于修改NAT网关端口转发规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyNatGatewayDestinationIpPortTranslationNatRule(request *ModifyNatGatewayDestinationIpPortTranslationNatRuleRequest) (response *ModifyNatGatewayDestinationIpPortTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewModifyNatGatewayDestinationIpPortTranslationNatRuleRequest()
	}

	response = NewModifyNatGatewayDestinationIpPortTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

// ModifyNatGatewayDestinationIpPortTranslationNatRule
// 本接口（ModifyNatGatewayDestinationIpPortTranslationNatRule）用于修改NAT网关端口转发规则。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyNatGatewayDestinationIpPortTranslationNatRuleWithContext(ctx context.Context, request *ModifyNatGatewayDestinationIpPortTranslationNatRuleRequest) (response *ModifyNatGatewayDestinationIpPortTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewModifyNatGatewayDestinationIpPortTranslationNatRuleRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNatGatewayDestinationIpPortTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNatGatewaySourceIpTranslationNatRuleRequest() (request *ModifyNatGatewaySourceIpTranslationNatRuleRequest) {
	request = &ModifyNatGatewaySourceIpTranslationNatRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyNatGatewaySourceIpTranslationNatRule")

	return
}

func NewModifyNatGatewaySourceIpTranslationNatRuleResponse() (response *ModifyNatGatewaySourceIpTranslationNatRuleResponse) {
	response = &ModifyNatGatewaySourceIpTranslationNatRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNatGatewaySourceIpTranslationNatRule
// 本接口（ModifyNatGatewaySourceIpTranslationNatRule）用于修改NAT网关SNAT转发规则。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION_UNBINDEIP = "UnsupportedOperation.UnbindEIP"
func (c *Client) ModifyNatGatewaySourceIpTranslationNatRule(request *ModifyNatGatewaySourceIpTranslationNatRuleRequest) (response *ModifyNatGatewaySourceIpTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewModifyNatGatewaySourceIpTranslationNatRuleRequest()
	}

	response = NewModifyNatGatewaySourceIpTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

// ModifyNatGatewaySourceIpTranslationNatRule
// 本接口（ModifyNatGatewaySourceIpTranslationNatRule）用于修改NAT网关SNAT转发规则。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION_UNBINDEIP = "UnsupportedOperation.UnbindEIP"
func (c *Client) ModifyNatGatewaySourceIpTranslationNatRuleWithContext(ctx context.Context, request *ModifyNatGatewaySourceIpTranslationNatRuleRequest) (response *ModifyNatGatewaySourceIpTranslationNatRuleResponse, err error) {
	if request == nil {
		request = NewModifyNatGatewaySourceIpTranslationNatRuleRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNatGatewaySourceIpTranslationNatRuleResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNetDetectRequest() (request *ModifyNetDetectRequest) {
	request = &ModifyNetDetectRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyNetDetect")

	return
}

func NewModifyNetDetectResponse() (response *ModifyNetDetectResponse) {
	response = &ModifyNetDetectResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNetDetect
// 本接口(ModifyNetDetect)用于修改网络探测参数。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_NEXTHOPMISMATCH = "InvalidParameter.NextHopMismatch"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_NETDETECTSAMEIP = "InvalidParameterValue.NetDetectSameIp"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
func (c *Client) ModifyNetDetect(request *ModifyNetDetectRequest) (response *ModifyNetDetectResponse, err error) {
	if request == nil {
		request = NewModifyNetDetectRequest()
	}

	response = NewModifyNetDetectResponse()
	err = c.Send(request, response)
	return
}

// ModifyNetDetect
// 本接口(ModifyNetDetect)用于修改网络探测参数。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_NEXTHOPMISMATCH = "InvalidParameter.NextHopMismatch"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_NETDETECTSAMEIP = "InvalidParameterValue.NetDetectSameIp"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_ECMPWITHUSERROUTE = "UnsupportedOperation.EcmpWithUserRoute"
func (c *Client) ModifyNetDetectWithContext(ctx context.Context, request *ModifyNetDetectRequest) (response *ModifyNetDetectResponse, err error) {
	if request == nil {
		request = NewModifyNetDetectRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNetDetectResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNetworkAclAttributeRequest() (request *ModifyNetworkAclAttributeRequest) {
	request = &ModifyNetworkAclAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyNetworkAclAttribute")

	return
}

func NewModifyNetworkAclAttributeResponse() (response *ModifyNetworkAclAttributeResponse) {
	response = &ModifyNetworkAclAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNetworkAclAttribute
// 本接口（ModifyNetworkAclAttribute）用于修改网络ACL属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyNetworkAclAttribute(request *ModifyNetworkAclAttributeRequest) (response *ModifyNetworkAclAttributeResponse, err error) {
	if request == nil {
		request = NewModifyNetworkAclAttributeRequest()
	}

	response = NewModifyNetworkAclAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyNetworkAclAttribute
// 本接口（ModifyNetworkAclAttribute）用于修改网络ACL属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyNetworkAclAttributeWithContext(ctx context.Context, request *ModifyNetworkAclAttributeRequest) (response *ModifyNetworkAclAttributeResponse, err error) {
	if request == nil {
		request = NewModifyNetworkAclAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNetworkAclAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNetworkAclEntriesRequest() (request *ModifyNetworkAclEntriesRequest) {
	request = &ModifyNetworkAclEntriesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyNetworkAclEntries")

	return
}

func NewModifyNetworkAclEntriesResponse() (response *ModifyNetworkAclEntriesResponse) {
	response = &ModifyNetworkAclEntriesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNetworkAclEntries
// 本接口（ModifyNetworkAclEntries）用于修改（包括添加和删除）网络ACL的入站规则和出站规则。在NetworkAclEntrySet参数中：
//
// * 若同时传入入站规则和出站规则，则重置原有的入站规则和出站规则，并分别导入传入的规则。
//
// * 若仅传入入站规则，则仅重置原有的入站规则，并导入传入的规则，不影响原有的出站规则（若仅传入出站规则，处理方式类似入站方向）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyNetworkAclEntries(request *ModifyNetworkAclEntriesRequest) (response *ModifyNetworkAclEntriesResponse, err error) {
	if request == nil {
		request = NewModifyNetworkAclEntriesRequest()
	}

	response = NewModifyNetworkAclEntriesResponse()
	err = c.Send(request, response)
	return
}

// ModifyNetworkAclEntries
// 本接口（ModifyNetworkAclEntries）用于修改（包括添加和删除）网络ACL的入站规则和出站规则。在NetworkAclEntrySet参数中：
//
// * 若同时传入入站规则和出站规则，则重置原有的入站规则和出站规则，并分别导入传入的规则。
//
// * 若仅传入入站规则，则仅重置原有的入站规则，并导入传入的规则，不影响原有的出站规则（若仅传入出站规则，处理方式类似入站方向）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyNetworkAclEntriesWithContext(ctx context.Context, request *ModifyNetworkAclEntriesRequest) (response *ModifyNetworkAclEntriesResponse, err error) {
	if request == nil {
		request = NewModifyNetworkAclEntriesRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNetworkAclEntriesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNetworkInterfaceAttributeRequest() (request *ModifyNetworkInterfaceAttributeRequest) {
	request = &ModifyNetworkInterfaceAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyNetworkInterfaceAttribute")

	return
}

func NewModifyNetworkInterfaceAttributeResponse() (response *ModifyNetworkInterfaceAttributeResponse) {
	response = &ModifyNetworkInterfaceAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNetworkInterfaceAttribute
// 本接口（ModifyNetworkInterfaceAttribute）用于修改弹性网卡属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyNetworkInterfaceAttribute(request *ModifyNetworkInterfaceAttributeRequest) (response *ModifyNetworkInterfaceAttributeResponse, err error) {
	if request == nil {
		request = NewModifyNetworkInterfaceAttributeRequest()
	}

	response = NewModifyNetworkInterfaceAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyNetworkInterfaceAttribute
// 本接口（ModifyNetworkInterfaceAttribute）用于修改弹性网卡属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyNetworkInterfaceAttributeWithContext(ctx context.Context, request *ModifyNetworkInterfaceAttributeRequest) (response *ModifyNetworkInterfaceAttributeResponse, err error) {
	if request == nil {
		request = NewModifyNetworkInterfaceAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNetworkInterfaceAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNetworkInterfaceQosRequest() (request *ModifyNetworkInterfaceQosRequest) {
	request = &ModifyNetworkInterfaceQosRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyNetworkInterfaceQos")

	return
}

func NewModifyNetworkInterfaceQosResponse() (response *ModifyNetworkInterfaceQosResponse) {
	response = &ModifyNetworkInterfaceQosResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNetworkInterfaceQos
// 修改弹性网卡服务质量
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyNetworkInterfaceQos(request *ModifyNetworkInterfaceQosRequest) (response *ModifyNetworkInterfaceQosResponse, err error) {
	if request == nil {
		request = NewModifyNetworkInterfaceQosRequest()
	}

	response = NewModifyNetworkInterfaceQosResponse()
	err = c.Send(request, response)
	return
}

// ModifyNetworkInterfaceQos
// 修改弹性网卡服务质量
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyNetworkInterfaceQosWithContext(ctx context.Context, request *ModifyNetworkInterfaceQosRequest) (response *ModifyNetworkInterfaceQosResponse, err error) {
	if request == nil {
		request = NewModifyNetworkInterfaceQosRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNetworkInterfaceQosResponse()
	err = c.Send(request, response)
	return
}

func NewModifyPrivateIpAddressesAttributeRequest() (request *ModifyPrivateIpAddressesAttributeRequest) {
	request = &ModifyPrivateIpAddressesAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyPrivateIpAddressesAttribute")

	return
}

func NewModifyPrivateIpAddressesAttributeResponse() (response *ModifyPrivateIpAddressesAttributeResponse) {
	response = &ModifyPrivateIpAddressesAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyPrivateIpAddressesAttribute
// 本接口（ModifyPrivateIpAddressesAttribute）用于修改弹性网卡内网IP属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) ModifyPrivateIpAddressesAttribute(request *ModifyPrivateIpAddressesAttributeRequest) (response *ModifyPrivateIpAddressesAttributeResponse, err error) {
	if request == nil {
		request = NewModifyPrivateIpAddressesAttributeRequest()
	}

	response = NewModifyPrivateIpAddressesAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyPrivateIpAddressesAttribute
// 本接口（ModifyPrivateIpAddressesAttribute）用于修改弹性网卡内网IP属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) ModifyPrivateIpAddressesAttributeWithContext(ctx context.Context, request *ModifyPrivateIpAddressesAttributeRequest) (response *ModifyPrivateIpAddressesAttributeResponse, err error) {
	if request == nil {
		request = NewModifyPrivateIpAddressesAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyPrivateIpAddressesAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyRouteTableAttributeRequest() (request *ModifyRouteTableAttributeRequest) {
	request = &ModifyRouteTableAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyRouteTableAttribute")

	return
}

func NewModifyRouteTableAttributeResponse() (response *ModifyRouteTableAttributeResponse) {
	response = &ModifyRouteTableAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyRouteTableAttribute
// 本接口（ModifyRouteTableAttribute）用于修改路由表（RouteTable）属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyRouteTableAttribute(request *ModifyRouteTableAttributeRequest) (response *ModifyRouteTableAttributeResponse, err error) {
	if request == nil {
		request = NewModifyRouteTableAttributeRequest()
	}

	response = NewModifyRouteTableAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyRouteTableAttribute
// 本接口（ModifyRouteTableAttribute）用于修改路由表（RouteTable）属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyRouteTableAttributeWithContext(ctx context.Context, request *ModifyRouteTableAttributeRequest) (response *ModifyRouteTableAttributeResponse, err error) {
	if request == nil {
		request = NewModifyRouteTableAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyRouteTableAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifySecurityGroupAttributeRequest() (request *ModifySecurityGroupAttributeRequest) {
	request = &ModifySecurityGroupAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifySecurityGroupAttribute")

	return
}

func NewModifySecurityGroupAttributeResponse() (response *ModifySecurityGroupAttributeResponse) {
	response = &ModifySecurityGroupAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifySecurityGroupAttribute
// 本接口（ModifySecurityGroupAttribute）用于修改安全组（SecurityGroupPolicy）属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifySecurityGroupAttribute(request *ModifySecurityGroupAttributeRequest) (response *ModifySecurityGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifySecurityGroupAttributeRequest()
	}

	response = NewModifySecurityGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifySecurityGroupAttribute
// 本接口（ModifySecurityGroupAttribute）用于修改安全组（SecurityGroupPolicy）属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifySecurityGroupAttributeWithContext(ctx context.Context, request *ModifySecurityGroupAttributeRequest) (response *ModifySecurityGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifySecurityGroupAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifySecurityGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifySecurityGroupPoliciesRequest() (request *ModifySecurityGroupPoliciesRequest) {
	request = &ModifySecurityGroupPoliciesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifySecurityGroupPolicies")

	return
}

func NewModifySecurityGroupPoliciesResponse() (response *ModifySecurityGroupPoliciesResponse) {
	response = &ModifySecurityGroupPoliciesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifySecurityGroupPolicies
// 本接口（ModifySecurityGroupPolicies）用于重置安全组出站和入站规则（SecurityGroupPolicy）。
//
// <ul>
//
// <li>该接口不支持自定义索引 PolicyIndex。</li>
//
// <li>在 SecurityGroupPolicySet 参数中：<ul>
//
//	<li> 如果指定 SecurityGroupPolicySet.Version 为0, 表示清空所有规则，并忽略 Egress 和 Ingress。</li>
//
//	<li> 如果指定 SecurityGroupPolicySet.Version 不为0, 在添加出站和入站规则（Egress 和 Ingress）时：<ul>
//
//		<li>Protocol 字段支持输入 TCP, UDP, ICMP, ICMPV6, GRE, ALL。</li>
//
//		<li>CidrBlock 字段允许输入符合 cidr 格式标准的任意字符串。(展开)在基础网络中，如果 CidrBlock 包含您的账户内的云服务器之外的设备在腾讯云的内网 IP，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。</li>
//
//		<li>Ipv6CidrBlock 字段允许输入符合 IPv6 cidr 格式标准的任意字符串。(展开)在基础网络中，如果Ipv6CidrBlock 包含您的账户内的云服务器之外的设备在腾讯云的内网 IPv6，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。</li>
//
//		<li>SecurityGroupId 字段允许输入与待修改的安全组位于相同项目中的安全组 ID，包括这个安全组 ID 本身，代表安全组下所有云服务器的内网 IP。使用这个字段时，这条规则用来匹配网络报文的过程中会随着被使用的这个ID所关联的云服务器变化而变化，不需要重新修改。</li>
//
//		<li>Port 字段允许输入一个单独端口号，或者用减号分隔的两个端口号代表端口范围，例如80或8000-8010。只有当 Protocol 字段是 TCP 或 UDP 时，Port 字段才被接受。</li>
//
//		<li>Action 字段只允许输入 ACCEPT 或 DROP。</li>
//
//		<li>CidrBlock, Ipv6CidrBlock, SecurityGroupId, AddressTemplate 四者是排他关系，不允许同时输入，Protocol + Port 和 ServiceTemplate 二者是排他关系，不允许同时输入。</li>
//
// </ul></li></ul></li>
//
// </ul>
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_DUPLICATEPOLICY = "UnsupportedOperation.DuplicatePolicy"
func (c *Client) ModifySecurityGroupPolicies(request *ModifySecurityGroupPoliciesRequest) (response *ModifySecurityGroupPoliciesResponse, err error) {
	if request == nil {
		request = NewModifySecurityGroupPoliciesRequest()
	}

	response = NewModifySecurityGroupPoliciesResponse()
	err = c.Send(request, response)
	return
}

// ModifySecurityGroupPolicies
// 本接口（ModifySecurityGroupPolicies）用于重置安全组出站和入站规则（SecurityGroupPolicy）。
//
// <ul>
//
// <li>该接口不支持自定义索引 PolicyIndex。</li>
//
// <li>在 SecurityGroupPolicySet 参数中：<ul>
//
//	<li> 如果指定 SecurityGroupPolicySet.Version 为0, 表示清空所有规则，并忽略 Egress 和 Ingress。</li>
//
//	<li> 如果指定 SecurityGroupPolicySet.Version 不为0, 在添加出站和入站规则（Egress 和 Ingress）时：<ul>
//
//		<li>Protocol 字段支持输入 TCP, UDP, ICMP, ICMPV6, GRE, ALL。</li>
//
//		<li>CidrBlock 字段允许输入符合 cidr 格式标准的任意字符串。(展开)在基础网络中，如果 CidrBlock 包含您的账户内的云服务器之外的设备在腾讯云的内网 IP，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。</li>
//
//		<li>Ipv6CidrBlock 字段允许输入符合 IPv6 cidr 格式标准的任意字符串。(展开)在基础网络中，如果Ipv6CidrBlock 包含您的账户内的云服务器之外的设备在腾讯云的内网 IPv6，并不代表此规则允许您访问这些设备，租户之间网络隔离规则优先于安全组中的内网规则。</li>
//
//		<li>SecurityGroupId 字段允许输入与待修改的安全组位于相同项目中的安全组 ID，包括这个安全组 ID 本身，代表安全组下所有云服务器的内网 IP。使用这个字段时，这条规则用来匹配网络报文的过程中会随着被使用的这个ID所关联的云服务器变化而变化，不需要重新修改。</li>
//
//		<li>Port 字段允许输入一个单独端口号，或者用减号分隔的两个端口号代表端口范围，例如80或8000-8010。只有当 Protocol 字段是 TCP 或 UDP 时，Port 字段才被接受。</li>
//
//		<li>Action 字段只允许输入 ACCEPT 或 DROP。</li>
//
//		<li>CidrBlock, Ipv6CidrBlock, SecurityGroupId, AddressTemplate 四者是排他关系，不允许同时输入，Protocol + Port 和 ServiceTemplate 二者是排他关系，不允许同时输入。</li>
//
// </ul></li></ul></li>
//
// </ul>
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_DUPLICATEPOLICY = "UnsupportedOperation.DuplicatePolicy"
func (c *Client) ModifySecurityGroupPoliciesWithContext(ctx context.Context, request *ModifySecurityGroupPoliciesRequest) (response *ModifySecurityGroupPoliciesResponse, err error) {
	if request == nil {
		request = NewModifySecurityGroupPoliciesRequest()
	}
	request.SetContext(ctx)

	response = NewModifySecurityGroupPoliciesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyServiceTemplateAttributeRequest() (request *ModifyServiceTemplateAttributeRequest) {
	request = &ModifyServiceTemplateAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyServiceTemplateAttribute")

	return
}

func NewModifyServiceTemplateAttributeResponse() (response *ModifyServiceTemplateAttributeResponse) {
	response = &ModifyServiceTemplateAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyServiceTemplateAttribute
// 本接口（ModifyServiceTemplateAttribute）用于修改协议端口模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) ModifyServiceTemplateAttribute(request *ModifyServiceTemplateAttributeRequest) (response *ModifyServiceTemplateAttributeResponse, err error) {
	if request == nil {
		request = NewModifyServiceTemplateAttributeRequest()
	}

	response = NewModifyServiceTemplateAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyServiceTemplateAttribute
// 本接口（ModifyServiceTemplateAttribute）用于修改协议端口模板
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) ModifyServiceTemplateAttributeWithContext(ctx context.Context, request *ModifyServiceTemplateAttributeRequest) (response *ModifyServiceTemplateAttributeResponse, err error) {
	if request == nil {
		request = NewModifyServiceTemplateAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyServiceTemplateAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyServiceTemplateGroupAttributeRequest() (request *ModifyServiceTemplateGroupAttributeRequest) {
	request = &ModifyServiceTemplateGroupAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyServiceTemplateGroupAttribute")

	return
}

func NewModifyServiceTemplateGroupAttributeResponse() (response *ModifyServiceTemplateGroupAttributeResponse) {
	response = &ModifyServiceTemplateGroupAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyServiceTemplateGroupAttribute
// 本接口（ModifyServiceTemplateGroupAttribute）用于修改协议端口模板集合。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyServiceTemplateGroupAttribute(request *ModifyServiceTemplateGroupAttributeRequest) (response *ModifyServiceTemplateGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifyServiceTemplateGroupAttributeRequest()
	}

	response = NewModifyServiceTemplateGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyServiceTemplateGroupAttribute
// 本接口（ModifyServiceTemplateGroupAttribute）用于修改协议端口模板集合。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyServiceTemplateGroupAttributeWithContext(ctx context.Context, request *ModifyServiceTemplateGroupAttributeRequest) (response *ModifyServiceTemplateGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifyServiceTemplateGroupAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyServiceTemplateGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifySubnetAttributeRequest() (request *ModifySubnetAttributeRequest) {
	request = &ModifySubnetAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifySubnetAttribute")

	return
}

func NewModifySubnetAttributeResponse() (response *ModifySubnetAttributeResponse) {
	response = &ModifySubnetAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifySubnetAttribute
// 本接口（ModifySubnetAttribute）用于修改子网属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifySubnetAttribute(request *ModifySubnetAttributeRequest) (response *ModifySubnetAttributeResponse, err error) {
	if request == nil {
		request = NewModifySubnetAttributeRequest()
	}

	response = NewModifySubnetAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifySubnetAttribute
// 本接口（ModifySubnetAttribute）用于修改子网属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifySubnetAttributeWithContext(ctx context.Context, request *ModifySubnetAttributeRequest) (response *ModifySubnetAttributeResponse, err error) {
	if request == nil {
		request = NewModifySubnetAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifySubnetAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyTemplateMemberRequest() (request *ModifyTemplateMemberRequest) {
	request = &ModifyTemplateMemberRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyTemplateMember")

	return
}

func NewModifyTemplateMemberResponse() (response *ModifyTemplateMemberResponse) {
	response = &ModifyTemplateMemberResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyTemplateMember
// 修改模版对象中的IP地址、协议端口、IP地址组、协议端口组。当前仅支持北京、泰国、北美地域请求。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyTemplateMember(request *ModifyTemplateMemberRequest) (response *ModifyTemplateMemberResponse, err error) {
	if request == nil {
		request = NewModifyTemplateMemberRequest()
	}

	response = NewModifyTemplateMemberResponse()
	err = c.Send(request, response)
	return
}

// ModifyTemplateMember
// 修改模版对象中的IP地址、协议端口、IP地址组、协议端口组。当前仅支持北京、泰国、北美地域请求。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyTemplateMemberWithContext(ctx context.Context, request *ModifyTemplateMemberRequest) (response *ModifyTemplateMemberResponse, err error) {
	if request == nil {
		request = NewModifyTemplateMemberRequest()
	}
	request.SetContext(ctx)

	response = NewModifyTemplateMemberResponse()
	err = c.Send(request, response)
	return
}

func NewModifyVpcAttributeRequest() (request *ModifyVpcAttributeRequest) {
	request = &ModifyVpcAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyVpcAttribute")

	return
}

func NewModifyVpcAttributeResponse() (response *ModifyVpcAttributeResponse) {
	response = &ModifyVpcAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyVpcAttribute
// 本接口（ModifyVpcAttribute）用于修改私有网络（VPC）的相关属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyVpcAttribute(request *ModifyVpcAttributeRequest) (response *ModifyVpcAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpcAttributeRequest()
	}

	response = NewModifyVpcAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyVpcAttribute
// 本接口（ModifyVpcAttribute）用于修改私有网络（VPC）的相关属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyVpcAttributeWithContext(ctx context.Context, request *ModifyVpcAttributeRequest) (response *ModifyVpcAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpcAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyVpcAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyVpcEndPointAttributeRequest() (request *ModifyVpcEndPointAttributeRequest) {
	request = &ModifyVpcEndPointAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyVpcEndPointAttribute")

	return
}

func NewModifyVpcEndPointAttributeResponse() (response *ModifyVpcEndPointAttributeResponse) {
	response = &ModifyVpcEndPointAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyVpcEndPointAttribute
// 修改终端节点属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCENOTFOUND_SVCNOTEXIST = "ResourceNotFound.SvcNotExist"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_SPECIALENDPOINTSERVICE = "UnsupportedOperation.SpecialEndPointService"
func (c *Client) ModifyVpcEndPointAttribute(request *ModifyVpcEndPointAttributeRequest) (response *ModifyVpcEndPointAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpcEndPointAttributeRequest()
	}

	response = NewModifyVpcEndPointAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyVpcEndPointAttribute
// 修改终端节点属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCENOTFOUND_SVCNOTEXIST = "ResourceNotFound.SvcNotExist"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_SPECIALENDPOINTSERVICE = "UnsupportedOperation.SpecialEndPointService"
func (c *Client) ModifyVpcEndPointAttributeWithContext(ctx context.Context, request *ModifyVpcEndPointAttributeRequest) (response *ModifyVpcEndPointAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpcEndPointAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyVpcEndPointAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyVpcEndPointServiceAttributeRequest() (request *ModifyVpcEndPointServiceAttributeRequest) {
	request = &ModifyVpcEndPointServiceAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyVpcEndPointServiceAttribute")

	return
}

func NewModifyVpcEndPointServiceAttributeResponse() (response *ModifyVpcEndPointServiceAttributeResponse) {
	response = &ModifyVpcEndPointServiceAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyVpcEndPointServiceAttribute
// 本接口（ModifyVpcEndPointServiceAttribute）用于修改终端节点服务属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) ModifyVpcEndPointServiceAttribute(request *ModifyVpcEndPointServiceAttributeRequest) (response *ModifyVpcEndPointServiceAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpcEndPointServiceAttributeRequest()
	}

	response = NewModifyVpcEndPointServiceAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyVpcEndPointServiceAttribute
// 本接口（ModifyVpcEndPointServiceAttribute）用于修改终端节点服务属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) ModifyVpcEndPointServiceAttributeWithContext(ctx context.Context, request *ModifyVpcEndPointServiceAttributeRequest) (response *ModifyVpcEndPointServiceAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpcEndPointServiceAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyVpcEndPointServiceAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyVpcEndPointServiceWhiteListRequest() (request *ModifyVpcEndPointServiceWhiteListRequest) {
	request = &ModifyVpcEndPointServiceWhiteListRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyVpcEndPointServiceWhiteList")

	return
}

func NewModifyVpcEndPointServiceWhiteListResponse() (response *ModifyVpcEndPointServiceWhiteListResponse) {
	response = &ModifyVpcEndPointServiceWhiteListResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyVpcEndPointServiceWhiteList
// 修改终端节点服务白名单属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
func (c *Client) ModifyVpcEndPointServiceWhiteList(request *ModifyVpcEndPointServiceWhiteListRequest) (response *ModifyVpcEndPointServiceWhiteListResponse, err error) {
	if request == nil {
		request = NewModifyVpcEndPointServiceWhiteListRequest()
	}

	response = NewModifyVpcEndPointServiceWhiteListResponse()
	err = c.Send(request, response)
	return
}

// ModifyVpcEndPointServiceWhiteList
// 修改终端节点服务白名单属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_UINNOTFOUND = "UnsupportedOperation.UinNotFound"
func (c *Client) ModifyVpcEndPointServiceWhiteListWithContext(ctx context.Context, request *ModifyVpcEndPointServiceWhiteListRequest) (response *ModifyVpcEndPointServiceWhiteListResponse, err error) {
	if request == nil {
		request = NewModifyVpcEndPointServiceWhiteListRequest()
	}
	request.SetContext(ctx)

	response = NewModifyVpcEndPointServiceWhiteListResponse()
	err = c.Send(request, response)
	return
}

func NewModifyVpnConnectionAttributeRequest() (request *ModifyVpnConnectionAttributeRequest) {
	request = &ModifyVpnConnectionAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyVpnConnectionAttribute")

	return
}

func NewModifyVpnConnectionAttributeResponse() (response *ModifyVpnConnectionAttributeResponse) {
	response = &ModifyVpnConnectionAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyVpnConnectionAttribute
// 本接口（ModifyVpnConnectionAttribute）用于修改VPN通道。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_VPNCONNCIDRCONFLICT = "InvalidParameterValue.VpnConnCidrConflict"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyVpnConnectionAttribute(request *ModifyVpnConnectionAttributeRequest) (response *ModifyVpnConnectionAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpnConnectionAttributeRequest()
	}

	response = NewModifyVpnConnectionAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyVpnConnectionAttribute
// 本接口（ModifyVpnConnectionAttribute）用于修改VPN通道。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_VPNCONNCIDRCONFLICT = "InvalidParameterValue.VpnConnCidrConflict"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyVpnConnectionAttributeWithContext(ctx context.Context, request *ModifyVpnConnectionAttributeRequest) (response *ModifyVpnConnectionAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpnConnectionAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyVpnConnectionAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyVpnGatewayAttributeRequest() (request *ModifyVpnGatewayAttributeRequest) {
	request = &ModifyVpnGatewayAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyVpnGatewayAttribute")

	return
}

func NewModifyVpnGatewayAttributeResponse() (response *ModifyVpnGatewayAttributeResponse) {
	response = &ModifyVpnGatewayAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyVpnGatewayAttribute
// 本接口（ModifyVpnGatewayAttribute）用于修改VPN网关属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyVpnGatewayAttribute(request *ModifyVpnGatewayAttributeRequest) (response *ModifyVpnGatewayAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpnGatewayAttributeRequest()
	}

	response = NewModifyVpnGatewayAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyVpnGatewayAttribute
// 本接口（ModifyVpnGatewayAttribute）用于修改VPN网关属性。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyVpnGatewayAttributeWithContext(ctx context.Context, request *ModifyVpnGatewayAttributeRequest) (response *ModifyVpnGatewayAttributeResponse, err error) {
	if request == nil {
		request = NewModifyVpnGatewayAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyVpnGatewayAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyVpnGatewayCcnRoutesRequest() (request *ModifyVpnGatewayCcnRoutesRequest) {
	request = &ModifyVpnGatewayCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyVpnGatewayCcnRoutes")

	return
}

func NewModifyVpnGatewayCcnRoutesResponse() (response *ModifyVpnGatewayCcnRoutesResponse) {
	response = &ModifyVpnGatewayCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyVpnGatewayCcnRoutes
// 本接口（ModifyVpnGatewayCcnRoutes）用于修改VPN网关云联网路由
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyVpnGatewayCcnRoutes(request *ModifyVpnGatewayCcnRoutesRequest) (response *ModifyVpnGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewModifyVpnGatewayCcnRoutesRequest()
	}

	response = NewModifyVpnGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// ModifyVpnGatewayCcnRoutes
// 本接口（ModifyVpnGatewayCcnRoutes）用于修改VPN网关云联网路由
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ModifyVpnGatewayCcnRoutesWithContext(ctx context.Context, request *ModifyVpnGatewayCcnRoutesRequest) (response *ModifyVpnGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewModifyVpnGatewayCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewModifyVpnGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyVpnGatewayRoutesRequest() (request *ModifyVpnGatewayRoutesRequest) {
	request = &ModifyVpnGatewayRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ModifyVpnGatewayRoutes")

	return
}

func NewModifyVpnGatewayRoutesResponse() (response *ModifyVpnGatewayRoutesResponse) {
	response = &ModifyVpnGatewayRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyVpnGatewayRoutes
// 修改VPN路由是否启用
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyVpnGatewayRoutes(request *ModifyVpnGatewayRoutesRequest) (response *ModifyVpnGatewayRoutesResponse, err error) {
	if request == nil {
		request = NewModifyVpnGatewayRoutesRequest()
	}

	response = NewModifyVpnGatewayRoutesResponse()
	err = c.Send(request, response)
	return
}

// ModifyVpnGatewayRoutes
// 修改VPN路由是否启用
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyVpnGatewayRoutesWithContext(ctx context.Context, request *ModifyVpnGatewayRoutesRequest) (response *ModifyVpnGatewayRoutesResponse, err error) {
	if request == nil {
		request = NewModifyVpnGatewayRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewModifyVpnGatewayRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewNotifyRoutesRequest() (request *NotifyRoutesRequest) {
	request = &NotifyRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "NotifyRoutes")

	return
}

func NewNotifyRoutesResponse() (response *NotifyRoutesResponse) {
	response = &NotifyRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// NotifyRoutes
// 路由表列表页操作增加“发布到云联网”，用于发布路由到云联网。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDROUTEID_NOTFOUND = "InvalidRouteId.NotFound"
//	INVALIDROUTETABLEID_MALFORMED = "InvalidRouteTableId.Malformed"
//	INVALIDROUTETABLEID_NOTFOUND = "InvalidRouteTableId.NotFound"
//	UNSUPPORTEDOPERATION_CCNNOTATTACHED = "UnsupportedOperation.CcnNotAttached"
//	UNSUPPORTEDOPERATION_NOTIFYCCN = "UnsupportedOperation.NotifyCcn"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) NotifyRoutes(request *NotifyRoutesRequest) (response *NotifyRoutesResponse, err error) {
	if request == nil {
		request = NewNotifyRoutesRequest()
	}

	response = NewNotifyRoutesResponse()
	err = c.Send(request, response)
	return
}

// NotifyRoutes
// 路由表列表页操作增加“发布到云联网”，用于发布路由到云联网。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDROUTEID_NOTFOUND = "InvalidRouteId.NotFound"
//	INVALIDROUTETABLEID_MALFORMED = "InvalidRouteTableId.Malformed"
//	INVALIDROUTETABLEID_NOTFOUND = "InvalidRouteTableId.NotFound"
//	UNSUPPORTEDOPERATION_CCNNOTATTACHED = "UnsupportedOperation.CcnNotAttached"
//	UNSUPPORTEDOPERATION_NOTIFYCCN = "UnsupportedOperation.NotifyCcn"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) NotifyRoutesWithContext(ctx context.Context, request *NotifyRoutesRequest) (response *NotifyRoutesResponse, err error) {
	if request == nil {
		request = NewNotifyRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewNotifyRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewRefreshDirectConnectGatewayRouteToNatGatewayRequest() (request *RefreshDirectConnectGatewayRouteToNatGatewayRequest) {
	request = &RefreshDirectConnectGatewayRouteToNatGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "RefreshDirectConnectGatewayRouteToNatGateway")

	return
}

func NewRefreshDirectConnectGatewayRouteToNatGatewayResponse() (response *RefreshDirectConnectGatewayRouteToNatGatewayResponse) {
	response = &RefreshDirectConnectGatewayRouteToNatGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RefreshDirectConnectGatewayRouteToNatGateway
// 刷新专线直连nat路由，更新nat到专线的路由表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) RefreshDirectConnectGatewayRouteToNatGateway(request *RefreshDirectConnectGatewayRouteToNatGatewayRequest) (response *RefreshDirectConnectGatewayRouteToNatGatewayResponse, err error) {
	if request == nil {
		request = NewRefreshDirectConnectGatewayRouteToNatGatewayRequest()
	}

	response = NewRefreshDirectConnectGatewayRouteToNatGatewayResponse()
	err = c.Send(request, response)
	return
}

// RefreshDirectConnectGatewayRouteToNatGateway
// 刷新专线直连nat路由，更新nat到专线的路由表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) RefreshDirectConnectGatewayRouteToNatGatewayWithContext(ctx context.Context, request *RefreshDirectConnectGatewayRouteToNatGatewayRequest) (response *RefreshDirectConnectGatewayRouteToNatGatewayResponse, err error) {
	if request == nil {
		request = NewRefreshDirectConnectGatewayRouteToNatGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewRefreshDirectConnectGatewayRouteToNatGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewRejectAttachCcnInstancesRequest() (request *RejectAttachCcnInstancesRequest) {
	request = &RejectAttachCcnInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "RejectAttachCcnInstances")

	return
}

func NewRejectAttachCcnInstancesResponse() (response *RejectAttachCcnInstancesResponse) {
	response = &RejectAttachCcnInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RejectAttachCcnInstances
// 本接口（RejectAttachCcnInstances）用于跨账号关联实例时，云联网所有者拒绝关联操作。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CCNNOTATTACHED = "UnsupportedOperation.CcnNotAttached"
//	UNSUPPORTEDOPERATION_NOTPENDINGCCNINSTANCE = "UnsupportedOperation.NotPendingCcnInstance"
func (c *Client) RejectAttachCcnInstances(request *RejectAttachCcnInstancesRequest) (response *RejectAttachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewRejectAttachCcnInstancesRequest()
	}

	response = NewRejectAttachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

// RejectAttachCcnInstances
// 本接口（RejectAttachCcnInstances）用于跨账号关联实例时，云联网所有者拒绝关联操作。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CCNNOTATTACHED = "UnsupportedOperation.CcnNotAttached"
//	UNSUPPORTEDOPERATION_NOTPENDINGCCNINSTANCE = "UnsupportedOperation.NotPendingCcnInstance"
func (c *Client) RejectAttachCcnInstancesWithContext(ctx context.Context, request *RejectAttachCcnInstancesRequest) (response *RejectAttachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewRejectAttachCcnInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewRejectAttachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewReleaseAddressesRequest() (request *ReleaseAddressesRequest) {
	request = &ReleaseAddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ReleaseAddresses")

	return
}

func NewReleaseAddressesResponse() (response *ReleaseAddressesResponse) {
	response = &ReleaseAddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ReleaseAddresses
// 本接口 (ReleaseAddresses) 用于释放一个或多个[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）。
//
// * 该操作不可逆，释放后 EIP 关联的 IP 地址将不再属于您的名下。
//
// * 只有状态为 UNBIND 的 EIP 才能进行释放操作。
//
// 可能返回的错误码:
//
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDADDRESSSTATE = "InvalidAddressState"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSINTERNETCHARGETYPECONFLICT = "InvalidParameterValue.AddressInternetChargeTypeConflict"
//	INVALIDPARAMETERVALUE_ADDRESSNOTEIP = "InvalidParameterValue.AddressNotEIP"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
func (c *Client) ReleaseAddresses(request *ReleaseAddressesRequest) (response *ReleaseAddressesResponse, err error) {
	if request == nil {
		request = NewReleaseAddressesRequest()
	}

	response = NewReleaseAddressesResponse()
	err = c.Send(request, response)
	return
}

// ReleaseAddresses
// 本接口 (ReleaseAddresses) 用于释放一个或多个[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）。
//
// * 该操作不可逆，释放后 EIP 关联的 IP 地址将不再属于您的名下。
//
// * 只有状态为 UNBIND 的 EIP 才能进行释放操作。
//
// 可能返回的错误码:
//
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
//	INVALIDADDRESSSTATE = "InvalidAddressState"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSINTERNETCHARGETYPECONFLICT = "InvalidParameterValue.AddressInternetChargeTypeConflict"
//	INVALIDPARAMETERVALUE_ADDRESSNOTEIP = "InvalidParameterValue.AddressNotEIP"
//	INVALIDPARAMETERVALUE_ADDRESSNOTFOUND = "InvalidParameterValue.AddressNotFound"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
func (c *Client) ReleaseAddressesWithContext(ctx context.Context, request *ReleaseAddressesRequest) (response *ReleaseAddressesResponse, err error) {
	if request == nil {
		request = NewReleaseAddressesRequest()
	}
	request.SetContext(ctx)

	response = NewReleaseAddressesResponse()
	err = c.Send(request, response)
	return
}

func NewReleaseIp6AddressesBandwidthRequest() (request *ReleaseIp6AddressesBandwidthRequest) {
	request = &ReleaseIp6AddressesBandwidthRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ReleaseIp6AddressesBandwidth")

	return
}

func NewReleaseIp6AddressesBandwidthResponse() (response *ReleaseIp6AddressesBandwidthResponse) {
	response = &ReleaseIp6AddressesBandwidthResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ReleaseIp6AddressesBandwidth
// 该接口用于给弹性公网IPv6地址释放带宽。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTFOUND = "InvalidParameterValue.AddressIpNotFound"
//	INVALIDPARAMETERVALUE_INVALIDIPV6 = "InvalidParameterValue.InvalidIpv6"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
func (c *Client) ReleaseIp6AddressesBandwidth(request *ReleaseIp6AddressesBandwidthRequest) (response *ReleaseIp6AddressesBandwidthResponse, err error) {
	if request == nil {
		request = NewReleaseIp6AddressesBandwidthRequest()
	}

	response = NewReleaseIp6AddressesBandwidthResponse()
	err = c.Send(request, response)
	return
}

// ReleaseIp6AddressesBandwidth
// 该接口用于给弹性公网IPv6地址释放带宽。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ADDRESSIDMALFORMED = "InvalidParameterValue.AddressIdMalformed"
//	INVALIDPARAMETERVALUE_ADDRESSIPNOTFOUND = "InvalidParameterValue.AddressIpNotFound"
//	INVALIDPARAMETERVALUE_INVALIDIPV6 = "InvalidParameterValue.InvalidIpv6"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
func (c *Client) ReleaseIp6AddressesBandwidthWithContext(ctx context.Context, request *ReleaseIp6AddressesBandwidthRequest) (response *ReleaseIp6AddressesBandwidthResponse, err error) {
	if request == nil {
		request = NewReleaseIp6AddressesBandwidthRequest()
	}
	request.SetContext(ctx)

	response = NewReleaseIp6AddressesBandwidthResponse()
	err = c.Send(request, response)
	return
}

func NewRemoveBandwidthPackageResourcesRequest() (request *RemoveBandwidthPackageResourcesRequest) {
	request = &RemoveBandwidthPackageResourcesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "RemoveBandwidthPackageResources")

	return
}

func NewRemoveBandwidthPackageResourcesResponse() (response *RemoveBandwidthPackageResourcesResponse) {
	response = &RemoveBandwidthPackageResourcesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RemoveBandwidthPackageResources
// 接口用于删除带宽包资源，包括[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)和[负载均衡](https://cloud.tencent.com/document/product/214/517)等
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_RESOURCEIDMALFORMED = "InvalidParameterValue.ResourceIdMalformed"
//	INVALIDPARAMETERVALUE_RESOURCENOTEXISTED = "InvalidParameterValue.ResourceNotExisted"
//	INVALIDPARAMETERVALUE_RESOURCENOTFOUND = "InvalidParameterValue.ResourceNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDRESOURCEPROTOCOL = "UnsupportedOperation.InvalidResourceProtocol"
func (c *Client) RemoveBandwidthPackageResources(request *RemoveBandwidthPackageResourcesRequest) (response *RemoveBandwidthPackageResourcesResponse, err error) {
	if request == nil {
		request = NewRemoveBandwidthPackageResourcesRequest()
	}

	response = NewRemoveBandwidthPackageResourcesResponse()
	err = c.Send(request, response)
	return
}

// RemoveBandwidthPackageResources
// 接口用于删除带宽包资源，包括[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)和[负载均衡](https://cloud.tencent.com/document/product/214/517)等
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGENOTFOUND = "InvalidParameterValue.BandwidthPackageNotFound"
//	INVALIDPARAMETERVALUE_RESOURCEIDMALFORMED = "InvalidParameterValue.ResourceIdMalformed"
//	INVALIDPARAMETERVALUE_RESOURCENOTEXISTED = "InvalidParameterValue.ResourceNotExisted"
//	INVALIDPARAMETERVALUE_RESOURCENOTFOUND = "InvalidParameterValue.ResourceNotFound"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDRESOURCEPROTOCOL = "UnsupportedOperation.InvalidResourceProtocol"
func (c *Client) RemoveBandwidthPackageResourcesWithContext(ctx context.Context, request *RemoveBandwidthPackageResourcesRequest) (response *RemoveBandwidthPackageResourcesResponse, err error) {
	if request == nil {
		request = NewRemoveBandwidthPackageResourcesRequest()
	}
	request.SetContext(ctx)

	response = NewRemoveBandwidthPackageResourcesResponse()
	err = c.Send(request, response)
	return
}

func NewRemoveIp6RulesRequest() (request *RemoveIp6RulesRequest) {
	request = &RemoveIp6RulesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "RemoveIp6Rules")

	return
}

func NewRemoveIp6RulesResponse() (response *RemoveIp6RulesResponse) {
	response = &RemoveIp6RulesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RemoveIp6Rules
// 1. 该接口用于删除IPV6转换规则
//
// 2. 支持批量删除同一个转换实例下的多个转换规则
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) RemoveIp6Rules(request *RemoveIp6RulesRequest) (response *RemoveIp6RulesResponse, err error) {
	if request == nil {
		request = NewRemoveIp6RulesRequest()
	}

	response = NewRemoveIp6RulesResponse()
	err = c.Send(request, response)
	return
}

// RemoveIp6Rules
// 1. 该接口用于删除IPV6转换规则
//
// 2. 支持批量删除同一个转换实例下的多个转换规则
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) RemoveIp6RulesWithContext(ctx context.Context, request *RemoveIp6RulesRequest) (response *RemoveIp6RulesResponse, err error) {
	if request == nil {
		request = NewRemoveIp6RulesRequest()
	}
	request.SetContext(ctx)

	response = NewRemoveIp6RulesResponse()
	err = c.Send(request, response)
	return
}

func NewRenewAddressesRequest() (request *RenewAddressesRequest) {
	request = &RenewAddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "RenewAddresses")

	return
}

func NewRenewAddressesResponse() (response *RenewAddressesResponse) {
	response = &RenewAddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RenewAddresses
// 该接口用于续费包月带宽计费模式的弹性公网IP
//
// 可能返回的错误码:
//
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
func (c *Client) RenewAddresses(request *RenewAddressesRequest) (response *RenewAddressesResponse, err error) {
	if request == nil {
		request = NewRenewAddressesRequest()
	}

	response = NewRenewAddressesResponse()
	err = c.Send(request, response)
	return
}

// RenewAddresses
// 该接口用于续费包月带宽计费模式的弹性公网IP
//
// 可能返回的错误码:
//
//	INVALIDADDRESSID_NOTFOUND = "InvalidAddressId.NotFound"
func (c *Client) RenewAddressesWithContext(ctx context.Context, request *RenewAddressesRequest) (response *RenewAddressesResponse, err error) {
	if request == nil {
		request = NewRenewAddressesRequest()
	}
	request.SetContext(ctx)

	response = NewRenewAddressesResponse()
	err = c.Send(request, response)
	return
}

func NewRenewVpnGatewayRequest() (request *RenewVpnGatewayRequest) {
	request = &RenewVpnGatewayRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "RenewVpnGateway")

	return
}

func NewRenewVpnGatewayResponse() (response *RenewVpnGatewayResponse) {
	response = &RenewVpnGatewayResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RenewVpnGateway
// 本接口（RenewVpnGateway）用于预付费（包年包月）VPN网关续费。目前只支持IPSEC网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) RenewVpnGateway(request *RenewVpnGatewayRequest) (response *RenewVpnGatewayResponse, err error) {
	if request == nil {
		request = NewRenewVpnGatewayRequest()
	}

	response = NewRenewVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

// RenewVpnGateway
// 本接口（RenewVpnGateway）用于预付费（包年包月）VPN网关续费。目前只支持IPSEC网关。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) RenewVpnGatewayWithContext(ctx context.Context, request *RenewVpnGatewayRequest) (response *RenewVpnGatewayResponse, err error) {
	if request == nil {
		request = NewRenewVpnGatewayRequest()
	}
	request.SetContext(ctx)

	response = NewRenewVpnGatewayResponse()
	err = c.Send(request, response)
	return
}

func NewReplaceDirectConnectGatewayCcnRoutesRequest() (request *ReplaceDirectConnectGatewayCcnRoutesRequest) {
	request = &ReplaceDirectConnectGatewayCcnRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ReplaceDirectConnectGatewayCcnRoutes")

	return
}

func NewReplaceDirectConnectGatewayCcnRoutesResponse() (response *ReplaceDirectConnectGatewayCcnRoutesResponse) {
	response = &ReplaceDirectConnectGatewayCcnRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ReplaceDirectConnectGatewayCcnRoutes
// 本接口（ReplaceDirectConnectGatewayCcnRoutes）根据路由ID（RouteId）修改指定的路由（Route），支持批量修改。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ReplaceDirectConnectGatewayCcnRoutes(request *ReplaceDirectConnectGatewayCcnRoutesRequest) (response *ReplaceDirectConnectGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewReplaceDirectConnectGatewayCcnRoutesRequest()
	}

	response = NewReplaceDirectConnectGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

// ReplaceDirectConnectGatewayCcnRoutes
// 本接口（ReplaceDirectConnectGatewayCcnRoutes）根据路由ID（RouteId）修改指定的路由（Route），支持批量修改。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ReplaceDirectConnectGatewayCcnRoutesWithContext(ctx context.Context, request *ReplaceDirectConnectGatewayCcnRoutesRequest) (response *ReplaceDirectConnectGatewayCcnRoutesResponse, err error) {
	if request == nil {
		request = NewReplaceDirectConnectGatewayCcnRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewReplaceDirectConnectGatewayCcnRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewReplaceRouteTableAssociationRequest() (request *ReplaceRouteTableAssociationRequest) {
	request = &ReplaceRouteTableAssociationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ReplaceRouteTableAssociation")

	return
}

func NewReplaceRouteTableAssociationResponse() (response *ReplaceRouteTableAssociationResponse) {
	response = &ReplaceRouteTableAssociationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ReplaceRouteTableAssociation
// 本接口（ReplaceRouteTableAssociation)用于修改子网（Subnet）关联的路由表（RouteTable）。
//
// * 一个子网只能关联一个路由表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) ReplaceRouteTableAssociation(request *ReplaceRouteTableAssociationRequest) (response *ReplaceRouteTableAssociationResponse, err error) {
	if request == nil {
		request = NewReplaceRouteTableAssociationRequest()
	}

	response = NewReplaceRouteTableAssociationResponse()
	err = c.Send(request, response)
	return
}

// ReplaceRouteTableAssociation
// 本接口（ReplaceRouteTableAssociation)用于修改子网（Subnet）关联的路由表（RouteTable）。
//
// * 一个子网只能关联一个路由表。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_VPCMISMATCH = "UnsupportedOperation.VpcMismatch"
func (c *Client) ReplaceRouteTableAssociationWithContext(ctx context.Context, request *ReplaceRouteTableAssociationRequest) (response *ReplaceRouteTableAssociationResponse, err error) {
	if request == nil {
		request = NewReplaceRouteTableAssociationRequest()
	}
	request.SetContext(ctx)

	response = NewReplaceRouteTableAssociationResponse()
	err = c.Send(request, response)
	return
}

func NewReplaceRoutesRequest() (request *ReplaceRoutesRequest) {
	request = &ReplaceRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ReplaceRoutes")

	return
}

func NewReplaceRoutesResponse() (response *ReplaceRoutesResponse) {
	response = &ReplaceRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ReplaceRoutes
// 本接口（ReplaceRoutes）根据路由策略ID（RouteId）修改指定的路由策略（Route），支持批量修改。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CIDRNOTINPEERVPC = "InvalidParameterValue.CidrNotInPeerVpc"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER_WITHGUESS = "UnknownParameter.WithGuess"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CDCSUBNETNOTSUPPORTUNLOCALGATEWAY = "UnsupportedOperation.CdcSubnetNotSupportUnLocalGateway"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
//	UNSUPPORTEDOPERATION_NORMALSUBNETNOTSUPPORTLOCALGATEWAY = "UnsupportedOperation.NormalSubnetNotSupportLocalGateway"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) ReplaceRoutes(request *ReplaceRoutesRequest) (response *ReplaceRoutesResponse, err error) {
	if request == nil {
		request = NewReplaceRoutesRequest()
	}

	response = NewReplaceRoutesResponse()
	err = c.Send(request, response)
	return
}

// ReplaceRoutes
// 本接口（ReplaceRoutes）根据路由策略ID（RouteId）修改指定的路由策略（Route），支持批量修改。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CIDRNOTINPEERVPC = "InvalidParameterValue.CidrNotInPeerVpc"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER_WITHGUESS = "UnknownParameter.WithGuess"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_CDCSUBNETNOTSUPPORTUNLOCALGATEWAY = "UnsupportedOperation.CdcSubnetNotSupportUnLocalGateway"
//	UNSUPPORTEDOPERATION_CONFLICTWITHDOCKERROUTE = "UnsupportedOperation.ConflictWithDockerRoute"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
//	UNSUPPORTEDOPERATION_NORMALSUBNETNOTSUPPORTLOCALGATEWAY = "UnsupportedOperation.NormalSubnetNotSupportLocalGateway"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) ReplaceRoutesWithContext(ctx context.Context, request *ReplaceRoutesRequest) (response *ReplaceRoutesResponse, err error) {
	if request == nil {
		request = NewReplaceRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewReplaceRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewReplaceSecurityGroupPolicyRequest() (request *ReplaceSecurityGroupPolicyRequest) {
	request = &ReplaceSecurityGroupPolicyRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ReplaceSecurityGroupPolicy")

	return
}

func NewReplaceSecurityGroupPolicyResponse() (response *ReplaceSecurityGroupPolicyResponse) {
	response = &ReplaceSecurityGroupPolicyResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ReplaceSecurityGroupPolicy
// 本接口（ReplaceSecurityGroupPolicy）用于替换单条安全组规则（SecurityGroupPolicy）。
//
// 单个请求中只能替换单个方向的一条规则, 必须要指定索引（PolicyIndex）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_CLBPOLICYLIMIT = "UnsupportedOperation.ClbPolicyLimit"
//	UNSUPPORTEDOPERATION_DUPLICATEPOLICY = "UnsupportedOperation.DuplicatePolicy"
//	UNSUPPORTEDOPERATION_VERSIONMISMATCH = "UnsupportedOperation.VersionMismatch"
func (c *Client) ReplaceSecurityGroupPolicy(request *ReplaceSecurityGroupPolicyRequest) (response *ReplaceSecurityGroupPolicyResponse, err error) {
	if request == nil {
		request = NewReplaceSecurityGroupPolicyRequest()
	}

	response = NewReplaceSecurityGroupPolicyResponse()
	err = c.Send(request, response)
	return
}

// ReplaceSecurityGroupPolicy
// 本接口（ReplaceSecurityGroupPolicy）用于替换单条安全组规则（SecurityGroupPolicy）。
//
// 单个请求中只能替换单个方向的一条规则, 必须要指定索引（PolicyIndex）。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_COEXIST = "InvalidParameter.Coexist"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION_CLBPOLICYLIMIT = "UnsupportedOperation.ClbPolicyLimit"
//	UNSUPPORTEDOPERATION_DUPLICATEPOLICY = "UnsupportedOperation.DuplicatePolicy"
//	UNSUPPORTEDOPERATION_VERSIONMISMATCH = "UnsupportedOperation.VersionMismatch"
func (c *Client) ReplaceSecurityGroupPolicyWithContext(ctx context.Context, request *ReplaceSecurityGroupPolicyRequest) (response *ReplaceSecurityGroupPolicyResponse, err error) {
	if request == nil {
		request = NewReplaceSecurityGroupPolicyRequest()
	}
	request.SetContext(ctx)

	response = NewReplaceSecurityGroupPolicyResponse()
	err = c.Send(request, response)
	return
}

func NewResetAttachCcnInstancesRequest() (request *ResetAttachCcnInstancesRequest) {
	request = &ResetAttachCcnInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ResetAttachCcnInstances")

	return
}

func NewResetAttachCcnInstancesResponse() (response *ResetAttachCcnInstancesResponse) {
	response = &ResetAttachCcnInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetAttachCcnInstances
// 本接口（ResetAttachCcnInstances）用于跨账号关联实例申请过期时，重新申请关联操作。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ResetAttachCcnInstances(request *ResetAttachCcnInstancesRequest) (response *ResetAttachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewResetAttachCcnInstancesRequest()
	}

	response = NewResetAttachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

// ResetAttachCcnInstances
// 本接口（ResetAttachCcnInstances）用于跨账号关联实例申请过期时，重新申请关联操作。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ResetAttachCcnInstancesWithContext(ctx context.Context, request *ResetAttachCcnInstancesRequest) (response *ResetAttachCcnInstancesResponse, err error) {
	if request == nil {
		request = NewResetAttachCcnInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewResetAttachCcnInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewResetNatGatewayConnectionRequest() (request *ResetNatGatewayConnectionRequest) {
	request = &ResetNatGatewayConnectionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ResetNatGatewayConnection")

	return
}

func NewResetNatGatewayConnectionResponse() (response *ResetNatGatewayConnectionResponse) {
	response = &ResetNatGatewayConnectionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetNatGatewayConnection
// 本接口（ResetNatGatewayConnection）用来NAT网关并发连接上限。
//
// 可能返回的错误码:
//
//	RESOURCEINUSE = "ResourceInUse"
//	UNSUPPORTEDOPERATION_UNPAIDORDERALREADYEXISTS = "UnsupportedOperation.UnpaidOrderAlreadyExists"
func (c *Client) ResetNatGatewayConnection(request *ResetNatGatewayConnectionRequest) (response *ResetNatGatewayConnectionResponse, err error) {
	if request == nil {
		request = NewResetNatGatewayConnectionRequest()
	}

	response = NewResetNatGatewayConnectionResponse()
	err = c.Send(request, response)
	return
}

// ResetNatGatewayConnection
// 本接口（ResetNatGatewayConnection）用来NAT网关并发连接上限。
//
// 可能返回的错误码:
//
//	RESOURCEINUSE = "ResourceInUse"
//	UNSUPPORTEDOPERATION_UNPAIDORDERALREADYEXISTS = "UnsupportedOperation.UnpaidOrderAlreadyExists"
func (c *Client) ResetNatGatewayConnectionWithContext(ctx context.Context, request *ResetNatGatewayConnectionRequest) (response *ResetNatGatewayConnectionResponse, err error) {
	if request == nil {
		request = NewResetNatGatewayConnectionRequest()
	}
	request.SetContext(ctx)

	response = NewResetNatGatewayConnectionResponse()
	err = c.Send(request, response)
	return
}

func NewResetRoutesRequest() (request *ResetRoutesRequest) {
	request = &ResetRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ResetRoutes")

	return
}

func NewResetRoutesResponse() (response *ResetRoutesResponse) {
	response = &ResetRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetRoutes
// 本接口（ResetRoutes）用于对某个路由表名称和所有路由策略（Route）进行重新设置。<br />
//
// 注意: 调用本接口是先删除当前路由表中所有路由策略, 再保存新提交的路由策略内容, 会引起网络中断。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CIDRNOTINPEERVPC = "InvalidParameterValue.CidrNotInPeerVpc"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) ResetRoutes(request *ResetRoutesRequest) (response *ResetRoutesResponse, err error) {
	if request == nil {
		request = NewResetRoutesRequest()
	}

	response = NewResetRoutesResponse()
	err = c.Send(request, response)
	return
}

// ResetRoutes
// 本接口（ResetRoutes）用于对某个路由表名称和所有路由策略（Route）进行重新设置。<br />
//
// 注意: 调用本接口是先删除当前路由表中所有路由策略, 再保存新提交的路由策略内容, 会引起网络中断。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CIDRNOTINPEERVPC = "InvalidParameterValue.CidrNotInPeerVpc"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_VPCCIDRCONFLICT = "InvalidParameterValue.VpcCidrConflict"
//	LIMITEXCEEDED = "LimitExceeded"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ECMP = "UnsupportedOperation.Ecmp"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) ResetRoutesWithContext(ctx context.Context, request *ResetRoutesRequest) (response *ResetRoutesResponse, err error) {
	if request == nil {
		request = NewResetRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewResetRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewResetVpnConnectionRequest() (request *ResetVpnConnectionRequest) {
	request = &ResetVpnConnectionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ResetVpnConnection")

	return
}

func NewResetVpnConnectionResponse() (response *ResetVpnConnectionResponse) {
	response = &ResetVpnConnectionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetVpnConnection
// 本接口(ResetVpnConnection)用于重置VPN通道。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ResetVpnConnection(request *ResetVpnConnectionRequest) (response *ResetVpnConnectionResponse, err error) {
	if request == nil {
		request = NewResetVpnConnectionRequest()
	}

	response = NewResetVpnConnectionResponse()
	err = c.Send(request, response)
	return
}

// ResetVpnConnection
// 本接口(ResetVpnConnection)用于重置VPN通道。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) ResetVpnConnectionWithContext(ctx context.Context, request *ResetVpnConnectionRequest) (response *ResetVpnConnectionResponse, err error) {
	if request == nil {
		request = NewResetVpnConnectionRequest()
	}
	request.SetContext(ctx)

	response = NewResetVpnConnectionResponse()
	err = c.Send(request, response)
	return
}

func NewResetVpnGatewayInternetMaxBandwidthRequest() (request *ResetVpnGatewayInternetMaxBandwidthRequest) {
	request = &ResetVpnGatewayInternetMaxBandwidthRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "ResetVpnGatewayInternetMaxBandwidth")

	return
}

func NewResetVpnGatewayInternetMaxBandwidthResponse() (response *ResetVpnGatewayInternetMaxBandwidthResponse) {
	response = &ResetVpnGatewayInternetMaxBandwidthResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetVpnGatewayInternetMaxBandwidth
// 本接口（ResetVpnGatewayInternetMaxBandwidth）调整VPN网关带宽上限。目前支持升级配置，如果是包年包月VPN网关需要在有效期内。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) ResetVpnGatewayInternetMaxBandwidth(request *ResetVpnGatewayInternetMaxBandwidthRequest) (response *ResetVpnGatewayInternetMaxBandwidthResponse, err error) {
	if request == nil {
		request = NewResetVpnGatewayInternetMaxBandwidthRequest()
	}

	response = NewResetVpnGatewayInternetMaxBandwidthResponse()
	err = c.Send(request, response)
	return
}

// ResetVpnGatewayInternetMaxBandwidth
// 本接口（ResetVpnGatewayInternetMaxBandwidth）调整VPN网关带宽上限。目前支持升级配置，如果是包年包月VPN网关需要在有效期内。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
func (c *Client) ResetVpnGatewayInternetMaxBandwidthWithContext(ctx context.Context, request *ResetVpnGatewayInternetMaxBandwidthRequest) (response *ResetVpnGatewayInternetMaxBandwidthResponse, err error) {
	if request == nil {
		request = NewResetVpnGatewayInternetMaxBandwidthRequest()
	}
	request.SetContext(ctx)

	response = NewResetVpnGatewayInternetMaxBandwidthResponse()
	err = c.Send(request, response)
	return
}

func NewSetCcnRegionBandwidthLimitsRequest() (request *SetCcnRegionBandwidthLimitsRequest) {
	request = &SetCcnRegionBandwidthLimitsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "SetCcnRegionBandwidthLimits")

	return
}

func NewSetCcnRegionBandwidthLimitsResponse() (response *SetCcnRegionBandwidthLimitsResponse) {
	response = &SetCcnRegionBandwidthLimitsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// SetCcnRegionBandwidthLimits
// 本接口（SetCcnRegionBandwidthLimits）用于设置云联网（CCN）各地域出带宽上限，或者地域间带宽上限。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_NOTPOSTPAIDCCNOPERATION = "UnsupportedOperation.NotPostpaidCcnOperation"
func (c *Client) SetCcnRegionBandwidthLimits(request *SetCcnRegionBandwidthLimitsRequest) (response *SetCcnRegionBandwidthLimitsResponse, err error) {
	if request == nil {
		request = NewSetCcnRegionBandwidthLimitsRequest()
	}

	response = NewSetCcnRegionBandwidthLimitsResponse()
	err = c.Send(request, response)
	return
}

// SetCcnRegionBandwidthLimits
// 本接口（SetCcnRegionBandwidthLimits）用于设置云联网（CCN）各地域出带宽上限，或者地域间带宽上限。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_NOTPOSTPAIDCCNOPERATION = "UnsupportedOperation.NotPostpaidCcnOperation"
func (c *Client) SetCcnRegionBandwidthLimitsWithContext(ctx context.Context, request *SetCcnRegionBandwidthLimitsRequest) (response *SetCcnRegionBandwidthLimitsResponse, err error) {
	if request == nil {
		request = NewSetCcnRegionBandwidthLimitsRequest()
	}
	request.SetContext(ctx)

	response = NewSetCcnRegionBandwidthLimitsResponse()
	err = c.Send(request, response)
	return
}

func NewTransformAddressRequest() (request *TransformAddressRequest) {
	request = &TransformAddressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "TransformAddress")

	return
}

func NewTransformAddressResponse() (response *TransformAddressResponse) {
	response = &TransformAddressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// TransformAddress
// 本接口 (TransformAddress) 用于将实例的普通公网 IP 转换为[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）。
//
// * 平台对用户每地域每日解绑 EIP 重新分配普通公网 IP 次数有所限制（可参见 [EIP 产品简介](/document/product/213/1941)）。上述配额可通过 [DescribeAddressQuota](https://cloud.tencent.com/document/api/213/1378) 接口获取。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED = "AddressQuotaLimitExceeded"
//	ADDRESSQUOTALIMITEXCEEDED_DAILYALLOCATE = "AddressQuotaLimitExceeded.DailyAllocate"
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_ALREADYBINDEIP = "InvalidInstanceId.AlreadyBindEip"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_INSTANCENOWANIP = "InvalidParameterValue.InstanceNoWanIP"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCESTATE = "InvalidParameterValue.InvalidInstanceState"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INVALIDADDRESSINTERNETCHARGETYPE = "UnsupportedOperation.InvalidAddressInternetChargeType"
func (c *Client) TransformAddress(request *TransformAddressRequest) (response *TransformAddressResponse, err error) {
	if request == nil {
		request = NewTransformAddressRequest()
	}

	response = NewTransformAddressResponse()
	err = c.Send(request, response)
	return
}

// TransformAddress
// 本接口 (TransformAddress) 用于将实例的普通公网 IP 转换为[弹性公网IP](https://cloud.tencent.com/document/product/213/1941)（简称 EIP）。
//
// * 平台对用户每地域每日解绑 EIP 重新分配普通公网 IP 次数有所限制（可参见 [EIP 产品简介](/document/product/213/1941)）。上述配额可通过 [DescribeAddressQuota](https://cloud.tencent.com/document/api/213/1378) 接口获取。
//
// 可能返回的错误码:
//
//	ADDRESSQUOTALIMITEXCEEDED = "AddressQuotaLimitExceeded"
//	ADDRESSQUOTALIMITEXCEEDED_DAILYALLOCATE = "AddressQuotaLimitExceeded.DailyAllocate"
//	INVALIDADDRESSID_BLOCKED = "InvalidAddressId.Blocked"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_ALREADYBINDEIP = "InvalidInstanceId.AlreadyBindEip"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_INSTANCENOWANIP = "InvalidParameterValue.InstanceNoWanIP"
//	INVALIDPARAMETERVALUE_INVALIDINSTANCESTATE = "InvalidParameterValue.InvalidInstanceState"
//	UNSUPPORTEDOPERATION_ADDRESSSTATUSNOTPERMIT = "UnsupportedOperation.AddressStatusNotPermit"
//	UNSUPPORTEDOPERATION_INVALIDADDRESSINTERNETCHARGETYPE = "UnsupportedOperation.InvalidAddressInternetChargeType"
func (c *Client) TransformAddressWithContext(ctx context.Context, request *TransformAddressRequest) (response *TransformAddressResponse, err error) {
	if request == nil {
		request = NewTransformAddressRequest()
	}
	request.SetContext(ctx)

	response = NewTransformAddressResponse()
	err = c.Send(request, response)
	return
}

func NewUnassignIpv6AddressesRequest() (request *UnassignIpv6AddressesRequest) {
	request = &UnassignIpv6AddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "UnassignIpv6Addresses")

	return
}

func NewUnassignIpv6AddressesResponse() (response *UnassignIpv6AddressesResponse) {
	response = &UnassignIpv6AddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UnassignIpv6Addresses
// 本接口（UnassignIpv6Addresses）用于释放弹性网卡`IPv6`地址。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_ATTACHMENTNOTFOUND = "UnauthorizedOperation.AttachmentNotFound"
//	UNAUTHORIZEDOPERATION_PRIMARYIP = "UnauthorizedOperation.PrimaryIp"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ATTACHMENTNOTFOUND = "UnsupportedOperation.AttachmentNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) UnassignIpv6Addresses(request *UnassignIpv6AddressesRequest) (response *UnassignIpv6AddressesResponse, err error) {
	if request == nil {
		request = NewUnassignIpv6AddressesRequest()
	}

	response = NewUnassignIpv6AddressesResponse()
	err = c.Send(request, response)
	return
}

// UnassignIpv6Addresses
// 本接口（UnassignIpv6Addresses）用于释放弹性网卡`IPv6`地址。<br />
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_ATTACHMENTNOTFOUND = "UnauthorizedOperation.AttachmentNotFound"
//	UNAUTHORIZEDOPERATION_PRIMARYIP = "UnauthorizedOperation.PrimaryIp"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ATTACHMENTNOTFOUND = "UnsupportedOperation.AttachmentNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) UnassignIpv6AddressesWithContext(ctx context.Context, request *UnassignIpv6AddressesRequest) (response *UnassignIpv6AddressesResponse, err error) {
	if request == nil {
		request = NewUnassignIpv6AddressesRequest()
	}
	request.SetContext(ctx)

	response = NewUnassignIpv6AddressesResponse()
	err = c.Send(request, response)
	return
}

func NewUnassignIpv6CidrBlockRequest() (request *UnassignIpv6CidrBlockRequest) {
	request = &UnassignIpv6CidrBlockRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "UnassignIpv6CidrBlock")

	return
}

func NewUnassignIpv6CidrBlockResponse() (response *UnassignIpv6CidrBlockResponse) {
	response = &UnassignIpv6CidrBlockResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UnassignIpv6CidrBlock
// 本接口（UnassignIpv6CidrBlock）用于释放IPv6网段。<br />
//
// 网段如果还有IP占用且未回收，则网段无法释放。
//
// 可能返回的错误码:
//
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) UnassignIpv6CidrBlock(request *UnassignIpv6CidrBlockRequest) (response *UnassignIpv6CidrBlockResponse, err error) {
	if request == nil {
		request = NewUnassignIpv6CidrBlockRequest()
	}

	response = NewUnassignIpv6CidrBlockResponse()
	err = c.Send(request, response)
	return
}

// UnassignIpv6CidrBlock
// 本接口（UnassignIpv6CidrBlock）用于释放IPv6网段。<br />
//
// 网段如果还有IP占用且未回收，则网段无法释放。
//
// 可能返回的错误码:
//
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) UnassignIpv6CidrBlockWithContext(ctx context.Context, request *UnassignIpv6CidrBlockRequest) (response *UnassignIpv6CidrBlockResponse, err error) {
	if request == nil {
		request = NewUnassignIpv6CidrBlockRequest()
	}
	request.SetContext(ctx)

	response = NewUnassignIpv6CidrBlockResponse()
	err = c.Send(request, response)
	return
}

func NewUnassignIpv6SubnetCidrBlockRequest() (request *UnassignIpv6SubnetCidrBlockRequest) {
	request = &UnassignIpv6SubnetCidrBlockRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "UnassignIpv6SubnetCidrBlock")

	return
}

func NewUnassignIpv6SubnetCidrBlockResponse() (response *UnassignIpv6SubnetCidrBlockResponse) {
	response = &UnassignIpv6SubnetCidrBlockResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UnassignIpv6SubnetCidrBlock
// 本接口（UnassignIpv6SubnetCidrBlock）用于释放IPv6子网段。<br />
//
// 子网段如果还有IP占用且未回收，则子网段无法释放。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) UnassignIpv6SubnetCidrBlock(request *UnassignIpv6SubnetCidrBlockRequest) (response *UnassignIpv6SubnetCidrBlockResponse, err error) {
	if request == nil {
		request = NewUnassignIpv6SubnetCidrBlockRequest()
	}

	response = NewUnassignIpv6SubnetCidrBlockResponse()
	err = c.Send(request, response)
	return
}

// UnassignIpv6SubnetCidrBlock
// 本接口（UnassignIpv6SubnetCidrBlock）用于释放IPv6子网段。<br />
//
// 子网段如果还有IP占用且未回收，则子网段无法释放。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) UnassignIpv6SubnetCidrBlockWithContext(ctx context.Context, request *UnassignIpv6SubnetCidrBlockRequest) (response *UnassignIpv6SubnetCidrBlockResponse, err error) {
	if request == nil {
		request = NewUnassignIpv6SubnetCidrBlockRequest()
	}
	request.SetContext(ctx)

	response = NewUnassignIpv6SubnetCidrBlockResponse()
	err = c.Send(request, response)
	return
}

func NewUnassignPrivateIpAddressesRequest() (request *UnassignPrivateIpAddressesRequest) {
	request = &UnassignPrivateIpAddressesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "UnassignPrivateIpAddresses")

	return
}

func NewUnassignPrivateIpAddressesResponse() (response *UnassignPrivateIpAddressesResponse) {
	response = &UnassignPrivateIpAddressesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UnassignPrivateIpAddresses
// 本接口（UnassignPrivateIpAddresses）用于弹性网卡退还内网 IP。
//
// * 退还弹性网卡上的辅助内网IP，接口自动解关联弹性公网 IP。不能退还弹性网卡的主内网IP。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ATTACHMENTNOTFOUND = "UnsupportedOperation.AttachmentNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) UnassignPrivateIpAddresses(request *UnassignPrivateIpAddressesRequest) (response *UnassignPrivateIpAddressesResponse, err error) {
	if request == nil {
		request = NewUnassignPrivateIpAddressesRequest()
	}

	response = NewUnassignPrivateIpAddressesResponse()
	err = c.Send(request, response)
	return
}

// UnassignPrivateIpAddresses
// 本接口（UnassignPrivateIpAddresses）用于弹性网卡退还内网 IP。
//
// * 退还弹性网卡上的辅助内网IP，接口自动解关联弹性公网 IP。不能退还弹性网卡的主内网IP。
//
// 本接口是异步完成，如需查询异步任务执行结果，请使用本接口返回的`RequestId`轮询`DescribeVpcTaskResult`接口。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_ATTACHMENTNOTFOUND = "UnsupportedOperation.AttachmentNotFound"
//	UNSUPPORTEDOPERATION_INVALIDSTATE = "UnsupportedOperation.InvalidState"
//	UNSUPPORTEDOPERATION_MUTEXOPERATIONTASKRUNNING = "UnsupportedOperation.MutexOperationTaskRunning"
func (c *Client) UnassignPrivateIpAddressesWithContext(ctx context.Context, request *UnassignPrivateIpAddressesRequest) (response *UnassignPrivateIpAddressesResponse, err error) {
	if request == nil {
		request = NewUnassignPrivateIpAddressesRequest()
	}
	request.SetContext(ctx)

	response = NewUnassignPrivateIpAddressesResponse()
	err = c.Send(request, response)
	return
}

func NewWithdrawNotifyRoutesRequest() (request *WithdrawNotifyRoutesRequest) {
	request = &WithdrawNotifyRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("vpc", APIVersion, "WithdrawNotifyRoutes")

	return
}

func NewWithdrawNotifyRoutesResponse() (response *WithdrawNotifyRoutesResponse) {
	response = &WithdrawNotifyRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// WithdrawNotifyRoutes
// 路由表列表页操作增加“从云联网撤销”，用于撤销已发布到云联网的路由。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDROUTEID_NOTFOUND = "InvalidRouteId.NotFound"
//	INVALIDROUTETABLEID_MALFORMED = "InvalidRouteTableId.Malformed"
//	INVALIDROUTETABLEID_NOTFOUND = "InvalidRouteTableId.NotFound"
//	UNSUPPORTEDOPERATION_NOTIFYCCN = "UnsupportedOperation.NotifyCcn"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) WithdrawNotifyRoutes(request *WithdrawNotifyRoutesRequest) (response *WithdrawNotifyRoutesResponse, err error) {
	if request == nil {
		request = NewWithdrawNotifyRoutesRequest()
	}

	response = NewWithdrawNotifyRoutesResponse()
	err = c.Send(request, response)
	return
}

// WithdrawNotifyRoutes
// 路由表列表页操作增加“从云联网撤销”，用于撤销已发布到云联网的路由。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE_MALFORMED = "InvalidParameterValue.Malformed"
//	INVALIDROUTEID_NOTFOUND = "InvalidRouteId.NotFound"
//	INVALIDROUTETABLEID_MALFORMED = "InvalidRouteTableId.Malformed"
//	INVALIDROUTETABLEID_NOTFOUND = "InvalidRouteTableId.NotFound"
//	UNSUPPORTEDOPERATION_NOTIFYCCN = "UnsupportedOperation.NotifyCcn"
//	UNSUPPORTEDOPERATION_SYSTEMROUTE = "UnsupportedOperation.SystemRoute"
func (c *Client) WithdrawNotifyRoutesWithContext(ctx context.Context, request *WithdrawNotifyRoutesRequest) (response *WithdrawNotifyRoutesResponse, err error) {
	if request == nil {
		request = NewWithdrawNotifyRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewWithdrawNotifyRoutesResponse()
	err = c.Send(request, response)
	return
}
