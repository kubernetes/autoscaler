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
type CreateScalingGroupResponse struct {
	// 伸缩组ID
	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
}

func (o CreateScalingGroupResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingGroupResponse", string(data)}, " ")
}
