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

// 启停伸缩组请求
type EnableOrDisableScalingGroupRequestBody struct {
	// 启用或停止伸缩组操作的标识。启用：resume 停止：pause
	Action *EnableOrDisableScalingGroupRequestBodyAction `json:"action,omitempty"`
}

func (o EnableOrDisableScalingGroupRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"EnableOrDisableScalingGroupRequestBody", string(data)}, " ")
}

type EnableOrDisableScalingGroupRequestBodyAction struct {
	value string
}

type EnableOrDisableScalingGroupRequestBodyActionEnum struct {
	RESUME EnableOrDisableScalingGroupRequestBodyAction
	PAUSE  EnableOrDisableScalingGroupRequestBodyAction
}

func GetEnableOrDisableScalingGroupRequestBodyActionEnum() EnableOrDisableScalingGroupRequestBodyActionEnum {
	return EnableOrDisableScalingGroupRequestBodyActionEnum{
		RESUME: EnableOrDisableScalingGroupRequestBodyAction{
			value: "resume",
		},
		PAUSE: EnableOrDisableScalingGroupRequestBodyAction{
			value: "pause",
		},
	}
}

func (c EnableOrDisableScalingGroupRequestBodyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *EnableOrDisableScalingGroupRequestBodyAction) UnmarshalJSON(b []byte) error {
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
