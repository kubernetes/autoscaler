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
type UpdateScalingGroupRequest struct {
	ScalingGroupId string                         `json:"scaling_group_id"`
	Body           *UpdateScalingGroupRequestBody `json:"body,omitempty"`
}

func (o UpdateScalingGroupRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateScalingGroupRequest", string(data)}, " ")
}
