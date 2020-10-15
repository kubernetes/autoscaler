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
type DeleteScalingTagsRequestBody struct {
	// 标签列表。action为delete时，tags结构体不能缺失，key不能为空，或者空字符串。
	Tags *[]TagsSingleValue `json:"tags,omitempty"`
	// 操作标识（区分大小写）：delete：删除。
	Action *DeleteScalingTagsRequestBodyAction `json:"action,omitempty"`
}

func (o DeleteScalingTagsRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingTagsRequestBody", string(data)}, " ")
}

type DeleteScalingTagsRequestBodyAction struct {
	value string
}

type DeleteScalingTagsRequestBodyActionEnum struct {
	DELETE DeleteScalingTagsRequestBodyAction
}

func GetDeleteScalingTagsRequestBodyActionEnum() DeleteScalingTagsRequestBodyActionEnum {
	return DeleteScalingTagsRequestBodyActionEnum{
		DELETE: DeleteScalingTagsRequestBodyAction{
			value: "delete",
		},
	}
}

func (c DeleteScalingTagsRequestBodyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteScalingTagsRequestBodyAction) UnmarshalJSON(b []byte) error {
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
