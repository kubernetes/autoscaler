package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 执行或启用或停止伸缩策略
type ExecuteScalingPolicyOption struct {
	// 执行或启用或停止伸缩策略操作的标识。执行：execute。启用：resume。停止：pause。

	Action ExecuteScalingPolicyOptionAction `json:"action"`
}

func (o ExecuteScalingPolicyOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ExecuteScalingPolicyOption struct{}"
	}

	return strings.Join([]string{"ExecuteScalingPolicyOption", string(data)}, " ")
}

type ExecuteScalingPolicyOptionAction struct {
	value string
}

type ExecuteScalingPolicyOptionActionEnum struct {
	EXECUTE ExecuteScalingPolicyOptionAction
}

func GetExecuteScalingPolicyOptionActionEnum() ExecuteScalingPolicyOptionActionEnum {
	return ExecuteScalingPolicyOptionActionEnum{
		EXECUTE: ExecuteScalingPolicyOptionAction{
			value: "execute",
		},
	}
}

func (c ExecuteScalingPolicyOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ExecuteScalingPolicyOptionAction) UnmarshalJSON(b []byte) error {
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
