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
type ShowScalingPolicyRequest struct {
	ScalingPolicyId string `json:"scaling_policy_id"`
}

func (o ShowScalingPolicyRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowScalingPolicyRequest", string(data)}, " ")
}
