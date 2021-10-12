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
type CompleteLifecycleActionRequest struct {
	ScalingGroupId string                              `json:"scaling_group_id"`
	Body           *CompleteLifecycleActionRequestBody `json:"body,omitempty"`
}

func (o CompleteLifecycleActionRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CompleteLifecycleActionRequest", string(data)}, " ")
}
