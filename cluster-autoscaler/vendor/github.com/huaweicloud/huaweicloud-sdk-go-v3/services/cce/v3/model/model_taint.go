package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 如下字段不可使用：  - node.kubernetes.io/memory-pressure - node.kubernetes.io/disk-pressure - node.kubernetes.io/out-of-disk - node.kubernetes.io/unschedulable - node.kubernetes.io/network-unavailable
type Taint struct {
	// 键

	Key *string `json:"key,omitempty"`
	// 值

	Value *string `json:"value,omitempty"`
	// 作用效果

	Effect *TaintEffect `json:"effect,omitempty"`
}

func (o Taint) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Taint struct{}"
	}

	return strings.Join([]string{"Taint", string(data)}, " ")
}

type TaintEffect struct {
	value string
}

type TaintEffectEnum struct {
	NO_SCHEDULE        TaintEffect
	PREFER_NO_SCHEDULE TaintEffect
	NO_EXECUTE         TaintEffect
}

func GetTaintEffectEnum() TaintEffectEnum {
	return TaintEffectEnum{
		NO_SCHEDULE: TaintEffect{
			value: "NoSchedule",
		},
		PREFER_NO_SCHEDULE: TaintEffect{
			value: "PreferNoSchedule",
		},
		NO_EXECUTE: TaintEffect{
			value: "NoExecute",
		},
	}
}

func (c TaintEffect) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *TaintEffect) UnmarshalJSON(b []byte) error {
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
