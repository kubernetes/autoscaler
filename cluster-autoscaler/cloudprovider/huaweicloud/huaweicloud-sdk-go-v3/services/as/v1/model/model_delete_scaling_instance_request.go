package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type DeleteScalingInstanceRequest struct {
	// 实例ID。

	InstanceId string `json:"instance_id"`
	// 实例移出伸缩组，是否删除云服务器实例。默认为no；可选值为yes或no。

	InstanceDelete *DeleteScalingInstanceRequestInstanceDelete `json:"instance_delete,omitempty"`
}

func (o DeleteScalingInstanceRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingInstanceRequest struct{}"
	}

	return strings.Join([]string{"DeleteScalingInstanceRequest", string(data)}, " ")
}

type DeleteScalingInstanceRequestInstanceDelete struct {
	value string
}

type DeleteScalingInstanceRequestInstanceDeleteEnum struct {
	YES DeleteScalingInstanceRequestInstanceDelete
	NO  DeleteScalingInstanceRequestInstanceDelete
}

func GetDeleteScalingInstanceRequestInstanceDeleteEnum() DeleteScalingInstanceRequestInstanceDeleteEnum {
	return DeleteScalingInstanceRequestInstanceDeleteEnum{
		YES: DeleteScalingInstanceRequestInstanceDelete{
			value: "yes",
		},
		NO: DeleteScalingInstanceRequestInstanceDelete{
			value: "no",
		},
	}
}

func (c DeleteScalingInstanceRequestInstanceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *DeleteScalingInstanceRequestInstanceDelete) UnmarshalJSON(b []byte) error {
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
