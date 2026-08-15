package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 启停伸缩组请求
type ResumeScalingGroupOption struct {
	// 启用或停止伸缩组操作的标识。启用：resume 停止：pause

	Action ResumeScalingGroupOptionAction `json:"action"`
}

func (o ResumeScalingGroupOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResumeScalingGroupOption struct{}"
	}

	return strings.Join([]string{"ResumeScalingGroupOption", string(data)}, " ")
}

type ResumeScalingGroupOptionAction struct {
	value string
}

type ResumeScalingGroupOptionActionEnum struct {
	RESUME ResumeScalingGroupOptionAction
}

func GetResumeScalingGroupOptionActionEnum() ResumeScalingGroupOptionActionEnum {
	return ResumeScalingGroupOptionActionEnum{
		RESUME: ResumeScalingGroupOptionAction{
			value: "resume",
		},
	}
}

func (c ResumeScalingGroupOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ResumeScalingGroupOptionAction) UnmarshalJSON(b []byte) error {
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
