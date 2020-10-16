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
type DeleteScalingNotificationRequest struct {
	ScalingGroupId string `json:"scaling_group_id"`
	TopicUrn       string `json:"topic_urn"`
}

func (o DeleteScalingNotificationRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingNotificationRequest", string(data)}, " ")
}
