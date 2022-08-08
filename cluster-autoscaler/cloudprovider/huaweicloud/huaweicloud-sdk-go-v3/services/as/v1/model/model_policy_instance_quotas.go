package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 配额列表
type PolicyInstanceQuotas struct {
	// 配额资源详情。

	Resources *[]PolicyInstanceResources `json:"resources,omitempty"`
}

func (o PolicyInstanceQuotas) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "PolicyInstanceQuotas struct{}"
	}

	return strings.Join([]string{"PolicyInstanceQuotas", string(data)}, " ")
}
