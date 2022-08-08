package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ListScalingPolicyExecuteLogsRequest struct {
	// 伸缩策略ID。

	ScalingPolicyId string `json:"scaling_policy_id"`
	// 日志ID。

	LogId *string `json:"log_id,omitempty"`
	// 伸缩资源类型：伸缩组：SCALING_GROUP。带宽：BANDWIDTH

	ScalingResourceType *ListScalingPolicyExecuteLogsRequestScalingResourceType `json:"scaling_resource_type,omitempty"`
	// 伸缩资源ID。

	ScalingResourceId *string `json:"scaling_resource_id,omitempty"`
	// 策略执行类型：SCHEDULED：自动触发（定时）。RECURRENCE：自动触发（周期）。ALARM：自动触发（告警）。MANUAL：手动触发。

	ExecuteType *ListScalingPolicyExecuteLogsRequestExecuteType `json:"execute_type,omitempty"`
	// 查询的起始时间，格式是“yyyy-MM-ddThh:mm:ssZ”。

	StartTime *string `json:"start_time,omitempty"`
	// 查询的截止时间，格式是“yyyy-MM-ddThh:mm:ssZ”。

	EndTime *string `json:"end_time,omitempty"`
	// 查询的起始行号，默认为0。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数，默认20，最大100。

	Limit *int32 `json:"limit,omitempty"`
}

func (o ListScalingPolicyExecuteLogsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingPolicyExecuteLogsRequest struct{}"
	}

	return strings.Join([]string{"ListScalingPolicyExecuteLogsRequest", string(data)}, " ")
}

type ListScalingPolicyExecuteLogsRequestScalingResourceType struct {
	value string
}

type ListScalingPolicyExecuteLogsRequestScalingResourceTypeEnum struct {
	SCALING_GROUP ListScalingPolicyExecuteLogsRequestScalingResourceType
	BANDWIDTH     ListScalingPolicyExecuteLogsRequestScalingResourceType
}

func GetListScalingPolicyExecuteLogsRequestScalingResourceTypeEnum() ListScalingPolicyExecuteLogsRequestScalingResourceTypeEnum {
	return ListScalingPolicyExecuteLogsRequestScalingResourceTypeEnum{
		SCALING_GROUP: ListScalingPolicyExecuteLogsRequestScalingResourceType{
			value: "SCALING_GROUP",
		},
		BANDWIDTH: ListScalingPolicyExecuteLogsRequestScalingResourceType{
			value: "BANDWIDTH",
		},
	}
}

func (c ListScalingPolicyExecuteLogsRequestScalingResourceType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingPolicyExecuteLogsRequestScalingResourceType) UnmarshalJSON(b []byte) error {
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

type ListScalingPolicyExecuteLogsRequestExecuteType struct {
	value string
}

type ListScalingPolicyExecuteLogsRequestExecuteTypeEnum struct {
	SCHEDULED  ListScalingPolicyExecuteLogsRequestExecuteType
	RECURRENCE ListScalingPolicyExecuteLogsRequestExecuteType
	ALARM      ListScalingPolicyExecuteLogsRequestExecuteType
	MANUAL     ListScalingPolicyExecuteLogsRequestExecuteType
}

func GetListScalingPolicyExecuteLogsRequestExecuteTypeEnum() ListScalingPolicyExecuteLogsRequestExecuteTypeEnum {
	return ListScalingPolicyExecuteLogsRequestExecuteTypeEnum{
		SCHEDULED: ListScalingPolicyExecuteLogsRequestExecuteType{
			value: "SCHEDULED",
		},
		RECURRENCE: ListScalingPolicyExecuteLogsRequestExecuteType{
			value: "RECURRENCE",
		},
		ALARM: ListScalingPolicyExecuteLogsRequestExecuteType{
			value: "ALARM",
		},
		MANUAL: ListScalingPolicyExecuteLogsRequestExecuteType{
			value: "MANUAL",
		},
	}
}

func (c ListScalingPolicyExecuteLogsRequestExecuteType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingPolicyExecuteLogsRequestExecuteType) UnmarshalJSON(b []byte) error {
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
