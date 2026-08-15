package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 请求正常
type QueryTagsOption struct {
	// 过滤条件，包含标签，最多包含10个Tag，结构体不能缺失。

	Tags *[]TagsMultiValue `json:"tags,omitempty"`
	// 过滤条件，包含任意标签，最多包含10个Tag。

	TagsAny *[]TagsMultiValue `json:"tags_any,omitempty"`
	// 过滤条件，不包含标签，最多包含10个Tag。

	NotTags *[]TagsMultiValue `json:"not_tags,omitempty"`
	// 过滤条件，不包含任意标签，最多包含10个Tag。

	NotTagsAny *[]TagsMultiValue `json:"not_tags_any,omitempty"`
	// 查询记录数（action为count时无此参数）如果action为filter默认为1000，limit最多为1000，不能为负数，最小值为1。

	Limit *string `json:"limit,omitempty"`
	// 分页位置标识（资源ID或索引位置）。

	Marker *string `json:"marker,omitempty"`
	// 操作标识（仅限filter，count）：filter（过滤）：即分页查询。count（查询总条数）：按照条件将总条数返回即可。

	Action QueryTagsOptionAction `json:"action"`
	// （索引位置），从offset指定的下一条数据开始查询。查询第一页数据时，不需要传入此参数。查询后续页码数据时，将查询前一页数据时，不需要传入此参数。查询后续页码数据时，将查询前一页数据时响应体中的值带入此参数。必须为数字，不能为负数。action：count时，无此参数。action：filter时，默认为0

	Offset *string `json:"offset,omitempty"`
	// 模糊搜索字段。

	Matches *[]Matches `json:"matches,omitempty"`
}

func (o QueryTagsOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "QueryTagsOption struct{}"
	}

	return strings.Join([]string{"QueryTagsOption", string(data)}, " ")
}

type QueryTagsOptionAction struct {
	value string
}

type QueryTagsOptionActionEnum struct {
	FILTER QueryTagsOptionAction
	COUNT  QueryTagsOptionAction
}

func GetQueryTagsOptionActionEnum() QueryTagsOptionActionEnum {
	return QueryTagsOptionActionEnum{
		FILTER: QueryTagsOptionAction{
			value: "filter",
		},
		COUNT: QueryTagsOptionAction{
			value: "count",
		},
	}
}

func (c QueryTagsOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *QueryTagsOptionAction) UnmarshalJSON(b []byte) error {
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
