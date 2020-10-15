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
type ListScalingPoliciesResponse struct {
	// 总记录数。
	TotalNumber *int32 `json:"total_number,omitempty"`
	// 查询的起始行号。
	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询记录数。
	Limit *int32 `json:"limit,omitempty"`
	// 伸缩策略列表
	ScalingPolicies *[]ScalingPolicyDetail `json:"scaling_policies,omitempty"`
}

func (o ListScalingPoliciesResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingPoliciesResponse", string(data)}, " ")
}
