package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type Resources struct {
	// 资源详情ID。

	ResourceId *string `json:"resource_id,omitempty"`
	// 资源详情。

	ResourceDetail *string `json:"resource_detail,omitempty"`
	// 标签列表，没有标签默认为空数组。

	Tags *[]ResourceTags `json:"tags,omitempty"`
	// 资源名称，没有默认为空字符串。

	ResourceName *string `json:"resource_name,omitempty"`
}

func (o Resources) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Resources struct{}"
	}

	return strings.Join([]string{"Resources", string(data)}, " ")
}
