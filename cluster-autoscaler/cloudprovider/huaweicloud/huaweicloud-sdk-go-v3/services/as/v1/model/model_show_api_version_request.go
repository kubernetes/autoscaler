package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ShowApiVersionRequest struct {
	// API版本ID。

	ApiVersion ShowApiVersionRequestApiVersion `json:"api_version"`
}

func (o ShowApiVersionRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowApiVersionRequest struct{}"
	}

	return strings.Join([]string{"ShowApiVersionRequest", string(data)}, " ")
}

type ShowApiVersionRequestApiVersion struct {
	value string
}

type ShowApiVersionRequestApiVersionEnum struct {
	V1 ShowApiVersionRequestApiVersion
	V2 ShowApiVersionRequestApiVersion
}

func GetShowApiVersionRequestApiVersionEnum() ShowApiVersionRequestApiVersionEnum {
	return ShowApiVersionRequestApiVersionEnum{
		V1: ShowApiVersionRequestApiVersion{
			value: "v1",
		},
		V2: ShowApiVersionRequestApiVersion{
			value: "v2",
		},
	}
}

func (c ShowApiVersionRequestApiVersion) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ShowApiVersionRequestApiVersion) UnmarshalJSON(b []byte) error {
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
