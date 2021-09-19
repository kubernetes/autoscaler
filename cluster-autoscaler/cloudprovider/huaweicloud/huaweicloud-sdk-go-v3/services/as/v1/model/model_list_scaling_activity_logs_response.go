/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

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
}

func (o ListScalingActivityLogsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingActivityLogsResponse", string(data)}, " ")
}
