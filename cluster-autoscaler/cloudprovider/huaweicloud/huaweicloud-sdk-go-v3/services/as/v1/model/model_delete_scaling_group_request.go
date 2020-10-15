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
type DeleteScalingGroupRequest struct {
	ScalingGroupId string                                `json:"scaling_group_id"`
	ForceDelete    *DeleteScalingGroupRequestForceDelete `json:"force_delete,omitempty"`
}

func (o DeleteScalingGroupRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingGroupRequest", string(data)}, " ")
}

type DeleteScalingGroupRequestForceDelete struct {
	value string
}

type DeleteScalingGroupRequestForceDeleteEnum struct {
	TRUE  DeleteScalingGroupRequestForceDelete
	FALSE DeleteScalingGroupRequestForceDelete
}

func GetDeleteScalingGroupRequestForceDeleteEnum() DeleteScalingGroupRequestForceDeleteEnum {
	return DeleteScalingGroupRequestForceDeleteEnum{
		TRUE: DeleteScalingGroupRequestForceDelete{
			value: "true",
		},
		FALSE: DeleteScalingGroupRequestForceDelete{
			value: "false",
		},
	}
}

func (c DeleteScalingGroupRequestForceDelete) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteScalingGroupRequestForceDelete) UnmarshalJSON(b []byte) error {
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
