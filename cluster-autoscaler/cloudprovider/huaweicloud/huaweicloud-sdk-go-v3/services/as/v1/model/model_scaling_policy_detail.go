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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// 伸缩策略
type ScalingPolicyDetail struct {
	// 伸缩组ID。
	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
	// 伸缩策略名称。
	ScalingPolicyName *string `json:"scaling_policy_name,omitempty"`
	// 伸缩策略ID。
	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	// 伸缩策略类型：ALARM：告警策略，此时alarm_id有返回，scheduled_policy不会返回。SCHEDULED：定时策略，此时alarm_id不会返回，scheduled_policy有返回，并且recurrence_type、recurrence_value、start_time和end_time不会返回。RECURRENCE：周期策略，此时alarm_id不会返回，scheduled_policy有返回，并且recurrence_type、recurrence_value、start_time和end_time有返回。
	ScalingPolicyType *ScalingPolicyDetailScalingPolicyType `json:"scaling_policy_type,omitempty"`
	// 告警ID，即告警规则的ID，当scaling_policy_type为ALARM时该项必选，此时scheduled_policy不生效。创建告警策略成功后，会自动为该告警ID对应的告警规则的alarm_actions字段增加类型为autoscaling的告警触发动作。告警ID通过查询云监控告警规则列表获取，详见《云监控API参考》的“查询告警规则列表”。
	AlarmId             *string              `json:"alarm_id,omitempty"`
	ScheduledPolicy     *ScheduledPolicy     `json:"scheduled_policy,omitempty"`
	ScalingPolicyAction *ScalingPolicyAction `json:"scaling_policy_action,omitempty"`
	// 冷却时间，取值范围0-86400，默认为300，单位是秒。
	CoolDownTime *int32 `json:"cool_down_time,omitempty"`
	// 创建伸缩策略时间，遵循UTC时间。
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
}

func (o ScalingPolicyDetail) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ScalingPolicyDetail", string(data)}, " ")
}

type ScalingPolicyDetailScalingPolicyType struct {
	value string
}

type ScalingPolicyDetailScalingPolicyTypeEnum struct {
	ALARM      ScalingPolicyDetailScalingPolicyType
	SCHEDULED  ScalingPolicyDetailScalingPolicyType
	RECURRENCE ScalingPolicyDetailScalingPolicyType
}

func GetScalingPolicyDetailScalingPolicyTypeEnum() ScalingPolicyDetailScalingPolicyTypeEnum {
	return ScalingPolicyDetailScalingPolicyTypeEnum{
		ALARM: ScalingPolicyDetailScalingPolicyType{
			value: "ALARM",
		},
		SCHEDULED: ScalingPolicyDetailScalingPolicyType{
			value: "SCHEDULED",
		},
		RECURRENCE: ScalingPolicyDetailScalingPolicyType{
			value: "RECURRENCE",
		},
	}
}

func (c ScalingPolicyDetailScalingPolicyType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ScalingPolicyDetailScalingPolicyType) UnmarshalJSON(b []byte) error {
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
