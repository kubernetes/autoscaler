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

package v20180525

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = "2018-05-25"

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

func NewAcquireClusterAdminRoleRequest() (request *AcquireClusterAdminRoleRequest) {
	request = &AcquireClusterAdminRoleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "AcquireClusterAdminRole")

	return
}

func NewAcquireClusterAdminRoleResponse() (response *AcquireClusterAdminRoleResponse) {
	response = &AcquireClusterAdminRoleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AcquireClusterAdminRole
// 通过此接口，可以获取集群的tke:admin的ClusterRole，即管理员角色，可以用于CAM侧高权限的用户，通过CAM策略给予子账户此接口权限，进而可以通过此接口直接获取到kubernetes集群内的管理员角色。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_KUBERNETESCLIENTBUILDERROR = "InternalError.KubernetesClientBuildError"
//	INTERNALERROR_KUBERNETESCREATEOPERATIONERROR = "InternalError.KubernetesCreateOperationError"
//	INTERNALERROR_KUBERNETESGETOPERATIONERROR = "InternalError.KubernetesGetOperationError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
//	UNSUPPORTEDOPERATION_NOTINWHITELIST = "UnsupportedOperation.NotInWhitelist"
func (c *Client) AcquireClusterAdminRole(request *AcquireClusterAdminRoleRequest) (response *AcquireClusterAdminRoleResponse, err error) {
	if request == nil {
		request = NewAcquireClusterAdminRoleRequest()
	}

	response = NewAcquireClusterAdminRoleResponse()
	err = c.Send(request, response)
	return
}

// AcquireClusterAdminRole
// 通过此接口，可以获取集群的tke:admin的ClusterRole，即管理员角色，可以用于CAM侧高权限的用户，通过CAM策略给予子账户此接口权限，进而可以通过此接口直接获取到kubernetes集群内的管理员角色。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_KUBERNETESCLIENTBUILDERROR = "InternalError.KubernetesClientBuildError"
//	INTERNALERROR_KUBERNETESCREATEOPERATIONERROR = "InternalError.KubernetesCreateOperationError"
//	INTERNALERROR_KUBERNETESGETOPERATIONERROR = "InternalError.KubernetesGetOperationError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
//	UNSUPPORTEDOPERATION_NOTINWHITELIST = "UnsupportedOperation.NotInWhitelist"
func (c *Client) AcquireClusterAdminRoleWithContext(ctx context.Context, request *AcquireClusterAdminRoleRequest) (response *AcquireClusterAdminRoleResponse, err error) {
	if request == nil {
		request = NewAcquireClusterAdminRoleRequest()
	}
	request.SetContext(ctx)

	response = NewAcquireClusterAdminRoleResponse()
	err = c.Send(request, response)
	return
}

func NewAddClusterCIDRRequest() (request *AddClusterCIDRRequest) {
	request = &AddClusterCIDRRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "AddClusterCIDR")

	return
}

func NewAddClusterCIDRResponse() (response *AddClusterCIDRResponse) {
	response = &AddClusterCIDRResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AddClusterCIDR
// 给GR集群增加可用的ClusterCIDR
//
// 可能返回的错误码:
//
//	INTERNALERROR_KUBECLIENTCREATE = "InternalError.KubeClientCreate"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CIDRMASKSIZEOUTOFRANGE = "InvalidParameter.CIDRMaskSizeOutOfRange"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERCLUSTER = "InvalidParameter.CidrConflictWithOtherCluster"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"
//	INVALIDPARAMETER_CIDRCONFLICTWITHVPCCIDR = "InvalidParameter.CidrConflictWithVpcCidr"
//	INVALIDPARAMETER_CIDRCONFLICTWITHVPCGLOBALROUTE = "InvalidParameter.CidrConflictWithVpcGlobalRoute"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) AddClusterCIDR(request *AddClusterCIDRRequest) (response *AddClusterCIDRResponse, err error) {
	if request == nil {
		request = NewAddClusterCIDRRequest()
	}

	response = NewAddClusterCIDRResponse()
	err = c.Send(request, response)
	return
}

// AddClusterCIDR
// 给GR集群增加可用的ClusterCIDR
//
// 可能返回的错误码:
//
//	INTERNALERROR_KUBECLIENTCREATE = "InternalError.KubeClientCreate"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CIDRMASKSIZEOUTOFRANGE = "InvalidParameter.CIDRMaskSizeOutOfRange"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERCLUSTER = "InvalidParameter.CidrConflictWithOtherCluster"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"
//	INVALIDPARAMETER_CIDRCONFLICTWITHVPCCIDR = "InvalidParameter.CidrConflictWithVpcCidr"
//	INVALIDPARAMETER_CIDRCONFLICTWITHVPCGLOBALROUTE = "InvalidParameter.CidrConflictWithVpcGlobalRoute"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) AddClusterCIDRWithContext(ctx context.Context, request *AddClusterCIDRRequest) (response *AddClusterCIDRResponse, err error) {
	if request == nil {
		request = NewAddClusterCIDRRequest()
	}
	request.SetContext(ctx)

	response = NewAddClusterCIDRResponse()
	err = c.Send(request, response)
	return
}

func NewAddExistedInstancesRequest() (request *AddExistedInstancesRequest) {
	request = &AddExistedInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "AddExistedInstances")

	return
}

func NewAddExistedInstancesResponse() (response *AddExistedInstancesResponse) {
	response = &AddExistedInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AddExistedInstances
// 添加已经存在的实例到集群
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) AddExistedInstances(request *AddExistedInstancesRequest) (response *AddExistedInstancesResponse, err error) {
	if request == nil {
		request = NewAddExistedInstancesRequest()
	}

	response = NewAddExistedInstancesResponse()
	err = c.Send(request, response)
	return
}

// AddExistedInstances
// 添加已经存在的实例到集群
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) AddExistedInstancesWithContext(ctx context.Context, request *AddExistedInstancesRequest) (response *AddExistedInstancesResponse, err error) {
	if request == nil {
		request = NewAddExistedInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewAddExistedInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewAddNodeToNodePoolRequest() (request *AddNodeToNodePoolRequest) {
	request = &AddNodeToNodePoolRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "AddNodeToNodePool")

	return
}

func NewAddNodeToNodePoolResponse() (response *AddNodeToNodePoolResponse) {
	response = &AddNodeToNodePoolResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AddNodeToNodePool
// 将集群内节点移入节点池
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) AddNodeToNodePool(request *AddNodeToNodePoolRequest) (response *AddNodeToNodePoolResponse, err error) {
	if request == nil {
		request = NewAddNodeToNodePoolRequest()
	}

	response = NewAddNodeToNodePoolResponse()
	err = c.Send(request, response)
	return
}

// AddNodeToNodePool
// 将集群内节点移入节点池
//
// 可能返回的错误码:
//
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) AddNodeToNodePoolWithContext(ctx context.Context, request *AddNodeToNodePoolRequest) (response *AddNodeToNodePoolResponse, err error) {
	if request == nil {
		request = NewAddNodeToNodePoolRequest()
	}
	request.SetContext(ctx)

	response = NewAddNodeToNodePoolResponse()
	err = c.Send(request, response)
	return
}

func NewAddVpcCniSubnetsRequest() (request *AddVpcCniSubnetsRequest) {
	request = &AddVpcCniSubnetsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "AddVpcCniSubnets")

	return
}

func NewAddVpcCniSubnetsResponse() (response *AddVpcCniSubnetsResponse) {
	response = &AddVpcCniSubnetsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// AddVpcCniSubnets
// 针对VPC-CNI模式的集群，增加集群容器网络可使用的子网
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) AddVpcCniSubnets(request *AddVpcCniSubnetsRequest) (response *AddVpcCniSubnetsResponse, err error) {
	if request == nil {
		request = NewAddVpcCniSubnetsRequest()
	}

	response = NewAddVpcCniSubnetsResponse()
	err = c.Send(request, response)
	return
}

// AddVpcCniSubnets
// 针对VPC-CNI模式的集群，增加集群容器网络可使用的子网
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) AddVpcCniSubnetsWithContext(ctx context.Context, request *AddVpcCniSubnetsRequest) (response *AddVpcCniSubnetsResponse, err error) {
	if request == nil {
		request = NewAddVpcCniSubnetsRequest()
	}
	request.SetContext(ctx)

	response = NewAddVpcCniSubnetsResponse()
	err = c.Send(request, response)
	return
}

func NewCheckInstancesUpgradeAbleRequest() (request *CheckInstancesUpgradeAbleRequest) {
	request = &CheckInstancesUpgradeAbleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CheckInstancesUpgradeAble")

	return
}

