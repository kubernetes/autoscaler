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
type ListLifeCycleHooksRequest struct {
	ScalingGroupId string `json:"scaling_group_id"`
}

func (o ListLifeCycleHooksRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListLifeCycleHooksRequest", string(data)}, " ")
}
