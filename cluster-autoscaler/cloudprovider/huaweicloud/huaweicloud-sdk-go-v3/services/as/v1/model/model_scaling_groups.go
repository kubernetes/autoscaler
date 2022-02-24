package model

import (
	"errors"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
)

// 伸缩组详情
type ScalingGroups struct {
	// 伸缩组名称。

	ScalingGroupName *string `json:"scaling_group_name,omitempty"`
	// 伸缩组ID。

	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
	// 伸缩组状态。

	ScalingGroupStatus *ScalingGroupsScalingGroupStatus `json:"scaling_group_status,omitempty"`
	// 伸缩配置ID。

	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	// 伸缩配置名称。

	ScalingConfigurationName *string `json:"scaling_configuration_name,omitempty"`
	// 伸缩组中当前实例数。

	CurrentInstanceNumber *int32 `json:"current_instance_number,omitempty"`
	// 伸缩组期望实例数。

	DesireInstanceNumber *int32 `json:"desire_instance_number,omitempty"`
	// 伸缩组最小实例数。

	MinInstanceNumber *int32 `json:"min_instance_number,omitempty"`
	// 伸缩组最大实例数

	MaxInstanceNumber *int32 `json:"max_instance_number,omitempty"`
	// 冷却时间，单位是秒。

	CoolDownTime *int32 `json:"cool_down_time,omitempty"`
	// 经典型负载均衡监听器ID，多个负载均衡监听器ID以逗号分隔。

	LbListenerId *string `json:"lb_listener_id,omitempty"`
	// 增强型负载均衡器信息，该参数为预留字段。

	LbaasListeners *[]LbaasListenersResult `json:"lbaas_listeners,omitempty"`
	// 可用分区信息

	AvailableZones *[]string `json:"available_zones,omitempty"`
	// 网络信息

	Networks *[]NetworksResult `json:"networks,omitempty"`
	// 安全组信息

	SecurityGroups *[]SecurityGroupsResult `json:"security_groups,omitempty"`
	// 创建伸缩组时间，遵循UTC时间。

	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 伸缩组所在的VPC ID。

	VpcId *string `json:"vpc_id,omitempty"`
	// 伸缩组详情。

	Detail *string `json:"detail,omitempty"`
	// 伸缩组伸缩标志。

	IsScaling *bool `json:"is_scaling,omitempty"`
	// 健康检查方式。

	HealthPeriodicAuditMethod *ScalingGroupsHealthPeriodicAuditMethod `json:"health_periodic_audit_method,omitempty"`
	// 健康检查的间隔时间。

	HealthPeriodicAuditTime *int32 `json:"health_periodic_audit_time,omitempty"`
	// 健康状况检查宽限期。

	HealthPeriodicAuditGracePeriod *int32 `json:"health_periodic_audit_grace_period,omitempty"`
	// 移除策略。

	InstanceTerminatePolicy *ScalingGroupsInstanceTerminatePolicy `json:"instance_terminate_policy,omitempty"`
	// 通知方式：EMAIL为发送邮件通知。

	Notifications *[]string `json:"notifications,omitempty"`
	// 删除云服务器是否删除云服务器绑定的弹性IP。

	DeletePublicip *bool `json:"delete_publicip,omitempty"`
	// 删除云服务器是否删除云服务器绑定的数据盘

	DeleteVolume *bool `json:"delete_volume,omitempty"`
	// 该参数为预留字段

	CloudLocationId *string `json:"cloud_location_id,omitempty"`
	// 企业项目ID

	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 伸缩组活动类型

	ActivityType *string `json:"activity_type,omitempty"`
	// 伸缩组扩缩容时目标AZ选择的优先级策略

	MultiAzPriorityPolicy *string `json:"multi_az_priority_policy,omitempty"`
	// 伸缩组描述信息

	Description *string `json:"description,omitempty"`
}

func (o ScalingGroups) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingGroups struct{}"
	}

	return strings.Join([]string{"ScalingGroups", string(data)}, " ")
}

type ScalingGroupsScalingGroupStatus struct {
	value string
}

type ScalingGroupsScalingGroupStatusEnum struct {
	INSERVICE ScalingGroupsScalingGroupStatus
	PAUSED    ScalingGroupsScalingGroupStatus
	ERROR     ScalingGroupsScalingGroupStatus
	DELETING  ScalingGroupsScalingGroupStatus
	FREEZED   ScalingGroupsScalingGroupStatus
}

func GetScalingGroupsScalingGroupStatusEnum() ScalingGroupsScalingGroupStatusEnum {
	return ScalingGroupsScalingGroupStatusEnum{
		INSERVICE: ScalingGroupsScalingGroupStatus{
			value: "INSERVICE",
		},
		PAUSED: ScalingGroupsScalingGroupStatus{
			value: "PAUSED",
		},
		ERROR: ScalingGroupsScalingGroupStatus{
			value: "ERROR",
		},
		DELETING: ScalingGroupsScalingGroupStatus{
			value: "DELETING",
		},
		FREEZED: ScalingGroupsScalingGroupStatus{
			value: "FREEZED",
		},
	}
}

func (c ScalingGroupsScalingGroupStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingGroupsScalingGroupStatus) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}

type ScalingGroupsHealthPeriodicAuditMethod struct {
	value string
}

type ScalingGroupsHealthPeriodicAuditMethodEnum struct {
	ELB_AUDIT  ScalingGroupsHealthPeriodicAuditMethod
	NOVA_AUDIT ScalingGroupsHealthPeriodicAuditMethod
}

func GetScalingGroupsHealthPeriodicAuditMethodEnum() ScalingGroupsHealthPeriodicAuditMethodEnum {
	return ScalingGroupsHealthPeriodicAuditMethodEnum{
		ELB_AUDIT: ScalingGroupsHealthPeriodicAuditMethod{
			value: "ELB_AUDIT",
		},
		NOVA_AUDIT: ScalingGroupsHealthPeriodicAuditMethod{
			value: "NOVA_AUDIT",
		},
	}
}

func (c ScalingGroupsHealthPeriodicAuditMethod) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingGroupsHealthPeriodicAuditMethod) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}

type ScalingGroupsInstanceTerminatePolicy struct {
	value string
}

type ScalingGroupsInstanceTerminatePolicyEnum struct {
	OLD_CONFIG_OLD_INSTANCE ScalingGroupsInstanceTerminatePolicy
	OLD_CONFIG_NEW_INSTANCE ScalingGroupsInstanceTerminatePolicy
	OLD_INSTANCE            ScalingGroupsInstanceTerminatePolicy
	NEW_INSTANCE            ScalingGroupsInstanceTerminatePolicy
}

func GetScalingGroupsInstanceTerminatePolicyEnum() ScalingGroupsInstanceTerminatePolicyEnum {
	return ScalingGroupsInstanceTerminatePolicyEnum{
		OLD_CONFIG_OLD_INSTANCE: ScalingGroupsInstanceTerminatePolicy{
			value: "OLD_CONFIG_OLD_INSTANCE",
		},
		OLD_CONFIG_NEW_INSTANCE: ScalingGroupsInstanceTerminatePolicy{
			value: "OLD_CONFIG_NEW_INSTANCE",
		},
		OLD_INSTANCE: ScalingGroupsInstanceTerminatePolicy{
			value: "OLD_INSTANCE",
		},
		NEW_INSTANCE: ScalingGroupsInstanceTerminatePolicy{
			value: "NEW_INSTANCE",
		},
	}
}

func (c ScalingGroupsInstanceTerminatePolicy) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingGroupsInstanceTerminatePolicy) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
