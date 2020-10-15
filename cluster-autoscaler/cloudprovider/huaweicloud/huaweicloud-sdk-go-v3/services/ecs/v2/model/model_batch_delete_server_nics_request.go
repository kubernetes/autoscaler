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
type BatchDeleteServerNicsRequest struct {
	ServerId string                            `json:"server_id"`
	Body     *BatchDeleteServerNicsRequestBody `json:"body,omitempty"`
}

func (o BatchDeleteServerNicsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchDeleteServerNicsRequest", string(data)}, " ")
}
