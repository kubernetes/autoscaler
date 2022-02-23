package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 伸缩策略执行日志列表
type ScalingPolicyExecuteLogList struct {
	// 策略执行状态：SUCCESS：成功。FAIL：失败。EXECUTING：执行中

	Status *ScalingPolicyExecuteLogListStatus `json:"status,omitempty"`
	// 策略执行失败原因。

	FailedReason *string `json:"failed_reason,omitempty"`
	// 策略执行类型：SCHEDULE：自动触发（定时）。RECURRENCE：自动触发（周期）。ALARM：自动警告（告警）。MANUAL：手动触发

	ExecuteType *ScalingPolicyExecuteLogListExecuteType `json:"execute_type,omitempty"`
	// 策略执行时间，遵循UTC时间。

	ExecuteTime *string `json:"execute_time,omitempty"`
	// 策略执行日志ID。

	Id *string `json:"id,omitempty"`
	// 租户id。

	TenantId *string `json:"tenant_id,omitempty"`
	// 伸缩策略ID。

	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
	// 伸缩资源类型：伸缩组：SCALING_GROUP 带宽：BANDWIDTH

	ScalingResourceType *ScalingPolicyExecuteLogListScalingResourceType `json:"scaling_resource_type,omitempty"`
	// 伸缩资源ID。

	ScalingResourceId *string `json:"scaling_resource_id,omitempty"`
	// 伸缩原始值。

	OldValue *string `json:"old_value,omitempty"`
	// 伸缩目标值。

	DesireValue *string `json:"desire_value,omitempty"`
	// 操作限制。当scaling_resource_type为BANDWIDTH时，且operation不为SET时，limit_value生效，单位为Mbit/s。此时，当operation为ADD时，limit_value表示最高带宽限制；当operation为REDUCE时，limit_value表示最低带宽限制。

	LimitValue *string `json:"limit_value,omitempty"`
	// 策略执行任务类型。ADD：添加。REMOVE：减少。SET：设置为

	Type *ScalingPolicyExecuteLogListType `json:"type,omitempty"`
	// 策略执行动作包含的具体任务

	JobRecords *[]JobRecords `json:"job_records,omitempty"`

	MetaData *EipMetaData `json:"meta_data,omitempty"`
}

func (o ScalingPolicyExecuteLogList) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingPolicyExecuteLogList struct{}"
	}

	return strings.Join([]string{"ScalingPolicyExecuteLogList", string(data)}, " ")
}

type ScalingPolicyExecuteLogListStatus struct {
	value string
}

type ScalingPolicyExecuteLogListStatusEnum struct {
	SUCCESS   ScalingPolicyExecuteLogListStatus
	FAIL      ScalingPolicyExecuteLogListStatus
	EXECUTING ScalingPolicyExecuteLogListStatus
}

func GetScalingPolicyExecuteLogListStatusEnum() ScalingPolicyExecuteLogListStatusEnum {
	return ScalingPolicyExecuteLogListStatusEnum{
		SUCCESS: ScalingPolicyExecuteLogListStatus{
			value: "SUCCESS",
		},
		FAIL: ScalingPolicyExecuteLogListStatus{
			value: "FAIL",
		},
		EXECUTING: ScalingPolicyExecuteLogListStatus{
			value: "EXECUTING",
		},
	}
}

func (c ScalingPolicyExecuteLogListStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingPolicyExecuteLogListStatus) UnmarshalJSON(b []byte) error {
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

type ScalingPolicyExecuteLogListExecuteType struct {
	value string
}

type ScalingPolicyExecuteLogListExecuteTypeEnum struct {
	SCHEDULE   ScalingPolicyExecuteLogListExecuteType
	RECURRENCE ScalingPolicyExecuteLogListExecuteType
	ALARM      ScalingPolicyExecuteLogListExecuteType
	MANUAL     ScalingPolicyExecuteLogListExecuteType
}

func GetScalingPolicyExecuteLogListExecuteTypeEnum() ScalingPolicyExecuteLogListExecuteTypeEnum {
	return ScalingPolicyExecuteLogListExecuteTypeEnum{
		SCHEDULE: ScalingPolicyExecuteLogListExecuteType{
			value: "SCHEDULE",
		},
		RECURRENCE: ScalingPolicyExecuteLogListExecuteType{
			value: "RECURRENCE",
		},
		ALARM: ScalingPolicyExecuteLogListExecuteType{
			value: "ALARM",
		},
		MANUAL: ScalingPolicyExecuteLogListExecuteType{
			value: "MANUAL",
		},
	}
}

func (c ScalingPolicyExecuteLogListExecuteType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingPolicyExecuteLogListExecuteType) UnmarshalJSON(b []byte) error {
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

type ScalingPolicyExecuteLogListScalingResourceType struct {
	value string
}

type ScalingPolicyExecuteLogListScalingResourceTypeEnum struct {
	SCALING_GROUP ScalingPolicyExecuteLogListScalingResourceType
	BANDWIDTH     ScalingPolicyExecuteLogListScalingResourceType
}

func GetScalingPolicyExecuteLogListScalingResourceTypeEnum() ScalingPolicyExecuteLogListScalingResourceTypeEnum {
	return ScalingPolicyExecuteLogListScalingResourceTypeEnum{
		SCALING_GROUP: ScalingPolicyExecuteLogListScalingResourceType{
			value: "SCALING_GROUP",
		},
		BANDWIDTH: ScalingPolicyExecuteLogListScalingResourceType{
			value: "BANDWIDTH",
		},
	}
}

func (c ScalingPolicyExecuteLogListScalingResourceType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingPolicyExecuteLogListScalingResourceType) UnmarshalJSON(b []byte) error {
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

type ScalingPolicyExecuteLogListType struct {
	value string
}

type ScalingPolicyExecuteLogListTypeEnum struct {
	ADD    ScalingPolicyExecuteLogListType
	REMOVE ScalingPolicyExecuteLogListType
	SET    ScalingPolicyExecuteLogListType
}

func GetScalingPolicyExecuteLogListTypeEnum() ScalingPolicyExecuteLogListTypeEnum {
	return ScalingPolicyExecuteLogListTypeEnum{
		ADD: ScalingPolicyExecuteLogListType{
			value: "ADD",
		},
		REMOVE: ScalingPolicyExecuteLogListType{
			value: "REMOVE",
		},
		SET: ScalingPolicyExecuteLogListType{
			value: "SET",
		},
	}
}

func (c ScalingPolicyExecuteLogListType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingPolicyExecuteLogListType) UnmarshalJSON(b []byte) error {
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
