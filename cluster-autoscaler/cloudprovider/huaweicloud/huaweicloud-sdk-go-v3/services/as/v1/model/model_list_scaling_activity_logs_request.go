package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ListScalingActivityLogsRequest struct {
	// 伸缩组ID。

	ScalingGroupId string `json:"scaling_group_id"`
	// 查询的起始时间，格式是“yyyy-MM-ddThh:mm:ssZ”。

	StartTime *string `json:"start_time,omitempty"`
	// 查询的截止时间，格式是“yyyy-MM-ddThh:mm:ssZ”。

	EndTime *string `json:"end_time,omitempty"`
	// 查询的起始行号，默认为0。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数，默认20，最大100。

	Limit *int32 `json:"limit,omitempty"`
}

func (o ListScalingActivityLogsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingActivityLogsRequest struct{}"
	}

	return strings.Join([]string{"ListScalingActivityLogsRequest", string(data)}, " ")
}
