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
type CreateScalingNotificationRequest struct {
	ScalingGroupId string                         `json:"scaling_group_id"`
	Body           *CreateNotificationRequestBody `json:"body,omitempty"`
}

func (o CreateScalingNotificationRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingNotificationRequest", string(data)}, " ")
}
