package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 创建伸缩策略（V2）
type CreateScalingPolicyV2Option struct {
	// 策略名称（1-64）字符，可以用中文、字母、数字、下划线、中划线的组合。

	ScalingPolicyName string `json:"scaling_policy_name"`
	// 伸缩资源ID，伸缩组唯一标识或带宽唯一标识。如果scaling_resource_type为SCALING_GROUP，对应伸缩组唯一标识。如果scaling_resource_type为BANDWIDTH，对应带宽唯一标识。

	ScalingResourceId string `json:"scaling_resource_id"`
	// 伸缩资源类型。伸缩组：SCALING_GROUP。带宽：BANDWIDTH。

	ScalingResourceType CreateScalingPolicyV2OptionScalingResourceType `json:"scaling_resource_type"`
	// 策略类型。告警策略：ALARM（与alarm_id对应）；定时策略：SCHEDULED（与scheduled_policy对应）；周期策略：RECURRENCE（与scheduled_policy对应）

	ScalingPolicyType CreateScalingPolicyV2OptionScalingPolicyType `json:"scaling_policy_type"`
	// 告警ID，即告警规则的ID，当scaling_policy_type为ALARM时该项必选，此时scheduled_policy不生效。创建告警策略成功后，会自动为该告警ID对应的告警规则的alarm_actions字段增加类型为autoscaling的告警触发动作。告警ID通过查询云监控告警规则列表获取，详见《云监控API参考》的“查询告警规则列表”。

	AlarmId *string `json:"alarm_id,omitempty"`

	ScheduledPolicy *ScheduledPolicy `json:"scheduled_policy,omitempty"`

	ScalingPolicyAction *ScalingPolicyActionV2 `json:"scaling_policy_action,omitempty"`
	// 冷却时间，取值范围0-86400，默认为300，单位是秒。

	CoolDownTime *int32 `json:"cool_down_time,omitempty"`
	// 伸缩策略描述（1-256个字符）

	Description *string `json:"description,omitempty"`
}

func (o CreateScalingPolicyV2Option) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingPolicyV2Option struct{}"
	}

	return strings.Join([]string{"CreateScalingPolicyV2Option", string(data)}, " ")
}

type CreateScalingPolicyV2OptionScalingResourceType struct {
	value string
}

type CreateScalingPolicyV2OptionScalingResourceTypeEnum struct {
	SCALING_GROUP CreateScalingPolicyV2OptionScalingResourceType
	BANDWIDTH     CreateScalingPolicyV2OptionScalingResourceType
}

func GetCreateScalingPolicyV2OptionScalingResourceTypeEnum() CreateScalingPolicyV2OptionScalingResourceTypeEnum {
	return CreateScalingPolicyV2OptionScalingResourceTypeEnum{
		SCALING_GROUP: CreateScalingPolicyV2OptionScalingResourceType{
			value: "SCALING_GROUP",
		},
		BANDWIDTH: CreateScalingPolicyV2OptionScalingResourceType{
			value: "BANDWIDTH",
		},
	}
}

func (c CreateScalingPolicyV2OptionScalingResourceType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *CreateScalingPolicyV2OptionScalingResourceType) UnmarshalJSON(b []byte) error {
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

type CreateScalingPolicyV2OptionScalingPolicyType struct {
	value string
}

type CreateScalingPolicyV2OptionScalingPolicyTypeEnum struct {
	ALARM      CreateScalingPolicyV2OptionScalingPolicyType
	SCHEDULED  CreateScalingPolicyV2OptionScalingPolicyType
	RECURRENCE CreateScalingPolicyV2OptionScalingPolicyType
}

func GetCreateScalingPolicyV2OptionScalingPolicyTypeEnum() CreateScalingPolicyV2OptionScalingPolicyTypeEnum {
	return CreateScalingPolicyV2OptionScalingPolicyTypeEnum{
		ALARM: CreateScalingPolicyV2OptionScalingPolicyType{
			value: "ALARM",
		},
		SCHEDULED: CreateScalingPolicyV2OptionScalingPolicyType{
			value: "SCHEDULED",
		},
		RECURRENCE: CreateScalingPolicyV2OptionScalingPolicyType{
			value: "RECURRENCE",
		},
	}
}

func (c CreateScalingPolicyV2OptionScalingPolicyType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *CreateScalingPolicyV2OptionScalingPolicyType) UnmarshalJSON(b []byte) error {
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
