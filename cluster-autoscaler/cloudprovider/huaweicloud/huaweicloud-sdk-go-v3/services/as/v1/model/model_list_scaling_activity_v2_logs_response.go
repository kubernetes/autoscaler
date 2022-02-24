package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListScalingActivityV2LogsResponse struct {
	// 总记录数。

	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的其实行号。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数。

	Limit *int32 `json:"limit,omitempty"`
	// 伸缩活动日志列表。

	ScalingActivityLog *[]ScalingActivityLogV2 `json:"scaling_activity_log,omitempty"`
	HttpStatusCode     int                     `json:"-"`
}

func (o ListScalingActivityV2LogsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingActivityV2LogsResponse struct{}"
	}

	return strings.Join([]string{"ListScalingActivityV2LogsResponse", string(data)}, " ")
}
