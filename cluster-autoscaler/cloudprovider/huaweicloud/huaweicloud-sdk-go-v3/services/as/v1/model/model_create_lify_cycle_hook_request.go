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
type CreateLifyCycleHookRequest struct {
	ScalingGroupId string                          `json:"scaling_group_id"`
	Body           *CreateLifeCycleHookRequestBody `json:"body,omitempty"`
}

func (o CreateLifyCycleHookRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateLifyCycleHookRequest", string(data)}, " ")
}
