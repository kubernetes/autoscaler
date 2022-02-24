package model

import (
	"errors"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
)

// 伸缩策略详情
type ScalingV2PolicyDetail struct {
	// 伸缩策略名称。

	ScalingPolicyName *string `json:"scaling_policy_name,omitempty"`
	// 伸缩策略ID。

	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	// 伸缩资源ID。

	ScalingResourceId *string `json:"scaling_resource_id,omitempty"`
	// 伸缩资源类型。伸缩组：SCALING_GROUP。带宽：BANDWIDTH。

	ScalingResourceType *ScalingV2PolicyDetailScalingResourceType `json:"scaling_resource_type,omitempty"`
	// 伸缩策略状态。INSERVICE：使用中。PAUSED：停止。EXECUTING：执行中。

	PolicyStatus *ScalingV2PolicyDetailPolicyStatus `json:"policy_status,omitempty"`
	// 伸缩策略类型：ALARM：告警策略，此时alarm_id有返回，scheduled_policy不会返回。SCHEDULED：定时策略，此时alarm_id不会返回，scheduled_policy有返回，并且recurrence_type、recurrence_value、start_time和end_time不会返回。RECURRENCE：周期策略，此时alarm_id不会返回，scheduled_policy有返回，并且recurrence_type、recurrence_value、start_time和end_time有返回。

	ScalingPolicyType *ScalingV2PolicyDetailScalingPolicyType `json:"scaling_policy_type,omitempty"`
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

func (o ScalingV2PolicyDetail) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingV2PolicyDetail struct{}"
	}

	return strings.Join([]string{"ScalingV2PolicyDetail", string(data)}, " ")
}

type ScalingV2PolicyDetailScalingResourceType struct {
	value string
}

type ScalingV2PolicyDetailScalingResourceTypeEnum struct {
	SCALING_GROUP ScalingV2PolicyDetailScalingResourceType
	BANDWIDTH     ScalingV2PolicyDetailScalingResourceType
}

func GetScalingV2PolicyDetailScalingResourceTypeEnum() ScalingV2PolicyDetailScalingResourceTypeEnum {
	return ScalingV2PolicyDetailScalingResourceTypeEnum{
		SCALING_GROUP: ScalingV2PolicyDetailScalingResourceType{
			value: "SCALING_GROUP",
		},
		BANDWIDTH: ScalingV2PolicyDetailScalingResourceType{
			value: "BANDWIDTH",
		},
	}
}

func (c ScalingV2PolicyDetailScalingResourceType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingV2PolicyDetailScalingResourceType) UnmarshalJSON(b []byte) error {
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

type ScalingV2PolicyDetailPolicyStatus struct {
	value string
}

type ScalingV2PolicyDetailPolicyStatusEnum struct {
	INSERVICE ScalingV2PolicyDetailPolicyStatus
	PAUSED    ScalingV2PolicyDetailPolicyStatus
	EXECUTING ScalingV2PolicyDetailPolicyStatus
}

func GetScalingV2PolicyDetailPolicyStatusEnum() ScalingV2PolicyDetailPolicyStatusEnum {
	return ScalingV2PolicyDetailPolicyStatusEnum{
		INSERVICE: ScalingV2PolicyDetailPolicyStatus{
			value: "INSERVICE",
		},
		PAUSED: ScalingV2PolicyDetailPolicyStatus{
			value: "PAUSED",
		},
		EXECUTING: ScalingV2PolicyDetailPolicyStatus{
			value: "EXECUTING",
		},
	}
}

func (c ScalingV2PolicyDetailPolicyStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingV2PolicyDetailPolicyStatus) UnmarshalJSON(b []byte) error {
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

type ScalingV2PolicyDetailScalingPolicyType struct {
	value string
}

type ScalingV2PolicyDetailScalingPolicyTypeEnum struct {
	ALARM      ScalingV2PolicyDetailScalingPolicyType
	SCHEDULED  ScalingV2PolicyDetailScalingPolicyType
	RECURRENCE ScalingV2PolicyDetailScalingPolicyType
}

func GetScalingV2PolicyDetailScalingPolicyTypeEnum() ScalingV2PolicyDetailScalingPolicyTypeEnum {
	return ScalingV2PolicyDetailScalingPolicyTypeEnum{
		ALARM: ScalingV2PolicyDetailScalingPolicyType{
			value: "ALARM",
		},
		SCHEDULED: ScalingV2PolicyDetailScalingPolicyType{
			value: "SCHEDULED",
		},
		RECURRENCE: ScalingV2PolicyDetailScalingPolicyType{
			value: "RECURRENCE",
		},
	}
}

func (c ScalingV2PolicyDetailScalingPolicyType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingV2PolicyDetailScalingPolicyType) UnmarshalJSON(b []byte) error {
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
