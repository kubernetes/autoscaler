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

// Request Object
type ExecuteScalingPolicyRequest struct {
	ScalingPolicyId string                           `json:"scaling_policy_id"`
	Body            *ExecuteScalingPolicyRequestBody `json:"body,omitempty"`
}

func (o ExecuteScalingPolicyRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ExecuteScalingPolicyRequest", string(data)}, " ")
}
