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
type DeleteScalingGroupResponse struct {
}

func (o DeleteScalingGroupResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingGroupResponse", string(data)}, " ")
}
