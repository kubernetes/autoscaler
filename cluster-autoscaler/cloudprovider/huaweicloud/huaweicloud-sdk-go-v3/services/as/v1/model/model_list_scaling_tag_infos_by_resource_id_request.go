package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ListScalingTagInfosByResourceIdRequest struct {
	// 资源类型，枚举类：scaling_group_tag。scaling_group_tag表示资源类型为伸缩组。

	ResourceType ListScalingTagInfosByResourceIdRequestResourceType `json:"resource_type"`
	// 资源ID。

	ResourceId string `json:"resource_id"`
}

func (o ListScalingTagInfosByResourceIdRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingTagInfosByResourceIdRequest struct{}"
	}

	return strings.Join([]string{"ListScalingTagInfosByResourceIdRequest", string(data)}, " ")
}

type ListScalingTagInfosByResourceIdRequestResourceType struct {
	value string
}

type ListScalingTagInfosByResourceIdRequestResourceTypeEnum struct {
	SCALING_GROUP_TAG ListScalingTagInfosByResourceIdRequestResourceType
}

func GetListScalingTagInfosByResourceIdRequestResourceTypeEnum() ListScalingTagInfosByResourceIdRequestResourceTypeEnum {
	return ListScalingTagInfosByResourceIdRequestResourceTypeEnum{
		SCALING_GROUP_TAG: ListScalingTagInfosByResourceIdRequestResourceType{
			value: "scaling_group_tag",
		},
	}
}

func (c ListScalingTagInfosByResourceIdRequestResourceType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingTagInfosByResourceIdRequestResourceType) UnmarshalJSON(b []byte) error {
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
