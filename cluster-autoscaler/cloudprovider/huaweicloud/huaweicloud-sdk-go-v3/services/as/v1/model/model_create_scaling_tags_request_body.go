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

// 标签列表
type CreateScalingTagsRequestBody struct {
	// 标签列表。
	Tags *[]TagsSingleValue `json:"tags,omitempty"`
	// 操作标识（区分大小写）：create：创建。若已经存在相同的key值则会覆盖对应的value值。
	Action *CreateScalingTagsRequestBodyAction `json:"action,omitempty"`
}

func (o CreateScalingTagsRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingTagsRequestBody", string(data)}, " ")
}

type CreateScalingTagsRequestBodyAction struct {
	value string
}

type CreateScalingTagsRequestBodyActionEnum struct {
	CREATE CreateScalingTagsRequestBodyAction
}

func GetCreateScalingTagsRequestBodyActionEnum() CreateScalingTagsRequestBodyActionEnum {
	return CreateScalingTagsRequestBodyActionEnum{
		CREATE: CreateScalingTagsRequestBodyAction{
			value: "create",
		},
	}
}

func (c CreateScalingTagsRequestBodyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateScalingTagsRequestBodyAction) UnmarshalJSON(b []byte) error {
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
