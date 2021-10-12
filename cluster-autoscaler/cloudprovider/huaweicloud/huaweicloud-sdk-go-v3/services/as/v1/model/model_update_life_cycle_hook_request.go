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
type UpdateLifeCycleHookRequest struct {
	ScalingGroupId    string                          `json:"scaling_group_id"`
	LifecycleHookName string                          `json:"lifecycle_hook_name"`
	Body              *UpdateLifeCycleHookRequestBody `json:"body,omitempty"`
}

func (o UpdateLifeCycleHookRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateLifeCycleHookRequest", string(data)}, " ")
}
