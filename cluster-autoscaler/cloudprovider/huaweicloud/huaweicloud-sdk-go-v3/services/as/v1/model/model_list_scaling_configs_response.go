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
type ListScalingConfigsResponse struct {
	// 总记录数。
	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的起始行号。
	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数。
	Limit *int32 `json:"limit,omitempty"`
	// 伸缩配置列表
	ScalingConfigurations *[]ScalingConfiguration `json:"scaling_configurations,omitempty"`
}

func (o ListScalingConfigsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingConfigsResponse", string(data)}, " ")
}
