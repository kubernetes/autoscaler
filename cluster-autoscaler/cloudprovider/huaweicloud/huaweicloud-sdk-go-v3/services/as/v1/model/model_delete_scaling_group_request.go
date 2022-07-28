package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type DeleteScalingGroupRequest struct {
	// 伸缩组ID。

	ScalingGroupId string `json:"scaling_group_id"`
	// 是否强制删除伸缩组。默认为no；可选值为yes或no。

	ForceDelete *DeleteScalingGroupRequestForceDelete `json:"force_delete,omitempty"`
}

func (o DeleteScalingGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingGroupRequest struct{}"
	}

	return strings.Join([]string{"DeleteScalingGroupRequest", string(data)}, " ")
}

type DeleteScalingGroupRequestForceDelete struct {
	value string
}

type DeleteScalingGroupRequestForceDeleteEnum struct {
	YES DeleteScalingGroupRequestForceDelete
	NO  DeleteScalingGroupRequestForceDelete
}

func GetDeleteScalingGroupRequestForceDeleteEnum() DeleteScalingGroupRequestForceDeleteEnum {
	return DeleteScalingGroupRequestForceDeleteEnum{
		YES: DeleteScalingGroupRequestForceDelete{
			value: "yes",
		},
		NO: DeleteScalingGroupRequestForceDelete{
			value: "no",
		},
	}
}

func (c DeleteScalingGroupRequestForceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
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
