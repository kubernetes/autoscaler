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

// 定时、周期任务策略
type ScheduledPolicy struct {
	// 触发时间，遵循UTC时间。如果scaling_policy_type为SCHEDULED，则格式为：YYYY-MM-DDThh:mmZ。如果scaling_policy_type为RECURRENCE，则格式为：hh:mm。
	LaunchTime *sdktime.SdkTime `json:"launch_time,omitempty"`
	// 周期触发类型，scaling_policy_type为RECURRENCE时该项必选。Daily：每天执行一次。Weekly：每周指定天执行一次。Monthly：每月指定天执行一次。
	RecurrenceType *ScheduledPolicyRecurrenceType `json:"recurrence_type,omitempty"`
	// 周期触发任务数值，scaling_policy_type为RECURRENCE时该项必选。类型为Daily时，该字段为null，表示每天执行类型为Weekly时，该字段取值范围为1-7，1表示星期日，以此类推，以”,”分割，例如：1,3,5。类型为Monthly时，该字段取值范围为1-31，分别表示每月的日期，以“,”分割，例如：1,10,13,28。
	RecurrenceValue *string `json:"recurrence_value,omitempty"`
	// 周期策略重复执行开始时间，遵循UTC时间。默认为当前时间，格式为：YYYY-MM-DDThh：mZ
	StartTime *sdktime.SdkTime `json:"start_time,omitempty"`
	// 周期策略重复执行结束时间，遵循UTC时间，scaling_policy_type为RECURRENCE时该项必选。当为周期类型策略时，不得早于当前时间和开始时间。格式为：YYYY-MM-DDThh：mmZ
	EndTime *sdktime.SdkTime `json:"end_time,omitempty"`
}

func (o ScheduledPolicy) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ScheduledPolicy", string(data)}, " ")
}

type ScheduledPolicyRecurrenceType struct {
	value string
}

type ScheduledPolicyRecurrenceTypeEnum struct {
	DAILY   ScheduledPolicyRecurrenceType
	WEEKLY  ScheduledPolicyRecurrenceType
	MONTHLY ScheduledPolicyRecurrenceType
}

func GetScheduledPolicyRecurrenceTypeEnum() ScheduledPolicyRecurrenceTypeEnum {
	return ScheduledPolicyRecurrenceTypeEnum{
		DAILY: ScheduledPolicyRecurrenceType{
			value: "Daily",
		},
		WEEKLY: ScheduledPolicyRecurrenceType{
			value: "Weekly",
		},
		MONTHLY: ScheduledPolicyRecurrenceType{
			value: "Monthly",
		},
	}
}

func (c ScheduledPolicyRecurrenceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ScheduledPolicyRecurrenceType) UnmarshalJSON(b []byte) error {
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
