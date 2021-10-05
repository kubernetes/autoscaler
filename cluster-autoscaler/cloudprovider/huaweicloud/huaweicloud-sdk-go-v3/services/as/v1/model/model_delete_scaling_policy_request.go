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
type DeleteScalingPolicyRequest struct {
	ScalingPolicyId string `json:"scaling_policy_id"`
}

func (o DeleteScalingPolicyRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingPolicyRequest", string(data)}, " ")
}
