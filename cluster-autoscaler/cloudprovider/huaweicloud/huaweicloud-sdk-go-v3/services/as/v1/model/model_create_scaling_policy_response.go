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
type CreateScalingPolicyResponse struct {
	// 伸缩策略ID。
	ScalingPolicyId *string `json:"scaling_policy_id,omitempty"`
}

func (o CreateScalingPolicyResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingPolicyResponse", string(data)}, " ")
}
