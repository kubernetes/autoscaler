package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

type Runtime struct {
	// 容器运行时，默认为“docker”

	Name *RuntimeName `json:"name,omitempty"`
}

func (o Runtime) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Runtime struct{}"
	}

	return strings.Join([]string{"Runtime", string(data)}, " ")
}

type RuntimeName struct {
	value string
}

type RuntimeNameEnum struct {
	DOCKER     RuntimeName
	CONTAINERD RuntimeName
}

func GetRuntimeNameEnum() RuntimeNameEnum {
	return RuntimeNameEnum{
		DOCKER: RuntimeName{
			value: "docker",
		},
		CONTAINERD: RuntimeName{
			value: "containerd",
		},
	}
}

func (c RuntimeName) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *RuntimeName) UnmarshalJSON(b []byte) error {
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
