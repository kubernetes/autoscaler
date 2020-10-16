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
type DeleteScalingNotificationResponse struct {
}

func (o DeleteScalingNotificationResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingNotificationResponse", string(data)}, " ")
}
