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

func NewAllocateHostsRequest() (request *AllocateHostsRequest) {
	request = &AllocateHostsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "AllocateHosts")

	return
}

func NewAllocateHostsResponse() (response *AllocateHostsResponse) {
	response = &AllocateHostsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AllocateHosts
// 本接口 (AllocateHosts) 用于创建一个或多个指定配置的CDH实例。
//
// * 当HostChargeType为PREPAID时，必须指定HostChargePrepaid参数。
//
// 可能返回的错误码:
//
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	RESOURCEINSUFFICIENT_ZONESOLDOUTFORSPECIFIEDINSTANCE = "ResourceInsufficient.ZoneSoldOutForSpecifiedInstance"
func (c *Client) AllocateHosts(request *AllocateHostsRequest) (response *AllocateHostsResponse, err error) {
	if request == nil {
		request = NewAllocateHostsRequest()
	}

	response = NewAllocateHostsResponse()
	err = c.Send(request, response)
	return
}

// AllocateHosts
// 本接口 (AllocateHosts) 用于创建一个或多个指定配置的CDH实例。
//
// * 当HostChargeType为PREPAID时，必须指定HostChargePrepaid参数。
//
// 可能返回的错误码:
//
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	RESOURCEINSUFFICIENT_ZONESOLDOUTFORSPECIFIEDINSTANCE = "ResourceInsufficient.ZoneSoldOutForSpecifiedInstance"
func (c *Client) AllocateHostsWithContext(ctx context.Context, request *AllocateHostsRequest) (response *AllocateHostsResponse, err error) {
	if request == nil {
		request = NewAllocateHostsRequest()
	}
	request.SetContext(ctx)

	response = NewAllocateHostsResponse()
	err = c.Send(request, response)
	return
}

func NewAssociateInstancesKeyPairsRequest() (request *AssociateInstancesKeyPairsRequest) {
	request = &AssociateInstancesKeyPairsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "AssociateInstancesKeyPairs")

	return
}

func NewAssociateInstancesKeyPairsResponse() (response *AssociateInstancesKeyPairsResponse) {
	response = &AssociateInstancesKeyPairsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssociateInstancesKeyPairs
// 本接口 (AssociateInstancesKeyPairs) 用于将密钥绑定到实例上。
//
// * 将密钥的公钥写入到实例的`SSH`配置当中，用户就可以通过该密钥的私钥来登录实例。
//
// * 如果实例原来绑定过密钥，那么原来的密钥将失效。
//
// * 如果实例原来是通过密码登录，绑定密钥后无法使用密码登录。
//
// * 支持批量操作。每次请求批量实例的上限为100。如果批量实例存在不允许操作的实例，操作会以特定错误码返回。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_KEYPAIRNOTSUPPORTED = "InvalidParameterValue.KeyPairNotSupported"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCEOSWINDOWS = "UnsupportedOperation.InstanceOsWindows"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) AssociateInstancesKeyPairs(request *AssociateInstancesKeyPairsRequest) (response *AssociateInstancesKeyPairsResponse, err error) {
	if request == nil {
		request = NewAssociateInstancesKeyPairsRequest()
	}

	response = NewAssociateInstancesKeyPairsResponse()
	err = c.Send(request, response)
	return
}

// AssociateInstancesKeyPairs
// 本接口 (AssociateInstancesKeyPairs) 用于将密钥绑定到实例上。
//
// * 将密钥的公钥写入到实例的`SSH`配置当中，用户就可以通过该密钥的私钥来登录实例。
//
// * 如果实例原来绑定过密钥，那么原来的密钥将失效。
//
// * 如果实例原来是通过密码登录，绑定密钥后无法使用密码登录。
//
// * 支持批量操作。每次请求批量实例的上限为100。如果批量实例存在不允许操作的实例，操作会以特定错误码返回。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_KEYPAIRNOTSUPPORTED = "InvalidParameterValue.KeyPairNotSupported"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCEOSWINDOWS = "UnsupportedOperation.InstanceOsWindows"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) AssociateInstancesKeyPairsWithContext(ctx context.Context, request *AssociateInstancesKeyPairsRequest) (response *AssociateInstancesKeyPairsResponse, err error) {
	if request == nil {
		request = NewAssociateInstancesKeyPairsRequest()
	}
	request.SetContext(ctx)

	response = NewAssociateInstancesKeyPairsResponse()
	err = c.Send(request, response)
	return
}

func NewAssociateSecurityGroupsRequest() (request *AssociateSecurityGroupsRequest) {
	request = &AssociateSecurityGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "AssociateSecurityGroups")

	return
}

func NewAssociateSecurityGroupsResponse() (response *AssociateSecurityGroupsResponse) {
	response = &AssociateSecurityGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AssociateSecurityGroups
// 本接口 (AssociateSecurityGroups) 用于绑定安全组到指定实例。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDSGID_MALFORMED = "InvalidSgId.Malformed"
//	LIMITEXCEEDED_ASSOCIATEUSGLIMITEXCEEDED = "LimitExceeded.AssociateUSGLimitExceeded"
//	LIMITEXCEEDED_CVMSVIFSPERSECGROUPLIMITEXCEEDED = "LimitExceeded.CvmsVifsPerSecGroupLimitExceeded"
//	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	SECGROUPACTIONFAILURE = "SecGroupActionFailure"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
func (c *Client) AssociateSecurityGroups(request *AssociateSecurityGroupsRequest) (response *AssociateSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewAssociateSecurityGroupsRequest()
	}

	response = NewAssociateSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

// AssociateSecurityGroups
// 本接口 (AssociateSecurityGroups) 用于绑定安全组到指定实例。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDSGID_MALFORMED = "InvalidSgId.Malformed"
//	LIMITEXCEEDED_ASSOCIATEUSGLIMITEXCEEDED = "LimitExceeded.AssociateUSGLimitExceeded"
//	LIMITEXCEEDED_CVMSVIFSPERSECGROUPLIMITEXCEEDED = "LimitExceeded.CvmsVifsPerSecGroupLimitExceeded"
//	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	SECGROUPACTIONFAILURE = "SecGroupActionFailure"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
func (c *Client) AssociateSecurityGroupsWithContext(ctx context.Context, request *AssociateSecurityGroupsRequest) (response *AssociateSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewAssociateSecurityGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewAssociateSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewCreateDisasterRecoverGroupRequest() (request *CreateDisasterRecoverGroupRequest) {
	request = &CreateDisasterRecoverGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "CreateDisasterRecoverGroup")

	return
}

func NewCreateDisasterRecoverGroupResponse() (response *CreateDisasterRecoverGroupResponse) {
	response = &CreateDisasterRecoverGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateDisasterRecoverGroup
// 本接口 (CreateDisasterRecoverGroup)用于创建[分散置放群组](https://cloud.tencent.com/document/product/213/15486)。创建好的置放群组，可在[创建实例](https://cloud.tencent.com/document/api/213/15730)时指定。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
func (c *Client) CreateDisasterRecoverGroup(request *CreateDisasterRecoverGroupRequest) (response *CreateDisasterRecoverGroupResponse, err error) {
	if request == nil {
		request = NewCreateDisasterRecoverGroupRequest()
	}

	response = NewCreateDisasterRecoverGroupResponse()
	err = c.Send(request, response)
	return
}

// CreateDisasterRecoverGroup
// 本接口 (CreateDisasterRecoverGroup)用于创建[分散置放群组](https://cloud.tencent.com/document/product/213/15486)。创建好的置放群组，可在[创建实例](https://cloud.tencent.com/document/api/213/15730)时指定。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
func (c *Client) CreateDisasterRecoverGroupWithContext(ctx context.Context, request *CreateDisasterRecoverGroupRequest) (response *CreateDisasterRecoverGroupResponse, err error) {
	if request == nil {
		request = NewCreateDisasterRecoverGroupRequest()
	}
	request.SetContext(ctx)

	response = NewCreateDisasterRecoverGroupResponse()
	err = c.Send(request, response)
	return
}

func NewCreateImageRequest() (request *CreateImageRequest) {
	request = &CreateImageRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "CreateImage")

	return
}

func NewCreateImageResponse() (response *CreateImageResponse) {
	response = &CreateImageResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateImage
// 本接口(CreateImage)用于将实例的系统盘制作为新镜像，创建后的镜像可以用于创建实例。
//
// 可能返回的错误码:
//
//	IMAGEQUOTALIMITEXCEEDED = "ImageQuotaLimitExceeded"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER_DATADISKIDCONTAINSROOTDISK = "InvalidParameter.DataDiskIdContainsRootDisk"
//	INVALIDPARAMETER_DATADISKNOTBELONGSPECIFIEDINSTANCE = "InvalidParameter.DataDiskNotBelongSpecifiedInstance"
//	INVALIDPARAMETER_DUPLICATESYSTEMSNAPSHOTS = "InvalidParameter.DuplicateSystemSnapshots"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_INVALIDDEPENDENCE = "InvalidParameter.InvalidDependence"
//	INVALIDPARAMETER_LOCALDATADISKNOTSUPPORT = "InvalidParameter.LocalDataDiskNotSupport"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETER_SPECIFYONEPARAMETER = "InvalidParameter.SpecifyOneParameter"
//	INVALIDPARAMETER_SWAPDISKNOTSUPPORT = "InvalidParameter.SwapDiskNotSupport"
//	INVALIDPARAMETER_SYSTEMSNAPSHOTNOTFOUND = "InvalidParameter.SystemSnapshotNotFound"
//	INVALIDPARAMETER_VALUETOOLARGE = "InvalidParameter.ValueTooLarge"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_TOOLARGE = "InvalidParameterValue.TooLarge"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEUNAVAILABLE_SNAPSHOTCREATING = "ResourceUnavailable.SnapshotCreating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_NOTSUPPORTINSTANCEIMAGE = "UnsupportedOperation.NotSupportInstanceImage"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) CreateImage(request *CreateImageRequest) (response *CreateImageResponse, err error) {
	if request == nil {
		request = NewCreateImageRequest()
	}

	response = NewCreateImageResponse()
	err = c.Send(request, response)
	return
}

// CreateImage
// 本接口(CreateImage)用于将实例的系统盘制作为新镜像，创建后的镜像可以用于创建实例。
//
// 可能返回的错误码:
//
//	IMAGEQUOTALIMITEXCEEDED = "ImageQuotaLimitExceeded"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER_DATADISKIDCONTAINSROOTDISK = "InvalidParameter.DataDiskIdContainsRootDisk"
//	INVALIDPARAMETER_DATADISKNOTBELONGSPECIFIEDINSTANCE = "InvalidParameter.DataDiskNotBelongSpecifiedInstance"
//	INVALIDPARAMETER_DUPLICATESYSTEMSNAPSHOTS = "InvalidParameter.DuplicateSystemSnapshots"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_INVALIDDEPENDENCE = "InvalidParameter.InvalidDependence"
//	INVALIDPARAMETER_LOCALDATADISKNOTSUPPORT = "InvalidParameter.LocalDataDiskNotSupport"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETER_SPECIFYONEPARAMETER = "InvalidParameter.SpecifyOneParameter"
//	INVALIDPARAMETER_SWAPDISKNOTSUPPORT = "InvalidParameter.SwapDiskNotSupport"
//	INVALIDPARAMETER_SYSTEMSNAPSHOTNOTFOUND = "InvalidParameter.SystemSnapshotNotFound"
//	INVALIDPARAMETER_VALUETOOLARGE = "InvalidParameter.ValueTooLarge"
//	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_TOOLARGE = "InvalidParameterValue.TooLarge"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEUNAVAILABLE_SNAPSHOTCREATING = "ResourceUnavailable.SnapshotCreating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_NOTSUPPORTINSTANCEIMAGE = "UnsupportedOperation.NotSupportInstanceImage"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) CreateImageWithContext(ctx context.Context, request *CreateImageRequest) (response *CreateImageResponse, err error) {
	if request == nil {
		request = NewCreateImageRequest()
	}
	request.SetContext(ctx)

	response = NewCreateImageResponse()
	err = c.Send(request, response)
	return
}

func NewCreateKeyPairRequest() (request *CreateKeyPairRequest) {
	request = &CreateKeyPairRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "CreateKeyPair")

	return
}

func NewCreateKeyPairResponse() (response *CreateKeyPairResponse) {
	response = &CreateKeyPairResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateKeyPair
// 本接口 (CreateKeyPair) 用于创建一个 `OpenSSH RSA` 密钥对，可以用于登录 `Linux` 实例。
//
// * 开发者只需指定密钥对名称，即可由系统自动创建密钥对，并返回所生成的密钥对的 `ID` 及其公钥、私钥的内容。
//
// * 密钥对名称不能和已经存在的密钥对的名称重复。
//
// * 私钥的内容可以保存到文件中作为 `SSH` 的一种认证方式。
//
// * 腾讯云不会保存用户的私钥，请妥善保管。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"
//	INVALIDKEYPAIRNAME_DUPLICATE = "InvalidKeyPairName.Duplicate"
//	INVALIDKEYPAIRNAMEEMPTY = "InvalidKeyPairNameEmpty"
//	INVALIDKEYPAIRNAMEINCLUDEILLEGALCHAR = "InvalidKeyPairNameIncludeIllegalChar"
//	INVALIDKEYPAIRNAMETOOLONG = "InvalidKeyPairNameTooLong"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) CreateKeyPair(request *CreateKeyPairRequest) (response *CreateKeyPairResponse, err error) {
	if request == nil {
		request = NewCreateKeyPairRequest()
	}

	response = NewCreateKeyPairResponse()
	err = c.Send(request, response)
	return
}

// CreateKeyPair
// 本接口 (CreateKeyPair) 用于创建一个 `OpenSSH RSA` 密钥对，可以用于登录 `Linux` 实例。
//
// * 开发者只需指定密钥对名称，即可由系统自动创建密钥对，并返回所生成的密钥对的 `ID` 及其公钥、私钥的内容。
//
// * 密钥对名称不能和已经存在的密钥对的名称重复。
//
// * 私钥的内容可以保存到文件中作为 `SSH` 的一种认证方式。
//
// * 腾讯云不会保存用户的私钥，请妥善保管。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"
//	INVALIDKEYPAIRNAME_DUPLICATE = "InvalidKeyPairName.Duplicate"
//	INVALIDKEYPAIRNAMEEMPTY = "InvalidKeyPairNameEmpty"
//	INVALIDKEYPAIRNAMEINCLUDEILLEGALCHAR = "InvalidKeyPairNameIncludeIllegalChar"
//	INVALIDKEYPAIRNAMETOOLONG = "InvalidKeyPairNameTooLong"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) CreateKeyPairWithContext(ctx context.Context, request *CreateKeyPairRequest) (response *CreateKeyPairResponse, err error) {
	if request == nil {
		request = NewCreateKeyPairRequest()
	}
	request.SetContext(ctx)

	response = NewCreateKeyPairResponse()
	err = c.Send(request, response)
	return
}

func NewCreateLaunchTemplateRequest() (request *CreateLaunchTemplateRequest) {
	request = &CreateLaunchTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "CreateLaunchTemplate")

	return
}

