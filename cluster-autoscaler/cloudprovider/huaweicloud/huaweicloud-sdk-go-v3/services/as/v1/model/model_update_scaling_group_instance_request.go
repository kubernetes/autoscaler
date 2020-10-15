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
type UpdateScalingGroupInstanceRequest struct {
	ScalingGroupId string                                 `json:"scaling_group_id"`
	Body           *UpdateScalingGroupInstanceRequestBody `json:"body,omitempty"`
}

func (o UpdateScalingGroupInstanceRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateScalingGroupInstanceRequest", string(data)}, " ")
}
