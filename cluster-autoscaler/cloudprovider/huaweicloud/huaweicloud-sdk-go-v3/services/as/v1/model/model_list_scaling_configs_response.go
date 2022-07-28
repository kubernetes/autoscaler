package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListScalingConfigsResponse struct {
	// 总记录数。

	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的起始行号。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数。

	Limit *int32 `json:"limit,omitempty"`
	// 伸缩配置列表

	ScalingConfigurations *[]ScalingConfiguration `json:"scaling_configurations,omitempty"`
	HttpStatusCode        int                     `json:"-"`
}

func (o ListScalingConfigsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingConfigsResponse struct{}"
	}

	return strings.Join([]string{"ListScalingConfigsResponse", string(data)}, " ")
}
