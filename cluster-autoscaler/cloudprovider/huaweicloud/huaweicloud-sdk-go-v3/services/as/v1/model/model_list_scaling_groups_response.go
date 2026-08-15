package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListScalingGroupsResponse struct {
	// 总记录数

	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的开始记录号

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数

	Limit *int32 `json:"limit,omitempty"`
	// 伸缩组列表

	ScalingGroups  *[]ScalingGroups `json:"scaling_groups,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ListScalingGroupsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingGroupsResponse struct{}"
	}

	return strings.Join([]string{"ListScalingGroupsResponse", string(data)}, " ")
}
