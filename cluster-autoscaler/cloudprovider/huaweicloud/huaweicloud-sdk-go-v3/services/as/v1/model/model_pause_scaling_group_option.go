package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 启停伸缩组请求
type PauseScalingGroupOption struct {
	// 启用或停止伸缩组操作的标识。启用：resume 停止：pause

	Action PauseScalingGroupOptionAction `json:"action"`
}

func (o PauseScalingGroupOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "PauseScalingGroupOption struct{}"
	}

	return strings.Join([]string{"PauseScalingGroupOption", string(data)}, " ")
}

type PauseScalingGroupOptionAction struct {
	value string
}

type PauseScalingGroupOptionActionEnum struct {
	PAUSE PauseScalingGroupOptionAction
}

func GetPauseScalingGroupOptionActionEnum() PauseScalingGroupOptionActionEnum {
	return PauseScalingGroupOptionActionEnum{
		PAUSE: PauseScalingGroupOptionAction{
			value: "pause",
		},
	}
}

func (c PauseScalingGroupOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *PauseScalingGroupOptionAction) UnmarshalJSON(b []byte) error {
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
