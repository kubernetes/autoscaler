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
type BatchAddServerNicsRequest struct {
	ServerId string                         `json:"server_id"`
	Body     *BatchAddServerNicsRequestBody `json:"body,omitempty"`
}

func (o BatchAddServerNicsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchAddServerNicsRequest", string(data)}, " ")
}
