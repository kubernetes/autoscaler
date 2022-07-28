package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 执行或启用或停止伸缩策略
type ResumeScalingPolicyOption struct {
	// 执行或启用或停止伸缩策略操作的标识。执行：execute。启用：resume。停止：pause。

	Action ResumeScalingPolicyOptionAction `json:"action"`
}

func (o ResumeScalingPolicyOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResumeScalingPolicyOption struct{}"
	}

	return strings.Join([]string{"ResumeScalingPolicyOption", string(data)}, " ")
}

type ResumeScalingPolicyOptionAction struct {
	value string
}

type ResumeScalingPolicyOptionActionEnum struct {
	RESUME ResumeScalingPolicyOptionAction
}

func GetResumeScalingPolicyOptionActionEnum() ResumeScalingPolicyOptionActionEnum {
	return ResumeScalingPolicyOptionActionEnum{
		RESUME: ResumeScalingPolicyOptionAction{
			value: "resume",
		},
	}
}

func (c ResumeScalingPolicyOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ResumeScalingPolicyOptionAction) UnmarshalJSON(b []byte) error {
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
