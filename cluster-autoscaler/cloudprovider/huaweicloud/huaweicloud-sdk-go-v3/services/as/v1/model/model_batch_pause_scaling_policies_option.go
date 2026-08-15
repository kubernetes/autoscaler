package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 批量操作弹性伸缩策略
type BatchPauseScalingPoliciesOption struct {
	// 伸缩策略ID。

	ScalingPolicyId []string `json:"scaling_policy_id"`
	// 是否强制删除伸缩策略。默认为no，可选值为yes或no。只有action为delete时，该字段才生效。

	ForceDelete *BatchPauseScalingPoliciesOptionForceDelete `json:"force_delete,omitempty"`
	// 批量操作伸缩策略action标识：删除：delete。启用：resume。停止：pause。

	Action BatchPauseScalingPoliciesOptionAction `json:"action"`
	// 是否删除告警策略使用的告警规则。可选值为yes或no，默认为no。  只有action为delete时，该字段才生效。

	DeleteAlarm *string `json:"delete_alarm,omitempty"`
}

func (o BatchPauseScalingPoliciesOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchPauseScalingPoliciesOption struct{}"
	}

	return strings.Join([]string{"BatchPauseScalingPoliciesOption", string(data)}, " ")
}

type BatchPauseScalingPoliciesOptionForceDelete struct {
	value string
}

type BatchPauseScalingPoliciesOptionForceDeleteEnum struct {
	NO  BatchPauseScalingPoliciesOptionForceDelete
	YES BatchPauseScalingPoliciesOptionForceDelete
}

func GetBatchPauseScalingPoliciesOptionForceDeleteEnum() BatchPauseScalingPoliciesOptionForceDeleteEnum {
	return BatchPauseScalingPoliciesOptionForceDeleteEnum{
		NO: BatchPauseScalingPoliciesOptionForceDelete{
			value: "no",
		},
		YES: BatchPauseScalingPoliciesOptionForceDelete{
			value: "yes",
		},
	}
}

func (c BatchPauseScalingPoliciesOptionForceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchPauseScalingPoliciesOptionForceDelete) UnmarshalJSON(b []byte) error {
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

type BatchPauseScalingPoliciesOptionAction struct {
	value string
}

type BatchPauseScalingPoliciesOptionActionEnum struct {
	PAUSE BatchPauseScalingPoliciesOptionAction
}

func GetBatchPauseScalingPoliciesOptionActionEnum() BatchPauseScalingPoliciesOptionActionEnum {
	return BatchPauseScalingPoliciesOptionActionEnum{
		PAUSE: BatchPauseScalingPoliciesOptionAction{
			value: "pause",
		},
	}
}

func (c BatchPauseScalingPoliciesOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchPauseScalingPoliciesOptionAction) UnmarshalJSON(b []byte) error {
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
