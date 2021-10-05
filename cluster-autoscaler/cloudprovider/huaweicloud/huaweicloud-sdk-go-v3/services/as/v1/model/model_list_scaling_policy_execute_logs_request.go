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

// Request Object
type ListScalingPolicyExecuteLogsRequest struct {
	ScalingPolicyId     string                                                  `json:"scaling_policy_id"`
	LogId               *string                                                 `json:"log_id,omitempty"`
	ScalingResourceType *ListScalingPolicyExecuteLogsRequestScalingResourceType `json:"scaling_resource_type,omitempty"`
	ScalingResourceId   *string                                                 `json:"scaling_resource_id,omitempty"`
	ExecuteType         *ListScalingPolicyExecuteLogsRequestExecuteType         `json:"execute_type,omitempty"`
	StartTime           *sdktime.SdkTime                                        `json:"start_time,omitempty"`
	EndTime             *sdktime.SdkTime                                        `json:"end_time,omitempty"`
	StartNumber         *int32                                                  `json:"start_number,omitempty"`
	Limit               *int32                                                  `json:"limit,omitempty"`
}

func (o ListScalingPolicyExecuteLogsRequest) String() string {
	data, _ := json.Marshal(o)
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
	return json.Marshal(c.value)
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
	return json.Marshal(c.value)
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
