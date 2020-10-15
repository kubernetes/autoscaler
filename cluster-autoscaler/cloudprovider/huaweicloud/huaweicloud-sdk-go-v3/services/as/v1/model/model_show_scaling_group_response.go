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
type ShowScalingGroupResponse struct {
	ScalingGroup *ScalingGroups `json:"scaling_group,omitempty"`
}

func (o ShowScalingGroupResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowScalingGroupResponse", string(data)}, " ")
}
