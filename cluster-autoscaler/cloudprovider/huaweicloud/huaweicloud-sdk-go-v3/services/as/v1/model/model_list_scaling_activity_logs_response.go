package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListScalingActivityLogsResponse struct {
	// 总记录数。

	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的其实行号。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数。

	Limit *int32 `json:"limit,omitempty"`
	// 伸缩活动日志列表。

	ScalingActivityLog *[]ScalingActivityLogList `json:"scaling_activity_log,omitempty"`
	HttpStatusCode     int                       `json:"-"`
}

func (o ListScalingActivityLogsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingActivityLogsResponse struct{}"
	}

	return strings.Join([]string{"ListScalingActivityLogsResponse", string(data)}, " ")
}
