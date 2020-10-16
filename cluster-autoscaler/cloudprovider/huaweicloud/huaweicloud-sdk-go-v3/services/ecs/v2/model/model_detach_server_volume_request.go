/*
 * ecs
 *
 * ECS Open API
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
type DetachServerVolumeRequest struct {
	ServerId   string                               `json:"server_id"`
	VolumeId   string                               `json:"volume_id"`
	DeleteFlag *DetachServerVolumeRequestDeleteFlag `json:"delete_flag,omitempty"`
}

func (o DetachServerVolumeRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DetachServerVolumeRequest", string(data)}, " ")
}

type DetachServerVolumeRequestDeleteFlag struct {
	value string
}

type DetachServerVolumeRequestDeleteFlagEnum struct {
	E_0 DetachServerVolumeRequestDeleteFlag
	E_1 DetachServerVolumeRequestDeleteFlag
}

func GetDetachServerVolumeRequestDeleteFlagEnum() DetachServerVolumeRequestDeleteFlagEnum {
	return DetachServerVolumeRequestDeleteFlagEnum{
		E_0: DetachServerVolumeRequestDeleteFlag{
			value: "0",
		},
		E_1: DetachServerVolumeRequestDeleteFlag{
			value: "1",
		},
	}
}

func (c DetachServerVolumeRequestDeleteFlag) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DetachServerVolumeRequestDeleteFlag) UnmarshalJSON(b []byte) error {
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
