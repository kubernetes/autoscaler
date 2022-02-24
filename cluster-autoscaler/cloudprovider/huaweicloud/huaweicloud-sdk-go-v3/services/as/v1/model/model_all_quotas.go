package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 配额列表
type AllQuotas struct {
	// 配额详情资源列表。

	Resources *[]AllResources `json:"resources,omitempty"`
}

func (o AllQuotas) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AllQuotas struct{}"
	}

	return strings.Join([]string{"AllQuotas", string(data)}, " ")
}
