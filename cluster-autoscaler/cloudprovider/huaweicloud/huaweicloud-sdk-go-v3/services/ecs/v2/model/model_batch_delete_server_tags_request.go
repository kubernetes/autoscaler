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
type BatchDeleteServerTagsRequest struct {
	ServerId string                            `json:"server_id"`
	Body     *BatchDeleteServerTagsRequestBody `json:"body,omitempty"`
}

func (o BatchDeleteServerTagsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchDeleteServerTagsRequest", string(data)}, " ")
}
