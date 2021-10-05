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
type CreateScalingGroupRequest struct {
	Body *CreateScalingGroupRequestBody `json:"body,omitempty"`
}

func (o CreateScalingGroupRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingGroupRequest", string(data)}, " ")
}
