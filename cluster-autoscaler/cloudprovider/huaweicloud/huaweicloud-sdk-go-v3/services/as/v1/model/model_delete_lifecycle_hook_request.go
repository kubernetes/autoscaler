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
type DeleteLifecycleHookRequest struct {
	ScalingGroupId    string `json:"scaling_group_id"`
	LifecycleHookName string `json:"lifecycle_hook_name"`
}

func (o DeleteLifecycleHookRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteLifecycleHookRequest", string(data)}, " ")
}
