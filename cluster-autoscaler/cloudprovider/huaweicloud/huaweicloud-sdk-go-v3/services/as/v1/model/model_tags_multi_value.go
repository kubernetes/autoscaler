package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type TagsMultiValue struct {
	// 资源标签键。最大长度127个unicode字符。key不能为空。（搜索时不对此参数做校验）。最多为10个，不能为空或者空字符串。且不能重复。

	Key string `json:"key"`
	// 资源标签值列表。每个值最大长度255个unicode字符，每个key下最多为10个，同一个key中values不能重复。

	Values []string `json:"values"`
}

func (o TagsMultiValue) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "TagsMultiValue struct{}"
	}

	return strings.Join([]string{"TagsMultiValue", string(data)}, " ")
}
