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
type BatchStopServersRequestBody struct {
	OsStop *BatchStopServersOption `json:"os-stop"`
}

func (o BatchStopServersRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchStopServersRequestBody", string(data)}, " ")
}
