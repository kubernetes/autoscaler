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
	"strings"
)

// 执行或启用或停止伸缩策略
type ExecuteScalingPolicyRequestBody struct {
	// 执行或启用或停止伸缩策略操作的标识。执行：execute。启用：resume。停止：pause。
	Action *ExecuteScalingPolicyRequestBodyAction `json:"action,omitempty"`
}

func (o ExecuteScalingPolicyRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ExecuteScalingPolicyRequestBody", string(data)}, " ")
}

type ExecuteScalingPolicyRequestBodyAction struct {
	value string
}

type ExecuteScalingPolicyRequestBodyActionEnum struct {
	EXECUTE ExecuteScalingPolicyRequestBodyAction
	RESUME  ExecuteScalingPolicyRequestBodyAction
	PAUSE   ExecuteScalingPolicyRequestBodyAction
}

func GetExecuteScalingPolicyRequestBodyActionEnum() ExecuteScalingPolicyRequestBodyActionEnum {
	return ExecuteScalingPolicyRequestBodyActionEnum{
		EXECUTE: ExecuteScalingPolicyRequestBodyAction{
			value: "execute",
		},
		RESUME: ExecuteScalingPolicyRequestBodyAction{
			value: "resume",
		},
		PAUSE: ExecuteScalingPolicyRequestBodyAction{
			value: "pause",
		},
	}
}

func (c ExecuteScalingPolicyRequestBodyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ExecuteScalingPolicyRequestBodyAction) UnmarshalJSON(b []byte) error {
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
