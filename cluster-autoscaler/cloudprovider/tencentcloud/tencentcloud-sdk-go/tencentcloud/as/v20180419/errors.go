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

const (
	// 此产品的特有错误码

	// 该请求账户未通过资格审计。
	ACCOUNTQUALIFICATIONRESTRICTIONS = "AccountQualificationRestrictions"

	// CVM接口调用失败。
	CALLCVMERROR = "CallCvmError"

	// 未生成伸缩活动。
	FAILEDOPERATION_NOACTIVITYTOGENERATE = "FailedOperation.NoActivityToGenerate"

	// 内部错误。
	INTERNALERROR = "InternalError"

	// Cmq 接口调用失败。
	INTERNALERROR_CALLCMQERROR = "InternalError.CallCmqError"

	// 内部接口调用失败。
	INTERNALERROR_CALLERROR = "InternalError.CallError"

	// LB 接口调用失败。
	INTERNALERROR_CALLLBERROR = "InternalError.CallLbError"

	// Monitor接口调用失败。
	INTERNALERROR_CALLMONITORERROR = "InternalError.CallMonitorError"

	// 通知服务接口调用失败。
	INTERNALERROR_CALLNOTIFICATIONERROR = "InternalError.CallNotificationError"

	// STS 接口调用失败。
	INTERNALERROR_CALLSTSERROR = "InternalError.CallStsError"

	// Tag 接口调用失败。
	INTERNALERROR_CALLTAGERROR = "InternalError.CallTagError"

	// Tvpc 接口调用失败。
	INTERNALERROR_CALLTVPCERROR = "InternalError.CallTvpcError"

	// VPC接口调用失败。
	INTERNALERROR_CALLVPCERROR = "InternalError.CallVpcError"

	// 调用其他服务异常。
	INTERNALERROR_CALLEEERROR = "InternalError.CalleeError"

	// 内部请求错误。
	INTERNALERROR_REQUESTERROR = "InternalError.RequestError"

	// 未找到该镜像。
	INVALIDIMAGEID_NOTFOUND = "InvalidImageId.NotFound"

	// 无效的启动配置。
	INVALIDLAUNCHCONFIGURATION = "InvalidLaunchConfiguration"

	// 启动配置ID无效。
	INVALIDLAUNCHCONFIGURATIONID = "InvalidLaunchConfigurationId"

	// 参数错误。
	INVALIDPARAMETER = "InvalidParameter"

	// 参数冲突，指定的多个参数冲突，不能同时存在。
	INVALIDPARAMETER_CONFLICT = "InvalidParameter.Conflict"

	// 主机名参数不适用于该镜像。
	INVALIDPARAMETER_HOSTNAMEUNAVAILABLE = "InvalidParameter.HostNameUnavailable"

	// 在特定场景下的不合法参数。
	INVALIDPARAMETER_INSCENARIO = "InvalidParameter.InScenario"

	// 无效的参数组合。
	INVALIDPARAMETER_INVALIDCOMBINATION = "InvalidParameter.InvalidCombination"

	// 指定的负载均衡器在当前伸缩组中没有找到。
	INVALIDPARAMETER_LOADBALANCERNOTINAUTOSCALINGGROUP = "InvalidParameter.LoadBalancerNotInAutoScalingGroup"

	// 参数缺失，两种参数之中必须指定其中一个。
	INVALIDPARAMETER_MUSTONEPARAMETER = "InvalidParameter.MustOneParameter"

	// 部分参数存在互斥应该删掉。
	INVALIDPARAMETER_PARAMETERMUSTBEDELETED = "InvalidParameter.ParameterMustBeDeleted"

	// 指定的两个参数冲突，不能同时存在。
	INVALIDPARAMETERCONFLICT = "InvalidParameterConflict"

	// 参数取值错误。
	INVALIDPARAMETERVALUE = "InvalidParameterValue"

	// 指定的基础容量过大，需小于等于最大实例数。
	INVALIDPARAMETERVALUE_BASECAPACITYTOOLARGE = "InvalidParameterValue.BaseCapacityTooLarge"

	// 在应当指定传统型负载均衡器的参数中，错误地指定了一个非传统型的负载均衡器。
	INVALIDPARAMETERVALUE_CLASSICLB = "InvalidParameterValue.ClassicLb"

	// 通知接收端类型冲突。
	INVALIDPARAMETERVALUE_CONFLICTNOTIFICATIONTARGET = "InvalidParameterValue.ConflictNotificationTarget"

	// 定时任务指定的Cron表达式无效。
	INVALIDPARAMETERVALUE_CRONEXPRESSIONILLEGAL = "InvalidParameterValue.CronExpressionIllegal"

	// CVM参数校验异常。
	INVALIDPARAMETERVALUE_CVMCONFIGURATIONERROR = "InvalidParameterValue.CvmConfigurationError"

	// CVM参数校验异常。
	INVALIDPARAMETERVALUE_CVMERROR = "InvalidParameterValue.CvmError"

	// 提供的应用型负载均衡器重复。
	INVALIDPARAMETERVALUE_DUPLICATEDFORWARDLB = "InvalidParameterValue.DuplicatedForwardLb"

	// 指定的子网重复。
	INVALIDPARAMETERVALUE_DUPLICATEDSUBNET = "InvalidParameterValue.DuplicatedSubnet"

	// 定时任务设置的结束时间在开始时间。
	INVALIDPARAMETERVALUE_ENDTIMEBEFORESTARTTIME = "InvalidParameterValue.EndTimeBeforeStartTime"

	// 无效的过滤器。
	INVALIDPARAMETERVALUE_FILTER = "InvalidParameterValue.Filter"

	// 在应当指定应用型负载均衡器的参数中，错误地指定了一个非应用型的负载均衡器。
	INVALIDPARAMETERVALUE_FORWARDLB = "InvalidParameterValue.ForwardLb"

	// 伸缩组名称重复。
	INVALIDPARAMETERVALUE_GROUPNAMEDUPLICATED = "InvalidParameterValue.GroupNameDuplicated"

	// 主机名不合法。
	INVALIDPARAMETERVALUE_HOSTNAMEILLEGAL = "InvalidParameterValue.HostNameIllegal"

	// 指定的镜像不存在。
	INVALIDPARAMETERVALUE_IMAGENOTFOUND = "InvalidParameterValue.ImageNotFound"

	// 设置的实例名称不合法。
	INVALIDPARAMETERVALUE_INSTANCENAMEILLEGAL = "InvalidParameterValue.InstanceNameIllegal"

	// 实例机型不支持。
	INVALIDPARAMETERVALUE_INSTANCETYPENOTSUPPORTED = "InvalidParameterValue.InstanceTypeNotSupported"

	// 伸缩活动ID无效。
	INVALIDPARAMETERVALUE_INVALIDACTIVITYID = "InvalidParameterValue.InvalidActivityId"

	// 伸缩组ID无效。
	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGGROUPID = "InvalidParameterValue.InvalidAutoScalingGroupId"

	// 通知ID无效。
	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGNOTIFICATIONID = "InvalidParameterValue.InvalidAutoScalingNotificationId"

	// 告警策略ID无效。
	INVALIDPARAMETERVALUE_INVALIDAUTOSCALINGPOLICYID = "InvalidParameterValue.InvalidAutoScalingPolicyId"

	// 为CLB指定的地域不合法。
	INVALIDPARAMETERVALUE_INVALIDCLBREGION = "InvalidParameterValue.InvalidClbRegion"

	// 过滤条件无效。
	INVALIDPARAMETERVALUE_INVALIDFILTER = "InvalidParameterValue.InvalidFilter"

	// 镜像ID无效。
	INVALIDPARAMETERVALUE_INVALIDIMAGEID = "InvalidParameterValue.InvalidImageId"

	// 实例ID无效。
	INVALIDPARAMETERVALUE_INVALIDINSTANCEID = "InvalidParameterValue.InvalidInstanceId"

	// 实例机型无效。
	INVALIDPARAMETERVALUE_INVALIDINSTANCETYPE = "InvalidParameterValue.InvalidInstanceType"

	// 输入的启动配置无效。
	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATION = "InvalidParameterValue.InvalidLaunchConfiguration"

	// 启动配置ID无效。
	INVALIDPARAMETERVALUE_INVALIDLAUNCHCONFIGURATIONID = "InvalidParameterValue.InvalidLaunchConfigurationId"

	// 生命周期挂钩ID无效。
	INVALIDPARAMETERVALUE_INVALIDLIFECYCLEHOOKID = "InvalidParameterValue.InvalidLifecycleHookId"

	// 指定的通知组 ID 不是数值字符串格式。
	INVALIDPARAMETERVALUE_INVALIDNOTIFICATIONUSERGROUPID = "InvalidParameterValue.InvalidNotificationUserGroupId"

	// 指定的PAI域名类型不支持。
	INVALIDPARAMETERVALUE_INVALIDPAIDOMAINNAMETYPE = "InvalidParameterValue.InvalidPaiDomainNameType"

	// 定时任务ID无效。
	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONID = "InvalidParameterValue.InvalidScheduledActionId"

	// 定时任务名称包含无效字符。
	INVALIDPARAMETERVALUE_INVALIDSCHEDULEDACTIONNAMEINCLUDEILLEGALCHAR = "InvalidParameterValue.InvalidScheduledActionNameIncludeIllegalChar"

	// 快照ID无效。
	INVALIDPARAMETERVALUE_INVALIDSNAPSHOTID = "InvalidParameterValue.InvalidSnapshotId"

	// 子网ID无效。
	INVALIDPARAMETERVALUE_INVALIDSUBNETID = "InvalidParameterValue.InvalidSubnetId"

	// 启动配置名称重复。
	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNAMEDUPLICATED = "InvalidParameterValue.LaunchConfigurationNameDuplicated"

	// 找不到指定启动配置。
	INVALIDPARAMETERVALUE_LAUNCHCONFIGURATIONNOTFOUND = "InvalidParameterValue.LaunchConfigurationNotFound"

	// 负载均衡器项目不一致。
	INVALIDPARAMETERVALUE_LBPROJECTINCONSISTENT = "InvalidParameterValue.LbProjectInconsistent"

	// 生命周期挂钩名称重复。
	INVALIDPARAMETERVALUE_LIFECYCLEHOOKNAMEDUPLICATED = "InvalidParameterValue.LifecycleHookNameDuplicated"

	// 取值超出限制。
	INVALIDPARAMETERVALUE_LIMITEXCEEDED = "InvalidParameterValue.LimitExceeded"

	// 无资源权限。
	INVALIDPARAMETERVALUE_NORESOURCEPERMISSION = "InvalidParameterValue.NoResourcePermission"

	// 提供的值不是浮点字符串格式。
	INVALIDPARAMETERVALUE_NOTSTRINGTYPEFLOAT = "InvalidParameterValue.NotStringTypeFloat"

	// 账号仅支持VPC网络。
	INVALIDPARAMETERVALUE_ONLYVPC = "InvalidParameterValue.OnlyVpc"

	// 项目ID不存在。
	INVALIDPARAMETERVALUE_PROJECTIDNOTFOUND = "InvalidParameterValue.ProjectIdNotFound"

	// 取值超出指定范围。
	INVALIDPARAMETERVALUE_RANGE = "InvalidParameterValue.Range"

	// 告警策略名称重复。
	INVALIDPARAMETERVALUE_SCALINGPOLICYNAMEDUPLICATE = "InvalidParameterValue.ScalingPolicyNameDuplicate"

	// 定时任务名称重复。
	INVALIDPARAMETERVALUE_SCHEDULEDACTIONNAMEDUPLICATE = "InvalidParameterValue.ScheduledActionNameDuplicate"

	// 伸缩组最大数量、最小数量、期望实例数取值不合法。
	INVALIDPARAMETERVALUE_SIZE = "InvalidParameterValue.Size"

	// 定时任务设置的开始时间在当前时间之前。
	INVALIDPARAMETERVALUE_STARTTIMEBEFORECURRENTTIME = "InvalidParameterValue.StartTimeBeforeCurrentTime"

	// 子网信息不合法。
	INVALIDPARAMETERVALUE_SUBNETIDS = "InvalidParameterValue.SubnetIds"

	// 负载均衡器四层监听器的后端端口重复。
	INVALIDPARAMETERVALUE_TARGETPORTDUPLICATED = "InvalidParameterValue.TargetPortDuplicated"

	// 指定的阈值不在有效范围。
	INVALIDPARAMETERVALUE_THRESHOLDOUTOFRANGE = "InvalidParameterValue.ThresholdOutOfRange"

	// 时间格式错误。
	INVALIDPARAMETERVALUE_TIMEFORMAT = "InvalidParameterValue.TimeFormat"

	// 取值过多。
	INVALIDPARAMETERVALUE_TOOLONG = "InvalidParameterValue.TooLong"

	// 输入参数值的长度小于最小值。
	INVALIDPARAMETERVALUE_TOOSHORT = "InvalidParameterValue.TooShort"

	// UserData格式错误。
	INVALIDPARAMETERVALUE_USERDATAFORMATERROR = "InvalidParameterValue.UserDataFormatError"

	// UserData长度过长。
	INVALIDPARAMETERVALUE_USERDATASIZEEXCEEDED = "InvalidParameterValue.UserDataSizeExceeded"

	// 用户组不存在。
	INVALIDPARAMETERVALUE_USERGROUPIDNOTFOUND = "InvalidParameterValue.UserGroupIdNotFound"

	// 指定的可用区与地域不匹配。
	INVALIDPARAMETERVALUE_ZONEMISMATCHREGION = "InvalidParameterValue.ZoneMismatchRegion"

	// 账户不支持该操作。
	INVALIDPERMISSION = "InvalidPermission"

	// 超过配额限制。
	LIMITEXCEEDED = "LimitExceeded"

	// 绑定指定的负载均衡器后，伸缩组绑定的负载均衡器总数超过了最大值。
	LIMITEXCEEDED_AFTERATTACHLBLIMITEXCEEDED = "LimitExceeded.AfterAttachLbLimitExceeded"

	// 伸缩组数量超过限制。
	LIMITEXCEEDED_AUTOSCALINGGROUPLIMITEXCEEDED = "LimitExceeded.AutoScalingGroupLimitExceeded"

	// 期望实例数超出限制。
	LIMITEXCEEDED_DESIREDCAPACITYLIMITEXCEEDED = "LimitExceeded.DesiredCapacityLimitExceeded"

	// 特定过滤器的值过多。
	LIMITEXCEEDED_FILTERVALUESTOOLONG = "LimitExceeded.FilterValuesTooLong"

	// 启动配置配额不足。
	LIMITEXCEEDED_LAUNCHCONFIGURATIONQUOTANOTENOUGH = "LimitExceeded.LaunchConfigurationQuotaNotEnough"

	// 最大实例数大于限制。
	LIMITEXCEEDED_MAXSIZELIMITEXCEEDED = "LimitExceeded.MaxSizeLimitExceeded"

	// 最小实例数低于限制。
	LIMITEXCEEDED_MINSIZELIMITEXCEEDED = "LimitExceeded.MinSizeLimitExceeded"

	// 当前剩余配额不足。
	LIMITEXCEEDED_QUOTANOTENOUGH = "LimitExceeded.QuotaNotEnough"

	// 定时任务数量超过限制。
	LIMITEXCEEDED_SCHEDULEDACTIONLIMITEXCEEDED = "LimitExceeded.ScheduledActionLimitExceeded"

	// 缺少参数错误。
	MISSINGPARAMETER = "MissingParameter"

	// 在特定场景下缺少参数。
	MISSINGPARAMETER_INSCENARIO = "MissingParameter.InScenario"

	// 竞价计费类型缺少对应的 InstanceMarketOptions 参数。
	MISSINGPARAMETER_INSTANCEMARKETOPTIONS = "MissingParameter.InstanceMarketOptions"

	// 伸缩组正在执行伸缩活动。
	RESOURCEINUSE_ACTIVITYINPROGRESS = "ResourceInUse.ActivityInProgress"

	// 伸缩组处于禁用状态。
	RESOURCEINUSE_AUTOSCALINGGROUPNOTACTIVE = "ResourceInUse.AutoScalingGroupNotActive"

	// 伸缩组内尚有正常实例。
	RESOURCEINUSE_INSTANCEINGROUP = "ResourceInUse.InstanceInGroup"

	// 指定的启动配置仍在伸缩组中使用。
	RESOURCEINUSE_LAUNCHCONFIGURATIONIDINUSE = "ResourceInUse.LaunchConfigurationIdInUse"

	// 超过伸缩组最大实例数。
	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPABOVEMAXSIZE = "ResourceInsufficient.AutoScalingGroupAboveMaxSize"

	// 少于伸缩组最小实例数。
	RESOURCEINSUFFICIENT_AUTOSCALINGGROUPBELOWMINSIZE = "ResourceInsufficient.AutoScalingGroupBelowMinSize"

	// 伸缩组内实例数超过最大实例数。
	RESOURCEINSUFFICIENT_INSERVICEINSTANCEABOVEMAXSIZE = "ResourceInsufficient.InServiceInstanceAboveMaxSize"

	// 伸缩组内实例数低于最小实例数。
	RESOURCEINSUFFICIENT_INSERVICEINSTANCEBELOWMINSIZE = "ResourceInsufficient.InServiceInstanceBelowMinSize"

	// 伸缩组不存在。
	RESOURCENOTFOUND_AUTOSCALINGGROUPNOTFOUND = "ResourceNotFound.AutoScalingGroupNotFound"

	// 通知不存在。
	RESOURCENOTFOUND_AUTOSCALINGNOTIFICATIONNOTFOUND = "ResourceNotFound.AutoScalingNotificationNotFound"

	// 指定的 CMQ queue 不存在。
	RESOURCENOTFOUND_CMQQUEUENOTFOUND = "ResourceNotFound.CmqQueueNotFound"

	// 指定的实例不存在。
	RESOURCENOTFOUND_INSTANCESNOTFOUND = "ResourceNotFound.InstancesNotFound"

	// 目标实例不在伸缩组内。
	RESOURCENOTFOUND_INSTANCESNOTINAUTOSCALINGGROUP = "ResourceNotFound.InstancesNotInAutoScalingGroup"

	// 指定的启动配置不存在。
	RESOURCENOTFOUND_LAUNCHCONFIGURATIONIDNOTFOUND = "ResourceNotFound.LaunchConfigurationIdNotFound"

	// 生命周期挂钩对应实例不存在。
	RESOURCENOTFOUND_LIFECYCLEHOOKINSTANCENOTFOUND = "ResourceNotFound.LifecycleHookInstanceNotFound"

	// 无法找到指定生命周期挂钩。
	RESOURCENOTFOUND_LIFECYCLEHOOKNOTFOUND = "ResourceNotFound.LifecycleHookNotFound"

	// 指定的Listener不存在。
	RESOURCENOTFOUND_LISTENERNOTFOUND = "ResourceNotFound.ListenerNotFound"

	// 找不到指定负载均衡器。
	RESOURCENOTFOUND_LOADBALANCERNOTFOUND = "ResourceNotFound.LoadBalancerNotFound"

	// 指定的Location不存在。
	RESOURCENOTFOUND_LOCATIONNOTFOUND = "ResourceNotFound.LocationNotFound"

	// 告警策略不存在。
	RESOURCENOTFOUND_SCALINGPOLICYNOTFOUND = "ResourceNotFound.ScalingPolicyNotFound"

	// 指定的定时任务不存在。
	RESOURCENOTFOUND_SCHEDULEDACTIONNOTFOUND = "ResourceNotFound.ScheduledActionNotFound"

	// TDMQ-CMQ 队列不存在。
	RESOURCENOTFOUND_TDMQCMQQUEUENOTFOUND = "ResourceNotFound.TDMQCMQQueueNotFound"

	// TDMQ-CMQ 主题不存在。
	RESOURCENOTFOUND_TDMQCMQTOPICNOTFOUND = "ResourceNotFound.TDMQCMQTopicNotFound"

	// 伸缩组状态异常。
	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPABNORMALSTATUS = "ResourceUnavailable.AutoScalingGroupAbnormalStatus"

	// 伸缩组被停用。
	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPDISABLED = "ResourceUnavailable.AutoScalingGroupDisabled"

	// 伸缩组正在活动中。
	RESOURCEUNAVAILABLE_AUTOSCALINGGROUPINACTIVITY = "ResourceUnavailable.AutoScalingGroupInActivity"

	// 指定的 CMQ Topic 无订阅者。
	RESOURCEUNAVAILABLE_CMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.CmqTopicHasNoSubscriber"

	// 实例和伸缩组Vpc不一致。
	RESOURCEUNAVAILABLE_CVMVPCINCONSISTENT = "ResourceUnavailable.CvmVpcInconsistent"

	// 指定的实例正在活动中。
	RESOURCEUNAVAILABLE_INSTANCEINOPERATION = "ResourceUnavailable.InstanceInOperation"

	// 实例不支持关机不收费。
	RESOURCEUNAVAILABLE_INSTANCENOTSUPPORTSTOPCHARGING = "ResourceUnavailable.InstanceNotSupportStopCharging"

	// 实例已存在于伸缩组中。
	RESOURCEUNAVAILABLE_INSTANCESALREADYINAUTOSCALINGGROUP = "ResourceUnavailable.InstancesAlreadyInAutoScalingGroup"

	// 启动配置状态异常。
	RESOURCEUNAVAILABLE_LAUNCHCONFIGURATIONSTATUSABNORMAL = "ResourceUnavailable.LaunchConfigurationStatusAbnormal"

	// CLB实例的后端地域与AS服务所在地域不一致。
	RESOURCEUNAVAILABLE_LBBACKENDREGIONINCONSISTENT = "ResourceUnavailable.LbBackendRegionInconsistent"

	// 负载均衡器项目不一致。
	RESOURCEUNAVAILABLE_LBPROJECTINCONSISTENT = "ResourceUnavailable.LbProjectInconsistent"

	// 负载均衡器VPC与伸缩组不一致。
	RESOURCEUNAVAILABLE_LBVPCINCONSISTENT = "ResourceUnavailable.LbVpcInconsistent"

	// 生命周期动作已经被设置。
	RESOURCEUNAVAILABLE_LIFECYCLEACTIONRESULTHASSET = "ResourceUnavailable.LifecycleActionResultHasSet"

	// LB 在指定的伸缩组内处于活动中。
	RESOURCEUNAVAILABLE_LOADBALANCERINOPERATION = "ResourceUnavailable.LoadBalancerInOperation"

	// 项目不一致。
	RESOURCEUNAVAILABLE_PROJECTINCONSISTENT = "ResourceUnavailable.ProjectInconsistent"

	// 关机实例不允许添加到伸缩组。
	RESOURCEUNAVAILABLE_STOPPEDINSTANCENOTALLOWATTACH = "ResourceUnavailable.StoppedInstanceNotAllowAttach"

	// TDMQ-CMQ 主题无订阅者。
	RESOURCEUNAVAILABLE_TDMQCMQTOPICHASNOSUBSCRIBER = "ResourceUnavailable.TDMQCMQTopicHasNoSubscriber"

	// 指定的可用区不可用。
	RESOURCEUNAVAILABLE_ZONEUNAVAILABLE = "ResourceUnavailable.ZoneUnavailable"
)
