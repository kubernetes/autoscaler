package model

import (
	"errors"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
)

// 伸缩策略
type ScalingPoliciesV2 struct {
	// 伸缩策略名称。

	ScalingPolicyName *string `json:"scaling_policy_name,omitempty"`
	// 伸缩策略ID。

	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	// 伸缩资源ID。

	ScalingResourceId *string `json:"scaling_resource_id,omitempty"`
	// 伸缩资源类型。伸缩组：SCALING_GROUP。带宽：BANDWIDTH。

	ScalingResourceType *ScalingPoliciesV2ScalingResourceType `json:"scaling_resource_type,omitempty"`
	// 伸缩策略状态。INSERVICE：使用中。PAUSED：停止。EXECUTING：执行中。

	PolicyStatus *ScalingPoliciesV2PolicyStatus `json:"policy_status,omitempty"`
	// 伸缩策略类型：ALARM：告警策略，此时alarm_id有返回，scheduled_policy不会返回。SCHEDULED：定时策略，此时alarm_id不会返回，scheduled_policy有返回，并且recurrence_type、recurrence_value、start_time和end_time不会返回。RECURRENCE：周期策略，此时alarm_id不会返回，scheduled_policy有返回，并且recurrence_type、recurrence_value、start_time和end_time有返回。

	ScalingPolicyType *ScalingPoliciesV2ScalingPolicyType `json:"scaling_policy_type,omitempty"`
	// 告警ID。

	AlarmId *string `json:"alarm_id,omitempty"`

	ScheduledPolicy *ScheduledPolicy `json:"scheduled_policy,omitempty"`

	ScalingPolicyAction *ScalingPolicyActionV2 `json:"scaling_policy_action,omitempty"`
	// 冷却时间，取值范围0-86400，默认为300，单位是秒。

	CoolDownTime *int32 `json:"cool_down_time,omitempty"`
	// 创建伸缩策略时间，遵循UTC时间

	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`

	MetaData *ScalingPolicyV2MetaData `json:"meta_data,omitempty"`
	// 伸缩策略描述（1-256个字符）

	Description *string `json:"description,omitempty"`
}

func (o ScalingPoliciesV2) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingPoliciesV2 struct{}"
	}

	return strings.Join([]string{"ScalingPoliciesV2", string(data)}, " ")
}

type ScalingPoliciesV2ScalingResourceType struct {
	value string
}

type ScalingPoliciesV2ScalingResourceTypeEnum struct {
	SCALING_GROUP ScalingPoliciesV2ScalingResourceType
	BANDWIDTH     ScalingPoliciesV2ScalingResourceType
}

func GetScalingPoliciesV2ScalingResourceTypeEnum() ScalingPoliciesV2ScalingResourceTypeEnum {
	return ScalingPoliciesV2ScalingResourceTypeEnum{
		SCALING_GROUP: ScalingPoliciesV2ScalingResourceType{
			value: "SCALING_GROUP",
		},
		BANDWIDTH: ScalingPoliciesV2ScalingResourceType{
			value: "BANDWIDTH",
		},
	}
}

func (c ScalingPoliciesV2ScalingResourceType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingPoliciesV2ScalingResourceType) UnmarshalJSON(b []byte) error {
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

type ScalingPoliciesV2PolicyStatus struct {
	value string
}

type ScalingPoliciesV2PolicyStatusEnum struct {
	INSERVICE ScalingPoliciesV2PolicyStatus
	PAUSED    ScalingPoliciesV2PolicyStatus
	EXECUTING ScalingPoliciesV2PolicyStatus
}

func GetScalingPoliciesV2PolicyStatusEnum() ScalingPoliciesV2PolicyStatusEnum {
	return ScalingPoliciesV2PolicyStatusEnum{
		INSERVICE: ScalingPoliciesV2PolicyStatus{
			value: "INSERVICE",
		},
		PAUSED: ScalingPoliciesV2PolicyStatus{
			value: "PAUSED",
		},
		EXECUTING: ScalingPoliciesV2PolicyStatus{
			value: "EXECUTING",
		},
	}
}

func (c ScalingPoliciesV2PolicyStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingPoliciesV2PolicyStatus) UnmarshalJSON(b []byte) error {
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

type ScalingPoliciesV2ScalingPolicyType struct {
	value string
}

type ScalingPoliciesV2ScalingPolicyTypeEnum struct {
	ALARM      ScalingPoliciesV2ScalingPolicyType
	SCHEDULED  ScalingPoliciesV2ScalingPolicyType
	RECURRENCE ScalingPoliciesV2ScalingPolicyType
}

func GetScalingPoliciesV2ScalingPolicyTypeEnum() ScalingPoliciesV2ScalingPolicyTypeEnum {
	return ScalingPoliciesV2ScalingPolicyTypeEnum{
		ALARM: ScalingPoliciesV2ScalingPolicyType{
			value: "ALARM",
		},
		SCHEDULED: ScalingPoliciesV2ScalingPolicyType{
			value: "SCHEDULED",
		},
		RECURRENCE: ScalingPoliciesV2ScalingPolicyType{
			value: "RECURRENCE",
		},
	}
}

func (c ScalingPoliciesV2ScalingPolicyType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingPoliciesV2ScalingPolicyType) UnmarshalJSON(b []byte) error {
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
