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
type BatchStartServersRequestBody struct {
	OsStart *BatchStartServersOption `json:"os-start"`
}

func (o BatchStartServersRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchStartServersRequestBody", string(data)}, " ")
}
