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
type CreateScalingTagsRequest struct {
	ResourceType CreateScalingTagsRequestResourceType `json:"resource_type"`
	ResourceId   string                               `json:"resource_id"`
	Body         *CreateScalingTagsRequestBody        `json:"body,omitempty"`
}

func (o CreateScalingTagsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingTagsRequest", string(data)}, " ")
}

type CreateScalingTagsRequestResourceType struct {
	value string
}

type CreateScalingTagsRequestResourceTypeEnum struct {
	SCALING_GROUP_TAG CreateScalingTagsRequestResourceType
}

func GetCreateScalingTagsRequestResourceTypeEnum() CreateScalingTagsRequestResourceTypeEnum {
	return CreateScalingTagsRequestResourceTypeEnum{
		SCALING_GROUP_TAG: CreateScalingTagsRequestResourceType{
			value: "scaling_group_tag",
		},
	}
}

func (c CreateScalingTagsRequestResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateScalingTagsRequestResourceType) UnmarshalJSON(b []byte) error {
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
