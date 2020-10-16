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
type DeleteScalingTagsRequest struct {
	ResourceType DeleteScalingTagsRequestResourceType `json:"resource_type"`
	ResourceId   string                               `json:"resource_id"`
	Body         *DeleteScalingTagsRequestBody        `json:"body,omitempty"`
}

func (o DeleteScalingTagsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingTagsRequest", string(data)}, " ")
}

type DeleteScalingTagsRequestResourceType struct {
	value string
}

type DeleteScalingTagsRequestResourceTypeEnum struct {
	SCALING_GROUP_TAG DeleteScalingTagsRequestResourceType
}

func GetDeleteScalingTagsRequestResourceTypeEnum() DeleteScalingTagsRequestResourceTypeEnum {
	return DeleteScalingTagsRequestResourceTypeEnum{
		SCALING_GROUP_TAG: DeleteScalingTagsRequestResourceType{
			value: "scaling_group_tag",
		},
	}
}

func (c DeleteScalingTagsRequestResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteScalingTagsRequestResourceType) UnmarshalJSON(b []byte) error {
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
