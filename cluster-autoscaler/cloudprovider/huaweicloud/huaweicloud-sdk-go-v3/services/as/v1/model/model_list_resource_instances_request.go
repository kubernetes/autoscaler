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
type ListResourceInstancesRequest struct {
	ResourceType ListResourceInstancesRequestResourceType `json:"resource_type"`
	Body         *ShowTagsRequestBody                     `json:"body,omitempty"`
}

func (o ListResourceInstancesRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListResourceInstancesRequest", string(data)}, " ")
}

type ListResourceInstancesRequestResourceType struct {
	value string
}

type ListResourceInstancesRequestResourceTypeEnum struct {
	SCALING_GROUP_TAG ListResourceInstancesRequestResourceType
}

func GetListResourceInstancesRequestResourceTypeEnum() ListResourceInstancesRequestResourceTypeEnum {
	return ListResourceInstancesRequestResourceTypeEnum{
		SCALING_GROUP_TAG: ListResourceInstancesRequestResourceType{
			value: "scaling_group_tag",
		},
	}
}

func (c ListResourceInstancesRequestResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListResourceInstancesRequestResourceType) UnmarshalJSON(b []byte) error {
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
