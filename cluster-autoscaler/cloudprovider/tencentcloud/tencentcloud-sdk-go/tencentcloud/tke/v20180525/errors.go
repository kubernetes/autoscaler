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

const (
	// 此产品的特有错误码

	// 操作失败。
	FAILEDOPERATION = "FailedOperation"

	// 子账户RBAC权限不足。
	FAILEDOPERATION_RBACFORBIDDEN = "FailedOperation.RBACForbidden"

	// 内部错误。
	INTERNALERROR = "InternalError"

	// 获取用户认证信息失败。
	INTERNALERROR_ACCOUNTCOMMON = "InternalError.AccountCommon"

	// 账户未通过认证。
	INTERNALERROR_ACCOUNTUSERNOTAUTHENTICATED = "InternalError.AccountUserNotAuthenticated"

	// 伸缩组资源创建报错。
	INTERNALERROR_ASCOMMON = "InternalError.AsCommon"

	// 没有权限。
	INTERNALERROR_CAMNOAUTH = "InternalError.CamNoAuth"

	// CIDR和其他集群CIDR冲突。
	INTERNALERROR_CIDRCONFLICTWITHOTHERCLUSTER = "InternalError.CidrConflictWithOtherCluster"

	// CIDR和其他路由冲突。
	INTERNALERROR_CIDRCONFLICTWITHOTHERROUTE = "InternalError.CidrConflictWithOtherRoute"

	// CIDR和vpc冲突。
	INTERNALERROR_CIDRCONFLICTWITHVPCCIDR = "InternalError.CidrConflictWithVpcCidr"

	// CIDR和全局路由冲突。
	INTERNALERROR_CIDRCONFLICTWITHVPCGLOBALROUTE = "InternalError.CidrConflictWithVpcGlobalRoute"

	// CIDR无效。
	INTERNALERROR_CIDRINVALI = "InternalError.CidrInvali"

	// CIDR掩码无效。
	INTERNALERROR_CIDRMASKSIZEOUTOFRANGE = "InternalError.CidrMaskSizeOutOfRange"

	// CIDR不在路由表之内。
	INTERNALERROR_CIDROUTOFROUTETABLE = "InternalError.CidrOutOfRouteTable"

	// 集群未找到。
	INTERNALERROR_CLUSTERNOTFOUND = "InternalError.ClusterNotFound"

	// 集群状态错误。
	INTERNALERROR_CLUSTERSTATE = "InternalError.ClusterState"

	// 集群节点版本过低。
	INTERNALERROR_CLUSTERUPGRADENODEVERSION = "InternalError.ClusterUpgradeNodeVersion"

	// 执行命令超时。
	INTERNALERROR_CMDTIMEOUT = "InternalError.CmdTimeout"

	// 内部HTTP客户端错误。
	INTERNALERROR_COMPONENTCLIENTHTTP = "InternalError.ComponentClientHttp"

	// 请求(http请求)其他云服务失败。
	INTERNALERROR_COMPONENTCLINETHTTP = "InternalError.ComponentClinetHttp"

	// 容器未找到。
	INTERNALERROR_CONTAINERNOTFOUND = "InternalError.ContainerNotFound"

	// 创建集群失败。
	INTERNALERROR_CREATEMASTERFAILED = "InternalError.CreateMasterFailed"

	// cvm创建节点报错。
	INTERNALERROR_CVMCOMMON = "InternalError.CvmCommon"

	// cvm不存在。
	INTERNALERROR_CVMNOTFOUND = "InternalError.CvmNotFound"

	// 存在云服务器在CVM侧查询不到。
	INTERNALERROR_CVMNUMBERNOTMATCH = "InternalError.CvmNumberNotMatch"

	// cvm状态不正常。
	INTERNALERROR_CVMSTATUS = "InternalError.CvmStatus"

	// db错误。
	INTERNALERROR_DB = "InternalError.Db"

	// DB错误。
	INTERNALERROR_DBAFFECTIVEDROWS = "InternalError.DbAffectivedRows"

	// 记录未找到。
	INTERNALERROR_DBRECORDNOTFOUND = "InternalError.DbRecordNotFound"

	// 获得当前安全组总数失败。
	INTERNALERROR_DFWGETUSGCOUNT = "InternalError.DfwGetUSGCount"

	// 获得安全组配额失败。
	INTERNALERROR_DFWGETUSGQUOTA = "InternalError.DfwGetUSGQuota"

	// 不支持空集群。
	INTERNALERROR_EMPTYCLUSTERNOTSUPPORT = "InternalError.EmptyClusterNotSupport"

	// 下一跳地址已关联CIDR。
	INTERNALERROR_GATEWAYALREADYASSOCIATEDCIDR = "InternalError.GatewayAlreadyAssociatedCidr"

	// 镜像未找到。
	INTERNALERROR_IMAGEIDNOTFOUND = "InternalError.ImageIdNotFound"

	// 初始化master失败。
	INTERNALERROR_INITMASTERFAILED = "InternalError.InitMasterFailed"

	// 无效CIDR。
	INTERNALERROR_INVALIDPRIVATENETWORKCIDR = "InternalError.InvalidPrivateNetworkCidr"

	// 连接用户Kubernetes集群失败。
	INTERNALERROR_KUBECLIENTCONNECTION = "InternalError.KubeClientConnection"

	// 创建集群Client出错。
	INTERNALERROR_KUBECLIENTCREATE = "InternalError.KubeClientCreate"

	// KubernetesAPI错误。
	INTERNALERROR_KUBECOMMON = "InternalError.KubeCommon"

	// kubernetes client建立失败。
	INTERNALERROR_KUBERNETESCLIENTBUILDERROR = "InternalError.KubernetesClientBuildError"

	// 创建Kubernetes资源失败。
	INTERNALERROR_KUBERNETESCREATEOPERATIONERROR = "InternalError.KubernetesCreateOperationError"

	// 删除Kubernetes资源失败。
	INTERNALERROR_KUBERNETESDELETEOPERATIONERROR = "InternalError.KubernetesDeleteOperationError"

	// 获取Kubernetes资源失败。
	INTERNALERROR_KUBERNETESGETOPERATIONERROR = "InternalError.KubernetesGetOperationError"

	// Kubernetes未知错误。
	INTERNALERROR_KUBERNETESINTERNAL = "InternalError.KubernetesInternal"

	// 底层调用CLB未知错误。
	INTERNALERROR_LBCOMMON = "InternalError.LbCommon"

	// 镜像OS不支持。
	INTERNALERROR_OSNOTSUPPORT = "InternalError.OsNotSupport"

	// Param。
	INTERNALERROR_PARAM = "InternalError.Param"

	// Pod未找到。
	INTERNALERROR_PODNOTFOUND = "InternalError.PodNotFound"

	// 集群不支持当前操作。
	INTERNALERROR_PUBLICCLUSTEROPNOTSUPPORT = "InternalError.PublicClusterOpNotSupport"

	// 超过配额限制。
	INTERNALERROR_QUOTAMAXCLSLIMIT = "InternalError.QuotaMaxClsLimit"

	// 超过配额限制。
	INTERNALERROR_QUOTAMAXNODLIMIT = "InternalError.QuotaMaxNodLimit"

	// 超过配额限制。
	INTERNALERROR_QUOTAMAXRTLIMIT = "InternalError.QuotaMaxRtLimit"

	// 安全组配额不足。
	INTERNALERROR_QUOTAUSGLIMIT = "InternalError.QuotaUSGLimit"

	// 资源已存在。
	INTERNALERROR_RESOURCEEXISTALREADY = "InternalError.ResourceExistAlready"

	// 路由表非空。
	INTERNALERROR_ROUTETABLENOTEMPTY = "InternalError.RouteTableNotEmpty"

	// 路由表不存在。
	INTERNALERROR_ROUTETABLENOTFOUND = "InternalError.RouteTableNotFound"

	// 创建任务失败。
	INTERNALERROR_TASKCREATEFAILED = "InternalError.TaskCreateFailed"

	// 任务当前所处状态不支持此操作。
	INTERNALERROR_TASKLIFESTATEERROR = "InternalError.TaskLifeStateError"

	// 任务未找到。
	INTERNALERROR_TASKNOTFOUND = "InternalError.TaskNotFound"

	// 内部错误。
	INTERNALERROR_UNEXCEPTEDINTERNAL = "InternalError.UnexceptedInternal"

	// 未知的内部错误。
	INTERNALERROR_UNEXPECTEDINTERNAL = "InternalError.UnexpectedInternal"

	// VPC未知错误。
	INTERNALERROR_VPCUNEXPECTEDERROR = "InternalError.VPCUnexpectedError"

	// VPC报错。
	INTERNALERROR_VPCCOMMON = "InternalError.VpcCommon"

	// 对等连接不存在。
	INTERNALERROR_VPCPEERNOTFOUND = "InternalError.VpcPeerNotFound"

	// 未发现vpc记录。
	INTERNALERROR_VPCRECODRNOTFOUND = "InternalError.VpcRecodrNotFound"

	// 白名单未知错误。
	INTERNALERROR_WHITELISTUNEXPECTEDERROR = "InternalError.WhitelistUnexpectedError"

	// 参数错误。
	INVALIDPARAMETER = "InvalidParameter"

	// 弹性伸缩组创建参数错误。
	INVALIDPARAMETER_ASCOMMONERROR = "InvalidParameter.AsCommonError"

	// CIDR掩码超出范围(集群CIDR范围 最小值: 10 最大值: 24)。
	INVALIDPARAMETER_CIDRMASKSIZEOUTOFRANGE = "InvalidParameter.CIDRMaskSizeOutOfRange"

	// CIDR和其他集群CIDR冲突。
	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERCLUSTER = "InvalidParameter.CidrConflictWithOtherCluster"

	// 创建的路由与已存在的其他路由产生冲突。
	INVALIDPARAMETER_CIDRCONFLICTWITHOTHERROUTE = "InvalidParameter.CidrConflictWithOtherRoute"

	// CIDR和vpc的CIDR冲突。
	INVALIDPARAMETER_CIDRCONFLICTWITHVPCCIDR = "InvalidParameter.CidrConflictWithVpcCidr"

	// 创建的路由与VPC下已存在的全局路由产生冲突。
	INVALIDPARAMETER_CIDRCONFLICTWITHVPCGLOBALROUTE = "InvalidParameter.CidrConflictWithVpcGlobalRoute"

	// 参数错误，CIDR不符合规范。
	INVALIDPARAMETER_CIDRINVALID = "InvalidParameter.CidrInvalid"

	// CIDR不在路由表之内。
	INVALIDPARAMETER_CIDROUTOFROUTETABLE = "InvalidParameter.CidrOutOfRouteTable"

	// 集群ID不存在。
	INVALIDPARAMETER_CLUSTERNOTFOUND = "InvalidParameter.ClusterNotFound"

	// 下一跳地址已关联CIDR。
	INVALIDPARAMETER_GATEWAYALREADYASSOCIATEDCIDR = "InvalidParameter.GatewayAlreadyAssociatedCidr"

	// 无效的私有CIDR网段。
	INVALIDPARAMETER_INVALIDPRIVATENETWORKCIDR = "InvalidParameter.InvalidPrivateNetworkCIDR"

	// 参数错误。
	INVALIDPARAMETER_PARAM = "InvalidParameter.Param"

	// Prometheus未关联本集群。
	INVALIDPARAMETER_PROMCLUSTERNOTFOUND = "InvalidParameter.PromClusterNotFound"

	// Prometheus实例不存在。
	INVALIDPARAMETER_PROMINSTANCENOTFOUND = "InvalidParameter.PromInstanceNotFound"

	// 资源未找到。
	INVALIDPARAMETER_RESOURCENOTFOUND = "InvalidParameter.ResourceNotFound"

	// 路由表非空。
	INVALIDPARAMETER_ROUTETABLENOTEMPTY = "InvalidParameter.RouteTableNotEmpty"

	// 超过配额限制。
	LIMITEXCEEDED = "LimitExceeded"

	// 缺少参数错误。
	MISSINGPARAMETER = "MissingParameter"

	// 操作被拒绝。
	OPERATIONDENIED = "OperationDenied"

	// 集群处于删除保护中，禁止删除。
	OPERATIONDENIED_CLUSTERINDELETIONPROTECTION = "OperationDenied.ClusterInDeletionProtection"

	// 资源被占用。
	RESOURCEINUSE = "ResourceInUse"

	// 资源不足。
	RESOURCEINSUFFICIENT = "ResourceInsufficient"

	// 资源不存在。
	RESOURCENOTFOUND = "ResourceNotFound"

	// 伸缩组不存在。
	RESOURCENOTFOUND_ASASGNOTEXIST = "ResourceNotFound.AsAsgNotExist"

	// 集群不存在。
	RESOURCENOTFOUND_CLUSTERNOTFOUND = "ResourceNotFound.ClusterNotFound"

	// 用户Kubernetes集群中未找到指定资源。
	RESOURCENOTFOUND_KUBERESOURCENOTFOUND = "ResourceNotFound.KubeResourceNotFound"

	// 未找到该kubernetes资源。
	RESOURCENOTFOUND_KUBERNETESRESOURCENOTFOUND = "ResourceNotFound.KubernetesResourceNotFound"

	// 找不到对应路由表。
	RESOURCENOTFOUND_ROUTETABLENOTFOUND = "ResourceNotFound.RouteTableNotFound"

	// 资源不可用。
	RESOURCEUNAVAILABLE = "ResourceUnavailable"

	// 集群状态不支持该操作。
	RESOURCEUNAVAILABLE_CLUSTERSTATE = "ResourceUnavailable.ClusterState"

	// Eks Container Instance状态不支持改操作。
	RESOURCEUNAVAILABLE_EKSCONTAINERSTATUS = "ResourceUnavailable.EksContainerStatus"

	// 资源售罄。
	RESOURCESSOLDOUT = "ResourcesSoldOut"

	// 未授权操作。
	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"

	// 无该接口CAM权限。
	UNAUTHORIZEDOPERATION_CAMNOAUTH = "UnauthorizedOperation.CamNoAuth"

	// 未知参数错误。
	UNKNOWNPARAMETER = "UnknownParameter"

	// 操作不支持。
	UNSUPPORTEDOPERATION = "UnsupportedOperation"

	// AS伸缩关闭导致无法开启CA。
	UNSUPPORTEDOPERATION_CAENABLEFAILED = "UnsupportedOperation.CaEnableFailed"

	// 非白名单用户。
	UNSUPPORTEDOPERATION_NOTINWHITELIST = "UnsupportedOperation.NotInWhitelist"
)
