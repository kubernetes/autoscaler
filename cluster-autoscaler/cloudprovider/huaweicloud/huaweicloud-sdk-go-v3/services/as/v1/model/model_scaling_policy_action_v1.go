package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 策略执行具体动作
type ScalingPolicyActionV1 struct {
	// 操作选项。ADD：添加实例。REMOVE/REDUCE：移除实例。SET：设置实例数为

	Operation *ScalingPolicyActionV1Operation `json:"operation,omitempty"`
	// 操作实例个数，默认为1。配置参数时，instance_number和instance_percentage参数只能选其中一个进行配置。

	InstanceNumber *int32 `json:"instance_number,omitempty"`
	// 操作实例百分比，将当前组容量增加、减少或设置为指定的百分比。当instance_number和instance_percentage参数均无配置时，则操作实例个数为1。配置参数时，instance_number和instance_percentage参数只能选其中一个进行配置。

	InstancePercentage *int32 `json:"instance_percentage,omitempty"`
}

func (o ScalingPolicyActionV1) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingPolicyActionV1 struct{}"
	}

	return strings.Join([]string{"ScalingPolicyActionV1", string(data)}, " ")
}

type ScalingPolicyActionV1Operation struct {
	value string
}

type ScalingPolicyActionV1OperationEnum struct {
	ADD    ScalingPolicyActionV1Operation
	REMOVE ScalingPolicyActionV1Operation
	REDUCE ScalingPolicyActionV1Operation
	SET    ScalingPolicyActionV1Operation
}

func GetScalingPolicyActionV1OperationEnum() ScalingPolicyActionV1OperationEnum {
	return ScalingPolicyActionV1OperationEnum{
		ADD: ScalingPolicyActionV1Operation{
			value: "ADD",
		},
		REMOVE: ScalingPolicyActionV1Operation{
			value: "REMOVE",
		},
		REDUCE: ScalingPolicyActionV1Operation{
			value: "REDUCE",
		},
		SET: ScalingPolicyActionV1Operation{
			value: "SET",
		},
	}
}

func (c ScalingPolicyActionV1Operation) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ScalingPolicyActionV1Operation) UnmarshalJSON(b []byte) error {
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
