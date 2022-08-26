package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type Matches struct {
	// 键。暂限定为resource_name。

	Key string `json:"key"`
	// 值。为固定字典值。每个值最大长度255个unicode字符。若为空字符串、resource_id时为精确匹配。

	Value string `json:"value"`
}

func (o Matches) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Matches struct{}"
	}

	return strings.Join([]string{"Matches", string(data)}, " ")
}
