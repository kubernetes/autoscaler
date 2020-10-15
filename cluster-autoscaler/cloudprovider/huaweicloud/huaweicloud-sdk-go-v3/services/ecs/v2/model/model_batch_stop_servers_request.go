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

// Request Object
type BatchStopServersRequest struct {
	Body *BatchStopServersRequestBody `json:"body,omitempty"`
}

func (o BatchStopServersRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchStopServersRequest", string(data)}, " ")
}
