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
type BatchCreateServerTagsRequest struct {
	ServerId string                            `json:"server_id"`
	Body     *BatchCreateServerTagsRequestBody `json:"body,omitempty"`
}

func (o BatchCreateServerTagsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchCreateServerTagsRequest", string(data)}, " ")
}
