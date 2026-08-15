package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 修改伸缩策略
type UpdateScalingV2PolicyOption struct {
	// 策略名称（1-64）字符，可以用中文、字母、数字、下划线、中划线的组合。

	ScalingPolicyName *string `json:"scaling_policy_name,omitempty"`
	// 伸缩资源ID，伸缩组唯一标识或带宽唯一标识。如果scaling_resource_type为SCALING_GROUP，对应伸缩组唯一标识。如果scaling_resource_type为BANDWIDTH，对应带宽唯一标识。

	ScalingResourceId *string `json:"scaling_resource_id,omitempty"`
	// 伸缩资源类型。伸缩组：SCALING_GROUP。带宽：BANDWIDTH。

	ScalingResourceType *UpdateScalingV2PolicyOptionScalingResourceType `json:"scaling_resource_type,omitempty"`
	// 策略类型。告警策略：ALARM（与alarm_id对应）；定时策略：SCHEDULED（与scheduled_policy对应）；周期策略：RECURRENCE（与scheduled_policy对应）

	ScalingPolicyType *UpdateScalingV2PolicyOptionScalingPolicyType `json:"scaling_policy_type,omitempty"`
	// 告警ID，即告警规则的ID，当scaling_policy_type为ALARM时该项必选，此时scheduled_policy不生效。创建告警策略成功后，会自动为该告警ID对应的告警规则的alarm_actions字段增加类型为autoscaling的告警触发动作。告警ID通过查询云监控告警规则列表获取，详见《云监控API参考》的“查询告警规则列表”。

	AlarmId *string `json:"alarm_id,omitempty"`

	ScheduledPolicy *ScheduledPolicy `json:"scheduled_policy,omitempty"`

	ScalingPolicyAction *ScalingPolicyActionV2 `json:"scaling_policy_action,omitempty"`
	// 冷却时间，取值范围0-86400，默认为300，单位是秒。

	CoolDownTime *int32 `json:"cool_down_time,omitempty"`
	// 伸缩策略描述（1-256个字符）

	Description *string `json:"description,omitempty"`
}

func (o UpdateScalingV2PolicyOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UpdateScalingV2PolicyOption struct{}"
	}

	return strings.Join([]string{"UpdateScalingV2PolicyOption", string(data)}, " ")
}

type UpdateScalingV2PolicyOptionScalingResourceType struct {
	value string
}

type UpdateScalingV2PolicyOptionScalingResourceTypeEnum struct {
	SCALING_GROUP UpdateScalingV2PolicyOptionScalingResourceType
	BANDWIDTH     UpdateScalingV2PolicyOptionScalingResourceType
}

func GetUpdateScalingV2PolicyOptionScalingResourceTypeEnum() UpdateScalingV2PolicyOptionScalingResourceTypeEnum {
	return UpdateScalingV2PolicyOptionScalingResourceTypeEnum{
		SCALING_GROUP: UpdateScalingV2PolicyOptionScalingResourceType{
			value: "SCALING_GROUP",
		},
		BANDWIDTH: UpdateScalingV2PolicyOptionScalingResourceType{
			value: "BANDWIDTH",
		},
	}
}

func (c UpdateScalingV2PolicyOptionScalingResourceType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *UpdateScalingV2PolicyOptionScalingResourceType) UnmarshalJSON(b []byte) error {
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

type UpdateScalingV2PolicyOptionScalingPolicyType struct {
	value string
}

type UpdateScalingV2PolicyOptionScalingPolicyTypeEnum struct {
	ALARM      UpdateScalingV2PolicyOptionScalingPolicyType
	SCHEDULED  UpdateScalingV2PolicyOptionScalingPolicyType
	RECURRENCE UpdateScalingV2PolicyOptionScalingPolicyType
}

func GetUpdateScalingV2PolicyOptionScalingPolicyTypeEnum() UpdateScalingV2PolicyOptionScalingPolicyTypeEnum {
	return UpdateScalingV2PolicyOptionScalingPolicyTypeEnum{
		ALARM: UpdateScalingV2PolicyOptionScalingPolicyType{
			value: "ALARM",
		},
		SCHEDULED: UpdateScalingV2PolicyOptionScalingPolicyType{
			value: "SCHEDULED",
		},
		RECURRENCE: UpdateScalingV2PolicyOptionScalingPolicyType{
			value: "RECURRENCE",
		},
	}
}

func (c UpdateScalingV2PolicyOptionScalingPolicyType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *UpdateScalingV2PolicyOptionScalingPolicyType) UnmarshalJSON(b []byte) error {
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
