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
type ListScalingInstancesResponse struct {
	// 总记录数。
	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的起始行号。
	StartNumber *int32 `json:"start_number,omitempty"`
	// 伸缩组实例详情。
	Limit *int32 `json:"limit,omitempty"`
	// 伸缩组实例详情。
	ScalingGroupInstances *[]ScalingGroupInstance `json:"scaling_group_instances,omitempty"`
}

func (o ListScalingInstancesResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingInstancesResponse", string(data)}, " ")
}
