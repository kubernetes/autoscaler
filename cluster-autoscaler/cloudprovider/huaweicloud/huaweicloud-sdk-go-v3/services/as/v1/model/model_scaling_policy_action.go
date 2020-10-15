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

// 策略执行具体动作
type ScalingPolicyAction struct {
	// 操作选项。ADD：添加实例。REMOVE/REDUCE：移除实例。SET：设置实例数为
	Operation *ScalingPolicyActionOperation `json:"operation,omitempty"`
	// 操作实例个数，默认为1。配置参数时，instance_number和instance_percentage参数只能选其中一个进行配置。
	InstanceNumber *int32 `json:"instance_number,omitempty"`
	// 操作实例百分比，将当前组容量增加、减少或设置为指定的百分比。当instance_number和instance_percentage参数均无配置时，则操作实例个数为1。配置参数时，instance_number和instance_percentage参数只能选其中一个进行配置。
	InstancePercentage *int32 `json:"instance_percentage,omitempty"`
}

func (o ScalingPolicyAction) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ScalingPolicyAction", string(data)}, " ")
}

type ScalingPolicyActionOperation struct {
	value string
}

type ScalingPolicyActionOperationEnum struct {
	ADD    ScalingPolicyActionOperation
	REMOVE ScalingPolicyActionOperation
	REDUCE ScalingPolicyActionOperation
	SET    ScalingPolicyActionOperation
}

func GetScalingPolicyActionOperationEnum() ScalingPolicyActionOperationEnum {
	return ScalingPolicyActionOperationEnum{
		ADD: ScalingPolicyActionOperation{
			value: "ADD",
		},
		REMOVE: ScalingPolicyActionOperation{
			value: "REMOVE",
		},
		REDUCE: ScalingPolicyActionOperation{
			value: "REDUCE",
		},
		SET: ScalingPolicyActionOperation{
			value: "SET",
		},
	}
}

func (c ScalingPolicyActionOperation) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ScalingPolicyActionOperation) UnmarshalJSON(b []byte) error {
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
