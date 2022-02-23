package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type ResourceTags struct {
	// 资源标签值。最大长度36个unicode字符。

	Key *string `json:"key,omitempty"`
	// 资源标签值。

	Value *string `json:"value,omitempty"`
}

func (o ResourceTags) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResourceTags struct{}"
	}

	return strings.Join([]string{"ResourceTags", string(data)}, " ")
}
