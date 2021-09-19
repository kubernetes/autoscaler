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

// Request Object
type DeleteScalingInstanceRequest struct {
	InstanceId     string                                      `json:"instance_id"`
	InstanceDelete *DeleteScalingInstanceRequestInstanceDelete `json:"instance_delete,omitempty"`
}

func (o DeleteScalingInstanceRequest) String() string {
	data, _ := json.Marshal(o)
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
	return json.Marshal(c.value)
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
