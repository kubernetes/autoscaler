package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListScalingTagInfosByResourceIdResponse struct {
	// 资源标签列表。

	Tags *[]TagsSingleValue `json:"tags,omitempty"`
	// 系统资源标签列表。

	SysTags        *[]TagsSingleValue `json:"sys_tags,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListScalingTagInfosByResourceIdResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingTagInfosByResourceIdResponse struct{}"
	}

	return strings.Join([]string{"ListScalingTagInfosByResourceIdResponse", string(data)}, " ")
}
