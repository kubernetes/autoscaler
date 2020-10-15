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
type BatchStartServersRequest struct {
	Body *BatchStartServersRequestBody `json:"body,omitempty"`
}

func (o BatchStartServersRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchStartServersRequest", string(data)}, " ")
}
