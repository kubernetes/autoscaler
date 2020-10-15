/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 修改伸缩组详情
type UpdateScalingGroupRequestBody struct {
	// 伸缩组名称(1-64个字符)，只能包含中文、字母、数字、下划线、中划线。
	ScalingGroupName *string `json:"scaling_group_name,omitempty"`
	// 期望实例数量，默认值为最小实例数。最小实例数<=期望实例数<=最大实例数
	DesireInstanceNumber *int32 `json:"desire_instance_number,omitempty"`
	// 最小实例数量。
	MinInstanceNumber *int32 `json:"min_instance_number,omitempty"`
	// 最大实例数量，大于等于最小实例数。
	MaxInstanceNumber *int32 `json:"max_instance_number,omitempty"`
	// 冷却时间，取值范围0-86400，单位是秒。
	CoolDownTime *int32 `json:"cool_down_time,omitempty"`
	// 可用分区信息。弹性伸缩活动中自动添加的云服务器会被创建在指定的可用区中。如果没有指定可用分区，会由系统自动指定可用分区。详情请参考地区和终端节点。仅当同时满足以下条件时才可以修改：伸缩组中无伸缩活动；实例数为0；伸缩组为非启用状态。
	AvailableZones *[]string `json:"available_zones,omitempty"`
	// 网络信息，最多支持选择5个子网，传入的第一个子网默认作为云服务器的主网卡。使用vpc_id通过查询VPC服务子网列表接口获取，仅当同时满足以下条件时，才可以修改：伸缩组中无伸缩活动；实例数为0；伸缩组为非启用状态。
	Networks *[]Networks `json:"networks,omitempty"`
	// 安全组信息，最多支持选择1个安全组。使用vpc_id通过查询VPC服务安全组列表接口获取，详见《虚拟私有云API参考》的“查询安全组列表”。当伸缩配置和伸缩组同时指定安全组时，将以伸缩配置中的安全组为准；当伸缩配置和伸缩组都没有指定安全组时，将使用默认安全组。为了使用灵活性更高，推荐在伸缩配置中指定安全组。仅当同时满足以下条件时，才可以修改：伸缩组无伸缩活动；实例数为0；伸缩组为非启用状态。
	SecurityGroups *[]SecurityGroups `json:"security_groups,omitempty"`
	// 弹性负载均衡（经典型）监听器ID，最多支持绑定3个负载均衡监听器，多个负载均衡监听器ID以逗号分隔。首先使用vpc_id通过查询ELB服务负载均衡器列表接口获取负载均衡器的ID，详见《弹性负载均衡API参考》的“查询负载均衡器列表”，再使用该ID查询监听器列表获取，详见《弹性负载均衡API参考》的“查询监听器列表”。仅当同时满足以下条件时，才可以修改：伸缩组无伸缩活动；实例数为0；伸缩组为非启用状态。
	LbListenerId *string `json:"lb_listener_id,omitempty"`
	// 弹性负载均衡器（增强型）信息，最多支持绑定3个负载均衡。该字段与lb_listener_id互斥。
	LbaasListeners *[]LbaasListeners `json:"lbaas_listeners,omitempty"`
	// 伸缩组实例健康检查方式：ELB_AUDIT和NOVA_AUDIT。当伸缩组参数中设置负载均衡时，默认为ELB_AUDIT；否则默认为NOVA_AUDIT。ELB_AUDIT表示负载均衡健康检查方式，在有监听器的伸缩组中有效。NOVA_AUDIT表示弹性伸缩自带的健康检查方式。
	HealthPeriodicAuditMethod *UpdateScalingGroupRequestBodyHealthPeriodicAuditMethod `json:"health_periodic_audit_method,omitempty"`
	// 伸缩组实例健康检查周期（分钟）：1、5、15、60、180。若设置为0，可以实现10秒级健康检查。
	HealthPeriodicAuditTime *UpdateScalingGroupRequestBodyHealthPeriodicAuditTime `json:"health_periodic_audit_time,omitempty"`
	// 伸缩组实例健康状况检查宽限期，取值范围0-86400，单位是秒。当实例加入伸缩组并且进入已启用状态后，健康状况检查宽限期才会启动，伸缩组会等健康状况检查宽限期结束后才检查实例的运行状况。当伸缩组实例健康检查方式为ELB_AUDIT时，该参数生效，若不设置该参数，默认为10分钟。
	HealthPeriodicAuditGracePeriod *int32 `json:"health_periodic_audit_grace_period,omitempty"`
	// 伸缩组实例移除策略：OLD_CONFIG_OLD_INSTANCE（默认）：从根据“较早创建的配置”创建的实例中筛选出较早创建的实例被优先移除。OLD_CONFIG_NEW_INSTANCE：从根据“较早创建的配置”创建的实例中筛选出较新创建的实例被优先移除。OLD_INSTANCE：较早创建的实例被优先移除。NEW_INSTANCE：较新创建的实例将被优先移除。
	InstanceTerminatePolicy *UpdateScalingGroupRequestBodyInstanceTerminatePolicy `json:"instance_terminate_policy,omitempty"`
	// 伸缩配置ID，通过查询弹性伸缩配置列表接口获取，详见查询弹性伸缩配置列表
	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	// 通知方式：EMAIL为发送邮件通知。该通知方式即将被废除，建议给弹性伸缩组配置通知功能。详见通知。
	Notifications *[]string `json:"notifications,omitempty"`
	// 配置删除云服务器时是否删除云服务器绑定的弹性IP。取值为true或false，默认为false。true：删除云服务器时，会同时删除绑定在云服务器上的弹性IP。当弹性IP的计费方式为包年包月时，不会被删除。false：删除云服务器时，仅解绑定在云服务器上的弹性IP，不删除弹性IP。
	DeletePublicip *bool `json:"delete_publicip,omitempty"`
	// 企业项目ID，用于指定伸缩组归属的企业项目。当伸缩组配置企业项目时，由该伸缩组创建的弹性云服务器将归属于该企业项目。当没有指定企业项目时，将使用企业项目ID为0的默认项目。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o UpdateScalingGroupRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateScalingGroupRequestBody", string(data)}, " ")
}