func NewCreateLaunchTemplateResponse() (response *CreateLaunchTemplateResponse) {
	response = &CreateLaunchTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateLaunchTemplate
// 本接口（CreateLaunchTemplate）用于创建实例启动模板。
//
// 实例启动模板是一种配置数据并可用于创建实例，其内容包含创建实例所需的配置，比如实例类型，数据盘和系统盘的类型和大小，以及安全组等信息。
//
// 初次创建实例模板后，其模板版本为默认版本1，新版本的创建可使用CreateLaunchTemplateVersion创建，版本号递增。默认情况下，在RunInstances中指定实例启动模板，若不指定模板版本号，则使用默认版本。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	INSTANCESQUOTALIMITEXCEEDED = "InstancesQuotaLimitExceeded"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_INVALIDIPFORMAT = "InvalidParameter.InvalidIpFormat"
//	INVALIDPARAMETER_LACKCORECOUNTORTHREADPERCORE = "InvalidParameter.LackCoreCountOrThreadPerCore"
//	INVALIDPARAMETER_PASSWORDNOTSUPPORTED = "InvalidParameter.PasswordNotSupported"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"
//	INVALIDPARAMETERVALUE_CORECOUNTVALUE = "InvalidParameterValue.CoreCountValue"
//	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"
//	INVALIDPARAMETERVALUE_INSTANCETYPEREQUIREDHPCCLUSTER = "InvalidParameterValue.InstanceTypeRequiredHpcCluster"
//	INVALIDPARAMETERVALUE_INSUFFICIENTOFFERING = "InvalidParameterValue.InsufficientOffering"
//	INVALIDPARAMETERVALUE_INSUFFICIENTPRICE = "InvalidParameterValue.InsufficientPrice"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATEDESCRIPTION = "InvalidParameterValue.InvalidLaunchTemplateDescription"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATENAME = "InvalidParameterValue.InvalidLaunchTemplateName"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATEVERSIONDESCRIPTION = "InvalidParameterValue.InvalidLaunchTemplateVersionDescription"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"
//	INVALIDPARAMETERVALUE_SUBNETNOTEXIST = "InvalidParameterValue.SubnetNotExist"
//	INVALIDPARAMETERVALUE_THREADPERCOREVALUE = "InvalidParameterValue.ThreadPerCoreValue"
//	INVALIDPARAMETERVALUE_VPCIDZONEIDNOTMATCH = "InvalidParameterValue.VpcIdZoneIdNotMatch"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPASSWORD = "InvalidPassword"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"
//	LIMITEXCEEDED_LAUNCHTEMPLATEQUOTA = "LimitExceeded.LaunchTemplateQuota"
//	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"
//	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"
//	LIMITEXCEEDED_USERSPOTQUOTA = "LimitExceeded.UserSpotQuota"
//	LIMITEXCEEDED_VPCSUBNETNUM = "LimitExceeded.VpcSubnetNum"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_DPDKINSTANCETYPEREQUIREDVPC = "MissingParameter.DPDKInstanceTypeRequiredVPC"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"
//	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	RESOURCESSOLDOUT_EIPINSUFFICIENT = "ResourcesSoldOut.EipInsufficient"
//	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
//	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_NOTSUPPORTIMPORTINSTANCESACTIONTIMER = "UnsupportedOperation.NotSupportImportInstancesActionTimer"
//	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"
//	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"
//	VPCIPISUSED = "VpcIpIsUsed"
func (c *Client) CreateLaunchTemplate(request *CreateLaunchTemplateRequest) (response *CreateLaunchTemplateResponse, err error) {
	if request == nil {
		request = NewCreateLaunchTemplateRequest()
	}

	response = NewCreateLaunchTemplateResponse()
	err = c.Send(request, response)
	return
}

// CreateLaunchTemplate
// 本接口（CreateLaunchTemplate）用于创建实例启动模板。
//
// 实例启动模板是一种配置数据并可用于创建实例，其内容包含创建实例所需的配置，比如实例类型，数据盘和系统盘的类型和大小，以及安全组等信息。
//
// 初次创建实例模板后，其模板版本为默认版本1，新版本的创建可使用CreateLaunchTemplateVersion创建，版本号递增。默认情况下，在RunInstances中指定实例启动模板，若不指定模板版本号，则使用默认版本。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	INSTANCESQUOTALIMITEXCEEDED = "InstancesQuotaLimitExceeded"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_INVALIDIPFORMAT = "InvalidParameter.InvalidIpFormat"
//	INVALIDPARAMETER_LACKCORECOUNTORTHREADPERCORE = "InvalidParameter.LackCoreCountOrThreadPerCore"
//	INVALIDPARAMETER_PASSWORDNOTSUPPORTED = "InvalidParameter.PasswordNotSupported"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"
//	INVALIDPARAMETERVALUE_CORECOUNTVALUE = "InvalidParameterValue.CoreCountValue"
//	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"
//	INVALIDPARAMETERVALUE_INSTANCETYPEREQUIREDHPCCLUSTER = "InvalidParameterValue.InstanceTypeRequiredHpcCluster"
//	INVALIDPARAMETERVALUE_INSUFFICIENTOFFERING = "InvalidParameterValue.InsufficientOffering"
//	INVALIDPARAMETERVALUE_INSUFFICIENTPRICE = "InvalidParameterValue.InsufficientPrice"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATEDESCRIPTION = "InvalidParameterValue.InvalidLaunchTemplateDescription"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATENAME = "InvalidParameterValue.InvalidLaunchTemplateName"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATEVERSIONDESCRIPTION = "InvalidParameterValue.InvalidLaunchTemplateVersionDescription"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"
//	INVALIDPARAMETERVALUE_SUBNETNOTEXIST = "InvalidParameterValue.SubnetNotExist"
//	INVALIDPARAMETERVALUE_THREADPERCOREVALUE = "InvalidParameterValue.ThreadPerCoreValue"
//	INVALIDPARAMETERVALUE_VPCIDZONEIDNOTMATCH = "InvalidParameterValue.VpcIdZoneIdNotMatch"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPASSWORD = "InvalidPassword"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"
//	LIMITEXCEEDED_LAUNCHTEMPLATEQUOTA = "LimitExceeded.LaunchTemplateQuota"
//	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"
//	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"
//	LIMITEXCEEDED_USERSPOTQUOTA = "LimitExceeded.UserSpotQuota"
//	LIMITEXCEEDED_VPCSUBNETNUM = "LimitExceeded.VpcSubnetNum"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_DPDKINSTANCETYPEREQUIREDVPC = "MissingParameter.DPDKInstanceTypeRequiredVPC"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"
//	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	RESOURCESSOLDOUT_EIPINSUFFICIENT = "ResourcesSoldOut.EipInsufficient"
//	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
//	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_NOTSUPPORTIMPORTINSTANCESACTIONTIMER = "UnsupportedOperation.NotSupportImportInstancesActionTimer"
//	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"
//	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"
//	VPCIPISUSED = "VpcIpIsUsed"
func (c *Client) CreateLaunchTemplateWithContext(ctx context.Context, request *CreateLaunchTemplateRequest) (response *CreateLaunchTemplateResponse, err error) {
	if request == nil {
		request = NewCreateLaunchTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewCreateLaunchTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewCreateLaunchTemplateVersionRequest() (request *CreateLaunchTemplateVersionRequest) {
	request = &CreateLaunchTemplateVersionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "CreateLaunchTemplateVersion")

	return
}

func NewCreateLaunchTemplateVersionResponse() (response *CreateLaunchTemplateVersionResponse) {
	response = &CreateLaunchTemplateVersionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateLaunchTemplateVersion
// 本接口（CreateLaunchTemplateVersion）根据指定的实例模板ID以及对应的模板版本号创建新的实例启动模板，若未指定模板版本号则使用默认版本号。每个实例启动模板最多创建30个版本。
//
// 可能返回的错误码:
//
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_INVALIDIPFORMAT = "InvalidParameter.InvalidIpFormat"
//	INVALIDPARAMETER_PASSWORDNOTSUPPORTED = "InvalidParameter.PasswordNotSupported"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"
//	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATEVERSIONDESCRIPTION = "InvalidParameterValue.InvalidLaunchTemplateVersionDescription"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"
//	INVALIDPARAMETERVALUE_SUBNETNOTEXIST = "InvalidParameterValue.SubnetNotExist"
//	INVALIDPARAMETERVALUE_THREADPERCOREVALUE = "InvalidParameterValue.ThreadPerCoreValue"
//	INVALIDPARAMETERVALUE_VPCIDZONEIDNOTMATCH = "InvalidParameterValue.VpcIdZoneIdNotMatch"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPASSWORD = "InvalidPassword"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"
//	LIMITEXCEEDED_LAUNCHTEMPLATEQUOTA = "LimitExceeded.LaunchTemplateQuota"
//	LIMITEXCEEDED_LAUNCHTEMPLATEVERSIONQUOTA = "LimitExceeded.LaunchTemplateVersionQuota"
//	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"
//	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"
//	LIMITEXCEEDED_USERSPOTQUOTA = "LimitExceeded.UserSpotQuota"
//	LIMITEXCEEDED_VPCSUBNETNUM = "LimitExceeded.VpcSubnetNum"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_DPDKINSTANCETYPEREQUIREDVPC = "MissingParameter.DPDKInstanceTypeRequiredVPC"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"
//	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	RESOURCESSOLDOUT_EIPINSUFFICIENT = "ResourcesSoldOut.EipInsufficient"
//	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
//	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"
//	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"
//	VPCIPISUSED = "VpcIpIsUsed"
func (c *Client) CreateLaunchTemplateVersion(request *CreateLaunchTemplateVersionRequest) (response *CreateLaunchTemplateVersionResponse, err error) {
	if request == nil {
		request = NewCreateLaunchTemplateVersionRequest()
	}

	response = NewCreateLaunchTemplateVersionResponse()
	err = c.Send(request, response)
	return
}

// CreateLaunchTemplateVersion
// 本接口（CreateLaunchTemplateVersion）根据指定的实例模板ID以及对应的模板版本号创建新的实例启动模板，若未指定模板版本号则使用默认版本号。每个实例启动模板最多创建30个版本。
//
// 可能返回的错误码:
//
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_INVALIDIPFORMAT = "InvalidParameter.InvalidIpFormat"
//	INVALIDPARAMETER_PASSWORDNOTSUPPORTED = "InvalidParameter.PasswordNotSupported"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"
//	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATEVERSIONDESCRIPTION = "InvalidParameterValue.InvalidLaunchTemplateVersionDescription"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"
//	INVALIDPARAMETERVALUE_SUBNETNOTEXIST = "InvalidParameterValue.SubnetNotExist"
//	INVALIDPARAMETERVALUE_THREADPERCOREVALUE = "InvalidParameterValue.ThreadPerCoreValue"
//	INVALIDPARAMETERVALUE_VPCIDZONEIDNOTMATCH = "InvalidParameterValue.VpcIdZoneIdNotMatch"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPASSWORD = "InvalidPassword"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"
//	LIMITEXCEEDED_LAUNCHTEMPLATEQUOTA = "LimitExceeded.LaunchTemplateQuota"
//	LIMITEXCEEDED_LAUNCHTEMPLATEVERSIONQUOTA = "LimitExceeded.LaunchTemplateVersionQuota"
//	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"
//	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"
//	LIMITEXCEEDED_USERSPOTQUOTA = "LimitExceeded.UserSpotQuota"
//	LIMITEXCEEDED_VPCSUBNETNUM = "LimitExceeded.VpcSubnetNum"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_DPDKINSTANCETYPEREQUIREDVPC = "MissingParameter.DPDKInstanceTypeRequiredVPC"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"
//	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	RESOURCESSOLDOUT_EIPINSUFFICIENT = "ResourcesSoldOut.EipInsufficient"
//	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
//	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"
//	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"
//	VPCIPISUSED = "VpcIpIsUsed"
func (c *Client) CreateLaunchTemplateVersionWithContext(ctx context.Context, request *CreateLaunchTemplateVersionRequest) (response *CreateLaunchTemplateVersionResponse, err error) {
	if request == nil {
		request = NewCreateLaunchTemplateVersionRequest()
	}
	request.SetContext(ctx)

	response = NewCreateLaunchTemplateVersionResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteDisasterRecoverGroupsRequest() (request *DeleteDisasterRecoverGroupsRequest) {
	request = &DeleteDisasterRecoverGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DeleteDisasterRecoverGroups")

	return
}

func NewDeleteDisasterRecoverGroupsResponse() (response *DeleteDisasterRecoverGroupsResponse) {
	response = &DeleteDisasterRecoverGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteDisasterRecoverGroups
// 本接口 (DeleteDisasterRecoverGroups)用于删除[分散置放群组](https://cloud.tencent.com/document/product/213/15486)。只有空的置放群组才能被删除，非空的群组需要先销毁组内所有云服务器，才能执行删除操作，不然会产生删除置放群组失败的错误。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_PLACEMENTSETNOTEMPTY = "FailedOperation.PlacementSetNotEmpty"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCENOTFOUND_INVALIDPLACEMENTSET = "ResourceNotFound.InvalidPlacementSet"
func (c *Client) DeleteDisasterRecoverGroups(request *DeleteDisasterRecoverGroupsRequest) (response *DeleteDisasterRecoverGroupsResponse, err error) {
	if request == nil {
		request = NewDeleteDisasterRecoverGroupsRequest()
	}

	response = NewDeleteDisasterRecoverGroupsResponse()
	err = c.Send(request, response)
	return
}

// DeleteDisasterRecoverGroups
// 本接口 (DeleteDisasterRecoverGroups)用于删除[分散置放群组](https://cloud.tencent.com/document/product/213/15486)。只有空的置放群组才能被删除，非空的群组需要先销毁组内所有云服务器，才能执行删除操作，不然会产生删除置放群组失败的错误。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_PLACEMENTSETNOTEMPTY = "FailedOperation.PlacementSetNotEmpty"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCENOTFOUND_INVALIDPLACEMENTSET = "ResourceNotFound.InvalidPlacementSet"
func (c *Client) DeleteDisasterRecoverGroupsWithContext(ctx context.Context, request *DeleteDisasterRecoverGroupsRequest) (response *DeleteDisasterRecoverGroupsResponse, err error) {
	if request == nil {
		request = NewDeleteDisasterRecoverGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteDisasterRecoverGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteImagesRequest() (request *DeleteImagesRequest) {
	request = &DeleteImagesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DeleteImages")

	return
}

func NewDeleteImagesResponse() (response *DeleteImagesResponse) {
	response = &DeleteImagesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteImages
// 本接口（DeleteImages）用于删除一个或多个镜像。
//
// * 当[镜像状态](https://cloud.tencent.com/document/product/213/15753#Image)为`创建中`和`使用中`时, 不允许删除。镜像状态可以通过[DescribeImages](https://cloud.tencent.com/document/api/213/9418)获取。
//
// * 每个地域最多只支持创建10个自定义镜像，删除镜像可以释放账户的配额。
//
// * 当镜像正在被其它账户分享时，不允许删除。
//
// 可能返回的错误码:
//
//	INVALIDIMAGEID_INSHARED = "InvalidImageId.InShared"
//	INVALIDIMAGEID_INCORRECTSTATE = "InvalidImageId.IncorrectState"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
func (c *Client) DeleteImages(request *DeleteImagesRequest) (response *DeleteImagesResponse, err error) {
	if request == nil {
		request = NewDeleteImagesRequest()
	}

	response = NewDeleteImagesResponse()
	err = c.Send(request, response)
	return
}

// DeleteImages
// 本接口（DeleteImages）用于删除一个或多个镜像。
//
// * 当[镜像状态](https://cloud.tencent.com/document/product/213/15753#Image)为`创建中`和`使用中`时, 不允许删除。镜像状态可以通过[DescribeImages](https://cloud.tencent.com/document/api/213/9418)获取。
//
// * 每个地域最多只支持创建10个自定义镜像，删除镜像可以释放账户的配额。
//
// * 当镜像正在被其它账户分享时，不允许删除。
//
// 可能返回的错误码:
//
//	INVALIDIMAGEID_INSHARED = "InvalidImageId.InShared"
//	INVALIDIMAGEID_INCORRECTSTATE = "InvalidImageId.IncorrectState"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
func (c *Client) DeleteImagesWithContext(ctx context.Context, request *DeleteImagesRequest) (response *DeleteImagesResponse, err error) {
	if request == nil {
		request = NewDeleteImagesRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteImagesResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteKeyPairsRequest() (request *DeleteKeyPairsRequest) {
	request = &DeleteKeyPairsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DeleteKeyPairs")

	return
}

func NewDeleteKeyPairsResponse() (response *DeleteKeyPairsResponse) {
	response = &DeleteKeyPairsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteKeyPairs
// 本接口 (DeleteKeyPairs) 用于删除已在腾讯云托管的密钥对。
//
// * 可以同时删除多个密钥对。
//
// * 不能删除已被实例或镜像引用的密钥对，所以需要独立判断是否所有密钥对都被成功删除。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_KEYPAIRNOTSUPPORTED = "InvalidParameterValue.KeyPairNotSupported"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DeleteKeyPairs(request *DeleteKeyPairsRequest) (response *DeleteKeyPairsResponse, err error) {
	if request == nil {
		request = NewDeleteKeyPairsRequest()
	}

	response = NewDeleteKeyPairsResponse()
	err = c.Send(request, response)
	return
}

// DeleteKeyPairs
// 本接口 (DeleteKeyPairs) 用于删除已在腾讯云托管的密钥对。
//
// * 可以同时删除多个密钥对。
//
// * 不能删除已被实例或镜像引用的密钥对，所以需要独立判断是否所有密钥对都被成功删除。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_KEYPAIRNOTSUPPORTED = "InvalidParameterValue.KeyPairNotSupported"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DeleteKeyPairsWithContext(ctx context.Context, request *DeleteKeyPairsRequest) (response *DeleteKeyPairsResponse, err error) {
	if request == nil {
		request = NewDeleteKeyPairsRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteKeyPairsResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteLaunchTemplateRequest() (request *DeleteLaunchTemplateRequest) {
	request = &DeleteLaunchTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DeleteLaunchTemplate")

	return
}

func NewDeleteLaunchTemplateResponse() (response *DeleteLaunchTemplateResponse) {
	response = &DeleteLaunchTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteLaunchTemplate
// 本接口（DeleteLaunchTemplate）用于删除一个实例启动模板。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
func (c *Client) DeleteLaunchTemplate(request *DeleteLaunchTemplateRequest) (response *DeleteLaunchTemplateResponse, err error) {
	if request == nil {
		request = NewDeleteLaunchTemplateRequest()
	}

	response = NewDeleteLaunchTemplateResponse()
	err = c.Send(request, response)
	return
}

// DeleteLaunchTemplate
// 本接口（DeleteLaunchTemplate）用于删除一个实例启动模板。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
func (c *Client) DeleteLaunchTemplateWithContext(ctx context.Context, request *DeleteLaunchTemplateRequest) (response *DeleteLaunchTemplateResponse, err error) {
	if request == nil {
		request = NewDeleteLaunchTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteLaunchTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteLaunchTemplateVersionsRequest() (request *DeleteLaunchTemplateVersionsRequest) {
	request = &DeleteLaunchTemplateVersionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DeleteLaunchTemplateVersions")

	return
}

func NewDeleteLaunchTemplateVersionsResponse() (response *DeleteLaunchTemplateVersionsResponse) {
	response = &DeleteLaunchTemplateVersionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteLaunchTemplateVersions
// 本接口（DeleteLaunchTemplateVersions）用于删除一个或者多个实例启动模板版本。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEDEFAULTVERSION = "InvalidParameterValue.LaunchTemplateDefaultVersion"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	MISSINGPARAMETER = "MissingParameter"
//	UNKNOWNPARAMETER = "UnknownParameter"
func (c *Client) DeleteLaunchTemplateVersions(request *DeleteLaunchTemplateVersionsRequest) (response *DeleteLaunchTemplateVersionsResponse, err error) {
	if request == nil {
		request = NewDeleteLaunchTemplateVersionsRequest()
	}

	response = NewDeleteLaunchTemplateVersionsResponse()
	err = c.Send(request, response)
	return
}

// DeleteLaunchTemplateVersions
// 本接口（DeleteLaunchTemplateVersions）用于删除一个或者多个实例启动模板版本。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEDEFAULTVERSION = "InvalidParameterValue.LaunchTemplateDefaultVersion"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	MISSINGPARAMETER = "MissingParameter"
//	UNKNOWNPARAMETER = "UnknownParameter"
func (c *Client) DeleteLaunchTemplateVersionsWithContext(ctx context.Context, request *DeleteLaunchTemplateVersionsRequest) (response *DeleteLaunchTemplateVersionsResponse, err error) {
	if request == nil {
		request = NewDeleteLaunchTemplateVersionsRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteLaunchTemplateVersionsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAccountQuotaRequest() (request *DescribeAccountQuotaRequest) {
	request = &DescribeAccountQuotaRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeAccountQuota")

	return
}

func NewDescribeAccountQuotaResponse() (response *DescribeAccountQuotaResponse) {
	response = &DescribeAccountQuotaResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAccountQuota
// 本接口(DescribeAccountQuota)用于查询用户配额详情。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
func (c *Client) DescribeAccountQuota(request *DescribeAccountQuotaRequest) (response *DescribeAccountQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeAccountQuotaRequest()
	}

	response = NewDescribeAccountQuotaResponse()
	err = c.Send(request, response)
	return
}

// DescribeAccountQuota
// 本接口(DescribeAccountQuota)用于查询用户配额详情。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
func (c *Client) DescribeAccountQuotaWithContext(ctx context.Context, request *DescribeAccountQuotaRequest) (response *DescribeAccountQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeAccountQuotaRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAccountQuotaResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeDisasterRecoverGroupQuotaRequest() (request *DescribeDisasterRecoverGroupQuotaRequest) {
	request = &DescribeDisasterRecoverGroupQuotaRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeDisasterRecoverGroupQuota")

	return
}

func NewDescribeDisasterRecoverGroupQuotaResponse() (response *DescribeDisasterRecoverGroupQuotaResponse) {
	response = &DescribeDisasterRecoverGroupQuotaResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeDisasterRecoverGroupQuota
// 本接口 (DescribeDisasterRecoverGroupQuota)用于查询[分散置放群组](https://cloud.tencent.com/document/product/213/15486)配额。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
func (c *Client) DescribeDisasterRecoverGroupQuota(request *DescribeDisasterRecoverGroupQuotaRequest) (response *DescribeDisasterRecoverGroupQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeDisasterRecoverGroupQuotaRequest()
	}

	response = NewDescribeDisasterRecoverGroupQuotaResponse()
	err = c.Send(request, response)
	return
}

// DescribeDisasterRecoverGroupQuota
// 本接口 (DescribeDisasterRecoverGroupQuota)用于查询[分散置放群组](https://cloud.tencent.com/document/product/213/15486)配额。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
func (c *Client) DescribeDisasterRecoverGroupQuotaWithContext(ctx context.Context, request *DescribeDisasterRecoverGroupQuotaRequest) (response *DescribeDisasterRecoverGroupQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeDisasterRecoverGroupQuotaRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeDisasterRecoverGroupQuotaResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeDisasterRecoverGroupsRequest() (request *DescribeDisasterRecoverGroupsRequest) {
	request = &DescribeDisasterRecoverGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeDisasterRecoverGroups")

	return
}

func NewDescribeDisasterRecoverGroupsResponse() (response *DescribeDisasterRecoverGroupsResponse) {
	response = &DescribeDisasterRecoverGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeDisasterRecoverGroups
// 本接口 (DescribeDisasterRecoverGroups)用于查询[分散置放群组](https://cloud.tencent.com/document/product/213/15486)信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
func (c *Client) DescribeDisasterRecoverGroups(request *DescribeDisasterRecoverGroupsRequest) (response *DescribeDisasterRecoverGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeDisasterRecoverGroupsRequest()
	}

	response = NewDescribeDisasterRecoverGroupsResponse()
	err = c.Send(request, response)
	return
}

// DescribeDisasterRecoverGroups
// 本接口 (DescribeDisasterRecoverGroups)用于查询[分散置放群组](https://cloud.tencent.com/document/product/213/15486)信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
func (c *Client) DescribeDisasterRecoverGroupsWithContext(ctx context.Context, request *DescribeDisasterRecoverGroupsRequest) (response *DescribeDisasterRecoverGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeDisasterRecoverGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeDisasterRecoverGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeHostsRequest() (request *DescribeHostsRequest) {
	request = &DescribeHostsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeHosts")

	return
}

func NewDescribeHostsResponse() (response *DescribeHostsResponse) {
	response = &DescribeHostsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeHosts
// 本接口 (DescribeHosts) 用于获取一个或多个CDH实例的详细信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeHosts(request *DescribeHostsRequest) (response *DescribeHostsResponse, err error) {
	if request == nil {
		request = NewDescribeHostsRequest()
	}

	response = NewDescribeHostsResponse()
	err = c.Send(request, response)
	return
}

// DescribeHosts
// 本接口 (DescribeHosts) 用于获取一个或多个CDH实例的详细信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeHostsWithContext(ctx context.Context, request *DescribeHostsRequest) (response *DescribeHostsResponse, err error) {
	if request == nil {
		request = NewDescribeHostsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeHostsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeImageQuotaRequest() (request *DescribeImageQuotaRequest) {
	request = &DescribeImageQuotaRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeImageQuota")

	return
}

func NewDescribeImageQuotaResponse() (response *DescribeImageQuotaResponse) {
	response = &DescribeImageQuotaResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeImageQuota
// 本接口(DescribeImageQuota)用于查询用户帐号的镜像配额。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeImageQuota(request *DescribeImageQuotaRequest) (response *DescribeImageQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeImageQuotaRequest()
	}

	response = NewDescribeImageQuotaResponse()
	err = c.Send(request, response)
	return
}

// DescribeImageQuota
// 本接口(DescribeImageQuota)用于查询用户帐号的镜像配额。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeImageQuotaWithContext(ctx context.Context, request *DescribeImageQuotaRequest) (response *DescribeImageQuotaResponse, err error) {
	if request == nil {
		request = NewDescribeImageQuotaRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeImageQuotaResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeImageSharePermissionRequest() (request *DescribeImageSharePermissionRequest) {
	request = &DescribeImageSharePermissionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeImageSharePermission")

	return
}

func NewDescribeImageSharePermissionResponse() (response *DescribeImageSharePermissionResponse) {
	response = &DescribeImageSharePermissionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeImageSharePermission
// 本接口（DescribeImageSharePermission）用于查询镜像分享信息。
//
// 可能返回的错误码:
//
//	INVALIDACCOUNTID_NOTFOUND = "InvalidAccountId.NotFound"
//	INVALIDACCOUNTIS_YOURSELF = "InvalidAccountIs.YourSelf"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	OVERQUOTA = "OverQuota"
//	UNAUTHORIZEDOPERATION_IMAGENOTBELONGTOACCOUNT = "UnauthorizedOperation.ImageNotBelongToAccount"
func (c *Client) DescribeImageSharePermission(request *DescribeImageSharePermissionRequest) (response *DescribeImageSharePermissionResponse, err error) {
	if request == nil {
		request = NewDescribeImageSharePermissionRequest()
	}

	response = NewDescribeImageSharePermissionResponse()
	err = c.Send(request, response)
	return
}

// DescribeImageSharePermission
// 本接口（DescribeImageSharePermission）用于查询镜像分享信息。
//
// 可能返回的错误码:
//
//	INVALIDACCOUNTID_NOTFOUND = "InvalidAccountId.NotFound"
//	INVALIDACCOUNTIS_YOURSELF = "InvalidAccountIs.YourSelf"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	OVERQUOTA = "OverQuota"
//	UNAUTHORIZEDOPERATION_IMAGENOTBELONGTOACCOUNT = "UnauthorizedOperation.ImageNotBelongToAccount"
func (c *Client) DescribeImageSharePermissionWithContext(ctx context.Context, request *DescribeImageSharePermissionRequest) (response *DescribeImageSharePermissionResponse, err error) {
	if request == nil {
		request = NewDescribeImageSharePermissionRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeImageSharePermissionResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeImagesRequest() (request *DescribeImagesRequest) {
	request = &DescribeImagesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeImages")

	return
}

func NewDescribeImagesResponse() (response *DescribeImagesResponse) {
	response = &DescribeImagesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeImages
// 本接口(DescribeImages) 用于查看镜像列表。
//
// * 可以通过指定镜像ID来查询指定镜像的详细信息，或通过设定过滤器来查询满足过滤条件的镜像的详细信息。
//
// * 指定偏移(Offset)和限制(Limit)来选择结果中的一部分，默认返回满足条件的前20个镜像信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INVALIDPARAMETERCOEXISTIMAGEIDSFILTERS = "InvalidParameter.InvalidParameterCoexistImageIdsFilters"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INVALIDPARAMETERVALUELIMIT = "InvalidParameterValue.InvalidParameterValueLimit"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNAUTHORIZEDOPERATION_PERMISSIONDENIED = "UnauthorizedOperation.PermissionDenied"
func (c *Client) DescribeImages(request *DescribeImagesRequest) (response *DescribeImagesResponse, err error) {
	if request == nil {
		request = NewDescribeImagesRequest()
	}

	response = NewDescribeImagesResponse()
	err = c.Send(request, response)
	return
}

// DescribeImages
// 本接口(DescribeImages) 用于查看镜像列表。
//
// * 可以通过指定镜像ID来查询指定镜像的详细信息，或通过设定过滤器来查询满足过滤条件的镜像的详细信息。
//
// * 指定偏移(Offset)和限制(Limit)来选择结果中的一部分，默认返回满足条件的前20个镜像信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INVALIDPARAMETERCOEXISTIMAGEIDSFILTERS = "InvalidParameter.InvalidParameterCoexistImageIdsFilters"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INVALIDPARAMETERVALUELIMIT = "InvalidParameterValue.InvalidParameterValueLimit"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNAUTHORIZEDOPERATION_PERMISSIONDENIED = "UnauthorizedOperation.PermissionDenied"
func (c *Client) DescribeImagesWithContext(ctx context.Context, request *DescribeImagesRequest) (response *DescribeImagesResponse, err error) {
	if request == nil {
		request = NewDescribeImagesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeImagesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeImportImageOsRequest() (request *DescribeImportImageOsRequest) {
	request = &DescribeImportImageOsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeImportImageOs")

	return
}

func NewDescribeImportImageOsResponse() (response *DescribeImportImageOsResponse) {
	response = &DescribeImportImageOsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeImportImageOs
// 查看可以导入的镜像操作系统信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INVALIDPARAMETERCOEXISTIMAGEIDSFILTERS = "InvalidParameter.InvalidParameterCoexistImageIdsFilters"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INVALIDPARAMETERVALUELIMIT = "InvalidParameterValue.InvalidParameterValueLimit"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNAUTHORIZEDOPERATION_PERMISSIONDENIED = "UnauthorizedOperation.PermissionDenied"
func (c *Client) DescribeImportImageOs(request *DescribeImportImageOsRequest) (response *DescribeImportImageOsResponse, err error) {
	if request == nil {
		request = NewDescribeImportImageOsRequest()
	}

	response = NewDescribeImportImageOsResponse()
	err = c.Send(request, response)
	return
}

// DescribeImportImageOs
// 查看可以导入的镜像操作系统信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INVALIDPARAMETERCOEXISTIMAGEIDSFILTERS = "InvalidParameter.InvalidParameterCoexistImageIdsFilters"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INVALIDPARAMETERVALUELIMIT = "InvalidParameterValue.InvalidParameterValueLimit"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNAUTHORIZEDOPERATION_PERMISSIONDENIED = "UnauthorizedOperation.PermissionDenied"
func (c *Client) DescribeImportImageOsWithContext(ctx context.Context, request *DescribeImportImageOsRequest) (response *DescribeImportImageOsResponse, err error) {
	if request == nil {
		request = NewDescribeImportImageOsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeImportImageOsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInstanceFamilyConfigsRequest() (request *DescribeInstanceFamilyConfigsRequest) {
	request = &DescribeInstanceFamilyConfigsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInstanceFamilyConfigs")

	return
}

func NewDescribeInstanceFamilyConfigsResponse() (response *DescribeInstanceFamilyConfigsResponse) {
	response = &DescribeInstanceFamilyConfigsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstanceFamilyConfigs
// 本接口（DescribeInstanceFamilyConfigs）查询当前用户和地域所支持的机型族列表信息。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
func (c *Client) DescribeInstanceFamilyConfigs(request *DescribeInstanceFamilyConfigsRequest) (response *DescribeInstanceFamilyConfigsResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceFamilyConfigsRequest()
	}

	response = NewDescribeInstanceFamilyConfigsResponse()
	err = c.Send(request, response)
	return
}

// DescribeInstanceFamilyConfigs
// 本接口（DescribeInstanceFamilyConfigs）查询当前用户和地域所支持的机型族列表信息。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
func (c *Client) DescribeInstanceFamilyConfigsWithContext(ctx context.Context, request *DescribeInstanceFamilyConfigsRequest) (response *DescribeInstanceFamilyConfigsResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceFamilyConfigsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInstanceFamilyConfigsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInstanceInternetBandwidthConfigsRequest() (request *DescribeInstanceInternetBandwidthConfigsRequest) {
	request = &DescribeInstanceInternetBandwidthConfigsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInstanceInternetBandwidthConfigs")

	return
}

func NewDescribeInstanceInternetBandwidthConfigsResponse() (response *DescribeInstanceInternetBandwidthConfigsResponse) {
	response = &DescribeInstanceInternetBandwidthConfigsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstanceInternetBandwidthConfigs
// 本接口 (DescribeInstanceInternetBandwidthConfigs) 用于查询实例带宽配置。
//
// * 只支持查询`BANDWIDTH_PREPAID`（ 预付费按带宽结算 ）计费模式的带宽配置。
//
// * 接口返回实例的所有带宽配置信息（包含历史的带宽配置信息）。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOTFOUNDEIP = "FailedOperation.NotFoundEIP"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeInstanceInternetBandwidthConfigs(request *DescribeInstanceInternetBandwidthConfigsRequest) (response *DescribeInstanceInternetBandwidthConfigsResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceInternetBandwidthConfigsRequest()
	}

	response = NewDescribeInstanceInternetBandwidthConfigsResponse()
	err = c.Send(request, response)
	return
}

// DescribeInstanceInternetBandwidthConfigs
// 本接口 (DescribeInstanceInternetBandwidthConfigs) 用于查询实例带宽配置。
//
// * 只支持查询`BANDWIDTH_PREPAID`（ 预付费按带宽结算 ）计费模式的带宽配置。
//
// * 接口返回实例的所有带宽配置信息（包含历史的带宽配置信息）。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOTFOUNDEIP = "FailedOperation.NotFoundEIP"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) DescribeInstanceInternetBandwidthConfigsWithContext(ctx context.Context, request *DescribeInstanceInternetBandwidthConfigsRequest) (response *DescribeInstanceInternetBandwidthConfigsResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceInternetBandwidthConfigsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInstanceInternetBandwidthConfigsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInstanceTypeConfigsRequest() (request *DescribeInstanceTypeConfigsRequest) {
	request = &DescribeInstanceTypeConfigsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInstanceTypeConfigs")

	return
}

func NewDescribeInstanceTypeConfigsResponse() (response *DescribeInstanceTypeConfigsResponse) {
	response = &DescribeInstanceTypeConfigsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstanceTypeConfigs
// 本接口 (DescribeInstanceTypeConfigs) 用于查询实例机型配置。
//
// * 可以根据`zone`、`instance-family`来查询实例机型配置。过滤条件详见过滤器[`Filter`](https://cloud.tencent.com/document/api/213/15753#Filter)。
//
// * 如果参数为空，返回指定地域的所有实例机型配置。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
func (c *Client) DescribeInstanceTypeConfigs(request *DescribeInstanceTypeConfigsRequest) (response *DescribeInstanceTypeConfigsResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceTypeConfigsRequest()
	}

	response = NewDescribeInstanceTypeConfigsResponse()
	err = c.Send(request, response)
	return
}

// DescribeInstanceTypeConfigs
// 本接口 (DescribeInstanceTypeConfigs) 用于查询实例机型配置。
//
// * 可以根据`zone`、`instance-family`来查询实例机型配置。过滤条件详见过滤器[`Filter`](https://cloud.tencent.com/document/api/213/15753#Filter)。
//
// * 如果参数为空，返回指定地域的所有实例机型配置。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
func (c *Client) DescribeInstanceTypeConfigsWithContext(ctx context.Context, request *DescribeInstanceTypeConfigsRequest) (response *DescribeInstanceTypeConfigsResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceTypeConfigsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInstanceTypeConfigsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInstanceVncUrlRequest() (request *DescribeInstanceVncUrlRequest) {
	request = &DescribeInstanceVncUrlRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInstanceVncUrl")

	return
}

func NewDescribeInstanceVncUrlResponse() (response *DescribeInstanceVncUrlResponse) {
	response = &DescribeInstanceVncUrlResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstanceVncUrl
// 本接口 ( DescribeInstanceVncUrl ) 用于查询实例管理终端地址，获取的地址可用于实例的 VNC 登录。
//
// * 处于 `STOPPED` 状态的机器无法使用此功能。
//
// * 管理终端地址的有效期为 15 秒，调用接口成功后如果 15 秒内不使用该链接进行访问，管理终端地址自动失效，您需要重新查询。
//
// * 管理终端地址一旦被访问，将自动失效，您需要重新查询。
//
// * 如果连接断开，每分钟内重新连接的次数不能超过 30 次。
//
// * 获取到 `InstanceVncUrl` 后，您需要在链接 <https://img.qcloud.com/qcloud/app/active_vnc/index.html?> 末尾加上参数 `InstanceVncUrl=xxxx`  。
//
//   - 参数 `InstanceVncUrl` ：调用接口成功后会返回的 `InstanceVncUrl` 的值。
//
//     最后组成的 URL 格式如下：
//
// “`
//
// https://img.qcloud.com/qcloud/app/active_vnc/index.html?InstanceVncUrl=wss%3A%2F%2Fbjvnc.qcloud.com%3A26789%2Fvnc%3Fs%3DaHpjWnRVMFNhYmxKdDM5MjRHNlVTSVQwajNUSW0wb2tBbmFtREFCTmFrcy8vUUNPMG0wSHZNOUUxRm5PMmUzWmFDcWlOdDJIbUJxSTZDL0RXcHZxYnZZMmRkWWZWcEZia2lyb09XMzdKNmM9
//
// “`
//
// 可能返回的错误码:
//
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCESTATE = "InvalidInstanceState"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) DescribeInstanceVncUrl(request *DescribeInstanceVncUrlRequest) (response *DescribeInstanceVncUrlResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceVncUrlRequest()
	}

	response = NewDescribeInstanceVncUrlResponse()
	err = c.Send(request, response)
	return
}

// DescribeInstanceVncUrl
// 本接口 ( DescribeInstanceVncUrl ) 用于查询实例管理终端地址，获取的地址可用于实例的 VNC 登录。
//
// * 处于 `STOPPED` 状态的机器无法使用此功能。
//
// * 管理终端地址的有效期为 15 秒，调用接口成功后如果 15 秒内不使用该链接进行访问，管理终端地址自动失效，您需要重新查询。
//
// * 管理终端地址一旦被访问，将自动失效，您需要重新查询。
//
// * 如果连接断开，每分钟内重新连接的次数不能超过 30 次。
//
// * 获取到 `InstanceVncUrl` 后，您需要在链接 <https://img.qcloud.com/qcloud/app/active_vnc/index.html?> 末尾加上参数 `InstanceVncUrl=xxxx`  。
//
//   - 参数 `InstanceVncUrl` ：调用接口成功后会返回的 `InstanceVncUrl` 的值。
//
//     最后组成的 URL 格式如下：
//
// “`
//
// https://img.qcloud.com/qcloud/app/active_vnc/index.html?InstanceVncUrl=wss%3A%2F%2Fbjvnc.qcloud.com%3A26789%2Fvnc%3Fs%3DaHpjWnRVMFNhYmxKdDM5MjRHNlVTSVQwajNUSW0wb2tBbmFtREFCTmFrcy8vUUNPMG0wSHZNOUUxRm5PMmUzWmFDcWlOdDJIbUJxSTZDL0RXcHZxYnZZMmRkWWZWcEZia2lyb09XMzdKNmM9
//
// “`
//
// 可能返回的错误码:
//
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCESTATE = "InvalidInstanceState"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) DescribeInstanceVncUrlWithContext(ctx context.Context, request *DescribeInstanceVncUrlRequest) (response *DescribeInstanceVncUrlResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceVncUrlRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInstanceVncUrlResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInstancesRequest() (request *DescribeInstancesRequest) {
	request = &DescribeInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInstances")

	return
}

func NewDescribeInstancesResponse() (response *DescribeInstancesResponse) {
	response = &DescribeInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstances
// 本接口 (DescribeInstances) 用于查询一个或多个实例的详细信息。
//
// * 可以根据实例`ID`、实例名称或者实例计费模式等信息来查询实例的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的实例。
//
// * 支持查询实例的最新操作（LatestOperation）以及最新操作状态(LatestOperationState)。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_ILLEGALTAGKEY = "FailedOperation.IllegalTagKey"
//	FAILEDOPERATION_ILLEGALTAGVALUE = "FailedOperation.IllegalTagValue"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_IPADDRESSMALFORMED = "InvalidParameterValue.IPAddressMalformed"
//	INVALIDPARAMETERVALUE_IPV6ADDRESSMALFORMED = "InvalidParameterValue.IPv6AddressMalformed"
//	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_INVALIDVAGUENAME = "InvalidParameterValue.InvalidVagueName"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_SUBNETIDMALFORMED = "InvalidParameterValue.SubnetIdMalformed"
//	INVALIDPARAMETERVALUE_TAGKEYNOTFOUND = "InvalidParameterValue.TagKeyNotFound"
//	INVALIDPARAMETERVALUE_VPCIDMALFORMED = "InvalidParameterValue.VpcIdMalformed"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDSGID_MALFORMED = "InvalidSgId.Malformed"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	UNAUTHORIZEDOPERATION_INVALIDTOKEN = "UnauthorizedOperation.InvalidToken"
func (c *Client) DescribeInstances(request *DescribeInstancesRequest) (response *DescribeInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeInstancesRequest()
	}

	response = NewDescribeInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeInstances
// 本接口 (DescribeInstances) 用于查询一个或多个实例的详细信息。
//
// * 可以根据实例`ID`、实例名称或者实例计费模式等信息来查询实例的详细信息。过滤信息详细请见过滤器`Filter`。
//
// * 如果参数为空，返回当前用户一定数量（`Limit`所指定的数量，默认为20）的实例。
//
// * 支持查询实例的最新操作（LatestOperation）以及最新操作状态(LatestOperationState)。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_ILLEGALTAGKEY = "FailedOperation.IllegalTagKey"
//	FAILEDOPERATION_ILLEGALTAGVALUE = "FailedOperation.IllegalTagValue"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_IPADDRESSMALFORMED = "InvalidParameterValue.IPAddressMalformed"
//	INVALIDPARAMETERVALUE_IPV6ADDRESSMALFORMED = "InvalidParameterValue.IPv6AddressMalformed"
//	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_INVALIDVAGUENAME = "InvalidParameterValue.InvalidVagueName"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_SUBNETIDMALFORMED = "InvalidParameterValue.SubnetIdMalformed"
//	INVALIDPARAMETERVALUE_TAGKEYNOTFOUND = "InvalidParameterValue.TagKeyNotFound"
//	INVALIDPARAMETERVALUE_VPCIDMALFORMED = "InvalidParameterValue.VpcIdMalformed"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDSGID_MALFORMED = "InvalidSgId.Malformed"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	UNAUTHORIZEDOPERATION_INVALIDTOKEN = "UnauthorizedOperation.InvalidToken"
func (c *Client) DescribeInstancesWithContext(ctx context.Context, request *DescribeInstancesRequest) (response *DescribeInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInstancesModificationRequest() (request *DescribeInstancesModificationRequest) {
	request = &DescribeInstancesModificationRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInstancesModification")

	return
}

func NewDescribeInstancesModificationResponse() (response *DescribeInstancesModificationResponse) {
	response = &DescribeInstancesModificationResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstancesModification
// 本接口 (DescribeInstancesModification) 用于查询指定实例支持调整的机型配置。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILY = "UnsupportedOperation.UnsupportedChangeInstanceFamily"
func (c *Client) DescribeInstancesModification(request *DescribeInstancesModificationRequest) (response *DescribeInstancesModificationResponse, err error) {
	if request == nil {
		request = NewDescribeInstancesModificationRequest()
	}

	response = NewDescribeInstancesModificationResponse()
	err = c.Send(request, response)
	return
}

// DescribeInstancesModification
// 本接口 (DescribeInstancesModification) 用于查询指定实例支持调整的机型配置。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILY = "UnsupportedOperation.UnsupportedChangeInstanceFamily"
func (c *Client) DescribeInstancesModificationWithContext(ctx context.Context, request *DescribeInstancesModificationRequest) (response *DescribeInstancesModificationResponse, err error) {
	if request == nil {
		request = NewDescribeInstancesModificationRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInstancesModificationResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInstancesOperationLimitRequest() (request *DescribeInstancesOperationLimitRequest) {
	request = &DescribeInstancesOperationLimitRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInstancesOperationLimit")

	return
}

func NewDescribeInstancesOperationLimitResponse() (response *DescribeInstancesOperationLimitResponse) {
	response = &DescribeInstancesOperationLimitResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstancesOperationLimit
// 本接口（DescribeInstancesOperationLimit）用于查询实例操作限制。
//
// * 目前支持调整配置操作限制次数查询。
//
// 可能返回的错误码:
//
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
func (c *Client) DescribeInstancesOperationLimit(request *DescribeInstancesOperationLimitRequest) (response *DescribeInstancesOperationLimitResponse, err error) {
	if request == nil {
		request = NewDescribeInstancesOperationLimitRequest()
	}

	response = NewDescribeInstancesOperationLimitResponse()
	err = c.Send(request, response)
	return
}

// DescribeInstancesOperationLimit
// 本接口（DescribeInstancesOperationLimit）用于查询实例操作限制。
//
// * 目前支持调整配置操作限制次数查询。
//
// 可能返回的错误码:
//
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
func (c *Client) DescribeInstancesOperationLimitWithContext(ctx context.Context, request *DescribeInstancesOperationLimitRequest) (response *DescribeInstancesOperationLimitResponse, err error) {
	if request == nil {
		request = NewDescribeInstancesOperationLimitRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInstancesOperationLimitResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInstancesStatusRequest() (request *DescribeInstancesStatusRequest) {
	request = &DescribeInstancesStatusRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInstancesStatus")

	return
}

func NewDescribeInstancesStatusResponse() (response *DescribeInstancesStatusResponse) {
	response = &DescribeInstancesStatusResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstancesStatus
// 本接口 (DescribeInstancesStatus) 用于查询一个或多个实例的状态。
//
// * 可以根据实例`ID`来查询实例的状态。
//
// * 如果参数为空，返回当前用户一定数量（Limit所指定的数量，默认为20）的实例状态。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNAUTHORIZEDOPERATION_INVALIDTOKEN = "UnauthorizedOperation.InvalidToken"
func (c *Client) DescribeInstancesStatus(request *DescribeInstancesStatusRequest) (response *DescribeInstancesStatusResponse, err error) {
	if request == nil {
		request = NewDescribeInstancesStatusRequest()
	}

	response = NewDescribeInstancesStatusResponse()
	err = c.Send(request, response)
	return
}

// DescribeInstancesStatus
// 本接口 (DescribeInstancesStatus) 用于查询一个或多个实例的状态。
//
// * 可以根据实例`ID`来查询实例的状态。
//
// * 如果参数为空，返回当前用户一定数量（Limit所指定的数量，默认为20）的实例状态。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNAUTHORIZEDOPERATION_INVALIDTOKEN = "UnauthorizedOperation.InvalidToken"
func (c *Client) DescribeInstancesStatusWithContext(ctx context.Context, request *DescribeInstancesStatusRequest) (response *DescribeInstancesStatusResponse, err error) {
	if request == nil {
		request = NewDescribeInstancesStatusRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInstancesStatusResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeInternetChargeTypeConfigsRequest() (request *DescribeInternetChargeTypeConfigsRequest) {
	request = &DescribeInternetChargeTypeConfigsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeInternetChargeTypeConfigs")

	return
}

func NewDescribeInternetChargeTypeConfigsResponse() (response *DescribeInternetChargeTypeConfigsResponse) {
	response = &DescribeInternetChargeTypeConfigsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInternetChargeTypeConfigs
// 本接口（DescribeInternetChargeTypeConfigs）用于查询网络的计费类型。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNAUTHORIZEDOPERATION_INVALIDTOKEN = "UnauthorizedOperation.InvalidToken"
func (c *Client) DescribeInternetChargeTypeConfigs(request *DescribeInternetChargeTypeConfigsRequest) (response *DescribeInternetChargeTypeConfigsResponse, err error) {
	if request == nil {
		request = NewDescribeInternetChargeTypeConfigsRequest()
	}

	response = NewDescribeInternetChargeTypeConfigsResponse()
	err = c.Send(request, response)
	return
}

// DescribeInternetChargeTypeConfigs
// 本接口（DescribeInternetChargeTypeConfigs）用于查询网络的计费类型。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNAUTHORIZEDOPERATION_INVALIDTOKEN = "UnauthorizedOperation.InvalidToken"
func (c *Client) DescribeInternetChargeTypeConfigsWithContext(ctx context.Context, request *DescribeInternetChargeTypeConfigsRequest) (response *DescribeInternetChargeTypeConfigsResponse, err error) {
	if request == nil {
		request = NewDescribeInternetChargeTypeConfigsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeInternetChargeTypeConfigsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeKeyPairsRequest() (request *DescribeKeyPairsRequest) {
	request = &DescribeKeyPairsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeKeyPairs")

	return
}

func NewDescribeKeyPairsResponse() (response *DescribeKeyPairsResponse) {
	response = &DescribeKeyPairsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeKeyPairs
// 本接口 (DescribeKeyPairs) 用于查询密钥对信息。
//
// * 密钥对是通过一种算法生成的一对密钥，在生成的密钥对中，一个向外界公开，称为公钥；另一个用户自己保留，称为私钥。密钥对的公钥内容可以通过这个接口查询，但私钥内容系统不保留。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUELIMIT = "InvalidParameterValueLimit"
//	INVALIDPARAMETERVALUEOFFSET = "InvalidParameterValueOffset"
func (c *Client) DescribeKeyPairs(request *DescribeKeyPairsRequest) (response *DescribeKeyPairsResponse, err error) {
	if request == nil {
		request = NewDescribeKeyPairsRequest()
	}

	response = NewDescribeKeyPairsResponse()
	err = c.Send(request, response)
	return
}

// DescribeKeyPairs
// 本接口 (DescribeKeyPairs) 用于查询密钥对信息。
//
// * 密钥对是通过一种算法生成的一对密钥，在生成的密钥对中，一个向外界公开，称为公钥；另一个用户自己保留，称为私钥。密钥对的公钥内容可以通过这个接口查询，但私钥内容系统不保留。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUELIMIT = "InvalidParameterValueLimit"
//	INVALIDPARAMETERVALUEOFFSET = "InvalidParameterValueOffset"
func (c *Client) DescribeKeyPairsWithContext(ctx context.Context, request *DescribeKeyPairsRequest) (response *DescribeKeyPairsResponse, err error) {
	if request == nil {
		request = NewDescribeKeyPairsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeKeyPairsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeLaunchTemplateVersionsRequest() (request *DescribeLaunchTemplateVersionsRequest) {
	request = &DescribeLaunchTemplateVersionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeLaunchTemplateVersions")

	return
}

func NewDescribeLaunchTemplateVersionsResponse() (response *DescribeLaunchTemplateVersionsResponse) {
	response = &DescribeLaunchTemplateVersionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeLaunchTemplateVersions
// 本接口（DescribeLaunchTemplateVersions）用于查询实例模板版本信息。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_NOTSUPPORTED = "InvalidParameterValue.NotSupported"
//	MISSINGPARAMETER = "MissingParameter"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeLaunchTemplateVersions(request *DescribeLaunchTemplateVersionsRequest) (response *DescribeLaunchTemplateVersionsResponse, err error) {
	if request == nil {
		request = NewDescribeLaunchTemplateVersionsRequest()
	}

	response = NewDescribeLaunchTemplateVersionsResponse()
	err = c.Send(request, response)
	return
}

// DescribeLaunchTemplateVersions
// 本接口（DescribeLaunchTemplateVersions）用于查询实例模板版本信息。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_NOTSUPPORTED = "InvalidParameterValue.NotSupported"
//	MISSINGPARAMETER = "MissingParameter"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeLaunchTemplateVersionsWithContext(ctx context.Context, request *DescribeLaunchTemplateVersionsRequest) (response *DescribeLaunchTemplateVersionsResponse, err error) {
	if request == nil {
		request = NewDescribeLaunchTemplateVersionsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeLaunchTemplateVersionsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeLaunchTemplatesRequest() (request *DescribeLaunchTemplatesRequest) {
	request = &DescribeLaunchTemplatesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeLaunchTemplates")

	return
}

func NewDescribeLaunchTemplatesResponse() (response *DescribeLaunchTemplatesResponse) {
	response = &DescribeLaunchTemplatesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeLaunchTemplates
// 本接口（DescribeLaunchTemplates）用于查询一个或者多个实例启动模板。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATENAME = "InvalidParameterValue.InvalidLaunchTemplateName"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeLaunchTemplates(request *DescribeLaunchTemplatesRequest) (response *DescribeLaunchTemplatesResponse, err error) {
	if request == nil {
		request = NewDescribeLaunchTemplatesRequest()
	}

	response = NewDescribeLaunchTemplatesResponse()
	err = c.Send(request, response)
	return
}

// DescribeLaunchTemplates
// 本接口（DescribeLaunchTemplates）用于查询一个或者多个实例启动模板。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATENAME = "InvalidParameterValue.InvalidLaunchTemplateName"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeLaunchTemplatesWithContext(ctx context.Context, request *DescribeLaunchTemplatesRequest) (response *DescribeLaunchTemplatesResponse, err error) {
	if request == nil {
		request = NewDescribeLaunchTemplatesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeLaunchTemplatesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeRegionsRequest() (request *DescribeRegionsRequest) {
	request = &DescribeRegionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeRegions")

	return
}

func NewDescribeRegionsResponse() (response *DescribeRegionsResponse) {
	response = &DescribeRegionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeRegions
// 本接口(DescribeRegions)用于查询地域信息。因平台策略原因，该接口暂时停止更新，为确保您正常调用，可切换至新链接：https://cloud.tencent.com/document/product/1278/55255。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATENAME = "InvalidParameterValue.InvalidLaunchTemplateName"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeRegions(request *DescribeRegionsRequest) (response *DescribeRegionsResponse, err error) {
	if request == nil {
		request = NewDescribeRegionsRequest()
	}

	response = NewDescribeRegionsResponse()
	err = c.Send(request, response)
	return
}

// DescribeRegions
// 本接口(DescribeRegions)用于查询地域信息。因平台策略原因，该接口暂时停止更新，为确保您正常调用，可切换至新链接：https://cloud.tencent.com/document/product/1278/55255。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATENAME = "InvalidParameterValue.InvalidLaunchTemplateName"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeRegionsWithContext(ctx context.Context, request *DescribeRegionsRequest) (response *DescribeRegionsResponse, err error) {
	if request == nil {
		request = NewDescribeRegionsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeRegionsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeReservedInstancesRequest() (request *DescribeReservedInstancesRequest) {
	request = &DescribeReservedInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeReservedInstances")

	return
}

func NewDescribeReservedInstancesResponse() (response *DescribeReservedInstancesResponse) {
	response = &DescribeReservedInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeReservedInstances
// 本接口(DescribeReservedInstances)可提供列出用户已购买的预留实例
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"
func (c *Client) DescribeReservedInstances(request *DescribeReservedInstancesRequest) (response *DescribeReservedInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeReservedInstancesRequest()
	}

	response = NewDescribeReservedInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeReservedInstances
// 本接口(DescribeReservedInstances)可提供列出用户已购买的预留实例
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"
func (c *Client) DescribeReservedInstancesWithContext(ctx context.Context, request *DescribeReservedInstancesRequest) (response *DescribeReservedInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeReservedInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeReservedInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeReservedInstancesConfigInfosRequest() (request *DescribeReservedInstancesConfigInfosRequest) {
	request = &DescribeReservedInstancesConfigInfosRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeReservedInstancesConfigInfos")

	return
}

func NewDescribeReservedInstancesConfigInfosResponse() (response *DescribeReservedInstancesConfigInfosResponse) {
	response = &DescribeReservedInstancesConfigInfosResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeReservedInstancesConfigInfos
// 本接口(DescribeReservedInstancesConfigInfos)供用户列出可购买预留实例机型配置。预留实例当前只针对国际站白名单用户开放。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"
func (c *Client) DescribeReservedInstancesConfigInfos(request *DescribeReservedInstancesConfigInfosRequest) (response *DescribeReservedInstancesConfigInfosResponse, err error) {
	if request == nil {
		request = NewDescribeReservedInstancesConfigInfosRequest()
	}

	response = NewDescribeReservedInstancesConfigInfosResponse()
	err = c.Send(request, response)
	return
}

// DescribeReservedInstancesConfigInfos
// 本接口(DescribeReservedInstancesConfigInfos)供用户列出可购买预留实例机型配置。预留实例当前只针对国际站白名单用户开放。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"
func (c *Client) DescribeReservedInstancesConfigInfosWithContext(ctx context.Context, request *DescribeReservedInstancesConfigInfosRequest) (response *DescribeReservedInstancesConfigInfosResponse, err error) {
	if request == nil {
		request = NewDescribeReservedInstancesConfigInfosRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeReservedInstancesConfigInfosResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeReservedInstancesOfferingsRequest() (request *DescribeReservedInstancesOfferingsRequest) {
	request = &DescribeReservedInstancesOfferingsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeReservedInstancesOfferings")

	return
}

func NewDescribeReservedInstancesOfferingsResponse() (response *DescribeReservedInstancesOfferingsResponse) {
	response = &DescribeReservedInstancesOfferingsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeReservedInstancesOfferings
// 本接口(DescribeReservedInstancesOfferings)供用户列出可购买的预留实例配置
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"
func (c *Client) DescribeReservedInstancesOfferings(request *DescribeReservedInstancesOfferingsRequest) (response *DescribeReservedInstancesOfferingsResponse, err error) {
	if request == nil {
		request = NewDescribeReservedInstancesOfferingsRequest()
	}

	response = NewDescribeReservedInstancesOfferingsResponse()
	err = c.Send(request, response)
	return
}

// DescribeReservedInstancesOfferings
// 本接口(DescribeReservedInstancesOfferings)供用户列出可购买的预留实例配置
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"
func (c *Client) DescribeReservedInstancesOfferingsWithContext(ctx context.Context, request *DescribeReservedInstancesOfferingsRequest) (response *DescribeReservedInstancesOfferingsResponse, err error) {
	if request == nil {
		request = NewDescribeReservedInstancesOfferingsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeReservedInstancesOfferingsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeZoneInstanceConfigInfosRequest() (request *DescribeZoneInstanceConfigInfosRequest) {
	request = &DescribeZoneInstanceConfigInfosRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeZoneInstanceConfigInfos")

	return
}

func NewDescribeZoneInstanceConfigInfosResponse() (response *DescribeZoneInstanceConfigInfosResponse) {
	response = &DescribeZoneInstanceConfigInfosResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeZoneInstanceConfigInfos
// 本接口(DescribeZoneInstanceConfigInfos) 获取可用区的机型信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	RESOURCEINSUFFICIENT_AVAILABILITYZONESOLDOUT = "ResourceInsufficient.AvailabilityZoneSoldOut"
func (c *Client) DescribeZoneInstanceConfigInfos(request *DescribeZoneInstanceConfigInfosRequest) (response *DescribeZoneInstanceConfigInfosResponse, err error) {
	if request == nil {
		request = NewDescribeZoneInstanceConfigInfosRequest()
	}

	response = NewDescribeZoneInstanceConfigInfosResponse()
	err = c.Send(request, response)
	return
}

// DescribeZoneInstanceConfigInfos
// 本接口(DescribeZoneInstanceConfigInfos) 获取可用区的机型信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	RESOURCEINSUFFICIENT_AVAILABILITYZONESOLDOUT = "ResourceInsufficient.AvailabilityZoneSoldOut"
func (c *Client) DescribeZoneInstanceConfigInfosWithContext(ctx context.Context, request *DescribeZoneInstanceConfigInfosRequest) (response *DescribeZoneInstanceConfigInfosResponse, err error) {
	if request == nil {
		request = NewDescribeZoneInstanceConfigInfosRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeZoneInstanceConfigInfosResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeZonesRequest() (request *DescribeZonesRequest) {
	request = &DescribeZonesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DescribeZones")

	return
}

func NewDescribeZonesResponse() (response *DescribeZonesResponse) {
	response = &DescribeZonesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeZones
// 本接口(DescribeZones)用于查询可用区信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	RESOURCEINSUFFICIENT_AVAILABILITYZONESOLDOUT = "ResourceInsufficient.AvailabilityZoneSoldOut"
func (c *Client) DescribeZones(request *DescribeZonesRequest) (response *DescribeZonesResponse, err error) {
	if request == nil {
		request = NewDescribeZonesRequest()
	}

	response = NewDescribeZonesResponse()
	err = c.Send(request, response)
	return
}

// DescribeZones
// 本接口(DescribeZones)用于查询可用区信息。
//
// 可能返回的错误码:
//
//	INVALIDFILTER = "InvalidFilter"
//	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	RESOURCEINSUFFICIENT_AVAILABILITYZONESOLDOUT = "ResourceInsufficient.AvailabilityZoneSoldOut"
func (c *Client) DescribeZonesWithContext(ctx context.Context, request *DescribeZonesRequest) (response *DescribeZonesResponse, err error) {
	if request == nil {
		request = NewDescribeZonesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeZonesResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateInstancesKeyPairsRequest() (request *DisassociateInstancesKeyPairsRequest) {
	request = &DisassociateInstancesKeyPairsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DisassociateInstancesKeyPairs")

	return
}

func NewDisassociateInstancesKeyPairsResponse() (response *DisassociateInstancesKeyPairsResponse) {
	response = &DisassociateInstancesKeyPairsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateInstancesKeyPairs
// 本接口 (DisassociateInstancesKeyPairs) 用于解除实例的密钥绑定关系。
//
// * 只支持[`STOPPED`](https://cloud.tencent.com/document/product/213/15753#InstanceStatus)状态的`Linux`操作系统的实例。
//
// * 解绑密钥后，实例可以通过原来设置的密码登录。
//
// * 如果原来没有设置密码，解绑后将无法使用 `SSH` 登录。可以调用 [ResetInstancesPassword](https://cloud.tencent.com/document/api/213/15736) 接口来设置登录密码。
//
// * 支持批量操作。每次请求批量实例的上限为100。如果批量实例存在不允许操作的实例，操作会以特定错误码返回。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCEOSWINDOWS = "UnsupportedOperation.InstanceOsWindows"
//	UNSUPPORTEDOPERATION_INSTANCESTATEBANNING = "UnsupportedOperation.InstanceStateBanning"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) DisassociateInstancesKeyPairs(request *DisassociateInstancesKeyPairsRequest) (response *DisassociateInstancesKeyPairsResponse, err error) {
	if request == nil {
		request = NewDisassociateInstancesKeyPairsRequest()
	}

	response = NewDisassociateInstancesKeyPairsResponse()
	err = c.Send(request, response)
	return
}

// DisassociateInstancesKeyPairs
// 本接口 (DisassociateInstancesKeyPairs) 用于解除实例的密钥绑定关系。
//
// * 只支持[`STOPPED`](https://cloud.tencent.com/document/product/213/15753#InstanceStatus)状态的`Linux`操作系统的实例。
//
// * 解绑密钥后，实例可以通过原来设置的密码登录。
//
// * 如果原来没有设置密码，解绑后将无法使用 `SSH` 登录。可以调用 [ResetInstancesPassword](https://cloud.tencent.com/document/api/213/15736) 接口来设置登录密码。
//
// * 支持批量操作。每次请求批量实例的上限为100。如果批量实例存在不允许操作的实例，操作会以特定错误码返回。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCEOSWINDOWS = "UnsupportedOperation.InstanceOsWindows"
//	UNSUPPORTEDOPERATION_INSTANCESTATEBANNING = "UnsupportedOperation.InstanceStateBanning"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) DisassociateInstancesKeyPairsWithContext(ctx context.Context, request *DisassociateInstancesKeyPairsRequest) (response *DisassociateInstancesKeyPairsResponse, err error) {
	if request == nil {
		request = NewDisassociateInstancesKeyPairsRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateInstancesKeyPairsResponse()
	err = c.Send(request, response)
	return
}

func NewDisassociateSecurityGroupsRequest() (request *DisassociateSecurityGroupsRequest) {
	request = &DisassociateSecurityGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "DisassociateSecurityGroups")

	return
}

func NewDisassociateSecurityGroupsResponse() (response *DisassociateSecurityGroupsResponse) {
	response = &DisassociateSecurityGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisassociateSecurityGroups
// 本接口 (DisassociateSecurityGroups) 用于解绑实例的指定安全组。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDSGID_MALFORMED = "InvalidSgId.Malformed"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	SECGROUPACTIONFAILURE = "SecGroupActionFailure"
func (c *Client) DisassociateSecurityGroups(request *DisassociateSecurityGroupsRequest) (response *DisassociateSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewDisassociateSecurityGroupsRequest()
	}

	response = NewDisassociateSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

// DisassociateSecurityGroups
// 本接口 (DisassociateSecurityGroups) 用于解绑实例的指定安全组。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDSGID_MALFORMED = "InvalidSgId.Malformed"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	SECGROUPACTIONFAILURE = "SecGroupActionFailure"
func (c *Client) DisassociateSecurityGroupsWithContext(ctx context.Context, request *DisassociateSecurityGroupsRequest) (response *DisassociateSecurityGroupsResponse, err error) {
	if request == nil {
		request = NewDisassociateSecurityGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDisassociateSecurityGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewImportImageRequest() (request *ImportImageRequest) {
	request = &ImportImageRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ImportImage")

	return
}

func NewImportImageResponse() (response *ImportImageResponse) {
	response = &ImportImageResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ImportImage
// 本接口(ImportImage)用于导入镜像，导入后的镜像可用于创建实例。目前支持 RAW、VHD、QCOW2、VMDK 镜像格式。
//
// 可能返回的错误码:
//
//	IMAGEQUOTALIMITEXCEEDED = "ImageQuotaLimitExceeded"
//	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"
//	INVALIDIMAGEOSTYPE_UNSUPPORTED = "InvalidImageOsType.Unsupported"
//	INVALIDIMAGEOSVERSION_UNSUPPORTED = "InvalidImageOsVersion.Unsupported"
//	INVALIDPARAMETER_INVALIDPARAMETERURLERROR = "InvalidParameter.InvalidParameterUrlError"
//	INVALIDPARAMETERVALUE_TOOLARGE = "InvalidParameterValue.TooLarge"
//	REGIONABILITYLIMIT_UNSUPPORTEDTOIMPORTIMAGE = "RegionAbilityLimit.UnsupportedToImportImage"
func (c *Client) ImportImage(request *ImportImageRequest) (response *ImportImageResponse, err error) {
	if request == nil {
		request = NewImportImageRequest()
	}

	response = NewImportImageResponse()
	err = c.Send(request, response)
	return
}

// ImportImage
// 本接口(ImportImage)用于导入镜像，导入后的镜像可用于创建实例。目前支持 RAW、VHD、QCOW2、VMDK 镜像格式。
//
// 可能返回的错误码:
//
//	IMAGEQUOTALIMITEXCEEDED = "ImageQuotaLimitExceeded"
//	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"
//	INVALIDIMAGEOSTYPE_UNSUPPORTED = "InvalidImageOsType.Unsupported"
//	INVALIDIMAGEOSVERSION_UNSUPPORTED = "InvalidImageOsVersion.Unsupported"
//	INVALIDPARAMETER_INVALIDPARAMETERURLERROR = "InvalidParameter.InvalidParameterUrlError"
//	INVALIDPARAMETERVALUE_TOOLARGE = "InvalidParameterValue.TooLarge"
//	REGIONABILITYLIMIT_UNSUPPORTEDTOIMPORTIMAGE = "RegionAbilityLimit.UnsupportedToImportImage"
func (c *Client) ImportImageWithContext(ctx context.Context, request *ImportImageRequest) (response *ImportImageResponse, err error) {
	if request == nil {
		request = NewImportImageRequest()
	}
	request.SetContext(ctx)

	response = NewImportImageResponse()
	err = c.Send(request, response)
	return
}

func NewImportKeyPairRequest() (request *ImportKeyPairRequest) {
	request = &ImportKeyPairRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ImportKeyPair")

	return
}

func NewImportKeyPairResponse() (response *ImportKeyPairResponse) {
	response = &ImportKeyPairResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ImportKeyPair
// 本接口 (ImportKeyPair) 用于导入密钥对。
//
// * 本接口的功能是将密钥对导入到用户账户，并不会自动绑定到实例。如需绑定可以使用[AssociasteInstancesKeyPair](https://cloud.tencent.com/document/api/213/9404)接口。
//
// * 需指定密钥对名称以及该密钥对的公钥文本。
//
// * 如果用户只有私钥，可以通过 `SSL` 工具将私钥转换成公钥后再导入。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"
//	INVALIDKEYPAIRNAME_DUPLICATE = "InvalidKeyPairName.Duplicate"
//	INVALIDKEYPAIRNAMEEMPTY = "InvalidKeyPairNameEmpty"
//	INVALIDKEYPAIRNAMEINCLUDEILLEGALCHAR = "InvalidKeyPairNameIncludeIllegalChar"
//	INVALIDKEYPAIRNAMETOOLONG = "InvalidKeyPairNameTooLong"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDPUBLICKEY_DUPLICATE = "InvalidPublicKey.Duplicate"
//	INVALIDPUBLICKEY_MALFORMED = "InvalidPublicKey.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) ImportKeyPair(request *ImportKeyPairRequest) (response *ImportKeyPairResponse, err error) {
	if request == nil {
		request = NewImportKeyPairRequest()
	}

	response = NewImportKeyPairResponse()
	err = c.Send(request, response)
	return
}

// ImportKeyPair
// 本接口 (ImportKeyPair) 用于导入密钥对。
//
// * 本接口的功能是将密钥对导入到用户账户，并不会自动绑定到实例。如需绑定可以使用[AssociasteInstancesKeyPair](https://cloud.tencent.com/document/api/213/9404)接口。
//
// * 需指定密钥对名称以及该密钥对的公钥文本。
//
// * 如果用户只有私钥，可以通过 `SSL` 工具将私钥转换成公钥后再导入。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"
//	INVALIDKEYPAIRNAME_DUPLICATE = "InvalidKeyPairName.Duplicate"
//	INVALIDKEYPAIRNAMEEMPTY = "InvalidKeyPairNameEmpty"
//	INVALIDKEYPAIRNAMEINCLUDEILLEGALCHAR = "InvalidKeyPairNameIncludeIllegalChar"
//	INVALIDKEYPAIRNAMETOOLONG = "InvalidKeyPairNameTooLong"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDPUBLICKEY_DUPLICATE = "InvalidPublicKey.Duplicate"
//	INVALIDPUBLICKEY_MALFORMED = "InvalidPublicKey.Malformed"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) ImportKeyPairWithContext(ctx context.Context, request *ImportKeyPairRequest) (response *ImportKeyPairResponse, err error) {
	if request == nil {
		request = NewImportKeyPairRequest()
	}
	request.SetContext(ctx)

	response = NewImportKeyPairResponse()
	err = c.Send(request, response)
	return
}

func NewInquirePricePurchaseReservedInstancesOfferingRequest() (request *InquirePricePurchaseReservedInstancesOfferingRequest) {
	request = &InquirePricePurchaseReservedInstancesOfferingRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquirePricePurchaseReservedInstancesOffering")

	return
}

func NewInquirePricePurchaseReservedInstancesOfferingResponse() (response *InquirePricePurchaseReservedInstancesOfferingResponse) {
	response = &InquirePricePurchaseReservedInstancesOfferingResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquirePricePurchaseReservedInstancesOffering
// 本接口(InquirePricePurchaseReservedInstancesOffering)用于创建预留实例询价。本接口仅允许针对购买限制范围内的预留实例配置进行询价。预留实例当前只针对国际站白名单用户开放。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
func (c *Client) InquirePricePurchaseReservedInstancesOffering(request *InquirePricePurchaseReservedInstancesOfferingRequest) (response *InquirePricePurchaseReservedInstancesOfferingResponse, err error) {
	if request == nil {
		request = NewInquirePricePurchaseReservedInstancesOfferingRequest()
	}

	response = NewInquirePricePurchaseReservedInstancesOfferingResponse()
	err = c.Send(request, response)
	return
}

// InquirePricePurchaseReservedInstancesOffering
// 本接口(InquirePricePurchaseReservedInstancesOffering)用于创建预留实例询价。本接口仅允许针对购买限制范围内的预留实例配置进行询价。预留实例当前只针对国际站白名单用户开放。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
func (c *Client) InquirePricePurchaseReservedInstancesOfferingWithContext(ctx context.Context, request *InquirePricePurchaseReservedInstancesOfferingRequest) (response *InquirePricePurchaseReservedInstancesOfferingResponse, err error) {
	if request == nil {
		request = NewInquirePricePurchaseReservedInstancesOfferingRequest()
	}
	request.SetContext(ctx)

	response = NewInquirePricePurchaseReservedInstancesOfferingResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceModifyInstancesChargeTypeRequest() (request *InquiryPriceModifyInstancesChargeTypeRequest) {
	request = &InquiryPriceModifyInstancesChargeTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquiryPriceModifyInstancesChargeType")

	return
}

func NewInquiryPriceModifyInstancesChargeTypeResponse() (response *InquiryPriceModifyInstancesChargeTypeResponse) {
	response = &InquiryPriceModifyInstancesChargeTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceModifyInstancesChargeType
// 本接口 (InquiryPriceModifyInstancesChargeType) 用于切换实例的计费模式询价。
//
// * 只支持从 `POSTPAID_BY_HOUR` 计费模式切换为`PREPAID`计费模式。
//
// * 关机不收费的实例、`BC1`和`BS1`机型族的实例、设置定时销毁的实例、竞价实例不支持该操作。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
func (c *Client) InquiryPriceModifyInstancesChargeType(request *InquiryPriceModifyInstancesChargeTypeRequest) (response *InquiryPriceModifyInstancesChargeTypeResponse, err error) {
	if request == nil {
		request = NewInquiryPriceModifyInstancesChargeTypeRequest()
	}

	response = NewInquiryPriceModifyInstancesChargeTypeResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceModifyInstancesChargeType
// 本接口 (InquiryPriceModifyInstancesChargeType) 用于切换实例的计费模式询价。
//
// * 只支持从 `POSTPAID_BY_HOUR` 计费模式切换为`PREPAID`计费模式。
//
// * 关机不收费的实例、`BC1`和`BS1`机型族的实例、设置定时销毁的实例、竞价实例不支持该操作。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
func (c *Client) InquiryPriceModifyInstancesChargeTypeWithContext(ctx context.Context, request *InquiryPriceModifyInstancesChargeTypeRequest) (response *InquiryPriceModifyInstancesChargeTypeResponse, err error) {
	if request == nil {
		request = NewInquiryPriceModifyInstancesChargeTypeRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceModifyInstancesChargeTypeResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceRenewInstancesRequest() (request *InquiryPriceRenewInstancesRequest) {
	request = &InquiryPriceRenewInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquiryPriceRenewInstances")

	return
}

func NewInquiryPriceRenewInstancesResponse() (response *InquiryPriceRenewInstancesResponse) {
	response = &InquiryPriceRenewInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceRenewInstances
// 本接口 (InquiryPriceRenewInstances) 用于续费包年包月实例询价。
//
// * 只支持查询包年包月实例的续费价格。
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INSTANCENOTSUPPORTEDMIXPRICINGMODEL = "InvalidParameterValue.InstanceNotSupportedMixPricingModel"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPERIOD = "InvalidPeriod"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
func (c *Client) InquiryPriceRenewInstances(request *InquiryPriceRenewInstancesRequest) (response *InquiryPriceRenewInstancesResponse, err error) {
	if request == nil {
		request = NewInquiryPriceRenewInstancesRequest()
	}

	response = NewInquiryPriceRenewInstancesResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceRenewInstances
// 本接口 (InquiryPriceRenewInstances) 用于续费包年包月实例询价。
//
// * 只支持查询包年包月实例的续费价格。
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INSTANCENOTSUPPORTEDMIXPRICINGMODEL = "InvalidParameterValue.InstanceNotSupportedMixPricingModel"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPERIOD = "InvalidPeriod"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
func (c *Client) InquiryPriceRenewInstancesWithContext(ctx context.Context, request *InquiryPriceRenewInstancesRequest) (response *InquiryPriceRenewInstancesResponse, err error) {
	if request == nil {
		request = NewInquiryPriceRenewInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceRenewInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceResetInstanceRequest() (request *InquiryPriceResetInstanceRequest) {
	request = &InquiryPriceResetInstanceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquiryPriceResetInstance")

	return
}

func NewInquiryPriceResetInstanceResponse() (response *InquiryPriceResetInstanceResponse) {
	response = &InquiryPriceResetInstanceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceResetInstance
// 本接口 (InquiryPriceResetInstance) 用于重装实例询价。
//
// * 如果指定了`ImageId`参数，则使用指定的镜像进行重装询价；否则按照当前实例使用的镜像进行重装询价。
//
// * 目前只支持[系统盘类型](https://cloud.tencent.com/document/api/213/15753#SystemDisk)是`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`类型的实例使用该接口实现`Linux`和`Windows`操作系统切换的重装询价。
//
// * 目前不支持境外地域的实例使用该接口实现`Linux`和`Windows`操作系统切换的重装询价。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEOSNAME = "InvalidParameterValue.InvalidImageOsName"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	UNSUPPORTEDOPERATION_RAWLOCALDISKINSREINSTALLTOQCOW2 = "UnsupportedOperation.RawLocalDiskInsReinstalltoQcow2"
func (c *Client) InquiryPriceResetInstance(request *InquiryPriceResetInstanceRequest) (response *InquiryPriceResetInstanceResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResetInstanceRequest()
	}

	response = NewInquiryPriceResetInstanceResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceResetInstance
// 本接口 (InquiryPriceResetInstance) 用于重装实例询价。
//
// * 如果指定了`ImageId`参数，则使用指定的镜像进行重装询价；否则按照当前实例使用的镜像进行重装询价。
//
// * 目前只支持[系统盘类型](https://cloud.tencent.com/document/api/213/15753#SystemDisk)是`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`类型的实例使用该接口实现`Linux`和`Windows`操作系统切换的重装询价。
//
// * 目前不支持境外地域的实例使用该接口实现`Linux`和`Windows`操作系统切换的重装询价。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEOSNAME = "InvalidParameterValue.InvalidImageOsName"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	UNSUPPORTEDOPERATION_RAWLOCALDISKINSREINSTALLTOQCOW2 = "UnsupportedOperation.RawLocalDiskInsReinstalltoQcow2"
func (c *Client) InquiryPriceResetInstanceWithContext(ctx context.Context, request *InquiryPriceResetInstanceRequest) (response *InquiryPriceResetInstanceResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResetInstanceRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceResetInstanceResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceResetInstancesInternetMaxBandwidthRequest() (request *InquiryPriceResetInstancesInternetMaxBandwidthRequest) {
	request = &InquiryPriceResetInstancesInternetMaxBandwidthRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquiryPriceResetInstancesInternetMaxBandwidth")

	return
}

func NewInquiryPriceResetInstancesInternetMaxBandwidthResponse() (response *InquiryPriceResetInstancesInternetMaxBandwidthResponse) {
	response = &InquiryPriceResetInstancesInternetMaxBandwidthResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceResetInstancesInternetMaxBandwidth
// 本接口 (InquiryPriceResetInstancesInternetMaxBandwidth) 用于调整实例公网带宽上限询价。
//
// * 不同机型带宽上限范围不一致，具体限制详见[公网带宽上限](https://cloud.tencent.com/document/product/213/12523)。
//
// * 对于`BANDWIDTH_PREPAID`计费方式的带宽，目前不支持调小带宽，且需要输入参数`StartTime`和`EndTime`，指定调整后的带宽的生效时间段。在这种场景下会涉及扣费，请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// * 对于 `TRAFFIC_POSTPAID_BY_HOUR`、 `BANDWIDTH_POSTPAID_BY_HOUR` 和 `BANDWIDTH_PACKAGE` 计费方式的带宽，使用该接口调整带宽上限是实时生效的，可以在带宽允许的范围内调大或者调小带宽，不支持输入参数 `StartTime` 和 `EndTime` 。
//
// * 接口不支持调整`BANDWIDTH_POSTPAID_BY_MONTH`计费方式的带宽。
//
// * 接口不支持批量调整 `BANDWIDTH_PREPAID` 和 `BANDWIDTH_POSTPAID_BY_HOUR` 计费方式的带宽。
//
// * 接口不支持批量调整混合计费方式的带宽。例如不支持同时调整`TRAFFIC_POSTPAID_BY_HOUR`和`BANDWIDTH_PACKAGE`计费方式的带宽。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_NOTFOUNDEIP = "FailedOperation.NotFoundEIP"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPERMISSION = "InvalidPermission"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) InquiryPriceResetInstancesInternetMaxBandwidth(request *InquiryPriceResetInstancesInternetMaxBandwidthRequest) (response *InquiryPriceResetInstancesInternetMaxBandwidthResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResetInstancesInternetMaxBandwidthRequest()
	}

	response = NewInquiryPriceResetInstancesInternetMaxBandwidthResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceResetInstancesInternetMaxBandwidth
// 本接口 (InquiryPriceResetInstancesInternetMaxBandwidth) 用于调整实例公网带宽上限询价。
//
// * 不同机型带宽上限范围不一致，具体限制详见[公网带宽上限](https://cloud.tencent.com/document/product/213/12523)。
//
// * 对于`BANDWIDTH_PREPAID`计费方式的带宽，目前不支持调小带宽，且需要输入参数`StartTime`和`EndTime`，指定调整后的带宽的生效时间段。在这种场景下会涉及扣费，请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// * 对于 `TRAFFIC_POSTPAID_BY_HOUR`、 `BANDWIDTH_POSTPAID_BY_HOUR` 和 `BANDWIDTH_PACKAGE` 计费方式的带宽，使用该接口调整带宽上限是实时生效的，可以在带宽允许的范围内调大或者调小带宽，不支持输入参数 `StartTime` 和 `EndTime` 。
//
// * 接口不支持调整`BANDWIDTH_POSTPAID_BY_MONTH`计费方式的带宽。
//
// * 接口不支持批量调整 `BANDWIDTH_PREPAID` 和 `BANDWIDTH_POSTPAID_BY_HOUR` 计费方式的带宽。
//
// * 接口不支持批量调整混合计费方式的带宽。例如不支持同时调整`TRAFFIC_POSTPAID_BY_HOUR`和`BANDWIDTH_PACKAGE`计费方式的带宽。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_NOTFOUNDEIP = "FailedOperation.NotFoundEIP"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPERMISSION = "InvalidPermission"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) InquiryPriceResetInstancesInternetMaxBandwidthWithContext(ctx context.Context, request *InquiryPriceResetInstancesInternetMaxBandwidthRequest) (response *InquiryPriceResetInstancesInternetMaxBandwidthResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResetInstancesInternetMaxBandwidthRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceResetInstancesInternetMaxBandwidthResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceResetInstancesTypeRequest() (request *InquiryPriceResetInstancesTypeRequest) {
	request = &InquiryPriceResetInstancesTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquiryPriceResetInstancesType")

	return
}

func NewInquiryPriceResetInstancesTypeResponse() (response *InquiryPriceResetInstancesTypeResponse) {
	response = &InquiryPriceResetInstancesTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceResetInstancesType
// 本接口 (InquiryPriceResetInstancesType) 用于调整实例的机型询价。
//
// * 目前只支持[系统盘类型](https://cloud.tencent.com/document/product/213/15753#SystemDisk)是`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`类型的实例使用该接口进行调整机型询价。
//
// * 目前不支持[CDH](https://cloud.tencent.com/document/product/416)实例使用该接口调整机型询价。
//
// * 对于包年包月实例，使用该接口会涉及扣费，请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYREFUNDPRICEFAILED = "FailedOperation.InquiryRefundPriceFailed"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPERMISSION = "InvalidPermission"
//	LIMITEXCEEDED_INSTANCETYPEBANDWIDTH = "LimitExceeded.InstanceTypeBandwidth"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	UNSUPPORTEDOPERATION_LOCALDATADISKCHANGEINSTANCEFAMILY = "UnsupportedOperation.LocalDataDiskChangeInstanceFamily"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILYTOSA3 = "UnsupportedOperation.UnsupportedChangeInstanceFamilyToSA3"
func (c *Client) InquiryPriceResetInstancesType(request *InquiryPriceResetInstancesTypeRequest) (response *InquiryPriceResetInstancesTypeResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResetInstancesTypeRequest()
	}

	response = NewInquiryPriceResetInstancesTypeResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceResetInstancesType
// 本接口 (InquiryPriceResetInstancesType) 用于调整实例的机型询价。
//
// * 目前只支持[系统盘类型](https://cloud.tencent.com/document/product/213/15753#SystemDisk)是`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`类型的实例使用该接口进行调整机型询价。
//
// * 目前不支持[CDH](https://cloud.tencent.com/document/product/416)实例使用该接口调整机型询价。
//
// * 对于包年包月实例，使用该接口会涉及扣费，请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYREFUNDPRICEFAILED = "FailedOperation.InquiryRefundPriceFailed"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPERMISSION = "InvalidPermission"
//	LIMITEXCEEDED_INSTANCETYPEBANDWIDTH = "LimitExceeded.InstanceTypeBandwidth"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	UNSUPPORTEDOPERATION_LOCALDATADISKCHANGEINSTANCEFAMILY = "UnsupportedOperation.LocalDataDiskChangeInstanceFamily"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILYTOSA3 = "UnsupportedOperation.UnsupportedChangeInstanceFamilyToSA3"
func (c *Client) InquiryPriceResetInstancesTypeWithContext(ctx context.Context, request *InquiryPriceResetInstancesTypeRequest) (response *InquiryPriceResetInstancesTypeResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResetInstancesTypeRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceResetInstancesTypeResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceResizeInstanceDisksRequest() (request *InquiryPriceResizeInstanceDisksRequest) {
	request = &InquiryPriceResizeInstanceDisksRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquiryPriceResizeInstanceDisks")

	return
}

func NewInquiryPriceResizeInstanceDisksResponse() (response *InquiryPriceResizeInstanceDisksResponse) {
	response = &InquiryPriceResizeInstanceDisksResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceResizeInstanceDisks
// 本接口 (InquiryPriceResizeInstanceDisks) 用于扩容实例的数据盘询价。
//
// * 目前只支持扩容非弹性数据盘（[`DescribeDisks`](https://cloud.tencent.com/document/api/362/16315)接口返回值中的`Portable`为`false`表示非弹性）询价，且[数据盘类型](https://cloud.tencent.com/document/product/213/15753#DataDisk)为：`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`。
//
// * 目前不支持[CDH](https://cloud.tencent.com/document/product/416)实例使用该接口扩容数据盘询价。* 仅支持包年包月实例随机器购买的数据盘。* 目前只支持扩容一块数据盘询价。
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_ATLEASTONE = "MissingParameter.AtLeastOne"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
func (c *Client) InquiryPriceResizeInstanceDisks(request *InquiryPriceResizeInstanceDisksRequest) (response *InquiryPriceResizeInstanceDisksResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResizeInstanceDisksRequest()
	}

	response = NewInquiryPriceResizeInstanceDisksResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceResizeInstanceDisks
// 本接口 (InquiryPriceResizeInstanceDisks) 用于扩容实例的数据盘询价。
//
// * 目前只支持扩容非弹性数据盘（[`DescribeDisks`](https://cloud.tencent.com/document/api/362/16315)接口返回值中的`Portable`为`false`表示非弹性）询价，且[数据盘类型](https://cloud.tencent.com/document/product/213/15753#DataDisk)为：`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`。
//
// * 目前不支持[CDH](https://cloud.tencent.com/document/product/416)实例使用该接口扩容数据盘询价。* 仅支持包年包月实例随机器购买的数据盘。* 目前只支持扩容一块数据盘询价。
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_ATLEASTONE = "MissingParameter.AtLeastOne"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
func (c *Client) InquiryPriceResizeInstanceDisksWithContext(ctx context.Context, request *InquiryPriceResizeInstanceDisksRequest) (response *InquiryPriceResizeInstanceDisksResponse, err error) {
	if request == nil {
		request = NewInquiryPriceResizeInstanceDisksRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceResizeInstanceDisksResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceRunInstancesRequest() (request *InquiryPriceRunInstancesRequest) {
	request = &InquiryPriceRunInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquiryPriceRunInstances")

	return
}

func NewInquiryPriceRunInstancesResponse() (response *InquiryPriceRunInstancesResponse) {
	response = &InquiryPriceRunInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceRunInstances
// 本接口(InquiryPriceRunInstances)用于创建实例询价。本接口仅允许针对购买限制范围内的实例配置进行询价, 详见：[创建实例](https://cloud.tencent.com/document/api/213/15730)。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_SNAPSHOTSIZELARGERTHANDATASIZE = "FailedOperation.SnapshotSizeLargerThanDataSize"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	FAILEDOPERATION_TATAGENTNOTSUPPORT = "FailedOperation.TatAgentNotSupport"
//	INSTANCESQUOTALIMITEXCEEDED = "InstancesQuotaLimitExceeded"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDNOTFOUND = "InvalidParameterValue.BandwidthPackageIdNotFound"
//	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"
//	INVALIDPARAMETERVALUE_INSTANCETYPEREQUIREDHPCCLUSTER = "InvalidParameterValue.InstanceTypeRequiredHpcCluster"
//	INVALIDPARAMETERVALUE_INSUFFICIENTPRICE = "InvalidParameterValue.InsufficientPrice"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEFORMAT = "InvalidParameterValue.InvalidImageFormat"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEOSNAME = "InvalidParameterValue.InvalidImageOsName"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"
//	INVALIDPASSWORD = "InvalidPassword"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_DISASTERRECOVERGROUP = "LimitExceeded.DisasterRecoverGroup"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"
//	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDINTERNATIONALUSER = "UnsupportedOperation.UnsupportedInternationalUser"
func (c *Client) InquiryPriceRunInstances(request *InquiryPriceRunInstancesRequest) (response *InquiryPriceRunInstancesResponse, err error) {
	if request == nil {
		request = NewInquiryPriceRunInstancesRequest()
	}

	response = NewInquiryPriceRunInstancesResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceRunInstances
// 本接口(InquiryPriceRunInstances)用于创建实例询价。本接口仅允许针对购买限制范围内的实例配置进行询价, 详见：[创建实例](https://cloud.tencent.com/document/api/213/15730)。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_SNAPSHOTSIZELARGERTHANDATASIZE = "FailedOperation.SnapshotSizeLargerThanDataSize"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	FAILEDOPERATION_TATAGENTNOTSUPPORT = "FailedOperation.TatAgentNotSupport"
//	INSTANCESQUOTALIMITEXCEEDED = "InstancesQuotaLimitExceeded"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDNOTFOUND = "InvalidParameterValue.BandwidthPackageIdNotFound"
//	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"
//	INVALIDPARAMETERVALUE_INSTANCETYPEREQUIREDHPCCLUSTER = "InvalidParameterValue.InstanceTypeRequiredHpcCluster"
//	INVALIDPARAMETERVALUE_INSUFFICIENTPRICE = "InvalidParameterValue.InsufficientPrice"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEFORMAT = "InvalidParameterValue.InvalidImageFormat"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEOSNAME = "InvalidParameterValue.InvalidImageOsName"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"
//	INVALIDPASSWORD = "InvalidPassword"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_DISASTERRECOVERGROUP = "LimitExceeded.DisasterRecoverGroup"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"
//	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDINTERNATIONALUSER = "UnsupportedOperation.UnsupportedInternationalUser"
func (c *Client) InquiryPriceRunInstancesWithContext(ctx context.Context, request *InquiryPriceRunInstancesRequest) (response *InquiryPriceRunInstancesResponse, err error) {
	if request == nil {
		request = NewInquiryPriceRunInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceRunInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewInquiryPriceTerminateInstancesRequest() (request *InquiryPriceTerminateInstancesRequest) {
	request = &InquiryPriceTerminateInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "InquiryPriceTerminateInstances")

	return
}

func NewInquiryPriceTerminateInstancesResponse() (response *InquiryPriceTerminateInstancesResponse) {
	response = &InquiryPriceTerminateInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// InquiryPriceTerminateInstances
// 本接口 (InquiryPriceTerminateInstances) 用于退还实例询价。
//
// * 查询退还实例可以返还的费用。
//
// * 支持批量操作，每次请求批量实例的上限为100。如果批量实例存在不允许操作的实例，操作会以特定[错误码](#4.-.E9.94.99.E8.AF.AF.E7.A0.81)返回。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYREFUNDPRICEFAILED = "FailedOperation.InquiryRefundPriceFailed"
//	FAILEDOPERATION_UNRETURNABLE = "FailedOperation.Unreturnable"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCENOTSUPPORTEDPREPAIDINSTANCE = "InvalidInstanceNotSupportedPrepaidInstance"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION_INSTANCEMIXEDPRICINGMODEL = "UnsupportedOperation.InstanceMixedPricingModel"
//	UNSUPPORTEDOPERATION_REGION = "UnsupportedOperation.Region"
func (c *Client) InquiryPriceTerminateInstances(request *InquiryPriceTerminateInstancesRequest) (response *InquiryPriceTerminateInstancesResponse, err error) {
	if request == nil {
		request = NewInquiryPriceTerminateInstancesRequest()
	}

	response = NewInquiryPriceTerminateInstancesResponse()
	err = c.Send(request, response)
	return
}

// InquiryPriceTerminateInstances
// 本接口 (InquiryPriceTerminateInstances) 用于退还实例询价。
//
// * 查询退还实例可以返还的费用。
//
// * 支持批量操作，每次请求批量实例的上限为100。如果批量实例存在不允许操作的实例，操作会以特定[错误码](#4.-.E9.94.99.E8.AF.AF.E7.A0.81)返回。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INQUIRYREFUNDPRICEFAILED = "FailedOperation.InquiryRefundPriceFailed"
//	FAILEDOPERATION_UNRETURNABLE = "FailedOperation.Unreturnable"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCENOTSUPPORTEDPREPAIDINSTANCE = "InvalidInstanceNotSupportedPrepaidInstance"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION_INSTANCEMIXEDPRICINGMODEL = "UnsupportedOperation.InstanceMixedPricingModel"
//	UNSUPPORTEDOPERATION_REGION = "UnsupportedOperation.Region"
func (c *Client) InquiryPriceTerminateInstancesWithContext(ctx context.Context, request *InquiryPriceTerminateInstancesRequest) (response *InquiryPriceTerminateInstancesResponse, err error) {
	if request == nil {
		request = NewInquiryPriceTerminateInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewInquiryPriceTerminateInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyDisasterRecoverGroupAttributeRequest() (request *ModifyDisasterRecoverGroupAttributeRequest) {
	request = &ModifyDisasterRecoverGroupAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyDisasterRecoverGroupAttribute")

	return
}

func NewModifyDisasterRecoverGroupAttributeResponse() (response *ModifyDisasterRecoverGroupAttributeResponse) {
	response = &ModifyDisasterRecoverGroupAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyDisasterRecoverGroupAttribute
// 本接口 (ModifyDisasterRecoverGroupAttribute)用于修改[分散置放群组](https://cloud.tencent.com/document/product/213/15486)属性。
//
// 可能返回的错误码:
//
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCENOTFOUND_INVALIDPLACEMENTSET = "ResourceNotFound.InvalidPlacementSet"
func (c *Client) ModifyDisasterRecoverGroupAttribute(request *ModifyDisasterRecoverGroupAttributeRequest) (response *ModifyDisasterRecoverGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifyDisasterRecoverGroupAttributeRequest()
	}

	response = NewModifyDisasterRecoverGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyDisasterRecoverGroupAttribute
// 本接口 (ModifyDisasterRecoverGroupAttribute)用于修改[分散置放群组](https://cloud.tencent.com/document/product/213/15486)属性。
//
// 可能返回的错误码:
//
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCENOTFOUND_INVALIDPLACEMENTSET = "ResourceNotFound.InvalidPlacementSet"
func (c *Client) ModifyDisasterRecoverGroupAttributeWithContext(ctx context.Context, request *ModifyDisasterRecoverGroupAttributeRequest) (response *ModifyDisasterRecoverGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifyDisasterRecoverGroupAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyDisasterRecoverGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyHostsAttributeRequest() (request *ModifyHostsAttributeRequest) {
	request = &ModifyHostsAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyHostsAttribute")

	return
}

func NewModifyHostsAttributeResponse() (response *ModifyHostsAttributeResponse) {
	response = &ModifyHostsAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyHostsAttribute
// 本接口（ModifyHostsAttribute）用于修改CDH实例的属性，如实例名称和续费标记等。参数HostName和RenewFlag必须设置其中一个，但不能同时设置。
//
// 可能返回的错误码:
//
//	INVALIDHOST_NOTSUPPORTED = "InvalidHost.NotSupported"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
func (c *Client) ModifyHostsAttribute(request *ModifyHostsAttributeRequest) (response *ModifyHostsAttributeResponse, err error) {
	if request == nil {
		request = NewModifyHostsAttributeRequest()
	}

	response = NewModifyHostsAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyHostsAttribute
// 本接口（ModifyHostsAttribute）用于修改CDH实例的属性，如实例名称和续费标记等。参数HostName和RenewFlag必须设置其中一个，但不能同时设置。
//
// 可能返回的错误码:
//
//	INVALIDHOST_NOTSUPPORTED = "InvalidHost.NotSupported"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
func (c *Client) ModifyHostsAttributeWithContext(ctx context.Context, request *ModifyHostsAttributeRequest) (response *ModifyHostsAttributeResponse, err error) {
	if request == nil {
		request = NewModifyHostsAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyHostsAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyImageAttributeRequest() (request *ModifyImageAttributeRequest) {
	request = &ModifyImageAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyImageAttribute")

	return
}

func NewModifyImageAttributeResponse() (response *ModifyImageAttributeResponse) {
	response = &ModifyImageAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyImageAttribute
// 本接口（ModifyImageAttribute）用于修改镜像属性。
//
// * 已分享的镜像无法修改属性。
//
// 可能返回的错误码:
//
//	INVALIDIMAGEID_INCORRECTSTATE = "InvalidImageId.IncorrectState"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"
//	INVALIDPARAMETER_VALUETOOLARGE = "InvalidParameter.ValueTooLarge"
//	INVALIDPARAMETERVALUE_TOOLARGE = "InvalidParameterValue.TooLarge"
func (c *Client) ModifyImageAttribute(request *ModifyImageAttributeRequest) (response *ModifyImageAttributeResponse, err error) {
	if request == nil {
		request = NewModifyImageAttributeRequest()
	}

	response = NewModifyImageAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyImageAttribute
// 本接口（ModifyImageAttribute）用于修改镜像属性。
//
// * 已分享的镜像无法修改属性。
//
// 可能返回的错误码:
//
//	INVALIDIMAGEID_INCORRECTSTATE = "InvalidImageId.IncorrectState"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"
//	INVALIDPARAMETER_VALUETOOLARGE = "InvalidParameter.ValueTooLarge"
//	INVALIDPARAMETERVALUE_TOOLARGE = "InvalidParameterValue.TooLarge"
func (c *Client) ModifyImageAttributeWithContext(ctx context.Context, request *ModifyImageAttributeRequest) (response *ModifyImageAttributeResponse, err error) {
	if request == nil {
		request = NewModifyImageAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyImageAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyImageSharePermissionRequest() (request *ModifyImageSharePermissionRequest) {
	request = &ModifyImageSharePermissionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyImageSharePermission")

	return
}

func NewModifyImageSharePermissionResponse() (response *ModifyImageSharePermissionResponse) {
	response = &ModifyImageSharePermissionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyImageSharePermission
// 本接口（ModifyImageSharePermission）用于修改镜像分享信息。
//
// * 分享镜像后，被分享账户可以通过该镜像创建实例。
//
// * 每个自定义镜像最多可共享给50个账户。
//
// * 分享镜像无法更改名称，描述，仅可用于创建实例。
//
// * 只支持分享到对方账户相同地域。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_ACCOUNTALREADYEXISTS = "FailedOperation.AccountAlreadyExists"
//	FAILEDOPERATION_ACCOUNTISYOURSELF = "FailedOperation.AccountIsYourSelf"
//	FAILEDOPERATION_NOTMASTERACCOUNT = "FailedOperation.NotMasterAccount"
//	FAILEDOPERATION_QIMAGESHAREFAILED = "FailedOperation.QImageShareFailed"
//	FAILEDOPERATION_RIMAGESHAREFAILED = "FailedOperation.RImageShareFailed"
//	INVALIDACCOUNTID_NOTFOUND = "InvalidAccountId.NotFound"
//	INVALIDACCOUNTIS_YOURSELF = "InvalidAccountIs.YourSelf"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	OVERQUOTA = "OverQuota"
//	UNAUTHORIZEDOPERATION_IMAGENOTBELONGTOACCOUNT = "UnauthorizedOperation.ImageNotBelongToAccount"
func (c *Client) ModifyImageSharePermission(request *ModifyImageSharePermissionRequest) (response *ModifyImageSharePermissionResponse, err error) {
	if request == nil {
		request = NewModifyImageSharePermissionRequest()
	}

	response = NewModifyImageSharePermissionResponse()
	err = c.Send(request, response)
	return
}

// ModifyImageSharePermission
// 本接口（ModifyImageSharePermission）用于修改镜像分享信息。
//
// * 分享镜像后，被分享账户可以通过该镜像创建实例。
//
// * 每个自定义镜像最多可共享给50个账户。
//
// * 分享镜像无法更改名称，描述，仅可用于创建实例。
//
// * 只支持分享到对方账户相同地域。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_ACCOUNTALREADYEXISTS = "FailedOperation.AccountAlreadyExists"
//	FAILEDOPERATION_ACCOUNTISYOURSELF = "FailedOperation.AccountIsYourSelf"
//	FAILEDOPERATION_NOTMASTERACCOUNT = "FailedOperation.NotMasterAccount"
//	FAILEDOPERATION_QIMAGESHAREFAILED = "FailedOperation.QImageShareFailed"
//	FAILEDOPERATION_RIMAGESHAREFAILED = "FailedOperation.RImageShareFailed"
//	INVALIDACCOUNTID_NOTFOUND = "InvalidAccountId.NotFound"
//	INVALIDACCOUNTIS_YOURSELF = "InvalidAccountIs.YourSelf"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	OVERQUOTA = "OverQuota"
//	UNAUTHORIZEDOPERATION_IMAGENOTBELONGTOACCOUNT = "UnauthorizedOperation.ImageNotBelongToAccount"
func (c *Client) ModifyImageSharePermissionWithContext(ctx context.Context, request *ModifyImageSharePermissionRequest) (response *ModifyImageSharePermissionResponse, err error) {
	if request == nil {
		request = NewModifyImageSharePermissionRequest()
	}
	request.SetContext(ctx)

	response = NewModifyImageSharePermissionResponse()
	err = c.Send(request, response)
	return
}

func NewModifyInstanceDiskTypeRequest() (request *ModifyInstanceDiskTypeRequest) {
	request = &ModifyInstanceDiskTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyInstanceDiskType")

	return
}

func NewModifyInstanceDiskTypeResponse() (response *ModifyInstanceDiskTypeResponse) {
	response = &ModifyInstanceDiskTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyInstanceDiskType
// 本接口 (ModifyInstanceDiskType) 用于修改实例硬盘介质类型。
//
// * 只支持实例的本地系统盘、本地数据盘转化成指定云硬盘介质。
//
// * 只支持实例在关机状态下转换成指定云硬盘介质。
//
// * 不支持竞价实例类型。
//
// * 修改前请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/378/4397)接口查询账户余额。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_INVALIDCLOUDDISKSOLDOUT = "InvalidParameter.InvalidCloudDiskSoldOut"
//	INVALIDPARAMETER_INVALIDINSTANCENOTSUPPORTED = "InvalidParameter.InvalidInstanceNotSupported"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LOCALDISKSIZERANGE = "InvalidParameterValue.LocalDiskSizeRange"
//	INVALIDPERMISSION = "InvalidPermission"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
func (c *Client) ModifyInstanceDiskType(request *ModifyInstanceDiskTypeRequest) (response *ModifyInstanceDiskTypeResponse, err error) {
	if request == nil {
		request = NewModifyInstanceDiskTypeRequest()
	}

	response = NewModifyInstanceDiskTypeResponse()
	err = c.Send(request, response)
	return
}

// ModifyInstanceDiskType
// 本接口 (ModifyInstanceDiskType) 用于修改实例硬盘介质类型。
//
// * 只支持实例的本地系统盘、本地数据盘转化成指定云硬盘介质。
//
// * 只支持实例在关机状态下转换成指定云硬盘介质。
//
// * 不支持竞价实例类型。
//
// * 修改前请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/378/4397)接口查询账户余额。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_INVALIDCLOUDDISKSOLDOUT = "InvalidParameter.InvalidCloudDiskSoldOut"
//	INVALIDPARAMETER_INVALIDINSTANCENOTSUPPORTED = "InvalidParameter.InvalidInstanceNotSupported"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LOCALDISKSIZERANGE = "InvalidParameterValue.LocalDiskSizeRange"
//	INVALIDPERMISSION = "InvalidPermission"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
func (c *Client) ModifyInstanceDiskTypeWithContext(ctx context.Context, request *ModifyInstanceDiskTypeRequest) (response *ModifyInstanceDiskTypeResponse, err error) {
	if request == nil {
		request = NewModifyInstanceDiskTypeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyInstanceDiskTypeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyInstancesAttributeRequest() (request *ModifyInstancesAttributeRequest) {
	request = &ModifyInstancesAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyInstancesAttribute")

	return
}

func NewModifyInstancesAttributeResponse() (response *ModifyInstancesAttributeResponse) {
	response = &ModifyInstancesAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyInstancesAttribute
// 本接口 (ModifyInstancesAttribute) 用于修改实例的属性（目前只支持修改实例的名称和关联的安全组）。
//
// * 每次请求必须指定实例的一种属性用于修改。
//
// * “实例名称”仅为方便用户自己管理之用，腾讯云并不以此名称作为在线支持或是进行实例管理操作的依据。
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 修改关联安全组时，子机原来关联的安全组会被解绑。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDPARAMETER_HOSTNAMEILLEGAL = "InvalidParameter.HostNameIllegal"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	LIMITEXCEEDED_ASSOCIATEUSGLIMITEXCEEDED = "LimitExceeded.AssociateUSGLimitExceeded"
//	LIMITEXCEEDED_CVMSVIFSPERSECGROUPLIMITEXCEEDED = "LimitExceeded.CvmsVifsPerSecGroupLimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ModifyInstancesAttribute(request *ModifyInstancesAttributeRequest) (response *ModifyInstancesAttributeResponse, err error) {
	if request == nil {
		request = NewModifyInstancesAttributeRequest()
	}

	response = NewModifyInstancesAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyInstancesAttribute
// 本接口 (ModifyInstancesAttribute) 用于修改实例的属性（目前只支持修改实例的名称和关联的安全组）。
//
// * 每次请求必须指定实例的一种属性用于修改。
//
// * “实例名称”仅为方便用户自己管理之用，腾讯云并不以此名称作为在线支持或是进行实例管理操作的依据。
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 修改关联安全组时，子机原来关联的安全组会被解绑。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDPARAMETER_HOSTNAMEILLEGAL = "InvalidParameter.HostNameIllegal"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	LIMITEXCEEDED_ASSOCIATEUSGLIMITEXCEEDED = "LimitExceeded.AssociateUSGLimitExceeded"
//	LIMITEXCEEDED_CVMSVIFSPERSECGROUPLIMITEXCEEDED = "LimitExceeded.CvmsVifsPerSecGroupLimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ModifyInstancesAttributeWithContext(ctx context.Context, request *ModifyInstancesAttributeRequest) (response *ModifyInstancesAttributeResponse, err error) {
	if request == nil {
		request = NewModifyInstancesAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyInstancesAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyInstancesChargeTypeRequest() (request *ModifyInstancesChargeTypeRequest) {
	request = &ModifyInstancesChargeTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyInstancesChargeType")

	return
}

func NewModifyInstancesChargeTypeResponse() (response *ModifyInstancesChargeTypeResponse) {
	response = &ModifyInstancesChargeTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyInstancesChargeType
// 本接口 (ModifyInstancesChargeType) 用于切换实例的计费模式。
//
// * 只支持从 `POSTPAID_BY_HOUR` 计费模式切换为`PREPAID`计费模式。
//
// * 关机不收费的实例、`BC1`和`BS1`机型族的实例、设置定时销毁的实例不支持该操作。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
func (c *Client) ModifyInstancesChargeType(request *ModifyInstancesChargeTypeRequest) (response *ModifyInstancesChargeTypeResponse, err error) {
	if request == nil {
		request = NewModifyInstancesChargeTypeRequest()
	}

	response = NewModifyInstancesChargeTypeResponse()
	err = c.Send(request, response)
	return
}

// ModifyInstancesChargeType
// 本接口 (ModifyInstancesChargeType) 用于切换实例的计费模式。
//
// * 只支持从 `POSTPAID_BY_HOUR` 计费模式切换为`PREPAID`计费模式。
//
// * 关机不收费的实例、`BC1`和`BS1`机型族的实例、设置定时销毁的实例不支持该操作。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
func (c *Client) ModifyInstancesChargeTypeWithContext(ctx context.Context, request *ModifyInstancesChargeTypeRequest) (response *ModifyInstancesChargeTypeResponse, err error) {
	if request == nil {
		request = NewModifyInstancesChargeTypeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyInstancesChargeTypeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyInstancesProjectRequest() (request *ModifyInstancesProjectRequest) {
	request = &ModifyInstancesProjectRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyInstancesProject")

	return
}

func NewModifyInstancesProjectResponse() (response *ModifyInstancesProjectResponse) {
	response = &ModifyInstancesProjectResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyInstancesProject
// 本接口 (ModifyInstancesProject) 用于修改实例所属项目。
//
// * 项目为一个虚拟概念，用户可以在一个账户下面建立多个项目，每个项目中管理不同的资源；将多个不同实例分属到不同项目中，后续使用 [`DescribeInstances`](https://cloud.tencent.com/document/api/213/15728)接口查询实例，项目ID可用于过滤结果。
//
// * 绑定负载均衡的实例不支持修改实例所属项目，请先使用[`DeregisterInstancesFromLoadBalancer`](https://cloud.tencent.com/document/api/214/1258)接口解绑负载均衡。
//
// [^_^]: # ( 修改实例所属项目会自动解关联实例原来关联的安全组，修改完成后可使用[`ModifyInstancesAttribute`](https://cloud.tencent.com/document/api/213/15739)接口关联安全组。)
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
func (c *Client) ModifyInstancesProject(request *ModifyInstancesProjectRequest) (response *ModifyInstancesProjectResponse, err error) {
	if request == nil {
		request = NewModifyInstancesProjectRequest()
	}

	response = NewModifyInstancesProjectResponse()
	err = c.Send(request, response)
	return
}

// ModifyInstancesProject
// 本接口 (ModifyInstancesProject) 用于修改实例所属项目。
//
// * 项目为一个虚拟概念，用户可以在一个账户下面建立多个项目，每个项目中管理不同的资源；将多个不同实例分属到不同项目中，后续使用 [`DescribeInstances`](https://cloud.tencent.com/document/api/213/15728)接口查询实例，项目ID可用于过滤结果。
//
// * 绑定负载均衡的实例不支持修改实例所属项目，请先使用[`DeregisterInstancesFromLoadBalancer`](https://cloud.tencent.com/document/api/214/1258)接口解绑负载均衡。
//
// [^_^]: # ( 修改实例所属项目会自动解关联实例原来关联的安全组，修改完成后可使用[`ModifyInstancesAttribute`](https://cloud.tencent.com/document/api/213/15739)接口关联安全组。)
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
func (c *Client) ModifyInstancesProjectWithContext(ctx context.Context, request *ModifyInstancesProjectRequest) (response *ModifyInstancesProjectResponse, err error) {
	if request == nil {
		request = NewModifyInstancesProjectRequest()
	}
	request.SetContext(ctx)

	response = NewModifyInstancesProjectResponse()
	err = c.Send(request, response)
	return
}

func NewModifyInstancesRenewFlagRequest() (request *ModifyInstancesRenewFlagRequest) {
	request = &ModifyInstancesRenewFlagRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyInstancesRenewFlag")

	return
}

func NewModifyInstancesRenewFlagResponse() (response *ModifyInstancesRenewFlagResponse) {
	response = &ModifyInstancesRenewFlagResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyInstancesRenewFlag
// 本接口 (ModifyInstancesRenewFlag) 用于修改包年包月实例续费标识。
//
// * 实例被标识为自动续费后，每次在实例到期时，会自动续费一个月。
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INVALIDINSTANCEAPPLICATIONROLEEMR = "FailedOperation.InvalidInstanceApplicationRoleEmr"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_UNDERWRITINGINSTANCETYPEONLYSUPPORTAUTORENEW = "UnsupportedOperation.UnderwritingInstanceTypeOnlySupportAutoRenew"
func (c *Client) ModifyInstancesRenewFlag(request *ModifyInstancesRenewFlagRequest) (response *ModifyInstancesRenewFlagResponse, err error) {
	if request == nil {
		request = NewModifyInstancesRenewFlagRequest()
	}

	response = NewModifyInstancesRenewFlagResponse()
	err = c.Send(request, response)
	return
}

// ModifyInstancesRenewFlag
// 本接口 (ModifyInstancesRenewFlag) 用于修改包年包月实例续费标识。
//
// * 实例被标识为自动续费后，每次在实例到期时，会自动续费一个月。
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INVALIDINSTANCEAPPLICATIONROLEEMR = "FailedOperation.InvalidInstanceApplicationRoleEmr"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_UNDERWRITINGINSTANCETYPEONLYSUPPORTAUTORENEW = "UnsupportedOperation.UnderwritingInstanceTypeOnlySupportAutoRenew"
func (c *Client) ModifyInstancesRenewFlagWithContext(ctx context.Context, request *ModifyInstancesRenewFlagRequest) (response *ModifyInstancesRenewFlagResponse, err error) {
	if request == nil {
		request = NewModifyInstancesRenewFlagRequest()
	}
	request.SetContext(ctx)

	response = NewModifyInstancesRenewFlagResponse()
	err = c.Send(request, response)
	return
}

func NewModifyInstancesVpcAttributeRequest() (request *ModifyInstancesVpcAttributeRequest) {
	request = &ModifyInstancesVpcAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyInstancesVpcAttribute")

	return
}

func NewModifyInstancesVpcAttributeResponse() (response *ModifyInstancesVpcAttributeResponse) {
	response = &ModifyInstancesVpcAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyInstancesVpcAttribute
// 本接口(ModifyInstancesVpcAttribute)用于修改实例vpc属性，如私有网络ip。
//
// * 此操作默认会关闭实例，完成后再启动。
//
// * 当指定私有网络ID和子网ID（子网必须在实例所在的可用区）与指定实例所在私有网络不一致时，会将实例迁移至指定的私有网络的子网下。执行此操作前请确保指定的实例上没有绑定[弹性网卡](https://cloud.tencent.com/document/product/576)和[负载均衡](https://cloud.tencent.com/document/product/214)。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	ENINOTALLOWEDCHANGESUBNET = "EniNotAllowedChangeSubnet"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCESTATE = "InvalidInstanceState"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_IPADDRESSMALFORMED = "InvalidParameterValue.IPAddressMalformed"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_IPV6NOTSUPPORTVPCMIGRATE = "UnsupportedOperation.IPv6NotSupportVpcMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEBANNING = "UnsupportedOperation.InstanceStateBanning"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_MODIFYVPCWITHCLB = "UnsupportedOperation.ModifyVPCWithCLB"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
//	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"
//	VPCIPISUSED = "VpcIpIsUsed"
func (c *Client) ModifyInstancesVpcAttribute(request *ModifyInstancesVpcAttributeRequest) (response *ModifyInstancesVpcAttributeResponse, err error) {
	if request == nil {
		request = NewModifyInstancesVpcAttributeRequest()
	}

	response = NewModifyInstancesVpcAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyInstancesVpcAttribute
// 本接口(ModifyInstancesVpcAttribute)用于修改实例vpc属性，如私有网络ip。
//
// * 此操作默认会关闭实例，完成后再启动。
//
// * 当指定私有网络ID和子网ID（子网必须在实例所在的可用区）与指定实例所在私有网络不一致时，会将实例迁移至指定的私有网络的子网下。执行此操作前请确保指定的实例上没有绑定[弹性网卡](https://cloud.tencent.com/document/product/576)和[负载均衡](https://cloud.tencent.com/document/product/214)。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	ENINOTALLOWEDCHANGESUBNET = "EniNotAllowedChangeSubnet"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCESTATE = "InvalidInstanceState"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_IPADDRESSMALFORMED = "InvalidParameterValue.IPAddressMalformed"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_IPV6NOTSUPPORTVPCMIGRATE = "UnsupportedOperation.IPv6NotSupportVpcMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEBANNING = "UnsupportedOperation.InstanceStateBanning"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_MODIFYVPCWITHCLB = "UnsupportedOperation.ModifyVPCWithCLB"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
//	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"
//	VPCIPISUSED = "VpcIpIsUsed"
func (c *Client) ModifyInstancesVpcAttributeWithContext(ctx context.Context, request *ModifyInstancesVpcAttributeRequest) (response *ModifyInstancesVpcAttributeResponse, err error) {
	if request == nil {
		request = NewModifyInstancesVpcAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyInstancesVpcAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyKeyPairAttributeRequest() (request *ModifyKeyPairAttributeRequest) {
	request = &ModifyKeyPairAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyKeyPairAttribute")

	return
}

func NewModifyKeyPairAttributeResponse() (response *ModifyKeyPairAttributeResponse) {
	response = &ModifyKeyPairAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyKeyPairAttribute
// 本接口 (ModifyKeyPairAttribute) 用于修改密钥对属性。
//
// * 修改密钥对ID所指定的密钥对的名称和描述信息。
//
// * 密钥对名称不能和已经存在的密钥对的名称重复。
//
// * 密钥对ID是密钥对的唯一标识，不可修改。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"
//	INVALIDKEYPAIRNAME_DUPLICATE = "InvalidKeyPairName.Duplicate"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) ModifyKeyPairAttribute(request *ModifyKeyPairAttributeRequest) (response *ModifyKeyPairAttributeResponse, err error) {
	if request == nil {
		request = NewModifyKeyPairAttributeRequest()
	}

	response = NewModifyKeyPairAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyKeyPairAttribute
// 本接口 (ModifyKeyPairAttribute) 用于修改密钥对属性。
//
// * 修改密钥对ID所指定的密钥对的名称和描述信息。
//
// * 密钥对名称不能和已经存在的密钥对的名称重复。
//
// * 密钥对ID是密钥对的唯一标识，不可修改。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"
//	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"
//	INVALIDKEYPAIRNAME_DUPLICATE = "InvalidKeyPairName.Duplicate"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	MISSINGPARAMETER = "MissingParameter"
func (c *Client) ModifyKeyPairAttributeWithContext(ctx context.Context, request *ModifyKeyPairAttributeRequest) (response *ModifyKeyPairAttributeResponse, err error) {
	if request == nil {
		request = NewModifyKeyPairAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyKeyPairAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyLaunchTemplateDefaultVersionRequest() (request *ModifyLaunchTemplateDefaultVersionRequest) {
	request = &ModifyLaunchTemplateDefaultVersionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ModifyLaunchTemplateDefaultVersion")

	return
}

func NewModifyLaunchTemplateDefaultVersionResponse() (response *ModifyLaunchTemplateDefaultVersionResponse) {
	response = &ModifyLaunchTemplateDefaultVersionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyLaunchTemplateDefaultVersion
// 本接口（ModifyLaunchTemplateDefaultVersion）用于修改实例启动模板默认版本。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERSETALREADY = "InvalidParameterValue.LaunchTemplateIdVerSetAlready"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	MISSINGPARAMETER = "MissingParameter"
//	UNKNOWNPARAMETER = "UnknownParameter"
func (c *Client) ModifyLaunchTemplateDefaultVersion(request *ModifyLaunchTemplateDefaultVersionRequest) (response *ModifyLaunchTemplateDefaultVersionResponse, err error) {
	if request == nil {
		request = NewModifyLaunchTemplateDefaultVersionRequest()
	}

	response = NewModifyLaunchTemplateDefaultVersionResponse()
	err = c.Send(request, response)
	return
}

// ModifyLaunchTemplateDefaultVersion
// 本接口（ModifyLaunchTemplateDefaultVersion）用于修改实例启动模板默认版本。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERSETALREADY = "InvalidParameterValue.LaunchTemplateIdVerSetAlready"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	MISSINGPARAMETER = "MissingParameter"
//	UNKNOWNPARAMETER = "UnknownParameter"
func (c *Client) ModifyLaunchTemplateDefaultVersionWithContext(ctx context.Context, request *ModifyLaunchTemplateDefaultVersionRequest) (response *ModifyLaunchTemplateDefaultVersionResponse, err error) {
	if request == nil {
		request = NewModifyLaunchTemplateDefaultVersionRequest()
	}
	request.SetContext(ctx)

	response = NewModifyLaunchTemplateDefaultVersionResponse()
	err = c.Send(request, response)
	return
}

func NewProgramFpgaImageRequest() (request *ProgramFpgaImageRequest) {
	request = &ProgramFpgaImageRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ProgramFpgaImage")

	return
}

func NewProgramFpgaImageResponse() (response *ProgramFpgaImageResponse) {
	response = &ProgramFpgaImageResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ProgramFpgaImage
// 本接口(ProgramFpgaImage)用于在线烧录由客户提供的FPGA镜像文件到指定实例的指定FPGA卡上。
//
// * 只支持对单个实例发起在线烧录FPGA镜像的操作。
//
// * 支持对单个实例的多块FPGA卡同时烧录FPGA镜像，DBDFs参数为空时，默认对指定实例的所有FPGA卡进行烧录。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INCORRECTFORMAT = "InvalidParameterValue.IncorrectFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNSUPPORTEDOPERATION_NOTFPGAINSTANCE = "UnsupportedOperation.NotFpgaInstance"
func (c *Client) ProgramFpgaImage(request *ProgramFpgaImageRequest) (response *ProgramFpgaImageResponse, err error) {
	if request == nil {
		request = NewProgramFpgaImageRequest()
	}

	response = NewProgramFpgaImageResponse()
	err = c.Send(request, response)
	return
}

// ProgramFpgaImage
// 本接口(ProgramFpgaImage)用于在线烧录由客户提供的FPGA镜像文件到指定实例的指定FPGA卡上。
//
// * 只支持对单个实例发起在线烧录FPGA镜像的操作。
//
// * 支持对单个实例的多块FPGA卡同时烧录FPGA镜像，DBDFs参数为空时，默认对指定实例的所有FPGA卡进行烧录。
//
// 可能返回的错误码:
//
//	INVALIDPARAMETERVALUE_INCORRECTFORMAT = "InvalidParameterValue.IncorrectFormat"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	UNSUPPORTEDOPERATION_NOTFPGAINSTANCE = "UnsupportedOperation.NotFpgaInstance"
func (c *Client) ProgramFpgaImageWithContext(ctx context.Context, request *ProgramFpgaImageRequest) (response *ProgramFpgaImageResponse, err error) {
	if request == nil {
		request = NewProgramFpgaImageRequest()
	}
	request.SetContext(ctx)

	response = NewProgramFpgaImageResponse()
	err = c.Send(request, response)
	return
}

func NewPurchaseReservedInstancesOfferingRequest() (request *PurchaseReservedInstancesOfferingRequest) {
	request = &PurchaseReservedInstancesOfferingRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "PurchaseReservedInstancesOffering")

	return
}

func NewPurchaseReservedInstancesOfferingResponse() (response *PurchaseReservedInstancesOfferingResponse) {
	response = &PurchaseReservedInstancesOfferingResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// PurchaseReservedInstancesOffering
// 本接口(PurchaseReservedInstancesOffering)用于用户购买一个或者多个指定配置的预留实例
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEOUTOFQUATA = "UnsupportedOperation.ReservedInstanceOutofQuata"
func (c *Client) PurchaseReservedInstancesOffering(request *PurchaseReservedInstancesOfferingRequest) (response *PurchaseReservedInstancesOfferingResponse, err error) {
	if request == nil {
		request = NewPurchaseReservedInstancesOfferingRequest()
	}

	response = NewPurchaseReservedInstancesOfferingResponse()
	err = c.Send(request, response)
	return
}

// PurchaseReservedInstancesOffering
// 本接口(PurchaseReservedInstancesOffering)用于用户购买一个或者多个指定配置的预留实例
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"
//	UNSUPPORTEDOPERATION_RESERVEDINSTANCEOUTOFQUATA = "UnsupportedOperation.ReservedInstanceOutofQuata"
func (c *Client) PurchaseReservedInstancesOfferingWithContext(ctx context.Context, request *PurchaseReservedInstancesOfferingRequest) (response *PurchaseReservedInstancesOfferingResponse, err error) {
	if request == nil {
		request = NewPurchaseReservedInstancesOfferingRequest()
	}
	request.SetContext(ctx)

	response = NewPurchaseReservedInstancesOfferingResponse()
	err = c.Send(request, response)
	return
}

func NewRebootInstancesRequest() (request *RebootInstancesRequest) {
	request = &RebootInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "RebootInstances")

	return
}

func NewRebootInstancesResponse() (response *RebootInstancesResponse) {
	response = &RebootInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RebootInstances
// 本接口 (RebootInstances) 用于重启实例。
//
// * 只有状态为`RUNNING`的实例才可以进行此操作。
//
// * 接口调用成功时，实例会进入`REBOOTING`状态；重启实例成功时，实例会进入`RUNNING`状态。
//
// * 支持强制重启。强制重启的效果等同于关闭物理计算机的电源开关再重新启动。强制重启可能会导致数据丢失或文件系统损坏，请仅在服务器不能正常重启时使用。
//
// * 支持批量操作，每次请求批量实例的上限为100。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) RebootInstances(request *RebootInstancesRequest) (response *RebootInstancesResponse, err error) {
	if request == nil {
		request = NewRebootInstancesRequest()
	}

	response = NewRebootInstancesResponse()
	err = c.Send(request, response)
	return
}

// RebootInstances
// 本接口 (RebootInstances) 用于重启实例。
//
// * 只有状态为`RUNNING`的实例才可以进行此操作。
//
// * 接口调用成功时，实例会进入`REBOOTING`状态；重启实例成功时，实例会进入`RUNNING`状态。
//
// * 支持强制重启。强制重启的效果等同于关闭物理计算机的电源开关再重新启动。强制重启可能会导致数据丢失或文件系统损坏，请仅在服务器不能正常重启时使用。
//
// * 支持批量操作，每次请求批量实例的上限为100。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) RebootInstancesWithContext(ctx context.Context, request *RebootInstancesRequest) (response *RebootInstancesResponse, err error) {
	if request == nil {
		request = NewRebootInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewRebootInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewRenewHostsRequest() (request *RenewHostsRequest) {
	request = &RenewHostsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "RenewHosts")

	return
}

func NewRenewHostsResponse() (response *RenewHostsResponse) {
	response = &RenewHostsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RenewHosts
// 本接口 (RenewHosts) 用于续费包年包月CDH实例。
//
// * 只支持操作包年包月实例，否则操作会以特定[错误码](#6.-.E9.94.99.E8.AF.AF.E7.A0.81)返回。
//
// * 续费时请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// 可能返回的错误码:
//
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDHOST_NOTSUPPORTED = "InvalidHost.NotSupported"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDPERIOD = "InvalidPeriod"
func (c *Client) RenewHosts(request *RenewHostsRequest) (response *RenewHostsResponse, err error) {
	if request == nil {
		request = NewRenewHostsRequest()
	}

	response = NewRenewHostsResponse()
	err = c.Send(request, response)
	return
}

// RenewHosts
// 本接口 (RenewHosts) 用于续费包年包月CDH实例。
//
// * 只支持操作包年包月实例，否则操作会以特定[错误码](#6.-.E9.94.99.E8.AF.AF.E7.A0.81)返回。
//
// * 续费时请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// 可能返回的错误码:
//
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDHOST_NOTSUPPORTED = "InvalidHost.NotSupported"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDPERIOD = "InvalidPeriod"
func (c *Client) RenewHostsWithContext(ctx context.Context, request *RenewHostsRequest) (response *RenewHostsResponse, err error) {
	if request == nil {
		request = NewRenewHostsRequest()
	}
	request.SetContext(ctx)

	response = NewRenewHostsResponse()
	err = c.Send(request, response)
	return
}

func NewRenewInstancesRequest() (request *RenewInstancesRequest) {
	request = &RenewInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "RenewInstances")

	return
}

func NewRenewInstancesResponse() (response *RenewInstancesResponse) {
	response = &RenewInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RenewInstances
// 本接口 (RenewInstances) 用于续费包年包月实例。
//
// * 只支持操作包年包月实例。
//
// * 续费时请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INVALIDINSTANCEAPPLICATIONROLEEMR = "FailedOperation.InvalidInstanceApplicationRoleEmr"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INSTANCENOTSUPPORTEDMIXPRICINGMODEL = "InvalidParameterValue.InstanceNotSupportedMixPricingModel"
//	INVALIDPERIOD = "InvalidPeriod"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
func (c *Client) RenewInstances(request *RenewInstancesRequest) (response *RenewInstancesResponse, err error) {
	if request == nil {
		request = NewRenewInstancesRequest()
	}

	response = NewRenewInstancesResponse()
	err = c.Send(request, response)
	return
}

// RenewInstances
// 本接口 (RenewInstances) 用于续费包年包月实例。
//
// * 只支持操作包年包月实例。
//
// * 续费时请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INVALIDINSTANCEAPPLICATIONROLEEMR = "FailedOperation.InvalidInstanceApplicationRoleEmr"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INSTANCENOTSUPPORTEDMIXPRICINGMODEL = "InvalidParameterValue.InstanceNotSupportedMixPricingModel"
//	INVALIDPERIOD = "InvalidPeriod"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
func (c *Client) RenewInstancesWithContext(ctx context.Context, request *RenewInstancesRequest) (response *RenewInstancesResponse, err error) {
	if request == nil {
		request = NewRenewInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewRenewInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewResetInstanceRequest() (request *ResetInstanceRequest) {
	request = &ResetInstanceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ResetInstance")

	return
}

func NewResetInstanceResponse() (response *ResetInstanceResponse) {
	response = &ResetInstanceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetInstance
// 本接口 (ResetInstance) 用于重装指定实例上的操作系统。
//
// * 如果指定了`ImageId`参数，则使用指定的镜像重装；否则按照当前实例使用的镜像进行重装。
//
// * 系统盘将会被格式化，并重置；请确保系统盘中无重要文件。
//
// * `Linux`和`Windows`系统互相切换时，该实例系统盘`ID`将发生变化，系统盘关联快照将无法回滚、恢复数据。
//
// * 密码不指定将会通过站内信下发随机密码。
//
// * 目前只支持[系统盘类型](https://cloud.tencent.com/document/api/213/9452#SystemDisk)是`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`类型的实例使用该接口实现`Linux`和`Windows`操作系统切换。
//
// * 目前不支持境外地域的实例使用该接口实现`Linux`和`Windows`操作系统切换。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER_HOSTNAMEILLEGAL = "InvalidParameter.HostNameIllegal"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEFORGIVENINSTANCETYPE = "InvalidParameterValue.InvalidImageForGivenInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDPASSWORD = "InvalidParameterValue.InvalidPassword"
//	INVALIDPARAMETERVALUE_KEYPAIRNOTFOUND = "InvalidParameterValue.KeyPairNotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPASSWORD = "InvalidPassword"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_CHCINSTALLCLOUDIMAGEWITHOUTDEPLOYNETWORK = "OperationDenied.ChcInstallCloudImageWithoutDeployNetwork"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	RESOURCEINUSE = "ResourceInUse"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"
//	UNSUPPORTEDOPERATION_RAWLOCALDISKINSREINSTALLTOQCOW2 = "UnsupportedOperation.RawLocalDiskInsReinstalltoQcow2"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ResetInstance(request *ResetInstanceRequest) (response *ResetInstanceResponse, err error) {
	if request == nil {
		request = NewResetInstanceRequest()
	}

	response = NewResetInstanceResponse()
	err = c.Send(request, response)
	return
}

// ResetInstance
// 本接口 (ResetInstance) 用于重装指定实例上的操作系统。
//
// * 如果指定了`ImageId`参数，则使用指定的镜像重装；否则按照当前实例使用的镜像进行重装。
//
// * 系统盘将会被格式化，并重置；请确保系统盘中无重要文件。
//
// * `Linux`和`Windows`系统互相切换时，该实例系统盘`ID`将发生变化，系统盘关联快照将无法回滚、恢复数据。
//
// * 密码不指定将会通过站内信下发随机密码。
//
// * 目前只支持[系统盘类型](https://cloud.tencent.com/document/api/213/9452#SystemDisk)是`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`类型的实例使用该接口实现`Linux`和`Windows`操作系统切换。
//
// * 目前不支持境外地域的实例使用该接口实现`Linux`和`Windows`操作系统切换。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER_HOSTNAMEILLEGAL = "InvalidParameter.HostNameIllegal"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEFORGIVENINSTANCETYPE = "InvalidParameterValue.InvalidImageForGivenInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDPASSWORD = "InvalidParameterValue.InvalidPassword"
//	INVALIDPARAMETERVALUE_KEYPAIRNOTFOUND = "InvalidParameterValue.KeyPairNotFound"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPASSWORD = "InvalidPassword"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_CHCINSTALLCLOUDIMAGEWITHOUTDEPLOYNETWORK = "OperationDenied.ChcInstallCloudImageWithoutDeployNetwork"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	RESOURCEINUSE = "ResourceInUse"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"
//	UNSUPPORTEDOPERATION_RAWLOCALDISKINSREINSTALLTOQCOW2 = "UnsupportedOperation.RawLocalDiskInsReinstalltoQcow2"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ResetInstanceWithContext(ctx context.Context, request *ResetInstanceRequest) (response *ResetInstanceResponse, err error) {
	if request == nil {
		request = NewResetInstanceRequest()
	}
	request.SetContext(ctx)

	response = NewResetInstanceResponse()
	err = c.Send(request, response)
	return
}

func NewResetInstancesInternetMaxBandwidthRequest() (request *ResetInstancesInternetMaxBandwidthRequest) {
	request = &ResetInstancesInternetMaxBandwidthRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ResetInstancesInternetMaxBandwidth")

	return
}

func NewResetInstancesInternetMaxBandwidthResponse() (response *ResetInstancesInternetMaxBandwidthResponse) {
	response = &ResetInstancesInternetMaxBandwidthResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetInstancesInternetMaxBandwidth
// 本接口 (ResetInstancesInternetMaxBandwidth) 用于调整实例公网带宽上限。
//
// * 不同机型带宽上限范围不一致，具体限制详见[公网带宽上限](https://cloud.tencent.com/document/product/213/12523)。
//
// * 对于 `BANDWIDTH_PREPAID` 计费方式的带宽，需要输入参数 `StartTime` 和 `EndTime` ，指定调整后的带宽的生效时间段。在这种场景下目前不支持调小带宽，会涉及扣费，请确保账户余额充足。可通过 [`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253) 接口查询账户余额。
//
// * 对于 `TRAFFIC_POSTPAID_BY_HOUR` 、 `BANDWIDTH_POSTPAID_BY_HOUR` 和 `BANDWIDTH_PACKAGE` 计费方式的带宽，使用该接口调整带宽上限是实时生效的，可以在带宽允许的范围内调大或者调小带宽，不支持输入参数 `StartTime` 和 `EndTime` 。
//
// * 接口不支持调整 `BANDWIDTH_POSTPAID_BY_MONTH` 计费方式的带宽。
//
// * 接口不支持批量调整 `BANDWIDTH_PREPAID` 和 `BANDWIDTH_POSTPAID_BY_HOUR` 计费方式的带宽。
//
// * 接口不支持批量调整混合计费方式的带宽。例如不支持同时调整 `TRAFFIC_POSTPAID_BY_HOUR` 和 `BANDWIDTH_PACKAGE` 计费方式的带宽。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOTFOUNDEIP = "FailedOperation.NotFoundEIP"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPERMISSION = "InvalidPermission"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ResetInstancesInternetMaxBandwidth(request *ResetInstancesInternetMaxBandwidthRequest) (response *ResetInstancesInternetMaxBandwidthResponse, err error) {
	if request == nil {
		request = NewResetInstancesInternetMaxBandwidthRequest()
	}

	response = NewResetInstancesInternetMaxBandwidthResponse()
	err = c.Send(request, response)
	return
}

// ResetInstancesInternetMaxBandwidth
// 本接口 (ResetInstancesInternetMaxBandwidth) 用于调整实例公网带宽上限。
//
// * 不同机型带宽上限范围不一致，具体限制详见[公网带宽上限](https://cloud.tencent.com/document/product/213/12523)。
//
// * 对于 `BANDWIDTH_PREPAID` 计费方式的带宽，需要输入参数 `StartTime` 和 `EndTime` ，指定调整后的带宽的生效时间段。在这种场景下目前不支持调小带宽，会涉及扣费，请确保账户余额充足。可通过 [`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253) 接口查询账户余额。
//
// * 对于 `TRAFFIC_POSTPAID_BY_HOUR` 、 `BANDWIDTH_POSTPAID_BY_HOUR` 和 `BANDWIDTH_PACKAGE` 计费方式的带宽，使用该接口调整带宽上限是实时生效的，可以在带宽允许的范围内调大或者调小带宽，不支持输入参数 `StartTime` 和 `EndTime` 。
//
// * 接口不支持调整 `BANDWIDTH_POSTPAID_BY_MONTH` 计费方式的带宽。
//
// * 接口不支持批量调整 `BANDWIDTH_PREPAID` 和 `BANDWIDTH_POSTPAID_BY_HOUR` 计费方式的带宽。
//
// * 接口不支持批量调整混合计费方式的带宽。例如不支持同时调整 `TRAFFIC_POSTPAID_BY_HOUR` 和 `BANDWIDTH_PACKAGE` 计费方式的带宽。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_NOTFOUNDEIP = "FailedOperation.NotFoundEIP"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPERMISSION = "InvalidPermission"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ResetInstancesInternetMaxBandwidthWithContext(ctx context.Context, request *ResetInstancesInternetMaxBandwidthRequest) (response *ResetInstancesInternetMaxBandwidthResponse, err error) {
	if request == nil {
		request = NewResetInstancesInternetMaxBandwidthRequest()
	}
	request.SetContext(ctx)

	response = NewResetInstancesInternetMaxBandwidthResponse()
	err = c.Send(request, response)
	return
}

func NewResetInstancesPasswordRequest() (request *ResetInstancesPasswordRequest) {
	request = &ResetInstancesPasswordRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ResetInstancesPassword")

	return
}

func NewResetInstancesPasswordResponse() (response *ResetInstancesPasswordResponse) {
	response = &ResetInstancesPasswordResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetInstancesPassword
// 本接口 (ResetInstancesPassword) 用于将实例操作系统的密码重置为用户指定的密码。
//
// *如果是修改系统管理云密码：实例的操作系统不同，管理员帐号也会不一样(`Windows`为`Administrator`，`Ubuntu`为`ubuntu`，其它系统为`root`)。
//
// * 重置处于运行中状态的实例密码，需要设置关机参数`ForceStop`为`TRUE`。如果没有显式指定强制关机参数，则只有处于关机状态的实例才允许执行重置密码操作。
//
// * 支持批量操作。将多个实例操作系统的密码重置为相同的密码。每次请求批量实例的上限为100。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDPASSWORD = "InvalidParameterValue.InvalidPassword"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPASSWORD = "InvalidPassword"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ResetInstancesPassword(request *ResetInstancesPasswordRequest) (response *ResetInstancesPasswordResponse, err error) {
	if request == nil {
		request = NewResetInstancesPasswordRequest()
	}

	response = NewResetInstancesPasswordResponse()
	err = c.Send(request, response)
	return
}

// ResetInstancesPassword
// 本接口 (ResetInstancesPassword) 用于将实例操作系统的密码重置为用户指定的密码。
//
// *如果是修改系统管理云密码：实例的操作系统不同，管理员帐号也会不一样(`Windows`为`Administrator`，`Ubuntu`为`ubuntu`，其它系统为`root`)。
//
// * 重置处于运行中状态的实例密码，需要设置关机参数`ForceStop`为`TRUE`。如果没有显式指定强制关机参数，则只有处于关机状态的实例才允许执行重置密码操作。
//
// * 支持批量操作。将多个实例操作系统的密码重置为相同的密码。每次请求批量实例的上限为100。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_INVALIDPASSWORD = "InvalidParameterValue.InvalidPassword"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"
//	INVALIDPASSWORD = "InvalidPassword"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ResetInstancesPasswordWithContext(ctx context.Context, request *ResetInstancesPasswordRequest) (response *ResetInstancesPasswordResponse, err error) {
	if request == nil {
		request = NewResetInstancesPasswordRequest()
	}
	request.SetContext(ctx)

	response = NewResetInstancesPasswordResponse()
	err = c.Send(request, response)
	return
}

func NewResetInstancesTypeRequest() (request *ResetInstancesTypeRequest) {
	request = &ResetInstancesTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ResetInstancesType")

	return
}

func NewResetInstancesTypeResponse() (response *ResetInstancesTypeResponse) {
	response = &ResetInstancesTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResetInstancesType
// 本接口 (ResetInstancesType) 用于调整实例的机型。
//
// * 目前只支持[系统盘类型](/document/api/213/9452#block_device)是CLOUD_BASIC、CLOUD_PREMIUM、CLOUD_SSD类型的实例使用该接口进行机型调整。
//
// * 目前不支持[CDH](https://cloud.tencent.com/document/product/416)实例使用该接口调整机型。对于包年包月实例，使用该接口会涉及扣费，请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// * 本接口为异步接口，调整实例配置请求发送成功后会返回一个RequestId，此时操作并未立即完成。实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表调整实例配置操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_PROMOTIONALPERIORESTRICTION = "FailedOperation.PromotionalPerioRestriction"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_GPUINSTANCEFAMILY = "InvalidParameterValue.GPUInstanceFamily"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_ENINUMLIMIT = "LimitExceeded.EniNumLimit"
//	LIMITEXCEEDED_INSTANCETYPEBANDWIDTH = "LimitExceeded.InstanceTypeBandwidth"
//	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	RESOURCESSOLDOUT_AVAILABLEZONE = "ResourcesSoldOut.AvailableZone"
//	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILY = "UnsupportedOperation.UnsupportedChangeInstanceFamily"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILYTOARM = "UnsupportedOperation.UnsupportedChangeInstanceFamilyToARM"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILYTOSA3 = "UnsupportedOperation.UnsupportedChangeInstanceFamilyToSA3"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCETOTHISINSTANCEFAMILY = "UnsupportedOperation.UnsupportedChangeInstanceToThisInstanceFamily"
func (c *Client) ResetInstancesType(request *ResetInstancesTypeRequest) (response *ResetInstancesTypeResponse, err error) {
	if request == nil {
		request = NewResetInstancesTypeRequest()
	}

	response = NewResetInstancesTypeResponse()
	err = c.Send(request, response)
	return
}

// ResetInstancesType
// 本接口 (ResetInstancesType) 用于调整实例的机型。
//
// * 目前只支持[系统盘类型](/document/api/213/9452#block_device)是CLOUD_BASIC、CLOUD_PREMIUM、CLOUD_SSD类型的实例使用该接口进行机型调整。
//
// * 目前不支持[CDH](https://cloud.tencent.com/document/product/416)实例使用该接口调整机型。对于包年包月实例，使用该接口会涉及扣费，请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// * 本接口为异步接口，调整实例配置请求发送成功后会返回一个RequestId，此时操作并未立即完成。实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表调整实例配置操作成功。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_PROMOTIONALPERIORESTRICTION = "FailedOperation.PromotionalPerioRestriction"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_GPUINSTANCEFAMILY = "InvalidParameterValue.GPUInstanceFamily"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_ENINUMLIMIT = "LimitExceeded.EniNumLimit"
//	LIMITEXCEEDED_INSTANCETYPEBANDWIDTH = "LimitExceeded.InstanceTypeBandwidth"
//	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	RESOURCESSOLDOUT_AVAILABLEZONE = "ResourcesSoldOut.AvailableZone"
//	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILY = "UnsupportedOperation.UnsupportedChangeInstanceFamily"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILYTOARM = "UnsupportedOperation.UnsupportedChangeInstanceFamilyToARM"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILYTOSA3 = "UnsupportedOperation.UnsupportedChangeInstanceFamilyToSA3"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCETOTHISINSTANCEFAMILY = "UnsupportedOperation.UnsupportedChangeInstanceToThisInstanceFamily"
func (c *Client) ResetInstancesTypeWithContext(ctx context.Context, request *ResetInstancesTypeRequest) (response *ResetInstancesTypeResponse, err error) {
	if request == nil {
		request = NewResetInstancesTypeRequest()
	}
	request.SetContext(ctx)

	response = NewResetInstancesTypeResponse()
	err = c.Send(request, response)
	return
}

func NewResizeInstanceDisksRequest() (request *ResizeInstanceDisksRequest) {
	request = &ResizeInstanceDisksRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "ResizeInstanceDisks")

	return
}

func NewResizeInstanceDisksResponse() (response *ResizeInstanceDisksResponse) {
	response = &ResizeInstanceDisksResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ResizeInstanceDisks
// 本接口 (ResizeInstanceDisks) 用于扩容实例的数据盘。
//
// * 目前只支持扩容非弹性盘（[`DescribeDisks`](https://cloud.tencent.com/document/api/362/16315)接口返回值中的`Portable`为`false`表示非弹性），且[数据盘类型](https://cloud.tencent.com/document/api/213/15753#DataDisk)为：`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`和[CDH](https://cloud.tencent.com/document/product/416)实例的`LOCAL_BASIC`、`LOCAL_SSD`类型数据盘。
//
// * 对于包年包月实例，使用该接口会涉及扣费，请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// * 目前只支持扩容一块数据盘。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// * 如果是系统盘，目前只支持扩容，不支持缩容。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_ATLEASTONE = "MissingParameter.AtLeastOne"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ResizeInstanceDisks(request *ResizeInstanceDisksRequest) (response *ResizeInstanceDisksResponse, err error) {
	if request == nil {
		request = NewResizeInstanceDisksRequest()
	}

	response = NewResizeInstanceDisksResponse()
	err = c.Send(request, response)
	return
}

// ResizeInstanceDisks
// 本接口 (ResizeInstanceDisks) 用于扩容实例的数据盘。
//
// * 目前只支持扩容非弹性盘（[`DescribeDisks`](https://cloud.tencent.com/document/api/362/16315)接口返回值中的`Portable`为`false`表示非弹性），且[数据盘类型](https://cloud.tencent.com/document/api/213/15753#DataDisk)为：`CLOUD_BASIC`、`CLOUD_PREMIUM`、`CLOUD_SSD`和[CDH](https://cloud.tencent.com/document/product/416)实例的`LOCAL_BASIC`、`LOCAL_SSD`类型数据盘。
//
// * 对于包年包月实例，使用该接口会涉及扣费，请确保账户余额充足。可通过[`DescribeAccountBalance`](https://cloud.tencent.com/document/product/555/20253)接口查询账户余额。
//
// * 目前只支持扩容一块数据盘。
//
// * 实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表操作成功。
//
// * 如果是系统盘，目前只支持扩容，不支持缩容。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_ATLEASTONE = "MissingParameter.AtLeastOne"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) ResizeInstanceDisksWithContext(ctx context.Context, request *ResizeInstanceDisksRequest) (response *ResizeInstanceDisksResponse, err error) {
	if request == nil {
		request = NewResizeInstanceDisksRequest()
	}
	request.SetContext(ctx)

	response = NewResizeInstanceDisksResponse()
	err = c.Send(request, response)
	return
}

func NewRunInstancesRequest() (request *RunInstancesRequest) {
	request = &RunInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "RunInstances")

	return
}

func NewRunInstancesResponse() (response *RunInstancesResponse) {
	response = &RunInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RunInstances
// 本接口 (RunInstances) 用于创建一个或多个指定配置的实例。
//
// * 实例创建成功后将自动开机启动，[实例状态](https://cloud.tencent.com/document/product/213/15753#InstanceStatus)变为“运行中”。
//
// * 预付费实例的购买会预先扣除本次实例购买所需金额，按小时后付费实例购买会预先冻结本次实例购买一小时内所需金额，在调用本接口前请确保账户余额充足。
//
// * 调用本接口创建实例，支持代金券自动抵扣（注意，代金券不可用于抵扣后付费冻结金额），详情请参考[代金券选用规则](https://cloud.tencent.com/document/product/555/7428)。
//
// * 本接口允许购买的实例数量遵循[CVM实例购买限制](https://cloud.tencent.com/document/product/213/2664)，所创建的实例和官网入口创建的实例共用配额。
//
// * 本接口为异步接口，当创建实例请求下发成功后会返回一个实例`ID`列表和一个`RequestId`，此时创建实例操作并未立即完成。在此期间实例的状态将会处于“PENDING”，实例创建结果可以通过调用 [DescribeInstancesStatus](https://cloud.tencent.com/document/product/213/15738)  接口查询，如果实例状态(InstanceState)由“PENDING”变为“RUNNING”，则代表实例创建成功，“LAUNCH_FAILED”代表实例创建失败。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"
//	FAILEDOPERATION_ILLEGALTAGVALUE = "FailedOperation.IllegalTagValue"
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_NOAVAILABLEIPADDRESSCOUNTINSUBNET = "FailedOperation.NoAvailableIpAddressCountInSubnet"
//	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"
//	FAILEDOPERATION_SNAPSHOTSIZELARGERTHANDATASIZE = "FailedOperation.SnapshotSizeLargerThanDataSize"
//	FAILEDOPERATION_SNAPSHOTSIZELESSTHANDATASIZE = "FailedOperation.SnapshotSizeLessThanDataSize"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	INSTANCESQUOTALIMITEXCEEDED = "InstancesQuotaLimitExceeded"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_HOSTIDSTATUSNOTSUPPORT = "InvalidParameter.HostIdStatusNotSupport"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_INVALIDIPFORMAT = "InvalidParameter.InvalidIpFormat"
//	INVALIDPARAMETER_LACKCORECOUNTORTHREADPERCORE = "InvalidParameter.LackCoreCountOrThreadPerCore"
//	INVALIDPARAMETER_PARAMETERCONFLICT = "InvalidParameter.ParameterConflict"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDNOTFOUND = "InvalidParameterValue.BandwidthPackageIdNotFound"
//	INVALIDPARAMETERVALUE_CHCHOSTSNOTFOUND = "InvalidParameterValue.ChcHostsNotFound"
//	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"
//	INVALIDPARAMETERVALUE_CORECOUNTVALUE = "InvalidParameterValue.CoreCountValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_IPADDRESSMALFORMED = "InvalidParameterValue.IPAddressMalformed"
//	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"
//	INVALIDPARAMETERVALUE_INSTANCETYPEREQUIREDHPCCLUSTER = "InvalidParameterValue.InstanceTypeRequiredHpcCluster"
//	INVALIDPARAMETERVALUE_INSUFFICIENTOFFERING = "InvalidParameterValue.InsufficientOffering"
//	INVALIDPARAMETERVALUE_INSUFFICIENTPRICE = "InvalidParameterValue.InsufficientPrice"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEFORGIVENINSTANCETYPE = "InvalidParameterValue.InvalidImageForGivenInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEFORMAT = "InvalidParameterValue.InvalidImageFormat"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEOSNAME = "InvalidParameterValue.InvalidImageOsName"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_INVALIDPASSWORD = "InvalidParameterValue.InvalidPassword"
//	INVALIDPARAMETERVALUE_INVALIDTIMEFORMAT = "InvalidParameterValue.InvalidTimeFormat"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_KEYPAIRNOTFOUND = "InvalidParameterValue.KeyPairNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"
//	INVALIDPARAMETERVALUE_SUBNETIDMALFORMED = "InvalidParameterValue.SubnetIdMalformed"
//	INVALIDPARAMETERVALUE_SUBNETNOTEXIST = "InvalidParameterValue.SubnetNotExist"
//	INVALIDPARAMETERVALUE_TAGKEYNOTFOUND = "InvalidParameterValue.TagKeyNotFound"
//	INVALIDPARAMETERVALUE_THREADPERCOREVALUE = "InvalidParameterValue.ThreadPerCoreValue"
//	INVALIDPARAMETERVALUE_VPCIDMALFORMED = "InvalidParameterValue.VpcIdMalformed"
//	INVALIDPARAMETERVALUE_VPCIDNOTEXIST = "InvalidParameterValue.VpcIdNotExist"
//	INVALIDPARAMETERVALUE_VPCIDZONEIDNOTMATCH = "InvalidParameterValue.VpcIdZoneIdNotMatch"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPASSWORD = "InvalidPassword"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_CVMSVIFSPERSECGROUPLIMITEXCEEDED = "LimitExceeded.CvmsVifsPerSecGroupLimitExceeded"
//	LIMITEXCEEDED_DISASTERRECOVERGROUP = "LimitExceeded.DisasterRecoverGroup"
//	LIMITEXCEEDED_IPV6ADDRESSNUM = "LimitExceeded.IPv6AddressNum"
//	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"
//	LIMITEXCEEDED_PREPAYQUOTA = "LimitExceeded.PrepayQuota"
//	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"
//	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"
//	LIMITEXCEEDED_USERSPOTQUOTA = "LimitExceeded.UserSpotQuota"
//	LIMITEXCEEDED_VPCSUBNETNUM = "LimitExceeded.VpcSubnetNum"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_DPDKINSTANCETYPEREQUIREDVPC = "MissingParameter.DPDKInstanceTypeRequiredVPC"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	OPERATIONDENIED_CHCINSTALLCLOUDIMAGEWITHOUTDEPLOYNETWORK = "OperationDenied.ChcInstallCloudImageWithoutDeployNetwork"
//	RESOURCEINSUFFICIENT_AVAILABILITYZONESOLDOUT = "ResourceInsufficient.AvailabilityZoneSoldOut"
//	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"
//	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	RESOURCESSOLDOUT_EIPINSUFFICIENT = "ResourcesSoldOut.EipInsufficient"
//	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
//	UNSUPPORTEDOPERATION_INVALIDREGIONDISKENCRYPT = "UnsupportedOperation.InvalidRegionDiskEncrypt"
//	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_NOTSUPPORTIMPORTINSTANCESACTIONTIMER = "UnsupportedOperation.NotSupportImportInstancesActionTimer"
//	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"
//	UNSUPPORTEDOPERATION_RAWLOCALDISKINSREINSTALLTOQCOW2 = "UnsupportedOperation.RawLocalDiskInsReinstalltoQcow2"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDINTERNATIONALUSER = "UnsupportedOperation.UnsupportedInternationalUser"
//	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"
//	VPCIPISUSED = "VpcIpIsUsed"
func (c *Client) RunInstances(request *RunInstancesRequest) (response *RunInstancesResponse, err error) {
	if request == nil {
		request = NewRunInstancesRequest()
	}

	response = NewRunInstancesResponse()
	err = c.Send(request, response)
	return
}

// RunInstances
// 本接口 (RunInstances) 用于创建一个或多个指定配置的实例。
//
// * 实例创建成功后将自动开机启动，[实例状态](https://cloud.tencent.com/document/product/213/15753#InstanceStatus)变为“运行中”。
//
// * 预付费实例的购买会预先扣除本次实例购买所需金额，按小时后付费实例购买会预先冻结本次实例购买一小时内所需金额，在调用本接口前请确保账户余额充足。
//
// * 调用本接口创建实例，支持代金券自动抵扣（注意，代金券不可用于抵扣后付费冻结金额），详情请参考[代金券选用规则](https://cloud.tencent.com/document/product/555/7428)。
//
// * 本接口允许购买的实例数量遵循[CVM实例购买限制](https://cloud.tencent.com/document/product/213/2664)，所创建的实例和官网入口创建的实例共用配额。
//
// * 本接口为异步接口，当创建实例请求下发成功后会返回一个实例`ID`列表和一个`RequestId`，此时创建实例操作并未立即完成。在此期间实例的状态将会处于“PENDING”，实例创建结果可以通过调用 [DescribeInstancesStatus](https://cloud.tencent.com/document/product/213/15738)  接口查询，如果实例状态(InstanceState)由“PENDING”变为“RUNNING”，则代表实例创建成功，“LAUNCH_FAILED”代表实例创建失败。
//
// 可能返回的错误码:
//
//	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"
//	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"
//	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"
//	FAILEDOPERATION_ILLEGALTAGVALUE = "FailedOperation.IllegalTagValue"
//	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"
//	FAILEDOPERATION_NOAVAILABLEIPADDRESSCOUNTINSUBNET = "FailedOperation.NoAvailableIpAddressCountInSubnet"
//	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"
//	FAILEDOPERATION_SNAPSHOTSIZELARGERTHANDATASIZE = "FailedOperation.SnapshotSizeLargerThanDataSize"
//	FAILEDOPERATION_SNAPSHOTSIZELESSTHANDATASIZE = "FailedOperation.SnapshotSizeLessThanDataSize"
//	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"
//	INSTANCESQUOTALIMITEXCEEDED = "InstancesQuotaLimitExceeded"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"
//	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"
//	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"
//	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"
//	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"
//	INVALIDPARAMETER_HOSTIDSTATUSNOTSUPPORT = "InvalidParameter.HostIdStatusNotSupport"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETER_INVALIDIPFORMAT = "InvalidParameter.InvalidIpFormat"
//	INVALIDPARAMETER_LACKCORECOUNTORTHREADPERCORE = "InvalidParameter.LackCoreCountOrThreadPerCore"
//	INVALIDPARAMETER_PARAMETERCONFLICT = "InvalidParameter.ParameterConflict"
//	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"
//	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDNOTFOUND = "InvalidParameterValue.BandwidthPackageIdNotFound"
//	INVALIDPARAMETERVALUE_CHCHOSTSNOTFOUND = "InvalidParameterValue.ChcHostsNotFound"
//	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"
//	INVALIDPARAMETERVALUE_CORECOUNTVALUE = "InvalidParameterValue.CoreCountValue"
//	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"
//	INVALIDPARAMETERVALUE_IPADDRESSMALFORMED = "InvalidParameterValue.IPAddressMalformed"
//	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"
//	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"
//	INVALIDPARAMETERVALUE_INSTANCETYPEREQUIREDHPCCLUSTER = "InvalidParameterValue.InstanceTypeRequiredHpcCluster"
//	INVALIDPARAMETERVALUE_INSUFFICIENTOFFERING = "InvalidParameterValue.InsufficientOffering"
//	INVALIDPARAMETERVALUE_INSUFFICIENTPRICE = "InvalidParameterValue.InsufficientPrice"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEFORGIVENINSTANCETYPE = "InvalidParameterValue.InvalidImageForGivenInstanceType"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEFORMAT = "InvalidParameterValue.InvalidImageFormat"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEOSNAME = "InvalidParameterValue.InvalidImageOsName"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"
//	INVALIDPARAMETERVALUE_INVALIDPASSWORD = "InvalidParameterValue.InvalidPassword"
//	INVALIDPARAMETERVALUE_INVALIDTIMEFORMAT = "InvalidParameterValue.InvalidTimeFormat"
//	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"
//	INVALIDPARAMETERVALUE_KEYPAIRNOTFOUND = "InvalidParameterValue.KeyPairNotFound"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"
//	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"
//	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"
//	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"
//	INVALIDPARAMETERVALUE_SUBNETIDMALFORMED = "InvalidParameterValue.SubnetIdMalformed"
//	INVALIDPARAMETERVALUE_SUBNETNOTEXIST = "InvalidParameterValue.SubnetNotExist"
//	INVALIDPARAMETERVALUE_TAGKEYNOTFOUND = "InvalidParameterValue.TagKeyNotFound"
//	INVALIDPARAMETERVALUE_THREADPERCOREVALUE = "InvalidParameterValue.ThreadPerCoreValue"
//	INVALIDPARAMETERVALUE_VPCIDMALFORMED = "InvalidParameterValue.VpcIdMalformed"
//	INVALIDPARAMETERVALUE_VPCIDNOTEXIST = "InvalidParameterValue.VpcIdNotExist"
//	INVALIDPARAMETERVALUE_VPCIDZONEIDNOTMATCH = "InvalidParameterValue.VpcIdZoneIdNotMatch"
//	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"
//	INVALIDPASSWORD = "InvalidPassword"
//	INVALIDPERIOD = "InvalidPeriod"
//	INVALIDPERMISSION = "InvalidPermission"
//	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"
//	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"
//	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"
//	LIMITEXCEEDED_CVMSVIFSPERSECGROUPLIMITEXCEEDED = "LimitExceeded.CvmsVifsPerSecGroupLimitExceeded"
//	LIMITEXCEEDED_DISASTERRECOVERGROUP = "LimitExceeded.DisasterRecoverGroup"
//	LIMITEXCEEDED_IPV6ADDRESSNUM = "LimitExceeded.IPv6AddressNum"
//	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"
//	LIMITEXCEEDED_PREPAYQUOTA = "LimitExceeded.PrepayQuota"
//	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"
//	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"
//	LIMITEXCEEDED_USERSPOTQUOTA = "LimitExceeded.UserSpotQuota"
//	LIMITEXCEEDED_VPCSUBNETNUM = "LimitExceeded.VpcSubnetNum"
//	MISSINGPARAMETER = "MissingParameter"
//	MISSINGPARAMETER_DPDKINSTANCETYPEREQUIREDVPC = "MissingParameter.DPDKInstanceTypeRequiredVPC"
//	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"
//	OPERATIONDENIED_CHCINSTALLCLOUDIMAGEWITHOUTDEPLOYNETWORK = "OperationDenied.ChcInstallCloudImageWithoutDeployNetwork"
//	RESOURCEINSUFFICIENT_AVAILABILITYZONESOLDOUT = "ResourceInsufficient.AvailabilityZoneSoldOut"
//	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"
//	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"
//	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"
//	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"
//	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"
//	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"
//	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"
//	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"
//	RESOURCESSOLDOUT_EIPINSUFFICIENT = "ResourcesSoldOut.EipInsufficient"
//	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"
//	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"
//	UNSUPPORTEDOPERATION_INVALIDREGIONDISKENCRYPT = "UnsupportedOperation.InvalidRegionDiskEncrypt"
//	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"
//	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"
//	UNSUPPORTEDOPERATION_NOTSUPPORTIMPORTINSTANCESACTIONTIMER = "UnsupportedOperation.NotSupportImportInstancesActionTimer"
//	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"
//	UNSUPPORTEDOPERATION_RAWLOCALDISKINSREINSTALLTOQCOW2 = "UnsupportedOperation.RawLocalDiskInsReinstalltoQcow2"
//	UNSUPPORTEDOPERATION_UNSUPPORTEDINTERNATIONALUSER = "UnsupportedOperation.UnsupportedInternationalUser"
//	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"
//	VPCIPISUSED = "VpcIpIsUsed"
func (c *Client) RunInstancesWithContext(ctx context.Context, request *RunInstancesRequest) (response *RunInstancesResponse, err error) {
	if request == nil {
		request = NewRunInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewRunInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewStartInstancesRequest() (request *StartInstancesRequest) {
	request = &StartInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "StartInstances")

	return
}

func NewStartInstancesResponse() (response *StartInstancesResponse) {
	response = &StartInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// StartInstances
// 本接口 (StartInstances) 用于启动一个或多个实例。
//
// * 只有状态为`STOPPED`的实例才可以进行此操作。
//
// * 接口调用成功时，实例会进入`STARTING`状态；启动实例成功时，实例会进入`RUNNING`状态。
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 本接口为异步接口，启动实例请求发送成功后会返回一个RequestId，此时操作并未立即完成。实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表启动实例操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
func (c *Client) StartInstances(request *StartInstancesRequest) (response *StartInstancesResponse, err error) {
	if request == nil {
		request = NewStartInstancesRequest()
	}

	response = NewStartInstancesResponse()
	err = c.Send(request, response)
	return
}

// StartInstances
// 本接口 (StartInstances) 用于启动一个或多个实例。
//
// * 只有状态为`STOPPED`的实例才可以进行此操作。
//
// * 接口调用成功时，实例会进入`STARTING`状态；启动实例成功时，实例会进入`RUNNING`状态。
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 本接口为异步接口，启动实例请求发送成功后会返回一个RequestId，此时操作并未立即完成。实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表启动实例操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
func (c *Client) StartInstancesWithContext(ctx context.Context, request *StartInstancesRequest) (response *StartInstancesResponse, err error) {
	if request == nil {
		request = NewStartInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewStartInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewStopInstancesRequest() (request *StopInstancesRequest) {
	request = &StopInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "StopInstances")

	return
}

func NewStopInstancesResponse() (response *StopInstancesResponse) {
	response = &StopInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// StopInstances
// 本接口 (StopInstances) 用于关闭一个或多个实例。
//
// * 只有状态为`RUNNING`的实例才可以进行此操作。
//
// * 接口调用成功时，实例会进入`STOPPING`状态；关闭实例成功时，实例会进入`STOPPED`状态。
//
// * 支持强制关闭。强制关机的效果等同于关闭物理计算机的电源开关。强制关机可能会导致数据丢失或文件系统损坏，请仅在服务器不能正常关机时使用。
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 本接口为异步接口，关闭实例请求发送成功后会返回一个RequestId，此时操作并未立即完成。实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表关闭实例操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) StopInstances(request *StopInstancesRequest) (response *StopInstancesResponse, err error) {
	if request == nil {
		request = NewStopInstancesRequest()
	}

	response = NewStopInstancesResponse()
	err = c.Send(request, response)
	return
}

// StopInstances
// 本接口 (StopInstances) 用于关闭一个或多个实例。
//
// * 只有状态为`RUNNING`的实例才可以进行此操作。
//
// * 接口调用成功时，实例会进入`STOPPING`状态；关闭实例成功时，实例会进入`STOPPED`状态。
//
// * 支持强制关闭。强制关机的效果等同于关闭物理计算机的电源开关。强制关机可能会导致数据丢失或文件系统损坏，请仅在服务器不能正常关机时使用。
//
// * 支持批量操作。每次请求批量实例的上限为100。
//
// * 本接口为异步接口，关闭实例请求发送成功后会返回一个RequestId，此时操作并未立即完成。实例操作结果可以通过调用 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728#.E7.A4.BA.E4.BE.8B3-.E6.9F.A5.E8.AF.A2.E5.AE.9E.E4.BE.8B.E7.9A.84.E6.9C.80.E6.96.B0.E6.93.8D.E4.BD.9C.E6.83.85.E5.86.B5) 接口查询，如果实例的最新操作状态(LatestOperationState)为“SUCCESS”，则代表关闭实例操作成功。
//
// 可能返回的错误码:
//
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"
func (c *Client) StopInstancesWithContext(ctx context.Context, request *StopInstancesRequest) (response *StopInstancesResponse, err error) {
	if request == nil {
		request = NewStopInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewStopInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewSyncImagesRequest() (request *SyncImagesRequest) {
	request = &SyncImagesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "SyncImages")

	return
}

func NewSyncImagesResponse() (response *SyncImagesResponse) {
	response = &SyncImagesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// SyncImages
// 本接口（SyncImages）用于将自定义镜像同步到其它地区。
//
// * 该接口每次调用只支持同步一个镜像。
//
// * 该接口支持多个同步地域。
//
// * 单个帐号在每个地域最多支持存在10个自定义镜像。
//
// 可能返回的错误码:
//
//	IMAGEQUOTALIMITEXCEEDED = "ImageQuotaLimitExceeded"
//	INVALIDIMAGEID_INCORRECTSTATE = "InvalidImageId.IncorrectState"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDIMAGEID_TOOLARGE = "InvalidImageId.TooLarge"
//	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
//	INVALIDREGION_UNAVAILABLE = "InvalidRegion.Unavailable"
//	UNSUPPORTEDOPERATION_REGION = "UnsupportedOperation.Region"
func (c *Client) SyncImages(request *SyncImagesRequest) (response *SyncImagesResponse, err error) {
	if request == nil {
		request = NewSyncImagesRequest()
	}

	response = NewSyncImagesResponse()
	err = c.Send(request, response)
	return
}

// SyncImages
// 本接口（SyncImages）用于将自定义镜像同步到其它地区。
//
// * 该接口每次调用只支持同步一个镜像。
//
// * 该接口支持多个同步地域。
//
// * 单个帐号在每个地域最多支持存在10个自定义镜像。
//
// 可能返回的错误码:
//
//	IMAGEQUOTALIMITEXCEEDED = "ImageQuotaLimitExceeded"
//	INVALIDIMAGEID_INCORRECTSTATE = "InvalidImageId.IncorrectState"
//	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"
//	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"
//	INVALIDIMAGEID_TOOLARGE = "InvalidImageId.TooLarge"
//	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"
//	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"
//	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"
//	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"
//	INVALIDREGION_UNAVAILABLE = "InvalidRegion.Unavailable"
//	UNSUPPORTEDOPERATION_REGION = "UnsupportedOperation.Region"
func (c *Client) SyncImagesWithContext(ctx context.Context, request *SyncImagesRequest) (response *SyncImagesResponse, err error) {
	if request == nil {
		request = NewSyncImagesRequest()
	}
	request.SetContext(ctx)

	response = NewSyncImagesResponse()
	err = c.Send(request, response)
	return
}

func NewTerminateInstancesRequest() (request *TerminateInstancesRequest) {
	request = &TerminateInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("cvm", APIVersion, "TerminateInstances")

	return
}

func NewTerminateInstancesResponse() (response *TerminateInstancesResponse) {
	response = &TerminateInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// TerminateInstances
// 本接口 (TerminateInstances) 用于主动退还实例。
//
// * 不再使用的实例，可通过本接口主动退还。
//
// * 按量计费的实例通过本接口可直接退还；包年包月实例如符合[退还规则](https://cloud.tencent.com/document/product/213/9711)，也可通过本接口主动退还。
//
// * 包年包月实例首次调用本接口，实例将被移至回收站，再次调用本接口，实例将被销毁，且不可恢复。按量计费实例调用本接口将被直接销毁
//
// * 支持批量操作，每次请求批量实例的上限为100。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INVALIDINSTANCEAPPLICATIONROLEEMR = "FailedOperation.InvalidInstanceApplicationRoleEmr"
//	FAILEDOPERATION_UNRETURNABLE = "FailedOperation.Unreturnable"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCENOTSUPPORTEDPREPAIDINSTANCE = "InvalidInstanceNotSupportedPrepaidInstance"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_NOTSUPPORTED = "InvalidParameterValue.NotSupported"
//	LIMITEXCEEDED_USERRETURNQUOTA = "LimitExceeded.UserReturnQuota"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCEMIXEDPRICINGMODEL = "UnsupportedOperation.InstanceMixedPricingModel"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_REGION = "UnsupportedOperation.Region"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_USERLIMITOPERATIONEXCEEDQUOTA = "UnsupportedOperation.UserLimitOperationExceedQuota"
func (c *Client) TerminateInstances(request *TerminateInstancesRequest) (response *TerminateInstancesResponse, err error) {
	if request == nil {
		request = NewTerminateInstancesRequest()
	}

	response = NewTerminateInstancesResponse()
	err = c.Send(request, response)
	return
}

// TerminateInstances
// 本接口 (TerminateInstances) 用于主动退还实例。
//
// * 不再使用的实例，可通过本接口主动退还。
//
// * 按量计费的实例通过本接口可直接退还；包年包月实例如符合[退还规则](https://cloud.tencent.com/document/product/213/9711)，也可通过本接口主动退还。
//
// * 包年包月实例首次调用本接口，实例将被移至回收站，再次调用本接口，实例将被销毁，且不可恢复。按量计费实例调用本接口将被直接销毁
//
// * 支持批量操作，每次请求批量实例的上限为100。
//
// 可能返回的错误码:
//
//	FAILEDOPERATION_INVALIDINSTANCEAPPLICATIONROLEEMR = "FailedOperation.InvalidInstanceApplicationRoleEmr"
//	FAILEDOPERATION_UNRETURNABLE = "FailedOperation.Unreturnable"
//	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"
//	INTERNALSERVERERROR = "InternalServerError"
//	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"
//	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"
//	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"
//	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"
//	INVALIDINSTANCENOTSUPPORTEDPREPAIDINSTANCE = "InvalidInstanceNotSupportedPrepaidInstance"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
//	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"
//	INVALIDPARAMETERVALUE_NOTSUPPORTED = "InvalidParameterValue.NotSupported"
//	LIMITEXCEEDED_USERRETURNQUOTA = "LimitExceeded.UserReturnQuota"
//	MISSINGPARAMETER = "MissingParameter"
//	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"
//	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"
//	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"
//	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"
//	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"
//	UNSUPPORTEDOPERATION_INSTANCEMIXEDPRICINGMODEL = "UnsupportedOperation.InstanceMixedPricingModel"
//	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"
//	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"
//	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"
//	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"
//	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"
//	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"
//	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"
//	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"
//	UNSUPPORTEDOPERATION_REGION = "UnsupportedOperation.Region"
//	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"
//	UNSUPPORTEDOPERATION_USERLIMITOPERATIONEXCEEDQUOTA = "UnsupportedOperation.UserLimitOperationExceedQuota"
func (c *Client) TerminateInstancesWithContext(ctx context.Context, request *TerminateInstancesRequest) (response *TerminateInstancesResponse, err error) {
	if request == nil {
		request = NewTerminateInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewTerminateInstancesResponse()
	err = c.Send(request, response)
	return
}
