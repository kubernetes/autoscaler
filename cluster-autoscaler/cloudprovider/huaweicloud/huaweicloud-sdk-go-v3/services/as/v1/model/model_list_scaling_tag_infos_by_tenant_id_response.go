package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListScalingTagInfosByTenantIdResponse struct {
	// 资源标签。

	Tags           *[]TagsMultiValue `json:"tags,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ListScalingTagInfosByTenantIdResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingTagInfosByTenantIdResponse struct{}"
	}

	return strings.Join([]string{"ListScalingTagInfosByTenantIdResponse", string(data)}, " ")
}