type UpdateScalingGroupRequestBodyHealthPeriodicAuditMethod struct {
	value string
}

type UpdateScalingGroupRequestBodyHealthPeriodicAuditMethodEnum struct {
	ELB_AUDIT  UpdateScalingGroupRequestBodyHealthPeriodicAuditMethod
	NOVA_AUDIT UpdateScalingGroupRequestBodyHealthPeriodicAuditMethod
}

func GetUpdateScalingGroupRequestBodyHealthPeriodicAuditMethodEnum() UpdateScalingGroupRequestBodyHealthPeriodicAuditMethodEnum {
	return UpdateScalingGroupRequestBodyHealthPeriodicAuditMethodEnum{
		ELB_AUDIT: UpdateScalingGroupRequestBodyHealthPeriodicAuditMethod{
			value: "ELB_AUDIT",
		},
		NOVA_AUDIT: UpdateScalingGroupRequestBodyHealthPeriodicAuditMethod{
			value: "NOVA_AUDIT",
		},
	}
}

func (c UpdateScalingGroupRequestBodyHealthPeriodicAuditMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateScalingGroupRequestBodyHealthPeriodicAuditMethod) UnmarshalJSON(b []byte) error {
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

type UpdateScalingGroupRequestBodyHealthPeriodicAuditTime struct {
	value int32
}

type UpdateScalingGroupRequestBodyHealthPeriodicAuditTimeEnum struct {
	E_0   UpdateScalingGroupRequestBodyHealthPeriodicAuditTime
	E_1   UpdateScalingGroupRequestBodyHealthPeriodicAuditTime
	E_5   UpdateScalingGroupRequestBodyHealthPeriodicAuditTime
	E_15  UpdateScalingGroupRequestBodyHealthPeriodicAuditTime
	E_60  UpdateScalingGroupRequestBodyHealthPeriodicAuditTime
	E_180 UpdateScalingGroupRequestBodyHealthPeriodicAuditTime
}

func GetUpdateScalingGroupRequestBodyHealthPeriodicAuditTimeEnum() UpdateScalingGroupRequestBodyHealthPeriodicAuditTimeEnum {
	return UpdateScalingGroupRequestBodyHealthPeriodicAuditTimeEnum{
		E_0: UpdateScalingGroupRequestBodyHealthPeriodicAuditTime{
			value: 0,
		}, E_1: UpdateScalingGroupRequestBodyHealthPeriodicAuditTime{
			value: 1,
		}, E_5: UpdateScalingGroupRequestBodyHealthPeriodicAuditTime{
			value: 5,
		}, E_15: UpdateScalingGroupRequestBodyHealthPeriodicAuditTime{
			value: 15,
		}, E_60: UpdateScalingGroupRequestBodyHealthPeriodicAuditTime{
			value: 60,
		}, E_180: UpdateScalingGroupRequestBodyHealthPeriodicAuditTime{
			value: 180,
		},
	}
}

func (c UpdateScalingGroupRequestBodyHealthPeriodicAuditTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateScalingGroupRequestBodyHealthPeriodicAuditTime) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("int32")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(int32)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to int32 error")
	}
}

type UpdateScalingGroupRequestBodyInstanceTerminatePolicy struct {
	value string
}

type UpdateScalingGroupRequestBodyInstanceTerminatePolicyEnum struct {
	OLD_CONFIG_OLD_INSTANCE UpdateScalingGroupRequestBodyInstanceTerminatePolicy
	OLD_CONFIG_NEW_INSTANCE UpdateScalingGroupRequestBodyInstanceTerminatePolicy
	OLD_INSTANCE            UpdateScalingGroupRequestBodyInstanceTerminatePolicy
	NEW_INSTANCE            UpdateScalingGroupRequestBodyInstanceTerminatePolicy
}

func GetUpdateScalingGroupRequestBodyInstanceTerminatePolicyEnum() UpdateScalingGroupRequestBodyInstanceTerminatePolicyEnum {
	return UpdateScalingGroupRequestBodyInstanceTerminatePolicyEnum{
		OLD_CONFIG_OLD_INSTANCE: UpdateScalingGroupRequestBodyInstanceTerminatePolicy{
			value: "OLD_CONFIG_OLD_INSTANCE",
		},
		OLD_CONFIG_NEW_INSTANCE: UpdateScalingGroupRequestBodyInstanceTerminatePolicy{
			value: "OLD_CONFIG_NEW_INSTANCE",
		},
		OLD_INSTANCE: UpdateScalingGroupRequestBodyInstanceTerminatePolicy{
			value: "OLD_INSTANCE",
		},
		NEW_INSTANCE: UpdateScalingGroupRequestBodyInstanceTerminatePolicy{
			value: "NEW_INSTANCE",
		},
	}
}

func (c UpdateScalingGroupRequestBodyInstanceTerminatePolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateScalingGroupRequestBodyInstanceTerminatePolicy) UnmarshalJSON(b []byte) error {
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