func NewCheckInstancesUpgradeAbleResponse() (response *CheckInstancesUpgradeAbleResponse) {
	response = &CheckInstancesUpgradeAbleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CheckInstancesUpgradeAble
// 检查给定节点列表中哪些是可升级的
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_KUBECLIENTCONNECTION = "InternalError.KubeClientConnection"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
func (c *Client) CheckInstancesUpgradeAble(request *CheckInstancesUpgradeAbleRequest) (response *CheckInstancesUpgradeAbleResponse, err error) {
	if request == nil {
		request = NewCheckInstancesUpgradeAbleRequest()
	}

	response = NewCheckInstancesUpgradeAbleResponse()
	err = c.Send(request, response)
	return
}

// CheckInstancesUpgradeAble
// 检查给定节点列表中哪些是可升级的
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_KUBECLIENTCONNECTION = "InternalError.KubeClientConnection"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
func (c *Client) CheckInstancesUpgradeAbleWithContext(ctx context.Context, request *CheckInstancesUpgradeAbleRequest) (response *CheckInstancesUpgradeAbleResponse, err error) {
	if request == nil {
		request = NewCheckInstancesUpgradeAbleRequest()
	}
	request.SetContext(ctx)

	response = NewCheckInstancesUpgradeAbleResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterRequest() (request *CreateClusterRequest) {
	request = &CreateClusterRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateCluster")

	return
}

func NewCreateClusterResponse() (response *CreateClusterResponse) {
	response = &CreateClusterResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateCluster
// 创建集群
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTCOMMON = "InternalError.AccountCommon"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CIDRCONFLICTWITHOTHERCLUSTER = "InternalError.CidrConflictWithOtherCluster"
//	INTERNALERROR_CIDRCONFLICTWITHOTHERROUTE = "InternalError.CidrConflictWithOtherRoute"
//	INTERNALERROR_CIDRCONFLICTWITHVPCCIDR = "InternalError.CidrConflictWithVpcCidr"
//	INTERNALERROR_CIDRCONFLICTWITHVPCGLOBALROUTE = "InternalError.CidrConflictWithVpcGlobalRoute"
//	INTERNALERROR_CIDRINVALI = "InternalError.CidrInvali"
//	INTERNALERROR_CIDRMASKSIZEOUTOFRANGE = "InternalError.CidrMaskSizeOutOfRange"
//	INTERNALERROR_COMPONENTCLINETHTTP = "InternalError.ComponentClinetHttp"
//	INTERNALERROR_CREATEMASTERFAILED = "InternalError.CreateMasterFailed"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNUMBERNOTMATCH = "InternalError.CvmNumberNotMatch"
//	INTERNALERROR_CVMSTATUS = "InternalError.CvmStatus"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_DFWGETUSGCOUNT = "InternalError.DfwGetUSGCount"
//	INTERNALERROR_DFWGETUSGQUOTA = "InternalError.DfwGetUSGQuota"
//	INTERNALERROR_INITMASTERFAILED = "InternalError.InitMasterFailed"
//	INTERNALERROR_INVALIDPRIVATENETWORKCIDR = "InternalError.InvalidPrivateNetworkCidr"
//	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"
//	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"
//	INTERNALERROR_QUOTAUSGLIMIT = "InternalError.QuotaUSGLimit"
//	INTERNALERROR_TASKCREATEFAILED = "InternalError.TaskCreateFailed"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CIDRMASKSIZEOUTOFRANGE = "InvalidParameter.CIDRMaskSizeOutOfRange"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERCLUSTER = "InvalidParameter.CidrConflictWithOtherCluster"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"
//	INVALIDPARAMETER_CIDRCONFLICTWITHVPCCIDR = "InvalidParameter.CidrConflictWithVpcCidr"
//	INVALIDPARAMETER_CIDRCONFLICTWITHVPCGLOBALROUTE = "InvalidParameter.CidrConflictWithVpcGlobalRoute"
//	INVALIDPARAMETER_CIDRINVALID = "InvalidParameter.CidrInvalid"
//	INVALIDPARAMETER_INVALIDPRIVATENETWORKCIDR = "InvalidParameter.InvalidPrivateNetworkCIDR"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) CreateCluster(request *CreateClusterRequest) (response *CreateClusterResponse, err error) {
	if request == nil {
		request = NewCreateClusterRequest()
	}

	response = NewCreateClusterResponse()
	err = c.Send(request, response)
	return
}

// CreateCluster
// 创建集群
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTCOMMON = "InternalError.AccountCommon"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CIDRCONFLICTWITHOTHERCLUSTER = "InternalError.CidrConflictWithOtherCluster"
//	INTERNALERROR_CIDRCONFLICTWITHOTHERROUTE = "InternalError.CidrConflictWithOtherRoute"
//	INTERNALERROR_CIDRCONFLICTWITHVPCCIDR = "InternalError.CidrConflictWithVpcCidr"
//	INTERNALERROR_CIDRCONFLICTWITHVPCGLOBALROUTE = "InternalError.CidrConflictWithVpcGlobalRoute"
//	INTERNALERROR_CIDRINVALI = "InternalError.CidrInvali"
//	INTERNALERROR_CIDRMASKSIZEOUTOFRANGE = "InternalError.CidrMaskSizeOutOfRange"
//	INTERNALERROR_COMPONENTCLINETHTTP = "InternalError.ComponentClinetHttp"
//	INTERNALERROR_CREATEMASTERFAILED = "InternalError.CreateMasterFailed"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNUMBERNOTMATCH = "InternalError.CvmNumberNotMatch"
//	INTERNALERROR_CVMSTATUS = "InternalError.CvmStatus"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_DFWGETUSGCOUNT = "InternalError.DfwGetUSGCount"
//	INTERNALERROR_DFWGETUSGQUOTA = "InternalError.DfwGetUSGQuota"
//	INTERNALERROR_INITMASTERFAILED = "InternalError.InitMasterFailed"
//	INTERNALERROR_INVALIDPRIVATENETWORKCIDR = "InternalError.InvalidPrivateNetworkCidr"
//	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"
//	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"
//	INTERNALERROR_QUOTAUSGLIMIT = "InternalError.QuotaUSGLimit"
//	INTERNALERROR_TASKCREATEFAILED = "InternalError.TaskCreateFailed"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CIDRMASKSIZEOUTOFRANGE = "InvalidParameter.CIDRMaskSizeOutOfRange"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERCLUSTER = "InvalidParameter.CidrConflictWithOtherCluster"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"
//	INVALIDPARAMETER_CIDRCONFLICTWITHVPCCIDR = "InvalidParameter.CidrConflictWithVpcCidr"
//	INVALIDPARAMETER_CIDRCONFLICTWITHVPCGLOBALROUTE = "InvalidParameter.CidrConflictWithVpcGlobalRoute"
//	INVALIDPARAMETER_CIDRINVALID = "InvalidParameter.CidrInvalid"
//	INVALIDPARAMETER_INVALIDPRIVATENETWORKCIDR = "InvalidParameter.InvalidPrivateNetworkCIDR"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) CreateClusterWithContext(ctx context.Context, request *CreateClusterRequest) (response *CreateClusterResponse, err error) {
	if request == nil {
		request = NewCreateClusterRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterAsGroupRequest() (request *CreateClusterAsGroupRequest) {
	request = &CreateClusterAsGroupRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateClusterAsGroup")

	return
}

func NewCreateClusterAsGroupResponse() (response *CreateClusterAsGroupResponse) {
	response = &CreateClusterAsGroupResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateClusterAsGroup
// 为已经存在的集群创建伸缩组
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateClusterAsGroup(request *CreateClusterAsGroupRequest) (response *CreateClusterAsGroupResponse, err error) {
	if request == nil {
		request = NewCreateClusterAsGroupRequest()
	}

	response = NewCreateClusterAsGroupResponse()
	err = c.Send(request, response)
	return
}

// CreateClusterAsGroup
// 为已经存在的集群创建伸缩组
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateClusterAsGroupWithContext(ctx context.Context, request *CreateClusterAsGroupRequest) (response *CreateClusterAsGroupResponse, err error) {
	if request == nil {
		request = NewCreateClusterAsGroupRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterAsGroupResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterEndpointRequest() (request *CreateClusterEndpointRequest) {
	request = &CreateClusterEndpointRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateClusterEndpoint")

	return
}

func NewCreateClusterEndpointResponse() (response *CreateClusterEndpointResponse) {
	response = &CreateClusterEndpointResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateClusterEndpoint
// 创建集群访问端口(独立集群开启内网/外网访问，托管集群支持开启内网访问)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_EMPTYCLUSTERNOTSUPPORT = "InternalError.EmptyClusterNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateClusterEndpoint(request *CreateClusterEndpointRequest) (response *CreateClusterEndpointResponse, err error) {
	if request == nil {
		request = NewCreateClusterEndpointRequest()
	}

	response = NewCreateClusterEndpointResponse()
	err = c.Send(request, response)
	return
}

// CreateClusterEndpoint
// 创建集群访问端口(独立集群开启内网/外网访问，托管集群支持开启内网访问)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_EMPTYCLUSTERNOTSUPPORT = "InternalError.EmptyClusterNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateClusterEndpointWithContext(ctx context.Context, request *CreateClusterEndpointRequest) (response *CreateClusterEndpointResponse, err error) {
	if request == nil {
		request = NewCreateClusterEndpointRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterEndpointResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterEndpointVipRequest() (request *CreateClusterEndpointVipRequest) {
	request = &CreateClusterEndpointVipRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateClusterEndpointVip")

	return
}

func NewCreateClusterEndpointVipResponse() (response *CreateClusterEndpointVipResponse) {
	response = &CreateClusterEndpointVipResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateClusterEndpointVip
// 创建托管集群外网访问端口（老的方式，仅支持托管集群外网端口）
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateClusterEndpointVip(request *CreateClusterEndpointVipRequest) (response *CreateClusterEndpointVipResponse, err error) {
	if request == nil {
		request = NewCreateClusterEndpointVipRequest()
	}

	response = NewCreateClusterEndpointVipResponse()
	err = c.Send(request, response)
	return
}

// CreateClusterEndpointVip
// 创建托管集群外网访问端口（老的方式，仅支持托管集群外网端口）
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateClusterEndpointVipWithContext(ctx context.Context, request *CreateClusterEndpointVipRequest) (response *CreateClusterEndpointVipResponse, err error) {
	if request == nil {
		request = NewCreateClusterEndpointVipRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterEndpointVipResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterInstancesRequest() (request *CreateClusterInstancesRequest) {
	request = &CreateClusterInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateClusterInstances")

	return
}

func NewCreateClusterInstancesResponse() (response *CreateClusterInstancesResponse) {
	response = &CreateClusterInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateClusterInstances
// 扩展(新建)集群节点
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTCOMMON = "InternalError.AccountCommon"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_COMPONENTCLINETHTTP = "InternalError.ComponentClinetHttp"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"
//	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"
//	INTERNALERROR_QUOTAMAXRTLIMIT = "InternalError.QuotaMaxRtLimit"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INTERNALERROR_VPCPEERNOTFOUND = "InternalError.VpcPeerNotFound"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateClusterInstances(request *CreateClusterInstancesRequest) (response *CreateClusterInstancesResponse, err error) {
	if request == nil {
		request = NewCreateClusterInstancesRequest()
	}

	response = NewCreateClusterInstancesResponse()
	err = c.Send(request, response)
	return
}

// CreateClusterInstances
// 扩展(新建)集群节点
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTCOMMON = "InternalError.AccountCommon"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_COMPONENTCLINETHTTP = "InternalError.ComponentClinetHttp"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"
//	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"
//	INTERNALERROR_QUOTAMAXRTLIMIT = "InternalError.QuotaMaxRtLimit"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INTERNALERROR_VPCPEERNOTFOUND = "InternalError.VpcPeerNotFound"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateClusterInstancesWithContext(ctx context.Context, request *CreateClusterInstancesRequest) (response *CreateClusterInstancesResponse, err error) {
	if request == nil {
		request = NewCreateClusterInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterNodePoolRequest() (request *CreateClusterNodePoolRequest) {
	request = &CreateClusterNodePoolRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateClusterNodePool")

	return
}

func NewCreateClusterNodePoolResponse() (response *CreateClusterNodePoolResponse) {
	response = &CreateClusterNodePoolResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateClusterNodePool
// 创建节点池
//
// 可能返回的错误码:
//
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_ASASGNOTEXIST = "ResourceNotFound.AsAsgNotExist"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) CreateClusterNodePool(request *CreateClusterNodePoolRequest) (response *CreateClusterNodePoolResponse, err error) {
	if request == nil {
		request = NewCreateClusterNodePoolRequest()
	}

	response = NewCreateClusterNodePoolResponse()
	err = c.Send(request, response)
	return
}

// CreateClusterNodePool
// 创建节点池
//
// 可能返回的错误码:
//
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_ASASGNOTEXIST = "ResourceNotFound.AsAsgNotExist"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) CreateClusterNodePoolWithContext(ctx context.Context, request *CreateClusterNodePoolRequest) (response *CreateClusterNodePoolResponse, err error) {
	if request == nil {
		request = NewCreateClusterNodePoolRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterNodePoolResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterNodePoolFromExistingAsgRequest() (request *CreateClusterNodePoolFromExistingAsgRequest) {
	request = &CreateClusterNodePoolFromExistingAsgRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateClusterNodePoolFromExistingAsg")

	return
}

func NewCreateClusterNodePoolFromExistingAsgResponse() (response *CreateClusterNodePoolFromExistingAsgResponse) {
	response = &CreateClusterNodePoolFromExistingAsgResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateClusterNodePoolFromExistingAsg
// 从伸缩组创建节点池
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
func (c *Client) CreateClusterNodePoolFromExistingAsg(request *CreateClusterNodePoolFromExistingAsgRequest) (response *CreateClusterNodePoolFromExistingAsgResponse, err error) {
	if request == nil {
		request = NewCreateClusterNodePoolFromExistingAsgRequest()
	}

	response = NewCreateClusterNodePoolFromExistingAsgResponse()
	err = c.Send(request, response)
	return
}

// CreateClusterNodePoolFromExistingAsg
// 从伸缩组创建节点池
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
func (c *Client) CreateClusterNodePoolFromExistingAsgWithContext(ctx context.Context, request *CreateClusterNodePoolFromExistingAsgRequest) (response *CreateClusterNodePoolFromExistingAsgResponse, err error) {
	if request == nil {
		request = NewCreateClusterNodePoolFromExistingAsgRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterNodePoolFromExistingAsgResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterRouteRequest() (request *CreateClusterRouteRequest) {
	request = &CreateClusterRouteRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateClusterRoute")

	return
}

func NewCreateClusterRouteResponse() (response *CreateClusterRouteResponse) {
	response = &CreateClusterRouteResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateClusterRoute
// 创建集群路由
//
// 可能返回的错误码:
//
//	INTERNALERROR_CIDRCONFLICTWITHOTHERROUTE = "InternalError.CidrConflictWithOtherRoute"
//	INTERNALERROR_CIDROUTOFROUTETABLE = "InternalError.CidrOutOfRouteTable"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_GATEWAYALREADYASSOCIATEDCIDR = "InternalError.GatewayAlreadyAssociatedCidr"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_ROUTETABLENOTFOUND = "InternalError.RouteTableNotFound"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"
//	INVALIDPARAMETER_CIDROUTOFROUTETABLE = "InvalidParameter.CidrOutOfRouteTable"
//	INVALIDPARAMETER_GATEWAYALREADYASSOCIATEDCIDR = "InvalidParameter.GatewayAlreadyAssociatedCidr"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_ROUTETABLENOTFOUND = "ResourceNotFound.RouteTableNotFound"
func (c *Client) CreateClusterRoute(request *CreateClusterRouteRequest) (response *CreateClusterRouteResponse, err error) {
	if request == nil {
		request = NewCreateClusterRouteRequest()
	}

	response = NewCreateClusterRouteResponse()
	err = c.Send(request, response)
	return
}

// CreateClusterRoute
// 创建集群路由
//
// 可能返回的错误码:
//
//	INTERNALERROR_CIDRCONFLICTWITHOTHERROUTE = "InternalError.CidrConflictWithOtherRoute"
//	INTERNALERROR_CIDROUTOFROUTETABLE = "InternalError.CidrOutOfRouteTable"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_GATEWAYALREADYASSOCIATEDCIDR = "InternalError.GatewayAlreadyAssociatedCidr"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_ROUTETABLENOTFOUND = "InternalError.RouteTableNotFound"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"
//	INVALIDPARAMETER_CIDROUTOFROUTETABLE = "InvalidParameter.CidrOutOfRouteTable"
//	INVALIDPARAMETER_GATEWAYALREADYASSOCIATEDCIDR = "InvalidParameter.GatewayAlreadyAssociatedCidr"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_ROUTETABLENOTFOUND = "ResourceNotFound.RouteTableNotFound"
func (c *Client) CreateClusterRouteWithContext(ctx context.Context, request *CreateClusterRouteRequest) (response *CreateClusterRouteResponse, err error) {
	if request == nil {
		request = NewCreateClusterRouteRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterRouteResponse()
	err = c.Send(request, response)
	return
}

func NewCreateClusterRouteTableRequest() (request *CreateClusterRouteTableRequest) {
	request = &CreateClusterRouteTableRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateClusterRouteTable")

	return
}

func NewCreateClusterRouteTableResponse() (response *CreateClusterRouteTableResponse) {
	response = &CreateClusterRouteTableResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateClusterRouteTable
// 创建集群路由表
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CIDRCONFLICTWITHOTHERCLUSTER = "InternalError.CidrConflictWithOtherCluster"
//	INTERNALERROR_CIDRCONFLICTWITHOTHERROUTE = "InternalError.CidrConflictWithOtherRoute"
//	INTERNALERROR_CIDRCONFLICTWITHVPCCIDR = "InternalError.CidrConflictWithVpcCidr"
//	INTERNALERROR_CIDRCONFLICTWITHVPCGLOBALROUTE = "InternalError.CidrConflictWithVpcGlobalRoute"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_QUOTAMAXRTLIMIT = "InternalError.QuotaMaxRtLimit"
//	INTERNALERROR_RESOURCEEXISTALREADY = "InternalError.ResourceExistAlready"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) CreateClusterRouteTable(request *CreateClusterRouteTableRequest) (response *CreateClusterRouteTableResponse, err error) {
	if request == nil {
		request = NewCreateClusterRouteTableRequest()
	}

	response = NewCreateClusterRouteTableResponse()
	err = c.Send(request, response)
	return
}

// CreateClusterRouteTable
// 创建集群路由表
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CIDRCONFLICTWITHOTHERCLUSTER = "InternalError.CidrConflictWithOtherCluster"
//	INTERNALERROR_CIDRCONFLICTWITHOTHERROUTE = "InternalError.CidrConflictWithOtherRoute"
//	INTERNALERROR_CIDRCONFLICTWITHVPCCIDR = "InternalError.CidrConflictWithVpcCidr"
//	INTERNALERROR_CIDRCONFLICTWITHVPCGLOBALROUTE = "InternalError.CidrConflictWithVpcGlobalRoute"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_QUOTAMAXRTLIMIT = "InternalError.QuotaMaxRtLimit"
//	INTERNALERROR_RESOURCEEXISTALREADY = "InternalError.ResourceExistAlready"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) CreateClusterRouteTableWithContext(ctx context.Context, request *CreateClusterRouteTableRequest) (response *CreateClusterRouteTableResponse, err error) {
	if request == nil {
		request = NewCreateClusterRouteTableRequest()
	}
	request.SetContext(ctx)

	response = NewCreateClusterRouteTableResponse()
	err = c.Send(request, response)
	return
}

func NewCreateEKSClusterRequest() (request *CreateEKSClusterRequest) {
	request = &CreateEKSClusterRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateEKSCluster")

	return
}

func NewCreateEKSClusterResponse() (response *CreateEKSClusterResponse) {
	response = &CreateEKSClusterResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateEKSCluster
// 创建弹性集群
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateEKSCluster(request *CreateEKSClusterRequest) (response *CreateEKSClusterResponse, err error) {
	if request == nil {
		request = NewCreateEKSClusterRequest()
	}

	response = NewCreateEKSClusterResponse()
	err = c.Send(request, response)
	return
}

// CreateEKSCluster
// 创建弹性集群
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) CreateEKSClusterWithContext(ctx context.Context, request *CreateEKSClusterRequest) (response *CreateEKSClusterResponse, err error) {
	if request == nil {
		request = NewCreateEKSClusterRequest()
	}
	request.SetContext(ctx)

	response = NewCreateEKSClusterResponse()
	err = c.Send(request, response)
	return
}

func NewCreateEKSContainerInstancesRequest() (request *CreateEKSContainerInstancesRequest) {
	request = &CreateEKSContainerInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateEKSContainerInstances")

	return
}

func NewCreateEKSContainerInstancesResponse() (response *CreateEKSContainerInstancesResponse) {
	response = &CreateEKSContainerInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateEKSContainerInstances
// 创建容器实例
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CMDTIMEOUT = "InternalError.CmdTimeout"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) CreateEKSContainerInstances(request *CreateEKSContainerInstancesRequest) (response *CreateEKSContainerInstancesResponse, err error) {
	if request == nil {
		request = NewCreateEKSContainerInstancesRequest()
	}

	response = NewCreateEKSContainerInstancesResponse()
	err = c.Send(request, response)
	return
}

// CreateEKSContainerInstances
// 创建容器实例
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CMDTIMEOUT = "InternalError.CmdTimeout"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) CreateEKSContainerInstancesWithContext(ctx context.Context, request *CreateEKSContainerInstancesRequest) (response *CreateEKSContainerInstancesResponse, err error) {
	if request == nil {
		request = NewCreateEKSContainerInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewCreateEKSContainerInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewCreatePrometheusAlertRuleRequest() (request *CreatePrometheusAlertRuleRequest) {
	request = &CreatePrometheusAlertRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreatePrometheusAlertRule")

	return
}

func NewCreatePrometheusAlertRuleResponse() (response *CreatePrometheusAlertRuleResponse) {
	response = &CreatePrometheusAlertRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreatePrometheusAlertRule
// 创建告警规则
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreatePrometheusAlertRule(request *CreatePrometheusAlertRuleRequest) (response *CreatePrometheusAlertRuleResponse, err error) {
	if request == nil {
		request = NewCreatePrometheusAlertRuleRequest()
	}

	response = NewCreatePrometheusAlertRuleResponse()
	err = c.Send(request, response)
	return
}

// CreatePrometheusAlertRule
// 创建告警规则
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) CreatePrometheusAlertRuleWithContext(ctx context.Context, request *CreatePrometheusAlertRuleRequest) (response *CreatePrometheusAlertRuleResponse, err error) {
	if request == nil {
		request = NewCreatePrometheusAlertRuleRequest()
	}
	request.SetContext(ctx)

	response = NewCreatePrometheusAlertRuleResponse()
	err = c.Send(request, response)
	return
}

func NewCreatePrometheusDashboardRequest() (request *CreatePrometheusDashboardRequest) {
	request = &CreatePrometheusDashboardRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreatePrometheusDashboard")

	return
}

func NewCreatePrometheusDashboardResponse() (response *CreatePrometheusDashboardResponse) {
	response = &CreatePrometheusDashboardResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreatePrometheusDashboard
// 创建grafana监控面板
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) CreatePrometheusDashboard(request *CreatePrometheusDashboardRequest) (response *CreatePrometheusDashboardResponse, err error) {
	if request == nil {
		request = NewCreatePrometheusDashboardRequest()
	}

	response = NewCreatePrometheusDashboardResponse()
	err = c.Send(request, response)
	return
}

// CreatePrometheusDashboard
// 创建grafana监控面板
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) CreatePrometheusDashboardWithContext(ctx context.Context, request *CreatePrometheusDashboardRequest) (response *CreatePrometheusDashboardResponse, err error) {
	if request == nil {
		request = NewCreatePrometheusDashboardRequest()
	}
	request.SetContext(ctx)

	response = NewCreatePrometheusDashboardResponse()
	err = c.Send(request, response)
	return
}

func NewCreatePrometheusTemplateRequest() (request *CreatePrometheusTemplateRequest) {
	request = &CreatePrometheusTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreatePrometheusTemplate")

	return
}

func NewCreatePrometheusTemplateResponse() (response *CreatePrometheusTemplateResponse) {
	response = &CreatePrometheusTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreatePrometheusTemplate
// 创建一个云原生Prometheus模板实例
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) CreatePrometheusTemplate(request *CreatePrometheusTemplateRequest) (response *CreatePrometheusTemplateResponse, err error) {
	if request == nil {
		request = NewCreatePrometheusTemplateRequest()
	}

	response = NewCreatePrometheusTemplateResponse()
	err = c.Send(request, response)
	return
}

// CreatePrometheusTemplate
// 创建一个云原生Prometheus模板实例
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) CreatePrometheusTemplateWithContext(ctx context.Context, request *CreatePrometheusTemplateRequest) (response *CreatePrometheusTemplateResponse, err error) {
	if request == nil {
		request = NewCreatePrometheusTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewCreatePrometheusTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteClusterRequest() (request *DeleteClusterRequest) {
	request = &DeleteClusterRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteCluster")

	return
}

func NewDeleteClusterResponse() (response *DeleteClusterResponse) {
	response = &DeleteClusterResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteCluster
// 删除集群(YUNAPI V3版本)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	OPERATIONDENIED_CLUSTERINDELETIONPROTECTION = "OperationDenied.ClusterInDeletionProtection"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
func (c *Client) DeleteCluster(request *DeleteClusterRequest) (response *DeleteClusterResponse, err error) {
	if request == nil {
		request = NewDeleteClusterRequest()
	}

	response = NewDeleteClusterResponse()
	err = c.Send(request, response)
	return
}

// DeleteCluster
// 删除集群(YUNAPI V3版本)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	OPERATIONDENIED_CLUSTERINDELETIONPROTECTION = "OperationDenied.ClusterInDeletionProtection"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
func (c *Client) DeleteClusterWithContext(ctx context.Context, request *DeleteClusterRequest) (response *DeleteClusterResponse, err error) {
	if request == nil {
		request = NewDeleteClusterRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteClusterResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteClusterAsGroupsRequest() (request *DeleteClusterAsGroupsRequest) {
	request = &DeleteClusterAsGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteClusterAsGroups")

	return
}

func NewDeleteClusterAsGroupsResponse() (response *DeleteClusterAsGroupsResponse) {
	response = &DeleteClusterAsGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteClusterAsGroups
// 删除集群伸缩组
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"
//	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"
//	INTERNALERROR_QUOTAMAXRTLIMIT = "InternalError.QuotaMaxRtLimit"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
func (c *Client) DeleteClusterAsGroups(request *DeleteClusterAsGroupsRequest) (response *DeleteClusterAsGroupsResponse, err error) {
	if request == nil {
		request = NewDeleteClusterAsGroupsRequest()
	}

	response = NewDeleteClusterAsGroupsResponse()
	err = c.Send(request, response)
	return
}

// DeleteClusterAsGroups
// 删除集群伸缩组
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"
//	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"
//	INTERNALERROR_QUOTAMAXRTLIMIT = "InternalError.QuotaMaxRtLimit"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
func (c *Client) DeleteClusterAsGroupsWithContext(ctx context.Context, request *DeleteClusterAsGroupsRequest) (response *DeleteClusterAsGroupsResponse, err error) {
	if request == nil {
		request = NewDeleteClusterAsGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteClusterAsGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteClusterEndpointRequest() (request *DeleteClusterEndpointRequest) {
	request = &DeleteClusterEndpointRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteClusterEndpoint")

	return
}

func NewDeleteClusterEndpointResponse() (response *DeleteClusterEndpointResponse) {
	response = &DeleteClusterEndpointResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteClusterEndpoint
// 删除集群访问端口(独立集群开启内网/外网访问，托管集群支持开启内网访问)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteClusterEndpoint(request *DeleteClusterEndpointRequest) (response *DeleteClusterEndpointResponse, err error) {
	if request == nil {
		request = NewDeleteClusterEndpointRequest()
	}

	response = NewDeleteClusterEndpointResponse()
	err = c.Send(request, response)
	return
}

// DeleteClusterEndpoint
// 删除集群访问端口(独立集群开启内网/外网访问，托管集群支持开启内网访问)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteClusterEndpointWithContext(ctx context.Context, request *DeleteClusterEndpointRequest) (response *DeleteClusterEndpointResponse, err error) {
	if request == nil {
		request = NewDeleteClusterEndpointRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteClusterEndpointResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteClusterEndpointVipRequest() (request *DeleteClusterEndpointVipRequest) {
	request = &DeleteClusterEndpointVipRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteClusterEndpointVip")

	return
}

func NewDeleteClusterEndpointVipResponse() (response *DeleteClusterEndpointVipResponse) {
	response = &DeleteClusterEndpointVipResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteClusterEndpointVip
// 删除托管集群外网访问端口（老的方式，仅支持托管集群外网端口）
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DFWGETUSGCOUNT = "InternalError.DfwGetUSGCount"
//	INTERNALERROR_DFWGETUSGQUOTA = "InternalError.DfwGetUSGQuota"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteClusterEndpointVip(request *DeleteClusterEndpointVipRequest) (response *DeleteClusterEndpointVipResponse, err error) {
	if request == nil {
		request = NewDeleteClusterEndpointVipRequest()
	}

	response = NewDeleteClusterEndpointVipResponse()
	err = c.Send(request, response)
	return
}

// DeleteClusterEndpointVip
// 删除托管集群外网访问端口（老的方式，仅支持托管集群外网端口）
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DFWGETUSGCOUNT = "InternalError.DfwGetUSGCount"
//	INTERNALERROR_DFWGETUSGQUOTA = "InternalError.DfwGetUSGQuota"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteClusterEndpointVipWithContext(ctx context.Context, request *DeleteClusterEndpointVipRequest) (response *DeleteClusterEndpointVipResponse, err error) {
	if request == nil {
		request = NewDeleteClusterEndpointVipRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteClusterEndpointVipResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteClusterInstancesRequest() (request *DeleteClusterInstancesRequest) {
	request = &DeleteClusterInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteClusterInstances")

	return
}

func NewDeleteClusterInstancesResponse() (response *DeleteClusterInstancesResponse) {
	response = &DeleteClusterInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteClusterInstances
// 删除集群中的实例
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) DeleteClusterInstances(request *DeleteClusterInstancesRequest) (response *DeleteClusterInstancesResponse, err error) {
	if request == nil {
		request = NewDeleteClusterInstancesRequest()
	}

	response = NewDeleteClusterInstancesResponse()
	err = c.Send(request, response)
	return
}

// DeleteClusterInstances
// 删除集群中的实例
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
func (c *Client) DeleteClusterInstancesWithContext(ctx context.Context, request *DeleteClusterInstancesRequest) (response *DeleteClusterInstancesResponse, err error) {
	if request == nil {
		request = NewDeleteClusterInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteClusterInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteClusterNodePoolRequest() (request *DeleteClusterNodePoolRequest) {
	request = &DeleteClusterNodePoolRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteClusterNodePool")

	return
}

func NewDeleteClusterNodePoolResponse() (response *DeleteClusterNodePoolResponse) {
	response = &DeleteClusterNodePoolResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteClusterNodePool
// 删除节点池
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DeleteClusterNodePool(request *DeleteClusterNodePoolRequest) (response *DeleteClusterNodePoolResponse, err error) {
	if request == nil {
		request = NewDeleteClusterNodePoolRequest()
	}

	response = NewDeleteClusterNodePoolResponse()
	err = c.Send(request, response)
	return
}

// DeleteClusterNodePool
// 删除节点池
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DeleteClusterNodePoolWithContext(ctx context.Context, request *DeleteClusterNodePoolRequest) (response *DeleteClusterNodePoolResponse, err error) {
	if request == nil {
		request = NewDeleteClusterNodePoolRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteClusterNodePoolResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteClusterRouteRequest() (request *DeleteClusterRouteRequest) {
	request = &DeleteClusterRouteRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteClusterRoute")

	return
}

func NewDeleteClusterRouteResponse() (response *DeleteClusterRouteResponse) {
	response = &DeleteClusterRouteResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteClusterRoute
// 删除集群路由
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_ROUTETABLENOTFOUND = "InternalError.RouteTableNotFound"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DeleteClusterRoute(request *DeleteClusterRouteRequest) (response *DeleteClusterRouteResponse, err error) {
	if request == nil {
		request = NewDeleteClusterRouteRequest()
	}

	response = NewDeleteClusterRouteResponse()
	err = c.Send(request, response)
	return
}

// DeleteClusterRoute
// 删除集群路由
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_ROUTETABLENOTFOUND = "InternalError.RouteTableNotFound"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DeleteClusterRouteWithContext(ctx context.Context, request *DeleteClusterRouteRequest) (response *DeleteClusterRouteResponse, err error) {
	if request == nil {
		request = NewDeleteClusterRouteRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteClusterRouteResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteClusterRouteTableRequest() (request *DeleteClusterRouteTableRequest) {
	request = &DeleteClusterRouteTableRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteClusterRouteTable")

	return
}

func NewDeleteClusterRouteTableResponse() (response *DeleteClusterRouteTableResponse) {
	response = &DeleteClusterRouteTableResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteClusterRouteTable
// 删除集群路由表
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_ROUTETABLENOTEMPTY = "InternalError.RouteTableNotEmpty"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"
func (c *Client) DeleteClusterRouteTable(request *DeleteClusterRouteTableRequest) (response *DeleteClusterRouteTableResponse, err error) {
	if request == nil {
		request = NewDeleteClusterRouteTableRequest()
	}

	response = NewDeleteClusterRouteTableResponse()
	err = c.Send(request, response)
	return
}

// DeleteClusterRouteTable
// 删除集群路由表
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_ROUTETABLENOTEMPTY = "InternalError.RouteTableNotEmpty"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"
func (c *Client) DeleteClusterRouteTableWithContext(ctx context.Context, request *DeleteClusterRouteTableRequest) (response *DeleteClusterRouteTableResponse, err error) {
	if request == nil {
		request = NewDeleteClusterRouteTableRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteClusterRouteTableResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteEKSClusterRequest() (request *DeleteEKSClusterRequest) {
	request = &DeleteEKSClusterRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteEKSCluster")

	return
}

func NewDeleteEKSClusterResponse() (response *DeleteEKSClusterResponse) {
	response = &DeleteEKSClusterResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteEKSCluster
// 删除弹性集群(yunapiv3)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteEKSCluster(request *DeleteEKSClusterRequest) (response *DeleteEKSClusterResponse, err error) {
	if request == nil {
		request = NewDeleteEKSClusterRequest()
	}

	response = NewDeleteEKSClusterResponse()
	err = c.Send(request, response)
	return
}

// DeleteEKSCluster
// 删除弹性集群(yunapiv3)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteEKSClusterWithContext(ctx context.Context, request *DeleteEKSClusterRequest) (response *DeleteEKSClusterResponse, err error) {
	if request == nil {
		request = NewDeleteEKSClusterRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteEKSClusterResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteEKSContainerInstancesRequest() (request *DeleteEKSContainerInstancesRequest) {
	request = &DeleteEKSContainerInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteEKSContainerInstances")

	return
}

func NewDeleteEKSContainerInstancesResponse() (response *DeleteEKSContainerInstancesResponse) {
	response = &DeleteEKSContainerInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteEKSContainerInstances
// 删除容器实例，可批量删除
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CONTAINERNOTFOUND = "InternalError.ContainerNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteEKSContainerInstances(request *DeleteEKSContainerInstancesRequest) (response *DeleteEKSContainerInstancesResponse, err error) {
	if request == nil {
		request = NewDeleteEKSContainerInstancesRequest()
	}

	response = NewDeleteEKSContainerInstancesResponse()
	err = c.Send(request, response)
	return
}

// DeleteEKSContainerInstances
// 删除容器实例，可批量删除
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CONTAINERNOTFOUND = "InternalError.ContainerNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DeleteEKSContainerInstancesWithContext(ctx context.Context, request *DeleteEKSContainerInstancesRequest) (response *DeleteEKSContainerInstancesResponse, err error) {
	if request == nil {
		request = NewDeleteEKSContainerInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDeleteEKSContainerInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDeletePrometheusAlertRuleRequest() (request *DeletePrometheusAlertRuleRequest) {
	request = &DeletePrometheusAlertRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeletePrometheusAlertRule")

	return
}

func NewDeletePrometheusAlertRuleResponse() (response *DeletePrometheusAlertRuleResponse) {
	response = &DeletePrometheusAlertRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeletePrometheusAlertRule
// 删除告警规则
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DeletePrometheusAlertRule(request *DeletePrometheusAlertRuleRequest) (response *DeletePrometheusAlertRuleResponse, err error) {
	if request == nil {
		request = NewDeletePrometheusAlertRuleRequest()
	}

	response = NewDeletePrometheusAlertRuleResponse()
	err = c.Send(request, response)
	return
}

// DeletePrometheusAlertRule
// 删除告警规则
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DeletePrometheusAlertRuleWithContext(ctx context.Context, request *DeletePrometheusAlertRuleRequest) (response *DeletePrometheusAlertRuleResponse, err error) {
	if request == nil {
		request = NewDeletePrometheusAlertRuleRequest()
	}
	request.SetContext(ctx)

	response = NewDeletePrometheusAlertRuleResponse()
	err = c.Send(request, response)
	return
}

func NewDeletePrometheusTemplateRequest() (request *DeletePrometheusTemplateRequest) {
	request = &DeletePrometheusTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeletePrometheusTemplate")

	return
}

func NewDeletePrometheusTemplateResponse() (response *DeletePrometheusTemplateResponse) {
	response = &DeletePrometheusTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeletePrometheusTemplate
// 删除一个云原生Prometheus配置模板
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) DeletePrometheusTemplate(request *DeletePrometheusTemplateRequest) (response *DeletePrometheusTemplateResponse, err error) {
	if request == nil {
		request = NewDeletePrometheusTemplateRequest()
	}

	response = NewDeletePrometheusTemplateResponse()
	err = c.Send(request, response)
	return
}

// DeletePrometheusTemplate
// 删除一个云原生Prometheus配置模板
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) DeletePrometheusTemplateWithContext(ctx context.Context, request *DeletePrometheusTemplateRequest) (response *DeletePrometheusTemplateResponse, err error) {
	if request == nil {
		request = NewDeletePrometheusTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewDeletePrometheusTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewDeletePrometheusTemplateSyncRequest() (request *DeletePrometheusTemplateSyncRequest) {
	request = &DeletePrometheusTemplateSyncRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeletePrometheusTemplateSync")

	return
}

func NewDeletePrometheusTemplateSyncResponse() (response *DeletePrometheusTemplateSyncResponse) {
	response = &DeletePrometheusTemplateSyncResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeletePrometheusTemplateSync
// 取消模板同步，这将会删除目标中该模板所生产的配置
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) DeletePrometheusTemplateSync(request *DeletePrometheusTemplateSyncRequest) (response *DeletePrometheusTemplateSyncResponse, err error) {
	if request == nil {
		request = NewDeletePrometheusTemplateSyncRequest()
	}

	response = NewDeletePrometheusTemplateSyncResponse()
	err = c.Send(request, response)
	return
}

// DeletePrometheusTemplateSync
// 取消模板同步，这将会删除目标中该模板所生产的配置
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) DeletePrometheusTemplateSyncWithContext(ctx context.Context, request *DeletePrometheusTemplateSyncRequest) (response *DeletePrometheusTemplateSyncResponse, err error) {
	if request == nil {
		request = NewDeletePrometheusTemplateSyncRequest()
	}
	request.SetContext(ctx)

	response = NewDeletePrometheusTemplateSyncResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeAvailableClusterVersionRequest() (request *DescribeAvailableClusterVersionRequest) {
	request = &DescribeAvailableClusterVersionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeAvailableClusterVersion")

	return
}

func NewDescribeAvailableClusterVersionResponse() (response *DescribeAvailableClusterVersionResponse) {
	response = &DescribeAvailableClusterVersionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeAvailableClusterVersion
// 获取集群可以升级的所有版本
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
func (c *Client) DescribeAvailableClusterVersion(request *DescribeAvailableClusterVersionRequest) (response *DescribeAvailableClusterVersionResponse, err error) {
	if request == nil {
		request = NewDescribeAvailableClusterVersionRequest()
	}

	response = NewDescribeAvailableClusterVersionResponse()
	err = c.Send(request, response)
	return
}

// DescribeAvailableClusterVersion
// 获取集群可以升级的所有版本
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
func (c *Client) DescribeAvailableClusterVersionWithContext(ctx context.Context, request *DescribeAvailableClusterVersionRequest) (response *DescribeAvailableClusterVersionResponse, err error) {
	if request == nil {
		request = NewDescribeAvailableClusterVersionRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeAvailableClusterVersionResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterAsGroupOptionRequest() (request *DescribeClusterAsGroupOptionRequest) {
	request = &DescribeClusterAsGroupOptionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterAsGroupOption")

	return
}

func NewDescribeClusterAsGroupOptionResponse() (response *DescribeClusterAsGroupOptionResponse) {
	response = &DescribeClusterAsGroupOptionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterAsGroupOption
// 集群弹性伸缩配置
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeClusterAsGroupOption(request *DescribeClusterAsGroupOptionRequest) (response *DescribeClusterAsGroupOptionResponse, err error) {
	if request == nil {
		request = NewDescribeClusterAsGroupOptionRequest()
	}

	response = NewDescribeClusterAsGroupOptionResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterAsGroupOption
// 集群弹性伸缩配置
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeClusterAsGroupOptionWithContext(ctx context.Context, request *DescribeClusterAsGroupOptionRequest) (response *DescribeClusterAsGroupOptionResponse, err error) {
	if request == nil {
		request = NewDescribeClusterAsGroupOptionRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterAsGroupOptionResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterAsGroupsRequest() (request *DescribeClusterAsGroupsRequest) {
	request = &DescribeClusterAsGroupsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterAsGroups")

	return
}

func NewDescribeClusterAsGroupsResponse() (response *DescribeClusterAsGroupsResponse) {
	response = &DescribeClusterAsGroupsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterAsGroups
// 集群关联的伸缩组列表
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PODNOTFOUND = "InternalError.PodNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INTERNALERROR_VPCPEERNOTFOUND = "InternalError.VpcPeerNotFound"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
func (c *Client) DescribeClusterAsGroups(request *DescribeClusterAsGroupsRequest) (response *DescribeClusterAsGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeClusterAsGroupsRequest()
	}

	response = NewDescribeClusterAsGroupsResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterAsGroups
// 集群关联的伸缩组列表
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PODNOTFOUND = "InternalError.PodNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INTERNALERROR_VPCPEERNOTFOUND = "InternalError.VpcPeerNotFound"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
func (c *Client) DescribeClusterAsGroupsWithContext(ctx context.Context, request *DescribeClusterAsGroupsRequest) (response *DescribeClusterAsGroupsResponse, err error) {
	if request == nil {
		request = NewDescribeClusterAsGroupsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterAsGroupsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterAuthenticationOptionsRequest() (request *DescribeClusterAuthenticationOptionsRequest) {
	request = &DescribeClusterAuthenticationOptionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterAuthenticationOptions")

	return
}

func NewDescribeClusterAuthenticationOptionsResponse() (response *DescribeClusterAuthenticationOptionsResponse) {
	response = &DescribeClusterAuthenticationOptionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterAuthenticationOptions
// 查看集群认证配置
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) DescribeClusterAuthenticationOptions(request *DescribeClusterAuthenticationOptionsRequest) (response *DescribeClusterAuthenticationOptionsResponse, err error) {
	if request == nil {
		request = NewDescribeClusterAuthenticationOptionsRequest()
	}

	response = NewDescribeClusterAuthenticationOptionsResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterAuthenticationOptions
// 查看集群认证配置
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) DescribeClusterAuthenticationOptionsWithContext(ctx context.Context, request *DescribeClusterAuthenticationOptionsRequest) (response *DescribeClusterAuthenticationOptionsResponse, err error) {
	if request == nil {
		request = NewDescribeClusterAuthenticationOptionsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterAuthenticationOptionsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterCommonNamesRequest() (request *DescribeClusterCommonNamesRequest) {
	request = &DescribeClusterCommonNamesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterCommonNames")

	return
}

func NewDescribeClusterCommonNamesResponse() (response *DescribeClusterCommonNamesResponse) {
	response = &DescribeClusterCommonNamesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterCommonNames
// 获取指定子账户在RBAC授权模式中对应kube-apiserver客户端证书的CommonName字段，如果没有客户端证书，将会签发一个，此接口有最大传入子账户数量上限，当前为50
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_WHITELISTUNEXPECTEDERROR = "InternalError.WhitelistUnexpectedError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
//	UNSUPPORTEDOPERATION_NOTINWHITELIST = "UnsupportedOperation.NotInWhitelist"
func (c *Client) DescribeClusterCommonNames(request *DescribeClusterCommonNamesRequest) (response *DescribeClusterCommonNamesResponse, err error) {
	if request == nil {
		request = NewDescribeClusterCommonNamesRequest()
	}

	response = NewDescribeClusterCommonNamesResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterCommonNames
// 获取指定子账户在RBAC授权模式中对应kube-apiserver客户端证书的CommonName字段，如果没有客户端证书，将会签发一个，此接口有最大传入子账户数量上限，当前为50
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_WHITELISTUNEXPECTEDERROR = "InternalError.WhitelistUnexpectedError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
//	UNSUPPORTEDOPERATION_NOTINWHITELIST = "UnsupportedOperation.NotInWhitelist"
func (c *Client) DescribeClusterCommonNamesWithContext(ctx context.Context, request *DescribeClusterCommonNamesRequest) (response *DescribeClusterCommonNamesResponse, err error) {
	if request == nil {
		request = NewDescribeClusterCommonNamesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterCommonNamesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterControllersRequest() (request *DescribeClusterControllersRequest) {
	request = &DescribeClusterControllersRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterControllers")

	return
}

func NewDescribeClusterControllersResponse() (response *DescribeClusterControllersResponse) {
	response = &DescribeClusterControllersResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterControllers
// 用于查询Kubernetes的各个原生控制器是否开启
//
// 可能返回的错误码:
//
//	INTERNALERROR_KUBECLIENTCREATE = "InternalError.KubeClientCreate"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DescribeClusterControllers(request *DescribeClusterControllersRequest) (response *DescribeClusterControllersResponse, err error) {
	if request == nil {
		request = NewDescribeClusterControllersRequest()
	}

	response = NewDescribeClusterControllersResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterControllers
// 用于查询Kubernetes的各个原生控制器是否开启
//
// 可能返回的错误码:
//
//	INTERNALERROR_KUBECLIENTCREATE = "InternalError.KubeClientCreate"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DescribeClusterControllersWithContext(ctx context.Context, request *DescribeClusterControllersRequest) (response *DescribeClusterControllersResponse, err error) {
	if request == nil {
		request = NewDescribeClusterControllersRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterControllersResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterEndpointStatusRequest() (request *DescribeClusterEndpointStatusRequest) {
	request = &DescribeClusterEndpointStatusRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterEndpointStatus")

	return
}

func NewDescribeClusterEndpointStatusResponse() (response *DescribeClusterEndpointStatusResponse) {
	response = &DescribeClusterEndpointStatusResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterEndpointStatus
// 查询集群访问端口状态(独立集群开启内网/外网访问，托管集群支持开启内网访问)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_KUBECLIENTCONNECTION = "InternalError.KubeClientConnection"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_KUBERNETESINTERNAL = "InternalError.KubernetesInternal"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeClusterEndpointStatus(request *DescribeClusterEndpointStatusRequest) (response *DescribeClusterEndpointStatusResponse, err error) {
	if request == nil {
		request = NewDescribeClusterEndpointStatusRequest()
	}

	response = NewDescribeClusterEndpointStatusResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterEndpointStatus
// 查询集群访问端口状态(独立集群开启内网/外网访问，托管集群支持开启内网访问)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_KUBECLIENTCONNECTION = "InternalError.KubeClientConnection"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_KUBERNETESINTERNAL = "InternalError.KubernetesInternal"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeClusterEndpointStatusWithContext(ctx context.Context, request *DescribeClusterEndpointStatusRequest) (response *DescribeClusterEndpointStatusResponse, err error) {
	if request == nil {
		request = NewDescribeClusterEndpointStatusRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterEndpointStatusResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterEndpointVipStatusRequest() (request *DescribeClusterEndpointVipStatusRequest) {
	request = &DescribeClusterEndpointVipStatusRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterEndpointVipStatus")

	return
}

func NewDescribeClusterEndpointVipStatusResponse() (response *DescribeClusterEndpointVipStatusResponse) {
	response = &DescribeClusterEndpointVipStatusResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterEndpointVipStatus
// 查询集群开启端口流程状态(仅支持托管集群外网端口)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_DFWGETUSGCOUNT = "InternalError.DfwGetUSGCount"
//	INTERNALERROR_DFWGETUSGQUOTA = "InternalError.DfwGetUSGQuota"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeClusterEndpointVipStatus(request *DescribeClusterEndpointVipStatusRequest) (response *DescribeClusterEndpointVipStatusResponse, err error) {
	if request == nil {
		request = NewDescribeClusterEndpointVipStatusRequest()
	}

	response = NewDescribeClusterEndpointVipStatusResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterEndpointVipStatus
// 查询集群开启端口流程状态(仅支持托管集群外网端口)
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_DFWGETUSGCOUNT = "InternalError.DfwGetUSGCount"
//	INTERNALERROR_DFWGETUSGQUOTA = "InternalError.DfwGetUSGQuota"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeClusterEndpointVipStatusWithContext(ctx context.Context, request *DescribeClusterEndpointVipStatusRequest) (response *DescribeClusterEndpointVipStatusResponse, err error) {
	if request == nil {
		request = NewDescribeClusterEndpointVipStatusRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterEndpointVipStatusResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterInstancesRequest() (request *DescribeClusterInstancesRequest) {
	request = &DescribeClusterInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterInstances")

	return
}

func NewDescribeClusterInstancesResponse() (response *DescribeClusterInstancesResponse) {
	response = &DescribeClusterInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterInstances
//
//	查询集群下节点实例信息
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_INITMASTERFAILED = "InternalError.InitMasterFailed"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) DescribeClusterInstances(request *DescribeClusterInstancesRequest) (response *DescribeClusterInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeClusterInstancesRequest()
	}

	response = NewDescribeClusterInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterInstances
//
//	查询集群下节点实例信息
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_INITMASTERFAILED = "InternalError.InitMasterFailed"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) DescribeClusterInstancesWithContext(ctx context.Context, request *DescribeClusterInstancesRequest) (response *DescribeClusterInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeClusterInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterKubeconfigRequest() (request *DescribeClusterKubeconfigRequest) {
	request = &DescribeClusterKubeconfigRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterKubeconfig")

	return
}

func NewDescribeClusterKubeconfigResponse() (response *DescribeClusterKubeconfigResponse) {
	response = &DescribeClusterKubeconfigResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterKubeconfig
// 获取集群的kubeconfig文件，不同子账户获取自己的kubeconfig文件，该文件中有每个子账户自己的kube-apiserver的客户端证书，默认首次调此接口时候创建客户端证书，时效20年，未授予任何权限，如果是集群所有者或者主账户，则默认是cluster-admin权限。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_KUBECLIENTCONNECTION = "InternalError.KubeClientConnection"
//	INTERNALERROR_KUBERNETESCLIENTBUILDERROR = "InternalError.KubernetesClientBuildError"
//	INTERNALERROR_KUBERNETESCREATEOPERATIONERROR = "InternalError.KubernetesCreateOperationError"
//	INTERNALERROR_KUBERNETESDELETEOPERATIONERROR = "InternalError.KubernetesDeleteOperationError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_WHITELISTUNEXPECTEDERROR = "InternalError.WhitelistUnexpectedError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	RESOURCENOTFOUND_KUBERNETESRESOURCENOTFOUND = "ResourceNotFound.KubernetesResourceNotFound"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
func (c *Client) DescribeClusterKubeconfig(request *DescribeClusterKubeconfigRequest) (response *DescribeClusterKubeconfigResponse, err error) {
	if request == nil {
		request = NewDescribeClusterKubeconfigRequest()
	}

	response = NewDescribeClusterKubeconfigResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterKubeconfig
// 获取集群的kubeconfig文件，不同子账户获取自己的kubeconfig文件，该文件中有每个子账户自己的kube-apiserver的客户端证书，默认首次调此接口时候创建客户端证书，时效20年，未授予任何权限，如果是集群所有者或者主账户，则默认是cluster-admin权限。
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_KUBECLIENTCONNECTION = "InternalError.KubeClientConnection"
//	INTERNALERROR_KUBERNETESCLIENTBUILDERROR = "InternalError.KubernetesClientBuildError"
//	INTERNALERROR_KUBERNETESCREATEOPERATIONERROR = "InternalError.KubernetesCreateOperationError"
//	INTERNALERROR_KUBERNETESDELETEOPERATIONERROR = "InternalError.KubernetesDeleteOperationError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_WHITELISTUNEXPECTEDERROR = "InternalError.WhitelistUnexpectedError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	RESOURCENOTFOUND_KUBERNETESRESOURCENOTFOUND = "ResourceNotFound.KubernetesResourceNotFound"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
func (c *Client) DescribeClusterKubeconfigWithContext(ctx context.Context, request *DescribeClusterKubeconfigRequest) (response *DescribeClusterKubeconfigResponse, err error) {
	if request == nil {
		request = NewDescribeClusterKubeconfigRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterKubeconfigResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterNodePoolDetailRequest() (request *DescribeClusterNodePoolDetailRequest) {
	request = &DescribeClusterNodePoolDetailRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterNodePoolDetail")

	return
}

func NewDescribeClusterNodePoolDetailResponse() (response *DescribeClusterNodePoolDetailResponse) {
	response = &DescribeClusterNodePoolDetailResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterNodePoolDetail
// 查询节点池详情
//
// 可能返回的错误码:
//
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
func (c *Client) DescribeClusterNodePoolDetail(request *DescribeClusterNodePoolDetailRequest) (response *DescribeClusterNodePoolDetailResponse, err error) {
	if request == nil {
		request = NewDescribeClusterNodePoolDetailRequest()
	}

	response = NewDescribeClusterNodePoolDetailResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterNodePoolDetail
// 查询节点池详情
//
// 可能返回的错误码:
//
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
func (c *Client) DescribeClusterNodePoolDetailWithContext(ctx context.Context, request *DescribeClusterNodePoolDetailRequest) (response *DescribeClusterNodePoolDetailResponse, err error) {
	if request == nil {
		request = NewDescribeClusterNodePoolDetailRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterNodePoolDetailResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterNodePoolsRequest() (request *DescribeClusterNodePoolsRequest) {
	request = &DescribeClusterNodePoolsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterNodePools")

	return
}

func NewDescribeClusterNodePoolsResponse() (response *DescribeClusterNodePoolsResponse) {
	response = &DescribeClusterNodePoolsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterNodePools
// 查询节点池列表
//
// 可能返回的错误码:
//
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
func (c *Client) DescribeClusterNodePools(request *DescribeClusterNodePoolsRequest) (response *DescribeClusterNodePoolsResponse, err error) {
	if request == nil {
		request = NewDescribeClusterNodePoolsRequest()
	}

	response = NewDescribeClusterNodePoolsResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterNodePools
// 查询节点池列表
//
// 可能返回的错误码:
//
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
func (c *Client) DescribeClusterNodePoolsWithContext(ctx context.Context, request *DescribeClusterNodePoolsRequest) (response *DescribeClusterNodePoolsResponse, err error) {
	if request == nil {
		request = NewDescribeClusterNodePoolsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterNodePoolsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterRouteTablesRequest() (request *DescribeClusterRouteTablesRequest) {
	request = &DescribeClusterRouteTablesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterRouteTables")

	return
}

func NewDescribeClusterRouteTablesResponse() (response *DescribeClusterRouteTablesResponse) {
	response = &DescribeClusterRouteTablesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterRouteTables
// 查询集群路由表
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
func (c *Client) DescribeClusterRouteTables(request *DescribeClusterRouteTablesRequest) (response *DescribeClusterRouteTablesResponse, err error) {
	if request == nil {
		request = NewDescribeClusterRouteTablesRequest()
	}

	response = NewDescribeClusterRouteTablesResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterRouteTables
// 查询集群路由表
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
func (c *Client) DescribeClusterRouteTablesWithContext(ctx context.Context, request *DescribeClusterRouteTablesRequest) (response *DescribeClusterRouteTablesResponse, err error) {
	if request == nil {
		request = NewDescribeClusterRouteTablesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterRouteTablesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterRoutesRequest() (request *DescribeClusterRoutesRequest) {
	request = &DescribeClusterRoutesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterRoutes")

	return
}

func NewDescribeClusterRoutesResponse() (response *DescribeClusterRoutesResponse) {
	response = &DescribeClusterRoutesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterRoutes
// 查询集群路由
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribeClusterRoutes(request *DescribeClusterRoutesRequest) (response *DescribeClusterRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeClusterRoutesRequest()
	}

	response = NewDescribeClusterRoutesResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterRoutes
// 查询集群路由
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribeClusterRoutesWithContext(ctx context.Context, request *DescribeClusterRoutesRequest) (response *DescribeClusterRoutesResponse, err error) {
	if request == nil {
		request = NewDescribeClusterRoutesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterRoutesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClusterSecurityRequest() (request *DescribeClusterSecurityRequest) {
	request = &DescribeClusterSecurityRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusterSecurity")

	return
}

func NewDescribeClusterSecurityResponse() (response *DescribeClusterSecurityResponse) {
	response = &DescribeClusterSecurityResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusterSecurity
// 集群的密钥信息
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_COMPONENTCLIENTHTTP = "InternalError.ComponentClientHttp"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_LBCOMMON = "InternalError.LbCommon"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_CIDRINVALID = "InvalidParameter.CidrInvalid"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	RESOURCENOTFOUND_KUBERESOURCENOTFOUND = "ResourceNotFound.KubeResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeClusterSecurity(request *DescribeClusterSecurityRequest) (response *DescribeClusterSecurityResponse, err error) {
	if request == nil {
		request = NewDescribeClusterSecurityRequest()
	}

	response = NewDescribeClusterSecurityResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusterSecurity
// 集群的密钥信息
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_COMPONENTCLIENTHTTP = "InternalError.ComponentClientHttp"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_LBCOMMON = "InternalError.LbCommon"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_CIDRINVALID = "InvalidParameter.CidrInvalid"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	RESOURCENOTFOUND_KUBERESOURCENOTFOUND = "ResourceNotFound.KubeResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeClusterSecurityWithContext(ctx context.Context, request *DescribeClusterSecurityRequest) (response *DescribeClusterSecurityResponse, err error) {
	if request == nil {
		request = NewDescribeClusterSecurityRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClusterSecurityResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeClustersRequest() (request *DescribeClustersRequest) {
	request = &DescribeClustersRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeClusters")

	return
}

func NewDescribeClustersResponse() (response *DescribeClustersResponse) {
	response = &DescribeClustersResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeClusters
// 查询集群列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"
//	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
func (c *Client) DescribeClusters(request *DescribeClustersRequest) (response *DescribeClustersResponse, err error) {
	if request == nil {
		request = NewDescribeClustersRequest()
	}

	response = NewDescribeClustersResponse()
	err = c.Send(request, response)
	return
}

// DescribeClusters
// 查询集群列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"
//	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"
//	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
func (c *Client) DescribeClustersWithContext(ctx context.Context, request *DescribeClustersRequest) (response *DescribeClustersResponse, err error) {
	if request == nil {
		request = NewDescribeClustersRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeClustersResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeEKSClusterCredentialRequest() (request *DescribeEKSClusterCredentialRequest) {
	request = &DescribeEKSClusterCredentialRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeEKSClusterCredential")

	return
}

func NewDescribeEKSClusterCredentialResponse() (response *DescribeEKSClusterCredentialResponse) {
	response = &DescribeEKSClusterCredentialResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeEKSClusterCredential
// 获取弹性容器集群的接入认证信息
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeEKSClusterCredential(request *DescribeEKSClusterCredentialRequest) (response *DescribeEKSClusterCredentialResponse, err error) {
	if request == nil {
		request = NewDescribeEKSClusterCredentialRequest()
	}

	response = NewDescribeEKSClusterCredentialResponse()
	err = c.Send(request, response)
	return
}

// DescribeEKSClusterCredential
// 获取弹性容器集群的接入认证信息
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeEKSClusterCredentialWithContext(ctx context.Context, request *DescribeEKSClusterCredentialRequest) (response *DescribeEKSClusterCredentialResponse, err error) {
	if request == nil {
		request = NewDescribeEKSClusterCredentialRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeEKSClusterCredentialResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeEKSClustersRequest() (request *DescribeEKSClustersRequest) {
	request = &DescribeEKSClustersRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeEKSClusters")

	return
}

func NewDescribeEKSClustersResponse() (response *DescribeEKSClustersResponse) {
	response = &DescribeEKSClustersResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeEKSClusters
// 查询弹性集群列表
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeEKSClusters(request *DescribeEKSClustersRequest) (response *DescribeEKSClustersResponse, err error) {
	if request == nil {
		request = NewDescribeEKSClustersRequest()
	}

	response = NewDescribeEKSClustersResponse()
	err = c.Send(request, response)
	return
}

// DescribeEKSClusters
// 查询弹性集群列表
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeEKSClustersWithContext(ctx context.Context, request *DescribeEKSClustersRequest) (response *DescribeEKSClustersResponse, err error) {
	if request == nil {
		request = NewDescribeEKSClustersRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeEKSClustersResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeEKSContainerInstanceEventRequest() (request *DescribeEKSContainerInstanceEventRequest) {
	request = &DescribeEKSContainerInstanceEventRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeEKSContainerInstanceEvent")

	return
}

func NewDescribeEKSContainerInstanceEventResponse() (response *DescribeEKSContainerInstanceEventResponse) {
	response = &DescribeEKSContainerInstanceEventResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeEKSContainerInstanceEvent
// 查询容器实例的事件
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	RESOURCESSOLDOUT = "ResourcesSoldOut"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeEKSContainerInstanceEvent(request *DescribeEKSContainerInstanceEventRequest) (response *DescribeEKSContainerInstanceEventResponse, err error) {
	if request == nil {
		request = NewDescribeEKSContainerInstanceEventRequest()
	}

	response = NewDescribeEKSContainerInstanceEventResponse()
	err = c.Send(request, response)
	return
}

// DescribeEKSContainerInstanceEvent
// 查询容器实例的事件
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	RESOURCEINSUFFICIENT = "ResourceInsufficient"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	RESOURCESSOLDOUT = "ResourcesSoldOut"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeEKSContainerInstanceEventWithContext(ctx context.Context, request *DescribeEKSContainerInstanceEventRequest) (response *DescribeEKSContainerInstanceEventResponse, err error) {
	if request == nil {
		request = NewDescribeEKSContainerInstanceEventRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeEKSContainerInstanceEventResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeEKSContainerInstanceRegionsRequest() (request *DescribeEKSContainerInstanceRegionsRequest) {
	request = &DescribeEKSContainerInstanceRegionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeEKSContainerInstanceRegions")

	return
}

func NewDescribeEKSContainerInstanceRegionsResponse() (response *DescribeEKSContainerInstanceRegionsResponse) {
	response = &DescribeEKSContainerInstanceRegionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeEKSContainerInstanceRegions
// 查询容器实例支持的地域
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribeEKSContainerInstanceRegions(request *DescribeEKSContainerInstanceRegionsRequest) (response *DescribeEKSContainerInstanceRegionsResponse, err error) {
	if request == nil {
		request = NewDescribeEKSContainerInstanceRegionsRequest()
	}

	response = NewDescribeEKSContainerInstanceRegionsResponse()
	err = c.Send(request, response)
	return
}

// DescribeEKSContainerInstanceRegions
// 查询容器实例支持的地域
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribeEKSContainerInstanceRegionsWithContext(ctx context.Context, request *DescribeEKSContainerInstanceRegionsRequest) (response *DescribeEKSContainerInstanceRegionsResponse, err error) {
	if request == nil {
		request = NewDescribeEKSContainerInstanceRegionsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeEKSContainerInstanceRegionsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeEKSContainerInstancesRequest() (request *DescribeEKSContainerInstancesRequest) {
	request = &DescribeEKSContainerInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeEKSContainerInstances")

	return
}

func NewDescribeEKSContainerInstancesResponse() (response *DescribeEKSContainerInstancesResponse) {
	response = &DescribeEKSContainerInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeEKSContainerInstances
// 查询容器实例
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	FAILEDOPERATION_RBACFORBIDDEN = "FailedOperation.RBACForbidden"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeEKSContainerInstances(request *DescribeEKSContainerInstancesRequest) (response *DescribeEKSContainerInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeEKSContainerInstancesRequest()
	}

	response = NewDescribeEKSContainerInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeEKSContainerInstances
// 查询容器实例
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	FAILEDOPERATION_RBACFORBIDDEN = "FailedOperation.RBACForbidden"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	RESOURCENOTFOUND = "ResourceNotFound"
func (c *Client) DescribeEKSContainerInstancesWithContext(ctx context.Context, request *DescribeEKSContainerInstancesRequest) (response *DescribeEKSContainerInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeEKSContainerInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeEKSContainerInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeEksContainerInstanceLogRequest() (request *DescribeEksContainerInstanceLogRequest) {
	request = &DescribeEksContainerInstanceLogRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeEksContainerInstanceLog")

	return
}

func NewDescribeEksContainerInstanceLogResponse() (response *DescribeEksContainerInstanceLogResponse) {
	response = &DescribeEksContainerInstanceLogResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeEksContainerInstanceLog
// 查询容器实例中容器日志
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CONTAINERNOTFOUND = "InternalError.ContainerNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE_EKSCONTAINERSTATUS = "ResourceUnavailable.EksContainerStatus"
func (c *Client) DescribeEksContainerInstanceLog(request *DescribeEksContainerInstanceLogRequest) (response *DescribeEksContainerInstanceLogResponse, err error) {
	if request == nil {
		request = NewDescribeEksContainerInstanceLogRequest()
	}

	response = NewDescribeEksContainerInstanceLogResponse()
	err = c.Send(request, response)
	return
}

// DescribeEksContainerInstanceLog
// 查询容器实例中容器日志
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CONTAINERNOTFOUND = "InternalError.ContainerNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE_EKSCONTAINERSTATUS = "ResourceUnavailable.EksContainerStatus"
func (c *Client) DescribeEksContainerInstanceLogWithContext(ctx context.Context, request *DescribeEksContainerInstanceLogRequest) (response *DescribeEksContainerInstanceLogResponse, err error) {
	if request == nil {
		request = NewDescribeEksContainerInstanceLogRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeEksContainerInstanceLogResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeEnableVpcCniProgressRequest() (request *DescribeEnableVpcCniProgressRequest) {
	request = &DescribeEnableVpcCniProgressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeEnableVpcCniProgress")

	return
}

func NewDescribeEnableVpcCniProgressResponse() (response *DescribeEnableVpcCniProgressResponse) {
	response = &DescribeEnableVpcCniProgressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeEnableVpcCniProgress
// 本接口用于查询开启vpc-cni模式的任务进度
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
func (c *Client) DescribeEnableVpcCniProgress(request *DescribeEnableVpcCniProgressRequest) (response *DescribeEnableVpcCniProgressResponse, err error) {
	if request == nil {
		request = NewDescribeEnableVpcCniProgressRequest()
	}

	response = NewDescribeEnableVpcCniProgressResponse()
	err = c.Send(request, response)
	return
}

// DescribeEnableVpcCniProgress
// 本接口用于查询开启vpc-cni模式的任务进度
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
func (c *Client) DescribeEnableVpcCniProgressWithContext(ctx context.Context, request *DescribeEnableVpcCniProgressRequest) (response *DescribeEnableVpcCniProgressResponse, err error) {
	if request == nil {
		request = NewDescribeEnableVpcCniProgressRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeEnableVpcCniProgressResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeExistedInstancesRequest() (request *DescribeExistedInstancesRequest) {
	request = &DescribeExistedInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeExistedInstances")

	return
}

func NewDescribeExistedInstancesResponse() (response *DescribeExistedInstancesResponse) {
	response = &DescribeExistedInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeExistedInstances
// 查询已经存在的节点，判断是否可以加入集群
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CREATEMASTERFAILED = "InternalError.CreateMasterFailed"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_INITMASTERFAILED = "InternalError.InitMasterFailed"
//	INTERNALERROR_INVALIDPRIVATENETWORKCIDR = "InternalError.InvalidPrivateNetworkCidr"
//	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeExistedInstances(request *DescribeExistedInstancesRequest) (response *DescribeExistedInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeExistedInstancesRequest()
	}

	response = NewDescribeExistedInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribeExistedInstances
// 查询已经存在的节点，判断是否可以加入集群
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CREATEMASTERFAILED = "InternalError.CreateMasterFailed"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_INITMASTERFAILED = "InternalError.InitMasterFailed"
//	INTERNALERROR_INVALIDPRIVATENETWORKCIDR = "InternalError.InvalidPrivateNetworkCidr"
//	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeExistedInstancesWithContext(ctx context.Context, request *DescribeExistedInstancesRequest) (response *DescribeExistedInstancesResponse, err error) {
	if request == nil {
		request = NewDescribeExistedInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeExistedInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeExternalClusterSpecRequest() (request *DescribeExternalClusterSpecRequest) {
	request = &DescribeExternalClusterSpecRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeExternalClusterSpec")

	return
}

func NewDescribeExternalClusterSpecResponse() (response *DescribeExternalClusterSpecResponse) {
	response = &DescribeExternalClusterSpecResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeExternalClusterSpec
// 获取导入第三方集群YAML定义
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_KUBERNETESCLIENTBUILDERROR = "InternalError.KubernetesClientBuildError"
//	INTERNALERROR_KUBERNETESCREATEOPERATIONERROR = "InternalError.KubernetesCreateOperationError"
//	INTERNALERROR_KUBERNETESDELETEOPERATIONERROR = "InternalError.KubernetesDeleteOperationError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_WHITELISTUNEXPECTEDERROR = "InternalError.WhitelistUnexpectedError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	RESOURCENOTFOUND_KUBERNETESRESOURCENOTFOUND = "ResourceNotFound.KubernetesResourceNotFound"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
func (c *Client) DescribeExternalClusterSpec(request *DescribeExternalClusterSpecRequest) (response *DescribeExternalClusterSpecResponse, err error) {
	if request == nil {
		request = NewDescribeExternalClusterSpecRequest()
	}

	response = NewDescribeExternalClusterSpecResponse()
	err = c.Send(request, response)
	return
}

// DescribeExternalClusterSpec
// 获取导入第三方集群YAML定义
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_KUBERNETESCLIENTBUILDERROR = "InternalError.KubernetesClientBuildError"
//	INTERNALERROR_KUBERNETESCREATEOPERATIONERROR = "InternalError.KubernetesCreateOperationError"
//	INTERNALERROR_KUBERNETESDELETEOPERATIONERROR = "InternalError.KubernetesDeleteOperationError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_WHITELISTUNEXPECTEDERROR = "InternalError.WhitelistUnexpectedError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	RESOURCENOTFOUND_KUBERNETESRESOURCENOTFOUND = "ResourceNotFound.KubernetesResourceNotFound"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
//	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"
func (c *Client) DescribeExternalClusterSpecWithContext(ctx context.Context, request *DescribeExternalClusterSpecRequest) (response *DescribeExternalClusterSpecResponse, err error) {
	if request == nil {
		request = NewDescribeExternalClusterSpecRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeExternalClusterSpecResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeImagesRequest() (request *DescribeImagesRequest) {
	request = &DescribeImagesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeImages")

	return
}

func NewDescribeImagesResponse() (response *DescribeImagesResponse) {
	response = &DescribeImagesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeImages
// 获取镜像信息
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeImages(request *DescribeImagesRequest) (response *DescribeImagesResponse, err error) {
	if request == nil {
		request = NewDescribeImagesRequest()
	}

	response = NewDescribeImagesResponse()
	err = c.Send(request, response)
	return
}

// DescribeImages
// 获取镜像信息
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeImagesWithContext(ctx context.Context, request *DescribeImagesRequest) (response *DescribeImagesResponse, err error) {
	if request == nil {
		request = NewDescribeImagesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeImagesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusAgentInstancesRequest() (request *DescribePrometheusAgentInstancesRequest) {
	request = &DescribePrometheusAgentInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusAgentInstances")

	return
}

func NewDescribePrometheusAgentInstancesResponse() (response *DescribePrometheusAgentInstancesResponse) {
	response = &DescribePrometheusAgentInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusAgentInstances
// 获取关联目标集群的实例列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribePrometheusAgentInstances(request *DescribePrometheusAgentInstancesRequest) (response *DescribePrometheusAgentInstancesResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusAgentInstancesRequest()
	}

	response = NewDescribePrometheusAgentInstancesResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusAgentInstances
// 获取关联目标集群的实例列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribePrometheusAgentInstancesWithContext(ctx context.Context, request *DescribePrometheusAgentInstancesRequest) (response *DescribePrometheusAgentInstancesResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusAgentInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusAgentInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusAgentsRequest() (request *DescribePrometheusAgentsRequest) {
	request = &DescribePrometheusAgentsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusAgents")

	return
}

func NewDescribePrometheusAgentsResponse() (response *DescribePrometheusAgentsResponse) {
	response = &DescribePrometheusAgentsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusAgents
// 获取被关联集群列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DescribePrometheusAgents(request *DescribePrometheusAgentsRequest) (response *DescribePrometheusAgentsResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusAgentsRequest()
	}

	response = NewDescribePrometheusAgentsResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusAgents
// 获取被关联集群列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DescribePrometheusAgentsWithContext(ctx context.Context, request *DescribePrometheusAgentsRequest) (response *DescribePrometheusAgentsResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusAgentsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusAgentsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusAlertHistoryRequest() (request *DescribePrometheusAlertHistoryRequest) {
	request = &DescribePrometheusAlertHistoryRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusAlertHistory")

	return
}

func NewDescribePrometheusAlertHistoryResponse() (response *DescribePrometheusAlertHistoryResponse) {
	response = &DescribePrometheusAlertHistoryResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusAlertHistory
// 获取告警历史
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DescribePrometheusAlertHistory(request *DescribePrometheusAlertHistoryRequest) (response *DescribePrometheusAlertHistoryResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusAlertHistoryRequest()
	}

	response = NewDescribePrometheusAlertHistoryResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusAlertHistory
// 获取告警历史
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DescribePrometheusAlertHistoryWithContext(ctx context.Context, request *DescribePrometheusAlertHistoryRequest) (response *DescribePrometheusAlertHistoryResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusAlertHistoryRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusAlertHistoryResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusAlertRuleRequest() (request *DescribePrometheusAlertRuleRequest) {
	request = &DescribePrometheusAlertRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusAlertRule")

	return
}

func NewDescribePrometheusAlertRuleResponse() (response *DescribePrometheusAlertRuleResponse) {
	response = &DescribePrometheusAlertRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusAlertRule
// 获取告警规则列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DescribePrometheusAlertRule(request *DescribePrometheusAlertRuleRequest) (response *DescribePrometheusAlertRuleResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusAlertRuleRequest()
	}

	response = NewDescribePrometheusAlertRuleResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusAlertRule
// 获取告警规则列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DescribePrometheusAlertRuleWithContext(ctx context.Context, request *DescribePrometheusAlertRuleRequest) (response *DescribePrometheusAlertRuleResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusAlertRuleRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusAlertRuleResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusInstanceRequest() (request *DescribePrometheusInstanceRequest) {
	request = &DescribePrometheusInstanceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusInstance")

	return
}

func NewDescribePrometheusInstanceResponse() (response *DescribePrometheusInstanceResponse) {
	response = &DescribePrometheusInstanceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusInstance
// 获取实例详细信息
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DescribePrometheusInstance(request *DescribePrometheusInstanceRequest) (response *DescribePrometheusInstanceResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusInstanceRequest()
	}

	response = NewDescribePrometheusInstanceResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusInstance
// 获取实例详细信息
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DescribePrometheusInstanceWithContext(ctx context.Context, request *DescribePrometheusInstanceRequest) (response *DescribePrometheusInstanceResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusInstanceRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusInstanceResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusOverviewsRequest() (request *DescribePrometheusOverviewsRequest) {
	request = &DescribePrometheusOverviewsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusOverviews")

	return
}

func NewDescribePrometheusOverviewsResponse() (response *DescribePrometheusOverviewsResponse) {
	response = &DescribePrometheusOverviewsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusOverviews
// 获取实例列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
func (c *Client) DescribePrometheusOverviews(request *DescribePrometheusOverviewsRequest) (response *DescribePrometheusOverviewsResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusOverviewsRequest()
	}

	response = NewDescribePrometheusOverviewsResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusOverviews
// 获取实例列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
func (c *Client) DescribePrometheusOverviewsWithContext(ctx context.Context, request *DescribePrometheusOverviewsRequest) (response *DescribePrometheusOverviewsResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusOverviewsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusOverviewsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusTargetsRequest() (request *DescribePrometheusTargetsRequest) {
	request = &DescribePrometheusTargetsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusTargets")

	return
}

func NewDescribePrometheusTargetsResponse() (response *DescribePrometheusTargetsResponse) {
	response = &DescribePrometheusTargetsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusTargets
// 获取targets信息
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PROMCLUSTERNOTFOUND = "InvalidParameter.PromClusterNotFound"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DescribePrometheusTargets(request *DescribePrometheusTargetsRequest) (response *DescribePrometheusTargetsResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusTargetsRequest()
	}

	response = NewDescribePrometheusTargetsResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusTargets
// 获取targets信息
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PROMCLUSTERNOTFOUND = "InvalidParameter.PromClusterNotFound"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
func (c *Client) DescribePrometheusTargetsWithContext(ctx context.Context, request *DescribePrometheusTargetsRequest) (response *DescribePrometheusTargetsResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusTargetsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusTargetsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusTemplateSyncRequest() (request *DescribePrometheusTemplateSyncRequest) {
	request = &DescribePrometheusTemplateSyncRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusTemplateSync")

	return
}

func NewDescribePrometheusTemplateSyncResponse() (response *DescribePrometheusTemplateSyncResponse) {
	response = &DescribePrometheusTemplateSyncResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusTemplateSync
// 获取模板同步信息
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) DescribePrometheusTemplateSync(request *DescribePrometheusTemplateSyncRequest) (response *DescribePrometheusTemplateSyncResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusTemplateSyncRequest()
	}

	response = NewDescribePrometheusTemplateSyncResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusTemplateSync
// 获取模板同步信息
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) DescribePrometheusTemplateSyncWithContext(ctx context.Context, request *DescribePrometheusTemplateSyncRequest) (response *DescribePrometheusTemplateSyncResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusTemplateSyncRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusTemplateSyncResponse()
	err = c.Send(request, response)
	return
}

func NewDescribePrometheusTemplatesRequest() (request *DescribePrometheusTemplatesRequest) {
	request = &DescribePrometheusTemplatesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribePrometheusTemplates")

	return
}

func NewDescribePrometheusTemplatesResponse() (response *DescribePrometheusTemplatesResponse) {
	response = &DescribePrometheusTemplatesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribePrometheusTemplates
// 拉取模板列表，默认模板将总是在最前面
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
func (c *Client) DescribePrometheusTemplates(request *DescribePrometheusTemplatesRequest) (response *DescribePrometheusTemplatesResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusTemplatesRequest()
	}

	response = NewDescribePrometheusTemplatesResponse()
	err = c.Send(request, response)
	return
}

// DescribePrometheusTemplates
// 拉取模板列表，默认模板将总是在最前面
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
func (c *Client) DescribePrometheusTemplatesWithContext(ctx context.Context, request *DescribePrometheusTemplatesRequest) (response *DescribePrometheusTemplatesResponse, err error) {
	if request == nil {
		request = NewDescribePrometheusTemplatesRequest()
	}
	request.SetContext(ctx)

	response = NewDescribePrometheusTemplatesResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeRegionsRequest() (request *DescribeRegionsRequest) {
	request = &DescribeRegionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeRegions")

	return
}

func NewDescribeRegionsResponse() (response *DescribeRegionsResponse) {
	response = &DescribeRegionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeRegions
// 获取容器服务支持的所有地域
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
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
// 获取容器服务支持的所有地域
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
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

func NewDescribeRouteTableConflictsRequest() (request *DescribeRouteTableConflictsRequest) {
	request = &DescribeRouteTableConflictsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeRouteTableConflicts")

	return
}

func NewDescribeRouteTableConflictsResponse() (response *DescribeRouteTableConflictsResponse) {
	response = &DescribeRouteTableConflictsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeRouteTableConflicts
// 查询路由表冲突列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CIDRMASKSIZEOUTOFRANGE = "InternalError.CidrMaskSizeOutOfRange"
//	INTERNALERROR_INVALIDPRIVATENETWORKCIDR = "InternalError.InvalidPrivateNetworkCidr"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribeRouteTableConflicts(request *DescribeRouteTableConflictsRequest) (response *DescribeRouteTableConflictsResponse, err error) {
	if request == nil {
		request = NewDescribeRouteTableConflictsRequest()
	}

	response = NewDescribeRouteTableConflictsResponse()
	err = c.Send(request, response)
	return
}

// DescribeRouteTableConflicts
// 查询路由表冲突列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CIDRMASKSIZEOUTOFRANGE = "InternalError.CidrMaskSizeOutOfRange"
//	INTERNALERROR_INVALIDPRIVATENETWORKCIDR = "InternalError.InvalidPrivateNetworkCidr"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) DescribeRouteTableConflictsWithContext(ctx context.Context, request *DescribeRouteTableConflictsRequest) (response *DescribeRouteTableConflictsResponse, err error) {
	if request == nil {
		request = NewDescribeRouteTableConflictsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeRouteTableConflictsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVersionsRequest() (request *DescribeVersionsRequest) {
	request = &DescribeVersionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeVersions")

	return
}

func NewDescribeVersionsResponse() (response *DescribeVersionsResponse) {
	response = &DescribeVersionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVersions
// 获取集群版本信息
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_COMPONENTCLINETHTTP = "InternalError.ComponentClinetHttp"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeVersions(request *DescribeVersionsRequest) (response *DescribeVersionsResponse, err error) {
	if request == nil {
		request = NewDescribeVersionsRequest()
	}

	response = NewDescribeVersionsResponse()
	err = c.Send(request, response)
	return
}

// DescribeVersions
// 获取集群版本信息
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_COMPONENTCLINETHTTP = "InternalError.ComponentClinetHttp"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeVersionsWithContext(ctx context.Context, request *DescribeVersionsRequest) (response *DescribeVersionsResponse, err error) {
	if request == nil {
		request = NewDescribeVersionsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVersionsResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeVpcCniPodLimitsRequest() (request *DescribeVpcCniPodLimitsRequest) {
	request = &DescribeVpcCniPodLimitsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeVpcCniPodLimits")

	return
}

func NewDescribeVpcCniPodLimitsResponse() (response *DescribeVpcCniPodLimitsResponse) {
	response = &DescribeVpcCniPodLimitsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeVpcCniPodLimits
// 本接口查询当前用户和地域在指定可用区下的机型可支持的最大 TKE VPC-CNI 网络模式的 Pod 数量
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	MISSINGPARAMETER = "MissingParameter"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeVpcCniPodLimits(request *DescribeVpcCniPodLimitsRequest) (response *DescribeVpcCniPodLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeVpcCniPodLimitsRequest()
	}

	response = NewDescribeVpcCniPodLimitsResponse()
	err = c.Send(request, response)
	return
}

// DescribeVpcCniPodLimits
// 本接口查询当前用户和地域在指定可用区下的机型可支持的最大 TKE VPC-CNI 网络模式的 Pod 数量
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	MISSINGPARAMETER = "MissingParameter"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) DescribeVpcCniPodLimitsWithContext(ctx context.Context, request *DescribeVpcCniPodLimitsRequest) (response *DescribeVpcCniPodLimitsResponse, err error) {
	if request == nil {
		request = NewDescribeVpcCniPodLimitsRequest()
	}
	request.SetContext(ctx)

	response = NewDescribeVpcCniPodLimitsResponse()
	err = c.Send(request, response)
	return
}

func NewDisableClusterDeletionProtectionRequest() (request *DisableClusterDeletionProtectionRequest) {
	request = &DisableClusterDeletionProtectionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DisableClusterDeletionProtection")

	return
}

func NewDisableClusterDeletionProtectionResponse() (response *DisableClusterDeletionProtectionResponse) {
	response = &DisableClusterDeletionProtectionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisableClusterDeletionProtection
// 关闭集群删除保护
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DisableClusterDeletionProtection(request *DisableClusterDeletionProtectionRequest) (response *DisableClusterDeletionProtectionResponse, err error) {
	if request == nil {
		request = NewDisableClusterDeletionProtectionRequest()
	}

	response = NewDisableClusterDeletionProtectionResponse()
	err = c.Send(request, response)
	return
}

// DisableClusterDeletionProtection
// 关闭集群删除保护
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DisableClusterDeletionProtectionWithContext(ctx context.Context, request *DisableClusterDeletionProtectionRequest) (response *DisableClusterDeletionProtectionResponse, err error) {
	if request == nil {
		request = NewDisableClusterDeletionProtectionRequest()
	}
	request.SetContext(ctx)

	response = NewDisableClusterDeletionProtectionResponse()
	err = c.Send(request, response)
	return
}

func NewDisableVpcCniNetworkTypeRequest() (request *DisableVpcCniNetworkTypeRequest) {
	request = &DisableVpcCniNetworkTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DisableVpcCniNetworkType")

	return
}

func NewDisableVpcCniNetworkTypeResponse() (response *DisableVpcCniNetworkTypeResponse) {
	response = &DisableVpcCniNetworkTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DisableVpcCniNetworkType
// 提供给附加了VPC-CNI能力的Global-Route集群关闭VPC-CNI
//
// 可能返回的错误码:
//
//	INTERNALERROR_KUBECLIENTCREATE = "InternalError.KubeClientCreate"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DisableVpcCniNetworkType(request *DisableVpcCniNetworkTypeRequest) (response *DisableVpcCniNetworkTypeResponse, err error) {
	if request == nil {
		request = NewDisableVpcCniNetworkTypeRequest()
	}

	response = NewDisableVpcCniNetworkTypeResponse()
	err = c.Send(request, response)
	return
}

// DisableVpcCniNetworkType
// 提供给附加了VPC-CNI能力的Global-Route集群关闭VPC-CNI
//
// 可能返回的错误码:
//
//	INTERNALERROR_KUBECLIENTCREATE = "InternalError.KubeClientCreate"
//	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) DisableVpcCniNetworkTypeWithContext(ctx context.Context, request *DisableVpcCniNetworkTypeRequest) (response *DisableVpcCniNetworkTypeResponse, err error) {
	if request == nil {
		request = NewDisableVpcCniNetworkTypeRequest()
	}
	request.SetContext(ctx)

	response = NewDisableVpcCniNetworkTypeResponse()
	err = c.Send(request, response)
	return
}

func NewEnableClusterDeletionProtectionRequest() (request *EnableClusterDeletionProtectionRequest) {
	request = &EnableClusterDeletionProtectionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "EnableClusterDeletionProtection")

	return
}

func NewEnableClusterDeletionProtectionResponse() (response *EnableClusterDeletionProtectionResponse) {
	response = &EnableClusterDeletionProtectionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// EnableClusterDeletionProtection
// 启用集群删除保护
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) EnableClusterDeletionProtection(request *EnableClusterDeletionProtectionRequest) (response *EnableClusterDeletionProtectionResponse, err error) {
	if request == nil {
		request = NewEnableClusterDeletionProtectionRequest()
	}

	response = NewEnableClusterDeletionProtectionResponse()
	err = c.Send(request, response)
	return
}

// EnableClusterDeletionProtection
// 启用集群删除保护
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) EnableClusterDeletionProtectionWithContext(ctx context.Context, request *EnableClusterDeletionProtectionRequest) (response *EnableClusterDeletionProtectionResponse, err error) {
	if request == nil {
		request = NewEnableClusterDeletionProtectionRequest()
	}
	request.SetContext(ctx)

	response = NewEnableClusterDeletionProtectionResponse()
	err = c.Send(request, response)
	return
}

func NewEnableVpcCniNetworkTypeRequest() (request *EnableVpcCniNetworkTypeRequest) {
	request = &EnableVpcCniNetworkTypeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "EnableVpcCniNetworkType")

	return
}

func NewEnableVpcCniNetworkTypeResponse() (response *EnableVpcCniNetworkTypeResponse) {
	response = &EnableVpcCniNetworkTypeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// EnableVpcCniNetworkType
// GR集群可以通过本接口附加vpc-cni容器网络插件，开启vpc-cni容器网络能力
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) EnableVpcCniNetworkType(request *EnableVpcCniNetworkTypeRequest) (response *EnableVpcCniNetworkTypeResponse, err error) {
	if request == nil {
		request = NewEnableVpcCniNetworkTypeRequest()
	}

	response = NewEnableVpcCniNetworkTypeResponse()
	err = c.Send(request, response)
	return
}

// EnableVpcCniNetworkType
// GR集群可以通过本接口附加vpc-cni容器网络插件，开启vpc-cni容器网络能力
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) EnableVpcCniNetworkTypeWithContext(ctx context.Context, request *EnableVpcCniNetworkTypeRequest) (response *EnableVpcCniNetworkTypeResponse, err error) {
	if request == nil {
		request = NewEnableVpcCniNetworkTypeRequest()
	}
	request.SetContext(ctx)

	response = NewEnableVpcCniNetworkTypeResponse()
	err = c.Send(request, response)
	return
}

func NewForwardApplicationRequestV3Request() (request *ForwardApplicationRequestV3Request) {
	request = &ForwardApplicationRequestV3Request{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ForwardApplicationRequestV3")

	return
}

func NewForwardApplicationRequestV3Response() (response *ForwardApplicationRequestV3Response) {
	response = &ForwardApplicationRequestV3Response{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ForwardApplicationRequestV3
// 操作TKE集群的addon
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	FAILEDOPERATION_RBACFORBIDDEN = "FailedOperation.RBACForbidden"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_NOTINWHITELIST = "UnsupportedOperation.NotInWhitelist"
func (c *Client) ForwardApplicationRequestV3(request *ForwardApplicationRequestV3Request) (response *ForwardApplicationRequestV3Response, err error) {
	if request == nil {
		request = NewForwardApplicationRequestV3Request()
	}

	response = NewForwardApplicationRequestV3Response()
	err = c.Send(request, response)
	return
}

// ForwardApplicationRequestV3
// 操作TKE集群的addon
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	FAILEDOPERATION_RBACFORBIDDEN = "FailedOperation.RBACForbidden"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
//	UNSUPPORTEDOPERATION_NOTINWHITELIST = "UnsupportedOperation.NotInWhitelist"
func (c *Client) ForwardApplicationRequestV3WithContext(ctx context.Context, request *ForwardApplicationRequestV3Request) (response *ForwardApplicationRequestV3Response, err error) {
	if request == nil {
		request = NewForwardApplicationRequestV3Request()
	}
	request.SetContext(ctx)

	response = NewForwardApplicationRequestV3Response()
	err = c.Send(request, response)
	return
}

func NewGetTkeAppChartListRequest() (request *GetTkeAppChartListRequest) {
	request = &GetTkeAppChartListRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "GetTkeAppChartList")

	return
}

func NewGetTkeAppChartListResponse() (response *GetTkeAppChartListResponse) {
	response = &GetTkeAppChartListResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// GetTkeAppChartList
// 获取TKE支持的App列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) GetTkeAppChartList(request *GetTkeAppChartListRequest) (response *GetTkeAppChartListResponse, err error) {
	if request == nil {
		request = NewGetTkeAppChartListRequest()
	}

	response = NewGetTkeAppChartListResponse()
	err = c.Send(request, response)
	return
}

// GetTkeAppChartList
// 获取TKE支持的App列表
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
func (c *Client) GetTkeAppChartListWithContext(ctx context.Context, request *GetTkeAppChartListRequest) (response *GetTkeAppChartListResponse, err error) {
	if request == nil {
		request = NewGetTkeAppChartListRequest()
	}
	request.SetContext(ctx)

	response = NewGetTkeAppChartListResponse()
	err = c.Send(request, response)
	return
}

func NewGetUpgradeInstanceProgressRequest() (request *GetUpgradeInstanceProgressRequest) {
	request = &GetUpgradeInstanceProgressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "GetUpgradeInstanceProgress")

	return
}

func NewGetUpgradeInstanceProgressResponse() (response *GetUpgradeInstanceProgressResponse) {
	response = &GetUpgradeInstanceProgressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// GetUpgradeInstanceProgress
// 获得节点升级当前的进度
//
// 可能返回的错误码:
//
//	INTERNALERROR_TASKNOTFOUND = "InternalError.TaskNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) GetUpgradeInstanceProgress(request *GetUpgradeInstanceProgressRequest) (response *GetUpgradeInstanceProgressResponse, err error) {
	if request == nil {
		request = NewGetUpgradeInstanceProgressRequest()
	}

	response = NewGetUpgradeInstanceProgressResponse()
	err = c.Send(request, response)
	return
}

// GetUpgradeInstanceProgress
// 获得节点升级当前的进度
//
// 可能返回的错误码:
//
//	INTERNALERROR_TASKNOTFOUND = "InternalError.TaskNotFound"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) GetUpgradeInstanceProgressWithContext(ctx context.Context, request *GetUpgradeInstanceProgressRequest) (response *GetUpgradeInstanceProgressResponse, err error) {
	if request == nil {
		request = NewGetUpgradeInstanceProgressRequest()
	}
	request.SetContext(ctx)

	response = NewGetUpgradeInstanceProgressResponse()
	err = c.Send(request, response)
	return
}

func NewModifyClusterAsGroupAttributeRequest() (request *ModifyClusterAsGroupAttributeRequest) {
	request = &ModifyClusterAsGroupAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyClusterAsGroupAttribute")

	return
}

func NewModifyClusterAsGroupAttributeResponse() (response *ModifyClusterAsGroupAttributeResponse) {
	response = &ModifyClusterAsGroupAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyClusterAsGroupAttribute
// 修改集群伸缩组属性
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"
//	INVALIDPARAMETER_CIDROUTOFROUTETABLE = "InvalidParameter.CidrOutOfRouteTable"
//	INVALIDPARAMETER_GATEWAYALREADYASSOCIATEDCIDR = "InvalidParameter.GatewayAlreadyAssociatedCidr"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyClusterAsGroupAttribute(request *ModifyClusterAsGroupAttributeRequest) (response *ModifyClusterAsGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifyClusterAsGroupAttributeRequest()
	}

	response = NewModifyClusterAsGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyClusterAsGroupAttribute
// 修改集群伸缩组属性
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"
//	INVALIDPARAMETER_CIDROUTOFROUTETABLE = "InvalidParameter.CidrOutOfRouteTable"
//	INVALIDPARAMETER_GATEWAYALREADYASSOCIATEDCIDR = "InvalidParameter.GatewayAlreadyAssociatedCidr"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyClusterAsGroupAttributeWithContext(ctx context.Context, request *ModifyClusterAsGroupAttributeRequest) (response *ModifyClusterAsGroupAttributeResponse, err error) {
	if request == nil {
		request = NewModifyClusterAsGroupAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyClusterAsGroupAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyClusterAsGroupOptionAttributeRequest() (request *ModifyClusterAsGroupOptionAttributeRequest) {
	request = &ModifyClusterAsGroupOptionAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyClusterAsGroupOptionAttribute")

	return
}

func NewModifyClusterAsGroupOptionAttributeResponse() (response *ModifyClusterAsGroupOptionAttributeResponse) {
	response = &ModifyClusterAsGroupOptionAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyClusterAsGroupOptionAttribute
// 修改集群弹性伸缩属性
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyClusterAsGroupOptionAttribute(request *ModifyClusterAsGroupOptionAttributeRequest) (response *ModifyClusterAsGroupOptionAttributeResponse, err error) {
	if request == nil {
		request = NewModifyClusterAsGroupOptionAttributeRequest()
	}

	response = NewModifyClusterAsGroupOptionAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyClusterAsGroupOptionAttribute
// 修改集群弹性伸缩属性
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"
//	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"
//	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyClusterAsGroupOptionAttributeWithContext(ctx context.Context, request *ModifyClusterAsGroupOptionAttributeRequest) (response *ModifyClusterAsGroupOptionAttributeResponse, err error) {
	if request == nil {
		request = NewModifyClusterAsGroupOptionAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyClusterAsGroupOptionAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyClusterAttributeRequest() (request *ModifyClusterAttributeRequest) {
	request = &ModifyClusterAttributeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyClusterAttribute")

	return
}

func NewModifyClusterAttributeResponse() (response *ModifyClusterAttributeResponse) {
	response = &ModifyClusterAttributeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyClusterAttribute
// 修改集群属性
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
func (c *Client) ModifyClusterAttribute(request *ModifyClusterAttributeRequest) (response *ModifyClusterAttributeResponse, err error) {
	if request == nil {
		request = NewModifyClusterAttributeRequest()
	}

	response = NewModifyClusterAttributeResponse()
	err = c.Send(request, response)
	return
}

// ModifyClusterAttribute
// 修改集群属性
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
func (c *Client) ModifyClusterAttributeWithContext(ctx context.Context, request *ModifyClusterAttributeRequest) (response *ModifyClusterAttributeResponse, err error) {
	if request == nil {
		request = NewModifyClusterAttributeRequest()
	}
	request.SetContext(ctx)

	response = NewModifyClusterAttributeResponse()
	err = c.Send(request, response)
	return
}

func NewModifyClusterAuthenticationOptionsRequest() (request *ModifyClusterAuthenticationOptionsRequest) {
	request = &ModifyClusterAuthenticationOptionsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyClusterAuthenticationOptions")

	return
}

func NewModifyClusterAuthenticationOptionsResponse() (response *ModifyClusterAuthenticationOptionsResponse) {
	response = &ModifyClusterAuthenticationOptionsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyClusterAuthenticationOptions
// 修改集群认证配置
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) ModifyClusterAuthenticationOptions(request *ModifyClusterAuthenticationOptionsRequest) (response *ModifyClusterAuthenticationOptionsResponse, err error) {
	if request == nil {
		request = NewModifyClusterAuthenticationOptionsRequest()
	}

	response = NewModifyClusterAuthenticationOptionsResponse()
	err = c.Send(request, response)
	return
}

// ModifyClusterAuthenticationOptions
// 修改集群认证配置
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) ModifyClusterAuthenticationOptionsWithContext(ctx context.Context, request *ModifyClusterAuthenticationOptionsRequest) (response *ModifyClusterAuthenticationOptionsResponse, err error) {
	if request == nil {
		request = NewModifyClusterAuthenticationOptionsRequest()
	}
	request.SetContext(ctx)

	response = NewModifyClusterAuthenticationOptionsResponse()
	err = c.Send(request, response)
	return
}

func NewModifyClusterEndpointSPRequest() (request *ModifyClusterEndpointSPRequest) {
	request = &ModifyClusterEndpointSPRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyClusterEndpointSP")

	return
}

func NewModifyClusterEndpointSPResponse() (response *ModifyClusterEndpointSPResponse) {
	response = &ModifyClusterEndpointSPResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyClusterEndpointSP
// 修改托管集群外网端口的安全策略（老的方式，仅支持托管集群外网端口）
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCUNEXPECTEDERROR = "InternalError.VPCUnexpectedError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyClusterEndpointSP(request *ModifyClusterEndpointSPRequest) (response *ModifyClusterEndpointSPResponse, err error) {
	if request == nil {
		request = NewModifyClusterEndpointSPRequest()
	}

	response = NewModifyClusterEndpointSPResponse()
	err = c.Send(request, response)
	return
}

// ModifyClusterEndpointSP
// 修改托管集群外网端口的安全策略（老的方式，仅支持托管集群外网端口）
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INTERNALERROR_VPCUNEXPECTEDERROR = "InternalError.VPCUnexpectedError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	OPERATIONDENIED = "OperationDenied"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) ModifyClusterEndpointSPWithContext(ctx context.Context, request *ModifyClusterEndpointSPRequest) (response *ModifyClusterEndpointSPResponse, err error) {
	if request == nil {
		request = NewModifyClusterEndpointSPRequest()
	}
	request.SetContext(ctx)

	response = NewModifyClusterEndpointSPResponse()
	err = c.Send(request, response)
	return
}

func NewModifyClusterNodePoolRequest() (request *ModifyClusterNodePoolRequest) {
	request = &ModifyClusterNodePoolRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyClusterNodePool")

	return
}

func NewModifyClusterNodePoolResponse() (response *ModifyClusterNodePoolResponse) {
	response = &ModifyClusterNodePoolResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyClusterNodePool
// 编辑节点池
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	OPERATIONDENIED = "OperationDenied"
//	UNSUPPORTEDOPERATION_CAENABLEFAILED = "UnsupportedOperation.CaEnableFailed"
func (c *Client) ModifyClusterNodePool(request *ModifyClusterNodePoolRequest) (response *ModifyClusterNodePoolResponse, err error) {
	if request == nil {
		request = NewModifyClusterNodePoolRequest()
	}

	response = NewModifyClusterNodePoolResponse()
	err = c.Send(request, response)
	return
}

// ModifyClusterNodePool
// 编辑节点池
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	OPERATIONDENIED = "OperationDenied"
//	UNSUPPORTEDOPERATION_CAENABLEFAILED = "UnsupportedOperation.CaEnableFailed"
func (c *Client) ModifyClusterNodePoolWithContext(ctx context.Context, request *ModifyClusterNodePoolRequest) (response *ModifyClusterNodePoolResponse, err error) {
	if request == nil {
		request = NewModifyClusterNodePoolRequest()
	}
	request.SetContext(ctx)

	response = NewModifyClusterNodePoolResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNodePoolDesiredCapacityAboutAsgRequest() (request *ModifyNodePoolDesiredCapacityAboutAsgRequest) {
	request = &ModifyNodePoolDesiredCapacityAboutAsgRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyNodePoolDesiredCapacityAboutAsg")

	return
}

func NewModifyNodePoolDesiredCapacityAboutAsgResponse() (response *ModifyNodePoolDesiredCapacityAboutAsgResponse) {
	response = &ModifyNodePoolDesiredCapacityAboutAsgResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNodePoolDesiredCapacityAboutAsg
// 修改节点池关联伸缩组的期望实例数
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_ASASGNOTEXIST = "ResourceNotFound.AsAsgNotExist"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
func (c *Client) ModifyNodePoolDesiredCapacityAboutAsg(request *ModifyNodePoolDesiredCapacityAboutAsgRequest) (response *ModifyNodePoolDesiredCapacityAboutAsgResponse, err error) {
	if request == nil {
		request = NewModifyNodePoolDesiredCapacityAboutAsgRequest()
	}

	response = NewModifyNodePoolDesiredCapacityAboutAsgResponse()
	err = c.Send(request, response)
	return
}

// ModifyNodePoolDesiredCapacityAboutAsg
// 修改节点池关联伸缩组的期望实例数
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND_ASASGNOTEXIST = "ResourceNotFound.AsAsgNotExist"
//	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"
//	UNKNOWNPARAMETER = "UnknownParameter"
func (c *Client) ModifyNodePoolDesiredCapacityAboutAsgWithContext(ctx context.Context, request *ModifyNodePoolDesiredCapacityAboutAsgRequest) (response *ModifyNodePoolDesiredCapacityAboutAsgResponse, err error) {
	if request == nil {
		request = NewModifyNodePoolDesiredCapacityAboutAsgRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNodePoolDesiredCapacityAboutAsgResponse()
	err = c.Send(request, response)
	return
}

func NewModifyNodePoolInstanceTypesRequest() (request *ModifyNodePoolInstanceTypesRequest) {
	request = &ModifyNodePoolInstanceTypesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyNodePoolInstanceTypes")

	return
}

func NewModifyNodePoolInstanceTypesResponse() (response *ModifyNodePoolInstanceTypesResponse) {
	response = &ModifyNodePoolInstanceTypesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyNodePoolInstanceTypes
// 修改节点池的机型配置
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) ModifyNodePoolInstanceTypes(request *ModifyNodePoolInstanceTypesRequest) (response *ModifyNodePoolInstanceTypesResponse, err error) {
	if request == nil {
		request = NewModifyNodePoolInstanceTypesRequest()
	}

	response = NewModifyNodePoolInstanceTypesResponse()
	err = c.Send(request, response)
	return
}

// ModifyNodePoolInstanceTypes
// 修改节点池的机型配置
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) ModifyNodePoolInstanceTypesWithContext(ctx context.Context, request *ModifyNodePoolInstanceTypesRequest) (response *ModifyNodePoolInstanceTypesResponse, err error) {
	if request == nil {
		request = NewModifyNodePoolInstanceTypesRequest()
	}
	request.SetContext(ctx)

	response = NewModifyNodePoolInstanceTypesResponse()
	err = c.Send(request, response)
	return
}

func NewModifyPrometheusAlertRuleRequest() (request *ModifyPrometheusAlertRuleRequest) {
	request = &ModifyPrometheusAlertRuleRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyPrometheusAlertRule")

	return
}

func NewModifyPrometheusAlertRuleResponse() (response *ModifyPrometheusAlertRuleResponse) {
	response = &ModifyPrometheusAlertRuleResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyPrometheusAlertRule
// 修改告警规则
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) ModifyPrometheusAlertRule(request *ModifyPrometheusAlertRuleRequest) (response *ModifyPrometheusAlertRuleResponse, err error) {
	if request == nil {
		request = NewModifyPrometheusAlertRuleRequest()
	}

	response = NewModifyPrometheusAlertRuleResponse()
	err = c.Send(request, response)
	return
}

// ModifyPrometheusAlertRule
// 修改告警规则
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) ModifyPrometheusAlertRuleWithContext(ctx context.Context, request *ModifyPrometheusAlertRuleRequest) (response *ModifyPrometheusAlertRuleResponse, err error) {
	if request == nil {
		request = NewModifyPrometheusAlertRuleRequest()
	}
	request.SetContext(ctx)

	response = NewModifyPrometheusAlertRuleResponse()
	err = c.Send(request, response)
	return
}

func NewModifyPrometheusTemplateRequest() (request *ModifyPrometheusTemplateRequest) {
	request = &ModifyPrometheusTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyPrometheusTemplate")

	return
}

func NewModifyPrometheusTemplateResponse() (response *ModifyPrometheusTemplateResponse) {
	response = &ModifyPrometheusTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyPrometheusTemplate
// 修改模板内容
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) ModifyPrometheusTemplate(request *ModifyPrometheusTemplateRequest) (response *ModifyPrometheusTemplateResponse, err error) {
	if request == nil {
		request = NewModifyPrometheusTemplateRequest()
	}

	response = NewModifyPrometheusTemplateResponse()
	err = c.Send(request, response)
	return
}

// ModifyPrometheusTemplate
// 修改模板内容
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) ModifyPrometheusTemplateWithContext(ctx context.Context, request *ModifyPrometheusTemplateRequest) (response *ModifyPrometheusTemplateResponse, err error) {
	if request == nil {
		request = NewModifyPrometheusTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewModifyPrometheusTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewRemoveNodeFromNodePoolRequest() (request *RemoveNodeFromNodePoolRequest) {
	request = &RemoveNodeFromNodePoolRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "RemoveNodeFromNodePool")

	return
}

func NewRemoveNodeFromNodePoolResponse() (response *RemoveNodeFromNodePoolResponse) {
	response = &RemoveNodeFromNodePoolResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RemoveNodeFromNodePool
// 移出节点池节点，但保留在集群内
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) RemoveNodeFromNodePool(request *RemoveNodeFromNodePoolRequest) (response *RemoveNodeFromNodePoolResponse, err error) {
	if request == nil {
		request = NewRemoveNodeFromNodePoolRequest()
	}

	response = NewRemoveNodeFromNodePoolResponse()
	err = c.Send(request, response)
	return
}

// RemoveNodeFromNodePool
// 移出节点池节点，但保留在集群内
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) RemoveNodeFromNodePoolWithContext(ctx context.Context, request *RemoveNodeFromNodePoolRequest) (response *RemoveNodeFromNodePoolResponse, err error) {
	if request == nil {
		request = NewRemoveNodeFromNodePoolRequest()
	}
	request.SetContext(ctx)

	response = NewRemoveNodeFromNodePoolResponse()
	err = c.Send(request, response)
	return
}

func NewRestartEKSContainerInstancesRequest() (request *RestartEKSContainerInstancesRequest) {
	request = &RestartEKSContainerInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "RestartEKSContainerInstances")

	return
}

func NewRestartEKSContainerInstancesResponse() (response *RestartEKSContainerInstancesResponse) {
	response = &RestartEKSContainerInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// RestartEKSContainerInstances
// 重启弹性容器实例，支持批量操作
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) RestartEKSContainerInstances(request *RestartEKSContainerInstancesRequest) (response *RestartEKSContainerInstancesResponse, err error) {
	if request == nil {
		request = NewRestartEKSContainerInstancesRequest()
	}

	response = NewRestartEKSContainerInstancesResponse()
	err = c.Send(request, response)
	return
}

// RestartEKSContainerInstances
// 重启弹性容器实例，支持批量操作
//
// 可能返回的错误码:
//
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) RestartEKSContainerInstancesWithContext(ctx context.Context, request *RestartEKSContainerInstancesRequest) (response *RestartEKSContainerInstancesResponse, err error) {
	if request == nil {
		request = NewRestartEKSContainerInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewRestartEKSContainerInstancesResponse()
	err = c.Send(request, response)
	return
}

func NewScaleInClusterMasterRequest() (request *ScaleInClusterMasterRequest) {
	request = &ScaleInClusterMasterRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ScaleInClusterMaster")

	return
}

func NewScaleInClusterMasterResponse() (response *ScaleInClusterMasterResponse) {
	response = &ScaleInClusterMasterResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ScaleInClusterMaster
// 缩容独立集群master节点
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	OPERATIONDENIED = "OperationDenied"
func (c *Client) ScaleInClusterMaster(request *ScaleInClusterMasterRequest) (response *ScaleInClusterMasterResponse, err error) {
	if request == nil {
		request = NewScaleInClusterMasterRequest()
	}

	response = NewScaleInClusterMasterResponse()
	err = c.Send(request, response)
	return
}

// ScaleInClusterMaster
// 缩容独立集群master节点
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	OPERATIONDENIED = "OperationDenied"
func (c *Client) ScaleInClusterMasterWithContext(ctx context.Context, request *ScaleInClusterMasterRequest) (response *ScaleInClusterMasterResponse, err error) {
	if request == nil {
		request = NewScaleInClusterMasterRequest()
	}
	request.SetContext(ctx)

	response = NewScaleInClusterMasterResponse()
	err = c.Send(request, response)
	return
}

func NewScaleOutClusterMasterRequest() (request *ScaleOutClusterMasterRequest) {
	request = &ScaleOutClusterMasterRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ScaleOutClusterMaster")

	return
}

func NewScaleOutClusterMasterResponse() (response *ScaleOutClusterMasterResponse) {
	response = &ScaleOutClusterMasterResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ScaleOutClusterMaster
// 扩容独立集群master节点
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	OPERATIONDENIED = "OperationDenied"
func (c *Client) ScaleOutClusterMaster(request *ScaleOutClusterMasterRequest) (response *ScaleOutClusterMasterResponse, err error) {
	if request == nil {
		request = NewScaleOutClusterMasterRequest()
	}

	response = NewScaleOutClusterMasterResponse()
	err = c.Send(request, response)
	return
}

// ScaleOutClusterMaster
// 扩容独立集群master节点
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	OPERATIONDENIED = "OperationDenied"
func (c *Client) ScaleOutClusterMasterWithContext(ctx context.Context, request *ScaleOutClusterMasterRequest) (response *ScaleOutClusterMasterResponse, err error) {
	if request == nil {
		request = NewScaleOutClusterMasterRequest()
	}
	request.SetContext(ctx)

	response = NewScaleOutClusterMasterResponse()
	err = c.Send(request, response)
	return
}

func NewSetNodePoolNodeProtectionRequest() (request *SetNodePoolNodeProtectionRequest) {
	request = &SetNodePoolNodeProtectionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "SetNodePoolNodeProtection")

	return
}

func NewSetNodePoolNodeProtectionResponse() (response *SetNodePoolNodeProtectionResponse) {
	response = &SetNodePoolNodeProtectionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// SetNodePoolNodeProtection
// 仅能设置节点池中处于伸缩组的节点
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) SetNodePoolNodeProtection(request *SetNodePoolNodeProtectionRequest) (response *SetNodePoolNodeProtectionResponse, err error) {
	if request == nil {
		request = NewSetNodePoolNodeProtectionRequest()
	}

	response = NewSetNodePoolNodeProtectionResponse()
	err = c.Send(request, response)
	return
}

// SetNodePoolNodeProtection
// 仅能设置节点池中处于伸缩组的节点
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
func (c *Client) SetNodePoolNodeProtectionWithContext(ctx context.Context, request *SetNodePoolNodeProtectionRequest) (response *SetNodePoolNodeProtectionResponse, err error) {
	if request == nil {
		request = NewSetNodePoolNodeProtectionRequest()
	}
	request.SetContext(ctx)

	response = NewSetNodePoolNodeProtectionResponse()
	err = c.Send(request, response)
	return
}

func NewSyncPrometheusTemplateRequest() (request *SyncPrometheusTemplateRequest) {
	request = &SyncPrometheusTemplateRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "SyncPrometheusTemplate")

	return
}

func NewSyncPrometheusTemplateResponse() (response *SyncPrometheusTemplateResponse) {
	response = &SyncPrometheusTemplateResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// SyncPrometheusTemplate
// 同步模板到实例或者集群
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_PROMCLUSTERNOTFOUND = "InvalidParameter.PromClusterNotFound"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) SyncPrometheusTemplate(request *SyncPrometheusTemplateRequest) (response *SyncPrometheusTemplateResponse, err error) {
	if request == nil {
		request = NewSyncPrometheusTemplateRequest()
	}

	response = NewSyncPrometheusTemplateResponse()
	err = c.Send(request, response)
	return
}

// SyncPrometheusTemplate
// 同步模板到实例或者集群
//
// 可能返回的错误码:
//
//	INTERNALERROR_DB = "InternalError.Db"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	INVALIDPARAMETER_PROMCLUSTERNOTFOUND = "InvalidParameter.PromClusterNotFound"
//	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"
//	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"
func (c *Client) SyncPrometheusTemplateWithContext(ctx context.Context, request *SyncPrometheusTemplateRequest) (response *SyncPrometheusTemplateResponse, err error) {
	if request == nil {
		request = NewSyncPrometheusTemplateRequest()
	}
	request.SetContext(ctx)

	response = NewSyncPrometheusTemplateResponse()
	err = c.Send(request, response)
	return
}

func NewUpdateClusterVersionRequest() (request *UpdateClusterVersionRequest) {
	request = &UpdateClusterVersionRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "UpdateClusterVersion")

	return
}

func NewUpdateClusterVersionResponse() (response *UpdateClusterVersionResponse) {
	response = &UpdateClusterVersionResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UpdateClusterVersion
// 升级集群 Master 组件到指定版本
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CLUSTERUPGRADENODEVERSION = "InternalError.ClusterUpgradeNodeVersion"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) UpdateClusterVersion(request *UpdateClusterVersionRequest) (response *UpdateClusterVersionResponse, err error) {
	if request == nil {
		request = NewUpdateClusterVersionRequest()
	}

	response = NewUpdateClusterVersionResponse()
	err = c.Send(request, response)
	return
}

// UpdateClusterVersion
// 升级集群 Master 组件到指定版本
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CLUSTERUPGRADENODEVERSION = "InternalError.ClusterUpgradeNodeVersion"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) UpdateClusterVersionWithContext(ctx context.Context, request *UpdateClusterVersionRequest) (response *UpdateClusterVersionResponse, err error) {
	if request == nil {
		request = NewUpdateClusterVersionRequest()
	}
	request.SetContext(ctx)

	response = NewUpdateClusterVersionResponse()
	err = c.Send(request, response)
	return
}

func NewUpdateEKSClusterRequest() (request *UpdateEKSClusterRequest) {
	request = &UpdateEKSClusterRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "UpdateEKSCluster")

	return
}

func NewUpdateEKSClusterResponse() (response *UpdateEKSClusterResponse) {
	response = &UpdateEKSClusterResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UpdateEKSCluster
// 修改弹性集群名称等属性
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) UpdateEKSCluster(request *UpdateEKSClusterRequest) (response *UpdateEKSClusterResponse, err error) {
	if request == nil {
		request = NewUpdateEKSClusterRequest()
	}

	response = NewUpdateEKSClusterResponse()
	err = c.Send(request, response)
	return
}

// UpdateEKSCluster
// 修改弹性集群名称等属性
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"
//	INVALIDPARAMETER = "InvalidParameter"
//	LIMITEXCEEDED = "LimitExceeded"
//	MISSINGPARAMETER = "MissingParameter"
//	RESOURCEINUSE = "ResourceInUse"
//	RESOURCENOTFOUND = "ResourceNotFound"
//	RESOURCEUNAVAILABLE = "ResourceUnavailable"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
//	UNKNOWNPARAMETER = "UnknownParameter"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) UpdateEKSClusterWithContext(ctx context.Context, request *UpdateEKSClusterRequest) (response *UpdateEKSClusterResponse, err error) {
	if request == nil {
		request = NewUpdateEKSClusterRequest()
	}
	request.SetContext(ctx)

	response = NewUpdateEKSClusterResponse()
	err = c.Send(request, response)
	return
}

func NewUpdateEKSContainerInstanceRequest() (request *UpdateEKSContainerInstanceRequest) {
	request = &UpdateEKSContainerInstanceRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "UpdateEKSContainerInstance")

	return
}

func NewUpdateEKSContainerInstanceResponse() (response *UpdateEKSContainerInstanceResponse) {
	response = &UpdateEKSContainerInstanceResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UpdateEKSContainerInstance
// 更新容器实例
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) UpdateEKSContainerInstance(request *UpdateEKSContainerInstanceRequest) (response *UpdateEKSContainerInstanceResponse, err error) {
	if request == nil {
		request = NewUpdateEKSContainerInstanceRequest()
	}

	response = NewUpdateEKSContainerInstanceResponse()
	err = c.Send(request, response)
	return
}

// UpdateEKSContainerInstance
// 更新容器实例
//
// 可能返回的错误码:
//
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	UNSUPPORTEDOPERATION = "UnsupportedOperation"
func (c *Client) UpdateEKSContainerInstanceWithContext(ctx context.Context, request *UpdateEKSContainerInstanceRequest) (response *UpdateEKSContainerInstanceResponse, err error) {
	if request == nil {
		request = NewUpdateEKSContainerInstanceRequest()
	}
	request.SetContext(ctx)

	response = NewUpdateEKSContainerInstanceResponse()
	err = c.Send(request, response)
	return
}

func NewUpgradeClusterInstancesRequest() (request *UpgradeClusterInstancesRequest) {
	request = &UpgradeClusterInstancesRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "UpgradeClusterInstances")

	return
}

func NewUpgradeClusterInstancesResponse() (response *UpgradeClusterInstancesResponse) {
	response = &UpgradeClusterInstancesResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// UpgradeClusterInstances
// 给集群的一批work节点进行升级
//
// 可能返回的错误码:
//
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_TASKLIFESTATEERROR = "InternalError.TaskLifeStateError"
//	INTERNALERROR_TASKNOTFOUND = "InternalError.TaskNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) UpgradeClusterInstances(request *UpgradeClusterInstancesRequest) (response *UpgradeClusterInstancesResponse, err error) {
	if request == nil {
		request = NewUpgradeClusterInstancesRequest()
	}

	response = NewUpgradeClusterInstancesResponse()
	err = c.Send(request, response)
	return
}

// UpgradeClusterInstances
// 给集群的一批work节点进行升级
//
// 可能返回的错误码:
//
//	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"
//	INTERNALERROR_PARAM = "InternalError.Param"
//	INTERNALERROR_TASKLIFESTATEERROR = "InternalError.TaskLifeStateError"
//	INTERNALERROR_TASKNOTFOUND = "InternalError.TaskNotFound"
//	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"
//	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"
//	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"
func (c *Client) UpgradeClusterInstancesWithContext(ctx context.Context, request *UpgradeClusterInstancesRequest) (response *UpgradeClusterInstancesResponse, err error) {
	if request == nil {
		request = NewUpgradeClusterInstancesRequest()
	}
	request.SetContext(ctx)

	response = NewUpgradeClusterInstancesResponse()
	err = c.Send(request, response)
	return
}
