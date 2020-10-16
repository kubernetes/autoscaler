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

// Response Object
type ListScalingNotificationsResponse struct {
	// 伸缩组通知列表。
	Topics *[]Topics `json:"topics,omitempty"`
}

func (o ListScalingNotificationsResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingNotificationsResponse", string(data)}, " ")
}
