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
type CreateScalingPolicyRequest struct {
	Body *CreateScalingPolicyRequestBody `json:"body,omitempty"`
}

func (o CreateScalingPolicyRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingPolicyRequest", string(data)}, " ")
}
