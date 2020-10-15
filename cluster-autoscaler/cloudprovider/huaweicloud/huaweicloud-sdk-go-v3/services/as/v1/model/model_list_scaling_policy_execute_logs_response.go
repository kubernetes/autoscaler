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
type ListScalingPolicyExecuteLogsResponse struct {
	// 总记录数。
	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的起始行号。
	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数。
	Limit *int32 `json:"limit,omitempty"`
	// 伸缩策略执行日志列表。
	ScalingPolicyExecuteLog *[]ScalingPolicyExecuteLogList `json:"scaling_policy_execute_log,omitempty"`
}

func (o ListScalingPolicyExecuteLogsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingPolicyExecuteLogsResponse", string(data)}, " ")
}
