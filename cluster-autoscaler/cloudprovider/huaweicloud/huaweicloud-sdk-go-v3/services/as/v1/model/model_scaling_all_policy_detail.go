package model

import (
	"errors"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
)

// 伸缩策略
type ScalingAllPolicyDetail struct {
	// 伸缩策略名称。

	ScalingPolicyName *string `json:"scaling_policy_name,omitempty"`
	// 伸缩策略ID。

	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	// 伸缩资源ID。

	ScalingResourceId *string `json:"scaling_resource_id,omitempty"`
	// 伸缩资源类型。伸缩组：SCALING_GROUP。带宽：BANDWIDTH。

	ScalingResourceType *ScalingAllPolicyDetailScalingResourceType `json:"scaling_resource_type,omitempty"`
	// 伸缩策略状态。INSERVICE：使用中。PAUSED：停止。EXECUTING：执行中。

	PolicyStatus *ScalingAllPolicyDetailPolicyStatus `json:"policy_status,omitempty"`
	// 伸缩策略类型：ALARM：告警策略，此时alarm_id有返回，scheduled_policy不会返回。SCHEDULED：定时策略，此时alarm_id不会返回，scheduled_policy有返回，并且recurrence_type、recurrence_value、start_time和end_time不会返回。RECURRENCE：周期策略，此时alarm_id不会返回，scheduled_policy有返回，并且recurrence_type、recurrence_value、start_time和end_time有返回。

	ScalingPolicyType *ScalingAllPolicyDetailScalingPolicyType `json:"scaling_policy_type,omitempty"`
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

func (o ScalingAllPolicyDetail) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingAllPolicyDetail struct{}"
	}

	return strings.Join([]string{"ScalingAllPolicyDetail", string(data)}, " ")
}

type ScalingAllPolicyDetailScalingResourceType struct {
	value string
}

type ScalingAllPolicyDetailScalingResourceTypeEnum struct {
	SCALING_GROUP ScalingAllPolicyDetailScalingResourceType
	BANDWIDTH     ScalingAllPolicyDetailScalingResourceType
}

func GetScalingAllPolicyDetailScalingResourceTypeEnum() ScalingAllPolicyDetailScalingResourceTypeEnum {
	return ScalingAllPolicyDetailScalingResourceTypeEnum{
		SCALING_GROUP: ScalingAllPolicyDetailScalingResourceType{
			value: "SCALING_GROUP",
		},
		BANDWIDTH: ScalingAllPolicyDetailScalingResourceType{
			value: "BANDWIDTH",
		},
	}
}

func (c ScalingAllPolicyDetailScalingResourceType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingAllPolicyDetailScalingResourceType) UnmarshalJSON(b []byte) error {
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

type ScalingAllPolicyDetailPolicyStatus struct {
	value string
}

type ScalingAllPolicyDetailPolicyStatusEnum struct {
	INSERVICE ScalingAllPolicyDetailPolicyStatus
	PAUSED    ScalingAllPolicyDetailPolicyStatus
	EXECUTING ScalingAllPolicyDetailPolicyStatus
}

func GetScalingAllPolicyDetailPolicyStatusEnum() ScalingAllPolicyDetailPolicyStatusEnum {
	return ScalingAllPolicyDetailPolicyStatusEnum{
		INSERVICE: ScalingAllPolicyDetailPolicyStatus{
			value: "INSERVICE",
		},
		PAUSED: ScalingAllPolicyDetailPolicyStatus{
			value: "PAUSED",
		},
		EXECUTING: ScalingAllPolicyDetailPolicyStatus{
			value: "EXECUTING",
		},
	}
}

func (c ScalingAllPolicyDetailPolicyStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingAllPolicyDetailPolicyStatus) UnmarshalJSON(b []byte) error {
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

type ScalingAllPolicyDetailScalingPolicyType struct {
	value string
}

type ScalingAllPolicyDetailScalingPolicyTypeEnum struct {
	ALARM      ScalingAllPolicyDetailScalingPolicyType
	SCHEDULED  ScalingAllPolicyDetailScalingPolicyType
	RECURRENCE ScalingAllPolicyDetailScalingPolicyType
}

func GetScalingAllPolicyDetailScalingPolicyTypeEnum() ScalingAllPolicyDetailScalingPolicyTypeEnum {
	return ScalingAllPolicyDetailScalingPolicyTypeEnum{
		ALARM: ScalingAllPolicyDetailScalingPolicyType{
			value: "ALARM",
		},
		SCHEDULED: ScalingAllPolicyDetailScalingPolicyType{
			value: "SCHEDULED",
		},
		RECURRENCE: ScalingAllPolicyDetailScalingPolicyType{
			value: "RECURRENCE",
		},
	}
}

func (c ScalingAllPolicyDetailScalingPolicyType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingAllPolicyDetailScalingPolicyType) UnmarshalJSON(b []byte) error {
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
