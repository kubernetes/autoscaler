/*
 * ecs
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// This is a auto create Body Object
type NovaCreateServersRequestBody struct {
	Server           *NovaCreateServersOption        `json:"server"`
	OsschedulerHints *NovaCreateServersSchedulerHint `json:"os:scheduler_hints,omitempty"`
}

func (o NovaCreateServersRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"NovaCreateServersRequestBody", string(data)}, " ")
}
