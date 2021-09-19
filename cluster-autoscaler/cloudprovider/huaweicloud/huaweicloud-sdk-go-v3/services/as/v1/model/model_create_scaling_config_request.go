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
type CreateScalingConfigRequest struct {
	Body *CreateScalingConfigRequestBody `json:"body,omitempty"`
}

func (o CreateScalingConfigRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingConfigRequest", string(data)}, " ")
}
