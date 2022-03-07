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

const (
	// 此产品的特有错误码

	// 该请求账户未通过资格审计。
	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"

	// 角色名鉴权失败
	AUTHFAILURE_CAMROLENAMEAUTHENTICATEFAILED = "AuthFailure.CamRoleNameAuthenticateFailed"

	// 弹性网卡不允许跨子网操作。
	ENINOTALLOWEDCHANGESUBNET = "EniNotAllowedChangeSubnet"

	// 账号已经存在
	FAILEDOPERATION_ACCOUNTALREADYEXISTS = "FailedOperation.AccountAlreadyExists"

	// 账号为当前用户
	FAILEDOPERATION_ACCOUNTISYOURSELF = "FailedOperation.AccountIsYourSelf"

	// 未找到指定的容灾组
	FAILEDOPERATION_DISASTERRECOVERGROUPNOTFOUND = "FailedOperation.DisasterRecoverGroupNotFound"

	// 标签键存在不合法字符
	FAILEDOPERATION_ILLEGALTAGKEY = "FailedOperation.IllegalTagKey"

	// 标签值存在不合法字符。
	FAILEDOPERATION_ILLEGALTAGVALUE = "FailedOperation.IllegalTagValue"

	// 询价失败
	FAILEDOPERATION_INQUIRYPRICEFAILED = "FailedOperation.InquiryPriceFailed"

	// 查询退换价格失败，找不到付款订单，请检查设备 `ins-xxxxxxx` 是否已过期。
	FAILEDOPERATION_INQUIRYREFUNDPRICEFAILED = "FailedOperation.InquiryRefundPriceFailed"

	// 请求不支持`EMR`的实例`ins-xxxxxxxx`。
	FAILEDOPERATION_INVALIDINSTANCEAPPLICATIONROLEEMR = "FailedOperation.InvalidInstanceApplicationRoleEmr"

	// 子网可用IP已耗尽。
	FAILEDOPERATION_NOAVAILABLEIPADDRESSCOUNTINSUBNET = "FailedOperation.NoAvailableIpAddressCountInSubnet"

	// 当前实例没有弹性IP
	FAILEDOPERATION_NOTFOUNDEIP = "FailedOperation.NotFoundEIP"

	// 账号为协作者，请填写主账号
	FAILEDOPERATION_NOTMASTERACCOUNT = "FailedOperation.NotMasterAccount"

	// 指定的置放群组非空。
	FAILEDOPERATION_PLACEMENTSETNOTEMPTY = "FailedOperation.PlacementSetNotEmpty"

	// 促销期内购买的实例不允许调整配置或计费模式。
	FAILEDOPERATION_PROMOTIONALPERIORESTRICTION = "FailedOperation.PromotionalPerioRestriction"

	// 镜像共享失败。
	FAILEDOPERATION_QIMAGESHAREFAILED = "FailedOperation.QImageShareFailed"

	// 镜像共享失败。
	FAILEDOPERATION_RIMAGESHAREFAILED = "FailedOperation.RImageShareFailed"

	// 安全组操作失败。
	FAILEDOPERATION_SECURITYGROUPACTIONFAILED = "FailedOperation.SecurityGroupActionFailed"

	// 快照容量大于磁盘大小，请选用更大的磁盘空间。
	FAILEDOPERATION_SNAPSHOTSIZELARGERTHANDATASIZE = "FailedOperation.SnapshotSizeLargerThanDataSize"

	// 不支持快照size小于云盘size。
	FAILEDOPERATION_SNAPSHOTSIZELESSTHANDATASIZE = "FailedOperation.SnapshotSizeLessThanDataSize"

	// 请求中指定的标签键为系统预留标签，禁止创建
	FAILEDOPERATION_TAGKEYRESERVED = "FailedOperation.TagKeyReserved"

	// 镜像是公共镜像并且启用了自动化助手服务，但它不符合 Linux&x86_64。
	FAILEDOPERATION_TATAGENTNOTSUPPORT = "FailedOperation.TatAgentNotSupport"

	// 实例无法退还。
	FAILEDOPERATION_UNRETURNABLE = "FailedOperation.Unreturnable"

	// 镜像配额超过了限制。
	IMAGEQUOTALIMITEXCEEDED = "ImageQuotaLimitExceeded"

	// 表示当前创建的实例个数超过了该账户允许购买的剩余配额数。
	INSTANCESQUOTALIMITEXCEEDED = "InstancesQuotaLimitExceeded"

	// 内部错误。
	INTERNALERROR = "InternalError"

	// 内部错误
	INTERNALERROR_TRADEUNKNOWNERROR = "InternalError.TradeUnknownError"

	// 操作内部错误。
	INTERNALSERVERERROR = "InternalServerError"

	// 账户余额不足。
	INVALIDACCOUNT_INSUFFICIENTBALANCE = "InvalidAccount.InsufficientBalance"

	// 账户有未支付订单。
	INVALIDACCOUNT_UNPAIDORDER = "InvalidAccount.UnpaidOrder"

	// 无效的账户Id。
	INVALIDACCOUNTID_NOTFOUND = "InvalidAccountId.NotFound"

	// 您无法共享镜像给自己。
	INVALIDACCOUNTIS_YOURSELF = "InvalidAccountIs.YourSelf"

	// 指定的ClientToken字符串长度超出限制，必须小于等于64字节。
	INVALIDCLIENTTOKEN_TOOLONG = "InvalidClientToken.TooLong"

	// 无效的过滤器。
	INVALIDFILTER = "InvalidFilter"

	// [`Filter`](/document/api/213/15753#Filter)。
	INVALIDFILTERVALUE_LIMITEXCEEDED = "InvalidFilterValue.LimitExceeded"

	// 不支持该宿主机实例执行指定的操作。
	INVALIDHOST_NOTSUPPORTED = "InvalidHost.NotSupported"

	// 无效[CDH](https://cloud.tencent.com/document/product/416) `ID`。指定的[CDH](https://cloud.tencent.com/document/product/416) `ID`格式错误。例如`ID`长度错误`host-1122`。
	INVALIDHOSTID_MALFORMED = "InvalidHostId.Malformed"

	// 指定的HostId不存在，或不属于该请求账号所有。
	INVALIDHOSTID_NOTFOUND = "InvalidHostId.NotFound"

	// 镜像处于共享中。
	INVALIDIMAGEID_INSHARED = "InvalidImageId.InShared"

	// 镜像状态不合法。
	INVALIDIMAGEID_INCORRECTSTATE = "InvalidImageId.IncorrectState"

	// 错误的镜像Id格式。
	INVALIDIMAGEID_MALFORMED = "InvalidImageId.Malformed"

	// 未找到该镜像。
	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"

	// 镜像大小超过限制。
	INVALIDIMAGEID_TOOLARGE = "InvalidImageId.TooLarge"

	// 镜像名称与原有镜像重复。
	INVALIDIMAGENAME_DUPLICATE = "InvalidImageName.Duplicate"

	// 不支持的操作系统类型。
	INVALIDIMAGEOSTYPE_UNSUPPORTED = "InvalidImageOsType.Unsupported"

	// 不支持的操作系统版本。
	INVALIDIMAGEOSVERSION_UNSUPPORTED = "InvalidImageOsVersion.Unsupported"

	// 不被支持的实例。
	INVALIDINSTANCE_NOTSUPPORTED = "InvalidInstance.NotSupported"

	// 无效实例`ID`。指定的实例`ID`格式错误。例如实例`ID`长度错误`ins-1122`。
	INVALIDINSTANCEID_MALFORMED = "InvalidInstanceId.Malformed"

	// 没有找到相应实例。
	INVALIDINSTANCEID_NOTFOUND = "InvalidInstanceId.NotFound"

	// 指定的InstanceName字符串长度超出限制，必须小于等于60字节。
	INVALIDINSTANCENAME_TOOLONG = "InvalidInstanceName.TooLong"

	// 该实例不满足包月[退还规则](https://cloud.tencent.com/document/product/213/9711)。
	INVALIDINSTANCENOTSUPPORTEDPREPAIDINSTANCE = "InvalidInstanceNotSupportedPrepaidInstance"

	// 指定实例的当前状态不能进行该操作。
	INVALIDINSTANCESTATE = "InvalidInstanceState"

	// 指定InstanceType参数格式不合法。
	INVALIDINSTANCETYPE_MALFORMED = "InvalidInstanceType.Malformed"

	// 密钥对数量超过限制。
	INVALIDKEYPAIR_LIMITEXCEEDED = "InvalidKeyPair.LimitExceeded"

	// 无效密钥对ID。指定的密钥对ID格式错误，例如 `ID` 长度错误`skey-1122`。
	INVALIDKEYPAIRID_MALFORMED = "InvalidKeyPairId.Malformed"

	// 无效密钥对ID。指定的密钥对ID不存在。
	INVALIDKEYPAIRID_NOTFOUND = "InvalidKeyPairId.NotFound"

	// 密钥对名称重复。
	INVALIDKEYPAIRNAME_DUPLICATE = "InvalidKeyPairName.Duplicate"

	// 密钥名称为空。
	INVALIDKEYPAIRNAMEEMPTY = "InvalidKeyPairNameEmpty"

	// 密钥名称包含非法字符。密钥名称只支持英文、数字和下划线。
	INVALIDKEYPAIRNAMEINCLUDEILLEGALCHAR = "InvalidKeyPairNameIncludeIllegalChar"

	// 密钥名称超过25个字符。
	INVALIDKEYPAIRNAMETOOLONG = "InvalidKeyPairNameTooLong"

	// 参数错误。
	INVALIDPARAMETER = "InvalidParameter"

	// DataDiskIds不应该传入RootDisk的Id。
	INVALIDPARAMETER_DATADISKIDCONTAINSROOTDISK = "InvalidParameter.DataDiskIdContainsRootDisk"

	// 指定的数据盘不属于指定的实例。
	INVALIDPARAMETER_DATADISKNOTBELONGSPECIFIEDINSTANCE = "InvalidParameter.DataDiskNotBelongSpecifiedInstance"

	// 只能包含一个系统盘快照。
	INVALIDPARAMETER_DUPLICATESYSTEMSNAPSHOTS = "InvalidParameter.DuplicateSystemSnapshots"

	// 该主机当前状态不支持该操作。
	INVALIDPARAMETER_HOSTIDSTATUSNOTSUPPORT = "InvalidParameter.HostIdStatusNotSupport"

	// 指定的hostName不符合规范。
	INVALIDPARAMETER_HOSTNAMEILLEGAL = "InvalidParameter.HostNameIllegal"

	// 当前接口不支持实例镜像。
	INVALIDPARAMETER_INSTANCEIMAGENOTSUPPORT = "InvalidParameter.InstanceImageNotSupport"

	// 云盘资源售罄。
	INVALIDPARAMETER_INVALIDCLOUDDISKSOLDOUT = "InvalidParameter.InvalidCloudDiskSoldOut"

	// 参数依赖不正确。
	INVALIDPARAMETER_INVALIDDEPENDENCE = "InvalidParameter.InvalidDependence"

	// 当前操作不支持该类型实例。
	INVALIDPARAMETER_INVALIDINSTANCENOTSUPPORTED = "InvalidParameter.InvalidInstanceNotSupported"

	// 指定的私有网络ip格式不正确。
	INVALIDPARAMETER_INVALIDIPFORMAT = "InvalidParameter.InvalidIpFormat"

	// 不能同时指定ImageIds和Filters。
	INVALIDPARAMETER_INVALIDPARAMETERCOEXISTIMAGEIDSFILTERS = "InvalidParameter.InvalidParameterCoexistImageIdsFilters"

	// 错误的url地址。
	INVALIDPARAMETER_INVALIDPARAMETERURLERROR = "InvalidParameter.InvalidParameterUrlError"

	// CoreCount和ThreadPerCore必须同时提供。
	INVALIDPARAMETER_LACKCORECOUNTORTHREADPERCORE = "InvalidParameter.LackCoreCountOrThreadPerCore"

	// 本地数据盘不支持创建实例镜像。
	INVALIDPARAMETER_LOCALDATADISKNOTSUPPORT = "InvalidParameter.LocalDataDiskNotSupport"

	// 不支持同时指定密钥登陆和保持镜像登陆方式。
	INVALIDPARAMETER_PARAMETERCONFLICT = "InvalidParameter.ParameterConflict"

	// 不支持设置登陆密码。
	INVALIDPARAMETER_PASSWORDNOTSUPPORTED = "InvalidParameter.PasswordNotSupported"

	// 指定的快照不存在。
	INVALIDPARAMETER_SNAPSHOTNOTFOUND = "InvalidParameter.SnapshotNotFound"

	// 多选一必选参数缺失。
	INVALIDPARAMETER_SPECIFYONEPARAMETER = "InvalidParameter.SpecifyOneParameter"

	// 不支持Swap盘。
	INVALIDPARAMETER_SWAPDISKNOTSUPPORT = "InvalidParameter.SwapDiskNotSupport"

	// 参数不包含系统盘快照。
	INVALIDPARAMETER_SYSTEMSNAPSHOTNOTFOUND = "InvalidParameter.SystemSnapshotNotFound"

	// 参数长度超过限制。
	INVALIDPARAMETER_VALUETOOLARGE = "InvalidParameter.ValueTooLarge"

	// 表示参数组合不正确。
	INVALIDPARAMETERCOMBINATION = "InvalidParameterCombination"

	// 指定的两个参数冲突，不能同时存在。 EIP只能绑定在实例上或指定网卡的指定内网 IP 上。
	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"

	// 参数取值错误。
	INVALIDPARAMETERVALUE = "InvalidParameterValue"

	// 共享带宽包ID不合要求，请提供规范的共享带宽包ID，类似bwp-xxxxxxxx，字母x代表小写字符或者数字。
	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDMALFORMED = "InvalidParameterValue.BandwidthPackageIdMalformed"

	// 请确认指定的带宽包是否存在。
	INVALIDPARAMETERVALUE_BANDWIDTHPACKAGEIDNOTFOUND = "InvalidParameterValue.BandwidthPackageIdNotFound"

	// 找不到对应的CHC物理服务器。
	INVALIDPARAMETERVALUE_CHCHOSTSNOTFOUND = "InvalidParameterValue.ChcHostsNotFound"

	// SSD云硬盘为数据盘时，购买大小不得小于100GB
	INVALIDPARAMETERVALUE_CLOUDSSDDATADISKSIZETOOSMALL = "InvalidParameterValue.CloudSsdDataDiskSizeTooSmall"

	// 核心计数不合法。
	INVALIDPARAMETERVALUE_CORECOUNTVALUE = "InvalidParameterValue.CoreCountValue"

	// 参数值重复。
	INVALIDPARAMETERVALUE_DUPLICATE = "InvalidParameterValue.Duplicate"

	// 非GPU实例不允许转为GPU实例。
	INVALIDPARAMETERVALUE_GPUINSTANCEFAMILY = "InvalidParameterValue.GPUInstanceFamily"

	// IP格式非法。
	INVALIDPARAMETERVALUE_IPADDRESSMALFORMED = "InvalidParameterValue.IPAddressMalformed"

	// ipv6地址无效
	INVALIDPARAMETERVALUE_IPV6ADDRESSMALFORMED = "InvalidParameterValue.IPv6AddressMalformed"

	// HostName参数值不合法
	INVALIDPARAMETERVALUE_ILLEGALHOSTNAME = "InvalidParameterValue.IllegalHostName"

	// 传参格式不对。
	INVALIDPARAMETERVALUE_INCORRECTFORMAT = "InvalidParameterValue.IncorrectFormat"

	// 不支持操作不同计费方式的实例。
	INVALIDPARAMETERVALUE_INSTANCENOTSUPPORTEDMIXPRICINGMODEL = "InvalidParameterValue.InstanceNotSupportedMixPricingModel"

	// 指定机型不存在
	INVALIDPARAMETERVALUE_INSTANCETYPENOTFOUND = "InvalidParameterValue.InstanceTypeNotFound"

	// 实例类型不可加入高性能计算集群。
	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTHPCCLUSTER = "InvalidParameterValue.InstanceTypeNotSupportHpcCluster"

	// 高性能计算实例需指定对应的高性能计算集群。
	INVALIDPARAMETERVALUE_INSTANCETYPEREQUIREDHPCCLUSTER = "InvalidParameterValue.InstanceTypeRequiredHpcCluster"

	// 竞价数量不足。
	INVALIDPARAMETERVALUE_INSUFFICIENTOFFERING = "InvalidParameterValue.InsufficientOffering"

	// 竞价失败。
	INVALIDPARAMETERVALUE_INSUFFICIENTPRICE = "InvalidParameterValue.InsufficientPrice"

	// 无效的appid。
	INVALIDPARAMETERVALUE_INVALIDAPPIDFORMAT = "InvalidParameterValue.InvalidAppIdFormat"

	// 镜像ID不支持指定的实例机型。
	INVALIDPARAMETERVALUE_INVALIDIMAGEFORGIVENINSTANCETYPE = "InvalidParameterValue.InvalidImageForGivenInstanceType"

	// 当前镜像为RAW格式，无法创建CVM，建议您选择其他镜像。
	INVALIDPARAMETERVALUE_INVALIDIMAGEFORMAT = "InvalidParameterValue.InvalidImageFormat"

	// 镜像不允许执行该操作
	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"

	// 当前地域不支持指定镜像所包含的操作系统。
	INVALIDPARAMETERVALUE_INVALIDIMAGEOSNAME = "InvalidParameterValue.InvalidImageOsName"

	// 镜像被其他操作占用，请检查，并稍后重试。
	INVALIDPARAMETERVALUE_INVALIDIMAGESTATE = "InvalidParameterValue.InvalidImageState"

	// IP地址不符合规范
	INVALIDPARAMETERVALUE_INVALIDIPFORMAT = "InvalidParameterValue.InvalidIpFormat"

	// 实例启动模板描述格式错误。
	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATEDESCRIPTION = "InvalidParameterValue.InvalidLaunchTemplateDescription"

	// 实例启动模板名称格式错误。
	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATENAME = "InvalidParameterValue.InvalidLaunchTemplateName"

	// 实例启动模板描述格式错误。
	INVALIDPARAMETERVALUE_INVALIDLAUNCHTEMPLATEVERSIONDESCRIPTION = "InvalidParameterValue.InvalidLaunchTemplateVersionDescription"

	// 参数值错误。
	INVALIDPARAMETERVALUE_INVALIDPARAMETERVALUELIMIT = "InvalidParameterValue.InvalidParameterValueLimit"

	// 无效密码。指定的密码不符合密码复杂度限制。例如密码长度不符合限制等。
	INVALIDPARAMETERVALUE_INVALIDPASSWORD = "InvalidParameterValue.InvalidPassword"

	// 时间格式不合法。
	INVALIDPARAMETERVALUE_INVALIDTIMEFORMAT = "InvalidParameterValue.InvalidTimeFormat"

	// UserData格式错误, 需要base64编码格式
	INVALIDPARAMETERVALUE_INVALIDUSERDATAFORMAT = "InvalidParameterValue.InvalidUserDataFormat"

	// 无效的模糊查询字符串。
	INVALIDPARAMETERVALUE_INVALIDVAGUENAME = "InvalidParameterValue.InvalidVagueName"

	// 请确认密钥是否存在。
	INVALIDPARAMETERVALUE_KEYPAIRNOTFOUND = "InvalidParameterValue.KeyPairNotFound"

	// 指定的密钥不支持当前操作。
	INVALIDPARAMETERVALUE_KEYPAIRNOTSUPPORTED = "InvalidParameterValue.KeyPairNotSupported"

	// 不支持删除默认启动模板版本。
	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEDEFAULTVERSION = "InvalidParameterValue.LaunchTemplateDefaultVersion"

	// 实例启动模板ID格式错误。
	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDMALFORMED = "InvalidParameterValue.LaunchTemplateIdMalformed"

	// 实例启动模板ID不存在。
	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdNotExisted"

	// 实例启动模板和版本ID组合不存在。
	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERNOTEXISTED = "InvalidParameterValue.LaunchTemplateIdVerNotExisted"

	// 指定的实例启动模板id不存在。
	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEIDVERSETALREADY = "InvalidParameterValue.LaunchTemplateIdVerSetAlready"

	// 实例启动模板未找到。
	INVALIDPARAMETERVALUE_LAUNCHTEMPLATENOTFOUND = "InvalidParameterValue.LaunchTemplateNotFound"

	// 无效的实例启动模板版本号。
	INVALIDPARAMETERVALUE_LAUNCHTEMPLATEVERSION = "InvalidParameterValue.LaunchTemplateVersion"

	// 参数值数量超过限制。
	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"

	// 本地盘的限制范围。
	INVALIDPARAMETERVALUE_LOCALDISKSIZERANGE = "InvalidParameterValue.LocalDiskSizeRange"

	// 参数值必须为开启DHCP的VPC
	INVALIDPARAMETERVALUE_MUSTDHCPENABLEDVPC = "InvalidParameterValue.MustDhcpEnabledVpc"

	// 不支持的操作。
	INVALIDPARAMETERVALUE_NOTSUPPORTED = "InvalidParameterValue.NotSupported"

	//  无效参数值。参数值取值范围不合法。
	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"

	// 快照ID不合要求，请提供规范的快照ID，类似snap-xxxxxxxx，字母x代表小写字符或者数字
	INVALIDPARAMETERVALUE_SNAPSHOTIDMALFORMED = "InvalidParameterValue.SnapshotIdMalformed"

	// 子网ID不合要求，请提供规范的子网ID，类似subnet-xxxxxxxx，字母x代表小写字符或者数字
	INVALIDPARAMETERVALUE_SUBNETIDMALFORMED = "InvalidParameterValue.SubnetIdMalformed"

	// 创建失败，您指定的子网不存在，请您重新指定
	INVALIDPARAMETERVALUE_SUBNETNOTEXIST = "InvalidParameterValue.SubnetNotExist"

	// 指定的标签不存在。
	INVALIDPARAMETERVALUE_TAGKEYNOTFOUND = "InvalidParameterValue.TagKeyNotFound"

	// 每核心线程数不合法。
	INVALIDPARAMETERVALUE_THREADPERCOREVALUE = "InvalidParameterValue.ThreadPerCoreValue"

	// 参数值超过最大限制。
	INVALIDPARAMETERVALUE_TOOLARGE = "InvalidParameterValue.TooLarge"

	// 无效参数值。参数值太长。
	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"

	// VPC ID`xxx`不合要求，请提供规范的Vpc ID， 类似vpc-xxxxxxxx，字母x代表小写字符或者数字。
	INVALIDPARAMETERVALUE_VPCIDMALFORMED = "InvalidParameterValue.VpcIdMalformed"

	// 指定的VpcId不存在。
	INVALIDPARAMETERVALUE_VPCIDNOTEXIST = "InvalidParameterValue.VpcIdNotExist"

	// VPC网络与实例不在同一可用区
	INVALIDPARAMETERVALUE_VPCIDZONEIDNOTMATCH = "InvalidParameterValue.VpcIdZoneIdNotMatch"

	// 请求不支持该可用区
	INVALIDPARAMETERVALUE_ZONENOTSUPPORTED = "InvalidParameterValue.ZoneNotSupported"

	// 参数值数量超过限制。
	INVALIDPARAMETERVALUELIMIT = "InvalidParameterValueLimit"

	// 无效参数值。指定的 `Offset` 无效。
	INVALIDPARAMETERVALUEOFFSET = "InvalidParameterValueOffset"

	// 无效密码。指定的密码不符合密码复杂度限制。例如密码长度不符合限制等。
	INVALIDPASSWORD = "InvalidPassword"

	// 无效时长。目前只支持时长：[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36]，单位：月。
	INVALIDPERIOD = "InvalidPeriod"

	// 账户不支持该操作。
	INVALIDPERMISSION = "InvalidPermission"

	// 无效的项目ID，指定的项目ID不存在。
	INVALIDPROJECTID_NOTFOUND = "InvalidProjectId.NotFound"

	// 无效密钥公钥。指定公钥已经存在。
	INVALIDPUBLICKEY_DUPLICATE = "InvalidPublicKey.Duplicate"

	// 无效密钥公钥。指定公钥格式错误，不符合`OpenSSH RSA`格式要求。
	INVALIDPUBLICKEY_MALFORMED = "InvalidPublicKey.Malformed"

	// 未找到该区域。
	INVALIDREGION_NOTFOUND = "InvalidRegion.NotFound"

	// 该区域目前不支持同步镜像。
	INVALIDREGION_UNAVAILABLE = "InvalidRegion.Unavailable"

	// 指定的`安全组ID`不存在。
	INVALIDSECURITYGROUPID_NOTFOUND = "InvalidSecurityGroupId.NotFound"

	// 指定的`安全组ID`格式错误，例如`实例ID`长度错误`sg-ide32`。
	INVALIDSGID_MALFORMED = "InvalidSgId.Malformed"

	// 指定的`zone`不存在。
	INVALIDZONE_MISMATCHREGION = "InvalidZone.MismatchRegion"

	// 一个实例绑定安全组数量不能超过5个
	LIMITEXCEEDED_ASSOCIATEUSGLIMITEXCEEDED = "LimitExceeded.AssociateUSGLimitExceeded"

	// 安全组关联云主机弹性网卡配额超限。
	LIMITEXCEEDED_CVMSVIFSPERSECGROUPLIMITEXCEEDED = "LimitExceeded.CvmsVifsPerSecGroupLimitExceeded"

	// 指定置放群组配额不足。
	LIMITEXCEEDED_DISASTERRECOVERGROUP = "LimitExceeded.DisasterRecoverGroup"

	// 特定实例当前ENI数量已超过目标实例类型的ENI允许的最大值，需删除部分ENI后重试。
	LIMITEXCEEDED_ENINUMLIMIT = "LimitExceeded.EniNumLimit"

	// IP数量超过网卡上限。
	LIMITEXCEEDED_IPV6ADDRESSNUM = "LimitExceeded.IPv6AddressNum"

	// 当前配额不足够生产指定数量的实例
	LIMITEXCEEDED_INSTANCEQUOTA = "LimitExceeded.InstanceQuota"

	// 目标实例规格不支持当前规格的外网带宽上限，不支持调整。具体可参考[公网网络带宽上限](https://cloud.tencent.com/document/product/213/12523)。
	LIMITEXCEEDED_INSTANCETYPEBANDWIDTH = "LimitExceeded.InstanceTypeBandwidth"

	// 实例启动模板数量超限。
	LIMITEXCEEDED_LAUNCHTEMPLATEQUOTA = "LimitExceeded.LaunchTemplateQuota"

	// 实例启动模板版本数量超限。
	LIMITEXCEEDED_LAUNCHTEMPLATEVERSIONQUOTA = "LimitExceeded.LaunchTemplateVersionQuota"

	// 预付费实例已购买数量已达到最大配额，请提升配额后重试。
	LIMITEXCEEDED_PREPAYQUOTA = "LimitExceeded.PrepayQuota"

	// 安全组限额不足
	LIMITEXCEEDED_SINGLEUSGQUOTA = "LimitExceeded.SingleUSGQuota"

	// 竞价实例类型配额不足
	LIMITEXCEEDED_SPOTQUOTA = "LimitExceeded.SpotQuota"

	// 退还失败，退还配额已达上限。
	LIMITEXCEEDED_USERRETURNQUOTA = "LimitExceeded.UserReturnQuota"

	// 竞价实例配额不足
	LIMITEXCEEDED_USERSPOTQUOTA = "LimitExceeded.UserSpotQuota"

	// 子网IP不足
	LIMITEXCEEDED_VPCSUBNETNUM = "LimitExceeded.VpcSubnetNum"

	// 缺少参数错误。
	MISSINGPARAMETER = "MissingParameter"

	// 缺少必要参数，请至少提供一个参数。
	MISSINGPARAMETER_ATLEASTONE = "MissingParameter.AtLeastOne"

	// DPDK实例机型要求VPC网络
	MISSINGPARAMETER_DPDKINSTANCETYPEREQUIREDVPC = "MissingParameter.DPDKInstanceTypeRequiredVPC"

	// 该实例类型必须开启云监控服务
	MISSINGPARAMETER_MONITORSERVICE = "MissingParameter.MonitorService"

	// 同样的任务正在运行。
	MUTEXOPERATION_TASKRUNNING = "MutexOperation.TaskRunning"

	// 不允许未配置部署网络的CHC安装云上镜像。
	OPERATIONDENIED_CHCINSTALLCLOUDIMAGEWITHOUTDEPLOYNETWORK = "OperationDenied.ChcInstallCloudImageWithoutDeployNetwork"

	// 实例正在执行其他操作，请稍后再试。
	OPERATIONDENIED_INSTANCEOPERATIONINPROGRESS = "OperationDenied.InstanceOperationInProgress"

	// 镜像共享超过配额。
	OVERQUOTA = "OverQuota"

	// 该地域不支持导入镜像。
	REGIONABILITYLIMIT_UNSUPPORTEDTOIMPORTIMAGE = "RegionAbilityLimit.UnsupportedToImportImage"

	// 资源被占用。
	RESOURCEINUSE = "ResourceInUse"

	// 该可用区已售罄
	RESOURCEINSUFFICIENT_AVAILABILITYZONESOLDOUT = "ResourceInsufficient.AvailabilityZoneSoldOut"

	// 指定的云盘规格已售罄
	RESOURCEINSUFFICIENT_CLOUDDISKSOLDOUT = "ResourceInsufficient.CloudDiskSoldOut"

	// 云盘参数不符合规范
	RESOURCEINSUFFICIENT_CLOUDDISKUNAVAILABLE = "ResourceInsufficient.CloudDiskUnavailable"

	// 实例个数超过容灾组的配额
	RESOURCEINSUFFICIENT_DISASTERRECOVERGROUPCVMQUOTA = "ResourceInsufficient.DisasterRecoverGroupCvmQuota"

	// 指定的实例类型库存不足。
	RESOURCEINSUFFICIENT_SPECIFIEDINSTANCETYPE = "ResourceInsufficient.SpecifiedInstanceType"

	// 指定的实例类型在选择的可用区已售罄。
	RESOURCEINSUFFICIENT_ZONESOLDOUTFORSPECIFIEDINSTANCE = "ResourceInsufficient.ZoneSoldOutForSpecifiedInstance"

	// 高性能计算集群不存在。
	RESOURCENOTFOUND_HPCCLUSTER = "ResourceNotFound.HpcCluster"

	// 指定的置放群组不存在。
	RESOURCENOTFOUND_INVALIDPLACEMENTSET = "ResourceNotFound.InvalidPlacementSet"

	// 无可用的缺省类型的CBS资源。
	RESOURCENOTFOUND_NODEFAULTCBS = "ResourceNotFound.NoDefaultCbs"

	// 无可用的缺省类型的CBS资源。
	RESOURCENOTFOUND_NODEFAULTCBSWITHREASON = "ResourceNotFound.NoDefaultCbsWithReason"

	// 该可用区不售卖此机型
	RESOURCEUNAVAILABLE_INSTANCETYPE = "ResourceUnavailable.InstanceType"

	// 快照正在创建过程中。
	RESOURCEUNAVAILABLE_SNAPSHOTCREATING = "ResourceUnavailable.SnapshotCreating"

	// 该可用区已售罄
	RESOURCESSOLDOUT_AVAILABLEZONE = "ResourcesSoldOut.AvailableZone"

	// 公网IP已售罄。
	RESOURCESSOLDOUT_EIPINSUFFICIENT = "ResourcesSoldOut.EipInsufficient"

	// 指定的实例类型已售罄。
	RESOURCESSOLDOUT_SPECIFIEDINSTANCETYPE = "ResourcesSoldOut.SpecifiedInstanceType"

	// 安全组服务接口调用通用错误。
	SECGROUPACTIONFAILURE = "SecGroupActionFailure"

	// 未授权操作。
	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"

	// 指定的镜像不属于用户。
	UNAUTHORIZEDOPERATION_IMAGENOTBELONGTOACCOUNT = "UnauthorizedOperation.ImageNotBelongToAccount"

	// 请确认Token是否有效。
	UNAUTHORIZEDOPERATION_INVALIDTOKEN = "UnauthorizedOperation.InvalidToken"

	// 您无法进行当前操作，请确认多因子认证（MFA）是否失效。
	UNAUTHORIZEDOPERATION_MFAEXPIRED = "UnauthorizedOperation.MFAExpired"

	// 没有权限进行此操作，请确认是否存在多因子认证（MFA）。
	UNAUTHORIZEDOPERATION_MFANOTFOUND = "UnauthorizedOperation.MFANotFound"

	// 无权操作指定的资源，请正确配置CAM策略。
	UNAUTHORIZEDOPERATION_PERMISSIONDENIED = "UnauthorizedOperation.PermissionDenied"

	// 未知参数错误。
	UNKNOWNPARAMETER = "UnknownParameter"

	// 操作不支持。
	UNSUPPORTEDOPERATION = "UnsupportedOperation"

	// 指定的实例付费模式或者网络付费模式不支持共享带宽包
	UNSUPPORTEDOPERATION_BANDWIDTHPACKAGEIDNOTSUPPORTED = "UnsupportedOperation.BandwidthPackageIdNotSupported"

	// IPv6实例不支持VPC迁移
	UNSUPPORTEDOPERATION_IPV6NOTSUPPORTVPCMIGRATE = "UnsupportedOperation.IPv6NotSupportVpcMigrate"

	// 请求不支持该实例计费模式
	UNSUPPORTEDOPERATION_INSTANCECHARGETYPE = "UnsupportedOperation.InstanceChargeType"

	// 不支持混合付费模式。
	UNSUPPORTEDOPERATION_INSTANCEMIXEDPRICINGMODEL = "UnsupportedOperation.InstanceMixedPricingModel"

	// 请求不支持操作系统为`Xserver windows2012cndatacenterx86_64`的实例`ins-xxxxxx` 。
	UNSUPPORTEDOPERATION_INSTANCEOSWINDOWS = "UnsupportedOperation.InstanceOsWindows"

	// 该子机处于封禁状态，请联系相关人员处理。
	UNSUPPORTEDOPERATION_INSTANCESTATEBANNING = "UnsupportedOperation.InstanceStateBanning"

	// 请求不支持永久故障的实例。
	UNSUPPORTEDOPERATION_INSTANCESTATECORRUPTED = "UnsupportedOperation.InstanceStateCorrupted"

	// 请求不支持进入救援模式的实例
	UNSUPPORTEDOPERATION_INSTANCESTATEENTERRESCUEMODE = "UnsupportedOperation.InstanceStateEnterRescueMode"

	// 不支持状态为 `ENTER_SERVICE_LIVE_MIGRATE`.的实例 `ins-xxxxxx` 。
	UNSUPPORTEDOPERATION_INSTANCESTATEENTERSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateEnterServiceLiveMigrate"

	// 请求不支持正在退出救援模式的实例
	UNSUPPORTEDOPERATION_INSTANCESTATEEXITRESCUEMODE = "UnsupportedOperation.InstanceStateExitRescueMode"

	// 不支持状态为 `EXIT_SERVICE_LIVE_MIGRATE`.的实例 `ins-xxxxxx` 。
	UNSUPPORTEDOPERATION_INSTANCESTATEEXITSERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateExitServiceLiveMigrate"

	// 操作不支持已冻结的实例。
	UNSUPPORTEDOPERATION_INSTANCESTATEFREEZING = "UnsupportedOperation.InstanceStateFreezing"

	// 请求不支持正在隔离状态的实例。
	UNSUPPORTEDOPERATION_INSTANCESTATEISOLATING = "UnsupportedOperation.InstanceStateIsolating"

	// 请求不支持创建未完成的实例
	UNSUPPORTEDOPERATION_INSTANCESTATEPENDING = "UnsupportedOperation.InstanceStatePending"

	// 请求不支持正在重启的实例
	UNSUPPORTEDOPERATION_INSTANCESTATEREBOOTING = "UnsupportedOperation.InstanceStateRebooting"

	// 请求不支持救援模式的实例
	UNSUPPORTEDOPERATION_INSTANCESTATERESCUEMODE = "UnsupportedOperation.InstanceStateRescueMode"

	// 请求不支持开机状态的实例
	UNSUPPORTEDOPERATION_INSTANCESTATERUNNING = "UnsupportedOperation.InstanceStateRunning"

	// 不支持正在服务迁移的实例，请稍后再试
	UNSUPPORTEDOPERATION_INSTANCESTATESERVICELIVEMIGRATE = "UnsupportedOperation.InstanceStateServiceLiveMigrate"

	// 请求不支持隔离状态的实例
	UNSUPPORTEDOPERATION_INSTANCESTATESHUTDOWN = "UnsupportedOperation.InstanceStateShutdown"

	// 实例开机中，不允许该操作。
	UNSUPPORTEDOPERATION_INSTANCESTATESTARTING = "UnsupportedOperation.InstanceStateStarting"

	// 请求不支持已关机的实例
	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPED = "UnsupportedOperation.InstanceStateStopped"

	// 请求不支持正在关机的实例
	UNSUPPORTEDOPERATION_INSTANCESTATESTOPPING = "UnsupportedOperation.InstanceStateStopping"

	// 不支持已销毁的实例
	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATED = "UnsupportedOperation.InstanceStateTerminated"

	// 请求不支持正在销毁的实例
	UNSUPPORTEDOPERATION_INSTANCESTATETERMINATING = "UnsupportedOperation.InstanceStateTerminating"

	// 不支持指定的磁盘
	UNSUPPORTEDOPERATION_INVALIDDISK = "UnsupportedOperation.InvalidDisk"

	// 当前操作只支持国际版用户。
	UNSUPPORTEDOPERATION_INVALIDPERMISSIONNONINTERNATIONALACCOUNT = "UnsupportedOperation.InvalidPermissionNonInternationalAccount"

	// 指定的地域不支持加密盘。
	UNSUPPORTEDOPERATION_INVALIDREGIONDISKENCRYPT = "UnsupportedOperation.InvalidRegionDiskEncrypt"

	// 密钥不支持Windows操作系统
	UNSUPPORTEDOPERATION_KEYPAIRUNSUPPORTEDWINDOWS = "UnsupportedOperation.KeyPairUnsupportedWindows"

	// 机型数据盘全为本地盘不支持跨机型调整。
	UNSUPPORTEDOPERATION_LOCALDATADISKCHANGEINSTANCEFAMILY = "UnsupportedOperation.LocalDataDiskChangeInstanceFamily"

	// 绑定负载均衡的实例，不支持修改vpc属性。
	UNSUPPORTEDOPERATION_MODIFYVPCWITHCLB = "UnsupportedOperation.ModifyVPCWithCLB"

	// 该实例类型不支持竞价计费
	UNSUPPORTEDOPERATION_NOINSTANCETYPESUPPORTSPOT = "UnsupportedOperation.NoInstanceTypeSupportSpot"

	// 当前实例不是FPGA机型。
	UNSUPPORTEDOPERATION_NOTFPGAINSTANCE = "UnsupportedOperation.NotFpgaInstance"

	// 针对当前实例设置定时任务失败。
	UNSUPPORTEDOPERATION_NOTSUPPORTIMPORTINSTANCESACTIONTIMER = "UnsupportedOperation.NotSupportImportInstancesActionTimer"

	// 操作不支持当前实例
	UNSUPPORTEDOPERATION_NOTSUPPORTINSTANCEIMAGE = "UnsupportedOperation.NotSupportInstanceImage"

	// 该操作仅支持预付费账户
	UNSUPPORTEDOPERATION_ONLYFORPREPAIDACCOUNT = "UnsupportedOperation.OnlyForPrepaidAccount"

	// 当前镜像不支持对该实例的重装操作。
	UNSUPPORTEDOPERATION_RAWLOCALDISKINSREINSTALLTOQCOW2 = "UnsupportedOperation.RawLocalDiskInsReinstalltoQcow2"

	// 不支持该地域
	UNSUPPORTEDOPERATION_REGION = "UnsupportedOperation.Region"

	// 当前用户暂不支持购买预留实例计费。
	UNSUPPORTEDOPERATION_RESERVEDINSTANCEINVISIBLEFORUSER = "UnsupportedOperation.ReservedInstanceInvisibleForUser"

	// 用户预留实例计费配额已达上限。
	UNSUPPORTEDOPERATION_RESERVEDINSTANCEOUTOFQUATA = "UnsupportedOperation.ReservedInstanceOutofQuata"

	// 请求不支持特殊机型的实例
	UNSUPPORTEDOPERATION_SPECIALINSTANCETYPE = "UnsupportedOperation.SpecialInstanceType"

	// 不支持关机不收费特性
	UNSUPPORTEDOPERATION_STOPPEDMODESTOPCHARGING = "UnsupportedOperation.StoppedModeStopCharging"

	// 该机型为包销机型，RenewFlag的值只允许设置为NOTIFY_AND_AUTO_RENEW。
	UNSUPPORTEDOPERATION_UNDERWRITINGINSTANCETYPEONLYSUPPORTAUTORENEW = "UnsupportedOperation.UnderwritingInstanceTypeOnlySupportAutoRenew"

	// 指定机型不支持跨机型调整配置。
	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILY = "UnsupportedOperation.UnsupportedChangeInstanceFamily"

	// 非ARM机型不支持调整到ARM机型。
	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILYTOARM = "UnsupportedOperation.UnsupportedChangeInstanceFamilyToARM"

	// 目标机型是SA3, 不支持变配。
	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCEFAMILYTOSA3 = "UnsupportedOperation.UnsupportedChangeInstanceFamilyToSA3"

	// 不支持实例变配到此类型机型。
	UNSUPPORTEDOPERATION_UNSUPPORTEDCHANGEINSTANCETOTHISINSTANCEFAMILY = "UnsupportedOperation.UnsupportedChangeInstanceToThisInstanceFamily"

	// 请求不支持国际版账号
	UNSUPPORTEDOPERATION_UNSUPPORTEDINTERNATIONALUSER = "UnsupportedOperation.UnsupportedInternationalUser"

	// 用户限额操作的配额不足。
	UNSUPPORTEDOPERATION_USERLIMITOPERATIONEXCEEDQUOTA = "UnsupportedOperation.UserLimitOperationExceedQuota"

	// 私有网络ip不在子网内。
	VPCADDRNOTINSUBNET = "VpcAddrNotInSubNet"

	// 私有网络ip已经被使用。
	VPCIPISUSED = "VpcIpIsUsed"
)
